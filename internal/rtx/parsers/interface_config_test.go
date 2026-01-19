package parsers

import (
	"strings"
	"testing"
)

func TestParseInterfaceConfig(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		interfaceName string
		expected      *InterfaceConfig
		wantErr       bool
	}{
		{
			name: "lan2 with DHCP and security filters",
			input: `description lan2 au
ip lan2 address dhcp
ip lan2 secure filter in 200020 200021 200022 200023 200024 200025 200103 200100 200102 200104 200101 200105 200099
ip lan2 secure filter out 200020 200021 200022 200023 200024 200025 200026 200027 200099 dynamic 200080 200081 200082 200083 200084 200085
ip lan2 nat descriptor 1000`,
			interfaceName: "lan2",
			expected: &InterfaceConfig{
				Name:        "lan2",
				Description: "au",
				IPAddress: &InterfaceIP{
					DHCP: true,
				},
				SecureFilterIn:   []int{200020, 200021, 200022, 200023, 200024, 200025, 200103, 200100, 200102, 200104, 200101, 200105, 200099},
				SecureFilterOut:  []int{200020, 200021, 200022, 200023, 200024, 200025, 200026, 200027, 200099},
				DynamicFilterOut: []int{200080, 200081, 200082, 200083, 200084, 200085},
				NATDescriptor:    1000,
			},
		},
		{
			name: "bridge1 with static IP",
			input: `description bridge1 "Internal bridge"
ip bridge1 address 192.168.1.253/16`,
			interfaceName: "bridge1",
			expected: &InterfaceConfig{
				Name:        "bridge1",
				Description: "Internal bridge",
				IPAddress: &InterfaceIP{
					Address: "192.168.1.253/16",
				},
			},
		},
		{
			name: "lan1 with proxyarp",
			input: `ip lan1 proxyarp on
ip lan1 address 192.168.1.1/24`,
			interfaceName: "lan1",
			expected: &InterfaceConfig{
				Name:     "lan1",
				ProxyARP: true,
				IPAddress: &InterfaceIP{
					Address: "192.168.1.1/24",
				},
			},
		},
		{
			name:          "lan1 with MTU",
			input:         `ip lan1 mtu 1500`,
			interfaceName: "lan1",
			expected: &InterfaceConfig{
				Name: "lan1",
				MTU:  1500,
			},
		},
		{
			name: "pp1 interface",
			input: `description pp1 "PPP connection"
ip pp1 address 10.0.0.1/30
ip pp1 secure filter in 100 101 102`,
			interfaceName: "pp1",
			expected: &InterfaceConfig{
				Name:        "pp1",
				Description: "PPP connection",
				IPAddress: &InterfaceIP{
					Address: "10.0.0.1/30",
				},
				SecureFilterIn: []int{100, 101, 102},
			},
		},
		{
			name:          "empty config",
			input:         "",
			interfaceName: "lan1",
			expected: &InterfaceConfig{
				Name: "lan1",
			},
		},
		{
			name: "tunnel interface",
			input: `ip tunnel1 address 172.16.0.1/30
ip tunnel1 secure filter in 300 301`,
			interfaceName: "tunnel1",
			expected: &InterfaceConfig{
				Name: "tunnel1",
				IPAddress: &InterfaceIP{
					Address: "172.16.0.1/30",
				},
				SecureFilterIn: []int{300, 301},
			},
		},
		{
			name: "secure filter out without dynamic",
			input: `ip lan2 secure filter out 100 101 102`,
			interfaceName: "lan2",
			expected: &InterfaceConfig{
				Name:            "lan2",
				SecureFilterOut: []int{100, 101, 102},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseInterfaceConfig(tt.input, tt.interfaceName)

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

			if result.Name != tt.expected.Name {
				t.Errorf("name = %q, want %q", result.Name, tt.expected.Name)
			}
			if result.Description != tt.expected.Description {
				t.Errorf("description = %q, want %q", result.Description, tt.expected.Description)
			}

			// Check IPAddress
			if tt.expected.IPAddress != nil {
				if result.IPAddress == nil {
					t.Errorf("ip_address is nil, expected %+v", tt.expected.IPAddress)
				} else {
					if result.IPAddress.Address != tt.expected.IPAddress.Address {
						t.Errorf("ip_address.address = %q, want %q", result.IPAddress.Address, tt.expected.IPAddress.Address)
					}
					if result.IPAddress.DHCP != tt.expected.IPAddress.DHCP {
						t.Errorf("ip_address.dhcp = %v, want %v", result.IPAddress.DHCP, tt.expected.IPAddress.DHCP)
					}
				}
			} else if result.IPAddress != nil {
				t.Errorf("ip_address = %+v, want nil", result.IPAddress)
			}

			// Check filters
			if !intSliceEqual(result.SecureFilterIn, tt.expected.SecureFilterIn) {
				t.Errorf("secure_filter_in = %v, want %v", result.SecureFilterIn, tt.expected.SecureFilterIn)
			}
			if !intSliceEqual(result.SecureFilterOut, tt.expected.SecureFilterOut) {
				t.Errorf("secure_filter_out = %v, want %v", result.SecureFilterOut, tt.expected.SecureFilterOut)
			}
			if !intSliceEqual(result.DynamicFilterOut, tt.expected.DynamicFilterOut) {
				t.Errorf("dynamic_filter_out = %v, want %v", result.DynamicFilterOut, tt.expected.DynamicFilterOut)
			}

			if result.NATDescriptor != tt.expected.NATDescriptor {
				t.Errorf("nat_descriptor = %d, want %d", result.NATDescriptor, tt.expected.NATDescriptor)
			}
			if result.ProxyARP != tt.expected.ProxyARP {
				t.Errorf("proxyarp = %v, want %v", result.ProxyARP, tt.expected.ProxyARP)
			}
			if result.MTU != tt.expected.MTU {
				t.Errorf("mtu = %d, want %d", result.MTU, tt.expected.MTU)
			}
		})
	}
}

