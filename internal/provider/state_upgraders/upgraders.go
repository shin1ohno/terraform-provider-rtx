// Package stateupgraders provides common types and utilities for Terraform state migration.
//
// # State Migration Overview
//
// Terraform providers store resource state in a schema-versioned format. When the schema
// changes in ways that require state transformation (e.g., renaming fields, restructuring
// data), the provider must implement state upgraders to migrate existing state files.
//
// # Version Numbering Convention
//
// Each resource that supports state migration should:
//  1. Define a SchemaVersion in the resource definition (integer, starts at 0)
//  2. Increment SchemaVersion when state structure changes
//  3. Provide StateUpgraders for each previous version
//
// Version 0 is the initial schema. When you make breaking changes:
//   - Increment SchemaVersion to 1
//   - Add a StateUpgrader that migrates V0 -> V1
//
// # When to Increment Schema Version
//
// Increment SchemaVersion when:
//   - Renaming an attribute (e.g., "name" -> "username")
//   - Changing attribute type (e.g., string -> list)
//   - Restructuring nested blocks
//   - Removing attributes that were previously stored in state
//
// Do NOT increment SchemaVersion when:
//   - Adding new Optional attributes (backward compatible)
//   - Adding new Computed attributes (backward compatible)
//   - Changing validation rules only
//
// # Example Usage
//
//	func resourceAdminUser() *schema.Resource {
//	    return &schema.Resource{
//	        // ... CRUD functions ...
//	        SchemaVersion: 1,
//	        StateUpgraders: []schema.StateUpgrader{
//	            stateupgraders.NewUpgrader(0, ResourceAdminUserV0Schema, UpgradeAdminUserV0),
//	        },
//	        Schema: currentSchema,
//	    }
//	}
//
// See https://developer.hashicorp.com/terraform/plugin/sdkv2/resources/state-migration
package stateupgraders

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// StateUpgradeFunc is an alias for schema.StateUpgradeFunc for documentation purposes.
// It receives the raw state as a map and returns the upgraded state.
//
// Parameters:
//   - ctx: Context for cancellation and deadline propagation
//   - rawState: The previous version's state as a map[string]interface{}
//   - meta: Provider meta information (client connections, etc.)
//
// Returns:
//   - map[string]interface{}: The upgraded state
//   - error: Any error that occurred during upgrade
type StateUpgradeFunc = schema.StateUpgradeFunc

// SchemaFunc is a function that returns the schema for a specific version.
// This is used to get the ImpliedType for state upgraders.
type SchemaFunc func() map[string]*schema.Schema

// NewUpgrader creates a schema.StateUpgrader entry for use in a resource's StateUpgraders slice.
//
// Parameters:
//   - version: The schema version this upgrader migrates FROM
//   - schemaFunc: Function that returns the schema for that version
//   - upgradeFunc: Function that performs the state migration
//
// Example:
//
//	StateUpgraders: []schema.StateUpgrader{
//	    stateupgraders.NewUpgrader(0, myResourceV0Schema, upgradeMyResourceV0),
//	    stateupgraders.NewUpgrader(1, myResourceV1Schema, upgradeMyResourceV1),
//	},
func NewUpgrader(version int, schemaFunc SchemaFunc, upgradeFunc StateUpgradeFunc) schema.StateUpgrader {
	return schema.StateUpgrader{
		Version: version,
		Type:    schemaForResource(schemaFunc()).CoreConfigSchema().ImpliedType(),
		Upgrade: upgradeFunc,
	}
}

// schemaForResource wraps a schema map into a minimal schema.Resource to get the ImpliedType.
func schemaForResource(s map[string]*schema.Schema) *schema.Resource {
	return &schema.Resource{
		Schema: s,
	}
}

// RenameAttribute is a helper function for the common pattern of renaming an attribute.
// It moves the value from oldKey to newKey and deletes the old key.
//
// Returns true if the rename was performed, false if oldKey didn't exist.
func RenameAttribute(rawState map[string]interface{}, oldKey, newKey string) bool {
	if rawState == nil {
		return false
	}
	if v, ok := rawState[oldKey]; ok {
		rawState[newKey] = v
		delete(rawState, oldKey)
		return true
	}
	return false
}

