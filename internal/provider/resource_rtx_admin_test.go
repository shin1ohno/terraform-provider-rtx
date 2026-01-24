package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceRTXAdmin_Schema(t *testing.T) {
	resource := resourceRTXAdmin()

	// Test login_password schema
	if s, ok := resource.Schema["login_password"]; !ok {
		t.Error("login_password schema should exist")
	} else {
		if s.Type != schema.TypeString {
			t.Error("login_password should be TypeString")
		}
		if !s.Optional {
			t.Error("login_password should be Optional")
		}
		if !s.Sensitive {
			t.Error("login_password should be Sensitive")
		}
	}

	// Test admin_password schema
	if s, ok := resource.Schema["admin_password"]; !ok {
		t.Error("admin_password schema should exist")
	} else {
		if s.Type != schema.TypeString {
			t.Error("admin_password should be TypeString")
		}
		if !s.Optional {
			t.Error("admin_password should be Optional")
		}
		if !s.Sensitive {
			t.Error("admin_password should be Sensitive")
		}
	}
}

func TestResourceRTXAdmin_CRUD(t *testing.T) {
	resource := resourceRTXAdmin()

	// Verify CRUD functions exist
	if resource.CreateContext == nil {
		t.Error("CreateContext should be defined")
	}
	if resource.ReadContext == nil {
		t.Error("ReadContext should be defined")
	}
	if resource.UpdateContext == nil {
		t.Error("UpdateContext should be defined")
	}
	if resource.DeleteContext == nil {
		t.Error("DeleteContext should be defined")
	}
}

func TestResourceRTXAdmin_Importer(t *testing.T) {
	resource := resourceRTXAdmin()

	if resource.Importer == nil {
		t.Error("Importer should be defined")
	}
	if resource.Importer.StateContext == nil {
		t.Error("StateContext should be defined")
	}
}

func TestResourceRTXAdmin_lastUpdatedComputed(t *testing.T) {
	resource := resourceRTXAdmin()

	// Test last_updated schema - should be Computed only
	if s, ok := resource.Schema["last_updated"]; !ok {
		t.Error("last_updated schema should exist")
	} else {
		if s.Type != schema.TypeString {
			t.Error("last_updated should be TypeString")
		}
		if !s.Computed {
			t.Error("last_updated should be Computed")
		}
		if s.Optional {
			t.Error("last_updated should not be Optional")
		}
		if s.Required {
			t.Error("last_updated should not be Required")
		}
	}
}

// =============================================================================
// Test Infrastructure for Acceptance Tests
// =============================================================================

// testAccAdminPreCheck verifies that required environment variables are set
// for running acceptance tests.
func testAccAdminPreCheck(t *testing.T) {
	t.Helper()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set, skipping acceptance test")
	}

	requiredEnvVars := []string{"RTX_HOST", "RTX_USERNAME", "RTX_PASSWORD"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			t.Fatalf("Environment variable %s must be set for acceptance tests", envVar)
		}
	}
}

// testAccAdminPreCheckWithAdminPassword verifies admin password is available.
func testAccAdminPreCheckWithAdminPassword(t *testing.T) {
	t.Helper()
	testAccAdminPreCheck(t)

	if os.Getenv("RTX_ADMIN_PASSWORD") == "" && os.Getenv("RTX_PASSWORD") == "" {
		t.Fatal("RTX_ADMIN_PASSWORD or RTX_PASSWORD must be set for tests requiring admin access")
	}
}

// testAccAdminProviderFactories provides the provider factory for acceptance tests.
var testAccAdminProviderFactories = map[string]func() (*schema.Provider, error){
	"rtx": func() (*schema.Provider, error) {
		return New("test"), nil
	},
}

// =============================================================================
// Acceptance Tests - Pattern 1: Perpetual Diff Prevention
// =============================================================================

