package tunnel

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// convertListToIntSlice converts a types.List to []int
func convertListToIntSlice(list types.List) []int {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	elements := list.Elements()
	if len(elements) == 0 {
		return nil
	}

	result := make([]int, 0, len(elements))
	for _, elem := range elements {
		if int64Val, ok := elem.(types.Int64); ok && !int64Val.IsNull() && !int64Val.IsUnknown() {
			result = append(result, int(int64Val.ValueInt64()))
		}
	}

	return result
}

// convertIntSliceToList converts []int to types.List
func convertIntSliceToList(ints []int) types.List {
	if len(ints) == 0 {
		return types.ListNull(types.Int64Type)
	}

	elements := make([]attr.Value, len(ints))
	for i, v := range ints {
		elements[i] = types.Int64Value(int64(v))
	}

	list, _ := types.ListValue(types.Int64Type, elements)
	return list
}

// Ensure context is used (for future use with diagnostics)
var _ = context.Background

// TunnelModel describes the unified tunnel resource data model.
type TunnelModel struct {
	TunnelID        types.Int64       `tfsdk:"tunnel_id"`
	Encapsulation   types.String      `tfsdk:"encapsulation"`
	Enabled         types.Bool        `tfsdk:"enabled"`
	Name            types.String      `tfsdk:"name"`
	TunnelInterface types.String      `tfsdk:"tunnel_interface"`
	IPsec           *TunnelIPsecModel `tfsdk:"ipsec"`
	L2TP            *TunnelL2TPModel  `tfsdk:"l2tp"`
}

// TunnelIPsecModel describes the IPsec nested block.
type TunnelIPsecModel struct {
	IPsecTunnelID   types.Int64             `tfsdk:"ipsec_tunnel_id"`
	LocalAddress    types.String            `tfsdk:"local_address"`
	RemoteAddress   types.String            `tfsdk:"remote_address"`
	PreSharedKey    types.String            `tfsdk:"pre_shared_key"`
	SecureFilterIn  types.List              `tfsdk:"secure_filter_in"`
	SecureFilterOut types.List              `tfsdk:"secure_filter_out"`
	TCPMSSLimit     types.String            `tfsdk:"tcp_mss_limit"`
	IPsecTransform  *IPsecTransformModel    `tfsdk:"ipsec_transform"`
	Keepalive       *TunnelIPsecKeepaliveModel `tfsdk:"keepalive"`
}

// TunnelIPsecKeepaliveModel describes the IPsec keepalive nested block.
type TunnelIPsecKeepaliveModel struct {
	Enabled  types.Bool   `tfsdk:"enabled"`
	Mode     types.String `tfsdk:"mode"`
	Interval types.Int64  `tfsdk:"interval"`
	Retry    types.Int64  `tfsdk:"retry"`
}

// IPsecTransformModel describes the IPsec transform nested block.
type IPsecTransformModel struct {
	Protocol         types.String `tfsdk:"protocol"`
	EncryptionAES256 types.Bool   `tfsdk:"encryption_aes256"`
	EncryptionAES128 types.Bool   `tfsdk:"encryption_aes128"`
	Encryption3DES   types.Bool   `tfsdk:"encryption_3des"`
	IntegritySHA256  types.Bool   `tfsdk:"integrity_sha256"`
	IntegritySHA1    types.Bool   `tfsdk:"integrity_sha1"`
	IntegrityMD5     types.Bool   `tfsdk:"integrity_md5"`
}

// TunnelL2TPModel describes the L2TP nested block.
type TunnelL2TPModel struct {
	Hostname       types.String            `tfsdk:"hostname"`
	LocalRouterID  types.String            `tfsdk:"local_router_id"`
	RemoteRouterID types.String            `tfsdk:"remote_router_id"`
	RemoteEndID    types.String            `tfsdk:"remote_end_id"`
	AlwaysOn       types.Bool              `tfsdk:"always_on"`
	TunnelAuth     *TunnelL2TPAuthModel    `tfsdk:"tunnel_auth"`
	Keepalive      *TunnelL2TPKeepaliveModel `tfsdk:"keepalive"`
}

// TunnelL2TPAuthModel describes the L2TP tunnel auth nested block.
type TunnelL2TPAuthModel struct {
	Enabled  types.Bool   `tfsdk:"enabled"`
	Password types.String `tfsdk:"password"`
}

// TunnelL2TPKeepaliveModel describes the L2TP keepalive nested block.
type TunnelL2TPKeepaliveModel struct {
	Enabled  types.Bool  `tfsdk:"enabled"`
	Interval types.Int64 `tfsdk:"interval"`
	Retry    types.Int64 `tfsdk:"retry"`
}

