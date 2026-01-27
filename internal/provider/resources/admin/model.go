package admin

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// AdminModel describes the resource data model.
type AdminModel struct {
	ID            types.String `tfsdk:"id"`
	LoginPassword types.String `tfsdk:"login_password"`
	AdminPassword types.String `tfsdk:"admin_password"`
	LastUpdated   types.String `tfsdk:"last_updated"`
}

// ToClient converts the Terraform model to a client.AdminConfig.
func (m *AdminModel) ToClient() client.AdminConfig {
	return client.AdminConfig{
		LoginPassword: fwhelpers.GetStringValue(m.LoginPassword),
		AdminPassword: fwhelpers.GetStringValue(m.AdminPassword),
	}
}

// FromClient updates the Terraform model from a client.AdminConfig.
// Note: Passwords are write-only and cannot be read back from the router.
func (m *AdminModel) FromClient(config *client.AdminConfig) {
	// Passwords cannot be read back for security reasons
	// Only set non-sensitive fields here if available
}
