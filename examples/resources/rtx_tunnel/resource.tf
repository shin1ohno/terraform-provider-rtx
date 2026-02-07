# Site-to-site IPsec VPN tunnel
resource "rtx_tunnel" "site_to_site_vpn" {
  tunnel_id     = 1
  encapsulation = "ipsec"
  enabled       = true

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

# L2TPv3 over IPsec tunnel (L2VPN)
resource "rtx_tunnel" "l2vpn" {
  tunnel_id     = 2
  encapsulation = "l2tpv3"
  enabled       = true

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
  }
}
