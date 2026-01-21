# Reconciliation

## Product principles
- Current implementation follows Cisco MAC ACL naming but diverges from RTX `ethernet filter` semantics; still keeps configuration-only state.

## Implementation alignment
- Implemented resource `rtx_access_list_mac` provides named ACL with permit/deny entries, MAC masks, ether_type, VLAN, logging, and interface binding via separate ACL resources.
- Gaps: spec expects numeric `ethernet filter` with pass/reject(-log/-nolog), DHCP-based matches, and direct interface application; current schema/action set and command mapping do not mirror RTX filter syntax, and import by filter number is unsupported.
