package parsers

import (
	"strings"
	"testing"
)

func TestParseIPFilterConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []IPFilter
		wantErr  bool
	}{
		{
			name:  "basic pass all filter",
			input: "ip filter 100 pass * * * * *",
			expected: []IPFilter{
				{
					Number:        100,
					Action:        "pass",
					SourceAddress: "*",
					DestAddress:   "*",
					Protocol:      "*",
					SourcePort:    "*",
					DestPort:      "*",
				},
			},
		},
		{
			name:  "reject filter with network",
			input: "ip filter 101 reject 192.168.1.0/24 * tcp * www",
			expected: []IPFilter{
				{
					Number:        101,
					Action:        "reject",
					SourceAddress: "192.168.1.0/24",
					DestAddress:   "*",
					Protocol:      "tcp",
					SourcePort:    "*",
					DestPort:      "www",
				},
			},
		},
		{
			name:  "filter with established",
			input: "ip filter 102 pass * * tcp * * established",
			expected: []IPFilter{
				{
					Number:        102,
					Action:        "pass",
					SourceAddress: "*",
					DestAddress:   "*",
					Protocol:      "tcp",
					SourcePort:    "*",
					DestPort:      "*",
					Established:   true,
				},
			},
		},
		{
			name:  "minimal filter",
			input: "ip filter 103 pass * * icmp",
			expected: []IPFilter{
				{
					Number:        103,
					Action:        "pass",
					SourceAddress: "*",
					DestAddress:   "*",
					Protocol:      "icmp",
				},
			},
		},
		{
			name: "multiple filters",
			input: `ip filter 100 pass * * * * *
ip filter 101 reject 10.0.0.0/8 * tcp * 22
ip filter 102 pass 192.168.1.0/24 * * * *`,
			expected: []IPFilter{
				{
					Number:        100,
					Action:        "pass",
					SourceAddress: "*",
					DestAddress:   "*",
					Protocol:      "*",
					SourcePort:    "*",
					DestPort:      "*",
				},
				{
					Number:        101,
					Action:        "reject",
					SourceAddress: "10.0.0.0/8",
					DestAddress:   "*",
					Protocol:      "tcp",
					SourcePort:    "*",
					DestPort:      "22",
				},
				{
					Number:        102,
					Action:        "pass",
					SourceAddress: "192.168.1.0/24",
					DestAddress:   "*",
					Protocol:      "*",
					SourcePort:    "*",
					DestPort:      "*",
				},
			},
		},
		{
			name: "skip dynamic and secure filter lines",
			input: `ip filter 100 pass * * * * *
ip filter dynamic 10 * * ftp
ip lan1 secure filter in 100`,
			expected: []IPFilter{
				{
					Number:        100,
					Action:        "pass",
					SourceAddress: "*",
					DestAddress:   "*",
					Protocol:      "*",
					SourcePort:    "*",
					DestPort:      "*",
				},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []IPFilter{},
		},
		{
			name:     "only comments",
			input:    "# This is a comment\n# Another comment",
			expected: []IPFilter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseIPFilterConfig(tt.input)

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
				if got.SourceAddress != expected.SourceAddress {
					t.Errorf("filter[%d].SourceAddress = %q, want %q", i, got.SourceAddress, expected.SourceAddress)
				}
				if got.DestAddress != expected.DestAddress {
					t.Errorf("filter[%d].DestAddress = %q, want %q", i, got.DestAddress, expected.DestAddress)
				}
				if got.Protocol != expected.Protocol {
					t.Errorf("filter[%d].Protocol = %q, want %q", i, got.Protocol, expected.Protocol)
				}
				if got.SourcePort != expected.SourcePort {
					t.Errorf("filter[%d].SourcePort = %q, want %q", i, got.SourcePort, expected.SourcePort)
				}
				if got.DestPort != expected.DestPort {
					t.Errorf("filter[%d].DestPort = %q, want %q", i, got.DestPort, expected.DestPort)
				}
				if got.Established != expected.Established {
					t.Errorf("filter[%d].Established = %v, want %v", i, got.Established, expected.Established)
				}
			}
		})
	}
}

func TestParseIPFilterDynamicConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []IPFilterDynamic
		wantErr  bool
	}{
		{
			name:  "basic dynamic filter",
			input: "ip filter dynamic 10 * * ftp",
			expected: []IPFilterDynamic{
				{
					Number:   10,
					Source:   "*",
					Dest:     "*",
					Protocol: "ftp",
				},
			},
		},
		{
			name:  "dynamic filter with syslog",
			input: "ip filter dynamic 20 * * www syslog on",
			expected: []IPFilterDynamic{
				{
					Number:   20,
					Source:   "*",
					Dest:     "*",
					Protocol: "www",
					SyslogOn: true,
				},
			},
		},
		{
			name:  "dynamic filter with source network",
			input: "ip filter dynamic 30 192.168.1.0/24 * smtp",
			expected: []IPFilterDynamic{
				{
					Number:   30,
					Source:   "192.168.1.0/24",
					Dest:     "*",
					Protocol: "smtp",
				},
			},
		},
		{
			name: "multiple dynamic filters",
			input: `ip filter dynamic 10 * * ftp
ip filter dynamic 20 * * www
ip filter dynamic 30 * * smtp syslog on`,
			expected: []IPFilterDynamic{
				{Number: 10, Source: "*", Dest: "*", Protocol: "ftp"},
				{Number: 20, Source: "*", Dest: "*", Protocol: "www"},
				{Number: 30, Source: "*", Dest: "*", Protocol: "smtp", SyslogOn: true},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []IPFilterDynamic{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseIPFilterDynamicConfig(tt.input)

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
				if got.Source != expected.Source {
					t.Errorf("filter[%d].Source = %q, want %q", i, got.Source, expected.Source)
				}
				if got.Dest != expected.Dest {
					t.Errorf("filter[%d].Dest = %q, want %q", i, got.Dest, expected.Dest)
				}
				if got.Protocol != expected.Protocol {
					t.Errorf("filter[%d].Protocol = %q, want %q", i, got.Protocol, expected.Protocol)
				}
				if got.SyslogOn != expected.SyslogOn {
					t.Errorf("filter[%d].SyslogOn = %v, want %v", i, got.SyslogOn, expected.SyslogOn)
				}
			}
		})
	}
}

func TestParseInterfaceSecureFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]map[string][]int
		wantErr  bool
	}{
		{
			name:  "single interface inbound",
			input: "ip lan1 secure filter in 100 101 102",
			expected: map[string]map[string][]int{
				"lan1": {"in": {100, 101, 102}},
			},
		},
		{
			name:  "single interface outbound",
			input: "ip lan1 secure filter out 200 201",
			expected: map[string]map[string][]int{
				"lan1": {"out": {200, 201}},
			},
		},
		{
			name: "multiple interfaces",
			input: `ip lan1 secure filter in 100 101
ip lan2 secure filter in 200 201
ip pp1 secure filter out 300`,
			expected: map[string]map[string][]int{
				"lan1": {"in": {100, 101}},
				"lan2": {"in": {200, 201}},
				"pp1":  {"out": {300}},
			},
		},
		{
			name: "interface with both directions",
			input: `ip lan1 secure filter in 100 101
ip lan1 secure filter out 200 201`,
			expected: map[string]map[string][]int{
				"lan1": {"in": {100, 101}, "out": {200, 201}},
			},
		},
		{
			name:  "filter with dynamic keyword",
			input: "ip lan1 secure filter in 100 101 dynamic 10 20",
			expected: map[string]map[string][]int{
				"lan1": {"in": {100, 101}}, // dynamic numbers are not included in static list
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
			result, err := ParseInterfaceSecureFilter(tt.input)

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
				t.Errorf("expected %d interfaces, got %d", len(tt.expected), len(result))
				return
			}

			for iface, expectedDirs := range tt.expected {
				gotDirs, ok := result[iface]
				if !ok {
					t.Errorf("interface %q not found in result", iface)
					continue
				}

				for dir, expectedNums := range expectedDirs {
					gotNums, ok := gotDirs[dir]
					if !ok {
						t.Errorf("direction %q not found for interface %q", dir, iface)
						continue
					}

					if len(gotNums) != len(expectedNums) {
						t.Errorf("interface %q direction %q: expected %d filters, got %d",
							iface, dir, len(expectedNums), len(gotNums))
						continue
					}

					for i, num := range expectedNums {
						if gotNums[i] != num {
							t.Errorf("interface %q direction %q filter[%d] = %d, want %d",
								iface, dir, i, gotNums[i], num)
						}
					}
				}
			}
		})
	}
}

