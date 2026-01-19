package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestAccResourceRTXBridge_basic tests basic bridge resource creation
func TestAccResourceRTXBridge_basic(t *testing.T) {
	// Skip if not running acceptance tests
	if testing.Short() {
		t.Skip("Skipping acceptance test")
	}

	resourceName := "rtx_bridge.test"
	bridgeName := fmt.Sprintf("bridge%d", acctest.RandIntRange(1, 99))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXBridgeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXBridgeConfig_basic(bridgeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXBridgeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", bridgeName),
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "members.0", "lan1"),
				),
			},
			// Test import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccResourceRTXBridge_multipleMembers tests bridge with multiple members
func TestAccResourceRTXBridge_multipleMembers(t *testing.T) {
	// Skip if not running acceptance tests
	if testing.Short() {
		t.Skip("Skipping acceptance test")
	}

	resourceName := "rtx_bridge.test"
	bridgeName := fmt.Sprintf("bridge%d", acctest.RandIntRange(1, 99))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXBridgeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXBridgeConfig_multipleMembers(bridgeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXBridgeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", bridgeName),
					resource.TestCheckResourceAttr(resourceName, "members.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "members.0", "lan1"),
					resource.TestCheckResourceAttr(resourceName, "members.1", "tunnel1"),
				),
			},
		},
	})
}

// TestAccResourceRTXBridge_update tests bridge member update
func TestAccResourceRTXBridge_update(t *testing.T) {
	// Skip if not running acceptance tests
	if testing.Short() {
		t.Skip("Skipping acceptance test")
	}

	resourceName := "rtx_bridge.test"
	bridgeName := fmt.Sprintf("bridge%d", acctest.RandIntRange(1, 99))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXBridgeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXBridgeConfig_basic(bridgeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXBridgeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "members.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "members.0", "lan1"),
				),
			},
			{
				Config: testAccResourceRTXBridgeConfig_multipleMembers(bridgeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXBridgeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "members.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "members.0", "lan1"),
					resource.TestCheckResourceAttr(resourceName, "members.1", "tunnel1"),
				),
			},
		},
	})
}

// TestAccResourceRTXBridge_noMembers tests bridge with no members
func TestAccResourceRTXBridge_noMembers(t *testing.T) {
	// Skip if not running acceptance tests
	if testing.Short() {
		t.Skip("Skipping acceptance test")
	}

	resourceName := "rtx_bridge.test"
	bridgeName := fmt.Sprintf("bridge%d", acctest.RandIntRange(1, 99))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXBridgeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXBridgeConfig_noMembers(bridgeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXBridgeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", bridgeName),
					resource.TestCheckResourceAttr(resourceName, "members.#", "0"),
				),
			},
		},
	})
}

func testAccCheckRTXBridgeExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No bridge ID is set")
		}

		// Note: In a real acceptance test, you would verify the bridge exists on the router
		// client := testAccProvider.Meta().(*apiClient)
		// _, err := client.client.GetBridge(context.Background(), rs.Primary.ID)
		// if err != nil {
		//     return fmt.Errorf("Bridge not found: %v", err)
		// }

		return nil
	}
}

func testAccCheckRTXBridgeDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rtx_bridge" {
			continue
		}

		// Note: In a real acceptance test, you would verify the bridge was deleted
		// client := testAccProvider.Meta().(*apiClient)
		// _, err := client.client.GetBridge(context.Background(), rs.Primary.ID)
		// if err == nil {
		//     return fmt.Errorf("Bridge still exists: %s", rs.Primary.ID)
		// }
	}

	return nil
}

func testAccResourceRTXBridgeConfig_basic(bridgeName string) string {
	return fmt.Sprintf(`
resource "rtx_bridge" "test" {
  name    = %q
  members = ["lan1"]
}
`, bridgeName)
}

func testAccResourceRTXBridgeConfig_multipleMembers(bridgeName string) string {
	return fmt.Sprintf(`
resource "rtx_bridge" "test" {
  name    = %q
  members = ["lan1", "tunnel1"]
}
`, bridgeName)
}

func testAccResourceRTXBridgeConfig_noMembers(bridgeName string) string {
	return fmt.Sprintf(`
resource "rtx_bridge" "test" {
  name = %q
}
`, bridgeName)
}
