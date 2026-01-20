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

	mu                    sync.Mutex
	session               Session
	executor              Executor
	active                bool
	dhcpService           *DHCPService
	dhcpScopeService      *DHCPScopeService
	ipv6PrefixService     *IPv6PrefixService
	systemService         *SystemService
	vlanService           *VLANService
	interfaceService      *InterfaceService
	staticRouteService    *StaticRouteService
	natMasqueradeService    *NATMasqueradeService
	natStaticService        *NATStaticService
	ethernetFilterService   *EthernetFilterService
	ipFilterService         *IPFilterService
	bgpService              *BGPService
	ospfService             *OSPFService
	ipsecTunnelService      *IPsecTunnelService
	l2tpService             *L2TPService
	pptpService             *PPTPService
	syslogService           *SyslogService
	snmpService             *SNMPService
	qosService              *QoSService
	scheduleService         *ScheduleService
	dnsService              *DNSService
	adminService            *AdminService
	serviceManager          *ServiceManager
	bridgeService           *BridgeService
	ipv6InterfaceService    *IPv6InterfaceService
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

	// Use dialer if provided (for testing/dependency injection)
	if c.dialer != nil {
		session, err := c.dialer.Dial(ctx, fmt.Sprintf("%s:%d", c.config.Host, c.config.Port), c.config)
		if err != nil {
			return err
		}
		c.session = session
	}
	c.executor = NewSimpleExecutor(sshConfig, addr, c.promptDetector, c.config)
	c.dhcpService = NewDHCPService(c.executor, c)
	c.dhcpScopeService = NewDHCPScopeService(c.executor, c)
	c.ipv6PrefixService = NewIPv6PrefixService(c.executor, c)
	c.systemService = NewSystemService(c.executor, c)
	c.vlanService = NewVLANService(c.executor, c)
	c.interfaceService = NewInterfaceService(c.executor, c)
	c.staticRouteService = NewStaticRouteService(c.executor, c)
	c.natMasqueradeService = NewNATMasqueradeService(c.executor, c)
	c.natStaticService = NewNATStaticService(c.executor, c)
	c.ethernetFilterService = NewEthernetFilterService(c.executor, c)
	c.ipFilterService = NewIPFilterService(c.executor, c)
	c.bgpService = NewBGPService(c.executor, c)
	c.ospfService = NewOSPFService(c.executor, c)
	c.ipsecTunnelService = NewIPsecTunnelService(c.executor, c)
	c.l2tpService = NewL2TPService(c.executor, c)
	c.pptpService = NewPPTPService(c.executor, c)
	c.syslogService = NewSyslogService(c.executor, c)
	c.snmpService = NewSNMPService(c.executor, c)
	c.qosService = NewQoSService(c.executor, c)
	c.scheduleService = NewScheduleService(c.executor, c)
	c.dnsService = NewDNSService(c.executor, c)
	c.adminService = NewAdminService(c.executor, c)
	c.serviceManager = NewServiceManager(c.executor, c)
	c.bridgeService = NewBridgeService(c.executor, c)
	c.ipv6InterfaceService = NewIPv6InterfaceService(c.executor, c)
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
	c.dhcpScopeService = nil
	c.ipv6PrefixService = nil
	c.systemService = nil
	c.vlanService = nil
	c.interfaceService = nil
	c.staticRouteService = nil
	c.natMasqueradeService = nil
	c.natStaticService = nil
	c.ethernetFilterService = nil
	c.ipFilterService = nil
	c.bgpService = nil
	c.ospfService = nil
	c.ipsecTunnelService = nil
	c.l2tpService = nil
	c.pptpService = nil
	c.syslogService = nil
	c.snmpService = nil
	c.qosService = nil
	c.scheduleService = nil
	c.dnsService = nil
	c.adminService = nil
	c.serviceManager = nil
	c.bridgeService = nil
	c.ipv6InterfaceService = nil

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
	session := c.session
	executor := c.executor
	c.mu.Unlock()

	var raw []byte
	var err error

	// Use session if available (for testing/dependency injection), otherwise use executor
	if session != nil {
		raw, err = session.Send(cmd.Payload)
	} else if executor != nil {
		raw, err = executor.Run(ctx, cmd.Payload)
	} else {
		return Result{}, fmt.Errorf("no session or executor available")
	}

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

// GetDHCPScope retrieves a DHCP scope configuration
func (c *rtxClient) GetDHCPScope(ctx context.Context, scopeID int) (*DHCPScope, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	dhcpScopeService := c.dhcpScopeService
	c.mu.Unlock()

	if dhcpScopeService == nil {
		return nil, fmt.Errorf("DHCP scope service not initialized")
	}

	return dhcpScopeService.GetScope(ctx, scopeID)
}

// CreateDHCPScope creates a new DHCP scope
func (c *rtxClient) CreateDHCPScope(ctx context.Context, scope DHCPScope) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	dhcpScopeService := c.dhcpScopeService
	c.mu.Unlock()

	if dhcpScopeService == nil {
		return fmt.Errorf("DHCP scope service not initialized")
	}

	return dhcpScopeService.CreateScope(ctx, scope)
}

// UpdateDHCPScope updates an existing DHCP scope
func (c *rtxClient) UpdateDHCPScope(ctx context.Context, scope DHCPScope) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	dhcpScopeService := c.dhcpScopeService
	c.mu.Unlock()

	if dhcpScopeService == nil {
		return fmt.Errorf("DHCP scope service not initialized")
	}

	return dhcpScopeService.UpdateScope(ctx, scope)
}

// DeleteDHCPScope removes a DHCP scope
func (c *rtxClient) DeleteDHCPScope(ctx context.Context, scopeID int) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	dhcpScopeService := c.dhcpScopeService
	c.mu.Unlock()

	if dhcpScopeService == nil {
		return fmt.Errorf("DHCP scope service not initialized")
	}

	return dhcpScopeService.DeleteScope(ctx, scopeID)
}

// ListDHCPScopes retrieves all DHCP scopes
func (c *rtxClient) ListDHCPScopes(ctx context.Context) ([]DHCPScope, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	dhcpScopeService := c.dhcpScopeService
	c.mu.Unlock()

	if dhcpScopeService == nil {
		return nil, fmt.Errorf("DHCP scope service not initialized")
	}

	return dhcpScopeService.ListScopes(ctx)
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

// GetIPv6Prefix retrieves an IPv6 prefix configuration
func (c *rtxClient) GetIPv6Prefix(ctx context.Context, prefixID int) (*IPv6Prefix, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	ipv6PrefixService := c.ipv6PrefixService
	c.mu.Unlock()

	if ipv6PrefixService == nil {
		return nil, fmt.Errorf("IPv6 prefix service not initialized")
	}

	return ipv6PrefixService.GetPrefix(ctx, prefixID)
}

// CreateIPv6Prefix creates a new IPv6 prefix
func (c *rtxClient) CreateIPv6Prefix(ctx context.Context, prefix IPv6Prefix) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ipv6PrefixService := c.ipv6PrefixService
	c.mu.Unlock()

	if ipv6PrefixService == nil {
		return fmt.Errorf("IPv6 prefix service not initialized")
	}

	return ipv6PrefixService.CreatePrefix(ctx, prefix)
}

// UpdateIPv6Prefix updates an existing IPv6 prefix
func (c *rtxClient) UpdateIPv6Prefix(ctx context.Context, prefix IPv6Prefix) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ipv6PrefixService := c.ipv6PrefixService
	c.mu.Unlock()

	if ipv6PrefixService == nil {
		return fmt.Errorf("IPv6 prefix service not initialized")
	}

	return ipv6PrefixService.UpdatePrefix(ctx, prefix)
}

// DeleteIPv6Prefix removes an IPv6 prefix
func (c *rtxClient) DeleteIPv6Prefix(ctx context.Context, prefixID int) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ipv6PrefixService := c.ipv6PrefixService
	c.mu.Unlock()

	if ipv6PrefixService == nil {
		return fmt.Errorf("IPv6 prefix service not initialized")
	}

	return ipv6PrefixService.DeletePrefix(ctx, prefixID)
}

