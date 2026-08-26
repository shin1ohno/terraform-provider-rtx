package access_list_ipv6_dynamic

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
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
