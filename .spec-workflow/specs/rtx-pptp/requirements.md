# Requirements: rtx_pptp

## Overview
Terraform resource for managing PPTP (Point-to-Point Tunneling Protocol) configuration on Yamaha RTX routers.

**Cisco Equivalent**: No direct equivalent (PPTP is deprecated in modern Cisco IOS)

## Cisco Compatibility

This resource follows general VPN naming patterns. Note that PPTP is considered legacy/deprecated for security reasons.

| RTX Attribute | Cisco Equivalent | Notes |
|---------------|------------------|-------|
| `enabled` | - | Service enable/disable |
| `shutdown` | `shutdown` | Admin state |
| `authentication` | - | Auth settings (PAP/CHAP/MS-CHAP) |

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure PPTP tunnel settings
- **Read**: Query PPTP tunnel status
- **Update**: Modify PPTP parameters
- **Delete**: Remove PPTP configuration

### 2. PPTP Server Mode
- Enable PPTP server function
- Listen address configuration
- Maximum simultaneous connections
- Authentication settings (PAP/CHAP/MS-CHAP)

### 3. PPTP Client Mode
- Remote PPTP server address
- Username and password
- Auto-reconnect settings
- Keep-alive configuration

### 4. Encryption Settings
- MPPE encryption (40/56/128-bit)
- Encryption requirement (required/optional)
- Stateless mode

### 5. IP Address Assignment
- Static IP assignment
- IP pool configuration
- DNS server assignment

### 6. PP Anonymous Configuration
- Configure anonymous PP for PPTP
- Bind to tunnel interface

### 7. Import Support
- Import existing PPTP configuration

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned PPTP configuration changes |
| `terraform apply` | ✅ Required | Create, update, or delete PPTP settings |
| `terraform destroy` | ✅ Required | Disable PPTP and remove configuration |
| `terraform import` | ✅ Required | Import existing PPTP configuration into state |
| `terraform refresh` | ✅ Required | Sync state with actual PPTP configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `pptp` (singleton resource)
- **Import Command**: `terraform import rtx_pptp.vpn_server pptp`
- **Post-Import**: All settings populated (passwords marked sensitive)

## Non-Functional Requirements

### 8. Validation
- Validate IP addresses
- Validate authentication settings
- Validate encryption options

### 9. Security
- Mark passwords as sensitive
- Note: PPTP is considered deprecated for security-sensitive applications

## RTX Commands Reference
```
pptp service on
pptp tunnel disconnect time <n>
pptp keepalive use on
pp select anonymous
pp bind tunnel<n>
pp auth accept <pap/chap/mschap/mschap-v2>
ppp ccp type mppe-any
```

## Example Usage
```hcl
# PPTP VPN Server (Legacy - consider using L2TP/IPsec or IKEv2)
resource "rtx_pptp" "vpn_server" {
  shutdown = false

  listen_address  = "0.0.0.0"
  max_connections = 10

  authentication {
    method   = "mschap-v2"
    username = "vpnuser"
    password = var.pptp_password
  }

  encryption {
    mppe_bits = 128
    required  = true
  }

  ip_pool {
    start = "192.168.200.100"
    end   = "192.168.200.200"
  }
}
```
