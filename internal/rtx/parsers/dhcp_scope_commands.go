package parsers

import (
	"fmt"
	"net"
	"strings"
)

// BuildDHCPScopeCreateCommand builds a command to create a DHCP scope
func BuildDHCPScopeCreateCommand(scope DhcpScope) string {
	// Base command: dhcp scope <id> <start-ip>-<end-ip>/<prefix>
	baseCmd := fmt.Sprintf("dhcp scope %d %s-%s/%d",
		scope.ID, scope.RangeStart, scope.RangeEnd, scope.Prefix)

	// Add only basic parameters (gateway and lease)
	var options []string

	if scope.Gateway != "" {
		options = append(options, "gateway", scope.Gateway)
	}

	if scope.Lease > 0 {
		// Convert seconds to minutes for RTX
		leaseMinutes := scope.Lease / 60
		if leaseMinutes <= 0 {
			leaseMinutes = 1 // Minimum lease time
		}
		
		// Set maxexpire to be larger than expire to avoid RTX error
		maxExpireMinutes := leaseMinutes * 7 // 7x expire time for maxexpire
		if maxExpireMinutes < 72*60 { // Ensure at least 72 hours for maxexpire (RTX default)
			maxExpireMinutes = 72 * 60
		}
		
		options = append(options, "expire", fmt.Sprintf("%d", leaseMinutes))
		options = append(options, "maxexpire", fmt.Sprintf("%d", maxExpireMinutes))
	}

	if len(options) > 0 {
		return fmt.Sprintf("%s %s", baseCmd, strings.Join(options, " "))
	}

	return baseCmd
}

// BuildDHCPScopeCreateCommands builds multiple commands to create a DHCP scope with all options
func BuildDHCPScopeCreateCommands(scope DhcpScope) []string {
	commands := []string{}

	// 1. Basic scope creation with gateway and lease
	baseCmd := BuildDHCPScopeCreateCommand(scope)
	commands = append(commands, baseCmd)

	// 2. DNS servers option
	if len(scope.DNSServers) > 0 {
		dnsCmd := fmt.Sprintf("dhcp scope option %d dns=%s", 
			scope.ID, strings.Join(scope.DNSServers, ","))
		commands = append(commands, dnsCmd)
	}

	// 3. Domain name option
	if scope.DomainName != "" {
		domainCmd := fmt.Sprintf("dhcp scope option %d domain=%s", 
			scope.ID, scope.DomainName)
		commands = append(commands, domainCmd)
	}

	return commands
}

// BuildDHCPScopeDeleteCommand builds a command to remove a DHCP scope
func BuildDHCPScopeDeleteCommand(scopeID int) string {
	return fmt.Sprintf("no dhcp scope %d", scopeID)
}

// BuildShowDHCPScopesCommand builds a command to show DHCP scope configurations
func BuildShowDHCPScopesCommand() string {
	return "show config | grep \"dhcp scope\""
}

// BuildDHCPScopeCreateCommandWithValidation builds DHCP scope commands with validation
func BuildDHCPScopeCreateCommandWithValidation(scope DhcpScope) ([]string, error) {
	// Validate required fields
	if scope.ID <= 0 || scope.ID > 255 {
		return nil, fmt.Errorf("scope ID must be between 1 and 255")
	}

	if scope.RangeStart == "" {
		return nil, fmt.Errorf("range_start is required")
	}

	if scope.RangeEnd == "" {
		return nil, fmt.Errorf("range_end is required")
	}

	// Validate IP addresses
	startIP := net.ParseIP(scope.RangeStart)
	if startIP == nil {
		return nil, fmt.Errorf("invalid range_start IP address: %s", scope.RangeStart)
	}

	endIP := net.ParseIP(scope.RangeEnd)
	if endIP == nil {
		return nil, fmt.Errorf("invalid range_end IP address: %s", scope.RangeEnd)
	}

	// Validate prefix
	if scope.Prefix < 8 || scope.Prefix > 32 {
		return nil, fmt.Errorf("prefix must be between 8 and 32")
	}

	// Validate gateway if provided
	if scope.Gateway != "" {
		if net.ParseIP(scope.Gateway) == nil {
			return nil, fmt.Errorf("invalid gateway IP address: %s", scope.Gateway)
		}
	}

	// Validate DNS servers if provided
	for i, dns := range scope.DNSServers {
		if net.ParseIP(dns) == nil {
			return nil, fmt.Errorf("invalid DNS server IP address at index %d: %s", i, dns)
		}
	}

	// Validate lease if provided
	if scope.Lease < 0 {
		return nil, fmt.Errorf("lease must be non-negative")
	}

	// Validate IP range order
	startIPv4 := startIP.To4()
	endIPv4 := endIP.To4()
	if startIPv4 == nil || endIPv4 == nil {
		return nil, fmt.Errorf("only IPv4 addresses are supported")
	}

	// Convert IP to uint32 for comparison
	startNum := ipToUint32(startIPv4)
	endNum := ipToUint32(endIPv4)
	if startNum >= endNum {
		return nil, fmt.Errorf("range_start must be less than range_end")
	}

	// Build commands using new multi-command function
	return BuildDHCPScopeCreateCommands(scope), nil
}

// BuildDHCPScopeUpdateCommand builds a command to update an existing DHCP scope
// RTX routers require scope deletion and recreation for updates to ensure consistency
func BuildDHCPScopeUpdateCommand(scope DhcpScope) ([]string, error) {
	// Validate the scope first using existing validation
	createCmds, err := BuildDHCPScopeCreateCommandWithValidation(scope)
	if err != nil {
		return nil, err
	}
	
	// Return sequence of commands: delete old scope, then create new scope
	deleteCmd := BuildDHCPScopeDeleteCommand(scope.ID)
	
	// Combine delete command with all creation commands
	updateCmds := []string{deleteCmd}
	updateCmds = append(updateCmds, createCmds...)
	
	return updateCmds, nil
}

// ipToUint32 converts an IPv4 address to uint32 for comparison
func ipToUint32(ip net.IP) uint32 {
	ipv4 := ip.To4()
	if ipv4 == nil {
		return 0
	}
	return uint32(ipv4[0])<<24 + uint32(ipv4[1])<<16 + uint32(ipv4[2])<<8 + uint32(ipv4[3])
}