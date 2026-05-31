package snmp_server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// SNMPServerModel describes the resource data model.
type SNMPServerModel struct {
	ID          types.String `tfsdk:"id"`
	Location    types.String `tfsdk:"location"`
	Contact     types.String `tfsdk:"contact"`
	ChassisID   types.String `tfsdk:"chassis_id"`
	Communities types.List   `tfsdk:"community"`
	Hosts       types.List   `tfsdk:"host"`
	EnableTraps types.List   `tfsdk:"enable_traps"`
	SNMPHost    types.List   `tfsdk:"snmp_host"`
	SNMPv2cHost types.List   `tfsdk:"snmpv2c_host"`
}

// CommunityModel describes a single SNMP community.
type CommunityModel struct {
	Name       types.String `tfsdk:"name"`
	Permission types.String `tfsdk:"permission"`
	ACL        types.String `tfsdk:"acl"`
}

// HostModel describes a single SNMP trap host.
type HostModel struct {
	IPAddress types.String `tfsdk:"ip_address"`
	Community types.String `tfsdk:"community"`
	Version   types.String `tfsdk:"version"`
}

// CommunityAttrTypes returns the attribute types for CommunityModel.
func CommunityAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":       types.StringType,
		"permission": types.StringType,
		"acl":        types.StringType,
	}
}

// HostAttrTypes returns the attribute types for HostModel.
func HostAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ip_address": types.StringType,
		"community":  types.StringType,
		"version":    types.StringType,
	}
}

// ToClient converts the Terraform model to a client.SNMPConfig.
func (m *SNMPServerModel) ToClient() client.SNMPConfig {
	config := client.SNMPConfig{
		SysLocation:   fwhelpers.GetStringValue(m.Location),
		SysContact:    fwhelpers.GetStringValue(m.Contact),
		SysName:       fwhelpers.GetStringValue(m.ChassisID),
		Communities:   []client.SNMPCommunity{},
		Hosts:         []client.SNMPHost{},
		TrapEnable:    []string{},
		HostAccessV1:  []string{},
		HostAccessV2c: []string{},
	}

	// Convert communities
	if !m.Communities.IsNull() && !m.Communities.IsUnknown() {
		var communities []CommunityModel
		m.Communities.ElementsAs(context.TODO(), &communities, false)
		for _, c := range communities {
			config.Communities = append(config.Communities, client.SNMPCommunity{
				Name:       fwhelpers.GetStringValue(c.Name),
				Permission: fwhelpers.GetStringValue(c.Permission),
				ACL:        fwhelpers.GetStringValue(c.ACL),
			})
		}
	}

	// Convert hosts
	if !m.Hosts.IsNull() && !m.Hosts.IsUnknown() {
		var hosts []HostModel
		m.Hosts.ElementsAs(context.TODO(), &hosts, false)
		for _, h := range hosts {
			config.Hosts = append(config.Hosts, client.SNMPHost{
				Address:   fwhelpers.GetStringValue(h.IPAddress),
				Community: fwhelpers.GetStringValue(h.Community),
				Version:   fwhelpers.GetStringValue(h.Version),
			})
		}
	}

	// Convert enable_traps
	if !m.EnableTraps.IsNull() && !m.EnableTraps.IsUnknown() {
		var traps []types.String
		m.EnableTraps.ElementsAs(context.TODO(), &traps, false)
		for _, t := range traps {
			config.TrapEnable = append(config.TrapEnable, fwhelpers.GetStringValue(t))
		}
	}

	// Convert snmp_host (SNMPv1 access-control)
	if !m.SNMPHost.IsNull() && !m.SNMPHost.IsUnknown() {
		var hosts []types.String
		m.SNMPHost.ElementsAs(context.TODO(), &hosts, false)
		for _, h := range hosts {
			config.HostAccessV1 = append(config.HostAccessV1, fwhelpers.GetStringValue(h))
		}
	}

	// Convert snmpv2c_host (SNMPv2c access-control)
	if !m.SNMPv2cHost.IsNull() && !m.SNMPv2cHost.IsUnknown() {
		var hosts []types.String
		m.SNMPv2cHost.ElementsAs(context.TODO(), &hosts, false)
		for _, h := range hosts {
			config.HostAccessV2c = append(config.HostAccessV2c, fwhelpers.GetStringValue(h))
		}
	}

	return config
}

