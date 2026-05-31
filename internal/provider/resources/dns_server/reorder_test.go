package dns_server

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestReorderServerSelectToMatchPlan_UnmatchedPlanEntryNoPanic reproduces the
// home-monitor rtx_dns_server.main crash (provider v0.14.2): auto mode
// (priority_start=1, priority_step=5) with 3 server_select blocks where the
// router read-back fails to match a plan entry. Before the fix,
// reorderServerSelectToMatchPlan pre-allocated reorderedValues with
// len == len(planSelects) and only filled matched slots, leaving a nil
// attr.Value that crashed types.ListValue -> basetypes.NewListValue
// (list_value.go:73) with a nil pointer dereference. After the fix, the
// unmatched slot is filled with the plan's own element, so (a) there is no nil
// slot (no panic) and (b) the result holds exactly the planned entries in plan
// order — which is what Terraform's post-apply consistency check requires
// because query_pattern is a Required attribute.
func TestReorderServerSelectToMatchPlan_UnmatchedPlanEntryNoPanic(t *testing.T) {
	ctx := context.Background()

	// Plan: 3 server_select blocks, auto mode (priorities 1, 6, 11).
	plan := &DNSServerModel{
		PriorityStart: types.Int64Value(1),
		PriorityStep:  types.Int64Value(5),
		ServerSelect: buildServerSelectList(t, []struct {
			priority     int64
			queryPattern string
			recordType   string
		}{
			{priority: 1, queryPattern: "home.local", recordType: "any"},
			{priority: 6, queryPattern: ".", recordType: "a"},
			{priority: 11, queryPattern: ".", recordType: "aaaa"},
		}),
	}

	// Router read-back: the "home.local" entry comes back with a renumbered
	// priority AND a record_type the matcher won't recognize, so plan entry 0
	// matches nothing (neither by priority nor by content). The "." entries
	// still match by priority.
	state := &DNSServerModel{
		PriorityStart: types.Int64Value(1),
		PriorityStep:  types.Int64Value(5),
		ServerSelect: buildServerSelectList(t, []struct {
			priority     int64
			queryPattern string
			recordType   string
		}{
			{priority: 2, queryPattern: "home.local", recordType: "ns"},
			{priority: 6, queryPattern: ".", recordType: "a"},
			{priority: 11, queryPattern: ".", recordType: "aaaa"},
		}),
	}

	var diags diag.Diagnostics
	// Must not panic (pre-fix this line crashed via NewListValue nil deref).
	state.reorderServerSelectToMatchPlan(ctx, plan, &diags)
	if diags.HasError() {
		t.Fatalf("reorderServerSelectToMatchPlan returned errors: %v", diags.Errors())
	}

	var result []DNSServerSelectModel
	d := state.ServerSelect.ElementsAs(ctx, &result, false)
	if d.HasError() {
		t.Fatalf("failed to extract result: %v", d.Errors())
	}

	// The result must hold exactly the planned entries in plan order: the
	// unmatched "home.local" slot is filled from the plan, the two "."
	// selectors carry their matched read-back values. A different order or
	// count would trip "Provider produced inconsistent result after apply".
	if len(result) != 3 {
		t.Fatalf("expected 3 entries in plan order, got %d", len(result))
	}
	if result[0].QueryPattern.ValueString() != "home.local" || result[0].RecordType.ValueString() != "any" {
		t.Errorf("expected first entry home.local/any (plan-filled), got %s/%s",
			result[0].QueryPattern.ValueString(), result[0].RecordType.ValueString())
	}
	if result[1].QueryPattern.ValueString() != "." || result[1].RecordType.ValueString() != "a" {
		t.Errorf("expected second entry ./a, got %s/%s",
			result[1].QueryPattern.ValueString(), result[1].RecordType.ValueString())
	}
	if result[2].QueryPattern.ValueString() != "." || result[2].RecordType.ValueString() != "aaaa" {
		t.Errorf("expected third entry ./aaaa, got %s/%s",
			result[2].QueryPattern.ValueString(), result[2].RecordType.ValueString())
	}
}

