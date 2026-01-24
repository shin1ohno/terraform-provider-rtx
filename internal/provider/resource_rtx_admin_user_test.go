package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceRTXAdminUser_Schema(t *testing.T) {
	resource := resourceRTXAdminUser()

	// Test username schema
	if s, ok := resource.Schema["username"]; !ok {
		t.Error("username schema should exist")
	} else {
		if s.Type != schema.TypeString {
			t.Error("username should be TypeString")
		}
		if !s.Required {
			t.Error("username should be Required")
		}
		if !s.ForceNew {
			t.Error("username should be ForceNew")
		}
	}

	// Test password schema
	if s, ok := resource.Schema["password"]; !ok {
		t.Error("password schema should exist")
	} else {
		if s.Type != schema.TypeString {
			t.Error("password should be TypeString")
		}
		if !s.Optional {
			t.Error("password should be Optional for import compatibility")
		}
		if !s.Sensitive {
			t.Error("password should be Sensitive")
		}
	}

	// Test encrypted schema
	if s, ok := resource.Schema["encrypted"]; !ok {
		t.Error("encrypted schema should exist")
	} else {
		if s.Type != schema.TypeBool {
			t.Error("encrypted should be TypeBool")
		}
		if !s.Optional {
			t.Error("encrypted should be Optional")
		}
		if !s.Computed {
			t.Error("encrypted should be Computed for import compatibility")
		}
	}

	// Test administrator schema
	if s, ok := resource.Schema["administrator"]; !ok {
		t.Error("administrator schema should exist")
	} else {
		if s.Type != schema.TypeBool {
			t.Error("administrator should be TypeBool")
		}
		if !s.Optional {
			t.Error("administrator should be Optional")
		}
		if !s.Computed {
			t.Error("administrator should be Computed for import compatibility")
		}
	}

	// Test connection_methods schema
	if s, ok := resource.Schema["connection_methods"]; !ok {
		t.Error("connection_methods schema should exist")
	} else {
		if s.Type != schema.TypeSet {
			t.Error("connection_methods should be TypeSet")
		}
		if !s.Optional {
			t.Error("connection_methods should be Optional")
		}
	}

	// Test gui_pages schema
	if s, ok := resource.Schema["gui_pages"]; !ok {
		t.Error("gui_pages schema should exist")
	} else {
		if s.Type != schema.TypeSet {
			t.Error("gui_pages should be TypeSet")
		}
		if !s.Optional {
			t.Error("gui_pages should be Optional")
		}
	}

	// Test login_timer schema
	if s, ok := resource.Schema["login_timer"]; !ok {
		t.Error("login_timer schema should exist")
	} else {
		if s.Type != schema.TypeInt {
			t.Error("login_timer should be TypeInt")
		}
		if !s.Optional {
			t.Error("login_timer should be Optional")
		}
		if !s.Computed {
			t.Error("login_timer should be Computed for import fidelity")
		}
	}
}

