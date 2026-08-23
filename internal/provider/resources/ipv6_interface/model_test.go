package ipv6_interface

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

// prefixRefAddr builds the prefix_ref form of an address block, as written for
// a prefix the router learns (ra-prefix@ / dhcp-prefix@).
func prefixRefAddr(prefixRef, interfaceID string) IPv6AddressModel {
	return IPv6AddressModel{
		Address:     types.StringNull(),
		PrefixRef:   types.StringValue(prefixRef),
		InterfaceID: types.StringValue(interfaceID),
	}
}

// literalAddr builds the literal form of an address block.
func literalAddr(address string) IPv6AddressModel {
	return IPv6AddressModel{
		Address:     types.StringValue(address),
		PrefixRef:   types.StringNull(),
		InterfaceID: types.StringNull(),
	}
}

func addressSummary(addrs []IPv6AddressModel) []string {
	out := make([]string, len(addrs))
	for i, a := range addrs {
		switch {
		case !a.PrefixRef.IsNull():
			out[i] = a.PrefixRef.ValueString() + a.InterfaceID.ValueString()
		default:
			out[i] = a.Address.ValueString()
		}
	}
	return out
}

func assertAddressOrder(t *testing.T, got, want []IPv6AddressModel) {
	t.Helper()

	gotSummary := addressSummary(got)
	wantSummary := addressSummary(want)

	if len(gotSummary) != len(wantSummary) {
		t.Fatalf("address count = %d %v, want %d %v", len(gotSummary), gotSummary, len(wantSummary), wantSummary)
	}
	for i := range wantSummary {
		if gotSummary[i] != wantSummary[i] {
			t.Errorf("address order = %v, want %v", gotSummary, wantSummary)
			return
		}
	}
}

// The live failure: lan1 is configured with a dhcp-prefix block first and a
// static ULA literal second, and the router echoes them the other way round.
// Six "Provider produced inconsistent result after apply" errors resulted,
// because address is a ListNestedBlock and order is significant.
func TestReorderAddressesToMatchPlan_SwappedMixedForms(t *testing.T) {
	planned := []IPv6AddressModel{
		prefixRefAddr("dhcp-prefix@lan2", "::1/64"),
		literalAddr("fd97:b085:767d::1/64"),
	}

	// Read-back arrives in the opposite order
	m := &IPv6InterfaceModel{
		Address: []IPv6AddressModel{
			literalAddr("fd97:b085:767d::1/64"),
			prefixRefAddr("dhcp-prefix@lan2", "::1/64"),
		},
	}

	m.reorderAddressesToMatchPlan(planned)

	assertAddressOrder(t, m.Address, planned)
}

func TestReorderAddressesToMatchPlan_AlreadyInPlanOrderIsUnchanged(t *testing.T) {
	planned := []IPv6AddressModel{
		prefixRefAddr("dhcp-prefix@lan2", "::1/64"),
		literalAddr("fd97:b085:767d::1/64"),
	}

	m := &IPv6InterfaceModel{
		Address: []IPv6AddressModel{
			prefixRefAddr("dhcp-prefix@lan2", "::1/64"),
			literalAddr("fd97:b085:767d::1/64"),
		},
	}

	m.reorderAddressesToMatchPlan(planned)

	assertAddressOrder(t, m.Address, planned)
}

func TestReorderAddressesToMatchPlan_ThreeAddressesFullyReversed(t *testing.T) {
	planned := []IPv6AddressModel{
		prefixRefAddr("dhcp-prefix@lan2", "::1/64"),
		literalAddr("fd97:b085:767d::1/64"),
		literalAddr("2001:db8::1/64"),
	}

	m := &IPv6InterfaceModel{
		Address: []IPv6AddressModel{
			literalAddr("2001:db8::1/64"),
			literalAddr("fd97:b085:767d::1/64"),
			prefixRefAddr("dhcp-prefix@lan2", "::1/64"),
		},
	}

	m.reorderAddressesToMatchPlan(planned)

	assertAddressOrder(t, m.Address, planned)
}

// An address the router returned but the plan does not mention is real drift.
// It must survive the reorder, after the matched entries, so Terraform still
// reports it instead of the provider quietly dropping it.
func TestReorderAddressesToMatchPlan_UnplannedRouterAddressIsKeptLast(t *testing.T) {
	planned := []IPv6AddressModel{
		prefixRefAddr("dhcp-prefix@lan2", "::1/64"),
	}

	m := &IPv6InterfaceModel{
		Address: []IPv6AddressModel{
			literalAddr("2001:db8::99/64"), // not in the plan
			prefixRefAddr("dhcp-prefix@lan2", "::1/64"),
		},
	}

	m.reorderAddressesToMatchPlan(planned)

	assertAddressOrder(t, m.Address, []IPv6AddressModel{
		prefixRefAddr("dhcp-prefix@lan2", "::1/64"),
		literalAddr("2001:db8::99/64"),
	})
}

