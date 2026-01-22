# Reconciliation

## Product principles
- Resource mirrors Cisco static route naming (`prefix/mask/next_hop`) and stores config-only state.

## Implementation alignment
- Supports prefix/mask with multiple next_hop objects (gateway or interface), distance, permanent flag, and filter; CRUD/import implemented and parser handles multiple gateways per prefix.
- Validation covers IP/mask formats and distance range.
- Gaps: no support for route tagging/metric-type, VRF/table selection, or preference weighting beyond distance; no guard against duplicate routes or conflicting next-hop definitions; route redistribution controls sit outside this resource.
