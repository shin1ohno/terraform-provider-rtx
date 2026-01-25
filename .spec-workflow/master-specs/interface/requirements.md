# Master Requirements: Interface Resources

## Overview

This specification covers the network interface resources for the Terraform RTX provider. These resources manage layer 2 and layer 3 interface configurations on Yamaha RTX routers, including physical interfaces (LAN), virtual interfaces (VLAN, Bridge), point-to-point interfaces (PP), and both IPv4 and IPv6 addressing.

## Alignment with Product Vision

These interface resources are foundational to the Terraform RTX provider, enabling Infrastructure-as-Code management of Yamaha RTX router network configurations. They support:

- Declarative network interface management
- Consistent configuration across environments
- Version-controlled network infrastructure
- Import of existing configurations

## Resources Covered

| Resource | Terraform Name | Type | Import Support | Description |
|----------|---------------|------|----------------|-------------|
| Interface | `rtx_interface` | per-interface | Yes | IPv4 interface configuration (LAN, Bridge, PP, Tunnel) |
| IPv6 Interface | `rtx_ipv6_interface` | per-interface | Yes | IPv6 interface configuration with RTADV/DHCPv6 |
| PP Interface | `rtx_pp_interface` | per-pp-number | Yes | Point-to-Point interface IP configuration |
| VLAN | `rtx_vlan` | per-vlan-id | Yes | 802.1Q VLAN interface on LAN ports |
| Bridge | `rtx_bridge` | per-bridge | Yes | Layer 2 bridge combining multiple interfaces |

---

## Resource 1: rtx_interface

### Overview

Manages IPv4 network interface configuration on RTX routers including IP addresses, security filters, NAT binding, and Ethernet filters.

### Functional Requirements

#### Core Operations

##### Create
- Configure IP address (static CIDR or DHCP)
- Set interface description
- Apply inbound/outbound security filters
- Apply dynamic filters for stateful inspection
- Bind NAT descriptor
- Enable/disable ProxyARP
- Set MTU value
- Apply Ethernet (L2) filters
- Save configuration to router flash

##### Read
- Retrieve all configured attributes from router
- Parse `show config` output for the interface
- Return "not found" if interface has no configuration

##### Update
- Detect changed attributes by comparing with current state
- Remove old configuration before applying new values
- Support partial updates (only changed attributes)
- Save configuration after updates

##### Delete
- Reset interface to factory defaults
- Remove all configured attributes
- Save configuration after reset

### Schema Definition

| Attribute | Type | Required | ForceNew | Computed | Validation | Description |
|-----------|------|----------|----------|----------|------------|-------------|
| `name` | string | Yes | Yes | No | `^(lan\|bridge\|pp\|tunnel)\d+$` | Interface name |
| `description` | string | No | No | No | - | Interface description |
| `ip_address` | block | No | No | No | MaxItems: 1 | IP address configuration |
| `ip_address.address` | string | No | No | No | CIDR notation | Static IP (e.g., "192.168.1.1/24") |
| `ip_address.dhcp` | bool | No | No | Yes | - | Use DHCP for IP assignment |
| `secure_filter_in` | list(int) | No | No | No | Each >= 1 | Inbound security filter numbers |
| `secure_filter_out` | list(int) | No | No | No | Each >= 1 | Outbound security filter numbers |
| `dynamic_filter_out` | list(int) | No | No | No | Each >= 1 | Dynamic filter numbers for stateful inspection |
| `nat_descriptor` | int | No | No | Yes | >= 0 | NAT descriptor ID |
| `proxyarp` | bool | No | No | Yes | - | Enable ProxyARP |
| `mtu` | int | No | No | Yes | 0-65535 | MTU size (0 = default) |
| `ethernet_filter_in` | list(int) | No | No | No | Each 1-512 | Inbound Ethernet filter numbers |
| `ethernet_filter_out` | list(int) | No | No | No | Each 1-512 | Outbound Ethernet filter numbers |
| `interface_name` | string | No | No | Yes | - | Computed interface name (same as `name`) |

### Import Specification

- **Import ID Format**: Interface name (e.g., `lan1`, `bridge1`, `pp1`, `tunnel1`)
- **Import Command**: `terraform import rtx_interface.example lan1`
- **Post-Import**: State is populated from router configuration

