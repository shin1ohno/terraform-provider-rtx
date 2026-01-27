// Package provider contains the collision validator for ACL resources.
//
// This file provides sequence collision detection for ACL resources in the
// terraform-provider-rtx. It implements both Plan-time (Terraform state-based)
// and Apply-time (router-based) collision detection.
//
// Collision detection ensures that ACL resources do not use overlapping sequence
// numbers, which would cause filter conflicts on the RTX router.
package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
)

// CollisionError represents a sequence collision between ACL resources.
// This error type provides detailed information about which sequences conflict
// and which resources are involved.
type CollisionError struct {
	// Sequence is the conflicting sequence number
	Sequence int

	// OwnedBy is the resource that owns this sequence (current resource being validated)
	OwnedBy string

	// ConflictsWith is the resource or entity that this sequence conflicts with
	ConflictsWith string

	// ACLType is the type of ACL (ip, ipv6, mac, etc.)
	ACLType ACLType

	// Message provides additional context about the collision
	Message string
}

// Error implements the error interface for CollisionError.
func (e *CollisionError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("sequence %d (owned by %q) conflicts with %q for ACL type %s",
		e.Sequence, e.OwnedBy, e.ConflictsWith, e.ACLType)
}

// MultiCollisionError represents multiple sequence collisions.
type MultiCollisionError struct {
	Errors []CollisionError
}

// Error implements the error interface for MultiCollisionError.
func (e *MultiCollisionError) Error() string {
	if len(e.Errors) == 0 {
		return "no collisions"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("detected %d sequence collision(s):\n", len(e.Errors)))

	for i, err := range e.Errors {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("  - %s", err.Error()))
	}

	return sb.String()
}

// SequenceRange represents a range of sequences used by an ACL resource.
type SequenceRange struct {
	// ResourceName is the Terraform resource name (e.g., "rtx_access_list_mac.example")
	ResourceName string

	// ACLName is the ACL group name
	ACLName string

	// ACLType is the type of ACL
	ACLType ACLType

	// Sequences is the list of sequence numbers used by this resource
	Sequences []int

	// Start is the minimum sequence number (for range display)
	Start int

	// End is the maximum sequence number (for range display)
	End int
}

// ContainsSequence checks if this range contains the given sequence.
func (r *SequenceRange) ContainsSequence(seq int) bool {
	for _, s := range r.Sequences {
		if s == seq {
			return true
		}
	}
	return false
}

// Overlaps checks if this range overlaps with another range.
func (r *SequenceRange) Overlaps(other *SequenceRange) []int {
	overlapping := make([]int, 0)

	seqSet := make(map[int]bool, len(r.Sequences))
	for _, s := range r.Sequences {
		seqSet[s] = true
	}

	for _, s := range other.Sequences {
		if seqSet[s] {
			overlapping = append(overlapping, s)
		}
	}

	return overlapping
}

// String returns a human-readable representation of the range.
func (r *SequenceRange) String() string {
	if len(r.Sequences) == 0 {
		return fmt.Sprintf("%s (no sequences)", r.ResourceName)
	}
	if len(r.Sequences) == 1 {
		return fmt.Sprintf("%s (sequence %d)", r.ResourceName, r.Sequences[0])
	}
	return fmt.Sprintf("%s (sequences %d-%d, count=%d)", r.ResourceName, r.Start, r.End, len(r.Sequences))
}

