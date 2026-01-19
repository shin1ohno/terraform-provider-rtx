package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// IPFilter represents a static IP filter rule on an RTX router
type IPFilter struct {
	Number        int    `json:"number"`                   // Filter number (1-65535)
	Action        string `json:"action"`                   // pass, reject, restrict, restrict-log
	SourceAddress string `json:"source_address"`           // Source IP/network or "*"
	SourceMask    string `json:"source_mask,omitempty"`    // Source mask (for non-CIDR format)
	DestAddress   string `json:"dest_address"`             // Destination IP/network or "*"
	DestMask      string `json:"dest_mask,omitempty"`      // Destination mask (for non-CIDR format)
	Protocol      string `json:"protocol"`                 // tcp, udp, icmp, ip, * (any)
	SourcePort    string `json:"source_port,omitempty"`    // Source port(s) or "*"
	DestPort      string `json:"dest_port,omitempty"`      // Destination port(s) or "*"
	Established   bool   `json:"established,omitempty"`    // Match established TCP connections
}

// IPFilterDynamic represents a dynamic (stateful) IP filter on an RTX router
type IPFilterDynamic struct {
	Number   int    `json:"number"`            // Filter number (1-65535)
	Source   string `json:"source"`            // Source address or "*"
	Dest     string `json:"dest"`              // Destination address or "*"
	Protocol string `json:"protocol"`          // Protocol (ftp, www, smtp, etc.)
	SyslogOn bool   `json:"syslog,omitempty"`  // Enable syslog for this filter
}

// IPFilterSet represents a set of IP filters grouped together
type IPFilterSet struct {
	SetNumber     int   `json:"set_number"`     // Set number
	FilterNumbers []int `json:"filter_numbers"` // List of filter numbers in this set
}

// ValidIPFilterActions defines the valid actions for IP filters
var ValidIPFilterActions = []string{"pass", "reject", "restrict", "restrict-log"}

// ValidIPFilterProtocols defines the valid protocols for IP filters
var ValidIPFilterProtocols = []string{"tcp", "udp", "icmp", "ip", "*", "gre", "esp", "ah", "icmp6"}

// ValidDynamicProtocols defines the valid protocols for dynamic filters
var ValidDynamicProtocols = []string{
	"ftp", "www", "smtp", "pop3", "dns", "domain", "telnet", "ssh",
	"tcp", "udp", "*",
}

// ParseIPFilterConfig parses the output of "show config" command for IP filter lines
func ParseIPFilterConfig(raw string) ([]IPFilter, error) {
	filters := []IPFilter{}
	lines := strings.Split(raw, "\n")

	// Pattern for static IP filter:
	// ip filter <n> <action> <src> <dst> <protocol> [<src_port>] [<dst_port>] [established]
	// The pattern matches required fields and captures optional ones
	filterPattern := regexp.MustCompile(`^\s*ip\s+filter\s+(\d+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)(?:\s+(\S+))?(?:\s+(\S+))?(?:\s+(\S+))?\s*$`)
	// Pattern to detect established keyword
	establishedPattern := regexp.MustCompile(`\bestablished\b`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip dynamic filter lines and interface secure filter lines
		if strings.Contains(line, "ip filter dynamic") ||
			strings.Contains(line, "secure filter") {
			continue
		}

		if matches := filterPattern.FindStringSubmatch(line); len(matches) >= 6 {
			number, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			filter := IPFilter{
				Number:        number,
				Action:        matches[2],
				SourceAddress: matches[3],
				DestAddress:   matches[4],
				Protocol:      matches[5],
			}

			// Check for established keyword in the line first
			hasEstablished := establishedPattern.MatchString(line)
			if hasEstablished {
				filter.Established = true
			}

			// Handle optional ports (skip "established" keyword)
			if len(matches) > 6 && matches[6] != "" && matches[6] != "established" {
				filter.SourcePort = matches[6]
			}
			if len(matches) > 7 && matches[7] != "" && matches[7] != "established" {
				filter.DestPort = matches[7]
			}

			filters = append(filters, filter)
		}
	}

	return filters, nil
}

