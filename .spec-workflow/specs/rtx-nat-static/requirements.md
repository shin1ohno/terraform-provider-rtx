# Requirements: rtx_nat_static

## Overview
Terraform resource for managing static NAT (1:1 mapping) on Yamaha RTX routers.

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure static NAT using `nat descriptor` commands
- **Read**: Query NAT descriptor configuration
- **Update**: Modify NAT mapping
- **Delete**: Remove NAT descriptor

### 2. NAT Descriptor
- Descriptor ID (1-65535)
- NAT type: static (one-to-one)
- Outer IP address
- Inner IP address

### 3. Address Mapping
- 1:1 IP address translation
- Bidirectional translation
- Support for multiple static mappings in one descriptor

### 4. Port Static NAT
- Specific port mapping (outer:port -> inner:port)
- Protocol specification (TCP/UDP)

### 5. Interface Binding
- Apply NAT descriptor to interface
- Direction configuration

### 6. Import Support
- Import existing static NAT by descriptor ID

## Non-Functional Requirements

### 7. Validation
- Validate IP address formats
- Validate descriptor ID uniqueness
- Prevent overlapping mappings

## RTX Commands Reference
```
nat descriptor type <id> static
nat descriptor static <id> <outer_ip>=<inner_ip>
nat descriptor static <id> <outer_ip>:<port>=<inner_ip>:<port> <protocol>
ip <interface> nat descriptor <id>
```

## Example Usage
```hcl
resource "rtx_nat_static" "webserver" {
  descriptor_id = 10

  mapping {
    outer_address = "203.0.113.10"
    inner_address = "192.168.1.10"
  }
}

resource "rtx_nat_static" "port_forward" {
  descriptor_id = 11

  port_mapping {
    protocol      = "tcp"
    outer_address = "203.0.113.1"
    outer_port    = 8080
    inner_address = "192.168.1.20"
    inner_port    = 80
  }
}
```
