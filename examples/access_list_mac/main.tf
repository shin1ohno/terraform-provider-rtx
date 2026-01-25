# RTX MAC Access List (Ethernet Filter) Configuration Examples

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

# Basic MAC access list - permit specific MAC addresses
resource "rtx_access_list_mac" "allow_known_devices" {
  name = "allowed_macs"

  entry {
    sequence        = 10
    ace_action      = "permit"
    source_address  = "00:11:22:33:44:55"
    destination_any = true
  }

  entry {
    sequence        = 20
    ace_action      = "permit"
    source_address  = "00:11:22:33:44:66"
    destination_any = true
  }

  # Deny all other traffic (implicit deny at end)
  entry {
    sequence        = 100
    ace_action      = "deny"
    source_any      = true
    destination_any = true
  }
}

# MAC access list with EtherType filtering
# Filter by protocol type (IPv4, IPv6, ARP, etc.)
resource "rtx_access_list_mac" "protocol_filter" {
  name = "protocol_filter"

  # Allow IPv4 traffic (EtherType 0x0800)
  entry {
    sequence        = 10
    ace_action      = "permit"
    source_any      = true
    destination_any = true
    ethertype       = "0x0800"
  }

  # Allow ARP traffic (EtherType 0x0806)
  entry {
    sequence        = 20
    ace_action      = "permit"
    source_any      = true
    destination_any = true
    ethertype       = "0x0806"
  }

  # Allow IPv6 traffic (EtherType 0x86DD)
  entry {
    sequence        = 30
    ace_action      = "permit"
    source_any      = true
    destination_any = true
    ethertype       = "0x86DD"
  }

  # Deny all other protocols
  entry {
    sequence        = 100
    ace_action      = "deny"
    source_any      = true
    destination_any = true
  }
}

# MAC access list with VLAN filtering
resource "rtx_access_list_mac" "vlan_filter" {
  name = "vlan_filter"

  # Allow traffic from VLAN 10
  entry {
    sequence        = 10
    ace_action      = "permit"
    source_any      = true
    destination_any = true
    vlan_id         = 10
  }

  # Allow traffic from VLAN 20
  entry {
    sequence        = 20
    ace_action      = "permit"
    source_any      = true
    destination_any = true
    vlan_id         = 20
  }

  # Deny all other VLAN traffic
  entry {
    sequence        = 100
    ace_action      = "deny"
    source_any      = true
    destination_any = true
  }
}

# Apply MAC access list to interface
resource "rtx_interface_mac_acl" "lan1_filter" {
  interface           = "lan1"
  mac_access_group_in = rtx_access_list_mac.allow_known_devices.name
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
