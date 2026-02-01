# Session Progress: Complete rtx_tunnel Resource Attributes

## Objective
Add missing attributes to the `rtx_tunnel` resource to ensure all RTX router config commands are mappable to Terraform attributes.

## Completed Tasks

### 1. Parser Struct Changes (`internal/rtx/parsers/tunnel.go`)
Added new fields:
- **Tunnel struct**: `EndpointName`, `EndpointNameType`
- **TunnelIPsec struct**: `NATTraversal`, `IKERemoteName`, `IKERemoteNameType`, `IKEKeepaliveLog`, `IKELog`
- **TunnelL2TP struct**: `KeepaliveLog` (DisconnectTime and SyslogEnabled already existed)

### 2. Regex Patterns and Parsing Logic
Added patterns for:
- `tunnel endpoint name <addr> [fqdn]`
- `ipsec ike nat-traversal N on/off`
- `ipsec ike remote name N <type> <value>`
- `ipsec ike keepalive log N on/off`
- `ipsec ike log N <options>`
- `l2tp tunnel disconnect time off/<seconds>`
- `l2tp keepalive log on/off`

### 3. Command Builder Functions
Added new functions:
- `BuildTunnelEndpointNameCommand(address, nameType string) string`
- `BuildIPsecIKENATTraversalCommand(tunnelID int, enabled bool) string`
- `BuildIPsecIKERemoteNameCommand(tunnelID int, nameType, value string) string`
- `BuildIPsecIKEKeepaliveLogCommand(tunnelID int, enabled bool) string`
- `BuildIPsecIKELogCommand(tunnelID int, options string) string`
- `BuildL2TPKeepaliveLogCommand(enabled bool) string`

Updated `BuildL2TPDisconnectTimeCommand` in `l2tp.go` to handle `off` case (seconds == 0).

### 4. Client Interfaces (`internal/client/interfaces.go`)
Updated structs:
- **Tunnel**: Added `EndpointName`, `EndpointNameType`
- **TunnelIPsec**: Added `NATTraversal`, `IKERemoteName`, `IKERemoteNameType`, `IKEKeepaliveLog`, `IKELog`
- **TunnelL2TP**: Added `KeepaliveLog`

### 5. Tunnel Service Converters (`internal/client/tunnel_service.go`)
Updated `convertToParserTunnel` and `convertFromParserTunnel` to include all new fields.

### 6. Terraform Model (`internal/provider/resources/tunnel/model.go`)
Added new fields:
- **TunnelModel**: `EndpointName`, `EndpointNameType`
- **TunnelIPsecModel**: `NATTraversal`, `IKERemoteName`, `IKERemoteNameType`, `IKEKeepaliveLog`, `IKELog`
- **TunnelL2TPModel**: `DisconnectTime`, `KeepaliveLog`, `Syslog`

Updated `ToClient()` and `FromClient()` methods.

### 7. Terraform Schema (`internal/provider/resources/tunnel/resource.go`)
Added new schema attributes:
- **Root level**: `endpoint_name`, `endpoint_name_type`
- **IPsec block**: `nat_traversal`, `ike_remote_name`, `ike_remote_name_type`, `ike_keepalive_log`, `ike_log`
- **L2TP block**: `disconnect_time`, `keepalive_log`, `syslog`

### 8. Tests (`internal/rtx/parsers/tunnel_test.go`)
- Updated `TestTunnelParser_ParseL2TPv3Tunnel` to test all new attributes
- Updated `TestBuildTunnelCommands_L2TPv3` to verify new commands
- Added `TestBuildNewTunnelCommands` for individual command builder tests

## New Attribute Summary

| Level | Terraform Attribute | RTX Command |
|-------|---------------------|-------------|
| Root | `endpoint_name` | `tunnel endpoint name <addr>` |
| Root | `endpoint_name_type` | `tunnel endpoint name <addr> fqdn` |
| IPsec | `nat_traversal` | `ipsec ike nat-traversal N on/off` |
| IPsec | `ike_remote_name` | `ipsec ike remote name N <type> <value>` |
| IPsec | `ike_remote_name_type` | (type field of above) |
| IPsec | `ike_keepalive_log` | `ipsec ike keepalive log N on/off` |
| IPsec | `ike_log` | `ipsec ike log N <options>` |
| L2TP | `disconnect_time` | `l2tp tunnel disconnect time off/<N>` |
| L2TP | `keepalive_log` | `l2tp keepalive log on/off` |
| L2TP | `syslog` | `l2tp syslog on/off` |

