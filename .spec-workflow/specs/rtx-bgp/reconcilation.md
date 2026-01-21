# Reconciliation

## Product principles
- Resource naming (`rtx_bgp`, `neighbor`, `network`) follows Cisco IOS XE patterns; config-only state upheld.
- Does not mix runtime neighbor status into state, matching “state clarity.”

## Implementation alignment
- Schema supports ASN, router_id, default_ipv4_unicast, log_neighbor_changes, neighbors (basic params, password, multihop/local_address), networks, and static/connected redistribution; import/export implemented.
- Service/parser handle basic enable/disable and neighbor/network mapping.
- Gaps: no address-family separation (IPv6, VPNv4), no route-maps/prefix/as-path/community controls, no advanced roles (RR/confederation), no timers/keepalive override per neighbor beyond basic fields; graceful restart and policy filtering absent.
