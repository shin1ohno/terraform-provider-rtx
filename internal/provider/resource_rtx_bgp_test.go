package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func TestBuildBGPConfigFromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.BGPConfig
	}{
		{
			name: "basic BGP config with ASN only",
			input: map[string]interface{}{
				"asn":                    "65001",
				"router_id":              "",
				"default_ipv4_unicast":   false,
				"log_neighbor_changes":   false,
				"redistribute_static":    false,
				"redistribute_connected": false,
				"neighbor":               []interface{}{},
				"network":                []interface{}{},
			},
			expected: client.BGPConfig{
				Enabled: true,
				ASN:     "65001",
			},
		},
		{
			name: "BGP config with router ID and options",
			input: map[string]interface{}{
				"asn":                    "65002",
				"router_id":              "1.1.1.1",
				"default_ipv4_unicast":   true,
				"log_neighbor_changes":   true,
				"redistribute_static":    true,
				"redistribute_connected": true,
				"neighbor":               []interface{}{},
				"network":                []interface{}{},
			},
			expected: client.BGPConfig{
				Enabled:               true,
				ASN:                   "65002",
				RouterID:              "1.1.1.1",
				DefaultIPv4Unicast:    true,
				LogNeighborChanges:    true,
				RedistributeStatic:    true,
				RedistributeConnected: true,
			},
		},
		{
			name: "BGP config with single neighbor",
			input: map[string]interface{}{
				"asn":                    "65003",
				"router_id":              "2.2.2.2",
				"default_ipv4_unicast":   false,
				"log_neighbor_changes":   false,
				"redistribute_static":    false,
				"redistribute_connected": false,
				"neighbor": []interface{}{
					map[string]interface{}{
						"index":         1,
						"ip":            "10.0.0.1",
						"remote_as":     "65100",
						"hold_time":     90,
						"keepalive":     30,
						"multihop":      0,
						"password":      "",
						"local_address": "",
					},
				},
				"network": []interface{}{},
			},
			expected: client.BGPConfig{
				Enabled:  true,
				ASN:      "65003",
				RouterID: "2.2.2.2",
				Neighbors: []client.BGPNeighbor{
					{
						ID:        1,
						IP:        "10.0.0.1",
						RemoteAS:  "65100",
						HoldTime:  90,
						Keepalive: 30,
					},
				},
			},
		},
		{
			name: "BGP config with multiple neighbors",
			input: map[string]interface{}{
				"asn":                    "65004",
				"router_id":              "3.3.3.3",
				"default_ipv4_unicast":   true,
				"log_neighbor_changes":   true,
				"redistribute_static":    false,
				"redistribute_connected": false,
				"neighbor": []interface{}{
					map[string]interface{}{
						"index":         1,
						"ip":            "10.0.0.1",
						"remote_as":     "65100",
						"hold_time":     90,
						"keepalive":     30,
						"multihop":      0,
						"password":      "",
						"local_address": "",
					},
					map[string]interface{}{
						"index":         2,
						"ip":            "10.0.0.2",
						"remote_as":     "65200",
						"hold_time":     180,
						"keepalive":     60,
						"multihop":      2,
						"password":      "secret123",
						"local_address": "192.168.1.1",
					},
				},
				"network": []interface{}{},
			},
			expected: client.BGPConfig{
				Enabled:            true,
				ASN:                "65004",
				RouterID:           "3.3.3.3",
				DefaultIPv4Unicast: true,
				LogNeighborChanges: true,
				Neighbors: []client.BGPNeighbor{
					{
						ID:        1,
						IP:        "10.0.0.1",
						RemoteAS:  "65100",
						HoldTime:  90,
						Keepalive: 30,
					},
					{
						ID:           2,
						IP:           "10.0.0.2",
						RemoteAS:     "65200",
						HoldTime:     180,
						Keepalive:    60,
						Multihop:     2,
						Password:     "secret123",
						LocalAddress: "192.168.1.1",
					},
				},
			},
		},
		{
			name: "BGP config with networks",
			input: map[string]interface{}{
				"asn":                    "65005",
				"router_id":              "4.4.4.4",
				"default_ipv4_unicast":   false,
				"log_neighbor_changes":   false,
				"redistribute_static":    false,
				"redistribute_connected": false,
				"neighbor":               []interface{}{},
				"network": []interface{}{
					map[string]interface{}{
						"prefix": "192.168.0.0",
						"mask":   "255.255.255.0",
					},
					map[string]interface{}{
						"prefix": "10.0.0.0",
						"mask":   "255.0.0.0",
					},
				},
			},
			expected: client.BGPConfig{
				Enabled:  true,
				ASN:      "65005",
				RouterID: "4.4.4.4",
				Networks: []client.BGPNetwork{
					{
						Prefix: "192.168.0.0",
						Mask:   "255.255.255.0",
					},
					{
						Prefix: "10.0.0.0",
						Mask:   "255.0.0.0",
					},
				},
			},
		},
		{
			name: "BGP config with 4-byte ASN",
			input: map[string]interface{}{
				"asn":                    "4200000001",
				"router_id":              "5.5.5.5",
				"default_ipv4_unicast":   false,
				"log_neighbor_changes":   false,
				"redistribute_static":    false,
				"redistribute_connected": false,
				"neighbor":               []interface{}{},
				"network":                []interface{}{},
			},
			expected: client.BGPConfig{
				Enabled:  true,
				ASN:      "4200000001",
				RouterID: "5.5.5.5",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceSchema := resourceRTXBGP().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			result := buildBGPConfigFromResourceData(d)

			assert.Equal(t, tt.expected.Enabled, result.Enabled)
			assert.Equal(t, tt.expected.ASN, result.ASN)
			assert.Equal(t, tt.expected.RouterID, result.RouterID)
			assert.Equal(t, tt.expected.DefaultIPv4Unicast, result.DefaultIPv4Unicast)
			assert.Equal(t, tt.expected.LogNeighborChanges, result.LogNeighborChanges)
			assert.Equal(t, tt.expected.RedistributeStatic, result.RedistributeStatic)
			assert.Equal(t, tt.expected.RedistributeConnected, result.RedistributeConnected)

			assert.Equal(t, len(tt.expected.Neighbors), len(result.Neighbors))
			for i, expectedNeighbor := range tt.expected.Neighbors {
				actualNeighbor := result.Neighbors[i]
				assert.Equal(t, expectedNeighbor.ID, actualNeighbor.ID, "neighbor[%d].ID", i)
				assert.Equal(t, expectedNeighbor.IP, actualNeighbor.IP, "neighbor[%d].IP", i)
				assert.Equal(t, expectedNeighbor.RemoteAS, actualNeighbor.RemoteAS, "neighbor[%d].RemoteAS", i)
				assert.Equal(t, expectedNeighbor.HoldTime, actualNeighbor.HoldTime, "neighbor[%d].HoldTime", i)
				assert.Equal(t, expectedNeighbor.Keepalive, actualNeighbor.Keepalive, "neighbor[%d].Keepalive", i)
				assert.Equal(t, expectedNeighbor.Multihop, actualNeighbor.Multihop, "neighbor[%d].Multihop", i)
				assert.Equal(t, expectedNeighbor.Password, actualNeighbor.Password, "neighbor[%d].Password", i)
				assert.Equal(t, expectedNeighbor.LocalAddress, actualNeighbor.LocalAddress, "neighbor[%d].LocalAddress", i)
			}

			assert.Equal(t, len(tt.expected.Networks), len(result.Networks))
			for i, expectedNetwork := range tt.expected.Networks {
				actualNetwork := result.Networks[i]
				assert.Equal(t, expectedNetwork.Prefix, actualNetwork.Prefix, "network[%d].Prefix", i)
				assert.Equal(t, expectedNetwork.Mask, actualNetwork.Mask, "network[%d].Mask", i)
			}
		})
	}
}

