# RTX PPTP Configuration Examples
#
# WARNING: PPTP is considered insecure and should only be used
# when no other VPN options are available. Consider using IPsec
# or L2TP/IPsec instead for better security.

terraform {
  required_providers {
    rtx = {
      source  = "github.com/sh1/rtx"
      version = "~> 0.1"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
}

# Basic PPTP server configuration
resource "rtx_pptp" "basic" {
  enabled = true

  authentication {
    method   = "mschap-v2"
    username = var.pptp_username
    password = var.pptp_password
  }

  encryption {
    mppe_bits = "128"
    required  = true
  }

  ip_pool {
    start = "192.168.50.10"
    end   = "192.168.50.50"
  }

  keepalive_enabled = true
  disconnect_time   = 3600
}

# PPTP server with optional encryption (for legacy clients)
resource "rtx_pptp" "legacy_compatible" {
  enabled = true

  authentication {
    method   = "chap"
    username = var.pptp_username
    password = var.pptp_password
  }

  # Encryption optional for legacy client compatibility
  encryption {
    mppe_bits = "128"
    required  = false
  }

  ip_pool {
    start = "10.0.100.10"
    end   = "10.0.100.100"
  }

  keepalive_enabled = true
}

# PPTP server disabled (standby)
resource "rtx_pptp" "standby" {
  enabled  = false
  shutdown = true

  authentication {
    method   = "mschap-v2"
    username = var.pptp_username
    password = var.pptp_password
  }

  encryption {
    mppe_bits = "128"
    required  = true
  }

  ip_pool {
    start = "192.168.51.10"
    end   = "192.168.51.50"
  }
}
