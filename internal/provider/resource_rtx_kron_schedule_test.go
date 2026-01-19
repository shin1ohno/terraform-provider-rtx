package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestResourceRTXKronScheduleSchema(t *testing.T) {
	resource := resourceRTXKronSchedule()

	// Test that the resource has the expected schema fields
	expectedFields := []string{
		"schedule_id",
		"name",
		"at_time",
		"day_of_week",
		"date",
		"recurring",
		"on_startup",
		"policy_list",
		"command_lines",
	}

	for _, field := range expectedFields {
		if resource.Schema[field] == nil {
			t.Errorf("expected %q field in schema", field)
		}
	}
}

func TestResourceRTXKronScheduleSchemaAttributes(t *testing.T) {
	resource := resourceRTXKronSchedule()

	tests := []struct {
		field    string
		expected struct {
			Type     schema.ValueType
			Required bool
			Optional bool
			ForceNew bool
		}
	}{
		{
			field: "schedule_id",
			expected: struct {
				Type     schema.ValueType
				Required bool
				Optional bool
				ForceNew bool
			}{
				Type:     schema.TypeInt,
				Required: true,
				Optional: false,
				ForceNew: true,
			},
		},
		{
			field: "name",
			expected: struct {
				Type     schema.ValueType
				Required bool
				Optional bool
				ForceNew bool
			}{
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
		},
		{
			field: "at_time",
			expected: struct {
				Type     schema.ValueType
				Required bool
				Optional bool
				ForceNew bool
			}{
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
		},
		{
			field: "on_startup",
			expected: struct {
				Type     schema.ValueType
				Required bool
				Optional bool
				ForceNew bool
			}{
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
		},
		{
			field: "recurring",
			expected: struct {
				Type     schema.ValueType
				Required bool
				Optional bool
				ForceNew bool
			}{
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				ForceNew: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			s := resource.Schema[tt.field]
			if s == nil {
				t.Fatalf("field %q not found in schema", tt.field)
			}

			if s.Type != tt.expected.Type {
				t.Errorf("field %q: Type = %v, want %v", tt.field, s.Type, tt.expected.Type)
			}
			if s.Required != tt.expected.Required {
				t.Errorf("field %q: Required = %v, want %v", tt.field, s.Required, tt.expected.Required)
			}
			if s.Optional != tt.expected.Optional {
				t.Errorf("field %q: Optional = %v, want %v", tt.field, s.Optional, tt.expected.Optional)
			}
			if s.ForceNew != tt.expected.ForceNew {
				t.Errorf("field %q: ForceNew = %v, want %v", tt.field, s.ForceNew, tt.expected.ForceNew)
			}
		})
	}
}

func TestValidateTimeFormatFunction(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid 12:00", value: "12:00", wantErr: false},
		{name: "valid 0:00", value: "0:00", wantErr: false},
		{name: "valid 23:59", value: "23:59", wantErr: false},
		{name: "valid 6:30", value: "6:30", wantErr: false},
		{name: "valid midnight", value: "00:00", wantErr: false},
		{name: "invalid hour 24", value: "24:00", wantErr: true},
		{name: "invalid hour 25", value: "25:00", wantErr: true},
		{name: "invalid minute 60", value: "12:60", wantErr: true},
		{name: "invalid format dash", value: "12-00", wantErr: true},
		{name: "invalid format no colon", value: "1200", wantErr: true},
		{name: "invalid format extra", value: "12:00:00", wantErr: true},
		{name: "empty allowed", value: "", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errors := validateTimeFormat(tt.value, "at_time")

			if tt.wantErr && len(errors) == 0 {
				t.Errorf("expected error for value %q, got none", tt.value)
			}
			if !tt.wantErr && len(errors) > 0 {
				t.Errorf("unexpected error for value %q: %v", tt.value, errors)
			}
		})
	}
}

func TestValidateDateFormatFunction(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "valid date", value: "2025/01/15", wantErr: false},
		{name: "valid year end", value: "2025/12/31", wantErr: false},
		{name: "valid year start", value: "2025/01/01", wantErr: false},
		{name: "valid 2099", value: "2099/12/31", wantErr: false},
		{name: "valid 2000", value: "2000/01/01", wantErr: false},
		{name: "invalid month 13", value: "2025/13/01", wantErr: true},
		{name: "invalid month 0", value: "2025/00/01", wantErr: true},
		{name: "invalid day 32", value: "2025/01/32", wantErr: true},
		{name: "invalid day 0", value: "2025/01/00", wantErr: true},
		{name: "invalid year 1999", value: "1999/01/01", wantErr: true},
		{name: "invalid year 2100", value: "2100/01/01", wantErr: true},
		{name: "invalid format dash", value: "2025-01-15", wantErr: true},
		{name: "invalid format no separator", value: "20250115", wantErr: true},
		{name: "empty allowed", value: "", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errors := validateDateFormat(tt.value, "date")

			if tt.wantErr && len(errors) == 0 {
				t.Errorf("expected error for value %q, got none", tt.value)
			}
			if !tt.wantErr && len(errors) > 0 {
				t.Errorf("unexpected error for value %q: %v", tt.value, errors)
			}
		})
	}
}

