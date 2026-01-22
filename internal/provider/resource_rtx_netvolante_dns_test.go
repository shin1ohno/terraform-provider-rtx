package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestBuildNetVolanteConfigFromResourceData_BasicConfig(t *testing.T) {
	input := map[string]interface{}{
		"interface":    "pp 1",
		"hostname":     "myrouter.aa0.netvolante.jp",
		"server":       1,
		"timeout":      60,
		"ipv6_enabled": false,
	}

	d := schema.TestResourceDataRaw(t, resourceRTXNetVolanteDNS().Schema, input)
	config := buildNetVolanteConfigFromResourceData(d)

	assert.Equal(t, "pp 1", config.Interface)
	assert.Equal(t, "myrouter.aa0.netvolante.jp", config.Hostname)
	assert.Equal(t, 1, config.Server)
	assert.Equal(t, 60, config.Timeout)
	assert.False(t, config.IPv6)
	assert.True(t, config.Use) // Should always be true when building config
}

func TestBuildNetVolanteConfigFromResourceData_WithIPv6(t *testing.T) {
	input := map[string]interface{}{
		"interface":    "lan2",
		"hostname":     "ipv6router.bb1.netvolante.jp",
		"server":       2,
		"timeout":      120,
		"ipv6_enabled": true,
	}

	d := schema.TestResourceDataRaw(t, resourceRTXNetVolanteDNS().Schema, input)
	config := buildNetVolanteConfigFromResourceData(d)

	assert.Equal(t, "lan2", config.Interface)
	assert.Equal(t, "ipv6router.bb1.netvolante.jp", config.Hostname)
	assert.Equal(t, 2, config.Server)
	assert.Equal(t, 120, config.Timeout)
	assert.True(t, config.IPv6)
}

func TestBuildNetVolanteConfigFromResourceData_DefaultValues(t *testing.T) {
	input := map[string]interface{}{
		"interface": "pp 1",
		"hostname":  "test.aa0.netvolante.jp",
	}

	d := schema.TestResourceDataRaw(t, resourceRTXNetVolanteDNS().Schema, input)
	config := buildNetVolanteConfigFromResourceData(d)

	assert.Equal(t, "pp 1", config.Interface)
	assert.Equal(t, "test.aa0.netvolante.jp", config.Hostname)
	assert.Equal(t, 1, config.Server)   // Default is 1
	assert.Equal(t, 60, config.Timeout) // Default is 60
	assert.False(t, config.IPv6)        // Default is false
}

func TestBuildNetVolanteConfigFromResourceData_WithAutoHostname(t *testing.T) {
	input := map[string]interface{}{
		"interface":     "lan2",
		"hostname":      "auto.aa0.netvolante.jp",
		"server":        1,
		"timeout":       60,
		"ipv6_enabled":  false,
		"auto_hostname": true,
	}

	d := schema.TestResourceDataRaw(t, resourceRTXNetVolanteDNS().Schema, input)
	config := buildNetVolanteConfigFromResourceData(d)

	assert.Equal(t, "lan2", config.Interface)
	assert.True(t, config.AutoHostname)
}

func TestResourceRTXNetVolanteDNSSchema(t *testing.T) {
	resource := resourceRTXNetVolanteDNS()

	// Verify required fields
	assert.NotNil(t, resource.Schema["interface"])
	assert.True(t, resource.Schema["interface"].Required)
	assert.True(t, resource.Schema["interface"].ForceNew)

	assert.NotNil(t, resource.Schema["hostname"])
	assert.True(t, resource.Schema["hostname"].Required)

	// Verify optional fields with defaults
	assert.NotNil(t, resource.Schema["server"])
	assert.True(t, resource.Schema["server"].Optional)
	assert.Equal(t, 1, resource.Schema["server"].Default)

	assert.NotNil(t, resource.Schema["timeout"])
	assert.True(t, resource.Schema["timeout"].Optional)
	assert.Equal(t, 60, resource.Schema["timeout"].Default)

	assert.NotNil(t, resource.Schema["ipv6_enabled"])
	assert.True(t, resource.Schema["ipv6_enabled"].Optional)
	assert.Equal(t, false, resource.Schema["ipv6_enabled"].Default)

	// Verify computed fields
	assert.NotNil(t, resource.Schema["auto_hostname"])
	assert.True(t, resource.Schema["auto_hostname"].Optional)
	assert.True(t, resource.Schema["auto_hostname"].Computed)
}

func TestBuildNetVolanteConfigFromResourceData_DifferentInterfaces(t *testing.T) {
	interfaces := []string{"pp 1", "pp 2", "lan1", "lan2", "lan3"}

	for _, iface := range interfaces {
		input := map[string]interface{}{
			"interface": iface,
			"hostname":  "test.aa0.netvolante.jp",
		}

		d := schema.TestResourceDataRaw(t, resourceRTXNetVolanteDNS().Schema, input)
		config := buildNetVolanteConfigFromResourceData(d)

		assert.Equal(t, iface, config.Interface)
	}
}

func TestBuildNetVolanteConfigFromResourceData_BothServers(t *testing.T) {
	// Test both valid server numbers
	for server := 1; server <= 2; server++ {
		input := map[string]interface{}{
			"interface": "pp 1",
			"hostname":  "test.aa0.netvolante.jp",
			"server":    server,
		}

		d := schema.TestResourceDataRaw(t, resourceRTXNetVolanteDNS().Schema, input)
		config := buildNetVolanteConfigFromResourceData(d)

		assert.Equal(t, server, config.Server)
	}
}
