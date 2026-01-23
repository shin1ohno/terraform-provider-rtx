package parsers

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DHCPBinding represents a DHCP static lease binding
type DHCPBinding struct {
	ScopeID             int    `json:"scope_id"`
	IPAddress           string `json:"ip_address"`
	MACAddress          string `json:"mac_address"`
	ClientIdentifier    string `json:"client_identifier,omitempty"`
	UseClientIdentifier bool   `json:"use_client_identifier"`
}

// DHCPBindingsParser is the interface for parsing DHCP binding information
type DHCPBindingsParser interface {
	ParseBindings(raw string, scopeID int) ([]DHCPBinding, error)
}

// dhcpBindingsParser handles parsing of DHCP binding output
type dhcpBindingsParser struct{}

// ParseBindings parses the output of "show config | grep 'dhcp scope bind'" command
func (p *dhcpBindingsParser) ParseBindings(raw string, scopeID int) ([]DHCPBinding, error) {
	var bindings []DHCPBinding
	lines := strings.Split(raw, "\n")

	// Regular expressions for different formats
	// RTX830 format: IP [ethernet] MAC (ethernet keyword appears before MAC if present)
	rtx830Pattern := regexp.MustCompile(`^\s*([0-9.]+)\s+(ethernet\s+([0-9a-fA-F:.-]+)|([0-9a-fA-F:.-]+))\s*$`)
	// RTX1210 format: IP MAC Type (Type appears after MAC)
	rtx1210Pattern := regexp.MustCompile(`^([0-9.]+)\s+([0-9a-fA-F:.-]+)\s+(MAC|ethernet)\s*$`)
	// show status dhcp format patterns for RTX1210
	// Line 1: 予約済みアドレス: IP
	// Line 2: (タイプ) クライアントID: (01) MAC
	staticIPPattern := regexp.MustCompile(`^\s*予約済みアドレス:\s*([0-9.]+)\s*$`)
	dynamicIPPattern := regexp.MustCompile(`^\s*割り当て中アドレス:\s*([0-9.]+)\s*$`)
	clientIDPattern := regexp.MustCompile(`^\s*\(タイプ\)\s*クライアントID:\s*\(01\)\s*([0-9a-fA-F\s]+)\s*$`)
	// show config format: dhcp scope bind SCOPE IP [01|ethernet] MAC (with spaces or colons)
	// Example: dhcp scope bind 1 192.168.1.20 01 00 30 93 11 0e 33
	// Example: dhcp scope bind 1 192.168.1.28 24:59:e5:54:5e:5a
	// Example: dhcp scope bind 1 192.168.1.23 ethernet b6:1a:27:ea:28:29
	configPattern := regexp.MustCompile(`^\s*dhcp\s+scope\s+bind\s+(\d+)\s+([0-9.]+)\s+(?:(01\s+)|ethernet\s+)?([0-9a-fA-F:\s]+)\s*$`)

	// For multi-line parsing
	var currentIP string
	var isStatic bool

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip header lines
		if strings.Contains(line, "DHCPスコープ番号") || strings.Contains(line, "ネットワークアドレス") ||
			strings.Contains(line, "ホスト名:") || strings.Contains(line, "リース残時間:") ||
			strings.Contains(line, "No bindings found") {
			continue
		}

		// Check for static IP pattern (予約済みアドレス)
		if matches := staticIPPattern.FindStringSubmatch(line); len(matches) >= 2 {
			currentIP = matches[1]
			isStatic = true
			continue
		}

		// Check for dynamic IP pattern (割り当て中アドレス) - skip these
		if matches := dynamicIPPattern.FindStringSubmatch(line); len(matches) >= 2 {
			currentIP = ""
			isStatic = false
			continue
		}

		// Check for client ID pattern - if we have a static IP, create binding
		if currentIP != "" && isStatic {
			if matches := clientIDPattern.FindStringSubmatch(line); len(matches) >= 2 {
				// Extract MAC address from client ID (remove spaces)
				macStr := strings.ReplaceAll(matches[1], " ", "")

				// Convert to standard format with colons
				var macParts []string
				for i := 0; i < len(macStr); i += 2 {
					if i+2 <= len(macStr) {
						macParts = append(macParts, macStr[i:i+2])
					}
				}
				macAddress := strings.Join(macParts, ":")

				binding := DHCPBinding{
					ScopeID:             scopeID,
					IPAddress:           currentIP,
					MACAddress:          macAddress,
					UseClientIdentifier: true, // Client ID format implies ethernet type
				}

				// Normalize MAC address
				normalizedMAC, err := NormalizeMACAddress(binding.MACAddress)
				if err != nil {
					return nil, fmt.Errorf("invalid MAC address %s: %w", binding.MACAddress, err)
				}
				binding.MACAddress = normalizedMAC

				bindings = append(bindings, binding)
				currentIP = ""
				isStatic = false
				continue
			}
		}

		// Try show config format first
		if matches := configPattern.FindStringSubmatch(line); len(matches) >= 5 {
			// Extract scope ID from the line itself
			extractedScopeID, _ := strconv.Atoi(matches[1])

			// Extract MAC address, handling both space-separated and colon-separated formats
			macStr := strings.TrimSpace(matches[4])

			// Check if it's prefixed with "01" or "ethernet" (client identifier types)
			useClientID := matches[3] != "" || strings.Contains(line, " ethernet ")

			// If MAC is space-separated, convert to colon format
			if strings.Contains(macStr, " ") {
				macParts := strings.Fields(macStr)
				macStr = strings.Join(macParts, ":")
			}

			binding := DHCPBinding{
				ScopeID:             extractedScopeID, // Use extracted scope ID from line
				IPAddress:           matches[2],
				MACAddress:          macStr,
				UseClientIdentifier: useClientID,
			}

			// Normalize MAC address
			normalizedMAC, err := NormalizeMACAddress(binding.MACAddress)
			if err != nil {
				return nil, fmt.Errorf("invalid MAC address %s: %w", binding.MACAddress, err)
			}
			binding.MACAddress = normalizedMAC

			bindings = append(bindings, binding)
			continue
		}

		var binding DHCPBinding
		binding.ScopeID = scopeID

		// Try RTX830 format
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
	// Handle custom client identifier
	if binding.ClientIdentifier != "" {
		// Format: client-id type:hex:hex:...
		// Normalize to lowercase for consistency
		normalizedClientID := strings.ToLower(binding.ClientIdentifier)
		return fmt.Sprintf("dhcp scope bind %d %s client-id %s",
			binding.ScopeID, binding.IPAddress, normalizedClientID)
	}

	// Handle MAC address with UseClientIdentifier flag
	if binding.UseClientIdentifier && binding.MACAddress != "" {
		// Legacy: ethernet MAC format
		return fmt.Sprintf("dhcp scope bind %d %s ethernet %s",
			binding.ScopeID, binding.IPAddress, binding.MACAddress)
	}

	// Plain MAC address binding
	return fmt.Sprintf("dhcp scope bind %d %s %s",
		binding.ScopeID, binding.IPAddress, binding.MACAddress)
}

