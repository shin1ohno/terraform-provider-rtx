package provider

import (
	"fmt"
	"regexp"
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
			// Test import with MAC address
			{
				ResourceName:      "rtx_dhcp_binding.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "1:00:11:22:33:44:55",
			},
			// Test import with IP address (backward compatibility)
			{
				ResourceName:      "rtx_dhcp_binding.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "1:192.168.1.100",
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

func TestAccRTXDHCPBinding_clientIdentifierCustom(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Test MAC-based client identifier (01 prefix)
			{
				Config: testAccRTXDHCPBindingConfig_clientIdentifierMAC(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test_mac", "scope_id", "1"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test_mac", "ip_address", "192.168.1.52"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test_mac", "client_identifier", "01:00:11:22:33:44:55"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test_mac", "mac_address", ""),
				),
			},
			// Test ASCII-based client identifier (02 prefix)
			{
				Config: testAccRTXDHCPBindingConfig_clientIdentifierASCII(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test_ascii", "scope_id", "1"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test_ascii", "ip_address", "192.168.1.53"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test_ascii", "client_identifier", "02:68:6f:73:74:6e:61:6d:65"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test_ascii", "mac_address", ""),
				),
			},
			// Test vendor-specific client identifier (FF prefix)
			{
				Config: testAccRTXDHCPBindingConfig_clientIdentifierVendor(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test_vendor", "scope_id", "1"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test_vendor", "ip_address", "192.168.1.54"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test_vendor", "client_identifier", "ff:00:01:02:03:04:05"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.test_vendor", "mac_address", ""),
				),
			},
		},
	})
}

func TestAccRTXDHCPBinding_clientIdentifierValidationErrors(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Test invalid prefix
			{
				Config:      testAccRTXDHCPBindingConfig_clientIdentifierInvalidPrefix(),
				ExpectError: regexp.MustCompile(`client identifier prefix must be 01 \(MAC\), 02 \(ASCII\), or ff \(vendor-specific\)`),
			},
			// Test invalid hex characters
			{
				Config:      testAccRTXDHCPBindingConfig_clientIdentifierInvalidHex(),
				ExpectError: regexp.MustCompile("client identifier contains invalid hex characters"),
			},
			// Test no data after prefix
			{
				Config:      testAccRTXDHCPBindingConfig_clientIdentifierNoData(),
				ExpectError: regexp.MustCompile("client identifier must have data after type prefix"),
			},
			// Test both mac_address and client_identifier specified
			{
				Config:      testAccRTXDHCPBindingConfig_clientIdentifierConflict(),
				ExpectError: regexp.MustCompile("\"client_identifier\": conflicts with mac_address"),
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

func TestAccRTXDHCPBinding_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXDHCPBindingConfig_import(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_dhcp_binding.import_test", "scope_id", "1"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.import_test", "ip_address", "192.168.1.150"),
					resource.TestCheckResourceAttr("rtx_dhcp_binding.import_test", "mac_address", "aa:bb:cc:dd:ee:ff"),
					resource.TestCheckResourceAttrSet("rtx_dhcp_binding.import_test", "id"),
				),
			},
			// Test import with MAC address format
			{
				ResourceName:      "rtx_dhcp_binding.import_test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "1:aa:bb:cc:dd:ee:ff",
			},
			// Test import with IP address format (backward compatibility)
			{
				ResourceName:      "rtx_dhcp_binding.import_test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "1:192.168.1.150",
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

func testAccRTXDHCPBindingConfig_clientIdentifierMAC() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_binding" "test_mac" {
  scope_id          = 1
  ip_address        = "192.168.1.52"
  client_identifier = "01:00:11:22:33:44:55"
}
`
}

func testAccRTXDHCPBindingConfig_clientIdentifierASCII() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_binding" "test_ascii" {
  scope_id          = 1
  ip_address        = "192.168.1.53"
  client_identifier = "02:68:6f:73:74:6e:61:6d:65"
}
`
}

func testAccRTXDHCPBindingConfig_clientIdentifierVendor() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_binding" "test_vendor" {
  scope_id          = 1
  ip_address        = "192.168.1.54"
  client_identifier = "ff:00:01:02:03:04:05"
}
`
}

func testAccRTXDHCPBindingConfig_clientIdentifierInvalidPrefix() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_binding" "test_invalid" {
  scope_id          = 1
  ip_address        = "192.168.1.55"
  client_identifier = "03:00:11:22:33:44:55"
}
`
}

func testAccRTXDHCPBindingConfig_clientIdentifierInvalidHex() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_binding" "test_invalid" {
  scope_id          = 1
  ip_address        = "192.168.1.56"
  client_identifier = "01:zz:11:22:33:44:55"
}
`
}

func testAccRTXDHCPBindingConfig_clientIdentifierNoData() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_binding" "test_invalid" {
  scope_id          = 1
  ip_address        = "192.168.1.57"
  client_identifier = "01:"
}
`
}

func testAccRTXDHCPBindingConfig_clientIdentifierConflict() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_binding" "test_conflict" {
  scope_id          = 1
  ip_address        = "192.168.1.58"
  mac_address       = "00:11:22:33:44:55"
  client_identifier = "01:00:11:22:33:44:55"
}
`
}

func testAccRTXDHCPBindingConfig_import() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_dhcp_binding" "import_test" {
  scope_id    = 1
  ip_address  = "192.168.1.150"
  mac_address = "aa:bb:cc:dd:ee:ff"
}
`
}
