package parsers

import (
	"testing"
)

func TestParseIPv6PrefixConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []IPv6Prefix
		wantErr  bool
	}{
		{
			name:  "static prefix",
			input: `ipv6 prefix 1 2001:db8:1234::/64`,
			expected: []IPv6Prefix{
				{
					ID:           1,
					Prefix:       "2001:db8:1234::",
					PrefixLength: 64,
					Source:       "static",
					Interface:    "",
				},
			},
		},
		{
			name:  "RA-derived prefix",
			input: `ipv6 prefix 2 ra-prefix@lan2::/64`,
			expected: []IPv6Prefix{
				{
					ID:           2,
					Prefix:       "",
					PrefixLength: 64,
					Source:       "ra",
					Interface:    "lan2",
				},
			},
		},
		{
			name:  "DHCPv6-PD prefix",
			input: `ipv6 prefix 3 dhcp-prefix@lan2::/48`,
			expected: []IPv6Prefix{
				{
					ID:           3,
					Prefix:       "",
					PrefixLength: 48,
					Source:       "dhcpv6-pd",
					Interface:    "lan2",
				},
			},
		},
		{
			name: "multiple prefixes",
			input: `ipv6 prefix 1 2001:db8:1234::/64
ipv6 prefix 2 ra-prefix@lan2::/64
ipv6 prefix 3 dhcp-prefix@lan2::/48`,
			expected: []IPv6Prefix{
				{
					ID:           1,
					Prefix:       "2001:db8:1234::",
					PrefixLength: 64,
					Source:       "static",
					Interface:    "",
				},
				{
					ID:           2,
					Prefix:       "",
					PrefixLength: 64,
					Source:       "ra",
					Interface:    "lan2",
				},
				{
					ID:           3,
					Prefix:       "",
					PrefixLength: 48,
					Source:       "dhcpv6-pd",
					Interface:    "lan2",
				},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []IPv6Prefix{},
		},
		{
			name:  "prefix with pp interface",
			input: `ipv6 prefix 10 ra-prefix@pp1::/64`,
			expected: []IPv6Prefix{
				{
					ID:           10,
					Prefix:       "",
					PrefixLength: 64,
					Source:       "ra",
					Interface:    "pp1",
				},
			},
		},
		{
			name:  "static prefix with full address",
			input: `ipv6 prefix 5 2001:db8:abcd:ef01::/56`,
			expected: []IPv6Prefix{
				{
					ID:           5,
					Prefix:       "2001:db8:abcd:ef01::",
					PrefixLength: 56,
					Source:       "static",
					Interface:    "",
				},
			},
		},
		{
			name:  "whitespace handling",
			input: `  ipv6 prefix 1 2001:db8::/32  `,
			expected: []IPv6Prefix{
				{
					ID:           1,
					Prefix:       "2001:db8::",
					PrefixLength: 32,
					Source:       "static",
					Interface:    "",
				},
			},
		},
	}

	parser := NewIPv6PrefixParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseIPv6PrefixConfig(tt.input)

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
				t.Errorf("expected %d prefixes, got %d", len(tt.expected), len(result))
				return
			}

			// Create a map for easier comparison (order may vary)
			resultMap := make(map[int]IPv6Prefix)
			for _, p := range result {
				resultMap[p.ID] = p
			}

			for _, expected := range tt.expected {
				got, ok := resultMap[expected.ID]
				if !ok {
					t.Errorf("prefix %d not found in result", expected.ID)
					continue
				}

				if got.Prefix != expected.Prefix {
					t.Errorf("prefix %d: Prefix = %q, want %q", expected.ID, got.Prefix, expected.Prefix)
				}
				if got.PrefixLength != expected.PrefixLength {
					t.Errorf("prefix %d: PrefixLength = %d, want %d", expected.ID, got.PrefixLength, expected.PrefixLength)
				}
				if got.Source != expected.Source {
					t.Errorf("prefix %d: Source = %q, want %q", expected.ID, got.Source, expected.Source)
				}
				if got.Interface != expected.Interface {
					t.Errorf("prefix %d: Interface = %q, want %q", expected.ID, got.Interface, expected.Interface)
				}
			}
		})
	}
}

