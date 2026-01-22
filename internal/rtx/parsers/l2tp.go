package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// L2TPConfig represents L2TP/L2TPv3 configuration on an RTX router
type L2TPConfig struct {
	ID               int            `json:"id"`                          // Tunnel ID
	Name             string         `json:"name,omitempty"`              // Description
	Version          string         `json:"version"`                     // "l2tp" (v2) or "l2tpv3" (v3)
	Mode             string         `json:"mode"`                        // "lns" (L2TPv2 server) or "l2vpn" (L2TPv3)
	Shutdown         bool           `json:"shutdown"`                    // Administratively shut down
	TunnelSource     string         `json:"tunnel_source"`               // Source IP/interface
	TunnelDest       string         `json:"tunnel_dest"`                 // Destination IP/FQDN
	TunnelDestType   string         `json:"tunnel_dest_type,omitempty"`  // "ip" or "fqdn"
	Authentication   *L2TPAuth      `json:"authentication,omitempty"`    // L2TPv2 authentication
	IPPool           *L2TPIPPool    `json:"ip_pool,omitempty"`           // L2TPv2 IP pool
	IPsecProfile     *L2TPIPsec     `json:"ipsec_profile,omitempty"`     // IPsec encryption
	L2TPv3Config     *L2TPv3Config  `json:"l2tpv3_config,omitempty"`     // L2TPv3-specific config
	KeepaliveEnabled bool           `json:"keepalive_enabled,omitempty"` // Keepalive enabled
	KeepaliveConfig  *L2TPKeepalive `json:"keepalive_config,omitempty"`  // Keepalive settings
	DisconnectTime   int            `json:"disconnect_time,omitempty"`   // Idle disconnect time
	AlwaysOn         bool           `json:"always_on,omitempty"`         // Always-on mode
	Enabled          bool           `json:"enabled"`                     // Service enabled
}

// L2TPAuth represents L2TPv2 authentication configuration
type L2TPAuth struct {
	Method   string `json:"method"`             // pap, chap, mschap, mschap-v2
	Username string `json:"username,omitempty"` // Local username
	Password string `json:"password,omitempty"` // Local password
}

// L2TPIPPool represents L2TPv2 IP pool configuration
type L2TPIPPool struct {
	Start string `json:"start"` // Start IP address
	End   string `json:"end"`   // End IP address
}

// L2TPIPsec represents L2TP over IPsec configuration
type L2TPIPsec struct {
	Enabled      bool   `json:"enabled"`                  // IPsec enabled
	PreSharedKey string `json:"pre_shared_key,omitempty"` // IPsec PSK
	TunnelID     int    `json:"tunnel_id,omitempty"`      // Associated IPsec tunnel ID
}

// L2TPv3Config represents L2TPv3-specific configuration
type L2TPv3Config struct {
	LocalRouterID   string          `json:"local_router_id"`            // Local router ID
	RemoteRouterID  string          `json:"remote_router_id"`           // Remote router ID
	RemoteEndID     string          `json:"remote_end_id,omitempty"`    // Remote end ID (hostname)
	SessionID       int             `json:"session_id,omitempty"`       // Session ID
	CookieSize      int             `json:"cookie_size,omitempty"`      // Cookie size (0, 4, 8)
	BridgeInterface string          `json:"bridge_interface,omitempty"` // Bridge interface
	TunnelAuth      *L2TPTunnelAuth `json:"tunnel_auth,omitempty"`      // Tunnel authentication
}

// L2TPTunnelAuth represents L2TPv3 tunnel authentication
type L2TPTunnelAuth struct {
	Enabled  bool   `json:"enabled"`            // Tunnel auth enabled
	Password string `json:"password,omitempty"` // Tunnel auth password
}

// L2TPKeepalive represents L2TP keepalive configuration
type L2TPKeepalive struct {
	Interval int `json:"interval"` // Keepalive interval
	Retry    int `json:"retry"`    // Retry count
}

// L2TPParser parses L2TP configuration output
type L2TPParser struct{}

// NewL2TPParser creates a new L2TP parser
func NewL2TPParser() *L2TPParser {
	return &L2TPParser{}
}

