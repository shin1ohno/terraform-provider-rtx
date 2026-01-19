package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// EthernetFilter represents an Ethernet (Layer 2) filter configuration on an RTX router
type EthernetFilter struct {
	Number    int    `json:"number"`               // Filter number (1-65535)
	Action    string `json:"action"`               // pass or reject
	SourceMAC string `json:"source_mac"`           // Source MAC address (* for any)
	DestMAC   string `json:"dest_mac"`             // Destination MAC address (* for any)
	EtherType string `json:"ether_type,omitempty"` // Ethernet type (e.g., 0x0800, 0x0806)
	VlanID    int    `json:"vlan_id,omitempty"`    // VLAN ID (1-4094, 0 means not specified)
}

// EthernetFilterParser parses Ethernet filter configuration output
type EthernetFilterParser struct{}

// NewEthernetFilterParser creates a new Ethernet filter parser
func NewEthernetFilterParser() *EthernetFilterParser {
	return &EthernetFilterParser{}
}

// NormalizeMAC converts MAC address between different formats
// Supported input formats:
//   - 00:11:22:33:44:55 (colon-separated)
//   - 0011.2233.4455 (Cisco dot notation)
//   - 00-11-22-33-44-55 (hyphen-separated)
//   - 001122334455 (no separator)
//   - * (wildcard)
//
// Output format: 00:11:22:33:44:55 (colon-separated, lowercase)
func NormalizeMAC(mac string) string {
	mac = strings.TrimSpace(mac)

	// Wildcard - return as-is
	if mac == "*" {
		return "*"
	}

	// Remove all separators and convert to lowercase
	clean := strings.ToLower(mac)
	clean = strings.ReplaceAll(clean, ":", "")
	clean = strings.ReplaceAll(clean, "-", "")
	clean = strings.ReplaceAll(clean, ".", "")

	// Check if we have exactly 12 hex characters
	if len(clean) != 12 {
		return mac // Return original if invalid
	}

	// Verify all characters are hex
	for _, c := range clean {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return mac // Return original if invalid
		}
	}

	// Format as colon-separated
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		clean[0:2], clean[2:4], clean[4:6],
		clean[6:8], clean[8:10], clean[10:12])
}

// ConvertMACToCisco converts a MAC address to Cisco dot notation (0011.2233.4455)
func ConvertMACToCisco(mac string) string {
	mac = strings.TrimSpace(mac)

	// Wildcard - return as-is
	if mac == "*" {
		return "*"
	}

	// Normalize first to get consistent format
	normalized := NormalizeMAC(mac)
	if normalized == mac && len(mac) != 17 {
		return mac // Return original if normalization failed
	}

	// Remove colons
	clean := strings.ReplaceAll(normalized, ":", "")
	if len(clean) != 12 {
		return mac
	}

	// Format as Cisco notation
	return fmt.Sprintf("%s.%s.%s",
		clean[0:4], clean[4:8], clean[8:12])
}

