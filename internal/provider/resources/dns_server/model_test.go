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

// buildHostsList constructs a types.List of host entries for test setup.
func buildHostsList(t *testing.T, entries []struct {
	recordType string
	name       string
	address    string
	ttl        int64
}) types.List {
	t.Helper()
	var diags diag.Diagnostics
	values := make([]attr.Value, len(entries))
	for i, e := range entries {
		obj, d := types.ObjectValue(
			DNSHostAttrTypes(),
			map[string]attr.Value{
				"type":    types.StringValue(e.recordType),
				"name":    types.StringValue(e.name),
				"address": types.StringValue(e.address),
				"ttl":     types.Int64Value(e.ttl),
			},
		)
		diags.Append(d...)
		values[i] = obj
	}
	if diags.HasError() {
		t.Fatalf("failed to build hosts list: %v", diags.Errors())
	}
	listVal, d := types.ListValue(types.ObjectType{AttrTypes: DNSHostAttrTypes()}, values)
	diags.Append(d...)
	if diags.HasError() {
		t.Fatalf("failed to build hosts list: %v", diags.Errors())
	}
	return listVal
}

func TestFromClient_HostsPreservesStateOrdering(t *testing.T) {
	ctx := context.Background()

	// Previous state has entries in specific order
	prevHosts := buildHostsList(t, []struct {
		recordType string
		name       string
		address    string
		ttl        int64
	}{
		{recordType: "a", name: "pro.home.local", address: "192.168.1.20", ttl: 0},
		{recordType: "a", name: "pro.home.local", address: "192.168.1.21", ttl: 0},
		{recordType: "a", name: "hnd.home.local", address: "192.168.1.253", ttl: 0},
	})

	model := &DNSServerModel{
		Hosts:        prevHosts,
		ServerSelect: types.ListNull(types.ObjectType{AttrTypes: DNSServerSelectAttrTypes()}),
	}

	// Router returns entries in different order
	routerConfig := &client.DNSConfig{
		Hosts: []client.DNSHost{
			{Type: "a", Name: "hnd.home.local", Address: "192.168.1.253"},
			{Type: "a", Name: "pro.home.local", Address: "192.168.1.21"},
			{Type: "a", Name: "pro.home.local", Address: "192.168.1.20"},
		},
	}

	var diags diag.Diagnostics
	model.FromClient(ctx, routerConfig, &diags)
	if diags.HasError() {
		t.Fatalf("FromClient returned errors: %v", diags.Errors())
	}

	var resultHosts []DNSHostModel
	d := model.Hosts.ElementsAs(ctx, &resultHosts, false)
	if d.HasError() {
		t.Fatalf("failed to extract result: %v", d.Errors())
	}

	if len(resultHosts) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(resultHosts))
	}

	// Order should match previous state
	if resultHosts[0].Address.ValueString() != "192.168.1.20" {
		t.Errorf("expected first entry 192.168.1.20, got %s", resultHosts[0].Address.ValueString())
	}
	if resultHosts[1].Address.ValueString() != "192.168.1.21" {
		t.Errorf("expected second entry 192.168.1.21, got %s", resultHosts[1].Address.ValueString())
	}
	if resultHosts[2].Address.ValueString() != "192.168.1.253" {
		t.Errorf("expected third entry 192.168.1.253, got %s", resultHosts[2].Address.ValueString())
	}
}

