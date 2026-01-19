package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// PPTPConfig represents PPTP configuration on an RTX router
type PPTPConfig struct {
	Shutdown         bool            `json:"shutdown"`                    // Administratively shut down
	ListenAddress    string          `json:"listen_address,omitempty"`    // Listen IP address
	MaxConnections   int             `json:"max_connections,omitempty"`   // Maximum concurrent connections
	Authentication   *PPTPAuth       `json:"authentication,omitempty"`    // Authentication settings
	Encryption       *PPTPEncryption `json:"encryption,omitempty"`        // MPPE encryption settings
	IPPool           *PPTPIPPool     `json:"ip_pool,omitempty"`           // IP pool for clients
	DisconnectTime   int             `json:"disconnect_time,omitempty"`   // Idle disconnect time
	KeepaliveEnabled bool            `json:"keepalive_enabled,omitempty"` // Keepalive enabled
	Enabled          bool            `json:"enabled"`                     // PPTP service enabled
}

// PPTPAuth represents PPTP authentication configuration
type PPTPAuth struct {
	Method   string `json:"method"`             // pap, chap, mschap, mschap-v2
	Username string `json:"username,omitempty"` // Local username
	Password string `json:"password,omitempty"` // Local password
}

// PPTPEncryption represents PPTP MPPE encryption configuration
type PPTPEncryption struct {
	MPPEBits int  `json:"mppe_bits,omitempty"` // 40, 56, or 128 bits
	Required bool `json:"required,omitempty"`  // Require encryption
}

// PPTPIPPool represents PPTP IP pool configuration
type PPTPIPPool struct {
	Start string `json:"start"` // Start IP address
	End   string `json:"end"`   // End IP address
}

// PPTPParser parses PPTP configuration output
type PPTPParser struct{}

// NewPPTPParser creates a new PPTP parser
func NewPPTPParser() *PPTPParser {
	return &PPTPParser{}
}

// ParsePPTPConfig parses the output of "show config | grep pptp" command
func (p *PPTPParser) ParsePPTPConfig(raw string) (*PPTPConfig, error) {
	config := &PPTPConfig{
		Shutdown: false,
		Enabled:  false,
	}

	lines := strings.Split(raw, "\n")

	// Patterns for PPTP configuration
	pptpServicePattern := regexp.MustCompile(`^\s*pptp\s+service\s+(on|off)\s*$`)
	pptpDisconnectTimePattern := regexp.MustCompile(`^\s*pptp\s+tunnel\s+disconnect\s+time\s+(\d+)\s*$`)
	pptpKeepalivePattern := regexp.MustCompile(`^\s*pptp\s+keepalive\s+use\s+(on|off)\s*$`)
	ppAuthAcceptPattern := regexp.MustCompile(`^\s*pp\s+auth\s+accept\s+(\S+)\s*$`)
	ppAuthMynamePattern := regexp.MustCompile(`^\s*pp\s+auth\s+myname\s+(\S+)\s+(\S+)\s*$`)
	pppCCPTypePattern := regexp.MustCompile(`^\s*ppp\s+ccp\s+type\s+(.+)\s*$`)
	ipPPRemotePoolPattern := regexp.MustCompile(`^\s*ip\s+pp\s+remote\s+address\s+pool\s+([0-9.]+)-([0-9.]+)\s*$`)
	pptpMaxConnectionsPattern := regexp.MustCompile(`^\s*pptp\s+syslog\s+(\d+)\s*$`) // Not exact, need to verify

	var inAnonymousPP bool

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// PPTP service on/off
		if matches := pptpServicePattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.Enabled = matches[1] == "on"
			continue
		}

		// PPTP disconnect time
		if matches := pptpDisconnectTimePattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.DisconnectTime, _ = strconv.Atoi(matches[1])
			continue
		}

		// PPTP keepalive
		if matches := pptpKeepalivePattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.KeepaliveEnabled = matches[1] == "on"
			continue
		}

		// PPTP max connections
		if matches := pptpMaxConnectionsPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.MaxConnections, _ = strconv.Atoi(matches[1])
			continue
		}

		// PP select anonymous context
		if strings.Contains(line, "pp select anonymous") {
			inAnonymousPP = true
			continue
		}

		// PP auth accept (within anonymous PP)
		if matches := ppAuthAcceptPattern.FindStringSubmatch(line); len(matches) >= 2 {
			if config.Authentication == nil {
				config.Authentication = &PPTPAuth{}
			}
			config.Authentication.Method = matches[1]
			continue
		}

		// PP auth myname (within anonymous PP)
		if matches := ppAuthMynamePattern.FindStringSubmatch(line); len(matches) >= 3 {
			if config.Authentication == nil {
				config.Authentication = &PPTPAuth{}
			}
			config.Authentication.Username = matches[1]
			config.Authentication.Password = matches[2]
			continue
		}

		// PPP CCP type (MPPE)
		if matches := pppCCPTypePattern.FindStringSubmatch(line); len(matches) >= 2 {
			ccpType := strings.ToLower(matches[1])
			if config.Encryption == nil {
				config.Encryption = &PPTPEncryption{}
			}
			if strings.Contains(ccpType, "mppe-128") {
				config.Encryption.MPPEBits = 128
			} else if strings.Contains(ccpType, "mppe-56") {
				config.Encryption.MPPEBits = 56
			} else if strings.Contains(ccpType, "mppe-40") {
				config.Encryption.MPPEBits = 40
			} else if strings.Contains(ccpType, "mppe") {
				config.Encryption.MPPEBits = 128 // Default to 128
			}
			if strings.Contains(ccpType, "require") {
				config.Encryption.Required = true
			}
			continue
		}

		// IP PP remote address pool (within anonymous PP)
		if matches := ipPPRemotePoolPattern.FindStringSubmatch(line); len(matches) >= 3 {
			if inAnonymousPP || config.IPPool == nil {
				config.IPPool = &PPTPIPPool{
					Start: matches[1],
					End:   matches[2],
				}
			}
			continue
		}
	}

	return config, nil
}

