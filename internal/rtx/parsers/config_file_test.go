package parsers

import (
	"testing"
)

func TestConfigFileParser_SplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int // expected line count after splitting
	}{
		{
			name:     "empty input",
			input:    "",
			expected: 0,
		},
		{
			name:     "single line",
			input:    "ip route default gateway 192.168.1.1",
			expected: 1,
		},
		{
			name:     "multiple lines with LF",
			input:    "line1\nline2\nline3",
			expected: 3,
		},
		{
			name:     "multiple lines with CRLF",
			input:    "line1\r\nline2\r\nline3",
			expected: 3,
		},
		{
			name:     "mixed line endings",
			input:    "line1\r\nline2\nline3\rline4",
			expected: 4,
		},
		{
			name:     "trailing newline",
			input:    "line1\nline2\n",
			expected: 2,
		},
	}

	parser := NewConfigFileParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if result.LineCount != tt.expected {
				t.Errorf("LineCount = %d, want %d", result.LineCount, tt.expected)
			}
		})
	}
}

func TestConfigFileParser_CommentRemoval(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedLines int // expected non-comment, non-empty lines
	}{
		{
			name:          "comment line only",
			input:         "# this is a comment",
			expectedLines: 0,
		},
		{
			name:          "comment with leading spaces",
			input:         "  # indented comment",
			expectedLines: 0,
		},
		{
			name: "mixed comments and commands",
			input: `# Admin section
login password test123
# Next section
ip route default gateway 192.168.1.1`,
			expectedLines: 2,
		},
		{
			name: "comment block from sample config",
			input: `#
# Admin
#
login password test-login-password-123
administrator password test-admin-password-456`,
			expectedLines: 2,
		},
	}

	parser := NewConfigFileParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if result.CommandCount != tt.expectedLines {
				t.Errorf("CommandCount = %d, want %d", result.CommandCount, tt.expectedLines)
			}
		})
	}
}

func TestConfigFileParser_TunnelContextTracking(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedContexts []ParseContext
	}{
		{
			name: "single tunnel select",
			input: `tunnel select 1
 tunnel encapsulation l2tpv3
 tunnel endpoint name test.example.com fqdn`,
			expectedContexts: []ParseContext{
				{Type: ContextTunnel, ID: 1},
			},
		},
		{
			name: "multiple tunnel selects",
			input: `tunnel select 1
 tunnel encapsulation l2tpv3
tunnel select 2
 tunnel encapsulation ipsec`,
			expectedContexts: []ParseContext{
				{Type: ContextTunnel, ID: 1},
				{Type: ContextTunnel, ID: 2},
			},
		},
		{
			name: "nested ipsec tunnel context",
			input: `tunnel select 1
 tunnel encapsulation l2tpv3
 ipsec tunnel 101
  ipsec sa policy 101 1 esp aes-cbc sha-hmac
  ipsec ike pre-shared-key 1 text test-secret
 l2tp tunnel auth on test-auth
 tunnel enable 1`,
			expectedContexts: []ParseContext{
				{Type: ContextTunnel, ID: 1},
				{Type: ContextIPsecTunnel, ID: 101},
			},
		},
	}

	parser := NewConfigFileParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if len(result.Contexts) != len(tt.expectedContexts) {
				t.Errorf("Context count = %d, want %d", len(result.Contexts), len(tt.expectedContexts))
				return
			}
			for i, ctx := range result.Contexts {
				if ctx.Type != tt.expectedContexts[i].Type {
					t.Errorf("Context[%d].Type = %v, want %v", i, ctx.Type, tt.expectedContexts[i].Type)
				}
				if ctx.ID != tt.expectedContexts[i].ID {
					t.Errorf("Context[%d].ID = %d, want %d", i, ctx.ID, tt.expectedContexts[i].ID)
				}
			}
		})
	}
}

