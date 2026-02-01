# Requirements: rtx_tunnel (Unified Tunnel Resource)

## Overview

Unify `rtx_ipsec_tunnel` and `rtx_l2tp` into a single `rtx_tunnel` resource that reflects RTX's actual command structure where `tunnel select N` is the parent container for both IPsec and L2TP settings.

## Background

### Current Design Problems

1. **Conceptual Mismatch**: `rtx_ipsec_tunnel` and `rtx_l2tp` are separate resources, but in RTX they share the same `tunnel select N` context
2. **Redundant References**: Both resources have `tunnel_id`, and `rtx_l2tp.ipsec_profile.tunnel_id` creates unnecessary coupling
3. **Confusing Naming**: `tunnel_id` vs `ipsec_tunnel_id` confusion

### RTX Command Structure

```
tunnel select 1                    <- Parent: tunnel interface
  tunnel encapsulation l2tpv3
  ipsec tunnel 101                 <- Child: IPsec binding
    ipsec sa policy 101 1 esp ...
    ipsec ike local address 1 ...
  l2tp hostname ...                <- Child: L2TP settings (sibling to IPsec)
  tunnel enable 1
```

### Industry Standard

- **Cisco VTI**: `interface Tunnel` is the parent, IPsec is an attribute
- **Juniper**: `st0` interface is the parent, IPsec is bound via `bind-interface`
- **Fortinet**: Phase1-interface creates tunnel interface automatically

## Functional Requirements

### FR-1: Unified Tunnel Resource

Create `rtx_tunnel` resource with:
- `tunnel_id` (Required): The tunnel interface ID (`tunnel select N`, 1-6000)
- `encapsulation` (Required): `ipsec`, `l2tpv3`, or `l2tp`
- `enabled` (Optional, Computed): Tunnel enable/disable (default: true)
- `name` (**Computed**, read-only): Tunnel description - RTX does not support setting descriptions within tunnel context
- `tunnel_interface` (Computed): The tunnel interface name (e.g., "tunnel1")
- `endpoint_name` (Optional): Tunnel endpoint name for DNS resolution
- `endpoint_name_type` (Optional): Endpoint name type (`fqdn`)
- `ipsec` block (Optional): IPsec configuration
- `l2tp` block (Optional): L2TP configuration (required when encapsulation includes L2TP)

### FR-2: IPsec Block

Nested `ipsec` block containing:
- `ipsec_tunnel_id` (Optional, Computed): The IPsec tunnel number (`ipsec tunnel N`). Defaults to `tunnel_id` if not specified.
- `local_address` (Optional, Computed): Local IKE endpoint
- `remote_address` (Optional, Computed): Remote IKE endpoint
- `pre_shared_key` (Required, Sensitive, **WriteOnly**): IKE PSK - not stored in state for security
- `nat_traversal` (Optional, Computed): Enable NAT traversal (default: false)
- `ike_remote_name` (Optional, Computed): IKE remote name value
- `ike_remote_name_type` (Optional, Computed): IKE remote name type (ipv4-addr, fqdn, user-fqdn, ipv6-addr, key-id, l2tpv3)
- `ike_keepalive_log` (Optional, Computed): Enable IKE keepalive logging (default: false)
- `ike_log` (Optional): IKE log options (e.g., "key-info message-info payload-info")
- `ipsec_transform` block: ESP/AH algorithms
- `keepalive` block: DPD/heartbeat settings
- `secure_filter_in` / `secure_filter_out` (Optional): Filter IDs
- `tcp_mss_limit` (Optional): TCP MSS limit setting

### FR-3: L2TP Block

Nested `l2tp` block containing:
- Common settings:
  - `hostname` (Optional): L2TP hostname for negotiation
  - `always_on` (Optional, Computed): Always-on mode (default: false)
  - `disconnect_time` (Optional, Computed): Disconnect time in seconds (0 = off)
  - `keepalive_log` (Optional, Computed): Enable L2TP keepalive logging (default: false)
  - `syslog` (Optional, Computed): Enable L2TP syslog (default: false)
  - `tunnel_auth` block: Tunnel authentication
  - `keepalive` block: L2TP keepalive settings
- For L2TPv3:
  - `local_router_id` / `remote_router_id` (Optional): Router IDs
  - `remote_end_id` (Optional): End ID