// BuildPPTPServiceCommand builds the command to enable/disable PPTP service
// Command format: pptp service on/off
func BuildPPTPServiceCommand(enabled bool) string {
	if enabled {
		return "pptp service on"
	}
	return "pptp service off"
}

// BuildPPTPTunnelDisconnectTimeCommand builds the command to set disconnect time
// Command format: pptp tunnel disconnect time <seconds>
func BuildPPTPTunnelDisconnectTimeCommand(seconds int) string {
	return fmt.Sprintf("pptp tunnel disconnect time %d", seconds)
}

// BuildPPTPKeepaliveCommand builds the command to enable/disable keepalive
// Command format: pptp keepalive use on/off
func BuildPPTPKeepaliveCommand(enabled bool) string {
	if enabled {
		return "pptp keepalive use on"
	}
	return "pptp keepalive use off"
}

// BuildPPTPAuthAcceptCommand builds the command to set authentication method
// Command format: pp auth accept <method>
func BuildPPTPAuthAcceptCommand(method string) string {
	return fmt.Sprintf("pp auth accept %s", method)
}

// BuildPPTPAuthMynameCommand builds the command to set credentials
// Command format: pp auth myname <username> <password>
func BuildPPTPAuthMynameCommand(username, password string) string {
	return fmt.Sprintf("pp auth myname %s %s", username, password)
}

// BuildPPPCCPTypeCommand builds the command to set MPPE encryption
// Command format: ppp ccp type mppe-128|mppe-any [require]
func BuildPPPCCPTypeCommand(enc PPTPEncryption) string {
	var ccpType string
	switch enc.MPPEBits {
	case 40:
		ccpType = "mppe-40"
	case 56:
		ccpType = "mppe-56"
	case 128:
		ccpType = "mppe-128"
	default:
		ccpType = "mppe-any"
	}

	if enc.Required {
		return fmt.Sprintf("ppp ccp type %s require", ccpType)
	}
	return fmt.Sprintf("ppp ccp type %s", ccpType)
}

// BuildPPTPIPPoolCommand builds the command to set IP pool
// Command format: ip pp remote address pool <start>-<end>
func BuildPPTPIPPoolCommand(start, end string) string {
	return fmt.Sprintf("ip pp remote address pool %s-%s", start, end)
}

// BuildDeletePPTPCommand builds the commands to disable PPTP
// Returns slice of commands to execute
func BuildDeletePPTPCommand() []string {
	return []string{
		"pptp service off",
		"no pptp tunnel disconnect time",
		"pptp keepalive use off",
	}
}

// BuildShowPPTPConfigCommand builds the command to show PPTP configuration
func BuildShowPPTPConfigCommand() string {
	return "show config | grep pptp"
}

// ValidatePPTPConfig validates a PPTP configuration
func ValidatePPTPConfig(config PPTPConfig) error {
	// Validate authentication
	if config.Authentication != nil {
		validMethods := map[string]bool{
			"pap": true, "chap": true, "mschap": true, "mschap-v2": true,
		}
		if !validMethods[config.Authentication.Method] {
			return fmt.Errorf("invalid authentication method: %s (must be pap, chap, mschap, or mschap-v2)", config.Authentication.Method)
		}
	}

	// Validate encryption
	if config.Encryption != nil {
		validBits := map[int]bool{40: true, 56: true, 128: true, 0: true}
		if !validBits[config.Encryption.MPPEBits] {
			return fmt.Errorf("invalid mppe_bits: %d (must be 40, 56, or 128)", config.Encryption.MPPEBits)
		}
	}

	// Validate IP pool
	if config.IPPool != nil {
		if !isValidIP(config.IPPool.Start) {
			return fmt.Errorf("invalid ip_pool start: %s", config.IPPool.Start)
		}
		if !isValidIP(config.IPPool.End) {
			return fmt.Errorf("invalid ip_pool end: %s", config.IPPool.End)
		}
	}

	// Validate listen address if provided
	if config.ListenAddress != "" && config.ListenAddress != "0.0.0.0" {
		if !isValidIP(config.ListenAddress) {
			return fmt.Errorf("invalid listen_address: %s", config.ListenAddress)
		}
	}

	// Validate disconnect time
	if config.DisconnectTime < 0 {
		return fmt.Errorf("disconnect_time must be non-negative")
	}

	return nil
}
