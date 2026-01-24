package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Test schema for helper function testing
func testSchemaForFieldHelpers() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"bool_field": {
			Type:     schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"int_field": {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
		},
		"string_field": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"string_list_field": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"string_set_field": {
			Type:     schema.TypeSet,
			Optional: true,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
	}
}

func TestFieldHelpers_GetBoolValue(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		key      string
		expected bool
	}{
		{
			name: "true value in config",
			config: map[string]interface{}{
				"bool_field": true,
			},
			key:      "bool_field",
			expected: true,
		},
		{
			name: "false value in config",
			config: map[string]interface{}{
				"bool_field": false,
			},
			key:      "bool_field",
			expected: false,
		},
		{
			name:     "zero value when not set",
			config:   map[string]interface{}{},
			key:      "bool_field",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, testSchemaForFieldHelpers(), tt.config)
			result := GetBoolValue(d, tt.key)
			if result != tt.expected {
				t.Errorf("GetBoolValue() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestFieldHelpers_GetIntValue(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		key      string
		expected int
	}{
		{
			name: "positive value in config",
			config: map[string]interface{}{
				"int_field": 42,
			},
			key:      "int_field",
			expected: 42,
		},
		{
			name: "negative value in config",
			config: map[string]interface{}{
				"int_field": -10,
			},
			key:      "int_field",
			expected: -10,
		},
		{
			name: "zero value in config",
			config: map[string]interface{}{
				"int_field": 0,
			},
			key:      "int_field",
			expected: 0,
		},
		{
			name:     "zero value when not set",
			config:   map[string]interface{}{},
			key:      "int_field",
			expected: 0,
		},
		{
			name: "large value in config",
			config: map[string]interface{}{
				"int_field": 999999,
			},
			key:      "int_field",
			expected: 999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, testSchemaForFieldHelpers(), tt.config)
			result := GetIntValue(d, tt.key)
			if result != tt.expected {
				t.Errorf("GetIntValue() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestFieldHelpers_GetStringValue(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		key      string
		expected string
	}{
		{
			name: "non-empty string in config",
			config: map[string]interface{}{
				"string_field": "hello",
			},
			key:      "string_field",
			expected: "hello",
		},
		{
			name: "empty string in config",
			config: map[string]interface{}{
				"string_field": "",
			},
			key:      "string_field",
			expected: "",
		},
		{
			name:     "empty string when not set",
			config:   map[string]interface{}{},
			key:      "string_field",
			expected: "",
		},
		{
			name: "string with spaces in config",
			config: map[string]interface{}{
				"string_field": "  spaced  ",
			},
			key:      "string_field",
			expected: "  spaced  ",
		},
		{
			name: "unicode string in config",
			config: map[string]interface{}{
				"string_field": "日本語テスト",
			},
			key:      "string_field",
			expected: "日本語テスト",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, testSchemaForFieldHelpers(), tt.config)
			result := GetStringValue(d, tt.key)
			if result != tt.expected {
				t.Errorf("GetStringValue() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestFieldHelpers_GetStringListValue(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		key      string
		expected []string
	}{
		{
			name: "list with multiple values",
			config: map[string]interface{}{
				"string_list_field": []interface{}{"a", "b", "c"},
			},
			key:      "string_list_field",
			expected: []string{"a", "b", "c"},
		},
		{
			name: "list with single value",
			config: map[string]interface{}{
				"string_list_field": []interface{}{"only"},
			},
			key:      "string_list_field",
			expected: []string{"only"},
		},
		{
			name: "empty list",
			config: map[string]interface{}{
				"string_list_field": []interface{}{},
			},
			key:      "string_list_field",
			expected: []string{},
		},
		{
			name:     "nil when not set",
			config:   map[string]interface{}{},
			key:      "string_list_field",
			expected: []string{},
		},
		{
			name: "list with unicode strings",
			config: map[string]interface{}{
				"string_list_field": []interface{}{"日本語", "test", "テスト"},
			},
			key:      "string_list_field",
			expected: []string{"日本語", "test", "テスト"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, testSchemaForFieldHelpers(), tt.config)
			result := GetStringListValue(d, tt.key)

			if len(result) != len(tt.expected) {
				t.Errorf("GetStringListValue() length = %d, expected %d", len(result), len(tt.expected))
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("GetStringListValue()[%d] = %v, expected %v", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestFieldHelpers_GetStringListValue_Set(t *testing.T) {
	tests := []struct {
		name           string
		config         map[string]interface{}
		key            string
		expectedLength int
	}{
		{
			name: "set with multiple values",
			config: map[string]interface{}{
				"string_set_field": []interface{}{"a", "b", "c"},
			},
			key:            "string_set_field",
			expectedLength: 3,
		},
		{
			name: "set with single value",
			config: map[string]interface{}{
				"string_set_field": []interface{}{"only"},
			},
			key:            "string_set_field",
			expectedLength: 1,
		},
		{
			name: "empty set",
			config: map[string]interface{}{
				"string_set_field": []interface{}{},
			},
			key:            "string_set_field",
			expectedLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, testSchemaForFieldHelpers(), tt.config)
			result := GetStringListValue(d, tt.key)

			if len(result) != tt.expectedLength {
				t.Errorf("GetStringListValue() for set: length = %d, expected %d", len(result), tt.expectedLength)
			}
		})
	}
}

func TestFieldHelpers_BoolPtr(t *testing.T) {
	tests := []struct {
		name     string
		value    bool
		expected bool
	}{
		{
			name:     "true value",
			value:    true,
			expected: true,
		},
		{
			name:     "false value",
			value:    false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BoolPtr(tt.value)

			if result == nil {
				t.Error("BoolPtr() returned nil, expected pointer")
				return
			}

			if *result != tt.expected {
				t.Errorf("BoolPtr() = %v, expected %v", *result, tt.expected)
			}
		})
	}
}

func TestFieldHelpers_IntPtr(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected int
	}{
		{
			name:     "positive value",
			value:    42,
			expected: 42,
		},
		{
			name:     "negative value",
			value:    -10,
			expected: -10,
		},
		{
			name:     "zero value",
			value:    0,
			expected: 0,
		},
		{
			name:     "large value",
			value:    999999,
			expected: 999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IntPtr(tt.value)

			if result == nil {
				t.Error("IntPtr() returned nil, expected pointer")
				return
			}

			if *result != tt.expected {
				t.Errorf("IntPtr() = %v, expected %v", *result, tt.expected)
			}
		})
	}
}

func TestFieldHelpers_StringPtr(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "non-empty string",
			value:    "hello",
			expected: "hello",
		},
		{
			name:     "empty string",
			value:    "",
			expected: "",
		},
		{
			name:     "string with spaces",
			value:    "  spaced  ",
			expected: "  spaced  ",
		},
		{
			name:     "unicode string",
			value:    "日本語テスト",
			expected: "日本語テスト",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringPtr(tt.value)

			if result == nil {
				t.Error("StringPtr() returned nil, expected pointer")
				return
			}

			if *result != tt.expected {
				t.Errorf("StringPtr() = %v, expected %v", *result, tt.expected)
			}
		})
	}
}

// Test that pointer functions return independent pointers
func TestFieldHelpers_PtrIndependence(t *testing.T) {
	t.Run("BoolPtr independence", func(t *testing.T) {
		p1 := BoolPtr(true)
		p2 := BoolPtr(true)

		if p1 == p2 {
			t.Error("BoolPtr() should return independent pointers")
		}
	})

	t.Run("IntPtr independence", func(t *testing.T) {
		p1 := IntPtr(42)
		p2 := IntPtr(42)

		if p1 == p2 {
			t.Error("IntPtr() should return independent pointers")
		}
	})

	t.Run("StringPtr independence", func(t *testing.T) {
		p1 := StringPtr("hello")
		p2 := StringPtr("hello")

		if p1 == p2 {
			t.Error("StringPtr() should return independent pointers")
		}
	})
}
