package access_list_mac

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// AccessListMACModel describes the resource data model.
type AccessListMACModel struct {
	Name          types.String `tfsdk:"name"`
	FilterID      types.Int64  `tfsdk:"filter_id"`
	SequenceStart types.Int64  `tfsdk:"sequence_start"`
	SequenceStep  types.Int64  `tfsdk:"sequence_step"`
	Applies       []ApplyModel `tfsdk:"apply"`
	Entries       []EntryModel `tfsdk:"entry"`
}

// ApplyModel describes an interface binding for MAC ACL.
type ApplyModel struct {
	Interface types.String `tfsdk:"interface"`
	Direction types.String `tfsdk:"direction"`
	FilterIDs types.List   `tfsdk:"filter_ids"`
}

// EntryModel describes a single entry in a MAC access list.
type EntryModel struct {
	Sequence               types.Int64     `tfsdk:"sequence"`
	AceAction              types.String    `tfsdk:"ace_action"`
	SourceAny              types.Bool      `tfsdk:"source_any"`
	SourceAddress          types.String    `tfsdk:"source_address"`
	SourceAddressMask      types.String    `tfsdk:"source_address_mask"`
	DestinationAny         types.Bool      `tfsdk:"destination_any"`
	DestinationAddress     types.String    `tfsdk:"destination_address"`
	DestinationAddressMask types.String    `tfsdk:"destination_address_mask"`
	EtherType              types.String    `tfsdk:"ether_type"`
	VlanID                 types.Int64     `tfsdk:"vlan_id"`
	Log                    types.Bool      `tfsdk:"log"`
	FilterID               types.Int64     `tfsdk:"filter_id"`
	DHCPMatch              *DHCPMatchModel `tfsdk:"dhcp_match"`
	Offset                 types.Int64     `tfsdk:"offset"`
	ByteList               types.List      `tfsdk:"byte_list"`
}

// DHCPMatchModel describes DHCP-based match settings.
type DHCPMatchModel struct {
	Type  types.String `tfsdk:"type"`
	Scope types.Int64  `tfsdk:"scope"`
}

// ApplyAttrTypes returns the attribute types for ApplyModel.
func ApplyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"interface":  types.StringType,
		"direction":  types.StringType,
		"filter_ids": types.ListType{ElemType: types.Int64Type},
	}
}

// EntryAttrTypes returns the attribute types for EntryModel.
func EntryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"sequence":                 types.Int64Type,
		"ace_action":               types.StringType,
		"source_any":               types.BoolType,
		"source_address":           types.StringType,
		"source_address_mask":      types.StringType,
		"destination_any":          types.BoolType,
		"destination_address":      types.StringType,
		"destination_address_mask": types.StringType,
		"ether_type":               types.StringType,
		"vlan_id":                  types.Int64Type,
		"log":                      types.BoolType,
		"filter_id":                types.Int64Type,
		"dhcp_match": types.ObjectType{AttrTypes: map[string]attr.Type{
			"type":  types.StringType,
			"scope": types.Int64Type,
		}},
		"offset":    types.Int64Type,
		"byte_list": types.ListType{ElemType: types.StringType},
	}
}

// DHCPMatchAttrTypes returns the attribute types for DHCPMatchModel.
func DHCPMatchAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":  types.StringType,
		"scope": types.Int64Type,
	}
}

// DefaultSequenceStep is the default step between sequence numbers in auto mode.
const DefaultSequenceStep = 10

