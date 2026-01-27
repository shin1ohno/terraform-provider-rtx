package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// IPsecTunnel represents an IPsec tunnel configuration on an RTX router
type IPsecTunnel struct {
	ID             int            `json:"id"`                      // Tunnel ID
	Name           string         `json:"name,omitempty"`          // Description/name
	LocalAddress   string         `json:"local_address"`           // Local endpoint IP
	RemoteAddress  string         `json:"remote_address"`          // Remote endpoint IP
	PreSharedKey   string         `json:"pre_shared_key"`          // IKE pre-shared key
	IKEv2Proposal  IKEv2Proposal  `json:"ikev2_proposal"`          // IKE Phase 1 proposal
	IPsecTransform IPsecTransform `json:"ipsec_transform"`         // IPsec Phase 2 transform
	LocalNetwork   string         `json:"local_network"`           // Local network CIDR
	RemoteNetwork  string         `json:"remote_network"`          // Remote network CIDR
	DPDEnabled     bool           `json:"dpd_enabled"`             // Dead Peer Detection enabled
	DPDInterval    int            `json:"dpd_interval,omitempty"`  // DPD interval in seconds
	DPDRetry       int            `json:"dpd_retry,omitempty"`     // DPD retry count
	Enabled        bool           `json:"enabled"`                 // Tunnel enabled
	SAPolicy       int            `json:"sa_policy,omitempty"`     // SA policy number
	IKELocalID     string         `json:"ike_local_id,omitempty"`  // IKE local ID
	IKERemoteID    string         `json:"ike_remote_id,omitempty"` // IKE remote ID
	NATTraversal   bool           `json:"nat_traversal,omitempty"` // NAT-T enabled
	PFSGroup       string         `json:"pfs_group,omitempty"`     // PFS DH group
}

// IKEv2Proposal represents IKE Phase 1 proposal settings
type IKEv2Proposal struct {
	EncryptionAES256 bool `json:"encryption_aes256"` // Use AES-256 encryption
	EncryptionAES128 bool `json:"encryption_aes128"` // Use AES-128 encryption
	Encryption3DES   bool `json:"encryption_3des"`   // Use 3DES encryption
	IntegritySHA256  bool `json:"integrity_sha256"`  // Use SHA-256 integrity
	IntegritySHA1    bool `json:"integrity_sha1"`    // Use SHA-1 integrity
	IntegrityMD5     bool `json:"integrity_md5"`     // Use MD5 integrity
	GroupFourteen    bool `json:"group_fourteen"`    // DH group 14 (2048-bit)
	GroupFive        bool `json:"group_five"`        // DH group 5 (1536-bit)
	GroupTwo         bool `json:"group_two"`         // DH group 2 (1024-bit)
	LifetimeSeconds  int  `json:"lifetime_seconds"`  // SA lifetime in seconds
}

// IPsecTransform represents IPsec Phase 2 transform settings
type IPsecTransform struct {
	Protocol         string `json:"protocol"`           // esp or ah
	EncryptionAES256 bool   `json:"encryption_aes256"`  // Use AES-256 encryption
	EncryptionAES128 bool   `json:"encryption_aes128"`  // Use AES-128 encryption
	Encryption3DES   bool   `json:"encryption_3des"`    // Use 3DES encryption
	IntegritySHA256  bool   `json:"integrity_sha256"`   // Use SHA-256-HMAC
	IntegritySHA1    bool   `json:"integrity_sha1"`     // Use SHA-1-HMAC
	IntegrityMD5     bool   `json:"integrity_md5"`      // Use MD5-HMAC
	PFSGroupFourteen bool   `json:"pfs_group_fourteen"` // PFS DH group 14
	PFSGroupFive     bool   `json:"pfs_group_five"`     // PFS DH group 5
	PFSGroupTwo      bool   `json:"pfs_group_two"`      // PFS DH group 2
	LifetimeSeconds  int    `json:"lifetime_seconds"`   // SA lifetime in seconds
}

// IPsecTunnelParser parses IPsec tunnel configuration output
type IPsecTunnelParser struct{}

// NewIPsecTunnelParser creates a new IPsec tunnel parser
func NewIPsecTunnelParser() *IPsecTunnelParser {
	return &IPsecTunnelParser{}
}