// SetDefaultIfMissing sets a default value for an attribute if it doesn't exist in state.
// This is useful when adding new required fields during state migration.
//
// Returns true if the default was set, false if the key already existed.
func SetDefaultIfMissing(rawState map[string]interface{}, key string, defaultValue interface{}) bool {
	if rawState == nil {
		return false
	}
	if _, ok := rawState[key]; !ok {
		rawState[key] = defaultValue
		return true
	}
	return false
}

// RemoveAttribute removes an attribute from the state.
// This is useful when deprecating and removing fields.
//
// Returns true if the attribute was removed, false if it didn't exist.
func RemoveAttribute(rawState map[string]interface{}, key string) bool {
	if rawState == nil {
		return false
	}
	if _, ok := rawState[key]; ok {
		delete(rawState, key)
		return true
	}
	return false
}

// TransformAttribute applies a transformation function to an attribute value.
// If the attribute doesn't exist or is nil, the transform is not applied.
//
// Returns true if the transformation was applied, false otherwise.
func TransformAttribute(rawState map[string]interface{}, key string, transform func(interface{}) interface{}) bool {
	if rawState == nil {
		return false
	}
	if v, ok := rawState[key]; ok && v != nil {
		rawState[key] = transform(v)
		return true
	}
	return false
}

// CopyAttribute copies the value from sourceKey to destKey without removing the source.
// This is useful when splitting one attribute into multiple.
//
// Returns true if the copy was performed, false if sourceKey didn't exist.
func CopyAttribute(rawState map[string]interface{}, sourceKey, destKey string) bool {
	if rawState == nil {
		return false
	}
	if v, ok := rawState[sourceKey]; ok {
		rawState[destKey] = v
		return true
	}
	return false
}

// MoveToNestedBlock moves a top-level attribute into a nested block.
// The nested block will be created as a list with a single element if it doesn't exist.
//
// Parameters:
//   - rawState: The state map to modify
//   - attrKey: The key of the attribute to move
//   - blockKey: The key of the nested block
//   - nestedAttrKey: The key to use inside the nested block
//
// Returns true if the move was performed, false otherwise.
func MoveToNestedBlock(rawState map[string]interface{}, attrKey, blockKey, nestedAttrKey string) bool {
	if rawState == nil {
		return false
	}
	v, ok := rawState[attrKey]
	if !ok {
		return false
	}

	// Get or create the nested block as a list
	var block []interface{}
	if existing, ok := rawState[blockKey].([]interface{}); ok && len(existing) > 0 {
		block = existing
	} else {
		block = []interface{}{make(map[string]interface{})}
	}

	// Set the value in the first element of the block
	if elem, ok := block[0].(map[string]interface{}); ok {
		elem[nestedAttrKey] = v
	}

	rawState[blockKey] = block
	delete(rawState, attrKey)
	return true
}

// ExtractFromNestedBlock extracts an attribute from a nested block to a top-level attribute.
// This is the reverse of MoveToNestedBlock.
//
// Parameters:
//   - rawState: The state map to modify
//   - blockKey: The key of the nested block
//   - nestedAttrKey: The key inside the nested block
//   - attrKey: The key for the top-level attribute
//
// Returns true if the extraction was performed, false otherwise.
func ExtractFromNestedBlock(rawState map[string]interface{}, blockKey, nestedAttrKey, attrKey string) bool {
	if rawState == nil {
		return false
	}

	block, ok := rawState[blockKey].([]interface{})
	if !ok || len(block) == 0 {
		return false
	}

	elem, ok := block[0].(map[string]interface{})
	if !ok {
		return false
	}

	v, ok := elem[nestedAttrKey]
	if !ok {
		return false
	}

	rawState[attrKey] = v
	delete(elem, nestedAttrKey)

	// If the nested block element is now empty, remove the block
	if len(elem) == 0 {
		delete(rawState, blockKey)
	}

	return true
}
