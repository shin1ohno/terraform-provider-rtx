# Schema Pattern Audit Report

This report documents the schema pattern compliance audit for all resources in the Terraform provider for RTX routers.

## Audit Summary

Total resources audited: 25+
Compliance level: Generally good, with some improvements recommended

## Resources Audited

### Singleton Resources (ID pattern: fixed string)

These resources correctly use singleton patterns:

| Resource | ID | Status | Notes |
|----------|-----|--------|-------|
| rtx_admin | "admin" | Compliant | Uses ForceNew on username |
| rtx_bgp | "bgp" | Compliant | Uses Optional+Computed correctly |
| rtx_ospf | "ospf" | Compliant | Uses Optional+Computed correctly |
| rtx_dns_server | "dns" | Compliant | Has DiffSuppressFunc on record_type |
| rtx_syslog | "syslog" | Compliant | Has DiffSuppressFunc on facility |
| rtx_snmp_server | "snmp" | Compliant | Uses Sensitive on community names |
| rtx_system | "system" | Compliant | Nested blocks well-structured |
| rtx_sshd | "sshd" | Compliant | - |
| rtx_sftpd | "sftpd" | Compliant | - |
| rtx_httpd | "httpd" | Compliant | - |

### Multi-Instance Resources (ID pattern: unique identifier)

| Resource | ID Pattern | Status | Notes |
|----------|-----------|--------|-------|
| rtx_admin_user | username | Compliant | ForceNew on username, Sensitive on password |
| rtx_interface | interface name | Compliant | ForceNew on name |
| rtx_static_route | destination | Compliant | ForceNew on destination |
| rtx_dhcp_scope | scope_id | Compliant | ForceNew on scope_id |
| rtx_dhcp_binding | scope_id/ip | Compliant | ForceNew on identifiers |
| rtx_nat_masquerade | descriptor_id | Compliant | ForceNew on descriptor_id |
| rtx_nat_static | descriptor_id | Compliant | CustomizeDiff for port validation |
| rtx_vlan | interface/vlan_id | Compliant | ForceNew on identifiers |
| rtx_bridge | bridge name | Compliant | ForceNew on name |
| rtx_ipsec_tunnel | tunnel_id | Compliant | Uses WriteOnlyStringSchema |
| rtx_pppoe | pp_number | Compliant | Uses WriteOnlyRequiredStringSchema |
| rtx_access_list_ip | sequence | Compliant | Has DiffSuppressFunc |

## Pattern Compliance Analysis

### 1. Attribute Configurability Patterns

#### Required Fields (Required: true)
All resources correctly use `Required: true` for mandatory user inputs:
- Identifiers (username, sequence, scope_id, vlan_id, etc.)
- Critical configuration (asn in BGP, router_id in OSPF)

#### Optional+Computed Fields (Optional: true, Computed: true)
This pattern is correctly applied for fields with API defaults:
- Boolean settings (enabled, administrator, redistribute_static, etc.)
- Numeric settings with defaults (port, timeout values)
- String settings with defaults (protocol, auth_method)

**Improvement Opportunity:**
Some resources use only `Optional: true` without `Computed: true` where the API may provide defaults. Consider reviewing:
- `timezone` in rtx_system (should potentially be Computed if router has default)
- `reconnect_interval` and `reconnect_attempts` in rtx_pppoe

#### Computed Fields (Computed: true only)
Correctly used for read-only values:
- `vlan_interface` in rtx_vlan (derived from vlan_id + interface)

### 2. ForceNew/RequiresReplace Pattern

All identifier fields correctly use `ForceNew: true`:
- `username` in rtx_admin_user
- `name` in rtx_interface
- `descriptor_id` in NAT resources
- `vlan_id` and `interface` in rtx_vlan
- `sequence` in access list resources
- `tunnel_id` in rtx_ipsec_tunnel
- `pp_number` in rtx_pppoe

### 3. Sensitive Fields Pattern

Correctly implemented for credential fields:
- `password` in rtx_admin_user (Sensitive: true)
- `pre_shared_key` in rtx_ipsec_tunnel (via WriteOnlyStringSchema)
- `password` in rtx_pppoe (via WriteOnlyRequiredStringSchema)
- `community.name` in rtx_snmp_server (Sensitive: true)
- `password` in BGP neighbor (Sensitive: true)

### 4. DiffSuppressFunc Pattern

