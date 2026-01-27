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

func resourceRTXAccessListExtended() *schema.Resource {
	return &schema.Resource{
		Description: "Manages IPv4 extended access lists (ACLs) on RTX routers. Extended ACLs provide granular control over packet filtering based on source/destination addresses, protocols, and ports. " +
			"Supports both manual sequence mode (explicit sequence on each entry) and auto sequence mode (sequence_start + sequence_step). " +
			"Optional apply blocks bind the ACL to interfaces.",
		CreateContext: resourceRTXAccessListExtendedCreate,
		ReadContext:   resourceRTXAccessListExtendedRead,
		UpdateContext: resourceRTXAccessListExtendedUpdate,
		DeleteContext: resourceRTXAccessListExtendedDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXAccessListExtendedImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the access list (used as identifier)",
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
				Description: "List of ACL entries",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sequence": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							Description:  "Sequence number (determines order). Required in manual mode (when sequence_start is not set). Auto-calculated in auto mode.",
							ValidateFunc: validation.IntBetween(1, MaxSequenceValue),
						},
						"ace_rule_action": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Action: 'permit' or 'deny'",
							ValidateFunc:     validation.StringInSlice([]string{"permit", "deny"}, true),
							DiffSuppressFunc: SuppressCaseDiff, // ACL actions are case-insensitive
						},
						"ace_rule_protocol": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Protocol: tcp, udp, icmp, ip, gre, esp, ah, or *",
							ValidateFunc:     validation.StringInSlice([]string{"tcp", "udp", "icmp", "ip", "gre", "esp", "ah", "*"}, true),
							DiffSuppressFunc: SuppressCaseDiff, // Protocol names are case-insensitive
						},
						"source_any": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Match any source address",
						},
						"source_prefix": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source IP address (e.g., '192.168.1.0')",
						},
						"source_prefix_mask": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source wildcard mask (e.g., '0.0.0.255')",
						},
						"source_port_equal": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source port equals (e.g., '80', '443')",
						},
						"source_port_range": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source port range (e.g., '1024-65535')",
						},
						"destination_any": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Match any destination address",
						},
						"destination_prefix": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination IP address (e.g., '10.0.0.0')",
						},
						"destination_prefix_mask": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination wildcard mask (e.g., '0.0.0.255')",
						},
						"destination_port_equal": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination port equals (e.g., '80', '443')",
						},
						"destination_port_range": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination port range (e.g., '1024-65535')",
						},
						"established": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Match established TCP connections (ACK or RST flag set)",
						},
						"log": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Enable logging for this entry",
						},
					},
				},
			},
		},

		CustomizeDiff: customdiff.All(
			validateAccessListExtendedEntries,
			ValidateACLSchema,
		),
	}
}

func validateAccessListExtendedEntries(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	entries := diff.Get("entry").([]interface{})

	for i, e := range entries {
		entry := e.(map[string]interface{})

		sourceAny := entry["source_any"].(bool)
		sourcePrefix := entry["source_prefix"].(string)
		destAny := entry["destination_any"].(bool)
		destPrefix := entry["destination_prefix"].(string)
		protocol := entry["ace_rule_protocol"].(string)
		established := entry["established"].(bool)

		// Either source_any or source_prefix must be specified
		if !sourceAny && sourcePrefix == "" {
			return fmt.Errorf("entry[%d]: either source_any must be true or source_prefix must be specified", i)
		}

		// Either destination_any or destination_prefix must be specified
		if !destAny && destPrefix == "" {
			return fmt.Errorf("entry[%d]: either destination_any must be true or destination_prefix must be specified", i)
		}

		// Established is only valid for TCP
		if established && strings.ToLower(protocol) != "tcp" {
			return fmt.Errorf("entry[%d]: established can only be set to true for tcp protocol", i)
		}
	}

	return nil
}

func resourceRTXAccessListExtendedCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_extended", d.Id())
	acl := buildAccessListExtendedFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended").Msgf("Creating access list extended: %+v", acl)

	err := apiClient.client.CreateAccessListExtended(ctx, acl)
	if err != nil {
		return diag.Errorf("Failed to create access list extended: %v", err)
	}

	d.SetId(acl.Name)

	// Handle apply blocks
	if len(acl.Applies) > 0 {
		if err := applyExtendedFiltersToInterfaces(ctx, apiClient, acl); err != nil {
			return diag.Errorf("Failed to apply filters to interfaces: %v", err)
		}
	}

	return resourceRTXAccessListExtendedRead(ctx, d, meta)
}

func resourceRTXAccessListExtendedRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_extended", d.Id())
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended").Msgf("Reading access list extended: %s", name)

	acl, err := apiClient.client.GetAccessListExtended(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended").Msgf("Access list extended %s not found, removing from state", name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read access list extended: %v", err)
	}

	if err := d.Set("name", acl.Name); err != nil {
		return diag.FromErr(err)
	}

	// Preserve sequence_start and sequence_step from config (they're not stored on router)
	// The values are already in state from the config

	entries := flattenAccessListExtendedEntries(acl.Entries)
	if err := d.Set("entry", entries); err != nil {
		return diag.FromErr(err)
	}

	// Read apply blocks from router
	applies, err := readExtendedApplies(ctx, apiClient, acl)
	if err != nil {
		logging.FromContext(ctx).Warn().Err(err).Msg("Failed to read interface filters, apply state may be stale")
	} else if len(applies) > 0 {
		if err := d.Set("apply", applies); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceRTXAccessListExtendedUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_extended", d.Id())
	acl := buildAccessListExtendedFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended").Msgf("Updating access list extended: %+v", acl)

	// Update entries
	err := apiClient.client.UpdateAccessListExtended(ctx, acl)
	if err != nil {
		return diag.Errorf("Failed to update access list extended: %v", err)
	}

	// Handle apply block changes
	if d.HasChange("apply") {
		oldApplies, newApplies := d.GetChange("apply")
		if err := updateExtendedApplies(ctx, apiClient, acl, oldApplies.([]interface{}), newApplies.([]interface{})); err != nil {
			return diag.Errorf("Failed to update interface filters: %v", err)
		}
	}

	return resourceRTXAccessListExtendedRead(ctx, d, meta)
}

func resourceRTXAccessListExtendedDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_extended", d.Id())
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended").Msgf("Deleting access list extended: %s", name)

	// Remove applies first
	acl := buildAccessListExtendedFromResourceData(d)
	if len(acl.Applies) > 0 {
		if err := removeExtendedFiltersFromInterfaces(ctx, apiClient, acl); err != nil {
			logging.FromContext(ctx).Warn().Err(err).Msg("Failed to remove filters from interfaces before delete")
		}
	}

	err := apiClient.client.DeleteAccessListExtended(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete access list extended: %v", err)
	}

	return nil
}

func resourceRTXAccessListExtendedImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_extended").Msgf("Importing access list extended: %s", name)

	acl, err := apiClient.client.GetAccessListExtended(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to import access list extended %s: %v", name, err)
	}

	d.SetId(name)
	d.Set("name", acl.Name)

	entries := flattenAccessListExtendedEntries(acl.Entries)
	d.Set("entry", entries)

	// Read apply blocks from router
	applies, err := readExtendedApplies(ctx, apiClient, acl)
	if err != nil {
		logging.FromContext(ctx).Warn().Err(err).Msg("Failed to read interface filters during import")
	} else if len(applies) > 0 {
		d.Set("apply", applies)
	}

	return []*schema.ResourceData{d}, nil
}

func buildAccessListExtendedFromResourceData(d *schema.ResourceData) client.AccessListExtended {
	sequenceStart := d.Get("sequence_start").(int)
	sequenceStep := d.Get("sequence_step").(int)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	acl := client.AccessListExtended{
		Name:    d.Get("name").(string),
		Entries: expandAccessListExtendedEntriesWithSequence(d.Get("entry").([]interface{}), sequenceStart, sequenceStep),
		Applies: expandExtendedApplies(d.Get("apply").([]interface{})),
	}
	return acl
}

