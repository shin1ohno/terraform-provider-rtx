package kron_schedule

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
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

// reconcileWithDesired restores the attributes Terraform requires to come back
// exactly as the practitioner wrote them.
//
// name, policy_list, day_of_week, at_time and date are Optional and NOT
// Computed, so terraform core's post-apply consistency check demands the
// applied value equal the planned one verbatim — null included. Three of them
// the RTX cannot store at all: name and policy_list are Terraform-side labels
// with no place in a `schedule at` line, and that line encodes weekdays inside
// its date token rather than in a field of its own, so read-back always returns
// them empty. at_time does come back, but in the router's own rendering: `0:00`
// is written and `show config` prints `00:00:00`.
//
// Only values equal modulo that formatting are replaced. A genuinely different
// time or date on the device is left in state, so the next plan shows the
// drift instead of hiding it. Same contract as fwhelpers/reorder.go: desired is
// the plan on Create/Update and the prior state on Read.
func (m *KronScheduleModel) reconcileWithDesired(desired *KronScheduleModel) {
	if desired == nil {
		return
	}

	m.Name = desired.Name
	m.PolicyList = desired.PolicyList
	m.DayOfWeek = desired.DayOfWeek

	if parsers.NormalizeScheduleTime(fwhelpers.GetStringValue(m.AtTime)) ==
		parsers.NormalizeScheduleTime(fwhelpers.GetStringValue(desired.AtTime)) {
		m.AtTime = desired.AtTime
	}

	if parsers.NormalizeScheduleDate(fwhelpers.GetStringValue(m.Date)) ==
		parsers.NormalizeScheduleDate(fwhelpers.GetStringValue(desired.Date)) {
		m.Date = desired.Date
	}
}
