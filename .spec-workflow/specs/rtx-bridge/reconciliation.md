# Reconciliation

## Product principles
- Bridge resource aligns with Cisco-style interface naming while keeping scope minimal and configuration-only.
- No operational status stored; uses direct RTX terms for member interfaces.

## Implementation alignment
- Schema covers bridge name and members; validation allows lan/lan-vlan/tunnel/pp/loopback members; CRUD/import implemented via `internal/client/bridge`.
- Supports multiple members and state round-trip.
- Gaps: no enforcement of bridge count limits or duplicate-member prevention across bridges; no explicit L2TP-specific validation; lacks member direction/options beyond membership.
