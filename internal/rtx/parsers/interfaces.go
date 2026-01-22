package parsers

import (
	"fmt"
	"regexp"
	"strings"
)

// Interface represents a network interface on an RTX router
type Interface struct {
	Name        string            `json:"name"`
	Kind        string            `json:"kind"` // lan, wan, pp, vlan
	AdminUp     bool              `json:"admin_up"`
	LinkUp      bool              `json:"link_up"`
	MAC         string            `json:"mac,omitempty"`
	IPv4        string            `json:"ipv4,omitempty"`
	IPv6        string            `json:"ipv6,omitempty"`
	MTU         int               `json:"mtu,omitempty"`
	Description string            `json:"description,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"` // For model-specific fields
}

// InterfacesParser is the interface for parsing interface information
type InterfacesParser interface {
	Parser
	ParseInterfaces(raw string) ([]Interface, error)
}

// BaseInterfacesParser provides common functionality for interface parsers
type BaseInterfacesParser struct {
	modelPatterns map[string]*regexp.Regexp
}

// rtx830InterfacesParser handles RTX830 interface output
type rtx830InterfacesParser struct {
	BaseInterfacesParser
}

// rtx12xxInterfacesParser handles RTX1210/1220 interface output
type rtx12xxInterfacesParser struct {
	BaseInterfacesParser
}

func init() {
	// Register RTX830 parser
	Register("interfaces", "RTX830", &rtx830InterfacesParser{
		BaseInterfacesParser: BaseInterfacesParser{
			modelPatterns: map[string]*regexp.Regexp{
				"interface": regexp.MustCompile(`^(LAN\d+|WAN\d+|PP\d+|VLAN\d+(?:\.\d+)?)\s*:\s*(.*)$`),
				"ipv4":      regexp.MustCompile(`IP\s*[Aa]ddress\s*:\s*([\d.]+(?:/\d+)?)`),
				"mac":       regexp.MustCompile(`MAC\s*[Aa]ddress\s*:\s*([0-9A-Fa-f:]+)`),
				"status":    regexp.MustCompile(`(UP|DOWN|up|down)`),
			},
		},
	})

	// Register RTX12xx parser
	rtx12xxParser := &rtx12xxInterfacesParser{
		BaseInterfacesParser: BaseInterfacesParser{
			modelPatterns: map[string]*regexp.Regexp{
				"interface": regexp.MustCompile(`^Interface\s+(LAN\d+|WAN\d+|PP\d+|VLAN\d+(?:\.\d+)?)`),
				"ipv4":      regexp.MustCompile(`IPv4\s*:\s*([\d.]+(?:/\d+)?)`),
				"ipv6":      regexp.MustCompile(`IPv6\s*:\s*([0-9a-fA-F:]+(?:/\d+)?)`),
				"mac":       regexp.MustCompile(`Ethernet\s+address\s*:\s*([0-9A-Fa-f:]+)`),
				"status":    regexp.MustCompile(`Status\s*:\s*(up|down)`),
				"mtu":       regexp.MustCompile(`MTU\s*:\s*(\d+)`),
			},
		},
	}
	Register("interfaces", "RTX1210", rtx12xxParser)
	Register("interfaces", "RTX1220", rtx12xxParser)

	// Create aliases for model families
	_ = RegisterAlias("interfaces", "RTX1210", "RTX12xx")
}

// Parse implements the Parser interface
func (p *rtx830InterfacesParser) Parse(raw string) (interface{}, error) {
	return p.ParseInterfaces(raw)
}

// CanHandle implements the Parser interface
func (p *rtx830InterfacesParser) CanHandle(model string) bool {
	return model == "RTX830"
}

