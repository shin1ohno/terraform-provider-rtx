package access_list_ipv6

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// AccessListIPv6Model describes the resource data model.
type AccessListIPv6Model struct {
	Name          types.String `tfsdk:"name"`
	SequenceStart types.Int64  `tfsdk:"sequence_start"`
	SequenceStep  types.Int64  `tfsdk:"sequence_step"`
	Entry         []EntryModel `tfsdk:"entry"`
	Apply         []ApplyModel `tfsdk:"apply"`
}

// EntryModel describes an IPv6 filter entry.
type EntryModel struct {
	Sequence    types.Int64  `tfsdk:"sequence"`
	Action      types.String `tfsdk:"action"`
	Source      types.String `tfsdk:"source"`
	Destination types.String `tfsdk:"destination"`
	Protocol    types.String `tfsdk:"protocol"`
	SourcePort  types.String `tfsdk:"source_port"`
	DestPort    types.String `tfsdk:"dest_port"`
	Log         types.Bool   `tfsdk:"log"`
}

// ApplyModel describes an interface binding configuration.
type ApplyModel struct {
	Interface types.String `tfsdk:"interface"`
	Direction types.String `tfsdk:"direction"`
	FilterIDs types.List   `tfsdk:"filter_ids"`
}

// EntryAttrTypes returns the attribute types for EntryModel.
func EntryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"sequence":    types.Int64Type,
		"action":      types.StringType,
		"source":      types.StringType,
		"destination": types.StringType,
		"protocol":    types.StringType,
		"source_port": types.StringType,
		"dest_port":   types.StringType,
		"log":         types.BoolType,
	}
}

// ApplyAttrTypes returns the attribute types for ApplyModel.
func ApplyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"interface": types.StringType,
		"direction": types.StringType,
		"filter_ids": types.ListType{
			ElemType: types.Int64Type,
		},
	}
}

// DefaultSequenceStep is the default step between sequence numbers.
const DefaultSequenceStep = 10

// ToFilters converts the Terraform model entries to client.IPFilter slice.
func (m *AccessListIPv6Model) ToFilters(ctx context.Context, diagnostics *diag.Diagnostics) []client.IPFilter {
	if len(m.Entry) == 0 {
		return nil
	}

	sequenceStart := fwhelpers.GetInt64Value(m.SequenceStart)
	sequenceStep := fwhelpers.GetInt64Value(m.SequenceStep)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	filters := make([]client.IPFilter, len(m.Entry))
	for i, entry := range m.Entry {
		// Determine sequence
		var seq int
		if sequenceStart > 0 {
			// Auto mode
			seq = sequenceStart + (i * sequenceStep)
		} else {
			// Manual mode
			seq = int(entry.Sequence.ValueInt64())
		}

		filters[i] = client.IPFilter{
			Number:        seq,
			Action:        fwhelpers.GetStringValue(entry.Action),
			SourceAddress: fwhelpers.GetStringValue(entry.Source),
			DestAddress:   fwhelpers.GetStringValue(entry.Destination),
			Protocol:      fwhelpers.GetStringValue(entry.Protocol),
			SourcePort:    fwhelpers.GetStringValue(entry.SourcePort),
			DestPort:      fwhelpers.GetStringValue(entry.DestPort),
		}
	}

	return filters
}

// GetExpectedSequences returns the sequence numbers expected based on configuration.
func (m *AccessListIPv6Model) GetExpectedSequences() []int {
	if len(m.Entry) == 0 {
		return nil
	}

	sequenceStart := fwhelpers.GetInt64Value(m.SequenceStart)
	sequenceStep := fwhelpers.GetInt64Value(m.SequenceStep)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	sequences := make([]int, 0, len(m.Entry))
	for i, entry := range m.Entry {
		var seq int
		if sequenceStart > 0 {
			seq = sequenceStart + (i * sequenceStep)
		} else {
			seq = int(entry.Sequence.ValueInt64())
		}
		if seq > 0 {
			sequences = append(sequences, seq)
		}
	}

	return sequences
}

