# Example: Basic DNS Server Configuration
# This example shows how to configure DNS server settings on an RTX router.

terraform {
  required_providers {
    rtx = {
      source  = "github.com/sh1/rtx"
      version = "~> 0.1"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
}

# Basic DNS Server Configuration
# Uses Google and Cloudflare public DNS servers
resource "rtx_dns_server" "basic" {
  name_servers = ["8.8.8.8", "8.8.4.4"]
  service_on   = true
}

# Full DNS Server Configuration Example
# Demonstrates all available options
resource "rtx_dns_server" "full" {
  # Enable DNS domain lookup (default: true)
  domain_lookup = true

  # Set default domain name for queries
  domain_name = "example.com"

  # Configure up to 3 DNS servers
  name_servers = ["8.8.8.8", "1.1.1.1"]

  # Domain-based DNS server selection
  # Use different DNS servers for specific domains/patterns
  server_select {
    priority      = 1
    servers       = ["192.168.1.1"]
    query_pattern = "internal.example.com"
  }

  server_select {
    priority      = 2
    servers       = ["10.0.0.1", "10.0.0.2"]
    query_pattern = "*.local"
  }

  # Advanced server select with EDNS and record type filtering
  server_select {
    priority      = 10
    servers       = ["10.0.0.53"]
    edns          = true
    record_type   = "any" # a, aaaa, ptr, mx, ns, cname, any
    query_pattern = "."   # Match all queries
  }

  # Static DNS host entries (local DNS overrides)
  hosts {
    name    = "router"
    address = "192.168.1.1"
  }

  hosts {
    name    = "nas"
    address = "192.168.1.10"
  }

  hosts {
    name    = "printer"
    address = "192.168.1.20"
  }

  # Enable DNS service
  service_on = true

  # Enable private address spoofing
  # Useful for security when using public DNS servers
  private_address_spoof = true
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

# Outputs
output "dns_config_id" {
  description = "DNS server configuration ID"
  value       = rtx_dns_server.basic.id
}