---

## Resource 2: rtx_ipv6_interface

### Overview

Manages IPv6 interface configuration including addresses, Router Advertisement (RTADV), DHCPv6, and IPv6 security filters.

### Functional Requirements

#### Core Operations

##### Create
- Configure multiple IPv6 addresses (static or prefix-based)
- Enable Router Advertisement with prefix configuration
- Set DHCPv6 service mode (server/client)
- Set IPv6 MTU
- Apply IPv6 security filters
- Save configuration

##### Read
- Retrieve all IPv6 attributes from router
- Parse RTADV configuration
- Return address list with prefix references

##### Update
- Replace address list when changed
- Update RTADV settings
- Modify DHCPv6 service mode
- Update security filters

##### Delete
- Remove all IPv6 configuration from interface
- Save configuration

### Schema Definition

| Attribute | Type | Required | ForceNew | Computed | Validation | Description |
|-----------|------|----------|----------|----------|------------|-------------|
| `interface` | string | Yes | Yes | No | `^(lan\|bridge\|pp\|tunnel)\d+$` | Interface name |
| `address` | list(block) | No | No | No | - | IPv6 address blocks |
| `address.address` | string | No | No | No | IPv6 CIDR | Static IPv6 address |
| `address.prefix_ref` | string | No | No | No | - | Prefix reference (e.g., "ra-prefix@lan2") |
| `address.interface_id` | string | No | No | No | - | Interface ID (e.g., "::1/64") |
| `rtadv` | block | No | No | No | MaxItems: 1 | Router Advertisement config |
| `rtadv.enabled` | bool | No | No | Yes | - | Enable RTADV |
| `rtadv.prefix_id` | int | Yes | No | No | >= 1 | IPv6 prefix ID to advertise |
| `rtadv.o_flag` | bool | No | No | Yes | - | Other Configuration Flag |
| `rtadv.m_flag` | bool | No | No | Yes | - | Managed Address Flag |
| `rtadv.lifetime` | int | No | No | Yes | >= 0 | Router lifetime (seconds) |
| `dhcpv6_service` | string | No | No | Yes | "server", "client", "" | DHCPv6 service mode |
| `mtu` | int | No | No | Yes | 0-65535 | IPv6 MTU (min 1280) |
| `secure_filter_in` | list(int) | No | No | No | Each >= 1 | IPv6 inbound filters |
| `secure_filter_out` | list(int) | No | No | No | Each >= 1 | IPv6 outbound filters |
| `dynamic_filter_out` | list(int) | No | No | No | Each >= 1 | IPv6 dynamic filters |

### Import Specification

- **Import ID Format**: Interface name (e.g., `lan1`)
- **Import Command**: `terraform import rtx_ipv6_interface.example lan1`

---

## Resource 3: rtx_pp_interface

### Overview

Manages IP configuration for Point-to-Point (PP) interfaces, typically used with PPPoE connections for WAN access.

### Functional Requirements

#### Core Operations

##### Create
- Configure IP address (static or "ipcp" for dynamic ISP assignment)
- Set MTU and TCP MSS limit
- Bind NAT descriptor
- Apply security filters
- Save configuration

##### Read
- Retrieve PP interface configuration
- Parse IP address, filters, and NAT settings

##### Update
- Update IP address assignment
- Modify MTU/TCP MSS
- Change NAT descriptor binding
- Update security filters

##### Delete
- Reset PP interface configuration
- Save configuration

### Schema Definition

| Attribute | Type | Required | ForceNew | Computed | Validation | Description |
|-----------|------|----------|----------|----------|------------|-------------|
| `pp_number` | int | Yes | Yes | No | >= 1 | PP interface number |
| `ip_address` | string | No | No | No | - | "ipcp" or CIDR notation |
| `mtu` | int | No | No | No | 0-1500 | MTU size |
| `tcp_mss` | int | No | No | No | 0-1460 | TCP MSS limit |
| `nat_descriptor` | int | No | No | No | >= 0 | NAT descriptor ID |
| `secure_filter_in` | list(int) | No | No | No | Each >= 1 | Inbound security filters |
| `secure_filter_out` | list(int) | No | No | No | Each >= 1 | Outbound security filters |
| `pp_interface` | string | No | No | Yes | - | Computed PP interface name (e.g., "pp1") |

