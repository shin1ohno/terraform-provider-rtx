package parsers

import (
	"fmt"
	"strings"
	"testing"
)

func TestBuildDHCPScopeCreateCommand(t *testing.T) {
	tests := []struct {
		name     string
		scope    DhcpScope
		expected string
	}{
		{
			name: "Basic scope with minimum fields",
			scope: DhcpScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
			},
			expected: "dhcp scope 1 192.168.1.100-192.168.1.200/24",
		},
		{
			name: "Scope with gateway",
			scope: DhcpScope{
				ID:         2,
				RangeStart: "10.0.0.10",
				RangeEnd:   "10.0.0.100",
				Prefix:     24,
				Gateway:    "10.0.0.1",
			},
			expected: "dhcp scope 2 10.0.0.10-10.0.0.100/24 gateway 10.0.0.1",
		},
		{
			name: "Scope with single DNS server",
			scope: DhcpScope{
				ID:         3,
				RangeStart: "172.16.0.50",
				RangeEnd:   "172.16.0.150",
				Prefix:     16,
				DNSServers: []string{"8.8.8.8"},
			},
			expected: "dhcp scope 3 172.16.0.50-172.16.0.150/16 dns 8.8.8.8",
		},
		{
			name: "Scope with multiple DNS servers",
			scope: DhcpScope{
				ID:         4,
				RangeStart: "192.168.10.20",
				RangeEnd:   "192.168.10.80",
				Prefix:     24,
				DNSServers: []string{"8.8.8.8", "8.8.4.4", "1.1.1.1"},
			},
			expected: "dhcp scope 4 192.168.10.20-192.168.10.80/24 dns 8.8.8.8 8.8.4.4 1.1.1.1",
		},
		{
			name: "Scope with lease time",
			scope: DhcpScope{
				ID:         5,
				RangeStart: "10.10.10.10",
				RangeEnd:   "10.10.10.50",
				Prefix:     28,
				Lease:      3600,
			},
			expected: "dhcp scope 5 10.10.10.10-10.10.10.50/28 lease 3600",
		},
		{
			name: "Scope with domain name",
			scope: DhcpScope{
				ID:         6,
				RangeStart: "192.168.0.10",
				RangeEnd:   "192.168.0.50",
				Prefix:     24,
				DomainName: "example.com",
			},
			expected: "dhcp scope 6 192.168.0.10-192.168.0.50/24 domain example.com",
		},
		{
			name: "Scope with all options",
			scope: DhcpScope{
				ID:         10,
				RangeStart: "192.168.100.10",
				RangeEnd:   "192.168.100.100",
				Prefix:     24,
				Gateway:    "192.168.100.1",
				DNSServers: []string{"192.168.100.1", "8.8.8.8"},
				Lease:      7200,
				DomainName: "internal.local",
			},
			expected: "dhcp scope 10 192.168.100.10-192.168.100.100/24 gateway 192.168.100.1 dns 192.168.100.1 8.8.8.8 lease 7200 domain internal.local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDHCPScopeCreateCommand(tt.scope)
			if result != tt.expected {
				t.Errorf("BuildDHCPScopeCreateCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDHCPScopeDeleteCommand(t *testing.T) {
	tests := []struct {
		name     string
		scopeID  int
		expected string
	}{
		{
			name:     "Delete scope 1",
			scopeID:  1,
			expected: "no dhcp scope 1",
		},
		{
			name:     "Delete scope 255",
			scopeID:  255,
			expected: "no dhcp scope 255",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDHCPScopeDeleteCommand(tt.scopeID)
			if result != tt.expected {
				t.Errorf("BuildDHCPScopeDeleteCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildShowDHCPScopesCommand(t *testing.T) {
	expected := `show config | grep "dhcp scope"`
	result := BuildShowDHCPScopesCommand()
	if result != expected {
		t.Errorf("BuildShowDHCPScopesCommand() = %q, want %q", result, expected)
	}
}

func TestBuildDHCPScopeCreateCommandWithValidation(t *testing.T) {
	tests := []struct {
		name        string
		scope       DhcpScope
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid basic scope",
			scope: DhcpScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
			},
			expectError: false,
		},
		{
			name: "Invalid scope ID - zero",
			scope: DhcpScope{
				ID:         0,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
			},
			expectError: true,
			errorMsg:    "scope ID must be between 1 and 255",
		},
		{
			name: "Invalid scope ID - too high",
			scope: DhcpScope{
				ID:         256,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
			},
			expectError: true,
			errorMsg:    "scope ID must be between 1 and 255",
		},
		{
			name: "Missing range_start",
			scope: DhcpScope{
				ID:         1,
				RangeStart: "",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
			},
			expectError: true,
			errorMsg:    "range_start is required",
		},
		{
			name: "Missing range_end",
			scope: DhcpScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "",
				Prefix:     24,
			},
			expectError: true,
			errorMsg:    "range_end is required",
		},
		{
			name: "Invalid range_start IP",
			scope: DhcpScope{
				ID:         1,
				RangeStart: "invalid.ip",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
			},
			expectError: true,
			errorMsg:    "invalid range_start IP address: invalid.ip",
		},
		{
			name: "Invalid range_end IP",
			scope: DhcpScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "invalid.ip",
				Prefix:     24,
			},
			expectError: true,
			errorMsg:    "invalid range_end IP address: invalid.ip",
		},
		{
			name: "Invalid prefix - too low",
			scope: DhcpScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     7,
			},
			expectError: true,
			errorMsg:    "prefix must be between 8 and 32",
		},
		{
			name: "Invalid prefix - too high",
			scope: DhcpScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     33,
			},
			expectError: true,
			errorMsg:    "prefix must be between 8 and 32",
		},
		{
			name: "Invalid gateway IP",
			scope: DhcpScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
				Gateway:    "invalid.gateway",
			},
			expectError: true,
			errorMsg:    "invalid gateway IP address: invalid.gateway",
		},
		{
			name: "Invalid DNS server IP",
			scope: DhcpScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
				DNSServers: []string{"8.8.8.8", "invalid.dns"},
			},
			expectError: true,
			errorMsg:    "invalid DNS server IP address at index 1: invalid.dns",
		},
		{
			name: "Invalid lease - negative",
			scope: DhcpScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
				Lease:      -1,
			},
			expectError: true,
			errorMsg:    "lease must be non-negative",
		},
		{
			name: "Invalid range - start >= end",
			scope: DhcpScope{
				ID:         1,
				RangeStart: "192.168.1.200",
				RangeEnd:   "192.168.1.100",
				Prefix:     24,
			},
			expectError: true,
			errorMsg:    "range_start must be less than range_end",
		},
		{
			name: "Invalid range - same IP",
			scope: DhcpScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.100",
				Prefix:     24,
			},
			expectError: true,
			errorMsg:    "range_start must be less than range_end",
		},
		{
			name: "Valid scope with all options",
			scope: DhcpScope{
				ID:         10,
				RangeStart: "192.168.100.10",
				RangeEnd:   "192.168.100.100",
				Prefix:     24,
				Gateway:    "192.168.100.1",
				DNSServers: []string{"192.168.100.1", "8.8.8.8"},
				Lease:      7200,
				DomainName: "internal.local",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildDHCPScopeCreateCommandWithValidation(tt.scope)

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

func TestIpToUint32(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected uint32
	}{
		{
			name:     "192.168.1.1",
			ip:       "192.168.1.1",
			expected: 0xc0a80101, // 192*256^3 + 168*256^2 + 1*256 + 1
		},
		{
			name:     "10.0.0.1",
			ip:       "10.0.0.1",
			expected: 0x0a000001, // 10*256^3 + 0*256^2 + 0*256 + 1
		},
		{
			name:     "0.0.0.0",
			ip:       "0.0.0.0",
			expected: 0x00000000,
		},
		{
			name:     "255.255.255.255",
			ip:       "255.255.255.255",
			expected: 0xffffffff,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := ParseIPHelper(tt.ip)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", tt.ip)
			}
			result := ipToUint32(ip)
			if result != tt.expected {
				t.Errorf("ipToUint32(%s) = 0x%x, want 0x%x", tt.ip, result, tt.expected)
			}
		})
	}
}

