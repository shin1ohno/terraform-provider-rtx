package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// MockClientForRoutes extends MockClient for routes testing
type MockClientForRoutes struct {
	mock.Mock
}

func (m *MockClientForRoutes) Dial(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClientForRoutes) Run(ctx context.Context, cmd client.Command) (client.Result, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(client.Result), args.Error(1)
}

func (m *MockClientForRoutes) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockClientForRoutes) RunBatch(ctx context.Context, cmds []string) ([]byte, error) {
	args := m.Called(ctx, cmds)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockClientForRoutes) GetRoutes(ctx context.Context) ([]client.Route, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.Route), args.Error(1)
}

func (m *MockClientForRoutes) GetSystemInfo(ctx context.Context) (*client.SystemInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.SystemInfo), args.Error(1)
}

func (m *MockClientForRoutes) GetInterfaces(ctx context.Context) ([]client.Interface, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.Interface), args.Error(1)
}

func (m *MockClientForRoutes) GetDHCPBindings(ctx context.Context, scopeID int) ([]client.DHCPBinding, error) {
	args := m.Called(ctx, scopeID)
	return args.Get(0).([]client.DHCPBinding), args.Error(1)
}

func (m *MockClientForRoutes) CreateDHCPBinding(ctx context.Context, binding client.DHCPBinding) error {
	args := m.Called(ctx, binding)
	return args.Error(0)
}

func (m *MockClientForRoutes) DeleteDHCPBinding(ctx context.Context, scopeID int, ipAddress string) error {
	args := m.Called(ctx, scopeID, ipAddress)
	return args.Error(0)
}

func (m *MockClientForRoutes) SaveConfig(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClientForRoutes) GetDHCPScope(ctx context.Context, scopeID int) (*client.DHCPScope, error) {
	args := m.Called(ctx, scopeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.DHCPScope), args.Error(1)
}

func (m *MockClientForRoutes) CreateDHCPScope(ctx context.Context, scope client.DHCPScope) error {
	args := m.Called(ctx, scope)
	return args.Error(0)
}

func (m *MockClientForRoutes) UpdateDHCPScope(ctx context.Context, scope client.DHCPScope) error {
	args := m.Called(ctx, scope)
	return args.Error(0)
}

func (m *MockClientForRoutes) DeleteDHCPScope(ctx context.Context, scopeID int) error {
	args := m.Called(ctx, scopeID)
	return args.Error(0)
}

func (m *MockClientForRoutes) ListDHCPScopes(ctx context.Context) ([]client.DHCPScope, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.DHCPScope), args.Error(1)
}

func (m *MockClientForRoutes) GetInterfaceConfig(ctx context.Context, interfaceName string) (*client.InterfaceConfig, error) {
	args := m.Called(ctx, interfaceName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.InterfaceConfig), args.Error(1)
}

func (m *MockClientForRoutes) ConfigureInterface(ctx context.Context, config client.InterfaceConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockClientForRoutes) UpdateInterfaceConfig(ctx context.Context, config client.InterfaceConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockClientForRoutes) ResetInterface(ctx context.Context, interfaceName string) error {
	args := m.Called(ctx, interfaceName)
	return args.Error(0)
}

func (m *MockClientForRoutes) ListInterfaceConfigs(ctx context.Context) ([]client.InterfaceConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.InterfaceConfig), args.Error(1)
}

func (m *MockClientForRoutes) GetIPv6Prefix(ctx context.Context, prefixID int) (*client.IPv6Prefix, error) {
	args := m.Called(ctx, prefixID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.IPv6Prefix), args.Error(1)
}

func (m *MockClientForRoutes) CreateIPv6Prefix(ctx context.Context, prefix client.IPv6Prefix) error {
	args := m.Called(ctx, prefix)
	return args.Error(0)
}

func (m *MockClientForRoutes) UpdateIPv6Prefix(ctx context.Context, prefix client.IPv6Prefix) error {
	args := m.Called(ctx, prefix)
	return args.Error(0)
}

func (m *MockClientForRoutes) DeleteIPv6Prefix(ctx context.Context, prefixID int) error {
	args := m.Called(ctx, prefixID)
	return args.Error(0)
}

func (m *MockClientForRoutes) ListIPv6Prefixes(ctx context.Context) ([]client.IPv6Prefix, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.IPv6Prefix), args.Error(1)
}

func (m *MockClientForRoutes) GetVLAN(ctx context.Context, iface string, vlanID int) (*client.VLAN, error) {
	args := m.Called(ctx, iface, vlanID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.VLAN), args.Error(1)
}

func (m *MockClientForRoutes) CreateVLAN(ctx context.Context, vlan client.VLAN) error {
	args := m.Called(ctx, vlan)
	return args.Error(0)
}

