# Reconciliation

## Product principles
- Focus on completing security/NAT coverage fits the “comprehensive RTX feature” goal while keeping Cisco-aligned naming.
- State handling stays configuration-only; no runtime data is persisted.

## Implementation alignment
- `rtx_access_list_ip` now accepts `restrict*` actions and tcpfin/tcprst/tcpsyn/established protocols; dynamic filters support protocol and filter-list forms; NAT masquerade static entries allow protocol-only (`esp`/`ah`/`gre`/`icmp`) with optional ports.
- Pattern catalogs exist for IP/NAT/Ethernet filters and dynamic filter parser supports wrapped lists, reducing prior import loss.
- Gaps: no validation tying `restrict` use to dynamic filters, dynamic filter protocol enum omits many documented services/timeouts, and Ethernet filtering is implemented as MAC ACL (permit/deny) rather than RTX `ethernet filter` with pass/reject and interface application; dynamic timer/global settings absent.
