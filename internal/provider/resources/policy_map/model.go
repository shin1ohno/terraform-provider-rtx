package policy_map

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// PolicyMapModel describes the resource data model.
type PolicyMapModel struct {
	Name    types.String `tfsdk:"name"`
	Classes types.List   `tfsdk:"class"`
}

// PolicyMapClassModel describes a class within a policy map.
type PolicyMapClassModel struct {
	Name             types.String `tfsdk:"name"`
	Priority         types.String `tfsdk:"priority"`
	BandwidthPercent types.Int64  `tfsdk:"bandwidth_percent"`
	PoliceCIR        types.Int64  `tfsdk:"police_cir"`
	QueueLimit       types.Int64  `tfsdk:"queue_limit"`
}

// PolicyMapClassAttrTypes returns the attribute types for PolicyMapClassModel.
func PolicyMapClassAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":              types.StringType,
		"priority":          types.StringType,
		"bandwidth_percent": types.Int64Type,
		"police_cir":        types.Int64Type,
		"queue_limit":       types.Int64Type,
	}
}

// ToClient converts the Terraform model to a client.PolicyMap.
func (m *PolicyMapModel) ToClient() client.PolicyMap {
	pm := client.PolicyMap{
		Name:    fwhelpers.GetStringValue(m.Name),
		Classes: []client.PolicyMapClass{},
	}

	if !m.Classes.IsNull() && !m.Classes.IsUnknown() {
		elements := m.Classes.Elements()
		pm.Classes = make([]client.PolicyMapClass, len(elements))

		for i, elem := range elements {
			objVal := elem.(types.Object)
			attrs := objVal.Attributes()

			class := client.PolicyMapClass{
				Name: getStringAttr(attrs, "name"),
			}

			if priority := getStringAttr(attrs, "priority"); priority != "" {
				class.Priority = priority
			}
			if bw := getInt64Attr(attrs, "bandwidth_percent"); bw != 0 {
				class.BandwidthPercent = bw
			}
			if cir := getInt64Attr(attrs, "police_cir"); cir != 0 {
				class.PoliceCIR = cir
			}
			if ql := getInt64Attr(attrs, "queue_limit"); ql != 0 {
				class.QueueLimit = ql
			}

			pm.Classes[i] = class
		}
	}

	return pm
}

// FromClient updates the Terraform model from a client.PolicyMap.
func (m *PolicyMapModel) FromClient(pm *client.PolicyMap) {
	m.Name = types.StringValue(pm.Name)

	if len(pm.Classes) > 0 {
		classElements := make([]attr.Value, len(pm.Classes))
		for i, class := range pm.Classes {
			attrs := map[string]attr.Value{
				"name":              types.StringValue(class.Name),
				"priority":          fwhelpers.StringValueOrNull(class.Priority),
				"bandwidth_percent": fwhelpers.Int64ValueOrNull(class.BandwidthPercent),
				"police_cir":        fwhelpers.Int64ValueOrNull(class.PoliceCIR),
				"queue_limit":       fwhelpers.Int64ValueOrNull(class.QueueLimit),
			}
			objVal, _ := types.ObjectValue(PolicyMapClassAttrTypes(), attrs)
			classElements[i] = objVal
		}
		listVal, _ := types.ListValue(types.ObjectType{AttrTypes: PolicyMapClassAttrTypes()}, classElements)
		m.Classes = listVal
	} else {
		m.Classes = types.ListValueMust(types.ObjectType{AttrTypes: PolicyMapClassAttrTypes()}, []attr.Value{})
	}
}

// Helper functions

func getStringAttr(attrs map[string]attr.Value, key string) string {
	if v, ok := attrs[key]; ok {
		if strVal, ok := v.(types.String); ok && !strVal.IsNull() && !strVal.IsUnknown() {
			return strVal.ValueString()
		}
	}
	return ""
}

func getInt64Attr(attrs map[string]attr.Value, key string) int {
	if v, ok := attrs[key]; ok {
		if intVal, ok := v.(types.Int64); ok && !intVal.IsNull() && !intVal.IsUnknown() {
			return int(intVal.ValueInt64())
		}
	}
	return 0
}
