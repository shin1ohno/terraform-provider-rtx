package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// EthernetFilter represents an Ethernet (Layer 2) filter configuration on an RTX router
type EthernetFilter struct {
	Number         int      `json:"number"`                    // Filter number (1-512)
	Action         string   `json:"action"`                    // pass-log, pass-nolog, reject-log, reject-nolog
	SourceMAC      string   `json:"source_mac,omitempty"`      // Source MAC address (* for any)
	DestinationMAC string   `json:"destination_mac,omitempty"` // Destination MAC address (* for any)
	DestMAC        string   `json:"dest_mac,omitempty"`        // Deprecated: Use DestinationMAC instead
	EtherType      string   `json:"ether_type,omitempty"`      // Ethernet type (e.g., 0x0800, 0x0806)
	VlanID         int      `json:"vlan_id,omitempty"`         // VLAN ID (1-4094, 0 means not specified)
	DHCPType       string   `json:"dhcp_type,omitempty"`       // DHCP filter type: dhcp-bind or dhcp-not-bind
	DHCPScope      int      `json:"dhcp_scope,omitempty"`      // DHCP scope number (for DHCP-based filters)
	Offset         int      `json:"offset,omitempty"`          // Byte offset for byte-match filtering
	ByteList       []string `json:"byte_list,omitempty"`       // Byte patterns for byte-match filtering
}

