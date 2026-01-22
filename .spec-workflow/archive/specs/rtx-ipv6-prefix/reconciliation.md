# Reconciliation

## Product principles
- Resource keeps Cisco-like naming while reflecting RTX prefix sources; state stores configuration only.

## Implementation alignment
- Supports prefix_id, prefix_length, source enum (static/ra/dhcpv6-pd), optional interface, and static prefix value; custom diff enforces prefix/interface presence based on source; CRUD/import implemented.
- Validation guards source/prefix/interface combinations and basic IPv6 formatting.
- Gaps: no deep IPv6 prefix validation (compression/length), no checks that referenced interface exists or prefix_length matches ISP constraints, and no metadata for prefix delegation parameters beyond length.
