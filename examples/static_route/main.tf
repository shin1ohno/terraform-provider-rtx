terraform {
  required_providers {
    rtx = {
      source  = "registry.terraform.io/sh1/rtx"
      version = "0.1.0"
    }
  }
}

provider "rtx" {
  host           = var.rtx_host
  username       = var.rtx_username
  password       = var.rtx_password
  admin_password = var.rtx_admin_password
  port           = var.rtx_port
  timeout        = 30
}

# Data source to read existing static routes
data "rtx_static_routes" "all" {}

# Display current static routes
output "current_static_routes" {
  value = data.rtx_static_routes.all.routes
}

# Test static route resource (commented out for safety)
# resource "rtx_static_route" "test_route" {
#   destination = "10.10.0.0/24"
#   gateway_ip  = "192.168.1.1"
#   metric      = 10
#   description = "Test route for Session 14"
# }