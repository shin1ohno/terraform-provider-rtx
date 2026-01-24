package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestBuildDNSConfigFromResourceData_BasicConfig(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceRTXDNSServer().Schema, map[string]interface{}{
		"domain_lookup":         true,
		"name_servers":          []interface{}{"8.8.8.8", "8.8.4.4"},
		"service_on":            true,
		"private_address_spoof": false,
	})

	config := buildDNSConfigFromResourceData(resourceData)

	if !config.DomainLookup {
		t.Error("Expected DomainLookup to be true")
	}
	if len(config.NameServers) != 2 {
		t.Errorf("Expected 2 name servers, got %d", len(config.NameServers))
	}
	if config.NameServers[0] != "8.8.8.8" {
		t.Errorf("Expected first server '8.8.8.8', got '%s'", config.NameServers[0])
	}
	if config.NameServers[1] != "8.8.4.4" {
		t.Errorf("Expected second server '8.8.4.4', got '%s'", config.NameServers[1])
	}
	if !config.ServiceOn {
		t.Error("Expected ServiceOn to be true")
	}
	if config.PrivateSpoof {
		t.Error("Expected PrivateSpoof to be false")
	}
}

func TestBuildDNSConfigFromResourceData_WithDomainName(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceRTXDNSServer().Schema, map[string]interface{}{
		"domain_lookup": true,
		"domain_name":   "example.com",
		"name_servers":  []interface{}{"8.8.8.8"},
		"service_on":    false,
	})

	config := buildDNSConfigFromResourceData(resourceData)

	if config.DomainName != "example.com" {
		t.Errorf("Expected domain name 'example.com', got '%s'", config.DomainName)
	}
}

func TestBuildDNSConfigFromResourceData_WithServerSelect(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceRTXDNSServer().Schema, map[string]interface{}{
		"domain_lookup": true,
		"server_select": []interface{}{
			map[string]interface{}{
				"id": 1,
				"server": []interface{}{
					map[string]interface{}{
						"address": "192.168.1.1",
						"edns":    false,
					},
				},
				"record_type":   "a",
				"query_pattern": "internal.example.com",
			},
			map[string]interface{}{
				"id": 2,
				"server": []interface{}{
					map[string]interface{}{
						"address": "10.0.0.1",
						"edns":    true,
					},
					map[string]interface{}{
						"address": "10.0.0.2",
						"edns":    false,
					},
				},
				"record_type":     "any",
				"query_pattern":   ".",
				"original_sender": "192.168.1.0/24",
				"restrict_pp":     1,
			},
		},
		"service_on": true,
	})

	config := buildDNSConfigFromResourceData(resourceData)

	if len(config.ServerSelect) != 2 {
		t.Fatalf("Expected 2 server select entries, got %d", len(config.ServerSelect))
	}

	// Check first entry
	if config.ServerSelect[0].ID != 1 {
		t.Errorf("Expected first select ID 1, got %d", config.ServerSelect[0].ID)
	}
	if len(config.ServerSelect[0].Servers) != 1 {
		t.Errorf("Expected 1 server in first select, got %d", len(config.ServerSelect[0].Servers))
	}
	if config.ServerSelect[0].Servers[0].Address != "192.168.1.1" {
		t.Errorf("Expected server address '192.168.1.1', got '%s'", config.ServerSelect[0].Servers[0].Address)
	}
	if config.ServerSelect[0].Servers[0].EDNS {
		t.Error("Expected EDNS to be false for first server")
	}
	if config.ServerSelect[0].QueryPattern != "internal.example.com" {
		t.Errorf("Expected query pattern 'internal.example.com', got '%s'", config.ServerSelect[0].QueryPattern)
	}

	// Check second entry
	if config.ServerSelect[1].ID != 2 {
		t.Errorf("Expected second select ID 2, got %d", config.ServerSelect[1].ID)
	}
	if len(config.ServerSelect[1].Servers) != 2 {
		t.Errorf("Expected 2 servers in second select, got %d", len(config.ServerSelect[1].Servers))
	}
	// Check per-server EDNS settings
	if !config.ServerSelect[1].Servers[0].EDNS {
		t.Error("Expected EDNS to be true for first server in second select")
	}
	if config.ServerSelect[1].Servers[1].EDNS {
		t.Error("Expected EDNS to be false for second server in second select")
	}
	if config.ServerSelect[1].RecordType != "any" {
		t.Errorf("Expected record type 'any', got '%s'", config.ServerSelect[1].RecordType)
	}
	if config.ServerSelect[1].OriginalSender != "192.168.1.0/24" {
		t.Errorf("Expected original sender '192.168.1.0/24', got '%s'", config.ServerSelect[1].OriginalSender)
	}
	if config.ServerSelect[1].RestrictPP != 1 {
		t.Errorf("Expected restrict_pp 1, got %d", config.ServerSelect[1].RestrictPP)
	}
}

