package client

import (
	"context"
	"strings"
	"testing"
)

// The command CreateSchedule puts on the wire is the whole bug this fix is
// about, so pin it per schedule shape.
func TestCreateScheduleCommandForm(t *testing.T) {
	tests := []struct {
		name     string
		schedule Schedule
		want     string
	}{
		{
			// The HND ntpdate schedule. `*` is the execution-context token the
			// RTX requires between the time and the command.
			name: "daily time schedule",
			schedule: Schedule{
				ID: 1, AtTime: "0:00", Recurring: true,
				Commands: []string{"ntpdate ntp.nict.jp syslog"},
			},
			want: "schedule at 1 0:00 * ntpdate ntp.nict.jp syslog",
		},
		{
			// The RTX has no weekday field — weekdays go in the date slot.
			// Before this branch existed the day was dropped and the schedule
			// silently became a daily one.
			name: "weekday schedule becomes a wildcard date",
			schedule: Schedule{
				ID: 2, AtTime: "8:00", DayOfWeek: "mon-fri", Recurring: true,
				Commands: []string{"pp auth accept on"},
			},
			want: "schedule at 2 */mon-fri 8:00 * pp auth accept on",
		},
		{
			name: "weekday spelling is normalized before it reaches the router",
			schedule: Schedule{
				ID: 3, AtTime: "9:00", DayOfWeek: "Sat, Sun", Recurring: true,
				Commands: []string{"nat masquerade on"},
			},
			want: "schedule at 3 */sat,sun 9:00 * nat masquerade on",
		},
		{
			name: "one-time date schedule",
			schedule: Schedule{
				ID: 4, AtTime: "23:59", Date: "2025/12/31",
				Commands: []string{"restart"},
			},
			want: "schedule at 4 2025/12/31 23:59 * restart",
		},
		{
			name: "startup schedule",
			schedule: Schedule{
				ID: 5, OnStartup: true,
				Commands: []string{"console info on"},
			},
			want: "schedule at 5 startup * console info on",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The mock matches on substring, so key the read fixture on
			// "show config" — keying it on "schedule at" would also swallow the
			// write command.
			exec := &mockExecutor{responses: map[string]string{"show config": ""}}
			svc := NewScheduleService(exec, nil)

			if err := svc.CreateSchedule(context.Background(), tt.schedule); err != nil {
				t.Fatalf("CreateSchedule() error = %v", err)
			}

			var found bool
			for _, cmd := range exec.executedCmds {
				if cmd == tt.want {
					found = true
				}
			}
			if !found {
				t.Errorf("commands issued = %q, want one of them to be %q", exec.executedCmds, tt.want)
			}
		})
	}
}

// A create against a router that already carries the identical line has to be
// adopted, not refused: an apply that wrote the line and then failed
// Terraform's post-apply consistency check leaves exactly that state, and
// refusing would strand it until someone hand-ran `no schedule at <id>`.
func TestCreateScheduleAdoptsIdenticalLine(t *testing.T) {
	desired := Schedule{
		ID: 1, AtTime: "0:00", Recurring: true,
		Commands: []string{"ntpdate ntp.nict.jp syslog"},
	}

	t.Run("identical line is adopted", func(t *testing.T) {
		exec := &mockExecutor{responses: map[string]string{
			// The router's own rendering of the same schedule.
			"show config": "schedule at 1 */* 00:00:00 * ntpdate ntp.nict.jp syslog\n",
		}}
		svc := NewScheduleService(exec, nil)

		if err := svc.CreateSchedule(context.Background(), desired); err != nil {
			t.Fatalf("CreateSchedule() error = %v, want nil", err)
		}
	})

	t.Run("a different line on the same ID is still a conflict", func(t *testing.T) {
		exec := &mockExecutor{responses: map[string]string{
			"show config": "schedule at 1 */* 03:00:00 * restart\n",
		}}
		svc := NewScheduleService(exec, nil)

		err := svc.CreateSchedule(context.Background(), desired)
		if err == nil || !strings.Contains(err.Error(), "already exists") {
			t.Fatalf("CreateSchedule() error = %v, want an 'already exists' error", err)
		}
	})
}
