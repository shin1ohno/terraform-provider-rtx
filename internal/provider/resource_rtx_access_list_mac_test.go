package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func TestBuildAccessListMACFromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.AccessListMAC
	}{
		{
			name: "basic MAC ACL with source any",
			input: map[string]interface{}{
				"name":      "test-mac-acl",
				"filter_id": 0,
				"entry": []interface{}{
					map[string]interface{}{
						"sequence":                 10,
						"ace_action":               "permit",
						"source_any":               true,
						"source_address":           "",
						"source_address_mask":      "",
						"destination_any":          false,
						"destination_address":      "00:11:22:33:44:55",
						"destination_address_mask": "",
						"ether_type":               "",
						"vlan_id":                  0,
						"log":                      false,
						"filter_id":                0,
						"offset":                   0,
						"byte_list":                []interface{}{},
						"dhcp_match":               []interface{}{},
					},
				},
			},
			expected: client.AccessListMAC{
				Name: "test-mac-acl",
				Entries: []client.AccessListMACEntry{
					{
						Sequence:           10,
						AceAction:          "permit",
						SourceAny:          true,
						DestinationAny:     false,
						DestinationAddress: "00:11:22:33:44:55",
						Log:                false,
					},
				},
			},
		},
		{
			name: "MAC ACL with source and destination addresses",
			input: map[string]interface{}{
				"name":      "mac-acl-full",
				"filter_id": 100,
				"entry": []interface{}{
					map[string]interface{}{
						"sequence":                 20,
						"ace_action":               "deny",
						"source_any":               false,
						"source_address":           "aa:bb:cc:dd:ee:ff",
						"source_address_mask":      "ff:ff:ff:00:00:00",
						"destination_any":          false,
						"destination_address":      "11:22:33:44:55:66",
						"destination_address_mask": "ff:ff:ff:ff:ff:ff",
						"ether_type":               "0x0800",
						"vlan_id":                  100,
						"log":                      true,
						"filter_id":                0,
						"offset":                   0,
						"byte_list":                []interface{}{},
						"dhcp_match":               []interface{}{},
					},
				},
			},
			expected: client.AccessListMAC{
				Name:     "mac-acl-full",
				FilterID: 100,
				Entries: []client.AccessListMACEntry{
					{
						Sequence:               20,
						AceAction:              "deny",
						SourceAny:              false,
						SourceAddress:          "aa:bb:cc:dd:ee:ff",
						SourceAddressMask:      "ff:ff:ff:00:00:00",
						DestinationAny:         false,
						DestinationAddress:     "11:22:33:44:55:66",
						DestinationAddressMask: "ff:ff:ff:ff:ff:ff",
						EtherType:              "0x0800",
						VlanID:                 100,
						Log:                    true,
						FilterID:               100,
					},
				},
			},
		},
		{
			name: "MAC ACL with RTX pass/reject actions",
			input: map[string]interface{}{
				"name":      "rtx-mac-acl",
				"filter_id": 200,
				"entry": []interface{}{
					map[string]interface{}{
						"sequence":                 10,
						"ace_action":               "pass-log",
						"source_any":               true,
						"source_address":           "",
						"source_address_mask":      "",
						"destination_any":          true,
						"destination_address":      "",
						"destination_address_mask": "",
						"ether_type":               "",
						"vlan_id":                  0,
						"log":                      false,
						"filter_id":                201,
						"offset":                   0,
						"byte_list":                []interface{}{},
						"dhcp_match":               []interface{}{},
					},
					map[string]interface{}{
						"sequence":                 20,
						"ace_action":               "reject-nolog",
						"source_any":               true,
						"source_address":           "",
						"source_address_mask":      "",
						"destination_any":          true,
						"destination_address":      "",
						"destination_address_mask": "",
						"ether_type":               "",
						"vlan_id":                  0,
						"log":                      false,
						"filter_id":                202,
						"offset":                   0,
						"byte_list":                []interface{}{},
						"dhcp_match":               []interface{}{},
					},
				},
			},
			expected: client.AccessListMAC{
				Name:     "rtx-mac-acl",
				FilterID: 200,
				Entries: []client.AccessListMACEntry{
					{
						Sequence:       10,
						AceAction:      "pass-log",
						SourceAny:      true,
						DestinationAny: true,
						FilterID:       201,
					},
					{
						Sequence:       20,
						AceAction:      "reject-nolog",
						SourceAny:      true,
						DestinationAny: true,
						FilterID:       202,
					},
				},
			},
		},
		{
			name: "MAC ACL with DHCP match",
			input: map[string]interface{}{
				"name":      "dhcp-mac-acl",
				"filter_id": 0,
				"entry": []interface{}{
					map[string]interface{}{
						"sequence":                 10,
						"ace_action":               "permit",
						"source_any":               true,
						"source_address":           "",
						"source_address_mask":      "",
						"destination_any":          true,
						"destination_address":      "",
						"destination_address_mask": "",
						"ether_type":               "",
						"vlan_id":                  0,
						"log":                      false,
						"filter_id":                0,
						"offset":                   0,
						"byte_list":                []interface{}{},
						"dhcp_match": []interface{}{
							map[string]interface{}{
								"type":  "dhcp-bind",
								"scope": 1,
							},
						},
					},
				},
			},
			expected: client.AccessListMAC{
				Name: "dhcp-mac-acl",
				Entries: []client.AccessListMACEntry{
					{
						Sequence:       10,
						AceAction:      "permit",
						SourceAny:      true,
						DestinationAny: true,
						DHCPType:       "dhcp-bind",
						DHCPScope:      1,
					},
				},
			},
		},
		{
			name: "MAC ACL with byte offset matching",
			input: map[string]interface{}{
				"name":      "byte-match-acl",
				"filter_id": 0,
				"entry": []interface{}{
					map[string]interface{}{
						"sequence":                 10,
						"ace_action":               "deny",
						"source_any":               true,
						"source_address":           "",
						"source_address_mask":      "",
						"destination_any":          true,
						"destination_address":      "",
						"destination_address_mask": "",
						"ether_type":               "",
						"vlan_id":                  0,
						"log":                      false,
						"filter_id":                0,
						"offset":                   14,
						"byte_list":                []interface{}{"45", "00"},
						"dhcp_match":               []interface{}{},
					},
				},
			},
			expected: client.AccessListMAC{
				Name: "byte-match-acl",
				Entries: []client.AccessListMACEntry{
					{
						Sequence:       10,
						AceAction:      "deny",
						SourceAny:      true,
						DestinationAny: true,
						Offset:         14,
						ByteList:       []string{"45", "00"},
					},
				},
			},
		},
		{
			name: "MAC ACL with apply block",
			input: map[string]interface{}{
				"name":      "applied-mac-acl",
				"filter_id": 300,
				"entry": []interface{}{
					map[string]interface{}{
						"sequence":                 10,
						"ace_action":               "pass",
						"source_any":               true,
						"source_address":           "",
						"source_address_mask":      "",
						"destination_any":          true,
						"destination_address":      "",
						"destination_address_mask": "",
						"ether_type":               "",
						"vlan_id":                  0,
						"log":                      false,
						"filter_id":                301,
						"offset":                   0,
						"byte_list":                []interface{}{},
						"dhcp_match":               []interface{}{},
					},
				},
				"apply": []interface{}{
					map[string]interface{}{
						"interface":  "lan1",
						"direction":  "in",
						"filter_ids": []interface{}{301, 302, 303},
					},
				},
			},
			expected: client.AccessListMAC{
				Name:     "applied-mac-acl",
				FilterID: 300,
				Entries: []client.AccessListMACEntry{
					{
						Sequence:       10,
						AceAction:      "pass",
						SourceAny:      true,
						DestinationAny: true,
						FilterID:       301,
					},
				},
				Apply: &client.MACApply{
					Interface: "lan1",
					Direction: "in",
					FilterIDs: []int{301, 302, 303},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceSchema := resourceRTXAccessListMAC().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			result := buildAccessListMACFromResourceData(d)

			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.FilterID, result.FilterID)
			assert.Equal(t, len(tt.expected.Entries), len(result.Entries))

			for i, expectedEntry := range tt.expected.Entries {
				actualEntry := result.Entries[i]
				assert.Equal(t, expectedEntry.Sequence, actualEntry.Sequence, "entry[%d].Sequence", i)
				assert.Equal(t, expectedEntry.AceAction, actualEntry.AceAction, "entry[%d].AceAction", i)
				assert.Equal(t, expectedEntry.SourceAny, actualEntry.SourceAny, "entry[%d].SourceAny", i)
				assert.Equal(t, expectedEntry.SourceAddress, actualEntry.SourceAddress, "entry[%d].SourceAddress", i)
				assert.Equal(t, expectedEntry.SourceAddressMask, actualEntry.SourceAddressMask, "entry[%d].SourceAddressMask", i)
				assert.Equal(t, expectedEntry.DestinationAny, actualEntry.DestinationAny, "entry[%d].DestinationAny", i)
				assert.Equal(t, expectedEntry.DestinationAddress, actualEntry.DestinationAddress, "entry[%d].DestinationAddress", i)
				assert.Equal(t, expectedEntry.DestinationAddressMask, actualEntry.DestinationAddressMask, "entry[%d].DestinationAddressMask", i)
				assert.Equal(t, expectedEntry.EtherType, actualEntry.EtherType, "entry[%d].EtherType", i)
				assert.Equal(t, expectedEntry.VlanID, actualEntry.VlanID, "entry[%d].VlanID", i)
				assert.Equal(t, expectedEntry.Log, actualEntry.Log, "entry[%d].Log", i)
				assert.Equal(t, expectedEntry.DHCPType, actualEntry.DHCPType, "entry[%d].DHCPType", i)
				assert.Equal(t, expectedEntry.DHCPScope, actualEntry.DHCPScope, "entry[%d].DHCPScope", i)
				assert.Equal(t, expectedEntry.Offset, actualEntry.Offset, "entry[%d].Offset", i)
				assert.Equal(t, expectedEntry.ByteList, actualEntry.ByteList, "entry[%d].ByteList", i)
			}

			if tt.expected.Apply != nil {
				assert.NotNil(t, result.Apply)
				assert.Equal(t, tt.expected.Apply.Interface, result.Apply.Interface)
				assert.Equal(t, tt.expected.Apply.Direction, result.Apply.Direction)
				assert.Equal(t, tt.expected.Apply.FilterIDs, result.Apply.FilterIDs)
			}
		})
	}
}

