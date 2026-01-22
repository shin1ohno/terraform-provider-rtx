package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestBuildSystemConfigFromResourceData(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected client.SystemConfig
	}{
		{
			name: "timezone only",
			input: map[string]interface{}{
				"timezone":      "+09:00",
				"console":       []interface{}{},
				"packet_buffer": []interface{}{},
				"statistics":    []interface{}{},
			},
			expected: client.SystemConfig{
				Timezone:      "+09:00",
				PacketBuffers: []client.PacketBufferConfig{},
			},
		},
		{
			name: "with console settings",
			input: map[string]interface{}{
				"timezone": "+09:00",
				"console": []interface{}{
					map[string]interface{}{
						"character": "ja.utf8",
						"lines":     "infinity",
						"prompt":    "RTX1200>",
					},
				},
				"packet_buffer": []interface{}{},
				"statistics":    []interface{}{},
			},
			expected: client.SystemConfig{
				Timezone: "+09:00",
				Console: &client.ConsoleConfig{
					Character: "ja.utf8",
					Lines:     "infinity",
					Prompt:    "RTX1200>",
				},
				PacketBuffers: []client.PacketBufferConfig{},
			},
		},
		{
			name: "with packet buffer settings",
			input: map[string]interface{}{
				"timezone": "-05:00",
				"console":  []interface{}{},
				"packet_buffer": []interface{}{
					map[string]interface{}{
						"size":       "small",
						"max_buffer": 1000,
						"max_free":   500,
					},
					map[string]interface{}{
						"size":       "middle",
						"max_buffer": 500,
						"max_free":   250,
					},
					map[string]interface{}{
						"size":       "large",
						"max_buffer": 100,
						"max_free":   50,
					},
				},
				"statistics": []interface{}{},
			},
			expected: client.SystemConfig{
				Timezone: "-05:00",
				PacketBuffers: []client.PacketBufferConfig{
					{Size: "small", MaxBuffer: 1000, MaxFree: 500},
					{Size: "middle", MaxBuffer: 500, MaxFree: 250},
					{Size: "large", MaxBuffer: 100, MaxFree: 50},
				},
			},
		},
		{
			name: "with statistics settings",
			input: map[string]interface{}{
				"timezone":      "+00:00",
				"console":       []interface{}{},
				"packet_buffer": []interface{}{},
				"statistics": []interface{}{
					map[string]interface{}{
						"traffic": true,
						"nat":     true,
					},
				},
			},
			expected: client.SystemConfig{
				Timezone:      "+00:00",
				PacketBuffers: []client.PacketBufferConfig{},
				Statistics: &client.StatisticsConfig{
					Traffic: true,
					NAT:     true,
				},
			},
		},
		{
			name: "full configuration",
			input: map[string]interface{}{
				"timezone": "+09:00",
				"console": []interface{}{
					map[string]interface{}{
						"character": "ascii",
						"lines":     "24",
						"prompt":    "router>",
					},
				},
				"packet_buffer": []interface{}{
					map[string]interface{}{
						"size":       "small",
						"max_buffer": 2000,
						"max_free":   1000,
					},
				},
				"statistics": []interface{}{
					map[string]interface{}{
						"traffic": true,
						"nat":     false,
					},
				},
			},
			expected: client.SystemConfig{
				Timezone: "+09:00",
				Console: &client.ConsoleConfig{
					Character: "ascii",
					Lines:     "24",
					Prompt:    "router>",
				},
				PacketBuffers: []client.PacketBufferConfig{
					{Size: "small", MaxBuffer: 2000, MaxFree: 1000},
				},
				Statistics: &client.StatisticsConfig{
					Traffic: true,
					NAT:     false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceSchema := resourceRTXSystem().Schema
			d := schema.TestResourceDataRaw(t, resourceSchema, tt.input)

			result := buildSystemConfigFromResourceData(d)

			assert.Equal(t, tt.expected.Timezone, result.Timezone)

			// Check console
			if tt.expected.Console != nil {
				assert.NotNil(t, result.Console)
				assert.Equal(t, tt.expected.Console.Character, result.Console.Character)
				assert.Equal(t, tt.expected.Console.Lines, result.Console.Lines)
				assert.Equal(t, tt.expected.Console.Prompt, result.Console.Prompt)
			} else {
				assert.Nil(t, result.Console)
			}

			// Check packet buffers
			assert.Equal(t, len(tt.expected.PacketBuffers), len(result.PacketBuffers))
			for i, expectedPB := range tt.expected.PacketBuffers {
				assert.Equal(t, expectedPB.Size, result.PacketBuffers[i].Size, "packet_buffer[%d].Size", i)
				assert.Equal(t, expectedPB.MaxBuffer, result.PacketBuffers[i].MaxBuffer, "packet_buffer[%d].MaxBuffer", i)
				assert.Equal(t, expectedPB.MaxFree, result.PacketBuffers[i].MaxFree, "packet_buffer[%d].MaxFree", i)
			}

			// Check statistics
			if tt.expected.Statistics != nil {
				assert.NotNil(t, result.Statistics)
				assert.Equal(t, tt.expected.Statistics.Traffic, result.Statistics.Traffic)
				assert.Equal(t, tt.expected.Statistics.NAT, result.Statistics.NAT)
			} else {
				assert.Nil(t, result.Statistics)
			}
		})
	}
}

