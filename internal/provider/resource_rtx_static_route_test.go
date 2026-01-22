package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func TestBuildStaticRouteFromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.StaticRoute
	}{
		{
			name: "default route with single gateway",
			input: map[string]interface{}{
				"prefix": "0.0.0.0",
				"mask":   "0.0.0.0",
				"next_hop": []interface{}{
					map[string]interface{}{
						"gateway":   "192.168.1.1",
						"interface": "",
						"distance":  1,
						"permanent": false,
						"filter":    0,
					},
				},
			},
			expected: client.StaticRoute{
				Prefix: "0.0.0.0",
				Mask:   "0.0.0.0",
				NextHops: []client.StaticRouteHop{
					{NextHop: "192.168.1.1", Distance: 1},
				},
			},
		},
		{
			name: "network route with interface",
			input: map[string]interface{}{
				"prefix": "10.0.0.0",
				"mask":   "255.0.0.0",
				"next_hop": []interface{}{
					map[string]interface{}{
						"gateway":   "",
						"interface": "pp 1",
						"distance":  10,
						"permanent": true,
						"filter":    0,
					},
				},
			},
			expected: client.StaticRoute{
				Prefix: "10.0.0.0",
				Mask:   "255.0.0.0",
				NextHops: []client.StaticRouteHop{
					{Interface: "pp 1", Distance: 10, Permanent: true},
				},
			},
		},
		{
			name: "route with tunnel interface",
			input: map[string]interface{}{
				"prefix": "172.16.0.0",
				"mask":   "255.255.0.0",
				"next_hop": []interface{}{
					map[string]interface{}{
						"gateway":   "",
						"interface": "tunnel 1",
						"distance":  50,
						"permanent": false,
						"filter":    100,
					},
				},
			},
			expected: client.StaticRoute{
				Prefix: "172.16.0.0",
				Mask:   "255.255.0.0",
				NextHops: []client.StaticRouteHop{
					{Interface: "tunnel 1", Distance: 50, Filter: 100},
				},
			},
		},
		{
			name: "route with multiple next hops (ECMP)",
			input: map[string]interface{}{
				"prefix": "192.168.100.0",
				"mask":   "255.255.255.0",
				"next_hop": []interface{}{
					map[string]interface{}{
						"gateway":   "10.0.0.1",
						"interface": "",
						"distance":  1,
						"permanent": false,
						"filter":    0,
					},
					map[string]interface{}{
						"gateway":   "10.0.0.2",
						"interface": "",
						"distance":  1,
						"permanent": false,
						"filter":    0,
					},
				},
			},
			expected: client.StaticRoute{
				Prefix: "192.168.100.0",
				Mask:   "255.255.255.0",
				NextHops: []client.StaticRouteHop{
					{NextHop: "10.0.0.1", Distance: 1},
					{NextHop: "10.0.0.2", Distance: 1},
				},
			},
		},
		{
			name: "route with failover next hops",
			input: map[string]interface{}{
				"prefix": "10.10.10.0",
				"mask":   "255.255.255.0",
				"next_hop": []interface{}{
					map[string]interface{}{
						"gateway":   "192.168.1.1",
						"interface": "",
						"distance":  10,
						"permanent": true,
						"filter":    0,
					},
					map[string]interface{}{
						"gateway":   "192.168.2.1",
						"interface": "",
						"distance":  20,
						"permanent": false,
						"filter":    0,
					},
				},
			},
			expected: client.StaticRoute{
				Prefix: "10.10.10.0",
				Mask:   "255.255.255.0",
				NextHops: []client.StaticRouteHop{
					{NextHop: "192.168.1.1", Distance: 10, Permanent: true},
					{NextHop: "192.168.2.1", Distance: 20},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceSchema := resourceRTXStaticRoute().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			result := buildStaticRouteFromResourceData(d)

			assert.Equal(t, tt.expected.Prefix, result.Prefix)
			assert.Equal(t, tt.expected.Mask, result.Mask)
			assert.Equal(t, len(tt.expected.NextHops), len(result.NextHops))

			for i, expectedHop := range tt.expected.NextHops {
				assert.Equal(t, expectedHop.NextHop, result.NextHops[i].NextHop, "next_hop[%d].NextHop", i)
				assert.Equal(t, expectedHop.Interface, result.NextHops[i].Interface, "next_hop[%d].Interface", i)
				assert.Equal(t, expectedHop.Distance, result.NextHops[i].Distance, "next_hop[%d].Distance", i)
				assert.Equal(t, expectedHop.Permanent, result.NextHops[i].Permanent, "next_hop[%d].Permanent", i)
				assert.Equal(t, expectedHop.Filter, result.NextHops[i].Filter, "next_hop[%d].Filter", i)
			}
		})
	}
}

func TestParseStaticRouteID(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		wantPrefix  string
		wantMask    string
		expectError bool
	}{
		{
			name:       "default route",
			id:         "0.0.0.0/0.0.0.0",
			wantPrefix: "0.0.0.0",
			wantMask:   "0.0.0.0",
		},
		{
			name:       "class A network",
			id:         "10.0.0.0/255.0.0.0",
			wantPrefix: "10.0.0.0",
			wantMask:   "255.0.0.0",
		},
		{
			name:       "class C network",
			id:         "192.168.1.0/255.255.255.0",
			wantPrefix: "192.168.1.0",
			wantMask:   "255.255.255.0",
		},
		{
			name:        "invalid format - no separator",
			id:          "10.0.0.0255.0.0.0",
			expectError: true,
		},
		{
			name:        "invalid prefix",
			id:          "invalid/255.0.0.0",
			expectError: true,
		},
		{
			name:        "invalid mask",
			id:          "10.0.0.0/invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, mask, err := parseStaticRouteID(tt.id)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantPrefix, prefix)
				assert.Equal(t, tt.wantMask, mask)
			}
		})
	}
}

