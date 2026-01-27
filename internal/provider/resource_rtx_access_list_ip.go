package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

// resourceRTXAccessListIP returns the schema for the rtx_access_list_ip resource.
// This is a group-based resource where multiple IP filter entries are managed together
// under a single name identifier.
func resourceRTXAccessListIP() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a group of IPv4 static filters (access list) on RTX routers. " +
			"This resource manages multiple IP filter rules as a single group using the RTX native 'ip filter' command. " +
			"Supports automatic sequence numbering or manual sequence assignment.",

		CreateContext: resourceRTXAccessListIPCreate,
		ReadContext:   resourceRTXAccessListIPRead,
		UpdateContext: resourceRTXAccessListIPUpdate,
		DeleteContext: resourceRTXAccessListIPDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXAccessListIPImport,
		},

		CustomizeDiff: customdiff.All(
			ValidateACLSchema,
			validateAccessListIPEntries,
		),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ACL group identifier. This name is used to reference the ACL in other resources and for Terraform state management.",
			},
			"sequence_start": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Starting sequence number for automatic sequence calculation. When set, sequence numbers are automatically assigned to entries based on their definition order. Mutually exclusive with entry-level sequence attributes.",
				ValidateFunc: validation.IntBetween(1, MaxSequenceValue),
			},
			"sequence_step": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      DefaultSequenceStep,
				Description:  fmt.Sprintf("Increment value for automatic sequence calculation. Only used when sequence_start is set. Default is %d.", DefaultSequenceStep),
				ValidateFunc: validation.IntBetween(1, MaxSequenceValue),
			},
			"apply": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of interface bindings. Each apply block binds this ACL to an interface in a specific direction.",
				Elem:        CommonApplySchema(),
			},
			"entry": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "List of IP filter entries. Each entry defines a single filter rule.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sequence": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							Description:  "Sequence number determines the order of evaluation. Required when sequence_start is not set (manual mode). Auto-calculated when sequence_start is set (auto mode).",
							ValidateFunc: validation.IntBetween(1, MaxSequenceValue),
						},
						"action": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Filter action: pass, reject, restrict, or restrict-log",
							ValidateFunc:     validation.StringInSlice([]string{"pass", "reject", "restrict", "restrict-log"}, true),
							DiffSuppressFunc: SuppressCaseDiff,
						},
						"source": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Source IP address/network in CIDR notation (e.g., '10.0.0.0/8') or '*' for any",
						},
						"destination": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Destination IP address/network in CIDR notation (e.g., '192.168.1.0/24') or '*' for any",
						},
						"protocol": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "*",
							Description:      "Protocol: tcp, udp, icmp, ip, gre, esp, ah, or * for any",
							ValidateFunc:     validation.StringInSlice([]string{"tcp", "udp", "udp,tcp", "tcp,udp", "icmp", "ip", "gre", "esp", "ah", "tcpfin", "tcprst", "*"}, true),
							DiffSuppressFunc: SuppressCaseDiff,
						},
						"source_port": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "*",
							Description: "Source port number, range (e.g., '1024-65535'), or '*' for any. Only valid for TCP/UDP.",
						},
						"dest_port": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "*",
							Description: "Destination port number, range (e.g., '80'), or '*' for any. Only valid for TCP/UDP.",
						},
						"established": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Match established TCP connections only. Only valid for TCP protocol.",
						},
						"log": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Enable logging when this entry matches traffic.",
						},
					},
				},
			},
		},
	}
}

// validateAccessListIPEntries validates IP filter entry constraints.
func validateAccessListIPEntries(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	entries := diff.Get("entry").([]interface{})

	for i, e := range entries {
		entry := e.(map[string]interface{})

		protocol := strings.ToLower(entry["protocol"].(string))
		established := entry["established"].(bool)
		sourcePort := entry["source_port"].(string)
		destPort := entry["dest_port"].(string)

		// Established is only valid for TCP
		if established && protocol != "tcp" {
			return fmt.Errorf("entry[%d]: established can only be set to true for tcp protocol", i)
		}

		// Port specifications valid for TCP/UDP and TCP-based protocols (tcpfin, tcprst)
		tcpBasedProtocols := protocol == "tcp" || protocol == "udp" || protocol == "tcp,udp" || protocol == "udp,tcp" || protocol == "tcpfin" || protocol == "tcprst"
		if !tcpBasedProtocols {
			if sourcePort != "*" && sourcePort != "" {
				return fmt.Errorf("entry[%d]: source_port can only be specified for tcp, udp, tcpfin, or tcprst protocols", i)
			}
			if destPort != "*" && destPort != "" {
				return fmt.Errorf("entry[%d]: dest_port can only be specified for tcp, udp, tcpfin, or tcprst protocols", i)
			}
		}
	}

	return nil
}

func resourceRTXAccessListIPCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	name := d.Get("name").(string)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ip", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ip").Msgf("Creating IP access list group: %s", name)

	// Build and create IP filters
	filters := buildIPFiltersFromResourceData(d)
	for _, filter := range filters {
		if err := apiClient.client.CreateIPFilter(ctx, filter); err != nil {
			return diag.Errorf("Failed to create IP filter %d: %v", filter.Number, err)
		}
	}

	// Handle apply blocks
	if err := applyIPFiltersToInterfacesFromResourceData(ctx, d, apiClient); err != nil {
		return diag.Errorf("Failed to apply IP filters to interfaces: %v", err)
	}

	d.SetId(name)

	return resourceRTXAccessListIPRead(ctx, d, meta)
}

func resourceRTXAccessListIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	name := d.Id()

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ip", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ip").Msgf("Reading IP access list group: %s", name)

	// Get the expected sequences from state to query
	sequences := getIPExpectedSequences(d)

	// Read each filter
	entries := make([]map[string]interface{}, 0, len(sequences))
	foundAny := false
	for _, seq := range sequences {
		filter, err := apiClient.client.GetIPFilter(ctx, seq)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				continue
			}
			return diag.Errorf("Failed to read IP filter %d: %v", seq, err)
		}
		foundAny = true

		e := map[string]interface{}{
			"sequence":    filter.Number,
			"action":      filter.Action,
			"source":      filter.SourceAddress,
			"destination": filter.DestAddress,
			"protocol":    filter.Protocol,
			"source_port": normalizePort(filter.SourcePort),
			"dest_port":   normalizePort(filter.DestPort),
			"established": filter.Established,
			"log":         false, // RTX doesn't return log status in filter read
		}
		entries = append(entries, e)
	}

	// If no entries found and we expected some, mark as deleted
	if !foundAny && len(sequences) > 0 {
		logger.Debug().Str("resource", "rtx_access_list_ip").Msgf("IP access list %s not found, removing from state", name)
		d.SetId("")
		return nil
	}

	// Set name
	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}

	// Set entries
	if err := d.Set("entry", entries); err != nil {
		return diag.FromErr(err)
	}

	// Read and set apply blocks
	if err := readIPApplyBlocksFromRouter(ctx, d, apiClient); err != nil {
		logger.Warn().Err(err).Msg("Failed to read apply blocks")
	}

	return nil
}

func resourceRTXAccessListIPUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	name := d.Id()

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ip", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ip").Msgf("Updating IP access list group: %s", name)

	// Handle entry changes
	if d.HasChange("entry") || d.HasChange("sequence_start") || d.HasChange("sequence_step") {
		// Get old sequences to potentially delete
		oldEntries, _ := d.GetChange("entry")
		oldSequences := extractIPSequencesFromEntryList(oldEntries.([]interface{}), d)
		newSequences := getIPExpectedSequences(d)

		// Delete removed sequences
		toDelete := findIPRemovedSequences(oldSequences, newSequences)
		for _, seq := range toDelete {
			if err := apiClient.client.DeleteIPFilter(ctx, seq); err != nil {
				if !strings.Contains(err.Error(), "not found") {
					logger.Warn().Err(err).Msgf("Failed to delete IP filter %d", seq)
				}
			}
		}

		// Create/update filters
		filters := buildIPFiltersFromResourceData(d)
		for _, filter := range filters {
			if err := apiClient.client.UpdateIPFilter(ctx, filter); err != nil {
				return diag.Errorf("Failed to update IP filter %d: %v", filter.Number, err)
			}
		}
	}

	// Handle apply changes
	if d.HasChange("apply") {
		oldApply, newApply := d.GetChange("apply")

		// Remove old applies
		for _, a := range oldApply.([]interface{}) {
			applyMap := a.(map[string]interface{})
			iface := applyMap["interface"].(string)
			direction := strings.ToLower(applyMap["direction"].(string))

			if err := apiClient.client.RemoveIPFiltersFromInterface(ctx, iface, direction); err != nil {
				logger.Warn().Err(err).Msgf("Failed to remove filters from %s %s", iface, direction)
			}
		}

		// Apply new applies
		for _, a := range newApply.([]interface{}) {
			applyMap := a.(map[string]interface{})
			iface := applyMap["interface"].(string)
			direction := strings.ToLower(applyMap["direction"].(string))
			filterIDs := extractIPFilterIDsFromApply(applyMap, d)

			if len(filterIDs) > 0 {
				if err := apiClient.client.ApplyIPFiltersToInterface(ctx, iface, direction, filterIDs); err != nil {
					return diag.Errorf("Failed to apply filters to interface %s %s: %v", iface, direction, err)
				}
			}
		}
	}

	return resourceRTXAccessListIPRead(ctx, d, meta)
}

func resourceRTXAccessListIPDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	name := d.Id()

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ip", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ip").Msgf("Deleting IP access list group: %s", name)

	// First remove apply blocks to free up filter references
	if v, ok := d.GetOk("apply"); ok {
		for _, a := range v.([]interface{}) {
			applyMap := a.(map[string]interface{})
			iface := applyMap["interface"].(string)
			direction := strings.ToLower(applyMap["direction"].(string))

			if err := apiClient.client.RemoveIPFiltersFromInterface(ctx, iface, direction); err != nil {
				logger.Warn().Err(err).Msgf("Failed to remove filters from %s %s", iface, direction)
			}
		}
	}

	// Get sequences to delete
	sequences := getIPExpectedSequences(d)

	// Delete all entries
	for _, seq := range sequences {
		if err := apiClient.client.DeleteIPFilter(ctx, seq); err != nil {
			if !strings.Contains(err.Error(), "not found") {
				return diag.Errorf("Failed to delete IP filter %d: %v", seq, err)
			}
		}
	}

	return nil
}

func resourceRTXAccessListIPImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Import format: name:seq1,seq2,seq3 or just name (imports all matching)
	// For simplicity, we'll import by name and assume manual sequence mode
	importID := d.Id()
	parts := strings.Split(importID, ":")

	name := parts[0]
	var sequences []int

	if len(parts) > 1 {
		// Parse comma-separated sequence numbers
		seqStrs := strings.Split(parts[1], ",")
		for _, s := range seqStrs {
			var seq int
			if _, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &seq); err == nil {
				sequences = append(sequences, seq)
			}
		}
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ip").Msgf("Importing IP access list: %s with sequences %v", name, sequences)

	d.SetId(name)
	if err := d.Set("name", name); err != nil {
		return nil, err
	}

	// If sequences provided, import those specific ones
	if len(sequences) > 0 {
		apiClient := meta.(*apiClient)
		entries := make([]map[string]interface{}, 0, len(sequences))

		for _, seq := range sequences {
			filter, err := apiClient.client.GetIPFilter(ctx, seq)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					continue
				}
				return nil, fmt.Errorf("failed to read IP filter %d: %v", seq, err)
			}

			e := map[string]interface{}{
				"sequence":    filter.Number,
				"action":      filter.Action,
				"source":      filter.SourceAddress,
				"destination": filter.DestAddress,
				"protocol":    filter.Protocol,
				"source_port": normalizePort(filter.SourcePort),
				"dest_port":   normalizePort(filter.DestPort),
				"established": filter.Established,
				"log":         false,
			}
			entries = append(entries, e)
		}

		if len(entries) == 0 {
			return nil, fmt.Errorf("no IP filters found with sequences %v", sequences)
		}

		if err := d.Set("entry", entries); err != nil {
			return nil, fmt.Errorf("failed to set state: %v", err)
		}
	}

	return []*schema.ResourceData{d}, nil
}

// buildIPFiltersFromResourceData builds IP filter structs from Terraform resource data.
func buildIPFiltersFromResourceData(d *schema.ResourceData) []client.IPFilter {
	sequenceStart := d.Get("sequence_start").(int)
	sequenceStep := d.Get("sequence_step").(int)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	entries := d.Get("entry").([]interface{})
	result := make([]client.IPFilter, 0, len(entries))

	for i, e := range entries {
		entry := e.(map[string]interface{})

		// Determine sequence
		var seq int
		if sequenceStart > 0 {
			// Auto mode
			seq = sequenceStart + (i * sequenceStep)
		} else {
			// Manual mode
			seq = entry["sequence"].(int)
		}

		filter := client.IPFilter{
			Number:        seq,
			Action:        entry["action"].(string),
			SourceAddress: entry["source"].(string),
			DestAddress:   entry["destination"].(string),
			Protocol:      entry["protocol"].(string),
			SourcePort:    entry["source_port"].(string),
			DestPort:      entry["dest_port"].(string),
			Established:   entry["established"].(bool),
		}

		result = append(result, filter)
	}

	return result
}

