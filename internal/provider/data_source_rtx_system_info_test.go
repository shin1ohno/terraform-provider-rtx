package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
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

// Mock client for unit tests
type MockClient struct {
	mock.Mock
}

func (m *MockClient) Dial(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClient) Run(ctx context.Context, cmd client.Command) (client.Result, error) {
	args := m.Called(ctx, cmd)
	return args.Get(0).(client.Result), args.Error(1)
}

func (m *MockClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockClient) GetInterfaces(ctx context.Context) ([]client.Interface, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.Interface), args.Error(1)
}

func (m *MockClient) GetSystemInfo(ctx context.Context) (*client.SystemInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.SystemInfo), args.Error(1)
}

func (m *MockClient) GetRoutes(ctx context.Context) ([]client.Route, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.Route), args.Error(1)
}

func (m *MockClient) GetDHCPBindings(ctx context.Context, scopeID int) ([]client.DHCPBinding, error) {
	args := m.Called(ctx, scopeID)
	return args.Get(0).([]client.DHCPBinding), args.Error(1)
}

func (m *MockClient) CreateDHCPBinding(ctx context.Context, binding client.DHCPBinding) error {
	args := m.Called(ctx, binding)
	return args.Error(0)
}

func (m *MockClient) DeleteDHCPBinding(ctx context.Context, scopeID int, ipAddress string) error {
	args := m.Called(ctx, scopeID, ipAddress)
	return args.Error(0)
}

func (m *MockClient) SaveConfig(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClient) GetDHCPScope(ctx context.Context, scopeID int) (*client.DHCPScope, error) {
	args := m.Called(ctx, scopeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.DHCPScope), args.Error(1)
}

func (m *MockClient) CreateDHCPScope(ctx context.Context, scope client.DHCPScope) error {
	args := m.Called(ctx, scope)
	return args.Error(0)
}

func (m *MockClient) UpdateDHCPScope(ctx context.Context, scope client.DHCPScope) error {
	args := m.Called(ctx, scope)
	return args.Error(0)
}

func (m *MockClient) DeleteDHCPScope(ctx context.Context, scopeID int) error {
	args := m.Called(ctx, scopeID)
	return args.Error(0)
}

func (m *MockClient) ListDHCPScopes(ctx context.Context) ([]client.DHCPScope, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.DHCPScope), args.Error(1)
}

func (m *MockClient) GetInterfaceConfig(ctx context.Context, interfaceName string) (*client.InterfaceConfig, error) {
	args := m.Called(ctx, interfaceName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.InterfaceConfig), args.Error(1)
}

func (m *MockClient) ConfigureInterface(ctx context.Context, config client.InterfaceConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockClient) UpdateInterfaceConfig(ctx context.Context, config client.InterfaceConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockClient) ResetInterface(ctx context.Context, interfaceName string) error {
	args := m.Called(ctx, interfaceName)
	return args.Error(0)
}

func (m *MockClient) ListInterfaceConfigs(ctx context.Context) ([]client.InterfaceConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.InterfaceConfig), args.Error(1)
}

func (m *MockClient) GetIPv6Prefix(ctx context.Context, prefixID int) (*client.IPv6Prefix, error) {
	args := m.Called(ctx, prefixID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.IPv6Prefix), args.Error(1)
}

func (m *MockClient) CreateIPv6Prefix(ctx context.Context, prefix client.IPv6Prefix) error {
	args := m.Called(ctx, prefix)
	return args.Error(0)
}

func (m *MockClient) UpdateIPv6Prefix(ctx context.Context, prefix client.IPv6Prefix) error {
	args := m.Called(ctx, prefix)
	return args.Error(0)
}

func (m *MockClient) DeleteIPv6Prefix(ctx context.Context, prefixID int) error {
	args := m.Called(ctx, prefixID)
	return args.Error(0)
}

func (m *MockClient) ListIPv6Prefixes(ctx context.Context) ([]client.IPv6Prefix, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.IPv6Prefix), args.Error(1)
}

func (m *MockClient) GetVLAN(ctx context.Context, iface string, vlanID int) (*client.VLAN, error) {
	args := m.Called(ctx, iface, vlanID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.VLAN), args.Error(1)
}

func (m *MockClient) CreateVLAN(ctx context.Context, vlan client.VLAN) error {
	args := m.Called(ctx, vlan)
	return args.Error(0)
}

func (m *MockClient) UpdateVLAN(ctx context.Context, vlan client.VLAN) error {
	args := m.Called(ctx, vlan)
	return args.Error(0)
}

func (m *MockClient) DeleteVLAN(ctx context.Context, iface string, vlanID int) error {
	args := m.Called(ctx, iface, vlanID)
	return args.Error(0)
}

func (m *MockClient) ListVLANs(ctx context.Context) ([]client.VLAN, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.VLAN), args.Error(1)
}

func (m *MockClient) GetSystemConfig(ctx context.Context) (*client.SystemConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.SystemConfig), args.Error(1)
}