// ToClient converts the Terraform model to a client.AccessListMAC.
func (m *AccessListMACModel) ToClient(ctx context.Context, diagnostics *diag.Diagnostics) client.AccessListMAC {
	sequenceStart := fwhelpers.GetInt64Value(m.SequenceStart)
	sequenceStep := fwhelpers.GetInt64Value(m.SequenceStep)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	acl := client.AccessListMAC{
		Name:          fwhelpers.GetStringValue(m.Name),
		FilterID:      fwhelpers.GetInt64Value(m.FilterID),
		SequenceStart: sequenceStart,
		SequenceStep:  sequenceStep,
		Entries:       make([]client.AccessListMACEntry, 0, len(m.Entries)),
		Applies:       make([]client.MACApply, 0, len(m.Applies)),
	}

	// Build entries
	for i, entry := range m.Entries {
		// Determine sequence based on mode
		var entrySequence int
		if sequenceStart > 0 {
			// Auto mode: calculate sequence
			entrySequence = sequenceStart + (i * sequenceStep)
		} else {
			// Manual mode: use explicit sequence
			entrySequence = fwhelpers.GetInt64Value(entry.Sequence)
		}

		aclEntry := client.AccessListMACEntry{
			Sequence:               entrySequence,
			AceAction:              fwhelpers.GetStringValue(entry.AceAction),
			SourceAny:              fwhelpers.GetBoolValue(entry.SourceAny),
			SourceAddress:          fwhelpers.GetStringValue(entry.SourceAddress),
			SourceAddressMask:      fwhelpers.GetStringValue(entry.SourceAddressMask),
			DestinationAny:         fwhelpers.GetBoolValue(entry.DestinationAny),
			DestinationAddress:     fwhelpers.GetStringValue(entry.DestinationAddress),
			DestinationAddressMask: fwhelpers.GetStringValue(entry.DestinationAddressMask),
			EtherType:              fwhelpers.GetStringValue(entry.EtherType),
			VlanID:                 fwhelpers.GetInt64Value(entry.VlanID),
			Log:                    fwhelpers.GetBoolValue(entry.Log),
			FilterID:               fwhelpers.GetInt64Value(entry.FilterID),
			Offset:                 fwhelpers.GetInt64Value(entry.Offset),
		}

		if aclEntry.FilterID == 0 && acl.FilterID > 0 {
			aclEntry.FilterID = acl.FilterID
		}

		// Handle byte_list
		if !entry.ByteList.IsNull() && !entry.ByteList.IsUnknown() {
			var byteList []types.String
			diagnostics.Append(entry.ByteList.ElementsAs(ctx, &byteList, false)...)
			if !diagnostics.HasError() {
				aclEntry.ByteList = fwhelpers.GetStringListValue(byteList)
			}
		}

		// Handle dhcp_match
		if entry.DHCPMatch != nil {
			aclEntry.DHCPType = fwhelpers.GetStringValue(entry.DHCPMatch.Type)
			aclEntry.DHCPScope = fwhelpers.GetInt64Value(entry.DHCPMatch.Scope)
		}

		acl.Entries = append(acl.Entries, aclEntry)
	}

	// Build applies list
	for _, apply := range m.Applies {
		var ids []int
		if !apply.FilterIDs.IsNull() && !apply.FilterIDs.IsUnknown() {
			var filterIDs []types.Int64
			diagnostics.Append(apply.FilterIDs.ElementsAs(ctx, &filterIDs, false)...)
			if !diagnostics.HasError() {
				for _, id := range filterIDs {
					ids = append(ids, int(id.ValueInt64()))
				}
			}
		}

		// If filter_ids is empty, populate with all entry sequences
		if len(ids) == 0 {
			for _, entry := range acl.Entries {
				ids = append(ids, entry.Sequence)
			}
		}

		macApply := client.MACApply{
			Interface: fwhelpers.GetStringValue(apply.Interface),
			Direction: fwhelpers.GetStringValue(apply.Direction),
			FilterIDs: ids,
		}
		acl.Applies = append(acl.Applies, macApply)
	}

	// Set legacy Apply field for backward compatibility with client layer
	if len(acl.Applies) > 0 {
		acl.Apply = &acl.Applies[0]
	}

	return acl
}

