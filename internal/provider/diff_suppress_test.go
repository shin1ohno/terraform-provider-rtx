// Package provider contains Terraform provider implementation for RTX routers.
package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSuppressCaseDiff tests case-insensitive string comparison.
func TestSuppressCaseDiff(t *testing.T) {
	tests := []struct {
		name     string
		old      string
		new      string
		expected bool
	}{
		// Equivalent values (should suppress diff)
		{
			name:     "same case",
			old:      "value",
			new:      "value",
			expected: true,
		},
		{
			name:     "different case - uppercase vs lowercase",
			old:      "VALUE",
			new:      "value",
			expected: true,
		},
		{
			name:     "different case - mixed case",
			old:      "MyHost",
			new:      "myhost",
			expected: true,
		},
		{
			name:     "protocol name case difference",
			old:      "TCP",
			new:      "tcp",
			expected: true,
		},
		{
			name:     "both empty",
			old:      "",
			new:      "",
			expected: true,
		},
		// Different values (should not suppress diff)
		{
			name:     "different values",
			old:      "value1",
			new:      "value2",
			expected: false,
		},
		{
			name:     "empty vs non-empty",
			old:      "",
			new:      "value",
			expected: false,
		},
		{
			name:     "non-empty vs empty",
			old:      "value",
			new:      "",
			expected: false,
		},
		// Edge cases
		{
			name:     "unicode characters",
			old:      "日本語",
			new:      "日本語",
			expected: true,
		},
		{
			name:     "special characters",
			old:      "test@example.com",
			new:      "TEST@EXAMPLE.COM",
			expected: true,
		},
		{
			name:     "numbers and letters",
			old:      "User123",
			new:      "user123",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SuppressCaseDiff("test_key", tt.old, tt.new, nil)
			assert.Equal(t, tt.expected, result, "SuppressCaseDiff(%q, %q)", tt.old, tt.new)
		})
	}
}

// TestSuppressWhitespaceDiff tests whitespace normalization comparison.
func TestSuppressWhitespaceDiff(t *testing.T) {
	tests := []struct {
		name     string
		old      string
		new      string
		expected bool
	}{
		// Equivalent values (should suppress diff)
		{
			name:     "same value no whitespace",
			old:      "value",
			new:      "value",
			expected: true,
		},
		{
			name:     "leading whitespace difference",
			old:      "  value",
			new:      "value",
			expected: true,
		},
		{
			name:     "trailing whitespace difference",
			old:      "value  ",
			new:      "value",
			expected: true,
		},
		{
			name:     "both leading and trailing whitespace",
			old:      "  value  ",
			new:      "value",
			expected: true,
		},
		{
			name:     "tab and newline whitespace",
			old:      "\tvalue\n",
			new:      "value",
			expected: true,
		},
		{
			name:     "both have different whitespace",
			old:      "  value  ",
			new:      "\tvalue\n",
			expected: true,
		},
		{
			name:     "both empty",
			old:      "",
			new:      "",
			expected: true,
		},
		{
			name:     "whitespace only vs empty",
			old:      "   ",
			new:      "",
			expected: true,
		},
		// Different values (should not suppress diff)
		{
			name:     "different values",
			old:      "value1",
			new:      "value2",
			expected: false,
		},
		{
			name:     "internal whitespace preserved - different",
			old:      "a b",
			new:      "a  b",
			expected: false,
		},
		{
			name:     "empty vs non-empty",
			old:      "",
			new:      "value",
			expected: false,
		},
		// Edge cases
		{
			name:     "internal whitespace preserved - same",
			old:      "a b",
			new:      "a b",
			expected: true,
		},
		{
			name:     "mixed whitespace characters",
			old:      "\t\n\r value \t\n\r",
			new:      "value",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SuppressWhitespaceDiff("test_key", tt.old, tt.new, nil)
			assert.Equal(t, tt.expected, result, "SuppressWhitespaceDiff(%q, %q)", tt.old, tt.new)
		})
	}
}