func TestConfigFileParser_PPContextTracking(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedContexts []ParseContext
	}{
		{
			name: "pp select anonymous",
			input: `pp select anonymous
 pp bind tunnel1
 pp auth request mschap-v2
 pp auth username vpnuser test-vpn-password-789
 pp enable anonymous`,
			expectedContexts: []ParseContext{
				{Type: ContextPP, ID: 0, Name: "anonymous"},
			},
		},
		{
			name: "pp select numbered",
			input: `pp select 1
 pppoe use lan2
 pp auth accept chap
 pp auth myname user@isp.com password123`,
			expectedContexts: []ParseContext{
				{Type: ContextPP, ID: 1},
			},
		},
		{
			name: "multiple pp selects",
			input: `pp select 1
 pppoe use lan2
pp select 2
 pppoe use lan3`,
			expectedContexts: []ParseContext{
				{Type: ContextPP, ID: 1},
				{Type: ContextPP, ID: 2},
			},
		},
	}

	parser := NewConfigFileParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if len(result.Contexts) != len(tt.expectedContexts) {
				t.Errorf("Context count = %d, want %d", len(result.Contexts), len(tt.expectedContexts))
				return
			}
			for i, ctx := range result.Contexts {
				if ctx.Type != tt.expectedContexts[i].Type {
					t.Errorf("Context[%d].Type = %v, want %v", i, ctx.Type, tt.expectedContexts[i].Type)
				}
				if ctx.ID != tt.expectedContexts[i].ID {
					t.Errorf("Context[%d].ID = %d, want %d", i, ctx.ID, tt.expectedContexts[i].ID)
				}
				if ctx.Name != tt.expectedContexts[i].Name {
					t.Errorf("Context[%d].Name = %q, want %q", i, ctx.Name, tt.expectedContexts[i].Name)
				}
			}
		})
	}
}

func TestConfigFileParser_MixedContexts(t *testing.T) {
	input := `#
# Tunnels
#
pp select anonymous
 pp bind tunnel1
 pp auth request mschap-v2
 pp auth username vpnuser test-vpn-password-789
 ppp ipcp ipaddress on
 pp enable anonymous

tunnel select 1
 tunnel encapsulation l2tpv3
 tunnel endpoint name test.example.com fqdn
 ipsec tunnel 101
  ipsec sa policy 101 1 esp aes-cbc sha-hmac
  ipsec ike pre-shared-key 1 text test-ike-psk-secret
  ipsec ike remote address 1 test.example.com
 l2tp tunnel auth on test-l2tp-auth-secret
 tunnel enable 1`

	parser := NewConfigFileParser()
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should have 3 contexts: pp anonymous, tunnel 1, ipsec tunnel 101
	expectedContexts := []ParseContext{
		{Type: ContextPP, ID: 0, Name: "anonymous"},
		{Type: ContextTunnel, ID: 1},
		{Type: ContextIPsecTunnel, ID: 101},
	}

	if len(result.Contexts) != len(expectedContexts) {
		t.Errorf("Context count = %d, want %d", len(result.Contexts), len(expectedContexts))
		for i, ctx := range result.Contexts {
			t.Logf("  Context[%d]: Type=%v, ID=%d, Name=%q", i, ctx.Type, ctx.ID, ctx.Name)
		}
		return
	}

	for i, ctx := range result.Contexts {
		if ctx.Type != expectedContexts[i].Type {
			t.Errorf("Context[%d].Type = %v, want %v", i, ctx.Type, expectedContexts[i].Type)
		}
		if ctx.ID != expectedContexts[i].ID {
			t.Errorf("Context[%d].ID = %d, want %d", i, ctx.ID, expectedContexts[i].ID)
		}
		if ctx.Name != expectedContexts[i].Name {
			t.Errorf("Context[%d].Name = %q, want %q", i, ctx.Name, expectedContexts[i].Name)
		}
	}
}

func TestConfigFileParser_CommandsByContext(t *testing.T) {
	input := `tunnel select 1
 tunnel encapsulation l2tpv3
 tunnel endpoint name test.example.com fqdn
 l2tp tunnel auth on secret123
 tunnel enable 1`

	parser := NewConfigFileParser()
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check that commands are associated with the tunnel context
	tunnelCtx := ParseContext{Type: ContextTunnel, ID: 1}
	commands := result.GetCommandsInContext(tunnelCtx)

	if len(commands) == 0 {
		t.Error("Expected commands in tunnel context, got none")
		return
	}

	// Verify some specific commands are present
	foundEncapsulation := false
	foundEndpoint := false
	foundAuth := false
	for _, cmd := range commands {
		if cmd.Line == "tunnel encapsulation l2tpv3" {
			foundEncapsulation = true
		}
		if cmd.Line == "tunnel endpoint name test.example.com fqdn" {
			foundEndpoint = true
		}
		if cmd.Line == "l2tp tunnel auth on secret123" {
			foundAuth = true
		}
	}

	if !foundEncapsulation {
		t.Error("Expected to find 'tunnel encapsulation l2tpv3' command")
	}
	if !foundEndpoint {
		t.Error("Expected to find 'tunnel endpoint name' command")
	}
	if !foundAuth {
		t.Error("Expected to find 'l2tp tunnel auth on' command")
	}
}

