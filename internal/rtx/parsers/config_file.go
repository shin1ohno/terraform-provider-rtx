package parsers

import (
	"regexp"
	"strconv"
	"strings"
)

// ExtractedPasswords contains all password and secret values extracted from config
type ExtractedPasswords struct {
	LoginPassword string              `json:"login_password,omitempty"`
	AdminPassword string              `json:"admin_password,omitempty"`
	Users         []ExtractedUserAuth `json:"users,omitempty"`
	IPsecPSK      []ExtractedIPsecPSK `json:"ipsec_psk,omitempty"`
	L2TPAuth      []ExtractedL2TPAuth `json:"l2tp_auth,omitempty"`
	PPAuth        []ExtractedPPAuth   `json:"pp_auth,omitempty"`
}

// ExtractedUserAuth represents a user authentication entry
type ExtractedUserAuth struct {
	Username  string `json:"username"`
	Password  string `json:"password,omitempty"` // Empty if encrypted
	Encrypted bool   `json:"encrypted"`
}

// ExtractedIPsecPSK represents an IPsec pre-shared key
type ExtractedIPsecPSK struct {
	ID        int    `json:"id"`
	Secret    string `json:"secret,omitempty"` // Empty if encrypted
	Encrypted bool   `json:"encrypted"`
}

// ExtractedL2TPAuth represents L2TP tunnel authentication
type ExtractedL2TPAuth struct {
	TunnelID int    `json:"tunnel_id"`
	Secret   string `json:"secret"`
}