// TestSuppressJSONDiff tests semantic JSON comparison.
func TestSuppressJSONDiff(t *testing.T) {
	tests := []struct {
		name     string
		old      string
		new      string
		expected bool
	}{
		// Equivalent values (should suppress diff)
		{
			name:     "identical JSON",
			old:      `{"a":1,"b":2}`,
			new:      `{"a":1,"b":2}`,
			expected: true,
		},
		{
			name:     "different key order",
			old:      `{"a":1,"b":2}`,
			new:      `{"b":2,"a":1}`,
			expected: true,
		},
		{
			name:     "whitespace difference",
			old:      `{"a": 1, "b": 2}`,
			new:      `{"a":1,"b":2}`,
			expected: true,
		},
		{
			name:     "pretty printed vs compact",
			old:      "{\n  \"a\": 1,\n  \"b\": 2\n}",
			new:      `{"a":1,"b":2}`,
			expected: true,
		},
		{
			name:     "array with same elements",
			old:      `[1,2,3]`,
			new:      `[1, 2, 3]`,
			expected: true,
		},
		{
			name:     "nested object",
			old:      `{"outer":{"inner":1}}`,
			new:      `{"outer": {"inner": 1}}`,
			expected: true,
		},
		{
			name:     "both empty strings",
			old:      "",
			new:      "",
			expected: true,
		},
		// Different values (should not suppress diff)
		{
			name:     "different values",
			old:      `{"a":1}`,
			new:      `{"a":2}`,
			expected: false,
		},
		{
			name:     "missing key",
			old:      `{"a":1,"b":2}`,
			new:      `{"a":1}`,
			expected: false,
		},
		{
			name:     "extra key",
			old:      `{"a":1}`,
			new:      `{"a":1,"b":2}`,
			expected: false,
		},
		{
			name:     "array order matters",
			old:      `[1,2,3]`,
			new:      `[3,2,1]`,
			expected: false,
		},
		{
			name:     "empty vs non-empty",
			old:      "",
			new:      `{"a":1}`,
			expected: false,
		},
		{
			name:     "non-empty vs empty",
			old:      `{"a":1}`,
			new:      "",
			expected: false,
		},
		// Invalid JSON (should not suppress diff - safe fallback)
		{
			name:     "old is invalid JSON",
			old:      "not json",
			new:      `{"a":1}`,
			expected: false,
		},
		{
			name:     "new is invalid JSON",
			old:      `{"a":1}`,
			new:      "not json",
			expected: false,
		},
		{
			name:     "both invalid JSON",
			old:      "not json",
			new:      "also not json",
			expected: false,
		},
		{
			name:     "truncated JSON",
			old:      `{"a":1`,
			new:      `{"a":1}`,
			expected: false,
		},
		// Edge cases
		{
			name:     "null values",
			old:      `{"a":null}`,
			new:      `{"a": null}`,
			expected: true,
		},
		{
			name:     "boolean values",
			old:      `{"flag":true}`,
			new:      `{"flag": true}`,
			expected: true,
		},
		{
			name:     "string values",
			old:      `{"name":"test"}`,
			new:      `{"name": "test"}`,
			expected: true,
		},
		{
			name:     "empty object",
			old:      `{}`,
			new:      `{ }`,
			expected: true,
		},
		{
			name:     "empty array",
			old:      `[]`,
			new:      `[ ]`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SuppressJSONDiff("test_key", tt.old, tt.new, nil)
			assert.Equal(t, tt.expected, result, "SuppressJSONDiff(%q, %q)", tt.old, tt.new)
		})
	}
}

