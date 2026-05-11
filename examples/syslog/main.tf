# RTX Syslog Configuration Example
#
# Note: Syslog is a singleton resource - only one configuration per router.

terraform {
  required_version = ">= 1.11"
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.13"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
}

# Syslog configuration with multiple hosts and log levels.
# Note: RTX firmware does not honor a non-default port on the `syslog host`
# command — receivers must listen on UDP 514. The provider exposes only the
# address attribute as a result.
resource "rtx_syslog" "main" {
  # Primary syslog server
  host {
    address = "192.168.1.100"
  }

  # Secondary syslog server (same UDP 514 default)
  host {
    address = "192.168.1.101"
  }

  # Remote syslog server
  host {
    address = "syslog.example.com"
  }

  # Source address for syslog messages
  local_address = "192.168.1.1"

  # Syslog facility
  facility = "local0"

  # Log levels
  notice = true
  info   = true
  debug  = false
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
