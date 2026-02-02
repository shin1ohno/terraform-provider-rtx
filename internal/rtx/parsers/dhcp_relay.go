package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DHCPRelayServerConfig represents DHCP relay server configuration
type DHCPRelayServerConfig struct {
	Servers []string `json:"servers"` // DHCP server IP addresses (max 4)
}

// DHCPRelaySelectConfig represents DHCP relay select configuration
type DHCPRelaySelectConfig struct {
	ScopeID int    `json:"scope_id"`
	Server  string `json:"server"`
}

// DHCPRelayParser parses DHCP relay configurations
type DHCPRelayParser struct{}

// NewDHCPRelayParser creates a new DHCP relay parser
func NewDHCPRelayParser() *DHCPRelayParser {
	return &DHCPRelayParser{}
}

// ParseRelayServerConfig parses the output of "show config | grep dhcp relay server"
func (p *DHCPRelayParser) ParseRelayServerConfig(raw string) (*DHCPRelayServerConfig, error) {
	config := &DHCPRelayServerConfig{
		Servers: []string{},
	}
	lines := strings.Split(raw, "\n")

	// Pattern: dhcp relay server IP1 [IP2 [IP3 [IP4]]]
	serverPattern := regexp.MustCompile(`^\s*dhcp\s+relay\s+server\s+(.+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if matches := serverPattern.FindStringSubmatch(line); len(matches) >= 2 {
			servers := strings.Fields(matches[1])
			for _, s := range servers {
				if s != "" {
					config.Servers = append(config.Servers, s)
				}
			}
			return config, nil
		}
	}

	return config, nil
}

// ParseRelaySelectConfig parses the output of "show config | grep dhcp relay select"
func (p *DHCPRelayParser) ParseRelaySelectConfig(raw string) ([]DHCPRelaySelectConfig, error) {
	var configs []DHCPRelaySelectConfig
	lines := strings.Split(raw, "\n")

	// Pattern: dhcp relay select SCOPE SERVER
	selectPattern := regexp.MustCompile(`^\s*dhcp\s+relay\s+select\s+(\d+)\s+([0-9.]+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if matches := selectPattern.FindStringSubmatch(line); len(matches) >= 3 {
			scopeID, _ := strconv.Atoi(matches[1])
			configs = append(configs, DHCPRelaySelectConfig{
				ScopeID: scopeID,
				Server:  matches[2],
			})
		}
	}

	return configs, nil
}

// BuildDHCPRelayServerCommand builds the dhcp relay server command
func BuildDHCPRelayServerCommand(servers []string) string {
	if len(servers) == 0 {
		return ""
	}
	// Max 4 servers
	if len(servers) > 4 {
		servers = servers[:4]
	}
	return fmt.Sprintf("dhcp relay server %s", strings.Join(servers, " "))
}

// BuildDeleteDHCPRelayServerCommand builds the no dhcp relay server command
func BuildDeleteDHCPRelayServerCommand() string {
	return "no dhcp relay server"
}

// BuildDHCPRelaySelectCommand builds the dhcp relay select command
func BuildDHCPRelaySelectCommand(scopeID int, server string) string {
	return fmt.Sprintf("dhcp relay select %d %s", scopeID, server)
}

// BuildDeleteDHCPRelaySelectCommand builds the no dhcp relay select command
func BuildDeleteDHCPRelaySelectCommand(scopeID int) string {
	return fmt.Sprintf("no dhcp relay select %d", scopeID)
}

// BuildShowDHCPRelayServerCommand builds the show command for DHCP relay server
func BuildShowDHCPRelayServerCommand() string {
	return "show config | grep \"dhcp relay server\""
}

// BuildShowDHCPRelaySelectCommand builds the show command for DHCP relay select
func BuildShowDHCPRelaySelectCommand() string {
	return "show config | grep \"dhcp relay select\""
}

// ValidateDHCPRelayServerConfig validates the relay server configuration
func ValidateDHCPRelayServerConfig(config DHCPRelayServerConfig) error {
	if len(config.Servers) > 4 {
		return fmt.Errorf("maximum 4 DHCP relay servers allowed, got %d", len(config.Servers))
	}
	for _, server := range config.Servers {
		if !isValidIP(server) {
			return fmt.Errorf("invalid DHCP relay server address: %s", server)
		}
	}
	return nil
}
