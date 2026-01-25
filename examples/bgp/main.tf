# RTX BGP Configuration Examples

terraform {
  required_providers {
    rtx = {
      source = "shin1ohno/rtx"
      version = "~> 0.5"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
}

# Basic BGP configuration with a single neighbor
resource "rtx_bgp" "basic" {
  asn       = 65001
  router_id = "10.0.0.1"

  neighbor {
    address    = "10.0.0.2"
    remote_as  = 65002
    keep_alive = 30
    hold_time  = 90
  }

  network {
    prefix = "192.168.0.0/24"
  }
}

# iBGP configuration with multiple neighbors
resource "rtx_bgp" "ibgp" {
  asn       = 65001
  router_id = "10.0.0.1"

  # iBGP neighbor (same AS)
  neighbor {
    address     = "10.0.0.2"
    remote_as   = 65001
    description = "iBGP peer"
  }

  # Another iBGP neighbor
  neighbor {
    address     = "10.0.0.3"
    remote_as   = 65001
    description = "iBGP peer 2"
  }
}

# eBGP configuration with route redistribution
resource "rtx_bgp" "ebgp" {
  asn       = 65001
  router_id = "203.0.113.1"

  neighbor {
    address    = "203.0.113.2"
    remote_as  = 65002
    keep_alive = 60
    hold_time  = 180
  }

  # Advertise networks
  network {
    prefix = "192.168.1.0/24"
  }

  network {
    prefix = "192.168.2.0/24"
  }

  # Redistribute from other protocols
  redistribution {
    protocol = "static"
  }

  redistribution {
    protocol = "connected"
  }
}
