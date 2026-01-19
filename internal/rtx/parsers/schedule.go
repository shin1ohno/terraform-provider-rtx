package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Schedule represents a schedule configuration on an RTX router
type Schedule struct {
	ID         int      `json:"id"`                    // Schedule ID (1-65535)
	Name       string   `json:"name,omitempty"`        // Schedule name/description
	AtTime     string   `json:"at_time,omitempty"`     // Time in HH:MM format
	DayOfWeek  string   `json:"day_of_week,omitempty"` // Day(s) of week (e.g., "mon-fri", "sat", "sun,mon")
	Date       string   `json:"date,omitempty"`        // Specific date in YYYY/MM/DD format
	Recurring  bool     `json:"recurring"`             // Whether schedule repeats
	OnStartup  bool     `json:"on_startup"`            // Execute at router startup
	PolicyList string   `json:"policy_list,omitempty"` // Policy/command list name
	Commands   []string `json:"commands,omitempty"`    // Commands to execute
	Enabled    bool     `json:"enabled"`               // Whether schedule is enabled
}

// KronPolicy represents a kron policy (command list) on an RTX router
type KronPolicy struct {
	Name     string   `json:"name"`               // Policy name
	Commands []string `json:"commands,omitempty"` // Commands in the policy
}

// ScheduleParser parses schedule configuration output
type ScheduleParser struct{}

// NewScheduleParser creates a new Schedule parser
func NewScheduleParser() *ScheduleParser {
	return &ScheduleParser{}
}

// ParseScheduleConfig parses the output of "show config" command for schedule configuration
// and returns a list of Schedules
func (p *ScheduleParser) ParseScheduleConfig(raw string) ([]Schedule, error) {
	schedules := make(map[int]*Schedule)
	lines := strings.Split(raw, "\n")

	// Patterns for different schedule configuration lines
	// schedule at <id> <time> <command>
	scheduleAtTimePattern := regexp.MustCompile(`^\s*schedule\s+at\s+(\d+)\s+(\d{1,2}:\d{2})\s+(.+)\s*$`)
	// schedule at <id> startup <command>
	scheduleAtStartupPattern := regexp.MustCompile(`^\s*schedule\s+at\s+(\d+)\s+startup\s+(.+)\s*$`)
	// schedule at <id> <date> <time> <command>
	scheduleAtDateTimePattern := regexp.MustCompile(`^\s*schedule\s+at\s+(\d+)\s+(\d{4}/\d{2}/\d{2})\s+(\d{1,2}:\d{2})\s+(.+)\s*$`)
	// schedule pp <n> <day> <time> connect/disconnect
	schedulePPPattern := regexp.MustCompile(`^\s*schedule\s+pp\s+(\d+)\s+([a-z,-]+)\s+(\d{1,2}:\d{2})\s+(connect|disconnect)\s*$`)
	// no schedule at <id>
	noSchedulePattern := regexp.MustCompile(`^\s*no\s+schedule\s+at\s+(\d+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Try schedule at date/time pattern first (more specific)
		if matches := scheduleAtDateTimePattern.FindStringSubmatch(line); len(matches) >= 5 {
			id, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}
			date := matches[2]
			time := matches[3]
			command := strings.TrimSpace(matches[4])

			schedule, exists := schedules[id]
			if !exists {
				schedule = &Schedule{
					ID:        id,
					Recurring: false,
					Enabled:   true,
				}
				schedules[id] = schedule
			}
			schedule.Date = date
			schedule.AtTime = time
			schedule.Commands = append(schedule.Commands, command)
			continue
		}

		// Try schedule at startup pattern
		if matches := scheduleAtStartupPattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}
			command := strings.TrimSpace(matches[2])

			schedule, exists := schedules[id]
			if !exists {
				schedule = &Schedule{
					ID:        id,
					Recurring: false,
					Enabled:   true,
				}
				schedules[id] = schedule
			}
			schedule.OnStartup = true
			schedule.Commands = append(schedule.Commands, command)
			continue
		}

		// Try schedule at time pattern (daily recurring)
		if matches := scheduleAtTimePattern.FindStringSubmatch(line); len(matches) >= 4 {
			id, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}
			time := matches[2]
			command := strings.TrimSpace(matches[3])

			schedule, exists := schedules[id]
			if !exists {
				schedule = &Schedule{
					ID:        id,
					Recurring: true,
					Enabled:   true,
				}
				schedules[id] = schedule
			}
			schedule.AtTime = time
			schedule.Commands = append(schedule.Commands, command)
			continue
		}

		// Try schedule pp pattern (PP interface schedule)
		if matches := schedulePPPattern.FindStringSubmatch(line); len(matches) >= 5 {
			ppNum, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}
			// Use negative IDs for PP schedules to differentiate
			id := -ppNum
			day := matches[2]
			time := matches[3]
			action := matches[4]

			schedule, exists := schedules[id]
			if !exists {
				schedule = &Schedule{
					ID:        id,
					Recurring: true,
					Enabled:   true,
				}
				schedules[id] = schedule
			}
			schedule.DayOfWeek = day
			schedule.AtTime = time
			schedule.Commands = append(schedule.Commands, action)
			continue
		}

		// Try no schedule pattern (disabled)
		if matches := noSchedulePattern.FindStringSubmatch(line); len(matches) >= 2 {
			id, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}
			if schedule, exists := schedules[id]; exists {
				schedule.Enabled = false
			}
			continue
		}
	}

	// Convert map to slice
	result := make([]Schedule, 0, len(schedules))
	for _, schedule := range schedules {
		result = append(result, *schedule)
	}

	return result, nil
}

// ParseSingleSchedule parses configuration for a specific schedule
func (p *ScheduleParser) ParseSingleSchedule(raw string, id int) (*Schedule, error) {
	schedules, err := p.ParseScheduleConfig(raw)
	if err != nil {
		return nil, err
	}

	for _, schedule := range schedules {
		if schedule.ID == id {
			return &schedule, nil
		}
	}

	return nil, fmt.Errorf("schedule %d not found", id)
}

// ParseKronPolicyConfig parses kron policy configurations
// Note: RTX routers don't have native kron policy support like Cisco,
// but we can simulate it using multiple schedule commands
func (p *ScheduleParser) ParseKronPolicyConfig(raw string) ([]KronPolicy, error) {
	policies := make(map[string]*KronPolicy)
	lines := strings.Split(raw, "\n")

	// Look for policy-like configurations
	// This could be implemented as comments or specific naming conventions
	// For now, we'll look for consecutive schedule commands with the same prefix

	// Pattern: # kron-policy-list <name>
	policyHeaderPattern := regexp.MustCompile(`^\s*#\s*kron-policy-list\s+(\S+)\s*$`)
	currentPolicy := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for policy header
		if matches := policyHeaderPattern.FindStringSubmatch(line); len(matches) >= 2 {
			currentPolicy = matches[1]
			if _, exists := policies[currentPolicy]; !exists {
				policies[currentPolicy] = &KronPolicy{
					Name:     currentPolicy,
					Commands: []string{},
				}
			}
			continue
		}

		// If we're in a policy context and see a command, add it
		if currentPolicy != "" && !strings.HasPrefix(line, "#") {
			if policy, exists := policies[currentPolicy]; exists {
				policy.Commands = append(policy.Commands, line)
			}
		}
	}

	// Convert map to slice
	result := make([]KronPolicy, 0, len(policies))
	for _, policy := range policies {
		result = append(result, *policy)
	}

	return result, nil
}