func TestConfigFileParser_GlobalCommands(t *testing.T) {
	input := `login password test123
ip route default gateway 192.168.1.1
tunnel select 1
 tunnel encapsulation l2tpv3
ip filter 100 pass * * * * *`

	parser := NewConfigFileParser()
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Get global (non-context) commands
	globalCommands := result.GetGlobalCommands()

	if len(globalCommands) != 3 {
		t.Errorf("Expected 3 global commands, got %d", len(globalCommands))
		for i, cmd := range globalCommands {
			t.Logf("  GlobalCommand[%d]: %q", i, cmd.Line)
		}
		return
	}

	// Verify the specific global commands
	expectedGlobal := []string{
		"login password test123",
		"ip route default gateway 192.168.1.1",
		"ip filter 100 pass * * * * *",
	}

	for i, cmd := range globalCommands {
		if cmd.Line != expectedGlobal[i] {
			t.Errorf("GlobalCommand[%d] = %q, want %q", i, cmd.Line, expectedGlobal[i])
		}
	}
}

func TestConfigFileParser_SampleConfig(t *testing.T) {
	// Sample config from requirements.md Appendix A
	sampleConfig := `#
# Admin
#
login password test-login-password-123
administrator password test-admin-password-456
login user testuser encrypted TESTENCRYPTEDHASH123456789
user attribute administrator=off connection=off gui-page=dashboard,lan-map,config login-timer=300
user attribute testuser connection=serial,telnet,remote,ssh,sftp,http gui-page=dashboard,lan-map,config login-timer=3600
timezone +09:00
console character ja.utf8
console prompt "[TEST-RTX] "

httpd host any
sshd service on
sshd host lan1
sftpd host lan1

#
# WAN connection
#
description lan2 test-wan
ip lan2 address 198.51.100.1/24
ip lan2 nat descriptor 1000
ip lan2 secure filter in 200020 200099
ip lan2 secure filter out 200099 dynamic 200080 200081

#
# IP configuration
#
ip route default gateway 198.51.100.254
ip route 10.0.0.0/8 gateway 192.0.2.1

#
# LAN configuration
#
ip lan1 address 192.0.2.253/24

#
# Services
#
dhcp service server
dhcp scope 1 192.0.2.100-192.0.2.199/24 gateway 192.0.2.253 expire 12:00
dhcp scope bind 1 192.0.2.100 01:00:11:22:33:44:55
dhcp scope option 1 dns=192.0.2.10

dns host lan1
dns service recursive
dns server select 1 192.0.2.10 edns=on any example.local
dns server select 500000 8.8.8.8 edns=on 8.8.4.4 edns=on any .

#
# Tunnels
#
pp select anonymous
 pp bind tunnel1
 pp auth request mschap-v2
 pp auth username vpnuser test-vpn-password-789
 ppp ipcp ipaddress on
 pp enable anonymous

tunnel select 1
 tunnel encapsulation l2tpv3
 tunnel endpoint name test.example.com fqdn
 ipsec tunnel 101
  ipsec sa policy 101 1 esp aes-cbc sha-hmac
  ipsec ike pre-shared-key 1 text test-ike-psk-secret
  ipsec ike remote address 1 test.example.com
 l2tp tunnel auth on test-l2tp-auth-secret
 tunnel enable 1

#
# Filters
#
ip filter 200020 reject * * udp,tcp 135 *
ip filter 200099 pass * * * * *
ip filter dynamic 200080 * * ftp
ip filter dynamic 200081 * * www

nat descriptor type 1000 masquerade
nat descriptor masquerade static 1000 1 192.0.2.253 tcp 22`

	parser := NewConfigFileParser()
	result, err := parser.Parse(sampleConfig)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Basic structure verification
	if result.LineCount == 0 {
		t.Error("Expected non-zero line count")
	}

	if result.CommandCount == 0 {
		t.Error("Expected non-zero command count")
	}

	// Verify contexts were detected
	expectedContextTypes := map[ContextType]int{
		ContextPP:          1, // pp select anonymous
		ContextTunnel:      1, // tunnel select 1
		ContextIPsecTunnel: 1, // ipsec tunnel 101
	}

	contextCounts := make(map[ContextType]int)
	for _, ctx := range result.Contexts {
		contextCounts[ctx.Type]++
	}

	for ctxType, expectedCount := range expectedContextTypes {
		if contextCounts[ctxType] != expectedCount {
			t.Errorf("Context type %v count = %d, want %d", ctxType, contextCounts[ctxType], expectedCount)
		}
	}

	// Verify global commands include expected entries
	globalCmds := result.GetGlobalCommands()
	foundLoginPassword := false
	foundIPRoute := false
	foundDHCPScope := false
	foundIPFilter := false

	for _, cmd := range globalCmds {
		if cmd.Line == "login password test-login-password-123" {
			foundLoginPassword = true
		}
		if cmd.Line == "ip route default gateway 198.51.100.254" {
			foundIPRoute = true
		}
		if cmd.Line == "dhcp scope 1 192.0.2.100-192.0.2.199/24 gateway 192.0.2.253 expire 12:00" {
			foundDHCPScope = true
		}
		if cmd.Line == "ip filter 200020 reject * * udp,tcp 135 *" {
			foundIPFilter = true
		}
	}

	if !foundLoginPassword {
		t.Error("Expected to find 'login password' in global commands")
	}
	if !foundIPRoute {
		t.Error("Expected to find 'ip route default' in global commands")
	}
	if !foundDHCPScope {
		t.Error("Expected to find 'dhcp scope' in global commands")
	}
	if !foundIPFilter {
		t.Error("Expected to find 'ip filter' in global commands")
	}
}

