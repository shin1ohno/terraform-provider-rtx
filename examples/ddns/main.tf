# RTX DDNS Configuration Examples
#
# This example demonstrates DDNS (Dynamic DNS) configuration for Yamaha RTX routers.
# Supports both NetVolante DNS (Yamaha's free service) and custom DDNS providers.

terraform {
  required_providers {
    rtx = {
      source = "shin1ohno/rtx"
      version = "~> 0.5"
    }
  }
}

provider "rtx" {
  host                = var.rtx_host
  username            = var.rtx_username
  password            = var.rtx_password
  admin_password      = var.rtx_admin_password
  skip_host_key_check = true
}

# =============================================================================
# Example 1: NetVolante DNS (Yamaha's Free DDNS Service)
# =============================================================================
# NetVolante DNS is a free DDNS service provided by Yamaha for RTX routers.
# It provides hostnames under the netvolante.jp domain.

resource "rtx_netvolante_dns" "primary" {
  interface    = "lan2"                       # Interface to monitor for IP changes
  hostname     = "myrouter.aa0.netvolante.jp" # Your NetVolante DNS hostname
  server       = 1                            # NetVolante server (1 or 2)
  timeout      = 60                           # Update timeout in seconds
  ipv6_enabled = false                        # Enable IPv6 address registration
}

# =============================================================================
# Example 2: Custom DDNS Provider (DynDNS, No-IP, etc.)
# =============================================================================
# Configure a custom DDNS provider using the generic DDNS resource.
# This supports any DDNS provider that uses HTTP-based updates.

# DynDNS-style provider
resource "rtx_ddns" "dyndns" {
  server_id = 1
  url       = "https://members.dyndns.org/nic/update"
  hostname  = var.dyndns_hostname
  username  = var.dyndns_username
  password  = var.dyndns_password
}

# No-IP provider
resource "rtx_ddns" "noip" {
  server_id = 2
  url       = "https://dynupdate.no-ip.com/nic/update"
  hostname  = var.noip_hostname
  username  = var.noip_username
  password  = var.noip_password
}

# =============================================================================
# Example 3: Multiple DDNS Services for Redundancy
# =============================================================================
# Configure multiple DDNS services for high availability.
# If one service fails, others can still provide name resolution.

resource "rtx_netvolante_dns" "backup" {
  interface    = "lan2"
  hostname     = "backup.aa0.netvolante.jp"
  server       = 2 # Use server 2 for redundancy
  timeout      = 60
  ipv6_enabled = false
}

resource "rtx_ddns" "cloudflare_style" {
  server_id = 3
  url       = "https://api.cloudflare.com/client/v4/zones/${var.cloudflare_zone_id}/dns_records/${var.cloudflare_record_id}"
  hostname  = var.cloudflare_hostname
  username  = var.cloudflare_email
  password  = var.cloudflare_api_key
}

# =============================================================================
# Example 4: DDNS Status Monitoring
# =============================================================================
# Use the data source to monitor DDNS registration status.

data "rtx_ddns_status" "all" {
  type = "all" # "netvolante", "custom", or "all"
}

# =============================================================================
# Variables
# =============================================================================

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

variable "rtx_admin_password" {
  description = "RTX router administrator password"
  type        = string
  sensitive   = true
}

# DynDNS credentials
variable "dyndns_hostname" {
  description = "DynDNS hostname"
  type        = string
  default     = "myhost.dyndns.org"
}

variable "dyndns_username" {
  description = "DynDNS username"
  type        = string
  default     = ""
}

variable "dyndns_password" {
  description = "DynDNS password"
  type        = string
  sensitive   = true
  default     = ""
}

# No-IP credentials
variable "noip_hostname" {
  description = "No-IP hostname"
  type        = string
  default     = "myhost.no-ip.org"
}

variable "noip_username" {
  description = "No-IP username/email"
  type        = string
  default     = ""
}

variable "noip_password" {
  description = "No-IP password"
  type        = string
  sensitive   = true
  default     = ""
}

# Cloudflare variables
variable "cloudflare_zone_id" {
  description = "Cloudflare zone ID"
  type        = string
  default     = ""
}

variable "cloudflare_record_id" {
  description = "Cloudflare DNS record ID"
  type        = string
  default     = ""
}

variable "cloudflare_hostname" {
  description = "Cloudflare hostname"
  type        = string
  default     = "myhost.example.com"
}

variable "cloudflare_email" {
  description = "Cloudflare account email"
  type        = string
  default     = ""
}

variable "cloudflare_api_key" {
  description = "Cloudflare API key"
  type        = string
  sensitive   = true
  default     = ""
}

# =============================================================================
# Outputs
# =============================================================================

output "netvolante_dns" {
  description = "NetVolante DNS configuration"
  value = {
    primary = {
      interface = rtx_netvolante_dns.primary.interface
      hostname  = rtx_netvolante_dns.primary.hostname
    }
    backup = {
      interface = rtx_netvolante_dns.backup.interface
      hostname  = rtx_netvolante_dns.backup.hostname
    }
  }
}

output "custom_ddns" {
  description = "Custom DDNS configuration"
  value = {
    dyndns = {
      server_id = rtx_ddns.dyndns.server_id
      hostname  = rtx_ddns.dyndns.hostname
    }
    noip = {
      server_id = rtx_ddns.noip.server_id
      hostname  = rtx_ddns.noip.hostname
    }
  }
}

output "ddns_status" {
  description = "Current DDNS registration status"
  value       = data.rtx_ddns_status.all.statuses
}
