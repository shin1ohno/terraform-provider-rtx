package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceRTXHTTPD_Schema(t *testing.T) {
	r := resourceRTXHTTPD()

	// Test host schema
	hostSchema := r.Schema["host"]
	if hostSchema.Type != schema.TypeString {
		t.Errorf("expected host type to be TypeString, got %v", hostSchema.Type)
	}
	if !hostSchema.Required {
		t.Error("expected host to be required")
	}

	// Test proxy_access schema
	proxySchema := r.Schema["proxy_access"]
	if proxySchema.Type != schema.TypeBool {
		t.Errorf("expected proxy_access type to be TypeBool, got %v", proxySchema.Type)
	}
	if proxySchema.Required {
		t.Error("expected proxy_access to be optional")
	}
	if proxySchema.Default != false {
		t.Errorf("expected proxy_access default to be false, got %v", proxySchema.Default)
	}
}

func TestResourceRTXHTTPD_HostValidation(t *testing.T) {
	testCases := []struct {
		value   string
		isValid bool
	}{
		{"any", true},
		{"lan1", true},
		{"lan2", true},
		{"pp1", true},
		{"bridge1", true},
		{"tunnel1", true},
		{"invalid", false},
		{"", false},
		{"eth0", false},
		{"vlan1", false},
	}

	r := resourceRTXHTTPD()
	validateFunc := r.Schema["host"].ValidateFunc

	for _, tc := range testCases {
		t.Run(tc.value, func(t *testing.T) {
			_, errs := validateFunc(tc.value, "host")
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

func TestAccResourceRTXHTTPD_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXHTTPDConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_httpd.test", "host", "any"),
					resource.TestCheckResourceAttr("rtx_httpd.test", "proxy_access", "false"),
				),
			},
		},
	})
}

func TestAccResourceRTXHTTPD_withProxyAccess(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXHTTPDConfig_withProxyAccess(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_httpd.test", "host", "lan1"),
					resource.TestCheckResourceAttr("rtx_httpd.test", "proxy_access", "true"),
				),
			},
		},
	})
}

func TestAccResourceRTXHTTPD_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXHTTPDConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_httpd.test", "host", "any"),
					resource.TestCheckResourceAttr("rtx_httpd.test", "proxy_access", "false"),
				),
			},
			{
				Config: testAccResourceRTXHTTPDConfig_withProxyAccess(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_httpd.test", "host", "lan1"),
					resource.TestCheckResourceAttr("rtx_httpd.test", "proxy_access", "true"),
				),
			},
		},
	})
}

func testAccResourceRTXHTTPDConfig_basic() string {
	return `
resource "rtx_httpd" "test" {
  host = "any"
}
`
}

func testAccResourceRTXHTTPDConfig_withProxyAccess() string {
	return `
resource "rtx_httpd" "test" {
  host         = "lan1"
  proxy_access = true
}
`
}
