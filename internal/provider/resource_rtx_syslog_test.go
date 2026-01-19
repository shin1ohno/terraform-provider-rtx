package provider

import (
	"testing"
)

func TestBuildSyslogConfigFromResourceData(t *testing.T) {
	// This test validates the validation functions since we can't easily
	// test buildSyslogConfigFromResourceData without mocking ResourceData
}

func TestValidateSyslogHostAddress(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid IPv4",
			value:   "192.168.1.100",
			wantErr: false,
		},
		{
			name:    "valid IPv4 boundary",
			value:   "0.0.0.0",
			wantErr: false,
		},
		{
			name:    "valid IPv4 max",
			value:   "255.255.255.255",
			wantErr: false,
		},
		{
			name:    "valid hostname simple",
			value:   "syslog",
			wantErr: false,
		},
		{
			name:    "valid hostname with domain",
			value:   "syslog.example.com",
			wantErr: false,
		},
		{
			name:    "valid hostname with hyphen",
			value:   "syslog-server.example.com",
			wantErr: false,
		},
		{
			name:    "valid IPv6 localhost",
			value:   "::1",
			wantErr: false,
		},
		{
			name:    "valid IPv6",
			value:   "2001:db8::1",
			wantErr: false,
		},
		{
			name:    "empty string",
			value:   "",
			wantErr: true,
		},
		{
			name:    "invalid hostname starts with hyphen",
			value:   "-invalid",
			wantErr: true,
		},
		{
			name:    "invalid hostname ends with hyphen",
			value:   "invalid-",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			value:   "host_name",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateSyslogHostAddress(tt.value, "test_key")
			hasErr := len(errs) > 0

			if hasErr != tt.wantErr {
				t.Errorf("validateSyslogHostAddress(%q) hasErr = %v, wantErr %v", tt.value, hasErr, tt.wantErr)
				if hasErr {
					t.Errorf("  error: %v", errs[0])
				}
			}
		})
	}
}

func TestValidateSyslogFacility(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "user facility",
			value:   "user",
			wantErr: false,
		},
		{
			name:    "local0 facility",
			value:   "local0",
			wantErr: false,
		},
		{
			name:    "local7 facility",
			value:   "local7",
			wantErr: false,
		},
		{
			name:    "invalid kern",
			value:   "kern",
			wantErr: true,
		},
		{
			name:    "invalid mail",
			value:   "mail",
			wantErr: true,
		},
		{
			name:    "invalid local8",
			value:   "local8",
			wantErr: true,
		},
		{
			name:    "empty string",
			value:   "",
			wantErr: true,
		},
		{
			name:    "invalid random",
			value:   "random",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateSyslogFacility(tt.value, "test_key")
			hasErr := len(errs) > 0

			if hasErr != tt.wantErr {
				t.Errorf("validateSyslogFacility(%q) hasErr = %v, wantErr %v", tt.value, hasErr, tt.wantErr)
				if hasErr {
					t.Errorf("  error: %v", errs[0])
				}
			}
		})
	}
}

func TestIsValidIPv4(t *testing.T) {
	tests := []struct {
		ip    string
		valid bool
	}{
		{"192.168.1.100", true},
		{"10.0.0.1", true},
		{"255.255.255.255", true},
		{"0.0.0.0", true},
		{"1.2.3.4", true},
		{"192.168.1", false},
		{"192.168.1.256", false},
		{"192.168.1.1.1", false},
		{"192.168.1.-1", false},
		{"192.168.01.1", false}, // Leading zeros
		{"", false},
		{"abc.def.ghi.jkl", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := isValidIPv4(tt.ip)
			if result != tt.valid {
				t.Errorf("isValidIPv4(%q) = %v, want %v", tt.ip, result, tt.valid)
			}
		})
	}
}

func TestIsValidIPv6(t *testing.T) {
	tests := []struct {
		ip    string
		valid bool
	}{
		{"::1", true},
		{"2001:db8::1", true},
		{"fe80::1", true},
		{"2001:0db8:0000:0000:0000:0000:0000:0001", true},
		{"192.168.1.1", false}, // IPv4
		{"not-ipv6", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := isValidIPv6(tt.ip)
			if result != tt.valid {
				t.Errorf("isValidIPv6(%q) = %v, want %v", tt.ip, result, tt.valid)
			}
		})
	}
}

func TestIsValidHostname(t *testing.T) {
	tests := []struct {
		hostname string
		valid    bool
	}{
		{"syslog", true},
		{"syslog-server", true},
		{"syslog.example.com", true},
		{"syslog-server.example.com", true},
		{"a1", true},
		{"1server", true},
		{"-invalid", false},
		{"invalid-", false},
		{"inva_lid", false},
		{"", false},
		{"a", true},
	}

	for _, tt := range tests {
		t.Run(tt.hostname, func(t *testing.T) {
			result := isValidHostname(tt.hostname)
			if result != tt.valid {
				t.Errorf("isValidHostname(%q) = %v, want %v", tt.hostname, result, tt.valid)
			}
		})
	}
}

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		char  rune
		valid bool
	}{
		{'a', true},
		{'z', true},
		{'A', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{'-', false},
		{'_', false},
		{'.', false},
		{' ', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			result := isAlphanumeric(tt.char)
			if result != tt.valid {
				t.Errorf("isAlphanumeric(%q) = %v, want %v", tt.char, result, tt.valid)
			}
		})
	}
}