// ParseL2TPConfig parses the output of "show config" for L2TP configuration
func (p *L2TPParser) ParseL2TPConfig(raw string) ([]L2TPConfig, error) {
	tunnels := make(map[int]*L2TPConfig)
	lines := strings.Split(raw, "\n")

	// L2TPv2 LNS patterns
	l2tpServicePattern := regexp.MustCompile(`^\s*l2tp\s+service\s+(on|off)\s*$`)
	ppSelectAnonymousPattern := regexp.MustCompile(`^\s*pp\s+select\s+anonymous\s*$`)
	ppBindTunnelPattern := regexp.MustCompile(`^\s*pp\s+bind\s+tunnel(\d+)\s*$`)
	ppAuthAcceptPattern := regexp.MustCompile(`^\s*pp\s+auth\s+accept\s+(\S+)\s*$`)
	ppAuthMynamePattern := regexp.MustCompile(`^\s*pp\s+auth\s+myname\s+(\S+)\s+(\S+)\s*$`)
	ipPPRemotePoolPattern := regexp.MustCompile(`^\s*ip\s+pp\s+remote\s+address\s+pool\s+([0-9.]+)-([0-9.]+)\s*$`)

	// L2TPv3 patterns
	tunnelSelectPattern := regexp.MustCompile(`^\s*tunnel\s+select\s+(\d+)\s*$`)
	tunnelEncapsulationPattern := regexp.MustCompile(`^\s*tunnel\s+encapsulation\s+(l2tp|l2tpv3)\s*$`)
	tunnelEndpointPattern := regexp.MustCompile(`^\s*tunnel\s+endpoint\s+address\s+([0-9.]+)\s+([0-9.]+)\s*$`)
	tunnelEndpointNamePattern := regexp.MustCompile(`^\s*tunnel\s+endpoint\s+name\s+(\S+)\s+(fqdn|ip)\s*$`)
	ipsecTunnelPattern := regexp.MustCompile(`^\s*ipsec\s+tunnel\s+(\d+)\s*$`)
	l2tpLocalRouterIDPattern := regexp.MustCompile(`^\s*l2tp\s+local\s+router-id\s+([0-9.]+)\s*$`)
	l2tpRemoteRouterIDPattern := regexp.MustCompile(`^\s*l2tp\s+remote\s+router-id\s+([0-9.]+)\s*$`)
	l2tpRemoteEndIDPattern := regexp.MustCompile(`^\s*l2tp\s+remote\s+end-id\s+(\S+)\s*$`)
	l2tpAlwaysOnPattern := regexp.MustCompile(`^\s*l2tp\s+always-on\s+(on|off)\s*$`)
	l2tpHostnamePattern := regexp.MustCompile(`^\s*l2tp\s+hostname\s+(\S+)\s*$`)
	l2tpTunnelAuthPattern := regexp.MustCompile(`^\s*l2tp\s+tunnel\s+auth\s+(on|off)(?:\s+(\S+))?\s*$`)
	l2tpKeepalivePattern := regexp.MustCompile(`^\s*l2tp\s+keepalive\s+use\s+on\s+(\d+)\s+(\d+)\s*$`)
	l2tpDisconnectTimePattern := regexp.MustCompile(`^\s*l2tp\s+tunnel\s+disconnect\s+time\s+(off|\d+)\s*$`)
	tunnelDescriptionPattern := regexp.MustCompile(`^\s*description\s+(.+)\s*$`)
	tunnelEnablePattern := regexp.MustCompile(`^\s*tunnel\s+enable\s+(\d+)\s*$`)

	var currentTunnelID int
	var l2tpServiceOn bool
	var inAnonymousPP bool
	var currentAnonymousConfig *L2TPConfig

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// L2TP service
		if matches := l2tpServicePattern.FindStringSubmatch(line); len(matches) >= 2 {
			l2tpServiceOn = matches[1] == "on"
			continue
		}

		// PP select anonymous (L2TPv2 LNS)
		if ppSelectAnonymousPattern.MatchString(line) {
			inAnonymousPP = true
			if currentAnonymousConfig == nil {
				currentAnonymousConfig = &L2TPConfig{
					ID:      0, // Anonymous PP
					Version: "l2tp",
					Mode:    "lns",
					Enabled: l2tpServiceOn,
				}
			}
			continue
		}

		// PP bind tunnel
		if matches := ppBindTunnelPattern.FindStringSubmatch(line); len(matches) >= 2 && inAnonymousPP {
			tunnelID, _ := strconv.Atoi(matches[1])
			if currentAnonymousConfig != nil {
				currentAnonymousConfig.ID = tunnelID
			}
			continue
		}

		// PP auth accept
		if matches := ppAuthAcceptPattern.FindStringSubmatch(line); len(matches) >= 2 && inAnonymousPP {
			if currentAnonymousConfig != nil {
				if currentAnonymousConfig.Authentication == nil {
					currentAnonymousConfig.Authentication = &L2TPAuth{}
				}
				currentAnonymousConfig.Authentication.Method = matches[1]
			}
			continue
		}

		// PP auth myname
		if matches := ppAuthMynamePattern.FindStringSubmatch(line); len(matches) >= 3 && inAnonymousPP {
			if currentAnonymousConfig != nil {
				if currentAnonymousConfig.Authentication == nil {
					currentAnonymousConfig.Authentication = &L2TPAuth{}
				}
				currentAnonymousConfig.Authentication.Username = matches[1]
				currentAnonymousConfig.Authentication.Password = matches[2]
			}
			continue
		}

		// IP PP remote address pool
		if matches := ipPPRemotePoolPattern.FindStringSubmatch(line); len(matches) >= 3 && inAnonymousPP {
			if currentAnonymousConfig != nil {
				currentAnonymousConfig.IPPool = &L2TPIPPool{
					Start: matches[1],
					End:   matches[2],
				}
			}
			continue
		}

		// Tunnel select (L2TPv3)
		if matches := tunnelSelectPattern.FindStringSubmatch(line); len(matches) >= 2 {
			id, _ := strconv.Atoi(matches[1])
			currentTunnelID = id
			inAnonymousPP = false
			if _, exists := tunnels[id]; !exists {
				tunnels[id] = &L2TPConfig{
					ID:      id,
					Version: "l2tpv3",
					Mode:    "l2vpn",
					Enabled: true,
				}
			}
			continue
		}

		// Tunnel encapsulation
		if matches := tunnelEncapsulationPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				if matches[1] == "l2tpv3" {
					tunnel.Version = "l2tpv3"
					tunnel.Mode = "l2vpn"
				} else {
					tunnel.Version = "l2tp"
				}
			}
			continue
		}

		// Tunnel endpoint
		if matches := tunnelEndpointPattern.FindStringSubmatch(line); len(matches) >= 3 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				tunnel.TunnelSource = matches[1]
				tunnel.TunnelDest = matches[2]
			}
			continue
		}

		// L2TP local router-id
		if matches := l2tpLocalRouterIDPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				if tunnel.L2TPv3Config == nil {
					tunnel.L2TPv3Config = &L2TPv3Config{}
				}
				tunnel.L2TPv3Config.LocalRouterID = matches[1]
			}
			continue
		}

		// L2TP remote router-id
		if matches := l2tpRemoteRouterIDPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				if tunnel.L2TPv3Config == nil {
					tunnel.L2TPv3Config = &L2TPv3Config{}
				}
				tunnel.L2TPv3Config.RemoteRouterID = matches[1]
			}
			continue
		}

		// L2TP remote end-id
		if matches := l2tpRemoteEndIDPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				if tunnel.L2TPv3Config == nil {
					tunnel.L2TPv3Config = &L2TPv3Config{}
				}
				tunnel.L2TPv3Config.RemoteEndID = matches[1]
			}
			continue
		}

		// L2TP always-on
		if matches := l2tpAlwaysOnPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				tunnel.AlwaysOn = matches[1] == "on"
			}
			continue
		}

		// L2TP keepalive
		if matches := l2tpKeepalivePattern.FindStringSubmatch(line); len(matches) >= 3 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				tunnel.KeepaliveEnabled = true
				interval, _ := strconv.Atoi(matches[1])
				retry, _ := strconv.Atoi(matches[2])
				tunnel.KeepaliveConfig = &L2TPKeepalive{
					Interval: interval,
					Retry:    retry,
				}
			}
			continue
		}

		// L2TP tunnel disconnect time
		if matches := l2tpDisconnectTimePattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				tunnel.DisconnectTime, _ = strconv.Atoi(matches[1])
			}
			continue
		}

		// Description
		if matches := tunnelDescriptionPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				tunnel.Name = strings.TrimSpace(matches[1])
			}
			continue
		}

		// Tunnel endpoint name (FQDN destination)
		if matches := tunnelEndpointNamePattern.FindStringSubmatch(line); len(matches) >= 3 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				tunnel.TunnelDest = matches[1]
				tunnel.TunnelDestType = matches[2]
			}
			continue
		}

		// IPsec tunnel association
		if matches := ipsecTunnelPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				ipsecID, _ := strconv.Atoi(matches[1])
				if tunnel.IPsecProfile == nil {
					tunnel.IPsecProfile = &L2TPIPsec{Enabled: true}
				}
				tunnel.IPsecProfile.TunnelID = ipsecID
			}
			continue
		}

		// L2TP hostname (used as tunnel name)
		if matches := l2tpHostnamePattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				tunnel.Name = matches[1]
			}
			continue
		}

		// L2TP tunnel auth
		if matches := l2tpTunnelAuthPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				if tunnel.L2TPv3Config == nil {
					tunnel.L2TPv3Config = &L2TPv3Config{}
				}
				if tunnel.L2TPv3Config.TunnelAuth == nil {
					tunnel.L2TPv3Config.TunnelAuth = &L2TPTunnelAuth{}
				}
				tunnel.L2TPv3Config.TunnelAuth.Enabled = matches[1] == "on"
				if len(matches) >= 3 && matches[2] != "" {
					tunnel.L2TPv3Config.TunnelAuth.Password = matches[2]
				}
			}
			continue
		}

		// Tunnel enable
		if matches := tunnelEnablePattern.FindStringSubmatch(line); len(matches) >= 2 {
			enabledID, _ := strconv.Atoi(matches[1])
			if tunnel, exists := tunnels[enabledID]; exists {
				tunnel.Enabled = true
			}
			continue
		}
	}

	// Add anonymous PP config if exists, merging with tunnel config
	if currentAnonymousConfig != nil && currentAnonymousConfig.ID > 0 {
		if existing, exists := tunnels[currentAnonymousConfig.ID]; exists {
			// Preserve enabled state and IPsec profile from tunnel config
			currentAnonymousConfig.Enabled = existing.Enabled
			if existing.IPsecProfile != nil {
				currentAnonymousConfig.IPsecProfile = existing.IPsecProfile
			}
		}
		tunnels[currentAnonymousConfig.ID] = currentAnonymousConfig
	}

	// Convert map to slice
	result := make([]L2TPConfig, 0, len(tunnels))
	for _, tunnel := range tunnels {
		result = append(result, *tunnel)
	}

	return result, nil
}

