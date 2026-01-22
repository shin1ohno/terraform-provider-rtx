package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestBuildAccessListExtendedFromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.AccessListExtended
	}{
		{
			name: "basic extended ACL with source any",
			input: map[string]interface{}{
				"name": "test-acl",
				"entry": []interface{}{
					map[string]interface{}{
						"sequence":                 10,
						"ace_rule_action":          "permit",
						"ace_rule_protocol":        "tcp",
						"source_any":               true,
						"source_prefix":            "",
						"source_prefix_mask":       "",
						"source_port_equal":        "",
						"source_port_range":        "",
						"destination_any":          false,
						"destination_prefix":       "192.168.1.0",
						"destination_prefix_mask":  "0.0.0.255",
						"destination_port_equal":   "80",
						"destination_port_range":   "",
						"established":              false,
						"log":                      false,
					},
				},
			},
			expected: client.AccessListExtended{
				Name: "test-acl",
				Entries: []client.AccessListExtendedEntry{
					{
						Sequence:              10,
						AceRuleAction:         "permit",
						AceRuleProtocol:       "tcp",
						SourceAny:             true,
						DestinationAny:        false,
						DestinationPrefix:     "192.168.1.0",
						DestinationPrefixMask: "0.0.0.255",
						DestinationPortEqual:  "80",
						Established:           false,
						Log:                   false,
					},
				},
			},
		},
		{
			name: "extended ACL with established TCP",
			input: map[string]interface{}{
				"name": "established-acl",
				"entry": []interface{}{
					map[string]interface{}{
						"sequence":                 20,
						"ace_rule_action":          "deny",
						"ace_rule_protocol":        "tcp",
						"source_any":               false,
						"source_prefix":            "10.0.0.0",
						"source_prefix_mask":       "0.255.255.255",
						"source_port_equal":        "",
						"source_port_range":        "",
						"destination_any":          true,
						"destination_prefix":       "",
						"destination_prefix_mask":  "",
						"destination_port_equal":   "",
						"destination_port_range":   "",
						"established":              true,
						"log":                      true,
					},
				},
			},
			expected: client.AccessListExtended{
				Name: "established-acl",
				Entries: []client.AccessListExtendedEntry{
					{
						Sequence:         20,
						AceRuleAction:    "deny",
						AceRuleProtocol:  "tcp",
						SourceAny:        false,
						SourcePrefix:     "10.0.0.0",
						SourcePrefixMask: "0.255.255.255",
						DestinationAny:   true,
						Established:      true,
						Log:              true,
					},
				},
			},
		},
		{
			name: "extended ACL with port range",
			input: map[string]interface{}{
				"name": "port-range-acl",
				"entry": []interface{}{
					map[string]interface{}{
						"sequence":                 30,
						"ace_rule_action":          "permit",
						"ace_rule_protocol":        "udp",
						"source_any":               true,
						"source_prefix":            "",
						"source_prefix_mask":       "",
						"source_port_equal":        "",
						"source_port_range":        "1024-65535",
						"destination_any":          true,
						"destination_prefix":       "",
						"destination_prefix_mask":  "",
						"destination_port_equal":   "",
						"destination_port_range":   "53-53",
						"established":              false,
						"log":                      false,
					},
				},
			},
			expected: client.AccessListExtended{
				Name: "port-range-acl",
				Entries: []client.AccessListExtendedEntry{
					{
						Sequence:             30,
						AceRuleAction:        "permit",
						AceRuleProtocol:      "udp",
						SourceAny:            true,
						SourcePortRange:      "1024-65535",
						DestinationAny:       true,
						DestinationPortRange: "53-53",
						Established:          false,
						Log:                  false,
					},
				},
			},
		},
		{
			name: "extended ACL with multiple entries",
			input: map[string]interface{}{
				"name": "multi-entry-acl",
				"entry": []interface{}{
					map[string]interface{}{
						"sequence":                 10,
						"ace_rule_action":          "permit",
						"ace_rule_protocol":        "icmp",
						"source_any":               true,
						"source_prefix":            "",
						"source_prefix_mask":       "",
						"source_port_equal":        "",
						"source_port_range":        "",
						"destination_any":          true,
						"destination_prefix":       "",
						"destination_prefix_mask":  "",
						"destination_port_equal":   "",
						"destination_port_range":   "",
						"established":              false,
						"log":                      false,
					},
					map[string]interface{}{
						"sequence":                 20,
						"ace_rule_action":          "deny",
						"ace_rule_protocol":        "ip",
						"source_any":               true,
						"source_prefix":            "",
						"source_prefix_mask":       "",
						"source_port_equal":        "",
						"source_port_range":        "",
						"destination_any":          true,
						"destination_prefix":       "",
						"destination_prefix_mask":  "",
						"destination_port_equal":   "",
						"destination_port_range":   "",
						"established":              false,
						"log":                      true,
					},
				},
			},
			expected: client.AccessListExtended{
				Name: "multi-entry-acl",
				Entries: []client.AccessListExtendedEntry{
					{
						Sequence:        10,
						AceRuleAction:   "permit",
						AceRuleProtocol: "icmp",
						SourceAny:       true,
						DestinationAny:  true,
						Established:     false,
						Log:             false,
					},
					{
						Sequence:        20,
						AceRuleAction:   "deny",
						AceRuleProtocol: "ip",
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
			resourceSchema := resourceRTXAccessListExtended().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			result := buildAccessListExtendedFromResourceData(d)

			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, len(tt.expected.Entries), len(result.Entries))

			for i, expectedEntry := range tt.expected.Entries {
				actualEntry := result.Entries[i]
				assert.Equal(t, expectedEntry.Sequence, actualEntry.Sequence, "entry[%d].Sequence", i)
				assert.Equal(t, expectedEntry.AceRuleAction, actualEntry.AceRuleAction, "entry[%d].AceRuleAction", i)
				assert.Equal(t, expectedEntry.AceRuleProtocol, actualEntry.AceRuleProtocol, "entry[%d].AceRuleProtocol", i)
				assert.Equal(t, expectedEntry.SourceAny, actualEntry.SourceAny, "entry[%d].SourceAny", i)
				assert.Equal(t, expectedEntry.SourcePrefix, actualEntry.SourcePrefix, "entry[%d].SourcePrefix", i)
				assert.Equal(t, expectedEntry.SourcePrefixMask, actualEntry.SourcePrefixMask, "entry[%d].SourcePrefixMask", i)
				assert.Equal(t, expectedEntry.SourcePortEqual, actualEntry.SourcePortEqual, "entry[%d].SourcePortEqual", i)
				assert.Equal(t, expectedEntry.SourcePortRange, actualEntry.SourcePortRange, "entry[%d].SourcePortRange", i)
				assert.Equal(t, expectedEntry.DestinationAny, actualEntry.DestinationAny, "entry[%d].DestinationAny", i)
				assert.Equal(t, expectedEntry.DestinationPrefix, actualEntry.DestinationPrefix, "entry[%d].DestinationPrefix", i)
				assert.Equal(t, expectedEntry.DestinationPrefixMask, actualEntry.DestinationPrefixMask, "entry[%d].DestinationPrefixMask", i)
				assert.Equal(t, expectedEntry.DestinationPortEqual, actualEntry.DestinationPortEqual, "entry[%d].DestinationPortEqual", i)
				assert.Equal(t, expectedEntry.DestinationPortRange, actualEntry.DestinationPortRange, "entry[%d].DestinationPortRange", i)
				assert.Equal(t, expectedEntry.Established, actualEntry.Established, "entry[%d].Established", i)
				assert.Equal(t, expectedEntry.Log, actualEntry.Log, "entry[%d].Log", i)
			}
		})
	}
}

