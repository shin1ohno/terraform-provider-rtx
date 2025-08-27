package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
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
		Timeout:        time.Duration(c.config.Timeout) * time.Second,
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
	
	err := c.session.Close()
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
	
	// Validate host key configuration
	if config.HostKey != "" && config.KnownHostsFile != "" {
		// Both specified - HostKey takes priority, but we warn about it in logs if needed
	}
	
	return nil
}