package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// NetVolanteConfig represents NetVolante DNS configuration on an RTX router
type NetVolanteConfig struct {
	Hostname     string `json:"hostname"`      // netvolante-dns hostname host <interface> <hostname>
	Interface    string `json:"interface"`     // Interface name (pp 1, lan1, etc.)
	Server       int    `json:"server"`        // netvolante-dns server 1|2
	Timeout      int    `json:"timeout"`       // netvolante-dns timeout <seconds> (default: 90)
	IPv6         bool   `json:"ipv6"`          // netvolante-dns use ipv6 <interface> on|off
	AutoHostname bool   `json:"auto_hostname"` // netvolante-dns auto hostname <interface> on|off
	Use          bool   `json:"use"`           // netvolante-dns use <interface> on|off
}

// DDNSServerConfig represents custom DDNS server configuration
type DDNSServerConfig struct {
	ID       int    `json:"id"`       // Server ID (1-4)
	URL      string `json:"url"`      // ddns server url <id> <url>
	Hostname string `json:"hostname"` // ddns server hostname <id> <hostname>
	Username string `json:"username"` // ddns server user <id> <username> <password>
	Password string `json:"password"` // Password (sensitive)
}

// DDNSConfig represents the overall DDNS configuration
type DDNSConfig struct {
	NetVolante    []NetVolanteConfig `json:"netvolante"`     // NetVolante DNS configurations
	CustomServers []DDNSServerConfig `json:"custom_servers"` // Custom DDNS server configurations
}

// DDNSStatus represents the status of DDNS registration
type DDNSStatus struct {
	Type         string `json:"type"`          // "netvolante" or "custom"
	Interface    string `json:"interface"`     // Interface name
	ServerID     int    `json:"server_id"`     // Server ID (for custom DDNS)
	Hostname     string `json:"hostname"`      // Registered hostname
	CurrentIP    string `json:"current_ip"`    // Currently registered IP address
	LastUpdate   string `json:"last_update"`   // Last update timestamp
	Status       string `json:"status"`        // Status: registered, updating, error
	ErrorMessage string `json:"error_message"` // Error message if any
}

// DDNSParser parses DDNS configuration output
type DDNSParser struct{}

// NewDDNSParser creates a new DDNS parser
func NewDDNSParser() *DDNSParser {
	return &DDNSParser{}
}

// ValidateHostname validates a hostname for DDNS
func ValidateHostname(hostname string) error {
	if hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}

	// RFC 1123 hostname validation
	// Hostname can contain letters, digits, hyphens, and dots
	// Each label must start and end with alphanumeric
	hostnamePattern := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if !hostnamePattern.MatchString(hostname) {
		return fmt.Errorf("invalid hostname format: %s", hostname)
	}

	// Maximum hostname length is 253 characters
	if len(hostname) > 253 {
		return fmt.Errorf("hostname too long: %d characters (max 253)", len(hostname))
	}

	return nil
}

// ValidateDDNSURL validates a DDNS update URL
func ValidateDDNSURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Basic URL validation - must start with http:// or https://
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("URL must start with http:// or https://: %s", url)
	}

	// URL should have a host part
	urlPattern := regexp.MustCompile(`^https?://[a-zA-Z0-9][a-zA-Z0-9\-\.]*[a-zA-Z0-9](/.*)?$`)
	if !urlPattern.MatchString(url) {
		return fmt.Errorf("invalid URL format: %s", url)
	}

	return nil
}

