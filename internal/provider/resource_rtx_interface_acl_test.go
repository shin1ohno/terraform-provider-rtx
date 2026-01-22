package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func TestBuildInterfaceACLFromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.InterfaceACL
	}{
		{
			name: "basic interface ACL with IP access groups",
			input: map[string]interface{}{
				"interface":                "lan1",
				"ip_access_group_in":       "acl-in",
				"ip_access_group_out":      "acl-out",
				"ipv6_access_group_in":     "",
				"ipv6_access_group_out":    "",
				"dynamic_filters_in":       []interface{}{},
				"dynamic_filters_out":      []interface{}{},
				"ipv6_dynamic_filters_in":  []interface{}{},
				"ipv6_dynamic_filters_out": []interface{}{},
			},
			expected: client.InterfaceACL{
				Interface:        "lan1",
				IPAccessGroupIn:  "acl-in",
				IPAccessGroupOut: "acl-out",
			},
		},
		{
			name: "interface ACL with IPv6 access groups",
			input: map[string]interface{}{
				"interface":                "lan2",
				"ip_access_group_in":       "",
				"ip_access_group_out":      "",
				"ipv6_access_group_in":     "acl-v6-in",
				"ipv6_access_group_out":    "acl-v6-out",
				"dynamic_filters_in":       []interface{}{},
				"dynamic_filters_out":      []interface{}{},
				"ipv6_dynamic_filters_in":  []interface{}{},
				"ipv6_dynamic_filters_out": []interface{}{},
			},
			expected: client.InterfaceACL{
				Interface:          "lan2",
				IPv6AccessGroupIn:  "acl-v6-in",
				IPv6AccessGroupOut: "acl-v6-out",
			},
		},
		{
			name: "interface ACL with dynamic filters",
			input: map[string]interface{}{
				"interface":                "pp1",
				"ip_access_group_in":       "",
				"ip_access_group_out":      "",
				"ipv6_access_group_in":     "",
				"ipv6_access_group_out":    "",
				"dynamic_filters_in":       []interface{}{100, 101, 102},
				"dynamic_filters_out":      []interface{}{200, 201},
				"ipv6_dynamic_filters_in":  []interface{}{},
				"ipv6_dynamic_filters_out": []interface{}{},
			},
			expected: client.InterfaceACL{
				Interface:         "pp1",
				DynamicFiltersIn:  []int{100, 101, 102},
				DynamicFiltersOut: []int{200, 201},
			},
		},
		{
			name: "interface ACL with IPv6 dynamic filters",
			input: map[string]interface{}{
				"interface":                "tunnel1",
				"ip_access_group_in":       "",
				"ip_access_group_out":      "",
				"ipv6_access_group_in":     "",
				"ipv6_access_group_out":    "",
				"dynamic_filters_in":       []interface{}{},
				"dynamic_filters_out":      []interface{}{},
				"ipv6_dynamic_filters_in":  []interface{}{300, 301},
				"ipv6_dynamic_filters_out": []interface{}{400},
			},
			expected: client.InterfaceACL{
				Interface:             "tunnel1",
				IPv6DynamicFiltersIn:  []int{300, 301},
				IPv6DynamicFiltersOut: []int{400},
			},
		},
		{
			name: "interface ACL with all options",
			input: map[string]interface{}{
				"interface":                "bridge1",
				"ip_access_group_in":       "full-acl-in",
				"ip_access_group_out":      "full-acl-out",
				"ipv6_access_group_in":     "full-v6-acl-in",
				"ipv6_access_group_out":    "full-v6-acl-out",
				"dynamic_filters_in":       []interface{}{10, 20},
				"dynamic_filters_out":      []interface{}{30},
				"ipv6_dynamic_filters_in":  []interface{}{40},
				"ipv6_dynamic_filters_out": []interface{}{50, 60},
			},
			expected: client.InterfaceACL{
				Interface:             "bridge1",
				IPAccessGroupIn:       "full-acl-in",
				IPAccessGroupOut:      "full-acl-out",
				IPv6AccessGroupIn:     "full-v6-acl-in",
				IPv6AccessGroupOut:    "full-v6-acl-out",
				DynamicFiltersIn:      []int{10, 20},
				DynamicFiltersOut:     []int{30},
				IPv6DynamicFiltersIn:  []int{40},
				IPv6DynamicFiltersOut: []int{50, 60},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceSchema := resourceRTXInterfaceACL().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			result := buildInterfaceACLFromResourceData(d)

			assert.Equal(t, tt.expected.Interface, result.Interface)
			assert.Equal(t, tt.expected.IPAccessGroupIn, result.IPAccessGroupIn)
			assert.Equal(t, tt.expected.IPAccessGroupOut, result.IPAccessGroupOut)
			assert.Equal(t, tt.expected.IPv6AccessGroupIn, result.IPv6AccessGroupIn)
			assert.Equal(t, tt.expected.IPv6AccessGroupOut, result.IPv6AccessGroupOut)
			assert.Equal(t, tt.expected.DynamicFiltersIn, result.DynamicFiltersIn)
			assert.Equal(t, tt.expected.DynamicFiltersOut, result.DynamicFiltersOut)
			assert.Equal(t, tt.expected.IPv6DynamicFiltersIn, result.IPv6DynamicFiltersIn)
			assert.Equal(t, tt.expected.IPv6DynamicFiltersOut, result.IPv6DynamicFiltersOut)
		})
	}
}

