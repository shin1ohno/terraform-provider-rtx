package access_list_extended

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// AccessListExtendedModel describes the resource data model.
type AccessListExtendedModel struct {
	Name          types.String `tfsdk:"name"`
	SequenceStart types.Int64  `tfsdk:"sequence_start"`
	SequenceStep  types.Int64  `tfsdk:"sequence_step"`
	Entry         types.List   `tfsdk:"entry"`
	Apply         types.List   `tfsdk:"apply"`
}

// EntryModel describes a single ACL entry.
type EntryModel struct {
	Sequence              types.Int64  `tfsdk:"sequence"`
	AceRuleAction         types.String `tfsdk:"ace_rule_action"`
	AceRuleProtocol       types.String `tfsdk:"ace_rule_protocol"`
	SourceAny             types.Bool   `tfsdk:"source_any"`
	SourcePrefix          types.String `tfsdk:"source_prefix"`
	SourcePrefixMask      types.String `tfsdk:"source_prefix_mask"`
	SourcePortEqual       types.String `tfsdk:"source_port_equal"`
	SourcePortRange       types.String `tfsdk:"source_port_range"`
	DestinationAny        types.Bool   `tfsdk:"destination_any"`
	DestinationPrefix     types.String `tfsdk:"destination_prefix"`
	DestinationPrefixMask types.String `tfsdk:"destination_prefix_mask"`
	DestinationPortEqual  types.String `tfsdk:"destination_port_equal"`
	DestinationPortRange  types.String `tfsdk:"destination_port_range"`
	Established           types.Bool   `tfsdk:"established"`
	Log                   types.Bool   `tfsdk:"log"`
}

// ApplyModel describes an interface binding.
type ApplyModel struct {
	Interface types.String `tfsdk:"interface"`
	Direction types.String `tfsdk:"direction"`
	Sequences types.List   `tfsdk:"sequences"`
}

// Constants for sequence calculation.
const (
	DefaultSequenceStep = 10
	// MaxSequenceValue is the maximum valid sequence number for RTX filters.
	// RTX routers support filter numbers up to 2147483647, but practical usage is typically under 1000000.
	MaxSequenceValue = 2147483647
)

// EntryAttrTypes returns the attribute types for EntryModel.
func EntryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"sequence":                types.Int64Type,
		"ace_rule_action":         types.StringType,
		"ace_rule_protocol":       types.StringType,
		"source_any":              types.BoolType,
		"source_prefix":           types.StringType,
		"source_prefix_mask":      types.StringType,
		"source_port_equal":       types.StringType,
		"source_port_range":       types.StringType,
		"destination_any":         types.BoolType,
		"destination_prefix":      types.StringType,
		"destination_prefix_mask": types.StringType,
		"destination_port_equal":  types.StringType,
		"destination_port_range":  types.StringType,
		"established":             types.BoolType,
		"log":                     types.BoolType,
	}
}

// ApplyAttrTypes returns the attribute types for ApplyModel.
func ApplyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"interface": types.StringType,
		"direction": types.StringType,
		"sequences": types.ListType{ElemType: types.Int64Type},
	}
}

// ToClient converts the Terraform model to a client.AccessListExtended.
func (m *AccessListExtendedModel) ToClient() client.AccessListExtended {
	acl := client.AccessListExtended{
		Name:    fwhelpers.GetStringValue(m.Name),
		Entries: m.entriesToClient(),
		Applies: m.appliesToClient(),
	}
	return acl
}

// entriesToClient converts entry models to client entries.
func (m *AccessListExtendedModel) entriesToClient() []client.AccessListExtendedEntry {
	if m.Entry.IsNull() || m.Entry.IsUnknown() {
		return []client.AccessListExtendedEntry{}
	}

	sequenceStart := fwhelpers.GetInt64Value(m.SequenceStart)
	sequenceStep := fwhelpers.GetInt64Value(m.SequenceStep)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	var entries []EntryModel
	m.Entry.ElementsAs(context.TODO(), &entries, false)

	result := make([]client.AccessListExtendedEntry, 0, len(entries))
	for i, entry := range entries {
		// Determine sequence number
		var sequence int
		if sequenceStart > 0 {
			// Auto mode: calculate sequence
			sequence = sequenceStart + (i * sequenceStep)
		} else {
			// Manual mode: use explicit sequence
			sequence = fwhelpers.GetInt64Value(entry.Sequence)
		}

		clientEntry := client.AccessListExtendedEntry{
			Sequence:              sequence,
			AceRuleAction:         fwhelpers.GetStringValue(entry.AceRuleAction),
			AceRuleProtocol:       fwhelpers.GetStringValue(entry.AceRuleProtocol),
			SourceAny:             fwhelpers.GetBoolValue(entry.SourceAny),
			SourcePrefix:          fwhelpers.GetStringValue(entry.SourcePrefix),
			SourcePrefixMask:      fwhelpers.GetStringValue(entry.SourcePrefixMask),
			SourcePortEqual:       fwhelpers.GetStringValue(entry.SourcePortEqual),
			SourcePortRange:       fwhelpers.GetStringValue(entry.SourcePortRange),
			DestinationAny:        fwhelpers.GetBoolValue(entry.DestinationAny),
			DestinationPrefix:     fwhelpers.GetStringValue(entry.DestinationPrefix),
			DestinationPrefixMask: fwhelpers.GetStringValue(entry.DestinationPrefixMask),
			DestinationPortEqual:  fwhelpers.GetStringValue(entry.DestinationPortEqual),
			DestinationPortRange:  fwhelpers.GetStringValue(entry.DestinationPortRange),
			Established:           fwhelpers.GetBoolValue(entry.Established),
			Log:                   fwhelpers.GetBoolValue(entry.Log),
		}
		result = append(result, clientEntry)
	}

	return result
}

