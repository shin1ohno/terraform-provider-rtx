# Requirements: rtx_bridge

## Overview
Terraform resource for managing bridge interface configurations on Yamaha RTX routers. Bridges allow multiple network segments (physical interfaces, tunnels) to operate as a single Layer 2 broadcast domain.

## Functional Requirements

### 1. CRUD Operations
- **Create**: Create bridge interface with member interfaces
- **Read**: Query bridge configuration and member list
- **Update**: Add or remove member interfaces
- **Delete**: Remove bridge and release member interfaces

### 2. Bridge Creation
- Create named bridge interfaces (bridge1, bridge2, etc.)
- RTX typically supports up to 8 bridges

### 3. Member Interface Management
- Add physical LAN interfaces (lan1, lan2, etc.)
- Add L2TPv3 tunnel interfaces (tunnel1, tunnel2, etc.)
- Add PP interfaces (pp1, pp2, etc.)
- Remove interfaces from bridge

### 4. Layer 2 VPN Support
- Support L2TPv3 tunnels as bridge members
- Enable multi-site L2 bridging

### 5. Import Support
- Import existing bridge configuration by name

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned bridge configuration changes |
| `terraform apply` | ✅ Required | Create, update, or delete bridge |
| `terraform destroy` | ✅ Required | Remove bridge and release members |
| `terraform import` | ✅ Required | Import existing bridge into state |
| `terraform refresh` | ✅ Required | Sync state with actual configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `<bridge_name>` (e.g., `bridge1`)
- **Import Command**: `terraform import rtx_bridge.main bridge1`
- **Post-Import**: All member interfaces must be populated

## Non-Functional Requirements

### 6. Validation
- Validate bridge name format (bridge1, bridge2, etc.)
- Validate member interface names
- Prevent duplicate member assignment across bridges

### 7. Dependencies
- L2TPv3 tunnels must be created before adding as bridge members
- IP address configuration handled by `rtx_interface`

## RTX Commands Reference
```
bridge member <bridge_name> <member1> [<member2> ...]
no bridge member <bridge_name>
show config | grep bridge
show status bridge
```

## Example Usage
```hcl
# Simple bridge with LAN interface
resource "rtx_bridge" "main" {
  name = "bridge1"

  members = ["lan1"]
}

# L2VPN bridge with multiple sites
resource "rtx_bridge" "l2vpn" {
  name = "bridge2"

  members = [
    "lan1",
    "tunnel1",
    "tunnel2",
  ]
}

# IP configuration for bridge (separate resource)
resource "rtx_interface" "bridge_ip" {
  name = rtx_bridge.main.name

  ip_address {
    address = "192.168.1.253/16"
  }
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state.
- Operational/runtime status must not be stored in state.
