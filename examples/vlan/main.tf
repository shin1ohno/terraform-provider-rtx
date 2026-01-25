# RTX VLAN Configuration Examples
#
# This example demonstrates how to configure VLANs on Yamaha RTX routers.
# VLANs use 802.1Q tagging to segment network traffic on LAN interfaces.

terraform {
  required_providers {
    rtx = {
      source  = "registry.terraform.io/sh1/rtx"
      version = "~> 0.2"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
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

# =============================================================================
# Base Interface Resource
# =============================================================================

resource "rtx_interface" "lan1" {
  name = "lan1"
}

# =============================================================================
# VLAN Examples with Resource References
# =============================================================================

# Example 1: Basic VLAN (without IP address)
# Creates a VLAN interface for 802.1Q tagging only
resource "rtx_vlan" "basic" {
  vlan_id   = 10
  interface = rtx_interface.lan1.interface_name
  name      = "Basic VLAN"
}

# Example 2: VLAN with IP Address
# Creates a VLAN interface with Layer 3 routing capability
resource "rtx_vlan" "management" {
  vlan_id    = 100
  interface  = rtx_interface.lan1.interface_name
  name       = "Management VLAN"
  ip_address = "192.168.100.1"
  ip_mask    = "255.255.255.0"
  shutdown   = false
}

# Example 3: Multiple VLANs on the same interface
# Demonstrates 802.1Q network segmentation with multiple VLANs on lan1
resource "rtx_vlan" "users" {
  vlan_id    = 20
  interface  = rtx_interface.lan1.interface_name
  name       = "Users VLAN"
  ip_address = "192.168.20.1"
  ip_mask    = "255.255.255.0"
  shutdown   = false
}

resource "rtx_vlan" "servers" {
  vlan_id    = 30
  interface  = rtx_interface.lan1.interface_name
  name       = "Servers VLAN"
  ip_address = "192.168.30.1"
  ip_mask    = "255.255.255.0"
  shutdown   = false
}

resource "rtx_vlan" "guest" {
  vlan_id    = 99
  interface  = rtx_interface.lan1.interface_name
  name       = "Guest VLAN"
  ip_address = "192.168.99.1"
  ip_mask    = "255.255.255.0"
  shutdown   = false
}

# Outputs
output "management_vlan_interface" {
  description = "The computed VLAN interface name for management VLAN"
  value       = rtx_vlan.management.vlan_interface
}

output "all_vlan_ids" {
  description = "All configured VLAN IDs"
  value = [
    rtx_vlan.basic.vlan_id,
    rtx_vlan.management.vlan_id,
    rtx_vlan.users.vlan_id,
    rtx_vlan.servers.vlan_id,
    rtx_vlan.guest.vlan_id,
  ]
}