func (m *MockClient) ConfigureSystem(ctx context.Context, config client.SystemConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockClient) UpdateSystemConfig(ctx context.Context, config client.SystemConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockClient) ResetSystem(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockClient) GetStaticRoute(ctx context.Context, prefix, mask string) (*client.StaticRoute, error) {
	args := m.Called(ctx, prefix, mask)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.StaticRoute), args.Error(1)
}

func (m *MockClient) CreateStaticRoute(ctx context.Context, route client.StaticRoute) error {
	args := m.Called(ctx, route)
	return args.Error(0)
}

func (m *MockClient) UpdateStaticRoute(ctx context.Context, route client.StaticRoute) error {
	args := m.Called(ctx, route)
	return args.Error(0)
}

func (m *MockClient) DeleteStaticRoute(ctx context.Context, prefix, mask string) error {
	args := m.Called(ctx, prefix, mask)
	return args.Error(0)
}

func (m *MockClient) ListStaticRoutes(ctx context.Context) ([]client.StaticRoute, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]client.StaticRoute), args.Error(1)
}

// NAT Masquerade methods
func (m *MockClient) GetNATMasquerade(ctx context.Context, descriptorID int) (*client.NATMasquerade, error) {
	panic("not implemented")
}

func (m *MockClient) CreateNATMasquerade(ctx context.Context, nat client.NATMasquerade) error {
	panic("not implemented")
}

func (m *MockClient) UpdateNATMasquerade(ctx context.Context, nat client.NATMasquerade) error {
	panic("not implemented")
}

func (m *MockClient) DeleteNATMasquerade(ctx context.Context, descriptorID int) error {
	panic("not implemented")
}

func (m *MockClient) ListNATMasquerades(ctx context.Context) ([]client.NATMasquerade, error) {
	panic("not implemented")
}

// NAT Static methods
func (m *MockClient) GetNATStatic(ctx context.Context, descriptorID int) (*client.NATStatic, error) {
	panic("not implemented")
}

func (m *MockClient) CreateNATStatic(ctx context.Context, nat client.NATStatic) error {
	panic("not implemented")
}

func (m *MockClient) UpdateNATStatic(ctx context.Context, nat client.NATStatic) error {
	panic("not implemented")
}

func (m *MockClient) DeleteNATStatic(ctx context.Context, descriptorID int) error {
	panic("not implemented")
}

func (m *MockClient) ListNATStatics(ctx context.Context) ([]client.NATStatic, error) {
	panic("not implemented")
}

// IP Filter methods
func (m *MockClient) GetIPFilter(ctx context.Context, number int) (*client.IPFilter, error) {
	panic("not implemented")
}

func (m *MockClient) CreateIPFilter(ctx context.Context, filter client.IPFilter) error {
	panic("not implemented")
}

func (m *MockClient) UpdateIPFilter(ctx context.Context, filter client.IPFilter) error {
	panic("not implemented")
}

func (m *MockClient) DeleteIPFilter(ctx context.Context, number int) error {
	panic("not implemented")
}

func (m *MockClient) ListIPFilters(ctx context.Context) ([]client.IPFilter, error) {
	panic("not implemented")
}

// IPv6 Filter methods
func (m *MockClient) GetIPv6Filter(ctx context.Context, number int) (*client.IPFilter, error) {
	panic("not implemented")
}

func (m *MockClient) CreateIPv6Filter(ctx context.Context, filter client.IPFilter) error {
	panic("not implemented")
}

func (m *MockClient) UpdateIPv6Filter(ctx context.Context, filter client.IPFilter) error {
	panic("not implemented")
}

func (m *MockClient) DeleteIPv6Filter(ctx context.Context, number int) error {
	panic("not implemented")
}

func (m *MockClient) ListIPv6Filters(ctx context.Context) ([]client.IPFilter, error) {
	panic("not implemented")
}

// IP Filter Dynamic methods
func (m *MockClient) GetIPFilterDynamic(ctx context.Context, number int) (*client.IPFilterDynamic, error) {
	panic("not implemented")
}

func (m *MockClient) CreateIPFilterDynamic(ctx context.Context, filter client.IPFilterDynamic) error {
	panic("not implemented")
}

func (m *MockClient) DeleteIPFilterDynamic(ctx context.Context, number int) error {
	panic("not implemented")
}

func (m *MockClient) ListIPFiltersDynamic(ctx context.Context) ([]client.IPFilterDynamic, error) {
	panic("not implemented")
}

// Ethernet Filter methods
func (m *MockClient) GetEthernetFilter(ctx context.Context, number int) (*client.EthernetFilter, error) {
	panic("not implemented")
}

func (m *MockClient) CreateEthernetFilter(ctx context.Context, filter client.EthernetFilter) error {
	panic("not implemented")
}

func (m *MockClient) UpdateEthernetFilter(ctx context.Context, filter client.EthernetFilter) error {
	panic("not implemented")
}

func (m *MockClient) DeleteEthernetFilter(ctx context.Context, number int) error {
	panic("not implemented")
}

func (m *MockClient) ListEthernetFilters(ctx context.Context) ([]client.EthernetFilter, error) {
	panic("not implemented")
}

// BGP methods
func (m *MockClient) GetBGPConfig(ctx context.Context) (*client.BGPConfig, error) {
	panic("not implemented")
}

