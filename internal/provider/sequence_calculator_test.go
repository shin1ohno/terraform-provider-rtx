// Package provider contains tests for the sequence calculator for ACL resources.
package provider

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSequenceMode_String tests the String method of SequenceMode.
func TestSequenceMode_String(t *testing.T) {
	tests := []struct {
		name     string
		mode     SequenceMode
		expected string
	}{
		{
			name:     "AutoMode",
			mode:     AutoMode,
			expected: "auto",
		},
		{
			name:     "ManualMode",
			mode:     ManualMode,
			expected: "manual",
		},
		{
			name:     "MixedMode",
			mode:     MixedMode,
			expected: "mixed",
		},
		{
			name:     "unknown mode",
			mode:     SequenceMode(999),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mode.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCalculateSequences tests the CalculateSequences function.
func TestCalculateSequences(t *testing.T) {
	tests := []struct {
		name          string
		start         int
		step          int
		count         int
		expected      []int
		expectedError error
	}{
		// Normal cases
		{
			name:     "basic calculation start=100 step=10 count=3",
			start:    100,
			step:     10,
			count:    3,
			expected: []int{100, 110, 120},
		},
		{
			name:     "start=1 step=1 count=5",
			start:    1,
			step:     1,
			count:    5,
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "start=10 step=10 count=1",
			start:    10,
			step:     10,
			count:    1,
			expected: []int{10},
		},
		{
			name:     "large step",
			start:    100,
			step:     1000,
			count:    3,
			expected: []int{100, 1100, 2100},
		},
		{
			name:     "default values start=10 step=10 count=5",
			start:    DefaultSequenceStart,
			step:     DefaultSequenceStep,
			count:    5,
			expected: []int{10, 20, 30, 40, 50},
		},
		// Edge cases
		{
			name:     "count=0 returns empty slice",
			start:    100,
			step:     10,
			count:    0,
			expected: []int{},
		},
		{
			name:     "minimum valid start",
			start:    MinSequence,
			step:     1,
			count:    3,
			expected: []int{1, 2, 3},
		},
		{
			name:     "near maximum sequence",
			start:    65525,
			step:     3,
			count:    3,
			expected: []int{65525, 65528, 65531},
		},
		// Error cases
		{
			name:          "step=0 returns error",
			start:         100,
			step:          0,
			count:         3,
			expected:      nil,
			expectedError: ErrInvalidStep,
		},
		{
			name:          "negative step returns error",
			start:         100,
			step:          -10,
			count:         3,
			expected:      nil,
			expectedError: ErrInvalidStep,
		},
		{
			name:          "negative count returns error",
			start:         100,
			step:          10,
			count:         -1,
			expected:      nil,
			expectedError: ErrInvalidCount,
		},
		{
			name:          "start=0 returns error",
			start:         0,
			step:          10,
			count:         3,
			expected:      nil,
			expectedError: ErrInvalidStart,
		},
		{
			name:          "negative start returns error",
			start:         -100,
			step:          10,
			count:         3,
			expected:      nil,
			expectedError: ErrInvalidStart,
		},
		{
			name:          "start exceeds MaxSequence returns error",
			start:         MaxSequence + 1,
			step:          10,
			count:         1,
			expected:      nil,
			expectedError: ErrSequenceOverflow,
		},
		{
			name:          "sequence overflow returns error",
			start:         65530,
			step:          10,
			count:         3,
			expected:      nil,
			expectedError: ErrSequenceOverflow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateSequences(tt.start, tt.step, tt.count)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError),
					"expected error %v, got %v", tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestValidateSequenceRange tests the ValidateSequenceRange function.
func TestValidateSequenceRange(t *testing.T) {
	tests := []struct {
		name          string
		start         int
		step          int
		count         int
		expectedError error
	}{
		// Valid cases
		{
			name:          "valid basic range",
			start:         100,
			step:          10,
			count:         3,
			expectedError: nil,
		},
		{
			name:          "valid minimum start",
			start:         MinSequence,
			step:          1,
			count:         10,
			expectedError: nil,
		},
		{
			name:          "valid single entry",
			start:         50000,
			step:          100,
			count:         1,
			expectedError: nil,
		},
		{
			name:          "valid count=0",
			start:         100,
			step:          10,
			count:         0,
			expectedError: nil,
		},
		{
			name:          "valid edge case at MaxSequence",
			start:         MaxSequence,
			step:          1,
			count:         1,
			expectedError: nil,
		},
		{
			name:          "valid large step with single entry",
			start:         1,
			step:          99999,
			count:         1,
			expectedError: nil,
		},
		// Invalid start cases
		{
			name:          "start=0 is invalid",
			start:         0,
			step:          10,
			count:         1,
			expectedError: ErrInvalidStart,
		},
		{
			name:          "negative start is invalid",
			start:         -1,
			step:          10,
			count:         1,
			expectedError: ErrInvalidStart,
		},
		{
			name:          "start exceeds MaxSequence",
			start:         MaxSequence + 1,
			step:          1,
			count:         1,
			expectedError: ErrSequenceOverflow,
		},
		// Invalid step cases
		{
			name:          "step=0 is invalid",
			start:         100,
			step:          0,
			count:         1,
			expectedError: ErrInvalidStep,
		},
		{
			name:          "negative step is invalid",
			start:         100,
			step:          -1,
			count:         1,
			expectedError: ErrInvalidStep,
		},
		// Invalid count cases
		{
			name:          "negative count is invalid",
			start:         100,
			step:          10,
			count:         -1,
			expectedError: ErrInvalidCount,
		},
		// Overflow cases
		{
			name:          "overflow at end of sequence",
			start:         65530,
			step:          10,
			count:         3,
			expectedError: ErrSequenceOverflow,
		},
		{
			name:          "overflow with large count",
			start:         1,
			step:          1,
			count:         MaxSequence + 1,
			expectedError: ErrSequenceOverflow,
		},
		{
			name:          "overflow with large step and count",
			start:         1,
			step:          100000000,
			count:         100,
			expectedError: ErrSequenceOverflow,
		},
		{
			name:          "overflow due to integer calculation",
			start:         1000000000,
			step:          1000000000,
			count:         10,
			expectedError: ErrSequenceOverflow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSequenceRange(tt.start, tt.step, tt.count)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError),
					"expected error %v, got %v", tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// testSchemaForSequenceCalculator returns a test schema for sequence calculator testing.
func testSchemaForSequenceCalculator() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"sequence_start": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"sequence_step": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"entry": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"sequence": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"action": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
		},
	}
}

// TestDetectSequenceMode tests the DetectSequenceMode function.
func TestDetectSequenceMode(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		expected SequenceMode
	}{
		// AutoMode cases
		{
			name: "auto mode with sequence_start only",
			config: map[string]interface{}{
				"sequence_start": 100,
			},
			expected: AutoMode,
		},
		{
			name: "auto mode with sequence_start and sequence_step",
			config: map[string]interface{}{
				"sequence_start": 100,
				"sequence_step":  20,
			},
			expected: AutoMode,
		},
		{
			name: "auto mode with entries without explicit sequences",
			config: map[string]interface{}{
				"sequence_start": 100,
				"entry": []interface{}{
					map[string]interface{}{
						"action":   "pass",
						"sequence": 0,
					},
					map[string]interface{}{
						"action": "deny",
					},
				},
			},
			expected: AutoMode,
		},
		// ManualMode cases
		{
			name:     "manual mode with no config",
			config:   map[string]interface{}{},
			expected: ManualMode,
		},
		{
			name: "manual mode with sequence_start=0",
			config: map[string]interface{}{
				"sequence_start": 0,
			},
			expected: ManualMode,
		},
		{
			name: "manual mode with entries with explicit sequences",
			config: map[string]interface{}{
				"entry": []interface{}{
					map[string]interface{}{
						"sequence": 100,
						"action":   "pass",
					},
					map[string]interface{}{
						"sequence": 200,
						"action":   "deny",
					},
				},
			},
			expected: ManualMode,
		},
		// MixedMode cases
		{
			name: "mixed mode with sequence_start and explicit entry sequences",
			config: map[string]interface{}{
				"sequence_start": 100,
				"entry": []interface{}{
					map[string]interface{}{
						"sequence": 500,
						"action":   "pass",
					},
				},
			},
			expected: MixedMode,
		},
		{
			name: "mixed mode with sequence_start and some entries with sequences",
			config: map[string]interface{}{
				"sequence_start": 100,
				"sequence_step":  10,
				"entry": []interface{}{
					map[string]interface{}{
						"action": "pass",
					},
					map[string]interface{}{
						"sequence": 200,
						"action":   "deny",
					},
				},
			},
			expected: MixedMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, testSchemaForSequenceCalculator(), tt.config)
			result := DetectSequenceMode(d)
			assert.Equal(t, tt.expected, result,
				"DetectSequenceMode() = %s, expected %s", result.String(), tt.expected.String())
		})
	}
}

// TestGetSequenceStartAndStep tests the GetSequenceStartAndStep function.
func TestGetSequenceStartAndStep(t *testing.T) {
	tests := []struct {
		name         string
		config       map[string]interface{}
		expectedSt   int
		expectedStep int
		expectedAuto bool
	}{
		// Auto mode cases
		{
			name: "auto mode with start and step",
			config: map[string]interface{}{
				"sequence_start": 100,
				"sequence_step":  20,
			},
			expectedSt:   100,
			expectedStep: 20,
			expectedAuto: true,
		},
		{
			name: "auto mode with start only (default step)",
			config: map[string]interface{}{
				"sequence_start": 50,
			},
			expectedSt:   50,
			expectedStep: DefaultSequenceStep,
			expectedAuto: true,
		},
		{
			name: "auto mode with large values",
			config: map[string]interface{}{
				"sequence_start": 90000,
				"sequence_step":  100,
			},
			expectedSt:   90000,
			expectedStep: 100,
			expectedAuto: true,
		},
		// Manual mode cases
		{
			name:         "manual mode with no config",
			config:       map[string]interface{}{},
			expectedSt:   0,
			expectedStep: 0,
			expectedAuto: false,
		},
		{
			name: "manual mode with sequence_start=0",
			config: map[string]interface{}{
				"sequence_start": 0,
			},
			expectedSt:   0,
			expectedStep: 0,
			expectedAuto: false,
		},
		{
			name: "manual mode with only sequence_step (no start)",
			config: map[string]interface{}{
				"sequence_step": 20,
			},
			expectedSt:   0,
			expectedStep: 0,
			expectedAuto: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, testSchemaForSequenceCalculator(), tt.config)
			start, step, isAuto := GetSequenceStartAndStep(d)

			assert.Equal(t, tt.expectedSt, start, "start")
			assert.Equal(t, tt.expectedStep, step, "step")
			assert.Equal(t, tt.expectedAuto, isAuto, "isAuto")
		})
	}
}

