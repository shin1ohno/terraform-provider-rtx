package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// IPv6InterfaceConfig represents IPv6 configuration for an RTX router interface
type IPv6InterfaceConfig struct {
	Interface        string        `json:"interface"`                    // Interface name (lan1, lan2, pp1, bridge1, tunnel1)
	Addresses        []IPv6Address `json:"addresses,omitempty"`          // IPv6 addresses
	RTADV            *RTADVConfig  `json:"rtadv,omitempty"`              // Router Advertisement configuration
	DHCPv6Service    string        `json:"dhcpv6_service,omitempty"`     // "server", "client", or "off"
	MTU              int           `json:"mtu,omitempty"`                // MTU size (0 = default)
	SecureFilterIn   []int         `json:"secure_filter_in,omitempty"`   // Inbound security filter numbers
	SecureFilterOut  []int         `json:"secure_filter_out,omitempty"`  // Outbound security filter numbers
	DynamicFilterOut []int         `json:"dynamic_filter_out,omitempty"` // Dynamic filters for outbound
}

// IPv6Address represents an IPv6 address configuration
type IPv6Address struct {
	Address     string `json:"address,omitempty"`      // Full IPv6 address with prefix (e.g., "2001:db8::1/64")
	PrefixRef   string `json:"prefix_ref,omitempty"`   // Prefix reference (e.g., "ra-prefix@lan2")
	InterfaceID string `json:"interface_id,omitempty"` // Interface ID (e.g., "::1/64")
}

// RTADVConfig represents Router Advertisement configuration
type RTADVConfig struct {
	Enabled  bool `json:"enabled"`            // RTADV enabled
	PrefixID int  `json:"prefix_id"`          // Prefix ID to advertise
	OFlag    bool `json:"o_flag"`             // Other Configuration Flag (O flag)
	MFlag    bool `json:"m_flag"`             // Managed Address Configuration Flag (M flag)
	Lifetime int  `json:"lifetime,omitempty"` // Router lifetime in seconds
}

// IPv6 interface name patterns for RTX routers
var (
	ipv6InterfaceNamePattern = regexp.MustCompile(`^(lan|bridge|pp|tunnel)\d+$`)
)

