# Requirements: rtx_nat_masquerade

## Overview
Terraform resource for managing NAT masquerade (dynamic NAPT) on Yamaha RTX routers.

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure NAT masquerade using `nat descriptor` commands
- **Read**: Query NAT descriptor configuration
- **Update**: Modify NAT descriptor parameters
- **Delete**: Remove NAT descriptor configuration

### 2. NAT Descriptor
- Descriptor ID (1-65535)
- NAT type: masquerade (many-to-one)
- Outer IP address (interface or specific IP)
- Inner network range

### 3. Port Translation
- Automatic port translation for outbound connections
- Port range configuration
- Protocol-specific settings (TCP/UDP/ICMP)

### 4. Interface Binding
- Apply NAT descriptor to interface
- Support LAN, PP, and tunnel interfaces
- Direction: in/out

### 5. Static Port Mapping
- Map external port to internal host:port
- Support TCP and UDP protocols
- Port range mapping

### 6. Import Support
- Import existing NAT descriptors by ID

## Non-Functional Requirements

### 7. Validation
- Validate descriptor ID uniqueness
- Validate IP address formats
- Validate port ranges

### 8. Logging
- Support NAT session logging configuration

## RTX Commands Reference
```
nat descriptor type <id> masquerade
nat descriptor address outer <id> <interface/ip>
nat descriptor address inner <id> <network>
ip <interface> nat descriptor <id>
show nat descriptor address
```

## Example Usage
```hcl
resource "rtx_nat_masquerade" "main" {
  descriptor_id = 1
  outer_address = "pp1"  # Use PP1 interface address
  inner_network = "192.168.1.0/24"
}

resource "rtx_nat_masquerade" "with_mapping" {
  descriptor_id = 2
  outer_address = "203.0.113.1"
  inner_network = "192.168.2.0/24"

  static_mapping {
    protocol      = "tcp"
    outer_port    = 443
    inner_address = "192.168.2.10"
    inner_port    = 443
  }
}
```