// ParseIPsecTunnelConfig parses the output of "show config" for IPsec tunnels
func (p *IPsecTunnelParser) ParseIPsecTunnelConfig(raw string) ([]IPsecTunnel, error) {
	tunnels := make(map[int]*IPsecTunnel)
	lines := strings.Split(raw, "\n")

	// Patterns for IPsec configuration
	tunnelSelectPattern := regexp.MustCompile(`^\s*tunnel\s+select\s+(\d+)\s*$`)
	ipsecTunnelPattern := regexp.MustCompile(`^\s*ipsec\s+tunnel\s+(\d+)\s*$`)
	ipsecSAPolicyPattern := regexp.MustCompile(`^\s*ipsec\s+sa\s+policy\s+(\d+)\s+(\d+)\s+(\w+)\s+(.+)\s*$`)
	ipsecIKELocalAddrPattern := regexp.MustCompile(`^\s*ipsec\s+ike\s+local\s+address\s+(\d+)\s+([0-9.]+)\s*$`)
	ipsecIKERemoteAddrPattern := regexp.MustCompile(`^\s*ipsec\s+ike\s+remote\s+address\s+(\d+)\s+([0-9.]+)\s*$`)
	ipsecIKEPreSharedKeyPattern := regexp.MustCompile(`^\s*ipsec\s+ike\s+pre-shared-key\s+(\d+)\s+text\s+(.+)\s*$`)
	ipsecIKEEncryptionPattern := regexp.MustCompile(`^\s*ipsec\s+ike\s+encryption\s+(\d+)\s+(.+)\s*$`)
	ipsecIKEHashPattern := regexp.MustCompile(`^\s*ipsec\s+ike\s+hash\s+(\d+)\s+(.+)\s*$`)
	ipsecIKEGroupPattern := regexp.MustCompile(`^\s*ipsec\s+ike\s+group\s+(\d+)\s+(.+)\s*$`)
	ipsecIKEKeepalivePattern := regexp.MustCompile(`^\s*ipsec\s+ike\s+keepalive\s+use\s+(\d+)\s+on\s+dpd\s+(\d+)\s*$`)
	ipsecIKEKeepaliveRetryPattern := regexp.MustCompile(`^\s*ipsec\s+ike\s+keepalive\s+use\s+(\d+)\s+on\s+dpd\s+(\d+)\s+(\d+)\s*$`)
	tunnelDescriptionPattern := regexp.MustCompile(`^\s*description\s+(.+)\s*$`)

	var currentTunnelID int

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Tunnel select
		if matches := tunnelSelectPattern.FindStringSubmatch(line); len(matches) >= 2 {
			id, _ := strconv.Atoi(matches[1])
			currentTunnelID = id
			if _, exists := tunnels[id]; !exists {
				tunnels[id] = &IPsecTunnel{
					ID:            id,
					DPDEnabled:    true,
					DPDInterval:   30,
					Enabled:       true,
					IKEv2Proposal: IKEv2Proposal{LifetimeSeconds: 28800},
					IPsecTransform: IPsecTransform{
						Protocol:        "esp",
						LifetimeSeconds: 3600,
					},
				}
			}
			continue
		}

		// IPsec tunnel
		if matches := ipsecTunnelPattern.FindStringSubmatch(line); len(matches) >= 2 {
			id, _ := strconv.Atoi(matches[1])
			if _, exists := tunnels[id]; !exists {
				tunnels[id] = &IPsecTunnel{
					ID:            id,
					DPDEnabled:    true,
					DPDInterval:   30,
					Enabled:       true,
					IKEv2Proposal: IKEv2Proposal{LifetimeSeconds: 28800},
					IPsecTransform: IPsecTransform{
						Protocol:        "esp",
						LifetimeSeconds: 3600,
					},
				}
			}
			continue
		}

		// IPsec SA policy
		// Format: ipsec sa policy <policy_num> <sa_num> <protocol> <algorithms>
		// The policy_num corresponds to the ipsec tunnel ID (e.g., ipsec tunnel 101 uses ipsec sa policy 101)
		if matches := ipsecSAPolicyPattern.FindStringSubmatch(line); len(matches) >= 5 {
			policyNum, _ := strconv.Atoi(matches[1])
			// matches[2] is the SA number (not tunnel ID)
			protocol := matches[3]
			algorithms := matches[4]

			// Use policy number to find the matching tunnel (ipsec tunnel ID = SA policy number)
			if tunnel, exists := tunnels[policyNum]; exists {
				tunnel.SAPolicy = policyNum
				tunnel.IPsecTransform.Protocol = protocol
				parseIPsecSAAlgorithms(algorithms, &tunnel.IPsecTransform)
			}
			continue
		}

		// IPsec IKE local address
		if matches := ipsecIKELocalAddrPattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, _ := strconv.Atoi(matches[1])
			if tunnel, exists := tunnels[id]; exists {
				tunnel.LocalAddress = matches[2]
			}
			continue
		}

		// IPsec IKE remote address
		if matches := ipsecIKERemoteAddrPattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, _ := strconv.Atoi(matches[1])
			if tunnel, exists := tunnels[id]; exists {
				tunnel.RemoteAddress = matches[2]
			}
			continue
		}

		// IPsec IKE pre-shared-key
		if matches := ipsecIKEPreSharedKeyPattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, _ := strconv.Atoi(matches[1])
			if tunnel, exists := tunnels[id]; exists {
				tunnel.PreSharedKey = strings.TrimSpace(matches[2])
			}
			continue
		}

		// IPsec IKE encryption
		if matches := ipsecIKEEncryptionPattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, _ := strconv.Atoi(matches[1])
			if tunnel, exists := tunnels[id]; exists {
				parseIKEEncryption(matches[2], &tunnel.IKEv2Proposal)
			}
			continue
		}

		// IPsec IKE hash
		if matches := ipsecIKEHashPattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, _ := strconv.Atoi(matches[1])
			if tunnel, exists := tunnels[id]; exists {
				parseIKEHash(matches[2], &tunnel.IKEv2Proposal)
			}
			continue
		}

		// IPsec IKE group
		if matches := ipsecIKEGroupPattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, _ := strconv.Atoi(matches[1])
			if tunnel, exists := tunnels[id]; exists {
				parseIKEGroup(matches[2], &tunnel.IKEv2Proposal)
			}
			continue
		}

		// IPsec IKE keepalive (DPD) with retry
		if matches := ipsecIKEKeepaliveRetryPattern.FindStringSubmatch(line); len(matches) >= 4 {
			id, _ := strconv.Atoi(matches[1])
			if tunnel, exists := tunnels[id]; exists {
				tunnel.DPDEnabled = true
				tunnel.DPDInterval, _ = strconv.Atoi(matches[2])
				tunnel.DPDRetry, _ = strconv.Atoi(matches[3])
			}
			continue
		}

		// IPsec IKE keepalive (DPD)
		if matches := ipsecIKEKeepalivePattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, _ := strconv.Atoi(matches[1])
			if tunnel, exists := tunnels[id]; exists {
				tunnel.DPDEnabled = true
				tunnel.DPDInterval, _ = strconv.Atoi(matches[2])
			}
			continue
		}

		// Description (within tunnel context)
		if currentTunnelID > 0 {
			if matches := tunnelDescriptionPattern.FindStringSubmatch(line); len(matches) >= 2 {
				if tunnel, exists := tunnels[currentTunnelID]; exists {
					tunnel.Name = strings.TrimSpace(matches[1])
				}
				continue
			}
		}
	}

	// Convert map to slice
	result := make([]IPsecTunnel, 0, len(tunnels))
	for _, tunnel := range tunnels {
		result = append(result, *tunnel)
	}

	return result, nil
}

