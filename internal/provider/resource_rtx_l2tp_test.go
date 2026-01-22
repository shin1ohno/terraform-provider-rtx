package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestBuildL2TPConfigFromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.L2TPConfig
	}{
		{
			name: "basic L2TPv2 LNS tunnel",
			input: map[string]interface{}{
				"tunnel_id":          1,
				"name":               "l2tp-server",
				"version":            "l2tp",
				"mode":               "lns",
				"shutdown":           false,
				"tunnel_source":      "",
				"tunnel_destination": "",
				"tunnel_dest_type":   "",
				"keepalive_enabled":  false,
				"keepalive_interval": 0,
				"keepalive_retry":    0,
				"disconnect_time":    0,
				"always_on":          false,
				"enabled":            true,
				"authentication":     []interface{}{},
				"ip_pool":            []interface{}{},
				"ipsec_profile":      []interface{}{},
				"l2tpv3_config":      []interface{}{},
			},
			expected: client.L2TPConfig{
				ID:      1,
				Name:    "l2tp-server",
				Version: "l2tp",
				Mode:    "lns",
				Enabled: true,
			},
		},
		{
			name: "L2TPv2 with authentication",
			input: map[string]interface{}{
				"tunnel_id":          2,
				"name":               "l2tp-auth",
				"version":            "l2tp",
				"mode":               "lns",
				"shutdown":           false,
				"tunnel_source":      "",
				"tunnel_destination": "",
				"tunnel_dest_type":   "",
				"keepalive_enabled":  true,
				"keepalive_interval": 30,
				"keepalive_retry":    3,
				"disconnect_time":    300,
				"always_on":          false,
				"enabled":            true,
				"authentication": []interface{}{
					map[string]interface{}{
						"method":   "mschap-v2",
						"username": "vpnuser",
						"password": "secret123",
					},
				},
				"ip_pool":       []interface{}{},
				"ipsec_profile": []interface{}{},
				"l2tpv3_config": []interface{}{},
			},
			expected: client.L2TPConfig{
				ID:               2,
				Name:             "l2tp-auth",
				Version:          "l2tp",
				Mode:             "lns",
				KeepaliveEnabled: true,
				DisconnectTime:   300,
				Enabled:          true,
				Authentication: &client.L2TPAuth{
					Method:   "mschap-v2",
					Username: "vpnuser",
					Password: "secret123",
				},
			},
		},
		{
			name: "L2TPv3 site-to-site tunnel",
			input: map[string]interface{}{
				"tunnel_id":          3,
				"name":               "l2tpv3-site",
				"version":            "l2tpv3",
				"mode":               "l2vpn",
				"shutdown":           false,
				"tunnel_source":      "192.168.1.1",
				"tunnel_destination": "10.0.0.1",
				"tunnel_dest_type":   "ip",
				"keepalive_enabled":  true,
				"keepalive_interval": 60,
				"keepalive_retry":    5,
				"disconnect_time":    0,
				"always_on":          true,
				"enabled":            true,
				"authentication":     []interface{}{},
				"ip_pool":            []interface{}{},
				"ipsec_profile":      []interface{}{},
				"l2tpv3_config": []interface{}{
					map[string]interface{}{
						"local_router_id":      "1.1.1.1",
						"remote_router_id":     "2.2.2.2",
						"remote_end_id":        "remote-router",
						"session_id":           100,
						"cookie_size":          8,
						"bridge_interface":     "bridge1",
						"tunnel_auth_enabled":  true,
						"tunnel_auth_password": "authpass",
					},
				},
			},
			expected: client.L2TPConfig{
				ID:               3,
				Name:             "l2tpv3-site",
				Version:          "l2tpv3",
				Mode:             "l2vpn",
				TunnelSource:     "192.168.1.1",
				TunnelDest:       "10.0.0.1",
				TunnelDestType:   "ip",
				KeepaliveEnabled: true,
				AlwaysOn:         true,
				Enabled:          true,
				L2TPv3Config: &client.L2TPv3Config{
					LocalRouterID:   "1.1.1.1",
					RemoteRouterID:  "2.2.2.2",
					RemoteEndID:     "remote-router",
					SessionID:       100,
					CookieSize:      8,
					BridgeInterface: "bridge1",
					TunnelAuth: &client.L2TPTunnelAuth{
						Enabled:  true,
						Password: "authpass",
					},
				},
			},
		},
		{
			name: "L2TP with IP pool and IPsec",
			input: map[string]interface{}{
				"tunnel_id":          4,
				"name":               "l2tp-ipsec",
				"version":            "l2tp",
				"mode":               "lns",
				"shutdown":           false,
				"tunnel_source":      "",
				"tunnel_destination": "",
				"tunnel_dest_type":   "",
				"keepalive_enabled":  false,
				"keepalive_interval": 0,
				"keepalive_retry":    0,
				"disconnect_time":    600,
				"always_on":          false,
				"enabled":            true,
				"authentication":     []interface{}{},
				"ip_pool": []interface{}{
					map[string]interface{}{
						"start": "10.10.10.100",
						"end":   "10.10.10.200",
					},
				},
				"ipsec_profile": []interface{}{
					map[string]interface{}{
						"enabled":        true,
						"pre_shared_key": "ipsec-psk",
						"tunnel_id":      5,
					},
				},
				"l2tpv3_config": []interface{}{},
			},
			expected: client.L2TPConfig{
				ID:             4,
				Name:           "l2tp-ipsec",
				Version:        "l2tp",
				Mode:           "lns",
				DisconnectTime: 600,
				Enabled:        true,
				IPPool: &client.L2TPIPPool{
					Start: "10.10.10.100",
					End:   "10.10.10.200",
				},
				IPsecProfile: &client.L2TPIPsec{
					Enabled:      true,
					PreSharedKey: "ipsec-psk",
					TunnelID:     5,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceSchema := resourceRTXL2TP().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			result := buildL2TPConfigFromResourceData(d)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Version, result.Version)
			assert.Equal(t, tt.expected.Mode, result.Mode)
			assert.Equal(t, tt.expected.TunnelSource, result.TunnelSource)
			assert.Equal(t, tt.expected.TunnelDest, result.TunnelDest)
			assert.Equal(t, tt.expected.KeepaliveEnabled, result.KeepaliveEnabled)
			assert.Equal(t, tt.expected.DisconnectTime, result.DisconnectTime)
			assert.Equal(t, tt.expected.AlwaysOn, result.AlwaysOn)
			assert.Equal(t, tt.expected.Enabled, result.Enabled)

			// Check authentication
			if tt.expected.Authentication != nil {
				assert.NotNil(t, result.Authentication)
				assert.Equal(t, tt.expected.Authentication.Method, result.Authentication.Method)
				assert.Equal(t, tt.expected.Authentication.Username, result.Authentication.Username)
				assert.Equal(t, tt.expected.Authentication.Password, result.Authentication.Password)
			}

			// Check IP pool
			if tt.expected.IPPool != nil {
				assert.NotNil(t, result.IPPool)
				assert.Equal(t, tt.expected.IPPool.Start, result.IPPool.Start)
				assert.Equal(t, tt.expected.IPPool.End, result.IPPool.End)
			}

			// Check IPsec profile
			if tt.expected.IPsecProfile != nil {
				assert.NotNil(t, result.IPsecProfile)
				assert.Equal(t, tt.expected.IPsecProfile.Enabled, result.IPsecProfile.Enabled)
				assert.Equal(t, tt.expected.IPsecProfile.PreSharedKey, result.IPsecProfile.PreSharedKey)
				assert.Equal(t, tt.expected.IPsecProfile.TunnelID, result.IPsecProfile.TunnelID)
			}

			// Check L2TPv3 config
			if tt.expected.L2TPv3Config != nil {
				assert.NotNil(t, result.L2TPv3Config)
				assert.Equal(t, tt.expected.L2TPv3Config.LocalRouterID, result.L2TPv3Config.LocalRouterID)
				assert.Equal(t, tt.expected.L2TPv3Config.RemoteRouterID, result.L2TPv3Config.RemoteRouterID)
				assert.Equal(t, tt.expected.L2TPv3Config.SessionID, result.L2TPv3Config.SessionID)
				assert.Equal(t, tt.expected.L2TPv3Config.CookieSize, result.L2TPv3Config.CookieSize)
				assert.Equal(t, tt.expected.L2TPv3Config.BridgeInterface, result.L2TPv3Config.BridgeInterface)
				if tt.expected.L2TPv3Config.TunnelAuth != nil {
					assert.NotNil(t, result.L2TPv3Config.TunnelAuth)
					assert.Equal(t, tt.expected.L2TPv3Config.TunnelAuth.Enabled, result.L2TPv3Config.TunnelAuth.Enabled)
				}
			}
		})
	}
}

