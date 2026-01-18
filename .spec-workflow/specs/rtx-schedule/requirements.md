# Requirements: rtx_schedule

## Overview
Terraform resource for managing scheduled tasks and time-based automation on Yamaha RTX routers.

## Functional Requirements

### 1. CRUD Operations
- **Create**: Define scheduled commands
- **Read**: Query schedule configuration
- **Update**: Modify schedule parameters
- **Delete**: Remove scheduled tasks

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

### 6. Schedule Groups
- Group related schedules
- Enable/disable groups

### 7. Import Support
- Import existing schedules

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
resource "rtx_schedule" "daily_backup" {
  id = 1

  time {
    hour   = 2
    minute = 0
  }

  recurrence = "daily"

  command = "copy config sd1:backup/config-$(date +%Y%m%d).txt"
}

resource "rtx_schedule" "weekly_reboot" {
  id = 2

  time {
    day_of_week = "sunday"
    hour        = 4
    minute      = 0
  }

  recurrence = "weekly"

  command = "restart"
}

resource "rtx_schedule" "business_hours_vpn" {
  id = 3

  time_range {
    start_hour = 8
    end_hour   = 18
  }

  days = ["monday", "tuesday", "wednesday", "thursday", "friday"]

  commands = [
    "pp select 1",
    "pp always-on on"
  ]
}

resource "rtx_schedule" "startup_task" {
  id = 10

  on_startup = true

  command = "syslog info 'Router started successfully'"
}
```
