package parsers

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// IPv6Prefix represents an IPv6 prefix definition on an RTX router
type IPv6Prefix struct {
	ID           int    `json:"id"`                  // Prefix ID (1-255)
	Prefix       string `json:"prefix"`              // Static prefix value (e.g., "2001:db8::")
	PrefixLength int    `json:"prefix_length"`       // Prefix length (e.g., 64)
	Source       string `json:"source"`              // "static", "ra", or "dhcpv6-pd"
	Interface    string `json:"interface,omitempty"` // Source interface for ra/pd
}

// IPv6PrefixParser parses IPv6 prefix configuration output
type IPv6PrefixParser struct{}

// NewIPv6PrefixParser creates a new IPv6 prefix parser
func NewIPv6PrefixParser() *IPv6PrefixParser {
	return &IPv6PrefixParser{}
}

// ParseIPv6PrefixConfig parses the output of "show config | grep ipv6 prefix" command
// and returns a list of IPv6 prefixes
func (p *IPv6PrefixParser) ParseIPv6PrefixConfig(raw string) ([]IPv6Prefix, error) {
	prefixes := make(map[int]*IPv6Prefix)
	lines := strings.Split(raw, "\n")

	// Patterns for different prefix types:
	// Static: ipv6 prefix <id> <prefix>::/<length>
	// RA: ipv6 prefix <id> ra-prefix@<interface>::/<length>
	// DHCPv6-PD: ipv6 prefix <id> dhcp-prefix@<interface>::/<length>
	staticPattern := regexp.MustCompile(`^\s*ipv6\s+prefix\s+(\d+)\s+([0-9a-fA-F:]+)::/(\d+)\s*$`)
	raPattern := regexp.MustCompile(`^\s*ipv6\s+prefix\s+(\d+)\s+ra-prefix@([^:]+)::/(\d+)\s*$`)
	dhcpPDPattern := regexp.MustCompile(`^\s*ipv6\s+prefix\s+(\d+)\s+dhcp-prefix@([^:]+)::/(\d+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try static prefix pattern
		if matches := staticPattern.FindStringSubmatch(line); len(matches) >= 4 {
			prefixID, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			prefixLength, err := strconv.Atoi(matches[3])
			if err != nil {
				continue
			}

			prefixes[prefixID] = &IPv6Prefix{
				ID:           prefixID,
				Prefix:       matches[2] + "::",
				PrefixLength: prefixLength,
				Source:       "static",
				Interface:    "",
			}
			continue
		}

		// Try RA prefix pattern
		if matches := raPattern.FindStringSubmatch(line); len(matches) >= 4 {
			prefixID, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			prefixLength, err := strconv.Atoi(matches[3])
			if err != nil {
				continue
			}

			prefixes[prefixID] = &IPv6Prefix{
				ID:           prefixID,
				Prefix:       "",
				PrefixLength: prefixLength,
				Source:       "ra",
				Interface:    matches[2],
			}
			continue
		}

		// Try DHCPv6-PD prefix pattern
		if matches := dhcpPDPattern.FindStringSubmatch(line); len(matches) >= 4 {
			prefixID, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			prefixLength, err := strconv.Atoi(matches[3])
			if err != nil {
				continue
			}

			prefixes[prefixID] = &IPv6Prefix{
				ID:           prefixID,
				Prefix:       "",
				PrefixLength: prefixLength,
				Source:       "dhcpv6-pd",
				Interface:    matches[2],
			}
			continue
		}
	}

	// Convert map to slice
	result := make([]IPv6Prefix, 0, len(prefixes))
	for _, prefix := range prefixes {
		result = append(result, *prefix)
	}

	return result, nil
}

// ParseSinglePrefix parses configuration for a specific prefix ID
func (p *IPv6PrefixParser) ParseSinglePrefix(raw string, prefixID int) (*IPv6Prefix, error) {
	prefixes, err := p.ParseIPv6PrefixConfig(raw)
	if err != nil {
		return nil, err
	}

	for _, prefix := range prefixes {
		if prefix.ID == prefixID {
			return &prefix, nil
		}
	}

	return nil, fmt.Errorf("prefix %d not found", prefixID)
}

// BuildIPv6PrefixCommand builds the command to create an IPv6 prefix
// Command formats:
// - Static: ipv6 prefix <id> <prefix>::/<length>
// - RA: ipv6 prefix <id> ra-prefix@<interface>::/<length>
// - DHCPv6-PD: ipv6 prefix <id> dhcp-prefix@<interface>::/<length>
func BuildIPv6PrefixCommand(prefix IPv6Prefix) string {
	switch prefix.Source {
	case "ra":
		return fmt.Sprintf("ipv6 prefix %d ra-prefix@%s::/%d", prefix.ID, prefix.Interface, prefix.PrefixLength)
	case "dhcpv6-pd":
		return fmt.Sprintf("ipv6 prefix %d dhcp-prefix@%s::/%d", prefix.ID, prefix.Interface, prefix.PrefixLength)
	default: // static
		// Remove trailing :: if present, then add it back for consistency
		prefixVal := strings.TrimSuffix(prefix.Prefix, "::")
		return fmt.Sprintf("ipv6 prefix %d %s::/%d", prefix.ID, prefixVal, prefix.PrefixLength)
	}
}

// BuildDeleteIPv6PrefixCommand builds the command to delete an IPv6 prefix
// Command format: no ipv6 prefix <id>
func BuildDeleteIPv6PrefixCommand(prefixID int) string {
	return fmt.Sprintf("no ipv6 prefix %d", prefixID)
}

// BuildShowIPv6PrefixCommand builds the command to show a specific IPv6 prefix
// Command format: show config | grep "ipv6 prefix <id>"
func BuildShowIPv6PrefixCommand(prefixID int) string {
	return fmt.Sprintf("show config | grep \"ipv6 prefix %d\"", prefixID)
}

// BuildShowAllIPv6PrefixesCommand builds the command to show all IPv6 prefixes
// Command format: show config | grep "ipv6 prefix"
func BuildShowAllIPv6PrefixesCommand() string {
	return "show config | grep \"ipv6 prefix\""
}

// ValidateIPv6Prefix validates an IPv6 prefix configuration
func ValidateIPv6Prefix(prefix IPv6Prefix) error {
	// Validate prefix ID range (1-255)
	if prefix.ID < 1 || prefix.ID > 255 {
		return fmt.Errorf("prefix ID must be between 1 and 255")
	}

	// Validate prefix length range (1-128)
	if prefix.PrefixLength < 1 || prefix.PrefixLength > 128 {
		return fmt.Errorf("prefix length must be between 1 and 128")
	}

	// Validate source type
	validSources := map[string]bool{
		"static":    true,
		"ra":        true,
		"dhcpv6-pd": true,
	}
	if !validSources[prefix.Source] {
		return fmt.Errorf("source must be one of: static, ra, dhcpv6-pd")
	}

	// Source-specific validation
	switch prefix.Source {
	case "static":
		if prefix.Prefix == "" {
			return fmt.Errorf("prefix is required for static source")
		}
		// Validate IPv6 prefix format
		if !isValidIPv6Prefix(prefix.Prefix) {
			return fmt.Errorf("invalid IPv6 prefix format: %s", prefix.Prefix)
		}
	case "ra":
		if prefix.Interface == "" {
			return fmt.Errorf("interface is required for ra source")
		}
	case "dhcpv6-pd":
		if prefix.Interface == "" {
			return fmt.Errorf("interface is required for dhcpv6-pd source")
		}
	}

	return nil
}

// isValidIPv6Prefix checks if a string is a valid IPv6 prefix (address portion only)
func isValidIPv6Prefix(prefix string) bool {
	// Remove trailing :: if present for parsing
	prefixVal := strings.TrimSuffix(prefix, "::")

	// Try to parse as IPv6 address
	// We need to append :: to make it a valid IPv6 address if it was stripped
	testAddr := prefixVal
	if !strings.Contains(testAddr, "::") {
		testAddr = prefixVal + "::"
	}

	ip := net.ParseIP(testAddr)
	return ip != nil
}