func TestResourceRTXAccessListMACSchema(t *testing.T) {
	resource := resourceRTXAccessListMAC()

	t.Run("name is required and ForceNew", func(t *testing.T) {
		assert.True(t, resource.Schema["name"].Required)
		assert.True(t, resource.Schema["name"].ForceNew)
	})

	t.Run("entry is required", func(t *testing.T) {
		assert.True(t, resource.Schema["entry"].Required)
	})

	t.Run("filter_id is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["filter_id"].Optional)
	})

	t.Run("apply is optional with MaxItems 1", func(t *testing.T) {
		assert.True(t, resource.Schema["apply"].Optional)
		assert.Equal(t, 1, resource.Schema["apply"].MaxItems)
	})

	t.Run("entry has correct nested schema", func(t *testing.T) {
		entrySchema := resource.Schema["entry"].Elem.(*schema.Resource).Schema

		assert.True(t, entrySchema["sequence"].Required)
		assert.True(t, entrySchema["ace_action"].Required)

		assert.True(t, entrySchema["source_any"].Optional)
		assert.True(t, entrySchema["source_address"].Optional)
		assert.True(t, entrySchema["source_address_mask"].Optional)
		assert.True(t, entrySchema["destination_any"].Optional)
		assert.True(t, entrySchema["destination_address"].Optional)
		assert.True(t, entrySchema["destination_address_mask"].Optional)
		assert.True(t, entrySchema["ether_type"].Optional)
		assert.True(t, entrySchema["vlan_id"].Optional)
		assert.True(t, entrySchema["log"].Optional)
		assert.True(t, entrySchema["filter_id"].Optional)
		assert.True(t, entrySchema["dhcp_match"].Optional)
		assert.True(t, entrySchema["offset"].Optional)
		assert.True(t, entrySchema["byte_list"].Optional)
	})
}