func TestResourceRTXBGPSchema(t *testing.T) {
	resource := resourceRTXBGP()

	t.Run("asn is required", func(t *testing.T) {
		assert.True(t, resource.Schema["asn"].Required)
	})

	t.Run("router_id is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["router_id"].Optional)
	})

	t.Run("default_ipv4_unicast is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["default_ipv4_unicast"].Optional)
		assert.True(t, resource.Schema["default_ipv4_unicast"].Computed)
	})

	t.Run("log_neighbor_changes is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["log_neighbor_changes"].Optional)
		assert.True(t, resource.Schema["log_neighbor_changes"].Computed)
	})

	t.Run("redistribute_static is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["redistribute_static"].Optional)
		assert.True(t, resource.Schema["redistribute_static"].Computed)
	})

	t.Run("redistribute_connected is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["redistribute_connected"].Optional)
		assert.True(t, resource.Schema["redistribute_connected"].Computed)
	})

	t.Run("neighbor is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["neighbor"].Optional)
	})

	t.Run("network is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["network"].Optional)
	})
}

func TestResourceRTXBGPNeighborSchema(t *testing.T) {
	resource := resourceRTXBGP()
	neighborSchema := resource.Schema["neighbor"].Elem.(*schema.Resource).Schema

	t.Run("neighbor index is required", func(t *testing.T) {
		assert.True(t, neighborSchema["index"].Required)
	})

	t.Run("neighbor ip is required", func(t *testing.T) {
		assert.True(t, neighborSchema["ip"].Required)
	})

	t.Run("neighbor remote_as is required", func(t *testing.T) {
		assert.True(t, neighborSchema["remote_as"].Required)
	})

	t.Run("neighbor hold_time is optional", func(t *testing.T) {
		assert.True(t, neighborSchema["hold_time"].Optional)
	})

	t.Run("neighbor keepalive is optional", func(t *testing.T) {
		assert.True(t, neighborSchema["keepalive"].Optional)
	})

	t.Run("neighbor multihop is optional", func(t *testing.T) {
		assert.True(t, neighborSchema["multihop"].Optional)
	})

	t.Run("neighbor password is optional and sensitive", func(t *testing.T) {
		assert.True(t, neighborSchema["password"].Optional)
		assert.True(t, neighborSchema["password"].Sensitive)
	})

	t.Run("neighbor local_address is optional", func(t *testing.T) {
		assert.True(t, neighborSchema["local_address"].Optional)
	})
}

