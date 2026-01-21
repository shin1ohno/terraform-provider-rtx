# Reconciliation

## Product principles
- DDNS resources align with Cisco-like naming (separate NetVolante/custom) and keep credentials sensitive; state clarity maintained.

## Implementation alignment
- `rtx_ddns` supports server_id/url/hostname/auth; `rtx_netvolante_dns` handles interface, hostname, server selection, IPv6 toggle, timeout, auto_hostname; status data source reports last update/status.
- Parsers/services exist for NetVolante and custom DDNS CRUD and status retrieval.
- Gaps: no update trigger controls (intervals, retries, interface up/down), no multi-WAN priority/failover logic, limited provider options (no API-key/custom params), and missing hostname availability checks; delete operations lack warnings for active registrations.