// parseIKEEncryption parses IKE encryption algorithm string
func parseIKEEncryption(enc string, proposal *IKEv2Proposal) {
	enc = strings.ToLower(enc)
	if strings.Contains(enc, "aes256") || strings.Contains(enc, "aes-cbc-256") {
		proposal.EncryptionAES256 = true
	}
	if strings.Contains(enc, "aes128") || strings.Contains(enc, "aes-cbc-128") || strings.Contains(enc, "aes-cbc") {
		proposal.EncryptionAES128 = true
	}
	if strings.Contains(enc, "3des") {
		proposal.Encryption3DES = true
	}
}

// parseIKEHash parses IKE hash algorithm string
func parseIKEHash(hash string, proposal *IKEv2Proposal) {
	hash = strings.ToLower(hash)
	if strings.Contains(hash, "sha256") || strings.Contains(hash, "sha2-256") {
		proposal.IntegritySHA256 = true
	}
	if strings.Contains(hash, "sha") && !strings.Contains(hash, "sha256") && !strings.Contains(hash, "sha2") {
		proposal.IntegritySHA1 = true
	}
	if strings.Contains(hash, "md5") {
		proposal.IntegrityMD5 = true
	}
}

// parseIKEGroup parses IKE DH group string
func parseIKEGroup(group string, proposal *IKEv2Proposal) {
	group = strings.ToLower(group)
	if strings.Contains(group, "modp2048") || strings.Contains(group, "14") {
		proposal.GroupFourteen = true
	}
	if strings.Contains(group, "modp1536") || strings.Contains(group, "5") {
		proposal.GroupFive = true
	}
	if strings.Contains(group, "modp1024") || strings.Contains(group, "2") {
		proposal.GroupTwo = true
	}
}

