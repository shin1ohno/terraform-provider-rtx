package access_list_ipv6_dynamic

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// DefaultSequenceStep is the default step between sequence numbers.
const DefaultSequenceStep = 10

// AccessListIPv6DynamicModel describes the resource data model.
type AccessListIPv6DynamicModel struct {
	Name          types.String `tfsdk:"name"`
	SequenceStart types.Int64  `tfsdk:"sequence_start"`
	SequenceStep  types.Int64  `tfsdk:"sequence_step"`
	Entries       []EntryModel `tfsdk:"entry"`
}

// EntryModel describes a single entry in the dynamic IPv6 access list.
type EntryModel struct {
	Sequence    types.Int64  `tfsdk:"sequence"`
	Source      types.String `tfsdk:"source"`
	Destination types.String `tfsdk:"destination"`
	Protocol    types.String `tfsdk:"protocol"`
	Syslog      types.Bool   `tfsdk:"syslog"`
}

// ToClient converts the Terraform model to a client.AccessListIPv6Dynamic.
func (m *AccessListIPv6DynamicModel) ToClient() client.AccessListIPv6Dynamic {
	sequenceStart := fwhelpers.GetInt64Value(m.SequenceStart)
	sequenceStep := fwhelpers.GetInt64Value(m.SequenceStep)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	acl := client.AccessListIPv6Dynamic{
		Name:    fwhelpers.GetStringValue(m.Name),
		Entries: make([]client.AccessListIPv6DynamicEntry, 0, len(m.Entries)),
	}

	for i, entry := range m.Entries {
		var seq int
		if sequenceStart > 0 {
			// Auto mode
			seq = sequenceStart + (i * sequenceStep)
		} else {
			// Manual mode
			seq = int(entry.Sequence.ValueInt64())
		}

		aclEntry := client.AccessListIPv6DynamicEntry{
			Sequence:    seq,
			Source:      fwhelpers.GetStringValue(entry.Source),
			Destination: fwhelpers.GetStringValue(entry.Destination),
			Protocol:    fwhelpers.GetStringValue(entry.Protocol),
			Syslog:      fwhelpers.GetBoolValue(entry.Syslog),
		}
		acl.Entries = append(acl.Entries, aclEntry)
	}

	return acl
}

// FromClient updates the Terraform model from a client.AccessListIPv6Dynamic.
// currentSeqs specifies which sequences to include (for filtering during read).
// When currentSeqs is empty (during import), it includes all entries from the router.
func (m *AccessListIPv6DynamicModel) FromClient(acl *client.AccessListIPv6Dynamic, currentSeqs map[int]bool) {
	m.Name = types.StringValue(acl.Name)

	sequenceStart := fwhelpers.GetInt64Value(m.SequenceStart)
	sequenceStep := fwhelpers.GetInt64Value(m.SequenceStep)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	// Build a map of entries from the router response
	entryMap := make(map[int]client.AccessListIPv6DynamicEntry)
	for _, entry := range acl.Entries {
		entryMap[entry.Sequence] = entry
	}

	// During import (currentSeqs is empty and no entries in model), include all router entries
	isImport := len(currentSeqs) == 0 && len(m.Entries) == 0
	if isImport {
		entries := make([]EntryModel, 0, len(acl.Entries))
		for _, entry := range acl.Entries {
			e := EntryModel{
				Sequence:    types.Int64Value(int64(entry.Sequence)),
				Source:      types.StringValue(entry.Source),
				Destination: types.StringValue(entry.Destination),
				Protocol:    types.StringValue(entry.Protocol),
				Syslog:      types.BoolValue(entry.Syslog),
			}
			entries = append(entries, e)
		}
		m.Entries = entries
		return
	}

	entries := make([]EntryModel, 0, len(m.Entries))
	for i, stateEntry := range m.Entries {
		var seq int
		if sequenceStart > 0 {
			// Auto mode
			seq = sequenceStart + (i * sequenceStep)
		} else {
			// Manual mode
			seq = int(stateEntry.Sequence.ValueInt64())
		}

		// Only include sequences that are already in state
		if len(currentSeqs) > 0 && !currentSeqs[seq] {
			continue
		}

		if entry, found := entryMap[seq]; found {
			e := EntryModel{
				Sequence:    types.Int64Value(int64(entry.Sequence)),
				Source:      types.StringValue(entry.Source),
				Destination: types.StringValue(entry.Destination),
				Protocol:    types.StringValue(entry.Protocol),
				Syslog:      types.BoolValue(entry.Syslog),
			}
			entries = append(entries, e)
		}
	}
	m.Entries = entries
}

// GetCurrentSequences returns a map of sequence numbers from the current entries.
func (m *AccessListIPv6DynamicModel) GetCurrentSequences() map[int]bool {
	sequenceStart := fwhelpers.GetInt64Value(m.SequenceStart)
	sequenceStep := fwhelpers.GetInt64Value(m.SequenceStep)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	seqs := make(map[int]bool)
	for i, entry := range m.Entries {
		var seq int
		if sequenceStart > 0 {
			// Auto mode
			seq = sequenceStart + (i * sequenceStep)
		} else {
			// Manual mode
			if !entry.Sequence.IsNull() && !entry.Sequence.IsUnknown() {
				seq = int(entry.Sequence.ValueInt64())
			}
		}
		if seq > 0 {
			seqs[seq] = true
		}
	}
	return seqs
}

// GetFilterNumbers returns the sequence numbers of all entries.
func (m *AccessListIPv6DynamicModel) GetFilterNumbers() []int {
	sequenceStart := fwhelpers.GetInt64Value(m.SequenceStart)
	sequenceStep := fwhelpers.GetInt64Value(m.SequenceStep)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	nums := make([]int, 0, len(m.Entries))
	for i, entry := range m.Entries {
		var seq int
		if sequenceStart > 0 {
			// Auto mode
			seq = sequenceStart + (i * sequenceStep)
		} else {
			// Manual mode
			if !entry.Sequence.IsNull() && !entry.Sequence.IsUnknown() {
				seq = int(entry.Sequence.ValueInt64())
			}
		}
		if seq > 0 {
			nums = append(nums, seq)
		}
	}
	return nums
}
