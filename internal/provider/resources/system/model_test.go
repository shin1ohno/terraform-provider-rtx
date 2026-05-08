package system

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func makePriorList(t *testing.T, mode string, attrTypes map[string]attr.Type, populated attr.Value) types.List {
	t.Helper()
	objType := types.ObjectType{AttrTypes: attrTypes}
	switch mode {
	case "null":
		return types.ListNull(objType)
	case "empty":
		return types.ListValueMust(objType, []attr.Value{})
	case "populated":
		return types.ListValueMust(objType, []attr.Value{populated})
	}
	t.Fatalf("unknown mode: %s", mode)
	return types.List{}
}

func priorConsoleObj() attr.Value {
	return types.ObjectValueMust(ConsoleModelType(), map[string]attr.Value{
		"character": types.StringValue("ascii"),
		"lines":     types.StringNull(),
		"prompt":    types.StringNull(),
	})
}

func priorPacketBufferObj() attr.Value {
	return types.ObjectValueMust(PacketBufferModelType(), map[string]attr.Value{
		"size":       types.StringValue("small"),
		"max_buffer": types.Int64Value(100),
		"max_free":   types.Int64Value(10),
	})
}

func priorStatisticsObj() attr.Value {
	return types.ObjectValueMust(StatisticsModelType(), map[string]attr.Value{
		"traffic": types.BoolValue(true),
		"nat":     types.BoolValue(false),
	})
}

// emptyConfig returns a SystemConfig that triggers the empty/nil branch for all
// three list fields.
func emptyConfig() *client.SystemConfig {
	return &client.SystemConfig{}
}

func nullPriors(t *testing.T) SystemModel {
	t.Helper()
	return SystemModel{
		Console:      types.ListNull(types.ObjectType{AttrTypes: ConsoleModelType()}),
		PacketBuffer: types.ListNull(types.ObjectType{AttrTypes: PacketBufferModelType()}),
		Statistics:   types.ListNull(types.ObjectType{AttrTypes: StatisticsModelType()}),
	}
}

func TestFromClient_Console_NullPreservation(t *testing.T) {
	cases := []struct {
		name     string
		prior    string
		config   *client.SystemConfig
		wantNull bool
		wantSize int
	}{
		{"empty + prior null stays null", "null", emptyConfig(), true, 0},
		{"empty + prior empty stays empty", "empty", emptyConfig(), false, 0},
		{"empty + prior populated overwrites to empty", "populated", emptyConfig(), false, 0},
		{"populated config wins", "null", &client.SystemConfig{Console: &client.ConsoleConfig{Character: "ascii"}}, false, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := nullPriors(t)
			m.Console = makePriorList(t, tc.prior, ConsoleModelType(), priorConsoleObj())
			m.FromClient(tc.config)
			if got := m.Console.IsNull(); got != tc.wantNull {
				t.Errorf("Console.IsNull() = %v, want %v", got, tc.wantNull)
			}
			if !tc.wantNull && len(m.Console.Elements()) != tc.wantSize {
				t.Errorf("len(Console.Elements()) = %d, want %d", len(m.Console.Elements()), tc.wantSize)
			}
		})
	}
}

func TestFromClient_PacketBuffer_NullPreservation(t *testing.T) {
	cases := []struct {
		name     string
		prior    string
		config   *client.SystemConfig
		wantNull bool
		wantSize int
	}{
		{"empty + prior null stays null", "null", emptyConfig(), true, 0},
		{"empty + prior empty stays empty", "empty", emptyConfig(), false, 0},
		{"empty + prior populated overwrites to empty", "populated", emptyConfig(), false, 0},
		{"populated config wins", "null", &client.SystemConfig{PacketBuffers: []client.PacketBufferConfig{{Size: "small", MaxBuffer: 100, MaxFree: 10}}}, false, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := nullPriors(t)
			m.PacketBuffer = makePriorList(t, tc.prior, PacketBufferModelType(), priorPacketBufferObj())
			m.FromClient(tc.config)
			if got := m.PacketBuffer.IsNull(); got != tc.wantNull {
				t.Errorf("PacketBuffer.IsNull() = %v, want %v", got, tc.wantNull)
			}
			if !tc.wantNull && len(m.PacketBuffer.Elements()) != tc.wantSize {
				t.Errorf("len(PacketBuffer.Elements()) = %d, want %d", len(m.PacketBuffer.Elements()), tc.wantSize)
			}
		})
	}
}

func TestFromClient_Statistics_NullPreservation(t *testing.T) {
	cases := []struct {
		name     string
		prior    string
		config   *client.SystemConfig
		wantNull bool
		wantSize int
	}{
		{"empty + prior null stays null", "null", emptyConfig(), true, 0},
		{"empty + prior empty stays empty", "empty", emptyConfig(), false, 0},
		{"empty + prior populated overwrites to empty", "populated", emptyConfig(), false, 0},
		{"populated config wins", "null", &client.SystemConfig{Statistics: &client.StatisticsConfig{Traffic: true, NAT: false}}, false, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := nullPriors(t)
			m.Statistics = makePriorList(t, tc.prior, StatisticsModelType(), priorStatisticsObj())
			m.FromClient(tc.config)
			if got := m.Statistics.IsNull(); got != tc.wantNull {
				t.Errorf("Statistics.IsNull() = %v, want %v", got, tc.wantNull)
			}
			if !tc.wantNull && len(m.Statistics.Elements()) != tc.wantSize {
				t.Errorf("len(Statistics.Elements()) = %d, want %d", len(m.Statistics.Elements()), tc.wantSize)
			}
		})
	}
}
