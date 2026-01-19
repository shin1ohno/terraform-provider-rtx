package parsers

import (
	"reflect"
	"testing"
)

func TestParseHTTPDConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *HTTPDConfig
	}{
		{
			name:  "empty config",
			input: "",
			expected: &HTTPDConfig{
				Host:        "",
				ProxyAccess: false,
			},
		},
		{
			name:  "host any",
			input: "httpd host any",
			expected: &HTTPDConfig{
				Host:        "any",
				ProxyAccess: false,
			},
		},
		{
			name:  "host specific interface",
			input: "httpd host lan1",
			expected: &HTTPDConfig{
				Host:        "lan1",
				ProxyAccess: false,
			},
		},
		{
			name:  "proxy access enabled",
			input: "httpd proxy-access l2ms permit on",
			expected: &HTTPDConfig{
				Host:        "",
				ProxyAccess: true,
			},
		},
		{
			name:  "proxy access disabled",
			input: "httpd proxy-access l2ms permit off",
			expected: &HTTPDConfig{
				Host:        "",
				ProxyAccess: false,
			},
		},
		{
			name: "full configuration",
			input: `httpd host lan1
httpd proxy-access l2ms permit on`,
			expected: &HTTPDConfig{
				Host:        "lan1",
				ProxyAccess: true,
			},
		},
		{
			name: "with comments and whitespace",
			input: `# HTTPD configuration
  httpd host any
  httpd proxy-access l2ms permit on
`,
			expected: &HTTPDConfig{
				Host:        "any",
				ProxyAccess: true,
			},
		},
	}

	parser := NewServiceParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseHTTPDConfig(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("got %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParseSSHDConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *SSHDConfig
	}{
		{
			name:  "empty config",
			input: "",
			expected: &SSHDConfig{
				Enabled: false,
				Hosts:   []string{},
				HostKey: "",
			},
		},
		{
			name:  "service on",
			input: "sshd service on",
			expected: &SSHDConfig{
				Enabled: true,
				Hosts:   []string{},
				HostKey: "",
			},
		},
		{
			name:  "service off",
			input: "sshd service off",
			expected: &SSHDConfig{
				Enabled: false,
				Hosts:   []string{},
				HostKey: "",
			},
		},
		{
			name:  "single host",
			input: "sshd host lan1",
			expected: &SSHDConfig{
				Enabled: false,
				Hosts:   []string{"lan1"},
				HostKey: "",
			},
		},
		{
			name:  "multiple hosts",
			input: "sshd host lan1 lan2 pp1",
			expected: &SSHDConfig{
				Enabled: false,
				Hosts:   []string{"lan1", "lan2", "pp1"},
				HostKey: "",
			},
		},
		{
			name: "full configuration",
			input: `sshd service on
sshd host lan1 lan2`,
			expected: &SSHDConfig{
				Enabled: true,
				Hosts:   []string{"lan1", "lan2"},
				HostKey: "",
			},
		},
		{
			name: "with host key",
			input: `sshd service on
sshd host lan1
sshd host key AAAAB3NzaC1yc2EAAAADAQABAAABgQC...`,
			expected: &SSHDConfig{
				Enabled: true,
				Hosts:   []string{"lan1"},
				HostKey: "AAAAB3NzaC1yc2EAAAADAQABAAABgQC...",
			},
		},
		{
			name: "host key generate (not stored)",
			input: `sshd service on
sshd host key generate`,
			expected: &SSHDConfig{
				Enabled: true,
				Hosts:   []string{},
				HostKey: "",
			},
		},
	}

	parser := NewServiceParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseSSHDConfig(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("got %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParseSFTPDConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *SFTPDConfig
	}{
		{
			name:  "empty config",
			input: "",
			expected: &SFTPDConfig{
				Hosts: []string{},
			},
		},
		{
			name:  "single host",
			input: "sftpd host lan1",
			expected: &SFTPDConfig{
				Hosts: []string{"lan1"},
			},
		},
		{
			name:  "multiple hosts",
			input: "sftpd host lan1 lan2",
			expected: &SFTPDConfig{
				Hosts: []string{"lan1", "lan2"},
			},
		},
		{
			name: "with whitespace",
			input: `  sftpd host lan1 lan2
`,
			expected: &SFTPDConfig{
				Hosts: []string{"lan1", "lan2"},
			},
		},
	}

	parser := NewServiceParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseSFTPDConfig(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("got %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestBuildHTTPDCommands(t *testing.T) {
	tests := []struct {
		name     string
		function func() string
		expected string
	}{
		{
			name:     "host any",
			function: func() string { return BuildHTTPDHostCommand("any") },
			expected: "httpd host any",
		},
		{
			name:     "host lan1",
			function: func() string { return BuildHTTPDHostCommand("lan1") },
			expected: "httpd host lan1",
		},
		{
			name:     "host empty",
			function: func() string { return BuildHTTPDHostCommand("") },
			expected: "",
		},
		{
			name:     "proxy access on",
			function: func() string { return BuildHTTPDProxyAccessCommand(true) },
			expected: "httpd proxy-access l2ms permit on",
		},
		{
			name:     "proxy access off",
			function: func() string { return BuildHTTPDProxyAccessCommand(false) },
			expected: "httpd proxy-access l2ms permit off",
		},
		{
			name:     "delete host",
			function: BuildDeleteHTTPDHostCommand,
			expected: "no httpd host",
		},
		{
			name:     "delete proxy access",
			function: BuildDeleteHTTPDProxyAccessCommand,
			expected: "httpd proxy-access l2ms permit off",
		},
		{
			name:     "show config",
			function: BuildShowHTTPDConfigCommand,
			expected: "show config | grep httpd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function()
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildSSHDCommands(t *testing.T) {
	tests := []struct {
		name     string
		function func() string
		expected string
	}{
		{
			name:     "service on",
			function: func() string { return BuildSSHDServiceCommand(true) },
			expected: "sshd service on",
		},
		{
			name:     "service off",
			function: func() string { return BuildSSHDServiceCommand(false) },
			expected: "sshd service off",
		},
		{
			name:     "host single",
			function: func() string { return BuildSSHDHostCommand([]string{"lan1"}) },
			expected: "sshd host lan1",
		},
		{
			name:     "host multiple",
			function: func() string { return BuildSSHDHostCommand([]string{"lan1", "lan2", "pp1"}) },
			expected: "sshd host lan1 lan2 pp1",
		},
		{
			name:     "host empty",
			function: func() string { return BuildSSHDHostCommand([]string{}) },
			expected: "",
		},
		{
			name:     "host key generate",
			function: BuildSSHDHostKeyGenerateCommand,
			expected: "sshd host key generate",
		},
		{
			name:     "delete service",
			function: BuildDeleteSSHDServiceCommand,
			expected: "no sshd service",
		},
		{
			name:     "delete host",
			function: BuildDeleteSSHDHostCommand,
			expected: "no sshd host",
		},
		{
			name:     "show config",
			function: BuildShowSSHDConfigCommand,
			expected: "show config | grep sshd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function()
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildSFTPDCommands(t *testing.T) {
	tests := []struct {
		name     string
		function func() string
		expected string
	}{
		{
			name:     "host single",
			function: func() string { return BuildSFTPDHostCommand([]string{"lan1"}) },
			expected: "sftpd host lan1",
		},
		{
			name:     "host multiple",
			function: func() string { return BuildSFTPDHostCommand([]string{"lan1", "lan2"}) },
			expected: "sftpd host lan1 lan2",
		},
		{
			name:     "host empty",
			function: func() string { return BuildSFTPDHostCommand([]string{}) },
			expected: "",
		},
		{
			name:     "delete host",
			function: BuildDeleteSFTPDHostCommand,
			expected: "no sftpd host",
		},
		{
			name:     "show config",
			function: BuildShowSFTPDConfigCommand,
			expected: "show config | grep sftpd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function()
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValidateHTTPDConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  HTTPDConfig
		wantErr bool
	}{
		{
			name:    "valid any",
			config:  HTTPDConfig{Host: "any", ProxyAccess: false},
			wantErr: false,
		},
		{
			name:    "valid lan1",
			config:  HTTPDConfig{Host: "lan1", ProxyAccess: true},
			wantErr: false,
		},
		{
			name:    "valid pp1",
			config:  HTTPDConfig{Host: "pp1", ProxyAccess: false},
			wantErr: false,
		},
		{
			name:    "valid bridge1",
			config:  HTTPDConfig{Host: "bridge1", ProxyAccess: false},
			wantErr: false,
		},
		{
			name:    "valid tunnel1",
			config:  HTTPDConfig{Host: "tunnel1", ProxyAccess: false},
			wantErr: false,
		},
		{
			name:    "empty host",
			config:  HTTPDConfig{Host: "", ProxyAccess: false},
			wantErr: true,
		},
		{
			name:    "invalid host",
			config:  HTTPDConfig{Host: "invalid", ProxyAccess: false},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHTTPDConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHTTPDConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSSHDConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  SSHDConfig
		wantErr bool
	}{
		{
			name:    "valid single host",
			config:  SSHDConfig{Enabled: true, Hosts: []string{"lan1"}},
			wantErr: false,
		},
		{
			name:    "valid multiple hosts",
			config:  SSHDConfig{Enabled: true, Hosts: []string{"lan1", "lan2", "pp1"}},
			wantErr: false,
		},
		{
			name:    "empty hosts",
			config:  SSHDConfig{Enabled: true, Hosts: []string{}},
			wantErr: false, // Empty hosts is valid (SSH service on but no host restriction)
		},
		{
			name:    "invalid interface",
			config:  SSHDConfig{Enabled: true, Hosts: []string{"invalid"}},
			wantErr: true,
		},
		{
			name:    "mixed valid and invalid",
			config:  SSHDConfig{Enabled: true, Hosts: []string{"lan1", "invalid"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSSHDConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSSHDConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSFTPDConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  SFTPDConfig
		wantErr bool
	}{
		{
			name:    "valid single host",
			config:  SFTPDConfig{Hosts: []string{"lan1"}},
			wantErr: false,
		},
		{
			name:    "valid multiple hosts",
			config:  SFTPDConfig{Hosts: []string{"lan1", "lan2"}},
			wantErr: false,
		},
		{
			name:    "empty hosts",
			config:  SFTPDConfig{Hosts: []string{}},
			wantErr: true, // SFTPD requires at least one host
		},
		{
			name:    "invalid interface",
			config:  SFTPDConfig{Hosts: []string{"invalid"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSFTPDConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSFTPDConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
