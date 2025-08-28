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

	// Add all supported parameters
	var options []string

	// Add gateway option
	if scope.Gateway != "" {
		options = append(options, "gateway", scope.Gateway)
	}

	// Add DNS servers option  
	if len(scope.DNSServers) > 0 {
		options = append(options, "dns")
		options = append(options, scope.DNSServers...)
	}

	// Add lease time option
	if scope.Lease > 0 {
		// RTX expects lease time in seconds, not minutes
		options = append(options, "lease", fmt.Sprintf("%d", scope.Lease))
	}

	// Add domain name option
	if scope.DomainName != "" {
		options = append(options, "domain", scope.DomainName)
	}

	if len(options) > 0 {
		return fmt.Sprintf("%s %s", baseCmd, strings.Join(options, " "))
	}

	return baseCmd
}

// BuildDHCPScopeCreateCommands builds multiple commands to create a DHCP scope with all options
// This function is deprecated in favor of single command approach
func BuildDHCPScopeCreateCommands(scope DhcpScope) []string {
	// Return single command instead of multiple commands
	cmd := BuildDHCPScopeCreateCommand(scope)
	return []string{cmd}
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

	// Build single command with all options
	cmd := BuildDHCPScopeCreateCommand(scope)
	return []string{cmd}, nil
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

	// Since createCmds now only contains one command, we can simplify
	updateCmds := []string{deleteCmd, createCmds[0]}

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
