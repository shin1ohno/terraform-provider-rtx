package ospf

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// OSPFModel describes the resource data model.
type OSPFModel struct {
	ID                          types.String `tfsdk:"id"`
	ProcessID                   types.Int64  `tfsdk:"process_id"`
	RouterID                    types.String `tfsdk:"router_id"`
	Distance                    types.Int64  `tfsdk:"distance"`
	DefaultInformationOriginate types.Bool   `tfsdk:"default_information_originate"`
	Networks                    types.List   `tfsdk:"network"`
	Areas                       types.List   `tfsdk:"area"`
	Neighbors                   types.List   `tfsdk:"neighbor"`
	RedistributeStatic          types.Bool   `tfsdk:"redistribute_static"`
	RedistributeConnected       types.Bool   `tfsdk:"redistribute_connected"`
}

// NetworkModel describes a network block within the OSPF resource.
type NetworkModel struct {
	IP       types.String `tfsdk:"ip"`
	Wildcard types.String `tfsdk:"wildcard"`
	Area     types.String `tfsdk:"area"`
}

// AreaModel describes an area block within the OSPF resource.
type AreaModel struct {
	AreaID    types.String `tfsdk:"area_id"`
	Type      types.String `tfsdk:"type"`
	NoSummary types.Bool   `tfsdk:"no_summary"`
}

// NeighborModel describes a neighbor block within the OSPF resource.
type NeighborModel struct {
	IP       types.String `tfsdk:"ip"`
	Priority types.Int64  `tfsdk:"priority"`
	Cost     types.Int64  `tfsdk:"cost"`
}

// NetworkModelAttrTypes returns the attribute types for NetworkModel.
func NetworkModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ip":       types.StringType,
		"wildcard": types.StringType,
		"area":     types.StringType,
	}
}

// AreaModelAttrTypes returns the attribute types for AreaModel.
func AreaModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"area_id":    types.StringType,
		"type":       types.StringType,
		"no_summary": types.BoolType,
	}
}

// NeighborModelAttrTypes returns the attribute types for NeighborModel.
func NeighborModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ip":       types.StringType,
		"priority": types.Int64Type,
		"cost":     types.Int64Type,
	}
}

// ToClient converts the Terraform model to a client.OSPFConfig.
func (m *OSPFModel) ToClient() client.OSPFConfig {
	config := client.OSPFConfig{
		Enabled:               true,
		ProcessID:             fwhelpers.GetInt64Value(m.ProcessID),
		RouterID:              fwhelpers.GetStringValue(m.RouterID),
		Distance:              fwhelpers.GetInt64Value(m.Distance),
		DefaultOriginate:      fwhelpers.GetBoolValue(m.DefaultInformationOriginate),
		RedistributeStatic:    fwhelpers.GetBoolValue(m.RedistributeStatic),
		RedistributeConnected: fwhelpers.GetBoolValue(m.RedistributeConnected),
	}

	// Convert networks
	if !m.Networks.IsNull() && !m.Networks.IsUnknown() {
		var networks []NetworkModel
		m.Networks.ElementsAs(nil, &networks, false)
		config.Networks = make([]client.OSPFNetwork, len(networks))
		for i, n := range networks {
			config.Networks[i] = client.OSPFNetwork{
				IP:       fwhelpers.GetStringValue(n.IP),
				Wildcard: fwhelpers.GetStringValue(n.Wildcard),
				Area:     fwhelpers.GetStringValue(n.Area),
			}
		}
	}

	// Convert areas
	if !m.Areas.IsNull() && !m.Areas.IsUnknown() {
		var areas []AreaModel
		m.Areas.ElementsAs(nil, &areas, false)
		config.Areas = make([]client.OSPFArea, len(areas))
		for i, a := range areas {
			config.Areas[i] = client.OSPFArea{
				ID:        fwhelpers.GetStringValue(a.AreaID),
				Type:      fwhelpers.GetStringValue(a.Type),
				NoSummary: fwhelpers.GetBoolValue(a.NoSummary),
			}
		}
	}

	// Convert neighbors
	if !m.Neighbors.IsNull() && !m.Neighbors.IsUnknown() {
		var neighbors []NeighborModel
		m.Neighbors.ElementsAs(nil, &neighbors, false)
		config.Neighbors = make([]client.OSPFNeighbor, len(neighbors))
		for i, n := range neighbors {
			config.Neighbors[i] = client.OSPFNeighbor{
				IP:       fwhelpers.GetStringValue(n.IP),
				Priority: fwhelpers.GetInt64Value(n.Priority),
				Cost:     fwhelpers.GetInt64Value(n.Cost),
			}
		}
	}

	return config
}

// FromClient updates the Terraform model from a client.OSPFConfig.
func (m *OSPFModel) FromClient(config *client.OSPFConfig) {
	m.ID = types.StringValue("ospf")
	m.ProcessID = types.Int64Value(int64(config.ProcessID))
	m.RouterID = types.StringValue(config.RouterID)
	m.Distance = types.Int64Value(int64(config.Distance))
	m.DefaultInformationOriginate = types.BoolValue(config.DefaultOriginate)
	m.RedistributeStatic = types.BoolValue(config.RedistributeStatic)
	m.RedistributeConnected = types.BoolValue(config.RedistributeConnected)

	// Convert networks
	if len(config.Networks) > 0 {
		networkElements := make([]attr.Value, len(config.Networks))
		for i, n := range config.Networks {
			networkElements[i], _ = types.ObjectValue(
				NetworkModelAttrTypes(),
				map[string]attr.Value{
					"ip":       types.StringValue(n.IP),
					"wildcard": types.StringValue(n.Wildcard),
					"area":     types.StringValue(n.Area),
				},
			)
		}
		m.Networks, _ = types.ListValue(types.ObjectType{AttrTypes: NetworkModelAttrTypes()}, networkElements)
	} else {
		m.Networks = types.ListNull(types.ObjectType{AttrTypes: NetworkModelAttrTypes()})
	}

	// Convert areas
	if len(config.Areas) > 0 {
		areaElements := make([]attr.Value, len(config.Areas))
		for i, a := range config.Areas {
			areaElements[i], _ = types.ObjectValue(
				AreaModelAttrTypes(),
				map[string]attr.Value{
					"area_id":    types.StringValue(a.ID),
					"type":       types.StringValue(a.Type),
					"no_summary": types.BoolValue(a.NoSummary),
				},
			)
		}
		m.Areas, _ = types.ListValue(types.ObjectType{AttrTypes: AreaModelAttrTypes()}, areaElements)
	} else {
		m.Areas = types.ListNull(types.ObjectType{AttrTypes: AreaModelAttrTypes()})
	}

	// Convert neighbors
	if len(config.Neighbors) > 0 {
		neighborElements := make([]attr.Value, len(config.Neighbors))
		for i, n := range config.Neighbors {
			neighborElements[i], _ = types.ObjectValue(
				NeighborModelAttrTypes(),
				map[string]attr.Value{
					"ip":       types.StringValue(n.IP),
					"priority": types.Int64Value(int64(n.Priority)),
					"cost":     types.Int64Value(int64(n.Cost)),
				},
			)
		}
		m.Neighbors, _ = types.ListValue(types.ObjectType{AttrTypes: NeighborModelAttrTypes()}, neighborElements)
	} else {
		m.Neighbors = types.ListNull(types.ObjectType{AttrTypes: NeighborModelAttrTypes()})
	}
}
