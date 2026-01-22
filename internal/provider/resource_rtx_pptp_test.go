package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func TestBuildPPTPConfigFromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.PPTPConfig
	}{
		{
			name: "basic PPTP with authentication",
			input: map[string]interface{}{
				"shutdown":          false,
				"listen_address":    "",
				"max_connections":   0,
				"disconnect_time":   0,
				"keepalive_enabled": false,
				"enabled":           true,
				"authentication": []interface{}{
					map[string]interface{}{
						"method":   "mschap-v2",
						"username": "vpnuser",
						"password": "secret123",
					},
				},
				"encryption": []interface{}{},
				"ip_pool":    []interface{}{},
			},
			expected: client.PPTPConfig{
				Enabled: true,
				Authentication: &client.PPTPAuth{
					Method:   "mschap-v2",
					Username: "vpnuser",
					Password: "secret123",
				},
			},
		},
		{
			name: "PPTP with all options",
			input: map[string]interface{}{
				"shutdown":          false,
				"listen_address":    "192.168.1.1",
				"max_connections":   10,
				"disconnect_time":   300,
				"keepalive_enabled": true,
				"enabled":           true,
				"authentication": []interface{}{
					map[string]interface{}{
						"method":   "chap",
						"username": "admin",
						"password": "adminpass",
					},
				},
				"encryption": []interface{}{
					map[string]interface{}{
						"mppe_bits": 128,
						"required":  true,
					},
				},
				"ip_pool": []interface{}{
					map[string]interface{}{
						"start": "10.0.0.100",
						"end":   "10.0.0.200",
					},
				},
			},
			expected: client.PPTPConfig{
				ListenAddress:    "192.168.1.1",
				MaxConnections:   10,
				DisconnectTime:   300,
				KeepaliveEnabled: true,
				Enabled:          true,
				Authentication: &client.PPTPAuth{
					Method:   "chap",
					Username: "admin",
					Password: "adminpass",
				},
				Encryption: &client.PPTPEncryption{
					MPPEBits: 128,
					Required: true,
				},
				IPPool: &client.PPTPIPPool{
					Start: "10.0.0.100",
					End:   "10.0.0.200",
				},
			},
		},
		{
			name: "PPTP with shutdown",
			input: map[string]interface{}{
				"shutdown":          true,
				"listen_address":    "",
				"max_connections":   5,
				"disconnect_time":   600,
				"keepalive_enabled": false,
				"enabled":           false,
				"authentication": []interface{}{
					map[string]interface{}{
						"method":   "pap",
						"username": "",
						"password": "",
					},
				},
				"encryption": []interface{}{},
				"ip_pool":    []interface{}{},
			},
			expected: client.PPTPConfig{
				Shutdown:       true,
				MaxConnections: 5,
				DisconnectTime: 600,
				Enabled:        false,
				Authentication: &client.PPTPAuth{
					Method: "pap",
				},
			},
		},
		{
			name: "PPTP with 40-bit MPPE encryption",
			input: map[string]interface{}{
				"shutdown":          false,
				"listen_address":    "",
				"max_connections":   0,
				"disconnect_time":   0,
				"keepalive_enabled": false,
				"enabled":           true,
				"authentication": []interface{}{
					map[string]interface{}{
						"method":   "mschap",
						"username": "user",
						"password": "pass",
					},
				},
				"encryption": []interface{}{
					map[string]interface{}{
						"mppe_bits": 40,
						"required":  false,
					},
				},
				"ip_pool": []interface{}{},
			},
			expected: client.PPTPConfig{
				Enabled: true,
				Authentication: &client.PPTPAuth{
					Method:   "mschap",
					Username: "user",
					Password: "pass",
				},
				Encryption: &client.PPTPEncryption{
					MPPEBits: 40,
					Required: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceSchema := resourceRTXPPTP().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			result := buildPPTPConfigFromResourceData(d)

			assert.Equal(t, tt.expected.Shutdown, result.Shutdown)
			assert.Equal(t, tt.expected.ListenAddress, result.ListenAddress)
			assert.Equal(t, tt.expected.MaxConnections, result.MaxConnections)
			assert.Equal(t, tt.expected.DisconnectTime, result.DisconnectTime)
			assert.Equal(t, tt.expected.KeepaliveEnabled, result.KeepaliveEnabled)
			assert.Equal(t, tt.expected.Enabled, result.Enabled)

			// Check authentication
			if tt.expected.Authentication != nil {
				assert.NotNil(t, result.Authentication)
				assert.Equal(t, tt.expected.Authentication.Method, result.Authentication.Method)
				assert.Equal(t, tt.expected.Authentication.Username, result.Authentication.Username)
				assert.Equal(t, tt.expected.Authentication.Password, result.Authentication.Password)
			}

			// Check encryption
			if tt.expected.Encryption != nil {
				assert.NotNil(t, result.Encryption)
				assert.Equal(t, tt.expected.Encryption.MPPEBits, result.Encryption.MPPEBits)
				assert.Equal(t, tt.expected.Encryption.Required, result.Encryption.Required)
			}

			// Check IP pool
			if tt.expected.IPPool != nil {
				assert.NotNil(t, result.IPPool)
				assert.Equal(t, tt.expected.IPPool.Start, result.IPPool.Start)
				assert.Equal(t, tt.expected.IPPool.End, result.IPPool.End)
			}
		})
	}
}

