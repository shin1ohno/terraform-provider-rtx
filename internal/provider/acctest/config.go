// Package acctest provides utilities for acceptance testing of Terraform resources.
package acctest

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// ConfigBuilder builds HCL configurations for tests using a fluent interface.
type ConfigBuilder struct {
	resourceType string
	resourceName string
	attributes   map[string]interface{}
	blocks       []*BlockBuilder
	dependsOn    []string
}

// BlockBuilder builds nested HCL blocks.
type BlockBuilder struct {
	blockType  string
	blockLabel string
	attributes map[string]interface{}
	blocks     []*BlockBuilder
}

// NewConfigBuilder creates a new configuration builder for a resource.
// resourceType is the Terraform resource type (e.g., "rtx_admin_user").
// resourceName is the local name for the resource (e.g., "test").
func NewConfigBuilder(resourceType, resourceName string) *ConfigBuilder {
	return &ConfigBuilder{
		resourceType: resourceType,
		resourceName: resourceName,
		attributes:   make(map[string]interface{}),
		blocks:       make([]*BlockBuilder, 0),
		dependsOn:    make([]string, 0),
	}
}

// SetAttribute sets a single attribute value.
// Supports string, int, int64, float64, bool, []string, []interface{}, and map[string]interface{}.
func (b *ConfigBuilder) SetAttribute(name string, value interface{}) *ConfigBuilder {
	b.attributes[name] = value
	return b
}

// SetAttributes sets multiple attributes at once.
func (b *ConfigBuilder) SetAttributes(attrs map[string]interface{}) *ConfigBuilder {
	for k, v := range attrs {
		b.attributes[k] = v
	}
	return b
}

// AddBlock adds a nested block to the resource.
func (b *ConfigBuilder) AddBlock(block *BlockBuilder) *ConfigBuilder {
	b.blocks = append(b.blocks, block)
	return b
}

// DependsOn adds resource dependencies.
func (b *ConfigBuilder) DependsOn(deps ...string) *ConfigBuilder {
	b.dependsOn = append(b.dependsOn, deps...)
	return b
}

// Build generates a valid HCL configuration string.
func (b *ConfigBuilder) Build() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("resource %q %q {\n", b.resourceType, b.resourceName))

	// Write attributes in sorted order for deterministic output
	writeAttributes(&buf, b.attributes, "  ")

	// Write nested blocks
	for _, block := range b.blocks {
		writeBlock(&buf, block, "  ")
	}

	// Write depends_on if present
	if len(b.dependsOn) > 0 {
		buf.WriteString("\n  depends_on = [\n")
		for _, dep := range b.dependsOn {
			buf.WriteString(fmt.Sprintf("    %s,\n", dep))
		}
		buf.WriteString("  ]\n")
	}

	buf.WriteString("}\n")

	return buf.String()
}

// ResourceAddress returns the full resource address (e.g., "rtx_admin_user.test").
func (b *ConfigBuilder) ResourceAddress() string {
	return fmt.Sprintf("%s.%s", b.resourceType, b.resourceName)
}

// NewBlockBuilder creates a new block builder.
// blockType is the block type (e.g., "timeouts", "connection").
// blockLabel is optional; use empty string for blocks without labels.
func NewBlockBuilder(blockType, blockLabel string) *BlockBuilder {
	return &BlockBuilder{
		blockType:  blockType,
		blockLabel: blockLabel,
		attributes: make(map[string]interface{}),
		blocks:     make([]*BlockBuilder, 0),
	}
}

// SetAttribute sets a single attribute on the block.
func (bb *BlockBuilder) SetAttribute(name string, value interface{}) *BlockBuilder {
	bb.attributes[name] = value
	return bb
}

// SetAttributes sets multiple attributes on the block.
func (bb *BlockBuilder) SetAttributes(attrs map[string]interface{}) *BlockBuilder {
	for k, v := range attrs {
		bb.attributes[k] = v
	}
	return bb
}

// AddBlock adds a nested block within this block.
func (bb *BlockBuilder) AddBlock(block *BlockBuilder) *BlockBuilder {
	bb.blocks = append(bb.blocks, block)
	return bb
}

// writeAttributes writes attribute key-value pairs to the buffer.
func writeAttributes(buf *bytes.Buffer, attrs map[string]interface{}, indent string) {
	// Sort keys for deterministic output
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := attrs[key]
		buf.WriteString(fmt.Sprintf("%s%s = %s\n", indent, key, formatValue(value)))
	}
}

// writeBlock writes a block to the buffer.
func writeBlock(buf *bytes.Buffer, block *BlockBuilder, indent string) {
	buf.WriteString("\n")
	if block.blockLabel != "" {
		buf.WriteString(fmt.Sprintf("%s%s %q {\n", indent, block.blockType, block.blockLabel))
	} else {
		buf.WriteString(fmt.Sprintf("%s%s {\n", indent, block.blockType))
	}

	writeAttributes(buf, block.attributes, indent+"  ")

	for _, nested := range block.blocks {
		writeBlock(buf, nested, indent+"  ")
	}

	buf.WriteString(fmt.Sprintf("%s}\n", indent))
}

// formatValue formats a Go value as an HCL value.
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return formatString(val)
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%g", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case []string:
		return formatStringSlice(val)
	case []interface{}:
		return formatInterfaceSlice(val)
	case map[string]interface{}:
		return formatMap(val)
	case map[string]string:
		return formatStringMap(val)
	case nil:
		return "null"
	default:
		// For unknown types, attempt to use string representation
		return fmt.Sprintf("%q", fmt.Sprintf("%v", val))
	}
}

// formatString properly escapes and quotes a string for HCL.
func formatString(s string) string {
	// Escape special characters
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return fmt.Sprintf("%q", s)
}

// formatStringSlice formats a string slice as an HCL list.
func formatStringSlice(ss []string) string {
	if len(ss) == 0 {
		return "[]"
	}
	parts := make([]string, len(ss))
	for i, s := range ss {
		parts[i] = formatString(s)
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
}

// formatInterfaceSlice formats an interface slice as an HCL list.
func formatInterfaceSlice(items []interface{}) string {
	if len(items) == 0 {
		return "[]"
	}
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = formatValue(item)
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
}

// formatMap formats a map as an HCL object.
func formatMap(m map[string]interface{}) string {
	if len(m) == 0 {
		return "{}"
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s = %s", k, formatValue(m[k])))
	}
	return fmt.Sprintf("{\n    %s\n  }", strings.Join(parts, "\n    "))
}

// formatStringMap formats a string map as an HCL object.
func formatStringMap(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s = %s", k, formatString(m[k])))
	}
	return fmt.Sprintf("{\n    %s\n  }", strings.Join(parts, "\n    "))
}

// MultiConfigBuilder helps build configurations with multiple resources.
type MultiConfigBuilder struct {
	builders []*ConfigBuilder
}

// NewMultiConfigBuilder creates a new multi-config builder.
func NewMultiConfigBuilder() *MultiConfigBuilder {
	return &MultiConfigBuilder{
		builders: make([]*ConfigBuilder, 0),
	}
}

// AddResource adds a ConfigBuilder to the multi-config.
func (m *MultiConfigBuilder) AddResource(b *ConfigBuilder) *MultiConfigBuilder {
	m.builders = append(m.builders, b)
	return m
}

// Build generates the combined HCL configuration.
func (m *MultiConfigBuilder) Build() string {
	var buf bytes.Buffer
	for i, b := range m.builders {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(b.Build())
	}
	return buf.String()
}