// TestReorderServerSelectToMatchPlan_NoMatchesAtAll covers the degenerate case
// where NO plan entry matches any router entry (e.g. a read-back that returns
// wholly different data). The result must be the plan's entries in plan order
// (the applied state must equal the plan; the divergent device entries surface
// as drift on the next Read), and must never contain a nil slot.
func TestReorderServerSelectToMatchPlan_NoMatchesAtAll(t *testing.T) {
	ctx := context.Background()

	plan := &DNSServerModel{
		PriorityStart: types.Int64Null(),
		PriorityStep:  types.Int64Null(),
		ServerSelect: buildServerSelectList(t, []struct {
			priority     int64
			queryPattern string
			recordType   string
		}{
			{priority: 100, queryPattern: "*.gone.com", recordType: "a"},
			{priority: 200, queryPattern: "*.missing.com", recordType: "a"},
		}),
	}

	state := &DNSServerModel{
		PriorityStart: types.Int64Null(),
		PriorityStep:  types.Int64Null(),
		ServerSelect: buildServerSelectList(t, []struct {
			priority     int64
			queryPattern string
			recordType   string
		}{
			{priority: 1, queryPattern: "*.actual.com", recordType: "a"},
			{priority: 2, queryPattern: "*.real.com", recordType: "a"},
		}),
	}

	var diags diag.Diagnostics
	state.reorderServerSelectToMatchPlan(ctx, plan, &diags)
	if diags.HasError() {
		t.Fatalf("reorderServerSelectToMatchPlan returned errors: %v", diags.Errors())
	}

	var result []DNSServerSelectModel
	d := state.ServerSelect.ElementsAs(ctx, &result, false)
	if d.HasError() {
		t.Fatalf("failed to extract result: %v", d.Errors())
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries (plan order), got %d", len(result))
	}
	if result[0].QueryPattern.ValueString() != "*.gone.com" {
		t.Errorf("expected first entry *.gone.com (plan-filled), got %s", result[0].QueryPattern.ValueString())
	}
	if result[1].QueryPattern.ValueString() != "*.missing.com" {
		t.Errorf("expected second entry *.missing.com (plan-filled), got %s", result[1].QueryPattern.ValueString())
	}
}

// TestReorderServerSelectToMatchPlan_ManualModeHappyPath confirms the common
// fully-matching manual-mode case still reorders to plan order after the fix.
func TestReorderServerSelectToMatchPlan_ManualModeHappyPath(t *testing.T) {
	ctx := context.Background()

	plan := &DNSServerModel{
		PriorityStart: types.Int64Null(),
		PriorityStep:  types.Int64Null(),
		ServerSelect: buildServerSelectList(t, []struct {
			priority     int64
			queryPattern string
			recordType   string
		}{
			{priority: 10, queryPattern: "*.example.com", recordType: "a"},
			{priority: 20, queryPattern: "*.test.com", recordType: "a"},
		}),
	}

	// Router returns the same content in reversed order.
	state := &DNSServerModel{
		PriorityStart: types.Int64Null(),
		PriorityStep:  types.Int64Null(),
		ServerSelect: buildServerSelectList(t, []struct {
			priority     int64
			queryPattern string
			recordType   string
		}{
			{priority: 20, queryPattern: "*.test.com", recordType: "a"},
			{priority: 10, queryPattern: "*.example.com", recordType: "a"},
		}),
	}

	var diags diag.Diagnostics
	state.reorderServerSelectToMatchPlan(ctx, plan, &diags)
	if diags.HasError() {
		t.Fatalf("reorderServerSelectToMatchPlan returned errors: %v", diags.Errors())
	}

	var result []DNSServerSelectModel
	d := state.ServerSelect.ElementsAs(ctx, &result, false)
	if d.HasError() {
		t.Fatalf("failed to extract result: %v", d.Errors())
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	// Must match plan order: example.com (priority 10) first.
	if result[0].Priority.ValueInt64() != 10 {
		t.Errorf("expected first entry priority 10, got %d", result[0].Priority.ValueInt64())
	}
	if result[1].Priority.ValueInt64() != 20 {
		t.Errorf("expected second entry priority 20, got %d", result[1].Priority.ValueInt64())
	}
}
