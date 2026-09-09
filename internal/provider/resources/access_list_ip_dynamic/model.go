package access_list_ip_dynamic

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// DefaultSequenceStep is the default step between sequence numbers.
const DefaultSequenceStep = 10

// AccessListIPDynamicModel describes the resource data model.
type AccessListIPDynamicModel struct {
	Name          types.String `tfsdk:"name"`
	SequenceStart types.Int64  `tfsdk:"sequence_start"`
	SequenceStep  types.Int64  `tfsdk:"sequence_step"`
	Entries       []EntryModel `tfsdk:"entry"`
}

// EntryModel describes a single dynamic filter entry.
type EntryModel struct {
	Sequence    types.Int64  `tfsdk:"sequence"`
	Source      types.String `tfsdk:"source"`
	Destination types.String `tfsdk:"destination"`
	Protocol    types.String `tfsdk:"protocol"`
	Syslog      types.Bool   `tfsdk:"syslog"`
	Timeout     types.Int64  `tfsdk:"timeout"`
}

// ToClient converts the Terraform model to a client.AccessListIPDynamic.
func (m *AccessListIPDynamicModel) ToClient() client.AccessListIPDynamic {
	sequenceStart := fwhelpers.GetInt64Value(m.SequenceStart)
	sequenceStep := fwhelpers.GetInt64Value(m.SequenceStep)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	acl := client.AccessListIPDynamic{
		Name:    fwhelpers.GetStringValue(m.Name),
		Entries: make([]client.AccessListIPDynamicEntry, 0, len(m.Entries)),
	}

	for i, entry := range m.Entries {
		var seq int
		if sequenceStart > 0 {
			// Auto mode
			seq = sequenceStart + (i * sequenceStep)
		} else {
			// Manual mode
			seq = fwhelpers.GetInt64Value(entry.Sequence)
		}

		aclEntry := client.AccessListIPDynamicEntry{
			Sequence:    seq,
			Source:      fwhelpers.GetStringValue(entry.Source),
			Destination: fwhelpers.GetStringValue(entry.Destination),
			Protocol:    fwhelpers.GetStringValue(entry.Protocol),
			Syslog:      fwhelpers.GetBoolValue(entry.Syslog),
		}

		if !entry.Timeout.IsNull() && !entry.Timeout.IsUnknown() {
			timeout := fwhelpers.GetInt64Value(entry.Timeout)
			if timeout > 0 {
				aclEntry.Timeout = &timeout
			}
		}

		acl.Entries = append(acl.Entries, aclEntry)
	}

	return acl
}

