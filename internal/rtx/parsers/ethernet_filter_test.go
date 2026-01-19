package parsers

import (
	"strings"
	"testing"
)

func TestNormalizeMAC(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "colon-separated lowercase",
			input:    "00:11:22:33:44:55",
			expected: "00:11:22:33:44:55",
		},
		{
			name:     "colon-separated uppercase",
			input:    "00:11:22:AA:BB:CC",
			expected: "00:11:22:aa:bb:cc",
		},
		{
			name:     "Cisco dot notation",
			input:    "0011.2233.4455",
			expected: "00:11:22:33:44:55",
		},
		{
			name:     "hyphen-separated",
			input:    "00-11-22-33-44-55",
			expected: "00:11:22:33:44:55",
		},
		{
			name:     "no separator",
			input:    "001122334455",
			expected: "00:11:22:33:44:55",
		},
		{
			name:     "wildcard",
			input:    "*",
			expected: "*",
		},
		{
			name:     "mixed case",
			input:    "AA:bb:CC:dd:EE:ff",
			expected: "aa:bb:cc:dd:ee:ff",
		},
		{
			name:     "with whitespace",
			input:    "  00:11:22:33:44:55  ",
			expected: "00:11:22:33:44:55",
		},
		{
			name:     "invalid - too short",
			input:    "00:11:22",
			expected: "00:11:22",
		},
		{
			name:     "invalid - non-hex",
			input:    "00:11:22:33:44:GG",
			expected: "00:11:22:33:44:GG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeMAC(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeMAC(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertMACToCisco(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "colon-separated to Cisco",
			input:    "00:11:22:33:44:55",
			expected: "0011.2233.4455",
		},
		{
			name:     "already Cisco format",
			input:    "0011.2233.4455",
			expected: "0011.2233.4455",
		},
		{
			name:     "hyphen-separated to Cisco",
			input:    "00-11-22-33-44-55",
			expected: "0011.2233.4455",
		},
		{
			name:     "no separator to Cisco",
			input:    "001122334455",
			expected: "0011.2233.4455",
		},
		{
			name:     "wildcard",
			input:    "*",
			expected: "*",
		},
		{
			name:     "uppercase to lowercase",
			input:    "AA:BB:CC:DD:EE:FF",
			expected: "aabb.ccdd.eeff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertMACToCisco(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertMACToCisco(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseEthernetFilterConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []EthernetFilter
		wantErr  bool
	}{
		{
			name:  "basic filter with all wildcards",
			input: `ethernet filter 1 pass * * *`,
			expected: []EthernetFilter{
				{
					Number:    1,
					Action:    "pass",
					SourceMAC: "*",
					DestMAC:   "*",
				},
			},
		},
		{
			name:  "filter with EtherType",
			input: `ethernet filter 1 pass * * 0x0800`,
			expected: []EthernetFilter{
				{
					Number:    1,
					Action:    "pass",
					SourceMAC: "*",
					DestMAC:   "*",
					EtherType: "0x0800",
				},
			},
		},
		{
			name:  "filter with source MAC",
			input: `ethernet filter 2 reject 00:11:22:33:44:55 * *`,
			expected: []EthernetFilter{
				{
					Number:    2,
					Action:    "reject",
					SourceMAC: "00:11:22:33:44:55",
					DestMAC:   "*",
				},
			},
		},
		{
			name:  "filter with VLAN ID",
			input: `ethernet filter 3 pass * * 0x0800 vlan 100`,
			expected: []EthernetFilter{
				{
					Number:    3,
					Action:    "pass",
					SourceMAC: "*",
					DestMAC:   "*",
					EtherType: "0x0800",
					VlanID:    100,
				},
			},
		},
		{
			name: "multiple filters",
			input: `ethernet filter 1 pass * * 0x0800
ethernet filter 2 reject 00:11:22:33:44:55 * *
ethernet filter 3 pass * ff:ff:ff:ff:ff:ff 0x0806`,
			expected: []EthernetFilter{
				{
					Number:    1,
					Action:    "pass",
					SourceMAC: "*",
					DestMAC:   "*",
					EtherType: "0x0800",
				},
				{
					Number:    2,
					Action:    "reject",
					SourceMAC: "00:11:22:33:44:55",
					DestMAC:   "*",
				},
				{
					Number:    3,
					Action:    "pass",
					SourceMAC: "*",
					DestMAC:   "ff:ff:ff:ff:ff:ff",
					EtherType: "0x0806",
				},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []EthernetFilter{},
		},
		{
			name: "mixed content with non-filter lines",
			input: `ip route default gateway 192.168.1.1
ethernet filter 1 pass * * 0x0800
dhcp scope 1 192.168.1.0/24`,
			expected: []EthernetFilter{
				{
					Number:    1,
					Action:    "pass",
					SourceMAC: "*",
					DestMAC:   "*",
					EtherType: "0x0800",
				},
			},
		},
		{
			name:  "filter with uppercase EtherType",
			input: `ethernet filter 1 pass * * 0x86DD`,
			expected: []EthernetFilter{
				{
					Number:    1,
					Action:    "pass",
					SourceMAC: "*",
					DestMAC:   "*",
					EtherType: "0x86dd",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseEthernetFilterConfig(tt.input)

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
				t.Errorf("expected %d filters, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				got := result[i]
				if got.Number != expected.Number {
					t.Errorf("filter[%d].Number = %d, want %d", i, got.Number, expected.Number)
				}
				if got.Action != expected.Action {
					t.Errorf("filter[%d].Action = %q, want %q", i, got.Action, expected.Action)
				}
				if got.SourceMAC != expected.SourceMAC {
					t.Errorf("filter[%d].SourceMAC = %q, want %q", i, got.SourceMAC, expected.SourceMAC)
				}
				if got.DestMAC != expected.DestMAC {
					t.Errorf("filter[%d].DestMAC = %q, want %q", i, got.DestMAC, expected.DestMAC)
				}
				if got.EtherType != expected.EtherType {
					t.Errorf("filter[%d].EtherType = %q, want %q", i, got.EtherType, expected.EtherType)
				}
				if got.VlanID != expected.VlanID {
					t.Errorf("filter[%d].VlanID = %d, want %d", i, got.VlanID, expected.VlanID)
				}
			}
		})
	}
}

func TestParseSingleEthernetFilter(t *testing.T) {
	input := `ethernet filter 1 pass * * 0x0800
ethernet filter 2 reject 00:11:22:33:44:55 * *
ethernet filter 3 pass * * 0x0806`

	tests := []struct {
		name         string
		filterNumber int
		expected     *EthernetFilter
		wantErr      bool
		errContains  string
	}{
		{
			name:         "find filter 1",
			filterNumber: 1,
			expected: &EthernetFilter{
				Number:    1,
				Action:    "pass",
				SourceMAC: "*",
				DestMAC:   "*",
				EtherType: "0x0800",
			},
		},
		{
			name:         "find filter 2",
			filterNumber: 2,
			expected: &EthernetFilter{
				Number:    2,
				Action:    "reject",
				SourceMAC: "00:11:22:33:44:55",
				DestMAC:   "*",
			},
		},
		{
			name:         "filter not found",
			filterNumber: 99,
			wantErr:      true,
			errContains:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSingleEthernetFilter(input, tt.filterNumber)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Number != tt.expected.Number {
				t.Errorf("Number = %d, want %d", result.Number, tt.expected.Number)
			}
			if result.Action != tt.expected.Action {
				t.Errorf("Action = %q, want %q", result.Action, tt.expected.Action)
			}
			if result.SourceMAC != tt.expected.SourceMAC {
				t.Errorf("SourceMAC = %q, want %q", result.SourceMAC, tt.expected.SourceMAC)
			}
			if result.EtherType != tt.expected.EtherType {
				t.Errorf("EtherType = %q, want %q", result.EtherType, tt.expected.EtherType)
			}
		})
	}
}

