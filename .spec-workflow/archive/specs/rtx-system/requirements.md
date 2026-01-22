# Requirements: rtx_system

## Overview
Terraform resource for managing system-level settings on Yamaha RTX routers. This includes timezone, console settings, packet buffer tuning, and statistics collection.

## Functional Requirements

### 1. CRUD Operations
- **Create**: Configure system settings
- **Read**: Query current system configuration
- **Update**: Modify system parameters
- **Delete**: Reset to default system settings

### 2. Timezone Configuration
- Configure timezone as UTC offset (e.g., +09:00 for JST)
- Support standard timezone format (±HH:MM)

### 3. Console Settings
- Character encoding (ja.utf8, ascii, ja.sjis)
- Lines per page (number or "infinity")
- Custom prompt string

### 4. Packet Buffer Tuning
- Configure small packet buffers (max-buffer, max-free)
- Configure middle packet buffers
- Configure large packet buffers
- High-traffic environment optimization

### 5. Statistics Collection
- Enable/disable traffic statistics
- Enable/disable NAT statistics

### 6. Import Support
- Import existing system configuration

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned system configuration changes |
| `terraform apply` | ✅ Required | Create, update, or delete system settings |
| `terraform destroy` | ✅ Required | Reset system to defaults |
| `terraform import` | ✅ Required | Import existing system settings |
| `terraform refresh` | ✅ Required | Sync state with actual configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `system` (singleton resource)
- **Import Command**: `terraform import rtx_system.main system`
- **Post-Import**: All attributes must be populated from router

## Non-Functional Requirements

### 7. Validation
- Validate timezone format (±HH:MM)
- Validate character encoding is supported
- Validate packet buffer values are positive integers
- Validate buffer size names (small, middle, large)

### 8. Performance
- Packet buffer changes may require restart for full effect

## RTX Commands Reference
```
timezone <utc_offset>
console character <encoding>
console lines <number|infinity>
console prompt "<prompt>"
system packet-buffer <size> max-buffer=<n> max-free=<n>
statistics traffic on|off
statistics nat on|off
```

## Example Usage
```hcl
resource "rtx_system" "main" {
  # Timezone (Japan Standard Time)
  timezone = "+09:00"

  # Console settings
  console {
    character = "ja.utf8"
    lines     = "infinity"
    prompt    = "[RTX1210] "
  }

  # Packet buffer tuning (high-traffic environment)
  packet_buffer {
    size       = "small"
    max_buffer = 5000
    max_free   = 1300
  }

  packet_buffer {
    size       = "middle"
    max_buffer = 10000
    max_free   = 4950
  }

  packet_buffer {
    size       = "large"
    max_buffer = 20000
    max_free   = 5600
  }

  # Statistics collection
  statistics {
    traffic = true
    nat     = true
  }
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state.
- Operational/runtime status must not be stored in state.
