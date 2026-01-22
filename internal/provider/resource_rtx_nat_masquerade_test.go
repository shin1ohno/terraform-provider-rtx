package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestBuildNATMasqueradeFromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.NATMasquerade
	}{
		{
			name: "basic NAT masquerade with ipcp",
			input: map[string]interface{}{
				"descriptor_id": 1,
				"outer_address": "ipcp",
				"inner_network": "192.168.1.0-192.168.1.255",
				"static_entry":  []interface{}{},
			},
			expected: client.NATMasquerade{
				DescriptorID: 1,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.1.0-192.168.1.255",
			},
		},
		{
			name: "NAT masquerade with interface",
			input: map[string]interface{}{
				"descriptor_id": 2,
				"outer_address": "pp1",
				"inner_network": "10.0.0.0-10.0.0.255",
				"static_entry":  []interface{}{},
			},
			expected: client.NATMasquerade{
				DescriptorID: 2,
				OuterAddress: "pp1",
				InnerNetwork: "10.0.0.0-10.0.0.255",
			},
		},
		{
			name: "NAT masquerade with TCP static entry",
			input: map[string]interface{}{
				"descriptor_id": 3,
				"outer_address": "ipcp",
				"inner_network": "192.168.0.0-192.168.0.255",
				"static_entry": []interface{}{
					map[string]interface{}{
						"entry_number":        1,
						"inside_local":        "192.168.0.100",
						"inside_local_port":   80,
						"outside_global":      "ipcp",
						"outside_global_port": 8080,
						"protocol":            "tcp",
					},
				},
			},
			expected: client.NATMasquerade{
				DescriptorID: 3,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.0.0-192.168.0.255",
				StaticEntries: []client.MasqueradeStaticEntry{
					{
						EntryNumber:       1,
						InsideLocal:       "192.168.0.100",
						InsideLocalPort:   intPtrNAT(80),
						OutsideGlobal:     "ipcp",
						OutsideGlobalPort: intPtrNAT(8080),
						Protocol:          "tcp",
					},
				},
			},
		},
		{
			name: "NAT masquerade with UDP static entry",
			input: map[string]interface{}{
				"descriptor_id": 4,
				"outer_address": "ipcp",
				"inner_network": "172.16.0.0-172.16.0.255",
				"static_entry": []interface{}{
					map[string]interface{}{
						"entry_number":        1,
						"inside_local":        "172.16.0.10",
						"inside_local_port":   53,
						"outside_global":      "ipcp",
						"outside_global_port": 53,
						"protocol":            "udp",
					},
				},
			},
			expected: client.NATMasquerade{
				DescriptorID: 4,
				OuterAddress: "ipcp",
				InnerNetwork: "172.16.0.0-172.16.0.255",
				StaticEntries: []client.MasqueradeStaticEntry{
					{
						EntryNumber:       1,
						InsideLocal:       "172.16.0.10",
						InsideLocalPort:   intPtrNAT(53),
						OutsideGlobal:     "ipcp",
						OutsideGlobalPort: intPtrNAT(53),
						Protocol:          "udp",
					},
				},
			},
		},
		{
			name: "NAT masquerade with protocol-only entry (ESP)",
			input: map[string]interface{}{
				"descriptor_id": 5,
				"outer_address": "ipcp",
				"inner_network": "192.168.10.0-192.168.10.255",
				"static_entry": []interface{}{
					map[string]interface{}{
						"entry_number":        1,
						"inside_local":        "192.168.10.1",
						"inside_local_port":   0,
						"outside_global":      "ipcp",
						"outside_global_port": 0,
						"protocol":            "esp",
					},
				},
			},
			expected: client.NATMasquerade{
				DescriptorID: 5,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.10.0-192.168.10.255",
				StaticEntries: []client.MasqueradeStaticEntry{
					{
						EntryNumber:   1,
						InsideLocal:   "192.168.10.1",
						OutsideGlobal: "ipcp",
						Protocol:      "esp",
					},
				},
			},
		},
		{
			name: "NAT masquerade with multiple static entries",
			input: map[string]interface{}{
				"descriptor_id": 6,
				"outer_address": "ipcp",
				"inner_network": "192.168.100.0-192.168.100.255",
				"static_entry": []interface{}{
					map[string]interface{}{
						"entry_number":        1,
						"inside_local":        "192.168.100.10",
						"inside_local_port":   80,
						"outside_global":      "ipcp",
						"outside_global_port": 80,
						"protocol":            "tcp",
					},
					map[string]interface{}{
						"entry_number":        2,
						"inside_local":        "192.168.100.10",
						"inside_local_port":   443,
						"outside_global":      "ipcp",
						"outside_global_port": 443,
						"protocol":            "tcp",
					},
					map[string]interface{}{
						"entry_number":        3,
						"inside_local":        "192.168.100.20",
						"inside_local_port":   0,
						"outside_global":      "ipcp",
						"outside_global_port": 0,
						"protocol":            "gre",
					},
				},
			},
			expected: client.NATMasquerade{
				DescriptorID: 6,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.100.0-192.168.100.255",
				StaticEntries: []client.MasqueradeStaticEntry{
					{
						EntryNumber:       1,
						InsideLocal:       "192.168.100.10",
						InsideLocalPort:   intPtrNAT(80),
						OutsideGlobal:     "ipcp",
						OutsideGlobalPort: intPtrNAT(80),
						Protocol:          "tcp",
					},
					{
						EntryNumber:       2,
						InsideLocal:       "192.168.100.10",
						InsideLocalPort:   intPtrNAT(443),
						OutsideGlobal:     "ipcp",
						OutsideGlobalPort: intPtrNAT(443),
						Protocol:          "tcp",
					},
					{
						EntryNumber:   3,
						InsideLocal:   "192.168.100.20",
						OutsideGlobal: "ipcp",
						Protocol:      "gre",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceSchema := resourceRTXNATMasquerade().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			result := buildNATMasqueradeFromResourceData(d)

			assert.Equal(t, tt.expected.DescriptorID, result.DescriptorID)
			assert.Equal(t, tt.expected.OuterAddress, result.OuterAddress)
			assert.Equal(t, tt.expected.InnerNetwork, result.InnerNetwork)

			assert.Equal(t, len(tt.expected.StaticEntries), len(result.StaticEntries))
			for i, expectedEntry := range tt.expected.StaticEntries {
				actualEntry := result.StaticEntries[i]
				assert.Equal(t, expectedEntry.EntryNumber, actualEntry.EntryNumber, "entry[%d].EntryNumber", i)
				assert.Equal(t, expectedEntry.InsideLocal, actualEntry.InsideLocal, "entry[%d].InsideLocal", i)
				assert.Equal(t, expectedEntry.OutsideGlobal, actualEntry.OutsideGlobal, "entry[%d].OutsideGlobal", i)
				assert.Equal(t, expectedEntry.Protocol, actualEntry.Protocol, "entry[%d].Protocol", i)

				// Check port pointers
				if expectedEntry.InsideLocalPort != nil {
					assert.NotNil(t, actualEntry.InsideLocalPort, "entry[%d].InsideLocalPort should not be nil", i)
					assert.Equal(t, *expectedEntry.InsideLocalPort, *actualEntry.InsideLocalPort, "entry[%d].InsideLocalPort", i)
				}
				if expectedEntry.OutsideGlobalPort != nil {
					assert.NotNil(t, actualEntry.OutsideGlobalPort, "entry[%d].OutsideGlobalPort should not be nil", i)
					assert.Equal(t, *expectedEntry.OutsideGlobalPort, *actualEntry.OutsideGlobalPort, "entry[%d].OutsideGlobalPort", i)
				}
			}
		})
	}
}

func intPtrNAT(i int) *int {
	return &i
}

func TestFlattenStaticEntries(t *testing.T) {
	tests := []struct {
		name     string
		entries  []client.MasqueradeStaticEntry
		expected []map[string]interface{}
	}{
		{
			name: "single TCP entry",
			entries: []client.MasqueradeStaticEntry{
				{
					EntryNumber:       1,
					InsideLocal:       "192.168.1.100",
					InsideLocalPort:   intPtrNAT(80),
					OutsideGlobal:     "ipcp",
					OutsideGlobalPort: intPtrNAT(8080),
					Protocol:          "tcp",
				},
			},
			expected: []map[string]interface{}{
				{
					"entry_number":        1,
					"inside_local":        "192.168.1.100",
					"inside_local_port":   80,
					"outside_global":      "ipcp",
					"outside_global_port": 8080,
					"protocol":            "tcp",
				},
			},
		},
		{
			name: "protocol-only entry (ESP)",
			entries: []client.MasqueradeStaticEntry{
				{
					EntryNumber:   1,
					InsideLocal:   "192.168.1.1",
					OutsideGlobal: "ipcp",
					Protocol:      "esp",
				},
			},
			expected: []map[string]interface{}{
				{
					"entry_number":   1,
					"inside_local":   "192.168.1.1",
					"outside_global": "ipcp",
					"protocol":       "esp",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenStaticEntries(tt.entries)

			assert.Equal(t, len(tt.expected), len(result))

			for i, expectedEntry := range tt.expected {
				actualEntry := result[i].(map[string]interface{})
				assert.Equal(t, expectedEntry["entry_number"], actualEntry["entry_number"])
				assert.Equal(t, expectedEntry["inside_local"], actualEntry["inside_local"])
				assert.Equal(t, expectedEntry["outside_global"], actualEntry["outside_global"])
				assert.Equal(t, expectedEntry["protocol"], actualEntry["protocol"])
			}
		})
	}
}

func TestResourceRTXNATMasqueradeSchema(t *testing.T) {
	resource := resourceRTXNATMasquerade()

	t.Run("descriptor_id is required and ForceNew", func(t *testing.T) {
		assert.True(t, resource.Schema["descriptor_id"].Required)
		assert.True(t, resource.Schema["descriptor_id"].ForceNew)
	})

	t.Run("outer_address is required", func(t *testing.T) {
		assert.True(t, resource.Schema["outer_address"].Required)
	})

	t.Run("inner_network is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["inner_network"].Optional)
	})

	t.Run("static_entry is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["static_entry"].Optional)
	})
}

