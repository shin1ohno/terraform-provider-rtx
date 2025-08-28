package client

import (
	"context"
	"time"
)

// Client is the main interface for interacting with RTX routers
type Client interface {
	// Dial establishes a connection to the RTX router
	Dial(ctx context.Context) error

	// Close terminates the connection
	Close() error

	// Run executes a command and returns the result
	Run(ctx context.Context, cmd Command) (Result, error)

	// GetSystemInfo retrieves system information from the router
	GetSystemInfo(ctx context.Context) (*SystemInfo, error)

	// GetInterfaces retrieves interface information from the router
	GetInterfaces(ctx context.Context) ([]Interface, error)

	// GetRoutes retrieves routing table information from the router
	GetRoutes(ctx context.Context) ([]Route, error)

	// GetStaticRoutes retrieves static route configurations from the router
	GetStaticRoutes(ctx context.Context) ([]StaticRoute, error)

	// GetDHCPScopes retrieves DHCP scope configurations from the router
	GetDHCPScopes(ctx context.Context) ([]DHCPScope, error)

	// GetDHCPScope retrieves a specific DHCP scope by ID
	GetDHCPScope(ctx context.Context, scopeID int) (*DHCPScope, error)

	// CreateDHCPScope creates a new DHCP scope
	CreateDHCPScope(ctx context.Context, scope DHCPScope) error

	// UpdateDHCPScope updates an existing DHCP scope
	UpdateDHCPScope(ctx context.Context, scope DHCPScope) error

	// DeleteDHCPScope removes a DHCP scope
	DeleteDHCPScope(ctx context.Context, scopeID int) error

	// GetDHCPBindings retrieves DHCP bindings for a scope
	GetDHCPBindings(ctx context.Context, scopeID int) ([]DHCPBinding, error)

	// CreateDHCPBinding creates a new DHCP binding
	CreateDHCPBinding(ctx context.Context, binding DHCPBinding) error

	// DeleteDHCPBinding removes a DHCP binding
	DeleteDHCPBinding(ctx context.Context, scopeID int, ipAddress string) error

	// SaveConfig saves the current configuration to persistent memory
	SaveConfig(ctx context.Context) error

	// Static Route management methods
	CreateStaticRoute(ctx context.Context, route StaticRoute) error
	GetStaticRoute(ctx context.Context, destination, gateway, iface string) (*StaticRoute, error)
	UpdateStaticRoute(ctx context.Context, route StaticRoute) error
	DeleteStaticRoute(ctx context.Context, destination, gateway, iface string) error
}

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

// Route represents a routing table entry on an RTX router
type Route struct {
	Destination string `json:"destination"`      // Network prefix (e.g., "192.168.1.0/24", "0.0.0.0/0")
	Gateway     string `json:"gateway"`          // Next hop gateway ("*" for directly connected routes)
	Interface   string `json:"interface"`        // Outgoing interface
	Protocol    string `json:"protocol"`         // S=static, C=connected, R=RIP, O=OSPF, B=BGP, D=DHCP
	Metric      *int   `json:"metric,omitempty"` // Route metric (optional)
}

// StaticRoute represents a static route configuration on an RTX router
type StaticRoute struct {
	Destination      string `json:"destination"`                  // Destination network prefix (required)
	GatewayIP        string `json:"gateway_ip,omitempty"`         // Next hop gateway IP address (optional)
	GatewayInterface string `json:"gateway_interface,omitempty"`  // Next hop gateway interface name (optional)
	Interface        string `json:"interface,omitempty"`          // Outgoing interface (optional)
	Metric           int    `json:"metric"`                       // Route metric (default: 1)
	Weight           int    `json:"weight"`                       // Route weight for ECMP (optional)
	Description      string `json:"description,omitempty"`        // Route description (optional)
	Hide             bool   `json:"hide,omitempty"`               // Hide flag (optional)
}

// DHCPScope represents a DHCP scope configuration
type DHCPScope struct {
	ID         int      `json:"id"`
	RangeStart string   `json:"range_start"`
	RangeEnd   string   `json:"range_end"`
	Prefix     int      `json:"prefix"`
	Gateway    string   `json:"gateway,omitempty"`
	DNSServers []string `json:"dns_servers,omitempty"`
	Lease      int      `json:"lease,omitempty"`
	DomainName string   `json:"domain_name,omitempty"`
}

// DHCPBinding represents a DHCP static lease binding
type DHCPBinding struct {
	ScopeID             int    `json:"scope_id"`
	IPAddress           string `json:"ip_address"`
	MACAddress          string `json:"mac_address"`
	ClientIdentifier    string `json:"client_identifier,omitempty"`
	UseClientIdentifier bool   `json:"use_client_identifier"`
}

// Command represents a command to be executed on the router
type Command struct {
	Key     string // Command identifier for parser lookup
	Payload string // Actual command string to send
}

// Result wraps the raw output and any parsed representation
type Result struct {
	Raw    []byte      // Raw command output
	Parsed interface{} // Parsed/typed representation
}

// Parser converts raw command output into structured data
type Parser interface {
	Parse(raw []byte) (interface{}, error)
}

// PromptDetector identifies when the router prompt appears in output
type PromptDetector interface {
	DetectPrompt(output []byte) (matched bool, prompt string)
}

// RetryStrategy determines retry behavior for transient failures
type RetryStrategy interface {
	Next(retry int) (delay time.Duration, giveUp bool)
}

// Session represents an SSH session with the router
type Session interface {
	Send(cmd string) ([]byte, error)
	Close() error
	SetAdminMode(bool) // Track if session is in administrator mode
}

// ConnDialer abstracts SSH connection creation
type ConnDialer interface {
	Dial(ctx context.Context, host string, config *Config) (Session, error)
}

// Config holds the SSH connection configuration
type Config struct {
	Host             string
	Port             int
	Username         string
	Password         string
	AdminPassword    string // Administrator password for configuration changes
	Timeout          int    // seconds
	HostKey          string // Fixed host key for verification (base64 encoded)
	KnownHostsFile   string // Path to known_hosts file
	SkipHostKeyCheck bool   // Skip host key verification (insecure)
}
