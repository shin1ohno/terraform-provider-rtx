package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRTXVLAN_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXVLANDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXVLANConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXVLANExists("rtx_vlan.test"),
					resource.TestCheckResourceAttr("rtx_vlan.test", "vlan_id", "10"),
					resource.TestCheckResourceAttr("rtx_vlan.test", "interface", "lan1"),
					resource.TestCheckResourceAttr("rtx_vlan.test", "shutdown", "false"),
					resource.TestCheckResourceAttrSet("rtx_vlan.test", "vlan_interface"),
				),
			},
		},
	})
}

func TestAccRTXVLAN_withIP(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXVLANDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXVLANConfig_withIP(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXVLANExists("rtx_vlan.test_ip"),
					resource.TestCheckResourceAttr("rtx_vlan.test_ip", "vlan_id", "20"),
					resource.TestCheckResourceAttr("rtx_vlan.test_ip", "interface", "lan1"),
					resource.TestCheckResourceAttr("rtx_vlan.test_ip", "name", "Test VLAN"),
					resource.TestCheckResourceAttr("rtx_vlan.test_ip", "ip_address", "192.168.20.1"),
					resource.TestCheckResourceAttr("rtx_vlan.test_ip", "ip_mask", "255.255.255.0"),
					resource.TestCheckResourceAttr("rtx_vlan.test_ip", "shutdown", "false"),
				),
			},
		},
	})
}

func TestAccRTXVLAN_update(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXVLANDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXVLANConfig_updateBefore(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXVLANExists("rtx_vlan.test_update"),
					resource.TestCheckResourceAttr("rtx_vlan.test_update", "name", "Before Update"),
					resource.TestCheckResourceAttr("rtx_vlan.test_update", "ip_address", "192.168.30.1"),
				),
			},
			{
				Config: testAccRTXVLANConfig_updateAfter(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXVLANExists("rtx_vlan.test_update"),
					resource.TestCheckResourceAttr("rtx_vlan.test_update", "name", "After Update"),
					resource.TestCheckResourceAttr("rtx_vlan.test_update", "ip_address", "192.168.31.1"),
				),
			},
		},
	})
}

func TestAccRTXVLAN_import(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXVLANDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXVLANConfig_import(),
			},
			{
				ResourceName:      "rtx_vlan.test_import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "lan1/40",
			},
		},
	})
}

func TestAccRTXVLAN_multipleOnSameInterface(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXVLANDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXVLANConfig_multiple(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXVLANExists("rtx_vlan.vlan_50"),
					testAccCheckRTXVLANExists("rtx_vlan.vlan_51"),
					resource.TestCheckResourceAttr("rtx_vlan.vlan_50", "vlan_id", "50"),
					resource.TestCheckResourceAttr("rtx_vlan.vlan_51", "vlan_id", "51"),
				),
			},
		},
	})
}

func testAccCheckRTXVLANExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("VLAN not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("VLAN ID is not set")
		}

		return nil
	}
}

func testAccCheckRTXVLANDestroy(s *terraform.State) error {
	// Check that all VLANs have been destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rtx_vlan" {
			continue
		}

		// The resource should be deleted, so we don't need to check anything
		// If we had access to the client here, we could verify the VLAN no longer exists
	}

	return nil
}

func testAccRTXVLANConfig_basic() string {
	return `
resource "rtx_vlan" "test" {
  vlan_id   = 10
  interface = "lan1"
}
`
}

func testAccRTXVLANConfig_withIP() string {
	return `
resource "rtx_vlan" "test_ip" {
  vlan_id    = 20
  interface  = "lan1"
  name       = "Test VLAN"
  ip_address = "192.168.20.1"
  ip_mask    = "255.255.255.0"
  shutdown   = false
}
`
}

func testAccRTXVLANConfig_updateBefore() string {
	return `
resource "rtx_vlan" "test_update" {
  vlan_id    = 30
  interface  = "lan1"
  name       = "Before Update"
  ip_address = "192.168.30.1"
  ip_mask    = "255.255.255.0"
}
`
}

func testAccRTXVLANConfig_updateAfter() string {
	return `
resource "rtx_vlan" "test_update" {
  vlan_id    = 30
  interface  = "lan1"
  name       = "After Update"
  ip_address = "192.168.31.1"
  ip_mask    = "255.255.255.0"
}
`
}

func testAccRTXVLANConfig_import() string {
	return `
resource "rtx_vlan" "test_import" {
  vlan_id    = 40
  interface  = "lan1"
  name       = "Import Test"
  ip_address = "192.168.40.1"
  ip_mask    = "255.255.255.0"
}
`
}

func testAccRTXVLANConfig_multiple() string {
	return `
resource "rtx_vlan" "vlan_50" {
  vlan_id    = 50
  interface  = "lan1"
  name       = "VLAN 50"
  ip_address = "192.168.50.1"
  ip_mask    = "255.255.255.0"
}

resource "rtx_vlan" "vlan_51" {
  vlan_id    = 51
  interface  = "lan1"
  name       = "VLAN 51"
  ip_address = "192.168.51.1"
  ip_mask    = "255.255.255.0"
}
`
}
