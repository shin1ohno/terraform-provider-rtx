package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DHCPScope represents a DHCP scope configuration on an RTX router
type DHCPScope struct {
	ScopeID       int              `json:"scope_id"`
	Network       string           `json:"network"`                  // CIDR notation: "192.168.1.0/24"
	RangeStart    string           `json:"range_start,omitempty"`    // Start IP of allocation range (if specified)
	RangeEnd      string           `json:"range_end,omitempty"`      // End IP of allocation range (if specified)
	LeaseTime     string           `json:"lease_time,omitempty"`     // Go duration format or "infinity"
	MaxLeaseTime  string           `json:"max_lease_time,omitempty"` // Maximum lease time when client requests longer (Go duration format or "infinity")
	ExcludeRanges []ExcludeRange   `json:"exclude_ranges,omitempty"` // Excluded IP ranges
	Options       DHCPScopeOptions `json:"options,omitempty"`        // DHCP options (dns, routers, etc.)
}

// DHCPScopeOptions represents DHCP options for a scope (Cisco-compatible naming)
type DHCPScopeOptions struct {
	DNSServers  []string `json:"dns_servers,omitempty"`  // DNS servers (max 3, option 6)
	Routers     []string `json:"routers,omitempty"`      // Default gateways (max 3, option 3)
	DomainName  string   `json:"domain_name,omitempty"`  // Domain name (option 15)
	Hostname    string   `json:"hostname,omitempty"`     // Hostname (option 12)
	WINSServers []string `json:"wins_servers,omitempty"` // WINS/NetBIOS name servers (max 3, option 44)
}

// ExcludeRange represents an IP range excluded from DHCP allocation
type ExcludeRange struct {
	Start string `json:"start"` // Start IP address
	End   string `json:"end"`   // End IP address
}

// DHCPScopeParser parses DHCP scope configuration output
type DHCPScopeParser struct{}

// NewDHCPScopeParser creates a new DHCP scope parser
func NewDHCPScopeParser() *DHCPScopeParser {
	return &DHCPScopeParser{}
}

