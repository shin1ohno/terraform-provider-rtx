package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func TestBuildIPFilterFromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.IPFilter
	}{
		{
			name: "basic filter with all required fields",
			input: map[string]interface{}{
				"filter_id":   200000,
				"action":      "reject",
				"source":      "10.0.0.0/8",
				"destination": "*",
				"protocol":    "tcp",
				"source_port": "*",
				"dest_port":   "135",
				"established": false,
			},
			expected: client.IPFilter{
				Number:        200000,
				Action:        "reject",
				SourceAddress: "10.0.0.0/8",
				DestAddress:   "*",
				Protocol:      "tcp",
				SourcePort:    "*",
				DestPort:      "135",
				Established:   false,
			},
		},
		{
			name: "filter with established TCP",
			input: map[string]interface{}{
				"filter_id":   100,
				"action":      "pass",
				"source":      "*",
				"destination": "192.168.1.0/24",
				"protocol":    "tcp",
				"source_port": "*",
				"dest_port":   "*",
				"established": true,
			},
			expected: client.IPFilter{
				Number:        100,
				Action:        "pass",
				SourceAddress: "*",
				DestAddress:   "192.168.1.0/24",
				Protocol:      "tcp",
				SourcePort:    "*",
				DestPort:      "*",
				Established:   true,
			},
		},
		{
			name: "filter with UDP and port range",
			input: map[string]interface{}{
				"filter_id":   300,
				"action":      "restrict",
				"source":      "172.16.0.0/12",
				"destination": "*",
				"protocol":    "udp",
				"source_port": "1024-65535",
				"dest_port":   "53",
				"established": false,
			},
			expected: client.IPFilter{
				Number:        300,
				Action:        "restrict",
				SourceAddress: "172.16.0.0/12",
				DestAddress:   "*",
				Protocol:      "udp",
				SourcePort:    "1024-65535",
				DestPort:      "53",
				Established:   false,
			},
		},
		{
			name: "filter with ICMP protocol",
			input: map[string]interface{}{
				"filter_id":   400,
				"action":      "pass",
				"source":      "*",
				"destination": "*",
				"protocol":    "icmp",
				"source_port": "*",
				"dest_port":   "*",
				"established": false,
			},
			expected: client.IPFilter{
				Number:        400,
				Action:        "pass",
				SourceAddress: "*",
				DestAddress:   "*",
				Protocol:      "icmp",
				SourcePort:    "*",
				DestPort:      "*",
				Established:   false,
			},
		},
		{
			name: "filter with any protocol",
			input: map[string]interface{}{
				"filter_id":   500,
				"action":      "reject",
				"source":      "0.0.0.0/0",
				"destination": "*",
				"protocol":    "*",
				"source_port": "*",
				"dest_port":   "*",
				"established": false,
			},
			expected: client.IPFilter{
				Number:        500,
				Action:        "reject",
				SourceAddress: "0.0.0.0/0",
				DestAddress:   "*",
				Protocol:      "*",
				SourcePort:    "*",
				DestPort:      "*",
				Established:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a resource data mock
			resourceSchema := resourceRTXAccessListIP().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			result := buildIPFilterFromResourceData(d)

			assert.Equal(t, tt.expected.Number, result.Number)
			assert.Equal(t, tt.expected.Action, result.Action)
			assert.Equal(t, tt.expected.SourceAddress, result.SourceAddress)
			assert.Equal(t, tt.expected.DestAddress, result.DestAddress)
			assert.Equal(t, tt.expected.Protocol, result.Protocol)
			assert.Equal(t, tt.expected.SourcePort, result.SourcePort)
			assert.Equal(t, tt.expected.DestPort, result.DestPort)
			assert.Equal(t, tt.expected.Established, result.Established)
		})
	}
}

func TestFlattenIPFilterToResourceData(t *testing.T) {
	tests := []struct {
		name     string
		filter   *client.IPFilter
		expected map[string]interface{}
	}{
		{
			name: "basic filter flattening",
			filter: &client.IPFilter{
				Number:        200000,
				Action:        "reject",
				SourceAddress: "10.0.0.0/8",
				DestAddress:   "*",
				Protocol:      "tcp",
				SourcePort:    "*",
				DestPort:      "135",
				Established:   false,
			},
			expected: map[string]interface{}{
				"filter_id":   200000,
				"action":      "reject",
				"source":      "10.0.0.0/8",
				"destination": "*",
				"protocol":    "tcp",
				"source_port": "*",
				"dest_port":   "135",
				"established": false,
			},
		},
		{
			name: "filter with established TCP",
			filter: &client.IPFilter{
				Number:        100,
				Action:        "pass",
				SourceAddress: "*",
				DestAddress:   "192.168.1.0/24",
				Protocol:      "tcp",
				SourcePort:    "",
				DestPort:      "",
				Established:   true,
			},
			expected: map[string]interface{}{
				"filter_id":   100,
				"action":      "pass",
				"source":      "*",
				"destination": "192.168.1.0/24",
				"protocol":    "tcp",
				"source_port": "*",
				"dest_port":   "*",
				"established": true,
			},
		},
		{
			name: "filter with empty ports converts to asterisk",
			filter: &client.IPFilter{
				Number:        300,
				Action:        "pass",
				SourceAddress: "*",
				DestAddress:   "*",
				Protocol:      "udp",
				SourcePort:    "",
				DestPort:      "",
				Established:   false,
			},
			expected: map[string]interface{}{
				"filter_id":   300,
				"action":      "pass",
				"source":      "*",
				"destination": "*",
				"protocol":    "udp",
				"source_port": "*",
				"dest_port":   "*",
				"established": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceSchema := resourceRTXAccessListIP().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{
				"filter_id":   0,
				"action":      "pass",
				"source":      "*",
				"destination": "*",
				"protocol":    "*",
				"source_port": "*",
				"dest_port":   "*",
				"established": false,
			})

			err := flattenIPFilterToResourceData(tt.filter, d)
			assert.NoError(t, err)

			assert.Equal(t, tt.expected["filter_id"], d.Get("filter_id"))
			assert.Equal(t, tt.expected["action"], d.Get("action"))
			assert.Equal(t, tt.expected["source"], d.Get("source"))
			assert.Equal(t, tt.expected["destination"], d.Get("destination"))
			assert.Equal(t, tt.expected["protocol"], d.Get("protocol"))
			assert.Equal(t, tt.expected["source_port"], d.Get("source_port"))
			assert.Equal(t, tt.expected["dest_port"], d.Get("dest_port"))
			assert.Equal(t, tt.expected["established"], d.Get("established"))
		})
	}
}