func TestResourceRTXAccessListMACSchemaValidation(t *testing.T) {
	resource := resourceRTXAccessListMAC()
	entrySchema := resource.Schema["entry"].Elem.(*schema.Resource).Schema

	t.Run("ace_action validation", func(t *testing.T) {
		validActions := []string{"permit", "deny", "pass-log", "pass-nolog", "reject-log", "reject-nolog", "pass", "reject"}
		for _, action := range validActions {
			_, errs := entrySchema["ace_action"].ValidateFunc(action, "ace_action")
			assert.Empty(t, errs, "action '%s' should be valid", action)
		}

		_, errs := entrySchema["ace_action"].ValidateFunc("invalid", "ace_action")
		assert.NotEmpty(t, errs, "action 'invalid' should be invalid")
	})

	t.Run("sequence validation", func(t *testing.T) {
		_, errs := entrySchema["sequence"].ValidateFunc(1, "sequence")
		assert.Empty(t, errs, "sequence 1 should be valid")

		_, errs = entrySchema["sequence"].ValidateFunc(99999, "sequence")
		assert.Empty(t, errs, "sequence 99999 should be valid")

		_, errs = entrySchema["sequence"].ValidateFunc(0, "sequence")
		assert.NotEmpty(t, errs, "sequence 0 should be invalid")

		_, errs = entrySchema["sequence"].ValidateFunc(100000, "sequence")
		assert.NotEmpty(t, errs, "sequence 100000 should be invalid")
	})

	t.Run("vlan_id validation", func(t *testing.T) {
		_, errs := entrySchema["vlan_id"].ValidateFunc(1, "vlan_id")
		assert.Empty(t, errs, "vlan_id 1 should be valid")

		_, errs = entrySchema["vlan_id"].ValidateFunc(4094, "vlan_id")
		assert.Empty(t, errs, "vlan_id 4094 should be valid")

		_, errs = entrySchema["vlan_id"].ValidateFunc(0, "vlan_id")
		assert.NotEmpty(t, errs, "vlan_id 0 should be invalid")

		_, errs = entrySchema["vlan_id"].ValidateFunc(4095, "vlan_id")
		assert.NotEmpty(t, errs, "vlan_id 4095 should be invalid")
	})

	t.Run("filter_id validation", func(t *testing.T) {
		_, errs := entrySchema["filter_id"].ValidateFunc(1, "filter_id")
		assert.Empty(t, errs, "filter_id 1 should be valid")

		_, errs = entrySchema["filter_id"].ValidateFunc(0, "filter_id")
		assert.NotEmpty(t, errs, "filter_id 0 should be invalid")
	})
}

