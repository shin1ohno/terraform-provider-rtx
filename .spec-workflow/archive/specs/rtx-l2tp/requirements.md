# Requirements: rtx_l2tp

## Overview
Terraform resource for managing L2TP (Layer 2 Tunneling Protocol) configuration on Yamaha RTX routers. Supports both L2TPv2 (remote access VPN) and L2TPv3 (site-to-site L2VPN).

### L2TP Versions Supported

| Version | Use Case | Description |
|---------|----------|-------------|
| **L2TPv2** | Remote Access VPN | L2TP/IPsec for smartphone/PC clients connecting to LAN. Server (LNS) mode only. |
| **L2TPv3** | Site-to-Site L2VPN | Ethernet frame tunneling between routers. Enables same L2 segment across multiple sites. |

**Cisco Equivalent**: `iosxe_interface_tunnel` with tunnel mode l2tp, `iosxe_aaa` for authentication

## Cisco Compatibility

This resource follows general VPN naming patterns where applicable:

| RTX Attribute | Cisco Equivalent | Notes |
|---------------|------------------|-------|
| `name` | `name` | Tunnel interface name |
| `mode` | `tunnel_mode` | Tunnel mode (l2tp) |
| `source` | `tunnel_source` | Local tunnel endpoint |
| `destination` | `tunnel_destination` | Remote tunnel endpoint |
| `shutdown` | `shutdown` | Admin state |

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure L2TP tunnel settings
- **Read**: Query L2TP tunnel status
  - Tunnel status is operational-only and MUST NOT be persisted in Terraform state
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

### 7. L2TPv3 Site-to-Site Configuration
- Local and remote router IDs
- Remote end ID (optional)
- Always-on connection mode
- Bridge interface binding
- IPsec transport mode (IKEv1)

### 8. Import Support
- Import existing L2TP configuration

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned L2TP configuration changes |
| `terraform apply` | ✅ Required | Create, update, or delete L2TP settings |
| `terraform destroy` | ✅ Required | Remove L2TP configuration cleanly |
| `terraform import` | ✅ Required | Import existing L2TP configuration into state |
| `terraform refresh` | ✅ Required | Sync state with actual L2TP configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `<tunnel_id>` (e.g., `1`)
- **Import Command**: `terraform import rtx_l2tp.vpn_server 1`
- **Post-Import**: All settings populated (credentials marked sensitive)

## Non-Functional Requirements

### 9. Validation
- Validate IP addresses
- Validate authentication credentials
- Validate tunnel parameters

### 10. Security
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
# L2TPv2 VPN Server (Remote Access)
resource "rtx_l2tp" "vpn_server" {
  name     = "L2TP_VPN"
  version  = "l2tp"  # L2TPv2
  mode     = "lns"   # L2TP Network Server
  shutdown = false

  tunnel_source      = "203.0.113.1"
  tunnel_destination = "0.0.0.0"  # Accept from any

  authentication {
    method   = "chap"
    username = "vpnuser"
    password = var.l2tp_password
  }

  ip_pool {
    start = "192.168.100.100"
    end   = "192.168.100.200"
  }

  # L2TP over IPsec
  ipsec_profile {
    enabled        = true
    pre_shared_key = var.ipsec_psk
  }
}

# L2TPv3 Site-to-Site L2VPN
resource "rtx_l2tp" "l2vpn_site_b" {
  name     = "L2VPN_SITE_B"
  version  = "l2tpv3"
  mode     = "l2vpn"
  shutdown = false

  l2tpv3_config {
    local_router_id  = "203.0.113.1"
    remote_router_id = "198.51.100.1"
    always_on        = true
    bridge_interface = "bridge1"
  }

  # IPsec transport mode (optional)
  ipsec_profile {
    enabled        = true
    pre_shared_key = var.ipsec_psk
  }
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state.
- Operational/runtime status must not be stored in state.