func (m *MockClient) ConfigureBGP(ctx context.Context, config client.BGPConfig) error {
	panic("not implemented")
}

func (m *MockClient) UpdateBGPConfig(ctx context.Context, config client.BGPConfig) error {
	panic("not implemented")
}

func (m *MockClient) ResetBGP(ctx context.Context) error {
	panic("not implemented")
}

// OSPF methods
func (m *MockClient) GetOSPF(ctx context.Context) (*client.OSPFConfig, error) {
	panic("not implemented")
}

func (m *MockClient) CreateOSPF(ctx context.Context, config client.OSPFConfig) error {
	panic("not implemented")
}

func (m *MockClient) UpdateOSPF(ctx context.Context, config client.OSPFConfig) error {
	panic("not implemented")
}

func (m *MockClient) DeleteOSPF(ctx context.Context) error {
	panic("not implemented")
}

// IPsec Tunnel methods
func (m *MockClient) GetIPsecTunnel(ctx context.Context, tunnelID int) (*client.IPsecTunnel, error) {
	panic("not implemented")
}

func (m *MockClient) CreateIPsecTunnel(ctx context.Context, tunnel client.IPsecTunnel) error {
	panic("not implemented")
}

func (m *MockClient) UpdateIPsecTunnel(ctx context.Context, tunnel client.IPsecTunnel) error {
	panic("not implemented")
}

func (m *MockClient) DeleteIPsecTunnel(ctx context.Context, tunnelID int) error {
	panic("not implemented")
}

func (m *MockClient) ListIPsecTunnels(ctx context.Context) ([]client.IPsecTunnel, error) {
	panic("not implemented")
}

// IPsec Transport methods
func (m *MockClient) GetIPsecTransport(ctx context.Context, transportID int) (*client.IPsecTransportConfig, error) {
	panic("not implemented")
}

func (m *MockClient) CreateIPsecTransport(ctx context.Context, transport client.IPsecTransportConfig) error {
	panic("not implemented")
}

func (m *MockClient) UpdateIPsecTransport(ctx context.Context, transport client.IPsecTransportConfig) error {
	panic("not implemented")
}

func (m *MockClient) DeleteIPsecTransport(ctx context.Context, transportID int) error {
	panic("not implemented")
}

func (m *MockClient) ListIPsecTransports(ctx context.Context) ([]client.IPsecTransportConfig, error) {
	panic("not implemented")
}

// L2TP methods
func (m *MockClient) GetL2TP(ctx context.Context, tunnelID int) (*client.L2TPConfig, error) {
	panic("not implemented")
}

func (m *MockClient) CreateL2TP(ctx context.Context, config client.L2TPConfig) error {
	panic("not implemented")
}

func (m *MockClient) UpdateL2TP(ctx context.Context, config client.L2TPConfig) error {
	panic("not implemented")
}

func (m *MockClient) DeleteL2TP(ctx context.Context, tunnelID int) error {
	panic("not implemented")
}

func (m *MockClient) ListL2TPs(ctx context.Context) ([]client.L2TPConfig, error) {
	panic("not implemented")
}

func (m *MockClient) GetL2TPServiceState(ctx context.Context) (*client.L2TPServiceState, error) {
	panic("not implemented")
}

func (m *MockClient) SetL2TPServiceState(ctx context.Context, enabled bool, protocols []string) error {
	panic("not implemented")
}

// PPTP methods
func (m *MockClient) GetPPTP(ctx context.Context) (*client.PPTPConfig, error) {
	panic("not implemented")
}

func (m *MockClient) CreatePPTP(ctx context.Context, config client.PPTPConfig) error {
	panic("not implemented")
}

func (m *MockClient) UpdatePPTP(ctx context.Context, config client.PPTPConfig) error {
	panic("not implemented")
}

func (m *MockClient) DeletePPTP(ctx context.Context) error {
	panic("not implemented")
}

// Syslog methods
func (m *MockClient) GetSyslogConfig(ctx context.Context) (*client.SyslogConfig, error) {
	panic("not implemented")
}

func (m *MockClient) ConfigureSyslog(ctx context.Context, config client.SyslogConfig) error {
	panic("not implemented")
}

func (m *MockClient) UpdateSyslogConfig(ctx context.Context, config client.SyslogConfig) error {
	panic("not implemented")
}

func (m *MockClient) ResetSyslog(ctx context.Context) error {
	panic("not implemented")
}

// QoS Class Map methods
func (m *MockClient) GetClassMap(ctx context.Context, name string) (*client.ClassMap, error) {
	panic("not implemented")
}

func (m *MockClient) CreateClassMap(ctx context.Context, cm client.ClassMap) error {
	panic("not implemented")
}

func (m *MockClient) UpdateClassMap(ctx context.Context, cm client.ClassMap) error {
	panic("not implemented")
}

func (m *MockClient) DeleteClassMap(ctx context.Context, name string) error {
	panic("not implemented")
}

func (m *MockClient) ListClassMaps(ctx context.Context) ([]client.ClassMap, error) {
	panic("not implemented")
}

