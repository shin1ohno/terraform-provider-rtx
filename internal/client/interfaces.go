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

	// GetDHCPBindings retrieves DHCP bindings for a scope
	GetDHCPBindings(ctx context.Context, scopeID int) ([]DHCPBinding, error)

	// CreateDHCPBinding creates a new DHCP binding
	CreateDHCPBinding(ctx context.Context, binding DHCPBinding) error

	// DeleteDHCPBinding removes a DHCP binding
	DeleteDHCPBinding(ctx context.Context, scopeID int, ipAddress string) error

	// GetDHCPScope retrieves a DHCP scope configuration
	GetDHCPScope(ctx context.Context, scopeID int) (*DHCPScope, error)

	// CreateDHCPScope creates a new DHCP scope
	CreateDHCPScope(ctx context.Context, scope DHCPScope) error

	// UpdateDHCPScope updates an existing DHCP scope
	UpdateDHCPScope(ctx context.Context, scope DHCPScope) error

	// DeleteDHCPScope removes a DHCP scope
	DeleteDHCPScope(ctx context.Context, scopeID int) error

	// ListDHCPScopes retrieves all DHCP scopes
	ListDHCPScopes(ctx context.Context) ([]DHCPScope, error)

	// GetIPv6Prefix retrieves an IPv6 prefix configuration
	GetIPv6Prefix(ctx context.Context, prefixID int) (*IPv6Prefix, error)

	// CreateIPv6Prefix creates a new IPv6 prefix
	CreateIPv6Prefix(ctx context.Context, prefix IPv6Prefix) error

	// UpdateIPv6Prefix updates an existing IPv6 prefix
	UpdateIPv6Prefix(ctx context.Context, prefix IPv6Prefix) error

	// DeleteIPv6Prefix removes an IPv6 prefix
	DeleteIPv6Prefix(ctx context.Context, prefixID int) error

	// ListIPv6Prefixes retrieves all IPv6 prefixes
	ListIPv6Prefixes(ctx context.Context) ([]IPv6Prefix, error)

	// GetVLAN retrieves a VLAN configuration
	GetVLAN(ctx context.Context, iface string, vlanID int) (*VLAN, error)

	// CreateVLAN creates a new VLAN
	CreateVLAN(ctx context.Context, vlan VLAN) error

	// UpdateVLAN updates an existing VLAN
	UpdateVLAN(ctx context.Context, vlan VLAN) error

	// DeleteVLAN removes a VLAN
	DeleteVLAN(ctx context.Context, iface string, vlanID int) error

	// ListVLANs retrieves all VLANs
	ListVLANs(ctx context.Context) ([]VLAN, error)

	// GetSystemConfig retrieves system configuration
	GetSystemConfig(ctx context.Context) (*SystemConfig, error)

	// ConfigureSystem sets system configuration
	ConfigureSystem(ctx context.Context, config SystemConfig) error

	// UpdateSystemConfig updates system configuration
	UpdateSystemConfig(ctx context.Context, config SystemConfig) error

	// ResetSystem resets system configuration to defaults
	ResetSystem(ctx context.Context) error

	// GetInterfaceConfig retrieves an interface configuration
	GetInterfaceConfig(ctx context.Context, interfaceName string) (*InterfaceConfig, error)

	// ConfigureInterface creates a new interface configuration
	ConfigureInterface(ctx context.Context, config InterfaceConfig) error

	// UpdateInterfaceConfig updates an existing interface configuration
	UpdateInterfaceConfig(ctx context.Context, config InterfaceConfig) error

	// ResetInterface removes interface configuration
	ResetInterface(ctx context.Context, interfaceName string) error

	// ListInterfaceConfigs retrieves all interface configurations
	ListInterfaceConfigs(ctx context.Context) ([]InterfaceConfig, error)

	// GetStaticRoute retrieves a static route configuration
	GetStaticRoute(ctx context.Context, prefix, mask string) (*StaticRoute, error)

	// CreateStaticRoute creates a new static route
	CreateStaticRoute(ctx context.Context, route StaticRoute) error

	// UpdateStaticRoute updates an existing static route
	UpdateStaticRoute(ctx context.Context, route StaticRoute) error

	// DeleteStaticRoute removes a static route
	DeleteStaticRoute(ctx context.Context, prefix, mask string) error

	// ListStaticRoutes retrieves all static routes
	ListStaticRoutes(ctx context.Context) ([]StaticRoute, error)

	// SaveConfig saves the current configuration to persistent memory
	SaveConfig(ctx context.Context) error

	// GetNATMasquerade retrieves a NAT masquerade configuration
	GetNATMasquerade(ctx context.Context, descriptorID int) (*NATMasquerade, error)

	// CreateNATMasquerade creates a new NAT masquerade
	CreateNATMasquerade(ctx context.Context, nat NATMasquerade) error

	// UpdateNATMasquerade updates an existing NAT masquerade
	UpdateNATMasquerade(ctx context.Context, nat NATMasquerade) error

	// DeleteNATMasquerade removes a NAT masquerade
	DeleteNATMasquerade(ctx context.Context, descriptorID int) error

	// ListNATMasquerades retrieves all NAT masquerades
	ListNATMasquerades(ctx context.Context) ([]NATMasquerade, error)

	// GetNATStatic retrieves a static NAT configuration
	GetNATStatic(ctx context.Context, descriptorID int) (*NATStatic, error)

	// CreateNATStatic creates a new static NAT
	CreateNATStatic(ctx context.Context, nat NATStatic) error

	// UpdateNATStatic updates an existing static NAT
	UpdateNATStatic(ctx context.Context, nat NATStatic) error

	// DeleteNATStatic removes a static NAT
	DeleteNATStatic(ctx context.Context, descriptorID int) error

	// ListNATStatics retrieves all static NATs
	ListNATStatics(ctx context.Context) ([]NATStatic, error)

	// GetIPFilter retrieves an IP filter configuration
	GetIPFilter(ctx context.Context, number int) (*IPFilter, error)

	// CreateIPFilter creates a new IP filter
	CreateIPFilter(ctx context.Context, filter IPFilter) error

	// UpdateIPFilter updates an existing IP filter
	UpdateIPFilter(ctx context.Context, filter IPFilter) error

	// DeleteIPFilter removes an IP filter
	DeleteIPFilter(ctx context.Context, number int) error

	// ListIPFilters retrieves all IP filters
	ListIPFilters(ctx context.Context) ([]IPFilter, error)

	// GetIPv6Filter retrieves an IPv6 filter configuration
	GetIPv6Filter(ctx context.Context, number int) (*IPFilter, error)

	// CreateIPv6Filter creates a new IPv6 filter
	CreateIPv6Filter(ctx context.Context, filter IPFilter) error

	// UpdateIPv6Filter updates an existing IPv6 filter
	UpdateIPv6Filter(ctx context.Context, filter IPFilter) error

	// DeleteIPv6Filter removes an IPv6 filter
	DeleteIPv6Filter(ctx context.Context, number int) error

	// ListIPv6Filters retrieves all IPv6 filters
	ListIPv6Filters(ctx context.Context) ([]IPFilter, error)

	// GetIPFilterDynamic retrieves a dynamic IP filter configuration
	GetIPFilterDynamic(ctx context.Context, number int) (*IPFilterDynamic, error)

	// CreateIPFilterDynamic creates a new dynamic IP filter
	CreateIPFilterDynamic(ctx context.Context, filter IPFilterDynamic) error

	// DeleteIPFilterDynamic removes a dynamic IP filter
	DeleteIPFilterDynamic(ctx context.Context, number int) error

	// ListIPFiltersDynamic retrieves all dynamic IP filters
	ListIPFiltersDynamic(ctx context.Context) ([]IPFilterDynamic, error)

	// GetEthernetFilter retrieves an Ethernet filter configuration
	GetEthernetFilter(ctx context.Context, number int) (*EthernetFilter, error)

	// CreateEthernetFilter creates a new Ethernet filter
	CreateEthernetFilter(ctx context.Context, filter EthernetFilter) error

	// UpdateEthernetFilter updates an existing Ethernet filter
	UpdateEthernetFilter(ctx context.Context, filter EthernetFilter) error

	// DeleteEthernetFilter removes an Ethernet filter
	DeleteEthernetFilter(ctx context.Context, number int) error

	// ListEthernetFilters retrieves all Ethernet filters
	ListEthernetFilters(ctx context.Context) ([]EthernetFilter, error)

	// BGP methods
	// GetBGPConfig retrieves BGP configuration
	GetBGPConfig(ctx context.Context) (*BGPConfig, error)

	// ConfigureBGP creates a new BGP configuration
	ConfigureBGP(ctx context.Context, config BGPConfig) error

	// UpdateBGPConfig updates BGP configuration
	UpdateBGPConfig(ctx context.Context, config BGPConfig) error

	// ResetBGP disables and removes BGP configuration
	ResetBGP(ctx context.Context) error

	// OSPF methods
	// GetOSPF retrieves OSPF configuration
	GetOSPF(ctx context.Context) (*OSPFConfig, error)

	// CreateOSPF creates OSPF configuration
	CreateOSPF(ctx context.Context, config OSPFConfig) error

	// UpdateOSPF updates OSPF configuration
	UpdateOSPF(ctx context.Context, config OSPFConfig) error

	// DeleteOSPF disables and removes OSPF configuration
	DeleteOSPF(ctx context.Context) error

	// IPsec Tunnel methods
	// GetIPsecTunnel retrieves an IPsec tunnel configuration
	GetIPsecTunnel(ctx context.Context, tunnelID int) (*IPsecTunnel, error)

	// CreateIPsecTunnel creates an IPsec tunnel
	CreateIPsecTunnel(ctx context.Context, tunnel IPsecTunnel) error

	// UpdateIPsecTunnel updates an IPsec tunnel
	UpdateIPsecTunnel(ctx context.Context, tunnel IPsecTunnel) error

	// DeleteIPsecTunnel removes an IPsec tunnel
	DeleteIPsecTunnel(ctx context.Context, tunnelID int) error

	// ListIPsecTunnels retrieves all IPsec tunnels
	ListIPsecTunnels(ctx context.Context) ([]IPsecTunnel, error)

	// IPsec Transport methods
	// GetIPsecTransport retrieves an IPsec transport configuration
	GetIPsecTransport(ctx context.Context, transportID int) (*IPsecTransportConfig, error)

	// CreateIPsecTransport creates an IPsec transport
	CreateIPsecTransport(ctx context.Context, transport IPsecTransportConfig) error

	// UpdateIPsecTransport updates an IPsec transport
	UpdateIPsecTransport(ctx context.Context, transport IPsecTransportConfig) error

	// DeleteIPsecTransport removes an IPsec transport
	DeleteIPsecTransport(ctx context.Context, transportID int) error

	// ListIPsecTransports retrieves all IPsec transports
	ListIPsecTransports(ctx context.Context) ([]IPsecTransportConfig, error)

	// L2TP methods
	// GetL2TP retrieves an L2TP/L2TPv3 tunnel configuration
	GetL2TP(ctx context.Context, tunnelID int) (*L2TPConfig, error)

	// CreateL2TP creates an L2TP/L2TPv3 tunnel
	CreateL2TP(ctx context.Context, config L2TPConfig) error

	// UpdateL2TP updates an L2TP/L2TPv3 tunnel
	UpdateL2TP(ctx context.Context, config L2TPConfig) error

	// DeleteL2TP removes an L2TP/L2TPv3 tunnel
	DeleteL2TP(ctx context.Context, tunnelID int) error

	// ListL2TPs retrieves all L2TP/L2TPv3 tunnels
	ListL2TPs(ctx context.Context) ([]L2TPConfig, error)

	// GetL2TPServiceState retrieves the L2TP service state (singleton)
	GetL2TPServiceState(ctx context.Context) (*L2TPServiceState, error)

	// SetL2TPServiceState sets the L2TP service state
	SetL2TPServiceState(ctx context.Context, enabled bool, protocols []string) error

	// PPTP methods
	// GetPPTP retrieves PPTP configuration
	GetPPTP(ctx context.Context) (*PPTPConfig, error)

	// CreatePPTP creates PPTP configuration
	CreatePPTP(ctx context.Context, config PPTPConfig) error

	// UpdatePPTP updates PPTP configuration
	UpdatePPTP(ctx context.Context, config PPTPConfig) error

	// DeletePPTP removes PPTP configuration
	DeletePPTP(ctx context.Context) error

	// Syslog methods (singleton resource)
	// GetSyslogConfig retrieves syslog configuration
	GetSyslogConfig(ctx context.Context) (*SyslogConfig, error)

	// ConfigureSyslog creates syslog configuration
	ConfigureSyslog(ctx context.Context, config SyslogConfig) error

	// UpdateSyslogConfig updates syslog configuration
	UpdateSyslogConfig(ctx context.Context, config SyslogConfig) error

	// ResetSyslog removes syslog configuration
	ResetSyslog(ctx context.Context) error

	// QoS Class Map methods
	// GetClassMap retrieves a class-map configuration
	GetClassMap(ctx context.Context, name string) (*ClassMap, error)

	// CreateClassMap creates a new class-map
	CreateClassMap(ctx context.Context, cm ClassMap) error

	// UpdateClassMap updates an existing class-map
	UpdateClassMap(ctx context.Context, cm ClassMap) error

	// DeleteClassMap removes a class-map
	DeleteClassMap(ctx context.Context, name string) error

	// ListClassMaps retrieves all class-maps
	ListClassMaps(ctx context.Context) ([]ClassMap, error)

	// QoS Policy Map methods
	// GetPolicyMap retrieves a policy-map configuration
	GetPolicyMap(ctx context.Context, name string) (*PolicyMap, error)

	// CreatePolicyMap creates a new policy-map
	CreatePolicyMap(ctx context.Context, pm PolicyMap) error

	// UpdatePolicyMap updates an existing policy-map
	UpdatePolicyMap(ctx context.Context, pm PolicyMap) error

	// DeletePolicyMap removes a policy-map
	DeletePolicyMap(ctx context.Context, name string) error

	// ListPolicyMaps retrieves all policy-maps
	ListPolicyMaps(ctx context.Context) ([]PolicyMap, error)

	// QoS Service Policy methods
	// GetServicePolicy retrieves a service-policy configuration
	GetServicePolicy(ctx context.Context, iface string, direction string) (*ServicePolicy, error)

	// CreateServicePolicy creates a new service-policy
	CreateServicePolicy(ctx context.Context, sp ServicePolicy) error

	// UpdateServicePolicy updates an existing service-policy
	UpdateServicePolicy(ctx context.Context, sp ServicePolicy) error

	// DeleteServicePolicy removes a service-policy
	DeleteServicePolicy(ctx context.Context, iface string, direction string) error

	// ListServicePolicies retrieves all service-policies
	ListServicePolicies(ctx context.Context) ([]ServicePolicy, error)

	// QoS Shape methods
	// GetShape retrieves a shape configuration
	GetShape(ctx context.Context, iface string, direction string) (*ShapeConfig, error)

	// CreateShape creates a new shape configuration
	CreateShape(ctx context.Context, sc ShapeConfig) error

	// UpdateShape updates an existing shape configuration
	UpdateShape(ctx context.Context, sc ShapeConfig) error

	// DeleteShape removes a shape configuration
	DeleteShape(ctx context.Context, iface string, direction string) error

	// ListShapes retrieves all shape configurations
	ListShapes(ctx context.Context) ([]ShapeConfig, error)

	// SNMP methods (singleton resource)
	// GetSNMP retrieves SNMP configuration
	GetSNMP(ctx context.Context) (*SNMPConfig, error)

	// CreateSNMP creates SNMP configuration
	CreateSNMP(ctx context.Context, config SNMPConfig) error

	// UpdateSNMP updates SNMP configuration
	UpdateSNMP(ctx context.Context, config SNMPConfig) error

	// DeleteSNMP removes SNMP configuration
	DeleteSNMP(ctx context.Context) error

	// Schedule methods
	// GetSchedule retrieves a schedule configuration
	GetSchedule(ctx context.Context, id int) (*Schedule, error)

	// CreateSchedule creates a new schedule
	CreateSchedule(ctx context.Context, schedule Schedule) error

	// UpdateSchedule updates an existing schedule
	UpdateSchedule(ctx context.Context, schedule Schedule) error

	// DeleteSchedule removes a schedule
	DeleteSchedule(ctx context.Context, id int) error

	// ListSchedules retrieves all schedules
	ListSchedules(ctx context.Context) ([]Schedule, error)

	// KronPolicy methods
	// GetKronPolicy retrieves a kron policy configuration
	GetKronPolicy(ctx context.Context, name string) (*KronPolicy, error)

	// CreateKronPolicy creates a new kron policy
	CreateKronPolicy(ctx context.Context, policy KronPolicy) error

	// UpdateKronPolicy updates an existing kron policy
	UpdateKronPolicy(ctx context.Context, policy KronPolicy) error

	// DeleteKronPolicy removes a kron policy
	DeleteKronPolicy(ctx context.Context, name string) error

	// ListKronPolicies retrieves all kron policies
	ListKronPolicies(ctx context.Context) ([]KronPolicy, error)

	// DNS methods (singleton resource)
	// GetDNS retrieves DNS server configuration
	GetDNS(ctx context.Context) (*DNSConfig, error)

	// ConfigureDNS creates DNS server configuration
	ConfigureDNS(ctx context.Context, config DNSConfig) error

	// UpdateDNS updates DNS server configuration
	UpdateDNS(ctx context.Context, config DNSConfig) error

	// ResetDNS removes DNS server configuration
	ResetDNS(ctx context.Context) error

	// Admin methods (singleton resource)
	// GetAdminConfig retrieves admin password configuration
	GetAdminConfig(ctx context.Context) (*AdminConfig, error)

	// ConfigureAdmin sets admin password configuration
	ConfigureAdmin(ctx context.Context, config AdminConfig) error

	// UpdateAdminConfig updates admin password configuration
	UpdateAdminConfig(ctx context.Context, config AdminConfig) error

	// ResetAdmin removes admin password configuration
	ResetAdmin(ctx context.Context) error

	// Admin User methods
	// GetAdminUser retrieves an admin user configuration
	GetAdminUser(ctx context.Context, username string) (*AdminUser, error)

	// CreateAdminUser creates a new admin user
	CreateAdminUser(ctx context.Context, user AdminUser) error

	// UpdateAdminUser updates an existing admin user
	UpdateAdminUser(ctx context.Context, user AdminUser) error

	// DeleteAdminUser removes an admin user
	DeleteAdminUser(ctx context.Context, username string) error

	// ListAdminUsers retrieves all admin users
	ListAdminUsers(ctx context.Context) ([]AdminUser, error)

	// HTTPD methods (singleton resource)
	// GetHTTPD retrieves HTTPD configuration
	GetHTTPD(ctx context.Context) (*HTTPDConfig, error)

	// ConfigureHTTPD creates HTTPD configuration
	ConfigureHTTPD(ctx context.Context, config HTTPDConfig) error

	// UpdateHTTPD updates HTTPD configuration
	UpdateHTTPD(ctx context.Context, config HTTPDConfig) error

	// ResetHTTPD removes HTTPD configuration
	ResetHTTPD(ctx context.Context) error

	// SSHD methods (singleton resource)
	// GetSSHD retrieves SSHD configuration
	GetSSHD(ctx context.Context) (*SSHDConfig, error)

	// ConfigureSSHD creates SSHD configuration
	ConfigureSSHD(ctx context.Context, config SSHDConfig) error

	// UpdateSSHD updates SSHD configuration
	UpdateSSHD(ctx context.Context, config SSHDConfig) error

	// ResetSSHD removes SSHD configuration
	ResetSSHD(ctx context.Context) error

	// SFTPD methods (singleton resource)
	// GetSFTPD retrieves SFTPD configuration
	GetSFTPD(ctx context.Context) (*SFTPDConfig, error)

	// ConfigureSFTPD creates SFTPD configuration
	ConfigureSFTPD(ctx context.Context, config SFTPDConfig) error

	// UpdateSFTPD updates SFTPD configuration
	UpdateSFTPD(ctx context.Context, config SFTPDConfig) error

	// ResetSFTPD removes SFTPD configuration
	ResetSFTPD(ctx context.Context) error

	// Bridge methods
	// GetBridge retrieves a bridge configuration
	GetBridge(ctx context.Context, name string) (*BridgeConfig, error)

	// CreateBridge creates a new bridge
	CreateBridge(ctx context.Context, bridge BridgeConfig) error

	// UpdateBridge updates an existing bridge
	UpdateBridge(ctx context.Context, bridge BridgeConfig) error

	// DeleteBridge removes a bridge
	DeleteBridge(ctx context.Context, name string) error

	// ListBridges retrieves all bridges
	ListBridges(ctx context.Context) ([]BridgeConfig, error)

	// IPv6 Interface methods
	// GetIPv6InterfaceConfig retrieves an IPv6 interface configuration
	GetIPv6InterfaceConfig(ctx context.Context, interfaceName string) (*IPv6InterfaceConfig, error)

	// ConfigureIPv6Interface creates a new IPv6 interface configuration
	ConfigureIPv6Interface(ctx context.Context, config IPv6InterfaceConfig) error

	// UpdateIPv6InterfaceConfig updates an existing IPv6 interface configuration
	UpdateIPv6InterfaceConfig(ctx context.Context, config IPv6InterfaceConfig) error

	// ResetIPv6Interface removes IPv6 interface configuration
	ResetIPv6Interface(ctx context.Context, interfaceName string) error

	// ListIPv6InterfaceConfigs retrieves all IPv6 interface configurations
	ListIPv6InterfaceConfigs(ctx context.Context) ([]IPv6InterfaceConfig, error)

	// Access List Extended (IPv4) methods
	// GetAccessListExtended retrieves an IPv4 extended access list
	GetAccessListExtended(ctx context.Context, name string) (*AccessListExtended, error)

	// CreateAccessListExtended creates a new IPv4 extended access list
	CreateAccessListExtended(ctx context.Context, acl AccessListExtended) error

	// UpdateAccessListExtended updates an existing IPv4 extended access list
	UpdateAccessListExtended(ctx context.Context, acl AccessListExtended) error

	// DeleteAccessListExtended removes an IPv4 extended access list
	DeleteAccessListExtended(ctx context.Context, name string) error

	// ListAccessListsExtended retrieves all IPv4 extended access lists
	ListAccessListsExtended(ctx context.Context) ([]AccessListExtended, error)

	// Access List Extended (IPv6) methods
	// GetAccessListExtendedIPv6 retrieves an IPv6 extended access list
	GetAccessListExtendedIPv6(ctx context.Context, name string) (*AccessListExtendedIPv6, error)

	// CreateAccessListExtendedIPv6 creates a new IPv6 extended access list
	CreateAccessListExtendedIPv6(ctx context.Context, acl AccessListExtendedIPv6) error

	// UpdateAccessListExtendedIPv6 updates an existing IPv6 extended access list
	UpdateAccessListExtendedIPv6(ctx context.Context, acl AccessListExtendedIPv6) error

	// DeleteAccessListExtendedIPv6 removes an IPv6 extended access list
	DeleteAccessListExtendedIPv6(ctx context.Context, name string) error

	// ListAccessListsExtendedIPv6 retrieves all IPv6 extended access lists
	ListAccessListsExtendedIPv6(ctx context.Context) ([]AccessListExtendedIPv6, error)

	// IP Filter Dynamic Config methods
	// GetIPFilterDynamicConfig retrieves the IP filter dynamic configuration
	GetIPFilterDynamicConfig(ctx context.Context) (*IPFilterDynamicConfig, error)

	// CreateIPFilterDynamicConfig creates the IP filter dynamic configuration
	CreateIPFilterDynamicConfig(ctx context.Context, config IPFilterDynamicConfig) error

	// UpdateIPFilterDynamicConfig updates the IP filter dynamic configuration
	UpdateIPFilterDynamicConfig(ctx context.Context, config IPFilterDynamicConfig) error

	// DeleteIPFilterDynamicConfig removes all IP filter dynamic entries
	DeleteIPFilterDynamicConfig(ctx context.Context) error

	// IPv6 Filter Dynamic Config methods
	// GetIPv6FilterDynamicConfig retrieves the IPv6 filter dynamic configuration
	GetIPv6FilterDynamicConfig(ctx context.Context) (*IPv6FilterDynamicConfig, error)

	// CreateIPv6FilterDynamicConfig creates the IPv6 filter dynamic configuration
	CreateIPv6FilterDynamicConfig(ctx context.Context, config IPv6FilterDynamicConfig) error

	// UpdateIPv6FilterDynamicConfig updates the IPv6 filter dynamic configuration
	UpdateIPv6FilterDynamicConfig(ctx context.Context, config IPv6FilterDynamicConfig) error

	// DeleteIPv6FilterDynamicConfig removes all IPv6 filter dynamic entries
	DeleteIPv6FilterDynamicConfig(ctx context.Context) error

	// Interface ACL methods
	// GetInterfaceACL retrieves ACL bindings for an interface
	GetInterfaceACL(ctx context.Context, iface string) (*InterfaceACL, error)

	// CreateInterfaceACL creates ACL bindings for an interface
	CreateInterfaceACL(ctx context.Context, acl InterfaceACL) error

	// UpdateInterfaceACL updates ACL bindings for an interface
	UpdateInterfaceACL(ctx context.Context, acl InterfaceACL) error

	// DeleteInterfaceACL removes ACL bindings from an interface
	DeleteInterfaceACL(ctx context.Context, iface string) error

	// ListInterfaceACLs retrieves all interface ACL bindings
	ListInterfaceACLs(ctx context.Context) ([]InterfaceACL, error)

	// Access List MAC methods
	// GetAccessListMAC retrieves a MAC access list
	GetAccessListMAC(ctx context.Context, name string) (*AccessListMAC, error)

	// CreateAccessListMAC creates a new MAC access list
	CreateAccessListMAC(ctx context.Context, acl AccessListMAC) error

	// UpdateAccessListMAC updates an existing MAC access list
	UpdateAccessListMAC(ctx context.Context, acl AccessListMAC) error

	// DeleteAccessListMAC removes a MAC access list
	DeleteAccessListMAC(ctx context.Context, name string) error

	// ListAccessListsMAC retrieves all MAC access lists
	ListAccessListsMAC(ctx context.Context) ([]AccessListMAC, error)

	// Interface MAC ACL methods
	// GetInterfaceMACACL retrieves MAC ACL bindings for an interface
	GetInterfaceMACACL(ctx context.Context, iface string) (*InterfaceMACACL, error)

	// CreateInterfaceMACACL creates MAC ACL bindings for an interface
	CreateInterfaceMACACL(ctx context.Context, acl InterfaceMACACL) error

	// UpdateInterfaceMACACL updates MAC ACL bindings for an interface
	UpdateInterfaceMACACL(ctx context.Context, acl InterfaceMACACL) error

	// DeleteInterfaceMACACL removes MAC ACL bindings from an interface
	DeleteInterfaceMACACL(ctx context.Context, iface string) error

	// ListInterfaceMACACLs retrieves all interface MAC ACL bindings
	ListInterfaceMACACLs(ctx context.Context) ([]InterfaceMACACL, error)
}

