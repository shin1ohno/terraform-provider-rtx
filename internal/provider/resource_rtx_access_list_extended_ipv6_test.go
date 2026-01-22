package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func TestBuildAccessListExtendedIPv6FromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.AccessListExtendedIPv6
	}{
		{
			name: "basic IPv6 extended ACL with source any",
			input: map[string]interface{}{
				"name": "test-ipv6-acl",
				"entry": []interface{}{
					map[string]interface{}{
						"sequence":                  10,
						"ace_rule_action":           "permit",
						"ace_rule_protocol":         "tcp",
						"source_any":                true,
						"source_prefix":             "",
						"source_prefix_length":      0,
						"source_port_equal":         "",
						"source_port_range":         "",
						"destination_any":           false,
						"destination_prefix":        "2001:db8:1::",
						"destination_prefix_length": 64,
						"destination_port_equal":    "80",
						"destination_port_range":    "",
						"established":               false,
						"log":                       false,
					},
				},
			},
			expected: client.AccessListExtendedIPv6{
				Name: "test-ipv6-acl",
				Entries: []client.AccessListExtendedIPv6Entry{
					{
						Sequence:                10,
						AceRuleAction:           "permit",
						AceRuleProtocol:         "tcp",
						SourceAny:               true,
						DestinationAny:          false,
						DestinationPrefix:       "2001:db8:1::",
						DestinationPrefixLength: 64,
						DestinationPortEqual:    "80",
						Established:             false,
						Log:                     false,
					},
				},
			},
		},
		{
			name: "IPv6 extended ACL with established TCP",
			input: map[string]interface{}{
				"name": "established-ipv6-acl",
				"entry": []interface{}{
					map[string]interface{}{
						"sequence":                  20,
						"ace_rule_action":           "deny",
						"ace_rule_protocol":         "tcp",
						"source_any":                false,
						"source_prefix":             "2001:db8::",
						"source_prefix_length":      32,
						"source_port_equal":         "",
						"source_port_range":         "",
						"destination_any":           true,
						"destination_prefix":        "",
						"destination_prefix_length": 0,
						"destination_port_equal":    "",
						"destination_port_range":    "",
						"established":               true,
						"log":                       true,
					},
				},
			},
			expected: client.AccessListExtendedIPv6{
				Name: "established-ipv6-acl",
				Entries: []client.AccessListExtendedIPv6Entry{
					{
						Sequence:           20,
						AceRuleAction:      "deny",
						AceRuleProtocol:    "tcp",
						SourceAny:          false,
						SourcePrefix:       "2001:db8::",
						SourcePrefixLength: 32,
						DestinationAny:     true,
						Established:        true,
						Log:                true,
					},
				},
			},
		},
		{
			name: "IPv6 extended ACL with icmpv6",
			input: map[string]interface{}{
				"name": "icmpv6-acl",
				"entry": []interface{}{
					map[string]interface{}{
						"sequence":                  30,
						"ace_rule_action":           "permit",
						"ace_rule_protocol":         "icmpv6",
						"source_any":                true,
						"source_prefix":             "",
						"source_prefix_length":      0,
						"source_port_equal":         "",
						"source_port_range":         "",
						"destination_any":           true,
						"destination_prefix":        "",
						"destination_prefix_length": 0,
						"destination_port_equal":    "",
						"destination_port_range":    "",
						"established":               false,
						"log":                       false,
					},
				},
			},
			expected: client.AccessListExtendedIPv6{
				Name: "icmpv6-acl",
				Entries: []client.AccessListExtendedIPv6Entry{
					{
						Sequence:        30,
						AceRuleAction:   "permit",
						AceRuleProtocol: "icmpv6",
						SourceAny:       true,
						DestinationAny:  true,
						Established:     false,
						Log:             false,
					},
				},
			},
		},
		{
			name: "IPv6 extended ACL with multiple entries",
			input: map[string]interface{}{
				"name": "multi-entry-ipv6-acl",
				"entry": []interface{}{
					map[string]interface{}{
						"sequence":                  10,
						"ace_rule_action":           "permit",
						"ace_rule_protocol":         "tcp",
						"source_any":                true,
						"source_prefix":             "",
						"source_prefix_length":      0,
						"source_port_equal":         "",
						"source_port_range":         "",
						"destination_any":           false,
						"destination_prefix":        "2001:db8:1::",
						"destination_prefix_length": 64,
						"destination_port_equal":    "443",
						"destination_port_range":    "",
						"established":               false,
						"log":                       false,
					},
					map[string]interface{}{
						"sequence":                  20,
						"ace_rule_action":           "deny",
						"ace_rule_protocol":         "ipv6",
						"source_any":                true,
						"source_prefix":             "",
						"source_prefix_length":      0,
						"source_port_equal":         "",
						"source_port_range":         "",
						"destination_any":           true,
						"destination_prefix":        "",
						"destination_prefix_length": 0,
						"destination_port_equal":    "",
						"destination_port_range":    "",
						"established":               false,
						"log":                       true,
					},
				},
			},
			expected: client.AccessListExtendedIPv6{
				Name: "multi-entry-ipv6-acl",
				Entries: []client.AccessListExtendedIPv6Entry{
					{
						Sequence:                10,
						AceRuleAction:           "permit",
						AceRuleProtocol:         "tcp",
						SourceAny:               true,
						DestinationAny:          false,
						DestinationPrefix:       "2001:db8:1::",
						DestinationPrefixLength: 64,
						DestinationPortEqual:    "443",
						Established:             false,
						Log:                     false,
					},
					{
						Sequence:        20,
						AceRuleAction:   "deny",
						AceRuleProtocol: "ipv6",
						SourceAny:       true,
						DestinationAny:  true,
						Established:     false,
						Log:             true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceSchema := resourceRTXAccessListExtendedIPv6().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			result := buildAccessListExtendedIPv6FromResourceData(d)

			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, len(tt.expected.Entries), len(result.Entries))

			for i, expectedEntry := range tt.expected.Entries {
				actualEntry := result.Entries[i]
				assert.Equal(t, expectedEntry.Sequence, actualEntry.Sequence, "entry[%d].Sequence", i)
				assert.Equal(t, expectedEntry.AceRuleAction, actualEntry.AceRuleAction, "entry[%d].AceRuleAction", i)
				assert.Equal(t, expectedEntry.AceRuleProtocol, actualEntry.AceRuleProtocol, "entry[%d].AceRuleProtocol", i)
				assert.Equal(t, expectedEntry.SourceAny, actualEntry.SourceAny, "entry[%d].SourceAny", i)
				assert.Equal(t, expectedEntry.SourcePrefix, actualEntry.SourcePrefix, "entry[%d].SourcePrefix", i)
				assert.Equal(t, expectedEntry.SourcePrefixLength, actualEntry.SourcePrefixLength, "entry[%d].SourcePrefixLength", i)
				assert.Equal(t, expectedEntry.DestinationAny, actualEntry.DestinationAny, "entry[%d].DestinationAny", i)
				assert.Equal(t, expectedEntry.DestinationPrefix, actualEntry.DestinationPrefix, "entry[%d].DestinationPrefix", i)
				assert.Equal(t, expectedEntry.DestinationPrefixLength, actualEntry.DestinationPrefixLength, "entry[%d].DestinationPrefixLength", i)
				assert.Equal(t, expectedEntry.Established, actualEntry.Established, "entry[%d].Established", i)
				assert.Equal(t, expectedEntry.Log, actualEntry.Log, "entry[%d].Log", i)
			}
		})
	}
}

