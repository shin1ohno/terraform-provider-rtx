package class_map

import (
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

	cm.MatchDestinationPort = fwhelpers.ListToIntSlice(m.MatchDestinationPort)
	cm.MatchSourcePort = fwhelpers.ListToIntSlice(m.MatchSourcePort)

	return cm
}

// FromClient updates the Terraform model from a client.ClassMap.
func (m *ClassMapModel) FromClient(cm *client.ClassMap) {
	m.Name = types.StringValue(cm.Name)
	m.MatchProtocol = fwhelpers.StringValueOrNull(cm.MatchProtocol)
	m.MatchDSCP = fwhelpers.StringValueOrNull(cm.MatchDSCP)
	m.MatchFilter = fwhelpers.Int64ValueOrNull(cm.MatchFilter)

	// Preserve empty list vs null: absence of port match on RTX is equivalent to empty list
	if cm.MatchDestinationPort == nil && !m.MatchDestinationPort.IsNull() {
		m.MatchDestinationPort = fwhelpers.IntSliceToList([]int{})
	} else {
		m.MatchDestinationPort = fwhelpers.IntSliceToList(cm.MatchDestinationPort)
	}
	if cm.MatchSourcePort == nil && !m.MatchSourcePort.IsNull() {
		m.MatchSourcePort = fwhelpers.IntSliceToList([]int{})
	} else {
		m.MatchSourcePort = fwhelpers.IntSliceToList(cm.MatchSourcePort)
	}
}
