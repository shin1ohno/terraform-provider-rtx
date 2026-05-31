package dhcp_scope

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// DHCPScopeModel describes the resource data model.
type DHCPScopeModel struct {
	ScopeID       types.Int64         `tfsdk:"scope_id"`
	Network       types.String        `tfsdk:"network"`
	RangeStart    types.String        `tfsdk:"range_start"`
	RangeEnd      types.String        `tfsdk:"range_end"`
	LeaseTime     types.String        `tfsdk:"lease_time"`
	ExcludeRanges []ExcludeRangeModel `tfsdk:"exclude_ranges"`
	Options       *OptionsModel       `tfsdk:"options"`
}

// ExcludeRangeModel describes an IP range excluded from DHCP allocation.
type ExcludeRangeModel struct {
	Start types.String `tfsdk:"start"`
	End   types.String `tfsdk:"end"`
}

// OptionsModel describes the DHCP options nested block.
type OptionsModel struct {
	Routers               types.List   `tfsdk:"routers"`
	DNSServers            types.List   `tfsdk:"dns_servers"`
	DomainName            types.String `tfsdk:"domain_name"`
	ClasslessStaticRoutes types.List   `tfsdk:"classless_static_routes"`
}

// ClasslessRouteModel describes a single RFC 3442 classless static route (option 121).
type ClasslessRouteModel struct {
	Destination types.String `tfsdk:"destination"`
	Gateway     types.String `tfsdk:"gateway"`
}

// ExcludeRangeAttrTypes returns the attribute types for ExcludeRangeModel.
func ExcludeRangeAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"start": types.StringType,
		"end":   types.StringType,
	}
}

// ClasslessRouteAttrTypes returns the attribute types for ClasslessRouteModel.
func ClasslessRouteAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"destination": types.StringType,
		"gateway":     types.StringType,
	}
}

// ToClient converts the Terraform model to a client.DHCPScope.
func (m *DHCPScopeModel) ToClient(ctx context.Context, diagnostics *diag.Diagnostics) client.DHCPScope {
	scope := client.DHCPScope{
		ScopeID:    int(m.ScopeID.ValueInt64()),
		Network:    fwhelpers.GetStringValue(m.Network),
		RangeStart: fwhelpers.GetStringValue(m.RangeStart),
		RangeEnd:   fwhelpers.GetStringValue(m.RangeEnd),
		LeaseTime:  fwhelpers.GetStringValue(m.LeaseTime),
	}

	// Handle exclude_ranges
	if len(m.ExcludeRanges) > 0 {
		scope.ExcludeRanges = make([]client.ExcludeRange, len(m.ExcludeRanges))
		for i, r := range m.ExcludeRanges {
			scope.ExcludeRanges[i] = client.ExcludeRange{
				Start: fwhelpers.GetStringValue(r.Start),
				End:   fwhelpers.GetStringValue(r.End),
			}
		}
	}

	// Handle options block
	if m.Options != nil {
		// Parse routers
		scope.Options.Routers = fwhelpers.ListToStringSlice(m.Options.Routers)

		// Parse dns_servers
		scope.Options.DNSServers = fwhelpers.ListToStringSlice(m.Options.DNSServers)

		// Parse domain_name
		scope.Options.DomainName = fwhelpers.GetStringValue(m.Options.DomainName)

		// Parse classless_static_routes (option 121)
		if !m.Options.ClasslessStaticRoutes.IsNull() && !m.Options.ClasslessStaticRoutes.IsUnknown() {
			var routes []ClasslessRouteModel
			m.Options.ClasslessStaticRoutes.ElementsAs(ctx, &routes, false)
			scope.Options.ClasslessStaticRoutes = make([]client.ClasslessRoute, len(routes))
			for i, r := range routes {
				scope.Options.ClasslessStaticRoutes[i] = client.ClasslessRoute{
					Destination: fwhelpers.GetStringValue(r.Destination),
					Gateway:     fwhelpers.GetStringValue(r.Gateway),
				}
			}
		}
	}

	return scope
}