func TestResourceRTXStaticRouteSchema(t *testing.T) {
	resource := resourceRTXStaticRoute()

	t.Run("prefix is required and ForceNew", func(t *testing.T) {
		assert.True(t, resource.Schema["prefix"].Required)
		assert.True(t, resource.Schema["prefix"].ForceNew)
	})

	t.Run("mask is required and ForceNew", func(t *testing.T) {
		assert.True(t, resource.Schema["mask"].Required)
		assert.True(t, resource.Schema["mask"].ForceNew)
	})

	t.Run("next_hop is required with MinItems 1", func(t *testing.T) {
		assert.True(t, resource.Schema["next_hop"].Required)
		assert.Equal(t, 1, resource.Schema["next_hop"].MinItems)
	})
}

func TestResourceRTXStaticRouteNextHopSchema(t *testing.T) {
	resource := resourceRTXStaticRoute()
	hopSchema := resource.Schema["next_hop"].Elem.(*schema.Resource).Schema

	t.Run("gateway is optional", func(t *testing.T) {
		assert.True(t, hopSchema["gateway"].Optional)
	})

	t.Run("interface is optional", func(t *testing.T) {
		assert.True(t, hopSchema["interface"].Optional)
	})

	t.Run("distance is optional and computed", func(t *testing.T) {
		assert.True(t, hopSchema["distance"].Optional)
		assert.True(t, hopSchema["distance"].Computed)
	})

	t.Run("distance validation", func(t *testing.T) {
		_, errs := hopSchema["distance"].ValidateFunc(1, "distance")
		assert.Empty(t, errs, "distance 1 should be valid")

		_, errs = hopSchema["distance"].ValidateFunc(100, "distance")
		assert.Empty(t, errs, "distance 100 should be valid")

		_, errs = hopSchema["distance"].ValidateFunc(0, "distance")
		assert.NotEmpty(t, errs, "distance 0 should be invalid")

		_, errs = hopSchema["distance"].ValidateFunc(101, "distance")
		assert.NotEmpty(t, errs, "distance 101 should be invalid")
	})

	t.Run("permanent is optional and computed", func(t *testing.T) {
		assert.True(t, hopSchema["permanent"].Optional)
		assert.True(t, hopSchema["permanent"].Computed)
	})

	t.Run("filter is optional and computed", func(t *testing.T) {
		assert.True(t, hopSchema["filter"].Optional)
		assert.True(t, hopSchema["filter"].Computed)
	})
}

func TestValidateRoutePrefix(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		isValid bool
	}{
		{"default route", "0.0.0.0", true},
		{"class A", "10.0.0.0", true},
		{"class B", "172.16.0.0", true},
		{"class C", "192.168.1.0", true},
		{"empty", "", false},
		{"invalid text", "invalid", false},
		{"IPv6 not allowed", "2001:db8::1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateRoutePrefix(tt.prefix, "prefix")
			if tt.isValid {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
			}
		})
	}
}

func TestValidateRouteMask(t *testing.T) {
	tests := []struct {
		name    string
		mask    string
		isValid bool
	}{
		{"all zeros", "0.0.0.0", true},
		{"class A", "255.0.0.0", true},
		{"class B", "255.255.0.0", true},
		{"class C", "255.255.255.0", true},
		{"/32 host", "255.255.255.255", true},
		{"/25", "255.255.255.128", true},
		{"empty", "", false},
		{"invalid text", "invalid", false},
		{"non-contiguous", "255.0.255.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateRouteMask(tt.mask, "mask")
			if tt.isValid {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
			}
		})
	}
}

func TestValidateOptionalIPAddressRoute(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		isValid bool
	}{
		{"valid IP", "192.168.1.1", true},
		{"empty (optional)", "", true},
		{"invalid text", "invalid", false},
		{"IPv6 not allowed", "2001:db8::1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateOptionalIPAddress(tt.ip, "ip")
			if tt.isValid {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
			}
		})
	}
}

func TestCidrPrefixToMask(t *testing.T) {
	tests := []struct {
		prefixLen int
		expected  string
	}{
		{0, "0.0.0.0"},
		{8, "255.0.0.0"},
		{16, "255.255.0.0"},
		{24, "255.255.255.0"},
		{32, "255.255.255.255"},
		{25, "255.255.255.128"},
		{-1, ""},
		{33, ""},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := cidrPrefixToMask(tt.prefixLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMaskToCIDRPrefix(t *testing.T) {
	tests := []struct {
		mask     string
		expected int
	}{
		{"0.0.0.0", 0},
		{"255.0.0.0", 8},
		{"255.255.0.0", 16},
		{"255.255.255.0", 24},
		{"255.255.255.255", 32},
		{"255.255.255.128", 25},
		{"invalid", -1},
	}

	for _, tt := range tests {
		t.Run(tt.mask, func(t *testing.T) {
			result := maskToCIDRPrefix(tt.mask)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResourceRTXStaticRouteImporter(t *testing.T) {
	resource := resourceRTXStaticRoute()

	t.Run("importer is configured", func(t *testing.T) {
		assert.NotNil(t, resource.Importer)
		assert.NotNil(t, resource.Importer.StateContext)
	})
}

func TestResourceRTXStaticRouteCRUDFunctions(t *testing.T) {
	resource := resourceRTXStaticRoute()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, resource.CreateContext)
		assert.NotNil(t, resource.ReadContext)
		assert.NotNil(t, resource.UpdateContext)
		assert.NotNil(t, resource.DeleteContext)
	})
}
