package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// PPPoEConfig represents PPPoE configuration on an RTX router
type PPPoEConfig struct {
	Number            int                 `json:"number"`                  // PP number (pp select <num>)
	Name              string              `json:"name,omitempty"`          // Description
	Interface         string              `json:"interface"`               // Physical interface (pppoe use <interface>)
	BindInterface     string              `json:"bind_interface"`          // Bind interface (pp bind <interface>)
	ServiceName       string              `json:"service_name"`            // pppoe service-name
	ACName            string              `json:"ac_name,omitempty"`       // pppoe ac-name
	Authentication    *PPPAuth            `json:"authentication"`          // Authentication settings
	AlwaysOn          bool                `json:"always_on"`               // pp always-on on|off
	Enabled           bool                `json:"enabled"`                 // pp enable <num>
	IPConfig          *PPIPConfig         `json:"ip_config"`               // IP configuration
	LCPEchoConfig     *LCPEchoConfig      `json:"lcp_echo,omitempty"`      // LCP echo (keepalive)
	DisconnectTimeout int                 `json:"disconnect_timeout"`      // pp disconnect time
	LCPReconnect      *LCPReconnectConfig `json:"lcp_reconnect,omitempty"` // Reconnect/backoff settings
}

// PPPAuth represents PPP authentication configuration
type PPPAuth struct {
	Method            string `json:"method"`                       // pap, chap, mschap, mschap-v2, accept-pap, accept-chap
	Username          string `json:"username"`                     // pp auth myname <username>
	Password          string `json:"password"`                     // pp auth myname <username> <password>
	EncryptedPassword string `json:"encrypted_password,omitempty"` // For import
}

// PPIPConfig represents PP interface IP configuration
type PPIPConfig struct {
	Address         string `json:"address"`                      // ip pp address <ip>/<mask> or "dhcp"
	MTU             int    `json:"mtu"`                          // ip pp mtu <size>
	TCPMSSLimit     int    `json:"tcp_mss_limit"`                // ip pp tcp mss limit <size>
	NATDescriptor   int    `json:"nat_descriptor"`               // ip pp nat descriptor <id>
	AccessListIPIn  string `json:"access_list_ip_in,omitempty"`  // Inbound IP access list name
	AccessListIPOut string `json:"access_list_ip_out,omitempty"` // Outbound IP access list name
}

// LCPEchoConfig represents LCP echo (keepalive) configuration
type LCPEchoConfig struct {
	Interval   int  `json:"interval"`    // LCP echo interval
	MaxRetries int  `json:"max_retries"` // Maximum retries
	Enabled    bool `json:"enabled"`     // Keepalive enabled
}

// LCPReconnectConfig represents reconnect/backoff settings
type LCPReconnectConfig struct {
	ReconnectInterval int `json:"reconnect_interval"` // Seconds between reconnect attempts
	ReconnectAttempts int `json:"reconnect_attempts"` // Max attempts (0 = unlimited)
}

// BuildPPReconnectCommand builds pp keepalive interval/retry commands
func BuildPPReconnectCommand(interval int, attempts int) string {
	if interval <= 0 {
		return ""
	}
	return fmt.Sprintf("pp keepalive interval %d retry-interval %d", interval, attempts)
}

// PPPParser parses PPP/PPPoE configuration output
type PPPParser struct{}

// NewPPPParser creates a new PPP parser
func NewPPPParser() *PPPParser {
	return &PPPParser{}
}

