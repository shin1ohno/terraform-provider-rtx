package client

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParserRegistry manages command parsers
type ParserRegistry struct {
	parsers map[string]Parser
}

// NewParserRegistry creates a new parser registry with default parsers
func NewParserRegistry() *ParserRegistry {
	r := &ParserRegistry{
		parsers: make(map[string]Parser),
	}

	// Register default parsers
	r.Register("show environment", &environmentParser{})
	r.Register("show status boot", &bootStatusParser{})
	r.Register("show config", &rawParser{})

	return r
}

// Register adds a parser for a specific command
func (r *ParserRegistry) Register(cmdKey string, parser Parser) {
	r.parsers[cmdKey] = parser
}

// Get retrieves a parser for a command, returns rawParser if not found
func (r *ParserRegistry) Get(cmdKey string) Parser {
	if p, ok := r.parsers[cmdKey]; ok {
		return p
	}
	return &rawParser{}
}

// rawParser returns the raw output without parsing
type rawParser struct{}

func (p *rawParser) Parse(raw []byte) (interface{}, error) {
	return string(raw), nil
}

// EnvironmentInfo represents the output of "show environment"
type EnvironmentInfo struct {
	Temperature float64
	CPUUsage    int
	MemoryUsage int
}

// environmentParser parses "show environment" command output
type environmentParser struct{}

func (p *environmentParser) Parse(raw []byte) (interface{}, error) {
	output := string(raw)
	info := &EnvironmentInfo{}

	// Parse temperature (example: "Temperature: 45.5C")
	if match := regexp.MustCompile(`Temperature:\s*([\d.]+)`).FindStringSubmatch(output); len(match) > 1 {
		temp, err := strconv.ParseFloat(match[1], 64)
		if err == nil {
			info.Temperature = temp
		}
	}

	// Parse CPU usage (example: "CPU: 25%")
	if match := regexp.MustCompile(`CPU:\s*(\d+)%`).FindStringSubmatch(output); len(match) > 1 {
		cpu, err := strconv.Atoi(match[1])
		if err == nil {
			info.CPUUsage = cpu
		}
	}

	// Parse memory usage (example: "Memory: 60%")
	if match := regexp.MustCompile(`Memory:\s*(\d+)%`).FindStringSubmatch(output); len(match) > 1 {
		mem, err := strconv.Atoi(match[1])
		if err == nil {
			info.MemoryUsage = mem
		}
	}

	return info, nil
}

// BootStatus represents the output of "show status boot"
type BootStatus struct {
	Version    string
	BootTime   time.Time
	Uptime     time.Duration
	LastReboot string
}

// bootStatusParser parses "show status boot" command output
type bootStatusParser struct{}

func (p *bootStatusParser) Parse(raw []byte) (interface{}, error) {
	output := string(raw)
	status := &BootStatus{}

	// Parse version (example: "RTX1210 Rev.14.01.38")
	if match := regexp.MustCompile(`RTX\d+\s+Rev\.([\d.]+)`).FindStringSubmatch(output); len(match) > 1 {
		status.Version = match[1]
	}

	// Parse uptime (example: "Uptime: 10 days 5:30:45")
	if match := regexp.MustCompile(`Uptime:\s*(\d+)\s*days?\s*(\d+):(\d+):(\d+)`).FindStringSubmatch(output); len(match) > 4 {
		days, _ := strconv.Atoi(match[1])
		hours, _ := strconv.Atoi(match[2])
		minutes, _ := strconv.Atoi(match[3])
		seconds, _ := strconv.Atoi(match[4])

		status.Uptime = time.Duration(days)*24*time.Hour +
			time.Duration(hours)*time.Hour +
			time.Duration(minutes)*time.Minute +
			time.Duration(seconds)*time.Second
	}

	// Parse last reboot reason
	if idx := strings.Index(output, "Reboot by"); idx != -1 {
		endIdx := strings.IndexByte(output[idx:], '\n')
		if endIdx == -1 {
			status.LastReboot = strings.TrimSpace(output[idx:])
		} else {
			status.LastReboot = strings.TrimSpace(output[idx : idx+endIdx])
		}
	}

	return status, nil
}

// ConfigSection represents a section of router configuration
type ConfigSection struct {
	Name  string
	Lines []string
}

// ParseError wraps parsing errors with context
type ParseError struct {
	Command string
	Err     error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("failed to parse %s output: %v", e.Command, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}