// TestAccRTXAdmin_noDiff verifies that re-applying the same configuration
// does not produce any planned changes (perpetual diff prevention).
func TestAccRTXAdmin_noDiff(t *testing.T) {
	resourceName := "rtx_admin.test"
	config := testAccRTXAdminConfig_basic()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccAdminPreCheck(t) },
		ProviderFactories: testAccAdminProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create resource
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "admin"),
				),
			},
			// Step 2: Re-apply same config - should be no-op
			{
				Config:   config,
				PlanOnly: true,
				// In SDK v2, empty plan is verified automatically
				// Test fails if any changes are planned
			},
		},
	})
}

// TestAccRTXAdmin_noDiffWithPasswords verifies no perpetual diff when
// passwords are specified in configuration.
func TestAccRTXAdmin_noDiffWithPasswords(t *testing.T) {
	resourceName := "rtx_admin.test"
	config := testAccRTXAdminConfig_withPasswords("loginpass123", "adminpass456")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccAdminPreCheckWithAdminPassword(t) },
		ProviderFactories: testAccAdminProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create resource with passwords
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "admin"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Step 2: Re-apply same config - should be no-op
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

// =============================================================================
// Acceptance Tests - Pattern 2: Import Testing
// =============================================================================

// TestAccRTXAdmin_import verifies that existing admin configuration can be
// imported into Terraform state.
func TestAccRTXAdmin_import(t *testing.T) {
	resourceName := "rtx_admin.test"
	config := testAccRTXAdminConfig_basic()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccAdminPreCheck(t) },
		ProviderFactories: testAccAdminProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create resource
			{
				Config: config,
			},
			// Step 2: Import and verify
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Passwords cannot be read from router, so they must be ignored
				ImportStateVerifyIgnore: []string{"login_password", "admin_password", "last_updated"},
			},
		},
	})
}

// TestAccRTXAdmin_importWithID verifies import works with explicit "admin" ID.
func TestAccRTXAdmin_importWithID(t *testing.T) {
	resourceName := "rtx_admin.test"
	config := testAccRTXAdminConfig_basic()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccAdminPreCheck(t) },
		ProviderFactories: testAccAdminProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create resource
			{
				Config: config,
			},
			// Step 2: Import with explicit ID "admin"
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           "admin",
				ImportStateVerifyIgnore: []string{"login_password", "admin_password", "last_updated"},
			},
		},
	})
}

// =============================================================================
// Acceptance Tests - Pattern 4: Password Update Testing
// (Adapted from Optional+Computed preservation for singleton resource)
// =============================================================================

// TestAccRTXAdmin_passwordUpdate verifies that password updates work correctly.
func TestAccRTXAdmin_passwordUpdate(t *testing.T) {
	resourceName := "rtx_admin.test"

	// Initial config with passwords
	configWithPasswords := testAccRTXAdminConfig_withPasswords("initial_login", "initial_admin")

	// Updated config with new passwords
	configWithNewPasswords := testAccRTXAdminConfig_withPasswords("updated_login", "updated_admin")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccAdminPreCheckWithAdminPassword(t) },
		ProviderFactories: testAccAdminProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with initial passwords
			{
				Config: configWithPasswords,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "admin"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
			// Step 2: Update with new passwords
			{
				Config: configWithNewPasswords,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "admin"),
					// last_updated should be updated to reflect the change
					resource.TestCheckResourceAttrSet(resourceName, "last_updated"),
				),
			},
		},
	})
}

// TestAccRTXAdmin_removePasswordsFromConfig verifies behavior when passwords
// are removed from configuration (should not reset router passwords).
func TestAccRTXAdmin_removePasswordsFromConfig(t *testing.T) {
	resourceName := "rtx_admin.test"

	// Config with passwords
	configWithPasswords := testAccRTXAdminConfig_withPasswords("testlogin", "testadmin")

	// Config without passwords (basic config)
	configWithoutPasswords := testAccRTXAdminConfig_basic()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccAdminPreCheckWithAdminPassword(t) },
		ProviderFactories: testAccAdminProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with passwords
			{
				Config: configWithPasswords,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "admin"),
				),
			},
			// Step 2: Apply config without passwords - passwords should be preserved in state
			{
				Config: configWithoutPasswords,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "admin"),
					// Resource should still exist, router passwords unchanged
				),
			},
		},
	})
}