func expandAccessListExtendedEntriesWithSequence(entries []interface{}, sequenceStart, sequenceStep int) []client.AccessListExtendedEntry {
	result := make([]client.AccessListExtendedEntry, 0, len(entries))

	for i, e := range entries {
		entry := e.(map[string]interface{})

		// Determine sequence number
		var sequence int
		if sequenceStart > 0 {
			// Auto mode: calculate sequence
			sequence = sequenceStart + (i * sequenceStep)
		} else {
			// Manual mode: use explicit sequence
			sequence = entry["sequence"].(int)
		}

		aclEntry := client.AccessListExtendedEntry{
			Sequence:        sequence,
			AceRuleAction:   entry["ace_rule_action"].(string),
			AceRuleProtocol: entry["ace_rule_protocol"].(string),
			SourceAny:       entry["source_any"].(bool),
			DestinationAny:  entry["destination_any"].(bool),
			Established:     entry["established"].(bool),
			Log:             entry["log"].(bool),
		}

		if v, ok := entry["source_prefix"].(string); ok && v != "" {
			aclEntry.SourcePrefix = v
		}
		if v, ok := entry["source_prefix_mask"].(string); ok && v != "" {
			aclEntry.SourcePrefixMask = v
		}
		if v, ok := entry["source_port_equal"].(string); ok && v != "" {
			aclEntry.SourcePortEqual = v
		}
		if v, ok := entry["source_port_range"].(string); ok && v != "" {
			aclEntry.SourcePortRange = v
		}
		if v, ok := entry["destination_prefix"].(string); ok && v != "" {
			aclEntry.DestinationPrefix = v
		}
		if v, ok := entry["destination_prefix_mask"].(string); ok && v != "" {
			aclEntry.DestinationPrefixMask = v
		}
		if v, ok := entry["destination_port_equal"].(string); ok && v != "" {
			aclEntry.DestinationPortEqual = v
		}
		if v, ok := entry["destination_port_range"].(string); ok && v != "" {
			aclEntry.DestinationPortRange = v
		}

		result = append(result, aclEntry)
	}

	return result
}

func flattenAccessListExtendedEntries(entries []client.AccessListExtendedEntry) []interface{} {
	result := make([]interface{}, 0, len(entries))

	for _, entry := range entries {
		e := map[string]interface{}{
			"sequence":                entry.Sequence,
			"ace_rule_action":         entry.AceRuleAction,
			"ace_rule_protocol":       entry.AceRuleProtocol,
			"source_any":              entry.SourceAny,
			"source_prefix":           entry.SourcePrefix,
			"source_prefix_mask":      entry.SourcePrefixMask,
			"source_port_equal":       entry.SourcePortEqual,
			"source_port_range":       entry.SourcePortRange,
			"destination_any":         entry.DestinationAny,
			"destination_prefix":      entry.DestinationPrefix,
			"destination_prefix_mask": entry.DestinationPrefixMask,
			"destination_port_equal":  entry.DestinationPortEqual,
			"destination_port_range":  entry.DestinationPortRange,
			"established":             entry.Established,
			"log":                     entry.Log,
		}
		result = append(result, e)
	}

	return result
}

func expandExtendedApplies(applies []interface{}) []client.ExtendedApply {
	result := make([]client.ExtendedApply, 0, len(applies))

	for _, a := range applies {
		applyMap := a.(map[string]interface{})
		apply := client.ExtendedApply{
			Interface: applyMap["interface"].(string),
			Direction: applyMap["direction"].(string),
		}

		// Extract filter_ids if specified
		if filterIDs, ok := applyMap["filter_ids"].([]interface{}); ok {
			for _, id := range filterIDs {
				apply.FilterIDs = append(apply.FilterIDs, id.(int))
			}
		}

		result = append(result, apply)
	}

	return result
}

// applyExtendedFiltersToInterfaces applies filters to interfaces using Client interface
// Extended ACL uses IP filters, so we use the IP filter apply methods
func applyExtendedFiltersToInterfaces(ctx context.Context, apiClient *apiClient, acl client.AccessListExtended) error {
	for _, apply := range acl.Applies {
		filterIDs := apply.FilterIDs
		// If no filter_ids specified, use all entry sequences
		if len(filterIDs) == 0 {
			for _, entry := range acl.Entries {
				filterIDs = append(filterIDs, entry.Sequence)
			}
		}

		err := apiClient.client.ApplyIPFiltersToInterface(ctx, apply.Interface, apply.Direction, filterIDs)
		if err != nil {
			return fmt.Errorf("failed to apply filters to %s %s: %w", apply.Interface, apply.Direction, err)
		}
	}

	return nil
}