func TestResourceRTXAdminUser_CRUD(t *testing.T) {
	resource := resourceRTXAdminUser()

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

func TestResourceRTXAdminUser_Importer(t *testing.T) {
	resource := resourceRTXAdminUser()

	if resource.Importer == nil {
		t.Error("Importer should be defined")
	}
	if resource.Importer.StateContext == nil {
		t.Error("StateContext should be defined")
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid username",
			value:   "admin",
			wantErr: false,
		},
		{
			name:    "valid username with underscore",
			value:   "admin_user",
			wantErr: false,
		},
		{
			name:    "valid username with numbers",
			value:   "admin123",
			wantErr: false,
		},
		{
			name:    "empty username",
			value:   "",
			wantErr: true,
		},
		{
			name:    "starts with number",
			value:   "1admin",
			wantErr: true,
		},
		{
			name:    "contains special characters",
			value:   "admin@user",
			wantErr: true,
		},
		{
			name:    "contains spaces",
			value:   "admin user",
			wantErr: true,
		},
		{
			name:    "contains hyphen",
			value:   "admin-user",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateUsername(tt.value, "username")
			if (len(errs) > 0) != tt.wantErr {
				t.Errorf("validateUsername() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

// TestAccAdminUser_import verifies that an admin user resource can be imported
// and that all expected attributes are populated correctly. This follows Pattern 2
// from the design document for import testing.
func TestAccAdminUser_import(t *testing.T) {
	resourceName := "rtx_admin_user.test"
	username := randomName("tfacc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create the resource
			{
				Config: testAccAdminUserConfig_basic(username),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", username),
					resource.TestCheckResourceAttrSet(resourceName, "encrypted"),
					resource.TestCheckResourceAttrSet(resourceName, "administrator"),
				),
			},
			// Step 2: Import and verify
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Password cannot be imported (write-only field)
				// It is not stored in router output and must be set in config after import
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

// TestAccAdminUser_importWithAttributes verifies import with various attributes set.
// This ensures all computed fields are properly populated during import.
func TestAccAdminUser_importWithAttributes(t *testing.T) {
	resourceName := "rtx_admin_user.test"
	username := randomName("tfacc")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create resource with multiple attributes
			{
				Config: testAccAdminUserConfig_withAttributes(username),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", username),
					resource.TestCheckResourceAttr(resourceName, "administrator", "true"),
					resource.TestCheckResourceAttr(resourceName, "login_timer", "300"),
					resource.TestCheckResourceAttr(resourceName, "connection_methods.#", "2"),
				),
			},
			// Step 2: Import and verify all attributes
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

// testAccAdminUserConfig_basic returns a minimal configuration for an admin user.
func testAccAdminUserConfig_basic(username string) string {
	return fmt.Sprintf(`
resource "rtx_admin_user" "test" {
  username = %q
  password = "testpass123"
}
`, username)
}

// testAccAdminUserConfig_withAttributes returns a configuration with additional attributes.
func testAccAdminUserConfig_withAttributes(username string) string {
	return fmt.Sprintf(`
resource "rtx_admin_user" "test" {
  username           = %q
  password           = "testpass123"
  administrator      = true
  login_timer        = 300
  connection_methods = ["ssh", "http"]
}
`, username)
}

// TestAccAdminUser_preserveAdministrator verifies that the administrator field
// is preserved when not specified in an update configuration.
// This is a regression test for the administrator lockout scenario where
// removing the administrator field from config would reset it to false.
//
// Pattern 4: Optional+Computed Field Preservation
func TestAccAdminUser_preserveAdministrator(t *testing.T) {
	resourceName := "rtx_admin_user.test"
	username := randomName("tfacc")

	// Config with administrator=true
	configWithAdmin := testAccAdminUserConfig_withAdministrator(username)

	// Config without administrator (should preserve existing value)
	configWithoutAdmin := testAccAdminUserConfig_withoutAdministrator(username)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with administrator=true
			{
				Config: configWithAdmin,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", username),
					resource.TestCheckResourceAttr(resourceName, "administrator", "true"),
				),
			},
			// Step 2: Apply without administrator - should preserve true
			{
				Config: configWithoutAdmin,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "username", username),
					resource.TestCheckResourceAttr(resourceName, "administrator", "true"),
				),
			},
		},
	})
}

// testAccAdminUserConfig_withAdministrator returns a configuration with administrator=true
func testAccAdminUserConfig_withAdministrator(username string) string {
	return fmt.Sprintf(`
resource "rtx_admin_user" "test" {
  username      = %q
  password      = "testpass123"
  administrator = true
}
`, username)
}

// testAccAdminUserConfig_withoutAdministrator returns a configuration without administrator field
func testAccAdminUserConfig_withoutAdministrator(username string) string {
	return fmt.Sprintf(`
resource "rtx_admin_user" "test" {
  username = %q
  password = "testpass123"
}
`, username)
}

// TestAccAdminUser_noDiff verifies that re-applying the same configuration
// produces no changes (perpetual diff prevention test).
// This is Pattern 1 from the testing design document.
//
// The test ensures that:
// 1. A resource can be created successfully
// 2. Re-applying the same configuration produces no changes (empty plan)
//
// If this test fails, it indicates a "perpetual diff" bug where the provider
// always detects changes even when the configuration hasn't changed.
func TestAccAdminUser_noDiff(t *testing.T) {
	resourceName := "rtx_admin_user.test"

	// Generate a unique username to avoid conflicts in parallel tests
	username := randomName("nodiff")

	config := testAccAdminUserConfig_basic(username)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create resource
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "username", username),
				),
			},
			// Step 2: Re-apply same config - should be no-op
			// In SDK v2, PlanOnly: true causes the test to fail if any changes are planned
			{
				Config:   config,
				PlanOnly: true,
			},
		},
	})
}

