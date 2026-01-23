# Design Document: Terraform Plan Differences Fix

## Overview

This document provides multi-angle analysis of the 4 terraform plan differences and proposes solutions for each, including actual RTX responses and implementation details.

---

## Difference 1: rtx_dhcp_scope.scope1 - Network Field is Null

### Problem Analysis

**Symptom:** State shows `network = ""` (empty string) even though RTX has `dhcp scope 1 192.168.1.20-192.168.1.99/16`

### Current State (from terraform.tfstate)
```
scope_id: 1
network: ""
range_start: ""
range_end: ""
lease_time: ""
options.routers: ['192.168.1.253']
```

### Actual RTX Response (from TF_LOG=DEBUG)
```
DHCP scope raw output: " show config | grep \"dhcp scope\"\r\n
Searching ...\r\n
dhcp scope 1 192.168.1.20-192.168.1.99/16 gateway 192.168.1.253 expire 12:00 ma\r\n
xexpire 24:00\r\n
dhcp scope bind 1 192.168.1.20 01 00 30 93 11 0e 33\r\n
dhcp scope bind 1 192.168.1.21 01 00 3e e1 c3 54 b4\r\n
dhcp scope bind 1 192.168.1.22 01 00 3e e1 c3 54 b5\r\n
dhcp scope bind 1 192.168.1.28 24:59:e5:54:5e:5a\r\n
dhcp scope bind 1 192.168.1.29 b8:0b:da:ef:77:33\r\n
dhcp scope option 1 router=192.168.1.253\r\n
[RTX1210] #"
```

### Root Cause Analysis

**Bug #1: Terminal Line Wrapping**

The RTX terminal wraps the output, splitting `maxexpire` into `ma` + `xexpire`:
```
dhcp scope 1 192.168.1.20-192.168.1.99/16 gateway 192.168.1.253 expire 12:00 ma
xexpire 24:00
```

The parser regex in `internal/rtx/parsers/dhcp_scope.go` (line 58) expects a single line:
```go
scopeRangePattern := regexp.MustCompile(
    `^\s*dhcp\s+scope\s+(\d+)\s+([0-9.]+)-([0-9.]+)/(\d+)` +
    `(?:\s+gateway\s+([0-9.]+))?(?:\s+expire\s+(\S+))?\s*$`
)
```

This regex ends with `$` and doesn't match because the line continues with ` maxexpire 24:00`.

**Bug #2: Network Calculation**

Even if line wrapping is fixed, line 89 has another bug:
```go
scope.Network = matches[2] + "/" + matches[4] // Store as start_ip/mask
```

This stores `192.168.1.20/16` instead of calculating the correct network `192.168.0.0/16`.

### Implementation Details

#### Code Flow
```
resourceRTXDHCPScopeRead (resource_rtx_dhcp_scope.go)
  → apiClient.client.GetDHCPScope(ctx, scopeID)
    → DHCPScopeService.GetScope (dhcp_scope_service.go:91)
      → parsers.BuildShowDHCPScopeCommand(scopeID)
        → "show config | grep \"dhcp scope\""
      → parser.ParseSingleScope(string(output), scopeID)
        → ParseScopeConfig (dhcp_scope.go:44)
          → scopeRangePattern.FindStringSubmatch(line) // FAILS due to line wrap
```

#### Parser Location
- File: `internal/rtx/parsers/dhcp_scope.go`
- Function: `ParseScopeConfig` (lines 44-203)
- Regex: `scopeRangePattern` (line 58)

### Proposed Solution

**Option A (Recommended): Fix Parser**

1. **Fix Line Wrapping**: Update `preprocessWrappedLines()` in `internal/client/parser.go` to handle text continuations (not just digits)

2. **Fix Regex**: Update `scopeRangePattern` to make `maxexpire` optional:
```go
scopeRangePattern := regexp.MustCompile(
    `^\s*dhcp\s+scope\s+(\d+)\s+([0-9.]+)-([0-9.]+)/(\d+)` +
    `(?:\s+gateway\s+([0-9.]+))?(?:\s+expire\s+(\S+))?` +
    `(?:\s+maxexpire\s+(\S+))?\s*$`
)
```

