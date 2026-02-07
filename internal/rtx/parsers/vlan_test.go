package parsers

import (
	"strings"
	"testing"
)

func TestParseVLANConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []VLAN
		wantErr  bool
	}{
		{
			name:  "basic VLAN",
			input: `vlan lan1/1 802.1q vid=10`,
			expected: []VLAN{
				{
					VlanID:        10,
					Interface:     "lan1",
					VlanInterface: "lan1/1",
					Shutdown:      false,
				},
			},
		},
		{
			name: "VLAN with IP address CIDR",
			input: `vlan lan1/1 802.1q vid=10
ip lan1/1 address 192.168.10.1/24`,
			expected: []VLAN{
				{
					VlanID:        10,
					Interface:     "lan1",
					VlanInterface: "lan1/1",
					IPAddress:     "192.168.10.1",
					IPMask:        "255.255.255.0",
					Shutdown:      false,
				},
			},
		},
		{
			name: "VLAN with IP address dotted mask",
			input: `vlan lan1/1 802.1q vid=10
ip lan1/1 address 192.168.10.1 255.255.255.0`,
			expected: []VLAN{
				{
					VlanID:        10,
					Interface:     "lan1",
					VlanInterface: "lan1/1",
					IPAddress:     "192.168.10.1",
					IPMask:        "255.255.255.0",
					Shutdown:      false,
				},
			},
		},
		{
			name: "VLAN with description",
			input: `vlan lan1/1 802.1q vid=10
description lan1/1 Management VLAN`,
			expected: []VLAN{
				{
					VlanID:        10,
					Name:          "Management VLAN",
					Interface:     "lan1",
					VlanInterface: "lan1/1",
					Shutdown:      false,
				},
			},
		},
		{
			name: "VLAN shutdown",
			input: `vlan lan1/1 802.1q vid=10
no lan1/1 enable`,
			expected: []VLAN{
				{
					VlanID:        10,
					Interface:     "lan1",
					VlanInterface: "lan1/1",
					Shutdown:      true,
				},
			},
		},
		{
			name: "VLAN enabled explicitly",
			input: `vlan lan1/1 802.1q vid=10
lan1/1 enable`,
			expected: []VLAN{
				{
					VlanID:        10,
					Interface:     "lan1",
					VlanInterface: "lan1/1",
					Shutdown:      false,
				},
			},
		},
		{
			name: "full VLAN configuration",
			input: `vlan lan1/1 802.1q vid=10
ip lan1/1 address 192.168.10.1/24
description lan1/1 Management VLAN
lan1/1 enable`,
			expected: []VLAN{
				{
					VlanID:        10,
					Name:          "Management VLAN",
					Interface:     "lan1",
					VlanInterface: "lan1/1",
					IPAddress:     "192.168.10.1",
					IPMask:        "255.255.255.0",
					Shutdown:      false,
				},
			},
		},
		{
			name: "multiple VLANs",
			input: `vlan lan1/1 802.1q vid=10
ip lan1/1 address 192.168.10.1/24
description lan1/1 Management
vlan lan1/2 802.1q vid=20
ip lan1/2 address 192.168.20.1/24
description lan1/2 Users`,
			expected: []VLAN{
				{
					VlanID:        10,
					Name:          "Management",
					Interface:     "lan1",
					VlanInterface: "lan1/1",
					IPAddress:     "192.168.10.1",
					IPMask:        "255.255.255.0",
					Shutdown:      false,
				},
				{
					VlanID:        20,
					Name:          "Users",
					Interface:     "lan1",
					VlanInterface: "lan1/2",
					IPAddress:     "192.168.20.1",
					IPMask:        "255.255.255.0",
					Shutdown:      false,
				},
			},
		},
		{
			name: "VLANs on different interfaces",
			input: `vlan lan1/1 802.1q vid=10
vlan lan2/1 802.1q vid=20`,
			expected: []VLAN{
				{
					VlanID:        10,
					Interface:     "lan1",
					VlanInterface: "lan1/1",
					Shutdown:      false,
				},
				{
					VlanID:        20,
					Interface:     "lan2",
					VlanInterface: "lan2/1",
					Shutdown:      false,
				},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []VLAN{},
		},
		{
			name:     "no VLAN config",
			input:    "some other config\nip lan1 address 192.168.1.1/24",
			expected: []VLAN{},
		},
	}

	parser := NewVLANParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseVLANConfig(tt.input)

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
				t.Errorf("expected %d VLANs, got %d", len(tt.expected), len(result))
				return
			}

			// Create a map for easier comparison (order may vary)
			resultMap := make(map[string]VLAN)
			for _, v := range result {
				resultMap[v.VlanInterface] = v
			}

			for _, expected := range tt.expected {
				got, ok := resultMap[expected.VlanInterface]
				if !ok {
					t.Errorf("VLAN %s not found in result", expected.VlanInterface)
					continue
				}

				if got.VlanID != expected.VlanID {
					t.Errorf("%s: vlan_id = %d, want %d", expected.VlanInterface, got.VlanID, expected.VlanID)
				}
				if got.Name != expected.Name {
					t.Errorf("%s: name = %q, want %q", expected.VlanInterface, got.Name, expected.Name)
				}
				if got.Interface != expected.Interface {
					t.Errorf("%s: interface = %q, want %q", expected.VlanInterface, got.Interface, expected.Interface)
				}
				if got.IPAddress != expected.IPAddress {
					t.Errorf("%s: ip_address = %q, want %q", expected.VlanInterface, got.IPAddress, expected.IPAddress)
				}
				if got.IPMask != expected.IPMask {
					t.Errorf("%s: ip_mask = %q, want %q", expected.VlanInterface, got.IPMask, expected.IPMask)
				}
				if got.Shutdown != expected.Shutdown {
					t.Errorf("%s: shutdown = %v, want %v", expected.VlanInterface, got.Shutdown, expected.Shutdown)
				}
			}
		})
	}
}