func TestBuildEthernetFilterCommand(t *testing.T) {
	tests := []struct {
		name     string
		filter   EthernetFilter
		expected string
	}{
		{
			name: "basic filter",
			filter: EthernetFilter{
				Number:    1,
				Action:    "pass",
				SourceMAC: "*",
				DestMAC:   "*",
			},
			expected: "ethernet filter 1 pass * *",
		},
		{
			name: "filter with EtherType",
			filter: EthernetFilter{
				Number:    1,
				Action:    "pass",
				SourceMAC: "*",
				DestMAC:   "*",
				EtherType: "0x0800",
			},
			expected: "ethernet filter 1 pass * * 0x0800",
		},
		{
			name: "reject with source MAC",
			filter: EthernetFilter{
				Number:    2,
				Action:    "reject",
				SourceMAC: "00:11:22:33:44:55",
				DestMAC:   "*",
			},
			expected: "ethernet filter 2 reject 00:11:22:33:44:55 *",
		},
		{
			name: "filter with VLAN ID",
			filter: EthernetFilter{
				Number:    3,
				Action:    "pass",
				SourceMAC: "*",
				DestMAC:   "*",
				EtherType: "0x0800",
				VlanID:    100,
			},
			expected: "ethernet filter 3 pass * * 0x0800 vlan 100",
		},
		{
			name: "empty MAC uses wildcard",
			filter: EthernetFilter{
				Number: 1,
				Action: "pass",
			},
			expected: "ethernet filter 1 pass * *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildEthernetFilterCommand(tt.filter)
			if result != tt.expected {
				t.Errorf("BuildEthernetFilterCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteEthernetFilterCommand(t *testing.T) {
	tests := []struct {
		name     string
		number   int
		expected string
	}{
		{
			name:     "delete filter 1",
			number:   1,
			expected: "no ethernet filter 1",
		},
		{
			name:     "delete filter 100",
			number:   100,
			expected: "no ethernet filter 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteEthernetFilterCommand(tt.number)
			if result != tt.expected {
				t.Errorf("BuildDeleteEthernetFilterCommand(%d) = %q, want %q", tt.number, result, tt.expected)
			}
		})
	}
}

func TestBuildInterfaceEthernetFilterCommand(t *testing.T) {
	tests := []struct {
		name       string
		iface      string
		direction  string
		filterNums []int
		expected   string
	}{
		{
			name:       "single filter in",
			iface:      "lan1",
			direction:  "in",
			filterNums: []int{1},
			expected:   "ethernet lan1 filter in 1",
		},
		{
			name:       "multiple filters in",
			iface:      "lan1",
			direction:  "in",
			filterNums: []int{1, 2, 3},
			expected:   "ethernet lan1 filter in 1 2 3",
		},
		{
			name:       "filter out",
			iface:      "lan2",
			direction:  "out",
			filterNums: []int{10, 20},
			expected:   "ethernet lan2 filter out 10 20",
		},
		{
			name:       "empty filter list",
			iface:      "lan1",
			direction:  "in",
			filterNums: []int{},
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildInterfaceEthernetFilterCommand(tt.iface, tt.direction, tt.filterNums)
			if result != tt.expected {
				t.Errorf("BuildInterfaceEthernetFilterCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteInterfaceEthernetFilterCommand(t *testing.T) {
	tests := []struct {
		name      string
		iface     string
		direction string
		expected  string
	}{
		{
			name:      "delete in filter",
			iface:     "lan1",
			direction: "in",
			expected:  "no ethernet lan1 filter in",
		},
		{
			name:      "delete out filter",
			iface:     "lan2",
			direction: "out",
			expected:  "no ethernet lan2 filter out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteInterfaceEthernetFilterCommand(tt.iface, tt.direction)
			if result != tt.expected {
				t.Errorf("BuildDeleteInterfaceEthernetFilterCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildShowEthernetFilterCommand(t *testing.T) {
	result := BuildShowEthernetFilterCommand(1)
	expected := `show config | grep "ethernet filter 1"`
	if result != expected {
		t.Errorf("BuildShowEthernetFilterCommand(1) = %q, want %q", result, expected)
	}
}

func TestBuildShowAllEthernetFiltersCommand(t *testing.T) {
	result := BuildShowAllEthernetFiltersCommand()
	expected := `show config | grep "ethernet filter"`
	if result != expected {
		t.Errorf("BuildShowAllEthernetFiltersCommand() = %q, want %q", result, expected)
	}
}

func TestValidateEthernetFilterNumber(t *testing.T) {
	tests := []struct {
		name    string
		number  int
		wantErr bool
	}{
		{
			name:    "valid minimum",
			number:  1,
			wantErr: false,
		},
		{
			name:    "valid maximum",
			number:  65535,
			wantErr: false,
		},
		{
			name:    "valid middle",
			number:  1000,
			wantErr: false,
		},
		{
			name:    "invalid zero",
			number:  0,
			wantErr: true,
		},
		{
			name:    "invalid negative",
			number:  -1,
			wantErr: true,
		},
		{
			name:    "invalid too large",
			number:  65536,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEthernetFilterNumber(tt.number)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for number %d, got nil", tt.number)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for number %d: %v", tt.number, err)
				}
			}
		})
	}
}

func TestValidateMAC(t *testing.T) {
	tests := []struct {
		name    string
		mac     string
		wantErr bool
	}{
		{
			name:    "valid colon-separated",
			mac:     "00:11:22:33:44:55",
			wantErr: false,
		},
		{
			name:    "valid Cisco notation",
			mac:     "0011.2233.4455",
			wantErr: false,
		},
		{
			name:    "valid hyphen-separated",
			mac:     "00-11-22-33-44-55",
			wantErr: false,
		},
		{
			name:    "valid no separator",
			mac:     "001122334455",
			wantErr: false,
		},
		{
			name:    "valid wildcard",
			mac:     "*",
			wantErr: false,
		},
		{
			name:    "valid uppercase",
			mac:     "AA:BB:CC:DD:EE:FF",
			wantErr: false,
		},
		{
			name:    "invalid too short",
			mac:     "00:11:22",
			wantErr: true,
		},
		{
			name:    "invalid too long",
			mac:     "00:11:22:33:44:55:66",
			wantErr: true,
		},
		{
			name:    "invalid non-hex",
			mac:     "00:11:22:33:44:GG",
			wantErr: true,
		},
		{
			name:    "invalid empty",
			mac:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMAC(tt.mac)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for MAC %q, got nil", tt.mac)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for MAC %q: %v", tt.mac, err)
				}
			}
		})
	}
}

func TestValidateEtherType(t *testing.T) {
	tests := []struct {
		name     string
		ethType  string
		wantErr  bool
	}{
		{
			name:    "valid IPv4",
			ethType: "0x0800",
			wantErr: false,
		},
		{
			name:    "valid ARP",
			ethType: "0x0806",
			wantErr: false,
		},
		{
			name:    "valid IPv6",
			ethType: "0x86DD",
			wantErr: false,
		},
		{
			name:    "valid lowercase",
			ethType: "0x86dd",
			wantErr: false,
		},
		{
			name:    "valid wildcard",
			ethType: "*",
			wantErr: false,
		},
		{
			name:    "valid empty",
			ethType: "",
			wantErr: false,
		},
		{
			name:    "valid short",
			ethType: "0x800",
			wantErr: false,
		},
		{
			name:    "invalid no prefix",
			ethType: "0800",
			wantErr: true,
		},
		{
			name:    "invalid non-hex",
			ethType: "0xGGGG",
			wantErr: true,
		},
		{
			name:    "invalid too long",
			ethType: "0x12345",
			wantErr: true,
		},
		{
			name:    "invalid prefix only",
			ethType: "0x",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEtherType(tt.ethType)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for EtherType %q, got nil", tt.ethType)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for EtherType %q: %v", tt.ethType, err)
				}
			}
		})
	}
}

