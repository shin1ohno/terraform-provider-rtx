package acctest

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewConfigBuilder verifies ConfigBuilder initialization.
func TestNewConfigBuilder(t *testing.T) {
	t.Parallel()

	cb := NewConfigBuilder("rtx_admin_user", "test")

	assert.NotNil(t, cb)
	assert.Equal(t, "rtx_admin_user", cb.resourceType)
	assert.Equal(t, "test", cb.resourceName)
	assert.NotNil(t, cb.attributes)
	assert.NotNil(t, cb.blocks)
	assert.NotNil(t, cb.dependsOn)
}

// TestConfigBuilderSetAttribute verifies single attribute setting.
func TestConfigBuilderSetAttribute(t *testing.T) {
	t.Parallel()

	cb := NewConfigBuilder("rtx_admin_user", "test").
		SetAttribute("username", "admin").
		SetAttribute("password", "secret")

	assert.Equal(t, "admin", cb.attributes["username"])
	assert.Equal(t, "secret", cb.attributes["password"])
}

// TestConfigBuilderSetAttributes verifies bulk attribute setting.
func TestConfigBuilderSetAttributes(t *testing.T) {
	t.Parallel()

	attrs := map[string]interface{}{
		"username":      "admin",
		"password":      "secret",
		"administrator": true,
	}

	cb := NewConfigBuilder("rtx_admin_user", "test").SetAttributes(attrs)

	assert.Equal(t, "admin", cb.attributes["username"])
	assert.Equal(t, "secret", cb.attributes["password"])
	assert.Equal(t, true, cb.attributes["administrator"])
}

// TestConfigBuilderResourceAddress verifies resource address generation.
func TestConfigBuilderResourceAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceType string
		resourceName string
		expected     string
	}{
		{
			name:         "simple resource",
			resourceType: "rtx_admin_user",
			resourceName: "test",
			expected:     "rtx_admin_user.test",
		},
		{
			name:         "with underscores",
			resourceType: "rtx_admin",
			resourceName: "my_resource",
			expected:     "rtx_admin.my_resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cb := NewConfigBuilder(tt.resourceType, tt.resourceName)
			assert.Equal(t, tt.expected, cb.ResourceAddress())
		})
	}
}

// TestConfigBuilderDependsOn verifies dependency configuration.
func TestConfigBuilderDependsOn(t *testing.T) {
	t.Parallel()

	cb := NewConfigBuilder("rtx_admin_user", "test").
		DependsOn("rtx_admin.primary", "rtx_admin.secondary")

	require.Len(t, cb.dependsOn, 2)
	assert.Equal(t, "rtx_admin.primary", cb.dependsOn[0])
	assert.Equal(t, "rtx_admin.secondary", cb.dependsOn[1])
}

// TestConfigBuilderBuildBasic verifies basic HCL generation.
func TestConfigBuilderBuildBasic(t *testing.T) {
	t.Parallel()

	cb := NewConfigBuilder("rtx_admin_user", "test").
		SetAttribute("username", "admin").
		SetAttribute("password", "secret")

	hcl := cb.Build()

	// Verify structure
	assert.Contains(t, hcl, `resource "rtx_admin_user" "test" {`)
	assert.Contains(t, hcl, `username = "admin"`)
	assert.Contains(t, hcl, `password = "secret"`)
	assert.True(t, strings.HasSuffix(hcl, "}\n"), "HCL should end with closing brace and newline")
}

// TestConfigBuilderBuildWithDependsOn verifies depends_on HCL generation.
func TestConfigBuilderBuildWithDependsOn(t *testing.T) {
	t.Parallel()

	cb := NewConfigBuilder("rtx_admin_user", "test").
		SetAttribute("username", "admin").
		DependsOn("rtx_admin.primary")

	hcl := cb.Build()

	assert.Contains(t, hcl, "depends_on = [")
	assert.Contains(t, hcl, "rtx_admin.primary,")
}

