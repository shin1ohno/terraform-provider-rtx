// Package provider contains tests for the collision validator for ACL resources.
package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCollisionError_Error tests the Error method of CollisionError.
func TestCollisionError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      CollisionError
		expected string
	}{
		{
			name: "custom message takes precedence",
			err: CollisionError{
				Sequence:      100,
				OwnedBy:       "resource-a",
				ConflictsWith: "resource-b",
				ACLType:       ACLTypeIP,
				Message:       "custom collision message",
			},
			expected: "custom collision message",
		},
		{
			name: "default message format when no custom message",
			err: CollisionError{
				Sequence:      200,
				OwnedBy:       "acl-group-1",
				ConflictsWith: "acl-group-2",
				ACLType:       ACLTypeMAC,
			},
			expected: `sequence 200 (owned by "acl-group-1") conflicts with "acl-group-2" for ACL type mac`,
		},
		{
			name: "ipv6 ACL type",
			err: CollisionError{
				Sequence:      300,
				OwnedBy:       "ipv6-acl",
				ConflictsWith: "other-ipv6-acl",
				ACLType:       ACLTypeIPv6,
			},
			expected: `sequence 300 (owned by "ipv6-acl") conflicts with "other-ipv6-acl" for ACL type ipv6`,
		},
		{
			name: "extended ACL type",
			err: CollisionError{
				Sequence:      500,
				OwnedBy:       "ext-acl",
				ConflictsWith: "router",
				ACLType:       ACLTypeExtended,
			},
			expected: `sequence 500 (owned by "ext-acl") conflicts with "router" for ACL type extended`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMultiCollisionError_Error tests the Error method of MultiCollisionError.
func TestMultiCollisionError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      MultiCollisionError
		contains []string
	}{
		{
			name:     "empty errors returns no collisions",
			err:      MultiCollisionError{Errors: []CollisionError{}},
			contains: []string{"no collisions"},
		},
		{
			name: "single collision",
			err: MultiCollisionError{
				Errors: []CollisionError{
					{
						Sequence:      100,
						OwnedBy:       "acl-a",
						ConflictsWith: "acl-b",
						ACLType:       ACLTypeIP,
					},
				},
			},
			contains: []string{"1 sequence collision", "sequence 100", "acl-a", "acl-b"},
		},
		{
			name: "multiple collisions",
			err: MultiCollisionError{
				Errors: []CollisionError{
					{
						Sequence:      100,
						OwnedBy:       "acl-1",
						ConflictsWith: "acl-2",
						ACLType:       ACLTypeIP,
					},
					{
						Sequence:      200,
						OwnedBy:       "acl-1",
						ConflictsWith: "acl-3",
						ACLType:       ACLTypeIP,
					},
					{
						Sequence:      300,
						OwnedBy:       "acl-4",
						ConflictsWith: "acl-5",
						ACLType:       ACLTypeMAC,
					},
				},
			},
			contains: []string{"3 sequence collision", "sequence 100", "sequence 200", "sequence 300"},
		},
		{
			name: "collision with custom message",
			err: MultiCollisionError{
				Errors: []CollisionError{
					{
						Sequence:      150,
						OwnedBy:       "test",
						ConflictsWith: "other",
						ACLType:       ACLTypeIP,
						Message:       "Custom error message for sequence 150",
					},
				},
			},
			contains: []string{"1 sequence collision", "Custom error message for sequence 150"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			for _, substr := range tt.contains {
				assert.Contains(t, result, substr)
			}
		})
	}
}

