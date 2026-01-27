// Package provider contains the sequence calculator for ACL resources.
//
// This file provides pure functions for calculating sequence numbers in automatic mode
// and detecting the sequence mode (auto vs manual) from Terraform resource data.
// These functions are used by all ACL resources that support the unified schema.
package provider

import (
	"errors"
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// SequenceMode represents how sequence numbers are assigned to ACL entries.
type SequenceMode int

const (
	// AutoMode indicates sequences are automatically calculated from sequence_start and sequence_step.
	// In this mode, users must not specify sequence numbers on individual entries.
	AutoMode SequenceMode = iota

	// ManualMode indicates sequences are explicitly specified on each entry.
	// In this mode, sequence_start must not be set (or set to 0).
	ManualMode

	// MixedMode indicates an invalid configuration where both auto and manual
	// sequence assignment are attempted. This mode should trigger a validation error.
	MixedMode
)

// String returns a human-readable name for the sequence mode.
func (m SequenceMode) String() string {
	switch m {
	case AutoMode:
		return "auto"
	case ManualMode:
		return "manual"
	case MixedMode:
		return "mixed"
	default:
		return "unknown"
	}
}

// ErrInvalidStep is returned when step is zero or negative.
var ErrInvalidStep = errors.New("step must be a positive integer")

// ErrInvalidCount is returned when count is negative.
var ErrInvalidCount = errors.New("count must be non-negative")

// ErrInvalidStart is returned when start is not positive.
var ErrInvalidStart = errors.New("start must be a positive integer")

// ErrSequenceOverflow is returned when calculated sequences exceed the valid range.
var ErrSequenceOverflow = errors.New("calculated sequence exceeds maximum allowed value")

// MaxSequence is the maximum allowed sequence number for RTX ACL entries.
// RTX routers support filter IDs from 1 to 65535.
// Reference: RTX Command Reference Chapter 13 (Ethernet Filtering), Chapter 14 (IP Packet Filtering)
const MaxSequence = 65535

// MinSequence is the minimum allowed sequence number for RTX ACL entries.
const MinSequence = 1

// DefaultSequenceStart is the default starting sequence number when auto mode is enabled.
const DefaultSequenceStart = 10

// DefaultSequenceStep is the default step between sequence numbers in auto mode.
const DefaultSequenceStep = 10

// CalculateSequences computes sequence numbers for ACL entries in automatic mode.
//
// This function generates a slice of sequence numbers starting from `start` and
// incrementing by `step` for each subsequent entry. The function is pure and has
// no side effects.
//
// Parameters:
//   - start: The first sequence number to use (must be positive)
//   - step: The increment between consecutive sequence numbers (must be positive)
//   - count: The number of entries to generate sequences for (must be non-negative)
//
// Returns:
//   - []int: A slice of calculated sequence numbers
//   - error: Non-nil if inputs are invalid or overflow would occur
//
// Example:
//
//	sequences, err := CalculateSequences(100, 10, 3)
//	// sequences = [100, 110, 120]
//
//	sequences, err := CalculateSequences(1, 1, 5)
//	// sequences = [1, 2, 3, 4, 5]
func CalculateSequences(start, step, count int) ([]int, error) {
	if err := ValidateSequenceRange(start, step, count); err != nil {
		return nil, err
	}

	if count == 0 {
		return []int{}, nil
	}

	sequences := make([]int, count)
	for i := 0; i < count; i++ {
		sequences[i] = start + (i * step)
	}

	return sequences, nil
}

// ValidateSequenceRange checks if the given parameters would produce valid sequences.
//
// This function validates that:
//   - start is positive and within valid range
//   - step is positive
//   - count is non-negative
//   - all calculated sequences fit within MaxSequence
//
// Parameters:
//   - start: The first sequence number
//   - step: The increment between sequences
//   - count: The number of sequences to generate
//
// Returns:
//   - error: Non-nil if any validation fails, with a descriptive message
//
// Example:
//
//	err := ValidateSequenceRange(99990, 10, 3)
//	// err contains overflow message (99990, 100000, 100010 - last two exceed 99999)
func ValidateSequenceRange(start, step, count int) error {
	if start < MinSequence {
		return fmt.Errorf("%w: got %d, minimum is %d", ErrInvalidStart, start, MinSequence)
	}

	if start > MaxSequence {
		return fmt.Errorf("%w: start %d exceeds maximum %d", ErrSequenceOverflow, start, MaxSequence)
	}

	if step <= 0 {
		return fmt.Errorf("%w: got %d", ErrInvalidStep, step)
	}

	if count < 0 {
		return fmt.Errorf("%w: got %d", ErrInvalidCount, count)
	}

	if count == 0 {
		return nil
	}

	// Calculate the last sequence number and check for overflow
	// Use int64 to detect overflow before it happens
	lastIndex := int64(count - 1)
	lastSequence := int64(start) + (lastIndex * int64(step))

	// Check for integer overflow in the calculation itself
	if lastSequence > math.MaxInt32 || lastSequence < 0 {
		return fmt.Errorf("%w: calculation overflow with start=%d, step=%d, count=%d",
			ErrSequenceOverflow, start, step, count)
	}

	if lastSequence > int64(MaxSequence) {
		return fmt.Errorf("%w: last sequence %d exceeds maximum %d (start=%d, step=%d, count=%d)",
			ErrSequenceOverflow, lastSequence, MaxSequence, start, step, count)
	}

	return nil
}

// DetectSequenceMode determines whether an ACL uses automatic or manual sequence assignment.
//
// The mode is determined by examining the resource data:
//   - If sequence_start is set and > 0: AutoMode (sequences calculated automatically)
//   - If sequence_start is not set or 0 and entries have sequences: ManualMode
//   - If sequence_start is set but entries also have explicit sequences: MixedMode (invalid)
//
// Parameters:
//   - d: The Terraform ResourceData containing the ACL configuration
//
// Returns:
//   - SequenceMode: The detected mode (AutoMode, ManualMode, or MixedMode)
//
// Note: MixedMode indicates an invalid configuration that should be rejected during validation.
func DetectSequenceMode(d *schema.ResourceData) SequenceMode {
	sequenceStart, hasStart := d.GetOk("sequence_start")
	isAutoMode := hasStart && sequenceStart.(int) > 0

	// Check if any entries have explicit sequences
	entries, hasEntries := d.GetOk("entry")
	hasExplicitSequences := false

	if hasEntries {
		entryList, ok := entries.([]interface{})
		if ok {
			for _, e := range entryList {
				entry, ok := e.(map[string]interface{})
				if !ok {
					continue
				}
				// Check if sequence is explicitly set (non-zero)
				if seq, exists := entry["sequence"]; exists {
					if seqInt, ok := seq.(int); ok && seqInt > 0 {
						hasExplicitSequences = true
						break
					}
				}
			}
		}
	}

	// Determine mode based on configuration
	if isAutoMode && hasExplicitSequences {
		return MixedMode
	}

	if isAutoMode {
		return AutoMode
	}

	return ManualMode
}

// GetSequenceStartAndStep retrieves sequence_start and sequence_step from ResourceData.
//
// If sequence_start is not set or is 0, this returns (0, 0, false) indicating manual mode.
// If sequence_step is not set, it defaults to DefaultSequenceStep.
//
// Parameters:
//   - d: The Terraform ResourceData
//
// Returns:
//   - start: The sequence start value (0 if manual mode)
//   - step: The sequence step value (0 if manual mode, defaults to DefaultSequenceStep otherwise)
//   - isAuto: True if auto mode is enabled
func GetSequenceStartAndStep(d *schema.ResourceData) (start, step int, isAuto bool) {
	startVal, hasStart := d.GetOk("sequence_start")
	if !hasStart || startVal.(int) == 0 {
		return 0, 0, false
	}

	start = startVal.(int)

	if stepVal, hasStep := d.GetOk("sequence_step"); hasStep {
		step = stepVal.(int)
	} else {
		step = DefaultSequenceStep
	}

	return start, step, true
}

// CalculateSequencesForResourceData computes sequences for entries in ResourceData.
//
// This is a convenience function that combines GetSequenceStartAndStep and CalculateSequences
// for use directly with Terraform ResourceData. If auto mode is not enabled, it returns nil.
//
// Parameters:
//   - d: The Terraform ResourceData
//
// Returns:
//   - []int: Calculated sequences if auto mode is enabled, nil otherwise
//   - error: Non-nil if calculation fails
func CalculateSequencesForResourceData(d *schema.ResourceData) ([]int, error) {
	start, step, isAuto := GetSequenceStartAndStep(d)
	if !isAuto {
		return nil, nil
	}

	// Count entries
	entries, hasEntries := d.GetOk("entry")
	if !hasEntries {
		return []int{}, nil
	}

	entryList, ok := entries.([]interface{})
	if !ok {
		return []int{}, nil
	}

	return CalculateSequences(start, step, len(entryList))
}
