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

// resourceRTXAccessListIPv6 returns the schema for the rtx_access_list_ipv6 resource.
// This is a group-based resource where multiple IPv6 filter entries are managed together
// under a single name identifier.
func resourceRTXAccessListIPv6() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a group of IPv6 static filters (access list) on RTX routers. " +
			"This resource manages multiple IPv6 filter rules as a single group using the RTX native 'ipv6 filter' command. " +
			"Supports automatic sequence numbering or manual sequence assignment.",

		CreateContext: resourceRTXAccessListIPv6Create,
		ReadContext:   resourceRTXAccessListIPv6Read,
		UpdateContext: resourceRTXAccessListIPv6Update,
		DeleteContext: resourceRTXAccessListIPv6Delete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXAccessListIPv6Import,
		},

		CustomizeDiff: customdiff.All(
			ValidateACLSchema,
			validateAccessListIPv6Entries,
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
				Description: "List of IPv6 filter entries. Each entry defines a single filter rule.",
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
							Description: "Source IPv6 address/prefix (e.g., '2001:db8::/32') or '*' for any",
						},
						"destination": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Destination IPv6 address/prefix (e.g., '2001:db8::1/128') or '*' for any",
						},
						"protocol": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "*",
							Description:      "Protocol: tcp, udp, icmp6, ip, gre, esp, ah, or * for any",
							ValidateFunc:     validation.StringInSlice([]string{"tcp", "udp", "icmp6", "ip", "gre", "esp", "ah", "*"}, true),
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

// validateAccessListIPv6Entries validates IPv6 filter entry constraints.
func validateAccessListIPv6Entries(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	entries := diff.Get("entry").([]interface{})

	for i, e := range entries {
		entry := e.(map[string]interface{})

		protocol := strings.ToLower(entry["protocol"].(string))
		sourcePort := entry["source_port"].(string)
		destPort := entry["dest_port"].(string)

		// Port specifications only valid for TCP/UDP
		if protocol != "tcp" && protocol != "udp" {
			if sourcePort != "*" && sourcePort != "" {
				return fmt.Errorf("entry[%d]: source_port can only be specified for tcp or udp protocols", i)
			}
			if destPort != "*" && destPort != "" {
				return fmt.Errorf("entry[%d]: dest_port can only be specified for tcp or udp protocols", i)
			}
		}
	}

	return nil
}

func resourceRTXAccessListIPv6Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	name := d.Get("name").(string)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ipv6").Msgf("Creating IPv6 access list group: %s", name)

	// Build and create IPv6 filters
	filters := buildIPv6FiltersFromResourceData(d)
	for _, filter := range filters {
		if err := apiClient.client.CreateIPv6Filter(ctx, filter); err != nil {
			return diag.Errorf("Failed to create IPv6 filter %d: %v", filter.Number, err)
		}
	}

	// Handle apply blocks
	if err := applyIPv6FiltersToInterfacesFromResourceData(ctx, d, apiClient); err != nil {
		return diag.Errorf("Failed to apply IPv6 filters to interfaces: %v", err)
	}

	d.SetId(name)

	return resourceRTXAccessListIPv6Read(ctx, d, meta)
}

func resourceRTXAccessListIPv6Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	name := d.Id()

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ipv6").Msgf("Reading IPv6 access list group: %s", name)

	// Get the expected sequences from state to query
	sequences := getIPv6ExpectedSequences(d)

	// Read each filter
	entries := make([]map[string]interface{}, 0, len(sequences))
	foundAny := false
	for _, seq := range sequences {
		filter, err := apiClient.client.GetIPv6Filter(ctx, seq)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				continue
			}
			return diag.Errorf("Failed to read IPv6 filter %d: %v", seq, err)
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
			"log":         false, // RTX doesn't return log status in filter read
		}
		entries = append(entries, e)
	}

	// If no entries found and we expected some, mark as deleted
	if !foundAny && len(sequences) > 0 {
		logger.Debug().Str("resource", "rtx_access_list_ipv6").Msgf("IPv6 access list %s not found, removing from state", name)
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
	if err := readIPv6ApplyBlocksFromRouter(ctx, d, apiClient); err != nil {
		logger.Warn().Err(err).Msg("Failed to read apply blocks")
	}

	return nil
}

func resourceRTXAccessListIPv6Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	name := d.Id()

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ipv6").Msgf("Updating IPv6 access list group: %s", name)

	// Handle entry changes
	if d.HasChange("entry") || d.HasChange("sequence_start") || d.HasChange("sequence_step") {
		// Get old sequences to potentially delete
		oldEntries, _ := d.GetChange("entry")
		oldSequences := extractIPv6SequencesFromEntryList(oldEntries.([]interface{}), d)
		newSequences := getIPv6ExpectedSequences(d)

		// Delete removed sequences
		toDelete := findIPv6RemovedSequences(oldSequences, newSequences)
		for _, seq := range toDelete {
			if err := apiClient.client.DeleteIPv6Filter(ctx, seq); err != nil {
				if !strings.Contains(err.Error(), "not found") {
					logger.Warn().Err(err).Msgf("Failed to delete IPv6 filter %d", seq)
				}
			}
		}

		// Create/update filters
		filters := buildIPv6FiltersFromResourceData(d)
		for _, filter := range filters {
			if err := apiClient.client.UpdateIPv6Filter(ctx, filter); err != nil {
				return diag.Errorf("Failed to update IPv6 filter %d: %v", filter.Number, err)
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

			if err := apiClient.client.RemoveIPv6FiltersFromInterface(ctx, iface, direction); err != nil {
				logger.Warn().Err(err).Msgf("Failed to remove filters from %s %s", iface, direction)
			}
		}

		// Apply new applies
		for _, a := range newApply.([]interface{}) {
			applyMap := a.(map[string]interface{})
			iface := applyMap["interface"].(string)
			direction := strings.ToLower(applyMap["direction"].(string))
			filterIDs := extractIPv6FilterIDsFromApply(applyMap, d)

			if len(filterIDs) > 0 {
				if err := apiClient.client.ApplyIPv6FiltersToInterface(ctx, iface, direction, filterIDs); err != nil {
					return diag.Errorf("Failed to apply filters to interface %s %s: %v", iface, direction, err)
				}
			}
		}
	}

	return resourceRTXAccessListIPv6Read(ctx, d, meta)
}