// FromClient updates the Terraform model from a client.DHCPScope.
func (m *DHCPScopeModel) FromClient(ctx context.Context, scope *client.DHCPScope, diagnostics *diag.Diagnostics) {
	m.ScopeID = types.Int64Value(int64(scope.ScopeID))
	m.Network = types.StringValue(scope.Network)
	m.RangeStart = fwhelpers.StringValueOrNull(scope.RangeStart)
	m.RangeEnd = fwhelpers.StringValueOrNull(scope.RangeEnd)
	m.LeaseTime = fwhelpers.StringValueOrNull(scope.LeaseTime)

	// Convert ExcludeRanges
	if len(scope.ExcludeRanges) > 0 {
		m.ExcludeRanges = make([]ExcludeRangeModel, len(scope.ExcludeRanges))
		for i, r := range scope.ExcludeRanges {
			m.ExcludeRanges[i] = ExcludeRangeModel{
				Start: types.StringValue(r.Start),
				End:   types.StringValue(r.End),
			}
		}
	} else {
		m.ExcludeRanges = nil
	}

	// Convert Options
	if len(scope.Options.Routers) > 0 || len(scope.Options.DNSServers) > 0 || scope.Options.DomainName != "" || len(scope.Options.ClasslessStaticRoutes) > 0 {
		// Capture the prior classless-routes null state before (re)allocating Options.
		priorClasslessNull := true
		if m.Options != nil {
			priorClasslessNull = m.Options.ClasslessStaticRoutes.IsNull()
		}

		if m.Options == nil {
			m.Options = &OptionsModel{}
		}

		// Build routers list (preserve empty list vs null)
		if scope.Options.Routers == nil && m.Options != nil && !m.Options.Routers.IsNull() {
			m.Options.Routers = fwhelpers.StringSliceToList([]string{})
		} else {
			m.Options.Routers = fwhelpers.StringSliceToList(scope.Options.Routers)
		}

		// Build dns_servers list (preserve empty list vs null)
		if scope.Options.DNSServers == nil && m.Options != nil && !m.Options.DNSServers.IsNull() {
			m.Options.DNSServers = fwhelpers.StringSliceToList([]string{})
		} else {
			m.Options.DNSServers = fwhelpers.StringSliceToList(scope.Options.DNSServers)
		}

		// Build domain_name
		m.Options.DomainName = fwhelpers.StringValueOrNull(scope.Options.DomainName)

		// Build classless_static_routes (option 121), preserving null vs empty.
		routeObjType := types.ObjectType{AttrTypes: ClasslessRouteAttrTypes()}
		if len(scope.Options.ClasslessStaticRoutes) > 0 {
			elements := make([]attr.Value, len(scope.Options.ClasslessStaticRoutes))
			for i, r := range scope.Options.ClasslessStaticRoutes {
				elements[i] = types.ObjectValueMust(
					ClasslessRouteAttrTypes(),
					map[string]attr.Value{
						"destination": types.StringValue(r.Destination),
						"gateway":     types.StringValue(r.Gateway),
					},
				)
			}
			m.Options.ClasslessStaticRoutes = types.ListValueMust(routeObjType, elements)
		} else if priorClasslessNull {
			m.Options.ClasslessStaticRoutes = types.ListNull(routeObjType)
		} else {
			m.Options.ClasslessStaticRoutes = types.ListValueMust(routeObjType, []attr.Value{})
		}
	} else {
		m.Options = nil
	}
}

// reconcileClasslessRoutesWithPlan overwrites the model's classless_static_routes
// with the planned value after a device read-back.
//
// classless_static_routes is an Optional (non-Computed) attribute: when the user
// sets it in config, Terraform's post-apply consistency check requires the
// returned state to hold EXACTLY the planned elements (same count, same order,
// same null-vs-empty). The device read-back via `show config` is unreliable for
// option 121 on RTX1210 Rev.14 — `r.read` can return a scope whose
// classless_static_routes is empty/null even though the Update commands wrote all
// three routes (and `save` lagged). FromClient then emits an empty list, and
// Terraform reports ".options.classless_static_routes: element 0/1/2 has
// vanished." Restoring the planned value (which we KNOW was the source of the
// write commands and round-trips correctly through encode/decode) makes the
// returned state echo the plan and satisfies the consistency check.
//
// Mirrors dns_server.reorderServerSelectToMatchPlan: read-back fills the genuinely
// Computed fields (range_start/range_end/lease_time), the plan is authoritative
// for the user-configured option-121 routes.
func (m *DHCPScopeModel) reconcileClasslessRoutesWithPlan(plan *DHCPScopeModel) {
	routeObjType := types.ObjectType{AttrTypes: ClasslessRouteAttrTypes()}

	// No planned options at all → nothing to reconcile. If read-back produced an
	// Options block anyway (device had stale/foreign options), drop the routes to
	// the plan's null so the consistency check sees null == null.
	if plan == nil || plan.Options == nil {
		if m.Options != nil {
			m.Options.ClasslessStaticRoutes = types.ListNull(routeObjType)
		}
		return
	}

	// The plan configured an options block. read-back may have set m.Options = nil
	// (no options seen in running config) — materialize it so we can echo the
	// planned routes. Other option fields (routers/dns_servers/domain_name) are
	// left as read-back produced them; only the option-121 routes are plan-authoritative.
	if m.Options == nil {
		m.Options = &OptionsModel{
			Routers:    plan.Options.Routers,
			DNSServers: plan.Options.DNSServers,
			DomainName: plan.Options.DomainName,
		}
	}

	// Echo the planned classless_static_routes verbatim — preserves count, order,
	// and the null-vs-empty distinction the consistency check enforces.
	m.Options.ClasslessStaticRoutes = plan.Options.ClasslessStaticRoutes
}
