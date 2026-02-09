package dns_server

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// buildServerSelectList constructs a types.List of server_select entries for test setup.
func buildServerSelectList(t *testing.T, entries []struct {
	priority     int64
	queryPattern string
	recordType   string
}) types.List {
	t.Helper()
	var diags diag.Diagnostics
	values := make([]attr.Value, len(entries))
	for i, e := range entries {
		serverList, d := types.ListValue(types.ObjectType{AttrTypes: DNSServerEntryAttrTypes()}, []attr.Value{})
		diags.Append(d...)
		obj, d := types.ObjectValue(
			DNSServerSelectAttrTypes(),
			map[string]attr.Value{
				"priority":        types.Int64Value(e.priority),
				"server":          serverList,
				"record_type":     fwhelpers.StringValueOrNull(e.recordType),
				"query_pattern":   types.StringValue(e.queryPattern),
				"original_sender": types.StringNull(),
				"restrict_pp":     types.Int64Value(0),
			},
		)
		diags.Append(d...)
		values[i] = obj
	}
	if diags.HasError() {
		t.Fatalf("failed to build server select list: %v", diags.Errors())
	}
	listVal, d := types.ListValue(types.ObjectType{AttrTypes: DNSServerSelectAttrTypes()}, values)
	diags.Append(d...)
	if diags.HasError() {
		t.Fatalf("failed to build server select list: %v", diags.Errors())
	}
	return listVal
}

func TestFromClient_PreservesStateOrdering(t *testing.T) {
	ctx := context.Background()

	// Previous state has entries in order: B(priority=10), A(priority=5)
	prevState := buildServerSelectList(t, []struct {
		priority     int64
		queryPattern string
		recordType   string
	}{
		{priority: 10, queryPattern: "*.example.com", recordType: "a"},
		{priority: 5, queryPattern: "*.test.com", recordType: "a"},
	})

	model := &DNSServerModel{
		ServerSelect: prevState,
	}

	// Router returns entries with updated priorities but same content
	routerConfig := &client.DNSConfig{
		ServerSelect: []client.DNSServerSelect{
			{ID: 1, QueryPattern: "*.test.com", RecordType: "a", Servers: []client.DNSServer{{Address: "8.8.8.8"}}},
			{ID: 6, QueryPattern: "*.example.com", RecordType: "a", Servers: []client.DNSServer{{Address: "1.1.1.1"}}},
		},
	}

	var diags diag.Diagnostics
	model.FromClient(ctx, routerConfig, &diags)
	if diags.HasError() {
		t.Fatalf("FromClient returned errors: %v", diags.Errors())
	}

	// Extract result entries
	var resultSelects []DNSServerSelectModel
	d := model.ServerSelect.ElementsAs(ctx, &resultSelects, false)
	if d.HasError() {
		t.Fatalf("failed to extract result: %v", d.Errors())
	}

	if len(resultSelects) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resultSelects))
	}

	// Order should match previous state: example.com first, test.com second
	if resultSelects[0].QueryPattern.ValueString() != "*.example.com" {
		t.Errorf("expected first entry to be *.example.com, got %s", resultSelects[0].QueryPattern.ValueString())
	}
	if resultSelects[1].QueryPattern.ValueString() != "*.test.com" {
		t.Errorf("expected second entry to be *.test.com, got %s", resultSelects[1].QueryPattern.ValueString())
	}

	// Values should be from the router (latest)
	if resultSelects[0].Priority.ValueInt64() != 6 {
		t.Errorf("expected first entry priority 6, got %d", resultSelects[0].Priority.ValueInt64())
	}
	if resultSelects[1].Priority.ValueInt64() != 1 {
		t.Errorf("expected second entry priority 1, got %d", resultSelects[1].Priority.ValueInt64())
	}
}

func TestFromClient_HandlesExtraRouterEntries(t *testing.T) {
	ctx := context.Background()

	// Previous state has 2 entries
	prevState := buildServerSelectList(t, []struct {
		priority     int64
		queryPattern string
		recordType   string
	}{
		{priority: 1, queryPattern: "*.example.com", recordType: "a"},
		{priority: 2, queryPattern: "*.test.com", recordType: "a"},
	})

	model := &DNSServerModel{
		ServerSelect: prevState,
	}

	// Router returns 3 entries (one extra that wasn't in state)
	routerConfig := &client.DNSConfig{
		ServerSelect: []client.DNSServerSelect{
			{ID: 1, QueryPattern: "*.example.com", RecordType: "a", Servers: []client.DNSServer{}},
			{ID: 5, QueryPattern: "*.test.com", RecordType: "a", Servers: []client.DNSServer{}},
			{ID: 10, QueryPattern: "*.new.com", RecordType: "a", Servers: []client.DNSServer{}},
		},
	}

	var diags diag.Diagnostics
	model.FromClient(ctx, routerConfig, &diags)
	if diags.HasError() {
		t.Fatalf("FromClient returned errors: %v", diags.Errors())
	}

	var resultSelects []DNSServerSelectModel
	d := model.ServerSelect.ElementsAs(ctx, &resultSelects, false)
	if d.HasError() {
		t.Fatalf("failed to extract result: %v", d.Errors())
	}

	if len(resultSelects) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(resultSelects))
	}

	// First two should match previous state order
	if resultSelects[0].QueryPattern.ValueString() != "*.example.com" {
		t.Errorf("expected first entry *.example.com, got %s", resultSelects[0].QueryPattern.ValueString())
	}
	if resultSelects[1].QueryPattern.ValueString() != "*.test.com" {
		t.Errorf("expected second entry *.test.com, got %s", resultSelects[1].QueryPattern.ValueString())
	}
	// Extra entry should be appended at the end
	if resultSelects[2].QueryPattern.ValueString() != "*.new.com" {
		t.Errorf("expected third entry *.new.com, got %s", resultSelects[2].QueryPattern.ValueString())
	}
}

