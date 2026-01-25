# RTX PPPoE Configuration Examples
#
# This example demonstrates PPPoE and PP interface configuration
# for Yamaha RTX routers, commonly used for NTT FLET'S and other
# Japanese ISP connections.

terraform {
  required_providers {
    rtx = {
      source  = "github.com/sh1/rtx"
      version = "~> 0.2"
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
# Example 1: Basic NTT FLET'S NGN PPPoE Connection
# =============================================================================
# This is the most common setup for Japanese fiber internet using NTT FLET'S.

resource "rtx_pppoe" "flets_primary" {
  pp_number      = 1
  name           = "NTT FLET'S NGN Primary"
  bind_interface = "lan2"
  username       = var.flets_username
  password       = var.flets_password
  auth_method    = "chap"
  always_on      = true
  enabled        = true
}

# PP interface IP configuration for the PPPoE connection
resource "rtx_pp_interface" "flets_primary" {
  pp_number      = rtx_pppoe.flets_primary.pp_number
  ip_address     = "ipcp" # Dynamic IP assignment from ISP
  mtu            = 1454   # Standard MTU for PPPoE (1500 - 8 - 38)
  tcp_mss        = 1414   # TCP MSS limit (MTU - 40)
  nat_descriptor = 1000   # Link to NAT masquerade

  # Basic security filters for WAN interface
  secure_filter_in  = [200020, 200021, 200022, 200023, 200024, 200025, 200099]
  secure_filter_out = [200020, 200021, 200022, 200023, 200024, 200025, 200099]

  depends_on = [rtx_pppoe.flets_primary]
}

# NAT masquerade for FLET'S connection
resource "rtx_nat_masquerade" "flets_primary" {
  descriptor_id = 1000
  outer_address = "ipcp" # Use the PPPoE assigned address
  inner_network = "192.168.1.0/24"
  enabled       = true
}

# =============================================================================
# Example 2: Multi-WAN Failover with Two ISPs
# =============================================================================
# Configure two PPPoE connections for redundancy.
# Primary connection via NTT FLET'S, secondary via another ISP.

resource "rtx_pppoe" "primary_wan" {
  pp_number      = 1
  name           = "Primary WAN - ISP-A"
  bind_interface = "lan2"
  username       = var.primary_isp_username
  password       = var.primary_isp_password
  auth_method    = "chap"
  always_on      = true
  enabled        = true
}

resource "rtx_pp_interface" "primary_wan" {
  pp_number      = rtx_pppoe.primary_wan.pp_number
  ip_address     = "ipcp"
  mtu            = 1454
  tcp_mss        = 1414
  nat_descriptor = 1001

  secure_filter_in  = [200020, 200021, 200022, 200023, 200024, 200025, 200099]
  secure_filter_out = [200020, 200021, 200022, 200023, 200024, 200025, 200099]

  depends_on = [rtx_pppoe.primary_wan]
}

resource "rtx_pppoe" "secondary_wan" {
  pp_number      = 2
  name           = "Secondary WAN - ISP-B"
  bind_interface = "lan3"
  username       = var.secondary_isp_username
  password       = var.secondary_isp_password
  auth_method    = "chap"
  always_on      = true
  enabled        = true
}

resource "rtx_pp_interface" "secondary_wan" {
  pp_number      = rtx_pppoe.secondary_wan.pp_number
  ip_address     = "ipcp"
  mtu            = 1454
  tcp_mss        = 1414
  nat_descriptor = 1002

  secure_filter_in  = [200020, 200021, 200022, 200023, 200024, 200025, 200099]
  secure_filter_out = [200020, 200021, 200022, 200023, 200024, 200025, 200099]

  depends_on = [rtx_pppoe.secondary_wan]
}

# NAT masquerade for primary WAN
resource "rtx_nat_masquerade" "primary_wan" {
  descriptor_id = 1001
  outer_address = "ipcp"
  inner_network = "192.168.1.0/24"
  enabled       = true
}

# NAT masquerade for secondary WAN
resource "rtx_nat_masquerade" "secondary_wan" {
  descriptor_id = 1002
  outer_address = "ipcp"
  inner_network = "192.168.1.0/24"
  enabled       = true
}

# Static routes for failover (primary via pp1, fallback to pp2)
resource "rtx_static_route" "default_primary" {
  destination = "default"
  gateway     = "pp 1"
  metric      = 1
}

resource "rtx_static_route" "default_secondary" {
  destination = "default"
  gateway     = "pp 2"
  metric      = 10 # Higher metric = lower priority
}

# =============================================================================
# Example 3: PPPoE with PAP Authentication (Legacy ISP)
# =============================================================================
# Some older ISPs may require PAP authentication instead of CHAP.

resource "rtx_pppoe" "legacy_isp" {
  pp_number      = 3
  name           = "Legacy ISP Connection"
  bind_interface = "lan2"
  username       = var.legacy_isp_username
  password       = var.legacy_isp_password
  auth_method    = "pap" # Use PAP for legacy compatibility
  always_on      = true
  enabled        = true
}

resource "rtx_pp_interface" "legacy_isp" {
  pp_number      = rtx_pppoe.legacy_isp.pp_number
  ip_address     = "ipcp"
  mtu            = 1454
  tcp_mss        = 1414
  nat_descriptor = 1003

  depends_on = [rtx_pppoe.legacy_isp]
}

# =============================================================================
# Example 4: PPPoE with Service Name (Multiple VLAN Services)
# =============================================================================
# Some ISPs provide multiple services on the same physical connection,
# distinguished by service name (e.g., Internet, VoIP, IPTV).

resource "rtx_pppoe" "internet_service" {
  pp_number      = 4
  name           = "Internet Service"
  bind_interface = "lan2"
  username       = var.multi_service_username
  password       = var.multi_service_password
  service_name   = "INTERNET" # Specific service name
  auth_method    = "chap"
  always_on      = true
  enabled        = true
}

resource "rtx_pp_interface" "internet_service" {
  pp_number      = rtx_pppoe.internet_service.pp_number
  ip_address     = "ipcp"
  mtu            = 1454
  tcp_mss        = 1414
  nat_descriptor = 1004

  depends_on = [rtx_pppoe.internet_service]
}

# =============================================================================
# Example 5: PPPoE with Disconnect Timeout (On-Demand Connection)
# =============================================================================
# Configure PPPoE to disconnect after a period of inactivity.
# Useful for metered connections or backup links.

resource "rtx_pppoe" "on_demand" {
  pp_number          = 5
  name               = "On-Demand Backup Link"
  bind_interface     = "lan3"
  username           = var.backup_isp_username
  password           = var.backup_isp_password
  auth_method        = "chap"
  always_on          = false # Allow disconnection
  disconnect_timeout = 300   # Disconnect after 5 minutes of idle
  enabled            = true
}

resource "rtx_pp_interface" "on_demand" {
  pp_number      = rtx_pppoe.on_demand.pp_number
  ip_address     = "ipcp"
  mtu            = 1454
  tcp_mss        = 1414
  nat_descriptor = 1005

  depends_on = [rtx_pppoe.on_demand]
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

# FLET'S credentials
variable "flets_username" {
  description = "NTT FLET'S PPPoE username (format: username@provider.ne.jp)"
  type        = string
  default     = "user@example.ne.jp"
}

variable "flets_password" {
  description = "NTT FLET'S PPPoE password"
  type        = string
  sensitive   = true
  default     = ""
}

# Multi-WAN credentials
variable "primary_isp_username" {
  description = "Primary ISP PPPoE username"
  type        = string
  default     = "primary@isp-a.ne.jp"
}

variable "primary_isp_password" {
  description = "Primary ISP PPPoE password"
  type        = string
  sensitive   = true
  default     = ""
}

variable "secondary_isp_username" {
  description = "Secondary ISP PPPoE username"
  type        = string
  default     = "secondary@isp-b.ne.jp"
}

variable "secondary_isp_password" {
  description = "Secondary ISP PPPoE password"
  type        = string
  sensitive   = true
  default     = ""
}

# Legacy ISP credentials
variable "legacy_isp_username" {
  description = "Legacy ISP PPPoE username"
  type        = string
  default     = "legacy-user"
}

variable "legacy_isp_password" {
  description = "Legacy ISP PPPoE password"
  type        = string
  sensitive   = true
  default     = ""
}

# Multi-service credentials
variable "multi_service_username" {
  description = "Multi-service ISP PPPoE username"
  type        = string
  default     = "multiservice@provider.ne.jp"
}

variable "multi_service_password" {
  description = "Multi-service ISP PPPoE password"
  type        = string
  sensitive   = true
  default     = ""
}

# Backup ISP credentials
variable "backup_isp_username" {
  description = "Backup ISP PPPoE username"
  type        = string
  default     = "backup@provider.ne.jp"
}

variable "backup_isp_password" {
  description = "Backup ISP PPPoE password"
  type        = string
  sensitive   = true
  default     = ""
}

# =============================================================================
# Outputs
# =============================================================================

output "flets_connection" {
  description = "FLET'S primary connection details"
  value = {
    pp_number      = rtx_pppoe.flets_primary.pp_number
    name           = rtx_pppoe.flets_primary.name
    bind_interface = rtx_pppoe.flets_primary.bind_interface
    enabled        = rtx_pppoe.flets_primary.enabled
  }
}

output "multi_wan_connections" {
  description = "Multi-WAN connection details"
  value = {
    primary = {
      pp_number = rtx_pppoe.primary_wan.pp_number
      name      = rtx_pppoe.primary_wan.name
    }
    secondary = {
      pp_number = rtx_pppoe.secondary_wan.pp_number
      name      = rtx_pppoe.secondary_wan.name
    }
  }
}
