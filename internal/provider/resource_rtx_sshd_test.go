package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceRTXSSHD_Schema(t *testing.T) {
	r := resourceRTXSSHD()

	// Test enabled schema
	enabledSchema := r.Schema["enabled"]
	if enabledSchema.Type != schema.TypeBool {
		t.Errorf("expected enabled type to be TypeBool, got %v", enabledSchema.Type)
	}
	if !enabledSchema.Required {
		t.Error("expected enabled to be required")
	}

	// Test hosts schema
	hostsSchema := r.Schema["hosts"]
	if hostsSchema.Type != schema.TypeList {
		t.Errorf("expected hosts type to be TypeList, got %v", hostsSchema.Type)
	}
	if hostsSchema.Required {
		t.Error("expected hosts to be optional")
	}

	// Test host_key schema
	hostKeySchema := r.Schema["host_key"]
	if hostKeySchema.Type != schema.TypeString {
		t.Errorf("expected host_key type to be TypeString, got %v", hostKeySchema.Type)
	}
	if !hostKeySchema.Computed {
		t.Error("expected host_key to be computed")
	}
	if !hostKeySchema.Sensitive {
		t.Error("expected host_key to be sensitive")
	}
}

func TestResourceRTXSSHD_HostsValidation(t *testing.T) {
	testCases := []struct {
		value   string
		isValid bool
	}{
		{"lan1", true},
		{"lan2", true},
		{"pp1", true},
		{"bridge1", true},
		{"tunnel1", true},
		{"any", false},
		{"invalid", false},
		{"", false},
		{"eth0", false},
	}

	r := resourceRTXSSHD()
	hostsSchema := r.Schema["hosts"]
	elemSchema := hostsSchema.Elem.(*schema.Schema)
	validateFunc := elemSchema.ValidateFunc

	for _, tc := range testCases {
		t.Run(tc.value, func(t *testing.T) {
			_, errs := validateFunc(tc.value, "hosts")
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

func TestAccResourceRTXSSHD_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXSSHDConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_sshd.test", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceRTXSSHD_withHosts(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXSSHDConfig_withHosts(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_sshd.test", "enabled", "true"),
					resource.TestCheckResourceAttr("rtx_sshd.test", "hosts.#", "2"),
					resource.TestCheckResourceAttr("rtx_sshd.test", "hosts.0", "lan1"),
					resource.TestCheckResourceAttr("rtx_sshd.test", "hosts.1", "lan2"),
				),
			},
		},
	})
}

func TestAccResourceRTXSSHD_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXSSHDConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_sshd.test", "enabled", "true"),
				),
			},
			{
				Config: testAccResourceRTXSSHDConfig_withHosts(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_sshd.test", "enabled", "true"),
					resource.TestCheckResourceAttr("rtx_sshd.test", "hosts.#", "2"),
				),
			},
		},
	})
}

func testAccResourceRTXSSHDConfig_basic() string {
	return `
resource "rtx_sshd" "test" {
  enabled = true
}
`
}

func testAccResourceRTXSSHDConfig_withHosts() string {
	return `
resource "rtx_sshd" "test" {
  enabled = true
  hosts   = ["lan1", "lan2"]
}
`
}
