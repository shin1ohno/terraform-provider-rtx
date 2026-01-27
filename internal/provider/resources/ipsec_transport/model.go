package ipsec_transport

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// IPsecTransportModel describes the resource data model.
type IPsecTransportModel struct {
	TransportID types.Int64  `tfsdk:"transport_id"`
	TunnelID    types.Int64  `tfsdk:"tunnel_id"`
	Protocol    types.String `tfsdk:"protocol"`
	Port        types.Int64  `tfsdk:"port"`
}

// ToClient converts the Terraform model to a client.IPsecTransportConfig.
func (m *IPsecTransportModel) ToClient() client.IPsecTransportConfig {
	return client.IPsecTransportConfig{
		TransportID: fwhelpers.GetInt64Value(m.TransportID),
		TunnelID:    fwhelpers.GetInt64Value(m.TunnelID),
		Protocol:    fwhelpers.GetStringValue(m.Protocol),
		Port:        fwhelpers.GetInt64Value(m.Port),
	}
}

// FromClient updates the Terraform model from a client.IPsecTransportConfig.
func (m *IPsecTransportModel) FromClient(transport *client.IPsecTransportConfig) {
	m.TransportID = types.Int64Value(int64(transport.TransportID))
	m.TunnelID = types.Int64Value(int64(transport.TunnelID))
	m.Protocol = fwhelpers.StringValueOrNull(transport.Protocol)
	m.Port = types.Int64Value(int64(transport.Port))
}
