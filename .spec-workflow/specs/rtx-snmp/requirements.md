# Requirements: rtx_snmp

## Overview
Terraform resource for managing SNMP (Simple Network Management Protocol) configuration on Yamaha RTX routers.

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
resource "rtx_snmp" "monitoring" {
  enabled = true

  system_name     = "rtx-router-01"
  system_location = "Tokyo DC Rack 42"
  system_contact  = "noc@example.com"

  community_ro = var.snmp_community_ro
  community_rw = var.snmp_community_rw

  allowed_hosts = ["10.0.0.0/8"]

  trap {
    host      = "10.0.0.100"
    community = var.snmp_trap_community
    version   = "2c"

    enabled_traps = ["linkDown", "linkUp", "authenticationFailure"]
  }
}

resource "rtx_snmp_v3_user" "admin" {
  username = "snmpadmin"

  auth_protocol = "sha"
  auth_password = var.snmpv3_auth_password

  priv_protocol = "aes"
  priv_password = var.snmpv3_priv_password
}
```
