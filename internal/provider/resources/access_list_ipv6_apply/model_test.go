package access_list_ipv6_apply

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func makePriorSequencesList(t *testing.T, mode string) types.List {
	t.Helper()
	switch mode {
	case "null":
		return types.ListNull(types.Int64Type)
	case "empty":
		return types.ListValueMust(types.Int64Type, []attr.Value{})
	case "populated":
		return types.ListValueMust(types.Int64Type, []attr.Value{types.Int64Value(1)})
	}
	t.Fatalf("unknown mode: %s", mode)
	return types.List{}
}

func TestSetSequencesFromInts_NullPreservation(t *testing.T) {
	cases := []struct {
		name     string
		prior    string
		ids      []int
		wantNull bool
		wantSize int
	}{
		{"empty ids + prior null stays null", "null", nil, true, 0},
		{"empty ids + prior empty stays empty", "empty", nil, false, 0},
		{"empty ids + prior populated overwrites to empty", "populated", nil, false, 0},
		{"populated ids over prior null", "null", []int{10, 20}, false, 2},
		{"populated ids over prior empty", "empty", []int{10}, false, 1},
		{"populated ids over prior populated", "populated", []int{30}, false, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &AccessListIPv6ApplyModel{Sequences: makePriorSequencesList(t, tc.prior)}
			m.SetSequencesFromInts(tc.ids)
			if got := m.Sequences.IsNull(); got != tc.wantNull {
				t.Errorf("Sequences.IsNull() = %v, want %v", got, tc.wantNull)
			}
			if !tc.wantNull {
				if size := len(m.Sequences.Elements()); size != tc.wantSize {
					t.Errorf("len(Sequences.Elements()) = %d, want %d", size, tc.wantSize)
				}
			}
		})
	}
}
