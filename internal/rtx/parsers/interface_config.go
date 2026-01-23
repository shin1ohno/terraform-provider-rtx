package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// InterfaceConfig represents interface configuration on an RTX router
type InterfaceConfig struct {
	Name              string       `json:"name"`                          // Interface name (lan1, lan2, pp1, bridge1, tunnel1)
	Description       string       `json:"description,omitempty"`         // Interface description
	IPAddress         *InterfaceIP `json:"ip_address,omitempty"`          // IPv4 address configuration
	SecureFilterIn    []int        `json:"secure_filter_in,omitempty"`    // Inbound security filter numbers
	SecureFilterOut   []int        `json:"secure_filter_out,omitempty"`   // Outbound security filter numbers
	DynamicFilterOut  []int        `json:"dynamic_filter_out,omitempty"`  // Dynamic filters for outbound
	EthernetFilterIn  []int        `json:"ethernet_filter_in,omitempty"`  // Inbound Ethernet (L2) filter numbers
	EthernetFilterOut []int        `json:"ethernet_filter_out,omitempty"` // Outbound Ethernet (L2) filter numbers
	NATDescriptor     int          `json:"nat_descriptor,omitempty"`      // NAT descriptor number (0 = none)
	ProxyARP          bool         `json:"proxyarp"`                      // Enable ProxyARP
	MTU               int          `json:"mtu,omitempty"`                 // MTU size (0 = default)
}

// InterfaceIP represents IP address configuration
type InterfaceIP struct {
	Address string `json:"address,omitempty"` // CIDR notation (192.168.1.1/24) or empty if DHCP
	DHCP    bool   `json:"dhcp"`              // Use DHCP for address assignment
}

// Interface name patterns for RTX routers
var (
	interfaceNamePattern = regexp.MustCompile(`^(lan|bridge|pp|tunnel)\d+$`)
)