// ParseIPHelper is a helper function for tests
func ParseIPHelper(ipStr string) []byte {
	ip := make([]byte, 4)
	// Simple parsing for test purposes
	switch ipStr {
	case "192.168.1.1":
		ip[0], ip[1], ip[2], ip[3] = 192, 168, 1, 1
	case "10.0.0.1":
		ip[0], ip[1], ip[2], ip[3] = 10, 0, 0, 1
	case "0.0.0.0":
		ip[0], ip[1], ip[2], ip[3] = 0, 0, 0, 0
	case "255.255.255.255":
		ip[0], ip[1], ip[2], ip[3] = 255, 255, 255, 255
	default:
		return nil
	}
	return ip
}

func TestBuildDHCPScopeUpdateCommand(t *testing.T) {
	tests := []struct {
		name        string
		scope       DhcpScope
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid scope update",
			scope: DhcpScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
				Gateway:    "192.168.1.1",
			},
			expectError: false,
		},
		{
			name: "Invalid scope update - bad scope ID",
			scope: DhcpScope{
				ID:         0,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
			},
			expectError: true,
			errorMsg:    "scope ID must be between 1 and 255",
		},
		{
			name: "Invalid scope update - invalid range",
			scope: DhcpScope{
				ID:         1,
				RangeStart: "192.168.1.200",
				RangeEnd:   "192.168.1.100",
				Prefix:     24,
			},
			expectError: true,
			errorMsg:    "range_start must be less than range_end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildDHCPScopeUpdateCommand(tt.scope)

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
				if len(result) != 2 {
					t.Errorf("Expected 2 commands, got %d", len(result))
				} else {
					expectedDelete := "no dhcp scope " + fmt.Sprintf("%d", tt.scope.ID)
					if result[0] != expectedDelete {
						t.Errorf("Expected delete command '%s', got '%s'", expectedDelete, result[0])
					}
					// Create command should start with "dhcp scope"
					if !strings.HasPrefix(result[1], "dhcp scope") {
						t.Errorf("Expected create command to start with 'dhcp scope', got '%s'", result[1])
					}
				}
			}
		})
	}
}