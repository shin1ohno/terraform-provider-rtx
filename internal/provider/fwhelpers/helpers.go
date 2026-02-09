package fwhelpers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringValueOrNull returns a types.String from a string value.
// If the value is empty, returns a null string.
func StringValueOrNull(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}

// Int64ValueOrNull returns a types.Int64 from an int value.
// If the value is 0, returns a null int64.
func Int64ValueOrNull(i int) types.Int64 {
	if i == 0 {
		return types.Int64Null()
	}
	return types.Int64Value(int64(i))
}

// BoolValue returns a types.Bool from a bool value.
func BoolValue(b bool) types.Bool {
	return types.BoolValue(b)
}

// GetStringValue extracts a string from a types.String.
// Returns empty string if null or unknown.
func GetStringValue(s types.String) string {
	if s.IsNull() || s.IsUnknown() {
		return ""
	}
	return s.ValueString()
}

// GetInt64Value extracts an int from a types.Int64.
// Returns 0 if null or unknown.
func GetInt64Value(i types.Int64) int {
	if i.IsNull() || i.IsUnknown() {
		return 0
	}
	return int(i.ValueInt64())
}

// GetBoolValue extracts a bool from a types.Bool.
// Returns false if null or unknown.
func GetBoolValue(b types.Bool) bool {
	if b.IsNull() || b.IsUnknown() {
		return false
	}
	return b.ValueBool()
}

// GetBoolValueWithDefault extracts a bool from a types.Bool.
// Returns the provided default if null or unknown.
func GetBoolValueWithDefault(b types.Bool, defaultValue bool) bool {
	if b.IsNull() || b.IsUnknown() {
		return defaultValue
	}
	return b.ValueBool()
}

// GetStringValueWithDefault extracts a string from a types.String.
// Returns the provided default if null or unknown.
func GetStringValueWithDefault(s types.String, defaultValue string) string {
	if s.IsNull() || s.IsUnknown() {
		return defaultValue
	}
	return s.ValueString()
}

// SetStringListValue converts a []string to []types.String.
func SetStringListValue(values []string) []types.String {
	result := make([]types.String, len(values))
	for i, v := range values {
		result[i] = types.StringValue(v)
	}
	return result
}

// GetStringListValue extracts []string from []types.String.
func GetStringListValue(values []types.String) []string {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = GetStringValue(v)
	}
	return result
}

// IntSliceToList converts []int to types.List, preserving nil vs empty distinction.
// nil → ListNull, []int{} → empty ListValue, populated → ListValue with elements.
func IntSliceToList(ints []int) types.List {
	if ints == nil {
		return types.ListNull(types.Int64Type)
	}

	elements := make([]attr.Value, len(ints))
	for i, v := range ints {
		elements[i] = types.Int64Value(int64(v))
	}

	list, _ := types.ListValue(types.Int64Type, elements)
	return list
}

// ListToIntSlice converts types.List to []int, preserving null vs empty distinction.
// null/unknown → nil, empty list → []int{} (NOT nil), populated → []int with values.
func ListToIntSlice(list types.List) []int {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	elements := list.Elements()
	result := make([]int, 0, len(elements))
	for _, elem := range elements {
		if int64Val, ok := elem.(types.Int64); ok && !int64Val.IsNull() && !int64Val.IsUnknown() {
			result = append(result, int(int64Val.ValueInt64()))
		}
	}

	return result
}

// StringSliceToList converts []string to types.List, preserving nil vs empty distinction.
// nil → ListNull, []string{} → empty ListValue, populated → ListValue with elements.
func StringSliceToList(strs []string) types.List {
	if strs == nil {
		return types.ListNull(types.StringType)
	}

	elements := make([]attr.Value, len(strs))
	for i, v := range strs {
		elements[i] = types.StringValue(v)
	}

	list, _ := types.ListValue(types.StringType, elements)
	return list
}

// ListToStringSlice converts types.List to []string, preserving null vs empty distinction.
// null/unknown → nil, empty list → []string{} (NOT nil), populated → []string with values.
func ListToStringSlice(list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	elements := list.Elements()
	result := make([]string, 0, len(elements))
	for _, elem := range elements {
		if strVal, ok := elem.(types.String); ok && !strVal.IsNull() && !strVal.IsUnknown() {
			result = append(result, strVal.ValueString())
		}
	}

	return result
}

// AppendDiagError is a helper to append an error diagnostic.
func AppendDiagError(diags *diag.Diagnostics, summary, detail string) {
	diags.AddError(summary, detail)
}

// AppendDiagWarning is a helper to append a warning diagnostic.
func AppendDiagWarning(diags *diag.Diagnostics, summary, detail string) {
	diags.AddWarning(summary, detail)
}

// LogDebug logs a debug message using the context's logger.
func LogDebug(ctx context.Context, msg string) {
	// Use Terraform's logging via tflog if needed
	// For now, this is a placeholder for structured logging
}
