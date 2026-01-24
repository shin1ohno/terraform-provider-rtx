package stateupgraders

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestNewUpgrader(t *testing.T) {
	schemaFunc := func() map[string]*schema.Schema {
		return map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		}
	}

	upgradeFunc := func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		return rawState, nil
	}

	upgrader := NewUpgrader(0, schemaFunc, upgradeFunc)

	assert.Equal(t, 0, upgrader.Version)
	assert.NotNil(t, upgrader.Type)
	assert.NotNil(t, upgrader.Upgrade)
}

func TestRenameAttribute(t *testing.T) {
	tests := []struct {
		name      string
		rawState  map[string]interface{}
		oldKey    string
		newKey    string
		expected  map[string]interface{}
		performed bool
	}{
		{
			name:      "rename existing attribute",
			rawState:  map[string]interface{}{"name": "test"},
			oldKey:    "name",
			newKey:    "username",
			expected:  map[string]interface{}{"username": "test"},
			performed: true,
		},
		{
			name:      "old key does not exist",
			rawState:  map[string]interface{}{"other": "value"},
			oldKey:    "name",
			newKey:    "username",
			expected:  map[string]interface{}{"other": "value"},
			performed: false,
		},
		{
			name:      "nil state",
			rawState:  nil,
			oldKey:    "name",
			newKey:    "username",
			expected:  nil,
			performed: false,
		},
		{
			name:      "empty state",
			rawState:  map[string]interface{}{},
			oldKey:    "name",
			newKey:    "username",
			expected:  map[string]interface{}{},
			performed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenameAttribute(tt.rawState, tt.oldKey, tt.newKey)
			assert.Equal(t, tt.performed, result)
			assert.Equal(t, tt.expected, tt.rawState)
		})
	}
}

func TestSetDefaultIfMissing(t *testing.T) {
	tests := []struct {
		name         string
		rawState     map[string]interface{}
		key          string
		defaultValue interface{}
		expected     map[string]interface{}
		wasSet       bool
	}{
		{
			name:         "key missing - set default",
			rawState:     map[string]interface{}{"other": "value"},
			key:          "enabled",
			defaultValue: true,
			expected:     map[string]interface{}{"other": "value", "enabled": true},
			wasSet:       true,
		},
		{
			name:         "key exists - no change",
			rawState:     map[string]interface{}{"enabled": false},
			key:          "enabled",
			defaultValue: true,
			expected:     map[string]interface{}{"enabled": false},
			wasSet:       false,
		},
		{
			name:         "nil state",
			rawState:     nil,
			key:          "enabled",
			defaultValue: true,
			expected:     nil,
			wasSet:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SetDefaultIfMissing(tt.rawState, tt.key, tt.defaultValue)
			assert.Equal(t, tt.wasSet, result)
			assert.Equal(t, tt.expected, tt.rawState)
		})
	}
}

func TestRemoveAttribute(t *testing.T) {
	tests := []struct {
		name     string
		rawState map[string]interface{}
		key      string
		expected map[string]interface{}
		removed  bool
	}{
		{
			name:     "remove existing attribute",
			rawState: map[string]interface{}{"name": "test", "other": "value"},
			key:      "name",
			expected: map[string]interface{}{"other": "value"},
			removed:  true,
		},
		{
			name:     "key does not exist",
			rawState: map[string]interface{}{"other": "value"},
			key:      "name",
			expected: map[string]interface{}{"other": "value"},
			removed:  false,
		},
		{
			name:     "nil state",
			rawState: nil,
			key:      "name",
			expected: nil,
			removed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveAttribute(tt.rawState, tt.key)
			assert.Equal(t, tt.removed, result)
			assert.Equal(t, tt.expected, tt.rawState)
		})
	}
}

func TestTransformAttribute(t *testing.T) {
	tests := []struct {
		name        string
		rawState    map[string]interface{}
		key         string
		transform   func(interface{}) interface{}
		expected    map[string]interface{}
		transformed bool
	}{
		{
			name:     "transform string to uppercase",
			rawState: map[string]interface{}{"name": "test"},
			key:      "name",
			transform: func(v interface{}) interface{} {
				if s, ok := v.(string); ok {
					return s + "_transformed"
				}
				return v
			},
			expected:    map[string]interface{}{"name": "test_transformed"},
			transformed: true,
		},
		{
			name:     "key does not exist",
			rawState: map[string]interface{}{"other": "value"},
			key:      "name",
			transform: func(v interface{}) interface{} {
				return v
			},
			expected:    map[string]interface{}{"other": "value"},
			transformed: false,
		},
		{
			name:     "nil value - no transform",
			rawState: map[string]interface{}{"name": nil},
			key:      "name",
			transform: func(v interface{}) interface{} {
				return "should_not_happen"
			},
			expected:    map[string]interface{}{"name": nil},
			transformed: false,
		},
		{
			name:     "nil state",
			rawState: nil,
			key:      "name",
			transform: func(v interface{}) interface{} {
				return v
			},
			expected:    nil,
			transformed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransformAttribute(tt.rawState, tt.key, tt.transform)
			assert.Equal(t, tt.transformed, result)
			assert.Equal(t, tt.expected, tt.rawState)
		})
	}
}

