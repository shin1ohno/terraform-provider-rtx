package parsers

import (
	"testing"
)

func TestBuildDHCPBindCommand(t *testing.T) {
	tests := []struct {
		name     string
		binding  DHCPBinding
		expected string
	}{
		{
			name: "Basic MAC binding",
			binding: DHCPBinding{
				ScopeID:             1,
				IPAddress:           "192.168.1.100",
				MACAddress:          "00:11:22:33:44:55",
				UseClientIdentifier: false,
			},
			expected: "dhcp scope bind 1 192.168.1.100 00:11:22:33:44:55",
		},
		{
			name: "Client identifier binding",
			binding: DHCPBinding{
				ScopeID:             1,
				IPAddress:           "192.168.1.101",
				MACAddress:          "00:aa:bb:cc:dd:ee",
				UseClientIdentifier: true,
			},
			expected: "dhcp scope bind 1 192.168.1.101 ethernet 00:aa:bb:cc:dd:ee",
		},
		{
			name: "Different scope",
			binding: DHCPBinding{
				ScopeID:             2,
				IPAddress:           "10.0.0.50",
				MACAddress:          "11:22:33:44:55:66",
				UseClientIdentifier: false,
			},
			expected: "dhcp scope bind 2 10.0.0.50 11:22:33:44:55:66",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDHCPBindCommand(tt.binding)
			if result != tt.expected {
				t.Errorf("BuildDHCPBindCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDHCPUnbindCommand(t *testing.T) {
	tests := []struct {
		name      string
		scopeID   int
		ipAddress string
		expected  string
	}{
		{
			name:      "Basic unbind",
			scopeID:   1,
			ipAddress: "192.168.1.100",
			expected:  "no dhcp scope bind 1 192.168.1.100",
		},
		{
			name:      "Different scope",
			scopeID:   2,
			ipAddress: "10.0.0.50",
			expected:  "no dhcp scope bind 2 10.0.0.50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDHCPUnbindCommand(tt.scopeID, tt.ipAddress)
			if result != tt.expected {
				t.Errorf("BuildDHCPUnbindCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildShowDHCPBindingsCommand(t *testing.T) {
	tests := []struct {
		name     string
		scopeID  int
		expected string
	}{
		{
			name:     "Scope 1",
			scopeID:  1,
			expected: "show dhcp scope bind 1",
		},
		{
			name:     "Scope 2",
			scopeID:  2,
			expected: "show dhcp scope bind 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildShowDHCPBindingsCommand(tt.scopeID)
			if result != tt.expected {
				t.Errorf("BuildShowDHCPBindingsCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}