func TestValidateDayOfWeekFunction(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "single day mon", value: "mon", wantErr: false},
		{name: "single day sun", value: "sun", wantErr: false},
		{name: "single day tue", value: "tue", wantErr: false},
		{name: "single day wed", value: "wed", wantErr: false},
		{name: "single day thu", value: "thu", wantErr: false},
		{name: "single day fri", value: "fri", wantErr: false},
		{name: "single day sat", value: "sat", wantErr: false},
		{name: "range mon-fri", value: "mon-fri", wantErr: false},
		{name: "range sat-sun", value: "sat-sun", wantErr: false},
		{name: "range sun-sat", value: "sun-sat", wantErr: false},
		{name: "comma list", value: "mon,wed,fri", wantErr: false},
		{name: "comma with spaces", value: "mon, wed, fri", wantErr: false},
		{name: "weekend", value: "sat,sun", wantErr: false},
		{name: "invalid day xyz", value: "xyz", wantErr: true},
		{name: "invalid range end", value: "mon-xyz", wantErr: true},
		{name: "invalid range start", value: "xyz-fri", wantErr: true},
		{name: "invalid in list", value: "mon,xyz,fri", wantErr: true},
		{name: "empty allowed", value: "", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, errors := validateDayOfWeek(tt.value, "day_of_week")

			if tt.wantErr && len(errors) == 0 {
				t.Errorf("expected error for value %q, got none", tt.value)
			}
			if !tt.wantErr && len(errors) > 0 {
				t.Errorf("unexpected error for value %q: %v", tt.value, errors)
			}
		})
	}
}

func TestResourceRTXKronScheduleConflictsWith(t *testing.T) {
	resource := resourceRTXKronSchedule()

	// Check at_time conflicts
	atTimeConflicts := resource.Schema["at_time"].ConflictsWith
	if len(atTimeConflicts) == 0 {
		t.Error("expected at_time to have ConflictsWith")
	}
	foundOnStartup := false
	for _, c := range atTimeConflicts {
		if c == "on_startup" {
			foundOnStartup = true
			break
		}
	}
	if !foundOnStartup {
		t.Error("expected at_time to conflict with on_startup")
	}

	// Check on_startup conflicts
	onStartupConflicts := resource.Schema["on_startup"].ConflictsWith
	expectedConflicts := []string{"at_time", "date", "day_of_week"}
	for _, expected := range expectedConflicts {
		found := false
		for _, c := range onStartupConflicts {
			if c == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected on_startup to conflict with %q", expected)
		}
	}

	// Check policy_list conflicts with command_lines
	policyListConflicts := resource.Schema["policy_list"].ConflictsWith
	foundCommandLines := false
	for _, c := range policyListConflicts {
		if c == "command_lines" {
			foundCommandLines = true
			break
		}
	}
	if !foundCommandLines {
		t.Error("expected policy_list to conflict with command_lines")
	}
}
