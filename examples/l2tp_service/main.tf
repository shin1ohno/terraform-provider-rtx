# L2TP service configuration for RTX router
#
# This is a singleton resource - only one can exist per router.
# It controls the global L2TP service state.

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

# Enable L2TP service with both L2TPv3 and L2TPv2 protocols
resource "rtx_l2tp_service" "main" {
  enabled   = true
  protocols = ["l2tpv3", "l2tp"]
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
