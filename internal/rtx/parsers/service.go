package parsers

import (
	"crypto/sha256"
	"encoding/base64"
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

// BuildShowSSHDStatusCommand builds the command to show SSHD host key
// Command format: show sshd host key
// Note: "show status sshd" is not supported on RTX routers
func BuildShowSSHDStatusCommand() string {
	return "show sshd host key"
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

// ParseSSHDHostKeyInfo parses host key information from "show sshd host key" output
// The output contains public keys in OpenSSH format:
//   - ssh-rsa AAAAB3NzaC1yc2E...
//   - ssh-dss AAAAB3NzaC1kc3M...
//
// Returns empty strings if no host key is found (not an error).
func ParseSSHDHostKeyInfo(output string) *SSHHostKeyInfo {
	info := &SSHHostKeyInfo{}

	// RTX router wraps long keys across multiple lines
	// We need to join continuation lines that are part of the key data
	lines := strings.Split(output, "\n")

	// Pattern for OpenSSH public key start: "ssh-rsa AAAA..." or "ssh-dss AAAA..."
	publicKeyStartPattern := regexp.MustCompile(`^(ssh-rsa|ssh-dss|ssh-ed25519|ecdsa-sha2-nistp\d+)\s+([A-Za-z0-9+/=]+)`)

	// Pattern for continuation lines (base64 characters only, may end with =)
	continuationPattern := regexp.MustCompile(`^[A-Za-z0-9+/=]+$`)

	// Pattern for prompt (indicates end of key data)
	promptPattern := regexp.MustCompile(`\[RTX\d+\]`)

	var currentAlgorithm string
	var currentKeyData strings.Builder

	flushCurrentKey := func() {
		if currentAlgorithm != "" && currentKeyData.Len() > 0 {
			keyData := currentKeyData.String()
			// Prefer RSA over DSS if we find multiple keys
			if info.Algorithm == "" || currentAlgorithm == "ssh-rsa" {
				info.Algorithm = currentAlgorithm
				info.Fingerprint = computeSSHFingerprint(keyData)
			}
		}
		currentAlgorithm = ""
		currentKeyData.Reset()
	}

	for _, line := range lines {
		// Normalize line endings and trim
		line = strings.TrimSpace(strings.ReplaceAll(line, "\r", ""))
		if line == "" {
			continue
		}

		// Check for prompt (end of output)
		if promptPattern.MatchString(line) {
			flushCurrentKey()
			break
		}

		// Try to match start of a new public key
		if matches := publicKeyStartPattern.FindStringSubmatch(line); len(matches) >= 3 {
			// Flush previous key if any
			flushCurrentKey()

			// Start new key
			currentAlgorithm = matches[1]
			currentKeyData.WriteString(matches[2])
			continue
		}

		// If we're collecting key data, check for continuation line
		if currentAlgorithm != "" && continuationPattern.MatchString(line) {
			currentKeyData.WriteString(line)
			continue
		}

		// Non-matching line while collecting - flush and reset
		if currentAlgorithm != "" {
			flushCurrentKey()
		}
	}

	// Flush any remaining key
	flushCurrentKey()

	return info
}

// computeSSHFingerprint computes the SHA256 fingerprint of an SSH public key
// Input: base64-encoded public key data (the middle part of "ssh-rsa AAAA... comment")
// Output: "SHA256:base64hash" format matching OpenSSH standard
func computeSSHFingerprint(keyDataBase64 string) string {
	// Empty input returns empty fingerprint
	if keyDataBase64 == "" {
		return ""
	}

	// Decode the base64 key data
	keyBytes, err := base64.StdEncoding.DecodeString(keyDataBase64)
	if err != nil {
		// If decoding fails, return empty string
		return ""
	}

	// Empty decoded bytes returns empty fingerprint
	if len(keyBytes) == 0 {
		return ""
	}

	// Compute SHA256 hash
	hash := sha256.Sum256(keyBytes)

	// Encode hash as base64 (OpenSSH uses base64, not hex)
	// Remove trailing '=' padding to match OpenSSH format
	fingerprint := base64.StdEncoding.EncodeToString(hash[:])
	fingerprint = strings.TrimRight(fingerprint, "=")

	return "SHA256:" + fingerprint
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
// RTX returns public keys in OpenSSH format, potentially wrapped across lines:
//
//	ssh-rsa AAAAB3NzaC1yc2E...
//	...continuation (base64)...
//	...= (end of base64)
//	user@host.local (comment on separate line)
//	ssh-ed25519 AAAAC3NzaC1lZDI1NTE5...
//
// Keys may be wrapped across multiple lines, and comments may be on separate lines.
// Returns an empty slice if no keys are found
func ParseSSHDAuthorizedKeys(output string) ([]SSHAuthorizedKey, error) {
	var keys []SSHAuthorizedKey

	lines := strings.Split(output, "\n")

	// RTX returns public keys in OpenSSH format, potentially wrapped across lines
	// Format: <type> <base64-key> <comment>
	// Keys may be wrapped across multiple lines, so we need to join them

	var currentKey strings.Builder
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip RTX prompts like "[RTX1210] #"
		if strings.HasPrefix(line, "[") {
			continue
		}

		// Check if this line starts a new key (starts with ssh- or ecdsa-)
		if strings.HasPrefix(line, "ssh-") || strings.HasPrefix(line, "ecdsa-") {
			// If we have a previous key being built, parse it first
			if currentKey.Len() > 0 {
				if key := parseOpenSSHKey(currentKey.String()); key != nil {
					keys = append(keys, *key)
				}
				currentKey.Reset()
			}
			currentKey.WriteString(line)
		} else if currentKey.Len() > 0 {
			// This is a continuation of the previous key (wrapped line)
			// Possible formats:
			// 1. Pure base64 continuation: "YlhH7Tlx24Q..."
			// 2. Pure comment: "user@host" (first char usually not base64-ish in context)
			// 3. Mixed: "keyDataEnd user@host" (base64 ending + space + comment)

			if strings.Contains(line, "@") {
				// Line contains @ - could be pure comment or mixed format
				spaceIdx := strings.Index(line, " ")
				if spaceIdx == -1 {
					// No space in line - might be weird email like "abc@def"
					// If starts with base64 char, it's continuation; otherwise add space
					if isBase64Char(rune(line[0])) {
						currentKey.WriteString(line)
					} else {
						currentKey.WriteString(" ")
						currentKey.WriteString(line)
					}
				} else if isBase64Char(rune(line[0])) {
					// Starts with base64 char and has space - line contains both key and comment
					// Join directly without extra space (space between key and comment is in the line)
					currentKey.WriteString(line)
				} else {
					// Doesn't start with base64 char - pure comment line
					currentKey.WriteString(" ")
					currentKey.WriteString(line)
				}
			} else {
				// No @ in line - this is pure base64 continuation
				currentKey.WriteString(line)
			}
		}
	}

	// Don't forget the last key
	if currentKey.Len() > 0 {
		if key := parseOpenSSHKey(currentKey.String()); key != nil {
			keys = append(keys, *key)
		}
	}

	return keys, nil
}

// parseOpenSSHKey parses a single OpenSSH format public key
// Format: <type> <base64-key> <comment>
// The base64 key ends with one of: A-Za-z0-9+/= and comment follows after space
func parseOpenSSHKey(keyLine string) *SSHAuthorizedKey {
	// First, extract the key type
	parts := strings.SplitN(keyLine, " ", 2)
	if len(parts) < 2 {
		return nil
	}

	keyType := parts[0]
	if !isValidSSHKeyType(keyType) {
		return nil
	}

	rest := strings.TrimSpace(parts[1])
	if rest == "" {
		return nil
	}

	// Find the boundary between base64 key and comment
	// The comment usually starts with something like "user@host" or similar
	// We look for the pattern: base64 followed by space followed by comment (containing @)
	var base64Key, comment string

	// Find the last occurrence of a pattern that looks like "= comment" or "key comment"
	// where comment contains @
	lastSpaceBeforeComment := -1
	for i := len(rest) - 1; i >= 0; i-- {
		if rest[i] == ' ' {
			remainingPart := rest[i+1:]
			if strings.Contains(remainingPart, "@") {
				lastSpaceBeforeComment = i
				break
			}
		}
	}

	if lastSpaceBeforeComment > 0 {
		base64Key = strings.TrimSpace(rest[:lastSpaceBeforeComment])
		comment = strings.TrimSpace(rest[lastSpaceBeforeComment+1:])
	} else {
		// No comment found, entire rest is the key
		base64Key = strings.TrimSpace(rest)
	}

	return &SSHAuthorizedKey{
		Type:        keyType,
		Fingerprint: base64Key,
		Comment:     comment,
	}
}

// isValidSSHKeyType checks if the string is a valid SSH key type
func isValidSSHKeyType(s string) bool {
	validTypes := []string{"ssh-rsa", "ssh-ed25519", "ssh-dss", "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521"}
	for _, t := range validTypes {
		if s == t {
			return true
		}
	}
	return false
}

// isBase64Char checks if a character is a valid base64 character
func isBase64Char(c rune) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '+' || c == '/' || c == '='
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