func TestFlattenAccessListExtendedEntries(t *testing.T) {
	tests := []struct {
		name     string
		entries  []client.AccessListExtendedEntry
		expected []map[string]interface{}
	}{
		{
			name: "single entry flattening",
			entries: []client.AccessListExtendedEntry{
				{
					Sequence:              10,
					AceRuleAction:         "permit",
					AceRuleProtocol:       "tcp",
					SourceAny:             true,
					DestinationPrefix:     "192.168.1.0",
					DestinationPrefixMask: "0.0.0.255",
					DestinationPortEqual:  "443",
					Established:           false,
					Log:                   false,
				},
			},
			expected: []map[string]interface{}{
				{
					"sequence":                 10,
					"ace_rule_action":          "permit",
					"ace_rule_protocol":        "tcp",
					"source_any":               true,
					"source_prefix":            "",
					"source_prefix_mask":       "",
					"source_port_equal":        "",
					"source_port_range":        "",
					"destination_any":          false,
					"destination_prefix":       "192.168.1.0",
					"destination_prefix_mask":  "0.0.0.255",
					"destination_port_equal":   "443",
					"destination_port_range":   "",
					"established":              false,
					"log":                      false,
				},
			},
		},
		{
			name: "multiple entries flattening",
			entries: []client.AccessListExtendedEntry{
				{
					Sequence:        10,
					AceRuleAction:   "permit",
					AceRuleProtocol: "icmp",
					SourceAny:       true,
					DestinationAny:  true,
				},
				{
					Sequence:        20,
					AceRuleAction:   "deny",
					AceRuleProtocol: "ip",
					SourceAny:       true,
					DestinationAny:  true,
					Log:             true,
				},
			},
			expected: []map[string]interface{}{
				{
					"sequence":                 10,
					"ace_rule_action":          "permit",
					"ace_rule_protocol":        "icmp",
					"source_any":               true,
					"source_prefix":            "",
					"source_prefix_mask":       "",
					"source_port_equal":        "",
					"source_port_range":        "",
					"destination_any":          true,
					"destination_prefix":       "",
					"destination_prefix_mask":  "",
					"destination_port_equal":   "",
					"destination_port_range":   "",
					"established":              false,
					"log":                      false,
				},
				{
					"sequence":                 20,
					"ace_rule_action":          "deny",
					"ace_rule_protocol":        "ip",
					"source_any":               true,
					"source_prefix":            "",
					"source_prefix_mask":       "",
					"source_port_equal":        "",
					"source_port_range":        "",
					"destination_any":          true,
					"destination_prefix":       "",
					"destination_prefix_mask":  "",
					"destination_port_equal":   "",
					"destination_port_range":   "",
					"established":              false,
					"log":                      true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenAccessListExtendedEntries(tt.entries)

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

func TestResourceRTXAccessListExtendedSchema(t *testing.T) {
	resource := resourceRTXAccessListExtended()

	t.Run("name is required and ForceNew", func(t *testing.T) {
		assert.True(t, resource.Schema["name"].Required)
		assert.True(t, resource.Schema["name"].ForceNew)
	})

	t.Run("entry is required", func(t *testing.T) {
		assert.True(t, resource.Schema["entry"].Required)
	})

	t.Run("entry has correct nested schema", func(t *testing.T) {
		entrySchema := resource.Schema["entry"].Elem.(*schema.Resource).Schema

		assert.True(t, entrySchema["sequence"].Required)
		assert.True(t, entrySchema["ace_rule_action"].Required)
		assert.True(t, entrySchema["ace_rule_protocol"].Required)

		assert.True(t, entrySchema["source_any"].Optional)
		assert.True(t, entrySchema["source_prefix"].Optional)
		assert.True(t, entrySchema["source_prefix_mask"].Optional)
		assert.True(t, entrySchema["destination_any"].Optional)
		assert.True(t, entrySchema["destination_prefix"].Optional)
		assert.True(t, entrySchema["destination_prefix_mask"].Optional)
		assert.True(t, entrySchema["established"].Optional)
		assert.True(t, entrySchema["log"].Optional)
	})
}

func TestResourceRTXAccessListExtendedSchemaValidation(t *testing.T) {
	resource := resourceRTXAccessListExtended()
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

	t.Run("ace_rule_protocol validation", func(t *testing.T) {
		validProtocols := []string{"tcp", "udp", "icmp", "ip", "gre", "esp", "ah", "*"}
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

		_, errs = entrySchema["sequence"].ValidateFunc(65536, "sequence")
		assert.NotEmpty(t, errs, "sequence 65536 should be invalid")
	})
}

func TestResourceRTXAccessListExtendedImporter(t *testing.T) {
	resource := resourceRTXAccessListExtended()

	t.Run("importer is configured", func(t *testing.T) {
		assert.NotNil(t, resource.Importer)
		assert.NotNil(t, resource.Importer.StateContext)
	})
}

func TestResourceRTXAccessListExtendedCRUDFunctions(t *testing.T) {
	resource := resourceRTXAccessListExtended()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, resource.CreateContext)
		assert.NotNil(t, resource.ReadContext)
		assert.NotNil(t, resource.UpdateContext)
		assert.NotNil(t, resource.DeleteContext)
	})
}

func TestResourceRTXAccessListExtendedCustomizeDiff(t *testing.T) {
	resource := resourceRTXAccessListExtended()

	t.Run("CustomizeDiff is configured", func(t *testing.T) {
		assert.NotNil(t, resource.CustomizeDiff)
	})
}