func TestResourceRTXPPTPSchema(t *testing.T) {
	resource := resourceRTXPPTP()

	t.Run("authentication is required with MaxItems 1", func(t *testing.T) {
		assert.True(t, resource.Schema["authentication"].Required)
		assert.Equal(t, 1, resource.Schema["authentication"].MaxItems)
	})

	t.Run("shutdown is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["shutdown"].Optional)
		assert.True(t, resource.Schema["shutdown"].Computed)
	})

	t.Run("listen_address is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["listen_address"].Optional)
		assert.True(t, resource.Schema["listen_address"].Computed)
	})

	t.Run("max_connections is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["max_connections"].Optional)
		assert.True(t, resource.Schema["max_connections"].Computed)
	})

	t.Run("encryption is optional with MaxItems 1", func(t *testing.T) {
		assert.True(t, resource.Schema["encryption"].Optional)
		assert.Equal(t, 1, resource.Schema["encryption"].MaxItems)
	})

	t.Run("ip_pool is optional with MaxItems 1", func(t *testing.T) {
		assert.True(t, resource.Schema["ip_pool"].Optional)
		assert.Equal(t, 1, resource.Schema["ip_pool"].MaxItems)
	})

	t.Run("disconnect_time is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["disconnect_time"].Optional)
		assert.True(t, resource.Schema["disconnect_time"].Computed)
	})

	t.Run("keepalive_enabled is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["keepalive_enabled"].Optional)
		assert.True(t, resource.Schema["keepalive_enabled"].Computed)
	})

	t.Run("enabled is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["enabled"].Optional)
		assert.True(t, resource.Schema["enabled"].Computed)
	})
}

func TestResourceRTXPPTPAuthenticationSchema(t *testing.T) {
	resource := resourceRTXPPTP()
	authSchema := resource.Schema["authentication"].Elem.(*schema.Resource).Schema

	t.Run("method is required", func(t *testing.T) {
		assert.True(t, authSchema["method"].Required)
	})

	t.Run("method validation", func(t *testing.T) {
		validMethods := []string{"pap", "chap", "mschap", "mschap-v2"}
		for _, method := range validMethods {
			_, errs := authSchema["method"].ValidateFunc(method, "method")
			assert.Empty(t, errs, "%s should be valid", method)
		}

		_, errs := authSchema["method"].ValidateFunc("invalid", "method")
		assert.NotEmpty(t, errs, "invalid should be rejected")
	})

	t.Run("password is sensitive", func(t *testing.T) {
		assert.True(t, authSchema["password"].Sensitive)
	})

	t.Run("username is optional", func(t *testing.T) {
		assert.True(t, authSchema["username"].Optional)
	})
}

func TestResourceRTXPPTPEncryptionSchema(t *testing.T) {
	resource := resourceRTXPPTP()
	encSchema := resource.Schema["encryption"].Elem.(*schema.Resource).Schema

	t.Run("mppe_bits is optional and computed", func(t *testing.T) {
		assert.True(t, encSchema["mppe_bits"].Optional)
		assert.True(t, encSchema["mppe_bits"].Computed)
	})

	t.Run("mppe_bits validation", func(t *testing.T) {
		validBits := []int{40, 56, 128}
		for _, bits := range validBits {
			_, errs := encSchema["mppe_bits"].ValidateFunc(bits, "mppe_bits")
			assert.Empty(t, errs, "%d should be valid", bits)
		}

		_, errs := encSchema["mppe_bits"].ValidateFunc(64, "mppe_bits")
		assert.NotEmpty(t, errs, "64 should be invalid")

		_, errs = encSchema["mppe_bits"].ValidateFunc(256, "mppe_bits")
		assert.NotEmpty(t, errs, "256 should be invalid")
	})

	t.Run("required is optional and computed", func(t *testing.T) {
		assert.True(t, encSchema["required"].Optional)
		assert.True(t, encSchema["required"].Computed)
	})
}

func TestResourceRTXPPTPIPPoolSchema(t *testing.T) {
	resource := resourceRTXPPTP()
	poolSchema := resource.Schema["ip_pool"].Elem.(*schema.Resource).Schema

	t.Run("start is required", func(t *testing.T) {
		assert.True(t, poolSchema["start"].Required)
	})

	t.Run("end is required", func(t *testing.T) {
		assert.True(t, poolSchema["end"].Required)
	})
}

func TestResourceRTXPPTPSchemaValidation(t *testing.T) {
	resource := resourceRTXPPTP()

	t.Run("max_connections validation", func(t *testing.T) {
		_, errs := resource.Schema["max_connections"].ValidateFunc(0, "max_connections")
		assert.Empty(t, errs, "0 should be valid (no limit)")

		_, errs = resource.Schema["max_connections"].ValidateFunc(100, "max_connections")
		assert.Empty(t, errs, "100 should be valid")

		_, errs = resource.Schema["max_connections"].ValidateFunc(-1, "max_connections")
		assert.NotEmpty(t, errs, "-1 should be invalid")
	})

	t.Run("disconnect_time validation", func(t *testing.T) {
		_, errs := resource.Schema["disconnect_time"].ValidateFunc(0, "disconnect_time")
		assert.Empty(t, errs, "0 should be valid (no timeout)")

		_, errs = resource.Schema["disconnect_time"].ValidateFunc(300, "disconnect_time")
		assert.Empty(t, errs, "300 should be valid")

		_, errs = resource.Schema["disconnect_time"].ValidateFunc(-1, "disconnect_time")
		assert.NotEmpty(t, errs, "-1 should be invalid")
	})
}

func TestResourceRTXPPTPImporter(t *testing.T) {
	resource := resourceRTXPPTP()

	t.Run("importer is configured", func(t *testing.T) {
		assert.NotNil(t, resource.Importer)
		assert.NotNil(t, resource.Importer.StateContext)
	})
}

func TestResourceRTXPPTPCRUDFunctions(t *testing.T) {
	resource := resourceRTXPPTP()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, resource.CreateContext)
		assert.NotNil(t, resource.ReadContext)
		assert.NotNil(t, resource.UpdateContext)
		assert.NotNil(t, resource.DeleteContext)
	})
}