// ListIPv6Prefixes retrieves all IPv6 prefixes
func (c *rtxClient) ListIPv6Prefixes(ctx context.Context) ([]IPv6Prefix, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	ipv6PrefixService := c.ipv6PrefixService
	c.mu.Unlock()

	if ipv6PrefixService == nil {
		return nil, fmt.Errorf("IPv6 prefix service not initialized")
	}

	return ipv6PrefixService.ListPrefixes(ctx)
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
// GetSystemConfig retrieves system configuration
func (c *rtxClient) GetSystemConfig(ctx context.Context) (*SystemConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	systemService := c.systemService
	c.mu.Unlock()

	if systemService == nil {
		return nil, fmt.Errorf("system service not initialized")
	}

	return systemService.Get(ctx)
}

// ConfigureSystem sets system configuration
func (c *rtxClient) ConfigureSystem(ctx context.Context, config SystemConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	systemService := c.systemService
	c.mu.Unlock()

	if systemService == nil {
		return fmt.Errorf("system service not initialized")
	}

	return systemService.Configure(ctx, config)
}

// UpdateSystemConfig updates system configuration
func (c *rtxClient) UpdateSystemConfig(ctx context.Context, config SystemConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	systemService := c.systemService
	c.mu.Unlock()

	if systemService == nil {
		return fmt.Errorf("system service not initialized")
	}

	return systemService.Update(ctx, config)
}

// ResetSystem resets system configuration to defaults
func (c *rtxClient) ResetSystem(ctx context.Context) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	systemService := c.systemService
	c.mu.Unlock()

	if systemService == nil {
		return fmt.Errorf("system service not initialized")
	}

	return systemService.Reset(ctx)
}

// GetVLAN retrieves a VLAN configuration
func (c *rtxClient) GetVLAN(ctx context.Context, iface string, vlanID int) (*VLAN, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	vlanService := c.vlanService
	c.mu.Unlock()

	if vlanService == nil {
		return nil, fmt.Errorf("VLAN service not initialized")
	}

	return vlanService.GetVLAN(ctx, iface, vlanID)
}

// CreateVLAN creates a new VLAN
func (c *rtxClient) CreateVLAN(ctx context.Context, vlan VLAN) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	vlanService := c.vlanService
	c.mu.Unlock()

	if vlanService == nil {
		return fmt.Errorf("VLAN service not initialized")
	}

	return vlanService.CreateVLAN(ctx, vlan)
}

// UpdateVLAN updates an existing VLAN
func (c *rtxClient) UpdateVLAN(ctx context.Context, vlan VLAN) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	vlanService := c.vlanService
	c.mu.Unlock()

	if vlanService == nil {
		return fmt.Errorf("VLAN service not initialized")
	}

	return vlanService.UpdateVLAN(ctx, vlan)
}

// DeleteVLAN removes a VLAN
func (c *rtxClient) DeleteVLAN(ctx context.Context, iface string, vlanID int) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	vlanService := c.vlanService
	c.mu.Unlock()

	if vlanService == nil {
		return fmt.Errorf("VLAN service not initialized")
	}

	return vlanService.DeleteVLAN(ctx, iface, vlanID)
}

// ListVLANs retrieves all VLANs
func (c *rtxClient) ListVLANs(ctx context.Context) ([]VLAN, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	vlanService := c.vlanService
	c.mu.Unlock()

	if vlanService == nil {
		return nil, fmt.Errorf("VLAN service not initialized")
	}

	return vlanService.ListVLANs(ctx)
}

// GetInterfaceConfig retrieves an interface configuration
func (c *rtxClient) GetInterfaceConfig(ctx context.Context, interfaceName string) (*InterfaceConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	interfaceService := c.interfaceService
	c.mu.Unlock()

	if interfaceService == nil {
		return nil, fmt.Errorf("interface service not initialized")
	}

	return interfaceService.Get(ctx, interfaceName)
}

// ConfigureInterface creates a new interface configuration
func (c *rtxClient) ConfigureInterface(ctx context.Context, config InterfaceConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	interfaceService := c.interfaceService
	c.mu.Unlock()

	if interfaceService == nil {
		return fmt.Errorf("interface service not initialized")
	}

	return interfaceService.Configure(ctx, config)
}

// UpdateInterfaceConfig updates an existing interface configuration
func (c *rtxClient) UpdateInterfaceConfig(ctx context.Context, config InterfaceConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	interfaceService := c.interfaceService
	c.mu.Unlock()

	if interfaceService == nil {
		return fmt.Errorf("interface service not initialized")
	}

	return interfaceService.Update(ctx, config)
}

// ResetInterface removes interface configuration (resets to defaults)
func (c *rtxClient) ResetInterface(ctx context.Context, interfaceName string) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	interfaceService := c.interfaceService
	c.mu.Unlock()

	if interfaceService == nil {
		return fmt.Errorf("interface service not initialized")
	}

	return interfaceService.Reset(ctx, interfaceName)
}

// ListInterfaceConfigs retrieves all interface configurations
func (c *rtxClient) ListInterfaceConfigs(ctx context.Context) ([]InterfaceConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	interfaceService := c.interfaceService
	c.mu.Unlock()

	if interfaceService == nil {
		return nil, fmt.Errorf("interface service not initialized")
	}

	return interfaceService.List(ctx)
}

// GetStaticRoute retrieves a static route configuration
func (c *rtxClient) GetStaticRoute(ctx context.Context, prefix, mask string) (*StaticRoute, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	staticRouteService := c.staticRouteService
	c.mu.Unlock()

	if staticRouteService == nil {
		return nil, fmt.Errorf("static route service not initialized")
	}

	return staticRouteService.GetRoute(ctx, prefix, mask)
}

// CreateStaticRoute creates a new static route
func (c *rtxClient) CreateStaticRoute(ctx context.Context, route StaticRoute) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	staticRouteService := c.staticRouteService
	c.mu.Unlock()

	if staticRouteService == nil {
		return fmt.Errorf("static route service not initialized")
	}

	return staticRouteService.CreateRoute(ctx, route)
}

// UpdateStaticRoute updates an existing static route
func (c *rtxClient) UpdateStaticRoute(ctx context.Context, route StaticRoute) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	staticRouteService := c.staticRouteService
	c.mu.Unlock()

	if staticRouteService == nil {
		return fmt.Errorf("static route service not initialized")
	}

	return staticRouteService.UpdateRoute(ctx, route)
}

// DeleteStaticRoute removes a static route
func (c *rtxClient) DeleteStaticRoute(ctx context.Context, prefix, mask string) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	staticRouteService := c.staticRouteService
	c.mu.Unlock()

	if staticRouteService == nil {
		return fmt.Errorf("static route service not initialized")
	}

	return staticRouteService.DeleteRoute(ctx, prefix, mask)
}

// ListStaticRoutes retrieves all static routes
func (c *rtxClient) ListStaticRoutes(ctx context.Context) ([]StaticRoute, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	staticRouteService := c.staticRouteService
	c.mu.Unlock()

	if staticRouteService == nil {
		return nil, fmt.Errorf("static route service not initialized")
	}

	return staticRouteService.ListRoutes(ctx)
}

// GetNATMasquerade retrieves a NAT masquerade configuration
func (c *rtxClient) GetNATMasquerade(ctx context.Context, descriptorID int) (*NATMasquerade, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	natMasqueradeService := c.natMasqueradeService
	c.mu.Unlock()

	if natMasqueradeService == nil {
		return nil, fmt.Errorf("NAT masquerade service not initialized")
	}

	return natMasqueradeService.Get(ctx, descriptorID)
}

