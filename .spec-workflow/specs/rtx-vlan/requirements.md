# Requirements: rtx_vlan

## Overview
Terraform resource for managing VLAN configuration on Yamaha RTX routers.

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
resource "rtx_vlan" "management" {
  interface = "lan1"
  vlan_id   = 10
  vid       = 10

  ip_address = "192.168.10.1/24"
}

resource "rtx_vlan" "users" {
  interface = "lan1"
  vlan_id   = 20
  vid       = 20

  ip_address = "192.168.20.1/24"
}
```
