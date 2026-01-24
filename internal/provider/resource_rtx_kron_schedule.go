package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/sh1/terraform-provider-rtx/internal/client"
)

func resourceRTXKronSchedule() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a kron schedule (scheduled task) on RTX routers. Schedules can be time-based (daily), startup-based, or date-specific (one-time).",
		CreateContext: resourceRTXKronScheduleCreate,
		ReadContext:   resourceRTXKronScheduleRead,
		UpdateContext: resourceRTXKronScheduleUpdate,
		DeleteContext: resourceRTXKronScheduleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRTXKronScheduleImport,
		},

		Schema: map[string]*schema.Schema{
			"schedule_id": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				Description:  "The schedule ID (1-65535). Must be unique across all schedules on the router.",
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional name/description for the schedule.",
			},
			"at_time": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Time to execute the schedule in HH:MM format (24-hour). Required unless on_startup is true.",
				ValidateFunc:  validateTimeFormat,
				ConflictsWith: []string{"on_startup"},
			},
			"day_of_week": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Day(s) of week to execute the schedule. Examples: 'mon', 'mon-fri', 'sat,sun'. If not specified with at_time, schedule runs daily.",
				ValidateFunc: validateDayOfWeek,
			},
			"date": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Specific date for one-time schedule execution in YYYY/MM/DD format. Cannot be combined with on_startup or day_of_week.",
				ValidateFunc:  validateDateFormat,
				ConflictsWith: []string{"on_startup", "day_of_week"},
			},
			"recurring": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether the schedule repeats. Automatically set to false for date-specific schedules.",
			},
			"on_startup": {
				Type:          schema.TypeBool,
				Optional:      true,
				Computed:      true,
				Description:   "Execute this schedule when the router starts up. Cannot be combined with at_time or date.",
				ConflictsWith: []string{"at_time", "date", "day_of_week"},
			},
			"policy_list": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Name of a kron policy to execute. Use this OR command_lines, not both.",
				ConflictsWith: []string{"command_lines"},
			},
			"command_lines": {
				Type:          schema.TypeList,
				Optional:      true,
				Description:   "List of commands to execute. Use this OR policy_list, not both.",
				ConflictsWith: []string{"policy_list"},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},

		CustomizeDiff: func(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
			// Validate that either at_time, on_startup, or date is specified
			atTime := diff.Get("at_time").(string)
			onStartup := diff.Get("on_startup").(bool)
			date := diff.Get("date").(string)

			if atTime == "" && !onStartup && date == "" {
				return fmt.Errorf("one of 'at_time', 'on_startup', or 'date' must be specified")
			}

			// Validate that either policy_list or command_lines is specified
			policyList := diff.Get("policy_list").(string)
			commandLines, _ := diff.GetOk("command_lines")
			commands := commandLines.([]interface{})

			if policyList == "" && len(commands) == 0 {
				return fmt.Errorf("either 'policy_list' or 'command_lines' must be specified")
			}

			// Auto-set recurring to false for date-specific schedules
			if date != "" {
				if err := diff.SetNew("recurring", false); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func resourceRTXKronScheduleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_kron_schedule", d.Id())
	schedule := buildScheduleFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_kron_schedule").Msgf("Creating kron schedule: %+v", schedule)

	err := apiClient.client.CreateSchedule(ctx, schedule)
	if err != nil {
		return diag.Errorf("Failed to create kron schedule: %v", err)
	}

	d.SetId(strconv.Itoa(schedule.ID))

	return resourceRTXKronScheduleRead(ctx, d, meta)
}

func resourceRTXKronScheduleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_kron_schedule", d.Id())
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid schedule ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_kron_schedule").Msgf("Reading kron schedule: %d", id)

	schedule, err := apiClient.client.GetSchedule(ctx, id)
	if err != nil {
		// Check if schedule doesn't exist
		if strings.Contains(err.Error(), "not found") {
			logging.FromContext(ctx).Debug().Str("resource", "rtx_kron_schedule").Msgf("Kron schedule %d not found, removing from state", id)
			d.SetId("")
			return nil
		}
		return diag.Errorf("Failed to read kron schedule: %v", err)
	}

	// Update the state
	if err := d.Set("schedule_id", schedule.ID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", schedule.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("at_time", schedule.AtTime); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("day_of_week", schedule.DayOfWeek); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("date", schedule.Date); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("recurring", schedule.Recurring); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("on_startup", schedule.OnStartup); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("policy_list", schedule.PolicyList); err != nil {
		return diag.FromErr(err)
	}
	if len(schedule.Commands) > 0 {
		if err := d.Set("command_lines", schedule.Commands); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceRTXKronScheduleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_kron_schedule", d.Id())
	schedule := buildScheduleFromResourceData(d)

	logging.FromContext(ctx).Debug().Str("resource", "rtx_kron_schedule").Msgf("Updating kron schedule: %+v", schedule)

	err := apiClient.client.UpdateSchedule(ctx, schedule)
	if err != nil {
		return diag.Errorf("Failed to update kron schedule: %v", err)
	}

	return resourceRTXKronScheduleRead(ctx, d, meta)
}

func resourceRTXKronScheduleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiClient := meta.(*apiClient)

	// Add resource context for command logging
	ctx = logging.WithResource(ctx, "rtx_kron_schedule", d.Id())
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.Errorf("Invalid schedule ID: %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_kron_schedule").Msgf("Deleting kron schedule: %d", id)

	err = apiClient.client.DeleteSchedule(ctx, id)
	if err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		return diag.Errorf("Failed to delete kron schedule: %v", err)
	}

	return nil
}

func resourceRTXKronScheduleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiClient := meta.(*apiClient)
	importID := d.Id()

	// Parse import ID as schedule ID
	id, err := strconv.Atoi(importID)
	if err != nil {
		return nil, fmt.Errorf("invalid import ID format, expected schedule ID (integer): %v", err)
	}

	logging.FromContext(ctx).Debug().Str("resource", "rtx_kron_schedule").Msgf("Importing kron schedule: %d", id)

	// Verify schedule exists
	schedule, err := apiClient.client.GetSchedule(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to import kron schedule %d: %v", id, err)
	}

	// Set all attributes
	d.SetId(strconv.Itoa(id))
	d.Set("schedule_id", schedule.ID)
	d.Set("name", schedule.Name)
	d.Set("at_time", schedule.AtTime)
	d.Set("day_of_week", schedule.DayOfWeek)
	d.Set("date", schedule.Date)
	d.Set("recurring", schedule.Recurring)
	d.Set("on_startup", schedule.OnStartup)
	d.Set("policy_list", schedule.PolicyList)
	if len(schedule.Commands) > 0 {
		d.Set("command_lines", schedule.Commands)
	}

	return []*schema.ResourceData{d}, nil
}

// buildScheduleFromResourceData creates a Schedule from Terraform resource data
func buildScheduleFromResourceData(d *schema.ResourceData) client.Schedule {
	schedule := client.Schedule{
		ID:        d.Get("schedule_id").(int),
		Recurring: d.Get("recurring").(bool),
		OnStartup: d.Get("on_startup").(bool),
		Enabled:   true, // Always enabled when managed by Terraform
	}

	if v, ok := d.GetOk("name"); ok {
		schedule.Name = v.(string)
	}

	if v, ok := d.GetOk("at_time"); ok {
		schedule.AtTime = v.(string)
	}

	if v, ok := d.GetOk("day_of_week"); ok {
		schedule.DayOfWeek = v.(string)
	}

	if v, ok := d.GetOk("date"); ok {
		schedule.Date = v.(string)
		schedule.Recurring = false // Date-specific schedules are one-time
	}

	if v, ok := d.GetOk("policy_list"); ok {
		schedule.PolicyList = v.(string)
	}

	if v, ok := d.GetOk("command_lines"); ok {
		commandsRaw := v.([]interface{})
		commands := make([]string, len(commandsRaw))
		for i, cmd := range commandsRaw {
			commands[i] = cmd.(string)
		}
		schedule.Commands = commands
	}

	return schedule
}

// validateTimeFormat validates a time string in HH:MM format
func validateTimeFormat(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, nil
	}

	timePattern := regexp.MustCompile(`^(\d{1,2}):(\d{2})$`)
	matches := timePattern.FindStringSubmatch(value)
	if len(matches) != 3 {
		return nil, []error{fmt.Errorf("%q must be in HH:MM format (e.g., '12:00', '6:30'), got %q", k, value)}
	}

	hour, _ := strconv.Atoi(matches[1])
	minute, _ := strconv.Atoi(matches[2])

	if hour < 0 || hour > 23 {
		return nil, []error{fmt.Errorf("%q has invalid hour %d, must be 0-23", k, hour)}
	}
	if minute < 0 || minute > 59 {
		return nil, []error{fmt.Errorf("%q has invalid minute %d, must be 0-59", k, minute)}
	}

	return nil, nil
}

// validateDateFormat validates a date string in YYYY/MM/DD format
func validateDateFormat(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, nil
	}

	datePattern := regexp.MustCompile(`^(\d{4})/(\d{2})/(\d{2})$`)
	matches := datePattern.FindStringSubmatch(value)
	if len(matches) != 4 {
		return nil, []error{fmt.Errorf("%q must be in YYYY/MM/DD format (e.g., '2025/01/15'), got %q", k, value)}
	}

	year, _ := strconv.Atoi(matches[1])
	month, _ := strconv.Atoi(matches[2])
	day, _ := strconv.Atoi(matches[3])

	if year < 2000 || year > 2099 {
		return nil, []error{fmt.Errorf("%q has invalid year %d, must be 2000-2099", k, year)}
	}
	if month < 1 || month > 12 {
		return nil, []error{fmt.Errorf("%q has invalid month %d, must be 1-12", k, month)}
	}
	if day < 1 || day > 31 {
		return nil, []error{fmt.Errorf("%q has invalid day %d, must be 1-31", k, day)}
	}

	return nil, nil
}

// validateDayOfWeek validates a day of week specification
func validateDayOfWeek(v interface{}, k string) ([]string, []error) {
	value := v.(string)

	if value == "" {
		return nil, nil
	}

	validDays := map[string]bool{
		"sun": true, "mon": true, "tue": true, "wed": true,
		"thu": true, "fri": true, "sat": true,
	}

	// Handle range format (e.g., "mon-fri")
	if strings.Contains(value, "-") {
		parts := strings.Split(value, "-")
		if len(parts) != 2 {
			return nil, []error{fmt.Errorf("%q has invalid day range format %q", k, value)}
		}
		if !validDays[strings.ToLower(parts[0])] {
			return nil, []error{fmt.Errorf("%q has invalid day %q in range", k, parts[0])}
		}
		if !validDays[strings.ToLower(parts[1])] {
			return nil, []error{fmt.Errorf("%q has invalid day %q in range", k, parts[1])}
		}
		return nil, nil
	}

	// Handle comma-separated format (e.g., "mon,wed,fri")
	parts := strings.Split(value, ",")
	for _, part := range parts {
		day := strings.ToLower(strings.TrimSpace(part))
		if !validDays[day] {
			return nil, []error{fmt.Errorf("%q has invalid day %q, must be one of: sun, mon, tue, wed, thu, fri, sat", k, day)}
		}
	}

	return nil, nil
}
