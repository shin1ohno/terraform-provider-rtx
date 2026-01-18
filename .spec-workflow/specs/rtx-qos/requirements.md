# Requirements: rtx_qos

## Overview
Terraform resource for managing QoS (Quality of Service) and traffic shaping on Yamaha RTX routers.

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
resource "rtx_qos" "wan_qos" {
  interface = "pp1"
  bandwidth = "100m"

  queue_type = "priority"

  class {
    id       = 1
    name     = "voip"
    priority = "high"

    match {
      protocol  = "udp"
      dest_port = "5060,10000-20000"
    }
  }

  class {
    id       = 2
    name     = "web"
    priority = "medium"

    match {
      protocol  = "tcp"
      dest_port = "80,443"
    }
  }

  class {
    id       = 3
    name     = "default"
    priority = "normal"

    match {
      source = "*"
    }
  }
}

resource "rtx_qos_shaper" "upload_limit" {
  interface = "lan1"
  direction = "out"

  bandwidth = "50m"
  burst     = "1m"
}
```
