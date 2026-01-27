package system

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// SystemModel describes the resource data model.
type SystemModel struct {
	ID           types.String `tfsdk:"id"`
	Timezone     types.String `tfsdk:"timezone"`
	Console      types.List   `tfsdk:"console"`
	PacketBuffer types.List   `tfsdk:"packet_buffer"`
	Statistics   types.List   `tfsdk:"statistics"`
}

// ConsoleModel describes the console nested block.
type ConsoleModel struct {
	Character types.String `tfsdk:"character"`
	Lines     types.String `tfsdk:"lines"`
	Prompt    types.String `tfsdk:"prompt"`
}

// PacketBufferModel describes the packet_buffer nested block.
type PacketBufferModel struct {
	Size      types.String `tfsdk:"size"`
	MaxBuffer types.Int64  `tfsdk:"max_buffer"`
	MaxFree   types.Int64  `tfsdk:"max_free"`
}

// StatisticsModel describes the statistics nested block.
type StatisticsModel struct {
	Traffic types.Bool `tfsdk:"traffic"`
	NAT     types.Bool `tfsdk:"nat"`
}

// ConsoleModelType returns the attribute types for ConsoleModel.
func ConsoleModelType() map[string]attr.Type {
	return map[string]attr.Type{
		"character": types.StringType,
		"lines":     types.StringType,
		"prompt":    types.StringType,
	}
}

// PacketBufferModelType returns the attribute types for PacketBufferModel.
func PacketBufferModelType() map[string]attr.Type {
	return map[string]attr.Type{
		"size":       types.StringType,
		"max_buffer": types.Int64Type,
		"max_free":   types.Int64Type,
	}
}

// StatisticsModelType returns the attribute types for StatisticsModel.
func StatisticsModelType() map[string]attr.Type {
	return map[string]attr.Type{
		"traffic": types.BoolType,
		"nat":     types.BoolType,
	}
}

// ToClient converts the Terraform model to a client.SystemConfig.
func (m *SystemModel) ToClient() client.SystemConfig {
	config := client.SystemConfig{
		Timezone:      fwhelpers.GetStringValue(m.Timezone),
		PacketBuffers: []client.PacketBufferConfig{},
	}

	// Handle console block
	if !m.Console.IsNull() && !m.Console.IsUnknown() && len(m.Console.Elements()) > 0 {
		var consoleModels []ConsoleModel
		m.Console.ElementsAs(context.Background(), &consoleModels, false)
		if len(consoleModels) > 0 {
			cm := consoleModels[0]
			config.Console = &client.ConsoleConfig{
				Character: fwhelpers.GetStringValue(cm.Character),
				Lines:     fwhelpers.GetStringValue(cm.Lines),
				Prompt:    fwhelpers.GetStringValue(cm.Prompt),
			}
		}
	}

	// Handle packet_buffer list
	if !m.PacketBuffer.IsNull() && !m.PacketBuffer.IsUnknown() {
		var pbModels []PacketBufferModel
		m.PacketBuffer.ElementsAs(context.Background(), &pbModels, false)
		for _, pb := range pbModels {
			config.PacketBuffers = append(config.PacketBuffers, client.PacketBufferConfig{
				Size:      fwhelpers.GetStringValue(pb.Size),
				MaxBuffer: fwhelpers.GetInt64Value(pb.MaxBuffer),
				MaxFree:   fwhelpers.GetInt64Value(pb.MaxFree),
			})
		}
	}

	// Handle statistics block
	if !m.Statistics.IsNull() && !m.Statistics.IsUnknown() && len(m.Statistics.Elements()) > 0 {
		var statsModels []StatisticsModel
		m.Statistics.ElementsAs(context.TODO(), &statsModels, false)
		if len(statsModels) > 0 {
			sm := statsModels[0]
			config.Statistics = &client.StatisticsConfig{
				Traffic: fwhelpers.GetBoolValue(sm.Traffic),
				NAT:     fwhelpers.GetBoolValue(sm.NAT),
			}
		}
	}

	return config
}

// FromClient updates the Terraform model from a client.SystemConfig.
func (m *SystemModel) FromClient(config *client.SystemConfig) {
	m.ID = types.StringValue("system")
	m.Timezone = fwhelpers.StringValueOrNull(config.Timezone)

	// Handle console
	if config.Console != nil && (config.Console.Character != "" || config.Console.Lines != "" || config.Console.Prompt != "") {
		consoleObj, _ := types.ObjectValue(ConsoleModelType(), map[string]attr.Value{
			"character": fwhelpers.StringValueOrNull(config.Console.Character),
			"lines":     fwhelpers.StringValueOrNull(config.Console.Lines),
			"prompt":    fwhelpers.StringValueOrNull(config.Console.Prompt),
		})
		m.Console, _ = types.ListValue(types.ObjectType{AttrTypes: ConsoleModelType()}, []attr.Value{consoleObj})
	} else {
		m.Console = types.ListValueMust(types.ObjectType{AttrTypes: ConsoleModelType()}, []attr.Value{})
	}

	// Handle packet_buffer
	if len(config.PacketBuffers) > 0 {
		pbValues := make([]attr.Value, len(config.PacketBuffers))
		for i, pb := range config.PacketBuffers {
			pbObj, _ := types.ObjectValue(PacketBufferModelType(), map[string]attr.Value{
				"size":       types.StringValue(pb.Size),
				"max_buffer": types.Int64Value(int64(pb.MaxBuffer)),
				"max_free":   types.Int64Value(int64(pb.MaxFree)),
			})
			pbValues[i] = pbObj
		}
		m.PacketBuffer, _ = types.ListValue(types.ObjectType{AttrTypes: PacketBufferModelType()}, pbValues)
	} else {
		m.PacketBuffer = types.ListValueMust(types.ObjectType{AttrTypes: PacketBufferModelType()}, []attr.Value{})
	}

	// Handle statistics
	if config.Statistics != nil {
		statsObj, _ := types.ObjectValue(StatisticsModelType(), map[string]attr.Value{
			"traffic": types.BoolValue(config.Statistics.Traffic),
			"nat":     types.BoolValue(config.Statistics.NAT),
		})
		m.Statistics, _ = types.ListValue(types.ObjectType{AttrTypes: StatisticsModelType()}, []attr.Value{statsObj})
	} else {
		m.Statistics = types.ListValueMust(types.ObjectType{AttrTypes: StatisticsModelType()}, []attr.Value{})
	}
}
