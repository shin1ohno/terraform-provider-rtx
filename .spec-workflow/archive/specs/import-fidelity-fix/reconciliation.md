# Reconciliation

## Product principles
- Import fidelity work directly supports “Import Support” and “Fail Safely” while keeping state limited to configuration data.
- Naming and schema remain Cisco-style; no operational status is persisted.

## Implementation alignment
- DNS server-select parsing now handles multi-server entries, record types, EDNS flags, and complex patterns; interface secure filter parsing captures long lists and dynamic IDs; static routes keep multiple next-hops/weights; L2TP tunnel auth fields are present; admin user attributes (login_timer/gui_pages) round-trip.
- Parser fixtures/tests added under `internal/rtx/parsers` and `testdata/import_fidelity` to cover the reported truncation/ordering issues.
- Gaps: imports still rely on schema defaults for some optional fields (e.g., passwords cannot be read, some computed booleans may mask missing values) and there is no explicit user-facing warning when router output omits fields; retry/robustness around large SSH outputs is minimal.
