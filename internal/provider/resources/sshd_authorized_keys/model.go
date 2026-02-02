package sshd_authorized_keys

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

// DefaultComment is the comment RTX adds when no comment is specified
const DefaultComment = "no comment"

// KeyModel represents a single SSH authorized key entry
type KeyModel struct {
	Key     types.String `tfsdk:"key"`
	Comment types.String `tfsdk:"comment"`
}

// KeyModelAttrTypes returns the attribute types for KeyModel
func KeyModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":     types.StringType,
		"comment": types.StringType,
	}
}

// SSHDAuthorizedKeysModel describes the resource data model.
type SSHDAuthorizedKeysModel struct {
	Username types.String `tfsdk:"username"`
	Keys     types.List   `tfsdk:"keys"` // List of KeyModel objects
	KeyCount types.Int64  `tfsdk:"key_count"`
}

// ToKeyStrings converts the Keys list to []string for the client.
// Format: "<type> <base64-key> <comment>"
func (m *SSHDAuthorizedKeysModel) ToKeyStrings(ctx context.Context, diags *diag.Diagnostics) []string {
	if m.Keys.IsNull() || m.Keys.IsUnknown() {
		return nil
	}

	var keyModels []KeyModel
	d := m.Keys.ElementsAs(ctx, &keyModels, false)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	result := make([]string, len(keyModels))
	for i, km := range keyModels {
		key := km.Key.ValueString()
		comment := km.Comment.ValueString()
		if comment == "" {
			comment = DefaultComment
		}
		result[i] = fmt.Sprintf("%s %s", key, comment)
	}
	return result
}

// FromClient updates the Terraform model from client.SSHAuthorizedKey slice.
func (m *SSHDAuthorizedKeysModel) FromClient(ctx context.Context, keys []client.SSHAuthorizedKey, diags *diag.Diagnostics) {
	m.KeyCount = types.Int64Value(int64(len(keys)))

	keyObjects := make([]attr.Value, len(keys))
	for i, key := range keys {
		keyStr := fmt.Sprintf("%s %s", key.Type, key.Fingerprint)

		comment := key.Comment
		if comment == "" {
			comment = DefaultComment
		}

		keyObj, d := types.ObjectValue(
			KeyModelAttrTypes(),
			map[string]attr.Value{
				"key":     types.StringValue(keyStr),
				"comment": types.StringValue(comment),
			},
		)
		diags.Append(d...)
		keyObjects[i] = keyObj
	}

	listVal, d := types.ListValue(types.ObjectType{AttrTypes: KeyModelAttrTypes()}, keyObjects)
	diags.Append(d...)
	m.Keys = listVal
}
