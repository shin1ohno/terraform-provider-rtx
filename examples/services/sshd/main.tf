# RTX SSHD (SSH Daemon) Configuration Example
#
# This example demonstrates how to configure the SSH daemon
# for secure remote access to the router.
#
# WARNING: Disabling SSH while connected via SSH may lock you out
# of remote management. Ensure you have alternative access (console, telnet)
# before making changes to SSH configuration.

terraform {
  required_providers {
    rtx = {
      source  = "github.com/sh1/rtx"
      version = "~> 0.2"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
}

# Example 1: Enable SSH service without interface restrictions
resource "rtx_sshd" "ssh_access" {
  enabled = true
}

# Example 2: Enable SSH service restricted to specific interfaces
# resource "rtx_sshd" "secure_ssh_access" {
#   enabled = true
#   hosts   = ["lan1", "lan2"]
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

output "sshd_enabled" {
  description = "Whether SSHD is enabled"
  value       = rtx_sshd.ssh_access.enabled
}

output "sshd_hosts" {
  description = "Interfaces SSHD is listening on"
  value       = rtx_sshd.ssh_access.hosts
}

# Note: The host_key is sensitive and not output here
# You can access it via rtx_sshd.ssh_access.host_key if needed
