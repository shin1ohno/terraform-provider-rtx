package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/sh1/terraform-provider-rtx/internal/client"
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
						"sequence":                10,
						"ace_rule_action":         "permit",
						"ace_rule_protocol":       "tcp",
						"source_any":              true,
						"source_prefix":           "",
						"source_prefix_mask":      "",
						"source_port_equal":       "",
						"source_port_range":       "",
						"destination_any":         false,
						"destination_prefix":      "192.168.1.0",
						"destination_prefix_mask": "0.0.0.255",
						"destination_port_equal":  "80",
						"destination_port_range":  "",
						"established":             false,
						"log":                     false,
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
						"sequence":                20,
						"ace_rule_action":         "deny",
						"ace_rule_protocol":       "tcp",
						"source_any":              false,
						"source_prefix":           "10.0.0.0",
						"source_prefix_mask":      "0.255.255.255",
						"source_port_equal":       "",
						"source_port_range":       "",
						"destination_any":         true,
						"destination_prefix":      "",
						"destination_prefix_mask": "",
						"destination_port_equal":  "",
						"destination_port_range":  "",
						"established":             true,
						"log":                     true,
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
						"sequence":                30,
						"ace_rule_action":         "permit",
						"ace_rule_protocol":       "udp",
						"source_any":              true,
						"source_prefix":           "",
						"source_prefix_mask":      "",
						"source_port_equal":       "",
						"source_port_range":       "1024-65535",
						"destination_any":         true,
						"destination_prefix":      "",
						"destination_prefix_mask": "",
						"destination_port_equal":  "",
						"destination_port_range":  "53-53",
						"established":             false,
						"log":                     false,
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
						"sequence":                10,
						"ace_rule_action":         "permit",
						"ace_rule_protocol":       "icmp",
						"source_any":              true,
						"source_prefix":           "",
						"source_prefix_mask":      "",
						"source_port_equal":       "",
						"source_port_range":       "",
						"destination_any":         true,
						"destination_prefix":      "",
						"destination_prefix_mask": "",
						"destination_port_equal":  "",
						"destination_port_range":  "",
						"established":             false,
						"log":                     false,
					},
					map[string]interface{}{
						"sequence":                20,
						"ace_rule_action":         "deny",
						"ace_rule_protocol":       "ip",
						"source_any":              true,
						"source_prefix":           "",
						"source_prefix_mask":      "",
						"source_port_equal":       "",
						"source_port_range":       "",
						"destination_any":         true,
						"destination_prefix":      "",
						"destination_prefix_mask": "",
						"destination_port_equal":  "",
						"destination_port_range":  "",
						"established":             false,
						"log":                     true,
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
					"sequence":                10,
					"ace_rule_action":         "permit",
					"ace_rule_protocol":       "tcp",
					"source_any":              true,
					"source_prefix":           "",
					"source_prefix_mask":      "",
					"source_port_equal":       "",
					"source_port_range":       "",
					"destination_any":         false,
					"destination_prefix":      "192.168.1.0",
					"destination_prefix_mask": "0.0.0.255",
					"destination_port_equal":  "443",
					"destination_port_range":  "",
					"established":             false,
					"log":                     false,
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
					"sequence":                10,
					"ace_rule_action":         "permit",
					"ace_rule_protocol":       "icmp",
					"source_any":              true,
					"source_prefix":           "",
					"source_prefix_mask":      "",
					"source_port_equal":       "",
					"source_port_range":       "",
					"destination_any":         true,
					"destination_prefix":      "",
					"destination_prefix_mask": "",
					"destination_port_equal":  "",
					"destination_port_range":  "",
					"established":             false,
					"log":                     false,
				},
				{
					"sequence":                20,
					"ace_rule_action":         "deny",
					"ace_rule_protocol":       "ip",
					"source_any":              true,
					"source_prefix":           "",
					"source_prefix_mask":      "",
					"source_port_equal":       "",
					"source_port_range":       "",
					"destination_any":         true,
					"destination_prefix":      "",
					"destination_prefix_mask": "",
					"destination_port_equal":  "",
					"destination_port_range":  "",
					"established":             false,
					"log":                     true,
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

		// sequence is Optional (computed in auto mode)
		assert.True(t, entrySchema["sequence"].Optional)
		assert.True(t, entrySchema["sequence"].Computed)
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

		_, errs = entrySchema["sequence"].ValidateFunc(2147483647, "sequence")
		assert.Empty(t, errs, "sequence 2147483647 should be valid")

		_, errs = entrySchema["sequence"].ValidateFunc(0, "sequence")
		assert.NotEmpty(t, errs, "sequence 0 should be invalid")

		_, errs = entrySchema["sequence"].ValidateFunc(2147483648, "sequence")
		assert.NotEmpty(t, errs, "sequence 2147483648 should be invalid")
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

// Acceptance Tests
// These tests require a real RTX router and TF_ACC=1 environment variable

func TestAccRTXAccessListExtended_AutoSequence(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXAccessListExtendedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXAccessListExtendedConfig_autoSequence(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRTXAccessListExtendedExists("rtx_access_list_extended.auto_seq"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.auto_seq", "name", "auto-seq-acl"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.auto_seq", "sequence_start", "100"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.auto_seq", "sequence_step", "10"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.auto_seq", "entry.#", "3"),
					// Auto-calculated sequences: 100, 110, 120
					resource.TestCheckResourceAttr("rtx_access_list_extended.auto_seq", "entry.0.sequence", "100"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.auto_seq", "entry.1.sequence", "110"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.auto_seq", "entry.2.sequence", "120"),
				),
			},
		},
	})
}