func TestBuildIPAddressCommand(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		ip       InterfaceIP
		expected string
	}{
		{
			name:  "static IP",
			iface: "lan2",
			ip: InterfaceIP{
				Address: "192.168.1.1/24",
			},
			expected: "ip lan2 address 192.168.1.1/24",
		},
		{
			name:  "DHCP",
			iface: "lan2",
			ip: InterfaceIP{
				DHCP: true,
			},
			expected: "ip lan2 address dhcp",
		},
		{
			name:  "bridge interface static",
			iface: "bridge1",
			ip: InterfaceIP{
				Address: "192.168.1.253/16",
			},
			expected: "ip bridge1 address 192.168.1.253/16",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildIPAddressCommand(tt.iface, tt.ip)
			if result != tt.expected {
				t.Errorf("BuildIPAddressCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteIPAddressCommand(t *testing.T) {
	result := BuildDeleteIPAddressCommand("lan2")
	expected := "no ip lan2 address"
	if result != expected {
		t.Errorf("BuildDeleteIPAddressCommand() = %q, want %q", result, expected)
	}
}

func TestBuildSecureFilterInCommand(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		filters  []int
		expected string
	}{
		{
			name:     "multiple filters",
			iface:    "lan2",
			filters:  []int{200020, 200021, 200022, 200099},
			expected: "ip lan2 secure filter in 200020 200021 200022 200099",
		},
		{
			name:     "single filter",
			iface:    "lan1",
			filters:  []int{100},
			expected: "ip lan1 secure filter in 100",
		},
		{
			name:     "empty filters",
			iface:    "lan1",
			filters:  []int{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildSecureFilterInCommand(tt.iface, tt.filters)
			if result != tt.expected {
				t.Errorf("BuildSecureFilterInCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildSecureFilterOutCommand(t *testing.T) {
	tests := []struct {
		name           string
		iface          string
		filters        []int
		dynamicFilters []int
		expected       string
	}{
		{
			name:           "filters with dynamic",
			iface:          "lan2",
			filters:        []int{200020, 200021, 200099},
			dynamicFilters: []int{200080, 200081},
			expected:       "ip lan2 secure filter out 200020 200021 200099 dynamic 200080 200081",
		},
		{
			name:           "filters without dynamic",
			iface:          "lan2",
			filters:        []int{100, 101, 102},
			dynamicFilters: []int{},
			expected:       "ip lan2 secure filter out 100 101 102",
		},
		{
			name:           "empty filters",
			iface:          "lan1",
			filters:        []int{},
			dynamicFilters: []int{},
			expected:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildSecureFilterOutCommand(tt.iface, tt.filters, tt.dynamicFilters)
			if result != tt.expected {
				t.Errorf("BuildSecureFilterOutCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteSecureFilterCommand(t *testing.T) {
	tests := []struct {
		name      string
		iface     string
		direction string
		expected  string
	}{
		{
			name:      "delete inbound filter",
			iface:     "lan2",
			direction: "in",
			expected:  "no ip lan2 secure filter in",
		},
		{
			name:      "delete outbound filter",
			iface:     "lan2",
			direction: "out",
			expected:  "no ip lan2 secure filter out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteSecureFilterCommand(tt.iface, tt.direction)
			if result != tt.expected {
				t.Errorf("BuildDeleteSecureFilterCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildNATDescriptorCommand(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		natID    int
		expected string
	}{
		{
			name:     "set NAT descriptor",
			iface:    "lan2",
			natID:    1000,
			expected: "ip lan2 nat descriptor 1000",
		},
		{
			name:     "zero NAT descriptor",
			iface:    "lan2",
			natID:    0,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNATDescriptorCommand(tt.iface, tt.natID)
			if result != tt.expected {
				t.Errorf("BuildNATDescriptorCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteNATDescriptorCommand(t *testing.T) {
	result := BuildDeleteNATDescriptorCommand("lan2")
	expected := "no ip lan2 nat descriptor"
	if result != expected {
		t.Errorf("BuildDeleteNATDescriptorCommand() = %q, want %q", result, expected)
	}
}

func TestBuildProxyARPCommand(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		enabled  bool
		expected string
	}{
		{
			name:     "enable proxyarp",
			iface:    "lan1",
			enabled:  true,
			expected: "ip lan1 proxyarp on",
		},
		{
			name:     "disable proxyarp",
			iface:    "lan1",
			enabled:  false,
			expected: "ip lan1 proxyarp off",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildProxyARPCommand(tt.iface, tt.enabled)
			if result != tt.expected {
				t.Errorf("BuildProxyARPCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDescriptionCommand(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		desc     string
		expected string
	}{
		{
			name:     "simple description",
			iface:    "lan2",
			desc:     "au",
			expected: `description lan2 "au"`,
		},
		{
			name:     "description with spaces",
			iface:    "bridge1",
			desc:     "Internal bridge",
			expected: `description bridge1 "Internal bridge"`,
		},
		{
			name:     "empty description",
			iface:    "lan1",
			desc:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDescriptionCommand(tt.iface, tt.desc)
			if result != tt.expected {
				t.Errorf("BuildDescriptionCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteDescriptionCommand(t *testing.T) {
	result := BuildDeleteDescriptionCommand("lan2")
	expected := "no description lan2"
	if result != expected {
		t.Errorf("BuildDeleteDescriptionCommand() = %q, want %q", result, expected)
	}
}

func TestBuildMTUCommand(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		mtu      int
		expected string
	}{
		{
			name:     "set MTU",
			iface:    "lan2",
			mtu:      1500,
			expected: "ip lan2 mtu 1500",
		},
		{
			name:     "zero MTU",
			iface:    "lan1",
			mtu:      0,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildMTUCommand(tt.iface, tt.mtu)
			if result != tt.expected {
				t.Errorf("BuildMTUCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteMTUCommand(t *testing.T) {
	result := BuildDeleteMTUCommand("lan2")
	expected := "no ip lan2 mtu"
	if result != expected {
		t.Errorf("BuildDeleteMTUCommand() = %q, want %q", result, expected)
	}
}

func TestBuildShowInterfaceConfigCommand(t *testing.T) {
	result := BuildShowInterfaceConfigCommand("lan2")
	expected := `show config | grep "lan2"`
	if result != expected {
		t.Errorf("BuildShowInterfaceConfigCommand() = %q, want %q", result, expected)
	}
}

func TestValidateInterfaceConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  InterfaceConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with static IP",
			config: InterfaceConfig{
				Name: "lan2",
				IPAddress: &InterfaceIP{
					Address: "192.168.1.1/24",
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with DHCP",
			config: InterfaceConfig{
				Name: "lan2",
				IPAddress: &InterfaceIP{
					DHCP: true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid config minimal",
			config: InterfaceConfig{
				Name: "lan1",
			},
			wantErr: false,
		},
		{
			name: "empty interface name",
			config: InterfaceConfig{
				Name: "",
			},
			wantErr: true,
			errMsg:  "interface name is required",
		},
		{
			name: "invalid interface name format",
			config: InterfaceConfig{
				Name: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid interface name",
		},
		{
			name: "invalid CIDR",
			config: InterfaceConfig{
				Name: "lan1",
				IPAddress: &InterfaceIP{
					Address: "192.168.1.1",
				},
			},
			wantErr: true,
			errMsg:  "CIDR notation",
		},
		{
			name: "both DHCP and static IP",
			config: InterfaceConfig{
				Name: "lan2",
				IPAddress: &InterfaceIP{
					Address: "192.168.1.1/24",
					DHCP:    true,
				},
			},
			wantErr: true,
			errMsg:  "cannot specify both",
		},
		{
			name: "negative filter number",
			config: InterfaceConfig{
				Name:           "lan1",
				SecureFilterIn: []int{-1},
			},
			wantErr: true,
			errMsg:  "filter numbers must be positive",
		},
		{
			name: "negative NAT descriptor",
			config: InterfaceConfig{
				Name:          "lan1",
				NATDescriptor: -1,
			},
			wantErr: true,
			errMsg:  "NAT descriptor must be non-negative",
		},
		{
			name: "invalid MTU (too small)",
			config: InterfaceConfig{
				Name: "lan1",
				MTU:  60,
			},
			wantErr: true,
			errMsg:  "MTU must be between",
		},
		{
			name: "valid bridge interface",
			config: InterfaceConfig{
				Name: "bridge1",
			},
			wantErr: false,
		},
		{
			name: "valid tunnel interface",
			config: InterfaceConfig{
				Name: "tunnel1",
			},
			wantErr: false,
		},
		{
			name: "valid pp interface",
			config: InterfaceConfig{
				Name: "pp1",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInterfaceConfig(tt.config)

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

func TestValidateInterfaceName(t *testing.T) {
	tests := []struct {
		name    string
		iface   string
		wantErr bool
	}{
		{"valid lan1", "lan1", false},
		{"valid lan2", "lan2", false},
		{"valid lan3", "lan3", false},
		{"valid bridge1", "bridge1", false},
		{"valid bridge10", "bridge10", false},
		{"valid pp1", "pp1", false},
		{"valid pp10", "pp10", false},
		{"valid tunnel1", "tunnel1", false},
		{"valid tunnel100", "tunnel100", false},
		{"invalid empty", "", true},
		{"invalid format", "invalid", true},
		{"invalid eth0", "eth0", true},
		{"invalid lan", "lan", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInterfaceName(tt.iface)
			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Helper functions

func intSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
