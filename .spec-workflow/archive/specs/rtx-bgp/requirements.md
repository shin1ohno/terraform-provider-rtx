# Requirements: rtx_bgp

## Overview
Terraform resource for managing BGP (Border Gateway Protocol) routing configuration on Yamaha RTX routers.

**Cisco Equivalent**: `iosxe_bgp`

## Cisco Compatibility

This resource follows Cisco IOS XE Terraform provider naming conventions:

| RTX Attribute | Cisco Equivalent | Notes |
|---------------|------------------|-------|
| `asn` | `asn` | Local AS number (string for 4-byte ASN) |
| `router_id` | - | Router ID (configure via iosxe_bgp_neighbor) |
| `default_ipv4_unicast` | `default_ipv4_unicast` | Enable IPv4 unicast by default |
| `log_neighbor_changes` | `log_neighbor_changes` | Log neighbor state changes |
| `neighbors` | - | Neighbors (use iosxe_bgp_neighbor) |

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure BGP routing process
- **Read**: Query BGP configuration and peer status
  - Peer status is operational-only and MUST NOT be persisted in Terraform state
- **Update**: Modify BGP parameters
- **Delete**: Remove BGP configuration

### 2. BGP Process Configuration
- Enable/disable BGP
- Local AS number (1-65535, or 4-byte ASN)
- Router ID
- Default local preference

### 3. Neighbor (Peer) Configuration
- Peer IP address
- Remote AS number
- eBGP or iBGP session
- Multihop configuration
- Timers (keepalive, hold)
- Password authentication (MD5)

### 4. Address Family Configuration
- IPv4 unicast
- IPv6 unicast
- Network announcements

### 5. Route Filtering
- Prefix lists
- AS path filters
- Route maps
- Community filters

### 6. Route Redistribution
- Redistribute static routes
- Redistribute connected routes
- Redistribute from OSPF

### 7. Advanced Features
- Route reflector configuration
- Confederation settings
- Graceful restart

### 8. Import Support
- Import existing BGP configuration

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned BGP configuration changes |
| `terraform apply` | ✅ Required | Create, update, or delete BGP settings |
| `terraform destroy` | ✅ Required | Disable BGP and remove all neighbor/network configs |
| `terraform import` | ✅ Required | Import existing BGP configuration into state |
| `terraform refresh` | ✅ Required | Sync state with actual BGP configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `bgp` (singleton resource)
- **Import Command**: `terraform import rtx_bgp.main bgp`
- **Post-Import**: All neighbors, networks, and filters populated (passwords sensitive)

## Non-Functional Requirements

### 9. Validation
- Validate AS number range
- Validate IP addresses
- Validate timer values

### 10. Security
- Mark peer passwords as sensitive

## RTX Commands Reference
```
bgp use on
bgp autonomous-system <asn>
bgp router id <router_id>
bgp neighbor <n> address <ip> as <asn>
bgp neighbor <n> hold-time <time>
bgp neighbor <n> local-address <ip>
bgp import filter <n> include <network>
bgp import from static
```

## Example Usage
```hcl
# BGP configuration - Cisco-compatible naming
resource "rtx_bgp" "main" {
  asn                  = "65001"
  router_id            = "1.1.1.1"
  default_ipv4_unicast = true
  log_neighbor_changes = true

  neighbors = [
    {
      ip        = "203.0.113.1"
      remote_as = "65002"
      hold_time = 90
      keepalive = 30
    },
    {
      ip        = "198.51.100.1"
      remote_as = "65003"
      multihop  = 2
      password  = var.bgp_password
    }
  ]

  networks = [
    {
      prefix = "192.168.0.0"
      mask   = "255.255.0.0"
    }
  ]

  redistribute_static = true
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state.
- Operational/runtime status must not be stored in state.