// BuildL2TPServiceCommand builds the command to enable/disable L2TP service
// Command format: l2tp service on/off
func BuildL2TPServiceCommand(enabled bool) string {
	if enabled {
		return "l2tp service on"
	}
	return "l2tp service off"
}

// BuildPPSelectAnonymousCommand builds the command to select anonymous PP
// Command format: pp select anonymous
func BuildPPSelectAnonymousCommand() string {
	return "pp select anonymous"
}

// BuildPPBindTunnelCommand builds the command to bind PP to tunnel
// Command format: pp bind tunnel<n>
func BuildPPBindTunnelCommand(tunnelID int) string {
	return fmt.Sprintf("pp bind tunnel%d", tunnelID)
}

// BuildPPAuthAcceptCommand builds the command to set authentication method
// Command format: pp auth accept <method>
func BuildPPAuthAcceptCommand(method string) string {
	return fmt.Sprintf("pp auth accept %s", method)
}

// BuildPPAuthMynameCommand builds the command to set credentials
// Command format: pp auth myname <username> <password>
func BuildPPAuthMynameCommand(username, password string) string {
	return fmt.Sprintf("pp auth myname %s %s", username, password)
}

// BuildIPPPRemotePoolCommand builds the command to set IP pool
// Command format: ip pp remote address pool <start>-<end>
func BuildIPPPRemotePoolCommand(start, end string) string {
	return fmt.Sprintf("ip pp remote address pool %s-%s", start, end)
}

