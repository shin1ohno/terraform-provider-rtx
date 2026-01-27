package dns_server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// DNSServerModel describes the resource data model.
type DNSServerModel struct {
	ID                  types.String `tfsdk:"id"`
	DomainLookup        types.Bool   `tfsdk:"domain_lookup"`
	DomainName          types.String `tfsdk:"domain_name"`
	NameServers         types.List   `tfsdk:"name_servers"`
	ServerSelect        types.List   `tfsdk:"server_select"`
	Hosts               types.List   `tfsdk:"hosts"`
	ServiceOn           types.Bool   `tfsdk:"service_on"`
	PrivateAddressSpoof types.Bool   `tfsdk:"private_address_spoof"`
}

// DNSServerSelectModel represents a domain-based DNS server selection entry.
type DNSServerSelectModel struct {
	Priority       types.Int64  `tfsdk:"priority"`
	Server         types.List   `tfsdk:"server"`
	RecordType     types.String `tfsdk:"record_type"`
	QueryPattern   types.String `tfsdk:"query_pattern"`
	OriginalSender types.String `tfsdk:"original_sender"`
	RestrictPP     types.Int64  `tfsdk:"restrict_pp"`
}

// DNSServerEntryModel represents a DNS server entry with EDNS setting.
type DNSServerEntryModel struct {
	Address types.String `tfsdk:"address"`
	EDNS    types.Bool   `tfsdk:"edns"`
}

// DNSHostModel represents a static DNS host entry.
type DNSHostModel struct {
	Name    types.String `tfsdk:"name"`
	Address types.String `tfsdk:"address"`
}

// ToClient converts the Terraform model to a client.DNSConfig.
func (m *DNSServerModel) ToClient(ctx context.Context, diags *diag.Diagnostics) client.DNSConfig {
	config := client.DNSConfig{
		DomainLookup: fwhelpers.GetBoolValue(m.DomainLookup),
		DomainName:   fwhelpers.GetStringValue(m.DomainName),
		ServiceOn:    fwhelpers.GetBoolValue(m.ServiceOn),
		PrivateSpoof: fwhelpers.GetBoolValue(m.PrivateAddressSpoof),
		NameServers:  []string{},
		ServerSelect: []client.DNSServerSelect{},
		Hosts:        []client.DNSHost{},
	}

	// Convert name_servers list
	if !m.NameServers.IsNull() && !m.NameServers.IsUnknown() {
		var nameServers []types.String
		d := m.NameServers.ElementsAs(ctx, &nameServers, false)
		diags.Append(d...)
		if !diags.HasError() {
			for _, ns := range nameServers {
				config.NameServers = append(config.NameServers, ns.ValueString())
			}
		}
	}

	// Convert server_select list
	if !m.ServerSelect.IsNull() && !m.ServerSelect.IsUnknown() {
		var serverSelects []DNSServerSelectModel
		d := m.ServerSelect.ElementsAs(ctx, &serverSelects, false)
		diags.Append(d...)
		if !diags.HasError() {
			for _, sel := range serverSelects {
				serverSelect := client.DNSServerSelect{
					ID:             int(sel.Priority.ValueInt64()),
					RecordType:     fwhelpers.GetStringValue(sel.RecordType),
					QueryPattern:   fwhelpers.GetStringValue(sel.QueryPattern),
					OriginalSender: fwhelpers.GetStringValue(sel.OriginalSender),
					RestrictPP:     int(sel.RestrictPP.ValueInt64()),
					Servers:        []client.DNSServer{},
				}

				// Set default record type if not specified
				if serverSelect.RecordType == "" {
					serverSelect.RecordType = "a"
				}

				// Convert server entries
				if !sel.Server.IsNull() && !sel.Server.IsUnknown() {
					var servers []DNSServerEntryModel
					d := sel.Server.ElementsAs(ctx, &servers, false)
					diags.Append(d...)
					if !diags.HasError() {
						for _, srv := range servers {
							serverSelect.Servers = append(serverSelect.Servers, client.DNSServer{
								Address: fwhelpers.GetStringValue(srv.Address),
								EDNS:    fwhelpers.GetBoolValue(srv.EDNS),
							})
						}
					}
				}

				config.ServerSelect = append(config.ServerSelect, serverSelect)
			}
		}
	}

	// Convert hosts list
	if !m.Hosts.IsNull() && !m.Hosts.IsUnknown() {
		var hosts []DNSHostModel
		d := m.Hosts.ElementsAs(ctx, &hosts, false)
		diags.Append(d...)
		if !diags.HasError() {
			for _, host := range hosts {
				config.Hosts = append(config.Hosts, client.DNSHost{
					Name:    fwhelpers.GetStringValue(host.Name),
					Address: fwhelpers.GetStringValue(host.Address),
				})
			}
		}
	}

	return config
}

