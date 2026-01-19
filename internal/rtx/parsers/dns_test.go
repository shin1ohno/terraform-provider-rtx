package parsers

import (
	"testing"
)

func TestParseDNSConfig_BasicConfiguration(t *testing.T) {
	raw := `
dns server 8.8.8.8 8.8.4.4
dns domain example.com
dns service on
dns private address spoof on
`
	parser := NewDNSParser()
	config, err := parser.ParseDNSConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse DNS config: %v", err)
	}

	// Check name servers
	if len(config.NameServers) != 2 {
		t.Errorf("Expected 2 name servers, got %d", len(config.NameServers))
	}
	if config.NameServers[0] != "8.8.8.8" {
		t.Errorf("Expected first server '8.8.8.8', got '%s'", config.NameServers[0])
	}
	if config.NameServers[1] != "8.8.4.4" {
		t.Errorf("Expected second server '8.8.4.4', got '%s'", config.NameServers[1])
	}

	// Check domain name
	if config.DomainName != "example.com" {
		t.Errorf("Expected domain name 'example.com', got '%s'", config.DomainName)
	}

	// Check service status
	if !config.ServiceOn {
		t.Error("Expected dns service to be on")
	}

	// Check private spoof
	if !config.PrivateSpoof {
		t.Error("Expected dns private address spoof to be on")
	}
}

func TestParseDNSConfig_ServerSelect(t *testing.T) {
	raw := `
dns server 8.8.8.8
dns server select 1 192.168.1.1 example.com
dns server select 2 10.0.0.1 10.0.0.2 *.local internal.net
`
	parser := NewDNSParser()
	config, err := parser.ParseDNSConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse DNS config: %v", err)
	}

	if len(config.ServerSelect) != 2 {
		t.Fatalf("Expected 2 server select entries, got %d", len(config.ServerSelect))
	}

	// Check first server select
	sel1 := config.ServerSelect[0]
	if sel1.ID != 1 {
		t.Errorf("Expected server select ID 1, got %d", sel1.ID)
	}
	if len(sel1.Servers) != 1 || sel1.Servers[0] != "192.168.1.1" {
		t.Errorf("Expected server ['192.168.1.1'], got %v", sel1.Servers)
	}
	if len(sel1.Domains) != 1 || sel1.Domains[0] != "example.com" {
		t.Errorf("Expected domains ['example.com'], got %v", sel1.Domains)
	}

	// Check second server select with multiple servers and domains
	sel2 := config.ServerSelect[1]
	if sel2.ID != 2 {
		t.Errorf("Expected server select ID 2, got %d", sel2.ID)
	}
	if len(sel2.Servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(sel2.Servers))
	}
	if len(sel2.Domains) != 2 {
		t.Errorf("Expected 2 domains, got %d", len(sel2.Domains))
	}
}

func TestParseDNSConfig_StaticHosts(t *testing.T) {
	raw := `
dns static myhost.local 192.168.1.100
dns static server1.example.com 10.0.0.1
dns static router.home 192.168.1.1
`
	parser := NewDNSParser()
	config, err := parser.ParseDNSConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse DNS config: %v", err)
	}

	if len(config.Hosts) != 3 {
		t.Fatalf("Expected 3 static hosts, got %d", len(config.Hosts))
	}

	// Check first host
	if config.Hosts[0].Name != "myhost.local" {
		t.Errorf("Expected hostname 'myhost.local', got '%s'", config.Hosts[0].Name)
	}
	if config.Hosts[0].Address != "192.168.1.100" {
		t.Errorf("Expected address '192.168.1.100', got '%s'", config.Hosts[0].Address)
	}
}

func TestParseDNSConfig_DomainLookup(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "domain lookup on",
			input:    "dns domain lookup on",
			expected: true,
		},
		{
			name:     "domain lookup off",
			input:    "dns domain lookup off",
			expected: false,
		},
		{
			name:     "no domain lookup",
			input:    "no dns domain lookup",
			expected: false,
		},
		{
			name:     "default (no config)",
			input:    "",
			expected: true,
		},
	}

	parser := NewDNSParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parser.ParseDNSConfig(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			if config.DomainLookup != tt.expected {
				t.Errorf("Expected DomainLookup=%v, got %v", tt.expected, config.DomainLookup)
			}
		})
	}
}

