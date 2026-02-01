# RTX Unified Tunnel Configuration Examples
#
# This resource provides a unified interface for managing tunnel configurations
# on RTX routers. It supports three encapsulation types:
# - ipsec: Site-to-site VPN tunnels
# - l2tpv3: L2VPN tunnels (L2TPv3 over IPsec)
# - l2tp: L2TPv2 remote access VPN

terraform {
  required_version = ">= 1.11"
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.8"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
}

# ============================================================================
# Example 1: Site-to-Site IPsec VPN Tunnel
# ============================================================================
# This creates a basic IPsec tunnel for connecting two sites.
resource "rtx_tunnel" "site_to_site_vpn" {
  tunnel_id     = 1
  encapsulation = "ipsec"
  enabled       = true
  # Note: 'name' is Computed (read-only) - use rtx_interface to set tunnel interface description

  ipsec {
    local_address  = "192.168.1.1"
    remote_address = "203.0.113.100"
    pre_shared_key = var.ipsec_psk

    ipsec_transform {
      protocol          = "esp"
      encryption_aes256 = true
      integrity_sha256  = true
    }

    keepalive {
      enabled  = true
      mode     = "dpd"
      interval = 30
      retry    = 3
    }
  }
}

# ============================================================================
# Example 2: L2TPv3 over IPsec Tunnel (L2VPN)
# ============================================================================
# This creates an L2VPN tunnel using L2TPv3 over IPsec.
# Useful for extending Layer 2 networks across sites.
resource "rtx_tunnel" "l2vpn" {
  tunnel_id     = 2
  encapsulation = "l2tpv3"
  enabled       = true

  # Optional: DNS-based endpoint resolution
  endpoint_name      = "branch.example.com"
  endpoint_name_type = "fqdn"

  ipsec {
    ipsec_tunnel_id = 101
    local_address   = "192.168.1.253"
    remote_address  = "branch.example.com"
    pre_shared_key  = var.ipsec_psk
    nat_traversal   = true

    ipsec_transform {
      protocol          = "esp"
      encryption_aes128 = true
      integrity_sha1    = true
    }
  }

  l2tp {
    hostname         = "main-router"
    local_router_id  = "192.168.1.253"
    remote_router_id = "192.168.2.254"
    remote_end_id    = "branch-router"
    always_on        = true

    tunnel_auth {
      enabled  = true
      password = var.l2tp_password
    }

    keepalive {
      enabled  = true
      interval = 60
      retry    = 5
    }
  }
}

# ============================================================================
# Example 3: L2TPv2 Remote Access VPN (LNS Mode)
# ============================================================================
# This configures the router as an L2TP Network Server (LNS)
# for remote access VPN connections.
# Note: L2TPv2 requires IPsec for transport encryption.
resource "rtx_tunnel" "remote_access" {
  tunnel_id     = 3
  encapsulation = "l2tp"
  enabled       = true

  ipsec {
    pre_shared_key = var.ipsec_psk

    ipsec_transform {
      protocol          = "esp"
      encryption_aes256 = true
      integrity_sha256  = true
    }
  }

  l2tp {
    hostname  = "vpn-server"
    always_on = false

    tunnel_auth {
      enabled  = true
      password = var.l2tp_password
    }

    keepalive {
      enabled  = true
      interval = 30
      retry    = 3
    }
  }
}

# ============================================================================
# Example 4: IPsec Tunnel with Security Filters
# ============================================================================
# This creates an IPsec tunnel with inbound/outbound security filters
# and TCP MSS limit configuration.
resource "rtx_tunnel" "filtered_vpn" {
  tunnel_id     = 4
  encapsulation = "ipsec"
  enabled       = true

  ipsec {
    local_address     = "10.0.0.1"
    remote_address    = "10.0.0.2"
    pre_shared_key    = var.ipsec_psk
    secure_filter_in  = [100, 101, 102]
    secure_filter_out = [200, 201]
    tcp_mss_limit     = "auto"

    ipsec_transform {
      protocol          = "esp"
      encryption_aes256 = true
      integrity_sha256  = true
    }
  }
}