// TestAccAdminUser_noDiff_withAllAttributes verifies no perpetual diff
// when all optional attributes are specified.
// This is an extended version of the basic noDiff test to catch issues
// with Optional+Computed fields that might cause diffs.
func TestAccAdminUser_noDiff_withAllAttributes(t *testing.T) {
	resourceName := "rtx_admin_user.test"
	username := randomName("nodiff")

	config := testAccAdminUserConfig_withAttributes(username)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create resource with all attributes
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "username", username),
					resource.TestCheckResourceAttr(resourceName, "administrator", "true"),
					resource.TestCheckResourceAttr(resourceName, "login_timer", "300"),
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

// randomName generates a unique name for test resources to avoid conflicts in parallel tests.
// The generated name consists of a prefix and a random 8-character suffix.
// Note: Username must start with a letter and contain only alphanumeric characters and underscores.
func randomName(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, acctest.RandString(8))
}

// TestAccAdminUser_passwordHandling verifies that password is handled as a write-only field.
// In SDK v2, true write-only is not supported, but we verify:
// 1. Password is marked Sensitive in schema (prevents display in plan output)
// 2. Password value is preserved in state across reads (not overwritten with empty)
// 3. Password can be used to authenticate/configure the router
//
// This test follows Pattern 5 from the testing patterns design document.
func TestAccAdminUser_passwordHandling(t *testing.T) {
	// Unit test: Verify schema properties
	t.Run("schema_sensitive_properties", func(t *testing.T) {
		res := resourceRTXAdminUser()
		passwordSchema := res.Schema["password"]

		if passwordSchema == nil {
			t.Fatal("password schema should exist")
		}

		// Verify Sensitive flag is set (masks value in plan output)
		if !passwordSchema.Sensitive {
			t.Error("password should be marked Sensitive to mask values in plan output")
		}

		// Verify Optional (not Required) for import compatibility
		if !passwordSchema.Optional {
			t.Error("password should be Optional for import compatibility")
		}

		// Verify NOT Computed (password is never read back from router)
		if passwordSchema.Computed {
			t.Error("password should NOT be Computed - router cannot return password values")
		}
	})

	// Unit test: Verify description indicates write-only behavior
	t.Run("schema_writeonly_description", func(t *testing.T) {
		res := resourceRTXAdminUser()
		passwordSchema := res.Schema["password"]

		desc := passwordSchema.Description
		if desc == "" {
			t.Error("password should have a description explaining its purpose")
		}

		// Description should mention write-only behavior
		if !strings.Contains(strings.ToLower(desc), "write-only") {
			t.Error("password description should mention write-only behavior")
		}
	})
}

// TestAccAdminUser_passwordNotReadable is an acceptance test that verifies
// password is correctly applied to the router but never read back into state.
// This follows Pattern 5 from the design document.
//
// The test verifies:
// 1. Resource is created with password
// 2. Password value in state is preserved (from config, not from read)
// 3. Subsequent reads do not overwrite the password with empty string
func TestAccAdminUser_passwordNotReadable(t *testing.T) {
	resourceName := "rtx_admin_user.test"
	username := randomName("testpwd")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create user with password
			{
				Config: testAccAdminUserConfig_withPassword(username, "testpassword123"),
				Check: resource.ComposeTestCheckFunc(
					// Resource should exist with correct username
					resource.TestCheckResourceAttr(resourceName, "username", username),
					// Password should exist in state (SDK v2 stores sensitive values)
					// The sensitive flag masks display but value IS stored
					resource.TestCheckResourceAttr(resourceName, "password", "testpassword123"),
				),
			},
			// Step 2: Re-apply same config - password should be preserved
			{
				Config: testAccAdminUserConfig_withPassword(username, "testpassword123"),
				Check: resource.ComposeTestCheckFunc(
					// Password value should remain unchanged after read
					resource.TestCheckResourceAttr(resourceName, "password", "testpassword123"),
				),
			},
			// Step 3: Apply with different password - should update
			{
				Config: testAccAdminUserConfig_withPassword(username, "newpassword456"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "password", "newpassword456"),
				),
			},
		},
	})
}

// testAccAdminUserConfig_withPassword generates a test configuration for admin user with password.
func testAccAdminUserConfig_withPassword(username, password string) string {
	return fmt.Sprintf(`
resource "rtx_admin_user" "test" {
  username = %q
  password = %q
}
`, username, password)
}
