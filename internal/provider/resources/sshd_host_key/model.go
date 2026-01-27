package sshd_host_key

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

// SSHDHostKeyModel describes the resource data model.
type SSHDHostKeyModel struct {
	ID          types.String `tfsdk:"id"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	Algorithm   types.String `tfsdk:"algorithm"`
}

// FromClient updates the Terraform model from a client.SSHHostKeyInfo.
func (m *SSHDHostKeyModel) FromClient(keyInfo *client.SSHHostKeyInfo) {
	m.ID = types.StringValue("sshd_host_key")
	m.Fingerprint = types.StringValue(keyInfo.Fingerprint)
	m.Algorithm = types.StringValue(keyInfo.Algorithm)
}
