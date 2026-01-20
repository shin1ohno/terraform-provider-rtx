package parsers

import (
	"testing"
)

// ============================================================================
// Validation Tests
// ============================================================================

func TestValidateHostname(t *testing.T) {
	tests := []struct {
		name      string
		hostname  string
		expectErr bool
	}{
		{
			name:      "valid simple hostname",
			hostname:  "myhost",
			expectErr: false,
		},
		{
			name:      "valid FQDN",
			hostname:  "myhost.example.com",
			expectErr: false,
		},
		{
			name:      "valid NetVolante hostname",
			hostname:  "myhost.aa0.netvolante.jp",
			expectErr: false,
		},
		{
			name:      "valid hostname with hyphens",
			hostname:  "my-host-name.example.com",
			expectErr: false,
		},
		{
			name:      "valid hostname with numbers",
			hostname:  "host123.example.com",
			expectErr: false,
		},
		{
			name:      "empty hostname",
			hostname:  "",
			expectErr: true,
		},
		{
			name:      "hostname starting with hyphen",
			hostname:  "-hostname.example.com",
			expectErr: true,
		},
		{
			name:      "hostname ending with hyphen",
			hostname:  "hostname-.example.com",
			expectErr: true,
		},
		{
			name:      "hostname with underscore",
			hostname:  "host_name.example.com",
			expectErr: true,
		},
		{
			name:      "hostname with spaces",
			hostname:  "host name.example.com",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHostname(tt.hostname)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateDDNSURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		expectErr bool
	}{
		{
			name:      "valid HTTPS URL",
			url:       "https://dynupdate.no-ip.com/nic/update",
			expectErr: false,
		},
		{
			name:      "valid HTTP URL",
			url:       "http://www.dyndns.org/nic/update",
			expectErr: false,
		},
		{
			name:      "valid URL with path",
			url:       "https://api.example.com/ddns/update",
			expectErr: false,
		},
		{
			name:      "valid URL with query",
			url:       "https://api.example.com/update?hostname=",
			expectErr: false,
		},
		{
			name:      "empty URL",
			url:       "",
			expectErr: true,
		},
		{
			name:      "URL without protocol",
			url:       "example.com/update",
			expectErr: true,
		},
		{
			name:      "FTP URL (invalid)",
			url:       "ftp://example.com/update",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDDNSURL(tt.url)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// ============================================================================
// NetVolante DNS Parser Tests
// ============================================================================

func TestParseNetVolanteDNS_BasicConfiguration(t *testing.T) {
	raw := `
netvolante-dns hostname host pp 1 myhost.aa0.netvolante.jp
netvolante-dns use pp 1 on
netvolante-dns go pp 1
`
	parser := NewDDNSParser()
	configs, err := parser.ParseNetVolanteDNS(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	cfg := configs[0]
	if cfg.Hostname != "myhost.aa0.netvolante.jp" {
		t.Errorf("Hostname: expected 'myhost.aa0.netvolante.jp', got '%s'", cfg.Hostname)
	}
	if cfg.Interface != "pp 1" {
		t.Errorf("Interface: expected 'pp 1', got '%s'", cfg.Interface)
	}
	if !cfg.Use {
		t.Error("Use: expected true, got false")
	}
}

func TestParseNetVolanteDNS_ServerSelection(t *testing.T) {
	raw := `
netvolante-dns server 2
netvolante-dns hostname host pp 1 myhost.aa0.netvolante.jp
`
	parser := NewDDNSParser()
	configs, err := parser.ParseNetVolanteDNS(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	if configs[0].Server != 2 {
		t.Errorf("Server: expected 2, got %d", configs[0].Server)
	}
}

func TestParseNetVolanteDNS_Timeout(t *testing.T) {
	raw := `
netvolante-dns timeout 120
netvolante-dns hostname host pp 1 myhost.aa0.netvolante.jp
`
	parser := NewDDNSParser()
	configs, err := parser.ParseNetVolanteDNS(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	if configs[0].Timeout != 120 {
		t.Errorf("Timeout: expected 120, got %d", configs[0].Timeout)
	}
}

func TestParseNetVolanteDNS_IPv6(t *testing.T) {
	raw := `
netvolante-dns hostname host pp 1 myhost.aa0.netvolante.jp
netvolante-dns use ipv6 pp 1 on
`
	parser := NewDDNSParser()
	configs, err := parser.ParseNetVolanteDNS(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	if !configs[0].IPv6 {
		t.Error("IPv6: expected true, got false")
	}
}

func TestParseNetVolanteDNS_AutoHostname(t *testing.T) {
	raw := `
netvolante-dns auto hostname pp 1 on
`
	parser := NewDDNSParser()
	configs, err := parser.ParseNetVolanteDNS(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	if !configs[0].AutoHostname {
		t.Error("AutoHostname: expected true, got false")
	}
}

func TestParseNetVolanteDNS_MultipleInterfaces(t *testing.T) {
	raw := `
netvolante-dns hostname host pp 1 host1.aa0.netvolante.jp
netvolante-dns hostname host pp 2 host2.aa0.netvolante.jp
netvolante-dns use pp 1 on
netvolante-dns use pp 2 on
`
	parser := NewDDNSParser()
	configs, err := parser.ParseNetVolanteDNS(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 2 {
		t.Fatalf("Expected 2 configs, got %d", len(configs))
	}
}

func TestParseNetVolanteDNS_InterfaceNormalization(t *testing.T) {
	raw := `
netvolante-dns hostname host pp1 myhost.aa0.netvolante.jp
`
	parser := NewDDNSParser()
	configs, err := parser.ParseNetVolanteDNS(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	// "pp1" should be normalized to "pp 1"
	if configs[0].Interface != "pp 1" {
		t.Errorf("Interface: expected 'pp 1', got '%s'", configs[0].Interface)
	}
}

func TestParseNetVolanteDNS_LAN1Interface(t *testing.T) {
	raw := `
netvolante-dns hostname host lan1 myhost.aa0.netvolante.jp
`
	parser := NewDDNSParser()
	configs, err := parser.ParseNetVolanteDNS(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	// "lan1" should remain as is
	if configs[0].Interface != "lan1" {
		t.Errorf("Interface: expected 'lan1', got '%s'", configs[0].Interface)
	}
}

func TestParseNetVolanteDNS_FullConfiguration(t *testing.T) {
	raw := `
netvolante-dns server 1
netvolante-dns timeout 90
netvolante-dns hostname host pp 1 myhost.aa0.netvolante.jp
netvolante-dns auto hostname pp 1 off
netvolante-dns use ipv6 pp 1 on
netvolante-dns use pp 1 on
netvolante-dns go pp 1
`
	parser := NewDDNSParser()
	configs, err := parser.ParseNetVolanteDNS(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	cfg := configs[0]
	if cfg.Hostname != "myhost.aa0.netvolante.jp" {
		t.Errorf("Hostname: expected 'myhost.aa0.netvolante.jp', got '%s'", cfg.Hostname)
	}
	if cfg.Interface != "pp 1" {
		t.Errorf("Interface: expected 'pp 1', got '%s'", cfg.Interface)
	}
	if cfg.Server != 1 {
		t.Errorf("Server: expected 1, got %d", cfg.Server)
	}
	if cfg.Timeout != 90 {
		t.Errorf("Timeout: expected 90, got %d", cfg.Timeout)
	}
	if !cfg.IPv6 {
		t.Error("IPv6: expected true, got false")
	}
	if cfg.AutoHostname {
		t.Error("AutoHostname: expected false, got true")
	}
	if !cfg.Use {
		t.Error("Use: expected true, got false")
	}
}

func TestParseNetVolanteDNS_Empty(t *testing.T) {
	parser := NewDDNSParser()
	configs, err := parser.ParseNetVolanteDNS("")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 0 {
		t.Errorf("Expected 0 configs, got %d", len(configs))
	}
}

func TestParseNetVolanteDNS_CommentsIgnored(t *testing.T) {
	raw := `
# NetVolante DNS configuration
netvolante-dns hostname host pp 1 myhost.aa0.netvolante.jp
# This is a comment
netvolante-dns use pp 1 on
`
	parser := NewDDNSParser()
	configs, err := parser.ParseNetVolanteDNS(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}
}

// ============================================================================
// Custom DDNS Parser Tests
// ============================================================================

func TestParseDDNSConfig_BasicConfiguration(t *testing.T) {
	raw := `
ddns server url 1 https://dynupdate.no-ip.com/nic/update
ddns server hostname 1 myhost.no-ip.org
ddns server user 1 myuser mypassword
ddns server go 1
`
	parser := NewDDNSParser()
	configs, err := parser.ParseDDNSConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	cfg := configs[0]
	if cfg.ID != 1 {
		t.Errorf("ID: expected 1, got %d", cfg.ID)
	}
	if cfg.URL != "https://dynupdate.no-ip.com/nic/update" {
		t.Errorf("URL: expected 'https://dynupdate.no-ip.com/nic/update', got '%s'", cfg.URL)
	}
	if cfg.Hostname != "myhost.no-ip.org" {
		t.Errorf("Hostname: expected 'myhost.no-ip.org', got '%s'", cfg.Hostname)
	}
	if cfg.Username != "myuser" {
		t.Errorf("Username: expected 'myuser', got '%s'", cfg.Username)
	}
	if cfg.Password != "mypassword" {
		t.Errorf("Password: expected 'mypassword', got '%s'", cfg.Password)
	}
}

func TestParseDDNSConfig_MultipleServers(t *testing.T) {
	raw := `
ddns server url 1 https://dynupdate.no-ip.com/nic/update
ddns server hostname 1 host1.no-ip.org
ddns server user 1 user1 pass1

ddns server url 2 https://www.dyndns.org/nic/update
ddns server hostname 2 host2.dyndns.org
ddns server user 2 user2 pass2
`
	parser := NewDDNSParser()
	configs, err := parser.ParseDDNSConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 2 {
		t.Fatalf("Expected 2 configs, got %d", len(configs))
	}

	// Configs should be sorted by ID
	if configs[0].ID != 1 {
		t.Errorf("First config ID: expected 1, got %d", configs[0].ID)
	}
	if configs[1].ID != 2 {
		t.Errorf("Second config ID: expected 2, got %d", configs[1].ID)
	}

	if configs[0].Hostname != "host1.no-ip.org" {
		t.Errorf("First hostname: expected 'host1.no-ip.org', got '%s'", configs[0].Hostname)
	}
	if configs[1].Hostname != "host2.dyndns.org" {
		t.Errorf("Second hostname: expected 'host2.dyndns.org', got '%s'", configs[1].Hostname)
	}
}

func TestParseDDNSConfig_PasswordWithSpecialChars(t *testing.T) {
	raw := `
ddns server url 1 https://example.com/update
ddns server user 1 myuser p@ssw0rd!#$%
`
	parser := NewDDNSParser()
	configs, err := parser.ParseDDNSConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	if configs[0].Password != "p@ssw0rd!#$%" {
		t.Errorf("Password: expected 'p@ssw0rd!#$%%', got '%s'", configs[0].Password)
	}
}

func TestParseDDNSConfig_URLOnly(t *testing.T) {
	raw := `
ddns server url 1 https://example.com/update
`
	parser := NewDDNSParser()
	configs, err := parser.ParseDDNSConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	if configs[0].URL != "https://example.com/update" {
		t.Errorf("URL: expected 'https://example.com/update', got '%s'", configs[0].URL)
	}
	if configs[0].Hostname != "" {
		t.Errorf("Hostname: expected empty, got '%s'", configs[0].Hostname)
	}
}

func TestParseDDNSConfig_Empty(t *testing.T) {
	parser := NewDDNSParser()
	configs, err := parser.ParseDDNSConfig("")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 0 {
		t.Errorf("Expected 0 configs, got %d", len(configs))
	}
}

func TestParseDDNSConfig_GoCommandOnly(t *testing.T) {
	raw := `
ddns server go 1
`
	parser := NewDDNSParser()
	configs, err := parser.ParseDDNSConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Go command alone should create a minimal config entry
	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	if configs[0].ID != 1 {
		t.Errorf("ID: expected 1, got %d", configs[0].ID)
	}
}

// ============================================================================
// DDNS Status Parser Tests
// ============================================================================

func TestParseDDNSStatus_NetVolante(t *testing.T) {
	raw := `
Interface: pp 1
Hostname: myhost.aa0.netvolante.jp
IP Address: 203.0.113.1
Status: registered
Last Update: 2024-01-20 10:30:00
`
	parser := NewDDNSParser()
	statuses, err := parser.ParseDDNSStatus(raw, "netvolante")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(statuses) != 1 {
		t.Fatalf("Expected 1 status, got %d", len(statuses))
	}

	status := statuses[0]
	if status.Type != "netvolante" {
		t.Errorf("Type: expected 'netvolante', got '%s'", status.Type)
	}
	if status.Interface != "pp 1" {
		t.Errorf("Interface: expected 'pp 1', got '%s'", status.Interface)
	}
	if status.Hostname != "myhost.aa0.netvolante.jp" {
		t.Errorf("Hostname: expected 'myhost.aa0.netvolante.jp', got '%s'", status.Hostname)
	}
	if status.CurrentIP != "203.0.113.1" {
		t.Errorf("CurrentIP: expected '203.0.113.1', got '%s'", status.CurrentIP)
	}
	if status.Status != "registered" {
		t.Errorf("Status: expected 'registered', got '%s'", status.Status)
	}
	if status.LastUpdate != "2024-01-20 10:30:00" {
		t.Errorf("LastUpdate: expected '2024-01-20 10:30:00', got '%s'", status.LastUpdate)
	}
}

func TestParseDDNSStatus_NetVolanteWithError(t *testing.T) {
	raw := `
Interface: pp 1
Hostname: myhost.aa0.netvolante.jp
Status: error
Error: Connection timeout
`
	parser := NewDDNSParser()
	statuses, err := parser.ParseDDNSStatus(raw, "netvolante")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(statuses) != 1 {
		t.Fatalf("Expected 1 status, got %d", len(statuses))
	}

	status := statuses[0]
	if status.Status != "error" {
		t.Errorf("Status: expected 'error', got '%s'", status.Status)
	}
	if status.ErrorMessage != "Connection timeout" {
		t.Errorf("ErrorMessage: expected 'Connection timeout', got '%s'", status.ErrorMessage)
	}
}

func TestParseDDNSStatus_Custom(t *testing.T) {
	raw := `
Server 1:
  Hostname: myhost.no-ip.org
  IP Address: 203.0.113.1
  Status: ok
  Last Update: 2024-01-20 10:30:00
`
	parser := NewDDNSParser()
	statuses, err := parser.ParseDDNSStatus(raw, "custom")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(statuses) != 1 {
		t.Fatalf("Expected 1 status, got %d", len(statuses))
	}

	status := statuses[0]
	if status.Type != "custom" {
		t.Errorf("Type: expected 'custom', got '%s'", status.Type)
	}
	if status.ServerID != 1 {
		t.Errorf("ServerID: expected 1, got %d", status.ServerID)
	}
	if status.Hostname != "myhost.no-ip.org" {
		t.Errorf("Hostname: expected 'myhost.no-ip.org', got '%s'", status.Hostname)
	}
	if status.CurrentIP != "203.0.113.1" {
		t.Errorf("CurrentIP: expected '203.0.113.1', got '%s'", status.CurrentIP)
	}
	if status.Status != "ok" {
		t.Errorf("Status: expected 'ok', got '%s'", status.Status)
	}
}

func TestParseDDNSStatus_AutoDetectNetVolante(t *testing.T) {
	raw := `
NetVolante DNS Status
Interface: pp 1
Hostname: myhost.aa0.netvolante.jp
`
	parser := NewDDNSParser()
	statuses, err := parser.ParseDDNSStatus(raw, "")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(statuses) != 1 {
		t.Fatalf("Expected 1 status, got %d", len(statuses))
	}

	if statuses[0].Type != "netvolante" {
		t.Errorf("Type: expected 'netvolante', got '%s'", statuses[0].Type)
	}
}

func TestParseDDNSStatus_Empty(t *testing.T) {
	parser := NewDDNSParser()
	statuses, err := parser.ParseDDNSStatus("", "")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(statuses) != 0 {
		t.Errorf("Expected 0 statuses, got %d", len(statuses))
	}
}

func TestParseDDNSStatus_MultipleNetVolante(t *testing.T) {
	raw := `
Interface: pp 1
Hostname: host1.aa0.netvolante.jp
IP Address: 203.0.113.1
Status: registered

Interface: pp 2
Hostname: host2.aa0.netvolante.jp
IP Address: 203.0.113.2
Status: registered
`
	parser := NewDDNSParser()
	statuses, err := parser.ParseDDNSStatus(raw, "netvolante")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(statuses) != 2 {
		t.Fatalf("Expected 2 statuses, got %d", len(statuses))
	}
}

// ============================================================================
// NetVolante Command Builder Tests
// ============================================================================

func TestBuildNetVolanteHostnameCommand(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		hostname string
		expected string
	}{
		{
			name:     "valid command",
			iface:    "pp 1",
			hostname: "myhost.aa0.netvolante.jp",
			expected: "netvolante-dns hostname host pp 1 myhost.aa0.netvolante.jp",
		},
		{
			name:     "empty interface",
			iface:    "",
			hostname: "myhost.aa0.netvolante.jp",
			expected: "",
		},
		{
			name:     "empty hostname",
			iface:    "pp 1",
			hostname: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNetVolanteHostnameCommand(tt.iface, tt.hostname)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildNetVolanteServerCommand(t *testing.T) {
	tests := []struct {
		name     string
		server   int
		expected string
	}{
		{
			name:     "server 1",
			server:   1,
			expected: "netvolante-dns server 1",
		},
		{
			name:     "server 2",
			server:   2,
			expected: "netvolante-dns server 2",
		},
		{
			name:     "invalid server 0",
			server:   0,
			expected: "",
		},
		{
			name:     "invalid server 3",
			server:   3,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNetVolanteServerCommand(tt.server)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildNetVolanteGoCommand(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		expected string
	}{
		{
			name:     "pp 1",
			iface:    "pp 1",
			expected: "netvolante-dns go pp 1",
		},
		{
			name:     "lan1",
			iface:    "lan1",
			expected: "netvolante-dns go lan1",
		},
		{
			name:     "empty",
			iface:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNetVolanteGoCommand(tt.iface)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildNetVolanteTimeoutCommand(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int
		expected string
	}{
		{
			name:     "90 seconds",
			seconds:  90,
			expected: "netvolante-dns timeout 90",
		},
		{
			name:     "120 seconds",
			seconds:  120,
			expected: "netvolante-dns timeout 120",
		},
		{
			name:     "zero",
			seconds:  0,
			expected: "",
		},
		{
			name:     "negative",
			seconds:  -1,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNetVolanteTimeoutCommand(tt.seconds)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildNetVolanteIPv6Command(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		enable   bool
		expected string
	}{
		{
			name:     "enable IPv6",
			iface:    "pp 1",
			enable:   true,
			expected: "netvolante-dns use ipv6 pp 1 on",
		},
		{
			name:     "disable IPv6",
			iface:    "pp 1",
			enable:   false,
			expected: "netvolante-dns use ipv6 pp 1 off",
		},
		{
			name:     "empty interface",
			iface:    "",
			enable:   true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNetVolanteIPv6Command(tt.iface, tt.enable)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildNetVolanteAutoHostnameCommand(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		enable   bool
		expected string
	}{
		{
			name:     "enable auto hostname",
			iface:    "pp 1",
			enable:   true,
			expected: "netvolante-dns auto hostname pp 1 on",
		},
		{
			name:     "disable auto hostname",
			iface:    "pp 1",
			enable:   false,
			expected: "netvolante-dns auto hostname pp 1 off",
		},
		{
			name:     "empty interface",
			iface:    "",
			enable:   true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNetVolanteAutoHostnameCommand(tt.iface, tt.enable)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildNetVolanteUseCommand(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		enable   bool
		expected string
	}{
		{
			name:     "enable use",
			iface:    "pp 1",
			enable:   true,
			expected: "netvolante-dns use pp 1 on",
		},
		{
			name:     "disable use",
			iface:    "pp 1",
			enable:   false,
			expected: "netvolante-dns use pp 1 off",
		},
		{
			name:     "empty interface",
			iface:    "",
			enable:   true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNetVolanteUseCommand(tt.iface, tt.enable)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildDeleteNetVolanteHostnameCommand(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		expected string
	}{
		{
			name:     "valid command",
			iface:    "pp 1",
			expected: "no netvolante-dns hostname host pp 1",
		},
		{
			name:     "empty interface",
			iface:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeleteNetVolanteHostnameCommand(tt.iface)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildNetVolanteCommand(t *testing.T) {
	config := NetVolanteConfig{
		Interface:    "pp 1",
		Hostname:     "myhost.aa0.netvolante.jp",
		Server:       2,
		Timeout:      120,
		IPv6:         true,
		AutoHostname: true,
		Use:          true,
	}

	commands := BuildNetVolanteCommand(config)

	// Should contain server, timeout, hostname, auto hostname, IPv6, use, and go commands
	if len(commands) < 5 {
		t.Errorf("Expected at least 5 commands, got %d: %v", len(commands), commands)
	}

	// Check for specific commands
	expectedCommands := []string{
		"netvolante-dns server 2",
		"netvolante-dns timeout 120",
		"netvolante-dns hostname host pp 1 myhost.aa0.netvolante.jp",
		"netvolante-dns auto hostname pp 1 on",
		"netvolante-dns use ipv6 pp 1 on",
		"netvolante-dns use pp 1 on",
		"netvolante-dns go pp 1",
	}

	for _, expected := range expectedCommands {
		found := false
		for _, cmd := range commands {
			if cmd == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command '%s' not found in %v", expected, commands)
		}
	}
}

// ============================================================================
// Custom DDNS Command Builder Tests
// ============================================================================

func TestBuildDDNSURLCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		url      string
		expected string
	}{
		{
			name:     "valid command",
			id:       1,
			url:      "https://dynupdate.no-ip.com/nic/update",
			expected: "ddns server url 1 https://dynupdate.no-ip.com/nic/update",
		},
		{
			name:     "invalid id 0",
			id:       0,
			url:      "https://example.com",
			expected: "",
		},
		{
			name:     "invalid id 5",
			id:       5,
			url:      "https://example.com",
			expected: "",
		},
		{
			name:     "empty url",
			id:       1,
			url:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDDNSURLCommand(tt.id, tt.url)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildDDNSHostnameCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		hostname string
		expected string
	}{
		{
			name:     "valid command",
			id:       1,
			hostname: "myhost.no-ip.org",
			expected: "ddns server hostname 1 myhost.no-ip.org",
		},
		{
			name:     "invalid id",
			id:       0,
			hostname: "myhost.no-ip.org",
			expected: "",
		},
		{
			name:     "empty hostname",
			id:       1,
			hostname: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDDNSHostnameCommand(tt.id, tt.hostname)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildDDNSUserCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		username string
		password string
		expected string
	}{
		{
			name:     "valid command",
			id:       1,
			username: "myuser",
			password: "mypassword",
			expected: "ddns server user 1 myuser mypassword",
		},
		{
			name:     "invalid id",
			id:       0,
			username: "myuser",
			password: "mypassword",
			expected: "",
		},
		{
			name:     "empty username",
			id:       1,
			username: "",
			password: "mypassword",
			expected: "",
		},
		{
			name:     "empty password",
			id:       1,
			username: "myuser",
			password: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDDNSUserCommand(tt.id, tt.username, tt.password)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildDDNSGoCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		expected string
	}{
		{
			name:     "valid id 1",
			id:       1,
			expected: "ddns server go 1",
		},
		{
			name:     "valid id 4",
			id:       4,
			expected: "ddns server go 4",
		},
		{
			name:     "invalid id 0",
			id:       0,
			expected: "",
		},
		{
			name:     "invalid id 5",
			id:       5,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDDNSGoCommand(tt.id)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildDeleteDDNSCommands(t *testing.T) {
	urlCmd := BuildDeleteDDNSURLCommand(1)
	if urlCmd != "no ddns server url 1" {
		t.Errorf("Expected 'no ddns server url 1', got '%s'", urlCmd)
	}

	hostnameCmd := BuildDeleteDDNSHostnameCommand(2)
	if hostnameCmd != "no ddns server hostname 2" {
		t.Errorf("Expected 'no ddns server hostname 2', got '%s'", hostnameCmd)
	}

	userCmd := BuildDeleteDDNSUserCommand(3)
	if userCmd != "no ddns server user 3" {
		t.Errorf("Expected 'no ddns server user 3', got '%s'", userCmd)
	}
}

func TestBuildDDNSCommand(t *testing.T) {
	config := DDNSServerConfig{
		ID:       1,
		URL:      "https://dynupdate.no-ip.com/nic/update",
		Hostname: "myhost.no-ip.org",
		Username: "myuser",
		Password: "mypassword",
	}

	commands := BuildDDNSCommand(config)

	// Should contain url, hostname, user, and go commands
	if len(commands) != 4 {
		t.Errorf("Expected 4 commands, got %d: %v", len(commands), commands)
	}

	expectedCommands := []string{
		"ddns server url 1 https://dynupdate.no-ip.com/nic/update",
		"ddns server hostname 1 myhost.no-ip.org",
		"ddns server user 1 myuser mypassword",
		"ddns server go 1",
	}

	for i, expected := range expectedCommands {
		if i < len(commands) && commands[i] != expected {
			t.Errorf("Command %d: expected '%s', got '%s'", i, expected, commands[i])
		}
	}
}

func TestBuildDeleteDDNSCommand(t *testing.T) {
	commands := BuildDeleteDDNSCommand(1)

	if len(commands) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(commands))
	}

	expectedCommands := []string{
		"no ddns server user 1",
		"no ddns server hostname 1",
		"no ddns server url 1",
	}

	for i, expected := range expectedCommands {
		if i < len(commands) && commands[i] != expected {
			t.Errorf("Command %d: expected '%s', got '%s'", i, expected, commands[i])
		}
	}
}

func TestBuildDeleteDDNSCommand_InvalidID(t *testing.T) {
	commands := BuildDeleteDDNSCommand(0)
	if commands != nil {
		t.Errorf("Expected nil for invalid ID, got %v", commands)
	}

	commands = BuildDeleteDDNSCommand(5)
	if commands != nil {
		t.Errorf("Expected nil for invalid ID, got %v", commands)
	}
}

// ============================================================================
// Show Status Command Tests
// ============================================================================

func TestBuildShowStatusCommands(t *testing.T) {
	nvCmd := BuildShowNetVolanteStatusCommand()
	if nvCmd != "show status netvolante-dns" {
		t.Errorf("Expected 'show status netvolante-dns', got '%s'", nvCmd)
	}

	ddnsCmd := BuildShowDDNSStatusCommand()
	if ddnsCmd != "show status ddns" {
		t.Errorf("Expected 'show status ddns', got '%s'", ddnsCmd)
	}
}

// ============================================================================
// Validation Tests
// ============================================================================

func TestValidateNetVolanteConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    NetVolanteConfig
		expectErr bool
	}{
		{
			name: "valid config",
			config: NetVolanteConfig{
				Interface: "pp 1",
				Hostname:  "myhost.aa0.netvolante.jp",
				Server:    1,
				Timeout:   90,
			},
			expectErr: false,
		},
		{
			name: "empty interface",
			config: NetVolanteConfig{
				Interface: "",
				Hostname:  "myhost.aa0.netvolante.jp",
			},
			expectErr: true,
		},
		{
			name: "invalid server",
			config: NetVolanteConfig{
				Interface: "pp 1",
				Server:    3,
			},
			expectErr: true,
		},
		{
			name: "invalid timeout",
			config: NetVolanteConfig{
				Interface: "pp 1",
				Timeout:   -1,
			},
			expectErr: true,
		},
		{
			name: "invalid hostname",
			config: NetVolanteConfig{
				Interface: "pp 1",
				Hostname:  "-invalid.hostname",
			},
			expectErr: true,
		},
		{
			name: "config without hostname (auto hostname)",
			config: NetVolanteConfig{
				Interface:    "pp 1",
				AutoHostname: true,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNetVolanteConfig(tt.config)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateDDNSServerConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    DDNSServerConfig
		expectErr bool
	}{
		{
			name: "valid config",
			config: DDNSServerConfig{
				ID:       1,
				URL:      "https://dynupdate.no-ip.com/nic/update",
				Hostname: "myhost.no-ip.org",
				Username: "myuser",
				Password: "mypassword",
			},
			expectErr: false,
		},
		{
			name: "invalid ID 0",
			config: DDNSServerConfig{
				ID: 0,
			},
			expectErr: true,
		},
		{
			name: "invalid ID 5",
			config: DDNSServerConfig{
				ID: 5,
			},
			expectErr: true,
		},
		{
			name: "invalid URL",
			config: DDNSServerConfig{
				ID:  1,
				URL: "ftp://invalid.com",
			},
			expectErr: true,
		},
		{
			name: "invalid hostname",
			config: DDNSServerConfig{
				ID:       1,
				Hostname: "-invalid.hostname",
			},
			expectErr: true,
		},
		{
			name: "username without password",
			config: DDNSServerConfig{
				ID:       1,
				Username: "myuser",
				Password: "",
			},
			expectErr: true,
		},
		{
			name: "password without username is ok",
			config: DDNSServerConfig{
				ID:       1,
				Password: "mypassword",
			},
			expectErr: false,
		},
		{
			name: "minimal valid config",
			config: DDNSServerConfig{
				ID: 1,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDDNSServerConfig(tt.config)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// ============================================================================
// Round-Trip Tests
// ============================================================================

func TestNetVolanteRoundTrip(t *testing.T) {
	parser := NewDDNSParser()

	original := NetVolanteConfig{
		Interface:    "pp 1",
		Hostname:     "myhost.aa0.netvolante.jp",
		Server:       2,
		Timeout:      120,
		IPv6:         true,
		AutoHostname: false,
		Use:          true,
	}

	// Build commands
	commands := BuildNetVolanteCommand(original)

	// Parse the commands back
	raw := ""
	for _, cmd := range commands {
		raw += cmd + "\n"
	}

	configs, err := parser.ParseNetVolanteDNS(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	parsed := configs[0]

	if parsed.Interface != original.Interface {
		t.Errorf("Interface: expected '%s', got '%s'", original.Interface, parsed.Interface)
	}
	if parsed.Hostname != original.Hostname {
		t.Errorf("Hostname: expected '%s', got '%s'", original.Hostname, parsed.Hostname)
	}
	if parsed.Server != original.Server {
		t.Errorf("Server: expected %d, got %d", original.Server, parsed.Server)
	}
	if parsed.Timeout != original.Timeout {
		t.Errorf("Timeout: expected %d, got %d", original.Timeout, parsed.Timeout)
	}
	if parsed.IPv6 != original.IPv6 {
		t.Errorf("IPv6: expected %v, got %v", original.IPv6, parsed.IPv6)
	}
	if parsed.Use != original.Use {
		t.Errorf("Use: expected %v, got %v", original.Use, parsed.Use)
	}
}

func TestCustomDDNSRoundTrip(t *testing.T) {
	parser := NewDDNSParser()

	original := DDNSServerConfig{
		ID:       1,
		URL:      "https://dynupdate.no-ip.com/nic/update",
		Hostname: "myhost.no-ip.org",
		Username: "myuser",
		Password: "mypassword",
	}

	// Build commands
	commands := BuildDDNSCommand(original)

	// Parse the commands back
	raw := ""
	for _, cmd := range commands {
		raw += cmd + "\n"
	}

	configs, err := parser.ParseDDNSConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	parsed := configs[0]

	if parsed.ID != original.ID {
		t.Errorf("ID: expected %d, got %d", original.ID, parsed.ID)
	}
	if parsed.URL != original.URL {
		t.Errorf("URL: expected '%s', got '%s'", original.URL, parsed.URL)
	}
	if parsed.Hostname != original.Hostname {
		t.Errorf("Hostname: expected '%s', got '%s'", original.Hostname, parsed.Hostname)
	}
	if parsed.Username != original.Username {
		t.Errorf("Username: expected '%s', got '%s'", original.Username, parsed.Username)
	}
	if parsed.Password != original.Password {
		t.Errorf("Password: expected '%s', got '%s'", original.Password, parsed.Password)
	}
}
