package parsers

// Gateway represents a single gateway configuration for a route
type Gateway struct {
	IP        string `json:"ip,omitempty"`        // Gateway IP address (optional)
	Interface string `json:"interface,omitempty"` // Gateway interface name (optional)
	Weight    int    `json:"weight"`              // ECMP weight for this gateway
	Hide      bool   `json:"hide,omitempty"`      // Hide this gateway from route display
}

// StaticRoute represents a static route configuration on an RTX router
// This is a local type to avoid circular imports with the client package
type StaticRoute struct {
	Destination      string    `json:"destination"`                  // Destination network prefix (required)
	GatewayIP        string    `json:"gateway_ip,omitempty"`         // Next hop gateway IP address (deprecated, use Gateways)
	GatewayInterface string    `json:"gateway_interface,omitempty"`  // Next hop gateway interface name (deprecated, use Gateways)
	Gateways         []Gateway `json:"gateways,omitempty"`           // List of gateways for this route
	Interface        string    `json:"interface,omitempty"`          // Outgoing interface (optional)
	Metric           int       `json:"metric"`                       // Route metric (default: 1)
	Weight           int       `json:"weight"`                       // Route weight for ECMP (deprecated, use per-gateway weight)
	Description      string    `json:"description,omitempty"`        // Route description (optional)
	Hide             bool      `json:"hide,omitempty"`               // Hide flag (deprecated, use per-gateway hide)
}

// ToClientStaticRoute converts parser StaticRoute to client StaticRoute
func (r *StaticRoute) ToClientStaticRoute() interface{} {
	// This will be used by the client package to convert types
	return struct {
		Destination      string `json:"destination"`
		GatewayIP        string `json:"gateway_ip,omitempty"`
		GatewayInterface string `json:"gateway_interface,omitempty"`
		Interface        string `json:"interface,omitempty"`
		Metric           int    `json:"metric"`
		Weight           int    `json:"weight"`
		Description      string `json:"description,omitempty"`
		Hide             bool   `json:"hide,omitempty"`
	}{
		Destination:      r.Destination,
		GatewayIP:        r.GatewayIP,
		GatewayInterface: r.GatewayInterface,
		Interface:        r.Interface,
		Metric:           r.Metric,
		Weight:           r.Weight,
		Description:      r.Description,
		Hide:             r.Hide,
	}
}