// ParseIPFilterDynamicConfig parses the output of "show config" for dynamic IP filter lines
func ParseIPFilterDynamicConfig(raw string) ([]IPFilterDynamic, error) {
	filters := []IPFilterDynamic{}
	lines := strings.Split(raw, "\n")

	// Pattern for dynamic IP filter:
	// ip filter dynamic <n> <src> <dst> <protocol> [options]
	dynamicPattern := regexp.MustCompile(`^\s*ip\s+filter\s+dynamic\s+(\d+)\s+(\S+)\s+(\S+)\s+(\S+)(?:\s+(.*))?$`)
	syslogPattern := regexp.MustCompile(`\bsyslog\s+on\b`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if matches := dynamicPattern.FindStringSubmatch(line); len(matches) >= 5 {
			number, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			filter := IPFilterDynamic{
				Number:   number,
				Source:   matches[2],
				Dest:     matches[3],
				Protocol: matches[4],
			}

			// Check for syslog option
			if len(matches) > 5 && matches[5] != "" {
				if syslogPattern.MatchString(matches[5]) {
					filter.SyslogOn = true
				}
			}

			filters = append(filters, filter)
		}
	}

	return filters, nil
}

// ParseInterfaceSecureFilter parses the interface secure filter configuration
// Returns a map of interface -> direction -> filter numbers
func ParseInterfaceSecureFilter(raw string) (map[string]map[string][]int, error) {
	result := make(map[string]map[string][]int)
	lines := strings.Split(raw, "\n")

	// Pattern: ip <interface> secure filter <direction> <filter_numbers...> [dynamic <dynamic_numbers...>]
	// Example: ip lan1 secure filter in 100 101 dynamic 10 20
	securePattern := regexp.MustCompile(`^\s*ip\s+(\S+)\s+secure\s+filter\s+(in|out)\s+(.+)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if matches := securePattern.FindStringSubmatch(line); len(matches) >= 4 {
			iface := matches[1]
			direction := matches[2]
			filterPart := matches[3]

			if result[iface] == nil {
				result[iface] = make(map[string][]int)
			}

			// Parse filter numbers (before "dynamic" keyword)
			filterNums := []int{}
			parts := strings.Fields(filterPart)
			for _, part := range parts {
				if part == "dynamic" {
					break // Stop at dynamic keyword
				}
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

// BuildIPFilterCommand builds the command to create an IP filter
// Command format: ip filter <n> <action> <src> <dst> <protocol> [<src_port>] [<dst_port>]
func BuildIPFilterCommand(filter IPFilter) string {
	parts := []string{
		"ip", "filter",
		strconv.Itoa(filter.Number),
		filter.Action,
		filter.SourceAddress,
		filter.DestAddress,
		filter.Protocol,
	}

	// Add source port if specified
	if filter.SourcePort != "" {
		parts = append(parts, filter.SourcePort)
	} else if filter.DestPort != "" {
		// If only dest port is specified, we need a placeholder for source port
		parts = append(parts, "*")
	}

	// Add destination port if specified
	if filter.DestPort != "" {
		parts = append(parts, filter.DestPort)
	}

	// Add established keyword for TCP
	if filter.Established && strings.ToLower(filter.Protocol) == "tcp" {
		parts = append(parts, "established")
	}

	return strings.Join(parts, " ")
}

// BuildIPFilterDynamicCommand builds the command to create a dynamic IP filter
// Command format: ip filter dynamic <n> <src> <dst> <protocol> [syslog on]
func BuildIPFilterDynamicCommand(filter IPFilterDynamic) string {
	parts := []string{
		"ip", "filter", "dynamic",
		strconv.Itoa(filter.Number),
		filter.Source,
		filter.Dest,
		filter.Protocol,
	}

	if filter.SyslogOn {
		parts = append(parts, "syslog", "on")
	}

	return strings.Join(parts, " ")
}

// BuildDeleteIPFilterCommand builds the command to delete an IP filter
// Command format: no ip filter <n>
func BuildDeleteIPFilterCommand(number int) string {
	return fmt.Sprintf("no ip filter %d", number)
}

// BuildDeleteIPFilterDynamicCommand builds the command to delete a dynamic IP filter
// Command format: no ip filter dynamic <n>
func BuildDeleteIPFilterDynamicCommand(number int) string {
	return fmt.Sprintf("no ip filter dynamic %d", number)
}

// BuildInterfaceSecureFilterCommand builds the command to apply filters to an interface
// Command format: ip <interface> secure filter <direction> <filter_numbers...>
func BuildInterfaceSecureFilterCommand(iface string, direction string, filterNums []int) string {
	parts := []string{"ip", iface, "secure", "filter", direction}
	for _, num := range filterNums {
		parts = append(parts, strconv.Itoa(num))
	}
	return strings.Join(parts, " ")
}

// BuildInterfaceSecureFilterWithDynamicCommand builds the command with both static and dynamic filters
// Command format: ip <interface> secure filter <direction> <static_nums...> dynamic <dynamic_nums...>
func BuildInterfaceSecureFilterWithDynamicCommand(iface string, direction string, staticNums []int, dynamicNums []int) string {
	parts := []string{"ip", iface, "secure", "filter", direction}
	for _, num := range staticNums {
		parts = append(parts, strconv.Itoa(num))
	}
	if len(dynamicNums) > 0 {
		parts = append(parts, "dynamic")
		for _, num := range dynamicNums {
			parts = append(parts, strconv.Itoa(num))
		}
	}
	return strings.Join(parts, " ")
}

// BuildDeleteInterfaceSecureFilterCommand builds the command to remove secure filter from interface
// Command format: no ip <interface> secure filter <direction>
func BuildDeleteInterfaceSecureFilterCommand(iface string, direction string) string {
	return fmt.Sprintf("no ip %s secure filter %s", iface, direction)
}

// BuildShowIPFilterCommand builds the command to show IP filter configuration
// Command format: show config | grep "ip filter"
func BuildShowIPFilterCommand() string {
	return "show config | grep \"ip filter\""
}

// BuildShowIPFilterByNumberCommand builds the command to show a specific IP filter
// Command format: show config | grep "ip filter <n>"
func BuildShowIPFilterByNumberCommand(number int) string {
	return fmt.Sprintf("show config | grep \"ip filter %d\"", number)
}

// ValidateIPFilterNumber validates that the filter number is in valid range (1-65535)
func ValidateIPFilterNumber(n int) error {
	if n < 1 || n > 65535 {
		return fmt.Errorf("filter number must be between 1 and 65535, got %d", n)
	}
	return nil
}

// ValidateIPFilterProtocol validates that the protocol is a valid IP filter protocol
func ValidateIPFilterProtocol(proto string) error {
	proto = strings.ToLower(proto)
	for _, valid := range ValidIPFilterProtocols {
		if proto == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid protocol: %s, must be one of: %s", proto, strings.Join(ValidIPFilterProtocols, ", "))
}

// ValidateIPFilterAction validates that the action is a valid IP filter action
func ValidateIPFilterAction(action string) error {
	action = strings.ToLower(action)
	for _, valid := range ValidIPFilterActions {
		if action == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid action: %s, must be one of: %s", action, strings.Join(ValidIPFilterActions, ", "))
}

// ValidateIPFilter validates a complete IP filter configuration
func ValidateIPFilter(filter IPFilter) error {
	if err := ValidateIPFilterNumber(filter.Number); err != nil {
		return err
	}

	if err := ValidateIPFilterAction(filter.Action); err != nil {
		return err
	}

	if filter.SourceAddress == "" {
		return fmt.Errorf("source address is required")
	}

	if filter.DestAddress == "" {
		return fmt.Errorf("destination address is required")
	}

	if err := ValidateIPFilterProtocol(filter.Protocol); err != nil {
		return err
	}

	// Validate established is only used with TCP
	if filter.Established && strings.ToLower(filter.Protocol) != "tcp" {
		return fmt.Errorf("established keyword can only be used with TCP protocol")
	}

	return nil
}

// ValidateIPFilterDynamic validates a dynamic IP filter configuration
func ValidateIPFilterDynamic(filter IPFilterDynamic) error {
	if err := ValidateIPFilterNumber(filter.Number); err != nil {
		return err
	}

	if filter.Source == "" {
		return fmt.Errorf("source is required")
	}

	if filter.Dest == "" {
		return fmt.Errorf("destination is required")
	}

	if filter.Protocol == "" {
		return fmt.Errorf("protocol is required")
	}

	// Validate protocol for dynamic filter
	proto := strings.ToLower(filter.Protocol)
	valid := false
	for _, p := range ValidDynamicProtocols {
		if proto == p {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid dynamic protocol: %s, must be one of: %s", filter.Protocol, strings.Join(ValidDynamicProtocols, ", "))
	}

	return nil
}

// ValidateIPFilterDirection validates the filter direction (in or out)
func ValidateIPFilterDirection(direction string) error {
	direction = strings.ToLower(direction)
	if direction != "in" && direction != "out" {
		return fmt.Errorf("direction must be 'in' or 'out', got: %s", direction)
	}
	return nil
}

// AccessListExtendedEntry represents a single entry in an IPv4 extended access list
type AccessListExtendedEntry struct {
	Sequence              int
	AceRuleAction         string
	AceRuleProtocol       string
	SourceAny             bool
	SourcePrefix          string
	SourcePrefixMask      string
	SourcePortEqual       string
	SourcePortRange       string
	DestinationAny        bool
	DestinationPrefix     string
	DestinationPrefixMask string
	DestinationPortEqual  string
	DestinationPortRange  string
	Established           bool
	Log                   bool
}

// AccessListExtendedIPv6Entry represents a single entry in an IPv6 extended access list
type AccessListExtendedIPv6Entry struct {
	Sequence                int
	AceRuleAction           string
	AceRuleProtocol         string
	SourceAny               bool
	SourcePrefix            string
	SourcePrefixLength      int
	SourcePortEqual         string
	SourcePortRange         string
	DestinationAny          bool
	DestinationPrefix       string
	DestinationPrefixLength int
	DestinationPortEqual    string
	DestinationPortRange    string
	Established             bool
	Log                     bool
}

// BuildAccessListExtendedEntryCommand builds an IP filter command from an ACL entry
// RTX command: ip filter <sequence> <action> <src> <dst> <protocol> [<src_port>] [<dst_port>] [established]
func BuildAccessListExtendedEntryCommand(entry AccessListExtendedEntry) string {
	// Build source address
	source := "*"
	if !entry.SourceAny && entry.SourcePrefix != "" {
		if entry.SourcePrefixMask != "" {
			source = entry.SourcePrefix + "/" + entry.SourcePrefixMask
		} else {
			source = entry.SourcePrefix
		}
	}

	// Build destination address
	dest := "*"
	if !entry.DestinationAny && entry.DestinationPrefix != "" {
		if entry.DestinationPrefixMask != "" {
			dest = entry.DestinationPrefix + "/" + entry.DestinationPrefixMask
		} else {
			dest = entry.DestinationPrefix
		}
	}

	// Map action
	action := "pass"
	if entry.AceRuleAction == "deny" {
		action = "reject"
	}

	// Build command parts
	parts := []string{
		"ip", "filter",
		strconv.Itoa(entry.Sequence),
		action,
		source,
		dest,
		entry.AceRuleProtocol,
	}

	// Add source port
	srcPort := "*"
	if entry.SourcePortEqual != "" {
		srcPort = entry.SourcePortEqual
	} else if entry.SourcePortRange != "" {
		srcPort = entry.SourcePortRange
	}

	// Add destination port
	dstPort := "*"
	if entry.DestinationPortEqual != "" {
		dstPort = entry.DestinationPortEqual
	} else if entry.DestinationPortRange != "" {
		dstPort = entry.DestinationPortRange
	}

	// Only add ports for tcp/udp
	proto := strings.ToLower(entry.AceRuleProtocol)
	if proto == "tcp" || proto == "udp" {
		if srcPort != "*" || dstPort != "*" {
			parts = append(parts, srcPort, dstPort)
		}
	}

	// Add established keyword for TCP
	if entry.Established && proto == "tcp" {
		parts = append(parts, "established")
	}

	return strings.Join(parts, " ")
}

// BuildAccessListExtendedIPv6EntryCommand builds an IPv6 filter command from an ACL entry
// RTX command: ipv6 filter <sequence> <action> <src> <dst> <protocol> [<src_port>] [<dst_port>]
func BuildAccessListExtendedIPv6EntryCommand(entry AccessListExtendedIPv6Entry) string {
	// Build source address
	source := "*"
	if !entry.SourceAny && entry.SourcePrefix != "" {
		if entry.SourcePrefixLength > 0 {
			source = fmt.Sprintf("%s/%d", entry.SourcePrefix, entry.SourcePrefixLength)
		} else {
			source = entry.SourcePrefix
		}
	}

	// Build destination address
	dest := "*"
	if !entry.DestinationAny && entry.DestinationPrefix != "" {
		if entry.DestinationPrefixLength > 0 {
			dest = fmt.Sprintf("%s/%d", entry.DestinationPrefix, entry.DestinationPrefixLength)
		} else {
			dest = entry.DestinationPrefix
		}
	}

	// Map action
	action := "pass"
	if entry.AceRuleAction == "deny" {
		action = "reject"
	}

	// Map protocol
	protocol := entry.AceRuleProtocol
	if protocol == "ipv6" {
		protocol = "ip"
	}

	// Build command parts
	parts := []string{
		"ipv6", "filter",
		strconv.Itoa(entry.Sequence),
		action,
		source,
		dest,
		protocol,
	}

	// Add source port
	srcPort := "*"
	if entry.SourcePortEqual != "" {
		srcPort = entry.SourcePortEqual
	} else if entry.SourcePortRange != "" {
		srcPort = entry.SourcePortRange
	}

	// Add destination port
	dstPort := "*"
	if entry.DestinationPortEqual != "" {
		dstPort = entry.DestinationPortEqual
	} else if entry.DestinationPortRange != "" {
		dstPort = entry.DestinationPortRange
	}

	// Only add ports for tcp/udp
	proto := strings.ToLower(protocol)
	if proto == "tcp" || proto == "udp" {
		if srcPort != "*" || dstPort != "*" {
			parts = append(parts, srcPort, dstPort)
		}
	}

	return strings.Join(parts, " ")
}

// BuildDeleteIPv6FilterCommand builds the command to delete an IPv6 filter
// Command format: no ipv6 filter <n>
func BuildDeleteIPv6FilterCommand(number int) string {
	return fmt.Sprintf("no ipv6 filter %d", number)
}

// BuildShowIPv6FilterCommand builds the command to show IPv6 filter configuration
func BuildShowIPv6FilterCommand() string {
	return "show config | grep \"ipv6 filter\""
}

// ParseIPv6FilterConfig parses the output of "show config" for IPv6 filter lines
func ParseIPv6FilterConfig(raw string) ([]IPFilter, error) {
	filters := []IPFilter{}
	lines := strings.Split(raw, "\n")

	// Pattern for IPv6 filter:
	// ipv6 filter <n> <action> <src> <dst> <protocol> [<src_port>] [<dst_port>]
	filterPattern := regexp.MustCompile(`^\s*ipv6\s+filter\s+(\d+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)(?:\s+(\S+))?(?:\s+(\S+))?\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip dynamic filter lines
		if strings.Contains(line, "ipv6 filter dynamic") {
			continue
		}

		if matches := filterPattern.FindStringSubmatch(line); len(matches) >= 6 {
			number, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			filter := IPFilter{
				Number:        number,
				Action:        matches[2],
				SourceAddress: matches[3],
				DestAddress:   matches[4],
				Protocol:      matches[5],
			}

			// Handle optional ports
			if len(matches) > 6 && matches[6] != "" {
				filter.SourcePort = matches[6]
			}
			if len(matches) > 7 && matches[7] != "" {
				filter.DestPort = matches[7]
			}

			filters = append(filters, filter)
		}
	}

	return filters, nil
}

