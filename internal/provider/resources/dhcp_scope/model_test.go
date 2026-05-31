package dhcp_scope

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

// buildClasslessRouteList constructs a types.List of classless_static_routes
// entries for test setup, matching the resource schema's nested object shape.
func buildClasslessRouteList(t *testing.T, routes []ClasslessRouteModel) types.List {
	t.Helper()
	routeObjType := types.ObjectType{AttrTypes: ClasslessRouteAttrTypes()}
	elements := make([]attr.Value, len(routes))
	for i, r := range routes {
		elements[i] = types.ObjectValueMust(
			ClasslessRouteAttrTypes(),
			map[string]attr.Value{
				"destination": r.Destination,
				"gateway":     r.Gateway,
			},
		)
	}
	return types.ListValueMust(routeObjType, elements)
}

// extractClasslessRoutes pulls the (destination, gateway) pairs out of a model's
// options.classless_static_routes list for assertion.
func extractClasslessRoutes(t *testing.T, m *DHCPScopeModel) []ClasslessRouteModel {
	t.Helper()
	if m.Options == nil {
		return nil
	}
	if m.Options.ClasslessStaticRoutes.IsNull() || m.Options.ClasslessStaticRoutes.IsUnknown() {
		return nil
	}
	var out []ClasslessRouteModel
	diags := m.Options.ClasslessStaticRoutes.ElementsAs(context.Background(), &out, false)
	if diags.HasError() {
		t.Fatalf("failed to extract classless routes: %v", diags)
	}
	return out
}

// homeMonitorPlanRoutes is the exact home-monitor rtx_dhcp_scope.ebisu_main
// classless_static_routes shape that triggered the v0.14.2 "element N has
// vanished" inconsistent-result-after-apply error:
//
//	0.0.0.0/0       -> 192.168.1.253   (default)
//	10.33.128.0/18  -> 192.168.1.60    (vpc_cidr)
//	100.64.0.0/10   -> 192.168.1.60    (tailscale_cgnat_cidr)
func homeMonitorPlanRoutes() []ClasslessRouteModel {
	return []ClasslessRouteModel{
		{Destination: types.StringValue("0.0.0.0/0"), Gateway: types.StringValue("192.168.1.253")},
		{Destination: types.StringValue("10.33.128.0/18"), Gateway: types.StringValue("192.168.1.60")},
		{Destination: types.StringValue("100.64.0.0/10"), Gateway: types.StringValue("192.168.1.60")},
	}
}

// newPlannedScopeModel builds a DHCPScopeModel as Terraform would deliver it from
// the plan: scope 1 with the home-monitor range, options, and 3 classless routes.
func newPlannedScopeModel(t *testing.T) DHCPScopeModel {
	t.Helper()
	routers := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("192.168.1.253")})
	dns := types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("192.168.1.61"),
		types.StringValue("1.1.1.1"),
	})
	return DHCPScopeModel{
		ScopeID:    types.Int64Value(1),
		Network:    types.StringValue("192.168.0.0/16"),
		RangeStart: types.StringValue("192.168.1.20"),
		RangeEnd:   types.StringValue("192.168.1.99"),
		LeaseTime:  types.StringValue("12h"),
		Options: &OptionsModel{
			Routers:               routers,
			DNSServers:            dns,
			DomainName:            types.StringNull(),
			ClasslessStaticRoutes: buildClasslessRouteList(t, homeMonitorPlanRoutes()),
		},
	}
}

