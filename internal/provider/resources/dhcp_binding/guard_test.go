package dhcp_binding

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func TestFindConflictingBinding(t *testing.T) {
	scope := []client.DHCPBinding{
		{ScopeID: 1, IPAddress: "192.168.1.25", MACAddress: "64:4b:f0:60:41:d8", UseClientIdentifier: true},
		{ScopeID: 1, IPAddress: "192.168.1.92", MACAddress: "64:4b:f0:37:27:0a", UseClientIdentifier: true},
		{ScopeID: 1, IPAddress: "192.168.1.60", MACAddress: "bc:24:11:aa:bb:cc"},
		// How ListBindings hands over an OUI range row (no IP, no MAC) and
		// an IPCP / text row (IP, no MAC).
		{ScopeID: 1},
		{ScopeID: 1, IPAddress: "192.168.1.201"},
		// Same IP and MAC in another scope.
		{ScopeID: 2, IPAddress: "192.168.1.92", MACAddress: "64:4b:f0:37:27:0a"},
	}

	tests := []struct {
		name     string
		planned  client.DHCPBinding
		existing []client.DHCPBinding
		wantIP   string // "" = no conflict
	}{
		{
			name:     "same IP and MAC — the rename race (SH1-6)",
			planned:  client.DHCPBinding{ScopeID: 1, IPAddress: "192.168.1.92", MACAddress: "64:4b:f0:37:27:0a", UseClientIdentifier: true},
			existing: scope,
			wantIP:   "192.168.1.92",
		},
		{
			name:     "same IP, different MAC",
			planned:  client.DHCPBinding{ScopeID: 1, IPAddress: "192.168.1.92", MACAddress: "00:11:22:33:44:55"},
			existing: scope,
			wantIP:   "192.168.1.92",
		},
		{
			name:     "same MAC, different IP, MAC in another notation",
			planned:  client.DHCPBinding{ScopeID: 1, IPAddress: "192.168.1.93", MACAddress: "64-4B-F0-37-27-0A"},
			existing: scope,
			wantIP:   "192.168.1.92",
		},
		{
			name:     "IP of a MAC-less row — a plain row cannot share it",
			planned:  client.DHCPBinding{ScopeID: 1, IPAddress: "192.168.1.201", MACAddress: "00:11:22:33:44:55"},
			existing: scope,
			wantIP:   "192.168.1.201",
		},
		{
			name:     "free IP and MAC",
			planned:  client.DHCPBinding{ScopeID: 1, IPAddress: "192.168.1.94", MACAddress: "00:11:22:33:44:55"},
			existing: scope,
			wantIP:   "",
		},
		{
			name:     "collision only in another scope",
			planned:  client.DHCPBinding{ScopeID: 3, IPAddress: "192.168.1.92", MACAddress: "64:4b:f0:37:27:0a"},
			existing: scope,
			wantIP:   "",
		},
		{
			name:     "empty scope",
			planned:  client.DHCPBinding{ScopeID: 1, IPAddress: "192.168.1.92", MACAddress: "64:4b:f0:37:27:0a"},
			existing: nil,
			wantIP:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findConflictingBinding(tt.planned, tt.existing)
			switch {
			case tt.wantIP == "" && got != nil:
				t.Fatalf("findConflictingBinding() = %+v, want nil", *got)
			case tt.wantIP != "" && got == nil:
				t.Fatalf("findConflictingBinding() = nil, want row for %s", tt.wantIP)
			case tt.wantIP != "" && got.IPAddress != tt.wantIP:
				t.Fatalf("findConflictingBinding() = row for %s, want %s", got.IPAddress, tt.wantIP)
			}
		})
	}
}

func TestConflictDiagnosticNamesTheRowAndBothWaysOut(t *testing.T) {
	var diags diag.Diagnostics
	planned := client.DHCPBinding{ScopeID: 1, IPAddress: "192.168.1.92", MACAddress: "64:4b:f0:37:27:0a"}
	hit := client.DHCPBinding{ScopeID: 1, IPAddress: "192.168.1.92", MACAddress: "64:4b:f0:37:27:0a"}

	conflictDiagnostic(planned, &hit, &diags)

	if !diags.HasError() || diags.ErrorsCount() != 1 {
		t.Fatalf("want exactly one error, got %v", diags)
	}
	detail := diags.Errors()[0].Detail()
	for _, want := range []string{
		"Scope 1 already binds 192.168.1.92 to 64:4b:f0:37:27:0a",
		"`moved` block",
		"terraform import <address> 1:64:4b:f0:37:27:0a",
	} {
		if !strings.Contains(detail, want) {
			t.Errorf("detail lacks %q:\n%s", want, detail)
		}
	}
}

// The D1 regression: a create whose read-back finds nothing must surface an
// error, not hand Terraform a null id that the next refresh silently drops.
func TestMissingAfterCreate(t *testing.T) {
	binding := client.DHCPBinding{ScopeID: 1, IPAddress: "192.168.1.92", MACAddress: "64:4b:f0:37:27:0a", UseClientIdentifier: true}

	t.Run("row present — no diagnostic", func(t *testing.T) {
		var diags diag.Diagnostics
		data := &DHCPBindingModel{ID: types.StringValue("1:64:4b:f0:37:27:0a")}

		if missingAfterCreate(data, binding, &diags) {
			t.Fatal("missingAfterCreate() = true for a present row")
		}
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
	})

	t.Run("row absent — error quoting the write and the show command", func(t *testing.T) {
		var diags diag.Diagnostics
		data := &DHCPBindingModel{ID: types.StringNull()}

		if !missingAfterCreate(data, binding, &diags) {
			t.Fatal("missingAfterCreate() = false for a null id")
		}
		if diags.ErrorsCount() != 1 {
			t.Fatalf("want exactly one error, got %v", diags)
		}
		detail := diags.Errors()[0].Detail()
		for _, want := range []string{
			"Scope 1 has no row for 192.168.1.92",
			"`dhcp scope bind 1 192.168.1.92 ethernet 64:4b:f0:37:27:0a` returned success",
			"`show config | grep \"dhcp scope bind 1\"` shows none",
		} {
			if !strings.Contains(detail, want) {
				t.Errorf("detail lacks %q:\n%s", want, detail)
			}
		}
	})
}