// FromClient updates the Terraform model from a client.AccessListIPDynamic.
// It only updates entries that are already in the state (by sequence number)
// to prevent filters from other access lists from leaking into this resource's state.
// When currentSeqs is empty (during import), it includes all entries from the router
// that match the configured sequence numbers in the model.
func (m *AccessListIPDynamicModel) FromClient(acl *client.AccessListIPDynamic, currentSeqs map[int]bool) {
	m.Name = types.StringValue(acl.Name)

	sequenceStart := fwhelpers.GetInt64Value(m.SequenceStart)
	sequenceStep := fwhelpers.GetInt64Value(m.SequenceStep)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	// Build a map of entries from the router response
	entryMap := make(map[int]client.AccessListIPDynamicEntry)
	for _, entry := range acl.Entries {
		entryMap[entry.Sequence] = entry
	}

	// During import (currentSeqs is empty and no entries in model), include all router entries
	isImport := len(currentSeqs) == 0 && len(m.Entries) == 0
	if isImport {
		newEntries := make([]EntryModel, 0, len(acl.Entries))
		for _, entry := range acl.Entries {
			newEntry := EntryModel{
				Sequence:    types.Int64Value(int64(entry.Sequence)),
				Source:      types.StringValue(entry.Source),
				Destination: types.StringValue(entry.Destination),
				Protocol:    types.StringValue(entry.Protocol),
				Syslog:      types.BoolValue(entry.Syslog),
			}

			if entry.Timeout != nil {
				newEntry.Timeout = types.Int64Value(int64(*entry.Timeout))
			} else {
				newEntry.Timeout = types.Int64Null()
			}

			newEntries = append(newEntries, newEntry)
		}
		m.Entries = newEntries
		return
	}

	// Update entries that are in the state
	newEntries := make([]EntryModel, 0, len(m.Entries))
	for i, stateEntry := range m.Entries {
		var seq int
		if sequenceStart > 0 {
			// Auto mode
			seq = sequenceStart + (i * sequenceStep)
		} else {
			// Manual mode
			seq = fwhelpers.GetInt64Value(stateEntry.Sequence)
		}

		if !currentSeqs[seq] {
			continue
		}

		if entry, found := entryMap[seq]; found {
			newEntry := EntryModel{
				Sequence:    types.Int64Value(int64(entry.Sequence)),
				Source:      types.StringValue(entry.Source),
				Destination: types.StringValue(entry.Destination),
				Protocol:    types.StringValue(entry.Protocol),
				Syslog:      types.BoolValue(entry.Syslog),
			}

			if entry.Timeout != nil {
				newEntry.Timeout = types.Int64Value(int64(*entry.Timeout))
			} else {
				newEntry.Timeout = types.Int64Null()
			}

			newEntries = append(newEntries, newEntry)
		}
	}

	m.Entries = newEntries
}

// GetCurrentSequences returns a map of sequence numbers currently in the model.
func (m *AccessListIPDynamicModel) GetCurrentSequences() map[int]bool {
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
				seq = fwhelpers.GetInt64Value(entry.Sequence)
			}
		}
		if seq > 0 {
			seqs[seq] = true
		}
	}
	return seqs
}

// GetFilterNumbers returns a slice of filter numbers (sequence numbers) from the model.
func (m *AccessListIPDynamicModel) GetFilterNumbers() []int {
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
				seq = fwhelpers.GetInt64Value(entry.Sequence)
			}
		}
		if seq > 0 {
			nums = append(nums, seq)
		}
	}
	return nums
}

// entryKey renders one dynamic filter entry in a form that is equal for a planned
// entry and for the router's copy of the same entry. Timeout is part of the key
// because it is part of the command; the form-2 lists (`filter <list> [in ..] [out ..]`)
// are too, so a form-2 row on the router never compares equal to the form-1 rows this
// resource writes.
func entryKey(source, destination, protocol string, syslog bool, timeout *int, filterList, inList, outList []int) string {
	t := "-"
	if timeout != nil {
		t = strconv.Itoa(*timeout)
	}
	return fmt.Sprintf("%s %s %s syslog=%t timeout=%s filter=%v in=%v out=%v",
		source, destination, protocol, syslog, t, filterList, inList, outList)
}

// PlannedEntryKeys returns sequence -> entryKey for every entry this plan will write.
func (m *AccessListIPDynamicModel) PlannedEntryKeys() map[int]string {
	acl := m.ToClient()
	keys := make(map[int]string, len(acl.Entries))
	for _, e := range acl.Entries {
		if e.Sequence <= 0 {
			continue
		}
		keys[e.Sequence] = entryKey(e.Source, e.Destination, e.Protocol, e.Syslog, e.Timeout, nil, nil, nil)
	}
	return keys
}

// RouterEntryKeys returns sequence -> entryKey for every dynamic IP filter the
// router holds, whichever resource (or nobody) owns it.
func RouterEntryKeys(config *client.IPFilterDynamicConfig) map[int]string {
	if config == nil {
		return map[int]string{}
	}
	keys := make(map[int]string, len(config.Entries))
	for _, e := range config.Entries {
		keys[e.Number] = entryKey(e.Source, e.Dest, e.Protocol, e.Syslog, e.Timeout, e.FilterList, e.InFilterList, e.OutFilterList)
	}
	return keys
}
