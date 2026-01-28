package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DNSConfig represents DNS server configuration on an RTX router
type DNSConfig struct {
	DomainLookup bool              `json:"domain_lookup"` // dns domain lookup enable/disable
	DomainName   string            `json:"domain_name"`   // dns domain name
	NameServers  []string          `json:"name_servers"`  // dns server <ip1> [<ip2>]
	ServerSelect []DNSServerSelect `json:"server_select"` // dns server select entries
	Hosts        []DNSHost         `json:"hosts"`         // dns static entries
	ServiceOn    bool              `json:"service_on"`    // dns service on/off
	PrivateSpoof bool              `json:"private_spoof"` // dns private address spoof on/off
}

// DNSServer represents a DNS server with its per-server EDNS setting
type DNSServer struct {
	Address string `json:"address"` // DNS server IP address (IPv4 or IPv6)
	EDNS    bool   `json:"edns"`    // Enable EDNS for this server
}

// DNSServerSelect represents a domain-based DNS server selection entry
type DNSServerSelect struct {
	ID             int         `json:"id"`              // Selector ID (1-65535)
	Servers        []DNSServer `json:"servers"`         // DNS servers with per-server EDNS
	RecordType     string      `json:"record_type"`     // DNS record type: a, aaaa, ptr, mx, ns, cname, any
	QueryPattern   string      `json:"query_pattern"`   // Domain pattern: ".", "*.example.com", etc.
	OriginalSender string      `json:"original_sender"` // Source IP/CIDR restriction
	RestrictPP     int         `json:"restrict_pp"`     // PP session restriction (0=none)
}

// DNSHost represents a static DNS host entry
type DNSHost struct {
	Name    string `json:"name"`    // Hostname
	Address string `json:"address"` // IP address
}

// validRecordTypes contains the valid DNS record types for server select
var validRecordTypes = map[string]bool{
	"a":     true,
	"aaaa":  true,
	"ptr":   true,
	"mx":    true,
	"ns":    true,
	"cname": true,
	"any":   true,
}

// DNSParser parses DNS configuration output
type DNSParser struct{}

// NewDNSParser creates a new DNS parser
func NewDNSParser() *DNSParser {
	return &DNSParser{}
}

