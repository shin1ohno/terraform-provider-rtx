package parsers

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// NATMasquerade represents a NAT masquerade configuration on an RTX router
type NATMasquerade struct {
	DescriptorID  int                      `json:"descriptor_id"`
	OuterAddress  string                   `json:"outer_address"`            // "ipcp", interface name, or specific IP
	InnerNetwork  string                   `json:"inner_network"`            // IP range: "192.168.1.0-192.168.1.255"
	StaticEntries []MasqueradeStaticEntry  `json:"static_entries,omitempty"` // Static port mappings
}

// MasqueradeStaticEntry represents a static port mapping entry
type MasqueradeStaticEntry struct {
	EntryNumber       int    `json:"entry_number"`
	InsideLocal       string `json:"inside_local"`        // Internal IP address
	InsideLocalPort   int    `json:"inside_local_port"`   // Internal port
	OutsideGlobal     string `json:"outside_global"`      // External IP address (or "ipcp")
	OutsideGlobalPort int    `json:"outside_global_port"` // External port
	Protocol          string `json:"protocol,omitempty"`  // "tcp", "udp", or empty for any
}

// ParseNATMasqueradeConfig parses the output of "show config" command
// for NAT descriptor masquerade lines
func ParseNATMasqueradeConfig(raw string) ([]NATMasquerade, error) {
	descriptors := make(map[int]*NATMasquerade)
	lines := strings.Split(raw, "\n")

	// nat descriptor type <id> masquerade
	typePattern := regexp.MustCompile(`^\s*nat\s+descriptor\s+type\s+(\d+)\s+masquerade\s*$`)
	// nat descriptor address outer <id> <address>
	outerPattern := regexp.MustCompile(`^\s*nat\s+descriptor\s+address\s+outer\s+(\d+)\s+(\S+)\s*$`)
	// nat descriptor address inner <id> <range>
	innerPattern := regexp.MustCompile(`^\s*nat\s+descriptor\s+address\s+inner\s+(\d+)\s+(\S+)\s*$`)
	// nat descriptor masquerade static <id> <entry> <outer:port>=<inner:port> [protocol]
	// Format: nat descriptor masquerade static 1 1 203.0.113.1:80=192.168.1.100:8080 tcp
	staticPattern := regexp.MustCompile(`^\s*nat\s+descriptor\s+masquerade\s+static\s+(\d+)\s+(\d+)\s+([^:]+):(\d+)=([^:]+):(\d+)(?:\s+(\S+))?\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try type pattern
		if matches := typePattern.FindStringSubmatch(line); len(matches) >= 2 {
			id, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			if _, exists := descriptors[id]; !exists {
				descriptors[id] = &NATMasquerade{
					DescriptorID:  id,
					StaticEntries: []MasqueradeStaticEntry{},
				}
			}
			continue
		}

		// Try outer address pattern
		if matches := outerPattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			desc, exists := descriptors[id]
			if !exists {
				desc = &NATMasquerade{
					DescriptorID:  id,
					StaticEntries: []MasqueradeStaticEntry{},
				}
				descriptors[id] = desc
			}
			desc.OuterAddress = matches[2]
			continue
		}

		// Try inner address pattern
		if matches := innerPattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			desc, exists := descriptors[id]
			if !exists {
				desc = &NATMasquerade{
					DescriptorID:  id,
					StaticEntries: []MasqueradeStaticEntry{},
				}
				descriptors[id] = desc
			}
			desc.InnerNetwork = matches[2]
			continue
		}

		// Try static entry pattern
		if matches := staticPattern.FindStringSubmatch(line); len(matches) >= 7 {
			id, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}
			entryNum, err := strconv.Atoi(matches[2])
			if err != nil {
				continue
			}
			outerPort, err := strconv.Atoi(matches[4])
			if err != nil {
				continue
			}
			innerPort, err := strconv.Atoi(matches[6])
			if err != nil {
				continue
			}

			desc, exists := descriptors[id]
			if !exists {
				desc = &NATMasquerade{
					DescriptorID:  id,
					StaticEntries: []MasqueradeStaticEntry{},
				}
				descriptors[id] = desc
			}

			entry := MasqueradeStaticEntry{
				EntryNumber:       entryNum,
				OutsideGlobal:     matches[3],
				OutsideGlobalPort: outerPort,
				InsideLocal:       matches[5],
				InsideLocalPort:   innerPort,
			}
			if len(matches) > 7 && matches[7] != "" {
				entry.Protocol = strings.ToLower(matches[7])
			}
			desc.StaticEntries = append(desc.StaticEntries, entry)
			continue
		}
	}

	// Convert map to slice
	result := make([]NATMasquerade, 0, len(descriptors))
	for _, desc := range descriptors {
		result = append(result, *desc)
	}

	return result, nil
}

// BuildNATDescriptorTypeMasqueradeCommand generates "nat descriptor type N masquerade" command
func BuildNATDescriptorTypeMasqueradeCommand(id int) string {
	return fmt.Sprintf("nat descriptor type %d masquerade", id)
}

// BuildNATDescriptorAddressOuterCommand generates "nat descriptor address outer N address" command
func BuildNATDescriptorAddressOuterCommand(id int, address string) string {
	return fmt.Sprintf("nat descriptor address outer %d %s", id, address)
}

// BuildNATDescriptorAddressInnerCommand generates "nat descriptor address inner N network" command
func BuildNATDescriptorAddressInnerCommand(id int, network string) string {
	return fmt.Sprintf("nat descriptor address inner %d %s", id, network)
}

// BuildNATMasqueradeStaticCommand generates static port mapping command
// Format: nat descriptor masquerade static <id> <entry> <outer:port>=<inner:port> [protocol]
func BuildNATMasqueradeStaticCommand(id int, entryNum int, entry MasqueradeStaticEntry) string {
	cmd := fmt.Sprintf("nat descriptor masquerade static %d %d %s:%d=%s:%d",
		id, entryNum,
		entry.OutsideGlobal, entry.OutsideGlobalPort,
		entry.InsideLocal, entry.InsideLocalPort)

	if entry.Protocol != "" {
		cmd += " " + strings.ToLower(entry.Protocol)
	}

	return cmd
}

// BuildDeleteNATMasqueradeCommand generates "no nat descriptor type N" command
func BuildDeleteNATMasqueradeCommand(id int) string {
	return fmt.Sprintf("no nat descriptor type %d", id)
}

// BuildInterfaceNATDescriptorCommand generates "ip <iface> nat descriptor N" command
func BuildInterfaceNATDescriptorCommand(iface string, descriptorID int) string {
	return fmt.Sprintf("ip %s nat descriptor %d", iface, descriptorID)
}

// BuildDeleteInterfaceNATDescriptorCommand generates "no ip <iface> nat descriptor N" command
func BuildDeleteInterfaceNATDescriptorCommand(iface string, descriptorID int) string {
	return fmt.Sprintf("no ip %s nat descriptor %d", iface, descriptorID)
}

// BuildDeleteNATMasqueradeStaticCommand generates command to delete a static entry
func BuildDeleteNATMasqueradeStaticCommand(id int, entryNum int) string {
	return fmt.Sprintf("no nat descriptor masquerade static %d %d", id, entryNum)
}

// BuildShowNATDescriptorCommand builds command to show NAT descriptor configuration
func BuildShowNATDescriptorCommand(id int) string {
	return fmt.Sprintf("show config | grep \"nat descriptor.*%d\"", id)
}

// BuildShowAllNATDescriptorsCommand builds command to show all NAT descriptors
func BuildShowAllNATDescriptorsCommand() string {
	return "show config | grep \"nat descriptor\""
}

// ValidateDescriptorID validates that descriptor ID is within valid range (1-65535)
func ValidateDescriptorID(id int) error {
	if id < 1 || id > 65535 {
		return fmt.Errorf("descriptor ID must be between 1 and 65535, got %d", id)
	}
	return nil
}

// ValidateCIDR validates CIDR format
func ValidateCIDR(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR notation: %s", cidr)
	}
	return nil
}

// ConvertCIDRToRange converts CIDR notation to IP range
// e.g., "192.168.1.0/24" -> ("192.168.1.0", "192.168.1.255")
func ConvertCIDRToRange(cidr string) (string, string, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", "", fmt.Errorf("invalid CIDR notation: %s", cidr)
	}

	// Get the network address (start)
	start := ipNet.IP.To4()
	if start == nil {
		return "", "", fmt.Errorf("only IPv4 CIDR is supported: %s", cidr)
	}

	// Calculate the broadcast address (end)
	mask := ipNet.Mask
	end := make(net.IP, len(start))
	for i := range start {
		end[i] = start[i] | ^mask[i]
	}

	return start.String(), end.String(), nil
}

// ConvertRangeToRTXFormat converts CIDR to RTX range format
// e.g., "192.168.1.0/24" -> "192.168.1.0-192.168.1.255"
func ConvertRangeToRTXFormat(cidr string) (string, error) {
	start, end, err := ConvertCIDRToRange(cidr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", start, end), nil
}

// ValidateNATPort validates that a port number is within valid range (1-65535)
func ValidateNATPort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}
	return nil
}

// ValidateNATProtocol validates that protocol is tcp, udp, or empty
func ValidateNATProtocol(protocol string) error {
	protocol = strings.ToLower(protocol)
	if protocol != "" && protocol != "tcp" && protocol != "udp" {
		return fmt.Errorf("protocol must be 'tcp', 'udp', or empty, got '%s'", protocol)
	}
	return nil
}

// ValidateOuterAddress validates outer address format
// Can be: "ipcp", interface name (e.g., "pp1"), or IP address
func ValidateOuterAddress(address string) error {
	if address == "" {
		return fmt.Errorf("outer address cannot be empty")
	}

	// "ipcp" is a special value for PPPoE
	if address == "ipcp" {
		return nil
	}

	// Check if it's an interface name (starts with common prefixes)
	if strings.HasPrefix(address, "pp") ||
		strings.HasPrefix(address, "lan") ||
		strings.HasPrefix(address, "tunnel") {
		return nil
	}

	// Check if it's a valid IP address
	if net.ParseIP(address) != nil {
		return nil
	}

	return fmt.Errorf("outer address must be 'ipcp', interface name, or valid IP address: %s", address)
}

// ValidateNATMasquerade validates a NAT masquerade configuration
func ValidateNATMasquerade(nat NATMasquerade) error {
	if err := ValidateDescriptorID(nat.DescriptorID); err != nil {
		return err
	}

	if err := ValidateOuterAddress(nat.OuterAddress); err != nil {
		return err
	}

	// Inner network should be in range format or CIDR
	if nat.InnerNetwork == "" {
		return fmt.Errorf("inner network cannot be empty")
	}

	// Validate static entries
	for i, entry := range nat.StaticEntries {
		if err := ValidateNATPort(entry.InsideLocalPort); err != nil {
			return fmt.Errorf("static entry %d: %w", i+1, err)
		}
		if err := ValidateNATPort(entry.OutsideGlobalPort); err != nil {
			return fmt.Errorf("static entry %d: %w", i+1, err)
		}
		if err := ValidateNATProtocol(entry.Protocol); err != nil {
			return fmt.Errorf("static entry %d: %w", i+1, err)
		}
		// Validate InsideLocal is a valid IP
		if net.ParseIP(entry.InsideLocal) == nil {
			return fmt.Errorf("static entry %d: invalid inside local IP: %s", i+1, entry.InsideLocal)
		}
	}

	return nil
}