3. **Fix Network Calculation**: Calculate proper network address from IP and prefix:
```go
// Instead of:
scope.Network = matches[2] + "/" + matches[4]

// Use:
network, _ := calculateNetworkAddress(matches[2], matches[4])
scope.Network = network  // Returns "192.168.0.0/16"
```

---

## Difference 2: rtx_ipv6_filter_dynamic.main - Never Imported

### Problem Analysis

**Symptom:** Resource shows "will be created" - no state exists

### Actual RTX Configuration (from TF_LOG=DEBUG)
```
ipv6 lan2 secure filter out 101099 dynamic 101080 101081 101082 101083 101084 1
01085 101098 101099
```

The IPv6 dynamic filters exist on the RTX and are referenced by the interface.

### Implementation Details

#### Import Capability Confirmed
```go
// internal/provider/resource_rtx_ipv6_filter_dynamic.go (lines 23-25)
Importer: &schema.ResourceImporter{
    StateContext: schema.ImportStatePassthroughContext,
},
```

The resource uses `ImportStatePassthroughContext`, which means:
- Import ID is directly used as the resource ID
- Read function populates the state

#### Read Function
```go
// internal/provider/resource_rtx_ipv6_filter_dynamic.go (lines 86-107)
func resourceRTXIPv6FilterDynamicRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    apiClient := meta.(*apiClient)
    config, err := apiClient.client.GetIPv6FilterDynamicConfig(ctx)
    // ...
    entries := flattenIPv6FilterDynamicEntries(config.Entries)
    d.Set("entry", entries)
    return nil
}
```

### Bug Discovery: Stub Implementation

**Import Test Result:**
```
$ terraform import rtx_ipv6_filter_dynamic.main main
Error: Failed to read IPv6 filter dynamic: IPv6 filter dynamic config not implemented
```

**Root Cause:** The `rtxClient.GetIPv6FilterDynamicConfig()` method is a stub that returns "not implemented":
```go
// internal/client/client.go (lines 3369-3372)
// IPv6 Filter Dynamic Config stub implementations
func (c *rtxClient) GetIPv6FilterDynamicConfig(ctx context.Context) (*IPv6FilterDynamicConfig, error) {
    return nil, fmt.Errorf("IPv6 filter dynamic config not implemented")
}
```

**Note:** A working implementation exists in `IPFilterService.GetIPv6FilterDynamicConfig()` (ip_filter_service.go:856), but `rtxClient` doesn't delegate to it.

### Proposed Solution

**Step 1: Fix rtxClient Implementation**

Update `internal/client/client.go` to delegate to IPFilterService:
```go
func (c *rtxClient) GetIPv6FilterDynamicConfig(ctx context.Context) (*IPv6FilterDynamicConfig, error) {
    ipFilterService := NewIPFilterService(c.executor, c)
    return ipFilterService.GetIPv6FilterDynamicConfig(ctx)
}
```

**Step 2: Run Import Command:**
```bash
terraform import rtx_ipv6_filter_dynamic.main main
```

---

## Difference 3: rtx_l2tp.tunnel1 - tunnel_auth_enabled Mismatch

### Problem Analysis

**Symptom:** State shows `tunnel_auth_enabled = true`, main.tf has `tunnel_auth_enabled = false`

### Current State (from terraform.tfstate)
```
tunnel_id: 1
tunnel_auth_enabled: True
tunnel_auth_password: (sensitive)
```

### main.tf Configuration
```hcl
resource "rtx_l2tp" "tunnel1" {
  # ...
  l2tpv3_config {
    local_router_id      = "192.168.1.253"
    remote_router_id     = "192.168.1.254"
    remote_end_id        = "shin1"
    tunnel_auth_enabled  = false  # <-- MISMATCH
  }
}
```

### Root Cause

This is a **configuration mismatch**, not a provider bug:
- The RTX router has `l2tp tunnel 1 auth on` configured
- The main.tf incorrectly specifies `tunnel_auth_enabled = false`
- The state correctly reflects the actual RTX configuration

### Proposed Solution

