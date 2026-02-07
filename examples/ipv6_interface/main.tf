# Example: IPv6 Interface Configuration for Yamaha RTX Routers
#
# This example demonstrates various IPv6 interface configurations:
# - LAN interface with Router Advertisement and DHCPv6 server
# - Additional interface with static IPv6 address

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
  host     = var.router_host
  username = var.router_username
  password = var.router_password
}

variable "router_host" {
  description = "RTX router hostname or IP address"
  type        = string
}

variable "router_username" {
  description = "Username for RTX router authentication"
  type        = string
}

variable "router_password" {
  description = "Password for RTX router authentication"
  type        = string
  sensitive   = true
}

# =============================================================================
# Example 1: LAN Interface with RA and DHCPv6 Server
# =============================================================================
# Define the IPv6 prefix to be used
resource "rtx_ipv6_prefix" "lan_prefix" {
  prefix_id     = 1
  prefix        = "2001:db8::"
  prefix_length = 64
  source        = "static"
}

# Configure LAN interface with SLAAC and DHCPv6
resource "rtx_ipv6_interface" "lan" {
  interface = "lan1"

  # Static IPv6 address for the router
  address {
    address = "2001:db8::1/64"
  }

  # Router Advertisement configuration
  # Enables SLAAC for clients
  rtadv {
    enabled   = true
    prefix_id = rtx_ipv6_prefix.lan_prefix.prefix_id
    o_flag    = true  # Clients should use DHCPv6 for other config (DNS, etc.)
    m_flag    = false # Clients use SLAAC for address configuration
    lifetime  = 1800  # Router lifetime in seconds
  }

  # DHCPv6 server for providing DNS servers and other options
  dhcpv6_service = "server"

  # Set MTU for IPv6
  mtu = 1500
}

# =============================================================================
# Example 2: WAN Interface with DHCPv6 Client
# =============================================================================
# Obtains IPv6 prefix delegation from ISP via DHCPv6
resource "rtx_ipv6_prefix" "wan_prefix" {
  prefix_id     = 2
  prefix_length = 64
  source        = "dhcpv6-pd"
  interface     = "lan2"
}

resource "rtx_ipv6_interface" "wan" {
  interface = "lan2"

  # Use DHCPv6 client to obtain address/prefix from ISP
  dhcpv6_service = "client"

  rtadv {
    enabled   = false
    prefix_id = rtx_ipv6_prefix.wan_prefix.prefix_id
  }
}

# =============================================================================
# Example 3: Full Configuration Example
# =============================================================================
resource "rtx_ipv6_prefix" "full_prefix" {
  prefix_id     = 3
  prefix        = "2001:db8:100::"
  prefix_length = 64
  source        = "static"
}

resource "rtx_ipv6_interface" "full_example" {
  interface = "lan3"

  # Primary address
  address {
    address = "2001:db8:100::1/64"
  }

  # Secondary address (link-local)
  address {
    address = "fe80::1/10"
  }

  # Router Advertisement with all options
  rtadv {
    enabled   = true
    prefix_id = rtx_ipv6_prefix.full_prefix.prefix_id
    o_flag    = true # Use DHCPv6 for other configuration
    m_flag    = true # Use DHCPv6 for managed addresses
    lifetime  = 3600 # 1 hour router lifetime
  }

  # DHCPv6 server mode
  dhcpv6_service = "server"

  # IPv6 MTU (minimum 1280)
  mtu = 1500
}

# Output the configured interfaces
output "lan_interface" {
  description = "LAN interface configuration"
  value       = rtx_ipv6_interface.lan.interface
}

output "wan_interface" {
  description = "WAN interface configuration"
  value       = rtx_ipv6_interface.wan.interface
}