// BuildTunnelEncapsulationCommand builds the command to set tunnel encapsulation
// Command format: tunnel encapsulation l2tp|l2tpv3
func BuildTunnelEncapsulationCommand(tunnelID int, version string) string {
	return fmt.Sprintf("tunnel encapsulation %s", version)
}

// BuildTunnelEndpointCommand builds the command to set tunnel endpoints
// Command format: tunnel endpoint address <local> <remote>
func BuildTunnelEndpointCommand(local, remote string) string {
	return fmt.Sprintf("tunnel endpoint address %s %s", local, remote)
}

// BuildL2TPLocalRouterIDCommand builds the command to set local router ID
// Command format: l2tp local router-id <ip>
func BuildL2TPLocalRouterIDCommand(routerID string) string {
	return fmt.Sprintf("l2tp local router-id %s", routerID)
}

// BuildL2TPRemoteRouterIDCommand builds the command to set remote router ID
// Command format: l2tp remote router-id <ip>
func BuildL2TPRemoteRouterIDCommand(routerID string) string {
	return fmt.Sprintf("l2tp remote router-id %s", routerID)
}

// BuildL2TPRemoteEndIDCommand builds the command to set remote end ID
// Command format: l2tp remote end-id <string>
func BuildL2TPRemoteEndIDCommand(endID string) string {
	return fmt.Sprintf("l2tp remote end-id %s", endID)
}

