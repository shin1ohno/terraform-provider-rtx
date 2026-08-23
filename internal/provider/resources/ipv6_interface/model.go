package ipv6_interface

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// IPv6InterfaceModel describes the resource data model.
type IPv6InterfaceModel struct {
	Interface     types.String       `tfsdk:"interface"`
	Address       []IPv6AddressModel `tfsdk:"address"`
	RTADV         *RTADVModel        `tfsdk:"rtadv"`
	DHCPv6Service types.String       `tfsdk:"dhcpv6_service"`
	MTU           types.Int64        `tfsdk:"mtu"`
}

// IPv6AddressModel describes an IPv6 address block.
type IPv6AddressModel struct {
	Address     types.String `tfsdk:"address"`
	PrefixRef   types.String `tfsdk:"prefix_ref"`
	InterfaceID types.String `tfsdk:"interface_id"`
}

// RTADVModel describes the Router Advertisement configuration.
type RTADVModel struct {
	Enabled   types.Bool  `tfsdk:"enabled"`
	PrefixIDs types.List  `tfsdk:"prefix_ids"`
	OFlag     types.Bool  `tfsdk:"o_flag"`
	MFlag     types.Bool  `tfsdk:"m_flag"`
	Lifetime  types.Int64 `tfsdk:"lifetime"`
}

// ToClient converts the Terraform model to a client.IPv6InterfaceConfig.
func (m *IPv6InterfaceModel) ToClient(ctx context.Context, diagnostics *diag.Diagnostics) client.IPv6InterfaceConfig {
	config := client.IPv6InterfaceConfig{
		Interface:     fwhelpers.GetStringValue(m.Interface),
		DHCPv6Service: fwhelpers.GetStringValue(m.DHCPv6Service),
		MTU:           fwhelpers.GetInt64Value(m.MTU),
	}

	// Handle address blocks
	if len(m.Address) > 0 {
		config.Addresses = make([]client.IPv6Address, len(m.Address))
		for i, addr := range m.Address {
			config.Addresses[i] = client.IPv6Address{
				Address:     fwhelpers.GetStringValue(addr.Address),
				PrefixRef:   fwhelpers.GetStringValue(addr.PrefixRef),
				InterfaceID: fwhelpers.GetStringValue(addr.InterfaceID),
			}
		}
	}

	// Handle rtadv block
	if m.RTADV != nil {
		config.RTADV = &client.RTADVConfig{
			Enabled:   fwhelpers.GetBoolValue(m.RTADV.Enabled),
			PrefixIDs: fwhelpers.ListToIntSlice(m.RTADV.PrefixIDs),
			OFlag:     fwhelpers.GetBoolValue(m.RTADV.OFlag),
			MFlag:     fwhelpers.GetBoolValue(m.RTADV.MFlag),
			Lifetime:  fwhelpers.GetInt64Value(m.RTADV.Lifetime),
		}
	}

	return config
}

// addressKey identifies an address block by content, so read-back entries can
// be matched to desired entries without depending on either side's ordering.
// ValueString() maps a null attribute to "", which is what we want here: an
// unset optional attribute and an empty one describe the same address.
type addressKey struct {
	address     string
	prefixRef   string
	interfaceID string
}

func makeAddressKey(addr IPv6AddressModel) addressKey {
	return addressKey{
		address:     addr.Address.ValueString(),
		prefixRef:   addr.PrefixRef.ValueString(),
		interfaceID: addr.InterfaceID.ValueString(),
	}
}

// reorderAddressesToMatchPlan reorders the read-back address list so that
// entries also present in desired appear in desired's order.
//
// The router does not echo `ipv6 <iface> address` lines in the order they were
// written, and address is a ListNestedBlock, so order is significant to
// Terraform: a read-back that merely permutes the configured addresses is
// reported as "Provider produced inconsistent result after apply". Nothing in
// the provider reorders (both the SSH and the SFTP read path append in source
// order), so the normalization has to happen here.
//
// desired is the plan on Create/Update and the prior state on Read. Passing the
// prior state on Read matters: without it a refresh would rewrite state in
// router order and every subsequent plan would show a spurious reordering diff.
//
// This only normalizes ordering. An address the router returned that desired
// does not mention is kept — it is real drift — and an address in desired that
// the router did not return is left out rather than fabricated, so genuine
// differences still surface.
//
// Same bug class and remedy as reorderServerSelectToMatchPlan in
// internal/provider/resources/dns_server/model.go.
func (m *IPv6InterfaceModel) reorderAddressesToMatchPlan(desired []IPv6AddressModel) {
	if len(desired) == 0 || len(m.Address) == 0 {
		return
	}

	// One index list per key so repeated identical addresses are consumed once each.
	byKey := make(map[addressKey][]int, len(m.Address))
	for i, addr := range m.Address {
		key := makeAddressKey(addr)
		byKey[key] = append(byKey[key], i)
	}

	reordered := make([]IPv6AddressModel, 0, len(m.Address))
	matched := make([]bool, len(m.Address))

	for _, want := range desired {
		key := makeAddressKey(want)
		indices := byKey[key]
		if len(indices) == 0 {
			continue
		}
		idx := indices[0]
		byKey[key] = indices[1:]
		matched[idx] = true
		reordered = append(reordered, m.Address[idx])
	}

	for i, addr := range m.Address {
		if !matched[i] {
			reordered = append(reordered, addr)
		}
	}

	m.Address = reordered
}

