# QoS Configuration Examples for Yamaha RTX Routers
# This file demonstrates various QoS configurations

terraform {
  required_version = ">= 1.11"
  required_providers {
    rtx = {
      source  = "shin1ohno/rtx"
      version = "~> 0.10"
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

# =============================================================================
# Base Interface Resources
# =============================================================================

resource "rtx_interface" "lan1" {
  name = "lan1"
}

resource "rtx_interface" "lan2" {
  name = "lan2"
}

# ============================================================================
# Example 1: Basic Class Map for VoIP Traffic
# ============================================================================
# This class map identifies VoIP traffic based on SIP port

resource "rtx_class_map" "voip" {
  name                   = "voip-traffic"
  match_protocol         = "sip"
  match_destination_port = [5060, 5061]
}

# ============================================================================
# Example 2: Class Map with DSCP Matching
# ============================================================================
# This class map matches traffic by DSCP value

resource "rtx_class_map" "expedited" {
  name       = "expedited-traffic"
  match_dscp = "ef"
}

# ============================================================================
# Example 3: Policy Map with Multiple Classes
# ============================================================================
# This policy map defines QoS actions for different traffic classes

resource "rtx_policy_map" "qos_policy" {
  name = "enterprise-qos"

  # High priority for VoIP traffic (30% bandwidth guaranteed)
  class {
    name              = rtx_class_map.voip.name
    priority          = "high"
    bandwidth_percent = 30
    queue_limit       = 64
  }

  # High priority for expedited traffic (20% bandwidth)
  class {
    name              = rtx_class_map.expedited.name
    priority          = "high"
    bandwidth_percent = 20
  }

  # Low priority for bulk data (50% bandwidth)
  class {
    name              = "bulk"
    priority          = "low"
    bandwidth_percent = 50
  }
}

# ============================================================================
# Example 4: Service Policy Attachment
# ============================================================================
# Attach the policy map to an interface

resource "rtx_service_policy" "lan1_qos" {
  interface  = rtx_interface.lan1.interface_name
  direction  = "output"
  policy_map = rtx_policy_map.qos_policy.name
}

# ============================================================================
# Example 5: Traffic Shaping
# ============================================================================
# Limit outbound traffic on LAN2 interface to 10 Mbps

resource "rtx_shape" "lan2_limit" {
  interface     = rtx_interface.lan2.interface_name
  direction     = "output"
  shape_average = 10000000 # 10 Mbps in bps
}

# ============================================================================
# Example 6: Complete QoS Setup
# ============================================================================
# A complete example showing class maps, policy map, service policy, and shaping

# Define class maps for different traffic types
resource "rtx_class_map" "realtime" {
  name                   = "realtime-apps"
  match_destination_port = [5060, 5061, 16384, 16385, 16386, 16387] # SIP + RTP
}

resource "rtx_class_map" "interactive" {
  name                   = "interactive-apps"
  match_destination_port = [22, 23, 3389] # SSH, Telnet, RDP
}

resource "rtx_class_map" "streaming" {
  name                   = "streaming-apps"
  match_destination_port = [80, 443, 8080] # HTTP/HTTPS
}

# Define policy map with bandwidth allocation
resource "rtx_policy_map" "complete_qos" {
  name = "complete-qos-policy"

  class {
    name              = rtx_class_map.realtime.name
    priority          = "high"
    bandwidth_percent = 25
    queue_limit       = 32
  }

  class {
    name              = rtx_class_map.interactive.name
    priority          = "high"
    bandwidth_percent = 25
  }

  class {
    name              = rtx_class_map.streaming.name
    priority          = "normal"
    bandwidth_percent = 30
  }

  # Remaining 20% for default traffic (implicit)
}

# Apply service policy to LAN interface
resource "rtx_service_policy" "complete_policy" {
  interface  = rtx_interface.lan2.interface_name
  direction  = "output"
  policy_map = rtx_policy_map.complete_qos.name
}

# Apply traffic shaping to limit bandwidth
resource "rtx_shape" "complete_shaping" {
  interface     = rtx_interface.lan2.interface_name
  direction     = "output"
  shape_average = 50000000 # 50 Mbps

  depends_on = [rtx_service_policy.complete_policy]
}

# ============================================================================
# Outputs
# ============================================================================

output "voip_class_map" {
  description = "VoIP class map name"
  value       = rtx_class_map.voip.name
}

output "lan2_shape_rate" {
  description = "LAN2 shaping rate in bps"
  value       = rtx_shape.lan2_limit.shape_average
}
