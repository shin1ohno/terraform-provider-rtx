package parsers

import (
	"fmt"
	"testing"
)

func TestParseNATMasqueradeConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []NATMasquerade
		wantErr  bool
	}{
		{
			name: "basic masquerade configuration",
			input: `nat descriptor type 1 masquerade
nat descriptor address outer 1 ipcp
nat descriptor address inner 1 192.168.1.0-192.168.1.255`,
			expected: []NATMasquerade{
				{
					DescriptorID:  1,
					OuterAddress:  "ipcp",
					InnerNetwork:  "192.168.1.0-192.168.1.255",
					StaticEntries: []MasqueradeStaticEntry{},
				},
			},
			wantErr: false,
		},
		{
			name: "masquerade with static entry",
			input: `nat descriptor type 2 masquerade
nat descriptor address outer 2 203.0.113.1
nat descriptor address inner 2 192.168.2.0-192.168.2.255
nat descriptor masquerade static 2 1 203.0.113.1:80=192.168.2.100:8080 tcp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 2,
					OuterAddress: "203.0.113.1",
					InnerNetwork: "192.168.2.0-192.168.2.255",
					StaticEntries: []MasqueradeStaticEntry{
						{
							EntryNumber:       1,
							OutsideGlobal:     "203.0.113.1",
							OutsideGlobalPort: intPtr(80),
							InsideLocal:       "192.168.2.100",
							InsideLocalPort:   intPtr(8080),
							Protocol:          "tcp",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "masquerade with multiple static entries",
			input: `nat descriptor type 3 masquerade
nat descriptor address outer 3 ipcp
nat descriptor address inner 3 10.0.0.0-10.0.0.255
nat descriptor masquerade static 3 1 ipcp:80=10.0.0.10:80 tcp
nat descriptor masquerade static 3 2 ipcp:443=10.0.0.10:443 tcp
nat descriptor masquerade static 3 3 ipcp:53=10.0.0.20:53 udp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 3,
					OuterAddress: "ipcp",
					InnerNetwork: "10.0.0.0-10.0.0.255",
					StaticEntries: []MasqueradeStaticEntry{
						{
							EntryNumber:       1,
							OutsideGlobal:     "ipcp",
							OutsideGlobalPort: intPtr(80),
							InsideLocal:       "10.0.0.10",
							InsideLocalPort:   intPtr(80),
							Protocol:          "tcp",
						},
						{
							EntryNumber:       2,
							OutsideGlobal:     "ipcp",
							OutsideGlobalPort: intPtr(443),
							InsideLocal:       "10.0.0.10",
							InsideLocalPort:   intPtr(443),
							Protocol:          "tcp",
						},
						{
							EntryNumber:       3,
							OutsideGlobal:     "ipcp",
							OutsideGlobalPort: intPtr(53),
							InsideLocal:       "10.0.0.20",
							InsideLocalPort:   intPtr(53),
							Protocol:          "udp",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "static entry without protocol",
			input: `nat descriptor type 4 masquerade
nat descriptor masquerade static 4 1 203.0.113.1:22=192.168.1.1:22`,
			expected: []NATMasquerade{
				{
					DescriptorID: 4,
					StaticEntries: []MasqueradeStaticEntry{
						{
							EntryNumber:       1,
							OutsideGlobal:     "203.0.113.1",
							OutsideGlobalPort: intPtr(22),
							InsideLocal:       "192.168.1.1",
							InsideLocalPort:   intPtr(22),
							Protocol:          "",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple descriptors",
			input: `nat descriptor type 1 masquerade
nat descriptor address outer 1 ipcp
nat descriptor address inner 1 192.168.1.0-192.168.1.255
nat descriptor type 2 masquerade
nat descriptor address outer 2 pp1
nat descriptor address inner 2 192.168.2.0-192.168.2.255`,
			expected: []NATMasquerade{
				{
					DescriptorID:  1,
					OuterAddress:  "ipcp",
					InnerNetwork:  "192.168.1.0-192.168.1.255",
					StaticEntries: []MasqueradeStaticEntry{},
				},
				{
					DescriptorID:  2,
					OuterAddress:  "pp1",
					InnerNetwork:  "192.168.2.0-192.168.2.255",
					StaticEntries: []MasqueradeStaticEntry{},
				},
			},
			wantErr: false,
		},
		{
			name:     "empty input",
			input:    "",
			expected: []NATMasquerade{},
			wantErr:  false,
		},
		{
			name:     "unrelated config lines",
			input:    "ip route default gateway 192.168.1.1\ndhcp scope 1 192.168.1.0/24",
			expected: []NATMasquerade{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseNATMasqueradeConfig(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d descriptors, got %d", len(tt.expected), len(result))
				return
			}

			// Create maps for comparison since order may not be guaranteed
			resultMap := make(map[int]NATMasquerade)
			for _, r := range result {
				resultMap[r.DescriptorID] = r
			}

			for _, exp := range tt.expected {
				got, ok := resultMap[exp.DescriptorID]
				if !ok {
					t.Errorf("descriptor %d not found in result", exp.DescriptorID)
					continue
				}

				if got.OuterAddress != exp.OuterAddress {
					t.Errorf("descriptor %d: outer address = %q, want %q", exp.DescriptorID, got.OuterAddress, exp.OuterAddress)
				}
				if got.InnerNetwork != exp.InnerNetwork {
					t.Errorf("descriptor %d: inner network = %q, want %q", exp.DescriptorID, got.InnerNetwork, exp.InnerNetwork)
				}
				if len(got.StaticEntries) != len(exp.StaticEntries) {
					t.Errorf("descriptor %d: static entries count = %d, want %d", exp.DescriptorID, len(got.StaticEntries), len(exp.StaticEntries))
					continue
				}

				for i, expEntry := range exp.StaticEntries {
					gotEntry := got.StaticEntries[i]
					if gotEntry.EntryNumber != expEntry.EntryNumber {
						t.Errorf("descriptor %d, entry %d: entry number = %d, want %d", exp.DescriptorID, i, gotEntry.EntryNumber, expEntry.EntryNumber)
					}
					if gotEntry.OutsideGlobal != expEntry.OutsideGlobal {
						t.Errorf("descriptor %d, entry %d: outside global = %q, want %q", exp.DescriptorID, i, gotEntry.OutsideGlobal, expEntry.OutsideGlobal)
					}
					// Compare port pointers
					if (gotEntry.OutsideGlobalPort == nil) != (expEntry.OutsideGlobalPort == nil) {
						t.Errorf("descriptor %d, entry %d: outside port nil mismatch, got nil=%v, want nil=%v", exp.DescriptorID, i, gotEntry.OutsideGlobalPort == nil, expEntry.OutsideGlobalPort == nil)
					} else if gotEntry.OutsideGlobalPort != nil && *gotEntry.OutsideGlobalPort != *expEntry.OutsideGlobalPort {
						t.Errorf("descriptor %d, entry %d: outside port = %d, want %d", exp.DescriptorID, i, *gotEntry.OutsideGlobalPort, *expEntry.OutsideGlobalPort)
					}
					if gotEntry.InsideLocal != expEntry.InsideLocal {
						t.Errorf("descriptor %d, entry %d: inside local = %q, want %q", exp.DescriptorID, i, gotEntry.InsideLocal, expEntry.InsideLocal)
					}
					// Compare port pointers
					if (gotEntry.InsideLocalPort == nil) != (expEntry.InsideLocalPort == nil) {
						t.Errorf("descriptor %d, entry %d: inside port nil mismatch, got nil=%v, want nil=%v", exp.DescriptorID, i, gotEntry.InsideLocalPort == nil, expEntry.InsideLocalPort == nil)
					} else if gotEntry.InsideLocalPort != nil && *gotEntry.InsideLocalPort != *expEntry.InsideLocalPort {
						t.Errorf("descriptor %d, entry %d: inside port = %d, want %d", exp.DescriptorID, i, *gotEntry.InsideLocalPort, *expEntry.InsideLocalPort)
					}
					if gotEntry.Protocol != expEntry.Protocol {
						t.Errorf("descriptor %d, entry %d: protocol = %q, want %q", exp.DescriptorID, i, gotEntry.Protocol, expEntry.Protocol)
					}
				}
			}
		})
	}
}

func TestBuildNATDescriptorTypeMasqueradeCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		expected string
	}{
		{
			name:     "descriptor 1",
			id:       1,
			expected: "nat descriptor type 1 masquerade",
		},
		{
			name:     "descriptor 100",
			id:       100,
			expected: "nat descriptor type 100 masquerade",
		},
		{
			name:     "max descriptor",
			id:       65535,
			expected: "nat descriptor type 65535 masquerade",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNATDescriptorTypeMasqueradeCommand(tt.id)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildNATDescriptorAddressOuterCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		address  string
		expected string
	}{
		{
			name:     "ipcp address",
			id:       1,
			address:  "ipcp",
			expected: "nat descriptor address outer 1 ipcp",
		},
		{
			name:     "specific IP",
			id:       2,
			address:  "203.0.113.1",
			expected: "nat descriptor address outer 2 203.0.113.1",
		},
		{
			name:     "interface name",
			id:       3,
			address:  "pp1",
			expected: "nat descriptor address outer 3 pp1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNATDescriptorAddressOuterCommand(tt.id, tt.address)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildNATDescriptorAddressInnerCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		network  string
		expected string
	}{
		{
			name:     "basic range",
			id:       1,
			network:  "192.168.1.0-192.168.1.255",
			expected: "nat descriptor address inner 1 192.168.1.0-192.168.1.255",
		},
		{
			name:     "smaller range",
			id:       2,
			network:  "10.0.0.1-10.0.0.10",
			expected: "nat descriptor address inner 2 10.0.0.1-10.0.0.10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNATDescriptorAddressInnerCommand(tt.id, tt.network)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildNATMasqueradeStaticCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		entryNum int
		entry    MasqueradeStaticEntry
		expected string
	}{
		{
			name:     "tcp with specific IP",
			id:       1,
			entryNum: 1,
			entry: MasqueradeStaticEntry{
				OutsideGlobal:     "203.0.113.1",
				OutsideGlobalPort: intPtr(80),
				InsideLocal:       "192.168.1.100",
				InsideLocalPort:   intPtr(8080),
				Protocol:          "tcp",
			},
			expected: "nat descriptor masquerade static 1 1 203.0.113.1:80=192.168.1.100:8080 tcp",
		},
		{
			name:     "udp with ipcp same ports",
			id:       2,
			entryNum: 3,
			entry: MasqueradeStaticEntry{
				OutsideGlobal:     "ipcp",
				OutsideGlobalPort: intPtr(53),
				InsideLocal:       "10.0.0.1",
				InsideLocalPort:   intPtr(53),
				Protocol:          "udp",
			},
			expected: "nat descriptor masquerade static 2 3 10.0.0.1 udp 53",
		},
		{
			name:     "tcp with ipcp different ports",
			id:       1000,
			entryNum: 4,
			entry: MasqueradeStaticEntry{
				OutsideGlobal:     "ipcp",
				OutsideGlobalPort: intPtr(8080),
				InsideLocal:       "192.168.1.100",
				InsideLocalPort:   intPtr(80),
				Protocol:          "tcp",
			},
			expected: "nat descriptor masquerade static 1000 4 192.168.1.100 tcp 8080=80",
		},
		{
			name:     "without protocol",
			id:       1,
			entryNum: 2,
			entry: MasqueradeStaticEntry{
				OutsideGlobal:     "203.0.113.1",
				OutsideGlobalPort: intPtr(22),
				InsideLocal:       "192.168.1.50",
				InsideLocalPort:   intPtr(22),
				Protocol:          "",
			},
			expected: "nat descriptor masquerade static 1 2 203.0.113.1:22=192.168.1.50:22",
		},
		{
			name:     "uppercase protocol normalized",
			id:       1,
			entryNum: 1,
			entry: MasqueradeStaticEntry{
				OutsideGlobal:     "203.0.113.1",
				OutsideGlobalPort: intPtr(443),
				InsideLocal:       "192.168.1.100",
				InsideLocalPort:   intPtr(443),
				Protocol:          "TCP",
			},
			expected: "nat descriptor masquerade static 1 1 203.0.113.1:443=192.168.1.100:443 tcp",
		},
		{
			name:     "protocol-only ESP",
			id:       1000,
			entryNum: 1,
			entry: MasqueradeStaticEntry{
				InsideLocal: "192.168.1.253",
				Protocol:    "esp",
			},
			expected: "nat descriptor masquerade static 1000 1 192.168.1.253 esp",
		},
		{
			name:     "protocol-only AH",
			id:       1000,
			entryNum: 2,
			entry: MasqueradeStaticEntry{
				InsideLocal: "192.168.1.254",
				Protocol:    "ah",
			},
			expected: "nat descriptor masquerade static 1000 2 192.168.1.254 ah",
		},
		{
			name:     "protocol-only GRE",
			id:       2000,
			entryNum: 1,
			entry: MasqueradeStaticEntry{
				InsideLocal: "10.0.0.100",
				Protocol:    "gre",
			},
			expected: "nat descriptor masquerade static 2000 1 10.0.0.100 gre",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNATMasqueradeStaticCommand(tt.id, tt.entryNum, tt.entry)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteNATMasqueradeCommand(t *testing.T) {
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
			result := BuildDeleteNATMasqueradeCommand(tt.id)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildInterfaceNATDescriptorCommand(t *testing.T) {
	tests := []struct {
		name         string
		iface        string
		descriptorID int
		expected     string
	}{
		{
			name:         "pp1 interface",
			iface:        "pp1",
			descriptorID: 1,
			expected:     "ip pp1 nat descriptor 1",
		},
		{
			name:         "lan2 interface",
			iface:        "lan2",
			descriptorID: 2,
			expected:     "ip lan2 nat descriptor 2",
		},
		{
			name:         "tunnel1 interface",
			iface:        "tunnel1",
			descriptorID: 100,
			expected:     "ip tunnel1 nat descriptor 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildInterfaceNATDescriptorCommand(tt.iface, tt.descriptorID)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValidateMasqueradeDescriptorID(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		wantErr bool
	}{
		{
			name:    "valid minimum",
			id:      1,
			wantErr: false,
		},
		{
			name:    "valid maximum",
			id:      65535,
			wantErr: false,
		},
		{
			name:    "valid middle",
			id:      1000,
			wantErr: false,
		},
		{
			name:    "zero is invalid",
			id:      0,
			wantErr: true,
		},
		{
			name:    "negative is invalid",
			id:      -1,
			wantErr: true,
		},
		{
			name:    "exceeds maximum",
			id:      65536,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDescriptorID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDescriptorID(%d) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCIDR(t *testing.T) {
	tests := []struct {
		name    string
		cidr    string
		wantErr bool
	}{
		{
			name:    "valid /24",
			cidr:    "192.168.1.0/24",
			wantErr: false,
		},
		{
			name:    "valid /16",
			cidr:    "10.0.0.0/16",
			wantErr: false,
		},
		{
			name:    "valid /32",
			cidr:    "192.168.1.1/32",
			wantErr: false,
		},
		{
			name:    "valid /8",
			cidr:    "10.0.0.0/8",
			wantErr: false,
		},
		{
			name:    "invalid no prefix",
			cidr:    "192.168.1.0",
			wantErr: true,
		},
		{
			name:    "invalid prefix too large",
			cidr:    "192.168.1.0/33",
			wantErr: true,
		},
		{
			name:    "invalid IP",
			cidr:    "999.999.999.999/24",
			wantErr: true,
		},
		{
			name:    "empty string",
			cidr:    "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			cidr:    "not-a-cidr",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCIDR(tt.cidr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCIDR(%q) error = %v, wantErr %v", tt.cidr, err, tt.wantErr)
			}
		})
	}
}

func TestConvertCIDRToRange(t *testing.T) {
	tests := []struct {
		name      string
		cidr      string
		wantStart string
		wantEnd   string
		wantErr   bool
	}{
		{
			name:      "class C /24",
			cidr:      "192.168.1.0/24",
			wantStart: "192.168.1.0",
			wantEnd:   "192.168.1.255",
			wantErr:   false,
		},
		{
			name:      "class B /16",
			cidr:      "172.16.0.0/16",
			wantStart: "172.16.0.0",
			wantEnd:   "172.16.255.255",
			wantErr:   false,
		},
		{
			name:      "class A /8",
			cidr:      "10.0.0.0/8",
			wantStart: "10.0.0.0",
			wantEnd:   "10.255.255.255",
			wantErr:   false,
		},
		{
			name:      "single host /32",
			cidr:      "192.168.1.100/32",
			wantStart: "192.168.1.100",
			wantEnd:   "192.168.1.100",
			wantErr:   false,
		},
		{
			name:      "/25 subnet",
			cidr:      "192.168.1.0/25",
			wantStart: "192.168.1.0",
			wantEnd:   "192.168.1.127",
			wantErr:   false,
		},
		{
			name:      "/30 subnet",
			cidr:      "192.168.1.0/30",
			wantStart: "192.168.1.0",
			wantEnd:   "192.168.1.3",
			wantErr:   false,
		},
		{
			name:    "invalid CIDR",
			cidr:    "not-a-cidr",
			wantErr: true,
		},
		{
			name:    "empty string",
			cidr:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := ConvertCIDRToRange(tt.cidr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if start != tt.wantStart {
				t.Errorf("start = %q, want %q", start, tt.wantStart)
			}
			if end != tt.wantEnd {
				t.Errorf("end = %q, want %q", end, tt.wantEnd)
			}
		})
	}
}

func TestConvertRangeToRTXFormat(t *testing.T) {
	tests := []struct {
		name     string
		cidr     string
		expected string
		wantErr  bool
	}{
		{
			name:     "/24 to range",
			cidr:     "192.168.1.0/24",
			expected: "192.168.1.0-192.168.1.255",
			wantErr:  false,
		},
		{
			name:     "/16 to range",
			cidr:     "10.0.0.0/16",
			expected: "10.0.0.0-10.0.255.255",
			wantErr:  false,
		},
		{
			name:    "invalid CIDR",
			cidr:    "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertRangeToRTXFormat(tt.cidr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValidateNATPort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{
			name:    "valid minimum",
			port:    1,
			wantErr: false,
		},
		{
			name:    "valid maximum",
			port:    65535,
			wantErr: false,
		},
		{
			name:    "valid common port",
			port:    80,
			wantErr: false,
		},
		{
			name:    "zero is invalid",
			port:    0,
			wantErr: true,
		},
		{
			name:    "negative is invalid",
			port:    -1,
			wantErr: true,
		},
		{
			name:    "exceeds maximum",
			port:    65536,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNATPort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNATPort(%d) error = %v, wantErr %v", tt.port, err, tt.wantErr)
			}
		})
	}
}

func TestValidateNATProtocol(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		wantErr  bool
	}{
		{
			name:     "tcp lowercase",
			protocol: "tcp",
			wantErr:  false,
		},
		{
			name:     "udp lowercase",
			protocol: "udp",
			wantErr:  false,
		},
		{
			name:     "TCP uppercase",
			protocol: "TCP",
			wantErr:  false,
		},
		{
			name:     "UDP uppercase",
			protocol: "UDP",
			wantErr:  false,
		},
		{
			name:     "empty is valid",
			protocol: "",
			wantErr:  false,
		},
		{
			name:     "icmp is valid (protocol-only NAT)",
			protocol: "icmp",
			wantErr:  false,
		},
		{
			name:     "esp is valid (protocol-only NAT for IPsec)",
			protocol: "esp",
			wantErr:  false,
		},
		{
			name:     "ah is valid (protocol-only NAT for IPsec)",
			protocol: "ah",
			wantErr:  false,
		},
		{
			name:     "gre is valid (protocol-only NAT for tunnels)",
			protocol: "gre",
			wantErr:  false,
		},
		{
			name:     "random string is invalid",
			protocol: "http",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNATProtocol(tt.protocol)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNATProtocol(%q) error = %v, wantErr %v", tt.protocol, err, tt.wantErr)
			}
		})
	}
}

// TestIsProtocolOnly tests the protocol-only helper function for ESP, AH, GRE, ICMP protocols
func TestIsProtocolOnly(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		expected bool
	}{
		{name: "esp is protocol-only", protocol: "esp", expected: true},
		{name: "ESP uppercase is protocol-only", protocol: "ESP", expected: true},
		{name: "ah is protocol-only", protocol: "ah", expected: true},
		{name: "AH uppercase is protocol-only", protocol: "AH", expected: true},
		{name: "gre is protocol-only", protocol: "gre", expected: true},
		{name: "GRE uppercase is protocol-only", protocol: "GRE", expected: true},
		{name: "icmp is protocol-only", protocol: "icmp", expected: true},
		{name: "ICMP uppercase is protocol-only", protocol: "ICMP", expected: true},
		{name: "tcp is not protocol-only", protocol: "tcp", expected: false},
		{name: "udp is not protocol-only", protocol: "udp", expected: false},
		{name: "empty is not protocol-only", protocol: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsProtocolOnly(tt.protocol)
			if result != tt.expected {
				t.Errorf("IsProtocolOnly(%q) = %v, want %v", tt.protocol, result, tt.expected)
			}
		})
	}
}

// TestBuildNATMasqueradeStaticCommand_ProtocolOnly tests command building for protocol-only entries (ESP, AH, GRE)
func TestBuildNATMasqueradeStaticCommand_ProtocolOnly(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		entryNum int
		entry    MasqueradeStaticEntry
		expected string
	}{
		{
			name:     "ESP protocol-only entry",
			id:       1000,
			entryNum: 1,
			entry: MasqueradeStaticEntry{
				InsideLocal: "192.168.1.253",
				Protocol:    "esp",
			},
			expected: "nat descriptor masquerade static 1000 1 192.168.1.253 esp",
		},
		{
			name:     "AH protocol-only entry",
			id:       1000,
			entryNum: 2,
			entry: MasqueradeStaticEntry{
				InsideLocal: "192.168.1.253",
				Protocol:    "ah",
			},
			expected: "nat descriptor masquerade static 1000 2 192.168.1.253 ah",
		},
		{
			name:     "GRE protocol-only entry",
			id:       1000,
			entryNum: 3,
			entry: MasqueradeStaticEntry{
				InsideLocal: "10.0.0.100",
				Protocol:    "gre",
			},
			expected: "nat descriptor masquerade static 1000 3 10.0.0.100 gre",
		},
		{
			name:     "ICMP protocol-only entry",
			id:       2000,
			entryNum: 1,
			entry: MasqueradeStaticEntry{
				InsideLocal: "192.168.0.1",
				Protocol:    "icmp",
			},
			expected: "nat descriptor masquerade static 2000 1 192.168.0.1 icmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNATMasqueradeStaticCommand(tt.id, tt.entryNum, tt.entry)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestParseNATMasqueradeConfig_ProtocolOnly tests parsing of protocol-only static entries
func TestParseNATMasqueradeConfig_ProtocolOnly(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []NATMasquerade
		wantErr  bool
	}{
		{
			name: "ESP protocol-only entry",
			input: `nat descriptor type 1000 masquerade
nat descriptor masquerade static 1000 1 192.168.1.253 esp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1000,
					StaticEntries: []MasqueradeStaticEntry{
						{
							EntryNumber: 1,
							InsideLocal: "192.168.1.253",
							Protocol:    "esp",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple protocol-only entries",
			input: `nat descriptor type 1000 masquerade
nat descriptor masquerade static 1000 1 192.168.1.253 esp
nat descriptor masquerade static 1000 2 192.168.1.253 ah
nat descriptor masquerade static 1000 3 10.0.0.100 gre`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1000,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, InsideLocal: "192.168.1.253", Protocol: "esp"},
						{EntryNumber: 2, InsideLocal: "192.168.1.253", Protocol: "ah"},
						{EntryNumber: 3, InsideLocal: "10.0.0.100", Protocol: "gre"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Mixed protocol-only and port-based entries",
			input: `nat descriptor type 1000 masquerade
nat descriptor masquerade static 1000 1 ipcp:80=192.168.1.100:80 tcp
nat descriptor masquerade static 1000 2 192.168.1.253 esp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1000,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(80), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(80), Protocol: "tcp"},
						{EntryNumber: 2, InsideLocal: "192.168.1.253", Protocol: "esp"},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseNATMasqueradeConfig(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d descriptors, got %d", len(tt.expected), len(result))
				return
			}

			resultMap := make(map[int]NATMasquerade)
			for _, r := range result {
				resultMap[r.DescriptorID] = r
			}

			for _, exp := range tt.expected {
				got, ok := resultMap[exp.DescriptorID]
				if !ok {
					t.Errorf("descriptor %d not found in result", exp.DescriptorID)
					continue
				}

				if len(got.StaticEntries) != len(exp.StaticEntries) {
					t.Errorf("descriptor %d: static entries count = %d, want %d", exp.DescriptorID, len(got.StaticEntries), len(exp.StaticEntries))
					continue
				}

				for i, expEntry := range exp.StaticEntries {
					gotEntry := got.StaticEntries[i]
					if gotEntry.EntryNumber != expEntry.EntryNumber {
						t.Errorf("descriptor %d, entry %d: entry number = %d, want %d", exp.DescriptorID, i, gotEntry.EntryNumber, expEntry.EntryNumber)
					}
					if gotEntry.InsideLocal != expEntry.InsideLocal {
						t.Errorf("descriptor %d, entry %d: inside local = %q, want %q", exp.DescriptorID, i, gotEntry.InsideLocal, expEntry.InsideLocal)
					}
					if gotEntry.Protocol != expEntry.Protocol {
						t.Errorf("descriptor %d, entry %d: protocol = %q, want %q", exp.DescriptorID, i, gotEntry.Protocol, expEntry.Protocol)
					}
					// Compare port pointers
					if (gotEntry.InsideLocalPort == nil) != (expEntry.InsideLocalPort == nil) {
						t.Errorf("descriptor %d, entry %d: inside port nil mismatch", exp.DescriptorID, i)
					} else if gotEntry.InsideLocalPort != nil && *gotEntry.InsideLocalPort != *expEntry.InsideLocalPort {
						t.Errorf("descriptor %d, entry %d: inside port = %d, want %d", exp.DescriptorID, i, *gotEntry.InsideLocalPort, *expEntry.InsideLocalPort)
					}
					if (gotEntry.OutsideGlobalPort == nil) != (expEntry.OutsideGlobalPort == nil) {
						t.Errorf("descriptor %d, entry %d: outside port nil mismatch", exp.DescriptorID, i)
					} else if gotEntry.OutsideGlobalPort != nil && *gotEntry.OutsideGlobalPort != *expEntry.OutsideGlobalPort {
						t.Errorf("descriptor %d, entry %d: outside port = %d, want %d", exp.DescriptorID, i, *gotEntry.OutsideGlobalPort, *expEntry.OutsideGlobalPort)
					}
				}
			}
		})
	}
}

func TestValidateOuterAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "ipcp",
			address: "ipcp",
			wantErr: false,
		},
		{
			name:    "primary",
			address: "primary",
			wantErr: false,
		},
		{
			name:    "secondary",
			address: "secondary",
			wantErr: false,
		},
		{
			name:    "pp1 interface",
			address: "pp1",
			wantErr: false,
		},
		{
			name:    "lan1 interface",
			address: "lan1",
			wantErr: false,
		},
		{
			name:    "tunnel1 interface",
			address: "tunnel1",
			wantErr: false,
		},
		{
			name:    "valid IP address",
			address: "203.0.113.1",
			wantErr: false,
		},
		{
			name:    "valid private IP",
			address: "192.168.1.1",
			wantErr: false,
		},
		{
			name:    "empty is invalid",
			address: "",
			wantErr: true,
		},
		{
			name:    "invalid address",
			address: "not-valid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOuterAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOuterAddress(%q) error = %v, wantErr %v", tt.address, err, tt.wantErr)
			}
		})
	}
}

