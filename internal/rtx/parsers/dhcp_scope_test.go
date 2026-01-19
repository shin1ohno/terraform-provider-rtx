package parsers

import (
	"testing"
)

func TestParseScopeConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []DHCPScope
		wantErr  bool
	}{
		{
			name:  "basic scope",
			input: `dhcp scope 1 192.168.1.0/24`,
			expected: []DHCPScope{
				{
					ScopeID:       1,
					Network:       "192.168.1.0/24",
					ExcludeRanges: []ExcludeRange{},
				},
			},
		},
		{
			name:  "scope with gateway (legacy format)",
			input: `dhcp scope 1 192.168.1.0/24 gateway 192.168.1.1`,
			expected: []DHCPScope{
				{
					ScopeID:       1,
					Network:       "192.168.1.0/24",
					Options:       DHCPScopeOptions{Routers: []string{"192.168.1.1"}},
					ExcludeRanges: []ExcludeRange{},
				},
			},
		},
		{
			name:  "scope with gateway and expire",
			input: `dhcp scope 1 192.168.1.0/24 gateway 192.168.1.1 expire 72:00`,
			expected: []DHCPScope{
				{
					ScopeID:       1,
					Network:       "192.168.1.0/24",
					Options:       DHCPScopeOptions{Routers: []string{"192.168.1.1"}},
					LeaseTime:     "72h",
					ExcludeRanges: []ExcludeRange{},
				},
			},
		},
		{
			name: "scope with DNS option",
			input: `dhcp scope 1 192.168.1.0/24
dhcp scope option 1 dns=8.8.8.8,8.8.4.4`,
			expected: []DHCPScope{
				{
					ScopeID:       1,
					Network:       "192.168.1.0/24",
					Options:       DHCPScopeOptions{DNSServers: []string{"8.8.8.8", "8.8.4.4"}},
					ExcludeRanges: []ExcludeRange{},
				},
			},
		},
		{
			name: "scope with router option",
			input: `dhcp scope 1 192.168.1.0/24
dhcp scope option 1 router=192.168.1.1`,
			expected: []DHCPScope{
				{
					ScopeID:       1,
					Network:       "192.168.1.0/24",
					Options:       DHCPScopeOptions{Routers: []string{"192.168.1.1"}},
					ExcludeRanges: []ExcludeRange{},
				},
			},
		},
		{
			name: "scope with DNS and router options",
			input: `dhcp scope 1 192.168.1.0/24
dhcp scope option 1 dns=8.8.8.8 router=192.168.1.1`,
			expected: []DHCPScope{
				{
					ScopeID: 1,
					Network: "192.168.1.0/24",
					Options: DHCPScopeOptions{
						DNSServers: []string{"8.8.8.8"},
						Routers:    []string{"192.168.1.1"},
					},
					ExcludeRanges: []ExcludeRange{},
				},
			},
		},
		{
			name: "scope with exclusion range",
			input: `dhcp scope 1 192.168.1.0/24
dhcp scope 1 except 192.168.1.1-192.168.1.10`,
			expected: []DHCPScope{
				{
					ScopeID: 1,
					Network: "192.168.1.0/24",
					ExcludeRanges: []ExcludeRange{
						{Start: "192.168.1.1", End: "192.168.1.10"},
					},
				},
			},
		},
		{
			name: "full scope configuration",
			input: `dhcp scope 1 192.168.1.0/24 expire 24:00
dhcp scope option 1 dns=8.8.8.8,8.8.4.4,1.1.1.1 router=192.168.1.1
dhcp scope 1 except 192.168.1.1-192.168.1.10
dhcp scope 1 except 192.168.1.250-192.168.1.254`,
			expected: []DHCPScope{
				{
					ScopeID:   1,
					Network:   "192.168.1.0/24",
					LeaseTime: "24h",
					Options: DHCPScopeOptions{
						DNSServers: []string{"8.8.8.8", "8.8.4.4", "1.1.1.1"},
						Routers:    []string{"192.168.1.1"},
					},
					ExcludeRanges: []ExcludeRange{
						{Start: "192.168.1.1", End: "192.168.1.10"},
						{Start: "192.168.1.250", End: "192.168.1.254"},
					},
				},
			},
		},
		{
			name: "multiple scopes",
			input: `dhcp scope 1 192.168.1.0/24
dhcp scope 2 192.168.2.0/24
dhcp scope option 1 dns=8.8.8.8 router=192.168.1.1
dhcp scope option 2 dns=1.1.1.1 router=192.168.2.1`,
			expected: []DHCPScope{
				{
					ScopeID: 1,
					Network: "192.168.1.0/24",
					Options: DHCPScopeOptions{
						DNSServers: []string{"8.8.8.8"},
						Routers:    []string{"192.168.1.1"},
					},
					ExcludeRanges: []ExcludeRange{},
				},
				{
					ScopeID: 2,
					Network: "192.168.2.0/24",
					Options: DHCPScopeOptions{
						DNSServers: []string{"1.1.1.1"},
						Routers:    []string{"192.168.2.1"},
					},
					ExcludeRanges: []ExcludeRange{},
				},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []DHCPScope{},
		},
	}

	parser := NewDHCPScopeParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseScopeConfig(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d scopes, got %d", len(tt.expected), len(result))
				return
			}

			// Create a map for easier comparison (order may vary)
			resultMap := make(map[int]DHCPScope)
			for _, s := range result {
				resultMap[s.ScopeID] = s
			}

			for _, expected := range tt.expected {
				got, ok := resultMap[expected.ScopeID]
				if !ok {
					t.Errorf("scope %d not found in result", expected.ScopeID)
					continue
				}

				if got.Network != expected.Network {
					t.Errorf("scope %d: network = %q, want %q", expected.ScopeID, got.Network, expected.Network)
				}
				if got.LeaseTime != expected.LeaseTime {
					t.Errorf("scope %d: lease_time = %q, want %q", expected.ScopeID, got.LeaseTime, expected.LeaseTime)
				}
				if len(got.Options.Routers) != len(expected.Options.Routers) {
					t.Errorf("scope %d: routers count = %d, want %d", expected.ScopeID, len(got.Options.Routers), len(expected.Options.Routers))
				}
				if len(got.Options.DNSServers) != len(expected.Options.DNSServers) {
					t.Errorf("scope %d: dns_servers count = %d, want %d", expected.ScopeID, len(got.Options.DNSServers), len(expected.Options.DNSServers))
				}
				if len(got.ExcludeRanges) != len(expected.ExcludeRanges) {
					t.Errorf("scope %d: exclude_ranges count = %d, want %d", expected.ScopeID, len(got.ExcludeRanges), len(expected.ExcludeRanges))
				}
			}
		})
	}
}