func TestValidateVlanID(t *testing.T) {
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
			id:      4094,
			wantErr: false,
		},
		{
			name:    "valid middle",
			id:      100,
			wantErr: false,
		},
		{
			name:    "valid zero (not specified)",
			id:      0,
			wantErr: false,
		},
		{
			name:    "invalid negative",
			id:      -1,
			wantErr: true,
		},
		{
			name:    "invalid too large",
			id:      4095,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVlanID(tt.id)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for VLAN ID %d, got nil", tt.id)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for VLAN ID %d: %v", tt.id, err)
				}
			}
		})
	}
}

func TestValidateEthernetFilterAction(t *testing.T) {
	tests := []struct {
		name    string
		action  string
		wantErr bool
	}{
		{
			name:    "valid pass",
			action:  "pass",
			wantErr: false,
		},
		{
			name:    "valid reject",
			action:  "reject",
			wantErr: false,
		},
		{
			name:    "valid uppercase PASS",
			action:  "PASS",
			wantErr: false,
		},
		{
			name:    "valid with whitespace",
			action:  "  pass  ",
			wantErr: false,
		},
		{
			name:    "invalid permit",
			action:  "permit",
			wantErr: true,
		},
		{
			name:    "invalid deny",
			action:  "deny",
			wantErr: true,
		},
		{
			name:    "invalid empty",
			action:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEthernetFilterAction(tt.action)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for action %q, got nil", tt.action)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for action %q: %v", tt.action, err)
				}
			}
		})
	}
}