// TestSuppressEquivalentIPDiff tests IP address comparison.
func TestSuppressEquivalentIPDiff(t *testing.T) {
	tests := []struct {
		name     string
		old      string
		new      string
		expected bool
	}{
		// Equivalent values (should suppress diff)
		{
			name:     "same IPv4",
			old:      "192.168.1.1",
			new:      "192.168.1.1",
			expected: true,
		},
		{
			name:     "same IPv6",
			old:      "2001:db8::1",
			new:      "2001:db8::1",
			expected: true,
		},
		{
			name:     "IPv6 compressed vs expanded",
			old:      "::1",
			new:      "0:0:0:0:0:0:0:1",
			expected: true,
		},
		{
			name:     "IPv6 various compression",
			old:      "2001:0db8:0000:0000:0000:0000:0000:0001",
			new:      "2001:db8::1",
			expected: true,
		},
		{
			name:     "both empty",
			old:      "",
			new:      "",
			expected: true,
		},
		// Different values (should not suppress diff)
		{
			name:     "different IPv4",
			old:      "192.168.1.1",
			new:      "192.168.1.2",
			expected: false,
		},
		{
			name:     "different IPv6",
			old:      "2001:db8::1",
			new:      "2001:db8::2",
			expected: false,
		},
		{
			name:     "IPv4 vs IPv4-mapped IPv6",
			old:      "192.168.1.1",
			new:      "::ffff:192.168.1.1",
			expected: true, // Go's net.IP.Equal() treats these as equivalent
		},
		{
			name:     "empty vs non-empty",
			old:      "",
			new:      "192.168.1.1",
			expected: false,
		},
		{
			name:     "non-empty vs empty",
			old:      "192.168.1.1",
			new:      "",
			expected: false,
		},
		// Invalid IP (falls back to string comparison)
		{
			name:     "old is invalid IP - same string",
			old:      "not-an-ip",
			new:      "not-an-ip",
			expected: true, // Falls back to string comparison
		},
		{
			name:     "old is invalid IP - different string",
			old:      "not-an-ip",
			new:      "192.168.1.1",
			expected: false,
		},
		{
			name:     "new is invalid IP",
			old:      "192.168.1.1",
			new:      "not-an-ip",
			expected: false,
		},
		{
			name:     "both invalid but same",
			old:      "invalid",
			new:      "invalid",
			expected: true,
		},
		{
			name:     "both invalid and different",
			old:      "invalid1",
			new:      "invalid2",
			expected: false,
		},
		// Edge cases
		{
			name:     "localhost IPv4",
			old:      "127.0.0.1",
			new:      "127.0.0.1",
			expected: true,
		},
		{
			name:     "localhost IPv6",
			old:      "::1",
			new:      "::1",
			expected: true,
		},
		{
			name:     "broadcast address",
			old:      "255.255.255.255",
			new:      "255.255.255.255",
			expected: true,
		},
		{
			name:     "zero address IPv4",
			old:      "0.0.0.0",
			new:      "0.0.0.0",
			expected: true,
		},
		{
			name:     "zero address IPv6",
			old:      "::",
			new:      "0:0:0:0:0:0:0:0",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SuppressEquivalentIPDiff("test_key", tt.old, tt.new, nil)
			assert.Equal(t, tt.expected, result, "SuppressEquivalentIPDiff(%q, %q)", tt.old, tt.new)
		})
	}
}

// TestSuppressCaseAndWhitespaceDiff tests combined case and whitespace normalization.
func TestSuppressCaseAndWhitespaceDiff(t *testing.T) {
	tests := []struct {
		name     string
		old      string
		new      string
		expected bool
	}{
		// Equivalent values (should suppress diff)
		{
			name:     "same value",
			old:      "value",
			new:      "value",
			expected: true,
		},
		{
			name:     "case difference only",
			old:      "VALUE",
			new:      "value",
			expected: true,
		},
		{
			name:     "whitespace difference only",
			old:      "  value  ",
			new:      "value",
			expected: true,
		},
		{
			name:     "both case and whitespace difference",
			old:      "  VALUE  ",
			new:      "value",
			expected: true,
		},
		{
			name:     "mixed case and various whitespace",
			old:      "MyHost",
			new:      "  myhost\n",
			expected: true,
		},
		{
			name:     "both empty",
			old:      "",
			new:      "",
			expected: true,
		},
		{
			name:     "whitespace only vs empty",
			old:      "   ",
			new:      "",
			expected: true,
		},
		// Different values (should not suppress diff)
		{
			name:     "different values",
			old:      "value1",
			new:      "value2",
			expected: false,
		},
		{
			name:     "internal whitespace matters",
			old:      "a b",
			new:      "a  b",
			expected: false,
		},
		// Edge cases
		{
			name:     "tabs and newlines with case",
			old:      "\tVALUE\n",
			new:      "value",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SuppressCaseAndWhitespaceDiff("test_key", tt.old, tt.new, nil)
			assert.Equal(t, tt.expected, result, "SuppressCaseAndWhitespaceDiff(%q, %q)", tt.old, tt.new)
		})
	}
}

