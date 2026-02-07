# Master Requirements: QoS Resources

## Overview

This document defines the requirements for Quality of Service (QoS) resources in the Terraform Provider for Yamaha RTX routers. QoS resources enable traffic classification, prioritization, bandwidth management, and traffic shaping on network interfaces. The QoS implementation follows a hierarchical model where class-maps classify traffic, policy-maps define actions for those classes, service-policies apply policies to interfaces, and shape configurations limit overall bandwidth.

## Alignment with Product Vision

These QoS resources support the infrastructure-as-code approach for managing Yamaha RTX router configurations, enabling:
- Automated deployment of consistent QoS policies across multiple routers
- Version-controlled traffic management configurations
- Reproducible network QoS settings for disaster recovery
- Integration with CI/CD pipelines for network automation

## Resource Summary

| Attribute | Value |
|-----------|-------|
| Resources | `rtx_class_map`, `rtx_policy_map`, `rtx_service_policy`, `rtx_shape` |
| Type | Collection (multiple instances per router) |
| Import Support | Yes (all resources) |
| Last Updated | 2026-01-23 |
| Source Specs | Implementation code analysis |

---

## Resource 1: rtx_class_map

### Resource Overview

Class-maps classify network traffic based on various match criteria including protocol, ports, DSCP values, and IP filters. They serve as traffic selectors that are referenced by policy-maps.

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_class_map` |
| Type | Collection |
| Import Support | Yes |
| Import ID Format | `{class-map-name}` |

### Functional Requirements

#### Create

- MUST create a new class-map configuration on the RTX router
- MUST validate the class-map name format (starts with letter, alphanumeric/underscore/hyphen only)
- MUST validate port numbers are within range 1-65535
- MUST support at least one match criterion (protocol, ports, DSCP, or filter)
- MUST save configuration to persistent storage after creation

#### Read

- MUST retrieve class-map configuration from the router
- MUST return all configured match criteria
- MUST handle non-existent class-maps gracefully (remove from state)

#### Update

- MUST update existing class-map match criteria
- MUST preserve the class-map name (ForceNew on name change)
- MUST save configuration after update

#### Delete

- MUST remove the class-map from the router
- MUST handle already-deleted class-maps gracefully
- MUST save configuration after deletion

### Attributes Schema

| Attribute | Type | Required | ForceNew | Description | Constraints |
|-----------|------|----------|----------|-------------|-------------|
| `name` | string | Yes | Yes | Class-map identifier | Must start with letter, alphanumeric/underscore/hyphen only |
| `match_protocol` | string | No | No | Protocol to match | e.g., "sip", "http", "ftp" |
| `match_destination_port` | list(int) | No | No | Destination ports to match | Each port: 1-65535 |
| `match_source_port` | list(int) | No | No | Source ports to match | Each port: 1-65535 |
| `match_dscp` | string | No | No | DSCP value to match | e.g., "ef", "af11", numeric "46" |
| `match_filter` | int | No | No | IP filter number reference | 1-65535 |

### Acceptance Criteria

1. WHEN a valid class-map name is provided THEN the system SHALL create the class-map
2. WHEN a class-map name starts with a number THEN the system SHALL reject with validation error
3. WHEN match_destination_port contains value > 65535 THEN the system SHALL reject with validation error
4. WHEN a class-map is deleted THEN the system SHALL remove it from router configuration
5. WHEN importing an existing class-map THEN the system SHALL populate all attributes correctly

---

## Resource 2: rtx_policy_map

### Resource Overview

Policy-maps define traffic treatment actions for classified traffic. They contain one or more class references, each with associated QoS parameters like priority, bandwidth allocation, policing rates, and queue limits.

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_policy_map` |
| Type | Collection |
| Import Support | Yes |
| Import ID Format | `{policy-map-name}` |

### Functional Requirements

#### Create

- MUST create a new policy-map with defined classes
- MUST validate policy-map name format (same rules as class-map)
- MUST validate total bandwidth_percent across all classes does not exceed 100%
- MUST validate priority values are "high", "normal", or "low"
- MUST require at least one class definition
- MUST save configuration after creation

