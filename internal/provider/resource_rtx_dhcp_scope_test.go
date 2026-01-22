package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func TestBuildDHCPScopeFromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.DHCPScope
	}{
		{
			name: "basic DHCP scope with network only",
			input: map[string]interface{}{
				"scope_id":       1,
				"network":        "192.168.1.0/24",
				"range_start":    "",
				"range_end":      "",
				"lease_time":     "",
				"exclude_ranges": []interface{}{},
				"options":        []interface{}{},
			},
			expected: client.DHCPScope{
				ScopeID: 1,
				Network: "192.168.1.0/24",
			},
		},
		{
			name: "DHCP scope with lease time",
			input: map[string]interface{}{
				"scope_id":       2,
				"network":        "10.0.0.0/24",
				"range_start":    "",
				"range_end":      "",
				"lease_time":     "72h",
				"exclude_ranges": []interface{}{},
				"options":        []interface{}{},
			},
			expected: client.DHCPScope{
				ScopeID:   2,
				Network:   "10.0.0.0/24",
				LeaseTime: "72h",
			},
		},
		{
			name: "DHCP scope with options",
			input: map[string]interface{}{
				"scope_id":       3,
				"network":        "172.16.0.0/24",
				"range_start":    "",
				"range_end":      "",
				"lease_time":     "24h",
				"exclude_ranges": []interface{}{},
				"options": []interface{}{
					map[string]interface{}{
						"routers":     []interface{}{"172.16.0.1"},
						"dns_servers": []interface{}{"8.8.8.8", "8.8.4.4"},
						"domain_name": "example.com",
					},
				},
			},
			expected: client.DHCPScope{
				ScopeID:   3,
				Network:   "172.16.0.0/24",
				LeaseTime: "24h",
				Options: client.DHCPScopeOptions{
					Routers:    []string{"172.16.0.1"},
					DNSServers: []string{"8.8.8.8", "8.8.4.4"},
					DomainName: "example.com",
				},
			},
		},
		{
			name: "DHCP scope with exclude ranges",
			input: map[string]interface{}{
				"scope_id":    4,
				"network":     "192.168.10.0/24",
				"range_start": "",
				"range_end":   "",
				"lease_time":  "",
				"exclude_ranges": []interface{}{
					map[string]interface{}{
						"start": "192.168.10.1",
						"end":   "192.168.10.10",
					},
					map[string]interface{}{
						"start": "192.168.10.250",
						"end":   "192.168.10.254",
					},
				},
				"options": []interface{}{},
			},
			expected: client.DHCPScope{
				ScopeID: 4,
				Network: "192.168.10.0/24",
				ExcludeRanges: []client.ExcludeRange{
					{Start: "192.168.10.1", End: "192.168.10.10"},
					{Start: "192.168.10.250", End: "192.168.10.254"},
				},
			},
		},
		{
			name: "DHCP scope with multiple routers",
			input: map[string]interface{}{
				"scope_id":       5,
				"network":        "10.10.0.0/24",
				"range_start":    "",
				"range_end":      "",
				"lease_time":     "infinite",
				"exclude_ranges": []interface{}{},
				"options": []interface{}{
					map[string]interface{}{
						"routers":     []interface{}{"10.10.0.1", "10.10.0.2", "10.10.0.3"},
						"dns_servers": []interface{}{"10.10.0.1"},
						"domain_name": "",
					},
				},
			},
			expected: client.DHCPScope{
				ScopeID:   5,
				Network:   "10.10.0.0/24",
				LeaseTime: "infinite",
				Options: client.DHCPScopeOptions{
					Routers:    []string{"10.10.0.1", "10.10.0.2", "10.10.0.3"},
					DNSServers: []string{"10.10.0.1"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceSchema := resourceRTXDHCPScope().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			result := buildDHCPScopeFromResourceData(d)

			assert.Equal(t, tt.expected.ScopeID, result.ScopeID)
			assert.Equal(t, tt.expected.Network, result.Network)
			assert.Equal(t, tt.expected.LeaseTime, result.LeaseTime)

			assert.Equal(t, len(tt.expected.ExcludeRanges), len(result.ExcludeRanges))
			for i, expectedRange := range tt.expected.ExcludeRanges {
				assert.Equal(t, expectedRange.Start, result.ExcludeRanges[i].Start, "exclude_ranges[%d].Start", i)
				assert.Equal(t, expectedRange.End, result.ExcludeRanges[i].End, "exclude_ranges[%d].End", i)
			}

			assert.Equal(t, tt.expected.Options.Routers, result.Options.Routers)
			assert.Equal(t, tt.expected.Options.DNSServers, result.Options.DNSServers)
			assert.Equal(t, tt.expected.Options.DomainName, result.Options.DomainName)
		})
	}
}

func TestResourceRTXDHCPScopeSchema(t *testing.T) {
	resource := resourceRTXDHCPScope()

	t.Run("scope_id is required and ForceNew", func(t *testing.T) {
		assert.True(t, resource.Schema["scope_id"].Required)
		assert.True(t, resource.Schema["scope_id"].ForceNew)
	})

	t.Run("network is required and ForceNew", func(t *testing.T) {
		assert.True(t, resource.Schema["network"].Required)
		assert.True(t, resource.Schema["network"].ForceNew)
	})

	t.Run("range_start is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["range_start"].Optional)
		assert.True(t, resource.Schema["range_start"].Computed)
	})

	t.Run("range_end is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["range_end"].Optional)
		assert.True(t, resource.Schema["range_end"].Computed)
	})

	t.Run("lease_time is optional and computed", func(t *testing.T) {
		assert.True(t, resource.Schema["lease_time"].Optional)
		assert.True(t, resource.Schema["lease_time"].Computed)
	})

	t.Run("exclude_ranges is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["exclude_ranges"].Optional)
	})

	t.Run("options is optional with MaxItems 1", func(t *testing.T) {
		assert.True(t, resource.Schema["options"].Optional)
		assert.Equal(t, 1, resource.Schema["options"].MaxItems)
	})
}

