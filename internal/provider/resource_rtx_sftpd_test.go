package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceRTXSFTPD_Schema(t *testing.T) {
	r := resourceRTXSFTPD()

	// Test hosts schema
	hostsSchema := r.Schema["hosts"]
	if hostsSchema.Type != schema.TypeList {
		t.Errorf("expected hosts type to be TypeList, got %v", hostsSchema.Type)
	}
	if !hostsSchema.Required {
		t.Error("expected hosts to be required")
	}
	if hostsSchema.MinItems != 1 {
		t.Errorf("expected MinItems to be 1, got %d", hostsSchema.MinItems)
	}
}

func TestResourceRTXSFTPD_HostsValidation(t *testing.T) {
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

	r := resourceRTXSFTPD()
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

func TestAccResourceRTXSFTPD_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXSFTPDConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_sftpd.test", "hosts.#", "1"),
					resource.TestCheckResourceAttr("rtx_sftpd.test", "hosts.0", "lan1"),
				),
			},
		},
	})
}

func TestAccResourceRTXSFTPD_multipleHosts(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXSFTPDConfig_multipleHosts(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_sftpd.test", "hosts.#", "2"),
					resource.TestCheckResourceAttr("rtx_sftpd.test", "hosts.0", "lan1"),
					resource.TestCheckResourceAttr("rtx_sftpd.test", "hosts.1", "lan2"),
				),
			},
		},
	})
}

func TestAccResourceRTXSFTPD_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXSFTPDConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_sftpd.test", "hosts.#", "1"),
					resource.TestCheckResourceAttr("rtx_sftpd.test", "hosts.0", "lan1"),
				),
			},
			{
				Config: testAccResourceRTXSFTPDConfig_multipleHosts(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("rtx_sftpd.test", "hosts.#", "2"),
					resource.TestCheckResourceAttr("rtx_sftpd.test", "hosts.0", "lan1"),
					resource.TestCheckResourceAttr("rtx_sftpd.test", "hosts.1", "lan2"),
				),
			},
		},
	})
}

func TestAccResourceRTXSFTPD_import(t *testing.T) {
	resourceName := "rtx_sftpd.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRTXSFTPDConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "hosts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "hosts.0", "lan1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceRTXSFTPD_noDiff(t *testing.T) {
	resourceName := "rtx_sftpd.test"
	config := testAccResourceRTXSFTPDConfig_multipleHosts()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "hosts.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "hosts.0", "lan1"),
					resource.TestCheckResourceAttr(resourceName, "hosts.1", "lan2"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

func testAccResourceRTXSFTPDConfig_basic() string {
	return `
resource "rtx_sftpd" "test" {
  hosts = ["lan1"]
}
`
}

func testAccResourceRTXSFTPDConfig_multipleHosts() string {
	return `
resource "rtx_sftpd" "test" {
  hosts = ["lan1", "lan2"]
}
`
}
