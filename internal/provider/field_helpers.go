package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// GetBoolValue returns the merged value from ResourceData for a bool field.
// If the field is set in the config, it returns the config value.
// If the field is not set in config but exists in state, it returns the state value.
// This behavior is automatic in Terraform's d.Get() method.
func GetBoolValue(d *schema.ResourceData, key string) bool {
	return d.Get(key).(bool)
}

// GetIntValue returns the merged value from ResourceData for an int field.
// If the field is set in the config, it returns the config value.
// If the field is not set in config but exists in state, it returns the state value.
// This behavior is automatic in Terraform's d.Get() method.
func GetIntValue(d *schema.ResourceData, key string) int {
	return d.Get(key).(int)
}

// GetStringValue returns the merged value from ResourceData for a string field.
// If the field is set in the config, it returns the config value.
// If the field is not set in config but exists in state, it returns the state value.
// This behavior is automatic in Terraform's d.Get() method.
func GetStringValue(d *schema.ResourceData, key string) string {
	return d.Get(key).(string)
}

// GetStringListValue returns the merged value from ResourceData for a string list/set field.
// If the field is set in the config, it returns the config value.
// If the field is not set in config but exists in state, it returns the state value.
// This function handles both TypeList and TypeSet with string elements.
func GetStringListValue(d *schema.ResourceData, key string) []string {
	rawValue := d.Get(key)

	// Handle TypeSet
	if set, ok := rawValue.(*schema.Set); ok {
		list := set.List()
		result := make([]string, len(list))
		for i, v := range list {
			result[i] = v.(string)
		}
		return result
	}

	// Handle TypeList ([]interface{})
	if list, ok := rawValue.([]interface{}); ok {
		result := make([]string, len(list))
		for i, v := range list {
			result[i] = v.(string)
		}
		return result
	}

	return nil
}

// BoolPtr returns a pointer to the given bool value.
// This is useful for creating optional bool fields in API structs
// where nil means "not specified".
func BoolPtr(v bool) *bool {
	return &v
}

// IntPtr returns a pointer to the given int value.
// This is useful for creating optional int fields in API structs
// where nil means "not specified".
func IntPtr(v int) *int {
	return &v
}

// StringPtr returns a pointer to the given string value.
// This is useful for creating optional string fields in API structs
// where nil means "not specified".
func StringPtr(v string) *string {
	return &v
}
