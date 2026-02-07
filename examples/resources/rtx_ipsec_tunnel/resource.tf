# Basic site-to-site IPsec VPN tunnel
resource "rtx_ipsec_tunnel" "site_to_site" {
  tunnel_id      = 1
  name           = "Office-to-Datacenter"
  local_address  = "203.0.113.1"
  remote_address = "198.51.100.1"
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
    lifetime_seconds   = 3600
  }
}

# IPsec tunnel with Dead Peer Detection
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

  dpd_enabled  = true
  dpd_interval = 30
  dpd_retry    = 5
}