// TestSuppressEquivalentCIDRDiff tests CIDR notation comparison.
func TestSuppressEquivalentCIDRDiff(t *testing.T) {
	tests := []struct {
		name     string
		old      string
		new      string
		expected bool
	}{
		// Equivalent values (should suppress diff)
		{
			name:     "same CIDR",
			old:      "192.168.1.0/24",
			new:      "192.168.1.0/24",
			expected: true,
		},
		{
			name:     "same network different host portion",
			old:      "192.168.1.5/24",
			new:      "192.168.1.100/24",
			expected: true, // Both normalize to 192.168.1.0/24
		},
		{
			name:     "IPv6 CIDR",
			old:      "2001:db8::/32",
			new:      "2001:db8::/32",
			expected: true,
		},
		{
			name:     "both empty",
			old:      "",
			new:      "",
			expected: true,
		},
		// Different values (should not suppress diff)
		{
			name:     "different networks",
			old:      "192.168.1.0/24",
			new:      "192.168.2.0/24",
			expected: false,
		},
		{
			name:     "different prefix lengths",
			old:      "192.168.1.0/24",
			new:      "192.168.1.0/16",
			expected: false,
		},
		{
			name:     "empty vs non-empty",
			old:      "",
			new:      "192.168.1.0/24",
			expected: false,
		},
		{
			name:     "non-empty vs empty",
			old:      "192.168.1.0/24",
			new:      "",
			expected: false,
		},
		// Invalid CIDR (falls back to string comparison)
		{
			name:     "old is invalid CIDR - same string",
			old:      "not-a-cidr",
			new:      "not-a-cidr",
			expected: true, // Falls back to string comparison
		},
		{
			name:     "old is invalid CIDR - different string",
			old:      "not-a-cidr",
			new:      "192.168.1.0/24",
			expected: false,
		},
		{
			name:     "new is invalid CIDR",
			old:      "192.168.1.0/24",
			new:      "not-a-cidr",
			expected: false,
		},
		{
			name:     "IP without prefix (invalid CIDR)",
			old:      "192.168.1.1",
			new:      "192.168.1.1",
			expected: true, // Falls back to string comparison
		},
		// Edge cases
		{
			name:     "smallest prefix /32",
			old:      "192.168.1.1/32",
			new:      "192.168.1.1/32",
			expected: true,
		},
		{
			name:     "largest prefix /0",
			old:      "0.0.0.0/0",
			new:      "0.0.0.0/0",
			expected: true,
		},
		{
			name:     "loopback CIDR",
			old:      "127.0.0.0/8",
			new:      "127.0.0.0/8",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SuppressEquivalentCIDRDiff("test_key", tt.old, tt.new, nil)
			assert.Equal(t, tt.expected, result, "SuppressEquivalentCIDRDiff(%q, %q)", tt.old, tt.new)
		})
	}
}

// TestSuppressBooleanStringDiff tests boolean string comparison.
func TestSuppressBooleanStringDiff(t *testing.T) {
	tests := []struct {
		name     string
		old      string
		new      string
		expected bool
	}{
		// Equivalent values (should suppress diff)
		{
			name:     "same true lowercase",
			old:      "true",
			new:      "true",
			expected: true,
		},
		{
			name:     "true case variation",
			old:      "TRUE",
			new:      "true",
			expected: true,
		},
		{
			name:     "True mixed case",
			old:      "True",
			new:      "true",
			expected: true,
		},
		{
			name:     "same false lowercase",
			old:      "false",
			new:      "false",
			expected: true,
		},
		{
			name:     "false case variation",
			old:      "FALSE",
			new:      "false",
			expected: true,
		},
		{
			name:     "False mixed case",
			old:      "False",
			new:      "false",
			expected: true,
		},
		{
			name:     "both empty",
			old:      "",
			new:      "",
			expected: true,
		},
		// Different values (should not suppress diff)
		{
			name:     "true vs false",
			old:      "true",
			new:      "false",
			expected: false,
		},
		{
			name:     "yes vs true (not equivalent)",
			old:      "yes",
			new:      "true",
			expected: false,
		},
		{
			name:     "no vs false (not equivalent)",
			old:      "no",
			new:      "false",
			expected: false,
		},
		{
			name:     "1 vs true (not equivalent)",
			old:      "1",
			new:      "true",
			expected: false,
		},
		{
			name:     "0 vs false (not equivalent)",
			old:      "0",
			new:      "false",
			expected: false,
		},
		{
			name:     "empty vs true",
			old:      "",
			new:      "true",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SuppressBooleanStringDiff("test_key", tt.old, tt.new, nil)
			assert.Equal(t, tt.expected, result, "SuppressBooleanStringDiff(%q, %q)", tt.old, tt.new)
		})
	}
}

