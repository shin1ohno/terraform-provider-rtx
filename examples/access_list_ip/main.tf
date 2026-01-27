# IPv4 Static Filter (Access List) Examples for RTX Router
#
# This example demonstrates the group-based ACL architecture where:
# - Multiple filter entries are defined in a single resource
# - Filters can be applied to interfaces using the `apply` block
# - Sequence numbers can be manually specified or auto-calculated

terraform {
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.6"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
}

# Example 1: Auto sequence mode with apply block
# Sequence numbers are automatically calculated from sequence_start
resource "rtx_access_list_ip" "wan_inbound" {
  name           = "wan-inbound"
  sequence_start = 100
  sequence_step  = 10

  # Block RFC1918 private networks from WAN
  entry {
    action      = "reject"
    source      = "10.0.0.0/8"
    destination = "*"
    protocol    = "*"
  }

  entry {
    action      = "reject"
    source      = "172.16.0.0/12"
    destination = "*"
    protocol    = "*"
  }

  entry {
    action      = "reject"
    source      = "192.168.0.0/16"
    destination = "*"
    protocol    = "*"
  }

  # Allow established TCP connections
  entry {
    action      = "pass"
    source      = "*"
    destination = "*"
    protocol    = "tcp"
    established = true
  }

  # Allow IPsec traffic
  entry {
    action      = "pass"
    source      = "*"
    destination = "*"
    protocol    = "udp"
    dest_port   = "500"
  }

  entry {
    action      = "pass"
    source      = "*"
    destination = "*"
    protocol    = "udp"
    dest_port   = "4500"
  }

  # Default pass rule
  entry {
    action      = "pass"
    source      = "*"
    destination = "*"
    protocol    = "*"
  }

  # Apply to WAN interface (lan2) for incoming traffic
  apply {
    interface = "lan2"
    direction = "in"
  }
}

# Example 2: Manual sequence mode
# Each entry has an explicit sequence number
resource "rtx_access_list_ip" "lan_security" {
  name = "lan-security"

  # Block NetBIOS ports
  entry {
    sequence    = 1000
    action      = "reject"
    source      = "*"
    destination = "*"
    protocol    = "udp"
    dest_port   = "135-139"
    log         = true
  }

  entry {
    sequence    = 1010
    action      = "reject"
    source      = "*"
    destination = "*"
    protocol    = "tcp"
    dest_port   = "135-139"
    log         = true
  }

  # Block SMB
  entry {
    sequence    = 1020
    action      = "reject"
    source      = "*"
    destination = "*"
    protocol    = "tcp"
    dest_port   = "445"
    log         = true
  }

  # Allow all other traffic
  entry {
    sequence    = 9999
    action      = "pass"
    source      = "*"
    destination = "*"
    protocol    = "*"
  }

  # Apply to LAN interface for outgoing traffic
  apply {
    interface = "lan1"
    direction = "out"
  }
}

# Example 3: ACL without apply block (filters defined but not applied)
# Use rtx_access_list_ip_apply resource to apply these filters
resource "rtx_access_list_ip" "shared_filters" {
  name           = "shared-filters"
  sequence_start = 5000
  sequence_step  = 100

  entry {
    action      = "reject"
    source      = "*"
    destination = "*"
    protocol    = "icmp"
  }

  entry {
    action      = "pass"
    source      = "*"
    destination = "*"
    protocol    = "*"
  }
}

# Apply shared_filters to multiple interfaces using rtx_access_list_ip_apply
resource "rtx_access_list_ip_apply" "shared_to_pp1" {
  access_list = rtx_access_list_ip.shared_filters.name
  interface   = "pp1"
  direction   = "in"
  # filter_ids is optional - if omitted, all entry sequences are applied
}

resource "rtx_access_list_ip_apply" "shared_to_tunnel1" {
  access_list = rtx_access_list_ip.shared_filters.name
  interface   = "tunnel1"
  direction   = "in"
  # Apply specific filter IDs in custom order
  filter_ids = [5100, 5000]
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
