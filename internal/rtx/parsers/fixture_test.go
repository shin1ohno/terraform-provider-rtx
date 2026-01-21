package parsers

import (
	"os"
	"path/filepath"
	"testing"
)

const testdataDir = "../testdata/import_fidelity"

func TestFixture_DNSServerSelectMultiServer(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(testdataDir, "dns_server_select_multi_server.txt"))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	parser := NewDNSParser()
	config, err := parser.ParseDNSConfig(string(data))
	if err != nil {
		t.Fatalf("Failed to parse DNS config: %v", err)
	}

	// REQ-1: Verify multiple server select entries are parsed
	if len(config.ServerSelect) < 5 {
		t.Errorf("Expected at least 5 server select entries, got %d", len(config.ServerSelect))
	}

	// Check entry 500000: Two servers, query pattern "."
	found500000 := false
	for _, sel := range config.ServerSelect {
		if sel.ID == 500000 {
			found500000 = true
			if len(sel.Servers) != 2 {
				t.Errorf("Entry 500000: expected 2 servers, got %d", len(sel.Servers))
			}
			if sel.Servers[0] != "1.1.1.1" || sel.Servers[1] != "1.0.0.1" {
				t.Errorf("Entry 500000: unexpected servers %v", sel.Servers)
			}
			if sel.QueryPattern != "." {
				t.Errorf("Entry 500000: expected query pattern '.', got %q", sel.QueryPattern)
			}
		}
	}
	if !found500000 {
		t.Error("Entry 500000 not found")
	}

	// Check entry 500100: IPv6 server, EDNS, AAAA record type
	found500100 := false
	for _, sel := range config.ServerSelect {
		if sel.ID == 500100 {
			found500100 = true
			if !sel.EDNS {
				t.Error("Entry 500100: EDNS should be enabled")
			}
			if sel.RecordType != "aaaa" {
				t.Errorf("Entry 500100: expected record type 'aaaa', got %q", sel.RecordType)
			}
		}
	}
	if !found500100 {
		t.Error("Entry 500100 not found")
	}

	// Check entry 500200: EDNS, type A, domain pattern, original sender
	found500200 := false
	for _, sel := range config.ServerSelect {
		if sel.ID == 500200 {
			found500200 = true
			if len(sel.Servers) != 2 {
				t.Errorf("Entry 500200: expected 2 servers, got %d", len(sel.Servers))
			}
			if !sel.EDNS {
				t.Error("Entry 500200: EDNS should be enabled")
			}
			if sel.RecordType != "a" {
				t.Errorf("Entry 500200: expected record type 'a', got %q", sel.RecordType)
			}
			if sel.QueryPattern != "*.google.com" {
				t.Errorf("Entry 500200: expected query pattern '*.google.com', got %q", sel.QueryPattern)
			}
			if sel.OriginalSender != "192.168.1.0/24" {
				t.Errorf("Entry 500200: expected original sender '192.168.1.0/24', got %q", sel.OriginalSender)
			}
		}
	}
	if !found500200 {
		t.Error("Entry 500200 not found")
	}

	// Verify other DNS settings
	if !config.ServiceOn {
		t.Error("DNS service should be on")
	}
	if !config.DomainLookup {
		t.Error("DNS domain lookup should be on")
	}
	if config.DomainName != "example.local" {
		t.Errorf("Expected domain name 'example.local', got %q", config.DomainName)
	}
}

func TestFixture_InterfaceFilterLongList(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(testdataDir, "interface_filter_long_list.txt"))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	// REQ-2: Test LAN1 with 13+ filter IDs
	config, err := ParseInterfaceConfig(string(data), "lan1")
	if err != nil {
		t.Fatalf("Failed to parse interface config: %v", err)
	}

	// Check inbound filters (should have 13 IDs)
	if len(config.SecureFilterIn) != 13 {
		t.Errorf("LAN1: expected 13 inbound filters, got %d: %v", len(config.SecureFilterIn), config.SecureFilterIn)
	}

	// Check outbound static filters (11 IDs)
	if len(config.SecureFilterOut) < 7 {
		t.Errorf("LAN1: expected at least 7 outbound static filters, got %d", len(config.SecureFilterOut))
	}

	// Check dynamic filters (3 IDs)
	if len(config.DynamicFilterOut) != 3 {
		t.Errorf("LAN1: expected 3 dynamic filters, got %d: %v", len(config.DynamicFilterOut), config.DynamicFilterOut)
	}

	// Test LAN2 with even longer list
	config2, err := ParseInterfaceConfig(string(data), "lan2")
	if err != nil {
		t.Fatalf("Failed to parse LAN2 config: %v", err)
	}

	if len(config2.SecureFilterIn) != 15 {
		t.Errorf("LAN2: expected 15 inbound filters, got %d", len(config2.SecureFilterIn))
	}

	if len(config2.DynamicFilterOut) != 5 {
		t.Errorf("LAN2: expected 5 dynamic filters, got %d", len(config2.DynamicFilterOut))
	}

	// Test LAN3 with 22 inbound filters
	config3, err := ParseInterfaceConfig(string(data), "lan3")
	if err != nil {
		t.Fatalf("Failed to parse LAN3 config: %v", err)
	}

	if len(config3.SecureFilterIn) != 22 {
		t.Errorf("LAN3: expected 22 inbound filters, got %d", len(config3.SecureFilterIn))
	}
}

