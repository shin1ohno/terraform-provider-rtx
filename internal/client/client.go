package client

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
	"golang.org/x/crypto/ssh"
)

// rtxClient is the concrete implementation of the Client interface
type rtxClient struct {
	config         *Config
	dialer         ConnDialer
	promptDetector PromptDetector
	parsers        map[string]Parser
	retryStrategy  RetryStrategy

	mu          sync.Mutex
	session     Session
	executor    Executor
	active      bool
	dhcpService *DHCPService
}

// NewClient creates a new RTX client instance
func NewClient(config *Config, opts ...Option) (Client, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	c := &rtxClient{
		config:         config,
		dialer:         &sshDialer{},
		promptDetector: &defaultPromptDetector{},
		parsers:        make(map[string]Parser),
		retryStrategy:  &noRetry{},
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// Option is a functional option for configuring the client
type Option func(*rtxClient)

// WithDialer sets a custom connection dialer
func WithDialer(dialer ConnDialer) Option {
	return func(c *rtxClient) {
		c.dialer = dialer
	}
}

// WithPromptDetector sets a custom prompt detector
func WithPromptDetector(detector PromptDetector) Option {
	return func(c *rtxClient) {
		c.promptDetector = detector
	}
}

// WithParser registers a parser for a specific command
func WithParser(cmdKey string, parser Parser) Option {
	return func(c *rtxClient) {
		c.parsers[cmdKey] = parser
	}
}

// WithRetryStrategy sets the retry strategy for transient failures
func WithRetryStrategy(strategy RetryStrategy) Option {
	return func(c *rtxClient) {
		c.retryStrategy = strategy
	}
}

// getHostKeyCallback returns the appropriate host key callback based on configuration
func (c *rtxClient) getHostKeyCallback() ssh.HostKeyCallback {
	if c.config.SkipHostKeyCheck {
		return ssh.InsecureIgnoreHostKey()
	}

	// Implement proper host key checking if needed
	// For now, we'll use InsecureIgnoreHostKey
	return ssh.InsecureIgnoreHostKey()
}

// Dial establishes a connection to the RTX router
func (c *rtxClient) Dial(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.active {
		return nil // Already connected
	}

	// For RTX routers, we'll use a simple executor that creates new connections per command
	// This is less efficient but more reliable given RTX's SSH implementation
	sshConfig := &ssh.ClientConfig{
		User: c.config.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(c.config.Password),
		},
		HostKeyCallback: c.getHostKeyCallback(),
		Timeout:         time.Duration(c.config.Timeout) * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	c.executor = NewSimpleExecutor(sshConfig, addr, c.promptDetector, c.config)
	c.dhcpService = NewDHCPService(c.executor, c)
	c.active = true
	return nil
}

// Close terminates the connection
func (c *rtxClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.active {
		return nil
	}

	var err error
	if c.session != nil {
		err = c.session.Close()
	}

	c.active = false
	c.session = nil
	c.executor = nil
	c.dhcpService = nil

	if err != nil {
		return fmt.Errorf("failed to close session: %w", err)
	}
	return nil
}

// Run executes a command and returns the result
func (c *rtxClient) Run(ctx context.Context, cmd Command) (Result, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return Result{}, fmt.Errorf("client not connected")
	}
	executor := c.executor
	c.mu.Unlock()

	// Execute command via executor
	raw, err := executor.Run(ctx, cmd.Payload)
	if err != nil {
		return Result{}, err
	}

	result := Result{Raw: raw}

	// Parse if parser is available
	if parser, ok := c.parsers[cmd.Key]; ok {
		parsed, err := parser.Parse(raw)
		if err != nil {
			return result, fmt.Errorf("%w: %v", ErrParse, err)
		}
		result.Parsed = parsed
	}

	return result, nil
}

// GetInterfaces retrieves interface information from the router
func (c *rtxClient) GetInterfaces(ctx context.Context) ([]Interface, error) {
	// First get system information to determine model
	systemInfo, err := c.GetSystemInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system info for parser selection: %w", err)
	}

	// Execute interface command based on model
	var cmdPayload string
	switch {
	case systemInfo.Model == "RTX830":
		cmdPayload = "show interface"
	default:
		cmdPayload = "show interface"
	}

	cmd := Command{
		Key:     "interfaces",
		Payload: cmdPayload,
	}

	result, err := c.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// Use the parser registry to parse interfaces
	parser, err := parsers.Get("interfaces", systemInfo.Model)
	if err != nil {
		return nil, fmt.Errorf("no parser available for model %s: %w", systemInfo.Model, err)
	}

	// Cast to InterfacesParser to access ParseInterfaces method
	interfacesParser, ok := parser.(parsers.InterfacesParser)
	if !ok {
		return nil, fmt.Errorf("parser for %s does not implement InterfacesParser", systemInfo.Model)
	}

	parsedInterfaces, err := interfacesParser.ParseInterfaces(string(result.Raw))
	if err != nil {
		return nil, fmt.Errorf("failed to parse interfaces: %w", err)
	}

	// Convert parsers.Interface to client.Interface
	interfaces := make([]Interface, len(parsedInterfaces))
	for i, pi := range parsedInterfaces {
		interfaces[i] = Interface{
			Name:        pi.Name,
			Kind:        pi.Kind,
			AdminUp:     pi.AdminUp,
			LinkUp:      pi.LinkUp,
			MAC:         pi.MAC,
			IPv4:        pi.IPv4,
			IPv6:        pi.IPv6,
			MTU:         pi.MTU,
			Description: pi.Description,
			Attributes:  pi.Attributes,
		}
	}

	return interfaces, nil
}