// Interface represents a network interface on an RTX router
type Interface struct {
	Name        string            `json:"name"`
	Kind        string            `json:"kind"`        // lan, wan, pp, vlan
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
	Destination string `json:"destination"`       // Network prefix (e.g., "192.168.1.0/24", "0.0.0.0/0")
	Gateway     string `json:"gateway"`           // Next hop gateway ("*" for directly connected routes)
	Interface   string `json:"interface"`         // Outgoing interface
	Protocol    string `json:"protocol"`          // S=static, C=connected, R=RIP, O=OSPF, B=BGP, D=DHCP
	Metric      *int   `json:"metric,omitempty"`  // Route metric (optional)
}

// DHCPBinding represents a DHCP static lease binding
type DHCPBinding struct {
	ScopeID             int    `json:"scope_id"`
	IPAddress           string `json:"ip_address"`
	MACAddress          string `json:"mac_address"`
	ClientIdentifier    string `json:"client_identifier,omitempty"`
	UseClientIdentifier bool   `json:"use_client_identifier"`
}

// DHCPScope represents a DHCP scope configuration on an RTX router
type DHCPScope struct {
	ScopeID       int              `json:"scope_id"`
	Network       string           `json:"network"`                  // CIDR notation: "192.168.1.0/24"
	LeaseTime     string           `json:"lease_time,omitempty"`     // Go duration format or "infinite"
	ExcludeRanges []ExcludeRange   `json:"exclude_ranges,omitempty"` // Excluded IP ranges
	Options       DHCPScopeOptions `json:"options,omitempty"`        // DHCP options (dns, routers, etc.)
}

