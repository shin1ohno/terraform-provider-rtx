package kron_policy

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// KronPolicyModel describes the resource data model.
type KronPolicyModel struct {
	Name         types.String `tfsdk:"name"`
	CommandLines types.List   `tfsdk:"command_lines"`
}

// ToClient converts the Terraform model to a client.KronPolicy.
func (m *KronPolicyModel) ToClient() client.KronPolicy {
	policy := client.KronPolicy{
		Name:     fwhelpers.GetStringValue(m.Name),
		Commands: getStringListValues(m.CommandLines),
	}

	// Ensure slice is not nil
	if policy.Commands == nil {
		policy.Commands = []string{}
	}

	return policy
}

// FromClient updates the Terraform model from a client.KronPolicy.
func (m *KronPolicyModel) FromClient(policy *client.KronPolicy) {
	m.Name = types.StringValue(policy.Name)
	m.CommandLines = stringSliceToList(policy.Commands)
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
	if slice == nil {
		return types.ListValueMust(types.StringType, []attr.Value{})
	}
	elements := make([]attr.Value, len(slice))
	for i, s := range slice {
		elements[i] = types.StringValue(s)
	}
	listVal, _ := types.ListValue(types.StringType, elements)
	return listVal
}
