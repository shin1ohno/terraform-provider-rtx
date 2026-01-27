package nat_masquerade

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// NATMasqueradeModel describes the resource data model.
type NATMasqueradeModel struct {
	ID           types.String `tfsdk:"id"`
	DescriptorID types.Int64  `tfsdk:"descriptor_id"`
	OuterAddress types.String `tfsdk:"outer_address"`
	InnerNetwork types.String `tfsdk:"inner_network"`
	StaticEntry  types.List   `tfsdk:"static_entry"`
}

// StaticEntryModel describes the static entry nested block model.
type StaticEntryModel struct {
	EntryNumber       types.Int64  `tfsdk:"entry_number"`
	InsideLocal       types.String `tfsdk:"inside_local"`
	InsideLocalPort   types.Int64  `tfsdk:"inside_local_port"`
	OutsideGlobal     types.String `tfsdk:"outside_global"`
	OutsideGlobalPort types.Int64  `tfsdk:"outside_global_port"`
	Protocol          types.String `tfsdk:"protocol"`
}

// StaticEntryAttrTypes returns the attribute types for StaticEntryModel.
func StaticEntryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"entry_number":        types.Int64Type,
		"inside_local":        types.StringType,
		"inside_local_port":   types.Int64Type,
		"outside_global":      types.StringType,
		"outside_global_port": types.Int64Type,
		"protocol":            types.StringType,
	}
}

// ToClient converts the Terraform model to a client.NATMasquerade.
func (m *NATMasqueradeModel) ToClient(ctx context.Context) (client.NATMasquerade, diag.Diagnostics) {
	var diags diag.Diagnostics

	nat := client.NATMasquerade{
		DescriptorID: fwhelpers.GetInt64Value(m.DescriptorID),
		OuterAddress: fwhelpers.GetStringValue(m.OuterAddress),
		InnerNetwork: fwhelpers.GetStringValue(m.InnerNetwork),
	}

	// Convert static entries
	if !m.StaticEntry.IsNull() && !m.StaticEntry.IsUnknown() {
		var entries []StaticEntryModel
		diags.Append(m.StaticEntry.ElementsAs(ctx, &entries, false)...)
		if diags.HasError() {
			return nat, diags
		}

		nat.StaticEntries = make([]client.MasqueradeStaticEntry, len(entries))
		for i, entry := range entries {
			nat.StaticEntries[i] = client.MasqueradeStaticEntry{
				EntryNumber:   fwhelpers.GetInt64Value(entry.EntryNumber),
				InsideLocal:   fwhelpers.GetStringValue(entry.InsideLocal),
				OutsideGlobal: fwhelpers.GetStringValue(entry.OutsideGlobal),
				Protocol:      fwhelpers.GetStringValue(entry.Protocol),
			}

			// Handle optional port fields
			if !entry.InsideLocalPort.IsNull() && !entry.InsideLocalPort.IsUnknown() {
				port := int(entry.InsideLocalPort.ValueInt64())
				nat.StaticEntries[i].InsideLocalPort = &port
			}

			if !entry.OutsideGlobalPort.IsNull() && !entry.OutsideGlobalPort.IsUnknown() {
				port := int(entry.OutsideGlobalPort.ValueInt64())
				nat.StaticEntries[i].OutsideGlobalPort = &port
			}
		}
	}

	return nat, diags
}

// FromClient updates the Terraform model from a client.NATMasquerade.
func (m *NATMasqueradeModel) FromClient(ctx context.Context, nat *client.NATMasquerade) diag.Diagnostics {
	var diags diag.Diagnostics

	m.DescriptorID = types.Int64Value(int64(nat.DescriptorID))
	m.OuterAddress = types.StringValue(nat.OuterAddress)
	m.InnerNetwork = fwhelpers.StringValueOrNull(nat.InnerNetwork)

	// Convert static entries
	if len(nat.StaticEntries) > 0 {
		entries := make([]attr.Value, len(nat.StaticEntries))
		for i, entry := range nat.StaticEntries {
			entryMap := map[string]attr.Value{
				"entry_number":        types.Int64Value(int64(entry.EntryNumber)),
				"inside_local":        types.StringValue(entry.InsideLocal),
				"inside_local_port":   types.Int64Null(),
				"outside_global":      types.StringValue(entry.OutsideGlobal),
				"outside_global_port": types.Int64Null(),
				"protocol":            fwhelpers.StringValueOrNull(entry.Protocol),
			}

			// Handle optional port fields
			if entry.InsideLocalPort != nil {
				entryMap["inside_local_port"] = types.Int64Value(int64(*entry.InsideLocalPort))
			}
			if entry.OutsideGlobalPort != nil {
				entryMap["outside_global_port"] = types.Int64Value(int64(*entry.OutsideGlobalPort))
			}

			objVal, objDiags := types.ObjectValue(StaticEntryAttrTypes(), entryMap)
			diags.Append(objDiags...)
			entries[i] = objVal
		}

		listVal, listDiags := types.ListValue(types.ObjectType{AttrTypes: StaticEntryAttrTypes()}, entries)
		diags.Append(listDiags...)
		m.StaticEntry = listVal
	} else {
		m.StaticEntry = types.ListNull(types.ObjectType{AttrTypes: StaticEntryAttrTypes()})
	}

	return diags
}
