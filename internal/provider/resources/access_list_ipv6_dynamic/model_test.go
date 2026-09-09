package access_list_ipv6_dynamic

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

func modelWithEntries(protocols ...string) *AccessListIPv6DynamicModel {
	m := &AccessListIPv6DynamicModel{
		Name:          types.StringValue("wan-outbound-ipv6-dynamic"),
		SequenceStart: types.Int64Value(1),
		SequenceStep:  types.Int64Value(5),
	}
	for _, p := range protocols {
		m.Entries = append(m.Entries, EntryModel{
			Source:      types.StringValue("*"),
			Destination: types.StringValue("*"),
			Protocol:    types.StringValue(p),
			Syslog:      types.BoolValue(false),
		})
	}
	return m
}

func routerEntry(seq int, source, protocol string) client.AccessListIPv6DynamicEntry {
	return client.AccessListIPv6DynamicEntry{
		Sequence: seq, Source: source, Destination: "*", Protocol: protocol,
	}
}

// The read must pick up an attribute change made on the device. This is the case
// that was mis-reported as "the provider never refreshes these resources" on
// 2026-08-23 — the observation had used `terraform state show`, which does not
// refresh at all.
func TestFromClient_EntryAttributeChangeOnRouterIsPickedUp(t *testing.T) {
	m := modelWithEntries("domain", "tcp", "udp")
	acl := &client.AccessListIPv6Dynamic{
		Name: "wan-outbound-ipv6-dynamic",
		Entries: []client.AccessListIPv6DynamicEntry{
			routerEntry(1, "dhcp-prefix@lan2::/64", "domain"),
			routerEntry(6, "dhcp-prefix@lan2::/64", "tcp"),
			routerEntry(11, "dhcp-prefix@lan2::/64", "udp"),
		},
	}

	m.FromClient(acl, m.GetCurrentSequences())

	if len(m.Entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(m.Entries))
	}
	for i, e := range m.Entries {
		if got := e.Source.ValueString(); got != "dhcp-prefix@lan2::/64" {
			t.Errorf("entry %d: source = %q, want the router's value", i, got)
		}
	}
}

// DELIBERATE, DO NOT "FIX": an entry the router has at a sequence this resource
// does not own is not adopted into state.
//
// The RTX has no concept of an ACL name — every `ipv6 filter dynamic <n>` line
// lives in one global namespace, and GetAccessListIPv6Dynamic returns all of them
// for whatever name it is asked about. currentSeqs is therefore the only expression
// of which sequences a given resource owns. Drop the filter and two
// rtx_access_list_ipv6_dynamic resources each adopt the other's entries and neither
// plan ever converges.
//
// The practical consequence is that an entry added on the device by hand is
// invisible until config declares it — which is correct, because until then no
// resource owns it.
func TestFromClient_RouterEntryOutsideOwnedSequencesIsNotAdopted(t *testing.T) {
	m := modelWithEntries("domain", "tcp")
	acl := &client.AccessListIPv6Dynamic{
		Name: "wan-outbound-ipv6-dynamic",
		Entries: []client.AccessListIPv6DynamicEntry{
			routerEntry(1, "*", "domain"),
			routerEntry(6, "*", "tcp"),
			routerEntry(11, "*", "udp"), // added out of band, owned by nobody
		},
	}

	m.FromClient(acl, m.GetCurrentSequences())

	if len(m.Entries) != 2 {
		t.Fatalf("got %d entries, want 2 — sequence 11 is not owned by this resource", len(m.Entries))
	}
}

// Import is the one case where the resource legitimately adopts everything: there
// is no prior state to scope it with.
func TestFromClient_ImportAdoptsEveryRouterEntry(t *testing.T) {
	m := &AccessListIPv6DynamicModel{
		Name:          types.StringValue("wan-outbound-ipv6-dynamic"),
		SequenceStart: types.Int64Value(1),
		SequenceStep:  types.Int64Value(5),
	}
	acl := &client.AccessListIPv6Dynamic{
		Name: "wan-outbound-ipv6-dynamic",
		Entries: []client.AccessListIPv6DynamicEntry{
			routerEntry(1, "*", "domain"),
			routerEntry(6, "dhcp-prefix@lan2::/64", "tcp"),
			routerEntry(11, "ra-prefix@lan2::/64", "udp"),
		},
	}

	m.FromClient(acl, m.GetCurrentSequences())

	if len(m.Entries) != 3 {
		t.Fatalf("got %d entries, want 3 on import", len(m.Entries))
	}
	if got := m.Entries[2].Source.ValueString(); got != "ra-prefix@lan2::/64" {
		t.Errorf("import dropped the prefix reference: source = %q", got)
	}
}

