# RTX Static Route Configuration Examples

terraform {
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.5"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
}

# Default route through a gateway
resource "rtx_static_route" "default" {
  prefix = "0.0.0.0"
  mask   = "0.0.0.0"

  next_hop {
    gateway  = "192.168.0.1"
    distance = 1
  }
}

# Network route to 10.0.0.0/8 through a specific gateway
resource "rtx_static_route" "private_network" {
  prefix = "10.0.0.0"
  mask   = "255.0.0.0"

  next_hop {
    gateway  = "192.168.1.1"
    distance = 1
  }
}

# Route through a PPPoE interface
resource "rtx_static_route" "via_pppoe" {
  prefix = "172.16.0.0"
  mask   = "255.240.0.0"

  next_hop {
    interface = "pp 1"
    distance  = 1
  }
}

# Route through a tunnel interface with keepalive (permanent)
resource "rtx_static_route" "vpn_tunnel" {
  prefix = "192.168.100.0"
  mask   = "255.255.255.0"

  next_hop {
    interface = "tunnel 1"
    distance  = 1
    permanent = true
  }
}

# Load balancing with multiple next hops (ECMP)
resource "rtx_static_route" "load_balanced" {
  prefix = "10.10.0.0"
  mask   = "255.255.0.0"

  next_hop {
    gateway  = "192.168.1.1"
    distance = 1
  }

  next_hop {
    gateway  = "192.168.2.1"
    distance = 1
  }
}

# Failover route with different distances (weights)
resource "rtx_static_route" "failover" {
  prefix = "10.20.0.0"
  mask   = "255.255.0.0"

  # Primary path
  next_hop {
    gateway  = "192.168.1.1"
    distance = 1
  }

  # Backup path (higher distance = lower priority)
  next_hop {
    gateway  = "192.168.2.1"
    distance = 10
  }
}

# Route with IP filter
resource "rtx_static_route" "filtered" {
  prefix = "10.30.0.0"
  mask   = "255.255.0.0"

  next_hop {
    gateway  = "192.168.1.1"
    distance = 1
    filter   = 100
  }
}
