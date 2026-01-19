package parsers

import (
	"fmt"
	"regexp"
	"strings"
)

// HTTPDConfig represents HTTP daemon configuration on an RTX router
type HTTPDConfig struct {
	Host        string `json:"host"`         // "any" or specific interface (e.g., "lan1")
	ProxyAccess bool   `json:"proxy_access"` // L2MS proxy access enabled
}

// SSHDConfig represents SSH daemon configuration on an RTX router
type SSHDConfig struct {
	Enabled bool     `json:"enabled"`            // sshd service on/off
	Hosts   []string `json:"hosts,omitempty"`    // Interface list (e.g., ["lan1", "lan2"])
	HostKey string   `json:"host_key,omitempty"` // RSA host key (sensitive)
}

// SFTPDConfig represents SFTP daemon configuration on an RTX router
type SFTPDConfig struct {
	Hosts []string `json:"hosts,omitempty"` // Interface list
}

// ServiceParser parses service daemon configuration output
type ServiceParser struct{}

// NewServiceParser creates a new service parser
func NewServiceParser() *ServiceParser {
	return &ServiceParser{}
}

// ParseHTTPDConfig parses HTTPD configuration from router output
// Parses lines like:
//   - httpd host any
//   - httpd host lan1
//   - httpd proxy-access l2ms permit on
func (p *ServiceParser) ParseHTTPDConfig(raw string) (*HTTPDConfig, error) {
	config := &HTTPDConfig{
		Host:        "", // Empty means not configured
		ProxyAccess: false,
	}

	lines := strings.Split(raw, "\n")

	// Pattern: httpd host <interface>
	hostPattern := regexp.MustCompile(`^\s*httpd\s+host\s+(\S+)\s*$`)
	// Pattern: httpd proxy-access l2ms permit on|off
	proxyPattern := regexp.MustCompile(`^\s*httpd\s+proxy-access\s+l2ms\s+permit\s+(on|off)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Try host pattern
		if matches := hostPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.Host = matches[1]
			continue
		}

		// Try proxy-access pattern
		if matches := proxyPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.ProxyAccess = matches[1] == "on"
			continue
		}
	}

	return config, nil
}

// ParseSSHDConfig parses SSHD configuration from router output
// Parses lines like:
//   - sshd service on
//   - sshd host lan1 lan2
//   - sshd host key generate
func (p *ServiceParser) ParseSSHDConfig(raw string) (*SSHDConfig, error) {
	config := &SSHDConfig{
		Enabled: false,
		Hosts:   []string{},
		HostKey: "",
	}

	lines := strings.Split(raw, "\n")

	// Pattern: sshd service on|off
	servicePattern := regexp.MustCompile(`^\s*sshd\s+service\s+(on|off)\s*$`)
	// Pattern: sshd host <interface1> [<interface2> ...]
	hostPattern := regexp.MustCompile(`^\s*sshd\s+host\s+(.+)\s*$`)
	// Pattern: sshd host key <key-data> (host key is on a single line)
	keyPattern := regexp.MustCompile(`^\s*sshd\s+host\s+key\s+(.+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Try service pattern
		if matches := servicePattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.Enabled = matches[1] == "on"
			continue
		}

		// Try host key pattern first (before host pattern)
		if matches := keyPattern.FindStringSubmatch(line); len(matches) >= 2 {
			keyValue := strings.TrimSpace(matches[1])
			// Skip "generate" keyword - that's not an actual key
			if keyValue != "generate" {
				config.HostKey = keyValue
			}
			continue
		}

		// Try host pattern (interface list)
		if matches := hostPattern.FindStringSubmatch(line); len(matches) >= 2 {
			hostsStr := strings.TrimSpace(matches[1])
			// Skip if this is a key command
			if strings.HasPrefix(hostsStr, "key ") {
				continue
			}
			// Parse space-separated interface list
			interfaces := strings.Fields(hostsStr)
			config.Hosts = append(config.Hosts, interfaces...)
			continue
		}
	}

	return config, nil
}

