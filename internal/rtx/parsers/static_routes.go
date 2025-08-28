package parsers

import (
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

// ParseStaticRoutes parses RTX static route configuration from raw text
func ParseStaticRoutes(raw []byte) ([]StaticRoute, error) {
	var routes []StaticRoute
	scanner := bufio.NewScanner(strings.NewReader(string(raw)))

	// Regex pattern for static route commands
	// Match "ip route <destination>" to catch both valid and potentially invalid route lines
	routePattern := regexp.MustCompile(`^ip route\s+(\S+)(?:\s+(.*))?$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}

		// Skip non-route lines like "ip route change log on", "ip routing process fast", etc.
		if strings.Contains(line, "ip route change") ||
			strings.Contains(line, "ip routing") ||
			strings.Contains(line, "ip filter") {
			continue
		}

		// Check if this line matches a static route command
		if routePattern.MatchString(line) {
			route, err := ParseStaticRoute(line)
			if err != nil {
				// Skip lines that don't parse as valid routes
				// This handles cases like "ip route change log on" that match the pattern
				// but aren't actual static route definitions
				continue
			}
			routes = append(routes, *route)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan input: %w", err)
	}

	return routes, nil
}

// ParseStaticRoute parses a single static route command line
func ParseStaticRoute(line string) (*StaticRoute, error) {
	if line == "" {
		return nil, fmt.Errorf("empty route command")
	}

	// Parse the command using regex
	// Pattern: ip route <dest> gateway <gw1> [gateway <gw2>] [interface <if>] [metric <m>] [weight <w>]
	parts := strings.Fields(line)
	if len(parts) < 4 || parts[0] != "ip" || parts[1] != "route" {
		return nil, fmt.Errorf("invalid route command format: %s", line)
	}

	route := &StaticRoute{
		Metric:      1, // Default metric
		Weight:      0, // Default weight
		Description: "",
		Hide:        false,
	}

	// Parse destination
	destination := parts[2]
	if destination == "default" {
		route.Destination = "0.0.0.0/0"
	} else {
		// Validate CIDR notation
		if _, _, err := net.ParseCIDR(destination); err != nil {
			return nil, fmt.Errorf("invalid destination CIDR '%s': %w", destination, err)
		}
		route.Destination = destination
	}

	// Find first gateway keyword (RTX supports multiple gateways)
	gatewayIdx := -1
	for i, part := range parts {
		if part == "gateway" {
			gatewayIdx = i
			break
		}
	}

	if gatewayIdx == -1 || gatewayIdx+1 >= len(parts) {
		return nil, fmt.Errorf("missing gateway in route command")
	}

	// Parse first gateway - could be IP address or interface
	gatewayValue := parts[gatewayIdx+1]
	
	// Check for "gateway interface <name>" format
	if gatewayValue == "interface" {
		if gatewayIdx+2 >= len(parts) {
			return nil, fmt.Errorf("missing interface name after 'gateway interface'")
		}
		route.GatewayInterface = parts[gatewayIdx+2]
		gatewayIdx++ // Skip the "interface" word
	} else {
		// Check if it's an IP address, special RTX gateway, or interface name
		if net.ParseIP(gatewayValue) != nil {
			// It's a valid IP address
			route.GatewayIP = gatewayValue
		} else if gatewayValue == "dhcp" {
			// RTX allows "gateway dhcp" and "gateway dhcp <interface>" for dynamic default routes
			// Check if next part is an interface for "gateway dhcp lan2" format
			if gatewayIdx+2 < len(parts) {
				nextPart := parts[gatewayIdx+2]
				// Check if next part looks like an interface name and is not a keyword
				if isValidInterfaceName(nextPart) && nextPart != "interface" && 
					nextPart != "metric" && nextPart != "weight" && nextPart != "gateway" {
					// Store as "dhcp lan2" format
					route.GatewayInterface = fmt.Sprintf("dhcp %s", nextPart)
					gatewayIdx++ // Skip this interface part in main parsing loop
				} else {
					// Just "gateway dhcp"
					route.GatewayInterface = gatewayValue
				}
			} else {
				// Just "gateway dhcp"
				route.GatewayInterface = gatewayValue
			}
		} else {
			// Assume it's an interface name
			route.GatewayInterface = gatewayValue
		}
	}

	// Create first gateway from parsed info
	var gateways []Gateway
	firstGateway := Gateway{
		Weight: 1, // Default weight
		Hide:   false,
	}
	
	if route.GatewayIP != "" {
		firstGateway.IP = route.GatewayIP
	} else if route.GatewayInterface != "" {
		firstGateway.Interface = route.GatewayInterface
	}
	gateways = append(gateways, firstGateway)

	// Parse remaining options including additional gateways
	for i := gatewayIdx + 2; i < len(parts); i++ {
		switch parts[i] {
		case "gateway":
			// Parse additional gateway
			if i+1 >= len(parts) {
				return nil, fmt.Errorf("missing gateway value after 'gateway'")
			}
			gatewayValue := parts[i+1]
			newGateway := Gateway{
				Weight: 1,    // Default weight
				Hide:   false, // Default hide
			}
			
			// Check for "gateway interface <name>" format
			if gatewayValue == "interface" {
				if i+2 >= len(parts) {
					return nil, fmt.Errorf("missing interface name after 'gateway interface'")
				}
				newGateway.Interface = parts[i+2]
				i++ // Skip the "interface" word
			} else if gatewayValue == "dhcp" {
				// Handle "gateway dhcp" and "gateway dhcp <interface>"
				if i+2 < len(parts) {
					nextPart := parts[i+2]
					// Check if next part is an interface name and not a keyword
					if isValidInterfaceName(nextPart) && nextPart != "interface" && 
						nextPart != "metric" && nextPart != "weight" && nextPart != "gateway" {
						// Store as "dhcp lan2" format
						newGateway.Interface = fmt.Sprintf("dhcp %s", nextPart)
						i++ // Skip the interface part
					} else {
						// Just "gateway dhcp"
						newGateway.Interface = gatewayValue
					}
				} else {
					// Just "gateway dhcp"
					newGateway.Interface = gatewayValue
				}
			} else if net.ParseIP(gatewayValue) != nil {
				// It's a valid IP address
				newGateway.IP = gatewayValue
			} else {
				// Assume it's an interface name
				newGateway.Interface = gatewayValue
			}
			
			gateways = append(gateways, newGateway)
			i++ // Skip the gateway value
			
		case "interface":
			if i+1 >= len(parts) {
				return nil, fmt.Errorf("missing interface name")
			}
			route.Interface = parts[i+1]
			i++ // Skip the interface name
			
		case "metric":
			if i+1 >= len(parts) {
				return nil, fmt.Errorf("missing metric value")
			}
			metric, err := strconv.Atoi(parts[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid metric value '%s': %w", parts[i+1], err)
			}
			route.Metric = metric
			i++ // Skip the metric value
			
		case "weight":
			if i+1 >= len(parts) {
				return nil, fmt.Errorf("missing weight value")
			}
			weight, err := strconv.Atoi(parts[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid weight value '%s': %w", parts[i+1], err)
			}
			// Apply weight to the last gateway
			if len(gateways) > 0 {
				gateways[len(gateways)-1].Weight = weight
			}
			i++ // Skip the weight value
			
		case "hide":
			// Apply hide to the last gateway
			if len(gateways) > 0 {
				gateways[len(gateways)-1].Hide = true
			}
		}
	}
	
	route.Gateways = gateways

	// Validate the parsed route
	if err := ValidateStaticRoute(*route); err != nil {
		return nil, fmt.Errorf("invalid route configuration: %w", err)
	}

	return route, nil
}

// BuildStaticRouteCommand builds RTX command for creating/updating a static route
func BuildStaticRouteCommand(route StaticRoute) string {
	var cmd strings.Builder

	// Start with ip route
	cmd.WriteString("ip route ")

	// Add destination
	if route.Destination == "0.0.0.0/0" {
		cmd.WriteString("default")
	} else {
		cmd.WriteString(route.Destination)
	}

	// Add gateways - use new format if available, fallback to legacy
	if len(route.Gateways) > 0 {
		for _, gw := range route.Gateways {
			cmd.WriteString(" gateway ")
			if gw.IP != "" {
				cmd.WriteString(gw.IP)
			} else if gw.Interface != "" {
				cmd.WriteString(gw.Interface)
			}
			
			// Add gateway-specific weight if not default
			if gw.Weight > 1 {
				cmd.WriteString(" weight ")
				cmd.WriteString(strconv.Itoa(gw.Weight))
			}
			
			// Add gateway-specific hide
			if gw.Hide {
				cmd.WriteString(" hide")
			}
		}
	} else {
		// Fallback to legacy fields for backwards compatibility
		cmd.WriteString(" gateway ")
		if route.GatewayIP != "" {
			cmd.WriteString(route.GatewayIP)
		} else if route.GatewayInterface != "" {
			cmd.WriteString(route.GatewayInterface)
		}
		
		// Add legacy weight if specified
		if route.Weight > 1 {
			cmd.WriteString(" weight ")
			cmd.WriteString(strconv.Itoa(route.Weight))
		}
		
		// Add legacy hide
		if route.Hide {
			cmd.WriteString(" hide")
		}
	}

	// Add interface if specified
	if route.Interface != "" {
		cmd.WriteString(" interface ")
		cmd.WriteString(route.Interface)
	}

	// Add metric if not default
	if route.Metric != 1 && route.Metric > 0 {
		cmd.WriteString(" metric ")
		cmd.WriteString(strconv.Itoa(route.Metric))
	}

	return cmd.String()
}

// BuildStaticRouteDeleteCommand builds RTX command for deleting a static route
func BuildStaticRouteDeleteCommand(route StaticRoute) string {
	var cmd strings.Builder

	// Start with no ip route
	cmd.WriteString("no ip route ")

	// Add destination
	if route.Destination == "0.0.0.0/0" {
		cmd.WriteString("default")
	} else {
		cmd.WriteString(route.Destination)
	}

	// Add gateway
	cmd.WriteString(" gateway ")
	if route.GatewayIP != "" {
		cmd.WriteString(route.GatewayIP)
	} else if route.GatewayInterface != "" {
		cmd.WriteString(route.GatewayInterface)
	}

	// Add interface if specified (needed for proper deletion)
	if route.Interface != "" {
		cmd.WriteString(" interface ")
		cmd.WriteString(route.Interface)
	}

	return cmd.String()
}

// ValidateStaticRoute validates a static route configuration
func ValidateStaticRoute(route StaticRoute) error {
	// Validate destination
	if route.Destination == "" {
		return fmt.Errorf("destination cannot be empty")
	}

	if route.Destination != "0.0.0.0/0" {
		if _, _, err := net.ParseCIDR(route.Destination); err != nil {
			return fmt.Errorf("invalid destination CIDR '%s': %w", route.Destination, err)
		}
	}

	// Validate gateway - check new format first, then fallback to legacy
	if len(route.Gateways) > 0 {
		// Validate each gateway in the new format
		for i, gw := range route.Gateways {
			hasIP := gw.IP != ""
			hasInterface := gw.Interface != ""
			
			if !hasIP && !hasInterface {
				return fmt.Errorf("gateway %d: either ip or interface must be specified", i)
			}
			
			if hasIP && hasInterface {
				return fmt.Errorf("gateway %d: cannot specify both ip and interface", i)
			}
			
			if hasIP {
				if net.ParseIP(gw.IP) == nil {
					return fmt.Errorf("gateway %d: invalid IP address '%s'", i, gw.IP)
				}
			}
			
			if hasInterface {
				if !isValidInterfaceName(gw.Interface) {
					return fmt.Errorf("gateway %d: invalid interface name '%s'", i, gw.Interface)
				}
			}
		}
	} else {
		// Fallback to legacy validation
		hasGatewayIP := route.GatewayIP != ""
		hasGatewayInterface := route.GatewayInterface != ""

		if !hasGatewayIP && !hasGatewayInterface {
			return fmt.Errorf("either gateway_ip or gateway_interface must be specified")
		}

		if hasGatewayIP && hasGatewayInterface {
			return fmt.Errorf("cannot specify both gateway_ip and gateway_interface")
		}

		// Validate gateway IP if specified
		if hasGatewayIP {
			if net.ParseIP(route.GatewayIP) == nil {
				return fmt.Errorf("invalid gateway IP address '%s'", route.GatewayIP)
			}
		}

		// Validate interface name if specified
		if hasGatewayInterface {
			if !isValidInterfaceName(route.GatewayInterface) {
				return fmt.Errorf("invalid gateway interface name '%s'", route.GatewayInterface)
			}
		}
	}

	if route.Interface != "" {
		if !isValidInterfaceName(route.Interface) {
			return fmt.Errorf("invalid interface name '%s'", route.Interface)
		}
	}

	// Validate metric range
	if route.Metric < 0 || route.Metric > 65535 {
		return fmt.Errorf("metric must be between 0 and 65535, got %d", route.Metric)
	}

	// Validate weight range
	if route.Weight < 0 || route.Weight > 255 {
		return fmt.Errorf("weight must be between 0 and 255, got %d", route.Weight)
	}

	return nil
}

// isValidInterfaceName checks if the interface name follows RTX naming convention
func isValidInterfaceName(name string) bool {
	// RTX interface patterns:
	// - wan1, lan1, pp1, tunnel1, loopback1, etc.
	// - dhcp (for gateway dhcp)
	// - dhcp lan2 (for gateway dhcp lan2)
	validPatterns := []string{
		`^(wan|lan|pp|tunnel|loopback)\d+$`,  // Standard interfaces
		`^dhcp$`,                             // DHCP gateway
		`^dhcp\s+(wan|lan|pp|tunnel)\d+$`,   // DHCP with specific interface
	}
	
	for _, pattern := range validPatterns {
		if matched, _ := regexp.MatchString(pattern, name); matched {
			return true
		}
	}
	
	return false
}