// ToClient converts the Terraform model to a client.Tunnel.
func (m *TunnelModel) ToClient() client.Tunnel {
	tunnel := client.Tunnel{
		ID:            int(m.TunnelID.ValueInt64()),
		Encapsulation: fwhelpers.GetStringValue(m.Encapsulation),
		Enabled:       fwhelpers.GetBoolValueWithDefault(m.Enabled, true),
		Name:          fwhelpers.GetStringValue(m.Name),
	}

	// Handle IPsec block
	if m.IPsec != nil {
		ipsecTunnelID := int(m.IPsec.IPsecTunnelID.ValueInt64())
		if ipsecTunnelID == 0 {
			ipsecTunnelID = tunnel.ID // Default to tunnel_id
		}

		tunnel.IPsec = &client.TunnelIPsec{
			IPsecTunnelID:   ipsecTunnelID,
			LocalAddress:    fwhelpers.GetStringValue(m.IPsec.LocalAddress),
			RemoteAddress:   fwhelpers.GetStringValue(m.IPsec.RemoteAddress),
			PreSharedKey:    fwhelpers.GetStringValue(m.IPsec.PreSharedKey),
			SecureFilterIn:  convertListToIntSlice(m.IPsec.SecureFilterIn),
			SecureFilterOut: convertListToIntSlice(m.IPsec.SecureFilterOut),
			TCPMSSLimit:     fwhelpers.GetStringValue(m.IPsec.TCPMSSLimit),
		}

		// Handle IPsec transform
		if m.IPsec.IPsecTransform != nil {
			tunnel.IPsec.Transform = client.IPsecTransform{
				Protocol:         fwhelpers.GetStringValueWithDefault(m.IPsec.IPsecTransform.Protocol, "esp"),
				EncryptionAES256: fwhelpers.GetBoolValue(m.IPsec.IPsecTransform.EncryptionAES256),
				EncryptionAES128: fwhelpers.GetBoolValue(m.IPsec.IPsecTransform.EncryptionAES128),
				Encryption3DES:   fwhelpers.GetBoolValue(m.IPsec.IPsecTransform.Encryption3DES),
				IntegritySHA256:  fwhelpers.GetBoolValue(m.IPsec.IPsecTransform.IntegritySHA256),
				IntegritySHA1:    fwhelpers.GetBoolValue(m.IPsec.IPsecTransform.IntegritySHA1),
				IntegrityMD5:     fwhelpers.GetBoolValue(m.IPsec.IPsecTransform.IntegrityMD5),
			}
		}

		// Handle IPsec keepalive
		if m.IPsec.Keepalive != nil {
			tunnel.IPsec.Keepalive = &client.TunnelIPsecKeepalive{
				Enabled:  fwhelpers.GetBoolValue(m.IPsec.Keepalive.Enabled),
				Mode:     fwhelpers.GetStringValueWithDefault(m.IPsec.Keepalive.Mode, "dpd"),
				Interval: fwhelpers.GetInt64Value(m.IPsec.Keepalive.Interval),
				Retry:    fwhelpers.GetInt64Value(m.IPsec.Keepalive.Retry),
			}
		}
	}

	// Handle L2TP block
	if m.L2TP != nil {
		tunnel.L2TP = &client.TunnelL2TP{
			Hostname:       fwhelpers.GetStringValue(m.L2TP.Hostname),
			LocalRouterID:  fwhelpers.GetStringValue(m.L2TP.LocalRouterID),
			RemoteRouterID: fwhelpers.GetStringValue(m.L2TP.RemoteRouterID),
			RemoteEndID:    fwhelpers.GetStringValue(m.L2TP.RemoteEndID),
			AlwaysOn:       fwhelpers.GetBoolValue(m.L2TP.AlwaysOn),
		}

		// Handle L2TP tunnel auth
		if m.L2TP.TunnelAuth != nil {
			tunnel.L2TP.TunnelAuth = &client.TunnelL2TPAuth{
				Enabled:  fwhelpers.GetBoolValue(m.L2TP.TunnelAuth.Enabled),
				Password: fwhelpers.GetStringValue(m.L2TP.TunnelAuth.Password),
			}
		}

		// Handle L2TP keepalive
		if m.L2TP.Keepalive != nil {
			tunnel.L2TP.Keepalive = &client.TunnelL2TPKeepalive{
				Enabled:  fwhelpers.GetBoolValue(m.L2TP.Keepalive.Enabled),
				Interval: fwhelpers.GetInt64Value(m.L2TP.Keepalive.Interval),
				Retry:    fwhelpers.GetInt64Value(m.L2TP.Keepalive.Retry),
			}
		}
	}

	return tunnel
}