func TestBuildVLANCommand(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		slot     int
		vlanID   int
		expected string
	}{
		{
			name:     "basic VLAN",
			iface:    "lan1",
			slot:     1,
			vlanID:   10,
			expected: "vlan lan1/1 802.1q vid=10",
		},
		{
			name:     "VLAN on lan2",
			iface:    "lan2",
			slot:     1,
			vlanID:   100,
			expected: "vlan lan2/1 802.1q vid=100",
		},
		{
			name:     "second slot",
			iface:    "lan1",
			slot:     2,
			vlanID:   20,
			expected: "vlan lan1/2 802.1q vid=20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildVLANCommand(tt.iface, tt.slot, tt.vlanID)
			if result != tt.expected {
				t.Errorf("BuildVLANCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildVLANIPCommand(t *testing.T) {
	tests := []struct {
		name          string
		vlanInterface string
		ipAddr        string
		mask          string
		expected      string
	}{
		{
			name:          "24-bit mask",
			vlanInterface: "lan1/1",
			ipAddr:        "192.168.10.1",
			mask:          "255.255.255.0",
			expected:      "ip lan1/1 address 192.168.10.1/24",
		},
		{
			name:          "16-bit mask",
			vlanInterface: "lan1/2",
			ipAddr:        "172.16.0.1",
			mask:          "255.255.0.0",
			expected:      "ip lan1/2 address 172.16.0.1/16",
		},
		{
			name:          "28-bit mask",
			vlanInterface: "lan2/1",
			ipAddr:        "10.0.0.1",
			mask:          "255.255.255.240",
			expected:      "ip lan2/1 address 10.0.0.1/28",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildVLANIPCommand(tt.vlanInterface, tt.ipAddr, tt.mask)
			if result != tt.expected {
				t.Errorf("BuildVLANIPCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildVLANDescriptionCommand(t *testing.T) {
	result := BuildVLANDescriptionCommand("lan1/1", "Management VLAN")
	expected := "description lan1/1 Management VLAN"
	if result != expected {
		t.Errorf("BuildVLANDescriptionCommand() = %q, want %q", result, expected)
	}
}

func TestBuildVLANEnableDisableCommand(t *testing.T) {
	enableResult := BuildVLANEnableCommand("lan1/1")
	expectedEnable := "lan1/1 enable"
	if enableResult != expectedEnable {
		t.Errorf("BuildVLANEnableCommand() = %q, want %q", enableResult, expectedEnable)
	}

	disableResult := BuildVLANDisableCommand("lan1/1")
	expectedDisable := "no lan1/1 enable"
	if disableResult != expectedDisable {
		t.Errorf("BuildVLANDisableCommand() = %q, want %q", disableResult, expectedDisable)
	}
}

func TestBuildDeleteVLANCommand(t *testing.T) {
	result := BuildDeleteVLANCommand("lan1/1")
	expected := "no vlan lan1/1"
	if result != expected {
		t.Errorf("BuildDeleteVLANCommand() = %q, want %q", result, expected)
	}
}

func TestBuildShowVLANCommand(t *testing.T) {
	result := BuildShowVLANCommand("lan1", 10)
	expected := `show config | grep "vlan lan1.*vid=10\|lan1/"`
	if result != expected {
		t.Errorf("BuildShowVLANCommand() = %q, want %q", result, expected)
	}
}

func TestValidateVLAN(t *testing.T) {
	tests := []struct {
		name    string
		vlan    VLAN
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid VLAN",
			vlan: VLAN{
				VlanID:    10,
				Interface: "lan1",
			},
			wantErr: false,
		},
		{
			name: "valid full VLAN",
			vlan: VLAN{
				VlanID:    10,
				Name:      "Management",
				Interface: "lan1",
				IPAddress: "192.168.10.1",
				IPMask:    "255.255.255.0",
			},
			wantErr: false,
		},
		{
			name: "VLAN ID too low (0)",
			vlan: VLAN{
				VlanID:    0,
				Interface: "lan1",
			},
			wantErr: true,
			errMsg:  "vlan_id must be 2-4094",
		},
		{
			name: "VLAN ID 1 is reserved",
			vlan: VLAN{
				VlanID:    1,
				Interface: "lan1",
			},
			wantErr: true,
			errMsg:  "vlan_id must be 2-4094",
		},
		{
			name: "VLAN ID too high",
			vlan: VLAN{
				VlanID:    4095,
				Interface: "lan1",
			},
			wantErr: true,
			errMsg:  "vlan_id must be 2-4094",
		},
		{
			name: "empty interface",
			vlan: VLAN{
				VlanID:    10,
				Interface: "",
			},
			wantErr: true,
			errMsg:  "interface is required",
		},
		{
			name: "invalid interface format",
			vlan: VLAN{
				VlanID:    10,
				Interface: "eth0",
			},
			wantErr: true,
			errMsg:  "interface must be in format 'lanN'",
		},
		{
			name: "invalid IP address",
			vlan: VLAN{
				VlanID:    10,
				Interface: "lan1",
				IPAddress: "invalid",
				IPMask:    "255.255.255.0",
			},
			wantErr: true,
			errMsg:  "invalid IP address",
		},
		{
			name: "IP without mask",
			vlan: VLAN{
				VlanID:    10,
				Interface: "lan1",
				IPAddress: "192.168.10.1",
			},
			wantErr: true,
			errMsg:  "ip_mask is required when ip_address is specified",
		},
		{
			name: "mask without IP",
			vlan: VLAN{
				VlanID:    10,
				Interface: "lan1",
				IPMask:    "255.255.255.0",
			},
			wantErr: true,
			errMsg:  "ip_address is required when ip_mask is specified",
		},
		{
			name: "invalid mask",
			vlan: VLAN{
				VlanID:    10,
				Interface: "lan1",
				IPAddress: "192.168.10.1",
				IPMask:    "invalid",
			},
			wantErr: true,
			errMsg:  "invalid IP mask",
		},
		{
			name: "valid VLAN ID 2 (minimum)",
			vlan: VLAN{
				VlanID:    2,
				Interface: "lan1",
			},
			wantErr: false,
		},
		{
			name: "valid VLAN ID 4094 (maximum)",
			vlan: VLAN{
				VlanID:    4094,
				Interface: "lan1",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVLAN(tt.vlan)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
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

func TestPrefixToMask(t *testing.T) {
	tests := []struct {
		prefix   int
		expected string
	}{
		{0, "0.0.0.0"},
		{8, "255.0.0.0"},
		{16, "255.255.0.0"},
		{24, "255.255.255.0"},
		{28, "255.255.255.240"},
		{32, "255.255.255.255"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := prefixToMask(tt.prefix)
			if result != tt.expected {
				t.Errorf("prefixToMask(%d) = %q, want %q", tt.prefix, result, tt.expected)
			}
		})
	}
}

func TestMaskToPrefix(t *testing.T) {
	tests := []struct {
		mask     string
		expected int
	}{
		{"0.0.0.0", 0},
		{"255.0.0.0", 8},
		{"255.255.0.0", 16},
		{"255.255.255.0", 24},
		{"255.255.255.240", 28},
		{"255.255.255.255", 32},
	}

	for _, tt := range tests {
		t.Run(tt.mask, func(t *testing.T) {
			result := maskToPrefix(tt.mask)
			if result != tt.expected {
				t.Errorf("maskToPrefix(%q) = %d, want %d", tt.mask, result, tt.expected)
			}
		})
	}
}

func TestFindNextAvailableSlot(t *testing.T) {
	tests := []struct {
		name          string
		existingVLANs []VLAN
		iface         string
		expected      int
	}{
		{
			name:          "no existing VLANs",
			existingVLANs: []VLAN{},
			iface:         "lan1",
			expected:      1,
		},
		{
			name: "one existing VLAN",
			existingVLANs: []VLAN{
				{VlanID: 10, Interface: "lan1", VlanInterface: "lan1/1"},
			},
			iface:    "lan1",
			expected: 2,
		},
		{
			name: "gap in slots",
			existingVLANs: []VLAN{
				{VlanID: 10, Interface: "lan1", VlanInterface: "lan1/1"},
				{VlanID: 30, Interface: "lan1", VlanInterface: "lan1/3"},
			},
			iface:    "lan1",
			expected: 2,
		},
		{
			name: "VLANs on different interfaces",
			existingVLANs: []VLAN{
				{VlanID: 10, Interface: "lan1", VlanInterface: "lan1/1"},
				{VlanID: 20, Interface: "lan2", VlanInterface: "lan2/1"},
			},
			iface:    "lan2",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindNextAvailableSlot(tt.existingVLANs, tt.iface)
			if result != tt.expected {
				t.Errorf("FindNextAvailableSlot() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestParseSingleVLAN(t *testing.T) {
	parser := NewVLANParser()

	input := `vlan lan1/1 802.1q vid=10
ip lan1/1 address 192.168.10.1/24
description lan1/1 Management
vlan lan1/2 802.1q vid=20
ip lan1/2 address 192.168.20.1/24`

	// Test finding existing VLAN
	vlan, err := parser.ParseSingleVLAN(input, "lan1", 10)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if vlan == nil {
		t.Fatal("expected VLAN, got nil")
	}
	if vlan.VlanID != 10 {
		t.Errorf("VlanID = %d, want 10", vlan.VlanID)
	}
	if vlan.Name != "Management" {
		t.Errorf("Name = %q, want %q", vlan.Name, "Management")
	}

	// Test not found
	_, err = parser.ParseSingleVLAN(input, "lan1", 99)
	if err == nil {
		t.Error("expected error for non-existent VLAN, got nil")
	}
}