// QoS Policy Map methods
func (m *MockClient) GetPolicyMap(ctx context.Context, name string) (*client.PolicyMap, error) {
	panic("not implemented")
}

func (m *MockClient) CreatePolicyMap(ctx context.Context, pm client.PolicyMap) error {
	panic("not implemented")
}

func (m *MockClient) UpdatePolicyMap(ctx context.Context, pm client.PolicyMap) error {
	panic("not implemented")
}

func (m *MockClient) DeletePolicyMap(ctx context.Context, name string) error {
	panic("not implemented")
}

func (m *MockClient) ListPolicyMaps(ctx context.Context) ([]client.PolicyMap, error) {
	panic("not implemented")
}

// QoS Service Policy methods
func (m *MockClient) GetServicePolicy(ctx context.Context, iface string, direction string) (*client.ServicePolicy, error) {
	panic("not implemented")
}

func (m *MockClient) CreateServicePolicy(ctx context.Context, sp client.ServicePolicy) error {
	panic("not implemented")
}

func (m *MockClient) UpdateServicePolicy(ctx context.Context, sp client.ServicePolicy) error {
	panic("not implemented")
}

func (m *MockClient) DeleteServicePolicy(ctx context.Context, iface string, direction string) error {
	panic("not implemented")
}

func (m *MockClient) ListServicePolicies(ctx context.Context) ([]client.ServicePolicy, error) {
	panic("not implemented")
}

// QoS Shape methods
func (m *MockClient) GetShape(ctx context.Context, iface string, direction string) (*client.ShapeConfig, error) {
	panic("not implemented")
}

func (m *MockClient) CreateShape(ctx context.Context, sc client.ShapeConfig) error {
	panic("not implemented")
}

func (m *MockClient) UpdateShape(ctx context.Context, sc client.ShapeConfig) error {
	panic("not implemented")
}

func (m *MockClient) DeleteShape(ctx context.Context, iface string, direction string) error {
	panic("not implemented")
}

func (m *MockClient) ListShapes(ctx context.Context) ([]client.ShapeConfig, error) {
	panic("not implemented")
}

// SNMP methods
func (m *MockClient) GetSNMP(ctx context.Context) (*client.SNMPConfig, error) {
	panic("not implemented")
}

func (m *MockClient) CreateSNMP(ctx context.Context, config client.SNMPConfig) error {
	panic("not implemented")
}

func (m *MockClient) UpdateSNMP(ctx context.Context, config client.SNMPConfig) error {
	panic("not implemented")
}

func (m *MockClient) DeleteSNMP(ctx context.Context) error {
	panic("not implemented")
}

// Schedule methods
func (m *MockClient) GetSchedule(ctx context.Context, id int) (*client.Schedule, error) {
	panic("not implemented")
}

func (m *MockClient) CreateSchedule(ctx context.Context, schedule client.Schedule) error {
	panic("not implemented")
}

func (m *MockClient) UpdateSchedule(ctx context.Context, schedule client.Schedule) error {
	panic("not implemented")
}

func (m *MockClient) DeleteSchedule(ctx context.Context, id int) error {
	panic("not implemented")
}

func (m *MockClient) ListSchedules(ctx context.Context) ([]client.Schedule, error) {
	panic("not implemented")
}

// Kron Policy methods
func (m *MockClient) GetKronPolicy(ctx context.Context, name string) (*client.KronPolicy, error) {
	panic("not implemented")
}

func (m *MockClient) CreateKronPolicy(ctx context.Context, policy client.KronPolicy) error {
	panic("not implemented")
}

func (m *MockClient) UpdateKronPolicy(ctx context.Context, policy client.KronPolicy) error {
	panic("not implemented")
}

func (m *MockClient) DeleteKronPolicy(ctx context.Context, name string) error {
	panic("not implemented")
}

func (m *MockClient) ListKronPolicies(ctx context.Context) ([]client.KronPolicy, error) {
	panic("not implemented")
}

// DNS methods
func (m *MockClient) GetDNS(ctx context.Context) (*client.DNSConfig, error) {
	panic("not implemented")
}

func (m *MockClient) ConfigureDNS(ctx context.Context, config client.DNSConfig) error {
	panic("not implemented")
}

func (m *MockClient) UpdateDNS(ctx context.Context, config client.DNSConfig) error {
	panic("not implemented")
}

func (m *MockClient) ResetDNS(ctx context.Context) error {
	panic("not implemented")
}

func (m *MockClient) GetAdminConfig(ctx context.Context) (*client.AdminConfig, error) {
	panic("not implemented")
}

func (m *MockClient) ConfigureAdmin(ctx context.Context, config client.AdminConfig) error {
	panic("not implemented")
}

func (m *MockClient) UpdateAdminConfig(ctx context.Context, config client.AdminConfig) error {
	panic("not implemented")
}

func (m *MockClient) ResetAdmin(ctx context.Context) error {
	panic("not implemented")
}

func (m *MockClient) GetAdminUser(ctx context.Context, username string) (*client.AdminUser, error) {
	panic("not implemented")
}

func (m *MockClient) CreateAdminUser(ctx context.Context, user client.AdminUser) error {
	panic("not implemented")
}

