package class_map

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// ClassMapModel describes the resource data model.
type ClassMapModel struct {
	Name                 types.String `tfsdk:"name"`
	MatchProtocol        types.String `tfsdk:"match_protocol"`
	MatchDestinationPort types.List   `tfsdk:"match_destination_port"`
	MatchSourcePort      types.List   `tfsdk:"match_source_port"`
	MatchDSCP            types.String `tfsdk:"match_dscp"`
	MatchFilter          types.Int64  `tfsdk:"match_filter"`
}

// ToClient converts the Terraform model to a client.ClassMap.
func (m *ClassMapModel) ToClient() client.ClassMap {
	cm := client.ClassMap{
		Name:          fwhelpers.GetStringValue(m.Name),
		MatchProtocol: fwhelpers.GetStringValue(m.MatchProtocol),
		MatchDSCP:     fwhelpers.GetStringValue(m.MatchDSCP),
		MatchFilter:   fwhelpers.GetInt64Value(m.MatchFilter),
	}

	cm.MatchDestinationPort = getInt64ListValues(m.MatchDestinationPort)
	cm.MatchSourcePort = getInt64ListValues(m.MatchSourcePort)

	return cm
}

// FromClient updates the Terraform model from a client.ClassMap.
func (m *ClassMapModel) FromClient(cm *client.ClassMap) {
	m.Name = types.StringValue(cm.Name)
	m.MatchProtocol = fwhelpers.StringValueOrNull(cm.MatchProtocol)
	m.MatchDSCP = fwhelpers.StringValueOrNull(cm.MatchDSCP)
	m.MatchFilter = fwhelpers.Int64ValueOrNull(cm.MatchFilter)

	m.MatchDestinationPort = intSliceToList(cm.MatchDestinationPort)
	m.MatchSourcePort = intSliceToList(cm.MatchSourcePort)
}

// Helper functions

func getInt64ListValues(list types.List) []int {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var result []int
	elements := list.Elements()
	for _, elem := range elements {
		if intVal, ok := elem.(types.Int64); ok {
			result = append(result, int(intVal.ValueInt64()))
		}
	}
	return result
}

func intSliceToList(slice []int) types.List {
	if len(slice) == 0 {
		return types.ListNull(types.Int64Type)
	}

	elements := make([]attr.Value, len(slice))
	for i, v := range slice {
		elements[i] = types.Int64Value(int64(v))
	}
	listVal, _ := types.ListValue(types.Int64Type, elements)
	return listVal
}
