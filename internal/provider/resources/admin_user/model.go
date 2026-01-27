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
func (m *AdminUserModel) FromClient(user *client.AdminUser) {
	m.Username = types.StringValue(user.Username)
	// Note: password is WriteOnly, so we don't read it back
	m.Encrypted = types.BoolValue(user.Encrypted)

	// Handle pointer fields
	if user.Attributes.Administrator != nil {
		m.Administrator = types.BoolValue(*user.Attributes.Administrator)
	} else {
		m.Administrator = types.BoolValue(false)
	}

	if user.Attributes.LoginTimer != nil {
		m.LoginTimer = types.Int64Value(int64(*user.Attributes.LoginTimer))
	} else {
		m.LoginTimer = types.Int64Null()
	}

	// Handle connection methods
	if user.Attributes.Connection != nil && len(user.Attributes.Connection) > 0 {
		m.ConnectionMethods = stringSliceToSet(user.Attributes.Connection)
	} else {
		m.ConnectionMethods = types.SetValueMust(types.StringType, []attr.Value{})
	}

	// Handle GUI pages
	if user.Attributes.GUIPages != nil && len(user.Attributes.GUIPages) > 0 {
		m.GUIPages = stringSliceToSet(user.Attributes.GUIPages)
	} else {
		m.GUIPages = types.SetValueMust(types.StringType, []attr.Value{})
	}
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