// BuildScheduleAtCommand builds a command to create a time-based schedule
// Command format: schedule at <id> <time> <command>
func BuildScheduleAtCommand(id int, time, command string) string {
	return fmt.Sprintf("schedule at %d %s %s", id, time, command)
}

// BuildScheduleAtStartupCommand builds a command to create a startup schedule
// Command format: schedule at <id> startup <command>
func BuildScheduleAtStartupCommand(id int, command string) string {
	return fmt.Sprintf("schedule at %d startup %s", id, command)
}

// BuildScheduleAtDateTimeCommand builds a command to create a date/time specific schedule
// Command format: schedule at <id> <date> <time> <command>
func BuildScheduleAtDateTimeCommand(id int, date, time, command string) string {
	return fmt.Sprintf("schedule at %d %s %s %s", id, date, time, command)
}

// BuildSchedulePPCommand builds a command to create a PP interface schedule
// Command format: schedule pp <n> <day> <time> connect/disconnect
func BuildSchedulePPCommand(ppNum int, dayOfWeek, time, action string) string {
	return fmt.Sprintf("schedule pp %d %s %s %s", ppNum, dayOfWeek, time, action)
}

// BuildDeleteScheduleCommand builds the command to delete a schedule
// Command format: no schedule at <id>
func BuildDeleteScheduleCommand(id int) string {
	return fmt.Sprintf("no schedule at %d", id)
}

// BuildDeleteSchedulePPCommand builds the command to delete a PP schedule
// Command format: no schedule pp <n> <day> <time>
func BuildDeleteSchedulePPCommand(ppNum int, dayOfWeek, time string) string {
	return fmt.Sprintf("no schedule pp %d %s %s", ppNum, dayOfWeek, time)
}

// BuildShowScheduleCommand builds the command to show schedule configuration
func BuildShowScheduleCommand() string {
	return "show config | grep schedule"
}

// BuildShowScheduleByIDCommand builds the command to show a specific schedule
func BuildShowScheduleByIDCommand(id int) string {
	return fmt.Sprintf("show config | grep \"schedule at %d\"", id)
}