func TestResourceRTXNATMasqueradeSchemaValidation(t *testing.T) {
	resource := resourceRTXNATMasquerade()

	t.Run("descriptor_id validation", func(t *testing.T) {
		_, errs := resource.Schema["descriptor_id"].ValidateFunc(1, "descriptor_id")
		assert.Empty(t, errs, "descriptor_id 1 should be valid")

		_, errs = resource.Schema["descriptor_id"].ValidateFunc(65535, "descriptor_id")
		assert.Empty(t, errs, "descriptor_id 65535 should be valid")

		_, errs = resource.Schema["descriptor_id"].ValidateFunc(0, "descriptor_id")
		assert.NotEmpty(t, errs, "descriptor_id 0 should be invalid")

		_, errs = resource.Schema["descriptor_id"].ValidateFunc(65536, "descriptor_id")
		assert.NotEmpty(t, errs, "descriptor_id 65536 should be invalid")
	})

	t.Run("outer_address validation", func(t *testing.T) {
		_, errs := resource.Schema["outer_address"].ValidateFunc("ipcp", "outer_address")
		assert.Empty(t, errs, "ipcp should be valid")

		_, errs = resource.Schema["outer_address"].ValidateFunc("pp1", "outer_address")
		assert.Empty(t, errs, "pp1 should be valid")

		_, errs = resource.Schema["outer_address"].ValidateFunc("192.168.1.1", "outer_address")
		assert.Empty(t, errs, "IP address should be valid")

		_, errs = resource.Schema["outer_address"].ValidateFunc("", "outer_address")
		assert.NotEmpty(t, errs, "empty should be invalid")
	})

	t.Run("inner_network validation (IP range)", func(t *testing.T) {
		_, errs := resource.Schema["inner_network"].ValidateFunc("192.168.1.0-192.168.1.255", "inner_network")
		assert.Empty(t, errs, "valid IP range should be accepted")

		_, errs = resource.Schema["inner_network"].ValidateFunc("10.0.0.1-10.0.0.100", "inner_network")
		assert.Empty(t, errs, "valid IP range should be accepted")

		_, errs = resource.Schema["inner_network"].ValidateFunc("192.168.1.0", "inner_network")
		assert.NotEmpty(t, errs, "single IP without range should be rejected")

		_, errs = resource.Schema["inner_network"].ValidateFunc("", "inner_network")
		assert.NotEmpty(t, errs, "empty should be rejected")
	})
}

