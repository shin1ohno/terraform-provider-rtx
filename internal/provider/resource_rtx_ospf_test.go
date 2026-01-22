package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestBuildOSPFConfigFromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.OSPFConfig
	}{
		{
			name: "basic OSPF with router ID only",
			input: map[string]interface{}{
				"process_id":                    0,
				"router_id":                     "1.1.1.1",
				"distance":                      0,
				"default_information_originate": false,
				"redistribute_static":           false,
				"redistribute_connected":        false,
				"network":                       []interface{}{},
				"area":                          []interface{}{},
				"neighbor":                      []interface{}{},
			},
			expected: client.OSPFConfig{
				Enabled:  true,
				RouterID: "1.1.1.1",
			},
		},
		{
			name: "OSPF with single network in backbone area",
			input: map[string]interface{}{
				"process_id":                    1,
				"router_id":                     "10.0.0.1",
				"distance":                      110,
				"default_information_originate": true,
				"redistribute_static":           true,
				"redistribute_connected":        true,
				"network": []interface{}{
					map[string]interface{}{
						"ip":       "192.168.1.0",
						"wildcard": "0.0.0.255",
						"area":     "0",
					},
				},
				"area":     []interface{}{},
				"neighbor": []interface{}{},
			},
			expected: client.OSPFConfig{
				Enabled:               true,
				ProcessID:             1,
				RouterID:              "10.0.0.1",
				Distance:              110,
				DefaultOriginate:      true,
				RedistributeStatic:    true,
				RedistributeConnected: true,
				Networks: []client.OSPFNetwork{
					{IP: "192.168.1.0", Wildcard: "0.0.0.255", Area: "0"},
				},
			},
		},
		{
			name: "OSPF with multiple networks and areas",
			input: map[string]interface{}{
				"process_id":                    1,
				"router_id":                     "10.0.0.2",
				"distance":                      0,
				"default_information_originate": false,
				"redistribute_static":           false,
				"redistribute_connected":        false,
				"network": []interface{}{
					map[string]interface{}{
						"ip":       "192.168.1.0",
						"wildcard": "0.0.0.255",
						"area":     "0.0.0.0",
					},
					map[string]interface{}{
						"ip":       "10.10.0.0",
						"wildcard": "0.0.255.255",
						"area":     "1",
					},
					map[string]interface{}{
						"ip":       "172.16.0.0",
						"wildcard": "0.0.0.255",
						"area":     "2",
					},
				},
				"area": []interface{}{
					map[string]interface{}{
						"id":         "0",
						"type":       "normal",
						"no_summary": false,
					},
					map[string]interface{}{
						"id":         "1",
						"type":       "stub",
						"no_summary": false,
					},
					map[string]interface{}{
						"id":         "2",
						"type":       "nssa",
						"no_summary": true,
					},
				},
				"neighbor": []interface{}{},
			},
			expected: client.OSPFConfig{
				Enabled:   true,
				ProcessID: 1,
				RouterID:  "10.0.0.2",
				Networks: []client.OSPFNetwork{
					{IP: "192.168.1.0", Wildcard: "0.0.0.255", Area: "0.0.0.0"},
					{IP: "10.10.0.0", Wildcard: "0.0.255.255", Area: "1"},
					{IP: "172.16.0.0", Wildcard: "0.0.0.255", Area: "2"},
				},
				Areas: []client.OSPFArea{
					{ID: "0", Type: "normal", NoSummary: false},
					{ID: "1", Type: "stub", NoSummary: false},
					{ID: "2", Type: "nssa", NoSummary: true},
				},
			},
		},
		{
			name: "OSPF with NBMA neighbors",
			input: map[string]interface{}{
				"process_id":                    1,
				"router_id":                     "172.16.0.1",
				"distance":                      0,
				"default_information_originate": false,
				"redistribute_static":           false,
				"redistribute_connected":        false,
				"network":                       []interface{}{},
				"area":                          []interface{}{},
				"neighbor": []interface{}{
					map[string]interface{}{
						"ip":       "172.16.0.2",
						"priority": 1,
						"cost":     10,
					},
					map[string]interface{}{
						"ip":       "172.16.0.3",
						"priority": 0,
						"cost":     20,
					},
				},
			},
			expected: client.OSPFConfig{
				Enabled:   true,
				ProcessID: 1,
				RouterID:  "172.16.0.1",
				Neighbors: []client.OSPFNeighbor{
					{IP: "172.16.0.2", Priority: 1, Cost: 10},
					{IP: "172.16.0.3", Priority: 0, Cost: 20},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceSchema := resourceRTXOSPF().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			result := buildOSPFConfigFromResourceData(d)

			assert.Equal(t, tt.expected.Enabled, result.Enabled)
			assert.Equal(t, tt.expected.ProcessID, result.ProcessID)
			assert.Equal(t, tt.expected.RouterID, result.RouterID)
			assert.Equal(t, tt.expected.Distance, result.Distance)
			assert.Equal(t, tt.expected.DefaultOriginate, result.DefaultOriginate)
			assert.Equal(t, tt.expected.RedistributeStatic, result.RedistributeStatic)
			assert.Equal(t, tt.expected.RedistributeConnected, result.RedistributeConnected)

			// Check networks
			assert.Equal(t, len(tt.expected.Networks), len(result.Networks))
			for i, expectedNet := range tt.expected.Networks {
				assert.Equal(t, expectedNet.IP, result.Networks[i].IP, "network[%d].IP", i)
				assert.Equal(t, expectedNet.Wildcard, result.Networks[i].Wildcard, "network[%d].Wildcard", i)
				assert.Equal(t, expectedNet.Area, result.Networks[i].Area, "network[%d].Area", i)
			}

			// Check areas
			assert.Equal(t, len(tt.expected.Areas), len(result.Areas))
			for i, expectedArea := range tt.expected.Areas {
				assert.Equal(t, expectedArea.ID, result.Areas[i].ID, "area[%d].ID", i)
				assert.Equal(t, expectedArea.Type, result.Areas[i].Type, "area[%d].Type", i)
				assert.Equal(t, expectedArea.NoSummary, result.Areas[i].NoSummary, "area[%d].NoSummary", i)
			}

			// Check neighbors
			assert.Equal(t, len(tt.expected.Neighbors), len(result.Neighbors))
			for i, expectedNeighbor := range tt.expected.Neighbors {
				assert.Equal(t, expectedNeighbor.IP, result.Neighbors[i].IP, "neighbor[%d].IP", i)
				assert.Equal(t, expectedNeighbor.Priority, result.Neighbors[i].Priority, "neighbor[%d].Priority", i)
				assert.Equal(t, expectedNeighbor.Cost, result.Neighbors[i].Cost, "neighbor[%d].Cost", i)
			}
		})
	}
}

