# Requirements: rtx_static_route

## Overview
Terraform resource for managing static IP routes on Yamaha RTX routers.

**Cisco Equivalent**: `iosxe_static_route`

## Cisco Compatibility

This resource follows Cisco IOS XE Terraform provider naming conventions:

| RTX Attribute | Cisco Equivalent | Notes |
|---------------|------------------|-------|
| `prefix` | `prefix` | Route destination network |
| `mask` | `mask` | Subnet mask (dotted decimal) |
| `next_hops` | `next_hops` | List of next hop configurations |
| `next_hop` | `next_hop` | Gateway IP address |
| `distance` | `distance` | Administrative distance / weight |
| `name` | `name` | Route description |
| `permanent` | `permanent` | Keep route even if interface down |

## Functional Requirements

### 1. CRUD Operations
- **Create**: Add static route entries using `ip route` command
- **Read**: Query current routing table to verify route existence
- **Update**: Modify route parameters (next_hop, distance, filters)
- **Delete**: Remove routes using `no ip route` command

### 2. Route Destinations
- Support `default` route (prefix=0.0.0.0, mask=0.0.0.0)
- Support network/prefix notation via `prefix` and `mask`
- Support host routes with mask=255.255.255.255

### 3. Next Hop Types
- IP address next_hop (e.g., `192.168.0.1`)
- PP interface (e.g., `pp 1`) via `interface` attribute
- Tunnel interface (e.g., `tunnel 1`) via `interface` attribute
- DHCP-derived gateway (e.g., `dhcp lan1`)
- NULL interface (for blackhole routes)
- Loopback interface

### 4. Route Parameters
- `distance`: Administrative distance / weight (default: 1)
- `permanent`: Keep route even when next_hop is unreachable (Cisco: `permanent`)
- `name`: Route description
- `filter`: Apply IP filter to route (RTX-specific)

### 5. Multi-path Routing
- Support multiple next_hops for load balancing
- Support failover with distance configuration

### 6. Import Support
- Import existing routes by prefix/mask
- Handle default route import

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned changes without applying; accurately diff current vs desired state |
| `terraform apply` | ✅ Required | Create, update, or delete routes on the RTX router |
| `terraform destroy` | ✅ Required | Remove all managed routes cleanly |
| `terraform import` | ✅ Required | Import existing routes into Terraform state |
| `terraform refresh` | ✅ Required | Sync Terraform state with actual router configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `<network>` (e.g., `default`, `192.168.1.0/24`)
- **Import Command**: `terraform import rtx_static_route.example 192.168.1.0/24`
- **Post-Import**: All attributes must be populated from router state

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
# Default route - Cisco-compatible syntax
resource "rtx_static_route" "default" {
  prefix = "0.0.0.0"
  mask   = "0.0.0.0"

  next_hops = [
    {
      next_hop  = "192.168.0.1"
      distance  = 1
      name      = "primary_gateway"
      permanent = true
    }
  ]
}

# Multi-path routing with failover
resource "rtx_static_route" "internal" {
  prefix = "10.0.0.0"
  mask   = "255.0.0.0"

  next_hops = [
    {
      next_hop = "192.168.1.1"
      distance = 10
      name     = "primary"
    },
    {
      next_hop = "192.168.2.1"
      distance = 20
      name     = "backup"
    }
  ]
}

# Tunnel interface route
resource "rtx_static_route" "vpn_route" {
  prefix = "172.16.0.0"
  mask   = "255.255.0.0"

  next_hops = [
    {
      interface = "tunnel 1"
      distance  = 1
      permanent = true
    }
  ]
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state.
- Operational/runtime status must not be stored in state.
