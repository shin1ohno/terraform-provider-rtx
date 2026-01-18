# Requirements: rtx_vlan

## Overview
Terraform resource for managing VLAN configuration on Yamaha RTX routers.

**Cisco Equivalent**: `iosxe_vlan`

## Cisco Compatibility

This resource follows Cisco IOS XE Terraform provider naming conventions:

| RTX Attribute | Cisco Equivalent | Notes |
|---------------|------------------|-------|
| `vlan_id` | `vlan_id` | VLAN identifier (1-4094) |
| `name` | `name` | VLAN name/description |
| `shutdown` | `shutdown` | Admin state (true = disabled) |
| `interface` | - | Parent interface (RTX-specific) |

## Functional Requirements

### 1. CRUD Operations
- **Create**: Create VLAN interface and configuration
- **Read**: Query VLAN configuration
- **Update**: Modify VLAN parameters
- **Delete**: Remove VLAN configuration

### 2. VLAN Interface
- VLAN ID (1-4094)
- Parent physical interface
- VLAN interface naming (e.g., `vlan1`)

### 3. Port Assignment
- Assign physical ports to VLANs
- Tagged (trunk) port configuration
- Untagged (access) port configuration

### 4. VLAN IP Configuration
- IP address assignment to VLAN interface
- DHCP client on VLAN interface
- Secondary IP addresses

### 5. Inter-VLAN Routing
- Enable routing between VLANs
- VLAN-based access control

### 6. Import Support
- Import existing VLAN by ID

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned VLAN configuration changes |
| `terraform apply` | ✅ Required | Create, update, or delete VLAN interfaces |
| `terraform destroy` | ✅ Required | Remove VLAN configuration from router |
| `terraform import` | ✅ Required | Import existing VLANs into Terraform state |
| `terraform refresh` | ✅ Required | Sync state with actual VLAN configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `<interface>/<vlan_id>` (e.g., `lan1/10`)
- **Import Command**: `terraform import rtx_vlan.management lan1/10`
- **Post-Import**: IP address and all VLAN settings must be populated

## Non-Functional Requirements

### 7. Validation
- Validate VLAN ID range (1-4094)
- Validate port availability
- Prevent duplicate VLAN IDs

### 8. Dependencies
- Depend on physical interface configuration

## RTX Commands Reference
```
vlan <interface>/<vlan_id> 802.1q vid=<vid>
ip <vlan_interface> address <ip/mask>
vlan port mapping <interface> <vlan_list>
```

## Example Usage
```hcl
# VLAN definition - Cisco-compatible naming
resource "rtx_vlan" "management" {
  vlan_id   = 10
  name      = "Management"
  shutdown  = false

  # RTX-specific: parent interface
  interface = "lan1"

  ip_address = "192.168.10.1"
  ip_mask    = "255.255.255.0"
}

resource "rtx_vlan" "users" {
  vlan_id   = 20
  name      = "Users"
  shutdown  = false

  interface = "lan1"

  ip_address = "192.168.20.1"
  ip_mask    = "255.255.255.0"
}
```