// parseIPsecSAAlgorithms parses IPsec SA policy algorithm string
// Format: "aes-cbc sha-hmac" or "aes-cbc-256 sha256-hmac" etc.
func parseIPsecSAAlgorithms(algs string, transform *IPsecTransform) {
	algs = strings.ToLower(algs)

	// Parse encryption algorithm
	if strings.Contains(algs, "aes-cbc-256") || strings.Contains(algs, "aes256") {
		transform.EncryptionAES256 = true
	} else if strings.Contains(algs, "aes-cbc") || strings.Contains(algs, "aes128") || strings.Contains(algs, "aes") {
		// aes-cbc without 256 suffix means 128-bit
		transform.EncryptionAES128 = true
	}
	if strings.Contains(algs, "3des") {
		transform.Encryption3DES = true
	}

	// Parse integrity/hash algorithm
	if strings.Contains(algs, "sha256-hmac") || strings.Contains(algs, "sha2-256-hmac") {
		transform.IntegritySHA256 = true
	} else if strings.Contains(algs, "sha-hmac") {
		// sha-hmac without 256 means SHA-1
		transform.IntegritySHA1 = true
	}
	if strings.Contains(algs, "md5-hmac") {
		transform.IntegrityMD5 = true
	}
}

// BuildTunnelSelectCommand builds the command to select a tunnel
// Command format: tunnel select <n>
func BuildTunnelSelectCommand(tunnelID int) string {
	return fmt.Sprintf("tunnel select %d", tunnelID)
}

// BuildIPsecTunnelCommand builds the command to create an IPsec tunnel
// Command format: ipsec tunnel <n>
func BuildIPsecTunnelCommand(tunnelID int) string {
	return fmt.Sprintf("ipsec tunnel %d", tunnelID)
}

// BuildIPsecIKELocalAddressCommand builds the command to set local address
// Command format: ipsec ike local address <n> <ip>
func BuildIPsecIKELocalAddressCommand(tunnelID int, ip string) string {
	return fmt.Sprintf("ipsec ike local address %d %s", tunnelID, ip)
}

// BuildIPsecIKERemoteAddressCommand builds the command to set remote address
// Command format: ipsec ike remote address <n> <ip>
func BuildIPsecIKERemoteAddressCommand(tunnelID int, ip string) string {
	return fmt.Sprintf("ipsec ike remote address %d %s", tunnelID, ip)
}

// BuildIPsecIKEPreSharedKeyCommand builds the command to set pre-shared key
// Command format: ipsec ike pre-shared-key <n> text <key>
func BuildIPsecIKEPreSharedKeyCommand(tunnelID int, key string) string {
	return fmt.Sprintf("ipsec ike pre-shared-key %d text %s", tunnelID, key)
}

// BuildIPsecIKEEncryptionCommand builds the command to set IKE encryption
// Command format: ipsec ike encryption <n> <algorithm>
func BuildIPsecIKEEncryptionCommand(tunnelID int, proposal IKEv2Proposal) string {
	var alg string
	if proposal.EncryptionAES256 {
		alg = "aes-cbc-256"
	} else if proposal.EncryptionAES128 {
		alg = "aes-cbc"
	} else if proposal.Encryption3DES {
		alg = "3des-cbc"
	} else {
		alg = "aes-cbc" // default
	}
	return fmt.Sprintf("ipsec ike encryption %d %s", tunnelID, alg)
}

// BuildIPsecIKEHashCommand builds the command to set IKE hash algorithm
// Command format: ipsec ike hash <n> <algorithm>
func BuildIPsecIKEHashCommand(tunnelID int, proposal IKEv2Proposal) string {
	var alg string
	if proposal.IntegritySHA256 {
		alg = "sha256"
	} else if proposal.IntegritySHA1 {
		alg = "sha"
	} else if proposal.IntegrityMD5 {
		alg = "md5"
	} else {
		alg = "sha256" // default
	}
	return fmt.Sprintf("ipsec ike hash %d %s", tunnelID, alg)
}

// BuildIPsecIKEGroupCommand builds the command to set IKE DH group
// Command format: ipsec ike group <n> <group>
func BuildIPsecIKEGroupCommand(tunnelID int, proposal IKEv2Proposal) string {
	var group string
	if proposal.GroupFourteen {
		group = "modp2048"
	} else if proposal.GroupFive {
		group = "modp1536"
	} else if proposal.GroupTwo {
		group = "modp1024"
	} else {
		group = "modp2048" // default
	}
	return fmt.Sprintf("ipsec ike group %d %s", tunnelID, group)
}

