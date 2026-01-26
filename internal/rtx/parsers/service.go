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
	Enabled    bool     `json:"enabled"`               // sshd service on/off
	Hosts      []string `json:"hosts,omitempty"`       // Interface list (e.g., ["lan1", "lan2"])
	HostKey    string   `json:"host_key,omitempty"`    // RSA host key (sensitive)
	AuthMethod string   `json:"auth_method,omitempty"` // SSH auth method: "password", "publickey", or "any" (default)
}

// SFTPDConfig represents SFTP daemon configuration on an RTX router
type SFTPDConfig struct {
	Hosts []string `json:"hosts,omitempty"` // Interface list
}

// SSHHostKeyInfo represents SSH host key information from "show status sshd"
type SSHHostKeyInfo struct {
	Fingerprint string `json:"fingerprint,omitempty"` // Host key fingerprint (e.g., SHA256:xxxxx or colon-separated hex)
	Algorithm   string `json:"algorithm,omitempty"`   // Key algorithm (RSA, ECDSA, ED25519, etc.)
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
//   - sshd auth method password|publickey
func (p *ServiceParser) ParseSSHDConfig(raw string) (*SSHDConfig, error) {
	config := &SSHDConfig{
		Enabled:    false,
		Hosts:      []string{},
		HostKey:    "",
		AuthMethod: "any", // Default: "any" (both password and publickey allowed)
	}

	lines := strings.Split(raw, "\n")

	// Pattern: sshd service on|off
	servicePattern := regexp.MustCompile(`^\s*sshd\s+service\s+(on|off)\s*$`)
	// Pattern: sshd host <interface1> [<interface2> ...]
	hostPattern := regexp.MustCompile(`^\s*sshd\s+host\s+(.+)\s*$`)
	// Pattern: sshd host key <key-data> (host key is on a single line)
	keyPattern := regexp.MustCompile(`^\s*sshd\s+host\s+key\s+(.+)\s*$`)
	// Pattern: sshd auth method password|publickey
	authMethodPattern := regexp.MustCompile(`^\s*sshd\s+auth\s+method\s+(password|publickey)\s*$`)

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

		// Try auth method pattern
		if matches := authMethodPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.AuthMethod = matches[1]
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

// BuildShowSSHDStatusCommand builds the command to show SSHD status
// Command format: show status sshd
func BuildShowSSHDStatusCommand() string {
	return "show status sshd"
}

// BuildSSHDAuthMethodCommand builds the command to set SSHD authentication method
// Command format:
//   - sshd auth method password (password only)
//   - sshd auth method publickey (public key only)
//   - no sshd auth method (any - both allowed, default)
func BuildSSHDAuthMethodCommand(method string) string {
	switch method {
	case "password":
		return "sshd auth method password"
	case "publickey":
		return "sshd auth method publickey"
	case "any", "":
		return "no sshd auth method"
	default:
		return "no sshd auth method"
	}
}

// BuildDeleteSSHDAuthMethodCommand builds the command to remove SSHD auth method configuration
// Command format: no sshd auth method
// This resets to default (any - both password and publickey allowed)
func BuildDeleteSSHDAuthMethodCommand() string {
	return "no sshd auth method"
}

// ParseSSHDHostKeyInfo parses host key information from "show status sshd" output
// The output typically contains lines like:
//   - SSH Host Key (RSA): XX:XX:XX:...
//   - SSH Host Key (ECDSA): XX:XX:XX:...
//   - ホストキーのフィンガープリント: SHA256:xxxxx
//
// Returns empty strings if no host key is found (not an error).
func ParseSSHDHostKeyInfo(output string) *SSHHostKeyInfo {
	info := &SSHHostKeyInfo{}

	lines := strings.Split(output, "\n")

	// Pattern 1: English format - "SSH Host Key (ALGORITHM): FINGERPRINT"
	englishPattern := regexp.MustCompile(`^\s*SSH\s+Host\s+Key\s*\(([^)]+)\)\s*:\s*(.+)\s*$`)

	// Pattern 2: Japanese format - "ホストキーのフィンガープリント: FINGERPRINT"
	// This format may include algorithm prefix like "SHA256:"
	japanesePattern := regexp.MustCompile(`^\s*ホストキーのフィンガープリント\s*:\s*(.+)\s*$`)

	// Pattern 3: Algorithm line - "ホストキーのアルゴリズム: ALGORITHM" or "Key Algorithm: ALGORITHM"
	algorithmPattern := regexp.MustCompile(`(?:ホストキーのアルゴリズム|Key\s+Algorithm)\s*:\s*(\S+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try English pattern first
		if matches := englishPattern.FindStringSubmatch(line); len(matches) >= 3 {
			info.Algorithm = strings.TrimSpace(matches[1])
			info.Fingerprint = strings.TrimSpace(matches[2])
			continue
		}

		// Try Japanese fingerprint pattern
		if matches := japanesePattern.FindStringSubmatch(line); len(matches) >= 2 {
			fingerprint := strings.TrimSpace(matches[1])
			// Check if fingerprint contains algorithm prefix (e.g., "SHA256:xxx")
			if idx := strings.Index(fingerprint, ":"); idx > 0 && idx < 10 {
				// Short prefix before colon suggests algorithm prefix
				prefix := fingerprint[:idx]
				if prefix == "SHA256" || prefix == "MD5" || prefix == "SHA1" {
					// Keep full fingerprint including prefix
					info.Fingerprint = fingerprint
				} else {
					info.Fingerprint = fingerprint
				}
			} else {
				info.Fingerprint = fingerprint
			}
			continue
		}

		// Try algorithm pattern
		if matches := algorithmPattern.FindStringSubmatch(line); len(matches) >= 2 {
			info.Algorithm = strings.TrimSpace(matches[1])
			continue
		}
	}

	return info
}

// ========== SSHD Authorized Keys Command Builders ==========

// SSHAuthorizedKey represents an SSH authorized key entry for parser output
type SSHAuthorizedKey struct {
	Type        string // Key type (e.g., "ED25519", "RSA")
	Fingerprint string // SHA256 fingerprint
	Comment     string // Key comment (e.g., "user@host")
}

// BuildShowSSHDAuthorizedKeysCommand builds the command to show SSH authorized keys for a user
// Command format: show sshd authorized-keys <username>
func BuildShowSSHDAuthorizedKeysCommand(username string) string {
	return fmt.Sprintf("show sshd authorized-keys %s", username)
}

// BuildImportSSHDAuthorizedKeysCommand builds the command to import SSH authorized keys for a user
// Command format: import sshd authorized-keys <username>
func BuildImportSSHDAuthorizedKeysCommand(username string) string {
	return fmt.Sprintf("import sshd authorized-keys %s", username)
}

// BuildDeleteSSHDAuthorizedKeysCommand builds the command to delete SSH authorized keys for a user
// Command format: delete /ssh/authorized_keys/<username>
func BuildDeleteSSHDAuthorizedKeysCommand(username string) string {
	return fmt.Sprintf("delete /ssh/authorized_keys/%s", username)
}

// ParseSSHDAuthorizedKeys parses the output of "show sshd authorized-keys" command
// Output format example:
//
//	256 SHA256:xxxx user@host (ED25519)
//	2048 SHA256:yyyy admin@pc (RSA)
//
// Returns an empty slice if no keys are found
func ParseSSHDAuthorizedKeys(output string) ([]SSHAuthorizedKey, error) {
	var keys []SSHAuthorizedKey

	lines := strings.Split(output, "\n")

	// Pattern: <bits> <fingerprint> <comment> (<type>)
	// e.g.: 256 SHA256:xxxx user@host (ED25519)
	keyPattern := regexp.MustCompile(`^\s*\d+\s+(SHA256:\S+)\s+(.+?)\s+\((\w+)\)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if matches := keyPattern.FindStringSubmatch(line); len(matches) >= 4 {
			key := SSHAuthorizedKey{
				Fingerprint: matches[1],
				Comment:     matches[2],
				Type:        matches[3],
			}
			keys = append(keys, key)
		}
	}

	return keys, nil
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