// FromClient updates the Terraform model from a client.IPv6InterfaceConfig.
func (m *IPv6InterfaceModel) FromClient(ctx context.Context, config *client.IPv6InterfaceConfig, diagnostics *diag.Diagnostics) {
	m.Interface = types.StringValue(config.Interface)
	m.DHCPv6Service = fwhelpers.StringValueOrNull(config.DHCPv6Service)
	m.MTU = fwhelpers.Int64ValueOrNull(config.MTU)

	// Convert Addresses
	if len(config.Addresses) > 0 {
		m.Address = make([]IPv6AddressModel, len(config.Addresses))
		for i, addr := range config.Addresses {
			m.Address[i] = IPv6AddressModel{
				Address:     fwhelpers.StringValueOrNull(addr.Address),
				PrefixRef:   fwhelpers.StringValueOrNull(addr.PrefixRef),
				InterfaceID: fwhelpers.StringValueOrNull(addr.InterfaceID),
			}
		}
	} else {
		m.Address = nil
	}

	// Convert RTADV
	// Preserve the rtadv block structure if it was configured, even if not enabled
	if config.RTADV != nil {
		if m.RTADV == nil {
			m.RTADV = &RTADVModel{}
		}
		m.RTADV.Enabled = types.BoolValue(config.RTADV.Enabled)
		// Only update PrefixIDs if the router returned any
		// Router may not return these consistently, so preserve existing if empty
		if len(config.RTADV.PrefixIDs) > 0 {
			m.RTADV.PrefixIDs = fwhelpers.IntSliceToList(config.RTADV.PrefixIDs)
		} else if m.RTADV.PrefixIDs.IsUnknown() || m.RTADV.PrefixIDs.IsNull() {
			// Unknown would leak into state as "known after apply"; a bare null
			// carries no element type, which the framework rejects on Set
			m.RTADV.PrefixIDs = types.ListNull(types.Int64Type)
		}
		// else: preserve existing m.RTADV.PrefixIDs (known value)
		m.RTADV.OFlag = types.BoolValue(config.RTADV.OFlag)
		m.RTADV.MFlag = types.BoolValue(config.RTADV.MFlag)
		// Only update Lifetime if router returned a non-zero value
		if config.RTADV.Lifetime != 0 {
			m.RTADV.Lifetime = types.Int64Value(int64(config.RTADV.Lifetime))
		} else if m.RTADV.Lifetime.IsUnknown() {
			// If existing value is unknown, set to null to avoid unknown after apply
			m.RTADV.Lifetime = types.Int64Null()
		}
		// else: preserve existing m.RTADV.Lifetime (known value)
	} else if m.RTADV != nil {
		// If rtadv was configured but not returned by router, preserve existing values
		// Only set enabled to false since router confirmed it's not enabled
		m.RTADV.Enabled = types.BoolValue(false)
		m.RTADV.OFlag = types.BoolValue(false)
		m.RTADV.MFlag = types.BoolValue(false)
		// Preserve PrefixIDs and Lifetime - don't clear them
		// But if they're unknown, set to null
		if m.RTADV.PrefixIDs.IsUnknown() || m.RTADV.PrefixIDs.IsNull() {
			m.RTADV.PrefixIDs = types.ListNull(types.Int64Type)
		}
		if m.RTADV.Lifetime.IsUnknown() {
			m.RTADV.Lifetime = types.Int64Null()
		}
	}
	// If both config.RTADV is nil and m.RTADV is nil, leave it as nil
}