func TestConfigFileParser_IndentedContext(t *testing.T) {
	// Test that indented lines are correctly associated with their context
	input := `tunnel select 1
 tunnel encapsulation l2tpv3
  ipsec tunnel 101
   ipsec sa policy 101 1 esp aes-cbc
 l2tp always-on on
tunnel enable 1`

	parser := NewConfigFileParser()
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// The ipsec tunnel 101 should be a nested context
	if len(result.Contexts) < 2 {
		t.Errorf("Expected at least 2 contexts, got %d", len(result.Contexts))
	}

	// Find tunnel context and ipsec context
	var tunnelCtx, ipsecCtx *ParseContext
	for i := range result.Contexts {
		if result.Contexts[i].Type == ContextTunnel && result.Contexts[i].ID == 1 {
			tunnelCtx = &result.Contexts[i]
		}
		if result.Contexts[i].Type == ContextIPsecTunnel && result.Contexts[i].ID == 101 {
			ipsecCtx = &result.Contexts[i]
		}
	}

	if tunnelCtx == nil {
		t.Error("Expected to find tunnel context with ID 1")
	}
	if ipsecCtx == nil {
		t.Error("Expected to find ipsec tunnel context with ID 101")
	}
}

// ============================================================================
// Resource Extraction Tests
// ============================================================================

func TestConfigFileParser_ExtractStaticRoutes(t *testing.T) {
	input := `ip route default gateway 198.51.100.254
ip route 10.0.0.0/8 gateway 192.0.2.1
ip route 172.16.0.0/12 gateway pp 1 weight 2`

	parser := NewConfigFileParser()
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	routes := result.ExtractStaticRoutes()

	if len(routes) != 3 {
		t.Errorf("Expected 3 routes, got %d", len(routes))
	}

	// Check default route
	foundDefault := false
	for _, r := range routes {
		if r.Prefix == "0.0.0.0" && r.Mask == "0.0.0.0" {
			foundDefault = true
			if len(r.NextHops) != 1 {
				t.Errorf("Expected 1 next hop for default route, got %d", len(r.NextHops))
			}
			if r.NextHops[0].NextHop != "198.51.100.254" {
				t.Errorf("Expected gateway 198.51.100.254, got %s", r.NextHops[0].NextHop)
			}
		}
	}
	if !foundDefault {
		t.Error("Expected to find default route")
	}
}