// A sequence removed on the device must not linger in state.
func TestFromClient_OwnedSequenceMissingFromRouterIsDropped(t *testing.T) {
	m := modelWithEntries("domain", "tcp", "udp")
	acl := &client.AccessListIPv6Dynamic{
		Name: "wan-outbound-ipv6-dynamic",
		Entries: []client.AccessListIPv6DynamicEntry{
			routerEntry(1, "*", "domain"),
			routerEntry(11, "*", "udp"),
		},
	}

	m.FromClient(acl, m.GetCurrentSequences())

	if len(m.Entries) != 2 {
		t.Fatalf("got %d entries, want 2 — sequence 6 is gone from the router", len(m.Entries))
	}
}

// wanOutbound is rtx_access_list_ipv6_dynamic.wan_outbound as home-monitor declares
// it (rtx-hnd.tf, 11 entries, start 1 step 5) and as the RTX1210 rendered it on
// 2026-09-08. The fixture is shared by the conflict tests below.
func wanOutbound() (*AccessListIPv6DynamicModel, *client.IPv6FilterDynamicConfig) {
	rows := []struct {
		source, protocol string
	}{
		{"*", "ftp"}, {"dhcp-prefix@lan2::/64", "domain"}, {"*", "www"}, {"*", "smtp"},
		{"*", "pop3"}, {"*", "submission"}, {"dhcp-prefix@lan2::/64", "tcp"},
		{"dhcp-prefix@lan2::/64", "udp"}, {"ra-prefix@lan2::/64", "domain"},
		{"ra-prefix@lan2::/64", "tcp"}, {"ra-prefix@lan2::/64", "udp"},
	}
	m := &AccessListIPv6DynamicModel{
		Name:          types.StringValue("wan-outbound-ipv6-dynamic"),
		SequenceStart: types.Int64Value(1),
		SequenceStep:  types.Int64Value(5),
	}
	router := &client.IPv6FilterDynamicConfig{}
	for i, r := range rows {
		m.Entries = append(m.Entries, EntryModel{
			Source:      types.StringValue(r.source),
			Destination: types.StringValue("*"),
			Protocol:    types.StringValue(r.protocol),
			Syslog:      types.BoolValue(false),
		})
		router.Entries = append(router.Entries, client.IPv6FilterDynamicEntry{
			Number: 1 + 5*i, Source: r.source, Dest: "*", Protocol: r.protocol,
		})
	}
	return m, router
}

// A planned entry and the router's rendering of the same entry must produce the
// same key, or the content-aware conflict check degrades to the sequence-only one.
func TestEntryKeys_PlannedAndRouterAgreeOnIdenticalRows(t *testing.T) {
	m, router := wanOutbound()
	planned, onRouter := m.PlannedEntryKeys(), RouterEntryKeys(router)
	if len(planned) != 11 || len(onRouter) != 11 {
		t.Fatalf("planned %d rows, router %d rows, want 11 and 11", len(planned), len(onRouter))
	}
	for seq, want := range planned {
		if got := onRouter[seq]; got != want {
			t.Errorf("seq %d: router key %q != planned key %q", seq, got, want)
		}
	}
}

func TestEntryKeys_SyslogIsPartOfTheKey(t *testing.T) {
	if entryKey("*", "*", "www", true) == entryKey("*", "*", "www", false) {
		t.Fatal("syslog=on and syslog=off rendered the same key")
	}
}

// 2026-09-08: the state carried 8 of the 11 rows (the last three were added to
// config after the state was last refreshed), the router already held all 11 from
// an earlier apply. With CheckSequenceConflicts the update failed on 41 46 51 —
// the resource's own rows — on every apply until the state was repaired by hand.
func TestSequenceConflicts_ShortStateOverIdenticalRouterRowsIsNotAConflict(t *testing.T) {
	m, router := wanOutbound()
	owned := []int{1, 6, 11, 16, 21, 26, 31, 36}

	if got := fwhelpers.CheckSequenceContentConflicts(m.PlannedEntryKeys(), RouterEntryKeys(router), owned); len(got) != 0 {
		t.Fatalf("conflicts = %v, want none: rows 41 46 51 are on the router in exactly the planned form", got)
	}
	// Create after `terraform state rm` — nothing owned, everything identical.
	if got := fwhelpers.CheckSequenceContentConflicts(m.PlannedEntryKeys(), RouterEntryKeys(router), nil); len(got) != 0 {
		t.Fatalf("create-path conflicts = %v, want none", got)
	}
}

// The guard still has to stop an overwrite of somebody else's row: same sequence,
// different content, not in this resource's state.
func TestSequenceConflicts_UnownedRowWithDifferentContentIsStillAConflict(t *testing.T) {
	m, router := wanOutbound()
	router.Entries[8].Source = "*" // seq 41 on the router: `* * domain`, not `ra-prefix@lan2::/64 * domain`
	owned := []int{1, 6, 11, 16, 21, 26, 31, 36}

	got := fwhelpers.CheckSequenceContentConflicts(m.PlannedEntryKeys(), RouterEntryKeys(router), owned)
	if len(got) != 1 || got[0] != 41 {
		t.Fatalf("conflicts = %v, want [41]", got)
	}
}
