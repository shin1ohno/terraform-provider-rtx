package ipv6_prefix

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// IPv6PrefixModel describes the resource data model.
type IPv6PrefixModel struct {
	PrefixID     types.Int64  `tfsdk:"prefix_id"`
	Prefix       types.String `tfsdk:"prefix"`
	PrefixLength types.Int64  `tfsdk:"prefix_length"`
	Source       types.String `tfsdk:"source"`
	Interface    types.String `tfsdk:"interface"`
}

// ToClient converts the Terraform model to a client.IPv6Prefix.
func (m *IPv6PrefixModel) ToClient() client.IPv6Prefix {
	return client.IPv6Prefix{
		ID:           fwhelpers.GetInt64Value(m.PrefixID),
		Prefix:       fwhelpers.GetStringValue(m.Prefix),
		PrefixLength: fwhelpers.GetInt64Value(m.PrefixLength),
		Source:       fwhelpers.GetStringValue(m.Source),
		Interface:    fwhelpers.GetStringValue(m.Interface),
	}
}

// FromClient updates the Terraform model from a client.IPv6Prefix.
func (m *IPv6PrefixModel) FromClient(prefix *client.IPv6Prefix) {
	m.PrefixID = types.Int64Value(int64(prefix.ID))
	m.Prefix = fwhelpers.StringValueOrNull(prefix.Prefix)
	m.PrefixLength = types.Int64Value(int64(prefix.PrefixLength))
	m.Source = types.StringValue(prefix.Source)
	m.Interface = fwhelpers.StringValueOrNull(prefix.Interface)
}
