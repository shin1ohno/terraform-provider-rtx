# Master Requirements: System Resources

## Overview

This specification covers system-level resources and data sources for managing Yamaha RTX router system configuration and retrieving system status information. The system resource group provides core functionality for router identity, timing, console settings, and operational status monitoring.

## Alignment with Product Vision

System resources form the foundation of RTX router management through Terraform:
- **Infrastructure as Code**: Enable version-controlled system configuration
- **Visibility**: Provide read-only access to system status, interfaces, routes, and DDNS status
- **Compliance**: Consistent timezone and console settings across router fleet

## Resource Summary

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_system` (resource), `rtx_system_info`, `rtx_interfaces`, `rtx_routes`, `rtx_ddns_status` (data sources) |
| Type | singleton (resource), read-only (data sources) |
| Import Support | yes (resource) |
| Last Updated | 2026-01-23 |
| Source Specs | Initial implementation |

---

## Resource: rtx_system

### Description

Manages system-level settings on RTX routers. This is a singleton resource - there is only one system configuration per router.

### Functional Requirements

#### Core Operations

##### Create
Creates system configuration by applying settings to the router via CLI commands. Uses fixed ID "system" for the singleton pattern.

##### Read
Retrieves current system configuration by parsing `show config` output filtered for system-related settings (timezone, console, packet-buffer, statistics).

##### Update
Compares desired state with current state and applies only changed settings. Supports incremental updates to minimize command execution.

##### Delete
Resets system configuration to defaults by issuing `no` commands for each configured setting.

### Terraform Schema

```hcl
resource "rtx_system" "example" {
  timezone = "+09:00"  # UTC offset (e.g., '+09:00' for JST, '-05:00' for EST)

  console {
    character = "ja.utf8"  # ja.utf8, ja.sjis, ascii, euc-jp
    lines     = "infinity" # positive integer or 'infinity'
    prompt    = "RTX>"     # custom prompt string
  }

  packet_buffer {
    size       = "small"
    max_buffer = 5000
    max_free   = 1300
  }

  packet_buffer {
    size       = "middle"
    max_buffer = 3000
    max_free   = 800
  }

  statistics {
    traffic = true
    nat     = true
  }
}
```

### Attribute Reference

| Attribute | Type | Required | Description | Validation |
|-----------|------|----------|-------------|------------|
| `timezone` | string | Optional | Timezone as UTC offset | Pattern: `^[+-]\d{2}:\d{2}$` (e.g., '+09:00', '-05:00') |
| `console` | block | Optional | Console settings (max 1 block) | - |
| `console.character` | string | Optional | Character encoding | Enum: `ja.utf8`, `ja.sjis`, `ascii`, `euc-jp` |
| `console.lines` | string | Optional | Lines per page | Positive integer or `infinity` |
| `console.prompt` | string | Optional | Custom prompt string | - |
| `packet_buffer` | block | Optional | Packet buffer tuning (max 3 blocks) | - |
| `packet_buffer.size` | string | Required | Buffer size category | Enum: `small`, `middle`, `large` |
| `packet_buffer.max_buffer` | int | Required | Maximum buffer count | Minimum: 1 |
| `packet_buffer.max_free` | int | Required | Maximum free buffer count | Minimum: 1, must be <= max_buffer |
| `statistics` | block | Optional | Statistics collection settings (max 1 block) | - |
| `statistics.traffic` | bool | Optional | Enable traffic statistics collection | Computed if not set |
| `statistics.nat` | bool | Optional | Enable NAT statistics collection | Computed if not set |

### RTX Commands Reference

```
# Timezone
timezone +09:00
no timezone

# Console settings
console character ja.utf8
console lines 24
console lines infinity
console prompt "RTX>"
no console character
no console lines
no console prompt

# Packet buffer tuning
system packet-buffer small max-buffer=5000 max-free=1300
system packet-buffer middle max-buffer=3000 max-free=800
system packet-buffer large max-buffer=2000 max-free=500
no system packet-buffer small

# Statistics
statistics traffic on
statistics traffic off
statistics nat on
statistics nat off
no statistics traffic
no statistics nat

# Show configuration
show config | grep -E "(timezone|console|packet-buffer|statistics)"
```

### Import Specification

- **Import ID Format**: `system`
- **Import Command**: `terraform import rtx_system.main system`
- **Post-Import**: All readable attributes are populated from router

---

## Data Source: rtx_system_info

### Description

Retrieves system information from an RTX router including model, firmware version, serial number, MAC address, and uptime.

### Terraform Schema

```hcl
data "rtx_system_info" "router" {}

output "router_model" {
  value = data.rtx_system_info.router.model
}
```

### Attribute Reference

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | string | Internal identifier (MD5 hash of system attributes) |
| `model` | string | RTX router model number (e.g., RTX1210, RTX830) |
| `firmware_version` | string | Firmware version running on the router |
| `serial_number` | string | Serial number of the router |
| `mac_address` | string | MAC address of the router |
| `uptime` | string | Uptime of the router |

### RTX Commands Reference

```
show environment
```

---

## Data Source: rtx_interfaces

### Description

Retrieves network interface information from an RTX router including LAN, WAN, PP, and VLAN interfaces.

### Terraform Schema

```hcl
data "rtx_interfaces" "all" {}

