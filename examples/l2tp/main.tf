# RTX L2TP Configuration Examples
#
# DEPRECATED: This resource is deprecated. Use rtx_tunnel instead.
# The rtx_tunnel resource provides a unified interface for all tunnel types
# (IPsec, L2TPv3, L2TPv2) and better reflects the RTX command structure.
# See examples/tunnel/ for the new resource examples.

terraform {
  required_version = ">= 1.11"
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.9"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
}

# L2TPv2 LNS (L2TP Network Server) configuration
# Used for remote access VPN
resource "rtx_l2tp" "remote_access" {
  tunnel_id = 1
  version   = "l2tp"
  mode      = "lns"
  name      = "Remote-Users"

  authentication {
    method   = "chap"
    username = var.l2tp_username
    password = var.l2tp_password
  }

  ip_pool {
    start = "192.168.100.10"
    end   = "192.168.100.50"
  }

  # Optional: use IPsec for encryption
  ipsec_profile {
    enabled = true
  }

  enabled = true
}

# L2TPv3 site-to-site configuration
# Used for layer 2 VPN between sites
resource "rtx_l2tp" "site_to_site" {
  tunnel_id          = 2
  version            = "l2tpv3"
  mode               = "l2vpn"
  name               = "Branch-L2VPN"
  tunnel_source      = "203.0.113.1"
  tunnel_destination = "198.51.100.1"

  l2tpv3_config {
    local_router_id      = "1.1.1.1"
    remote_router_id     = "2.2.2.2"
    remote_end_id        = "branch-office"
    session_id           = 100
    tunnel_auth_enabled  = true
    tunnel_auth_password = var.tunnel_password
    bridge_interface     = "bridge1"
  }

  keepalive_enabled  = true
  keepalive_interval = 30
  keepalive_retry    = 5
  always_on          = true
  enabled            = true
}

# L2TPv3 with cookie for security
resource "rtx_l2tp" "with_cookie" {
  tunnel_id          = 3
  version            = "l2tpv3"
  mode               = "l2vpn"
  name               = "Datacenter-PW"
  tunnel_source      = "203.0.113.1"
  tunnel_destination = "10.0.0.2"

  l2tpv3_config {
    local_router_id  = "3.3.3.3"
    remote_router_id = "4.4.4.4"
    remote_end_id    = "dc-peer"
    session_id       = 1001
    cookie_size      = 8
  }

  enabled = true
}

# L2TP tunnel disabled (standby configuration)
resource "rtx_l2tp" "standby" {
  tunnel_id = 4
  version   = "l2tp"
  mode      = "lns"
  name      = "Backup-VPN"
  shutdown  = true

  authentication {
    method   = "mschap-v2"
    username = var.l2tp_username
    password = var.l2tp_password
  }

  ip_pool {
    start = "192.168.200.10"
    end   = "192.168.200.50"
  }

  enabled = false
}
