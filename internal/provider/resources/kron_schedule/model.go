package kron_schedule

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// KronScheduleModel describes the resource data model.
type KronScheduleModel struct {
	ScheduleID   types.Int64  `tfsdk:"schedule_id"`
	Name         types.String `tfsdk:"name"`
	AtTime       types.String `tfsdk:"at_time"`
	DayOfWeek    types.String `tfsdk:"day_of_week"`
	Date         types.String `tfsdk:"date"`
	Recurring    types.Bool   `tfsdk:"recurring"`
	OnStartup    types.Bool   `tfsdk:"on_startup"`
	PolicyList   types.String `tfsdk:"policy_list"`
	CommandLines types.List   `tfsdk:"command_lines"`
}

// ToClient converts the Terraform model to a client.Schedule.
func (m *KronScheduleModel) ToClient() client.Schedule {
	schedule := client.Schedule{
		ID:        fwhelpers.GetInt64Value(m.ScheduleID),
		Name:      fwhelpers.GetStringValue(m.Name),
		AtTime:    fwhelpers.GetStringValue(m.AtTime),
		DayOfWeek: fwhelpers.GetStringValue(m.DayOfWeek),
		Date:      fwhelpers.GetStringValue(m.Date),
		Recurring: fwhelpers.GetBoolValue(m.Recurring),
		OnStartup: fwhelpers.GetBoolValue(m.OnStartup),
		Enabled:   true, // Always enabled when managed by Terraform
	}

	if !m.PolicyList.IsNull() && !m.PolicyList.IsUnknown() {
		schedule.PolicyList = m.PolicyList.ValueString()
	}

	// Convert command lines list
	schedule.Commands = fwhelpers.ListToStringSlice(m.CommandLines)

	// Date-specific schedules are one-time
	if schedule.Date != "" {
		schedule.Recurring = false
	}

	return schedule
}

// FromClient updates the Terraform model from a client.Schedule.
func (m *KronScheduleModel) FromClient(schedule *client.Schedule) {
	m.ScheduleID = types.Int64Value(int64(schedule.ID))
	m.Name = fwhelpers.StringValueOrNull(schedule.Name)
	m.AtTime = fwhelpers.StringValueOrNull(schedule.AtTime)
	m.DayOfWeek = fwhelpers.StringValueOrNull(schedule.DayOfWeek)
	m.Date = fwhelpers.StringValueOrNull(schedule.Date)
	m.Recurring = types.BoolValue(schedule.Recurring)
	m.OnStartup = types.BoolValue(schedule.OnStartup)
	m.PolicyList = fwhelpers.StringValueOrNull(schedule.PolicyList)

	// Preserve empty list vs null: no commands on RTX is equivalent to empty list
	if schedule.Commands == nil && !m.CommandLines.IsNull() {
		m.CommandLines = fwhelpers.StringSliceToList([]string{})
	} else {
		m.CommandLines = fwhelpers.StringSliceToList(schedule.Commands)
	}
}