func (m *MockClient) UpdateAdminUser(ctx context.Context, user client.AdminUser) error {
	panic("not implemented")
}

func (m *MockClient) DeleteAdminUser(ctx context.Context, username string) error {
	panic("not implemented")
}

func (m *MockClient) ListAdminUsers(ctx context.Context) ([]client.AdminUser, error) {
	panic("not implemented")
}

func (m *MockClient) GetHTTPD(ctx context.Context) (*client.HTTPDConfig, error) {
	panic("not implemented")
}

func (m *MockClient) ConfigureHTTPD(ctx context.Context, config client.HTTPDConfig) error {
	panic("not implemented")
}

func (m *MockClient) UpdateHTTPD(ctx context.Context, config client.HTTPDConfig) error {
	panic("not implemented")
}

func (m *MockClient) ResetHTTPD(ctx context.Context) error {
	panic("not implemented")
}

func (m *MockClient) GetSSHD(ctx context.Context) (*client.SSHDConfig, error) {
	panic("not implemented")
}

func (m *MockClient) ConfigureSSHD(ctx context.Context, config client.SSHDConfig) error {
	panic("not implemented")
}

func (m *MockClient) UpdateSSHD(ctx context.Context, config client.SSHDConfig) error {
	panic("not implemented")
}

func (m *MockClient) ResetSSHD(ctx context.Context) error {
	panic("not implemented")
}

func (m *MockClient) GetSFTPD(ctx context.Context) (*client.SFTPDConfig, error) {
	panic("not implemented")
}

func (m *MockClient) ConfigureSFTPD(ctx context.Context, config client.SFTPDConfig) error {
	panic("not implemented")
}

func (m *MockClient) UpdateSFTPD(ctx context.Context, config client.SFTPDConfig) error {
	panic("not implemented")
}

func (m *MockClient) ResetSFTPD(ctx context.Context) error {
	panic("not implemented")
}

// Bridge methods
func (m *MockClient) GetBridge(ctx context.Context, name string) (*client.BridgeConfig, error) {
	panic("not implemented")
}

func (m *MockClient) CreateBridge(ctx context.Context, bridge client.BridgeConfig) error {
	panic("not implemented")
}

func (m *MockClient) UpdateBridge(ctx context.Context, bridge client.BridgeConfig) error {
	panic("not implemented")
}

func (m *MockClient) DeleteBridge(ctx context.Context, name string) error {
	panic("not implemented")
}

func (m *MockClient) ListBridges(ctx context.Context) ([]client.BridgeConfig, error) {
	panic("not implemented")
}

// IPv6 Interface methods
func (m *MockClient) GetIPv6InterfaceConfig(ctx context.Context, interfaceName string) (*client.IPv6InterfaceConfig, error) {
	panic("not implemented")
}

func (m *MockClient) ConfigureIPv6Interface(ctx context.Context, config client.IPv6InterfaceConfig) error {
	panic("not implemented")
}

func (m *MockClient) UpdateIPv6InterfaceConfig(ctx context.Context, config client.IPv6InterfaceConfig) error {
	panic("not implemented")
}

func (m *MockClient) ResetIPv6Interface(ctx context.Context, interfaceName string) error {
	panic("not implemented")
}

func (m *MockClient) ListIPv6InterfaceConfigs(ctx context.Context) ([]client.IPv6InterfaceConfig, error) {
	panic("not implemented")
}

// Access List Extended (IPv4)
func (m *MockClient) GetAccessListExtended(ctx context.Context, name string) (*client.AccessListExtended, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClient) CreateAccessListExtended(ctx context.Context, acl client.AccessListExtended) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) UpdateAccessListExtended(ctx context.Context, acl client.AccessListExtended) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) DeleteAccessListExtended(ctx context.Context, name string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) ListAccessListsExtended(ctx context.Context) ([]client.AccessListExtended, error) {
	return nil, fmt.Errorf("not implemented")
}

// Access List Extended (IPv6)
func (m *MockClient) GetAccessListExtendedIPv6(ctx context.Context, name string) (*client.AccessListExtendedIPv6, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClient) CreateAccessListExtendedIPv6(ctx context.Context, acl client.AccessListExtendedIPv6) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) UpdateAccessListExtendedIPv6(ctx context.Context, acl client.AccessListExtendedIPv6) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) DeleteAccessListExtendedIPv6(ctx context.Context, name string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) ListAccessListsExtendedIPv6(ctx context.Context) ([]client.AccessListExtendedIPv6, error) {
	return nil, fmt.Errorf("not implemented")
}

