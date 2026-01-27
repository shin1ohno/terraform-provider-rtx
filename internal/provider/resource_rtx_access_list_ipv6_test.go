package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestResourceRTXAccessListIPv6Schema(t *testing.T) {
	resource := resourceRTXAccessListIPv6()

	t.Run("name is required and forces new", func(t *testing.T) {
		assert.True(t, resource.Schema["name"].Required)
		assert.True(t, resource.Schema["name"].ForceNew)
	})

	t.Run("sequence_start is optional", func(t *testing.T) {
		assert.True(t, resource.Schema["sequence_start"].Optional)
	})

	t.Run("sequence_step is optional with default", func(t *testing.T) {
		assert.True(t, resource.Schema["sequence_step"].Optional)
		assert.Equal(t, DefaultSequenceStep, resource.Schema["sequence_step"].Default)
	})

	t.Run("entry is required list", func(t *testing.T) {
		assert.True(t, resource.Schema["entry"].Required)
		assert.Equal(t, 1, resource.Schema["entry"].MinItems)
	})

	t.Run("apply is optional list", func(t *testing.T) {
		assert.True(t, resource.Schema["apply"].Optional)
	})
}

func TestResourceRTXAccessListIPv6EntrySchema(t *testing.T) {
	resource := resourceRTXAccessListIPv6()
	entrySchema := resource.Schema["entry"].Elem.(*schema.Resource).Schema

	t.Run("sequence is optional and computed", func(t *testing.T) {
		assert.True(t, entrySchema["sequence"].Optional)
		assert.True(t, entrySchema["sequence"].Computed)
	})

	t.Run("action is required", func(t *testing.T) {
		assert.True(t, entrySchema["action"].Required)
	})

	t.Run("source is required", func(t *testing.T) {
		assert.True(t, entrySchema["source"].Required)
	})

	t.Run("destination is required", func(t *testing.T) {
		assert.True(t, entrySchema["destination"].Required)
	})

	t.Run("protocol is optional with default", func(t *testing.T) {
		assert.True(t, entrySchema["protocol"].Optional)
		assert.Equal(t, "*", entrySchema["protocol"].Default)
	})

	t.Run("source_port is optional with default", func(t *testing.T) {
		assert.True(t, entrySchema["source_port"].Optional)
		assert.Equal(t, "*", entrySchema["source_port"].Default)
	})

	t.Run("dest_port is optional with default", func(t *testing.T) {
		assert.True(t, entrySchema["dest_port"].Optional)
		assert.Equal(t, "*", entrySchema["dest_port"].Default)
	})

	t.Run("log is optional with default false", func(t *testing.T) {
		assert.True(t, entrySchema["log"].Optional)
		assert.Equal(t, false, entrySchema["log"].Default)
	})
}

func TestResourceRTXAccessListIPv6SchemaValidation(t *testing.T) {
	resource := resourceRTXAccessListIPv6()
	entrySchema := resource.Schema["entry"].Elem.(*schema.Resource).Schema

	t.Run("action validation", func(t *testing.T) {
		validActions := []string{"pass", "reject", "restrict", "restrict-log"}
		for _, action := range validActions {
			_, errs := entrySchema["action"].ValidateFunc(action, "action")
			assert.Empty(t, errs, "action '%s' should be valid", action)
		}

		_, errs := entrySchema["action"].ValidateFunc("invalid", "action")
		assert.NotEmpty(t, errs, "action 'invalid' should be invalid")
	})

	t.Run("protocol validation includes icmp6", func(t *testing.T) {
		// IPv6-specific: icmp6 instead of icmp
		validProtocols := []string{"tcp", "udp", "icmp6", "ip", "gre", "esp", "ah", "*"}
		for _, proto := range validProtocols {
			_, errs := entrySchema["protocol"].ValidateFunc(proto, "protocol")
			assert.Empty(t, errs, "protocol '%s' should be valid", proto)
		}

		_, errs := entrySchema["protocol"].ValidateFunc("icmp", "protocol")
		assert.NotEmpty(t, errs, "protocol 'icmp' should be invalid for IPv6")

		_, errs = entrySchema["protocol"].ValidateFunc("invalid", "protocol")
		assert.NotEmpty(t, errs, "protocol 'invalid' should be invalid")
	})

	t.Run("sequence validation allows valid values", func(t *testing.T) {
		// Valid range: 1-65535
		_, errs := entrySchema["sequence"].ValidateFunc(1, "sequence")
		assert.Empty(t, errs, "sequence 1 should be valid")

		_, errs = entrySchema["sequence"].ValidateFunc(65535, "sequence")
		assert.Empty(t, errs, "sequence 65535 should be valid")

		_, errs = entrySchema["sequence"].ValidateFunc(65536, "sequence")
		assert.NotEmpty(t, errs, "sequence 65536 should be invalid")

		_, errs = entrySchema["sequence"].ValidateFunc(0, "sequence")
		assert.NotEmpty(t, errs, "sequence 0 should be invalid")
	})
}

func TestResourceRTXAccessListIPv6Importer(t *testing.T) {
	resource := resourceRTXAccessListIPv6()

	t.Run("importer is configured", func(t *testing.T) {
		assert.NotNil(t, resource.Importer)
		assert.NotNil(t, resource.Importer.StateContext)
	})
}

func TestResourceRTXAccessListIPv6CRUDFunctions(t *testing.T) {
	resource := resourceRTXAccessListIPv6()

	t.Run("CRUD functions are configured", func(t *testing.T) {
		assert.NotNil(t, resource.CreateContext)
		assert.NotNil(t, resource.ReadContext)
		assert.NotNil(t, resource.UpdateContext)
		assert.NotNil(t, resource.DeleteContext)
	})
}

func TestResourceRTXAccessListIPv6CustomizeDiff(t *testing.T) {
	resource := resourceRTXAccessListIPv6()

	t.Run("customize diff is configured", func(t *testing.T) {
		assert.NotNil(t, resource.CustomizeDiff)
	})
}
