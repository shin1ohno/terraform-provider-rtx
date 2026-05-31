package parsers

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

// SNMPConfig represents SNMP configuration on an RTX router
type SNMPConfig struct {
	SysName       string          `json:"sysname,omitempty"`         // System name
	SysLocation   string          `json:"syslocation,omitempty"`     // System location
	SysContact    string          `json:"syscontact,omitempty"`      // System contact
	Communities   []SNMPCommunity `json:"communities,omitempty"`     // SNMP communities
	Hosts         []SNMPHost      `json:"hosts,omitempty"`           // SNMP trap hosts
	TrapEnable    []string        `json:"trap_enable,omitempty"`     // Enabled trap types
	HostAccessV1  []string        `json:"host_access_v1,omitempty"`  // SNMPv1 host access-control tokens (any|none|range|lanN|bridgeN — bare IP routes to trap Hosts)
	HostAccessV2c []string        `json:"host_access_v2c,omitempty"` // SNMPv2c host access-control tokens
}

// ipv4Octet matches a single decimal octet 0-255 (no leading-zero overflow like
// "999"). Used to compose the IPv4 / IPv4-range sub-patterns below.
const ipv4Octet = `(?:25[0-5]|2[0-4][0-9]|1?[0-9]?[0-9])`

// ipv4Re is a bare dotted-quad IPv4 address.
const ipv4Re = ipv4Octet + `\.` + ipv4Octet + `\.` + ipv4Octet + `\.` + ipv4Octet

// SNMPHostAccessV1Pattern matches a valid `snmp host <v>` (SNMPv1) access-control
// token: "any", "none", an IPv4 range "A.B.C.D-E.F.G.H", or an interface token
// (lanN / bridgeN). A BARE IPv4 is intentionally EXCLUDED: in this provider's
// model `snmp host <ip>` is the trap-receiver form (the existing `host` block),
// and the two identical command forms cannot be disambiguated on read-back. The
// IPv4 range contains a hyphen, so it does not collide with the trap host parse.
// Single source of truth — referenced by ParseSNMPConfig, ValidateSNMPConfig,
// and the snmp_host schema validator in resource.go.
var SNMPHostAccessV1Pattern = regexp.MustCompile(
	`^(?:any|none|` + ipv4Re + `-` + ipv4Re + `|lan\d+|bridge\d+)$`,
)

// SNMPHostAccessV2cPattern matches a valid `snmpv2c host <v>` (SNMPv2c)
// access-control token: everything SNMPHostAccessV1Pattern accepts PLUS a bare
// IPv4 address. `snmpv2c host <ip>` has no competing trap-receiver parse, so a
// bare IPv4 round-trips cleanly here. Single source of truth — referenced by
// ValidateSNMPConfig and the snmpv2c_host schema validator in resource.go.
var SNMPHostAccessV2cPattern = regexp.MustCompile(
	`^(?:any|none|` + ipv4Re + `(?:-` + ipv4Re + `)?|lan\d+|bridge\d+)$`,
)

// pureIPv4Pattern matches a bare IPv4 address (no range, no interface token).
// Used by ParseSNMPConfig to route a bare `snmp host <ip>` to the trap Hosts
// block rather than the SNMPv1 access-control list.
var pureIPv4Pattern = regexp.MustCompile(`^` + ipv4Re + `$`)

// isSNMPHostAccessV1Token reports whether v is a valid SNMPv1 (`snmp host`)
// access-control token (no bare IPv4).
func isSNMPHostAccessV1Token(v string) bool {
	return SNMPHostAccessV1Pattern.MatchString(v)
}

// isSNMPHostAccessV2cToken reports whether v is a valid SNMPv2c
// (`snmpv2c host`) access-control token (bare IPv4 allowed).
func isSNMPHostAccessV2cToken(v string) bool {
	return SNMPHostAccessV2cPattern.MatchString(v)
}