// IP Filter Dynamic Config
func (m *MockClient) GetIPFilterDynamicConfig(ctx context.Context) (*client.IPFilterDynamicConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClient) CreateIPFilterDynamicConfig(ctx context.Context, config client.IPFilterDynamicConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) UpdateIPFilterDynamicConfig(ctx context.Context, config client.IPFilterDynamicConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) DeleteIPFilterDynamicConfig(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

// IPv6 Filter Dynamic Config
func (m *MockClient) GetIPv6FilterDynamicConfig(ctx context.Context) (*client.IPv6FilterDynamicConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClient) CreateIPv6FilterDynamicConfig(ctx context.Context, config client.IPv6FilterDynamicConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) UpdateIPv6FilterDynamicConfig(ctx context.Context, config client.IPv6FilterDynamicConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) DeleteIPv6FilterDynamicConfig(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

// Interface ACL
func (m *MockClient) GetInterfaceACL(ctx context.Context, iface string) (*client.InterfaceACL, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClient) CreateInterfaceACL(ctx context.Context, acl client.InterfaceACL) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) UpdateInterfaceACL(ctx context.Context, acl client.InterfaceACL) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) DeleteInterfaceACL(ctx context.Context, iface string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) ListInterfaceACLs(ctx context.Context) ([]client.InterfaceACL, error) {
	return nil, fmt.Errorf("not implemented")
}

// Access List MAC
func (m *MockClient) GetAccessListMAC(ctx context.Context, name string) (*client.AccessListMAC, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClient) CreateAccessListMAC(ctx context.Context, acl client.AccessListMAC) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) UpdateAccessListMAC(ctx context.Context, acl client.AccessListMAC) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) DeleteAccessListMAC(ctx context.Context, name string, filterNums []int) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) ListAccessListsMAC(ctx context.Context) ([]client.AccessListMAC, error) {
	return nil, fmt.Errorf("not implemented")
}

// Interface MAC ACL
func (m *MockClient) GetInterfaceMACACL(ctx context.Context, iface string) (*client.InterfaceMACACL, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClient) CreateInterfaceMACACL(ctx context.Context, acl client.InterfaceMACACL) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) UpdateInterfaceMACACL(ctx context.Context, acl client.InterfaceMACACL) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) DeleteInterfaceMACACL(ctx context.Context, iface string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) ListInterfaceMACACLs(ctx context.Context) ([]client.InterfaceMACACL, error) {
	return nil, fmt.Errorf("not implemented")
}

// DDNS - NetVolante DNS methods
func (m *MockClient) GetNetVolanteDNS(ctx context.Context) ([]client.NetVolanteConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClient) GetNetVolanteDNSByInterface(ctx context.Context, iface string) (*client.NetVolanteConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClient) ConfigureNetVolanteDNS(ctx context.Context, config client.NetVolanteConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) UpdateNetVolanteDNS(ctx context.Context, config client.NetVolanteConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) DeleteNetVolanteDNS(ctx context.Context, iface string) error {
	return fmt.Errorf("not implemented")
}

// DDNS - Custom DDNS methods
func (m *MockClient) GetDDNS(ctx context.Context) ([]client.DDNSServerConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClient) GetDDNSByID(ctx context.Context, id int) (*client.DDNSServerConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClient) ConfigureDDNS(ctx context.Context, config client.DDNSServerConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) UpdateDDNS(ctx context.Context, config client.DDNSServerConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) DeleteDDNS(ctx context.Context, id int) error {
	return fmt.Errorf("not implemented")
}

// DDNS - Status methods
func (m *MockClient) GetNetVolanteDNSStatus(ctx context.Context) ([]client.DDNSStatus, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClient) GetDDNSStatus(ctx context.Context) ([]client.DDNSStatus, error) {
	return nil, fmt.Errorf("not implemented")
}

// PP Interface methods
func (m *MockClient) GetPPInterfaceConfig(ctx context.Context, ppNum int) (*client.PPIPConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClient) ConfigurePPInterface(ctx context.Context, ppNum int, config client.PPIPConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) UpdatePPInterfaceConfig(ctx context.Context, ppNum int, config client.PPIPConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) ResetPPInterfaceConfig(ctx context.Context, ppNum int) error {
	return fmt.Errorf("not implemented")
}

// PPPoE methods
func (m *MockClient) ListPPPoE(ctx context.Context) ([]client.PPPoEConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClient) GetPPPoE(ctx context.Context, ppNum int) (*client.PPPoEConfig, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockClient) CreatePPPoE(ctx context.Context, config client.PPPoEConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) UpdatePPPoE(ctx context.Context, config client.PPPoEConfig) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) DeletePPPoE(ctx context.Context, ppNum int) error {
	return fmt.Errorf("not implemented")
}

func (m *MockClient) GetPPConnectionStatus(ctx context.Context, ppNum int) (*client.PPConnectionStatus, error) {
	return nil, fmt.Errorf("not implemented")
}

func TestRTXSystemInfoDataSourceSchema(t *testing.T) {
	dataSource := dataSourceRTXSystemInfo()

	// Test that the data source is properly configured
	assert.NotNil(t, dataSource)
	assert.NotNil(t, dataSource.Schema)
	assert.NotNil(t, dataSource.ReadContext)

	// Test schema structure
	schemaMap := dataSource.Schema

	// Check required fields
	requiredFields := []string{"model", "firmware_version", "serial_number", "mac_address", "uptime"}
	for _, field := range requiredFields {
		assert.Contains(t, schemaMap, field, "Schema should contain %s field", field)
		assert.Equal(t, schema.TypeString, schemaMap[field].Type, "%s should be of type string", field)
		assert.True(t, schemaMap[field].Computed, "%s should be computed", field)
	}

	// Check that id field exists and is computed
	assert.Contains(t, schemaMap, "id")
	assert.Equal(t, schema.TypeString, schemaMap["id"].Type)
	assert.True(t, schemaMap["id"].Computed)
}

