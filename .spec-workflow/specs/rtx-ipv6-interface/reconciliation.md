# Reconciliation

## Product principles
- Schema mirrors Cisco-esque naming (rtadv/dhcpv6/filters) while using RTX constructs; state remains configuration-only.

## Implementation alignment
- Supports multiple addresses (static or prefix_ref+interface_id), RTADV block (o_flag/m_flag/lifetime), dhcpv6_service enum, IPv6 MTU, secure_filter in/out, dynamic_filter_out; CRUD/import implemented with validations.
- Parser handles long filter lists per import-fidelity fixes.
- Gaps: no explicit RA enable/disable guard when rtadv omitted, no validation coupling prefix_ref with interface_id, DHCPv6 mode lacks IR flag/options, and no checks for prefix existence or filter dependencies; router lifetime/MTU bounds not enforced against IPv6 minimums beyond basic range.
