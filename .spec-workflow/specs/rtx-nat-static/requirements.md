# Requirements: rtx_nat_static

## Overview
Terraform resource for managing static NAT (1:1 mapping) on Yamaha RTX routers.

**Cisco Equivalent**: `iosxe_nat` (static inside source)

## Cisco Compatibility

This resource follows Cisco IOS XE Terraform provider naming conventions:

| RTX Attribute | Cisco Equivalent | Notes |
|---------------|------------------|-------|
| `inside_local` | `inside_local_address` | Inside local IP address |
| `outside_global` | `outside_global_address` | Outside global IP address |
| `protocol` | `protocol` | TCP/UDP for port NAT |
| `inside_local_port` | `inside_local_port` | Inside port number |
| `outside_global_port` | `outside_global_port` | Outside port number |

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

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned static NAT changes |
| `terraform apply` | ✅ Required | Create, update, or delete static NAT mappings |
| `terraform destroy` | ✅ Required | Remove NAT descriptor configuration |
| `terraform import` | ✅ Required | Import existing static NAT into Terraform state |
| `terraform refresh` | ✅ Required | Sync state with actual router configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `<descriptor_id>` (e.g., `10`)
- **Import Command**: `terraform import rtx_nat_static.webserver 10`
- **Post-Import**: All mappings must be populated from router state

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
# Static 1:1 NAT - Cisco-compatible naming
resource "rtx_nat_static" "webserver" {
  id = 10

  static_entries = [
    {
      inside_local   = "192.168.1.10"
      outside_global = "203.0.113.10"
    }
  ]
}

# Port-based static NAT
resource "rtx_nat_static" "port_forward" {
  id = 11

  static_entries = [
    {
      protocol           = "tcp"
      inside_local       = "192.168.1.20"
      inside_local_port  = 80
      outside_global     = "203.0.113.1"
      outside_global_port = 8080
    }
  ]
}
```
