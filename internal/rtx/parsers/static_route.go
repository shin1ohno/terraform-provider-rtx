package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// StaticRoute represents a static route configuration on an RTX router
type StaticRoute struct {
	Prefix   string    `json:"prefix"`    // Route destination (e.g., "0.0.0.0" for default)
	Mask     string    `json:"mask"`      // Subnet mask (e.g., "0.0.0.0" for default)
	NextHops []NextHop `json:"next_hops"` // List of next hops
}

// NextHop represents a next hop configuration for a static route
type NextHop struct {
	NextHop   string `json:"next_hop,omitempty"`  // Gateway IP address
	Interface string `json:"interface,omitempty"` // Interface (pp 1, tunnel 1, etc.)
	Distance  int    `json:"distance"`            // Administrative distance (weight)
	Name      string `json:"name,omitempty"`      // Route description
	Permanent bool   `json:"permanent"`           // Keep route when interface down
	Filter    int    `json:"filter,omitempty"`    // IP filter number (RTX-specific)
}

// StaticRouteParser parses static route configuration output
type StaticRouteParser struct{}

// NewStaticRouteParser creates a new static route parser
func NewStaticRouteParser() *StaticRouteParser {
	return &StaticRouteParser{}
}

// ParseRouteConfig parses the output of "show config | grep ip route" command
// and returns a list of static routes grouped by prefix/mask
func (p *StaticRouteParser) ParseRouteConfig(raw string) ([]StaticRoute, error) {
	routes := make(map[string]*StaticRoute) // key: "prefix/mask"
	lines := strings.Split(raw, "\n")

	// Pattern for ip route commands
	// ip route default gateway <gateway> [weight <n>] [filter <n>] [hide]
	// ip route <network>/<prefix> gateway <gateway> [weight <n>] [filter <n>] [hide]
	// ip route <network>/<prefix> gateway pp <n> [weight <n>] [filter <n>] [hide]
	// ip route <network>/<prefix> gateway tunnel <n> [weight <n>] [filter <n>] [hide]
	// ip route <network>/<prefix> gateway dhcp <interface>
	routePattern := regexp.MustCompile(`^\s*ip\s+route\s+(\S+)\s+gateway\s+(.+?)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip 'no ip route' lines
		if strings.HasPrefix(line, "no ") {
			continue
		}

		matches := routePattern.FindStringSubmatch(line)
		if len(matches) < 3 {
			continue
		}

		network := matches[1]
		gatewayPart := matches[2]

		// Parse network: "default" or "x.x.x.x/y" or "x.x.x.x/x.x.x.x"
		prefix, mask := parseNetworkNotation(network)
		if prefix == "" {
			continue
		}

		routeKey := fmt.Sprintf("%s/%s", prefix, mask)

		// Create or get existing route
		route, exists := routes[routeKey]
		if !exists {
			route = &StaticRoute{
				Prefix:   prefix,
				Mask:     mask,
				NextHops: []NextHop{},
			}
			routes[routeKey] = route
		}

		// Parse gateway/next hop - handle multiple gateways on same line (ECMP)
		// RTX syntax: "gateway 192.168.1.20 gateway 192.168.1.21"
		gatewayParts := splitOnGateway(gatewayPart)
		for _, gw := range gatewayParts {
			hop := parseGatewayPart(gw)
			route.NextHops = append(route.NextHops, hop)
		}
	}

	// Convert map to slice
	result := make([]StaticRoute, 0, len(routes))
	for _, route := range routes {
		result = append(result, *route)
	}

	return result, nil
}

// ParseSingleRoute parses configuration for a specific route prefix/mask
func (p *StaticRouteParser) ParseSingleRoute(raw string, prefix, mask string) (*StaticRoute, error) {
	routes, err := p.ParseRouteConfig(raw)
	if err != nil {
		return nil, err
	}

	for _, route := range routes {
		if route.Prefix == prefix && route.Mask == mask {
			return &route, nil
		}
	}

	return nil, fmt.Errorf("route %s/%s not found", prefix, mask)
}

// parseNetworkNotation converts various network formats to prefix/mask
// Supports: "default", "x.x.x.x/y" (CIDR), "x.x.x.x/x.x.x.x" (dotted mask)
func parseNetworkNotation(network string) (prefix, mask string) {
	if network == "default" {
		return "0.0.0.0", "0.0.0.0"
	}

	parts := strings.Split(network, "/")
	if len(parts) != 2 {
		return "", ""
	}

	prefix = parts[0]
	if !isValidIP(prefix) {
		return "", ""
	}

	// Check if mask is CIDR length or dotted decimal
	if strings.Contains(parts[1], ".") {
		// Dotted decimal mask
		mask = parts[1]
		if !isValidIP(mask) {
			return "", ""
		}
	} else {
		// CIDR prefix length
		prefixLen, err := strconv.Atoi(parts[1])
		if err != nil || prefixLen < 0 || prefixLen > 32 {
			return "", ""
		}
		mask = cidrToMask(prefixLen)
	}

	return prefix, mask
}

// cidrToMask converts CIDR prefix length to dotted decimal mask
func cidrToMask(prefixLen int) string {
	if prefixLen < 0 || prefixLen > 32 {
		return ""
	}

	mask := uint32(0xFFFFFFFF) << (32 - prefixLen)
	return fmt.Sprintf("%d.%d.%d.%d",
		(mask>>24)&0xFF,
		(mask>>16)&0xFF,
		(mask>>8)&0xFF,
		mask&0xFF)
}

// maskToCIDR converts dotted decimal mask to CIDR prefix length
func maskToCIDR(mask string) int {
	parts := strings.Split(mask, ".")
	if len(parts) != 4 {
		return -1
	}

	bits := 0
	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil || num < 0 || num > 255 {
			return -1
		}
		for i := 7; i >= 0; i-- {
			if (num & (1 << i)) != 0 {
				bits++
			}
		}
	}
	return bits
}

// parseGatewayPart parses the gateway specification from a route command
func parseGatewayPart(gatewayPart string) NextHop {
	hop := NextHop{
		Distance: 1, // Default distance
	}

	// Tokenize the gateway part
	tokens := strings.Fields(gatewayPart)
	if len(tokens) == 0 {
		return hop
	}

	i := 0

	// Parse gateway type
	switch {
	case tokens[i] == "pp" && i+1 < len(tokens):
		hop.Interface = fmt.Sprintf("pp %s", tokens[i+1])
		i += 2
	case tokens[i] == "tunnel" && i+1 < len(tokens):
		hop.Interface = fmt.Sprintf("tunnel %s", tokens[i+1])
		i += 2
	case tokens[i] == "dhcp" && i+1 < len(tokens):
		hop.Interface = fmt.Sprintf("dhcp %s", tokens[i+1])
		i += 2
	case tokens[i] == "null":
		hop.Interface = "null"
		i++
	case tokens[i] == "loopback":
		hop.Interface = "loopback"
		i++
	case isValidIP(tokens[i]):
		hop.NextHop = tokens[i]
		i++
	default:
		// Unknown format, try to use as-is
		hop.Interface = tokens[i]
		i++
	}

	// Parse optional parameters
	for i < len(tokens) {
		switch tokens[i] {
		case "weight":
			if i+1 < len(tokens) {
				weight, err := strconv.Atoi(tokens[i+1])
				if err == nil {
					hop.Distance = weight
				}
				i += 2
			} else {
				i++
			}
		case "filter":
			if i+1 < len(tokens) {
				filter, err := strconv.Atoi(tokens[i+1])
				if err == nil {
					hop.Filter = filter
				}
				i += 2
			} else {
				i++
			}
		case "hide":
			hop.Permanent = false
			i++
		case "keepalive":
			hop.Permanent = true
			i++
		case "name":
			// Collect name until next keyword or end
			if i+1 < len(tokens) {
				i++
				nameParts := []string{}
				for i < len(tokens) && !isKeyword(tokens[i]) {
					nameParts = append(nameParts, tokens[i])
					i++
				}
				hop.Name = strings.Join(nameParts, " ")
			} else {
				i++
			}
		default:
			i++
		}
	}

	return hop
}

// isKeyword checks if a token is a known parameter keyword
func isKeyword(token string) bool {
	keywords := []string{"weight", "filter", "hide", "keepalive", "name"}
	for _, kw := range keywords {
		if token == kw {
			return true
		}
	}
	return false
}

// splitOnGateway splits a gateway specification into multiple parts for ECMP routes
// Input: "192.168.1.20 gateway 192.168.1.21" or "192.168.1.20 weight 1 gateway 192.168.1.21 weight 1"
// Output: ["192.168.1.20", "192.168.1.21"] or ["192.168.1.20 weight 1", "192.168.1.21 weight 1"]
func splitOnGateway(gatewayPart string) []string {
	// Split on " gateway " to handle multiple next hops
	parts := strings.Split(gatewayPart, " gateway ")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	if len(result) == 0 {
		return []string{gatewayPart}
	}
	return result
}

// BuildIPRouteCommand builds the command to create a static route
// Command format: ip route <network> gateway <gateway> [weight <n>] [filter <n>] [hide] [keepalive]
func BuildIPRouteCommand(route StaticRoute, hop NextHop) string {
	// Determine network string
	network := formatNetworkNotation(route.Prefix, route.Mask)

	var cmd strings.Builder
	cmd.WriteString(fmt.Sprintf("ip route %s gateway ", network))

	// Add gateway
	if hop.Interface != "" {
		cmd.WriteString(hop.Interface)
	} else if hop.NextHop != "" {
		cmd.WriteString(hop.NextHop)
	} else {
		// Should not happen, but handle gracefully
		return ""
	}

	// Add optional parameters
	if hop.Distance > 1 {
		cmd.WriteString(fmt.Sprintf(" weight %d", hop.Distance))
	}

	if hop.Filter > 0 {
		cmd.WriteString(fmt.Sprintf(" filter %d", hop.Filter))
	}

	if hop.Permanent {
		cmd.WriteString(" keepalive")
	}

	// Note: name parameter is typically not supported in standard ip route command
	// It's more of a Terraform-only tracking field

	return cmd.String()
}

// BuildDeleteIPRouteCommand builds the command to delete a static route
// Command format: no ip route <network> [gateway <gateway>]
func BuildDeleteIPRouteCommand(prefix, mask string, hop *NextHop) string {
	network := formatNetworkNotation(prefix, mask)

	if hop == nil {
		// Delete all routes for this network
		return fmt.Sprintf("no ip route %s", network)
	}

	// Delete specific next hop
	var cmd strings.Builder
	cmd.WriteString(fmt.Sprintf("no ip route %s gateway ", network))

	if hop.Interface != "" {
		cmd.WriteString(hop.Interface)
	} else if hop.NextHop != "" {
		cmd.WriteString(hop.NextHop)
	}

	return cmd.String()
}

// BuildShowIPRouteConfigCommand builds the command to show static route configuration
// Command format: show config | grep "ip route"
func BuildShowIPRouteConfigCommand() string {
	return `show config | grep "ip route"`
}

// BuildShowSingleRouteConfigCommand builds the command to show a specific route's configuration
func BuildShowSingleRouteConfigCommand(prefix, mask string) string {
	network := formatNetworkNotation(prefix, mask)
	return fmt.Sprintf(`show config | grep "ip route %s"`, network)
}

// formatNetworkNotation formats prefix/mask for RTX command
// Converts "0.0.0.0/0.0.0.0" to "default"
// Converts other addresses to CIDR notation "x.x.x.x/y"
func formatNetworkNotation(prefix, mask string) string {
	if prefix == "0.0.0.0" && mask == "0.0.0.0" {
		return "default"
	}

	// Convert mask to CIDR prefix length
	prefixLen := maskToCIDR(mask)
	if prefixLen < 0 {
		// Fallback to dotted notation
		return fmt.Sprintf("%s/%s", prefix, mask)
	}

	return fmt.Sprintf("%s/%d", prefix, prefixLen)
}

// ValidateStaticRoute validates a static route configuration
func ValidateStaticRoute(route StaticRoute) error {
	if route.Prefix == "" {
		return fmt.Errorf("prefix is required")
	}

	if !isValidIP(route.Prefix) {
		return fmt.Errorf("invalid prefix: %s", route.Prefix)
	}

	if route.Mask == "" {
		return fmt.Errorf("mask is required")
	}

	if !isValidIP(route.Mask) {
		return fmt.Errorf("invalid mask: %s", route.Mask)
	}

	if len(route.NextHops) == 0 {
		return fmt.Errorf("at least one next_hop is required")
	}

	for i, hop := range route.NextHops {
		if err := validateNextHop(hop); err != nil {
			return fmt.Errorf("next_hop[%d]: %w", i, err)
		}
	}

	return nil
}

// validateNextHop validates a single next hop configuration
func validateNextHop(hop NextHop) error {
	// Must have either next_hop (IP) or interface
	if hop.NextHop == "" && hop.Interface == "" {
		return fmt.Errorf("either next_hop or interface must be specified")
	}

	// Validate next_hop IP if specified
	if hop.NextHop != "" && !isValidIP(hop.NextHop) {
		return fmt.Errorf("invalid next_hop IP address: %s", hop.NextHop)
	}

	// Validate interface format if specified
	if hop.Interface != "" {
		if !isValidInterface(hop.Interface) {
			return fmt.Errorf("invalid interface format: %s", hop.Interface)
		}
	}

	// Validate distance (weight) range
	if hop.Distance < 0 || hop.Distance > 100 {
		return fmt.Errorf("distance must be between 0 and 100")
	}

	// Validate filter number
	if hop.Filter < 0 {
		return fmt.Errorf("filter must be non-negative")
	}

	return nil
}

// isValidInterface checks if an interface name is valid for RTX routers
func isValidInterface(iface string) bool {
	validPrefixes := []string{
		"pp ",      // PPPoE/PPP interface
		"tunnel ",  // Tunnel interface
		"dhcp ",    // DHCP-derived gateway
		"lan",      // LAN interface
		"null",     // Null interface (blackhole)
		"loopback", // Loopback interface
	}

	for _, prefix := range validPrefixes {
		if strings.HasPrefix(iface, prefix) || iface == prefix {
			return true
		}
	}

	return false
}
