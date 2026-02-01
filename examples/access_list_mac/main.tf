# RTX MAC Access List (Ethernet Filter) Configuration Examples
#
# This example demonstrates the group-based MAC ACL architecture where:
# - Multiple filter entries are defined in a single resource
# - Filters can be applied to interfaces using the `apply` block
# - Sequence numbers can be manually specified or auto-calculated
#
# Note: MAC filters are only supported on Ethernet interfaces (lan, bridge).
# PP and Tunnel interfaces are NOT supported for MAC filtering.

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

# Example 1: Auto sequence mode - permit specific MAC addresses
resource "rtx_access_list_mac" "allow_known_devices" {
  name           = "allowed-macs"
  sequence_start = 10
  sequence_step  = 10

  entry {
    ace_action      = "permit"
    source_address  = "00:11:22:33:44:55"
    destination_any = true
  }

  entry {
    ace_action      = "permit"
    source_address  = "00:11:22:33:44:66"
    destination_any = true
  }

  # Deny all other traffic (implicit deny at end)
  entry {
    ace_action      = "deny"
    source_any      = true
    destination_any = true
    log             = true
  }

  # Apply to LAN interface for incoming traffic
  apply {
    interface = "lan1"
    direction = "in"
  }
}

# Example 2: Manual sequence mode with EtherType filtering
# Filter by protocol type (IPv4, IPv6, ARP, etc.)
resource "rtx_access_list_mac" "protocol_filter" {
  name = "protocol-filter"

  # Allow IPv4 traffic (EtherType 0x0800)
  entry {
    sequence        = 10
    ace_action      = "permit"
    source_any      = true
    destination_any = true
    ether_type      = "0x0800"
  }

  # Allow ARP traffic (EtherType 0x0806)
  entry {
    sequence        = 20
    ace_action      = "permit"
    source_any      = true
    destination_any = true
    ether_type      = "0x0806"
  }

  # Allow IPv6 traffic (EtherType 0x86DD)
  entry {
    sequence        = 30
    ace_action      = "permit"
    source_any      = true
    destination_any = true
    ether_type      = "0x86DD"
  }

  # Deny all other protocols
  entry {
    sequence        = 100
    ace_action      = "deny"
    source_any      = true
    destination_any = true
    log             = true
  }

  # Apply to bridge interface
  apply {
    interface = "bridge1"
    direction = "in"
  }
}

# Example 3: VLAN filtering
resource "rtx_access_list_mac" "vlan_filter" {
  name           = "vlan-filter"
  sequence_start = 10
  sequence_step  = 10

  # Allow traffic from VLAN 10
  entry {
    ace_action      = "permit"
    source_any      = true
    destination_any = true
    vlan_id         = 10
  }

  # Allow traffic from VLAN 20
  entry {
    ace_action      = "permit"
    source_any      = true
    destination_any = true
    vlan_id         = 20
  }

  # Deny all other VLAN traffic
  entry {
    ace_action      = "deny"
    source_any      = true
    destination_any = true
  }

  apply {
    interface = "lan1"
    direction = "in"
  }
}

# Example 4: DHCP-based filtering
resource "rtx_access_list_mac" "dhcp_binding" {
  name           = "dhcp-binding"
  sequence_start = 100
  sequence_step  = 10

  # Allow traffic from DHCP-bound clients
  entry {
    ace_action      = "permit"
    source_any      = true
    destination_any = true
    dhcp_match {
      type  = "dhcp-bind"
      scope = 1
    }
  }

  # Deny non-DHCP-bound traffic
  entry {
    ace_action      = "deny"
    source_any      = true
    destination_any = true
    log             = true
  }

  apply {
    interface = "lan1"
    direction = "in"
  }
}

# Example 5: ACL without apply block
# Use rtx_access_list_mac_apply to bind to multiple interfaces
resource "rtx_access_list_mac" "shared_mac_filters" {
  name           = "shared-mac-filters"
  sequence_start = 1000
  sequence_step  = 100

  entry {
    ace_action      = "permit"
    source_any      = true
    destination_any = true
    ether_type      = "0x0800"
  }

  entry {
    ace_action      = "deny"
    source_any      = true
    destination_any = true
  }
}

# Apply shared filters to lan2
resource "rtx_access_list_mac_apply" "shared_to_lan2" {
  access_list = rtx_access_list_mac.shared_mac_filters.name
  interface   = "lan2"
  direction   = "in"
  filter_ids  = [1000, 1100]
}

# Apply shared filters to lan3 with specific filter order
resource "rtx_access_list_mac_apply" "shared_to_lan3" {
  access_list = rtx_access_list_mac.shared_mac_filters.name
  interface   = "lan3"
  direction   = "out"
  filter_ids  = [1100, 1000]
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
