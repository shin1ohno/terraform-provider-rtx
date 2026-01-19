package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SystemConfig represents system-level configuration on an RTX router
type SystemConfig struct {
	Timezone      string               `json:"timezone,omitempty"`       // UTC offset (e.g., "+09:00")
	Console       *ConsoleConfig       `json:"console,omitempty"`        // Console settings
	PacketBuffers []PacketBufferConfig `json:"packet_buffers,omitempty"` // Packet buffer tuning
	Statistics    *StatisticsConfig    `json:"statistics,omitempty"`     // Statistics collection
}

// ConsoleConfig represents console settings
type ConsoleConfig struct {
	Character string `json:"character,omitempty"` // Character encoding (ja.utf8, ascii, ja.sjis)
	Lines     string `json:"lines,omitempty"`     // Lines per page (number or "infinity")
	Prompt    string `json:"prompt,omitempty"`    // Custom prompt string
}

// PacketBufferConfig represents packet buffer tuning for each size
type PacketBufferConfig struct {
	Size      string `json:"size"`       // "small", "middle", or "large"
	MaxBuffer int    `json:"max_buffer"` // Maximum buffer count
	MaxFree   int    `json:"max_free"`   // Maximum free buffer count
}

// StatisticsConfig represents statistics collection settings
type StatisticsConfig struct {
	Traffic bool `json:"traffic"` // Traffic statistics
	NAT     bool `json:"nat"`     // NAT statistics
}

// SystemParser parses system configuration output
type SystemParser struct{}

// NewSystemParser creates a new system parser
func NewSystemParser() *SystemParser {
	return &SystemParser{}
}