// ParseEthernetFilterConfig parses the output of "show config" command
// and extracts Ethernet filter configurations
func ParseEthernetFilterConfig(raw string) ([]EthernetFilter, error) {
	filters := []EthernetFilter{}
	lines := strings.Split(raw, "\n")

	// Pattern: ethernet filter <n> <action> <src_mac> <dst_mac> [<eth_type>] [vlan <vlan_id>]
	filterPattern := regexp.MustCompile(`^\s*ethernet\s+filter\s+(\d+)\s+(pass|reject)\s+(\S+)\s+(\S+)(?:\s+(0x[0-9a-fA-F]+|\*))?(?:\s+vlan\s+(\d+))?\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := filterPattern.FindStringSubmatch(line)
		if len(matches) >= 5 {
			number, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			filter := EthernetFilter{
				Number:    number,
				Action:    matches[2],
				SourceMAC: matches[3],
				DestMAC:   matches[4],
			}

			// EtherType (optional)
			if len(matches) > 5 && matches[5] != "" && matches[5] != "*" {
				filter.EtherType = strings.ToLower(matches[5])
			}

			// VLAN ID (optional)
			if len(matches) > 6 && matches[6] != "" {
				vlanID, err := strconv.Atoi(matches[6])
				if err == nil {
					filter.VlanID = vlanID
				}
			}

			filters = append(filters, filter)
		}
	}

	return filters, nil
}

// ParseSingleEthernetFilter parses configuration for a specific filter number
func ParseSingleEthernetFilter(raw string, filterNumber int) (*EthernetFilter, error) {
	filters, err := ParseEthernetFilterConfig(raw)
	if err != nil {
		return nil, err
	}

	for _, filter := range filters {
		if filter.Number == filterNumber {
			return &filter, nil
		}
	}

	return nil, fmt.Errorf("ethernet filter %d not found", filterNumber)
}

// BuildEthernetFilterCommand builds the command to create an Ethernet filter
// Command format: ethernet filter <n> <action> <src_mac> <dst_mac> [<eth_type>] [vlan <vlan_id>]
func BuildEthernetFilterCommand(filter EthernetFilter) string {
	srcMAC := filter.SourceMAC
	if srcMAC == "" {
		srcMAC = "*"
	}

	dstMAC := filter.DestMAC
	if dstMAC == "" {
		dstMAC = "*"
	}

	cmd := fmt.Sprintf("ethernet filter %d %s %s %s", filter.Number, filter.Action, srcMAC, dstMAC)

	// Add EtherType if specified
	if filter.EtherType != "" {
		cmd += fmt.Sprintf(" %s", filter.EtherType)
	}

	// Add VLAN ID if specified
	if filter.VlanID > 0 {
		cmd += fmt.Sprintf(" vlan %d", filter.VlanID)
	}

	return cmd
}

// BuildDeleteEthernetFilterCommand builds the command to delete an Ethernet filter
// Command format: no ethernet filter <n>
func BuildDeleteEthernetFilterCommand(number int) string {
	return fmt.Sprintf("no ethernet filter %d", number)
}

// BuildInterfaceEthernetFilterCommand builds the command to apply filters to an interface
// Command format: ethernet <interface> filter <direction> <filter_list>
func BuildInterfaceEthernetFilterCommand(iface string, direction string, filterNums []int) string {
	if len(filterNums) == 0 {
		return ""
	}

	// Convert filter numbers to strings
	filterStrs := make([]string, len(filterNums))
	for i, n := range filterNums {
		filterStrs[i] = strconv.Itoa(n)
	}

	return fmt.Sprintf("ethernet %s filter %s %s", iface, direction, strings.Join(filterStrs, " "))
}

// BuildDeleteInterfaceEthernetFilterCommand builds the command to remove filters from an interface
// Command format: no ethernet <interface> filter <direction>
func BuildDeleteInterfaceEthernetFilterCommand(iface string, direction string) string {
	return fmt.Sprintf("no ethernet %s filter %s", iface, direction)
}

// BuildShowEthernetFilterCommand builds the command to show Ethernet filter configuration
// Command format: show config | grep "ethernet filter <n>"
func BuildShowEthernetFilterCommand(number int) string {
	return fmt.Sprintf("show config | grep \"ethernet filter %d\"", number)
}

// BuildShowAllEthernetFiltersCommand builds the command to show all Ethernet filters
// Command format: show config | grep "ethernet filter"
func BuildShowAllEthernetFiltersCommand() string {
	return "show config | grep \"ethernet filter\""
}

// ValidateEthernetFilterNumber validates that the filter number is in valid range (1-65535)
func ValidateEthernetFilterNumber(n int) error {
	if n < 1 || n > 65535 {
		return fmt.Errorf("ethernet filter number must be between 1 and 65535, got %d", n)
	}
	return nil
}

// ValidateMAC validates a MAC address format
// Valid formats:
//   - 00:11:22:33:44:55 (colon-separated)
//   - 0011.2233.4455 (Cisco dot notation)
//   - 00-11-22-33-44-55 (hyphen-separated)
//   - 001122334455 (no separator)
//   - * (wildcard)
func ValidateMAC(mac string) error {
	mac = strings.TrimSpace(mac)

	// Wildcard is valid
	if mac == "*" {
		return nil
	}

	// Remove all separators
	clean := strings.ToLower(mac)
	clean = strings.ReplaceAll(clean, ":", "")
	clean = strings.ReplaceAll(clean, "-", "")
	clean = strings.ReplaceAll(clean, ".", "")

	// Check length
	if len(clean) != 12 {
		return fmt.Errorf("invalid MAC address format: %s (expected 12 hex characters)", mac)
	}

	// Check if all characters are hex
	for _, c := range clean {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return fmt.Errorf("invalid MAC address format: %s (non-hex character found)", mac)
		}
	}

	return nil
}

// ValidateEtherType validates an Ethernet type value
// Valid formats:
//   - 0x0800 (IPv4)
//   - 0x0806 (ARP)
//   - 0x86DD (IPv6)
//   - * (wildcard)
func ValidateEtherType(ethType string) error {
	ethType = strings.TrimSpace(ethType)

	// Empty or wildcard is valid
	if ethType == "" || ethType == "*" {
		return nil
	}

	// Must start with 0x
	if !strings.HasPrefix(strings.ToLower(ethType), "0x") {
		return fmt.Errorf("EtherType must be in hex format (e.g., 0x0800), got %s", ethType)
	}

	// Check hex value after 0x
	hexPart := ethType[2:]
	if len(hexPart) == 0 || len(hexPart) > 4 {
		return fmt.Errorf("EtherType must be 1-4 hex digits after 0x, got %s", ethType)
	}

	// Validate hex characters
	for _, c := range strings.ToLower(hexPart) {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return fmt.Errorf("invalid EtherType format: %s (non-hex character found)", ethType)
		}
	}

	return nil
}

// ValidateVlanID validates a VLAN ID (1-4094)
func ValidateVlanID(id int) error {
	// 0 means not specified, which is valid
	if id == 0 {
		return nil
	}

	if id < 1 || id > 4094 {
		return fmt.Errorf("VLAN ID must be between 1 and 4094, got %d", id)
	}
	return nil
}

// ValidateEthernetFilterAction validates the filter action (pass or reject)
func ValidateEthernetFilterAction(action string) error {
	action = strings.ToLower(strings.TrimSpace(action))
	if action != "pass" && action != "reject" {
		return fmt.Errorf("action must be 'pass' or 'reject', got '%s'", action)
	}
	return nil
}

// ValidateEthernetFilterDirection validates the filter direction (in or out)
func ValidateEthernetFilterDirection(direction string) error {
	direction = strings.ToLower(strings.TrimSpace(direction))
	if direction != "in" && direction != "out" {
		return fmt.Errorf("direction must be 'in' or 'out', got '%s'", direction)
	}
	return nil
}

// ValidateEthernetFilter validates an Ethernet filter configuration
func ValidateEthernetFilter(filter EthernetFilter) error {
	if err := ValidateEthernetFilterNumber(filter.Number); err != nil {
		return err
	}

	if err := ValidateEthernetFilterAction(filter.Action); err != nil {
		return err
	}

	if filter.SourceMAC != "" {
		if err := ValidateMAC(filter.SourceMAC); err != nil {
			return fmt.Errorf("invalid source MAC: %w", err)
		}
	}

	if filter.DestMAC != "" {
		if err := ValidateMAC(filter.DestMAC); err != nil {
			return fmt.Errorf("invalid destination MAC: %w", err)
		}
	}

	if filter.EtherType != "" {
		if err := ValidateEtherType(filter.EtherType); err != nil {
			return err
		}
	}

	if err := ValidateVlanID(filter.VlanID); err != nil {
		return err
	}

	return nil
}

// AccessListMACEntry represents a single entry in a MAC access list
type AccessListMACEntry struct {
	Sequence               int
	AceAction              string
	SourceAny              bool
	SourceAddress          string
	SourceAddressMask      string
	DestinationAny         bool
	DestinationAddress     string
	DestinationAddressMask string
	EtherType              string
	VlanID                 int
	Log                    bool
}

// BuildAccessListMACEntryCommand builds an Ethernet filter command from a MAC ACL entry
// RTX command: ethernet filter <sequence> <action> <src_mac> <dst_mac> [<eth_type>] [vlan <vlan_id>]
func BuildAccessListMACEntryCommand(entry AccessListMACEntry) string {
	// Build source MAC
	srcMAC := "*"
	if !entry.SourceAny && entry.SourceAddress != "" {
		srcMAC = entry.SourceAddress
		// Note: RTX doesn't support MAC masks in basic ethernet filter
		// We keep the mask for internal representation but don't use it in command
	}

	// Build destination MAC
	dstMAC := "*"
	if !entry.DestinationAny && entry.DestinationAddress != "" {
		dstMAC = entry.DestinationAddress
	}

	// Map action
	action := "pass"
	if entry.AceAction == "deny" {
		action = "reject"
	}

	// Build command
	cmd := fmt.Sprintf("ethernet filter %d %s %s %s", entry.Sequence, action, srcMAC, dstMAC)

	// Add EtherType if specified
	if entry.EtherType != "" {
		cmd += fmt.Sprintf(" %s", entry.EtherType)
	}

	// Add VLAN ID if specified
	if entry.VlanID > 0 {
		cmd += fmt.Sprintf(" vlan %d", entry.VlanID)
	}

	return cmd
}

// BuildMACAccessListInterfaceCommand builds the command to apply MAC filters to an interface
// Command format: ethernet <interface> filter <direction> <filter_numbers...>
func BuildMACAccessListInterfaceCommand(iface string, direction string, filterNums []int) string {
	if len(filterNums) == 0 {
		return ""
	}

	// Convert filter numbers to strings
	filterStrs := make([]string, len(filterNums))
	for i, n := range filterNums {
		filterStrs[i] = strconv.Itoa(n)
	}

	return fmt.Sprintf("ethernet %s filter %s %s", iface, direction, strings.Join(filterStrs, " "))
}

// BuildDeleteMACAccessListInterfaceCommand builds the command to remove MAC filters from an interface
// Command format: no ethernet <interface> filter <direction>
func BuildDeleteMACAccessListInterfaceCommand(iface string, direction string) string {
	return fmt.Sprintf("no ethernet %s filter %s", iface, direction)
}

// ParseInterfaceEthernetFilter parses the interface Ethernet filter configuration
// Returns a map of interface -> direction -> filter numbers
func ParseInterfaceEthernetFilter(raw string) (map[string]map[string][]int, error) {
	result := make(map[string]map[string][]int)
	lines := strings.Split(raw, "\n")

	// Pattern: ethernet <interface> filter <direction> <filter_numbers...>
	filterPattern := regexp.MustCompile(`^\s*ethernet\s+(\S+)\s+filter\s+(in|out)\s+(.+)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if matches := filterPattern.FindStringSubmatch(line); len(matches) >= 4 {
			iface := matches[1]
			direction := matches[2]
			filterPart := matches[3]

			if result[iface] == nil {
				result[iface] = make(map[string][]int)
			}

			// Parse filter numbers
			filterNums := []int{}
			parts := strings.Fields(filterPart)
			for _, part := range parts {
				num, err := strconv.Atoi(part)
				if err == nil {
					filterNums = append(filterNums, num)
				}
			}

			result[iface][direction] = filterNums
		}
	}

	return result, nil
}

// BuildShowInterfaceEthernetFilterCommand builds the command to show interface Ethernet filter configuration
func BuildShowInterfaceEthernetFilterCommand() string {
	return "show config | grep \"ethernet .* filter\""
}