func TestBuildIPFilterCommand(t *testing.T) {
	tests := []struct {
		name     string
		filter   IPFilter
		expected string
	}{
		{
			name: "pass all filter",
			filter: IPFilter{
				Number:        100,
				Action:        "pass",
				SourceAddress: "*",
				DestAddress:   "*",
				Protocol:      "*",
				SourcePort:    "*",
				DestPort:      "*",
			},
			expected: "ip filter 100 pass * * * * *",
		},
		{
			name: "reject with network and port",
			filter: IPFilter{
				Number:        101,
				Action:        "reject",
				SourceAddress: "192.168.1.0/24",
				DestAddress:   "*",
				Protocol:      "tcp",
				SourcePort:    "*",
				DestPort:      "www",
			},
			expected: "ip filter 101 reject 192.168.1.0/24 * tcp * www",
		},
		{
			name: "filter with established",
			filter: IPFilter{
				Number:        102,
				Action:        "pass",
				SourceAddress: "*",
				DestAddress:   "*",
				Protocol:      "tcp",
				SourcePort:    "*",
				DestPort:      "*",
				Established:   true,
			},
			expected: "ip filter 102 pass * * tcp * * established",
		},
		{
			name: "minimal filter without ports",
			filter: IPFilter{
				Number:        103,
				Action:        "pass",
				SourceAddress: "*",
				DestAddress:   "*",
				Protocol:      "icmp",
			},
			expected: "ip filter 103 pass * * icmp",
		},
		{
			name: "filter with only dest port",
			filter: IPFilter{
				Number:        104,
				Action:        "pass",
				SourceAddress: "*",
				DestAddress:   "*",
				Protocol:      "tcp",
				DestPort:      "22",
			},
			expected: "ip filter 104 pass * * tcp * 22",
		},
		{
			name: "restrict-log action",
			filter: IPFilter{
				Number:        105,
				Action:        "restrict-log",
				SourceAddress: "10.0.0.0/8",
				DestAddress:   "*",
				Protocol:      "ip",
			},
			expected: "ip filter 105 restrict-log 10.0.0.0/8 * ip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildIPFilterCommand(tt.filter)
			if result != tt.expected {
				t.Errorf("BuildIPFilterCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildIPFilterDynamicCommand(t *testing.T) {
	tests := []struct {
		name     string
		filter   IPFilterDynamic
		expected string
	}{
		{
			name: "basic dynamic filter",
			filter: IPFilterDynamic{
				Number:   10,
				Source:   "*",
				Dest:     "*",
				Protocol: "ftp",
			},
			expected: "ip filter dynamic 10 * * ftp",
		},
		{
			name: "dynamic filter with syslog",
			filter: IPFilterDynamic{
				Number:   20,
				Source:   "*",
				Dest:     "*",
				Protocol: "www",
				SyslogOn: true,
			},
			expected: "ip filter dynamic 20 * * www syslog on",
		},
		{
			name: "dynamic filter with source network",
			filter: IPFilterDynamic{
				Number:   30,
				Source:   "192.168.1.0/24",
				Dest:     "*",
				Protocol: "smtp",
			},
			expected: "ip filter dynamic 30 192.168.1.0/24 * smtp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildIPFilterDynamicCommand(tt.filter)
			if result != tt.expected {
				t.Errorf("BuildIPFilterDynamicCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteIPFilterCommand(t *testing.T) {
	tests := []struct {
		name     string
		number   int
		expected string
	}{
		{
			name:     "delete filter 100",
			number:   100,
			expected: "no ip filter 100",
		},
		{
			name:     "delete filter 65535",
			number:   65535,
			expected: "no ip filter 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteIPFilterCommand(tt.number)
			if result != tt.expected {
				t.Errorf("BuildDeleteIPFilterCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteIPFilterDynamicCommand(t *testing.T) {
	tests := []struct {
		name     string
		number   int
		expected string
	}{
		{
			name:     "delete dynamic filter 10",
			number:   10,
			expected: "no ip filter dynamic 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteIPFilterDynamicCommand(tt.number)
			if result != tt.expected {
				t.Errorf("BuildDeleteIPFilterDynamicCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildInterfaceSecureFilterCommand(t *testing.T) {
	tests := []struct {
		name       string
		iface      string
		direction  string
		filterNums []int
		expected   string
	}{
		{
			name:       "single filter inbound",
			iface:      "lan1",
			direction:  "in",
			filterNums: []int{100},
			expected:   "ip lan1 secure filter in 100",
		},
		{
			name:       "multiple filters inbound",
			iface:      "lan1",
			direction:  "in",
			filterNums: []int{100, 101, 102},
			expected:   "ip lan1 secure filter in 100 101 102",
		},
		{
			name:       "pp interface outbound",
			iface:      "pp1",
			direction:  "out",
			filterNums: []int{200, 201},
			expected:   "ip pp1 secure filter out 200 201",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildInterfaceSecureFilterCommand(tt.iface, tt.direction, tt.filterNums)
			if result != tt.expected {
				t.Errorf("BuildInterfaceSecureFilterCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildInterfaceSecureFilterWithDynamicCommand(t *testing.T) {
	tests := []struct {
		name        string
		iface       string
		direction   string
		staticNums  []int
		dynamicNums []int
		expected    string
	}{
		{
			name:        "static only",
			iface:       "lan1",
			direction:   "in",
			staticNums:  []int{100, 101},
			dynamicNums: []int{},
			expected:    "ip lan1 secure filter in 100 101",
		},
		{
			name:        "static and dynamic",
			iface:       "lan1",
			direction:   "in",
			staticNums:  []int{100, 101},
			dynamicNums: []int{10, 20},
			expected:    "ip lan1 secure filter in 100 101 dynamic 10 20",
		},
		{
			name:        "multiple dynamic filters",
			iface:       "pp1",
			direction:   "out",
			staticNums:  []int{200},
			dynamicNums: []int{30, 40, 50},
			expected:    "ip pp1 secure filter out 200 dynamic 30 40 50",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildInterfaceSecureFilterWithDynamicCommand(tt.iface, tt.direction, tt.staticNums, tt.dynamicNums)
			if result != tt.expected {
				t.Errorf("BuildInterfaceSecureFilterWithDynamicCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteInterfaceSecureFilterCommand(t *testing.T) {
	tests := []struct {
		name      string
		iface     string
		direction string
		expected  string
	}{
		{
			name:      "delete inbound filter",
			iface:     "lan1",
			direction: "in",
			expected:  "no ip lan1 secure filter in",
		},
		{
			name:      "delete outbound filter",
			iface:     "pp1",
			direction: "out",
			expected:  "no ip pp1 secure filter out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteInterfaceSecureFilterCommand(tt.iface, tt.direction)
			if result != tt.expected {
				t.Errorf("BuildDeleteInterfaceSecureFilterCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildShowIPFilterCommand(t *testing.T) {
	result := BuildShowIPFilterCommand()
	expected := `show config | grep "ip filter"`
	if result != expected {
		t.Errorf("BuildShowIPFilterCommand() = %q, want %q", result, expected)
	}
}

func TestBuildShowIPFilterByNumberCommand(t *testing.T) {
	result := BuildShowIPFilterByNumberCommand(100)
	expected := `show config | grep "ip filter 100"`
	if result != expected {
		t.Errorf("BuildShowIPFilterByNumberCommand() = %q, want %q", result, expected)
	}
}

func TestValidateIPFilterNumber(t *testing.T) {
	tests := []struct {
		name    string
		number  int
		wantErr bool
		errMsg  string
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
			number:  32768,
			wantErr: false,
		},
		{
			name:    "zero",
			number:  0,
			wantErr: true,
			errMsg:  "filter number must be between 1 and 65535",
		},
		{
			name:    "negative",
			number:  -1,
			wantErr: true,
			errMsg:  "filter number must be between 1 and 65535",
		},
		{
			name:    "too large",
			number:  65536,
			wantErr: true,
			errMsg:  "filter number must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPFilterNumber(tt.number)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateIPFilterProtocol(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		wantErr  bool
	}{
		{name: "tcp", protocol: "tcp", wantErr: false},
		{name: "udp", protocol: "udp", wantErr: false},
		{name: "icmp", protocol: "icmp", wantErr: false},
		{name: "ip", protocol: "ip", wantErr: false},
		{name: "wildcard", protocol: "*", wantErr: false},
		{name: "gre", protocol: "gre", wantErr: false},
		{name: "esp", protocol: "esp", wantErr: false},
		{name: "ah", protocol: "ah", wantErr: false},
		{name: "icmp6", protocol: "icmp6", wantErr: false},
		{name: "uppercase TCP", protocol: "TCP", wantErr: false},
		{name: "invalid protocol", protocol: "invalid", wantErr: true},
		{name: "empty", protocol: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPFilterProtocol(tt.protocol)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for protocol %q, got nil", tt.protocol)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for protocol %q: %v", tt.protocol, err)
				}
			}
		})
	}
}

func TestValidateIPFilterAction(t *testing.T) {
	tests := []struct {
		name    string
		action  string
		wantErr bool
	}{
		{name: "pass", action: "pass", wantErr: false},
		{name: "reject", action: "reject", wantErr: false},
		{name: "restrict", action: "restrict", wantErr: false},
		{name: "restrict-log", action: "restrict-log", wantErr: false},
		{name: "uppercase PASS", action: "PASS", wantErr: false},
		{name: "invalid action", action: "allow", wantErr: true},
		{name: "deny", action: "deny", wantErr: true},
		{name: "empty", action: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPFilterAction(tt.action)
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

func TestValidateIPFilter(t *testing.T) {
	tests := []struct {
		name    string
		filter  IPFilter
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid filter",
			filter: IPFilter{
				Number:        100,
				Action:        "pass",
				SourceAddress: "*",
				DestAddress:   "*",
				Protocol:      "tcp",
			},
			wantErr: false,
		},
		{
			name: "valid filter with established",
			filter: IPFilter{
				Number:        100,
				Action:        "pass",
				SourceAddress: "*",
				DestAddress:   "*",
				Protocol:      "tcp",
				Established:   true,
			},
			wantErr: false,
		},
		{
			name: "invalid number",
			filter: IPFilter{
				Number:        0,
				Action:        "pass",
				SourceAddress: "*",
				DestAddress:   "*",
				Protocol:      "tcp",
			},
			wantErr: true,
			errMsg:  "filter number must be between 1 and 65535",
		},
		{
			name: "invalid action",
			filter: IPFilter{
				Number:        100,
				Action:        "invalid",
				SourceAddress: "*",
				DestAddress:   "*",
				Protocol:      "tcp",
			},
			wantErr: true,
			errMsg:  "invalid action",
		},
		{
			name: "empty source address",
			filter: IPFilter{
				Number:        100,
				Action:        "pass",
				SourceAddress: "",
				DestAddress:   "*",
				Protocol:      "tcp",
			},
			wantErr: true,
			errMsg:  "source address is required",
		},
		{
			name: "empty dest address",
			filter: IPFilter{
				Number:        100,
				Action:        "pass",
				SourceAddress: "*",
				DestAddress:   "",
				Protocol:      "tcp",
			},
			wantErr: true,
			errMsg:  "destination address is required",
		},
		{
			name: "invalid protocol",
			filter: IPFilter{
				Number:        100,
				Action:        "pass",
				SourceAddress: "*",
				DestAddress:   "*",
				Protocol:      "invalid",
			},
			wantErr: true,
			errMsg:  "invalid protocol",
		},
		{
			name: "established with non-tcp",
			filter: IPFilter{
				Number:        100,
				Action:        "pass",
				SourceAddress: "*",
				DestAddress:   "*",
				Protocol:      "udp",
				Established:   true,
			},
			wantErr: true,
			errMsg:  "established keyword can only be used with TCP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPFilter(tt.filter)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateIPFilterDynamic(t *testing.T) {
	tests := []struct {
		name    string
		filter  IPFilterDynamic
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid dynamic filter",
			filter: IPFilterDynamic{
				Number:   10,
				Source:   "*",
				Dest:     "*",
				Protocol: "ftp",
			},
			wantErr: false,
		},
		{
			name: "valid with syslog",
			filter: IPFilterDynamic{
				Number:   20,
				Source:   "*",
				Dest:     "*",
				Protocol: "www",
				SyslogOn: true,
			},
			wantErr: false,
		},
		{
			name: "invalid number",
			filter: IPFilterDynamic{
				Number:   0,
				Source:   "*",
				Dest:     "*",
				Protocol: "ftp",
			},
			wantErr: true,
			errMsg:  "filter number must be between 1 and 65535",
		},
		{
			name: "empty source",
			filter: IPFilterDynamic{
				Number:   10,
				Source:   "",
				Dest:     "*",
				Protocol: "ftp",
			},
			wantErr: true,
			errMsg:  "source is required",
		},
		{
			name: "empty dest",
			filter: IPFilterDynamic{
				Number:   10,
				Source:   "*",
				Dest:     "",
				Protocol: "ftp",
			},
			wantErr: true,
			errMsg:  "destination is required",
		},
		{
			name: "empty protocol",
			filter: IPFilterDynamic{
				Number:   10,
				Source:   "*",
				Dest:     "*",
				Protocol: "",
			},
			wantErr: true,
			errMsg:  "protocol is required",
		},
		{
			name: "invalid dynamic protocol",
			filter: IPFilterDynamic{
				Number:   10,
				Source:   "*",
				Dest:     "*",
				Protocol: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid dynamic protocol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPFilterDynamic(tt.filter)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateIPFilterDirection(t *testing.T) {
	tests := []struct {
		name      string
		direction string
		wantErr   bool
	}{
		{name: "in", direction: "in", wantErr: false},
		{name: "out", direction: "out", wantErr: false},
		{name: "uppercase IN", direction: "IN", wantErr: false},
		{name: "uppercase OUT", direction: "OUT", wantErr: false},
		{name: "inbound", direction: "inbound", wantErr: true},
		{name: "outbound", direction: "outbound", wantErr: true},
		{name: "empty", direction: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPFilterDirection(tt.direction)
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
