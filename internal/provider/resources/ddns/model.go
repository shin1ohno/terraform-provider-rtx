package ddns

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// DDNSModel describes the resource data model.
type DDNSModel struct {
	ServerID types.Int64  `tfsdk:"server_id"`
	URL      types.String `tfsdk:"url"`
	Hostname types.String `tfsdk:"hostname"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// ToClient converts the Terraform model to a client.DDNSServerConfig.
func (m *DDNSModel) ToClient() client.DDNSServerConfig {
	return client.DDNSServerConfig{
		ID:       fwhelpers.GetInt64Value(m.ServerID),
		URL:      fwhelpers.GetStringValue(m.URL),
		Hostname: fwhelpers.GetStringValue(m.Hostname),
		Username: fwhelpers.GetStringValue(m.Username),
		Password: fwhelpers.GetStringValue(m.Password),
	}
}

// FromClient updates the Terraform model from a client.DDNSServerConfig.
// Note: Password is not read back from router for security reasons.
func (m *DDNSModel) FromClient(config *client.DDNSServerConfig) {
	m.ServerID = types.Int64Value(int64(config.ID))
	m.URL = types.StringValue(config.URL)
	m.Hostname = types.StringValue(config.Hostname)
	m.Username = fwhelpers.StringValueOrNull(config.Username)
	// Note: password is WriteOnly, so we don't read it back
}