func TestFixture_StaticRouteMultiGateway(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(testdataDir, "static_route_multi_gateway.txt"))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	parser := NewStaticRouteParser()
	routes, err := parser.ParseRouteConfig(string(data))
	if err != nil {
		t.Fatalf("Failed to parse route config: %v", err)
	}

	// REQ-3: Verify routes are parsed
	if len(routes) < 5 {
		t.Errorf("Expected at least 5 unique routes, got %d", len(routes))
	}

	// Find the 10.33.128.0/21 route and verify multiple gateways
	var multiGwRoute *StaticRoute
	for i, route := range routes {
		if route.Prefix == "10.33.128.0" && route.Mask == "255.255.248.0" {
			multiGwRoute = &routes[i]
			break
		}
	}

	if multiGwRoute == nil {
		t.Fatal("Route 10.33.128.0/21 not found")
	}

	if len(multiGwRoute.NextHops) != 3 {
		t.Errorf("Expected 3 next hops for 10.33.128.0/21, got %d", len(multiGwRoute.NextHops))
	}

	// Check weight values
	weights := make(map[string]int)
	for _, hop := range multiGwRoute.NextHops {
		weights[hop.NextHop] = hop.Distance
	}

	if weights["192.168.1.20"] != 1 {
		t.Errorf("Gateway 192.168.1.20 should have weight 1, got %d", weights["192.168.1.20"])
	}
	if weights["192.168.1.21"] != 10 {
		t.Errorf("Gateway 192.168.1.21 should have weight 10, got %d", weights["192.168.1.21"])
	}
	if weights["192.168.1.22"] != 20 {
		t.Errorf("Gateway 192.168.1.22 should have weight 20, got %d", weights["192.168.1.22"])
	}

	// Find route with filter attribute
	var filterRoute *StaticRoute
	for i, route := range routes {
		if route.Prefix == "172.16.0.0" && route.Mask == "255.240.0.0" {
			filterRoute = &routes[i]
			break
		}
	}

	if filterRoute == nil {
		t.Fatal("Route 172.16.0.0/12 not found")
	}

	if len(filterRoute.NextHops) != 2 {
		t.Errorf("Expected 2 next hops for 172.16.0.0/12, got %d", len(filterRoute.NextHops))
	}

	// Check filter values
	for _, hop := range filterRoute.NextHops {
		if hop.NextHop == "192.168.1.1" && hop.Filter != 1 {
			t.Errorf("Gateway 192.168.1.1 should have filter 1, got %d", hop.Filter)
		}
		if hop.NextHop == "192.168.1.2" && hop.Filter != 2 {
			t.Errorf("Gateway 192.168.1.2 should have filter 2, got %d", hop.Filter)
		}
	}
}