func TestResourceRTXInterfaceACLSchema(t *testing.T) {
	resource := resourceRTXInterfaceACL()

	t.Run("interface is required and ForceNew", func(t *testing.T) {
		assert.True(t, resource.Schema["interface"].Required)
		assert.True(t, resource.Schema["interface"].ForceNew)
	})

	t.Run("ip_access_group_in is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["ip_access_group_in"].Optional)
	})

	t.Run("ip_access_group_out is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["ip_access_group_out"].Optional)
	})

	t.Run("ipv6_access_group_in is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["ipv6_access_group_in"].Optional)
	})

	t.Run("ipv6_access_group_out is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["ipv6_access_group_out"].Optional)
	})

	t.Run("dynamic_filters_in is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["dynamic_filters_in"].Optional)
	})

	t.Run("dynamic_filters_out is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["dynamic_filters_out"].Optional)
	})

	t.Run("ipv6_dynamic_filters_in is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["ipv6_dynamic_filters_in"].Optional)
	})

	t.Run("ipv6_dynamic_filters_out is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["ipv6_dynamic_filters_out"].Optional)
	})
}

func TestResourceRTXInterfaceACLSchemaValidation(t *testing.T) {
	resource := resourceRTXInterfaceACL()

	t.Run("interface validation - valid interfaces", func(t *testing.T) {
		validInterfaces := []string{"lan1", "lan2", "pp1", "pp10", "tunnel1", "bridge1", "vlan100"}
		for _, iface := range validInterfaces {
			_, errs := resource.Schema["interface"].ValidateFunc(iface, "interface")
			assert.Empty(t, errs, "interface '%s' should be valid", iface)
		}
	})

	t.Run("interface validation - invalid interfaces", func(t *testing.T) {
		invalidInterfaces := []string{"eth0", "invalid", "WAN", "LAN1"}
		for _, iface := range invalidInterfaces {
			_, errs := resource.Schema["interface"].ValidateFunc(iface, "interface")
			assert.NotEmpty(t, errs, "interface '%s' should be invalid", iface)
		}
	})
}

func TestResourceRTXInterfaceACLImporter(t *testing.T) {
	resource := resourceRTXInterfaceACL()

	t.Run("importer is configured", func(t *testing.T) {
		assert.NotNil(t, resource.Importer)
		assert.NotNil(t, resource.Importer.StateContext)
	})
}

func TestResourceRTXInterfaceACLCRUDFunctions(t *testing.T) {
	resource := resourceRTXInterfaceACL()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, resource.CreateContext)
		assert.NotNil(t, resource.ReadContext)
		assert.NotNil(t, resource.UpdateContext)
		assert.NotNil(t, resource.DeleteContext)
	})
}

func TestInterfaceACLNameRegex(t *testing.T) {
	tests := []struct {
		name     string
		iface    string
		expected bool
	}{
		{"lan1", "lan1", true},
		{"lan10", "lan10", true},
		{"pp1", "pp1", true},
		{"pp99", "pp99", true},
		{"tunnel1", "tunnel1", true},
		{"tunnel100", "tunnel100", true},
		{"bridge1", "bridge1", true},
		{"vlan1", "vlan1", true},
		{"vlan4094", "vlan4094", true},
		{"invalid", "invalid", false},
		{"eth0", "eth0", false},
		{"LAN1", "LAN1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interfaceACLNameRegex.MatchString(tt.iface)
			assert.Equal(t, tt.expected, result)
		})
	}
}
