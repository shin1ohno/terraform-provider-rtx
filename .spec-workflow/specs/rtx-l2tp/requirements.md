# Requirements: rtx_l2tp

## Overview
Terraform resource for managing L2TP (Layer 2 Tunneling Protocol) configuration on Yamaha RTX routers.

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure L2TP tunnel settings
- **Read**: Query L2TP tunnel status
- **Update**: Modify L2TP parameters
- **Delete**: Remove L2TP configuration

### 2. L2TP Tunnel Configuration
- Tunnel endpoint (local/remote)
- Tunnel type (LAC/LNS)
- Always-on or on-demand connection

### 3. L2TP Server (LNS) Mode
- Accept incoming L2TP connections
- IP address pool for clients
- Authentication settings (PAP/CHAP)
- Maximum simultaneous connections

### 4. L2TP Client (LAC) Mode
- Remote LNS address
- Username and password
- Auto-reconnect settings

### 5. IPsec Integration
- L2TP over IPsec configuration
- Pre-shared key or certificate
- Encryption settings

### 6. PP Anonymous Configuration
- Configure anonymous PP for L2TP
- IP address assignment

### 7. Import Support
- Import existing L2TP configuration

## Non-Functional Requirements

### 8. Validation
- Validate IP addresses
- Validate authentication credentials
- Validate tunnel parameters

### 9. Security
- Mark passwords as sensitive
- Secure credential handling

## RTX Commands Reference
```
pp select anonymous
pp bind tunnel<n>
pp auth accept <pap/chap>
pp auth myname <name> <password>
ppp ipcp ipaddress on
l2tp tunnel disconnect time <n>
l2tp keepalive use on
l2tp keepalive log on
```

## Example Usage
```hcl
resource "rtx_l2tp" "vpn_server" {
  mode = "lns"

  tunnel_id = 1

  local_address = "203.0.113.1"

  authentication {
    method   = "chap"
    username = "vpnuser"
    password = var.l2tp_password
  }

  ip_pool {
    start = "192.168.100.100"
    end   = "192.168.100.200"
  }

  ipsec_enabled = true
  pre_shared_key = var.ipsec_psk
}
```
