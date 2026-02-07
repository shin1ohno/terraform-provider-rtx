# RTX BGP Configuration Examples
#
# Note: BGP is a singleton resource - only one BGP configuration can exist per router.

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

# BGP configuration with iBGP and eBGP neighbors
resource "rtx_bgp" "main" {
  asn       = "65001"
  router_id = "10.0.0.1"

  # Enable logging neighbor state changes
  log_neighbor_changes = true

  # iBGP neighbor (same AS)
  neighbor {
    index     = 1
    ip        = "10.0.0.2"
    remote_as = "65001"
    keepalive = 30
    hold_time = 90
  }

  # Another iBGP neighbor
  neighbor {
    index     = 2
    ip        = "10.0.0.3"
    remote_as = "65001"
    keepalive = 30
    hold_time = 90
  }

  # eBGP neighbor (different AS)
  neighbor {
    index     = 3
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

  network {
    prefix = "192.168.1.0"
    mask   = "255.255.255.0"
  }

  # Redistribute routes
  redistribute_static    = true
  redistribute_connected = true
}

variable "rtx_host" {
  description = "RTX router hostname or IP address"
  type        = string
}

variable "rtx_username" {
  description = "RTX router username"
  type        = string
}

variable "rtx_password" {
  description = "RTX router password"
  type        = string
  sensitive   = true
}
