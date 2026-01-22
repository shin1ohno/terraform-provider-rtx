# Reconciliation

## Product principles
- Schema keeps Cisco-like access-list naming while honoring RTX actions (`restrict*`) and avoids storing runtime counters.

## Implementation alignment
- Static filter resource supports filter_id, actions (pass/reject/restrict variants), protocols incl. tcpfin/tcprst/tcpsyn/established, source/destination, ports, and established flag; CRUD/import implemented.
- Dynamic filter resource supports protocol-based and filter-list forms with syslog/timeout options and parser support for long lists.
- Gaps: no logging variants for pass/reject, no tcpflag custom masks or icmp-info/error distinctions, no validation tying restrict rules to dynamic filters, and no interface/application resource as spec implies; dynamic protocol enum omits many documented services.
