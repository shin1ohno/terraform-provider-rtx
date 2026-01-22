package parsers

import (
	"strings"
	"testing"
)

func TestParseSyslogConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *SyslogConfig
		wantErr  bool
	}{
		{
			name:  "single host without port",
			input: `syslog host 192.168.1.100`,
			expected: &SyslogConfig{
				Hosts: []SyslogHost{
					{Address: "192.168.1.100", Port: 0},
				},
			},
		},
		{
			name:  "single host with port",
			input: `syslog host 192.168.1.100 1514`,
			expected: &SyslogConfig{
				Hosts: []SyslogHost{
					{Address: "192.168.1.100", Port: 1514},
				},
			},
		},
		{
			name: "multiple hosts",
			input: `syslog host 192.168.1.100
syslog host 192.168.1.101 1514`,
			expected: &SyslogConfig{
				Hosts: []SyslogHost{
					{Address: "192.168.1.100", Port: 0},
					{Address: "192.168.1.101", Port: 1514},
				},
			},
		},
		{
			name:  "host with hostname",
			input: `syslog host syslog.example.com`,
			expected: &SyslogConfig{
				Hosts: []SyslogHost{
					{Address: "syslog.example.com", Port: 0},
				},
			},
		},
		{
			name:  "local address only",
			input: `syslog local address 192.168.1.1`,
			expected: &SyslogConfig{
				Hosts:        []SyslogHost{},
				LocalAddress: "192.168.1.1",
			},
		},
		{
			name:  "facility only",
			input: `syslog facility local0`,
			expected: &SyslogConfig{
				Hosts:    []SyslogHost{},
				Facility: "local0",
			},
		},
		{
			name:  "facility user",
			input: `syslog facility user`,
			expected: &SyslogConfig{
				Hosts:    []SyslogHost{},
				Facility: "user",
			},
		},
		{
			name:  "notice on",
			input: `syslog notice on`,
			expected: &SyslogConfig{
				Hosts:  []SyslogHost{},
				Notice: true,
			},
		},
		{
			name:  "notice off",
			input: `syslog notice off`,
			expected: &SyslogConfig{
				Hosts:  []SyslogHost{},
				Notice: false,
			},
		},
		{
			name:  "info on",
			input: `syslog info on`,
			expected: &SyslogConfig{
				Hosts: []SyslogHost{},
				Info:  true,
			},
		},
		{
			name:  "debug on",
			input: `syslog debug on`,
			expected: &SyslogConfig{
				Hosts: []SyslogHost{},
				Debug: true,
			},
		},
		{
			name: "all log levels enabled",
			input: `syslog notice on
syslog info on
syslog debug on`,
			expected: &SyslogConfig{
				Hosts:  []SyslogHost{},
				Notice: true,
				Info:   true,
				Debug:  true,
			},
		},
		{
			name: "full configuration",
			input: `syslog host 192.168.1.100
syslog host 192.168.1.101 1514
syslog local address 192.168.1.1
syslog facility local0
syslog notice on
syslog info on
syslog debug off`,
			expected: &SyslogConfig{
				Hosts: []SyslogHost{
					{Address: "192.168.1.100", Port: 0},
					{Address: "192.168.1.101", Port: 1514},
				},
				LocalAddress: "192.168.1.1",
				Facility:     "local0",
				Notice:       true,
				Info:         true,
				Debug:        false,
			},
		},
		{
			name:  "empty input",
			input: "",
			expected: &SyslogConfig{
				Hosts: []SyslogHost{},
			},
		},
		{
			name:  "whitespace only",
			input: "   \n\n   \n",
			expected: &SyslogConfig{
				Hosts: []SyslogHost{},
			},
		},
		{
			name: "mixed with other config lines",
			input: `some other config
syslog host 192.168.1.100
another config line
syslog notice on`,
			expected: &SyslogConfig{
				Hosts: []SyslogHost{
					{Address: "192.168.1.100", Port: 0},
				},
				Notice: true,
			},
		},
	}

	parser := NewSyslogParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseSyslogConfig(tt.input)

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

			// Check hosts
			if len(result.Hosts) != len(tt.expected.Hosts) {
				t.Errorf("hosts count = %d, want %d", len(result.Hosts), len(tt.expected.Hosts))
			} else {
				for i, host := range result.Hosts {
					if host.Address != tt.expected.Hosts[i].Address {
						t.Errorf("hosts[%d].address = %q, want %q", i, host.Address, tt.expected.Hosts[i].Address)
					}
					if host.Port != tt.expected.Hosts[i].Port {
						t.Errorf("hosts[%d].port = %d, want %d", i, host.Port, tt.expected.Hosts[i].Port)
					}
				}
			}

			// Check local address
			if result.LocalAddress != tt.expected.LocalAddress {
				t.Errorf("local_address = %q, want %q", result.LocalAddress, tt.expected.LocalAddress)
			}

			// Check facility
			if result.Facility != tt.expected.Facility {
				t.Errorf("facility = %q, want %q", result.Facility, tt.expected.Facility)
			}

			// Check notice
			if result.Notice != tt.expected.Notice {
				t.Errorf("notice = %v, want %v", result.Notice, tt.expected.Notice)
			}

			// Check info
			if result.Info != tt.expected.Info {
				t.Errorf("info = %v, want %v", result.Info, tt.expected.Info)
			}

			// Check debug
			if result.Debug != tt.expected.Debug {
				t.Errorf("debug = %v, want %v", result.Debug, tt.expected.Debug)
			}
		})
	}
}