// GetRoutes retrieves routing table information from the router
func (c *rtxClient) GetRoutes(ctx context.Context) ([]Route, error) {
	// First get system information to determine model
	systemInfo, err := c.GetSystemInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system info for parser selection: %w", err)
	}

	// Execute route command based on model
	var cmdPayload string
	switch {
	case systemInfo.Model == "RTX830":
		cmdPayload = "show ip route"
	default:
		cmdPayload = "show ip route"
	}

	cmd := Command{
		Key:     "routes",
		Payload: cmdPayload,
	}

	result, err := c.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// Use the parser registry to parse routes
	parser, err := parsers.Get("routes", systemInfo.Model)
	if err != nil {
		return nil, fmt.Errorf("no parser available for model %s: %w", systemInfo.Model, err)
	}

	// Cast to RoutesParser to access ParseRoutes method
	routesParser, ok := parser.(parsers.RoutesParser)
	if !ok {
		return nil, fmt.Errorf("parser for %s does not implement RoutesParser", systemInfo.Model)
	}

	parsedRoutes, err := routesParser.ParseRoutes(string(result.Raw))
	if err != nil {
		return nil, fmt.Errorf("failed to parse routes: %w", err)
	}

	// Convert parsers.Route to client.Route
	routes := make([]Route, len(parsedRoutes))
	for i, pr := range parsedRoutes {
		routes[i] = Route{
			Destination: pr.Destination,
			Gateway:     pr.Gateway,
			Interface:   pr.Interface,
			Protocol:    pr.Protocol,
			Metric:      pr.Metric,
		}
	}

	return routes, nil
}


// GetDHCPScopes retrieves DHCP scope configurations from the router
func (c *rtxClient) GetDHCPScopes(ctx context.Context) ([]DHCPScope, error) {
	// First get system information to determine model
	systemInfo, err := c.GetSystemInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system info for parser selection: %w", err)
	}

	// Execute DHCP scope command based on model
	var cmdPayload string
	switch {
	case systemInfo.Model == "RTX830":
		cmdPayload = "show running-config | grep \"dhcp scope\""
	default:
		// RTX1210 and other RTX series use 'show config'
		cmdPayload = "show config | grep \"dhcp scope\""
	}

	cmd := Command{
		Key:     "dhcp_scope",
		Payload: cmdPayload,
	}

	result, err := c.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// Use the parser registry to parse DHCP scopes
	parser, err := parsers.Get("dhcp_scope", systemInfo.Model)
	if err != nil {
		return nil, fmt.Errorf("no parser available for model %s: %w", systemInfo.Model, err)
	}

	// Cast to DhcpScopeParser to access ParseDhcpScopes method
	dhcpScopeParser, ok := parser.(parsers.DhcpScopeParser)
	if !ok {
		return nil, fmt.Errorf("parser for %s does not implement DhcpScopeParser", systemInfo.Model)
	}

	parsedScopes, err := dhcpScopeParser.ParseDhcpScopes(string(result.Raw))
	if err != nil {
		return nil, fmt.Errorf("failed to parse DHCP scopes: %w", err)
	}

	// Convert parsers.DhcpScope to client.DHCPScope
	scopes := make([]DHCPScope, len(parsedScopes))
	for i, ps := range parsedScopes {
		scopes[i] = DHCPScope{
			ID:         ps.ID,
			RangeStart: ps.RangeStart,
			RangeEnd:   ps.RangeEnd,
			Prefix:     ps.Prefix,
			Gateway:    ps.Gateway,
			DNSServers: ps.DNSServers,
			Lease:      ps.Lease,
			DomainName: ps.DomainName,
		}
	}

	return scopes, nil
}

