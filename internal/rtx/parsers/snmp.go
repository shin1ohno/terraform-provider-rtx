package parsers

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

// SNMPConfig represents SNMP configuration on an RTX router
type SNMPConfig struct {
	SysName     string          `json:"sysname,omitempty"`     // System name
	SysLocation string          `json:"syslocation,omitempty"` // System location
	SysContact  string          `json:"syscontact,omitempty"`  // System contact
	Communities []SNMPCommunity `json:"communities,omitempty"` // SNMP communities
	Hosts       []SNMPHost      `json:"hosts,omitempty"`       // SNMP trap hosts
	TrapEnable  []string        `json:"trap_enable,omitempty"` // Enabled trap types
}

// SNMPCommunity represents an SNMP community configuration
type SNMPCommunity struct {
	Name       string `json:"name"`                 // Community string name
	Permission string `json:"permission"`           // "ro" (read-only) or "rw" (read-write)
	ACL        string `json:"acl,omitempty"`        // Access control list (optional)
}

// SNMPHost represents an SNMP trap host configuration
type SNMPHost struct {
	Address   string `json:"address"`             // IP address of trap receiver
	Community string `json:"community,omitempty"` // Community string for traps
	Version   string `json:"version,omitempty"`   // SNMP version (1, 2c)
}

// SNMPParser parses SNMP configuration output
type SNMPParser struct{}

// NewSNMPParser creates a new SNMP parser
func NewSNMPParser() *SNMPParser {
	return &SNMPParser{}
}

// ParseSNMPConfig parses the output of "show config | grep snmp" command
func (p *SNMPParser) ParseSNMPConfig(raw string) (*SNMPConfig, error) {
	config := &SNMPConfig{
		Communities: []SNMPCommunity{},
		Hosts:       []SNMPHost{},
		TrapEnable:  []string{},
	}

	lines := strings.Split(raw, "\n")

	// Patterns for SNMP configuration
	// snmp sysname <name>
	sysNamePattern := regexp.MustCompile(`^\s*snmp\s+sysname\s+(.+?)\s*$`)
	// snmp syslocation <location>
	sysLocationPattern := regexp.MustCompile(`^\s*snmp\s+syslocation\s+(.+?)\s*$`)
	// snmp syscontact <contact>
	sysContactPattern := regexp.MustCompile(`^\s*snmp\s+syscontact\s+(.+?)\s*$`)
	// snmp community read-only <string> [<acl>]
	communityROPattern := regexp.MustCompile(`^\s*snmp\s+community\s+read-only\s+(\S+)(?:\s+(\S+))?\s*$`)
	// snmp community read-write <string> [<acl>]
	communityRWPattern := regexp.MustCompile(`^\s*snmp\s+community\s+read-write\s+(\S+)(?:\s+(\S+))?\s*$`)
	// snmp host <ip> [community <string>] [version <ver>]
	hostPattern := regexp.MustCompile(`^\s*snmp\s+host\s+([0-9.]+)(?:\s+community\s+(\S+))?(?:\s+version\s+(\S+))?\s*$`)
	// snmp host <ip> (simple form)
	hostSimplePattern := regexp.MustCompile(`^\s*snmp\s+host\s+([0-9.]+)\s*$`)
	// snmp trap community <string>
	trapCommunityPattern := regexp.MustCompile(`^\s*snmp\s+trap\s+community\s+(\S+)\s*$`)
	// snmp trap enable snmp <types>
	trapEnablePattern := regexp.MustCompile(`^\s*snmp\s+trap\s+enable\s+snmp\s+(.+?)\s*$`)

	var trapCommunity string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// snmp sysname
		if matches := sysNamePattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.SysName = strings.TrimSpace(matches[1])
			continue
		}

		// snmp syslocation
		if matches := sysLocationPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.SysLocation = strings.TrimSpace(matches[1])
			continue
		}

		// snmp syscontact
		if matches := sysContactPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.SysContact = strings.TrimSpace(matches[1])
			continue
		}

		// snmp community read-only
		if matches := communityROPattern.FindStringSubmatch(line); len(matches) >= 2 {
			community := SNMPCommunity{
				Name:       matches[1],
				Permission: "ro",
			}
			if len(matches) >= 3 && matches[2] != "" {
				community.ACL = matches[2]
			}
			config.Communities = append(config.Communities, community)
			continue
		}

		// snmp community read-write
		if matches := communityRWPattern.FindStringSubmatch(line); len(matches) >= 2 {
			community := SNMPCommunity{
				Name:       matches[1],
				Permission: "rw",
			}
			if len(matches) >= 3 && matches[2] != "" {
				community.ACL = matches[2]
			}
			config.Communities = append(config.Communities, community)
			continue
		}

		// snmp trap community (save for later use with hosts)
		if matches := trapCommunityPattern.FindStringSubmatch(line); len(matches) >= 2 {
			trapCommunity = matches[1]
			continue
		}

		// snmp host (full form)
		if matches := hostPattern.FindStringSubmatch(line); len(matches) >= 2 {
			host := SNMPHost{
				Address: matches[1],
			}
			if len(matches) >= 3 && matches[2] != "" {
				host.Community = matches[2]
			}
			if len(matches) >= 4 && matches[3] != "" {
				host.Version = matches[3]
			}
			// If community not specified in host line, use trap community
			if host.Community == "" && trapCommunity != "" {
				host.Community = trapCommunity
			}
			config.Hosts = append(config.Hosts, host)
			continue
		}

		// snmp host (simple form)
		if matches := hostSimplePattern.FindStringSubmatch(line); len(matches) >= 2 {
			host := SNMPHost{
				Address:   matches[1],
				Community: trapCommunity,
			}
			config.Hosts = append(config.Hosts, host)
			continue
		}

		// snmp trap enable snmp
		if matches := trapEnablePattern.FindStringSubmatch(line); len(matches) >= 2 {
			// Parse the trap types (space-separated)
			trapTypes := strings.Fields(matches[1])
			config.TrapEnable = append(config.TrapEnable, trapTypes...)
			continue
		}
	}

	return config, nil
}