func TestFixture_L2TPTunnelAuth(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(testdataDir, "l2tp_tunnel_auth.txt"))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	parser := NewL2TPParser()
	tunnels, err := parser.ParseL2TPConfig(string(data))
	if err != nil {
		t.Fatalf("Failed to parse L2TP config: %v", err)
	}

	// REQ-4: Verify tunnels are parsed
	if len(tunnels) < 4 {
		t.Errorf("Expected at least 4 tunnels, got %d", len(tunnels))
	}

	// Find tunnel 1 with auth on and password
	var tunnel1 *L2TPConfig
	for i, tunnel := range tunnels {
		if tunnel.ID == 1 {
			tunnel1 = &tunnels[i]
			break
		}
	}

	if tunnel1 == nil {
		t.Fatal("Tunnel 1 not found")
	}

	if tunnel1.Version != "l2tpv3" {
		t.Errorf("Tunnel 1: expected version 'l2tpv3', got %q", tunnel1.Version)
	}

	if tunnel1.L2TPv3Config == nil {
		t.Fatal("Tunnel 1: L2TPv3Config is nil")
	}

	if tunnel1.L2TPv3Config.TunnelAuth == nil {
		t.Fatal("Tunnel 1: TunnelAuth is nil")
	}

	if !tunnel1.L2TPv3Config.TunnelAuth.Enabled {
		t.Error("Tunnel 1: TunnelAuth should be enabled")
	}

	if tunnel1.L2TPv3Config.TunnelAuth.Password != "TEST_PASSWORD_2word" {
		t.Errorf("Tunnel 1: expected password 'TEST_PASSWORD_2word', got %q", tunnel1.L2TPv3Config.TunnelAuth.Password)
	}

	// Find tunnel 2 with auth on but no password
	var tunnel2 *L2TPConfig
	for i, tunnel := range tunnels {
		if tunnel.ID == 2 {
			tunnel2 = &tunnels[i]
			break
		}
	}

	if tunnel2 != nil && tunnel2.L2TPv3Config != nil && tunnel2.L2TPv3Config.TunnelAuth != nil {
		if !tunnel2.L2TPv3Config.TunnelAuth.Enabled {
			t.Error("Tunnel 2: TunnelAuth should be enabled")
		}
	}

	// Find tunnel 4 with IPsec
	var tunnel4 *L2TPConfig
	for i, tunnel := range tunnels {
		if tunnel.ID == 4 {
			tunnel4 = &tunnels[i]
			break
		}
	}

	if tunnel4 != nil {
		if tunnel4.IPsecProfile == nil {
			t.Error("Tunnel 4: IPsec profile should be set")
		} else if tunnel4.IPsecProfile.TunnelID != 4 {
			t.Errorf("Tunnel 4: expected IPsec tunnel ID 4, got %d", tunnel4.IPsecProfile.TunnelID)
		}
	}
}

func TestFixture_AdminUserFullAttributes(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(testdataDir, "admin_user_full_attributes.txt"))
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	parser := NewAdminParser()
	config, err := parser.ParseAdminConfig(string(data))
	if err != nil {
		t.Fatalf("Failed to parse admin config: %v", err)
	}

	// REQ-5: Verify users are parsed
	if len(config.Users) < 5 {
		t.Errorf("Expected at least 5 users, got %d", len(config.Users))
	}

	// Find admin user and verify all attributes
	var adminUser *UserConfig
	for i, user := range config.Users {
		if user.Username == "admin" {
			adminUser = &config.Users[i]
			break
		}
	}

	if adminUser == nil {
		t.Fatal("Admin user not found")
	}

	if !adminUser.Attributes.Administrator {
		t.Error("Admin user should have administrator=on")
	}

	expectedConnections := []string{"serial", "telnet", "ssh"}
	if len(adminUser.Attributes.Connection) != len(expectedConnections) {
		t.Errorf("Admin: expected %d connections, got %d", len(expectedConnections), len(adminUser.Attributes.Connection))
	}

	expectedGUIPages := []string{"dashboard", "lan-map", "config"}
	if len(adminUser.Attributes.GUIPages) != len(expectedGUIPages) {
		t.Errorf("Admin: expected %d GUI pages, got %d", len(expectedGUIPages), len(adminUser.Attributes.GUIPages))
	}

	if adminUser.Attributes.LoginTimer != 3600 {
		t.Errorf("Admin: expected login-timer=3600, got %d", adminUser.Attributes.LoginTimer)
	}

	// Find operator user with encrypted password
	var operatorUser *UserConfig
	for i, user := range config.Users {
		if user.Username == "operator" {
			operatorUser = &config.Users[i]
			break
		}
	}

	if operatorUser == nil {
		t.Fatal("Operator user not found")
	}

	if !operatorUser.Encrypted {
		t.Error("Operator user should have encrypted password")
	}

	if operatorUser.Attributes.Administrator {
		t.Error("Operator user should have administrator=off")
	}

	// Find fullaccess user with all connection types
	var fullUser *UserConfig
	for i, user := range config.Users {
		if user.Username == "fullaccess" {
			fullUser = &config.Users[i]
			break
		}
	}

	if fullUser == nil {
		t.Fatal("Fullaccess user not found")
	}

	if len(fullUser.Attributes.Connection) != 6 {
		t.Errorf("Fullaccess: expected 6 connection types, got %d: %v",
			len(fullUser.Attributes.Connection), fullUser.Attributes.Connection)
	}

	if fullUser.Attributes.LoginTimer != 0 {
		t.Errorf("Fullaccess: expected login-timer=0, got %d", fullUser.Attributes.LoginTimer)
	}
}