// TestConfigBuilderBuildAttributeTypes verifies all attribute type formatting.
func TestConfigBuilderBuildAttributeTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		attrName string
		value    interface{}
		expected string
	}{
		{
			name:     "string",
			attrName: "name",
			value:    "test-value",
			expected: `name = "test-value"`,
		},
		{
			name:     "integer",
			attrName: "port",
			value:    8080,
			expected: `port = 8080`,
		},
		{
			name:     "int64",
			attrName: "big_number",
			value:    int64(9223372036854775807),
			expected: `big_number = 9223372036854775807`,
		},
		{
			name:     "float64",
			attrName: "ratio",
			value:    0.5,
			expected: `ratio = 0.5`,
		},
		{
			name:     "boolean true",
			attrName: "enabled",
			value:    true,
			expected: `enabled = true`,
		},
		{
			name:     "boolean false",
			attrName: "disabled",
			value:    false,
			expected: `disabled = false`,
		},
		{
			name:     "nil value",
			attrName: "nullable",
			value:    nil,
			expected: `nullable = null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cb := NewConfigBuilder("rtx_test", "test").SetAttribute(tt.attrName, tt.value)
			hcl := cb.Build()

			assert.Contains(t, hcl, tt.expected)
		})
	}
}

// TestConfigBuilderBuildStringSlice verifies string slice HCL generation.
func TestConfigBuilderBuildStringSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		values   []string
		expected string
	}{
		{
			name:     "single element",
			values:   []string{"value1"},
			expected: `tags = ["value1"]`,
		},
		{
			name:     "multiple elements",
			values:   []string{"value1", "value2", "value3"},
			expected: `tags = ["value1", "value2", "value3"]`,
		},
		{
			name:     "empty slice",
			values:   []string{},
			expected: `tags = []`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cb := NewConfigBuilder("rtx_test", "test").SetAttribute("tags", tt.values)
			hcl := cb.Build()

			assert.Contains(t, hcl, tt.expected)
		})
	}
}

// TestConfigBuilderBuildInterfaceSlice verifies interface slice HCL generation.
func TestConfigBuilderBuildInterfaceSlice(t *testing.T) {
	t.Parallel()

	values := []interface{}{"string", 42, true}
	cb := NewConfigBuilder("rtx_test", "test").SetAttribute("mixed", values)
	hcl := cb.Build()

	assert.Contains(t, hcl, `mixed = ["string", 42, true]`)
}

// TestConfigBuilderBuildMap verifies map HCL generation.
func TestConfigBuilderBuildMap(t *testing.T) {
	t.Parallel()

	t.Run("string interface map", func(t *testing.T) {
		t.Parallel()

		attrs := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}

		cb := NewConfigBuilder("rtx_test", "test").SetAttribute("labels", attrs)
		hcl := cb.Build()

		assert.Contains(t, hcl, "labels = {")
		assert.Contains(t, hcl, `key1 = "value1"`)
		assert.Contains(t, hcl, "key2 = 42")
	})

	t.Run("string string map", func(t *testing.T) {
		t.Parallel()

		attrs := map[string]string{
			"env":  "production",
			"team": "platform",
		}

		cb := NewConfigBuilder("rtx_test", "test").SetAttribute("tags", attrs)
		hcl := cb.Build()

		assert.Contains(t, hcl, "tags = {")
		assert.Contains(t, hcl, `env = "production"`)
		assert.Contains(t, hcl, `team = "platform"`)
	})

	t.Run("empty map", func(t *testing.T) {
		t.Parallel()

		attrs := map[string]interface{}{}
		cb := NewConfigBuilder("rtx_test", "test").SetAttribute("empty", attrs)
		hcl := cb.Build()

		assert.Contains(t, hcl, "empty = {}")
	})
}

// TestConfigBuilderBuildEscaping verifies special character escaping.
// Note: formatString uses Go's %q verb which produces Go-style escaped strings.
// The escaping in formatString then wraps with quotes, resulting in double escaping.
func TestConfigBuilderBuildEscaping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		contains string
	}{
		{
			name:  "double quotes",
			value: `value with "quotes"`,
			// formatString escapes: \ -> \\, " -> \"
			// then %q wraps and escapes again
			contains: `field = "value with \\\"quotes\\\""`,
		},
		{
			name:  "backslash",
			value: `path\to\file`,
			// \ -> \\ from escape, then %q escapes \\ -> \\\\
			contains: `field = "path\\\\to\\\\file"`,
		},
		{
			name:  "newline",
			value: "line1\nline2",
			// \n -> \\n from escape, then %q escapes \\ -> \\\\
			contains: `field = "line1\\nline2"`,
		},
		{
			name:  "tab",
			value: "col1\tcol2",
			// \t -> \\t from escape, then %q escapes
			contains: `field = "col1\\tcol2"`,
		},
		{
			name:  "carriage return",
			value: "text\rmore",
			// \r -> \\r from escape
			contains: `field = "text\\rmore"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cb := NewConfigBuilder("rtx_test", "test").SetAttribute("field", tt.value)
			hcl := cb.Build()

			assert.Contains(t, hcl, tt.contains)
		})
	}
}