// FromFilters updates the Terraform model entries from client.IPFilter slice.
func (m *AccessListIPv6Model) FromFilters(ctx context.Context, filters []*client.IPFilter, diagnostics *diag.Diagnostics) {
	if len(filters) == 0 {
		m.Entry = nil
		return
	}

	m.Entry = make([]EntryModel, len(filters))
	for i, filter := range filters {
		m.Entry[i] = EntryModel{
			Sequence:    types.Int64Value(int64(filter.Number)),
			Action:      types.StringValue(filter.Action),
			Source:      types.StringValue(filter.SourceAddress),
			Destination: types.StringValue(filter.DestAddress),
			Protocol:    types.StringValue(filter.Protocol),
			SourcePort:  types.StringValue(normalizePort(filter.SourcePort)),
			DestPort:    types.StringValue(normalizePort(filter.DestPort)),
			Log:         types.BoolValue(false), // RTX doesn't return log status in filter read
		}
	}
}

// GetApplyFilterIDs extracts filter IDs from an apply block, falling back to entry sequences.
func (m *AccessListIPv6Model) GetApplyFilterIDs(apply *ApplyModel) []int {
	if apply.FilterIDs.IsNull() || apply.FilterIDs.IsUnknown() {
		// Fall back to all entry sequences
		return m.GetExpectedSequences()
	}

	elements := apply.FilterIDs.Elements()
	if len(elements) == 0 {
		return m.GetExpectedSequences()
	}

	ids := make([]int, 0, len(elements))
	for _, elem := range elements {
		if intVal, ok := elem.(types.Int64); ok && !intVal.IsNull() && !intVal.IsUnknown() {
			ids = append(ids, int(intVal.ValueInt64()))
		}
	}
	return ids
}

// SetApplyFilterIDs sets the filter IDs on an apply block.
func SetApplyFilterIDs(apply *ApplyModel, filterIDs []int) {
	if len(filterIDs) == 0 {
		apply.FilterIDs = types.ListValueMust(types.Int64Type, []attr.Value{})
		return
	}

	elements := make([]attr.Value, len(filterIDs))
	for i, id := range filterIDs {
		elements[i] = types.Int64Value(int64(id))
	}
	apply.FilterIDs = types.ListValueMust(types.Int64Type, elements)
}

// normalizePort normalizes port values for display.
func normalizePort(port string) string {
	if port == "" {
		return "*"
	}
	return port
}

// FindRemovedSequences finds sequences that were in old but not in new.
func FindRemovedSequences(old, new []int) []int {
	newSet := make(map[int]bool)
	for _, seq := range new {
		newSet[seq] = true
	}

	var removed []int
	for _, seq := range old {
		if !newSet[seq] {
			removed = append(removed, seq)
		}
	}

	return removed
}

// ExtractSequencesFromEntries extracts sequences from entry list with sequence calculation.
func ExtractSequencesFromEntries(entries []EntryModel, sequenceStart, sequenceStep int) []int {
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	sequences := make([]int, 0, len(entries))
	for i, entry := range entries {
		var seq int
		if sequenceStart > 0 {
			seq = sequenceStart + (i * sequenceStep)
		} else {
			seq = int(entry.Sequence.ValueInt64())
		}
		if seq > 0 {
			sequences = append(sequences, seq)
		}
	}

	return sequences
}

// ValidateEntryPorts validates that port specifications are only used for TCP/UDP.
func ValidateEntryPorts(entries []EntryModel) error {
	for i, entry := range entries {
		protocol := strings.ToLower(fwhelpers.GetStringValue(entry.Protocol))
		sourcePort := fwhelpers.GetStringValue(entry.SourcePort)
		destPort := fwhelpers.GetStringValue(entry.DestPort)

		// Port specifications only valid for TCP/UDP
		if protocol != "tcp" && protocol != "udp" {
			if sourcePort != "*" && sourcePort != "" {
				return &EntryValidationError{
					Index:   i,
					Field:   "source_port",
					Message: "can only be specified for tcp or udp protocols",
				}
			}
			if destPort != "*" && destPort != "" {
				return &EntryValidationError{
					Index:   i,
					Field:   "dest_port",
					Message: "can only be specified for tcp or udp protocols",
				}
			}
		}
	}
	return nil
}

// EntryValidationError represents an entry validation error.
type EntryValidationError struct {
	Index   int
	Field   string
	Message string
}

func (e *EntryValidationError) Error() string {
	return "entry[" + string(rune('0'+e.Index)) + "]: " + e.Field + " " + e.Message
}