// ParseNetVolanteDNS parses NetVolante DNS configuration from router output
func (p *DDNSParser) ParseNetVolanteDNS(raw string) ([]NetVolanteConfig, error) {
	configs := make(map[string]*NetVolanteConfig) // key: interface name
	lines := strings.Split(raw, "\n")

	// Interface patterns: "pp 1", "pp1", "lan1", "tunnel 1", etc.
	// Pattern: netvolante-dns hostname host <interface> <hostname>
	// Interface can be "pp 1", "pp1", "lan1", "tunnel 1", etc.
	hostnamePattern := regexp.MustCompile(`^\s*netvolante-dns\s+hostname\s+host\s+((?:pp|tunnel)\s*\d+|lan\d+)\s+(\S+)\s*$`)
	// Pattern: netvolante-dns server <1|2>
	serverPattern := regexp.MustCompile(`^\s*netvolante-dns\s+server\s+(\d+)\s*$`)
	// Pattern: netvolante-dns go <interface>
	goPattern := regexp.MustCompile(`^\s*netvolante-dns\s+go\s+((?:pp|tunnel)\s*\d+|lan\d+)\s*$`)
	// Pattern: netvolante-dns timeout <seconds>
	timeoutPattern := regexp.MustCompile(`^\s*netvolante-dns\s+timeout\s+(\d+)\s*$`)
	// Pattern: netvolante-dns use ipv6 <interface> on|off
	ipv6Pattern := regexp.MustCompile(`^\s*netvolante-dns\s+use\s+ipv6\s+((?:pp|tunnel)\s*\d+|lan\d+)\s+(on|off)\s*$`)
	// Pattern: netvolante-dns auto hostname <interface> on|off
	autoHostnamePattern := regexp.MustCompile(`^\s*netvolante-dns\s+auto\s+hostname\s+((?:pp|tunnel)\s*\d+|lan\d+)\s+(on|off)\s*$`)
	// Pattern: netvolante-dns use <interface> on|off
	usePattern := regexp.MustCompile(`^\s*netvolante-dns\s+use\s+((?:pp|tunnel)\s*\d+|lan\d+)\s+(on|off)\s*$`)

	globalServer := 1
	globalTimeout := 90

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse hostname host
		if matches := hostnamePattern.FindStringSubmatch(line); len(matches) >= 3 {
			iface := normalizeInterface(matches[1])
			if _, exists := configs[iface]; !exists {
				configs[iface] = &NetVolanteConfig{
					Interface: iface,
					Server:    globalServer,
					Timeout:   globalTimeout,
					Use:       true, // default enabled if hostname is set
				}
			}
			configs[iface].Hostname = matches[2]
			continue
		}

		// Parse server selection (global)
		if matches := serverPattern.FindStringSubmatch(line); len(matches) >= 2 {
			if server, err := strconv.Atoi(matches[1]); err == nil {
				globalServer = server
				// Update all existing configs
				for _, cfg := range configs {
					cfg.Server = server
				}
			}
			continue
		}

		// Parse timeout (global)
		if matches := timeoutPattern.FindStringSubmatch(line); len(matches) >= 2 {
			if timeout, err := strconv.Atoi(matches[1]); err == nil {
				globalTimeout = timeout
				// Update all existing configs
				for _, cfg := range configs {
					cfg.Timeout = timeout
				}
			}
			continue
		}

		// Parse IPv6 use
		if matches := ipv6Pattern.FindStringSubmatch(line); len(matches) >= 3 {
			iface := normalizeInterface(matches[1])
			if _, exists := configs[iface]; !exists {
				configs[iface] = &NetVolanteConfig{
					Interface: iface,
					Server:    globalServer,
					Timeout:   globalTimeout,
				}
			}
			configs[iface].IPv6 = matches[2] == "on"
			continue
		}

		// Parse auto hostname
		if matches := autoHostnamePattern.FindStringSubmatch(line); len(matches) >= 3 {
			iface := normalizeInterface(matches[1])
			if _, exists := configs[iface]; !exists {
				configs[iface] = &NetVolanteConfig{
					Interface: iface,
					Server:    globalServer,
					Timeout:   globalTimeout,
				}
			}
			configs[iface].AutoHostname = matches[2] == "on"
			continue
		}

		// Parse use (enable/disable)
		// Must be after ipv6 pattern to avoid matching "netvolante-dns use ipv6 ..."
		if matches := usePattern.FindStringSubmatch(line); len(matches) >= 3 {
			// Skip if this is an IPv6 use command (already handled)
			if matches[1] == "ipv6" {
				continue
			}
			iface := normalizeInterface(matches[1])
			if _, exists := configs[iface]; !exists {
				configs[iface] = &NetVolanteConfig{
					Interface: iface,
					Server:    globalServer,
					Timeout:   globalTimeout,
				}
			}
			configs[iface].Use = matches[2] == "on"
			continue
		}

		// Parse go command (triggers registration, indicates interface is configured)
		if matches := goPattern.FindStringSubmatch(line); len(matches) >= 2 {
			iface := normalizeInterface(matches[1])
			if _, exists := configs[iface]; !exists {
				configs[iface] = &NetVolanteConfig{
					Interface: iface,
					Server:    globalServer,
					Timeout:   globalTimeout,
					Use:       true,
				}
			}
			continue
		}
	}

	// Convert map to slice
	result := make([]NetVolanteConfig, 0, len(configs))
	for _, cfg := range configs {
		result = append(result, *cfg)
	}

	return result, nil
}

