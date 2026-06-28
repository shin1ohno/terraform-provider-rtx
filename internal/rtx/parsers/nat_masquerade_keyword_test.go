package parsers

import "testing"

// RTX renders well-known ports as service keywords in "show config"
// (80 -> www, 443 -> https). The static parser must resolve them back to numbers.
func TestParseNATMasqueradeStaticKeywordPort(t *testing.T) {
	raw := "nat descriptor type 1000 masquerade\n" +
		"nat descriptor masquerade static 1000 1 192.168.0.200 tcp www\n" +
		"nat descriptor masquerade static 1000 2 192.168.0.200 tcp https\n"

	nats, err := ParseNATMasqueradeConfig(raw)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(nats) != 1 {
		t.Fatalf("got %d descriptors, want 1", len(nats))
	}
	if len(nats[0].StaticEntries) != 2 {
		t.Fatalf("got %d static entries, want 2", len(nats[0].StaticEntries))
	}

	want := map[int]int{1: 80, 2: 443}
	for _, e := range nats[0].StaticEntries {
		if e.InsideLocalPort == nil || *e.InsideLocalPort != want[e.EntryNumber] {
			t.Errorf("entry %d InsideLocalPort = %v, want %d", e.EntryNumber, e.InsideLocalPort, want[e.EntryNumber])
		}
		if e.Protocol != "tcp" {
			t.Errorf("entry %d Protocol = %q, want tcp", e.EntryNumber, e.Protocol)
		}
	}
}
