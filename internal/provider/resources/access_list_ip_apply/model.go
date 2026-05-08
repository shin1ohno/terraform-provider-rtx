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
	Sequences  types.List   `tfsdk:"sequences"`
}

// GetSequencesAsInts returns the sequences as a slice of integers.
func (m *AccessListIPApplyModel) GetSequencesAsInts() []int {
	if m.Sequences.IsNull() || m.Sequences.IsUnknown() {
		return nil
	}

	var result []int
	elements := m.Sequences.Elements()
	for _, elem := range elements {
		if intVal, ok := elem.(types.Int64); ok && !intVal.IsNull() && !intVal.IsUnknown() {
			result = append(result, int(intVal.ValueInt64()))
		}
	}
	return result
}

// SetSequencesFromInts sets the sequences from a slice of integers.
// Preserves prior null vs empty distinction: when the router returns no IDs,
// leave state null if the config never set the attribute, otherwise materialize
// an empty list. Mirrors the null-preservation invariant used by snmp_server's
// FromClient (commit 00c069b) — framework rejects "was null, but now empty list"
// on the apply consistency check otherwise.
func (m *AccessListIPApplyModel) SetSequencesFromInts(ids []int) {
	if len(ids) > 0 {
		elements := make([]attr.Value, len(ids))
		for i, id := range ids {
			elements[i] = types.Int64Value(int64(id))
		}
		m.Sequences = types.ListValueMust(types.Int64Type, elements)
	} else if m.Sequences.IsNull() {
		m.Sequences = types.ListNull(types.Int64Type)
	} else {
		m.Sequences = types.ListValueMust(types.Int64Type, []attr.Value{})
	}
}