// CreateNATMasquerade creates a new NAT masquerade
func (c *rtxClient) CreateNATMasquerade(ctx context.Context, nat NATMasquerade) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	natMasqueradeService := c.natMasqueradeService
	c.mu.Unlock()

	if natMasqueradeService == nil {
		return fmt.Errorf("NAT masquerade service not initialized")
	}

	return natMasqueradeService.Create(ctx, nat)
}

// UpdateNATMasquerade updates an existing NAT masquerade
func (c *rtxClient) UpdateNATMasquerade(ctx context.Context, nat NATMasquerade) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	natMasqueradeService := c.natMasqueradeService
	c.mu.Unlock()

	if natMasqueradeService == nil {
		return fmt.Errorf("NAT masquerade service not initialized")
	}

	return natMasqueradeService.Update(ctx, nat)
}

// DeleteNATMasquerade removes a NAT masquerade
func (c *rtxClient) DeleteNATMasquerade(ctx context.Context, descriptorID int) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	natMasqueradeService := c.natMasqueradeService
	c.mu.Unlock()

	if natMasqueradeService == nil {
		return fmt.Errorf("NAT masquerade service not initialized")
	}

	return natMasqueradeService.Delete(ctx, descriptorID)
}

// ListNATMasquerades retrieves all NAT masquerades
func (c *rtxClient) ListNATMasquerades(ctx context.Context) ([]NATMasquerade, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	natMasqueradeService := c.natMasqueradeService
	c.mu.Unlock()

	if natMasqueradeService == nil {
		return nil, fmt.Errorf("NAT masquerade service not initialized")
	}

	return natMasqueradeService.List(ctx)
}

// GetNATStatic retrieves a NAT static configuration
func (c *rtxClient) GetNATStatic(ctx context.Context, descriptorID int) (*NATStatic, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	natStaticService := c.natStaticService
	c.mu.Unlock()

	if natStaticService == nil {
		return nil, fmt.Errorf("NAT static service not initialized")
	}

	return natStaticService.Get(ctx, descriptorID)
}

// CreateNATStatic creates a new NAT static
func (c *rtxClient) CreateNATStatic(ctx context.Context, nat NATStatic) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	natStaticService := c.natStaticService
	c.mu.Unlock()

	if natStaticService == nil {
		return fmt.Errorf("NAT static service not initialized")
	}

	return natStaticService.Create(ctx, nat)
}

// UpdateNATStatic updates an existing NAT static
func (c *rtxClient) UpdateNATStatic(ctx context.Context, nat NATStatic) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	natStaticService := c.natStaticService
	c.mu.Unlock()

	if natStaticService == nil {
		return fmt.Errorf("NAT static service not initialized")
	}

	return natStaticService.Update(ctx, nat)
}

// DeleteNATStatic removes a NAT static
func (c *rtxClient) DeleteNATStatic(ctx context.Context, descriptorID int) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	natStaticService := c.natStaticService
	c.mu.Unlock()

	if natStaticService == nil {
		return fmt.Errorf("NAT static service not initialized")
	}

	return natStaticService.Delete(ctx, descriptorID)
}

// ListNATStatics retrieves all NAT statics
func (c *rtxClient) ListNATStatics(ctx context.Context) ([]NATStatic, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	natStaticService := c.natStaticService
	c.mu.Unlock()

	if natStaticService == nil {
		return nil, fmt.Errorf("NAT static service not initialized")
	}

	return natStaticService.List(ctx)
}

// GetEthernetFilter retrieves an Ethernet filter configuration
func (c *rtxClient) GetEthernetFilter(ctx context.Context, number int) (*EthernetFilter, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	ethernetFilterService := c.ethernetFilterService
	c.mu.Unlock()

	if ethernetFilterService == nil {
		return nil, fmt.Errorf("Ethernet filter service not initialized")
	}

	return ethernetFilterService.GetFilter(ctx, number)
}

// CreateEthernetFilter creates a new Ethernet filter
func (c *rtxClient) CreateEthernetFilter(ctx context.Context, filter EthernetFilter) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ethernetFilterService := c.ethernetFilterService
	c.mu.Unlock()

	if ethernetFilterService == nil {
		return fmt.Errorf("Ethernet filter service not initialized")
	}

	return ethernetFilterService.CreateFilter(ctx, filter)
}

// UpdateEthernetFilter updates an existing Ethernet filter
func (c *rtxClient) UpdateEthernetFilter(ctx context.Context, filter EthernetFilter) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ethernetFilterService := c.ethernetFilterService
	c.mu.Unlock()

	if ethernetFilterService == nil {
		return fmt.Errorf("Ethernet filter service not initialized")
	}

	return ethernetFilterService.UpdateFilter(ctx, filter)
}

// DeleteEthernetFilter removes an Ethernet filter
func (c *rtxClient) DeleteEthernetFilter(ctx context.Context, number int) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ethernetFilterService := c.ethernetFilterService
	c.mu.Unlock()

	if ethernetFilterService == nil {
		return fmt.Errorf("Ethernet filter service not initialized")
	}

	return ethernetFilterService.DeleteFilter(ctx, number)
}

// ListEthernetFilters retrieves all Ethernet filters
func (c *rtxClient) ListEthernetFilters(ctx context.Context) ([]EthernetFilter, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	ethernetFilterService := c.ethernetFilterService
	c.mu.Unlock()

	if ethernetFilterService == nil {
		return nil, fmt.Errorf("Ethernet filter service not initialized")
	}

	return ethernetFilterService.ListFilters(ctx)
}

// GetIPFilter retrieves an IP filter configuration
func (c *rtxClient) GetIPFilter(ctx context.Context, number int) (*IPFilter, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	ipFilterService := c.ipFilterService
	c.mu.Unlock()

	if ipFilterService == nil {
		return nil, fmt.Errorf("IP filter service not initialized")
	}

	return ipFilterService.GetFilter(ctx, number)
}

// CreateIPFilter creates a new IP filter
func (c *rtxClient) CreateIPFilter(ctx context.Context, filter IPFilter) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ipFilterService := c.ipFilterService
	c.mu.Unlock()

	if ipFilterService == nil {
		return fmt.Errorf("IP filter service not initialized")
	}

	return ipFilterService.CreateFilter(ctx, filter)
}

// UpdateIPFilter updates an existing IP filter
func (c *rtxClient) UpdateIPFilter(ctx context.Context, filter IPFilter) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ipFilterService := c.ipFilterService
	c.mu.Unlock()

	if ipFilterService == nil {
		return fmt.Errorf("IP filter service not initialized")
	}

	return ipFilterService.UpdateFilter(ctx, filter)
}

// DeleteIPFilter removes an IP filter
func (c *rtxClient) DeleteIPFilter(ctx context.Context, number int) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ipFilterService := c.ipFilterService
	c.mu.Unlock()

	if ipFilterService == nil {
		return fmt.Errorf("IP filter service not initialized")
	}

	return ipFilterService.DeleteFilter(ctx, number)
}

// ListIPFilters retrieves all IP filters
func (c *rtxClient) ListIPFilters(ctx context.Context) ([]IPFilter, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	ipFilterService := c.ipFilterService
	c.mu.Unlock()

	if ipFilterService == nil {
		return nil, fmt.Errorf("IP filter service not initialized")
	}

	return ipFilterService.ListFilters(ctx)
}

// GetIPFilterDynamic retrieves a dynamic IP filter configuration
func (c *rtxClient) GetIPFilterDynamic(ctx context.Context, number int) (*IPFilterDynamic, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	ipFilterService := c.ipFilterService
	c.mu.Unlock()

	if ipFilterService == nil {
		return nil, fmt.Errorf("IP filter service not initialized")
	}

	return ipFilterService.GetDynamicFilter(ctx, number)
}

// CreateIPFilterDynamic creates a new dynamic IP filter
func (c *rtxClient) CreateIPFilterDynamic(ctx context.Context, filter IPFilterDynamic) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ipFilterService := c.ipFilterService
	c.mu.Unlock()

	if ipFilterService == nil {
		return fmt.Errorf("IP filter service not initialized")
	}

	return ipFilterService.CreateDynamicFilter(ctx, filter)
}

