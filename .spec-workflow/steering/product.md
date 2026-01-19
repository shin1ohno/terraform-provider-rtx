# Product Overview

## Product Purpose

Terraform Provider for Yamaha RTX series routers enables Infrastructure as Code (IaC) management of enterprise-grade network infrastructure. It provides comprehensive coverage of all RTX router features—from basic networking and routing to advanced VPN, security, and QoS configurations. The provider solves the challenge of manually configuring and maintaining RTX router settings by providing declarative, version-controlled, and reproducible network configurations.

## Target Users

- **Network Administrators**: Managing enterprise networks with Yamaha RTX routers
- **DevOps Engineers**: Integrating network infrastructure into CI/CD pipelines
- **System Integrators**: Deploying consistent network configurations across multiple sites
- **IT Teams in Japan**: Primary market where RTX routers are widely deployed

### User Pain Points
- Manual CLI configuration is error-prone and time-consuming
- No version control for network configurations
- Difficult to replicate configurations across environments
- Lack of automation for network infrastructure changes

## Key Features

This provider implements comprehensive support for all Yamaha RTX router features:

### Network & Routing
- **IP Configuration**: IP addresses, routing tables, static routes, policy routing
- **OSPF/OSPFv3**: Dynamic routing with OSPF for IPv4 and IPv6
- **BGP**: Border Gateway Protocol configuration and peer management
- **IPv6**: Full IPv6 support including addressing and routing

### Network Services
- **DHCP**: Scopes, static bindings, relay, and advanced options
- **DNS**: DNS server, resolver, and forwarding configuration
- **NAT/NAPT**: Network address translation rules and mappings
- **NAT46/DNS46**: IPv4-IPv6 translation mechanisms

### VPN & Tunneling
- **IPsec**: IKE/IPsec VPN tunnels with full parameter control
- **L2TP**: Layer 2 Tunneling Protocol configuration
- **PPTP**: Point-to-Point Tunneling Protocol
- **IPIP Tunneling**: IP-in-IP encapsulation
- **Cloud VPN**: Integration with cloud service VPN connections

### Security & Filtering
- **Ethernet Filters**: Layer 2 packet filtering
- **IP Filters**: Layer 3/4 packet filtering and firewall rules
- **URL Filters**: Web content filtering

### Quality of Service
- **QoS/Traffic Shaping**: Priority control and bandwidth management
- **DPI**: Deep Packet Inspection configuration

### Interfaces & Connectivity
- **VLAN**: Virtual LAN configuration
- **Bridge Interface**: Layer 2 bridging
- **Flexible LAN/WAN Ports**: Port role assignment
- **PPP/PPPoE**: Point-to-Point Protocol configuration
- **Mobile Internet**: Mobile broadband connectivity
- **ISDN/PRI**: Legacy connectivity options

### Monitoring & Management
- **SNMP**: SNMP agent and trap configuration
- **RADIUS**: Authentication server integration
- **SIP**: VoIP/SIP functionality
- **Logging**: Syslog and logging configuration
- **Statistics**: Traffic and performance statistics
- **Diagnostics**: Network diagnostic tools

### Automation & Integration
- **Schedules**: Time-based automation rules
- **Email Notifications**: Trigger-based alerts
- **HTTP Server**: Built-in web server configuration
- **Lua Scripts**: Custom scripting support
- **UPnP**: Universal Plug and Play settings
- **L2MS/YNO**: Yamaha Network Organizer integration

### System Management
- **Configuration Persistence**: Automatic saving to persistent memory
- **Import Support**: Import existing configurations into Terraform state
- **USB/External Memory**: External storage configuration
- **NetVolante DNS**: Dynamic DNS service integration

## Business Objectives

- Enable complete IaC management of Yamaha RTX router configurations
- Reduce configuration errors through declarative definitions and validation
- Support enterprise deployment patterns with proper state management
- Provide seamless integration with existing Terraform workflows

## Success Metrics

- **Test Coverage**: Comprehensive unit and acceptance tests for all resources
- **Documentation**: Complete schema documentation and usage examples
- **Reliability**: Stable SSH connections and proper error handling
- **Adoption**: Usage in production environments for RTX router management

