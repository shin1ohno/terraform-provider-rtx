# RTX OSPF Configuration Examples
#
# Note: OSPF is a singleton resource - only one OSPF configuration per router.

terraform {
  required_version = ">= 1.11"
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.9"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
}

# OSPF configuration with multiple areas and features
resource "rtx_ospf" "main" {
  process_id = 1
  router_id  = "10.0.0.1"
  distance   = 110

  # Backbone area (Area 0)
  area {
    area_id = "0.0.0.0"
    type    = "normal"
  }

  # Stub area (Area 1)
  area {
    area_id    = "0.0.0.1"
    type       = "stub"
    no_summary = false
  }

  # NSSA area (Area 2)
  area {
    area_id    = "0.0.0.2"
    type       = "nssa"
    no_summary = true
  }

  # Networks in backbone area
  network {
    ip       = "10.0.0.0"
    wildcard = "0.0.0.255"
    area     = "0.0.0.0"
  }

  network {
    ip       = "192.168.0.0"
    wildcard = "0.0.0.255"
    area     = "0.0.0.0"
  }

  # Network in stub area
  network {
    ip       = "192.168.1.0"
    wildcard = "0.0.0.255"
    area     = "0.0.0.1"
  }

  # Network in NSSA area
  network {
    ip       = "192.168.2.0"
    wildcard = "0.0.0.255"
    area     = "0.0.0.2"
  }

  # Static neighbors for NBMA networks
  neighbor {
    ip       = "10.0.0.2"
    priority = 1
  }

  neighbor {
    ip       = "10.0.0.3"
    priority = 0
    cost     = 10
  }

  # Redistribute routes into OSPF
  redistribute_static    = true
  redistribute_connected = true

  # Originate default route
  default_information_originate = true
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
