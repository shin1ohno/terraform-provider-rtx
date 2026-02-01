package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Tunnel represents a unified tunnel configuration (rtx_tunnel resource)
// This combines IPsec and L2TP settings under a single tunnel select N context
type Tunnel struct {
	ID            int          `json:"id"`                    // tunnel select N
	Encapsulation string       `json:"encapsulation"`         // "ipsec", "l2tpv3", or "l2tp"
	Enabled       bool         `json:"enabled"`               // tunnel enable N
	Name          string       `json:"name,omitempty"`        // Description
	IPsec         *TunnelIPsec `json:"ipsec,omitempty"`       // IPsec configuration
	L2TP          *TunnelL2TP  `json:"l2tp,omitempty"`        // L2TP configuration
}

// TunnelIPsec represents IPsec settings within a unified tunnel
type TunnelIPsec struct {
	IPsecTunnelID   int                   `json:"ipsec_tunnel_id"`             // ipsec tunnel N (Computed: defaults to tunnel_id)
	LocalAddress    string                `json:"local_address,omitempty"`     // ipsec ike local address
	RemoteAddress   string                `json:"remote_address,omitempty"`    // ipsec ike remote address
	PreSharedKey    string                `json:"pre_shared_key"`              // ipsec ike pre-shared-key
	IKEv2Proposal   IKEv2Proposal         `json:"ikev2_proposal"`              // IKE Phase 1 proposal
	Transform       IPsecTransform        `json:"transform"`                   // IPsec Phase 2 transform
	Keepalive       *TunnelIPsecKeepalive `json:"keepalive,omitempty"`         // DPD/heartbeat settings
	SecureFilterIn  []int                 `json:"secure_filter_in,omitempty"`  // ip tunnel secure filter in
	SecureFilterOut []int                 `json:"secure_filter_out,omitempty"` // ip tunnel secure filter out
	TCPMSSLimit     string                `json:"tcp_mss_limit,omitempty"`     // ip tunnel tcp mss limit
}

// TunnelIPsecKeepalive represents IPsec keepalive/DPD settings within a tunnel
type TunnelIPsecKeepalive struct {
	Enabled  bool   `json:"enabled"`  // Keepalive enabled
	Mode     string `json:"mode"`     // "dpd" or "heartbeat"
	Interval int    `json:"interval"` // Interval in seconds
	Retry    int    `json:"retry"`    // Retry count
}

// TunnelL2TP represents L2TP settings within a unified tunnel
type TunnelL2TP struct {
	// Common L2TP settings
	Hostname       string               `json:"hostname,omitempty"`        // l2tp hostname
	AlwaysOn       bool                 `json:"always_on,omitempty"`       // l2tp always-on
	DisconnectTime int                  `json:"disconnect_time,omitempty"` // Idle disconnect time (0 = disabled)
	Keepalive      *TunnelL2TPKeepalive `json:"keepalive,omitempty"`       // l2tp keepalive use
	SyslogEnabled  bool                 `json:"syslog_enabled,omitempty"`  // l2tp syslog on

	// L2TPv3 specific
	LocalRouterID  string           `json:"local_router_id,omitempty"`  // l2tp local router-id
	RemoteRouterID string           `json:"remote_router_id,omitempty"` // l2tp remote router-id
	RemoteEndID    string           `json:"remote_end_id,omitempty"`    // l2tp remote end-id
	TunnelAuth     *TunnelL2TPAuth  `json:"tunnel_auth,omitempty"`      // l2tp tunnel auth

	// L2TPv2 specific (remote access)
	Authentication *L2TPAuth   `json:"authentication,omitempty"` // PPP authentication
	IPPool         *L2TPIPPool `json:"ip_pool,omitempty"`        // Client IP pool
}

// TunnelL2TPKeepalive represents L2TP keepalive settings within a tunnel
type TunnelL2TPKeepalive struct {
	Enabled  bool `json:"enabled"`  // l2tp keepalive use on
	Interval int  `json:"interval"` // Interval in seconds
	Retry    int  `json:"retry"`    // Retry count
}

