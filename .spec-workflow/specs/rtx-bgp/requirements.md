# Requirements: rtx_bgp

## Overview
Terraform resource for managing BGP (Border Gateway Protocol) routing configuration on Yamaha RTX routers.

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure BGP routing process
- **Read**: Query BGP configuration and peer status
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
resource "rtx_bgp" "main" {
  enabled   = true
  as_number = 65001
  router_id = "1.1.1.1"

  neighbor {
    id             = 1
    address        = "203.0.113.1"
    remote_as      = 65002
    hold_time      = 90
    keepalive_time = 30
  }

  neighbor {
    id             = 2
    address        = "198.51.100.1"
    remote_as      = 65003
    multihop       = 2
    password       = var.bgp_password
  }

  network {
    prefix = "192.168.0.0/16"
  }

  redistribute {
    protocol = "static"
  }
}
```