// ParseInterfaces parses RTX830 interface output
func (p *rtx830InterfacesParser) ParseInterfaces(raw string) ([]Interface, error) {
	var interfaces []Interface
	lines := strings.Split(raw, "\n")

	var currentInterface *Interface

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this is a new interface line
		if match := p.modelPatterns["interface"].FindStringSubmatch(line); len(match) > 1 {
			// Save previous interface if exists
			if currentInterface != nil {
				interfaces = append(interfaces, *currentInterface)
			}

			// Start new interface
			currentInterface = &Interface{
				Name:       match[1],
				Kind:       getInterfaceKind(match[1]),
				Attributes: make(map[string]string),
			}

			// Check status in the same line
			if p.modelPatterns["status"].MatchString(match[2]) {
				status := p.modelPatterns["status"].FindString(match[2])
				currentInterface.LinkUp = strings.ToLower(status) == "up"
				currentInterface.AdminUp = currentInterface.LinkUp // Simplified for now
			}
			continue
		}

		if currentInterface == nil {
			continue
		}

		// Parse IPv4 address
		if match := p.modelPatterns["ipv4"].FindStringSubmatch(line); len(match) > 1 {
			currentInterface.IPv4 = match[1]
		}

		// Parse MAC address
		if match := p.modelPatterns["mac"].FindStringSubmatch(line); len(match) > 1 {
			currentInterface.MAC = strings.ToUpper(match[1])
		}
	}

	// Don't forget the last interface
	if currentInterface != nil {
		interfaces = append(interfaces, *currentInterface)
	}

	return interfaces, nil
}

// Parse implements the Parser interface
func (p *rtx12xxInterfacesParser) Parse(raw string) (interface{}, error) {
	return p.ParseInterfaces(raw)
}

// CanHandle implements the Parser interface
func (p *rtx12xxInterfacesParser) CanHandle(model string) bool {
	return strings.HasPrefix(model, "RTX12")
}

// ParseInterfaces parses RTX1210/1220 interface output
func (p *rtx12xxInterfacesParser) ParseInterfaces(raw string) ([]Interface, error) {
	var interfaces []Interface
	lines := strings.Split(raw, "\n")

	var currentInterface *Interface

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this is a new interface line
		if match := p.modelPatterns["interface"].FindStringSubmatch(line); len(match) > 1 {
			// Save previous interface if exists
			if currentInterface != nil {
				interfaces = append(interfaces, *currentInterface)
			}

			// Start new interface
			currentInterface = &Interface{
				Name:       match[1],
				Kind:       getInterfaceKind(match[1]),
				Attributes: make(map[string]string),
			}
			continue
		}

		if currentInterface == nil {
			continue
		}

		// Parse status
		if match := p.modelPatterns["status"].FindStringSubmatch(line); len(match) > 1 {
			currentInterface.LinkUp = strings.ToLower(match[1]) == "up"
			currentInterface.AdminUp = currentInterface.LinkUp
		}

		// Parse IPv4 address
		if match := p.modelPatterns["ipv4"].FindStringSubmatch(line); len(match) > 1 {
			currentInterface.IPv4 = match[1]
		}

		// Parse IPv6 address
		if match := p.modelPatterns["ipv6"].FindStringSubmatch(line); len(match) > 1 {
			currentInterface.IPv6 = match[1]
		}

		// Parse MAC address
		if match := p.modelPatterns["mac"].FindStringSubmatch(line); len(match) > 1 {
			currentInterface.MAC = strings.ToUpper(match[1])
		}

		// Parse MTU
		if match := p.modelPatterns["mtu"].FindStringSubmatch(line); len(match) > 1 {
			_, _ = fmt.Sscanf(match[1], "%d", &currentInterface.MTU)
		}
	}

	// Don't forget the last interface
	if currentInterface != nil {
		interfaces = append(interfaces, *currentInterface)
	}

	return interfaces, nil
}

// getInterfaceKind determines the interface type from its name
func getInterfaceKind(name string) string {
	switch {
	case strings.HasPrefix(name, "LAN"):
		return "lan"
	case strings.HasPrefix(name, "WAN"):
		return "wan"
	case strings.HasPrefix(name, "PP"):
		return "pp"
	case strings.HasPrefix(name, "VLAN"):
		return "vlan"
	default:
		return "unknown"
	}
}