// ValidEthernetFilterActions defines the valid actions for Ethernet filters
var ValidEthernetFilterActions = []string{"pass-log", "pass-nolog", "reject-log", "reject-nolog", "pass", "reject"}

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
// Handles both MAC-based and DHCP-based filter formats:
// - MAC-based: ethernet filter <id> <action> <src_mac> [<dst_mac>] [offset=N byte1 byte2 ...]
// - DHCP-based: ethernet filter <id> <action> dhcp-bind|dhcp-not-bind [<scope>]
func ParseEthernetFilterConfig(raw string) ([]EthernetFilter, error) {
	filters := []EthernetFilter{}
	lines := strings.Split(raw, "\n")

	// Pattern for MAC-based filter:
	// ethernet filter <n> <action> <src_mac> <dst_mac> [<eth_type>] [vlan <vlan_id>]
	// Note: Use .*$ instead of \s*$ to handle RTX terminal line wrapping
	macFilterPattern := regexp.MustCompile(`^\s*ethernet\s+filter\s+(\d+)\s+(pass-log|pass-nolog|reject-log|reject-nolog|pass|reject)\s+(\S+)\s+(\S+)(?:\s+(0x[0-9a-fA-F]+|\*))?(?:\s+vlan\s+(\d+))?.*$`)

	// Pattern for DHCP-based filter:
	// ethernet filter <n> <action> dhcp-bind|dhcp-not-bind [<scope>]
	dhcpFilterPattern := regexp.MustCompile(`^\s*ethernet\s+filter\s+(\d+)\s+(pass-log|pass-nolog|reject-log|reject-nolog|pass|reject)\s+(dhcp-bind|dhcp-not-bind)(?:\s+(\d+))?.*$`)

	// Pattern for filter with offset and byte_list:
	// ethernet filter <n> <action> <src_mac> [<dst_mac>] offset=<N> <byte1> <byte2> ...
	offsetFilterPattern := regexp.MustCompile(`^\s*ethernet\s+filter\s+(\d+)\s+(pass-log|pass-nolog|reject-log|reject-nolog|pass|reject)\s+(\S+)(?:\s+(\S+))?\s+offset=(\d+)\s+(.+)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try offset filter pattern first (most specific)
		if matches := offsetFilterPattern.FindStringSubmatch(line); len(matches) >= 7 {
			number, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}
			offset, err := strconv.Atoi(matches[5])
			if err != nil {
				continue
			}

			filter := EthernetFilter{
				Number:         number,
				Action:         matches[2],
				SourceMAC:      matches[3],
				DestinationMAC: matches[4],
				Offset:         offset,
				ByteList:       strings.Fields(matches[6]),
			}

			// For backward compatibility
			filter.DestMAC = filter.DestinationMAC

			filters = append(filters, filter)
			continue
		}

		// Try DHCP-based filter pattern
		if matches := dhcpFilterPattern.FindStringSubmatch(line); len(matches) >= 4 {
			number, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			filter := EthernetFilter{
				Number:   number,
				Action:   matches[2],
				DHCPType: matches[3],
			}

			// DHCP scope (optional)
			if len(matches) > 4 && matches[4] != "" {
				scope, err := strconv.Atoi(matches[4])
				if err == nil {
					filter.DHCPScope = scope
				}
			}

			filters = append(filters, filter)
			continue
		}

		// Try MAC-based filter pattern
		if matches := macFilterPattern.FindStringSubmatch(line); len(matches) >= 5 {
			number, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			filter := EthernetFilter{
				Number:         number,
				Action:         matches[2],
				SourceMAC:      matches[3],
				DestinationMAC: matches[4],
			}

			// For backward compatibility
			filter.DestMAC = filter.DestinationMAC

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
// Handles both MAC-based and DHCP-based filter formats:
// - MAC-based: ethernet filter <n> <action> <src_mac> [<dst_mac>] [<eth_type>] [vlan <vlan_id>]
// - DHCP-based: ethernet filter <n> <action> dhcp-bind|dhcp-not-bind [<scope>]
// - With offset: ethernet filter <n> <action> <src_mac> [<dst_mac>] offset=<N> <byte1> <byte2> ...
func BuildEthernetFilterCommand(filter EthernetFilter) string {
	// DHCP-based filter
	if filter.DHCPType != "" {
		cmd := fmt.Sprintf("ethernet filter %d %s %s", filter.Number, filter.Action, filter.DHCPType)
		if filter.DHCPScope > 0 {
			cmd += fmt.Sprintf(" %d", filter.DHCPScope)
		}
		return cmd
	}

	// MAC-based filter
	srcMAC := filter.SourceMAC
	if srcMAC == "" {
		srcMAC = "*"
	}

	// Use DestinationMAC if set, fall back to DestMAC for backward compatibility
	dstMAC := filter.DestinationMAC
	if dstMAC == "" {
		dstMAC = filter.DestMAC
	}
	if dstMAC == "" {
		dstMAC = "*"
	}

	cmd := fmt.Sprintf("ethernet filter %d %s %s %s", filter.Number, filter.Action, srcMAC, dstMAC)

	// Add offset and byte_list if specified
	if filter.Offset > 0 && len(filter.ByteList) > 0 {
		cmd += fmt.Sprintf(" offset=%d %s", filter.Offset, strings.Join(filter.ByteList, " "))
		return cmd
	}

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

// ValidateEthernetFilterNumber validates that the filter number is in valid range (1-512)
func ValidateEthernetFilterNumber(n int) error {
	if n < 1 || n > 512 {
		return fmt.Errorf("ethernet filter number must be between 1 and 512, got %d", n)
	}
	return nil
}

// ValidateMACAddress validates a MAC address format for ethernet filters
// Valid formats:
//   - xx:xx:xx:xx:xx:xx (colon-separated)
//   - * (wildcard)
func ValidateMACAddress(mac string) error {
	mac = strings.TrimSpace(mac)

	// Empty is valid (will default to *)
	if mac == "" {
		return nil
	}

	// Wildcard is valid
	if mac == "*" {
		return nil
	}

	// Pattern for colon-separated MAC address: xx:xx:xx:xx:xx:xx
	macPattern := regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$`)
	if !macPattern.MatchString(mac) {
		return fmt.Errorf("invalid MAC address format: %s (expected xx:xx:xx:xx:xx:xx or *)", mac)
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
//   - *:*:*:*:*:* (RTX wildcard format)
func ValidateMAC(mac string) error {
	mac = strings.TrimSpace(mac)

	// Wildcards are valid - both "*" and "*:*:*:*:*:*" formats
	if mac == "*" || mac == "*:*:*:*:*:*" {
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

// ValidateEthernetFilterAction validates the filter action
// Valid actions: pass-log, pass-nolog, reject-log, reject-nolog, pass, reject
func ValidateEthernetFilterAction(action string) error {
	action = strings.ToLower(strings.TrimSpace(action))
	for _, valid := range ValidEthernetFilterActions {
		if action == valid {
			return nil
		}
	}
	return fmt.Errorf("action must be one of %v, got '%s'", ValidEthernetFilterActions, action)
}

// ValidateDHCPType validates the DHCP filter type
// Valid types: dhcp-bind, dhcp-not-bind
func ValidateDHCPType(dhcpType string) error {
	dhcpType = strings.ToLower(strings.TrimSpace(dhcpType))
	if dhcpType == "" {
		return nil // Empty is valid for non-DHCP filters
	}
	if dhcpType != "dhcp-bind" && dhcpType != "dhcp-not-bind" {
		return fmt.Errorf("dhcp_type must be 'dhcp-bind' or 'dhcp-not-bind', got '%s'", dhcpType)
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

	// DHCP-based filter validation
	if filter.DHCPType != "" {
		if err := ValidateDHCPType(filter.DHCPType); err != nil {
			return err
		}
		// DHCP filters should not have MAC addresses set
		if filter.SourceMAC != "" && filter.SourceMAC != "*" {
			return fmt.Errorf("DHCP-based filter should not have source MAC address")
		}
		if (filter.DestinationMAC != "" && filter.DestinationMAC != "*") ||
			(filter.DestMAC != "" && filter.DestMAC != "*") {
			return fmt.Errorf("DHCP-based filter should not have destination MAC address")
		}
		return nil
	}

	// MAC-based filter validation
	if filter.SourceMAC != "" {
		if err := ValidateMAC(filter.SourceMAC); err != nil {
			return fmt.Errorf("invalid source MAC: %w", err)
		}
	}

	// Check both DestinationMAC and DestMAC for backward compatibility
	destMAC := filter.DestinationMAC
	if destMAC == "" {
		destMAC = filter.DestMAC
	}
	if destMAC != "" {
		if err := ValidateMAC(destMAC); err != nil {
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

	// Validate offset and byte_list consistency
	if filter.Offset > 0 && len(filter.ByteList) == 0 {
		return fmt.Errorf("byte_list is required when offset is specified")
	}
	if filter.Offset == 0 && len(filter.ByteList) > 0 {
		return fmt.Errorf("offset is required when byte_list is specified")
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
	FilterID               int
	DHCPType               string
	DHCPScope              int
	Offset                 int
	ByteList               []string
}

// BuildAccessListMACEntryCommand builds an Ethernet filter command from a MAC ACL entry
// RTX command: ethernet filter <sequence> <action> <src_mac> <dst_mac> [<eth_type>] [vlan <vlan_id>]
func BuildAccessListMACEntryCommand(entry AccessListMACEntry) string {
	number := entry.Sequence
	if entry.FilterID > 0 {
		number = entry.FilterID
	}

	action := entry.AceAction
	switch entry.AceAction {
	case "permit":
		action = "pass"
	case "deny":
		action = "reject"
	}

	filter := EthernetFilter{
		Number:         number,
		Action:         action,
		SourceMAC:      "*",
		DestinationMAC: "*",
		EtherType:      entry.EtherType,
		VlanID:         entry.VlanID,
		DHCPType:       entry.DHCPType,
		DHCPScope:      entry.DHCPScope,
		Offset:         entry.Offset,
		ByteList:       entry.ByteList,
	}

	if !entry.SourceAny && entry.SourceAddress != "" {
		filter.SourceMAC = entry.SourceAddress
	}
	if !entry.DestinationAny && entry.DestinationAddress != "" {
		filter.DestinationMAC = entry.DestinationAddress
		filter.DestMAC = entry.DestinationAddress
	}

	return BuildEthernetFilterCommand(filter)
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

// EthernetFilterApplication represents ethernet filter binding to an interface
// This struct provides a cleaner interface for working with ethernet filter application
// to specific interfaces compared to the map-based ParseInterfaceEthernetFilter function.
type EthernetFilterApplication struct {
	Interface string `json:"interface"` // lan1, lan2, etc.
	Direction string `json:"direction"` // in, out
	Filters   []int  `json:"filters"`   // Filter numbers in order
}

// ParseEthernetFilterApplication parses "ethernet <interface> filter in/out" commands
// and returns a list of EthernetFilterApplication structs.
// This is an alternative to ParseInterfaceEthernetFilter that returns a struct-based result.
//
// Example input:
//
//	ethernet lan1 filter in 1 100
//	ethernet lan1 filter out 2 200
//	ethernet lan2 filter in 10 20 30
func ParseEthernetFilterApplication(raw string) ([]EthernetFilterApplication, error) {
	result := []EthernetFilterApplication{}
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

			// Parse filter numbers
			filterNums := []int{}
			parts := strings.Fields(filterPart)
			for _, part := range parts {
				num, err := strconv.Atoi(part)
				if err == nil {
					filterNums = append(filterNums, num)
				}
			}

			// Only add if we have at least one valid filter number
			if len(filterNums) > 0 {
				result = append(result, EthernetFilterApplication{
					Interface: iface,
					Direction: direction,
					Filters:   filterNums,
				})
			}
		}
	}

	return result, nil
}

// ParseSingleEthernetFilterApplication parses the configuration and returns
// the filter application for a specific interface and direction.
// Returns nil if no matching configuration is found.
func ParseSingleEthernetFilterApplication(raw string, iface string, direction string) (*EthernetFilterApplication, error) {
	apps, err := ParseEthernetFilterApplication(raw)
	if err != nil {
		return nil, err
	}

	for _, app := range apps {
		if app.Interface == iface && app.Direction == direction {
			return &app, nil
		}
	}

	return nil, nil // Not found, but not an error
}

// BuildEthernetFilterApplicationCommand generates the CLI command from an EthernetFilterApplication struct
// Command format: ethernet <interface> filter <direction> <filter_numbers...>
// Returns empty string if Filters is empty.
func BuildEthernetFilterApplicationCommand(app EthernetFilterApplication) string {
	if len(app.Filters) == 0 {
		return ""
	}

	// Convert filter numbers to strings
	filterStrs := make([]string, len(app.Filters))
	for i, n := range app.Filters {
		filterStrs[i] = strconv.Itoa(n)
	}

	return fmt.Sprintf("ethernet %s filter %s %s", app.Interface, app.Direction, strings.Join(filterStrs, " "))
}

// BuildDeleteEthernetFilterApplicationCommand generates the CLI command to remove filters from an interface
// Command format: no ethernet <interface> filter <direction>
func BuildDeleteEthernetFilterApplicationCommand(iface string, direction string) string {
	return fmt.Sprintf("no ethernet %s filter %s", iface, direction)
}

// ValidateEthernetFilterApplication validates an EthernetFilterApplication struct
func ValidateEthernetFilterApplication(app EthernetFilterApplication) error {
	// Validate interface name - should be a valid interface like lan1, lan2, etc.
	if app.Interface == "" {
		return fmt.Errorf("interface name is required")
	}

	// Validate direction
	if err := ValidateEthernetFilterDirection(app.Direction); err != nil {
		return err
	}

	// Validate filter numbers
	for _, num := range app.Filters {
		if err := ValidateEthernetFilterNumber(num); err != nil {
			return fmt.Errorf("invalid filter number %d: %w", num, err)
		}
	}

	return nil
}
