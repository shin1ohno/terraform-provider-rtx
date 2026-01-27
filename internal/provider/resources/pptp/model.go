package pptp

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// PPTPModel describes the resource data model.
type PPTPModel struct {
	ID               types.String         `tfsdk:"id"`
	Shutdown         types.Bool           `tfsdk:"shutdown"`
	ListenAddress    types.String         `tfsdk:"listen_address"`
	MaxConnections   types.Int64          `tfsdk:"max_connections"`
	Authentication   *AuthenticationModel `tfsdk:"authentication"`
	Encryption       *EncryptionModel     `tfsdk:"encryption"`
	IPPool           *IPPoolModel         `tfsdk:"ip_pool"`
	DisconnectTime   types.Int64          `tfsdk:"disconnect_time"`
	KeepaliveEnabled types.Bool           `tfsdk:"keepalive_enabled"`
	Enabled          types.Bool           `tfsdk:"enabled"`
}

// AuthenticationModel describes the authentication block.
type AuthenticationModel struct {
	Method   types.String `tfsdk:"method"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// EncryptionModel describes the encryption block.
type EncryptionModel struct {
	MPPEBits types.Int64 `tfsdk:"mppe_bits"`
	Required types.Bool  `tfsdk:"required"`
}

// IPPoolModel describes the ip_pool block.
type IPPoolModel struct {
	Start types.String `tfsdk:"start"`
	End   types.String `tfsdk:"end"`
}

// ToClient converts the Terraform model to a client.PPTPConfig.
func (m *PPTPModel) ToClient() client.PPTPConfig {
	config := client.PPTPConfig{
		Shutdown:         fwhelpers.GetBoolValue(m.Shutdown),
		ListenAddress:    fwhelpers.GetStringValue(m.ListenAddress),
		MaxConnections:   fwhelpers.GetInt64Value(m.MaxConnections),
		DisconnectTime:   fwhelpers.GetInt64Value(m.DisconnectTime),
		KeepaliveEnabled: fwhelpers.GetBoolValue(m.KeepaliveEnabled),
		Enabled:          fwhelpers.GetBoolValue(m.Enabled),
	}

	// Handle authentication
	if m.Authentication != nil {
		config.Authentication = &client.PPTPAuth{
			Method:   fwhelpers.GetStringValue(m.Authentication.Method),
			Username: fwhelpers.GetStringValue(m.Authentication.Username),
			Password: fwhelpers.GetStringValue(m.Authentication.Password),
		}
	}

	// Handle encryption
	if m.Encryption != nil {
		config.Encryption = &client.PPTPEncryption{
			MPPEBits: fwhelpers.GetInt64Value(m.Encryption.MPPEBits),
			Required: fwhelpers.GetBoolValue(m.Encryption.Required),
		}
	}

	// Handle IP pool
	if m.IPPool != nil {
		config.IPPool = &client.PPTPIPPool{
			Start: fwhelpers.GetStringValue(m.IPPool.Start),
			End:   fwhelpers.GetStringValue(m.IPPool.End),
		}
	}

	return config
}

// FromClient updates the Terraform model from a client.PPTPConfig.
func (m *PPTPModel) FromClient(config *client.PPTPConfig) {
	m.ID = types.StringValue("pptp")
	m.Shutdown = types.BoolValue(config.Shutdown)
	m.ListenAddress = fwhelpers.StringValueOrNull(config.ListenAddress)
	m.MaxConnections = fwhelpers.Int64ValueOrNull(config.MaxConnections)
	m.DisconnectTime = fwhelpers.Int64ValueOrNull(config.DisconnectTime)
	m.KeepaliveEnabled = types.BoolValue(config.KeepaliveEnabled)
	m.Enabled = types.BoolValue(config.Enabled)

	// Handle authentication
	if config.Authentication != nil {
		m.Authentication = &AuthenticationModel{
			Method:   types.StringValue(config.Authentication.Method),
			Username: fwhelpers.StringValueOrNull(config.Authentication.Username),
			Password: fwhelpers.StringValueOrNull(config.Authentication.Password),
		}
	} else {
		m.Authentication = nil
	}

	// Handle encryption
	if config.Encryption != nil {
		m.Encryption = &EncryptionModel{
			MPPEBits: fwhelpers.Int64ValueOrNull(config.Encryption.MPPEBits),
			Required: types.BoolValue(config.Encryption.Required),
		}
	} else {
		m.Encryption = nil
	}

	// Handle IP pool
	if config.IPPool != nil {
		m.IPPool = &IPPoolModel{
			Start: types.StringValue(config.IPPool.Start),
			End:   types.StringValue(config.IPPool.End),
		}
	} else {
		m.IPPool = nil
	}
}