// DeleteIPFilterDynamic removes a dynamic IP filter
func (c *rtxClient) DeleteIPFilterDynamic(ctx context.Context, number int) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ipFilterService := c.ipFilterService
	c.mu.Unlock()

	if ipFilterService == nil {
		return fmt.Errorf("IP filter service not initialized")
	}

	return ipFilterService.DeleteDynamicFilter(ctx, number)
}

// ListIPFiltersDynamic retrieves all dynamic IP filters
func (c *rtxClient) ListIPFiltersDynamic(ctx context.Context) ([]IPFilterDynamic, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	ipFilterService := c.ipFilterService
	c.mu.Unlock()

	if ipFilterService == nil {
		return nil, fmt.Errorf("IP filter service not initialized")
	}

	return ipFilterService.ListDynamicFilters(ctx)
}

// GetBGPConfig retrieves BGP configuration
func (c *rtxClient) GetBGPConfig(ctx context.Context) (*BGPConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	bgpService := c.bgpService
	c.mu.Unlock()

	if bgpService == nil {
		return nil, fmt.Errorf("BGP service not initialized")
	}

	return bgpService.Get(ctx)
}

// ConfigureBGP creates a new BGP configuration
func (c *rtxClient) ConfigureBGP(ctx context.Context, config BGPConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	bgpService := c.bgpService
	c.mu.Unlock()

	if bgpService == nil {
		return fmt.Errorf("BGP service not initialized")
	}

	return bgpService.Configure(ctx, config)
}

// UpdateBGPConfig updates BGP configuration
func (c *rtxClient) UpdateBGPConfig(ctx context.Context, config BGPConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	bgpService := c.bgpService
	c.mu.Unlock()

	if bgpService == nil {
		return fmt.Errorf("BGP service not initialized")
	}

	return bgpService.Update(ctx, config)
}

// ResetBGP disables and removes BGP configuration
func (c *rtxClient) ResetBGP(ctx context.Context) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	bgpService := c.bgpService
	c.mu.Unlock()

	if bgpService == nil {
		return fmt.Errorf("BGP service not initialized")
	}

	return bgpService.Reset(ctx)
}

// GetOSPF retrieves OSPF configuration
func (c *rtxClient) GetOSPF(ctx context.Context) (*OSPFConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	ospfService := c.ospfService
	c.mu.Unlock()

	if ospfService == nil {
		return nil, fmt.Errorf("OSPF service not initialized")
	}

	return ospfService.Get(ctx)
}

// CreateOSPF creates OSPF configuration
func (c *rtxClient) CreateOSPF(ctx context.Context, config OSPFConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ospfService := c.ospfService
	c.mu.Unlock()

	if ospfService == nil {
		return fmt.Errorf("OSPF service not initialized")
	}

	return ospfService.Configure(ctx, config)
}

// UpdateOSPF updates OSPF configuration
func (c *rtxClient) UpdateOSPF(ctx context.Context, config OSPFConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ospfService := c.ospfService
	c.mu.Unlock()

	if ospfService == nil {
		return fmt.Errorf("OSPF service not initialized")
	}

	return ospfService.Update(ctx, config)
}

// DeleteOSPF disables and removes OSPF configuration
func (c *rtxClient) DeleteOSPF(ctx context.Context) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ospfService := c.ospfService
	c.mu.Unlock()

	if ospfService == nil {
		return fmt.Errorf("OSPF service not initialized")
	}

	return ospfService.Reset(ctx)
}

// GetIPsecTunnel retrieves an IPsec tunnel configuration
func (c *rtxClient) GetIPsecTunnel(ctx context.Context, tunnelID int) (*IPsecTunnel, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	ipsecService := c.ipsecTunnelService
	c.mu.Unlock()

	if ipsecService == nil {
		return nil, fmt.Errorf("IPsec tunnel service not initialized")
	}

	return ipsecService.Get(ctx, tunnelID)
}

// CreateIPsecTunnel creates an IPsec tunnel
func (c *rtxClient) CreateIPsecTunnel(ctx context.Context, tunnel IPsecTunnel) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ipsecService := c.ipsecTunnelService
	c.mu.Unlock()

	if ipsecService == nil {
		return fmt.Errorf("IPsec tunnel service not initialized")
	}

	return ipsecService.Create(ctx, tunnel)
}

// UpdateIPsecTunnel updates an IPsec tunnel
func (c *rtxClient) UpdateIPsecTunnel(ctx context.Context, tunnel IPsecTunnel) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ipsecService := c.ipsecTunnelService
	c.mu.Unlock()

	if ipsecService == nil {
		return fmt.Errorf("IPsec tunnel service not initialized")
	}

	return ipsecService.Update(ctx, tunnel)
}

// DeleteIPsecTunnel removes an IPsec tunnel
func (c *rtxClient) DeleteIPsecTunnel(ctx context.Context, tunnelID int) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ipsecService := c.ipsecTunnelService
	c.mu.Unlock()

	if ipsecService == nil {
		return fmt.Errorf("IPsec tunnel service not initialized")
	}

	return ipsecService.Delete(ctx, tunnelID)
}

// ListIPsecTunnels retrieves all IPsec tunnels
func (c *rtxClient) ListIPsecTunnels(ctx context.Context) ([]IPsecTunnel, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	ipsecService := c.ipsecTunnelService
	c.mu.Unlock()

	if ipsecService == nil {
		return nil, fmt.Errorf("IPsec tunnel service not initialized")
	}

	return ipsecService.List(ctx)
}

// GetL2TP retrieves an L2TP/L2TPv3 tunnel configuration
func (c *rtxClient) GetL2TP(ctx context.Context, tunnelID int) (*L2TPConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	l2tpService := c.l2tpService
	c.mu.Unlock()

	if l2tpService == nil {
		return nil, fmt.Errorf("L2TP service not initialized")
	}

	return l2tpService.Get(ctx, tunnelID)
}

// CreateL2TP creates an L2TP/L2TPv3 tunnel
func (c *rtxClient) CreateL2TP(ctx context.Context, config L2TPConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	l2tpService := c.l2tpService
	c.mu.Unlock()

	if l2tpService == nil {
		return fmt.Errorf("L2TP service not initialized")
	}

	return l2tpService.Create(ctx, config)
}

// UpdateL2TP updates an L2TP/L2TPv3 tunnel
func (c *rtxClient) UpdateL2TP(ctx context.Context, config L2TPConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	l2tpService := c.l2tpService
	c.mu.Unlock()

	if l2tpService == nil {
		return fmt.Errorf("L2TP service not initialized")
	}

	return l2tpService.Update(ctx, config)
}

// DeleteL2TP removes an L2TP/L2TPv3 tunnel
func (c *rtxClient) DeleteL2TP(ctx context.Context, tunnelID int) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	l2tpService := c.l2tpService
	c.mu.Unlock()

	if l2tpService == nil {
		return fmt.Errorf("L2TP service not initialized")
	}

	return l2tpService.Delete(ctx, tunnelID)
}

// ListL2TPs retrieves all L2TP/L2TPv3 tunnels
func (c *rtxClient) ListL2TPs(ctx context.Context) ([]L2TPConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	l2tpService := c.l2tpService
	c.mu.Unlock()

	if l2tpService == nil {
		return nil, fmt.Errorf("L2TP service not initialized")
	}

	return l2tpService.List(ctx)
}

// GetPPTP retrieves PPTP configuration
func (c *rtxClient) GetPPTP(ctx context.Context) (*PPTPConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	pptpService := c.pptpService
	c.mu.Unlock()

	if pptpService == nil {
		return nil, fmt.Errorf("PPTP service not initialized")
	}

	return pptpService.Get(ctx)
}

// CreatePPTP creates PPTP configuration
func (c *rtxClient) CreatePPTP(ctx context.Context, config PPTPConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	pptpService := c.pptpService
	c.mu.Unlock()

	if pptpService == nil {
		return fmt.Errorf("PPTP service not initialized")
	}

	return pptpService.Create(ctx, config)
}