func TestBuildSyslogHostCommand(t *testing.T) {
	tests := []struct {
		name     string
		host     SyslogHost
		expected string
	}{
		{
			name:     "host without port",
			host:     SyslogHost{Address: "192.168.1.100", Port: 0},
			expected: "syslog host 192.168.1.100",
		},
		{
			name:     "host with default port",
			host:     SyslogHost{Address: "192.168.1.100", Port: 514},
			expected: "syslog host 192.168.1.100",
		},
		{
			name:     "host with custom port",
			host:     SyslogHost{Address: "192.168.1.100", Port: 1514},
			expected: "syslog host 192.168.1.100 1514",
		},
		{
			name:     "hostname without port",
			host:     SyslogHost{Address: "syslog.example.com", Port: 0},
			expected: "syslog host syslog.example.com",
		},
		{
			name:     "hostname with custom port",
			host:     SyslogHost{Address: "syslog.example.com", Port: 5514},
			expected: "syslog host syslog.example.com 5514",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildSyslogHostCommand(tt.host)
			if result != tt.expected {
				t.Errorf("BuildSyslogHostCommand(%+v) = %q, want %q", tt.host, result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteSyslogHostCommand(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected string
	}{
		{
			name:     "IP address",
			address:  "192.168.1.100",
			expected: "no syslog host 192.168.1.100",
		},
		{
			name:     "hostname",
			address:  "syslog.example.com",
			expected: "no syslog host syslog.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteSyslogHostCommand(tt.address)
			if result != tt.expected {
				t.Errorf("BuildDeleteSyslogHostCommand(%q) = %q, want %q", tt.address, result, tt.expected)
			}
		})
	}
}

func TestBuildSyslogLocalAddressCommand(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected string
	}{
		{
			name:     "simple IP",
			address:  "192.168.1.1",
			expected: "syslog local address 192.168.1.1",
		},
		{
			name:     "another IP",
			address:  "10.0.0.1",
			expected: "syslog local address 10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildSyslogLocalAddressCommand(tt.address)
			if result != tt.expected {
				t.Errorf("BuildSyslogLocalAddressCommand(%q) = %q, want %q", tt.address, result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteSyslogLocalAddressCommand(t *testing.T) {
	result := BuildDeleteSyslogLocalAddressCommand()
	expected := "no syslog local address"
	if result != expected {
		t.Errorf("BuildDeleteSyslogLocalAddressCommand() = %q, want %q", result, expected)
	}
}

func TestBuildSyslogFacilityCommand(t *testing.T) {
	tests := []struct {
		name     string
		facility string
		expected string
	}{
		{
			name:     "user facility",
			facility: "user",
			expected: "syslog facility user",
		},
		{
			name:     "local0 facility",
			facility: "local0",
			expected: "syslog facility local0",
		},
		{
			name:     "local7 facility",
			facility: "local7",
			expected: "syslog facility local7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildSyslogFacilityCommand(tt.facility)
			if result != tt.expected {
				t.Errorf("BuildSyslogFacilityCommand(%q) = %q, want %q", tt.facility, result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteSyslogFacilityCommand(t *testing.T) {
	result := BuildDeleteSyslogFacilityCommand()
	expected := "no syslog facility"
	if result != expected {
		t.Errorf("BuildDeleteSyslogFacilityCommand() = %q, want %q", result, expected)
	}
}

func TestBuildSyslogLevelCommands(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name:     "notice on",
			fn:       func() string { return BuildSyslogNoticeCommand(true) },
			expected: "syslog notice on",
		},
		{
			name:     "notice off",
			fn:       func() string { return BuildSyslogNoticeCommand(false) },
			expected: "syslog notice off",
		},
		{
			name:     "info on",
			fn:       func() string { return BuildSyslogInfoCommand(true) },
			expected: "syslog info on",
		},
		{
			name:     "info off",
			fn:       func() string { return BuildSyslogInfoCommand(false) },
			expected: "syslog info off",
		},
		{
			name:     "debug on",
			fn:       func() string { return BuildSyslogDebugCommand(true) },
			expected: "syslog debug on",
		},
		{
			name:     "debug off",
			fn:       func() string { return BuildSyslogDebugCommand(false) },
			expected: "syslog debug off",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteSyslogCommand(t *testing.T) {
	config := &SyslogConfig{
		Hosts: []SyslogHost{
			{Address: "192.168.1.100", Port: 0},
			{Address: "192.168.1.101", Port: 1514},
		},
		LocalAddress: "192.168.1.1",
		Facility:     "local0",
		Notice:       true,
		Info:         true,
		Debug:        true,
	}

	commands := BuildDeleteSyslogCommand(config)

	expected := []string{
		"no syslog host 192.168.1.100",
		"no syslog host 192.168.1.101",
		"no syslog local address",
		"no syslog facility",
		"syslog notice off",
		"syslog info off",
		"syslog debug off",
	}

	if len(commands) != len(expected) {
		t.Errorf("command count = %d, want %d", len(commands), len(expected))
		return
	}

	for i, cmd := range commands {
		if cmd != expected[i] {
			t.Errorf("commands[%d] = %q, want %q", i, cmd, expected[i])
		}
	}
}

func TestBuildDeleteSyslogCommand_Minimal(t *testing.T) {
	config := &SyslogConfig{
		Hosts:  []SyslogHost{},
		Notice: false,
		Info:   false,
		Debug:  false,
	}

	commands := BuildDeleteSyslogCommand(config)

	// Should be empty since nothing is configured
	if len(commands) != 0 {
		t.Errorf("expected empty commands, got %v", commands)
	}
}

func TestBuildShowSyslogConfigCommand(t *testing.T) {
	result := BuildShowSyslogConfigCommand()
	expected := `show config | grep syslog`
	if result != expected {
		t.Errorf("BuildShowSyslogConfigCommand() = %q, want %q", result, expected)
	}
}

func TestValidateSyslogConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *SyslogConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid full config",
			config: &SyslogConfig{
				Hosts: []SyslogHost{
					{Address: "192.168.1.100", Port: 514},
					{Address: "syslog.example.com", Port: 1514},
				},
				LocalAddress: "192.168.1.1",
				Facility:     "local0",
				Notice:       true,
				Info:         true,
				Debug:        false,
			},
			wantErr: false,
		},
		{
			name: "valid minimal config",
			config: &SyslogConfig{
				Hosts: []SyslogHost{
					{Address: "192.168.1.100", Port: 0},
				},
			},
			wantErr: false,
		},
		{
			name: "valid empty config",
			config: &SyslogConfig{
				Hosts: []SyslogHost{},
			},
			wantErr: false,
		},
		{
			name: "valid host with hostname",
			config: &SyslogConfig{
				Hosts: []SyslogHost{
					{Address: "syslog.example.com", Port: 0},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid host - empty address",
			config: &SyslogConfig{
				Hosts: []SyslogHost{
					{Address: "", Port: 514},
				},
			},
			wantErr: true,
			errMsg:  "address is required",
		},
		{
			name: "invalid host - bad port",
			config: &SyslogConfig{
				Hosts: []SyslogHost{
					{Address: "192.168.1.100", Port: 70000},
				},
			},
			wantErr: true,
			errMsg:  "port must be between 0 and 65535",
		},
		{
			name: "invalid host - negative port",
			config: &SyslogConfig{
				Hosts: []SyslogHost{
					{Address: "192.168.1.100", Port: -1},
				},
			},
			wantErr: true,
			errMsg:  "port must be between 0 and 65535",
		},
		{
			name: "invalid local address",
			config: &SyslogConfig{
				Hosts:        []SyslogHost{},
				LocalAddress: "invalid-ip",
			},
			wantErr: true,
			errMsg:  "invalid local_address format",
		},
		{
			name: "invalid facility",
			config: &SyslogConfig{
				Hosts:    []SyslogHost{},
				Facility: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid facility",
		},
		{
			name: "invalid facility - kern",
			config: &SyslogConfig{
				Hosts:    []SyslogHost{},
				Facility: "kern",
			},
			wantErr: true,
			errMsg:  "invalid facility",
		},
		{
			name: "valid all facilities",
			config: &SyslogConfig{
				Hosts:    []SyslogHost{},
				Facility: "user",
			},
			wantErr: false,
		},
		{
			name: "valid host address with dots",
			config: &SyslogConfig{
				Hosts: []SyslogHost{
					{Address: "192.168.1", Port: 0}, // Valid as hostname (looks like internal DNS)
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSyslogConfig(tt.config)

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

func TestIsValidSyslogHost(t *testing.T) {
	tests := []struct {
		host  string
		valid bool
	}{
		{"192.168.1.100", true},
		{"10.0.0.1", true},
		{"255.255.255.255", true},
		{"0.0.0.0", true},
		{"syslog.example.com", true},
		{"syslog-server", true},
		{"syslog1", true},
		{"log.local", true},
		{"::1", true},         // IPv6 localhost
		{"2001:db8::1", true}, // IPv6
		{"192.168.1", true},   // Looks like a hostname with dots (could be internal DNS)
		{"-invalid", false},   // Invalid hostname (starts with hyphen)
		{"", false},           // Empty
		{"with space", false}, // Invalid hostname
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			result := isValidSyslogHost(tt.host)
			if result != tt.valid {
				t.Errorf("isValidSyslogHost(%q) = %v, want %v", tt.host, result, tt.valid)
			}
		})
	}
}

func TestIsValidFacility(t *testing.T) {
	tests := []struct {
		facility string
		valid    bool
	}{
		{"user", true},
		{"local0", true},
		{"local1", true},
		{"local2", true},
		{"local3", true},
		{"local4", true},
		{"local5", true},
		{"local6", true},
		{"local7", true},
		{"kern", false},
		{"mail", false},
		{"daemon", false},
		{"auth", false},
		{"local8", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.facility, func(t *testing.T) {
			result := isValidFacility(tt.facility)
			if result != tt.valid {
				t.Errorf("isValidFacility(%q) = %v, want %v", tt.facility, result, tt.valid)
			}
		})
	}
}
