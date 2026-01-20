package parsers

import (
	"testing"
)

func TestParseRouteConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []StaticRoute
		wantErr  bool
	}{
		{
			name:  "default route with IP gateway",
			input: `ip route default gateway 192.168.0.1`,
			expected: []StaticRoute{
				{
					Prefix: "0.0.0.0",
					Mask:   "0.0.0.0",
					NextHops: []NextHop{
						{
							NextHop:  "192.168.0.1",
							Distance: 1,
						},
					},
				},
			},
		},
		{
			name:  "default route with weight",
			input: `ip route default gateway 192.168.0.1 weight 10`,
			expected: []StaticRoute{
				{
					Prefix: "0.0.0.0",
					Mask:   "0.0.0.0",
					NextHops: []NextHop{
						{
							NextHop:  "192.168.0.1",
							Distance: 10,
						},
					},
				},
			},
		},
		{
			name:  "network route with CIDR",
			input: `ip route 10.0.0.0/8 gateway 192.168.1.1`,
			expected: []StaticRoute{
				{
					Prefix: "10.0.0.0",
					Mask:   "255.0.0.0",
					NextHops: []NextHop{
						{
							NextHop:  "192.168.1.1",
							Distance: 1,
						},
					},
				},
			},
		},
		{
			name:  "network route with pp interface",
			input: `ip route 172.16.0.0/12 gateway pp 1`,
			expected: []StaticRoute{
				{
					Prefix: "172.16.0.0",
					Mask:   "255.240.0.0",
					NextHops: []NextHop{
						{
							Interface: "pp 1",
							Distance:  1,
						},
					},
				},
			},
		},
		{
			name:  "network route with tunnel interface",
			input: `ip route 192.168.100.0/24 gateway tunnel 1`,
			expected: []StaticRoute{
				{
					Prefix: "192.168.100.0",
					Mask:   "255.255.255.0",
					NextHops: []NextHop{
						{
							Interface: "tunnel 1",
							Distance:  1,
						},
					},
				},
			},
		},
		{
			name:  "route with filter",
			input: `ip route 10.10.0.0/16 gateway 192.168.1.1 filter 100`,
			expected: []StaticRoute{
				{
					Prefix: "10.10.0.0",
					Mask:   "255.255.0.0",
					NextHops: []NextHop{
						{
							NextHop:  "192.168.1.1",
							Distance: 1,
							Filter:   100,
						},
					},
				},
			},
		},
		{
			name: "multi-hop route (load balancing)",
			input: `ip route 10.0.0.0/8 gateway 192.168.1.1 weight 1
ip route 10.0.0.0/8 gateway 192.168.2.1 weight 2`,
			expected: []StaticRoute{
				{
					Prefix: "10.0.0.0",
					Mask:   "255.0.0.0",
					NextHops: []NextHop{
						{
							NextHop:  "192.168.1.1",
							Distance: 1,
						},
						{
							NextHop:  "192.168.2.1",
							Distance: 2,
						},
					},
				},
			},
		},
		{
			name: "multi-hop route with 3 gateways (failover/load balancing)",
			input: `ip route 192.168.100.0/24 gateway 10.0.0.1 weight 1
ip route 192.168.100.0/24 gateway 10.0.0.2 weight 2
ip route 192.168.100.0/24 gateway 10.0.0.3 weight 3`,
			expected: []StaticRoute{
				{
					Prefix: "192.168.100.0",
					Mask:   "255.255.255.0",
					NextHops: []NextHop{
						{
							NextHop:  "10.0.0.1",
							Distance: 1,
						},
						{
							NextHop:  "10.0.0.2",
							Distance: 2,
						},
						{
							NextHop:  "10.0.0.3",
							Distance: 3,
						},
					},
				},
			},
		},
		{
			name: "multiple routes",
			input: `ip route default gateway 192.168.0.1
ip route 10.0.0.0/8 gateway 192.168.1.1
ip route 172.16.0.0/12 gateway tunnel 1`,
			expected: []StaticRoute{
				{
					Prefix: "0.0.0.0",
					Mask:   "0.0.0.0",
					NextHops: []NextHop{
						{
							NextHop:  "192.168.0.1",
							Distance: 1,
						},
					},
				},
				{
					Prefix: "10.0.0.0",
					Mask:   "255.0.0.0",
					NextHops: []NextHop{
						{
							NextHop:  "192.168.1.1",
							Distance: 1,
						},
					},
				},
				{
					Prefix: "172.16.0.0",
					Mask:   "255.240.0.0",
					NextHops: []NextHop{
						{
							Interface: "tunnel 1",
							Distance:  1,
						},
					},
				},
			},
		},
		{
			name:  "route with keepalive (permanent)",
			input: `ip route 10.0.0.0/8 gateway 192.168.1.1 keepalive`,
			expected: []StaticRoute{
				{
					Prefix: "10.0.0.0",
					Mask:   "255.0.0.0",
					NextHops: []NextHop{
						{
							NextHop:   "192.168.1.1",
							Distance:  1,
							Permanent: true,
						},
					},
				},
			},
		},
		{
			name:  "route with dhcp interface",
			input: `ip route default gateway dhcp lan1`,
			expected: []StaticRoute{
				{
					Prefix: "0.0.0.0",
					Mask:   "0.0.0.0",
					NextHops: []NextHop{
						{
							Interface: "dhcp lan1",
							Distance:  1,
						},
					},
				},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []StaticRoute{},
		},
		{
			name:     "comment lines ignored",
			input:    "# ip route default gateway 192.168.0.1",
			expected: []StaticRoute{},
		},
		{
			name:     "no ip route lines ignored",
			input:    "no ip route default",
			expected: []StaticRoute{},
		},
	}

	parser := NewStaticRouteParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseRouteConfig(tt.input)

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
				t.Errorf("expected %d routes, got %d", len(tt.expected), len(result))
				return
			}

			// Create a map for easier comparison
			resultMap := make(map[string]StaticRoute)
			for _, r := range result {
				key := r.Prefix + "/" + r.Mask
				resultMap[key] = r
			}

			for _, expected := range tt.expected {
				key := expected.Prefix + "/" + expected.Mask
				got, ok := resultMap[key]
				if !ok {
					t.Errorf("route %s not found in result", key)
					continue
				}

				if got.Prefix != expected.Prefix {
					t.Errorf("prefix = %q, want %q", got.Prefix, expected.Prefix)
				}
				if got.Mask != expected.Mask {
					t.Errorf("mask = %q, want %q", got.Mask, expected.Mask)
				}
				if len(got.NextHops) != len(expected.NextHops) {
					t.Errorf("next_hops count = %d, want %d", len(got.NextHops), len(expected.NextHops))
					continue
				}

				for i, expectedHop := range expected.NextHops {
					gotHop := got.NextHops[i]
					if gotHop.NextHop != expectedHop.NextHop {
						t.Errorf("next_hops[%d].next_hop = %q, want %q", i, gotHop.NextHop, expectedHop.NextHop)
					}
					if gotHop.Interface != expectedHop.Interface {
						t.Errorf("next_hops[%d].interface = %q, want %q", i, gotHop.Interface, expectedHop.Interface)
					}
					if gotHop.Distance != expectedHop.Distance {
						t.Errorf("next_hops[%d].distance = %d, want %d", i, gotHop.Distance, expectedHop.Distance)
					}
					if gotHop.Filter != expectedHop.Filter {
						t.Errorf("next_hops[%d].filter = %d, want %d", i, gotHop.Filter, expectedHop.Filter)
					}
					if gotHop.Permanent != expectedHop.Permanent {
						t.Errorf("next_hops[%d].permanent = %v, want %v", i, gotHop.Permanent, expectedHop.Permanent)
					}
				}
			}
		})
	}
}

