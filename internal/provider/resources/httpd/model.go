package httpd

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// HTTPDModel describes the resource data model.
type HTTPDModel struct {
	ID          types.String `tfsdk:"id"`
	Host        types.String `tfsdk:"host"`
	ProxyAccess types.Bool   `tfsdk:"proxy_access"`
}

// ToClient converts the Terraform model to a client.HTTPDConfig.
func (m *HTTPDModel) ToClient() client.HTTPDConfig {
	return client.HTTPDConfig{
		Host:        fwhelpers.GetStringValue(m.Host),
		ProxyAccess: fwhelpers.GetBoolValue(m.ProxyAccess),
	}
}

// FromClient updates the Terraform model from a client.HTTPDConfig.
func (m *HTTPDModel) FromClient(config *client.HTTPDConfig) {
	m.ID = types.StringValue("httpd")
	m.Host = types.StringValue(config.Host)
	m.ProxyAccess = types.BoolValue(config.ProxyAccess)
}
