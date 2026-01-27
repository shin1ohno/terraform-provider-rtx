package sftpd

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

// SFTPDModel describes the resource data model.
type SFTPDModel struct {
	ID    types.String `tfsdk:"id"`
	Hosts types.List   `tfsdk:"hosts"`
}

// ToClient converts the Terraform model to a client.SFTPDConfig.
func (m *SFTPDModel) ToClient() client.SFTPDConfig {
	config := client.SFTPDConfig{
		Hosts: []string{},
	}

	if !m.Hosts.IsNull() && !m.Hosts.IsUnknown() {
		elements := m.Hosts.Elements()
		hosts := make([]string, len(elements))
		for i, elem := range elements {
			if strVal, ok := elem.(types.String); ok {
				hosts[i] = strVal.ValueString()
			}
		}
		config.Hosts = hosts
	}

	return config
}

// FromClient updates the Terraform model from a client.SFTPDConfig.
func (m *SFTPDModel) FromClient(config *client.SFTPDConfig) {
	m.ID = types.StringValue("sftpd")

	if len(config.Hosts) > 0 {
		elements := make([]attr.Value, len(config.Hosts))
		for i, h := range config.Hosts {
			elements[i] = types.StringValue(h)
		}
		m.Hosts, _ = types.ListValue(types.StringType, elements)
	} else {
		m.Hosts, _ = types.ListValue(types.StringType, []attr.Value{})
	}
}
