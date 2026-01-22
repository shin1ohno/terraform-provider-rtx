package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func TestBuildIPv6FilterDynamicConfigFromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.IPv6FilterDynamicConfig
	}{
		{
			name: "basic entry with www protocol",
			input: map[string]interface{}{
				"entry": []interface{}{
					map[string]interface{}{
						"number":      100,
						"source":      "*",
						"destination": "*",
						"protocol":    "www",
						"syslog":      false,
					},
				},
			},
			expected: client.IPv6FilterDynamicConfig{
				Entries: []client.IPv6FilterDynamicEntry{
					{
						Number:   100,
						Source:   "*",
						Dest:     "*",
						Protocol: "www",
						Syslog:   false,
					},
				},
			},
		},
		{
			name: "entry with submission protocol",
			input: map[string]interface{}{
				"entry": []interface{}{
					map[string]interface{}{
						"number":      200,
						"source":      "2001:db8::/32",
						"destination": "*",
						"protocol":    "submission",
						"syslog":      true,
					},
				},
			},
			expected: client.IPv6FilterDynamicConfig{
				Entries: []client.IPv6FilterDynamicEntry{
					{
						Number:   200,
						Source:   "2001:db8::/32",
						Dest:     "*",
						Protocol: "submission",
						Syslog:   true,
					},
				},
			},
		},
		{
			name: "entry with smtp protocol",
			input: map[string]interface{}{
				"entry": []interface{}{
					map[string]interface{}{
						"number":      300,
						"source":      "*",
						"destination": "2001:db8::1",
						"protocol":    "smtp",
						"syslog":      false,
					},
				},
			},
			expected: client.IPv6FilterDynamicConfig{
				Entries: []client.IPv6FilterDynamicEntry{
					{
						Number:   300,
						Source:   "*",
						Dest:     "2001:db8::1",
						Protocol: "smtp",
						Syslog:   false,
					},
				},
			},
		},
		{
			name: "multiple entries",
			input: map[string]interface{}{
				"entry": []interface{}{
					map[string]interface{}{
						"number":      100,
						"source":      "*",
						"destination": "*",
						"protocol":    "ftp",
						"syslog":      false,
					},
					map[string]interface{}{
						"number":      101,
						"source":      "*",
						"destination": "*",
						"protocol":    "submission",
						"syslog":      true,
					},
				},
			},
			expected: client.IPv6FilterDynamicConfig{
				Entries: []client.IPv6FilterDynamicEntry{
					{
						Number:   100,
						Source:   "*",
						Dest:     "*",
						Protocol: "ftp",
						Syslog:   false,
					},
					{
						Number:   101,
						Source:   "*",
						Dest:     "*",
						Protocol: "submission",
						Syslog:   true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceRTXIPv6FilterDynamic().Schema, tt.input)
			result := buildIPv6FilterDynamicConfigFromResourceData(d)
			assert.Equal(t, len(tt.expected.Entries), len(result.Entries))
			for i, expectedEntry := range tt.expected.Entries {
				assert.Equal(t, expectedEntry.Number, result.Entries[i].Number)
				assert.Equal(t, expectedEntry.Source, result.Entries[i].Source)
				assert.Equal(t, expectedEntry.Dest, result.Entries[i].Dest)
				assert.Equal(t, expectedEntry.Protocol, result.Entries[i].Protocol)
				assert.Equal(t, expectedEntry.Syslog, result.Entries[i].Syslog)
			}
		})
	}
}

func TestFlattenIPv6FilterDynamicEntries(t *testing.T) {
	tests := []struct {
		name     string
		input    []client.IPv6FilterDynamicEntry
		expected []interface{}
	}{
		{
			name: "single entry with submission protocol",
			input: []client.IPv6FilterDynamicEntry{
				{
					Number:   100,
					Source:   "*",
					Dest:     "*",
					Protocol: "submission",
					Syslog:   true,
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"number":      100,
					"source":      "*",
					"destination": "*",
					"protocol":    "submission",
					"syslog":      true,
				},
			},
		},
		{
			name: "multiple entries",
			input: []client.IPv6FilterDynamicEntry{
				{
					Number:   200,
					Source:   "2001:db8::/32",
					Dest:     "*",
					Protocol: "www",
					Syslog:   false,
				},
				{
					Number:   201,
					Source:   "*",
					Dest:     "2001:db8::1",
					Protocol: "submission",
					Syslog:   true,
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"number":      200,
					"source":      "2001:db8::/32",
					"destination": "*",
					"protocol":    "www",
					"syslog":      false,
				},
				map[string]interface{}{
					"number":      201,
					"source":      "*",
					"destination": "2001:db8::1",
					"protocol":    "submission",
					"syslog":      true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenIPv6FilterDynamicEntries(tt.input)
			assert.Equal(t, len(tt.expected), len(result))
			for i, expected := range tt.expected {
				expectedMap := expected.(map[string]interface{})
				resultMap := result[i].(map[string]interface{})
				assert.Equal(t, expectedMap["number"], resultMap["number"])
				assert.Equal(t, expectedMap["source"], resultMap["source"])
				assert.Equal(t, expectedMap["destination"], resultMap["destination"])
				assert.Equal(t, expectedMap["protocol"], resultMap["protocol"])
				assert.Equal(t, expectedMap["syslog"], resultMap["syslog"])
			}
		})
	}
}

func TestResourceRTXIPv6FilterDynamicSchema(t *testing.T) {
	resource := resourceRTXIPv6FilterDynamic()

	// Verify entry field exists and is required
	assert.NotNil(t, resource.Schema["entry"])
	assert.True(t, resource.Schema["entry"].Required)

	// Get the entry schema
	entryResource := resource.Schema["entry"].Elem.(*schema.Resource)

	// Verify required fields in entry
	assert.NotNil(t, entryResource.Schema["number"])
	assert.True(t, entryResource.Schema["number"].Required)

	assert.NotNil(t, entryResource.Schema["source"])
	assert.True(t, entryResource.Schema["source"].Required)

	assert.NotNil(t, entryResource.Schema["destination"])
	assert.True(t, entryResource.Schema["destination"].Required)

	assert.NotNil(t, entryResource.Schema["protocol"])
	assert.True(t, entryResource.Schema["protocol"].Required)

	// Verify optional syslog field
	assert.NotNil(t, entryResource.Schema["syslog"])
	assert.True(t, entryResource.Schema["syslog"].Optional)
	assert.Equal(t, false, entryResource.Schema["syslog"].Default)
}

func TestResourceRTXIPv6FilterDynamicSchemaProtocolValidation(t *testing.T) {
	resource := resourceRTXIPv6FilterDynamic()
	entryResource := resource.Schema["entry"].Elem.(*schema.Resource)

	// Test that submission protocol is valid (schema-level validation)
	// This tests that the ValidateFunc allows submission
	protocolSchema := entryResource.Schema["protocol"]
	assert.NotNil(t, protocolSchema.ValidateFunc)

	// Valid protocols should not produce errors
	validProtocols := []string{"ftp", "www", "smtp", "submission", "pop3", "dns", "domain", "telnet", "ssh", "tcp", "udp", "*"}
	for _, proto := range validProtocols {
		_, errors := protocolSchema.ValidateFunc(proto, "protocol")
		assert.Empty(t, errors, "protocol %q should be valid", proto)
	}

	// Invalid protocol should produce error
	_, errors := protocolSchema.ValidateFunc("invalid", "protocol")
	assert.NotEmpty(t, errors, "protocol 'invalid' should not be valid")
}
