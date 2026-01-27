package netvolante_dns

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// NetVolanteDNSModel describes the resource data model.
type NetVolanteDNSModel struct {
	Interface    types.String `tfsdk:"interface"`
	Hostname     types.String `tfsdk:"hostname"`
	Server       types.Int64  `tfsdk:"server"`
	Timeout      types.Int64  `tfsdk:"timeout"`
	IPv6Enabled  types.Bool   `tfsdk:"ipv6_enabled"`
	AutoHostname types.Bool   `tfsdk:"auto_hostname"`
}

// ToClient converts the Terraform model to a client.NetVolanteConfig.
func (m *NetVolanteDNSModel) ToClient() client.NetVolanteConfig {
	return client.NetVolanteConfig{
		Interface:    fwhelpers.GetStringValue(m.Interface),
		Hostname:     fwhelpers.GetStringValue(m.Hostname),
		Server:       fwhelpers.GetInt64Value(m.Server),
		Timeout:      fwhelpers.GetInt64Value(m.Timeout),
		IPv6:         fwhelpers.GetBoolValue(m.IPv6Enabled),
		AutoHostname: fwhelpers.GetBoolValue(m.AutoHostname),
		Use:          true,
	}
}

// FromClient updates the Terraform model from a client.NetVolanteConfig.
func (m *NetVolanteDNSModel) FromClient(config *client.NetVolanteConfig) {
	m.Interface = types.StringValue(config.Interface)
	m.Hostname = types.StringValue(config.Hostname)
	m.Server = types.Int64Value(int64(config.Server))
	m.Timeout = types.Int64Value(int64(config.Timeout))
	m.IPv6Enabled = types.BoolValue(config.IPv6)
	m.AutoHostname = types.BoolValue(config.AutoHostname)
}
