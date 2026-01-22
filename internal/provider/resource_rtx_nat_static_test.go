package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestBuildNATStaticFromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.NATStatic
	}{
		{
			name: "basic static NAT without ports",
			input: map[string]interface{}{
				"descriptor_id": 1,
				"entry": []interface{}{
					map[string]interface{}{
						"inside_local":        "192.168.1.10",
						"inside_local_port":   0,
						"outside_global":      "203.0.113.10",
						"outside_global_port": 0,
						"protocol":            "",
					},
				},
			},
			expected: client.NATStatic{
				DescriptorID: 1,
				Entries: []client.NATStaticEntry{
					{
						InsideLocal:   "192.168.1.10",
						OutsideGlobal: "203.0.113.10",
					},
				},
			},
		},
		{
			name: "static NAT with TCP port mapping",
			input: map[string]interface{}{
				"descriptor_id": 2,
				"entry": []interface{}{
					map[string]interface{}{
						"inside_local":        "192.168.1.20",
						"inside_local_port":   80,
						"outside_global":      "203.0.113.20",
						"outside_global_port": 8080,
						"protocol":            "tcp",
					},
				},
			},
			expected: client.NATStatic{
				DescriptorID: 2,
				Entries: []client.NATStaticEntry{
					{
						InsideLocal:       "192.168.1.20",
						InsideLocalPort:   intPtrStatic(80),
						OutsideGlobal:     "203.0.113.20",
						OutsideGlobalPort: intPtrStatic(8080),
						Protocol:          "tcp",
					},
				},
			},
		},
		{
			name: "static NAT with UDP port mapping",
			input: map[string]interface{}{
				"descriptor_id": 3,
				"entry": []interface{}{
					map[string]interface{}{
						"inside_local":        "192.168.1.30",
						"inside_local_port":   53,
						"outside_global":      "203.0.113.30",
						"outside_global_port": 5353,
						"protocol":            "udp",
					},
				},
			},
			expected: client.NATStatic{
				DescriptorID: 3,
				Entries: []client.NATStaticEntry{
					{
						InsideLocal:       "192.168.1.30",
						InsideLocalPort:   intPtrStatic(53),
						OutsideGlobal:     "203.0.113.30",
						OutsideGlobalPort: intPtrStatic(5353),
						Protocol:          "udp",
					},
				},
			},
		},
		{
			name: "multiple static NAT entries",
			input: map[string]interface{}{
				"descriptor_id": 4,
				"entry": []interface{}{
					map[string]interface{}{
						"inside_local":        "192.168.1.100",
						"inside_local_port":   0,
						"outside_global":      "203.0.113.100",
						"outside_global_port": 0,
						"protocol":            "",
					},
					map[string]interface{}{
						"inside_local":        "192.168.1.101",
						"inside_local_port":   443,
						"outside_global":      "203.0.113.101",
						"outside_global_port": 443,
						"protocol":            "tcp",
					},
					map[string]interface{}{
						"inside_local":        "192.168.1.102",
						"inside_local_port":   22,
						"outside_global":      "203.0.113.102",
						"outside_global_port": 2222,
						"protocol":            "tcp",
					},
				},
			},
			expected: client.NATStatic{
				DescriptorID: 4,
				Entries: []client.NATStaticEntry{
					{
						InsideLocal:   "192.168.1.100",
						OutsideGlobal: "203.0.113.100",
					},
					{
						InsideLocal:       "192.168.1.101",
						InsideLocalPort:   intPtrStatic(443),
						OutsideGlobal:     "203.0.113.101",
						OutsideGlobalPort: intPtrStatic(443),
						Protocol:          "tcp",
					},
					{
						InsideLocal:       "192.168.1.102",
						InsideLocalPort:   intPtrStatic(22),
						OutsideGlobal:     "203.0.113.102",
						OutsideGlobalPort: intPtrStatic(2222),
						Protocol:          "tcp",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceSchema := resourceRTXNATStatic().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			result := buildNATStaticFromResourceData(d)

			assert.Equal(t, tt.expected.DescriptorID, result.DescriptorID)
			assert.Equal(t, len(tt.expected.Entries), len(result.Entries))

			for i, expectedEntry := range tt.expected.Entries {
				assert.Equal(t, expectedEntry.InsideLocal, result.Entries[i].InsideLocal, "entry[%d].InsideLocal", i)
				assert.Equal(t, expectedEntry.OutsideGlobal, result.Entries[i].OutsideGlobal, "entry[%d].OutsideGlobal", i)
				assert.Equal(t, expectedEntry.Protocol, result.Entries[i].Protocol, "entry[%d].Protocol", i)

				if expectedEntry.InsideLocalPort != nil {
					assert.NotNil(t, result.Entries[i].InsideLocalPort, "entry[%d].InsideLocalPort should not be nil", i)
					assert.Equal(t, *expectedEntry.InsideLocalPort, *result.Entries[i].InsideLocalPort, "entry[%d].InsideLocalPort", i)
				}
				if expectedEntry.OutsideGlobalPort != nil {
					assert.NotNil(t, result.Entries[i].OutsideGlobalPort, "entry[%d].OutsideGlobalPort should not be nil", i)
					assert.Equal(t, *expectedEntry.OutsideGlobalPort, *result.Entries[i].OutsideGlobalPort, "entry[%d].OutsideGlobalPort", i)
				}
			}
		})
	}
}

// Helper function specific to NAT static tests
func intPtrStatic(i int) *int {
	return &i
}

func TestExpandNATStaticEntries(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []client.NATStaticEntry
	}{
		{
			name: "single entry without ports",
			input: []interface{}{
				map[string]interface{}{
					"inside_local":        "10.0.0.1",
					"inside_local_port":   0,
					"outside_global":      "1.2.3.4",
					"outside_global_port": 0,
					"protocol":            "",
				},
			},
			expected: []client.NATStaticEntry{
				{
					InsideLocal:   "10.0.0.1",
					OutsideGlobal: "1.2.3.4",
				},
			},
		},
		{
			name: "single entry with ports",
			input: []interface{}{
				map[string]interface{}{
					"inside_local":        "10.0.0.2",
					"inside_local_port":   8080,
					"outside_global":      "1.2.3.5",
					"outside_global_port": 80,
					"protocol":            "tcp",
				},
			},
			expected: []client.NATStaticEntry{
				{
					InsideLocal:       "10.0.0.2",
					InsideLocalPort:   intPtrStatic(8080),
					OutsideGlobal:     "1.2.3.5",
					OutsideGlobalPort: intPtrStatic(80),
					Protocol:          "tcp",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandNATStaticEntries(tt.input)

			assert.Equal(t, len(tt.expected), len(result))
			for i, expected := range tt.expected {
				assert.Equal(t, expected.InsideLocal, result[i].InsideLocal)
				assert.Equal(t, expected.OutsideGlobal, result[i].OutsideGlobal)
				assert.Equal(t, expected.Protocol, result[i].Protocol)
			}
		})
	}
}

func TestFlattenNATStaticEntries(t *testing.T) {
	tests := []struct {
		name     string
		input    []client.NATStaticEntry
		expected []interface{}
	}{
		{
			name: "entry without ports",
			input: []client.NATStaticEntry{
				{
					InsideLocal:   "192.168.1.1",
					OutsideGlobal: "1.1.1.1",
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"inside_local":        "192.168.1.1",
					"inside_local_port":   0,
					"outside_global":      "1.1.1.1",
					"outside_global_port": 0,
					"protocol":            "",
				},
			},
		},
		{
			name: "entry with ports",
			input: []client.NATStaticEntry{
				{
					InsideLocal:       "192.168.1.2",
					InsideLocalPort:   intPtrStatic(80),
					OutsideGlobal:     "2.2.2.2",
					OutsideGlobalPort: intPtrStatic(8080),
					Protocol:          "tcp",
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"inside_local":        "192.168.1.2",
					"inside_local_port":   80,
					"outside_global":      "2.2.2.2",
					"outside_global_port": 8080,
					"protocol":            "tcp",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenNATStaticEntries(tt.input)

			assert.Equal(t, len(tt.expected), len(result))
			for i, expected := range tt.expected {
				expectedMap := expected.(map[string]interface{})
				resultMap := result[i].(map[string]interface{})

				assert.Equal(t, expectedMap["inside_local"], resultMap["inside_local"])
				assert.Equal(t, expectedMap["outside_global"], resultMap["outside_global"])
				assert.Equal(t, expectedMap["protocol"], resultMap["protocol"])
				assert.Equal(t, expectedMap["inside_local_port"], resultMap["inside_local_port"])
				assert.Equal(t, expectedMap["outside_global_port"], resultMap["outside_global_port"])
			}
		})
	}
}

func TestResourceRTXNATStaticSchema(t *testing.T) {
	resource := resourceRTXNATStatic()

	t.Run("descriptor_id is required and ForceNew", func(t *testing.T) {
		assert.True(t, resource.Schema["descriptor_id"].Required)
		assert.True(t, resource.Schema["descriptor_id"].ForceNew)
	})

	t.Run("entry is required", func(t *testing.T) {
		assert.True(t, resource.Schema["entry"].Required)
	})
}

func TestResourceRTXNATStaticSchemaValidation(t *testing.T) {
	resource := resourceRTXNATStatic()

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
}

func TestResourceRTXNATStaticEntrySchema(t *testing.T) {
	resource := resourceRTXNATStatic()
	entrySchema := resource.Schema["entry"].Elem.(*schema.Resource).Schema

	t.Run("inside_local is required", func(t *testing.T) {
		assert.True(t, entrySchema["inside_local"].Required)
	})

	t.Run("outside_global is required", func(t *testing.T) {
		assert.True(t, entrySchema["outside_global"].Required)
	})

	t.Run("ports are optional", func(t *testing.T) {
		assert.True(t, entrySchema["inside_local_port"].Optional)
		assert.True(t, entrySchema["outside_global_port"].Optional)
	})

	t.Run("protocol is optional", func(t *testing.T) {
		assert.True(t, entrySchema["protocol"].Optional)
	})

	t.Run("protocol validation", func(t *testing.T) {
		_, errs := entrySchema["protocol"].ValidateFunc("tcp", "protocol")
		assert.Empty(t, errs, "tcp should be valid")

		_, errs = entrySchema["protocol"].ValidateFunc("udp", "protocol")
		assert.Empty(t, errs, "udp should be valid")

		_, errs = entrySchema["protocol"].ValidateFunc("invalid", "protocol")
		assert.NotEmpty(t, errs, "invalid should be rejected")
	})

	t.Run("port validation", func(t *testing.T) {
		_, errs := entrySchema["inside_local_port"].ValidateFunc(1, "inside_local_port")
		assert.Empty(t, errs, "port 1 should be valid")

		_, errs = entrySchema["inside_local_port"].ValidateFunc(65535, "inside_local_port")
		assert.Empty(t, errs, "port 65535 should be valid")

		_, errs = entrySchema["inside_local_port"].ValidateFunc(0, "inside_local_port")
		assert.NotEmpty(t, errs, "port 0 should be invalid")

		_, errs = entrySchema["inside_local_port"].ValidateFunc(65536, "inside_local_port")
		assert.NotEmpty(t, errs, "port 65536 should be invalid")
	})
}

func TestValidateNATIPAddressStatic(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		isValid bool
	}{
		{"valid IP", "192.168.1.1", true},
		{"valid public IP", "203.0.113.10", true},
		{"invalid empty", "", false},
		{"invalid text", "invalid", false},
		{"IPv6 not allowed", "2001:db8::1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateNATIPAddress(tt.ip, "ip")
			if tt.isValid {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
			}
		})
	}
}

func TestResourceRTXNATStaticImporter(t *testing.T) {
	resource := resourceRTXNATStatic()

	t.Run("importer is configured", func(t *testing.T) {
		assert.NotNil(t, resource.Importer)
		assert.NotNil(t, resource.Importer.StateContext)
	})
}

func TestResourceRTXNATStaticCRUDFunctions(t *testing.T) {
	resource := resourceRTXNATStatic()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, resource.CreateContext)
		assert.NotNil(t, resource.ReadContext)
		assert.NotNil(t, resource.UpdateContext)
		assert.NotNil(t, resource.DeleteContext)
	})

	t.Run("CustomizeDiff is configured", func(t *testing.T) {
		assert.NotNil(t, resource.CustomizeDiff)
	})
}
