package syslog

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// SyslogModel describes the resource data model.
type SyslogModel struct {
	ID           types.String `tfsdk:"id"`
	Hosts        types.Set    `tfsdk:"host"`
	LocalAddress types.String `tfsdk:"local_address"`
	Facility     types.String `tfsdk:"facility"`
	Notice       types.Bool   `tfsdk:"notice"`
	Info         types.Bool   `tfsdk:"info"`
	Debug        types.Bool   `tfsdk:"debug"`
}

// HostModel describes a single syslog host.
type HostModel struct {
	Address types.String `tfsdk:"address"`
	Port    types.Int64  `tfsdk:"port"`
}

// HostAttrTypes returns the attribute types for HostModel.
func HostAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"address": types.StringType,
		"port":    types.Int64Type,
	}
}

// ToClient converts the Terraform model to a client.SyslogConfig.
func (m *SyslogModel) ToClient(ctx context.Context) (client.SyslogConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	config := client.SyslogConfig{
		LocalAddress: fwhelpers.GetStringValue(m.LocalAddress),
		Facility:     fwhelpers.GetStringValue(m.Facility),
		Notice:       fwhelpers.GetBoolValue(m.Notice),
		Info:         fwhelpers.GetBoolValue(m.Info),
		Debug:        fwhelpers.GetBoolValue(m.Debug),
		Hosts:        []client.SyslogHost{},
	}

	// Convert hosts set to client.SyslogHost slice
	if !m.Hosts.IsNull() && !m.Hosts.IsUnknown() {
		var hostModels []HostModel
		diags.Append(m.Hosts.ElementsAs(ctx, &hostModels, false)...)
		if diags.HasError() {
			return config, diags
		}

		for _, h := range hostModels {
			host := client.SyslogHost{
				Address: fwhelpers.GetStringValue(h.Address),
				Port:    fwhelpers.GetInt64Value(h.Port),
			}
			config.Hosts = append(config.Hosts, host)
		}
	}

	return config, diags
}

// FromClient updates the Terraform model from a client.SyslogConfig.
func (m *SyslogModel) FromClient(ctx context.Context, config *client.SyslogConfig) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue("syslog")
	m.LocalAddress = fwhelpers.StringValueOrNull(config.LocalAddress)
	m.Facility = fwhelpers.StringValueOrNull(config.Facility)
	m.Notice = types.BoolValue(config.Notice)
	m.Info = types.BoolValue(config.Info)
	m.Debug = types.BoolValue(config.Debug)

	// Convert hosts to set
	if len(config.Hosts) > 0 {
		hostValues := make([]attr.Value, len(config.Hosts))
		for i, host := range config.Hosts {
			hostObj, d := types.ObjectValue(HostAttrTypes(), map[string]attr.Value{
				"address": types.StringValue(host.Address),
				"port":    types.Int64Value(int64(host.Port)),
			})
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			hostValues[i] = hostObj
		}
		hostSet, d := types.SetValue(types.ObjectType{AttrTypes: HostAttrTypes()}, hostValues)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		m.Hosts = hostSet
	} else {
		m.Hosts = types.SetValueMust(types.ObjectType{AttrTypes: HostAttrTypes()}, []attr.Value{})
	}

	return diags
}

// HostObjectType returns the object type for host set elements.
func HostObjectType() basetypes.ObjectType {
	return types.ObjectType{AttrTypes: HostAttrTypes()}
}
