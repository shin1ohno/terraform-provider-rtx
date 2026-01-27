package interface_resource

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// InterfaceModel describes the resource data model.
type InterfaceModel struct {
	Name          types.String    `tfsdk:"name"`
	InterfaceName types.String    `tfsdk:"interface_name"`
	Description   types.String    `tfsdk:"description"`
	IPAddress     *IPAddressModel `tfsdk:"ip_address"`
	NATDescriptor types.Int64     `tfsdk:"nat_descriptor"`
	ProxyARP      types.Bool      `tfsdk:"proxyarp"`
	MTU           types.Int64     `tfsdk:"mtu"`
}

// IPAddressModel describes the IP address nested block.
type IPAddressModel struct {
	Address types.String `tfsdk:"address"`
	DHCP    types.Bool   `tfsdk:"dhcp"`
}

// ToClient converts the Terraform model to a client.InterfaceConfig.
func (m *InterfaceModel) ToClient() client.InterfaceConfig {
	config := client.InterfaceConfig{
		Name:          fwhelpers.GetStringValue(m.Name),
		Description:   fwhelpers.GetStringValue(m.Description),
		NATDescriptor: fwhelpers.GetInt64Value(m.NATDescriptor),
		ProxyARP:      fwhelpers.GetBoolValue(m.ProxyARP),
		MTU:           fwhelpers.GetInt64Value(m.MTU),
	}

	// Handle IP address block
	if m.IPAddress != nil {
		config.IPAddress = &client.InterfaceIP{
			Address: fwhelpers.GetStringValue(m.IPAddress.Address),
			DHCP:    fwhelpers.GetBoolValue(m.IPAddress.DHCP),
		}
	}

	return config
}

// FromClient updates the Terraform model from a client.InterfaceConfig.
func (m *InterfaceModel) FromClient(config *client.InterfaceConfig) {
	m.Name = types.StringValue(config.Name)
	m.InterfaceName = types.StringValue(config.Name)
	m.Description = fwhelpers.StringValueOrNull(config.Description)
	m.NATDescriptor = fwhelpers.Int64ValueOrNull(config.NATDescriptor)
	m.ProxyARP = types.BoolValue(config.ProxyARP)
	m.MTU = fwhelpers.Int64ValueOrNull(config.MTU)

	// Handle IP address block
	if config.IPAddress != nil && (config.IPAddress.Address != "" || config.IPAddress.DHCP) {
		if m.IPAddress == nil {
			m.IPAddress = &IPAddressModel{}
		}
		m.IPAddress.Address = fwhelpers.StringValueOrNull(config.IPAddress.Address)
		m.IPAddress.DHCP = types.BoolValue(config.IPAddress.DHCP)
	} else {
		m.IPAddress = nil
	}
}