// ParseIPv6FilterDynamicConfig parses the output of "show config" for IPv6 dynamic filter lines
func ParseIPv6FilterDynamicConfig(raw string) ([]IPFilterDynamic, error) {
	filters := []IPFilterDynamic{}
	lines := strings.Split(raw, "\n")

	// Pattern for dynamic IPv6 filter:
	// ipv6 filter dynamic <n> <src> <dst> <protocol> [options]
	dynamicPattern := regexp.MustCompile(`^\s*ipv6\s+filter\s+dynamic\s+(\d+)\s+(\S+)\s+(\S+)\s+(\S+)(?:\s+(.*))?$`)
	syslogPattern := regexp.MustCompile(`\bsyslog\s+on\b`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if matches := dynamicPattern.FindStringSubmatch(line); len(matches) >= 5 {
			number, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			filter := IPFilterDynamic{
				Number:   number,
				Source:   matches[2],
				Dest:     matches[3],
				Protocol: matches[4],
			}

			// Check for syslog option
			if len(matches) > 5 && matches[5] != "" {
				if syslogPattern.MatchString(matches[5]) {
					filter.SyslogOn = true
				}
			}

			filters = append(filters, filter)
		}
	}

	return filters, nil
}

// BuildIPv6FilterDynamicCommand builds the command to create a dynamic IPv6 filter
// Command format: ipv6 filter dynamic <n> <src> <dst> <protocol> [syslog on]
func BuildIPv6FilterDynamicCommand(filter IPFilterDynamic) string {
	parts := []string{
		"ipv6", "filter", "dynamic",
		strconv.Itoa(filter.Number),
		filter.Source,
		filter.Dest,
		filter.Protocol,
	}

	if filter.SyslogOn {
		parts = append(parts, "syslog", "on")
	}

	return strings.Join(parts, " ")
}

