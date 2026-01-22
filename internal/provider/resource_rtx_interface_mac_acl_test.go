package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func TestBuildInterfaceMACACLFromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.InterfaceMACACL
	}{
		{
			name: "basic MAC ACL with inbound only",
			input: map[string]interface{}{
				"interface":            "lan1",
				"mac_access_group_in":  "mac-acl-in",
				"mac_access_group_out": "",
			},
			expected: client.InterfaceMACACL{
				Interface:        "lan1",
				MACAccessGroupIn: "mac-acl-in",
			},
		},
		{
			name: "MAC ACL with outbound only",
			input: map[string]interface{}{
				"interface":            "lan2",
				"mac_access_group_in":  "",
				"mac_access_group_out": "mac-acl-out",
			},
			expected: client.InterfaceMACACL{
				Interface:         "lan2",
				MACAccessGroupOut: "mac-acl-out",
			},
		},
		{
			name: "MAC ACL with both directions",
			input: map[string]interface{}{
				"interface":            "bridge1",
				"mac_access_group_in":  "mac-filter-in",
				"mac_access_group_out": "mac-filter-out",
			},
			expected: client.InterfaceMACACL{
				Interface:         "bridge1",
				MACAccessGroupIn:  "mac-filter-in",
				MACAccessGroupOut: "mac-filter-out",
			},
		},
		{
			name: "MAC ACL on VLAN interface",
			input: map[string]interface{}{
				"interface":            "vlan100",
				"mac_access_group_in":  "vlan-mac-acl",
				"mac_access_group_out": "",
			},
			expected: client.InterfaceMACACL{
				Interface:        "vlan100",
				MACAccessGroupIn: "vlan-mac-acl",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceSchema := resourceRTXInterfaceMACACL().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			result := buildInterfaceMACACLFromResourceData(d)

			assert.Equal(t, tt.expected.Interface, result.Interface)
			assert.Equal(t, tt.expected.MACAccessGroupIn, result.MACAccessGroupIn)
			assert.Equal(t, tt.expected.MACAccessGroupOut, result.MACAccessGroupOut)
		})
	}
}

func TestResourceRTXInterfaceMACACLSchema(t *testing.T) {
	resource := resourceRTXInterfaceMACACL()

	t.Run("interface is required and ForceNew", func(t *testing.T) {
		assert.True(t, resource.Schema["interface"].Required)
		assert.True(t, resource.Schema["interface"].ForceNew)
	})

	t.Run("mac_access_group_in is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["mac_access_group_in"].Optional)
	})

	t.Run("mac_access_group_out is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["mac_access_group_out"].Optional)
	})
}

func TestResourceRTXInterfaceMACACLSchemaValidation(t *testing.T) {
	resource := resourceRTXInterfaceMACACL()

	t.Run("interface validation - valid interfaces", func(t *testing.T) {
		validInterfaces := []string{"lan1", "lan2", "bridge1", "bridge10", "vlan1", "vlan4094"}
		for _, iface := range validInterfaces {
			_, errs := resource.Schema["interface"].ValidateFunc(iface, "interface")
			assert.Empty(t, errs, "interface '%s' should be valid", iface)
		}
	})

	t.Run("interface validation - invalid interfaces", func(t *testing.T) {
		// MAC ACL only applies to LAN, bridge, and VLAN interfaces (not pp or tunnel)
		invalidInterfaces := []string{"pp1", "tunnel1", "invalid", "eth0", "LAN1"}
		for _, iface := range invalidInterfaces {
			_, errs := resource.Schema["interface"].ValidateFunc(iface, "interface")
			assert.NotEmpty(t, errs, "interface '%s' should be invalid", iface)
		}
	})
}

func TestResourceRTXInterfaceMACACLImporter(t *testing.T) {
	resource := resourceRTXInterfaceMACACL()

	t.Run("importer is configured", func(t *testing.T) {
		assert.NotNil(t, resource.Importer)
		assert.NotNil(t, resource.Importer.StateContext)
	})
}

func TestResourceRTXInterfaceMACACLCRUDFunctions(t *testing.T) {
	resource := resourceRTXInterfaceMACACL()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, resource.CreateContext)
		assert.NotNil(t, resource.ReadContext)
		assert.NotNil(t, resource.UpdateContext)
		assert.NotNil(t, resource.DeleteContext)
	})
}

func TestInterfaceMACACLNameRegex(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		expected bool
	}{
		{"lan1", "lan1", true},
		{"lan10", "lan10", true},
		{"bridge1", "bridge1", true},
		{"bridge10", "bridge10", true},
		{"vlan1", "vlan1", true},
		{"vlan4094", "vlan4094", true},
		{"pp1 not allowed", "pp1", false},
		{"tunnel1 not allowed", "tunnel1", false},
		{"invalid", "invalid", false},
		{"eth0", "eth0", false},
		{"LAN1 case sensitive", "LAN1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interfaceMACACLNameRegex.MatchString(tt.iface)
			assert.Equal(t, tt.expected, result)
		})
	}
}