func TestResourceRTXOSPFSchema(t *testing.T) {
	resource := resourceRTXOSPF()

	t.Run("router_id is required", func(t *testing.T) {
		assert.True(t, resource.Schema["router_id"].Required)
	})

	t.Run("process_id is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["process_id"].Optional)
		assert.True(t, resource.Schema["process_id"].Computed)
	})

	t.Run("distance is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["distance"].Optional)
		assert.True(t, resource.Schema["distance"].Computed)
	})

	t.Run("default_information_originate is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["default_information_originate"].Optional)
		assert.True(t, resource.Schema["default_information_originate"].Computed)
	})

	t.Run("redistribute options are optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["redistribute_static"].Optional)
		assert.True(t, resource.Schema["redistribute_static"].Computed)
		assert.True(t, resource.Schema["redistribute_connected"].Optional)
		assert.True(t, resource.Schema["redistribute_connected"].Computed)
	})

	t.Run("network is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["network"].Optional)
	})

	t.Run("area is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["area"].Optional)
	})

	t.Run("neighbor is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["neighbor"].Optional)
	})
}

func TestResourceRTXOSPFSchemaValidation(t *testing.T) {
	resource := resourceRTXOSPF()

	t.Run("process_id validation", func(t *testing.T) {
		_, errs := resource.Schema["process_id"].ValidateFunc(1, "process_id")
		assert.Empty(t, errs, "process_id 1 should be valid")

		_, errs = resource.Schema["process_id"].ValidateFunc(0, "process_id")
		assert.NotEmpty(t, errs, "process_id 0 should be invalid")
	})

	t.Run("distance validation", func(t *testing.T) {
		_, errs := resource.Schema["distance"].ValidateFunc(1, "distance")
		assert.Empty(t, errs, "distance 1 should be valid")

		_, errs = resource.Schema["distance"].ValidateFunc(255, "distance")
		assert.Empty(t, errs, "distance 255 should be valid")

		_, errs = resource.Schema["distance"].ValidateFunc(0, "distance")
		assert.NotEmpty(t, errs, "distance 0 should be invalid")

		_, errs = resource.Schema["distance"].ValidateFunc(256, "distance")
		assert.NotEmpty(t, errs, "distance 256 should be invalid")
	})
}