func TestParseSingleIPv6Prefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		prefixID int
		expected *IPv6Prefix
		wantErr  bool
	}{
		{
			name:     "find existing prefix",
			input:    "ipv6 prefix 1 2001:db8::/64\nipv6 prefix 2 ra-prefix@lan2::/64",
			prefixID: 1,
			expected: &IPv6Prefix{
				ID:           1,
				Prefix:       "2001:db8::",
				PrefixLength: 64,
				Source:       "static",
				Interface:    "",
			},
		},
		{
			name:     "prefix not found",
			input:    "ipv6 prefix 1 2001:db8::/64",
			prefixID: 99,
			expected: nil,
			wantErr:  true,
		},
	}

	parser := NewIPv6PrefixParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseSinglePrefix(tt.input, tt.prefixID)

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

			if result.ID != tt.expected.ID {
				t.Errorf("ID = %d, want %d", result.ID, tt.expected.ID)
			}
			if result.Prefix != tt.expected.Prefix {
				t.Errorf("Prefix = %q, want %q", result.Prefix, tt.expected.Prefix)
			}
			if result.PrefixLength != tt.expected.PrefixLength {
				t.Errorf("PrefixLength = %d, want %d", result.PrefixLength, tt.expected.PrefixLength)
			}
			if result.Source != tt.expected.Source {
				t.Errorf("Source = %q, want %q", result.Source, tt.expected.Source)
			}
			if result.Interface != tt.expected.Interface {
				t.Errorf("Interface = %q, want %q", result.Interface, tt.expected.Interface)
			}
		})
	}
}

