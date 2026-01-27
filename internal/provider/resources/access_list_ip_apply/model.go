package access_list_ip_apply

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AccessListIPApplyModel describes the resource data model.
type AccessListIPApplyModel struct {
	AccessList types.String `tfsdk:"access_list"`
	Interface  types.String `tfsdk:"interface"`
	Direction  types.String `tfsdk:"direction"`
	FilterIDs  types.List   `tfsdk:"filter_ids"`
}

// GetFilterIDsAsInts returns the filter IDs as a slice of integers.
func (m *AccessListIPApplyModel) GetFilterIDsAsInts() []int {
	if m.FilterIDs.IsNull() || m.FilterIDs.IsUnknown() {
		return nil
	}

	var result []int
	elements := m.FilterIDs.Elements()
	for _, elem := range elements {
		if intVal, ok := elem.(types.Int64); ok && !intVal.IsNull() && !intVal.IsUnknown() {
			result = append(result, int(intVal.ValueInt64()))
		}
	}
	return result
}

// SetFilterIDsFromInts sets the filter IDs from a slice of integers.
func (m *AccessListIPApplyModel) SetFilterIDsFromInts(ids []int) {
	if len(ids) == 0 {
		m.FilterIDs = types.ListValueMust(types.Int64Type, []attr.Value{})
		return
	}

	elements := make([]attr.Value, len(ids))
	for i, id := range ids {
		elements[i] = types.Int64Value(int64(id))
	}
	m.FilterIDs = types.ListValueMust(types.Int64Type, elements)
}