#### Read

- MUST retrieve policy-map configuration with all class definitions
- MUST return all QoS parameters for each class
- MUST handle non-existent policy-maps gracefully

#### Update

- MUST support adding, removing, or modifying classes
- MUST re-validate bandwidth_percent totals on update
- MUST preserve policy-map name (ForceNew)
- MUST save configuration after update

#### Delete

- MUST remove the policy-map from the router
- MUST handle policy-maps in use by service-policies appropriately
- MUST save configuration after deletion

### Attributes Schema

| Attribute | Type | Required | ForceNew | Description | Constraints |
|-----------|------|----------|----------|-------------|-------------|
| `name` | string | Yes | Yes | Policy-map identifier | Must start with letter, alphanumeric/underscore/hyphen only |
| `class` | list(object) | Yes | No | Class definitions | At least one class required |

#### Class Block Schema

| Attribute | Type | Required | Description | Constraints |
|-----------|------|----------|-------------|-------------|
| `name` | string | Yes | Class name (references class-map) | Non-empty string |
| `priority` | string | No | Priority level | "high", "normal", or "low" |
| `bandwidth_percent` | int | No | Bandwidth allocation percentage | 1-100, total across classes <= 100 |
| `police_cir` | int | No | Committed Information Rate (bps) | Positive integer |
| `queue_limit` | int | No | Queue depth limit | Positive integer |

### Acceptance Criteria

1. WHEN total bandwidth_percent exceeds 100% THEN the system SHALL reject with validation error
2. WHEN priority is set to invalid value THEN the system SHALL reject with validation error
3. WHEN policy-map has no classes THEN the system SHALL reject with validation error
4. WHEN class references non-existent class-map THEN Terraform plan SHOULD succeed (runtime validation)
5. WHEN importing policy-map THEN the system SHALL populate all class attributes

---

## Resource 3: rtx_service_policy

### Resource Overview

Service-policies attach policy-maps to network interfaces in a specific direction (input or output). They activate QoS treatment for traffic flowing through the interface.

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_service_policy` |
| Type | Collection |
| Import Support | Yes |
| Import ID Format | `{interface}:{direction}` |

### Functional Requirements

#### Create

- MUST attach the specified policy-map to the interface
- MUST validate direction is "input" or "output"
- MUST validate interface name is non-empty
- MUST validate policy_map is non-empty
- MUST execute `queue {interface} type {policy_map}` command
- MUST save configuration after creation

#### Read

- MUST retrieve service-policy configuration from router
- MUST parse `show config | grep "queue {interface}"` output
- MUST handle non-existent service-policies gracefully

#### Update

- MUST allow changing the policy_map (in-place update)
- MUST NOT allow changing interface or direction (ForceNew)
- MUST delete and recreate for policy_map changes
- MUST save configuration after update

#### Delete

- MUST remove the service-policy from the interface
- MUST execute `no queue {interface} type` command
- MUST handle already-deleted service-policies gracefully
- MUST save configuration after deletion

### Attributes Schema

| Attribute | Type | Required | ForceNew | Description | Constraints |
|-----------|------|----------|----------|-------------|-------------|
| `interface` | string | Yes | Yes | Interface name | e.g., "lan1", "wan1", "pp1" |
| `direction` | string | Yes | Yes | Traffic direction | "input" or "output" |
| `policy_map` | string | Yes | No | Policy-map or queue type name | e.g., "priority", "cbq", custom name |

### Acceptance Criteria

1. WHEN direction is not "input" or "output" THEN the system SHALL reject with validation error
2. WHEN interface is empty THEN the system SHALL reject with validation error
3. WHEN service-policy is applied THEN router SHALL have `queue {interface} type {policy_map}` configured
4. WHEN service-policy is deleted THEN router SHALL NOT have queue type for interface
5. WHEN importing with format "{interface}:{direction}" THEN the system SHALL parse correctly

---

## Resource 4: rtx_shape

### Resource Overview

Shape configurations limit the overall bandwidth on an interface using traffic shaping. This controls the maximum rate at which traffic can egress or ingress the interface.

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_shape` |
| Type | Collection |
| Import Support | Yes |
| Import ID Format | `{interface}:{direction}` |