func TestResourceRTXDHCPScopeSchemaValidation(t *testing.T) {
	resource := resourceRTXDHCPScope()

	t.Run("scope_id validation", func(t *testing.T) {
		_, errs := resource.Schema["scope_id"].ValidateFunc(1, "scope_id")
		assert.Empty(t, errs, "scope_id 1 should be valid")

		_, errs = resource.Schema["scope_id"].ValidateFunc(0, "scope_id")
		assert.NotEmpty(t, errs, "scope_id 0 should be invalid")
	})

	t.Run("network validation (CIDR)", func(t *testing.T) {
		_, errs := resource.Schema["network"].ValidateFunc("192.168.1.0/24", "network")
		assert.Empty(t, errs, "valid CIDR should be accepted")

		_, errs = resource.Schema["network"].ValidateFunc("10.0.0.0/8", "network")
		assert.Empty(t, errs, "valid CIDR should be accepted")

		_, errs = resource.Schema["network"].ValidateFunc("invalid", "network")
		assert.NotEmpty(t, errs, "invalid CIDR should be rejected")
	})

	t.Run("range_start validation (IP address)", func(t *testing.T) {
		_, errs := resource.Schema["range_start"].ValidateFunc("192.168.1.1", "range_start")
		assert.Empty(t, errs, "valid IP should be accepted")

		_, errs = resource.Schema["range_start"].ValidateFunc("", "range_start")
		assert.Empty(t, errs, "empty should be accepted (optional)")

		_, errs = resource.Schema["range_start"].ValidateFunc("invalid", "range_start")
		assert.NotEmpty(t, errs, "invalid IP should be rejected")
	})

	t.Run("lease_time validation", func(t *testing.T) {
		_, errs := resource.Schema["lease_time"].ValidateFunc("72h", "lease_time")
		assert.Empty(t, errs, "72h should be valid")

		_, errs = resource.Schema["lease_time"].ValidateFunc("30m", "lease_time")
		assert.Empty(t, errs, "30m should be valid")

		_, errs = resource.Schema["lease_time"].ValidateFunc("infinite", "lease_time")
		assert.Empty(t, errs, "infinite should be valid")

		_, errs = resource.Schema["lease_time"].ValidateFunc("", "lease_time")
		assert.Empty(t, errs, "empty should be valid")

		_, errs = resource.Schema["lease_time"].ValidateFunc("invalid", "lease_time")
		assert.NotEmpty(t, errs, "invalid format should be rejected")
	})
}