func TestResourceRTXNATMasqueradeStaticEntrySchema(t *testing.T) {
	resource := resourceRTXNATMasquerade()
	staticEntrySchema := resource.Schema["static_entry"].Elem.(*schema.Resource).Schema

	t.Run("entry_number is required", func(t *testing.T) {
		assert.True(t, staticEntrySchema["entry_number"].Required)
	})

	t.Run("inside_local is required", func(t *testing.T) {
		assert.True(t, staticEntrySchema["inside_local"].Required)
	})

	t.Run("inside_local_port is optional", func(t *testing.T) {
		assert.True(t, staticEntrySchema["inside_local_port"].Optional)
	})

	t.Run("outside_global is optional with default ipcp", func(t *testing.T) {
		assert.True(t, staticEntrySchema["outside_global"].Optional)
		assert.Equal(t, "ipcp", staticEntrySchema["outside_global"].Default)
	})

	t.Run("outside_global_port is optional", func(t *testing.T) {
		assert.True(t, staticEntrySchema["outside_global_port"].Optional)
	})

	t.Run("protocol validation", func(t *testing.T) {
		validProtocols := []string{"tcp", "udp", "esp", "ah", "gre", "icmp"}
		for _, proto := range validProtocols {
			_, errs := staticEntrySchema["protocol"].ValidateFunc(proto, "protocol")
			assert.Empty(t, errs, "protocol '%s' should be valid", proto)
		}

		_, errs := staticEntrySchema["protocol"].ValidateFunc("invalid", "protocol")
		assert.NotEmpty(t, errs, "protocol 'invalid' should be invalid")
	})

	t.Run("port validation", func(t *testing.T) {
		_, errs := staticEntrySchema["inside_local_port"].ValidateFunc(1, "inside_local_port")
		assert.Empty(t, errs, "port 1 should be valid")

		_, errs = staticEntrySchema["inside_local_port"].ValidateFunc(65535, "inside_local_port")
		assert.Empty(t, errs, "port 65535 should be valid")

		_, errs = staticEntrySchema["inside_local_port"].ValidateFunc(0, "inside_local_port")
		assert.NotEmpty(t, errs, "port 0 should be invalid")

		_, errs = staticEntrySchema["inside_local_port"].ValidateFunc(65536, "inside_local_port")
		assert.NotEmpty(t, errs, "port 65536 should be invalid")
	})
}

