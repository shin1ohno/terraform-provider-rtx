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

func TestBuildPPIPConfigFromResourceData_WithSecurityFilters(t *testing.T) {
	input := map[string]interface{}{
		"pp_number":         1,
		"ip_address":        "ipcp",
		"mtu":               1454,
		"secure_filter_in":  []interface{}{200020, 200021, 200022, 200099},
		"secure_filter_out": []interface{}{200020, 200021, 200022, 200099},
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPInterface().Schema, input)
	config := buildPPIPConfigFromResourceData(d)

	assert.Equal(t, "ipcp", config.Address)
	assert.Len(t, config.SecureFilterIn, 4)
	assert.Equal(t, 200020, config.SecureFilterIn[0])
	assert.Equal(t, 200021, config.SecureFilterIn[1])
	assert.Equal(t, 200022, config.SecureFilterIn[2])
	assert.Equal(t, 200099, config.SecureFilterIn[3])

	assert.Len(t, config.SecureFilterOut, 4)
	assert.Equal(t, 200020, config.SecureFilterOut[0])
	assert.Equal(t, 200099, config.SecureFilterOut[3])
}

func TestBuildPPIPConfigFromResourceData_DefaultValues(t *testing.T) {
	input := map[string]interface{}{
		"pp_number": 1,
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPInterface().Schema, input)
	config := buildPPIPConfigFromResourceData(d)

	assert.Equal(t, "", config.Address)     // Default is empty
	assert.Equal(t, 0, config.MTU)          // Default is 0
	assert.Equal(t, 0, config.TCPMSSLimit)  // Default is 0
	assert.Equal(t, 0, config.NATDescriptor) // Default is 0
	assert.Nil(t, config.SecureFilterIn)
	assert.Nil(t, config.SecureFilterOut)
}

func TestBuildPPIPConfigFromResourceData_FullConfig(t *testing.T) {
	input := map[string]interface{}{
		"pp_number":         1,
		"ip_address":        "ipcp",
		"mtu":               1454,
		"tcp_mss":           1414,
		"nat_descriptor":    1000,
		"secure_filter_in":  []interface{}{200020, 200021, 200022},
		"secure_filter_out": []interface{}{200030, 200031},
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPInterface().Schema, input)
	config := buildPPIPConfigFromResourceData(d)

	assert.Equal(t, "ipcp", config.Address)
	assert.Equal(t, 1454, config.MTU)
	assert.Equal(t, 1414, config.TCPMSSLimit)
	assert.Equal(t, 1000, config.NATDescriptor)
	assert.Len(t, config.SecureFilterIn, 3)
	assert.Len(t, config.SecureFilterOut, 2)
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
	assert.Equal(t, 0, resource.Schema["mtu"].Default)

	assert.NotNil(t, resource.Schema["tcp_mss"])
	assert.True(t, resource.Schema["tcp_mss"].Optional)
	assert.Equal(t, 0, resource.Schema["tcp_mss"].Default)

	assert.NotNil(t, resource.Schema["nat_descriptor"])
	assert.True(t, resource.Schema["nat_descriptor"].Optional)
	assert.Equal(t, 0, resource.Schema["nat_descriptor"].Default)

	assert.NotNil(t, resource.Schema["secure_filter_in"])
	assert.True(t, resource.Schema["secure_filter_in"].Optional)
	assert.Equal(t, schema.TypeList, resource.Schema["secure_filter_in"].Type)

	assert.NotNil(t, resource.Schema["secure_filter_out"])
	assert.True(t, resource.Schema["secure_filter_out"].Optional)
	assert.Equal(t, schema.TypeList, resource.Schema["secure_filter_out"].Type)
}

func TestBuildPPIPConfigFromResourceData_OnlyInputFilters(t *testing.T) {
	input := map[string]interface{}{
		"pp_number":        1,
		"ip_address":       "ipcp",
		"secure_filter_in": []interface{}{100, 101, 102},
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPInterface().Schema, input)
	config := buildPPIPConfigFromResourceData(d)

	assert.Len(t, config.SecureFilterIn, 3)
	assert.Nil(t, config.SecureFilterOut)
}

func TestBuildPPIPConfigFromResourceData_OnlyOutputFilters(t *testing.T) {
	input := map[string]interface{}{
		"pp_number":         1,
		"ip_address":        "ipcp",
		"secure_filter_out": []interface{}{200, 201},
	}

	d := schema.TestResourceDataRaw(t, resourceRTXPPInterface().Schema, input)
	config := buildPPIPConfigFromResourceData(d)

	assert.Nil(t, config.SecureFilterIn)
	assert.Len(t, config.SecureFilterOut, 2)
}
