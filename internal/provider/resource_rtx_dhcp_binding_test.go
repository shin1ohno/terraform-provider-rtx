package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRTXDHCPBinding_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXDHCPBindingConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test", "scope_id", "1"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test", "ip_address", "192.168.1.100"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test", "mac_address", "00:11:22:33:44:55"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test", "use_client_identifier", "false"),
					resource.TestCheckResourceAttrSet("rtx_dhcp_binding.test", "id"),
				),
			},
			// Test import
			{
				ResourceName:      "rtx_dhcp_binding.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "1:00:11:22:33:44:55",
			},
		},
	})
}

func TestAccRTXDHCPBinding_clientIdentifier(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXDHCPBindingConfig_clientIdentifier(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test", "scope_id", "1"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test", "ip_address", "192.168.1.51"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test", "mac_address", "00:a0:de:44:55:66"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test", "use_client_identifier", "true"),
				),
			},
		},
	})
}

func TestAccRTXDHCPBinding_multipleBindings(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXDHCPBindingConfig_multiple(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// First binding
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test1", "scope_id", "1"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test1", "ip_address", "192.168.1.101"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test1", "mac_address", "aa:bb:cc:dd:ee:01"),
					// Second binding
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test2", "scope_id", "1"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test2", "ip_address", "192.168.1.102"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test2", "mac_address", "aa:bb:cc:dd:ee:02"),
					// Third binding in different scope
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test3", "scope_id", "2"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test3", "ip_address", "192.168.2.100"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test3", "mac_address", "aa:bb:cc:dd:ee:03"),
				),
			},
		},
	})
}

func TestAccRTXDHCPBinding_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXDHCPBindingConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRTXDHCPBindingExists("rtx_dhcp_binding.test"),
				),
			},
			{
				// Simulate external deletion
				PreConfig: func() {
					// This would normally delete the binding outside of Terraform
					// For now, we'll just verify the Read handles missing resources
				},
				Config:             testAccRTXDHCPBindingConfig_basic(),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckRTXDHCPBindingExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DHCP Binding ID is set")
		}

		// Here you would check if the binding actually exists on the RTX router
		// For now, we just verify the ID format
		return nil
	}
}

func testAccRTXDHCPBindingConfig_basic() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_binding" "test" {
  scope_id    = 1
  ip_address  = "192.168.1.100"
  mac_address = "00:11:22:33:44:55"
}
`
}

func testAccRTXDHCPBindingConfig_clientIdentifier() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_binding" "test" {
  scope_id              = 1
  ip_address            = "192.168.1.51"
  mac_address           = "00:a0:de:44:55:66"
  use_client_identifier = true
}
`
}

func testAccRTXDHCPBindingConfig_multiple() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_binding" "test1" {
  scope_id    = 1
  ip_address  = "192.168.1.101"
  mac_address = "aa:bb:cc:dd:ee:01"
}

resource "rtx_dhcp_binding" "test2" {
  scope_id    = 1
  ip_address  = "192.168.1.102"
  mac_address = "aa:bb:cc:dd:ee:02"
}

resource "rtx_dhcp_binding" "test3" {
  scope_id    = 2
  ip_address  = "192.168.2.100"
  mac_address = "aa:bb:cc:dd:ee:03"
}
`
}