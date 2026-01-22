# Requirements: rtx_schedule

## Overview
Terraform resources for managing scheduled tasks and time-based automation on Yamaha RTX routers.

**Cisco Equivalent**: `iosxe_kron_schedule`, `iosxe_kron_policy`

## Cisco Compatibility

These resources follow Cisco Kron scheduler naming patterns:

| RTX Attribute | Cisco Equivalent | Notes |
|---------------|------------------|-------|
| `name` | `name` | Schedule name |
| `time` | `at_time` | Execution time |
| `recurring` | `recurring` | Recurring schedule |
| `command` | `command_line` | Command to execute |
| `policy_list` | `policy_list` | Associated policy |

## Covered Resources

This specification covers two Terraform resources:

- **`rtx_kron_policy`**: Defines command lists to execute
- **`rtx_kron_schedule`**: Defines when to execute policies

## Functional Requirements

### 1. CRUD Operations
- **Create**: Define policies and schedules
- **Read**: Query policy and schedule configuration
- **Update**: Modify policy commands and schedule parameters
- **Delete**: Remove policies and schedules

### 2. Schedule Definition
- Schedule ID/name
- Execution time (cron-like syntax)
- Command to execute
- Enable/disable individual schedules

### 3. Time Specification
- One-time execution (specific date/time)
- Recurring execution (daily, weekly, monthly)
- Interval-based execution (every N minutes/hours)
- Day of week specification
- Time range (start and end time)

### 4. Command Execution
- Single command
- Command sequence
- Configuration commands
- Show commands (for logging)

### 5. Conditional Execution
- Execute only if condition is met
- Interface status conditions
- Connection status conditions
  - Runtime status is operational-only and MUST NOT be persisted in Terraform state

### 6. Schedule Groups
- Group related schedules
- Enable/disable groups

### 7. Import Support
- Import existing schedules

## Terraform Command Support

This resource must fully support all standard Terraform workflow commands:

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | ✅ Required | Show planned schedule changes |
| `terraform apply` | ✅ Required | Create, update, or delete scheduled tasks |
| `terraform destroy` | ✅ Required | Remove scheduled tasks from router |
| `terraform import` | ✅ Required | Import existing schedules into state |
| `terraform refresh` | ✅ Required | Sync state with actual schedule configuration |
| `terraform state` | ✅ Required | Support state inspection and manipulation |

### Import Specification
- **Import ID Format**: `<schedule_id>` (e.g., `1`)
- **Import Command**: `terraform import rtx_kron_schedule.daily_backup 1`
- **Post-Import**: All schedule parameters must be populated from router

## Non-Functional Requirements

### 8. Validation
- Validate time format
- Validate command syntax
- Validate schedule ID uniqueness

### 9. Timezone
- Use router's configured timezone
- Handle daylight saving time

## RTX Commands Reference
```
schedule at <id> <time> <command>
schedule at <id> startup <command>
schedule at <id> <date> <time> <command>
schedule pp 1 <day> <time> connect
schedule pp 1 <day> <time> disconnect
```

## Example Usage
```hcl
# Kron policy (command list) - Cisco-compatible naming
resource "rtx_kron_policy" "backup_commands" {
  name = "BACKUP_POLICY"

  command_lines = [
    "copy config sd1:backup/config-$(date +%Y%m%d).txt"
  ]
}

# Kron schedule - Cisco-compatible naming
resource "rtx_kron_schedule" "daily_backup" {
  name = "DAILY_BACKUP"

  at_time   = "02:00"
  recurring = true

  policy_list = "BACKUP_POLICY"
}

resource "rtx_kron_schedule" "weekly_reboot" {
  name = "WEEKLY_REBOOT"

  at_time     = "04:00"
  day_of_week = "sunday"
  recurring   = true

  policy_list = "REBOOT_POLICY"
}

resource "rtx_kron_policy" "reboot_commands" {
  name = "REBOOT_POLICY"

  command_lines = ["restart"]
}

# Startup schedule
resource "rtx_kron_schedule" "startup_task" {
  name = "STARTUP_LOG"

  on_startup = true

  policy_list = "STARTUP_POLICY"
}

resource "rtx_kron_policy" "startup_commands" {
  name = "STARTUP_POLICY"

  command_lines = [
    "syslog info 'Router started successfully'"
  ]
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state.
- Operational/runtime status must not be stored in state.