func TestFlattenAccessListExtendedIPv6Entries(t *testing.T) {
	tests := []struct {
		name     string
		entries  []client.AccessListExtendedIPv6Entry
		expected []map[string]interface{}
	}{
		{
			name: "single IPv6 entry flattening",
			entries: []client.AccessListExtendedIPv6Entry{
				{
					Sequence:                10,
					AceRuleAction:           "permit",
					AceRuleProtocol:         "tcp",
					SourceAny:               true,
					DestinationPrefix:       "2001:db8::",
					DestinationPrefixLength: 32,
					DestinationPortEqual:    "443",
					Established:             false,
					Log:                     false,
				},
			},
			expected: []map[string]interface{}{
				{
					"sequence":                  10,
					"ace_rule_action":           "permit",
					"ace_rule_protocol":         "tcp",
					"source_any":                true,
					"source_prefix":             "",
					"source_prefix_length":      0,
					"source_port_equal":         "",
					"source_port_range":         "",
					"destination_any":           false,
					"destination_prefix":        "2001:db8::",
					"destination_prefix_length": 32,
					"destination_port_equal":    "443",
					"destination_port_range":    "",
					"established":               false,
					"log":                       false,
				},
			},
		},
		{
			name: "multiple IPv6 entries flattening",
			entries: []client.AccessListExtendedIPv6Entry{
				{
					Sequence:        10,
					AceRuleAction:   "permit",
					AceRuleProtocol: "icmpv6",
					SourceAny:       true,
					DestinationAny:  true,
				},
				{
					Sequence:        20,
					AceRuleAction:   "deny",
					AceRuleProtocol: "ipv6",
					SourceAny:       true,
					DestinationAny:  true,
					Log:             true,
				},
			},
			expected: []map[string]interface{}{
				{
					"sequence":                  10,
					"ace_rule_action":           "permit",
					"ace_rule_protocol":         "icmpv6",
					"source_any":                true,
					"source_prefix":             "",
					"source_prefix_length":      0,
					"source_port_equal":         "",
					"source_port_range":         "",
					"destination_any":           true,
					"destination_prefix":        "",
					"destination_prefix_length": 0,
					"destination_port_equal":    "",
					"destination_port_range":    "",
					"established":               false,
					"log":                       false,
				},
				{
					"sequence":                  20,
					"ace_rule_action":           "deny",
					"ace_rule_protocol":         "ipv6",
					"source_any":                true,
					"source_prefix":             "",
					"source_prefix_length":      0,
					"source_port_equal":         "",
					"source_port_range":         "",
					"destination_any":           true,
					"destination_prefix":        "",
					"destination_prefix_length": 0,
					"destination_port_equal":    "",
					"destination_port_range":    "",
					"established":               false,
					"log":                       true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenAccessListExtendedIPv6Entries(tt.entries)

			assert.Equal(t, len(tt.expected), len(result))

			for i, expectedEntry := range tt.expected {
				actualEntry := result[i].(map[string]interface{})
				assert.Equal(t, expectedEntry["sequence"], actualEntry["sequence"])
				assert.Equal(t, expectedEntry["ace_rule_action"], actualEntry["ace_rule_action"])
				assert.Equal(t, expectedEntry["ace_rule_protocol"], actualEntry["ace_rule_protocol"])
				assert.Equal(t, expectedEntry["source_any"], actualEntry["source_any"])
				assert.Equal(t, expectedEntry["destination_any"], actualEntry["destination_any"])
				assert.Equal(t, expectedEntry["established"], actualEntry["established"])
				assert.Equal(t, expectedEntry["log"], actualEntry["log"])
			}
		})
	}
}

func TestResourceRTXAccessListExtendedIPv6Schema(t *testing.T) {
	resource := resourceRTXAccessListExtendedIPv6()

	t.Run("name is required and ForceNew", func(t *testing.T) {
		assert.True(t, resource.Schema["name"].Required)
		assert.True(t, resource.Schema["name"].ForceNew)
	})

	t.Run("entry is required", func(t *testing.T) {
		assert.True(t, resource.Schema["entry"].Required)
	})

	t.Run("entry has correct nested schema for IPv6", func(t *testing.T) {
		entrySchema := resource.Schema["entry"].Elem.(*schema.Resource).Schema

		assert.True(t, entrySchema["sequence"].Required)
		assert.True(t, entrySchema["ace_rule_action"].Required)
		assert.True(t, entrySchema["ace_rule_protocol"].Required)

		assert.True(t, entrySchema["source_any"].Optional)
		assert.True(t, entrySchema["source_prefix"].Optional)
		assert.True(t, entrySchema["source_prefix_length"].Optional)
		assert.True(t, entrySchema["destination_any"].Optional)
		assert.True(t, entrySchema["destination_prefix"].Optional)
		assert.True(t, entrySchema["destination_prefix_length"].Optional)
		assert.True(t, entrySchema["established"].Optional)
		assert.True(t, entrySchema["log"].Optional)
	})
}

func TestResourceRTXAccessListExtendedIPv6SchemaValidation(t *testing.T) {
	resource := resourceRTXAccessListExtendedIPv6()
	entrySchema := resource.Schema["entry"].Elem.(*schema.Resource).Schema

	t.Run("ace_rule_action validation", func(t *testing.T) {
		validActions := []string{"permit", "deny"}
		for _, action := range validActions {
			_, errs := entrySchema["ace_rule_action"].ValidateFunc(action, "ace_rule_action")
			assert.Empty(t, errs, "action '%s' should be valid", action)
		}

		_, errs := entrySchema["ace_rule_action"].ValidateFunc("invalid", "ace_rule_action")
		assert.NotEmpty(t, errs, "action 'invalid' should be invalid")
	})

	t.Run("ace_rule_protocol validation for IPv6", func(t *testing.T) {
		validProtocols := []string{"tcp", "udp", "icmpv6", "ipv6", "ip", "*"}
		for _, proto := range validProtocols {
			_, errs := entrySchema["ace_rule_protocol"].ValidateFunc(proto, "ace_rule_protocol")
			assert.Empty(t, errs, "protocol '%s' should be valid", proto)
		}

		_, errs := entrySchema["ace_rule_protocol"].ValidateFunc("invalid", "ace_rule_protocol")
		assert.NotEmpty(t, errs, "protocol 'invalid' should be invalid")
	})

	t.Run("sequence validation", func(t *testing.T) {
		_, errs := entrySchema["sequence"].ValidateFunc(1, "sequence")
		assert.Empty(t, errs, "sequence 1 should be valid")

		_, errs = entrySchema["sequence"].ValidateFunc(65535, "sequence")
		assert.Empty(t, errs, "sequence 65535 should be valid")

		_, errs = entrySchema["sequence"].ValidateFunc(0, "sequence")
		assert.NotEmpty(t, errs, "sequence 0 should be invalid")
	})

	t.Run("source_prefix_length validation", func(t *testing.T) {
		_, errs := entrySchema["source_prefix_length"].ValidateFunc(0, "source_prefix_length")
		assert.Empty(t, errs, "prefix_length 0 should be valid")

		_, errs = entrySchema["source_prefix_length"].ValidateFunc(128, "source_prefix_length")
		assert.Empty(t, errs, "prefix_length 128 should be valid")

		_, errs = entrySchema["source_prefix_length"].ValidateFunc(129, "source_prefix_length")
		assert.NotEmpty(t, errs, "prefix_length 129 should be invalid")
	})

	t.Run("destination_prefix_length validation", func(t *testing.T) {
		_, errs := entrySchema["destination_prefix_length"].ValidateFunc(64, "destination_prefix_length")
		assert.Empty(t, errs, "prefix_length 64 should be valid")

		_, errs = entrySchema["destination_prefix_length"].ValidateFunc(-1, "destination_prefix_length")
		assert.NotEmpty(t, errs, "prefix_length -1 should be invalid")
	})
}

func TestResourceRTXAccessListExtendedIPv6Importer(t *testing.T) {
	resource := resourceRTXAccessListExtendedIPv6()

	t.Run("importer is configured", func(t *testing.T) {
		assert.NotNil(t, resource.Importer)
		assert.NotNil(t, resource.Importer.StateContext)
	})
}

func TestResourceRTXAccessListExtendedIPv6CRUDFunctions(t *testing.T) {
	resource := resourceRTXAccessListExtendedIPv6()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, resource.CreateContext)
		assert.NotNil(t, resource.ReadContext)
		assert.NotNil(t, resource.UpdateContext)
		assert.NotNil(t, resource.DeleteContext)
	})
}

func TestResourceRTXAccessListExtendedIPv6CustomizeDiff(t *testing.T) {
	resource := resourceRTXAccessListExtendedIPv6()

	t.Run("CustomizeDiff is configured", func(t *testing.T) {
		assert.NotNil(t, resource.CustomizeDiff)
	})
}
