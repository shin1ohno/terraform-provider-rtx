# RTX L2TP Configuration Examples

terraform {
  required_version = ">= 1.11"
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.7"
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
  tunnel_id   = 1
  mode        = "lns"
  tunnel_name = "remote-users"
  hostname    = "vpn.example.com"

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
}

# L2TPv3 site-to-site configuration
# Used for layer 2 VPN between sites
resource "rtx_l2tp" "site_to_site" {
  tunnel_id   = 2
  mode        = "l2vpn"
  tunnel_name = "branch-l2vpn"
  hostname    = "main-office"

  l2tpv3_config {
    remote_end_id     = "branch-office"
    local_session_id  = 100
    remote_session_id = 200
    tunnel_password   = var.tunnel_password
    bridge_interface  = "bridge1"
  }
}

# L2TPv3 with pseudowire
resource "rtx_l2tp" "pseudowire" {
  tunnel_id   = 3
  mode        = "l2vpn"
  tunnel_name = "datacenter-pw"
  hostname    = "dc-router"

  l2tpv3_config {
    remote_end_id     = "dc-peer"
    local_session_id  = 1001
    remote_session_id = 1001
    tunnel_password   = var.tunnel_password
    pseudowire_type   = "ethernet"
    vlan_id           = 100
  }
}

# L2TP tunnel disabled (standby configuration)
resource "rtx_l2tp" "standby" {
  tunnel_id   = 4
  mode        = "lns"
  tunnel_name = "backup-vpn"
  hostname    = "backup.example.com"
  shutdown    = true

  authentication {
    method   = "mschap-v2"
    username = var.l2tp_username
    password = var.l2tp_password
  }

  ip_pool {
    start = "192.168.200.10"
    end   = "192.168.200.50"
  }
}