// BuildSNMPSysNameCommand builds the command to set the system name
// Command format: snmp sysname <name>
func BuildSNMPSysNameCommand(name string) string {
	return fmt.Sprintf("snmp sysname %s", name)
}

// BuildSNMPSysLocationCommand builds the command to set the system location
// Command format: snmp syslocation <location>
func BuildSNMPSysLocationCommand(location string) string {
	return fmt.Sprintf("snmp syslocation %s", location)
}

// BuildSNMPSysContactCommand builds the command to set the system contact
// Command format: snmp syscontact <contact>
func BuildSNMPSysContactCommand(contact string) string {
	return fmt.Sprintf("snmp syscontact %s", contact)
}

// BuildSNMPCommunityCommand builds the command to configure an SNMP community
// Command format: snmp community read-only|read-write <string> [<acl>]
func BuildSNMPCommunityCommand(community SNMPCommunity) string {
	permission := "read-only"
	if community.Permission == "rw" {
		permission = "read-write"
	}
	if community.ACL != "" {
		return fmt.Sprintf("snmp community %s %s %s", permission, community.Name, community.ACL)
	}
	return fmt.Sprintf("snmp community %s %s", permission, community.Name)
}

// BuildSNMPHostCommand builds the command to configure an SNMP trap host
// Command format: snmp host <ip>
func BuildSNMPHostCommand(host SNMPHost) string {
	return fmt.Sprintf("snmp host %s", host.Address)
}

// BuildSNMPTrapCommunityCommand builds the command to set the trap community
// Command format: snmp trap community <string>
func BuildSNMPTrapCommunityCommand(community string) string {
	return fmt.Sprintf("snmp trap community %s", community)
}

// BuildSNMPTrapEnableCommand builds the command to enable SNMP traps
// Command format: snmp trap enable snmp <types>
func BuildSNMPTrapEnableCommand(trapTypes []string) string {
	return fmt.Sprintf("snmp trap enable snmp %s", strings.Join(trapTypes, " "))
}

// BuildDeleteSNMPSysNameCommand builds the command to remove the system name
// Command format: no snmp sysname
func BuildDeleteSNMPSysNameCommand() string {
	return "no snmp sysname"
}

// BuildDeleteSNMPSysLocationCommand builds the command to remove the system location
// Command format: no snmp syslocation
func BuildDeleteSNMPSysLocationCommand() string {
	return "no snmp syslocation"
}

