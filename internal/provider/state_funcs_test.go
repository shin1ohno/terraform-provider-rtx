// Package provider contains Terraform provider implementation for RTX routers.
package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNormalizeIPAddress tests IP address normalization to canonical form.
func TestNormalizeIPAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		// Valid IPv4 addresses
		{
			name:     "standard IPv4",
			input:    "192.168.1.1",
			expected: "192.168.1.1",
		},
		{
			name:     "localhost IPv4",
			input:    "127.0.0.1",
			expected: "127.0.0.1",
		},
		// Note: IPv4 with leading zeros (e.g., "192.168.001.001") is not valid
		// per RFC and net.ParseIP returns nil, so it's returned unchanged.
		{
			name:     "broadcast address",
			input:    "255.255.255.255",
			expected: "255.255.255.255",
		},
		{
			name:     "zero address IPv4",
			input:    "0.0.0.0",
			expected: "0.0.0.0",
		},
		// Valid IPv6 addresses
		{
			name:     "IPv6 compressed form",
			input:    "2001:db8::1",
			expected: "2001:db8::1",
		},
		{
			name:     "IPv6 expanded form normalizes to compressed",
			input:    "2001:0db8:0000:0000:0000:0000:0000:0001",
			expected: "2001:db8::1",
		},
		{
			name:     "IPv6 localhost",
			input:    "::1",
			expected: "::1",
		},
		{
			name:     "IPv6 zero address",
			input:    "::",
			expected: "::",
		},
		{
			name:     "IPv6 full form",
			input:    "0:0:0:0:0:0:0:1",
			expected: "::1",
		},
		// IPv4-mapped IPv6 addresses convert to IPv4
		{
			name:     "IPv4-mapped IPv6 converts to IPv4",
			input:    "::ffff:192.168.1.1",
			expected: "192.168.1.1",
		},
		{
			name:     "IPv4-mapped IPv6 localhost",
			input:    "::ffff:127.0.0.1",
			expected: "127.0.0.1",
		},
		// Empty and invalid inputs
		{
			name:     "empty string returns empty",
			input:    "",
			expected: "",
		},
		{
			name:     "invalid IP returns unchanged",
			input:    "not-an-ip",
			expected: "not-an-ip",
		},
		{
			name:     "hostname returns unchanged",
			input:    "example.com",
			expected: "example.com",
		},
		{
			name:     "IP with CIDR returns unchanged",
			input:    "192.168.1.0/24",
			expected: "192.168.1.0/24",
		},
		{
			name:     "partial IP returns unchanged",
			input:    "192.168",
			expected: "192.168",
		},
		{
			name:     "IP with port returns unchanged",
			input:    "192.168.1.1:8080",
			expected: "192.168.1.1:8080",
		},
		// Non-string inputs
		{
			name:     "integer input returns empty",
			input:    12345,
			expected: "",
		},
		{
			name:     "nil input returns empty",
			input:    nil,
			expected: "",
		},
		{
			name:     "boolean input returns empty",
			input:    true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeIPAddress(tt.input)
			assert.Equal(t, tt.expected, result, "normalizeIPAddress(%v)", tt.input)
		})
	}
}

// TestNormalizeIPAddress_Idempotency verifies that normalizing twice produces the same result.
func TestNormalizeIPAddress_Idempotency(t *testing.T) {
	tests := []string{
		"192.168.1.1",
		"2001:db8::1",
		"2001:0db8:0000:0000:0000:0000:0000:0001",
		"::ffff:192.168.1.1",
		"::1",
		"not-an-ip",
		"",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			first := normalizeIPAddress(input)
			second := normalizeIPAddress(first)
			assert.Equal(t, first, second, "normalizeIPAddress should be idempotent for %q", input)
		})
	}
}

