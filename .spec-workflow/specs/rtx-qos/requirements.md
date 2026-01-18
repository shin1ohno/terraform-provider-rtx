# Requirements: rtx_qos

## Overview
Terraform resource for managing QoS (Quality of Service) and traffic shaping on Yamaha RTX routers.

**Cisco Equivalent**: `iosxe_policy_map`, `iosxe_class_map`, `iosxe_service_policy`

## Cisco Compatibility

This resource follows Cisco MQC (Modular QoS CLI) naming patterns:

| RTX Attribute | Cisco Equivalent | Notes |
|---------------|------------------|-------|
| `name` | `name` | Policy/class name |
| `class` | `class` | Traffic class definition |
| `match` | `match_*` | Match conditions |
| `bandwidth` | `bandwidth_percent` | Bandwidth allocation |
| `priority` | `priority` | Priority queuing |
| `police` | `police_cir` | Traffic policing |
| `shape` | `shape_average` | Traffic shaping |

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure QoS policies
- **Read**: Query QoS configuration
- **Update**: Modify QoS parameters
- **Delete**: Remove QoS configuration

### 2. Traffic Classification
- Classify by source/destination IP
- Classify by protocol and port
- Classify by DSCP/TOS
- Classify by interface

### 3. Priority Queuing
- Define priority levels (high, medium, normal, low)
- Map traffic classes to priorities
- Strict priority queuing
- Weighted fair queuing

### 4. Bandwidth Control
- Interface bandwidth limit
- Per-class bandwidth allocation
- Minimum guaranteed bandwidth
- Maximum burst bandwidth

### 5. Traffic Shaping
- Rate limiting (policing)
- Traffic shaping (smoothing)
- Burst size configuration

### 6. Queue Configuration
- Queue depth
- Drop policy (tail drop, WRED)
- ECN support

### 7. Import Support
- Import existing QoS configuration

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned QoS configuration changes |
| `terraform apply` | ✅ Required | Create, update, or delete QoS policies |
| `terraform destroy` | ✅ Required | Remove QoS configuration from interface |
| `terraform import` | ✅ Required | Import existing QoS configuration into state |
| `terraform refresh` | ✅ Required | Sync state with actual QoS configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `<interface>` (e.g., `pp1`)
- **Import Command**: `terraform import rtx_qos.wan_qos pp1`
- **Post-Import**: All classes and queue settings must be populated

## Non-Functional Requirements

### 8. Validation
- Validate bandwidth values
- Validate priority levels
- Validate class definitions

### 9. Performance
- Minimize latency for high-priority traffic

## RTX Commands Reference
```
queue <interface> type priority
queue <interface> class filter <n> <filter>
queue <interface> class priority <class> <priority>
speed <interface> <bandwidth>
queue <interface> length <class> <length>
```

## Example Usage
```hcl
# Class map - Cisco MQC style
resource "rtx_class_map" "voip" {
  name = "VOIP"

  match_protocol = "udp"
  match_destination_port = ["5060", "10000-20000"]
}

resource "rtx_class_map" "web" {
  name = "WEB"

  match_protocol = "tcp"
  match_destination_port = ["80", "443"]
}

# Policy map - Cisco MQC style
resource "rtx_policy_map" "wan_qos" {
  name = "WAN_QOS"

  class {
    name     = "VOIP"
    priority = true
    police_cir = 1000000  # 1 Mbps
  }

  class {
    name             = "WEB"
    bandwidth_percent = 30
  }

  class {
    name             = "class-default"
    bandwidth_percent = 20
    queue_limit      = 64
  }
}

# Apply policy to interface
resource "rtx_service_policy" "wan" {
  interface  = "pp1"
  direction  = "output"
  policy_map = "WAN_QOS"
}

# Traffic shaping
resource "rtx_shape" "upload_limit" {
  interface     = "lan1"
  direction     = "output"
  shape_average = 50000000  # 50 Mbps
  shape_burst   = 1000000   # 1 Mbps burst
}
```
