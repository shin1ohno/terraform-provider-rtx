package access_list_ip_dynamic

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

func intp(v int) *int { return &v }

// lanOutbound mirrors the IPv4 sibling of the 2026-09-08 case: three form-1 rows
// with start 100 step 5, one of them with a timeout.
func lanOutbound() (*AccessListIPDynamicModel, *client.IPFilterDynamicConfig) {
	m := &AccessListIPDynamicModel{
		Name:          types.StringValue("lan-outbound-dynamic"),
		SequenceStart: types.Int64Value(100),
		SequenceStep:  types.Int64Value(5),
		Entries: []EntryModel{
			{Source: types.StringValue("*"), Destination: types.StringValue("*"), Protocol: types.StringValue("ftp"), Syslog: types.BoolValue(false), Timeout: types.Int64Null()},
			{Source: types.StringValue("*"), Destination: types.StringValue("*"), Protocol: types.StringValue("www"), Syslog: types.BoolValue(true), Timeout: types.Int64Null()},
			{Source: types.StringValue("192.168.1.0/24"), Destination: types.StringValue("*"), Protocol: types.StringValue("tcp"), Syslog: types.BoolValue(false), Timeout: types.Int64Value(600)},
		},
	}
	router := &client.IPFilterDynamicConfig{Entries: []client.IPFilterDynamicEntry{
		{Number: 100, Source: "*", Dest: "*", Protocol: "ftp"},
		{Number: 105, Source: "*", Dest: "*", Protocol: "www", Syslog: true},
		{Number: 110, Source: "192.168.1.0/24", Dest: "*", Protocol: "tcp", Timeout: intp(600)},
	}}
	return m, router
}

func TestEntryKeys_PlannedAndRouterAgreeOnIdenticalRows(t *testing.T) {
	m, router := lanOutbound()
	planned, onRouter := m.PlannedEntryKeys(), RouterEntryKeys(router)
	if len(planned) != 3 || len(onRouter) != 3 {
		t.Fatalf("planned %d rows, router %d rows, want 3 and 3", len(planned), len(onRouter))
	}
	for seq, want := range planned {
		if got := onRouter[seq]; got != want {
			t.Errorf("seq %d: router key %q != planned key %q", seq, got, want)
		}
	}
}

func TestSequenceConflicts_ShortStateOverIdenticalRouterRowsIsNotAConflict(t *testing.T) {
	m, router := lanOutbound()
	if got := fwhelpers.CheckSequenceContentConflicts(m.PlannedEntryKeys(), RouterEntryKeys(router), []int{100}); len(got) != 0 {
		t.Fatalf("conflicts = %v, want none", got)
	}
	if got := fwhelpers.CheckSequenceContentConflicts(m.PlannedEntryKeys(), RouterEntryKeys(router), nil); len(got) != 0 {
		t.Fatalf("create-path conflicts = %v, want none", got)
	}
}

// A different timeout is different content: the command the resource would write
// is not the command the router holds.
func TestSequenceConflicts_TimeoutIsPartOfTheContent(t *testing.T) {
	m, router := lanOutbound()
	router.Entries[2].Timeout = intp(300)
	got := fwhelpers.CheckSequenceContentConflicts(m.PlannedEntryKeys(), RouterEntryKeys(router), []int{100, 105})
	if len(got) != 1 || got[0] != 110 {
		t.Fatalf("conflicts = %v, want [110]", got)
	}
}

// A form-2 row (`ip filter dynamic N * * filter 200 in 201`) on the router at a
// planned sequence is somebody else's row even when source/dest happen to match
// and the parser left Protocol empty — the lists make it a different entry.
func TestSequenceConflicts_FormTwoRouterRowNeverMatchesAFormOnePlan(t *testing.T) {
	m, router := lanOutbound()
	router.Entries[0] = client.IPFilterDynamicEntry{Number: 100, Source: "*", Dest: "*", FilterList: []int{200}, InFilterList: []int{201}}
	got := fwhelpers.CheckSequenceContentConflicts(m.PlannedEntryKeys(), RouterEntryKeys(router), nil)
	if len(got) != 1 || got[0] != 100 {
		t.Fatalf("conflicts = %v, want [100]", got)
	}
}