// UpdatePPTP updates PPTP configuration
func (c *rtxClient) UpdatePPTP(ctx context.Context, config PPTPConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	pptpService := c.pptpService
	c.mu.Unlock()

	if pptpService == nil {
		return fmt.Errorf("PPTP service not initialized")
	}

	return pptpService.Update(ctx, config)
}

// DeletePPTP removes PPTP configuration
func (c *rtxClient) DeletePPTP(ctx context.Context) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	pptpService := c.pptpService
	c.mu.Unlock()

	if pptpService == nil {
		return fmt.Errorf("PPTP service not initialized")
	}

	return pptpService.Delete(ctx)
}

// GetSyslogConfig retrieves syslog configuration
func (c *rtxClient) GetSyslogConfig(ctx context.Context) (*SyslogConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	syslogService := c.syslogService
	c.mu.Unlock()

	if syslogService == nil {
		return nil, fmt.Errorf("syslog service not initialized")
	}

	return syslogService.Get(ctx)
}

// ConfigureSyslog creates syslog configuration
func (c *rtxClient) ConfigureSyslog(ctx context.Context, config SyslogConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	syslogService := c.syslogService
	c.mu.Unlock()

	if syslogService == nil {
		return fmt.Errorf("syslog service not initialized")
	}

	return syslogService.Configure(ctx, config)
}

// UpdateSyslogConfig updates syslog configuration
func (c *rtxClient) UpdateSyslogConfig(ctx context.Context, config SyslogConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	syslogService := c.syslogService
	c.mu.Unlock()

	if syslogService == nil {
		return fmt.Errorf("syslog service not initialized")
	}

	return syslogService.Update(ctx, config)
}

// ResetSyslog removes syslog configuration
func (c *rtxClient) ResetSyslog(ctx context.Context) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	syslogService := c.syslogService
	c.mu.Unlock()

	if syslogService == nil {
		return fmt.Errorf("syslog service not initialized")
	}

	return syslogService.Reset(ctx)
}

// GetDNS retrieves DNS server configuration
func (c *rtxClient) GetDNS(ctx context.Context) (*DNSConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	dnsService := c.dnsService
	c.mu.Unlock()

	if dnsService == nil {
		return nil, fmt.Errorf("DNS service not initialized")
	}

	return dnsService.Get(ctx)
}

// ConfigureDNS creates DNS server configuration
func (c *rtxClient) ConfigureDNS(ctx context.Context, config DNSConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	dnsService := c.dnsService
	c.mu.Unlock()

	if dnsService == nil {
		return fmt.Errorf("DNS service not initialized")
	}

	return dnsService.Configure(ctx, config)
}

// UpdateDNS updates DNS server configuration
func (c *rtxClient) UpdateDNS(ctx context.Context, config DNSConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	dnsService := c.dnsService
	c.mu.Unlock()

	if dnsService == nil {
		return fmt.Errorf("DNS service not initialized")
	}

	return dnsService.Update(ctx, config)
}

// ResetDNS removes DNS server configuration
func (c *rtxClient) ResetDNS(ctx context.Context) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	dnsService := c.dnsService
	c.mu.Unlock()

	if dnsService == nil {
		return fmt.Errorf("DNS service not initialized")
	}

	return dnsService.Reset(ctx)
}

// ========== QoS Class Map Methods ==========

// GetClassMap retrieves a class-map configuration
func (c *rtxClient) GetClassMap(ctx context.Context, name string) (*ClassMap, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return nil, fmt.Errorf("QoS service not initialized")
	}

	return qosService.GetClassMap(ctx, name)
}

// CreateClassMap creates a new class-map
func (c *rtxClient) CreateClassMap(ctx context.Context, cm ClassMap) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return fmt.Errorf("QoS service not initialized")
	}

	return qosService.CreateClassMap(ctx, cm)
}

// UpdateClassMap updates an existing class-map
func (c *rtxClient) UpdateClassMap(ctx context.Context, cm ClassMap) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return fmt.Errorf("QoS service not initialized")
	}

	return qosService.UpdateClassMap(ctx, cm)
}

// DeleteClassMap removes a class-map
func (c *rtxClient) DeleteClassMap(ctx context.Context, name string) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return fmt.Errorf("QoS service not initialized")
	}

	return qosService.DeleteClassMap(ctx, name)
}

// ListClassMaps retrieves all class-maps
func (c *rtxClient) ListClassMaps(ctx context.Context) ([]ClassMap, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return nil, fmt.Errorf("QoS service not initialized")
	}

	return qosService.ListClassMaps(ctx)
}

// ========== QoS Policy Map Methods ==========

// GetPolicyMap retrieves a policy-map configuration
func (c *rtxClient) GetPolicyMap(ctx context.Context, name string) (*PolicyMap, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return nil, fmt.Errorf("QoS service not initialized")
	}

	return qosService.GetPolicyMap(ctx, name)
}

// CreatePolicyMap creates a new policy-map
func (c *rtxClient) CreatePolicyMap(ctx context.Context, pm PolicyMap) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return fmt.Errorf("QoS service not initialized")
	}

	return qosService.CreatePolicyMap(ctx, pm)
}

// UpdatePolicyMap updates an existing policy-map
func (c *rtxClient) UpdatePolicyMap(ctx context.Context, pm PolicyMap) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return fmt.Errorf("QoS service not initialized")
	}

	return qosService.UpdatePolicyMap(ctx, pm)
}

// DeletePolicyMap removes a policy-map
func (c *rtxClient) DeletePolicyMap(ctx context.Context, name string) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return fmt.Errorf("QoS service not initialized")
	}

	return qosService.DeletePolicyMap(ctx, name)
}

// ListPolicyMaps retrieves all policy-maps
func (c *rtxClient) ListPolicyMaps(ctx context.Context) ([]PolicyMap, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return nil, fmt.Errorf("QoS service not initialized")
	}

	return qosService.ListPolicyMaps(ctx)
}

// ========== QoS Service Policy Methods ==========

// GetServicePolicy retrieves a service-policy configuration
func (c *rtxClient) GetServicePolicy(ctx context.Context, iface string, direction string) (*ServicePolicy, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return nil, fmt.Errorf("QoS service not initialized")
	}

	return qosService.GetServicePolicy(ctx, iface, direction)
}

// CreateServicePolicy creates a new service-policy
func (c *rtxClient) CreateServicePolicy(ctx context.Context, sp ServicePolicy) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return fmt.Errorf("QoS service not initialized")
	}

	return qosService.CreateServicePolicy(ctx, sp)
}

// UpdateServicePolicy updates an existing service-policy
func (c *rtxClient) UpdateServicePolicy(ctx context.Context, sp ServicePolicy) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return fmt.Errorf("QoS service not initialized")
	}

	return qosService.UpdateServicePolicy(ctx, sp)
}

// DeleteServicePolicy removes a service-policy
func (c *rtxClient) DeleteServicePolicy(ctx context.Context, iface string, direction string) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return fmt.Errorf("QoS service not initialized")
	}

	return qosService.DeleteServicePolicy(ctx, iface, direction)
}

// ListServicePolicies retrieves all service-policies
func (c *rtxClient) ListServicePolicies(ctx context.Context) ([]ServicePolicy, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return nil, fmt.Errorf("QoS service not initialized")
	}

	return qosService.ListServicePolicies(ctx)
}

// ========== QoS Shape Methods ==========

// GetShape retrieves a shape configuration
func (c *rtxClient) GetShape(ctx context.Context, iface string, direction string) (*ShapeConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return nil, fmt.Errorf("QoS service not initialized")
	}

	return qosService.GetShape(ctx, iface, direction)
}

// CreateShape creates a new shape configuration
func (c *rtxClient) CreateShape(ctx context.Context, sc ShapeConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return fmt.Errorf("QoS service not initialized")
	}

	return qosService.CreateShape(ctx, sc)
}