// ParseSFTPDConfig parses SFTPD configuration from router output
// Parses lines like:
//   - sftpd host lan1 lan2
func (p *ServiceParser) ParseSFTPDConfig(raw string) (*SFTPDConfig, error) {
	config := &SFTPDConfig{
		Hosts: []string{},
	}

	lines := strings.Split(raw, "\n")

	// Pattern: sftpd host <interface1> [<interface2> ...]
	hostPattern := regexp.MustCompile(`^\s*sftpd\s+host\s+(.+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Try host pattern
		if matches := hostPattern.FindStringSubmatch(line); len(matches) >= 2 {
			hostsStr := strings.TrimSpace(matches[1])
			// Parse space-separated interface list
			interfaces := strings.Fields(hostsStr)
			config.Hosts = append(config.Hosts, interfaces...)
			continue
		}
	}

	return config, nil
}

// ========== HTTPD Command Builders ==========

// BuildHTTPDHostCommand builds the command to set HTTPD host
// Command format: httpd host <any|interface>
func BuildHTTPDHostCommand(host string) string {
	if host == "" {
		return ""
	}
	return fmt.Sprintf("httpd host %s", host)
}

// BuildHTTPDProxyAccessCommand builds the command to set HTTPD proxy access
// Command format: httpd proxy-access l2ms permit on|off
func BuildHTTPDProxyAccessCommand(enabled bool) string {
	state := "off"
	if enabled {
		state = "on"
	}
	return fmt.Sprintf("httpd proxy-access l2ms permit %s", state)
}

// BuildDeleteHTTPDHostCommand builds the command to remove HTTPD host configuration
// Command format: no httpd host
func BuildDeleteHTTPDHostCommand() string {
	return "no httpd host"
}

// BuildDeleteHTTPDProxyAccessCommand builds the command to disable HTTPD proxy access
// Command format: httpd proxy-access l2ms permit off
func BuildDeleteHTTPDProxyAccessCommand() string {
	return "httpd proxy-access l2ms permit off"
}

// BuildShowHTTPDConfigCommand builds the command to show HTTPD configuration
// Command format: show config | grep httpd
func BuildShowHTTPDConfigCommand() string {
	return "show config | grep httpd"
}

// ========== SSHD Command Builders ==========

// BuildSSHDServiceCommand builds the command to enable/disable SSHD service
// Command format: sshd service on|off
func BuildSSHDServiceCommand(enabled bool) string {
	state := "off"
	if enabled {
		state = "on"
	}
	return fmt.Sprintf("sshd service %s", state)
}

// BuildSSHDHostCommand builds the command to set SSHD hosts
// Command format: sshd host <interface1> [<interface2> ...]
func BuildSSHDHostCommand(hosts []string) string {
	if len(hosts) == 0 {
		return ""
	}
	return fmt.Sprintf("sshd host %s", strings.Join(hosts, " "))
}

// BuildSSHDHostKeyGenerateCommand builds the command to generate SSH host key
// Command format: sshd host key generate
func BuildSSHDHostKeyGenerateCommand() string {
	return "sshd host key generate"
}

// BuildDeleteSSHDServiceCommand builds the command to disable SSHD service
// Command format: no sshd service
func BuildDeleteSSHDServiceCommand() string {
	return "no sshd service"
}

// BuildDeleteSSHDHostCommand builds the command to remove SSHD host configuration
// Command format: no sshd host
func BuildDeleteSSHDHostCommand() string {
	return "no sshd host"
}

// BuildShowSSHDConfigCommand builds the command to show SSHD configuration
// Command format: show config | grep sshd
func BuildShowSSHDConfigCommand() string {
	return "show config | grep sshd"
}

// ========== SFTPD Command Builders ==========

// BuildSFTPDHostCommand builds the command to set SFTPD hosts
// Command format: sftpd host <interface1> [<interface2> ...]
func BuildSFTPDHostCommand(hosts []string) string {
	if len(hosts) == 0 {
		return ""
	}
	return fmt.Sprintf("sftpd host %s", strings.Join(hosts, " "))
}

// BuildDeleteSFTPDHostCommand builds the command to remove SFTPD host configuration
// Command format: no sftpd host
func BuildDeleteSFTPDHostCommand() string {
	return "no sftpd host"
}

// BuildShowSFTPDConfigCommand builds the command to show SFTPD configuration
// Command format: show config | grep sftpd
func BuildShowSFTPDConfigCommand() string {
	return "show config | grep sftpd"
}

// ========== Validation Functions ==========

// ValidateHTTPDConfig validates HTTPD configuration
func ValidateHTTPDConfig(config HTTPDConfig) error {
	if config.Host == "" {
		return fmt.Errorf("host is required")
	}

	// Valid hosts: "any" or interface name (lan1, lan2, pp1, etc.)
	validHostPattern := regexp.MustCompile(`^(any|lan\d+|pp\d+|bridge\d+|tunnel\d+)$`)
	if !validHostPattern.MatchString(config.Host) {
		return fmt.Errorf("invalid host: %s (must be 'any' or interface name like lan1, pp1)", config.Host)
	}

	return nil
}

// ValidateSSHDConfig validates SSHD configuration
func ValidateSSHDConfig(config SSHDConfig) error {
	// Validate interface names
	validIfacePattern := regexp.MustCompile(`^(lan\d+|pp\d+|bridge\d+|tunnel\d+)$`)
	for _, host := range config.Hosts {
		if !validIfacePattern.MatchString(host) {
			return fmt.Errorf("invalid interface: %s (must be interface name like lan1, pp1)", host)
		}
	}

	return nil
}

// ValidateSFTPDConfig validates SFTPD configuration
func ValidateSFTPDConfig(config SFTPDConfig) error {
	if len(config.Hosts) == 0 {
		return fmt.Errorf("at least one host interface is required")
	}

	// Validate interface names
	validIfacePattern := regexp.MustCompile(`^(lan\d+|pp\d+|bridge\d+|tunnel\d+)$`)
	for _, host := range config.Hosts {
		if !validIfacePattern.MatchString(host) {
			return fmt.Errorf("invalid interface: %s (must be interface name like lan1, pp1)", host)
		}
	}

	return nil
}