// DHCPScopeOptions represents DHCP options for a scope (Cisco-compatible naming)
type DHCPScopeOptions struct {
	DNSServers []string `json:"dns_servers,omitempty"` // DNS servers (max 3)
	Routers    []string `json:"routers,omitempty"`     // Default gateways (max 3)
	DomainName string   `json:"domain_name,omitempty"` // Domain name
}

// ExcludeRange represents an IP range excluded from DHCP allocation
type ExcludeRange struct {
	Start string `json:"start"` // Start IP address
	End   string `json:"end"`   // End IP address
}

// IPv6Prefix represents an IPv6 prefix definition on an RTX router
type IPv6Prefix struct {
	ID           int    `json:"id"`                  // Prefix ID (1-255)
	Prefix       string `json:"prefix"`              // Static prefix value (e.g., "2001:db8::")
	PrefixLength int    `json:"prefix_length"`       // Prefix length (e.g., 64)
	Source       string `json:"source"`              // "static", "ra", or "dhcpv6-pd"
	Interface    string `json:"interface,omitempty"` // Source interface for ra/pd
}

// VLAN represents a VLAN configuration on an RTX router
type VLAN struct {
	VlanID        int    `json:"vlan_id"`              // VLAN ID (1-4094)
	Name          string `json:"name,omitempty"`       // VLAN name/description
	Interface     string `json:"interface"`            // Parent interface (lan1, lan2)
	VlanInterface string `json:"vlan_interface"`       // Computed: lan1/1, lan1/2, etc.
	IPAddress     string `json:"ip_address,omitempty"` // IP address
	IPMask        string `json:"ip_mask,omitempty"`    // Subnet mask
	Shutdown      bool   `json:"shutdown"`             // Admin state (true = shutdown)
}

