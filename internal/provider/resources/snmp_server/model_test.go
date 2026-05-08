package snmp_server

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

// priorState captures the three list-typed fields used by the null-preservation
// branch in FromClient.
type priorState int

const (
	priorNull priorState = iota
	priorEmpty
	priorPopulated
)

func makePriorCommunities(t *testing.T, s priorState) types.List {
	t.Helper()
	objType := types.ObjectType{AttrTypes: CommunityAttrTypes()}
	switch s {
	case priorNull:
		return types.ListNull(objType)
	case priorEmpty:
		return types.ListValueMust(objType, []attr.Value{})
	case priorPopulated:
		return types.ListValueMust(objType, []attr.Value{
			types.ObjectValueMust(CommunityAttrTypes(), map[string]attr.Value{
				"name":       types.StringValue("public"),
				"permission": types.StringValue("ro"),
				"acl":        types.StringNull(),
			}),
		})
	}
	t.Fatalf("unhandled priorState: %d", s)
	return types.List{}
}

func makePriorHosts(t *testing.T, s priorState) types.List {
	t.Helper()
	objType := types.ObjectType{AttrTypes: HostAttrTypes()}
	switch s {
	case priorNull:
		return types.ListNull(objType)
	case priorEmpty:
		return types.ListValueMust(objType, []attr.Value{})
	case priorPopulated:
		return types.ListValueMust(objType, []attr.Value{
			types.ObjectValueMust(HostAttrTypes(), map[string]attr.Value{
				"ip_address": types.StringValue("192.0.2.1"),
				"community":  types.StringNull(),
				"version":    types.StringNull(),
			}),
		})
	}
	t.Fatalf("unhandled priorState: %d", s)
	return types.List{}
}

func makePriorEnableTraps(t *testing.T, s priorState) types.List {
	t.Helper()
	switch s {
	case priorNull:
		return types.ListNull(types.StringType)
	case priorEmpty:
		return types.ListValueMust(types.StringType, []attr.Value{})
	case priorPopulated:
		return types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("coldstart"),
		})
	}
	t.Fatalf("unhandled priorState: %d", s)
	return types.List{}
}

// expectedShape describes what FromClient must produce for a list-typed field.
type expectedShape struct {
	null bool
	size int
}

func assertListShape(t *testing.T, fieldName string, got types.List, want expectedShape) {
	t.Helper()
	if got.IsNull() != want.null {
		t.Errorf("%s: IsNull = %v, want %v", fieldName, got.IsNull(), want.null)
	}
	if !want.null {
		if size := len(got.Elements()); size != want.size {
			t.Errorf("%s: len(Elements) = %d, want %d", fieldName, size, want.size)
		}
	}
}

func TestFromClient_Communities_NullPreservation(t *testing.T) {
	cases := []struct {
		name       string
		prior      priorState
		clientData []client.SNMPCommunity
		want       expectedShape
	}{
		{"empty config + prior null stays null", priorNull, nil, expectedShape{null: true}},
		{"empty config + prior empty stays empty", priorEmpty, nil, expectedShape{null: false, size: 0}},
		{"empty config + prior populated overwrites to empty (drift)", priorPopulated, nil, expectedShape{null: false, size: 0}},
		{"populated config wins over prior null", priorNull, []client.SNMPCommunity{{Name: "public", Permission: "ro"}}, expectedShape{null: false, size: 1}},
		{"populated config wins over prior empty", priorEmpty, []client.SNMPCommunity{{Name: "public", Permission: "ro"}}, expectedShape{null: false, size: 1}},
		{"populated config wins over prior populated", priorPopulated, []client.SNMPCommunity{{Name: "private", Permission: "rw"}}, expectedShape{null: false, size: 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &SNMPServerModel{Communities: makePriorCommunities(t, tc.prior)}
			cfg := &client.SNMPConfig{Communities: tc.clientData}
			m.FromClient(cfg)
			assertListShape(t, "Communities", m.Communities, tc.want)
		})
	}
}

func TestFromClient_Hosts_NullPreservation(t *testing.T) {
	cases := []struct {
		name       string
		prior      priorState
		clientData []client.SNMPHost
		want       expectedShape
	}{
		{"empty config + prior null stays null", priorNull, nil, expectedShape{null: true}},
		{"empty config + prior empty stays empty", priorEmpty, nil, expectedShape{null: false, size: 0}},
		{"empty config + prior populated overwrites to empty (drift)", priorPopulated, nil, expectedShape{null: false, size: 0}},
		{"populated config wins over prior null", priorNull, []client.SNMPHost{{Address: "192.0.2.1"}}, expectedShape{null: false, size: 1}},
		{"populated config wins over prior empty", priorEmpty, []client.SNMPHost{{Address: "192.0.2.1"}}, expectedShape{null: false, size: 1}},
		{"populated config wins over prior populated", priorPopulated, []client.SNMPHost{{Address: "192.0.2.2"}}, expectedShape{null: false, size: 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &SNMPServerModel{Hosts: makePriorHosts(t, tc.prior)}
			cfg := &client.SNMPConfig{Hosts: tc.clientData}
			m.FromClient(cfg)
			assertListShape(t, "Hosts", m.Hosts, tc.want)
		})
	}
}

func TestFromClient_EnableTraps_NullPreservation(t *testing.T) {
	cases := []struct {
		name       string
		prior      priorState
		clientData []string
		want       expectedShape
	}{
		// This case is the precise reproduction of the 2026-05-08 bug report:
		// home-monitor's rtx_snmp_server.hnd has no `enable_traps` block, so the
		// plan resolves to null; the router has no traps enabled, so config
		// arrives empty. Pre-fix this triggered "was null, but now
		// cty.ListValEmpty(cty.String)" on apply.
		{"empty config + prior null stays null", priorNull, nil, expectedShape{null: true}},
		{"empty config + prior empty stays empty", priorEmpty, nil, expectedShape{null: false, size: 0}},
		{"empty config + prior populated overwrites to empty (drift)", priorPopulated, nil, expectedShape{null: false, size: 0}},
		{"populated config wins over prior null", priorNull, []string{"coldstart"}, expectedShape{null: false, size: 1}},
		{"populated config wins over prior empty", priorEmpty, []string{"coldstart"}, expectedShape{null: false, size: 1}},
		{"populated config wins over prior populated", priorPopulated, []string{"warmstart"}, expectedShape{null: false, size: 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &SNMPServerModel{EnableTraps: makePriorEnableTraps(t, tc.prior)}
			cfg := &client.SNMPConfig{TrapEnable: tc.clientData}
			m.FromClient(cfg)
			assertListShape(t, "EnableTraps", m.EnableTraps, tc.want)
		})
	}
}
