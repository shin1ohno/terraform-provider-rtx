# Requirements: rtx_ipsec_tunnel

## Overview
Terraform resource for managing IPsec VPN tunnels on Yamaha RTX routers.

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure IPsec tunnel with IKE settings
- **Read**: Query tunnel and SA status
- **Update**: Modify tunnel parameters
- **Delete**: Remove tunnel configuration

### 2. Tunnel Configuration
- Tunnel ID (1-100+)
- Tunnel mode (tunnel/transport)
- Local and remote endpoints
- Encryption domain (interesting traffic)

### 3. IKE Phase 1 (Main Mode / Aggressive Mode)
- IKE version (IKEv1/IKEv2)
- Pre-shared key or certificate authentication
- Encryption algorithm (AES, 3DES)
- Hash algorithm (SHA-256, SHA-1, MD5)
- DH group (1, 2, 5, 14, 15, 16)
- SA lifetime

### 4. IKE Phase 2 (Quick Mode)
- IPsec protocol (ESP/AH)
- Encryption algorithm
- Authentication algorithm
- PFS group
- SA lifetime

### 5. Peer Configuration
- Remote peer IP address or FQDN
- Local identity
- Remote identity
- NAT-T settings

### 6. Keepalive and DPD
- Dead Peer Detection (DPD)
- Keepalive interval
- Retry count

### 7. Import Support
- Import existing tunnel by ID

## Non-Functional Requirements

### 8. Validation
- Validate IP addresses and FQDNs
- Validate algorithm combinations
- Validate key formats

### 9. Security
- Secure handling of pre-shared keys
- Mark PSK as sensitive in state

## RTX Commands Reference
```
tunnel select <n>
ipsec tunnel <n>
ipsec sa policy <n> <tunnel> esp aes-cbc sha-hmac
ipsec ike pre-shared-key <n> text <key>
ipsec ike remote address <n> <ip>
ipsec ike encryption <n> aes-cbc
ipsec ike group <n> modp1024
ipsec ike keepalive use <n> on dpd
```

## Example Usage
```hcl
resource "rtx_ipsec_tunnel" "site_to_site" {
  tunnel_id = 1

  local_address  = "203.0.113.1"
  remote_address = "198.51.100.1"

  pre_shared_key = var.ipsec_psk

  phase1 {
    encryption = "aes256-cbc"
    hash       = "sha256"
    dh_group   = 14
    lifetime   = 28800
  }

  phase2 {
    protocol   = "esp"
    encryption = "aes256-cbc"
    auth       = "sha256-hmac"
    pfs_group  = 14
    lifetime   = 3600
  }

  local_network  = "192.168.1.0/24"
  remote_network = "192.168.2.0/24"

  dpd_enabled  = true
  dpd_interval = 30
}
```
