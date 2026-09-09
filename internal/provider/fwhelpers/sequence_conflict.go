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

// CheckSequenceContentConflicts reports the planned sequences that would overwrite
// somebody else's filter. Both maps go from sequence number to a canonical rendering
// of the entry at that sequence: planned is what the resource is about to write,
// existing is what the router holds now. A planned sequence conflicts only when the
// router already holds it, this resource does not own it, AND the router's rendering
// differs from the planned one.
//
// The identical-content exemption is the point. The filter resources, static and
// dynamic alike, do not read back rows they do not own (each Read walks the sequences
// recorded in state, see access_list_ipv6_dynamic/model.go FromClient and the
// GetIPFilter loop in access_list_ip/resource.go), so any state that carries fewer entries than the router — a state built before
// entries were appended to config, a partial apply, a `terraform state rm` — makes the
// rows this resource wrote last time look like somebody else's. With a bare-sequence
// check the resource then refuses to write its own output again and stays wedged
// until the state is repaired by hand (rtx_access_list_ipv6_dynamic.wan_outbound,
// 2026-09-08: state held 8 of 11 rows, every apply failed on sequences 41 46 51).
// Re-writing a row the router already holds in exactly the planned form changes
// nothing on the router, so it is not a conflict. A row with different content at a
// planned sequence still is, whoever wrote it.
func CheckSequenceContentConflicts(planned, existing map[int]string, currentState []int) []int {
	currentSet := make(map[int]bool, len(currentState))
	for _, seq := range currentState {
		currentSet[seq] = true
	}

	var conflicts []int
	for seq, want := range planned {
		if currentSet[seq] {
			continue
		}
		have, onRouter := existing[seq]
		if onRouter && have != want {
			conflicts = append(conflicts, seq)
		}
	}

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
