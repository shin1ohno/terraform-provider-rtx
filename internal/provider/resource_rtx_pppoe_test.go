package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestBuildPPPoEConfigFromResourceData_BasicConfig(t *testing.T) {
	input := map[string]interface{}{
		"pp_number":      1,
		"name":           "NTT FLET'S NGN",
		"bind_interface": "lan2",
		"username":       "user@example.ne.jp",
		"password":       "secret123",
		"auth_method":    "chap",
		"always_on":      true,
		"enabled":        true,
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPPoE().Schema, input)
	config := buildPPPoEConfigFromResourceData(d)

	assert.Equal(t, 1, config.Number)
	assert.Equal(t, "NTT FLET'S NGN", config.Name)
	assert.Equal(t, "lan2", config.BindInterface)
	assert.True(t, config.AlwaysOn)
	assert.True(t, config.Enabled)

	// Check authentication
	assert.NotNil(t, config.Authentication)
	assert.Equal(t, "chap", config.Authentication.Method)
	assert.Equal(t, "user@example.ne.jp", config.Authentication.Username)
	assert.Equal(t, "secret123", config.Authentication.Password)
}

func TestBuildPPPoEConfigFromResourceData_WithServiceName(t *testing.T) {
	input := map[string]interface{}{
		"pp_number":      2,
		"bind_interface": "lan3",
		"username":       "user@isp.jp",
		"password":       "pass",
		"service_name":   "INTERNET",
		"auth_method":    "chap",
		"always_on":      true,
		"enabled":        true,
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPPoE().Schema, input)
	config := buildPPPoEConfigFromResourceData(d)

	assert.Equal(t, 2, config.Number)
	assert.Equal(t, "INTERNET", config.ServiceName)
}

func TestBuildPPPoEConfigFromResourceData_WithACName(t *testing.T) {
	input := map[string]interface{}{
		"pp_number":      3,
		"bind_interface": "lan2",
		"username":       "user@provider.ne.jp",
		"password":       "password",
		"ac_name":        "ACCESS_CONCENTRATOR",
		"auth_method":    "pap",
		"always_on":      false,
		"enabled":        true,
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPPoE().Schema, input)
	config := buildPPPoEConfigFromResourceData(d)

	assert.Equal(t, 3, config.Number)
	assert.Equal(t, "ACCESS_CONCENTRATOR", config.ACName)
	assert.Equal(t, "pap", config.Authentication.Method)
	assert.False(t, config.AlwaysOn)
}

func TestBuildPPPoEConfigFromResourceData_WithDisconnectTimeout(t *testing.T) {
	input := map[string]interface{}{
		"pp_number":          4,
		"bind_interface":     "lan2",
		"username":           "backup@isp.jp",
		"password":           "backup_pass",
		"auth_method":        "chap",
		"always_on":          false,
		"disconnect_timeout": 300,
		"enabled":            true,
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPPoE().Schema, input)
	config := buildPPPoEConfigFromResourceData(d)

	assert.Equal(t, 4, config.Number)
	assert.False(t, config.AlwaysOn)
	assert.Equal(t, 300, config.DisconnectTimeout)
}

func TestBuildPPPoEConfigFromResourceData_DefaultValues(t *testing.T) {
	input := map[string]interface{}{
		"pp_number":      1,
		"bind_interface": "lan2",
		"username":       "user",
		"password":       "pass",
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPPoE().Schema, input)
	config := buildPPPoEConfigFromResourceData(d)

	assert.Equal(t, "chap", config.Authentication.Method) // Default auth method
	assert.True(t, config.AlwaysOn)                        // Default is true
	assert.True(t, config.Enabled)                         // Default is true
	assert.Equal(t, 0, config.DisconnectTimeout)           // Default is 0
}

func TestBuildPPPoEConfigFromResourceData_AllAuthMethods(t *testing.T) {
	authMethods := []string{"pap", "chap", "mschap", "mschap-v2"}

	for _, method := range authMethods {
		input := map[string]interface{}{
			"pp_number":      1,
			"bind_interface": "lan2",
			"username":       "user",
			"password":       "pass",
			"auth_method":    method,
			"always_on":      true,
			"enabled":        true,
		}

		d := schema.TestResourceDataRaw(t, resourceRTXPPPoE().Schema, input)
		config := buildPPPoEConfigFromResourceData(d)

		assert.Equal(t, method, config.Authentication.Method)
	}
}

func TestResourceRTXPPPoESchema(t *testing.T) {
	resource := resourceRTXPPPoE()

	// Verify required fields
	assert.NotNil(t, resource.Schema["pp_number"])
	assert.True(t, resource.Schema["pp_number"].Required)
	assert.True(t, resource.Schema["pp_number"].ForceNew)

	assert.NotNil(t, resource.Schema["bind_interface"])
	assert.True(t, resource.Schema["bind_interface"].Required)

	assert.NotNil(t, resource.Schema["username"])
	assert.True(t, resource.Schema["username"].Required)

	assert.NotNil(t, resource.Schema["password"])
	assert.True(t, resource.Schema["password"].Required)
	assert.True(t, resource.Schema["password"].Sensitive)

	// Verify optional fields
	assert.NotNil(t, resource.Schema["name"])
	assert.True(t, resource.Schema["name"].Optional)

	assert.NotNil(t, resource.Schema["service_name"])
	assert.True(t, resource.Schema["service_name"].Optional)

	assert.NotNil(t, resource.Schema["ac_name"])
	assert.True(t, resource.Schema["ac_name"].Optional)

	assert.NotNil(t, resource.Schema["auth_method"])
	assert.True(t, resource.Schema["auth_method"].Optional)
	assert.Equal(t, "chap", resource.Schema["auth_method"].Default)

	assert.NotNil(t, resource.Schema["always_on"])
	assert.True(t, resource.Schema["always_on"].Optional)
	assert.Equal(t, true, resource.Schema["always_on"].Default)

	assert.NotNil(t, resource.Schema["disconnect_timeout"])
	assert.True(t, resource.Schema["disconnect_timeout"].Optional)
	assert.Equal(t, 0, resource.Schema["disconnect_timeout"].Default)

	assert.NotNil(t, resource.Schema["enabled"])
	assert.True(t, resource.Schema["enabled"].Optional)
	assert.Equal(t, true, resource.Schema["enabled"].Default)
}
