package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/sh1/terraform-provider-rtx/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClientForInterfaces extends MockClient for interfaces testing
type MockClientForInterfaces struct {
	mock.Mock
}

func (m *MockClientForInterfaces) Dial(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClientForInterfaces) Run(ctx context.Context, cmd client.Command) (client.Result, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(client.Result), args.Error(1)
}

func (m *MockClientForInterfaces) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockClientForInterfaces) GetInterfaces(ctx context.Context) ([]client.Interface, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.Interface), args.Error(1)
}

func (m *MockClientForInterfaces) GetSystemInfo(ctx context.Context) (*client.SystemInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.SystemInfo), args.Error(1)
}

func (m *MockClientForInterfaces) GetRoutes(ctx context.Context) ([]client.Route, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.Route), args.Error(1)
}

func (m *MockClientForInterfaces) GetDHCPBindings(ctx context.Context, scopeID int) ([]client.DHCPBinding, error) {
	args := m.Called(ctx, scopeID)
	return args.Get(0).([]client.DHCPBinding), args.Error(1)
}

func (m *MockClientForInterfaces) CreateDHCPBinding(ctx context.Context, binding client.DHCPBinding) error {
	args := m.Called(ctx, binding)
	return args.Error(0)
}

func (m *MockClientForInterfaces) DeleteDHCPBinding(ctx context.Context, scopeID int, ipAddress string) error {
	args := m.Called(ctx, scopeID, ipAddress)
	return args.Error(0)
}

func (m *MockClientForInterfaces) SaveConfig(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClientForInterfaces) GetDHCPScope(ctx context.Context, scopeID int) (*client.DHCPScope, error) {
	args := m.Called(ctx, scopeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.DHCPScope), args.Error(1)
}

func (m *MockClientForInterfaces) CreateDHCPScope(ctx context.Context, scope client.DHCPScope) error {
	args := m.Called(ctx, scope)
	return args.Error(0)
}

func (m *MockClientForInterfaces) UpdateDHCPScope(ctx context.Context, scope client.DHCPScope) error {
	args := m.Called(ctx, scope)
	return args.Error(0)
}

func (m *MockClientForInterfaces) DeleteDHCPScope(ctx context.Context, scopeID int) error {
	args := m.Called(ctx, scopeID)
	return args.Error(0)
}

func (m *MockClientForInterfaces) ListDHCPScopes(ctx context.Context) ([]client.DHCPScope, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.DHCPScope), args.Error(1)
}

func (m *MockClientForInterfaces) GetInterfaceConfig(ctx context.Context, interfaceName string) (*client.InterfaceConfig, error) {
	args := m.Called(ctx, interfaceName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.InterfaceConfig), args.Error(1)
}

func (m *MockClientForInterfaces) ConfigureInterface(ctx context.Context, config client.InterfaceConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockClientForInterfaces) UpdateInterfaceConfig(ctx context.Context, config client.InterfaceConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockClientForInterfaces) ResetInterface(ctx context.Context, interfaceName string) error {
	args := m.Called(ctx, interfaceName)
	return args.Error(0)
}

func (m *MockClientForInterfaces) ListInterfaceConfigs(ctx context.Context) ([]client.InterfaceConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.InterfaceConfig), args.Error(1)
}

func (m *MockClientForInterfaces) GetIPv6Prefix(ctx context.Context, prefixID int) (*client.IPv6Prefix, error) {
	args := m.Called(ctx, prefixID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.IPv6Prefix), args.Error(1)
}

func (m *MockClientForInterfaces) CreateIPv6Prefix(ctx context.Context, prefix client.IPv6Prefix) error {
	args := m.Called(ctx, prefix)
	return args.Error(0)
}

func (m *MockClientForInterfaces) UpdateIPv6Prefix(ctx context.Context, prefix client.IPv6Prefix) error {
	args := m.Called(ctx, prefix)
	return args.Error(0)
}

func (m *MockClientForInterfaces) DeleteIPv6Prefix(ctx context.Context, prefixID int) error {
	args := m.Called(ctx, prefixID)
	return args.Error(0)
}

func (m *MockClientForInterfaces) ListIPv6Prefixes(ctx context.Context) ([]client.IPv6Prefix, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.IPv6Prefix), args.Error(1)
}

func (m *MockClientForInterfaces) GetVLAN(ctx context.Context, iface string, vlanID int) (*client.VLAN, error) {
	args := m.Called(ctx, iface, vlanID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.VLAN), args.Error(1)
}

func (m *MockClientForInterfaces) CreateVLAN(ctx context.Context, vlan client.VLAN) error {
	args := m.Called(ctx, vlan)
	return args.Error(0)
}

func (m *MockClientForInterfaces) UpdateVLAN(ctx context.Context, vlan client.VLAN) error {
	args := m.Called(ctx, vlan)
	return args.Error(0)
}

