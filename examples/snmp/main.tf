# Basic SNMP configuration with read-only community
resource "rtx_snmp_server" "basic" {
  location   = "Server Room A"
  contact    = "admin@example.com"
  chassis_id = "RTX830-Main"

  community {
    name       = var.snmp_community_ro
    permission = "ro"
  }
}

# Full SNMP configuration with multiple communities, trap hosts, and security settings
resource "rtx_snmp_server" "full" {
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
    acl        = "10" # Restrict to specific hosts via ACL
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

# Minimal SNMP configuration - monitoring only
resource "rtx_snmp_server" "monitoring_only" {
  chassis_id = "RTX830-Branch01"

  community {
    name       = var.snmp_community_ro
    permission = "ro"
    acl        = "20" # Only allow monitoring subnet
  }

  host {
    ip_address = "10.0.0.50"
    version    = "2c"
  }

  enable_traps = ["all"]
}
