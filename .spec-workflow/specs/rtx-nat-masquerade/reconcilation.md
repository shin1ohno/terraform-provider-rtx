# Reconciliation

## Product principles
- Resource naming follows Cisco NAT patterns while using RTX descriptor concepts; state is config-only.

## Implementation alignment
- Schema covers descriptor_id, outer_address, optional inner_network, and static_entry with protocol-only support (`esp/ah/gre/icmp`) or tcp/udp ports; CRUD/import implemented.
- Validation on descriptor_id and IP/port ranges present; protocol-only mappings satisfy ESP/AH use cases.
- Gaps: no interface binding field or inbound/outbound application, no descriptor type/inside network range enforcement, no session logging or port-range mapping controls, and lacks port translation range/overload tuning; dependency with interfaces not enforced.
