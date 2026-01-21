# Reconciliation

## Product principles
- Goal supports “Fail Safely” and import fidelity by strengthening parser coverage; no conflicts with Cisco naming/state clarity.

## Implementation alignment
- Pattern catalogs exist, but there is no automated extraction from the 59-chapter docs, no categorized pattern inventory, and no generated coverage/gap report.
- Test fixtures are not organized per-command/expected layout described; interface/value type exhaustiveness is not codified.
- Gaps: need `[書式]`/`[設定値]` scraping, categorization by feature, generated table-driven tests with coverage metrics, and interface-name/value-type combinatorial tests.
