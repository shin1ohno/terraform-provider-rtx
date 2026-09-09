package access_list_ip

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// entryList builds the entry list attribute the way a plan carries it, with the
// schema defaults ("*" for protocol and ports, false for the bools) already applied.
func entryList(entries ...EntryModel) types.List {
	values := make([]attr.Value, len(entries))
	for i, e := range entries {
		values[i] = entryToObjectValue(e)
	}
	return types.ListValueMust(types.ObjectType{AttrTypes: EntryModelAttrTypes()}, values)
}

func entry(action, src, dst, proto, sport, dport string, established bool) EntryModel {
	return EntryModel{
		Sequence:    types.Int64Null(),
		Action:      types.StringValue(action),
		Source:      types.StringValue(src),
		Destination: types.StringValue(dst),
		Protocol:    types.StringValue(proto),
		SourcePort:  types.StringValue(sport),
		DestPort:    types.StringValue(dport),
		Established: types.BoolValue(established),
		Log:         types.BoolValue(false),
	}
}

// lanIn is a static list of the shape that gets wedged: start 200000 step 1,
// three rows, and a state that may hold fewer of them than the router does.
func lanIn() (*AccessListIPModel, []client.IPFilter) {
	m := &AccessListIPModel{
		Name:          types.StringValue("lan-in"),
		SequenceStart: types.Int64Value(200000),
		SequenceStep:  types.Int64Value(1),
		Entry: entryList(
			entry("pass", "192.168.1.0/24", "*", "*", "*", "*", false),
			entry("pass", "*", "*", "tcp", "*", "22", false),
			entry("reject", "*", "*", "tcp", "*", "*", true),
		),
	}
	router := []client.IPFilter{
		{Number: 200000, Action: "pass", SourceAddress: "192.168.1.0/24", DestAddress: "*", Protocol: "*", SourcePort: "*", DestPort: "*"},
		{Number: 200001, Action: "pass", SourceAddress: "*", DestAddress: "*", Protocol: "tcp", SourcePort: "*", DestPort: "22"},
		{Number: 200002, Action: "reject", SourceAddress: "*", DestAddress: "*", Protocol: "tcp", SourcePort: "*", DestPort: "*", Established: true},
	}
	return m, router
}

func TestFilterKeys_PlannedAndRouterAgreeOnIdenticalRows(t *testing.T) {
	m, router := lanIn()
	planned, onRouter := m.PlannedFilterKeys(), RouterFilterKeys(router)
	if len(planned) != 3 || len(onRouter) != 3 {
		t.Fatalf("planned %d rows, router %d rows, want 3 and 3", len(planned), len(onRouter))
	}
	for seq, want := range planned {
		if got := onRouter[seq]; got != want {
			t.Errorf("seq %d: router key %q != planned key %q", seq, got, want)
		}
	}
}

// `show config` leaves protocol and ports empty when the row says `*`; the key
// normalises them the same way SetEntriesFromFilters does, so the empty form and
// the explicit form are the same content.
func TestFilterKeys_EmptyProtocolAndPortsReadAsAny(t *testing.T) {
	m, router := lanIn()
	router[0].Protocol, router[0].SourcePort, router[0].DestPort = "", "", ""
	if got, want := RouterFilterKeys(router)[200000], m.PlannedFilterKeys()[200000]; got != want {
		t.Fatalf("router key %q != planned key %q", got, want)
	}
}

func TestSequenceConflicts_ShortStateOverIdenticalRouterRowsIsNotAConflict(t *testing.T) {
	m, router := lanIn()
	for _, owned := range [][]int{{200000}, nil} {
		if got := fwhelpers.CheckSequenceContentConflicts(m.PlannedFilterKeys(), RouterFilterKeys(router), owned); len(got) != 0 {
			t.Fatalf("owned %v: conflicts = %v, want none", owned, got)
		}
	}
}

func TestSequenceConflicts_UnownedRowWithDifferentContentIsStillAConflict(t *testing.T) {
	m, router := lanIn()
	router[2].Established = false // somebody else's `reject * * tcp * *` at our sequence
	got := fwhelpers.CheckSequenceContentConflicts(m.PlannedFilterKeys(), RouterFilterKeys(router), []int{200000, 200001})
	if want := []int{200002}; !reflect.DeepEqual(got, want) {
		t.Fatalf("conflicts = %v, want %v", got, want)
	}
}