// TestSequenceRange_ContainsSequence tests the ContainsSequence method.
func TestSequenceRange_ContainsSequence(t *testing.T) {
	tests := []struct {
		name     string
		seqRange SequenceRange
		testSeq  int
		expected bool
	}{
		{
			name: "sequence is contained",
			seqRange: SequenceRange{
				Sequences: []int{100, 110, 120, 130},
			},
			testSeq:  110,
			expected: true,
		},
		{
			name: "sequence is not contained",
			seqRange: SequenceRange{
				Sequences: []int{100, 110, 120, 130},
			},
			testSeq:  115,
			expected: false,
		},
		{
			name: "first sequence",
			seqRange: SequenceRange{
				Sequences: []int{100, 200, 300},
			},
			testSeq:  100,
			expected: true,
		},
		{
			name: "last sequence",
			seqRange: SequenceRange{
				Sequences: []int{100, 200, 300},
			},
			testSeq:  300,
			expected: true,
		},
		{
			name: "empty sequences",
			seqRange: SequenceRange{
				Sequences: []int{},
			},
			testSeq:  100,
			expected: false,
		},
		{
			name: "single sequence match",
			seqRange: SequenceRange{
				Sequences: []int{50},
			},
			testSeq:  50,
			expected: true,
		},
		{
			name: "single sequence no match",
			seqRange: SequenceRange{
				Sequences: []int{50},
			},
			testSeq:  51,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.seqRange.ContainsSequence(tt.testSeq)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSequenceRange_Overlaps tests the Overlaps method.
func TestSequenceRange_Overlaps(t *testing.T) {
	tests := []struct {
		name     string
		range1   SequenceRange
		range2   SequenceRange
		expected []int
	}{
		{
			name: "no overlap - non-overlapping ranges",
			range1: SequenceRange{
				Sequences: []int{100, 110, 120},
			},
			range2: SequenceRange{
				Sequences: []int{200, 210, 220},
			},
			expected: []int{},
		},
		{
			name: "single sequence overlap",
			range1: SequenceRange{
				Sequences: []int{100, 110, 120},
			},
			range2: SequenceRange{
				Sequences: []int{120, 130, 140},
			},
			expected: []int{120},
		},
		{
			name: "multiple sequence overlap",
			range1: SequenceRange{
				Sequences: []int{100, 110, 120, 130},
			},
			range2: SequenceRange{
				Sequences: []int{110, 120, 200},
			},
			expected: []int{110, 120},
		},
		{
			name: "complete overlap",
			range1: SequenceRange{
				Sequences: []int{100, 110, 120},
			},
			range2: SequenceRange{
				Sequences: []int{100, 110, 120},
			},
			expected: []int{100, 110, 120},
		},
		{
			name: "one range contains the other",
			range1: SequenceRange{
				Sequences: []int{100, 110, 120, 130, 140},
			},
			range2: SequenceRange{
				Sequences: []int{110, 120, 130},
			},
			expected: []int{110, 120, 130},
		},
		{
			name: "empty first range",
			range1: SequenceRange{
				Sequences: []int{},
			},
			range2: SequenceRange{
				Sequences: []int{100, 110, 120},
			},
			expected: []int{},
		},
		{
			name: "empty second range",
			range1: SequenceRange{
				Sequences: []int{100, 110, 120},
			},
			range2: SequenceRange{
				Sequences: []int{},
			},
			expected: []int{},
		},
		{
			name: "both ranges empty",
			range1: SequenceRange{
				Sequences: []int{},
			},
			range2: SequenceRange{
				Sequences: []int{},
			},
			expected: []int{},
		},
		{
			name: "interleaved sequences with some overlap",
			range1: SequenceRange{
				Sequences: []int{100, 120, 140, 160},
			},
			range2: SequenceRange{
				Sequences: []int{110, 120, 130, 140, 150},
			},
			expected: []int{120, 140},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.range1.Overlaps(&tt.range2)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

// TestSequenceRange_String tests the String method.
func TestSequenceRange_String(t *testing.T) {
	tests := []struct {
		name     string
		seqRange SequenceRange
		contains []string
	}{
		{
			name: "no sequences",
			seqRange: SequenceRange{
				ResourceName: "test-resource",
				Sequences:    []int{},
			},
			contains: []string{"test-resource", "no sequences"},
		},
		{
			name: "single sequence",
			seqRange: SequenceRange{
				ResourceName: "single-seq-resource",
				Sequences:    []int{100},
			},
			contains: []string{"single-seq-resource", "sequence 100"},
		},
		{
			name: "multiple sequences",
			seqRange: SequenceRange{
				ResourceName: "multi-seq-resource",
				Sequences:    []int{100, 110, 120, 130},
				Start:        100,
				End:          130,
			},
			contains: []string{"multi-seq-resource", "100", "130", "count=4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.seqRange.String()
			for _, substr := range tt.contains {
				assert.Contains(t, result, substr)
			}
		})
	}
}

// TestValidateSequenceRangeNoOverlap tests the ValidateSequenceRangeNoOverlap function.
func TestValidateSequenceRangeNoOverlap(t *testing.T) {
	tests := []struct {
		name        string
		ranges      []SequenceRange
		expectError bool
		errorCount  int
		errorSeqs   []int
	}{
		{
			name:        "empty ranges - no error",
			ranges:      []SequenceRange{},
			expectError: false,
		},
		{
			name: "single range - no error",
			ranges: []SequenceRange{
				{
					ResourceName: "acl-1",
					ACLType:      ACLTypeIP,
					Sequences:    []int{100, 110, 120},
				},
			},
			expectError: false,
		},
		{
			name: "two non-overlapping ranges - no error",
			ranges: []SequenceRange{
				{
					ResourceName: "acl-1",
					ACLType:      ACLTypeIP,
					Sequences:    []int{100, 110, 120},
				},
				{
					ResourceName: "acl-2",
					ACLType:      ACLTypeIP,
					Sequences:    []int{200, 210, 220},
				},
			},
			expectError: false,
		},
		{
			name: "different ACL types - no collision even if overlapping",
			ranges: []SequenceRange{
				{
					ResourceName: "acl-ip",
					ACLType:      ACLTypeIP,
					Sequences:    []int{100, 110, 120},
				},
				{
					ResourceName: "acl-mac",
					ACLType:      ACLTypeMAC,
					Sequences:    []int{100, 110, 120},
				},
			},
			expectError: false,
		},
		{
			name: "overlapping ranges same ACL type - error",
			ranges: []SequenceRange{
				{
					ResourceName: "acl-1",
					ACLType:      ACLTypeIP,
					Sequences:    []int{100, 110, 120},
				},
				{
					ResourceName: "acl-2",
					ACLType:      ACLTypeIP,
					Sequences:    []int{120, 130, 140},
				},
			},
			expectError: true,
			errorCount:  1,
			errorSeqs:   []int{120},
		},
		{
			name: "multiple overlapping sequences - multiple errors",
			ranges: []SequenceRange{
				{
					ResourceName: "acl-1",
					ACLType:      ACLTypeIPv6,
					Sequences:    []int{100, 110, 120, 130},
				},
				{
					ResourceName: "acl-2",
					ACLType:      ACLTypeIPv6,
					Sequences:    []int{110, 120, 200},
				},
			},
			expectError: true,
			errorCount:  2,
			errorSeqs:   []int{110, 120},
		},
		{
			name: "three ranges with one overlapping pair",
			ranges: []SequenceRange{
				{
					ResourceName: "acl-1",
					ACLType:      ACLTypeMAC,
					Sequences:    []int{100, 110},
				},
				{
					ResourceName: "acl-2",
					ACLType:      ACLTypeMAC,
					Sequences:    []int{200, 210},
				},
				{
					ResourceName: "acl-3",
					ACLType:      ACLTypeMAC,
					Sequences:    []int{210, 220},
				},
			},
			expectError: true,
			errorCount:  1,
			errorSeqs:   []int{210},
		},
		{
			name: "multiple ranges all overlapping",
			ranges: []SequenceRange{
				{
					ResourceName: "acl-1",
					ACLType:      ACLTypeIP,
					Sequences:    []int{100, 110, 120},
				},
				{
					ResourceName: "acl-2",
					ACLType:      ACLTypeIP,
					Sequences:    []int{100, 120, 130},
				},
				{
					ResourceName: "acl-3",
					ACLType:      ACLTypeIP,
					Sequences:    []int{100, 150},
				},
			},
			expectError: true,
			errorCount:  4, // 100,120 from acl-1 vs acl-2, 100 from acl-1 vs acl-3, 100 from acl-2 vs acl-3
			errorSeqs:   []int{100, 120},
		},
		{
			name: "mixed ACL types with some overlap in same type",
			ranges: []SequenceRange{
				{
					ResourceName: "ip-acl-1",
					ACLType:      ACLTypeIP,
					Sequences:    []int{100, 110},
				},
				{
					ResourceName: "mac-acl-1",
					ACLType:      ACLTypeMAC,
					Sequences:    []int{100, 110},
				},
				{
					ResourceName: "ip-acl-2",
					ACLType:      ACLTypeIP,
					Sequences:    []int{110, 120},
				},
			},
			expectError: true,
			errorCount:  1,
			errorSeqs:   []int{110},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSequenceRangeNoOverlap(tt.ranges)

			if !tt.expectError {
				assert.NoError(t, err)
				return
			}

			require.Error(t, err)

			multiErr, ok := err.(*MultiCollisionError)
			require.True(t, ok, "error should be *MultiCollisionError")
			assert.Len(t, multiErr.Errors, tt.errorCount)

			// Verify expected sequences are in the errors
			foundSeqs := make(map[int]bool)
			for _, collErr := range multiErr.Errors {
				foundSeqs[collErr.Sequence] = true
			}
			for _, seq := range tt.errorSeqs {
				assert.True(t, foundSeqs[seq], "expected sequence %d in errors", seq)
			}
		})
	}
}

// TestBuildCollisionErrorMessage tests the BuildCollisionErrorMessage function.
func TestBuildCollisionErrorMessage(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		suggestedStart int
		suggestedStep  int
		contains       []string
	}{
		{
			name: "with suggested start",
			err: &CollisionError{
				Sequence:      100,
				OwnedBy:       "acl-1",
				ConflictsWith: "acl-2",
				ACLType:       ACLTypeIP,
			},
			suggestedStart: 500,
			suggestedStep:  10,
			contains: []string{
				"sequence 100",
				"Suggested actions",
				"sequence_start = 500",
				"different sequence_start",
			},
		},
		{
			name: "without suggested start",
			err: &CollisionError{
				Sequence:      200,
				OwnedBy:       "acl-a",
				ConflictsWith: "acl-b",
				ACLType:       ACLTypeMAC,
			},
			suggestedStart: 0,
			suggestedStep:  10,
			contains: []string{
				"sequence 200",
				"Suggested actions",
				"manual mode",
				"Review existing ACLs",
			},
		},
		{
			name: "multi collision error",
			err: &MultiCollisionError{
				Errors: []CollisionError{
					{Sequence: 100, OwnedBy: "a", ConflictsWith: "b", ACLType: ACLTypeIP},
					{Sequence: 110, OwnedBy: "a", ConflictsWith: "c", ACLType: ACLTypeIP},
				},
			},
			suggestedStart: 1000,
			suggestedStep:  10,
			contains: []string{
				"2 sequence collision",
				"Suggested actions",
				"sequence_start = 1000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildCollisionErrorMessage(tt.err, tt.suggestedStart, tt.suggestedStep)

			for _, substr := range tt.contains {
				assert.Contains(t, result, substr)
			}
		})
	}
}

// TestFindNextAvailableSequenceStart tests the FindNextAvailableSequenceStart function.
func TestFindNextAvailableSequenceStart(t *testing.T) {
	tests := []struct {
		name              string
		existingSequences map[int]string
		entryCount        int
		step              int
		preferredStart    int
		expected          int
	}{
		{
			name:              "empty sequences - use preferred start",
			existingSequences: map[int]string{},
			entryCount:        3,
			step:              10,
			preferredStart:    100,
			expected:          100,
		},
		{
			name:              "preferred start available",
			existingSequences: map[int]string{200: "other", 210: "other"},
			entryCount:        3,
			step:              10,
			preferredStart:    100,
			expected:          100,
		},
		{
			name:              "preferred start conflicts - find alternative",
			existingSequences: map[int]string{100: "existing", 110: "existing"},
			entryCount:        3,
			step:              10,
			preferredStart:    100,
			expected:          10, // First candidate
		},
		{
			name:              "no preferred start - auto detect",
			existingSequences: map[int]string{},
			entryCount:        3,
			step:              10,
			preferredStart:    0,
			expected:          10,
		},
		{
			name:              "first candidates taken - try 100",
			existingSequences: map[int]string{10: "a", 20: "b", 30: "c"},
			entryCount:        3,
			step:              10,
			preferredStart:    0,
			expected:          100,
		},
		{
			name:              "entryCount=0 returns preferred start",
			existingSequences: map[int]string{100: "existing"},
			entryCount:        0,
			step:              10,
			preferredStart:    500,
			expected:          500,
		},
		{
			name:              "all common candidates taken - use gap after last",
			existingSequences: map[int]string{10: "a", 100: "b", 1000: "c", 10000: "d"},
			entryCount:        3,
			step:              10,
			preferredStart:    0,
			expected:          10100, // Next hundred after 10000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindNextAvailableSequenceStart(
				tt.existingSequences,
				tt.entryCount,
				tt.step,
				tt.preferredStart,
			)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsRangeFree tests that isRangeFree correctly identifies free ranges.
func TestIsRangeFree(t *testing.T) {
	// Test through FindNextAvailableSequenceStart behavior
	tests := []struct {
		name              string
		existingSequences map[int]string
		entryCount        int
		step              int
		preferredStart    int
		shouldBePreferred bool
	}{
		{
			name:              "range is free",
			existingSequences: map[int]string{50: "a"},
			entryCount:        3,
			step:              10,
			preferredStart:    100,
			shouldBePreferred: true,
		},
		{
			name:              "range has conflict at start",
			existingSequences: map[int]string{100: "conflict"},
			entryCount:        3,
			step:              10,
			preferredStart:    100,
			shouldBePreferred: false,
		},
		{
			name:              "range has conflict in middle",
			existingSequences: map[int]string{110: "conflict"},
			entryCount:        3,
			step:              10,
			preferredStart:    100,
			shouldBePreferred: false,
		},
		{
			name:              "range has conflict at end",
			existingSequences: map[int]string{120: "conflict"},
			entryCount:        3,
			step:              10,
			preferredStart:    100,
			shouldBePreferred: false,
		},
		{
			name:              "adjacent sequences are ok",
			existingSequences: map[int]string{90: "before", 130: "after"},
			entryCount:        3,
			step:              10,
			preferredStart:    100,
			shouldBePreferred: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindNextAvailableSequenceStart(
				tt.existingSequences,
				tt.entryCount,
				tt.step,
				tt.preferredStart,
			)
			if tt.shouldBePreferred {
				assert.Equal(t, tt.preferredStart, result)
			} else {
				assert.NotEqual(t, tt.preferredStart, result)
			}
		})
	}
}

// TestBuildSequenceRange tests the buildSequenceRange function.
func TestBuildSequenceRange(t *testing.T) {
	tests := []struct {
		name          string
		resourceID    string
		aclName       string
		aclType       ACLType
		sequences     []int
		expectedStart int
		expectedEnd   int
	}{
		{
			name:          "empty sequences",
			resourceID:    "res-1",
			aclName:       "acl-1",
			aclType:       ACLTypeIP,
			sequences:     []int{},
			expectedStart: 0,
			expectedEnd:   0,
		},
		{
			name:          "single sequence",
			resourceID:    "res-2",
			aclName:       "acl-2",
			aclType:       ACLTypeMAC,
			sequences:     []int{100},
			expectedStart: 100,
			expectedEnd:   100,
		},
		{
			name:          "multiple sequences in order",
			resourceID:    "res-3",
			aclName:       "acl-3",
			aclType:       ACLTypeIPv6,
			sequences:     []int{100, 110, 120},
			expectedStart: 100,
			expectedEnd:   120,
		},
		{
			name:          "multiple sequences out of order",
			resourceID:    "res-4",
			aclName:       "acl-4",
			aclType:       ACLTypeExtended,
			sequences:     []int{300, 100, 200},
			expectedStart: 100,
			expectedEnd:   300,
		},
		{
			name:          "sequences with gaps",
			resourceID:    "res-5",
			aclName:       "acl-5",
			aclType:       ACLTypeIP,
			sequences:     []int{10, 50, 1000, 500},
			expectedStart: 10,
			expectedEnd:   1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildSequenceRange(tt.resourceID, tt.aclName, tt.aclType, tt.sequences)

			require.NotNil(t, result)
			assert.Equal(t, tt.resourceID, result.ResourceName)
			assert.Equal(t, tt.aclName, result.ACLName)
			assert.Equal(t, tt.aclType, result.ACLType)
			assert.Equal(t, tt.sequences, result.Sequences)

			if len(tt.sequences) > 0 {
				assert.Equal(t, tt.expectedStart, result.Start)
				assert.Equal(t, tt.expectedEnd, result.End)
			}
		})
	}
}

// TestCollisionErrorTypes ensures error types implement error interface correctly.
func TestCollisionErrorTypes(t *testing.T) {
	t.Run("CollisionError implements error", func(t *testing.T) {
		var err error = &CollisionError{
			Sequence:      100,
			OwnedBy:       "test",
			ConflictsWith: "other",
			ACLType:       ACLTypeIP,
		}
		assert.NotEmpty(t, err.Error())
	})

	t.Run("MultiCollisionError implements error", func(t *testing.T) {
		var err error = &MultiCollisionError{
			Errors: []CollisionError{
				{Sequence: 100, OwnedBy: "test", ConflictsWith: "other", ACLType: ACLTypeIP},
			},
		}
		assert.NotEmpty(t, err.Error())
	})
}