func TestConfigFileParser_ExtractDHCPScopes(t *testing.T) {
	input := `dhcp service server
dhcp scope 1 192.0.2.100-192.0.2.199/24 gateway 192.0.2.253 expire 12:00
dhcp scope bind 1 192.0.2.100 01:00:11:22:33:44:55
dhcp scope option 1 dns=192.0.2.10`

	parser := NewConfigFileParser()
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	scopes := result.ExtractDHCPScopes()

	if len(scopes) != 1 {
		t.Errorf("Expected 1 scope, got %d", len(scopes))
		return
	}

	scope := scopes[0]
	if scope.ScopeID != 1 {
		t.Errorf("Expected scope ID 1, got %d", scope.ScopeID)
	}
	if scope.RangeStart != "192.0.2.100" {
		t.Errorf("Expected range start 192.0.2.100, got %s", scope.RangeStart)
	}
	if scope.RangeEnd != "192.0.2.199" {
		t.Errorf("Expected range end 192.0.2.199, got %s", scope.RangeEnd)
	}
	if len(scope.Options.DNSServers) != 1 || scope.Options.DNSServers[0] != "192.0.2.10" {
		t.Errorf("Expected DNS server 192.0.2.10, got %v", scope.Options.DNSServers)
	}
}

func TestConfigFileParser_ExtractNATMasquerade(t *testing.T) {
	input := `nat descriptor type 1000 masquerade
nat descriptor masquerade static 1000 1 192.0.2.253 tcp 22`

	parser := NewConfigFileParser()
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	nats := result.ExtractNATMasquerade()

	if len(nats) != 1 {
		t.Errorf("Expected 1 NAT descriptor, got %d", len(nats))
		return
	}

	nat := nats[0]
	if nat.DescriptorID != 1000 {
		t.Errorf("Expected descriptor ID 1000, got %d", nat.DescriptorID)
	}
	if len(nat.StaticEntries) != 1 {
		t.Errorf("Expected 1 static entry, got %d", len(nat.StaticEntries))
	}
}

func TestConfigFileParser_ExtractIPFilters(t *testing.T) {
	input := `ip filter 200020 reject * * udp,tcp 135 *
ip filter 200099 pass * * * * *
ip filter dynamic 200080 * * ftp
ip filter dynamic 200081 * * www`

	parser := NewConfigFileParser()
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	staticFilters := result.ExtractIPFilters()
	dynamicFilters := result.ExtractIPFiltersDynamic()

	if len(staticFilters) != 2 {
		t.Errorf("Expected 2 static filters, got %d", len(staticFilters))
	}

	if len(dynamicFilters) != 2 {
		t.Errorf("Expected 2 dynamic filters, got %d", len(dynamicFilters))
	}

	// Check specific filter
	foundReject := false
	for _, f := range staticFilters {
		if f.Number == 200020 {
			foundReject = true
			if f.Action != "reject" {
				t.Errorf("Expected action 'reject', got '%s'", f.Action)
			}
		}
	}
	if !foundReject {
		t.Error("Expected to find filter 200020")
	}
}

func TestConfigFileParser_ExtractPasswords(t *testing.T) {
	input := `login password test-login-password-123
administrator password test-admin-password-456
login user testuser encrypted TESTENCRYPTEDHASH123456789`

	parser := NewConfigFileParser()
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	passwords := result.ExtractPasswords()

	// Check login password
	if passwords.LoginPassword != "test-login-password-123" {
		t.Errorf("Expected login password 'test-login-password-123', got '%s'", passwords.LoginPassword)
	}

	// Check admin password
	if passwords.AdminPassword != "test-admin-password-456" {
		t.Errorf("Expected admin password 'test-admin-password-456', got '%s'", passwords.AdminPassword)
	}

	// Check user (encrypted password should be marked as unknown)
	if len(passwords.Users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(passwords.Users))
	}
	if passwords.Users[0].Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", passwords.Users[0].Username)
	}
	if !passwords.Users[0].Encrypted {
		t.Error("Expected user password to be marked as encrypted")
	}
}

