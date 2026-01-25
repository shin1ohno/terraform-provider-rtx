# RTX NAT Masquerade (PAT/NAPT) Configuration Examples

terraform {
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.5"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
}

# =============================================================================
# Dependent Resources
# =============================================================================

resource "rtx_interface" "lan2" {
  name = "lan2"
}

resource "rtx_pppoe" "primary" {
  pp_number      = 1
  name           = "Primary WAN"
  bind_interface = rtx_interface.lan2.interface_name
  username       = "user@provider.ne.jp"
  password       = "example!PASS123"
  auth_method    = "chap"
  always_on      = true
  enabled        = true
}

# =============================================================================
# NAT Masquerade Examples
# =============================================================================

# Basic NAT masquerade for PPPoE connection
# Maps internal network to the PPPoE-assigned IP address
resource "rtx_nat_masquerade" "pppoe" {
  descriptor_id = 1
  outer_address = "ipcp"
  inner_network = "192.168.1.0-192.168.1.255"
}

# NAT masquerade with static port forwarding
# Allows external access to internal services
resource "rtx_nat_masquerade" "with_port_forwarding" {
  descriptor_id = 2
  outer_address = rtx_pppoe.primary.pp_interface
  inner_network = "192.168.2.0-192.168.2.255"

  # Forward HTTP traffic to internal web server
  static_entry {
    entry_number        = 1
    inside_local        = "192.168.2.10"
    inside_local_port   = 80
    outside_global_port = 80
    protocol            = "tcp"
  }

  # Forward HTTPS traffic to internal web server
  static_entry {
    entry_number        = 2
    inside_local        = "192.168.2.10"
    inside_local_port   = 443
    outside_global_port = 443
    protocol            = "tcp"
  }

  # Forward SSH to internal server on non-standard external port
  static_entry {
    entry_number        = 3
    inside_local        = "192.168.2.20"
    inside_local_port   = 22
    outside_global_port = 2222
    protocol            = "tcp"
  }
}

# NAT masquerade with ESP protocol forwarding for VPN passthrough
resource "rtx_nat_masquerade" "vpn_passthrough" {
  descriptor_id = 3
  outer_address = "ipcp"
  inner_network = "192.168.3.0-192.168.3.255"

  # Forward ESP (IPsec) to VPN server
  static_entry {
    entry_number = 1
    inside_local = "192.168.3.100"
    protocol     = "esp"
  }

  # Forward IKE (UDP 500) to VPN server
  static_entry {
    entry_number        = 2
    inside_local        = "192.168.3.100"
    inside_local_port   = 500
    outside_global_port = 500
    protocol            = "udp"
  }

  # Forward NAT-T (UDP 4500) to VPN server
  static_entry {
    entry_number        = 3
    inside_local        = "192.168.3.100"
    inside_local_port   = 4500
    outside_global_port = 4500
    protocol            = "udp"
  }
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
