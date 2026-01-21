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
			number:  512,
			wantErr: false,
		},
		{
			name:    "valid middle",
			number:  256,
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
			number:  513,
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
		name    string
		ethType string
		wantErr bool
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
			errContains: "between 1 and 512",
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

func TestParseDHCPBasedEthernetFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []EthernetFilter
	}{
		{
			name:  "dhcp-bind filter without scope",
			input: `ethernet filter 1 pass-log dhcp-bind`,
			expected: []EthernetFilter{
				{
					Number:   1,
					Action:   "pass-log",
					DHCPType: "dhcp-bind",
				},
			},
		},
		{
			name:  "dhcp-not-bind filter without scope",
			input: `ethernet filter 2 reject-nolog dhcp-not-bind`,
			expected: []EthernetFilter{
				{
					Number:   2,
					Action:   "reject-nolog",
					DHCPType: "dhcp-not-bind",
				},
			},
		},
		{
			name:  "dhcp-bind filter with scope",
			input: `ethernet filter 3 pass-nolog dhcp-bind 1`,
			expected: []EthernetFilter{
				{
					Number:    3,
					Action:    "pass-nolog",
					DHCPType:  "dhcp-bind",
					DHCPScope: 1,
				},
			},
		},
		{
			name:  "dhcp-not-bind filter with scope",
			input: `ethernet filter 4 reject-log dhcp-not-bind 2`,
			expected: []EthernetFilter{
				{
					Number:    4,
					Action:    "reject-log",
					DHCPType:  "dhcp-not-bind",
					DHCPScope: 2,
				},
			},
		},
		{
			name: "mixed DHCP and MAC filters",
			input: `ethernet filter 1 pass-log dhcp-bind
ethernet filter 2 reject-nolog 00:11:22:33:44:55 *
ethernet filter 3 pass-nolog dhcp-not-bind 1`,
			expected: []EthernetFilter{
				{
					Number:   1,
					Action:   "pass-log",
					DHCPType: "dhcp-bind",
				},
				{
					Number:         2,
					Action:         "reject-nolog",
					SourceMAC:      "00:11:22:33:44:55",
					DestinationMAC: "*",
					DestMAC:        "*",
				},
				{
					Number:    3,
					Action:    "pass-nolog",
					DHCPType:  "dhcp-not-bind",
					DHCPScope: 1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseEthernetFilterConfig(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d filters, got %d", len(tt.expected), len(result))
			}

			for i, expected := range tt.expected {
				got := result[i]
				if got.Number != expected.Number {
					t.Errorf("filter[%d].Number = %d, want %d", i, got.Number, expected.Number)
				}
				if got.Action != expected.Action {
					t.Errorf("filter[%d].Action = %q, want %q", i, got.Action, expected.Action)
				}
				if got.DHCPType != expected.DHCPType {
					t.Errorf("filter[%d].DHCPType = %q, want %q", i, got.DHCPType, expected.DHCPType)
				}
				if got.DHCPScope != expected.DHCPScope {
					t.Errorf("filter[%d].DHCPScope = %d, want %d", i, got.DHCPScope, expected.DHCPScope)
				}
				if got.SourceMAC != expected.SourceMAC {
					t.Errorf("filter[%d].SourceMAC = %q, want %q", i, got.SourceMAC, expected.SourceMAC)
				}
				if got.DestinationMAC != expected.DestinationMAC {
					t.Errorf("filter[%d].DestinationMAC = %q, want %q", i, got.DestinationMAC, expected.DestinationMAC)
				}
			}
		})
	}
}