func TestConfigFileParser_ExtractPasswords_IPsecPSK(t *testing.T) {
	input := `tunnel select 1
 tunnel encapsulation l2tpv3
 ipsec tunnel 101
  ipsec ike pre-shared-key 1 text test-ike-psk-secret
  ipsec ike remote address 1 test.example.com
 l2tp tunnel auth on test-l2tp-auth-secret
 tunnel enable 1`

	parser := NewConfigFileParser()
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	passwords := result.ExtractPasswords()

	// Check IPsec PSK
	if len(passwords.IPsecPSK) != 1 {
		t.Errorf("Expected 1 IPsec PSK, got %d", len(passwords.IPsecPSK))
	} else {
		psk := passwords.IPsecPSK[0]
		if psk.ID != 1 {
			t.Errorf("Expected IPsec PSK ID 1, got %d", psk.ID)
		}
		if psk.Secret != "test-ike-psk-secret" {
			t.Errorf("Expected PSK 'test-ike-psk-secret', got '%s'", psk.Secret)
		}
	}

	// Check L2TP tunnel auth
	if len(passwords.L2TPAuth) != 1 {
		t.Errorf("Expected 1 L2TP auth, got %d", len(passwords.L2TPAuth))
	} else {
		auth := passwords.L2TPAuth[0]
		if auth.TunnelID != 1 {
			t.Errorf("Expected L2TP tunnel ID 1, got %d", auth.TunnelID)
		}
		if auth.Secret != "test-l2tp-auth-secret" {
			t.Errorf("Expected L2TP secret 'test-l2tp-auth-secret', got '%s'", auth.Secret)
		}
	}
}

func TestConfigFileParser_ExtractPasswords_PPAuth(t *testing.T) {
	input := `pp select anonymous
 pp bind tunnel1
 pp auth request mschap-v2
 pp auth username vpnuser test-vpn-password-789
 ppp ipcp ipaddress on
 pp enable anonymous`

	parser := NewConfigFileParser()
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	passwords := result.ExtractPasswords()

	// Check PP auth username/password
	if len(passwords.PPAuth) != 1 {
		t.Errorf("Expected 1 PP auth entry, got %d", len(passwords.PPAuth))
	} else {
		ppAuth := passwords.PPAuth[0]
		if ppAuth.PPName != "anonymous" {
			t.Errorf("Expected PP name 'anonymous', got '%s'", ppAuth.PPName)
		}
		if ppAuth.Username != "vpnuser" {
			t.Errorf("Expected username 'vpnuser', got '%s'", ppAuth.Username)
		}
		if ppAuth.Password != "test-vpn-password-789" {
			t.Errorf("Expected password 'test-vpn-password-789', got '%s'", ppAuth.Password)
		}
	}
}