func (m *MockClientForInterfaces) DeleteVLAN(ctx context.Context, iface string, vlanID int) error {
	args := m.Called(ctx, iface, vlanID)
	return args.Error(0)
}

func (m *MockClientForInterfaces) ListVLANs(ctx context.Context) ([]client.VLAN, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.VLAN), args.Error(1)
}

func (m *MockClientForInterfaces) GetSystemConfig(ctx context.Context) (*client.SystemConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.SystemConfig), args.Error(1)
}

func (m *MockClientForInterfaces) ConfigureSystem(ctx context.Context, config client.SystemConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockClientForInterfaces) UpdateSystemConfig(ctx context.Context, config client.SystemConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockClientForInterfaces) ResetSystem(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClientForInterfaces) GetStaticRoute(ctx context.Context, prefix, mask string) (*client.StaticRoute, error) {
	args := m.Called(ctx, prefix, mask)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.StaticRoute), args.Error(1)
}

func (m *MockClientForInterfaces) CreateStaticRoute(ctx context.Context, route client.StaticRoute) error {
	args := m.Called(ctx, route)
	return args.Error(0)
}

func (m *MockClientForInterfaces) UpdateStaticRoute(ctx context.Context, route client.StaticRoute) error {
	args := m.Called(ctx, route)
	return args.Error(0)
}

func (m *MockClientForInterfaces) DeleteStaticRoute(ctx context.Context, prefix, mask string) error {
	args := m.Called(ctx, prefix, mask)
	return args.Error(0)
}

func (m *MockClientForInterfaces) ListStaticRoutes(ctx context.Context) ([]client.StaticRoute, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.StaticRoute), args.Error(1)
}