// GetDHCPScope retrieves a specific DHCP scope by ID
func (c *rtxClient) GetDHCPScope(ctx context.Context, scopeID int) (*DHCPScope, error) {
	if scopeID <= 0 || scopeID > 255 {
		return nil, fmt.Errorf("scope_id must be between 1 and 255")
	}

	scopes, err := c.GetDHCPScopes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve DHCP scopes: %w", err)
	}

	// Find the specific scope
	for _, scope := range scopes {
		if scope.ID == scopeID {
			return &scope, nil
		}
	}

	// Return ErrNotFound if scope doesn't exist
	return nil, ErrNotFound
}

// CreateDHCPScope creates a new DHCP scope
func (c *rtxClient) CreateDHCPScope(ctx context.Context, scope DHCPScope) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	dhcpService := c.dhcpService
	c.mu.Unlock()

	if dhcpService == nil {
		return fmt.Errorf("DHCP service not initialized")
	}

	return dhcpService.CreateScope(ctx, scope)
}

// UpdateDHCPScope updates an existing DHCP scope
func (c *rtxClient) UpdateDHCPScope(ctx context.Context, scope DHCPScope) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	dhcpService := c.dhcpService
	c.mu.Unlock()

	if dhcpService == nil {
		return fmt.Errorf("DHCP service not initialized")
	}

	return dhcpService.UpdateScope(ctx, scope)
}

// DeleteDHCPScope removes a DHCP scope
func (c *rtxClient) DeleteDHCPScope(ctx context.Context, scopeID int) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	dhcpService := c.dhcpService
	c.mu.Unlock()

	if dhcpService == nil {
		return fmt.Errorf("DHCP service not initialized")
	}

	return dhcpService.DeleteScope(ctx, scopeID)
}

// GetDHCPBindings retrieves DHCP bindings for a scope
func (c *rtxClient) GetDHCPBindings(ctx context.Context, scopeID int) ([]DHCPBinding, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	dhcpService := c.dhcpService
	c.mu.Unlock()

	if dhcpService == nil {
		return nil, fmt.Errorf("DHCP service not initialized")
	}

	return dhcpService.ListBindings(ctx, scopeID)
}

// CreateDHCPBinding creates a new DHCP binding
func (c *rtxClient) CreateDHCPBinding(ctx context.Context, binding DHCPBinding) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	dhcpService := c.dhcpService
	c.mu.Unlock()

	if dhcpService == nil {
		return fmt.Errorf("DHCP service not initialized")
	}

	return dhcpService.CreateBinding(ctx, binding)
}

// DeleteDHCPBinding removes a DHCP binding
func (c *rtxClient) DeleteDHCPBinding(ctx context.Context, scopeID int, ipAddress string) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	dhcpService := c.dhcpService
	c.mu.Unlock()

	if dhcpService == nil {
		return fmt.Errorf("DHCP service not initialized")
	}

	return dhcpService.DeleteBinding(ctx, scopeID, ipAddress)
}

// SaveConfig saves the current configuration to persistent memory
func (c *rtxClient) SaveConfig(ctx context.Context) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	executor := c.executor
	c.mu.Unlock()

	// Execute save command
	_, err := executor.Run(ctx, "save")
	if err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return nil
}

// Static Route management methods

