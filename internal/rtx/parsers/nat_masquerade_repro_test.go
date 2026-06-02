package parsers

import "testing"

// Reproduces the "block count changed from 6 to 5" bug observed on RTX hnd
// nat descriptor 1000 after adding the tailscale udp 41641 static entry.
// Uses the exact SFTP-rendered config lines from the live device.
func TestReproNat1000SixEntries(t *testing.T) {
	raw := "" +
		"ip lan2 nat descriptor 1000\n" +
		"nat descriptor type 1000 masquerade\n" +
		"nat descriptor address outer 1000 primary\n" +
		"nat descriptor address inner 1000 192.168.1.0-192.168.1.255\n" +
		"nat descriptor masquerade incoming 1000 reject \n" +
		"nat descriptor masquerade static 1000 1 192.168.1.253 esp\n" +
		"nat descriptor masquerade static 1000 2 192.168.1.253 udp 500\n" +
		"nat descriptor masquerade static 1000 3 192.168.1.253 udp 4500\n" +
		"nat descriptor masquerade static 1000 4 192.168.1.253 udp 1701\n" +
		"nat descriptor masquerade static 1000 5 192.168.1.60 udp 41641\n" +
		"nat descriptor masquerade static 1000 900 192.168.1.20 tcp 55000\n"

	nats, err := ParseNATMasqueradeConfig(raw)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, n := range nats {
		if n.DescriptorID != 1000 {
			continue
		}
		found = true
		t.Logf("desc 1000 static entries: %d", len(n.StaticEntries))
		for _, e := range n.StaticEntries {
			port := -1
			if e.InsideLocalPort != nil {
				port = *e.InsideLocalPort
			}
			t.Logf("  entry=%d ip=%s proto=%s port=%d outer=%q", e.EntryNumber, e.InsideLocal, e.Protocol, port, e.OutsideGlobal)
		}
		if len(n.StaticEntries) != 6 {
			t.Errorf("expected 6 static entries, got %d", len(n.StaticEntries))
		}
	}
	if !found {
		t.Fatal("descriptor 1000 not found")
	}
}
