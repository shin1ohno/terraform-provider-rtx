package dhcp_binding

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// findConflictingBinding returns the row in existing that a create of planned
// would collide with — one in the same scope that already holds the planned
// IP, or the planned MAC — or nil when the scope has no such row.
//
// `dhcp scope bind` overwrites silently, so without this check a create on
// top of a row Terraform knows under another address succeeds and leaves two
// instances claiming one line. An OUI range row arrives from ListBindings
// with neither IP nor MAC and never matches; IPCP and text rows keep their
// IP and match on it.
func findConflictingBinding(planned client.DHCPBinding, existing []client.DHCPBinding) *client.DHCPBinding {
	plannedMAC := ""
	if planned.MACAddress != "" {
		plannedMAC, _ = normalizeMACAddress(planned.MACAddress)
	}

	for i := range existing {
		row := &existing[i]
		if row.ScopeID != planned.ScopeID {
			continue
		}
		if planned.IPAddress != "" && row.IPAddress == planned.IPAddress {
			return row
		}
		if plannedMAC != "" && row.MACAddress != "" {
			if rowMAC, _ := normalizeMACAddress(row.MACAddress); rowMAC == plannedMAC {
				return row
			}
		}
	}
	return nil
}

// conflictDiagnostic names the row a create collided with and the two ways
// out. It cannot tell the two apart: a sibling instance being destroyed in
// the same apply for the same row (a for_each key rename) looks exactly like
// a row nobody in state owns.
func conflictDiagnostic(planned client.DHCPBinding, hit *client.DHCPBinding, diagnostics *diag.Diagnostics) {
	fwhelpers.AppendDiagError(diagnostics,
		"DHCP binding already exists on the router",
		fmt.Sprintf("Scope %d already binds %s to %s, and this resource is not the one Terraform holds it under. "+
			"If another rtx_dhcp_binding for the same row is being destroyed in this apply (a for_each key rename), "+
			"declare the rename with a `moved` block instead: the destroy and this create run concurrently, and "+
			"the destroy removes the row even when this create wrote it first. Otherwise adopt the row with "+
			"`terraform import <address> %d:%s`.",
			planned.ScopeID, hit.IPAddress, hit.MACAddress, planned.ScopeID, hit.MACAddress),
	)
}

// missingAfterCreate reports the router having no row for the binding right
// after `dhcp scope bind` returned success, and returns true when it did.
//
// read() nulls the id when the scope's `show config` has no row for the MAC.
// Straight after a create that means the write did not land — the RTX
// rejects a malformed command with a message none of the write path's error
// patterns match — or that it landed and was removed before this read-back,
// which is what a concurrent destroy of a sibling instance for the same row
// does. Naming it here beats writing a null id into state and letting the
// next refresh quietly drop the resource.
func missingAfterCreate(data *DHCPBindingModel, binding client.DHCPBinding, diagnostics *diag.Diagnostics) bool {
	if !data.ID.IsNull() {
		return false
	}

	fwhelpers.AppendDiagError(diagnostics,
		"DHCP binding not found after create",
		fmt.Sprintf("Scope %d has no row for %s after `%s` returned success (`%s` shows none). "+
			"Either the router rejected the command with a message the provider does not recognise — "+
			"run it on the console to see the error — or another operation removed the row between "+
			"the write and this read-back; re-run the apply.",
			binding.ScopeID, binding.IPAddress,
			parsers.BuildDHCPBindCommand(toParserBinding(binding)),
			parsers.BuildShowDHCPBindingsCommand(binding.ScopeID)),
	)
	return true
}

// toParserBinding is the client → parsers projection the write path uses, so
// the command quoted in a diagnostic is the one that was actually sent.
func toParserBinding(binding client.DHCPBinding) parsers.DHCPBinding {
	return parsers.DHCPBinding{
		ScopeID:             binding.ScopeID,
		IPAddress:           binding.IPAddress,
		MACAddress:          binding.MACAddress,
		ClientIdentifier:    binding.ClientIdentifier,
		UseClientIdentifier: binding.UseClientIdentifier,
	}
}
