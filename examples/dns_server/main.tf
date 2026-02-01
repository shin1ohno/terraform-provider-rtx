# Example: Basic DNS Server Configuration
#
# This example shows how to configure DNS server settings on an RTX router.
# Note: DNS server is a singleton resource - only one configuration per router.

terraform {
  required_version = ">= 1.11"
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.8"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
}

# DNS Server Configuration
# Demonstrates available options including server selection and static hosts
resource "rtx_dns_server" "main" {
  # Enable DNS domain lookup (default: true)
  domain_lookup = true

  # Set default domain name for queries
  domain_name = "example.com"

  # Configure up to 3 DNS servers (Google and Cloudflare public DNS)
  name_servers = ["8.8.8.8", "1.1.1.1"]

  # Domain-based DNS server selection
  # Use different DNS servers for specific domains/patterns

  # Internal domain resolution
  server_select {
    priority      = 1
    query_pattern = "internal.example.com"
    server {
      address = "192.168.1.1"
    }
  }

  # Local network resolution with multiple servers
  server_select {
    priority      = 2
    query_pattern = "*.local"
    server {
      address = "10.0.0.1"
    }
    server {
      address = "10.0.0.2"
    }
  }

  # Advanced server select with EDNS and record type filtering
  server_select {
    priority      = 10
    query_pattern = "."
    record_type   = "any"
    server {
      address = "10.0.0.53"
      edns    = true
    }
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
  value       = rtx_dns_server.main.id
}
