package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DHCPScope represents a DHCP scope configuration on an RTX router
type DHCPScope struct {
	ScopeID       int            `json:"scope_id"`
	Network       string         `json:"network"`                  // CIDR notation: "192.168.1.0/24"
	Gateway       string         `json:"gateway,omitempty"`        // Default gateway
	DNSServers    []string       `json:"dns_servers,omitempty"`    // Up to 3 DNS servers
	LeaseTime     string         `json:"lease_time,omitempty"`     // Go duration format or "infinite"
	ExcludeRanges []ExcludeRange `json:"exclude_ranges,omitempty"` // Excluded IP ranges
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
	// dhcp scope <id> <network>/<prefix> [gateway <gateway>] [expire <time>]
	scopePattern := regexp.MustCompile(`^\s*dhcp\s+scope\s+(\d+)\s+([0-9.]+/\d+)(?:\s+gateway\s+([0-9.]+))?(?:\s+expire\s+(\S+))?\s*$`)
	// dhcp scope option <id> dns=<dns1>[,<dns2>[,<dns3>]]
	dnsPattern := regexp.MustCompile(`^\s*dhcp\s+scope\s+option\s+(\d+)\s+dns=([0-9.,]+)\s*$`)
	// dhcp scope <id> except <start>-<end>
	exceptPattern := regexp.MustCompile(`^\s*dhcp\s+scope\s+(\d+)\s+except\s+([0-9.]+)-([0-9.]+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try scope definition pattern
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
			if len(matches) > 3 && matches[3] != "" {
				scope.Gateway = matches[3]
			}
			if len(matches) > 4 && matches[4] != "" {
				scope.LeaseTime = convertRTXLeaseTimeToGo(matches[4])
			}
			continue
		}

		// Try DNS option pattern
		if matches := dnsPattern.FindStringSubmatch(line); len(matches) >= 3 {
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

			dnsServers := strings.Split(matches[2], ",")
			for _, dns := range dnsServers {
				dns = strings.TrimSpace(dns)
				if dns != "" {
					scope.DNSServers = append(scope.DNSServers, dns)
				}
			}
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
// Command format: dhcp scope <id> <network>/<prefix> [gateway <gateway>] [expire <time>]
func BuildDHCPScopeCommand(scope DHCPScope) string {
	cmd := fmt.Sprintf("dhcp scope %d %s", scope.ScopeID, scope.Network)

	if scope.Gateway != "" {
		cmd += fmt.Sprintf(" gateway %s", scope.Gateway)
	}

	if scope.LeaseTime != "" {
		rtxTime := convertGoLeaseTimeToRTX(scope.LeaseTime)
		if rtxTime != "" {
			cmd += fmt.Sprintf(" expire %s", rtxTime)
		}
	}

	return cmd
}

// BuildDHCPScopeOptionsCommand builds the command to set DNS servers for a scope
// Command format: dhcp scope option <id> dns=<dns1>[,<dns2>[,<dns3>]]
func BuildDHCPScopeOptionsCommand(scopeID int, dnsServers []string) string {
	if len(dnsServers) == 0 {
		return ""
	}

	// Limit to 3 DNS servers (RTX limitation)
	servers := dnsServers
	if len(servers) > 3 {
		servers = servers[:3]
	}

	return fmt.Sprintf("dhcp scope option %d dns=%s", scopeID, strings.Join(servers, ","))
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

// BuildDeleteDHCPScopeOptionsCommand builds the command to remove DNS options
// Command format: no dhcp scope option <id> dns
func BuildDeleteDHCPScopeOptionsCommand(scopeID int) string {
	return fmt.Sprintf("no dhcp scope option %d dns", scopeID)
}

// BuildDeleteDHCPScopeExceptCommand builds the command to remove an exclusion range
// Command format: no dhcp scope <id> except <start>-<end>
func BuildDeleteDHCPScopeExceptCommand(scopeID int, excludeRange ExcludeRange) string {
	return fmt.Sprintf("no dhcp scope %d except %s-%s", scopeID, excludeRange.Start, excludeRange.End)
}

// BuildShowDHCPScopeCommand builds the command to show DHCP scope configuration
// Command format: show config | grep "dhcp scope <id>"
func BuildShowDHCPScopeCommand(scopeID int) string {
	return fmt.Sprintf("show config | grep \"dhcp scope %d\"", scopeID)
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

	// Validate gateway if provided
	if scope.Gateway != "" && !isValidIP(scope.Gateway) {
		return fmt.Errorf("gateway must be a valid IP address")
	}

	// Validate DNS servers
	if len(scope.DNSServers) > 3 {
		return fmt.Errorf("maximum 3 DNS servers allowed")
	}
	for _, dns := range scope.DNSServers {
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