**Option A (Recommended): Update main.tf**

```hcl
l2tpv3_config {
  local_router_id      = "192.168.1.253"
  remote_router_id     = "192.168.1.254"
  remote_end_id        = "shin1"
  tunnel_auth_enabled  = true  # Match RTX config
  # tunnel_auth_password is sensitive - handled separately
}
```

Note: The `tunnel_auth_password` is marked as sensitive in the state and doesn't need to be in main.tf if it's already configured on the router.

**Option B: Remove auth from RTX**

Running `terraform apply` with `tunnel_auth_enabled = false` will:
- Execute `no l2tp tunnel 1 auth` on the RTX
- Remove tunnel authentication
- **Risk:** May break L2TP connectivity if remote peer expects auth

---

## Difference 4: rtx_nat_masquerade.nat1000 - Never Imported

### Problem Analysis

**Symptom:** Resource shows "will be created" - no state exists

### Actual RTX Configuration
Based on the interface configuration referencing `nat descriptor 1000`:
```
nat descriptor type 1000 masquerade
nat descriptor address outer 1000 primary
nat descriptor masquerade static 1000 1 192.168.1.253 esp
nat descriptor masquerade static 1000 2 192.168.1.253 udp 500
nat descriptor masquerade static 1000 3 192.168.1.253 udp 4500
nat descriptor masquerade static 1000 4 192.168.1.253 udp 1701
nat descriptor masquerade static 1000 900 192.168.1.20 tcp 55000
```

### Implementation Details

#### Import Capability Confirmed
```go
// internal/provider/resource_rtx_nat_masquerade.go (lines 25-27)
Importer: &schema.ResourceImporter{
    StateContext: resourceRTXNATMasqueradeImport,
},
```

The resource has a custom import function that accepts the descriptor ID.

#### Schema for Protocol-Only Entries
```go
// internal/provider/resource_rtx_nat_masquerade.go
"protocol": {
    Type:         schema.TypeString,
    Optional:     true,
    Description:  "Protocol: 'tcp', 'udp' (require ports), or 'esp', 'ah', 'gre', 'icmp' (protocol-only, no ports)",
    ValidateFunc: validation.StringInSlice([]string{"tcp", "udp", "esp", "ah", "gre", "icmp", ""}, true),
},
```

ESP protocol IS supported - port fields are optional for protocol-only entries.

### Bug Discovery: Descriptor Not Found

**Import Test Result:**
```
$ terraform import rtx_nat_masquerade.nat1000 1000
Error: failed to import NAT masquerade 1000: NAT masquerade with descriptor ID 1000 not found
```

**Root Cause Investigation:**

The `NATMasqueradeService.Get()` function:
1. Builds command: `show config | grep "nat descriptor" | grep -E "( 1000 | 1000$)"`
2. Parses output with `ParseNATMasqueradeConfig()`
3. Returns "not found" if no matching descriptor ID

Possible issues:
- The grep -E pattern may not work correctly on RTX
- The pattern `( 1000 | 1000$)` might not match RTX output format
- Parser may not be handling the output correctly

**Action Required:** Run with TF_LOG=DEBUG to capture actual RTX output

### main.tf Configuration (Updated)
```hcl
resource "rtx_nat_masquerade" "nat1000" {
  descriptor_id = 1000
  outer_address = "primary"

  # Entry 1: ESP protocol (IPsec)
  static_entry {
    entry_number   = 1
    inside_local   = "192.168.1.253"
    outside_global = "primary"
    protocol       = "esp"
  }

  static_entry {
    entry_number        = 2
    inside_local        = "192.168.1.253"
    inside_local_port   = 500
    outside_global      = "primary"
    outside_global_port = 500
    protocol            = "udp"
  }
  # ... additional entries
}
```

### Proposed Solution

**Step 1: Debug grep pattern issue**
- Run with TF_LOG=DEBUG to capture NAT descriptor output
- Verify grep -E works on RTX

**Step 2: Fix provider if needed**

**Step 3: Run Import Command:**
```bash
terraform import rtx_nat_masquerade.nat1000 1000
```

---

## Implementation Strategy

