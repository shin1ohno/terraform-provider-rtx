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