## Product Principles

1. **Cisco-Compatible Syntax**: Align resource and attribute naming with Cisco IOS XE Terraform provider (`CiscoDevNet/iosxe`) conventions wherever possible. This enables network administrators to manage both Cisco and RTX devices with minimal learning curve and consistent Terraform configurations.
2. **Follow RTX CLI Semantics**: Where Cisco conventions don't apply, mirror RTX router terminology for familiarity
3. **Fail Safely**: Validate configurations before applying; provide clear error messages
4. **Leverage Existing Patterns**: Reuse established Terraform provider conventions and Go SSH infrastructure
5. **Minimal Scope**: Only implement features that provide clear value for IaC management
6. **State Clarity**: Persist only configuration in Terraform state; do not store operational/runtime status to avoid perpetual diffs

## API Design Philosophy: Cisco Compatibility

To maximize interoperability and reduce cognitive load for network engineers, this provider follows Cisco IOS XE Terraform provider naming conventions:

### Resource Naming
| RTX Resource | Cisco Equivalent | Pattern |
|--------------|------------------|---------|
| `rtx_static_route` | `iosxe_static_route` | `<provider>_static_route` |
| `rtx_ospf` | `iosxe_ospf` | `<provider>_ospf` |
| `rtx_bgp` | `iosxe_bgp` | `<provider>_bgp` |
| `rtx_vlan` | `iosxe_vlan` | `<provider>_vlan` |
| `rtx_nat` | `iosxe_nat` | `<provider>_nat` |
| `rtx_access_list` | `iosxe_access_list_extended` | `<provider>_access_list*` |
| `rtx_interface` | `iosxe_interface_ethernet` | `<provider>_interface*` |

### Attribute Naming Conventions
| Attribute | Example | Notes |
|-----------|---------|-------|
| `process_id` | OSPF process ID | Not just `id` |
| `router_id` | OSPF/BGP router ID | Dotted decimal format |
| `asn` | BGP AS number | String for 4-byte ASN support |
| `vlan_id` | VLAN identifier | Integer 1-4094 |
| `prefix` / `mask` | Route destination | Separate prefix and mask |
| `next_hops` | Route next hops | List of objects |
| `networks` | OSPF/BGP networks | List of network objects |
| `neighbors` | Routing neighbors | List of neighbor objects |
| `entries` | ACL/filter rules | List of rule entries |
| `shutdown` | Admin state | Boolean, true = disabled |

### Nested Block Patterns
```hcl
# Cisco-style nested blocks for RTX
resource "rtx_ospf" "backbone" {
  process_id = 1
  router_id  = "1.1.1.1"

  networks = [
    {
      ip       = "10.0.0.0"
      wildcard = "0.0.0.255"
      area     = "0"
    }
  ]

  neighbors = [
    {
      ip       = "2.2.2.2"
      priority = 10
      cost     = 100
    }
  ]
}

resource "rtx_static_route" "default" {
  prefix = "0.0.0.0"
  mask   = "0.0.0.0"

  next_hops = [
    {
      next_hop = "192.168.1.1"
      distance = 10
      name     = "primary_gateway"
    }
  ]
}
```

### Benefits of Cisco Compatibility
- **Unified Workflow**: Manage Cisco and RTX devices with similar Terraform configurations
- **Transferable Knowledge**: Skills learned on one provider apply to the other
- **Copy-Paste Friendly**: Easier migration of configurations between vendors
- **Reduced Errors**: Consistent naming reduces configuration mistakes

## Monitoring & Visibility

- **CLI-based**: Standard Terraform CLI output and logging
- **Debug Logging**: TF_LOG environment variable for detailed operation tracing
- **State File**: Terraform state tracks all managed resources
- **Plan Output**: Clear diff visualization before applying changes

## Future Vision

### Potential Enhancements
- **Multi-Router Support**: Manage multiple RTX routers from a single configuration
- **Configuration Backup**: Export and restore complete router configurations
- **Configuration Diff**: Compare running and saved configurations
- **Cluster Support**: HA/failover cluster management
- **Terraform Cloud Integration**: Remote state and collaboration features
- **Provider Versioning**: Compatibility across different RTX firmware versions
