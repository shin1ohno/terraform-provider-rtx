package parsers

import (
	"strings"
	"testing"
)

func TestParseNATStaticConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []NATStatic
		wantErr  bool
	}{
		{
			name: "simple 1:1 static NAT",
			input: `
nat descriptor type 1 static
nat descriptor static 1 203.0.113.1=192.168.1.1
`,
			expected: []NATStatic{
				{
					DescriptorID: 1,
					Entries: []NATStaticEntry{
						{
							OutsideGlobal: "203.0.113.1",
							InsideLocal:   "192.168.1.1",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "port-based static NAT",
			input: `
nat descriptor type 10 static
nat descriptor static 10 203.0.113.1:80=192.168.1.1:8080 tcp
`,
			expected: []NATStatic{
				{
					DescriptorID: 10,
					Entries: []NATStaticEntry{
						{
							OutsideGlobal:     "203.0.113.1",
							OutsideGlobalPort: 80,
							InsideLocal:       "192.168.1.1",
							InsideLocalPort:   8080,
							Protocol:          "tcp",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple entries in one descriptor",
			input: `
nat descriptor type 5 static
nat descriptor static 5 203.0.113.10=192.168.1.10
nat descriptor static 5 203.0.113.11=192.168.1.11
nat descriptor static 5 203.0.113.1:443=192.168.1.1:8443 tcp
`,
			expected: []NATStatic{
				{
					DescriptorID: 5,
					Entries: []NATStaticEntry{
						{
							OutsideGlobal: "203.0.113.10",
							InsideLocal:   "192.168.1.10",
						},
						{
							OutsideGlobal: "203.0.113.11",
							InsideLocal:   "192.168.1.11",
						},
						{
							OutsideGlobal:     "203.0.113.1",
							OutsideGlobalPort: 443,
							InsideLocal:       "192.168.1.1",
							InsideLocalPort:   8443,
							Protocol:          "tcp",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "UDP port-based NAT",
			input: `
nat descriptor type 20 static
nat descriptor static 20 203.0.113.1:53=192.168.1.1:5353 udp
`,
			expected: []NATStatic{
				{
					DescriptorID: 20,
					Entries: []NATStaticEntry{
						{
							OutsideGlobal:     "203.0.113.1",
							OutsideGlobalPort: 53,
							InsideLocal:       "192.168.1.1",
							InsideLocalPort:   5353,
							Protocol:          "udp",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple descriptors",
			input: `
nat descriptor type 1 static
nat descriptor static 1 203.0.113.1=192.168.1.1
nat descriptor type 2 static
nat descriptor static 2 203.0.113.2=192.168.1.2
`,
			expected: []NATStatic{
				{
					DescriptorID: 1,
					Entries: []NATStaticEntry{
						{
							OutsideGlobal: "203.0.113.1",
							InsideLocal:   "192.168.1.1",
						},
					},
				},
				{
					DescriptorID: 2,
					Entries: []NATStaticEntry{
						{
							OutsideGlobal: "203.0.113.2",
							InsideLocal:   "192.168.1.2",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:     "empty input",
			input:    "",
			expected: []NATStatic{},
			wantErr:  false,
		},
		{
			name: "entry without type definition",
			input: `
nat descriptor static 99 203.0.113.1=192.168.1.1
`,
			expected: []NATStatic{
				{
					DescriptorID: 99,
					Entries: []NATStaticEntry{
						{
							OutsideGlobal: "203.0.113.1",
							InsideLocal:   "192.168.1.1",
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseNATStaticConfig(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNATStaticConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("ParseNATStaticConfig() got %d descriptors, want %d", len(result), len(tt.expected))
				return
			}

			// Create a map for easier comparison since order is not guaranteed
			resultMap := make(map[int]NATStatic)
			for _, nat := range result {
				resultMap[nat.DescriptorID] = nat
			}

			for _, expected := range tt.expected {
				got, ok := resultMap[expected.DescriptorID]
				if !ok {
					t.Errorf("ParseNATStaticConfig() missing descriptor ID %d", expected.DescriptorID)
					continue
				}

				if len(got.Entries) != len(expected.Entries) {
					t.Errorf("ParseNATStaticConfig() descriptor %d: got %d entries, want %d",
						expected.DescriptorID, len(got.Entries), len(expected.Entries))
					continue
				}

				for i, expectedEntry := range expected.Entries {
					gotEntry := got.Entries[i]
					if gotEntry.InsideLocal != expectedEntry.InsideLocal {
						t.Errorf("Entry %d: InsideLocal = %s, want %s", i, gotEntry.InsideLocal, expectedEntry.InsideLocal)
					}
					if gotEntry.InsideLocalPort != expectedEntry.InsideLocalPort {
						t.Errorf("Entry %d: InsideLocalPort = %d, want %d", i, gotEntry.InsideLocalPort, expectedEntry.InsideLocalPort)
					}
					if gotEntry.OutsideGlobal != expectedEntry.OutsideGlobal {
						t.Errorf("Entry %d: OutsideGlobal = %s, want %s", i, gotEntry.OutsideGlobal, expectedEntry.OutsideGlobal)
					}
					if gotEntry.OutsideGlobalPort != expectedEntry.OutsideGlobalPort {
						t.Errorf("Entry %d: OutsideGlobalPort = %d, want %d", i, gotEntry.OutsideGlobalPort, expectedEntry.OutsideGlobalPort)
					}
					if gotEntry.Protocol != expectedEntry.Protocol {
						t.Errorf("Entry %d: Protocol = %s, want %s", i, gotEntry.Protocol, expectedEntry.Protocol)
					}
				}
			}
		})
	}
}

func TestBuildNATDescriptorTypeStaticCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		expected string
	}{
		{
			name:     "descriptor ID 1",
			id:       1,
			expected: "nat descriptor type 1 static",
		},
		{
			name:     "descriptor ID 100",
			id:       100,
			expected: "nat descriptor type 100 static",
		},
		{
			name:     "maximum descriptor ID",
			id:       65535,
			expected: "nat descriptor type 65535 static",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNATDescriptorTypeStaticCommand(tt.id)
			if result != tt.expected {
				t.Errorf("BuildNATDescriptorTypeStaticCommand(%d) = %s, want %s", tt.id, result, tt.expected)
			}
		})
	}
}

func TestBuildNATStaticMappingCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		entry    NATStaticEntry
		expected string
	}{
		{
			name: "simple 1:1 mapping",
			id:   1,
			entry: NATStaticEntry{
				OutsideGlobal: "203.0.113.1",
				InsideLocal:   "192.168.1.1",
			},
			expected: "nat descriptor static 1 203.0.113.1=192.168.1.1",
		},
		{
			name: "different IPs",
			id:   10,
			entry: NATStaticEntry{
				OutsideGlobal: "198.51.100.50",
				InsideLocal:   "10.0.0.100",
			},
			expected: "nat descriptor static 10 198.51.100.50=10.0.0.100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNATStaticMappingCommand(tt.id, tt.entry)
			if result != tt.expected {
				t.Errorf("BuildNATStaticMappingCommand() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestBuildNATStaticPortMappingCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		entry    NATStaticEntry
		expected string
	}{
		{
			name: "TCP port mapping",
			id:   1,
			entry: NATStaticEntry{
				OutsideGlobal:     "203.0.113.1",
				OutsideGlobalPort: 80,
				InsideLocal:       "192.168.1.1",
				InsideLocalPort:   8080,
				Protocol:          "tcp",
			},
			expected: "nat descriptor static 1 203.0.113.1:80=192.168.1.1:8080 tcp",
		},
		{
			name: "UDP port mapping",
			id:   5,
			entry: NATStaticEntry{
				OutsideGlobal:     "203.0.113.1",
				OutsideGlobalPort: 53,
				InsideLocal:       "192.168.1.10",
				InsideLocalPort:   5353,
				Protocol:          "UDP",
			},
			expected: "nat descriptor static 5 203.0.113.1:53=192.168.1.10:5353 udp",
		},
		{
			name: "HTTPS port mapping",
			id:   10,
			entry: NATStaticEntry{
				OutsideGlobal:     "198.51.100.1",
				OutsideGlobalPort: 443,
				InsideLocal:       "10.0.0.1",
				InsideLocalPort:   8443,
				Protocol:          "tcp",
			},
			expected: "nat descriptor static 10 198.51.100.1:443=10.0.0.1:8443 tcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNATStaticPortMappingCommand(tt.id, tt.entry)
			if result != tt.expected {
				t.Errorf("BuildNATStaticPortMappingCommand() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteNATStaticCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		expected string
	}{
		{
			name:     "delete descriptor 1",
			id:       1,
			expected: "no nat descriptor type 1",
		},
		{
			name:     "delete descriptor 100",
			id:       100,
			expected: "no nat descriptor type 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteNATStaticCommand(tt.id)
			if result != tt.expected {
				t.Errorf("BuildDeleteNATStaticCommand(%d) = %s, want %s", tt.id, result, tt.expected)
			}
		})
	}
}

func TestBuildInterfaceNATCommand(t *testing.T) {
	tests := []struct {
		name         string
		iface        string
		descriptorID int
		expected     string
	}{
		{
			name:         "lan1 interface",
			iface:        "lan1",
			descriptorID: 1,
			expected:     "ip lan1 nat descriptor 1",
		},
		{
			name:         "lan2 interface",
			iface:        "lan2",
			descriptorID: 10,
			expected:     "ip lan2 nat descriptor 10",
		},
		{
			name:         "pp1 interface",
			iface:        "pp1",
			descriptorID: 100,
			expected:     "ip pp1 nat descriptor 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildInterfaceNATCommand(tt.iface, tt.descriptorID)
			if result != tt.expected {
				t.Errorf("BuildInterfaceNATCommand() = %s, want %s", result, tt.expected)
			}
		})
	}
}

// Note: ValidateDescriptorID and ValidatePort tests are in nat_masquerade_test.go
// as those functions are defined there and shared across NAT parsers

func TestValidateNATStaticProtocol(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		wantErr  bool
	}{
		{
			name:     "valid tcp lowercase",
			protocol: "tcp",
			wantErr:  false,
		},
		{
			name:     "valid udp lowercase",
			protocol: "udp",
			wantErr:  false,
		},
		{
			name:     "valid TCP uppercase",
			protocol: "TCP",
			wantErr:  false,
		},
		{
			name:     "valid UDP uppercase",
			protocol: "UDP",
			wantErr:  false,
		},
		{
			name:     "invalid protocol icmp",
			protocol: "icmp",
			wantErr:  true,
		},
		{
			name:     "invalid empty protocol",
			protocol: "",
			wantErr:  true,
		},
		{
			name:     "invalid protocol any",
			protocol: "any",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNATStaticProtocol(tt.protocol)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateNATStaticProtocol(%s) error = %v, wantErr %v", tt.protocol, err, tt.wantErr)
			}
		})
	}
}

func TestValidateNATStaticEntry(t *testing.T) {
	tests := []struct {
		name    string
		entry   NATStaticEntry
		wantErr bool
	}{
		{
			name: "valid simple entry",
			entry: NATStaticEntry{
				InsideLocal:   "192.168.1.1",
				OutsideGlobal: "203.0.113.1",
			},
			wantErr: false,
		},
		{
			name: "valid port-based entry",
			entry: NATStaticEntry{
				InsideLocal:       "192.168.1.1",
				InsideLocalPort:   8080,
				OutsideGlobal:     "203.0.113.1",
				OutsideGlobalPort: 80,
				Protocol:          "tcp",
			},
			wantErr: false,
		},
		{
			name: "missing inside_local",
			entry: NATStaticEntry{
				OutsideGlobal: "203.0.113.1",
			},
			wantErr: true,
		},
		{
			name: "missing outside_global",
			entry: NATStaticEntry{
				InsideLocal: "192.168.1.1",
			},
			wantErr: true,
		},
		{
			name: "invalid inside_local IP",
			entry: NATStaticEntry{
				InsideLocal:   "invalid",
				OutsideGlobal: "203.0.113.1",
			},
			wantErr: true,
		},
		{
			name: "invalid outside_global IP",
			entry: NATStaticEntry{
				InsideLocal:   "192.168.1.1",
				OutsideGlobal: "invalid",
			},
			wantErr: true,
		},
		{
			name: "port NAT missing inside port",
			entry: NATStaticEntry{
				InsideLocal:       "192.168.1.1",
				OutsideGlobal:     "203.0.113.1",
				OutsideGlobalPort: 80,
				Protocol:          "tcp",
			},
			wantErr: true,
		},
		{
			name: "port NAT missing outside port",
			entry: NATStaticEntry{
				InsideLocal:     "192.168.1.1",
				InsideLocalPort: 8080,
				OutsideGlobal:   "203.0.113.1",
				Protocol:        "tcp",
			},
			wantErr: true,
		},
		{
			name: "port NAT missing protocol",
			entry: NATStaticEntry{
				InsideLocal:       "192.168.1.1",
				InsideLocalPort:   8080,
				OutsideGlobal:     "203.0.113.1",
				OutsideGlobalPort: 80,
			},
			wantErr: true,
		},
		{
			name: "port NAT invalid protocol",
			entry: NATStaticEntry{
				InsideLocal:       "192.168.1.1",
				InsideLocalPort:   8080,
				OutsideGlobal:     "203.0.113.1",
				OutsideGlobalPort: 80,
				Protocol:          "icmp",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNATStaticEntry(tt.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNATStaticEntry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNATStatic(t *testing.T) {
	tests := []struct {
		name    string
		nat     NATStatic
		wantErr bool
	}{
		{
			name: "valid configuration",
			nat: NATStatic{
				DescriptorID: 1,
				Entries: []NATStaticEntry{
					{
						InsideLocal:   "192.168.1.1",
						OutsideGlobal: "203.0.113.1",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid descriptor ID",
			nat: NATStatic{
				DescriptorID: 0,
				Entries: []NATStaticEntry{
					{
						InsideLocal:   "192.168.1.1",
						OutsideGlobal: "203.0.113.1",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid entry",
			nat: NATStatic{
				DescriptorID: 1,
				Entries: []NATStaticEntry{
					{
						InsideLocal: "invalid",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty entries is valid",
			nat: NATStatic{
				DescriptorID: 1,
				Entries:      []NATStaticEntry{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNATStatic(tt.nat)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNATStatic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsPortBasedNAT(t *testing.T) {
	tests := []struct {
		name     string
		entry    NATStaticEntry
		expected bool
	}{
		{
			name: "simple 1:1 NAT",
			entry: NATStaticEntry{
				InsideLocal:   "192.168.1.1",
				OutsideGlobal: "203.0.113.1",
			},
			expected: false,
		},
		{
			name: "port-based NAT",
			entry: NATStaticEntry{
				InsideLocal:       "192.168.1.1",
				InsideLocalPort:   8080,
				OutsideGlobal:     "203.0.113.1",
				OutsideGlobalPort: 80,
				Protocol:          "tcp",
			},
			expected: true,
		},
		{
			name: "partial port NAT (missing protocol)",
			entry: NATStaticEntry{
				InsideLocal:       "192.168.1.1",
				InsideLocalPort:   8080,
				OutsideGlobal:     "203.0.113.1",
				OutsideGlobalPort: 80,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPortBasedNAT(tt.entry)
			if result != tt.expected {
				t.Errorf("IsPortBasedNAT() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildNATStaticCommands(t *testing.T) {
	tests := []struct {
		name     string
		nat      NATStatic
		expected []string
	}{
		{
			name: "simple 1:1 NAT",
			nat: NATStatic{
				DescriptorID: 1,
				Entries: []NATStaticEntry{
					{
						InsideLocal:   "192.168.1.1",
						OutsideGlobal: "203.0.113.1",
					},
				},
			},
			expected: []string{
				"nat descriptor type 1 static",
				"nat descriptor static 1 203.0.113.1=192.168.1.1",
			},
		},
		{
			name: "port-based NAT",
			nat: NATStatic{
				DescriptorID: 10,
				Entries: []NATStaticEntry{
					{
						InsideLocal:       "192.168.1.1",
						InsideLocalPort:   8080,
						OutsideGlobal:     "203.0.113.1",
						OutsideGlobalPort: 80,
						Protocol:          "tcp",
					},
				},
			},
			expected: []string{
				"nat descriptor type 10 static",
				"nat descriptor static 10 203.0.113.1:80=192.168.1.1:8080 tcp",
			},
		},
		{
			name: "mixed NAT entries",
			nat: NATStatic{
				DescriptorID: 5,
				Entries: []NATStaticEntry{
					{
						InsideLocal:   "192.168.1.10",
						OutsideGlobal: "203.0.113.10",
					},
					{
						InsideLocal:       "192.168.1.1",
						InsideLocalPort:   8080,
						OutsideGlobal:     "203.0.113.1",
						OutsideGlobalPort: 80,
						Protocol:          "tcp",
					},
				},
			},
			expected: []string{
				"nat descriptor type 5 static",
				"nat descriptor static 5 203.0.113.10=192.168.1.10",
				"nat descriptor static 5 203.0.113.1:80=192.168.1.1:8080 tcp",
			},
		},
		{
			name: "no entries",
			nat: NATStatic{
				DescriptorID: 1,
				Entries:      []NATStaticEntry{},
			},
			expected: []string{
				"nat descriptor type 1 static",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNATStaticCommands(tt.nat)
			if len(result) != len(tt.expected) {
				t.Errorf("BuildNATStaticCommands() returned %d commands, want %d", len(result), len(tt.expected))
				return
			}
			for i, cmd := range result {
				if cmd != tt.expected[i] {
					t.Errorf("BuildNATStaticCommands()[%d] = %s, want %s", i, cmd, tt.expected[i])
				}
			}
		})
	}
}

func TestBuildDeleteNATStaticMappingCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		entry    NATStaticEntry
		expected string
	}{
		{
			name: "delete 1:1 mapping",
			id:   1,
			entry: NATStaticEntry{
				OutsideGlobal: "203.0.113.1",
				InsideLocal:   "192.168.1.1",
			},
			expected: "no nat descriptor static 1 203.0.113.1=192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteNATStaticMappingCommand(tt.id, tt.entry)
			if result != tt.expected {
				t.Errorf("BuildDeleteNATStaticMappingCommand() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteNATStaticPortMappingCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		entry    NATStaticEntry
		expected string
	}{
		{
			name: "delete port mapping",
			id:   1,
			entry: NATStaticEntry{
				OutsideGlobal:     "203.0.113.1",
				OutsideGlobalPort: 80,
				InsideLocal:       "192.168.1.1",
				InsideLocalPort:   8080,
				Protocol:          "tcp",
			},
			expected: "no nat descriptor static 1 203.0.113.1:80=192.168.1.1:8080 tcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteNATStaticPortMappingCommand(tt.id, tt.entry)
			if result != tt.expected {
				t.Errorf("BuildDeleteNATStaticPortMappingCommand() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteInterfaceNATCommand(t *testing.T) {
	tests := []struct {
		name         string
		iface        string
		descriptorID int
		expected     string
	}{
		{
			name:         "delete lan2 NAT binding",
			iface:        "lan2",
			descriptorID: 1,
			expected:     "no ip lan2 nat descriptor 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteInterfaceNATCommand(tt.iface, tt.descriptorID)
			if result != tt.expected {
				t.Errorf("BuildDeleteInterfaceNATCommand() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestBuildShowNATStaticCommand(t *testing.T) {
	tests := []struct {
		name         string
		descriptorID int
		expected     string
	}{
		{
			name:         "show descriptor 1",
			descriptorID: 1,
			expected:     "show config | grep \"nat descriptor.*1\"",
		},
		{
			name:         "show descriptor 100",
			descriptorID: 100,
			expected:     "show config | grep \"nat descriptor.*100\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildShowNATStaticCommand(tt.descriptorID)
			if result != tt.expected {
				t.Errorf("BuildShowNATStaticCommand() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestBuildShowAllNATStaticCommand(t *testing.T) {
	expected := "show config | grep \"nat descriptor\""
	result := BuildShowAllNATStaticCommand()
	if result != expected {
		t.Errorf("BuildShowAllNATStaticCommand() = %s, want %s", result, expected)
	}
}

func TestNATStaticParserParseSingleNATStatic(t *testing.T) {
	parser := NewNATStaticParser()

	input := `
nat descriptor type 1 static
nat descriptor static 1 203.0.113.1=192.168.1.1
nat descriptor type 2 static
nat descriptor static 2 203.0.113.2=192.168.1.2
`

	// Test finding existing descriptor
	result, err := parser.ParseSingleNATStatic(input, 1)
	if err != nil {
		t.Errorf("ParseSingleNATStatic() error = %v", err)
		return
	}
	if result.DescriptorID != 1 {
		t.Errorf("ParseSingleNATStatic() DescriptorID = %d, want 1", result.DescriptorID)
	}

	// Test finding non-existing descriptor
	_, err = parser.ParseSingleNATStatic(input, 99)
	if err == nil {
		t.Errorf("ParseSingleNATStatic() expected error for non-existing descriptor")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("ParseSingleNATStatic() error = %v, want error containing 'not found'", err)
	}
}
