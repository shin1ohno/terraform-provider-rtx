package parsers

import (
	"regexp"
	"strconv"
	"strings"
)

// Route represents a routing table entry on an RTX router
type Route struct {
	Destination string `json:"destination"`         // Network prefix (e.g., "192.168.1.0/24", "0.0.0.0/0")
	Gateway     string `json:"gateway"`            // Next hop gateway ("*" for directly connected routes)
	Interface   string `json:"interface"`          // Outgoing interface
	Protocol    string `json:"protocol"`           // S=static, C=connected, R=RIP, O=OSPF, B=BGP, D=DHCP
	Metric      *int   `json:"metric,omitempty"`   // Route metric (optional)
}

// RoutesParser is the interface for parsing routing table information
type RoutesParser interface {
	Parser
	ParseRoutes(raw string) ([]Route, error)
}

// BaseRoutesParser provides common functionality for route parsers
type BaseRoutesParser struct {
	modelPatterns map[string]*regexp.Regexp
}

// rtx830RoutesParser handles RTX830 route output
type rtx830RoutesParser struct {
	BaseRoutesParser
}

// rtx12xxRoutesParser handles RTX1210/1220 route output
type rtx12xxRoutesParser struct {
	BaseRoutesParser
}

func init() {
	// Register RTX830 routes parser
	Register("routes", "RTX830", &rtx830RoutesParser{
		BaseRoutesParser: BaseRoutesParser{
			modelPatterns: map[string]*regexp.Regexp{
				// RTX830 route format: "S   0.0.0.0/0         via 192.168.1.1    dev LAN1 metric 1"
				"route": regexp.MustCompile(`^([SCROBPD])\s+(\S+)\s+(?:via\s+(\S+))?\s*(?:dev\s+(\S+))?\s*(?:metric\s+(\d+))?`),
			},
		},
	})
	
	// Register RTX12xx routes parser
	rtx12xxParser := &rtx12xxRoutesParser{
		BaseRoutesParser: BaseRoutesParser{
			modelPatterns: map[string]*regexp.Regexp{
				// RTX12xx route format: "Destination     Gateway         Interface   Protocol Metric"
				//                        "0.0.0.0/0      192.168.1.1     LAN1        S        1"
				//                        "192.168.1.0/24 *               LAN1        C        -"
				"header": regexp.MustCompile(`^Destination\s+Gateway\s+Interface\s+Protocol\s+Metric`),
				"route":  regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+|-)$`),
			},
		},
	}
	Register("routes", "RTX1210", rtx12xxParser)
	Register("routes", "RTX1220", rtx12xxParser)
	
	// Create aliases for model families
	RegisterAlias("routes", "RTX1210", "RTX12xx")
}

// Parse implements the Parser interface
func (p *rtx830RoutesParser) Parse(raw string) (interface{}, error) {
	return p.ParseRoutes(raw)
}

// CanHandle implements the Parser interface
func (p *rtx830RoutesParser) CanHandle(model string) bool {
	return model == "RTX830"
}

// ParseRoutes parses RTX830 route output
func (p *rtx830RoutesParser) ParseRoutes(raw string) ([]Route, error) {
	routes := make([]Route, 0)
	lines := strings.Split(raw, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Try to match route pattern
		if match := p.modelPatterns["route"].FindStringSubmatch(line); len(match) >= 3 {
			route := Route{
				Protocol:    match[1],
				Destination: match[2],
			}
			
			// Gateway (match[3]) - may be empty for connected routes
			if len(match) > 3 && match[3] != "" {
				route.Gateway = match[3]
			} else {
				route.Gateway = "*" // Connected route
			}
			
			// Interface (match[4])
			if len(match) > 4 && match[4] != "" {
				route.Interface = match[4]
			}
			
			// Metric (match[5])
			if len(match) > 5 && match[5] != "" {
				if metric, err := strconv.Atoi(match[5]); err == nil {
					route.Metric = &metric
				}
			}
			
			routes = append(routes, route)
		}
	}
	
	return routes, nil
}

// Parse implements the Parser interface
func (p *rtx12xxRoutesParser) Parse(raw string) (interface{}, error) {
	return p.ParseRoutes(raw)
}

// CanHandle implements the Parser interface
func (p *rtx12xxRoutesParser) CanHandle(model string) bool {
	return strings.HasPrefix(model, "RTX12")
}

// ParseRoutes parses RTX1210/1220 route output
func (p *rtx12xxRoutesParser) ParseRoutes(raw string) ([]Route, error) {
	routes := make([]Route, 0)
	lines := strings.Split(raw, "\n")
	
	headerFound := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Check for header line
		if p.modelPatterns["header"].MatchString(line) {
			headerFound = true
			continue
		}
		
		// Skip until we find the header
		if !headerFound {
			continue
		}
		
		// Try to match route pattern
		if match := p.modelPatterns["route"].FindStringSubmatch(line); len(match) >= 6 {
			route := Route{
				Destination: match[1],
				Gateway:     match[2],
				Interface:   match[3],
				Protocol:    match[4],
			}
			
			// Handle metric (match[5])
			if match[5] != "-" && match[5] != "" {
				if metric, err := strconv.Atoi(match[5]); err == nil {
					route.Metric = &metric
				}
			}
			
			routes = append(routes, route)
		}
	}
	
	return routes, nil
}