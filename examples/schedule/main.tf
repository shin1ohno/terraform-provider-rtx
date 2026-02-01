# Example: RTX Kron Schedule Resources
#
# This example demonstrates various schedule configurations:
# - Kron policies (command lists)
# - Daily recurring schedules
# - Weekly schedules with day_of_week
# - One-time date-specific schedules

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

# ============================================================================
# Example 1: Basic Policy (reusable command set)
# ============================================================================

# Define a policy with multiple commands (can be referenced by schedules)
resource "rtx_kron_policy" "daily_backup" {
  name = "daily_backup"
  command_lines = [
    "save",
    "show config | mail admin@example.com"
  ]
}

# ============================================================================
# Example 2: Daily Schedule
# ============================================================================

# Create a daily schedule at 3:00 AM
resource "rtx_kron_schedule" "backup_schedule" {
  schedule_id = 1
  name        = "Daily Backup"
  at_time     = "3:00"
  recurring   = true

  command_lines = [
    "save",
    "show config | mail admin@example.com"
  ]
}

# ============================================================================
# Example 3: Weekly Schedule with day_of_week
# ============================================================================

# Create a weekly schedule every Sunday at 2:00 AM
resource "rtx_kron_schedule" "weekly_maintenance" {
  schedule_id = 2
  name        = "Weekly Maintenance"
  at_time     = "2:00"
  day_of_week = "sun"
  recurring   = true

  command_lines = [
    "clear dns cache",
    "clear ip filter dynamic",
    "show status | mail admin@example.com"
  ]
}

# ============================================================================
# Example 4: Weekday Schedule (Monday-Friday)
# ============================================================================

# Weekday monitoring schedule with inline commands
resource "rtx_kron_schedule" "weekday_report" {
  schedule_id = 3
  name        = "Weekday Status Report"
  at_time     = "8:00"
  day_of_week = "mon-fri"
  recurring   = true

  command_lines = [
    "show environment",
    "show ip route summary"
  ]
}

# ============================================================================
# Example 5: One-time Date-specific Schedule
# ============================================================================

# Scheduled maintenance for a specific date
resource "rtx_kron_schedule" "planned_maintenance" {
  schedule_id = 20
  name        = "Planned Maintenance 2025"
  date        = "2025/06/15"
  at_time     = "23:00"
  recurring   = false

  command_lines = [
    "pp disable 1",
    "save"
  ]
}

# ============================================================================
# Example 6: Daily Log Rotation
# ============================================================================

resource "rtx_kron_schedule" "log_rotation" {
  schedule_id = 30
  name        = "Daily Log Rotation"
  at_time     = "0:00"
  recurring   = true

  command_lines = [
    "clear syslog"
  ]
}

# ============================================================================
# Outputs
# ============================================================================

output "backup_schedule_id" {
  description = "ID of the daily backup schedule"
  value       = rtx_kron_schedule.backup_schedule.schedule_id
}

output "policy_name" {
  description = "Name of the backup policy"
  value       = rtx_kron_policy.daily_backup.name
}