func TestParseDNSConfig_ServiceAndSpoof(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		serviceOn     bool
		privateSpoof  bool
	}{
		{
			name:         "both on",
			input:        "dns service on\ndns private address spoof on",
			serviceOn:    true,
			privateSpoof: true,
		},
		{
			name:         "both off",
			input:        "dns service off\ndns private address spoof off",
			serviceOn:    false,
			privateSpoof: false,
		},
		{
			name:         "service on spoof off",
			input:        "dns service on\ndns private address spoof off",
			serviceOn:    true,
			privateSpoof: false,
		},
		{
			name:         "default (no config)",
			input:        "",
			serviceOn:    false,
			privateSpoof: false,
		},
	}

	parser := NewDNSParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parser.ParseDNSConfig(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}
			if config.ServiceOn != tt.serviceOn {
				t.Errorf("Expected ServiceOn=%v, got %v", tt.serviceOn, config.ServiceOn)
			}
			if config.PrivateSpoof != tt.privateSpoof {
				t.Errorf("Expected PrivateSpoof=%v, got %v", tt.privateSpoof, config.PrivateSpoof)
			}
		})
	}
}

func TestParseDNSConfig_FullConfiguration(t *testing.T) {
	raw := `
dns domain lookup on
dns domain example.com
dns server 8.8.8.8 1.1.1.1
dns server select 1 192.168.1.1 internal.example.com
dns server select 2 10.0.0.1 *.local
dns static router 192.168.1.1
dns static nas 192.168.1.10
dns service on
dns private address spoof on
`
	parser := NewDNSParser()
	config, err := parser.ParseDNSConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse DNS config: %v", err)
	}

	if !config.DomainLookup {
		t.Error("Expected DomainLookup to be true")
	}
	if config.DomainName != "example.com" {
		t.Errorf("Expected domain 'example.com', got '%s'", config.DomainName)
	}
	if len(config.NameServers) != 2 {
		t.Errorf("Expected 2 name servers, got %d", len(config.NameServers))
	}
	if len(config.ServerSelect) != 2 {
		t.Errorf("Expected 2 server select entries, got %d", len(config.ServerSelect))
	}
	if len(config.Hosts) != 2 {
		t.Errorf("Expected 2 static hosts, got %d", len(config.Hosts))
	}
	if !config.ServiceOn {
		t.Error("Expected ServiceOn to be true")
	}
	if !config.PrivateSpoof {
		t.Error("Expected PrivateSpoof to be true")
	}
}