// BuildL2TPAlwaysOnCommand builds the command to set always-on mode
// Command format: l2tp always-on on/off
func BuildL2TPAlwaysOnCommand(enabled bool) string {
	if enabled {
		return "l2tp always-on on"
	}
	return "l2tp always-on off"
}

// BuildL2TPKeepaliveCommand builds the command to configure keepalive
// Command format: l2tp keepalive use on <interval> <retry>
func BuildL2TPKeepaliveCommand(interval, retry int) string {
	return fmt.Sprintf("l2tp keepalive use on %d %d", interval, retry)
}

// BuildL2TPKeepaliveOffCommand builds the command to disable keepalive
// Command format: l2tp keepalive use off
func BuildL2TPKeepaliveOffCommand() string {
	return "l2tp keepalive use off"
}

// BuildL2TPDisconnectTimeCommand builds the command to set disconnect time
// Command format: l2tp tunnel disconnect time <seconds>
func BuildL2TPDisconnectTimeCommand(seconds int) string {
	return fmt.Sprintf("l2tp tunnel disconnect time %d", seconds)
}

// BuildDeleteL2TPTunnelCommand builds the command to delete L2TP tunnel
// Command format: no tunnel select <n>
func BuildDeleteL2TPTunnelCommand(tunnelID int) string {
	return fmt.Sprintf("no tunnel select %d", tunnelID)
}

// BuildShowL2TPConfigCommand builds the command to show L2TP configuration
// Uses full "show config" output since RTX routers don't support pipe commands
func BuildShowL2TPConfigCommand() string {
	return "show config"
}