- For L2TPv2 (remote access):
  - `authentication` block: PAP/CHAP settings
  - `ip_pool` block: Client IP pool

### FR-4: Encapsulation Modes

| Encapsulation | IPsec Block | L2TP Block | Use Case |
|---------------|-------------|------------|----------|
| `ipsec` | Required | Not allowed | Site-to-site IPsec VPN |
| `l2tpv3` | Optional | Required | L2VPN (with optional IPsec) |
| `l2tp` | Required | Required | L2TPv2 remote access (always over IPsec) |

### FR-5: Backward Compatibility

- Deprecate `rtx_ipsec_tunnel` and `rtx_l2tp` resources
- Provide migration guide in documentation
- No state migration tools required (user handles migration)

## Non-Functional Requirements

### NFR-1: Documentation

- Update all docs to use `rtx_tunnel`
- Document migration from old resources
- Explain RTX command structure mapping

### NFR-2: Examples

- Update `examples/import/main.tf` to use new resource
- Provide examples for each encapsulation mode

### NFR-3: Testing

- Unit tests for parser
- Integration tests for CRUD operations
- Acceptance tests with real router

## Terraform Schema Example

> **Note:** The `name` attribute is **Computed** (read-only). RTX does not support setting tunnel descriptions within the tunnel context. Use `rtx_interface` to set the tunnel interface description if needed.

```hcl
# Pure IPsec site-to-site VPN
resource "rtx_tunnel" "site_to_site" {
  tunnel_id     = 1
  encapsulation = "ipsec"
  enabled       = true
  # 'name' is Computed (read-only) - do not set

  ipsec {
    # ipsec_tunnel_id is Computed (defaults to tunnel_id)
    local_address   = "203.0.113.1"
    remote_address  = "198.51.100.1"
    pre_shared_key  = var.ipsec_psk

    ipsec_transform {
      protocol          = "esp"
      encryption_aes128 = true
      integrity_sha1    = true
    }

    keepalive {
      enabled  = true
      mode     = "dpd"
      interval = 30
      retry    = 3
    }
  }
}

# L2TPv3 over IPsec (site-to-site L2VPN)
resource "rtx_tunnel" "l2vpn" {
  tunnel_id     = 2
  encapsulation = "l2tpv3"
  enabled       = true

  # Optional: DNS-based endpoint resolution
  endpoint_name      = "branch.example.com"
  endpoint_name_type = "fqdn"

  ipsec {
    # ipsec_tunnel_id is Computed (defaults to tunnel_id)
    local_address   = "192.168.1.253"
    remote_address  = "itm.ohno.be"
    pre_shared_key  = var.ipsec_psk
    nat_traversal   = true  # Enable NAT traversal

    ipsec_transform {
      protocol          = "esp"
      encryption_aes128 = true
      integrity_sha1    = true
    }

    keepalive {
      enabled  = true
      mode     = "heartbeat"
      interval = 10
      retry    = 6
    }

    secure_filter_in = [200028, 200099]
    tcp_mss_limit    = "auto"
  }

  l2tp {
    hostname         = "ebisu-RTX1210"
    local_router_id  = "192.168.1.253"
    remote_router_id = "192.168.1.254"
    remote_end_id    = "shin1"
    always_on        = true
    syslog           = true  # Enable L2TP syslog

    tunnel_auth {
      enabled  = true
      password = var.l2tp_password
    }

    keepalive {
      enabled  = true
      interval = 60
      retry    = 3
    }
  }
}

# L2TPv2 remote access (always over IPsec)
resource "rtx_tunnel" "remote_access" {
  tunnel_id     = 3
  encapsulation = "l2tp"
  enabled       = true

  ipsec {
    # ipsec_tunnel_id is Computed (defaults to tunnel_id)
    pre_shared_key  = var.ipsec_psk

    ipsec_transform {
      protocol          = "esp"
      encryption_aes128 = true
      integrity_sha1    = true
    }
  }

  l2tp {
    authentication {
      method   = "chap"
      username = "vpnuser"
      password = var.l2tp_password
    }

    ip_pool {
      start = "192.168.100.100"
      end   = "192.168.100.200"
    }
  }
}

# Reference computed attributes
output "tunnel_interface" {
  description = "The tunnel interface name (e.g., 'tunnel1')"
  value       = rtx_tunnel.site_to_site.tunnel_interface
}
```