func TestFromClient_EmptyPreviousState(t *testing.T) {
	ctx := context.Background()

	// No previous state (null list)
	model := &DNSServerModel{
		ServerSelect: types.ListNull(types.ObjectType{AttrTypes: DNSServerSelectAttrTypes()}),
	}

	// Router returns entries in non-sorted order
	routerConfig := &client.DNSConfig{
		ServerSelect: []client.DNSServerSelect{
			{ID: 10, QueryPattern: "*.example.com", RecordType: "a", Servers: []client.DNSServer{}},
			{ID: 3, QueryPattern: "*.test.com", RecordType: "a", Servers: []client.DNSServer{}},
			{ID: 7, QueryPattern: "*.foo.com", RecordType: "a", Servers: []client.DNSServer{}},
		},
	}

	var diags diag.Diagnostics
	model.FromClient(ctx, routerConfig, &diags)
	if diags.HasError() {
		t.Fatalf("FromClient returned errors: %v", diags.Errors())
	}

	var resultSelects []DNSServerSelectModel
	d := model.ServerSelect.ElementsAs(ctx, &resultSelects, false)
	if d.HasError() {
		t.Fatalf("failed to extract result: %v", d.Errors())
	}

	if len(resultSelects) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(resultSelects))
	}

	// Should be sorted by ID when no previous state
	if resultSelects[0].Priority.ValueInt64() != 3 {
		t.Errorf("expected first entry priority 3, got %d", resultSelects[0].Priority.ValueInt64())
	}
	if resultSelects[1].Priority.ValueInt64() != 7 {
		t.Errorf("expected second entry priority 7, got %d", resultSelects[1].Priority.ValueInt64())
	}
	if resultSelects[2].Priority.ValueInt64() != 10 {
		t.Errorf("expected third entry priority 10, got %d", resultSelects[2].Priority.ValueInt64())
	}
}

func TestFromClient_DeletedStateEntries(t *testing.T) {
	ctx := context.Background()

	// Previous state has 3 entries
	prevState := buildServerSelectList(t, []struct {
		priority     int64
		queryPattern string
		recordType   string
	}{
		{priority: 1, queryPattern: "*.example.com", recordType: "a"},
		{priority: 2, queryPattern: "*.deleted.com", recordType: "a"},
		{priority: 3, queryPattern: "*.test.com", recordType: "a"},
	})

	model := &DNSServerModel{
		ServerSelect: prevState,
	}

	// Router returns only 2 entries (middle one was deleted)
	routerConfig := &client.DNSConfig{
		ServerSelect: []client.DNSServerSelect{
			{ID: 1, QueryPattern: "*.example.com", RecordType: "a", Servers: []client.DNSServer{}},
			{ID: 3, QueryPattern: "*.test.com", RecordType: "a", Servers: []client.DNSServer{}},
		},
	}

	var diags diag.Diagnostics
	model.FromClient(ctx, routerConfig, &diags)
	if diags.HasError() {
		t.Fatalf("FromClient returned errors: %v", diags.Errors())
	}

	var resultSelects []DNSServerSelectModel
	d := model.ServerSelect.ElementsAs(ctx, &resultSelects, false)
	if d.HasError() {
		t.Fatalf("failed to extract result: %v", d.Errors())
	}

	if len(resultSelects) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resultSelects))
	}

	// Order should match previous state, skipping deleted entry
	if resultSelects[0].QueryPattern.ValueString() != "*.example.com" {
		t.Errorf("expected first entry *.example.com, got %s", resultSelects[0].QueryPattern.ValueString())
	}
	if resultSelects[1].QueryPattern.ValueString() != "*.test.com" {
		t.Errorf("expected second entry *.test.com, got %s", resultSelects[1].QueryPattern.ValueString())
	}
}

func TestNormalizeRecordType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "a"},
		{"a", "a"},
		{"A", "a"},
		{"aaaa", "aaaa"},
		{"AAAA", "aaaa"},
		{" a ", "a"},
	}

	for _, tt := range tests {
		result := normalizeRecordType(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeRecordType(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