// GetStaticRoutes retrieves all static route configurations from the router
func (c *rtxClient) GetStaticRoutes(ctx context.Context) ([]StaticRoute, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	executor := c.executor
	c.mu.Unlock()

	// Get system information to determine the correct command
	systemInfo, err := c.GetSystemInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system info for parser selection: %w", err)
	}

	// Execute command to get static route configuration
	var cmdPayload string
	switch {
	case systemInfo.Model == "RTX830":
		cmdPayload = "show running-config | grep \"ip route\""
	default:
		// RTX1210 and other RTX series use 'show config'
		cmdPayload = "show config | grep \"ip route\""
	}

	output, err := executor.Run(ctx, cmdPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to get static routes: %w", err)
	}

	// Debug: log the actual output we got from router
	if len(output) == 0 {
		fmt.Printf("DEBUG: Got empty output from command: %s\n", cmdPayload)
	} else {
		fmt.Printf("DEBUG: Got output from router (%d bytes): %q\n", len(output), output)
	}

	// Parse static routes using parser
	parserRoutes, err := parsers.ParseStaticRoutes([]byte(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse static routes: %w", err)
	}

	fmt.Printf("DEBUG: Parsed %d routes from output\n", len(parserRoutes))

	// Convert parser StaticRoute to client StaticRoute
	var routes []StaticRoute
	for _, parserRoute := range parserRoutes {
		clientRoute := StaticRoute{
			Destination:      parserRoute.Destination,
			GatewayIP:        parserRoute.GatewayIP,
			GatewayInterface: parserRoute.GatewayInterface,
			Interface:        parserRoute.Interface,
			Metric:           parserRoute.Metric,
			Weight:           parserRoute.Weight,
			Description:      parserRoute.Description,
			Hide:             parserRoute.Hide,
		}
		
		// Copy new Gateways array
		if len(parserRoute.Gateways) > 0 {
			clientRoute.Gateways = make([]Gateway, len(parserRoute.Gateways))
			for i, gw := range parserRoute.Gateways {
				clientRoute.Gateways[i] = Gateway{
					IP:        gw.IP,
					Interface: gw.Interface,
					Weight:    gw.Weight,
					Hide:      gw.Hide,
				}
			}
		}
		
		routes = append(routes, clientRoute)
	}

	return routes, nil
}

// CreateStaticRoute creates a static route on the RTX router
func (c *rtxClient) CreateStaticRoute(ctx context.Context, route StaticRoute) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	executor := c.executor
	c.mu.Unlock()

	// Convert client StaticRoute to parser StaticRoute for command generation
	parserRoute := parsers.StaticRoute{
		Destination: route.Destination,
		Interface:   route.Interface,
		Metric:      route.Metric,
		Description: route.Description,
	}

	// Handle new Gateways array or fallback to legacy fields
	if len(route.Gateways) > 0 {
		// Use new Gateways array
		parserRoute.Gateways = make([]parsers.Gateway, len(route.Gateways))
		for i, gw := range route.Gateways {
			parserRoute.Gateways[i] = parsers.Gateway{
				IP:        gw.IP,
				Interface: gw.Interface,
				Weight:    gw.Weight,
				Hide:      gw.Hide,
			}
		}
	} else {
		// Fallback to legacy single gateway fields for backward compatibility
		parserRoute.GatewayIP = route.GatewayIP
		parserRoute.GatewayInterface = route.GatewayInterface
		parserRoute.Weight = route.Weight
		parserRoute.Hide = route.Hide
	}

	// Validate the route first
	if err := parsers.ValidateStaticRoute(parserRoute); err != nil {
		return fmt.Errorf("invalid route configuration: %w", err)
	}

	// Generate and execute the command
	cmd := parsers.BuildStaticRouteCommand(parserRoute)
	_, err := executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to create static route: %w", err)
	}

	// Save configuration to make it persistent
	return c.SaveConfig(ctx)
}