### Import Specification

- **Import ID Format**: PP number as string (e.g., `1`)
- **Import Command**: `terraform import rtx_pp_interface.wan 1`

---

## Resource 4: rtx_vlan

### Overview

Manages 802.1Q VLAN interfaces on LAN ports for network segmentation.

### Functional Requirements

#### Core Operations

##### Create
- Create VLAN with specified ID on parent interface
- Automatically assign next available slot number
- Configure IP address and mask
- Set VLAN name/description
- Configure administrative shutdown state
- Save configuration

##### Read
- Retrieve VLAN configuration by interface and VLAN ID
- Return computed `vlan_interface` name (e.g., "lan1/1")
- Parse IP address and shutdown state

##### Update
- Update IP address and mask
- Change VLAN description
- Toggle shutdown state

##### Delete
- Remove VLAN interface
- Save configuration

### Schema Definition

| Attribute | Type | Required | ForceNew | Computed | Validation | Description |
|-----------|------|----------|----------|----------|------------|-------------|
| `vlan_id` | int | Yes | Yes | No | 1-4094 | VLAN ID |
| `interface` | string | Yes | Yes | No | `^lan\d+$` | Parent LAN interface |
| `name` | string | No | No | No | - | VLAN description |
| `ip_address` | string | No | No | No | IPv4 format | IP address |
| `ip_mask` | string | No | No | No | Dotted decimal | Subnet mask |
| `shutdown` | bool | No | No | Yes | - | Administrative shutdown |
| `vlan_interface` | string | No | No | Yes | - | Computed name (e.g., "lan1/1") |

### Validation Rules

- `ip_mask` is required when `ip_address` is specified
- `ip_address` is required when `ip_mask` is specified

### Import Specification

- **Import ID Format**: `interface/vlan_id` (e.g., `lan1/10`)
- **Import Command**: `terraform import rtx_vlan.guest lan1/10`

---

## Resource 5: rtx_bridge

### Overview

Manages Layer 2 bridge configurations that combine multiple interfaces into a single broadcast domain.

### Functional Requirements

#### Core Operations

##### Create
- Create bridge with specified name
- Add member interfaces
- Save configuration

##### Read
- Retrieve bridge configuration
- List member interfaces

##### Update
- Replace member interface list
- Save configuration

##### Delete
- Remove bridge configuration
- Save configuration

### Schema Definition

| Attribute | Type | Required | ForceNew | Computed | Validation | Description |
|-----------|------|----------|----------|----------|------------|-------------|
| `name` | string | Yes | Yes | No | `^bridge\d+$` | Bridge name |
| `members` | list(string) | No | No | No | Valid interface names | Member interfaces |
| `interface_name` | string | No | No | Yes | - | Computed interface name (same as `name`) |

### Valid Member Interface Patterns

- `lan\d+` - Physical LAN interfaces
- `lan\d+/\d+` - VLAN interfaces
- `tunnel\d+` - Tunnel interfaces
- `pp\d+` - PP interfaces
- `loopback\d+` - Loopback interfaces
- `bridge\d+` - Nested bridges (rare)

### Import Specification

- **Import ID Format**: Bridge name (e.g., `bridge1`)
- **Import Command**: `terraform import rtx_bridge.internal bridge1`

---

## Non-Functional Requirements

### Code Architecture and Modularity

- **Single Responsibility Principle**: Each resource file handles one resource type
- **Modular Design**: Shared validation functions across resources
- **Service Layer Separation**: Client services handle RTX communication, resources handle Terraform lifecycle
- **Clear Interfaces**: Parsers convert between RTX output and Go structs

### Performance

- SSH connection reuse across operations
- Batch operations where possible
- Efficient config parsing with regex

### Security

- Credentials handled by provider configuration
- No sensitive data stored in Terraform state beyond what's necessary
- SSH key authentication supported

### Reliability

- Idempotent operations (delete succeeds if resource already gone)
- Configuration saved after each operation
- Error detection via command output parsing

### Validation

- Interface names validated against allowed patterns
- IP addresses validated for format and range
- Filter numbers validated for allowed ranges
- Cross-field validation (e.g., ip_address requires ip_mask for VLAN)

---