func TestResourceRTXBGPNetworkSchema(t *testing.T) {
	resource := resourceRTXBGP()
	networkSchema := resource.Schema["network"].Elem.(*schema.Resource).Schema

	t.Run("network prefix is required", func(t *testing.T) {
		assert.True(t, networkSchema["prefix"].Required)
	})

	t.Run("network mask is required", func(t *testing.T) {
		assert.True(t, networkSchema["mask"].Required)
	})
}

func TestResourceRTXBGPSchemaValidation(t *testing.T) {
	resource := resourceRTXBGP()
	neighborSchema := resource.Schema["neighbor"].Elem.(*schema.Resource).Schema

	t.Run("asn validation", func(t *testing.T) {
		// Valid ASNs
		validASNs := []string{"1", "65535", "65001", "4294967295", "4200000001"}
		for _, asn := range validASNs {
			_, errs := resource.Schema["asn"].ValidateFunc(asn, "asn")
			assert.Empty(t, errs, "ASN '%s' should be valid", asn)
		}

		// Invalid ASNs
		invalidASNs := []string{"0", "", "abc", "-1"}
		for _, asn := range invalidASNs {
			_, errs := resource.Schema["asn"].ValidateFunc(asn, "asn")
			assert.NotEmpty(t, errs, "ASN '%s' should be invalid", asn)
		}
	})

	t.Run("router_id validation", func(t *testing.T) {
		// Valid router IDs (IPv4 addresses)
		validIPs := []string{"1.1.1.1", "192.168.1.1", "10.0.0.1", ""}
		for _, ip := range validIPs {
			_, errs := resource.Schema["router_id"].ValidateFunc(ip, "router_id")
			assert.Empty(t, errs, "router_id '%s' should be valid", ip)
		}

		// Invalid router IDs
		invalidIPs := []string{"invalid", "256.256.256.256", "2001:db8::1"}
		for _, ip := range invalidIPs {
			_, errs := resource.Schema["router_id"].ValidateFunc(ip, "router_id")
			assert.NotEmpty(t, errs, "router_id '%s' should be invalid", ip)
		}
	})

	t.Run("neighbor index validation", func(t *testing.T) {
		_, errs := neighborSchema["index"].ValidateFunc(1, "index")
		assert.Empty(t, errs, "index 1 should be valid")

		_, errs = neighborSchema["index"].ValidateFunc(0, "index")
		assert.NotEmpty(t, errs, "index 0 should be invalid")
	})

	t.Run("neighbor hold_time validation", func(t *testing.T) {
		_, errs := neighborSchema["hold_time"].ValidateFunc(3, "hold_time")
		assert.Empty(t, errs, "hold_time 3 should be valid")

		_, errs = neighborSchema["hold_time"].ValidateFunc(28800, "hold_time")
		assert.Empty(t, errs, "hold_time 28800 should be valid")

		_, errs = neighborSchema["hold_time"].ValidateFunc(2, "hold_time")
		assert.NotEmpty(t, errs, "hold_time 2 should be invalid")

		_, errs = neighborSchema["hold_time"].ValidateFunc(28801, "hold_time")
		assert.NotEmpty(t, errs, "hold_time 28801 should be invalid")
	})

	t.Run("neighbor keepalive validation", func(t *testing.T) {
		_, errs := neighborSchema["keepalive"].ValidateFunc(1, "keepalive")
		assert.Empty(t, errs, "keepalive 1 should be valid")

		_, errs = neighborSchema["keepalive"].ValidateFunc(21845, "keepalive")
		assert.Empty(t, errs, "keepalive 21845 should be valid")

		_, errs = neighborSchema["keepalive"].ValidateFunc(0, "keepalive")
		assert.NotEmpty(t, errs, "keepalive 0 should be invalid")

		_, errs = neighborSchema["keepalive"].ValidateFunc(21846, "keepalive")
		assert.NotEmpty(t, errs, "keepalive 21846 should be invalid")
	})

	t.Run("neighbor multihop validation", func(t *testing.T) {
		_, errs := neighborSchema["multihop"].ValidateFunc(1, "multihop")
		assert.Empty(t, errs, "multihop 1 should be valid")

		_, errs = neighborSchema["multihop"].ValidateFunc(255, "multihop")
		assert.Empty(t, errs, "multihop 255 should be valid")

		_, errs = neighborSchema["multihop"].ValidateFunc(0, "multihop")
		assert.NotEmpty(t, errs, "multihop 0 should be invalid")

		_, errs = neighborSchema["multihop"].ValidateFunc(256, "multihop")
		assert.NotEmpty(t, errs, "multihop 256 should be invalid")
	})
}

