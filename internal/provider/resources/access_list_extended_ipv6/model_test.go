package access_list_extended_ipv6

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func makePriorEntries(t *testing.T, mode string) types.List {
	t.Helper()
	objType := types.ObjectType{AttrTypes: EntryAttrTypes()}
	switch mode {
	case "null":
		return types.ListNull(objType)
	case "empty":
		return types.ListValueMust(objType, []attr.Value{})
	case "populated":
		return types.ListValueMust(objType, []attr.Value{
			types.ObjectValueMust(EntryAttrTypes(), map[string]attr.Value{
				"sequence":                  types.Int64Value(1),
				"ace_rule_action":           types.StringValue("permit"),
				"ace_rule_protocol":         types.StringValue("tcp"),
				"source_any":                types.BoolValue(true),
				"source_prefix":             types.StringNull(),
				"source_prefix_length":      types.Int64Null(),
				"source_port_equal":         types.StringNull(),
				"source_port_range":         types.StringNull(),
				"destination_any":           types.BoolValue(true),
				"destination_prefix":        types.StringNull(),
				"destination_prefix_length": types.Int64Null(),
				"destination_port_equal":    types.StringNull(),
				"destination_port_range":    types.StringNull(),
				"established":               types.BoolValue(false),
				"log":                       types.BoolValue(false),
			}),
		})
	}
	t.Fatalf("unknown mode: %s", mode)
	return types.List{}
}

func TestFromClient_Entries_NullPreservation(t *testing.T) {
	cases := []struct {
		name     string
		prior    string
		entries  []client.AccessListExtendedIPv6Entry
		wantNull bool
		wantSize int
	}{
		{"empty entries + prior null stays null", "null", nil, true, 0},
		{"empty entries + prior empty stays empty", "empty", nil, false, 0},
		{"empty entries + prior populated overwrites to empty", "populated", nil, false, 0},
		{"populated entries over prior null", "null", []client.AccessListExtendedIPv6Entry{{Sequence: 1, AceRuleAction: "permit", AceRuleProtocol: "tcp", SourceAny: true, DestinationAny: true}}, false, 1},
		{"populated entries over prior empty", "empty", []client.AccessListExtendedIPv6Entry{{Sequence: 2, AceRuleAction: "deny", AceRuleProtocol: "udp", SourceAny: true, DestinationAny: true}}, false, 1},
		{"populated entries over prior populated", "populated", []client.AccessListExtendedIPv6Entry{{Sequence: 3, AceRuleAction: "permit", AceRuleProtocol: "icmp", SourceAny: true, DestinationAny: true}}, false, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &AccessListExtendedIPv6Model{Entries: makePriorEntries(t, tc.prior)}
			m.FromClient(&client.AccessListExtendedIPv6{Name: "test", Entries: tc.entries})
			if got := m.Entries.IsNull(); got != tc.wantNull {
				t.Errorf("Entries.IsNull() = %v, want %v", got, tc.wantNull)
			}
			if !tc.wantNull {
				if size := len(m.Entries.Elements()); size != tc.wantSize {
					t.Errorf("len(Entries.Elements()) = %d, want %d", size, tc.wantSize)
				}
			}
		})
	}
}
