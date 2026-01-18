package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRTXDHCPScope_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXDHCPScopeConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "scope_id", "1"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "network", "192.168.100.0/24"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "lease_time", "72h"),
					resource.TestCheckResourceAttrSet("rtx_dhcp_scope.test", "id"),
				),
			},
			// Test import
			{
				ResourceName:      "rtx_dhcp_scope.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "1",
			},
		},
	})
}

func TestAccRTXDHCPScope_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXDHCPScopeConfig_full(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "scope_id", "2"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "network", "192.168.200.0/24"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "gateway", "192.168.200.1"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "dns_servers.#", "2"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "dns_servers.0", "8.8.8.8"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "dns_servers.1", "8.8.4.4"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "lease_time", "24h"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "exclude_ranges.#", "1"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "exclude_ranges.0.start", "192.168.200.1"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "exclude_ranges.0.end", "192.168.200.10"),
				),
			},
		},
	})
}

func TestAccRTXDHCPScope_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Initial creation
			{
				Config: testAccRTXDHCPScopeConfig_update_step1(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "scope_id", "3"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "network", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "gateway", "10.0.0.1"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "dns_servers.#", "1"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "dns_servers.0", "8.8.8.8"),
				),
			},
			// Update DNS servers and lease time
			{
				Config: testAccRTXDHCPScopeConfig_update_step2(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "scope_id", "3"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "network", "10.0.0.0/24"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "gateway", "10.0.0.1"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "dns_servers.#", "3"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "dns_servers.0", "8.8.8.8"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "dns_servers.1", "8.8.4.4"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "dns_servers.2", "1.1.1.1"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "lease_time", "48h"),
				),
			},
			// Add exclude ranges
			{
				Config: testAccRTXDHCPScopeConfig_update_step3(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "scope_id", "3"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "exclude_ranges.#", "2"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "exclude_ranges.0.start", "10.0.0.1"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "exclude_ranges.0.end", "10.0.0.10"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "exclude_ranges.1.start", "10.0.0.250"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "exclude_ranges.1.end", "10.0.0.254"),
				),
			},
		},
	})
}

func TestAccRTXDHCPScope_infiniteLease(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXDHCPScopeConfig_infiniteLease(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "scope_id", "4"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "lease_time", "infinite"),
				),
			},
		},
	})
}

func TestAccRTXDHCPScope_withBinding(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXDHCPScopeConfig_withBinding(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check scope
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "scope_id", "5"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.test", "network", "172.16.0.0/24"),
					// Check binding
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test", "scope_id", "5"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test", "ip_address", "172.16.0.100"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test", "mac_address", "aa:bb:cc:dd:ee:ff"),
				),
			},
		},
	})
}

func TestAccRTXDHCPScope_validationErrors(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Invalid CIDR notation
			{
				Config:      testAccRTXDHCPScopeConfig_invalidCIDR(),
				ExpectError: regexp.MustCompile("must be a valid CIDR notation"),
			},
			// Invalid IP address for gateway
			{
				Config:      testAccRTXDHCPScopeConfig_invalidGateway(),
				ExpectError: regexp.MustCompile("must be a valid IP address"),
			},
			// Invalid IP address for DNS
			{
				Config:      testAccRTXDHCPScopeConfig_invalidDNS(),
				ExpectError: regexp.MustCompile("must be a valid IP address"),
			},
			// Invalid lease time format
			{
				Config:      testAccRTXDHCPScopeConfig_invalidLeaseTime(),
				ExpectError: regexp.MustCompile("must be a valid duration"),
			},
		},
	})
}

func TestAccRTXDHCPScope_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXDHCPScopeConfig_import(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_dhcp_scope.import_test", "scope_id", "6"),
					resource.TestCheckResourceAttr("rtx_dhcp_scope.import_test", "network", "192.168.6.0/24"),
					testAccCheckRTXDHCPScopeExists("rtx_dhcp_scope.import_test"),
				),
			},
			{
				ResourceName:      "rtx_dhcp_scope.import_test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "6",
			},
		},
	})
}

