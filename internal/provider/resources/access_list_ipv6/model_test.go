package access_list_ipv6

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

func entry(action, src, dst, proto, sport, dport string) EntryModel {
	return EntryModel{
		Sequence:    types.Int64Null(),
		Action:      types.StringValue(action),
		Source:      types.StringValue(src),
		Destination: types.StringValue(dst),
		Protocol:    types.StringValue(proto),
		SourcePort:  types.StringValue(sport),
		DestPort:    types.StringValue(dport),
		Log:         types.BoolValue(false),
	}
}

// wanIn is the IPv6 static list of the 2026-09-08 shape: start 101000 step 1,
// three rows, and a state that may hold fewer of them than the router does.
func wanIn() (*AccessListIPv6Model, []client.IPFilter) {
	m := &AccessListIPv6Model{
		Name:          types.StringValue("wan-in"),
		SequenceStart: types.Int64Value(101000),
		SequenceStep:  types.Int64Value(1),
		Entry: []EntryModel{
			entry("pass", "*", "*", "icmp6", "*", "*"),
			entry("pass", "*", "*", "udp", "*", "546"),
			entry("reject", "*", "*", "*", "*", "*"),
		},
	}
	router := []client.IPFilter{
		{Number: 101000, Action: "pass", SourceAddress: "*", DestAddress: "*", Protocol: "icmp6", SourcePort: "*", DestPort: "*"},
		{Number: 101001, Action: "pass", SourceAddress: "*", DestAddress: "*", Protocol: "udp", SourcePort: "*", DestPort: "546"},
		{Number: 101002, Action: "reject", SourceAddress: "*", DestAddress: "*", Protocol: "*", SourcePort: "*", DestPort: "*"},
	}
	return m, router
}

func plannedKeys(t *testing.T, m *AccessListIPv6Model) map[int]string {
	t.Helper()
	var diags diag.Diagnostics
	keys := m.PlannedFilterKeys(context.Background(), &diags)
	if diags.HasError() {
		t.Fatalf("PlannedFilterKeys: %v", diags)
	}
	return keys
}

func TestFilterKeys_PlannedAndRouterAgreeOnIdenticalRows(t *testing.T) {
	m, router := wanIn()
	planned, onRouter := plannedKeys(t, m), RouterFilterKeys(router)
	if len(planned) != 3 || len(onRouter) != 3 {
		t.Fatalf("planned %d rows, router %d rows, want 3 and 3", len(planned), len(onRouter))
	}
	for seq, want := range planned {
		if got := onRouter[seq]; got != want {
			t.Errorf("seq %d: router key %q != planned key %q", seq, got, want)
		}
	}
}

func TestFilterKeys_EmptyProtocolAndPortsReadAsAny(t *testing.T) {
	m, router := wanIn()
	router[2].Protocol, router[2].SourcePort, router[2].DestPort = "", "", ""
	if got, want := RouterFilterKeys(router)[101002], plannedKeys(t, m)[101002]; got != want {
		t.Fatalf("router key %q != planned key %q", got, want)
	}
}

func TestSequenceConflicts_ShortStateOverIdenticalRouterRowsIsNotAConflict(t *testing.T) {
	m, router := wanIn()
	for _, owned := range [][]int{{101000}, nil} {
		if got := fwhelpers.CheckSequenceContentConflicts(plannedKeys(t, m), RouterFilterKeys(router), owned); len(got) != 0 {
			t.Fatalf("owned %v: conflicts = %v, want none", owned, got)
		}
	}
}

func TestSequenceConflicts_UnownedRowWithDifferentContentIsStillAConflict(t *testing.T) {
	m, router := wanIn()
	router[1].DestPort = "547" // somebody else's row at our sequence
	got := fwhelpers.CheckSequenceContentConflicts(plannedKeys(t, m), RouterFilterKeys(router), []int{101000})
	if want := []int{101001}; !reflect.DeepEqual(got, want) {
		t.Fatalf("conflicts = %v, want %v", got, want)
	}
}
