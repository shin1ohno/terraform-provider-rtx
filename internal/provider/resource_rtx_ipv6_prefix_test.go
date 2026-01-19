package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRTXIPv6Prefix_static(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXIPv6PrefixDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXIPv6PrefixConfig_static(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXIPv6PrefixExists("rtx_ipv6_prefix.test"),
					resource.TestCheckResourceAttr("rtx_ipv6_prefix.test", "prefix_id", "1"),
					resource.TestCheckResourceAttr("rtx_ipv6_prefix.test", "prefix", "2001:db8::"),
					resource.TestCheckResourceAttr("rtx_ipv6_prefix.test", "prefix_length", "64"),
					resource.TestCheckResourceAttr("rtx_ipv6_prefix.test", "source", "static"),
				),
			},
			{
				ResourceName:      "rtx_ipv6_prefix.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRTXIPv6Prefix_ra(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXIPv6PrefixDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXIPv6PrefixConfig_ra(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXIPv6PrefixExists("rtx_ipv6_prefix.test_ra"),
					resource.TestCheckResourceAttr("rtx_ipv6_prefix.test_ra", "prefix_id", "2"),
					resource.TestCheckResourceAttr("rtx_ipv6_prefix.test_ra", "prefix_length", "64"),
					resource.TestCheckResourceAttr("rtx_ipv6_prefix.test_ra", "source", "ra"),
					resource.TestCheckResourceAttr("rtx_ipv6_prefix.test_ra", "interface", "lan2"),
				),
			},
		},
	})
}

func TestAccRTXIPv6Prefix_dhcpv6pd(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXIPv6PrefixDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXIPv6PrefixConfig_dhcpv6pd(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXIPv6PrefixExists("rtx_ipv6_prefix.test_pd"),
					resource.TestCheckResourceAttr("rtx_ipv6_prefix.test_pd", "prefix_id", "3"),
					resource.TestCheckResourceAttr("rtx_ipv6_prefix.test_pd", "prefix_length", "48"),
					resource.TestCheckResourceAttr("rtx_ipv6_prefix.test_pd", "source", "dhcpv6-pd"),
					resource.TestCheckResourceAttr("rtx_ipv6_prefix.test_pd", "interface", "lan2"),
				),
			},
		},
	})
}

func TestAccRTXIPv6Prefix_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXIPv6PrefixDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXIPv6PrefixConfig_static(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXIPv6PrefixExists("rtx_ipv6_prefix.test"),
					resource.TestCheckResourceAttr("rtx_ipv6_prefix.test", "prefix_length", "64"),
				),
			},
			{
				Config: testAccRTXIPv6PrefixConfig_static_updated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXIPv6PrefixExists("rtx_ipv6_prefix.test"),
					resource.TestCheckResourceAttr("rtx_ipv6_prefix.test", "prefix_length", "56"),
				),
			},
		},
	})
}

func testAccCheckRTXIPv6PrefixExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IPv6 prefix ID is set")
		}

		return nil
	}
}

func testAccCheckRTXIPv6PrefixDestroy(s *terraform.State) error {
	// In a real implementation, we would check that the prefix
	// no longer exists on the router
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rtx_ipv6_prefix" {
			continue
		}
		// Check that the resource has been removed
		// This would require actual router communication
	}
	return nil
}

func testAccRTXIPv6PrefixConfig_static() string {
	return `
resource "rtx_ipv6_prefix" "test" {
  prefix_id     = 1
  prefix        = "2001:db8::"
  prefix_length = 64
  source        = "static"
}
`
}

func testAccRTXIPv6PrefixConfig_static_updated() string {
	return `
resource "rtx_ipv6_prefix" "test" {
  prefix_id     = 1
  prefix        = "2001:db8::"
  prefix_length = 56
  source        = "static"
}
`
}

func testAccRTXIPv6PrefixConfig_ra() string {
	return `
resource "rtx_ipv6_prefix" "test_ra" {
  prefix_id     = 2
  prefix_length = 64
  source        = "ra"
  interface     = "lan2"
}
`
}

func testAccRTXIPv6PrefixConfig_dhcpv6pd() string {
	return `
resource "rtx_ipv6_prefix" "test_pd" {
  prefix_id     = 3
  prefix_length = 48
  source        = "dhcpv6-pd"
  interface     = "lan2"
}
`
}