// TestNormalizeLowercase tests lowercase string normalization.
func TestNormalizeLowercase(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		// Basic transformations
		{
			name:     "all uppercase to lowercase",
			input:    "HELLO",
			expected: "hello",
		},
		{
			name:     "mixed case to lowercase",
			input:    "HelloWorld",
			expected: "helloworld",
		},
		{
			name:     "already lowercase unchanged",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "with numbers",
			input:    "User123",
			expected: "user123",
		},
		{
			name:     "with special characters",
			input:    "Test@Example.Com",
			expected: "test@example.com",
		},
		// Whitespace preservation
		{
			name:     "leading whitespace preserved",
			input:    "  HELLO",
			expected: "  hello",
		},
		{
			name:     "trailing whitespace preserved",
			input:    "HELLO  ",
			expected: "hello  ",
		},
		{
			name:     "internal whitespace preserved",
			input:    "HELLO WORLD",
			expected: "hello world",
		},
		{
			name:     "tabs and newlines preserved",
			input:    "\tHELLO\n",
			expected: "\thello\n",
		},
		// Empty and edge cases
		{
			name:     "empty string unchanged",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only unchanged",
			input:    "   ",
			expected: "   ",
		},
		{
			name:     "unicode characters",
			input:    "HELLO",
			expected: "hello",
		},
		// Non-string inputs
		{
			name:     "integer input returns empty",
			input:    12345,
			expected: "",
		},
		{
			name:     "nil input returns empty",
			input:    nil,
			expected: "",
		},
		{
			name:     "boolean input returns empty",
			input:    true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeLowercase(tt.input)
			assert.Equal(t, tt.expected, result, "normalizeLowercase(%v)", tt.input)
		})
	}
}

// TestNormalizeLowercase_Idempotency verifies that normalizing twice produces the same result.
func TestNormalizeLowercase_Idempotency(t *testing.T) {
	tests := []string{
		"HELLO",
		"HelloWorld",
		"hello",
		"  HELLO  ",
		"Test123",
		"",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			first := normalizeLowercase(input)
			second := normalizeLowercase(first)
			assert.Equal(t, first, second, "normalizeLowercase should be idempotent for %q", input)
		})
	}
}

// TestNormalizeJSON tests JSON normalization to consistent formatting.
func TestNormalizeJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		// Key ordering normalization
		{
			name:     "keys sorted alphabetically",
			input:    `{"b":2,"a":1}`,
			expected: `{"a":1,"b":2}`,
		},
		{
			name:     "already sorted unchanged",
			input:    `{"a":1,"b":2}`,
			expected: `{"a":1,"b":2}`,
		},
		// Whitespace normalization
		{
			name:     "remove whitespace around colons",
			input:    `{ "key" : "value" }`,
			expected: `{"key":"value"}`,
		},
		{
			name:     "remove whitespace in arrays",
			input:    `[1, 2, 3]`,
			expected: `[1,2,3]`,
		},
		{
			name:     "pretty printed to compact",
			input:    "{\n  \"a\": 1,\n  \"b\": 2\n}",
			expected: `{"a":1,"b":2}`,
		},
		// Nested structures
		{
			name:     "nested objects",
			input:    `{"outer": {"inner": 1}}`,
			expected: `{"outer":{"inner":1}}`,
		},
		{
			name:     "nested arrays",
			input:    `{"arr": [1, 2, 3]}`,
			expected: `{"arr":[1,2,3]}`,
		},
		{
			name:     "deeply nested",
			input:    `{"a": {"b": {"c": 1}}}`,
			expected: `{"a":{"b":{"c":1}}}`,
		},
		// Different value types
		{
			name:     "null value",
			input:    `{"a": null}`,
			expected: `{"a":null}`,
		},
		{
			name:     "boolean values",
			input:    `{"flag": true, "other": false}`,
			expected: `{"flag":true,"other":false}`,
		},
		{
			name:     "number values",
			input:    `{"int": 42, "float": 3.14}`,
			expected: `{"float":3.14,"int":42}`,
		},
		{
			name:     "string with special chars",
			input:    `{"msg": "hello\nworld"}`,
			expected: `{"msg":"hello\nworld"}`,
		},
		// Empty structures
		{
			name:     "empty object",
			input:    `{ }`,
			expected: `{}`,
		},
		{
			name:     "empty array",
			input:    `[ ]`,
			expected: `[]`,
		},
		{
			name:     "empty string unchanged",
			input:    "",
			expected: "",
		},
		// Invalid JSON returns unchanged
		{
			name:     "invalid JSON returns unchanged",
			input:    "not json",
			expected: "not json",
		},
		{
			name:     "truncated JSON returns unchanged",
			input:    `{"a":1`,
			expected: `{"a":1`,
		},
		{
			name:     "trailing comma returns unchanged",
			input:    `{"a":1,}`,
			expected: `{"a":1,}`,
		},
		// Non-string inputs
		{
			name:     "integer input returns empty",
			input:    12345,
			expected: "",
		},
		{
			name:     "nil input returns empty",
			input:    nil,
			expected: "",
		},
		{
			name:     "boolean input returns empty",
			input:    true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeJSON(tt.input)
			assert.Equal(t, tt.expected, result, "normalizeJSON(%v)", tt.input)
		})
	}
}

