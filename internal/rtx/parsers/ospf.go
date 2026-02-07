package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// OSPFConfig represents OSPF configuration on an RTX router
type OSPFConfig struct {
	Enabled               bool           `json:"enabled"`
	ProcessID             int            `json:"process_id,omitempty"`             // OSPF process ID (default: 1)
	RouterID              string         `json:"router_id"`                        // Router ID (required)
	Distance              int            `json:"distance,omitempty"`               // Administrative distance (default: 110)
	DefaultOriginate      bool           `json:"default_originate,omitempty"`      // Originate default route
	Networks              []OSPFNetwork  `json:"networks,omitempty"`               // OSPF networks
	Areas                 []OSPFArea     `json:"areas,omitempty"`                  // OSPF areas
	Neighbors             []OSPFNeighbor `json:"neighbors,omitempty"`              // OSPF neighbors (NBMA)
	RedistributeStatic    bool           `json:"redistribute_static,omitempty"`    // Redistribute static routes
	RedistributeConnected bool           `json:"redistribute_connected,omitempty"` // Redistribute connected routes
}

// OSPFNetwork represents an OSPF network configuration
type OSPFNetwork struct {
	IP       string `json:"ip"`       // Network IP address
	Wildcard string `json:"wildcard"` // Wildcard mask
	Area     string `json:"area"`     // Area ID (decimal or dotted decimal)
}

// OSPFArea represents an OSPF area configuration
type OSPFArea struct {
	ID        string `json:"id"`                   // Area ID (decimal or dotted decimal)
	Type      string `json:"type,omitempty"`       // Area type: normal, stub (RTX does not support nssa)
	NoSummary bool   `json:"no_summary,omitempty"` // Totally stubby (no summary LSAs)
}

// OSPFNeighbor represents an OSPF neighbor configuration (for NBMA networks)
type OSPFNeighbor struct {
	IP       string `json:"ip"`                 // Neighbor IP address
	Priority int    `json:"priority,omitempty"` // Neighbor priority (0-255)
	Cost     int    `json:"cost,omitempty"`     // Cost to neighbor
}

// OSPFParser parses OSPF configuration output
type OSPFParser struct{}

// NewOSPFParser creates a new OSPF parser
func NewOSPFParser() *OSPFParser {
	return &OSPFParser{}
}

