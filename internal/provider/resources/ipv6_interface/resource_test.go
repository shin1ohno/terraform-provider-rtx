package ipv6_interface

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// upgradeStateV0 runs the v0 -> v1 state upgrader against a prior state whose
// rtadv block is built from rtadvAttrs (nil means the whole block is null).
func upgradeStateV0(ctx context.Context, t *testing.T, rtadvAttrs map[string]tftypes.Value) IPv6InterfaceModel {
	t.Helper()

	r := &IPv6InterfaceResource{}

	upgrader, ok := r.UpgradeState(ctx)[0]
	if !ok {
		t.Fatalf("no state upgrader registered for schema version 0")
	}
	if upgrader.PriorSchema == nil {
		t.Fatalf("state upgrader for version 0 has no PriorSchema")
	}

	priorSchema := *upgrader.PriorSchema
	priorType := priorSchema.Type().TerraformType(ctx)

	rtadvType := priorType.(tftypes.Object).AttributeTypes["rtadv"]
	rtadv := tftypes.NewValue(rtadvType, nil)
	if rtadvAttrs != nil {
		rtadv = tftypes.NewValue(rtadvType, rtadvAttrs)
	}

	priorRaw := tftypes.NewValue(priorType, map[string]tftypes.Value{
		"interface":      tftypes.NewValue(tftypes.String, "lan1"),
		"dhcpv6_service": tftypes.NewValue(tftypes.String, "server"),
		"mtu":            tftypes.NewValue(tftypes.Number, 1500),
		"address":        tftypes.NewValue(priorType.(tftypes.Object).AttributeTypes["address"], nil),
		"rtadv":          rtadv,
	})

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	req := resource.UpgradeStateRequest{
		State: &tfsdk.State{Raw: priorRaw, Schema: priorSchema},
	}
	resp := &resource.UpgradeStateResponse{
		State: tfsdk.State{
			Raw:    tftypes.NewValue(schemaResp.Schema.Type().TerraformType(ctx), nil),
			Schema: schemaResp.Schema,
		},
	}

	upgrader.StateUpgrader(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("state upgrade produced errors: %v", resp.Diagnostics.Errors())
	}

	var upgraded IPv6InterfaceModel
	if diags := resp.State.Get(ctx, &upgraded); diags.HasError() {
		t.Fatalf("reading upgraded state: %v", diags.Errors())
	}

	return upgraded
}

func TestUpgradeStateV0ToV1_PrefixIDBecomesSingleElementList(t *testing.T) {
	ctx := context.Background()

	upgraded := upgradeStateV0(ctx, t, map[string]tftypes.Value{
		"enabled":   tftypes.NewValue(tftypes.Bool, true),
		"prefix_id": tftypes.NewValue(tftypes.Number, 2),
		"o_flag":    tftypes.NewValue(tftypes.Bool, true),
		"m_flag":    tftypes.NewValue(tftypes.Bool, false),
		"lifetime":  tftypes.NewValue(tftypes.Number, 1800),
	})

	if upgraded.Interface.ValueString() != "lan1" {
		t.Errorf("Interface = %q, want lan1", upgraded.Interface.ValueString())
	}
	if upgraded.DHCPv6Service.ValueString() != "server" {
		t.Errorf("DHCPv6Service = %q, want server", upgraded.DHCPv6Service.ValueString())
	}
	if upgraded.MTU.ValueInt64() != 1500 {
		t.Errorf("MTU = %d, want 1500", upgraded.MTU.ValueInt64())
	}

	if upgraded.RTADV == nil {
		t.Fatalf("RTADV = nil, want the block to survive the upgrade")
	}

	want := fwhelpers.IntSliceToList([]int{2})
	if !upgraded.RTADV.PrefixIDs.Equal(want) {
		t.Errorf("PrefixIDs = %v, want %v", upgraded.RTADV.PrefixIDs, want)
	}

	if !upgraded.RTADV.Enabled.ValueBool() {
		t.Errorf("Enabled = false, want true")
	}
	if !upgraded.RTADV.OFlag.ValueBool() {
		t.Errorf("OFlag = false, want true")
	}
	if upgraded.RTADV.MFlag.ValueBool() {
		t.Errorf("MFlag = true, want false")
	}
	if upgraded.RTADV.Lifetime.ValueInt64() != 1800 {
		t.Errorf("Lifetime = %d, want 1800", upgraded.RTADV.Lifetime.ValueInt64())
	}
}

func TestUpgradeStateV0ToV1_NullPrefixIDStaysNull(t *testing.T) {
	ctx := context.Background()

	upgraded := upgradeStateV0(ctx, t, map[string]tftypes.Value{
		"enabled":   tftypes.NewValue(tftypes.Bool, false),
		"prefix_id": tftypes.NewValue(tftypes.Number, nil),
		"o_flag":    tftypes.NewValue(tftypes.Bool, false),
		"m_flag":    tftypes.NewValue(tftypes.Bool, false),
		"lifetime":  tftypes.NewValue(tftypes.Number, nil),
	})

	if upgraded.RTADV == nil {
		t.Fatalf("RTADV = nil, want the block to survive the upgrade")
	}
	if !upgraded.RTADV.PrefixIDs.IsNull() {
		t.Errorf("PrefixIDs = %v, want null", upgraded.RTADV.PrefixIDs)
	}
	if upgraded.RTADV.PrefixIDs.ElementType(ctx) != types.Int64Type {
		t.Errorf("PrefixIDs element type = %v, want Int64", upgraded.RTADV.PrefixIDs.ElementType(ctx))
	}
}

func TestUpgradeStateV0ToV1_NullRTADVBlockStaysNull(t *testing.T) {
	ctx := context.Background()

	upgraded := upgradeStateV0(ctx, t, nil)

	if upgraded.RTADV != nil {
		t.Errorf("RTADV = %+v, want nil", upgraded.RTADV)
	}
}
