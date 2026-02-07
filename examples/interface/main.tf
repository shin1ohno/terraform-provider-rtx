# Example: RTX Interface Configuration
#
# This example demonstrates various interface configuration scenarios
# for Yamaha RTX routers using the terraform-provider-rtx.

terraform {
  required_version = ">= 1.11"
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.10"
    }
  }
}

# Provider configuration (use environment variables for credentials)
provider "rtx" {
  host                = var.rtx_host
  username            = var.rtx_username
  password            = var.rtx_password
  admin_password      = var.rtx_admin_password
  skip_host_key_check = true
}

# Example 1: WAN interface with DHCP (typical ISP connection)
resource "rtx_interface" "wan" {
  name        = "lan2"
  description = "WAN connection to ISP"

  ip_address {
    dhcp = true
  }

  # NAT descriptor binding (reference rtx_nat_masquerade or rtx_nat_static resources)
  nat_descriptor = 1000
}

# Example 2: LAN interface with static IP
resource "rtx_interface" "lan" {
  name        = "lan1"
  description = "Internal LAN"

  ip_address {
    address = "192.168.1.1/24"
  }

  # Enable ProxyARP for this interface
  proxyarp = true
}

# Example 3: Bridge interface for internal network
resource "rtx_interface" "bridge" {
  name        = "bridge1"
  description = "Internal bridge network"

  ip_address {
    address = "192.168.100.1/24"
  }

  # Custom MTU
  mtu = 1500
}

# Example 4: Minimal interface configuration
resource "rtx_interface" "minimal" {
  name = "lan3"

  ip_address {
    address = "10.0.0.1/24"
  }
}

# Variables
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

variable "rtx_admin_password" {
  description = "RTX router administrator password"
  type        = string
  sensitive   = true
}

# Outputs
output "wan_interface" {
  description = "WAN interface configuration"
  value = {
    name           = rtx_interface.wan.name
    description    = rtx_interface.wan.description
    nat_descriptor = rtx_interface.wan.nat_descriptor
  }
}

output "lan_interface" {
  description = "LAN interface configuration"
  value = {
    name       = rtx_interface.lan.name
    ip_address = rtx_interface.lan.ip_address
    proxyarp   = rtx_interface.lan.proxyarp
  }
}