// A planned address the router did not return is genuine drift too. The
// reorder must not fabricate it, or an address that silently failed to apply
// would look like it succeeded.
func TestReorderAddressesToMatchPlan_MissingRouterAddressIsNotFabricated(t *testing.T) {
	planned := []IPv6AddressModel{
		prefixRefAddr("dhcp-prefix@lan2", "::1/64"),
		literalAddr("fd97:b085:767d::1/64"),
	}

	m := &IPv6InterfaceModel{
		Address: []IPv6AddressModel{
			literalAddr("fd97:b085:767d::1/64"),
		},
	}

	m.reorderAddressesToMatchPlan(planned)

	assertAddressOrder(t, m.Address, []IPv6AddressModel{
		literalAddr("fd97:b085:767d::1/64"),
	})
}

func TestReorderAddressesToMatchPlan_DuplicateAddressesEachConsumedOnce(t *testing.T) {
	planned := []IPv6AddressModel{
		literalAddr("2001:db8::1/64"),
		literalAddr("2001:db8::1/64"),
		prefixRefAddr("dhcp-prefix@lan2", "::1/64"),
	}

	m := &IPv6InterfaceModel{
		Address: []IPv6AddressModel{
			prefixRefAddr("dhcp-prefix@lan2", "::1/64"),
			literalAddr("2001:db8::1/64"),
			literalAddr("2001:db8::1/64"),
		},
	}

	m.reorderAddressesToMatchPlan(planned)

	assertAddressOrder(t, m.Address, planned)
}

func TestReorderAddressesToMatchPlan_EmptyInputsAreNoOps(t *testing.T) {
	// No desired ordering to apply
	m := &IPv6InterfaceModel{Address: []IPv6AddressModel{literalAddr("2001:db8::1/64")}}
	m.reorderAddressesToMatchPlan(nil)
	assertAddressOrder(t, m.Address, []IPv6AddressModel{literalAddr("2001:db8::1/64")})

	// Router returned nothing
	empty := &IPv6InterfaceModel{}
	empty.reorderAddressesToMatchPlan([]IPv6AddressModel{literalAddr("2001:db8::1/64")})
	if empty.Address != nil {
		t.Errorf("Address = %v, want nil", empty.Address)
	}
}

// The pipeline Create/Update actually run: the plan is captured, FromClient
// overwrites Address with the router's ordering, then the reorder restores it.
// Guards the call-site contract that the snapshot is taken BEFORE the read-back
// — a helper that works in isolation is useless if it is handed post-read data.
func TestFromClientThenReorder_RestoresPlanOrder(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	m := &IPv6InterfaceModel{
		Interface: types.StringValue("lan1"),
		Address: []IPv6AddressModel{
			prefixRefAddr("dhcp-prefix@lan2", "::1/64"),
			literalAddr("fd97:b085:767d::1/64"),
		},
	}

	// What Create/Update capture before reading back
	plannedAddresses := m.Address

	// The router echoes the two addresses in the opposite order
	m.FromClient(ctx, &client.IPv6InterfaceConfig{
		Interface: "lan1",
		Addresses: []client.IPv6Address{
			{Address: "fd97:b085:767d::1/64"},
			{PrefixRef: "dhcp-prefix@lan2", InterfaceID: "::1/64"},
		},
	}, &diags)

	if diags.HasError() {
		t.Fatalf("FromClient produced errors: %v", diags.Errors())
	}

	// Without the reorder this is the state that trips Terraform
	assertAddressOrder(t, m.Address, []IPv6AddressModel{
		literalAddr("fd97:b085:767d::1/64"),
		prefixRefAddr("dhcp-prefix@lan2", "::1/64"),
	})

	m.reorderAddressesToMatchPlan(plannedAddresses)

	assertAddressOrder(t, m.Address, []IPv6AddressModel{
		prefixRefAddr("dhcp-prefix@lan2", "::1/64"),
		literalAddr("fd97:b085:767d::1/64"),
	})
}

// A null optional attribute and an empty-string one describe the same address,
// so they must match: FromClient produces null via StringValueOrNull while a
// plan can carry "".
func TestReorderAddressesToMatchPlan_NullAndEmptyStringMatch(t *testing.T) {
	planned := []IPv6AddressModel{
		{
			Address:     types.StringValue(""),
			PrefixRef:   types.StringValue("dhcp-prefix@lan2"),
			InterfaceID: types.StringValue("::1/64"),
		},
		literalAddr("fd97:b085:767d::1/64"),
	}

	m := &IPv6InterfaceModel{
		Address: []IPv6AddressModel{
			literalAddr("fd97:b085:767d::1/64"),
			prefixRefAddr("dhcp-prefix@lan2", "::1/64"), // Address is null, not ""
		},
	}

	m.reorderAddressesToMatchPlan(planned)

	assertAddressOrder(t, m.Address, []IPv6AddressModel{
		prefixRefAddr("dhcp-prefix@lan2", "::1/64"),
		literalAddr("fd97:b085:767d::1/64"),
	})
}