## Command Generation Rules

Commands are only generated when:
- `endpoint_name`: Non-empty
- `nat_traversal`: `true` (false is default, no command generated)
- `ike_remote_name/type`: Both non-empty
- `ike_keepalive_log`: `true` (false is default, no command generated)
- `ike_log`: Non-empty
- `disconnect_time`: Always generated (0 = "off")
- `keepalive_log`: `true` (false is default, no command generated)
- `syslog`: `true` (false is default, no command generated)

### 9. Example Configuration (`examples/import/main.tf`)
Updated both tunnel resources with new attributes:
- `rtx_tunnel.hnd_itm` (L2TPv3): Added all new attributes
- `rtx_tunnel.remote_access` (L2TP): Added applicable new attributes

### 10. Bug Fix: disconnect_time Inconsistency
Fixed `FromClient()` in model.go to use `types.Int64Value()` instead of `fwhelpers.Int64ValueOrNull()` for `disconnect_time`.

**Problem**: `Int64ValueOrNull(0)` returns null, but 0 is a valid value meaning "off" for disconnect_time. This caused "Provider produced inconsistent result after apply" errors.

**Solution**: Always use `types.Int64Value(int64(tunnel.L2TP.DisconnectTime))` since 0 is semantically meaningful.

## Test Results
All tests pass:
```
ok  	github.com/sh1/terraform-provider-rtx/internal/client	8.315s
ok  	github.com/sh1/terraform-provider-rtx/internal/rtx/parsers	0.264s
```

## Example Configuration
```hcl
resource "rtx_tunnel" "hnd_itm" {
  tunnel_id          = 1
  encapsulation      = "l2tpv3"
  enabled            = true
  endpoint_name      = "itm.ohno.be"
  endpoint_name_type = "fqdn"

  ipsec {
    ipsec_tunnel_id      = 101
    local_address        = "192.168.1.253"
    remote_address       = "itm.ohno.be"
    pre_shared_key       = var.admin_password
    secure_filter_in     = [200028, 200099]
    tcp_mss_limit        = "auto"
    nat_traversal        = true
    ike_remote_name      = "key-id"
    ike_remote_name_type = "l2tpv3"
    ike_keepalive_log    = false
    ike_log              = "key-info message-info payload-info"

    ipsec_transform {
      protocol          = "esp"
      encryption_aes128 = true
      integrity_sha1    = true
    }

    keepalive {
      enabled  = true
      mode     = "heartbeat"
      interval = 10
      retry    = 6
    }
  }

  l2tp {
    hostname         = "ebisu-RTX1210"
    local_router_id  = "192.168.1.253"
    remote_router_id = "192.168.1.254"
    remote_end_id    = "shin1"
    always_on        = true
    disconnect_time  = 0  # "off"
    keepalive_log    = false
    syslog           = true

    tunnel_auth {
      enabled  = true
      password = var.admin_password
    }

    keepalive {
      enabled  = true
      interval = 60
      retry    = 3
    }
  }
}
```

### 11. Bug Fix: Extra IKE Commands Breaking Tunnel Connection

**Problem**: L2TPv3 tunnel connection was not being established despite Terraform plan showing no changes.

**Root Cause**: `buildTunnelIPsecCommands` in `internal/rtx/parsers/tunnel.go` was unconditionally generating:
- `ipsec ike encryption N aes-cbc`
- `ipsec ike hash N sha256`
- `ipsec ike group N modp2048`

These commands did NOT exist in the working configuration (both hnd and itm routers). The peer router (itm) used default IKE negotiation without explicit encryption/hash/group settings. When hnd explicitly specified these values, IKE negotiation failed due to mismatch.

**Comparison**:
| Command | Expected (hnd/config.txt) | Terraform Generated | itm Config |
|---------|---------------------------|---------------------|------------|
| `ipsec ike encryption 1 aes-cbc` | **Missing** | Generated | **Missing** |
| `ipsec ike hash 1 sha256` | **Missing** | Generated | **Missing** |
| `ipsec ike group 1 modp2048` | **Missing** | Generated | **Missing** |

**Solution**: Added `isIKEv2ProposalSet()` helper function and conditional command generation:

```go
// isIKEv2ProposalSet returns true if any IKEv2 proposal settings are explicitly configured
func isIKEv2ProposalSet(proposal IKEv2Proposal) bool {
    return proposal.EncryptionAES256 || proposal.EncryptionAES128 || proposal.Encryption3DES ||
        proposal.IntegritySHA256 || proposal.IntegritySHA1 || proposal.IntegrityMD5 ||
        proposal.GroupFourteen || proposal.GroupFive || proposal.GroupTwo
}

// In buildTunnelIPsecCommands:
if isIKEv2ProposalSet(ipsec.IKEv2Proposal) {
    commands = append(commands, BuildIPsecIKEEncryptionCommand(tunnelID, ipsec.IKEv2Proposal))
    commands = append(commands, BuildIPsecIKEHashCommand(tunnelID, ipsec.IKEv2Proposal))
    commands = append(commands, BuildIPsecIKEGroupCommand(tunnelID, ipsec.IKEv2Proposal))
}
```

Now IKEv2 proposal commands are only generated when explicitly configured in Terraform. When not set, the router uses its default negotiation behavior, matching the peer router.

---

## Session: 2026-02-01 - Sync Examples and Spec with Implementation

### Objective
Synchronize examples and spec documentation with the current `rtx_tunnel` implementation.

### Completed Tasks