func TestConfigFileParser_ExtractFromSampleConfig(t *testing.T) {
	// Full sample config from requirements.md
	sampleConfig := `#
# Admin
#
login password test-login-password-123
administrator password test-admin-password-456
login user testuser encrypted TESTENCRYPTEDHASH123456789
user attribute administrator=off connection=off gui-page=dashboard,lan-map,config login-timer=300
user attribute testuser connection=serial,telnet,remote,ssh,sftp,http gui-page=dashboard,lan-map,config login-timer=3600
timezone +09:00
console character ja.utf8
console prompt "[TEST-RTX] "

httpd host any
sshd service on
sshd host lan1
sftpd host lan1

#
# WAN connection
#
description lan2 test-wan
ip lan2 address 198.51.100.1/24
ip lan2 nat descriptor 1000
ip lan2 secure filter in 200020 200099
ip lan2 secure filter out 200099 dynamic 200080 200081

#
# IP configuration
#
ip route default gateway 198.51.100.254
ip route 10.0.0.0/8 gateway 192.0.2.1

#
# LAN configuration
#
ip lan1 address 192.0.2.253/24

#
# Services
#
dhcp service server
dhcp scope 1 192.0.2.100-192.0.2.199/24 gateway 192.0.2.253 expire 12:00
dhcp scope bind 1 192.0.2.100 01:00:11:22:33:44:55
dhcp scope option 1 dns=192.0.2.10

dns host lan1
dns service recursive
dns server select 1 192.0.2.10 edns=on any example.local
dns server select 500000 8.8.8.8 edns=on 8.8.4.4 edns=on any .

#
# Tunnels
#
pp select anonymous
 pp bind tunnel1
 pp auth request mschap-v2
 pp auth username vpnuser test-vpn-password-789
 ppp ipcp ipaddress on
 pp enable anonymous

tunnel select 1
 tunnel encapsulation l2tpv3
 tunnel endpoint name test.example.com fqdn
 ipsec tunnel 101
  ipsec sa policy 101 1 esp aes-cbc sha-hmac
  ipsec ike pre-shared-key 1 text test-ike-psk-secret
  ipsec ike remote address 1 test.example.com
 l2tp tunnel auth on test-l2tp-auth-secret
 tunnel enable 1

#
# Filters
#
ip filter 200020 reject * * udp,tcp 135 *
ip filter 200099 pass * * * * *
ip filter dynamic 200080 * * ftp
ip filter dynamic 200081 * * www

nat descriptor type 1000 masquerade
nat descriptor masquerade static 1000 1 192.0.2.253 tcp 22`

	parser := NewConfigFileParser()
	result, err := parser.Parse(sampleConfig)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Test static routes extraction
	routes := result.ExtractStaticRoutes()
	if len(routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(routes))
	}

	// Test DHCP scopes extraction
	scopes := result.ExtractDHCPScopes()
	if len(scopes) != 1 {
		t.Errorf("Expected 1 DHCP scope, got %d", len(scopes))
	}

	// Test NAT masquerade extraction
	nats := result.ExtractNATMasquerade()
	if len(nats) != 1 {
		t.Errorf("Expected 1 NAT descriptor, got %d", len(nats))
	}

	// Test IP filters extraction
	staticFilters := result.ExtractIPFilters()
	if len(staticFilters) != 2 {
		t.Errorf("Expected 2 static filters, got %d", len(staticFilters))
	}

	dynamicFilters := result.ExtractIPFiltersDynamic()
	if len(dynamicFilters) != 2 {
		t.Errorf("Expected 2 dynamic filters, got %d", len(dynamicFilters))
	}

	// Test passwords extraction
	passwords := result.ExtractPasswords()
	if passwords.LoginPassword != "test-login-password-123" {
		t.Errorf("Expected login password, got '%s'", passwords.LoginPassword)
	}
	if passwords.AdminPassword != "test-admin-password-456" {
		t.Errorf("Expected admin password, got '%s'", passwords.AdminPassword)
	}
	if len(passwords.IPsecPSK) != 1 {
		t.Errorf("Expected 1 IPsec PSK, got %d", len(passwords.IPsecPSK))
	}
	if len(passwords.L2TPAuth) != 1 {
		t.Errorf("Expected 1 L2TP auth, got %d", len(passwords.L2TPAuth))
	}
	if len(passwords.PPAuth) != 1 {
		t.Errorf("Expected 1 PP auth, got %d", len(passwords.PPAuth))
	}
}

func TestConfigFileParser_ExtractIPsecTunnels(t *testing.T) {
	input := `tunnel select 1
 tunnel encapsulation ipsec
 ipsec tunnel 1
  ipsec sa policy 1 1 esp aes-cbc sha-hmac
 description test-ipsec-tunnel
 tunnel enable 1

ipsec ike local address 1 192.168.1.1
ipsec ike remote address 1 203.0.113.1
ipsec ike pre-shared-key 1 text test-psk-secret
ipsec ike encryption 1 aes-cbc-256
ipsec ike hash 1 sha256
ipsec ike group 1 modp2048`

	parser := NewConfigFileParser()
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	tunnels := result.ExtractIPsecTunnels()

	if len(tunnels) == 0 {
		t.Fatal("Expected at least 1 IPsec tunnel, got 0")
	}

	// Check tunnel properties
	tunnel := tunnels[0]
	if tunnel.ID != 1 {
		t.Errorf("Expected tunnel ID 1, got %d", tunnel.ID)
	}
	if tunnel.LocalAddress != "192.168.1.1" {
		t.Errorf("Expected local address 192.168.1.1, got %s", tunnel.LocalAddress)
	}
	if tunnel.RemoteAddress != "203.0.113.1" {
		t.Errorf("Expected remote address 203.0.113.1, got %s", tunnel.RemoteAddress)
	}
	if tunnel.PreSharedKey != "test-psk-secret" {
		t.Errorf("Expected PSK 'test-psk-secret', got '%s'", tunnel.PreSharedKey)
	}
}

