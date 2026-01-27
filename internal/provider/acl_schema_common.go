// Package provider contains ACL schema common utilities for Terraform provider.
//
// This file provides shared schema definitions and helper functions for all ACL
// (Access Control List) resources in the terraform-provider-rtx. It establishes
// a unified pattern for ACL definition across IP, IPv6, MAC, and Extended ACL types.
//
// The unified ACL design supports:
//   - Automatic sequence calculation based on entry order (sequence_start + sequence_step)
//   - Manual sequence mode where each entry specifies its own sequence number
//   - Multiple apply blocks for binding ACLs to interfaces
//   - Consistent validation across all ACL types
//
// This file works in conjunction with sequence_calculator.go which provides:
//   - SequenceMode type (AutoMode, ManualMode, MixedMode)
//   - CalculateSequences() for computing sequence numbers
//   - ValidateSequenceRange() for overflow detection
//   - DetectSequenceMode() for mode detection
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// ACLType represents the type of ACL for command generation and validation.
type ACLType string

const (
	// ACLTypeIP represents IPv4 access lists.
	ACLTypeIP ACLType = "ip"
	// ACLTypeIPv6 represents IPv6 access lists.
	ACLTypeIPv6 ACLType = "ipv6"
	// ACLTypeMAC represents MAC address access lists.
	ACLTypeMAC ACLType = "mac"
	// ACLTypeIPDynamic represents dynamic IPv4 access lists.
	ACLTypeIPDynamic ACLType = "ip_dynamic"
	// ACLTypeIPv6Dynamic represents dynamic IPv6 access lists.
	ACLTypeIPv6Dynamic ACLType = "ipv6_dynamic"
	// ACLTypeExtended represents extended access lists.
	ACLTypeExtended ACLType = "extended"
)

// MaxSequenceValue is the maximum valid sequence number for RTX filters.
// RTX routers support sequence numbers from 1 to 65535.
// Reference: RTX Command Reference Chapter 13 (Ethernet Filtering), Chapter 14 (IP Packet Filtering)
// This is used for schema validation. The MaxSequence constant in sequence_calculator.go
// is used for calculation validation.
const MaxSequenceValue = 65535

// CommonACL represents the common fields extracted from ACL resource data.
type CommonACL struct {
	Name          string
	SequenceStart int // 0 = manual mode
	SequenceStep  int // default 1
	Entries       []ACLEntry
	Applies       []ACLApply
}

// ACLEntry represents a single entry in an ACL with its sequence and fields.
type ACLEntry struct {
	Sequence int                    // calculated or explicit sequence number
	Fields   map[string]interface{} // ACL-type specific fields
}

// ACLApply represents an interface binding configuration for an ACL.
type ACLApply struct {
	Interface string // e.g., "lan1", "bridge1", "pp1"
	Direction string // "in" or "out"
	FilterIDs []int  // empty = all sequences
}

// CommonACLSchema returns shared schema attributes for all ACL types.
//
// This function provides the common top-level attributes that every ACL resource
// should include: name, sequence_start, sequence_step, and apply blocks.
//
// Schema attributes:
//   - name: ACL group identifier (required, immutable)
//   - sequence_start: Starting sequence for auto mode (optional)
//   - sequence_step: Increment for auto mode (optional, default 10)
//   - apply: List of interface bindings (optional)
//
// Usage:
//
//	schema := map[string]*schema.Schema{
//	    ...CommonACLSchema(),
//	    "entry": { /* ACL-type specific entry schema */ },
//	}
func CommonACLSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "ACL group identifier. This name is used to reference the ACL in other resources.",
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
	}
}

// CommonApplySchema returns the schema for apply blocks.
//
// The apply block schema defines how an ACL is bound to a network interface.
// Each apply block specifies the interface, direction, and optionally which
// filter IDs to apply.
//
// Schema attributes:
//   - interface: Target interface name (required)
//   - direction: Traffic direction "in" or "out" (required)
//   - filter_ids: Specific filter IDs to apply (optional, defaults to all)
//
// Returns a *schema.Resource suitable for use as an Elem in a TypeList schema.
func CommonApplySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"interface": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Interface to apply the ACL to (e.g., lan1, bridge1, pp1, tunnel1).",
			},
			"direction": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Direction to apply the ACL: 'in' for incoming traffic, 'out' for outgoing traffic.",
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
					ValidateFunc: validation.IntBetween(1, MaxSequenceValue),
				},
			},
		},
	}
}