### Phase 1: Configuration Fixes (No Code Changes) ✅

**Task 1: Update main.tf for L2TP** ✅ COMPLETED
- Changed `tunnel_auth_enabled = false` to `tunnel_auth_enabled = true`
- Location: `examples/import/main.tf`

### Phase 2: Fix Provider Bugs (Code Changes Required)

**Task 2: Fix IPv6 Filter Dynamic Stub**
- Location: `internal/client/client.go` (lines 3369-3384)
- Fix: Delegate to IPFilterService implementation instead of returning stub error
- Files to modify:
  - `internal/client/client.go` - Replace stub with delegation

**Task 3: Debug NAT Masquerade Import**
- Run `TF_LOG=DEBUG terraform import rtx_nat_masquerade.nat1000 1000`
- Capture actual RTX output for `show config | grep "nat descriptor" | grep -E "( 1000 | 1000$)"`
- Fix grep pattern or parser as needed
- Files to potentially modify:
  - `internal/rtx/parsers/nat_masquerade.go` - Fix BuildShowNATDescriptorCommand()

**Task 4: Fix DHCP Scope Parser**
1. Update regex to handle `maxexpire` parameter
2. Add network address calculation function
3. Handle terminal line wrapping for text continuations
- Files to modify:
  - `internal/rtx/parsers/dhcp_scope.go`
  - `internal/client/parser.go` (if line wrapping fix needed)

### Phase 3: Run Imports (After Bug Fixes)

**Task 5: Import IPv6 Dynamic Filter**
```bash
cd examples/import
terraform import rtx_ipv6_filter_dynamic.main main
```

**Task 6: Import NAT Masquerade**
```bash
terraform import rtx_nat_masquerade.nat1000 1000
```

### Phase 4: Verification

**Task 7: Final Verification**
```bash
terraform plan -parallelism=2
# Expected: "No changes. Your infrastructure matches the configuration."
```

---

## Summary of Discovered Bugs

| Bug | Location | Symptom | Fix Required |
|-----|----------|---------|--------------|
| IPv6 Filter Dynamic Stub | client.go:3369-3384 | "not implemented" error | Delegate to IPFilterService |
| NAT Masquerade Import | nat_masquerade.go:332 | "descriptor not found" | Debug grep pattern |
| DHCP Scope Line Wrap | dhcp_scope.go:58 | Regex fails on wrapped lines | Handle maxexpire |
| DHCP Scope Network | dhcp_scope.go:89 | Wrong network address | Calculate from IP+prefix |

---

## Risk Assessment

| Fix | Risk Level | Impact | Mitigation |
|-----|------------|--------|------------|
| L2TP main.tf Update | None | Config only | Already completed |
| IPv6 Filter Dynamic Stub | Low | Code change | Simple delegation |
| NAT Import Debug | Medium | May need grep fix | TF_LOG debug first |
| DHCP Scope Parser Fix | Medium | Code change | Unit tests required |

---

## Code References

### DHCP Scope Parsing
- Parser: `internal/rtx/parsers/dhcp_scope.go:44-203`
- Service: `internal/client/dhcp_scope_service.go:91-111`
- Resource: `internal/provider/resource_rtx_dhcp_scope.go`

### IPv6 Filter Dynamic
- Resource: `internal/provider/resource_rtx_ipv6_filter_dynamic.go`
- Importer: Line 23-25 (ImportStatePassthroughContext)
- Read: Lines 86-107

### NAT Masquerade
- Resource: `internal/provider/resource_rtx_nat_masquerade.go`
- Importer: Line 25-27 (resourceRTXNATMasqueradeImport)
- ESP support: Schema at line 72-76

### L2TP
- Resource: `internal/provider/resource_rtx_l2tp.go`
- State: tunnel_auth_enabled correctly reflects RTX config

---

## Testing Strategy

1. **Unit Tests:** Add tests for DHCP scope parsing with `maxexpire`
2. **Import Tests:** Verify import commands work for IPv6/NAT
3. **Plan Tests:** Verify clean terraform plan after all fixes
4. **Round-Trip Test:** Import → Plan → No changes