func TestResourceRTXSystemSchema(t *testing.T) {
	resource := resourceRTXSystem()

	t.Run("timezone is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["timezone"].Optional)
	})

	t.Run("console is optional with MaxItems 1", func(t *testing.T) {
		assert.True(t, resource.Schema["console"].Optional)
		assert.Equal(t, 1, resource.Schema["console"].MaxItems)
	})

	t.Run("packet_buffer is optional with MaxItems 3", func(t *testing.T) {
		assert.True(t, resource.Schema["packet_buffer"].Optional)
		assert.Equal(t, 3, resource.Schema["packet_buffer"].MaxItems)
	})

	t.Run("statistics is optional with MaxItems 1", func(t *testing.T) {
		assert.True(t, resource.Schema["statistics"].Optional)
		assert.Equal(t, 1, resource.Schema["statistics"].MaxItems)
	})
}

func TestResourceRTXSystemConsoleSchema(t *testing.T) {
	resource := resourceRTXSystem()
	consoleSchema := resource.Schema["console"].Elem.(*schema.Resource).Schema

	t.Run("character is optional", func(t *testing.T) {
		assert.True(t, consoleSchema["character"].Optional)
	})

	t.Run("character validation", func(t *testing.T) {
		validCharsets := []string{"ja.utf8", "ja.sjis", "ascii", "euc-jp"}
		for _, charset := range validCharsets {
			_, errs := consoleSchema["character"].ValidateFunc(charset, "character")
			assert.Empty(t, errs, "%s should be valid", charset)
		}

		_, errs := consoleSchema["character"].ValidateFunc("invalid", "character")
		assert.NotEmpty(t, errs, "invalid should be rejected")
	})

	t.Run("lines is optional", func(t *testing.T) {
		assert.True(t, consoleSchema["lines"].Optional)
	})

	t.Run("prompt is optional", func(t *testing.T) {
		assert.True(t, consoleSchema["prompt"].Optional)
	})
}

func TestResourceRTXSystemPacketBufferSchema(t *testing.T) {
	resource := resourceRTXSystem()
	pbSchema := resource.Schema["packet_buffer"].Elem.(*schema.Resource).Schema

	t.Run("size is required", func(t *testing.T) {
		assert.True(t, pbSchema["size"].Required)
	})

	t.Run("size validation", func(t *testing.T) {
		validSizes := []string{"small", "middle", "large"}
		for _, size := range validSizes {
			_, errs := pbSchema["size"].ValidateFunc(size, "size")
			assert.Empty(t, errs, "%s should be valid", size)
		}

		_, errs := pbSchema["size"].ValidateFunc("invalid", "size")
		assert.NotEmpty(t, errs, "invalid should be rejected")
	})

	t.Run("max_buffer is required", func(t *testing.T) {
		assert.True(t, pbSchema["max_buffer"].Required)
	})

	t.Run("max_buffer validation", func(t *testing.T) {
		_, errs := pbSchema["max_buffer"].ValidateFunc(1, "max_buffer")
		assert.Empty(t, errs, "1 should be valid")

		_, errs = pbSchema["max_buffer"].ValidateFunc(0, "max_buffer")
		assert.NotEmpty(t, errs, "0 should be invalid")
	})

	t.Run("max_free is required", func(t *testing.T) {
		assert.True(t, pbSchema["max_free"].Required)
	})

	t.Run("max_free validation", func(t *testing.T) {
		_, errs := pbSchema["max_free"].ValidateFunc(1, "max_free")
		assert.Empty(t, errs, "1 should be valid")

		_, errs = pbSchema["max_free"].ValidateFunc(0, "max_free")
		assert.NotEmpty(t, errs, "0 should be invalid")
	})
}

func TestResourceRTXSystemStatisticsSchema(t *testing.T) {
	resource := resourceRTXSystem()
	statsSchema := resource.Schema["statistics"].Elem.(*schema.Resource).Schema

	t.Run("traffic is optional and computed", func(t *testing.T) {
		assert.True(t, statsSchema["traffic"].Optional)
		assert.True(t, statsSchema["traffic"].Computed)
	})

	t.Run("nat is optional and computed", func(t *testing.T) {
		assert.True(t, statsSchema["nat"].Optional)
		assert.True(t, statsSchema["nat"].Computed)
	})
}

func TestValidateTimezone(t *testing.T) {
	tests := []struct {
		name    string
		tz      string
		isValid bool
	}{
		{"JST", "+09:00", true},
		{"EST", "-05:00", true},
		{"UTC", "+00:00", true},
		{"empty", "", true},
		{"IST +05:30", "+05:30", true},
		{"invalid format", "JST", false},
		{"missing colon", "+0900", false},
		{"missing sign", "09:00", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateTimezone(tt.tz, "timezone")
			if tt.isValid {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
			}
		})
	}
}

func TestValidateConsoleLines(t *testing.T) {
	tests := []struct {
		name    string
		lines   string
		isValid bool
	}{
		{"infinity", "infinity", true},
		{"positive integer", "24", true},
		{"large number", "1000", true},
		{"empty", "", true},
		{"zero", "0", false},
		{"negative", "-1", false},
		{"text", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errs := validateConsoleLines(tt.lines, "lines")
			if tt.isValid {
				assert.Empty(t, errs)
			} else {
				assert.NotEmpty(t, errs)
			}
		})
	}
}

func TestResourceRTXSystemImporter(t *testing.T) {
	resource := resourceRTXSystem()

	t.Run("importer is configured", func(t *testing.T) {
		assert.NotNil(t, resource.Importer)
		assert.NotNil(t, resource.Importer.StateContext)
	})
}

func TestResourceRTXSystemCRUDFunctions(t *testing.T) {
	resource := resourceRTXSystem()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, resource.CreateContext)
		assert.NotNil(t, resource.ReadContext)
		assert.NotNil(t, resource.UpdateContext)
		assert.NotNil(t, resource.DeleteContext)
	})
}
