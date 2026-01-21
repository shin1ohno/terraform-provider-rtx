# Reconciliation

## Product principles
- SNMP resource uses Cisco-like naming and keeps community strings sensitive; runtime status not stored.

## Implementation alignment
- Supports sysLocation/contact/name, communities (ro/rw + ACL), trap hosts (v1/v2c), and trap enable list; singleton CRUD/import implemented.
- Validates IP/permissions and treats communities/trap strings as sensitive.
- Gaps: no SNMPv3 users/groups, no engine ID/auth/priv settings, limited trap type validation, no interface binding or view control, and no safeguards against weak community usage.
