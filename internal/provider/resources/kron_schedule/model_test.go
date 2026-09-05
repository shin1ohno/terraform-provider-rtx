package kron_schedule

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/provider/fwhelpers"
)

// Terraform core compares the post-apply state against the plan attribute by
// attribute, so a value the router merely re-formatted has to be handed back in
// the practitioner's own spelling. Anything genuinely different has to survive
// into state instead, or real drift becomes invisible.
func TestReconcileWithDesired(t *testing.T) {
	tests := []struct {
		name          string
		fromRouter    KronScheduleModel
		desired       *KronScheduleModel
		wantAtTime    types.String
		wantDate      types.String
		wantName      types.String
		wantDayOfWeek types.String
	}{
		{
			name: "router rendering of the same instant keeps the configured spelling",
			fromRouter: KronScheduleModel{
				AtTime: types.StringValue("00:00:00"),
				Date:   types.StringNull(),
				Name:   types.StringNull(),
			},
			desired: &KronScheduleModel{
				AtTime: types.StringValue("0:00"),
				Date:   types.StringNull(),
				Name:   types.StringNull(),
			},
			wantAtTime:    types.StringValue("0:00"),
			wantDate:      types.StringNull(),
			wantName:      types.StringNull(),
			wantDayOfWeek: types.StringNull(),
		},
		{
			name: "a genuinely different time stays in state so the drift shows",
			fromRouter: KronScheduleModel{
				AtTime: types.StringValue("01:00:00"),
			},
			desired: &KronScheduleModel{
				AtTime: types.StringValue("0:00"),
			},
			wantAtTime:    types.StringValue("01:00:00"),
			wantDate:      types.StringNull(),
			wantName:      types.StringNull(),
			wantDayOfWeek: types.StringNull(),
		},
		{
			name: "zero-padding differences in a one-time date are not drift",
			fromRouter: KronScheduleModel{
				AtTime: types.StringValue("09:00"),
				Date:   types.StringValue("2025/1/5"),
			},
			desired: &KronScheduleModel{
				AtTime: types.StringValue("09:00"),
				Date:   types.StringValue("2025/01/05"),
			},
			wantAtTime:    types.StringValue("09:00"),
			wantDate:      types.StringValue("2025/01/05"),
			wantName:      types.StringNull(),
			wantDayOfWeek: types.StringNull(),
		},
		{
			// name has no field in a `schedule at` line, so config is its only
			// source of truth and it comes back unconditionally.
			name: "a label the router cannot store comes back from the config",
			fromRouter: KronScheduleModel{
				AtTime: types.StringValue("0:00"),
				Name:   types.StringNull(),
			},
			desired: &KronScheduleModel{
				AtTime: types.StringValue("0:00"),
				Name:   types.StringValue("nightly ntpdate"),
			},
			wantAtTime:    types.StringValue("0:00"),
			wantDate:      types.StringNull(),
			wantName:      types.StringValue("nightly ntpdate"),
			wantDayOfWeek: types.StringNull(),
		},
		{
			name: "weekday spelling differences are not drift",
			fromRouter: KronScheduleModel{
				AtTime:    types.StringValue("8:00"),
				DayOfWeek: types.StringValue("mon,wed,fri"),
			},
			desired: &KronScheduleModel{
				AtTime:    types.StringValue("8:00"),
				DayOfWeek: types.StringValue("mon, wed, fri"),
			},
			wantAtTime:    types.StringValue("8:00"),
			wantDate:      types.StringNull(),
			wantName:      types.StringNull(),
			wantDayOfWeek: types.StringValue("mon, wed, fri"),
		},
		{
			// day_of_week is NOT echoed blindly: the write path emits it as the
			// `*/<days>` date token, so the router really does report it back,
			// and echoing would claim success for a weekday schedule the device
			// is running daily.
			name: "a weekday the router did not return stays absent so the drift shows",
			fromRouter: KronScheduleModel{
				AtTime:    types.StringValue("8:00"),
				DayOfWeek: types.StringNull(),
			},
			desired: &KronScheduleModel{
				AtTime:    types.StringValue("8:00"),
				DayOfWeek: types.StringValue("mon-fri"),
			},
			wantAtTime:    types.StringValue("8:00"),
			wantDate:      types.StringNull(),
			wantName:      types.StringNull(),
			wantDayOfWeek: types.StringNull(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fromRouter
			got.reconcileWithDesired(tt.desired)

			if !got.AtTime.Equal(tt.wantAtTime) {
				t.Errorf("AtTime = %v, want %v", got.AtTime, tt.wantAtTime)
			}
			if !got.Date.Equal(tt.wantDate) {
				t.Errorf("Date = %v, want %v", got.Date, tt.wantDate)
			}
			if !got.Name.Equal(tt.wantName) {
				t.Errorf("Name = %v, want %v", got.Name, tt.wantName)
			}
			if !got.DayOfWeek.Equal(tt.wantDayOfWeek) {
				t.Errorf("DayOfWeek = %v, want %v", got.DayOfWeek, tt.wantDayOfWeek)
			}
		})
	}
}