func TestBuildDHCPBasedEthernetFilterCommand(t *testing.T) {
	tests := []struct {
		name     string
		filter   EthernetFilter
		expected string
	}{
		{
			name: "dhcp-bind without scope",
			filter: EthernetFilter{
				Number:   1,
				Action:   "pass-log",
				DHCPType: "dhcp-bind",
			},
			expected: "ethernet filter 1 pass-log dhcp-bind",
		},
		{
			name: "dhcp-not-bind without scope",
			filter: EthernetFilter{
				Number:   2,
				Action:   "reject-nolog",
				DHCPType: "dhcp-not-bind",
			},
			expected: "ethernet filter 2 reject-nolog dhcp-not-bind",
		},
		{
			name: "dhcp-bind with scope",
			filter: EthernetFilter{
				Number:    3,
				Action:    "pass-nolog",
				DHCPType:  "dhcp-bind",
				DHCPScope: 1,
			},
			expected: "ethernet filter 3 pass-nolog dhcp-bind 1",
		},
		{
			name: "dhcp-not-bind with scope",
			filter: EthernetFilter{
				Number:    4,
				Action:    "reject-log",
				DHCPType:  "dhcp-not-bind",
				DHCPScope: 2,
			},
			expected: "ethernet filter 4 reject-log dhcp-not-bind 2",
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

func TestParseOffsetBasedEthernetFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []EthernetFilter
	}{
		{
			name:  "filter with offset and byte_list",
			input: `ethernet filter 1 pass-log * * offset=14 0x08 0x00`,
			expected: []EthernetFilter{
				{
					Number:         1,
					Action:         "pass-log",
					SourceMAC:      "*",
					DestinationMAC: "*",
					DestMAC:        "*",
					Offset:         14,
					ByteList:       []string{"0x08", "0x00"},
				},
			},
		},
		{
			name:  "filter with offset and multiple bytes",
			input: `ethernet filter 2 reject-nolog 00:11:22:33:44:55 * offset=20 0xff 0xff 0xff 0xff`,
			expected: []EthernetFilter{
				{
					Number:         2,
					Action:         "reject-nolog",
					SourceMAC:      "00:11:22:33:44:55",
					DestinationMAC: "*",
					DestMAC:        "*",
					Offset:         20,
					ByteList:       []string{"0xff", "0xff", "0xff", "0xff"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseEthernetFilterConfig(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d filters, got %d", len(tt.expected), len(result))
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
				if got.DestinationMAC != expected.DestinationMAC {
					t.Errorf("filter[%d].DestinationMAC = %q, want %q", i, got.DestinationMAC, expected.DestinationMAC)
				}
				if got.Offset != expected.Offset {
					t.Errorf("filter[%d].Offset = %d, want %d", i, got.Offset, expected.Offset)
				}
				if len(got.ByteList) != len(expected.ByteList) {
					t.Errorf("filter[%d].ByteList length = %d, want %d", i, len(got.ByteList), len(expected.ByteList))
				} else {
					for j, b := range expected.ByteList {
						if got.ByteList[j] != b {
							t.Errorf("filter[%d].ByteList[%d] = %q, want %q", i, j, got.ByteList[j], b)
						}
					}
				}
			}
		})
	}
}

func TestBuildOffsetBasedEthernetFilterCommand(t *testing.T) {
	tests := []struct {
		name     string
		filter   EthernetFilter
		expected string
	}{
		{
			name: "filter with offset and byte_list",
			filter: EthernetFilter{
				Number:         1,
				Action:         "pass-log",
				SourceMAC:      "*",
				DestinationMAC: "*",
				Offset:         14,
				ByteList:       []string{"0x08", "0x00"},
			},
			expected: "ethernet filter 1 pass-log * * offset=14 0x08 0x00",
		},
		{
			name: "filter with offset and multiple bytes",
			filter: EthernetFilter{
				Number:         2,
				Action:         "reject-nolog",
				SourceMAC:      "00:11:22:33:44:55",
				DestinationMAC: "*",
				Offset:         20,
				ByteList:       []string{"0xff", "0xff", "0xff", "0xff"},
			},
			expected: "ethernet filter 2 reject-nolog 00:11:22:33:44:55 * offset=20 0xff 0xff 0xff 0xff",
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

func TestValidateMACAddress(t *testing.T) {
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
			name:    "valid wildcard",
			mac:     "*",
			wantErr: false,
		},
		{
			name:    "valid empty",
			mac:     "",
			wantErr: false,
		},
		{
			name:    "valid uppercase",
			mac:     "AA:BB:CC:DD:EE:FF",
			wantErr: false,
		},
		{
			name:    "invalid Cisco notation",
			mac:     "0011.2233.4455",
			wantErr: true,
		},
		{
			name:    "invalid hyphen-separated",
			mac:     "00-11-22-33-44-55",
			wantErr: true,
		},
		{
			name:    "invalid no separator",
			mac:     "001122334455",
			wantErr: true,
		},
		{
			name:    "invalid too short",
			mac:     "00:11:22",
			wantErr: true,
		},
		{
			name:    "invalid non-hex",
			mac:     "00:11:22:33:44:GG",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMACAddress(tt.mac)
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

func TestValidateDHCPType(t *testing.T) {
	tests := []struct {
		name     string
		dhcpType string
		wantErr  bool
	}{
		{
			name:     "valid dhcp-bind",
			dhcpType: "dhcp-bind",
			wantErr:  false,
		},
		{
			name:     "valid dhcp-not-bind",
			dhcpType: "dhcp-not-bind",
			wantErr:  false,
		},
		{
			name:     "valid empty",
			dhcpType: "",
			wantErr:  false,
		},
		{
			name:     "invalid dhcp-bound",
			dhcpType: "dhcp-bound",
			wantErr:  true,
		},
		{
			name:     "invalid other",
			dhcpType: "other",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDHCPType(tt.dhcpType)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for DHCP type %q, got nil", tt.dhcpType)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for DHCP type %q: %v", tt.dhcpType, err)
				}
			}
		})
	}
}

func TestValidateEthernetFilterWithNewActions(t *testing.T) {
	tests := []struct {
		name    string
		filter  EthernetFilter
		wantErr bool
	}{
		{
			name: "valid pass-log action",
			filter: EthernetFilter{
				Number:   1,
				Action:   "pass-log",
				DHCPType: "dhcp-bind",
			},
			wantErr: false,
		},
		{
			name: "valid pass-nolog action",
			filter: EthernetFilter{
				Number:   2,
				Action:   "pass-nolog",
				DHCPType: "dhcp-not-bind",
			},
			wantErr: false,
		},
		{
			name: "valid reject-log action",
			filter: EthernetFilter{
				Number:    3,
				Action:    "reject-log",
				DHCPType:  "dhcp-bind",
				DHCPScope: 1,
			},
			wantErr: false,
		},
		{
			name: "valid reject-nolog action",
			filter: EthernetFilter{
				Number:    4,
				Action:    "reject-nolog",
				DHCPType:  "dhcp-not-bind",
				DHCPScope: 2,
			},
			wantErr: false,
		},
		{
			name: "valid DHCP filter with scope",
			filter: EthernetFilter{
				Number:    5,
				Action:    "pass-log",
				DHCPType:  "dhcp-bind",
				DHCPScope: 10,
			},
			wantErr: false,
		},
		{
			name: "valid filter with offset and byte_list",
			filter: EthernetFilter{
				Number:         6,
				Action:         "pass-nolog",
				SourceMAC:      "*",
				DestinationMAC: "*",
				Offset:         14,
				ByteList:       []string{"0x08", "0x00"},
			},
			wantErr: false,
		},
		{
			name: "invalid filter number too large",
			filter: EthernetFilter{
				Number:   513,
				Action:   "pass-log",
				DHCPType: "dhcp-bind",
			},
			wantErr: true,
		},
		{
			name: "invalid DHCP filter with source MAC",
			filter: EthernetFilter{
				Number:    1,
				Action:    "pass-log",
				DHCPType:  "dhcp-bind",
				SourceMAC: "00:11:22:33:44:55",
			},
			wantErr: true,
		},
		{
			name: "invalid offset without byte_list",
			filter: EthernetFilter{
				Number:         1,
				Action:         "pass-log",
				SourceMAC:      "*",
				DestinationMAC: "*",
				Offset:         14,
			},
			wantErr: true,
		},
		{
			name: "invalid byte_list without offset",
			filter: EthernetFilter{
				Number:         1,
				Action:         "pass-log",
				SourceMAC:      "*",
				DestinationMAC: "*",
				ByteList:       []string{"0x08", "0x00"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEthernetFilter(tt.filter)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestParseEthernetFilterConfigWithNewActions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []EthernetFilter
	}{
		{
			name:  "pass-log action",
			input: `ethernet filter 1 pass-log * * 0x0800`,
			expected: []EthernetFilter{
				{
					Number:         1,
					Action:         "pass-log",
					SourceMAC:      "*",
					DestinationMAC: "*",
					DestMAC:        "*",
					EtherType:      "0x0800",
				},
			},
		},
		{
			name:  "pass-nolog action",
			input: `ethernet filter 2 pass-nolog 00:11:22:33:44:55 *`,
			expected: []EthernetFilter{
				{
					Number:         2,
					Action:         "pass-nolog",
					SourceMAC:      "00:11:22:33:44:55",
					DestinationMAC: "*",
					DestMAC:        "*",
				},
			},
		},
		{
			name:  "reject-log action",
			input: `ethernet filter 3 reject-log * ff:ff:ff:ff:ff:ff`,
			expected: []EthernetFilter{
				{
					Number:         3,
					Action:         "reject-log",
					SourceMAC:      "*",
					DestinationMAC: "ff:ff:ff:ff:ff:ff",
					DestMAC:        "ff:ff:ff:ff:ff:ff",
				},
			},
		},
		{
			name:  "reject-nolog action",
			input: `ethernet filter 4 reject-nolog 00:11:22:33:44:55 ff:ff:ff:ff:ff:ff 0x0806`,
			expected: []EthernetFilter{
				{
					Number:         4,
					Action:         "reject-nolog",
					SourceMAC:      "00:11:22:33:44:55",
					DestinationMAC: "ff:ff:ff:ff:ff:ff",
					DestMAC:        "ff:ff:ff:ff:ff:ff",
					EtherType:      "0x0806",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseEthernetFilterConfig(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d filters, got %d", len(tt.expected), len(result))
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
				if got.DestinationMAC != expected.DestinationMAC {
					t.Errorf("filter[%d].DestinationMAC = %q, want %q", i, got.DestinationMAC, expected.DestinationMAC)
				}
				if got.EtherType != expected.EtherType {
					t.Errorf("filter[%d].EtherType = %q, want %q", i, got.EtherType, expected.EtherType)
				}
			}
		})
	}
}

func TestBuildEthernetFilterCommandWithNewActions(t *testing.T) {
	tests := []struct {
		name     string
		filter   EthernetFilter
		expected string
	}{
		{
			name: "pass-log with EtherType",
			filter: EthernetFilter{
				Number:         1,
				Action:         "pass-log",
				SourceMAC:      "*",
				DestinationMAC: "*",
				EtherType:      "0x0800",
			},
			expected: "ethernet filter 1 pass-log * * 0x0800",
		},
		{
			name: "pass-nolog with source MAC",
			filter: EthernetFilter{
				Number:         2,
				Action:         "pass-nolog",
				SourceMAC:      "00:11:22:33:44:55",
				DestinationMAC: "*",
			},
			expected: "ethernet filter 2 pass-nolog 00:11:22:33:44:55 *",
		},
		{
			name: "reject-log with dest MAC",
			filter: EthernetFilter{
				Number:         3,
				Action:         "reject-log",
				SourceMAC:      "*",
				DestinationMAC: "ff:ff:ff:ff:ff:ff",
			},
			expected: "ethernet filter 3 reject-log * ff:ff:ff:ff:ff:ff",
		},
		{
			name: "reject-nolog with all options",
			filter: EthernetFilter{
				Number:         4,
				Action:         "reject-nolog",
				SourceMAC:      "00:11:22:33:44:55",
				DestinationMAC: "ff:ff:ff:ff:ff:ff",
				EtherType:      "0x0806",
				VlanID:         100,
			},
			expected: "ethernet filter 4 reject-nolog 00:11:22:33:44:55 ff:ff:ff:ff:ff:ff 0x0806 vlan 100",
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

func TestValidateEthernetFilterNumber512(t *testing.T) {
	tests := []struct {
		name    string
		number  int
		wantErr bool
	}{
		{
			name:    "valid 512",
			number:  512,
			wantErr: false,
		},
		{
			name:    "invalid 513",
			number:  513,
			wantErr: true,
		},
		{
			name:    "invalid 65535 (old max)",
			number:  65535,
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

func TestBuildAccessListMACEntryCommand(t *testing.T) {
	tests := []struct {
		name     string
		entry    AccessListMACEntry
		expected string
	}{
		{
			name: "permit maps to pass with offset/bytes",
			entry: AccessListMACEntry{
				Sequence:  5,
				AceAction: "permit",
				SourceAny: true,
				Offset:    14,
				ByteList:  []string{"0x08", "0x00"},
			},
			expected: "ethernet filter 5 pass * * offset=14 0x08 0x00",
		},
		{
			name: "explicit filter_id with pass-log and dhcp scope",
			entry: AccessListMACEntry{
				FilterID:  10,
				AceAction: "pass-log",
				SourceAny: true,
				DHCPType:  "dhcp-bind",
				DHCPScope: 2,
			},
			expected: "ethernet filter 10 pass-log dhcp-bind 2",
		},
		{
			name: "reject-log keeps action and dest mac",
			entry: AccessListMACEntry{
				Sequence:           7,
				AceAction:          "reject-log",
				SourceAny:          true,
				DestinationAny:     false,
				DestinationAddress: "ff:ff:ff:ff:ff:ff",
				EtherType:          "0x0806",
			},
			expected: "ethernet filter 7 reject-log * ff:ff:ff:ff:ff:ff 0x0806",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BuildAccessListMACEntryCommand(tt.entry); got != tt.expected {
				t.Errorf("BuildAccessListMACEntryCommand() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseEthernetFilterApplication(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []EthernetFilterApplication
		wantErr  bool
	}{
		{
			name:  "single filter in",
			input: `ethernet lan1 filter in 1`,
			expected: []EthernetFilterApplication{
				{
					Interface: "lan1",
					Direction: "in",
					Filters:   []int{1},
				},
			},
		},
		{
			name:  "multiple filters in",
			input: `ethernet lan1 filter in 1 100`,
			expected: []EthernetFilterApplication{
				{
					Interface: "lan1",
					Direction: "in",
					Filters:   []int{1, 100},
				},
			},
		},
		{
			name:  "filter out",
			input: `ethernet lan2 filter out 10 20 30`,
			expected: []EthernetFilterApplication{
				{
					Interface: "lan2",
					Direction: "out",
					Filters:   []int{10, 20, 30},
				},
			},
		},
		{
			name: "multiple interfaces and directions",
			input: `ethernet lan1 filter in 1 100
ethernet lan1 filter out 2 200
ethernet lan2 filter in 10 20 30`,
			expected: []EthernetFilterApplication{
				{
					Interface: "lan1",
					Direction: "in",
					Filters:   []int{1, 100},
				},
				{
					Interface: "lan1",
					Direction: "out",
					Filters:   []int{2, 200},
				},
				{
					Interface: "lan2",
					Direction: "in",
					Filters:   []int{10, 20, 30},
				},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []EthernetFilterApplication{},
		},
		{
			name: "mixed content with non-filter lines",
			input: `ip route default gateway 192.168.1.1
ethernet lan1 filter in 1 100
dhcp scope 1 192.168.1.0/24`,
			expected: []EthernetFilterApplication{
				{
					Interface: "lan1",
					Direction: "in",
					Filters:   []int{1, 100},
				},
			},
		},
		{
			name:  "many filters",
			input: `ethernet lan1 filter in 1 2 3 4 5 6 7 8 9 10`,
			expected: []EthernetFilterApplication{
				{
					Interface: "lan1",
					Direction: "in",
					Filters:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				},
			},
		},
		{
			name:  "max filter number",
			input: `ethernet lan1 filter in 512`,
			expected: []EthernetFilterApplication{
				{
					Interface: "lan1",
					Direction: "in",
					Filters:   []int{512},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseEthernetFilterApplication(tt.input)

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
				t.Errorf("expected %d applications, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				got := result[i]
				if got.Interface != expected.Interface {
					t.Errorf("app[%d].Interface = %q, want %q", i, got.Interface, expected.Interface)
				}
				if got.Direction != expected.Direction {
					t.Errorf("app[%d].Direction = %q, want %q", i, got.Direction, expected.Direction)
				}
				if len(got.Filters) != len(expected.Filters) {
					t.Errorf("app[%d].Filters length = %d, want %d", i, len(got.Filters), len(expected.Filters))
					continue
				}
				for j, f := range expected.Filters {
					if got.Filters[j] != f {
						t.Errorf("app[%d].Filters[%d] = %d, want %d", i, j, got.Filters[j], f)
					}
				}
			}
		})
	}
}

func TestParseSingleEthernetFilterApplication(t *testing.T) {
	input := `ethernet lan1 filter in 1 100
ethernet lan1 filter out 2 200
ethernet lan2 filter in 10 20 30`

	tests := []struct {
		name      string
		iface     string
		direction string
		expected  *EthernetFilterApplication
	}{
		{
			name:      "find lan1 in",
			iface:     "lan1",
			direction: "in",
			expected: &EthernetFilterApplication{
				Interface: "lan1",
				Direction: "in",
				Filters:   []int{1, 100},
			},
		},
		{
			name:      "find lan1 out",
			iface:     "lan1",
			direction: "out",
			expected: &EthernetFilterApplication{
				Interface: "lan1",
				Direction: "out",
				Filters:   []int{2, 200},
			},
		},
		{
			name:      "find lan2 in",
			iface:     "lan2",
			direction: "in",
			expected: &EthernetFilterApplication{
				Interface: "lan2",
				Direction: "in",
				Filters:   []int{10, 20, 30},
			},
		},
		{
			name:      "not found - wrong interface",
			iface:     "lan3",
			direction: "in",
			expected:  nil,
		},
		{
			name:      "not found - wrong direction",
			iface:     "lan2",
			direction: "out",
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSingleEthernetFilterApplication(input, tt.iface, tt.direction)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("expected %+v, got nil", tt.expected)
				return
			}

			if result.Interface != tt.expected.Interface {
				t.Errorf("Interface = %q, want %q", result.Interface, tt.expected.Interface)
			}
			if result.Direction != tt.expected.Direction {
				t.Errorf("Direction = %q, want %q", result.Direction, tt.expected.Direction)
			}
			if len(result.Filters) != len(tt.expected.Filters) {
				t.Errorf("Filters length = %d, want %d", len(result.Filters), len(tt.expected.Filters))
				return
			}
			for i, f := range tt.expected.Filters {
				if result.Filters[i] != f {
					t.Errorf("Filters[%d] = %d, want %d", i, result.Filters[i], f)
				}
			}
		})
	}
}

func TestBuildEthernetFilterApplicationCommand(t *testing.T) {
	tests := []struct {
		name     string
		app      EthernetFilterApplication
		expected string
	}{
		{
			name: "single filter in",
			app: EthernetFilterApplication{
				Interface: "lan1",
				Direction: "in",
				Filters:   []int{1},
			},
			expected: "ethernet lan1 filter in 1",
		},
		{
			name: "multiple filters in",
			app: EthernetFilterApplication{
				Interface: "lan1",
				Direction: "in",
				Filters:   []int{1, 100},
			},
			expected: "ethernet lan1 filter in 1 100",
		},
		{
			name: "filter out",
			app: EthernetFilterApplication{
				Interface: "lan2",
				Direction: "out",
				Filters:   []int{10, 20},
			},
			expected: "ethernet lan2 filter out 10 20",
		},
		{
			name: "empty filters returns empty string",
			app: EthernetFilterApplication{
				Interface: "lan1",
				Direction: "in",
				Filters:   []int{},
			},
			expected: "",
		},
		{
			name: "many filters",
			app: EthernetFilterApplication{
				Interface: "lan1",
				Direction: "in",
				Filters:   []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			},
			expected: "ethernet lan1 filter in 1 2 3 4 5 6 7 8 9 10",
		},
		{
			name: "max filter number",
			app: EthernetFilterApplication{
				Interface: "lan1",
				Direction: "in",
				Filters:   []int{512},
			},
			expected: "ethernet lan1 filter in 512",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildEthernetFilterApplicationCommand(tt.app)
			if result != tt.expected {
				t.Errorf("BuildEthernetFilterApplicationCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteEthernetFilterApplicationCommand(t *testing.T) {
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
			result := BuildDeleteEthernetFilterApplicationCommand(tt.iface, tt.direction)
			if result != tt.expected {
				t.Errorf("BuildDeleteEthernetFilterApplicationCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValidateEthernetFilterApplication(t *testing.T) {
	tests := []struct {
		name        string
		app         EthernetFilterApplication
		wantErr     bool
		errContains string
	}{
		{
			name: "valid single filter",
			app: EthernetFilterApplication{
				Interface: "lan1",
				Direction: "in",
				Filters:   []int{1},
			},
			wantErr: false,
		},
		{
			name: "valid multiple filters",
			app: EthernetFilterApplication{
				Interface: "lan1",
				Direction: "out",
				Filters:   []int{1, 100, 512},
			},
			wantErr: false,
		},
		{
			name: "valid empty filters",
			app: EthernetFilterApplication{
				Interface: "lan1",
				Direction: "in",
				Filters:   []int{},
			},
			wantErr: false,
		},
		{
			name: "invalid empty interface",
			app: EthernetFilterApplication{
				Interface: "",
				Direction: "in",
				Filters:   []int{1},
			},
			wantErr:     true,
			errContains: "interface name is required",
		},
		{
			name: "invalid direction",
			app: EthernetFilterApplication{
				Interface: "lan1",
				Direction: "both",
				Filters:   []int{1},
			},
			wantErr:     true,
			errContains: "direction",
		},
		{
			name: "invalid filter number zero",
			app: EthernetFilterApplication{
				Interface: "lan1",
				Direction: "in",
				Filters:   []int{0},
			},
			wantErr:     true,
			errContains: "between 1 and 512",
		},
		{
			name: "invalid filter number too large",
			app: EthernetFilterApplication{
				Interface: "lan1",
				Direction: "in",
				Filters:   []int{513},
			},
			wantErr:     true,
			errContains: "between 1 and 512",
		},
		{
			name: "invalid filter number in list",
			app: EthernetFilterApplication{
				Interface: "lan1",
				Direction: "in",
				Filters:   []int{1, 100, 1000},
			},
			wantErr:     true,
			errContains: "between 1 and 512",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEthernetFilterApplication(tt.app)
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

func TestParseInterfaceEthernetFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]map[string][]int
	}{
		{
			name:  "single interface in filter",
			input: `ethernet lan1 filter in 1 100`,
			expected: map[string]map[string][]int{
				"lan1": {
					"in": {1, 100},
				},
			},
		},
		{
			name: "multiple interfaces and directions",
			input: `ethernet lan1 filter in 1 100
ethernet lan1 filter out 2 200
ethernet lan2 filter in 10 20`,
			expected: map[string]map[string][]int{
				"lan1": {
					"in":  {1, 100},
					"out": {2, 200},
				},
				"lan2": {
					"in": {10, 20},
				},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: map[string]map[string][]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseInterfaceEthernetFilter(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d interfaces, got %d", len(tt.expected), len(result))
				return
			}

			for iface, dirs := range tt.expected {
				gotDirs, ok := result[iface]
				if !ok {
					t.Errorf("interface %q not found in result", iface)
					continue
				}
				for dir, filters := range dirs {
					gotFilters, ok := gotDirs[dir]
					if !ok {
						t.Errorf("interface %q direction %q not found in result", iface, dir)
						continue
					}
					if len(gotFilters) != len(filters) {
						t.Errorf("interface %q direction %q: expected %d filters, got %d", iface, dir, len(filters), len(gotFilters))
						continue
					}
					for i, f := range filters {
						if gotFilters[i] != f {
							t.Errorf("interface %q direction %q filter[%d]: expected %d, got %d", iface, dir, i, f, gotFilters[i])
						}
					}
				}
			}
		})
	}
}