Currently implemented:
- `SuppressCaseDiff` for case-insensitive comparisons:
  - `record_type` in rtx_dns_server
  - `facility` in rtx_syslog
  - `action` in rtx_access_list_ip
  - `protocol` in rtx_access_list_ip

**Improvement Opportunities:**
Consider adding DiffSuppressFunc for:
- IP address fields (IPv4 format normalization: leading zeros)
- Port range fields (e.g., "80" vs "80-80")
- Protocol fields in other resources (consistency)

### 5. Nested Block Patterns

Correctly implemented patterns:

#### TypeList for ordered collections:
- `neighbor` in rtx_bgp and rtx_ospf
- `network` in rtx_bgp and rtx_ospf
- `static_entry` in rtx_nat_masquerade
- `entry` in rtx_nat_static
- `filter_rules` patterns in access lists
- `server_select` in rtx_dns_server

#### TypeSet for unordered collections:
- `host` in rtx_syslog (correctly uses TypeSet)

#### Single nested block (MaxItems: 1):
- `console` in rtx_system
- `statistics` in rtx_system
- `ikev2_proposal` in rtx_ipsec_tunnel
- `ipsec_transform` in rtx_ipsec_tunnel

### 6. Zero Value Handling

Most resources correctly use `GetOk()` pattern for optional fields.

**Improvement Opportunities:**
- Some resources use direct `d.Get()` without checking if value was explicitly set
- Boolean fields should use `GetOkExists()` where `false` is meaningful

### 7. Validation Patterns

Well-implemented custom validators:
- `validateASN` for BGP AS numbers
- `validateIPv4Address` for IP addresses
- `validateTimezone` for timezone formats
- `validateBridgeName` and `validateBridgeMember` for bridge configuration
- `validateVLANInterfaceName` and `validateSubnetMask` for VLAN
- `validateConsoleLines` for system console

## Recommended Changes

### High Priority

1. **Add DiffSuppressFunc for IP addresses**
   - Files: Multiple resources with IP address fields
   - Add `SuppressEquivalentIPDiff` function from design.md
   - Apply to: `ip_address`, `local_address`, `remote_address` fields

2. **Consistent Sensitive field handling**
   - Ensure all password/key fields use `Sensitive: true`
   - Consider using helper functions consistently

### Medium Priority

3. **Add Optional+Computed to fields with API defaults**
   - Review fields that are Optional-only but have router defaults
   - Update to Optional+Computed pattern

4. **Extend DiffSuppressFunc usage**
   - Add case-insensitive comparison to more protocol/action fields
   - Consider JSON comparison for complex configuration strings

### Low Priority

5. **Documentation improvements**
   - Add `Description` to all nested block schemas
   - Standardize description formatting

6. **StateFunc for normalization**
   - Consider adding `StateFunc` for IP address normalization
   - Consider adding `StateFunc` for protocol name lowercase

## Pattern Decision Tree Application

Based on the design document decision tree:

```
Is the value always provided by the user?
├── YES → Required: true
│   Examples: username, sequence, vlan_id, scope_id
└── NO
    ├── Is the value ever provided by the user?
    │   ├── NO → Computed: true (read-only)
    │   │   Examples: vlan_interface
    │   └── YES
    │       ├── Can API/router provide a default?
    │       │   ├── YES → Optional: true, Computed: true
    │       │   │   Examples: enabled, administrator, port
    │       │   └── NO → Optional: true
    │       │       Examples: description, name
    │       └── Is change immutable after create?
    │           ├── YES → Add ForceNew: true
    │           │   Examples: username, tunnel_id
    │           └── NO → (no additional flag)
    └── Is it sensitive (password, key)?
        ├── YES → Add Sensitive: true
        │   Examples: password, pre_shared_key
        └── NO → (no additional flag)
```

All resources follow this decision tree appropriately.

## Files Requiring Updates

Based on this audit, the following files may benefit from pattern updates:

| File | Change Type | Priority |
|------|-------------|----------|
| Multiple | Add SuppressEquivalentIPDiff | Medium |
| resource_rtx_system.go | Consider Optional+Computed for timezone | Low |
| resource_rtx_pppoe.go | Consider Optional+Computed for reconnect fields | Low |

## Conclusion

The codebase demonstrates good adherence to schema design patterns. The main opportunities for improvement are:
1. Adding DiffSuppressFunc for semantic IP address equality
2. Reviewing fields for Optional+Computed pattern where API defaults exist
3. Ensuring consistent use of Sensitive and WriteOnly patterns

No critical compliance issues were found.
