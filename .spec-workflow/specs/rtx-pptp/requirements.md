# Requirements: rtx_pptp

## Overview
Terraform resource for managing PPTP (Point-to-Point Tunneling Protocol) configuration on Yamaha RTX routers.

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
resource "rtx_pptp" "vpn_server" {
  enabled = true

  listen_address = "0.0.0.0"
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