func TestParseSingleRoute(t *testing.T) {
	parser := NewStaticRouteParser()

	input := `ip route default gateway 192.168.0.1
ip route 10.0.0.0/8 gateway 192.168.1.1`

	route, err := parser.ParseSingleRoute(input, "10.0.0.0", "255.0.0.0")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if route.Prefix != "10.0.0.0" {
		t.Errorf("prefix = %q, want %q", route.Prefix, "10.0.0.0")
	}
	if route.Mask != "255.0.0.0" {
		t.Errorf("mask = %q, want %q", route.Mask, "255.0.0.0")
	}

	// Test not found
	_, err = parser.ParseSingleRoute(input, "172.16.0.0", "255.240.0.0")
	if err == nil {
		t.Errorf("expected error for non-existent route, got nil")
	}
}

func TestParseSingleRouteMultiGateway(t *testing.T) {
	parser := NewStaticRouteParser()

	// Simulate output from: show config | grep "ip route 192.168.100.0/24"
	// This should return ALL lines matching the prefix, including multiple gateways
	input := `ip route 192.168.100.0/24 gateway 10.0.0.1 weight 1
ip route 192.168.100.0/24 gateway 10.0.0.2 weight 2
ip route 192.168.100.0/24 gateway 10.0.0.3 weight 3`

	route, err := parser.ParseSingleRoute(input, "192.168.100.0", "255.255.255.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if route.Prefix != "192.168.100.0" {
		t.Errorf("prefix = %q, want %q", route.Prefix, "192.168.100.0")
	}
	if route.Mask != "255.255.255.0" {
		t.Errorf("mask = %q, want %q", route.Mask, "255.255.255.0")
	}

	// Verify all 3 gateways are captured
	if len(route.NextHops) != 3 {
		t.Fatalf("expected 3 next_hops, got %d", len(route.NextHops))
	}

	expectedHops := []struct {
		nextHop  string
		distance int
	}{
		{"10.0.0.1", 1},
		{"10.0.0.2", 2},
		{"10.0.0.3", 3},
	}

	for i, expected := range expectedHops {
		got := route.NextHops[i]
		if got.NextHop != expected.nextHop {
			t.Errorf("next_hops[%d].next_hop = %q, want %q", i, got.NextHop, expected.nextHop)
		}
		if got.Distance != expected.distance {
			t.Errorf("next_hops[%d].distance = %d, want %d", i, got.Distance, expected.distance)
		}
	}
}

