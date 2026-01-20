package parsers

import (
	"testing"
)

// ============================================================================
// PPPoE Parser Tests
// ============================================================================

func TestParsePPPoEConfig_BasicConfiguration(t *testing.T) {
	raw := `
pp select 1
 description pp NTT FLETS
 pppoe use lan2
 pp auth accept pap chap
 pp auth myname username password
 pp always-on on
 ip pp address 192.168.1.1/24
 ip pp mtu 1454
 ip pp nat descriptor 1
pp enable 1
`
	parser := NewPPPParser()
	configs, err := parser.ParsePPPoEConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	cfg := configs[0]
	if cfg.Number != 1 {
		t.Errorf("Number: expected 1, got %d", cfg.Number)
	}
	if cfg.Name != "NTT FLETS" {
		t.Errorf("Name: expected 'NTT FLETS', got '%s'", cfg.Name)
	}
	if cfg.Interface != "lan2" {
		t.Errorf("Interface: expected 'lan2', got '%s'", cfg.Interface)
	}
	if cfg.Authentication == nil {
		t.Fatal("Authentication should not be nil")
	}
	if cfg.Authentication.Method != "pap" {
		t.Errorf("Auth method: expected 'pap', got '%s'", cfg.Authentication.Method)
	}
	if cfg.Authentication.Username != "username" {
		t.Errorf("Auth username: expected 'username', got '%s'", cfg.Authentication.Username)
	}
	if cfg.Authentication.Password != "password" {
		t.Errorf("Auth password: expected 'password', got '%s'", cfg.Authentication.Password)
	}
	if !cfg.AlwaysOn {
		t.Error("AlwaysOn: expected true, got false")
	}
	if !cfg.Enabled {
		t.Error("Enabled: expected true, got false")
	}
	if cfg.IPConfig == nil {
		t.Fatal("IPConfig should not be nil")
	}
	if cfg.IPConfig.Address != "192.168.1.1/24" {
		t.Errorf("Address: expected '192.168.1.1/24', got '%s'", cfg.IPConfig.Address)
	}
	if cfg.IPConfig.MTU != 1454 {
		t.Errorf("MTU: expected 1454, got %d", cfg.IPConfig.MTU)
	}
	if cfg.IPConfig.NATDescriptor != 1 {
		t.Errorf("NATDescriptor: expected 1, got %d", cfg.IPConfig.NATDescriptor)
	}
}

func TestParsePPPoEConfig_MultipleInterfaces(t *testing.T) {
	raw := `
pp select 1
 description pp Primary ISP
 pppoe use lan2
 pp auth accept chap
 pp auth myname user1 pass1
 pp always-on on
 ip pp nat descriptor 1
pp enable 1

pp select 2
 description pp Backup ISP
 pppoe use lan3
 pp auth accept chap
 pp auth myname user2 pass2
 pp always-on off
 ip pp nat descriptor 2
`
	parser := NewPPPParser()
	configs, err := parser.ParsePPPoEConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 2 {
		t.Fatalf("Expected 2 configs, got %d", len(configs))
	}

	// Find config with Number 1
	var cfg1, cfg2 *PPPoEConfig
	for i := range configs {
		if configs[i].Number == 1 {
			cfg1 = &configs[i]
		}
		if configs[i].Number == 2 {
			cfg2 = &configs[i]
		}
	}

	if cfg1 == nil {
		t.Fatal("PP 1 config not found")
	}
	if cfg2 == nil {
		t.Fatal("PP 2 config not found")
	}

	if cfg1.Name != "Primary ISP" {
		t.Errorf("PP 1 Name: expected 'Primary ISP', got '%s'", cfg1.Name)
	}
	if cfg1.Interface != "lan2" {
		t.Errorf("PP 1 Interface: expected 'lan2', got '%s'", cfg1.Interface)
	}
	if cfg1.Enabled != true {
		t.Error("PP 1 Enabled: expected true")
	}

	if cfg2.Name != "Backup ISP" {
		t.Errorf("PP 2 Name: expected 'Backup ISP', got '%s'", cfg2.Name)
	}
	if cfg2.Interface != "lan3" {
		t.Errorf("PP 2 Interface: expected 'lan3', got '%s'", cfg2.Interface)
	}
	if cfg2.AlwaysOn != false {
		t.Error("PP 2 AlwaysOn: expected false")
	}
}