func TestAccRTXAccessListExtended_MultipleApply(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXAccessListExtendedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXAccessListExtendedConfig_multipleApply(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRTXAccessListExtendedExists("rtx_access_list_extended.multi_apply"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.multi_apply", "name", "multi-apply-acl"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.multi_apply", "apply.#", "2"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.multi_apply", "apply.0.interface", "lan1"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.multi_apply", "apply.0.direction", "in"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.multi_apply", "apply.1.interface", "lan1"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.multi_apply", "apply.1.direction", "out"),
				),
			},
		},
	})
}

func TestAccRTXAccessListExtended_UpdateAddEntry(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXAccessListExtendedDestroy,
		Steps: []resource.TestStep{
			// Initial creation with 2 entries
			{
				Config: testAccRTXAccessListExtendedConfig_updateInitial(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRTXAccessListExtendedExists("rtx_access_list_extended.update_test"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.update_test", "name", "update-test-acl"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.update_test", "entry.#", "2"),
				),
			},
			// Update: add a third entry
			{
				Config: testAccRTXAccessListExtendedConfig_updateAddEntry(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRTXAccessListExtendedExists("rtx_access_list_extended.update_test"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.update_test", "name", "update-test-acl"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.update_test", "entry.#", "3"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.update_test", "entry.2.ace_rule_action", "deny"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.update_test", "entry.2.ace_rule_protocol", "ip"),
				),
			},
		},
	})
}

func TestAccRTXAccessListExtended_Import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckRTXAccessListExtendedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXAccessListExtendedConfig_import(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRTXAccessListExtendedExists("rtx_access_list_extended.import_test"),
					resource.TestCheckResourceAttr("rtx_access_list_extended.import_test", "name", "import-test-acl"),
				),
			},
			{
				ResourceName:      "rtx_access_list_extended.import_test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "import-test-acl",
				// sequence_start and sequence_step are not stored on router, so skip verification
				ImportStateVerifyIgnore: []string{"sequence_start", "sequence_step"},
			},
		},
	})
}

// Helper functions for acceptance tests

func testAccCheckRTXAccessListExtendedExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Access List Extended ID is set")
		}

		return nil
	}
}

func testAccCheckRTXAccessListExtendedDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "rtx_access_list_extended" {
			continue
		}

		// Resource should be deleted. If we had client access, we could verify.
	}

	return nil
}