// FromClient updates the Terraform model from a client.DNSConfig.
func (m *DNSServerModel) FromClient(ctx context.Context, config *client.DNSConfig, diags *diag.Diagnostics) {
	m.ID = types.StringValue("dns")
	m.DomainLookup = types.BoolValue(config.DomainLookup)
	m.DomainName = fwhelpers.StringValueOrNull(config.DomainName)
	m.ServiceOn = types.BoolValue(config.ServiceOn)
	m.PrivateAddressSpoof = types.BoolValue(config.PrivateSpoof)

	// Convert name_servers
	if len(config.NameServers) > 0 {
		nameServerValues := make([]attr.Value, len(config.NameServers))
		for i, ns := range config.NameServers {
			nameServerValues[i] = types.StringValue(ns)
		}
		listVal, d := types.ListValue(types.StringType, nameServerValues)
		diags.Append(d...)
		m.NameServers = listVal
	} else {
		m.NameServers = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Convert server_select
	if len(config.ServerSelect) > 0 {
		serverSelectValues := make([]attr.Value, len(config.ServerSelect))
		for i, sel := range config.ServerSelect {
			// Convert servers
			serverValues := make([]attr.Value, len(sel.Servers))
			for j, srv := range sel.Servers {
				serverObj, d := types.ObjectValue(
					DNSServerEntryAttrTypes(),
					map[string]attr.Value{
						"address": types.StringValue(srv.Address),
						"edns":    types.BoolValue(srv.EDNS),
					},
				)
				diags.Append(d...)
				serverValues[j] = serverObj
			}

			serverListVal, d := types.ListValue(types.ObjectType{AttrTypes: DNSServerEntryAttrTypes()}, serverValues)
			diags.Append(d...)

			selectObj, d := types.ObjectValue(
				DNSServerSelectAttrTypes(),
				map[string]attr.Value{
					"priority":        types.Int64Value(int64(sel.ID)),
					"server":          serverListVal,
					"record_type":     fwhelpers.StringValueOrNull(sel.RecordType),
					"query_pattern":   types.StringValue(sel.QueryPattern),
					"original_sender": fwhelpers.StringValueOrNull(sel.OriginalSender),
					"restrict_pp":     types.Int64Value(int64(sel.RestrictPP)),
				},
			)
			diags.Append(d...)
			serverSelectValues[i] = selectObj
		}
		listVal, d := types.ListValue(types.ObjectType{AttrTypes: DNSServerSelectAttrTypes()}, serverSelectValues)
		diags.Append(d...)
		m.ServerSelect = listVal
	} else {
		m.ServerSelect = types.ListValueMust(types.ObjectType{AttrTypes: DNSServerSelectAttrTypes()}, []attr.Value{})
	}

	// Convert hosts
	if len(config.Hosts) > 0 {
		hostValues := make([]attr.Value, len(config.Hosts))
		for i, host := range config.Hosts {
			hostObj, d := types.ObjectValue(
				DNSHostAttrTypes(),
				map[string]attr.Value{
					"name":    types.StringValue(host.Name),
					"address": types.StringValue(host.Address),
				},
			)
			diags.Append(d...)
			hostValues[i] = hostObj
		}
		listVal, d := types.ListValue(types.ObjectType{AttrTypes: DNSHostAttrTypes()}, hostValues)
		diags.Append(d...)
		m.Hosts = listVal
	} else {
		m.Hosts = types.ListValueMust(types.ObjectType{AttrTypes: DNSHostAttrTypes()}, []attr.Value{})
	}
}

// DNSServerEntryAttrTypes returns the attribute types for DNSServerEntryModel.
func DNSServerEntryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"address": types.StringType,
		"edns":    types.BoolType,
	}
}

// DNSServerSelectAttrTypes returns the attribute types for DNSServerSelectModel.
func DNSServerSelectAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"priority":        types.Int64Type,
		"server":          types.ListType{ElemType: types.ObjectType{AttrTypes: DNSServerEntryAttrTypes()}},
		"record_type":     types.StringType,
		"query_pattern":   types.StringType,
		"original_sender": types.StringType,
		"restrict_pp":     types.Int64Type,
	}
}

// DNSHostAttrTypes returns the attribute types for DNSHostModel.
func DNSHostAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":    types.StringType,
		"address": types.StringType,
	}
}
