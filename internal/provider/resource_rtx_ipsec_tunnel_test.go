package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func TestBuildIPsecTunnelFromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.IPsecTunnel
	}{
		{
			name: "basic IPsec tunnel",
			input: map[string]interface{}{
				"tunnel_id":       1,
				"name":            "vpn-tunnel-1",
				"local_address":   "192.168.1.1",
				"remote_address":  "10.0.0.1",
				"pre_shared_key":  "secret123",
				"local_network":   "192.168.1.0/24",
				"remote_network":  "10.0.0.0/24",
				"dpd_enabled":     true,
				"dpd_interval":    10,
				"dpd_retry":       3,
				"enabled":         true,
				"ikev2_proposal":  []interface{}{},
				"ipsec_transform": []interface{}{},
			},
			expected: client.IPsecTunnel{
				ID:            1,
				Name:          "vpn-tunnel-1",
				LocalAddress:  "192.168.1.1",
				RemoteAddress: "10.0.0.1",
				PreSharedKey:  "secret123",
				LocalNetwork:  "192.168.1.0/24",
				RemoteNetwork: "10.0.0.0/24",
				DPDEnabled:    true,
				DPDInterval:   10,
				DPDRetry:      3,
				Enabled:       true,
			},
		},
		{
			name: "IPsec tunnel with IKEv2 proposal",
			input: map[string]interface{}{
				"tunnel_id":      2,
				"name":           "vpn-tunnel-2",
				"local_address":  "172.16.0.1",
				"remote_address": "172.16.1.1",
				"pre_shared_key": "key456",
				"local_network":  "",
				"remote_network": "",
				"dpd_enabled":    false,
				"dpd_interval":   0,
				"dpd_retry":      0,
				"enabled":        true,
				"ikev2_proposal": []interface{}{
					map[string]interface{}{
						"encryption_aes256": true,
						"encryption_aes128": false,
						"encryption_3des":   false,
						"integrity_sha256":  true,
						"integrity_sha1":    false,
						"integrity_md5":     false,
						"group_fourteen":    true,
						"group_five":        false,
						"group_two":         false,
						"lifetime_seconds":  86400,
					},
				},
				"ipsec_transform": []interface{}{},
			},
			expected: client.IPsecTunnel{
				ID:            2,
				Name:          "vpn-tunnel-2",
				LocalAddress:  "172.16.0.1",
				RemoteAddress: "172.16.1.1",
				PreSharedKey:  "key456",
				Enabled:       true,
				IKEv2Proposal: client.IKEv2Proposal{
					EncryptionAES256: true,
					IntegritySHA256:  true,
					GroupFourteen:    true,
					LifetimeSeconds:  86400,
				},
			},
		},
		{
			name: "IPsec tunnel with IPsec transform",
			input: map[string]interface{}{
				"tunnel_id":      3,
				"name":           "vpn-tunnel-3",
				"local_address":  "10.10.0.1",
				"remote_address": "10.20.0.1",
				"pre_shared_key": "key789",
				"local_network":  "10.10.0.0/24",
				"remote_network": "10.20.0.0/24",
				"dpd_enabled":    true,
				"dpd_interval":   15,
				"dpd_retry":      5,
				"enabled":        true,
				"ikev2_proposal": []interface{}{},
				"ipsec_transform": []interface{}{
					map[string]interface{}{
						"protocol":           "esp",
						"encryption_aes256":  true,
						"encryption_aes128":  false,
						"encryption_3des":    false,
						"integrity_sha256":   true,
						"integrity_sha1":     false,
						"integrity_md5":      false,
						"pfs_group_fourteen": true,
						"pfs_group_five":     false,
						"pfs_group_two":      false,
						"lifetime_seconds":   3600,
					},
				},
			},
			expected: client.IPsecTunnel{
				ID:            3,
				Name:          "vpn-tunnel-3",
				LocalAddress:  "10.10.0.1",
				RemoteAddress: "10.20.0.1",
				PreSharedKey:  "key789",
				LocalNetwork:  "10.10.0.0/24",
				RemoteNetwork: "10.20.0.0/24",
				DPDEnabled:    true,
				DPDInterval:   15,
				DPDRetry:      5,
				Enabled:       true,
				IPsecTransform: client.IPsecTransform{
					Protocol:         "esp",
					EncryptionAES256: true,
					IntegritySHA256:  true,
					PFSGroupFourteen: true,
					LifetimeSeconds:  3600,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceSchema := resourceRTXIPsecTunnel().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			result := buildIPsecTunnelFromResourceData(d)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.LocalAddress, result.LocalAddress)
			assert.Equal(t, tt.expected.RemoteAddress, result.RemoteAddress)
			assert.Equal(t, tt.expected.PreSharedKey, result.PreSharedKey)
			assert.Equal(t, tt.expected.LocalNetwork, result.LocalNetwork)
			assert.Equal(t, tt.expected.RemoteNetwork, result.RemoteNetwork)
			assert.Equal(t, tt.expected.DPDEnabled, result.DPDEnabled)
			assert.Equal(t, tt.expected.DPDInterval, result.DPDInterval)
			assert.Equal(t, tt.expected.DPDRetry, result.DPDRetry)
			assert.Equal(t, tt.expected.Enabled, result.Enabled)

			// IKEv2 Proposal
			assert.Equal(t, tt.expected.IKEv2Proposal.EncryptionAES256, result.IKEv2Proposal.EncryptionAES256)
			assert.Equal(t, tt.expected.IKEv2Proposal.IntegritySHA256, result.IKEv2Proposal.IntegritySHA256)
			assert.Equal(t, tt.expected.IKEv2Proposal.GroupFourteen, result.IKEv2Proposal.GroupFourteen)
			assert.Equal(t, tt.expected.IKEv2Proposal.LifetimeSeconds, result.IKEv2Proposal.LifetimeSeconds)

			// IPsec Transform
			assert.Equal(t, tt.expected.IPsecTransform.Protocol, result.IPsecTransform.Protocol)
			assert.Equal(t, tt.expected.IPsecTransform.EncryptionAES256, result.IPsecTransform.EncryptionAES256)
			assert.Equal(t, tt.expected.IPsecTransform.IntegritySHA256, result.IPsecTransform.IntegritySHA256)
			assert.Equal(t, tt.expected.IPsecTransform.PFSGroupFourteen, result.IPsecTransform.PFSGroupFourteen)
			assert.Equal(t, tt.expected.IPsecTransform.LifetimeSeconds, result.IPsecTransform.LifetimeSeconds)
		})
	}
}

