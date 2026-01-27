package nat_static

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// NATStaticModel describes the resource data model.
type NATStaticModel struct {
	DescriptorID types.Int64 `tfsdk:"descriptor_id"`
	Entry        types.List  `tfsdk:"entry"`
}

// NATStaticEntryModel describes a single static NAT entry.
type NATStaticEntryModel struct {
	InsideLocal       types.String `tfsdk:"inside_local"`
	InsideLocalPort   types.Int64  `tfsdk:"inside_local_port"`
	OutsideGlobal     types.String `tfsdk:"outside_global"`
	OutsideGlobalPort types.Int64  `tfsdk:"outside_global_port"`
	Protocol          types.String `tfsdk:"protocol"`
}

// EntryObjectType returns the object type for NAT static entries.
func EntryObjectType() map[string]attr.Type {
	return map[string]attr.Type{
		"inside_local":        types.StringType,
		"inside_local_port":   types.Int64Type,
		"outside_global":      types.StringType,
		"outside_global_port": types.Int64Type,
		"protocol":            types.StringType,
	}
}

// ToClient converts the Terraform model to a client.NATStatic.
func (m *NATStaticModel) ToClient() client.NATStatic {
	nat := client.NATStatic{
		DescriptorID: fwhelpers.GetInt64Value(m.DescriptorID),
		Entries:      make([]client.NATStaticEntry, 0),
	}

	if !m.Entry.IsNull() && !m.Entry.IsUnknown() {
		elements := m.Entry.Elements()
		for _, elem := range elements {
			objVal, ok := elem.(types.Object)
			if !ok {
				continue
			}

			attrs := objVal.Attributes()
			entry := client.NATStaticEntry{}

			if v, ok := attrs["inside_local"].(types.String); ok && !v.IsNull() {
				entry.InsideLocal = v.ValueString()
			}
			if v, ok := attrs["outside_global"].(types.String); ok && !v.IsNull() {
				entry.OutsideGlobal = v.ValueString()
			}
			if v, ok := attrs["protocol"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
				entry.Protocol = v.ValueString()
			}
			if v, ok := attrs["inside_local_port"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
				port := int(v.ValueInt64())
				if port > 0 {
					entry.InsideLocalPort = &port
				}
			}
			if v, ok := attrs["outside_global_port"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
				port := int(v.ValueInt64())
				if port > 0 {
					entry.OutsideGlobalPort = &port
				}
			}

			nat.Entries = append(nat.Entries, entry)
		}
	}

	return nat
}

// FromClient updates the Terraform model from a client.NATStatic.
func (m *NATStaticModel) FromClient(nat *client.NATStatic) {
	m.DescriptorID = types.Int64Value(int64(nat.DescriptorID))

	entries := make([]attr.Value, len(nat.Entries))
	for i, entry := range nat.Entries {
		insideLocalPort := types.Int64Null()
		if entry.InsideLocalPort != nil && *entry.InsideLocalPort > 0 {
			insideLocalPort = types.Int64Value(int64(*entry.InsideLocalPort))
		}

		outsideGlobalPort := types.Int64Null()
		if entry.OutsideGlobalPort != nil && *entry.OutsideGlobalPort > 0 {
			outsideGlobalPort = types.Int64Value(int64(*entry.OutsideGlobalPort))
		}

		protocol := types.StringNull()
		if entry.Protocol != "" {
			protocol = types.StringValue(entry.Protocol)
		}

		entryAttrs := map[string]attr.Value{
			"inside_local":        types.StringValue(entry.InsideLocal),
			"inside_local_port":   insideLocalPort,
			"outside_global":      types.StringValue(entry.OutsideGlobal),
			"outside_global_port": outsideGlobalPort,
			"protocol":            protocol,
		}

		entries[i] = types.ObjectValueMust(EntryObjectType(), entryAttrs)
	}

	m.Entry = types.ListValueMust(types.ObjectType{AttrTypes: EntryObjectType()}, entries)
}
