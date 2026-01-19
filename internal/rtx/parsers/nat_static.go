package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// NATStatic represents a static NAT descriptor configuration on an RTX router
type NATStatic struct {
	DescriptorID int              `json:"descriptor_id"`
	Entries      []NATStaticEntry `json:"entries,omitempty"`
}

// NATStaticEntry represents a single static NAT mapping entry
type NATStaticEntry struct {
	InsideLocal       string `json:"inside_local"`                  // Inside local IP address
	InsideLocalPort   int    `json:"inside_local_port,omitempty"`   // Inside local port (for port NAT)
	OutsideGlobal     string `json:"outside_global"`                // Outside global IP address
	OutsideGlobalPort int    `json:"outside_global_port,omitempty"` // Outside global port (for port NAT)
	Protocol          string `json:"protocol,omitempty"`            // Protocol: tcp, udp (for port NAT)
}

// NATStaticParser parses static NAT configuration output
type NATStaticParser struct{}

// NewNATStaticParser creates a new static NAT parser
func NewNATStaticParser() *NATStaticParser {
	return &NATStaticParser{}
}

// ParseNATStaticConfig parses the output of "show config" command
// and returns a list of static NAT descriptors
func ParseNATStaticConfig(raw string) ([]NATStatic, error) {
	descriptors := make(map[int]*NATStatic)
	lines := strings.Split(raw, "\n")

	// Pattern for NAT descriptor type definition
	// nat descriptor type <id> static
	typePattern := regexp.MustCompile(`^\s*nat\s+descriptor\s+type\s+(\d+)\s+static\s*$`)

	// Pattern for 1:1 static NAT mapping
	// nat descriptor static <id> <outer_ip>=<inner_ip>
	staticPattern := regexp.MustCompile(`^\s*nat\s+descriptor\s+static\s+(\d+)\s+([0-9.]+)=([0-9.]+)\s*$`)

	// Pattern for port-based static NAT mapping
	// nat descriptor static <id> <outer_ip>:<port>=<inner_ip>:<port> <protocol>
	portStaticPattern := regexp.MustCompile(`^\s*nat\s+descriptor\s+static\s+(\d+)\s+([0-9.]+):(\d+)=([0-9.]+):(\d+)\s+(tcp|udp)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try NAT descriptor type pattern
		if matches := typePattern.FindStringSubmatch(line); len(matches) >= 2 {
			descriptorID, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			if _, exists := descriptors[descriptorID]; !exists {
				descriptors[descriptorID] = &NATStatic{
					DescriptorID: descriptorID,
					Entries:      []NATStaticEntry{},
				}
			}
			continue
		}

		// Try port-based static NAT pattern (check this before simple static pattern)
		if matches := portStaticPattern.FindStringSubmatch(line); len(matches) >= 7 {
			descriptorID, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			descriptor, exists := descriptors[descriptorID]
			if !exists {
				descriptor = &NATStatic{
					DescriptorID: descriptorID,
					Entries:      []NATStaticEntry{},
				}
				descriptors[descriptorID] = descriptor
			}

			outsidePort, _ := strconv.Atoi(matches[3])
			insidePort, _ := strconv.Atoi(matches[5])

			entry := NATStaticEntry{
				OutsideGlobal:     matches[2],
				OutsideGlobalPort: outsidePort,
				InsideLocal:       matches[4],
				InsideLocalPort:   insidePort,
				Protocol:          strings.ToLower(matches[6]),
			}
			descriptor.Entries = append(descriptor.Entries, entry)
			continue
		}

		// Try simple 1:1 static NAT pattern
		if matches := staticPattern.FindStringSubmatch(line); len(matches) >= 4 {
			descriptorID, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			descriptor, exists := descriptors[descriptorID]
			if !exists {
				descriptor = &NATStatic{
					DescriptorID: descriptorID,
					Entries:      []NATStaticEntry{},
				}
				descriptors[descriptorID] = descriptor
			}

			entry := NATStaticEntry{
				OutsideGlobal: matches[2],
				InsideLocal:   matches[3],
			}
			descriptor.Entries = append(descriptor.Entries, entry)
			continue
		}
	}

	// Convert map to slice
	result := make([]NATStatic, 0, len(descriptors))
	for _, descriptor := range descriptors {
		result = append(result, *descriptor)
	}

	return result, nil
}

// ParseSingleNATStatic parses configuration for a specific descriptor ID
func (p *NATStaticParser) ParseSingleNATStatic(raw string, descriptorID int) (*NATStatic, error) {
	descriptors, err := ParseNATStaticConfig(raw)
	if err != nil {
		return nil, err
	}

	for _, descriptor := range descriptors {
		if descriptor.DescriptorID == descriptorID {
			return &descriptor, nil
		}
	}

	return nil, fmt.Errorf("NAT descriptor %d not found", descriptorID)
}

// BuildNATDescriptorTypeStaticCommand builds the command to set NAT descriptor type to static
// Command format: nat descriptor type <id> static
func BuildNATDescriptorTypeStaticCommand(id int) string {
	return fmt.Sprintf("nat descriptor type %d static", id)
}

// BuildNATStaticMappingCommand builds the command for 1:1 static NAT mapping
// Command format: nat descriptor static <id> <outer_ip>=<inner_ip>
func BuildNATStaticMappingCommand(id int, entry NATStaticEntry) string {
	return fmt.Sprintf("nat descriptor static %d %s=%s", id, entry.OutsideGlobal, entry.InsideLocal)
}

// BuildNATStaticPortMappingCommand builds the command for port-based static NAT mapping
// Command format: nat descriptor static <id> <outer_ip>:<port>=<inner_ip>:<port> <protocol>
func BuildNATStaticPortMappingCommand(id int, entry NATStaticEntry) string {
	return fmt.Sprintf("nat descriptor static %d %s:%d=%s:%d %s",
		id, entry.OutsideGlobal, entry.OutsideGlobalPort,
		entry.InsideLocal, entry.InsideLocalPort,
		strings.ToLower(entry.Protocol))
}

// BuildDeleteNATStaticCommand builds the command to delete a NAT descriptor
// Command format: no nat descriptor type <id>
func BuildDeleteNATStaticCommand(id int) string {
	return fmt.Sprintf("no nat descriptor type %d", id)
}

// BuildDeleteNATStaticMappingCommand builds the command to delete a specific 1:1 NAT mapping
// Command format: no nat descriptor static <id> <outer_ip>=<inner_ip>
func BuildDeleteNATStaticMappingCommand(id int, entry NATStaticEntry) string {
	return fmt.Sprintf("no nat descriptor static %d %s=%s", id, entry.OutsideGlobal, entry.InsideLocal)
}

// BuildDeleteNATStaticPortMappingCommand builds the command to delete a port-based NAT mapping
// Command format: no nat descriptor static <id> <outer_ip>:<port>=<inner_ip>:<port> <protocol>
func BuildDeleteNATStaticPortMappingCommand(id int, entry NATStaticEntry) string {
	return fmt.Sprintf("no nat descriptor static %d %s:%d=%s:%d %s",
		id, entry.OutsideGlobal, entry.OutsideGlobalPort,
		entry.InsideLocal, entry.InsideLocalPort,
		strings.ToLower(entry.Protocol))
}

// BuildInterfaceNATCommand builds the command to apply NAT descriptor to an interface
// Command format: ip <interface> nat descriptor <id>
func BuildInterfaceNATCommand(iface string, descriptorID int) string {
	return fmt.Sprintf("ip %s nat descriptor %d", iface, descriptorID)
}

// BuildDeleteInterfaceNATCommand builds the command to remove NAT descriptor from an interface
// Command format: no ip <interface> nat descriptor <id>
func BuildDeleteInterfaceNATCommand(iface string, descriptorID int) string {
	return fmt.Sprintf("no ip %s nat descriptor %d", iface, descriptorID)
}

// BuildShowNATStaticCommand builds the command to show NAT descriptor configuration
// Command format: show config | grep "nat descriptor"
func BuildShowNATStaticCommand(descriptorID int) string {
	return fmt.Sprintf("show config | grep \"nat descriptor.*%d\"", descriptorID)
}

// BuildShowAllNATStaticCommand builds the command to show all NAT descriptors
// Command format: show config | grep "nat descriptor"
func BuildShowAllNATStaticCommand() string {
	return "show config | grep \"nat descriptor\""
}

// validateNATStaticProtocol validates the protocol string for port-based NAT (tcp or udp required)
func validateNATStaticProtocol(protocol string) error {
	proto := strings.ToLower(protocol)
	if proto != "tcp" && proto != "udp" {
		return fmt.Errorf("protocol must be 'tcp' or 'udp', got '%s'", protocol)
	}
	return nil
}

// ValidateNATStaticEntry validates a NAT static entry
func ValidateNATStaticEntry(entry NATStaticEntry) error {
	if entry.InsideLocal == "" {
		return fmt.Errorf("inside_local is required")
	}
	if !isValidIP(entry.InsideLocal) {
		return fmt.Errorf("invalid inside_local IP address: %s", entry.InsideLocal)
	}

	if entry.OutsideGlobal == "" {
		return fmt.Errorf("outside_global is required")
	}
	if !isValidIP(entry.OutsideGlobal) {
		return fmt.Errorf("invalid outside_global IP address: %s", entry.OutsideGlobal)
	}

	// Port-based NAT validation
	isPortNAT := entry.InsideLocalPort > 0 || entry.OutsideGlobalPort > 0 || entry.Protocol != ""

	if isPortNAT {
		if entry.InsideLocalPort == 0 {
			return fmt.Errorf("inside_local_port is required for port-based NAT")
		}
		if err := ValidateNATPort(entry.InsideLocalPort); err != nil {
			return fmt.Errorf("invalid inside_local_port: %w", err)
		}

		if entry.OutsideGlobalPort == 0 {
			return fmt.Errorf("outside_global_port is required for port-based NAT")
		}
		if err := ValidateNATPort(entry.OutsideGlobalPort); err != nil {
			return fmt.Errorf("invalid outside_global_port: %w", err)
		}

		if entry.Protocol == "" {
			return fmt.Errorf("protocol is required for port-based NAT")
		}
		if err := validateNATStaticProtocol(entry.Protocol); err != nil {
			return err
		}
	}

	return nil
}

// ValidateNATStatic validates a complete NAT static configuration
func ValidateNATStatic(nat NATStatic) error {
	if err := ValidateDescriptorID(nat.DescriptorID); err != nil {
		return err
	}

	for i, entry := range nat.Entries {
		if err := ValidateNATStaticEntry(entry); err != nil {
			return fmt.Errorf("entry %d: %w", i, err)
		}
	}

	return nil
}

// IsPortBasedNAT checks if an entry is a port-based NAT (vs simple 1:1 NAT)
func IsPortBasedNAT(entry NATStaticEntry) bool {
	return entry.InsideLocalPort > 0 && entry.OutsideGlobalPort > 0 && entry.Protocol != ""
}

// BuildNATStaticCommands builds all commands needed to create a NAT static configuration
func BuildNATStaticCommands(nat NATStatic) []string {
	var commands []string

	// First, set the NAT descriptor type
	commands = append(commands, BuildNATDescriptorTypeStaticCommand(nat.DescriptorID))

	// Then add each mapping
	for _, entry := range nat.Entries {
		if IsPortBasedNAT(entry) {
			commands = append(commands, BuildNATStaticPortMappingCommand(nat.DescriptorID, entry))
		} else {
			commands = append(commands, BuildNATStaticMappingCommand(nat.DescriptorID, entry))
		}
	}

	return commands
}
