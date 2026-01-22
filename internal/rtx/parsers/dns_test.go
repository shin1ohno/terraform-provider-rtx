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
dns server select 2 10.0.0.1 edns=on 10.0.0.2 edns=on any .
`
	parser := NewDNSParser()
	config, err := parser.ParseDNSConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse DNS config: %v", err)
	}

	if len(config.ServerSelect) != 2 {
		t.Fatalf("Expected 2 server select entries, got %d", len(config.ServerSelect))
	}

	// Check first server select (no EDNS)
	sel1 := config.ServerSelect[0]
	if sel1.ID != 1 {
		t.Errorf("Expected server select ID 1, got %d", sel1.ID)
	}
	if len(sel1.Servers) != 1 || sel1.Servers[0].Address != "192.168.1.1" {
		t.Errorf("Expected server ['192.168.1.1'], got %v", sel1.Servers)
	}
	if sel1.Servers[0].EDNS {
		t.Error("Expected first server EDNS to be false")
	}
	if sel1.QueryPattern != "example.com" {
		t.Errorf("Expected query pattern 'example.com', got '%s'", sel1.QueryPattern)
	}

	// Check second server select with per-server EDNS and record type
	sel2 := config.ServerSelect[1]
	if sel2.ID != 2 {
		t.Errorf("Expected server select ID 2, got %d", sel2.ID)
	}
	if len(sel2.Servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(sel2.Servers))
	}
	if sel2.Servers[0].Address != "10.0.0.1" || !sel2.Servers[0].EDNS {
		t.Errorf("Expected first server '10.0.0.1' with EDNS=true, got %+v", sel2.Servers[0])
	}
	if sel2.Servers[1].Address != "10.0.0.2" || !sel2.Servers[1].EDNS {
		t.Errorf("Expected second server '10.0.0.2' with EDNS=true, got %+v", sel2.Servers[1])
	}
	if sel2.RecordType != "any" {
		t.Errorf("Expected record type 'any', got '%s'", sel2.RecordType)
	}
	if sel2.QueryPattern != "." {
		t.Errorf("Expected query pattern '.', got '%s'", sel2.QueryPattern)
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
		name         string
		input        string
		serviceOn    bool
		privateSpoof bool
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
			name:         "service recursive",
			input:        "dns service recursive",
			serviceOn:    true,
			privateSpoof: false,
		},
		{
			name:         "service recursive with spoof on",
			input:        "dns service recursive\ndns private address spoof on",
			serviceOn:    true,
			privateSpoof: true,
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
			name: "single server simple pattern",
			sel: DNSServerSelect{
				ID:           1,
				Servers:      []DNSServer{{Address: "192.168.1.1", EDNS: false}},
				QueryPattern: "example.com",
			},
			expected: "dns server select 1 192.168.1.1 example.com",
		},
		{
			name: "multiple servers with per-server EDNS and any record type",
			sel: DNSServerSelect{
				ID: 2,
				Servers: []DNSServer{
					{Address: "10.0.0.1", EDNS: true},
					{Address: "10.0.0.2", EDNS: true},
				},
				RecordType:   "any",
				QueryPattern: ".",
			},
			expected: "dns server select 2 10.0.0.1 edns=on 10.0.0.2 edns=on any .",
		},
		{
			name: "with original sender",
			sel: DNSServerSelect{
				ID:             3,
				Servers:        []DNSServer{{Address: "192.168.1.1", EDNS: false}},
				QueryPattern:   "*.corp.example.com",
				OriginalSender: "192.168.1.0/24",
			},
			expected: "dns server select 3 192.168.1.1 *.corp.example.com 192.168.1.0/24",
		},
		{
			name: "with restrict pp",
			sel: DNSServerSelect{
				ID:           4,
				Servers:      []DNSServer{{Address: "10.0.0.53", EDNS: false}},
				QueryPattern: ".",
				RestrictPP:   1,
			},
			expected: "dns server select 4 10.0.0.53 . restrict pp 1",
		},
		{
			name: "full options with per-server EDNS",
			sel: DNSServerSelect{
				ID:             10,
				Servers:        []DNSServer{{Address: "10.0.0.53", EDNS: true}},
				RecordType:     "aaaa",
				QueryPattern:   "*.corp.example.com",
				OriginalSender: "192.168.1.0/24",
				RestrictPP:     1,
			},
			expected: "dns server select 10 10.0.0.53 edns=on aaaa *.corp.example.com 192.168.1.0/24 restrict pp 1",
		},
		{
			name: "invalid - no servers",
			sel: DNSServerSelect{
				ID:           1,
				Servers:      []DNSServer{},
				QueryPattern: "example.com",
			},
			expected: "",
		},
		{
			name: "invalid - no query pattern",
			sel: DNSServerSelect{
				ID:           1,
				Servers:      []DNSServer{{Address: "192.168.1.1", EDNS: false}},
				QueryPattern: "",
			},
			expected: "",
		},
		{
			name: "mixed EDNS - first on, second off",
			sel: DNSServerSelect{
				ID: 5,
				Servers: []DNSServer{
					{Address: "10.0.0.1", EDNS: true},
					{Address: "10.0.0.2", EDNS: false},
				},
				RecordType:   "aaaa",
				QueryPattern: ".",
			},
			expected: "dns server select 5 10.0.0.1 edns=on 10.0.0.2 aaaa .",
		},
		// === ID boundary values ===
		{
			name: "ID minimum (1)",
			sel: DNSServerSelect{
				ID:           1,
				Servers:      []DNSServer{{Address: "8.8.8.8", EDNS: false}},
				QueryPattern: ".",
			},
			expected: "dns server select 1 8.8.8.8 .",
		},
		{
			name: "ID maximum (65535)",
			sel: DNSServerSelect{
				ID:           65535,
				Servers:      []DNSServer{{Address: "8.8.8.8", EDNS: false}},
				QueryPattern: ".",
			},
			expected: "dns server select 65535 8.8.8.8 .",
		},
		{
			name: "invalid - ID zero",
			sel: DNSServerSelect{
				ID:           0,
				Servers:      []DNSServer{{Address: "8.8.8.8", EDNS: false}},
				QueryPattern: ".",
			},
			expected: "",
		},
		// === All record types ===
		{
			name: "record type: a (default, omitted)",
			sel: DNSServerSelect{
				ID:           10,
				Servers:      []DNSServer{{Address: "8.8.8.8", EDNS: false}},
				RecordType:   "a",
				QueryPattern: ".",
			},
			expected: "dns server select 10 8.8.8.8 .",
		},
		{
			name: "record type: ptr",
			sel: DNSServerSelect{
				ID:           11,
				Servers:      []DNSServer{{Address: "8.8.8.8", EDNS: false}},
				RecordType:   "ptr",
				QueryPattern: ".",
			},
			expected: "dns server select 11 8.8.8.8 ptr .",
		},
		{
			name: "record type: mx",
			sel: DNSServerSelect{
				ID:           12,
				Servers:      []DNSServer{{Address: "8.8.8.8", EDNS: false}},
				RecordType:   "mx",
				QueryPattern: ".",
			},
			expected: "dns server select 12 8.8.8.8 mx .",
		},
		{
			name: "record type: ns",
			sel: DNSServerSelect{
				ID:           13,
				Servers:      []DNSServer{{Address: "8.8.8.8", EDNS: false}},
				RecordType:   "ns",
				QueryPattern: ".",
			},
			expected: "dns server select 13 8.8.8.8 ns .",
		},
		{
			name: "record type: cname",
			sel: DNSServerSelect{
				ID:           14,
				Servers:      []DNSServer{{Address: "8.8.8.8", EDNS: false}},
				RecordType:   "cname",
				QueryPattern: ".",
			},
			expected: "dns server select 14 8.8.8.8 cname .",
		},
		// === IPv6 servers ===
		{
			name: "single IPv6 server",
			sel: DNSServerSelect{
				ID:           20,
				Servers:      []DNSServer{{Address: "2001:4860:4860::8888", EDNS: false}},
				QueryPattern: ".",
			},
			expected: "dns server select 20 2001:4860:4860::8888 .",
		},
		{
			name: "two IPv6 servers with EDNS",
			sel: DNSServerSelect{
				ID: 21,
				Servers: []DNSServer{
					{Address: "2606:4700:4700::1111", EDNS: true},
					{Address: "2606:4700:4700::1001", EDNS: true},
				},
				RecordType:   "aaaa",
				QueryPattern: ".",
			},
			expected: "dns server select 21 2606:4700:4700::1111 edns=on 2606:4700:4700::1001 edns=on aaaa .",
		},
		{
			name: "mixed IPv4 and IPv6",
			sel: DNSServerSelect{
				ID: 22,
				Servers: []DNSServer{
					{Address: "8.8.8.8", EDNS: true},
					{Address: "2001:4860:4860::8888", EDNS: true},
				},
				QueryPattern: ".",
			},
			expected: "dns server select 22 8.8.8.8 edns=on 2001:4860:4860::8888 edns=on .",
		},
		// === Query pattern variations ===
		{
			name: "query pattern: wildcard subdomain",
			sel: DNSServerSelect{
				ID:           30,
				Servers:      []DNSServer{{Address: "10.0.0.53", EDNS: false}},
				QueryPattern: "*.internal.corp.example.com",
			},
			expected: "dns server select 30 10.0.0.53 *.internal.corp.example.com",
		},
		{
			name: "query pattern: simple domain",
			sel: DNSServerSelect{
				ID:           31,
				Servers:      []DNSServer{{Address: "10.0.0.53", EDNS: false}},
				QueryPattern: "home.local",
			},
			expected: "dns server select 31 10.0.0.53 home.local",
		},
		// === OriginalSender variations ===
		{
			name: "original sender: single IP",
			sel: DNSServerSelect{
				ID:             40,
				Servers:        []DNSServer{{Address: "10.0.0.53", EDNS: false}},
				QueryPattern:   ".",
				OriginalSender: "192.168.1.100",
			},
			expected: "dns server select 40 10.0.0.53 . 192.168.1.100",
		},
		{
			name: "original sender: CIDR /16",
			sel: DNSServerSelect{
				ID:             41,
				Servers:        []DNSServer{{Address: "10.0.0.53", EDNS: false}},
				QueryPattern:   ".",
				OriginalSender: "10.0.0.0/16",
			},
			expected: "dns server select 41 10.0.0.53 . 10.0.0.0/16",
		},
		// === EDNS combinations ===
		{
			name: "EDNS: both servers off",
			sel: DNSServerSelect{
				ID: 50,
				Servers: []DNSServer{
					{Address: "1.1.1.1", EDNS: false},
					{Address: "1.0.0.1", EDNS: false},
				},
				QueryPattern: ".",
			},
			expected: "dns server select 50 1.1.1.1 1.0.0.1 .",
		},
		{
			name: "EDNS: second server only",
			sel: DNSServerSelect{
				ID: 51,
				Servers: []DNSServer{
					{Address: "1.1.1.1", EDNS: false},
					{Address: "1.0.0.1", EDNS: true},
				},
				QueryPattern: ".",
			},
			expected: "dns server select 51 1.1.1.1 1.0.0.1 edns=on .",
		},
		// === Complex real-world example ===
		{
			name: "real-world: Cloudflare IPv6 with AAAA filter",
			sel: DNSServerSelect{
				ID: 500100,
				Servers: []DNSServer{
					{Address: "2606:4700:4700::1111", EDNS: true},
					{Address: "2606:4700:4700::1001", EDNS: true},
				},
				RecordType:   "aaaa",
				QueryPattern: ".",
			},
			expected: "dns server select 500100 2606:4700:4700::1111 edns=on 2606:4700:4700::1001 edns=on aaaa .",
		},
		{
			name: "real-world: internal DNS with restrict",
			sel: DNSServerSelect{
				ID:             100,
				Servers:        []DNSServer{{Address: "192.168.1.1", EDNS: false}},
				RecordType:     "any",
				QueryPattern:   "*.corp.local",
				OriginalSender: "192.168.0.0/16",
				RestrictPP:     1,
			},
			expected: "dns server select 100 192.168.1.1 any *.corp.local 192.168.0.0/16 restrict pp 1",
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
	// When enabled, should output "dns service recursive" (preferred form)
	if result := BuildDNSServiceCommand(true); result != "dns service recursive" {
		t.Errorf("Expected 'dns service recursive', got '%s'", result)
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
					{ID: 1, Servers: []DNSServer{{Address: "192.168.1.1"}}, QueryPattern: "example.com"},
				},
			},
			expectErr: false,
		},
		{
			name: "invalid server select ID",
			config: DNSConfig{
				ServerSelect: []DNSServerSelect{
					{ID: 0, Servers: []DNSServer{{Address: "192.168.1.1"}}, QueryPattern: "example.com"},
				},
			},
			expectErr: true,
		},
		{
			name: "server select no servers",
			config: DNSConfig{
				ServerSelect: []DNSServerSelect{
					{ID: 1, Servers: []DNSServer{}, QueryPattern: "example.com"},
				},
			},
			expectErr: true,
		},
		{
			name: "server select no query pattern",
			config: DNSConfig{
				ServerSelect: []DNSServerSelect{
					{ID: 1, Servers: []DNSServer{{Address: "192.168.1.1"}}, QueryPattern: ""},
				},
			},
			expectErr: true,
		},
		{
			name: "server select invalid record type",
			config: DNSConfig{
				ServerSelect: []DNSServerSelect{
					{ID: 1, Servers: []DNSServer{{Address: "192.168.1.1"}}, QueryPattern: ".", RecordType: "invalid"},
				},
			},
			expectErr: true,
		},
		{
			name: "server select valid record type",
			config: DNSConfig{
				ServerSelect: []DNSServerSelect{
					{ID: 1, Servers: []DNSServer{{Address: "192.168.1.1"}}, QueryPattern: ".", RecordType: "aaaa"},
				},
			},
			expectErr: false,
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

// TestParseDNSServerSelectEDNS verifies that edns=on is correctly extracted
// and stored in the EDNS boolean field, not in QueryPattern
func TestParseDNSServerSelectEDNS(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedEDNS       bool
		expectedRecordType string
		expectedPattern    string
		expectedSender     string
		expectedRestrictPP int
	}{
		{
			name:               "edns=on before record type",
			input:              "dns server select 1 10.0.0.1 10.0.0.2 edns=on any .",
			expectedEDNS:       true,
			expectedRecordType: "any",
			expectedPattern:    ".",
		},
		{
			name:               "edns=on without record type",
			input:              "dns server select 2 10.0.0.1 edns=on example.com",
			expectedEDNS:       true,
			expectedRecordType: "", // not specified in input, preserved as empty
			expectedPattern:    "example.com",
		},
		{
			name:               "no edns flag",
			input:              "dns server select 3 10.0.0.1 aaaa example.com",
			expectedEDNS:       false,
			expectedRecordType: "aaaa",
			expectedPattern:    "example.com",
		},
		{
			name:               "edns=on with original sender",
			input:              "dns server select 4 10.0.0.1 edns=on example.com 192.168.1.0/24",
			expectedEDNS:       true,
			expectedRecordType: "", // not specified in input
			expectedPattern:    "example.com",
			expectedSender:     "192.168.1.0/24",
		},
		{
			name:               "edns=on with restrict pp",
			input:              "dns server select 5 10.0.0.1 edns=on . restrict pp 1",
			expectedEDNS:       true,
			expectedRecordType: "", // not specified in input
			expectedPattern:    ".",
			expectedRestrictPP: 1,
		},
		{
			name:               "full options with edns=on",
			input:              "dns server select 6 10.0.0.53 edns=on aaaa *.corp.example.com 192.168.1.0/24 restrict pp 2",
			expectedEDNS:       true,
			expectedRecordType: "aaaa",
			expectedPattern:    "*.corp.example.com",
			expectedSender:     "192.168.1.0/24",
			expectedRestrictPP: 2,
		},
		{
			name:               "domain with equals sign (not edns)",
			input:              "dns server select 7 10.0.0.1 something=test.com",
			expectedEDNS:       false,
			expectedRecordType: "", // not specified in input
			expectedPattern:    "something=test.com",
		},
		{
			name:               "edns=on with wildcard domain",
			input:              "dns server select 8 192.168.1.1 edns=on *.internal.corp",
			expectedEDNS:       true,
			expectedRecordType: "", // not specified in input
			expectedPattern:    "*.internal.corp",
		},
		{
			name:               "ptr record type with edns=on",
			input:              "dns server select 9 10.0.0.53 edns=on ptr .",
			expectedEDNS:       true,
			expectedRecordType: "ptr",
			expectedPattern:    ".",
		},
	}

	parser := NewDNSParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parser.ParseDNSConfig(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			if len(config.ServerSelect) != 1 {
				t.Fatalf("Expected 1 server select entry, got %d", len(config.ServerSelect))
			}

			sel := config.ServerSelect[0]

			// Verify per-server EDNS is correctly set (check the first server)
			if len(sel.Servers) == 0 {
				t.Fatalf("Expected at least 1 server, got 0")
			}
			// Check if any server has EDNS enabled (for these single-server or multi-server tests)
			hasEDNS := false
			for _, srv := range sel.Servers {
				if srv.EDNS {
					hasEDNS = true
					break
				}
			}
			if hasEDNS != tt.expectedEDNS {
				t.Errorf("EDNS: expected %v, got %v (servers: %+v)", tt.expectedEDNS, hasEDNS, sel.Servers)
			}

			// Ensure QueryPattern does NOT contain "edns=on"
			if sel.QueryPattern == "edns=on" {
				t.Errorf("QueryPattern incorrectly contains 'edns=on': %s", sel.QueryPattern)
			}

			if sel.RecordType != tt.expectedRecordType {
				t.Errorf("RecordType: expected %s, got %s", tt.expectedRecordType, sel.RecordType)
			}

			if sel.QueryPattern != tt.expectedPattern {
				t.Errorf("QueryPattern: expected %s, got %s", tt.expectedPattern, sel.QueryPattern)
			}

			if sel.OriginalSender != tt.expectedSender {
				t.Errorf("OriginalSender: expected %s, got %s", tt.expectedSender, sel.OriginalSender)
			}

			if sel.RestrictPP != tt.expectedRestrictPP {
				t.Errorf("RestrictPP: expected %d, got %d", tt.expectedRestrictPP, sel.RestrictPP)
			}
		})
	}
}

// TestParseDNSServerSelectMultiEDNS tests parsing of multi-server entries with interleaved edns=on (REQ-3)
// Format: dns server select <id> <server1> [edns=on] <server2> [edns=on] [record_type] <query_pattern>
func TestParseDNSServerSelectMultiEDNS(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedServers []DNSServer // Now includes per-server EDNS
		expectedType    string
		expectedPattern string
	}{
		{
			name:  "two IPv6 servers with interleaved edns=on",
			input: "dns server select 500100 2606:4700:4700::1111 edns=on 2606:4700:4700::1001 edns=on aaaa .",
			expectedServers: []DNSServer{
				{Address: "2606:4700:4700::1111", EDNS: true},
				{Address: "2606:4700:4700::1001", EDNS: true},
			},
			expectedType:    "aaaa",
			expectedPattern: ".",
		},
		{
			name:  "two IPv4 servers with interleaved edns=on",
			input: "dns server select 1 8.8.8.8 edns=on 8.8.4.4 edns=on .",
			expectedServers: []DNSServer{
				{Address: "8.8.8.8", EDNS: true},
				{Address: "8.8.4.4", EDNS: true},
			},
			expectedType:    "",
			expectedPattern: ".",
		},
		{
			name:  "first server with edns=on, second without",
			input: "dns server select 2 1.1.1.1 edns=on 1.0.0.1 aaaa .",
			expectedServers: []DNSServer{
				{Address: "1.1.1.1", EDNS: true},
				{Address: "1.0.0.1", EDNS: false},
			},
			expectedType:    "aaaa",
			expectedPattern: ".",
		},
		{
			name:  "single server with edns=on",
			input: "dns server select 3 1.1.1.1 edns=on .",
			expectedServers: []DNSServer{
				{Address: "1.1.1.1", EDNS: true},
			},
			expectedType:    "",
			expectedPattern: ".",
		},
		{
			name:  "two servers with trailing edns=on (applies to last server)",
			input: "dns server select 4 8.8.8.8 8.8.4.4 edns=on .",
			expectedServers: []DNSServer{
				{Address: "8.8.8.8", EDNS: false},
				{Address: "8.8.4.4", EDNS: true},
			},
			expectedType:    "",
			expectedPattern: ".",
		},
		{
			name:  "mixed IPv4 and IPv6 with interleaved edns=on",
			input: "dns server select 5 8.8.8.8 edns=on 2606:4700:4700::1111 edns=on any .",
			expectedServers: []DNSServer{
				{Address: "8.8.8.8", EDNS: true},
				{Address: "2606:4700:4700::1111", EDNS: true},
			},
			expectedType:    "any",
			expectedPattern: ".",
		},
	}

	parser := NewDNSParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parser.ParseDNSConfig(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			if len(config.ServerSelect) != 1 {
				t.Fatalf("Expected 1 server select entry, got %d", len(config.ServerSelect))
			}

			sel := config.ServerSelect[0]

			// Verify servers with per-server EDNS
			if len(sel.Servers) != len(tt.expectedServers) {
				t.Errorf("Servers count: expected %d, got %d", len(tt.expectedServers), len(sel.Servers))
			} else {
				for i, expected := range tt.expectedServers {
					if sel.Servers[i].Address != expected.Address {
						t.Errorf("Server[%d] Address: expected %s, got %s", i, expected.Address, sel.Servers[i].Address)
					}
					if sel.Servers[i].EDNS != expected.EDNS {
						t.Errorf("Server[%d] EDNS: expected %v, got %v", i, expected.EDNS, sel.Servers[i].EDNS)
					}
				}
			}

			// Verify record type
			if sel.RecordType != tt.expectedType {
				t.Errorf("RecordType: expected %s, got %s", tt.expectedType, sel.RecordType)
			}

			// Verify query pattern
			if sel.QueryPattern != tt.expectedPattern {
				t.Errorf("QueryPattern: expected %s, got %s", tt.expectedPattern, sel.QueryPattern)
			}
		})
	}
}

// TestParseDNSServerSelectFieldOrder verifies correct field parsing order per RTX spec (REQ-1)
// Command format: dns server select <id> <server1> [edns=on] [server2] [edns=on] [record_type] <query_pattern> [original_sender] [restrict pp n]
func TestParseDNSServerSelectFieldOrder(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedServers []DNSServer
		expectedType    string
		expectedPattern string
		expectedSender  string
		expectedPP      int
	}{
		{
			name:  "two servers with aaaa record type",
			input: "dns server select 1 10.0.0.1 10.0.0.2 aaaa .",
			expectedServers: []DNSServer{
				{Address: "10.0.0.1", EDNS: false},
				{Address: "10.0.0.2", EDNS: false},
			},
			expectedType:    "aaaa",
			expectedPattern: ".",
		},
		{
			name:  "two servers with dot query pattern only",
			input: "dns server select 2 8.8.8.8 8.8.4.4 .",
			expectedServers: []DNSServer{
				{Address: "8.8.8.8", EDNS: false},
				{Address: "8.8.4.4", EDNS: false},
			},
			expectedType:    "", // not specified
			expectedPattern: ".",
		},
		{
			name:  "two servers with trailing edns (applies to last server)",
			input: "dns server select 3 10.0.0.1 10.0.0.2 edns=on any .",
			expectedServers: []DNSServer{
				{Address: "10.0.0.1", EDNS: false},
				{Address: "10.0.0.2", EDNS: true},
			},
			expectedType:    "any",
			expectedPattern: ".",
		},
		{
			name:  "single server with original_sender after query pattern",
			input: "dns server select 4 192.168.1.1 example.com 192.168.0.0/24",
			expectedServers: []DNSServer{
				{Address: "192.168.1.1", EDNS: false},
			},
			expectedType:    "",
			expectedPattern: "example.com",
			expectedSender:  "192.168.0.0/24",
		},
		{
			name:  "two servers should not confuse second server as original_sender",
			input: "dns server select 5 10.0.0.1 10.0.0.2 example.com",
			expectedServers: []DNSServer{
				{Address: "10.0.0.1", EDNS: false},
				{Address: "10.0.0.2", EDNS: false},
			},
			expectedType:    "",
			expectedPattern: "example.com",
			expectedSender:  "", // 10.0.0.2 is a server, not original_sender
		},
		{
			name:  "full options with per-server edns",
			input: "dns server select 6 10.0.0.53 edns=on 10.0.0.54 edns=on aaaa *.corp.example.com 192.168.1.0/24 restrict pp 1",
			expectedServers: []DNSServer{
				{Address: "10.0.0.53", EDNS: true},
				{Address: "10.0.0.54", EDNS: true},
			},
			expectedType:    "aaaa",
			expectedPattern: "*.corp.example.com",
			expectedSender:  "192.168.1.0/24",
			expectedPP:      1,
		},
		{
			name:  "aaaa record type preserved not defaulted to a",
			input: "dns server select 7 10.0.0.1 aaaa example.com",
			expectedServers: []DNSServer{
				{Address: "10.0.0.1", EDNS: false},
			},
			expectedType:    "aaaa", // must be preserved, not defaulted to "a"
			expectedPattern: "example.com",
		},
		{
			name:  "dot query pattern not misinterpreted",
			input: "dns server select 8 10.0.0.1 10.0.0.2 .",
			expectedServers: []DNSServer{
				{Address: "10.0.0.1", EDNS: false},
				{Address: "10.0.0.2", EDNS: false},
			},
			expectedType:    "",
			expectedPattern: ".", // dot is query pattern, not record type
		},
	}

	parser := NewDNSParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parser.ParseDNSConfig(tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(config.ServerSelect) != 1 {
				t.Fatalf("Expected 1 server select entry, got %d", len(config.ServerSelect))
			}

			sel := config.ServerSelect[0]

			// Verify servers array with per-server EDNS
			if len(sel.Servers) != len(tt.expectedServers) {
				t.Errorf("Servers count: expected %d, got %d (servers: %v)",
					len(tt.expectedServers), len(sel.Servers), sel.Servers)
			}
			for i, expected := range tt.expectedServers {
				if i < len(sel.Servers) {
					if sel.Servers[i].Address != expected.Address {
						t.Errorf("Server[%d] Address: expected %s, got %s", i, expected.Address, sel.Servers[i].Address)
					}
					if sel.Servers[i].EDNS != expected.EDNS {
						t.Errorf("Server[%d] EDNS: expected %v, got %v", i, expected.EDNS, sel.Servers[i].EDNS)
					}
				}
			}

			if sel.RecordType != tt.expectedType {
				t.Errorf("RecordType: expected %q, got %q", tt.expectedType, sel.RecordType)
			}

			if sel.QueryPattern != tt.expectedPattern {
				t.Errorf("QueryPattern: expected %q, got %q", tt.expectedPattern, sel.QueryPattern)
			}

			if sel.OriginalSender != tt.expectedSender {
				t.Errorf("OriginalSender: expected %q, got %q", tt.expectedSender, sel.OriginalSender)
			}

			if sel.RestrictPP != tt.expectedPP {
				t.Errorf("RestrictPP: expected %d, got %d", tt.expectedPP, sel.RestrictPP)
			}
		})
	}
}

// TestDNSServerSelectRoundTrip verifies that parsing and building produce consistent results
func TestDNSServerSelectRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		input    DNSServerSelect
		expected string
	}{
		{
			name: "with per-server EDNS enabled",
			input: DNSServerSelect{
				ID:           1,
				Servers:      []DNSServer{{Address: "10.0.0.1", EDNS: true}},
				RecordType:   "any",
				QueryPattern: ".",
			},
			expected: "dns server select 1 10.0.0.1 edns=on any .",
		},
		{
			name: "without EDNS",
			input: DNSServerSelect{
				ID:           2,
				Servers:      []DNSServer{{Address: "192.168.1.1", EDNS: false}},
				RecordType:   "", // empty means default, not emitted in command
				QueryPattern: "example.com",
			},
			expected: "dns server select 2 192.168.1.1 example.com",
		},
		{
			name: "per-server EDNS with aaaa record type",
			input: DNSServerSelect{
				ID: 3,
				Servers: []DNSServer{
					{Address: "10.0.0.53", EDNS: true},
					{Address: "10.0.0.54", EDNS: true},
				},
				RecordType:   "aaaa",
				QueryPattern: "*.ipv6.corp",
			},
			expected: "dns server select 3 10.0.0.53 edns=on 10.0.0.54 edns=on aaaa *.ipv6.corp",
		},
	}

	parser := NewDNSParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the command
			cmd := BuildDNSServerSelectCommand(tt.input)
			if cmd != tt.expected {
				t.Errorf("Build: expected %q, got %q", tt.expected, cmd)
			}

			// Parse the command back
			config, err := parser.ParseDNSConfig(cmd)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(config.ServerSelect) != 1 {
				t.Fatalf("Expected 1 entry, got %d", len(config.ServerSelect))
			}

			sel := config.ServerSelect[0]

			// Verify round-trip consistency
			if sel.ID != tt.input.ID {
				t.Errorf("ID: expected %d, got %d", tt.input.ID, sel.ID)
			}
			// Verify servers with per-server EDNS
			if len(sel.Servers) != len(tt.input.Servers) {
				t.Errorf("Servers count: expected %d, got %d", len(tt.input.Servers), len(sel.Servers))
			} else {
				for i, expected := range tt.input.Servers {
					if sel.Servers[i].Address != expected.Address {
						t.Errorf("Server[%d] Address: expected %s, got %s", i, expected.Address, sel.Servers[i].Address)
					}
					if sel.Servers[i].EDNS != expected.EDNS {
						t.Errorf("Server[%d] EDNS: expected %v, got %v", i, expected.EDNS, sel.Servers[i].EDNS)
					}
				}
			}
			if sel.RecordType != tt.input.RecordType {
				t.Errorf("RecordType: expected %s, got %s", tt.input.RecordType, sel.RecordType)
			}
			if sel.QueryPattern != tt.input.QueryPattern {
				t.Errorf("QueryPattern: expected %s, got %s", tt.input.QueryPattern, sel.QueryPattern)
			}
		})
	}
}

// TestParseDNSServerSelectREQ1Cases verifies specific test cases from REQ-1
// These test cases ensure the parser handles the field order correctly
func TestParseDNSServerSelectREQ1Cases(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedServers []DNSServer
		expectedType    string
		expectedPattern string
		expectedSender  string
		expectedPP      int
	}{
		{
			// Test Case 1: Two servers with dot query pattern and restrict pp
			// Input: 1.1.1.1 1.0.0.1 . restrict pp 1
			// Expected: servers=[1.1.1.1, 1.0.0.1], query_pattern=., no original_sender
			name:  "REQ1-case1: two servers dot restrict",
			input: "dns server select 1 1.1.1.1 1.0.0.1 . restrict pp 1",
			expectedServers: []DNSServer{
				{Address: "1.1.1.1", EDNS: false},
				{Address: "1.0.0.1", EDNS: false},
			},
			expectedType:    "",
			expectedPattern: ".",
			expectedSender:  "",
			expectedPP:      1,
		},
		{
			// Test Case 2: Single server with edns, aaaa record type, and dot query pattern
			// Input: 1.1.1.1 edns=on aaaa .
			// Expected: servers=[1.1.1.1 with edns=true], record_type=aaaa, query_pattern=.
			name:  "REQ1-case2: edns aaaa dot",
			input: "dns server select 2 1.1.1.1 edns=on aaaa .",
			expectedServers: []DNSServer{
				{Address: "1.1.1.1", EDNS: true},
			},
			expectedType:    "aaaa",
			expectedPattern: ".",
			expectedSender:  "",
			expectedPP:      0,
		},
		{
			// Test Case 3: Single server with record type 'a', domain pattern, and original_sender
			// Input: 8.8.8.8 a example.com 192.168.1.0/24
			// Expected: servers=[8.8.8.8], record_type=a, query_pattern=example.com, original_sender=192.168.1.0/24
			name:  "REQ1-case3: a type with original_sender",
			input: "dns server select 3 8.8.8.8 a example.com 192.168.1.0/24",
			expectedServers: []DNSServer{
				{Address: "8.8.8.8", EDNS: false},
			},
			expectedType:    "a",
			expectedPattern: "example.com",
			expectedSender:  "192.168.1.0/24",
			expectedPP:      0,
		},
	}

	parser := NewDNSParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parser.ParseDNSConfig(tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(config.ServerSelect) != 1 {
				t.Fatalf("Expected 1 server select entry, got %d", len(config.ServerSelect))
			}

			sel := config.ServerSelect[0]

			// Verify servers array with per-server EDNS
			if len(sel.Servers) != len(tt.expectedServers) {
				t.Errorf("Servers count: expected %d, got %d (servers: %v)",
					len(tt.expectedServers), len(sel.Servers), sel.Servers)
			}
			for i, expected := range tt.expectedServers {
				if i < len(sel.Servers) {
					if sel.Servers[i].Address != expected.Address {
						t.Errorf("Server[%d] Address: expected %s, got %s", i, expected.Address, sel.Servers[i].Address)
					}
					if sel.Servers[i].EDNS != expected.EDNS {
						t.Errorf("Server[%d] EDNS: expected %v, got %v", i, expected.EDNS, sel.Servers[i].EDNS)
					}
				}
			}

			if sel.RecordType != tt.expectedType {
				t.Errorf("RecordType: expected %q, got %q", tt.expectedType, sel.RecordType)
			}

			if sel.QueryPattern != tt.expectedPattern {
				t.Errorf("QueryPattern: expected %q, got %q", tt.expectedPattern, sel.QueryPattern)
			}

			if sel.OriginalSender != tt.expectedSender {
				t.Errorf("OriginalSender: expected %q, got %q", tt.expectedSender, sel.OriginalSender)
			}

			if sel.RestrictPP != tt.expectedPP {
				t.Errorf("RestrictPP: expected %d, got %d", tt.expectedPP, sel.RestrictPP)
			}
		})
	}
}

// TestParseDNSServerSelectStrictOrder verifies strict field order parsing (Task 1 success criteria)
// Success criteria from spec:
// - Test with two servers captures both in servers array
// - record_type aaaa preserved
// - query_pattern . captured correctly
func TestParseDNSServerSelectStrictOrder(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedServers []DNSServer
		expectedType    string
		expectedPattern string
		expectedSender  string
		expectedPP      int
	}{
		{
			// Success criteria 1: Two servers captured in servers array
			name:  "two servers both captured",
			input: "dns server select 1 10.0.0.1 10.0.0.2 example.com",
			expectedServers: []DNSServer{
				{Address: "10.0.0.1", EDNS: false},
				{Address: "10.0.0.2", EDNS: false},
			},
			expectedPattern: "example.com",
		},
		{
			// Success criteria 2: record_type aaaa preserved
			name:  "aaaa record type preserved",
			input: "dns server select 2 10.0.0.1 aaaa example.com",
			expectedServers: []DNSServer{
				{Address: "10.0.0.1", EDNS: false},
			},
			expectedType:    "aaaa",
			expectedPattern: "example.com",
		},
		{
			// Success criteria 3: query_pattern . captured correctly
			name:  "dot query pattern captured",
			input: "dns server select 3 10.0.0.1 10.0.0.2 .",
			expectedServers: []DNSServer{
				{Address: "10.0.0.1", EDNS: false},
				{Address: "10.0.0.2", EDNS: false},
			},
			expectedType:    "", // dot is not record type
			expectedPattern: ".",
		},
		{
			// Combined: two servers, aaaa, and dot pattern
			name:  "two servers aaaa dot combined",
			input: "dns server select 4 10.0.0.1 10.0.0.2 aaaa .",
			expectedServers: []DNSServer{
				{Address: "10.0.0.1", EDNS: false},
				{Address: "10.0.0.2", EDNS: false},
			},
			expectedType:    "aaaa",
			expectedPattern: ".",
		},
		{
			// IPv6 server addresses
			name:  "IPv6 servers",
			input: "dns server select 5 2001:db8::1 2001:db8::2 aaaa .",
			expectedServers: []DNSServer{
				{Address: "2001:db8::1", EDNS: false},
				{Address: "2001:db8::2", EDNS: false},
			},
			expectedType:    "aaaa",
			expectedPattern: ".",
		},
		{
			// Verify second server is NOT parsed as original_sender
			name:  "second server not as original_sender",
			input: "dns server select 6 192.168.1.1 192.168.1.2 .",
			expectedServers: []DNSServer{
				{Address: "192.168.1.1", EDNS: false},
				{Address: "192.168.1.2", EDNS: false},
			},
			expectedPattern: ".",
			expectedSender:  "", // 192.168.1.2 should be in servers, not original_sender
		},
		{
			// original_sender comes AFTER query_pattern
			name:  "original_sender after query_pattern",
			input: "dns server select 7 10.0.0.1 example.com 192.168.0.0/24",
			expectedServers: []DNSServer{
				{Address: "10.0.0.1", EDNS: false},
			},
			expectedPattern: "example.com",
			expectedSender:  "192.168.0.0/24",
		},
		{
			// Full example with all fields in correct order (trailing edns applies to last server)
			name:  "all fields correct order",
			input: "dns server select 8 10.0.0.1 10.0.0.2 edns=on aaaa . 192.168.1.0/24 restrict pp 1",
			expectedServers: []DNSServer{
				{Address: "10.0.0.1", EDNS: false},
				{Address: "10.0.0.2", EDNS: true},
			},
			expectedType:    "aaaa",
			expectedPattern: ".",
			expectedSender:  "192.168.1.0/24",
			expectedPP:      1,
		},
	}

	parser := NewDNSParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parser.ParseDNSConfig(tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(config.ServerSelect) != 1 {
				t.Fatalf("Expected 1 server select entry, got %d", len(config.ServerSelect))
			}

			sel := config.ServerSelect[0]

			// Verify servers count and values with per-server EDNS
			if len(sel.Servers) != len(tt.expectedServers) {
				t.Errorf("Servers count: expected %d, got %d (servers: %v)",
					len(tt.expectedServers), len(sel.Servers), sel.Servers)
			} else {
				for i, expected := range tt.expectedServers {
					if sel.Servers[i].Address != expected.Address {
						t.Errorf("Server[%d] Address: expected %s, got %s", i, expected.Address, sel.Servers[i].Address)
					}
					if sel.Servers[i].EDNS != expected.EDNS {
						t.Errorf("Server[%d] EDNS: expected %v, got %v", i, expected.EDNS, sel.Servers[i].EDNS)
					}
				}
			}

			if sel.RecordType != tt.expectedType {
				t.Errorf("RecordType: expected %q, got %q", tt.expectedType, sel.RecordType)
			}

			if sel.QueryPattern != tt.expectedPattern {
				t.Errorf("QueryPattern: expected %q, got %q", tt.expectedPattern, sel.QueryPattern)
			}

			if sel.OriginalSender != tt.expectedSender {
				t.Errorf("OriginalSender: expected %q, got %q", tt.expectedSender, sel.OriginalSender)
			}

			if sel.RestrictPP != tt.expectedPP {
				t.Errorf("RestrictPP: expected %d, got %d", tt.expectedPP, sel.RestrictPP)
			}
		})
	}
}

// TestParseDNSServerSelectMaxTwoServers verifies that server parsing is limited to max 2 IPs (Bug 2 fix)
// This prevents the 3rd+ fields from being incorrectly parsed as servers
func TestParseDNSServerSelectMaxTwoServers(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedServers []DNSServer
		expectedType    string
		expectedPattern string
		expectedSender  string
		expectedPP      int
	}{
		{
			// Bug case: 2 servers + trailing edns=on + aaaa
			name:  "2 servers with trailing edns=on and aaaa record_type",
			input: "dns server select 1 10.0.0.1 10.0.0.2 edns=on aaaa .",
			expectedServers: []DNSServer{
				{Address: "10.0.0.1", EDNS: false},
				{Address: "10.0.0.2", EDNS: true},
			},
			expectedType:    "aaaa",
			expectedPattern: ".",
		},
		{
			// Verify query_pattern "." is not confused with IP
			name:  "dot query pattern not confused with IP",
			input: "dns server select 2 8.8.8.8 8.8.4.4 .",
			expectedServers: []DNSServer{
				{Address: "8.8.8.8", EDNS: false},
				{Address: "8.8.4.4", EDNS: false},
			},
			expectedType:    "",
			expectedPattern: ".",
		},
		{
			// Test aaaa record_type parsing with trailing edns
			name:  "aaaa record type with trailing edns",
			input: "dns server select 3 1.1.1.1 1.0.0.1 edns=on aaaa . restrict pp 1",
			expectedServers: []DNSServer{
				{Address: "1.1.1.1", EDNS: false},
				{Address: "1.0.0.1", EDNS: true},
			},
			expectedType:    "aaaa",
			expectedPattern: ".",
			expectedPP:      1,
		},
		{
			// Verify 3rd IP is NOT parsed as server (max 2 servers)
			// The 3rd IP-looking field should not be in servers array
			name:  "max 2 servers enforced - 3rd IP becomes query pattern",
			input: "dns server select 4 10.0.0.1 10.0.0.2 192.168.1.1",
			expectedServers: []DNSServer{
				{Address: "10.0.0.1", EDNS: false},
				{Address: "10.0.0.2", EDNS: false},
			},
			expectedType:    "",
			expectedPattern: "192.168.1.1", // 3rd IP becomes query pattern, not server
		},
		{
			// Verify original_sender still works after query pattern
			name:  "2 servers with original_sender after pattern",
			input: "dns server select 5 10.0.0.1 10.0.0.2 example.com 192.168.0.0/24",
			expectedServers: []DNSServer{
				{Address: "10.0.0.1", EDNS: false},
				{Address: "10.0.0.2", EDNS: false},
			},
			expectedType:    "",
			expectedPattern: "example.com",
			expectedSender:  "192.168.0.0/24",
		},
		{
			// Full options: 2 servers + per-server edns + aaaa + pattern + original_sender + restrict
			name:  "full options with max servers and per-server edns",
			input: "dns server select 6 10.0.0.53 edns=on 10.0.0.54 edns=on aaaa *.corp.local 192.168.1.0/24 restrict pp 2",
			expectedServers: []DNSServer{
				{Address: "10.0.0.53", EDNS: true},
				{Address: "10.0.0.54", EDNS: true},
			},
			expectedType:    "aaaa",
			expectedPattern: "*.corp.local",
			expectedSender:  "192.168.1.0/24",
			expectedPP:      2,
		},
	}

	parser := NewDNSParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := parser.ParseDNSConfig(tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(config.ServerSelect) != 1 {
				t.Fatalf("Expected 1 server select entry, got %d", len(config.ServerSelect))
			}

			sel := config.ServerSelect[0]

			// Verify server count is limited to max 2
			if len(sel.Servers) > 2 {
				t.Errorf("Servers count exceeded max 2: got %d (servers: %v)", len(sel.Servers), sel.Servers)
			}

			// Verify expected servers with per-server EDNS
			if len(sel.Servers) != len(tt.expectedServers) {
				t.Errorf("Servers count: expected %d, got %d (servers: %v)",
					len(tt.expectedServers), len(sel.Servers), sel.Servers)
			} else {
				for i, expected := range tt.expectedServers {
					if sel.Servers[i].Address != expected.Address {
						t.Errorf("Server[%d] Address: expected %s, got %s", i, expected.Address, sel.Servers[i].Address)
					}
					if sel.Servers[i].EDNS != expected.EDNS {
						t.Errorf("Server[%d] EDNS: expected %v, got %v", i, expected.EDNS, sel.Servers[i].EDNS)
					}
				}
			}

			if sel.RecordType != tt.expectedType {
				t.Errorf("RecordType: expected %q, got %q", tt.expectedType, sel.RecordType)
			}

			if sel.QueryPattern != tt.expectedPattern {
				t.Errorf("QueryPattern: expected %q, got %q", tt.expectedPattern, sel.QueryPattern)
			}

			if sel.OriginalSender != tt.expectedSender {
				t.Errorf("OriginalSender: expected %q, got %q", tt.expectedSender, sel.OriginalSender)
			}

			if sel.RestrictPP != tt.expectedPP {
				t.Errorf("RestrictPP: expected %d, got %d", tt.expectedPP, sel.RestrictPP)
			}
		})
	}
}

// TestParseDNSConfig_LineWrapping tests that the parser correctly handles
// RTX router output where long lines are wrapped (e.g., "edns=on" becomes "edns\n=on")
func TestParseDNSConfig_LineWrapping(t *testing.T) {
	testCases := []struct {
		name string
		raw  string
	}{
		{
			name: "LF line ending",
			raw:  "dns server select 500100 2606:4700:4700::1111 edns=on 2606:4700:4700::1001 edns\n=on aaaa .\n",
		},
		{
			name: "CRLF line ending",
			raw:  "dns server select 500100 2606:4700:4700::1111 edns=on 2606:4700:4700::1001 edns\r\n=on aaaa .\r\n",
		},
		{
			name: "trailing whitespace before wrap",
			raw:  "dns server select 500100 2606:4700:4700::1111 edns=on 2606:4700:4700::1001 edns   \n=on aaaa .\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewDNSParser()
			config, err := parser.ParseDNSConfig(tc.raw)
			if err != nil {
				t.Fatalf("Failed to parse DNS config: %v", err)
			}

			if len(config.ServerSelect) != 1 {
				t.Fatalf("Expected 1 server_select entry, got %d", len(config.ServerSelect))
			}

			sel := config.ServerSelect[0]

			// Verify ID
			if sel.ID != 500100 {
				t.Errorf("Expected ID 500100, got %d", sel.ID)
			}

			// Verify servers
			if len(sel.Servers) != 2 {
				t.Fatalf("Expected 2 servers, got %d", len(sel.Servers))
			}

			// First server
			if sel.Servers[0].Address != "2606:4700:4700::1111" {
				t.Errorf("Server[0] Address: expected 2606:4700:4700::1111, got %s", sel.Servers[0].Address)
			}
			if !sel.Servers[0].EDNS {
				t.Error("Server[0] EDNS: expected true, got false")
			}

			// Second server - this is the one affected by line wrapping
			if sel.Servers[1].Address != "2606:4700:4700::1001" {
				t.Errorf("Server[1] Address: expected 2606:4700:4700::1001, got %s", sel.Servers[1].Address)
			}
			if !sel.Servers[1].EDNS {
				t.Error("Server[1] EDNS: expected true, got false (line wrapping issue?)")
			}

			// Verify record type
			if sel.RecordType != "aaaa" {
				t.Errorf("RecordType: expected aaaa, got %q", sel.RecordType)
			}

			// Verify query pattern
			if sel.QueryPattern != "." {
				t.Errorf("QueryPattern: expected '.', got %q", sel.QueryPattern)
			}
		})
	}
}
