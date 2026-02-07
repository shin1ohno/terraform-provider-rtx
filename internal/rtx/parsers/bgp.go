package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// BGPConfig represents BGP configuration on an RTX router
type BGPConfig struct {
	Enabled               bool          `json:"enabled"`
	ASN                   string        `json:"asn"`                              // String for 4-byte ASN support
	RouterID              string        `json:"router_id,omitempty"`              // Optional router ID
	DefaultIPv4Unicast    bool          `json:"default_ipv4_unicast"`             // Default: true
	LogNeighborChanges    bool          `json:"log_neighbor_changes"`             // Default: true
	Neighbors             []BGPNeighbor `json:"neighbors,omitempty"`              // BGP neighbors
	Networks              []BGPNetwork  `json:"networks,omitempty"`               // Announced networks
	RedistributeStatic    bool          `json:"redistribute_static,omitempty"`    // Redistribute static routes
	RedistributeConnected bool          `json:"redistribute_connected,omitempty"` // Redistribute connected routes
}

// BGPNeighbor represents a BGP neighbor configuration
type BGPNeighbor struct {
	ID           int    `json:"id"`                      // Neighbor ID (1-based)
	IP           string `json:"ip"`                      // Neighbor IP address
	RemoteAS     string `json:"remote_as"`               // Remote AS number
	HoldTime     int    `json:"hold_time,omitempty"`     // Hold time in seconds
	Keepalive    int    `json:"keepalive,omitempty"`     // Keepalive interval
	Multihop     int    `json:"multihop,omitempty"`      // eBGP multihop TTL
	Password     string `json:"password,omitempty"`      // MD5 authentication password (pre-shared-key)
	LocalAddress string `json:"local_address,omitempty"` // Local address for session
	Passive      bool   `json:"passive,omitempty"`       // Passive mode
}

// BGPNetwork represents a BGP network announcement
type BGPNetwork struct {
	Prefix string `json:"prefix"` // Network prefix
	Mask   string `json:"mask"`   // Network mask (dotted decimal)
}

// BGPParser parses BGP configuration output
type BGPParser struct{}

// NewBGPParser creates a new BGP parser
func NewBGPParser() *BGPParser {
	return &BGPParser{}
}

