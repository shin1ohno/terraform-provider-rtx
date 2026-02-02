# RTX SNMP Server Configuration Example
#
# Note: SNMP server is a singleton resource - only one configuration per router.

terraform {
  required_version = ">= 1.11"
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.9"
    }
  }
}

provider "rtx" {
  host     = var.rtx_host
  username = var.rtx_username
  password = var.rtx_password
}

# SNMP Server configuration with communities and trap hosts
resource "rtx_snmp_server" "main" {
  location   = "Tokyo Data Center, Rack 12"
  contact    = "noc@example.com"
  chassis_id = "RTX830-Core"

  # Read-only community for monitoring systems
  community {
    name       = var.snmp_community_ro
    permission = "ro"
  }

  # Read-write community for management (use with caution)
  community {
    name       = var.snmp_community_rw
    permission = "rw"
    acl        = "10"
  }

  # Trap receiver - primary monitoring server
  host {
    ip_address = "192.168.1.100"
    version    = "2c"
  }

  # Trap receiver - backup monitoring server
  host {
    ip_address = "192.168.1.101"
    version    = "2c"
  }

  # Enable specific trap types
  enable_traps = ["coldstart", "warmstart", "linkdown", "linkup", "authentication"]
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

variable "snmp_community_ro" {
  description = "SNMP read-only community string"
  type        = string
  sensitive   = true
  default     = "public"
}

variable "snmp_community_rw" {
  description = "SNMP read-write community string"
  type        = string
  sensitive   = true
  default     = "private"
}
