package vlan

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// VLANModel describes the resource data model.
type VLANModel struct {
	VlanID        types.Int64  `tfsdk:"vlan_id"`
	Interface     types.String `tfsdk:"interface"`
	Name          types.String `tfsdk:"name"`
	IPAddress     types.String `tfsdk:"ip_address"`
	IPMask        types.String `tfsdk:"ip_mask"`
	Shutdown      types.Bool   `tfsdk:"shutdown"`
	VlanInterface types.String `tfsdk:"vlan_interface"`
}

// ToClient converts the Terraform model to a client.VLAN.
func (m *VLANModel) ToClient() client.VLAN {
	return client.VLAN{
		VlanID:        fwhelpers.GetInt64Value(m.VlanID),
		Interface:     fwhelpers.GetStringValue(m.Interface),
		Name:          fwhelpers.GetStringValue(m.Name),
		IPAddress:     fwhelpers.GetStringValue(m.IPAddress),
		IPMask:        fwhelpers.GetStringValue(m.IPMask),
		Shutdown:      fwhelpers.GetBoolValue(m.Shutdown),
		VlanInterface: fwhelpers.GetStringValue(m.VlanInterface),
	}
}

// FromClient updates the Terraform model from a client.VLAN.
func (m *VLANModel) FromClient(vlan *client.VLAN) {
	m.VlanID = types.Int64Value(int64(vlan.VlanID))
	m.Interface = types.StringValue(vlan.Interface)
	m.Name = fwhelpers.StringValueOrNull(vlan.Name)
	m.IPAddress = fwhelpers.StringValueOrNull(vlan.IPAddress)
	m.IPMask = fwhelpers.StringValueOrNull(vlan.IPMask)
	m.Shutdown = types.BoolValue(vlan.Shutdown)
	m.VlanInterface = types.StringValue(vlan.VlanInterface)
}

// ID returns the resource ID in "interface/vlan_id" format.
func (m *VLANModel) ID() string {
	return m.Interface.ValueString() + "/" + m.VlanID.String()
}
