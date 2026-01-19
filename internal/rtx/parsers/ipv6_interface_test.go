package parsers

import (
	"testing"
)

func TestParseIPv6InterfaceConfig(t *testing.T) {
	tests := []struct {
		name          string
		raw           string
		interfaceName string
		want          *IPv6InterfaceConfig
		wantErr       bool
	}{
		{
			name: "static IPv6 address",
			raw: `ipv6 lan1 address 2001:db8::1/64
`,
			interfaceName: "lan1",
			want: &IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{
					{Address: "2001:db8::1/64"},
				},
			},
		},
		{
			name: "prefix-based address",
			raw: `ipv6 lan1 address ra-prefix@lan2::1/64
`,
			interfaceName: "lan1",
			want: &IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{
					{PrefixRef: "ra-prefix@lan2", InterfaceID: "::1/64"},
				},
			},
		},
		{
			name: "multiple addresses",
			raw: `ipv6 lan1 address 2001:db8::1/64
ipv6 lan1 address fe80::1/10
`,
			interfaceName: "lan1",
			want: &IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{
					{Address: "2001:db8::1/64"},
					{Address: "fe80::1/10"},
				},
			},
		},
		{
			name: "RTADV configuration",
			raw: `ipv6 lan1 rtadv send 1 o_flag=on m_flag=off
`,
			interfaceName: "lan1",
			want: &IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{},
				RTADV: &RTADVConfig{
					Enabled:  true,
					PrefixID: 1,
					OFlag:    true,
					MFlag:    false,
				},
			},
		},
		{
			name: "RTADV with lifetime",
			raw: `ipv6 lan1 rtadv send 1 o_flag=on m_flag=on lifetime=1800
`,
			interfaceName: "lan1",
			want: &IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{},
				RTADV: &RTADVConfig{
					Enabled:  true,
					PrefixID: 1,
					OFlag:    true,
					MFlag:    true,
					Lifetime: 1800,
				},
			},
		},
		{
			name: "DHCPv6 server",
			raw: `ipv6 lan1 dhcp service server
`,
			interfaceName: "lan1",
			want: &IPv6InterfaceConfig{
				Interface:     "lan1",
				Addresses:     []IPv6Address{},
				DHCPv6Service: "server",
			},
		},
		{
			name: "DHCPv6 client",
			raw: `ipv6 lan2 dhcp service client
`,
			interfaceName: "lan2",
			want: &IPv6InterfaceConfig{
				Interface:     "lan2",
				Addresses:     []IPv6Address{},
				DHCPv6Service: "client",
			},
		},
		{
			name: "MTU setting",
			raw: `ipv6 lan1 mtu 1500
`,
			interfaceName: "lan1",
			want: &IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{},
				MTU:       1500,
			},
		},
		{
			name: "security filters",
			raw: `ipv6 lan1 secure filter in 1 2 3
ipv6 lan1 secure filter out 10 20 30
`,
			interfaceName: "lan1",
			want: &IPv6InterfaceConfig{
				Interface:       "lan1",
				Addresses:       []IPv6Address{},
				SecureFilterIn:  []int{1, 2, 3},
				SecureFilterOut: []int{10, 20, 30},
			},
		},
		{
			name: "security filters with dynamic",
			raw: `ipv6 lan1 secure filter in 1 2
ipv6 lan1 secure filter out 10 20 dynamic 100 101
`,
			interfaceName: "lan1",
			want: &IPv6InterfaceConfig{
				Interface:        "lan1",
				Addresses:        []IPv6Address{},
				SecureFilterIn:   []int{1, 2},
				SecureFilterOut:  []int{10, 20},
				DynamicFilterOut: []int{100, 101},
			},
		},
		{
			name: "full configuration",
			raw: `ipv6 lan1 address 2001:db8::1/64
ipv6 lan1 rtadv send 1 o_flag=on m_flag=off
ipv6 lan1 dhcp service server
ipv6 lan1 mtu 1500
ipv6 lan1 secure filter in 1 2 3
ipv6 lan1 secure filter out 10 20 dynamic 100
`,
			interfaceName: "lan1",
			want: &IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{
					{Address: "2001:db8::1/64"},
				},
				RTADV: &RTADVConfig{
					Enabled:  true,
					PrefixID: 1,
					OFlag:    true,
					MFlag:    false,
				},
				DHCPv6Service:    "server",
				MTU:              1500,
				SecureFilterIn:   []int{1, 2, 3},
				SecureFilterOut:  []int{10, 20},
				DynamicFilterOut: []int{100},
			},
		},
		{
			name:          "empty configuration",
			raw:           ``,
			interfaceName: "lan1",
			want: &IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIPv6InterfaceConfig(tt.raw, tt.interfaceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseIPv6InterfaceConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Compare interface name
			if got.Interface != tt.want.Interface {
				t.Errorf("Interface = %v, want %v", got.Interface, tt.want.Interface)
			}

			// Compare addresses
			if len(got.Addresses) != len(tt.want.Addresses) {
				t.Errorf("Addresses length = %v, want %v", len(got.Addresses), len(tt.want.Addresses))
			} else {
				for i, addr := range got.Addresses {
					if addr != tt.want.Addresses[i] {
						t.Errorf("Addresses[%d] = %v, want %v", i, addr, tt.want.Addresses[i])
					}
				}
			}

			// Compare RTADV
			if (got.RTADV == nil) != (tt.want.RTADV == nil) {
				t.Errorf("RTADV = %v, want %v", got.RTADV, tt.want.RTADV)
			} else if got.RTADV != nil {
				if *got.RTADV != *tt.want.RTADV {
					t.Errorf("RTADV = %+v, want %+v", *got.RTADV, *tt.want.RTADV)
				}
			}

			// Compare other fields
			if got.DHCPv6Service != tt.want.DHCPv6Service {
				t.Errorf("DHCPv6Service = %v, want %v", got.DHCPv6Service, tt.want.DHCPv6Service)
			}
			if got.MTU != tt.want.MTU {
				t.Errorf("MTU = %v, want %v", got.MTU, tt.want.MTU)
			}

			// Compare filters
			if !intSlicesEqual(got.SecureFilterIn, tt.want.SecureFilterIn) {
				t.Errorf("SecureFilterIn = %v, want %v", got.SecureFilterIn, tt.want.SecureFilterIn)
			}
			if !intSlicesEqual(got.SecureFilterOut, tt.want.SecureFilterOut) {
				t.Errorf("SecureFilterOut = %v, want %v", got.SecureFilterOut, tt.want.SecureFilterOut)
			}
			if !intSlicesEqual(got.DynamicFilterOut, tt.want.DynamicFilterOut) {
				t.Errorf("DynamicFilterOut = %v, want %v", got.DynamicFilterOut, tt.want.DynamicFilterOut)
			}
		})
	}
}

