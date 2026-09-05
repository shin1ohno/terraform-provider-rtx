package parsers

import (
	"testing"
)

// The RTX requires an execution-context token between the time and the command
// ("schedule at ID [日付] 時刻 * コマンド..." per RTX1210 Rev.14.01.42
// `schedule at ?`), and it re-renders what was written: a line written as
// "schedule at 1 0:00 * ntpdate ntp.nict.jp syslog" can be printed back by
// `show config` as "schedule at 1 */* 00:00:00 * ntpdate ntp.nict.jp syslog".
// These cases pin both renderings, because the provider has to survive either.
func TestParseScheduleAtRenderedForms(t *testing.T) {
	parser := NewScheduleParser()

	tests := []struct {
		name          string
		line          string
		wantID        int
		wantDate      string
		wantAtTime    string
		wantRecurring bool
		wantOnStartup bool
		wantCommand   string
	}{
		{
			// Captured verbatim from the ITM RTX830 Rev.15.02.31 on 2026-09-05
			// via `show config | grep "schedule at 1"`.
			name:          "daily wildcard date as the router renders it",
			line:          "schedule at 1 */* 00:00:00 * ntpdate ntp.nict.jp syslog",
			wantID:        1,
			wantAtTime:    "00:00:00",
			wantRecurring: true,
			wantCommand:   "ntpdate ntp.nict.jp syslog",
		},
		{
			name:          "same schedule in the form the provider writes",
			line:          "schedule at 1 0:00 * ntpdate ntp.nict.jp syslog",
			wantID:        1,
			wantAtTime:    "0:00",
			wantRecurring: true,
			wantCommand:   "ntpdate ntp.nict.jp syslog",
		},
		{
			name:          "weekday selector stays in date and remains recurring",
			line:          "schedule at 1 */mon-fri 8:00 * pp auth accept on",
			wantID:        1,
			wantDate:      "*/mon-fri",
			wantAtTime:    "8:00",
			wantRecurring: true,
			wantCommand:   "pp auth accept on",
		},
		{
			name:          "month/day repeats every year",
			line:          "schedule at 1 1/1 0:0 * ip route 192.168.0.0/24 gateway pp 2",
			wantID:        1,
			wantDate:      "1/1",
			wantAtTime:    "0:0",
			wantRecurring: true,
			wantCommand:   "ip route 192.168.0.0/24 gateway pp 2",
		},
		{
			name:        "full date is one-time",
			line:        "schedule at 2 2024/12/31 23:59 * restart",
			wantID:      2,
			wantDate:    "2024/12/31",
			wantAtTime:  "23:59",
			wantCommand: "restart",
		},
		{
			name:          "startup form",
			line:          "schedule at 1 startup * console info on",
			wantID:        1,
			wantOnStartup: true,
			wantCommand:   "console info on",
		},
		{
			name:          "wildcard inside the clock",
			line:          "schedule at 1 12:*:00 * lua script.lua",
			wantID:        1,
			wantAtTime:    "12:*:00",
			wantRecurring: true,
			wantCommand:   "lua script.lua",
		},
		{
			name:          "pp context is consumed, not left in the command",
			line:          "schedule at 3 */mon-fri 17:00 pp 1 isdn auto connect off",
			wantID:        3,
			wantDate:      "*/mon-fri",
			wantAtTime:    "17:00",
			wantRecurring: true,
			wantCommand:   "isdn auto connect off",
		},
		{
			name:          "tunnel context is consumed",
			line:          "schedule at 1 8:00 tunnel 1 ipsec sa delete 1",
			wantID:        1,
			wantAtTime:    "8:00",
			wantRecurring: true,
			wantCommand:   "ipsec sa delete 1",
		},
		{
			name:          "switch context by topology path",
			line:          "schedule at 2 */* 03:00 switch lan1:4 switch control function execute restart",
			wantID:        2,
			wantAtTime:    "03:00",
			wantRecurring: true,
			wantCommand:   "switch control function execute restart",
		},
		{
			name:          "switch context by MAC address",
			line:          "schedule at 1 */* 03:00 switch 00:a0:de:01:02:03 switch control function execute restart",
			wantID:        1,
			wantAtTime:    "03:00",
			wantRecurring: true,
			wantCommand:   "switch control function execute restart",
		},
		{
			// A context-less line whose command happens to start with the word
			// `switch` must not have two of its words eaten as a context.
			name:          "command starting with switch and no context token",
			line:          "schedule at 4 12:00 switch control function execute restart",
			wantID:        4,
			wantAtTime:    "12:00",
			wantRecurring: true,
			wantCommand:   "switch control function execute restart",
		},
		{
			// Same guard for `pp`: `pp select 1` is a command, not a context.
			name:          "command starting with pp and no context token",
			line:          "schedule at 5 23:59 pp select 1",
			wantID:        5,
			wantAtTime:    "23:59",
			wantRecurring: true,
			wantCommand:   "pp select 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedules, err := parser.ParseScheduleConfig(tt.line)
			if err != nil {
				t.Fatalf("ParseScheduleConfig() error = %v", err)
			}
			if len(schedules) != 1 {
				t.Fatalf("parsed %d schedules, want 1 (line %q)", len(schedules), tt.line)
			}
			s := schedules[0]
			if s.ID != tt.wantID {
				t.Errorf("ID = %d, want %d", s.ID, tt.wantID)
			}
			if s.Date != tt.wantDate {
				t.Errorf("Date = %q, want %q", s.Date, tt.wantDate)
			}
			if s.AtTime != tt.wantAtTime {
				t.Errorf("AtTime = %q, want %q", s.AtTime, tt.wantAtTime)
			}
			if s.Recurring != tt.wantRecurring {
				t.Errorf("Recurring = %v, want %v", s.Recurring, tt.wantRecurring)
			}
			if s.OnStartup != tt.wantOnStartup {
				t.Errorf("OnStartup = %v, want %v", s.OnStartup, tt.wantOnStartup)
			}
			if len(s.Commands) != 1 || s.Commands[0] != tt.wantCommand {
				t.Errorf("Commands = %q, want [%q]", s.Commands, tt.wantCommand)
			}
			if !s.Enabled {
				t.Errorf("Enabled = false, want true")
			}
		})
	}
}

