# Requirements: rtx_snmp

## Overview
Terraform resources for managing SNMP (Simple Network Management Protocol) configuration on Yamaha RTX routers.

**Cisco Equivalent**: `iosxe_snmp_server`

## Cisco Compatibility

These resources follow Cisco SNMP server naming patterns:

| RTX Attribute | Cisco Equivalent | Notes |
|---------------|------------------|-------|
| `community` | `community` | SNMP community string |
| `location` | `location` | System location |
| `contact` | `contact` | System contact |
| `host` | `host` | Trap destination |
| `enable_traps` | `enable_traps` | Enable specific traps |
| `view` | `view` | SNMP view configuration |

## Covered Resources

This specification covers two Terraform resources:

- **`rtx_snmp_server`**: SNMP agent, communities, trap destinations
- **`rtx_snmp_server_user`**: SNMPv3 user configuration

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure SNMP agent settings
- **Read**: Query SNMP configuration
- **Update**: Modify SNMP settings
- **Delete**: Disable SNMP and remove configuration

### 2. SNMP Agent Configuration
- Enable/disable SNMP agent
- System contact information
- System location
- System name

### 3. SNMP v1/v2c Configuration
- Community string (read-only)
- Community string (read-write)
- Access control by source IP

### 4. SNMP v3 Configuration
- User authentication (MD5, SHA)
- Privacy/encryption (DES, AES)
- Security levels (noAuthNoPriv, authNoPriv, authPriv)
- Engine ID

### 5. SNMP Traps
- Trap destination (manager IP)
- Trap community string
- Trap types (link up/down, auth failure, etc.)
- Trap version (v1, v2c, v3)

### 6. Access Control
- Allowed hosts/networks
- View configuration
- Access mode (read-only, read-write)

### 7. Import Support
- Import existing SNMP configuration

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned SNMP configuration changes |
| `terraform apply` | ✅ Required | Create, update, or delete SNMP settings |
| `terraform destroy` | ✅ Required | Disable SNMP agent and remove configuration |
| `terraform import` | ✅ Required | Import existing SNMP configuration into state |
| `terraform refresh` | ✅ Required | Sync state with actual SNMP configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `snmp` (singleton resource)
- **Import Command**: `terraform import rtx_snmp.monitoring snmp`
- **Post-Import**: All settings populated (community strings marked sensitive)

## Non-Functional Requirements

### 8. Validation
- Validate IP addresses
- Validate community string format
- Validate SNMPv3 credentials

### 9. Security
- Mark community strings as sensitive
- Mark SNMPv3 credentials as sensitive

## RTX Commands Reference
```
snmp host <ip> [community <string>]
snmp community read-only <string>
snmp community read-write <string>
snmp trap community <string>
snmp trap host <ip>
snmp trap enable snmp [all|<trap_type>]
snmp sysname <name>
snmp syslocation <location>
snmp syscontact <contact>
```

## Example Usage
```hcl
# SNMP server configuration - Cisco-compatible naming
resource "rtx_snmp_server" "monitoring" {
  location = "Tokyo DC Rack 42"
  contact  = "noc@example.com"

  # Community strings
  communities = [
    {
      name       = var.snmp_community_ro
      permission = "ro"
      acl        = "SNMP_ACCESS"
    },
    {
      name       = var.snmp_community_rw
      permission = "rw"
      acl        = "SNMP_ACCESS"
    }
  ]

  # Trap destinations
  hosts = [
    {
      address   = "10.0.0.100"
      community = var.snmp_trap_community
      version   = "2c"
    }
  ]

  # Enable traps
  enable_traps = ["snmp", "linkdown", "linkup"]
}

# SNMPv3 user
resource "rtx_snmp_server_user" "admin" {
  username = "snmpadmin"
  group    = "ADMIN_GROUP"

  auth_protocol = "sha"
  auth_password = var.snmpv3_auth_password

  priv_protocol = "aes128"
  priv_password = var.snmpv3_priv_password
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state.
- Operational/runtime status must not be stored in state.