func TestParsePPPoEConfig_WithServiceName(t *testing.T) {
	raw := `
pp select 1
 pppoe use lan2
 pppoe service-name FLET'S
 pp auth accept chap
 pp auth myname user pass
 pp always-on on
pp enable 1
`
	parser := NewPPPParser()
	configs, err := parser.ParsePPPoEConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	if configs[0].ServiceName != "FLET'S" {
		t.Errorf("ServiceName: expected \"FLET'S\", got '%s'", configs[0].ServiceName)
	}
}

func TestParsePPPoEConfig_WithTCPMSSLimit(t *testing.T) {
	raw := `
pp select 1
 pppoe use lan2
 pp auth accept chap
 pp auth myname user pass
 ip pp mtu 1454
 ip pp tcp mss limit 1414
 pp always-on on
pp enable 1
`
	parser := NewPPPParser()
	configs, err := parser.ParsePPPoEConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	if configs[0].IPConfig == nil {
		t.Fatal("IPConfig should not be nil")
	}
	if configs[0].IPConfig.MTU != 1454 {
		t.Errorf("MTU: expected 1454, got %d", configs[0].IPConfig.MTU)
	}
	if configs[0].IPConfig.TCPMSSLimit != 1414 {
		t.Errorf("TCPMSSLimit: expected 1414, got %d", configs[0].IPConfig.TCPMSSLimit)
	}
}

func TestParsePPPoEConfig_WithSecureFilter(t *testing.T) {
	raw := `
pp select 1
 pppoe use lan2
 pp auth accept chap
 pp auth myname user pass
 ip pp secure filter in 100 101 102
 ip pp secure filter out 200 201
 pp always-on on
pp enable 1
`
	parser := NewPPPParser()
	configs, err := parser.ParsePPPoEConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	if configs[0].IPConfig == nil {
		t.Fatal("IPConfig should not be nil")
	}
	if len(configs[0].IPConfig.SecureFilterIn) != 3 {
		t.Errorf("SecureFilterIn: expected 3 filters, got %d", len(configs[0].IPConfig.SecureFilterIn))
	}
	if len(configs[0].IPConfig.SecureFilterOut) != 2 {
		t.Errorf("SecureFilterOut: expected 2 filters, got %d", len(configs[0].IPConfig.SecureFilterOut))
	}
}

func TestParsePPPoEConfig_DisconnectTime(t *testing.T) {
	raw := `
pp select 1
 pppoe use lan2
 pp auth accept chap
 pp auth myname user pass
 pp disconnect time 60
 pp always-on off
pp enable 1
`
	parser := NewPPPParser()
	configs, err := parser.ParsePPPoEConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	if configs[0].DisconnectTimeout != 60 {
		t.Errorf("DisconnectTimeout: expected 60, got %d", configs[0].DisconnectTimeout)
	}
}

func TestParsePPPoEConfig_Empty(t *testing.T) {
	parser := NewPPPParser()
	configs, err := parser.ParsePPPoEConfig("")
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 0 {
		t.Errorf("Expected 0 configs, got %d", len(configs))
	}
}

func TestParsePPPoEConfig_CommentsIgnored(t *testing.T) {
	raw := `
# PPPoE configuration
pp select 1
 pppoe use lan2
 # Authentication
 pp auth accept chap
 pp auth myname user pass
 pp always-on on
! This is also a comment
pp enable 1
`
	parser := NewPPPParser()
	configs, err := parser.ParsePPPoEConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}
}

func TestParsePPPoEConfig_PPBind(t *testing.T) {
	raw := `
pp select 1
 pp bind lan2
 pp auth accept chap
 pp auth myname user pass
 pp always-on on
pp enable 1
`
	parser := NewPPPParser()
	configs, err := parser.ParsePPPoEConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	if configs[0].BindInterface != "lan2" {
		t.Errorf("BindInterface: expected 'lan2', got '%s'", configs[0].BindInterface)
	}
}

func TestParsePPPoEConfig_ACName(t *testing.T) {
	raw := `
pp select 1
 pppoe use lan2
 pppoe ac-name test-ac
 pp auth accept chap
 pp auth myname user pass
 pp always-on on
pp enable 1
`
	parser := NewPPPParser()
	configs, err := parser.ParsePPPoEConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	if configs[0].ACName != "test-ac" {
		t.Errorf("ACName: expected 'test-ac', got '%s'", configs[0].ACName)
	}
}

// ============================================================================
// PP Interface Config Tests
// ============================================================================