// ParseDNSConfig parses the output of "show config" command for DNS configuration
func (p *DNSParser) ParseDNSConfig(raw string) (*DNSConfig, error) {
	config := &DNSConfig{
		DomainLookup: true,  // Default: enabled
		ServiceOn:    false, // Default: off
		PrivateSpoof: false, // Default: off
		NameServers:  []string{},
		ServerSelect: []DNSServerSelect{},
		Hosts:        []DNSHost{},
	}

	// Pre-process: Join lines that were wrapped by RTX router
	// RTX router wraps long lines at ~80 chars, e.g., "edns=on" becomes "edns\n=on"
	// Lines starting with "=" are continuation lines and should be joined to previous
	// Also handle CRLF by normalizing line endings first
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")
	rawLines := strings.Split(raw, "\n")
	var joinedLines []string
	for _, line := range rawLines {
		trimmed := strings.TrimSpace(line)
		// If line starts with "=", it's a continuation of the previous line
		if strings.HasPrefix(trimmed, "=") && len(joinedLines) > 0 {
			// Trim the previous line before joining to remove any trailing whitespace
			prevLine := strings.TrimRight(joinedLines[len(joinedLines)-1], " \t\r")
			joinedLines[len(joinedLines)-1] = prevLine + trimmed
		} else {
			joinedLines = append(joinedLines, line)
		}
	}
	raw = strings.Join(joinedLines, "\n")

	lines := strings.Split(raw, "\n")

	// Patterns for different DNS configuration lines
	// dns domain lookup on/off
	domainLookupPattern := regexp.MustCompile(`^\s*dns\s+domain\s+lookup\s+(on|off)\s*$`)
	// dns domain <name>
	domainNamePattern := regexp.MustCompile(`^\s*dns\s+domain\s+(\S+)\s*$`)
	// dns server <ip1> [<ip2>] [<ip3>]
	dnsServerPattern := regexp.MustCompile(`^\s*dns\s+server\s+(\S+)(?:\s+(\S+))?(?:\s+(\S+))?\s*$`)
	// dns server select <id> <server> [<server2>] <domain1> [<domain2>...]
	// Format: dns server select <id> <server(s)> <domain(s)>
	dnsServerSelectPattern := regexp.MustCompile(`^\s*dns\s+server\s+select\s+(\d+)\s+(.+)\s*$`)
	// dns static <hostname> <ip>
	dnsStaticPattern := regexp.MustCompile(`^\s*dns\s+static\s+(\S+)\s+(\S+)\s*$`)
	// dns service on/off/recursive
	dnsServicePattern := regexp.MustCompile(`^\s*dns\s+service\s+(on|off|recursive)\s*$`)
	// dns private address spoof on/off
	dnsPrivateSpoofPattern := regexp.MustCompile(`^\s*dns\s+private\s+address\s+spoof\s+(on|off)\s*$`)
	// no dns domain lookup (disable)
	noDomainLookupPattern := regexp.MustCompile(`^\s*no\s+dns\s+domain\s+lookup\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Try domain lookup pattern
		if matches := domainLookupPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.DomainLookup = matches[1] == "on"
			continue
		}

		// Try no domain lookup pattern
		if noDomainLookupPattern.MatchString(line) {
			config.DomainLookup = false
			continue
		}

		// Try DNS service pattern
		if matches := dnsServicePattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.ServiceOn = (matches[1] == "on" || matches[1] == "recursive")
			continue
		}

		// Try DNS private spoof pattern
		if matches := dnsPrivateSpoofPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.PrivateSpoof = matches[1] == "on"
			continue
		}

		// Try DNS server select pattern (must be before dns server pattern)
		if matches := dnsServerSelectPattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, err := strconv.Atoi(matches[1])
			if err != nil {
				continue
			}

			sel := parseDNSServerSelectFields(id, matches[2])
			if sel != nil {
				config.ServerSelect = append(config.ServerSelect, *sel)
			}
			continue
		}

		// Try DNS server pattern
		if matches := dnsServerPattern.FindStringSubmatch(line); len(matches) >= 2 {
			for i := 1; i < len(matches); i++ {
				if matches[i] != "" {
					config.NameServers = append(config.NameServers, matches[i])
				}
			}
			continue
		}

		// Try domain name pattern (must be after dns domain lookup)
		if matches := domainNamePattern.FindStringSubmatch(line); len(matches) >= 2 {
			// Avoid matching "dns domain lookup" as domain name
			if matches[1] != "lookup" {
				config.DomainName = matches[1]
			}
			continue
		}

		// Try DNS static pattern
		if matches := dnsStaticPattern.FindStringSubmatch(line); len(matches) >= 3 {
			config.Hosts = append(config.Hosts, DNSHost{
				Name:    matches[1],
				Address: matches[2],
			})
			continue
		}
	}

	return config, nil
}

// isValidIPForDNS checks if a string looks like an IP address (for DNS server parsing)
func isValidIPForDNS(s string) bool {
	// Check IPv4
	ipv4Pattern := regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)
	if ipv4Pattern.MatchString(s) {
		return true
	}

	// Check IPv6 (basic check)
	if strings.Contains(s, ":") && !strings.Contains(s, "*") {
		return true
	}

	return false
}

// isIPOrCIDR checks if a string looks like an IP address or CIDR notation
func isIPOrCIDR(s string) bool {
	// Check for CIDR notation
	if strings.Contains(s, "/") {
		parts := strings.Split(s, "/")
		if len(parts) == 2 && isValidIPForDNS(parts[0]) {
			return true
		}
	}
	// Check for IP range (192.168.1.0-192.168.1.255)
	if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		if len(parts) == 2 && isValidIPForDNS(parts[0]) && isValidIPForDNS(parts[1]) {
			return true
		}
	}
	return isValidIPForDNS(s)
}

// parseDNSServerSelectFields parses the fields after "dns server select <id>"
// Format: <server1> [edns=on|edns=off] [<server2> [edns=on|edns=off]] [record_type] <query-pattern> [original-sender] [restrict pp n]
//
// The parsing follows strict RTX command order:
// 1. servers (1-2 IPs) with optional edns=on/edns=off after each server (per-server EDNS)
// 2. record_type if in validRecordTypes AND not "." (dot is always query_pattern)
// 3. query_pattern (required, first non-IP/non-keyword after optional fields)
// 4. original_sender (optional, IP/CIDR after query_pattern)
// 5. restrict pp n if present
func parseDNSServerSelectFields(id int, rest string) *DNSServerSelect {
	fields := strings.Fields(rest)
	if len(fields) < 2 {
		return nil
	}

	sel := &DNSServerSelect{
		ID:      id,
		Servers: []DNSServer{},
		// RecordType left empty - defaults to "a" implicitly when building commands
	}

	i := 0

	// Phase 1: Parse servers with interleaved edns=on/edns=off options
	// RTX routers support at most 2 DNS servers per select entry
	// Format can be: <server1> edns=on <server2> edns=on (interleaved)
	//            or: <server1> <server2> edns=on (trailing - applies to all)
	const maxServers = 2
	for i < len(fields) && len(sel.Servers) < maxServers {
		if isValidIPForDNS(fields[i]) {
			server := DNSServer{
				Address: fields[i],
				EDNS:    false, // Default: EDNS disabled
			}
			i++
			// Check for edns=on or edns=off immediately after this server
			if i < len(fields) {
				if fields[i] == "edns=on" {
					server.EDNS = true
					i++
				} else if fields[i] == "edns=off" {
					server.EDNS = false
					i++
				}
			}
			sel.Servers = append(sel.Servers, server)
		} else {
			break // Not an IP, move to next phase
		}
	}

	if len(sel.Servers) == 0 {
		return nil
	}

	// Phase 2: Check for record_type (must come before query_pattern)
	// Note: "." is NOT a valid record type, it's always a query pattern
	if i < len(fields) && validRecordTypes[fields[i]] && fields[i] != "." {
		sel.RecordType = fields[i]
		i++
	}

	// Phase 3: Parse query_pattern (required)
	// The next non-IP/non-keyword field is the query pattern
	if i < len(fields) {
		// query_pattern can be: ".", "*.example.com", "example.com", etc.
		// It is NOT an IP address at this point (servers already parsed)
		sel.QueryPattern = fields[i]
		i++
	}

	if sel.QueryPattern == "" {
		return nil
	}

	// Phase 4: Check for original_sender (optional, IP/CIDR after query_pattern)
	if i < len(fields) && isIPOrCIDR(fields[i]) {
		sel.OriginalSender = fields[i]
		i++
	}

	// Phase 5: Check for "restrict pp n" (must be at the end)
	if i < len(fields) && fields[i] == "restrict" && i+2 < len(fields) && fields[i+1] == "pp" {
		if pp, err := strconv.Atoi(fields[i+2]); err == nil {
			sel.RestrictPP = pp
		}
		// i += 3 // Not needed as we're done parsing
	}

	return sel
}

// BuildDNSServerCommand builds the command to set DNS servers
// Command format: dns server <ip1> [<ip2>] [<ip3>]
func BuildDNSServerCommand(servers []string) string {
	if len(servers) == 0 {
		return ""
	}
	return fmt.Sprintf("dns server %s", strings.Join(servers, " "))
}

// BuildDeleteDNSServerCommand builds the command to remove DNS servers
// Command format: no dns server
func BuildDeleteDNSServerCommand() string {
	return "no dns server"
}

// BuildDNSServerSelectCommand builds the command for domain-based DNS server selection
// Command format: dns server select <id> <server1> [edns=on] [<server2> [edns=on]] [type] <query-pattern> [original-sender] [restrict pp n]
func BuildDNSServerSelectCommand(sel DNSServerSelect) string {
	if sel.ID < 1 || len(sel.Servers) == 0 || sel.QueryPattern == "" {
		return ""
	}

	parts := []string{
		"dns server select",
		strconv.Itoa(sel.ID),
	}

	// Add servers with per-server EDNS options
	for _, server := range sel.Servers {
		parts = append(parts, server.Address)
		if server.EDNS {
			parts = append(parts, "edns=on")
		}
	}

	// Add record type if not default "a"
	if sel.RecordType != "" && sel.RecordType != "a" {
		parts = append(parts, sel.RecordType)
	}

	// Add query pattern (required)
	parts = append(parts, sel.QueryPattern)

	// Add original sender if specified
	if sel.OriginalSender != "" {
		parts = append(parts, sel.OriginalSender)
	}

	// Add restrict pp if specified
	if sel.RestrictPP > 0 {
		parts = append(parts, "restrict", "pp", strconv.Itoa(sel.RestrictPP))
	}

	return strings.Join(parts, " ")
}

// BuildDeleteDNSServerSelectCommand builds the command to remove a DNS server select entry
// Command format: no dns server select <id>
func BuildDeleteDNSServerSelectCommand(id int) string {
	return fmt.Sprintf("no dns server select %d", id)
}

// BuildDNSStaticCommand builds the command for a static DNS host entry
// Command format: dns static <hostname> <ip>
func BuildDNSStaticCommand(host DNSHost) string {
	if host.Name == "" || host.Address == "" {
		return ""
	}
	return fmt.Sprintf("dns static %s %s", host.Name, host.Address)
}

// BuildDeleteDNSStaticCommand builds the command to remove a static DNS host entry
// Command format: no dns static <hostname>
func BuildDeleteDNSStaticCommand(hostname string) string {
	return fmt.Sprintf("no dns static %s", hostname)
}

// BuildDNSServiceCommand builds the command to enable/disable DNS service
// Command format: dns service recursive/off (recursive is preferred form for enabled)
func BuildDNSServiceCommand(enable bool) string {
	if enable {
		return "dns service recursive"
	}
	return "dns service off"
}

// BuildDNSPrivateSpoofCommand builds the command to enable/disable DNS private address spoofing
// Command format: dns private address spoof on/off
func BuildDNSPrivateSpoofCommand(enable bool) string {
	if enable {
		return "dns private address spoof on"
	}
	return "dns private address spoof off"
}

// BuildDNSDomainLookupCommand builds the command to enable/disable DNS domain lookup
// Command format: dns domain lookup on/off (or no dns domain lookup)
func BuildDNSDomainLookupCommand(enable bool) string {
	if enable {
		return "dns domain lookup on"
	}
	return "no dns domain lookup"
}

// BuildDNSDomainNameCommand builds the command to set the domain name
// Command format: dns domain <name>
func BuildDNSDomainNameCommand(name string) string {
	if name == "" {
		return ""
	}
	return fmt.Sprintf("dns domain %s", name)
}

// BuildDeleteDNSDomainNameCommand builds the command to remove the domain name
// Command format: no dns domain
func BuildDeleteDNSDomainNameCommand() string {
	return "no dns domain"
}

// BuildDeleteDNSCommand builds commands to remove all DNS configuration
func BuildDeleteDNSCommand() []string {
	return []string{
		"no dns server",
		"no dns domain",
		"dns service off",
		"dns private address spoof off",
	}
}

// BuildShowDNSConfigCommand builds the command to show DNS configuration
func BuildShowDNSConfigCommand() string {
	return "show config | grep dns"
}

// ValidateDNSConfig validates a DNS configuration
func ValidateDNSConfig(config DNSConfig) error {
	// Validate name servers
	for _, server := range config.NameServers {
		if !isValidIPForDNS(server) {
			return fmt.Errorf("invalid DNS server IP address: %s", server)
		}
	}

	// Maximum 3 name servers
	if len(config.NameServers) > 3 {
		return fmt.Errorf("maximum 3 DNS servers allowed, got %d", len(config.NameServers))
	}

	// Validate server select entries
	for _, sel := range config.ServerSelect {
		if sel.ID < 1 || sel.ID > 65535 {
			return fmt.Errorf("dns server select ID must be between 1 and 65535, got %d", sel.ID)
		}
		if len(sel.Servers) == 0 {
			return fmt.Errorf("dns server select %d must have at least one server", sel.ID)
		}
		if len(sel.Servers) > 2 {
			return fmt.Errorf("dns server select %d: maximum 2 servers allowed, got %d", sel.ID, len(sel.Servers))
		}
		if sel.QueryPattern == "" {
			return fmt.Errorf("dns server select %d must have a query pattern", sel.ID)
		}
		// Validate record type if specified
		if sel.RecordType != "" && !validRecordTypes[sel.RecordType] {
			return fmt.Errorf("dns server select %d: invalid record type %q, must be one of: a, aaaa, ptr, mx, ns, cname, any", sel.ID, sel.RecordType)
		}
		for _, server := range sel.Servers {
			if !isValidIPForDNS(server.Address) {
				return fmt.Errorf("dns server select %d: invalid server IP address: %s", sel.ID, server.Address)
			}
		}
	}

	// Validate static hosts
	for _, host := range config.Hosts {
		if host.Name == "" {
			return fmt.Errorf("dns static host name cannot be empty")
		}
		if !isValidIPForDNS(host.Address) {
			return fmt.Errorf("dns static host %s: invalid IP address: %s", host.Name, host.Address)
		}
	}

	return nil
}
