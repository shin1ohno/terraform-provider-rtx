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

// resourceRTXAccessListMAC returns the schema for the rtx_access_list_mac resource
func resourceRTXAccessListMAC() *schema.Resource {
	return &schema.Resource{
		Description: "Manages MAC address access lists on RTX routers. " +
			"MAC ACLs filter traffic based on source and destination MAC addresses. " +
			"Supports automatic sequence numbering (auto mode) or manual sequence assignment.",

		CreateContext: resourceRTXAccessListMACCreate,
		ReadContext:   resourceRTXAccessListMACRead,
		UpdateContext: resourceRTXAccessListMACUpdate,
		DeleteContext: resourceRTXAccessListMACDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXAccessListMACImport,
		},

		CustomizeDiff: customdiff.All(
			validateMACACLSchema,
		),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Access list name (identifier)",
			},
			"filter_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				Description:  "Optional RTX filter ID to enable numeric ethernet filter mode. If not specified, derived from first entry.",
				ValidateFunc: validation.IntAtLeast(1),
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
				Computed:    true,
				Description: "List of interface bindings. Each apply block binds this ACL to an interface in a specific direction. Multiple apply blocks are supported.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"interface": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Interface to apply filters (e.g., lan1, bridge1). MAC ACLs cannot be applied to PP or Tunnel interfaces.",
						},
						"direction": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Direction to apply filters (in or out)",
							ValidateFunc:     validation.StringInSlice([]string{"in", "out"}, true),
							DiffSuppressFunc: SuppressCaseDiff,
						},
						"filter_ids": {
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							Description: "Specific filter IDs (sequence numbers) to apply in order. If omitted, all entry sequences are applied in order.",
							Elem: &schema.Schema{
								Type:         schema.TypeInt,
								ValidateFunc: validation.IntAtLeast(1),
							},
						},
					},
				},
			},
			"entry": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of MAC ACL entries",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sequence": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							Description:  "Sequence number (determines order of evaluation). Required in manual mode (when sequence_start is not set). Auto-calculated in auto mode (when sequence_start is set).",
							ValidateFunc: validation.IntBetween(1, MaxSequence),
						},
						"ace_action": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Action to take (permit/deny or RTX pass/reject with log/nolog)",
							ValidateFunc:     validation.StringInSlice([]string{"permit", "deny", "pass-log", "pass-nolog", "reject-log", "reject-nolog", "pass", "reject"}, true),
							DiffSuppressFunc: SuppressEquivalentACLActionDiff,
						},
						"source_any": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Match any source MAC address",
						},
						"source_address": {
							Type:             schema.TypeString,
							Optional:         true,
							Description:      "Source MAC address (e.g., 00:00:00:00:00:00)",
							DiffSuppressFunc: SuppressMACAddressWhenAnyIsTrue,
						},
						"source_address_mask": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source MAC wildcard mask",
						},
						"destination_any": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Match any destination MAC address",
						},
						"destination_address": {
							Type:             schema.TypeString,
							Optional:         true,
							Description:      "Destination MAC address (e.g., 00:00:00:00:00:00)",
							DiffSuppressFunc: SuppressMACAddressWhenAnyIsTrue,
						},
						"destination_address_mask": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination MAC wildcard mask",
						},
						"ether_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Ethernet type (e.g., 0x0800 for IPv4, 0x0806 for ARP)",
						},
						"vlan_id": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "VLAN ID to match",
							ValidateFunc: validation.IntBetween(1, 4094),
						},
						"log": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Enable logging for this entry",
						},
						"filter_id": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Explicit filter number for this entry (overrides sequence)",
							ValidateFunc: validation.IntAtLeast(1),
						},
						"dhcp_match": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "DHCP-based match settings",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										Description:  "DHCP match type (dhcp-bind or dhcp-not-bind)",
										ValidateFunc: validation.StringInSlice([]string{"dhcp-bind", "dhcp-not-bind"}, false),
									},
									"scope": {
										Type:         schema.TypeInt,
										Optional:     true,
										Description:  "DHCP scope number",
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"offset": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Offset for byte matching",
							ValidateFunc: validation.IntAtLeast(0),
						},
						"byte_list": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Byte list (hex) for offset matching",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