func TestReconcileWithDesiredNilIsNoOp(t *testing.T) {
	m := KronScheduleModel{AtTime: types.StringValue("00:00:00")}
	m.reconcileWithDesired(nil)

	if !m.AtTime.Equal(types.StringValue("00:00:00")) {
		t.Errorf("AtTime = %v, want unchanged", m.AtTime)
	}
}

// The whole point of the round trip: what the practitioner writes for the HND
// ntpdate schedule has to survive Create's read-back under EITHER rendering the
// router might return — the verbatim form or the re-rendered `00:00:00`.
func TestFromClientThenReconcileMatchesPlan(t *testing.T) {
	planned := func() KronScheduleModel {
		return KronScheduleModel{
			ScheduleID:   types.Int64Value(1),
			Name:         types.StringNull(),
			AtTime:       types.StringValue("0:00"),
			DayOfWeek:    types.StringNull(),
			Date:         types.StringNull(),
			Recurring:    types.BoolValue(true),
			OnStartup:    types.BoolValue(false),
			PolicyList:   types.StringNull(),
			CommandLines: fwhelpers.StringSliceToList([]string{"ntpdate ntp.nict.jp syslog"}),
		}
	}

	readBacks := map[string]client.Schedule{
		"router re-renders the clock": {
			ID:        1,
			AtTime:    "00:00:00",
			Recurring: true,
			Commands:  []string{"ntpdate ntp.nict.jp syslog"},
			Enabled:   true,
		},
		"router preserves the written form": {
			ID:        1,
			AtTime:    "0:00",
			Recurring: true,
			Commands:  []string{"ntpdate ntp.nict.jp syslog"},
			Enabled:   true,
		},
	}

	for name, schedule := range readBacks {
		t.Run(name, func(t *testing.T) {
			plan := planned()
			state := planned()

			state.FromClient(&schedule)
			state.reconcileWithDesired(&plan)

			if !state.ScheduleID.Equal(plan.ScheduleID) {
				t.Errorf("schedule_id = %v, want %v", state.ScheduleID, plan.ScheduleID)
			}
			if !state.AtTime.Equal(plan.AtTime) {
				t.Errorf("at_time = %v, want %v", state.AtTime, plan.AtTime)
			}
			if !state.Date.Equal(plan.Date) {
				t.Errorf("date = %v, want %v", state.Date, plan.Date)
			}
			if !state.Name.Equal(plan.Name) {
				t.Errorf("name = %v, want %v", state.Name, plan.Name)
			}
			if !state.DayOfWeek.Equal(plan.DayOfWeek) {
				t.Errorf("day_of_week = %v, want %v", state.DayOfWeek, plan.DayOfWeek)
			}
			if !state.PolicyList.Equal(plan.PolicyList) {
				t.Errorf("policy_list = %v, want %v", state.PolicyList, plan.PolicyList)
			}
			if !state.Recurring.Equal(plan.Recurring) {
				t.Errorf("recurring = %v, want %v", state.Recurring, plan.Recurring)
			}
			if !state.OnStartup.Equal(plan.OnStartup) {
				t.Errorf("on_startup = %v, want %v", state.OnStartup, plan.OnStartup)
			}
			if !state.CommandLines.Equal(plan.CommandLines) {
				t.Errorf("command_lines = %v, want %v", state.CommandLines, plan.CommandLines)
			}
		})
	}
}