// ValidateL2TPConfig validates an L2TP configuration
func ValidateL2TPConfig(config L2TPConfig) error {
	if config.ID < 0 {
		return fmt.Errorf("tunnel id must be non-negative")
	}

	if config.Version != "l2tp" && config.Version != "l2tpv3" {
		return fmt.Errorf("invalid version: must be 'l2tp' or 'l2tpv3'")
	}

	if config.Mode != "lns" && config.Mode != "l2vpn" {
		return fmt.Errorf("invalid mode: must be 'lns' or 'l2vpn'")
	}

	// L2TPv2 LNS validation
	if config.Version == "l2tp" && config.Mode == "lns" {
		if config.Authentication != nil {
			validMethods := map[string]bool{
				"pap": true, "chap": true, "mschap": true, "mschap-v2": true,
			}
			if !validMethods[config.Authentication.Method] {
				return fmt.Errorf("invalid authentication method: %s", config.Authentication.Method)
			}
		}
		if config.IPPool != nil {
			if !isValidIP(config.IPPool.Start) {
				return fmt.Errorf("invalid ip_pool start: %s", config.IPPool.Start)
			}
			if !isValidIP(config.IPPool.End) {
				return fmt.Errorf("invalid ip_pool end: %s", config.IPPool.End)
			}
		}
	}

	// L2TPv3 validation
	if config.Version == "l2tpv3" {
		if config.TunnelSource != "" && !isValidIP(config.TunnelSource) {
			return fmt.Errorf("invalid tunnel_source: %s", config.TunnelSource)
		}
		if config.TunnelDest != "" && !isValidIP(config.TunnelDest) {
			// Could be FQDN
			if config.TunnelDestType != "fqdn" && !isValidIP(config.TunnelDest) {
				return fmt.Errorf("invalid tunnel_dest: %s", config.TunnelDest)
			}
		}
		if config.L2TPv3Config != nil {
			if config.L2TPv3Config.LocalRouterID != "" && !isValidIP(config.L2TPv3Config.LocalRouterID) {
				return fmt.Errorf("invalid local_router_id: %s", config.L2TPv3Config.LocalRouterID)
			}
			if config.L2TPv3Config.RemoteRouterID != "" && !isValidIP(config.L2TPv3Config.RemoteRouterID) {
				return fmt.Errorf("invalid remote_router_id: %s", config.L2TPv3Config.RemoteRouterID)
			}
		}
	}

	return nil
}

// L2TPService represents the L2TP service state
type L2TPService struct {
	Enabled   bool     `json:"enabled"`
	Protocols []string `json:"protocols,omitempty"` // ["l2tpv3", "l2tp"]
}

// ParseL2TPServiceConfig parses the output containing "l2tp service on/off [protocols...]"
// RTX command format: l2tp service on [l2tpv3] [l2tp]
// or: l2tp service off
func ParseL2TPServiceConfig(raw string) (*L2TPService, error) {
	lines := strings.Split(raw, "\n")

	// Pattern: l2tp service on [l2tpv3] [l2tp]
	// or: l2tp service off
	servicePattern := regexp.MustCompile(`^\s*l2tp\s+service\s+(on|off)(?:\s+(.+))?\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if matches := servicePattern.FindStringSubmatch(line); len(matches) >= 2 {
			service := &L2TPService{
				Enabled:   matches[1] == "on",
				Protocols: []string{},
			}

			// Parse protocols if service is on and protocols are specified
			if service.Enabled && len(matches) >= 3 && matches[2] != "" {
				protocolsPart := strings.TrimSpace(matches[2])
				if protocolsPart != "" {
					protocols := strings.Fields(protocolsPart)
					for _, proto := range protocols {
						// Valid protocols: l2tpv3, l2tp
						if proto == "l2tpv3" || proto == "l2tp" {
							service.Protocols = append(service.Protocols, proto)
						}
					}
				}
			}

			return service, nil
		}
	}

	// If no l2tp service line found, assume service is off (default state)
	return &L2TPService{
		Enabled:   false,
		Protocols: []string{},
	}, nil
}

// BuildL2TPServiceCommandWithProtocols builds the command to enable/disable L2TP service with protocols
// Command format: l2tp service on [l2tpv3] [l2tp]
// or: l2tp service off
func BuildL2TPServiceCommandWithProtocols(enabled bool, protocols []string) string {
	if !enabled {
		return "l2tp service off"
	}

	if len(protocols) == 0 {
		return "l2tp service on"
	}

	// Build command with protocols
	// Example: l2tp service on l2tpv3 l2tp
	return fmt.Sprintf("l2tp service on %s", strings.Join(protocols, " "))
}
