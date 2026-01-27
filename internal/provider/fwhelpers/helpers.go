package fwhelpers

import (
	"context"

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