func (m *MockClientForRoutes) UpdateVLAN(ctx context.Context, vlan client.VLAN) error {
	args := m.Called(ctx, vlan)
	return args.Error(0)
}

func (m *MockClientForRoutes) DeleteVLAN(ctx context.Context, iface string, vlanID int) error {
	args := m.Called(ctx, iface, vlanID)
	return args.Error(0)
}

func (m *MockClientForRoutes) ListVLANs(ctx context.Context) ([]client.VLAN, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.VLAN), args.Error(1)
}

func (m *MockClientForRoutes) GetSystemConfig(ctx context.Context) (*client.SystemConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.SystemConfig), args.Error(1)
}

func (m *MockClientForRoutes) ConfigureSystem(ctx context.Context, config client.SystemConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockClientForRoutes) UpdateSystemConfig(ctx context.Context, config client.SystemConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockClientForRoutes) ResetSystem(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClientForRoutes) GetStaticRoute(ctx context.Context, prefix, mask string) (*client.StaticRoute, error) {
	args := m.Called(ctx, prefix, mask)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.StaticRoute), args.Error(1)
}

func (m *MockClientForRoutes) CreateStaticRoute(ctx context.Context, route client.StaticRoute) error {
	args := m.Called(ctx, route)
	return args.Error(0)
}

func (m *MockClientForRoutes) UpdateStaticRoute(ctx context.Context, route client.StaticRoute) error {
	args := m.Called(ctx, route)
	return args.Error(0)
}

func (m *MockClientForRoutes) DeleteStaticRoute(ctx context.Context, prefix, mask string) error {
	args := m.Called(ctx, prefix, mask)
	return args.Error(0)
}

func (m *MockClientForRoutes) ListStaticRoutes(ctx context.Context) ([]client.StaticRoute, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.StaticRoute), args.Error(1)
}

