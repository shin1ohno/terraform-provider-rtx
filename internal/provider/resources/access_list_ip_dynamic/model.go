package access_list_ip_dynamic

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// AccessListIPDynamicModel describes the resource data model.
type AccessListIPDynamicModel struct {
	Name    types.String `tfsdk:"name"`
	Entries []EntryModel `tfsdk:"entry"`
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
	acl := client.AccessListIPDynamic{
		Name:    fwhelpers.GetStringValue(m.Name),
		Entries: make([]client.AccessListIPDynamicEntry, 0, len(m.Entries)),
	}

	for _, entry := range m.Entries {
		aclEntry := client.AccessListIPDynamicEntry{
			Sequence:    fwhelpers.GetInt64Value(entry.Sequence),
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
func (m *AccessListIPDynamicModel) FromClient(acl *client.AccessListIPDynamic, currentSeqs map[int]bool) {
	m.Name = types.StringValue(acl.Name)

	// Build a map of entries from the router response
	entryMap := make(map[int]client.AccessListIPDynamicEntry)
	for _, entry := range acl.Entries {
		entryMap[entry.Sequence] = entry
	}

	// Update entries that are in the state
	newEntries := make([]EntryModel, 0, len(m.Entries))
	for _, stateEntry := range m.Entries {
		seq := fwhelpers.GetInt64Value(stateEntry.Sequence)
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
	seqs := make(map[int]bool)
	for _, entry := range m.Entries {
		if !entry.Sequence.IsNull() && !entry.Sequence.IsUnknown() {
			seqs[fwhelpers.GetInt64Value(entry.Sequence)] = true
		}
	}
	return seqs
}

// GetFilterNumbers returns a slice of filter numbers (sequence numbers) from the model.
func (m *AccessListIPDynamicModel) GetFilterNumbers() []int {
	nums := make([]int, 0, len(m.Entries))
	for _, entry := range m.Entries {
		if !entry.Sequence.IsNull() && !entry.Sequence.IsUnknown() {
			seq := fwhelpers.GetInt64Value(entry.Sequence)
			if seq > 0 {
				nums = append(nums, seq)
			}
		}
	}
	return nums
}