// ParsePPPoEConfig parses the output of "show config" for PPPoE configuration
func (p *PPPParser) ParsePPPoEConfig(raw string) ([]PPPoEConfig, error) {
	configs := make(map[int]*PPPoEConfig)
	lines := strings.Split(raw, "\n")

	// Patterns for PP interface commands
	ppSelectPattern := regexp.MustCompile(`^\s*pp\s+select\s+(\d+)\s*$`)
	ppDescriptionPattern := regexp.MustCompile(`^\s*description\s+(pp\s+)?(.+)\s*$`)
	pppoeUsePattern := regexp.MustCompile(`^\s*pppoe\s+use\s+(\S+)\s*$`)
	ppBindPattern := regexp.MustCompile(`^\s*pp\s+bind\s+(\S+)\s*$`)
	pppoeServiceNamePattern := regexp.MustCompile(`^\s*pppoe\s+service-name\s+(.+)\s*$`)
	pppoeACNamePattern := regexp.MustCompile(`^\s*pppoe\s+ac-name\s+(.+)\s*$`)
	ppAuthAcceptPattern := regexp.MustCompile(`^\s*pp\s+auth\s+accept\s+(\S+)(?:\s+(\S+))?\s*$`)
	ppAuthMynamePattern := regexp.MustCompile(`^\s*pp\s+auth\s+myname\s+(\S+)\s+(\S+)\s*$`)
	ppAlwaysOnPattern := regexp.MustCompile(`^\s*pp\s+always-on\s+(on|off)\s*$`)
	ppEnablePattern := regexp.MustCompile(`^\s*pp\s+enable\s+(\d+)\s*$`)
	ppDisconnectTimePattern := regexp.MustCompile(`^\s*pp\s+disconnect\s+time\s+(off|\d+)\s*$`)
	ppReconnectPattern := regexp.MustCompile(`^\s*pp\s+keepalive\s+interval\s+(\d+)\s+retry-interval\s+(\d+)\s*$`)

	// IP PP patterns
	ipPPAddressPattern := regexp.MustCompile(`^\s*ip\s+pp\s+address\s+(\S+)\s*$`)
	ipPPMTUPattern := regexp.MustCompile(`^\s*ip\s+pp\s+mtu\s+(\d+)\s*$`)
	ipPPTCPMSSPattern := regexp.MustCompile(`^\s*ip\s+pp\s+tcp\s+mss\s+limit\s+(\d+)\s*$`)
	ipPPNATDescriptorPattern := regexp.MustCompile(`^\s*ip\s+pp\s+nat\s+descriptor\s+(\d+)\s*$`)
	ipPPSecureFilterInPattern := regexp.MustCompile(`^\s*ip\s+pp\s+secure\s+filter\s+in\s+(.+)\s*$`)
	ipPPSecureFilterOutPattern := regexp.MustCompile(`^\s*ip\s+pp\s+secure\s+filter\s+out\s+(.+)\s*$`)

	// LCP echo patterns
	pppLcpMruPattern := regexp.MustCompile(`^\s*ppp\s+lcp\s+mru\s+on\s+(\d+)\s*$`)
	pppIpcpIPAddressPattern := regexp.MustCompile(`^\s*ppp\s+ipcp\s+ipaddress\s+(on|off)\s*$`)
	pppCcpPattern := regexp.MustCompile(`^\s*ppp\s+ccp\s+type\s+(\S+)\s*$`)

	var currentPPNum int
	var currentConfig *PPPoEConfig

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}

		// PP select
		if matches := ppSelectPattern.FindStringSubmatch(line); len(matches) >= 2 {
			// Save previous config if exists
			if currentConfig != nil && currentPPNum > 0 {
				configs[currentPPNum] = currentConfig
			}

			ppNum, _ := strconv.Atoi(matches[1])
			currentPPNum = ppNum

			// Create new config
			currentConfig = &PPPoEConfig{
				Number:   ppNum,
				IPConfig: &PPIPConfig{},
			}
			continue
		}

		// Skip if no current PP context
		if currentConfig == nil {
			continue
		}

		// Description
		if matches := ppDescriptionPattern.FindStringSubmatch(line); len(matches) >= 3 {
			currentConfig.Name = strings.TrimSpace(matches[2])
			continue
		}

		// PPPoE use
		if matches := pppoeUsePattern.FindStringSubmatch(line); len(matches) >= 2 {
			currentConfig.Interface = matches[1]
			continue
		}

		// PP bind
		if matches := ppBindPattern.FindStringSubmatch(line); len(matches) >= 2 {
			currentConfig.BindInterface = matches[1]
			continue
		}

		// PPPoE service-name
		if matches := pppoeServiceNamePattern.FindStringSubmatch(line); len(matches) >= 2 {
			currentConfig.ServiceName = strings.TrimSpace(matches[1])
			continue
		}

		// PPPoE ac-name
		if matches := pppoeACNamePattern.FindStringSubmatch(line); len(matches) >= 2 {
			currentConfig.ACName = strings.TrimSpace(matches[1])
			continue
		}

		// PP auth accept
		if matches := ppAuthAcceptPattern.FindStringSubmatch(line); len(matches) >= 2 {
			if currentConfig.Authentication == nil {
				currentConfig.Authentication = &PPPAuth{}
			}
			currentConfig.Authentication.Method = matches[1]
			continue
		}

		// PP auth myname
		if matches := ppAuthMynamePattern.FindStringSubmatch(line); len(matches) >= 3 {
			if currentConfig.Authentication == nil {
				currentConfig.Authentication = &PPPAuth{}
			}
			currentConfig.Authentication.Username = matches[1]
			currentConfig.Authentication.Password = matches[2]
			continue
		}

		// PP always-on
		if matches := ppAlwaysOnPattern.FindStringSubmatch(line); len(matches) >= 2 {
			currentConfig.AlwaysOn = matches[1] == "on"
			continue
		}

		// PP disconnect time
		if matches := ppDisconnectTimePattern.FindStringSubmatch(line); len(matches) >= 2 {
			if matches[1] != "off" {
				currentConfig.DisconnectTimeout, _ = strconv.Atoi(matches[1])
			}
			continue
		}

		// PP reconnect/backoff (reuse keepalive syntax)
		if matches := ppReconnectPattern.FindStringSubmatch(line); len(matches) == 3 {
			if currentConfig.LCPReconnect == nil {
				currentConfig.LCPReconnect = &LCPReconnectConfig{}
			}
			currentConfig.LCPReconnect.ReconnectInterval, _ = strconv.Atoi(matches[1])
			currentConfig.LCPReconnect.ReconnectAttempts, _ = strconv.Atoi(matches[2])
			continue
		}

		// IP PP address
		if matches := ipPPAddressPattern.FindStringSubmatch(line); len(matches) >= 2 {
			if currentConfig.IPConfig == nil {
				currentConfig.IPConfig = &PPIPConfig{}
			}
			currentConfig.IPConfig.Address = matches[1]
			continue
		}

		// IP PP MTU
		if matches := ipPPMTUPattern.FindStringSubmatch(line); len(matches) >= 2 {
			if currentConfig.IPConfig == nil {
				currentConfig.IPConfig = &PPIPConfig{}
			}
			currentConfig.IPConfig.MTU, _ = strconv.Atoi(matches[1])
			continue
		}

		// IP PP TCP MSS limit
		if matches := ipPPTCPMSSPattern.FindStringSubmatch(line); len(matches) >= 2 {
			if currentConfig.IPConfig == nil {
				currentConfig.IPConfig = &PPIPConfig{}
			}
			currentConfig.IPConfig.TCPMSSLimit, _ = strconv.Atoi(matches[1])
			continue
		}

		// IP PP NAT descriptor
		if matches := ipPPNATDescriptorPattern.FindStringSubmatch(line); len(matches) >= 2 {
			if currentConfig.IPConfig == nil {
				currentConfig.IPConfig = &PPIPConfig{}
			}
			currentConfig.IPConfig.NATDescriptor, _ = strconv.Atoi(matches[1])
			continue
		}

		// IP PP secure filter in (access list name)
		if matches := ipPPSecureFilterInPattern.FindStringSubmatch(line); len(matches) >= 2 {
			if currentConfig.IPConfig == nil {
				currentConfig.IPConfig = &PPIPConfig{}
			}
			currentConfig.IPConfig.AccessListIPIn = strings.TrimSpace(matches[1])
			continue
		}

		// IP PP secure filter out (access list name)
		if matches := ipPPSecureFilterOutPattern.FindStringSubmatch(line); len(matches) >= 2 {
			if currentConfig.IPConfig == nil {
				currentConfig.IPConfig = &PPIPConfig{}
			}
			currentConfig.IPConfig.AccessListIPOut = strings.TrimSpace(matches[1])
			continue
		}

		// Ignore other ppp settings for now (lcp mru, ipcp, ccp)
		_ = pppLcpMruPattern
		_ = pppIpcpIPAddressPattern
		_ = pppCcpPattern
	}

	// Save last config
	if currentConfig != nil && currentPPNum > 0 {
		configs[currentPPNum] = currentConfig
	}

	// Check for pp enable commands (second pass - after all configs are stored)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if matches := ppEnablePattern.FindStringSubmatch(line); len(matches) >= 2 {
			ppNum, _ := strconv.Atoi(matches[1])
			if cfg, exists := configs[ppNum]; exists {
				cfg.Enabled = true
			}
		}
	}

	// Convert map to slice
	result := make([]PPPoEConfig, 0, len(configs))
	for _, cfg := range configs {
		result = append(result, *cfg)
	}

	return result, nil
}

