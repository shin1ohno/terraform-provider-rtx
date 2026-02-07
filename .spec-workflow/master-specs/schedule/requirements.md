# Master Requirements: Schedule Resources

## Overview

The Schedule resources manage scheduled task execution on Yamaha RTX routers. RTX routers support the `schedule at` command for time-based, startup-based, and date-specific task automation. This includes two Terraform resources:

1. **rtx_kron_schedule** - Manages scheduled tasks that execute commands at specific times, on startup, or on specific dates
2. **rtx_kron_policy** - Manages named command lists (Terraform-level abstraction, as RTX doesn't have native kron policy support)

## Alignment with Product Vision

These resources support the product vision by enabling:
- Automated maintenance tasks (backups, restarts, config saves)
- Scheduled connectivity management (PP interface connect/disconnect)
- Time-based configuration changes
- Startup initialization sequences
- Infrastructure-as-Code approach to router automation

## Resource Summary

### rtx_kron_schedule

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_kron_schedule` |
| Type | collection (identified by schedule_id) |
| Import Support | yes |
| Last Updated | 2026-01-23 |
| Source | Implementation code analysis |

### rtx_kron_policy

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_kron_policy` |
| Type | collection (identified by name) |
| Import Support | yes (partial - commands cannot be imported) |
| Last Updated | 2026-01-23 |
| Source | Implementation code analysis |

## Functional Requirements

### rtx_kron_schedule

#### Core Operations

##### Create
- Validates schedule configuration before sending to router
- Checks if schedule with same ID already exists (returns error if exists)
- Builds appropriate RTX command based on schedule type (time/startup/date)
- Executes schedule command(s) on router
- Saves configuration to persistent memory
- Sets Terraform resource ID to schedule_id

##### Read
- Retrieves schedule configuration from router using `show config | grep "schedule at <id>"`
- Parses schedule output to extract time, date, startup flag, and commands
- Updates Terraform state with current values
- Removes resource from state if schedule not found on router

##### Update
- Deletes existing schedule commands first
- Creates new schedule with updated values
- Saves configuration to persistent memory
- This is a delete-and-recreate operation at the RTX level

##### Delete
- Executes `no schedule at <id>` command
- Saves configuration to persistent memory
- Handles "not found" gracefully (idempotent)

#### Attributes

| Attribute | Type | Required | ForceNew | Description |
|-----------|------|----------|----------|-------------|
| `schedule_id` | int | yes | yes | Schedule ID (1-65535), must be unique |
| `name` | string | no | no | Optional description/name for the schedule |
| `at_time` | string | no* | no | Execution time in HH:MM format (24-hour) |
| `day_of_week` | string | no | no | Day(s) of week: `mon`, `mon-fri`, `sat,sun` |
| `date` | string | no* | no | Specific date in YYYY/MM/DD format |
| `recurring` | bool | no | no | Whether schedule repeats (auto-set false for date-based) |
| `on_startup` | bool | no* | no | Execute when router starts |
| `policy_list` | string | no** | no | Name of kron policy to execute |
| `command_lines` | list(string) | no** | no | Commands to execute |

*One of `at_time`, `on_startup`, or `date` must be specified
**One of `policy_list` or `command_lines` must be specified

#### Attribute Constraints

- `at_time` conflicts with: `on_startup`
- `date` conflicts with: `on_startup`, `day_of_week`
- `on_startup` conflicts with: `at_time`, `date`, `day_of_week`
- `policy_list` conflicts with: `command_lines`

#### Validation Rules

| Field | Rule |
|-------|------|
| `schedule_id` | Integer between 1 and 65535 |
| `at_time` | HH:MM format, hour 0-23, minute 0-59 |
| `date` | YYYY/MM/DD format, year 2000-2099, month 1-12, day 1-31 |
| `day_of_week` | Valid days: sun, mon, tue, wed, thu, fri, sat; supports ranges (mon-fri) and lists (sat,sun) |

### rtx_kron_policy

#### Core Operations

##### Create
- Validates policy name format
- Validates commands list is not empty
- Stores policy as logical grouping (RTX doesn't have native kron policy)
- Sets Terraform resource ID to policy name

##### Read
- Policy is managed at Terraform level only
- Reads back values from state (no device query)
- Validates state consistency

##### Update
- Validates updated policy configuration
- Updates Terraform state

##### Delete
- Removes policy from Terraform state
- No router command needed (policy is Terraform-level abstraction)

#### Attributes

| Attribute | Type | Required | ForceNew | Description |
|-----------|------|----------|----------|-------------|
| `name` | string | yes | yes | Policy name (max 64 chars) |
| `command_lines` | list(string) | yes | no | Commands to execute (min 1) |

#### Validation Rules

| Field | Rule |
|-------|------|
| `name` | Must start with letter, contain only letters/numbers/underscores/hyphens, max 64 chars |
| `command_lines` | At least one command required, each command must not be empty |

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Separate files for provider resource, client service, and parser
- **Modular Design**: Parser functions are reusable for command building and output parsing
- **Dependency Management**: Service depends on executor interface for testability
- **Clear Interfaces**: Well-defined Schedule and KronPolicy data structures

### Performance
- Schedule operations are lightweight (single command execution)
- Configuration save adds latency but ensures persistence
- Parser uses efficient regex-based line processing

### Security
- Commands are passed to router as-is; validation prevents injection
- No sensitive data stored in schedule configuration
- Connection uses SSH with proper authentication

### Reliability
- Idempotent delete operations (handles "not found" gracefully)
- Context cancellation support for all operations
- Configuration saved after each change for persistence

### Validation
- All input validated before sending to router
- Time format validation (HH:MM, 0-23 hour, 0-59 minute)
- Date format validation (YYYY/MM/DD, 2000-2099 year range)
- Day of week validation (sun-sat, ranges, lists)
- Schedule ID range validation (1-65535)
- Policy name format validation

## RTX Commands Reference

### Schedule Commands

```
# Time-based schedule (daily recurring)
schedule at <id> <HH:MM> <command>

# Startup schedule
schedule at <id> startup <command>

# Date/time specific schedule (one-time)
schedule at <id> <YYYY/MM/DD> <HH:MM> <command>

# PP interface schedule (day-of-week based)
schedule pp <pp-num> <day-spec> <HH:MM> connect|disconnect

# Delete schedule
no schedule at <id>

# Delete PP schedule
no schedule pp <pp-num> <day-spec> <HH:MM>

# Show schedules
show config | grep schedule
show config | grep "schedule at <id>"
```

### Examples

```
# Daily backup at 3:00 AM
schedule at 1 3:00 save

# Execute on router startup
schedule at 2 startup dhcp service server

# One-time schedule for maintenance
schedule at 3 2025/12/31 23:59 restart

# PP interface connect on weekdays
schedule pp 1 mon-fri 8:00 connect

# PP interface disconnect on weekends
schedule pp 1 sat,sun 22:00 disconnect
```

## Terraform Command Support

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | Required | Shows changes to schedule configuration |
| `terraform apply` | Required | Creates/updates/deletes schedules |
| `terraform destroy` | Required | Removes all managed schedules |
| `terraform import` | Required | Imports existing schedules |
| `terraform refresh` | Required | Refreshes state from router |
| `terraform state` | Required | Standard state operations |

### Import Specification

#### rtx_kron_schedule
- **Import ID Format**: `<schedule_id>` (integer)
- **Import Command**: `terraform import rtx_kron_schedule.example 1`
- **Post-Import**: All attributes populated from router configuration

#### rtx_kron_policy
- **Import ID Format**: `<policy_name>` (string)
- **Import Command**: `terraform import rtx_kron_policy.example my_policy`
- **Post-Import**: Only name is imported; user must update with correct commands (RTX doesn't store policies)

## Example Usage

### Daily Backup Schedule

```hcl
resource "rtx_kron_schedule" "daily_backup" {
  schedule_id = 1
  name        = "Daily configuration backup"
  at_time     = "3:00"
  recurring   = true

  command_lines = [
    "save"
  ]
}
```

### Startup Initialization

```hcl
resource "rtx_kron_schedule" "startup_init" {
  schedule_id = 2
  on_startup  = true

  command_lines = [
    "dhcp service server",
    "pp select 1"
  ]
}
```

### One-Time Maintenance

```hcl
resource "rtx_kron_schedule" "maintenance" {
  schedule_id = 3
  date        = "2025/12/31"
  at_time     = "23:59"
  recurring   = false

  command_lines = [
    "restart"
  ]
}
```

### Schedule with Policy Reference

```hcl
resource "rtx_kron_policy" "backup_tasks" {
  name = "daily_backup"

  command_lines = [
    "save",
    "show ip route"
  ]
}

resource "rtx_kron_schedule" "scheduled_backup" {
  schedule_id = 10
  at_time     = "2:00"
  recurring   = true
  policy_list = rtx_kron_policy.backup_tasks.name
}
```

### Weekday PP Connection Schedule

```hcl
# Note: day_of_week is supported but PP schedules use a different
# command format (schedule pp) which is handled internally
resource "rtx_kron_schedule" "weekday_connect" {
  schedule_id = 20
  at_time     = "8:00"
  day_of_week = "mon-fri"

  command_lines = [
    "pp select 1",
    "connect"
  ]
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state
- Schedule enabled status is always `true` when managed by Terraform
- Runtime/operational status is not stored in state
- Policy commands are stored in Terraform state (not on router)
- Import reads configuration from router; state is updated accordingly

## Change History

| Date | Source Spec | Changes |
|------|-------------|---------|
| 2026-01-23 | Implementation analysis | Initial master spec creation from existing code |
| 2026-02-07 | Implementation Audit | Full audit against implementation code |
