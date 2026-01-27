package admin_user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// AdminUserModel describes the resource data model.
type AdminUserModel struct {
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
	Encrypted         types.Bool   `tfsdk:"encrypted"`
	Administrator     types.Bool   `tfsdk:"administrator"`
	ConnectionMethods types.Set    `tfsdk:"connection_methods"`
	GUIPages          types.Set    `tfsdk:"gui_pages"`
	LoginTimer        types.Int64  `tfsdk:"login_timer"`
}

// ToClient converts the Terraform model to a client.AdminUser.
func (m *AdminUserModel) ToClient() client.AdminUser {
	user := client.AdminUser{
		Username:  fwhelpers.GetStringValue(m.Username),
		Password:  fwhelpers.GetStringValue(m.Password),
		Encrypted: fwhelpers.GetBoolValue(m.Encrypted),
		Attributes: client.AdminUserAttributes{
			Administrator: boolPtr(fwhelpers.GetBoolValue(m.Administrator)),
			LoginTimer:    intPtr(fwhelpers.GetInt64Value(m.LoginTimer)),
			Connection:    getStringSetValues(m.ConnectionMethods),
			GUIPages:      getStringSetValues(m.GUIPages),
		},
	}

	// Ensure slices are not nil
	if user.Attributes.Connection == nil {
		user.Attributes.Connection = []string{}
	}
	if user.Attributes.GUIPages == nil {
		user.Attributes.GUIPages = []string{}
	}

	return user
}

// FromClient updates the Terraform model from a client.AdminUser.
// Note: Some fields are not returned correctly by the router, so we preserve the existing
// model values for those fields. This is essential for preventing drift on refresh.
func (m *AdminUserModel) FromClient(user *client.AdminUser) {
	m.Username = types.StringValue(user.Username)
	// Note: password is WriteOnly, so we don't read it back
	// Note: encrypted is config-only - router may return different value, so we preserve existing
	// (Update function will restore planned value)

	// Administrator is the only attribute that the router returns correctly
	if user.Attributes.Administrator != nil {
		m.Administrator = types.BoolValue(*user.Attributes.Administrator)
	} else if m.Administrator.IsNull() || m.Administrator.IsUnknown() {
		// Only set default if not already configured
		m.Administrator = types.BoolValue(false)
	}
	// else: preserve existing model value

	// Note: The following attributes are config-only - the router may return different values
	// or use different naming conventions. We preserve the existing model values to prevent
	// drift on refresh. The Update function ensures planned values are used after updates.

	// LoginTimer: preserve existing value (router may not return it consistently)
	// ConnectionMethods: preserve existing value (router uses different naming)
	// GUIPages: preserve existing value (router may not return it consistently)
}

// Helper functions

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}

func getStringSetValues(set types.Set) []string {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}

	var result []string
	elements := set.Elements()
	for _, elem := range elements {
		if strVal, ok := elem.(types.String); ok {
			result = append(result, strVal.ValueString())
		}
	}
	return result
}

func stringSliceToSet(slice []string) types.Set {
	elements := make([]attr.Value, len(slice))
	for i, s := range slice {
		elements[i] = types.StringValue(s)
	}
	setVal, _ := types.SetValue(types.StringType, elements)
	return setVal
}

// Helper to convert set to string slice with context (for diagnostics if needed)
func getStringSetValuesWithContext(ctx context.Context, set types.Set) []string {
	return getStringSetValues(set)
}
