package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// VLAN represents a VLAN configuration on an RTX router
type VLAN struct {
	VlanID        int    `json:"vlan_id"`              // VLAN ID (2-4094, 1 is reserved)
	Name          string `json:"name,omitempty"`       // VLAN name/description
	Interface     string `json:"interface"`            // Parent interface (lan1, lan2)
	VlanInterface string `json:"vlan_interface"`       // Computed: lan1/1, lan1/2, etc.
	IPAddress     string `json:"ip_address,omitempty"` // IP address
	IPMask        string `json:"ip_mask,omitempty"`    // Subnet mask
	Shutdown      bool   `json:"shutdown"`             // Admin state (true = shutdown)
}

// VLANParser parses VLAN configuration output
type VLANParser struct{}

// NewVLANParser creates a new VLAN parser
func NewVLANParser() *VLANParser {
	return &VLANParser{}
}

// ParseVLANConfig parses the output of "show config" command for VLAN configuration
// and returns a list of VLANs
func (p *VLANParser) ParseVLANConfig(raw string) ([]VLAN, error) {
	vlans := make(map[string]*VLAN) // key: vlan_interface (e.g., "lan1/1")
	lines := strings.Split(raw, "\n")

	// Patterns for different VLAN configuration lines
	// vlan <interface>/<n> 802.1q vid=<vlan_id>
	vlanPattern := regexp.MustCompile(`^\s*vlan\s+(\w+)/(\d+)\s+802\.1q\s+vid=(\d+)\s*$`)
	// ip <vlan_interface> address <ip>/<prefix> or ip <vlan_interface> address <ip> <mask>
	ipPatternCIDR := regexp.MustCompile(`^\s*ip\s+(\w+/\d+)\s+address\s+([0-9.]+)/(\d+)\s*$`)
	ipPatternMask := regexp.MustCompile(`^\s*ip\s+(\w+/\d+)\s+address\s+([0-9.]+)\s+([0-9.]+)\s*$`)
	// description <vlan_interface> <name>
	descPattern := regexp.MustCompile(`^\s*description\s+(\w+/\d+)\s+(.+)\s*$`)
	// <vlan_interface> enable or no <vlan_interface> enable
	enablePattern := regexp.MustCompile(`^\s*(\w+/\d+)\s+enable\s*$`)
	noEnablePattern := regexp.MustCompile(`^\s*no\s+(\w+/\d+)\s+enable\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try VLAN definition pattern: vlan lan1/1 802.1q vid=10
		if matches := vlanPattern.FindStringSubmatch(line); len(matches) >= 4 {
			iface := matches[1]
			slot := matches[2]
			vlanID, err := strconv.Atoi(matches[3])
			if err != nil {
				continue
			}

			vlanInterface := fmt.Sprintf("%s/%s", iface, slot)
			vlan, exists := vlans[vlanInterface]
			if !exists {
				vlan = &VLAN{
					Interface:     iface,
					VlanInterface: vlanInterface,
					Shutdown:      false, // Default: enabled
				}
				vlans[vlanInterface] = vlan
			}
			vlan.VlanID = vlanID
			continue
		}

		// Try IP address pattern with CIDR notation
		if matches := ipPatternCIDR.FindStringSubmatch(line); len(matches) >= 4 {
			vlanInterface := matches[1]
			ipAddr := matches[2]
			prefix, err := strconv.Atoi(matches[3])
			if err != nil {
				continue
			}

			vlan, exists := vlans[vlanInterface]
			if !exists {
				// Create VLAN entry if not exists (might be defined later)
				parts := strings.Split(vlanInterface, "/")
				if len(parts) == 2 {
					vlan = &VLAN{
						Interface:     parts[0],
						VlanInterface: vlanInterface,
						Shutdown:      false,
					}
					vlans[vlanInterface] = vlan
				} else {
					continue
				}
			}
			vlan.IPAddress = ipAddr
			vlan.IPMask = prefixToMask(prefix)
			continue
		}

		// Try IP address pattern with dotted decimal mask
		if matches := ipPatternMask.FindStringSubmatch(line); len(matches) >= 4 {
			vlanInterface := matches[1]
			ipAddr := matches[2]
			mask := matches[3]

			vlan, exists := vlans[vlanInterface]
			if !exists {
				parts := strings.Split(vlanInterface, "/")
				if len(parts) == 2 {
					vlan = &VLAN{
						Interface:     parts[0],
						VlanInterface: vlanInterface,
						Shutdown:      false,
					}
					vlans[vlanInterface] = vlan
				} else {
					continue
				}
			}
			vlan.IPAddress = ipAddr
			vlan.IPMask = mask
			continue
		}

		// Try description pattern
		if matches := descPattern.FindStringSubmatch(line); len(matches) >= 3 {
			vlanInterface := matches[1]
			name := strings.TrimSpace(matches[2])

			vlan, exists := vlans[vlanInterface]
			if !exists {
				parts := strings.Split(vlanInterface, "/")
				if len(parts) == 2 {
					vlan = &VLAN{
						Interface:     parts[0],
						VlanInterface: vlanInterface,
						Shutdown:      false,
					}
					vlans[vlanInterface] = vlan
				} else {
					continue
				}
			}
			vlan.Name = name
			continue
		}

		// Try no enable pattern (shutdown)
		if matches := noEnablePattern.FindStringSubmatch(line); len(matches) >= 2 {
			vlanInterface := matches[1]
			if vlan, exists := vlans[vlanInterface]; exists {
				vlan.Shutdown = true
			}
			continue
		}

		// Try enable pattern (not shutdown) - this is the default state
		if matches := enablePattern.FindStringSubmatch(line); len(matches) >= 2 {
			vlanInterface := matches[1]
			if vlan, exists := vlans[vlanInterface]; exists {
				vlan.Shutdown = false
			}
			continue
		}
	}

	// Convert map to slice
	result := make([]VLAN, 0, len(vlans))
	for _, vlan := range vlans {
		// Only include VLANs that have a valid VLAN ID
		if vlan.VlanID > 0 {
			result = append(result, *vlan)
		}
	}

	return result, nil
}

// ParseSingleVLAN parses configuration for a specific VLAN interface
func (p *VLANParser) ParseSingleVLAN(raw string, iface string, vlanID int) (*VLAN, error) {
	vlans, err := p.ParseVLANConfig(raw)
	if err != nil {
		return nil, err
	}

	for _, vlan := range vlans {
		if vlan.Interface == iface && vlan.VlanID == vlanID {
			return &vlan, nil
		}
	}

	return nil, fmt.Errorf("VLAN %d on interface %s not found", vlanID, iface)
}

// BuildVLANCommand builds the command to create a VLAN interface
// Command format: vlan <interface>/<n> 802.1q vid=<vlan_id>
// Note: The slot number <n> is determined by finding the next available slot
func BuildVLANCommand(iface string, slot int, vlanID int) string {
	return fmt.Sprintf("vlan %s/%d 802.1q vid=%d", iface, slot, vlanID)
}

// BuildVLANIPCommand builds the command to set IP address on a VLAN interface
// Command format: ip <vlan_interface> address <ip>/<prefix>
func BuildVLANIPCommand(vlanInterface string, ipAddr string, mask string) string {
	prefix := maskToPrefix(mask)
	return fmt.Sprintf("ip %s address %s/%d", vlanInterface, ipAddr, prefix)
}

// BuildDeleteVLANIPCommand builds the command to remove IP address from a VLAN interface
// Command format: no ip <vlan_interface> address
func BuildDeleteVLANIPCommand(vlanInterface string) string {
	return fmt.Sprintf("no ip %s address", vlanInterface)
}

// BuildVLANDescriptionCommand builds the command to set description on a VLAN interface
// Command format: description <vlan_interface> <name>
func BuildVLANDescriptionCommand(vlanInterface string, name string) string {
	return fmt.Sprintf("description %s %s", vlanInterface, name)
}

// BuildDeleteVLANDescriptionCommand builds the command to remove description from a VLAN interface
// Command format: no description <vlan_interface>
func BuildDeleteVLANDescriptionCommand(vlanInterface string) string {
	return fmt.Sprintf("no description %s", vlanInterface)
}

// BuildVLANEnableCommand builds the command to enable a VLAN interface
// Command format: <vlan_interface> enable
func BuildVLANEnableCommand(vlanInterface string) string {
	return fmt.Sprintf("%s enable", vlanInterface)
}

// BuildVLANDisableCommand builds the command to disable a VLAN interface
// Command format: no <vlan_interface> enable
func BuildVLANDisableCommand(vlanInterface string) string {
	return fmt.Sprintf("no %s enable", vlanInterface)
}

// BuildDeleteVLANCommand builds the command to delete a VLAN interface
// Command format: no vlan <interface>/<n>
func BuildDeleteVLANCommand(vlanInterface string) string {
	return fmt.Sprintf("no vlan %s", vlanInterface)
}

// BuildShowVLANCommand builds the command to show VLAN configuration
// Command format: show config | grep "<interface>/<n>\|vlan"
func BuildShowVLANCommand(iface string, vlanID int) string {
	// To get all VLAN-related config, we need to search for both the vlan definition
	// and the IP/description configuration for the VLAN interface
	return fmt.Sprintf("show config | grep \"vlan %s.*vid=%d\\|%s/\"", iface, vlanID, iface)
}

// BuildShowAllVLANsCommand builds the command to show all VLAN configurations
// Command format: show config | grep "vlan\|lan./"
func BuildShowAllVLANsCommand() string {
	return "show config | grep \"vlan\\|lan./\""
}

// ValidateVLAN validates a VLAN configuration
func ValidateVLAN(vlan VLAN) error {
	// Validate VLAN ID (2-4094, VLAN ID 1 is reserved)
	if vlan.VlanID < 2 || vlan.VlanID > 4094 {
		return fmt.Errorf("vlan_id must be 2-4094 (1 is reserved), got %d", vlan.VlanID)
	}

	// Validate interface name
	if vlan.Interface == "" {
		return fmt.Errorf("interface is required")
	}

	// Validate interface format (lan1, lan2, lan3, etc.)
	validIfacePattern := regexp.MustCompile(`^lan\d+$`)
	if !validIfacePattern.MatchString(vlan.Interface) {
		return fmt.Errorf("interface must be in format 'lanN' (e.g., lan1, lan2), got %s", vlan.Interface)
	}

	// Validate IP address if provided
	if vlan.IPAddress != "" {
		if !isValidIP(vlan.IPAddress) {
			return fmt.Errorf("invalid IP address: %s", vlan.IPAddress)
		}
		// If IP is provided, mask must also be provided
		if vlan.IPMask == "" {
			return fmt.Errorf("ip_mask is required when ip_address is specified")
		}
		if !isValidMask(vlan.IPMask) {
			return fmt.Errorf("invalid IP mask: %s", vlan.IPMask)
		}
	}

	// If mask is provided, IP should also be provided
	if vlan.IPMask != "" && vlan.IPAddress == "" {
		return fmt.Errorf("ip_address is required when ip_mask is specified")
	}

	return nil
}

// isValidMask checks if a string is a valid subnet mask
func isValidMask(mask string) bool {
	if !isValidIP(mask) {
		return false
	}

	// Convert to binary and check it's a valid mask (contiguous 1s followed by 0s)
	parts := strings.Split(mask, ".")
	var binary uint32
	for _, part := range parts {
		num, _ := strconv.Atoi(part)
		binary = (binary << 8) | uint32(num)
	}

	// A valid mask should be: some number of 1s followed by 0s
	// After inverting, it should be contiguous 0s followed by 1s
	inverted := ^binary
	// Adding 1 to inverted should result in power of 2
	if inverted == 0 {
		return true // All 1s (255.255.255.255)
	}
	return (inverted & (inverted + 1)) == 0
}

// prefixToMask converts CIDR prefix length to dotted decimal subnet mask
func prefixToMask(prefix int) string {
	if prefix < 0 || prefix > 32 {
		return ""
	}

	var mask uint32 = 0xFFFFFFFF << (32 - prefix)
	return fmt.Sprintf("%d.%d.%d.%d",
		(mask>>24)&0xFF,
		(mask>>16)&0xFF,
		(mask>>8)&0xFF,
		mask&0xFF)
}

// maskToPrefix converts dotted decimal subnet mask to CIDR prefix length
func maskToPrefix(mask string) int {
	parts := strings.Split(mask, ".")
	if len(parts) != 4 {
		return 0
	}

	var binary uint32
	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil || num < 0 || num > 255 {
			return 0
		}
		binary = (binary << 8) | uint32(num)
	}

	// Count leading 1s
	prefix := 0
	for i := 31; i >= 0; i-- {
		if (binary & (1 << i)) != 0 {
			prefix++
		} else {
			break
		}
	}

	return prefix
}

// FindNextAvailableSlot finds the next available slot number for a VLAN on an interface
// This is used when creating a new VLAN to determine which slot to use
func FindNextAvailableSlot(existingVLANs []VLAN, iface string) int {
	usedSlots := make(map[int]bool)

	for _, vlan := range existingVLANs {
		if vlan.Interface == iface {
			// Extract slot number from VlanInterface (e.g., "lan1/1" -> 1)
			parts := strings.Split(vlan.VlanInterface, "/")
			if len(parts) == 2 {
				slot, err := strconv.Atoi(parts[1])
				if err == nil {
					usedSlots[slot] = true
				}
			}
		}
	}

	// Find the first available slot starting from 1
	for slot := 1; slot <= 4096; slot++ {
		if !usedSlots[slot] {
			return slot
		}
	}

	return -1 // No available slots
}