// ParseScopeConfig parses the output of "show config | grep dhcp scope" command
// and returns a list of DHCP scopes
func (p *DHCPScopeParser) ParseScopeConfig(raw string) ([]DHCPScope, error) {
	scopes := make(map[int]*DHCPScope)
	lines := strings.Split(raw, "\n")

	// Patterns for different scope configuration lines
	// dhcp scope <id> <network>/<prefix> [gateway <ip>] [expire <time>] [maxexpire <time>]
	// Note: gateway is legacy format; modern RTX uses "dhcp scope option <id> router=..."
	// Both "dhcp scope 1 192.168.0.0/16 expire 24:00" (no gateway) and
	// "dhcp scope 1 192.168.0.0/16 gateway 192.168.0.1 expire 24:00" must be supported
	scopePattern := regexp.MustCompile(`^\s*dhcp\s+scope\s+(\d+)\s+([0-9.]+/\d+)(?:\s+gateway\s+([0-9.]+))?(?:\s+expire\s+(\S+))?(?:\s+maxexpire\s+(\S+))?.*$`)
	// Pattern for scope with expire but no gateway (expire comes right after network)
	scopeExpireOnlyPattern := regexp.MustCompile(`^\s*dhcp\s+scope\s+(\d+)\s+([0-9.]+/\d+)\s+expire\s+(\S+)(?:\s+maxexpire\s+(\S+))?.*$`)
	// Pattern for IP range format: dhcp scope <id> <start_ip>-<end_ip>/<mask> [gateway <ip>] [expire <time>] [maxexpire <time>]
	// e.g., "dhcp scope 1 192.168.1.20-192.168.1.99/16 gateway 192.168.1.253 expire 12:00"
	scopeRangePattern := regexp.MustCompile(`^\s*dhcp\s+scope\s+(\d+)\s+([0-9.]+)-([0-9.]+)/(\d+)(?:\s+gateway\s+([0-9.]+))?(?:\s+expire\s+(\S+))?(?:\s+maxexpire\s+(\S+))?.*$`)
	// dhcp scope option <id> dns=<dns1>[,<dns2>[,<dns3>]] [router=<gw1>[,<gw2>]] [domain=<domain>]
	optionPattern := regexp.MustCompile(`^\s*dhcp\s+scope\s+option\s+(\d+)\s+(.+)\s*$`)
	// dhcp scope <id> except <start>-<end>
	exceptPattern := regexp.MustCompile(`^\s*dhcp\s+scope\s+(\d+)\s+except\s+([0-9.]+)-([0-9.]+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try IP range pattern first (most specific pattern)
		// This handles "dhcp scope 1 192.168.1.20-192.168.1.99/16 gateway 192.168.1.253 expire 12:00 maxexpire 24:00"
		if matches := scopeRangePattern.FindStringSubmatch(line); len(matches) >= 5 {
			scopeID, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			scope, exists := scopes[scopeID]
			if !exists {
				scope = &DHCPScope{
					ScopeID:       scopeID,
					ExcludeRanges: []ExcludeRange{},
				}
				scopes[scopeID] = scope
			}

			scope.RangeStart = matches[2]
			scope.RangeEnd = matches[3]
			scope.Network = calculateNetworkAddress(matches[2], matches[4]) // Calculate proper network address from start IP and prefix
			// Gateway (legacy format) - convert to Options.Routers
			if len(matches) > 5 && matches[5] != "" {
				scope.Options.Routers = []string{matches[5]}
			}
			// Expire time
			if len(matches) > 6 && matches[6] != "" {
				scope.LeaseTime = convertRTXLeaseTimeToGo(matches[6])
			}
			// Max expire time
			if len(matches) > 7 && matches[7] != "" {
				scope.MaxLeaseTime = convertRTXLeaseTimeToGo(matches[7])
			}
			continue
		}

		// Try scope with expire only pattern (more specific pattern)
		// This handles "dhcp scope 1 192.168.0.0/16 expire 24:00 maxexpire 48:00" (no gateway)
		if matches := scopeExpireOnlyPattern.FindStringSubmatch(line); len(matches) >= 4 {
			scopeID, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			scope, exists := scopes[scopeID]
			if !exists {
				scope = &DHCPScope{
					ScopeID:       scopeID,
					ExcludeRanges: []ExcludeRange{},
				}
				scopes[scopeID] = scope
			}

			scope.Network = matches[2]
			scope.LeaseTime = convertRTXLeaseTimeToGo(matches[3])
			// Max expire time
			if len(matches) > 4 && matches[4] != "" {
				scope.MaxLeaseTime = convertRTXLeaseTimeToGo(matches[4])
			}
			continue
		}

		// Try scope definition pattern (with optional gateway)
		if matches := scopePattern.FindStringSubmatch(line); len(matches) >= 3 {
			scopeID, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			scope, exists := scopes[scopeID]
			if !exists {
				scope = &DHCPScope{
					ScopeID:       scopeID,
					ExcludeRanges: []ExcludeRange{},
				}
				scopes[scopeID] = scope
			}

			scope.Network = matches[2]
			// Legacy gateway support: convert to Options.Routers
			if len(matches) > 3 && matches[3] != "" {
				scope.Options.Routers = []string{matches[3]}
			}
			if len(matches) > 4 && matches[4] != "" {
				scope.LeaseTime = convertRTXLeaseTimeToGo(matches[4])
			}
			// Max expire time
			if len(matches) > 5 && matches[5] != "" {
				scope.MaxLeaseTime = convertRTXLeaseTimeToGo(matches[5])
			}
			continue
		}

		// Try option pattern (dns=, router=, domain=)
		if matches := optionPattern.FindStringSubmatch(line); len(matches) >= 3 {
			scopeID, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			scope, exists := scopes[scopeID]
			if !exists {
				scope = &DHCPScope{
					ScopeID:       scopeID,
					ExcludeRanges: []ExcludeRange{},
				}
				scopes[scopeID] = scope
			}

			// Parse option string (e.g., "dns=1.1.1.1,8.8.8.8 router=192.168.1.1")
			optionStr := matches[2]
			parseOptions(optionStr, &scope.Options)
			continue
		}

		// Try except (exclusion) pattern
		if matches := exceptPattern.FindStringSubmatch(line); len(matches) >= 4 {
			scopeID, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			scope, exists := scopes[scopeID]
			if !exists {
				scope = &DHCPScope{
					ScopeID:       scopeID,
					ExcludeRanges: []ExcludeRange{},
				}
				scopes[scopeID] = scope
			}

			scope.ExcludeRanges = append(scope.ExcludeRanges, ExcludeRange{
				Start: matches[2],
				End:   matches[3],
			})
			continue
		}
	}

	// Convert map to slice
	result := make([]DHCPScope, 0, len(scopes))
	for _, scope := range scopes {
		result = append(result, *scope)
	}

	return result, nil
}

// ParseSingleScope parses configuration for a specific scope ID
func (p *DHCPScopeParser) ParseSingleScope(raw string, scopeID int) (*DHCPScope, error) {
	scopes, err := p.ParseScopeConfig(raw)
	if err != nil {
		return nil, err
	}

	for _, scope := range scopes {
		if scope.ScopeID == scopeID {
			return &scope, nil
		}
	}

	return nil, fmt.Errorf("scope %d not found", scopeID)
}

// parseOptions parses option string like "dns=1.1.1.1,8.8.8.8 router=192.168.1.1 domain=example.com hostname=router wins_server=192.168.1.10"
func parseOptions(optionStr string, opts *DHCPScopeOptions) {
	// Split by space to get individual key=value pairs
	parts := strings.Fields(optionStr)
	for _, part := range parts {
		if idx := strings.Index(part, "="); idx != -1 {
			key := strings.ToLower(part[:idx])
			value := part[idx+1:]

			switch key {
			case "dns", "6": // DNS servers (option 6)
				servers := strings.Split(value, ",")
				for _, s := range servers {
					s = strings.TrimSpace(s)
					if s != "" {
						opts.DNSServers = append(opts.DNSServers, s)
					}
				}
			case "router", "3": // Default gateways (option 3)
				routers := strings.Split(value, ",")
				for _, r := range routers {
					r = strings.TrimSpace(r)
					if r != "" {
						opts.Routers = append(opts.Routers, r)
					}
				}
			case "domain", "15": // Domain name (option 15)
				opts.DomainName = value
			case "hostname", "12": // Hostname (option 12)
				opts.Hostname = value
			case "wins_server", "44": // WINS/NetBIOS name servers (option 44)
				servers := strings.Split(value, ",")
				for _, s := range servers {
					s = strings.TrimSpace(s)
					if s != "" {
						opts.WINSServers = append(opts.WINSServers, s)
					}
				}
			}
		}
	}
}

// convertRTXLeaseTimeToGo converts RTX lease time format (h:mm or "infinite") to Go duration
func convertRTXLeaseTimeToGo(rtxTime string) string {
	if rtxTime == "infinite" {
		return "infinite"
	}

	// RTX format: h:mm (e.g., "3:00" for 3 hours, "72:00" for 72 hours)
	parts := strings.Split(rtxTime, ":")
	if len(parts) == 2 {
		hours, err := strconv.Atoi(parts[0])
		if err != nil {
			return rtxTime
		}
		minutes, err := strconv.Atoi(parts[1])
		if err != nil {
			return rtxTime
		}

		totalMinutes := hours*60 + minutes
		if totalMinutes%60 == 0 {
			return fmt.Sprintf("%dh", totalMinutes/60)
		}
		return fmt.Sprintf("%dm", totalMinutes)
	}

	return rtxTime
}

// convertGoLeaseTimeToRTX converts Go duration format to RTX lease time format
func convertGoLeaseTimeToRTX(goDuration string) string {
	if goDuration == "" {
		return ""
	}
	if goDuration == "infinite" {
		return "infinite"
	}

	// Parse Go duration-like format (e.g., "72h", "30m", "1h30m")
	goDuration = strings.ToLower(goDuration)

	totalMinutes := 0

	// Handle hours
	if idx := strings.Index(goDuration, "h"); idx != -1 {
		hours, err := strconv.Atoi(goDuration[:idx])
		if err == nil {
			totalMinutes += hours * 60
		}
		goDuration = goDuration[idx+1:]
	}

	// Handle minutes
	if idx := strings.Index(goDuration, "m"); idx != -1 {
		minutes, err := strconv.Atoi(goDuration[:idx])
		if err == nil {
			totalMinutes += minutes
		}
	}

	if totalMinutes == 0 {
		// Try parsing as plain hours
		hours, err := strconv.Atoi(strings.TrimSuffix(goDuration, "h"))
		if err == nil {
			totalMinutes = hours * 60
		}
	}

	if totalMinutes == 0 {
		return goDuration // Return as-is if parsing failed
	}

	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	return fmt.Sprintf("%d:%02d", hours, minutes)
}

// BuildDHCPScopeCommand builds the command to create a DHCP scope
// Command format: dhcp scope <id> <network>/<prefix> [gateway <ip>] [expire <time>] [maxexpire <time>]
// Or range format: dhcp scope <id> <start_ip>-<end_ip>/<prefix> [gateway <ip>] [expire <time>] [maxexpire <time>]
// Note: Gateway/routers are preferably configured via options command, but legacy format is still supported
func BuildDHCPScopeCommand(scope DHCPScope) string {
	var networkPart string
	if scope.RangeStart != "" && scope.RangeEnd != "" {
		// Use range format: start-end/prefix
		// Extract prefix from Network (e.g., "192.168.0.0/16" -> "16")
		prefix := ""
		if idx := strings.Index(scope.Network, "/"); idx >= 0 {
			prefix = scope.Network[idx+1:]
		}
		networkPart = fmt.Sprintf("%s-%s/%s", scope.RangeStart, scope.RangeEnd, prefix)
	} else {
		// Use CIDR format: network/prefix
		networkPart = scope.Network
	}

	cmd := fmt.Sprintf("dhcp scope %d %s", scope.ScopeID, networkPart)

	// Gateway (legacy format) - only output if Routers is set and we want legacy output
	// Note: Modern RTX prefers "dhcp scope option" for router settings
	// We don't output gateway here to encourage use of dhcp scope option command

	if scope.LeaseTime != "" {
		rtxTime := convertGoLeaseTimeToRTX(scope.LeaseTime)
		if rtxTime != "" {
			cmd += fmt.Sprintf(" expire %s", rtxTime)
		}
	}

	if scope.MaxLeaseTime != "" {
		rtxTime := convertGoLeaseTimeToRTX(scope.MaxLeaseTime)
		if rtxTime != "" {
			cmd += fmt.Sprintf(" maxexpire %s", rtxTime)
		}
	}

	return cmd
}

// BuildDHCPScopeOptionsCommand builds the command to set DHCP options for a scope
// Command format: dhcp scope option <id> [dns=<dns1>,<dns2>] [router=<gw1>,<gw2>] [domain=<domain>] [hostname=<name>] [wins_server=<ip1>,<ip2>]
func BuildDHCPScopeOptionsCommand(scopeID int, opts DHCPScopeOptions) string {
	var parts []string

	// DNS servers (max 3, option 6)
	if len(opts.DNSServers) > 0 {
		servers := opts.DNSServers
		if len(servers) > 3 {
			servers = servers[:3]
		}
		parts = append(parts, fmt.Sprintf("dns=%s", strings.Join(servers, ",")))
	}

	// Routers/default gateways (max 3, option 3)
	if len(opts.Routers) > 0 {
		routers := opts.Routers
		if len(routers) > 3 {
			routers = routers[:3]
		}
		parts = append(parts, fmt.Sprintf("router=%s", strings.Join(routers, ",")))
	}

	// Domain name (option 15)
	if opts.DomainName != "" {
		parts = append(parts, fmt.Sprintf("domain=%s", opts.DomainName))
	}

	// Hostname (option 12)
	if opts.Hostname != "" {
		parts = append(parts, fmt.Sprintf("hostname=%s", opts.Hostname))
	}

	// WINS/NetBIOS name servers (max 3, option 44)
	if len(opts.WINSServers) > 0 {
		servers := opts.WINSServers
		if len(servers) > 3 {
			servers = servers[:3]
		}
		parts = append(parts, fmt.Sprintf("wins_server=%s", strings.Join(servers, ",")))
	}

	if len(parts) == 0 {
		return ""
	}

	return fmt.Sprintf("dhcp scope option %d %s", scopeID, strings.Join(parts, " "))
}

// BuildDHCPScopeExceptCommand builds the command to add an exclusion range
// Command format: dhcp scope <id> except <start>-<end>
func BuildDHCPScopeExceptCommand(scopeID int, excludeRange ExcludeRange) string {
	return fmt.Sprintf("dhcp scope %d except %s-%s", scopeID, excludeRange.Start, excludeRange.End)
}

// BuildDeleteDHCPScopeCommand builds the command to delete a DHCP scope
// Command format: no dhcp scope <id>
func BuildDeleteDHCPScopeCommand(scopeID int) string {
	return fmt.Sprintf("no dhcp scope %d", scopeID)
}

// BuildDeleteDHCPScopeOptionsCommand builds the command to remove all DHCP options
// Command format: no dhcp scope option <id>
func BuildDeleteDHCPScopeOptionsCommand(scopeID int) string {
	return fmt.Sprintf("no dhcp scope option %d", scopeID)
}

// BuildDeleteDHCPScopeExceptCommand builds the command to remove an exclusion range
// Command format: no dhcp scope <id> except <start>-<end>
func BuildDeleteDHCPScopeExceptCommand(scopeID int, excludeRange ExcludeRange) string {
	return fmt.Sprintf("no dhcp scope %d except %s-%s", scopeID, excludeRange.Start, excludeRange.End)
}

// BuildShowDHCPScopeCommand builds the command to show DHCP scope configuration
// This uses a broad grep pattern to capture all dhcp scope lines, then relies on
// the parser to filter by scope ID. Direct grep like "dhcp scope 1" misses
// "dhcp scope option 1" lines because the ID is not immediately after "dhcp scope ".
// Command format: show config | grep "dhcp scope"
func BuildShowDHCPScopeCommand(scopeID int) string {
	// Use broad pattern - the parser will filter by scopeID
	return "show config | grep \"dhcp scope\""
}

// BuildShowAllDHCPScopesCommand builds the command to show all DHCP scopes
// Command format: show config | grep "dhcp scope"
func BuildShowAllDHCPScopesCommand() string {
	return "show config | grep \"dhcp scope\""
}

// ValidateDHCPScope validates a DHCP scope configuration
func ValidateDHCPScope(scope DHCPScope) error {
	if scope.ScopeID <= 0 {
		return fmt.Errorf("scope_id must be positive")
	}

	if scope.Network == "" {
		return fmt.Errorf("network is required")
	}

	// Validate CIDR format
	if !isValidCIDR(scope.Network) {
		return fmt.Errorf("network must be in CIDR notation (e.g., 192.168.1.0/24)")
	}

	// Validate routers (default gateways)
	if len(scope.Options.Routers) > 3 {
		return fmt.Errorf("maximum 3 routers (default gateways) allowed")
	}
	for _, router := range scope.Options.Routers {
		if !isValidIP(router) {
			return fmt.Errorf("invalid router address: %s", router)
		}
	}

	// Validate DNS servers
	if len(scope.Options.DNSServers) > 3 {
		return fmt.Errorf("maximum 3 DNS servers allowed")
	}
	for _, dns := range scope.Options.DNSServers {
		if !isValidIP(dns) {
			return fmt.Errorf("invalid DNS server address: %s", dns)
		}
	}

	// Validate exclude ranges
	for _, r := range scope.ExcludeRanges {
		if !isValidIP(r.Start) {
			return fmt.Errorf("invalid exclude range start address: %s", r.Start)
		}
		if !isValidIP(r.End) {
			return fmt.Errorf("invalid exclude range end address: %s", r.End)
		}
	}

	return nil
}

// isValidCIDR checks if a string is a valid CIDR notation
func isValidCIDR(cidr string) bool {
	parts := strings.Split(cidr, "/")
	if len(parts) != 2 {
		return false
	}

	if !isValidIP(parts[0]) {
		return false
	}

	prefix, err := strconv.Atoi(parts[1])
	if err != nil || prefix < 0 || prefix > 32 {
		return false
	}

	return true
}

// isValidIP checks if a string is a valid IPv4 address
func isValidIP(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil || num < 0 || num > 255 {
			return false
		}
	}

	return true
}

// calculateNetworkAddress calculates the network address from an IP and prefix length
// For example: "192.168.1.20" with prefix "16" returns "192.168.0.0/16"
func calculateNetworkAddress(ip string, prefixLen string) string {
	prefix, err := strconv.Atoi(prefixLen)
	if err != nil || prefix < 0 || prefix > 32 {
		return ip + "/" + prefixLen // Return as-is if invalid
	}

	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return ip + "/" + prefixLen // Return as-is if invalid
	}

	// Parse IP octets
	octets := make([]int, 4)
	for i, part := range parts {
		octet, err := strconv.Atoi(part)
		if err != nil || octet < 0 || octet > 255 {
			return ip + "/" + prefixLen // Return as-is if invalid
		}
		octets[i] = octet
	}

	// Apply network mask
	// Create a 32-bit mask with 'prefix' leading 1s
	var mask uint32 = 0xFFFFFFFF << (32 - prefix)

	// Convert IP to 32-bit integer
	ipInt := uint32(octets[0])<<24 | uint32(octets[1])<<16 | uint32(octets[2])<<8 | uint32(octets[3])

	// Apply mask to get network address
	networkInt := ipInt & mask

	// Convert back to dotted notation
	networkOctets := []int{
		int((networkInt >> 24) & 0xFF),
		int((networkInt >> 16) & 0xFF),
		int((networkInt >> 8) & 0xFF),
		int(networkInt & 0xFF),
	}

	return fmt.Sprintf("%d.%d.%d.%d/%s", networkOctets[0], networkOctets[1], networkOctets[2], networkOctets[3], prefixLen)
}
