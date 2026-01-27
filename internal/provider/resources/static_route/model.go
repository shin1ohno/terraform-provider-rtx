package static_route

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// StaticRouteModel describes the resource data model.
type StaticRouteModel struct {
	ID       types.String   `tfsdk:"id"`
	Prefix   types.String   `tfsdk:"prefix"`
	Mask     types.String   `tfsdk:"mask"`
	NextHops []NextHopModel `tfsdk:"next_hop"`
}

// NextHopModel describes the next hop nested block.
type NextHopModel struct {
	Gateway   types.String `tfsdk:"gateway"`
	Interface types.String `tfsdk:"interface"`
	Distance  types.Int64  `tfsdk:"distance"`
	Permanent types.Bool   `tfsdk:"permanent"`
	Filter    types.Int64  `tfsdk:"filter"`
}

// ToClient converts the Terraform model to a client.StaticRoute.
func (m *StaticRouteModel) ToClient() client.StaticRoute {
	route := client.StaticRoute{
		Prefix: fwhelpers.GetStringValue(m.Prefix),
		Mask:   fwhelpers.GetStringValue(m.Mask),
	}

	// Handle next hops
	if len(m.NextHops) > 0 {
		nextHops := make([]client.StaticRouteHop, len(m.NextHops))
		for i, hop := range m.NextHops {
			nextHops[i] = client.StaticRouteHop{
				NextHop:   fwhelpers.GetStringValue(hop.Gateway),
				Interface: fwhelpers.GetStringValue(hop.Interface),
				Distance:  fwhelpers.GetInt64Value(hop.Distance),
				Permanent: fwhelpers.GetBoolValue(hop.Permanent),
				Filter:    fwhelpers.GetInt64Value(hop.Filter),
			}
		}
		route.NextHops = nextHops
	}

	return route
}

// FromClient updates the Terraform model from a client.StaticRoute.
func (m *StaticRouteModel) FromClient(route *client.StaticRoute) {
	m.Prefix = types.StringValue(route.Prefix)
	m.Mask = types.StringValue(route.Mask)
	m.ID = types.StringValue(route.Prefix + "/" + route.Mask)

	// Update next hops
	if len(route.NextHops) > 0 {
		nextHops := make([]NextHopModel, len(route.NextHops))
		for i, hop := range route.NextHops {
			nextHops[i] = NextHopModel{
				Gateway:   fwhelpers.StringValueOrNull(hop.NextHop),
				Interface: fwhelpers.StringValueOrNull(hop.Interface),
				Distance:  fwhelpers.Int64ValueOrNull(hop.Distance),
				Permanent: types.BoolValue(hop.Permanent),
				Filter:    fwhelpers.Int64ValueOrNull(hop.Filter),
			}
		}
		m.NextHops = nextHops
	}
}
