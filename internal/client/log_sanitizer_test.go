package client

import "testing"

func TestSanitizeCommandForLog(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Password patterns
		{
			name:     "login password command",
			input:    "login password secret123",
			expected: "[REDACTED - contains sensitive data]",
		},
		{
			name:     "administrator password command",
			input:    "administrator password mypass",
			expected: "[REDACTED - contains sensitive data]",
		},
		{
			name:     "uppercase PASSWORD",
			input:    "LOGIN PASSWORD SECRET",
			expected: "[REDACTED - contains sensitive data]",
		},
		{
			name:     "mixed case Password",
			input:    "administrator Password test",
			expected: "[REDACTED - contains sensitive data]",
		},

		// Pre-shared-key patterns
		{
			name:     "ipsec pre-shared-key",
			input:    "ipsec sa policy 1 1 esp aes-cbc sha-hmac pre-shared-key text mykey",
			expected: "[REDACTED - contains sensitive data]",
		},
		{
			name:     "uppercase PRE-SHARED-KEY",
			input:    "ipsec PRE-SHARED-KEY text test",
			expected: "[REDACTED - contains sensitive data]",
		},

		// Secret patterns
		{
			name:     "bgp secret",
			input:    "bgp neighbor 192.168.1.1 secret mysecret",
			expected: "[REDACTED - contains sensitive data]",
		},
		{
			name:     "uppercase SECRET",
			input:    "bgp neighbor SECRET test",
			expected: "[REDACTED - contains sensitive data]",
		},

		// Community patterns (SNMP)
		{
			name:     "snmp community",
			input:    "snmp community public",
			expected: "[REDACTED - contains sensitive data]",
		},
		{
			name:     "uppercase COMMUNITY",
			input:    "snmp COMMUNITY private",
			expected: "[REDACTED - contains sensitive data]",
		},

		// Safe commands (should NOT be redacted)
		{
			name:     "ip address command",
			input:    "ip lan1 address 192.168.1.1/24",
			expected: "ip lan1 address 192.168.1.1/24",
		},
		{
			name:     "show config command",
			input:    "show config",
			expected: "show config",
		},
		{
			name:     "interface command",
			input:    "ip lan1 description Office Network",
			expected: "ip lan1 description Office Network",
		},
		{
			name:     "routing command",
			input:    "ip route default gateway pp 1",
			expected: "ip route default gateway pp 1",
		},

		// Edge cases
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "   ",
		},
		{
			name:     "password keyword only",
			input:    "password",
			expected: "[REDACTED - contains sensitive data]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeCommandForLog(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeCommandForLog(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLogCommand(t *testing.T) {
	// LogCommand uses log.Printf internally, so we just verify it doesn't panic
	// and properly calls SanitizeCommandForLog

	// These should not panic
	LogCommand("[DEBUG] Test", "show config")
	LogCommand("[DEBUG] Test", "login password secret")
	LogCommand("[DEBUG] Test", "")
}