// ParseIPv6InterfaceConfig parses the output of "show config | grep <interface>" command
// and returns the IPv6 interface configuration
func ParseIPv6InterfaceConfig(raw string, interfaceName string) (*IPv6InterfaceConfig, error) {
	config := &IPv6InterfaceConfig{
		Interface: interfaceName,
		Addresses: []IPv6Address{},
	}

	// Preprocess to handle wrapped lines (long filter lists can span multiple lines)
	raw = preprocessWrappedLines(raw)
	lines := strings.Split(raw, "\n")

	// Patterns for parsing IPv6 interface configuration
	// ipv6 <interface> address <address>/<prefix>
	// ipv6 <interface> address <prefix-ref>::<interface-id>/<prefix>
	ipv6AddrPattern := regexp.MustCompile(`^\s*ipv6\s+` + regexp.QuoteMeta(interfaceName) + `\s+address\s+(\S+)\s*$`)
	// ipv6 <interface> rtadv send <prefix_id> [o_flag=on|off] [m_flag=on|off] [lifetime=<seconds>]
	rtadvPattern := regexp.MustCompile(`^\s*ipv6\s+` + regexp.QuoteMeta(interfaceName) + `\s+rtadv\s+send\s+(.+)\s*$`)
	// ipv6 <interface> dhcp service server|client
	dhcpPattern := regexp.MustCompile(`^\s*ipv6\s+` + regexp.QuoteMeta(interfaceName) + `\s+dhcp\s+service\s+(server|client)\s*$`)
	// ipv6 <interface> mtu <size>
	mtuPattern := regexp.MustCompile(`^\s*ipv6\s+` + regexp.QuoteMeta(interfaceName) + `\s+mtu\s+(\d+)\s*$`)
	// ipv6 <interface> secure filter in <filter_list>
	filterInPattern := regexp.MustCompile(`^\s*ipv6\s+` + regexp.QuoteMeta(interfaceName) + `\s+secure\s+filter\s+in\s+(.+)\s*$`)
	// ipv6 <interface> secure filter out <filter_list> [dynamic <dynamic_filter_list>]
	filterOutPattern := regexp.MustCompile(`^\s*ipv6\s+` + regexp.QuoteMeta(interfaceName) + `\s+secure\s+filter\s+out\s+(.+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse IPv6 address
		if matches := ipv6AddrPattern.FindStringSubmatch(line); len(matches) >= 2 {
			addr := parseIPv6Address(matches[1])
			config.Addresses = append(config.Addresses, addr)
			continue
		}

		// Parse RTADV configuration
		if matches := rtadvPattern.FindStringSubmatch(line); len(matches) >= 2 {
			rtadv := parseRTADVConfig(matches[1])
			config.RTADV = rtadv
			continue
		}

		// Parse DHCPv6 service
		if matches := dhcpPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.DHCPv6Service = matches[1]
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
	}

	return config, nil
}

// parseIPv6Address parses an IPv6 address string
// Handles formats like:
// - "2001:db8::1/64" (static address)
// - "ra-prefix@lan2::1/64" (prefix-based address)
// - "dhcp-prefix@lan2::1/64" (DHCPv6-PD based address)
func parseIPv6Address(addrStr string) IPv6Address {
	addr := IPv6Address{}

	// Check for prefix reference (contains @)
	if strings.Contains(addrStr, "@") {
		// Format: prefix-ref::interface-id/prefix
		// e.g., "ra-prefix@lan2::1/64"
		parts := strings.SplitN(addrStr, "::", 2)
		if len(parts) == 2 {
			addr.PrefixRef = parts[0]
			addr.InterfaceID = "::" + parts[1]
		} else {
			addr.Address = addrStr
		}
	} else {
		// Static address
		addr.Address = addrStr
	}

	return addr
}

// parseRTADVConfig parses RTADV configuration string
// Format: <prefix_id> [o_flag=on|off] [m_flag=on|off] [lifetime=<seconds>]
func parseRTADVConfig(rtadvStr string) *RTADVConfig {
	rtadv := &RTADVConfig{
		Enabled: true,
	}

	parts := strings.Fields(rtadvStr)
	if len(parts) == 0 {
		return nil
	}

	// First part is the prefix ID
	prefixID, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil
	}
	rtadv.PrefixID = prefixID

	// Parse optional flags
	for _, part := range parts[1:] {
		if strings.HasPrefix(strings.ToLower(part), "o_flag=") {
			value := strings.TrimPrefix(strings.ToLower(part), "o_flag=")
			rtadv.OFlag = value == "on"
		} else if strings.HasPrefix(strings.ToLower(part), "m_flag=") {
			value := strings.TrimPrefix(strings.ToLower(part), "m_flag=")
			rtadv.MFlag = value == "on"
		} else if strings.HasPrefix(strings.ToLower(part), "lifetime=") {
			value := strings.TrimPrefix(part, "lifetime=")
			lifetime, err := strconv.Atoi(value)
			if err == nil {
				rtadv.Lifetime = lifetime
			}
		}
	}

	return rtadv
}

// BuildIPv6AddressCommand builds the command to set IPv6 address
// Command format: ipv6 <interface> address <address>
func BuildIPv6AddressCommand(iface string, addr IPv6Address) string {
	if addr.PrefixRef != "" && addr.InterfaceID != "" {
		// Prefix-based address: ipv6 lan1 address ra-prefix@lan2::1/64
		return fmt.Sprintf("ipv6 %s address %s%s", iface, addr.PrefixRef, addr.InterfaceID)
	}
	if addr.Address != "" {
		return fmt.Sprintf("ipv6 %s address %s", iface, addr.Address)
	}
	return ""
}

// BuildDeleteIPv6AddressCommand builds the command to remove IPv6 address
// Command format: no ipv6 <interface> address [<address>]
func BuildDeleteIPv6AddressCommand(iface string, addr *IPv6Address) string {
	if addr == nil {
		// Remove all addresses
		return fmt.Sprintf("no ipv6 %s address", iface)
	}
	if addr.PrefixRef != "" && addr.InterfaceID != "" {
		return fmt.Sprintf("no ipv6 %s address %s%s", iface, addr.PrefixRef, addr.InterfaceID)
	}
	if addr.Address != "" {
		return fmt.Sprintf("no ipv6 %s address %s", iface, addr.Address)
	}
	return fmt.Sprintf("no ipv6 %s address", iface)
}

// BuildIPv6RTADVCommand builds the command to configure Router Advertisement
// Command format: ipv6 <interface> rtadv send <prefix_id> [o_flag=on|off] [m_flag=on|off] [lifetime=<seconds>]
func BuildIPv6RTADVCommand(iface string, rtadv RTADVConfig) string {
	if !rtadv.Enabled {
		return ""
	}

	cmd := fmt.Sprintf("ipv6 %s rtadv send %d", iface, rtadv.PrefixID)

	if rtadv.OFlag {
		cmd += " o_flag=on"
	} else {
		cmd += " o_flag=off"
	}

	if rtadv.MFlag {
		cmd += " m_flag=on"
	} else {
		cmd += " m_flag=off"
	}

	if rtadv.Lifetime > 0 {
		cmd += fmt.Sprintf(" lifetime=%d", rtadv.Lifetime)
	}

	return cmd
}

// BuildDeleteIPv6RTADVCommand builds the command to remove Router Advertisement
// Command format: no ipv6 <interface> rtadv send
func BuildDeleteIPv6RTADVCommand(iface string) string {
	return fmt.Sprintf("no ipv6 %s rtadv send", iface)
}

// BuildIPv6DHCPv6Command builds the command to configure DHCPv6 service
// Command format: ipv6 <interface> dhcp service server|client
func BuildIPv6DHCPv6Command(iface string, service string) string {
	if service == "" || service == "off" {
		return ""
	}
	return fmt.Sprintf("ipv6 %s dhcp service %s", iface, service)
}

// BuildDeleteIPv6DHCPv6Command builds the command to remove DHCPv6 service
// Command format: no ipv6 <interface> dhcp service
func BuildDeleteIPv6DHCPv6Command(iface string) string {
	return fmt.Sprintf("no ipv6 %s dhcp service", iface)
}

// BuildIPv6MTUCommand builds the command to set IPv6 MTU
// Command format: ipv6 <interface> mtu <size>
func BuildIPv6MTUCommand(iface string, mtu int) string {
	if mtu <= 0 {
		return ""
	}
	return fmt.Sprintf("ipv6 %s mtu %d", iface, mtu)
}

// BuildDeleteIPv6MTUCommand builds the command to remove IPv6 MTU configuration
// Command format: no ipv6 <interface> mtu
func BuildDeleteIPv6MTUCommand(iface string) string {
	return fmt.Sprintf("no ipv6 %s mtu", iface)
}

// BuildIPv6SecureFilterInCommand builds the command to set inbound IPv6 security filter
// Command format: ipv6 <interface> secure filter in <filter_list>
func BuildIPv6SecureFilterInCommand(iface string, filters []int) string {
	if len(filters) == 0 {
		return ""
	}
	filterStrs := make([]string, len(filters))
	for i, f := range filters {
		filterStrs[i] = strconv.Itoa(f)
	}
	return fmt.Sprintf("ipv6 %s secure filter in %s", iface, strings.Join(filterStrs, " "))
}

// BuildIPv6SecureFilterOutCommand builds the command to set outbound IPv6 security filter
// Command format: ipv6 <interface> secure filter out <filter_list> [dynamic <dynamic_filter_list>]
func BuildIPv6SecureFilterOutCommand(iface string, filters []int, dynamicFilters []int) string {
	if len(filters) == 0 {
		return ""
	}
	filterStrs := make([]string, len(filters))
	for i, f := range filters {
		filterStrs[i] = strconv.Itoa(f)
	}

	cmd := fmt.Sprintf("ipv6 %s secure filter out %s", iface, strings.Join(filterStrs, " "))

	if len(dynamicFilters) > 0 {
		dynamicStrs := make([]string, len(dynamicFilters))
		for i, f := range dynamicFilters {
			dynamicStrs[i] = strconv.Itoa(f)
		}
		cmd += " dynamic " + strings.Join(dynamicStrs, " ")
	}

	return cmd
}

// BuildDeleteIPv6SecureFilterCommand builds the command to remove IPv6 security filter
// Command format: no ipv6 <interface> secure filter in|out
func BuildDeleteIPv6SecureFilterCommand(iface string, direction string) string {
	return fmt.Sprintf("no ipv6 %s secure filter %s", iface, direction)
}

// BuildShowIPv6InterfaceConfigCommand builds the command to show IPv6 interface configuration
// Command format: show config | grep "ipv6 <interface>"
func BuildShowIPv6InterfaceConfigCommand(interfaceName string) string {
	return fmt.Sprintf(`show config | grep "ipv6 %s"`, interfaceName)
}

// BuildDeleteIPv6InterfaceCommands builds commands to remove all IPv6 interface configuration
func BuildDeleteIPv6InterfaceCommands(iface string) []string {
	return []string{
		fmt.Sprintf("no ipv6 %s address", iface),
		fmt.Sprintf("no ipv6 %s rtadv send", iface),
		fmt.Sprintf("no ipv6 %s dhcp service", iface),
		fmt.Sprintf("no ipv6 %s mtu", iface),
		fmt.Sprintf("no ipv6 %s secure filter in", iface),
		fmt.Sprintf("no ipv6 %s secure filter out", iface),
	}
}

// ValidateIPv6InterfaceConfig validates an IPv6 interface configuration
func ValidateIPv6InterfaceConfig(config IPv6InterfaceConfig) error {
	// Validate interface name
	if err := ValidateIPv6InterfaceName(config.Interface); err != nil {
		return err
	}

	// Validate addresses
	for i, addr := range config.Addresses {
		if err := validateIPv6Address(addr); err != nil {
			return fmt.Errorf("address[%d]: %w", i, err)
		}
	}

	// Validate RTADV configuration
	if config.RTADV != nil {
		if config.RTADV.PrefixID <= 0 {
			return fmt.Errorf("RTADV prefix_id must be positive")
		}
	}

	// Validate DHCPv6 service
	if config.DHCPv6Service != "" {
		service := strings.ToLower(config.DHCPv6Service)
		if service != "server" && service != "client" && service != "off" {
			return fmt.Errorf("DHCPv6 service must be 'server', 'client', or 'off'")
		}
	}

	// Validate MTU
	if config.MTU != 0 && (config.MTU < 1280 || config.MTU > 65535) {
		return fmt.Errorf("IPv6 MTU must be between 1280 and 65535")
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

	return nil
}

// ValidateIPv6InterfaceName validates an IPv6 interface name
func ValidateIPv6InterfaceName(name string) error {
	if name == "" {
		return fmt.Errorf("interface name is required")
	}
	if !ipv6InterfaceNamePattern.MatchString(name) {
		return fmt.Errorf("invalid interface name: %s (expected lan1, lan2, bridge1, pp1, tunnel1, etc.)", name)
	}
	return nil
}

// validateIPv6Address validates an IPv6Address structure
func validateIPv6Address(addr IPv6Address) error {
	// Must have either a static address or a prefix reference
	hasStatic := addr.Address != ""
	hasPrefixRef := addr.PrefixRef != "" && addr.InterfaceID != ""

	if !hasStatic && !hasPrefixRef {
		return fmt.Errorf("address must have either a static address or prefix_ref with interface_id")
	}

	if hasStatic && hasPrefixRef {
		return fmt.Errorf("address cannot have both static address and prefix_ref")
	}

	// Validate static address format (should contain / for prefix length)
	if hasStatic && !strings.Contains(addr.Address, "/") {
		return fmt.Errorf("static address must include prefix length (e.g., 2001:db8::1/64)")
	}

	// Validate prefix reference format
	if addr.PrefixRef != "" && !strings.Contains(addr.PrefixRef, "@") {
		return fmt.Errorf("prefix_ref must include source interface (e.g., ra-prefix@lan2)")
	}

	// Validate interface ID format
	if addr.InterfaceID != "" && !strings.HasPrefix(addr.InterfaceID, "::") {
		return fmt.Errorf("interface_id must start with :: (e.g., ::1/64)")
	}

	return nil
}

// IsValidIPv6CIDR checks if a string is a valid IPv6 CIDR notation
func IsValidIPv6CIDR(cidr string) bool {
	parts := strings.Split(cidr, "/")
	if len(parts) != 2 {
		return false
	}

	// Check prefix length
	prefix, err := strconv.Atoi(parts[1])
	if err != nil || prefix < 0 || prefix > 128 {
		return false
	}

	// Basic IPv6 format check (contains colons)
	addr := parts[0]
	return strings.Contains(addr, ":")
}