func TestParsePPInterfaceConfig(t *testing.T) {
	raw := `
pp select 1
 pppoe use lan2
 pp auth accept chap
 pp auth myname user pass
 ip pp address 192.168.1.1/24
 ip pp mtu 1454
 ip pp tcp mss limit 1414
 ip pp nat descriptor 1
 ip pp secure filter in 100 101
 ip pp secure filter out 200
 pp always-on on
pp enable 1

pp select 2
 pppoe use lan3
 ip pp address 192.168.2.1/24
 ip pp mtu 1500
`
	parser := NewPPPParser()
	config, err := parser.ParsePPInterfaceConfig(raw, 1)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if config.Address != "192.168.1.1/24" {
		t.Errorf("Address: expected '192.168.1.1/24', got '%s'", config.Address)
	}
	if config.MTU != 1454 {
		t.Errorf("MTU: expected 1454, got %d", config.MTU)
	}
	if config.TCPMSSLimit != 1414 {
		t.Errorf("TCPMSSLimit: expected 1414, got %d", config.TCPMSSLimit)
	}
	if config.NATDescriptor != 1 {
		t.Errorf("NATDescriptor: expected 1, got %d", config.NATDescriptor)
	}
	if len(config.SecureFilterIn) != 2 {
		t.Errorf("SecureFilterIn: expected 2, got %d", len(config.SecureFilterIn))
	}
	if len(config.SecureFilterOut) != 1 {
		t.Errorf("SecureFilterOut: expected 1, got %d", len(config.SecureFilterOut))
	}
}

// ============================================================================
// Command Builder Tests
// ============================================================================