// TestNormalizeJSON_Idempotency verifies that normalizing twice produces the same result.
func TestNormalizeJSON_Idempotency(t *testing.T) {
	tests := []string{
		`{"b":2,"a":1}`,
		`{ "key" : "value" }`,
		`[1, 2, 3]`,
		`{"nested": {"inner": 1}}`,
		`{}`,
		`[]`,
		"not json",
		"",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			first := normalizeJSON(input)
			second := normalizeJSON(first)
			assert.Equal(t, first, second, "normalizeJSON should be idempotent for %q", input)
		})
	}
}

// TestNormalizeTrimmedString tests whitespace trimming normalization.
func TestNormalizeTrimmedString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		// Leading whitespace removal
		{
			name:     "leading spaces removed",
			input:    "  hello",
			expected: "hello",
		},
		{
			name:     "leading tabs removed",
			input:    "\thello",
			expected: "hello",
		},
		{
			name:     "leading newlines removed",
			input:    "\nhello",
			expected: "hello",
		},
		// Trailing whitespace removal
		{
			name:     "trailing spaces removed",
			input:    "hello  ",
			expected: "hello",
		},
		{
			name:     "trailing tabs removed",
			input:    "hello\t",
			expected: "hello",
		},
		{
			name:     "trailing newlines removed",
			input:    "hello\n",
			expected: "hello",
		},
		// Both leading and trailing
		{
			name:     "both leading and trailing removed",
			input:    "  hello  ",
			expected: "hello",
		},
		{
			name:     "mixed whitespace characters",
			input:    "\t\n\r hello \t\n\r",
			expected: "hello",
		},
		// Internal whitespace preserved
		{
			name:     "internal spaces preserved",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "multiple internal spaces preserved",
			input:    "hello  world",
			expected: "hello  world",
		},
		{
			name:     "internal tabs preserved",
			input:    "hello\tworld",
			expected: "hello\tworld",
		},
		// Edge cases
		{
			name:     "empty string unchanged",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only becomes empty",
			input:    "   ",
			expected: "",
		},
		{
			name:     "no trimming needed",
			input:    "hello",
			expected: "hello",
		},
		// Non-string inputs
		{
			name:     "integer input returns empty",
			input:    12345,
			expected: "",
		},
		{
			name:     "nil input returns empty",
			input:    nil,
			expected: "",
		},
		{
			name:     "boolean input returns empty",
			input:    true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeTrimmedString(tt.input)
			assert.Equal(t, tt.expected, result, "normalizeTrimmedString(%v)", tt.input)
		})
	}
}

