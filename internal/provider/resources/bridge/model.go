package bridge

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// BridgeModel describes the resource data model.
type BridgeModel struct {
	Name          types.String `tfsdk:"name"`
	InterfaceName types.String `tfsdk:"interface_name"`
	Members       types.List   `tfsdk:"members"`
}

// ToClient converts the Terraform model to a client.BridgeConfig.
func (m *BridgeModel) ToClient() client.BridgeConfig {
	bridge := client.BridgeConfig{
		Name:    fwhelpers.GetStringValue(m.Name),
		Members: []string{},
	}

	if !m.Members.IsNull() && !m.Members.IsUnknown() {
		elements := m.Members.Elements()
		for _, elem := range elements {
			if strVal, ok := elem.(types.String); ok {
				bridge.Members = append(bridge.Members, strVal.ValueString())
			}
		}
	}

	return bridge
}

// FromClient updates the Terraform model from a client.BridgeConfig.
func (m *BridgeModel) FromClient(bridge *client.BridgeConfig) {
	m.Name = types.StringValue(bridge.Name)
	m.InterfaceName = types.StringValue(bridge.Name)

	if len(bridge.Members) > 0 {
		elements := make([]attr.Value, len(bridge.Members))
		for i, member := range bridge.Members {
			elements[i] = types.StringValue(member)
		}
		m.Members = types.ListValueMust(types.StringType, elements)
	} else {
		m.Members = types.ListValueMust(types.StringType, []attr.Value{})
	}
}
