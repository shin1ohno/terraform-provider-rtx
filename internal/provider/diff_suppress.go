// Package provider contains Terraform provider implementation for RTX routers.
package provider

import (
	"encoding/json"
	"net"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// DiffSuppressFunc Library
//
// This file provides reusable diff suppression functions for Terraform schema definitions.
// These functions allow semantic equality comparison, suppressing irrelevant diffs when
// values are functionally equivalent but syntactically different.
//
// Usage in schema definitions:
//
//	"field_name": {
//	    Type:             schema.TypeString,
//	    Optional:         true,
//	    DiffSuppressFunc: SuppressCaseDiff,
//	},
//
// Note: When migrating to terraform-plugin-framework, these functions should be
// replaced with custom types implementing SemanticEquals() method.

// SuppressCaseDiff ignores case differences when comparing string values.
// This is useful for fields where the router/API normalizes case differently
// than the user input (e.g., hostnames, protocol names).
//
// Examples:
//   - "TCP" and "tcp" are considered equal
//   - "MyHost" and "myhost" are considered equal
//   - "" and "" are considered equal
//
// Returns true if values are equal ignoring case, false otherwise.
func SuppressCaseDiff(k, old, new string, d *schema.ResourceData) bool {
	return strings.EqualFold(old, new)
}

// SuppressWhitespaceDiff ignores leading and trailing whitespace differences.
// This is useful for fields where whitespace may be trimmed by the router/API
// or where users may accidentally include extra whitespace.
//
// Examples:
//   - "  value  " and "value" are considered equal
//   - "\tvalue\n" and "value" are considered equal
//   - "a b" and "a b" are considered equal (internal whitespace preserved)
//
// Returns true if values are equal after trimming whitespace, false otherwise.
func SuppressWhitespaceDiff(k, old, new string, d *schema.ResourceData) bool {
	return strings.TrimSpace(old) == strings.TrimSpace(new)
}

// SuppressJSONDiff compares JSON strings semantically rather than syntactically.
// This is useful for JSON configuration fields where key ordering or formatting
// may differ but the data is functionally identical.
//
// Examples:
//   - `{"a":1,"b":2}` and `{"b":2,"a":1}` are considered equal
//   - `{"a": 1}` and `{"a":1}` are considered equal (whitespace ignored)
//   - `[1,2,3]` and `[1, 2, 3]` are considered equal
//
// Returns true if values are semantically equal JSON, false if:
//   - Values differ semantically
//   - Either value is not valid JSON (safe fallback to show diff)
//   - Either value is empty (both empty = true)
func SuppressJSONDiff(k, old, new string, d *schema.ResourceData) bool {
	// Handle empty strings - both empty is equal, one empty is different
	if old == "" && new == "" {
		return true
	}
	if old == "" || new == "" {
		return false
	}

	var oldJSON, newJSON interface{}

	if err := json.Unmarshal([]byte(old), &oldJSON); err != nil {
		// Old value is not valid JSON - return false (safe fallback to show diff)
		return false
	}

	if err := json.Unmarshal([]byte(new), &newJSON); err != nil {
		// New value is not valid JSON - return false (safe fallback to show diff)
		return false
	}

	return reflect.DeepEqual(oldJSON, newJSON)
}

// SuppressEquivalentIPDiff compares IP addresses accounting for different representations.
// This is useful for IP address fields where the format may vary but represent
// the same address.
//
// Examples:
//   - "192.168.1.1" and "192.168.001.001" may be parsed as equal
//   - "::1" and "0:0:0:0:0:0:0:1" are considered equal (IPv6 equivalence)
//   - "192.168.1.1" and "192.168.1.1" are considered equal
//
// Returns true if values represent the same IP address, false if:
//   - Values represent different IP addresses
//   - Either value is not a valid IP address (falls back to string comparison)
//   - Both values are empty strings (considered equal)
func SuppressEquivalentIPDiff(k, old, new string, d *schema.ResourceData) bool {
	// Handle empty strings
	if old == "" && new == "" {
		return true
	}
	if old == "" || new == "" {
		return false
	}

	oldIP := net.ParseIP(old)
	newIP := net.ParseIP(new)

	// If either is not a valid IP, fall back to string comparison
	if oldIP == nil || newIP == nil {
		return old == new
	}

	return oldIP.Equal(newIP)
}

// SuppressCaseAndWhitespaceDiff combines case-insensitive and whitespace normalization.
// This is useful for fields where both case and whitespace may vary.
//
// Examples:
//   - "  VALUE  " and "value" are considered equal
//   - "MyHost" and "  myhost\n" are considered equal
//
// Returns true if values are equal after trimming whitespace and ignoring case.
func SuppressCaseAndWhitespaceDiff(k, old, new string, d *schema.ResourceData) bool {
	return strings.EqualFold(strings.TrimSpace(old), strings.TrimSpace(new))
}

// SuppressEquivalentCIDRDiff compares CIDR notation network addresses.
// This normalizes the network address portion while preserving the prefix length.
//
// Examples:
//   - "192.168.1.0/24" and "192.168.1.0/24" are considered equal
//   - "192.168.1.5/24" and "192.168.1.0/24" may differ (different host portions)
//   - "10.0.0.0/8" and "10.0.0.0/8" are considered equal
//
// Returns true if values represent the same CIDR block, false if:
//   - Values represent different networks
//   - Either value is not a valid CIDR notation (falls back to string comparison)
func SuppressEquivalentCIDRDiff(k, old, new string, d *schema.ResourceData) bool {
	// Handle empty strings
	if old == "" && new == "" {
		return true
	}
	if old == "" || new == "" {
		return false
	}

	_, oldNet, oldErr := net.ParseCIDR(old)
	_, newNet, newErr := net.ParseCIDR(new)

	// If either is not a valid CIDR, fall back to string comparison
	if oldErr != nil || newErr != nil {
		return old == new
	}

	// Compare the network address and mask
	return oldNet.String() == newNet.String()
}

// SuppressBooleanStringDiff compares boolean-like string values.
// This handles various representations of true/false values.
//
// Examples:
//   - "true" and "True" and "TRUE" are considered equal
//   - "false" and "False" and "FALSE" are considered equal
//   - "yes" and "true" are NOT considered equal (explicit matching only)
//
// Returns true if values represent the same boolean, false otherwise.
func SuppressBooleanStringDiff(k, old, new string, d *schema.ResourceData) bool {
	return strings.EqualFold(old, new)
}

// SuppressEmptyStringDiff treats empty string and missing value as equivalent.
// This is useful when the API returns empty string but Terraform config has no value.
//
// Examples:
//   - "" (empty) and "" (empty) are considered equal
//
// Note: This function only compares the old and new string values directly.
// For null vs empty distinction, use d.GetOk() in CRUD functions instead.
func SuppressEmptyStringDiff(k, old, new string, d *schema.ResourceData) bool {
	return old == new
}

// SuppressMACAddressWhenAnyIsTrue suppresses diff for MAC address fields
// when the corresponding *_any field is true. This handles the case where
// RTX router normalizes "*:*:*:*:*:*" to *_any=true internally.
//
// The key format is expected to be like "entry.0.source_address" or
// "entry.0.destination_address". This function extracts the entry index
// and checks the corresponding source_any or destination_any field.
//
// Examples:
//   - If source_any=true, source_address diff is suppressed
//   - If destination_any=true, destination_address diff is suppressed
//   - Wildcard address "*:*:*:*:*:*" with any=false shows diff
//
// Returns true if diff should be suppressed, false otherwise.
func SuppressMACAddressWhenAnyIsTrue(k, old, new string, d *schema.ResourceData) bool {
	// Extract the base path to find the corresponding *_any field
	// Key format: "entry.0.source_address" or "entry.0.destination_address"
	parts := strings.Split(k, ".")
	if len(parts) < 3 {
		return old == new
	}

	// Determine which field we're checking
	fieldName := parts[len(parts)-1]
	basePath := strings.Join(parts[:len(parts)-1], ".")

	var anyFieldPath string
	if fieldName == "source_address" {
		anyFieldPath = basePath + ".source_any"
	} else if fieldName == "destination_address" {
		anyFieldPath = basePath + ".destination_any"
	} else {
		return old == new
	}

	// Check if the corresponding *_any field is true
	if anyVal, ok := d.GetOk(anyFieldPath); ok && anyVal.(bool) {
		// *_any is true, suppress diff for the address field
		return true
	}

	// Also treat wildcard address as equivalent to *_any=true
	wildcardMAC := "*:*:*:*:*:*"
	if old == wildcardMAC || new == wildcardMAC {
		// If either is wildcard and either *_any or address matches wildcard semantically
		if (old == wildcardMAC && new == "") || (old == "" && new == wildcardMAC) {
			return true
		}
		if old == wildcardMAC && new == wildcardMAC {
			return true
		}
	}

	return old == new
}

// SuppressEquivalentACLActionDiff compares ACL action values accounting for
// equivalent representations. RTX routers may normalize action values differently
// than Terraform configuration.
//
// Equivalent action groups:
//   - "permit", "pass", "pass-nolog" are all equivalent (allow without logging)
//   - "deny", "reject", "reject-nolog" are all equivalent (deny without logging)
//   - "pass-log" stands alone (allow with logging)
//   - "reject-log" stands alone (deny with logging)
//
// Examples:
//   - "permit" and "pass" are considered equal
//   - "pass", "pass-nolog", and "pass-log" are considered equal
//   - "reject", "reject-nolog", and "reject-log" are considered equal
//   - The router may normalize action values and add/remove logging suffixes
//
// Returns true if values represent the same action semantically.
func SuppressEquivalentACLActionDiff(k, old, new string, d *schema.ResourceData) bool {
	// Normalize both values to lowercase for comparison
	oldLower := strings.ToLower(strings.TrimSpace(old))
	newLower := strings.ToLower(strings.TrimSpace(new))

	// Direct match (including case-insensitive)
	if oldLower == newLower {
		return true
	}

	// Define equivalent action groups
	// Note: All pass/permit variants are equivalent, and all reject/deny variants are equivalent.
	// The router may normalize these values and add/remove logging suffixes.
	permitGroup := map[string]bool{"permit": true, "pass": true, "pass-nolog": true, "pass-log": true}
	denyGroup := map[string]bool{"deny": true, "reject": true, "reject-nolog": true, "reject-log": true}

	// Check if both are in the same equivalence group
	if permitGroup[oldLower] && permitGroup[newLower] {
		return true
	}
	if denyGroup[oldLower] && denyGroup[newLower] {
		return true
	}

	return false
}
