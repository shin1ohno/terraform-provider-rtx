terraform {
  required_providers {
    rtx = {
      source  = "registry.terraform.io/sh1/rtx"
      version = "0.1.0"
    }
  }
}

provider "rtx" {
  host                 = var.rtx_host
  username             = var.rtx_username
  password             = var.rtx_password
  admin_password       = var.rtx_admin_password
  port                 = var.rtx_port
  timeout              = 30
  skip_host_key_check  = var.skip_host_key_check
}

# Data source to read existing static routes
data "rtx_static_routes" "all" {}

# Display current static routes
output "current_static_routes" {
  value = data.rtx_static_routes.all.routes
}

# Static route resources - uncomment after import
# Import existing routes with: terraform import rtx_static_route.route_name "destination/gateway"

# Actual static routes from router config to import
# Based on: ip route 10.33.128.0/21 gateway 192.168.1.20 gateway 192.168.1.21
# Based on: ip route 100.64.0.0/10 gateway 192.168.1.20 gateway 192.168.1.21

# Routes based on actual router configuration
# ip route default gateway dhcp lan2
resource "rtx_static_route" "default_route" {
  destination = "0.0.0.0/0"
  
  gateways {
    interface = "dhcp lan2"
  }
  
  description = "Default route via DHCP-obtained gateway on LAN2"
}

resource "rtx_static_route" "vpc" {
  destination = "10.33.128.0/21"
  
  gateways {
    ip = "192.168.1.20"
  }
  
  gateways {
    ip = "192.168.1.21"
  }
  
  description = "AWS VPC network with failover"
}

resource "rtx_static_route" "tailscale" {
  destination = "100.64.0.0/10"
  
  gateways {
    ip = "192.168.1.20"
  }
  
  gateways {
    ip = "192.168.1.21"
  }
  
  description = "Tailscale network route"
}
