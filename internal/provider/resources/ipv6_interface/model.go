package ipv6_interface

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// IPv6InterfaceModel describes the resource data model.
type IPv6InterfaceModel struct {
	Interface     types.String       `tfsdk:"interface"`
	Address       []IPv6AddressModel `tfsdk:"address"`
	RTADV         *RTADVModel        `tfsdk:"rtadv"`
	DHCPv6Service types.String       `tfsdk:"dhcpv6_service"`
	MTU           types.Int64        `tfsdk:"mtu"`
}

// IPv6AddressModel describes an IPv6 address block.
type IPv6AddressModel struct {
	Address     types.String `tfsdk:"address"`
	PrefixRef   types.String `tfsdk:"prefix_ref"`
	InterfaceID types.String `tfsdk:"interface_id"`
}

// RTADVModel describes the Router Advertisement configuration.
type RTADVModel struct {
	Enabled  types.Bool  `tfsdk:"enabled"`
	PrefixID types.Int64 `tfsdk:"prefix_id"`
	OFlag    types.Bool  `tfsdk:"o_flag"`
	MFlag    types.Bool  `tfsdk:"m_flag"`
	Lifetime types.Int64 `tfsdk:"lifetime"`
}

// ToClient converts the Terraform model to a client.IPv6InterfaceConfig.
func (m *IPv6InterfaceModel) ToClient(ctx context.Context, diagnostics *diag.Diagnostics) client.IPv6InterfaceConfig {
	config := client.IPv6InterfaceConfig{
		Interface:     fwhelpers.GetStringValue(m.Interface),
		DHCPv6Service: fwhelpers.GetStringValue(m.DHCPv6Service),
		MTU:           fwhelpers.GetInt64Value(m.MTU),
	}

	// Handle address blocks
	if len(m.Address) > 0 {
		config.Addresses = make([]client.IPv6Address, len(m.Address))
		for i, addr := range m.Address {
			config.Addresses[i] = client.IPv6Address{
				Address:     fwhelpers.GetStringValue(addr.Address),
				PrefixRef:   fwhelpers.GetStringValue(addr.PrefixRef),
				InterfaceID: fwhelpers.GetStringValue(addr.InterfaceID),
			}
		}
	}

	// Handle rtadv block
	if m.RTADV != nil {
		config.RTADV = &client.RTADVConfig{
			Enabled:  fwhelpers.GetBoolValue(m.RTADV.Enabled),
			PrefixID: fwhelpers.GetInt64Value(m.RTADV.PrefixID),
			OFlag:    fwhelpers.GetBoolValue(m.RTADV.OFlag),
			MFlag:    fwhelpers.GetBoolValue(m.RTADV.MFlag),
			Lifetime: fwhelpers.GetInt64Value(m.RTADV.Lifetime),
		}
	}

	return config
}

// FromClient updates the Terraform model from a client.IPv6InterfaceConfig.
func (m *IPv6InterfaceModel) FromClient(ctx context.Context, config *client.IPv6InterfaceConfig, diagnostics *diag.Diagnostics) {
	m.Interface = types.StringValue(config.Interface)
	m.DHCPv6Service = fwhelpers.StringValueOrNull(config.DHCPv6Service)
	m.MTU = fwhelpers.Int64ValueOrNull(config.MTU)

	// Convert Addresses
	if len(config.Addresses) > 0 {
		m.Address = make([]IPv6AddressModel, len(config.Addresses))
		for i, addr := range config.Addresses {
			m.Address[i] = IPv6AddressModel{
				Address:     fwhelpers.StringValueOrNull(addr.Address),
				PrefixRef:   fwhelpers.StringValueOrNull(addr.PrefixRef),
				InterfaceID: fwhelpers.StringValueOrNull(addr.InterfaceID),
			}
		}
	} else {
		m.Address = nil
	}

	// Convert RTADV
	if config.RTADV != nil && config.RTADV.Enabled {
		if m.RTADV == nil {
			m.RTADV = &RTADVModel{}
		}
		m.RTADV.Enabled = types.BoolValue(config.RTADV.Enabled)
		m.RTADV.PrefixID = types.Int64Value(int64(config.RTADV.PrefixID))
		m.RTADV.OFlag = types.BoolValue(config.RTADV.OFlag)
		m.RTADV.MFlag = types.BoolValue(config.RTADV.MFlag)
		m.RTADV.Lifetime = fwhelpers.Int64ValueOrNull(config.RTADV.Lifetime)
	} else {
		m.RTADV = nil
	}
}