func TestBuildDHCPScopeCommand(t *testing.T) {
	tests := []struct {
		name     string
		scope    DHCPScope
		expected string
	}{
		{
			name: "basic scope",
			scope: DHCPScope{
				ScopeID: 1,
				Network: "192.168.1.0/24",
			},
			expected: "dhcp scope 1 192.168.1.0/24",
		},
		{
			name: "scope with lease time",
			scope: DHCPScope{
				ScopeID:   1,
				Network:   "192.168.1.0/24",
				LeaseTime: "72h",
			},
			expected: "dhcp scope 1 192.168.1.0/24 expire 72:00",
		},
		{
			name: "scope with infinite lease",
			scope: DHCPScope{
				ScopeID:   1,
				Network:   "192.168.1.0/24",
				LeaseTime: "infinite",
			},
			expected: "dhcp scope 1 192.168.1.0/24 expire infinite",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDHCPScopeCommand(tt.scope)
			if result != tt.expected {
				t.Errorf("BuildDHCPScopeCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDHCPScopeOptionsCommand(t *testing.T) {
	tests := []struct {
		name     string
		scopeID  int
		options  DHCPScopeOptions
		expected string
	}{
		{
			name:     "single DNS",
			scopeID:  1,
			options:  DHCPScopeOptions{DNSServers: []string{"8.8.8.8"}},
			expected: "dhcp scope option 1 dns=8.8.8.8",
		},
		{
			name:     "multiple DNS",
			scopeID:  1,
			options:  DHCPScopeOptions{DNSServers: []string{"8.8.8.8", "8.8.4.4"}},
			expected: "dhcp scope option 1 dns=8.8.8.8,8.8.4.4",
		},
		{
			name:     "single router",
			scopeID:  1,
			options:  DHCPScopeOptions{Routers: []string{"192.168.1.1"}},
			expected: "dhcp scope option 1 router=192.168.1.1",
		},
		{
			name:    "DNS and router",
			scopeID: 1,
			options: DHCPScopeOptions{
				DNSServers: []string{"8.8.8.8"},
				Routers:    []string{"192.168.1.1"},
			},
			expected: "dhcp scope option 1 dns=8.8.8.8 router=192.168.1.1",
		},
		{
			name:    "DNS, router and domain",
			scopeID: 1,
			options: DHCPScopeOptions{
				DNSServers: []string{"8.8.8.8"},
				Routers:    []string{"192.168.1.1"},
				DomainName: "example.local",
			},
			expected: "dhcp scope option 1 dns=8.8.8.8 router=192.168.1.1 domain=example.local",
		},
		{
			name:    "more than three DNS (truncated)",
			scopeID: 1,
			options: DHCPScopeOptions{
				DNSServers: []string{"8.8.8.8", "8.8.4.4", "1.1.1.1", "9.9.9.9"},
			},
			expected: "dhcp scope option 1 dns=8.8.8.8,8.8.4.4,1.1.1.1",
		},
		{
			name:     "empty options",
			scopeID:  1,
			options:  DHCPScopeOptions{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDHCPScopeOptionsCommand(tt.scopeID, tt.options)
			if result != tt.expected {
				t.Errorf("BuildDHCPScopeOptionsCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDHCPScopeExceptCommand(t *testing.T) {
	tests := []struct {
		name         string
		scopeID      int
		excludeRange ExcludeRange
		expected     string
	}{
		{
			name:         "basic range",
			scopeID:      1,
			excludeRange: ExcludeRange{Start: "192.168.1.1", End: "192.168.1.10"},
			expected:     "dhcp scope 1 except 192.168.1.1-192.168.1.10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDHCPScopeExceptCommand(tt.scopeID, tt.excludeRange)
			if result != tt.expected {
				t.Errorf("BuildDHCPScopeExceptCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteDHCPScopeCommand(t *testing.T) {
	result := BuildDeleteDHCPScopeCommand(1)
	expected := "no dhcp scope 1"
	if result != expected {
		t.Errorf("BuildDeleteDHCPScopeCommand() = %q, want %q", result, expected)
	}
}

func TestBuildShowDHCPScopeCommand(t *testing.T) {
	result := BuildShowDHCPScopeCommand(1)
	expected := `show config | grep "dhcp scope 1"`
	if result != expected {
		t.Errorf("BuildShowDHCPScopeCommand() = %q, want %q", result, expected)
	}
}

func TestValidateDHCPScope(t *testing.T) {
	tests := []struct {
		name    string
		scope   DHCPScope
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid scope",
			scope: DHCPScope{
				ScopeID: 1,
				Network: "192.168.1.0/24",
			},
			wantErr: false,
		},
		{
			name: "valid full scope",
			scope: DHCPScope{
				ScopeID: 1,
				Network: "192.168.1.0/24",
				Options: DHCPScopeOptions{
					Routers:    []string{"192.168.1.1"},
					DNSServers: []string{"8.8.8.8", "8.8.4.4"},
				},
				ExcludeRanges: []ExcludeRange{
					{Start: "192.168.1.1", End: "192.168.1.10"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid scope_id",
			scope: DHCPScope{
				ScopeID: 0,
				Network: "192.168.1.0/24",
			},
			wantErr: true,
			errMsg:  "scope_id must be positive",
		},
		{
			name: "empty network",
			scope: DHCPScope{
				ScopeID: 1,
				Network: "",
			},
			wantErr: true,
			errMsg:  "network is required",
		},
		{
			name: "invalid network format",
			scope: DHCPScope{
				ScopeID: 1,
				Network: "192.168.1.0",
			},
			wantErr: true,
			errMsg:  "network must be in CIDR notation",
		},
		{
			name: "invalid router",
			scope: DHCPScope{
				ScopeID: 1,
				Network: "192.168.1.0/24",
				Options: DHCPScopeOptions{Routers: []string{"invalid"}},
			},
			wantErr: true,
			errMsg:  "invalid router address",
		},
		{
			name: "too many DNS servers",
			scope: DHCPScope{
				ScopeID: 1,
				Network: "192.168.1.0/24",
				Options: DHCPScopeOptions{
					DNSServers: []string{"8.8.8.8", "8.8.4.4", "1.1.1.1", "9.9.9.9"},
				},
			},
			wantErr: true,
			errMsg:  "maximum 3 DNS servers allowed",
		},
		{
			name: "invalid DNS server",
			scope: DHCPScope{
				ScopeID: 1,
				Network: "192.168.1.0/24",
				Options: DHCPScopeOptions{DNSServers: []string{"invalid"}},
			},
			wantErr: true,
			errMsg:  "invalid DNS server address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDHCPScope(tt.scope)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestConvertLeaseTime(t *testing.T) {
	tests := []struct {
		name    string
		goTime  string
		rtxTime string
	}{
		{"hours only", "72h", "72:00"},
		{"minutes only", "30m", "0:30"},
		{"hours and minutes", "1h30m", "1:30"},
		{"infinite", "infinite", "infinite"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name+" (go to rtx)", func(t *testing.T) {
			result := convertGoLeaseTimeToRTX(tt.goTime)
			if result != tt.rtxTime {
				t.Errorf("convertGoLeaseTimeToRTX(%q) = %q, want %q", tt.goTime, result, tt.rtxTime)
			}
		})

		if tt.goTime != "" && tt.rtxTime != "" {
			t.Run(tt.name+" (rtx to go)", func(t *testing.T) {
				result := convertRTXLeaseTimeToGo(tt.rtxTime)
				// For round-trip, we need to normalize
				if tt.goTime == "infinite" {
					if result != "infinite" {
						t.Errorf("convertRTXLeaseTimeToGo(%q) = %q, want %q", tt.rtxTime, result, "infinite")
					}
				}
			})
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
