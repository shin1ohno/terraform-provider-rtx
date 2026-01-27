package sshd_authorized_keys

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

// SSHDAuthorizedKeysModel describes the resource data model.
type SSHDAuthorizedKeysModel struct {
	Username types.String `tfsdk:"username"`
	Keys     types.List   `tfsdk:"keys"`
	KeyCount types.Int64  `tfsdk:"key_count"`
}

// ToKeyStrings converts the Keys list to []string for the client.
func (m *SSHDAuthorizedKeysModel) ToKeyStrings() []string {
	if m.Keys.IsNull() || m.Keys.IsUnknown() {
		return nil
	}

	elements := m.Keys.Elements()
	result := make([]string, len(elements))
	for i, elem := range elements {
		if strVal, ok := elem.(types.String); ok {
			result[i] = strVal.ValueString()
		}
	}
	return result
}

// FromClient updates the Terraform model from client.SSHAuthorizedKey slice.
func (m *SSHDAuthorizedKeysModel) FromClient(keys []client.SSHAuthorizedKey) {
	m.KeyCount = types.Int64Value(int64(len(keys)))

	// Reconstruct full key strings from parsed keys
	// Format: <type> <base64-key> [comment]
	keyStrings := make([]attr.Value, len(keys))
	for i, key := range keys {
		var keyStr string
		if key.Comment != "" {
			keyStr = fmt.Sprintf("%s %s %s", key.Type, key.Fingerprint, key.Comment)
		} else {
			keyStr = fmt.Sprintf("%s %s", key.Type, key.Fingerprint)
		}
		keyStrings[i] = types.StringValue(keyStr)
	}

	m.Keys = types.ListValueMust(types.StringType, keyStrings)
}