func TestBuildDNSServerCommand(t *testing.T) {
	tests := []struct {
		name     string
		servers  []string
		expected string
	}{
		{
			name:     "single server",
			servers:  []string{"8.8.8.8"},
			expected: "dns server 8.8.8.8",
		},
		{
			name:     "two servers",
			servers:  []string{"8.8.8.8", "8.8.4.4"},
			expected: "dns server 8.8.8.8 8.8.4.4",
		},
		{
			name:     "three servers",
			servers:  []string{"8.8.8.8", "8.8.4.4", "1.1.1.1"},
			expected: "dns server 8.8.8.8 8.8.4.4 1.1.1.1",
		},
		{
			name:     "empty",
			servers:  []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDNSServerCommand(tt.servers)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildDNSServerSelectCommand(t *testing.T) {
	tests := []struct {
		name     string
		sel      DNSServerSelect
		expected string
	}{
		{
			name: "single server single domain",
			sel: DNSServerSelect{
				ID:      1,
				Servers: []string{"192.168.1.1"},
				Domains: []string{"example.com"},
			},
			expected: "dns server select 1 192.168.1.1 example.com",
		},
		{
			name: "multiple servers multiple domains",
			sel: DNSServerSelect{
				ID:      2,
				Servers: []string{"10.0.0.1", "10.0.0.2"},
				Domains: []string{"*.local", "internal.net"},
			},
			expected: "dns server select 2 10.0.0.1 10.0.0.2 *.local internal.net",
		},
		{
			name: "invalid - no servers",
			sel: DNSServerSelect{
				ID:      1,
				Servers: []string{},
				Domains: []string{"example.com"},
			},
			expected: "",
		},
		{
			name: "invalid - no domains",
			sel: DNSServerSelect{
				ID:      1,
				Servers: []string{"192.168.1.1"},
				Domains: []string{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDNSServerSelectCommand(tt.sel)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildDNSStaticCommand(t *testing.T) {
	tests := []struct {
		name     string
		host     DNSHost
		expected string
	}{
		{
			name:     "valid host",
			host:     DNSHost{Name: "myhost", Address: "192.168.1.100"},
			expected: "dns static myhost 192.168.1.100",
		},
		{
			name:     "empty name",
			host:     DNSHost{Name: "", Address: "192.168.1.100"},
			expected: "",
		},
		{
			name:     "empty address",
			host:     DNSHost{Name: "myhost", Address: ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDNSStaticCommand(tt.host)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildDNSServiceCommand(t *testing.T) {
	if result := BuildDNSServiceCommand(true); result != "dns service on" {
		t.Errorf("Expected 'dns service on', got '%s'", result)
	}
	if result := BuildDNSServiceCommand(false); result != "dns service off" {
		t.Errorf("Expected 'dns service off', got '%s'", result)
	}
}

func TestBuildDNSPrivateSpoofCommand(t *testing.T) {
	if result := BuildDNSPrivateSpoofCommand(true); result != "dns private address spoof on" {
		t.Errorf("Expected 'dns private address spoof on', got '%s'", result)
	}
	if result := BuildDNSPrivateSpoofCommand(false); result != "dns private address spoof off" {
		t.Errorf("Expected 'dns private address spoof off', got '%s'", result)
	}
}

func TestBuildDNSDomainLookupCommand(t *testing.T) {
	if result := BuildDNSDomainLookupCommand(true); result != "dns domain lookup on" {
		t.Errorf("Expected 'dns domain lookup on', got '%s'", result)
	}
	if result := BuildDNSDomainLookupCommand(false); result != "no dns domain lookup" {
		t.Errorf("Expected 'no dns domain lookup', got '%s'", result)
	}
}

func TestBuildDNSDomainNameCommand(t *testing.T) {
	if result := BuildDNSDomainNameCommand("example.com"); result != "dns domain example.com" {
		t.Errorf("Expected 'dns domain example.com', got '%s'", result)
	}
	if result := BuildDNSDomainNameCommand(""); result != "" {
		t.Errorf("Expected empty string for empty domain, got '%s'", result)
	}
}

func TestBuildDeleteDNSCommands(t *testing.T) {
	if result := BuildDeleteDNSServerCommand(); result != "no dns server" {
		t.Errorf("Expected 'no dns server', got '%s'", result)
	}
	if result := BuildDeleteDNSServerSelectCommand(5); result != "no dns server select 5" {
		t.Errorf("Expected 'no dns server select 5', got '%s'", result)
	}
	if result := BuildDeleteDNSStaticCommand("myhost"); result != "no dns static myhost" {
		t.Errorf("Expected 'no dns static myhost', got '%s'", result)
	}
	if result := BuildDeleteDNSDomainNameCommand(); result != "no dns domain" {
		t.Errorf("Expected 'no dns domain', got '%s'", result)
	}
}

func TestValidateDNSConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    DNSConfig
		expectErr bool
	}{
		{
			name: "valid basic config",
			config: DNSConfig{
				NameServers: []string{"8.8.8.8", "8.8.4.4"},
			},
			expectErr: false,
		},
		{
			name: "invalid name server",
			config: DNSConfig{
				NameServers: []string{"invalid"},
			},
			expectErr: true,
		},
		{
			name: "too many name servers",
			config: DNSConfig{
				NameServers: []string{"8.8.8.8", "8.8.4.4", "1.1.1.1", "1.0.0.1"},
			},
			expectErr: true,
		},
		{
			name: "valid server select",
			config: DNSConfig{
				ServerSelect: []DNSServerSelect{
					{ID: 1, Servers: []string{"192.168.1.1"}, Domains: []string{"example.com"}},
				},
			},
			expectErr: false,
		},
		{
			name: "invalid server select ID",
			config: DNSConfig{
				ServerSelect: []DNSServerSelect{
					{ID: 0, Servers: []string{"192.168.1.1"}, Domains: []string{"example.com"}},
				},
			},
			expectErr: true,
		},
		{
			name: "server select no servers",
			config: DNSConfig{
				ServerSelect: []DNSServerSelect{
					{ID: 1, Servers: []string{}, Domains: []string{"example.com"}},
				},
			},
			expectErr: true,
		},
		{
			name: "server select no domains",
			config: DNSConfig{
				ServerSelect: []DNSServerSelect{
					{ID: 1, Servers: []string{"192.168.1.1"}, Domains: []string{}},
				},
			},
			expectErr: true,
		},
		{
			name: "valid static host",
			config: DNSConfig{
				Hosts: []DNSHost{
					{Name: "myhost", Address: "192.168.1.100"},
				},
			},
			expectErr: false,
		},
		{
			name: "static host empty name",
			config: DNSConfig{
				Hosts: []DNSHost{
					{Name: "", Address: "192.168.1.100"},
				},
			},
			expectErr: true,
		},
		{
			name: "static host invalid address",
			config: DNSConfig{
				Hosts: []DNSHost{
					{Name: "myhost", Address: "invalid"},
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDNSConfig(tt.config)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestBuildShowDNSConfigCommand(t *testing.T) {
	result := BuildShowDNSConfigCommand()
	if result != "show config | grep dns" {
		t.Errorf("Expected 'show config | grep dns', got '%s'", result)
	}
}

func TestBuildDeleteDNSCommand(t *testing.T) {
	commands := BuildDeleteDNSCommand()
	if len(commands) != 4 {
		t.Errorf("Expected 4 delete commands, got %d", len(commands))
	}
	expected := []string{
		"no dns server",
		"no dns domain",
		"dns service off",
		"dns private address spoof off",
	}
	for i, cmd := range commands {
		if cmd != expected[i] {
			t.Errorf("Command %d: expected '%s', got '%s'", i, expected[i], cmd)
		}
	}
}