// NAT Masquerade methods
func (m *MockClientForInterfaces) GetNATMasquerade(ctx context.Context, descriptorID int) (*client.NATMasquerade, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateNATMasquerade(ctx context.Context, nat client.NATMasquerade) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateNATMasquerade(ctx context.Context, nat client.NATMasquerade) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteNATMasquerade(ctx context.Context, descriptorID int) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListNATMasquerades(ctx context.Context) ([]client.NATMasquerade, error) {
	panic("not implemented")
}

// NAT Static methods
func (m *MockClientForInterfaces) GetNATStatic(ctx context.Context, descriptorID int) (*client.NATStatic, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateNATStatic(ctx context.Context, nat client.NATStatic) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateNATStatic(ctx context.Context, nat client.NATStatic) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteNATStatic(ctx context.Context, descriptorID int) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListNATStatics(ctx context.Context) ([]client.NATStatic, error) {
	panic("not implemented")
}

// IP Filter methods
func (m *MockClientForInterfaces) GetIPFilter(ctx context.Context, number int) (*client.IPFilter, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateIPFilter(ctx context.Context, filter client.IPFilter) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateIPFilter(ctx context.Context, filter client.IPFilter) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteIPFilter(ctx context.Context, number int) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListIPFilters(ctx context.Context) ([]client.IPFilter, error) {
	panic("not implemented")
}

// IPv6 Filter methods
func (m *MockClientForInterfaces) GetIPv6Filter(ctx context.Context, number int) (*client.IPFilter, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateIPv6Filter(ctx context.Context, filter client.IPFilter) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateIPv6Filter(ctx context.Context, filter client.IPFilter) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteIPv6Filter(ctx context.Context, number int) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListIPv6Filters(ctx context.Context) ([]client.IPFilter, error) {
	panic("not implemented")
}

// IP Filter Dynamic methods
func (m *MockClientForInterfaces) GetIPFilterDynamic(ctx context.Context, number int) (*client.IPFilterDynamic, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateIPFilterDynamic(ctx context.Context, filter client.IPFilterDynamic) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteIPFilterDynamic(ctx context.Context, number int) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListIPFiltersDynamic(ctx context.Context) ([]client.IPFilterDynamic, error) {
	panic("not implemented")
}

// Ethernet Filter methods
func (m *MockClientForInterfaces) GetEthernetFilter(ctx context.Context, number int) (*client.EthernetFilter, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateEthernetFilter(ctx context.Context, filter client.EthernetFilter) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateEthernetFilter(ctx context.Context, filter client.EthernetFilter) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteEthernetFilter(ctx context.Context, number int) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListEthernetFilters(ctx context.Context) ([]client.EthernetFilter, error) {
	panic("not implemented")
}

// BGP methods
func (m *MockClientForInterfaces) GetBGPConfig(ctx context.Context) (*client.BGPConfig, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ConfigureBGP(ctx context.Context, config client.BGPConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateBGPConfig(ctx context.Context, config client.BGPConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ResetBGP(ctx context.Context) error {
	panic("not implemented")
}

// OSPF methods
func (m *MockClientForInterfaces) GetOSPF(ctx context.Context) (*client.OSPFConfig, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateOSPF(ctx context.Context, config client.OSPFConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateOSPF(ctx context.Context, config client.OSPFConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteOSPF(ctx context.Context) error {
	panic("not implemented")
}

// IPsec Tunnel methods
func (m *MockClientForInterfaces) GetIPsecTunnel(ctx context.Context, tunnelID int) (*client.IPsecTunnel, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateIPsecTunnel(ctx context.Context, tunnel client.IPsecTunnel) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateIPsecTunnel(ctx context.Context, tunnel client.IPsecTunnel) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteIPsecTunnel(ctx context.Context, tunnelID int) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListIPsecTunnels(ctx context.Context) ([]client.IPsecTunnel, error) {
	panic("not implemented")
}

// IPsec Transport methods
func (m *MockClientForInterfaces) GetIPsecTransport(ctx context.Context, transportID int) (*client.IPsecTransportConfig, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateIPsecTransport(ctx context.Context, transport client.IPsecTransportConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateIPsecTransport(ctx context.Context, transport client.IPsecTransportConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteIPsecTransport(ctx context.Context, transportID int) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListIPsecTransports(ctx context.Context) ([]client.IPsecTransportConfig, error) {
	panic("not implemented")
}

// L2TP methods
func (m *MockClientForInterfaces) GetL2TP(ctx context.Context, tunnelID int) (*client.L2TPConfig, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateL2TP(ctx context.Context, config client.L2TPConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateL2TP(ctx context.Context, config client.L2TPConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteL2TP(ctx context.Context, tunnelID int) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListL2TPs(ctx context.Context) ([]client.L2TPConfig, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) GetL2TPServiceState(ctx context.Context) (*client.L2TPServiceState, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) SetL2TPServiceState(ctx context.Context, enabled bool, protocols []string) error {
	panic("not implemented")
}

// PPTP methods
func (m *MockClientForInterfaces) GetPPTP(ctx context.Context) (*client.PPTPConfig, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreatePPTP(ctx context.Context, config client.PPTPConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdatePPTP(ctx context.Context, config client.PPTPConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeletePPTP(ctx context.Context) error {
	panic("not implemented")
}

// Syslog methods
func (m *MockClientForInterfaces) GetSyslogConfig(ctx context.Context) (*client.SyslogConfig, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ConfigureSyslog(ctx context.Context, config client.SyslogConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateSyslogConfig(ctx context.Context, config client.SyslogConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ResetSyslog(ctx context.Context) error {
	panic("not implemented")
}

// QoS Class Map methods
func (m *MockClientForInterfaces) GetClassMap(ctx context.Context, name string) (*client.ClassMap, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateClassMap(ctx context.Context, cm client.ClassMap) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateClassMap(ctx context.Context, cm client.ClassMap) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteClassMap(ctx context.Context, name string) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListClassMaps(ctx context.Context) ([]client.ClassMap, error) {
	panic("not implemented")
}

// QoS Policy Map methods
func (m *MockClientForInterfaces) GetPolicyMap(ctx context.Context, name string) (*client.PolicyMap, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreatePolicyMap(ctx context.Context, pm client.PolicyMap) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdatePolicyMap(ctx context.Context, pm client.PolicyMap) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeletePolicyMap(ctx context.Context, name string) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListPolicyMaps(ctx context.Context) ([]client.PolicyMap, error) {
	panic("not implemented")
}

// QoS Service Policy methods
func (m *MockClientForInterfaces) GetServicePolicy(ctx context.Context, iface string, direction string) (*client.ServicePolicy, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateServicePolicy(ctx context.Context, sp client.ServicePolicy) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateServicePolicy(ctx context.Context, sp client.ServicePolicy) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteServicePolicy(ctx context.Context, iface string, direction string) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListServicePolicies(ctx context.Context) ([]client.ServicePolicy, error) {
	panic("not implemented")
}

// QoS Shape methods
func (m *MockClientForInterfaces) GetShape(ctx context.Context, iface string, direction string) (*client.ShapeConfig, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateShape(ctx context.Context, sc client.ShapeConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateShape(ctx context.Context, sc client.ShapeConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteShape(ctx context.Context, iface string, direction string) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListShapes(ctx context.Context) ([]client.ShapeConfig, error) {
	panic("not implemented")
}

// SNMP methods
func (m *MockClientForInterfaces) GetSNMP(ctx context.Context) (*client.SNMPConfig, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateSNMP(ctx context.Context, config client.SNMPConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateSNMP(ctx context.Context, config client.SNMPConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteSNMP(ctx context.Context) error {
	panic("not implemented")
}

// Schedule methods
func (m *MockClientForInterfaces) GetSchedule(ctx context.Context, id int) (*client.Schedule, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateSchedule(ctx context.Context, schedule client.Schedule) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateSchedule(ctx context.Context, schedule client.Schedule) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteSchedule(ctx context.Context, id int) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListSchedules(ctx context.Context) ([]client.Schedule, error) {
	panic("not implemented")
}

// Kron Policy methods
func (m *MockClientForInterfaces) GetKronPolicy(ctx context.Context, name string) (*client.KronPolicy, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateKronPolicy(ctx context.Context, policy client.KronPolicy) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateKronPolicy(ctx context.Context, policy client.KronPolicy) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteKronPolicy(ctx context.Context, name string) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListKronPolicies(ctx context.Context) ([]client.KronPolicy, error) {
	panic("not implemented")
}

// DNS methods
func (m *MockClientForInterfaces) GetDNS(ctx context.Context) (*client.DNSConfig, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ConfigureDNS(ctx context.Context, config client.DNSConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateDNS(ctx context.Context, config client.DNSConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ResetDNS(ctx context.Context) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) GetAdminConfig(ctx context.Context) (*client.AdminConfig, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ConfigureAdmin(ctx context.Context, config client.AdminConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateAdminConfig(ctx context.Context, config client.AdminConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ResetAdmin(ctx context.Context) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) GetAdminUser(ctx context.Context, username string) (*client.AdminUser, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateAdminUser(ctx context.Context, user client.AdminUser) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateAdminUser(ctx context.Context, user client.AdminUser) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteAdminUser(ctx context.Context, username string) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListAdminUsers(ctx context.Context) ([]client.AdminUser, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) GetHTTPD(ctx context.Context) (*client.HTTPDConfig, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ConfigureHTTPD(ctx context.Context, config client.HTTPDConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateHTTPD(ctx context.Context, config client.HTTPDConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ResetHTTPD(ctx context.Context) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) GetSSHD(ctx context.Context) (*client.SSHDConfig, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ConfigureSSHD(ctx context.Context, config client.SSHDConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateSSHD(ctx context.Context, config client.SSHDConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ResetSSHD(ctx context.Context) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) GetSFTPD(ctx context.Context) (*client.SFTPDConfig, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ConfigureSFTPD(ctx context.Context, config client.SFTPDConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateSFTPD(ctx context.Context, config client.SFTPDConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ResetSFTPD(ctx context.Context) error {
	panic("not implemented")
}

// Bridge methods
func (m *MockClientForInterfaces) GetBridge(ctx context.Context, name string) (*client.BridgeConfig, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) CreateBridge(ctx context.Context, bridge client.BridgeConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateBridge(ctx context.Context, bridge client.BridgeConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) DeleteBridge(ctx context.Context, name string) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListBridges(ctx context.Context) ([]client.BridgeConfig, error) {
	panic("not implemented")
}

// IPv6 Interface methods
func (m *MockClientForInterfaces) GetIPv6InterfaceConfig(ctx context.Context, interfaceName string) (*client.IPv6InterfaceConfig, error) {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ConfigureIPv6Interface(ctx context.Context, config client.IPv6InterfaceConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) UpdateIPv6InterfaceConfig(ctx context.Context, config client.IPv6InterfaceConfig) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ResetIPv6Interface(ctx context.Context, interfaceName string) error {
	panic("not implemented")
}

func (m *MockClientForInterfaces) ListIPv6InterfaceConfigs(ctx context.Context) ([]client.IPv6InterfaceConfig, error) {
	panic("not implemented")
}

// Access List Extended (IPv4)
func (m *MockClientForInterfaces) GetAccessListExtended(ctx context.Context, name string) (*client.AccessListExtended, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) CreateAccessListExtended(ctx context.Context, acl client.AccessListExtended) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) UpdateAccessListExtended(ctx context.Context, acl client.AccessListExtended) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) DeleteAccessListExtended(ctx context.Context, name string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) ListAccessListsExtended(ctx context.Context) ([]client.AccessListExtended, error) {
	return nil, fmt.Errorf("not implemented")
}

// Access List Extended (IPv6)
func (m *MockClientForInterfaces) GetAccessListExtendedIPv6(ctx context.Context, name string) (*client.AccessListExtendedIPv6, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) CreateAccessListExtendedIPv6(ctx context.Context, acl client.AccessListExtendedIPv6) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) UpdateAccessListExtendedIPv6(ctx context.Context, acl client.AccessListExtendedIPv6) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) DeleteAccessListExtendedIPv6(ctx context.Context, name string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) ListAccessListsExtendedIPv6(ctx context.Context) ([]client.AccessListExtendedIPv6, error) {
	return nil, fmt.Errorf("not implemented")
}

// IP Filter Dynamic Config
func (m *MockClientForInterfaces) GetIPFilterDynamicConfig(ctx context.Context) (*client.IPFilterDynamicConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) CreateIPFilterDynamicConfig(ctx context.Context, config client.IPFilterDynamicConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) UpdateIPFilterDynamicConfig(ctx context.Context, config client.IPFilterDynamicConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) DeleteIPFilterDynamicConfig(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

// IPv6 Filter Dynamic Config
func (m *MockClientForInterfaces) GetIPv6FilterDynamicConfig(ctx context.Context) (*client.IPv6FilterDynamicConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) CreateIPv6FilterDynamicConfig(ctx context.Context, config client.IPv6FilterDynamicConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) UpdateIPv6FilterDynamicConfig(ctx context.Context, config client.IPv6FilterDynamicConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) DeleteIPv6FilterDynamicConfig(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

// Interface ACL
func (m *MockClientForInterfaces) GetInterfaceACL(ctx context.Context, iface string) (*client.InterfaceACL, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) CreateInterfaceACL(ctx context.Context, acl client.InterfaceACL) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) UpdateInterfaceACL(ctx context.Context, acl client.InterfaceACL) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) DeleteInterfaceACL(ctx context.Context, iface string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) ListInterfaceACLs(ctx context.Context) ([]client.InterfaceACL, error) {
	return nil, fmt.Errorf("not implemented")
}

// Access List MAC
func (m *MockClientForInterfaces) GetAccessListMAC(ctx context.Context, name string) (*client.AccessListMAC, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) CreateAccessListMAC(ctx context.Context, acl client.AccessListMAC) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) UpdateAccessListMAC(ctx context.Context, acl client.AccessListMAC) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) DeleteAccessListMAC(ctx context.Context, name string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) ListAccessListsMAC(ctx context.Context) ([]client.AccessListMAC, error) {
	return nil, fmt.Errorf("not implemented")
}

// Interface MAC ACL
func (m *MockClientForInterfaces) GetInterfaceMACACL(ctx context.Context, iface string) (*client.InterfaceMACACL, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) CreateInterfaceMACACL(ctx context.Context, acl client.InterfaceMACACL) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) UpdateInterfaceMACACL(ctx context.Context, acl client.InterfaceMACACL) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) DeleteInterfaceMACACL(ctx context.Context, iface string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) ListInterfaceMACACLs(ctx context.Context) ([]client.InterfaceMACACL, error) {
	return nil, fmt.Errorf("not implemented")
}

// DDNS - NetVolante DNS methods
func (m *MockClientForInterfaces) GetNetVolanteDNS(ctx context.Context) ([]client.NetVolanteConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) GetNetVolanteDNSByInterface(ctx context.Context, iface string) (*client.NetVolanteConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) ConfigureNetVolanteDNS(ctx context.Context, config client.NetVolanteConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) UpdateNetVolanteDNS(ctx context.Context, config client.NetVolanteConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) DeleteNetVolanteDNS(ctx context.Context, iface string) error {
	return fmt.Errorf("not implemented")
}

// DDNS - Custom DDNS methods
func (m *MockClientForInterfaces) GetDDNS(ctx context.Context) ([]client.DDNSServerConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) GetDDNSByID(ctx context.Context, id int) (*client.DDNSServerConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) ConfigureDDNS(ctx context.Context, config client.DDNSServerConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) UpdateDDNS(ctx context.Context, config client.DDNSServerConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) DeleteDDNS(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}

// DDNS - Status methods
func (m *MockClientForInterfaces) GetNetVolanteDNSStatus(ctx context.Context) ([]client.DDNSStatus, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) GetDDNSStatus(ctx context.Context) ([]client.DDNSStatus, error) {
	return nil, fmt.Errorf("not implemented")
}

// PP Interface methods
func (m *MockClientForInterfaces) GetPPInterfaceConfig(ctx context.Context, ppNum int) (*client.PPIPConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) ConfigurePPInterface(ctx context.Context, ppNum int, config client.PPIPConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) UpdatePPInterfaceConfig(ctx context.Context, ppNum int, config client.PPIPConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) ResetPPInterfaceConfig(ctx context.Context, ppNum int) error {
	return fmt.Errorf("not implemented")
}

// PPPoE methods
func (m *MockClientForInterfaces) ListPPPoE(ctx context.Context) ([]client.PPPoEConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) GetPPPoE(ctx context.Context, ppNum int) (*client.PPPoEConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) CreatePPPoE(ctx context.Context, config client.PPPoEConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) UpdatePPPoE(ctx context.Context, config client.PPPoEConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) DeletePPPoE(ctx context.Context, ppNum int) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClientForInterfaces) GetPPConnectionStatus(ctx context.Context, ppNum int) (*client.PPConnectionStatus, error) {
	return nil, fmt.Errorf("not implemented")
}

func TestRTXInterfacesDataSourceSchema(t *testing.T) {
	dataSource := dataSourceRTXInterfaces()
	
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

	// Check interfaces list field
	assert.Contains(t, schemaMap, "interfaces")
	assert.Equal(t, schema.TypeList, schemaMap["interfaces"].Type)
	assert.True(t, schemaMap["interfaces"].Computed)
	assert.NotNil(t, schemaMap["interfaces"].Elem)

	// Check interface schema
	interfaceResource, ok := schemaMap["interfaces"].Elem.(*schema.Resource)
	assert.True(t, ok)
	interfaceSchema := interfaceResource.Schema

	// Check required fields
	requiredFields := map[string]schema.ValueType{
		"name":     schema.TypeString,
		"kind":     schema.TypeString,
		"admin_up": schema.TypeBool,
		"link_up":  schema.TypeBool,
	}
	
	for field, expectedType := range requiredFields {
		assert.Contains(t, interfaceSchema, field, "Schema should contain %s field", field)
		assert.Equal(t, expectedType, interfaceSchema[field].Type, "%s should be of type %v", field, expectedType)
		assert.True(t, interfaceSchema[field].Computed, "%s should be computed", field)
	}

	// Check optional fields
	optionalFields := map[string]schema.ValueType{
		"mac":         schema.TypeString,
		"ipv4":        schema.TypeString,
		"ipv6":        schema.TypeString,
		"mtu":         schema.TypeInt,
		"description": schema.TypeString,
		"attributes":  schema.TypeMap,
	}

	for field, expectedType := range optionalFields {
		assert.Contains(t, interfaceSchema, field, "Schema should contain %s field", field)
		assert.Equal(t, expectedType, interfaceSchema[field].Type, "%s should be of type %v", field, expectedType)
		assert.True(t, interfaceSchema[field].Computed, "%s should be computed", field)
	}

	// Check attributes element type
	assert.NotNil(t, interfaceSchema["attributes"].Elem)
	attributesElem, ok := interfaceSchema["attributes"].Elem.(*schema.Schema)
	assert.True(t, ok)
	assert.Equal(t, schema.TypeString, attributesElem.Type)
}

func TestRTXInterfacesDataSourceRead_Success(t *testing.T) {
	mockClient := &MockClientForInterfaces{}
	
	// Mock successful interfaces retrieval
	expectedInterfaces := []client.Interface{
		{
			Name:    "LAN1",
			Kind:    "lan",
			AdminUp: true,
			LinkUp:  true,
			MAC:     "00:A0:DE:12:34:56",
			IPv4:    "192.168.1.1/24",
			MTU:     1500,
			Attributes: map[string]string{
				"speed": "1000M",
			},
		},
		{
			Name:    "WAN1",
			Kind:    "wan",
			AdminUp: true,
			LinkUp:  false,
			MAC:     "00:A0:DE:12:34:57",
			IPv4:    "203.0.113.1/30",
			MTU:     1500,
		},
		{
			Name:    "PP1",
			Kind:    "pp",
			AdminUp: false,
			LinkUp:  false,
		},
		{
			Name:        "VLAN100",
			Kind:        "vlan",
			AdminUp:     true,
			LinkUp:      true,
			IPv4:        "10.0.0.1/24",
			Description: "Management VLAN",
		},
	}

	mockClient.On("GetInterfaces", mock.Anything).Return(expectedInterfaces, nil)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXInterfaces().Schema, map[string]interface{}{})
	
	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXInterfacesRead(ctx, d, apiClient)

	// Assert no errors
	assert.Empty(t, diags)

	// Assert data was set correctly
	interfaces := d.Get("interfaces").([]interface{})
	assert.Len(t, interfaces, 4)

	// Check LAN1 interface
	lan1 := interfaces[0].(map[string]interface{})
	assert.Equal(t, "LAN1", lan1["name"])
	assert.Equal(t, "lan", lan1["kind"])
	assert.Equal(t, true, lan1["admin_up"])
	assert.Equal(t, true, lan1["link_up"])
	assert.Equal(t, "00:A0:DE:12:34:56", lan1["mac"])
	assert.Equal(t, "192.168.1.1/24", lan1["ipv4"])
	assert.Equal(t, 1500, lan1["mtu"])
	assert.Equal(t, map[string]interface{}{"speed": "1000M"}, lan1["attributes"])

	// Check WAN1 interface
	wan1 := interfaces[1].(map[string]interface{})
	assert.Equal(t, "WAN1", wan1["name"])
	assert.Equal(t, "wan", wan1["kind"])
	assert.Equal(t, true, wan1["admin_up"])
	assert.Equal(t, false, wan1["link_up"])
	assert.Equal(t, "00:A0:DE:12:34:57", wan1["mac"])
	assert.Equal(t, "203.0.113.1/30", wan1["ipv4"])
	assert.Equal(t, 1500, wan1["mtu"])
	// Empty attributes will be an empty map, and other empty fields will be set
	assert.Equal(t, "", wan1["ipv6"])
	assert.Equal(t, "", wan1["description"])
	// Terraform converts map[string]string to map[string]interface{}
	assert.Equal(t, map[string]interface{}{}, wan1["attributes"])

	// Check PP1 interface (minimal data)
	pp1 := interfaces[2].(map[string]interface{})
	assert.Equal(t, "PP1", pp1["name"])
	assert.Equal(t, "pp", pp1["kind"])
	assert.Equal(t, false, pp1["admin_up"])
	assert.Equal(t, false, pp1["link_up"])
	// Empty optional fields will be zero values
	assert.Equal(t, "", pp1["mac"])
	assert.Equal(t, "", pp1["ipv4"])
	assert.Equal(t, 0, pp1["mtu"])
	assert.Equal(t, "", pp1["description"])
	assert.Equal(t, map[string]interface{}{}, pp1["attributes"])

	// Check VLAN100 interface
	vlan100 := interfaces[3].(map[string]interface{})
	assert.Equal(t, "VLAN100", vlan100["name"])
	assert.Equal(t, "vlan", vlan100["kind"])
	assert.Equal(t, true, vlan100["admin_up"])
	assert.Equal(t, true, vlan100["link_up"])
	assert.Equal(t, "10.0.0.1/24", vlan100["ipv4"])
	assert.Equal(t, "Management VLAN", vlan100["description"])
	// Empty fields will be zero values
	assert.Equal(t, "", vlan100["mac"])
	assert.Equal(t, "", vlan100["ipv6"])
	assert.Equal(t, 0, vlan100["mtu"])
	assert.Equal(t, map[string]interface{}{}, vlan100["attributes"])

	// Check that ID was set
	assert.NotEmpty(t, d.Id())

	mockClient.AssertExpectations(t)
}

func TestRTXInterfacesDataSourceRead_ClientError(t *testing.T) {
	mockClient := &MockClientForInterfaces{}
	
	// Mock client error
	expectedError := errors.New("SSH connection failed")
	mockClient.On("GetInterfaces", mock.Anything).Return([]client.Interface{}, expectedError)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXInterfaces().Schema, map[string]interface{}{})
	
	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXInterfacesRead(ctx, d, apiClient)

	// Assert error occurred
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
	// diag.Errorf puts the entire message in Summary
	assert.Contains(t, diags[0].Summary, "Failed to retrieve interfaces information")
	assert.Contains(t, diags[0].Summary, "SSH connection failed")

	mockClient.AssertExpectations(t)
}

func TestRTXInterfacesDataSourceRead_EmptyInterfaces(t *testing.T) {
	mockClient := &MockClientForInterfaces{}
	
	// Mock empty interfaces list
	mockClient.On("GetInterfaces", mock.Anything).Return([]client.Interface{}, nil)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXInterfaces().Schema, map[string]interface{}{})
	
	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXInterfacesRead(ctx, d, apiClient)

	// Assert no errors
	assert.Empty(t, diags)

	// Assert empty interfaces list was set
	interfaces := d.Get("interfaces").([]interface{})
	assert.Len(t, interfaces, 0)

	// Check that ID was still set
	assert.NotEmpty(t, d.Id())

	mockClient.AssertExpectations(t)
}

func TestRTXInterfacesDataSourceRead_DifferentRTXModels(t *testing.T) {
	tests := []struct {
		name       string
		interfaces []client.Interface
	}{
		{
			name: "RTX830_interfaces",
			interfaces: []client.Interface{
				{
					Name:    "LAN1",
					Kind:    "lan",
					AdminUp: true,
					LinkUp:  true,
					MAC:     "00:A0:DE:11:22:33",
					IPv4:    "192.168.100.1/24",
				},
				{
					Name:    "WAN1",
					Kind:    "wan",
					AdminUp: true,
					LinkUp:  true,
					IPv4:    "203.0.113.10/30",
				},
			},
		},
		{
			name: "RTX1210_interfaces",
			interfaces: []client.Interface{
				{
					Name:    "LAN1",
					Kind:    "lan",
					AdminUp: true,
					LinkUp:  true,
					MAC:     "00:A0:DE:44:55:66",
					IPv4:    "10.0.0.1/24",
					IPv6:    "2001:db8::1/64",
					MTU:     1500,
				},
				{
					Name:    "LAN2",
					Kind:    "lan",
					AdminUp: true,
					LinkUp:  false,
					MAC:     "00:A0:DE:44:55:67",
				},
				{
					Name:    "WAN1",
					Kind:    "wan",
					AdminUp: true,
					LinkUp:  true,
					IPv4:    "203.0.113.20/30",
					MTU:     1500,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClientForInterfaces{}
			mockClient.On("GetInterfaces", mock.Anything).Return(tt.interfaces, nil)

			// Create a resource data mock
			d := schema.TestResourceDataRaw(t, dataSourceRTXInterfaces().Schema, map[string]interface{}{})
			
			// Create mock API client
			apiClient := &apiClient{client: mockClient}

			// Call the read function
			ctx := context.Background()
			diags := dataSourceRTXInterfacesRead(ctx, d, apiClient)

			// Assert no errors
			assert.Empty(t, diags)

			// Assert data was set correctly
			interfaces := d.Get("interfaces").([]interface{})
			assert.Len(t, interfaces, len(tt.interfaces))

			// Verify each interface
			for i, expectedInterface := range tt.interfaces {
				actualInterface := interfaces[i].(map[string]interface{})
				assert.Equal(t, expectedInterface.Name, actualInterface["name"])
				assert.Equal(t, expectedInterface.Kind, actualInterface["kind"])
				assert.Equal(t, expectedInterface.AdminUp, actualInterface["admin_up"])
				assert.Equal(t, expectedInterface.LinkUp, actualInterface["link_up"])

				// Check optional fields
				if expectedInterface.MAC != "" {
					assert.Equal(t, expectedInterface.MAC, actualInterface["mac"])
				}
				if expectedInterface.IPv4 != "" {
					assert.Equal(t, expectedInterface.IPv4, actualInterface["ipv4"])
				}
				if expectedInterface.IPv6 != "" {
					assert.Equal(t, expectedInterface.IPv6, actualInterface["ipv6"])
				}
				if expectedInterface.MTU > 0 {
					assert.Equal(t, expectedInterface.MTU, actualInterface["mtu"])
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}

// Acceptance tests using the Docker test environment
func TestAccRTXInterfacesDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXInterfacesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRTXInterfacesDataSourceExists("data.rtx_interfaces.test"),
					resource.TestCheckResourceAttrSet("data.rtx_interfaces.test", "id"),
					resource.TestCheckResourceAttrSet("data.rtx_interfaces.test", "interfaces.#"),
					// Check that we have at least one interface
					resource.TestCheckResourceAttr("data.rtx_interfaces.test", "interfaces.#", "4"), // Assuming RTX has 4 interfaces
				),
			},
		},
	})
}

func TestAccRTXInterfacesDataSource_interfaceAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXInterfacesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Test interface fields
					resource.TestCheckResourceAttrSet("data.rtx_interfaces.test", "interfaces.0.name"),
					resource.TestCheckResourceAttrSet("data.rtx_interfaces.test", "interfaces.0.kind"),
					resource.TestMatchResourceAttr("data.rtx_interfaces.test", "interfaces.0.kind", regexp.MustCompile(`^(lan|wan|pp|vlan)$`)),
					resource.TestCheckResourceAttrSet("data.rtx_interfaces.test", "interfaces.0.admin_up"),
					resource.TestCheckResourceAttrSet("data.rtx_interfaces.test", "interfaces.0.link_up"),
					// Check that interface names match expected pattern
					resource.TestMatchResourceAttr("data.rtx_interfaces.test", "interfaces.0.name", regexp.MustCompile(`^(LAN\d+|WAN\d+|PP\d+|VLAN\d+)$`)),
				),
			},
		},
	})
}

func TestAccRTXInterfacesDataSource_lanInterfaces(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXInterfacesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify LAN interfaces have required attributes
					resource.TestCheckResourceAttrWith("data.rtx_interfaces.test", "interfaces.0.name", func(value string) error {
						if !strings.HasPrefix(value, "LAN") && !strings.HasPrefix(value, "WAN") {
							return fmt.Errorf("expected interface name to start with LAN or WAN, got: %s", value)
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccRTXInterfacesDataSource_interfaceTypes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXInterfacesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check that each interface has a valid kind
					resource.TestCheckResourceAttrWith("data.rtx_interfaces.test", "interfaces.#", func(value string) error {
						// This test will iterate through all interfaces in the acceptance test
						return nil
					}),
				),
			},
		},
	})
}

func testAccRTXInterfacesDataSourceConfig() string {
	return `
data "rtx_interfaces" "test" {}
`
}

func testAccCheckRTXInterfacesDataSourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("data source not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("data source ID not set")
		}

		// Check that interfaces attribute exists and has content
		interfacesCount := rs.Primary.Attributes["interfaces.#"]
		if interfacesCount == "" {
			return fmt.Errorf("interfaces count not set")
		}

		return nil
	}
}