// UpdateShape updates an existing shape configuration
func (c *rtxClient) UpdateShape(ctx context.Context, sc ShapeConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return fmt.Errorf("QoS service not initialized")
	}

	return qosService.UpdateShape(ctx, sc)
}

// DeleteShape removes a shape configuration
func (c *rtxClient) DeleteShape(ctx context.Context, iface string, direction string) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return fmt.Errorf("QoS service not initialized")
	}

	return qosService.DeleteShape(ctx, iface, direction)
}

// ListShapes retrieves all shape configurations
func (c *rtxClient) ListShapes(ctx context.Context) ([]ShapeConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	qosService := c.qosService
	c.mu.Unlock()

	if qosService == nil {
		return nil, fmt.Errorf("QoS service not initialized")
	}

	return qosService.ListShapes(ctx)
}


// GetSchedule retrieves a schedule configuration
func (c *rtxClient) GetSchedule(ctx context.Context, id int) (*Schedule, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	scheduleService := c.scheduleService
	c.mu.Unlock()

	if scheduleService == nil {
		return nil, fmt.Errorf("schedule service not initialized")
	}

	return scheduleService.GetSchedule(ctx, id)
}

// CreateSchedule creates a new schedule
func (c *rtxClient) CreateSchedule(ctx context.Context, schedule Schedule) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	scheduleService := c.scheduleService
	c.mu.Unlock()

	if scheduleService == nil {
		return fmt.Errorf("schedule service not initialized")
	}

	return scheduleService.CreateSchedule(ctx, schedule)
}

// UpdateSchedule updates an existing schedule
func (c *rtxClient) UpdateSchedule(ctx context.Context, schedule Schedule) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	scheduleService := c.scheduleService
	c.mu.Unlock()

	if scheduleService == nil {
		return fmt.Errorf("schedule service not initialized")
	}

	return scheduleService.UpdateSchedule(ctx, schedule)
}

// DeleteSchedule removes a schedule
func (c *rtxClient) DeleteSchedule(ctx context.Context, id int) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	scheduleService := c.scheduleService
	c.mu.Unlock()

	if scheduleService == nil {
		return fmt.Errorf("schedule service not initialized")
	}

	return scheduleService.DeleteSchedule(ctx, id)
}

// ListSchedules retrieves all schedules
func (c *rtxClient) ListSchedules(ctx context.Context) ([]Schedule, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	scheduleService := c.scheduleService
	c.mu.Unlock()

	if scheduleService == nil {
		return nil, fmt.Errorf("schedule service not initialized")
	}

	return scheduleService.ListSchedules(ctx)
}

// GetKronPolicy retrieves a kron policy configuration
func (c *rtxClient) GetKronPolicy(ctx context.Context, name string) (*KronPolicy, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	scheduleService := c.scheduleService
	c.mu.Unlock()

	if scheduleService == nil {
		return nil, fmt.Errorf("schedule service not initialized")
	}

	return scheduleService.GetKronPolicy(ctx, name)
}

// CreateKronPolicy creates a new kron policy
func (c *rtxClient) CreateKronPolicy(ctx context.Context, policy KronPolicy) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	scheduleService := c.scheduleService
	c.mu.Unlock()

	if scheduleService == nil {
		return fmt.Errorf("schedule service not initialized")
	}

	return scheduleService.CreateKronPolicy(ctx, policy)
}

// UpdateKronPolicy updates an existing kron policy
func (c *rtxClient) UpdateKronPolicy(ctx context.Context, policy KronPolicy) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	scheduleService := c.scheduleService
	c.mu.Unlock()

	if scheduleService == nil {
		return fmt.Errorf("schedule service not initialized")
	}

	return scheduleService.UpdateKronPolicy(ctx, policy)
}

// DeleteKronPolicy removes a kron policy
func (c *rtxClient) DeleteKronPolicy(ctx context.Context, name string) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	scheduleService := c.scheduleService
	c.mu.Unlock()

	if scheduleService == nil {
		return fmt.Errorf("schedule service not initialized")
	}

	return scheduleService.DeleteKronPolicy(ctx, name)
}

// ListKronPolicies retrieves all kron policies
func (c *rtxClient) ListKronPolicies(ctx context.Context) ([]KronPolicy, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	scheduleService := c.scheduleService
	c.mu.Unlock()

	if scheduleService == nil {
		return nil, fmt.Errorf("schedule service not initialized")
	}

	return scheduleService.ListKronPolicies(ctx)
}

// ========== SNMP Methods ==========

// GetSNMP retrieves SNMP configuration
func (c *rtxClient) GetSNMP(ctx context.Context) (*SNMPConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	snmpService := c.snmpService
	c.mu.Unlock()

	if snmpService == nil {
		return nil, fmt.Errorf("SNMP service not initialized")
	}

	return snmpService.Get(ctx)
}

// CreateSNMP creates SNMP configuration
func (c *rtxClient) CreateSNMP(ctx context.Context, config SNMPConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	snmpService := c.snmpService
	c.mu.Unlock()

	if snmpService == nil {
		return fmt.Errorf("SNMP service not initialized")
	}

	return snmpService.Create(ctx, config)
}

// UpdateSNMP updates SNMP configuration
func (c *rtxClient) UpdateSNMP(ctx context.Context, config SNMPConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	snmpService := c.snmpService
	c.mu.Unlock()

	if snmpService == nil {
		return fmt.Errorf("SNMP service not initialized")
	}

	return snmpService.Update(ctx, config)
}

// DeleteSNMP removes SNMP configuration
func (c *rtxClient) DeleteSNMP(ctx context.Context) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	snmpService := c.snmpService
	c.mu.Unlock()

	if snmpService == nil {
		return fmt.Errorf("SNMP service not initialized")
	}

	return snmpService.Delete(ctx)
}

// ========== Admin Methods ==========

// GetAdminConfig retrieves admin password configuration
func (c *rtxClient) GetAdminConfig(ctx context.Context) (*AdminConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	adminService := c.adminService
	c.mu.Unlock()

	if adminService == nil {
		return nil, fmt.Errorf("admin service not initialized")
	}

	return adminService.GetAdminConfig(ctx)
}

// ConfigureAdmin sets admin password configuration
func (c *rtxClient) ConfigureAdmin(ctx context.Context, config AdminConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	adminService := c.adminService
	c.mu.Unlock()

	if adminService == nil {
		return fmt.Errorf("admin service not initialized")
	}

	return adminService.ConfigureAdmin(ctx, config)
}

// UpdateAdminConfig updates admin password configuration
func (c *rtxClient) UpdateAdminConfig(ctx context.Context, config AdminConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	adminService := c.adminService
	c.mu.Unlock()

	if adminService == nil {
		return fmt.Errorf("admin service not initialized")
	}

	return adminService.UpdateAdminConfig(ctx, config)
}

// ResetAdmin removes admin password configuration
func (c *rtxClient) ResetAdmin(ctx context.Context) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	adminService := c.adminService
	c.mu.Unlock()

	if adminService == nil {
		return fmt.Errorf("admin service not initialized")
	}

	return adminService.ResetAdmin(ctx)
}

// ========== Admin User Methods ==========

// GetAdminUser retrieves an admin user configuration
func (c *rtxClient) GetAdminUser(ctx context.Context, username string) (*AdminUser, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	adminService := c.adminService
	c.mu.Unlock()

	if adminService == nil {
		return nil, fmt.Errorf("admin service not initialized")
	}

	return adminService.GetAdminUser(ctx, username)
}

// CreateAdminUser creates a new admin user
func (c *rtxClient) CreateAdminUser(ctx context.Context, user AdminUser) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	adminService := c.adminService
	c.mu.Unlock()

	if adminService == nil {
		return fmt.Errorf("admin service not initialized")
	}

	return adminService.CreateAdminUser(ctx, user)
}