// ParseOSPFConfig parses the output of "show config | grep ospf" command
func (p *OSPFParser) ParseOSPFConfig(raw string) (*OSPFConfig, error) {
	config := &OSPFConfig{
		Enabled:   false,
		ProcessID: 1,
		Distance:  110,
		Networks:  []OSPFNetwork{},
		Areas:     []OSPFArea{},
		Neighbors: []OSPFNeighbor{},
	}

	lines := strings.Split(raw, "\n")
	areas := make(map[string]*OSPFArea)

	// Patterns for OSPF configuration
	ospfUsePattern := regexp.MustCompile(`^\s*ospf\s+use\s+(on|off)\s*$`)
	ospfRouterIDPattern := regexp.MustCompile(`^\s*ospf\s+router\s+id\s+([0-9.]+)\s*$`)
	ospfAreaPattern := regexp.MustCompile(`^\s*ospf\s+area\s+([0-9.]+)\s*$`)
	ospfAreaStubPattern := regexp.MustCompile(`^\s*ospf\s+area\s+([0-9.]+)\s+stub(?:\s+(no-summary))?\s*$`)
	// Note: RTX does not support NSSA areas
	ospfImportStaticPattern := regexp.MustCompile(`^\s*ospf\s+import\s+from\s+static\s*$`)
	// Note: RTX supports ospf import from static, rip, bgp (not connected)
	// ip <interface> ospf area <area>
	ipOspfAreaPattern := regexp.MustCompile(`^\s*ip\s+(\S+)\s+ospf\s+area\s+([0-9.]+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// OSPF use on/off
		if matches := ospfUsePattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.Enabled = matches[1] == "on"
			continue
		}

		// OSPF router ID
		if matches := ospfRouterIDPattern.FindStringSubmatch(line); len(matches) >= 2 {
			config.RouterID = matches[1]
			continue
		}

		// OSPF area (normal)
		if matches := ospfAreaPattern.FindStringSubmatch(line); len(matches) >= 2 {
			areaID := matches[1]
			if _, exists := areas[areaID]; !exists {
				areas[areaID] = &OSPFArea{ID: areaID, Type: "normal"}
			}
			continue
		}

		// OSPF area stub
		if matches := ospfAreaStubPattern.FindStringSubmatch(line); len(matches) >= 2 {
			areaID := matches[1]
			area, exists := areas[areaID]
			if !exists {
				area = &OSPFArea{ID: areaID}
				areas[areaID] = area
			}
			area.Type = "stub"
			if len(matches) > 2 && matches[2] == "no-summary" {
				area.NoSummary = true
			}
			continue
		}

		// Note: NSSA is not supported by RTX routers

		// ip <interface> ospf area <area>
		if matches := ipOspfAreaPattern.FindStringSubmatch(line); len(matches) >= 3 {
			// This tells us interface is in area, add to networks
			interfaceName := matches[1]
			areaID := matches[2]
			config.Networks = append(config.Networks, OSPFNetwork{
				IP:   interfaceName, // Store interface name as IP for now
				Area: areaID,
			})
			continue
		}

		// OSPF redistribute static
		if ospfImportStaticPattern.MatchString(line) {
			config.RedistributeStatic = true
			continue
		}

		// Note: RTX supports ospf import from static, rip, bgp (not connected)
	}

	// Convert areas map to slice
	for _, area := range areas {
		config.Areas = append(config.Areas, *area)
	}

	return config, nil
}

// BuildOSPFEnableCommand builds the command to enable OSPF
// Command format: ospf use on
func BuildOSPFEnableCommand() string {
	return "ospf use on"
}

// BuildOSPFDisableCommand builds the command to disable OSPF
// Command format: ospf use off
func BuildOSPFDisableCommand() string {
	return "ospf use off"
}

// BuildOSPFRouterIDCommand builds the command to set the router ID
// Command format: ospf router id <router_id>
func BuildOSPFRouterIDCommand(routerID string) string {
	return fmt.Sprintf("ospf router id %s", routerID)
}

// BuildOSPFAreaCommand builds the command to configure an OSPF area
// Command format: ospf area <area_id> [stub] [no-summary]
// Note: RTX does not support NSSA areas
func BuildOSPFAreaCommand(area OSPFArea) string {
	cmd := fmt.Sprintf("ospf area %s", area.ID)

	if area.Type == "stub" {
		cmd += " stub"
		if area.NoSummary {
			cmd += " no-summary"
		}
	}

	return cmd
}

// BuildIPOSPFAreaCommand builds the command to assign interface to OSPF area
// Command format: ip <interface> ospf area <area>
func BuildIPOSPFAreaCommand(interfaceName, areaID string) string {
	return fmt.Sprintf("ip %s ospf area %s", interfaceName, areaID)
}

// BuildOSPFImportCommand builds the command for route redistribution
// Command format: ospf import from static|connected
func BuildOSPFImportCommand(routeType string) string {
	return fmt.Sprintf("ospf import from %s", routeType)
}

// BuildDeleteOSPFAreaCommand builds the command to delete an OSPF area
// Command format: no ospf area <area_id>
func BuildDeleteOSPFAreaCommand(areaID string) string {
	return fmt.Sprintf("no ospf area %s", areaID)
}

// BuildDeleteIPOSPFAreaCommand builds the command to remove interface from OSPF
// Command format: no ip <interface> ospf area
func BuildDeleteIPOSPFAreaCommand(interfaceName string) string {
	return fmt.Sprintf("no ip %s ospf area", interfaceName)
}

// BuildDeleteOSPFImportCommand removes route redistribution
// Command format: no ospf import from static|connected
func BuildDeleteOSPFImportCommand(routeType string) string {
	return fmt.Sprintf("no ospf import from %s", routeType)
}

// BuildShowOSPFConfigCommand builds the command to show OSPF configuration
func BuildShowOSPFConfigCommand() string {
	return "show config | grep ospf"
}

// ValidateOSPFConfig validates an OSPF configuration
func ValidateOSPFConfig(config OSPFConfig) error {
	// Validate router ID
	if config.RouterID == "" {
		return fmt.Errorf("router_id is required")
	}
	if !isValidIP(config.RouterID) {
		return fmt.Errorf("invalid router_id: must be a valid IPv4 address")
	}

	// Validate areas
	for _, area := range config.Areas {
		if !isValidAreaID(area.ID) {
			return fmt.Errorf("invalid area id: %s (must be decimal or dotted decimal)", area.ID)
		}
		if area.Type != "" && area.Type != "normal" && area.Type != "stub" {
			return fmt.Errorf("invalid area type: %s (must be normal or stub)", area.Type)
		}
	}

	// Validate networks
	for _, network := range config.Networks {
		if network.IP != "" && !isValidIP(network.IP) && !strings.HasPrefix(network.IP, "lan") && !strings.HasPrefix(network.IP, "pp") {
			// Allow interface names like lan1, pp1, etc.
			return fmt.Errorf("invalid network ip: %s", network.IP)
		}
		if network.Area != "" && !isValidAreaID(network.Area) {
			return fmt.Errorf("invalid network area: %s", network.Area)
		}
	}

	// Validate neighbors
	for _, neighbor := range config.Neighbors {
		if !isValidIP(neighbor.IP) {
			return fmt.Errorf("invalid neighbor ip: %s", neighbor.IP)
		}
		if neighbor.Priority < 0 || neighbor.Priority > 255 {
			return fmt.Errorf("neighbor priority must be between 0 and 255")
		}
		if neighbor.Cost < 0 || neighbor.Cost > 65535 {
			return fmt.Errorf("neighbor cost must be between 0 and 65535")
		}
	}

	return nil
}

// isValidAreaID checks if a string is a valid OSPF area ID
// Valid formats: decimal (0-4294967295) or dotted decimal (x.x.x.x)
func isValidAreaID(areaID string) bool {
	// Try decimal format
	if num, err := strconv.ParseUint(areaID, 10, 32); err == nil {
		_ = num // Valid decimal
		return true
	}

	// Try dotted decimal format
	return isValidIP(areaID)
}
