package logging

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
)

func TestContainsSensitivePattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "password in string",
			input:    "login password mypass",
			expected: true,
		},
		{
			name:     "PASSWORD uppercase",
			input:    "LOGIN PASSWORD MYPASS",
			expected: true,
		},
		{
			name:     "pre-shared-key",
			input:    "ipsec pre-shared-key text secret123",
			expected: true,
		},
		{
			name:     "secret in string",
			input:    "bgp neighbor secret key123",
			expected: true,
		},
		{
			name:     "community string",
			input:    "snmp community public",
			expected: true,
		},
		{
			name:     "token in string",
			input:    "auth token abc123",
			expected: true,
		},
		{
			name:     "key in string",
			input:    "api key xyz",
			expected: true,
		},
		{
			name:     "no sensitive pattern",
			input:    "ip route 192.168.1.0/24",
			expected: false,
		},
		{
			name:     "interface name",
			input:    "interface lan1",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsSensitivePattern(tt.input)
			if result != tt.expected {
				t.Errorf("containsSensitivePattern(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "password redacted",
			input:    "login password mypass",
			expected: redactedMessage,
		},
		{
			name:     "pre-shared-key redacted",
			input:    "ipsec pre-shared-key text secret123",
			expected: redactedMessage,
		},
		{
			name:     "safe command unchanged",
			input:    "ip route 192.168.1.0/24",
			expected: "ip route 192.168.1.0/24",
		},
		{
			name:     "community redacted",
			input:    "snmp community public",
			expected: redactedMessage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsSensitiveField(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		expected  bool
	}{
		{name: "password is sensitive", fieldName: "password", expected: true},
		{name: "PASSWORD uppercase", fieldName: "PASSWORD", expected: true},
		{name: "pre_shared_key is sensitive", fieldName: "pre_shared_key", expected: true},
		{name: "secret is sensitive", fieldName: "secret", expected: true},
		{name: "community is sensitive", fieldName: "community", expected: true},
		{name: "token is sensitive", fieldName: "token", expected: true},
		{name: "api_key is sensitive", fieldName: "api_key", expected: true},
		{name: "credential is sensitive", fieldName: "credential", expected: true},
		{name: "command is not sensitive", fieldName: "command", expected: false},
		{name: "host is not sensitive", fieldName: "host", expected: false},
		{name: "resource_id is not sensitive", fieldName: "resource_id", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSensitiveField(tt.fieldName)
			if result != tt.expected {
				t.Errorf("IsSensitiveField(%q) = %v, want %v", tt.fieldName, result, tt.expected)
			}
		})
	}
}

func TestSanitizeMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "password field redacted",
			input: map[string]interface{}{
				"user":     "admin",
				"password": "secret123",
			},
			expected: map[string]interface{}{
				"user":     "admin",
				"password": redactedMessage,
			},
		},
		{
			name: "sensitive value in non-sensitive field redacted",
			input: map[string]interface{}{
				"command": "login password mypass",
				"host":    "192.168.1.1",
			},
			expected: map[string]interface{}{
				"command": redactedMessage,
				"host":    "192.168.1.1",
			},
		},
		{
			name: "safe map unchanged",
			input: map[string]interface{}{
				"host":    "192.168.1.1",
				"port":    22,
				"command": "show config",
			},
			expected: map[string]interface{}{
				"host":    "192.168.1.1",
				"port":    22,
				"command": "show config",
			},
		},
		{
			name: "multiple sensitive fields",
			input: map[string]interface{}{
				"api_key":  "abc123",
				"token":    "xyz789",
				"resource": "nat_masquerade",
			},
			expected: map[string]interface{}{
				"api_key":  redactedMessage,
				"token":    redactedMessage,
				"resource": "nat_masquerade",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeMap(tt.input)
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("SanitizeMap()[%q] = %v, want %v", k, result[k], v)
				}
			}
		})
	}
}

func TestSanitizingHook(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Hook(SanitizingHook{})

	// Log a message with sensitive content
	logger.Info().Str("command", "test").Msg("login password mypass")

	output := buf.String()
	// The hook should have added the sanitized field
	if !bytes.Contains(buf.Bytes(), []byte("sanitized")) {
		t.Errorf("expected 'sanitized' field in output for sensitive message, got %q", output)
	}
}

func TestSanitizingHook_SafeMessage(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Hook(SanitizingHook{})

	// Log a message without sensitive content
	logger.Info().Str("command", "show config").Msg("executing command")

	output := buf.String()
	// The hook should NOT have added the sanitized field
	if bytes.Contains(buf.Bytes(), []byte("sanitized")) {
		t.Errorf("unexpected 'sanitized' field in output for safe message, got %q", output)
	}
}

func TestNewSanitizedLogger(t *testing.T) {
	logger := NewSanitizedLogger()
	// Just verify it doesn't panic and returns a valid logger
	logger.Info().Msg("test")
}
