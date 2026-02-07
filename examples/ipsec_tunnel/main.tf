# RTX IPsec Tunnel Configuration Examples
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
      version = "~> 0.10"
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
  name           = "Office-to-Datacenter"
  local_address  = "203.0.113.1"
  remote_address = "198.51.100.1"
  pre_shared_key = var.psk

  # IKEv2 proposal (Phase 1)
  ikev2_proposal {
    encryption_aes256 = true
    integrity_sha256  = true
    group_fourteen    = true
    lifetime_seconds  = 86400
  }

  # IPsec transform (Phase 2)
  ipsec_transform {
    protocol           = "esp"
    encryption_aes256  = true
    integrity_sha256   = true
    pfs_group_fourteen = true
    lifetime_seconds   = 3600
  }
}

# IPsec tunnel with Dead Peer Detection (DPD)
resource "rtx_ipsec_tunnel" "with_dpd" {
  tunnel_id      = 2
  name           = "Branch-Office"
  local_address  = "203.0.113.1"
  remote_address = "198.51.100.2"
  pre_shared_key = var.psk

  ikev2_proposal {
    encryption_aes256 = true
    integrity_sha256  = true
    group_fourteen    = true
    lifetime_seconds  = 86400
  }

  ipsec_transform {
    protocol           = "esp"
    encryption_aes256  = true
    integrity_sha256   = true
    pfs_group_fourteen = true
    lifetime_seconds   = 7200
  }

  # Enable DPD for detecting dead peers
  dpd_enabled  = true
  dpd_interval = 30
  dpd_retry    = 5
}

# IPsec tunnel with security filters
resource "rtx_ipsec_tunnel" "with_filters" {
  tunnel_id      = 3
  name           = "Datacenter"
  local_address  = "10.0.0.1"
  remote_address = "10.0.0.2"
  pre_shared_key = var.psk

  ikev2_proposal {
    encryption_aes128 = true
    integrity_sha1    = true
    group_fourteen    = true
  }

  ipsec_transform {
    protocol          = "esp"
    encryption_aes128 = true
    integrity_sha1    = true
  }

  # Security filters and TCP MSS limit
  secure_filter_in  = [100, 101, 102]
  secure_filter_out = [200, 201]
  tcp_mss_limit     = "auto"
}

# IPsec tunnel disabled (pre-configured but not active)
resource "rtx_ipsec_tunnel" "standby" {
  tunnel_id      = 4
  name           = "Standby-Site"
  local_address  = "203.0.113.1"
  remote_address = "192.0.2.1"
  pre_shared_key = var.psk
  enabled        = false

  ikev2_proposal {
    encryption_aes128 = true
    integrity_sha256  = true
    group_fourteen    = true
    lifetime_seconds  = 86400
  }

  ipsec_transform {
    protocol           = "esp"
    encryption_aes128  = true
    integrity_sha256   = true
    pfs_group_fourteen = true
    lifetime_seconds   = 3600
  }
}