func TestResourceRTXL2TPSchema(t *testing.T) {
	resource := resourceRTXL2TP()

	t.Run("tunnel_id is required and ForceNew", func(t *testing.T) {
		assert.True(t, resource.Schema["tunnel_id"].Required)
		assert.True(t, resource.Schema["tunnel_id"].ForceNew)
	})

	t.Run("version is required and ForceNew", func(t *testing.T) {
		assert.True(t, resource.Schema["version"].Required)
		assert.True(t, resource.Schema["version"].ForceNew)
	})

	t.Run("mode is required and ForceNew", func(t *testing.T) {
		assert.True(t, resource.Schema["mode"].Required)
		assert.True(t, resource.Schema["mode"].ForceNew)
	})

	t.Run("name is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["name"].Optional)
	})

	t.Run("authentication is optional with MaxItems 1", func(t *testing.T) {
		assert.True(t, resource.Schema["authentication"].Optional)
		assert.Equal(t, 1, resource.Schema["authentication"].MaxItems)
	})

	t.Run("ip_pool is optional with MaxItems 1", func(t *testing.T) {
		assert.True(t, resource.Schema["ip_pool"].Optional)
		assert.Equal(t, 1, resource.Schema["ip_pool"].MaxItems)
	})

	t.Run("ipsec_profile is optional with MaxItems 1", func(t *testing.T) {
		assert.True(t, resource.Schema["ipsec_profile"].Optional)
		assert.Equal(t, 1, resource.Schema["ipsec_profile"].MaxItems)
	})

	t.Run("l2tpv3_config is optional with MaxItems 1", func(t *testing.T) {
		assert.True(t, resource.Schema["l2tpv3_config"].Optional)
		assert.Equal(t, 1, resource.Schema["l2tpv3_config"].MaxItems)
	})
}

