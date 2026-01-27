# RTX System Configuration Example
#
# This example demonstrates how to configure system-level settings
# on a Yamaha RTX router.

terraform {
  required_version = ">= 1.11"
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.7"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password

  # Optionally set admin password if different from login password
  # admin_password = var.rtx_admin_password

  # Skip host key check for testing (not recommended for production)
  skip_host_key_check = var.skip_host_key_check
}

# Configure system settings
resource "rtx_system" "main" {
  # Set timezone to JST (Japan Standard Time)
  timezone = "+09:00"

  # Configure console settings
  console {
    character = "ja.utf8"
    lines     = "infinity"
    prompt    = "[RTX1210] "
  }

  # Tune packet buffers for high-performance networking
  packet_buffer {
    size       = "small"
    max_buffer = 5000
    max_free   = 1300
  }

  packet_buffer {
    size       = "middle"
    max_buffer = 10000
    max_free   = 4950
  }

  packet_buffer {
    size       = "large"
    max_buffer = 20000
    max_free   = 5600
  }

  # Enable statistics collection
  statistics {
    traffic = true
    nat     = true
  }
}

# Output the configured timezone
output "configured_timezone" {
  value       = rtx_system.main.timezone
  description = "The configured timezone"
}