func TestBuildIPv6PrefixCommand(t *testing.T) {
	tests := []struct {
		name     string
		prefix   IPv6Prefix
		expected string
	}{
		{
			name: "static prefix",
			prefix: IPv6Prefix{
				ID:           1,
				Prefix:       "2001:db8:1234::",
				PrefixLength: 64,
				Source:       "static",
			},
			expected: "ipv6 prefix 1 2001:db8:1234::/64",
		},
		{
			name: "RA-derived prefix",
			prefix: IPv6Prefix{
				ID:           2,
				PrefixLength: 64,
				Source:       "ra",
				Interface:    "lan2",
			},
			expected: "ipv6 prefix 2 ra-prefix@lan2::/64",
		},
		{
			name: "DHCPv6-PD prefix",
			prefix: IPv6Prefix{
				ID:           3,
				PrefixLength: 48,
				Source:       "dhcpv6-pd",
				Interface:    "lan2",
			},
			expected: "ipv6 prefix 3 dhcp-prefix@lan2::/48",
		},
		{
			name: "static prefix with different length",
			prefix: IPv6Prefix{
				ID:           10,
				Prefix:       "2001:db8:abcd::",
				PrefixLength: 56,
				Source:       "static",
			},
			expected: "ipv6 prefix 10 2001:db8:abcd::/56",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildIPv6PrefixCommand(tt.prefix)
			if result != tt.expected {
				t.Errorf("BuildIPv6PrefixCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteIPv6PrefixCommand(t *testing.T) {
	tests := []struct {
		name     string
		prefixID int
		expected string
	}{
		{
			name:     "delete prefix 1",
			prefixID: 1,
			expected: "no ipv6 prefix 1",
		},
		{
			name:     "delete prefix 255",
			prefixID: 255,
			expected: "no ipv6 prefix 255",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteIPv6PrefixCommand(tt.prefixID)
			if result != tt.expected {
				t.Errorf("BuildDeleteIPv6PrefixCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildShowIPv6PrefixCommand(t *testing.T) {
	result := BuildShowIPv6PrefixCommand(1)
	expected := `show config | grep "ipv6 prefix 1"`
	if result != expected {
		t.Errorf("BuildShowIPv6PrefixCommand() = %q, want %q", result, expected)
	}
}

func TestBuildShowAllIPv6PrefixesCommand(t *testing.T) {
	result := BuildShowAllIPv6PrefixesCommand()
	expected := `show config | grep "ipv6 prefix"`
	if result != expected {
		t.Errorf("BuildShowAllIPv6PrefixesCommand() = %q, want %q", result, expected)
	}
}

func TestValidateIPv6Prefix(t *testing.T) {
	tests := []struct {
		name    string
		prefix  IPv6Prefix
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid static prefix",
			prefix: IPv6Prefix{
				ID:           1,
				Prefix:       "2001:db8::",
				PrefixLength: 64,
				Source:       "static",
			},
			wantErr: false,
		},
		{
			name: "valid RA prefix",
			prefix: IPv6Prefix{
				ID:           2,
				PrefixLength: 64,
				Source:       "ra",
				Interface:    "lan2",
			},
			wantErr: false,
		},
		{
			name: "valid DHCPv6-PD prefix",
			prefix: IPv6Prefix{
				ID:           3,
				PrefixLength: 48,
				Source:       "dhcpv6-pd",
				Interface:    "lan2",
			},
			wantErr: false,
		},
		{
			name: "invalid prefix ID - zero",
			prefix: IPv6Prefix{
				ID:           0,
				Prefix:       "2001:db8::",
				PrefixLength: 64,
				Source:       "static",
			},
			wantErr: true,
			errMsg:  "prefix ID must be between 1 and 255",
		},
		{
			name: "invalid prefix ID - too large",
			prefix: IPv6Prefix{
				ID:           256,
				Prefix:       "2001:db8::",
				PrefixLength: 64,
				Source:       "static",
			},
			wantErr: true,
			errMsg:  "prefix ID must be between 1 and 255",
		},
		{
			name: "invalid prefix length - zero",
			prefix: IPv6Prefix{
				ID:           1,
				Prefix:       "2001:db8::",
				PrefixLength: 0,
				Source:       "static",
			},
			wantErr: true,
			errMsg:  "prefix length must be between 1 and 128",
		},
		{
			name: "invalid prefix length - too large",
			prefix: IPv6Prefix{
				ID:           1,
				Prefix:       "2001:db8::",
				PrefixLength: 129,
				Source:       "static",
			},
			wantErr: true,
			errMsg:  "prefix length must be between 1 and 128",
		},
		{
			name: "invalid source",
			prefix: IPv6Prefix{
				ID:           1,
				Prefix:       "2001:db8::",
				PrefixLength: 64,
				Source:       "invalid",
			},
			wantErr: true,
			errMsg:  "source must be one of: static, ra, dhcpv6-pd",
		},
		{
			name: "static prefix missing prefix value",
			prefix: IPv6Prefix{
				ID:           1,
				Prefix:       "",
				PrefixLength: 64,
				Source:       "static",
			},
			wantErr: true,
			errMsg:  "prefix is required for static source",
		},
		{
			name: "RA prefix missing interface",
			prefix: IPv6Prefix{
				ID:           2,
				PrefixLength: 64,
				Source:       "ra",
				Interface:    "",
			},
			wantErr: true,
			errMsg:  "interface is required for ra source",
		},
		{
			name: "DHCPv6-PD prefix missing interface",
			prefix: IPv6Prefix{
				ID:           3,
				PrefixLength: 48,
				Source:       "dhcpv6-pd",
				Interface:    "",
			},
			wantErr: true,
			errMsg:  "interface is required for dhcpv6-pd source",
		},
		{
			name: "invalid IPv6 prefix format",
			prefix: IPv6Prefix{
				ID:           1,
				Prefix:       "not-an-ipv6",
				PrefixLength: 64,
				Source:       "static",
			},
			wantErr: true,
			errMsg:  "invalid IPv6 prefix format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPv6Prefix(tt.prefix)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !ipv6ContainsString(err.Error(), tt.errMsg) {
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

// ipv6ContainsString checks if s contains substr
func ipv6ContainsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
