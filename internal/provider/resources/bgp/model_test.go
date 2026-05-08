package bgp

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func makePriorNeighbors(t *testing.T, mode string) types.List {
	t.Helper()
	objType := types.ObjectType{AttrTypes: NeighborAttrTypes()}
	switch mode {
	case "null":
		return types.ListNull(objType)
	case "empty":
		return types.ListValueMust(objType, []attr.Value{})
	case "populated":
		return types.ListValueMust(objType, []attr.Value{
			types.ObjectValueMust(NeighborAttrTypes(), map[string]attr.Value{
				"index":         types.Int64Value(1),
				"ip":            types.StringValue("10.0.0.1"),
				"remote_as":     types.StringValue("65001"),
				"hold_time":     types.Int64Null(),
				"keepalive":     types.Int64Null(),
				"multihop":      types.Int64Null(),
				"password":      types.StringNull(),
				"local_address": types.StringNull(),
			}),
		})
	}
	t.Fatalf("unknown mode: %s", mode)
	return types.List{}
}

func makePriorNetworks(t *testing.T, mode string) types.List {
	t.Helper()
	objType := types.ObjectType{AttrTypes: NetworkAttrTypes()}
	switch mode {
	case "null":
		return types.ListNull(objType)
	case "empty":
		return types.ListValueMust(objType, []attr.Value{})
	case "populated":
		return types.ListValueMust(objType, []attr.Value{
			types.ObjectValueMust(NetworkAttrTypes(), map[string]attr.Value{
				"prefix": types.StringValue("192.168.0.0"),
				"mask":   types.StringValue("255.255.0.0"),
			}),
		})
	}
	t.Fatalf("unknown mode: %s", mode)
	return types.List{}
}

func TestFromClient_Neighbors_NullPreservation(t *testing.T) {
	cases := []struct {
		name      string
		prior     string
		neighbors []client.BGPNeighbor
		wantNull  bool
		wantSize  int
	}{
		{"empty + prior null stays null", "null", nil, true, 0},
		{"empty + prior empty stays empty", "empty", nil, false, 0},
		{"empty + prior populated overwrites to empty", "populated", nil, false, 0},
		{"populated over prior null", "null", []client.BGPNeighbor{{ID: 1, IP: "10.0.0.1", RemoteAS: "65001"}}, false, 1},
		{"populated over prior empty", "empty", []client.BGPNeighbor{{ID: 2, IP: "10.0.0.2", RemoteAS: "65002"}}, false, 1},
		{"populated over prior populated", "populated", []client.BGPNeighbor{{ID: 3, IP: "10.0.0.3", RemoteAS: "65003"}}, false, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &BGPModel{
				Neighbors: makePriorNeighbors(t, tc.prior),
				Networks:  types.ListNull(types.ObjectType{AttrTypes: NetworkAttrTypes()}),
			}
			m.FromClient(&client.BGPConfig{ASN: "65000", Neighbors: tc.neighbors})
			if got := m.Neighbors.IsNull(); got != tc.wantNull {
				t.Errorf("Neighbors.IsNull() = %v, want %v", got, tc.wantNull)
			}
			if !tc.wantNull && len(m.Neighbors.Elements()) != tc.wantSize {
				t.Errorf("len(Neighbors.Elements()) = %d, want %d", len(m.Neighbors.Elements()), tc.wantSize)
			}
		})
	}
}

func TestFromClient_Networks_NullPreservation(t *testing.T) {
	cases := []struct {
		name     string
		prior    string
		networks []client.BGPNetwork
		wantNull bool
		wantSize int
	}{
		{"empty + prior null stays null", "null", nil, true, 0},
		{"empty + prior empty stays empty", "empty", nil, false, 0},
		{"empty + prior populated overwrites to empty", "populated", nil, false, 0},
		{"populated over prior null", "null", []client.BGPNetwork{{Prefix: "192.168.0.0", Mask: "255.255.0.0"}}, false, 1},
		{"populated over prior empty", "empty", []client.BGPNetwork{{Prefix: "172.16.0.0", Mask: "255.240.0.0"}}, false, 1},
		{"populated over prior populated", "populated", []client.BGPNetwork{{Prefix: "10.0.0.0", Mask: "255.0.0.0"}}, false, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &BGPModel{
				Neighbors: types.ListNull(types.ObjectType{AttrTypes: NeighborAttrTypes()}),
				Networks:  makePriorNetworks(t, tc.prior),
			}
			m.FromClient(&client.BGPConfig{ASN: "65000", Networks: tc.networks})
			if got := m.Networks.IsNull(); got != tc.wantNull {
				t.Errorf("Networks.IsNull() = %v, want %v", got, tc.wantNull)
			}
			if !tc.wantNull && len(m.Networks.Elements()) != tc.wantSize {
				t.Errorf("len(Networks.Elements()) = %d, want %d", len(m.Networks.Elements()), tc.wantSize)
			}
		})
	}
}
