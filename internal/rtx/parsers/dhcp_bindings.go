package parsers

import (
	"fmt"
	"regexp"
	"strings"
)

// DHCPBinding represents a DHCP static lease binding
type DHCPBinding struct {
	ScopeID             int    `json:"scope_id"`
	IPAddress           string `json:"ip_address"`
	MACAddress          string `json:"mac_address"`
	UseClientIdentifier bool   `json:"use_client_identifier"`
}

// DHCPBindingsParser is the interface for parsing DHCP binding information
type DHCPBindingsParser interface {
	ParseBindings(raw string, scopeID int) ([]DHCPBinding, error)
}

// dhcpBindingsParser handles parsing of DHCP binding output
type dhcpBindingsParser struct{}

// ParseBindings parses the output of "show dhcp scope bind {scope_id}" command
func (p *dhcpBindingsParser) ParseBindings(raw string, scopeID int) ([]DHCPBinding, error) {
	var bindings []DHCPBinding
	lines := strings.Split(raw, "\n")
	
	// Regular expressions for different formats
	// RTX830 format: IP [ethernet] MAC (ethernet keyword appears before MAC if present)
	rtx830Pattern := regexp.MustCompile(`^\s*([0-9.]+)\s+(ethernet\s+([0-9a-fA-F:.-]+)|([0-9a-fA-F:.-]+))\s*$`)
	// RTX1210 format: IP MAC Type (Type appears after MAC)
	rtx1210Pattern := regexp.MustCompile(`^([0-9.]+)\s+([0-9a-fA-F:.-]+)\s+(MAC|ethernet)\s*$`)
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Skip header lines
		if strings.Contains(line, "Scope") || strings.Contains(line, "Bindings") || 
		   strings.Contains(line, "IP Address") || strings.Contains(line, "MAC Address") ||
		   strings.Contains(line, "No bindings found") {
			continue
		}
		
		var binding DHCPBinding
		binding.ScopeID = scopeID
		
		// Try RTX830 format first
		if matches := rtx830Pattern.FindStringSubmatch(line); len(matches) >= 3 {
			binding.IPAddress = matches[1]
			
			// Check if matches[3] exists (ethernet case) or use matches[4] (no ethernet case)
			if matches[3] != "" {
				// "ethernet MAC" case
				binding.UseClientIdentifier = true
				binding.MACAddress = matches[3]
			} else if matches[4] != "" {
				// "MAC" case
				binding.UseClientIdentifier = false
				binding.MACAddress = matches[4]
			} else {
				// Fallback (shouldn't happen with correct regex)
				binding.UseClientIdentifier = false
				binding.MACAddress = matches[2]
			}
		} else if matches := rtx1210Pattern.FindStringSubmatch(line); len(matches) >= 4 {
			// Try RTX1210 format
			binding.IPAddress = matches[1]
			binding.MACAddress = matches[2]
			binding.UseClientIdentifier = strings.ToLower(matches[3]) == "ethernet"
		} else {
			// Skip lines that don't match any pattern
			continue
		}
		
		// Normalize MAC address
		normalizedMAC, err := NormalizeMACAddress(binding.MACAddress)
		if err != nil {
			return nil, fmt.Errorf("invalid MAC address %s: %w", binding.MACAddress, err)
		}
		binding.MACAddress = normalizedMAC
		
		bindings = append(bindings, binding)
	}
	
	return bindings, nil
}

// NormalizeMACAddress converts various MAC address formats to standard colon-separated lowercase
func NormalizeMACAddress(mac string) (string, error) {
	// Remove all separators
	cleaned := strings.ToLower(mac)
	cleaned = strings.ReplaceAll(cleaned, ":", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	
	// Validate length
	if len(cleaned) != 12 {
		return "", fmt.Errorf("MAC address must be 12 hex digits, got %d", len(cleaned))
	}
	
	// Validate characters
	if !regexp.MustCompile(`^[0-9a-f]{12}$`).MatchString(cleaned) {
		return "", fmt.Errorf("MAC address contains invalid characters")
	}
	
	// Format with colons
	result := fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		cleaned[0:2], cleaned[2:4], cleaned[4:6],
		cleaned[6:8], cleaned[8:10], cleaned[10:12])
	
	return result, nil
}

// NewDHCPBindingsParser creates a new DHCP bindings parser
func NewDHCPBindingsParser() DHCPBindingsParser {
	return &dhcpBindingsParser{}
}

// BuildDHCPBindCommand builds a command to create a DHCP binding
func BuildDHCPBindCommand(binding DHCPBinding) string {
	if binding.UseClientIdentifier {
		return fmt.Sprintf("dhcp scope bind %d %s ethernet %s",
			binding.ScopeID, binding.IPAddress, binding.MACAddress)
	}
	return fmt.Sprintf("dhcp scope bind %d %s %s",
		binding.ScopeID, binding.IPAddress, binding.MACAddress)
}

// BuildDHCPUnbindCommand builds a command to remove a DHCP binding
func BuildDHCPUnbindCommand(scopeID int, ipAddress string) string {
	return fmt.Sprintf("no dhcp scope bind %d %s", scopeID, ipAddress)
}

// BuildShowDHCPBindingsCommand builds a command to show DHCP bindings for a scope
func BuildShowDHCPBindingsCommand(scopeID int) string {
	return fmt.Sprintf("show dhcp scope bind %d", scopeID)
}