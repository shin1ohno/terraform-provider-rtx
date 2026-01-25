# RTX SFTPD (SFTP Daemon) Configuration Example
#
# This example demonstrates how to configure the SFTP daemon
# for secure file transfer to the router.
#
# Note: SFTPD requires SSHD to be enabled for the service to work.
# Make sure to configure rtx_sshd with enabled = true before using SFTPD.

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

# First, ensure SSHD is enabled (SFTPD depends on SSH)
resource "rtx_sshd" "ssh" {
  enabled = true
  hosts   = ["lan1"]
}

# Then configure SFTPD
resource "rtx_sftpd" "file_transfer" {
  hosts = ["lan1"]

  depends_on = [rtx_sshd.ssh]
}

# Example: SFTPD on multiple interfaces
# resource "rtx_sftpd" "multi_interface" {
#   hosts = ["lan1", "lan2"]
#
#   depends_on = [rtx_sshd.ssh]
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

output "sftpd_hosts" {
  description = "Interfaces SFTPD is listening on"
  value       = rtx_sftpd.file_transfer.hosts
}