func TestFromClient_HostsHandlesNewEntries(t *testing.T) {
	ctx := context.Background()

	// Previous state has 2 entries
	prevHosts := buildHostsList(t, []struct {
		recordType string
		name       string
		address    string
		ttl        int64
	}{
		{recordType: "a", name: "hnd.home.local", address: "192.168.1.253", ttl: 0},
		{recordType: "a", name: "pro.home.local", address: "192.168.1.20", ttl: 0},
	})

	model := &DNSServerModel{
		Hosts:        prevHosts,
		ServerSelect: types.ListNull(types.ObjectType{AttrTypes: DNSServerSelectAttrTypes()}),
	}

	// Router returns 3 entries (one new)
	routerConfig := &client.DNSConfig{
		Hosts: []client.DNSHost{
			{Type: "a", Name: "pro.home.local", Address: "192.168.1.20"},
			{Type: "a", Name: "itm.home.local", Address: "192.168.1.254"},
			{Type: "a", Name: "hnd.home.local", Address: "192.168.1.253"},
		},
	}

	var diags diag.Diagnostics
	model.FromClient(ctx, routerConfig, &diags)
	if diags.HasError() {
		t.Fatalf("FromClient returned errors: %v", diags.Errors())
	}

	var resultHosts []DNSHostModel
	d := model.Hosts.ElementsAs(ctx, &resultHosts, false)
	if d.HasError() {
		t.Fatalf("failed to extract result: %v", d.Errors())
	}

	if len(resultHosts) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(resultHosts))
	}

	// First two match previous state order
	if resultHosts[0].Name.ValueString() != "hnd.home.local" {
		t.Errorf("expected first entry hnd.home.local, got %s", resultHosts[0].Name.ValueString())
	}
	if resultHosts[1].Name.ValueString() != "pro.home.local" {
		t.Errorf("expected second entry pro.home.local, got %s", resultHosts[1].Name.ValueString())
	}
	// New entry appended at end
	if resultHosts[2].Name.ValueString() != "itm.home.local" {
		t.Errorf("expected third entry itm.home.local, got %s", resultHosts[2].Name.ValueString())
	}
}

func TestFromClient_HostsEmptyPreviousState(t *testing.T) {
	ctx := context.Background()

	// No previous state
	model := &DNSServerModel{
		Hosts:        types.ListNull(types.ObjectType{AttrTypes: DNSHostAttrTypes()}),
		ServerSelect: types.ListNull(types.ObjectType{AttrTypes: DNSServerSelectAttrTypes()}),
	}

	routerConfig := &client.DNSConfig{
		Hosts: []client.DNSHost{
			{Type: "a", Name: "pro.home.local", Address: "192.168.1.20"},
			{Type: "a", Name: "hnd.home.local", Address: "192.168.1.253"},
		},
	}

	var diags diag.Diagnostics
	model.FromClient(ctx, routerConfig, &diags)
	if diags.HasError() {
		t.Fatalf("FromClient returned errors: %v", diags.Errors())
	}

	var resultHosts []DNSHostModel
	d := model.Hosts.ElementsAs(ctx, &resultHosts, false)
	if d.HasError() {
		t.Fatalf("failed to extract result: %v", d.Errors())
	}

	if len(resultHosts) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resultHosts))
	}

	// Should be in router order (no previous state to match)
	if resultHosts[0].Name.ValueString() != "pro.home.local" {
		t.Errorf("expected first entry pro.home.local, got %s", resultHosts[0].Name.ValueString())
	}
	if resultHosts[1].Name.ValueString() != "hnd.home.local" {
		t.Errorf("expected second entry hnd.home.local, got %s", resultHosts[1].Name.ValueString())
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

func makePriorServerSelect(t *testing.T, mode string) types.List {
	t.Helper()
	objType := types.ObjectType{AttrTypes: DNSServerSelectAttrTypes()}
	switch mode {
	case "null":
		return types.ListNull(objType)
	case "empty":
		return types.ListValueMust(objType, []attr.Value{})
	case "populated":
		serverList := types.ListValueMust(types.ObjectType{AttrTypes: DNSServerEntryAttrTypes()}, []attr.Value{})
		return types.ListValueMust(objType, []attr.Value{
			types.ObjectValueMust(DNSServerSelectAttrTypes(), map[string]attr.Value{
				"priority":        types.Int64Value(10),
				"server":          serverList,
				"record_type":     types.StringValue("a"),
				"query_pattern":   types.StringValue("example.com"),
				"original_sender": types.StringNull(),
				"restrict_pp":     types.Int64Value(0),
			}),
		})
	}
	t.Fatalf("unknown mode: %s", mode)
	return types.List{}
}

func makePriorHosts(t *testing.T, mode string) types.List {
	t.Helper()
	objType := types.ObjectType{AttrTypes: DNSHostAttrTypes()}
	switch mode {
	case "null":
		return types.ListNull(objType)
	case "empty":
		return types.ListValueMust(objType, []attr.Value{})
	case "populated":
		return types.ListValueMust(objType, []attr.Value{
			types.ObjectValueMust(DNSHostAttrTypes(), map[string]attr.Value{
				"type":    types.StringValue("a"),
				"name":    types.StringValue("router.local"),
				"address": types.StringValue("192.0.2.1"),
				"ttl":     types.Int64Value(0),
			}),
		})
	}
	t.Fatalf("unknown mode: %s", mode)
	return types.List{}
}

func TestFromClient_ServerSelect_NullPreservation(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name     string
		prior    string
		entries  []client.DNSServerSelect
		wantNull bool
		wantSize int
	}{
		{"empty + prior null stays null", "null", nil, true, 0},
		{"empty + prior empty stays empty", "empty", nil, false, 0},
		{"empty + prior populated overwrites to empty", "populated", nil, false, 0},
		{"populated over prior null", "null", []client.DNSServerSelect{{ID: 10, RecordType: "a", QueryPattern: "example.com"}}, false, 1},
		{"populated over prior empty", "empty", []client.DNSServerSelect{{ID: 20, RecordType: "a", QueryPattern: "test.com"}}, false, 1},
		{"populated over prior populated", "populated", []client.DNSServerSelect{{ID: 30, RecordType: "a", QueryPattern: "other.com"}}, false, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var diags diag.Diagnostics
			m := &DNSServerModel{
				ServerSelect: makePriorServerSelect(t, tc.prior),
				Hosts:        types.ListNull(types.ObjectType{AttrTypes: DNSHostAttrTypes()}),
				NameServers:  types.ListNull(types.StringType),
			}
			m.FromClient(ctx, &client.DNSConfig{ServerSelect: tc.entries}, &diags)
			if diags.HasError() {
				t.Fatalf("FromClient returned errors: %v", diags.Errors())
			}
			if got := m.ServerSelect.IsNull(); got != tc.wantNull {
				t.Errorf("ServerSelect.IsNull() = %v, want %v", got, tc.wantNull)
			}
			if !tc.wantNull && len(m.ServerSelect.Elements()) != tc.wantSize {
				t.Errorf("len(ServerSelect.Elements()) = %d, want %d", len(m.ServerSelect.Elements()), tc.wantSize)
			}
		})
	}
}