func TestBuildIPRouteCommand(t *testing.T) {
	tests := []struct {
		name     string
		route    StaticRoute
		hop      NextHop
		expected string
	}{
		{
			name: "default route with IP gateway",
			route: StaticRoute{
				Prefix: "0.0.0.0",
				Mask:   "0.0.0.0",
			},
			hop: NextHop{
				NextHop:  "192.168.0.1",
				Distance: 1,
			},
			expected: "ip route default gateway 192.168.0.1",
		},
		{
			name: "default route with weight",
			route: StaticRoute{
				Prefix: "0.0.0.0",
				Mask:   "0.0.0.0",
			},
			hop: NextHop{
				NextHop:  "192.168.0.1",
				Distance: 10,
			},
			expected: "ip route default gateway 192.168.0.1 weight 10",
		},
		{
			name: "network route with CIDR",
			route: StaticRoute{
				Prefix: "10.0.0.0",
				Mask:   "255.0.0.0",
			},
			hop: NextHop{
				NextHop:  "192.168.1.1",
				Distance: 1,
			},
			expected: "ip route 10.0.0.0/8 gateway 192.168.1.1",
		},
		{
			name: "route with pp interface",
			route: StaticRoute{
				Prefix: "172.16.0.0",
				Mask:   "255.240.0.0",
			},
			hop: NextHop{
				Interface: "pp 1",
				Distance:  1,
			},
			expected: "ip route 172.16.0.0/12 gateway pp 1",
		},
		{
			name: "route with tunnel interface",
			route: StaticRoute{
				Prefix: "192.168.100.0",
				Mask:   "255.255.255.0",
			},
			hop: NextHop{
				Interface: "tunnel 1",
				Distance:  1,
			},
			expected: "ip route 192.168.100.0/24 gateway tunnel 1",
		},
		{
			name: "route with filter",
			route: StaticRoute{
				Prefix: "10.10.0.0",
				Mask:   "255.255.0.0",
			},
			hop: NextHop{
				NextHop:  "192.168.1.1",
				Distance: 1,
				Filter:   100,
			},
			expected: "ip route 10.10.0.0/16 gateway 192.168.1.1 filter 100",
		},
		{
			name: "route with keepalive",
			route: StaticRoute{
				Prefix: "10.0.0.0",
				Mask:   "255.0.0.0",
			},
			hop: NextHop{
				NextHop:   "192.168.1.1",
				Distance:  1,
				Permanent: true,
			},
			expected: "ip route 10.0.0.0/8 gateway 192.168.1.1 keepalive",
		},
		{
			name: "route with all options",
			route: StaticRoute{
				Prefix: "10.0.0.0",
				Mask:   "255.0.0.0",
			},
			hop: NextHop{
				NextHop:   "192.168.1.1",
				Distance:  20,
				Filter:    50,
				Permanent: true,
			},
			expected: "ip route 10.0.0.0/8 gateway 192.168.1.1 weight 20 filter 50 keepalive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildIPRouteCommand(tt.route, tt.hop)
			if result != tt.expected {
				t.Errorf("BuildIPRouteCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteIPRouteCommand(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		mask     string
		hop      *NextHop
		expected string
	}{
		{
			name:     "delete all routes for network",
			prefix:   "10.0.0.0",
			mask:     "255.0.0.0",
			hop:      nil,
			expected: "no ip route 10.0.0.0/8",
		},
		{
			name:     "delete default route",
			prefix:   "0.0.0.0",
			mask:     "0.0.0.0",
			hop:      nil,
			expected: "no ip route default",
		},
		{
			name:   "delete specific next hop",
			prefix: "10.0.0.0",
			mask:   "255.0.0.0",
			hop: &NextHop{
				NextHop: "192.168.1.1",
			},
			expected: "no ip route 10.0.0.0/8 gateway 192.168.1.1",
		},
		{
			name:   "delete pp interface route",
			prefix: "172.16.0.0",
			mask:   "255.240.0.0",
			hop: &NextHop{
				Interface: "pp 1",
			},
			expected: "no ip route 172.16.0.0/12 gateway pp 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteIPRouteCommand(tt.prefix, tt.mask, tt.hop)
			if result != tt.expected {
				t.Errorf("BuildDeleteIPRouteCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildShowIPRouteConfigCommand(t *testing.T) {
	result := BuildShowIPRouteConfigCommand()
	expected := `show config | grep "ip route"`
	if result != expected {
		t.Errorf("BuildShowIPRouteConfigCommand() = %q, want %q", result, expected)
	}
}

func TestBuildShowSingleRouteConfigCommand(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		mask     string
		expected string
	}{
		{
			name:     "default route",
			prefix:   "0.0.0.0",
			mask:     "0.0.0.0",
			expected: `show config | grep "ip route default"`,
		},
		{
			name:     "network route",
			prefix:   "10.0.0.0",
			mask:     "255.0.0.0",
			expected: `show config | grep "ip route 10.0.0.0/8"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildShowSingleRouteConfigCommand(tt.prefix, tt.mask)
			if result != tt.expected {
				t.Errorf("BuildShowSingleRouteConfigCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValidateStaticRoute(t *testing.T) {
	tests := []struct {
		name    string
		route   StaticRoute
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid route with IP gateway",
			route: StaticRoute{
				Prefix: "10.0.0.0",
				Mask:   "255.0.0.0",
				NextHops: []NextHop{
					{NextHop: "192.168.1.1", Distance: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "valid default route",
			route: StaticRoute{
				Prefix: "0.0.0.0",
				Mask:   "0.0.0.0",
				NextHops: []NextHop{
					{NextHop: "192.168.0.1", Distance: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "valid route with interface",
			route: StaticRoute{
				Prefix: "172.16.0.0",
				Mask:   "255.240.0.0",
				NextHops: []NextHop{
					{Interface: "tunnel 1", Distance: 1},
				},
			},
			wantErr: false,
		},
		{
			name: "valid multi-hop route",
			route: StaticRoute{
				Prefix: "10.0.0.0",
				Mask:   "255.0.0.0",
				NextHops: []NextHop{
					{NextHop: "192.168.1.1", Distance: 1},
					{NextHop: "192.168.2.1", Distance: 10},
				},
			},
			wantErr: false,
		},
		{
			name: "empty prefix",
			route: StaticRoute{
				Prefix: "",
				Mask:   "255.0.0.0",
				NextHops: []NextHop{
					{NextHop: "192.168.1.1", Distance: 1},
				},
			},
			wantErr: true,
			errMsg:  "prefix is required",
		},
		{
			name: "invalid prefix",
			route: StaticRoute{
				Prefix: "invalid",
				Mask:   "255.0.0.0",
				NextHops: []NextHop{
					{NextHop: "192.168.1.1", Distance: 1},
				},
			},
			wantErr: true,
			errMsg:  "invalid prefix",
		},
		{
			name: "empty mask",
			route: StaticRoute{
				Prefix: "10.0.0.0",
				Mask:   "",
				NextHops: []NextHop{
					{NextHop: "192.168.1.1", Distance: 1},
				},
			},
			wantErr: true,
			errMsg:  "mask is required",
		},
		{
			name: "invalid mask",
			route: StaticRoute{
				Prefix: "10.0.0.0",
				Mask:   "invalid",
				NextHops: []NextHop{
					{NextHop: "192.168.1.1", Distance: 1},
				},
			},
			wantErr: true,
			errMsg:  "invalid mask",
		},
		{
			name: "no next hops",
			route: StaticRoute{
				Prefix:   "10.0.0.0",
				Mask:     "255.0.0.0",
				NextHops: []NextHop{},
			},
			wantErr: true,
			errMsg:  "at least one next_hop is required",
		},
		{
			name: "next hop missing both gateway and interface",
			route: StaticRoute{
				Prefix: "10.0.0.0",
				Mask:   "255.0.0.0",
				NextHops: []NextHop{
					{Distance: 1},
				},
			},
			wantErr: true,
			errMsg:  "either next_hop or interface must be specified",
		},
		{
			name: "invalid next hop IP",
			route: StaticRoute{
				Prefix: "10.0.0.0",
				Mask:   "255.0.0.0",
				NextHops: []NextHop{
					{NextHop: "invalid", Distance: 1},
				},
			},
			wantErr: true,
			errMsg:  "invalid next_hop IP address",
		},
		{
			name: "invalid interface",
			route: StaticRoute{
				Prefix: "10.0.0.0",
				Mask:   "255.0.0.0",
				NextHops: []NextHop{
					{Interface: "invalid 1", Distance: 1},
				},
			},
			wantErr: true,
			errMsg:  "invalid interface format",
		},
		{
			name: "distance out of range (negative)",
			route: StaticRoute{
				Prefix: "10.0.0.0",
				Mask:   "255.0.0.0",
				NextHops: []NextHop{
					{NextHop: "192.168.1.1", Distance: -1},
				},
			},
			wantErr: true,
			errMsg:  "distance must be between 0 and 100",
		},
		{
			name: "distance out of range (too high)",
			route: StaticRoute{
				Prefix: "10.0.0.0",
				Mask:   "255.0.0.0",
				NextHops: []NextHop{
					{NextHop: "192.168.1.1", Distance: 101},
				},
			},
			wantErr: true,
			errMsg:  "distance must be between 0 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStaticRoute(tt.route)

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

func TestCIDRToMask(t *testing.T) {
	tests := []struct {
		cidr     int
		expected string
	}{
		{0, "0.0.0.0"},
		{8, "255.0.0.0"},
		{12, "255.240.0.0"},
		{16, "255.255.0.0"},
		{24, "255.255.255.0"},
		{32, "255.255.255.255"},
	}

	for _, tt := range tests {
		result := cidrToMask(tt.cidr)
		if result != tt.expected {
			t.Errorf("cidrToMask(%d) = %q, want %q", tt.cidr, result, tt.expected)
		}
	}
}

func TestMaskToCIDR(t *testing.T) {
	tests := []struct {
		mask     string
		expected int
	}{
		{"0.0.0.0", 0},
		{"255.0.0.0", 8},
		{"255.240.0.0", 12},
		{"255.255.0.0", 16},
		{"255.255.255.0", 24},
		{"255.255.255.255", 32},
	}

	for _, tt := range tests {
		result := maskToCIDR(tt.mask)
		if result != tt.expected {
			t.Errorf("maskToCIDR(%q) = %d, want %d", tt.mask, result, tt.expected)
		}
	}
}

func TestIsValidInterface(t *testing.T) {
	tests := []struct {
		iface    string
		expected bool
	}{
		{"pp 1", true},
		{"pp 10", true},
		{"tunnel 1", true},
		{"tunnel 100", true},
		{"dhcp lan1", true},
		{"null", true},
		{"loopback", true},
		{"lan1", true},
		{"invalid", false},
		{"eth0", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isValidInterface(tt.iface)
		if result != tt.expected {
			t.Errorf("isValidInterface(%q) = %v, want %v", tt.iface, result, tt.expected)
		}
	}
}