func TestResourceRTXAccessListMACApplySchemaValidation(t *testing.T) {
	resource := resourceRTXAccessListMAC()
	applySchema := resource.Schema["apply"].Elem.(*schema.Resource).Schema

	t.Run("apply interface is required", func(t *testing.T) {
		assert.True(t, applySchema["interface"].Required)
	})

	t.Run("apply direction is required", func(t *testing.T) {
		assert.True(t, applySchema["direction"].Required)
	})

	t.Run("apply direction validation", func(t *testing.T) {
		validDirections := []string{"in", "out"}
		for _, dir := range validDirections {
			_, errs := applySchema["direction"].ValidateFunc(dir, "direction")
			assert.Empty(t, errs, "direction '%s' should be valid", dir)
		}

		_, errs := applySchema["direction"].ValidateFunc("invalid", "direction")
		assert.NotEmpty(t, errs, "direction 'invalid' should be invalid")
	})

	t.Run("apply filter_ids is required", func(t *testing.T) {
		assert.True(t, applySchema["filter_ids"].Required)
	})
}

func TestResourceRTXAccessListMACImporter(t *testing.T) {
	resource := resourceRTXAccessListMAC()

	t.Run("importer is configured", func(t *testing.T) {
		assert.NotNil(t, resource.Importer)
		assert.NotNil(t, resource.Importer.StateContext)
	})
}

func TestResourceRTXAccessListMACCRUDFunctions(t *testing.T) {
	resource := resourceRTXAccessListMAC()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, resource.CreateContext)
		assert.NotNil(t, resource.ReadContext)
		assert.NotNil(t, resource.UpdateContext)
		assert.NotNil(t, resource.DeleteContext)
	})
}