// validateMACACLSchema validates the MAC ACL schema for auto/manual mode consistency
// and validates apply block interface compatibility.
func validateMACACLSchema(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	// Validate sequence mode consistency
	sequenceStart, hasSequenceStart := diff.GetOk("sequence_start")
	sequenceStep := diff.Get("sequence_step").(int)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	entries := diff.Get("entry").([]interface{})
	autoMode := hasSequenceStart && sequenceStart.(int) > 0

	// Track sequences for duplicate detection
	usedSequences := make(map[int]int)

	for i, e := range entries {
		entry := e.(map[string]interface{})
		entrySeq, hasEntrySeq := entry["sequence"]
		entrySeqVal := 0
		if hasEntrySeq {
			if seq, ok := entrySeq.(int); ok {
				entrySeqVal = seq
			}
		}

		if autoMode {
			// Auto mode: entry-level sequence should not be specified
			if entrySeqVal > 0 {
				return fmt.Errorf("entry[%d]: sequence cannot be specified when sequence_start is set (auto mode). Remove the sequence attribute or use manual mode by removing sequence_start", i)
			}

			// Calculate the sequence for overflow check
			calculatedSeq := sequenceStart.(int) + (i * sequenceStep)
			if calculatedSeq > MaxSequenceValue {
				return fmt.Errorf("entry[%d]: calculated sequence %d exceeds maximum value %d. Reduce sequence_start or sequence_step, or reduce number of entries", i, calculatedSeq, MaxSequenceValue)
			}

			// Check for duplicates
			if prevIdx, exists := usedSequences[calculatedSeq]; exists {
				return fmt.Errorf("entry[%d]: calculated sequence %d conflicts with entry[%d]. Increase sequence_step to avoid collisions", i, calculatedSeq, prevIdx)
			}
			usedSequences[calculatedSeq] = i
		} else {
			// Manual mode: entry-level sequence is required
			if entrySeqVal <= 0 {
				return fmt.Errorf("entry[%d]: sequence must be specified when sequence_start is not set (manual mode). Add a sequence attribute to each entry or use auto mode by setting sequence_start", i)
			}

			// Check for duplicates in manual mode
			if prevIdx, exists := usedSequences[entrySeqVal]; exists {
				return fmt.Errorf("entry[%d]: sequence %d is already used by entry[%d]. Each entry must have a unique sequence number", i, entrySeqVal, prevIdx)
			}
			usedSequences[entrySeqVal] = i
		}
	}

	// Validate apply blocks
	applies, hasApplies := diff.GetOk("apply")
	if !hasApplies {
		return nil
	}

	applyList := applies.([]interface{})
	appliedTo := make(map[string]int)

	for i, a := range applyList {
		applyMap := a.(map[string]interface{})
		iface := applyMap["interface"].(string)
		direction := applyMap["direction"].(string)

		// Validate interface compatibility for MAC ACLs
		if err := InterfaceSupportsACLType(iface, ACLTypeMAC); err != nil {
			return fmt.Errorf("apply[%d]: %v", i, err)
		}

		// Check for duplicate interface+direction
		key := fmt.Sprintf("%s:%s", iface, strings.ToLower(direction))
		if prevIdx, exists := appliedTo[key]; exists {
			return fmt.Errorf("apply[%d]: interface %s direction %s is already specified in apply[%d]. Remove the duplicate apply block", i, iface, direction, prevIdx)
		}
		appliedTo[key] = i

		// Validate filter_ids if specified
		if filterIDs, ok := applyMap["filter_ids"].([]interface{}); ok && len(filterIDs) > 0 {
			seenIDs := make(map[int]bool)
			for j, id := range filterIDs {
				filterID := id.(int)
				if seenIDs[filterID] {
					return fmt.Errorf("apply[%d].filter_ids[%d]: filter ID %d is duplicated. Remove duplicate filter IDs", i, j, filterID)
				}
				seenIDs[filterID] = true
			}
		}
	}

	return nil
}

func resourceRTXAccessListMACCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_mac", d.Id())
	acl := buildAccessListMACFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_mac").Msgf("Creating MAC access list: %s", acl.Name)

	err := apiClient.client.CreateAccessListMAC(ctx, acl)
	if err != nil {
		return diag.Errorf("Failed to create MAC access list: %v", err)
	}

	d.SetId(acl.Name)

	return resourceRTXAccessListMACRead(ctx, d, meta)
}

func resourceRTXAccessListMACRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_mac", d.Id())
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_mac").Msgf("Reading MAC access list: %s", name)

	acl, err := apiClient.client.GetAccessListMAC(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logging.FromContext(ctx).Warn().Str("resource", "rtx_access_list_mac").Msgf("MAC access list %s not found, removing from state", name)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read MAC access list: %v", err)
	}

	d.Set("name", acl.Name)
	d.Set("filter_id", acl.FilterID)

	// Set sequence_start and sequence_step if they were in the config
	if acl.SequenceStart > 0 {
		d.Set("sequence_start", acl.SequenceStart)
		if acl.SequenceStep > 0 {
			d.Set("sequence_step", acl.SequenceStep)
		}
	}

	entries := make([]map[string]interface{}, 0, len(acl.Entries))
	wildcardMAC := "*:*:*:*:*:*"
	for _, entry := range acl.Entries {
		// Detect wildcard addresses and set *_any fields accordingly
		sourceAny := entry.SourceAny || entry.SourceAddress == wildcardMAC
		destinationAny := entry.DestinationAny || entry.DestinationAddress == wildcardMAC

		// When *_any is true, clear the address to match the config pattern
		sourceAddress := entry.SourceAddress
		destinationAddress := entry.DestinationAddress
		if sourceAny && sourceAddress == wildcardMAC {
			sourceAddress = ""
		}
		if destinationAny && destinationAddress == wildcardMAC {
			destinationAddress = ""
		}

		e := map[string]interface{}{
			"sequence":                 entry.Sequence,
			"ace_action":               entry.AceAction,
			"source_any":               sourceAny,
			"source_address":           sourceAddress,
			"source_address_mask":      entry.SourceAddressMask,
			"destination_any":          destinationAny,
			"destination_address":      destinationAddress,
			"destination_address_mask": entry.DestinationAddressMask,
			"ether_type":               entry.EtherType,
			"vlan_id":                  entry.VlanID,
			"log":                      entry.Log,
			"filter_id":                entry.FilterID,
			"offset":                   entry.Offset,
			"byte_list":                entry.ByteList,
		}

		if entry.DHCPType != "" {
			e["dhcp_match"] = []map[string]interface{}{
				{
					"type":  entry.DHCPType,
					"scope": entry.DHCPScope,
				},
			}
		}

		entries = append(entries, e)
	}
	d.Set("entry", entries)

	// Handle multiple applies (prefer Applies over legacy Apply)
	if len(acl.Applies) > 0 {
		applyList := make([]map[string]interface{}, 0, len(acl.Applies))
		for _, apply := range acl.Applies {
			applyList = append(applyList, map[string]interface{}{
				"interface":  apply.Interface,
				"direction":  apply.Direction,
				"filter_ids": apply.FilterIDs,
			})
		}
		d.Set("apply", applyList)
	} else if acl.Apply != nil {
		// Legacy single apply support
		d.Set("apply", []map[string]interface{}{
			{
				"interface":  acl.Apply.Interface,
				"direction":  acl.Apply.Direction,
				"filter_ids": acl.Apply.FilterIDs,
			},
		})
	} else {
		d.Set("apply", nil)
	}

	return nil
}

func resourceRTXAccessListMACUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_mac", d.Id())
	acl := buildAccessListMACFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_mac").Msgf("Updating MAC access list: %s", acl.Name)

	err := apiClient.client.UpdateAccessListMAC(ctx, acl)
	if err != nil {
		return diag.Errorf("Failed to update MAC access list: %v", err)
	}

	return resourceRTXAccessListMACRead(ctx, d, meta)
}

func resourceRTXAccessListMACDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_access_list_mac", d.Id())
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_mac").Msgf("Deleting MAC access list: %s", name)

	// Collect filter numbers to delete
	var filterNums []int
	sequenceStart := d.Get("sequence_start").(int)
	sequenceStep := d.Get("sequence_step").(int)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	entries := d.Get("entry").([]interface{})
	for i, e := range entries {
		entry := e.(map[string]interface{})
		var num int

		// Determine filter number based on mode
		if sequenceStart > 0 {
			// Auto mode: calculate sequence
			num = sequenceStart + (i * sequenceStep)
		} else {
			// Manual mode: use explicit filter_id or sequence
			num = entry["filter_id"].(int)
			if num == 0 {
				num = entry["sequence"].(int)
			}
		}

		if num > 0 {
			filterNums = append(filterNums, num)
		}
	}

	err := apiClient.client.DeleteAccessListMAC(ctx, name, filterNums)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete MAC access list: %v", err)
	}

	return nil
}

func resourceRTXAccessListMACImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	name := d.Id()

	logging.FromContext(ctx).Debug().Str("resource", "rtx_access_list_mac").Msgf("Importing MAC access list: %s", name)

	acl, err := apiClient.client.GetAccessListMAC(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to import MAC access list %s: %v", name, err)
	}

	d.SetId(name)
	d.Set("name", acl.Name)
	d.Set("filter_id", acl.FilterID)

	// Import sequence_start and sequence_step if present
	if acl.SequenceStart > 0 {
		d.Set("sequence_start", acl.SequenceStart)
		if acl.SequenceStep > 0 {
			d.Set("sequence_step", acl.SequenceStep)
		}
	}

	entries := make([]map[string]interface{}, 0, len(acl.Entries))
	for _, entry := range acl.Entries {
		e := map[string]interface{}{
			"sequence":                 entry.Sequence,
			"ace_action":               entry.AceAction,
			"source_any":               entry.SourceAny,
			"source_address":           entry.SourceAddress,
			"source_address_mask":      entry.SourceAddressMask,
			"destination_any":          entry.DestinationAny,
			"destination_address":      entry.DestinationAddress,
			"destination_address_mask": entry.DestinationAddressMask,
			"ether_type":               entry.EtherType,
			"vlan_id":                  entry.VlanID,
			"log":                      entry.Log,
			"filter_id":                entry.FilterID,
			"offset":                   entry.Offset,
			"byte_list":                entry.ByteList,
		}
		if entry.DHCPType != "" {
			e["dhcp_match"] = []map[string]interface{}{
				{
					"type":  entry.DHCPType,
					"scope": entry.DHCPScope,
				},
			}
		}
		entries = append(entries, e)
	}
	d.Set("entry", entries)

	// Import multiple applies
	if len(acl.Applies) > 0 {
		applyList := make([]map[string]interface{}, 0, len(acl.Applies))
		for _, apply := range acl.Applies {
			applyList = append(applyList, map[string]interface{}{
				"interface":  apply.Interface,
				"direction":  apply.Direction,
				"filter_ids": apply.FilterIDs,
			})
		}
		d.Set("apply", applyList)
	} else if acl.Apply != nil {
		d.Set("apply", []map[string]interface{}{
			{
				"interface":  acl.Apply.Interface,
				"direction":  acl.Apply.Direction,
				"filter_ids": acl.Apply.FilterIDs,
			},
		})
	} else {
		d.Set("apply", nil)
	}

	return []*schema.ResourceData{d}, nil
}

func buildAccessListMACFromResourceData(d *schema.ResourceData) client.AccessListMAC {
	sequenceStart := d.Get("sequence_start").(int)
	sequenceStep := d.Get("sequence_step").(int)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	acl := client.AccessListMAC{
		Name:          d.Get("name").(string),
		FilterID:      d.Get("filter_id").(int),
		SequenceStart: sequenceStart,
		SequenceStep:  sequenceStep,
		Entries:       make([]client.AccessListMACEntry, 0),
		Applies:       make([]client.MACApply, 0),
	}

	entries := d.Get("entry").([]interface{})
	for i, e := range entries {
		entry := e.(map[string]interface{})

		// Determine sequence based on mode
		var entrySequence int
		if sequenceStart > 0 {
			// Auto mode: calculate sequence
			entrySequence = sequenceStart + (i * sequenceStep)
		} else {
			// Manual mode: use explicit sequence
			entrySequence = entry["sequence"].(int)
		}

		aclEntry := client.AccessListMACEntry{
			Sequence:               entrySequence,
			AceAction:              entry["ace_action"].(string),
			SourceAny:              entry["source_any"].(bool),
			SourceAddress:          entry["source_address"].(string),
			SourceAddressMask:      entry["source_address_mask"].(string),
			DestinationAny:         entry["destination_any"].(bool),
			DestinationAddress:     entry["destination_address"].(string),
			DestinationAddressMask: entry["destination_address_mask"].(string),
			EtherType:              entry["ether_type"].(string),
			VlanID:                 entry["vlan_id"].(int),
			Log:                    entry["log"].(bool),
			FilterID:               entry["filter_id"].(int),
			Offset:                 entry["offset"].(int),
		}

		if aclEntry.FilterID == 0 && acl.FilterID > 0 {
			aclEntry.FilterID = acl.FilterID
		}

		if v, ok := entry["byte_list"].([]interface{}); ok {
			for _, b := range v {
				aclEntry.ByteList = append(aclEntry.ByteList, b.(string))
			}
		}

		if v, ok := entry["dhcp_match"].([]interface{}); ok && len(v) > 0 {
			m := v[0].(map[string]interface{})
			aclEntry.DHCPType = m["type"].(string)
			aclEntry.DHCPScope = m["scope"].(int)
		}

		acl.Entries = append(acl.Entries, aclEntry)
	}

	// Build applies list
	if v, ok := d.GetOk("apply"); ok {
		applyList := v.([]interface{})
		for _, a := range applyList {
			m := a.(map[string]interface{})
			var ids []int
			if rawIDs, ok := m["filter_ids"].([]interface{}); ok {
				for _, id := range rawIDs {
					ids = append(ids, id.(int))
				}
			}

			// If filter_ids is empty, populate with all entry sequences
			if len(ids) == 0 {
				for _, entry := range acl.Entries {
					ids = append(ids, entry.Sequence)
				}
			}

			apply := client.MACApply{
				Interface: m["interface"].(string),
				Direction: strings.ToLower(m["direction"].(string)),
				FilterIDs: ids,
			}
			acl.Applies = append(acl.Applies, apply)
		}

		// Also set legacy Apply field for backward compatibility with client layer
		if len(acl.Applies) > 0 {
			acl.Apply = &acl.Applies[0]
		}
	}

	return acl
}