// BuildDeleteSNMPSysContactCommand builds the command to remove the system contact
// Command format: no snmp syscontact
func BuildDeleteSNMPSysContactCommand() string {
	return "no snmp syscontact"
}

// BuildDeleteSNMPCommunityCommand builds the command to remove an SNMP community
// Command format: no snmp community read-only|read-write <string>
func BuildDeleteSNMPCommunityCommand(community SNMPCommunity) string {
	permission := "read-only"
	if community.Permission == "rw" {
		permission = "read-write"
	}
	return fmt.Sprintf("no snmp community %s %s", permission, community.Name)
}

// BuildDeleteSNMPHostCommand builds the command to remove an SNMP trap host
// Command format: no snmp host <ip>
func BuildDeleteSNMPHostCommand(address string) string {
	return fmt.Sprintf("no snmp host %s", address)
}

// BuildDeleteSNMPTrapCommunityCommand builds the command to remove the trap community
// Command format: no snmp trap community
func BuildDeleteSNMPTrapCommunityCommand() string {
	return "no snmp trap community"
}

// BuildDeleteSNMPTrapEnableCommand builds the command to disable SNMP traps
// Command format: no snmp trap enable snmp
func BuildDeleteSNMPTrapEnableCommand() string {
	return "no snmp trap enable snmp"
}

// BuildShowSNMPConfigCommand builds the command to show SNMP configuration
func BuildShowSNMPConfigCommand() string {
	return "show config | grep snmp"
}

// ValidateSNMPConfig validates an SNMP configuration
func ValidateSNMPConfig(config SNMPConfig) error {
	// Validate communities
	for _, community := range config.Communities {
		if community.Name == "" {
			return fmt.Errorf("community name cannot be empty")
		}
		if community.Permission != "ro" && community.Permission != "rw" {
			return fmt.Errorf("community permission must be 'ro' or 'rw', got '%s'", community.Permission)
		}
		// Validate community string length (typical SNMP limit)
		if len(community.Name) > 64 {
			return fmt.Errorf("community name '%s' exceeds maximum length of 64 characters", community.Name)
		}
	}

	// Validate hosts
	for _, host := range config.Hosts {
		if host.Address == "" {
			return fmt.Errorf("host address cannot be empty")
		}
		if net.ParseIP(host.Address) == nil {
			return fmt.Errorf("invalid host IP address: %s", host.Address)
		}
		// Validate version if specified
		if host.Version != "" && host.Version != "1" && host.Version != "2c" {
			return fmt.Errorf("invalid SNMP version '%s', must be '1' or '2c'", host.Version)
		}
	}

	// Validate trap types
	validTrapTypes := map[string]bool{
		"all":             true,
		"authentication":  true,
		"coldstart":       true,
		"warmstart":       true,
		"linkdown":        true,
		"linkup":          true,
		"enterprise":      true,
	}
	for _, trapType := range config.TrapEnable {
		if !validTrapTypes[strings.ToLower(trapType)] {
			return fmt.Errorf("invalid trap type '%s'", trapType)
		}
	}

	return nil
}

// ValidateSNMPCommunity validates a single SNMP community
func ValidateSNMPCommunity(community SNMPCommunity) error {
	if community.Name == "" {
		return fmt.Errorf("community name cannot be empty")
	}
	if community.Permission != "ro" && community.Permission != "rw" {
		return fmt.Errorf("community permission must be 'ro' or 'rw', got '%s'", community.Permission)
	}
	if len(community.Name) > 64 {
		return fmt.Errorf("community name '%s' exceeds maximum length of 64 characters", community.Name)
	}
	return nil
}

// ValidateSNMPHost validates a single SNMP host
func ValidateSNMPHost(host SNMPHost) error {
	if host.Address == "" {
		return fmt.Errorf("host address cannot be empty")
	}
	if net.ParseIP(host.Address) == nil {
		return fmt.Errorf("invalid host IP address: %s", host.Address)
	}
	if host.Version != "" && host.Version != "1" && host.Version != "2c" {
		return fmt.Errorf("invalid SNMP version '%s', must be '1' or '2c'", host.Version)
	}
	return nil
}