// getIPExpectedSequences returns the sequence numbers expected based on state.
func getIPExpectedSequences(d *schema.ResourceData) []int {
	sequenceStart := d.Get("sequence_start").(int)
	sequenceStep := d.Get("sequence_step").(int)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	entries := d.Get("entry").([]interface{})
	sequences := make([]int, 0, len(entries))

	for i, e := range entries {
		entry := e.(map[string]interface{})

		var seq int
		if sequenceStart > 0 {
			seq = sequenceStart + (i * sequenceStep)
		} else {
			seq = entry["sequence"].(int)
		}

		if seq > 0 {
			sequences = append(sequences, seq)
		}
	}

	return sequences
}

// applyIPFiltersToInterfacesFromResourceData handles the apply blocks during create.
func applyIPFiltersToInterfacesFromResourceData(ctx context.Context, d *schema.ResourceData, apiClient *apiClient) error {
	v, ok := d.GetOk("apply")
	if !ok {
		return nil
	}

	for _, a := range v.([]interface{}) {
		applyMap := a.(map[string]interface{})
		iface := applyMap["interface"].(string)
		direction := strings.ToLower(applyMap["direction"].(string))
		filterIDs := extractIPFilterIDsFromApply(applyMap, d)

		if len(filterIDs) > 0 {
			if err := apiClient.client.ApplyIPFiltersToInterface(ctx, iface, direction, filterIDs); err != nil {
				return fmt.Errorf("failed to apply filters to interface %s %s: %w", iface, direction, err)
			}
		}
	}

	return nil
}

// readIPApplyBlocksFromRouter reads apply block state from the router.
func readIPApplyBlocksFromRouter(ctx context.Context, d *schema.ResourceData, apiClient *apiClient) error {
	v, ok := d.GetOk("apply")
	if !ok {
		return nil
	}

	applyList := v.([]interface{})
	updatedApplies := make([]map[string]interface{}, 0, len(applyList))

	for _, a := range applyList {
		applyMap := a.(map[string]interface{})
		iface := applyMap["interface"].(string)
		direction := strings.ToLower(applyMap["direction"].(string))

		filterIDs, err := apiClient.client.GetIPInterfaceFilters(ctx, iface, direction)
		if err != nil {
			return fmt.Errorf("failed to get filters for interface %s %s: %w", iface, direction, err)
		}

		updatedApply := map[string]interface{}{
			"interface":  iface,
			"direction":  direction,
			"filter_ids": filterIDs,
		}
		updatedApplies = append(updatedApplies, updatedApply)
	}

	return d.Set("apply", updatedApplies)
}

// extractIPFilterIDsFromApply extracts filter IDs from apply map, falling back to entry sequences.
func extractIPFilterIDsFromApply(applyMap map[string]interface{}, d *schema.ResourceData) []int {
	if rawIDs, ok := applyMap["filter_ids"].([]interface{}); ok && len(rawIDs) > 0 {
		ids := make([]int, 0, len(rawIDs))
		for _, id := range rawIDs {
			ids = append(ids, id.(int))
		}
		return ids
	}

	// Fall back to all entry sequences
	return getIPExpectedSequences(d)
}

// extractIPSequencesFromEntryList extracts sequences from an entry list.
func extractIPSequencesFromEntryList(entries []interface{}, d *schema.ResourceData) []int {
	sequenceStart := d.Get("sequence_start").(int)
	sequenceStep := d.Get("sequence_step").(int)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	sequences := make([]int, 0, len(entries))

	for i, e := range entries {
		entry := e.(map[string]interface{})

		var seq int
		if sequenceStart > 0 {
			seq = sequenceStart + (i * sequenceStep)
		} else {
			seq = entry["sequence"].(int)
		}

		if seq > 0 {
			sequences = append(sequences, seq)
		}
	}

	return sequences
}

// findIPRemovedSequences finds sequences that were in old but not in new.
func findIPRemovedSequences(old, new []int) []int {
	newSet := make(map[int]bool)
	for _, seq := range new {
		newSet[seq] = true
	}

	var removed []int
	for _, seq := range old {
		if !newSet[seq] {
			removed = append(removed, seq)
		}
	}

	return removed
}

// normalizePort converts empty or missing port to "*"
func normalizePort(port string) string {
	if port == "" {
		return "*"
	}
	return port
}
