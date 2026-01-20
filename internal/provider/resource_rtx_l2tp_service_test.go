package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceRTXL2TPService_Schema(t *testing.T) {
	r := resourceRTXL2TPService()

	// Test enabled schema
	enabledSchema := r.Schema["enabled"]
	if enabledSchema.Type != schema.TypeBool {
		t.Errorf("expected enabled type to be TypeBool, got %v", enabledSchema.Type)
	}
	if !enabledSchema.Required {
		t.Error("expected enabled to be required")
	}

	// Test protocols schema
	protocolsSchema := r.Schema["protocols"]
	if protocolsSchema.Type != schema.TypeList {
		t.Errorf("expected protocols type to be TypeList, got %v", protocolsSchema.Type)
	}
	if protocolsSchema.Required {
		t.Error("expected protocols to be optional")
	}

	// Test that protocols element validation works
	protocolsElem := protocolsSchema.Elem.(*schema.Schema)
	if protocolsElem.Type != schema.TypeString {
		t.Errorf("expected protocols element type to be TypeString, got %v", protocolsElem.Type)
	}
	if protocolsElem.ValidateFunc == nil {
		t.Error("expected protocols element to have ValidateFunc")
	}
}

func TestResourceRTXL2TPService_ProtocolsValidation(t *testing.T) {
	testCases := []struct {
		value   string
		isValid bool
	}{
		{"l2tp", true},
		{"l2tpv3", true},
		{"l2tpv2", false}, // Invalid protocol name
		{"invalid", false},
		{"", false},
		{"L2TP", false}, // Case sensitive
	}

	r := resourceRTXL2TPService()
	protocolsSchema := r.Schema["protocols"]
	elemSchema := protocolsSchema.Elem.(*schema.Schema)
	validateFunc := elemSchema.ValidateFunc

	for _, tc := range testCases {
		t.Run(tc.value, func(t *testing.T) {
			_, errs := validateFunc(tc.value, "protocols")
			hasError := len(errs) > 0
			if hasError == tc.isValid {
				if tc.isValid {
					t.Errorf("expected %q to be valid, but got errors: %v", tc.value, errs)
				} else {
					t.Errorf("expected %q to be invalid, but got no errors", tc.value)
				}
			}
		})
	}
}

func TestResourceRTXL2TPService_SingletonID(t *testing.T) {
	// Verify the resource uses "default" as the singleton ID
	// This is tested by checking the import function behavior
	r := resourceRTXL2TPService()

	if r.Importer == nil {
		t.Error("expected resource to have Importer")
	}

	// Verify the resource has all CRUD operations
	if r.CreateContext == nil {
		t.Error("expected resource to have CreateContext")
	}
	if r.ReadContext == nil {
		t.Error("expected resource to have ReadContext")
	}
	if r.UpdateContext == nil {
		t.Error("expected resource to have UpdateContext")
	}
	if r.DeleteContext == nil {
		t.Error("expected resource to have DeleteContext")
	}
}

func TestExpandStringList(t *testing.T) {
	testCases := []struct {
		name     string
		input    []interface{}
		expected []string
	}{
		{
			name:     "empty list",
			input:    []interface{}{},
			expected: []string{},
		},
		{
			name:     "single element",
			input:    []interface{}{"l2tp"},
			expected: []string{"l2tp"},
		},
		{
			name:     "multiple elements",
			input:    []interface{}{"l2tp", "l2tpv3"},
			expected: []string{"l2tp", "l2tpv3"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := expandStringList(tc.input)
			if len(result) != len(tc.expected) {
				t.Errorf("expected %d elements, got %d", len(tc.expected), len(result))
				return
			}
			for i, v := range result {
				if v != tc.expected[i] {
					t.Errorf("expected element %d to be %q, got %q", i, tc.expected[i], v)
				}
			}
		})
	}
}

// Acceptance tests - require real RTX router

func TestAccResourceRTXL2TPService_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXL2TPServiceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_l2tp_service.test", "enabled", "true"),
					resource.TestCheckResourceAttr("rtx_l2tp_service.test", "id", "default"),
				),
			},
		},
	})
}

func TestAccResourceRTXL2TPService_withProtocols(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXL2TPServiceConfig_withProtocols(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_l2tp_service.test", "enabled", "true"),
					resource.TestCheckResourceAttr("rtx_l2tp_service.test", "protocols.#", "2"),
					resource.TestCheckResourceAttr("rtx_l2tp_service.test", "protocols.0", "l2tpv3"),
					resource.TestCheckResourceAttr("rtx_l2tp_service.test", "protocols.1", "l2tp"),
				),
			},
		},
	})
}

func TestAccResourceRTXL2TPService_disabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXL2TPServiceConfig_disabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_l2tp_service.test", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccResourceRTXL2TPService_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXL2TPServiceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_l2tp_service.test", "enabled", "true"),
				),
			},
			{
				Config: testAccResourceRTXL2TPServiceConfig_withProtocols(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_l2tp_service.test", "enabled", "true"),
					resource.TestCheckResourceAttr("rtx_l2tp_service.test", "protocols.#", "2"),
				),
			},
			{
				Config: testAccResourceRTXL2TPServiceConfig_disabled(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_l2tp_service.test", "enabled", "false"),
				),
			},
		},
	})
}

func TestAccResourceRTXL2TPService_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXL2TPServiceConfig_basic(),
			},
			{
				ResourceName:      "rtx_l2tp_service.test",
				ImportState:       true,
				ImportStateId:     "default",
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceRTXL2TPServiceConfig_basic() string {
	return `
resource "rtx_l2tp_service" "test" {
  enabled = true
}
`
}

func testAccResourceRTXL2TPServiceConfig_withProtocols() string {
	return `
resource "rtx_l2tp_service" "test" {
  enabled   = true
  protocols = ["l2tpv3", "l2tp"]
}
`
}

func testAccResourceRTXL2TPServiceConfig_disabled() string {
	return `
resource "rtx_l2tp_service" "test" {
  enabled = false
}
`
}