// ParseSystemConfig parses the output of system configuration commands
func (p *SystemParser) ParseSystemConfig(raw string) (*SystemConfig, error) {
	config := &SystemConfig{
		PacketBuffers: []PacketBufferConfig{},
	}
	lines := strings.Split(raw, "\n")

	// Patterns for different configuration lines
	// timezone +09:00
	timezonePattern := regexp.MustCompile(`^\s*timezone\s+([\+\-]?\d{2}:\d{2})\s*$`)
	// console character ja.utf8
	consoleCharPattern := regexp.MustCompile(`^\s*console\s+character\s+(\S+)\s*$`)
	// console lines infinity | console lines 24
	consoleLinesPattern := regexp.MustCompile(`^\s*console\s+lines\s+(\S+)\s*$`)
	// console prompt "string" or console prompt string
	consolePromptPattern := regexp.MustCompile(`^\s*console\s+prompt\s+"?([^"]*)"?\s*$`)
	// system packet-buffer small max-buffer=5000 max-free=1300
	packetBufferPattern := regexp.MustCompile(`^\s*system\s+packet-buffer\s+(small|middle|large)\s+max-buffer=(\d+)\s+max-free=(\d+)\s*$`)
	// statistics traffic on|off
	statsTrafficPattern := regexp.MustCompile(`^\s*statistics\s+traffic\s+(on|off)\s*$`)
	// statistics nat on|off
	statsNATPattern := regexp.MustCompile(`^\s*statistics\s+nat\s+(on|off)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try timezone pattern
		if matches := timezonePattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.Timezone = matches[1]
			continue
		}

		// Try console character pattern
		if matches := consoleCharPattern.FindStringSubmatch(line); len(matches) >= 2 {
			if config.Console == nil {
				config.Console = &ConsoleConfig{}
			}
			config.Console.Character = matches[1]
			continue
		}

		// Try console lines pattern
		if matches := consoleLinesPattern.FindStringSubmatch(line); len(matches) >= 2 {
			if config.Console == nil {
				config.Console = &ConsoleConfig{}
			}
			config.Console.Lines = matches[1]
			continue
		}

		// Try console prompt pattern
		if matches := consolePromptPattern.FindStringSubmatch(line); len(matches) >= 2 {
			if config.Console == nil {
				config.Console = &ConsoleConfig{}
			}
			config.Console.Prompt = matches[1]
			continue
		}

		// Try packet buffer pattern
		if matches := packetBufferPattern.FindStringSubmatch(line); len(matches) >= 4 {
			maxBuffer, _ := strconv.Atoi(matches[2])
			maxFree, _ := strconv.Atoi(matches[3])
			config.PacketBuffers = append(config.PacketBuffers, PacketBufferConfig{
				Size:      matches[1],
				MaxBuffer: maxBuffer,
				MaxFree:   maxFree,
			})
			continue
		}

		// Try statistics traffic pattern
		if matches := statsTrafficPattern.FindStringSubmatch(line); len(matches) >= 2 {
			if config.Statistics == nil {
				config.Statistics = &StatisticsConfig{}
			}
			config.Statistics.Traffic = matches[1] == "on"
			continue
		}

		// Try statistics NAT pattern
		if matches := statsNATPattern.FindStringSubmatch(line); len(matches) >= 2 {
			if config.Statistics == nil {
				config.Statistics = &StatisticsConfig{}
			}
			config.Statistics.NAT = matches[1] == "on"
			continue
		}
	}

	return config, nil
}

// BuildTimezoneCommand builds the command to set timezone
// Command format: timezone <utc_offset>
func BuildTimezoneCommand(tz string) string {
	return fmt.Sprintf("timezone %s", tz)
}

// BuildDeleteTimezoneCommand builds the command to remove timezone setting
func BuildDeleteTimezoneCommand() string {
	return "no timezone"
}

// BuildConsoleCharacterCommand builds the command to set console character encoding
// Command format: console character <encoding>
func BuildConsoleCharacterCommand(encoding string) string {
	return fmt.Sprintf("console character %s", encoding)
}

// BuildDeleteConsoleCharacterCommand builds the command to remove console character setting
func BuildDeleteConsoleCharacterCommand() string {
	return "no console character"
}

// BuildConsoleLinesCommand builds the command to set console lines
// Command format: console lines <number|infinity>
func BuildConsoleLinesCommand(lines string) string {
	return fmt.Sprintf("console lines %s", lines)
}

// BuildDeleteConsoleLinesCommand builds the command to remove console lines setting
func BuildDeleteConsoleLinesCommand() string {
	return "no console lines"
}

// BuildConsolePromptCommand builds the command to set console prompt
// Command format: console prompt "<prompt>"
func BuildConsolePromptCommand(prompt string) string {
	// Quote the prompt if it contains spaces
	if strings.Contains(prompt, " ") {
		return fmt.Sprintf("console prompt \"%s\"", prompt)
	}
	return fmt.Sprintf("console prompt %s", prompt)
}

// BuildDeleteConsolePromptCommand builds the command to remove console prompt setting
func BuildDeleteConsolePromptCommand() string {
	return "no console prompt"
}

// BuildPacketBufferCommand builds the command to set packet buffer settings
// Command format: system packet-buffer <size> max-buffer=<n> max-free=<n>
func BuildPacketBufferCommand(config PacketBufferConfig) string {
	return fmt.Sprintf("system packet-buffer %s max-buffer=%d max-free=%d",
		config.Size, config.MaxBuffer, config.MaxFree)
}

// BuildDeletePacketBufferCommand builds the command to remove packet buffer setting
func BuildDeletePacketBufferCommand(size string) string {
	return fmt.Sprintf("no system packet-buffer %s", size)
}

// BuildStatisticsTrafficCommand builds the command to set traffic statistics
// Command format: statistics traffic on|off
func BuildStatisticsTrafficCommand(enabled bool) string {
	if enabled {
		return "statistics traffic on"
	}
	return "statistics traffic off"
}

// BuildDeleteStatisticsTrafficCommand builds the command to remove traffic statistics setting
func BuildDeleteStatisticsTrafficCommand() string {
	return "no statistics traffic"
}

// BuildStatisticsNATCommand builds the command to set NAT statistics
// Command format: statistics nat on|off
func BuildStatisticsNATCommand(enabled bool) string {
	if enabled {
		return "statistics nat on"
	}
	return "statistics nat off"
}

// BuildDeleteStatisticsNATCommand builds the command to remove NAT statistics setting
func BuildDeleteStatisticsNATCommand() string {
	return "no statistics nat"
}

// BuildShowSystemConfigCommand builds the command to show system configuration
// Command format: show config | grep -E "(timezone|console|packet-buffer|statistics)"
func BuildShowSystemConfigCommand() string {
	return `show config | grep -E "(timezone|console|packet-buffer|statistics)"`
}

// BuildDeleteSystemCommands builds all commands needed to reset system configuration
func BuildDeleteSystemCommands(config *SystemConfig) []string {
	var commands []string

	if config.Timezone != "" {
		commands = append(commands, BuildDeleteTimezoneCommand())
	}

	if config.Console != nil {
		if config.Console.Character != "" {
			commands = append(commands, BuildDeleteConsoleCharacterCommand())
		}
		if config.Console.Lines != "" {
			commands = append(commands, BuildDeleteConsoleLinesCommand())
		}
		if config.Console.Prompt != "" {
			commands = append(commands, BuildDeleteConsolePromptCommand())
		}
	}

	for _, pb := range config.PacketBuffers {
		commands = append(commands, BuildDeletePacketBufferCommand(pb.Size))
	}

	if config.Statistics != nil {
		commands = append(commands, BuildDeleteStatisticsTrafficCommand())
		commands = append(commands, BuildDeleteStatisticsNATCommand())
	}

	return commands
}

// ValidateSystemConfig validates a system configuration
func ValidateSystemConfig(config *SystemConfig) error {
	// Validate timezone format
	if config.Timezone != "" {
		if !isValidTimezone(config.Timezone) {
			return fmt.Errorf("invalid timezone format: %s (expected ±HH:MM)", config.Timezone)
		}
	}

	// Validate console settings
	if config.Console != nil {
		if config.Console.Character != "" {
			if !isValidCharacterEncoding(config.Console.Character) {
				return fmt.Errorf("invalid character encoding: %s (expected: ja.utf8, ja.sjis, ascii)", config.Console.Character)
			}
		}

		if config.Console.Lines != "" {
			if !isValidConsoleLines(config.Console.Lines) {
				return fmt.Errorf("invalid console lines: %s (expected positive integer or 'infinity')", config.Console.Lines)
			}
		}
	}

	// Validate packet buffers
	for _, pb := range config.PacketBuffers {
		if !isValidPacketBufferSize(pb.Size) {
			return fmt.Errorf("invalid packet buffer size: %s (expected: small, middle, large)", pb.Size)
		}
		if pb.MaxBuffer <= 0 {
			return fmt.Errorf("max_buffer must be positive for size %s", pb.Size)
		}
		if pb.MaxFree <= 0 {
			return fmt.Errorf("max_free must be positive for size %s", pb.Size)
		}
		if pb.MaxFree > pb.MaxBuffer {
			return fmt.Errorf("max_free cannot exceed max_buffer for size %s", pb.Size)
		}
	}

	return nil
}

// isValidTimezone checks if a string is a valid timezone format (±HH:MM)
func isValidTimezone(tz string) bool {
	pattern := regexp.MustCompile(`^[\+\-]\d{2}:\d{2}$`)
	return pattern.MatchString(tz)
}

// isValidCharacterEncoding checks if a string is a valid character encoding
func isValidCharacterEncoding(encoding string) bool {
	validEncodings := []string{"ja.utf8", "ja.sjis", "ascii", "euc-jp"}
	for _, valid := range validEncodings {
		if encoding == valid {
			return true
		}
	}
	return false
}

// isValidConsoleLines checks if a string is a valid console lines setting
func isValidConsoleLines(lines string) bool {
	if lines == "infinity" {
		return true
	}
	n, err := strconv.Atoi(lines)
	return err == nil && n > 0
}

// isValidPacketBufferSize checks if a string is a valid packet buffer size
func isValidPacketBufferSize(size string) bool {
	return size == "small" || size == "middle" || size == "large"
}
