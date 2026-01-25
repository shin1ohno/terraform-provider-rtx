# Example: IPv6 Interface Configuration for Yamaha RTX Routers
#
# This example demonstrates various IPv6 interface configurations:
# - WAN interface with DHCPv6 client (obtains prefix from ISP)
# - LAN interface with Router Advertisement and DHCPv6 server
# - Bridge interface with static IPv6 address
# - Security filter configuration

terraform {
  required_providers {
    rtx = {
      source = "sh1/rtx"
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
# Base Interface Resources
# =============================================================================

resource "rtx_interface" "lan1" {
  name = "lan1"
}

resource "rtx_interface" "lan2" {
  name = "lan2"
}

resource "rtx_interface" "lan3" {
  name = "lan3"
}

resource "rtx_bridge" "bridge1" {
  name    = "bridge1"
  members = [rtx_interface.lan1.interface_name]
}

# ============================================================
# Example 1: WAN Interface with DHCPv6 Client
# ============================================================
# Obtains IPv6 prefix delegation from ISP via DHCPv6
resource "rtx_ipv6_interface" "wan" {
  interface = rtx_interface.lan2.interface_name

  # Use DHCPv6 client to obtain address/prefix from ISP
  dhcpv6_service = "client"
}

# ============================================================
# Example 2: LAN Interface with RA and DHCPv6 Server
# ============================================================
# First, define the IPv6 prefix to be used
resource "rtx_ipv6_prefix" "lan_prefix" {
  id            = 1
  prefix        = "2001:db8::"
  prefix_length = 64
  source        = "static"
}

# Configure LAN interface with SLAAC and DHCPv6
resource "rtx_ipv6_interface" "lan" {
  interface = rtx_interface.lan1.interface_name

  # Static IPv6 address for the router
  address {
    address = "2001:db8::1/64"
  }

  # Router Advertisement configuration
  # Enables SLAAC for clients
  rtadv {
    enabled   = true
    prefix_id = rtx_ipv6_prefix.lan_prefix.id
    o_flag    = true  # Clients should use DHCPv6 for other config (DNS, etc.)
    m_flag    = false # Clients use SLAAC for address configuration
    lifetime  = 1800  # Router lifetime in seconds
  }

  # DHCPv6 server for providing DNS servers and other options
  dhcpv6_service = "server"

  # Set MTU for IPv6
  mtu = 1500
}

# ============================================================
# Example 3: Bridge Interface with Static Address
# ============================================================
resource "rtx_ipv6_interface" "bridge" {
  interface = rtx_bridge.bridge1.interface_name

  # Multiple static IPv6 addresses
  address {
    address = "2001:db8:1::1/64"
  }

  address {
    address = "fd00::1/64" # ULA address for local network
  }
}

# ============================================================
# Example 4: Interface with Prefix-Based Address
# ============================================================
# Uses prefix obtained from another interface via RA or DHCPv6-PD
resource "rtx_ipv6_interface" "prefix_based" {
  interface = rtx_interface.lan3.interface_name

  # Address derived from prefix received on lan2
  address {
    prefix_ref   = "ra-prefix@${rtx_interface.lan2.interface_name}"
    interface_id = "::1/64"
  }

  # Also add a link-local derived address
  address {
    prefix_ref   = "dhcp-prefix@${rtx_interface.lan2.interface_name}"
    interface_id = "::1/64"
  }
}

# ============================================================
# Example 5: Security Filter Configuration
# ============================================================
resource "rtx_ipv6_interface" "secured" {
  interface = rtx_interface.lan1.interface_name

  address {
    address = "2001:db8:2::1/64"
  }

  # Inbound security filters (first match wins)
  # Filter 1: Allow established connections
  # Filter 2: Allow ICMPv6
  # Filter 3: Deny all other
  secure_filter_in = [1, 2, 3]

  # Outbound security filters with stateful inspection
  secure_filter_out  = [10, 20, 30]
  dynamic_filter_out = [100, 101] # Dynamic filters for stateful inspection
}

# ============================================================
# Example 6: Full Configuration Example
# ============================================================
resource "rtx_ipv6_prefix" "full_prefix" {
  id            = 2
  prefix        = "2001:db8:100::"
  prefix_length = 64
  source        = "static"
}

resource "rtx_ipv6_interface" "full_example" {
  interface = rtx_interface.lan1.interface_name

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
    prefix_id = rtx_ipv6_prefix.full_prefix.id
    o_flag    = true # Use DHCPv6 for other configuration
    m_flag    = true # Use DHCPv6 for managed addresses
    lifetime  = 3600 # 1 hour router lifetime
  }

  # DHCPv6 server mode
  dhcpv6_service = "server"

  # IPv6 MTU (minimum 1280)
  mtu = 1500

  # Security filters
  secure_filter_in   = [1, 2, 3, 4, 5]
  secure_filter_out  = [10, 20, 30]
  dynamic_filter_out = [100]
}

# Output the configured interfaces
output "wan_interface" {
  description = "WAN interface configuration"
  value       = rtx_ipv6_interface.wan.interface
}

output "lan_interface" {
  description = "LAN interface configuration"
  value       = rtx_ipv6_interface.lan.interface
}
