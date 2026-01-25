package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRTXIPv6Interface_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXIPv6InterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXIPv6InterfaceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXIPv6InterfaceExists("rtx_ipv6_interface.test"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "interface", "lan1"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "address.0.address", "2001:db8::1/64"),
				),
			},
		},
	})
}

func TestAccRTXIPv6Interface_rtadv(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXIPv6InterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXIPv6InterfaceConfig_rtadv(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXIPv6InterfaceExists("rtx_ipv6_interface.test"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "interface", "lan1"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "rtadv.0.enabled", "true"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "rtadv.0.prefix_id", "1"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "rtadv.0.o_flag", "true"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "rtadv.0.m_flag", "false"),
				),
			},
		},
	})
}

func TestAccRTXIPv6Interface_dhcpv6Server(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXIPv6InterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXIPv6InterfaceConfig_dhcpv6Server(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXIPv6InterfaceExists("rtx_ipv6_interface.test"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "interface", "lan1"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "dhcpv6_service", "server"),
				),
			},
		},
	})
}

func TestAccRTXIPv6Interface_dhcpv6Client(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXIPv6InterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXIPv6InterfaceConfig_dhcpv6Client(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXIPv6InterfaceExists("rtx_ipv6_interface.test"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "interface", "lan2"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "dhcpv6_service", "client"),
				),
			},
		},
	})
}

func TestAccRTXIPv6Interface_accessLists(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXIPv6InterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXIPv6InterfaceConfig_accessLists(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXIPv6InterfaceExists("rtx_ipv6_interface.test"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "interface", "lan1"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "access_list_ipv6_in", "ipv6-in-acl"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "access_list_ipv6_out", "ipv6-out-acl"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "access_list_ipv6_dynamic_out", "ipv6-dynamic-out-acl"),
				),
			},
		},
	})
}

func TestAccRTXIPv6Interface_prefixBasedAddress(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXIPv6InterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXIPv6InterfaceConfig_prefixBasedAddress(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXIPv6InterfaceExists("rtx_ipv6_interface.test"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "interface", "lan1"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "address.0.prefix_ref", "ra-prefix@lan2"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "address.0.interface_id", "::1/64"),
				),
			},
		},
	})
}

func TestAccRTXIPv6Interface_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXIPv6InterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXIPv6InterfaceConfig_full(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXIPv6InterfaceExists("rtx_ipv6_interface.test"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "interface", "lan1"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "address.0.address", "2001:db8::1/64"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "rtadv.0.enabled", "true"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "rtadv.0.prefix_id", "1"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "dhcpv6_service", "server"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "mtu", "1500"),
				),
			},
		},
	})
}

func TestAccRTXIPv6Interface_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXIPv6InterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXIPv6InterfaceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXIPv6InterfaceExists("rtx_ipv6_interface.test"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "address.0.address", "2001:db8::1/64"),
				),
			},
			{
				Config: testAccRTXIPv6InterfaceConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXIPv6InterfaceExists("rtx_ipv6_interface.test"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "address.0.address", "2001:db8::2/64"),
					resource.TestCheckResourceAttr("rtx_ipv6_interface.test", "mtu", "1400"),
				),
			},
		},
	})
}

func TestAccRTXIPv6Interface_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXIPv6InterfaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXIPv6InterfaceConfig_basic(),
			},
			{
				ResourceName:      "rtx_ipv6_interface.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckRTXIPv6InterfaceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		return nil
	}
}

func testAccCheckRTXIPv6InterfaceDestroy(s *terraform.State) error {
	// In actual tests, we would verify that the IPv6 interface configuration
	// has been removed from the router
	return nil
}

func testAccRTXIPv6InterfaceConfig_basic() string {
	return `
resource "rtx_ipv6_interface" "test" {
  interface = "lan1"

  address {
    address = "2001:db8::1/64"
  }
}
`
}

func testAccRTXIPv6InterfaceConfig_rtadv() string {
	return `
resource "rtx_ipv6_prefix" "test" {
  id            = 1
  prefix        = "2001:db8::"
  prefix_length = 64
  source        = "static"
}

resource "rtx_ipv6_interface" "test" {
  interface = "lan1"

  address {
    address = "2001:db8::1/64"
  }

  rtadv {
    enabled   = true
    prefix_id = rtx_ipv6_prefix.test.id
    o_flag    = true
    m_flag    = false
  }
}
`
}

func testAccRTXIPv6InterfaceConfig_dhcpv6Server() string {
	return `
resource "rtx_ipv6_interface" "test" {
  interface = "lan1"

  address {
    address = "2001:db8::1/64"
  }

  dhcpv6_service = "server"
}
`
}

func testAccRTXIPv6InterfaceConfig_dhcpv6Client() string {
	return `
resource "rtx_ipv6_interface" "test" {
  interface = "lan2"

  dhcpv6_service = "client"
}
`
}

func testAccRTXIPv6InterfaceConfig_accessLists() string {
	return `
resource "rtx_ipv6_interface" "test" {
  interface = "lan1"

  address {
    address = "2001:db8::1/64"
  }

  access_list_ipv6_in          = "ipv6-in-acl"
  access_list_ipv6_out         = "ipv6-out-acl"
  access_list_ipv6_dynamic_out = "ipv6-dynamic-out-acl"
}
`
}

func testAccRTXIPv6InterfaceConfig_prefixBasedAddress() string {
	return `
resource "rtx_ipv6_interface" "test" {
  interface = "lan1"

  address {
    prefix_ref   = "ra-prefix@lan2"
    interface_id = "::1/64"
  }
}
`
}

func testAccRTXIPv6InterfaceConfig_full() string {
	return `
resource "rtx_ipv6_prefix" "test" {
  id            = 1
  prefix        = "2001:db8::"
  prefix_length = 64
  source        = "static"
}

resource "rtx_ipv6_interface" "test" {
  interface = "lan1"

  address {
    address = "2001:db8::1/64"
  }

  rtadv {
    enabled   = true
    prefix_id = rtx_ipv6_prefix.test.id
    o_flag    = true
    m_flag    = true
    lifetime  = 1800
  }

  dhcpv6_service = "server"
  mtu            = 1500

  access_list_ipv6_in  = "ipv6-in-acl"
  access_list_ipv6_out = "ipv6-out-acl"
}
`
}

func testAccRTXIPv6InterfaceConfig_updated() string {
	return `
resource "rtx_ipv6_interface" "test" {
  interface = "lan1"

  address {
    address = "2001:db8::2/64"
  }

  mtu = 1400
}
`
}
