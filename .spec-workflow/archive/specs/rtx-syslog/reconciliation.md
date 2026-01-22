# Reconciliation

## Product principles
- Syslog resource keeps Cisco-like facility/level naming and stores only configuration; no runtime buffers kept.

## Implementation alignment
- Supports multiple hosts with address/port, local_address, facility, and level toggles (notice/info/debug); singleton CRUD/import implemented.
- Validation for hosts/facility present.
- Gaps: no per-host level, transport (TCP/TLS) selection, source interface binding, or message format controls; lacks rate-limit/logging options and timezone handling.
