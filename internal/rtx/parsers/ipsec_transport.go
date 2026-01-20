package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// IPsecTransport represents an IPsec transport mode configuration on an RTX router
type IPsecTransport struct {
	TransportID int    `json:"transport_id"` // Transport number
	TunnelID    int    `json:"tunnel_id"`    // Associated tunnel number
	Protocol    string `json:"protocol"`     // Protocol ("udp" typically)
	Port        int    `json:"port"`         // Port number (1701 for L2TP)
}

// ParseIPsecTransportConfig parses the output of "show config" for IPsec transport configurations
// Command format: ipsec transport N M proto port
// where N is transport ID, M is tunnel ID, proto is protocol (e.g., udp), port is port number
func ParseIPsecTransportConfig(raw string) ([]IPsecTransport, error) {
	var transports []IPsecTransport
	lines := strings.Split(raw, "\n")

	// Pattern: ipsec transport <transport_id> <tunnel_id> <protocol> <port>
	transportPattern := regexp.MustCompile(`^\s*ipsec\s+transport\s+(\d+)\s+(\d+)\s+(\w+)\s+(\d+)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if matches := transportPattern.FindStringSubmatch(line); len(matches) >= 5 {
			transportID, _ := strconv.Atoi(matches[1])
			tunnelID, _ := strconv.Atoi(matches[2])
			protocol := matches[3]
			port, _ := strconv.Atoi(matches[4])

			transport := IPsecTransport{
				TransportID: transportID,
				TunnelID:    tunnelID,
				Protocol:    protocol,
				Port:        port,
			}
			transports = append(transports, transport)
		}
	}

	return transports, nil
}

// BuildIPsecTransportCommand builds the command to create an IPsec transport configuration
// Command format: ipsec transport <transport_id> <tunnel_id> <protocol> <port>
func BuildIPsecTransportCommand(t IPsecTransport) string {
	return fmt.Sprintf("ipsec transport %d %d %s %d", t.TransportID, t.TunnelID, t.Protocol, t.Port)
}

// BuildDeleteIPsecTransportCommand builds the command to delete an IPsec transport configuration
// Command format: no ipsec transport <transport_id>
func BuildDeleteIPsecTransportCommand(transportID int) string {
	return fmt.Sprintf("no ipsec transport %d", transportID)
}

// BuildShowIPsecTransportCommand builds the command to show IPsec transport configuration
func BuildShowIPsecTransportCommand() string {
	return "show config | grep \"ipsec transport\""
}

// ValidateIPsecTransport validates an IPsec transport configuration
func ValidateIPsecTransport(t IPsecTransport) error {
	if t.TransportID <= 0 {
		return fmt.Errorf("transport_id must be positive")
	}

	if t.TunnelID <= 0 {
		return fmt.Errorf("tunnel_id must be positive")
	}

	if t.Protocol == "" {
		return fmt.Errorf("protocol is required")
	}

	// Validate protocol is a known value
	validProtocols := []string{"udp", "tcp"}
	isValid := false
	for _, p := range validProtocols {
		if strings.EqualFold(t.Protocol, p) {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("protocol must be one of: %s", strings.Join(validProtocols, ", "))
	}

	if t.Port <= 0 || t.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	return nil
}
