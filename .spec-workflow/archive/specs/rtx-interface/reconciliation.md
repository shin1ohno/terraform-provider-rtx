# Reconciliation

## Product principles
- Interface schema mirrors Cisco-style naming (name/description/mtu/shutdown-equivalent proxyarp) while using RTX-specific secure filter/NAT descriptors; state excludes runtime link status.

## Implementation alignment
- Resource supports description, single IP block (static or DHCP), secure_filter_in/out, dynamic_filter_out, nat_descriptor binding, proxyarp, and mtu; validations on names/integers present; import/read round-trips filters and NAT descriptors.
- Parser fixes address long filter lists and dynamic filters per import-fidelity work.
- Gaps: no support for multiple/secondary addresses, shutdown/admin state, VRRP/HSRP equivalents, or per-interface DHCP options; lacks guardrails on NAT/filters existence and no warning when both DHCP and address omitted.
