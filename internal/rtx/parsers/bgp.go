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
	Password     string `json:"password,omitempty"`      // MD5 authentication password
	LocalAddress string `json:"local_address,omitempty"` // Local address for session
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
	bgpUsePattern := regexp.MustCompile(`^\s*bgp\s+use\s+(on|off)\s*$`)
	bgpASNPattern := regexp.MustCompile(`^\s*bgp\s+autonomous-system\s+(\d+)\s*$`)
	bgpRouterIDPattern := regexp.MustCompile(`^\s*bgp\s+router\s+id\s+([0-9.]+)\s*$`)
	bgpNeighborAddrPattern := regexp.MustCompile(`^\s*bgp\s+neighbor\s+(\d+)\s+address\s+([0-9.]+)\s+as\s+(\d+)\s*$`)
	bgpNeighborHoldTimePattern := regexp.MustCompile(`^\s*bgp\s+neighbor\s+(\d+)\s+hold-time\s+(\d+)\s*$`)
	bgpNeighborKeepalivePattern := regexp.MustCompile(`^\s*bgp\s+neighbor\s+(\d+)\s+keepalive\s+(\d+)\s*$`)
	bgpNeighborMultihopPattern := regexp.MustCompile(`^\s*bgp\s+neighbor\s+(\d+)\s+multihop\s+(\d+)\s*$`)
	bgpNeighborPasswordPattern := regexp.MustCompile(`^\s*bgp\s+neighbor\s+(\d+)\s+password\s+(.+)\s*$`)
	bgpNeighborLocalAddrPattern := regexp.MustCompile(`^\s*bgp\s+neighbor\s+(\d+)\s+local-address\s+([0-9.]+)\s*$`)
	bgpImportFilterPattern := regexp.MustCompile(`^\s*bgp\s+import\s+filter\s+\d+\s+include\s+([0-9.]+)/([0-9.]+)\s*$`)
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

		// BGP neighbor address and AS
		if matches := bgpNeighborAddrPattern.FindStringSubmatch(line); len(matches) >= 4 {
			id, _ := strconv.Atoi(matches[1])
			neighbor, exists := neighbors[id]
			if !exists {
				neighbor = &BGPNeighbor{ID: id}
				neighbors[id] = neighbor
			}
			neighbor.IP = matches[2]
			neighbor.RemoteAS = matches[3]
			continue
		}

		// BGP neighbor hold-time
		if matches := bgpNeighborHoldTimePattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, _ := strconv.Atoi(matches[1])
			neighbor, exists := neighbors[id]
			if !exists {
				neighbor = &BGPNeighbor{ID: id}
				neighbors[id] = neighbor
			}
			neighbor.HoldTime, _ = strconv.Atoi(matches[2])
			continue
		}

		// BGP neighbor keepalive
		if matches := bgpNeighborKeepalivePattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, _ := strconv.Atoi(matches[1])
			neighbor, exists := neighbors[id]
			if !exists {
				neighbor = &BGPNeighbor{ID: id}
				neighbors[id] = neighbor
			}
			neighbor.Keepalive, _ = strconv.Atoi(matches[2])
			continue
		}

		// BGP neighbor multihop
		if matches := bgpNeighborMultihopPattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, _ := strconv.Atoi(matches[1])
			neighbor, exists := neighbors[id]
			if !exists {
				neighbor = &BGPNeighbor{ID: id}
				neighbors[id] = neighbor
			}
			neighbor.Multihop, _ = strconv.Atoi(matches[2])
			continue
		}

		// BGP neighbor password
		if matches := bgpNeighborPasswordPattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, _ := strconv.Atoi(matches[1])
			neighbor, exists := neighbors[id]
			if !exists {
				neighbor = &BGPNeighbor{ID: id}
				neighbors[id] = neighbor
			}
			neighbor.Password = strings.TrimSpace(matches[2])
			continue
		}

		// BGP neighbor local-address
		if matches := bgpNeighborLocalAddrPattern.FindStringSubmatch(line); len(matches) >= 3 {
			id, _ := strconv.Atoi(matches[1])
			neighbor, exists := neighbors[id]
			if !exists {
				neighbor = &BGPNeighbor{ID: id}
				neighbors[id] = neighbor
			}
			neighbor.LocalAddress = matches[2]
			continue
		}

		// BGP import filter (network announcement)
		if matches := bgpImportFilterPattern.FindStringSubmatch(line); len(matches) >= 3 {
			config.Networks = append(config.Networks, BGPNetwork{
				Prefix: matches[1],
				Mask:   matches[2],
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
// Command format: bgp neighbor <n> address <ip> as <asn>
func BuildBGPNeighborCommand(neighbor BGPNeighbor) string {
	return fmt.Sprintf("bgp neighbor %d address %s as %s", neighbor.ID, neighbor.IP, neighbor.RemoteAS)
}

// BuildBGPNeighborHoldTimeCommand builds the command to set neighbor hold-time
// Command format: bgp neighbor <n> hold-time <seconds>
func BuildBGPNeighborHoldTimeCommand(neighborID, holdTime int) string {
	return fmt.Sprintf("bgp neighbor %d hold-time %d", neighborID, holdTime)
}

// BuildBGPNeighborKeepaliveCommand builds the command to set neighbor keepalive
// Command format: bgp neighbor <n> keepalive <seconds>
func BuildBGPNeighborKeepaliveCommand(neighborID, keepalive int) string {
	return fmt.Sprintf("bgp neighbor %d keepalive %d", neighborID, keepalive)
}

// BuildBGPNeighborMultihopCommand builds the command to set eBGP multihop
// Command format: bgp neighbor <n> multihop <ttl>
func BuildBGPNeighborMultihopCommand(neighborID, ttl int) string {
	return fmt.Sprintf("bgp neighbor %d multihop %d", neighborID, ttl)
}

// BuildBGPNeighborPasswordCommand builds the command to set neighbor MD5 password
// Command format: bgp neighbor <n> password <password>
func BuildBGPNeighborPasswordCommand(neighborID int, password string) string {
	return fmt.Sprintf("bgp neighbor %d password %s", neighborID, password)
}

// BuildBGPNeighborLocalAddressCommand builds the command to set local address
// Command format: bgp neighbor <n> local-address <ip>
func BuildBGPNeighborLocalAddressCommand(neighborID int, localAddress string) string {
	return fmt.Sprintf("bgp neighbor %d local-address %s", neighborID, localAddress)
}

// BuildBGPNetworkCommand builds the command to announce a network
// Command format: bgp import filter <n> include <prefix>/<mask>
func BuildBGPNetworkCommand(filterID int, network BGPNetwork) string {
	return fmt.Sprintf("bgp import filter %d include %s/%s", filterID, network.Prefix, network.Mask)
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

	// Validate ASN (1-4294967295 for 4-byte ASN)
	asn, err := strconv.ParseUint(config.ASN, 10, 32)
	if err != nil || asn == 0 {
		return fmt.Errorf("invalid asn: must be between 1 and 4294967295")
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
		remoteASN, err := strconv.ParseUint(neighbor.RemoteAS, 10, 32)
		if err != nil || remoteASN == 0 {
			return fmt.Errorf("invalid neighbor remote_as: must be between 1 and 4294967295")
		}
		if neighbor.LocalAddress != "" && !isValidIP(neighbor.LocalAddress) {
			return fmt.Errorf("invalid neighbor local_address: %s", neighbor.LocalAddress)
		}
		if neighbor.HoldTime != 0 && (neighbor.HoldTime < 3 || neighbor.HoldTime > 65535) {
			return fmt.Errorf("neighbor hold_time must be between 3 and 65535")
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
