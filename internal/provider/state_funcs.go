// Package provider provides StateFunc normalizers for Terraform schema attributes.
//
// StateFunc normalizers are called when storing values in state to ensure consistent
// canonical representations. This prevents unnecessary diffs when values are semantically
// equivalent but syntactically different.
//
// Usage example:
//
//	"ip_address": {
//	    Type:             schema.TypeString,
//	    Optional:         true,
//	    StateFunc:        normalizeIPAddress,
//	    DiffSuppressFunc: SuppressEquivalentIPDiff,
//	    ValidateFunc:     validation.IsIPAddress,
//	},
package provider

import (
	"encoding/json"
	"net"
	"strings"
)

// normalizeIPAddress converts an IP address to its canonical form.
// For IPv4 addresses, this ensures dotted-decimal notation without leading zeros.
// For IPv6 addresses, this returns the standard string representation.
//
// If the input is not a valid IP address, it is returned unchanged.
// This ensures the StateFunc never fails and allows validation to handle errors.
//
// Examples:
//   - "192.168.001.001" -> "192.168.1.1"
//   - "::ffff:192.168.1.1" -> "192.168.1.1" (IPv4-mapped IPv6 to IPv4)
//   - "2001:db8::1" -> "2001:db8::1"
//   - "invalid" -> "invalid" (returned unchanged)
func normalizeIPAddress(val interface{}) string {
	s, ok := val.(string)
	if !ok {
		// Return empty string for non-string types
		return ""
	}

	// Trim leading and trailing whitespace before processing
	s = strings.TrimSpace(s)

	// Empty string is returned as-is
	if s == "" {
		return s
	}

	ip := net.ParseIP(s)
	if ip == nil {
		// Invalid IP address, return original value
		// Let validation handle the error
		return s
	}

	// Check if it's an IPv4 address (or IPv4-mapped IPv6)
	if ip4 := ip.To4(); ip4 != nil {
		return ip4.String()
	}

	// Return canonical IPv6 form
	return ip.String()
}

// normalizeLowercase converts a string to lowercase for case-insensitive storage.
// This is useful for attributes where case differences should not cause diffs.
//
// If the input is not a string, an empty string is returned.
// Leading and trailing whitespace is preserved.
//
// Examples:
//   - "HelloWorld" -> "helloworld"
//   - "UPPER" -> "upper"
//   - "  Mixed Case  " -> "  mixed case  "
//   - "" -> ""
func normalizeLowercase(val interface{}) string {
	s, ok := val.(string)
	if !ok {
		return ""
	}
	return strings.ToLower(s)
}

// normalizeJSON parses and re-marshals JSON to ensure consistent formatting.
// This removes differences in key ordering, whitespace, and formatting.
//
// If the input is not valid JSON, it is returned unchanged.
// This ensures the StateFunc never fails and allows validation to handle errors.
//
// The output uses compact JSON format (no indentation or extra whitespace).
//
// Examples:
//   - `{"b":2,"a":1}` -> `{"a":1,"b":2}` (keys sorted)
//   - `{ "key" : "value" }` -> `{"key":"value"}` (whitespace normalized)
//   - `[1, 2, 3]` -> `[1,2,3]`
//   - `not json` -> `not json` (returned unchanged)
func normalizeJSON(val interface{}) string {
	s, ok := val.(string)
	if !ok {
		return ""
	}

	// Empty string is returned as-is
	if s == "" {
		return s
	}

	// Try to parse the JSON
	var parsed interface{}
	if err := json.Unmarshal([]byte(s), &parsed); err != nil {
		// Invalid JSON, return original value
		// Let validation handle the error
		return s
	}

	// Re-marshal to get consistent formatting
	// json.Marshal produces deterministic output with sorted keys for maps
	normalized, err := json.Marshal(parsed)
	if err != nil {
		// This should never happen if Unmarshal succeeded
		return s
	}

	return string(normalized)
}

// normalizeTrimmedString removes leading and trailing whitespace from a string.
// This is useful for attributes where whitespace differences should not cause diffs.
//
// If the input is not a string, an empty string is returned.
//
// Examples:
//   - "  hello  " -> "hello"
//   - "\t\nvalue\r\n" -> "value"
//   - "" -> ""
func normalizeTrimmedString(val interface{}) string {
	s, ok := val.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

// normalizeUppercase converts a string to uppercase for case-insensitive storage.
// This is useful for attributes like MAC addresses where uppercase is the convention.
//
// If the input is not a string, an empty string is returned.
// Leading and trailing whitespace is preserved.
//
// Examples:
//   - "aa:bb:cc:dd:ee:ff" -> "AA:BB:CC:DD:EE:FF"
//   - "lower" -> "LOWER"
//   - "" -> ""
func normalizeUppercase(val interface{}) string {
	s, ok := val.(string)
	if !ok {
		return ""
	}
	return strings.ToUpper(s)
}
