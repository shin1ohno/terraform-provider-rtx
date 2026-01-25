# RTX IPsec Tunnel Configuration Examples

terraform {
  required_providers {
    rtx = {
      source  = "registry.terraform.io/sh1/rtx"
      version = "~> 0.2"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
}

# Basic Site-to-Site IPsec VPN tunnel
resource "rtx_ipsec_tunnel" "site_to_site" {
  tunnel_id      = 1
  local_id       = "203.0.113.1"
  remote_id      = "203.0.113.2"
  remote_name    = "remote-office"
  pre_shared_key = var.psk

  # IKEv2 proposal
  ikev2_proposal {
    encryption = "aes256-cbc"
    hash       = "sha256"
    dh_group   = 14
    lifetime   = 86400
  }

  # IPsec transform
  ipsec_transform {
    encryption       = "aes256-cbc"
    authentication   = "sha256-hmac"
    pfs_group        = 14
    lifetime_seconds = 3600
  }
}

# IPsec tunnel with Dead Peer Detection (DPD)
resource "rtx_ipsec_tunnel" "with_dpd" {
  tunnel_id      = 2
  local_id       = "203.0.113.1"
  remote_id      = "198.51.100.1"
  remote_name    = "branch-office"
  pre_shared_key = var.psk

  ikev2_proposal {
    encryption = "aes256-cbc"
    hash       = "sha512"
    dh_group   = 15
    lifetime   = 86400
  }

  ipsec_transform {
    encryption       = "aes256-gcm"
    authentication   = "sha512-hmac"
    pfs_group        = 15
    lifetime_seconds = 7200
  }

  # Enable DPD for detecting dead peers
  dpd {
    enabled  = true
    interval = 30
    retry    = 5
    action   = "restart"
  }
}

# IPsec tunnel with AES-GCM (AEAD cipher)
resource "rtx_ipsec_tunnel" "aead" {
  tunnel_id      = 3
  local_id       = "10.0.0.1"
  remote_id      = "10.0.0.2"
  remote_name    = "datacenter"
  pre_shared_key = var.psk

  ikev2_proposal {
    encryption = "aes256-gcm"
    hash       = "sha384"
    dh_group   = 19
    lifetime   = 86400
  }

  ipsec_transform {
    encryption       = "aes256-gcm"
    authentication   = "sha384-hmac"
    pfs_group        = 19
    lifetime_seconds = 3600
    lifetime_bytes   = 4294967296 # 4GB
  }
}

# IPsec tunnel disabled (pre-configured but not active)
resource "rtx_ipsec_tunnel" "standby" {
  tunnel_id      = 4
  local_id       = "203.0.113.1"
  remote_id      = "192.0.2.1"
  remote_name    = "standby-site"
  pre_shared_key = var.psk
  shutdown       = true

  ikev2_proposal {
    encryption = "aes128-cbc"
    hash       = "sha256"
    dh_group   = 14
    lifetime   = 86400
  }

  ipsec_transform {
    encryption       = "aes128-cbc"
    authentication   = "sha256-hmac"
    pfs_group        = 14
    lifetime_seconds = 3600
  }
}