// ParsePPInterfaceConfig parses PP interface specific settings
func (p *PPPParser) ParsePPInterfaceConfig(raw string, ppNum int) (*PPIPConfig, error) {
	config := &PPIPConfig{}
	lines := strings.Split(raw, "\n")

	inPPContext := false
	ppSelectPattern := regexp.MustCompile(fmt.Sprintf(`^\s*pp\s+select\s+%d\s*$`, ppNum))
	ppSelectOtherPattern := regexp.MustCompile(`^\s*pp\s+select\s+\d+\s*$`)

	ipPPAddressPattern := regexp.MustCompile(`^\s*ip\s+pp\s+address\s+(\S+)\s*$`)
	ipPPMTUPattern := regexp.MustCompile(`^\s*ip\s+pp\s+mtu\s+(\d+)\s*$`)
	ipPPTCPMSSPattern := regexp.MustCompile(`^\s*ip\s+pp\s+tcp\s+mss\s+limit\s+(\d+)\s*$`)
	ipPPNATDescriptorPattern := regexp.MustCompile(`^\s*ip\s+pp\s+nat\s+descriptor\s+(\d+)\s*$`)
	ipPPSecureFilterInPattern := regexp.MustCompile(`^\s*ip\s+pp\s+secure\s+filter\s+in\s+(.+)\s*$`)
	ipPPSecureFilterOutPattern := regexp.MustCompile(`^\s*ip\s+pp\s+secure\s+filter\s+out\s+(.+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if we enter the target PP context
		if ppSelectPattern.MatchString(line) {
			inPPContext = true
			continue
		}

		// Check if we leave the PP context
		if inPPContext && ppSelectOtherPattern.MatchString(line) {
			break
		}

		if !inPPContext {
			continue
		}

		// Parse IP PP commands
		if matches := ipPPAddressPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.Address = matches[1]
			continue
		}

		if matches := ipPPMTUPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.MTU, _ = strconv.Atoi(matches[1])
			continue
		}

		if matches := ipPPTCPMSSPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.TCPMSSLimit, _ = strconv.Atoi(matches[1])
			continue
		}

		if matches := ipPPNATDescriptorPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.NATDescriptor, _ = strconv.Atoi(matches[1])
			continue
		}

		if matches := ipPPSecureFilterInPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.AccessListIPIn = strings.TrimSpace(matches[1])
			continue
		}

		if matches := ipPPSecureFilterOutPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.AccessListIPOut = strings.TrimSpace(matches[1])
			continue
		}
	}

	return config, nil
}

// ============================================================================
// Command Builders
// ============================================================================

// BuildPPSelectCommand builds "pp select <num>" command
func BuildPPSelectCommand(ppNum int) string {
	if ppNum < 1 {
		return ""
	}
	return fmt.Sprintf("pp select %d", ppNum)
}

// BuildPPDescriptionCommand builds "description" command for PP interface
func BuildPPDescriptionCommand(description string) string {
	if description == "" {
		return ""
	}
	return fmt.Sprintf("description pp %s", description)
}

// BuildPPPoEUseCommand builds "pppoe use <interface>" command
func BuildPPPoEUseCommand(iface string) string {
	if iface == "" {
		return ""
	}
	return fmt.Sprintf("pppoe use %s", iface)
}

// BuildPPBindCommand builds "pp bind <interface>" command
func BuildPPBindCommand(iface string) string {
	if iface == "" {
		return ""
	}
	return fmt.Sprintf("pp bind %s", iface)
}

// BuildPPPoEServiceNameCommand builds "pppoe service-name <name>" command
func BuildPPPoEServiceNameCommand(serviceName string) string {
	if serviceName == "" {
		return ""
	}
	return fmt.Sprintf("pppoe service-name %s", serviceName)
}

// BuildPPPoEACNameCommand builds "pppoe ac-name <name>" command
func BuildPPPoEACNameCommand(acName string) string {
	if acName == "" {
		return ""
	}
	return fmt.Sprintf("pppoe ac-name %s", acName)
}

// BuildPPPAuthAcceptCommand builds "pp auth accept <method>" command
func BuildPPPAuthAcceptCommand(method string) string {
	if method == "" {
		return ""
	}
	return fmt.Sprintf("pp auth accept %s", method)
}

// BuildPPPAuthMynameCommand builds "pp auth myname <username> <password>" command
func BuildPPPAuthMynameCommand(username, password string) string {
	if username == "" || password == "" {
		return ""
	}
	return fmt.Sprintf("pp auth myname %s %s", username, password)
}

// BuildPPAlwaysOnCommand builds "pp always-on <on|off>" command
func BuildPPAlwaysOnCommand(enabled bool) string {
	if enabled {
		return "pp always-on on"
	}
	return "pp always-on off"
}

// BuildPPDisconnectTimeCommand builds "pp disconnect time <seconds>" command
func BuildPPDisconnectTimeCommand(seconds int) string {
	if seconds <= 0 {
		return "pp disconnect time off"
	}
	return fmt.Sprintf("pp disconnect time %d", seconds)
}

// BuildPPEnableCommand builds "pp enable <num>" command
func BuildPPEnableCommand(ppNum int) string {
	if ppNum < 1 {
		return ""
	}
	return fmt.Sprintf("pp enable %d", ppNum)
}

// BuildPPDisableCommand builds "pp disable <num>" command
func BuildPPDisableCommand(ppNum int) string {
	if ppNum < 1 {
		return ""
	}
	return fmt.Sprintf("pp disable %d", ppNum)
}

// BuildIPPPAddressCommand builds "ip pp address <address>" command
func BuildIPPPAddressCommand(address string) string {
	if address == "" {
		return ""
	}
	return fmt.Sprintf("ip pp address %s", address)
}

// BuildIPPPMTUCommand builds "ip pp mtu <size>" command
func BuildIPPPMTUCommand(mtu int) string {
	if mtu <= 0 {
		return ""
	}
	return fmt.Sprintf("ip pp mtu %d", mtu)
}

// BuildIPPPTCPMSSLimitCommand builds "ip pp tcp mss limit <size>" command
func BuildIPPPTCPMSSLimitCommand(mss int) string {
	if mss <= 0 {
		return ""
	}
	return fmt.Sprintf("ip pp tcp mss limit %d", mss)
}

// BuildIPPPNATDescriptorCommand builds "ip pp nat descriptor <id>" command
func BuildIPPPNATDescriptorCommand(descriptorID int) string {
	if descriptorID <= 0 {
		return ""
	}
	return fmt.Sprintf("ip pp nat descriptor %d", descriptorID)
}

// BuildIPPPSecureFilterInCommand builds "ip pp secure filter in <name>" command
func BuildIPPPSecureFilterInCommand(accessListName string) string {
	if accessListName == "" {
		return ""
	}
	return fmt.Sprintf("ip pp secure filter in %s", accessListName)
}

// BuildIPPPSecureFilterOutCommand builds "ip pp secure filter out <name>" command
func BuildIPPPSecureFilterOutCommand(accessListName string) string {
	if accessListName == "" {
		return ""
	}
	return fmt.Sprintf("ip pp secure filter out %s", accessListName)
}

// BuildDeleteIPPPAddressCommand builds "no ip pp address" command
func BuildDeleteIPPPAddressCommand() string {
	return "no ip pp address"
}

// BuildDeleteIPPPMTUCommand builds "no ip pp mtu" command
func BuildDeleteIPPPMTUCommand() string {
	return "no ip pp mtu"
}

// BuildDeleteIPPPNATDescriptorCommand builds "no ip pp nat descriptor" command
func BuildDeleteIPPPNATDescriptorCommand() string {
	return "no ip pp nat descriptor"
}

// BuildDeleteIPPPSecureFilterInCommand builds "no ip pp secure filter in" command
func BuildDeleteIPPPSecureFilterInCommand() string {
	return "no ip pp secure filter in"
}

// BuildDeleteIPPPSecureFilterOutCommand builds "no ip pp secure filter out" command
func BuildDeleteIPPPSecureFilterOutCommand() string {
	return "no ip pp secure filter out"
}

// ============================================================================
// Full Command Builders
// ============================================================================

// BuildPPPoECommand builds all commands for a PPPoE configuration
func BuildPPPoECommand(config PPPoEConfig) []string {
	var commands []string

	// pp select
	if cmd := BuildPPSelectCommand(config.Number); cmd != "" {
		commands = append(commands, cmd)
	}

	// description
	if cmd := BuildPPDescriptionCommand(config.Name); cmd != "" {
		commands = append(commands, cmd)
	}

	// pppoe use
	if cmd := BuildPPPoEUseCommand(config.Interface); cmd != "" {
		commands = append(commands, cmd)
	}

	// pp bind
	if cmd := BuildPPBindCommand(config.BindInterface); cmd != "" {
		commands = append(commands, cmd)
	}

	// pppoe service-name
	if cmd := BuildPPPoEServiceNameCommand(config.ServiceName); cmd != "" {
		commands = append(commands, cmd)
	}

	// pppoe ac-name
	if cmd := BuildPPPoEACNameCommand(config.ACName); cmd != "" {
		commands = append(commands, cmd)
	}

	// pp auth accept
	if config.Authentication != nil && config.Authentication.Method != "" {
		if cmd := BuildPPPAuthAcceptCommand(config.Authentication.Method); cmd != "" {
			commands = append(commands, cmd)
		}
	}

	// pp auth myname
	if config.Authentication != nil && config.Authentication.Username != "" {
		if cmd := BuildPPPAuthMynameCommand(config.Authentication.Username, config.Authentication.Password); cmd != "" {
			commands = append(commands, cmd)
		}
	}

	// pp always-on
	commands = append(commands, BuildPPAlwaysOnCommand(config.AlwaysOn))

	// pp disconnect time
	if config.DisconnectTimeout > 0 {
		commands = append(commands, BuildPPDisconnectTimeCommand(config.DisconnectTimeout))
	}

	// IP PP settings
	if config.IPConfig != nil {
		// ip pp address
		if cmd := BuildIPPPAddressCommand(config.IPConfig.Address); cmd != "" {
			commands = append(commands, cmd)
		}

		// ip pp mtu
		if cmd := BuildIPPPMTUCommand(config.IPConfig.MTU); cmd != "" {
			commands = append(commands, cmd)
		}

		// ip pp tcp mss limit
		if cmd := BuildIPPPTCPMSSLimitCommand(config.IPConfig.TCPMSSLimit); cmd != "" {
			commands = append(commands, cmd)
		}

		// ip pp nat descriptor
		if cmd := BuildIPPPNATDescriptorCommand(config.IPConfig.NATDescriptor); cmd != "" {
			commands = append(commands, cmd)
		}

		// ip pp secure filter in
		if cmd := BuildIPPPSecureFilterInCommand(config.IPConfig.AccessListIPIn); cmd != "" {
			commands = append(commands, cmd)
		}

		// ip pp secure filter out
		if cmd := BuildIPPPSecureFilterOutCommand(config.IPConfig.AccessListIPOut); cmd != "" {
			commands = append(commands, cmd)
		}
	}

	// pp enable
	if config.Enabled {
		if cmd := BuildPPEnableCommand(config.Number); cmd != "" {
			commands = append(commands, cmd)
		}
	}

	// reconnect/backoff
	if config.LCPReconnect != nil && config.LCPReconnect.ReconnectInterval > 0 {
		if cmd := BuildPPReconnectCommand(config.LCPReconnect.ReconnectInterval, config.LCPReconnect.ReconnectAttempts); cmd != "" {
			commands = append(commands, cmd)
		}
	}

	return commands
}

// BuildDeletePPPoECommand builds commands to delete a PPPoE configuration
func BuildDeletePPPoECommand(ppNum int) []string {
	if ppNum < 1 {
		return nil
	}

	return []string{
		BuildPPDisableCommand(ppNum),
		BuildPPSelectCommand(ppNum),
		"no description",
		"no pppoe use",
		"no pp bind",
		"no pppoe service-name",
		"no pp auth accept",
		"no pp auth myname",
		"pp always-on off",
		"no ip pp address",
		"no ip pp mtu",
		"no ip pp nat descriptor",
		"no ip pp secure filter in",
		"no ip pp secure filter out",
	}
}

// BuildShowPPConfigCommand builds "show config pp <num>" command
func BuildShowPPConfigCommand(ppNum int) string {
	if ppNum < 1 {
		return "show config"
	}
	return fmt.Sprintf("show config | grep -A100 'pp select %d'", ppNum)
}

// ============================================================================
// Validation
// ============================================================================

// ValidatePPPoEConfig validates a PPPoE configuration
func ValidatePPPoEConfig(config PPPoEConfig) error {
	if config.Number < 1 {
		return fmt.Errorf("PP number must be >= 1")
	}

	if config.Interface == "" && config.BindInterface == "" {
		return fmt.Errorf("either interface or bind_interface must be specified")
	}

	// Validate interface name
	if config.Interface != "" {
		if !isValidPPPoEInterface(config.Interface) {
			return fmt.Errorf("invalid PPPoE interface: %s", config.Interface)
		}
	}

	// Validate authentication
	if config.Authentication != nil {
		if config.Authentication.Method != "" {
			validMethods := []string{"pap", "chap", "mschap", "mschap-v2", "none"}
			found := false
			for _, m := range validMethods {
				if config.Authentication.Method == m {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("invalid authentication method: %s", config.Authentication.Method)
			}
		}

		if config.Authentication.Username != "" && config.Authentication.Password == "" {
			return fmt.Errorf("password is required when username is specified")
		}
	}

	// Validate IP config
	if config.IPConfig != nil {
		if config.IPConfig.MTU > 0 && (config.IPConfig.MTU < 64 || config.IPConfig.MTU > 1500) {
			return fmt.Errorf("MTU must be between 64 and 1500: %d", config.IPConfig.MTU)
		}

		if config.IPConfig.TCPMSSLimit > 0 && (config.IPConfig.TCPMSSLimit < 1 || config.IPConfig.TCPMSSLimit > 1460) {
			return fmt.Errorf("TCP MSS limit must be between 1 and 1460: %d", config.IPConfig.TCPMSSLimit)
		}
	}

	if config.LCPReconnect != nil {
		if config.LCPReconnect.ReconnectInterval < 0 {
			return fmt.Errorf("reconnect interval must be >= 0")
		}
		if config.LCPReconnect.ReconnectAttempts < 0 {
			return fmt.Errorf("reconnect attempts must be >= 0")
		}
	}

	return nil
}

// isValidPPPoEInterface checks if an interface name is valid for PPPoE
func isValidPPPoEInterface(iface string) bool {
	// Valid interfaces: lan1, lan2, lan3, lan4, etc.
	pattern := regexp.MustCompile(`^lan\d+$`)
	return pattern.MatchString(iface)
}

// ValidatePPIPConfig validates PP interface IP configuration
func ValidatePPIPConfig(config PPIPConfig) error {
	if config.MTU > 0 && (config.MTU < 64 || config.MTU > 1500) {
		return fmt.Errorf("MTU must be between 64 and 1500: %d", config.MTU)
	}

	if config.TCPMSSLimit > 0 && (config.TCPMSSLimit < 1 || config.TCPMSSLimit > 1460) {
		return fmt.Errorf("TCP MSS limit must be between 1 and 1460: %d", config.TCPMSSLimit)
	}

	if config.NATDescriptor > 0 && (config.NATDescriptor < 1 || config.NATDescriptor > 65535) {
		return fmt.Errorf("NAT descriptor ID must be between 1 and 65535: %d", config.NATDescriptor)
	}

	return nil
}
