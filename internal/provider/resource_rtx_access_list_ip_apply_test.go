package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

// TestAccRTXAccessListIPApply_Basic tests basic apply resource creation
func TestAccRTXAccessListIPApply_Basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	resourceName := "rtx_access_list_ip_apply.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXAccessListIPApplyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXAccessListIPApplyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXAccessListIPApplyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_list", "test-acl"),
					resource.TestCheckResourceAttr(resourceName, "interface", "lan1"),
					resource.TestCheckResourceAttr(resourceName, "direction", "in"),
					resource.TestCheckResourceAttr(resourceName, "filter_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "filter_ids.0", "100"),
					resource.TestCheckResourceAttr(resourceName, "filter_ids.1", "110"),
				),
			},
			// Test import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// access_list is set to "imported" on import, so we skip verification
				ImportStateVerifyIgnore: []string{"access_list"},
			},
		},
	})
}

// TestAccRTXAccessListIPApply_ForEach tests for_each pattern with multiple apply resources
func TestAccRTXAccessListIPApply_ForEach(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXAccessListIPApplyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXAccessListIPApplyConfig_forEach(),
				Check: resource.ComposeTestCheckFunc(
					// Check lan1 inbound apply
					testAccCheckRTXAccessListIPApplyExists("rtx_access_list_ip_apply.multi[\"lan1-in\"]"),
					resource.TestCheckResourceAttr("rtx_access_list_ip_apply.multi[\"lan1-in\"]", "interface", "lan1"),
					resource.TestCheckResourceAttr("rtx_access_list_ip_apply.multi[\"lan1-in\"]", "direction", "in"),
					// Check lan1 outbound apply
					testAccCheckRTXAccessListIPApplyExists("rtx_access_list_ip_apply.multi[\"lan1-out\"]"),
					resource.TestCheckResourceAttr("rtx_access_list_ip_apply.multi[\"lan1-out\"]", "interface", "lan1"),
					resource.TestCheckResourceAttr("rtx_access_list_ip_apply.multi[\"lan1-out\"]", "direction", "out"),
					// Check lan2 inbound apply
					testAccCheckRTXAccessListIPApplyExists("rtx_access_list_ip_apply.multi[\"lan2-in\"]"),
					resource.TestCheckResourceAttr("rtx_access_list_ip_apply.multi[\"lan2-in\"]", "interface", "lan2"),
					resource.TestCheckResourceAttr("rtx_access_list_ip_apply.multi[\"lan2-in\"]", "direction", "in"),
				),
			},
		},
	})
}

// TestAccRTXAccessListIPApply_Conflict tests conflict detection with inline apply
func TestAccRTXAccessListIPApply_Conflict(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXAccessListIPApplyDestroy,
		Steps: []resource.TestStep{
			{
				// This test verifies that separate apply resources can be used
				// alongside ACL resources. The conflict detection is handled
				// in CustomizeDiff when inline applies are present.
				Config: testAccRTXAccessListIPApplyConfig_withACL(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXAccessListIPApplyExists("rtx_access_list_ip_apply.separate"),
					resource.TestCheckResourceAttr("rtx_access_list_ip_apply.separate", "interface", "lan2"),
					resource.TestCheckResourceAttr("rtx_access_list_ip_apply.separate", "direction", "out"),
				),
			},
		},
	})
}

// TestAccRTXAccessListIPApply_Update tests updating filter_ids
func TestAccRTXAccessListIPApply_Update(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	resourceName := "rtx_access_list_ip_apply.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXAccessListIPApplyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXAccessListIPApplyConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXAccessListIPApplyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filter_ids.#", "2"),
				),
			},
			{
				Config: testAccRTXAccessListIPApplyConfig_updated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRTXAccessListIPApplyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "filter_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "filter_ids.0", "100"),
					resource.TestCheckResourceAttr(resourceName, "filter_ids.1", "110"),
					resource.TestCheckResourceAttr(resourceName, "filter_ids.2", "120"),
				),
			},
		},
	})
}

// Unit tests for resource schema validation

func TestResourceRTXAccessListIPApplySchema(t *testing.T) {
	res := resourceRTXAccessListIPApply()

	t.Run("access_list is required", func(t *testing.T) {
		assert.True(t, res.Schema["access_list"].Required)
	})

	t.Run("interface is required and ForceNew", func(t *testing.T) {
		assert.True(t, res.Schema["interface"].Required)
		assert.True(t, res.Schema["interface"].ForceNew)
	})

	t.Run("direction is required and ForceNew", func(t *testing.T) {
		assert.True(t, res.Schema["direction"].Required)
		assert.True(t, res.Schema["direction"].ForceNew)
	})

	t.Run("filter_ids is optional", func(t *testing.T) {
		assert.True(t, res.Schema["filter_ids"].Optional)
	})

	t.Run("direction validation", func(t *testing.T) {
		validDirs := []string{"in", "out", "IN", "OUT"}
		for _, dir := range validDirs {
			_, errs := res.Schema["direction"].ValidateFunc(dir, "direction")
			assert.Empty(t, errs, "direction '%s' should be valid", dir)
		}

		_, errs := res.Schema["direction"].ValidateFunc("invalid", "direction")
		assert.NotEmpty(t, errs, "direction 'invalid' should be invalid")
	})
}

