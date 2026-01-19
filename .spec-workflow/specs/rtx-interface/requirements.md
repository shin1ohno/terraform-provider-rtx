# Requirements: rtx_interface

## Overview
Terraform resource for managing network interface configurations on Yamaha RTX routers. This includes IP address assignment, security filters, NAT descriptors, and other interface-level settings.

**Cisco Equivalent**: `iosxe_interface`

## Cisco Compatibility

This resource follows Cisco IOS XE Terraform provider naming conventions:

| RTX Attribute | Cisco Equivalent | Notes |
|---------------|------------------|-------|
| `name` | `name` | Interface identifier (lan1, lan2, etc.) |
| `description` | `description` | Interface description |
| `ip_address` | `ipv4_address` | IP address configuration |
| `mtu` | `mtu` | Maximum transmission unit |
| `shutdown` | `shutdown` | Admin state (RTX: always up) |

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure interface settings using RTX `ip` commands
- **Read**: Query current interface configuration
- **Update**: Modify interface parameters
- **Delete**: Remove custom interface configuration (reset to defaults)

### 2. IP Address Configuration
- Static IP address with CIDR notation
- DHCP client mode
- Secondary IP addresses (if supported)

### 3. Security Filter Application
- Inbound security filters (`secure filter in`)
- Outbound security filters (`secure filter out`)
- Dynamic outbound filters for stateful inspection

### 4. NAT Descriptor Binding
- Bind NAT descriptor to interface
- Support for masquerade and static NAT

### 5. ProxyARP Configuration
- Enable/disable ProxyARP on interface
- Required for certain bridging scenarios

### 6. MTU Configuration
- Configure interface MTU size
- Validate against minimum/maximum limits

### 7. Import Support
- Import existing interface configuration by name

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned interface configuration changes |
| `terraform apply` | ✅ Required | Create, update, or delete interface settings |
| `terraform destroy` | ✅ Required | Reset interface to default configuration |
| `terraform import` | ✅ Required | Import existing interface into Terraform state |
| `terraform refresh` | ✅ Required | Sync state with actual interface configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `<interface_name>` (e.g., `lan1`, `lan2`)
- **Import Command**: `terraform import rtx_interface.wan lan2`
- **Post-Import**: All attributes including filters and NAT descriptor must be populated

## Non-Functional Requirements

### 8. Validation
- Validate interface name format (lan1, lan2, bridge1, pp1, tunnel1)
- Validate IP address CIDR notation
- Validate filter numbers are positive integers
- Validate NAT descriptor is positive integer
- Validate MTU is within valid range

### 9. Dependencies
- Security filters (use `rtx_ip_filter` to define)
- NAT descriptors (use `rtx_nat_masquerade` or `rtx_nat_static` to define)

## RTX Commands Reference
```
ip <interface> address <ip>/<prefix>
ip <interface> address dhcp
ip <interface> secure filter in <filter_list>
ip <interface> secure filter out <filter_list> [dynamic <dynamic_filter_list>]
ip <interface> nat descriptor <descriptor_id>
ip <interface> proxyarp on|off
ip <interface> mtu <size>
description <interface> <description>
```

## Example Usage
```hcl
# WAN interface with DHCP and security filters
resource "rtx_interface" "wan" {
  name        = "lan2"
  description = "WAN connection"

  ip_address {
    dhcp = true
  }

  secure_filter_in  = [200020, 200021, 200022, 200099]
  secure_filter_out = [200020, 200021, 200099]
  dynamic_filter_out = [200080, 200081, 200082]

  nat_descriptor = 1000
}

# LAN interface with static IP
resource "rtx_interface" "lan" {
  name = "lan1"

  ip_address {
    address = "192.168.1.1/24"
  }

  proxyarp = true
}

# Bridge interface
resource "rtx_interface" "bridge" {
  name        = "bridge1"
  description = "Internal bridge"

  ip_address {
    address = "192.168.1.253/16"
  }
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state.
- Operational/runtime status must not be stored in state.
