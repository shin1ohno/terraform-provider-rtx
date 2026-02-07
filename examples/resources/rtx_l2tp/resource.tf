# L2TPv2 LNS for remote access VPN
resource "rtx_l2tp" "remote_access" {
  tunnel_id = 1
  version   = "l2tp"
  mode      = "lns"
  name      = "Remote-Users"

  authentication {
    method   = "chap"
    username = "vpn-user"
    password = "example!PASS123"
  }

  ip_pool {
    start = "192.168.100.10"
    end   = "192.168.100.50"
  }

  enabled = true
}

# L2TPv3 site-to-site L2VPN
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
    tunnel_auth_enabled  = true
    tunnel_auth_password = "example!PASS123"
  }

  keepalive_enabled  = true
  keepalive_interval = 30
  keepalive_retry    = 5
  always_on          = true
  enabled            = true
}
