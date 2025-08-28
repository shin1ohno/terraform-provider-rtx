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
	routePattern := regexp.MustCompile(`^ip route (.+)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}

		// Check if this line matches a static route command
		if routePattern.MatchString(line) {
			route, err := ParseStaticRoute(line)
			if err != nil {
				return nil, fmt.Errorf("failed to parse static route line '%s': %w", line, err)
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
	// Pattern: ip route <dest> gateway <gw> [interface <if>] [metric <m>] [weight <w>]
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

	// Find gateway keyword
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

	// Parse gateway - could be IP address or interface
	gatewayValue := parts[gatewayIdx+1]
	
	// Check for "gateway interface <name>" format
	if gatewayValue == "interface" {
		if gatewayIdx+2 >= len(parts) {
			return nil, fmt.Errorf("missing interface name after 'gateway interface'")
		}
		route.GatewayInterface = parts[gatewayIdx+2]
		gatewayIdx++ // Skip the "interface" word
	} else {
		// Check if it's an IP address or interface name
		if net.ParseIP(gatewayValue) != nil {
			route.GatewayIP = gatewayValue
		} else {
			// Assume it's an interface name
			route.GatewayInterface = gatewayValue
		}
	}

	// Parse remaining options
	for i := gatewayIdx + 2; i < len(parts); i++ {
		switch parts[i] {
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
			route.Weight = weight
			i++ // Skip the weight value
		}
	}

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

	// Add gateway
	cmd.WriteString(" gateway ")
	if route.GatewayIP != "" {
		cmd.WriteString(route.GatewayIP)
	} else if route.GatewayInterface != "" {
		cmd.WriteString(route.GatewayInterface)
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

	// Add weight if specified
	if route.Weight > 0 {
		cmd.WriteString(" weight ")
		cmd.WriteString(strconv.Itoa(route.Weight))
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

	// Validate gateway
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
	// RTX interface pattern: wan1, lan1, pp1, tunnel1, loopback1, etc.
	pattern := regexp.MustCompile(`^(wan|lan|pp|tunnel|loopback)\d+$`)
	return pattern.MatchString(name)
}