# BGP configuration with iBGP and eBGP neighbors
resource "rtx_bgp" "main" {
  asn       = "65001"
  router_id = "10.0.0.1"

  log_neighbor_changes = true

  # iBGP neighbor (same AS)
  neighbor {
    index     = 1
    ip        = "10.0.0.2"
    remote_as = "65001"
    keepalive = 30
    hold_time = 90
  }

  # eBGP neighbor (different AS)
  neighbor {
    index     = 2
    ip        = "203.0.113.2"
    remote_as = "65002"
    keepalive = 60
    hold_time = 180
  }

  # Advertise networks
  network {
    prefix = "192.168.0.0"
    mask   = "255.255.255.0"
  }

  redistribute_static    = true
  redistribute_connected = true
}