// FromClient updates the Terraform model from a client.SNMPConfig.
func (m *SNMPServerModel) FromClient(config *client.SNMPConfig) {
	m.ID = types.StringValue("snmp")
	m.Location = fwhelpers.StringValueOrNull(config.SysLocation)
	m.Contact = fwhelpers.StringValueOrNull(config.SysContact)
	m.ChassisID = fwhelpers.StringValueOrNull(config.SysName)

	// Convert communities. Preserve prior null vs empty distinction: when the
	// router returns no communities, leave state null if the config never set
	// the block, otherwise materialize an empty list so an explicit `community`
	// block of zero entries is preserved.
	communityType := types.ObjectType{AttrTypes: CommunityAttrTypes()}
	if len(config.Communities) > 0 {
		communityValues := make([]attr.Value, len(config.Communities))
		for i, c := range config.Communities {
			communityValues[i] = types.ObjectValueMust(CommunityAttrTypes(), map[string]attr.Value{
				"name":       types.StringValue(c.Name),
				"permission": types.StringValue(c.Permission),
				"acl":        fwhelpers.StringValueOrNull(c.ACL),
			})
		}
		m.Communities = types.ListValueMust(communityType, communityValues)
	} else if m.Communities.IsNull() {
		m.Communities = types.ListNull(communityType)
	} else {
		m.Communities = types.ListValueMust(communityType, []attr.Value{})
	}

	// Convert hosts. Same null-preservation invariant as Communities.
	hostType := types.ObjectType{AttrTypes: HostAttrTypes()}
	if len(config.Hosts) > 0 {
		hostValues := make([]attr.Value, len(config.Hosts))
		for i, h := range config.Hosts {
			hostValues[i] = types.ObjectValueMust(HostAttrTypes(), map[string]attr.Value{
				"ip_address": types.StringValue(h.Address),
				"community":  fwhelpers.StringValueOrNull(h.Community),
				"version":    fwhelpers.StringValueOrNull(h.Version),
			})
		}
		m.Hosts = types.ListValueMust(hostType, hostValues)
	} else if m.Hosts.IsNull() {
		m.Hosts = types.ListNull(hostType)
	} else {
		m.Hosts = types.ListValueMust(hostType, []attr.Value{})
	}

	// Convert enable_traps. Same null-preservation invariant: framework rejects
	// the apply with "was null, but now cty.ListValEmpty(cty.String)" when an
	// unset attribute resolves to an empty list.
	if len(config.TrapEnable) > 0 {
		trapValues := make([]attr.Value, len(config.TrapEnable))
		for i, t := range config.TrapEnable {
			trapValues[i] = types.StringValue(t)
		}
		m.EnableTraps = types.ListValueMust(types.StringType, trapValues)
	} else if m.EnableTraps.IsNull() {
		m.EnableTraps = types.ListNull(types.StringType)
	} else {
		m.EnableTraps = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Convert snmp_host (SNMPv1 access-control). Same null-preservation invariant.
	m.SNMPHost = stringListWithNullPreservation(config.HostAccessV1, m.SNMPHost)

	// Convert snmpv2c_host (SNMPv2c access-control). Same null-preservation invariant.
	m.SNMPv2cHost = stringListWithNullPreservation(config.HostAccessV2c, m.SNMPv2cHost)
}

// stringListWithNullPreservation materializes a string list from values while
// preserving the prior null-vs-empty distinction (mirrors the inline branches
// used for Communities/Hosts/EnableTraps): a populated config wins; an empty
// config keeps the prior state null if it was null, otherwise becomes an empty
// list. This guards against the framework's "was null, but now
// cty.ListValEmpty(cty.String)" inconsistent-result error on apply.
func stringListWithNullPreservation(values []string, prior types.List) types.List {
	if len(values) > 0 {
		elems := make([]attr.Value, len(values))
		for i, v := range values {
			elems[i] = types.StringValue(v)
		}
		return types.ListValueMust(types.StringType, elems)
	}
	if prior.IsNull() {
		return types.ListNull(types.StringType)
	}
	return types.ListValueMust(types.StringType, []attr.Value{})
}