func TestResourceRTXIPsecTunnelSchema(t *testing.T) {
	resource := resourceRTXIPsecTunnel()

	t.Run("tunnel_id is required and ForceNew", func(t *testing.T) {
		assert.True(t, resource.Schema["tunnel_id"].Required)
		assert.True(t, resource.Schema["tunnel_id"].ForceNew)
	})

	t.Run("name is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["name"].Optional)
	})

	t.Run("local_address is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["local_address"].Optional)
	})

	t.Run("remote_address is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["remote_address"].Optional)
	})

	t.Run("pre_shared_key is optional and sensitive", func(t *testing.T) {
		assert.True(t, resource.Schema["pre_shared_key"].Optional)
		assert.True(t, resource.Schema["pre_shared_key"].Sensitive)
	})

	t.Run("ikev2_proposal is optional with MaxItems 1", func(t *testing.T) {
		assert.True(t, resource.Schema["ikev2_proposal"].Optional)
		assert.Equal(t, 1, resource.Schema["ikev2_proposal"].MaxItems)
	})

	t.Run("ipsec_transform is optional with MaxItems 1", func(t *testing.T) {
		assert.True(t, resource.Schema["ipsec_transform"].Optional)
		assert.Equal(t, 1, resource.Schema["ipsec_transform"].MaxItems)
	})

	t.Run("dpd_enabled is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["dpd_enabled"].Optional)
		assert.True(t, resource.Schema["dpd_enabled"].Computed)
	})

	t.Run("enabled is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["enabled"].Optional)
		assert.True(t, resource.Schema["enabled"].Computed)
	})
}

func TestResourceRTXIPsecTunnelSchemaValidation(t *testing.T) {
	resource := resourceRTXIPsecTunnel()

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

	t.Run("local_address validation", func(t *testing.T) {
		_, errs := resource.Schema["local_address"].ValidateFunc("192.168.1.1", "local_address")
		assert.Empty(t, errs, "valid IP should be accepted")

		_, errs = resource.Schema["local_address"].ValidateFunc("", "local_address")
		assert.Empty(t, errs, "empty should be accepted (optional)")

		_, errs = resource.Schema["local_address"].ValidateFunc("invalid", "local_address")
		assert.NotEmpty(t, errs, "invalid IP should be rejected")
	})
}