// ValidateSchedule validates a Schedule configuration
func ValidateSchedule(schedule Schedule) error {
	// Validate ID
	if schedule.ID < 1 || schedule.ID > 65535 {
		return fmt.Errorf("schedule id must be between 1 and 65535, got %d", schedule.ID)
	}

	// Validate time format if specified
	if schedule.AtTime != "" {
		if err := ValidateTimeFormat(schedule.AtTime); err != nil {
			return err
		}
	}

	// Validate date format if specified
	if schedule.Date != "" {
		if err := ValidateDateFormat(schedule.Date); err != nil {
			return err
		}
	}

	// Validate day of week if specified
	if schedule.DayOfWeek != "" {
		if err := ValidateDayOfWeek(schedule.DayOfWeek); err != nil {
			return err
		}
	}

	// Must have either time, startup, or date specified
	if schedule.AtTime == "" && !schedule.OnStartup && schedule.Date == "" {
		return fmt.Errorf("schedule must have at_time, on_startup, or date specified")
	}

	// Cannot have both startup and time/date
	if schedule.OnStartup && (schedule.AtTime != "" || schedule.Date != "") {
		return fmt.Errorf("on_startup cannot be combined with at_time or date")
	}

	// Must have at least one command
	if len(schedule.Commands) == 0 && schedule.PolicyList == "" {
		return fmt.Errorf("schedule must have at least one command or policy_list")
	}

	return nil
}

// ValidateTimeFormat validates a time string in HH:MM format
func ValidateTimeFormat(timeStr string) error {
	timePattern := regexp.MustCompile(`^(\d{1,2}):(\d{2})$`)
	matches := timePattern.FindStringSubmatch(timeStr)
	if len(matches) != 3 {
		return fmt.Errorf("invalid time format %q, expected HH:MM", timeStr)
	}

	hour, _ := strconv.Atoi(matches[1])
	minute, _ := strconv.Atoi(matches[2])

	if hour < 0 || hour > 23 {
		return fmt.Errorf("invalid hour %d in time %q, must be 0-23", hour, timeStr)
	}
	if minute < 0 || minute > 59 {
		return fmt.Errorf("invalid minute %d in time %q, must be 0-59", minute, timeStr)
	}

	return nil
}

// ValidateDateFormat validates a date string in YYYY/MM/DD format
func ValidateDateFormat(dateStr string) error {
	datePattern := regexp.MustCompile(`^(\d{4})/(\d{2})/(\d{2})$`)
	matches := datePattern.FindStringSubmatch(dateStr)
	if len(matches) != 4 {
		return fmt.Errorf("invalid date format %q, expected YYYY/MM/DD", dateStr)
	}

	year, _ := strconv.Atoi(matches[1])
	month, _ := strconv.Atoi(matches[2])
	day, _ := strconv.Atoi(matches[3])

	if year < 2000 || year > 2099 {
		return fmt.Errorf("invalid year %d in date %q, must be 2000-2099", year, dateStr)
	}
	if month < 1 || month > 12 {
		return fmt.Errorf("invalid month %d in date %q, must be 1-12", month, dateStr)
	}
	if day < 1 || day > 31 {
		return fmt.Errorf("invalid day %d in date %q, must be 1-31", day, dateStr)
	}

	return nil
}

// ValidateDayOfWeek validates a day of week specification
// Valid formats: "mon", "tue", "mon-fri", "sat,sun", "mon,wed,fri"
func ValidateDayOfWeek(dayStr string) error {
	validDays := map[string]bool{
		"sun": true, "mon": true, "tue": true, "wed": true,
		"thu": true, "fri": true, "sat": true,
	}

	// Handle range format (e.g., "mon-fri")
	if strings.Contains(dayStr, "-") {
		parts := strings.Split(dayStr, "-")
		if len(parts) != 2 {
			return fmt.Errorf("invalid day range format %q", dayStr)
		}
		if !validDays[strings.ToLower(parts[0])] {
			return fmt.Errorf("invalid day %q in range", parts[0])
		}
		if !validDays[strings.ToLower(parts[1])] {
			return fmt.Errorf("invalid day %q in range", parts[1])
		}
		return nil
	}

	// Handle comma-separated format (e.g., "mon,wed,fri")
	parts := strings.Split(dayStr, ",")
	for _, part := range parts {
		day := strings.ToLower(strings.TrimSpace(part))
		if !validDays[day] {
			return fmt.Errorf("invalid day %q, must be one of: sun, mon, tue, wed, thu, fri, sat", day)
		}
	}

	return nil
}

// ValidateKronPolicy validates a KronPolicy configuration
func ValidateKronPolicy(policy KronPolicy) error {
	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}

	// Validate policy name format (alphanumeric and underscores only)
	namePattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	if !namePattern.MatchString(policy.Name) {
		return fmt.Errorf("policy name %q must start with a letter and contain only letters, numbers, underscores, and hyphens", policy.Name)
	}

	if len(policy.Commands) == 0 {
		return fmt.Errorf("policy must have at least one command")
	}

	return nil
}
