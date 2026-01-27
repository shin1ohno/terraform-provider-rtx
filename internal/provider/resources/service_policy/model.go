package service_policy

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// ServicePolicyModel describes the resource data model.
type ServicePolicyModel struct {
	ID        types.String `tfsdk:"id"`
	Interface types.String `tfsdk:"interface"`
	Direction types.String `tfsdk:"direction"`
	PolicyMap types.String `tfsdk:"policy_map"`
}

// ToClient converts the Terraform model to a client.ServicePolicy.
func (m *ServicePolicyModel) ToClient() client.ServicePolicy {
	return client.ServicePolicy{
		Interface: fwhelpers.GetStringValue(m.Interface),
		Direction: fwhelpers.GetStringValue(m.Direction),
		PolicyMap: fwhelpers.GetStringValue(m.PolicyMap),
	}
}

// FromClient updates the Terraform model from a client.ServicePolicy.
func (m *ServicePolicyModel) FromClient(sp *client.ServicePolicy) {
	m.Interface = types.StringValue(sp.Interface)
	m.Direction = types.StringValue(sp.Direction)
	m.PolicyMap = types.StringValue(sp.PolicyMap)
}
