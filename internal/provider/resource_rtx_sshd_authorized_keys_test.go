package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestRTXSSHDAuthorizedKeysResourceSchema(t *testing.T) {
	resource := resourceRTXSSHDAuthorizedKeys()

	// Test that the resource is properly configured
	assert.NotNil(t, resource)

	// Verify all expected attributes exist
	expectedAttrs := []string{"username", "keys", "key_count"}
	for _, attr := range expectedAttrs {
		_, ok := resource.Schema[attr]
		assert.True(t, ok, "Expected attribute %s to exist", attr)
	}

	// Test username attribute
	usernameAttr := resource.Schema["username"]
	assert.Equal(t, schema.TypeString, usernameAttr.Type)
	assert.True(t, usernameAttr.Required)
	assert.True(t, usernameAttr.ForceNew)
	assert.NotNil(t, usernameAttr.ValidateFunc)

	// Test keys attribute
	keysAttr := resource.Schema["keys"]
	assert.Equal(t, schema.TypeList, keysAttr.Type)
	assert.True(t, keysAttr.Required)
	assert.NotNil(t, keysAttr.Elem)
	elemSchema := keysAttr.Elem.(*schema.Schema)
	assert.Equal(t, schema.TypeString, elemSchema.Type)

	// Test key_count attribute
	keyCountAttr := resource.Schema["key_count"]
	assert.Equal(t, schema.TypeInt, keyCountAttr.Type)
	assert.True(t, keyCountAttr.Computed)
	assert.False(t, keyCountAttr.Required)

	// Verify CRUD functions are set
	assert.NotNil(t, resource.CreateContext)
	assert.NotNil(t, resource.ReadContext)
	assert.NotNil(t, resource.UpdateContext)
	assert.NotNil(t, resource.DeleteContext)

	// Verify import is supported
	assert.NotNil(t, resource.Importer)
	assert.NotNil(t, resource.Importer.StateContext)
}

func TestRTXSSHDAuthorizedKeysResourceDescription(t *testing.T) {
	resource := resourceRTXSSHDAuthorizedKeys()

	// Verify description is set
	assert.NotEmpty(t, resource.Description)
	assert.Contains(t, resource.Description, "SSH authorized keys")

	// Verify attribute descriptions
	assert.NotEmpty(t, resource.Schema["username"].Description)
	assert.NotEmpty(t, resource.Schema["keys"].Description)
	assert.NotEmpty(t, resource.Schema["key_count"].Description)
}

func TestRTXSSHDAuthorizedKeysUsernameValidation(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		expectError bool
	}{
		{
			name:        "valid username - simple",
			username:    "admin",
			expectError: false,
		},
		{
			name:        "valid username - with numbers",
			username:    "user123",
			expectError: false,
		},
		{
			name:        "valid username - with underscore",
			username:    "admin_user",
			expectError: false,
		},
		{
			name:        "valid username - mixed case",
			username:    "AdminUser",
			expectError: false,
		},
		{
			name:        "invalid username - empty",
			username:    "",
			expectError: true,
		},
		{
			name:        "invalid username - starts with number",
			username:    "1admin",
			expectError: true,
		},
		{
			name:        "invalid username - contains special char",
			username:    "admin@user",
			expectError: true,
		},
		{
			name:        "invalid username - contains dash",
			username:    "admin-user",
			expectError: true,
		},
		{
			name:        "invalid username - contains space",
			username:    "admin user",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateUsername(tt.username, "username")

			if tt.expectError {
				assert.NotEmpty(t, errs, "Expected validation error for username: %s", tt.username)
			} else {
				assert.Empty(t, errs, "Expected no validation error for username: %s", tt.username)
			}
		})
	}
}
