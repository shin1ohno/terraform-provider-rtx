package l2tp

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// L2TPModel describes the resource data model.
type L2TPModel struct {
	TunnelID          types.Int64          `tfsdk:"tunnel_id"`
	TunnelInterface   types.String         `tfsdk:"tunnel_interface"`
	InterfaceName     types.String         `tfsdk:"interface_name"`
	Name              types.String         `tfsdk:"name"`
	Version           types.String         `tfsdk:"version"`
	Mode              types.String         `tfsdk:"mode"`
	Shutdown          types.Bool           `tfsdk:"shutdown"`
	TunnelSource      types.String         `tfsdk:"tunnel_source"`
	TunnelDestination types.String         `tfsdk:"tunnel_destination"`
	TunnelDestType    types.String         `tfsdk:"tunnel_dest_type"`
	KeepaliveEnabled  types.Bool           `tfsdk:"keepalive_enabled"`
	KeepaliveInterval types.Int64          `tfsdk:"keepalive_interval"`
	KeepaliveRetry    types.Int64          `tfsdk:"keepalive_retry"`
	DisconnectTime    types.Int64          `tfsdk:"disconnect_time"`
	AlwaysOn          types.Bool           `tfsdk:"always_on"`
	Enabled           types.Bool           `tfsdk:"enabled"`
	Authentication    *AuthenticationModel `tfsdk:"authentication"`
	IPPool            *IPPoolModel         `tfsdk:"ip_pool"`
	IPsecProfile      *IPsecProfileModel   `tfsdk:"ipsec_profile"`
	L2TPv3Config      *L2TPv3ConfigModel   `tfsdk:"l2tpv3_config"`
}

