# Default route through a gateway
resource "rtx_static_route" "default" {
  prefix = "0.0.0.0"
  mask   = "0.0.0.0"

  next_hop {
    gateway  = "192.168.0.1"
    distance = 1
  }
}

# Network route to 10.0.0.0/8 via gateway
resource "rtx_static_route" "private" {
  prefix = "10.0.0.0"
  mask   = "255.0.0.0"

  next_hop {
    gateway  = "192.168.1.1"
    distance = 1
  }
}

# Route through PPPoE interface with failover
resource "rtx_static_route" "internet" {
  prefix = "0.0.0.0"
  mask   = "0.0.0.0"

  # Primary: PPPoE connection
  next_hop {
    interface = "pp 1"
    distance  = 1
    permanent = true
  }

  # Backup: Secondary gateway
  next_hop {
    gateway  = "192.168.0.254"
    distance = 10
  }
}

# VPN tunnel route
resource "rtx_static_route" "vpn" {
  prefix = "192.168.100.0"
  mask   = "255.255.255.0"

  next_hop {
    interface = "tunnel 1"
    distance  = 1
    permanent = true
  }
}
