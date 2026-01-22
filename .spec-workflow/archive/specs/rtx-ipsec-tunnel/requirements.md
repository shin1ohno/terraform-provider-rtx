# Requirements: rtx_ipsec_tunnel

## Overview
Terraform resource for managing IPsec VPN tunnels on Yamaha RTX routers.

**Cisco Equivalent**: `iosxe_crypto_ikev2_proposal`, `iosxe_crypto_ipsec_profile`, `iosxe_interface_tunnel`

## Cisco Compatibility

This resource follows Cisco IOS XE Terraform provider naming conventions:

| RTX Attribute | Cisco Equivalent | Notes |
|---------------|------------------|-------|
| `name` | `name` | Tunnel/proposal name |
| `encryption_aes_cbc_256` | `encryption_aes_cbc_256` | AES-256 encryption |
| `integrity_sha256` | `integrity_sha256` | SHA-256 integrity |
| `group_fourteen` | `group_fourteen` | DH Group 14 (2048-bit) |
| `group_sixteen` | `group_sixteen` | DH Group 16 (4096-bit) |
| `lifetime` | `lifetime_seconds` | SA lifetime |

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure IPsec tunnel with IKE settings
- **Read**: Query tunnel and SA status
  - Tunnel/SA status is operational-only and MUST NOT be persisted in Terraform state
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

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned IPsec tunnel changes |
| `terraform apply` | ✅ Required | Create, update, or delete IPsec tunnels |
| `terraform destroy` | ✅ Required | Remove tunnel and clear SAs |
| `terraform import` | ✅ Required | Import existing tunnels into Terraform state |
| `terraform refresh` | ✅ Required | Sync state with actual tunnel configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `<tunnel_id>` (e.g., `1`)
- **Import Command**: `terraform import rtx_ipsec_tunnel.site_to_site 1`
- **Post-Import**: All Phase1/Phase2 settings must be populated (PSK marked sensitive)

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
# IPsec Tunnel - integrated configuration for RTX
resource "rtx_ipsec_tunnel" "site_to_site" {
  id   = 1
  name = "site-to-site-vpn"

  local_address  = "203.0.113.1"
  remote_address = "198.51.100.1"

  pre_shared_key = var.ipsec_psk  # Sensitive

  # IKE Phase 1 (IKEv2)
  ikev2_proposal {
    encryption_aes_cbc_256 = true
    integrity_sha256       = true
    group_fourteen         = true  # DH Group 14
    lifetime_seconds       = 28800
  }

  # IPsec Phase 2
  ipsec_transform {
    protocol               = "esp"
    encryption_aes_cbc_256 = true
    integrity_sha256_hmac  = true
    pfs_group_fourteen     = true
    lifetime_seconds       = 3600
  }

  local_network  = "192.168.1.0/24"
  remote_network = "192.168.2.0/24"

  dpd_enabled  = true
  dpd_interval = 30
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state.
- Operational/runtime status must not be stored in state.