func TestCopyAttribute(t *testing.T) {
	tests := []struct {
		name      string
		rawState  map[string]interface{}
		sourceKey string
		destKey   string
		expected  map[string]interface{}
		copied    bool
	}{
		{
			name:      "copy existing attribute",
			rawState:  map[string]interface{}{"source": "value"},
			sourceKey: "source",
			destKey:   "dest",
			expected:  map[string]interface{}{"source": "value", "dest": "value"},
			copied:    true,
		},
		{
			name:      "source does not exist",
			rawState:  map[string]interface{}{"other": "value"},
			sourceKey: "source",
			destKey:   "dest",
			expected:  map[string]interface{}{"other": "value"},
			copied:    false,
		},
		{
			name:      "nil state",
			rawState:  nil,
			sourceKey: "source",
			destKey:   "dest",
			expected:  nil,
			copied:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CopyAttribute(tt.rawState, tt.sourceKey, tt.destKey)
			assert.Equal(t, tt.copied, result)
			assert.Equal(t, tt.expected, tt.rawState)
		})
	}
}

func TestMoveToNestedBlock(t *testing.T) {
	tests := []struct {
		name          string
		rawState      map[string]interface{}
		attrKey       string
		blockKey      string
		nestedAttrKey string
		expected      map[string]interface{}
		moved         bool
	}{
		{
			name:          "move attribute to new nested block",
			rawState:      map[string]interface{}{"timeout": 30},
			attrKey:       "timeout",
			blockKey:      "timeouts",
			nestedAttrKey: "create",
			expected: map[string]interface{}{
				"timeouts": []interface{}{
					map[string]interface{}{"create": 30},
				},
			},
			moved: true,
		},
		{
			name: "move attribute to existing nested block",
			rawState: map[string]interface{}{
				"timeout": 30,
				"timeouts": []interface{}{
					map[string]interface{}{"delete": 10},
				},
			},
			attrKey:       "timeout",
			blockKey:      "timeouts",
			nestedAttrKey: "create",
			expected: map[string]interface{}{
				"timeouts": []interface{}{
					map[string]interface{}{"delete": 10, "create": 30},
				},
			},
			moved: true,
		},
		{
			name:          "attribute does not exist",
			rawState:      map[string]interface{}{"other": "value"},
			attrKey:       "timeout",
			blockKey:      "timeouts",
			nestedAttrKey: "create",
			expected:      map[string]interface{}{"other": "value"},
			moved:         false,
		},
		{
			name:          "nil state",
			rawState:      nil,
			attrKey:       "timeout",
			blockKey:      "timeouts",
			nestedAttrKey: "create",
			expected:      nil,
			moved:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MoveToNestedBlock(tt.rawState, tt.attrKey, tt.blockKey, tt.nestedAttrKey)
			assert.Equal(t, tt.moved, result)
			assert.Equal(t, tt.expected, tt.rawState)
		})
	}
}

func TestExtractFromNestedBlock(t *testing.T) {
	tests := []struct {
		name          string
		rawState      map[string]interface{}
		blockKey      string
		nestedAttrKey string
		attrKey       string
		expected      map[string]interface{}
		extracted     bool
	}{
		{
			name: "extract from nested block - block becomes empty",
			rawState: map[string]interface{}{
				"timeouts": []interface{}{
					map[string]interface{}{"create": 30},
				},
			},
			blockKey:      "timeouts",
			nestedAttrKey: "create",
			attrKey:       "timeout",
			expected: map[string]interface{}{
				"timeout": 30,
			},
			extracted: true,
		},
		{
			name: "extract from nested block - block has other attrs",
			rawState: map[string]interface{}{
				"timeouts": []interface{}{
					map[string]interface{}{"create": 30, "delete": 10},
				},
			},
			blockKey:      "timeouts",
			nestedAttrKey: "create",
			attrKey:       "timeout",
			expected: map[string]interface{}{
				"timeout": 30,
				"timeouts": []interface{}{
					map[string]interface{}{"delete": 10},
				},
			},
			extracted: true,
		},
		{
			name:          "block does not exist",
			rawState:      map[string]interface{}{"other": "value"},
			blockKey:      "timeouts",
			nestedAttrKey: "create",
			attrKey:       "timeout",
			expected:      map[string]interface{}{"other": "value"},
			extracted:     false,
		},
		{
			name: "nested attribute does not exist",
			rawState: map[string]interface{}{
				"timeouts": []interface{}{
					map[string]interface{}{"delete": 10},
				},
			},
			blockKey:      "timeouts",
			nestedAttrKey: "create",
			attrKey:       "timeout",
			expected: map[string]interface{}{
				"timeouts": []interface{}{
					map[string]interface{}{"delete": 10},
				},
			},
			extracted: false,
		},
		{
			name:          "nil state",
			rawState:      nil,
			blockKey:      "timeouts",
			nestedAttrKey: "create",
			attrKey:       "timeout",
			expected:      nil,
			extracted:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractFromNestedBlock(tt.rawState, tt.blockKey, tt.nestedAttrKey, tt.attrKey)
			assert.Equal(t, tt.extracted, result)
			assert.Equal(t, tt.expected, tt.rawState)
		})
	}
}
