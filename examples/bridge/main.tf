# RTX Bridge Configuration Examples
#
# This file demonstrates various bridge configuration scenarios
# for Yamaha RTX routers using the rtx_bridge resource.

terraform {
  required_version = ">= 1.11"
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.9"
    }
  }
}

# =============================================================================
# Dependent Resources (for reference examples)
# =============================================================================

# Interface resources
resource "rtx_interface" "lan1" {
  name = "lan1"
}

resource "rtx_interface" "lan2" {
  name = "lan2"
}

resource "rtx_interface" "lan3" {
  name = "lan3"
}

# L2TP tunnel resources
resource "rtx_l2tp" "tunnel1" {
  tunnel_id          = 1
  name               = "l2tp-tunnel-1"
  version            = "l2tpv3"
  mode               = "l2vpn"
  tunnel_destination = "192.0.2.1"

  l2tpv3_config {
    local_router_id  = "192.168.1.1"
    remote_router_id = "192.168.1.2"
    remote_end_id    = "remote1"
  }
}

resource "rtx_l2tp" "tunnel2" {
  tunnel_id          = 2
  name               = "l2tp-tunnel-2"
  version            = "l2tpv3"
  mode               = "l2vpn"
  tunnel_destination = "192.0.2.2"

  l2tpv3_config {
    local_router_id  = "192.168.1.1"
    remote_router_id = "192.168.1.3"
    remote_end_id    = "remote2"
  }
}

# =============================================================================
# Bridge Examples with Resource References
# =============================================================================

# Example 1: Basic Bridge with Single LAN Member
# Creates a simple bridge with one LAN interface
resource "rtx_bridge" "basic" {
  name    = "bridge1"
  members = [rtx_interface.lan1.interface_name]
}

# Example 2: Multi-Member Bridge
# Combines LAN and tunnel interfaces into a single broadcast domain
resource "rtx_bridge" "multi_member" {
  name    = "bridge2"
  members = [rtx_interface.lan2.interface_name, rtx_l2tp.tunnel1.tunnel_interface]
}

# Example 3: L2VPN Bridge
# Bridges a LAN interface with multiple L2TPv3 tunnels for Layer 2 VPN
resource "rtx_bridge" "l2vpn" {
  name = "bridge3"
  members = [
    rtx_interface.lan3.interface_name,
    rtx_l2tp.tunnel1.tunnel_interface,
    rtx_l2tp.tunnel2.tunnel_interface
  ]
}

# Example 4: Empty Bridge
# Creates a bridge with no initial members (members can be added later)
resource "rtx_bridge" "empty" {
  name = "bridge4"
}

# =============================================================================
# Output examples
# =============================================================================

output "basic_bridge_name" {
  description = "Name of the basic bridge"
  value       = rtx_bridge.basic.name
}

output "l2vpn_bridge_members" {
  description = "Members of the L2VPN bridge"
  value       = rtx_bridge.l2vpn.members
}
