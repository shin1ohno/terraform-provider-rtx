package ipsec_tunnel

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// IPsecTunnelModel describes the resource data model.
type IPsecTunnelModel struct {
	TunnelID        types.Int64          `tfsdk:"tunnel_id"`
	Name            types.String         `tfsdk:"name"`
	LocalAddress    types.String         `tfsdk:"local_address"`
	RemoteAddress   types.String         `tfsdk:"remote_address"`
	PreSharedKey    types.String         `tfsdk:"pre_shared_key"`
	LocalNetwork    types.String         `tfsdk:"local_network"`
	RemoteNetwork   types.String         `tfsdk:"remote_network"`
	DPDEnabled      types.Bool           `tfsdk:"dpd_enabled"`
	DPDInterval     types.Int64          `tfsdk:"dpd_interval"`
	DPDRetry        types.Int64          `tfsdk:"dpd_retry"`
	KeepaliveMode   types.String         `tfsdk:"keepalive_mode"`
	Enabled         types.Bool           `tfsdk:"enabled"`
	TunnelInterface types.String         `tfsdk:"tunnel_interface"`
	IKEv2Proposal   *IKEv2ProposalModel  `tfsdk:"ikev2_proposal"`
	IPsecTransform  *IPsecTransformModel `tfsdk:"ipsec_transform"`
}

// IKEv2ProposalModel describes the IKEv2 proposal nested block.
type IKEv2ProposalModel struct {
	EncryptionAES256 types.Bool  `tfsdk:"encryption_aes256"`
	EncryptionAES128 types.Bool  `tfsdk:"encryption_aes128"`
	Encryption3DES   types.Bool  `tfsdk:"encryption_3des"`
	IntegritySHA256  types.Bool  `tfsdk:"integrity_sha256"`
	IntegritySHA1    types.Bool  `tfsdk:"integrity_sha1"`
	IntegrityMD5     types.Bool  `tfsdk:"integrity_md5"`
	GroupFourteen    types.Bool  `tfsdk:"group_fourteen"`
	GroupFive        types.Bool  `tfsdk:"group_five"`
	GroupTwo         types.Bool  `tfsdk:"group_two"`
	LifetimeSeconds  types.Int64 `tfsdk:"lifetime_seconds"`
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
	PFSGroupFourteen types.Bool   `tfsdk:"pfs_group_fourteen"`
	PFSGroupFive     types.Bool   `tfsdk:"pfs_group_five"`
	PFSGroupTwo      types.Bool   `tfsdk:"pfs_group_two"`
	LifetimeSeconds  types.Int64  `tfsdk:"lifetime_seconds"`
}