// TestSuppressEmptyStringDiff tests empty string comparison.
func TestSuppressEmptyStringDiff(t *testing.T) {
	tests := []struct {
		name     string
		old      string
		new      string
		expected bool
	}{
		// Equivalent values (should suppress diff)
		{
			name:     "both empty",
			old:      "",
			new:      "",
			expected: true,
		},
		{
			name:     "same non-empty",
			old:      "value",
			new:      "value",
			expected: true,
		},
		// Different values (should not suppress diff)
		{
			name:     "empty vs non-empty",
			old:      "",
			new:      "value",
			expected: false,
		},
		{
			name:     "non-empty vs empty",
			old:      "value",
			new:      "",
			expected: false,
		},
		{
			name:     "different non-empty values",
			old:      "value1",
			new:      "value2",
			expected: false,
		},
		// Edge cases
		{
			name:     "whitespace is not empty",
			old:      " ",
			new:      "",
			expected: false,
		},
		{
			name:     "same whitespace",
			old:      " ",
			new:      " ",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SuppressEmptyStringDiff("test_key", tt.old, tt.new, nil)
			assert.Equal(t, tt.expected, result, "SuppressEmptyStringDiff(%q, %q)", tt.old, tt.new)
		})
	}
}

// TestDiffSuppressFuncs_NilResourceData tests that functions handle nil ResourceData gracefully.
// This is important because DiffSuppressFunc receives ResourceData which could be nil in some contexts.
func TestDiffSuppressFuncs_NilResourceData(t *testing.T) {
	// All functions should handle nil ResourceData without panicking
	t.Run("SuppressCaseDiff with nil ResourceData", func(t *testing.T) {
		assert.NotPanics(t, func() {
			SuppressCaseDiff("key", "old", "new", nil)
		})
	})

	t.Run("SuppressWhitespaceDiff with nil ResourceData", func(t *testing.T) {
		assert.NotPanics(t, func() {
			SuppressWhitespaceDiff("key", "old", "new", nil)
		})
	})

	t.Run("SuppressJSONDiff with nil ResourceData", func(t *testing.T) {
		assert.NotPanics(t, func() {
			SuppressJSONDiff("key", `{"a":1}`, `{"a":1}`, nil)
		})
	})

	t.Run("SuppressEquivalentIPDiff with nil ResourceData", func(t *testing.T) {
		assert.NotPanics(t, func() {
			SuppressEquivalentIPDiff("key", "192.168.1.1", "192.168.1.1", nil)
		})
	})

	t.Run("SuppressCaseAndWhitespaceDiff with nil ResourceData", func(t *testing.T) {
		assert.NotPanics(t, func() {
			SuppressCaseAndWhitespaceDiff("key", "old", "new", nil)
		})
	})

	t.Run("SuppressEquivalentCIDRDiff with nil ResourceData", func(t *testing.T) {
		assert.NotPanics(t, func() {
			SuppressEquivalentCIDRDiff("key", "192.168.1.0/24", "192.168.1.0/24", nil)
		})
	})

	t.Run("SuppressBooleanStringDiff with nil ResourceData", func(t *testing.T) {
		assert.NotPanics(t, func() {
			SuppressBooleanStringDiff("key", "true", "false", nil)
		})
	})

	t.Run("SuppressEmptyStringDiff with nil ResourceData", func(t *testing.T) {
		assert.NotPanics(t, func() {
			SuppressEmptyStringDiff("key", "", "", nil)
		})
	})
}