// appliesToClient converts apply models to client applies.
func (m *AccessListExtendedModel) appliesToClient() []client.ExtendedApply {
	if m.Apply.IsNull() || m.Apply.IsUnknown() {
		return []client.ExtendedApply{}
	}

	var applies []ApplyModel
	m.Apply.ElementsAs(context.TODO(), &applies, false)

	result := make([]client.ExtendedApply, 0, len(applies))
	for _, apply := range applies {
		clientApply := client.ExtendedApply{
			Interface: fwhelpers.GetStringValue(apply.Interface),
			Direction: fwhelpers.GetStringValue(apply.Direction),
		}

		// Extract sequences
		if !apply.Sequences.IsNull() && !apply.Sequences.IsUnknown() {
			var sequences []types.Int64
			apply.Sequences.ElementsAs(context.TODO(), &sequences, false)
			for _, id := range sequences {
				clientApply.FilterIDs = append(clientApply.FilterIDs, int(id.ValueInt64()))
			}
		}

		result = append(result, clientApply)
	}

	return result
}

// FromClient updates the Terraform model from a client.AccessListExtended.
func (m *AccessListExtendedModel) FromClient(acl *client.AccessListExtended) {
	m.Name = types.StringValue(acl.Name)
	// Note: SequenceStart and SequenceStep are config-only, not stored on router

	// Convert entries
	entryValues := make([]attr.Value, len(acl.Entries))
	for i, entry := range acl.Entries {
		entryValues[i] = types.ObjectValueMust(EntryAttrTypes(), map[string]attr.Value{
			"sequence":                types.Int64Value(int64(entry.Sequence)),
			"ace_rule_action":         types.StringValue(entry.AceRuleAction),
			"ace_rule_protocol":       types.StringValue(entry.AceRuleProtocol),
			"source_any":              types.BoolValue(entry.SourceAny),
			"source_prefix":           fwhelpers.StringValueOrNull(entry.SourcePrefix),
			"source_prefix_mask":      fwhelpers.StringValueOrNull(entry.SourcePrefixMask),
			"source_port_equal":       fwhelpers.StringValueOrNull(entry.SourcePortEqual),
			"source_port_range":       fwhelpers.StringValueOrNull(entry.SourcePortRange),
			"destination_any":         types.BoolValue(entry.DestinationAny),
			"destination_prefix":      fwhelpers.StringValueOrNull(entry.DestinationPrefix),
			"destination_prefix_mask": fwhelpers.StringValueOrNull(entry.DestinationPrefixMask),
			"destination_port_equal":  fwhelpers.StringValueOrNull(entry.DestinationPortEqual),
			"destination_port_range":  fwhelpers.StringValueOrNull(entry.DestinationPortRange),
			"established":             types.BoolValue(entry.Established),
			"log":                     types.BoolValue(entry.Log),
		})
	}
	m.Entry = types.ListValueMust(types.ObjectType{AttrTypes: EntryAttrTypes()}, entryValues)
}

// SetAppliesFromClient sets the apply blocks from client data.
func (m *AccessListExtendedModel) SetAppliesFromClient(applies []client.ExtendedApply) {
	if len(applies) == 0 {
		m.Apply = types.ListNull(types.ObjectType{AttrTypes: ApplyAttrTypes()})
		return
	}

	applyValues := make([]attr.Value, len(applies))
	for i, apply := range applies {
		// Convert sequences
		sequenceValues := make([]attr.Value, len(apply.FilterIDs))
		for j, id := range apply.FilterIDs {
			sequenceValues[j] = types.Int64Value(int64(id))
		}

		applyValues[i] = types.ObjectValueMust(ApplyAttrTypes(), map[string]attr.Value{
			"interface": types.StringValue(apply.Interface),
			"direction": types.StringValue(apply.Direction),
			"sequences": types.ListValueMust(types.Int64Type, sequenceValues),
		})
	}
	m.Apply = types.ListValueMust(types.ObjectType{AttrTypes: ApplyAttrTypes()}, applyValues)
}
