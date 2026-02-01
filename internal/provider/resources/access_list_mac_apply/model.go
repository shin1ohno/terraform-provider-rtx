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
	Sequences  types.List   `tfsdk:"sequences"`
}

// GetSequencesAsInts extracts sequences as a slice of integers.
func (m *AccessListMACApplyModel) GetSequencesAsInts() []int {
	if m.Sequences.IsNull() || m.Sequences.IsUnknown() {
		return nil
	}

	elements := m.Sequences.Elements()
	result := make([]int, 0, len(elements))
	for _, elem := range elements {
		if intVal, ok := elem.(types.Int64); ok && !intVal.IsNull() && !intVal.IsUnknown() {
			result = append(result, int(intVal.ValueInt64()))
		}
	}
	return result
}

// SetSequencesFromInts sets sequences from a slice of integers.
func (m *AccessListMACApplyModel) SetSequencesFromInts(sequences []int) {
	if len(sequences) == 0 {
		m.Sequences = types.ListValueMust(types.Int64Type, []attr.Value{})
		return
	}

	elements := make([]attr.Value, len(sequences))
	for i, id := range sequences {
		elements[i] = types.Int64Value(int64(id))
	}
	m.Sequences = types.ListValueMust(types.Int64Type, elements)
}
