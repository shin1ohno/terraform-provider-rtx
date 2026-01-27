package access_list_ip

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// AccessListIPModel describes the resource data model for IP access list.
type AccessListIPModel struct {
	Name          types.String `tfsdk:"name"`
	SequenceStart types.Int64  `tfsdk:"sequence_start"`
	SequenceStep  types.Int64  `tfsdk:"sequence_step"`
	Apply         types.List   `tfsdk:"apply"`
	Entry         types.List   `tfsdk:"entry"`
}

// EntryModel describes a single IP filter entry.
type EntryModel struct {
	Sequence    types.Int64  `tfsdk:"sequence"`
	Action      types.String `tfsdk:"action"`
	Source      types.String `tfsdk:"source"`
	Destination types.String `tfsdk:"destination"`
	Protocol    types.String `tfsdk:"protocol"`
	SourcePort  types.String `tfsdk:"source_port"`
	DestPort    types.String `tfsdk:"dest_port"`
	Established types.Bool   `tfsdk:"established"`
	Log         types.Bool   `tfsdk:"log"`
}

// ApplyModel describes an interface binding configuration.
type ApplyModel struct {
	Interface types.String `tfsdk:"interface"`
	Direction types.String `tfsdk:"direction"`
	FilterIDs types.List   `tfsdk:"filter_ids"`
}

// EntryModelAttrTypes returns the attribute types for EntryModel.
func EntryModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"sequence":    types.Int64Type,
		"action":      types.StringType,
		"source":      types.StringType,
		"destination": types.StringType,
		"protocol":    types.StringType,
		"source_port": types.StringType,
		"dest_port":   types.StringType,
		"established": types.BoolType,
		"log":         types.BoolType,
	}
}

// ApplyModelAttrTypes returns the attribute types for ApplyModel.
func ApplyModelAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"interface":  types.StringType,
		"direction":  types.StringType,
		"filter_ids": types.ListType{ElemType: types.Int64Type},
	}
}

// DefaultSequenceStep is the default step between sequence numbers.
const DefaultSequenceStep = 10

// ToClientFilters converts the Terraform model to client.IPFilter slice.
func (m *AccessListIPModel) ToClientFilters() []client.IPFilter {
	sequenceStart := fwhelpers.GetInt64Value(m.SequenceStart)
	sequenceStep := fwhelpers.GetInt64Value(m.SequenceStep)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	if m.Entry.IsNull() || m.Entry.IsUnknown() {
		return nil
	}

	var entries []EntryModel
	m.Entry.ElementsAs(context.TODO(), &entries, false)

	result := make([]client.IPFilter, 0, len(entries))

	for i, entry := range entries {
		var seq int
		if sequenceStart > 0 {
			// Auto mode
			seq = sequenceStart + (i * sequenceStep)
		} else {
			// Manual mode
			seq = fwhelpers.GetInt64Value(entry.Sequence)
		}

		filter := client.IPFilter{
			Number:        seq,
			Action:        fwhelpers.GetStringValue(entry.Action),
			SourceAddress: fwhelpers.GetStringValue(entry.Source),
			DestAddress:   fwhelpers.GetStringValue(entry.Destination),
			Protocol:      getStringWithDefault(entry.Protocol, "*"),
			SourcePort:    getStringWithDefault(entry.SourcePort, "*"),
			DestPort:      getStringWithDefault(entry.DestPort, "*"),
			Established:   fwhelpers.GetBoolValue(entry.Established),
		}

		result = append(result, filter)
	}

	return result
}

// GetExpectedSequences returns the sequence numbers expected based on state.
func (m *AccessListIPModel) GetExpectedSequences() []int {
	sequenceStart := fwhelpers.GetInt64Value(m.SequenceStart)
	sequenceStep := fwhelpers.GetInt64Value(m.SequenceStep)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	if m.Entry.IsNull() || m.Entry.IsUnknown() {
		return nil
	}

	var entries []EntryModel
	m.Entry.ElementsAs(context.TODO(), &entries, false)

	sequences := make([]int, 0, len(entries))

	for i, entry := range entries {
		var seq int
		if sequenceStart > 0 {
			seq = sequenceStart + (i * sequenceStep)
		} else {
			seq = fwhelpers.GetInt64Value(entry.Sequence)
		}

		if seq > 0 {
			sequences = append(sequences, seq)
		}
	}

	return sequences
}

// GetApplies returns the apply configurations.
func (m *AccessListIPModel) GetApplies() []ApplyModel {
	if m.Apply.IsNull() || m.Apply.IsUnknown() {
		return nil
	}

	var applies []ApplyModel
	m.Apply.ElementsAs(context.TODO(), &applies, false)
	return applies
}

// SetEntriesFromFilters updates the model entries from client.IPFilter slice.
func (m *AccessListIPModel) SetEntriesFromFilters(filters []client.IPFilter) {
	entries := make([]EntryModel, 0, len(filters))

	for _, filter := range filters {
		entry := EntryModel{
			Sequence:    types.Int64Value(int64(filter.Number)),
			Action:      types.StringValue(filter.Action),
			Source:      types.StringValue(filter.SourceAddress),
			Destination: types.StringValue(filter.DestAddress),
			Protocol:    types.StringValue(normalizePort(filter.Protocol)),
			SourcePort:  types.StringValue(normalizePort(filter.SourcePort)),
			DestPort:    types.StringValue(normalizePort(filter.DestPort)),
			Established: types.BoolValue(filter.Established),
			Log:         types.BoolValue(false), // RTX doesn't return log status
		}
		entries = append(entries, entry)
	}

	entryValues := make([]attr.Value, len(entries))
	for i, e := range entries {
		entryValues[i] = entryToObjectValue(e)
	}

	m.Entry = types.ListValueMust(types.ObjectType{AttrTypes: EntryModelAttrTypes()}, entryValues)
}

// entryToObjectValue converts an EntryModel to an attr.Value.
func entryToObjectValue(e EntryModel) attr.Value {
	return types.ObjectValueMust(EntryModelAttrTypes(), map[string]attr.Value{
		"sequence":    e.Sequence,
		"action":      e.Action,
		"source":      e.Source,
		"destination": e.Destination,
		"protocol":    e.Protocol,
		"source_port": e.SourcePort,
		"dest_port":   e.DestPort,
		"established": e.Established,
		"log":         e.Log,
	})
}

// applyToObjectValue converts an ApplyModel to an attr.Value.
func applyToObjectValue(a ApplyModel) attr.Value {
	return types.ObjectValueMust(ApplyModelAttrTypes(), map[string]attr.Value{
		"interface":  a.Interface,
		"direction":  a.Direction,
		"filter_ids": a.FilterIDs,
	})
}

// SetAppliesFromRouter updates the apply configurations from router state.
func (m *AccessListIPModel) SetAppliesFromRouter(applies []ApplyModel) {
	applyValues := make([]attr.Value, len(applies))
	for i, a := range applies {
		applyValues[i] = applyToObjectValue(a)
	}

	m.Apply = types.ListValueMust(types.ObjectType{AttrTypes: ApplyModelAttrTypes()}, applyValues)
}

// Helper functions

func getStringWithDefault(s types.String, defaultVal string) string {
	if s.IsNull() || s.IsUnknown() {
		return defaultVal
	}
	v := s.ValueString()
	if v == "" {
		return defaultVal
	}
	return v
}

func normalizePort(port string) string {
	if port == "" {
		return "*"
	}
	return port
}
