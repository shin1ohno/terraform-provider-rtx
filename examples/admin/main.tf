# RTX Admin Configuration Example
#
# This example demonstrates how to configure admin passwords and user accounts
# on a Yamaha RTX router using the Terraform provider.

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

# Variables for sensitive information
variable "rtx_host" {
  description = "RTX router hostname or IP address"
  type        = string
}

variable "rtx_username" {
  description = "Username for RTX router authentication"
  type        = string
}

variable "rtx_password" {
  description = "Password for RTX router authentication"
  type        = string
  sensitive   = true
}

variable "login_password" {
  description = "Login password for the RTX router"
  type        = string
  sensitive   = true
  default     = ""
}

variable "admin_password" {
  description = "Administrator password for the RTX router"
  type        = string
  sensitive   = true
  default     = ""
}

# Configure router-level passwords (singleton resource)
# This sets the login password and administrator password for the router itself
resource "rtx_admin" "main" {
  login_password = var.login_password
  admin_password = var.admin_password
}

# Create an administrator user with full access
resource "rtx_admin_user" "admin" {
  username      = "admin"
  password      = "SecureAdminPass123!"
  administrator = true
  connection    = ["ssh", "telnet", "http", "serial"]
  gui_pages     = ["dashboard", "lan-map", "config"]
  login_timer   = 0 # No timeout
}

# Create an operator user with limited access
resource "rtx_admin_user" "operator" {
  username      = "operator"
  password      = "OperatorPass456!"
  administrator = false
  connection    = ["ssh", "http"]
  gui_pages     = ["dashboard", "lan-map"]
  login_timer   = 300 # 5 minute timeout
}

# Create a read-only user for monitoring
resource "rtx_admin_user" "monitor" {
  username      = "monitor"
  password      = "MonitorPass789!"
  administrator = false
  connection    = ["http"]
  gui_pages     = ["dashboard"]
  login_timer   = 600 # 10 minute timeout
}

# Outputs
output "admin_user_created" {
  description = "Admin user has been created"
  value       = rtx_admin_user.admin.username
}

output "operator_user_created" {
  description = "Operator user has been created"
  value       = rtx_admin_user.operator.username
}

output "monitor_user_created" {
  description = "Monitor user has been created"
  value       = rtx_admin_user.monitor.username
}
