package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestBuildPPIPConfigFromResourceData_BasicConfig(t *testing.T) {
	input := map[string]interface{}{
		"pp_number":  1,
		"ip_address": "ipcp",
		"mtu":        1454,
		"tcp_mss":    1414,
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPInterface().Schema, input)
	config := buildPPIPConfigFromResourceData(d)

	assert.Equal(t, "ipcp", config.Address)
	assert.Equal(t, 1454, config.MTU)
	assert.Equal(t, 1414, config.TCPMSSLimit)
}

func TestBuildPPIPConfigFromResourceData_WithNATDescriptor(t *testing.T) {
	input := map[string]interface{}{
		"pp_number":      2,
		"ip_address":     "ipcp",
		"mtu":            1454,
		"tcp_mss":        1414,
		"nat_descriptor": 1000,
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPInterface().Schema, input)
	config := buildPPIPConfigFromResourceData(d)

	assert.Equal(t, "ipcp", config.Address)
	assert.Equal(t, 1000, config.NATDescriptor)
}

func TestBuildPPIPConfigFromResourceData_WithStaticIP(t *testing.T) {
	input := map[string]interface{}{
		"pp_number":  3,
		"ip_address": "203.0.113.100/30",
		"mtu":        1500,
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPInterface().Schema, input)
	config := buildPPIPConfigFromResourceData(d)

	assert.Equal(t, "203.0.113.100/30", config.Address)
	assert.Equal(t, 1500, config.MTU)
}

func TestBuildPPIPConfigFromResourceData_WithAccessLists(t *testing.T) {
	input := map[string]interface{}{
		"pp_number":          1,
		"ip_address":         "ipcp",
		"mtu":                1454,
		"access_list_ip_in":  "pp-secure-in",
		"access_list_ip_out": "pp-secure-out",
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPInterface().Schema, input)
	config := buildPPIPConfigFromResourceData(d)

	assert.Equal(t, "ipcp", config.Address)
	assert.Equal(t, "pp-secure-in", config.AccessListIPIn)
	assert.Equal(t, "pp-secure-out", config.AccessListIPOut)
}

func TestBuildPPIPConfigFromResourceData_DefaultValues(t *testing.T) {
	input := map[string]interface{}{
		"pp_number": 1,
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPInterface().Schema, input)
	config := buildPPIPConfigFromResourceData(d)

	assert.Equal(t, "", config.Address)        // Default is empty
	assert.Equal(t, 0, config.MTU)             // Default is 0
	assert.Equal(t, 0, config.TCPMSSLimit)     // Default is 0
	assert.Equal(t, 0, config.NATDescriptor)   // Default is 0
	assert.Equal(t, "", config.AccessListIPIn) // Default is empty
	assert.Equal(t, "", config.AccessListIPOut)
}

func TestBuildPPIPConfigFromResourceData_FullConfig(t *testing.T) {
	input := map[string]interface{}{
		"pp_number":          1,
		"ip_address":         "ipcp",
		"mtu":                1454,
		"tcp_mss":            1414,
		"nat_descriptor":     1000,
		"access_list_ip_in":  "pp-secure-in",
		"access_list_ip_out": "pp-secure-out",
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPInterface().Schema, input)
	config := buildPPIPConfigFromResourceData(d)

	assert.Equal(t, "ipcp", config.Address)
	assert.Equal(t, 1454, config.MTU)
	assert.Equal(t, 1414, config.TCPMSSLimit)
	assert.Equal(t, 1000, config.NATDescriptor)
	assert.Equal(t, "pp-secure-in", config.AccessListIPIn)
	assert.Equal(t, "pp-secure-out", config.AccessListIPOut)
}

func TestResourceRTXPPInterfaceSchema(t *testing.T) {
	resource := resourceRTXPPInterface()

	// Verify required fields
	assert.NotNil(t, resource.Schema["pp_number"])
	assert.True(t, resource.Schema["pp_number"].Required)
	assert.True(t, resource.Schema["pp_number"].ForceNew)

	// Verify optional fields with defaults
	assert.NotNil(t, resource.Schema["ip_address"])
	assert.True(t, resource.Schema["ip_address"].Optional)
	assert.Equal(t, "", resource.Schema["ip_address"].Default)

	assert.NotNil(t, resource.Schema["mtu"])
	assert.True(t, resource.Schema["mtu"].Optional)
	assert.True(t, resource.Schema["mtu"].Computed) // Changed from Default to Computed for field preservation

	assert.NotNil(t, resource.Schema["tcp_mss"])
	assert.True(t, resource.Schema["tcp_mss"].Optional)
	assert.True(t, resource.Schema["tcp_mss"].Computed) // Changed from Default to Computed for field preservation

	assert.NotNil(t, resource.Schema["nat_descriptor"])
	assert.True(t, resource.Schema["nat_descriptor"].Optional)
	assert.True(t, resource.Schema["nat_descriptor"].Computed) // Changed from Default to Computed for field preservation

	// Access list attributes
	assert.NotNil(t, resource.Schema["access_list_ip_in"])
	assert.True(t, resource.Schema["access_list_ip_in"].Optional)
	assert.Equal(t, schema.TypeString, resource.Schema["access_list_ip_in"].Type)

	assert.NotNil(t, resource.Schema["access_list_ip_out"])
	assert.True(t, resource.Schema["access_list_ip_out"].Optional)
	assert.Equal(t, schema.TypeString, resource.Schema["access_list_ip_out"].Type)
}

func TestBuildPPIPConfigFromResourceData_OnlyInputAccessList(t *testing.T) {
	input := map[string]interface{}{
		"pp_number":         1,
		"ip_address":        "ipcp",
		"access_list_ip_in": "pp-in-acl",
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPInterface().Schema, input)
	config := buildPPIPConfigFromResourceData(d)

	assert.Equal(t, "pp-in-acl", config.AccessListIPIn)
	assert.Equal(t, "", config.AccessListIPOut)
}

func TestBuildPPIPConfigFromResourceData_OnlyOutputAccessList(t *testing.T) {
	input := map[string]interface{}{
		"pp_number":          1,
		"ip_address":         "ipcp",
		"access_list_ip_out": "pp-out-acl",
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPInterface().Schema, input)
	config := buildPPIPConfigFromResourceData(d)

	assert.Equal(t, "", config.AccessListIPIn)
	assert.Equal(t, "pp-out-acl", config.AccessListIPOut)
}
