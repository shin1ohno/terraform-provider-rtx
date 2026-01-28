# IPv6 Static Filter (Access List) Examples for RTX Router
#
# This example demonstrates the group-based ACL architecture where:
# - Multiple filter entries are defined in a single resource
# - Filters can be applied to interfaces using the `apply` block
# - Sequence numbers can be manually specified or auto-calculated

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

# Example 1: Auto sequence mode with apply block
# Essential IPv6 filters for WAN interface
resource "rtx_access_list_ipv6" "wan_inbound" {
  name           = "wan-inbound-ipv6"
  sequence_start = 100
  sequence_step  = 10

  # Allow ICMPv6 (required for IPv6 to function properly)
  entry {
    action      = "pass"
    source      = "*"
    destination = "*"
    protocol    = "icmp6"
  }

  # Allow DHCPv6 client (UDP 546)
  entry {
    action      = "pass"
    source      = "*"
    destination = "*"
    protocol    = "udp"
    dest_port   = "546"
  }

  # Allow DHCPv6 server (UDP 547)
  entry {
    action      = "pass"
    source      = "*"
    destination = "*"
    protocol    = "udp"
    dest_port   = "547"
  }

  # Default pass rule
  entry {
    action      = "pass"
    source      = "*"
    destination = "*"
    protocol    = "*"
  }

  # Apply to WAN interface for incoming traffic
  apply {
    interface = "lan2"
    direction = "in"
  }
}

# Example 2: Manual sequence mode with logging
# SSH access control from trusted prefixes
resource "rtx_access_list_ipv6" "ssh_access" {
  name = "ssh-access-ipv6"

  # Allow SSH from documentation prefix (example)
  entry {
    sequence    = 1000
    action      = "pass"
    source      = "2001:db8::/32"
    destination = "*"
    protocol    = "tcp"
    dest_port   = "22"
    log         = true
  }

  # Allow SSH from another trusted prefix
  entry {
    sequence    = 1010
    action      = "pass"
    source      = "2001:db8:1::/48"
    destination = "*"
    protocol    = "tcp"
    dest_port   = "22"
    log         = true
  }

  # Reject all other SSH attempts
  entry {
    sequence    = 1020
    action      = "reject"
    source      = "*"
    destination = "*"
    protocol    = "tcp"
    dest_port   = "22"
    log         = true
  }

  # Pass all other traffic
  entry {
    sequence    = 9999
    action      = "pass"
    source      = "*"
    destination = "*"
    protocol    = "*"
  }

  # Apply to LAN interface
  apply {
    interface = "lan1"
    direction = "in"
  }
}

# Example 3: ACL without apply block
# Use rtx_access_list_ipv6_apply to bind to interfaces
resource "rtx_access_list_ipv6" "shared_filters" {
  name           = "shared-ipv6-filters"
  sequence_start = 5000
  sequence_step  = 100

  # Block specific protocol
  entry {
    action      = "reject"
    source      = "*"
    destination = "*"
    protocol    = "udp"
    dest_port   = "53"
  }

  # Pass everything else
  entry {
    action      = "pass"
    source      = "*"
    destination = "*"
    protocol    = "*"
  }
}

# Apply shared filters to PP interface
resource "rtx_access_list_ipv6_apply" "shared_to_pp1" {
  access_list = rtx_access_list_ipv6.shared_filters.name
  interface   = "pp1"
  direction   = "in"
}

# Apply shared filters to tunnel interface with specific filter order
resource "rtx_access_list_ipv6_apply" "shared_to_tunnel1" {
  access_list = rtx_access_list_ipv6.shared_filters.name
  interface   = "tunnel1"
  direction   = "in"
  filter_ids  = [5100, 5000]
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