// TunnelL2TPAuth represents L2TPv3 tunnel authentication
type TunnelL2TPAuth struct {
	Enabled  bool   `json:"enabled"`            // l2tp tunnel auth on
	Password string `json:"password,omitempty"` // Tunnel auth password
}

// TunnelParser parses unified tunnel configuration output
type TunnelParser struct{}

// NewTunnelParser creates a new unified tunnel parser
func NewTunnelParser() *TunnelParser {
	return &TunnelParser{}
}

// ParseTunnelConfig parses the output of "show config" for unified tunnels
func (p *TunnelParser) ParseTunnelConfig(raw string) ([]Tunnel, error) {
	tunnels := make(map[int]*Tunnel)
	lines := strings.Split(raw, "\n")

	// Tunnel patterns
	tunnelSelectPattern := regexp.MustCompile(`^\s*tunnel\s+select\s+(\d+)\s*$`)
	tunnelEncapsulationPattern := regexp.MustCompile(`^\s*tunnel\s+encapsulation\s+(l2tp|l2tpv3)\s*$`)
	tunnelEnablePattern := regexp.MustCompile(`^\s*tunnel\s+enable\s+(\d+)\s*$`)
	tunnelDescriptionPattern := regexp.MustCompile(`^\s*description\s+(.+)\s*$`)

	// IPsec patterns
	ipsecTunnelPattern := regexp.MustCompile(`^\s*ipsec\s+tunnel\s+(\d+)\s*$`)
	ipsecSAPolicyPattern := regexp.MustCompile(`^\s*ipsec\s+sa\s+policy\s+(\d+)\s+(\d+)\s+(\w+)\s+(.+)\s*$`)
	ipsecIKELocalAddrPattern := regexp.MustCompile(`^\s*ipsec\s+ike\s+local\s+address\s+(\d+)\s+(\S+)\s*$`)
	ipsecIKERemoteAddrPattern := regexp.MustCompile(`^\s*ipsec\s+ike\s+remote\s+address\s+(\d+)\s+(\S+)\s*$`)
	ipsecIKEPreSharedKeyPattern := regexp.MustCompile(`^\s*ipsec\s+ike\s+pre-shared-key\s+(\d+)\s+text\s+(.+)\s*$`)
	ipsecIKEEncryptionPattern := regexp.MustCompile(`^\s*ipsec\s+ike\s+encryption\s+(\d+)\s+(.+)\s*$`)
	ipsecIKEHashPattern := regexp.MustCompile(`^\s*ipsec\s+ike\s+hash\s+(\d+)\s+(.+)\s*$`)
	ipsecIKEGroupPattern := regexp.MustCompile(`^\s*ipsec\s+ike\s+group\s+(\d+)\s+(.+)\s*$`)
	ipsecIKEKeepaliveRetryPattern := regexp.MustCompile(`^\s*ipsec\s+ike\s+keepalive\s+use\s+(\d+)\s+on\s+dpd\s+(\d+)\s+(\d+)\s*$`)
	ipsecIKEKeepalivePattern := regexp.MustCompile(`^\s*ipsec\s+ike\s+keepalive\s+use\s+(\d+)\s+on\s+dpd\s+(\d+)\s*$`)
	ipsecIKEKeepaliveHeartbeatPattern := regexp.MustCompile(`^\s*ipsec\s+ike\s+keepalive\s+use\s+(\d+)\s+on\s+heartbeat\s+(\d+)\s+(\d+)\s*$`)
	ipTunnelSecureFilterPattern := regexp.MustCompile(`^\s*ip\s+tunnel\s+secure\s+filter\s+(in|out)\s+(.+)$`)
	ipTunnelTCPMSSPattern := regexp.MustCompile(`^\s*ip\s+tunnel\s+tcp\s+mss\s+limit\s+(\S+)\s*$`)

	// L2TP patterns
	l2tpHostnamePattern := regexp.MustCompile(`^\s*l2tp\s+hostname\s+(\S+)\s*$`)
	l2tpLocalRouterIDPattern := regexp.MustCompile(`^\s*l2tp\s+local\s+router-id\s+([0-9.]+)\s*$`)
	l2tpRemoteRouterIDPattern := regexp.MustCompile(`^\s*l2tp\s+remote\s+router-id\s+([0-9.]+)\s*$`)
	l2tpRemoteEndIDPattern := regexp.MustCompile(`^\s*l2tp\s+remote\s+end-id\s+(\S+)\s*$`)
	l2tpAlwaysOnPattern := regexp.MustCompile(`^\s*l2tp\s+always-on\s+(on|off)\s*$`)
	l2tpTunnelAuthPattern := regexp.MustCompile(`^\s*l2tp\s+tunnel\s+auth\s+(on|off)(?:\s+(\S+))?\s*$`)
	l2tpKeepalivePattern := regexp.MustCompile(`^\s*l2tp\s+keepalive\s+use\s+on\s+(\d+)\s+(\d+)\s*$`)
	l2tpSyslogPattern := regexp.MustCompile(`^\s*l2tp\s+syslog\s+(on|off)\s*$`)

	// L2TPv2 LNS patterns (anonymous PP context)
	ppSelectAnonymousPattern := regexp.MustCompile(`^\s*pp\s+select\s+anonymous\s*$`)
	ppBindTunnelPattern := regexp.MustCompile(`^\s*pp\s+bind\s+tunnel(\d+)\s*$`)
	ppAuthAcceptPattern := regexp.MustCompile(`^\s*pp\s+auth\s+accept\s+(\S+)\s*$`)
	ppAuthRequestPattern := regexp.MustCompile(`^\s*pp\s+auth\s+request\s+(\S+)\s*$`)
	ppAuthMynamePattern := regexp.MustCompile(`^\s*pp\s+auth\s+myname\s+(\S+)\s+(\S+)\s*$`)
	ipPPRemotePoolPattern := regexp.MustCompile(`^\s*ip\s+pp\s+remote\s+address\s+pool\s+([0-9.]+)-([0-9.]+)\s*$`)

	var currentTunnelID int
	var inAnonymousPP bool
	var anonymousPPTunnelID int
	var anonymousAuth *L2TPAuth
	var anonymousIPPool *L2TPIPPool

	for _, rawLine := range lines {
		// Check if line is indented (within a context) before trimming
		// RTX config uses single space indentation for context-specific commands
		isIndented := len(rawLine) > 0 && (rawLine[0] == ' ' || rawLine[0] == '\t')

		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		// Reset tunnel context when we encounter a non-indented line that's not a tunnel/pp select
		// This prevents global descriptions from being assigned to tunnels
		if !isIndented && currentTunnelID > 0 {
			if !strings.HasPrefix(line, "tunnel select") && !strings.HasPrefix(line, "tunnel enable") {
				currentTunnelID = 0
			}
		}

		// Tunnel select - creates a new tunnel entry
		if matches := tunnelSelectPattern.FindStringSubmatch(line); len(matches) >= 2 {
			id, _ := strconv.Atoi(matches[1])
			currentTunnelID = id
			inAnonymousPP = false
			if _, exists := tunnels[id]; !exists {
				tunnels[id] = &Tunnel{
					ID:            id,
					Encapsulation: "ipsec", // Default to IPsec, will be overridden if l2tp encapsulation found
					Enabled:       false,   // Will be set true when tunnel enable found
				}
			}
			continue
		}

		// Tunnel encapsulation - determines encapsulation type
		if matches := tunnelEncapsulationPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				tunnel.Encapsulation = matches[1]
			}
			continue
		}

		// IPsec tunnel - records ipsec tunnel ID and creates IPsec block
		if matches := ipsecTunnelPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			ipsecTunnelID, _ := strconv.Atoi(matches[1])
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				if tunnel.IPsec == nil {
					tunnel.IPsec = &TunnelIPsec{
						IKEv2Proposal: IKEv2Proposal{LifetimeSeconds: 28800},
						Transform: IPsecTransform{
							Protocol:        "esp",
							LifetimeSeconds: 3600,
						},
					}
				}
				tunnel.IPsec.IPsecTunnelID = ipsecTunnelID
			}
			continue
		}

		// IPsec SA policy
		if matches := ipsecSAPolicyPattern.FindStringSubmatch(line); len(matches) >= 5 && currentTunnelID > 0 {
			policyNum, _ := strconv.Atoi(matches[1])
			protocol := matches[3]
			algorithms := matches[4]

			if tunnel, exists := tunnels[currentTunnelID]; exists && tunnel.IPsec != nil {
				if tunnel.IPsec.IPsecTunnelID == policyNum {
					tunnel.IPsec.Transform.Protocol = protocol
					parseIPsecSAAlgorithms(algorithms, &tunnel.IPsec.Transform)
				}
			}
			continue
		}

		// IPsec IKE local address
		// Note: IKE gateway ID may differ from IPsec tunnel ID, so we assign to current tunnel context
		if matches := ipsecIKELocalAddrPattern.FindStringSubmatch(line); len(matches) >= 3 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists && tunnel.IPsec != nil {
				tunnel.IPsec.LocalAddress = matches[2]
			}
			continue
		}

		// IPsec IKE remote address
		// Note: IKE gateway ID may differ from IPsec tunnel ID, so we assign to current tunnel context
		if matches := ipsecIKERemoteAddrPattern.FindStringSubmatch(line); len(matches) >= 3 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists && tunnel.IPsec != nil {
				tunnel.IPsec.RemoteAddress = matches[2]
			}
			continue
		}

		// IPsec IKE pre-shared-key
		// Note: IKE gateway ID may differ from IPsec tunnel ID, so we assign to current tunnel context
		if matches := ipsecIKEPreSharedKeyPattern.FindStringSubmatch(line); len(matches) >= 3 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists && tunnel.IPsec != nil {
				tunnel.IPsec.PreSharedKey = strings.TrimSpace(matches[2])
			}
			continue
		}

		// IPsec IKE encryption
		// Note: IKE gateway ID may differ from IPsec tunnel ID, so we assign to current tunnel context
		if matches := ipsecIKEEncryptionPattern.FindStringSubmatch(line); len(matches) >= 3 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists && tunnel.IPsec != nil {
				parseIKEEncryption(matches[2], &tunnel.IPsec.IKEv2Proposal)
			}
			continue
		}

		// IPsec IKE hash
		// Note: IKE gateway ID may differ from IPsec tunnel ID, so we assign to current tunnel context
		if matches := ipsecIKEHashPattern.FindStringSubmatch(line); len(matches) >= 3 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists && tunnel.IPsec != nil {
				parseIKEHash(matches[2], &tunnel.IPsec.IKEv2Proposal)
			}
			continue
		}

		// IPsec IKE group
		// Note: IKE gateway ID may differ from IPsec tunnel ID, so we assign to current tunnel context
		if matches := ipsecIKEGroupPattern.FindStringSubmatch(line); len(matches) >= 3 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists && tunnel.IPsec != nil {
				parseIKEGroup(matches[2], &tunnel.IPsec.IKEv2Proposal)
			}
			continue
		}

		// IPsec IKE keepalive (DPD) with retry
		// Note: IKE gateway ID may differ from IPsec tunnel ID, so we assign to current tunnel context
		if matches := ipsecIKEKeepaliveRetryPattern.FindStringSubmatch(line); len(matches) >= 4 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists && tunnel.IPsec != nil {
				interval, _ := strconv.Atoi(matches[2])
				retry, _ := strconv.Atoi(matches[3])
				tunnel.IPsec.Keepalive = &TunnelIPsecKeepalive{
					Enabled:  true,
					Mode:     "dpd",
					Interval: interval,
					Retry:    retry,
				}
			}
			continue
		}

		// IPsec IKE keepalive (DPD) without retry
		// Note: IKE gateway ID may differ from IPsec tunnel ID, so we assign to current tunnel context
		if matches := ipsecIKEKeepalivePattern.FindStringSubmatch(line); len(matches) >= 3 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists && tunnel.IPsec != nil {
				interval, _ := strconv.Atoi(matches[2])
				tunnel.IPsec.Keepalive = &TunnelIPsecKeepalive{
					Enabled:  true,
					Mode:     "dpd",
					Interval: interval,
				}
			}
			continue
		}

		// IPsec IKE keepalive (heartbeat)
		// Note: IKE gateway ID may differ from IPsec tunnel ID, so we assign to current tunnel context
		if matches := ipsecIKEKeepaliveHeartbeatPattern.FindStringSubmatch(line); len(matches) >= 4 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists && tunnel.IPsec != nil {
				interval, _ := strconv.Atoi(matches[2])
				retry, _ := strconv.Atoi(matches[3])
				tunnel.IPsec.Keepalive = &TunnelIPsecKeepalive{
					Enabled:  true,
					Mode:     "heartbeat",
					Interval: interval,
					Retry:    retry,
				}
			}
			continue
		}

		// IP tunnel secure filter
		if matches := ipTunnelSecureFilterPattern.FindStringSubmatch(line); len(matches) >= 3 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists && tunnel.IPsec != nil {
				direction := matches[1]
				filterIDsStr := strings.TrimSpace(matches[2])
				filterIDs := parseFilterIDs(filterIDsStr)
				if direction == "in" {
					tunnel.IPsec.SecureFilterIn = filterIDs
				} else {
					tunnel.IPsec.SecureFilterOut = filterIDs
				}
			}
			continue
		}

		// IP tunnel TCP MSS limit
		if matches := ipTunnelTCPMSSPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists && tunnel.IPsec != nil {
				tunnel.IPsec.TCPMSSLimit = strings.TrimSpace(matches[1])
			}
			continue
		}

		// Description (within tunnel context)
		if matches := tunnelDescriptionPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				tunnel.Name = strings.TrimSpace(matches[1])
			}
			continue
		}

		// L2TP hostname
		if matches := l2tpHostnamePattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				if tunnel.L2TP == nil {
					tunnel.L2TP = &TunnelL2TP{}
				}
				tunnel.L2TP.Hostname = matches[1]
			}
			continue
		}

		// L2TP local router-id
		if matches := l2tpLocalRouterIDPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				if tunnel.L2TP == nil {
					tunnel.L2TP = &TunnelL2TP{}
				}
				tunnel.L2TP.LocalRouterID = matches[1]
			}
			continue
		}

		// L2TP remote router-id
		if matches := l2tpRemoteRouterIDPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				if tunnel.L2TP == nil {
					tunnel.L2TP = &TunnelL2TP{}
				}
				tunnel.L2TP.RemoteRouterID = matches[1]
			}
			continue
		}

		// L2TP remote end-id
		if matches := l2tpRemoteEndIDPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				if tunnel.L2TP == nil {
					tunnel.L2TP = &TunnelL2TP{}
				}
				tunnel.L2TP.RemoteEndID = matches[1]
			}
			continue
		}

		// L2TP always-on
		if matches := l2tpAlwaysOnPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				if tunnel.L2TP == nil {
					tunnel.L2TP = &TunnelL2TP{}
				}
				tunnel.L2TP.AlwaysOn = matches[1] == "on"
			}
			continue
		}

		// L2TP tunnel auth
		if matches := l2tpTunnelAuthPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				if tunnel.L2TP == nil {
					tunnel.L2TP = &TunnelL2TP{}
				}
				tunnel.L2TP.TunnelAuth = &TunnelL2TPAuth{
					Enabled: matches[1] == "on",
				}
				if len(matches) >= 3 && matches[2] != "" {
					tunnel.L2TP.TunnelAuth.Password = matches[2]
				}
			}
			continue
		}

		// L2TP keepalive
		if matches := l2tpKeepalivePattern.FindStringSubmatch(line); len(matches) >= 3 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				if tunnel.L2TP == nil {
					tunnel.L2TP = &TunnelL2TP{}
				}
				interval, _ := strconv.Atoi(matches[1])
				retry, _ := strconv.Atoi(matches[2])
				tunnel.L2TP.Keepalive = &TunnelL2TPKeepalive{
					Enabled:  true,
					Interval: interval,
					Retry:    retry,
				}
			}
			continue
		}

		// L2TP syslog
		if matches := l2tpSyslogPattern.FindStringSubmatch(line); len(matches) >= 2 && currentTunnelID > 0 {
			if tunnel, exists := tunnels[currentTunnelID]; exists {
				if tunnel.L2TP == nil {
					tunnel.L2TP = &TunnelL2TP{}
				}
				tunnel.L2TP.SyslogEnabled = matches[1] == "on"
			}
			continue
		}

		// PP select anonymous (L2TPv2 LNS)
		if ppSelectAnonymousPattern.MatchString(line) {
			inAnonymousPP = true
			currentTunnelID = 0 // Reset tunnel context
			continue
		}

		// PP bind tunnel (L2TPv2)
		if matches := ppBindTunnelPattern.FindStringSubmatch(line); len(matches) >= 2 && inAnonymousPP {
			tunnelID, _ := strconv.Atoi(matches[1])
			anonymousPPTunnelID = tunnelID
			continue
		}

		// PP auth accept (L2TPv2)
		if matches := ppAuthAcceptPattern.FindStringSubmatch(line); len(matches) >= 2 && inAnonymousPP {
			if anonymousAuth == nil {
				anonymousAuth = &L2TPAuth{}
			}
			anonymousAuth.Method = matches[1]
			continue
		}

		// PP auth request (L2TPv2)
		if matches := ppAuthRequestPattern.FindStringSubmatch(line); len(matches) >= 2 && inAnonymousPP {
			if anonymousAuth == nil {
				anonymousAuth = &L2TPAuth{}
			}
			anonymousAuth.RequestMethod = matches[1]
			continue
		}

		// PP auth myname (L2TPv2)
		if matches := ppAuthMynamePattern.FindStringSubmatch(line); len(matches) >= 3 && inAnonymousPP {
			if anonymousAuth == nil {
				anonymousAuth = &L2TPAuth{}
			}
			anonymousAuth.Username = matches[1]
			anonymousAuth.Password = matches[2]
			continue
		}

		// IP PP remote address pool (L2TPv2)
		if matches := ipPPRemotePoolPattern.FindStringSubmatch(line); len(matches) >= 3 && inAnonymousPP {
			anonymousIPPool = &L2TPIPPool{
				Start: matches[1],
				End:   matches[2],
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

	// Apply anonymous PP settings to the associated tunnel
	if anonymousPPTunnelID > 0 {
		if tunnel, exists := tunnels[anonymousPPTunnelID]; exists {
			if tunnel.L2TP == nil {
				tunnel.L2TP = &TunnelL2TP{}
			}
			tunnel.L2TP.Authentication = anonymousAuth
			tunnel.L2TP.IPPool = anonymousIPPool
		}
	}

	// Convert map to slice
	result := make([]Tunnel, 0, len(tunnels))
	for _, tunnel := range tunnels {
		result = append(result, *tunnel)
	}

	return result, nil
}

// BuildTunnelCommands builds all commands for a unified tunnel configuration
func BuildTunnelCommands(tunnel Tunnel) []string {
	var commands []string

	// tunnel select N
	commands = append(commands, BuildTunnelSelectCommand(tunnel.ID))

	// tunnel encapsulation (for L2TP)
	if tunnel.Encapsulation == "l2tpv3" || tunnel.Encapsulation == "l2tp" {
		commands = append(commands, BuildTunnelEncapsulationCommand(tunnel.ID, tunnel.Encapsulation))
	}

	// Note: description command is not generated for tunnels.
	// The tunnel.Name is read from the config but not written back
	// because RTX doesn't support "description" command within tunnel select context.
	// To set a tunnel description, use a separate rtx_interface resource for the tunnel interface.

	// IPsec commands
	if tunnel.IPsec != nil {
		ipsecCmds := buildTunnelIPsecCommands(tunnel.ID, tunnel.IPsec)
		commands = append(commands, ipsecCmds...)
	}

	// L2TP commands
	if tunnel.L2TP != nil {
		l2tpCmds := buildTunnelL2TPCommands(tunnel.Encapsulation, tunnel.L2TP)
		commands = append(commands, l2tpCmds...)
	}

	// tunnel enable/disable
	if tunnel.Enabled {
		commands = append(commands, BuildTunnelEnableCommand(tunnel.ID))
	}

	return commands
}

// buildTunnelDescription builds the description command within tunnel context
// Command format: description <text>
func buildTunnelDescription(description string) string {
	return fmt.Sprintf("description %s", description)
}

// buildTunnelIPsecCommands builds IPsec-related commands within tunnel context
func buildTunnelIPsecCommands(tunnelID int, ipsec *TunnelIPsec) []string {
	var commands []string

	// Use ipsec_tunnel_id, default to tunnel_id if not set
	ipsecTunnelID := ipsec.IPsecTunnelID
	if ipsecTunnelID == 0 {
		ipsecTunnelID = tunnelID
	}

	// ipsec tunnel N
	commands = append(commands, BuildIPsecTunnelCommand(ipsecTunnelID))

	// ipsec sa policy
	commands = append(commands, BuildIPsecSAPolicyCommand(ipsecTunnelID, 1, ipsec.Transform))

	// IKE commands use tunnelID (from tunnel select N), not ipsecTunnelID (from ipsec tunnel N)
	// This is because RTX uses separate ID spaces for IPsec tunnels and IKE gateways

	// ipsec ike local address
	if ipsec.LocalAddress != "" {
		commands = append(commands, BuildIPsecIKELocalAddressCommand(tunnelID, ipsec.LocalAddress))
	}

	// ipsec ike remote address
	if ipsec.RemoteAddress != "" {
		commands = append(commands, BuildIPsecIKERemoteAddressCommand(tunnelID, ipsec.RemoteAddress))
	}

	// ipsec ike pre-shared-key (skip if empty - WriteOnly attribute may not be available on Update)
	if ipsec.PreSharedKey != "" {
		commands = append(commands, BuildIPsecIKEPreSharedKeyCommand(tunnelID, ipsec.PreSharedKey))
	}

	// ipsec ike encryption
	commands = append(commands, BuildIPsecIKEEncryptionCommand(tunnelID, ipsec.IKEv2Proposal))

	// ipsec ike hash
	commands = append(commands, BuildIPsecIKEHashCommand(tunnelID, ipsec.IKEv2Proposal))

	// ipsec ike group
	commands = append(commands, BuildIPsecIKEGroupCommand(tunnelID, ipsec.IKEv2Proposal))

	// ipsec ike keepalive
	if ipsec.Keepalive != nil && ipsec.Keepalive.Enabled {
		if ipsec.Keepalive.Mode == "heartbeat" {
			commands = append(commands, BuildIPsecIKEKeepaliveHeartbeatCommand(tunnelID, ipsec.Keepalive.Interval, ipsec.Keepalive.Retry))
		} else {
			commands = append(commands, BuildIPsecIKEKeepaliveCommand(tunnelID, ipsec.Keepalive.Interval, ipsec.Keepalive.Retry))
		}
	}

	// ip tunnel secure filter in
	if len(ipsec.SecureFilterIn) > 0 {
		commands = append(commands, BuildIPTunnelSecureFilterCommand("in", ipsec.SecureFilterIn))
	}

	// ip tunnel secure filter out
	if len(ipsec.SecureFilterOut) > 0 {
		commands = append(commands, BuildIPTunnelSecureFilterCommand("out", ipsec.SecureFilterOut))
	}

	// ip tunnel tcp mss limit
	if ipsec.TCPMSSLimit != "" {
		commands = append(commands, BuildIPTunnelTCPMSSLimitCommand(ipsec.TCPMSSLimit))
	}

	return commands
}

// buildTunnelL2TPCommands builds L2TP-related commands within tunnel context
func buildTunnelL2TPCommands(encapsulation string, l2tp *TunnelL2TP) []string {
	var commands []string

	// l2tp hostname
	if l2tp.Hostname != "" {
		commands = append(commands, BuildL2TPHostnameCommand(l2tp.Hostname))
	}

	// L2TPv3 specific commands
	if encapsulation == "l2tpv3" {
		// l2tp local router-id
		if l2tp.LocalRouterID != "" {
			commands = append(commands, BuildL2TPLocalRouterIDCommand(l2tp.LocalRouterID))
		}

		// l2tp remote router-id
		if l2tp.RemoteRouterID != "" {
			commands = append(commands, BuildL2TPRemoteRouterIDCommand(l2tp.RemoteRouterID))
		}

		// l2tp remote end-id
		if l2tp.RemoteEndID != "" {
			commands = append(commands, BuildL2TPRemoteEndIDCommand(l2tp.RemoteEndID))
		}

		// l2tp tunnel auth
		if l2tp.TunnelAuth != nil && l2tp.TunnelAuth.Enabled {
			commands = append(commands, BuildL2TPTunnelAuthCommand(l2tp.TunnelAuth.Password))
		}
	}

	// l2tp always-on
	if l2tp.AlwaysOn {
		commands = append(commands, BuildL2TPAlwaysOnCommand(true))
	}

	// l2tp keepalive
	if l2tp.Keepalive != nil && l2tp.Keepalive.Enabled {
		commands = append(commands, BuildL2TPKeepaliveCommand(l2tp.Keepalive.Interval, l2tp.Keepalive.Retry))
	}

	// l2tp syslog
	if l2tp.SyslogEnabled {
		commands = append(commands, BuildL2TPSyslogCommand(true))
	}

	return commands
}

// BuildL2TPHostnameCommand builds the L2TP hostname command
// Command format: l2tp hostname <name>
func BuildL2TPHostnameCommand(hostname string) string {
	return fmt.Sprintf("l2tp hostname %s", hostname)
}

// BuildL2TPTunnelAuthCommand builds the L2TP tunnel auth command
// Command format: l2tp tunnel auth on [password]
func BuildL2TPTunnelAuthCommand(password string) string {
	if password != "" {
		return fmt.Sprintf("l2tp tunnel auth on %s", password)
	}
	return "l2tp tunnel auth on"
}

// BuildL2TPSyslogCommand builds the L2TP syslog command
// Command format: l2tp syslog on/off
func BuildL2TPSyslogCommand(enabled bool) string {
	if enabled {
		return "l2tp syslog on"
	}
	return "l2tp syslog off"
}

// BuildDeleteTunnelCommands builds commands to delete a tunnel
func BuildDeleteTunnelCommands(tunnelID int) []string {
	return []string{
		BuildTunnelSelectCommand(tunnelID),
		fmt.Sprintf("no ipsec tunnel %d", tunnelID),
		BuildDeleteTunnelSelectCommand(tunnelID),
	}
}

// ValidateTunnel validates a unified tunnel configuration
func ValidateTunnel(tunnel Tunnel) error {
	if tunnel.ID <= 0 {
		return fmt.Errorf("tunnel_id must be positive")
	}

	if tunnel.Encapsulation == "" {
		return fmt.Errorf("encapsulation is required")
	}

	validEncapsulations := map[string]bool{"ipsec": true, "l2tpv3": true, "l2tp": true}
	if !validEncapsulations[tunnel.Encapsulation] {
		return fmt.Errorf("encapsulation must be 'ipsec', 'l2tpv3', or 'l2tp'")
	}

	// Validate based on encapsulation type
	switch tunnel.Encapsulation {
	case "ipsec":
		if tunnel.IPsec == nil {
			return fmt.Errorf("ipsec block is required for encapsulation 'ipsec'")
		}
		if tunnel.L2TP != nil {
			return fmt.Errorf("l2tp block is not allowed for encapsulation 'ipsec'")
		}
	case "l2tpv3":
		if tunnel.L2TP == nil {
			return fmt.Errorf("l2tp block is required for encapsulation 'l2tpv3'")
		}
		// IPsec is optional for L2TPv3
	case "l2tp":
		if tunnel.IPsec == nil {
			return fmt.Errorf("ipsec block is required for encapsulation 'l2tp'")
		}
		if tunnel.L2TP == nil {
			return fmt.Errorf("l2tp block is required for encapsulation 'l2tp'")
		}
	}

	// Note: pre_shared_key validation is handled by Terraform schema (Required: true)
	// We don't validate it here because it's a WriteOnly attribute and won't be
	// available during Update operations.

	return nil
}
