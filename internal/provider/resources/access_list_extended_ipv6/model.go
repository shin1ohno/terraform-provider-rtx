package access_list_extended_ipv6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// AccessListExtendedIPv6Model describes the resource data model.
type AccessListExtendedIPv6Model struct {
	Name    types.String `tfsdk:"name"`
	Entries types.List   `tfsdk:"entry"`
}

// EntryModel describes the entry nested block data model.
type EntryModel struct {
	Sequence                types.Int64  `tfsdk:"sequence"`
	AceRuleAction           types.String `tfsdk:"ace_rule_action"`
	AceRuleProtocol         types.String `tfsdk:"ace_rule_protocol"`
	SourceAny               types.Bool   `tfsdk:"source_any"`
	SourcePrefix            types.String `tfsdk:"source_prefix"`
	SourcePrefixLength      types.Int64  `tfsdk:"source_prefix_length"`
	SourcePortEqual         types.String `tfsdk:"source_port_equal"`
	SourcePortRange         types.String `tfsdk:"source_port_range"`
	DestinationAny          types.Bool   `tfsdk:"destination_any"`
	DestinationPrefix       types.String `tfsdk:"destination_prefix"`
	DestinationPrefixLength types.Int64  `tfsdk:"destination_prefix_length"`
	DestinationPortEqual    types.String `tfsdk:"destination_port_equal"`
	DestinationPortRange    types.String `tfsdk:"destination_port_range"`
	Established             types.Bool   `tfsdk:"established"`
	Log                     types.Bool   `tfsdk:"log"`
}

// EntryAttrTypes returns the attribute types for the entry nested block.
func EntryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"sequence":                  types.Int64Type,
		"ace_rule_action":           types.StringType,
		"ace_rule_protocol":         types.StringType,
		"source_any":                types.BoolType,
		"source_prefix":             types.StringType,
		"source_prefix_length":      types.Int64Type,
		"source_port_equal":         types.StringType,
		"source_port_range":         types.StringType,
		"destination_any":           types.BoolType,
		"destination_prefix":        types.StringType,
		"destination_prefix_length": types.Int64Type,
		"destination_port_equal":    types.StringType,
		"destination_port_range":    types.StringType,
		"established":               types.BoolType,
		"log":                       types.BoolType,
	}
}

// ToClient converts the Terraform model to a client.AccessListExtendedIPv6.
func (m *AccessListExtendedIPv6Model) ToClient() client.AccessListExtendedIPv6 {
	acl := client.AccessListExtendedIPv6{
		Name: fwhelpers.GetStringValue(m.Name),
	}

	// Convert entries
	if !m.Entries.IsNull() && !m.Entries.IsUnknown() {
		var entries []EntryModel
		m.Entries.ElementsAs(context.Background(), &entries, false)
		acl.Entries = make([]client.AccessListExtendedIPv6Entry, len(entries))
		for i, e := range entries {
			acl.Entries[i] = client.AccessListExtendedIPv6Entry{
				Sequence:                int(e.Sequence.ValueInt64()),
				AceRuleAction:           fwhelpers.GetStringValue(e.AceRuleAction),
				AceRuleProtocol:         fwhelpers.GetStringValue(e.AceRuleProtocol),
				SourceAny:               fwhelpers.GetBoolValue(e.SourceAny),
				SourcePrefix:            fwhelpers.GetStringValue(e.SourcePrefix),
				SourcePrefixLength:      fwhelpers.GetInt64Value(e.SourcePrefixLength),
				SourcePortEqual:         fwhelpers.GetStringValue(e.SourcePortEqual),
				SourcePortRange:         fwhelpers.GetStringValue(e.SourcePortRange),
				DestinationAny:          fwhelpers.GetBoolValue(e.DestinationAny),
				DestinationPrefix:       fwhelpers.GetStringValue(e.DestinationPrefix),
				DestinationPrefixLength: fwhelpers.GetInt64Value(e.DestinationPrefixLength),
				DestinationPortEqual:    fwhelpers.GetStringValue(e.DestinationPortEqual),
				DestinationPortRange:    fwhelpers.GetStringValue(e.DestinationPortRange),
				Established:             fwhelpers.GetBoolValue(e.Established),
				Log:                     fwhelpers.GetBoolValue(e.Log),
			}
		}
	}

	return acl
}

// FromClient updates the Terraform model from a client.AccessListExtendedIPv6.
func (m *AccessListExtendedIPv6Model) FromClient(acl *client.AccessListExtendedIPv6) {
	m.Name = types.StringValue(acl.Name)

	// Convert entries
	if len(acl.Entries) > 0 {
		entries := make([]attr.Value, len(acl.Entries))
		for i, e := range acl.Entries {
			entries[i] = types.ObjectValueMust(
				EntryAttrTypes(),
				map[string]attr.Value{
					"sequence":                  types.Int64Value(int64(e.Sequence)),
					"ace_rule_action":           types.StringValue(e.AceRuleAction),
					"ace_rule_protocol":         types.StringValue(e.AceRuleProtocol),
					"source_any":                types.BoolValue(e.SourceAny),
					"source_prefix":             fwhelpers.StringValueOrNull(e.SourcePrefix),
					"source_prefix_length":      fwhelpers.Int64ValueOrNull(e.SourcePrefixLength),
					"source_port_equal":         fwhelpers.StringValueOrNull(e.SourcePortEqual),
					"source_port_range":         fwhelpers.StringValueOrNull(e.SourcePortRange),
					"destination_any":           types.BoolValue(e.DestinationAny),
					"destination_prefix":        fwhelpers.StringValueOrNull(e.DestinationPrefix),
					"destination_prefix_length": fwhelpers.Int64ValueOrNull(e.DestinationPrefixLength),
					"destination_port_equal":    fwhelpers.StringValueOrNull(e.DestinationPortEqual),
					"destination_port_range":    fwhelpers.StringValueOrNull(e.DestinationPortRange),
					"established":               types.BoolValue(e.Established),
					"log":                       types.BoolValue(e.Log),
				},
			)
		}
		m.Entries = types.ListValueMust(types.ObjectType{AttrTypes: EntryAttrTypes()}, entries)
	} else {
		m.Entries = types.ListValueMust(types.ObjectType{AttrTypes: EntryAttrTypes()}, []attr.Value{})
	}
}
