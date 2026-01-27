package bgp

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// BGPModel describes the resource data model.
type BGPModel struct {
	ID                    types.String `tfsdk:"id"`
	ASN                   types.String `tfsdk:"asn"`
	RouterID              types.String `tfsdk:"router_id"`
	DefaultIPv4Unicast    types.Bool   `tfsdk:"default_ipv4_unicast"`
	LogNeighborChanges    types.Bool   `tfsdk:"log_neighbor_changes"`
	Neighbors             types.List   `tfsdk:"neighbor"`
	Networks              types.List   `tfsdk:"network"`
	RedistributeStatic    types.Bool   `tfsdk:"redistribute_static"`
	RedistributeConnected types.Bool   `tfsdk:"redistribute_connected"`
}

// NeighborModel describes the neighbor nested block data model.
type NeighborModel struct {
	Index        types.Int64  `tfsdk:"index"`
	IP           types.String `tfsdk:"ip"`
	RemoteAS     types.String `tfsdk:"remote_as"`
	HoldTime     types.Int64  `tfsdk:"hold_time"`
	Keepalive    types.Int64  `tfsdk:"keepalive"`
	Multihop     types.Int64  `tfsdk:"multihop"`
	Password     types.String `tfsdk:"password"`
	LocalAddress types.String `tfsdk:"local_address"`
}

// NetworkModel describes the network nested block data model.
type NetworkModel struct {
	Prefix types.String `tfsdk:"prefix"`
	Mask   types.String `tfsdk:"mask"`
}

// NeighborAttrTypes returns the attribute types for the neighbor nested block.
func NeighborAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"index":         types.Int64Type,
		"ip":            types.StringType,
		"remote_as":     types.StringType,
		"hold_time":     types.Int64Type,
		"keepalive":     types.Int64Type,
		"multihop":      types.Int64Type,
		"password":      types.StringType,
		"local_address": types.StringType,
	}
}

// NetworkAttrTypes returns the attribute types for the network nested block.
func NetworkAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"prefix": types.StringType,
		"mask":   types.StringType,
	}
}

// ToClient converts the Terraform model to a client.BGPConfig.
func (m *BGPModel) ToClient() client.BGPConfig {
	config := client.BGPConfig{
		Enabled:               true,
		ASN:                   fwhelpers.GetStringValue(m.ASN),
		RouterID:              fwhelpers.GetStringValue(m.RouterID),
		DefaultIPv4Unicast:    fwhelpers.GetBoolValue(m.DefaultIPv4Unicast),
		LogNeighborChanges:    fwhelpers.GetBoolValue(m.LogNeighborChanges),
		RedistributeStatic:    fwhelpers.GetBoolValue(m.RedistributeStatic),
		RedistributeConnected: fwhelpers.GetBoolValue(m.RedistributeConnected),
	}

	// Convert neighbors
	if !m.Neighbors.IsNull() && !m.Neighbors.IsUnknown() {
		var neighbors []NeighborModel
		m.Neighbors.ElementsAs(context.TODO(), &neighbors, false)
		config.Neighbors = make([]client.BGPNeighbor, len(neighbors))
		for i, n := range neighbors {
			config.Neighbors[i] = client.BGPNeighbor{
				ID:           fwhelpers.GetInt64Value(n.Index),
				IP:           fwhelpers.GetStringValue(n.IP),
				RemoteAS:     fwhelpers.GetStringValue(n.RemoteAS),
				HoldTime:     fwhelpers.GetInt64Value(n.HoldTime),
				Keepalive:    fwhelpers.GetInt64Value(n.Keepalive),
				Multihop:     fwhelpers.GetInt64Value(n.Multihop),
				Password:     fwhelpers.GetStringValue(n.Password),
				LocalAddress: fwhelpers.GetStringValue(n.LocalAddress),
			}
		}
	}

	// Convert networks
	if !m.Networks.IsNull() && !m.Networks.IsUnknown() {
		var networks []NetworkModel
		m.Networks.ElementsAs(context.TODO(), &networks, false)
		config.Networks = make([]client.BGPNetwork, len(networks))
		for i, n := range networks {
			config.Networks[i] = client.BGPNetwork{
				Prefix: fwhelpers.GetStringValue(n.Prefix),
				Mask:   fwhelpers.GetStringValue(n.Mask),
			}
		}
	}

	return config
}

// FromClient updates the Terraform model from a client.BGPConfig.
func (m *BGPModel) FromClient(config *client.BGPConfig) {
	m.ID = types.StringValue("bgp")
	m.ASN = types.StringValue(config.ASN)
	m.RouterID = fwhelpers.StringValueOrNull(config.RouterID)
	m.DefaultIPv4Unicast = types.BoolValue(config.DefaultIPv4Unicast)
	m.LogNeighborChanges = types.BoolValue(config.LogNeighborChanges)
	m.RedistributeStatic = types.BoolValue(config.RedistributeStatic)
	m.RedistributeConnected = types.BoolValue(config.RedistributeConnected)

	// Convert neighbors
	if len(config.Neighbors) > 0 {
		neighbors := make([]attr.Value, len(config.Neighbors))
		for i, n := range config.Neighbors {
			neighbors[i] = types.ObjectValueMust(
				NeighborAttrTypes(),
				map[string]attr.Value{
					"index":         types.Int64Value(int64(n.ID)),
					"ip":            types.StringValue(n.IP),
					"remote_as":     types.StringValue(n.RemoteAS),
					"hold_time":     fwhelpers.Int64ValueOrNull(n.HoldTime),
					"keepalive":     fwhelpers.Int64ValueOrNull(n.Keepalive),
					"multihop":      fwhelpers.Int64ValueOrNull(n.Multihop),
					"password":      fwhelpers.StringValueOrNull(n.Password),
					"local_address": fwhelpers.StringValueOrNull(n.LocalAddress),
				},
			)
		}
		m.Neighbors = types.ListValueMust(types.ObjectType{AttrTypes: NeighborAttrTypes()}, neighbors)
	} else {
		m.Neighbors = types.ListValueMust(types.ObjectType{AttrTypes: NeighborAttrTypes()}, []attr.Value{})
	}

	// Convert networks
	if len(config.Networks) > 0 {
		networks := make([]attr.Value, len(config.Networks))
		for i, n := range config.Networks {
			networks[i] = types.ObjectValueMust(
				NetworkAttrTypes(),
				map[string]attr.Value{
					"prefix": types.StringValue(n.Prefix),
					"mask":   types.StringValue(n.Mask),
				},
			)
		}
		m.Networks = types.ListValueMust(types.ObjectType{AttrTypes: NetworkAttrTypes()}, networks)
	} else {
		m.Networks = types.ListValueMust(types.ObjectType{AttrTypes: NetworkAttrTypes()}, []attr.Value{})
	}
}
