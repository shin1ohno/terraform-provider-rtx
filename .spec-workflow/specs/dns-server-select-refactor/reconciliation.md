# Reconciliation

## Product principles
- New schema separates EDNS/record_type/query_pattern to mirror Cisco-style clarity; state remains config-only.
- Naming stays aligned with Cisco IOS XE patterns and RTX terminology without storing operational data.

## Implementation alignment
- `internal/provider/resource_rtx_dns_server.go` exposes `edns`, `record_type`, `query_pattern`, `original_sender`, `restrict_pp`; parser/build helpers cover multi-server select entries.
- Legacy `domains` field removed; parser tests exercise field ordering and EDNS handling.
- Gaps: `query_pattern` is optional instead of required, `original_sender` lacks IP/CIDR validation, and no migration warning if legacy configs are present.
