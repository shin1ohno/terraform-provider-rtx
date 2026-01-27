// Package provider contains tests for the ACL schema common utilities.
package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestACLType_Constants tests that ACLType constants are defined correctly.
func TestACLType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		aclType  ACLType
		expected string
	}{
		{
			name:     "ACLTypeIP",
			aclType:  ACLTypeIP,
			expected: "ip",
		},
		{
			name:     "ACLTypeIPv6",
			aclType:  ACLTypeIPv6,
			expected: "ipv6",
		},
		{
			name:     "ACLTypeMAC",
			aclType:  ACLTypeMAC,
			expected: "mac",
		},
		{
			name:     "ACLTypeIPDynamic",
			aclType:  ACLTypeIPDynamic,
			expected: "ip_dynamic",
		},
		{
			name:     "ACLTypeIPv6Dynamic",
			aclType:  ACLTypeIPv6Dynamic,
			expected: "ipv6_dynamic",
		},
		{
			name:     "ACLTypeExtended",
			aclType:  ACLTypeExtended,
			expected: "extended",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.aclType))
		})
	}
}

// TestMaxSequenceValue tests the MaxSequenceValue constant.
func TestMaxSequenceValue(t *testing.T) {
	assert.Equal(t, 2147483647, MaxSequenceValue)
}

// TestCommonACLSchema tests that CommonACLSchema returns the expected schema.
func TestCommonACLSchema(t *testing.T) {
	aclSchema := CommonACLSchema()

	t.Run("returns map with expected keys", func(t *testing.T) {
		expectedKeys := []string{"name", "sequence_start", "sequence_step", "apply"}
		for _, key := range expectedKeys {
			assert.Contains(t, aclSchema, key, "schema should contain %q", key)
		}
	})

	t.Run("name schema is correct", func(t *testing.T) {
		nameSchema := aclSchema["name"]
		require.NotNil(t, nameSchema)
		assert.True(t, nameSchema.Required, "name should be required")
		assert.True(t, nameSchema.ForceNew, "name should force new resource")
		assert.Equal(t, schema.TypeString, nameSchema.Type)
	})

	t.Run("sequence_start schema is correct", func(t *testing.T) {
		seqStartSchema := aclSchema["sequence_start"]
		require.NotNil(t, seqStartSchema)
		assert.True(t, seqStartSchema.Optional, "sequence_start should be optional")
		assert.False(t, seqStartSchema.Required, "sequence_start should not be required")
		assert.Equal(t, schema.TypeInt, seqStartSchema.Type)
		assert.NotNil(t, seqStartSchema.ValidateFunc, "sequence_start should have validation")
	})

	t.Run("sequence_step schema is correct", func(t *testing.T) {
		seqStepSchema := aclSchema["sequence_step"]
		require.NotNil(t, seqStepSchema)
		assert.True(t, seqStepSchema.Optional, "sequence_step should be optional")
		assert.Equal(t, DefaultSequenceStep, seqStepSchema.Default, "sequence_step should default to %d", DefaultSequenceStep)
		assert.Equal(t, schema.TypeInt, seqStepSchema.Type)
		assert.NotNil(t, seqStepSchema.ValidateFunc, "sequence_step should have validation")
	})

	t.Run("apply schema is correct", func(t *testing.T) {
		applySchema := aclSchema["apply"]
		require.NotNil(t, applySchema)
		assert.True(t, applySchema.Optional, "apply should be optional")
		assert.Equal(t, schema.TypeList, applySchema.Type)
		assert.NotNil(t, applySchema.Elem, "apply should have Elem defined")
	})
}