func TestResourceRTXOSPFNetworkSchema(t *testing.T) {
	resource := resourceRTXOSPF()
	networkSchema := resource.Schema["network"].Elem.(*schema.Resource).Schema

	t.Run("ip is required", func(t *testing.T) {
		assert.True(t, networkSchema["ip"].Required)
	})

	t.Run("wildcard is required", func(t *testing.T) {
		assert.True(t, networkSchema["wildcard"].Required)
	})

	t.Run("area is required", func(t *testing.T) {
		assert.True(t, networkSchema["area"].Required)
	})
}

func TestResourceRTXOSPFAreaSchema(t *testing.T) {
	resource := resourceRTXOSPF()
	areaSchema := resource.Schema["area"].Elem.(*schema.Resource).Schema

	t.Run("id is required", func(t *testing.T) {
		assert.True(t, areaSchema["id"].Required)
	})

	t.Run("type is optional and computed", func(t *testing.T) {
		assert.True(t, areaSchema["type"].Optional)
		assert.True(t, areaSchema["type"].Computed)
	})

	t.Run("type validation", func(t *testing.T) {
		validTypes := []string{"normal", "stub", "nssa"}
		for _, areaType := range validTypes {
			_, errs := areaSchema["type"].ValidateFunc(areaType, "type")
			assert.Empty(t, errs, "%s should be valid", areaType)
		}

		_, errs := areaSchema["type"].ValidateFunc("invalid", "type")
		assert.NotEmpty(t, errs, "invalid should be rejected")
	})

	t.Run("no_summary is optional and computed", func(t *testing.T) {
		assert.True(t, areaSchema["no_summary"].Optional)
		assert.True(t, areaSchema["no_summary"].Computed)
	})
}

func TestResourceRTXOSPFNeighborSchema(t *testing.T) {
	resource := resourceRTXOSPF()
	neighborSchema := resource.Schema["neighbor"].Elem.(*schema.Resource).Schema

	t.Run("ip is required", func(t *testing.T) {
		assert.True(t, neighborSchema["ip"].Required)
	})

	t.Run("priority is optional and computed", func(t *testing.T) {
		assert.True(t, neighborSchema["priority"].Optional)
		assert.True(t, neighborSchema["priority"].Computed)
	})

	t.Run("priority validation", func(t *testing.T) {
		_, errs := neighborSchema["priority"].ValidateFunc(0, "priority")
		assert.Empty(t, errs, "priority 0 should be valid")

		_, errs = neighborSchema["priority"].ValidateFunc(255, "priority")
		assert.Empty(t, errs, "priority 255 should be valid")

		_, errs = neighborSchema["priority"].ValidateFunc(-1, "priority")
		assert.NotEmpty(t, errs, "priority -1 should be invalid")

		_, errs = neighborSchema["priority"].ValidateFunc(256, "priority")
		assert.NotEmpty(t, errs, "priority 256 should be invalid")
	})

	t.Run("cost is optional and computed", func(t *testing.T) {
		assert.True(t, neighborSchema["cost"].Optional)
		assert.True(t, neighborSchema["cost"].Computed)
	})
}

func TestResourceRTXOSPFImporter(t *testing.T) {
	resource := resourceRTXOSPF()

	t.Run("importer is configured", func(t *testing.T) {
		assert.NotNil(t, resource.Importer)
		assert.NotNil(t, resource.Importer.StateContext)
	})
}

func TestResourceRTXOSPFCRUDFunctions(t *testing.T) {
	resource := resourceRTXOSPF()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, resource.CreateContext)
		assert.NotNil(t, resource.ReadContext)
		assert.NotNil(t, resource.UpdateContext)
		assert.NotNil(t, resource.DeleteContext)
	})
}
