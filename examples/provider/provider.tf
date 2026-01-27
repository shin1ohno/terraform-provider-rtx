terraform {
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.6"
    }
  }
}

# Configure the RTX Provider
provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
  port     = var.rtx_port
  timeout  = var.rtx_timeout
}

# RTX router system information
data "rtx_system_info" "router" {}

output "router_info" {
  value = {
    model            = data.rtx_system_info.router.model
    firmware_version = data.rtx_system_info.router.firmware_version
    uptime           = data.rtx_system_info.router.uptime
  }
}
