package fwhelpers

import (
	"fmt"
	"sort"
)

// SequenceConflict represents a detected sequence conflict.
type SequenceConflict struct {
	Sequence int
	Message  string
}

// CheckSequenceConflicts checks for conflicts between planned sequences and existing sequences.
// It returns a list of conflicting sequences.
//
// Parameters:
//   - planned: sequences that the resource wants to use
//   - existing: all sequences currently on the router
//   - currentState: sequences that the resource currently owns (will be excluded from conflict check)
//
// Returns sequences from planned that exist in existing but not in currentState.
func CheckSequenceConflicts(planned, existing, currentState []int) []int {
	// Build a set of existing sequences
	existingSet := make(map[int]bool, len(existing))
	for _, seq := range existing {
		existingSet[seq] = true
	}

	// Build a set of current state sequences (owned by this resource)
	currentSet := make(map[int]bool, len(currentState))
	for _, seq := range currentState {
		currentSet[seq] = true
	}

	// Find conflicts: sequences that are planned, exist on router, but not owned by this resource
	var conflicts []int
	for _, seq := range planned {
		if existingSet[seq] && !currentSet[seq] {
			conflicts = append(conflicts, seq)
		}
	}

	// Sort for deterministic output
	sort.Ints(conflicts)
	return conflicts
}

// FormatSequenceConflictError creates a formatted error message for sequence conflicts.
func FormatSequenceConflictError(resourceType, resourceName string, conflicts []int) string {
	return fmt.Sprintf(`The following sequence numbers are already in use on the router: %v

This may be caused by:
- Another Terraform resource using the same sequences
- Manual configuration on the router

To resolve, use different sequence_start values for each %s resource.
For example, if one resource uses sequence_start=1, another should use sequence_start=100.`,
		conflicts, resourceType)
}

// CalculateSequences calculates the sequence numbers based on start, step, and count.
// If start is 0, returns nil (manual mode).
func CalculateSequences(start, step, count int) []int {
	if start == 0 {
		return nil
	}
	if step == 0 {
		step = 10 // default step
	}

	sequences := make([]int, count)
	for i := 0; i < count; i++ {
		sequences[i] = start + (i * step)
	}
	return sequences
}