func TestResourceRTXAccessListIPSchema(t *testing.T) {
	resource := resourceRTXAccessListIP()

	// Test that required fields are marked as required
	t.Run("filter_id is required", func(t *testing.T) {
		assert.True(t, resource.Schema["filter_id"].Required)
		assert.True(t, resource.Schema["filter_id"].ForceNew)
	})

	t.Run("action is required", func(t *testing.T) {
		assert.True(t, resource.Schema["action"].Required)
	})

	t.Run("source is required", func(t *testing.T) {
		assert.True(t, resource.Schema["source"].Required)
	})

	t.Run("destination is required", func(t *testing.T) {
		assert.True(t, resource.Schema["destination"].Required)
	})

	t.Run("protocol is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["protocol"].Optional)
		assert.True(t, resource.Schema["protocol"].Computed)
	})

	t.Run("source_port is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["source_port"].Optional)
		assert.True(t, resource.Schema["source_port"].Computed)
	})

	t.Run("dest_port is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["dest_port"].Optional)
		assert.True(t, resource.Schema["dest_port"].Computed)
	})

	t.Run("established is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["established"].Optional)
		assert.True(t, resource.Schema["established"].Computed)
	})
}

func TestResourceRTXAccessListIPSchemaValidation(t *testing.T) {
	resource := resourceRTXAccessListIP()

	t.Run("action validation", func(t *testing.T) {
		validActions := []string{"pass", "reject", "restrict", "restrict-log"}
		for _, action := range validActions {
			_, errs := resource.Schema["action"].ValidateFunc(action, "action")
			assert.Empty(t, errs, "action '%s' should be valid", action)
		}

		_, errs := resource.Schema["action"].ValidateFunc("invalid", "action")
		assert.NotEmpty(t, errs, "action 'invalid' should be invalid")
	})

	t.Run("protocol validation", func(t *testing.T) {
		validProtocols := []string{"tcp", "udp", "icmp", "ip", "gre", "esp", "ah", "tcpfin", "tcprst", "*"}
		for _, proto := range validProtocols {
			_, errs := resource.Schema["protocol"].ValidateFunc(proto, "protocol")
			assert.Empty(t, errs, "protocol '%s' should be valid", proto)
		}

		_, errs := resource.Schema["protocol"].ValidateFunc("invalid", "protocol")
		assert.NotEmpty(t, errs, "protocol 'invalid' should be invalid")
	})

	t.Run("filter_id validation", func(t *testing.T) {
		// Valid range: 1-2147483647
		_, errs := resource.Schema["filter_id"].ValidateFunc(1, "filter_id")
		assert.Empty(t, errs, "filter_id 1 should be valid")

		_, errs = resource.Schema["filter_id"].ValidateFunc(65535, "filter_id")
		assert.Empty(t, errs, "filter_id 65535 should be valid")

		_, errs = resource.Schema["filter_id"].ValidateFunc(0, "filter_id")
		assert.NotEmpty(t, errs, "filter_id 0 should be invalid")

		// 65536 is valid since range is 1-2147483647
		_, errs = resource.Schema["filter_id"].ValidateFunc(65536, "filter_id")
		assert.Empty(t, errs, "filter_id 65536 should be valid")
	})
}

func TestResourceRTXAccessListIPImporter(t *testing.T) {
	resource := resourceRTXAccessListIP()

	t.Run("importer is configured", func(t *testing.T) {
		assert.NotNil(t, resource.Importer)
		assert.NotNil(t, resource.Importer.StateContext)
	})
}

func TestResourceRTXAccessListIPCRUDFunctions(t *testing.T) {
	resource := resourceRTXAccessListIP()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, resource.CreateContext)
		assert.NotNil(t, resource.ReadContext)
		assert.NotNil(t, resource.UpdateContext)
		assert.NotNil(t, resource.DeleteContext)
	})
}