### Functional Requirements

#### Create

- MUST configure traffic shaping on the specified interface
- MUST validate shape_average is positive (>= 1)
- MUST validate direction is "input" or "output"
- MUST execute `speed {interface} {bandwidth}` command
- MUST save configuration after creation

#### Read

- MUST retrieve shape configuration from router
- MUST parse `speed {interface}` from configuration
- MUST handle non-existent shape configurations gracefully

#### Update

- MUST allow changing shape_average and shape_burst (in-place)
- MUST NOT allow changing interface or direction (ForceNew)
- MUST execute speed command with new values
- MUST save configuration after update

#### Delete

- MUST remove traffic shaping from the interface
- MUST execute `no speed {interface}` command
- MUST handle already-deleted shapes gracefully
- MUST save configuration after deletion

### Attributes Schema

| Attribute | Type | Required | ForceNew | Description | Constraints |
|-----------|------|----------|----------|-------------|-------------|
| `interface` | string | Yes | Yes | Interface name | e.g., "lan1", "wan1" |
| `direction` | string | Yes | Yes | Traffic direction | "input" or "output" |
| `shape_average` | int | Yes | No | Average rate limit (bps) | >= 1 |
| `shape_burst` | int | No | No | Burst size (bytes) | >= 0 |

### Acceptance Criteria

1. WHEN shape_average is 0 or negative THEN the system SHALL reject with validation error
2. WHEN direction is invalid THEN the system SHALL reject with validation error
3. WHEN shape is applied THEN router SHALL have `speed {interface} {bandwidth}` configured
4. WHEN shape is deleted THEN router SHALL NOT have speed command for interface
5. WHEN importing with format "{interface}:{direction}" THEN the system SHALL parse correctly

---

## Non-Functional Requirements

### Code Architecture and Modularity

- **Single Responsibility Principle**: Each resource file handles one QoS resource type
- **Modular Design**: Parser, service, and provider layers are isolated
- **Dependency Management**: QoS service depends on executor and parser modules
- **Clear Interfaces**: Client interface defines QoS method contracts

### Performance

- All QoS operations SHOULD complete within 30 seconds under normal network conditions
- Configuration parsing SHOULD handle large configurations (100+ queue rules)
- List operations SHOULD efficiently scan configuration without loading entire config into memory

### Security

- No sensitive data is stored in QoS configurations
- All commands executed via authenticated SSH session
- Configuration changes require administrator privileges on router

### Reliability

- Operations MUST be idempotent (repeated applies produce same result)
- Failed operations MUST NOT leave router in inconsistent state
- Configuration MUST be saved to persistent storage after each change
- "Not found" errors during delete SHOULD be handled gracefully

### Validation

- Name validation: Must start with letter, alphanumeric/underscore/hyphen only
- Port validation: 1-65535 range
- Bandwidth validation: 1-100% per class, total <= 100%
- Direction validation: "input" or "output" only
- Rate validation: Positive integers for bandwidth rates

---

## RTX Commands Reference

### Class Map Commands

```
# Class maps use IP filters and queue class configurations
ip filter <filter_num> <action> <source> <dest> <protocol> [options]
queue <interface> class filter <class_num> <filter_num>
```

### Policy Map Commands

```
# Queue type and class configuration
queue <interface> type <priority|cbq|fifo|shaping>
queue <interface> class priority <class_num> <high|normal|low>
queue <interface> class bandwidth <class_num> <percent>
queue <interface> length <class_num> <length>
```

### Service Policy Commands

```
# Apply queue type to interface
queue <interface> type <queue_type>

# Remove queue type
no queue <interface> type
```