func TestBuildIPv6AddressCommand(t *testing.T) {
	tests := []struct {
		name  string
		iface string
		addr  IPv6Address
		want  string
	}{
		{
			name:  "static address",
			iface: "lan1",
			addr:  IPv6Address{Address: "2001:db8::1/64"},
			want:  "ipv6 lan1 address 2001:db8::1/64",
		},
		{
			name:  "prefix-based address",
			iface: "lan1",
			addr:  IPv6Address{PrefixRef: "ra-prefix@lan2", InterfaceID: "::1/64"},
			want:  "ipv6 lan1 address ra-prefix@lan2::1/64",
		},
		{
			name:  "empty address",
			iface: "lan1",
			addr:  IPv6Address{},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildIPv6AddressCommand(tt.iface, tt.addr)
			if got != tt.want {
				t.Errorf("BuildIPv6AddressCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildDeleteIPv6AddressCommand(t *testing.T) {
	tests := []struct {
		name  string
		iface string
		addr  *IPv6Address
		want  string
	}{
		{
			name:  "delete all addresses",
			iface: "lan1",
			addr:  nil,
			want:  "no ipv6 lan1 address",
		},
		{
			name:  "delete specific static address",
			iface: "lan1",
			addr:  &IPv6Address{Address: "2001:db8::1/64"},
			want:  "no ipv6 lan1 address 2001:db8::1/64",
		},
		{
			name:  "delete prefix-based address",
			iface: "lan1",
			addr:  &IPv6Address{PrefixRef: "ra-prefix@lan2", InterfaceID: "::1/64"},
			want:  "no ipv6 lan1 address ra-prefix@lan2::1/64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildDeleteIPv6AddressCommand(tt.iface, tt.addr)
			if got != tt.want {
				t.Errorf("BuildDeleteIPv6AddressCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildIPv6RTADVCommand(t *testing.T) {
	tests := []struct {
		name  string
		iface string
		rtadv RTADVConfig
		want  string
	}{
		{
			name:  "basic RTADV",
			iface: "lan1",
			rtadv: RTADVConfig{Enabled: true, PrefixID: 1, OFlag: true, MFlag: false},
			want:  "ipv6 lan1 rtadv send 1 o_flag=on m_flag=off",
		},
		{
			name:  "RTADV with m_flag",
			iface: "lan1",
			rtadv: RTADVConfig{Enabled: true, PrefixID: 2, OFlag: false, MFlag: true},
			want:  "ipv6 lan1 rtadv send 2 o_flag=off m_flag=on",
		},
		{
			name:  "RTADV with lifetime",
			iface: "lan1",
			rtadv: RTADVConfig{Enabled: true, PrefixID: 1, OFlag: true, MFlag: true, Lifetime: 1800},
			want:  "ipv6 lan1 rtadv send 1 o_flag=on m_flag=on lifetime=1800",
		},
		{
			name:  "disabled RTADV",
			iface: "lan1",
			rtadv: RTADVConfig{Enabled: false, PrefixID: 1},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildIPv6RTADVCommand(tt.iface, tt.rtadv)
			if got != tt.want {
				t.Errorf("BuildIPv6RTADVCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildIPv6DHCPv6Command(t *testing.T) {
	tests := []struct {
		name    string
		iface   string
		service string
		want    string
	}{
		{
			name:    "DHCPv6 server",
			iface:   "lan1",
			service: "server",
			want:    "ipv6 lan1 dhcp service server",
		},
		{
			name:    "DHCPv6 client",
			iface:   "lan2",
			service: "client",
			want:    "ipv6 lan2 dhcp service client",
		},
		{
			name:    "DHCPv6 off",
			iface:   "lan1",
			service: "off",
			want:    "",
		},
		{
			name:    "empty service",
			iface:   "lan1",
			service: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildIPv6DHCPv6Command(tt.iface, tt.service)
			if got != tt.want {
				t.Errorf("BuildIPv6DHCPv6Command() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildIPv6MTUCommand(t *testing.T) {
	tests := []struct {
		name  string
		iface string
		mtu   int
		want  string
	}{
		{
			name:  "set MTU",
			iface: "lan1",
			mtu:   1500,
			want:  "ipv6 lan1 mtu 1500",
		},
		{
			name:  "minimum IPv6 MTU",
			iface: "lan1",
			mtu:   1280,
			want:  "ipv6 lan1 mtu 1280",
		},
		{
			name:  "zero MTU (no command)",
			iface: "lan1",
			mtu:   0,
			want:  "",
		},
		{
			name:  "negative MTU (no command)",
			iface: "lan1",
			mtu:   -1,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildIPv6MTUCommand(tt.iface, tt.mtu)
			if got != tt.want {
				t.Errorf("BuildIPv6MTUCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildIPv6SecureFilterCommands(t *testing.T) {
	t.Run("inbound filter", func(t *testing.T) {
		got := BuildIPv6SecureFilterInCommand("lan1", []int{1, 2, 3})
		want := "ipv6 lan1 secure filter in 1 2 3"
		if got != want {
			t.Errorf("BuildIPv6SecureFilterInCommand() = %q, want %q", got, want)
		}
	})

	t.Run("empty inbound filter", func(t *testing.T) {
		got := BuildIPv6SecureFilterInCommand("lan1", []int{})
		if got != "" {
			t.Errorf("BuildIPv6SecureFilterInCommand() = %q, want empty", got)
		}
	})

	t.Run("outbound filter", func(t *testing.T) {
		got := BuildIPv6SecureFilterOutCommand("lan1", []int{10, 20}, nil)
		want := "ipv6 lan1 secure filter out 10 20"
		if got != want {
			t.Errorf("BuildIPv6SecureFilterOutCommand() = %q, want %q", got, want)
		}
	})

	t.Run("outbound filter with dynamic", func(t *testing.T) {
		got := BuildIPv6SecureFilterOutCommand("lan1", []int{10, 20}, []int{100, 101})
		want := "ipv6 lan1 secure filter out 10 20 dynamic 100 101"
		if got != want {
			t.Errorf("BuildIPv6SecureFilterOutCommand() = %q, want %q", got, want)
		}
	})

	t.Run("delete filter", func(t *testing.T) {
		got := BuildDeleteIPv6SecureFilterCommand("lan1", "in")
		want := "no ipv6 lan1 secure filter in"
		if got != want {
			t.Errorf("BuildDeleteIPv6SecureFilterCommand() = %q, want %q", got, want)
		}
	})
}

func TestBuildShowIPv6InterfaceConfigCommand(t *testing.T) {
	got := BuildShowIPv6InterfaceConfigCommand("lan1")
	want := `show config | grep "ipv6 lan1"`
	if got != want {
		t.Errorf("BuildShowIPv6InterfaceConfigCommand() = %q, want %q", got, want)
	}
}

func TestBuildDeleteIPv6InterfaceCommands(t *testing.T) {
	got := BuildDeleteIPv6InterfaceCommands("lan1")
	want := []string{
		"no ipv6 lan1 address",
		"no ipv6 lan1 rtadv send",
		"no ipv6 lan1 dhcp service",
		"no ipv6 lan1 mtu",
		"no ipv6 lan1 secure filter in",
		"no ipv6 lan1 secure filter out",
	}

	if len(got) != len(want) {
		t.Errorf("BuildDeleteIPv6InterfaceCommands() returned %d commands, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("BuildDeleteIPv6InterfaceCommands()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestValidateIPv6InterfaceConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  IPv6InterfaceConfig
		wantErr bool
	}{
		{
			name: "valid config with static address",
			config: IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{{Address: "2001:db8::1/64"}},
			},
			wantErr: false,
		},
		{
			name: "valid config with prefix-based address",
			config: IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{{PrefixRef: "ra-prefix@lan2", InterfaceID: "::1/64"}},
			},
			wantErr: false,
		},
		{
			name: "valid full config",
			config: IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{{Address: "2001:db8::1/64"}},
				RTADV:     &RTADVConfig{Enabled: true, PrefixID: 1},
				MTU:       1500,
			},
			wantErr: false,
		},
		{
			name: "invalid interface name",
			config: IPv6InterfaceConfig{
				Interface: "invalid",
			},
			wantErr: true,
		},
		{
			name: "empty interface name",
			config: IPv6InterfaceConfig{
				Interface: "",
			},
			wantErr: true,
		},
		{
			name: "invalid MTU (too small)",
			config: IPv6InterfaceConfig{
				Interface: "lan1",
				MTU:       1000,
			},
			wantErr: true,
		},
		{
			name: "invalid MTU (too large)",
			config: IPv6InterfaceConfig{
				Interface: "lan1",
				MTU:       100000,
			},
			wantErr: true,
		},
		{
			name: "invalid RTADV prefix_id",
			config: IPv6InterfaceConfig{
				Interface: "lan1",
				RTADV:     &RTADVConfig{Enabled: true, PrefixID: 0},
			},
			wantErr: true,
		},
		{
			name: "invalid DHCPv6 service",
			config: IPv6InterfaceConfig{
				Interface:     "lan1",
				DHCPv6Service: "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid filter number",
			config: IPv6InterfaceConfig{
				Interface:      "lan1",
				SecureFilterIn: []int{0},
			},
			wantErr: true,
		},
		{
			name: "address without prefix length",
			config: IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{{Address: "2001:db8::1"}},
			},
			wantErr: true,
		},
		{
			name: "prefix ref without @",
			config: IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{{PrefixRef: "ra-prefix", InterfaceID: "::1/64"}},
			},
			wantErr: true,
		},
		{
			name: "interface_id without ::",
			config: IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{{PrefixRef: "ra-prefix@lan2", InterfaceID: "1/64"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPv6InterfaceConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIPv6InterfaceConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateIPv6InterfaceName(t *testing.T) {
	tests := []struct {
		name    string
		ifName  string
		wantErr bool
	}{
		{"lan1", "lan1", false},
		{"lan2", "lan2", false},
		{"lan10", "lan10", false},
		{"bridge1", "bridge1", false},
		{"pp1", "pp1", false},
		{"tunnel1", "tunnel1", false},
		{"empty", "", true},
		{"invalid", "invalid", true},
		{"eth0", "eth0", true},
		{"lan", "lan", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPv6InterfaceName(tt.ifName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIPv6InterfaceName(%q) error = %v, wantErr %v", tt.ifName, err, tt.wantErr)
			}
		})
	}
}

func TestIsValidIPv6CIDR(t *testing.T) {
	tests := []struct {
		name string
		cidr string
		want bool
	}{
		{"valid /64", "2001:db8::1/64", true},
		{"valid /128", "2001:db8::1/128", true},
		{"valid /0", "::/0", true},
		{"no prefix", "2001:db8::1", false},
		{"invalid prefix", "2001:db8::1/129", false},
		{"negative prefix", "2001:db8::1/-1", false},
		{"no colon", "invalid/64", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidIPv6CIDR(tt.cidr)
			if got != tt.want {
				t.Errorf("IsValidIPv6CIDR(%q) = %v, want %v", tt.cidr, got, tt.want)
			}
		})
	}
}

// intSlicesEqual compares two int slices for equality
func intSlicesEqual(a, b []int) bool {
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