#### 1. Updated `examples/tunnel/main.tf`
- Removed `name` attribute from 4 examples (it's Computed/read-only)
- Added `endpoint_name` and `endpoint_name_type` to Example 2 (L2TPv3)
- Added `nat_traversal = true` to Example 2 (L2TPv3)
- Added `ipsec` block to Example 3 (L2TPv2) - required for L2TP encapsulation

#### 2. Updated VPN Master Spec (`.spec-workflow/master-specs/vpn/`)
- **requirements.md**: Added `rtx_tunnel` as Resource 0 with full attribute documentation
- **requirements.md**: Added deprecation notices to `rtx_ipsec_tunnel` and `rtx_l2tp`
- **requirements.md**: Updated Resources Summary table with status column
- **design.md**: Added TunnelService component and architecture
- **design.md**: Added Tunnel data models (Tunnel, TunnelIPsec, TunnelL2TP, etc.)
- **design.md**: Updated File Structure with tunnel resource files

#### 3. Updated Feature Spec (`.spec-workflow/specs/rtx-tunnel-unified/`)
- **design.md**: Updated Data Models with new attributes:
  - `EndpointName`, `EndpointNameType` (Tunnel)
  - `NATTraversal`, `IKERemoteName`, `IKERemoteNameType`, `IKEKeepaliveLog`, `IKELog` (TunnelIPsec)
  - `DisconnectTime`, `KeepaliveLog`, `Syslog` (TunnelL2TP)
- **design.md**: Updated Terraform Schema to show `name` as Computed (not Optional)
- **requirements.md**: Updated FR-1, FR-2, FR-3 with new attributes
- **requirements.md**: Fixed example code to not set `name`
- **tasks.md**: Marked completed tasks (1-8, 10-11) with ✅

### Key Findings

1. **`name` is Computed**: RTX does not support setting tunnel descriptions within the tunnel context. The `name` attribute is read-only.

2. **L2TPv2 requires IPsec**: When `encapsulation = "l2tp"`, the `ipsec` block with `pre_shared_key` is required.

3. **New attributes implemented but not documented**:
   - `endpoint_name`, `endpoint_name_type` - DNS resolution for tunnel endpoints
   - `nat_traversal` - NAT traversal support
   - `ike_remote_name`, `ike_remote_name_type` - IKE remote identification
   - `ike_keepalive_log`, `ike_log` - IKE logging options
   - `disconnect_time`, `keepalive_log`, `syslog` - L2TP options

### Validation Results
- `go generate ./...` - Success
- `terraform validate` on examples/tunnel - Success
- `go test ./internal/client/... -run TestTunnel` - All tests pass
- `go test ./internal/rtx/parsers/... -run TestTunnel` - All tests pass

---

## Session: 2026-02-01 - Full Codebase Example Sync

### Objective
Audit and fix ALL examples in the codebase to match the actual implementation schemas.

### Fixed Examples

| Example | Issue | Fix |
|---------|-------|-----|
| `admin/main.tf` | `connection` → `connection_methods` | Renamed attribute |
| `interface/main.tf` | Non-existent `access_list_ip_*` attributes | Removed unsupported attributes |
| `bgp/main.tf` | Multiple schema issues | Complete rewrite: `address`→`ip`, added `index`, `keep_alive`→`keepalive`, `network.prefix`→`ip+wildcard+area`, removed `redistribution` blocks |
| `bgp/variables.tf` | Duplicate variable declarations | Deleted file |
| `ddns/main.tf` | Non-existent `rtx_ddns_status` data source | Removed data source and related output |
| `dns_server/main.tf` | `server_select.servers` → nested `server` blocks | Restructured to use `server { address }` pattern |
| `ipsec_transport/main.tf` | Missing terraform/provider blocks | Added required blocks |
| `ipsec_tunnel/main.tf` | Wrong attribute names | Fixed: `local_id`→`local_address`, `remote_id`→`remote_address`, `ikev2_proposal.encryption`→`encryption_aes256`, `dpd.enabled`→`dpd_enabled` |
| `ipv6_interface/main.tf` | Required `rtadv.prefix_id` missing | Simplified to include rtadv for all interfaces |
| `l2tp/main.tf` | Multiple schema issues | Fixed: `tunnel_name`→`name`, `l2tpv3_config.local_session_id`→`local_router_id`, etc. |
| `l2tp_service/main.tf` | Missing terraform/provider blocks | Added required blocks |
| `ospf/main.tf` | Multiple schema issues | Fixed: `network.prefix`→`ip+wildcard+area`, `neighbor.address`→`ip`, merged into singleton resource |
| `ospf/variables.tf` | Duplicate variable declarations | Deleted file |
| `pppoe/main.tf` | Non-existent `access_list_ip_*` and `secure_filter_*` on pp_interface | Removed unsupported attributes |
| `qos/main.tf` | Invalid `wan1` interface name, non-existent `rtx_ip_filter` resource | Changed to `lan2`, removed ip_filter example |
| `schedule/main.tf` | `policy_list` attribute not working as expected | Changed to use `command_lines` directly |
| `sftp-enabled/main.tf` | `name`→`username`, `administrator = "on"`→`= true` | Fixed rtx_admin_user attributes |
| `snmp/main.tf` | Missing terraform/provider blocks | Added required blocks |
| `syslog/main.tf` | Missing terraform/provider blocks | Added required blocks |

### Common Issues Found

1. **Missing terraform/provider blocks**: Many examples had no terraform version or provider configuration
2. **Incorrect attribute names**: Many examples used hypothetical names that don't match the actual schema
3. **Non-existent resources**: Some examples used resources that don't exist (e.g., `rtx_ip_filter`, `rtx_ddns_status`)
4. **Schema mismatches**: Nested blocks vs lists, required vs optional attributes
5. **Duplicate variable files**: Some examples had both inline variables and separate variables.tf files

### Validation
All examples now pass `terraform validate` (with only provider override warnings).

### Files Modified
- 18 example main.tf files updated
- 2 variables.tf files deleted (bgp, ospf)

---

## Session: 2026-02-01 - Master Spec Audit and Sync

### Objective
Audit and update all 12 master specs to accurately reflect the current implementation.

### Summary of Updates

#### Plugin Framework Updates (All Specs)
Changed "Terraform Plugin SDK v2" to "**Terraform Plugin Framework**" in:
- interface/design.md
- routing/design.md
- management/design.md
- admin/design.md
- dhcp/design.md
- dns/design.md
- ppp/design.md
- qos/design.md
- nat/design.md
- access-list/design.md (date fix)

#### Interface Spec
- Marked 5 filter attributes as **not yet implemented**:
  - `secure_filter_in`, `secure_filter_out`
  - `dynamic_filter_out`
  - `ethernet_filter_in`, `ethernet_filter_out`
- Updated file structure to modular pattern

#### Routing Spec
- Fixed BGP neighbor attribute: `id` → `index`
- Fixed OSPF area attribute: `id` → `area_id`
- Updated schema examples

#### Management Spec
- Marked `rtx_sftpd` as **not implemented**
- Added status column to Resources Summary table

#### PPP Spec
- Marked `secure_filter_in` and `secure_filter_out` as **not yet implemented** in requirements.md

#### DHCP Spec
- Clarified `hostname` and `description` are "**Terraform-only, not sent to router**"

#### VPN Spec (Already Updated)
- `rtx_tunnel` unified resource already documented
- Deprecation notices already in place for `rtx_ipsec_tunnel` and `rtx_l2tp`

### Files Modified

| Spec | requirements.md | design.md |
|------|-----------------|-----------|
| interface | ✅ | ✅ |
| routing | ✅ | ✅ |
| management | ✅ | ✅ |
| admin | - | ✅ |
| dhcp | - | ✅ |
| dns | - | ✅ |
| ppp | ✅ | ✅ |
| qos | - | ✅ |
| nat | - | ✅ |
| ipv6 | - | - (already correct) |
| access-list | - | ✅ (date fix) |
| vpn | ✅ (already updated) | ✅ (already updated) |

### Key Findings

1. **All resources use Terraform Plugin Framework**, not the deprecated SDK v2
2. **Some documented attributes are not yet implemented**:
   - Interface: 5 filter attributes
   - PPP: secure_filter_in/out
   - Management: rtx_sftpd resource
3. **File structure has been modernized** to `internal/provider/resources/{name}/` pattern
4. **Attribute naming issues** in routing spec (id vs index, id vs area_id)

---

## Session: 2026-02-01 - Validation Fixes and ACL Naming Unification

### Objective
Fix validation issues blocking `terraform apply` after import and continue ACL naming unification.

### Completed Tasks

#### 1. NAT Masquerade "primary" Validation Fix
**Problem**: `outer_address = "primary"` was rejected during validation.

**Root Cause**: `ValidateOuterAddress` in `internal/rtx/parsers/nat_masquerade.go` didn't include "primary" or "secondary" as valid values.

**Solution**: Added check for "primary" and "secondary" values:
```go
// "primary" and "secondary" are valid RTX values for using interface IP
if address == "primary" || address == "secondary" {
    return nil
}
```

**Files Modified**:
- `internal/rtx/parsers/nat_masquerade.go:443-446`
- `internal/rtx/parsers/nat_masquerade_test.go:1756-1757` - Updated test to expect success

#### 2. Admin User Password Validation Analysis
**Problem**: Bug report claimed "password is required" error during import.

**Analysis**: Reviewed the code flow:
- `admin_user/resource.go`: `password` is defined as `Optional`
- `admin_service.go:GetAdminUser`: Does NOT call validation
- `admin_service.go:UpdateAdminUser`: Uses `ValidateUserConfigForAttributeUpdate` when password is empty

**Conclusion**: The code already handles password-less operations correctly. Password is only required for `CreateAdminUser`. For imported resources where password is unknown, `UpdateAdminUser` works without password validation.

#### 3. ACL Import Entry Loading Analysis
**Problem**: Bug report claimed entries are not saved to state after import.

**Analysis**:
- Static ACLs (`access_list_ip`, `access_list_ipv6`) support importing with sequence numbers: `terraform import rtx_access_list_ip.test name:1,2,3`
- Dynamic ACLs (`access_list_ip_dynamic`, `access_list_ipv6_dynamic`) intentionally don't import entries because RTX doesn't track filter-to-list associations
- Current terraform.tfstate shows all entries are properly populated

**Code**: Import functions use `entryToObjectValue()` to properly convert entries to ListValue for state storage (lines 504-510 in `access_list_ipv6/resource.go`)

#### 4. Terraform State Attribute Naming Update
**Problem**: State file used old attribute names (`filter_ids`, `dynamic_filter_ids`) while schema uses new names (`sequences`, `dynamic_sequences`).

**Solution**: Updated terraform.tfstate to use new attribute names:
- `filter_ids` → `sequences`
- `dynamic_filter_ids` → `dynamic_sequences`

#### 5. Example Configuration Update
**File**: `examples/import/main.tf`

Updated MAC ACL apply blocks:
```hcl
# Before
apply {
  filter_ids = [1, 100]
}

# After
apply {
  sequences = [1, 100]
}
```

### Test Results
All tests pass:
```
ok  github.com/sh1/terraform-provider-rtx/internal/rtx/parsers  0.305s
```

### Summary of Validation Fixes

| Issue | Status | Resolution |
|-------|--------|------------|
| NAT Masquerade "primary" | ✅ Fixed | Added "primary"/"secondary" to valid values |
| Admin User password | ✅ Already Working | Code already uses password-less validation for updates |
| ACL import entries | ✅ Working | Static ACLs support sequence import, dynamic ACLs by design |

### Files Modified
- `internal/rtx/parsers/nat_masquerade.go`
- `internal/rtx/parsers/nat_masquerade_test.go`
- `examples/import/main.tf`
- `examples/import/terraform.tfstate`

---

## Session: 2026-02-01 - NAT Masquerade Static Entry Command Fix

### Objective
Fix NAT masquerade static entry command format when `outside_global = "ipcp"`.

### Problem
When `outside_global = "ipcp"`, the command was generated as:
```
nat descriptor masquerade static 1000 2 ipcp:500=192.168.1.253:500 udp
```
This format is **not valid** on RTX routers and produces:
```
エラー: IPアドレスが認識できません
```

### Solution
RTX routers have two different formats for static masquerade entries:

| Scenario | Format |
|----------|--------|
| Specific outer IP | `nat descriptor masquerade static <id> <entry> <outer:port>=<inner:port> [protocol]` |
| ipcp (PPPoE dynamic) | `nat descriptor masquerade static <id> <entry> <inner_ip> <protocol> <port>` |
| ipcp (different ports) | `nat descriptor masquerade static <id> <entry> <inner_ip> <protocol> <outer_port>=<inner_port>` |
| Protocol-only (ESP, AH, GRE) | `nat descriptor masquerade static <id> <entry> <inner_ip> <protocol>` |

Updated `BuildNATMasqueradeStaticCommand` in `internal/rtx/parsers/nat_masquerade.go` to:
1. Check if `OutsideGlobal` is "ipcp" or empty
2. If yes, generate the alternate format: `<inner_ip> <protocol> <port>`
3. If ports differ, use: `<inner_ip> <protocol> <outer_port>=<inner_port>`
4. If specific IP, use original format: `<outer:port>=<inner:port> [protocol]`

### Test Cases Updated

| Entry | Expected Command |
|-------|------------------|
| inside=192.168.1.253, outside=ipcp, port=500, protocol=udp | `nat descriptor masquerade static 1000 2 192.168.1.253 udp 500` |
| inside=192.168.1.100, outside=ipcp, outer=8080, inner=80, protocol=tcp | `nat descriptor masquerade static 1000 3 192.168.1.100 tcp 8080=80` |
| inside=192.168.1.253, protocol=esp (no ports) | `nat descriptor masquerade static 1000 1 192.168.1.253 esp` |
| inside=192.168.1.100, outside=203.0.113.1, port=80, protocol=tcp | `nat descriptor masquerade static 1000 4 203.0.113.1:80=192.168.1.100:80 tcp` |

### Files Modified
- `internal/rtx/parsers/nat_masquerade.go` - Fixed command generation for ipcp
- `internal/rtx/parsers/nat_masquerade_test.go` - Updated test cases for new format

### Test Results
All NAT tests pass (100+ test cases).

---

## Session: 2026-02-01 - Master Spec File Path Sync

### Objective
Update all master specs to reflect the new modular file structure (`resources/{name}/resource.go + model.go`).

### Background
All 49 resources have been migrated from the old `resource_rtx_*.go` pattern to the new modular structure:
```
internal/provider/resources/{name}/
├── resource.go  # Resource CRUD implementation
├── model.go     # Data model with ToClient/FromClient
└── resource_test.go  # Tests
```

### Updated Master Specs

| Spec | Key Changes |
|------|-------------|
| **routing** | Updated bgp/, ospf/, static_route/ paths; Fixed architecture diagram |
| **vpn** | Updated ipsec_tunnel/, ipsec_transport/, l2tp/, l2tp_service/, pptp/ paths; tunnel/ already correct |
| **schedule** | Updated kron_schedule/, kron_policy/ paths; Updated to Plugin Framework interfaces |
| **filter** | Updated to access_list_ip_apply/, access_list_mac_apply/, access_list_ip_dynamic/, access_list_ipv6_dynamic/ |
| **system** | Updated system/ path; Updated to Plugin Framework; Fixed datasources/ path |
| **ipv6** | Updated ipv6_prefix/ path; Updated to Plugin Framework interfaces |

### Pattern Applied
For each spec:
1. Changed "Resource File" → "Resource Directory" in Resource Summary
2. Updated architecture mermaid diagrams with new file names
3. Updated File Structure section with modular directories
4. Updated component interfaces to show Plugin Framework pattern
5. Added Change History entry: "2026-02-01 | Structure Sync | Updated file paths..."

### Files Modified
- `.spec-workflow/master-specs/routing/design.md`
- `.spec-workflow/master-specs/vpn/design.md`
- `.spec-workflow/master-specs/schedule/design.md`
- `.spec-workflow/master-specs/filter/design.md`
- `.spec-workflow/master-specs/system/design.md`
- `.spec-workflow/master-specs/ipv6/design.md`

### Previously Updated Specs (in earlier session)
- access-list/design.md
- admin/design.md
- nat/design.md
- dhcp/design.md
- dns/design.md
- ppp/design.md
- qos/design.md
- interface/design.md
- management/design.md

### Verification
All 16 master specs now reference the correct modular file structure.
