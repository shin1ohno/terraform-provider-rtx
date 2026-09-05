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
	// Context is the execution-context token the line carried: "*" (the router
	// itself), "pp <n>", "tunnel <n>" or "switch <sw>". rtx_kron_schedule only
	// models "*"; the field exists so the resource can warn rather than
	// silently rewrite a pp/tunnel/switch schedule as "*" on the next apply.
	Context string `json:"context,omitempty"`
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
	// ParseScheduleConfig used to return the map's range order, so ListSchedules
	// handed back a different order on every call. Track first-seen order.
	order := make([]int, 0, 8)
	lines := strings.Split(raw, "\n")

	// The `schedule at` forms are tokenized rather than matched by one regexp —
	// see parseScheduleAtLine. The two forms below have a fixed shape and stay
	// as patterns.
	// schedule pp <n> <day> <time> connect/disconnect
	schedulePPPattern := regexp.MustCompile(`^\s*schedule\s+pp\s+(\d+)\s+([a-z,-]+)\s+(\d{1,2}:\d{2})\s+(connect|disconnect)\s*$`)
	// no schedule at <id>
	noSchedulePattern := regexp.MustCompile(`^\s*no\s+schedule\s+at\s+(\d+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// schedule at <id> [<date>] <time>|startup <context> <command...>
		if parsed, ok := parseScheduleAtLine(line); ok {
			if existing, dup := schedules[parsed.ID]; dup {
				// The RTX only allows one `schedule at` line per ID, so a second
				// line for the same ID means the caller concatenated configs.
				// Keep the first line's timing and collect the extra command.
				existing.Commands = append(existing.Commands, parsed.Commands...)
				continue
			}
			s := parsed
			schedules[parsed.ID] = &s
			order = append(order, parsed.ID)
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
				order = append(order, id)
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

	result := make([]Schedule, 0, len(order))
	for _, id := range order {
		result = append(result, *schedules[id])
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

// scheduleAnyContext is the RTX execution-context token that means "run this in
// the router's own context" — as opposed to `pp <n>`, `tunnel <n>` or
// `switch <sw>`. RTX1210 Rev.14.01.42 `schedule at ?` prints the grammar as
//
//	schedule at ID [日付] 時刻 * コマンド...
//	schedule at ID [日付] 時刻 pp 相手先番号 コマンド...
//	schedule at ID [日付] 時刻 tunnel トンネル番号 コマンド...
//	schedule at ID [日付] 時刻 switch スイッチ コマンド...
//
// so the token is mandatory on the wire; a `schedule at` line written without
// it is rejected by the router.
const scheduleAnyContext = "*"

// scheduleEveryDay is how the RTX renders "every day" in the optional date
// slot. It carries no information the model does not already hold in
// Recurring, and it is not a value the rtx_kron_schedule `date` attribute can
// hold, so the parser folds it into Recurring instead of surfacing it.
const scheduleEveryDay = "*/*"

var (
	// The ID and everything the RTX printed after it.
	scheduleAtPrefixPattern = regexp.MustCompile(`^\s*schedule\s+at\s+(\d+)\s+(.*)$`)

	// A one-time date. `show config` renders these as YYYY/MM/DD.
	scheduleFullDatePattern = regexp.MustCompile(`^\d{4}/\d{1,2}/\d{1,2}$`)

	// A repeating date: month/day with wildcards, or a weekday selector —
	// `*/*`, `*/mon-fri`, `*/sat,sun`, `1/1`.
	scheduleRepeatDatePattern = regexp.MustCompile(`^(\d{1,2}|\*)/(\*|\d{1,2}|(sun|mon|tue|wed|thu|fri|sat)([,-](sun|mon|tue|wed|thu|fri|sat))*)$`)

	// A clock token. The router renders HH:MM:SS; operators write H:MM. Any
	// field may be a wildcard (`12:*:00`, `*:00`).
	scheduleTimeTokenPattern = regexp.MustCompile(`^(\d{1,2}|\*):(\d{1,2}|\*)(:(\d{1,2}|\*))?$`)

	scheduleNumberPattern = regexp.MustCompile(`^\d+$`)

	// Zero-padding forms, used to compare two spellings of the same instant.
	scheduleClockPattern    = regexp.MustCompile(`^(\d{1,2}):(\d{1,2})(?::(\d{1,2}))?$`)
	scheduleCalendarPattern = regexp.MustCompile(`^(\d{4})/(\d{1,2})/(\d{1,2})$`)
)

// scheduleTokens splits s on whitespace and returns each token together with
// the byte offset just past it. The offsets let the caller slice the ORIGINAL
// string for the trailing command rather than re-joining tokens, which would
// collapse whatever internal spacing the operator wrote (a quoted
// `syslog notice 'Timer  expired'` keeps its spaces).
func scheduleTokens(s string) (toks []string, ends []int) {
	i := 0
	for i < len(s) {
		for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
			i++
		}
		if i >= len(s) {
			break
		}
		start := i
		for i < len(s) && s[i] != ' ' && s[i] != '\t' {
			i++
		}
		toks = append(toks, s[start:i])
		ends = append(ends, i)
	}
	return toks, ends
}

// parseScheduleAtLine parses one `schedule at` configuration line into a
// Schedule, returning false for anything it does not model (a `+TIMER` line,
// or a line that is not a schedule at all).
//
// Two things make this a tokenizer instead of a regexp. The date slot is
// optional AND the router re-renders what was written into it — a line written
// as `schedule at 1 0:00 * ntpdate ntp.nict.jp syslog` can come back as
// `schedule at 1 */* 00:00:00 * ntpdate ntp.nict.jp syslog`, which the old
// fixed patterns could not match, so Read reported the schedule missing right
// after Create wrote it. And the execution-context token has to be CONSUMED:
// left in place it lands at the head of Commands[0] and the command_lines
// attribute drifts by a leading "*".
//
// The context token is treated as optional on read even though the device
// requires it, so configs captured by hand or from older firmware still parse.
func parseScheduleAtLine(line string) (Schedule, bool) {
	m := scheduleAtPrefixPattern.FindStringSubmatch(line)
	if m == nil {
		return Schedule{}, false
	}
	id, err := strconv.Atoi(m[1])
	if err != nil {
		return Schedule{}, false
	}

	rest := m[2]
	toks, ends := scheduleTokens(rest)
	if len(toks) == 0 {
		return Schedule{}, false
	}

	schedule := Schedule{ID: id, Enabled: true, Context: scheduleAnyContext}
	oneTime := false
	i := 0

	// [<date>]
	switch {
	case scheduleFullDatePattern.MatchString(toks[i]):
		schedule.Date = toks[i]
		oneTime = true
		i++
	case scheduleRepeatDatePattern.MatchString(toks[i]):
		// The RTX has no day-of-week field of its own: weekdays live in the
		// date slot as `*/<days>`. Surface those in DayOfWeek so a read-back
		// lands on the attribute the practitioner actually wrote — `date` only
		// accepts YYYY/MM/DD, so a `*/mon-fri` left in Date could never round
		// trip. `*/*` says nothing Recurring does not, so it is dropped.
		switch dow := strings.TrimPrefix(toks[i], "*/"); {
		case toks[i] == scheduleEveryDay:
		case strings.HasPrefix(toks[i], "*/") && ValidateDayOfWeek(dow) == nil:
			schedule.DayOfWeek = dow
		default:
			schedule.Date = toks[i]
		}
		i++
	}
	if i >= len(toks) {
		return Schedule{}, false
	}

	// <time> | startup
	switch {
	case toks[i] == "startup":
		schedule.OnStartup = true
		i++
	case scheduleTimeTokenPattern.MatchString(toks[i]):
		schedule.AtTime = toks[i]
		i++
	default:
		return Schedule{}, false
	}
	schedule.Recurring = !schedule.OnStartup && !oneTime

	// <context>. `switch` takes a MAC address or a topology path, both of which
	// contain a colon; requiring that keeps a context-less line whose command
	// happens to start with the word `switch` from being mis-split.
	if i < len(toks) {
		switch {
		case toks[i] == scheduleAnyContext:
			i++
		case (toks[i] == "pp" || toks[i] == "tunnel") && i+1 < len(toks) && scheduleNumberPattern.MatchString(toks[i+1]):
			schedule.Context = toks[i] + " " + toks[i+1]
			i += 2
		case toks[i] == "switch" && i+1 < len(toks) && strings.Contains(toks[i+1], ":"):
			schedule.Context = toks[i] + " " + toks[i+1]
			i += 2
		}
	}
	if i == 0 || i > len(toks) {
		return Schedule{}, false
	}

	command := strings.TrimSpace(rest[ends[i-1]:])
	if command == "" {
		return Schedule{}, false
	}
	schedule.Commands = []string{command}

	return schedule, true
}

