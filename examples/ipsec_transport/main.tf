# IPsec transport mode configuration for RTX router
#
# IPsec transport is used for L2TP over IPsec configurations,
# mapping L2TP traffic (UDP 1701) to IPsec tunnels.

terraform {
  required_version = ">= 1.11"
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.9"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
}

# Map L2TP traffic to IPsec tunnel 101
resource "rtx_ipsec_transport" "l2tp_tunnel1" {
  transport_id = 1
  tunnel_id    = 101
  protocol     = "udp"
  port         = 1701
}

# Map L2TP traffic to IPsec tunnel 3 (another remote site)
resource "rtx_ipsec_transport" "l2tp_tunnel2" {
  transport_id = 3
  tunnel_id    = 3
  protocol     = "udp"
  port         = 1701
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
