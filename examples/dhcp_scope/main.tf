# Example: Basic DHCP Scope Configuration
# This example demonstrates how to create and manage DHCP scopes on RTX routers

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

# Basic DHCP Scope
# Creates a simple DHCP scope with default settings
resource "rtx_dhcp_scope" "basic" {
  scope_id = 1
  network  = "192.168.1.0/24"
}

# Full-featured DHCP Scope
# Creates a DHCP scope with all available options configured
resource "rtx_dhcp_scope" "full" {
  scope_id = 2
  network  = "192.168.2.0/24"

  # Default gateway for DHCP clients
  gateway = "192.168.2.1"

  # DNS servers (maximum 3)
  dns_servers = [
    "8.8.8.8",
    "8.8.4.4",
    "1.1.1.1"
  ]

  # Lease duration (Go duration format or "infinite")
  lease_time = "24h"

  # Exclude ranges - addresses that won't be assigned by DHCP
  exclude_ranges {
    start = "192.168.2.1"
    end   = "192.168.2.10"
  }

  exclude_ranges {
    start = "192.168.2.250"
    end   = "192.168.2.254"
  }
}

# DHCP Scope with Binding
# Demonstrates dependency between scope and binding resources
resource "rtx_dhcp_scope" "with_binding" {
  scope_id = 3
  network  = "192.168.3.0/24"
  gateway  = "192.168.3.1"

  dns_servers = ["8.8.8.8"]
  lease_time  = "72h"
}

# Static DHCP binding within the scope
# Note: The binding automatically depends on the scope through scope_id
resource "rtx_dhcp_binding" "server" {
  scope_id    = rtx_dhcp_scope.with_binding.scope_id
  ip_address  = "192.168.3.100"
  mac_address = "00:11:22:33:44:55"

  # Ensure scope exists before creating binding
  depends_on = [rtx_dhcp_scope.with_binding]
}

# Outputs
output "basic_scope_id" {
  description = "ID of the basic DHCP scope"
  value       = rtx_dhcp_scope.basic.scope_id
}

output "full_scope_network" {
  description = "Network of the full-featured DHCP scope"
  value       = rtx_dhcp_scope.full.network
}

output "binding_ip" {
  description = "IP address of the static binding"
  value       = rtx_dhcp_binding.server.ip_address
}
