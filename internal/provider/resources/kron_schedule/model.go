package kron_schedule

import (
	"strings"

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
// Computed, so terraform core's post-apply consistency check demands the applied
// value equal the planned one verbatim — null included. name and policy_list the
// RTX cannot store at all: they are Terraform-side labels with no place in a
// `schedule at` line, so config is their only source of truth and they are
// echoed unconditionally. at_time, date and day_of_week DO come back, but in the
// router's own spelling — `0:00` is written and `show config` prints
// `00:00:00`, and a weekday is carried as the date token `*/mon-fri`.
//
// Those three are only replaced when they are equal modulo that formatting. A
// genuinely different time, date or weekday on the device stays in state so the
// next plan shows the drift instead of hiding it — which matters most for
// day_of_week, where an unconditional echo would report success for a weekday
// schedule the router is actually running daily. Same contract as
// fwhelpers/reorder.go: desired is the plan on Create/Update and the prior
// state on Read.
func (m *KronScheduleModel) reconcileWithDesired(desired *KronScheduleModel) {
	if desired == nil {
		return
	}

	m.Name = preferDesired(m.Name, desired.Name)
	m.PolicyList = preferDesired(m.PolicyList, desired.PolicyList)

	if parsers.NormalizeScheduleTime(fwhelpers.GetStringValue(m.AtTime)) ==
		parsers.NormalizeScheduleTime(fwhelpers.GetStringValue(desired.AtTime)) {
		m.AtTime = preferDesired(m.AtTime, desired.AtTime)
	}

	if parsers.NormalizeScheduleDate(fwhelpers.GetStringValue(m.Date)) ==
		parsers.NormalizeScheduleDate(fwhelpers.GetStringValue(desired.Date)) {
		m.Date = preferDesired(m.Date, desired.Date)
	}

	if parsers.ScheduleDayOfWeekKey(fwhelpers.GetStringValue(m.DayOfWeek)) ==
		parsers.ScheduleDayOfWeekKey(fwhelpers.GetStringValue(desired.DayOfWeek)) {
		m.DayOfWeek = preferDesired(m.DayOfWeek, desired.DayOfWeek)
	}

	// The router's config rendering is trimmed, and so is every line this
	// package parses, so a command written with surrounding whitespace can
	// never be read back verbatim. Echo the configured element when it differs
	// only by that, rather than failing an apply over a space.
	if !desired.CommandLines.IsUnknown() &&
		sameCommandsIgnoringSurroundingSpace(
			fwhelpers.ListToStringSlice(m.CommandLines),
			fwhelpers.ListToStringSlice(desired.CommandLines)) {
		m.CommandLines = desired.CommandLines
	}
}

func sameCommandsIgnoringSurroundingSpace(actual, desired []string) bool {
	if len(actual) != len(desired) {
		return false
	}
	for i := range actual {
		if strings.TrimSpace(actual[i]) != strings.TrimSpace(desired[i]) {
			return false
		}
	}
	return true
}

// preferDesired returns the practitioner's value unless it is unknown, which
// can never be written into state.
func preferDesired(actual, desired types.String) types.String {
	if desired.IsUnknown() {
		return actual
	}
	return desired
}
