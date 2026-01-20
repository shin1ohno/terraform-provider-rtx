package parsers

import (
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
							OutsideGlobalPort: 80,
							InsideLocal:       "192.168.2.100",
							InsideLocalPort:   8080,
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
							OutsideGlobalPort: 80,
							InsideLocal:       "10.0.0.10",
							InsideLocalPort:   80,
							Protocol:          "tcp",
						},
						{
							EntryNumber:       2,
							OutsideGlobal:     "ipcp",
							OutsideGlobalPort: 443,
							InsideLocal:       "10.0.0.10",
							InsideLocalPort:   443,
							Protocol:          "tcp",
						},
						{
							EntryNumber:       3,
							OutsideGlobal:     "ipcp",
							OutsideGlobalPort: 53,
							InsideLocal:       "10.0.0.20",
							InsideLocalPort:   53,
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
							OutsideGlobalPort: 22,
							InsideLocal:       "192.168.1.1",
							InsideLocalPort:   22,
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
					if gotEntry.OutsideGlobalPort != expEntry.OutsideGlobalPort {
						t.Errorf("descriptor %d, entry %d: outside port = %d, want %d", exp.DescriptorID, i, gotEntry.OutsideGlobalPort, expEntry.OutsideGlobalPort)
					}
					if gotEntry.InsideLocal != expEntry.InsideLocal {
						t.Errorf("descriptor %d, entry %d: inside local = %q, want %q", exp.DescriptorID, i, gotEntry.InsideLocal, expEntry.InsideLocal)
					}
					if gotEntry.InsideLocalPort != expEntry.InsideLocalPort {
						t.Errorf("descriptor %d, entry %d: inside port = %d, want %d", exp.DescriptorID, i, gotEntry.InsideLocalPort, expEntry.InsideLocalPort)
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
				OutsideGlobalPort: 80,
				InsideLocal:       "192.168.1.100",
				InsideLocalPort:   8080,
				Protocol:          "tcp",
			},
			expected: "nat descriptor masquerade static 1 1 203.0.113.1:80=192.168.1.100:8080 tcp",
		},
		{
			name:     "udp with ipcp",
			id:       2,
			entryNum: 3,
			entry: MasqueradeStaticEntry{
				OutsideGlobal:     "ipcp",
				OutsideGlobalPort: 53,
				InsideLocal:       "10.0.0.1",
				InsideLocalPort:   53,
				Protocol:          "udp",
			},
			expected: "nat descriptor masquerade static 2 3 ipcp:53=10.0.0.1:53 udp",
		},
		{
			name:     "without protocol",
			id:       1,
			entryNum: 2,
			entry: MasqueradeStaticEntry{
				OutsideGlobal:     "203.0.113.1",
				OutsideGlobalPort: 22,
				InsideLocal:       "192.168.1.50",
				InsideLocalPort:   22,
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
				OutsideGlobalPort: 443,
				InsideLocal:       "192.168.1.100",
				InsideLocalPort:   443,
				Protocol:          "TCP",
			},
			expected: "nat descriptor masquerade static 1 1 203.0.113.1:443=192.168.1.100:443 tcp",
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
			name:     "icmp is invalid",
			protocol: "icmp",
			wantErr:  true,
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
						OutsideGlobalPort: 80,
						InsideLocal:       "192.168.1.100",
						InsideLocalPort:   8080,
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
			name: "invalid static entry port",
			nat: NATMasquerade{
				DescriptorID: 1,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.1.0-192.168.1.255",
				StaticEntries: []MasqueradeStaticEntry{
					{
						EntryNumber:       1,
						OutsideGlobal:     "203.0.113.1",
						OutsideGlobalPort: 0,
						InsideLocal:       "192.168.1.100",
						InsideLocalPort:   8080,
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
						OutsideGlobalPort: 80,
						InsideLocal:       "not-an-ip",
						InsideLocalPort:   8080,
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
						OutsideGlobalPort: 80,
						InsideLocal:       "192.168.1.100",
						InsideLocalPort:   8080,
						Protocol:          "icmp",
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
			expected: `show config | grep -E "nat descriptor (type|address outer|address inner|masquerade static) 1 "`,
		},
		{
			name:     "descriptor 10",
			id:       10,
			expected: `show config | grep -E "nat descriptor (type|address outer|address inner|masquerade static) 10 "`,
		},
		{
			name:     "descriptor 100",
			id:       100,
			expected: `show config | grep -E "nat descriptor (type|address outer|address inner|masquerade static) 100 "`,
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
