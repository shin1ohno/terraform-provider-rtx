package parsers

import (
	"reflect"
	"strings"
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
				Enabled:    false,
				Hosts:      []string{},
				HostKey:    "",
				AuthMethod: "any",
			},
		},
		{
			name:  "service on",
			input: "sshd service on",
			expected: &SSHDConfig{
				Enabled:    true,
				Hosts:      []string{},
				HostKey:    "",
				AuthMethod: "any",
			},
		},
		{
			name:  "service off",
			input: "sshd service off",
			expected: &SSHDConfig{
				Enabled:    false,
				Hosts:      []string{},
				HostKey:    "",
				AuthMethod: "any",
			},
		},
		{
			name:  "single host",
			input: "sshd host lan1",
			expected: &SSHDConfig{
				Enabled:    false,
				Hosts:      []string{"lan1"},
				HostKey:    "",
				AuthMethod: "any",
			},
		},
		{
			name:  "multiple hosts",
			input: "sshd host lan1 lan2 pp1",
			expected: &SSHDConfig{
				Enabled:    false,
				Hosts:      []string{"lan1", "lan2", "pp1"},
				HostKey:    "",
				AuthMethod: "any",
			},
		},
		{
			name: "full configuration",
			input: `sshd service on
sshd host lan1 lan2`,
			expected: &SSHDConfig{
				Enabled:    true,
				Hosts:      []string{"lan1", "lan2"},
				HostKey:    "",
				AuthMethod: "any",
			},
		},
		{
			name: "with host key",
			input: `sshd service on
sshd host lan1
sshd host key AAAAB3NzaC1yc2EAAAADAQABAAABgQC...`,
			expected: &SSHDConfig{
				Enabled:    true,
				Hosts:      []string{"lan1"},
				HostKey:    "AAAAB3NzaC1yc2EAAAADAQABAAABgQC...",
				AuthMethod: "any",
			},
		},
		{
			name: "host key generate (not stored)",
			input: `sshd service on
sshd host key generate`,
			expected: &SSHDConfig{
				Enabled:    true,
				Hosts:      []string{},
				HostKey:    "",
				AuthMethod: "any",
			},
		},
		{
			name:  "auth method password",
			input: "sshd auth method password",
			expected: &SSHDConfig{
				Enabled:    false,
				Hosts:      []string{},
				HostKey:    "",
				AuthMethod: "password",
			},
		},
		{
			name:  "auth method publickey",
			input: "sshd auth method publickey",
			expected: &SSHDConfig{
				Enabled:    false,
				Hosts:      []string{},
				HostKey:    "",
				AuthMethod: "publickey",
			},
		},
		{
			name: "full configuration with auth method",
			input: `sshd service on
sshd host lan1
sshd auth method publickey`,
			expected: &SSHDConfig{
				Enabled:    true,
				Hosts:      []string{"lan1"},
				HostKey:    "",
				AuthMethod: "publickey",
			},
		},
		{
			name:  "auth method with whitespace",
			input: `  sshd auth method password  `,
			expected: &SSHDConfig{
				Enabled:    false,
				Hosts:      []string{},
				HostKey:    "",
				AuthMethod: "password",
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

func TestBuildShowSSHDStatusCommand(t *testing.T) {
	result := BuildShowSSHDStatusCommand()
	expected := "show sshd host key"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestBuildSSHDAuthMethodCommand(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected string
	}{
		{
			name:     "password method",
			method:   "password",
			expected: "sshd auth method password",
		},
		{
			name:     "publickey method",
			method:   "publickey",
			expected: "sshd auth method publickey",
		},
		{
			name:     "any method",
			method:   "any",
			expected: "no sshd auth method",
		},
		{
			name:     "empty method defaults to any",
			method:   "",
			expected: "no sshd auth method",
		},
		{
			name:     "unknown method defaults to any",
			method:   "unknown",
			expected: "no sshd auth method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildSSHDAuthMethodCommand(tt.method)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteSSHDAuthMethodCommand(t *testing.T) {
	result := BuildDeleteSSHDAuthMethodCommand()
	expected := "no sshd auth method"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestParseSSHDHostKeyInfo(t *testing.T) {
	// Test cases for "show sshd host key" output (OpenSSH public key format)
	// Using valid base64 strings that represent actual SSH key data format
	tests := []struct {
		name              string
		input             string
		expectAlgorithm   string
		expectFingerprint bool // true if we expect a non-empty fingerprint starting with "SHA256:"
	}{
		{
			name:              "empty output",
			input:             "",
			expectAlgorithm:   "",
			expectFingerprint: false,
		},
		{
			name:              "ssh-rsa key",
			input:             "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQC7",
			expectAlgorithm:   "ssh-rsa",
			expectFingerprint: true,
		},
		{
			name:              "ssh-dss key",
			input:             "ssh-dss AAAAB3NzaC1kc3MAAACBAM0=",
			expectAlgorithm:   "ssh-dss",
			expectFingerprint: true,
		},
		{
			name:              "ssh-ed25519 key",
			input:             "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIA==",
			expectAlgorithm:   "ssh-ed25519",
			expectFingerprint: true,
		},
		{
			name: "multiple keys prefers RSA",
			input: `ssh-dss AAAAB3NzaC1kc3MAAACBAM0=
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQC7`,
			expectAlgorithm:   "ssh-rsa",
			expectFingerprint: true,
		},
		{
			name: "RTX actual output format with multiple lines",
			input: `ssh-dss AAAAB3NzaC1kc3MAAACBAM0=
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQC7`,
			expectAlgorithm:   "ssh-rsa",
			expectFingerprint: true,
		},
		{
			name: "RTX wrapped key with line continuation",
			input: `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQDLGxJbUtgNvfAxr+kwHhfvGehvnm61nStY2DnXYzd
tQe0fdGmPUewxLhQO68z2e0morNpwsu96EFU0R6gFftr1/zvSOan82FrXom8RyudM0WyUX5GHMvcCSR
CZRSMw0nEqPXbOCaKr6596YJZxY6wXKzTghO6LwVW78jvhDTbs+Q==
[RTX1210] >`,
			expectAlgorithm:   "ssh-rsa",
			expectFingerprint: true,
		},
		{
			name: "RTX output with both DSS and RSA wrapped keys",
			input: `ssh-dss AAAAB3NzaC1kc3MAAACBAJnCmRWBNTPHkE8awFpxNEc8G7t9RNQAO9XDlQlCrK79qKZS3Yt
Wtn4iMi3R5ppyfjOj/G2jXimj3+pUg+nXjQ0BCIqHUvUZlZhE8aw4BB7/YnbJPxonrBe2PXgx7b7ynp
cDvDEvrH/I1NwWCaFyCswugPC/V6CZSStrYrnpQ+FvAAAAFQDTus3D+g==
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQDLGxJbUtgNvfAxr+kwHhfvGehvnm61nStY2DnXYzd
tQe0fdGmPUewxLhQO68z2e0morNpwsu96EFU0R6gFftr1/zvSOan82FrXom8RyudM0WyUX5GHMvcCSR
CZRSMw0nEqPXbOCaKr6596YJZxY6wXKzTghO6LwVW78jvhDTbs+Q==
[RTX1210] >`,
			expectAlgorithm:   "ssh-rsa",
			expectFingerprint: true,
		},
		{
			name:              "with leading whitespace",
			input:             "  ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQC7",
			expectAlgorithm:   "ssh-rsa",
			expectFingerprint: true,
		},
		{
			name: "no host key (other output)",
			input: `SSHD Status
Service: running
Connections: 0`,
			expectAlgorithm:   "",
			expectFingerprint: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseSSHDHostKeyInfo(tt.input)
			if result.Algorithm != tt.expectAlgorithm {
				t.Errorf("algorithm: got %q, want %q", result.Algorithm, tt.expectAlgorithm)
			}
			if tt.expectFingerprint {
				if result.Fingerprint == "" {
					t.Errorf("expected non-empty fingerprint, got empty")
				} else if !strings.HasPrefix(result.Fingerprint, "SHA256:") {
					t.Errorf("fingerprint should start with 'SHA256:', got %q", result.Fingerprint)
				}
			} else {
				if result.Fingerprint != "" {
					t.Errorf("expected empty fingerprint, got %q", result.Fingerprint)
				}
			}
		})
	}
}

func TestComputeSSHFingerprint(t *testing.T) {
	tests := []struct {
		name        string
		keyData     string
		expectEmpty bool
	}{
		{
			name:        "valid base64",
			keyData:     "AAAAB3NzaC1yc2EAAAADAQABAAAAgQDLGxJb",
			expectEmpty: false,
		},
		{
			name:        "invalid base64",
			keyData:     "!!!not-base64!!!",
			expectEmpty: true,
		},
		{
			name:        "empty input",
			keyData:     "",
			expectEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeSSHFingerprint(tt.keyData)
			if tt.expectEmpty {
				if result != "" {
					t.Errorf("expected empty fingerprint, got %q", result)
				}
			} else {
				if result == "" {
					t.Errorf("expected non-empty fingerprint")
				}
				if !strings.HasPrefix(result, "SHA256:") {
					t.Errorf("fingerprint should start with 'SHA256:', got %q", result)
				}
			}
		})
	}
}

func TestBuildShowSSHDAuthorizedKeysCommand(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected string
	}{
		{
			name:     "admin user",
			username: "admin",
			expected: "show sshd authorized-keys admin",
		},
		{
			name:     "root user",
			username: "root",
			expected: "show sshd authorized-keys root",
		},
		{
			name:     "user with numbers",
			username: "user123",
			expected: "show sshd authorized-keys user123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildShowSSHDAuthorizedKeysCommand(tt.username)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildImportSSHDAuthorizedKeysCommand(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected string
	}{
		{
			name:     "admin user",
			username: "admin",
			expected: "import sshd authorized-keys admin",
		},
		{
			name:     "root user",
			username: "root",
			expected: "import sshd authorized-keys root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildImportSSHDAuthorizedKeysCommand(tt.username)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteSSHDAuthorizedKeysCommand(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected string
	}{
		{
			name:     "admin user",
			username: "admin",
			expected: "delete /ssh/authorized_keys/admin",
		},
		{
			name:     "root user",
			username: "root",
			expected: "delete /ssh/authorized_keys/root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteSSHDAuthorizedKeysCommand(tt.username)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseSSHDAuthorizedKeys(t *testing.T) {
	// ParseSSHDAuthorizedKeys parses OpenSSH format public keys from RTX router
	// Format: <type> <base64-key> <comment>
	tests := []struct {
		name     string
		input    string
		expected []SSHAuthorizedKey
	}{
		{
			name:     "empty output",
			input:    "",
			expected: []SSHAuthorizedKey{},
		},
		{
			name:     "no keys message",
			input:    "No authorized keys found",
			expected: []SSHAuthorizedKey{},
		},
		{
			name:  "single ED25519 key",
			input: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBtest user@host",
			expected: []SSHAuthorizedKey{
				{
					Type:        "ssh-ed25519",
					Fingerprint: "AAAAC3NzaC1lZDI1NTE5AAAAIBtest",
					Comment:     "user@host",
				},
			},
		},
		{
			name:  "single RSA key",
			input: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQC7 admin@pc",
			expected: []SSHAuthorizedKey{
				{
					Type:        "ssh-rsa",
					Fingerprint: "AAAAB3NzaC1yc2EAAAADAQABAAAAgQC7",
					Comment:     "admin@pc",
				},
			},
		},
		{
			name: "multiple keys",
			input: `ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBtest user@host
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQC7 admin@pc
ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTY= test@server`,
			expected: []SSHAuthorizedKey{
				{
					Type:        "ssh-ed25519",
					Fingerprint: "AAAAC3NzaC1lZDI1NTE5AAAAIBtest",
					Comment:     "user@host",
				},
				{
					Type:        "ssh-rsa",
					Fingerprint: "AAAAB3NzaC1yc2EAAAADAQABAAAAgQC7",
					Comment:     "admin@pc",
				},
				{
					Type:        "ecdsa-sha2-nistp256",
					Fingerprint: "AAAAE2VjZHNhLXNoYTItbmlzdHAyNTY=",
					Comment:     "test@server",
				},
			},
		},
		{
			name: "with leading/trailing whitespace",
			input: `  ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBtest1 user1@host1
  ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQC7 user2@host2  `,
			expected: []SSHAuthorizedKey{
				{
					Type:        "ssh-ed25519",
					Fingerprint: "AAAAC3NzaC1lZDI1NTE5AAAAIBtest1",
					Comment:     "user1@host1",
				},
				{
					Type:        "ssh-rsa",
					Fingerprint: "AAAAB3NzaC1yc2EAAAADAQABAAAAgQC7",
					Comment:     "user2@host2",
				},
			},
		},
		{
			name: "with empty lines",
			input: `
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBtest user@host

`,
			expected: []SSHAuthorizedKey{
				{
					Type:        "ssh-ed25519",
					Fingerprint: "AAAAC3NzaC1lZDI1NTE5AAAAIBtest",
					Comment:     "user@host",
				},
			},
		},
		{
			name:  "key without comment",
			input: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBtest",
			expected: []SSHAuthorizedKey{
				{
					Type:        "ssh-ed25519",
					Fingerprint: "AAAAC3NzaC1lZDI1NTE5AAAAIBtest",
					Comment:     "",
				},
			},
		},
		{
			name: "RTX prompt lines ignored",
			input: `[RTX1210] # show sshd authorized-keys admin
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBtest user@host
[RTX1210] #`,
			expected: []SSHAuthorizedKey{
				{
					Type:        "ssh-ed25519",
					Fingerprint: "AAAAC3NzaC1lZDI1NTE5AAAAIBtest",
					Comment:     "user@host",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSSHDAuthorizedKeys(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Handle nil vs empty slice comparison
			if len(result) == 0 && len(tt.expected) == 0 {
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("got %+v, want %+v", result, tt.expected)
			}
		})
	}
}