// ToClient converts the Terraform model to a client.IPsecTunnel.
func (m *IPsecTunnelModel) ToClient() client.IPsecTunnel {
	tunnel := client.IPsecTunnel{
		ID:            int(m.TunnelID.ValueInt64()),
		Name:          fwhelpers.GetStringValue(m.Name),
		LocalAddress:  fwhelpers.GetStringValue(m.LocalAddress),
		RemoteAddress: fwhelpers.GetStringValue(m.RemoteAddress),
		PreSharedKey:  fwhelpers.GetStringValue(m.PreSharedKey),
		LocalNetwork:  fwhelpers.GetStringValue(m.LocalNetwork),
		RemoteNetwork: fwhelpers.GetStringValue(m.RemoteNetwork),
		DPDEnabled:    fwhelpers.GetBoolValue(m.DPDEnabled),
		DPDInterval:   fwhelpers.GetInt64Value(m.DPDInterval),
		DPDRetry:      fwhelpers.GetInt64Value(m.DPDRetry),
		KeepaliveMode: fwhelpers.GetStringValue(m.KeepaliveMode),
		Enabled:       fwhelpers.GetBoolValueWithDefault(m.Enabled, true),
	}

	// Handle IKEv2 proposal
	if m.IKEv2Proposal != nil {
		tunnel.IKEv2Proposal = client.IKEv2Proposal{
			EncryptionAES256: fwhelpers.GetBoolValue(m.IKEv2Proposal.EncryptionAES256),
			EncryptionAES128: fwhelpers.GetBoolValue(m.IKEv2Proposal.EncryptionAES128),
			Encryption3DES:   fwhelpers.GetBoolValue(m.IKEv2Proposal.Encryption3DES),
			IntegritySHA256:  fwhelpers.GetBoolValue(m.IKEv2Proposal.IntegritySHA256),
			IntegritySHA1:    fwhelpers.GetBoolValue(m.IKEv2Proposal.IntegritySHA1),
			IntegrityMD5:     fwhelpers.GetBoolValue(m.IKEv2Proposal.IntegrityMD5),
			GroupFourteen:    fwhelpers.GetBoolValue(m.IKEv2Proposal.GroupFourteen),
			GroupFive:        fwhelpers.GetBoolValue(m.IKEv2Proposal.GroupFive),
			GroupTwo:         fwhelpers.GetBoolValue(m.IKEv2Proposal.GroupTwo),
			LifetimeSeconds:  fwhelpers.GetInt64Value(m.IKEv2Proposal.LifetimeSeconds),
		}
	}

	// Handle IPsec transform
	if m.IPsecTransform != nil {
		tunnel.IPsecTransform = client.IPsecTransform{
			Protocol:         fwhelpers.GetStringValue(m.IPsecTransform.Protocol),
			EncryptionAES256: fwhelpers.GetBoolValue(m.IPsecTransform.EncryptionAES256),
			EncryptionAES128: fwhelpers.GetBoolValue(m.IPsecTransform.EncryptionAES128),
			Encryption3DES:   fwhelpers.GetBoolValue(m.IPsecTransform.Encryption3DES),
			IntegritySHA256:  fwhelpers.GetBoolValue(m.IPsecTransform.IntegritySHA256),
			IntegritySHA1:    fwhelpers.GetBoolValue(m.IPsecTransform.IntegritySHA1),
			IntegrityMD5:     fwhelpers.GetBoolValue(m.IPsecTransform.IntegrityMD5),
			PFSGroupFourteen: fwhelpers.GetBoolValue(m.IPsecTransform.PFSGroupFourteen),
			PFSGroupFive:     fwhelpers.GetBoolValue(m.IPsecTransform.PFSGroupFive),
			PFSGroupTwo:      fwhelpers.GetBoolValue(m.IPsecTransform.PFSGroupTwo),
			LifetimeSeconds:  fwhelpers.GetInt64Value(m.IPsecTransform.LifetimeSeconds),
		}
	}

	return tunnel
}