### Shape Commands

```
# Set interface speed limit
speed <interface> <bandwidth_bps>

# Remove speed limit
no speed <interface>
```

### Show Commands

```
# Show QoS configuration for specific interface
show config | grep "queue <interface>\|speed <interface>"

# Show all QoS configuration
show config | grep "queue\|speed"
```

---

## Terraform Command Support

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | Required | Preview QoS configuration changes |
| `terraform apply` | Required | Apply QoS configuration to router |
| `terraform destroy` | Required | Remove QoS configurations |
| `terraform import` | Required | Import existing QoS configurations |
| `terraform refresh` | Required | Sync state with current router configuration |
| `terraform state` | Required | Manage Terraform state |

### Import Specifications

#### rtx_class_map

- **Import ID Format**: `{name}` (e.g., `voip-traffic`)
- **Import Command**: `terraform import rtx_class_map.example voip-traffic`
- **Post-Import**: Verify all match criteria are populated

#### rtx_policy_map

- **Import ID Format**: `{name}` (e.g., `qos-policy`)
- **Import Command**: `terraform import rtx_policy_map.example qos-policy`
- **Post-Import**: Verify all class definitions and parameters are populated

#### rtx_service_policy

- **Import ID Format**: `{interface}:{direction}` (e.g., `lan1:output`)
- **Import Command**: `terraform import rtx_service_policy.example lan1:output`
- **Post-Import**: Verify interface, direction, and policy_map are populated

#### rtx_shape

- **Import ID Format**: `{interface}:{direction}` (e.g., `lan1:output`)
- **Import Command**: `terraform import rtx_shape.example lan1:output`
- **Post-Import**: Verify interface, direction, and shape_average are populated

---

## Example Usage

### Complete QoS Configuration

```hcl
# Define traffic classification
resource "rtx_class_map" "voip" {
  name                   = "voip-traffic"
  match_protocol         = "sip"
  match_destination_port = [5060, 5061]
}

resource "rtx_class_map" "video" {
  name       = "video-traffic"
  match_dscp = "af41"
}

resource "rtx_class_map" "bulk_data" {
  name         = "bulk-data"
  match_filter = 100
}

# Define QoS policy
resource "rtx_policy_map" "corporate_qos" {
  name = "corporate-qos"

  class {
    name              = "voip-traffic"
    priority          = "high"
    bandwidth_percent = 20
    queue_limit       = 64
  }

  class {
    name              = "video-traffic"
    priority          = "normal"
    bandwidth_percent = 30
  }

  class {
    name              = "bulk-data"
    priority          = "low"
    bandwidth_percent = 50
  }
}

# Apply policy to interface
resource "rtx_service_policy" "lan1_output" {
  interface  = "lan1"
  direction  = "output"
  policy_map = "priority"
}

# Configure traffic shaping
resource "rtx_shape" "wan1_output" {
  interface     = "wan1"
  direction     = "output"
  shape_average = 100000000  # 100 Mbps
  shape_burst   = 15000      # 15KB burst
}
```

### Basic Priority Queue

```hcl
resource "rtx_service_policy" "simple_priority" {
  interface  = "lan1"
  direction  = "output"
  policy_map = "priority"
}
```

### CBQ (Class-Based Queuing)

```hcl
resource "rtx_service_policy" "cbq_policy" {
  interface  = "lan2"
  direction  = "output"
  policy_map = "cbq"
}
```

---

## State Handling

- Only configuration attributes are persisted in Terraform state
- Operational/runtime statistics (queue depths, drop counts) are NOT stored in state
- Resource IDs are derived from configuration attributes:
  - class_map: `{name}`
  - policy_map: `{name}`
  - service_policy: `{interface}:{direction}`
  - shape: `{interface}:{direction}`

---

## Change History

| Date | Source | Changes |
|------|--------|---------|
| 2026-01-23 | Implementation analysis | Initial master spec creation from existing code |
| 2026-02-07 | Implementation Audit | Full audit against implementation code |
