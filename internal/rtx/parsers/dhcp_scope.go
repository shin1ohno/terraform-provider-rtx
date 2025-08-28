package parsers

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// DhcpScope represents a DHCP scope configuration on an RTX router
type DhcpScope struct {
	ID         int      `json:"id"`
	RangeStart string   `json:"range_start"`
	RangeEnd   string   `json:"range_end"`
	Prefix     int      `json:"prefix"`
	Gateway    string   `json:"gateway,omitempty"`
	DNSServers []string `json:"dns_servers,omitempty"`
	Lease      int      `json:"lease,omitempty"`
	DomainName string   `json:"domain_name,omitempty"`
}

// DhcpScopeParser is the interface for parsing DHCP scope information
type DhcpScopeParser interface {
	Parser
	ParseDhcpScopes(raw string) ([]*DhcpScope, error)
}

// BaseDhcpScopeParser provides common functionality for DHCP scope parsers
type BaseDhcpScopeParser struct {
	modelPatterns map[string]*regexp.Regexp
}

// rtx830DhcpScopeParser handles RTX830 DHCP scope output
type rtx830DhcpScopeParser struct {
	BaseDhcpScopeParser
}

// rtx12xxDhcpScopeParser handles RTX1210/1220 DHCP scope output
type rtx12xxDhcpScopeParser struct {
	BaseDhcpScopeParser
}

func init() {
	// Register RTX830 parser
	Register("dhcp_scope", "RTX830", &rtx830DhcpScopeParser{
		BaseDhcpScopeParser: BaseDhcpScopeParser{
			modelPatterns: map[string]*regexp.Regexp{
				"scope": regexp.MustCompile(`^dhcp\s+scope\s+(\d+)\s+(\S+)\s*(.*)$`),
			},
		},
	})

	// Register RTX12xx parser
	rtx12xxParser := &rtx12xxDhcpScopeParser{
		BaseDhcpScopeParser: BaseDhcpScopeParser{
			modelPatterns: map[string]*regexp.Regexp{
				"scope": regexp.MustCompile(`^dhcp\s+scope\s+(\d+)\s+(\S+)\s*(.*)$`),
			},
		},
	}
	Register("dhcp_scope", "RTX1210", rtx12xxParser)
	Register("dhcp_scope", "RTX1220", rtx12xxParser)

	// Create aliases for model families
	if err := RegisterAlias("dhcp_scope", "RTX1210", "RTX12xx"); err != nil {
		log.Printf("[WARN] Failed to register alias for dhcp_scope RTX12xx: %v", err)
	}
}

// ParseDhcpScope parses a single DHCP scope configuration line
func ParseDhcpScope(configLine string) (*DhcpScope, error) {
	configLine = strings.TrimSpace(configLine)
	if configLine == "" {
		return nil, fmt.Errorf("empty configuration line")
	}

	// Basic pattern to match "dhcp scope ID RANGE/PREFIX [options]"
	pattern := regexp.MustCompile(`^dhcp\s+scope\s+(\d+)\s+(\S+)\s*(.*)$`)
	matches := pattern.FindStringSubmatch(configLine)
	if len(matches) < 3 {
		return nil, fmt.Errorf("invalid dhcp scope format")
	}

	// Parse scope ID
	id, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid scope ID: %v", err)
	}
	if id <= 0 {
		return nil, fmt.Errorf("scope ID must be positive")
	}

	// Parse IP range and prefix
	scope := &DhcpScope{ID: id}
	if err := parseIPRangeAndPrefix(scope, matches[2]); err != nil {
		return nil, err
	}

	// Parse options if present
	if len(matches) > 2 && strings.TrimSpace(matches[3]) != "" {
		if err := parseOptions(scope, matches[3]); err != nil {
			return nil, err
		}
	}

	return scope, nil
}

// parseIPRangeAndPrefix parses the "START-END/PREFIX" part
func parseIPRangeAndPrefix(scope *DhcpScope, rangeStr string) error {
	// Split by '/' to separate range from prefix
	parts := strings.Split(rangeStr, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format: expected RANGE/PREFIX")
	}

	// Parse prefix
	prefix, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid prefix: %v", err)
	}
	if prefix < 0 || prefix > 32 {
		return fmt.Errorf("prefix must be between 0 and 32")
	}
	scope.Prefix = prefix

	// Parse IP range
	rangeParts := strings.Split(parts[0], "-")
	if len(rangeParts) != 2 {
		return fmt.Errorf("invalid IP range format: expected START-END")
	}

	startIP := strings.TrimSpace(rangeParts[0])
	endIP := strings.TrimSpace(rangeParts[1])

	// Validate IP addresses
	if net.ParseIP(startIP) == nil {
		return fmt.Errorf("invalid start IP address: %s", startIP)
	}
	if net.ParseIP(endIP) == nil {
		return fmt.Errorf("invalid end IP address: %s", endIP)
	}

	scope.RangeStart = startIP
	scope.RangeEnd = endIP

	return nil
}

