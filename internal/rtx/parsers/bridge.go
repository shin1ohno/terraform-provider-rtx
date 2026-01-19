package parsers

import (
	"fmt"
	"regexp"
	"strings"
)

// BridgeConfig represents a bridge configuration on an RTX router
type BridgeConfig struct {
	Name    string   `json:"name"`    // Bridge name (bridge1, bridge2, etc.)
	Members []string `json:"members"` // Member interfaces (lan1, tunnel1, etc.)
}

// BridgeParser parses bridge configuration output
type BridgeParser struct{}

// NewBridgeParser creates a new bridge parser
func NewBridgeParser() *BridgeParser {
	return &BridgeParser{}
}

// ParseBridgeConfig parses the output of "show config | grep bridge" command
// and returns a list of bridge configurations
// Example input lines:
//   bridge member bridge1 lan1
//   bridge member bridge1 lan1 tunnel1
//   bridge member bridge2 lan2
func (p *BridgeParser) ParseBridgeConfig(raw string) ([]BridgeConfig, error) {
	bridges := make(map[string]*BridgeConfig)
	lines := strings.Split(raw, "\n")

	// Pattern for bridge member command: bridge member <name> <member1> [<member2>...]
	bridgePattern := regexp.MustCompile(`^\s*bridge\s+member\s+(\S+)\s+(.+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try bridge member pattern
		if matches := bridgePattern.FindStringSubmatch(line); len(matches) >= 3 {
			name := matches[1]
			membersStr := strings.TrimSpace(matches[2])
			members := strings.Fields(membersStr)

			bridge, exists := bridges[name]
			if !exists {
				bridge = &BridgeConfig{
					Name:    name,
					Members: []string{},
				}
				bridges[name] = bridge
			}
			bridge.Members = members
		}
	}

	// Convert map to slice
	result := make([]BridgeConfig, 0, len(bridges))
	for _, bridge := range bridges {
		result = append(result, *bridge)
	}

	return result, nil
}

// ParseSingleBridge parses configuration for a specific bridge
func (p *BridgeParser) ParseSingleBridge(raw string, name string) (*BridgeConfig, error) {
	bridges, err := p.ParseBridgeConfig(raw)
	if err != nil {
		return nil, err
	}

	for _, bridge := range bridges {
		if bridge.Name == name {
			return &bridge, nil
		}
	}

	return nil, fmt.Errorf("bridge %s not found", name)
}

// BuildBridgeMemberCommand builds the command to create/update a bridge
// Command format: bridge member <name> <member1> [<member2>...]
func BuildBridgeMemberCommand(name string, members []string) string {
	if len(members) == 0 {
		return fmt.Sprintf("bridge member %s", name)
	}
	return fmt.Sprintf("bridge member %s %s", name, strings.Join(members, " "))
}

// BuildDeleteBridgeCommand builds the command to delete a bridge
// Command format: no bridge member <name>
func BuildDeleteBridgeCommand(name string) string {
	return fmt.Sprintf("no bridge member %s", name)
}

// BuildShowBridgeCommand builds the command to show bridge configuration
// Command format: show config | grep bridge
func BuildShowBridgeCommand(name string) string {
	return fmt.Sprintf("show config | grep \"bridge member %s\"", name)
}

// BuildShowAllBridgesCommand builds the command to show all bridge configurations
// Command format: show config | grep "bridge member"
func BuildShowAllBridgesCommand() string {
	return "show config | grep \"bridge member\""
}

// ValidateBridgeName validates the bridge name format
// Bridge names must be in format "bridgeN" (e.g., bridge1, bridge2)
func ValidateBridgeName(name string) error {
	if name == "" {
		return fmt.Errorf("bridge name is required")
	}

	// Validate bridge name format (bridge[0-9]+)
	validNamePattern := regexp.MustCompile(`^bridge\d+$`)
	if !validNamePattern.MatchString(name) {
		return fmt.Errorf("bridge name must be in format 'bridgeN' (e.g., bridge1, bridge2), got %s", name)
	}

	return nil
}

// ValidateBridgeMember validates a bridge member interface name
// Valid formats: lan*, tunnel*, pp*
func ValidateBridgeMember(member string) error {
	if member == "" {
		return fmt.Errorf("member interface name is required")
	}

	// Valid member patterns
	validPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^lan\d+$`),           // lan1, lan2, etc.
		regexp.MustCompile(`^lan\d+/\d+$`),       // lan1/1 (VLAN interfaces)
		regexp.MustCompile(`^tunnel\d+$`),        // tunnel1, tunnel2, etc.
		regexp.MustCompile(`^pp\d+$`),            // pp1, pp2, etc.
		regexp.MustCompile(`^loopback\d+$`),      // loopback1, etc.
		regexp.MustCompile(`^bridge\d+$`),        // nested bridge (rare but possible)
	}

	for _, pattern := range validPatterns {
		if pattern.MatchString(member) {
			return nil
		}
	}

	return fmt.Errorf("invalid member interface name %q, must be lan*, lan*/*, tunnel*, pp*, loopback*, or bridge*", member)
}

// ValidateBridge validates a complete bridge configuration
func ValidateBridge(bridge BridgeConfig) error {
	if err := ValidateBridgeName(bridge.Name); err != nil {
		return err
	}

	// Validate each member
	for _, member := range bridge.Members {
		if err := ValidateBridgeMember(member); err != nil {
			return fmt.Errorf("invalid bridge member: %w", err)
		}
	}

	// Check for duplicate members
	seen := make(map[string]bool)
	for _, member := range bridge.Members {
		if seen[member] {
			return fmt.Errorf("duplicate member interface: %s", member)
		}
		seen[member] = true
	}

	return nil
}
