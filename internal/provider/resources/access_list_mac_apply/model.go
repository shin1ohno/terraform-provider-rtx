package access_list_mac_apply

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AccessListMACApplyModel describes the resource data model.
type AccessListMACApplyModel struct {
	ID         types.String `tfsdk:"id"`
	AccessList types.String `tfsdk:"access_list"`
	Interface  types.String `tfsdk:"interface"`
	Direction  types.String `tfsdk:"direction"`
	FilterIDs  types.List   `tfsdk:"filter_ids"`
}

// GetFilterIDsAsInts extracts filter IDs as a slice of integers.
func (m *AccessListMACApplyModel) GetFilterIDsAsInts() []int {
	if m.FilterIDs.IsNull() || m.FilterIDs.IsUnknown() {
		return nil
	}

	elements := m.FilterIDs.Elements()
	result := make([]int, 0, len(elements))
	for _, elem := range elements {
		if intVal, ok := elem.(types.Int64); ok && !intVal.IsNull() && !intVal.IsUnknown() {
			result = append(result, int(intVal.ValueInt64()))
		}
	}
	return result
}

// SetFilterIDsFromInts sets filter IDs from a slice of integers.
func (m *AccessListMACApplyModel) SetFilterIDsFromInts(filterIDs []int) {
	if len(filterIDs) == 0 {
		m.FilterIDs = types.ListValueMust(types.Int64Type, []attr.Value{})
		return
	}

	elements := make([]attr.Value, len(filterIDs))
	for i, id := range filterIDs {
		elements[i] = types.Int64Value(int64(id))
	}
	m.FilterIDs = types.ListValueMust(types.Int64Type, elements)
}