func TestFromClient_Hosts_NullPreservation(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name     string
		prior    string
		entries  []client.DNSHost
		wantNull bool
		wantSize int
	}{
		{"empty + prior null stays null", "null", nil, true, 0},
		{"empty + prior empty stays empty", "empty", nil, false, 0},
		{"empty + prior populated overwrites to empty", "populated", nil, false, 0},
		{"populated over prior null", "null", []client.DNSHost{{Type: "a", Name: "h.local", Address: "192.0.2.1"}}, false, 1},
		{"populated over prior empty", "empty", []client.DNSHost{{Type: "a", Name: "h2.local", Address: "192.0.2.2"}}, false, 1},
		{"populated over prior populated", "populated", []client.DNSHost{{Type: "a", Name: "h3.local", Address: "192.0.2.3"}}, false, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var diags diag.Diagnostics
			m := &DNSServerModel{
				ServerSelect: types.ListNull(types.ObjectType{AttrTypes: DNSServerSelectAttrTypes()}),
				Hosts:        makePriorHosts(t, tc.prior),
				NameServers:  types.ListNull(types.StringType),
			}
			m.FromClient(ctx, &client.DNSConfig{Hosts: tc.entries}, &diags)
			if diags.HasError() {
				t.Fatalf("FromClient returned errors: %v", diags.Errors())
			}
			if got := m.Hosts.IsNull(); got != tc.wantNull {
				t.Errorf("Hosts.IsNull() = %v, want %v", got, tc.wantNull)
			}
			if !tc.wantNull && len(m.Hosts.Elements()) != tc.wantSize {
				t.Errorf("len(Hosts.Elements()) = %d, want %d", len(m.Hosts.Elements()), tc.wantSize)
			}
		})
	}
}
