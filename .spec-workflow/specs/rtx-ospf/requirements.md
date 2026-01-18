# Requirements: rtx_ospf

## Overview
Terraform resource for managing OSPF (Open Shortest Path First) routing configuration on Yamaha RTX routers.

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure OSPF routing process
- **Read**: Query OSPF configuration and neighbor status
- **Update**: Modify OSPF parameters
- **Delete**: Remove OSPF configuration

### 2. OSPF Process Configuration
- Enable/disable OSPF
- Router ID
- Reference bandwidth
- SPF calculation timers

### 3. Area Configuration
- Area ID (0.0.0.0 for backbone)
- Area type (normal, stub, NSSA, totally stubby)
- Virtual links for non-contiguous areas
- Area range summarization

### 4. Interface Configuration
- Network assignment to areas
- Interface cost
- Hello and dead intervals
- Priority for DR/BDR election
- Authentication (none, simple, MD5)

### 5. Route Redistribution
- Redistribute static routes
- Redistribute connected routes
- Redistribute from other protocols
- Metric and metric-type for redistributed routes

### 6. Passive Interface
- Mark interfaces as passive (no hello)
- Useful for stub networks

### 7. Import Support
- Import existing OSPF configuration

## Non-Functional Requirements

### 8. Validation
- Validate router ID format
- Validate area ID format
- Validate timer ranges
- Validate cost values

### 9. Dependencies
- Require interface IP configuration

## RTX Commands Reference
```
ospf use on
ospf router id <router_id>
ospf area <area> <interface>
ospf import from static
ospf import from ospf
ip <interface> ospf area <area>
ip <interface> ospf cost <cost>
ip <interface> ospf priority <priority>
```

## Example Usage
```hcl
resource "rtx_ospf" "main" {
  enabled   = true
  router_id = "1.1.1.1"

  area {
    id   = "0.0.0.0"
    type = "normal"
  }

  area {
    id   = "0.0.0.1"
    type = "stub"
    no_summary = false
  }

  network {
    prefix = "192.168.1.0/24"
    area   = "0.0.0.0"
  }

  network {
    prefix = "10.0.0.0/24"
    area   = "0.0.0.1"
  }

  redistribute {
    protocol = "static"
    metric   = 100
    metric_type = 2
  }
}
```