// BuildDHCPUnbindCommand builds a command to remove a DHCP binding
func BuildDHCPUnbindCommand(scopeID int, ipAddress string) string {
	return fmt.Sprintf("no dhcp scope bind %d %s", scopeID, ipAddress)
}

// BuildShowDHCPBindingsCommand builds a command to show DHCP bindings for a scope
func BuildShowDHCPBindingsCommand(scopeID int) string {
	return fmt.Sprintf("show config | grep \"dhcp scope bind %d\"", scopeID)
}

// BuildDHCPBindCommandWithValidation builds a DHCP bind command with validation
func BuildDHCPBindCommandWithValidation(binding DHCPBinding) (string, error) {
	// Validate required fields
	if binding.ScopeID <= 0 {
		return "", errors.New("invalid scope ID")
	}

	if binding.IPAddress == "" {
		return "", errors.New("IP address is required")
	}

	// Validate client identification
	hasMAC := binding.MACAddress != ""
	hasClientID := binding.ClientIdentifier != ""

	// Special case for empty client identifier validation
	if !hasMAC && binding.ClientIdentifier == "" {
		return "", errors.New("client identifier cannot be empty")
	}

	if !hasMAC && !hasClientID {
		return "", errors.New("either MAC address or client identifier must be specified")
	}

	if hasMAC && hasClientID {
		return "", errors.New("cannot specify both MAC address and client identifier")
	}

	// Validate client identifier format if present
	if hasClientID {
		if err := validateClientIdentifier(binding.ClientIdentifier); err != nil {
			return "", err // Return error directly to match expected messages
		}
	}

	// If MAC address validation is needed, validate it
	if hasMAC {
		if _, err := NormalizeMACAddress(binding.MACAddress); err != nil {
			return "", fmt.Errorf("invalid MAC address: %v", err)
		}
	}

	// Build command using existing function
	return BuildDHCPBindCommand(binding), nil
}

// validateClientIdentifier validates the client identifier format
func validateClientIdentifier(identifier string) error {
	if identifier == "" {
		return errors.New("client identifier cannot be empty")
	}

	// Check format: type:data
	parts := strings.Split(identifier, ":")
	if len(parts) < 2 {
		return errors.New("invalid client identifier format")
	}

	// Check if we have data after the prefix
	if len(parts) == 2 && parts[1] == "" {
		return errors.New("client identifier must have data after type prefix")
	}

	// Check prefix is supported (01, 02, or FF)
	prefix := strings.ToLower(parts[0])
	if prefix != "01" && prefix != "02" && prefix != "ff" {
		return errors.New("unsupported client identifier prefix")
	}

	// Validate each hex part
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		if len(part) != 2 {
			return errors.New("invalid hex characters in client identifier")
		}

		for _, c := range part {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return errors.New("invalid hex characters in client identifier")
			}
		}
	}

	// Check length limit (255 bytes max) - each part represents 1 byte
	if len(parts) > 128 {
		return errors.New("client identifier too long (max 255 bytes)")
	}

	return nil
}