func TestResourceRTXDHCPScopeOptionsSchema(t *testing.T) {
	resource := resourceRTXDHCPScope()
	optionsSchema := resource.Schema["options"].Elem.(*schema.Resource).Schema

	t.Run("routers is optional with MaxItems 3", func(t *testing.T) {
		assert.True(t, optionsSchema["routers"].Optional)
		assert.Equal(t, 3, optionsSchema["routers"].MaxItems)
	})

	t.Run("dns_servers is optional with MaxItems 3", func(t *testing.T) {
		assert.True(t, optionsSchema["dns_servers"].Optional)
		assert.Equal(t, 3, optionsSchema["dns_servers"].MaxItems)
	})

	t.Run("domain_name is optional", func(t *testing.T) {
		assert.True(t, optionsSchema["domain_name"].Optional)
	})
}

func TestResourceRTXDHCPScopeExcludeRangesSchema(t *testing.T) {
	resource := resourceRTXDHCPScope()
	excludeRangesSchema := resource.Schema["exclude_ranges"].Elem.(*schema.Resource).Schema

	t.Run("start is required", func(t *testing.T) {
		assert.True(t, excludeRangesSchema["start"].Required)
	})

	t.Run("end is required", func(t *testing.T) {
		assert.True(t, excludeRangesSchema["end"].Required)
	})
}

func TestResourceRTXDHCPScopeImporter(t *testing.T) {
	resource := resourceRTXDHCPScope()

	t.Run("importer is configured", func(t *testing.T) {
		assert.NotNil(t, resource.Importer)
		assert.NotNil(t, resource.Importer.StateContext)
	})
}

func TestResourceRTXDHCPScopeCRUDFunctions(t *testing.T) {
	resource := resourceRTXDHCPScope()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, resource.CreateContext)
		assert.NotNil(t, resource.ReadContext)
		assert.NotNil(t, resource.UpdateContext)
		assert.NotNil(t, resource.DeleteContext)
	})
}

func TestValidateCIDR(t *testing.T) {
	tests := []struct {
		name    string
		cidr    string
		isValid bool
	}{
		{"valid /24 network", "192.168.1.0/24", true},
		{"valid /8 network", "10.0.0.0/8", true},
		{"valid /32 host", "192.168.1.1/32", true},
		{"invalid no prefix", "192.168.1.0", false},
		{"invalid text", "invalid", false},
		{"invalid prefix too large", "192.168.1.0/33", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateCIDR(tt.cidr, "cidr")
			if tt.isValid {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
			}
		})
	}
}

func TestValidateIPAddressDHCPScope(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		isValid bool
	}{
		{"valid IP", "192.168.1.1", true},
		{"valid zero", "0.0.0.0", true},
		{"valid broadcast", "255.255.255.255", true},
		{"empty (optional)", "", true},
		{"invalid text", "invalid", false},
		{"invalid out of range", "256.1.1.1", false},
		{"IPv6 not allowed", "2001:db8::1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateIPAddress(tt.ip, "ip")
			if tt.isValid {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
			}
		})
	}
}

func TestValidateLeaseTime(t *testing.T) {
	tests := []struct {
		name    string
		lease   string
		isValid bool
	}{
		{"72 hours", "72h", true},
		{"30 minutes", "30m", true},
		{"1 hour 30 minutes", "1h30m", true},
		{"infinite", "infinite", true},
		{"empty", "", true},
		{"no unit", "72", false},
		{"invalid unit", "72d", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateLeaseTime(tt.lease, "lease_time")
			if tt.isValid {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
			}
		})
	}
}
