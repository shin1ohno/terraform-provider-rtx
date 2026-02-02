package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DHCPInterfaceConfig represents DHCP service configuration on an interface
type DHCPInterfaceConfig struct {
	Interface    string `json:"interface"`     // Interface name (lan1, lan2, etc.)
	ScopeID      int    `json:"scope_id"`      // DHCP scope number
	RelayEnabled bool   `json:"relay_enabled"` // Whether relay is enabled
}

// DHCPInterfaceParser parses DHCP interface configurations
type DHCPInterfaceParser struct{}

// NewDHCPInterfaceParser creates a new DHCP interface parser
func NewDHCPInterfaceParser() *DHCPInterfaceParser {
	return &DHCPInterfaceParser{}
}

// ParseInterfaceDHCPConfig parses the output of "show config | grep ip INTERFACE dhcp"
func (p *DHCPInterfaceParser) ParseInterfaceDHCPConfig(raw string) ([]DHCPInterfaceConfig, error) {
	var configs []DHCPInterfaceConfig
	lines := strings.Split(raw, "\n")

	// Pattern: ip INTERFACE dhcp service server [SCOPE]
	// Pattern: ip INTERFACE dhcp service relay
	servicePattern := regexp.MustCompile(`^\s*ip\s+(\S+)\s+dhcp\s+service\s+(server|relay)(?:\s+(\d+))?\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if matches := servicePattern.FindStringSubmatch(line); len(matches) >= 3 {
			config := DHCPInterfaceConfig{
				Interface:    matches[1],
				RelayEnabled: matches[2] == "relay",
			}

			if len(matches) > 3 && matches[3] != "" {
				config.ScopeID, _ = strconv.Atoi(matches[3])
			}

			configs = append(configs, config)
		}
	}

	return configs, nil
}

// BuildInterfaceDHCPServiceCommand builds the ip interface dhcp service command
func BuildInterfaceDHCPServiceCommand(iface string, serviceType string, scopeID int) string {
	if serviceType == "relay" {
		return fmt.Sprintf("ip %s dhcp service relay", iface)
	}
	if scopeID > 0 {
		return fmt.Sprintf("ip %s dhcp service server %d", iface, scopeID)
	}
	return fmt.Sprintf("ip %s dhcp service server", iface)
}

// BuildDeleteInterfaceDHCPServiceCommand builds the no ip interface dhcp service command
func BuildDeleteInterfaceDHCPServiceCommand(iface string) string {
	return fmt.Sprintf("no ip %s dhcp service", iface)
}

// BuildShowInterfaceDHCPServiceCommand builds the show command for interface DHCP service
func BuildShowInterfaceDHCPServiceCommand(iface string) string {
	return fmt.Sprintf("show config | grep \"ip %s dhcp service\"", iface)
}

// ValidateDHCPInterfaceName validates the interface name for DHCP service
func ValidateDHCPInterfaceName(iface string) error {
	// Valid interfaces: lan1, lan2, ..., vlan1, vlan2, ..., bridge1, etc.
	validPattern := regexp.MustCompile(`^(lan|vlan|bridge|loopback|tunnel|pp)\d+$`)
	if !validPattern.MatchString(iface) {
		return fmt.Errorf("invalid interface name: %s", iface)
	}
	return nil
}
