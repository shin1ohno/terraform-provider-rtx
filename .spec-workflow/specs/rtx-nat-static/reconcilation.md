# Reconciliation

## Product principles
- Cisco-style naming for static NAT entries maintained; state restricted to config mappings.

## Implementation alignment
- Resource supports descriptor_id, static entries with inside_local/inside_local_port, outside_global/port, and protocol selection; CRUD/import implemented.
- Validation on IDs/ports exists and parser handles multiple entries.
- Gaps: descriptor type/binding to interfaces not modeled, no bidirectional/overlap checks, no port-range mapping or logging controls, and no guard against conflicting mappings across descriptors.
