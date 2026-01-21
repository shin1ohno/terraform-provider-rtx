# Reconciliation

## Product principles
- Schema mirrors Cisco-style DHCP scope naming while staying RTX-specific; state-only persistence upheld.

## Implementation alignment
- Resource supports scope_id/network (CIDR), lease_time, exclude_ranges, and options (routers/dns/domain); CRUD/import wired through client/parsers.
- Validation covers CIDR/IP formats; dependency with bindings can be expressed via Terraform graph.
- Gaps: no explicit start/end range derivation or excludes validation against network, no warnings for active leases on delete, and lease_time handling limited (no infinite sentinel beyond string), plus missing per-scope gateway/DNS count enforcement and DHCP option breadth beyond routers/DNS/domain.