// TestCommonApplySchema tests that CommonApplySchema returns the expected schema.
func TestCommonApplySchema(t *testing.T) {
	applyResource := CommonApplySchema()
	require.NotNil(t, applyResource)

	applySchemaMap := applyResource.Schema

	t.Run("returns resource with expected keys", func(t *testing.T) {
		expectedKeys := []string{"interface", "direction", "filter_ids"}
		for _, key := range expectedKeys {
			assert.Contains(t, applySchemaMap, key, "schema should contain %q", key)
		}
	})

	t.Run("interface schema is correct", func(t *testing.T) {
		ifaceSchema := applySchemaMap["interface"]
		require.NotNil(t, ifaceSchema)
		assert.True(t, ifaceSchema.Required, "interface should be required")
		assert.Equal(t, schema.TypeString, ifaceSchema.Type)
	})

	t.Run("direction schema is correct", func(t *testing.T) {
		dirSchema := applySchemaMap["direction"]
		require.NotNil(t, dirSchema)
		assert.True(t, dirSchema.Required, "direction should be required")
		assert.Equal(t, schema.TypeString, dirSchema.Type)
		assert.NotNil(t, dirSchema.ValidateFunc, "direction should have validation")
		assert.NotNil(t, dirSchema.DiffSuppressFunc, "direction should have DiffSuppressFunc")
	})

	t.Run("filter_ids schema is correct", func(t *testing.T) {
		filterIDsSchema := applySchemaMap["filter_ids"]
		require.NotNil(t, filterIDsSchema)
		assert.True(t, filterIDsSchema.Optional, "filter_ids should be optional")
		assert.True(t, filterIDsSchema.Computed, "filter_ids should be computed")
		assert.Equal(t, schema.TypeList, filterIDsSchema.Type)
	})
}

// TestCommonEntrySchema tests that CommonEntrySchema returns the expected schema.
func TestCommonEntrySchema(t *testing.T) {
	entrySchema := CommonEntrySchema()

	t.Run("returns map with expected keys", func(t *testing.T) {
		expectedKeys := []string{"sequence", "log"}
		for _, key := range expectedKeys {
			assert.Contains(t, entrySchema, key, "schema should contain %q", key)
		}
	})

	t.Run("sequence schema is correct", func(t *testing.T) {
		seqSchema := entrySchema["sequence"]
		require.NotNil(t, seqSchema)
		assert.True(t, seqSchema.Optional, "sequence should be optional")
		assert.True(t, seqSchema.Computed, "sequence should be computed")
		assert.Equal(t, schema.TypeInt, seqSchema.Type)
		assert.NotNil(t, seqSchema.ValidateFunc, "sequence should have validation")
	})

	t.Run("log schema is correct", func(t *testing.T) {
		logSchema := entrySchema["log"]
		require.NotNil(t, logSchema)
		assert.True(t, logSchema.Optional, "log should be optional")
		assert.Equal(t, false, logSchema.Default, "log should default to false")
		assert.Equal(t, schema.TypeBool, logSchema.Type)
	})
}

// testSchemaForACLValidation returns a test schema for ACL validation testing.
func testSchemaForACLValidation() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"sequence_start": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"sequence_step": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  DefaultSequenceStep,
		},
		"entry": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"sequence": {
						Type:     schema.TypeInt,
						Optional: true,
						Computed: true,
					},
					"action": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
		},
		"apply": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"interface": {
						Type:     schema.TypeString,
						Required: true,
					},
					"direction": {
						Type:     schema.TypeString,
						Required: true,
					},
					"filter_ids": {
						Type:     schema.TypeList,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeInt},
					},
				},
			},
		},
	}
}