// UpdateAdminUser updates an existing admin user
func (c *rtxClient) UpdateAdminUser(ctx context.Context, user AdminUser) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	adminService := c.adminService
	c.mu.Unlock()

	if adminService == nil {
		return fmt.Errorf("admin service not initialized")
	}

	return adminService.UpdateAdminUser(ctx, user)
}

// DeleteAdminUser removes an admin user
func (c *rtxClient) DeleteAdminUser(ctx context.Context, username string) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	adminService := c.adminService
	c.mu.Unlock()

	if adminService == nil {
		return fmt.Errorf("admin service not initialized")
	}

	return adminService.DeleteAdminUser(ctx, username)
}

// ListAdminUsers retrieves all admin users
func (c *rtxClient) ListAdminUsers(ctx context.Context) ([]AdminUser, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	adminService := c.adminService
	c.mu.Unlock()

	if adminService == nil {
		return nil, fmt.Errorf("admin service not initialized")
	}

	return adminService.ListAdminUsers(ctx)
}

// ========== HTTPD Methods ==========

// GetHTTPD retrieves HTTPD configuration
func (c *rtxClient) GetHTTPD(ctx context.Context) (*HTTPDConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	serviceManager := c.serviceManager
	c.mu.Unlock()

	if serviceManager == nil {
		return nil, fmt.Errorf("service manager not initialized")
	}

	return serviceManager.GetHTTPD(ctx)
}

// ConfigureHTTPD creates HTTPD configuration
func (c *rtxClient) ConfigureHTTPD(ctx context.Context, config HTTPDConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	serviceManager := c.serviceManager
	c.mu.Unlock()

	if serviceManager == nil {
		return fmt.Errorf("service manager not initialized")
	}

	return serviceManager.ConfigureHTTPD(ctx, config)
}

// UpdateHTTPD updates HTTPD configuration
func (c *rtxClient) UpdateHTTPD(ctx context.Context, config HTTPDConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	serviceManager := c.serviceManager
	c.mu.Unlock()

	if serviceManager == nil {
		return fmt.Errorf("service manager not initialized")
	}

	return serviceManager.UpdateHTTPD(ctx, config)
}

// ResetHTTPD removes HTTPD configuration
func (c *rtxClient) ResetHTTPD(ctx context.Context) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	serviceManager := c.serviceManager
	c.mu.Unlock()

	if serviceManager == nil {
		return fmt.Errorf("service manager not initialized")
	}

	return serviceManager.ResetHTTPD(ctx)
}

// ========== SSHD Methods ==========

// GetSSHD retrieves SSHD configuration
func (c *rtxClient) GetSSHD(ctx context.Context) (*SSHDConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	serviceManager := c.serviceManager
	c.mu.Unlock()

	if serviceManager == nil {
		return nil, fmt.Errorf("service manager not initialized")
	}

	return serviceManager.GetSSHD(ctx)
}

// ConfigureSSHD creates SSHD configuration
func (c *rtxClient) ConfigureSSHD(ctx context.Context, config SSHDConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	serviceManager := c.serviceManager
	c.mu.Unlock()

	if serviceManager == nil {
		return fmt.Errorf("service manager not initialized")
	}

	return serviceManager.ConfigureSSHD(ctx, config)
}

// UpdateSSHD updates SSHD configuration
func (c *rtxClient) UpdateSSHD(ctx context.Context, config SSHDConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	serviceManager := c.serviceManager
	c.mu.Unlock()

	if serviceManager == nil {
		return fmt.Errorf("service manager not initialized")
	}

	return serviceManager.UpdateSSHD(ctx, config)
}

// ResetSSHD removes SSHD configuration
func (c *rtxClient) ResetSSHD(ctx context.Context) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	serviceManager := c.serviceManager
	c.mu.Unlock()

	if serviceManager == nil {
		return fmt.Errorf("service manager not initialized")
	}

	return serviceManager.ResetSSHD(ctx)
}

// ========== SFTPD Methods ==========

// GetSFTPD retrieves SFTPD configuration
func (c *rtxClient) GetSFTPD(ctx context.Context) (*SFTPDConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	serviceManager := c.serviceManager
	c.mu.Unlock()

	if serviceManager == nil {
		return nil, fmt.Errorf("service manager not initialized")
	}

	return serviceManager.GetSFTPD(ctx)
}

// ConfigureSFTPD creates SFTPD configuration
func (c *rtxClient) ConfigureSFTPD(ctx context.Context, config SFTPDConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	serviceManager := c.serviceManager
	c.mu.Unlock()

	if serviceManager == nil {
		return fmt.Errorf("service manager not initialized")
	}

	return serviceManager.ConfigureSFTPD(ctx, config)
}

// UpdateSFTPD updates SFTPD configuration
func (c *rtxClient) UpdateSFTPD(ctx context.Context, config SFTPDConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	serviceManager := c.serviceManager
	c.mu.Unlock()

	if serviceManager == nil {
		return fmt.Errorf("service manager not initialized")
	}

	return serviceManager.UpdateSFTPD(ctx, config)
}

// ResetSFTPD removes SFTPD configuration
func (c *rtxClient) ResetSFTPD(ctx context.Context) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	serviceManager := c.serviceManager
	c.mu.Unlock()

	if serviceManager == nil {
		return fmt.Errorf("service manager not initialized")
	}

	return serviceManager.ResetSFTPD(ctx)
}


// ========== Bridge Methods ==========

// GetBridge retrieves a bridge configuration
func (c *rtxClient) GetBridge(ctx context.Context, name string) (*BridgeConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	bridgeService := c.bridgeService
	c.mu.Unlock()

	if bridgeService == nil {
		return nil, fmt.Errorf("bridge service not initialized")
	}

	return bridgeService.GetBridge(ctx, name)
}

// CreateBridge creates a new bridge
func (c *rtxClient) CreateBridge(ctx context.Context, bridge BridgeConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	bridgeService := c.bridgeService
	c.mu.Unlock()

	if bridgeService == nil {
		return fmt.Errorf("bridge service not initialized")
	}

	return bridgeService.CreateBridge(ctx, bridge)
}

// UpdateBridge updates an existing bridge
func (c *rtxClient) UpdateBridge(ctx context.Context, bridge BridgeConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	bridgeService := c.bridgeService
	c.mu.Unlock()

	if bridgeService == nil {
		return fmt.Errorf("bridge service not initialized")
	}

	return bridgeService.UpdateBridge(ctx, bridge)
}

// DeleteBridge removes a bridge
func (c *rtxClient) DeleteBridge(ctx context.Context, name string) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	bridgeService := c.bridgeService
	c.mu.Unlock()

	if bridgeService == nil {
		return fmt.Errorf("bridge service not initialized")
	}

	return bridgeService.DeleteBridge(ctx, name)
}

// ListBridges retrieves all bridges
func (c *rtxClient) ListBridges(ctx context.Context) ([]BridgeConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	bridgeService := c.bridgeService
	c.mu.Unlock()

	if bridgeService == nil {
		return nil, fmt.Errorf("bridge service not initialized")
	}

	return bridgeService.ListBridges(ctx)
}

// ========== IPv6 Interface Methods ==========

// GetIPv6InterfaceConfig retrieves an IPv6 interface configuration
func (c *rtxClient) GetIPv6InterfaceConfig(ctx context.Context, interfaceName string) (*IPv6InterfaceConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	ipv6InterfaceService := c.ipv6InterfaceService
	c.mu.Unlock()

	if ipv6InterfaceService == nil {
		return nil, fmt.Errorf("IPv6 interface service not initialized")
	}

	return ipv6InterfaceService.Get(ctx, interfaceName)
}

// ConfigureIPv6Interface creates a new IPv6 interface configuration
func (c *rtxClient) ConfigureIPv6Interface(ctx context.Context, config IPv6InterfaceConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ipv6InterfaceService := c.ipv6InterfaceService
	c.mu.Unlock()

	if ipv6InterfaceService == nil {
		return fmt.Errorf("IPv6 interface service not initialized")
	}

	return ipv6InterfaceService.Configure(ctx, config)
}