func TestRTXSystemInfoDataSourceRead_Success(t *testing.T) {
	mockClient := &MockClient{}

	// Mock successful system info retrieval
	expectedSystemInfo := &client.SystemInfo{
		Model:           "RTX1210",
		FirmwareVersion: "Rev.14.01.27",
		SerialNumber:    "ABC123456789",
		MACAddress:      "00:a0:de:12:34:56",
		Uptime:          "15 days, 10:30:25",
	}

	mockClient.On("GetSystemInfo", mock.Anything).Return(expectedSystemInfo, nil)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXSystemInfo().Schema, map[string]interface{}{})

	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXSystemInfoRead(ctx, d, apiClient)

	// Assert no errors
	assert.Empty(t, diags)

	// Assert data was set correctly
	assert.Equal(t, "RTX1210", d.Get("model").(string))
	assert.Equal(t, "Rev.14.01.27", d.Get("firmware_version").(string))
	assert.Equal(t, "ABC123456789", d.Get("serial_number").(string))
	assert.Equal(t, "00:a0:de:12:34:56", d.Get("mac_address").(string))
	assert.Equal(t, "15 days, 10:30:25", d.Get("uptime").(string))
	assert.NotEmpty(t, d.Id())

	mockClient.AssertExpectations(t)
}

func TestRTXSystemInfoDataSourceRead_ClientError(t *testing.T) {
	mockClient := &MockClient{}

	// Mock client error
	expectedError := errors.New("SSH connection failed")
	mockClient.On("GetSystemInfo", mock.Anything).Return((*client.SystemInfo)(nil), expectedError)

	// Create a resource data mock
	d := schema.TestResourceDataRaw(t, dataSourceRTXSystemInfo().Schema, map[string]interface{}{})

	// Create mock API client
	apiClient := &apiClient{client: mockClient}

	// Call the read function
	ctx := context.Background()
	diags := dataSourceRTXSystemInfoRead(ctx, d, apiClient)

	// Assert error occurred
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, "Failed to retrieve system information")
	assert.Contains(t, diags[0].Summary, "SSH connection failed")

	mockClient.AssertExpectations(t)
}

// Note: Parse error test removed as GetSystemInfo now handles parsing internally
// and parseSystemInfo doesn't return errors - it returns a SystemInfo struct with empty fields
// if parsing fails, which is still valid for the data source.

// TestRTXSystemInfoDataSourceRead_CommandError removed since client.Result doesn't have Error field

// Test parsing of various RTX output formats
func TestParseSystemInfo(t *testing.T) {
	tests := []struct {
		name           string
		output         string
		expectedModel  string
		expectedFW     string
		expectedSerial string
		expectedMAC    string
		expectedUptime string
		shouldError    bool
	}{
		{
			name: "Standard RTX1210 format",
			output: `Model: RTX1210
Firmware Version: Rev.14.01.27
Serial Number: ABC123456789
MAC Address: 00:a0:de:12:34:56
Uptime: 15 days, 10:30:25`,
			expectedModel:  "RTX1210",
			expectedFW:     "Rev.14.01.27",
			expectedSerial: "ABC123456789",
			expectedMAC:    "00:a0:de:12:34:56",
			expectedUptime: "15 days, 10:30:25",
			shouldError:    false,
		},
		{
			name: "RTX830 format with different spacing",
			output: `Model:    RTX830
Firmware Version:  Rev.15.02.10
Serial Number:   XYZ987654321
MAC Address:    aa:bb:cc:dd:ee:ff
Uptime:   3 days, 5:15:42`,
			expectedModel:  "RTX830",
			expectedFW:     "Rev.15.02.10",
			expectedSerial: "XYZ987654321",
			expectedMAC:    "aa:bb:cc:dd:ee:ff",
			expectedUptime: "3 days, 5:15:42",
			shouldError:    false,
		},
		{
			name: "Short uptime format",
			output: `Model: RTX1200
Firmware Version: Rev.10.01.80
Serial Number: DEF456789123
MAC Address: 12:34:56:78:9a:bc
Uptime: 2:15:30`,
			expectedModel:  "RTX1200",
			expectedFW:     "Rev.10.01.80",
			expectedSerial: "DEF456789123",
			expectedMAC:    "12:34:56:78:9a:bc",
			expectedUptime: "2:15:30",
			shouldError:    false,
		},
		{
			name:        "Missing model field",
			output:      `Firmware Version: Rev.14.01.27`,
			shouldError: true,
		},
		{
			name:        "Empty output",
			output:      "",
			shouldError: true,
		},
		{
			name:        "Invalid format",
			output:      "This is not a valid RTX output",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSystemInfo(tt.output)

			if tt.shouldError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedModel, result.Model)
			assert.Equal(t, tt.expectedFW, result.FirmwareVersion)
			assert.Equal(t, tt.expectedSerial, result.SerialNumber)
			assert.Equal(t, tt.expectedMAC, result.MACAddress)
			assert.Equal(t, tt.expectedUptime, result.Uptime)
		})
	}
}