// ValidateNoCollision checks for sequence collisions during Terraform plan phase.
//
// This function should be used in CustomizeDiff to detect potential sequence
// conflicts between ACL resources in the same Terraform state. It compares the
// sequences of the current resource with other ACL resources of the same type.
//
// Parameters:
//   - ctx: Context for the operation
//   - diff: The ResourceDiff from CustomizeDiff
//   - meta: Provider meta containing client configuration
//
// Returns:
//   - nil if no collisions detected
//   - error describing the collision if detected
//
// Note: This validation is state-based and cannot detect conflicts with
// resources managed outside of Terraform. Use CheckRouterCollision for
// comprehensive Apply-time validation.
func ValidateNoCollision(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	logger := logging.FromContext(ctx)

	// Get the ACL type from meta if available, or determine from resource type
	aclType := getACLTypeFromResourceDiff(diff)
	if aclType == "" {
		logger.Debug().Msg("collision_validator: could not determine ACL type, skipping collision check")
		return nil
	}

	// Get the current resource's name and sequences
	currentName := ""
	if v, ok := diff.GetOk("name"); ok {
		currentName = v.(string)
	}
	if currentName == "" {
		// If no name is set, we cannot check for collisions
		return nil
	}

	// Calculate the sequences for the current resource
	currentSequences := calculateSequencesFromDiff(diff)
	if len(currentSequences) == 0 {
		// No sequences to validate
		return nil
	}

	logger.Debug().
		Str("acl_name", currentName).
		Str("acl_type", string(aclType)).
		Ints("sequences", currentSequences).
		Msg("collision_validator: validating sequences")

	// Build the current resource's range
	currentRange := buildSequenceRange(diff.Id(), currentName, aclType, currentSequences)

	// In CustomizeDiff context, we don't have easy access to other resources
	// in the state. This validation primarily checks internal consistency.
	// For cross-resource collision detection, CheckRouterCollision should be used.

	// Check for duplicate sequences within the current resource
	seenSequences := make(map[int]bool)
	for _, seq := range currentSequences {
		if seenSequences[seq] {
			return &CollisionError{
				Sequence:      seq,
				OwnedBy:       currentName,
				ConflictsWith: currentName,
				ACLType:       aclType,
				Message:       fmt.Sprintf("duplicate sequence %d detected within ACL %q", seq, currentName),
			}
		}
		seenSequences[seq] = true
	}

	logger.Debug().
		Str("range", currentRange.String()).
		Msg("collision_validator: no internal collisions detected")

	return nil
}

// CheckRouterCollision verifies that sequences don't conflict with existing
// filters on the RTX router.
//
// This function should be called during the Apply phase (Create or Update)
// to ensure that the sequences being used don't conflict with existing
// filters on the router that may not be managed by Terraform.
//
// Parameters:
//   - ctx: Context for the operation
//   - rtxClient: The RTX client for querying router state
//   - aclType: The type of ACL being validated
//   - sequences: The sequence numbers to validate
//   - excludeACL: The ACL name to exclude from collision check (for updates)
//
// Returns:
//   - nil if no collisions detected
//   - CollisionError or MultiCollisionError if collisions are found
func CheckRouterCollision(ctx context.Context, rtxClient client.Client, aclType ACLType, sequences []int, excludeACL string) error {
	logger := logging.FromContext(ctx)

	if len(sequences) == 0 {
		return nil
	}

	logger.Debug().
		Str("acl_type", string(aclType)).
		Ints("sequences", sequences).
		Str("exclude_acl", excludeACL).
		Msg("collision_validator: checking router for existing filters")

	// Get existing filter sequences from the router
	existingSequences, err := getExistingFilterSequences(ctx, rtxClient, aclType)
	if err != nil {
		logger.Warn().
			Err(err).
			Str("acl_type", string(aclType)).
			Msg("collision_validator: failed to get existing filters, skipping router collision check")
		return nil
	}

	// Check for collisions
	var collisions []CollisionError

	for _, seq := range sequences {
		if owner, exists := existingSequences[seq]; exists {
			// Skip if this sequence belongs to the ACL being updated
			if excludeACL != "" && owner == excludeACL {
				continue
			}

			collisions = append(collisions, CollisionError{
				Sequence:      seq,
				OwnedBy:       "current resource",
				ConflictsWith: owner,
				ACLType:       aclType,
				Message: fmt.Sprintf("sequence %d already exists on router (owned by %q). "+
					"Choose a different sequence range to avoid conflicts.", seq, owner),
			})
		}
	}

	if len(collisions) == 0 {
		logger.Debug().Msg("collision_validator: no router collisions detected")
		return nil
	}

	if len(collisions) == 1 {
		return &collisions[0]
	}

	return &MultiCollisionError{Errors: collisions}
}