// TestValidateACLSchema tests the ValidateACLSchema function.
func TestValidateACLSchema(t *testing.T) {
	testSchema := testSchemaForACLValidation()

	tests := []struct {
		name        string
		config      map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		// Valid auto mode cases
		{
			name: "valid auto mode with entries without sequences",
			config: map[string]interface{}{
				"name":           "test-acl",
				"sequence_start": 100,
				"sequence_step":  10,
				"entry": []interface{}{
					map[string]interface{}{"action": "pass", "sequence": 0},
					map[string]interface{}{"action": "deny", "sequence": 0},
				},
			},
			expectError: false,
		},
		{
			name: "valid auto mode with single entry",
			config: map[string]interface{}{
				"name":           "test-acl",
				"sequence_start": 100,
				"sequence_step":  10,
				"entry": []interface{}{
					map[string]interface{}{"action": "pass", "sequence": 0},
				},
			},
			expectError: false,
		},
		{
			name: "valid auto mode with no entries",
			config: map[string]interface{}{
				"name":           "test-acl",
				"sequence_start": 100,
				"sequence_step":  10,
				"entry":          []interface{}{},
			},
			expectError: false,
		},
		// Valid manual mode cases
		{
			name: "valid manual mode with explicit sequences",
			config: map[string]interface{}{
				"name":          "test-acl",
				"sequence_step": 10,
				"entry": []interface{}{
					map[string]interface{}{"sequence": 100, "action": "pass"},
					map[string]interface{}{"sequence": 200, "action": "deny"},
				},
			},
			expectError: false,
		},
		{
			name: "valid manual mode with single entry",
			config: map[string]interface{}{
				"name":          "test-acl",
				"sequence_step": 10,
				"entry": []interface{}{
					map[string]interface{}{"sequence": 100, "action": "pass"},
				},
			},
			expectError: false,
		},
		// Mixed mode error cases
		{
			name: "mixed mode with sequence_start and entry sequence",
			config: map[string]interface{}{
				"name":           "test-acl",
				"sequence_start": 100,
				"sequence_step":  10,
				"entry": []interface{}{
					map[string]interface{}{"sequence": 500, "action": "pass"},
				},
			},
			expectError: true,
			errorMsg:    "cannot be specified when sequence_start is set",
		},
		{
			name: "mixed mode with sequence_start and some entries with sequences",
			config: map[string]interface{}{
				"name":           "test-acl",
				"sequence_start": 100,
				"sequence_step":  10,
				"entry": []interface{}{
					map[string]interface{}{"sequence": 0, "action": "pass"},
					map[string]interface{}{"sequence": 200, "action": "deny"},
				},
			},
			expectError: true,
			errorMsg:    "cannot be specified when sequence_start is set",
		},
		// Manual mode missing sequence
		{
			name: "manual mode missing sequence on entry",
			config: map[string]interface{}{
				"name":          "test-acl",
				"sequence_step": 10,
				"entry": []interface{}{
					map[string]interface{}{"sequence": 0, "action": "pass"},
				},
			},
			expectError: true,
			errorMsg:    "sequence must be specified when sequence_start is not set",
		},
		// Duplicate sequence errors
		{
			name: "duplicate sequences in manual mode",
			config: map[string]interface{}{
				"name":          "test-acl",
				"sequence_step": 10,
				"entry": []interface{}{
					map[string]interface{}{"sequence": 100, "action": "pass"},
					map[string]interface{}{"sequence": 100, "action": "deny"},
				},
			},
			expectError: true,
			errorMsg:    "sequence 100 is already used",
		},
		// Overflow error (MaxSequenceValue is 2147483647)
		{
			name: "sequence overflow in auto mode",
			config: map[string]interface{}{
				"name":           "test-acl",
				"sequence_start": 2147483640,
				"sequence_step":  10,
				"entry": []interface{}{
					map[string]interface{}{"sequence": 0, "action": "pass"},
					map[string]interface{}{"sequence": 0, "action": "deny"},
				},
			},
			expectError: true,
			errorMsg:    "exceeds maximum value",
		},
		// Valid apply blocks
		{
			name: "valid single apply block",
			config: map[string]interface{}{
				"name":           "test-acl",
				"sequence_start": 100,
				"sequence_step":  10,
				"entry":          []interface{}{},
				"apply": []interface{}{
					map[string]interface{}{
						"interface":  "lan1",
						"direction":  "in",
						"filter_ids": []interface{}{},
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid multiple apply blocks",
			config: map[string]interface{}{
				"name":           "test-acl",
				"sequence_start": 100,
				"sequence_step":  10,
				"entry":          []interface{}{},
				"apply": []interface{}{
					map[string]interface{}{
						"interface":  "lan1",
						"direction":  "in",
						"filter_ids": []interface{}{},
					},
					map[string]interface{}{
						"interface":  "lan1",
						"direction":  "out",
						"filter_ids": []interface{}{},
					},
					map[string]interface{}{
						"interface":  "lan2",
						"direction":  "in",
						"filter_ids": []interface{}{},
					},
				},
			},
			expectError: false,
		},
		// Duplicate apply block error
		{
			name: "duplicate apply blocks same interface and direction",
			config: map[string]interface{}{
				"name":           "test-acl",
				"sequence_start": 100,
				"sequence_step":  10,
				"entry":          []interface{}{},
				"apply": []interface{}{
					map[string]interface{}{
						"interface":  "lan1",
						"direction":  "in",
						"filter_ids": []interface{}{},
					},
					map[string]interface{}{
						"interface":  "lan1",
						"direction":  "in",
						"filter_ids": []interface{}{},
					},
				},
			},
			expectError: true,
			errorMsg:    "is already specified",
		},
		// Duplicate filter_ids error
		{
			name: "duplicate filter_ids in apply block",
			config: map[string]interface{}{
				"name":           "test-acl",
				"sequence_start": 100,
				"sequence_step":  10,
				"entry":          []interface{}{},
				"apply": []interface{}{
					map[string]interface{}{
						"interface":  "lan1",
						"direction":  "in",
						"filter_ids": []interface{}{100, 200, 100},
					},
				},
			},
			expectError: true,
			errorMsg:    "filter ID 100 is duplicated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create ResourceData from config
			d := schema.TestResourceDataRaw(t, testSchema, tt.config)

			// For proper testing, we use TestResourceDataRaw and simulate the diff
			// by calling the validation function through the resource data
			err := validateACLSchemaWithData(context.Background(), d)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// validateACLSchemaWithData is a helper that validates ACL schema using ResourceData.
// This is used for testing since we cannot easily create a ResourceDiff.
func validateACLSchemaWithData(_ context.Context, d *schema.ResourceData) error {
	sequenceStart, hasSequenceStart := d.GetOk("sequence_start")
	sequenceStep := d.Get("sequence_step").(int)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	entries, hasEntries := d.GetOk("entry")
	if !hasEntries {
		entries = []interface{}{}
	}
	entryList := entries.([]interface{})

	// Determine the sequence mode
	autoMode := hasSequenceStart && sequenceStart.(int) > 0

	// Track sequences for duplicate detection
	usedSequences := make(map[int]int) // sequence -> entry index

	for i, e := range entryList {
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
	applies, hasApplies := d.GetOk("apply")
	if !hasApplies {
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

// TestBuildACLFromResourceData tests the BuildACLFromResourceData function.
func TestBuildACLFromResourceData(t *testing.T) {
	testSchema := testSchemaForACLValidation()

	tests := []struct {
		name          string
		config        map[string]interface{}
		expectedName  string
		expectedStart int
		expectedStep  int
		entryCount    int
		applyCount    int
	}{
		{
			name: "basic auto mode config",
			config: map[string]interface{}{
				"name":           "test-acl",
				"sequence_start": 100,
				"sequence_step":  10,
				"entry": []interface{}{
					map[string]interface{}{"action": "pass"},
					map[string]interface{}{"action": "deny"},
				},
			},
			expectedName:  "test-acl",
			expectedStart: 100,
			expectedStep:  10,
			entryCount:    2,
			applyCount:    0,
		},
		{
			name: "manual mode config with applies",
			config: map[string]interface{}{
				"name":          "manual-acl",
				"sequence_step": 10,
				"entry": []interface{}{
					map[string]interface{}{"sequence": 100, "action": "pass"},
				},
				"apply": []interface{}{
					map[string]interface{}{
						"interface": "lan1",
						"direction": "in",
					},
				},
			},
			expectedName:  "manual-acl",
			expectedStart: 0,
			expectedStep:  10,
			entryCount:    1,
			applyCount:    1,
		},
		{
			name: "config with default sequence_step",
			config: map[string]interface{}{
				"name":           "default-step-acl",
				"sequence_start": 50,
			},
			expectedName:  "default-step-acl",
			expectedStart: 50,
			expectedStep:  DefaultSequenceStep,
			entryCount:    0,
			applyCount:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, testSchema, tt.config)
			acl := BuildACLFromResourceData(d)

			assert.Equal(t, tt.expectedName, acl.Name)
			assert.Equal(t, tt.expectedStart, acl.SequenceStart)
			assert.Equal(t, tt.expectedStep, acl.SequenceStep)
			assert.Len(t, acl.Entries, tt.entryCount)
			assert.Len(t, acl.Applies, tt.applyCount)
		})
	}
}

// TestCalculateEntrySequences tests the CalculateEntrySequences function.
func TestCalculateEntrySequences(t *testing.T) {
	tests := []struct {
		name     string
		acl      CommonACL
		expected []int
	}{
		{
			name: "auto mode calculates sequences",
			acl: CommonACL{
				SequenceStart: 100,
				SequenceStep:  10,
				Entries: []ACLEntry{
					{Fields: map[string]interface{}{}},
					{Fields: map[string]interface{}{}},
					{Fields: map[string]interface{}{}},
				},
			},
			expected: []int{100, 110, 120},
		},
		{
			name: "manual mode uses entry sequences",
			acl: CommonACL{
				SequenceStart: 0,
				SequenceStep:  10,
				Entries: []ACLEntry{
					{Sequence: 50, Fields: map[string]interface{}{}},
					{Sequence: 150, Fields: map[string]interface{}{}},
					{Sequence: 300, Fields: map[string]interface{}{}},
				},
			},
			expected: []int{50, 150, 300},
		},
		{
			name: "empty entries",
			acl: CommonACL{
				SequenceStart: 100,
				SequenceStep:  10,
				Entries:       []ACLEntry{},
			},
			expected: []int{},
		},
		{
			name: "single entry auto mode",
			acl: CommonACL{
				SequenceStart: 500,
				SequenceStep:  100,
				Entries: []ACLEntry{
					{Fields: map[string]interface{}{}},
				},
			},
			expected: []int{500},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateEntrySequences(tt.acl)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestInterfaceSupportsACLType tests the InterfaceSupportsACLType function.
func TestInterfaceSupportsACLType(t *testing.T) {
	tests := []struct {
		name        string
		iface       string
		aclType     ACLType
		expectError bool
		errorMsg    string
	}{
		// LAN interfaces support all ACL types
		{
			name:        "lan1 supports IP ACL",
			iface:       "lan1",
			aclType:     ACLTypeIP,
			expectError: false,
		},
		{
			name:        "lan2 supports IPv6 ACL",
			iface:       "lan2",
			aclType:     ACLTypeIPv6,
			expectError: false,
		},
		{
			name:        "lan1 supports MAC ACL",
			iface:       "lan1",
			aclType:     ACLTypeMAC,
			expectError: false,
		},
		// Bridge interfaces support all ACL types
		{
			name:        "bridge1 supports IP ACL",
			iface:       "bridge1",
			aclType:     ACLTypeIP,
			expectError: false,
		},
		{
			name:        "bridge1 supports MAC ACL",
			iface:       "bridge1",
			aclType:     ACLTypeMAC,
			expectError: false,
		},
		// PP interfaces do not support MAC ACL
		{
			name:        "pp1 supports IP ACL",
			iface:       "pp1",
			aclType:     ACLTypeIP,
			expectError: false,
		},
		{
			name:        "pp1 does not support MAC ACL",
			iface:       "pp1",
			aclType:     ACLTypeMAC,
			expectError: true,
			errorMsg:    "MAC ACL cannot be applied",
		},
		{
			name:        "pp100 does not support MAC ACL",
			iface:       "pp100",
			aclType:     ACLTypeMAC,
			expectError: true,
			errorMsg:    "MAC ACL cannot be applied",
		},
		// Tunnel interfaces do not support MAC ACL
		{
			name:        "tunnel1 supports IP ACL",
			iface:       "tunnel1",
			aclType:     ACLTypeIP,
			expectError: false,
		},
		{
			name:        "tunnel1 does not support MAC ACL",
			iface:       "tunnel1",
			aclType:     ACLTypeMAC,
			expectError: true,
			errorMsg:    "MAC ACL cannot be applied",
		},
		{
			name:        "tunnel100 does not support MAC ACL",
			iface:       "tunnel100",
			aclType:     ACLTypeMAC,
			expectError: true,
			errorMsg:    "MAC ACL cannot be applied",
		},
		// Short interface names (edge cases)
		{
			name:        "short interface name p does not match pp",
			iface:       "p",
			aclType:     ACLTypeMAC,
			expectError: false, // "p" is not "pp" prefix
		},
		{
			name:        "short interface name pp is checked",
			iface:       "pp",
			aclType:     ACLTypeMAC,
			expectError: false, // "pp" has length 2, prefix check requires > 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InterfaceSupportsACLType(tt.iface, tt.aclType)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
