package parsers

import (
	"reflect"
	"testing"
)

func TestParseStaticRoutes(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []StaticRoute
		wantErr  bool
	}{
		{
			name: "single default route with IP gateway",
			input: []byte(`ip route default gateway 192.168.1.1
exit`),
			expected: []StaticRoute{
				{
					Destination: "0.0.0.0/0",
					GatewayIP:   "192.168.1.1",
					Interface:   "",
					Metric:      1,
					Weight:      0,
					Description: "",
					Hide:        false,
				},
			},
			wantErr: false,
		},
		{
			name: "single route with CIDR and IP gateway",
			input: []byte(`ip route 192.168.100.0/24 gateway 192.168.1.254
exit`),
			expected: []StaticRoute{
				{
					Destination: "192.168.100.0/24",
					GatewayIP:   "192.168.1.254",
					Interface:   "",
					Metric:      1,
					Weight:      0,
					Description: "",
					Hide:        false,
				},
			},
			wantErr: false,
		},
		{
			name: "single route with interface gateway",
			input: []byte(`ip route 10.0.0.0/8 gateway pp1
exit`),
			expected: []StaticRoute{
				{
					Destination:      "10.0.0.0/8",
					GatewayIP:        "",
					GatewayInterface: "pp1",
					Interface:        "",
					Metric:           1,
					Weight:           0,
					Description:      "",
					Hide:             false,
				},
			},
			wantErr: false,
		},
		{
			name: "route with IP gateway and interface",
			input: []byte(`ip route 172.16.0.0/16 gateway 192.168.1.1 interface lan2
exit`),
			expected: []StaticRoute{
				{
					Destination: "172.16.0.0/16",
					GatewayIP:   "192.168.1.1",
					Interface:   "lan2",
					Metric:      1,
					Weight:      0,
					Description: "",
					Hide:        false,
				},
			},
			wantErr: false,
		},
		{
			name: "route with metric",
			input: []byte(`ip route 10.10.0.0/16 gateway 192.168.1.1 metric 50
exit`),
			expected: []StaticRoute{
				{
					Destination: "10.10.0.0/16",
					GatewayIP:   "192.168.1.1",
					Interface:   "",
					Metric:      50,
					Weight:      0,
					Description: "",
					Hide:        false,
				},
			},
			wantErr: false,
		},
		{
			name: "route with weight (ECMP)",
			input: []byte(`ip route 10.20.0.0/16 gateway 192.168.1.1 weight 100
exit`),
			expected: []StaticRoute{
				{
					Destination: "10.20.0.0/16",
					GatewayIP:   "192.168.1.1",
					Interface:   "",
					Metric:      1,
					Weight:      100,
					Description: "",
					Hide:        false,
				},
			},
			wantErr: false,
		},
		{
			name: "route with all options",
			input: []byte(`ip route 192.168.200.0/24 gateway 192.168.1.100 interface wan1 metric 10 weight 200
exit`),
			expected: []StaticRoute{
				{
					Destination: "192.168.200.0/24",
					GatewayIP:   "192.168.1.100",
					Interface:   "wan1",
					Metric:      10,
					Weight:      200,
					Description: "",
					Hide:        false,
				},
			},
			wantErr: false,
		},
		{
			name: "multiple routes",
			input: []byte(`ip route default gateway 192.168.1.1
ip route 10.0.0.0/8 gateway pp1
ip route 172.16.0.0/12 gateway 192.168.1.254 metric 100
exit`),
			expected: []StaticRoute{
				{
					Destination: "0.0.0.0/0",
					GatewayIP:   "192.168.1.1",
					Interface:   "",
					Metric:      1,
					Weight:      0,
					Description: "",
					Hide:        false,
				},
				{
					Destination:      "10.0.0.0/8",
					GatewayIP:        "",
					GatewayInterface: "pp1",
					Interface:        "",
					Metric:           1,
					Weight:           0,
					Description:      "",
					Hide:             false,
				},
				{
					Destination: "172.16.0.0/12",
					GatewayIP:   "192.168.1.254",
					Interface:   "",
					Metric:      100,
					Weight:      0,
					Description: "",
					Hide:        false,
				},
			},
			wantErr: false,
		},
		{
			name: "RTX830 format variations",
			input: []byte(`ip route default gateway 192.168.1.1
ip route 10.0.0.0/8 gateway interface pp1
ip route 192.168.100.0/24 gateway 192.168.1.254
exit`),
			expected: []StaticRoute{
				{
					Destination: "0.0.0.0/0",
					GatewayIP:   "192.168.1.1",
					Interface:   "",
					Metric:      1,
					Weight:      0,
					Description: "",
					Hide:        false,
				},
				{
					Destination:      "10.0.0.0/8",
					GatewayIP:        "",
					GatewayInterface: "pp1",
					Interface:        "",
					Metric:           1,
					Weight:           0,
					Description:      "",
					Hide:             false,
				},
				{
					Destination: "192.168.100.0/24",
					GatewayIP:   "192.168.1.254",
					Interface:   "",
					Metric:      1,
					Weight:      0,
					Description: "",
					Hide:        false,
				},
			},
			wantErr: false,
		},
		{
			name:     "empty input",
			input:    []byte(""),
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "no matching routes",
			input:    []byte("show version\nexit"),
			expected: nil,
			wantErr:  false,
		},
		{
			name: "invalid CIDR",
			input: []byte(`ip route 999.999.999.999/99 gateway 192.168.1.1
exit`),
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid gateway IP",
			input: []byte(`ip route 192.168.1.0/24 gateway 999.999.999.999
exit`),
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid metric",
			input: []byte(`ip route 192.168.1.0/24 gateway 192.168.1.1 metric abc
exit`),
			expected: nil,
			wantErr:  true,
		},
		{
			name: "invalid weight",
			input: []byte(`ip route 192.168.1.0/24 gateway 192.168.1.1 weight xyz
exit`),
			expected: nil,
			wantErr:  true,
		},
		{
			name: "missing gateway",
			input: []byte(`ip route 192.168.1.0/24
exit`),
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseStaticRoutes(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseStaticRoutes() expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseStaticRoutes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseStaticRoutes() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParseStaticRoute(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *StaticRoute
		wantErr  bool
	}{
		{
			name:  "default route",
			input: "ip route default gateway 192.168.1.1",
			expected: &StaticRoute{
				Destination: "0.0.0.0/0",
				GatewayIP:   "192.168.1.1",
				Interface:   "",
				Metric:      1,
				Weight:      0,
				Description: "",
				Hide:        false,
			},
			wantErr: false,
		},
		{
			name:  "interface gateway with 'interface' keyword",
			input: "ip route 10.0.0.0/8 gateway interface pp1",
			expected: &StaticRoute{
				Destination:      "10.0.0.0/8",
				GatewayIP:        "",
				GatewayInterface: "pp1",
				Interface:        "",
				Metric:           1,
				Weight:           0,
				Description:      "",
				Hide:             false,
			},
			wantErr: false,
		},
		{
			name:  "interface gateway without 'interface' keyword",
			input: "ip route 10.0.0.0/8 gateway pp1",
			expected: &StaticRoute{
				Destination:      "10.0.0.0/8",
				GatewayIP:        "",
				GatewayInterface: "pp1",
				Interface:        "",
				Metric:           1,
				Weight:           0,
				Description:      "",
				Hide:             false,
			},
			wantErr: false,
		},
		{
			name:  "complex route with all options",
			input: "ip route 192.168.100.0/24 gateway 10.0.0.1 interface lan2 metric 50 weight 100",
			expected: &StaticRoute{
				Destination: "192.168.100.0/24",
				GatewayIP:   "10.0.0.1",
				Interface:   "lan2",
				Metric:      50,
				Weight:      100,
				Description: "",
				Hide:        false,
			},
			wantErr: false,
		},
		{
			name:     "invalid route format",
			input:    "invalid route command",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseStaticRoute(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseStaticRoute() expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseStaticRoute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseStaticRoute() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestBuildStaticRouteCommand(t *testing.T) {
	tests := []struct {
		name     string
		route    StaticRoute
		expected string
	}{
		{
			name: "default route with IP gateway",
			route: StaticRoute{
				Destination: "0.0.0.0/0",
				GatewayIP:   "192.168.1.1",
			},
			expected: "ip route default gateway 192.168.1.1",
		},
		{
			name: "route with IP gateway and interface",
			route: StaticRoute{
				Destination: "192.168.100.0/24",
				GatewayIP:   "192.168.1.254",
				Interface:   "wan1",
			},
			expected: "ip route 192.168.100.0/24 gateway 192.168.1.254 interface wan1",
		},
		{
			name: "route with interface gateway",
			route: StaticRoute{
				Destination:      "10.0.0.0/8",
				GatewayInterface: "pp1",
			},
			expected: "ip route 10.0.0.0/8 gateway pp1",
		},
		{
			name: "route with metric",
			route: StaticRoute{
				Destination: "172.16.0.0/12",
				GatewayIP:   "192.168.1.1",
				Metric:      100,
			},
			expected: "ip route 172.16.0.0/12 gateway 192.168.1.1 metric 100",
		},
		{
			name: "route with weight",
			route: StaticRoute{
				Destination: "10.10.0.0/16",
				GatewayIP:   "192.168.1.1",
				Weight:      200,
			},
			expected: "ip route 10.10.0.0/16 gateway 192.168.1.1 weight 200",
		},
		{
			name: "route with all options",
			route: StaticRoute{
				Destination: "192.168.200.0/24",
				GatewayIP:   "192.168.1.100",
				Interface:   "wan1",
				Metric:      10,
				Weight:      50,
			},
			expected: "ip route 192.168.200.0/24 gateway 192.168.1.100 interface wan1 metric 10 weight 50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildStaticRouteCommand(tt.route)
			if result != tt.expected {
				t.Errorf("BuildStaticRouteCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildStaticRouteDeleteCommand(t *testing.T) {
	tests := []struct {
		name     string
		route    StaticRoute
		expected string
	}{
		{
			name: "delete default route",
			route: StaticRoute{
				Destination: "0.0.0.0/0",
				GatewayIP:   "192.168.1.1",
			},
			expected: "no ip route default gateway 192.168.1.1",
		},
		{
			name: "delete route with interface gateway",
			route: StaticRoute{
				Destination:      "10.0.0.0/8",
				GatewayInterface: "pp1",
			},
			expected: "no ip route 10.0.0.0/8 gateway pp1",
		},
		{
			name: "delete route with IP gateway and interface",
			route: StaticRoute{
				Destination: "192.168.100.0/24",
				GatewayIP:   "192.168.1.254",
				Interface:   "wan1",
			},
			expected: "no ip route 192.168.100.0/24 gateway 192.168.1.254 interface wan1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildStaticRouteDeleteCommand(tt.route)
			if result != tt.expected {
				t.Errorf("BuildStaticRouteDeleteCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValidateStaticRoute(t *testing.T) {
	tests := []struct {
		name    string
		route   StaticRoute
		wantErr bool
	}{
		{
			name: "valid route with IP gateway",
			route: StaticRoute{
				Destination: "192.168.1.0/24",
				GatewayIP:   "192.168.1.1",
			},
			wantErr: false,
		},
		{
			name: "valid route with interface gateway",
			route: StaticRoute{
				Destination:      "10.0.0.0/8",
				GatewayInterface: "pp1",
			},
			wantErr: false,
		},
		{
			name: "invalid destination CIDR",
			route: StaticRoute{
				Destination: "invalid-cidr",
				GatewayIP:   "192.168.1.1",
			},
			wantErr: true,
		},
		{
			name: "invalid gateway IP",
			route: StaticRoute{
				Destination: "192.168.1.0/24",
				GatewayIP:   "invalid-ip",
			},
			wantErr: true,
		},
		{
			name: "no gateway specified",
			route: StaticRoute{
				Destination: "192.168.1.0/24",
			},
			wantErr: true,
		},
		{
			name: "both gateway IP and interface specified",
			route: StaticRoute{
				Destination:      "192.168.1.0/24",
				GatewayIP:        "192.168.1.1",
				GatewayInterface: "pp1",
			},
			wantErr: true,
		},
		{
			name: "invalid metric range",
			route: StaticRoute{
				Destination: "192.168.1.0/24",
				GatewayIP:   "192.168.1.1",
				Metric:      70000, // > 65535
			},
			wantErr: true,
		},
		{
			name: "invalid weight range",
			route: StaticRoute{
				Destination: "192.168.1.0/24",
				GatewayIP:   "192.168.1.1",
				Weight:      300, // > 255
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStaticRoute(tt.route)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStaticRoute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}