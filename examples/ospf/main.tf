# RTX OSPF Configuration Examples

terraform {
  required_providers {
    rtx = {
      source  = "registry.terraform.io/sh1/rtx"
      version = "~> 0.2"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
}

# Basic OSPF configuration
resource "rtx_ospf" "basic" {
  process_id = 1
  router_id  = "10.0.0.1"

  # Define OSPF area
  area {
    area_id = "0.0.0.0"
    type    = "normal"
  }

  # Network to include in OSPF
  network {
    prefix  = "192.168.1.0/24"
    area_id = "0.0.0.0"
  }
}

# OSPF with multiple areas (including stub area)
resource "rtx_ospf" "multi_area" {
  process_id = 1
  router_id  = "10.0.0.1"
  distance   = 110

  # Backbone area
  area {
    area_id = "0.0.0.0"
    type    = "normal"
  }

  # Stub area
  area {
    area_id      = "0.0.0.1"
    type         = "stub"
    no_summary   = false
    default_cost = 10
  }

  # Network in backbone
  network {
    prefix  = "10.0.0.0/24"
    area_id = "0.0.0.0"
  }

  # Network in stub area
  network {
    prefix  = "192.168.1.0/24"
    area_id = "0.0.0.1"
  }
}

# OSPF with NSSA (Not So Stubby Area)
resource "rtx_ospf" "nssa" {
  process_id = 1
  router_id  = "10.0.0.1"

  area {
    area_id = "0.0.0.0"
    type    = "normal"
  }

  area {
    area_id    = "0.0.0.2"
    type       = "nssa"
    no_summary = true
  }

  network {
    prefix  = "10.0.0.0/24"
    area_id = "0.0.0.0"
  }
}

# OSPF with static neighbor (for NBMA networks)
resource "rtx_ospf" "nbma" {
  process_id = 1
  router_id  = "10.0.0.1"

  area {
    area_id = "0.0.0.0"
    type    = "normal"
  }

  neighbor {
    address  = "10.0.0.2"
    priority = 1
  }

  neighbor {
    address  = "10.0.0.3"
    priority = 0
  }

  network {
    prefix  = "10.0.0.0/24"
    area_id = "0.0.0.0"
  }
}