// ParseBGPConfig parses the output of "show config | grep bgp" command
func (p *BGPParser) ParseBGPConfig(raw string) (*BGPConfig, error) {
	config := &BGPConfig{
		Enabled:            false,
		DefaultIPv4Unicast: true,
		LogNeighborChanges: true,
		Neighbors:          []BGPNeighbor{},
		Networks:           []BGPNetwork{},
	}

	lines := strings.Split(raw, "\n")
	neighbors := make(map[int]*BGPNeighbor)

	// Patterns for BGP configuration
	// Reference: RTX Command Reference
	bgpUsePattern := regexp.MustCompile(`^\s*bgp\s+use\s+(on|off)\s*$`)
	bgpASNPattern := regexp.MustCompile(`^\s*bgp\s+autonomous-system\s+(\d+)\s*$`)
	bgpRouterIDPattern := regexp.MustCompile(`^\s*bgp\s+router\s+id\s+([0-9.]+)\s*$`)
	// New format: bgp neighbor <n> <as> <ip> [options...]
	// Example: bgp neighbor 1 65002 203.0.113.1
	// Example: bgp neighbor 1 65002 192.168.1.2 hold-time=90 local-address=192.168.1.1 passive=on
	bgpNeighborPattern := regexp.MustCompile(`^\s*bgp\s+neighbor\s+(\d+)\s+(\d+)\s+([0-9.]+)(.*)$`)
	// Pre-shared key: bgp neighbor pre-shared-key <n> text <password>
	bgpNeighborPreSharedKeyPattern := regexp.MustCompile(`^\s*bgp\s+neighbor\s+pre-shared-key\s+(\d+)\s+text\s+(.+)\s*$`)
	bgpImportFilterPattern := regexp.MustCompile(`^\s*bgp\s+import\s+filter\s+\d+\s+include\s+([0-9.]+)/(\d+)\s*$`)
	bgpImportStaticPattern := regexp.MustCompile(`^\s*bgp\s+import\s+from\s+static\s*$`)
	bgpImportConnectedPattern := regexp.MustCompile(`^\s*bgp\s+import\s+from\s+connected\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// BGP use on/off
		if matches := bgpUsePattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.Enabled = matches[1] == "on"
			continue
		}

		// BGP AS number
		if matches := bgpASNPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.ASN = matches[1]
			continue
		}

		// BGP router ID
		if matches := bgpRouterIDPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.RouterID = matches[1]
			continue
		}

		// BGP neighbor: bgp neighbor <n> <as> <ip> [options...]
		if matches := bgpNeighborPattern.FindStringSubmatch(line); len(matches) >= 4 {
			id, _ := strconv.Atoi(matches[1])
			neighbor, exists := neighbors[id]
			if !exists {
				neighbor = &BGPNeighbor{ID: id}
				neighbors[id] = neighbor
			}
			neighbor.RemoteAS = matches[2]
			neighbor.IP = matches[3]

			// Parse optional parameters (e.g., hold-time=90 local-address=192.168.1.1 passive=on)
			if len(matches) > 4 && matches[4] != "" {
				options := strings.TrimSpace(matches[4])
				parseNeighborOptions(neighbor, options)
			}
			continue
		}

		// BGP neighbor pre-shared-key: bgp neighbor pre-shared-key <n> text <password>
		if matches := bgpNeighborPreSharedKeyPattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, _ := strconv.Atoi(matches[1])
			neighbor, exists := neighbors[id]
			if !exists {
				neighbor = &BGPNeighbor{ID: id}
				neighbors[id] = neighbor
			}
			neighbor.Password = strings.TrimSpace(matches[2])
			continue
		}

		// BGP import filter (network announcement)
		// Input format: bgp import filter <n> include <prefix>/<cidr>
		// Store as: Prefix=IP, Mask=dotted decimal
		if matches := bgpImportFilterPattern.FindStringSubmatch(line); len(matches) >= 3 {
			cidr, _ := strconv.Atoi(matches[2])
			mask := cidrToMask(cidr)
			config.Networks = append(config.Networks, BGPNetwork{
				Prefix: matches[1],
				Mask:   mask,
			})
			continue
		}

		// BGP redistribute static
		if bgpImportStaticPattern.MatchString(line) {
			config.RedistributeStatic = true
			continue
		}

		// BGP redistribute connected
		if bgpImportConnectedPattern.MatchString(line) {
			config.RedistributeConnected = true
			continue
		}
	}

	// Convert neighbors map to slice
	for _, neighbor := range neighbors {
		config.Neighbors = append(config.Neighbors, *neighbor)
	}

	return config, nil
}

// parseNeighborOptions parses inline options from neighbor command
// Example options: hold-time=90 local-address=192.168.1.1 passive=on
func parseNeighborOptions(neighbor *BGPNeighbor, options string) {
	// Parse key=value pairs
	parts := strings.Fields(options)
	for _, part := range parts {
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) != 2 {
				continue
			}
			key := kv[0]
			value := kv[1]

			switch key {
			case "hold-time":
				neighbor.HoldTime, _ = strconv.Atoi(value)
			case "keepalive":
				neighbor.Keepalive, _ = strconv.Atoi(value)
			case "local-address":
				neighbor.LocalAddress = value
			case "passive":
				neighbor.Passive = (value == "on")
			case "multihop":
				neighbor.Multihop, _ = strconv.Atoi(value)
			}
		}
	}
}

// BuildBGPUseCommand builds the command to enable/disable BGP
// Command format: bgp use on/off
func BuildBGPUseCommand(enabled bool) string {
	if enabled {
		return "bgp use on"
	}
	return "bgp use off"
}

// BuildBGPASNCommand builds the command to set the AS number
// Command format: bgp autonomous-system <asn>
func BuildBGPASNCommand(asn string) string {
	return fmt.Sprintf("bgp autonomous-system %s", asn)
}

// BuildBGPRouterIDCommand builds the command to set the router ID
// Command format: bgp router id <router_id>
func BuildBGPRouterIDCommand(routerID string) string {
	return fmt.Sprintf("bgp router id %s", routerID)
}

// BuildBGPNeighborCommand builds the command to configure a BGP neighbor
// Command format: bgp neighbor <n> <as> <ip> [options...]
// Reference: RTX Command Reference
func BuildBGPNeighborCommand(neighbor BGPNeighbor) string {
	cmd := fmt.Sprintf("bgp neighbor %d %s %s", neighbor.ID, neighbor.RemoteAS, neighbor.IP)

	// Add optional inline parameters
	if neighbor.HoldTime > 0 {
		cmd += fmt.Sprintf(" hold-time=%d", neighbor.HoldTime)
	}
	if neighbor.LocalAddress != "" {
		cmd += fmt.Sprintf(" local-address=%s", neighbor.LocalAddress)
	}
	if neighbor.Passive {
		cmd += " passive=on"
	}

	return cmd
}

// BuildBGPNeighborPreSharedKeyCommand builds the command to set neighbor MD5 password
// Command format: bgp neighbor pre-shared-key <n> text <password>
// Reference: RTX Command Reference
func BuildBGPNeighborPreSharedKeyCommand(neighborID int, password string) string {
	return fmt.Sprintf("bgp neighbor pre-shared-key %d text %s", neighborID, password)
}

// BuildBGPNeighborPasswordCommand is an alias for BuildBGPNeighborPreSharedKeyCommand
// for backward compatibility
// Command format: bgp neighbor pre-shared-key <n> text <password>
func BuildBGPNeighborPasswordCommand(neighborID int, password string) string {
	return BuildBGPNeighborPreSharedKeyCommand(neighborID, password)
}

// BuildBGPNetworkCommand builds the command to announce a network
// Command format: bgp import filter <n> include <prefix>/<cidr>
// Reference: RTX uses CIDR notation (e.g., 192.168.0.0/16)
func BuildBGPNetworkCommand(filterID int, network BGPNetwork) string {
	// Convert dotted decimal mask to CIDR if needed
	cidr := maskToCIDR(network.Mask)
	if cidr == 0 {
		// Assume it's already a CIDR string
		return fmt.Sprintf("bgp import filter %d include %s/%s", filterID, network.Prefix, network.Mask)
	}
	return fmt.Sprintf("bgp import filter %d include %s/%d", filterID, network.Prefix, cidr)
}

// BuildBGPRedistributeCommand builds the command for route redistribution
// Command format: bgp import from static|connected
func BuildBGPRedistributeCommand(routeType string) string {
	return fmt.Sprintf("bgp import from %s", routeType)
}

// BuildDeleteBGPNeighborCommand builds the command to delete a BGP neighbor
// Command format: no bgp neighbor <n>
func BuildDeleteBGPNeighborCommand(neighborID int) string {
	return fmt.Sprintf("no bgp neighbor %d", neighborID)
}

// BuildDeleteBGPNetworkCommand builds the command to remove a network announcement
// Command format: no bgp import filter <n>
func BuildDeleteBGPNetworkCommand(filterID int) string {
	return fmt.Sprintf("no bgp import filter %d", filterID)
}

// BuildDeleteBGPRedistributeCommand removes route redistribution
// Command format: no bgp import from static|connected
func BuildDeleteBGPRedistributeCommand(routeType string) string {
	return fmt.Sprintf("no bgp import from %s", routeType)
}

// BuildShowBGPConfigCommand builds the command to show BGP configuration
func BuildShowBGPConfigCommand() string {
	return "show config | grep bgp"
}

// ValidateBGPConfig validates a BGP configuration
func ValidateBGPConfig(config BGPConfig) error {
	if config.ASN == "" {
		return fmt.Errorf("asn is required")
	}

	// Validate ASN (1-65535 for 2-byte ASN only - RTX limitation)
	asn, err := strconv.ParseUint(config.ASN, 10, 16)
	if err != nil || asn == 0 || asn > 65535 {
		return fmt.Errorf("invalid AS number: must be between 1 and 65535")
	}

	// Validate router ID if provided
	if config.RouterID != "" && !isValidIP(config.RouterID) {
		return fmt.Errorf("invalid router_id: must be a valid IPv4 address")
	}

	// Validate neighbors
	for _, neighbor := range config.Neighbors {
		if neighbor.ID <= 0 {
			return fmt.Errorf("neighbor id must be positive")
		}
		if !isValidIP(neighbor.IP) {
			return fmt.Errorf("invalid neighbor ip: %s", neighbor.IP)
		}
		if neighbor.RemoteAS == "" {
			return fmt.Errorf("neighbor remote_as is required")
		}
		remoteASN, err := strconv.ParseUint(neighbor.RemoteAS, 10, 16)
		if err != nil || remoteASN == 0 || remoteASN > 65535 {
			return fmt.Errorf("invalid neighbor remote_as: must be between 1 and 65535")
		}
		if neighbor.LocalAddress != "" && !isValidIP(neighbor.LocalAddress) {
			return fmt.Errorf("invalid neighbor local_address: %s", neighbor.LocalAddress)
		}
		if neighbor.HoldTime != 0 && (neighbor.HoldTime < 3 || neighbor.HoldTime > 28800) {
			return fmt.Errorf("neighbor hold_time must be between 3 and 28800")
		}
		if neighbor.Keepalive != 0 && (neighbor.Keepalive < 1 || neighbor.Keepalive > 21845) {
			return fmt.Errorf("neighbor keepalive must be between 1 and 21845")
		}
		if neighbor.Multihop != 0 && (neighbor.Multihop < 1 || neighbor.Multihop > 255) {
			return fmt.Errorf("neighbor multihop must be between 1 and 255")
		}
	}

	// Validate networks
	for _, network := range config.Networks {
		if !isValidIP(network.Prefix) {
			return fmt.Errorf("invalid network prefix: %s", network.Prefix)
		}
		if !isValidIP(network.Mask) {
			return fmt.Errorf("invalid network mask: %s", network.Mask)
		}
	}

	return nil
}