// normalizeInterface normalizes interface names (e.g., "pp1" -> "pp 1")
func normalizeInterface(iface string) string {
	// Handle "pp1", "pp2", etc.
	ppPattern := regexp.MustCompile(`^pp(\d+)$`)
	if matches := ppPattern.FindStringSubmatch(iface); len(matches) >= 2 {
		return "pp " + matches[1]
	}
	// Handle "tunnel1", "tunnel2", etc.
	tunnelPattern := regexp.MustCompile(`^tunnel(\d+)$`)
	if matches := tunnelPattern.FindStringSubmatch(iface); len(matches) >= 2 {
		return "tunnel " + matches[1]
	}
	return iface
}

// ParseDDNSConfig parses custom DDNS server configuration from router output
func (p *DDNSParser) ParseDDNSConfig(raw string) ([]DDNSServerConfig, error) {
	configs := make(map[int]*DDNSServerConfig)
	lines := strings.Split(raw, "\n")

	// Pattern: ddns server url <id> <url>
	urlPattern := regexp.MustCompile(`^\s*ddns\s+server\s+url\s+(\d+)\s+(\S+)\s*$`)
	// Pattern: ddns server hostname <id> <hostname>
	hostnamePattern := regexp.MustCompile(`^\s*ddns\s+server\s+hostname\s+(\d+)\s+(\S+)\s*$`)
	// Pattern: ddns server user <id> <username> <password>
	// Note: password may contain special characters, so we match until end of line
	userPattern := regexp.MustCompile(`^\s*ddns\s+server\s+user\s+(\d+)\s+(\S+)\s+(.+?)\s*$`)
	// Pattern: ddns server go <id>
	goPattern := regexp.MustCompile(`^\s*ddns\s+server\s+go\s+(\d+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse URL
		if matches := urlPattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}
			if _, exists := configs[id]; !exists {
				configs[id] = &DDNSServerConfig{ID: id}
			}
			configs[id].URL = matches[2]
			continue
		}

		// Parse hostname
		if matches := hostnamePattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}
			if _, exists := configs[id]; !exists {
				configs[id] = &DDNSServerConfig{ID: id}
			}
			configs[id].Hostname = matches[2]
			continue
		}

		// Parse user credentials
		if matches := userPattern.FindStringSubmatch(line); len(matches) >= 4 {
			id, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}
			if _, exists := configs[id]; !exists {
				configs[id] = &DDNSServerConfig{ID: id}
			}
			configs[id].Username = matches[2]
			configs[id].Password = strings.TrimSpace(matches[3])
			continue
		}

		// Parse go command (indicates server is configured)
		if matches := goPattern.FindStringSubmatch(line); len(matches) >= 2 {
			id, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}
			if _, exists := configs[id]; !exists {
				configs[id] = &DDNSServerConfig{ID: id}
			}
			continue
		}
	}

	// Convert map to slice, sorted by ID
	result := make([]DDNSServerConfig, 0, len(configs))
	for i := 1; i <= 4; i++ { // DDNS server IDs are typically 1-4
		if cfg, exists := configs[i]; exists {
			result = append(result, *cfg)
		}
	}

	return result, nil
}

// ParseDDNSStatus parses DDNS status from "show status netvolante-dns" or "show status ddns" output
func (p *DDNSParser) ParseDDNSStatus(raw string, statusType string) ([]DDNSStatus, error) {
	var statuses []DDNSStatus
	lines := strings.Split(raw, "\n")

	if statusType == "netvolante" {
		statuses = p.parseNetVolanteStatus(lines)
	} else if statusType == "custom" {
		statuses = p.parseCustomDDNSStatus(lines)
	} else {
		// Try to detect type from content
		if strings.Contains(raw, "netvolante") || strings.Contains(raw, "NetVolante") {
			statuses = p.parseNetVolanteStatus(lines)
		} else {
			statuses = p.parseCustomDDNSStatus(lines)
		}
	}

	return statuses, nil
}

// parseNetVolanteStatus parses NetVolante DNS status
func (p *DDNSParser) parseNetVolanteStatus(lines []string) []DDNSStatus {
	var statuses []DDNSStatus

	// Patterns for NetVolante DNS status
	// Format may vary, common patterns:
	// Interface: pp 1
	// Hostname: myhost.aa0.netvolante.jp
	// IP Address: 203.0.113.1
	// Status: registered
	// Last Update: 2024-01-20 10:30:00

	interfacePattern := regexp.MustCompile(`(?i)interface\s*:\s*(\S+.*)`)
	hostnamePattern := regexp.MustCompile(`(?i)hostname\s*:\s*(\S+)`)
	ipPattern := regexp.MustCompile(`(?i)ip\s*address\s*:\s*(\S+)`)
	statusPattern := regexp.MustCompile(`(?i)status\s*:\s*(\S+)`)
	lastUpdatePattern := regexp.MustCompile(`(?i)last\s*update\s*:\s*(.+)`)
	errorPattern := regexp.MustCompile(`(?i)error\s*:\s*(.+)`)

	var current *DDNSStatus

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if current != nil && current.Hostname != "" {
				current.Type = "netvolante"
				statuses = append(statuses, *current)
				current = nil
			}
			continue
		}

		if current == nil {
			current = &DDNSStatus{}
		}

		if matches := interfacePattern.FindStringSubmatch(line); len(matches) >= 2 {
			current.Interface = strings.TrimSpace(matches[1])
			continue
		}

		if matches := hostnamePattern.FindStringSubmatch(line); len(matches) >= 2 {
			current.Hostname = matches[1]
			continue
		}

		if matches := ipPattern.FindStringSubmatch(line); len(matches) >= 2 {
			current.CurrentIP = matches[1]
			continue
		}

		if matches := statusPattern.FindStringSubmatch(line); len(matches) >= 2 {
			current.Status = strings.ToLower(matches[1])
			continue
		}

		if matches := lastUpdatePattern.FindStringSubmatch(line); len(matches) >= 2 {
			current.LastUpdate = strings.TrimSpace(matches[1])
			continue
		}

		if matches := errorPattern.FindStringSubmatch(line); len(matches) >= 2 {
			current.ErrorMessage = strings.TrimSpace(matches[1])
			continue
		}
	}

	// Don't forget the last entry
	if current != nil && current.Hostname != "" {
		current.Type = "netvolante"
		statuses = append(statuses, *current)
	}

	return statuses
}

// parseCustomDDNSStatus parses custom DDNS status
func (p *DDNSParser) parseCustomDDNSStatus(lines []string) []DDNSStatus {
	var statuses []DDNSStatus

	// Patterns for custom DDNS status
	// Format may vary, common patterns:
	// Server 1:
	//   URL: https://dynupdate.no-ip.com/nic/update
	//   Hostname: myhost.no-ip.org
	//   IP Address: 203.0.113.1
	//   Status: ok
	//   Last Update: 2024-01-20 10:30:00

	serverPattern := regexp.MustCompile(`(?i)server\s*(\d+)\s*:?`)
	hostnamePattern := regexp.MustCompile(`(?i)hostname\s*:\s*(\S+)`)
	ipPattern := regexp.MustCompile(`(?i)ip\s*address\s*:\s*(\S+)`)
	statusPattern := regexp.MustCompile(`(?i)status\s*:\s*(\S+)`)
	lastUpdatePattern := regexp.MustCompile(`(?i)last\s*update\s*:\s*(.+)`)
	errorPattern := regexp.MustCompile(`(?i)error\s*:\s*(.+)`)

	var current *DDNSStatus

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for new server section
		if matches := serverPattern.FindStringSubmatch(line); len(matches) >= 2 {
			if current != nil && current.Hostname != "" {
				current.Type = "custom"
				statuses = append(statuses, *current)
			}
			serverID, _ := strconv.Atoi(matches[1])
			current = &DDNSStatus{ServerID: serverID}
			continue
		}

		if current == nil {
			current = &DDNSStatus{}
		}

		if matches := hostnamePattern.FindStringSubmatch(line); len(matches) >= 2 {
			current.Hostname = matches[1]
			continue
		}

		if matches := ipPattern.FindStringSubmatch(line); len(matches) >= 2 {
			current.CurrentIP = matches[1]
			continue
		}

		if matches := statusPattern.FindStringSubmatch(line); len(matches) >= 2 {
			current.Status = strings.ToLower(matches[1])
			continue
		}

		if matches := lastUpdatePattern.FindStringSubmatch(line); len(matches) >= 2 {
			current.LastUpdate = strings.TrimSpace(matches[1])
			continue
		}

		if matches := errorPattern.FindStringSubmatch(line); len(matches) >= 2 {
			current.ErrorMessage = strings.TrimSpace(matches[1])
			continue
		}
	}

	// Don't forget the last entry
	if current != nil && current.Hostname != "" {
		current.Type = "custom"
		statuses = append(statuses, *current)
	}

	return statuses
}

// BuildNetVolanteHostnameCommand builds the command to set NetVolante DNS hostname
// Command format: netvolante-dns hostname host <interface> <hostname>
func BuildNetVolanteHostnameCommand(iface, hostname string) string {
	if iface == "" || hostname == "" {
		return ""
	}
	return fmt.Sprintf("netvolante-dns hostname host %s %s", iface, hostname)
}

// BuildNetVolanteServerCommand builds the command to select NetVolante DNS server
// Command format: netvolante-dns server <1|2>
func BuildNetVolanteServerCommand(server int) string {
	if server < 1 || server > 2 {
		return ""
	}
	return fmt.Sprintf("netvolante-dns server %d", server)
}

// BuildNetVolanteGoCommand builds the command to trigger NetVolante DNS registration
// Command format: netvolante-dns go <interface>
func BuildNetVolanteGoCommand(iface string) string {
	if iface == "" {
		return ""
	}
	return fmt.Sprintf("netvolante-dns go %s", iface)
}

// BuildNetVolanteTimeoutCommand builds the command to set NetVolante DNS timeout
// Command format: netvolante-dns timeout <seconds>
func BuildNetVolanteTimeoutCommand(seconds int) string {
	if seconds <= 0 {
		return ""
	}
	return fmt.Sprintf("netvolante-dns timeout %d", seconds)
}

// BuildNetVolanteIPv6Command builds the command to enable/disable IPv6 for NetVolante DNS
// Command format: netvolante-dns use ipv6 <interface> on|off
func BuildNetVolanteIPv6Command(iface string, enable bool) string {
	if iface == "" {
		return ""
	}
	onOff := "off"
	if enable {
		onOff = "on"
	}
	return fmt.Sprintf("netvolante-dns use ipv6 %s %s", iface, onOff)
}

// BuildNetVolanteAutoHostnameCommand builds the command to enable/disable auto hostname
// Command format: netvolante-dns auto hostname <interface> on|off
func BuildNetVolanteAutoHostnameCommand(iface string, enable bool) string {
	if iface == "" {
		return ""
	}
	onOff := "off"
	if enable {
		onOff = "on"
	}
	return fmt.Sprintf("netvolante-dns auto hostname %s %s", iface, onOff)
}

// BuildNetVolanteUseCommand builds the command to enable/disable NetVolante DNS for an interface
// Command format: netvolante-dns use <interface> on|off
func BuildNetVolanteUseCommand(iface string, enable bool) string {
	if iface == "" {
		return ""
	}
	onOff := "off"
	if enable {
		onOff = "on"
	}
	return fmt.Sprintf("netvolante-dns use %s %s", iface, onOff)
}

// BuildDeleteNetVolanteHostnameCommand builds the command to remove NetVolante DNS hostname
// Command format: no netvolante-dns hostname host <interface>
func BuildDeleteNetVolanteHostnameCommand(iface string) string {
	if iface == "" {
		return ""
	}
	return fmt.Sprintf("no netvolante-dns hostname host %s", iface)
}

// BuildNetVolanteCommand builds a complete set of commands for NetVolante DNS configuration
func BuildNetVolanteCommand(config NetVolanteConfig) []string {
	var commands []string

	// Set server if not default
	if config.Server > 0 && config.Server != 1 {
		if cmd := BuildNetVolanteServerCommand(config.Server); cmd != "" {
			commands = append(commands, cmd)
		}
	}

	// Set timeout if not default
	if config.Timeout > 0 && config.Timeout != 90 {
		if cmd := BuildNetVolanteTimeoutCommand(config.Timeout); cmd != "" {
			commands = append(commands, cmd)
		}
	}

	// Set hostname
	if config.Hostname != "" {
		if cmd := BuildNetVolanteHostnameCommand(config.Interface, config.Hostname); cmd != "" {
			commands = append(commands, cmd)
		}
	}

	// Set auto hostname if enabled
	if config.AutoHostname {
		if cmd := BuildNetVolanteAutoHostnameCommand(config.Interface, true); cmd != "" {
			commands = append(commands, cmd)
		}
	}

	// Set IPv6 if enabled
	if config.IPv6 {
		if cmd := BuildNetVolanteIPv6Command(config.Interface, true); cmd != "" {
			commands = append(commands, cmd)
		}
	}

	// Enable/disable use
	if cmd := BuildNetVolanteUseCommand(config.Interface, config.Use); cmd != "" {
		commands = append(commands, cmd)
	}

	// Trigger registration
	if config.Use && config.Interface != "" {
		if cmd := BuildNetVolanteGoCommand(config.Interface); cmd != "" {
			commands = append(commands, cmd)
		}
	}

	return commands
}

// BuildDDNSURLCommand builds the command to set custom DDNS server URL
// Command format: ddns server url <id> <url>
func BuildDDNSURLCommand(id int, url string) string {
	if id < 1 || id > 4 || url == "" {
		return ""
	}
	return fmt.Sprintf("ddns server url %d %s", id, url)
}

// BuildDDNSHostnameCommand builds the command to set custom DDNS hostname
// Command format: ddns server hostname <id> <hostname>
func BuildDDNSHostnameCommand(id int, hostname string) string {
	if id < 1 || id > 4 || hostname == "" {
		return ""
	}
	return fmt.Sprintf("ddns server hostname %d %s", id, hostname)
}

// BuildDDNSUserCommand builds the command to set custom DDNS credentials
// Command format: ddns server user <id> <username> <password>
func BuildDDNSUserCommand(id int, username, password string) string {
	if id < 1 || id > 4 || username == "" || password == "" {
		return ""
	}
	return fmt.Sprintf("ddns server user %d %s %s", id, username, password)
}

// BuildDDNSGoCommand builds the command to trigger custom DDNS update
// Command format: ddns server go <id>
func BuildDDNSGoCommand(id int) string {
	if id < 1 || id > 4 {
		return ""
	}
	return fmt.Sprintf("ddns server go %d", id)
}

// BuildDeleteDDNSURLCommand builds the command to remove custom DDNS URL
// Command format: no ddns server url <id>
func BuildDeleteDDNSURLCommand(id int) string {
	if id < 1 || id > 4 {
		return ""
	}
	return fmt.Sprintf("no ddns server url %d", id)
}

// BuildDeleteDDNSHostnameCommand builds the command to remove custom DDNS hostname
// Command format: no ddns server hostname <id>
func BuildDeleteDDNSHostnameCommand(id int) string {
	if id < 1 || id > 4 {
		return ""
	}
	return fmt.Sprintf("no ddns server hostname %d", id)
}

// BuildDeleteDDNSUserCommand builds the command to remove custom DDNS credentials
// Command format: no ddns server user <id>
func BuildDeleteDDNSUserCommand(id int) string {
	if id < 1 || id > 4 {
		return ""
	}
	return fmt.Sprintf("no ddns server user %d", id)
}

// BuildDDNSCommand builds a complete set of commands for custom DDNS configuration
func BuildDDNSCommand(config DDNSServerConfig) []string {
	var commands []string

	// Set URL
	if config.URL != "" {
		if cmd := BuildDDNSURLCommand(config.ID, config.URL); cmd != "" {
			commands = append(commands, cmd)
		}
	}

	// Set hostname
	if config.Hostname != "" {
		if cmd := BuildDDNSHostnameCommand(config.ID, config.Hostname); cmd != "" {
			commands = append(commands, cmd)
		}
	}

	// Set credentials
	if config.Username != "" && config.Password != "" {
		if cmd := BuildDDNSUserCommand(config.ID, config.Username, config.Password); cmd != "" {
			commands = append(commands, cmd)
		}
	}

	// Trigger update
	if config.URL != "" {
		if cmd := BuildDDNSGoCommand(config.ID); cmd != "" {
			commands = append(commands, cmd)
		}
	}

	return commands
}

// BuildDeleteDDNSCommand builds commands to remove all custom DDNS configuration for a server ID
func BuildDeleteDDNSCommand(id int) []string {
	if id < 1 || id > 4 {
		return nil
	}
	return []string{
		BuildDeleteDDNSUserCommand(id),
		BuildDeleteDDNSHostnameCommand(id),
		BuildDeleteDDNSURLCommand(id),
	}
}

// BuildShowNetVolanteStatusCommand builds the command to show NetVolante DNS status
func BuildShowNetVolanteStatusCommand() string {
	return "show status netvolante-dns"
}

// BuildShowDDNSStatusCommand builds the command to show custom DDNS status
func BuildShowDDNSStatusCommand() string {
	return "show status ddns"
}

// ValidateNetVolanteConfig validates a NetVolante DNS configuration
func ValidateNetVolanteConfig(config NetVolanteConfig) error {
	if config.Interface == "" {
		return fmt.Errorf("interface cannot be empty")
	}

	if config.Hostname != "" {
		// NetVolante hostnames must end with .netvolante.jp or be in specific format
		if err := ValidateHostname(config.Hostname); err != nil {
			return fmt.Errorf("invalid hostname: %w", err)
		}
	}

	if config.Server != 0 && (config.Server < 1 || config.Server > 2) {
		return fmt.Errorf("server must be 1 or 2, got %d", config.Server)
	}

	if config.Timeout != 0 && config.Timeout < 1 {
		return fmt.Errorf("timeout must be positive, got %d", config.Timeout)
	}

	return nil
}

// ValidateDDNSServerConfig validates a custom DDNS server configuration
func ValidateDDNSServerConfig(config DDNSServerConfig) error {
	if config.ID < 1 || config.ID > 4 {
		return fmt.Errorf("server ID must be between 1 and 4, got %d", config.ID)
	}

	if config.URL != "" {
		if err := ValidateDDNSURL(config.URL); err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}
	}

	if config.Hostname != "" {
		if err := ValidateHostname(config.Hostname); err != nil {
			return fmt.Errorf("invalid hostname: %w", err)
		}
	}

	// If username is set, password must also be set
	if config.Username != "" && config.Password == "" {
		return fmt.Errorf("password is required when username is specified")
	}

	return nil
}
