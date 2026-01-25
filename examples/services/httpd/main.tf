# RTX HTTPD (Web Interface) Configuration Example
#
# This example demonstrates how to configure the HTTP daemon
# for web-based router management.

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

# Example 1: Allow HTTP access from any interface
resource "rtx_httpd" "web_management" {
  host = "any"
}

# Example 2: Restrict HTTP access to LAN interface only with L2MS proxy access
# resource "rtx_httpd" "secure_web_management" {
#   host         = "lan1"
#   proxy_access = true
# }

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

output "httpd_host" {
  description = "Interface HTTPD is listening on"
  value       = rtx_httpd.web_management.host
}