func TestBuildDNSConfigFromResourceData_WithHosts(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceRTXDNSServer().Schema, map[string]interface{}{
		"domain_lookup": true,
		"hosts": []interface{}{
			map[string]interface{}{
				"name":    "router",
				"address": "192.168.1.1",
			},
			map[string]interface{}{
				"name":    "nas",
				"address": "192.168.1.10",
			},
		},
		"service_on": true,
	})

	config := buildDNSConfigFromResourceData(resourceData)

	if len(config.Hosts) != 2 {
		t.Fatalf("Expected 2 hosts, got %d", len(config.Hosts))
	}

	if config.Hosts[0].Name != "router" {
		t.Errorf("Expected first host name 'router', got '%s'", config.Hosts[0].Name)
	}
	if config.Hosts[0].Address != "192.168.1.1" {
		t.Errorf("Expected first host address '192.168.1.1', got '%s'", config.Hosts[0].Address)
	}
	if config.Hosts[1].Name != "nas" {
		t.Errorf("Expected second host name 'nas', got '%s'", config.Hosts[1].Name)
	}
	if config.Hosts[1].Address != "192.168.1.10" {
		t.Errorf("Expected second host address '192.168.1.10', got '%s'", config.Hosts[1].Address)
	}
}

func TestBuildDNSConfigFromResourceData_FullConfig(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceRTXDNSServer().Schema, map[string]interface{}{
		"domain_lookup": false,
		"domain_name":   "example.com",
		"name_servers":  []interface{}{"8.8.8.8", "1.1.1.1"},
		"server_select": []interface{}{
			map[string]interface{}{
				"id": 1,
				"server": []interface{}{
					map[string]interface{}{
						"address": "192.168.1.1",
						"edns":    false,
					},
				},
				"record_type":   "a",
				"query_pattern": "internal.example.com",
			},
		},
		"hosts": []interface{}{
			map[string]interface{}{
				"name":    "router",
				"address": "192.168.1.1",
			},
		},
		"service_on":            true,
		"private_address_spoof": true,
	})

	config := buildDNSConfigFromResourceData(resourceData)

	if config.DomainLookup {
		t.Error("Expected DomainLookup to be false")
	}
	if config.DomainName != "example.com" {
		t.Errorf("Expected domain name 'example.com', got '%s'", config.DomainName)
	}
	if len(config.NameServers) != 2 {
		t.Errorf("Expected 2 name servers, got %d", len(config.NameServers))
	}
	if len(config.ServerSelect) != 1 {
		t.Errorf("Expected 1 server select, got %d", len(config.ServerSelect))
	}
	if len(config.Hosts) != 1 {
		t.Errorf("Expected 1 host, got %d", len(config.Hosts))
	}
	if !config.ServiceOn {
		t.Error("Expected ServiceOn to be true")
	}
	if !config.PrivateSpoof {
		t.Error("Expected PrivateSpoof to be true")
	}
}

func TestBuildDNSConfigFromResourceData_EmptyConfig(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, resourceRTXDNSServer().Schema, map[string]interface{}{
		"domain_lookup": true,
		"service_on":    false,
	})

	config := buildDNSConfigFromResourceData(resourceData)

	if config.DomainName != "" {
		t.Errorf("Expected empty domain name, got '%s'", config.DomainName)
	}
	if len(config.NameServers) != 0 {
		t.Errorf("Expected 0 name servers, got %d", len(config.NameServers))
	}
	if len(config.ServerSelect) != 0 {
		t.Errorf("Expected 0 server select entries, got %d", len(config.ServerSelect))
	}
	if len(config.Hosts) != 0 {
		t.Errorf("Expected 0 hosts, got %d", len(config.Hosts))
	}
}

// Acceptance tests for rtx_dns_server

func TestAccRTXDNSServer_basic(t *testing.T) {
	resourceName := "rtx_dns_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXDNSServerConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "domain_lookup", "true"),
					resource.TestCheckResourceAttr(resourceName, "service_on", "true"),
					resource.TestCheckResourceAttr(resourceName, "name_servers.#", "2"),
				),
			},
		},
	})
}

func TestAccRTXDNSServer_import(t *testing.T) {
	resourceName := "rtx_dns_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXDNSServerConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "domain_lookup", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRTXDNSServer_noDiff(t *testing.T) {
	resourceName := "rtx_dns_server.test"
	config := testAccRTXDNSServerConfig_withDomain()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example.local"),
					resource.TestCheckResourceAttr(resourceName, "service_on", "true"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

func TestAccRTXDNSServer_update(t *testing.T) {
	resourceName := "rtx_dns_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXDNSServerConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name_servers.#", "2"),
				),
			},
			{
				Config: testAccRTXDNSServerConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name_servers.#", "1"),
				),
			},
		},
	})
}

func testAccRTXDNSServerConfig_basic() string {
	return `
resource "rtx_dns_server" "test" {
  domain_lookup = true
  name_servers  = ["8.8.8.8", "8.8.4.4"]
  service_on    = true
}
`
}

func testAccRTXDNSServerConfig_withDomain() string {
	return `
resource "rtx_dns_server" "test" {
  domain_lookup = true
  domain_name   = "example.local"
  name_servers  = ["8.8.8.8"]
  service_on    = true
}
`
}

func testAccRTXDNSServerConfig_updated() string {
	return `
resource "rtx_dns_server" "test" {
  domain_lookup = true
  name_servers  = ["1.1.1.1"]
  service_on    = true
}
`
}
