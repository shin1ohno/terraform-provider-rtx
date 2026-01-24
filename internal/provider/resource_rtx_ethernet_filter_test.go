package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func TestBuildEthernetFilterFromResourceData_MACBasedFilter(t *testing.T) {
	tests := []struct {
		name       string
		input      map[string]interface{}
		wantNumber int
		wantAction string
		wantSrcMAC string
		wantDstMAC string
	}{
		{
			name: "basic MAC filter with pass-log",
			input: map[string]interface{}{
				"sequence":        100,
				"action":          "pass-log",
				"source_mac":      "00:11:22:33:44:55",
				"destination_mac": "*",
			},
			wantNumber: 100,
			wantAction: "pass-log",
			wantSrcMAC: "00:11:22:33:44:55",
			wantDstMAC: "*",
		},
		{
			name: "MAC filter with reject-nolog",
			input: map[string]interface{}{
				"sequence":        200,
				"action":          "reject-nolog",
				"source_mac":      "*",
				"destination_mac": "aa:bb:cc:dd:ee:ff",
			},
			wantNumber: 200,
			wantAction: "reject-nolog",
			wantSrcMAC: "*",
			wantDstMAC: "aa:bb:cc:dd:ee:ff",
		},
		{
			name: "MAC filter with both MACs specified",
			input: map[string]interface{}{
				"sequence":        300,
				"action":          "pass-nolog",
				"source_mac":      "11:22:33:44:55:66",
				"destination_mac": "66:55:44:33:22:11",
			},
			wantNumber: 300,
			wantAction: "pass-nolog",
			wantSrcMAC: "11:22:33:44:55:66",
			wantDstMAC: "66:55:44:33:22:11",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceRTXEthernetFilter().Schema, tt.input)
			filter := buildEthernetFilterFromResourceData(d)

			assert.Equal(t, tt.wantNumber, filter.Number)
			assert.Equal(t, tt.wantAction, filter.Action)
			assert.Equal(t, tt.wantSrcMAC, filter.SourceMAC)
			assert.Equal(t, tt.wantDstMAC, filter.DestMAC)
		})
	}
}

func TestBuildEthernetFilterFromResourceData_WithEtherType(t *testing.T) {
	input := map[string]interface{}{
		"sequence":        100,
		"action":          "pass-log",
		"source_mac":      "*",
		"destination_mac": "*",
		"ether_type":      "0x0800",
	}

	d := schema.TestResourceDataRaw(t, resourceRTXEthernetFilter().Schema, input)
	filter := buildEthernetFilterFromResourceData(d)

	assert.Equal(t, 100, filter.Number)
	assert.Equal(t, "pass-log", filter.Action)
	assert.Equal(t, "0x0800", filter.EtherType)
}

func TestBuildEthernetFilterFromResourceData_WithVlanID(t *testing.T) {
	input := map[string]interface{}{
		"sequence":        100,
		"action":          "reject-log",
		"source_mac":      "*",
		"destination_mac": "*",
		"vlan_id":         100,
	}

	d := schema.TestResourceDataRaw(t, resourceRTXEthernetFilter().Schema, input)
	filter := buildEthernetFilterFromResourceData(d)

	assert.Equal(t, 100, filter.Number)
	assert.Equal(t, "reject-log", filter.Action)
	assert.Equal(t, 100, filter.VlanID)
}

func TestResourceRTXEthernetFilterSchema(t *testing.T) {
	resource := resourceRTXEthernetFilter()

	// Verify required fields
	assert.NotNil(t, resource.Schema["sequence"])
	assert.True(t, resource.Schema["sequence"].Required)
	assert.True(t, resource.Schema["sequence"].ForceNew)

	assert.NotNil(t, resource.Schema["action"])
	assert.True(t, resource.Schema["action"].Required)

	// Verify optional fields
	assert.NotNil(t, resource.Schema["source_mac"])
	assert.True(t, resource.Schema["source_mac"].Optional)

	assert.NotNil(t, resource.Schema["destination_mac"])
	assert.True(t, resource.Schema["destination_mac"].Optional)

	assert.NotNil(t, resource.Schema["ether_type"])
	assert.True(t, resource.Schema["ether_type"].Optional)

	assert.NotNil(t, resource.Schema["vlan_id"])
	assert.True(t, resource.Schema["vlan_id"].Optional)

	assert.NotNil(t, resource.Schema["dhcp_type"])
	assert.True(t, resource.Schema["dhcp_type"].Optional)

	assert.NotNil(t, resource.Schema["dhcp_scope"])
	assert.True(t, resource.Schema["dhcp_scope"].Optional)

	// Verify ConflictsWith settings
	assert.Contains(t, resource.Schema["source_mac"].ConflictsWith, "dhcp_type")
	assert.Contains(t, resource.Schema["dhcp_type"].ConflictsWith, "source_mac")
}

func TestFlattenEthernetFilterToResourceData(t *testing.T) {
	tests := []struct {
		name      string
		sequence  int
		action    string
		sourceMAC string
		destMAC   string
		etherType string
		vlanID    int
	}{
		{
			name:      "basic MAC filter",
			sequence:  100,
			action:    "pass-log",
			sourceMAC: "00:11:22:33:44:55",
			destMAC:   "*",
		},
		{
			name:      "filter with ether_type",
			sequence:  200,
			action:    "reject-nolog",
			sourceMAC: "*",
			destMAC:   "*",
			etherType: "0x0806",
		},
		{
			name:      "filter with vlan_id",
			sequence:  300,
			action:    "pass-nolog",
			sourceMAC: "*",
			destMAC:   "*",
			vlanID:    200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create filter
			filter := &client.EthernetFilter{
				Number:    tt.sequence,
				Action:    tt.action,
				SourceMAC: tt.sourceMAC,
				DestMAC:   tt.destMAC,
				EtherType: tt.etherType,
				VlanID:    tt.vlanID,
			}

			// Create empty resource data
			input := map[string]interface{}{
				"sequence": 0,
				"action":   "",
			}
			d := schema.TestResourceDataRaw(t, resourceRTXEthernetFilter().Schema, input)

			// Flatten
			err := flattenEthernetFilterToResourceData(filter, d)
			assert.NoError(t, err)

			// Verify
			assert.Equal(t, tt.sequence, d.Get("sequence"))
			assert.Equal(t, tt.action, d.Get("action"))
			if tt.sourceMAC != "" {
				assert.Equal(t, tt.sourceMAC, d.Get("source_mac"))
			}
			if tt.destMAC != "" {
				assert.Equal(t, tt.destMAC, d.Get("destination_mac"))
			}
			if tt.etherType != "" {
				assert.Equal(t, tt.etherType, d.Get("ether_type"))
			}
			if tt.vlanID > 0 {
				assert.Equal(t, tt.vlanID, d.Get("vlan_id"))
			}
		})
	}
}