// CommonEntrySchema returns base schema for entry blocks.
//
// This provides the common attributes shared by all ACL entry types.
// ACL-specific fields should be added to this base by the individual
// resource implementations.
//
// Schema attributes:
//   - sequence: Entry sequence number (optional in auto mode, required in manual mode)
//   - log: Enable logging for this entry (optional)
//
// Usage:
//
//	entrySchema := CommonEntrySchema()
//	entrySchema["ace_action"] = &schema.Schema{ /* ACL-specific */ }
//	entrySchema["source_any"] = &schema.Schema{ /* ACL-specific */ }
func CommonEntrySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"sequence": {
			Type:         schema.TypeInt,
			Optional:     true,
			Computed:     true,
			Description:  "Sequence number determines the order of evaluation. Required when sequence_start is not set (manual mode). Auto-calculated when sequence_start is set (auto mode).",
			ValidateFunc: validation.IntBetween(1, MaxSequenceValue),
		},
		"log": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enable logging when this entry matches traffic.",
		},
	}
}

// BuildACLFromResourceData extracts common ACL fields from ResourceData.
//
// This helper function reads the common ACL attributes from a Terraform
// ResourceData object and returns a CommonACL struct. It handles both
// auto and manual sequence modes.
//
// Parameters:
//   - d: The Terraform ResourceData containing the ACL configuration
//
// Returns:
//   - CommonACL struct populated with name, sequence settings, entries, and applies
//
// Note: This function only extracts the common fields. ACL-type specific
// entry fields must be processed separately by the resource implementation.
func BuildACLFromResourceData(d *schema.ResourceData) CommonACL {
	acl := CommonACL{
		Name:          d.Get("name").(string),
		SequenceStart: d.Get("sequence_start").(int),
		SequenceStep:  d.Get("sequence_step").(int),
		Entries:       make([]ACLEntry, 0),
		Applies:       make([]ACLApply, 0),
	}

	// Handle default sequence step (use DefaultSequenceStep from sequence_calculator.go)
	if acl.SequenceStep == 0 {
		acl.SequenceStep = DefaultSequenceStep
	}

	// Extract entries
	if v, ok := d.GetOk("entry"); ok {
		entries := v.([]interface{})
		for i, e := range entries {
			entry := e.(map[string]interface{})
			aclEntry := ACLEntry{
				Fields: entry,
			}

			// Determine sequence based on mode
			if acl.SequenceStart > 0 {
				// Auto mode: calculate sequence
				aclEntry.Sequence = acl.SequenceStart + (i * acl.SequenceStep)
			} else if seq, ok := entry["sequence"].(int); ok && seq > 0 {
				// Manual mode: use explicit sequence
				aclEntry.Sequence = seq
			}

			acl.Entries = append(acl.Entries, aclEntry)
		}
	}

	// Extract apply blocks
	if v, ok := d.GetOk("apply"); ok {
		applyList := v.([]interface{})
		for _, a := range applyList {
			applyMap := a.(map[string]interface{})
			apply := ACLApply{
				Interface: applyMap["interface"].(string),
				Direction: applyMap["direction"].(string),
			}

			// Extract filter_ids if specified
			if filterIDs, ok := applyMap["filter_ids"].([]interface{}); ok {
				for _, id := range filterIDs {
					apply.FilterIDs = append(apply.FilterIDs, id.(int))
				}
			}

			// If filter_ids is empty, populate with all entry sequences
			if len(apply.FilterIDs) == 0 {
				for _, entry := range acl.Entries {
					apply.FilterIDs = append(apply.FilterIDs, entry.Sequence)
				}
			}

			acl.Applies = append(acl.Applies, apply)
		}
	}

	return acl
}