// TestCalculateSequencesForResourceData tests the CalculateSequencesForResourceData function.
func TestCalculateSequencesForResourceData(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]interface{}
		expected      []int
		expectedError bool
	}{
		// Auto mode cases
		{
			name: "auto mode with 3 entries",
			config: map[string]interface{}{
				"sequence_start": 100,
				"sequence_step":  10,
				"entry": []interface{}{
					map[string]interface{}{"action": "pass"},
					map[string]interface{}{"action": "deny"},
					map[string]interface{}{"action": "pass"},
				},
			},
			expected: []int{100, 110, 120},
		},
		{
			name: "auto mode with default step",
			config: map[string]interface{}{
				"sequence_start": 50,
				"entry": []interface{}{
					map[string]interface{}{"action": "pass"},
					map[string]interface{}{"action": "deny"},
				},
			},
			expected: []int{50, 60},
		},
		{
			name: "auto mode with no entries",
			config: map[string]interface{}{
				"sequence_start": 100,
				"sequence_step":  10,
			},
			expected: []int{},
		},
		{
			name: "auto mode with empty entry list",
			config: map[string]interface{}{
				"sequence_start": 100,
				"sequence_step":  10,
				"entry":          []interface{}{},
			},
			expected: []int{},
		},
		// Manual mode cases (returns nil)
		{
			name:     "manual mode returns nil",
			config:   map[string]interface{}{},
			expected: nil,
		},
		{
			name: "manual mode with sequence_start=0 returns nil",
			config: map[string]interface{}{
				"sequence_start": 0,
				"entry": []interface{}{
					map[string]interface{}{"action": "pass", "sequence": 100},
				},
			},
			expected: nil,
		},
		// Error case
		{
			name: "auto mode with overflow",
			config: map[string]interface{}{
				"sequence_start": 65530,
				"sequence_step":  10,
				"entry": []interface{}{
					map[string]interface{}{"action": "pass"},
					map[string]interface{}{"action": "deny"},
					map[string]interface{}{"action": "pass"},
				},
			},
			expected:      nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, testSchemaForSequenceCalculator(), tt.config)
			result, err := CalculateSequencesForResourceData(d)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestConstants verifies that constants have expected values.
func TestConstants(t *testing.T) {
	t.Run("MaxSequence", func(t *testing.T) {
		assert.Equal(t, 65535, MaxSequence)
	})

	t.Run("MinSequence", func(t *testing.T) {
		assert.Equal(t, 1, MinSequence)
	})

	t.Run("DefaultSequenceStart", func(t *testing.T) {
		assert.Equal(t, 10, DefaultSequenceStart)
	})

	t.Run("DefaultSequenceStep", func(t *testing.T) {
		assert.Equal(t, 10, DefaultSequenceStep)
	})
}

// TestErrors verifies that error variables are defined correctly.
func TestErrors(t *testing.T) {
	t.Run("ErrInvalidStep is defined", func(t *testing.T) {
		assert.NotNil(t, ErrInvalidStep)
		assert.Contains(t, ErrInvalidStep.Error(), "step")
	})

	t.Run("ErrInvalidCount is defined", func(t *testing.T) {
		assert.NotNil(t, ErrInvalidCount)
		assert.Contains(t, ErrInvalidCount.Error(), "count")
	})

	t.Run("ErrInvalidStart is defined", func(t *testing.T) {
		assert.NotNil(t, ErrInvalidStart)
		assert.Contains(t, ErrInvalidStart.Error(), "start")
	})

	t.Run("ErrSequenceOverflow is defined", func(t *testing.T) {
		assert.NotNil(t, ErrSequenceOverflow)
		assert.Contains(t, ErrSequenceOverflow.Error(), "maximum")
	})
}

// TestValidateSequenceRange_ErrorMessages verifies that error messages contain useful information.
func TestValidateSequenceRange_ErrorMessages(t *testing.T) {
	tests := []struct {
		name             string
		start            int
		step             int
		count            int
		expectedContains []string
	}{
		{
			name:             "invalid start error contains start value",
			start:            0,
			step:             10,
			count:            1,
			expectedContains: []string{"0", "minimum"},
		},
		{
			name:             "invalid step error contains step value",
			start:            100,
			step:             0,
			count:            1,
			expectedContains: []string{"0"},
		},
		{
			name:             "invalid count error contains count value",
			start:            100,
			step:             10,
			count:            -5,
			expectedContains: []string{"-5"},
		},
		{
			name:             "overflow error contains relevant values",
			start:            65530,
			step:             10,
			count:            3,
			expectedContains: []string{"65530", "exceed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSequenceRange(tt.start, tt.step, tt.count)
			require.Error(t, err)

			errMsg := err.Error()
			for _, substr := range tt.expectedContains {
				assert.Contains(t, errMsg, substr,
					"error message should contain %q", substr)
			}
		})
	}
}

// TestCalculateSequences_LargeValues tests calculation with large but valid values.
func TestCalculateSequences_LargeValues(t *testing.T) {
	tests := []struct {
		name  string
		start int
		step  int
		count int
	}{
		{
			name:  "near max with step 1",
			start: MaxSequence - 9,
			step:  1,
			count: 10,
		},
		{
			name:  "large start single entry",
			start: MaxSequence,
			step:  1,
			count: 1,
		},
		{
			name:  "mid range large count",
			start: 50000,
			step:  1,
			count: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CalculateSequences(tt.start, tt.step, tt.count)
			require.NoError(t, err)
			require.Len(t, result, tt.count)

			// Verify first and last values
			assert.Equal(t, tt.start, result[0])
			if tt.count > 1 {
				expectedLast := tt.start + (tt.count-1)*tt.step
				assert.Equal(t, expectedLast, result[tt.count-1])
			}

			// Verify all values are within valid range
			for i, seq := range result {
				assert.GreaterOrEqual(t, seq, MinSequence, "sequence[%d]", i)
				assert.LessOrEqual(t, seq, MaxSequence, "sequence[%d]", i)
			}
		})
	}
}