// NormalizeScheduleTime renders a schedule clock token in the router's own
// zero-padded HH:MM:SS form so that two spellings of the same instant — the
// `0:00` an operator writes and the `00:00:00` `show config` prints back —
// compare equal. A token with a wildcard field is returned unchanged, so it
// only ever compares equal to itself.
func NormalizeScheduleTime(t string) string {
	m := scheduleClockPattern.FindStringSubmatch(t)
	if m == nil {
		return t
	}
	hour, _ := strconv.Atoi(m[1])
	minute, _ := strconv.Atoi(m[2])
	second := 0
	if m[3] != "" {
		second, _ = strconv.Atoi(m[3])
	}
	return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
}

// NormalizeScheduleDate renders a one-time date in zero-padded YYYY/MM/DD form
// for the same reason NormalizeScheduleTime exists. Wildcard and weekday dates
// are returned unchanged.
func NormalizeScheduleDate(d string) string {
	m := scheduleCalendarPattern.FindStringSubmatch(d)
	if m == nil {
		return d
	}
	year, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	day, _ := strconv.Atoi(m[3])
	return fmt.Sprintf("%04d/%02d/%02d", year, month, day)
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
// Command format: schedule at <id> <time> * <command>
//
// The `*` is the mandatory execution-context token — see scheduleAnyContext.
// Without it the RTX rejects the line, and because the write path only treats
// output matching a known error string as a failure, the rejection surfaces
// later as a Read that cannot find the schedule.
func BuildScheduleAtCommand(id int, time, command string) string {
	return fmt.Sprintf("schedule at %d %s %s %s", id, time, scheduleAnyContext, command)
}

// BuildScheduleAtStartupCommand builds a command to create a startup schedule
// Command format: schedule at <id> startup * <command>
func BuildScheduleAtStartupCommand(id int, command string) string {
	return fmt.Sprintf("schedule at %d startup %s %s", id, scheduleAnyContext, command)
}

// BuildScheduleAtDateTimeCommand builds a command to create a date/time specific schedule
// Command format: schedule at <id> <date> <time> * <command>
func BuildScheduleAtDateTimeCommand(id int, date, time, command string) string {
	return fmt.Sprintf("schedule at %d %s %s %s %s", id, date, time, scheduleAnyContext, command)
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

// NormalizeScheduleDayOfWeek renders a weekday selector in one spelling so
// `"mon, wed, fri"` from config and `"mon,wed,fri"` from the router compare
// equal. Ranges pass through unchanged.
func NormalizeScheduleDayOfWeek(d string) string {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(d)), ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return strings.Join(parts, ",")
}

// IsScheduleOneTimeDate reports whether a date token names a single calendar
// day. `1/1`, `*/mon-fri` and `*/*` all repeat.
func IsScheduleOneTimeDate(date string) bool {
	return scheduleFullDatePattern.MatchString(date)
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

// ValidateTimeFormat validates a time string in HH:MM or HH:MM:SS format.
//
// The seconds field is accepted because that is how `show config` renders a
// schedule back (`0:00` is stored, `00:00:00` is printed), so a value read off
// the router has to survive being written into config unchanged.
func ValidateTimeFormat(timeStr string) error {
	timePattern := regexp.MustCompile(`^(\d{1,2}):(\d{2})(?::(\d{2}))?$`)
	matches := timePattern.FindStringSubmatch(timeStr)
	if len(matches) != 4 {
		return fmt.Errorf("invalid time format %q, expected HH:MM or HH:MM:SS", timeStr)
	}

	hour, _ := strconv.Atoi(matches[1])
	minute, _ := strconv.Atoi(matches[2])
	if matches[3] != "" {
		if second, _ := strconv.Atoi(matches[3]); second < 0 || second > 59 {
			return fmt.Errorf("invalid second %d in time %q, must be 0-59", second, timeStr)
		}
	}

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
