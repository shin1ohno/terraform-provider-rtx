package sshd

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// SSHDModel describes the resource data model.
type SSHDModel struct {
	ID         types.String `tfsdk:"id"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	Hosts      types.List   `tfsdk:"hosts"`
	HostKey    types.String `tfsdk:"host_key"`
	AuthMethod types.String `tfsdk:"auth_method"`
}

// ToClient converts the Terraform model to a client.SSHDConfig.
func (m *SSHDModel) ToClient() client.SSHDConfig {
	config := client.SSHDConfig{
		Enabled:    fwhelpers.GetBoolValue(m.Enabled),
		Hosts:      getStringListValues(m.Hosts),
		AuthMethod: fwhelpers.GetStringValue(m.AuthMethod),
	}

	// Ensure Hosts is not nil
	if config.Hosts == nil {
		config.Hosts = []string{}
	}

	return config
}

// FromClient updates the Terraform model from a client.SSHDConfig.
func (m *SSHDModel) FromClient(config *client.SSHDConfig) {
	m.Enabled = types.BoolValue(config.Enabled)
	m.AuthMethod = types.StringValue(config.AuthMethod)

	// Handle host key
	if config.HostKey != "" {
		m.HostKey = types.StringValue(config.HostKey)
	} else {
		m.HostKey = types.StringNull()
	}

	// Handle hosts list
	if len(config.Hosts) > 0 {
		m.Hosts = stringSliceToList(config.Hosts)
	} else {
		m.Hosts = types.ListValueMust(types.StringType, []attr.Value{})
	}
}

// Helper functions

func getStringListValues(list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var result []string
	elements := list.Elements()
	for _, elem := range elements {
		if strVal, ok := elem.(types.String); ok {
			result = append(result, strVal.ValueString())
		}
	}
	return result
}

func stringSliceToList(slice []string) types.List {
	elements := make([]attr.Value, len(slice))
	for i, s := range slice {
		elements[i] = types.StringValue(s)
	}
	listVal, _ := types.ListValue(types.StringType, elements)
	return listVal
}