func TestResourceRTXNATMasqueradeImporter(t *testing.T) {
	resource := resourceRTXNATMasquerade()

	t.Run("importer is configured", func(t *testing.T) {
		assert.NotNil(t, resource.Importer)
		assert.NotNil(t, resource.Importer.StateContext)
	})
}

func TestResourceRTXNATMasqueradeCRUDFunctions(t *testing.T) {
	resource := resourceRTXNATMasquerade()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, resource.CreateContext)
		assert.NotNil(t, resource.ReadContext)
		assert.NotNil(t, resource.UpdateContext)
		assert.NotNil(t, resource.DeleteContext)
	})
}

func TestParseNATMasqueradeID(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		expected    int
		shouldError bool
	}{
		{"valid ID 1", "1", 1, false},
		{"valid ID 100", "100", 100, false},
		{"valid ID 65535", "65535", 65535, false},
		{"invalid - zero", "0", 0, true},
		{"invalid - negative", "-1", 0, true},
		{"invalid - too large", "65536", 0, true},
		{"invalid - not a number", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseNATMasqueradeID(tt.id)
			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValidateIPRange(t *testing.T) {
	tests := []struct {
		name    string
		ipRange string
		isValid bool
	}{
		{"valid range", "192.168.1.0-192.168.1.255", true},
		{"valid small range", "10.0.0.1-10.0.0.10", true},
		{"single IP (invalid)", "192.168.1.1", false},
		{"empty (invalid)", "", false},
		{"no separator (invalid)", "192.168.1.0 192.168.1.255", false},
		{"invalid start IP", "invalid-192.168.1.255", false},
		{"invalid end IP", "192.168.1.0-invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateIPRange(tt.ipRange, "inner_network")
			if tt.isValid {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
			}
		})
	}
}