// GetStaticRoute retrieves a specific static route by its key components
func (c *rtxClient) GetStaticRoute(ctx context.Context, destination, gateway, iface string) (*StaticRoute, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	c.mu.Unlock()

	// Get all static routes and find the matching one
	routes, err := c.GetStaticRoutes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get static routes: %w", err)
	}

	for _, route := range routes {
		// For new ID format (destination only), match by destination
		if route.Destination == destination {
			// If gateway and iface are empty (new format), return the route
			if gateway == "" && iface == "" {
				return &route, nil
			}
			
			// Legacy format: also check gateway and interface match
			gatewayMatches := false
			
			// Handle multi-gateway case: gateway parameter might be comma-separated
			if gateway != "" {
				gatewayList := strings.Split(gateway, ",")
				
				// Check gateway match - support both new and legacy formats
				if len(route.Gateways) > 0 {
					// For multi-gateway routes, check if all gateways from ID are present
					if len(gatewayList) > 1 {
						// Multi-gateway case: check if the route contains all expected gateways
						matchedGateways := 0
						for _, expectedGw := range gatewayList {
							for _, routeGw := range route.Gateways {
								if (routeGw.IP != "" && routeGw.IP == expectedGw) || 
								   (routeGw.Interface != "" && routeGw.Interface == expectedGw) {
									matchedGateways++
									break
								}
							}
						}
						gatewayMatches = (matchedGateways == len(gatewayList))
					} else {
						// Single gateway case: match any gateway in the list
						for _, gw := range route.Gateways {
							if (gw.IP != "" && gw.IP == gateway) || 
							   (gw.Interface != "" && gw.Interface == gateway) {
								gatewayMatches = true
								break
							}
						}
					}
				} else {
					// Check legacy format
					if route.GatewayIP != "" && route.GatewayIP == gateway {
						gatewayMatches = true
					} else if route.GatewayInterface != "" && route.GatewayInterface == gateway {
						gatewayMatches = true
					}
				}
			} else {
				gatewayMatches = true // No gateway specified means match any
			}
			
			// Check interface match (empty interface matches empty string)
			interfaceMatches := (iface == "" || route.Interface == iface)
			
			if gatewayMatches && interfaceMatches {
				return &route, nil
			}
		}
	}

	return nil, fmt.Errorf("static route not found: %s via %s interface %s", destination, gateway, iface)
}

// UpdateStaticRoute updates an existing static route
func (c *rtxClient) UpdateStaticRoute(ctx context.Context, route StaticRoute) error {
	// For static routes, we implement update as delete-then-create
	// since RTX doesn't support in-place updates for route parameters like metric/weight
	
	// First, delete the existing route
	var gateway string
	if len(route.Gateways) > 0 {
		// Use first gateway from new array format  
		if route.Gateways[0].IP != "" {
			gateway = route.Gateways[0].IP
		} else {
			gateway = route.Gateways[0].Interface
		}
	} else {
		// Fallback to legacy fields
		gateway = route.GatewayIP
		if gateway == "" {
			gateway = route.GatewayInterface
		}
	}
	
	err := c.DeleteStaticRoute(ctx, route.Destination, gateway, route.Interface)
	if err != nil {
		// If delete fails with "not found", that's okay - we'll just create
		if fmt.Sprintf("%v", err) != "static route not found" {
			return fmt.Errorf("failed to delete existing route for update: %w", err)
		}
	}

	// Then create the new route
	return c.CreateStaticRoute(ctx, route)
}

// DeleteStaticRoute removes a static route from the RTX router
func (c *rtxClient) DeleteStaticRoute(ctx context.Context, destination, gateway, iface string) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	executor := c.executor
	c.mu.Unlock()

	// Create a route object for command generation
	parserRoute := parsers.StaticRoute{
		Destination: destination,
		Interface:   iface,
	}

	// Set the appropriate gateway field
	if gateway != "" {
		// Use net.ParseIP to properly determine if gateway is an IP address
		if net.ParseIP(gateway) != nil {
			// It's an IP address
			parserRoute.GatewayIP = gateway
		} else {
			// It's an interface name
			parserRoute.GatewayInterface = gateway
		}
	}

	// Generate and execute the delete command
	cmd := parsers.BuildStaticRouteDeleteCommand(parserRoute)
	_, err := executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete static route: %w", err)
	}

	// Save configuration to make it persistent
	return c.SaveConfig(ctx)
}

// validateConfig checks if the configuration is valid
func validateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if config.Host == "" {
		return fmt.Errorf("host is required")
	}
	if config.Username == "" {
		return fmt.Errorf("username is required")
	}
	if config.Password == "" {
		return fmt.Errorf("password is required")
	}
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", config.Port)
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 // Default timeout
	}

	// Validate host key configuration - both can be specified
	// HostKey takes priority over KnownHostsFile when both are provided

	return nil
}