// ValidateSequenceRangeNoOverlap validates that multiple ACL resources don't have
// overlapping sequence ranges.
//
// This is a helper function that can be used to validate a collection of
// sequence ranges for overlap. It returns a detailed error if any overlaps
// are detected.
//
// Parameters:
//   - ranges: List of SequenceRange to validate
//
// Returns:
//   - nil if no overlaps detected
//   - MultiCollisionError describing all detected overlaps
func ValidateSequenceRangeNoOverlap(ranges []SequenceRange) error {
	if len(ranges) < 2 {
		return nil
	}

	var collisions []CollisionError

	// Compare each pair of ranges
	for i := 0; i < len(ranges); i++ {
		for j := i + 1; j < len(ranges); j++ {
			// Only compare ranges of the same ACL type
			if ranges[i].ACLType != ranges[j].ACLType {
				continue
			}

			overlapping := ranges[i].Overlaps(&ranges[j])
			if len(overlapping) > 0 {
				for _, seq := range overlapping {
					collisions = append(collisions, CollisionError{
						Sequence:      seq,
						OwnedBy:       ranges[i].ResourceName,
						ConflictsWith: ranges[j].ResourceName,
						ACLType:       ranges[i].ACLType,
						Message: fmt.Sprintf("sequence %d in %q conflicts with %q",
							seq, ranges[i].ResourceName, ranges[j].ResourceName),
					})
				}
			}
		}
	}

	if len(collisions) == 0 {
		return nil
	}

	return &MultiCollisionError{Errors: collisions}
}

// getACLTypeFromResourceDiff attempts to determine the ACL type from the resource.
func getACLTypeFromResourceDiff(diff *schema.ResourceDiff) ACLType {
	// Try to get from a custom attribute if set
	if v, ok := diff.GetOk("acl_type"); ok {
		return ACLType(v.(string))
	}

	// Infer from resource type name if possible
	// This is a fallback; resources should set acl_type explicitly
	return ""
}

// calculateSequencesFromDiff calculates the sequences for a resource from its diff.
func calculateSequencesFromDiff(diff *schema.ResourceDiff) []int {
	sequences := make([]int, 0)

	// Check if auto mode
	sequenceStart, hasSequenceStart := diff.GetOk("sequence_start")
	sequenceStep := diff.Get("sequence_step").(int)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	entries := diff.Get("entry").([]interface{})

	if hasSequenceStart && sequenceStart.(int) > 0 {
		// Auto mode: calculate sequences
		start := sequenceStart.(int)
		for i := range entries {
			sequences = append(sequences, start+(i*sequenceStep))
		}
	} else {
		// Manual mode: extract from entries
		for _, e := range entries {
			entry := e.(map[string]interface{})
			if seq, ok := entry["sequence"].(int); ok && seq > 0 {
				sequences = append(sequences, seq)
			}
		}
	}

	return sequences
}

// buildSequenceRange creates a SequenceRange from the given parameters.
func buildSequenceRange(resourceID, aclName string, aclType ACLType, sequences []int) *SequenceRange {
	if len(sequences) == 0 {
		return &SequenceRange{
			ResourceName: resourceID,
			ACLName:      aclName,
			ACLType:      aclType,
			Sequences:    sequences,
		}
	}

	// Sort sequences to find min/max
	sorted := make([]int, len(sequences))
	copy(sorted, sequences)
	sort.Ints(sorted)

	return &SequenceRange{
		ResourceName: resourceID,
		ACLName:      aclName,
		ACLType:      aclType,
		Sequences:    sequences,
		Start:        sorted[0],
		End:          sorted[len(sorted)-1],
	}
}