// Acceptance tests using Docker test environment
func TestAccRTXSystemInfoDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXSystemInfoDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRTXSystemInfoDataSourceExists("data.rtx_system_info.test"),
					resource.TestCheckResourceAttrSet("data.rtx_system_info.test", "id"),
					resource.TestCheckResourceAttrSet("data.rtx_system_info.test", "model"),
					resource.TestCheckResourceAttrSet("data.rtx_system_info.test", "firmware_version"),
					resource.TestCheckResourceAttrSet("data.rtx_system_info.test", "serial_number"),
					resource.TestCheckResourceAttrSet("data.rtx_system_info.test", "mac_address"),
					resource.TestCheckResourceAttrSet("data.rtx_system_info.test", "uptime"),
					// Test that model matches RTX pattern
					resource.TestMatchResourceAttr("data.rtx_system_info.test", "model", regexp.MustCompile(`^RTX\d+$`)),
					// Test that firmware version has correct format
					resource.TestMatchResourceAttr("data.rtx_system_info.test", "firmware_version", regexp.MustCompile(`^Rev\.\d+\.\d+\.\d+$`)),
					// Test that MAC address has correct format
					resource.TestMatchResourceAttr("data.rtx_system_info.test", "mac_address", regexp.MustCompile(`^([0-9a-fA-F]{2}:){5}[0-9a-fA-F]{2}$`)),
				),
			},
		},
	})
}

func TestAccRTXSystemInfoDataSource_attributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRTXSystemInfoDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check that all required attributes are present and non-empty
					resource.TestCheckResourceAttrWith("data.rtx_system_info.test", "model", func(value string) error {
						if strings.TrimSpace(value) == "" {
							return fmt.Errorf("model should not be empty")
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("data.rtx_system_info.test", "firmware_version", func(value string) error {
						if strings.TrimSpace(value) == "" {
							return fmt.Errorf("firmware_version should not be empty")
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("data.rtx_system_info.test", "serial_number", func(value string) error {
						if strings.TrimSpace(value) == "" {
							return fmt.Errorf("serial_number should not be empty")
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("data.rtx_system_info.test", "mac_address", func(value string) error {
						if strings.TrimSpace(value) == "" {
							return fmt.Errorf("mac_address should not be empty")
						}
						return nil
					}),
					resource.TestCheckResourceAttrWith("data.rtx_system_info.test", "uptime", func(value string) error {
						if strings.TrimSpace(value) == "" {
							return fmt.Errorf("uptime should not be empty")
						}
						return nil
					}),
				),
			},
		},
	})
}

func testAccRTXSystemInfoDataSourceConfig() string {
	return `
data "rtx_system_info" "test" {}
`
}

func testAccCheckRTXSystemInfoDataSourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("data source not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("data source ID not set")
		}

		return nil
	}
}

// Helper functions for tests
func testAccPreCheck(t *testing.T) {
	// Skip if not running acceptance tests
	if !isAccTest() {
		t.Skip("Skipping acceptance test")
	}

	// Check required environment variables for Docker test environment
	requiredEnvVars := []string{
		"RTX_HOST",
		"RTX_USERNAME",
		"RTX_PASSWORD",
	}

	for _, envVar := range requiredEnvVars {
		if value := getEnvVar(envVar); value == "" {
			t.Fatalf("Environment variable %s must be set for acceptance tests", envVar)
		}
	}
}

// isAccTest checks if acceptance tests are enabled
func isAccTest() bool {
	return os.Getenv("TF_ACC") != ""
}

// getEnvVar gets environment variable value
func getEnvVar(name string) string {
	return os.Getenv(name)
}

// testAccProviderFactories provides test provider factories for acceptance tests
var testAccProviderFactories = map[string]func() (*schema.Provider, error){
	"rtx": func() (*schema.Provider, error) {
		return New("test"), nil
	},
}

// parseSystemInfo parses RTX system information from command output
func parseSystemInfo(output string) (*client.SystemInfo, error) {
	if strings.TrimSpace(output) == "" {
		return nil, errors.New("empty output")
	}

	info := &client.SystemInfo{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "Model":
			info.Model = value
		case "Firmware Version":
			info.FirmwareVersion = value
		case "Serial Number":
			info.SerialNumber = value
		case "MAC Address":
			info.MACAddress = value
		case "Uptime":
			info.Uptime = value
		}
	}

	// Validate that we got all required fields
	if info.Model == "" {
		return nil, errors.New("model not found in output")
	}
	if info.FirmwareVersion == "" {
		return nil, errors.New("firmware version not found in output")
	}
	if info.SerialNumber == "" {
		return nil, errors.New("serial number not found in output")
	}
	if info.MACAddress == "" {
		return nil, errors.New("MAC address not found in output")
	}
	if info.Uptime == "" {
		return nil, errors.New("uptime not found in output")
	}

	return info, nil
}