// `*/*` says nothing the model does not already carry in Recurring, and the
// rtx_kron_schedule `date` attribute cannot hold it (its validator only accepts
// YYYY/MM/DD). Surfacing it would make every plan show a diff the practitioner
// cannot write away, so the parser folds it into Recurring instead.
func TestParseScheduleAtEveryDayDateIsNotSurfaced(t *testing.T) {
	parser := NewScheduleParser()

	schedules, err := parser.ParseScheduleConfig("schedule at 1 */* 03:00 * restart")
	if err != nil {
		t.Fatalf("ParseScheduleConfig() error = %v", err)
	}
	if len(schedules) != 1 {
		t.Fatalf("parsed %d schedules, want 1", len(schedules))
	}
	if schedules[0].Date != "" {
		t.Errorf("Date = %q, want \"\" (the wildcard date must not reach the schema)", schedules[0].Date)
	}
	if !schedules[0].Recurring {
		t.Errorf("Recurring = false, want true")
	}
}

// The `+TIMER` form is valid RTX syntax the model has no field for. It must be
// skipped rather than half-parsed into a schedule with a bogus time.
func TestParseScheduleAtSkipsUnmodelledForms(t *testing.T) {
	parser := NewScheduleParser()

	schedules, err := parser.ParseScheduleConfig("schedule at 1 +600 * restart")
	if err != nil {
		t.Fatalf("ParseScheduleConfig() error = %v", err)
	}
	if len(schedules) != 0 {
		t.Errorf("parsed %+v, want no schedules", schedules)
	}
}

