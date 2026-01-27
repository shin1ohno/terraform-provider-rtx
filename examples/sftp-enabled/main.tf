# SFTP-Enabled RTX Router Configuration Example
#
# This example demonstrates a complete setup for enabling SFTP on an RTX router.
# SFTP (SSH File Transfer Protocol) enables secure file transfers to and from
# the router, which is useful for configuration backups, firmware updates,
# and accessing log files.
#
# Prerequisites:
#   - SSH must be enabled before SFTP can be used
#   - Users need appropriate connection permissions for SFTP access
#
# Usage:
#   terraform init
#   terraform plan -var-file="terraform.tfvars"
#   terraform apply -var-file="terraform.tfvars"

terraform {
  required_version = ">= 1.11"
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.7"
    }
  }
}

# Provider configuration
# For production use, always verify the SSH host key
provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password

  # Security: Verify the router's SSH host key
  # Option 1: Use known_hosts file (default)
  # known_hosts_file = "~/.ssh/known_hosts"

  # Option 2: Provide the host key directly (recommended for automation)
  # ssh_host_key = var.rtx_ssh_host_key

  # Option 3: Skip verification (TESTING ONLY - NOT SECURE)
  # skip_host_key_check = true
}

# =============================================================================
# Step 1: Enable SSH daemon (required for SFTP)
# =============================================================================
# The SSH daemon must be enabled before SFTP can function.
# This configures sshd to listen on the specified interfaces.

resource "rtx_sshd" "ssh" {
  enabled = true

  # Restrict SSH access to specific interfaces for security
  # Options: "lan1", "lan2", "wan1", etc.
  # Omit hosts to allow SSH from all interfaces
  hosts = var.ssh_interfaces
}

# =============================================================================
# Step 2: Enable SFTP daemon
# =============================================================================
# The SFTP daemon provides secure file transfer functionality.
# It depends on SSH being enabled first.

resource "rtx_sftpd" "sftp" {
  # List of interfaces that SFTP will accept connections from
  # For security, restrict to internal/management interfaces only
  hosts = var.sftp_interfaces

  # Ensure SSH is enabled before configuring SFTP
  depends_on = [rtx_sshd.ssh]
}

# =============================================================================
# Step 3: Configure admin user with SFTP access (optional)
# =============================================================================
# Create an admin user that can access the router via SFTP.
# This user will have full administrative privileges.

resource "rtx_admin_user" "terraform_admin" {
  name     = var.admin_username
  password = var.admin_password

  # Enable administrator privileges
  # This allows the user to make configuration changes
  administrator = "on"

  # Allow connections via SSH and SFTP
  # connection = "ssh,sftp"
}

# =============================================================================
# Variables
# =============================================================================

variable "rtx_host" {
  description = "RTX router hostname or IP address"
  type        = string
}

variable "rtx_username" {
  description = "RTX router login username"
  type        = string
}

variable "rtx_password" {
  description = "RTX router login password"
  type        = string
  sensitive   = true
}

variable "admin_username" {
  description = "Admin user name for SFTP access"
  type        = string
  default     = "terraform"
}

variable "admin_password" {
  description = "Admin user password"
  type        = string
  sensitive   = true
}

variable "ssh_interfaces" {
  description = "Interfaces to enable SSH on"
  type        = list(string)
  default     = ["lan1"]
}

variable "sftp_interfaces" {
  description = "Interfaces to enable SFTP on"
  type        = list(string)
  default     = ["lan1"]
}

# =============================================================================
# Outputs
# =============================================================================

output "sshd_enabled" {
  description = "Whether SSH daemon is enabled"
  value       = rtx_sshd.ssh.enabled
}

output "sshd_hosts" {
  description = "Interfaces SSH is listening on"
  value       = rtx_sshd.ssh.hosts
}

output "sftpd_hosts" {
  description = "Interfaces SFTP is listening on"
  value       = rtx_sftpd.sftp.hosts
}

output "admin_user" {
  description = "Admin user configured for SFTP access"
  value       = rtx_admin_user.terraform_admin.name
}