func TestConfigFileParser_ExtractL2TPTunnels(t *testing.T) {
	input := `tunnel select 1
 tunnel encapsulation l2tpv3
 tunnel endpoint address 192.168.1.1 203.0.113.1
 l2tp local router-id 192.168.1.1
 l2tp remote router-id 203.0.113.1
 l2tp always-on on
 description test-l2tp-tunnel
 tunnel enable 1`

	parser := NewConfigFileParser()
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	tunnels := result.ExtractL2TPTunnels()

	if len(tunnels) == 0 {
		t.Fatal("Expected at least 1 L2TP tunnel, got 0")
	}

	// Check tunnel properties
	tunnel := tunnels[0]
	if tunnel.ID != 1 {
		t.Errorf("Expected tunnel ID 1, got %d", tunnel.ID)
	}
	if tunnel.Version != "l2tpv3" {
		t.Errorf("Expected version 'l2tpv3', got '%s'", tunnel.Version)
	}
	if tunnel.TunnelSource != "192.168.1.1" {
		t.Errorf("Expected tunnel source 192.168.1.1, got %s", tunnel.TunnelSource)
	}
	if tunnel.TunnelDest != "203.0.113.1" {
		t.Errorf("Expected tunnel dest 203.0.113.1, got %s", tunnel.TunnelDest)
	}
	if !tunnel.AlwaysOn {
		t.Error("Expected always-on to be true")
	}
}

func TestConfigFileParser_ExtractL2TPTunnels_WithAnonymousPP(t *testing.T) {
	input := `l2tp service on

pp select anonymous
 pp bind tunnel1
 pp auth accept mschap-v2
 pp auth myname testuser testpass
 ppp ipcp ipaddress on
 pp enable anonymous

tunnel select 1
 tunnel encapsulation l2tpv3
 tunnel endpoint address 192.168.1.1 203.0.113.1
 tunnel enable 1`

	parser := NewConfigFileParser()
	result, err := parser.Parse(input)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	tunnels := result.ExtractL2TPTunnels()

	// Should extract from both tunnel context and pp anonymous context
	if len(tunnels) == 0 {
		t.Fatal("Expected at least 1 L2TP tunnel, got 0")
	}
}

func TestConfigFileParser_ExtractL2TPService(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectEnabled  bool
		expectNil      bool
		expectProtocol []string
	}{
		{
			name:          "l2tp service on",
			input:         "l2tp service on",
			expectEnabled: true,
			expectNil:     false,
		},
		{
			name:          "l2tp service off",
			input:         "l2tp service off",
			expectEnabled: false,
			expectNil:     false,
		},
		{
			name:           "l2tp service on with protocols",
			input:          "l2tp service on l2tpv3 l2tp",
			expectEnabled:  true,
			expectNil:      false,
			expectProtocol: []string{"l2tpv3", "l2tp"},
		},
		{
			name:      "no l2tp service command",
			input:     "ip route default gateway 192.168.1.1",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewConfigFileParser()
			result, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			service := result.ExtractL2TPService()

			if tt.expectNil {
				if service != nil {
					t.Errorf("Expected nil service, got %+v", service)
				}
				return
			}

			if service == nil {
				t.Fatal("Expected non-nil service, got nil")
			}

			if service.Enabled != tt.expectEnabled {
				t.Errorf("Expected Enabled=%v, got %v", tt.expectEnabled, service.Enabled)
			}

			if tt.expectProtocol != nil {
				if len(service.Protocols) != len(tt.expectProtocol) {
					t.Errorf("Expected %d protocols, got %d", len(tt.expectProtocol), len(service.Protocols))
				} else {
					for i, proto := range tt.expectProtocol {
						if service.Protocols[i] != proto {
							t.Errorf("Expected protocol[%d]=%s, got %s", i, proto, service.Protocols[i])
						}
					}
				}
			}
		})
	}
}