// SNMPCommunity represents an SNMP community configuration
type SNMPCommunity struct {
	Name       string `json:"name"`          // Community string name
	Permission string `json:"permission"`    // "ro" (read-only) or "rw" (read-write)
	ACL        string `json:"acl,omitempty"` // Access control list (optional)
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
		Communities:   []SNMPCommunity{},
		Hosts:         []SNMPHost{},
		TrapEnable:    []string{},
		HostAccessV1:  []string{},
		HostAccessV2c: []string{},
	}

	// De-wrap RTX console line-wrapping first. The read path always uses SSH
	// `show config | grep snmp`, and the RTX console wraps long lines at 80
	// columns with a bare CRLF (no continuation marker). Without reassembly a
	// wrapped fragment of a long pre-existing field — e.g. `snmp trap enable
	// snmp <many types>` or a long `snmp syslocation <...>` — matches no pattern
	// and is silently dropped, producing "Provider produced inconsistent result
	// after apply". Mirrors dns.go / dhcp_scope.go. A genuine logical line starts
	// with "snmp " or "snmpv2c "; any other non-empty, non-prompt line is a wrap
	// continuation rejoined to the preceding line.
	lines := dewrapConsoleLines(raw, func(trimmed string) bool {
		return strings.HasPrefix(trimmed, "snmp ") || strings.HasPrefix(trimmed, "snmpv2c ")
	})

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
	// snmp host <token> (SNMPv1 access-control form: any|none|range|lanN|bridgeN, or pure IP)
	hostAccessPattern := regexp.MustCompile(`^\s*snmp\s+host\s+(\S+)\s*$`)
	// snmpv2c host <host> [<ro> [<rw>]] (SNMPv2c access-control). The optional
	// trailing ro/rw community tokens (`(?:\s+\S+)*`) are matched but
	// intentionally NOT captured: the provider manages host access only, and
	// BuildSNMPv2cHostCommand emits the host-only form. Capturing communities
	// here would be lossy on the build side anyway.
	hostAccessV2cPattern := regexp.MustCompile(`^\s*snmpv2c\s+host\s+(\S+)(?:\s+\S+)*\s*$`)
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

		// snmpv2c host <host> [<ro> [<rw>]] (SNMPv2c access-control).
		// Only the host token is captured; communities are ignored for now.
		if matches := hostAccessV2cPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.HostAccessV2c = append(config.HostAccessV2c, matches[1])
			continue
		}

		// snmp host (full form: trap receiver with community/version)
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

		// snmp host (simple form: pure IP trap receiver)
		if matches := hostSimplePattern.FindStringSubmatch(line); len(matches) >= 2 {
			host := SNMPHost{
				Address:   matches[1],
				Community: trapCommunity,
			}
			config.Hosts = append(config.Hosts, host)
			continue
		}

		// snmp host <token> (SNMPv1 access-control: any|none|range|lanN|bridgeN).
		// Pure IPs are handled above and routed to Hosts (trap receivers); only
		// non-IP access-control tokens reach this branch. SNMPHostAccessV1Pattern
		// already excludes a bare IPv4, but the explicit pureIPv4Pattern guard
		// makes the trap-vs-access split self-documenting at the branch.
		if matches := hostAccessPattern.FindStringSubmatch(line); len(matches) >= 2 {
			token := matches[1]
			if !pureIPv4Pattern.MatchString(token) && isSNMPHostAccessV1Token(token) {
				config.HostAccessV1 = append(config.HostAccessV1, token)
				continue
			}
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

// BuildSNMPHostAccessCommand builds the SNMPv1 host access-control command.
// Command format: snmp host <v>  (v = any|none|IP|range|lanN|bridgeN)
func BuildSNMPHostAccessCommand(v string) string {
	return fmt.Sprintf("snmp host %s", v)
}

// BuildSNMPv2cHostCommand builds the SNMPv2c host access-control command.
// Command format: snmpv2c host <v>  (v = any|none|IP|range|lanN|bridgeN)
func BuildSNMPv2cHostCommand(v string) string {
	return fmt.Sprintf("snmpv2c host %s", v)
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

// BuildDeleteSNMPHostAccessCommand builds the command to remove an SNMPv1 host
// access-control entry.
// Command format: no snmp host <v>
func BuildDeleteSNMPHostAccessCommand(v string) string {
	return fmt.Sprintf("no snmp host %s", v)
}

// BuildDeleteSNMPv2cHostCommand builds the command to remove an SNMPv2c host
// access-control entry.
// Command format: no snmpv2c host <v>
func BuildDeleteSNMPv2cHostCommand(v string) string {
	return fmt.Sprintf("no snmpv2c host %s", v)
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
		"all":            true,
		"authentication": true,
		"coldstart":      true,
		"warmstart":      true,
		"linkdown":       true,
		"linkup":         true,
		"enterprise":     true,
	}
	for _, trapType := range config.TrapEnable {
		if !validTrapTypes[strings.ToLower(trapType)] {
			return fmt.Errorf("invalid trap type '%s'", trapType)
		}
	}

	// Validate SNMPv1 host access-control tokens. A bare IPv4 is NOT accepted
	// here (it would round-trip into the trap Hosts block); IP-specific v1
	// access remains expressible via the existing `host` block.
	for _, v := range config.HostAccessV1 {
		if !isSNMPHostAccessV1Token(v) {
			return fmt.Errorf("invalid snmp_host value '%s' (expected any|none|IPv4-range|lanN|bridgeN; a bare IPv4 is not allowed — use the host block for SNMPv1 trap receivers, or snmpv2c_host)", v)
		}
		if err := validateSNMPAccessIPs(v); err != nil {
			return fmt.Errorf("invalid snmp_host value '%s': %w", v, err)
		}
	}

	// Validate SNMPv2c host access-control tokens (bare IPv4 allowed).
	for _, v := range config.HostAccessV2c {
		if !isSNMPHostAccessV2cToken(v) {
			return fmt.Errorf("invalid snmpv2c_host value '%s' (expected any|none|IPv4|IPv4-range|lanN|bridgeN)", v)
		}
		if err := validateSNMPAccessIPs(v); err != nil {
			return fmt.Errorf("invalid snmpv2c_host value '%s': %w", v, err)
		}
	}

	return nil
}

// validateSNMPAccessIPs applies a net.ParseIP backstop to any IPv4 / IPv4-range
// access-control token (mirrors the trap Hosts net.ParseIP guard above), so a
// malformed IP that the regex's octet limits miss is still rejected on desired
// config. Non-IP tokens (any|none|lanN|bridgeN) pass through untouched.
// ValidateSNMPConfig runs only on Create/Update (desired config), never on
// device reads, so this will not reject device state.
func validateSNMPAccessIPs(v string) error {
	if strings.Contains(v, "-") {
		parts := strings.SplitN(v, "-", 2)
		for _, p := range parts {
			if net.ParseIP(p) == nil {
				return fmt.Errorf("invalid IPv4 address in range: %s", p)
			}
		}
		return nil
	}
	// A bare dotted-quad (only reachable for snmpv2c_host). Interface tokens and
	// any|none contain no dot, so this check is skipped for them.
	if pureIPv4Pattern.MatchString(v) {
		if net.ParseIP(v) == nil {
			return fmt.Errorf("invalid IPv4 address: %s", v)
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