// BuildDeleteIPv6FilterDynamicCommand builds the command to delete a dynamic IPv6 filter
// Command format: no ipv6 filter dynamic <n>
func BuildDeleteIPv6FilterDynamicCommand(number int) string {
	return fmt.Sprintf("no ipv6 filter dynamic %d", number)
}

// ParseInterfaceIPv6SecureFilter parses the interface IPv6 secure filter configuration
// Returns a map of interface -> direction -> filter numbers
func ParseInterfaceIPv6SecureFilter(raw string) (map[string]map[string][]int, error) {
	result := make(map[string]map[string][]int)
	lines := strings.Split(raw, "\n")

	// Pattern: ipv6 <interface> secure filter <direction> <filter_numbers...> [dynamic <dynamic_numbers...>]
	securePattern := regexp.MustCompile(`^\s*ipv6\s+(\S+)\s+secure\s+filter\s+(in|out)\s+(.+)$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if matches := securePattern.FindStringSubmatch(line); len(matches) >= 4 {
			iface := matches[1]
			direction := matches[2]
			filterPart := matches[3]

			if result[iface] == nil {
				result[iface] = make(map[string][]int)
			}

			// Parse filter numbers (before "dynamic" keyword)
			filterNums := []int{}
			parts := strings.Fields(filterPart)
			for _, part := range parts {
				if part == "dynamic" {
					break
				}
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

// BuildInterfaceIPv6SecureFilterCommand builds the command to apply IPv6 filters to an interface
// Command format: ipv6 <interface> secure filter <direction> <filter_numbers...>
func BuildInterfaceIPv6SecureFilterCommand(iface string, direction string, filterNums []int) string {
	parts := []string{"ipv6", iface, "secure", "filter", direction}
	for _, num := range filterNums {
		parts = append(parts, strconv.Itoa(num))
	}
	return strings.Join(parts, " ")
}

// BuildInterfaceIPv6SecureFilterWithDynamicCommand builds the command with both static and dynamic filters
// Command format: ipv6 <interface> secure filter <direction> <static_nums...> dynamic <dynamic_nums...>
func BuildInterfaceIPv6SecureFilterWithDynamicCommand(iface string, direction string, staticNums []int, dynamicNums []int) string {
	parts := []string{"ipv6", iface, "secure", "filter", direction}
	for _, num := range staticNums {
		parts = append(parts, strconv.Itoa(num))
	}
	if len(dynamicNums) > 0 {
		parts = append(parts, "dynamic")
		for _, num := range dynamicNums {
			parts = append(parts, strconv.Itoa(num))
		}
	}
	return strings.Join(parts, " ")
}

// BuildDeleteInterfaceIPv6SecureFilterCommand builds the command to remove IPv6 secure filter from interface
// Command format: no ipv6 <interface> secure filter <direction>
func BuildDeleteInterfaceIPv6SecureFilterCommand(iface string, direction string) string {
	return fmt.Sprintf("no ipv6 %s secure filter %s", iface, direction)
}