func TestResourceRTXL2TPSchemaValidation(t *testing.T) {
	resource := resourceRTXL2TP()

	t.Run("tunnel_id validation", func(t *testing.T) {
		_, errs := resource.Schema["tunnel_id"].ValidateFunc(1, "tunnel_id")
		assert.Empty(t, errs, "tunnel_id 1 should be valid")

		_, errs = resource.Schema["tunnel_id"].ValidateFunc(65535, "tunnel_id")
		assert.Empty(t, errs, "tunnel_id 65535 should be valid")

		_, errs = resource.Schema["tunnel_id"].ValidateFunc(0, "tunnel_id")
		assert.NotEmpty(t, errs, "tunnel_id 0 should be invalid")

		_, errs = resource.Schema["tunnel_id"].ValidateFunc(65536, "tunnel_id")
		assert.NotEmpty(t, errs, "tunnel_id 65536 should be invalid")
	})

	t.Run("version validation", func(t *testing.T) {
		_, errs := resource.Schema["version"].ValidateFunc("l2tp", "version")
		assert.Empty(t, errs, "l2tp should be valid")

		_, errs = resource.Schema["version"].ValidateFunc("l2tpv3", "version")
		assert.Empty(t, errs, "l2tpv3 should be valid")

		_, errs = resource.Schema["version"].ValidateFunc("invalid", "version")
		assert.NotEmpty(t, errs, "invalid should be rejected")
	})

	t.Run("mode validation", func(t *testing.T) {
		_, errs := resource.Schema["mode"].ValidateFunc("lns", "mode")
		assert.Empty(t, errs, "lns should be valid")

		_, errs = resource.Schema["mode"].ValidateFunc("l2vpn", "mode")
		assert.Empty(t, errs, "l2vpn should be valid")

		_, errs = resource.Schema["mode"].ValidateFunc("invalid", "mode")
		assert.NotEmpty(t, errs, "invalid should be rejected")
	})
}

func TestResourceRTXL2TPAuthenticationSchema(t *testing.T) {
	resource := resourceRTXL2TP()
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
}

func TestResourceRTXL2TPL2TPv3ConfigSchema(t *testing.T) {
	resource := resourceRTXL2TP()
	l2tpv3Schema := resource.Schema["l2tpv3_config"].Elem.(*schema.Resource).Schema

	t.Run("cookie_size validation", func(t *testing.T) {
		validSizes := []int{0, 4, 8}
		for _, size := range validSizes {
			_, errs := l2tpv3Schema["cookie_size"].ValidateFunc(size, "cookie_size")
			assert.Empty(t, errs, "%d should be valid", size)
		}

		_, errs := l2tpv3Schema["cookie_size"].ValidateFunc(3, "cookie_size")
		assert.NotEmpty(t, errs, "3 should be invalid")
	})

	t.Run("tunnel_auth_password is sensitive", func(t *testing.T) {
		assert.True(t, l2tpv3Schema["tunnel_auth_password"].Sensitive)
	})
}

func TestResourceRTXL2TPImporter(t *testing.T) {
	resource := resourceRTXL2TP()

	t.Run("importer is configured", func(t *testing.T) {
		assert.NotNil(t, resource.Importer)
		assert.NotNil(t, resource.Importer.StateContext)
	})
}

func TestResourceRTXL2TPCRUDFunctions(t *testing.T) {
	resource := resourceRTXL2TP()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, resource.CreateContext)
		assert.NotNil(t, resource.ReadContext)
		assert.NotNil(t, resource.UpdateContext)
		assert.NotNil(t, resource.DeleteContext)
	})
}
