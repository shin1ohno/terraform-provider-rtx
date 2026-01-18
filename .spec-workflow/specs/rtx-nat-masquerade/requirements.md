# Requirements: rtx_nat_masquerade

## Overview
Terraform resource for managing NAT masquerade (dynamic NAPT) on Yamaha RTX routers.

**Cisco Equivalent**: `iosxe_nat` (inside source with overload)

## Cisco Compatibility

This resource follows Cisco IOS XE Terraform provider naming conventions:

| RTX Attribute | Cisco Equivalent | Notes |
|---------------|------------------|-------|
| `id` | `id` | NAT descriptor ID |
| `inside_source` | `inside_source_interfaces` | Inside source configuration |
| `interface` | `interface` | Outside interface |
| `overload` | `overload` | Enable PAT (always true for masquerade) |
| `acl` | `access_list` | Source access list |

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

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned NAT changes without applying |
| `terraform apply` | ✅ Required | Create, update, or delete NAT masquerade configuration |
| `terraform destroy` | ✅ Required | Remove NAT descriptor and interface bindings |
| `terraform import` | ✅ Required | Import existing NAT descriptors into Terraform state |
| `terraform refresh` | ✅ Required | Sync state with actual router NAT configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `<descriptor_id>` (e.g., `1`)
- **Import Command**: `terraform import rtx_nat_masquerade.main 1`
- **Post-Import**: All attributes including static mappings must be populated

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
# Basic NAT masquerade - Cisco-compatible naming
resource "rtx_nat" "main" {
  id = 1

  inside_source {
    acl       = "192.168.1.0/24"
    interface = "pp1"
    overload  = true
  }
}

# NAT with static port mapping
resource "rtx_nat" "with_mapping" {
  id = 2

  inside_source {
    acl       = "192.168.2.0/24"
    interface = "pp1"
    overload  = true
  }

  static_entries = [
    {
      inside_local   = "192.168.2.10:443"
      outside_global = "203.0.113.1:443"
      protocol       = "tcp"
    }
  ]
}
```