// SystemConfig represents system-level configuration on an RTX router
type SystemConfig struct {
	Timezone      string               `json:"timezone,omitempty"`       // UTC offset (e.g., "+09:00")
	Console       *ConsoleConfig       `json:"console,omitempty"`        // Console settings
	PacketBuffers []PacketBufferConfig `json:"packet_buffers,omitempty"` // Packet buffer tuning
	Statistics    *StatisticsConfig    `json:"statistics,omitempty"`     // Statistics collection
}

// ConsoleConfig represents console settings
type ConsoleConfig struct {
	Character string `json:"character,omitempty"` // Character encoding (ja.utf8, ascii, ja.sjis)
	Lines     string `json:"lines,omitempty"`     // Lines per page (number or "infinity")
	Prompt    string `json:"prompt,omitempty"`    // Custom prompt string
}

// PacketBufferConfig represents packet buffer tuning for each size
type PacketBufferConfig struct {
	Size      string `json:"size"`       // "small", "middle", or "large"
	MaxBuffer int    `json:"max_buffer"` // Maximum buffer count
	MaxFree   int    `json:"max_free"`   // Maximum free buffer count
}

// StatisticsConfig represents statistics collection settings
type StatisticsConfig struct {
	Traffic bool `json:"traffic"` // Traffic statistics
	NAT     bool `json:"nat"`     // NAT statistics
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

// InterfaceConfig represents interface configuration on an RTX router
type InterfaceConfig struct {
	Name             string       `json:"name"`                         // Interface name (lan1, lan2, pp1, bridge1, tunnel1)
	Description      string       `json:"description,omitempty"`        // Interface description
	IPAddress        *InterfaceIP `json:"ip_address,omitempty"`         // IPv4 address configuration
	SecureFilterIn   []int        `json:"secure_filter_in,omitempty"`   // Inbound security filter numbers
	SecureFilterOut  []int        `json:"secure_filter_out,omitempty"`  // Outbound security filter numbers
	DynamicFilterOut []int        `json:"dynamic_filter_out,omitempty"` // Dynamic filters for outbound
	NATDescriptor    int          `json:"nat_descriptor,omitempty"`     // NAT descriptor number (0 = none)
	ProxyARP         bool         `json:"proxyarp"`                     // Enable ProxyARP
	MTU              int          `json:"mtu,omitempty"`                // MTU size (0 = default)
}

// InterfaceIP represents IP address configuration
type InterfaceIP struct {
	Address string `json:"address,omitempty"` // CIDR notation (192.168.1.1/24) or empty if DHCP
	DHCP    bool   `json:"dhcp"`              // Use DHCP for address assignment
}

// StaticRoute represents a static route configuration on an RTX router
type StaticRoute struct {
	Prefix   string           `json:"prefix"`    // Route destination (e.g., "0.0.0.0" for default)
	Mask     string           `json:"mask"`      // Subnet mask (e.g., "0.0.0.0" for default)
	NextHops []StaticRouteHop `json:"next_hops"` // List of next hops
}

// StaticRouteHop represents a next hop configuration for a static route
type StaticRouteHop struct {
	NextHop   string `json:"next_hop,omitempty"`  // Gateway IP address
	Interface string `json:"interface,omitempty"` // Interface (pp 1, tunnel 1, etc.)
	Distance  int    `json:"distance"`            // Administrative distance (weight)
	Name      string `json:"name,omitempty"`      // Route description
	Permanent bool   `json:"permanent"`           // Keep route when interface down
	Filter    int    `json:"filter,omitempty"`    // IP filter number (RTX-specific)
}

// NATMasquerade represents a NAT masquerade configuration on an RTX router
type NATMasquerade struct {
	DescriptorID  int                     `json:"descriptor_id"`            // NAT descriptor ID (1-65535)
	OuterAddress  string                  `json:"outer_address"`            // "ipcp", interface name, or specific IP
	InnerNetwork  string                  `json:"inner_network"`            // IP range: "192.168.1.0-192.168.1.255"
	StaticEntries []MasqueradeStaticEntry `json:"static_entries,omitempty"` // Static port mappings
}

// MasqueradeStaticEntry represents a static port mapping entry for NAT masquerade
type MasqueradeStaticEntry struct {
	EntryNumber       int    `json:"entry_number"`        // Entry number for identification
	InsideLocal       string `json:"inside_local"`        // Internal IP address
	InsideLocalPort   int    `json:"inside_local_port"`   // Internal port
	OutsideGlobal     string `json:"outside_global"`      // External IP address (or "ipcp")
	OutsideGlobalPort int    `json:"outside_global_port"` // External port
	Protocol          string `json:"protocol,omitempty"`  // "tcp", "udp", or empty for any
}

// NATStatic represents a static NAT descriptor configuration on an RTX router
type NATStatic struct {
	DescriptorID int              `json:"descriptor_id"` // NAT descriptor ID (1-65535)
	Entries      []NATStaticEntry `json:"entries,omitempty"`
}

// NATStaticEntry represents a single static NAT mapping entry
type NATStaticEntry struct {
	InsideLocal       string `json:"inside_local"`                  // Inside local IP address
	InsideLocalPort   int    `json:"inside_local_port,omitempty"`   // Inside local port (for port NAT)
	OutsideGlobal     string `json:"outside_global"`                // Outside global IP address
	OutsideGlobalPort int    `json:"outside_global_port,omitempty"` // Outside global port (for port NAT)
	Protocol          string `json:"protocol,omitempty"`            // Protocol: tcp, udp (for port NAT)
}

// EthernetFilter represents an Ethernet (Layer 2) filter configuration on an RTX router
type EthernetFilter struct {
	Number    int    `json:"number"`               // Filter number (1-65535)
	Action    string `json:"action"`               // pass or reject
	SourceMAC string `json:"source_mac"`           // Source MAC address (* for any)
	DestMAC   string `json:"dest_mac"`             // Destination MAC address (* for any)
	EtherType string `json:"ether_type,omitempty"` // Ethernet type (e.g., 0x0800, 0x0806)
	VlanID    int    `json:"vlan_id,omitempty"`    // VLAN ID (1-4094, 0 means not specified)
}

// IPFilter represents a static IP filter rule on an RTX router
type IPFilter struct {
	Number        int    `json:"number"`                   // Filter number (1-65535)
	Action        string `json:"action"`                   // pass, reject, restrict, restrict-log
	SourceAddress string `json:"source_address"`           // Source IP/network or "*"
	SourceMask    string `json:"source_mask,omitempty"`    // Source mask (for non-CIDR format)
	DestAddress   string `json:"dest_address"`             // Destination IP/network or "*"
	DestMask      string `json:"dest_mask,omitempty"`      // Destination mask (for non-CIDR format)
	Protocol      string `json:"protocol"`                 // tcp, udp, icmp, ip, * (any)
	SourcePort    string `json:"source_port,omitempty"`    // Source port(s) or "*"
	DestPort      string `json:"dest_port,omitempty"`      // Destination port(s) or "*"
	Established   bool   `json:"established,omitempty"`    // Match established TCP connections
}

// IPFilterDynamic represents a dynamic (stateful) IP filter on an RTX router
type IPFilterDynamic struct {
	Number   int    `json:"number"`            // Filter number (1-65535)
	Source   string `json:"source"`            // Source address or "*"
	Dest     string `json:"dest"`              // Destination address or "*"
	Protocol string `json:"protocol"`          // Protocol (ftp, www, smtp, etc.)
	SyslogOn bool   `json:"syslog,omitempty"`  // Enable syslog for this filter
}

// BGPConfig represents BGP configuration on an RTX router
type BGPConfig struct {
	Enabled               bool          `json:"enabled"`
	ASN                   string        `json:"asn"`                              // String for 4-byte ASN support
	RouterID              string        `json:"router_id,omitempty"`              // Optional router ID
	DefaultIPv4Unicast    bool          `json:"default_ipv4_unicast"`             // Default: true
	LogNeighborChanges    bool          `json:"log_neighbor_changes"`             // Default: true
	Neighbors             []BGPNeighbor `json:"neighbors,omitempty"`              // BGP neighbors
	Networks              []BGPNetwork  `json:"networks,omitempty"`               // Announced networks
	RedistributeStatic    bool          `json:"redistribute_static,omitempty"`    // Redistribute static routes
	RedistributeConnected bool          `json:"redistribute_connected,omitempty"` // Redistribute connected routes
}

// BGPNeighbor represents a BGP neighbor configuration
type BGPNeighbor struct {
	ID           int    `json:"id"`                      // Neighbor ID (1-based)
	IP           string `json:"ip"`                      // Neighbor IP address
	RemoteAS     string `json:"remote_as"`               // Remote AS number
	HoldTime     int    `json:"hold_time,omitempty"`     // Hold time in seconds
	Keepalive    int    `json:"keepalive,omitempty"`     // Keepalive interval
	Multihop     int    `json:"multihop,omitempty"`      // eBGP multihop TTL
	Password     string `json:"password,omitempty"`      // MD5 authentication password
	LocalAddress string `json:"local_address,omitempty"` // Local address for session
}

// BGPNetwork represents a BGP network announcement
type BGPNetwork struct {
	Prefix string `json:"prefix"` // Network prefix
	Mask   string `json:"mask"`   // Network mask (dotted decimal)
}

// OSPFConfig represents OSPF configuration on an RTX router
type OSPFConfig struct {
	Enabled               bool           `json:"enabled"`
	ProcessID             int            `json:"process_id,omitempty"`             // OSPF process ID (default: 1)
	RouterID              string         `json:"router_id"`                        // Router ID (required)
	Distance              int            `json:"distance,omitempty"`               // Administrative distance (default: 110)
	DefaultOriginate      bool           `json:"default_originate,omitempty"`      // Originate default route
	Networks              []OSPFNetwork  `json:"networks,omitempty"`               // OSPF networks
	Areas                 []OSPFArea     `json:"areas,omitempty"`                  // OSPF areas
	Neighbors             []OSPFNeighbor `json:"neighbors,omitempty"`              // OSPF neighbors (NBMA)
	RedistributeStatic    bool           `json:"redistribute_static,omitempty"`    // Redistribute static routes
	RedistributeConnected bool           `json:"redistribute_connected,omitempty"` // Redistribute connected routes
}

// OSPFNetwork represents an OSPF network configuration
type OSPFNetwork struct {
	IP       string `json:"ip"`       // Network IP address or interface
	Wildcard string `json:"wildcard"` // Wildcard mask
	Area     string `json:"area"`     // Area ID (decimal or dotted decimal)
}

// OSPFArea represents an OSPF area configuration
type OSPFArea struct {
	ID        string `json:"id"`                   // Area ID (decimal or dotted decimal)
	Type      string `json:"type,omitempty"`       // Area type: normal, stub, nssa
	NoSummary bool   `json:"no_summary,omitempty"` // Totally stubby/NSSA (no summary LSAs)
}

// OSPFNeighbor represents an OSPF neighbor configuration (for NBMA networks)
type OSPFNeighbor struct {
	IP       string `json:"ip"`                 // Neighbor IP address
	Priority int    `json:"priority,omitempty"` // Neighbor priority (0-255)
	Cost     int    `json:"cost,omitempty"`     // Cost to neighbor
}

// IPsecTunnel represents an IPsec tunnel configuration on an RTX router
type IPsecTunnel struct {
	ID             int            `json:"id"`                     // Tunnel ID
	Name           string         `json:"name,omitempty"`         // Description/name
	LocalAddress   string         `json:"local_address"`          // Local endpoint IP
	RemoteAddress  string         `json:"remote_address"`         // Remote endpoint IP
	PreSharedKey   string         `json:"pre_shared_key"`         // IKE pre-shared key
	IKEv2Proposal  IKEv2Proposal  `json:"ikev2_proposal"`         // IKE Phase 1 proposal
	IPsecTransform IPsecTransform `json:"ipsec_transform"`        // IPsec Phase 2 transform
	LocalNetwork   string         `json:"local_network"`          // Local network CIDR
	RemoteNetwork  string         `json:"remote_network"`         // Remote network CIDR
	DPDEnabled     bool           `json:"dpd_enabled"`            // Dead Peer Detection enabled
	DPDInterval    int            `json:"dpd_interval,omitempty"` // DPD interval in seconds
	DPDRetry       int            `json:"dpd_retry,omitempty"`    // DPD retry count
	Enabled        bool           `json:"enabled"`                // Tunnel enabled
}

// IKEv2Proposal represents IKE Phase 1 proposal settings
type IKEv2Proposal struct {
	EncryptionAES256 bool `json:"encryption_aes256"` // Use AES-256 encryption
	EncryptionAES128 bool `json:"encryption_aes128"` // Use AES-128 encryption
	Encryption3DES   bool `json:"encryption_3des"`   // Use 3DES encryption
	IntegritySHA256  bool `json:"integrity_sha256"`  // Use SHA-256 integrity
	IntegritySHA1    bool `json:"integrity_sha1"`    // Use SHA-1 integrity
	IntegrityMD5     bool `json:"integrity_md5"`     // Use MD5 integrity
	GroupFourteen    bool `json:"group_fourteen"`    // DH group 14 (2048-bit)
	GroupFive        bool `json:"group_five"`        // DH group 5 (1536-bit)
	GroupTwo         bool `json:"group_two"`         // DH group 2 (1024-bit)
	LifetimeSeconds  int  `json:"lifetime_seconds"`  // SA lifetime in seconds
}

// IPsecTransform represents IPsec Phase 2 transform settings
type IPsecTransform struct {
	Protocol         string `json:"protocol"`           // esp or ah
	EncryptionAES256 bool   `json:"encryption_aes256"`  // Use AES-256 encryption
	EncryptionAES128 bool   `json:"encryption_aes128"`  // Use AES-128 encryption
	Encryption3DES   bool   `json:"encryption_3des"`    // Use 3DES encryption
	IntegritySHA256  bool   `json:"integrity_sha256"`   // Use SHA-256-HMAC
	IntegritySHA1    bool   `json:"integrity_sha1"`     // Use SHA-1-HMAC
	IntegrityMD5     bool   `json:"integrity_md5"`      // Use MD5-HMAC
	PFSGroupFourteen bool   `json:"pfs_group_fourteen"` // PFS DH group 14
	PFSGroupFive     bool   `json:"pfs_group_five"`     // PFS DH group 5
	PFSGroupTwo      bool   `json:"pfs_group_two"`      // PFS DH group 2
	LifetimeSeconds  int    `json:"lifetime_seconds"`   // SA lifetime in seconds
}

// L2TPConfig represents L2TP/L2TPv3 configuration on an RTX router
type L2TPConfig struct {
	ID               int            `json:"id"`                          // Tunnel ID
	Name             string         `json:"name,omitempty"`              // Description
	Version          string         `json:"version"`                     // "l2tp" (v2) or "l2tpv3" (v3)
	Mode             string         `json:"mode"`                        // "lns" (L2TPv2 server) or "l2vpn" (L2TPv3)
	Shutdown         bool           `json:"shutdown"`                    // Administratively shut down
	TunnelSource     string         `json:"tunnel_source"`               // Source IP/interface
	TunnelDest       string         `json:"tunnel_dest"`                 // Destination IP/FQDN
	TunnelDestType   string         `json:"tunnel_dest_type,omitempty"`  // "ip" or "fqdn"
	Authentication   *L2TPAuth      `json:"authentication,omitempty"`    // L2TPv2 authentication
	IPPool           *L2TPIPPool    `json:"ip_pool,omitempty"`           // L2TPv2 IP pool
	IPsecProfile     *L2TPIPsec     `json:"ipsec_profile,omitempty"`     // IPsec encryption
	L2TPv3Config     *L2TPv3Config  `json:"l2tpv3_config,omitempty"`     // L2TPv3-specific config
	KeepaliveEnabled bool           `json:"keepalive_enabled,omitempty"` // Keepalive enabled
	KeepaliveConfig  *L2TPKeepalive `json:"keepalive_config,omitempty"`  // Keepalive settings
	DisconnectTime   int            `json:"disconnect_time,omitempty"`   // Idle disconnect time
	AlwaysOn         bool           `json:"always_on,omitempty"`         // Always-on mode
	Enabled          bool           `json:"enabled"`                     // Service enabled
}

// L2TPAuth represents L2TPv2 authentication configuration
type L2TPAuth struct {
	Method   string `json:"method"`             // pap, chap, mschap, mschap-v2
	Username string `json:"username,omitempty"` // Local username
	Password string `json:"password,omitempty"` // Local password
}

// L2TPIPPool represents L2TPv2 IP pool configuration
type L2TPIPPool struct {
	Start string `json:"start"` // Start IP address
	End   string `json:"end"`   // End IP address
}

// L2TPIPsec represents L2TP over IPsec configuration
type L2TPIPsec struct {
	Enabled      bool   `json:"enabled"`                  // IPsec enabled
	PreSharedKey string `json:"pre_shared_key,omitempty"` // IPsec PSK
	TunnelID     int    `json:"tunnel_id,omitempty"`      // Associated IPsec tunnel ID
}

// L2TPv3Config represents L2TPv3-specific configuration
type L2TPv3Config struct {
	LocalRouterID   string          `json:"local_router_id"`            // Local router ID
	RemoteRouterID  string          `json:"remote_router_id"`           // Remote router ID
	RemoteEndID     string          `json:"remote_end_id,omitempty"`    // Remote end ID (hostname)
	SessionID       int             `json:"session_id,omitempty"`       // Session ID
	CookieSize      int             `json:"cookie_size,omitempty"`      // Cookie size (0, 4, 8)
	BridgeInterface string          `json:"bridge_interface,omitempty"` // Bridge interface
	TunnelAuth      *L2TPTunnelAuth `json:"tunnel_auth,omitempty"`      // Tunnel authentication
}

// L2TPTunnelAuth represents L2TPv3 tunnel authentication
type L2TPTunnelAuth struct {
	Enabled  bool   `json:"enabled"`            // Tunnel auth enabled
	Password string `json:"password,omitempty"` // Tunnel auth password
}

// L2TPKeepalive represents L2TP keepalive configuration
type L2TPKeepalive struct {
	Interval int `json:"interval"` // Keepalive interval
	Retry    int `json:"retry"`    // Retry count
}

// L2TPServiceState represents the L2TP service state (singleton)
type L2TPServiceState struct {
	Enabled   bool     `json:"enabled"`             // Service enabled/disabled
	Protocols []string `json:"protocols,omitempty"` // Enabled protocols: "l2tpv3", "l2tp"
}

// PPTPConfig represents PPTP configuration on an RTX router
type PPTPConfig struct {
	Shutdown         bool            `json:"shutdown"`                    // Administratively shut down
	ListenAddress    string          `json:"listen_address,omitempty"`    // Listen IP address
	MaxConnections   int             `json:"max_connections,omitempty"`   // Maximum concurrent connections
	Authentication   *PPTPAuth       `json:"authentication,omitempty"`    // Authentication settings
	Encryption       *PPTPEncryption `json:"encryption,omitempty"`        // MPPE encryption settings
	IPPool           *PPTPIPPool     `json:"ip_pool,omitempty"`           // IP pool for clients
	DisconnectTime   int             `json:"disconnect_time,omitempty"`   // Idle disconnect time
	KeepaliveEnabled bool            `json:"keepalive_enabled,omitempty"` // Keepalive enabled
	Enabled          bool            `json:"enabled"`                     // PPTP service enabled
}

// PPTPAuth represents PPTP authentication configuration
type PPTPAuth struct {
	Method   string `json:"method"`             // pap, chap, mschap, mschap-v2
	Username string `json:"username,omitempty"` // Local username
	Password string `json:"password,omitempty"` // Local password
}

// PPTPEncryption represents PPTP MPPE encryption configuration
type PPTPEncryption struct {
	MPPEBits int  `json:"mppe_bits,omitempty"` // 40, 56, or 128 bits
	Required bool `json:"required,omitempty"`  // Require encryption
}

// PPTPIPPool represents PPTP IP pool configuration
type PPTPIPPool struct {
	Start string `json:"start"` // Start IP address
	End   string `json:"end"`   // End IP address
}

// DNSConfig represents DNS server configuration on an RTX router
type DNSConfig struct {
	DomainLookup bool              `json:"domain_lookup"` // dns domain lookup enable/disable
	DomainName   string            `json:"domain_name"`   // dns domain name
	NameServers  []string          `json:"name_servers"`  // dns server <ip1> [<ip2>]
	ServerSelect []DNSServerSelect `json:"server_select"` // dns server select entries
	Hosts        []DNSHost         `json:"hosts"`         // dns static entries
	ServiceOn    bool              `json:"service_on"`    // dns service on/off
	PrivateSpoof bool              `json:"private_spoof"` // dns private address spoof on/off
}

// DNSServerSelect represents a domain-based DNS server selection entry
type DNSServerSelect struct {
	ID             int      `json:"id"`              // Selector ID (1-65535)
	Servers        []string `json:"servers"`         // DNS server IPs
	EDNS           bool     `json:"edns"`            // Enable EDNS (Extension mechanisms for DNS)
	RecordType     string   `json:"record_type"`     // DNS record type: a, aaaa, ptr, mx, ns, cname, any
	QueryPattern   string   `json:"query_pattern"`   // Domain pattern: ".", "*.example.com", etc.
	OriginalSender string   `json:"original_sender"` // Source IP/CIDR restriction
	RestrictPP     int      `json:"restrict_pp"`     // PP session restriction (0=none)
}

// DNSHost represents a static DNS host entry
type DNSHost struct {
	Name    string `json:"name"`    // Hostname
	Address string `json:"address"` // IP address
}

// ClassMap represents a class-map configuration for traffic classification
type ClassMap struct {
	Name                 string `json:"name"`                              // Class map name
	MatchProtocol        string `json:"match_protocol,omitempty"`          // Protocol to match
	MatchDestinationPort []int  `json:"match_destination_port,omitempty"`  // Destination ports to match
	MatchSourcePort      []int  `json:"match_source_port,omitempty"`       // Source ports to match
	MatchDSCP            string `json:"match_dscp,omitempty"`              // DSCP value to match
	MatchFilter          int    `json:"match_filter,omitempty"`            // IP filter number to match
}

// PolicyMap represents a policy-map configuration
type PolicyMap struct {
	Name    string           `json:"name"`              // Policy map name
	Classes []PolicyMapClass `json:"classes,omitempty"` // Policy map classes
}

// PolicyMapClass represents a class within a policy map
type PolicyMapClass struct {
	Name             string `json:"name"`                        // Class name (references class-map)
	Priority         string `json:"priority,omitempty"`          // Priority level (high, normal, low)
	BandwidthPercent int    `json:"bandwidth_percent,omitempty"` // Bandwidth percentage
	PoliceCIR        int    `json:"police_cir,omitempty"`        // Committed Information Rate in bps
	QueueLimit       int    `json:"queue_limit,omitempty"`       // Queue limit (depth)
}

// ServicePolicy represents a service-policy attachment to an interface
type ServicePolicy struct {
	Interface string `json:"interface"` // Interface name
	Direction string `json:"direction"` // input or output
	PolicyMap string `json:"policy_map"` // Policy map name
}

// ShapeConfig represents traffic shaping configuration
type ShapeConfig struct {
	Interface    string `json:"interface"`             // Interface name
	Direction    string `json:"direction"`             // input or output
	ShapeAverage int    `json:"shape_average"`         // Average rate in bps
	ShapeBurst   int    `json:"shape_burst,omitempty"` // Burst size in bytes
}

// SyslogConfig represents syslog configuration on an RTX router
type SyslogConfig struct {
	Hosts        []SyslogHost `json:"hosts,omitempty"`         // Syslog destination hosts
	LocalAddress string       `json:"local_address,omitempty"` // Source IP address for syslog
	Facility     string       `json:"facility,omitempty"`      // Syslog facility (e.g., user, local0-local7)
	Notice       bool         `json:"notice"`                  // Log notice level messages
	Info         bool         `json:"info"`                    // Log info level messages
	Debug        bool         `json:"debug"`                   // Log debug level messages
}

// SyslogHost represents a syslog destination host
type SyslogHost struct {
	Address string `json:"address"`        // IP address or hostname
	Port    int    `json:"port,omitempty"` // UDP port (default 514)
}

// SNMPConfig represents SNMP configuration on an RTX router
type SNMPConfig struct {
	SysName     string          `json:"sysname,omitempty"`     // System name
	SysLocation string          `json:"syslocation,omitempty"` // System location
	SysContact  string          `json:"syscontact,omitempty"`  // System contact
	Communities []SNMPCommunity `json:"communities,omitempty"` // SNMP communities
	Hosts       []SNMPHost      `json:"hosts,omitempty"`       // SNMP trap hosts
	TrapEnable  []string        `json:"trap_enable,omitempty"` // Enabled trap types
}

// SNMPCommunity represents an SNMP community configuration
type SNMPCommunity struct {
	Name       string `json:"name"`          // Community string name (Sensitive)
	Permission string `json:"permission"`    // "ro" (read-only) or "rw" (read-write)
	ACL        string `json:"acl,omitempty"` // Access control list (optional)
}

// SNMPHost represents an SNMP trap host configuration
type SNMPHost struct {
	Address   string `json:"address"`             // IP address of trap receiver
	Community string `json:"community,omitempty"` // Community string for traps (Sensitive)
	Version   string `json:"version,omitempty"`   // SNMP version (1, 2c)
}

// Schedule represents a schedule configuration on an RTX router
type Schedule struct {
	ID          int      `json:"id"`                     // Schedule ID (1-65535)
	Name        string   `json:"name,omitempty"`         // Schedule name/description
	AtTime      string   `json:"at_time,omitempty"`      // Time in HH:MM format
	DayOfWeek   string   `json:"day_of_week,omitempty"`  // Day(s) of week (e.g., "mon-fri", "sat", "sun,mon")
	Date        string   `json:"date,omitempty"`         // Specific date in YYYY/MM/DD format
	Recurring   bool     `json:"recurring"`              // Whether schedule repeats
	OnStartup   bool     `json:"on_startup"`             // Execute at router startup
	PolicyList  string   `json:"policy_list,omitempty"`  // Policy/command list name
	Commands    []string `json:"commands,omitempty"`     // Commands to execute
	Enabled     bool     `json:"enabled"`                // Whether schedule is enabled
	PPInterface int      `json:"pp_interface,omitempty"` // PP interface number for PP schedules
}

// KronPolicy represents a kron policy (command list) on an RTX router
type KronPolicy struct {
	Name     string   `json:"name"`               // Policy name
	Commands []string `json:"commands,omitempty"` // Commands in the policy
}

// AdminConfig represents the admin configuration on an RTX router
type AdminConfig struct {
	LoginPassword string `json:"login_password"` // Login password (sensitive)
	AdminPassword string `json:"admin_password"` // Administrator password (sensitive)
}

// AdminUser represents a user account on an RTX router
type AdminUser struct {
	Username   string              `json:"username"`   // Username
	Password   string              `json:"password"`   // Password (sensitive)
	Encrypted  bool                `json:"encrypted"`  // Whether password is encrypted
	Attributes AdminUserAttributes `json:"attributes"` // User attributes
}

// AdminUserAttributes represents user attribute configuration
type AdminUserAttributes struct {
	Administrator bool     `json:"administrator"`  // Whether user has administrator privileges
	Connection    []string `json:"connection"`     // Allowed connection types: serial, telnet, remote, ssh, sftp, http
	GUIPages      []string `json:"gui_pages"`      // Allowed GUI pages: dashboard, lan-map, config
	LoginTimer    int      `json:"login_timer"`    // Login timeout in seconds (0 = infinite)
}

// HTTPDConfig represents HTTP daemon configuration on an RTX router
type HTTPDConfig struct {
	Host        string `json:"host"`         // "any" or specific interface (e.g., "lan1")
	ProxyAccess bool   `json:"proxy_access"` // L2MS proxy access enabled
}

// SSHDConfig represents SSH daemon configuration on an RTX router
type SSHDConfig struct {
	Enabled bool     `json:"enabled"`            // sshd service on/off
	Hosts   []string `json:"hosts,omitempty"`    // Interface list (e.g., ["lan1", "lan2"])
	HostKey string   `json:"host_key,omitempty"` // RSA host key (sensitive)
}

// SFTPDConfig represents SFTP daemon configuration on an RTX router
type SFTPDConfig struct {
	Hosts []string `json:"hosts,omitempty"` // Interface list
}

// BridgeConfig represents an Ethernet bridge configuration on an RTX router
type BridgeConfig struct {
	Name    string   `json:"name"`    // Bridge name (bridge1, bridge2, etc.)
	Members []string `json:"members"` // Member interfaces (lan1, tunnel1, etc.)
}

// IPv6InterfaceConfig represents IPv6 configuration for an RTX router interface
type IPv6InterfaceConfig struct {
	Interface        string        `json:"interface"`                    // Interface name (lan1, lan2, pp1, bridge1, tunnel1)
	Addresses        []IPv6Address `json:"addresses,omitempty"`          // IPv6 addresses
	RTADV            *RTADVConfig  `json:"rtadv,omitempty"`              // Router Advertisement configuration
	DHCPv6Service    string        `json:"dhcpv6_service,omitempty"`     // "server", "client", or "off"
	MTU              int           `json:"mtu,omitempty"`                // MTU size (0 = default)
	SecureFilterIn   []int         `json:"secure_filter_in,omitempty"`   // Inbound security filter numbers
	SecureFilterOut  []int         `json:"secure_filter_out,omitempty"`  // Outbound security filter numbers
	DynamicFilterOut []int         `json:"dynamic_filter_out,omitempty"` // Dynamic filters for outbound
}

// IPv6Address represents an IPv6 address configuration
type IPv6Address struct {
	Address     string `json:"address,omitempty"`      // Full IPv6 address with prefix (e.g., "2001:db8::1/64")
	PrefixRef   string `json:"prefix_ref,omitempty"`   // Prefix reference (e.g., "ra-prefix@lan2")
	InterfaceID string `json:"interface_id,omitempty"` // Interface ID (e.g., "::1/64")
}

// RTADVConfig represents Router Advertisement configuration
type RTADVConfig struct {
	Enabled  bool `json:"enabled"`            // RTADV enabled
	PrefixID int  `json:"prefix_id"`          // Prefix ID to advertise
	OFlag    bool `json:"o_flag"`             // Other Configuration Flag (O flag)
	MFlag    bool `json:"m_flag"`             // Managed Address Configuration Flag (M flag)
	Lifetime int  `json:"lifetime,omitempty"` // Router lifetime in seconds
}

// AccessListExtended represents an IPv4 extended access list (Cisco-compatible naming)
type AccessListExtended struct {
	Name    string                     `json:"name"`    // ACL name (identifier)
	Entries []AccessListExtendedEntry  `json:"entries"` // List of ACL entries
}

// AccessListExtendedEntry represents a single entry in an IPv4 extended access list
type AccessListExtendedEntry struct {
	Sequence              int    `json:"sequence"`                         // Sequence number (determines order)
	AceRuleAction         string `json:"ace_rule_action"`                  // permit or deny
	AceRuleProtocol       string `json:"ace_rule_protocol"`                // Protocol: tcp, udp, icmp, ip, etc.
	SourceAny             bool   `json:"source_any,omitempty"`             // Source is any
	SourcePrefix          string `json:"source_prefix,omitempty"`          // Source IP address
	SourcePrefixMask      string `json:"source_prefix_mask,omitempty"`     // Source wildcard mask
	SourcePortEqual       string `json:"source_port_equal,omitempty"`      // Source port equals
	SourcePortRange       string `json:"source_port_range,omitempty"`      // Source port range (e.g., "1024-65535")
	DestinationAny        bool   `json:"destination_any,omitempty"`        // Destination is any
	DestinationPrefix     string `json:"destination_prefix,omitempty"`     // Destination IP address
	DestinationPrefixMask string `json:"destination_prefix_mask,omitempty"` // Destination wildcard mask
	DestinationPortEqual  string `json:"destination_port_equal,omitempty"` // Destination port equals
	DestinationPortRange  string `json:"destination_port_range,omitempty"` // Destination port range
	Established           bool   `json:"established,omitempty"`            // Match established TCP connections
	Log                   bool   `json:"log,omitempty"`                    // Enable logging
}

// AccessListExtendedIPv6 represents an IPv6 extended access list
type AccessListExtendedIPv6 struct {
	Name    string                         `json:"name"`    // ACL name (identifier)
	Entries []AccessListExtendedIPv6Entry  `json:"entries"` // List of ACL entries
}

// AccessListExtendedIPv6Entry represents a single entry in an IPv6 extended access list
type AccessListExtendedIPv6Entry struct {
	Sequence              int    `json:"sequence"`                         // Sequence number (determines order)
	AceRuleAction         string `json:"ace_rule_action"`                  // permit or deny
	AceRuleProtocol       string `json:"ace_rule_protocol"`                // Protocol: tcp, udp, icmpv6, ipv6, etc.
	SourceAny             bool   `json:"source_any,omitempty"`             // Source is any
	SourcePrefix          string `json:"source_prefix,omitempty"`          // Source IPv6 address/prefix
	SourcePrefixLength    int    `json:"source_prefix_length,omitempty"`   // Source prefix length
	SourcePortEqual       string `json:"source_port_equal,omitempty"`      // Source port equals
	SourcePortRange       string `json:"source_port_range,omitempty"`      // Source port range
	DestinationAny        bool   `json:"destination_any,omitempty"`        // Destination is any
	DestinationPrefix     string `json:"destination_prefix,omitempty"`     // Destination IPv6 address/prefix
	DestinationPrefixLength int  `json:"destination_prefix_length,omitempty"` // Destination prefix length
	DestinationPortEqual  string `json:"destination_port_equal,omitempty"` // Destination port equals
	DestinationPortRange  string `json:"destination_port_range,omitempty"` // Destination port range
	Established           bool   `json:"established,omitempty"`            // Match established TCP connections
	Log                   bool   `json:"log,omitempty"`                    // Enable logging
}

// IPFilterDynamicConfig represents a collection of dynamic IP filters (stateful inspection)
type IPFilterDynamicConfig struct {
	Entries []IPFilterDynamicEntry `json:"entries"` // List of dynamic filter entries
}

// IPFilterDynamicEntry represents a single dynamic IP filter entry
type IPFilterDynamicEntry struct {
	Number   int    `json:"number"`            // Filter number (unique identifier)
	Source   string `json:"source"`            // Source address or "*"
	Dest     string `json:"dest"`              // Destination address or "*"
	Protocol string `json:"protocol"`          // Protocol (ftp, www, smtp, etc.)
	Syslog   bool   `json:"syslog,omitempty"`  // Enable syslog for this filter
}

// IPv6FilterDynamicConfig represents a collection of IPv6 dynamic filters
type IPv6FilterDynamicConfig struct {
	Entries []IPv6FilterDynamicEntry `json:"entries"` // List of dynamic filter entries
}

// IPv6FilterDynamicEntry represents a single IPv6 dynamic filter entry
type IPv6FilterDynamicEntry struct {
	Number   int    `json:"number"`            // Filter number (unique identifier)
	Source   string `json:"source"`            // Source address or "*"
	Dest     string `json:"dest"`              // Destination address or "*"
	Protocol string `json:"protocol"`          // Protocol (ftp, www, smtp, etc.)
	Syslog   bool   `json:"syslog,omitempty"`  // Enable syslog for this filter
}

// InterfaceACL represents ACL bindings to an interface
type InterfaceACL struct {
	Interface            string `json:"interface"`                        // Interface name (lan1, pp1, etc.)
	IPAccessGroupIn      string `json:"ip_access_group_in,omitempty"`     // Inbound IPv4 ACL name
	IPAccessGroupOut     string `json:"ip_access_group_out,omitempty"`    // Outbound IPv4 ACL name
	IPv6AccessGroupIn    string `json:"ipv6_access_group_in,omitempty"`   // Inbound IPv6 ACL name
	IPv6AccessGroupOut   string `json:"ipv6_access_group_out,omitempty"`  // Outbound IPv6 ACL name
	DynamicFiltersIn     []int  `json:"dynamic_filters_in,omitempty"`     // Inbound dynamic filter numbers
	DynamicFiltersOut    []int  `json:"dynamic_filters_out,omitempty"`    // Outbound dynamic filter numbers
	IPv6DynamicFiltersIn []int  `json:"ipv6_dynamic_filters_in,omitempty"` // Inbound IPv6 dynamic filter numbers
	IPv6DynamicFiltersOut []int `json:"ipv6_dynamic_filters_out,omitempty"` // Outbound IPv6 dynamic filter numbers
}

// AccessListMAC represents a MAC address access list
type AccessListMAC struct {
	Name    string              `json:"name"`    // ACL name (identifier)
	Entries []AccessListMACEntry `json:"entries"` // List of MAC ACL entries
}

// AccessListMACEntry represents a single entry in a MAC access list
type AccessListMACEntry struct {
	Sequence           int    `json:"sequence"`                      // Sequence number (determines order)
	AceAction          string `json:"ace_action"`                    // permit or deny
	SourceAny          bool   `json:"source_any,omitempty"`          // Source is any
	SourceAddress      string `json:"source_address,omitempty"`      // Source MAC address
	SourceAddressMask  string `json:"source_address_mask,omitempty"` // Source MAC wildcard mask
	DestinationAny     bool   `json:"destination_any,omitempty"`     // Destination is any
	DestinationAddress string `json:"destination_address,omitempty"` // Destination MAC address
	DestinationAddressMask string `json:"destination_address_mask,omitempty"` // Destination MAC wildcard mask
	EtherType          string `json:"ether_type,omitempty"`          // Ethernet type (0x0800, 0x0806, etc.)
	VlanID             int    `json:"vlan_id,omitempty"`             // VLAN ID
	Log                bool   `json:"log,omitempty"`                 // Enable logging
}

// InterfaceMACACL represents MAC ACL bindings to an interface
type InterfaceMACACL struct {
	Interface         string `json:"interface"`                     // Interface name (lan1, lan2, etc.)
	MACAccessGroupIn  string `json:"mac_access_group_in,omitempty"` // Inbound MAC ACL name
	MACAccessGroupOut string `json:"mac_access_group_out,omitempty"` // Outbound MAC ACL name
}

// IPsecTransportConfig represents an IPsec transport mode configuration on an RTX router
type IPsecTransportConfig struct {
	TransportID int    `json:"transport_id"` // Transport number
	TunnelID    int    `json:"tunnel_id"`    // Associated tunnel number
	Protocol    string `json:"protocol"`     // Protocol ("udp" typically)
	Port        int    `json:"port"`         // Port number (1701 for L2TP)
}
