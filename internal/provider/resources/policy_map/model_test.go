package policy_map

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func makePriorClasses(t *testing.T, mode string) types.List {
	t.Helper()
	objType := types.ObjectType{AttrTypes: PolicyMapClassAttrTypes()}
	switch mode {
	case "null":
		return types.ListNull(objType)
	case "empty":
		return types.ListValueMust(objType, []attr.Value{})
	case "populated":
		return types.ListValueMust(objType, []attr.Value{
			types.ObjectValueMust(PolicyMapClassAttrTypes(), map[string]attr.Value{
				"name":              types.StringValue("voice"),
				"priority":          types.StringValue("high"),
				"bandwidth_percent": types.Int64Null(),
				"police_cir":        types.Int64Null(),
				"queue_limit":       types.Int64Null(),
			}),
		})
	}
	t.Fatalf("unknown mode: %s", mode)
	return types.List{}
}

func TestFromClient_Classes_NullPreservation(t *testing.T) {
	cases := []struct {
		name     string
		prior    string
		classes  []client.PolicyMapClass
		wantNull bool
		wantSize int
	}{
		{"empty + prior null stays null", "null", nil, true, 0},
		{"empty + prior empty stays empty", "empty", nil, false, 0},
		{"empty + prior populated overwrites to empty", "populated", nil, false, 0},
		{"populated over prior null", "null", []client.PolicyMapClass{{Name: "voice", Priority: "high"}}, false, 1},
		{"populated over prior empty", "empty", []client.PolicyMapClass{{Name: "data"}}, false, 1},
		{"populated over prior populated", "populated", []client.PolicyMapClass{{Name: "video", Priority: "normal"}}, false, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &PolicyMapModel{Classes: makePriorClasses(t, tc.prior)}
			m.FromClient(&client.PolicyMap{Name: "qos", Classes: tc.classes})
			if got := m.Classes.IsNull(); got != tc.wantNull {
				t.Errorf("Classes.IsNull() = %v, want %v", got, tc.wantNull)
			}
			if !tc.wantNull && len(m.Classes.Elements()) != tc.wantSize {
				t.Errorf("len(Classes.Elements()) = %d, want %d", len(m.Classes.Elements()), tc.wantSize)
			}
		})
	}
}