// Test configuration generators

func testAccRTXAccessListExtendedConfig_autoSequence() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_access_list_extended" "auto_seq" {
  name           = "auto-seq-acl"
  sequence_start = 100
  sequence_step  = 10

  entry {
    ace_rule_action   = "permit"
    ace_rule_protocol = "tcp"
    source_any        = true
    destination_any   = false
    destination_prefix      = "192.168.1.0"
    destination_prefix_mask = "0.0.0.255"
    destination_port_equal  = "80"
  }

  entry {
    ace_rule_action   = "permit"
    ace_rule_protocol = "tcp"
    source_any        = true
    destination_any   = false
    destination_prefix      = "192.168.1.0"
    destination_prefix_mask = "0.0.0.255"
    destination_port_equal  = "443"
  }

  entry {
    ace_rule_action   = "deny"
    ace_rule_protocol = "ip"
    source_any        = true
    destination_any   = true
    log               = true
  }
}
`
}

func testAccRTXAccessListExtendedConfig_multipleApply() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_access_list_extended" "multi_apply" {
  name           = "multi-apply-acl"
  sequence_start = 100
  sequence_step  = 10

  entry {
    ace_rule_action   = "permit"
    ace_rule_protocol = "tcp"
    source_any        = true
    destination_any   = false
    destination_prefix      = "10.0.0.0"
    destination_prefix_mask = "0.255.255.255"
    destination_port_equal  = "22"
  }

  entry {
    ace_rule_action   = "deny"
    ace_rule_protocol = "ip"
    source_any        = true
    destination_any   = true
  }

  apply {
    interface  = "lan1"
    direction  = "in"
    filter_ids = [100, 110]
  }

  apply {
    interface  = "lan1"
    direction  = "out"
    filter_ids = [100, 110]
  }
}
`
}

func testAccRTXAccessListExtendedConfig_updateInitial() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_access_list_extended" "update_test" {
  name           = "update-test-acl"
  sequence_start = 100
  sequence_step  = 10

  entry {
    ace_rule_action   = "permit"
    ace_rule_protocol = "icmp"
    source_any        = true
    destination_any   = true
  }

  entry {
    ace_rule_action   = "permit"
    ace_rule_protocol = "tcp"
    source_any        = true
    destination_any   = false
    destination_prefix      = "192.168.0.0"
    destination_prefix_mask = "0.0.255.255"
    established       = true
  }
}
`
}

func testAccRTXAccessListExtendedConfig_updateAddEntry() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_access_list_extended" "update_test" {
  name           = "update-test-acl"
  sequence_start = 100
  sequence_step  = 10

  entry {
    ace_rule_action   = "permit"
    ace_rule_protocol = "icmp"
    source_any        = true
    destination_any   = true
  }

  entry {
    ace_rule_action   = "permit"
    ace_rule_protocol = "tcp"
    source_any        = true
    destination_any   = false
    destination_prefix      = "192.168.0.0"
    destination_prefix_mask = "0.0.255.255"
    established       = true
  }

  entry {
    ace_rule_action   = "deny"
    ace_rule_protocol = "ip"
    source_any        = true
    destination_any   = true
    log               = true
  }
}
`
}

func testAccRTXAccessListExtendedConfig_import() string {
	return `
provider "rtx" {
  host                 = "localhost"
  port                 = 2222
  username             = "testuser"
  password             = "testpass"
  skip_host_key_check  = true
}

resource "rtx_access_list_extended" "import_test" {
  name           = "import-test-acl"
  sequence_start = 100
  sequence_step  = 10

  entry {
    ace_rule_action   = "permit"
    ace_rule_protocol = "tcp"
    source_any        = true
    destination_any   = false
    destination_prefix      = "172.16.0.0"
    destination_prefix_mask = "0.15.255.255"
    destination_port_equal  = "8080"
  }

  entry {
    ace_rule_action   = "deny"
    ace_rule_protocol = "ip"
    source_any        = true
    destination_any   = true
  }
}
`
}