## RTX Commands Reference

### rtx_interface

```
ip <interface> address <cidr>
ip <interface> address dhcp
ip <interface> description <text>
ip <interface> secure filter in <numbers...>
ip <interface> secure filter out <numbers...> [dynamic <numbers...>]
ip <interface> nat descriptor <id>
ip <interface> proxyarp on|off
ip <interface> mtu <value>
ethernet <interface> filter in <numbers...>
ethernet <interface> filter out <numbers...>
no ip <interface> address
no ip <interface> description
no ip <interface> secure filter in
no ip <interface> secure filter out
```

### rtx_ipv6_interface

```
ipv6 <interface> address <cidr>
ipv6 <interface> address <prefix-ref>::<interface-id>/<prefix>
ipv6 <interface> rtadv send <prefix-id> [o_flag=on|off] [m_flag=on|off]
ipv6 <interface> dhcp service <server|client>
ipv6 <interface> mtu <value>
ipv6 <interface> secure filter in <numbers...>
ipv6 <interface> secure filter out <numbers...> [dynamic <numbers...>]
no ipv6 <interface> address
no ipv6 <interface> rtadv send
no ipv6 <interface> dhcp service
```

### rtx_pp_interface

```
pp select <number>
  ip pp address <cidr>
  ip pp address ipcp
  ip pp mtu <value>
  ip pp tcp mss limit <value>
  ip pp nat descriptor <id>
  ip pp secure filter in <numbers...>
  ip pp secure filter out <numbers...>
pp select none
```

### rtx_vlan

```
vlan <interface>/<slot> 802.1q vid=<vlan_id>
ip <interface>/<slot> address <ip> <mask>
description <interface>/<slot> <name>
<interface>/<slot> use off
<interface>/<slot> use on
no vlan <interface>/<slot>
```

### rtx_bridge

```
bridge member <bridge-name> <interface1> [<interface2>...]
no bridge member <bridge-name>
```

---

## Terraform Command Support

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | Required | Detect drift and show planned changes |
| `terraform apply` | Required | Apply configuration changes |
| `terraform destroy` | Required | Remove resources and reset interfaces |
| `terraform import` | Required | Import existing router configuration |
| `terraform refresh` | Required | Update state from actual router config |
| `terraform state` | Required | Standard state management |

---

## Example Usage

### Complete Network Configuration

```hcl
# Internal LAN interface
resource "rtx_interface" "lan1" {
  name        = "lan1"
  description = "Internal LAN"

  ip_address {
    address = "192.168.1.1/24"
  }

  secure_filter_in  = [1, 2, 3]
  secure_filter_out = [10, 11, 12]
  proxyarp          = true
}

# IPv6 configuration for internal LAN
resource "rtx_ipv6_interface" "lan1_ipv6" {
  interface = "lan1"

  address {
    address = "2001:db8::1/64"
  }

  rtadv {
    enabled   = true
    prefix_id = 1
    o_flag    = true
  }

  dhcpv6_service = "server"
}

# PPPoE WAN interface
resource "rtx_pp_interface" "wan" {
  pp_number     = 1
  ip_address    = "ipcp"
  mtu           = 1454
  tcp_mss       = 1414
  nat_descriptor = 1000

  secure_filter_in  = [200020, 200021, 200099]
  secure_filter_out = [200020, 200021, 200099]
}

# Guest VLAN
resource "rtx_vlan" "guest" {
  vlan_id    = 10
  interface  = "lan1"
  name       = "Guest Network"
  ip_address = "192.168.10.1"
  ip_mask    = "255.255.255.0"
}

# Bridge for LAN aggregation
resource "rtx_bridge" "internal" {
  name    = "bridge1"
  members = ["lan1", "lan2"]
}
```

---

## State Handling

- Only configuration attributes are persisted in Terraform state
- Operational/runtime status (link state, traffic counters) not stored
- Computed fields (`vlan_interface`) derived from router response
- Import populates state from actual router configuration

---

## Change History

| Date | Source | Changes |
|------|--------|---------|
| 2025-01-23 | Implementation Analysis | Initial master spec from codebase analysis |
| 2026-01-25 | Implementation Sync | Add computed `interface_name` for rtx_interface/rtx_bridge, `pp_interface` for rtx_pp_interface |
