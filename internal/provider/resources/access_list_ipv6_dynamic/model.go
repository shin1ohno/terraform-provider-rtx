package access_list_ipv6_dynamic

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// AccessListIPv6DynamicModel describes the resource data model.
type AccessListIPv6DynamicModel struct {
	Name    types.String `tfsdk:"name"`
	Entries []EntryModel `tfsdk:"entry"`
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
	acl := client.AccessListIPv6Dynamic{
		Name:    fwhelpers.GetStringValue(m.Name),
		Entries: make([]client.AccessListIPv6DynamicEntry, 0, len(m.Entries)),
	}

	for _, entry := range m.Entries {
		aclEntry := client.AccessListIPv6DynamicEntry{
			Sequence:    int(entry.Sequence.ValueInt64()),
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
func (m *AccessListIPv6DynamicModel) FromClient(acl *client.AccessListIPv6Dynamic, currentSeqs map[int]bool) {
	m.Name = types.StringValue(acl.Name)

	entries := make([]EntryModel, 0, len(acl.Entries))
	for _, entry := range acl.Entries {
		// Only include sequences that are already in state
		// This prevents filters from other access lists from appearing here
		if len(currentSeqs) > 0 && !currentSeqs[entry.Sequence] {
			continue
		}

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
}

// GetCurrentSequences returns a map of sequence numbers from the current entries.
func (m *AccessListIPv6DynamicModel) GetCurrentSequences() map[int]bool {
	seqs := make(map[int]bool)
	for _, entry := range m.Entries {
		if !entry.Sequence.IsNull() && !entry.Sequence.IsUnknown() {
			seqs[int(entry.Sequence.ValueInt64())] = true
		}
	}
	return seqs
}

// GetFilterNumbers returns the sequence numbers of all entries.
func (m *AccessListIPv6DynamicModel) GetFilterNumbers() []int {
	nums := make([]int, 0, len(m.Entries))
	for _, entry := range m.Entries {
		if !entry.Sequence.IsNull() && !entry.Sequence.IsUnknown() {
			nums = append(nums, int(entry.Sequence.ValueInt64()))
		}
	}
	return nums
}