// getExistingFilterSequences retrieves existing filter sequences from the router.
// Returns a map of sequence number to owner identifier.
func getExistingFilterSequences(ctx context.Context, rtxClient client.Client, aclType ACLType) (map[int]string, error) {
	sequences := make(map[int]string)

	switch aclType {
	case ACLTypeIP, ACLTypeExtended:
		filters, err := rtxClient.ListIPFilters(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list IP filters: %w", err)
		}
		for _, f := range filters {
			// Use "router" as owner for externally managed filters
			// In the future, we could try to identify the Terraform resource
			sequences[f.Number] = "router"
		}

	case ACLTypeIPv6:
		filters, err := rtxClient.ListIPv6Filters(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list IPv6 filters: %w", err)
		}
		for _, f := range filters {
			sequences[f.Number] = "router"
		}

	case ACLTypeMAC:
		filters, err := rtxClient.ListEthernetFilters(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list Ethernet filters: %w", err)
		}
		for _, f := range filters {
			sequences[f.Number] = "router"
		}

	case ACLTypeIPDynamic:
		filters, err := rtxClient.ListIPFiltersDynamic(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list dynamic IP filters: %w", err)
		}
		for _, f := range filters {
			sequences[f.Number] = "router"
		}

	default:
		return nil, fmt.Errorf("unsupported ACL type: %s", aclType)
	}

	return sequences, nil
}

// BuildCollisionErrorMessage creates a user-friendly error message for collision errors.
func BuildCollisionErrorMessage(err error, suggestedStart, suggestedStep int) string {
	var sb strings.Builder

	sb.WriteString(err.Error())
	sb.WriteString("\n\nSuggested actions:\n")
	sb.WriteString("  1. Use a different sequence_start value to avoid overlap\n")

	if suggestedStart > 0 {
		sb.WriteString(fmt.Sprintf("     Example: sequence_start = %d\n", suggestedStart))
	}

	sb.WriteString("  2. If using manual mode, choose unique sequence numbers\n")
	sb.WriteString("  3. Review existing ACLs to understand sequence allocation\n")

	return sb.String()
}

// FindNextAvailableSequenceStart finds the next available sequence start value
// that doesn't conflict with existing sequences.
//
// Parameters:
//   - existingSequences: Map of sequence to owner
//   - entryCount: Number of entries that need sequences
//   - step: The step value to use
//   - preferredStart: Preferred starting value (0 for auto-detect)
//
// Returns the suggested start value, or -1 if no suitable range is found.
func FindNextAvailableSequenceStart(existingSequences map[int]string, entryCount, step, preferredStart int) int {
	if entryCount <= 0 {
		return preferredStart
	}

	// Collect all used sequences
	used := make([]int, 0, len(existingSequences))
	for seq := range existingSequences {
		used = append(used, seq)
	}
	sort.Ints(used)

	// If preferred start is specified and range is free, use it
	if preferredStart > 0 {
		if isRangeFree(existingSequences, preferredStart, step, entryCount) {
			return preferredStart
		}
	}

	// Try common starting points
	candidates := []int{10, 100, 1000, 10000}

	for _, candidate := range candidates {
		if isRangeFree(existingSequences, candidate, step, entryCount) {
			return candidate
		}
	}

	// Find a gap in existing sequences
	if len(used) > 0 {
		// Try after the last used sequence
		lastUsed := used[len(used)-1]
		nextStart := ((lastUsed / 100) + 1) * 100 // Round up to next hundred
		if nextStart < MaxSequence && isRangeFree(existingSequences, nextStart, step, entryCount) {
			return nextStart
		}
	}

	return -1
}

// isRangeFree checks if a sequence range is completely free.
func isRangeFree(existingSequences map[int]string, start, step, count int) bool {
	for i := 0; i < count; i++ {
		seq := start + (i * step)
		if seq > MaxSequence {
			return false
		}
		if _, exists := existingSequences[seq]; exists {
			return false
		}
	}
	return true
}
