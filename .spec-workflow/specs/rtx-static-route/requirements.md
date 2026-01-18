# Requirements: rtx_static_route

## Overview
Terraform resource for managing static IP routes on Yamaha RTX routers.

## Functional Requirements

### 1. CRUD Operations
- **Create**: Add static route entries using `ip route` command
- **Read**: Query current routing table to verify route existence
- **Update**: Modify route parameters (gateway, weight, filters)
- **Delete**: Remove routes using `no ip route` command

### 2. Route Destinations
- Support `default` route (default gateway)
- Support network/prefix notation (e.g., `192.168.1.0/24`)
- Support host routes with /32 prefix

### 3. Gateway Types
- IP address gateway (e.g., `192.168.0.1`)
- PP interface gateway (e.g., `pp 1`)
- Tunnel interface gateway (e.g., `tunnel 1`)
- DHCP-derived gateway (e.g., `dhcp lan1`)
- NULL interface (for blackhole routes)
- Loopback interface

### 4. Route Parameters
- `weight`: Load balancing weight (default: 1)
- `hide`: Hide route when gateway is unreachable
- `keepalive`: Enable keepalive for route monitoring
- `filter`: Apply IP filter to route

### 5. Multi-path Routing
- Support multiple gateways for load balancing
- Support failover with weight=0

### 6. Import Support
- Import existing routes by network/prefix
- Handle default route import

## Non-Functional Requirements

### 7. Validation
- Validate IP address and CIDR notation
- Validate gateway interface existence
- Validate weight range (0-100)

### 8. State Management
- Track route state in Terraform
- Handle external route modifications gracefully

## RTX Commands Reference
```
ip route <network> gateway <gateway> [weight <n>] [hide] [filter <n>]
no ip route <network>
show ip route
```

## Example Usage
```hcl
resource "rtx_static_route" "default" {
  network = "default"
  gateway {
    address = "192.168.0.1"
  }
}

resource "rtx_static_route" "internal" {
  network = "10.0.0.0/8"
  gateway {
    interface = "tunnel 1"
    weight    = 1
    hide      = true
  }
}
```