func TestValidateNATMasquerade(t *testing.T) {
	tests := []struct {
		name    string
		nat     NATMasquerade
		wantErr bool
	}{
		{
			name: "valid basic config",
			nat: NATMasquerade{
				DescriptorID:  1,
				OuterAddress:  "ipcp",
				InnerNetwork:  "192.168.1.0-192.168.1.255",
				StaticEntries: []MasqueradeStaticEntry{},
			},
			wantErr: false,
		},
		{
			name: "valid with static entries",
			nat: NATMasquerade{
				DescriptorID: 2,
				OuterAddress: "203.0.113.1",
				InnerNetwork: "192.168.1.0-192.168.1.255",
				StaticEntries: []MasqueradeStaticEntry{
					{
						EntryNumber:       1,
						OutsideGlobal:     "203.0.113.1",
						OutsideGlobalPort: intPtr(80),
						InsideLocal:       "192.168.1.100",
						InsideLocalPort:   intPtr(8080),
						Protocol:          "tcp",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid descriptor ID zero",
			nat: NATMasquerade{
				DescriptorID: 0,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.1.0-192.168.1.255",
			},
			wantErr: true,
		},
		{
			name: "invalid descriptor ID too large",
			nat: NATMasquerade{
				DescriptorID: 70000,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.1.0-192.168.1.255",
			},
			wantErr: true,
		},
		{
			name: "empty outer address",
			nat: NATMasquerade{
				DescriptorID: 1,
				OuterAddress: "",
				InnerNetwork: "192.168.1.0-192.168.1.255",
			},
			wantErr: true,
		},
		{
			name: "empty inner network",
			nat: NATMasquerade{
				DescriptorID: 1,
				OuterAddress: "ipcp",
				InnerNetwork: "",
			},
			wantErr: true,
		},
		{
			name: "invalid static entry port - zero",
			nat: NATMasquerade{
				DescriptorID: 1,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.1.0-192.168.1.255",
				StaticEntries: []MasqueradeStaticEntry{
					{
						EntryNumber:       1,
						OutsideGlobal:     "203.0.113.1",
						OutsideGlobalPort: intPtr(0),
						InsideLocal:       "192.168.1.100",
						InsideLocalPort:   intPtr(8080),
						Protocol:          "tcp",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid static entry inside IP",
			nat: NATMasquerade{
				DescriptorID: 1,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.1.0-192.168.1.255",
				StaticEntries: []MasqueradeStaticEntry{
					{
						EntryNumber:       1,
						OutsideGlobal:     "203.0.113.1",
						OutsideGlobalPort: intPtr(80),
						InsideLocal:       "not-an-ip",
						InsideLocalPort:   intPtr(8080),
						Protocol:          "tcp",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid static entry protocol",
			nat: NATMasquerade{
				DescriptorID: 1,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.1.0-192.168.1.255",
				StaticEntries: []MasqueradeStaticEntry{
					{
						EntryNumber:       1,
						OutsideGlobal:     "203.0.113.1",
						OutsideGlobalPort: intPtr(80),
						InsideLocal:       "192.168.1.100",
						InsideLocalPort:   intPtr(8080),
						Protocol:          "invalid",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid protocol-only ESP entry",
			nat: NATMasquerade{
				DescriptorID: 1000,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.1.0-192.168.1.255",
				StaticEntries: []MasqueradeStaticEntry{
					{
						EntryNumber: 1,
						InsideLocal: "192.168.1.253",
						Protocol:    "esp",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "protocol-only entry should not have ports",
			nat: NATMasquerade{
				DescriptorID: 1000,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.1.0-192.168.1.255",
				StaticEntries: []MasqueradeStaticEntry{
					{
						EntryNumber:       1,
						InsideLocal:       "192.168.1.253",
						InsideLocalPort:   intPtr(80),
						OutsideGlobalPort: intPtr(80),
						Protocol:          "esp",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNATMasquerade(tt.nat)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNATMasquerade() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildDeleteInterfaceNATDescriptorCommand(t *testing.T) {
	tests := []struct {
		name         string
		iface        string
		descriptorID int
		expected     string
	}{
		{
			name:         "pp1 interface",
			iface:        "pp1",
			descriptorID: 1,
			expected:     "no ip pp1 nat descriptor 1",
		},
		{
			name:         "lan2 interface",
			iface:        "lan2",
			descriptorID: 2,
			expected:     "no ip lan2 nat descriptor 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteInterfaceNATDescriptorCommand(tt.iface, tt.descriptorID)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteNATMasqueradeStaticCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		entryNum int
		expected string
	}{
		{
			name:     "delete entry 1",
			id:       1,
			entryNum: 1,
			expected: "no nat descriptor masquerade static 1 1",
		},
		{
			name:     "delete entry 5 from descriptor 10",
			id:       10,
			entryNum: 5,
			expected: "no nat descriptor masquerade static 10 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteNATMasqueradeStaticCommand(tt.id, tt.entryNum)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildShowNATDescriptorCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		expected string
	}{
		{
			name:     "descriptor 1",
			id:       1,
			expected: `show config | grep "nat descriptor.*1"`,
		},
		{
			name:     "descriptor 10",
			id:       10,
			expected: `show config | grep "nat descriptor.*10"`,
		},
		{
			name:     "descriptor 100",
			id:       100,
			expected: `show config | grep "nat descriptor.*100"`,
		},
		{
			name:     "descriptor 1000 - RTX compatible pattern without grep -E",
			id:       1000,
			expected: `show config | grep "nat descriptor.*1000"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildShowNATDescriptorCommand(tt.id)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildShowAllNATDescriptorsCommand(t *testing.T) {
	result := BuildShowAllNATDescriptorsCommand()
	expected := `show config | grep "nat descriptor"`
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

// =============================================================================
// Additional test cases from nat.yaml patterns
// =============================================================================

// TestParseNATMasqueradeConfig_YAMLPatterns tests additional patterns from nat.yaml
func TestParseNATMasqueradeConfig_YAMLPatterns(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []NATMasquerade
		wantErr  bool
	}{
		{
			name: "outer address with primary keyword",
			input: `nat descriptor type 1 masquerade
nat descriptor address outer 1 primary
nat descriptor address inner 1 auto`,
			expected: []NATMasquerade{
				{
					DescriptorID:  1,
					OuterAddress:  "primary",
					InnerNetwork:  "auto",
					StaticEntries: []MasqueradeStaticEntry{},
				},
			},
			wantErr: false,
		},
		{
			name: "outer address with IP range",
			input: `nat descriptor type 2 masquerade
nat descriptor address outer 2 203.0.113.1-203.0.113.16
nat descriptor address inner 2 192.168.0.1-192.168.0.100`,
			expected: []NATMasquerade{
				{
					DescriptorID:  2,
					OuterAddress:  "203.0.113.1-203.0.113.16",
					InnerNetwork:  "192.168.0.1-192.168.0.100",
					StaticEntries: []MasqueradeStaticEntry{},
				},
			},
			wantErr: false,
		},
		{
			name: "outer address with specific IP",
			input: `nat descriptor type 1 masquerade
nat descriptor address outer 1 203.0.113.1
nat descriptor address inner 1 192.168.1.0-192.168.1.255`,
			expected: []NATMasquerade{
				{
					DescriptorID:  1,
					OuterAddress:  "203.0.113.1",
					InnerNetwork:  "192.168.1.0-192.168.1.255",
					StaticEntries: []MasqueradeStaticEntry{},
				},
			},
			wantErr: false,
		},
		{
			name: "inner address with auto keyword",
			input: `nat descriptor type 1 masquerade
nat descriptor address outer 1 ipcp
nat descriptor address inner 1 auto`,
			expected: []NATMasquerade{
				{
					DescriptorID:  1,
					OuterAddress:  "ipcp",
					InnerNetwork:  "auto",
					StaticEntries: []MasqueradeStaticEntry{},
				},
			},
			wantErr: false,
		},
		{
			name: "masquerade static with UDP DNS forwarding",
			input: `nat descriptor type 1 masquerade
nat descriptor address outer 1 ipcp
nat descriptor address inner 1 192.168.1.0-192.168.1.255
nat descriptor masquerade static 1 3 ipcp:53=192.168.1.200:53 udp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					OuterAddress: "ipcp",
					InnerNetwork: "192.168.1.0-192.168.1.255",
					StaticEntries: []MasqueradeStaticEntry{
						{
							EntryNumber:       3,
							OutsideGlobal:     "ipcp",
							OutsideGlobalPort: intPtr(53),
							InsideLocal:       "192.168.1.200",
							InsideLocalPort:   intPtr(53),
							Protocol:          "udp",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "masquerade static with different outer and inner ports",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 2 203.0.113.1:8080=192.168.1.100:80 tcp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{
							EntryNumber:       2,
							OutsideGlobal:     "203.0.113.1",
							OutsideGlobalPort: intPtr(8080),
							InsideLocal:       "192.168.1.100",
							InsideLocalPort:   intPtr(80),
							Protocol:          "tcp",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "large descriptor ID",
			input: `nat descriptor type 65535 masquerade
nat descriptor address outer 65535 ipcp
nat descriptor address inner 65535 10.0.0.0-10.0.0.255`,
			expected: []NATMasquerade{
				{
					DescriptorID:  65535,
					OuterAddress:  "ipcp",
					InnerNetwork:  "10.0.0.0-10.0.0.255",
					StaticEntries: []MasqueradeStaticEntry{},
				},
			},
			wantErr: false,
		},
		{
			name: "descriptor ID 1000 with VPN passthrough",
			input: `nat descriptor type 1000 masquerade
nat descriptor address outer 1000 primary
nat descriptor address inner 1000 192.168.1.0-192.168.1.255
nat descriptor masquerade static 1000 1 192.168.1.253 esp
nat descriptor masquerade static 1000 2 192.168.1.253 ah`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1000,
					OuterAddress: "primary",
					InnerNetwork: "192.168.1.0-192.168.1.255",
					StaticEntries: []MasqueradeStaticEntry{
						{
							EntryNumber:   1,
							InsideLocal:   "192.168.1.253",
							Protocol:      "esp",
							OutsideGlobal: "ipcp",
						},
						{
							EntryNumber:   2,
							InsideLocal:   "192.168.1.253",
							Protocol:      "ah",
							OutsideGlobal: "ipcp",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "descriptor 1000 with multiple mixed static entries",
			input: `nat descriptor type 1000 masquerade
nat descriptor address outer 1000 ipcp
nat descriptor address inner 1000 10.0.0.0-10.0.255.255
nat descriptor masquerade static 1000 1 10.0.0.100 tcp 443
nat descriptor masquerade static 1000 2 10.0.0.100 tcp 8443=443
nat descriptor masquerade static 1000 3 10.0.0.200 esp
nat descriptor masquerade static 1000 4 10.0.0.200 gre`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1000,
					OuterAddress: "ipcp",
					InnerNetwork: "10.0.0.0-10.0.255.255",
					StaticEntries: []MasqueradeStaticEntry{
						{
							EntryNumber:       1,
							InsideLocal:       "10.0.0.100",
							InsideLocalPort:   intPtr(443),
							OutsideGlobalPort: intPtr(443),
							Protocol:          "tcp",
							OutsideGlobal:     "ipcp",
						},
						{
							EntryNumber:       2,
							InsideLocal:       "10.0.0.100",
							InsideLocalPort:   intPtr(443),
							OutsideGlobalPort: intPtr(8443),
							Protocol:          "tcp",
							OutsideGlobal:     "ipcp",
						},
						{
							EntryNumber:   3,
							InsideLocal:   "10.0.0.200",
							Protocol:      "esp",
							OutsideGlobal: "ipcp",
						},
						{
							EntryNumber:   4,
							InsideLocal:   "10.0.0.200",
							Protocol:      "gre",
							OutsideGlobal: "ipcp",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "protocol-only entries icmp",
			input: `nat descriptor type 500 masquerade
nat descriptor address outer 500 pp1
nat descriptor address inner 500 172.16.0.0-172.16.0.255
nat descriptor masquerade static 500 1 172.16.0.1 icmp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 500,
					OuterAddress: "pp1",
					InnerNetwork: "172.16.0.0-172.16.0.255",
					StaticEntries: []MasqueradeStaticEntry{
						{
							EntryNumber:   1,
							InsideLocal:   "172.16.0.1",
							Protocol:      "icmp",
							OutsideGlobal: "ipcp",
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseNATMasqueradeConfig(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d descriptors, got %d", len(tt.expected), len(result))
				return
			}

			resultMap := make(map[int]NATMasquerade)
			for _, r := range result {
				resultMap[r.DescriptorID] = r
			}

			for _, exp := range tt.expected {
				got, ok := resultMap[exp.DescriptorID]
				if !ok {
					t.Errorf("descriptor %d not found in result", exp.DescriptorID)
					continue
				}

				if got.OuterAddress != exp.OuterAddress {
					t.Errorf("descriptor %d: outer address = %q, want %q", exp.DescriptorID, got.OuterAddress, exp.OuterAddress)
				}
				if got.InnerNetwork != exp.InnerNetwork {
					t.Errorf("descriptor %d: inner network = %q, want %q", exp.DescriptorID, got.InnerNetwork, exp.InnerNetwork)
				}
				if len(got.StaticEntries) != len(exp.StaticEntries) {
					t.Errorf("descriptor %d: static entries count = %d, want %d", exp.DescriptorID, len(got.StaticEntries), len(exp.StaticEntries))
					continue
				}

				for i, expEntry := range exp.StaticEntries {
					gotEntry := got.StaticEntries[i]
					if gotEntry.EntryNumber != expEntry.EntryNumber {
						t.Errorf("descriptor %d, entry %d: entry number = %d, want %d", exp.DescriptorID, i, gotEntry.EntryNumber, expEntry.EntryNumber)
					}
					if gotEntry.OutsideGlobal != expEntry.OutsideGlobal {
						t.Errorf("descriptor %d, entry %d: outside global = %q, want %q", exp.DescriptorID, i, gotEntry.OutsideGlobal, expEntry.OutsideGlobal)
					}
					// Compare port pointers
					if (gotEntry.OutsideGlobalPort == nil) != (expEntry.OutsideGlobalPort == nil) {
						t.Errorf("descriptor %d, entry %d: outside port nil mismatch, got nil=%v, want nil=%v", exp.DescriptorID, i, gotEntry.OutsideGlobalPort == nil, expEntry.OutsideGlobalPort == nil)
					} else if gotEntry.OutsideGlobalPort != nil && *gotEntry.OutsideGlobalPort != *expEntry.OutsideGlobalPort {
						t.Errorf("descriptor %d, entry %d: outside port = %d, want %d", exp.DescriptorID, i, *gotEntry.OutsideGlobalPort, *expEntry.OutsideGlobalPort)
					}
					if gotEntry.InsideLocal != expEntry.InsideLocal {
						t.Errorf("descriptor %d, entry %d: inside local = %q, want %q", exp.DescriptorID, i, gotEntry.InsideLocal, expEntry.InsideLocal)
					}
					// Compare port pointers
					if (gotEntry.InsideLocalPort == nil) != (expEntry.InsideLocalPort == nil) {
						t.Errorf("descriptor %d, entry %d: inside port nil mismatch, got nil=%v, want nil=%v", exp.DescriptorID, i, gotEntry.InsideLocalPort == nil, expEntry.InsideLocalPort == nil)
					} else if gotEntry.InsideLocalPort != nil && *gotEntry.InsideLocalPort != *expEntry.InsideLocalPort {
						t.Errorf("descriptor %d, entry %d: inside port = %d, want %d", exp.DescriptorID, i, *gotEntry.InsideLocalPort, *expEntry.InsideLocalPort)
					}
					if gotEntry.Protocol != expEntry.Protocol {
						t.Errorf("descriptor %d, entry %d: protocol = %q, want %q", exp.DescriptorID, i, gotEntry.Protocol, expEntry.Protocol)
					}
				}
			}
		})
	}
}

// TestValidateOuterAddress_YAMLPatterns tests outer address validation with nat.yaml patterns
func TestValidateOuterAddress_YAMLPatterns(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		// Valid cases from nat.yaml
		{name: "ipcp keyword", address: "ipcp", wantErr: false},
		{name: "primary keyword", address: "primary", wantErr: false},     // primary is a valid RTX outer address value
		{name: "secondary keyword", address: "secondary", wantErr: false}, // secondary is a valid RTX outer address value
		{name: "pp1 interface", address: "pp1", wantErr: false},
		{name: "pp10 interface", address: "pp10", wantErr: false},
		{name: "lan1 interface", address: "lan1", wantErr: false},
		{name: "lan2 interface", address: "lan2", wantErr: false},
		{name: "tunnel1 interface", address: "tunnel1", wantErr: false},
		{name: "specific IP", address: "203.0.113.1", wantErr: false},
		{name: "private IP", address: "192.168.1.1", wantErr: false},
		{name: "loopback IP", address: "127.0.0.1", wantErr: false},
		// Invalid cases
		{name: "empty string", address: "", wantErr: true},
		{name: "random text", address: "invalid", wantErr: true},
		{name: "IP with extra characters", address: "192.168.1.1.1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOuterAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOuterAddress(%q) error = %v, wantErr %v", tt.address, err, tt.wantErr)
			}
		})
	}
}

// TestValidateMasqueradeDescriptorID_Range tests descriptor ID range validation
func TestValidateMasqueradeDescriptorID_Range(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		wantErr bool
	}{
		// From nat.yaml: range is 1-2147483647, but implementation uses 1-65535
		{name: "minimum valid", id: 1, wantErr: false},
		{name: "typical value", id: 100, wantErr: false},
		{name: "maximum implementation limit", id: 65535, wantErr: false},
		{name: "just over maximum", id: 65536, wantErr: true},
		{name: "zero", id: 0, wantErr: true},
		{name: "negative", id: -1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDescriptorID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDescriptorID(%d) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

// TestBuildNATMasqueradeStaticCommand_YAMLPatterns tests command building with nat.yaml patterns
func TestBuildNATMasqueradeStaticCommand_YAMLPatterns(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		entryNum int
		entry    MasqueradeStaticEntry
		expected string
	}{
		{
			name:     "HTTP port forwarding same port with ipcp",
			id:       1,
			entryNum: 1,
			entry: MasqueradeStaticEntry{
				OutsideGlobal:     "ipcp",
				OutsideGlobalPort: intPtr(80),
				InsideLocal:       "192.168.1.100",
				InsideLocalPort:   intPtr(80),
				Protocol:          "tcp",
			},
			expected: "nat descriptor masquerade static 1 1 192.168.1.100 tcp 80",
		},
		{
			name:     "HTTP port forwarding different ports with ipcp",
			id:       1,
			entryNum: 2,
			entry: MasqueradeStaticEntry{
				OutsideGlobal:     "ipcp",
				OutsideGlobalPort: intPtr(8080),
				InsideLocal:       "192.168.1.100",
				InsideLocalPort:   intPtr(80),
				Protocol:          "tcp",
			},
			expected: "nat descriptor masquerade static 1 2 192.168.1.100 tcp 8080=80",
		},
		{
			name:     "DNS UDP forwarding with ipcp",
			id:       1,
			entryNum: 3,
			entry: MasqueradeStaticEntry{
				OutsideGlobal:     "ipcp",
				OutsideGlobalPort: intPtr(53),
				InsideLocal:       "192.168.1.200",
				InsideLocalPort:   intPtr(53),
				Protocol:          "udp",
			},
			expected: "nat descriptor masquerade static 1 3 192.168.1.200 udp 53",
		},
		{
			name:     "specific outer IP address",
			id:       2,
			entryNum: 1,
			entry: MasqueradeStaticEntry{
				OutsideGlobal:     "203.0.113.1",
				OutsideGlobalPort: intPtr(443),
				InsideLocal:       "192.168.1.100",
				InsideLocalPort:   intPtr(443),
				Protocol:          "tcp",
			},
			expected: "nat descriptor masquerade static 2 1 203.0.113.1:443=192.168.1.100:443 tcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNATMasqueradeStaticCommand(tt.id, tt.entryNum, tt.entry)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestBuildNATDescriptorAddressOuterCommand_YAMLPatterns tests outer address command building
func TestBuildNATDescriptorAddressOuterCommand_YAMLPatterns(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		address  string
		expected string
	}{
		// Cases from nat.yaml section 23.4
		{
			name:     "ipcp for PPPoE",
			id:       1,
			address:  "ipcp",
			expected: "nat descriptor address outer 1 ipcp",
		},
		{
			name:     "specific IP address",
			id:       1,
			address:  "203.0.113.1",
			expected: "nat descriptor address outer 1 203.0.113.1",
		},
		{
			name:     "IP address range",
			id:       2,
			address:  "203.0.113.1-203.0.113.16",
			expected: "nat descriptor address outer 2 203.0.113.1-203.0.113.16",
		},
		{
			name:     "primary keyword",
			id:       1,
			address:  "primary",
			expected: "nat descriptor address outer 1 primary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNATDescriptorAddressOuterCommand(tt.id, tt.address)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestNATMasqueradeProtocolOnlyRoundTrip tests the round-trip: parse → build → parse for protocol-only entries
func TestNATMasqueradeProtocolOnlyRoundTrip(t *testing.T) {
	tests := []struct {
		name         string
		descriptorID int
		input        string
		expected     MasqueradeStaticEntry
	}{
		{
			name:         "ESP round-trip",
			descriptorID: 1000,
			input:        "nat descriptor masquerade static 1000 1 192.168.1.253 esp",
			expected: MasqueradeStaticEntry{
				EntryNumber: 1,
				InsideLocal: "192.168.1.253",
				Protocol:    "esp",
			},
		},
		{
			name:         "AH round-trip",
			descriptorID: 1000,
			input:        "nat descriptor masquerade static 1000 2 192.168.1.253 ah",
			expected: MasqueradeStaticEntry{
				EntryNumber: 2,
				InsideLocal: "192.168.1.253",
				Protocol:    "ah",
			},
		},
		{
			name:         "GRE round-trip",
			descriptorID: 2000,
			input:        "nat descriptor masquerade static 2000 1 10.0.0.100 gre",
			expected: MasqueradeStaticEntry{
				EntryNumber: 1,
				InsideLocal: "10.0.0.100",
				Protocol:    "gre",
			},
		},
		{
			name:         "ICMP round-trip",
			descriptorID: 500,
			input:        "nat descriptor masquerade static 500 5 172.16.0.1 icmp",
			expected: MasqueradeStaticEntry{
				EntryNumber: 5,
				InsideLocal: "172.16.0.1",
				Protocol:    "icmp",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Parse the original input
			config := fmt.Sprintf("nat descriptor type %d masquerade\n%s", tt.descriptorID, tt.input)
			result, err := ParseNATMasqueradeConfig(config)
			if err != nil {
				t.Fatalf("Parse step 1 failed: %v", err)
			}

			if len(result) != 1 || len(result[0].StaticEntries) != 1 {
				t.Fatalf("Expected 1 descriptor with 1 static entry, got %d descriptors", len(result))
			}

			parsedEntry := result[0].StaticEntries[0]

			// Verify parsed entry matches expected
			if parsedEntry.EntryNumber != tt.expected.EntryNumber {
				t.Errorf("Entry number mismatch: got %d, want %d", parsedEntry.EntryNumber, tt.expected.EntryNumber)
			}
			if parsedEntry.InsideLocal != tt.expected.InsideLocal {
				t.Errorf("InsideLocal mismatch: got %q, want %q", parsedEntry.InsideLocal, tt.expected.InsideLocal)
			}
			if parsedEntry.Protocol != tt.expected.Protocol {
				t.Errorf("Protocol mismatch: got %q, want %q", parsedEntry.Protocol, tt.expected.Protocol)
			}
			// Protocol-only entries should have nil ports
			if parsedEntry.InsideLocalPort != nil {
				t.Errorf("InsideLocalPort should be nil for protocol-only entry, got %d", *parsedEntry.InsideLocalPort)
			}
			if parsedEntry.OutsideGlobalPort != nil {
				t.Errorf("OutsideGlobalPort should be nil for protocol-only entry, got %d", *parsedEntry.OutsideGlobalPort)
			}

			// Step 2: Build command from parsed entry
			builtCmd := BuildNATMasqueradeStaticCommand(result[0].DescriptorID, parsedEntry.EntryNumber, parsedEntry)

			// Verify built command matches original input
			if builtCmd != tt.input {
				t.Errorf("Build command mismatch:\n  got:  %q\n  want: %q", builtCmd, tt.input)
			}

			// Step 3: Parse the built command again
			config2 := fmt.Sprintf("nat descriptor type %d masquerade\n%s", tt.descriptorID, builtCmd)
			result2, err := ParseNATMasqueradeConfig(config2)
			if err != nil {
				t.Fatalf("Parse step 2 failed: %v", err)
			}

			if len(result2) != 1 || len(result2[0].StaticEntries) != 1 {
				t.Fatalf("Expected 1 descriptor with 1 static entry after round-trip, got %d descriptors", len(result2))
			}

			// Verify round-trip produces identical entry
			roundTripEntry := result2[0].StaticEntries[0]
			if roundTripEntry.EntryNumber != parsedEntry.EntryNumber {
				t.Errorf("Round-trip entry number mismatch: got %d, want %d", roundTripEntry.EntryNumber, parsedEntry.EntryNumber)
			}
			if roundTripEntry.InsideLocal != parsedEntry.InsideLocal {
				t.Errorf("Round-trip InsideLocal mismatch: got %q, want %q", roundTripEntry.InsideLocal, parsedEntry.InsideLocal)
			}
			if roundTripEntry.Protocol != parsedEntry.Protocol {
				t.Errorf("Round-trip Protocol mismatch: got %q, want %q", roundTripEntry.Protocol, parsedEntry.Protocol)
			}
		})
	}
}

// TestNATMasqueradePortBasedRoundTrip tests the round-trip: parse → build → parse for port-based entries
func TestNATMasqueradePortBasedRoundTrip(t *testing.T) {
	tests := []struct {
		name         string
		descriptorID int
		input        string
		expected     MasqueradeStaticEntry
	}{
		{
			name:         "TCP with same ports round-trip (ipcp format)",
			descriptorID: 1,
			input:        "nat descriptor masquerade static 1 1 192.168.1.100 tcp 80",
			expected: MasqueradeStaticEntry{
				EntryNumber:       1,
				OutsideGlobal:     "ipcp",
				OutsideGlobalPort: intPtr(80),
				InsideLocal:       "192.168.1.100",
				InsideLocalPort:   intPtr(80),
				Protocol:          "tcp",
			},
		},
		{
			name:         "TCP with different ports round-trip (ipcp format)",
			descriptorID: 1,
			input:        "nat descriptor masquerade static 1 2 192.168.1.100 tcp 8080=80",
			expected: MasqueradeStaticEntry{
				EntryNumber:       2,
				OutsideGlobal:     "ipcp",
				OutsideGlobalPort: intPtr(8080),
				InsideLocal:       "192.168.1.100",
				InsideLocalPort:   intPtr(80),
				Protocol:          "tcp",
			},
		},
		{
			name:         "UDP with different ports round-trip (specific IP)",
			descriptorID: 2,
			input:        "nat descriptor masquerade static 2 3 203.0.113.1:8080=10.0.0.50:80 udp",
			expected: MasqueradeStaticEntry{
				EntryNumber:       3,
				OutsideGlobal:     "203.0.113.1",
				OutsideGlobalPort: intPtr(8080),
				InsideLocal:       "10.0.0.50",
				InsideLocalPort:   intPtr(80),
				Protocol:          "udp",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Parse the original input
			config := fmt.Sprintf("nat descriptor type %d masquerade\n%s", tt.descriptorID, tt.input)
			result, err := ParseNATMasqueradeConfig(config)
			if err != nil {
				t.Fatalf("Parse step 1 failed: %v", err)
			}

			if len(result) != 1 || len(result[0].StaticEntries) != 1 {
				t.Fatalf("Expected 1 descriptor with 1 static entry, got %d descriptors", len(result))
			}

			parsedEntry := result[0].StaticEntries[0]

			// Step 2: Build command from parsed entry
			builtCmd := BuildNATMasqueradeStaticCommand(result[0].DescriptorID, parsedEntry.EntryNumber, parsedEntry)

			// Verify built command matches original input
			if builtCmd != tt.input {
				t.Errorf("Build command mismatch:\n  got:  %q\n  want: %q", builtCmd, tt.input)
			}

			// Step 3: Parse the built command again
			config2 := fmt.Sprintf("nat descriptor type %d masquerade\n%s", tt.descriptorID, builtCmd)
			result2, err := ParseNATMasqueradeConfig(config2)
			if err != nil {
				t.Fatalf("Parse step 2 failed: %v", err)
			}

			if len(result2) != 1 || len(result2[0].StaticEntries) != 1 {
				t.Fatalf("Expected 1 descriptor with 1 static entry after round-trip, got %d descriptors", len(result2))
			}

			// Verify round-trip produces identical entry
			roundTripEntry := result2[0].StaticEntries[0]
			if roundTripEntry.EntryNumber != parsedEntry.EntryNumber {
				t.Errorf("Round-trip entry number mismatch")
			}
			if roundTripEntry.OutsideGlobal != parsedEntry.OutsideGlobal {
				t.Errorf("Round-trip OutsideGlobal mismatch: got %q, want %q", roundTripEntry.OutsideGlobal, parsedEntry.OutsideGlobal)
			}
			if *roundTripEntry.OutsideGlobalPort != *parsedEntry.OutsideGlobalPort {
				t.Errorf("Round-trip OutsideGlobalPort mismatch")
			}
			if roundTripEntry.InsideLocal != parsedEntry.InsideLocal {
				t.Errorf("Round-trip InsideLocal mismatch")
			}
			if *roundTripEntry.InsideLocalPort != *parsedEntry.InsideLocalPort {
				t.Errorf("Round-trip InsideLocalPort mismatch")
			}
			if roundTripEntry.Protocol != parsedEntry.Protocol {
				t.Errorf("Round-trip Protocol mismatch")
			}
		})
	}
}

// TestBuildNATDescriptorAddressInnerCommand_YAMLPatterns tests inner address command building
func TestBuildNATDescriptorAddressInnerCommand_YAMLPatterns(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		network  string
		expected string
	}{
		// Cases from nat.yaml section 23.5
		{
			name:     "full /24 range",
			id:       1,
			network:  "192.168.1.0-192.168.1.255",
			expected: "nat descriptor address inner 1 192.168.1.0-192.168.1.255",
		},
		{
			name:     "auto keyword",
			id:       1,
			network:  "auto",
			expected: "nat descriptor address inner 1 auto",
		},
		{
			name:     "specific host range",
			id:       2,
			network:  "192.168.0.1-192.168.0.100",
			expected: "nat descriptor address inner 2 192.168.0.1-192.168.0.100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNATDescriptorAddressInnerCommand(tt.id, tt.network)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Edge case tests for NAT (Tasks 10-12)
// =============================================================================

// TestParseNATPortRanges tests parsing of single port, port ranges, and multiple protocols
// Task 10: NAT edge cases - Port ranges
func TestParseNATPortRanges(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []NATMasquerade
		wantErr  bool
	}{
		{
			name: "single port - HTTP",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 ipcp:80=192.168.1.100:80 tcp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{
							EntryNumber:       1,
							OutsideGlobal:     "ipcp",
							OutsideGlobalPort: intPtr(80),
							InsideLocal:       "192.168.1.100",
							InsideLocalPort:   intPtr(80),
							Protocol:          "tcp",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "single port - well-known ports",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 ipcp:22=192.168.1.10:22 tcp
nat descriptor masquerade static 1 2 ipcp:23=192.168.1.10:23 tcp
nat descriptor masquerade static 1 3 ipcp:25=192.168.1.20:25 tcp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(22), InsideLocal: "192.168.1.10", InsideLocalPort: intPtr(22), Protocol: "tcp"},
						{EntryNumber: 2, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(23), InsideLocal: "192.168.1.10", InsideLocalPort: intPtr(23), Protocol: "tcp"},
						{EntryNumber: 3, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(25), InsideLocal: "192.168.1.20", InsideLocalPort: intPtr(25), Protocol: "tcp"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "high port numbers",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 ipcp:8080=192.168.1.100:80 tcp
nat descriptor masquerade static 1 2 ipcp:8443=192.168.1.100:443 tcp
nat descriptor masquerade static 1 3 ipcp:65535=192.168.1.100:65535 tcp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(8080), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(80), Protocol: "tcp"},
						{EntryNumber: 2, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(8443), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(443), Protocol: "tcp"},
						{EntryNumber: 3, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(65535), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(65535), Protocol: "tcp"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "minimum port number 1",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 ipcp:1=192.168.1.100:1 tcp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(1), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(1), Protocol: "tcp"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "port forwarding with different inner and outer ports",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 203.0.113.1:80=192.168.1.100:8080 tcp
nat descriptor masquerade static 1 2 203.0.113.1:443=192.168.1.100:8443 tcp
nat descriptor masquerade static 1 3 203.0.113.1:53=192.168.1.200:5353 udp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, OutsideGlobal: "203.0.113.1", OutsideGlobalPort: intPtr(80), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(8080), Protocol: "tcp"},
						{EntryNumber: 2, OutsideGlobal: "203.0.113.1", OutsideGlobalPort: intPtr(443), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(8443), Protocol: "tcp"},
						{EntryNumber: 3, OutsideGlobal: "203.0.113.1", OutsideGlobalPort: intPtr(53), InsideLocal: "192.168.1.200", InsideLocalPort: intPtr(5353), Protocol: "udp"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "common service ports",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 ipcp:21=192.168.1.100:21 tcp
nat descriptor masquerade static 1 2 ipcp:110=192.168.1.100:110 tcp
nat descriptor masquerade static 1 3 ipcp:143=192.168.1.100:143 tcp
nat descriptor masquerade static 1 4 ipcp:993=192.168.1.100:993 tcp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(21), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(21), Protocol: "tcp"},
						{EntryNumber: 2, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(110), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(110), Protocol: "tcp"},
						{EntryNumber: 3, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(143), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(143), Protocol: "tcp"},
						{EntryNumber: 4, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(993), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(993), Protocol: "tcp"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple descriptors with different ports",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 ipcp:80=192.168.1.100:80 tcp
nat descriptor type 2 masquerade
nat descriptor masquerade static 2 1 ipcp:443=192.168.2.100:443 tcp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(80), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(80), Protocol: "tcp"},
					},
				},
				{
					DescriptorID: 2,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(443), InsideLocal: "192.168.2.100", InsideLocalPort: intPtr(443), Protocol: "tcp"},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseNATMasqueradeConfig(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d descriptors, got %d", len(tt.expected), len(result))
				return
			}

			resultMap := make(map[int]NATMasquerade)
			for _, r := range result {
				resultMap[r.DescriptorID] = r
			}

			for _, exp := range tt.expected {
				got, ok := resultMap[exp.DescriptorID]
				if !ok {
					t.Errorf("descriptor %d not found in result", exp.DescriptorID)
					continue
				}

				if len(got.StaticEntries) != len(exp.StaticEntries) {
					t.Errorf("descriptor %d: static entries count = %d, want %d", exp.DescriptorID, len(got.StaticEntries), len(exp.StaticEntries))
					continue
				}

				for i, expEntry := range exp.StaticEntries {
					gotEntry := got.StaticEntries[i]
					if gotEntry.EntryNumber != expEntry.EntryNumber {
						t.Errorf("descriptor %d, entry %d: entry number = %d, want %d", exp.DescriptorID, i, gotEntry.EntryNumber, expEntry.EntryNumber)
					}
					// Compare port pointers
					if (gotEntry.OutsideGlobalPort == nil) != (expEntry.OutsideGlobalPort == nil) {
						t.Errorf("descriptor %d, entry %d: outside port nil mismatch, got nil=%v, want nil=%v", exp.DescriptorID, i, gotEntry.OutsideGlobalPort == nil, expEntry.OutsideGlobalPort == nil)
					} else if gotEntry.OutsideGlobalPort != nil && *gotEntry.OutsideGlobalPort != *expEntry.OutsideGlobalPort {
						t.Errorf("descriptor %d, entry %d: outside port = %d, want %d", exp.DescriptorID, i, *gotEntry.OutsideGlobalPort, *expEntry.OutsideGlobalPort)
					}
					if (gotEntry.InsideLocalPort == nil) != (expEntry.InsideLocalPort == nil) {
						t.Errorf("descriptor %d, entry %d: inside port nil mismatch, got nil=%v, want nil=%v", exp.DescriptorID, i, gotEntry.InsideLocalPort == nil, expEntry.InsideLocalPort == nil)
					} else if gotEntry.InsideLocalPort != nil && *gotEntry.InsideLocalPort != *expEntry.InsideLocalPort {
						t.Errorf("descriptor %d, entry %d: inside port = %d, want %d", exp.DescriptorID, i, *gotEntry.InsideLocalPort, *expEntry.InsideLocalPort)
					}
					if gotEntry.Protocol != expEntry.Protocol {
						t.Errorf("descriptor %d, entry %d: protocol = %q, want %q", exp.DescriptorID, i, gotEntry.Protocol, expEntry.Protocol)
					}
				}
			}
		})
	}
}

// TestParseNATProtocols tests parsing of different protocols (TCP, UDP, etc.)
// Task 11: NAT edge cases - Protocols
func TestParseNATProtocols(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []NATMasquerade
		wantErr  bool
	}{
		{
			name: "TCP-only protocol",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 ipcp:80=192.168.1.100:80 tcp
nat descriptor masquerade static 1 2 ipcp:443=192.168.1.100:443 tcp
nat descriptor masquerade static 1 3 ipcp:22=192.168.1.100:22 tcp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(80), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(80), Protocol: "tcp"},
						{EntryNumber: 2, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(443), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(443), Protocol: "tcp"},
						{EntryNumber: 3, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(22), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(22), Protocol: "tcp"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "UDP-only protocol",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 ipcp:53=192.168.1.200:53 udp
nat descriptor masquerade static 1 2 ipcp:123=192.168.1.200:123 udp
nat descriptor masquerade static 1 3 ipcp:161=192.168.1.200:161 udp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(53), InsideLocal: "192.168.1.200", InsideLocalPort: intPtr(53), Protocol: "udp"},
						{EntryNumber: 2, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(123), InsideLocal: "192.168.1.200", InsideLocalPort: intPtr(123), Protocol: "udp"},
						{EntryNumber: 3, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(161), InsideLocal: "192.168.1.200", InsideLocalPort: intPtr(161), Protocol: "udp"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "mixed TCP and UDP protocols",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 ipcp:80=192.168.1.100:80 tcp
nat descriptor masquerade static 1 2 ipcp:53=192.168.1.200:53 udp
nat descriptor masquerade static 1 3 ipcp:443=192.168.1.100:443 tcp
nat descriptor masquerade static 1 4 ipcp:123=192.168.1.200:123 udp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(80), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(80), Protocol: "tcp"},
						{EntryNumber: 2, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(53), InsideLocal: "192.168.1.200", InsideLocalPort: intPtr(53), Protocol: "udp"},
						{EntryNumber: 3, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(443), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(443), Protocol: "tcp"},
						{EntryNumber: 4, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(123), InsideLocal: "192.168.1.200", InsideLocalPort: intPtr(123), Protocol: "udp"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "without protocol (any protocol)",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 ipcp:80=192.168.1.100:80`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(80), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(80), Protocol: ""},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "uppercase protocol TCP",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 ipcp:80=192.168.1.100:80 TCP`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(80), InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(80), Protocol: "tcp"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "uppercase protocol UDP",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 ipcp:53=192.168.1.200:53 UDP`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(53), InsideLocal: "192.168.1.200", InsideLocalPort: intPtr(53), Protocol: "udp"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "DNS on both TCP and UDP",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 ipcp:53=192.168.1.200:53 tcp
nat descriptor masquerade static 1 2 ipcp:53=192.168.1.200:53 udp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(53), InsideLocal: "192.168.1.200", InsideLocalPort: intPtr(53), Protocol: "tcp"},
						{EntryNumber: 2, OutsideGlobal: "ipcp", OutsideGlobalPort: intPtr(53), InsideLocal: "192.168.1.200", InsideLocalPort: intPtr(53), Protocol: "udp"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "alternate format with tcp protocol",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 192.168.1.100 tcp 80`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, InsideLocal: "192.168.1.100", InsideLocalPort: intPtr(80), OutsideGlobalPort: intPtr(80), Protocol: "tcp"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "alternate format with udp protocol",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 192.168.1.200 udp 53`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, InsideLocal: "192.168.1.200", InsideLocalPort: intPtr(53), OutsideGlobalPort: intPtr(53), Protocol: "udp"},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseNATMasqueradeConfig(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d descriptors, got %d", len(tt.expected), len(result))
				return
			}

			resultMap := make(map[int]NATMasquerade)
			for _, r := range result {
				resultMap[r.DescriptorID] = r
			}

			for _, exp := range tt.expected {
				got, ok := resultMap[exp.DescriptorID]
				if !ok {
					t.Errorf("descriptor %d not found in result", exp.DescriptorID)
					continue
				}

				if len(got.StaticEntries) != len(exp.StaticEntries) {
					t.Errorf("descriptor %d: static entries count = %d, want %d", exp.DescriptorID, len(got.StaticEntries), len(exp.StaticEntries))
					continue
				}

				for i, expEntry := range exp.StaticEntries {
					gotEntry := got.StaticEntries[i]
					if gotEntry.Protocol != expEntry.Protocol {
						t.Errorf("descriptor %d, entry %d: protocol = %q, want %q", exp.DescriptorID, i, gotEntry.Protocol, expEntry.Protocol)
					}
					if gotEntry.InsideLocal != expEntry.InsideLocal {
						t.Errorf("descriptor %d, entry %d: inside local = %q, want %q", exp.DescriptorID, i, gotEntry.InsideLocal, expEntry.InsideLocal)
					}
					// Compare port pointers
					if (gotEntry.InsideLocalPort == nil) != (expEntry.InsideLocalPort == nil) {
						t.Errorf("descriptor %d, entry %d: inside port nil mismatch, got nil=%v, want nil=%v", exp.DescriptorID, i, gotEntry.InsideLocalPort == nil, expEntry.InsideLocalPort == nil)
					} else if gotEntry.InsideLocalPort != nil && *gotEntry.InsideLocalPort != *expEntry.InsideLocalPort {
						t.Errorf("descriptor %d, entry %d: inside port = %d, want %d", exp.DescriptorID, i, *gotEntry.InsideLocalPort, *expEntry.InsideLocalPort)
					}
					if (gotEntry.OutsideGlobalPort == nil) != (expEntry.OutsideGlobalPort == nil) {
						t.Errorf("descriptor %d, entry %d: outside port nil mismatch, got nil=%v, want nil=%v", exp.DescriptorID, i, gotEntry.OutsideGlobalPort == nil, expEntry.OutsideGlobalPort == nil)
					} else if gotEntry.OutsideGlobalPort != nil && *gotEntry.OutsideGlobalPort != *expEntry.OutsideGlobalPort {
						t.Errorf("descriptor %d, entry %d: outside port = %d, want %d", exp.DescriptorID, i, *gotEntry.OutsideGlobalPort, *expEntry.OutsideGlobalPort)
					}
				}
			}
		})
	}
}

// TestParseNATDynamic tests parsing of dynamic NAT configurations including pools and overload
// Task 12: NAT edge cases - Dynamic configurations
func TestParseNATDynamic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []NATMasquerade
		wantErr  bool
	}{
		{
			name: "basic masquerade with outer address range",
			input: `nat descriptor type 1 masquerade
nat descriptor address outer 1 203.0.113.1-203.0.113.16
nat descriptor address inner 1 192.168.1.0-192.168.1.255`,
			expected: []NATMasquerade{
				{
					DescriptorID:  1,
					OuterAddress:  "203.0.113.1-203.0.113.16",
					InnerNetwork:  "192.168.1.0-192.168.1.255",
					StaticEntries: []MasqueradeStaticEntry{},
				},
			},
			wantErr: false,
		},
		{
			name: "PAT (Port Address Translation) with ipcp",
			input: `nat descriptor type 1 masquerade
nat descriptor address outer 1 ipcp
nat descriptor address inner 1 192.168.0.0-192.168.0.255`,
			expected: []NATMasquerade{
				{
					DescriptorID:  1,
					OuterAddress:  "ipcp",
					InnerNetwork:  "192.168.0.0-192.168.0.255",
					StaticEntries: []MasqueradeStaticEntry{},
				},
			},
			wantErr: false,
		},
		{
			name: "PAT with single outer IP",
			input: `nat descriptor type 1 masquerade
nat descriptor address outer 1 203.0.113.1
nat descriptor address inner 1 10.0.0.0-10.0.255.255`,
			expected: []NATMasquerade{
				{
					DescriptorID:  1,
					OuterAddress:  "203.0.113.1",
					InnerNetwork:  "10.0.0.0-10.0.255.255",
					StaticEntries: []MasqueradeStaticEntry{},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple inner address ranges with different descriptors",
			input: `nat descriptor type 1 masquerade
nat descriptor address outer 1 ipcp
nat descriptor address inner 1 192.168.1.0-192.168.1.255
nat descriptor type 2 masquerade
nat descriptor address outer 2 ipcp
nat descriptor address inner 2 192.168.2.0-192.168.2.255`,
			expected: []NATMasquerade{
				{
					DescriptorID:  1,
					OuterAddress:  "ipcp",
					InnerNetwork:  "192.168.1.0-192.168.1.255",
					StaticEntries: []MasqueradeStaticEntry{},
				},
				{
					DescriptorID:  2,
					OuterAddress:  "ipcp",
					InnerNetwork:  "192.168.2.0-192.168.2.255",
					StaticEntries: []MasqueradeStaticEntry{},
				},
			},
			wantErr: false,
		},
		{
			name: "auto inner address",
			input: `nat descriptor type 1 masquerade
nat descriptor address outer 1 ipcp
nat descriptor address inner 1 auto`,
			expected: []NATMasquerade{
				{
					DescriptorID:  1,
					OuterAddress:  "ipcp",
					InnerNetwork:  "auto",
					StaticEntries: []MasqueradeStaticEntry{},
				},
			},
			wantErr: false,
		},
		{
			name: "primary outer address",
			input: `nat descriptor type 1 masquerade
nat descriptor address outer 1 primary
nat descriptor address inner 1 192.168.1.0-192.168.1.255`,
			expected: []NATMasquerade{
				{
					DescriptorID:  1,
					OuterAddress:  "primary",
					InnerNetwork:  "192.168.1.0-192.168.1.255",
					StaticEntries: []MasqueradeStaticEntry{},
				},
			},
			wantErr: false,
		},
		{
			name: "pool with static entries combined",
			input: `nat descriptor type 1 masquerade
nat descriptor address outer 1 203.0.113.1-203.0.113.10
nat descriptor address inner 1 192.168.0.0-192.168.0.255
nat descriptor masquerade static 1 1 203.0.113.1:80=192.168.0.100:80 tcp
nat descriptor masquerade static 1 2 203.0.113.1:443=192.168.0.100:443 tcp`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					OuterAddress: "203.0.113.1-203.0.113.10",
					InnerNetwork: "192.168.0.0-192.168.0.255",
					StaticEntries: []MasqueradeStaticEntry{
						{EntryNumber: 1, OutsideGlobal: "203.0.113.1", OutsideGlobalPort: intPtr(80), InsideLocal: "192.168.0.100", InsideLocalPort: intPtr(80), Protocol: "tcp"},
						{EntryNumber: 2, OutsideGlobal: "203.0.113.1", OutsideGlobalPort: intPtr(443), InsideLocal: "192.168.0.100", InsideLocalPort: intPtr(443), Protocol: "tcp"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "small inner address range",
			input: `nat descriptor type 1 masquerade
nat descriptor address outer 1 ipcp
nat descriptor address inner 1 192.168.1.100-192.168.1.110`,
			expected: []NATMasquerade{
				{
					DescriptorID:  1,
					OuterAddress:  "ipcp",
					InnerNetwork:  "192.168.1.100-192.168.1.110",
					StaticEntries: []MasqueradeStaticEntry{},
				},
			},
			wantErr: false,
		},
		{
			name: "class A inner address range",
			input: `nat descriptor type 1 masquerade
nat descriptor address outer 1 ipcp
nat descriptor address inner 1 10.0.0.0-10.255.255.255`,
			expected: []NATMasquerade{
				{
					DescriptorID:  1,
					OuterAddress:  "ipcp",
					InnerNetwork:  "10.0.0.0-10.255.255.255",
					StaticEntries: []MasqueradeStaticEntry{},
				},
			},
			wantErr: false,
		},
		{
			name: "class B inner address range",
			input: `nat descriptor type 1 masquerade
nat descriptor address outer 1 ipcp
nat descriptor address inner 1 172.16.0.0-172.31.255.255`,
			expected: []NATMasquerade{
				{
					DescriptorID:  1,
					OuterAddress:  "ipcp",
					InnerNetwork:  "172.16.0.0-172.31.255.255",
					StaticEntries: []MasqueradeStaticEntry{},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseNATMasqueradeConfig(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d descriptors, got %d", len(tt.expected), len(result))
				return
			}

			resultMap := make(map[int]NATMasquerade)
			for _, r := range result {
				resultMap[r.DescriptorID] = r
			}

			for _, exp := range tt.expected {
				got, ok := resultMap[exp.DescriptorID]
				if !ok {
					t.Errorf("descriptor %d not found in result", exp.DescriptorID)
					continue
				}

				if got.OuterAddress != exp.OuterAddress {
					t.Errorf("descriptor %d: outer address = %q, want %q", exp.DescriptorID, got.OuterAddress, exp.OuterAddress)
				}
				if got.InnerNetwork != exp.InnerNetwork {
					t.Errorf("descriptor %d: inner network = %q, want %q", exp.DescriptorID, got.InnerNetwork, exp.InnerNetwork)
				}
				if len(got.StaticEntries) != len(exp.StaticEntries) {
					t.Errorf("descriptor %d: static entries count = %d, want %d", exp.DescriptorID, len(got.StaticEntries), len(exp.StaticEntries))
					continue
				}

				for i, expEntry := range exp.StaticEntries {
					gotEntry := got.StaticEntries[i]
					if gotEntry.EntryNumber != expEntry.EntryNumber {
						t.Errorf("descriptor %d, entry %d: entry number = %d, want %d", exp.DescriptorID, i, gotEntry.EntryNumber, expEntry.EntryNumber)
					}
					if gotEntry.OutsideGlobal != expEntry.OutsideGlobal {
						t.Errorf("descriptor %d, entry %d: outside global = %q, want %q", exp.DescriptorID, i, gotEntry.OutsideGlobal, expEntry.OutsideGlobal)
					}
					// Compare port pointers
					if (gotEntry.OutsideGlobalPort == nil) != (expEntry.OutsideGlobalPort == nil) {
						t.Errorf("descriptor %d, entry %d: outside port nil mismatch, got nil=%v, want nil=%v", exp.DescriptorID, i, gotEntry.OutsideGlobalPort == nil, expEntry.OutsideGlobalPort == nil)
					} else if gotEntry.OutsideGlobalPort != nil && *gotEntry.OutsideGlobalPort != *expEntry.OutsideGlobalPort {
						t.Errorf("descriptor %d, entry %d: outside port = %d, want %d", exp.DescriptorID, i, *gotEntry.OutsideGlobalPort, *expEntry.OutsideGlobalPort)
					}
					if gotEntry.InsideLocal != expEntry.InsideLocal {
						t.Errorf("descriptor %d, entry %d: inside local = %q, want %q", exp.DescriptorID, i, gotEntry.InsideLocal, expEntry.InsideLocal)
					}
					// Compare port pointers
					if (gotEntry.InsideLocalPort == nil) != (expEntry.InsideLocalPort == nil) {
						t.Errorf("descriptor %d, entry %d: inside port nil mismatch, got nil=%v, want nil=%v", exp.DescriptorID, i, gotEntry.InsideLocalPort == nil, expEntry.InsideLocalPort == nil)
					} else if gotEntry.InsideLocalPort != nil && *gotEntry.InsideLocalPort != *expEntry.InsideLocalPort {
						t.Errorf("descriptor %d, entry %d: inside port = %d, want %d", exp.DescriptorID, i, *gotEntry.InsideLocalPort, *expEntry.InsideLocalPort)
					}
					if gotEntry.Protocol != expEntry.Protocol {
						t.Errorf("descriptor %d, entry %d: protocol = %q, want %q", exp.DescriptorID, i, gotEntry.Protocol, expEntry.Protocol)
					}
				}
			}
		})
	}
}

// TestParseNATMasqueradeStaticAltFormat tests the alternate format from nat.yaml 23.8
// Format: nat descriptor masquerade static <id> <entry_id> <inner_ip> <protocol> [<outer_port>=]<inner_port>
func TestParseNATMasqueradeStaticAltFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []NATMasquerade
		wantErr  bool
	}{
		{
			name: "alternate format - HTTP same port",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 192.168.1.100 tcp 80`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{
							EntryNumber:       1,
							InsideLocal:       "192.168.1.100",
							InsideLocalPort:   intPtr(80),
							OutsideGlobalPort: intPtr(80), // Same as inside port
							Protocol:          "tcp",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "alternate format - UDP DNS same port",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 3 192.168.1.200 udp 53`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{
							EntryNumber:       3,
							InsideLocal:       "192.168.1.200",
							InsideLocalPort:   intPtr(53),
							OutsideGlobalPort: intPtr(53),
							Protocol:          "udp",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "alternate format - different outer and inner ports",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 2 192.168.1.100 tcp 8080=80`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{
							EntryNumber:       2,
							InsideLocal:       "192.168.1.100",
							InsideLocalPort:   intPtr(80),
							OutsideGlobalPort: intPtr(8080),
							Protocol:          "tcp",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "mixed formats in same descriptor",
			input: `nat descriptor type 1 masquerade
nat descriptor address outer 1 ipcp
nat descriptor address inner 1 192.168.1.0-192.168.1.255
nat descriptor masquerade static 1 1 192.168.1.100 tcp 80
nat descriptor masquerade static 1 2 ipcp:8080=192.168.1.100:8080 tcp
nat descriptor masquerade static 1 3 192.168.1.200 udp 53`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					OuterAddress: "ipcp",
					InnerNetwork: "192.168.1.0-192.168.1.255",
					StaticEntries: []MasqueradeStaticEntry{
						{
							EntryNumber:       1,
							InsideLocal:       "192.168.1.100",
							InsideLocalPort:   intPtr(80),
							OutsideGlobalPort: intPtr(80),
							Protocol:          "tcp",
						},
						{
							EntryNumber:       2,
							OutsideGlobal:     "ipcp",
							OutsideGlobalPort: intPtr(8080),
							InsideLocal:       "192.168.1.100",
							InsideLocalPort:   intPtr(8080),
							Protocol:          "tcp",
						},
						{
							EntryNumber:       3,
							InsideLocal:       "192.168.1.200",
							InsideLocalPort:   intPtr(53),
							OutsideGlobalPort: intPtr(53),
							Protocol:          "udp",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "all examples from nat.yaml 23.8",
			input: `nat descriptor type 1 masquerade
nat descriptor masquerade static 1 1 192.168.1.100 tcp 80
nat descriptor masquerade static 1 2 192.168.1.100 tcp 8080=80
nat descriptor masquerade static 1 3 192.168.1.200 udp 53`,
			expected: []NATMasquerade{
				{
					DescriptorID: 1,
					StaticEntries: []MasqueradeStaticEntry{
						{
							EntryNumber:       1,
							InsideLocal:       "192.168.1.100",
							InsideLocalPort:   intPtr(80),
							OutsideGlobalPort: intPtr(80),
							Protocol:          "tcp",
						},
						{
							EntryNumber:       2,
							InsideLocal:       "192.168.1.100",
							InsideLocalPort:   intPtr(80),
							OutsideGlobalPort: intPtr(8080),
							Protocol:          "tcp",
						},
						{
							EntryNumber:       3,
							InsideLocal:       "192.168.1.200",
							InsideLocalPort:   intPtr(53),
							OutsideGlobalPort: intPtr(53),
							Protocol:          "udp",
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseNATMasqueradeConfig(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d descriptors, got %d", len(tt.expected), len(result))
				return
			}

			resultMap := make(map[int]NATMasquerade)
			for _, r := range result {
				resultMap[r.DescriptorID] = r
			}

			for _, exp := range tt.expected {
				got, ok := resultMap[exp.DescriptorID]
				if !ok {
					t.Errorf("descriptor %d not found in result", exp.DescriptorID)
					continue
				}

				if got.OuterAddress != exp.OuterAddress {
					t.Errorf("descriptor %d: outer address = %q, want %q", exp.DescriptorID, got.OuterAddress, exp.OuterAddress)
				}
				if got.InnerNetwork != exp.InnerNetwork {
					t.Errorf("descriptor %d: inner network = %q, want %q", exp.DescriptorID, got.InnerNetwork, exp.InnerNetwork)
				}
				if len(got.StaticEntries) != len(exp.StaticEntries) {
					t.Errorf("descriptor %d: static entries count = %d, want %d", exp.DescriptorID, len(got.StaticEntries), len(exp.StaticEntries))
					continue
				}

				for i, expEntry := range exp.StaticEntries {
					gotEntry := got.StaticEntries[i]
					if gotEntry.EntryNumber != expEntry.EntryNumber {
						t.Errorf("descriptor %d, entry %d: entry number = %d, want %d", exp.DescriptorID, i, gotEntry.EntryNumber, expEntry.EntryNumber)
					}
					if gotEntry.InsideLocal != expEntry.InsideLocal {
						t.Errorf("descriptor %d, entry %d: inside local = %q, want %q", exp.DescriptorID, i, gotEntry.InsideLocal, expEntry.InsideLocal)
					}
					// Compare port pointers
					if (gotEntry.InsideLocalPort == nil) != (expEntry.InsideLocalPort == nil) {
						t.Errorf("descriptor %d, entry %d: inside port nil mismatch, got nil=%v, want nil=%v", exp.DescriptorID, i, gotEntry.InsideLocalPort == nil, expEntry.InsideLocalPort == nil)
					} else if gotEntry.InsideLocalPort != nil && *gotEntry.InsideLocalPort != *expEntry.InsideLocalPort {
						t.Errorf("descriptor %d, entry %d: inside port = %d, want %d", exp.DescriptorID, i, *gotEntry.InsideLocalPort, *expEntry.InsideLocalPort)
					}
					if (gotEntry.OutsideGlobalPort == nil) != (expEntry.OutsideGlobalPort == nil) {
						t.Errorf("descriptor %d, entry %d: outside port nil mismatch, got nil=%v, want nil=%v", exp.DescriptorID, i, gotEntry.OutsideGlobalPort == nil, expEntry.OutsideGlobalPort == nil)
					} else if gotEntry.OutsideGlobalPort != nil && *gotEntry.OutsideGlobalPort != *expEntry.OutsideGlobalPort {
						t.Errorf("descriptor %d, entry %d: outside port = %d, want %d", exp.DescriptorID, i, *gotEntry.OutsideGlobalPort, *expEntry.OutsideGlobalPort)
					}
					if gotEntry.Protocol != expEntry.Protocol {
						t.Errorf("descriptor %d, entry %d: protocol = %q, want %q", exp.DescriptorID, i, gotEntry.Protocol, expEntry.Protocol)
					}
					// Only check OutsideGlobal if expected
					if expEntry.OutsideGlobal != "" && gotEntry.OutsideGlobal != expEntry.OutsideGlobal {
						t.Errorf("descriptor %d, entry %d: outside global = %q, want %q", exp.DescriptorID, i, gotEntry.OutsideGlobal, expEntry.OutsideGlobal)
					}
				}
			}
		})
	}
}