// TestUpdateReadBack_ClasslessRoutesSurviveDeviceDrop reproduces the v0.14.2
// "Provider produced inconsistent result after apply: element 0 has vanished"
// failure. It simulates the Update path: the plan holds 3 classless routes, the
// device WRITE succeeds, but the post-apply read-back (FromClient) returns a
// scope whose option-121 routes were dropped (RTX1210 Rev.14 running-vs-saved
// lag). Before the fix, the returned model has 0 routes — Terraform rejects it.
// After reconcileClasslessRoutesWithPlan, the model echoes all 3 planned routes
// in order.
func TestUpdateReadBack_ClasslessRoutesSurviveDeviceDrop(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	// 1. Start from the plan (what req.Plan.Get returns into `data`).
	data := newPlannedScopeModel(t)

	// 2. Capture a separate copy of the plan (what req.Plan.Get returns into
	//    `planData` — the new second Get in Create/Update).
	planData := newPlannedScopeModel(t)

	// 3. Simulate read-back from the device. This is the failure mode commit #10
	//    documented: the running config the provider reads back DOES NOT contain
	//    the option-121 routes (they were written but `show config` returned the
	//    pre-change / lagged state). dns_servers and routers DID round-trip.
	readBack := &client.DHCPScope{
		ScopeID:    1,
		Network:    "192.168.0.0/16",
		RangeStart: "192.168.1.20",
		RangeEnd:   "192.168.1.99",
		LeaseTime:  "12h",
		Options: client.DHCPScopeOptions{
			Routers:    []string{"192.168.1.253"},
			DNSServers: []string{"192.168.1.61", "1.1.1.1"},
			// ClasslessStaticRoutes deliberately empty — the device dropped them.
		},
	}
	data.FromClient(ctx, readBack, &diags)
	if diags.HasError() {
		t.Fatalf("FromClient returned diagnostics: %v", diags)
	}

	// Sanity: this IS the bug. Without reconciliation the read-back model has
	// fewer routes than the plan, which is exactly what makes Terraform report
	// "element N has vanished".
	preFix := extractClasslessRoutes(t, &data)
	if len(preFix) == len(homeMonitorPlanRoutes()) {
		t.Fatalf("test premise invalid: read-back unexpectedly preserved %d routes; "+
			"the fixture must simulate the device DROPPING option-121 routes", len(preFix))
	}

	// 4. Apply the fix: reconcile the read-back state with the plan.
	data.reconcileClasslessRoutesWithPlan(&planData)

	// 5. Assert the returned state now echoes the plan EXACTLY: 3 routes, in order.
	got := extractClasslessRoutes(t, &data)
	want := homeMonitorPlanRoutes()
	if len(got) != len(want) {
		t.Fatalf("classless route count = %d, want %d (routes vanished)", len(got), len(want))
	}
	for i := range want {
		if got[i].Destination.ValueString() != want[i].Destination.ValueString() {
			t.Errorf("route[%d] destination = %q, want %q", i,
				got[i].Destination.ValueString(), want[i].Destination.ValueString())
		}
		if got[i].Gateway.ValueString() != want[i].Gateway.ValueString() {
			t.Errorf("route[%d] gateway = %q, want %q", i,
				got[i].Gateway.ValueString(), want[i].Gateway.ValueString())
		}
	}

	// The reconciled list must equal the plan list value-for-value so Terraform's
	// consistency check (planned == returned) passes.
	if !data.Options.ClasslessStaticRoutes.Equal(planData.Options.ClasslessStaticRoutes) {
		t.Errorf("reconciled classless_static_routes != planned value\n got: %v\nwant: %v",
			data.Options.ClasslessStaticRoutes, planData.Options.ClasslessStaticRoutes)
	}
}

// TestReconcileClasslessRoutes_NullPlanStaysNull verifies that when the plan did
// not configure classless_static_routes, reconciliation does not invent routes
// from a device read-back (which would itself be an inconsistent result).
func TestReconcileClasslessRoutes_NullPlanStaysNull(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	// Plan: scope with options but NO classless routes (null list).
	// `data` and `planData` are built independently — in the real resource path
	// they come from two separate req.Plan.Get calls and must NOT alias the same
	// *OptionsModel (FromClient mutates data.Options in place).
	newNullPlan := func() DHCPScopeModel {
		return DHCPScopeModel{
			ScopeID: types.Int64Value(1),
			Network: types.StringValue("192.168.0.0/16"),
			Options: &OptionsModel{
				Routers:               types.ListValueMust(types.StringType, []attr.Value{types.StringValue("192.168.1.253")}),
				DNSServers:            types.ListNull(types.StringType),
				DomainName:            types.StringNull(),
				ClasslessStaticRoutes: types.ListNull(types.ObjectType{AttrTypes: ClasslessRouteAttrTypes()}),
			},
		}
	}
	planData := newNullPlan()
	data := newNullPlan()

	// Device read-back happens to return some foreign option-121 routes.
	readBack := &client.DHCPScope{
		ScopeID: 1,
		Network: "192.168.0.0/16",
		Options: client.DHCPScopeOptions{
			Routers: []string{"192.168.1.253"},
			ClasslessStaticRoutes: []client.ClasslessRoute{
				{Destination: "0.0.0.0/0", Gateway: "192.168.1.1"},
			},
		},
	}
	data.FromClient(ctx, readBack, &diags)
	if diags.HasError() {
		t.Fatalf("FromClient returned diagnostics: %v", diags)
	}

	data.reconcileClasslessRoutesWithPlan(&planData)

	if data.Options == nil {
		t.Fatal("options unexpectedly nil after reconcile")
	}
	if !data.Options.ClasslessStaticRoutes.IsNull() {
		t.Errorf("classless_static_routes = %v, want null (plan had no routes)",
			data.Options.ClasslessStaticRoutes)
	}
}