// ParseInterfaceConfig parses the output of "show config | grep <interface>" command
// and returns the interface configuration
func ParseInterfaceConfig(raw string, interfaceName string) (*InterfaceConfig, error) {
	config := &InterfaceConfig{
		Name: interfaceName,
	}

	// Preprocess to handle wrapped lines (long filter lists can span multiple lines)
	raw = preprocessWrappedLines(raw)
	lines := strings.Split(raw, "\n")

	// Patterns for parsing interface configuration
	// description <interface> <desc> or description <interface> "desc with spaces"
	descPattern := regexp.MustCompile(`^\s*description\s+` + regexp.QuoteMeta(interfaceName) + `\s+(?:"([^"]+)"|(\S+))\s*$`)
	// ip <interface> address <ip>/<prefix> or ip <interface> address dhcp
	ipAddrPattern := regexp.MustCompile(`^\s*ip\s+` + regexp.QuoteMeta(interfaceName) + `\s+address\s+(\S+)\s*$`)
	// ip <interface> secure filter in <filter_list>
	filterInPattern := regexp.MustCompile(`^\s*ip\s+` + regexp.QuoteMeta(interfaceName) + `\s+secure\s+filter\s+in\s+(.+)\s*$`)
	// ip <interface> secure filter out <filter_list> [dynamic <dynamic_filter_list>]
	filterOutPattern := regexp.MustCompile(`^\s*ip\s+` + regexp.QuoteMeta(interfaceName) + `\s+secure\s+filter\s+out\s+(.+)\s*$`)
	// ethernet <interface> filter in <filter_list>
	ethFilterInPattern := regexp.MustCompile(`^\s*ethernet\s+` + regexp.QuoteMeta(interfaceName) + `\s+filter\s+in\s+(.+)\s*$`)
	// ethernet <interface> filter out <filter_list>
	ethFilterOutPattern := regexp.MustCompile(`^\s*ethernet\s+` + regexp.QuoteMeta(interfaceName) + `\s+filter\s+out\s+(.+)\s*$`)
	// ip <interface> nat descriptor <id>
	natPattern := regexp.MustCompile(`^\s*ip\s+` + regexp.QuoteMeta(interfaceName) + `\s+nat\s+descriptor\s+(\d+)\s*$`)
	// ip <interface> proxyarp on|off
	proxyarpPattern := regexp.MustCompile(`^\s*ip\s+` + regexp.QuoteMeta(interfaceName) + `\s+proxyarp\s+(on|off)\s*$`)
	// ip <interface> mtu <size>
	mtuPattern := regexp.MustCompile(`^\s*ip\s+` + regexp.QuoteMeta(interfaceName) + `\s+mtu\s+(\d+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse description
		if matches := descPattern.FindStringSubmatch(line); len(matches) >= 2 {
			if matches[1] != "" {
				config.Description = matches[1]
			} else if len(matches) > 2 && matches[2] != "" {
				config.Description = matches[2]
			}
			continue
		}

		// Parse IP address
		if matches := ipAddrPattern.FindStringSubmatch(line); len(matches) >= 2 {
			addr := matches[1]
			if strings.ToLower(addr) == "dhcp" {
				config.IPAddress = &InterfaceIP{DHCP: true}
			} else {
				config.IPAddress = &InterfaceIP{Address: addr}
			}
			continue
		}

		// Parse inbound security filter
		if matches := filterInPattern.FindStringSubmatch(line); len(matches) >= 2 {
			filters := parseFilterList(matches[1])
			config.SecureFilterIn = filters
			continue
		}

		// Parse outbound security filter (may include dynamic)
		if matches := filterOutPattern.FindStringSubmatch(line); len(matches) >= 2 {
			filterStr := matches[1]
			// Check for dynamic keyword
			dynamicIdx := strings.Index(strings.ToLower(filterStr), "dynamic")
			if dynamicIdx != -1 {
				staticPart := strings.TrimSpace(filterStr[:dynamicIdx])
				dynamicPart := strings.TrimSpace(filterStr[dynamicIdx+len("dynamic"):])
				config.SecureFilterOut = parseFilterList(staticPart)
				config.DynamicFilterOut = parseFilterList(dynamicPart)
			} else {
				config.SecureFilterOut = parseFilterList(filterStr)
			}
			continue
		}

		// Parse inbound Ethernet filter
		if matches := ethFilterInPattern.FindStringSubmatch(line); len(matches) >= 2 {
			filters := parseFilterList(matches[1])
			config.EthernetFilterIn = filters
			continue
		}

		// Parse outbound Ethernet filter
		if matches := ethFilterOutPattern.FindStringSubmatch(line); len(matches) >= 2 {
			filters := parseFilterList(matches[1])
			config.EthernetFilterOut = filters
			continue
		}

		// Parse NAT descriptor
		if matches := natPattern.FindStringSubmatch(line); len(matches) >= 2 {
			natID, err := strconv.Atoi(matches[1])
			if err == nil {
				config.NATDescriptor = natID
			}
			continue
		}

		// Parse ProxyARP
		if matches := proxyarpPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.ProxyARP = strings.ToLower(matches[1]) == "on"
			continue
		}

		// Parse MTU
		if matches := mtuPattern.FindStringSubmatch(line); len(matches) >= 2 {
			mtu, err := strconv.Atoi(matches[1])
			if err == nil {
				config.MTU = mtu
			}
			continue
		}
	}

	return config, nil
}

// parseFilterList parses a space-separated list of filter numbers
func parseFilterList(filterStr string) []int {
	var filters []int
	parts := strings.Fields(filterStr)
	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err == nil && num > 0 {
			filters = append(filters, num)
		}
	}
	return filters
}

// BuildIPAddressCommand builds the command to set IP address
// Command format: ip <interface> address <ip>/<prefix> or ip <interface> address dhcp
func BuildIPAddressCommand(iface string, ip InterfaceIP) string {
	if ip.DHCP {
		return fmt.Sprintf("ip %s address dhcp", iface)
	}
	if ip.Address != "" {
		return fmt.Sprintf("ip %s address %s", iface, ip.Address)
	}
	return ""
}

// BuildDeleteIPAddressCommand builds the command to remove IP address
// Command format: no ip <interface> address
func BuildDeleteIPAddressCommand(iface string) string {
	return fmt.Sprintf("no ip %s address", iface)
}

// BuildSecureFilterInCommand builds the command to set inbound security filter
// Command format: ip <interface> secure filter in <filter_list>
func BuildSecureFilterInCommand(iface string, filters []int) string {
	if len(filters) == 0 {
		return ""
	}
	filterStrs := make([]string, len(filters))
	for i, f := range filters {
		filterStrs[i] = strconv.Itoa(f)
	}
	return fmt.Sprintf("ip %s secure filter in %s", iface, strings.Join(filterStrs, " "))
}

// BuildSecureFilterOutCommand builds the command to set outbound security filter
// Command format: ip <interface> secure filter out <filter_list> [dynamic <dynamic_filter_list>]
func BuildSecureFilterOutCommand(iface string, filters []int, dynamicFilters []int) string {
	if len(filters) == 0 {
		return ""
	}
	filterStrs := make([]string, len(filters))
	for i, f := range filters {
		filterStrs[i] = strconv.Itoa(f)
	}

	cmd := fmt.Sprintf("ip %s secure filter out %s", iface, strings.Join(filterStrs, " "))

	if len(dynamicFilters) > 0 {
		dynamicStrs := make([]string, len(dynamicFilters))
		for i, f := range dynamicFilters {
			dynamicStrs[i] = strconv.Itoa(f)
		}
		cmd += " dynamic " + strings.Join(dynamicStrs, " ")
	}

	return cmd
}

// BuildDeleteSecureFilterCommand builds the command to remove security filter
// Command format: no ip <interface> secure filter in|out
func BuildDeleteSecureFilterCommand(iface string, direction string) string {
	return fmt.Sprintf("no ip %s secure filter %s", iface, direction)
}

// BuildNATDescriptorCommand builds the command to set NAT descriptor
// Command format: ip <interface> nat descriptor <id>
func BuildNATDescriptorCommand(iface string, natID int) string {
	if natID <= 0 {
		return ""
	}
	return fmt.Sprintf("ip %s nat descriptor %d", iface, natID)
}

// BuildDeleteNATDescriptorCommand builds the command to remove NAT descriptor
// Command format: no ip <interface> nat descriptor
func BuildDeleteNATDescriptorCommand(iface string) string {
	return fmt.Sprintf("no ip %s nat descriptor", iface)
}

// BuildProxyARPCommand builds the command to set ProxyARP
// Command format: ip <interface> proxyarp on|off
func BuildProxyARPCommand(iface string, enabled bool) string {
	state := "off"
	if enabled {
		state = "on"
	}
	return fmt.Sprintf("ip %s proxyarp %s", iface, state)
}

// BuildDescriptionCommand builds the command to set interface description
// Command format: description <interface> "<description>"
func BuildDescriptionCommand(iface string, desc string) string {
	if desc == "" {
		return ""
	}
	return fmt.Sprintf(`description %s "%s"`, iface, desc)
}

// BuildDeleteDescriptionCommand builds the command to remove interface description
// Command format: no description <interface>
func BuildDeleteDescriptionCommand(iface string) string {
	return fmt.Sprintf("no description %s", iface)
}

// BuildMTUCommand builds the command to set MTU
// Command format: ip <interface> mtu <size>
func BuildMTUCommand(iface string, mtu int) string {
	if mtu <= 0 {
		return ""
	}
	return fmt.Sprintf("ip %s mtu %d", iface, mtu)
}

// BuildDeleteMTUCommand builds the command to remove MTU configuration
// Command format: no ip <interface> mtu
func BuildDeleteMTUCommand(iface string) string {
	return fmt.Sprintf("no ip %s mtu", iface)
}

// BuildShowInterfaceConfigCommand builds the command to show interface configuration
// Command format: show config | grep "<interface>"
func BuildShowInterfaceConfigCommand(interfaceName string) string {
	return fmt.Sprintf(`show config | grep "%s"`, interfaceName)
}

// ValidateInterfaceConfig validates an interface configuration
func ValidateInterfaceConfig(config InterfaceConfig) error {
	// Validate interface name
	if err := ValidateInterfaceName(config.Name); err != nil {
		return err
	}

	// Validate IP address configuration
	if config.IPAddress != nil {
		if config.IPAddress.DHCP && config.IPAddress.Address != "" {
			return fmt.Errorf("cannot specify both DHCP and static IP address")
		}
		if config.IPAddress.Address != "" && !isValidCIDR(config.IPAddress.Address) {
			return fmt.Errorf("IP address must be in CIDR notation (e.g., 192.168.1.1/24)")
		}
	}

	// Validate filter numbers
	for _, f := range config.SecureFilterIn {
		if f <= 0 {
			return fmt.Errorf("filter numbers must be positive integers")
		}
	}
	for _, f := range config.SecureFilterOut {
		if f <= 0 {
			return fmt.Errorf("filter numbers must be positive integers")
		}
	}
	for _, f := range config.DynamicFilterOut {
		if f <= 0 {
			return fmt.Errorf("filter numbers must be positive integers")
		}
	}

	// Validate Ethernet filter numbers
	for _, f := range config.EthernetFilterIn {
		if f <= 0 {
			return fmt.Errorf("ethernet filter numbers must be positive integers")
		}
	}
	for _, f := range config.EthernetFilterOut {
		if f <= 0 {
			return fmt.Errorf("ethernet filter numbers must be positive integers")
		}
	}

	// Validate NAT descriptor
	if config.NATDescriptor < 0 {
		return fmt.Errorf("NAT descriptor must be non-negative")
	}

	// Validate MTU
	if config.MTU != 0 && (config.MTU < 68 || config.MTU > 65535) {
		return fmt.Errorf("MTU must be between 68 and 65535")
	}

	return nil
}

// ValidateInterfaceName validates an interface name
func ValidateInterfaceName(name string) error {
	if name == "" {
		return fmt.Errorf("interface name is required")
	}
	if !interfaceNamePattern.MatchString(name) {
		return fmt.Errorf("invalid interface name: %s (expected lan1, lan2, bridge1, pp1, tunnel1, etc.)", name)
	}
	return nil
}

// endsWithDigit returns true if the string ends with a digit (0-9)
// Note: This checks the ORIGINAL string without trimming, because trailing whitespace
// indicates the number was complete (not a mid-number wrap scenario)
func endsWithDigit(s string) bool {
	if len(s) == 0 {
		return false
	}
	return s[len(s)-1] >= '0' && s[len(s)-1] <= '9'
}

// startsWithDigit returns true if the string starts with a digit (0-9)
func startsWithDigit(s string) bool {
	if len(s) == 0 {
		return false
	}
	return s[0] >= '0' && s[0] <= '9'
}

// preprocessWrappedLines handles RTX output where long filter lists wrap to multiple lines.
// RTX wraps long filter lists by continuing on the next line with just numbers.
// This function joins those continuation lines back together.
//
// IMPORTANT: RTX terminal wrapping can split numbers at arbitrary positions.
// When a line ends with a digit and the ORIGINAL continuation line starts
// directly with a digit (no leading whitespace), this indicates a mid-number wrap,
// and we join WITHOUT a space to preserve the number.
//
// Example input (mid-number wrap - no leading space on continuation):
//
//	ip lan2 secure filter in 200020 20010
//	0 200102
//
// Example output:
//
//	ip lan2 secure filter in 200020 200100 200102
//
// Example input (normal wrap - has leading space on continuation):
//
//	ip lan2 secure filter in 200020 200021
//	 200022 200023
//
// Example output:
//
//	ip lan2 secure filter in 200020 200021 200022 200023
func preprocessWrappedLines(input string) string {
	if input == "" {
		return ""
	}

	// Normalize line endings
	input = strings.ReplaceAll(input, "\r\n", "\n")

	// Pattern to detect a continuation line:
	// - RTX wraps long lines at ~80 chars, continuation starts with space(s) then digit
	// - Also handle lines that start directly with a digit (no leading space)
	// - May contain numbers, 'dynamic' keyword, and spaces
	// This handles both simple number continuations and "numbers dynamic numbers" patterns
	continuationPattern := regexp.MustCompile(`^(\s+)?\d`)

	lines := strings.Split(input, "\n")
	var result []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Look ahead for continuation lines (lines starting with a digit)
		for i+1 < len(lines) && continuationPattern.MatchString(strings.TrimSpace(lines[i+1])) {
			i++
			rawNextLine := lines[i]
			nextLine := strings.TrimSpace(rawNextLine)

			// Check if we need to join without space (mid-number wrap)
			// Mid-number wrap occurs when:
			// 1. Current line ends with a digit
			// 2. Original continuation line starts DIRECTLY with a digit (no leading whitespace)
			// This distinguishes between:
			// - "... 20010\n0 ..." -> mid-number wrap (join without space)
			// - "... 200021\n 200022 ..." -> normal wrap (join with space)
			if endsWithDigit(line) && startsWithDigit(rawNextLine) {
				line = line + nextLine // Join without space to preserve the number
			} else {
				line = line + " " + nextLine // Join with space
			}
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}