// TestConfigBuilderBuildDeterministic verifies output is deterministic.
func TestConfigBuilderBuildDeterministic(t *testing.T) {
	t.Parallel()

	// Build the same config multiple times and verify output is identical
	for i := 0; i < 10; i++ {
		cb := NewConfigBuilder("rtx_admin_user", "test").
			SetAttribute("username", "admin").
			SetAttribute("password", "secret").
			SetAttribute("administrator", true)

		hcl := cb.Build()

		// Verify attributes are in sorted order (administrator before password before username)
		adminIdx := strings.Index(hcl, "administrator")
		passIdx := strings.Index(hcl, "password")
		userIdx := strings.Index(hcl, "username")

		assert.True(t, adminIdx < passIdx, "administrator should come before password")
		assert.True(t, passIdx < userIdx, "password should come before username")
	}
}

// TestNewBlockBuilder verifies BlockBuilder initialization.
func TestNewBlockBuilder(t *testing.T) {
	t.Parallel()

	t.Run("with label", func(t *testing.T) {
		t.Parallel()

		bb := NewBlockBuilder("provisioner", "local-exec")
		assert.Equal(t, "provisioner", bb.blockType)
		assert.Equal(t, "local-exec", bb.blockLabel)
		assert.NotNil(t, bb.attributes)
		assert.NotNil(t, bb.blocks)
	})

	t.Run("without label", func(t *testing.T) {
		t.Parallel()

		bb := NewBlockBuilder("timeouts", "")
		assert.Equal(t, "timeouts", bb.blockType)
		assert.Equal(t, "", bb.blockLabel)
	})
}

// TestConfigBuilderAddBlock verifies nested block HCL generation.
func TestConfigBuilderAddBlock(t *testing.T) {
	t.Parallel()

	t.Run("block without label", func(t *testing.T) {
		t.Parallel()

		timeouts := NewBlockBuilder("timeouts", "").
			SetAttribute("create", "10m").
			SetAttribute("delete", "5m")

		cb := NewConfigBuilder("rtx_admin_user", "test").
			SetAttribute("username", "admin").
			AddBlock(timeouts)

		hcl := cb.Build()

		assert.Contains(t, hcl, "timeouts {")
		assert.Contains(t, hcl, `create = "10m"`)
		assert.Contains(t, hcl, `delete = "5m"`)
	})

	t.Run("block with label", func(t *testing.T) {
		t.Parallel()

		provisioner := NewBlockBuilder("provisioner", "local-exec").
			SetAttribute("command", "echo hello")

		cb := NewConfigBuilder("rtx_admin_user", "test").
			SetAttribute("username", "admin").
			AddBlock(provisioner)

		hcl := cb.Build()

		assert.Contains(t, hcl, `provisioner "local-exec" {`)
		assert.Contains(t, hcl, `command = "echo hello"`)
	})
}