func TestResourceRTXIPsecTunnelIKEv2ProposalSchema(t *testing.T) {
	resource := resourceRTXIPsecTunnel()
	ikev2Schema := resource.Schema["ikev2_proposal"].Elem.(*schema.Resource).Schema

	t.Run("encryption options are optional and computed", func(t *testing.T) {
		assert.True(t, ikev2Schema["encryption_aes256"].Optional)
		assert.True(t, ikev2Schema["encryption_aes256"].Computed)
		assert.True(t, ikev2Schema["encryption_aes128"].Optional)
		assert.True(t, ikev2Schema["encryption_aes128"].Computed)
		assert.True(t, ikev2Schema["encryption_3des"].Optional)
		assert.True(t, ikev2Schema["encryption_3des"].Computed)
	})

	t.Run("integrity options are optional and computed", func(t *testing.T) {
		assert.True(t, ikev2Schema["integrity_sha256"].Optional)
		assert.True(t, ikev2Schema["integrity_sha256"].Computed)
		assert.True(t, ikev2Schema["integrity_sha1"].Optional)
		assert.True(t, ikev2Schema["integrity_sha1"].Computed)
		assert.True(t, ikev2Schema["integrity_md5"].Optional)
		assert.True(t, ikev2Schema["integrity_md5"].Computed)
	})

	t.Run("DH group options are optional and computed", func(t *testing.T) {
		assert.True(t, ikev2Schema["group_fourteen"].Optional)
		assert.True(t, ikev2Schema["group_fourteen"].Computed)
		assert.True(t, ikev2Schema["group_five"].Optional)
		assert.True(t, ikev2Schema["group_five"].Computed)
		assert.True(t, ikev2Schema["group_two"].Optional)
		assert.True(t, ikev2Schema["group_two"].Computed)
	})

	t.Run("lifetime_seconds validation", func(t *testing.T) {
		_, errs := ikev2Schema["lifetime_seconds"].ValidateFunc(60, "lifetime_seconds")
		assert.Empty(t, errs, "60 should be valid")

		_, errs = ikev2Schema["lifetime_seconds"].ValidateFunc(59, "lifetime_seconds")
		assert.NotEmpty(t, errs, "59 should be invalid")
	})
}

func TestResourceRTXIPsecTunnelIPsecTransformSchema(t *testing.T) {
	resource := resourceRTXIPsecTunnel()
	transformSchema := resource.Schema["ipsec_transform"].Elem.(*schema.Resource).Schema

	t.Run("protocol validation", func(t *testing.T) {
		_, errs := transformSchema["protocol"].ValidateFunc("esp", "protocol")
		assert.Empty(t, errs, "esp should be valid")

		_, errs = transformSchema["protocol"].ValidateFunc("ah", "protocol")
		assert.Empty(t, errs, "ah should be valid")

		_, errs = transformSchema["protocol"].ValidateFunc("invalid", "protocol")
		assert.NotEmpty(t, errs, "invalid should be rejected")
	})

	t.Run("PFS group options are optional and computed", func(t *testing.T) {
		assert.True(t, transformSchema["pfs_group_fourteen"].Optional)
		assert.True(t, transformSchema["pfs_group_fourteen"].Computed)
		assert.True(t, transformSchema["pfs_group_five"].Optional)
		assert.True(t, transformSchema["pfs_group_five"].Computed)
		assert.True(t, transformSchema["pfs_group_two"].Optional)
		assert.True(t, transformSchema["pfs_group_two"].Computed)
	})
}

func TestResourceRTXIPsecTunnelImporter(t *testing.T) {
	resource := resourceRTXIPsecTunnel()

	t.Run("importer is configured", func(t *testing.T) {
		assert.NotNil(t, resource.Importer)
		assert.NotNil(t, resource.Importer.StateContext)
	})
}

func TestResourceRTXIPsecTunnelCRUDFunctions(t *testing.T) {
	resource := resourceRTXIPsecTunnel()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, resource.CreateContext)
		assert.NotNil(t, resource.ReadContext)
		assert.NotNil(t, resource.UpdateContext)
		assert.NotNil(t, resource.DeleteContext)
	})
}