func TestResourceRTXBGPImporter(t *testing.T) {
	resource := resourceRTXBGP()

	t.Run("importer is configured", func(t *testing.T) {
		assert.NotNil(t, resource.Importer)
		assert.NotNil(t, resource.Importer.StateContext)
	})
}

func TestResourceRTXBGPCRUDFunctions(t *testing.T) {
	resource := resourceRTXBGP()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, resource.CreateContext)
		assert.NotNil(t, resource.ReadContext)
		assert.NotNil(t, resource.UpdateContext)
		assert.NotNil(t, resource.DeleteContext)
	})
}

func TestValidateASN(t *testing.T) {
	tests := []struct {
		name    string
		asn     string
		isValid bool
	}{
		{"valid 2-byte ASN", "65001", true},
		{"valid 4-byte ASN", "4200000001", true},
		{"minimum ASN", "1", true},
		{"maximum ASN", "4294967295", true},
		{"zero ASN", "0", false},
		{"empty ASN", "", false},
		{"non-numeric ASN", "abc", false},
		{"negative ASN", "-1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateASN(tt.asn, "asn")
			if tt.isValid {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
			}
		})
	}
}

func TestValidateIPv4Address(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		isValid bool
	}{
		{"valid IP", "192.168.1.1", true},
		{"valid loopback", "127.0.0.1", true},
		{"valid zero address", "0.0.0.0", true},
		{"empty IP (optional)", "", true},
		{"invalid IP", "invalid", false},
		{"out of range IP", "256.256.256.256", false},
		{"IPv6 address", "2001:db8::1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateIPv4Address(tt.ip, "ip")
			if tt.isValid {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
			}
		})
	}
}