// ValidateACLSchema is a CustomizeDiff validation function that validates
// the ACL schema for auto/manual mode consistency.
//
// This function should be used in the CustomizeDiff field of ACL resources
// to ensure that:
//  1. When sequence_start is set (auto mode), entries do not have explicit sequence
//  2. When sequence_start is not set (manual mode), all entries have explicit sequence
//  3. Sequence values are within valid range after calculation
//  4. No duplicate sequences exist
//
// Usage:
//
//	&schema.Resource{
//	    CustomizeDiff: customdiff.All(
//	        ValidateACLSchema,
//	        // other validations...
//	    ),
//	}
//
// Returns nil if validation passes, or an error describing the validation failure.
func ValidateACLSchema(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	sequenceStart, hasSequenceStart := diff.GetOk("sequence_start")
	sequenceStep := diff.Get("sequence_step").(int)
	// Use DefaultSequenceStep from sequence_calculator.go
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	entries := diff.Get("entry").([]interface{})

	// Determine the sequence mode
	autoMode := hasSequenceStart && sequenceStart.(int) > 0

	// Track sequences for duplicate detection
	usedSequences := make(map[int]int) // sequence -> entry index

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

			// Check for duplicates (in auto mode, this shouldn't happen unless step=0)
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
	return validateApplyBlocks(diff)
}

// validateApplyBlocks validates the apply blocks for consistency.
func validateApplyBlocks(diff *schema.ResourceDiff) error {
	applies, ok := diff.GetOk("apply")
	if !ok {
		return nil
	}

	applyList := applies.([]interface{})

	// Track interface+direction combinations for duplicate detection
	appliedTo := make(map[string]int) // "interface:direction" -> apply index

	for i, a := range applyList {
		applyMap := a.(map[string]interface{})
		iface := applyMap["interface"].(string)
		direction := applyMap["direction"].(string)

		key := fmt.Sprintf("%s:%s", iface, direction)
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

// Note: DetectSequenceMode is defined in sequence_calculator.go and returns
// AutoMode, ManualMode, or MixedMode based on the ACL configuration.
// Use that function to detect the sequence mode from ResourceData.

// CalculateEntrySequences calculates sequence numbers for all entries.
//
// This function computes the sequence numbers for ACL entries based on the
// configured mode (auto or manual). In auto mode, sequences are calculated
// from sequence_start and sequence_step. In manual mode, the explicit
// sequence values are used.
//
// Parameters:
//   - acl: The CommonACL containing configuration and entries
//
// Returns:
//   - A slice of calculated sequence numbers in entry order
//
// Note: This function does not validate the sequences. Use ValidateACLSchema
// in CustomizeDiff to ensure sequences are valid before this function is called.
func CalculateEntrySequences(acl CommonACL) []int {
	sequences := make([]int, len(acl.Entries))

	for i, entry := range acl.Entries {
		if acl.SequenceStart > 0 {
			// Auto mode: calculate sequence
			sequences[i] = acl.SequenceStart + (i * acl.SequenceStep)
		} else {
			// Manual mode: use entry's sequence
			sequences[i] = entry.Sequence
		}
	}

	return sequences
}

// InterfaceSupportsACLType checks if an interface type supports a given ACL type.
//
// This validation function determines whether a specific ACL type can be applied
// to a given interface. For example, MAC ACLs cannot be applied to PP or Tunnel
// interfaces.
//
// Parameters:
//   - interfaceName: The name of the interface (e.g., "lan1", "pp1", "tunnel1")
//   - aclType: The type of ACL being applied
//
// Returns:
//   - nil if the ACL type is supported on the interface
//   - error describing the incompatibility if not supported
func InterfaceSupportsACLType(interfaceName string, aclType ACLType) error {
	// Determine interface type from name
	isPP := len(interfaceName) > 2 && interfaceName[:2] == "pp"
	isTunnel := len(interfaceName) > 6 && interfaceName[:6] == "tunnel"

	// MAC ACLs are not supported on PP and Tunnel interfaces
	if aclType == ACLTypeMAC && (isPP || isTunnel) {
		return fmt.Errorf("MAC ACL cannot be applied to %s interfaces. MAC filtering is only supported on LAN and Bridge interfaces", interfaceName)
	}

	return nil
}
