# Requirements: rtx_ospf

## Overview
Terraform resource for managing OSPF (Open Shortest Path First) routing configuration on Yamaha RTX routers.

**Cisco Equivalent**: `iosxe_ospf`

## Cisco Compatibility

This resource follows Cisco IOS XE Terraform provider naming conventions:

| RTX Attribute | Cisco Equivalent | Notes |
|---------------|------------------|-------|
| `process_id` | `process_id` | OSPF process identifier |
| `router_id` | `router_id` | Router ID (dotted decimal) |
| `networks` | `networks` | List of network/area mappings |
| `neighbors` | `neighbors` | Static neighbor list |
| `distance` | `distance` | Administrative distance |
| `default_information_originate` | `default_information_originate` | Advertise default route |

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure OSPF routing process
- **Read**: Query OSPF configuration and neighbor status
  - Neighbor status is operational-only and MUST NOT be persisted in Terraform state
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

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned OSPF configuration changes |
| `terraform apply` | ✅ Required | Create, update, or delete OSPF settings |
| `terraform destroy` | ✅ Required | Disable OSPF and remove all area/network configs |
| `terraform import` | ✅ Required | Import existing OSPF configuration into state |
| `terraform refresh` | ✅ Required | Sync state with actual OSPF configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `ospf` (singleton resource)
- **Import Command**: `terraform import rtx_ospf.main ospf`
- **Post-Import**: All areas, networks, and redistribution settings populated

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
# OSPF configuration - Cisco-compatible naming
resource "rtx_ospf" "backbone" {
  process_id = 1
  router_id  = "1.1.1.1"
  distance   = 110

  default_information_originate = true

  networks = [
    {
      ip       = "192.168.1.0"
      wildcard = "0.0.0.255"
      area     = "0"
    },
    {
      ip       = "10.0.0.0"
      wildcard = "0.0.0.255"
      area     = "1"
    }
  ]

  neighbors = [
    {
      ip       = "192.168.1.2"
      priority = 10
      cost     = 100
    }
  ]

  # RTX-specific: area type configuration
  areas = [
    {
      id   = "0"
      type = "normal"
    },
    {
      id         = "1"
      type       = "stub"
      no_summary = false
    }
  ]

  redistribute_static = true
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state.
- Operational/runtime status must not be stored in state.