func TestBuildPPSelectCommand(t *testing.T) {
	tests := []struct {
		name     string
		ppNum    int
		expected string
	}{
		{"valid pp 1", 1, "pp select 1"},
		{"valid pp 10", 10, "pp select 10"},
		{"invalid pp 0", 0, ""},
		{"invalid pp -1", -1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPPSelectCommand(tt.ppNum)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildPPDescriptionCommand(t *testing.T) {
	tests := []struct {
		name        string
		description string
		expected    string
	}{
		{"valid description", "NTT FLETS", "description pp NTT FLETS"},
		{"empty description", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPPDescriptionCommand(tt.description)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildPPPoEUseCommand(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		expected string
	}{
		{"lan2", "lan2", "pppoe use lan2"},
		{"lan3", "lan3", "pppoe use lan3"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPPPoEUseCommand(tt.iface)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildPPBindCommand(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		expected string
	}{
		{"lan2", "lan2", "pp bind lan2"},
		{"tunnel1", "tunnel1", "pp bind tunnel1"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPPBindCommand(tt.iface)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildPPPoEServiceNameCommand(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		expected    string
	}{
		{"FLETS", "FLET'S", "pppoe service-name FLET'S"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPPPoEServiceNameCommand(tt.serviceName)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildPPPAuthAcceptCommand(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected string
	}{
		{"pap", "pap", "pp auth accept pap"},
		{"chap", "chap", "pp auth accept chap"},
		{"mschap-v2", "mschap-v2", "pp auth accept mschap-v2"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPPPAuthAcceptCommand(tt.method)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildPPPAuthMynameCommand(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		expected string
	}{
		{"valid", "user", "pass", "pp auth myname user pass"},
		{"special chars", "user@isp.com", "p@ss!word", "pp auth myname user@isp.com p@ss!word"},
		{"empty username", "", "pass", ""},
		{"empty password", "user", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPPPAuthMynameCommand(tt.username, tt.password)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildPPAlwaysOnCommand(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		expected string
	}{
		{"enabled", true, "pp always-on on"},
		{"disabled", false, "pp always-on off"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPPAlwaysOnCommand(tt.enabled)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildPPDisconnectTimeCommand(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int
		expected string
	}{
		{"60 seconds", 60, "pp disconnect time 60"},
		{"300 seconds", 300, "pp disconnect time 300"},
		{"off", 0, "pp disconnect time off"},
		{"negative", -1, "pp disconnect time off"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPPDisconnectTimeCommand(tt.seconds)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildPPEnableCommand(t *testing.T) {
	tests := []struct {
		name     string
		ppNum    int
		expected string
	}{
		{"pp 1", 1, "pp enable 1"},
		{"pp 10", 10, "pp enable 10"},
		{"invalid", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildPPEnableCommand(tt.ppNum)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildIPPPAddressCommand(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected string
	}{
		{"with mask", "192.168.1.1/24", "ip pp address 192.168.1.1/24"},
		{"dhcp", "dhcp", "ip pp address dhcp"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildIPPPAddressCommand(tt.address)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildIPPPMTUCommand(t *testing.T) {
	tests := []struct {
		name     string
		mtu      int
		expected string
	}{
		{"1454", 1454, "ip pp mtu 1454"},
		{"1500", 1500, "ip pp mtu 1500"},
		{"zero", 0, ""},
		{"negative", -1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildIPPPMTUCommand(tt.mtu)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildIPPPTCPMSSLimitCommand(t *testing.T) {
	tests := []struct {
		name     string
		mss      int
		expected string
	}{
		{"1414", 1414, "ip pp tcp mss limit 1414"},
		{"1360", 1360, "ip pp tcp mss limit 1360"},
		{"zero", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildIPPPTCPMSSLimitCommand(tt.mss)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildIPPPNATDescriptorCommand(t *testing.T) {
	tests := []struct {
		name        string
		descriptor  int
		expected    string
	}{
		{"descriptor 1", 1, "ip pp nat descriptor 1"},
		{"descriptor 100", 100, "ip pp nat descriptor 100"},
		{"zero", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildIPPPNATDescriptorCommand(tt.descriptor)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildIPPPSecureFilterCommands(t *testing.T) {
	tests := []struct {
		name      string
		filterIDs []int
		expected  string
	}{
		{"single", []int{100}, "ip pp secure filter in 100"},
		{"multiple", []int{100, 101, 102}, "ip pp secure filter in 100 101 102"},
		{"empty", []int{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildIPPPSecureFilterInCommand(tt.filterIDs)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// Full Command Builder Tests
// ============================================================================

func TestBuildPPPoECommand(t *testing.T) {
	config := PPPoEConfig{
		Number:    1,
		Name:      "NTT FLETS",
		Interface: "lan2",
		Authentication: &PPPAuth{
			Method:   "chap",
			Username: "user",
			Password: "pass",
		},
		AlwaysOn: true,
		Enabled:  true,
		IPConfig: &PPIPConfig{
			Address:       "192.168.1.1/24",
			MTU:           1454,
			TCPMSSLimit:   1414,
			NATDescriptor: 1,
		},
	}

	commands := BuildPPPoECommand(config)

	// Verify essential commands are present
	expectedCommands := []string{
		"pp select 1",
		"description pp NTT FLETS",
		"pppoe use lan2",
		"pp auth accept chap",
		"pp auth myname user pass",
		"pp always-on on",
		"ip pp address 192.168.1.1/24",
		"ip pp mtu 1454",
		"ip pp tcp mss limit 1414",
		"ip pp nat descriptor 1",
		"pp enable 1",
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

func TestBuildDeletePPPoECommand(t *testing.T) {
	commands := BuildDeletePPPoECommand(1)

	if commands == nil {
		t.Fatal("Commands should not be nil")
	}

	// Verify pp disable is first
	if commands[0] != "pp disable 1" {
		t.Errorf("First command should be 'pp disable 1', got '%s'", commands[0])
	}

	// Verify pp select is present
	found := false
	for _, cmd := range commands {
		if cmd == "pp select 1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("pp select 1 not found in delete commands")
	}
}

func TestBuildDeletePPPoECommand_InvalidPPNum(t *testing.T) {
	commands := BuildDeletePPPoECommand(0)
	if commands != nil {
		t.Errorf("Expected nil for invalid PP num, got %v", commands)
	}
}

// ============================================================================
// Validation Tests
// ============================================================================

func TestValidatePPPoEConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    PPPoEConfig
		expectErr bool
	}{
		{
			name: "valid config",
			config: PPPoEConfig{
				Number:    1,
				Interface: "lan2",
				Authentication: &PPPAuth{
					Method:   "chap",
					Username: "user",
					Password: "pass",
				},
			},
			expectErr: false,
		},
		{
			name: "valid with bind interface",
			config: PPPoEConfig{
				Number:        1,
				BindInterface: "lan2",
			},
			expectErr: false,
		},
		{
			name: "invalid pp number 0",
			config: PPPoEConfig{
				Number:    0,
				Interface: "lan2",
			},
			expectErr: true,
		},
		{
			name: "missing interface",
			config: PPPoEConfig{
				Number: 1,
			},
			expectErr: true,
		},
		{
			name: "invalid interface name",
			config: PPPoEConfig{
				Number:    1,
				Interface: "invalid",
			},
			expectErr: true,
		},
		{
			name: "invalid auth method",
			config: PPPoEConfig{
				Number:    1,
				Interface: "lan2",
				Authentication: &PPPAuth{
					Method: "invalid",
				},
			},
			expectErr: true,
		},
		{
			name: "username without password",
			config: PPPoEConfig{
				Number:    1,
				Interface: "lan2",
				Authentication: &PPPAuth{
					Username: "user",
				},
			},
			expectErr: true,
		},
		{
			name: "invalid MTU too small",
			config: PPPoEConfig{
				Number:    1,
				Interface: "lan2",
				IPConfig: &PPIPConfig{
					MTU: 10,
				},
			},
			expectErr: true,
		},
		{
			name: "invalid MTU too large",
			config: PPPoEConfig{
				Number:    1,
				Interface: "lan2",
				IPConfig: &PPIPConfig{
					MTU: 2000,
				},
			},
			expectErr: true,
		},
		{
			name: "invalid TCP MSS",
			config: PPPoEConfig{
				Number:    1,
				Interface: "lan2",
				IPConfig: &PPIPConfig{
					TCPMSSLimit: 2000,
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePPPoEConfig(tt.config)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidatePPIPConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    PPIPConfig
		expectErr bool
	}{
		{
			name: "valid config",
			config: PPIPConfig{
				Address:       "192.168.1.1/24",
				MTU:           1454,
				TCPMSSLimit:   1414,
				NATDescriptor: 1,
			},
			expectErr: false,
		},
		{
			name: "invalid MTU",
			config: PPIPConfig{
				MTU: 9999,
			},
			expectErr: true,
		},
		{
			name: "invalid TCP MSS",
			config: PPIPConfig{
				TCPMSSLimit: 9999,
			},
			expectErr: true,
		},
		{
			name: "invalid NAT descriptor",
			config: PPIPConfig{
				NATDescriptor: 99999,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePPIPConfig(tt.config)
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

func TestPPPoERoundTrip(t *testing.T) {
	parser := NewPPPParser()

	original := PPPoEConfig{
		Number:    1,
		Name:      "Test ISP",
		Interface: "lan2",
		Authentication: &PPPAuth{
			Method:   "chap",
			Username: "testuser",
			Password: "testpass",
		},
		AlwaysOn: true,
		IPConfig: &PPIPConfig{
			Address:       "192.168.1.1/24",
			MTU:           1454,
			TCPMSSLimit:   1414,
			NATDescriptor: 1,
		},
	}

	// Build commands
	commands := BuildPPPoECommand(original)

	// Add pp enable command at the end
	raw := ""
	for _, cmd := range commands {
		raw += cmd + "\n"
	}

	// Parse back
	configs, err := parser.ParsePPPoEConfig(raw)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(configs) != 1 {
		t.Fatalf("Expected 1 config, got %d", len(configs))
	}

	parsed := configs[0]

	if parsed.Number != original.Number {
		t.Errorf("Number: expected %d, got %d", original.Number, parsed.Number)
	}
	if parsed.Name != original.Name {
		t.Errorf("Name: expected '%s', got '%s'", original.Name, parsed.Name)
	}
	if parsed.Interface != original.Interface {
		t.Errorf("Interface: expected '%s', got '%s'", original.Interface, parsed.Interface)
	}
	if parsed.Authentication == nil {
		t.Fatal("Authentication should not be nil")
	}
	if parsed.Authentication.Method != original.Authentication.Method {
		t.Errorf("Auth method: expected '%s', got '%s'", original.Authentication.Method, parsed.Authentication.Method)
	}
	if parsed.Authentication.Username != original.Authentication.Username {
		t.Errorf("Auth username: expected '%s', got '%s'", original.Authentication.Username, parsed.Authentication.Username)
	}
	if parsed.AlwaysOn != original.AlwaysOn {
		t.Errorf("AlwaysOn: expected %v, got %v", original.AlwaysOn, parsed.AlwaysOn)
	}
	if parsed.IPConfig == nil {
		t.Fatal("IPConfig should not be nil")
	}
	if parsed.IPConfig.MTU != original.IPConfig.MTU {
		t.Errorf("MTU: expected %d, got %d", original.IPConfig.MTU, parsed.IPConfig.MTU)
	}
}