// BuildIPsecSAPolicyCommand builds the command to set SA policy
// Command format: ipsec sa policy <policy_num> <tunnel_id> <protocol> <algorithms>
func BuildIPsecSAPolicyCommand(policyNum, tunnelID int, transform IPsecTransform) string {
	protocol := transform.Protocol
	if protocol == "" {
		protocol = "esp"
	}

	var enc, hash string
	if transform.EncryptionAES256 {
		enc = "aes-cbc-256"
	} else if transform.EncryptionAES128 {
		enc = "aes-cbc"
	} else if transform.Encryption3DES {
		enc = "3des-cbc"
	} else {
		enc = "aes-cbc"
	}

	if transform.IntegritySHA256 {
		hash = "sha256-hmac"
	} else if transform.IntegritySHA1 {
		hash = "sha-hmac"
	} else if transform.IntegrityMD5 {
		hash = "md5-hmac"
	} else {
		hash = "sha256-hmac"
	}

	return fmt.Sprintf("ipsec sa policy %d %d %s %s %s", policyNum, tunnelID, protocol, enc, hash)
}

// BuildIPsecIKEKeepaliveCommand builds the command to enable DPD
// Command format: ipsec ike keepalive use <n> on dpd <interval> [retry]
func BuildIPsecIKEKeepaliveCommand(tunnelID, interval, retry int) string {
	if retry > 0 {
		return fmt.Sprintf("ipsec ike keepalive use %d on dpd %d %d", tunnelID, interval, retry)
	}
	return fmt.Sprintf("ipsec ike keepalive use %d on dpd %d", tunnelID, interval)
}

// BuildIPsecIKEKeepaliveOffCommand builds the command to disable DPD
// Command format: ipsec ike keepalive use <n> off
func BuildIPsecIKEKeepaliveOffCommand(tunnelID int) string {
	return fmt.Sprintf("ipsec ike keepalive use %d off", tunnelID)
}

// BuildDeleteIPsecTunnelCommand builds the command to delete an IPsec tunnel
// Command format: no ipsec tunnel <n>
func BuildDeleteIPsecTunnelCommand(tunnelID int) string {
	return fmt.Sprintf("no ipsec tunnel %d", tunnelID)
}

// BuildDeleteTunnelSelectCommand builds the command to delete tunnel select
// Command format: no tunnel select <n>
func BuildDeleteTunnelSelectCommand(tunnelID int) string {
	return fmt.Sprintf("no tunnel select %d", tunnelID)
}

// BuildShowIPsecConfigCommand builds the command to show IPsec configuration
func BuildShowIPsecConfigCommand() string {
	return "show config | grep ipsec"
}

// ValidateIPsecTunnel validates an IPsec tunnel configuration
func ValidateIPsecTunnel(tunnel IPsecTunnel) error {
	if tunnel.ID <= 0 {
		return fmt.Errorf("tunnel id must be positive")
	}

	// LocalAddress is optional - when empty, the router uses appropriate local address
	if tunnel.LocalAddress != "" && !isValidIP(tunnel.LocalAddress) {
		return fmt.Errorf("invalid local_address: %s", tunnel.LocalAddress)
	}

	// RemoteAddress is optional for some IPsec configurations (e.g., L2TP anonymous)
	if tunnel.RemoteAddress != "" && !isValidIP(tunnel.RemoteAddress) {
		return fmt.Errorf("invalid remote_address: %s", tunnel.RemoteAddress)
	}

	// PreSharedKey is optional for some IPsec configurations (e.g., when using certificates
	// or when defined elsewhere like in transport mode for L2TP)
	// If provided, just ensure it's not empty string after being explicitly set

	if tunnel.LocalNetwork != "" && !isValidCIDR(tunnel.LocalNetwork) {
		return fmt.Errorf("invalid local_network: %s", tunnel.LocalNetwork)
	}

	if tunnel.RemoteNetwork != "" && !isValidCIDR(tunnel.RemoteNetwork) {
		return fmt.Errorf("invalid remote_network: %s", tunnel.RemoteNetwork)
	}

	if tunnel.DPDInterval < 0 || tunnel.DPDInterval > 3600 {
		return fmt.Errorf("dpd_interval must be between 0 and 3600")
	}

	return nil
}
