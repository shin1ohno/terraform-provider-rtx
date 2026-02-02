package parsers

import (
	"fmt"
	"regexp"
	"strings"
)

// DHCPServiceConfig represents DHCP service configuration
type DHCPServiceConfig struct {
	ServiceType string `json:"service_type"` // "server", "relay", or "" (disabled)
}

// DHCPServiceParser parses DHCP service configuration
type DHCPServiceParser struct{}

// NewDHCPServiceParser creates a new DHCP service parser
func NewDHCPServiceParser() *DHCPServiceParser {
	return &DHCPServiceParser{}
}

// ParseServiceConfig parses the output of "show config | grep dhcp service" command
func (p *DHCPServiceParser) ParseServiceConfig(raw string) (*DHCPServiceConfig, error) {
	config := &DHCPServiceConfig{}
	lines := strings.Split(raw, "\n")

	// Pattern: dhcp service server|relay
	servicePattern := regexp.MustCompile(`^\s*dhcp\s+service\s+(server|relay)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if matches := servicePattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.ServiceType = matches[1]
			return config, nil
		}
	}

	return config, nil
}

// BuildDHCPServiceCommand builds the dhcp service command
func BuildDHCPServiceCommand(serviceType string) string {
	if serviceType == "" {
		return ""
	}
	return fmt.Sprintf("dhcp service %s", serviceType)
}

// BuildDeleteDHCPServiceCommand builds the no dhcp service command
func BuildDeleteDHCPServiceCommand() string {
	return "no dhcp service"
}

// BuildShowDHCPServiceCommand builds the show command for DHCP service
func BuildShowDHCPServiceCommand() string {
	return "show config | grep \"dhcp service\""
}

// ValidateDHCPServiceType validates the DHCP service type
func ValidateDHCPServiceType(serviceType string) error {
	switch serviceType {
	case "server", "relay", "":
		return nil
	default:
		return fmt.Errorf("invalid DHCP service type: %s (must be 'server' or 'relay')", serviceType)
	}
}
