package syslog

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func makePriorHosts(t *testing.T, mode string) types.Set {
	t.Helper()
	objType := types.ObjectType{AttrTypes: HostAttrTypes()}
	switch mode {
	case "null":
		return types.SetNull(objType)
	case "empty":
		return types.SetValueMust(objType, []attr.Value{})
	case "populated":
		return types.SetValueMust(objType, []attr.Value{
			types.ObjectValueMust(HostAttrTypes(), map[string]attr.Value{
				"address": types.StringValue("192.0.2.1"),
				"port":    types.Int64Value(514),
			}),
		})
	}
	t.Fatalf("unknown mode: %s", mode)
	return types.Set{}
}

func TestFromClient_Hosts_NullPreservation(t *testing.T) {
	cases := []struct {
		name     string
		prior    string
		hosts    []client.SyslogHost
		wantNull bool
		wantSize int
	}{
		{"empty + prior null stays null", "null", nil, true, 0},
		{"empty + prior empty stays empty", "empty", nil, false, 0},
		{"empty + prior populated overwrites to empty", "populated", nil, false, 0},
		{"populated over prior null", "null", []client.SyslogHost{{Address: "10.0.0.1", Port: 514}}, false, 1},
		{"populated over prior empty", "empty", []client.SyslogHost{{Address: "10.0.0.2"}}, false, 1},
		{"populated over prior populated", "populated", []client.SyslogHost{{Address: "10.0.0.3", Port: 1514}}, false, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &SyslogModel{Hosts: makePriorHosts(t, tc.prior)}
			diags := m.FromClient(context.Background(), &client.SyslogConfig{Hosts: tc.hosts})
			if diags.HasError() {
				t.Fatalf("FromClient returned errors: %v", diags.Errors())
			}
			if got := m.Hosts.IsNull(); got != tc.wantNull {
				t.Errorf("Hosts.IsNull() = %v, want %v", got, tc.wantNull)
			}
			if !tc.wantNull && len(m.Hosts.Elements()) != tc.wantSize {
				t.Errorf("len(Hosts.Elements()) = %d, want %d", len(m.Hosts.Elements()), tc.wantSize)
			}
		})
	}
}