// What the builders emit has to come back out of the parser unchanged, since
// that is the exact path Create takes: write the command, then re-read it.
func TestScheduleBuildThenParse(t *testing.T) {
	parser := NewScheduleParser()

	t.Run("time schedule", func(t *testing.T) {
		line := BuildScheduleAtCommand(1, "0:00", "ntpdate ntp.nict.jp syslog")
		if want := "schedule at 1 0:00 * ntpdate ntp.nict.jp syslog"; line != want {
			t.Fatalf("BuildScheduleAtCommand() = %q, want %q", line, want)
		}
		schedules, err := parser.ParseScheduleConfig(line)
		if err != nil || len(schedules) != 1 {
			t.Fatalf("ParseScheduleConfig(%q) = %+v, %v", line, schedules, err)
		}
		if got := schedules[0].Commands[0]; got != "ntpdate ntp.nict.jp syslog" {
			t.Errorf("Commands[0] = %q, want %q", got, "ntpdate ntp.nict.jp syslog")
		}
		if schedules[0].AtTime != "0:00" {
			t.Errorf("AtTime = %q, want %q", schedules[0].AtTime, "0:00")
		}
	})

	t.Run("startup schedule", func(t *testing.T) {
		line := BuildScheduleAtStartupCommand(2, "console info on")
		if want := "schedule at 2 startup * console info on"; line != want {
			t.Fatalf("BuildScheduleAtStartupCommand() = %q, want %q", line, want)
		}
		schedules, err := parser.ParseScheduleConfig(line)
		if err != nil || len(schedules) != 1 {
			t.Fatalf("ParseScheduleConfig(%q) = %+v, %v", line, schedules, err)
		}
		if !schedules[0].OnStartup || schedules[0].Commands[0] != "console info on" {
			t.Errorf("parsed %+v, want OnStartup with command %q", schedules[0], "console info on")
		}
	})

	t.Run("date/time schedule", func(t *testing.T) {
		line := BuildScheduleAtDateTimeCommand(3, "2025/01/15", "09:00", "save")
		if want := "schedule at 3 2025/01/15 09:00 * save"; line != want {
			t.Fatalf("BuildScheduleAtDateTimeCommand() = %q, want %q", line, want)
		}
		schedules, err := parser.ParseScheduleConfig(line)
		if err != nil || len(schedules) != 1 {
			t.Fatalf("ParseScheduleConfig(%q) = %+v, %v", line, schedules, err)
		}
		if schedules[0].Recurring {
			t.Errorf("Recurring = true, want false for a full date")
		}
		if schedules[0].Commands[0] != "save" {
			t.Errorf("Commands[0] = %q, want %q", schedules[0].Commands[0], "save")
		}
	})
}

func TestNormalizeScheduleTime(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"0:00", "00:00:00"},
		{"00:00:00", "00:00:00"},
		{"9:5", "09:05:00"},
		{"23:59", "23:59:00"},
		{"23:59:59", "23:59:59"},
		// Wildcards carry no numeric value, so they stay as written and only
		// ever compare equal to themselves.
		{"12:*:00", "12:*:00"},
		{"", ""},
	}

	for _, tt := range tests {
		if got := NormalizeScheduleTime(tt.in); got != tt.want {
			t.Errorf("NormalizeScheduleTime(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestNormalizeScheduleDate(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"2025/1/5", "2025/01/05"},
		{"2025/01/05", "2025/01/05"},
		{"*/mon-fri", "*/mon-fri"},
		{"", ""},
	}

	for _, tt := range tests {
		if got := NormalizeScheduleDate(tt.in); got != tt.want {
			t.Errorf("NormalizeScheduleDate(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// The command has to be sliced out of the original line rather than rebuilt by
// joining tokens, or quoted arguments lose their spacing.
func TestParseScheduleAtPreservesCommandSpacing(t *testing.T) {
	parser := NewScheduleParser()

	line := "schedule at 2 0:00 * syslog notice 'nightly  sync'"
	schedules, err := parser.ParseScheduleConfig(line)
	if err != nil || len(schedules) != 1 {
		t.Fatalf("ParseScheduleConfig(%q) = %+v, %v", line, schedules, err)
	}
	if want := "syslog notice 'nightly  sync'"; schedules[0].Commands[0] != want {
		t.Errorf("Commands[0] = %q, want %q", schedules[0].Commands[0], want)
	}
}