output "lan1_ip" {
  value = [for iface in data.rtx_interfaces.all.interfaces : iface.ipv4 if iface.name == "LAN1"][0]
}
```

### Attribute Reference

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | string | Internal identifier (MD5 hash of interface data) |
| `interfaces` | list | List of network interfaces |
| `interfaces.name` | string | Interface name (e.g., LAN1, WAN1, PP1, VLAN1) |
| `interfaces.kind` | string | Interface type: `lan`, `wan`, `pp`, or `vlan` |
| `interfaces.admin_up` | bool | Whether interface is administratively up |
| `interfaces.link_up` | bool | Whether physical link is up |
| `interfaces.mac` | string | MAC address of the interface |
| `interfaces.ipv4` | string | IPv4 address assigned to the interface |
| `interfaces.ipv6` | string | IPv6 address assigned to the interface |
| `interfaces.mtu` | int | Maximum Transmission Unit (MTU) |
| `interfaces.description` | string | Interface description |
| `interfaces.attributes` | map(string) | Additional model-specific attributes |

### RTX Commands Reference

```
show status lan1
show status wan1
show status pp 1
```

---

## Data Source: rtx_routes

### Description

Retrieves routing table information from an RTX router including static routes, connected routes, and dynamic routing protocol entries.

### Terraform Schema

```hcl
data "rtx_routes" "all" {}

output "default_gateway" {
  value = [for route in data.rtx_routes.all.routes : route.gateway if route.destination == "0.0.0.0/0"][0]
}
```

### Attribute Reference

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | string | Internal identifier (MD5 hash of route data) |
| `routes` | list | List of routes in the routing table |
| `routes.destination` | string | Destination network prefix (e.g., '192.168.1.0/24', '0.0.0.0/0') |
| `routes.gateway` | string | Next hop gateway IP ('*' for directly connected) |
| `routes.interface` | string | Outgoing interface |
| `routes.protocol` | string | Route protocol code |
| `routes.metric` | int | Route metric (cost), 0 if not specified |

### Protocol Codes

| Code | Protocol |
|------|----------|
| S | Static route |
| C | Connected (directly attached network) |
| R | RIP |
| O | OSPF |
| B | BGP |
| D | DHCP |

### RTX Commands Reference

```
show ip route
```

---

## Data Source: rtx_ddns_status

### Description

Retrieves DDNS registration status from RTX routers including both NetVolante DNS (Yamaha's free DDNS service) and custom DDNS providers.

### Terraform Schema

```hcl
# Get all DDNS status
data "rtx_ddns_status" "all" {
  type = "all"
}

# Get only NetVolante status
data "rtx_ddns_status" "netvolante" {
  type = "netvolante"
}

# Get only custom DDNS status
data "rtx_ddns_status" "custom" {
  type = "custom"
}
```

### Attribute Reference

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `type` | string | Optional | Filter type: `netvolante`, `custom`, or `all` (default) |
| `statuses` | list | Computed | List of DDNS status entries |
| `statuses.type` | string | - | DDNS type: `netvolante` or `custom` |
| `statuses.interface` | string | - | Interface (for NetVolante DNS) |
| `statuses.server_id` | int | - | Server ID (for custom DDNS, 0 for NetVolante) |
| `statuses.hostname` | string | - | Registered hostname |
| `statuses.current_ip` | string | - | Currently registered IP address |
| `statuses.last_update` | string | - | Last successful update timestamp |
| `statuses.status` | string | - | Status: `success`, `error`, or `pending` |
| `statuses.error_message` | string | - | Error message if status is `error` |

### RTX Commands Reference

```
show status netvolante-dns
show status ddns
```

---

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Each component handles one concern (provider, service, parser)
- **Modular Design**: Parsers are registered per model for extensibility
- **Dependency Management**: Clear interfaces between layers
- **Clear Interfaces**: Client interface abstracts implementation details

### Performance
- Batch command execution to minimize SSH round-trips
- Configuration save after batch operations (not per-command)
- MD5-based ID generation for data sources (efficient comparison)

### Security
- No secrets stored in Terraform state for system resources
- All communication via authenticated SSH session
- Administrator mode required for configuration changes

### Reliability
- Context cancellation support throughout
- Graceful handling of command failures during reset operations
- Read-after-write pattern ensures state consistency

### Validation
- Timezone format validated as `±HH:MM`
- Console character encoding validated against allowed values
- Packet buffer constraints: max_free <= max_buffer
- Console lines: positive integer or 'infinity'

## Terraform Command Support

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | Required | Shows changes to system configuration |
| `terraform apply` | Required | Applies system configuration changes |
| `terraform destroy` | Required | Resets system configuration to defaults |
| `terraform import` | Required | Imports existing system configuration |
| `terraform refresh` | Required | Refreshes state from router |
| `terraform state` | Required | Manages local state |

## State Handling

- Only configuration attributes are persisted in Terraform state
- Runtime/operational status (uptime, etc.) are only in data sources
- Data sources generate dynamic IDs based on content hash
- Resource uses fixed ID "system" for singleton pattern

## Change History

| Date | Source Spec | Changes |
|------|-------------|---------|
| 2026-01-23 | Initial | Created master spec from implementation |
