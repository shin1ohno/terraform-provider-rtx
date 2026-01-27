package l2tp_service

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

// L2TPServiceModel describes the resource data model.
type L2TPServiceModel struct {
	ID        types.String   `tfsdk:"id"`
	Enabled   types.Bool     `tfsdk:"enabled"`
	Protocols []types.String `tfsdk:"protocols"`
}

// ToClient converts the Terraform model to client parameters.
func (m *L2TPServiceModel) ToClient() (bool, []string) {
	enabled := m.Enabled.ValueBool()

	protocols := make([]string, len(m.Protocols))
	for i, p := range m.Protocols {
		protocols[i] = p.ValueString()
	}

	return enabled, protocols
}

// FromClient updates the Terraform model from a client.L2TPServiceState.
func (m *L2TPServiceModel) FromClient(state *client.L2TPServiceState) {
	m.ID = types.StringValue("default")
	m.Enabled = types.BoolValue(state.Enabled)

	if len(state.Protocols) > 0 {
		m.Protocols = make([]types.String, len(state.Protocols))
		for i, p := range state.Protocols {
			m.Protocols[i] = types.StringValue(p)
		}
	} else {
		m.Protocols = nil
	}
}