// FromClient updates the Terraform model from a client.Tunnel.
func (m *TunnelModel) FromClient(tunnel *client.Tunnel) {
	m.TunnelID = types.Int64Value(int64(tunnel.ID))
	m.Encapsulation = types.StringValue(tunnel.Encapsulation)
	m.Enabled = types.BoolValue(tunnel.Enabled)
	m.Name = fwhelpers.StringValueOrNull(tunnel.Name)
	m.TunnelInterface = types.StringValue(fmt.Sprintf("tunnel%d", tunnel.ID))

	// Handle IPsec block
	if tunnel.IPsec != nil {
		if m.IPsec == nil {
			m.IPsec = &TunnelIPsecModel{}
		}
		m.IPsec.IPsecTunnelID = types.Int64Value(int64(tunnel.IPsec.IPsecTunnelID))
		m.IPsec.LocalAddress = fwhelpers.StringValueOrNull(tunnel.IPsec.LocalAddress)
		m.IPsec.RemoteAddress = fwhelpers.StringValueOrNull(tunnel.IPsec.RemoteAddress)
		// Note: pre_shared_key is WriteOnly, so we don't read it back
		m.IPsec.SecureFilterIn = convertIntSliceToList(tunnel.IPsec.SecureFilterIn)
		m.IPsec.SecureFilterOut = convertIntSliceToList(tunnel.IPsec.SecureFilterOut)
		m.IPsec.TCPMSSLimit = fwhelpers.StringValueOrNull(tunnel.IPsec.TCPMSSLimit)

		// Handle IPsec transform
		if m.IPsec.IPsecTransform == nil {
			m.IPsec.IPsecTransform = &IPsecTransformModel{}
		}
		m.IPsec.IPsecTransform.Protocol = fwhelpers.StringValueOrNull(tunnel.IPsec.Transform.Protocol)
		m.IPsec.IPsecTransform.EncryptionAES256 = types.BoolValue(tunnel.IPsec.Transform.EncryptionAES256)
		m.IPsec.IPsecTransform.EncryptionAES128 = types.BoolValue(tunnel.IPsec.Transform.EncryptionAES128)
		m.IPsec.IPsecTransform.Encryption3DES = types.BoolValue(tunnel.IPsec.Transform.Encryption3DES)
		m.IPsec.IPsecTransform.IntegritySHA256 = types.BoolValue(tunnel.IPsec.Transform.IntegritySHA256)
		m.IPsec.IPsecTransform.IntegritySHA1 = types.BoolValue(tunnel.IPsec.Transform.IntegritySHA1)
		m.IPsec.IPsecTransform.IntegrityMD5 = types.BoolValue(tunnel.IPsec.Transform.IntegrityMD5)

		// Handle IPsec keepalive
		if tunnel.IPsec.Keepalive != nil {
			if m.IPsec.Keepalive == nil {
				m.IPsec.Keepalive = &TunnelIPsecKeepaliveModel{}
			}
			m.IPsec.Keepalive.Enabled = types.BoolValue(tunnel.IPsec.Keepalive.Enabled)
			m.IPsec.Keepalive.Mode = fwhelpers.StringValueOrNull(tunnel.IPsec.Keepalive.Mode)
			m.IPsec.Keepalive.Interval = fwhelpers.Int64ValueOrNull(tunnel.IPsec.Keepalive.Interval)
			m.IPsec.Keepalive.Retry = fwhelpers.Int64ValueOrNull(tunnel.IPsec.Keepalive.Retry)
		}
	}

	// Handle L2TP block
	if tunnel.L2TP != nil {
		if m.L2TP == nil {
			m.L2TP = &TunnelL2TPModel{}
		}
		m.L2TP.Hostname = fwhelpers.StringValueOrNull(tunnel.L2TP.Hostname)
		m.L2TP.LocalRouterID = fwhelpers.StringValueOrNull(tunnel.L2TP.LocalRouterID)
		m.L2TP.RemoteRouterID = fwhelpers.StringValueOrNull(tunnel.L2TP.RemoteRouterID)
		m.L2TP.RemoteEndID = fwhelpers.StringValueOrNull(tunnel.L2TP.RemoteEndID)
		m.L2TP.AlwaysOn = types.BoolValue(tunnel.L2TP.AlwaysOn)

		// Handle L2TP tunnel auth
		if tunnel.L2TP.TunnelAuth != nil {
			if m.L2TP.TunnelAuth == nil {
				m.L2TP.TunnelAuth = &TunnelL2TPAuthModel{}
			}
			m.L2TP.TunnelAuth.Enabled = types.BoolValue(tunnel.L2TP.TunnelAuth.Enabled)
			// Note: password is WriteOnly, so we don't read it back
		}

		// Handle L2TP keepalive
		if tunnel.L2TP.Keepalive != nil {
			if m.L2TP.Keepalive == nil {
				m.L2TP.Keepalive = &TunnelL2TPKeepaliveModel{}
			}
			m.L2TP.Keepalive.Enabled = types.BoolValue(tunnel.L2TP.Keepalive.Enabled)
			m.L2TP.Keepalive.Interval = fwhelpers.Int64ValueOrNull(tunnel.L2TP.Keepalive.Interval)
			m.L2TP.Keepalive.Retry = fwhelpers.Int64ValueOrNull(tunnel.L2TP.Keepalive.Retry)
		}
	}
}

// ID returns the resource identifier.
func (m *TunnelModel) ID() string {
	return fmt.Sprintf("%d", m.TunnelID.ValueInt64())
}
