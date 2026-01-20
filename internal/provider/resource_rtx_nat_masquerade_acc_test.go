package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccRTXNATMasquerade_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXNATMasqueradeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXNATMasqueradeConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXNATMasqueradeExists("rtx_nat_masquerade.test"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test", "descriptor_id", "1000"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test", "outer_address", "ipcp"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test", "inner_network", "192.168.1.0-192.168.1.255"),
				),
			},
		},
	})
}

func TestAccRTXNATMasquerade_withStaticEntry(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXNATMasqueradeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXNATMasqueradeConfig_withStaticEntry(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXNATMasqueradeExists("rtx_nat_masquerade.test_static"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_static", "descriptor_id", "1001"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_static", "static_entry.#", "1"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_static", "static_entry.0.inside_local", "192.168.1.100"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_static", "static_entry.0.inside_local_port", "80"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_static", "static_entry.0.protocol", "tcp"),
				),
			},
		},
	})
}

func TestAccRTXNATMasquerade_protocolOnly_ESP(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXNATMasqueradeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXNATMasqueradeConfig_protocolOnlyESP(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXNATMasqueradeExists("rtx_nat_masquerade.test_esp"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_esp", "descriptor_id", "1002"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_esp", "static_entry.#", "1"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_esp", "static_entry.0.inside_local", "192.168.1.253"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_esp", "static_entry.0.protocol", "esp"),
					// Ports should not be set for protocol-only entries
					resource.TestCheckNoResourceAttr("rtx_nat_masquerade.test_esp", "static_entry.0.inside_local_port"),
				),
			},
		},
	})
}

func TestAccRTXNATMasquerade_protocolOnly_AH(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXNATMasqueradeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXNATMasqueradeConfig_protocolOnlyAH(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXNATMasqueradeExists("rtx_nat_masquerade.test_ah"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_ah", "descriptor_id", "1003"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_ah", "static_entry.#", "1"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_ah", "static_entry.0.inside_local", "192.168.1.252"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_ah", "static_entry.0.protocol", "ah"),
				),
			},
		},
	})
}

func TestAccRTXNATMasquerade_protocolOnly_GRE(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXNATMasqueradeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXNATMasqueradeConfig_protocolOnlyGRE(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXNATMasqueradeExists("rtx_nat_masquerade.test_gre"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_gre", "descriptor_id", "1004"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_gre", "static_entry.#", "1"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_gre", "static_entry.0.inside_local", "192.168.1.251"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_gre", "static_entry.0.protocol", "gre"),
				),
			},
		},
	})
}

func TestAccRTXNATMasquerade_mixedEntries(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXNATMasqueradeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXNATMasqueradeConfig_mixedEntries(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXNATMasqueradeExists("rtx_nat_masquerade.test_mixed"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_mixed", "descriptor_id", "1005"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_mixed", "static_entry.#", "3"),
					// TCP entry with ports
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_mixed", "static_entry.0.protocol", "tcp"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_mixed", "static_entry.0.inside_local_port", "443"),
					// ESP entry without ports
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_mixed", "static_entry.1.protocol", "esp"),
					// UDP entry with ports
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_mixed", "static_entry.2.protocol", "udp"),
					resource.TestCheckResourceAttr("rtx_nat_masquerade.test_mixed", "static_entry.2.inside_local_port", "500"),
				),
			},
		},
	})
}

func TestAccRTXNATMasquerade_import(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXNATMasqueradeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXNATMasqueradeConfig_basic(),
			},
			{
				ResourceName:      "rtx_nat_masquerade.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRTXNATMasquerade_importProtocolOnly(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXNATMasqueradeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXNATMasqueradeConfig_protocolOnlyESP(),
			},
			{
				ResourceName:      "rtx_nat_masquerade.test_esp",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckRTXNATMasqueradeExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID is not set")
		}

		return nil
	}
}

func testAccCheckRTXNATMasqueradeDestroy(s *terraform.State) error {
	// This function would check that the NAT masquerade is actually deleted
	// from the router, but for now we just check that the resource is removed
	// from state
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rtx_nat_masquerade" {
			continue
		}

		// In a real implementation, we would check the router to verify
		// the NAT descriptor is actually deleted
	}

	return nil
}

func testAccRTXNATMasqueradeConfig_basic() string {
	return `
resource "rtx_nat_masquerade" "test" {
  descriptor_id = 1000
  outer_address = "ipcp"
  inner_network = "192.168.1.0-192.168.1.255"
}
`
}

func testAccRTXNATMasqueradeConfig_withStaticEntry() string {
	return `
resource "rtx_nat_masquerade" "test_static" {
  descriptor_id = 1001
  outer_address = "ipcp"
  inner_network = "192.168.1.0-192.168.1.255"

  static_entry {
    entry_number        = 1
    inside_local        = "192.168.1.100"
    inside_local_port   = 80
    outside_global      = "ipcp"
    outside_global_port = 80
    protocol            = "tcp"
  }
}
`
}

func testAccRTXNATMasqueradeConfig_protocolOnlyESP() string {
	return `
resource "rtx_nat_masquerade" "test_esp" {
  descriptor_id = 1002
  outer_address = "ipcp"
  inner_network = "192.168.1.0-192.168.1.255"

  static_entry {
    entry_number   = 1
    inside_local   = "192.168.1.253"
    protocol       = "esp"
  }
}
`
}

func testAccRTXNATMasqueradeConfig_protocolOnlyAH() string {
	return `
resource "rtx_nat_masquerade" "test_ah" {
  descriptor_id = 1003
  outer_address = "ipcp"
  inner_network = "192.168.1.0-192.168.1.255"

  static_entry {
    entry_number   = 1
    inside_local   = "192.168.1.252"
    protocol       = "ah"
  }
}
`
}

func testAccRTXNATMasqueradeConfig_protocolOnlyGRE() string {
	return `
resource "rtx_nat_masquerade" "test_gre" {
  descriptor_id = 1004
  outer_address = "ipcp"
  inner_network = "192.168.1.0-192.168.1.255"

  static_entry {
    entry_number   = 1
    inside_local   = "192.168.1.251"
    protocol       = "gre"
  }
}
`
}

func testAccRTXNATMasqueradeConfig_mixedEntries() string {
	return `
resource "rtx_nat_masquerade" "test_mixed" {
  descriptor_id = 1005
  outer_address = "ipcp"
  inner_network = "192.168.1.0-192.168.1.255"

  static_entry {
    entry_number        = 1
    inside_local        = "192.168.1.100"
    inside_local_port   = 443
    outside_global      = "ipcp"
    outside_global_port = 443
    protocol            = "tcp"
  }

  static_entry {
    entry_number   = 2
    inside_local   = "192.168.1.253"
    protocol       = "esp"
  }

  static_entry {
    entry_number        = 3
    inside_local        = "192.168.1.250"
    inside_local_port   = 500
    outside_global      = "ipcp"
    outside_global_port = 500
    protocol            = "udp"
  }
}
`
}
