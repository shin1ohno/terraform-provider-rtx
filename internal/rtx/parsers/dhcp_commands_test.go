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
		// New client identifier test cases
		{
			name: "MAC-based client identifier (01 prefix)",
			binding: DHCPBinding{
				ScopeID:          1,
				IPAddress:        "192.168.1.102",
				ClientIdentifier: "01:00:11:22:33:44:55",
			},
			expected: "dhcp scope bind 1 192.168.1.102 client-id 01:00:11:22:33:44:55",
		},
		{
			name: "ASCII-based client identifier (02 prefix)",
			binding: DHCPBinding{
				ScopeID:          1,
				IPAddress:        "192.168.1.103",
				ClientIdentifier: "02:68:6f:73:74:6e:61:6d:65", // "hostname" in hex
			},
			expected: "dhcp scope bind 1 192.168.1.103 client-id 02:68:6f:73:74:6e:61:6d:65",
		},
		{
			name: "Vendor-specific client identifier (FF prefix)",
			binding: DHCPBinding{
				ScopeID:          2,
				IPAddress:        "10.0.0.51",
				ClientIdentifier: "ff:00:01:02:03:04:05",
			},
			expected: "dhcp scope bind 2 10.0.0.51 client-id ff:00:01:02:03:04:05",
		},
		{
			name: "Custom client identifier with mixed case",
			binding: DHCPBinding{
				ScopeID:          1,
				IPAddress:        "192.168.1.104",
				ClientIdentifier: "01:AA:BB:CC:DD:EE:FF",
			},
			expected: "dhcp scope bind 1 192.168.1.104 client-id 01:aa:bb:cc:dd:ee:ff",
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

func TestBuildDHCPBindCommandClientIdentifierValidation(t *testing.T) {
	tests := []struct {
		name        string
		binding     DHCPBinding
		expectError bool
		errorMsg    string
	}{
		{
			name: "Invalid client identifier format - no colon",
			binding: DHCPBinding{
				ScopeID:          1,
				IPAddress:        "192.168.1.100",
				ClientIdentifier: "0100112233445566",
			},
			expectError: true,
			errorMsg:    "invalid client identifier format",
		},
		{
			name: "Invalid client identifier format - invalid hex",
			binding: DHCPBinding{
				ScopeID:          1,
				IPAddress:        "192.168.1.100",
				ClientIdentifier: "01:zz:11:22:33:44:55",
			},
			expectError: true,
			errorMsg:    "invalid hex characters in client identifier",
		},
		{
			name: "Invalid client identifier format - unsupported prefix",
			binding: DHCPBinding{
				ScopeID:          1,
				IPAddress:        "192.168.1.100",
				ClientIdentifier: "03:00:11:22:33:44:55",
			},
			expectError: true,
			errorMsg:    "unsupported client identifier prefix",
		},
		{
			name: "Empty client identifier",
			binding: DHCPBinding{
				ScopeID:          1,
				IPAddress:        "192.168.1.100",
				ClientIdentifier: "",
			},
			expectError: true,
			errorMsg:    "client identifier cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildDHCPBindCommandWithValidation(tt.binding)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result == "" {
					t.Errorf("Expected command result but got empty string")
				}
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
			expected: `show config | grep "dhcp scope bind 1"`,
		},
		{
			name:     "Scope 2",
			scopeID:  2,
			expected: `show config | grep "dhcp scope bind 2"`,
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
