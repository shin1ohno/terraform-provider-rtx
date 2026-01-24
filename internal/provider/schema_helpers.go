// Package provider contains schema helper functions for Terraform provider development.
//
// This file provides reusable schema patterns for sensitive and write-only fields
// in the terraform-plugin-sdk/v2. These helpers establish consistent patterns
// across all resources in the provider.
package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// WriteOnlyStringSchema returns a schema for write-only string fields in SDK v2.
//
// This is a best-effort implementation for SDK v2 which does not natively support
// write-only attributes. The Sensitive flag ensures the value is not displayed
// in plan output, but the value will still be stored in state.
//
// To achieve true write-only behavior in SDK v2:
//  1. Use this schema definition for the attribute
//  2. In the Read function, do NOT attempt to read this value from the remote system
//  3. The value from state will be preserved across reads
//
// When migrating to terraform-plugin-framework, replace with:
//
//	schema.StringAttribute{
//	    Optional:  true,
//	    Sensitive: true,
//	    WriteOnly: true,  // Available in terraform-plugin-framework
//	}
//
// SDK v2 Limitations:
//   - Value is still stored in state (though encrypted by Terraform)
//   - Cannot truly prevent value from appearing in state file
//   - Must rely on Read function discipline to not overwrite with empty string
//
// Parameters:
//   - description: Human-readable description of the field's purpose
//
// Returns:
//   - *schema.Schema configured for write-only-like behavior in SDK v2
func WriteOnlyStringSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Sensitive:   true,
		Description: description + " (write-only: value is sent to device but not read back)",
	}
}

// WriteOnlyRequiredStringSchema returns a schema for required write-only string fields.
//
// Similar to WriteOnlyStringSchema but marks the field as required instead of optional.
// Use this for credentials or secrets that must be provided during resource creation.
//
// SDK v2 Limitations:
//   - Same limitations as WriteOnlyStringSchema
//   - Value must be provided on every apply if the resource is being modified
//
// Parameters:
//   - description: Human-readable description of the field's purpose
//
// Returns:
//   - *schema.Schema configured for required write-only-like behavior in SDK v2
func WriteOnlyRequiredStringSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Sensitive:   true,
		Description: description + " (write-only: value is sent to device but not read back)",
	}
}

// SensitiveStringSchema returns a schema for sensitive but readable fields.
//
// Unlike WriteOnlyStringSchema, this is for fields that contain sensitive data
// but can and should be read back from the remote system. The value will be
// stored in state and masked in plan/apply output.
//
// Use cases:
//   - API keys that need to be compared for drift detection
//   - Secrets that the remote system can return (though often hashed)
//   - Configuration values that happen to be sensitive
//
// Parameters:
//   - description: Human-readable description of the field's purpose
//   - required: Whether the field is required (true) or optional (false)
//
// Returns:
//   - *schema.Schema configured for sensitive string field
func SensitiveStringSchema(description string, required bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Required:    required,
		Optional:    !required,
		Sensitive:   true,
		Description: description,
	}
}

// SensitiveComputedStringSchema returns a schema for sensitive computed fields.
//
// Use this for sensitive values that are generated or assigned by the remote
// system and cannot be set by the user. Examples include generated tokens,
// certificates, or other system-assigned secrets.
//
// Parameters:
//   - description: Human-readable description of the field's purpose
//
// Returns:
//   - *schema.Schema configured for sensitive computed string field
func SensitiveComputedStringSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
		Description: description,
	}
}

// ImmutableStringSchema returns a schema for immutable identifier fields.
//
// Use this for fields that cannot be changed after resource creation and
// require the resource to be destroyed and recreated if modified.
// Common examples include usernames, resource IDs, or other unique identifiers.
//
// SDK v2 uses ForceNew: true to implement this behavior.
// In terraform-plugin-framework, use RequiresReplace() plan modifier instead.
//
// Parameters:
//   - description: Human-readable description of the field's purpose
//
// Returns:
//   - *schema.Schema configured for immutable required string field
func ImmutableStringSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: description + " (cannot be changed after creation)",
	}
}

// ImmutableOptionalStringSchema returns a schema for optional immutable fields.
//
// Similar to ImmutableStringSchema but for optional fields. If the value is set
// during creation and later changed, the resource will be destroyed and recreated.
//
// Parameters:
//   - description: Human-readable description of the field's purpose
//
// Returns:
//   - *schema.Schema configured for immutable optional string field
func ImmutableOptionalStringSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Description: description + " (cannot be changed after creation)",
	}
}

// OptionalComputedStringSchema returns a schema for optional fields with API defaults.
//
// Use this for fields where the user may provide a value, but if omitted,
// the remote system will assign a default value. The Computed flag ensures
// that Terraform will not show a diff when the user doesn't specify a value
// and the API provides one.
//
// Important: When reading state, always set this field to the value returned
// by the remote system to enable drift detection.
//
// Parameters:
//   - description: Human-readable description of the field's purpose
//
// Returns:
//   - *schema.Schema configured for optional field with computed default
func OptionalComputedStringSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: description,
	}
}

// OptionalComputedBoolSchema returns a schema for optional boolean fields with API defaults.
//
// Similar to OptionalComputedStringSchema but for boolean values.
//
// Parameters:
//   - description: Human-readable description of the field's purpose
//
// Returns:
//   - *schema.Schema configured for optional boolean with computed default
func OptionalComputedBoolSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: description,
	}
}

// OptionalComputedIntSchema returns a schema for optional integer fields with API defaults.
//
// Similar to OptionalComputedStringSchema but for integer values.
//
// Parameters:
//   - description: Human-readable description of the field's purpose
//
// Returns:
//   - *schema.Schema configured for optional integer with computed default
func OptionalComputedIntSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
		Description: description,
	}
}