// AuthenticationModel describes the authentication nested block.
type AuthenticationModel struct {
	Method        types.String `tfsdk:"method"`
	RequestMethod types.String `tfsdk:"request_method"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
}

// IPPoolModel describes the IP pool nested block.
type IPPoolModel struct {
	Start types.String `tfsdk:"start"`
	End   types.String `tfsdk:"end"`
}

// IPsecProfileModel describes the IPsec profile nested block.
type IPsecProfileModel struct {
	Enabled      types.Bool   `tfsdk:"enabled"`
	PreSharedKey types.String `tfsdk:"pre_shared_key"`
	TunnelID     types.Int64  `tfsdk:"tunnel_id"`
}

// L2TPv3ConfigModel describes the L2TPv3 config nested block.
type L2TPv3ConfigModel struct {
	LocalRouterID      types.String `tfsdk:"local_router_id"`
	RemoteRouterID     types.String `tfsdk:"remote_router_id"`
	RemoteEndID        types.String `tfsdk:"remote_end_id"`
	SessionID          types.Int64  `tfsdk:"session_id"`
	CookieSize         types.Int64  `tfsdk:"cookie_size"`
	BridgeInterface    types.String `tfsdk:"bridge_interface"`
	TunnelAuthEnabled  types.Bool   `tfsdk:"tunnel_auth_enabled"`
	TunnelAuthPassword types.String `tfsdk:"tunnel_auth_password"`
}

// ToClient converts the Terraform model to a client.L2TPConfig.
func (m *L2TPModel) ToClient() client.L2TPConfig {
	config := client.L2TPConfig{
		ID:               int(m.TunnelID.ValueInt64()),
		Name:             fwhelpers.GetStringValue(m.Name),
		Version:          fwhelpers.GetStringValue(m.Version),
		Mode:             fwhelpers.GetStringValue(m.Mode),
		Shutdown:         fwhelpers.GetBoolValue(m.Shutdown),
		TunnelSource:     fwhelpers.GetStringValue(m.TunnelSource),
		TunnelDest:       fwhelpers.GetStringValue(m.TunnelDestination),
		TunnelDestType:   fwhelpers.GetStringValue(m.TunnelDestType),
		KeepaliveEnabled: fwhelpers.GetBoolValue(m.KeepaliveEnabled),
		DisconnectTime:   fwhelpers.GetInt64Value(m.DisconnectTime),
		AlwaysOn:         fwhelpers.GetBoolValue(m.AlwaysOn),
		Enabled:          fwhelpers.GetBoolValueWithDefault(m.Enabled, true),
	}

	// Handle authentication
	if m.Authentication != nil {
		config.Authentication = &client.L2TPAuth{
			Method:        fwhelpers.GetStringValue(m.Authentication.Method),
			RequestMethod: fwhelpers.GetStringValue(m.Authentication.RequestMethod),
			Username:      fwhelpers.GetStringValue(m.Authentication.Username),
			Password:      fwhelpers.GetStringValue(m.Authentication.Password),
		}
	}

	// Handle IP pool
	if m.IPPool != nil {
		config.IPPool = &client.L2TPIPPool{
			Start: fwhelpers.GetStringValue(m.IPPool.Start),
			End:   fwhelpers.GetStringValue(m.IPPool.End),
		}
	}

	// Handle IPsec profile
	if m.IPsecProfile != nil {
		config.IPsecProfile = &client.L2TPIPsec{
			Enabled:      fwhelpers.GetBoolValue(m.IPsecProfile.Enabled),
			PreSharedKey: fwhelpers.GetStringValue(m.IPsecProfile.PreSharedKey),
			TunnelID:     fwhelpers.GetInt64Value(m.IPsecProfile.TunnelID),
		}
	}

	// Handle L2TPv3 config
	if m.L2TPv3Config != nil {
		config.L2TPv3Config = &client.L2TPv3Config{
			LocalRouterID:   fwhelpers.GetStringValue(m.L2TPv3Config.LocalRouterID),
			RemoteRouterID:  fwhelpers.GetStringValue(m.L2TPv3Config.RemoteRouterID),
			RemoteEndID:     fwhelpers.GetStringValue(m.L2TPv3Config.RemoteEndID),
			SessionID:       fwhelpers.GetInt64Value(m.L2TPv3Config.SessionID),
			CookieSize:      fwhelpers.GetInt64Value(m.L2TPv3Config.CookieSize),
			BridgeInterface: fwhelpers.GetStringValue(m.L2TPv3Config.BridgeInterface),
		}
		if fwhelpers.GetBoolValue(m.L2TPv3Config.TunnelAuthEnabled) {
			config.L2TPv3Config.TunnelAuth = &client.L2TPTunnelAuth{
				Enabled:  true,
				Password: fwhelpers.GetStringValue(m.L2TPv3Config.TunnelAuthPassword),
			}
		}
	}

	// Handle keepalive config
	if fwhelpers.GetBoolValue(m.KeepaliveEnabled) {
		config.KeepaliveConfig = &client.L2TPKeepalive{
			Interval: fwhelpers.GetInt64Value(m.KeepaliveInterval),
			Retry:    fwhelpers.GetInt64Value(m.KeepaliveRetry),
		}
	}

	return config
}

// FromClient updates the Terraform model from a client.L2TPConfig.
func (m *L2TPModel) FromClient(config *client.L2TPConfig) {
	m.TunnelID = types.Int64Value(int64(config.ID))
	m.TunnelInterface = types.StringValue(fmt.Sprintf("tunnel%d", config.ID))
	m.InterfaceName = types.StringValue(fmt.Sprintf("tunnel%d", config.ID))
	m.Name = fwhelpers.StringValueOrNull(config.Name)
	m.Version = types.StringValue(config.Version)
	m.Mode = types.StringValue(config.Mode)
	m.Shutdown = types.BoolValue(config.Shutdown)
	m.TunnelSource = fwhelpers.StringValueOrNull(config.TunnelSource)
	m.TunnelDestination = fwhelpers.StringValueOrNull(config.TunnelDest)
	m.TunnelDestType = fwhelpers.StringValueOrNull(config.TunnelDestType)
	m.KeepaliveEnabled = types.BoolValue(config.KeepaliveEnabled)
	// DisconnectTime: 0 is a valid value (means no timeout), so always use Int64Value
	m.DisconnectTime = types.Int64Value(int64(config.DisconnectTime))
	m.AlwaysOn = types.BoolValue(config.AlwaysOn)
	m.Enabled = types.BoolValue(config.Enabled)

	// Update authentication - skip for L2TPv2 LNS as router parsing is unreliable
	// For LNS mode, preserve existing state values
	if config.Authentication != nil && !(config.Version == "l2tp" && config.Mode == "lns") {
		if m.Authentication == nil {
			m.Authentication = &AuthenticationModel{}
		}
		m.Authentication.Method = fwhelpers.StringValueOrNull(config.Authentication.Method)
		m.Authentication.RequestMethod = fwhelpers.StringValueOrNull(config.Authentication.RequestMethod)
		m.Authentication.Username = fwhelpers.StringValueOrNull(config.Authentication.Username)
		m.Authentication.Password = fwhelpers.StringValueOrNull(config.Authentication.Password)
	}

	// Update IP pool - skip for L2TPv2 LNS as router parsing is unreliable
	if config.IPPool != nil && !(config.Version == "l2tp" && config.Mode == "lns") {
		if m.IPPool == nil {
			m.IPPool = &IPPoolModel{}
		}
		m.IPPool.Start = types.StringValue(config.IPPool.Start)
		m.IPPool.End = types.StringValue(config.IPPool.End)
	}

	// Update IPsec profile
	if config.IPsecProfile != nil {
		if m.IPsecProfile == nil {
			m.IPsecProfile = &IPsecProfileModel{}
		}
		m.IPsecProfile.Enabled = types.BoolValue(config.IPsecProfile.Enabled)
		m.IPsecProfile.PreSharedKey = fwhelpers.StringValueOrNull(config.IPsecProfile.PreSharedKey)
		m.IPsecProfile.TunnelID = fwhelpers.Int64ValueOrNull(config.IPsecProfile.TunnelID)
	}

	// Update L2TPv3 config
	if config.L2TPv3Config != nil {
		if m.L2TPv3Config == nil {
			m.L2TPv3Config = &L2TPv3ConfigModel{}
		}
		m.L2TPv3Config.LocalRouterID = fwhelpers.StringValueOrNull(config.L2TPv3Config.LocalRouterID)
		m.L2TPv3Config.RemoteRouterID = fwhelpers.StringValueOrNull(config.L2TPv3Config.RemoteRouterID)
		m.L2TPv3Config.RemoteEndID = fwhelpers.StringValueOrNull(config.L2TPv3Config.RemoteEndID)
		m.L2TPv3Config.SessionID = fwhelpers.Int64ValueOrNull(config.L2TPv3Config.SessionID)
		m.L2TPv3Config.CookieSize = fwhelpers.Int64ValueOrNull(config.L2TPv3Config.CookieSize)
		m.L2TPv3Config.BridgeInterface = fwhelpers.StringValueOrNull(config.L2TPv3Config.BridgeInterface)
		if config.L2TPv3Config.TunnelAuth != nil {
			m.L2TPv3Config.TunnelAuthEnabled = types.BoolValue(config.L2TPv3Config.TunnelAuth.Enabled)
			// Note: tunnel_auth_password is WriteOnly, so we don't read it back
		} else {
			m.L2TPv3Config.TunnelAuthEnabled = types.BoolValue(false)
		}
	}

	// Update keepalive config
	if config.KeepaliveConfig != nil {
		m.KeepaliveInterval = fwhelpers.Int64ValueOrNull(config.KeepaliveConfig.Interval)
		m.KeepaliveRetry = fwhelpers.Int64ValueOrNull(config.KeepaliveConfig.Retry)
	} else {
		// When keepalive is not enabled, set to null
		m.KeepaliveInterval = types.Int64Null()
		m.KeepaliveRetry = types.Int64Null()
	}
}