// removeExtendedFiltersFromInterfaces removes filters from interfaces
func removeExtendedFiltersFromInterfaces(ctx context.Context, apiClient *apiClient, acl client.AccessListExtended) error {
	for _, apply := range acl.Applies {
		err := apiClient.client.RemoveIPFiltersFromInterface(ctx, apply.Interface, apply.Direction)
		if err != nil {
			return fmt.Errorf("failed to remove filters from %s %s: %w", apply.Interface, apply.Direction, err)
		}
	}

	return nil
}

// updateExtendedApplies handles changes to apply blocks
func updateExtendedApplies(ctx context.Context, apiClient *apiClient, acl client.AccessListExtended, oldApplies, newApplies []interface{}) error {
	// Build maps for comparison
	oldMap := make(map[string]client.ExtendedApply)
	for _, a := range expandExtendedApplies(oldApplies) {
		key := fmt.Sprintf("%s:%s", a.Interface, a.Direction)
		oldMap[key] = a
	}

	newMap := make(map[string]client.ExtendedApply)
	for _, a := range expandExtendedApplies(newApplies) {
		key := fmt.Sprintf("%s:%s", a.Interface, a.Direction)
		newMap[key] = a
	}

	// Remove old applies that are not in new
	for key, oldApply := range oldMap {
		if _, exists := newMap[key]; !exists {
			err := apiClient.client.RemoveIPFiltersFromInterface(ctx, oldApply.Interface, oldApply.Direction)
			if err != nil {
				return fmt.Errorf("failed to remove filters from %s %s: %w", oldApply.Interface, oldApply.Direction, err)
			}
		}
	}

	// Add or update new applies
	for key, newApply := range newMap {
		filterIDs := newApply.FilterIDs
		// If no filter_ids specified, use all entry sequences
		if len(filterIDs) == 0 {
			for _, entry := range acl.Entries {
				filterIDs = append(filterIDs, entry.Sequence)
			}
		}

		oldApply, exists := oldMap[key]
		if !exists || !equalIntSlices(oldApply.FilterIDs, filterIDs) {
			// Apply or update
			err := apiClient.client.ApplyIPFiltersToInterface(ctx, newApply.Interface, newApply.Direction, filterIDs)
			if err != nil {
				return fmt.Errorf("failed to apply filters to %s %s: %w", newApply.Interface, newApply.Direction, err)
			}
		}
	}

	return nil
}

// readExtendedApplies reads the current apply state from the router
// Note: Since the router stores filters by sequence number (not by ACL name),
// we read all interface filters and try to match them with our entry sequences.
// This approach has limitations - it cannot distinguish between filters from
// different ACL resources using the same sequence numbers.
func readExtendedApplies(ctx context.Context, apiClient *apiClient, acl *client.AccessListExtended) ([]interface{}, error) {
	// Build set of our entry sequences for filtering
	ourSequences := make(map[int]bool)
	for _, entry := range acl.Entries {
		ourSequences[entry.Sequence] = true
	}

	// Read applies from the configuration we maintain in state
	// rather than trying to reverse-engineer from router config
	// This is because the router doesn't track which ACL resource owns which filter
	result := make([]interface{}, 0)
	for _, apply := range acl.Applies {
		filterIDs := apply.FilterIDs
		if len(filterIDs) == 0 {
			// If no explicit filter_ids, use all sequences
			for _, entry := range acl.Entries {
				filterIDs = append(filterIDs, entry.Sequence)
			}
		}

		// Verify each filter is still applied by querying the router
		currentFilters, err := apiClient.client.GetIPInterfaceFilters(ctx, apply.Interface, apply.Direction)
		if err != nil {
			// If we can't read filters, skip this apply
			continue
		}

		// Find matching filter IDs
		currentSet := make(map[int]bool)
		for _, id := range currentFilters {
			currentSet[id] = true
		}

		matchingIDs := make([]int, 0)
		for _, id := range filterIDs {
			if currentSet[id] && ourSequences[id] {
				matchingIDs = append(matchingIDs, id)
			}
		}

		if len(matchingIDs) > 0 {
			applyMap := map[string]interface{}{
				"interface":  apply.Interface,
				"direction":  apply.Direction,
				"filter_ids": matchingIDs,
			}
			result = append(result, applyMap)
		}
	}

	return result, nil
}

// equalIntSlices compares two int slices for equality
func equalIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
