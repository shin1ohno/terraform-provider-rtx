# Reconciliation

## Product principles
- Kron-style naming mirrors Cisco scheduler patterns; state captures only configured commands/times.

## Implementation alignment
- `rtx_kron_policy` stores command_lines; `rtx_kron_schedule` supports name, at_time, recurring/day-of-week, startup flag, and policy_list; CRUD/import implemented.
- Time validation exists for basic HH:MM and startup scheduling.
- Gaps: no interval-based or date-specific schedules, no interface/condition-based triggers, no schedule groups/enable flags, limited validation of command syntax, and no timezone/DST handling beyond router default.
