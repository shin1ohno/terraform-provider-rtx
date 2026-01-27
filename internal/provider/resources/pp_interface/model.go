package pp_interface

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// PPInterfaceModel describes the resource data model.
type PPInterfaceModel struct {
	ID            types.String `tfsdk:"id"`
	PPNumber      types.Int64  `tfsdk:"pp_number"`
	IPAddress     types.String `tfsdk:"ip_address"`
	MTU           types.Int64  `tfsdk:"mtu"`
	TCPMSS        types.Int64  `tfsdk:"tcp_mss"`
	NATDescriptor types.Int64  `tfsdk:"nat_descriptor"`
	PPInterface   types.String `tfsdk:"pp_interface"`
}

// ToClient converts the Terraform model to a client.PPIPConfig.
func (m *PPInterfaceModel) ToClient() client.PPIPConfig {
	return client.PPIPConfig{
		Address:       fwhelpers.GetStringValue(m.IPAddress),
		MTU:           fwhelpers.GetInt64Value(m.MTU),
		TCPMSSLimit:   fwhelpers.GetInt64Value(m.TCPMSS),
		NATDescriptor: fwhelpers.GetInt64Value(m.NATDescriptor),
	}
}

// FromClient updates the Terraform model from a client.PPIPConfig.
func (m *PPInterfaceModel) FromClient(ppNum int, config *client.PPIPConfig) {
	m.ID = types.StringValue(fmt.Sprintf("%d", ppNum))
	m.PPNumber = types.Int64Value(int64(ppNum))
	m.IPAddress = types.StringValue(config.Address)
	m.MTU = types.Int64Value(int64(config.MTU))
	m.TCPMSS = types.Int64Value(int64(config.TCPMSSLimit))
	m.NATDescriptor = types.Int64Value(int64(config.NATDescriptor))
	m.PPInterface = types.StringValue(fmt.Sprintf("pp%d", ppNum))
}
