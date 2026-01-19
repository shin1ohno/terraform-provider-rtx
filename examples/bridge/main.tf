# RTX Bridge Configuration Examples
#
# This file demonstrates various bridge configuration scenarios
# for Yamaha RTX routers using the rtx_bridge resource.

terraform {
  required_providers {
    rtx = {
      source = "sh1/rtx"
    }
  }
}

# Example 1: Basic Bridge with Single LAN Member
# Creates a simple bridge with one LAN interface
resource "rtx_bridge" "basic" {
  name    = "bridge1"
  members = ["lan1"]
}

# Example 2: Multi-Member Bridge
# Combines LAN and tunnel interfaces into a single broadcast domain
resource "rtx_bridge" "multi_member" {
  name    = "bridge2"
  members = ["lan2", "tunnel1"]
}

# Example 3: L2VPN Bridge
# Bridges a LAN interface with multiple L2TPv3 tunnels for Layer 2 VPN
resource "rtx_bridge" "l2vpn" {
  name    = "bridge3"
  members = ["lan3", "tunnel1", "tunnel2"]
}

# Example 4: VLAN Bridge
# Bridges VLAN sub-interfaces together
resource "rtx_bridge" "vlan_bridge" {
  name    = "bridge4"
  members = ["lan1/1", "lan1/2"]
}

# Example 5: Bridge with PP Interface
# Bridges a LAN interface with a PPP interface
resource "rtx_bridge" "with_pp" {
  name    = "bridge5"
  members = ["lan1", "pp1"]
}

# Example 6: Empty Bridge
# Creates a bridge with no initial members (members can be added later)
resource "rtx_bridge" "empty" {
  name = "bridge6"
}

# Output examples
output "basic_bridge_name" {
  description = "Name of the basic bridge"
  value       = rtx_bridge.basic.name
}

output "l2vpn_bridge_members" {
  description = "Members of the L2VPN bridge"
  value       = rtx_bridge.l2vpn.members
}