// ExtractedPPAuth represents PP authentication credentials
type ExtractedPPAuth struct {
	PPID     int    `json:"pp_id"` // 0 for anonymous
	PPName   string `json:"pp_name,omitempty"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// ContextType represents the type of configuration context
type ContextType int

const (
	// ContextGlobal represents global (non-context) commands
	ContextGlobal ContextType = iota
	// ContextTunnel represents tunnel select N context
	ContextTunnel
	// ContextPP represents pp select N context
	ContextPP
	// ContextIPsecTunnel represents nested ipsec tunnel N context
	ContextIPsecTunnel
)

// String returns a string representation of the context type
func (ct ContextType) String() string {
	switch ct {
	case ContextGlobal:
		return "global"
	case ContextTunnel:
		return "tunnel"
	case ContextPP:
		return "pp"
	case ContextIPsecTunnel:
		return "ipsec-tunnel"
	default:
		return "unknown"
	}
}

// ParseContext represents a configuration context (tunnel select, pp select, etc.)
type ParseContext struct {
	Type ContextType // Type of context
	ID   int         // Numeric ID (e.g., tunnel number, pp number)
	Name string      // Optional name (e.g., "anonymous" for pp select anonymous)
}

// ParsedCommand represents a parsed command line with its context
type ParsedCommand struct {
	Line        string        // The command line (trimmed)
	Context     *ParseContext // The context this command belongs to (nil for global)
	LineNumber  int           // Original line number in the config
	IndentLevel int           // Indentation level (number of leading spaces/tabs)
}

// ParsedConfig represents the parsed configuration file
type ParsedConfig struct {
	Raw          string          // Raw content for debugging
	LineCount    int             // Total number of non-empty lines
	CommandCount int             // Number of command lines (excluding comments)
	Contexts     []ParseContext  // All contexts found in the config
	Commands     []ParsedCommand // All parsed commands
}

// GetCommandsInContext returns all commands within a specific context
func (pc *ParsedConfig) GetCommandsInContext(ctx ParseContext) []ParsedCommand {
	var result []ParsedCommand
	for _, cmd := range pc.Commands {
		if cmd.Context != nil &&
			cmd.Context.Type == ctx.Type &&
			cmd.Context.ID == ctx.ID &&
			cmd.Context.Name == ctx.Name {
			result = append(result, cmd)
		}
	}
	return result
}

// GetGlobalCommands returns all commands not in any context
func (pc *ParsedConfig) GetGlobalCommands() []ParsedCommand {
	var result []ParsedCommand
	for _, cmd := range pc.Commands {
		if cmd.Context == nil {
			result = append(result, cmd)
		}
	}
	return result
}

// ConfigFileParser parses RTX router config.txt files
type ConfigFileParser struct {
	// Patterns for context detection
	tunnelSelectPattern      *regexp.Regexp
	ppSelectPattern          *regexp.Regexp
	ppSelectAnonymousPattern *regexp.Regexp
	ipsecTunnelPattern       *regexp.Regexp
}

// NewConfigFileParser creates a new config file parser
func NewConfigFileParser() *ConfigFileParser {
	return &ConfigFileParser{
		tunnelSelectPattern:      regexp.MustCompile(`^\s*tunnel\s+select\s+(\d+)\s*$`),
		ppSelectPattern:          regexp.MustCompile(`^\s*pp\s+select\s+(\d+)\s*$`),
		ppSelectAnonymousPattern: regexp.MustCompile(`^\s*pp\s+select\s+anonymous\s*$`),
		ipsecTunnelPattern:       regexp.MustCompile(`^\s*ipsec\s+tunnel\s+(\d+)\s*$`),
	}
}

// Parse parses the raw config file content
func (p *ConfigFileParser) Parse(raw string) (*ParsedConfig, error) {
	result := &ParsedConfig{
		Raw:      raw,
		Contexts: []ParseContext{},
		Commands: []ParsedCommand{},
	}

	// Normalize line endings
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")

	lines := strings.Split(raw, "\n")

	// Context tracking
	var currentContext *ParseContext
	var contextStack []ParseContext     // Stack for nested contexts
	contextMap := make(map[string]bool) // Track unique contexts

	lineNumber := 0
	for _, line := range lines {
		lineNumber++

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}
		result.LineCount++

		// Get indentation level
		indentLevel := p.getIndentLevel(line)
		trimmedLine := strings.TrimSpace(line)

		// Skip comment lines
		if strings.HasPrefix(trimmedLine, "#") {
			continue
		}
		result.CommandCount++

		// Check for context changes
		newContext := p.detectContext(trimmedLine, indentLevel)
		if newContext != nil {
			// Handle nested contexts (e.g., ipsec tunnel within tunnel select)
			if newContext.Type == ContextIPsecTunnel && currentContext != nil &&
				currentContext.Type == ContextTunnel {
				// Push current context to stack
				contextStack = append(contextStack, *currentContext)
			} else if newContext.Type == ContextTunnel || newContext.Type == ContextPP {
				// Clear the stack for new top-level contexts
				contextStack = nil
			}

			currentContext = newContext

			// Track unique contexts
			contextKey := p.contextKey(*newContext)
			if !contextMap[contextKey] {
				contextMap[contextKey] = true
				result.Contexts = append(result.Contexts, *newContext)
			}
			continue
		}

		// Check if this is a context-exiting command (tunnel enable N, pp enable)
		// This can happen at any indent level within the context
		if currentContext != nil && p.isContextExitLine(trimmedLine) {
			// This command belongs to the current context but marks its end
			// Create command first, then exit context
			cmd := ParsedCommand{
				Line:        trimmedLine,
				LineNumber:  lineNumber,
				IndentLevel: indentLevel,
			}
			ctx := *currentContext
			cmd.Context = &ctx
			result.Commands = append(result.Commands, cmd)

			// Now exit the context
			contextStack = nil
			currentContext = nil
			continue
		}

		// Check if we're exiting a context based on indentation
		// If indentation is 0, we might be back at global level
		if indentLevel == 0 && currentContext != nil {
			if !p.isContextualCommand(trimmedLine, currentContext) {
				// Reset context for non-contextual commands at indent 0
				if len(contextStack) > 0 {
					// Pop from stack
					currentContext = &contextStack[len(contextStack)-1]
					contextStack = contextStack[:len(contextStack)-1]
				} else {
					// Not in any nested context, this is a global command
					currentContext = nil
				}
			}
		}

		// Check if we're in an IPsec tunnel context and see a non-IPsec command
		// that belongs to the parent tunnel context (like l2tp, tunnel endpoint, etc.)
		if currentContext != nil && currentContext.Type == ContextIPsecTunnel && len(contextStack) > 0 {
			// Check if this command should belong to the parent tunnel context
			if strings.HasPrefix(trimmedLine, "l2tp ") ||
				strings.HasPrefix(trimmedLine, "tunnel endpoint") ||
				strings.HasPrefix(trimmedLine, "tunnel enable") ||
				strings.HasPrefix(trimmedLine, "ip tunnel ") {
				// Pop back to parent tunnel context
				currentContext = &contextStack[len(contextStack)-1]
				contextStack = contextStack[:len(contextStack)-1]
			}
		}

		// Create parsed command
		cmd := ParsedCommand{
			Line:        trimmedLine,
			LineNumber:  lineNumber,
			IndentLevel: indentLevel,
		}

		// Assign context (clone to avoid reference issues)
		if currentContext != nil {
			ctx := *currentContext
			cmd.Context = &ctx
		}

		result.Commands = append(result.Commands, cmd)
	}

	return result, nil
}

// detectContext checks if a line starts a new context and returns it
func (p *ConfigFileParser) detectContext(line string, indentLevel int) *ParseContext {
	// tunnel select N
	if matches := p.tunnelSelectPattern.FindStringSubmatch(line); len(matches) >= 2 {
		id, _ := strconv.Atoi(matches[1])
		return &ParseContext{Type: ContextTunnel, ID: id}
	}

	// pp select anonymous
	if p.ppSelectAnonymousPattern.MatchString(line) {
		return &ParseContext{Type: ContextPP, ID: 0, Name: "anonymous"}
	}

	// pp select N
	if matches := p.ppSelectPattern.FindStringSubmatch(line); len(matches) >= 2 {
		id, _ := strconv.Atoi(matches[1])
		return &ParseContext{Type: ContextPP, ID: id}
	}

	// ipsec tunnel N (nested context)
	if matches := p.ipsecTunnelPattern.FindStringSubmatch(line); len(matches) >= 2 {
		id, _ := strconv.Atoi(matches[1])
		return &ParseContext{Type: ContextIPsecTunnel, ID: id}
	}

	return nil
}

// getIndentLevel returns the number of leading whitespace characters
func (p *ConfigFileParser) getIndentLevel(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' || ch == '\t' {
			count++
		} else {
			break
		}
	}
	return count
}

// contextKey returns a unique string key for a context
func (p *ConfigFileParser) contextKey(ctx ParseContext) string {
	if ctx.Name != "" {
		return ctx.Type.String() + ":" + ctx.Name
	}
	return ctx.Type.String() + ":" + strconv.Itoa(ctx.ID)
}

// isContextExitLine checks if a line is an enable/disable command for the context
func (p *ConfigFileParser) isContextExitLine(line string) bool {
	// tunnel enable N, pp enable N, etc.
	enablePattern := regexp.MustCompile(`^(tunnel|pp)\s+enable\s+`)
	disablePattern := regexp.MustCompile(`^(tunnel|pp)\s+disable\s+`)
	return enablePattern.MatchString(line) || disablePattern.MatchString(line)
}

// isContextualCommand checks if a command is typically found within the current context
func (p *ConfigFileParser) isContextualCommand(line string, ctx *ParseContext) bool {
	if ctx == nil {
		return false
	}

	switch ctx.Type {
	case ContextTunnel:
		// Commands typically found in tunnel context
		tunnelCommands := []string{
			"tunnel ", "ipsec ", "l2tp ", "description ",
		}
		for _, prefix := range tunnelCommands {
			if strings.HasPrefix(line, prefix) {
				return true
			}
		}
	case ContextPP:
		// Commands typically found in pp context
		ppCommands := []string{
			"pp ", "pppoe ", "ppp ", "ip pp ", "description ",
		}
		for _, prefix := range ppCommands {
			if strings.HasPrefix(line, prefix) {
				return true
			}
		}
	case ContextIPsecTunnel:
		// Commands in ipsec tunnel context
		return strings.HasPrefix(line, "ipsec ")
	}

	return false
}

// ============================================================================
// Resource Extraction Methods
// ============================================================================

// ExtractStaticRoutes extracts static routes from parsed config using the existing parser
func (pc *ParsedConfig) ExtractStaticRoutes() []StaticRoute {
	// Build raw config string from global commands that start with "ip route"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "ip route ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	parser := NewStaticRouteParser()
	routes, _ := parser.ParseRouteConfig(raw)
	return routes
}

// ExtractDHCPScopes extracts DHCP scopes from parsed config using the existing parser
func (pc *ParsedConfig) ExtractDHCPScopes() []DHCPScope {
	// Build raw config string from global commands that start with "dhcp scope"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "dhcp scope ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	parser := NewDHCPScopeParser()
	scopes, _ := parser.ParseScopeConfig(raw)
	return scopes
}

// ExtractNATMasquerade extracts NAT masquerade descriptors from parsed config
func (pc *ParsedConfig) ExtractNATMasquerade() []NATMasquerade {
	// Build raw config string from global commands that start with "nat descriptor"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "nat descriptor ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	nats, _ := ParseNATMasqueradeConfig(raw)
	return nats
}

// ExtractIPFilters extracts static IP filters from parsed config
func (pc *ParsedConfig) ExtractIPFilters() []IPFilter {
	// Build raw config string from global commands that start with "ip filter"
	// but exclude "ip filter dynamic" lines
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "ip filter ") && !strings.HasPrefix(cmd.Line, "ip filter dynamic ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	filters, _ := ParseIPFilterConfig(raw)
	return filters
}

// ExtractIPFiltersDynamic extracts dynamic IP filters from parsed config
func (pc *ParsedConfig) ExtractIPFiltersDynamic() []IPFilterDynamic {
	// Build raw config string from global commands that start with "ip filter dynamic"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "ip filter dynamic ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	filters, _ := ParseIPFilterDynamicConfig(raw)
	return filters
}

// ExtractHTTPD extracts HTTPD configuration from parsed config using the service parser
func (pc *ParsedConfig) ExtractHTTPD() *HTTPDConfig {
	// Build raw config string from global commands that start with "httpd"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "httpd ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	parser := NewServiceParser()
	config, _ := parser.ParseHTTPDConfig(raw)
	return config
}

// ExtractSSHD extracts SSHD configuration from parsed config using the service parser
func (pc *ParsedConfig) ExtractSSHD() *SSHDConfig {
	// Build raw config string from global commands that start with "sshd"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "sshd ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	parser := NewServiceParser()
	config, _ := parser.ParseSSHDConfig(raw)
	return config
}

// ExtractSFTPD extracts SFTPD configuration from parsed config using the service parser
func (pc *ParsedConfig) ExtractSFTPD() *SFTPDConfig {
	// Build raw config string from global commands that start with "sftpd"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "sftpd ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	parser := NewServiceParser()
	config, _ := parser.ParseSFTPDConfig(raw)
	return config
}

// ExtractDNSServer extracts DNS configuration from parsed config using the DNS parser
func (pc *ParsedConfig) ExtractDNSServer() *DNSConfig {
	// Build raw config string from global commands that start with "dns"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "dns ") || strings.HasPrefix(cmd.Line, "no dns ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	parser := NewDNSParser()
	config, _ := parser.ParseDNSConfig(raw)
	return config
}

// ExtractSystem extracts system configuration from parsed config using the system parser
func (pc *ParsedConfig) ExtractSystem() *SystemConfig {
	// Build raw config string from global commands that match system config patterns
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "timezone ") ||
			strings.HasPrefix(cmd.Line, "console ") ||
			strings.HasPrefix(cmd.Line, "system packet-buffer ") ||
			strings.HasPrefix(cmd.Line, "statistics ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	parser := NewSystemParser()
	config, _ := parser.ParseSystemConfig(raw)
	return config
}

// ExtractAdmin extracts admin configuration (login password, administrator password) from parsed config
func (pc *ParsedConfig) ExtractAdmin() *AdminConfig {
	// Build raw config string from global commands that match admin config patterns
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "login password ") ||
			strings.HasPrefix(cmd.Line, "administrator password ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	parser := NewAdminParser()
	config, _ := parser.ParseAdminConfig(raw)
	return config
}

// ExtractAdminUsers extracts admin users (login user, user attribute) from parsed config
func (pc *ParsedConfig) ExtractAdminUsers() []UserConfig {
	// Build raw config string from global commands that match user config patterns
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "login user ") ||
			strings.HasPrefix(cmd.Line, "user attribute ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	parser := NewAdminParser()
	config, _ := parser.ParseAdminConfig(raw)
	return config.Users
}

// ExtractSyslog extracts syslog configuration from parsed config
func (pc *ParsedConfig) ExtractSyslog() *SyslogConfig {
	// Build raw config string from global commands that start with "syslog"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "syslog ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	parser := NewSyslogParser()
	config, _ := parser.ParseSyslogConfig(raw)
	return config
}

// ExtractPasswords extracts all password and secret values from parsed config
func (pc *ParsedConfig) ExtractPasswords() ExtractedPasswords {
	result := ExtractedPasswords{
		Users:    []ExtractedUserAuth{},
		IPsecPSK: []ExtractedIPsecPSK{},
		L2TPAuth: []ExtractedL2TPAuth{},
		PPAuth:   []ExtractedPPAuth{},
	}

	// Patterns for password extraction
	loginPasswordPattern := regexp.MustCompile(`^login\s+password\s+(.+)$`)
	adminPasswordPattern := regexp.MustCompile(`^administrator\s+password\s+(.+)$`)
	loginUserEncryptedPattern := regexp.MustCompile(`^login\s+user\s+(\S+)\s+encrypted\s+(\S+)$`)
	loginUserPlainPattern := regexp.MustCompile(`^login\s+user\s+(\S+)\s+(.+)$`)
	ipsecPSKTextPattern := regexp.MustCompile(`^ipsec\s+ike\s+pre-shared-key\s+(\d+)\s+text\s+(\S+)$`)
	l2tpAuthPattern := regexp.MustCompile(`^l2tp\s+tunnel\s+auth\s+on\s+(\S+)$`)
	ppAuthUsernamePattern := regexp.MustCompile(`^pp\s+auth\s+username\s+(\S+)\s+(.+)$`)

	// Extract from global commands
	for _, cmd := range pc.GetGlobalCommands() {
		// Login password
		if matches := loginPasswordPattern.FindStringSubmatch(cmd.Line); len(matches) >= 2 {
			result.LoginPassword = matches[1]
			continue
		}

		// Admin password
		if matches := adminPasswordPattern.FindStringSubmatch(cmd.Line); len(matches) >= 2 {
			result.AdminPassword = matches[1]
			continue
		}

		// Login user (encrypted)
		if matches := loginUserEncryptedPattern.FindStringSubmatch(cmd.Line); len(matches) >= 3 {
			result.Users = append(result.Users, ExtractedUserAuth{
				Username:  matches[1],
				Password:  "", // Encrypted passwords are not extractable
				Encrypted: true,
			})
			continue
		}

		// Login user (plaintext) - must check after encrypted pattern
		if matches := loginUserPlainPattern.FindStringSubmatch(cmd.Line); len(matches) >= 3 {
			// Skip if already matched as encrypted (contains "encrypted" keyword)
			if strings.Contains(cmd.Line, " encrypted ") {
				continue
			}
			result.Users = append(result.Users, ExtractedUserAuth{
				Username:  matches[1],
				Password:  matches[2],
				Encrypted: false,
			})
			continue
		}
	}

	// Extract from all commands (checking context where relevant)
	// This approach handles nested contexts better than iterating by context
	var currentTunnelID int
	for _, cmd := range pc.Commands {
		// Track current tunnel ID from context
		if cmd.Context != nil && cmd.Context.Type == ContextTunnel {
			currentTunnelID = cmd.Context.ID
		}

		// L2TP tunnel auth (can be in tunnel or ipsec tunnel context)
		if matches := l2tpAuthPattern.FindStringSubmatch(cmd.Line); len(matches) >= 2 {
			// Use the tunnel ID from the command's context, or the tracked one
			tunnelID := currentTunnelID
			if cmd.Context != nil && cmd.Context.Type == ContextTunnel {
				tunnelID = cmd.Context.ID
			}
			result.L2TPAuth = append(result.L2TPAuth, ExtractedL2TPAuth{
				TunnelID: tunnelID,
				Secret:   matches[1],
			})
		}

		// IPsec IKE pre-shared-key text
		if matches := ipsecPSKTextPattern.FindStringSubmatch(cmd.Line); len(matches) >= 3 {
			id, _ := strconv.Atoi(matches[1])
			result.IPsecPSK = append(result.IPsecPSK, ExtractedIPsecPSK{
				ID:        id,
				Secret:    matches[2],
				Encrypted: false,
			})
		}
	}

	// Extract from PP contexts (pp auth username)
	for _, ctx := range pc.Contexts {
		if ctx.Type == ContextPP {
			cmds := pc.GetCommandsInContext(ctx)
			for _, cmd := range cmds {
				// PP auth username
				if matches := ppAuthUsernamePattern.FindStringSubmatch(cmd.Line); len(matches) >= 3 {
					ppAuth := ExtractedPPAuth{
						PPID:     ctx.ID,
						Username: matches[1],
						Password: matches[2],
					}
					if ctx.Name != "" {
						ppAuth.PPName = ctx.Name
					}
					result.PPAuth = append(result.PPAuth, ppAuth)
				}
			}
		}
	}

	return result
}

// ExtractAccessListIPv6 extracts static IPv6 filters from parsed config
func (pc *ParsedConfig) ExtractAccessListIPv6() []IPFilter {
	// Build raw config string from global commands that start with "ipv6 filter"
	// but exclude "ipv6 filter dynamic" lines
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "ipv6 filter ") && !strings.HasPrefix(cmd.Line, "ipv6 filter dynamic ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	filters, _ := ParseIPv6FilterConfig(raw)
	return filters
}

// ExtractIPv6FiltersDynamic extracts dynamic IPv6 filters from parsed config
func (pc *ParsedConfig) ExtractIPv6FiltersDynamic() []IPFilterDynamic {
	// Build raw config string from global commands that start with "ipv6 filter dynamic"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "ipv6 filter dynamic ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	filters, _ := ParseIPv6FilterDynamicConfig(raw)
	return filters
}

// ExtractEthernetFilters extracts Ethernet filters from parsed config
func (pc *ParsedConfig) ExtractEthernetFilters() []EthernetFilter {
	// Build raw config string from global commands that start with "ethernet filter"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "ethernet filter ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	filters, _ := ParseEthernetFilterConfig(raw)
	return filters
}

// ExtractIPv6Prefixes extracts IPv6 prefixes from parsed config
func (pc *ParsedConfig) ExtractIPv6Prefixes() []IPv6Prefix {
	// Build raw config string from global commands that start with "ipv6 prefix"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "ipv6 prefix ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	parser := NewIPv6PrefixParser()
	prefixes, _ := parser.ParseIPv6PrefixConfig(raw)
	return prefixes
}

// ExtractBGP extracts BGP configuration from parsed config using the BGP parser
func (pc *ParsedConfig) ExtractBGP() *BGPConfig {
	// Build raw config string from global commands that start with "bgp"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "bgp ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	parser := NewBGPParser()
	config, _ := parser.ParseBGPConfig(raw)
	return config
}

// ExtractOSPF extracts OSPF configuration from parsed config using the OSPF parser
func (pc *ParsedConfig) ExtractOSPF() *OSPFConfig {
	// Build raw config string from global commands that start with "ospf" or "ip ... ospf"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "ospf ") || strings.Contains(cmd.Line, " ospf ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	parser := NewOSPFParser()
	config, _ := parser.ParseOSPFConfig(raw)
	return config
}

// ExtractPPTP extracts PPTP configuration from parsed config using the PPTP parser
func (pc *ParsedConfig) ExtractPPTP() *PPTPConfig {
	// Build raw config string from global commands that start with "pptp"
	// Also include related pp and ppp commands for complete PPTP config
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "pptp ") {
			lines = append(lines, cmd.Line)
		}
	}

	// Also include pp select anonymous context commands for PPTP
	for _, ctx := range pc.Contexts {
		if ctx.Type == ContextPP && ctx.Name == "anonymous" {
			cmds := pc.GetCommandsInContext(ctx)
			for _, cmd := range cmds {
				lines = append(lines, cmd.Line)
			}
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	parser := NewPPTPParser()
	config, _ := parser.ParsePPTPConfig(raw)
	return config
}

// ExtractSNMPServer extracts SNMP configuration from parsed config using the SNMP parser
func (pc *ParsedConfig) ExtractSNMPServer() *SNMPConfig {
	// Build raw config string from global commands that start with "snmp"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "snmp ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	parser := NewSNMPParser()
	config, _ := parser.ParseSNMPConfig(raw)
	return config
}

// ExtractDHCPBindings extracts DHCP static lease bindings from parsed config
func (pc *ParsedConfig) ExtractDHCPBindings() []DHCPBinding {
	// Build raw config string from global commands that start with "dhcp scope bind"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "dhcp scope bind ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	parser := NewDHCPBindingsParser()
	// ParseBindings requires a scopeID, but we pass 0 since the scopeID is extracted from each line
	bindings, _ := parser.ParseBindings(raw, 0)
	return bindings
}

// ExtractBridges extracts bridge configurations from parsed config
func (pc *ParsedConfig) ExtractBridges() []BridgeConfig {
	// Build raw config string from global commands that start with "bridge member"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "bridge member ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	parser := NewBridgeParser()
	bridges, _ := parser.ParseBridgeConfig(raw)
	return bridges
}

// ExtractNATStatic extracts static NAT descriptors from parsed config
func (pc *ParsedConfig) ExtractNATStatic() []NATStatic {
	// Build raw config string from global commands that match NAT static patterns
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		// Include "nat descriptor type N static" and "nat descriptor static N" lines
		if strings.HasPrefix(cmd.Line, "nat descriptor type ") && strings.Contains(cmd.Line, " static") {
			lines = append(lines, cmd.Line)
		} else if strings.HasPrefix(cmd.Line, "nat descriptor static ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	nats, _ := ParseNATStaticConfig(raw)
	return nats
}

// ExtractIPsecTransports extracts IPsec transport configurations from parsed config
func (pc *ParsedConfig) ExtractIPsecTransports() []IPsecTransport {
	// Build raw config string from global commands that start with "ipsec transport"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "ipsec transport ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	transports, _ := ParseIPsecTransportConfig(raw)
	return transports
}

// ExtractInterfaces extracts interface configurations from parsed config.
// It identifies all interfaces mentioned in the config and returns their configurations.
func (pc *ParsedConfig) ExtractInterfaces() map[string]*InterfaceConfig {
	// First, identify all unique interface names from relevant commands
	interfaceNames := make(map[string]bool)

	// Patterns to extract interface names from various commands
	ipPattern := regexp.MustCompile(`^ip\s+(lan\d+|pp\d+|tunnel\d+|bridge\d+)\s+`)
	descPattern := regexp.MustCompile(`^description\s+(lan\d+|pp\d+|tunnel\d+|bridge\d+)\s+`)
	ethernetPattern := regexp.MustCompile(`^ethernet\s+(lan\d+)\s+filter\s+`)

	for _, cmd := range pc.GetGlobalCommands() {
		// Match ip <interface> ...
		if matches := ipPattern.FindStringSubmatch(cmd.Line); len(matches) >= 2 {
			interfaceNames[matches[1]] = true
		}
		// Match description <interface> ...
		if matches := descPattern.FindStringSubmatch(cmd.Line); len(matches) >= 2 {
			interfaceNames[matches[1]] = true
		}
		// Match ethernet <interface> filter ...
		if matches := ethernetPattern.FindStringSubmatch(cmd.Line); len(matches) >= 2 {
			interfaceNames[matches[1]] = true
		}
	}

	if len(interfaceNames) == 0 {
		return nil
	}

	// For each interface, collect relevant lines and parse
	result := make(map[string]*InterfaceConfig)

	for ifaceName := range interfaceNames {
		// Collect all lines relevant to this interface
		var lines []string
		for _, cmd := range pc.GetGlobalCommands() {
			// Match commands that contain this interface name
			if strings.HasPrefix(cmd.Line, "ip "+ifaceName+" ") ||
				strings.HasPrefix(cmd.Line, "description "+ifaceName+" ") ||
				strings.HasPrefix(cmd.Line, "ethernet "+ifaceName+" ") {
				lines = append(lines, cmd.Line)
			}
		}

		if len(lines) > 0 {
			raw := strings.Join(lines, "\n")
			config, err := ParseInterfaceConfig(raw, ifaceName)
			if err == nil && config != nil {
				result[ifaceName] = config
			}
		}
	}

	return result
}

// ExtractIPsecTunnels extracts IPsec tunnel configurations from parsed config.
// This requires context-aware parsing from "tunnel select N" contexts that have ipsec configuration.
func (pc *ParsedConfig) ExtractIPsecTunnels() []IPsecTunnel {
	var tunnels []IPsecTunnel

	// Iterate through all tunnel contexts
	for _, ctx := range pc.Contexts {
		if ctx.Type != ContextTunnel {
			continue
		}

		cmds := pc.GetCommandsInContext(ctx)
		if len(cmds) == 0 {
			continue
		}

		// Check if this tunnel has IPsec configuration
		hasIPsec := false
		for _, cmd := range cmds {
			if strings.HasPrefix(cmd.Line, "ipsec tunnel ") ||
				strings.HasPrefix(cmd.Line, "tunnel encapsulation ipsec") {
				hasIPsec = true
				break
			}
		}

		if !hasIPsec {
			continue
		}

		// Build raw config string for this tunnel including tunnel select line
		var lines []string
		lines = append(lines, "tunnel select "+strconv.Itoa(ctx.ID))
		for _, cmd := range cmds {
			lines = append(lines, cmd.Line)
		}

		// Also get commands from nested ipsec tunnel contexts
		for _, ipsecCtx := range pc.Contexts {
			if ipsecCtx.Type == ContextIPsecTunnel {
				ipsecCmds := pc.GetCommandsInContext(ipsecCtx)
				for _, cmd := range ipsecCmds {
					lines = append(lines, cmd.Line)
				}
			}
		}

		// Also add global ipsec commands that reference this tunnel ID
		for _, cmd := range pc.GetGlobalCommands() {
			// Match ipsec commands that reference this tunnel ID
			if strings.HasPrefix(cmd.Line, "ipsec ike ") ||
				strings.HasPrefix(cmd.Line, "ipsec sa policy ") {
				// Check if the command references this tunnel ID
				idPattern := regexp.MustCompile(`\b` + strconv.Itoa(ctx.ID) + `\b`)
				if idPattern.MatchString(cmd.Line) {
					lines = append(lines, cmd.Line)
				}
			}
		}

		raw := strings.Join(lines, "\n")
		parser := NewIPsecTunnelParser()
		parsed, err := parser.ParseIPsecTunnelConfig(raw)
		if err == nil && len(parsed) > 0 {
			tunnels = append(tunnels, parsed...)
		}
	}

	return tunnels
}

// ExtractL2TPTunnels extracts L2TP tunnel configurations from parsed config.
// This requires context-aware parsing from "tunnel select N" contexts that have L2TP configuration.
func (pc *ParsedConfig) ExtractL2TPTunnels() []L2TPConfig {
	var tunnels []L2TPConfig

	// Iterate through all tunnel contexts
	for _, ctx := range pc.Contexts {
		if ctx.Type != ContextTunnel {
			continue
		}

		cmds := pc.GetCommandsInContext(ctx)
		if len(cmds) == 0 {
			continue
		}

		// Check if this tunnel has L2TP configuration
		hasL2TP := false
		for _, cmd := range cmds {
			if strings.HasPrefix(cmd.Line, "tunnel encapsulation l2tp") ||
				strings.HasPrefix(cmd.Line, "l2tp ") {
				hasL2TP = true
				break
			}
		}

		if !hasL2TP {
			continue
		}

		// Build raw config string for this tunnel including tunnel select line
		var lines []string
		lines = append(lines, "tunnel select "+strconv.Itoa(ctx.ID))
		for _, cmd := range cmds {
			lines = append(lines, cmd.Line)
		}

		raw := strings.Join(lines, "\n")
		parser := NewL2TPParser()
		parsed, err := parser.ParseL2TPConfig(raw)
		if err == nil && len(parsed) > 0 {
			tunnels = append(tunnels, parsed...)
		}
	}

	// Also check for pp select anonymous context for L2TPv2 LNS configuration
	for _, ctx := range pc.Contexts {
		if ctx.Type == ContextPP && ctx.Name == "anonymous" {
			cmds := pc.GetCommandsInContext(ctx)
			if len(cmds) == 0 {
				continue
			}

			// Build raw config string including pp select anonymous line
			var lines []string
			lines = append(lines, "pp select anonymous")
			for _, cmd := range cmds {
				lines = append(lines, cmd.Line)
			}

			// Check for l2tp service command in global commands
			for _, cmd := range pc.GetGlobalCommands() {
				if strings.HasPrefix(cmd.Line, "l2tp service ") {
					lines = append(lines, cmd.Line)
				}
			}

			raw := strings.Join(lines, "\n")
			parser := NewL2TPParser()
			parsed, err := parser.ParseL2TPConfig(raw)
			if err == nil && len(parsed) > 0 {
				// Only add if not already present (by ID)
				for _, p := range parsed {
					found := false
					for _, existing := range tunnels {
						if existing.ID == p.ID {
							found = true
							break
						}
					}
					if !found {
						tunnels = append(tunnels, p)
					}
				}
			}
		}
	}

	return tunnels
}

// ExtractL2TPService extracts L2TP service configuration from parsed config.
// This extracts the "l2tp service" global command.
func (pc *ParsedConfig) ExtractL2TPService() *L2TPService {
	// Build raw config string from global commands that start with "l2tp service"
	var lines []string
	for _, cmd := range pc.GetGlobalCommands() {
		if strings.HasPrefix(cmd.Line, "l2tp service ") {
			lines = append(lines, cmd.Line)
		}
	}

	if len(lines) == 0 {
		return nil
	}

	raw := strings.Join(lines, "\n")
	service, err := ParseL2TPServiceConfig(raw)
	if err != nil {
		return nil
	}
	return service
}

// ExtractIPv6Interfaces extracts IPv6 interface configurations from parsed config.
// It identifies all interfaces with IPv6 configuration and returns their configurations.
func (pc *ParsedConfig) ExtractIPv6Interfaces() map[string]*IPv6InterfaceConfig {
	// First, identify all unique interface names from ipv6 commands
	interfaceNames := make(map[string]bool)

	// Pattern to extract interface names from ipv6 commands
	ipv6Pattern := regexp.MustCompile(`^ipv6\s+(lan\d+|pp\d+|tunnel\d+|bridge\d+)\s+`)

	for _, cmd := range pc.GetGlobalCommands() {
		if matches := ipv6Pattern.FindStringSubmatch(cmd.Line); len(matches) >= 2 {
			interfaceNames[matches[1]] = true
		}
	}

	if len(interfaceNames) == 0 {
		return nil
	}

	// For each interface, collect relevant lines and parse
	result := make(map[string]*IPv6InterfaceConfig)

	for ifaceName := range interfaceNames {
		// Collect all ipv6 lines relevant to this interface
		var lines []string
		for _, cmd := range pc.GetGlobalCommands() {
			if strings.HasPrefix(cmd.Line, "ipv6 "+ifaceName+" ") {
				lines = append(lines, cmd.Line)
			}
		}

		if len(lines) > 0 {
			raw := strings.Join(lines, "\n")
			config, err := ParseIPv6InterfaceConfig(raw, ifaceName)
			if err == nil && config != nil {
				result[ifaceName] = config
			}
		}
	}

	return result
}
