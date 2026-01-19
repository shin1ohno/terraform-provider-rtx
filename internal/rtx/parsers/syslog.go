package parsers

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// SyslogConfig represents syslog configuration on an RTX router
type SyslogConfig struct {
	Hosts        []SyslogHost `json:"hosts,omitempty"`         // Syslog destination hosts
	LocalAddress string       `json:"local_address,omitempty"` // Source IP address for syslog
	Facility     string       `json:"facility,omitempty"`      // Syslog facility (e.g., user, local0-local7)
	Notice       bool         `json:"notice"`                  // Log notice level messages
	Info         bool         `json:"info"`                    // Log info level messages
	Debug        bool         `json:"debug"`                   // Log debug level messages
}

// SyslogHost represents a syslog destination host
type SyslogHost struct {
	Address string `json:"address"`        // IP address or hostname
	Port    int    `json:"port,omitempty"` // UDP port (default 514)
}

// SyslogParser parses syslog configuration output
type SyslogParser struct{}

// NewSyslogParser creates a new syslog parser
func NewSyslogParser() *SyslogParser {
	return &SyslogParser{}
}

// ParseSyslogConfig parses the output of syslog configuration commands
func (p *SyslogParser) ParseSyslogConfig(raw string) (*SyslogConfig, error) {
	config := &SyslogConfig{
		Hosts:  []SyslogHost{},
		Notice: false, // Default off
		Info:   false, // Default off
		Debug:  false, // Default off
	}
	lines := strings.Split(raw, "\n")

	// Patterns for different configuration lines
	// syslog host <address> [<port>]
	hostWithPortPattern := regexp.MustCompile(`^\s*syslog\s+host\s+(\S+)\s+(\d+)\s*$`)
	hostPattern := regexp.MustCompile(`^\s*syslog\s+host\s+(\S+)\s*$`)
	// syslog local address <ip>
	localAddressPattern := regexp.MustCompile(`^\s*syslog\s+local\s+address\s+(\S+)\s*$`)
	// syslog facility <facility>
	facilityPattern := regexp.MustCompile(`^\s*syslog\s+facility\s+(\S+)\s*$`)
	// syslog notice on|off
	noticePattern := regexp.MustCompile(`^\s*syslog\s+notice\s+(on|off)\s*$`)
	// syslog info on|off
	infoPattern := regexp.MustCompile(`^\s*syslog\s+info\s+(on|off)\s*$`)
	// syslog debug on|off
	debugPattern := regexp.MustCompile(`^\s*syslog\s+debug\s+(on|off)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try syslog host with port pattern
		if matches := hostWithPortPattern.FindStringSubmatch(line); len(matches) >= 3 {
			port, err := strconv.Atoi(matches[2])
			if err != nil {
				continue
			}
			config.Hosts = append(config.Hosts, SyslogHost{
				Address: matches[1],
				Port:    port,
			})
			continue
		}

		// Try syslog host pattern (without port)
		if matches := hostPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.Hosts = append(config.Hosts, SyslogHost{
				Address: matches[1],
				Port:    0, // Default port (514)
			})
			continue
		}

		// Try local address pattern
		if matches := localAddressPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.LocalAddress = matches[1]
			continue
		}

		// Try facility pattern
		if matches := facilityPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.Facility = matches[1]
			continue
		}

		// Try notice pattern
		if matches := noticePattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.Notice = matches[1] == "on"
			continue
		}

		// Try info pattern
		if matches := infoPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.Info = matches[1] == "on"
			continue
		}

		// Try debug pattern
		if matches := debugPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.Debug = matches[1] == "on"
			continue
		}
	}

	return config, nil
}

// BuildSyslogHostCommand builds the command to add a syslog host
// Command format: syslog host <address> [<port>]
func BuildSyslogHostCommand(host SyslogHost) string {
	if host.Port > 0 && host.Port != 514 {
		return fmt.Sprintf("syslog host %s %d", host.Address, host.Port)
	}
	return fmt.Sprintf("syslog host %s", host.Address)
}

// BuildDeleteSyslogHostCommand builds the command to remove a syslog host
// Command format: no syslog host <address>
func BuildDeleteSyslogHostCommand(address string) string {
	return fmt.Sprintf("no syslog host %s", address)
}

// BuildSyslogLocalAddressCommand builds the command to set syslog local address
// Command format: syslog local address <ip>
func BuildSyslogLocalAddressCommand(address string) string {
	return fmt.Sprintf("syslog local address %s", address)
}

// BuildDeleteSyslogLocalAddressCommand builds the command to remove syslog local address
func BuildDeleteSyslogLocalAddressCommand() string {
	return "no syslog local address"
}

// BuildSyslogFacilityCommand builds the command to set syslog facility
// Command format: syslog facility <facility>
func BuildSyslogFacilityCommand(facility string) string {
	return fmt.Sprintf("syslog facility %s", facility)
}

// BuildDeleteSyslogFacilityCommand builds the command to remove syslog facility setting
func BuildDeleteSyslogFacilityCommand() string {
	return "no syslog facility"
}

// BuildSyslogLevelCommand builds the command to set a syslog level
// Command format: syslog <level> on|off
func BuildSyslogLevelCommand(level string, enabled bool) string {
	state := "off"
	if enabled {
		state = "on"
	}
	return fmt.Sprintf("syslog %s %s", level, state)
}

// BuildSyslogNoticeCommand builds the command to set syslog notice level
func BuildSyslogNoticeCommand(enabled bool) string {
	return BuildSyslogLevelCommand("notice", enabled)
}

// BuildSyslogInfoCommand builds the command to set syslog info level
func BuildSyslogInfoCommand(enabled bool) string {
	return BuildSyslogLevelCommand("info", enabled)
}

// BuildSyslogDebugCommand builds the command to set syslog debug level
func BuildSyslogDebugCommand(enabled bool) string {
	return BuildSyslogLevelCommand("debug", enabled)
}

// BuildDeleteSyslogCommand builds commands to completely remove syslog configuration
func BuildDeleteSyslogCommand(config *SyslogConfig) []string {
	var commands []string

	// Remove all hosts first
	for _, host := range config.Hosts {
		commands = append(commands, BuildDeleteSyslogHostCommand(host.Address))
	}

	// Remove local address if set
	if config.LocalAddress != "" {
		commands = append(commands, BuildDeleteSyslogLocalAddressCommand())
	}

	// Remove facility if set
	if config.Facility != "" {
		commands = append(commands, BuildDeleteSyslogFacilityCommand())
	}

	// Turn off all log levels
	if config.Notice {
		commands = append(commands, BuildSyslogNoticeCommand(false))
	}
	if config.Info {
		commands = append(commands, BuildSyslogInfoCommand(false))
	}
	if config.Debug {
		commands = append(commands, BuildSyslogDebugCommand(false))
	}

	return commands
}

// BuildShowSyslogConfigCommand builds the command to show syslog configuration
// Command format: show config | grep syslog
func BuildShowSyslogConfigCommand() string {
	return `show config | grep syslog`
}

// ValidateSyslogConfig validates a syslog configuration
func ValidateSyslogConfig(config *SyslogConfig) error {
	// Validate hosts
	for i, host := range config.Hosts {
		if host.Address == "" {
			return fmt.Errorf("host[%d]: address is required", i)
		}
		// Validate IP address format
		if !isValidSyslogHost(host.Address) {
			return fmt.Errorf("host[%d]: invalid address format: %s", i, host.Address)
		}
		// Validate port range
		if host.Port < 0 || host.Port > 65535 {
			return fmt.Errorf("host[%d]: port must be between 0 and 65535, got %d", i, host.Port)
		}
	}

	// Validate local address if set
	if config.LocalAddress != "" {
		if !isValidIP(config.LocalAddress) {
			return fmt.Errorf("invalid local_address format: %s", config.LocalAddress)
		}
	}

	// Validate facility if set
	if config.Facility != "" {
		if !isValidFacility(config.Facility) {
			return fmt.Errorf("invalid facility: %s (expected: user, local0-local7)", config.Facility)
		}
	}

	return nil
}

// isValidSyslogHost checks if a string is a valid syslog host (IP address or hostname)
func isValidSyslogHost(host string) bool {
	// Check if it's a valid IP address
	if net.ParseIP(host) != nil {
		return true
	}
	// Check if it's a valid hostname pattern (basic validation)
	// Hostname can contain alphanumeric characters, dots, and hyphens
	hostnamePattern := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-\.]*[a-zA-Z0-9])?$`)
	return hostnamePattern.MatchString(host)
}

// isValidFacility checks if a string is a valid syslog facility
func isValidFacility(facility string) bool {
	validFacilities := []string{
		"user", "local0", "local1", "local2", "local3",
		"local4", "local5", "local6", "local7",
	}
	for _, valid := range validFacilities {
		if facility == valid {
			return true
		}
	}
	return false
}
