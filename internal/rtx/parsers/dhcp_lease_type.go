package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DHCPLeaseTypeConfig represents DHCP scope lease type configuration
type DHCPLeaseTypeConfig struct {
	ScopeID   int    `json:"scope_id"`
	LeaseType string `json:"lease_type"` // "bind-only", "bind-priority", "lease-only"
}

// DHCPLeaseTypeParser parses DHCP lease type configurations
type DHCPLeaseTypeParser struct{}

// NewDHCPLeaseTypeParser creates a new DHCP lease type parser
func NewDHCPLeaseTypeParser() *DHCPLeaseTypeParser {
	return &DHCPLeaseTypeParser{}
}

// ParseLeaseTypeConfig parses the output of "show config | grep dhcp scope lease type"
func (p *DHCPLeaseTypeParser) ParseLeaseTypeConfig(raw string) ([]DHCPLeaseTypeConfig, error) {
	var configs []DHCPLeaseTypeConfig
	lines := strings.Split(raw, "\n")

	// Pattern: dhcp scope lease type SCOPE TYPE
	leaseTypePattern := regexp.MustCompile(`^\s*dhcp\s+scope\s+lease\s+type\s+(\d+)\s+(bind-only|bind-priority|lease-only)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if matches := leaseTypePattern.FindStringSubmatch(line); len(matches) >= 3 {
			scopeID, _ := strconv.Atoi(matches[1])
			configs = append(configs, DHCPLeaseTypeConfig{
				ScopeID:   scopeID,
				LeaseType: matches[2],
			})
		}
	}

	return configs, nil
}

// BuildDHCPLeaseTypeCommand builds the dhcp scope lease type command
func BuildDHCPLeaseTypeCommand(scopeID int, leaseType string) string {
	return fmt.Sprintf("dhcp scope lease type %d %s", scopeID, leaseType)
}

// BuildDeleteDHCPLeaseTypeCommand builds the no dhcp scope lease type command
func BuildDeleteDHCPLeaseTypeCommand(scopeID int) string {
	return fmt.Sprintf("no dhcp scope lease type %d", scopeID)
}

// BuildShowDHCPLeaseTypeCommand builds the show command for DHCP lease type
func BuildShowDHCPLeaseTypeCommand() string {
	return "show config | grep \"dhcp scope lease type\""
}

// ValidateDHCPLeaseType validates the DHCP lease type
func ValidateDHCPLeaseType(leaseType string) error {
	switch leaseType {
	case "bind-only", "bind-priority", "lease-only":
		return nil
	default:
		return fmt.Errorf("invalid DHCP lease type: %s (must be 'bind-only', 'bind-priority', or 'lease-only')", leaseType)
	}
}
