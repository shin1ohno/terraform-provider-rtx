package parsers

import (
	"strings"
	"testing"
)

func TestParseScheduleConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Schedule
		wantErr  bool
	}{
		{
			name:  "basic time schedule",
			input: `schedule at 1 12:00 show ip route`,
			expected: []Schedule{
				{
					ID:        1,
					AtTime:    "12:00",
					Commands:  []string{"show ip route"},
					Recurring: true,
					Enabled:   true,
				},
			},
		},
		{
			name:  "startup schedule",
			input: `schedule at 2 startup pp select 1`,
			expected: []Schedule{
				{
					ID:        2,
					OnStartup: true,
					Commands:  []string{"pp select 1"},
					Recurring: false,
					Enabled:   true,
				},
			},
		},
		{
			name:  "date/time schedule",
			input: `schedule at 3 2025/01/15 09:00 save`,
			expected: []Schedule{
				{
					ID:        3,
					Date:      "2025/01/15",
					AtTime:    "09:00",
					Commands:  []string{"save"},
					Recurring: false,
					Enabled:   true,
				},
			},
		},
		{
			name:  "PP schedule with day of week",
			input: `schedule pp 1 mon-fri 8:00 connect`,
			expected: []Schedule{
				{
					ID:        -1,
					DayOfWeek: "mon-fri",
					AtTime:    "8:00",
					Commands:  []string{"connect"},
					Recurring: true,
					Enabled:   true,
				},
			},
		},
		{
			name: "multiple schedules",
			input: `schedule at 1 6:00 show ip route
schedule at 2 startup dhcp service server
schedule at 3 2025/12/31 23:59 save`,
			expected: []Schedule{
				{
					ID:        1,
					AtTime:    "6:00",
					Commands:  []string{"show ip route"},
					Recurring: true,
					Enabled:   true,
				},
				{
					ID:        2,
					OnStartup: true,
					Commands:  []string{"dhcp service server"},
					Recurring: false,
					Enabled:   true,
				},
				{
					ID:        3,
					Date:      "2025/12/31",
					AtTime:    "23:59",
					Commands:  []string{"save"},
					Recurring: false,
					Enabled:   true,
				},
			},
		},
		{
			name: "schedule with multiple commands (same ID)",
			input: `schedule at 1 12:00 show ip route
schedule at 1 12:00 show ip interface`,
			expected: []Schedule{
				{
					ID:        1,
					AtTime:    "12:00",
					Commands:  []string{"show ip route", "show ip interface"},
					Recurring: true,
					Enabled:   true,
				},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []Schedule{},
		},
		{
			name:     "no schedule config",
			input:    "ip lan1 address 192.168.1.1/24\nsome other config",
			expected: []Schedule{},
		},
		{
			name: "schedule with comments",
			input: `# Daily backup
schedule at 1 3:00 save
# Weekly maintenance
schedule at 2 startup pp select 1`,
			expected: []Schedule{
				{
					ID:        1,
					AtTime:    "3:00",
					Commands:  []string{"save"},
					Recurring: true,
					Enabled:   true,
				},
				{
					ID:        2,
					OnStartup: true,
					Commands:  []string{"pp select 1"},
					Recurring: false,
					Enabled:   true,
				},
			},
		},
	}

	parser := NewScheduleParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseScheduleConfig(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d schedules, got %d", len(tt.expected), len(result))
				return
			}

			// Create a map for easier comparison (order may vary)
			resultMap := make(map[int]Schedule)
			for _, s := range result {
				resultMap[s.ID] = s
			}

			for _, expected := range tt.expected {
				got, ok := resultMap[expected.ID]
				if !ok {
					t.Errorf("schedule %d not found in result", expected.ID)
					continue
				}

				if got.AtTime != expected.AtTime {
					t.Errorf("schedule %d: at_time = %q, want %q", expected.ID, got.AtTime, expected.AtTime)
				}
				if got.Date != expected.Date {
					t.Errorf("schedule %d: date = %q, want %q", expected.ID, got.Date, expected.Date)
				}
				if got.DayOfWeek != expected.DayOfWeek {
					t.Errorf("schedule %d: day_of_week = %q, want %q", expected.ID, got.DayOfWeek, expected.DayOfWeek)
				}
				if got.OnStartup != expected.OnStartup {
					t.Errorf("schedule %d: on_startup = %v, want %v", expected.ID, got.OnStartup, expected.OnStartup)
				}
				if got.Recurring != expected.Recurring {
					t.Errorf("schedule %d: recurring = %v, want %v", expected.ID, got.Recurring, expected.Recurring)
				}
				if got.Enabled != expected.Enabled {
					t.Errorf("schedule %d: enabled = %v, want %v", expected.ID, got.Enabled, expected.Enabled)
				}
				if len(got.Commands) != len(expected.Commands) {
					t.Errorf("schedule %d: commands count = %d, want %d", expected.ID, len(got.Commands), len(expected.Commands))
				} else {
					for i, cmd := range expected.Commands {
						if got.Commands[i] != cmd {
							t.Errorf("schedule %d: command[%d] = %q, want %q", expected.ID, i, got.Commands[i], cmd)
						}
					}
				}
			}
		})
	}
}

func TestBuildScheduleAtCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		time     string
		command  string
		expected string
	}{
		{
			name:     "basic daily schedule",
			id:       1,
			time:     "12:00",
			command:  "save",
			expected: "schedule at 1 12:00 save",
		},
		{
			name:     "early morning schedule",
			id:       5,
			time:     "6:30",
			command:  "show ip route",
			expected: "schedule at 5 6:30 show ip route",
		},
		{
			name:     "late night schedule",
			id:       10,
			time:     "23:59",
			command:  "pp select 1",
			expected: "schedule at 10 23:59 pp select 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildScheduleAtCommand(tt.id, tt.time, tt.command)
			if result != tt.expected {
				t.Errorf("BuildScheduleAtCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildScheduleAtStartupCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		command  string
		expected string
	}{
		{
			name:     "startup with simple command",
			id:       1,
			command:  "dhcp service server",
			expected: "schedule at 1 startup dhcp service server",
		},
		{
			name:     "startup with pp select",
			id:       2,
			command:  "pp select 1",
			expected: "schedule at 2 startup pp select 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildScheduleAtStartupCommand(tt.id, tt.command)
			if result != tt.expected {
				t.Errorf("BuildScheduleAtStartupCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildScheduleAtDateTimeCommand(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		date     string
		time     string
		command  string
		expected string
	}{
		{
			name:     "one-time schedule",
			id:       1,
			date:     "2025/01/15",
			time:     "09:00",
			command:  "save",
			expected: "schedule at 1 2025/01/15 09:00 save",
		},
		{
			name:     "year-end schedule",
			id:       10,
			date:     "2025/12/31",
			time:     "23:59",
			command:  "restart",
			expected: "schedule at 10 2025/12/31 23:59 restart",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildScheduleAtDateTimeCommand(tt.id, tt.date, tt.time, tt.command)
			if result != tt.expected {
				t.Errorf("BuildScheduleAtDateTimeCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildSchedulePPCommand(t *testing.T) {
	tests := []struct {
		name      string
		ppNum     int
		dayOfWeek string
		time      string
		action    string
		expected  string
	}{
		{
			name:      "weekday connect",
			ppNum:     1,
			dayOfWeek: "mon-fri",
			time:      "8:00",
			action:    "connect",
			expected:  "schedule pp 1 mon-fri 8:00 connect",
		},
		{
			name:      "weekend disconnect",
			ppNum:     1,
			dayOfWeek: "sat,sun",
			time:      "22:00",
			action:    "disconnect",
			expected:  "schedule pp 1 sat,sun 22:00 disconnect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildSchedulePPCommand(tt.ppNum, tt.dayOfWeek, tt.time, tt.action)
			if result != tt.expected {
				t.Errorf("BuildSchedulePPCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildDeleteScheduleCommand(t *testing.T) {
	result := BuildDeleteScheduleCommand(5)
	expected := "no schedule at 5"
	if result != expected {
		t.Errorf("BuildDeleteScheduleCommand() = %q, want %q", result, expected)
	}
}

func TestBuildDeleteSchedulePPCommand(t *testing.T) {
	result := BuildDeleteSchedulePPCommand(1, "mon-fri", "8:00")
	expected := "no schedule pp 1 mon-fri 8:00"
	if result != expected {
		t.Errorf("BuildDeleteSchedulePPCommand() = %q, want %q", result, expected)
	}
}

func TestValidateTimeFormat(t *testing.T) {
	tests := []struct {
		name    string
		time    string
		wantErr bool
		errMsg  string
	}{
		{name: "valid 12:00", time: "12:00", wantErr: false},
		{name: "valid 0:00", time: "0:00", wantErr: false},
		{name: "valid 23:59", time: "23:59", wantErr: false},
		{name: "valid 6:30", time: "6:30", wantErr: false},
		{name: "invalid hour 25", time: "25:00", wantErr: true, errMsg: "invalid hour"},
		{name: "invalid minute 60", time: "12:60", wantErr: true, errMsg: "invalid minute"},
		{name: "invalid format", time: "12-00", wantErr: true, errMsg: "invalid time format"},
		{name: "empty", time: "", wantErr: true, errMsg: "invalid time format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeFormat(tt.time)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateDateFormat(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		wantErr bool
		errMsg  string
	}{
		{name: "valid date", date: "2025/01/15", wantErr: false},
		{name: "valid year end", date: "2025/12/31", wantErr: false},
		{name: "valid year start", date: "2025/01/01", wantErr: false},
		{name: "invalid month 13", date: "2025/13/01", wantErr: true, errMsg: "invalid month"},
		{name: "invalid month 0", date: "2025/00/01", wantErr: true, errMsg: "invalid month"},
		{name: "invalid day 32", date: "2025/01/32", wantErr: true, errMsg: "invalid day"},
		{name: "invalid day 0", date: "2025/01/00", wantErr: true, errMsg: "invalid day"},
		{name: "invalid year 1999", date: "1999/01/01", wantErr: true, errMsg: "invalid year"},
		{name: "invalid format", date: "2025-01-15", wantErr: true, errMsg: "invalid date format"},
		{name: "empty", date: "", wantErr: true, errMsg: "invalid date format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDateFormat(tt.date)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateDayOfWeek(t *testing.T) {
	tests := []struct {
		name    string
		day     string
		wantErr bool
		errMsg  string
	}{
		{name: "single day mon", day: "mon", wantErr: false},
		{name: "single day sun", day: "sun", wantErr: false},
		{name: "range mon-fri", day: "mon-fri", wantErr: false},
		{name: "range sat-sun", day: "sat-sun", wantErr: false},
		{name: "comma list", day: "mon,wed,fri", wantErr: false},
		{name: "comma with spaces", day: "mon, wed, fri", wantErr: false},
		{name: "invalid day", day: "xyz", wantErr: true, errMsg: "invalid day"},
		{name: "invalid range", day: "mon-xyz", wantErr: true, errMsg: "invalid day"},
		{name: "invalid in list", day: "mon,xyz,fri", wantErr: true, errMsg: "invalid day"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDayOfWeek(tt.day)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateSchedule(t *testing.T) {
	tests := []struct {
		name     string
		schedule Schedule
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid time schedule",
			schedule: Schedule{
				ID:       1,
				AtTime:   "12:00",
				Commands: []string{"save"},
			},
			wantErr: false,
		},
		{
			name: "valid startup schedule",
			schedule: Schedule{
				ID:        1,
				OnStartup: true,
				Commands:  []string{"pp select 1"},
			},
			wantErr: false,
		},
		{
			name: "valid date/time schedule",
			schedule: Schedule{
				ID:       1,
				Date:     "2025/01/15",
				AtTime:   "09:00",
				Commands: []string{"save"},
			},
			wantErr: false,
		},
		{
			name: "invalid ID 0",
			schedule: Schedule{
				ID:       0,
				AtTime:   "12:00",
				Commands: []string{"save"},
			},
			wantErr: true,
			errMsg:  "schedule id must be between 1 and 65535",
		},
		{
			name: "invalid ID too high",
			schedule: Schedule{
				ID:       65536,
				AtTime:   "12:00",
				Commands: []string{"save"},
			},
			wantErr: true,
			errMsg:  "schedule id must be between 1 and 65535",
		},
		{
			name: "invalid time format",
			schedule: Schedule{
				ID:       1,
				AtTime:   "25:00",
				Commands: []string{"save"},
			},
			wantErr: true,
			errMsg:  "invalid hour",
		},
		{
			name: "invalid date format",
			schedule: Schedule{
				ID:       1,
				Date:     "invalid",
				Commands: []string{"save"},
			},
			wantErr: true,
			errMsg:  "invalid date format",
		},
		{
			name: "missing time/startup/date",
			schedule: Schedule{
				ID:       1,
				Commands: []string{"save"},
			},
			wantErr: true,
			errMsg:  "must have at_time, on_startup, or date specified",
		},
		{
			name: "startup with time conflict",
			schedule: Schedule{
				ID:        1,
				OnStartup: true,
				AtTime:    "12:00",
				Commands:  []string{"save"},
			},
			wantErr: true,
			errMsg:  "on_startup cannot be combined with at_time or date",
		},
		{
			name: "no commands",
			schedule: Schedule{
				ID:     1,
				AtTime: "12:00",
			},
			wantErr: true,
			errMsg:  "must have at least one command or policy_list",
		},
		{
			name: "valid with policy_list",
			schedule: Schedule{
				ID:         1,
				AtTime:     "12:00",
				PolicyList: "daily_backup",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSchedule(tt.schedule)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateKronPolicy(t *testing.T) {
	tests := []struct {
		name    string
		policy  KronPolicy
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid policy",
			policy: KronPolicy{
				Name:     "daily_backup",
				Commands: []string{"save", "show ip route"},
			},
			wantErr: false,
		},
		{
			name: "valid policy with hyphen",
			policy: KronPolicy{
				Name:     "my-policy-1",
				Commands: []string{"save"},
			},
			wantErr: false,
		},
		{
			name: "valid policy with underscore",
			policy: KronPolicy{
				Name:     "my_policy_2",
				Commands: []string{"save"},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			policy: KronPolicy{
				Name:     "",
				Commands: []string{"save"},
			},
			wantErr: true,
			errMsg:  "policy name is required",
		},
		{
			name: "name starts with number",
			policy: KronPolicy{
				Name:     "1policy",
				Commands: []string{"save"},
			},
			wantErr: true,
			errMsg:  "must start with a letter",
		},
		{
			name: "name with spaces",
			policy: KronPolicy{
				Name:     "my policy",
				Commands: []string{"save"},
			},
			wantErr: true,
			errMsg:  "must start with a letter",
		},
		{
			name: "no commands",
			policy: KronPolicy{
				Name:     "empty_policy",
				Commands: []string{},
			},
			wantErr: true,
			errMsg:  "must have at least one command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKronPolicy(tt.policy)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseSingleSchedule(t *testing.T) {
	parser := NewScheduleParser()

	input := `schedule at 1 12:00 save
schedule at 2 startup dhcp service server
schedule at 3 2025/01/15 09:00 restart`

	// Test finding existing schedule
	schedule, err := parser.ParseSingleSchedule(input, 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if schedule == nil {
		t.Fatal("expected schedule, got nil")
	}
	if schedule.ID != 1 {
		t.Errorf("ID = %d, want 1", schedule.ID)
	}
	if schedule.AtTime != "12:00" {
		t.Errorf("AtTime = %q, want %q", schedule.AtTime, "12:00")
	}

	// Test finding startup schedule
	schedule, err = parser.ParseSingleSchedule(input, 2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !schedule.OnStartup {
		t.Error("expected OnStartup = true")
	}

	// Test not found
	_, err = parser.ParseSingleSchedule(input, 99)
	if err == nil {
		t.Error("expected error for non-existent schedule, got nil")
	}
}

func TestBuildShowScheduleCommand(t *testing.T) {
	result := BuildShowScheduleCommand()
	expected := "show config | grep schedule"
	if result != expected {
		t.Errorf("BuildShowScheduleCommand() = %q, want %q", result, expected)
	}
}

func TestBuildShowScheduleByIDCommand(t *testing.T) {
	result := BuildShowScheduleByIDCommand(5)
	expected := `show config | grep "schedule at 5"`
	if result != expected {
		t.Errorf("BuildShowScheduleByIDCommand() = %q, want %q", result, expected)
	}
}
