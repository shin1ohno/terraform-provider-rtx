# Tasks: Terraform Plan Differences Fix

## Implementation Tasks - ALL COMPLETED

### Task 1: Fix L2TP Tunnel Configuration Mismatch ✅ COMPLETED

- [x] 1.1 Verify actual RTX configuration has `l2tp tunnel 1 auth on`
- [x] 1.2 Update main.tf: Set `tunnel_auth_enabled = true`
- [x] 1.3 Add `Computed: true` to `tunnel_auth_password` schema
- [x] 1.4 Run `terraform plan` to verify no changes for L2TP tunnel

### Task 2: Fix IPv6 Filter Dynamic Stub Implementation ✅ COMPLETED

- [x] 2.1 Fix stub in `internal/client/client.go` (lines 3369-3384)
- [x] 2.2 Replace stub with delegation to IPFilterService
- [x] 2.3 Also fix Create/Update/Delete stubs
- [x] 2.4 Run `terraform import rtx_ipv6_filter_dynamic.main main`
- [x] 2.5 Run `terraform plan` to verify no changes for IPv6 filter

### Task 3: Debug and Fix NAT Masquerade Import ✅ COMPLETED

- [x] 3.1 Discovered RTX doesn't support `grep -E` (extended regex)
- [x] 3.2 Changed grep pattern from `grep -E "( 1000 | 1000$)"` to `grep "nat descriptor.*1000"`
- [x] 3.3 Added `OutsideGlobal: "ipcp"` default to all parser patterns
- [x] 3.4 Removed `outside_global` from main.tf (uses schema default)
- [x] 3.5 Run `terraform import rtx_nat_masquerade.nat1000 1000`
- [x] 3.6 Run `terraform plan` to verify no changes for NAT masquerade

### Task 4: Fix DHCP Scope Parser ✅ COMPLETED

- [x] 4.1 Fix regex to handle line wrapping:
  - Changed `\s*$` to `.*$` in scopePattern, scopeExpireOnlyPattern, scopeRangePattern
  - This handles lines ending with partial text like ` ma` from `maxexpire`
- [x] 4.2 Added `calculateNetworkAddress()` function to convert IP+prefix to network CIDR
- [x] 4.3 Updated line 92 to use `calculateNetworkAddress(matches[2], matches[4])`
- [x] 4.4 Removed gateway-to-routers logic to prevent duplicates (routers from options take precedence)
- [x] 4.5 Re-import: `terraform state rm rtx_dhcp_scope.scope1 && terraform import rtx_dhcp_scope.scope1 1`
- [x] 4.6 Run `terraform plan` to verify no changes for DHCP scope

### Task 5: Final Verification ✅ COMPLETED

- [x] 5.1 Run full terraform plan: `terraform plan -parallelism=2`
- [x] 5.2 Verify output shows **"No changes. Your infrastructure matches the configuration."**
- [x] 5.3 All differences resolved
- [x] 5.4 Update SESSION_PROGRESS.md with completion status

## Dependencies

```
Task 1 ✅ → completed
Task 2 ✅ → completed
Task 3 ✅ → completed
Task 4 ✅ → completed
Task 5 ✅ → completed (depends on Tasks 1-4)
```

## Bug Summary

| Bug | File | Line | Status |
|-----|------|------|--------|
| IPv6 Filter Dynamic Stub | client.go | 3369-3384 | ✅ Fixed |
| NAT Masquerade grep pattern | nat_masquerade.go | 341 | ✅ Fixed |
| NAT Masquerade OutsideGlobal default | nat_masquerade.go | multiple | ✅ Fixed |
| DHCP Scope line wrap regex | dhcp_scope.go | 54, 57, 61 | ✅ Fixed |
| DHCP Scope network calculation | dhcp_scope.go | 92 | ✅ Fixed |
| DHCP Scope routers duplicate | dhcp_scope.go | 93-96 | ✅ Fixed |
| L2TP password Computed flag | resource_rtx_l2tp.go | 207 | ✅ Fixed |

## Final Test Results

```bash
$ terraform plan -parallelism=2
...
No changes. Your infrastructure matches the configuration.

Terraform has compared your real infrastructure against your configuration
and found no differences, so no changes are needed.
```

## Files Modified

- `internal/rtx/parsers/dhcp_scope.go` - regex patterns, calculateNetworkAddress, gateway logic
- `internal/rtx/parsers/nat_masquerade.go` - grep pattern, OutsideGlobal defaults
- `internal/client/client.go` - IPv6 Filter Dynamic stub → IPFilterService delegation
- `internal/provider/resource_rtx_l2tp.go` - tunnel_auth_password Computed flag
- `examples/import/main.tf` - tunnel_auth_enabled=true, removed outside_global