// parseOptions parses the options part of the DHCP scope line
func parseOptions(scope *DhcpScope, optionsStr string) error {
	tokens := strings.Fields(optionsStr)

	for i := 0; i < len(tokens); i++ {
		switch tokens[i] {
		case "gateway":
			if i+1 >= len(tokens) {
				return fmt.Errorf("gateway option requires an IP address")
			}
			i++
			gateway := tokens[i]
			if net.ParseIP(gateway) == nil {
				return fmt.Errorf("invalid gateway IP address: %s", gateway)
			}
			scope.Gateway = gateway

		case "dns":
			if i+1 >= len(tokens) {
				return fmt.Errorf("dns option requires at least one IP address")
			}

			// Collect DNS servers until we hit the next option or end
			var dnsServers []string
			for j := i + 1; j < len(tokens); j++ {
				if isOptionKeyword(tokens[j]) {
					break
				}
				if net.ParseIP(tokens[j]) == nil {
					return fmt.Errorf("invalid DNS server IP address: %s", tokens[j])
				}
				dnsServers = append(dnsServers, tokens[j])
				i = j
			}
			scope.DNSServers = dnsServers

		case "lease":
			if i+1 >= len(tokens) {
				return fmt.Errorf("lease option requires a number")
			}
			i++
			lease, err := strconv.Atoi(tokens[i])
			if err != nil {
				return fmt.Errorf("invalid lease value: %v", err)
			}
			if lease < 0 {
				return fmt.Errorf("lease must be non-negative")
			}
			scope.Lease = lease

		case "domain":
			if i+1 >= len(tokens) {
				return fmt.Errorf("domain option requires a domain name")
			}
			i++
			scope.DomainName = tokens[i]

		case "expire":
			if i+1 >= len(tokens) {
				return fmt.Errorf("expire option requires a time value")
			}
			i++
			leaseTime, err := parseTimeToSeconds(tokens[i])
			if err != nil {
				return fmt.Errorf("invalid expire value: %v", err)
			}
			scope.Lease = leaseTime

		case "maxexpire":
			if i+1 >= len(tokens) {
				return fmt.Errorf("maxexpire option requires a time value")
			}
			i++
			// For now, we'll just skip maxexpire as our scope struct doesn't store it separately
			// but we need to parse it to avoid "unknown option" error
			_, err := parseTimeToSeconds(tokens[i])
			if err != nil {
				return fmt.Errorf("invalid maxexpire value: %v", err)
			}
			// Skip storing maxexpire for now

		case "ma":
			// Skip "ma" option which is sometimes appended after expire
			// This appears to be related to manual/max settings on RTX routers
			// We'll ignore it for now as it doesn't affect our scope configuration

		default:
			// Log unknown options but don't fail parsing to handle RTX router variations
			// This allows the parser to be more resilient to firmware differences
			// Unknown options are silently skipped
			continue
		}
	}

	return nil
}

// parseTimeToSeconds converts time formats to seconds
// Supports both minutes (e.g., "1440") and hh:mm format (e.g., "24:00")
func parseTimeToSeconds(timeStr string) (int, error) {
	timeStr = strings.TrimSpace(timeStr)

	// Check if it's in hh:mm format
	if strings.Contains(timeStr, ":") {
		parts := strings.Split(timeStr, ":")
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid time format: expected hh:mm")
		}

		hours, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid hours: %v", err)
		}

		minutes, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %v", err)
		}

		if hours < 0 || minutes < 0 || minutes >= 60 {
			return 0, fmt.Errorf("invalid time values")
		}

		return (hours*60 + minutes) * 60, nil // Convert to seconds
	}

	// Otherwise, treat as minutes
	minutes, err := strconv.Atoi(timeStr)
	if err != nil {
		return 0, fmt.Errorf("invalid minutes value: %v", err)
	}

	if minutes < 0 {
		return 0, fmt.Errorf("minutes must be non-negative")
	}

	return minutes * 60, nil // Convert to seconds
}

// isOptionKeyword checks if a token is a known option keyword
func isOptionKeyword(token string) bool {
	keywords := []string{"gateway", "dns", "lease", "domain", "expire", "maxexpire", "ma"}
	for _, keyword := range keywords {
		if token == keyword {
			return true
		}
	}
	return false
}

// Parse implements the Parser interface
func (p *rtx830DhcpScopeParser) Parse(raw string) (interface{}, error) {
	return p.ParseDhcpScopes(raw)
}

// CanHandle implements the Parser interface
func (p *rtx830DhcpScopeParser) CanHandle(model string) bool {
	return model == "RTX830"
}

// ParseDhcpScopes parses RTX830 DHCP scope output
func (p *rtx830DhcpScopeParser) ParseDhcpScopes(raw string) ([]*DhcpScope, error) {
	scopes := make([]*DhcpScope, 0)
	lines := strings.Split(raw, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line matches dhcp scope pattern (but not dhcp scope bind)
		if strings.HasPrefix(line, "dhcp scope ") && !strings.HasPrefix(line, "dhcp scope bind") && !strings.HasPrefix(line, "dhcp scope option") {
			scope, err := ParseDhcpScope(line)
			if err != nil {
				return nil, fmt.Errorf("failed to parse line '%s': %v", line, err)
			}
			scopes = append(scopes, scope)
		}
	}

	return scopes, nil
}

// Parse implements the Parser interface
func (p *rtx12xxDhcpScopeParser) Parse(raw string) (interface{}, error) {
	return p.ParseDhcpScopes(raw)
}

// CanHandle implements the Parser interface
func (p *rtx12xxDhcpScopeParser) CanHandle(model string) bool {
	return strings.HasPrefix(model, "RTX12")
}

// ParseDhcpScopes parses RTX1210/1220 DHCP scope output
func (p *rtx12xxDhcpScopeParser) ParseDhcpScopes(raw string) ([]*DhcpScope, error) {
	scopes := make([]*DhcpScope, 0)
	lines := strings.Split(raw, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line matches dhcp scope pattern (but not dhcp scope bind)
		if strings.HasPrefix(line, "dhcp scope ") && !strings.HasPrefix(line, "dhcp scope bind") && !strings.HasPrefix(line, "dhcp scope option") {
			scope, err := ParseDhcpScope(line)
			if err != nil {
				return nil, fmt.Errorf("failed to parse line '%s': %v", line, err)
			}
			scopes = append(scopes, scope)
		}
	}

	return scopes, nil
}