func resourceRTXAccessListIPv6Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)
	name := d.Id()

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_ipv6", name)
	logger := logging.FromContext(ctx)

	logger.Debug().Str("resource", "rtx_access_list_ipv6").Msgf("Deleting IPv6 access list group: %s", name)

	// First remove apply blocks to free up filter references
	if v, ok := d.GetOk("apply"); ok {
		for _, a := range v.([]interface{}) {
			applyMap := a.(map[string]interface{})
			iface := applyMap["interface"].(string)
			direction := strings.ToLower(applyMap["direction"].(string))

			if err := apiClient.client.RemoveIPv6FiltersFromInterface(ctx, iface, direction); err != nil {
				logger.Warn().Err(err).Msgf("Failed to remove filters from %s %s", iface, direction)
			}
		}
	}

	// Get sequences to delete
	sequences := getIPv6ExpectedSequences(d)

	// Delete all entries
	for _, seq := range sequences {
		if err := apiClient.client.DeleteIPv6Filter(ctx, seq); err != nil {
			if !strings.Contains(err.Error(), "not found") {
				return diag.Errorf("Failed to delete IPv6 filter %d: %v", seq, err)
			}
		}
	}

	return nil
}

func resourceRTXAccessListIPv6Import(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_ipv6").Msgf("Importing IPv6 access list: %s with sequences %v", name, sequences)

	d.SetId(name)
	if err := d.Set("name", name); err != nil {
		return nil, err
	}

	// If sequences provided, import those specific ones
	if len(sequences) > 0 {
		apiClient := meta.(*apiClient)
		entries := make([]map[string]interface{}, 0, len(sequences))

		for _, seq := range sequences {
			filter, err := apiClient.client.GetIPv6Filter(ctx, seq)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					continue
				}
				return nil, fmt.Errorf("failed to read IPv6 filter %d: %v", seq, err)
			}

			e := map[string]interface{}{
				"sequence":    filter.Number,
				"action":      filter.Action,
				"source":      filter.SourceAddress,
				"destination": filter.DestAddress,
				"protocol":    filter.Protocol,
				"source_port": normalizePort(filter.SourcePort),
				"dest_port":   normalizePort(filter.DestPort),
				"log":         false,
			}
			entries = append(entries, e)
		}

		if len(entries) == 0 {
			return nil, fmt.Errorf("no IPv6 filters found with sequences %v", sequences)
		}

		if err := d.Set("entry", entries); err != nil {
			return nil, fmt.Errorf("failed to set state: %v", err)
		}
	}

	return []*schema.ResourceData{d}, nil
}

// buildIPv6FiltersFromResourceData builds IPv6 filter structs from Terraform resource data.
func buildIPv6FiltersFromResourceData(d *schema.ResourceData) []client.IPFilter {
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
		}

		result = append(result, filter)
	}

	return result
}

// getIPv6ExpectedSequences returns the sequence numbers expected based on state.
func getIPv6ExpectedSequences(d *schema.ResourceData) []int {
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

// applyIPv6FiltersToInterfacesFromResourceData handles the apply blocks during create.
func applyIPv6FiltersToInterfacesFromResourceData(ctx context.Context, d *schema.ResourceData, apiClient *apiClient) error {
	v, ok := d.GetOk("apply")
	if !ok {
		return nil
	}

	for _, a := range v.([]interface{}) {
		applyMap := a.(map[string]interface{})
		iface := applyMap["interface"].(string)
		direction := strings.ToLower(applyMap["direction"].(string))
		filterIDs := extractIPv6FilterIDsFromApply(applyMap, d)

		if len(filterIDs) > 0 {
			if err := apiClient.client.ApplyIPv6FiltersToInterface(ctx, iface, direction, filterIDs); err != nil {
				return fmt.Errorf("failed to apply filters to interface %s %s: %w", iface, direction, err)
			}
		}
	}

	return nil
}

// readIPv6ApplyBlocksFromRouter reads apply block state from the router.
func readIPv6ApplyBlocksFromRouter(ctx context.Context, d *schema.ResourceData, apiClient *apiClient) error {
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

		filterIDs, err := apiClient.client.GetIPv6InterfaceFilters(ctx, iface, direction)
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

// extractIPv6FilterIDsFromApply extracts filter IDs from apply map, falling back to entry sequences.
func extractIPv6FilterIDsFromApply(applyMap map[string]interface{}, d *schema.ResourceData) []int {
	if rawIDs, ok := applyMap["filter_ids"].([]interface{}); ok && len(rawIDs) > 0 {
		ids := make([]int, 0, len(rawIDs))
		for _, id := range rawIDs {
			ids = append(ids, id.(int))
		}
		return ids
	}

	// Fall back to all entry sequences
	return getIPv6ExpectedSequences(d)
}

// extractIPv6SequencesFromEntryList extracts sequences from an entry list.
func extractIPv6SequencesFromEntryList(entries []interface{}, d *schema.ResourceData) []int {
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

// findIPv6RemovedSequences finds sequences that were in old but not in new.
func findIPv6RemovedSequences(old, new []int) []int {
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