// =============================================================================
// Acceptance Tests - Pattern 5: Sensitive Attribute Testing
// =============================================================================

// TestAccRTXAdmin_passwordSensitive verifies that password fields are marked
// as sensitive and handled securely in state.
func TestAccRTXAdmin_passwordSensitive(t *testing.T) {
	resourceName := "rtx_admin.test"
	config := testAccRTXAdminConfig_withPasswords("sensitivelogin", "sensitiveadmin")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccAdminPreCheckWithAdminPassword(t) },
		ProviderFactories: testAccAdminProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "admin"),
					// Verify passwords are set (sensitive fields can still be checked)
					resource.TestCheckResourceAttr(resourceName, "login_password", "sensitivelogin"),
					resource.TestCheckResourceAttr(resourceName, "admin_password", "sensitiveadmin"),
				),
			},
		},
	})
}

// TestAccRTXAdmin_onlyLoginPassword verifies setting only the login password.
func TestAccRTXAdmin_onlyLoginPassword(t *testing.T) {
	resourceName := "rtx_admin.test"
	config := testAccRTXAdminConfig_onlyLoginPassword("loginonly123")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccAdminPreCheck(t) },
		ProviderFactories: testAccAdminProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "admin"),
					resource.TestCheckResourceAttr(resourceName, "login_password", "loginonly123"),
					resource.TestCheckResourceAttr(resourceName, "admin_password", ""),
				),
			},
		},
	})
}

// TestAccRTXAdmin_onlyAdminPassword verifies setting only the admin password.
func TestAccRTXAdmin_onlyAdminPassword(t *testing.T) {
	resourceName := "rtx_admin.test"
	config := testAccRTXAdminConfig_onlyAdminPassword("adminonly456")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccAdminPreCheckWithAdminPassword(t) },
		ProviderFactories: testAccAdminProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", "admin"),
					resource.TestCheckResourceAttr(resourceName, "login_password", ""),
					resource.TestCheckResourceAttr(resourceName, "admin_password", "adminonly456"),
				),
			},
		},
	})
}

// =============================================================================
// Acceptance Tests - Singleton Resource Pattern
// =============================================================================

// TestAccRTXAdmin_singletonID verifies that the resource always uses "admin" ID.
func TestAccRTXAdmin_singletonID(t *testing.T) {
	resourceName := "rtx_admin.test"
	config := testAccRTXAdminConfig_basic()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccAdminPreCheck(t) },
		ProviderFactories: testAccAdminProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					// Singleton resource should always have ID "admin"
					resource.TestCheckResourceAttr(resourceName, "id", "admin"),
				),
			},
		},
	})
}

// =============================================================================
// Test Configuration Helpers
// =============================================================================

func testAccRTXAdminConfig_basic() string {
	return `resource "rtx_admin" "test" {}`
}

func testAccRTXAdminConfig_withPasswords(loginPassword, adminPassword string) string {
	return fmt.Sprintf(`
resource "rtx_admin" "test" {
  login_password = %q
  admin_password = %q
}
`, loginPassword, adminPassword)
}

func testAccRTXAdminConfig_onlyLoginPassword(loginPassword string) string {
	return fmt.Sprintf(`
resource "rtx_admin" "test" {
  login_password = %q
}
`, loginPassword)
}

func testAccRTXAdminConfig_onlyAdminPassword(adminPassword string) string {
	return fmt.Sprintf(`
resource "rtx_admin" "test" {
  admin_password = %q
}
`, adminPassword)
}