// FromClient updates the Terraform model from a client.AccessListMAC.
func (m *AccessListMACModel) FromClient(ctx context.Context, acl *client.AccessListMAC, diagnostics *diag.Diagnostics) {
	m.Name = types.StringValue(acl.Name)
	m.FilterID = fwhelpers.Int64ValueOrNull(acl.FilterID)

	// Set sequence_start and sequence_step if they were in the config
	if acl.SequenceStart > 0 {
		m.SequenceStart = types.Int64Value(int64(acl.SequenceStart))
		if acl.SequenceStep > 0 {
			m.SequenceStep = types.Int64Value(int64(acl.SequenceStep))
		}
	} else {
		m.SequenceStart = types.Int64Null()
	}

	// Build entries
	wildcardMAC := "*:*:*:*:*:*"
	if len(acl.Entries) > 0 {
		m.Entries = make([]EntryModel, 0, len(acl.Entries))
		for _, entry := range acl.Entries {
			// Detect wildcard addresses and set *_any fields accordingly
			sourceAny := entry.SourceAny || entry.SourceAddress == wildcardMAC
			destinationAny := entry.DestinationAny || entry.DestinationAddress == wildcardMAC

			// When *_any is true, clear the address to match the config pattern
			sourceAddress := entry.SourceAddress
			destinationAddress := entry.DestinationAddress
			if sourceAny && sourceAddress == wildcardMAC {
				sourceAddress = ""
			}
			if destinationAny && destinationAddress == wildcardMAC {
				destinationAddress = ""
			}

			entryModel := EntryModel{
				Sequence:               types.Int64Value(int64(entry.Sequence)),
				AceAction:              types.StringValue(entry.AceAction),
				SourceAny:              types.BoolValue(sourceAny),
				SourceAddress:          fwhelpers.StringValueOrNull(sourceAddress),
				SourceAddressMask:      fwhelpers.StringValueOrNull(entry.SourceAddressMask),
				DestinationAny:         types.BoolValue(destinationAny),
				DestinationAddress:     fwhelpers.StringValueOrNull(destinationAddress),
				DestinationAddressMask: fwhelpers.StringValueOrNull(entry.DestinationAddressMask),
				EtherType:              fwhelpers.StringValueOrNull(entry.EtherType),
				VlanID:                 fwhelpers.Int64ValueOrNull(entry.VlanID),
				Log:                    types.BoolValue(entry.Log),
				FilterID:               fwhelpers.Int64ValueOrNull(entry.FilterID),
				Offset:                 fwhelpers.Int64ValueOrNull(entry.Offset),
			}

			// Handle dhcp_match
			if entry.DHCPType != "" {
				entryModel.DHCPMatch = &DHCPMatchModel{
					Type:  types.StringValue(entry.DHCPType),
					Scope: fwhelpers.Int64ValueOrNull(entry.DHCPScope),
				}
			}

			// Handle byte_list
			if len(entry.ByteList) > 0 {
				byteElements := make([]attr.Value, len(entry.ByteList))
				for i, b := range entry.ByteList {
					byteElements[i] = types.StringValue(b)
				}
				list, diags := types.ListValue(types.StringType, byteElements)
				diagnostics.Append(diags...)
				if diagnostics.HasError() {
					return
				}
				entryModel.ByteList = list
			} else {
				entryModel.ByteList = types.ListNull(types.StringType)
			}

			m.Entries = append(m.Entries, entryModel)
		}
	} else {
		m.Entries = nil
	}

	// Build applies
	if len(acl.Applies) > 0 {
		m.Applies = make([]ApplyModel, 0, len(acl.Applies))
		for _, apply := range acl.Applies {
			applyModel := ApplyModel{
				Interface: types.StringValue(apply.Interface),
				Direction: types.StringValue(apply.Direction),
			}

			// Build filter_ids list
			if len(apply.FilterIDs) > 0 {
				filterElements := make([]attr.Value, len(apply.FilterIDs))
				for i, id := range apply.FilterIDs {
					filterElements[i] = types.Int64Value(int64(id))
				}
				list, diags := types.ListValue(types.Int64Type, filterElements)
				diagnostics.Append(diags...)
				if diagnostics.HasError() {
					return
				}
				applyModel.FilterIDs = list
			} else {
				applyModel.FilterIDs = types.ListNull(types.Int64Type)
			}

			m.Applies = append(m.Applies, applyModel)
		}
	} else if acl.Apply != nil {
		// Legacy single apply support
		applyModel := ApplyModel{
			Interface: types.StringValue(acl.Apply.Interface),
			Direction: types.StringValue(acl.Apply.Direction),
		}

		if len(acl.Apply.FilterIDs) > 0 {
			filterElements := make([]attr.Value, len(acl.Apply.FilterIDs))
			for i, id := range acl.Apply.FilterIDs {
				filterElements[i] = types.Int64Value(int64(id))
			}
			list, diags := types.ListValue(types.Int64Type, filterElements)
			diagnostics.Append(diags...)
			if diagnostics.HasError() {
				return
			}
			applyModel.FilterIDs = list
		} else {
			applyModel.FilterIDs = types.ListNull(types.Int64Type)
		}

		m.Applies = []ApplyModel{applyModel}
	} else {
		m.Applies = nil
	}
}

// GetFilterNumbersForDelete returns the filter numbers to delete based on the model configuration.
func (m *AccessListMACModel) GetFilterNumbersForDelete(ctx context.Context, diagnostics *diag.Diagnostics) []int {
	sequenceStart := fwhelpers.GetInt64Value(m.SequenceStart)
	sequenceStep := fwhelpers.GetInt64Value(m.SequenceStep)
	if sequenceStep == 0 {
		sequenceStep = DefaultSequenceStep
	}

	var filterNums []int
	for i, entry := range m.Entries {
		var num int

		// Determine filter number based on mode
		if sequenceStart > 0 {
			// Auto mode: calculate sequence
			num = sequenceStart + (i * sequenceStep)
		} else {
			// Manual mode: use explicit filter_id or sequence
			num = fwhelpers.GetInt64Value(entry.FilterID)
			if num == 0 {
				num = fwhelpers.GetInt64Value(entry.Sequence)
			}
		}

		if num > 0 {
			filterNums = append(filterNums, num)
		}
	}

	return filterNums
}