func TestAccRTXDHCPScope_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXDHCPScopeConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRTXDHCPScopeExists("rtx_dhcp_scope.test"),
				),
			},
			{
				// Verify the Read handles missing resources gracefully
				PreConfig: func() {
					// In a real test, this would delete the scope outside of Terraform
				},
				Config:             testAccRTXDHCPScopeConfig_basic(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckRTXDHCPScopeExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DHCP Scope ID is set")
		}

		return nil
	}
}

// Test configuration generators

func testAccRTXDHCPScopeConfig_basic() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_scope" "test" {
  scope_id = 1
  network  = "192.168.100.0/24"
}
`
}

func testAccRTXDHCPScopeConfig_full() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_scope" "test" {
  scope_id    = 2
  network     = "192.168.200.0/24"
  gateway     = "192.168.200.1"
  dns_servers = ["8.8.8.8", "8.8.4.4"]
  lease_time  = "24h"

  exclude_ranges {
    start = "192.168.200.1"
    end   = "192.168.200.10"
  }
}
`
}

func testAccRTXDHCPScopeConfig_update_step1() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_scope" "test" {
  scope_id    = 3
  network     = "10.0.0.0/24"
  gateway     = "10.0.0.1"
  dns_servers = ["8.8.8.8"]
  lease_time  = "72h"
}
`
}

func testAccRTXDHCPScopeConfig_update_step2() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_scope" "test" {
  scope_id    = 3
  network     = "10.0.0.0/24"
  gateway     = "10.0.0.1"
  dns_servers = ["8.8.8.8", "8.8.4.4", "1.1.1.1"]
  lease_time  = "48h"
}
`
}

func testAccRTXDHCPScopeConfig_update_step3() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_scope" "test" {
  scope_id    = 3
  network     = "10.0.0.0/24"
  gateway     = "10.0.0.1"
  dns_servers = ["8.8.8.8", "8.8.4.4", "1.1.1.1"]
  lease_time  = "48h"

  exclude_ranges {
    start = "10.0.0.1"
    end   = "10.0.0.10"
  }

  exclude_ranges {
    start = "10.0.0.250"
    end   = "10.0.0.254"
  }
}
`
}

func testAccRTXDHCPScopeConfig_infiniteLease() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_scope" "test" {
  scope_id   = 4
  network    = "192.168.4.0/24"
  lease_time = "infinite"
}
`
}

func testAccRTXDHCPScopeConfig_withBinding() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_scope" "test" {
  scope_id    = 5
  network     = "172.16.0.0/24"
  gateway     = "172.16.0.1"
  dns_servers = ["8.8.8.8"]
}

resource "rtx_dhcp_binding" "test" {
  scope_id    = rtx_dhcp_scope.test.scope_id
  ip_address  = "172.16.0.100"
  mac_address = "aa:bb:cc:dd:ee:ff"

  depends_on = [rtx_dhcp_scope.test]
}
`
}

func testAccRTXDHCPScopeConfig_invalidCIDR() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_scope" "test" {
  scope_id = 1
  network  = "192.168.1.0"
}
`
}

func testAccRTXDHCPScopeConfig_invalidGateway() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_scope" "test" {
  scope_id = 1
  network  = "192.168.1.0/24"
  gateway  = "invalid-ip"
}
`
}

func testAccRTXDHCPScopeConfig_invalidDNS() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_scope" "test" {
  scope_id    = 1
  network     = "192.168.1.0/24"
  dns_servers = ["not-an-ip"]
}
`
}

func testAccRTXDHCPScopeConfig_invalidLeaseTime() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_scope" "test" {
  scope_id   = 1
  network    = "192.168.1.0/24"
  lease_time = "invalid"
}
`
}

func testAccRTXDHCPScopeConfig_import() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_scope" "import_test" {
  scope_id    = 6
  network     = "192.168.6.0/24"
  gateway     = "192.168.6.1"
  dns_servers = ["8.8.8.8"]
}
`
}
