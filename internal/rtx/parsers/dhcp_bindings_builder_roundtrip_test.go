package parsers

import (
	"reflect"
	"testing"
)

// Every branch of BuildDHCPBindCommand must parse back to the struct it was
// built from: the resource writes with the builder and reads with the parser,
// so a branch the parser cannot read is a resource that never converges.
func TestDHCPBindingsRoundTrip_BuildParse_EveryBuilderBranch(t *testing.T) {
	parser := NewDHCPBindingsParser()

	cases := []struct {
		name    string
		binding DHCPBinding
		want    string
	}{
		{
			name:    "plain MAC (chaddr)",
			binding: DHCPBinding{ScopeID: 1, IPAddress: "192.168.1.10", MACAddress: "00:11:22:33:44:55"},
			want:    "dhcp scope bind 1 192.168.1.10 00:11:22:33:44:55",
		},
		{
			name:    "ethernet MAC (client identifier, type 01)",
			binding: DHCPBinding{ScopeID: 1, IPAddress: "192.168.1.92", MACAddress: "64:4b:f0:37:27:0a", UseClientIdentifier: true},
			want:    "dhcp scope bind 1 192.168.1.92 ethernet 64:4b:f0:37:27:0a",
		},
		{
			name:    "text client identifier",
			binding: DHCPBinding{ScopeID: 1, IPAddress: "192.168.1.30", TextIdentifier: "client01"},
			want:    "dhcp scope bind 1 192.168.1.30 text client01",
		},
		{
			name:    "OUI range",
			binding: DHCPBinding{ScopeID: 1, IPRangeStart: "192.168.1.200", IPRangeEnd: "192.168.1.210", OUI: "00:a0:de"},
			want:    "dhcp scope bind 1 192.168.1.200-192.168.1.210 00:a0:de:*",
		},
		{
			name:    "ipcp",
			binding: DHCPBinding{ScopeID: 1, IPAddress: "192.168.1.40", IPCP: true},
			want:    "dhcp scope bind 1 192.168.1.40 ipcp",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := BuildDHCPBindCommand(tc.binding)
			if cmd != tc.want {
				t.Fatalf("BuildDHCPBindCommand() = %q, want %q", cmd, tc.want)
			}

			got, err := parser.ParseBindings(cmd, tc.binding.ScopeID)
			if err != nil {
				t.Fatalf("ParseBindings(%q) error: %v", cmd, err)
			}
			if len(got) != 1 {
				t.Fatalf("ParseBindings(%q) = %d rows, want 1", cmd, len(got))
			}
			if !reflect.DeepEqual(got[0], tc.binding) {
				t.Errorf("round-trip changed the binding:\n  built from: %+v\n  parsed as:  %+v", tc.binding, got[0])
			}
		})
	}
}

// The client-id branch is the one exception, and the reason the resource
// rejects client_identifier at plan time: the builder renders
// `client-id <id>`, which none of the parser's patterns read, so a binding
// written this way is invisible to every later Read.
func TestDHCPBindingsRoundTrip_BuildParse_ClientIDBranchDoesNotParse(t *testing.T) {
	parser := NewDHCPBindingsParser()
	binding := DHCPBinding{ScopeID: 1, IPAddress: "192.168.1.50", ClientIdentifier: "01:00:11:22:33:44:55"}

	cmd := BuildDHCPBindCommand(binding)
	if want := "dhcp scope bind 1 192.168.1.50 client-id 01:00:11:22:33:44:55"; cmd != want {
		t.Fatalf("BuildDHCPBindCommand() = %q, want %q", cmd, want)
	}

	got, err := parser.ParseBindings(cmd, 1)
	if err != nil {
		t.Fatalf("ParseBindings(%q) error: %v", cmd, err)
	}
	if len(got) != 0 {
		t.Fatalf("ParseBindings(%q) = %+v; the client-id form parsing now means the plan-time reject in the dhcp_binding resource can be lifted", cmd, got)
	}
}