// FromClient updates the Terraform model from a client.IPsecTunnel.
func (m *IPsecTunnelModel) FromClient(tunnel *client.IPsecTunnel) {
	m.TunnelID = types.Int64Value(int64(tunnel.ID))
	m.Name = fwhelpers.StringValueOrNull(tunnel.Name)
	m.LocalAddress = fwhelpers.StringValueOrNull(tunnel.LocalAddress)
	m.RemoteAddress = fwhelpers.StringValueOrNull(tunnel.RemoteAddress)
	// Note: pre_shared_key is WriteOnly, so we don't read it back
	m.LocalNetwork = fwhelpers.StringValueOrNull(tunnel.LocalNetwork)
	m.RemoteNetwork = fwhelpers.StringValueOrNull(tunnel.RemoteNetwork)

	// DPD values: config-only attributes
	// Router may not return correct values (DPD keepalive status parsing is unreliable),
	// so always preserve existing state values. Only set initial values if currently unknown/null.
	if m.DPDEnabled.IsUnknown() || m.DPDEnabled.IsNull() {
		m.DPDEnabled = types.BoolValue(tunnel.DPDEnabled)
	}
	// else: preserve existing m.DPDEnabled

	if m.DPDInterval.IsUnknown() || m.DPDInterval.IsNull() {
		m.DPDInterval = fwhelpers.Int64ValueOrNull(tunnel.DPDInterval)
	}
	// else: preserve existing m.DPDInterval

	if m.DPDRetry.IsUnknown() || m.DPDRetry.IsNull() {
		m.DPDRetry = fwhelpers.Int64ValueOrNull(tunnel.DPDRetry)
	}
	// else: preserve existing m.DPDRetry

	// KeepaliveMode: preserve existing state value if set
	if m.KeepaliveMode.IsUnknown() || m.KeepaliveMode.IsNull() {
		m.KeepaliveMode = fwhelpers.StringValueOrNull(tunnel.KeepaliveMode)
	}

	m.Enabled = types.BoolValue(tunnel.Enabled)
	m.TunnelInterface = types.StringValue(fmt.Sprintf("tunnel%d", tunnel.ID))

	// Update IKEv2 proposal - only update if the block already exists in state
	// This preserves the user's intent when they don't specify ikev2_proposal
	if m.IKEv2Proposal != nil {
		m.IKEv2Proposal.EncryptionAES256 = types.BoolValue(tunnel.IKEv2Proposal.EncryptionAES256)
		m.IKEv2Proposal.EncryptionAES128 = types.BoolValue(tunnel.IKEv2Proposal.EncryptionAES128)
		m.IKEv2Proposal.Encryption3DES = types.BoolValue(tunnel.IKEv2Proposal.Encryption3DES)
		m.IKEv2Proposal.IntegritySHA256 = types.BoolValue(tunnel.IKEv2Proposal.IntegritySHA256)
		m.IKEv2Proposal.IntegritySHA1 = types.BoolValue(tunnel.IKEv2Proposal.IntegritySHA1)
		m.IKEv2Proposal.IntegrityMD5 = types.BoolValue(tunnel.IKEv2Proposal.IntegrityMD5)
		m.IKEv2Proposal.GroupFourteen = types.BoolValue(tunnel.IKEv2Proposal.GroupFourteen)
		m.IKEv2Proposal.GroupFive = types.BoolValue(tunnel.IKEv2Proposal.GroupFive)
		m.IKEv2Proposal.GroupTwo = types.BoolValue(tunnel.IKEv2Proposal.GroupTwo)
		m.IKEv2Proposal.LifetimeSeconds = fwhelpers.Int64ValueOrNull(tunnel.IKEv2Proposal.LifetimeSeconds)
	}
	// If m.IKEv2Proposal is nil, leave it nil - user didn't specify the block

	// Update IPsec transform
	if m.IPsecTransform == nil {
		m.IPsecTransform = &IPsecTransformModel{}
	}
	m.IPsecTransform.Protocol = fwhelpers.StringValueOrNull(tunnel.IPsecTransform.Protocol)
	m.IPsecTransform.EncryptionAES256 = types.BoolValue(tunnel.IPsecTransform.EncryptionAES256)
	m.IPsecTransform.EncryptionAES128 = types.BoolValue(tunnel.IPsecTransform.EncryptionAES128)
	m.IPsecTransform.Encryption3DES = types.BoolValue(tunnel.IPsecTransform.Encryption3DES)
	m.IPsecTransform.IntegritySHA256 = types.BoolValue(tunnel.IPsecTransform.IntegritySHA256)
	m.IPsecTransform.IntegritySHA1 = types.BoolValue(tunnel.IPsecTransform.IntegritySHA1)
	m.IPsecTransform.IntegrityMD5 = types.BoolValue(tunnel.IPsecTransform.IntegrityMD5)
	m.IPsecTransform.PFSGroupFourteen = types.BoolValue(tunnel.IPsecTransform.PFSGroupFourteen)
	m.IPsecTransform.PFSGroupFive = types.BoolValue(tunnel.IPsecTransform.PFSGroupFive)
	m.IPsecTransform.PFSGroupTwo = types.BoolValue(tunnel.IPsecTransform.PFSGroupTwo)
	m.IPsecTransform.LifetimeSeconds = fwhelpers.Int64ValueOrNull(tunnel.IPsecTransform.LifetimeSeconds)
}