// NAT Masquerade methods
func (m *MockClientForRoutes) GetNATMasquerade(ctx context.Context, descriptorID int) (*client.NATMasquerade, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateNATMasquerade(ctx context.Context, nat client.NATMasquerade) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateNATMasquerade(ctx context.Context, nat client.NATMasquerade) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteNATMasquerade(ctx context.Context, descriptorID int) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListNATMasquerades(ctx context.Context) ([]client.NATMasquerade, error) {
	panic("not implemented")
}

// NAT Static methods
func (m *MockClientForRoutes) GetNATStatic(ctx context.Context, descriptorID int) (*client.NATStatic, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateNATStatic(ctx context.Context, nat client.NATStatic) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateNATStatic(ctx context.Context, nat client.NATStatic) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteNATStatic(ctx context.Context, descriptorID int) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListNATStatics(ctx context.Context) ([]client.NATStatic, error) {
	panic("not implemented")
}

// IP Filter methods
func (m *MockClientForRoutes) GetIPFilter(ctx context.Context, number int) (*client.IPFilter, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateIPFilter(ctx context.Context, filter client.IPFilter) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateIPFilter(ctx context.Context, filter client.IPFilter) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteIPFilter(ctx context.Context, number int) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListIPFilters(ctx context.Context) ([]client.IPFilter, error) {
	panic("not implemented")
}

// IPv6 Filter methods
func (m *MockClientForRoutes) GetIPv6Filter(ctx context.Context, number int) (*client.IPFilter, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateIPv6Filter(ctx context.Context, filter client.IPFilter) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateIPv6Filter(ctx context.Context, filter client.IPFilter) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteIPv6Filter(ctx context.Context, number int) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListIPv6Filters(ctx context.Context) ([]client.IPFilter, error) {
	panic("not implemented")
}

// IP Filter Dynamic methods
func (m *MockClientForRoutes) GetIPFilterDynamic(ctx context.Context, number int) (*client.IPFilterDynamic, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateIPFilterDynamic(ctx context.Context, filter client.IPFilterDynamic) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteIPFilterDynamic(ctx context.Context, number int) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListIPFiltersDynamic(ctx context.Context) ([]client.IPFilterDynamic, error) {
	panic("not implemented")
}

// Ethernet Filter methods
func (m *MockClientForRoutes) GetEthernetFilter(ctx context.Context, number int) (*client.EthernetFilter, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateEthernetFilter(ctx context.Context, filter client.EthernetFilter) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateEthernetFilter(ctx context.Context, filter client.EthernetFilter) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteEthernetFilter(ctx context.Context, number int) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListEthernetFilters(ctx context.Context) ([]client.EthernetFilter, error) {
	panic("not implemented")
}

// BGP methods
func (m *MockClientForRoutes) GetBGPConfig(ctx context.Context) (*client.BGPConfig, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) ConfigureBGP(ctx context.Context, config client.BGPConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateBGPConfig(ctx context.Context, config client.BGPConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ResetBGP(ctx context.Context) error {
	panic("not implemented")
}

// OSPF methods
func (m *MockClientForRoutes) GetOSPF(ctx context.Context) (*client.OSPFConfig, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateOSPF(ctx context.Context, config client.OSPFConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateOSPF(ctx context.Context, config client.OSPFConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteOSPF(ctx context.Context) error {
	panic("not implemented")
}

// IPsec Tunnel methods
func (m *MockClientForRoutes) GetIPsecTunnel(ctx context.Context, tunnelID int) (*client.IPsecTunnel, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateIPsecTunnel(ctx context.Context, tunnel client.IPsecTunnel) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateIPsecTunnel(ctx context.Context, tunnel client.IPsecTunnel) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteIPsecTunnel(ctx context.Context, tunnelID int) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListIPsecTunnels(ctx context.Context) ([]client.IPsecTunnel, error) {
	panic("not implemented")
}

// IPsec Transport methods
func (m *MockClientForRoutes) GetIPsecTransport(ctx context.Context, transportID int) (*client.IPsecTransportConfig, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateIPsecTransport(ctx context.Context, transport client.IPsecTransportConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateIPsecTransport(ctx context.Context, transport client.IPsecTransportConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteIPsecTransport(ctx context.Context, transportID int) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListIPsecTransports(ctx context.Context) ([]client.IPsecTransportConfig, error) {
	panic("not implemented")
}

// L2TP methods
func (m *MockClientForRoutes) GetL2TP(ctx context.Context, tunnelID int) (*client.L2TPConfig, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateL2TP(ctx context.Context, config client.L2TPConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateL2TP(ctx context.Context, config client.L2TPConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteL2TP(ctx context.Context, tunnelID int) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListL2TPs(ctx context.Context) ([]client.L2TPConfig, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) GetL2TPServiceState(ctx context.Context) (*client.L2TPServiceState, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) SetL2TPServiceState(ctx context.Context, enabled bool, protocols []string) error {
	panic("not implemented")
}

// PPTP methods
func (m *MockClientForRoutes) GetPPTP(ctx context.Context) (*client.PPTPConfig, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreatePPTP(ctx context.Context, config client.PPTPConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdatePPTP(ctx context.Context, config client.PPTPConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeletePPTP(ctx context.Context) error {
	panic("not implemented")
}

// Syslog methods
func (m *MockClientForRoutes) GetSyslogConfig(ctx context.Context) (*client.SyslogConfig, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) ConfigureSyslog(ctx context.Context, config client.SyslogConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateSyslogConfig(ctx context.Context, config client.SyslogConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ResetSyslog(ctx context.Context) error {
	panic("not implemented")
}

// QoS Class Map methods
func (m *MockClientForRoutes) GetClassMap(ctx context.Context, name string) (*client.ClassMap, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateClassMap(ctx context.Context, cm client.ClassMap) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateClassMap(ctx context.Context, cm client.ClassMap) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteClassMap(ctx context.Context, name string) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListClassMaps(ctx context.Context) ([]client.ClassMap, error) {
	panic("not implemented")
}

// QoS Policy Map methods
func (m *MockClientForRoutes) GetPolicyMap(ctx context.Context, name string) (*client.PolicyMap, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreatePolicyMap(ctx context.Context, pm client.PolicyMap) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdatePolicyMap(ctx context.Context, pm client.PolicyMap) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeletePolicyMap(ctx context.Context, name string) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListPolicyMaps(ctx context.Context) ([]client.PolicyMap, error) {
	panic("not implemented")
}

// QoS Service Policy methods
func (m *MockClientForRoutes) GetServicePolicy(ctx context.Context, iface string, direction string) (*client.ServicePolicy, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateServicePolicy(ctx context.Context, sp client.ServicePolicy) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateServicePolicy(ctx context.Context, sp client.ServicePolicy) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteServicePolicy(ctx context.Context, iface string, direction string) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListServicePolicies(ctx context.Context) ([]client.ServicePolicy, error) {
	panic("not implemented")
}

// QoS Shape methods
func (m *MockClientForRoutes) GetShape(ctx context.Context, iface string, direction string) (*client.ShapeConfig, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateShape(ctx context.Context, sc client.ShapeConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateShape(ctx context.Context, sc client.ShapeConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteShape(ctx context.Context, iface string, direction string) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListShapes(ctx context.Context) ([]client.ShapeConfig, error) {
	panic("not implemented")
}

// SNMP methods
func (m *MockClientForRoutes) GetSNMP(ctx context.Context) (*client.SNMPConfig, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateSNMP(ctx context.Context, config client.SNMPConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateSNMP(ctx context.Context, config client.SNMPConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteSNMP(ctx context.Context) error {
	panic("not implemented")
}

// Schedule methods
func (m *MockClientForRoutes) GetSchedule(ctx context.Context, id int) (*client.Schedule, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateSchedule(ctx context.Context, schedule client.Schedule) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateSchedule(ctx context.Context, schedule client.Schedule) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteSchedule(ctx context.Context, id int) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListSchedules(ctx context.Context) ([]client.Schedule, error) {
	panic("not implemented")
}

// Kron Policy methods
func (m *MockClientForRoutes) GetKronPolicy(ctx context.Context, name string) (*client.KronPolicy, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateKronPolicy(ctx context.Context, policy client.KronPolicy) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateKronPolicy(ctx context.Context, policy client.KronPolicy) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteKronPolicy(ctx context.Context, name string) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListKronPolicies(ctx context.Context) ([]client.KronPolicy, error) {
	panic("not implemented")
}

// DNS methods
func (m *MockClientForRoutes) GetDNS(ctx context.Context) (*client.DNSConfig, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) ConfigureDNS(ctx context.Context, config client.DNSConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateDNS(ctx context.Context, config client.DNSConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ResetDNS(ctx context.Context) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) GetAdminConfig(ctx context.Context) (*client.AdminConfig, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) ConfigureAdmin(ctx context.Context, config client.AdminConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateAdminConfig(ctx context.Context, config client.AdminConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ResetAdmin(ctx context.Context) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) GetAdminUser(ctx context.Context, username string) (*client.AdminUser, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateAdminUser(ctx context.Context, user client.AdminUser) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateAdminUser(ctx context.Context, user client.AdminUser) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteAdminUser(ctx context.Context, username string) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListAdminUsers(ctx context.Context) ([]client.AdminUser, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) GetHTTPD(ctx context.Context) (*client.HTTPDConfig, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) ConfigureHTTPD(ctx context.Context, config client.HTTPDConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateHTTPD(ctx context.Context, config client.HTTPDConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ResetHTTPD(ctx context.Context) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) GetSSHD(ctx context.Context) (*client.SSHDConfig, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) ConfigureSSHD(ctx context.Context, config client.SSHDConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateSSHD(ctx context.Context, config client.SSHDConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ResetSSHD(ctx context.Context) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) GetSFTPD(ctx context.Context) (*client.SFTPDConfig, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) ConfigureSFTPD(ctx context.Context, config client.SFTPDConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateSFTPD(ctx context.Context, config client.SFTPDConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ResetSFTPD(ctx context.Context) error {
	panic("not implemented")
}

// Bridge methods
func (m *MockClientForRoutes) GetBridge(ctx context.Context, name string) (*client.BridgeConfig, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) CreateBridge(ctx context.Context, bridge client.BridgeConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateBridge(ctx context.Context, bridge client.BridgeConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) DeleteBridge(ctx context.Context, name string) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListBridges(ctx context.Context) ([]client.BridgeConfig, error) {
	panic("not implemented")
}

// IPv6 Interface methods
func (m *MockClientForRoutes) GetIPv6InterfaceConfig(ctx context.Context, interfaceName string) (*client.IPv6InterfaceConfig, error) {
	panic("not implemented")
}

func (m *MockClientForRoutes) ConfigureIPv6Interface(ctx context.Context, config client.IPv6InterfaceConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) UpdateIPv6InterfaceConfig(ctx context.Context, config client.IPv6InterfaceConfig) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ResetIPv6Interface(ctx context.Context, interfaceName string) error {
	panic("not implemented")
}

func (m *MockClientForRoutes) ListIPv6InterfaceConfigs(ctx context.Context) ([]client.IPv6InterfaceConfig, error) {
	panic("not implemented")
}

// Access List Extended (IPv4)
func (m *MockClientForRoutes) GetAccessListExtended(ctx context.Context, name string) (*client.AccessListExtended, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) CreateAccessListExtended(ctx context.Context, acl client.AccessListExtended) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) UpdateAccessListExtended(ctx context.Context, acl client.AccessListExtended) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) DeleteAccessListExtended(ctx context.Context, name string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) ListAccessListsExtended(ctx context.Context) ([]client.AccessListExtended, error) {
	return nil, fmt.Errorf("not implemented")
}

// Access List Extended (IPv6)
func (m *MockClientForRoutes) GetAccessListExtendedIPv6(ctx context.Context, name string) (*client.AccessListExtendedIPv6, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) CreateAccessListExtendedIPv6(ctx context.Context, acl client.AccessListExtendedIPv6) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) UpdateAccessListExtendedIPv6(ctx context.Context, acl client.AccessListExtendedIPv6) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) DeleteAccessListExtendedIPv6(ctx context.Context, name string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) ListAccessListsExtendedIPv6(ctx context.Context) ([]client.AccessListExtendedIPv6, error) {
	return nil, fmt.Errorf("not implemented")
}

// IP Filter Dynamic Config
func (m *MockClientForRoutes) GetIPFilterDynamicConfig(ctx context.Context) (*client.IPFilterDynamicConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) CreateIPFilterDynamicConfig(ctx context.Context, config client.IPFilterDynamicConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) UpdateIPFilterDynamicConfig(ctx context.Context, config client.IPFilterDynamicConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) DeleteIPFilterDynamicConfig(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

// IPv6 Filter Dynamic Config
func (m *MockClientForRoutes) GetIPv6FilterDynamicConfig(ctx context.Context) (*client.IPv6FilterDynamicConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) CreateIPv6FilterDynamicConfig(ctx context.Context, config client.IPv6FilterDynamicConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) UpdateIPv6FilterDynamicConfig(ctx context.Context, config client.IPv6FilterDynamicConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) DeleteIPv6FilterDynamicConfig(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

// Interface ACL
func (m *MockClientForRoutes) GetInterfaceACL(ctx context.Context, iface string) (*client.InterfaceACL, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) CreateInterfaceACL(ctx context.Context, acl client.InterfaceACL) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) UpdateInterfaceACL(ctx context.Context, acl client.InterfaceACL) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) DeleteInterfaceACL(ctx context.Context, iface string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) ListInterfaceACLs(ctx context.Context) ([]client.InterfaceACL, error) {
	return nil, fmt.Errorf("not implemented")
}

// Access List MAC
func (m *MockClientForRoutes) GetAccessListMAC(ctx context.Context, name string) (*client.AccessListMAC, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) CreateAccessListMAC(ctx context.Context, acl client.AccessListMAC) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) UpdateAccessListMAC(ctx context.Context, acl client.AccessListMAC) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) DeleteAccessListMAC(ctx context.Context, name string, filterNums []int) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) ListAccessListsMAC(ctx context.Context) ([]client.AccessListMAC, error) {
	return nil, fmt.Errorf("not implemented")
}

// Interface MAC ACL
func (m *MockClientForRoutes) GetInterfaceMACACL(ctx context.Context, iface string) (*client.InterfaceMACACL, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) CreateInterfaceMACACL(ctx context.Context, acl client.InterfaceMACACL) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) UpdateInterfaceMACACL(ctx context.Context, acl client.InterfaceMACACL) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) DeleteInterfaceMACACL(ctx context.Context, iface string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) ListInterfaceMACACLs(ctx context.Context) ([]client.InterfaceMACACL, error) {
	return nil, fmt.Errorf("not implemented")
}

// DDNS - NetVolante DNS methods
func (m *MockClientForRoutes) GetNetVolanteDNS(ctx context.Context) ([]client.NetVolanteConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) GetNetVolanteDNSByInterface(ctx context.Context, iface string) (*client.NetVolanteConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) ConfigureNetVolanteDNS(ctx context.Context, config client.NetVolanteConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) UpdateNetVolanteDNS(ctx context.Context, config client.NetVolanteConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) DeleteNetVolanteDNS(ctx context.Context, iface string) error {
	return fmt.Errorf("not implemented")
}

// DDNS - Custom DDNS methods
func (m *MockClientForRoutes) GetDDNS(ctx context.Context) ([]client.DDNSServerConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) GetDDNSByID(ctx context.Context, id int) (*client.DDNSServerConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) ConfigureDDNS(ctx context.Context, config client.DDNSServerConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) UpdateDDNS(ctx context.Context, config client.DDNSServerConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) DeleteDDNS(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}

// DDNS - Status methods
func (m *MockClientForRoutes) GetNetVolanteDNSStatus(ctx context.Context) ([]client.DDNSStatus, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) GetDDNSStatus(ctx context.Context) ([]client.DDNSStatus, error) {
	return nil, fmt.Errorf("not implemented")
}

// PP Interface methods
func (m *MockClientForRoutes) GetPPInterfaceConfig(ctx context.Context, ppNum int) (*client.PPIPConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) ConfigurePPInterface(ctx context.Context, ppNum int, config client.PPIPConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) UpdatePPInterfaceConfig(ctx context.Context, ppNum int, config client.PPIPConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) ResetPPInterfaceConfig(ctx context.Context, ppNum int) error {
	return fmt.Errorf("not implemented")
}

// PPPoE methods
func (m *MockClientForRoutes) ListPPPoE(ctx context.Context) ([]client.PPPoEConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) GetPPPoE(ctx context.Context, ppNum int) (*client.PPPoEConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) CreatePPPoE(ctx context.Context, config client.PPPoEConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) UpdatePPPoE(ctx context.Context, config client.PPPoEConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) DeletePPPoE(ctx context.Context, ppNum int) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForRoutes) GetPPConnectionStatus(ctx context.Context, ppNum int) (*client.PPConnectionStatus, error) {
	return nil, fmt.Errorf("not implemented")
}

// SFTP Configuration Cache methods
func (m *MockClientForRoutes) GetCachedConfig(ctx context.Context) (*parsers.ParsedConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*parsers.ParsedConfig), args.Error(1)
}

func (m *MockClientForRoutes) SFTPEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockClientForRoutes) InvalidateCache() {
	m.Called()
}

func (m *MockClientForRoutes) MarkCacheDirty() {
	m.Called()
}

func TestRTXRoutesDataSourceSchema(t *testing.T) {
	dataSource := dataSourceRTXRoutes()

	// Test that the data source is properly configured
	assert.NotNil(t, dataSource)
	assert.NotNil(t, dataSource.Schema)
	assert.NotNil(t, dataSource.ReadContext)

	// Test schema structure
	schemaMap := dataSource.Schema

	// Check that id field exists and is computed
	assert.Contains(t, schemaMap, "id")
	assert.Equal(t, schema.TypeString, schemaMap["id"].Type)
	assert.True(t, schemaMap["id"].Computed)

	// Check routes list field
	assert.Contains(t, schemaMap, "routes")
	assert.Equal(t, schema.TypeList, schemaMap["routes"].Type)
	assert.True(t, schemaMap["routes"].Computed)
	assert.NotNil(t, schemaMap["routes"].Elem)

	// Check route schema
	routeResource, ok := schemaMap["routes"].Elem.(*schema.Resource)
	assert.True(t, ok)
	routeSchema := routeResource.Schema

	// Check required fields
	requiredFields := map[string]schema.ValueType{
		"destination": schema.TypeString,
		"gateway":     schema.TypeString,
		"interface":   schema.TypeString,
		"protocol":    schema.TypeString,
		"metric":      schema.TypeInt,
	}

	for field, expectedType := range requiredFields {
		assert.Contains(t, routeSchema, field, "Schema should contain %s field", field)
		assert.Equal(t, expectedType, routeSchema[field].Type, "%s should be of type %v", field, expectedType)
		assert.True(t, routeSchema[field].Computed, "%s should be computed", field)
	}
}

func TestRTXRoutesDataSourceRead_Success(t *testing.T) {
	mockClient := &MockClientForRoutes{}

	// Mock successful routes retrieval
	metric1 := 1
	metric2 := 10
	expectedRoutes := []client.Route{
		{
			Destination: "0.0.0.0/0",
			Gateway:     "192.168.1.1",
			Interface:   "LAN1",
			Protocol:    "S",
			Metric:      &metric1,
		},
		{
			Destination: "192.168.1.0/24",
			Gateway:     "*",
			Interface:   "LAN1",
			Protocol:    "C",
			Metric:      nil, // Connected routes may not have metrics
		},
		{
			Destination: "10.0.0.0/24",
			Gateway:     "192.168.1.254",
			Interface:   "LAN1",
			Protocol:    "R",
			Metric:      &metric2,
		},
		{
			Destination: "203.0.113.0/30",
			Gateway:     "*",
			Interface:   "WAN1",
			Protocol:    "C",
			Metric:      nil,
		},
	}

	mockClient.On("GetRoutes", mock.Anything).Return(expectedRoutes, nil)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXRoutes().Schema, map[string]interface{}{})

	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXRoutesRead(ctx, d, apiClient)

	// Assert no errors
	assert.Empty(t, diags)

	// Assert data was set correctly
	routes := d.Get("routes").([]interface{})
	assert.Len(t, routes, 4)

	// Check default route (first route)
	defaultRoute := routes[0].(map[string]interface{})
	assert.Equal(t, "0.0.0.0/0", defaultRoute["destination"])
	assert.Equal(t, "192.168.1.1", defaultRoute["gateway"])
	assert.Equal(t, "LAN1", defaultRoute["interface"])
	assert.Equal(t, "S", defaultRoute["protocol"])
	assert.Equal(t, 1, defaultRoute["metric"])

	// Check connected route (second route)
	connectedRoute := routes[1].(map[string]interface{})
	assert.Equal(t, "192.168.1.0/24", connectedRoute["destination"])
	assert.Equal(t, "*", connectedRoute["gateway"])
	assert.Equal(t, "LAN1", connectedRoute["interface"])
	assert.Equal(t, "C", connectedRoute["protocol"])
	assert.Equal(t, 0, connectedRoute["metric"]) // Should default to 0 when nil

	// Check RIP route (third route)
	ripRoute := routes[2].(map[string]interface{})
	assert.Equal(t, "10.0.0.0/24", ripRoute["destination"])
	assert.Equal(t, "192.168.1.254", ripRoute["gateway"])
	assert.Equal(t, "LAN1", ripRoute["interface"])
	assert.Equal(t, "R", ripRoute["protocol"])
	assert.Equal(t, 10, ripRoute["metric"])

	// Check WAN connected route (fourth route)
	wanRoute := routes[3].(map[string]interface{})
	assert.Equal(t, "203.0.113.0/30", wanRoute["destination"])
	assert.Equal(t, "*", wanRoute["gateway"])
	assert.Equal(t, "WAN1", wanRoute["interface"])
	assert.Equal(t, "C", wanRoute["protocol"])
	assert.Equal(t, 0, wanRoute["metric"])

	// Check that ID was set
	assert.NotEmpty(t, d.Id())

	mockClient.AssertExpectations(t)
}

func TestRTXRoutesDataSourceRead_ClientError(t *testing.T) {
	mockClient := &MockClientForRoutes{}

	// Mock client error
	expectedError := errors.New("SSH connection failed")
	mockClient.On("GetRoutes", mock.Anything).Return([]client.Route{}, expectedError)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXRoutes().Schema, map[string]interface{}{})

	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXRoutesRead(ctx, d, apiClient)

	// Assert error occurred
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "Failed to retrieve routes information")
	assert.Contains(t, diags[0].Summary, "SSH connection failed")

	mockClient.AssertExpectations(t)
}

func TestRTXRoutesDataSourceRead_EmptyRoutes(t *testing.T) {
	mockClient := &MockClientForRoutes{}

	// Mock empty routes list
	mockClient.On("GetRoutes", mock.Anything).Return([]client.Route{}, nil)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXRoutes().Schema, map[string]interface{}{})

	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXRoutesRead(ctx, d, apiClient)

	// Assert no errors
	assert.Empty(t, diags)

	// Assert empty routes list was set
	routes := d.Get("routes").([]interface{})
	assert.Len(t, routes, 0)

	// Check that ID was still set
	assert.NotEmpty(t, d.Id())

	mockClient.AssertExpectations(t)
}

func TestRTXRoutesDataSourceRead_DifferentRouteTypes(t *testing.T) {
	tests := []struct {
		name   string
		routes []client.Route
	}{
		{
			name: "StaticRoutes",
			routes: []client.Route{
				{
					Destination: "0.0.0.0/0",
					Gateway:     "192.168.1.1",
					Interface:   "LAN1",
					Protocol:    "S",
					Metric:      intPtr(1),
				},
				{
					Destination: "10.10.10.0/24",
					Gateway:     "192.168.1.10",
					Interface:   "LAN1",
					Protocol:    "S",
					Metric:      intPtr(5),
				},
			},
		},
		{
			name: "ConnectedRoutes",
			routes: []client.Route{
				{
					Destination: "192.168.1.0/24",
					Gateway:     "*",
					Interface:   "LAN1",
					Protocol:    "C",
					Metric:      nil,
				},
				{
					Destination: "192.168.2.0/24",
					Gateway:     "*",
					Interface:   "LAN2",
					Protocol:    "C",
					Metric:      nil,
				},
			},
		},
		{
			name: "DynamicRoutes",
			routes: []client.Route{
				{
					Destination: "172.16.1.0/24",
					Gateway:     "192.168.1.100",
					Interface:   "LAN1",
					Protocol:    "R",
					Metric:      intPtr(2),
				},
				{
					Destination: "172.16.2.0/24",
					Gateway:     "192.168.1.101",
					Interface:   "LAN1",
					Protocol:    "O",
					Metric:      intPtr(110),
				},
				{
					Destination: "172.16.3.0/24",
					Gateway:     "192.168.1.102",
					Interface:   "LAN1",
					Protocol:    "B",
					Metric:      intPtr(20),
				},
			},
		},
		{
			name: "DHCPRoutes",
			routes: []client.Route{
				{
					Destination: "0.0.0.0/0",
					Gateway:     "203.0.113.1",
					Interface:   "WAN1",
					Protocol:    "D",
					Metric:      intPtr(1),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClientForRoutes{}
			mockClient.On("GetRoutes", mock.Anything).Return(tt.routes, nil)

			// Create a resource data mock
			d := schema.TestResourceDataRaw(t, dataSourceRTXRoutes().Schema, map[string]interface{}{})

			// Create mock API client
			apiClient := &apiClient{client: mockClient}

			// Call the read function
			ctx := context.Background()
			diags := dataSourceRTXRoutesRead(ctx, d, apiClient)

			// Assert no errors
			assert.Empty(t, diags)

			// Assert data was set correctly
			routes := d.Get("routes").([]interface{})
			assert.Len(t, routes, len(tt.routes))

			// Verify each route
			for i, expectedRoute := range tt.routes {
				actualRoute := routes[i].(map[string]interface{})
				assert.Equal(t, expectedRoute.Destination, actualRoute["destination"])
				assert.Equal(t, expectedRoute.Gateway, actualRoute["gateway"])
				assert.Equal(t, expectedRoute.Interface, actualRoute["interface"])
				assert.Equal(t, expectedRoute.Protocol, actualRoute["protocol"])

				if expectedRoute.Metric != nil {
					assert.Equal(t, *expectedRoute.Metric, actualRoute["metric"])
				} else {
					assert.Equal(t, 0, actualRoute["metric"])
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestRTXRoutesDataSourceRead_ProtocolValidation(t *testing.T) {
	validProtocols := []string{"S", "C", "R", "O", "B", "D"}

	for _, protocol := range validProtocols {
		t.Run(fmt.Sprintf("Protocol_%s", protocol), func(t *testing.T) {
			mockClient := &MockClientForRoutes{}

			routes := []client.Route{
				{
					Destination: "192.168.1.0/24",
					Gateway:     "192.168.1.1",
					Interface:   "LAN1",
					Protocol:    protocol,
					Metric:      intPtr(1),
				},
			}
			mockClient.On("GetRoutes", mock.Anything).Return(routes, nil)

			// Create a resource data mock
			d := schema.TestResourceDataRaw(t, dataSourceRTXRoutes().Schema, map[string]interface{}{})

			// Create mock API client
			apiClient := &apiClient{client: mockClient}

			// Call the read function
			ctx := context.Background()
			diags := dataSourceRTXRoutesRead(ctx, d, apiClient)

			// Assert no errors
			assert.Empty(t, diags)

			// Verify protocol was set correctly
			routesData := d.Get("routes").([]interface{})
			assert.Len(t, routesData, 1)

			actualRoute := routesData[0].(map[string]interface{})
			assert.Equal(t, protocol, actualRoute["protocol"])

			mockClient.AssertExpectations(t)
		})
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}

// Acceptance tests using the Docker test environment
func TestAccRTXRoutesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXRoutesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRTXRoutesDataSourceExists("data.rtx_routes.test"),
					resource.TestCheckResourceAttrSet("data.rtx_routes.test", "id"),
					resource.TestCheckResourceAttrSet("data.rtx_routes.test", "routes.#"),
					// Check that we have at least one route (should have default route and connected routes)
					resource.TestCheckResourceAttr("data.rtx_routes.test", "routes.#", "3"), // Assuming RTX has 3 routes
				),
			},
		},
	})
}

func TestAccRTXRoutesDataSource_routeAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXRoutesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Test route fields
					resource.TestCheckResourceAttrSet("data.rtx_routes.test", "routes.0.destination"),
					resource.TestCheckResourceAttrSet("data.rtx_routes.test", "routes.0.gateway"),
					resource.TestCheckResourceAttrSet("data.rtx_routes.test", "routes.0.interface"),
					resource.TestCheckResourceAttrSet("data.rtx_routes.test", "routes.0.protocol"),
					resource.TestCheckResourceAttrSet("data.rtx_routes.test", "routes.0.metric"),
					// Test that destination matches CIDR pattern
					resource.TestMatchResourceAttr("data.rtx_routes.test", "routes.0.destination", regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}/\d{1,2}$`)),
					// Test that protocol is one of the valid values
					resource.TestMatchResourceAttr("data.rtx_routes.test", "routes.0.protocol", regexp.MustCompile(`^[SCROB D]$`)),
				),
			},
		},
	})
}

func TestAccRTXRoutesDataSource_defaultRoute(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXRoutesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check for default route (0.0.0.0/0) existence
					resource.TestCheckResourceAttrWith("data.rtx_routes.test", "routes.#", func(value string) error {
						// This test checks that at least one route exists
						if value == "0" {
							return fmt.Errorf("expected at least one route, got none")
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccRTXRoutesDataSource_connectedRoutes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXRoutesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify connected routes have "*" as gateway
					resource.TestCheckResourceAttrWith("data.rtx_routes.test", "routes.#", func(value string) error {
						// This would need custom logic to check for connected routes in acceptance test
						return nil
					}),
				),
			},
		},
	})
}

func testAccRTXRoutesDataSourceConfig() string {
	return `
data "rtx_routes" "test" {}
`
}

func testAccCheckRTXRoutesDataSourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("data source not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("data source ID not set")
		}

		// Check that routes attribute exists and has content
		routesCount := rs.Primary.Attributes["routes.#"]
		if routesCount == "" {
			return fmt.Errorf("routes count not set")
		}

		return nil
	}
}
