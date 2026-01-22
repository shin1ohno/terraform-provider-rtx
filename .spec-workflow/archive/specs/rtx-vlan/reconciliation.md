# Reconciliation

## Product principles
- Resource naming matches Cisco VLAN patterns; state contains config only and avoids operational status.

## Implementation alignment
- Schema includes vlan_id, interface (parent), name, ip_address/mask, shutdown flag, and computed vlan_interface; CRUD/import implemented with validations.
- ID parsing uses interface/vlan_id format consistent with import spec.
- Gaps: no port assignment/tagging, no DHCP client/secondary IPs, no inter-VLAN routing controls or access policy bindings, and no validation of parent interface availability; shutdown default is computed rather than explicit.
