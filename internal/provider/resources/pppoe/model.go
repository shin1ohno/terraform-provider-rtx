package pppoe

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// PPPoEModel describes the resource data model.
type PPPoEModel struct {
	ID                types.String `tfsdk:"id"`
	PPNumber          types.Int64  `tfsdk:"pp_number"`
	Name              types.String `tfsdk:"name"`
	BindInterface     types.String `tfsdk:"bind_interface"`
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
	ServiceName       types.String `tfsdk:"service_name"`
	ACName            types.String `tfsdk:"ac_name"`
	AuthMethod        types.String `tfsdk:"auth_method"`
	AlwaysOn          types.Bool   `tfsdk:"always_on"`
	DisconnectTimeout types.Int64  `tfsdk:"disconnect_timeout"`
	ReconnectInterval types.Int64  `tfsdk:"reconnect_interval"`
	ReconnectAttempts types.Int64  `tfsdk:"reconnect_attempts"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	PPInterface       types.String `tfsdk:"pp_interface"`
}

// ToClient converts the Terraform model to a client.PPPoEConfig.
func (m *PPPoEModel) ToClient() client.PPPoEConfig {
	config := client.PPPoEConfig{
		Number:            fwhelpers.GetInt64Value(m.PPNumber),
		Name:              fwhelpers.GetStringValue(m.Name),
		BindInterface:     fwhelpers.GetStringValue(m.BindInterface),
		ServiceName:       fwhelpers.GetStringValue(m.ServiceName),
		ACName:            fwhelpers.GetStringValue(m.ACName),
		AlwaysOn:          fwhelpers.GetBoolValue(m.AlwaysOn),
		Enabled:           fwhelpers.GetBoolValue(m.Enabled),
		DisconnectTimeout: fwhelpers.GetInt64Value(m.DisconnectTimeout),
	}

	// Set authentication
	config.Authentication = &client.PPPAuth{
		Method:   fwhelpers.GetStringValue(m.AuthMethod),
		Username: fwhelpers.GetStringValue(m.Username),
		Password: fwhelpers.GetStringValue(m.Password),
	}

	// Reconnect/keepalive settings
	if !m.ReconnectInterval.IsNull() && !m.ReconnectInterval.IsUnknown() {
		config.LCPReconnect = &client.LCPReconnectConfig{
			ReconnectInterval: fwhelpers.GetInt64Value(m.ReconnectInterval),
			ReconnectAttempts: fwhelpers.GetInt64Value(m.ReconnectAttempts),
		}
	}

	return config
}

// FromClient updates the Terraform model from a client.PPPoEConfig.
func (m *PPPoEModel) FromClient(config *client.PPPoEConfig) {
	m.ID = types.StringValue(fmt.Sprintf("%d", config.Number))
	m.PPNumber = types.Int64Value(int64(config.Number))
	m.Name = fwhelpers.StringValueOrNull(config.Name)
	m.BindInterface = types.StringValue(config.BindInterface)
	m.ServiceName = fwhelpers.StringValueOrNull(config.ServiceName)
	m.ACName = fwhelpers.StringValueOrNull(config.ACName)
	m.AlwaysOn = types.BoolValue(config.AlwaysOn)
	m.DisconnectTimeout = types.Int64Value(int64(config.DisconnectTimeout))
	m.Enabled = types.BoolValue(config.Enabled)
	m.PPInterface = types.StringValue(fmt.Sprintf("pp%d", config.Number))

	// Handle LCP reconnect settings
	if config.LCPReconnect != nil {
		m.ReconnectInterval = types.Int64Value(int64(config.LCPReconnect.ReconnectInterval))
		m.ReconnectAttempts = types.Int64Value(int64(config.LCPReconnect.ReconnectAttempts))
	} else {
		m.ReconnectInterval = types.Int64Null()
		m.ReconnectAttempts = types.Int64Null()
	}

	// Set authentication attributes if available
	if config.Authentication != nil {
		m.Username = types.StringValue(config.Authentication.Username)
		m.AuthMethod = fwhelpers.StringValueOrNull(config.Authentication.Method)
		// Note: Password is WriteOnly - we don't read it back from router
	}
}