// TestNormalizeTrimmedString_Idempotency verifies that normalizing twice produces the same result.
func TestNormalizeTrimmedString_Idempotency(t *testing.T) {
	tests := []string{
		"  hello  ",
		"\t\nhello\r\n",
		"hello world",
		"hello",
		"   ",
		"",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			first := normalizeTrimmedString(input)
			second := normalizeTrimmedString(first)
			assert.Equal(t, first, second, "normalizeTrimmedString should be idempotent for %q", input)
		})
	}
}

// TestNormalizeUppercase tests uppercase string normalization.
func TestNormalizeUppercase(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		// Basic transformations
		{
			name:     "all lowercase to uppercase",
			input:    "hello",
			expected: "HELLO",
		},
		{
			name:     "mixed case to uppercase",
			input:    "HelloWorld",
			expected: "HELLOWORLD",
		},
		{
			name:     "already uppercase unchanged",
			input:    "HELLO",
			expected: "HELLO",
		},
		{
			name:     "with numbers",
			input:    "user123",
			expected: "USER123",
		},
		{
			name:     "MAC address lowercase",
			input:    "aa:bb:cc:dd:ee:ff",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "MAC address mixed case",
			input:    "Aa:Bb:Cc:Dd:Ee:Ff",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		// Whitespace preservation
		{
			name:     "leading whitespace preserved",
			input:    "  hello",
			expected: "  HELLO",
		},
		{
			name:     "trailing whitespace preserved",
			input:    "hello  ",
			expected: "HELLO  ",
		},
		{
			name:     "internal whitespace preserved",
			input:    "hello world",
			expected: "HELLO WORLD",
		},
		{
			name:     "tabs and newlines preserved",
			input:    "\thello\n",
			expected: "\tHELLO\n",
		},
		// Empty and edge cases
		{
			name:     "empty string unchanged",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only unchanged",
			input:    "   ",
			expected: "   ",
		},
		// Non-string inputs
		{
			name:     "integer input returns empty",
			input:    12345,
			expected: "",
		},
		{
			name:     "nil input returns empty",
			input:    nil,
			expected: "",
		},
		{
			name:     "boolean input returns empty",
			input:    true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeUppercase(tt.input)
			assert.Equal(t, tt.expected, result, "normalizeUppercase(%v)", tt.input)
		})
	}
}

// TestNormalizeUppercase_Idempotency verifies that normalizing twice produces the same result.
func TestNormalizeUppercase_Idempotency(t *testing.T) {
	tests := []string{
		"hello",
		"HelloWorld",
		"HELLO",
		"  hello  ",
		"aa:bb:cc:dd:ee:ff",
		"",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			first := normalizeUppercase(input)
			second := normalizeUppercase(first)
			assert.Equal(t, first, second, "normalizeUppercase should be idempotent for %q", input)
		})
	}
}

// TestStateFuncs_NoPanic verifies that all normalizers handle any input without panicking.
func TestStateFuncs_NoPanic(t *testing.T) {
	// Various problematic inputs that should not cause panic
	inputs := []interface{}{
		nil,
		"",
		"normal string",
		123,
		3.14,
		true,
		false,
		[]string{"a", "b"},
		map[string]int{"a": 1},
	}

	for _, input := range inputs {
		t.Run("normalizeIPAddress", func(t *testing.T) {
			assert.NotPanics(t, func() {
				normalizeIPAddress(input)
			})
		})

		t.Run("normalizeLowercase", func(t *testing.T) {
			assert.NotPanics(t, func() {
				normalizeLowercase(input)
			})
		})

		t.Run("normalizeJSON", func(t *testing.T) {
			assert.NotPanics(t, func() {
				normalizeJSON(input)
			})
		})

		t.Run("normalizeTrimmedString", func(t *testing.T) {
			assert.NotPanics(t, func() {
				normalizeTrimmedString(input)
			})
		})

		t.Run("normalizeUppercase", func(t *testing.T) {
			assert.NotPanics(t, func() {
				normalizeUppercase(input)
			})
		})
	}
}