func TestValidateEthernetFilterDirection(t *testing.T) {
	tests := []struct {
		name      string
		direction string
		wantErr   bool
	}{
		{
			name:      "valid in",
			direction: "in",
			wantErr:   false,
		},
		{
			name:      "valid out",
			direction: "out",
			wantErr:   false,
		},
		{
			name:      "valid uppercase IN",
			direction: "IN",
			wantErr:   false,
		},
		{
			name:      "valid with whitespace",
			direction: "  out  ",
			wantErr:   false,
		},
		{
			name:      "invalid inbound",
			direction: "inbound",
			wantErr:   true,
		},
		{
			name:      "invalid empty",
			direction: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEthernetFilterDirection(tt.direction)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for direction %q, got nil", tt.direction)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for direction %q: %v", tt.direction, err)
				}
			}
		})
	}
}

func TestValidateEthernetFilter(t *testing.T) {
	tests := []struct {
		name        string
		filter      EthernetFilter
		wantErr     bool
		errContains string
	}{
		{
			name: "valid filter",
			filter: EthernetFilter{
				Number:    1,
				Action:    "pass",
				SourceMAC: "*",
				DestMAC:   "*",
			},
			wantErr: false,
		},
		{
			name: "valid full filter",
			filter: EthernetFilter{
				Number:    100,
				Action:    "reject",
				SourceMAC: "00:11:22:33:44:55",
				DestMAC:   "ff:ff:ff:ff:ff:ff",
				EtherType: "0x0800",
				VlanID:    100,
			},
			wantErr: false,
		},
		{
			name: "invalid filter number",
			filter: EthernetFilter{
				Number:    0,
				Action:    "pass",
				SourceMAC: "*",
				DestMAC:   "*",
			},
			wantErr:     true,
			errContains: "between 1 and 65535",
		},
		{
			name: "invalid action",
			filter: EthernetFilter{
				Number:    1,
				Action:    "permit",
				SourceMAC: "*",
				DestMAC:   "*",
			},
			wantErr:     true,
			errContains: "pass",
		},
		{
			name: "invalid source MAC",
			filter: EthernetFilter{
				Number:    1,
				Action:    "pass",
				SourceMAC: "invalid",
				DestMAC:   "*",
			},
			wantErr:     true,
			errContains: "source MAC",
		},
		{
			name: "invalid destination MAC",
			filter: EthernetFilter{
				Number:    1,
				Action:    "pass",
				SourceMAC: "*",
				DestMAC:   "invalid",
			},
			wantErr:     true,
			errContains: "destination MAC",
		},
		{
			name: "invalid EtherType",
			filter: EthernetFilter{
				Number:    1,
				Action:    "pass",
				SourceMAC: "*",
				DestMAC:   "*",
				EtherType: "invalid",
			},
			wantErr:     true,
			errContains: "EtherType",
		},
		{
			name: "invalid VLAN ID",
			filter: EthernetFilter{
				Number:    1,
				Action:    "pass",
				SourceMAC: "*",
				DestMAC:   "*",
				VlanID:    5000,
			},
			wantErr:     true,
			errContains: "VLAN ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEthernetFilter(tt.filter)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