// UpdateIPv6InterfaceConfig updates an existing IPv6 interface configuration
func (c *rtxClient) UpdateIPv6InterfaceConfig(ctx context.Context, config IPv6InterfaceConfig) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ipv6InterfaceService := c.ipv6InterfaceService
	c.mu.Unlock()

	if ipv6InterfaceService == nil {
		return fmt.Errorf("IPv6 interface service not initialized")
	}

	return ipv6InterfaceService.Update(ctx, config)
}

// ResetIPv6Interface removes IPv6 interface configuration
func (c *rtxClient) ResetIPv6Interface(ctx context.Context, interfaceName string) error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return fmt.Errorf("client not connected")
	}
	ipv6InterfaceService := c.ipv6InterfaceService
	c.mu.Unlock()

	if ipv6InterfaceService == nil {
		return fmt.Errorf("IPv6 interface service not initialized")
	}

	return ipv6InterfaceService.Reset(ctx, interfaceName)
}

// ListIPv6InterfaceConfigs retrieves all IPv6 interface configurations
func (c *rtxClient) ListIPv6InterfaceConfigs(ctx context.Context) ([]IPv6InterfaceConfig, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil, fmt.Errorf("client not connected")
	}
	ipv6InterfaceService := c.ipv6InterfaceService
	c.mu.Unlock()

	if ipv6InterfaceService == nil {
		return nil, fmt.Errorf("IPv6 interface service not initialized")
	}

	return ipv6InterfaceService.List(ctx)
}

// Access List Extended (IPv4) stub implementations
func (c *rtxClient) GetAccessListExtended(ctx context.Context, name string) (*AccessListExtended, error) {
	return nil, fmt.Errorf("access list extended not implemented")
}

func (c *rtxClient) CreateAccessListExtended(ctx context.Context, acl AccessListExtended) error {
	return fmt.Errorf("access list extended not implemented")
}

func (c *rtxClient) UpdateAccessListExtended(ctx context.Context, acl AccessListExtended) error {
	return fmt.Errorf("access list extended not implemented")
}

func (c *rtxClient) DeleteAccessListExtended(ctx context.Context, name string) error {
	return fmt.Errorf("access list extended not implemented")
}

func (c *rtxClient) ListAccessListsExtended(ctx context.Context) ([]AccessListExtended, error) {
	return nil, fmt.Errorf("access list extended not implemented")
}

// Access List Extended (IPv6) stub implementations
func (c *rtxClient) GetAccessListExtendedIPv6(ctx context.Context, name string) (*AccessListExtendedIPv6, error) {
	return nil, fmt.Errorf("access list extended IPv6 not implemented")
}

func (c *rtxClient) CreateAccessListExtendedIPv6(ctx context.Context, acl AccessListExtendedIPv6) error {
	return fmt.Errorf("access list extended IPv6 not implemented")
}

func (c *rtxClient) UpdateAccessListExtendedIPv6(ctx context.Context, acl AccessListExtendedIPv6) error {
	return fmt.Errorf("access list extended IPv6 not implemented")
}

func (c *rtxClient) DeleteAccessListExtendedIPv6(ctx context.Context, name string) error {
	return fmt.Errorf("access list extended IPv6 not implemented")
}

func (c *rtxClient) ListAccessListsExtendedIPv6(ctx context.Context) ([]AccessListExtendedIPv6, error) {
	return nil, fmt.Errorf("access list extended IPv6 not implemented")
}

// IP Filter Dynamic Config stub implementations
func (c *rtxClient) GetIPFilterDynamicConfig(ctx context.Context) (*IPFilterDynamicConfig, error) {
	return nil, fmt.Errorf("IP filter dynamic config not implemented")
}

func (c *rtxClient) CreateIPFilterDynamicConfig(ctx context.Context, config IPFilterDynamicConfig) error {
	return fmt.Errorf("IP filter dynamic config not implemented")
}

func (c *rtxClient) UpdateIPFilterDynamicConfig(ctx context.Context, config IPFilterDynamicConfig) error {
	return fmt.Errorf("IP filter dynamic config not implemented")
}

func (c *rtxClient) DeleteIPFilterDynamicConfig(ctx context.Context) error {
	return fmt.Errorf("IP filter dynamic config not implemented")
}

// IPv6 Filter Dynamic Config stub implementations
func (c *rtxClient) GetIPv6FilterDynamicConfig(ctx context.Context) (*IPv6FilterDynamicConfig, error) {
	return nil, fmt.Errorf("IPv6 filter dynamic config not implemented")
}

func (c *rtxClient) CreateIPv6FilterDynamicConfig(ctx context.Context, config IPv6FilterDynamicConfig) error {
	return fmt.Errorf("IPv6 filter dynamic config not implemented")
}

func (c *rtxClient) UpdateIPv6FilterDynamicConfig(ctx context.Context, config IPv6FilterDynamicConfig) error {
	return fmt.Errorf("IPv6 filter dynamic config not implemented")
}

func (c *rtxClient) DeleteIPv6FilterDynamicConfig(ctx context.Context) error {
	return fmt.Errorf("IPv6 filter dynamic config not implemented")
}

// Interface ACL stub implementations
func (c *rtxClient) GetInterfaceACL(ctx context.Context, iface string) (*InterfaceACL, error) {
	return nil, fmt.Errorf("interface ACL not implemented")
}

func (c *rtxClient) CreateInterfaceACL(ctx context.Context, acl InterfaceACL) error {
	return fmt.Errorf("interface ACL not implemented")
}

func (c *rtxClient) UpdateInterfaceACL(ctx context.Context, acl InterfaceACL) error {
	return fmt.Errorf("interface ACL not implemented")
}

func (c *rtxClient) DeleteInterfaceACL(ctx context.Context, iface string) error {
	return fmt.Errorf("interface ACL not implemented")
}

func (c *rtxClient) ListInterfaceACLs(ctx context.Context) ([]InterfaceACL, error) {
	return nil, fmt.Errorf("interface ACL not implemented")
}

// Access List MAC stub implementations
func (c *rtxClient) GetAccessListMAC(ctx context.Context, name string) (*AccessListMAC, error) {
	return nil, fmt.Errorf("access list MAC not implemented")
}

func (c *rtxClient) CreateAccessListMAC(ctx context.Context, acl AccessListMAC) error {
	return fmt.Errorf("access list MAC not implemented")
}

func (c *rtxClient) UpdateAccessListMAC(ctx context.Context, acl AccessListMAC) error {
	return fmt.Errorf("access list MAC not implemented")
}

func (c *rtxClient) DeleteAccessListMAC(ctx context.Context, name string) error {
	return fmt.Errorf("access list MAC not implemented")
}

func (c *rtxClient) ListAccessListsMAC(ctx context.Context) ([]AccessListMAC, error) {
	return nil, fmt.Errorf("access list MAC not implemented")
}

// Interface MAC ACL stub implementations
func (c *rtxClient) GetInterfaceMACACL(ctx context.Context, iface string) (*InterfaceMACACL, error) {
	return nil, fmt.Errorf("interface MAC ACL not implemented")
}

func (c *rtxClient) CreateInterfaceMACACL(ctx context.Context, acl InterfaceMACACL) error {
	return fmt.Errorf("interface MAC ACL not implemented")
}

func (c *rtxClient) UpdateInterfaceMACACL(ctx context.Context, acl InterfaceMACACL) error {
	return fmt.Errorf("interface MAC ACL not implemented")
}

func (c *rtxClient) DeleteInterfaceMACACL(ctx context.Context, iface string) error {
	return fmt.Errorf("interface MAC ACL not implemented")
}

func (c *rtxClient) ListInterfaceMACACLs(ctx context.Context) ([]InterfaceMACACL, error) {
	return nil, fmt.Errorf("interface MAC ACL not implemented")
}
