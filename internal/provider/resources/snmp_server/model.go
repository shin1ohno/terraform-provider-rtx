package snmp_server

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
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
		SysLocation: fwhelpers.GetStringValue(m.Location),
		SysContact:  fwhelpers.GetStringValue(m.Contact),
		SysName:     fwhelpers.GetStringValue(m.ChassisID),
		Communities: []client.SNMPCommunity{},
		Hosts:       []client.SNMPHost{},
		TrapEnable:  []string{},
	}

	// Convert communities
	if !m.Communities.IsNull() && !m.Communities.IsUnknown() {
		var communities []CommunityModel
		m.Communities.ElementsAs(nil, &communities, false)
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
		m.Hosts.ElementsAs(nil, &hosts, false)
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
		m.EnableTraps.ElementsAs(nil, &traps, false)
		for _, t := range traps {
			config.TrapEnable = append(config.TrapEnable, fwhelpers.GetStringValue(t))
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

	// Convert communities
	if len(config.Communities) > 0 {
		communityValues := make([]attr.Value, len(config.Communities))
		for i, c := range config.Communities {
			communityValues[i] = types.ObjectValueMust(CommunityAttrTypes(), map[string]attr.Value{
				"name":       types.StringValue(c.Name),
				"permission": types.StringValue(c.Permission),
				"acl":        fwhelpers.StringValueOrNull(c.ACL),
			})
		}
		m.Communities = types.ListValueMust(types.ObjectType{AttrTypes: CommunityAttrTypes()}, communityValues)
	} else {
		m.Communities = types.ListValueMust(types.ObjectType{AttrTypes: CommunityAttrTypes()}, []attr.Value{})
	}

	// Convert hosts
	if len(config.Hosts) > 0 {
		hostValues := make([]attr.Value, len(config.Hosts))
		for i, h := range config.Hosts {
			hostValues[i] = types.ObjectValueMust(HostAttrTypes(), map[string]attr.Value{
				"ip_address": types.StringValue(h.Address),
				"community":  fwhelpers.StringValueOrNull(h.Community),
				"version":    fwhelpers.StringValueOrNull(h.Version),
			})
		}
		m.Hosts = types.ListValueMust(types.ObjectType{AttrTypes: HostAttrTypes()}, hostValues)
	} else {
		m.Hosts = types.ListValueMust(types.ObjectType{AttrTypes: HostAttrTypes()}, []attr.Value{})
	}

	// Convert enable_traps
	if len(config.TrapEnable) > 0 {
		trapValues := make([]attr.Value, len(config.TrapEnable))
		for i, t := range config.TrapEnable {
			trapValues[i] = types.StringValue(t)
		}
		m.EnableTraps = types.ListValueMust(types.StringType, trapValues)
	} else {
		m.EnableTraps = types.ListValueMust(types.StringType, []attr.Value{})
	}
}

// convertParsedSNMPConfig converts a parser SNMPConfig to a client SNMPConfig.
func convertParsedSNMPConfig(parsed *parsers.SNMPConfig) *client.SNMPConfig {
	config := &client.SNMPConfig{
		SysName:     parsed.SysName,
		SysLocation: parsed.SysLocation,
		SysContact:  parsed.SysContact,
		TrapEnable:  parsed.TrapEnable,
	}

	// Convert Communities
	config.Communities = make([]client.SNMPCommunity, len(parsed.Communities))
	for i, c := range parsed.Communities {
		config.Communities[i] = client.SNMPCommunity{
			Name:       c.Name,
			Permission: c.Permission,
			ACL:        c.ACL,
		}
	}

	// Convert Hosts
	config.Hosts = make([]client.SNMPHost, len(parsed.Hosts))
	for i, h := range parsed.Hosts {
		config.Hosts[i] = client.SNMPHost{
			Address:   h.Address,
			Community: h.Community,
			Version:   h.Version,
		}
	}

	return config
}