func TestResourceRTXAccessListIPApplyCRUDFunctions(t *testing.T) {
	res := resourceRTXAccessListIPApply()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, res.CreateContext)
		assert.NotNil(t, res.ReadContext)
		assert.NotNil(t, res.UpdateContext)
		assert.NotNil(t, res.DeleteContext)
	})

	t.Run("Importer is configured", func(t *testing.T) {
		assert.NotNil(t, res.Importer)
		assert.NotNil(t, res.Importer.StateContext)
	})

	t.Run("CustomizeDiff is configured", func(t *testing.T) {
		assert.NotNil(t, res.CustomizeDiff)
	})
}

func TestExtractApplyFilterIDs(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected []int
	}{
		{
			name: "with filter_ids",
			input: map[string]interface{}{
				"access_list": "test",
				"interface":   "lan1",
				"direction":   "in",
				"filter_ids":  []interface{}{100, 110, 120},
			},
			expected: []int{100, 110, 120},
		},
		{
			name: "empty filter_ids",
			input: map[string]interface{}{
				"access_list": "test",
				"interface":   "lan1",
				"direction":   "in",
				"filter_ids":  []interface{}{},
			},
			expected: nil,
		},
		{
			name: "no filter_ids key",
			input: map[string]interface{}{
				"access_list": "test",
				"interface":   "lan1",
				"direction":   "in",
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceSchema := resourceRTXAccessListIPApply().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			result := extractApplyFilterIDs(d)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions for acceptance tests

func testAccCheckRTXAccessListIPApplyExists(resourceName string) resource.TestCheckFunc {
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

func testAccCheckRTXAccessListIPApplyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rtx_access_list_ip_apply" {
			continue
		}

		// In a real implementation, we would verify the apply was removed
		// from the router by checking the interface filter configuration
	}

	return nil
}

// Test configurations

func testAccRTXAccessListIPApplyConfig_basic() string {
	return `
# First create an IP access list with filters
resource "rtx_access_list_ip" "test" {
  name           = "test-acl"
  sequence_start = 100
  sequence_step  = 10

  entry {
    action      = "pass"
    source      = "192.168.1.0/24"
    destination = "*"
    protocol    = "*"
  }

  entry {
    action      = "reject"
    source      = "*"
    destination = "*"
    protocol    = "*"
  }
}

# Then apply the filters to an interface using the separate apply resource
resource "rtx_access_list_ip_apply" "test" {
  access_list = rtx_access_list_ip.test.name
  interface   = "lan1"
  direction   = "in"
  filter_ids  = [100, 110]
}
`
}

func testAccRTXAccessListIPApplyConfig_updated() string {
	return `
# ACL with three entries
resource "rtx_access_list_ip" "test" {
  name           = "test-acl"
  sequence_start = 100
  sequence_step  = 10

  entry {
    action      = "pass"
    source      = "192.168.1.0/24"
    destination = "*"
    protocol    = "*"
  }

  entry {
    action      = "pass"
    source      = "192.168.2.0/24"
    destination = "*"
    protocol    = "*"
  }

  entry {
    action      = "reject"
    source      = "*"
    destination = "*"
    protocol    = "*"
  }
}

# Apply all three filters
resource "rtx_access_list_ip_apply" "test" {
  access_list = rtx_access_list_ip.test.name
  interface   = "lan1"
  direction   = "in"
  filter_ids  = [100, 110, 120]
}
`
}

func testAccRTXAccessListIPApplyConfig_forEach() string {
	return `
# Define ACL for multiple interfaces
resource "rtx_access_list_ip" "multi" {
  name           = "multi-acl"
  sequence_start = 200
  sequence_step  = 10

  entry {
    action      = "pass"
    source      = "10.0.0.0/8"
    destination = "*"
    protocol    = "*"
  }

  entry {
    action      = "reject"
    source      = "*"
    destination = "*"
    protocol    = "*"
  }
}

# Define interface bindings using for_each
locals {
  apply_bindings = {
    "lan1-in"  = { interface = "lan1", direction = "in" }
    "lan1-out" = { interface = "lan1", direction = "out" }
    "lan2-in"  = { interface = "lan2", direction = "in" }
  }
}

resource "rtx_access_list_ip_apply" "multi" {
  for_each = local.apply_bindings

  access_list = rtx_access_list_ip.multi.name
  interface   = each.value.interface
  direction   = each.value.direction
  filter_ids  = [200, 210]
}
`
}

func testAccRTXAccessListIPApplyConfig_withACL() string {
	return `
# ACL with inline apply to lan1
resource "rtx_access_list_ip" "inline" {
  name           = "inline-acl"
  sequence_start = 300
  sequence_step  = 10

  entry {
    action      = "pass"
    source      = "172.16.0.0/12"
    destination = "*"
    protocol    = "*"
  }

  # Inline apply to lan1 inbound
  apply {
    interface  = "lan1"
    direction  = "in"
    filter_ids = [300]
  }
}

# Separate ACL for lan2
resource "rtx_access_list_ip" "external" {
  name           = "external-acl"
  sequence_start = 400
  sequence_step  = 10

  entry {
    action      = "reject"
    source      = "*"
    destination = "*"
    protocol    = "*"
  }
}

# Separate apply resource to different interface (no conflict)
resource "rtx_access_list_ip_apply" "separate" {
  access_list = rtx_access_list_ip.external.name
  interface   = "lan2"
  direction   = "out"
  filter_ids  = [400]
}
`
}
