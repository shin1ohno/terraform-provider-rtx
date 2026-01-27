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
		{
			name:  "restrict-nolog action",
			input: "ip filter 110 restrict-nolog 192.168.0.0/16 * ip",
			expected: []IPFilter{
				{
					Number:        110,
					Action:        "restrict-nolog",
					SourceAddress: "192.168.0.0/16",
					DestAddress:   "*",
					Protocol:      "ip",
				},
			},
		},
		{
			name:  "tcpfin protocol filter",
			input: "ip filter 111 pass * * tcpfin",
			expected: []IPFilter{
				{
					Number:        111,
					Action:        "pass",
					SourceAddress: "*",
					DestAddress:   "*",
					Protocol:      "tcpfin",
				},
			},
		},
		{
			name:  "tcprst protocol filter",
			input: "ip filter 112 reject * * tcprst",
			expected: []IPFilter{
				{
					Number:        112,
					Action:        "reject",
					SourceAddress: "*",
					DestAddress:   "*",
					Protocol:      "tcprst",
				},
			},
		},
		{
			name:  "tcpsyn protocol filter",
			input: "ip filter 113 reject 0.0.0.0/0 * tcpsyn",
			expected: []IPFilter{
				{
					Number:        113,
					Action:        "reject",
					SourceAddress: "0.0.0.0/0",
					DestAddress:   "*",
					Protocol:      "tcpsyn",
				},
			},
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
		{
			name: "restrict-nolog action",
			filter: IPFilter{
				Number:        106,
				Action:        "restrict-nolog",
				SourceAddress: "10.0.0.0/8",
				DestAddress:   "*",
				Protocol:      "ip",
			},
			expected: "ip filter 106 restrict-nolog 10.0.0.0/8 * ip",
		},
		{
			name: "tcpfin protocol",
			filter: IPFilter{
				Number:        107,
				Action:        "pass",
				SourceAddress: "*",
				DestAddress:   "*",
				Protocol:      "tcpfin",
			},
			expected: "ip filter 107 pass * * tcpfin",
		},
		{
			name: "tcprst protocol",
			filter: IPFilter{
				Number:        108,
				Action:        "reject",
				SourceAddress: "*",
				DestAddress:   "*",
				Protocol:      "tcprst",
			},
			expected: "ip filter 108 reject * * tcprst",
		},
		{
			name: "tcpsyn protocol",
			filter: IPFilter{
				Number:        109,
				Action:        "reject",
				SourceAddress: "*",
				DestAddress:   "*",
				Protocol:      "tcpsyn",
			},
			expected: "ip filter 109 reject * * tcpsyn",
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
			errMsg:  "filter number must be between 1 and 2147483647",
		},
		{
			name:    "negative",
			number:  -1,
			wantErr: true,
			errMsg:  "filter number must be between 1 and 2147483647",
		},
		{
			name:    "large valid (200000)",
			number:  200000,
			wantErr: false,
		},
		{
			name:    "large valid (500000)",
			number:  500000,
			wantErr: false,
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
		{name: "tcpfin", protocol: "tcpfin", wantErr: false},
		{name: "tcprst", protocol: "tcprst", wantErr: false},
		{name: "tcpsyn", protocol: "tcpsyn", wantErr: false},
		{name: "established", protocol: "established", wantErr: false},
		{name: "uppercase TCP", protocol: "TCP", wantErr: false},
		{name: "invalid protocol", protocol: "invalid", wantErr: true},
		{name: "empty", protocol: "", wantErr: true},
		// Compound protocol tests (RTX supports comma-separated protocols)
		{name: "compound udp,tcp", protocol: "udp,tcp", wantErr: false},
		{name: "compound tcp,udp", protocol: "tcp,udp", wantErr: false},
		{name: "compound with spaces", protocol: "udp, tcp", wantErr: false},
		{name: "compound uppercase", protocol: "UDP,TCP", wantErr: false},
		{name: "compound three protocols", protocol: "tcp,udp,icmp", wantErr: false},
		{name: "compound with invalid", protocol: "tcp,invalid", wantErr: true},
		{name: "compound all invalid", protocol: "invalid1,invalid2", wantErr: true},
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
		{name: "restrict-nolog", action: "restrict-nolog", wantErr: false},
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
			errMsg:  "filter number must be between 1 and 2147483647",
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
			errMsg:  "filter number must be between 1 and 2147483647",
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

// ============================================================================
// Tests for Extended Dynamic IP Filter Parser and Builder
// ============================================================================

func TestParseIPFilterDynamicConfigExtended(t *testing.T) {
	timeout60 := 60
	timeout120 := 120
	timeout3600 := 3600

	tests := []struct {
		name     string
		input    string
		expected []IPFilterDynamic
		wantErr  bool
	}{
		// Form 1: Protocol-based dynamic filters
		{
			name:  "Form 1 - basic ftp protocol",
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
			name:  "Form 1 - www protocol",
			input: "ip filter dynamic 20 * * www",
			expected: []IPFilterDynamic{
				{
					Number:   20,
					Source:   "*",
					Dest:     "*",
					Protocol: "www",
				},
			},
		},
		{
			name:  "Form 1 - smtp protocol with network",
			input: "ip filter dynamic 30 192.168.1.0/24 10.0.0.0/8 smtp",
			expected: []IPFilterDynamic{
				{
					Number:   30,
					Source:   "192.168.1.0/24",
					Dest:     "10.0.0.0/8",
					Protocol: "smtp",
				},
			},
		},
		{
			name:  "Form 1 - tcp protocol",
			input: "ip filter dynamic 40 * * tcp",
			expected: []IPFilterDynamic{
				{
					Number:   40,
					Source:   "*",
					Dest:     "*",
					Protocol: "tcp",
				},
			},
		},
		{
			name:  "Form 1 - udp protocol",
			input: "ip filter dynamic 50 * * udp",
			expected: []IPFilterDynamic{
				{
					Number:   50,
					Source:   "*",
					Dest:     "*",
					Protocol: "udp",
				},
			},
		},
		{
			name:  "Form 1 - wildcard protocol",
			input: "ip filter dynamic 60 * * *",
			expected: []IPFilterDynamic{
				{
					Number:   60,
					Source:   "*",
					Dest:     "*",
					Protocol: "*",
				},
			},
		},
		{
			name:  "Form 1 - extended protocol https",
			input: "ip filter dynamic 70 * * https",
			expected: []IPFilterDynamic{
				{
					Number:   70,
					Source:   "*",
					Dest:     "*",
					Protocol: "https",
				},
			},
		},
		{
			name:  "Form 1 - extended protocol sip",
			input: "ip filter dynamic 80 * * sip",
			expected: []IPFilterDynamic{
				{
					Number:   80,
					Source:   "*",
					Dest:     "*",
					Protocol: "sip",
				},
			},
		},
		{
			name:  "Form 1 - extended protocol pptp",
			input: "ip filter dynamic 90 * * pptp",
			expected: []IPFilterDynamic{
				{
					Number:   90,
					Source:   "*",
					Dest:     "*",
					Protocol: "pptp",
				},
			},
		},
		// Form 1 with syslog option
		{
			name:  "Form 1 - with syslog on",
			input: "ip filter dynamic 100 * * ftp syslog on",
			expected: []IPFilterDynamic{
				{
					Number:   100,
					Source:   "*",
					Dest:     "*",
					Protocol: "ftp",
					SyslogOn: true,
				},
			},
		},
		{
			name:  "Form 1 - with syslog off (default)",
			input: "ip filter dynamic 110 * * www syslog off",
			expected: []IPFilterDynamic{
				{
					Number:   110,
					Source:   "*",
					Dest:     "*",
					Protocol: "www",
					SyslogOn: false,
				},
			},
		},
		// Form 1 with timeout option
		{
			name:  "Form 1 - with timeout",
			input: "ip filter dynamic 120 * * tcp timeout=60",
			expected: []IPFilterDynamic{
				{
					Number:   120,
					Source:   "*",
					Dest:     "*",
					Protocol: "tcp",
					Timeout:  &timeout60,
				},
			},
		},
		{
			name:  "Form 1 - with timeout 3600",
			input: "ip filter dynamic 130 * * udp timeout=3600",
			expected: []IPFilterDynamic{
				{
					Number:   130,
					Source:   "*",
					Dest:     "*",
					Protocol: "udp",
					Timeout:  &timeout3600,
				},
			},
		},
		// Form 1 with combined options
		{
			name:  "Form 1 - syslog on and timeout",
			input: "ip filter dynamic 140 * * smtp syslog on timeout=120",
			expected: []IPFilterDynamic{
				{
					Number:   140,
					Source:   "*",
					Dest:     "*",
					Protocol: "smtp",
					SyslogOn: true,
					Timeout:  &timeout120,
				},
			},
		},
		{
			name:  "Form 1 - timeout and syslog on (reverse order)",
			input: "ip filter dynamic 150 * * pop3 timeout=60 syslog on",
			expected: []IPFilterDynamic{
				{
					Number:   150,
					Source:   "*",
					Dest:     "*",
					Protocol: "pop3",
					SyslogOn: true,
					Timeout:  &timeout60,
				},
			},
		},
		// Form 2: Filter-reference form
		{
			name:  "Form 2 - single filter list",
			input: "ip filter dynamic 200 * * filter 100",
			expected: []IPFilterDynamic{
				{
					Number:     200,
					Source:     "*",
					Dest:       "*",
					FilterList: []int{100},
				},
			},
		},
		{
			name:  "Form 2 - multiple filter list",
			input: "ip filter dynamic 210 * * filter 100 101 102",
			expected: []IPFilterDynamic{
				{
					Number:     210,
					Source:     "*",
					Dest:       "*",
					FilterList: []int{100, 101, 102},
				},
			},
		},
		{
			name:  "Form 2 - filter with in list",
			input: "ip filter dynamic 220 * * filter 100 in 200",
			expected: []IPFilterDynamic{
				{
					Number:       220,
					Source:       "*",
					Dest:         "*",
					FilterList:   []int{100},
					InFilterList: []int{200},
				},
			},
		},
		{
			name:  "Form 2 - filter with out list",
			input: "ip filter dynamic 230 * * filter 100 out 300",
			expected: []IPFilterDynamic{
				{
					Number:        230,
					Source:        "*",
					Dest:          "*",
					FilterList:    []int{100},
					OutFilterList: []int{300},
				},
			},
		},
		{
			name:  "Form 2 - filter with in and out lists",
			input: "ip filter dynamic 240 * * filter 100 101 in 200 201 out 300 301",
			expected: []IPFilterDynamic{
				{
					Number:        240,
					Source:        "*",
					Dest:          "*",
					FilterList:    []int{100, 101},
					InFilterList:  []int{200, 201},
					OutFilterList: []int{300, 301},
				},
			},
		},
		{
			name:  "Form 2 - filter with only in and out (no main filter)",
			input: "ip filter dynamic 250 * * filter 100 in 200 202 204 out 300",
			expected: []IPFilterDynamic{
				{
					Number:        250,
					Source:        "*",
					Dest:          "*",
					FilterList:    []int{100},
					InFilterList:  []int{200, 202, 204},
					OutFilterList: []int{300},
				},
			},
		},
		// Form 2 with options
		{
			name:  "Form 2 - filter with syslog on",
			input: "ip filter dynamic 260 * * filter 100 syslog on",
			expected: []IPFilterDynamic{
				{
					Number:     260,
					Source:     "*",
					Dest:       "*",
					FilterList: []int{100},
					SyslogOn:   true,
				},
			},
		},
		{
			name:  "Form 2 - filter with timeout",
			input: "ip filter dynamic 270 * * filter 100 timeout=60",
			expected: []IPFilterDynamic{
				{
					Number:     270,
					Source:     "*",
					Dest:       "*",
					FilterList: []int{100},
					Timeout:    &timeout60,
				},
			},
		},
		{
			name:  "Form 2 - full configuration",
			input: "ip filter dynamic 280 192.168.1.0/24 10.0.0.0/8 filter 100 101 in 200 out 300 syslog on timeout=120",
			expected: []IPFilterDynamic{
				{
					Number:        280,
					Source:        "192.168.1.0/24",
					Dest:          "10.0.0.0/8",
					FilterList:    []int{100, 101},
					InFilterList:  []int{200},
					OutFilterList: []int{300},
					SyslogOn:      true,
					Timeout:       &timeout120,
				},
			},
		},
		// Multiple filters
		{
			name: "Multiple dynamic filters - mixed forms",
			input: `ip filter dynamic 10 * * ftp
ip filter dynamic 20 * * www syslog on
ip filter dynamic 30 * * filter 100 101 in 200 out 300
ip filter dynamic 40 192.168.1.0/24 * tcp timeout=60`,
			expected: []IPFilterDynamic{
				{Number: 10, Source: "*", Dest: "*", Protocol: "ftp"},
				{Number: 20, Source: "*", Dest: "*", Protocol: "www", SyslogOn: true},
				{Number: 30, Source: "*", Dest: "*", FilterList: []int{100, 101}, InFilterList: []int{200}, OutFilterList: []int{300}},
				{Number: 40, Source: "192.168.1.0/24", Dest: "*", Protocol: "tcp", Timeout: &timeout60},
			},
		},
		// Edge cases
		{
			name:     "Empty input",
			input:    "",
			expected: []IPFilterDynamic{},
		},
		{
			name:     "Only comments",
			input:    "# This is a comment\n# Another comment",
			expected: []IPFilterDynamic{},
		},
		{
			name:     "Skip static filter lines",
			input:    "ip filter 100 pass * * tcp",
			expected: []IPFilterDynamic{},
		},
		{
			name: "Mixed static and dynamic filters",
			input: `ip filter 100 pass * * tcp
ip filter dynamic 10 * * ftp
ip lan1 secure filter in 100`,
			expected: []IPFilterDynamic{
				{Number: 10, Source: "*", Dest: "*", Protocol: "ftp"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseIPFilterDynamicConfigExtended(tt.input)

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
				// Compare Timeout
				if (got.Timeout == nil) != (expected.Timeout == nil) {
					t.Errorf("filter[%d].Timeout = %v, want %v", i, got.Timeout, expected.Timeout)
				} else if got.Timeout != nil && *got.Timeout != *expected.Timeout {
					t.Errorf("filter[%d].Timeout = %d, want %d", i, *got.Timeout, *expected.Timeout)
				}
				// Compare FilterList
				if !equalIntSlice(got.FilterList, expected.FilterList) {
					t.Errorf("filter[%d].FilterList = %v, want %v", i, got.FilterList, expected.FilterList)
				}
				// Compare InFilterList
				if !equalIntSlice(got.InFilterList, expected.InFilterList) {
					t.Errorf("filter[%d].InFilterList = %v, want %v", i, got.InFilterList, expected.InFilterList)
				}
				// Compare OutFilterList
				if !equalIntSlice(got.OutFilterList, expected.OutFilterList) {
					t.Errorf("filter[%d].OutFilterList = %v, want %v", i, got.OutFilterList, expected.OutFilterList)
				}
			}
		})
	}
}

func TestBuildIPFilterDynamicCommandExtended(t *testing.T) {
	timeout60 := 60
	timeout120 := 120

	tests := []struct {
		name     string
		filter   IPFilterDynamic
		expected string
	}{
		// Form 1: Protocol-based dynamic filters
		{
			name: "Form 1 - basic ftp protocol",
			filter: IPFilterDynamic{
				Number:   10,
				Source:   "*",
				Dest:     "*",
				Protocol: "ftp",
			},
			expected: "ip filter dynamic 10 * * ftp",
		},
		{
			name: "Form 1 - www protocol with network",
			filter: IPFilterDynamic{
				Number:   20,
				Source:   "192.168.1.0/24",
				Dest:     "10.0.0.0/8",
				Protocol: "www",
			},
			expected: "ip filter dynamic 20 192.168.1.0/24 10.0.0.0/8 www",
		},
		{
			name: "Form 1 - with syslog on",
			filter: IPFilterDynamic{
				Number:   30,
				Source:   "*",
				Dest:     "*",
				Protocol: "smtp",
				SyslogOn: true,
			},
			expected: "ip filter dynamic 30 * * smtp syslog on",
		},
		{
			name: "Form 1 - with timeout",
			filter: IPFilterDynamic{
				Number:   40,
				Source:   "*",
				Dest:     "*",
				Protocol: "tcp",
				Timeout:  &timeout60,
			},
			expected: "ip filter dynamic 40 * * tcp timeout=60",
		},
		{
			name: "Form 1 - with syslog on and timeout",
			filter: IPFilterDynamic{
				Number:   50,
				Source:   "*",
				Dest:     "*",
				Protocol: "udp",
				SyslogOn: true,
				Timeout:  &timeout120,
			},
			expected: "ip filter dynamic 50 * * udp syslog on timeout=120",
		},
		{
			name: "Form 1 - extended protocol https",
			filter: IPFilterDynamic{
				Number:   60,
				Source:   "*",
				Dest:     "*",
				Protocol: "https",
			},
			expected: "ip filter dynamic 60 * * https",
		},
		{
			name: "Form 1 - extended protocol sip",
			filter: IPFilterDynamic{
				Number:   70,
				Source:   "*",
				Dest:     "*",
				Protocol: "sip",
			},
			expected: "ip filter dynamic 70 * * sip",
		},
		{
			name: "Form 1 - extended protocol pptp",
			filter: IPFilterDynamic{
				Number:   80,
				Source:   "*",
				Dest:     "*",
				Protocol: "pptp",
			},
			expected: "ip filter dynamic 80 * * pptp",
		},
		{
			name: "Form 1 - extended protocol ipsec-nat-t",
			filter: IPFilterDynamic{
				Number:   90,
				Source:   "*",
				Dest:     "*",
				Protocol: "ipsec-nat-t",
			},
			expected: "ip filter dynamic 90 * * ipsec-nat-t",
		},
		// Form 2: Filter-reference form
		{
			name: "Form 2 - single filter list",
			filter: IPFilterDynamic{
				Number:     100,
				Source:     "*",
				Dest:       "*",
				FilterList: []int{100},
			},
			expected: "ip filter dynamic 100 * * filter 100",
		},
		{
			name: "Form 2 - multiple filter list",
			filter: IPFilterDynamic{
				Number:     110,
				Source:     "*",
				Dest:       "*",
				FilterList: []int{100, 101, 102},
			},
			expected: "ip filter dynamic 110 * * filter 100 101 102",
		},
		{
			name: "Form 2 - filter with in list",
			filter: IPFilterDynamic{
				Number:       120,
				Source:       "*",
				Dest:         "*",
				FilterList:   []int{100},
				InFilterList: []int{200, 201},
			},
			expected: "ip filter dynamic 120 * * filter 100 in 200 201",
		},
		{
			name: "Form 2 - filter with out list",
			filter: IPFilterDynamic{
				Number:        130,
				Source:        "*",
				Dest:          "*",
				FilterList:    []int{100},
				OutFilterList: []int{300, 301},
			},
			expected: "ip filter dynamic 130 * * filter 100 out 300 301",
		},
		{
			name: "Form 2 - filter with in and out lists",
			filter: IPFilterDynamic{
				Number:        140,
				Source:        "*",
				Dest:          "*",
				FilterList:    []int{100, 101},
				InFilterList:  []int{200},
				OutFilterList: []int{300},
			},
			expected: "ip filter dynamic 140 * * filter 100 101 in 200 out 300",
		},
		{
			name: "Form 2 - filter with syslog on",
			filter: IPFilterDynamic{
				Number:     150,
				Source:     "*",
				Dest:       "*",
				FilterList: []int{100},
				SyslogOn:   true,
			},
			expected: "ip filter dynamic 150 * * filter 100 syslog on",
		},
		{
			name: "Form 2 - filter with timeout",
			filter: IPFilterDynamic{
				Number:     160,
				Source:     "*",
				Dest:       "*",
				FilterList: []int{100},
				Timeout:    &timeout60,
			},
			expected: "ip filter dynamic 160 * * filter 100 timeout=60",
		},
		{
			name: "Form 2 - full configuration",
			filter: IPFilterDynamic{
				Number:        170,
				Source:        "192.168.1.0/24",
				Dest:          "10.0.0.0/8",
				FilterList:    []int{100, 101},
				InFilterList:  []int{200, 201},
				OutFilterList: []int{300, 301},
				SyslogOn:      true,
				Timeout:       &timeout120,
			},
			expected: "ip filter dynamic 170 192.168.1.0/24 10.0.0.0/8 filter 100 101 in 200 201 out 300 301 syslog on timeout=120",
		},
		// Edge cases
		{
			name: "Form 2 - empty in list",
			filter: IPFilterDynamic{
				Number:        180,
				Source:        "*",
				Dest:          "*",
				FilterList:    []int{100},
				InFilterList:  []int{},
				OutFilterList: []int{300},
			},
			expected: "ip filter dynamic 180 * * filter 100 out 300",
		},
		{
			name: "Form 2 - empty out list",
			filter: IPFilterDynamic{
				Number:       190,
				Source:       "*",
				Dest:         "*",
				FilterList:   []int{100},
				InFilterList: []int{200},
			},
			expected: "ip filter dynamic 190 * * filter 100 in 200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildIPFilterDynamicCommandExtended(tt.filter)
			if result != tt.expected {
				t.Errorf("BuildIPFilterDynamicCommandExtended() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestParseAndBuildRoundTripExtended tests that parsing and building produces the same command
func TestParseAndBuildRoundTripExtended(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// Form 1 cases
		{
			name:  "Form 1 - basic ftp",
			input: "ip filter dynamic 10 * * ftp",
		},
		{
			name:  "Form 1 - with network",
			input: "ip filter dynamic 20 192.168.1.0/24 10.0.0.0/8 www",
		},
		{
			name:  "Form 1 - with syslog",
			input: "ip filter dynamic 30 * * smtp syslog on",
		},
		{
			name:  "Form 1 - with timeout",
			input: "ip filter dynamic 40 * * tcp timeout=60",
		},
		{
			name:  "Form 1 - with syslog and timeout",
			input: "ip filter dynamic 50 * * udp syslog on timeout=120",
		},
		// Form 2 cases
		{
			name:  "Form 2 - single filter",
			input: "ip filter dynamic 100 * * filter 100",
		},
		{
			name:  "Form 2 - multiple filters",
			input: "ip filter dynamic 110 * * filter 100 101 102",
		},
		{
			name:  "Form 2 - with in list",
			input: "ip filter dynamic 120 * * filter 100 in 200 201",
		},
		{
			name:  "Form 2 - with out list",
			input: "ip filter dynamic 130 * * filter 100 out 300",
		},
		{
			name:  "Form 2 - full config",
			input: "ip filter dynamic 140 192.168.1.0/24 * filter 100 101 in 200 out 300 syslog on timeout=60",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the input
			filters, err := ParseIPFilterDynamicConfigExtended(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			if len(filters) != 1 {
				t.Fatalf("Expected 1 filter, got %d", len(filters))
			}

			// Build command from parsed filter
			result := BuildIPFilterDynamicCommandExtended(filters[0])

			// Compare (should match original input)
			if result != tt.input {
				t.Errorf("Round-trip failed:\n  input:  %q\n  output: %q", tt.input, result)
			}
		})
	}
}

// Helper function to compare int slices
func equalIntSlice(a, b []int) bool {
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

// ============================================================================
// Tests for IPv6 Dynamic Filter Parser
// ============================================================================

func TestParseIPv6FilterDynamicConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []IPFilterDynamic
		wantErr  bool
	}{
		{
			name:  "basic ftp protocol",
			input: "ipv6 filter dynamic 10 * * ftp",
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
			name:  "domain protocol",
			input: "ipv6 filter dynamic 20 * * domain",
			expected: []IPFilterDynamic{
				{
					Number:   20,
					Source:   "*",
					Dest:     "*",
					Protocol: "domain",
				},
			},
		},
		{
			name:  "www protocol",
			input: "ipv6 filter dynamic 30 * * www",
			expected: []IPFilterDynamic{
				{
					Number:   30,
					Source:   "*",
					Dest:     "*",
					Protocol: "www",
				},
			},
		},
		{
			name:  "smtp protocol",
			input: "ipv6 filter dynamic 40 * * smtp",
			expected: []IPFilterDynamic{
				{
					Number:   40,
					Source:   "*",
					Dest:     "*",
					Protocol: "smtp",
				},
			},
		},
		{
			name:  "pop3 protocol",
			input: "ipv6 filter dynamic 50 * * pop3",
			expected: []IPFilterDynamic{
				{
					Number:   50,
					Source:   "*",
					Dest:     "*",
					Protocol: "pop3",
				},
			},
		},
		{
			name:  "submission protocol",
			input: "ipv6 filter dynamic 60 * * submission",
			expected: []IPFilterDynamic{
				{
					Number:   60,
					Source:   "*",
					Dest:     "*",
					Protocol: "submission",
				},
			},
		},
		{
			name:  "tcp protocol",
			input: "ipv6 filter dynamic 70 * * tcp",
			expected: []IPFilterDynamic{
				{
					Number:   70,
					Source:   "*",
					Dest:     "*",
					Protocol: "tcp",
				},
			},
		},
		{
			name:  "udp protocol",
			input: "ipv6 filter dynamic 80 * * udp",
			expected: []IPFilterDynamic{
				{
					Number:   80,
					Source:   "*",
					Dest:     "*",
					Protocol: "udp",
				},
			},
		},
		{
			name:  "with syslog on",
			input: "ipv6 filter dynamic 100 * * ftp syslog on",
			expected: []IPFilterDynamic{
				{
					Number:   100,
					Source:   "*",
					Dest:     "*",
					Protocol: "ftp",
					SyslogOn: true,
				},
			},
		},
		{
			name:  "with IPv6 source network",
			input: "ipv6 filter dynamic 110 2001:db8::/32 * www",
			expected: []IPFilterDynamic{
				{
					Number:   110,
					Source:   "2001:db8::/32",
					Dest:     "*",
					Protocol: "www",
				},
			},
		},
		{
			name:  "with IPv6 source and dest networks",
			input: "ipv6 filter dynamic 120 2001:db8:1::/48 2001:db8:2::/48 smtp",
			expected: []IPFilterDynamic{
				{
					Number:   120,
					Source:   "2001:db8:1::/48",
					Dest:     "2001:db8:2::/48",
					Protocol: "smtp",
				},
			},
		},
		{
			name: "multiple IPv6 dynamic filters - 8 protocols",
			input: `ipv6 filter dynamic 10 * * ftp
ipv6 filter dynamic 20 * * domain
ipv6 filter dynamic 30 * * www
ipv6 filter dynamic 40 * * smtp
ipv6 filter dynamic 50 * * pop3
ipv6 filter dynamic 60 * * submission
ipv6 filter dynamic 70 * * tcp
ipv6 filter dynamic 80 * * udp`,
			expected: []IPFilterDynamic{
				{Number: 10, Source: "*", Dest: "*", Protocol: "ftp"},
				{Number: 20, Source: "*", Dest: "*", Protocol: "domain"},
				{Number: 30, Source: "*", Dest: "*", Protocol: "www"},
				{Number: 40, Source: "*", Dest: "*", Protocol: "smtp"},
				{Number: 50, Source: "*", Dest: "*", Protocol: "pop3"},
				{Number: 60, Source: "*", Dest: "*", Protocol: "submission"},
				{Number: 70, Source: "*", Dest: "*", Protocol: "tcp"},
				{Number: 80, Source: "*", Dest: "*", Protocol: "udp"},
			},
		},
		{
			name: "multiple filters with syslog options",
			input: `ipv6 filter dynamic 10 * * ftp syslog on
ipv6 filter dynamic 20 * * www
ipv6 filter dynamic 30 * * smtp syslog on`,
			expected: []IPFilterDynamic{
				{Number: 10, Source: "*", Dest: "*", Protocol: "ftp", SyslogOn: true},
				{Number: 20, Source: "*", Dest: "*", Protocol: "www", SyslogOn: false},
				{Number: 30, Source: "*", Dest: "*", Protocol: "smtp", SyslogOn: true},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []IPFilterDynamic{},
		},
		{
			name:     "only comments",
			input:    "# This is a comment\n# Another comment",
			expected: []IPFilterDynamic{},
		},
		{
			name: "skip static IPv6 filter lines",
			input: `ipv6 filter 100 pass * * tcp
ipv6 filter dynamic 10 * * ftp
ipv6 lan1 secure filter in 100`,
			expected: []IPFilterDynamic{
				{Number: 10, Source: "*", Dest: "*", Protocol: "ftp"},
			},
		},
		{
			name: "mixed IPv4 and IPv6 dynamic filters - only IPv6 parsed",
			input: `ip filter dynamic 10 * * ftp
ipv6 filter dynamic 20 * * www
ip filter dynamic 30 * * smtp`,
			expected: []IPFilterDynamic{
				{Number: 20, Source: "*", Dest: "*", Protocol: "www"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseIPv6FilterDynamicConfig(tt.input)

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

func TestBuildIPv6FilterDynamicCommand(t *testing.T) {
	tests := []struct {
		name     string
		filter   IPFilterDynamic
		expected string
	}{
		{
			name: "basic ftp protocol",
			filter: IPFilterDynamic{
				Number:   10,
				Source:   "*",
				Dest:     "*",
				Protocol: "ftp",
			},
			expected: "ipv6 filter dynamic 10 * * ftp",
		},
		{
			name: "domain protocol",
			filter: IPFilterDynamic{
				Number:   20,
				Source:   "*",
				Dest:     "*",
				Protocol: "domain",
			},
			expected: "ipv6 filter dynamic 20 * * domain",
		},
		{
			name: "www protocol with syslog",
			filter: IPFilterDynamic{
				Number:   30,
				Source:   "*",
				Dest:     "*",
				Protocol: "www",
				SyslogOn: true,
			},
			expected: "ipv6 filter dynamic 30 * * www syslog on",
		},
		{
			name: "smtp protocol with IPv6 network",
			filter: IPFilterDynamic{
				Number:   40,
				Source:   "2001:db8::/32",
				Dest:     "*",
				Protocol: "smtp",
			},
			expected: "ipv6 filter dynamic 40 2001:db8::/32 * smtp",
		},
		{
			name: "tcp protocol with both IPv6 networks",
			filter: IPFilterDynamic{
				Number:   50,
				Source:   "2001:db8:1::/48",
				Dest:     "2001:db8:2::/48",
				Protocol: "tcp",
			},
			expected: "ipv6 filter dynamic 50 2001:db8:1::/48 2001:db8:2::/48 tcp",
		},
		{
			name: "submission protocol with syslog",
			filter: IPFilterDynamic{
				Number:   60,
				Source:   "*",
				Dest:     "*",
				Protocol: "submission",
				SyslogOn: true,
			},
			expected: "ipv6 filter dynamic 60 * * submission syslog on",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildIPv6FilterDynamicCommand(tt.filter)
			if result != tt.expected {
				t.Errorf("BuildIPv6FilterDynamicCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteIPv6FilterDynamicCommand(t *testing.T) {
	tests := []struct {
		name     string
		number   int
		expected string
	}{
		{
			name:     "delete dynamic filter 10",
			number:   10,
			expected: "no ipv6 filter dynamic 10",
		},
		{
			name:     "delete dynamic filter 65535",
			number:   65535,
			expected: "no ipv6 filter dynamic 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteIPv6FilterDynamicCommand(tt.number)
			if result != tt.expected {
				t.Errorf("BuildDeleteIPv6FilterDynamicCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestParseAndBuildIPv6DynamicRoundTrip tests that parsing and building produces the same command
func TestParseAndBuildIPv6DynamicRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "basic ftp",
			input: "ipv6 filter dynamic 10 * * ftp",
		},
		{
			name:  "domain protocol",
			input: "ipv6 filter dynamic 20 * * domain",
		},
		{
			name:  "www protocol",
			input: "ipv6 filter dynamic 30 * * www",
		},
		{
			name:  "with syslog",
			input: "ipv6 filter dynamic 40 * * smtp syslog on",
		},
		{
			name:  "with IPv6 network",
			input: "ipv6 filter dynamic 50 2001:db8::/32 * tcp",
		},
		{
			name:  "with both IPv6 networks",
			input: "ipv6 filter dynamic 60 2001:db8:1::/48 2001:db8:2::/48 udp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the input
			filters, err := ParseIPv6FilterDynamicConfig(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			if len(filters) != 1 {
				t.Fatalf("Expected 1 filter, got %d", len(filters))
			}

			// Build command from parsed filter
			result := BuildIPv6FilterDynamicCommand(filters[0])

			// Compare (should match original input)
			if result != tt.input {
				t.Errorf("Round-trip failed:\n  input:  %q\n  output: %q", tt.input, result)
			}
		})
	}
}