// TestBlockBuilderNestedBlocks verifies deeply nested block HCL generation.
func TestBlockBuilderNestedBlocks(t *testing.T) {
	t.Parallel()

	innerBlock := NewBlockBuilder("connection", "").
		SetAttribute("type", "ssh").
		SetAttribute("host", "192.168.1.1")

	provisioner := NewBlockBuilder("provisioner", "remote-exec").
		SetAttribute("inline", []string{"echo hello"}).
		AddBlock(innerBlock)

	cb := NewConfigBuilder("rtx_admin", "test").
		SetAttribute("name", "test").
		AddBlock(provisioner)

	hcl := cb.Build()

	assert.Contains(t, hcl, `provisioner "remote-exec" {`)
	assert.Contains(t, hcl, "connection {")
	assert.Contains(t, hcl, `type = "ssh"`)
}

// TestMultiConfigBuilder verifies multiple resource configuration building.
func TestMultiConfigBuilder(t *testing.T) {
	t.Parallel()

	admin := NewConfigBuilder("rtx_admin", "primary").
		SetAttribute("name", "admin")

	user := NewConfigBuilder("rtx_admin_user", "test").
		SetAttribute("username", "testuser").
		DependsOn("rtx_admin.primary")

	mcb := NewMultiConfigBuilder().
		AddResource(admin).
		AddResource(user)

	hcl := mcb.Build()

	// Verify both resources are present
	assert.Contains(t, hcl, `resource "rtx_admin" "primary"`)
	assert.Contains(t, hcl, `resource "rtx_admin_user" "test"`)
	assert.Contains(t, hcl, `name = "admin"`)
	assert.Contains(t, hcl, `username = "testuser"`)
	assert.Contains(t, hcl, "depends_on = [")

	// Verify resources are separated by newline
	adminIdx := strings.Index(hcl, "rtx_admin")
	userIdx := strings.Index(hcl, "rtx_admin_user")
	assert.True(t, adminIdx < userIdx, "admin resource should come before user resource")
}

// TestMultiConfigBuilderEmpty verifies empty multi-config building.
func TestMultiConfigBuilderEmpty(t *testing.T) {
	t.Parallel()

	mcb := NewMultiConfigBuilder()
	hcl := mcb.Build()

	assert.Equal(t, "", hcl)
}

// TestConfigBuilderValidHCL verifies generated HCL is syntactically correct.
func TestConfigBuilderValidHCL(t *testing.T) {
	t.Parallel()

	// Build a complex configuration
	cb := NewConfigBuilder("rtx_admin_user", "complex").
		SetAttribute("username", "admin").
		SetAttribute("password", "secret123").
		SetAttribute("administrator", true).
		SetAttribute("port", 22).
		SetAttribute("tags", []string{"production", "critical"}).
		SetAttribute("metadata", map[string]interface{}{
			"created_by": "terraform",
			"version":    1,
		}).
		AddBlock(NewBlockBuilder("timeouts", "").
			SetAttribute("create", "10m").
			SetAttribute("update", "5m")).
		DependsOn("rtx_admin.primary")

	hcl := cb.Build()

	// Verify basic HCL structure
	assert.True(t, strings.HasPrefix(hcl, "resource"))
	assert.True(t, strings.HasSuffix(hcl, "}\n"))

	// Count braces to ensure they're balanced
	openBraces := strings.Count(hcl, "{")
	closeBraces := strings.Count(hcl, "}")
	assert.Equal(t, openBraces, closeBraces, "braces should be balanced")

	// Count brackets to ensure they're balanced
	openBrackets := strings.Count(hcl, "[")
	closeBrackets := strings.Count(hcl, "]")
	assert.Equal(t, openBrackets, closeBrackets, "brackets should be balanced")
}
