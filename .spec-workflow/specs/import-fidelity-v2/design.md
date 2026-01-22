# Design Document: Import Fidelity Fix v2

## Overview

This design addresses four import fidelity issues in the RTX Terraform provider parsers and resource schemas.

## Architecture Decisions

### AD-1: DHCP Scope IP Range Format

**Current Behavior**:
- Parser only captures network CIDR (e.g., `192.168.0.0/16`)
- IP range format (`192.168.1.20-192.168.1.99/16`) is partially parsed

**Proposed Solution**:
- Add `range_start` and `range_end` fields to `DHCPScope` struct
- Update `scopePattern` regex to capture IP range format
- Modify resource schema to include range fields

**Files to Modify**:
- `internal/rtx/parsers/dhcp_scope.go`: Add range parsing
- `internal/provider/resource_rtx_dhcp_scope.go`: Update schema

### AD-2: User Attribute Parsing Enhancement

**Current Behavior**:
- `ParseUserAttributeString` in `admin.go` parses some attributes
- `login-timer` and `gui-page` values may not be fully extracted

**Root Cause Analysis**:
Looking at config.txt line 9:
```
user attribute shin1ohno connection=serial,telnet,remote,ssh,sftp,http gui-page=dashboard,lan-map,config login-timer=3600
```

The parser needs to handle:
1. `login-timer=3600` (integer value)
2. `gui-page=dashboard,lan-map,config` (comma-separated list)

**Proposed Solution**:
- Verify/fix `ParseUserAttributeString` function
- Ensure regex patterns capture all attribute formats
- Test with actual config.txt examples

**Files to Modify**:
- `internal/rtx/parsers/admin.go`: Fix attribute parsing
- `internal/rtx/parsers/admin_test.go`: Add test cases

### AD-3: NAT Masquerade Protocol-Only Entry

**Current Behavior**:
- NAT static entries require port fields
- Protocol-only entries (ESP, AH, GRE, ICMP) are not supported

**Proposed Solution**:
- Update `MasqueradeStaticEntry` struct to allow optional ports
- Add `IsProtocolOnly` field or detect from protocol type
- Modify parser to handle both formats:
  - Port-based: `ipcp:80=192.168.1.10:8080 tcp`
  - Protocol-only: `192.168.1.253 esp`

**Command Format Analysis**:
```
# Port-based NAT static entry
nat descriptor masquerade static 1000 2 192.168.1.253 udp 500

# Protocol-only NAT static entry
nat descriptor masquerade static 1000 1 192.168.1.253 esp
```

**Files to Modify**:
- `internal/rtx/parsers/nat_masquerade.go`: Add protocol-only parsing
- `internal/client/nat_masquerade_service.go`: Handle protocol-only entries
- `internal/provider/resource_rtx_nat_masquerade.go`: Update schema

## Data Model Changes

### DHCPScope Struct Enhancement

```go
type DHCPScope struct {
    ScopeID     int
    Network     string  // CIDR format: 192.168.0.0/16
    RangeStart  string  // Optional: 192.168.1.20
    RangeEnd    string  // Optional: 192.168.1.99
    Gateway     string
    LeaseTime   string
    MaxExpire   string
    // ... existing fields
}
```

### MasqueradeStaticEntry Enhancement

```go
type MasqueradeStaticEntry struct {
    EntryNumber       int
    InsideLocal       string
    InsideLocalPort   *int    // nil for protocol-only
    OutsideGlobal     string
    OutsideGlobalPort *int    // nil for protocol-only
    Protocol          string
    IsProtocolOnly    bool    // true for esp, ah, gre, icmp
}
```

## Implementation Approach

1. **Phase 1**: Fix DHCP scope IP range parsing
2. **Phase 2**: Fix user attribute parsing
3. **Phase 3**: Add NAT protocol-only entry support

Each phase includes:
- Parser modification
- Unit test addition
- Round-trip verification

## Testing Strategy

### Unit Tests

- `TestParseDHCPScopeIPRange`: Test IP range format parsing
- `TestParseUserAttributeWithLoginTimer`: Test login-timer extraction
- `TestParseUserAttributeWithGuiPage`: Test gui-page extraction
- `TestParseNATMasqueradeProtocolOnly`: Test ESP/AH entry parsing

### Integration Tests

- Verify terraform plan shows no diff after import
- Test with actual config.txt from RTX1210-Ebisu

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Breaking existing DHCP scope configs | Maintain backward compatibility with network-only format |
| User attribute parsing regression | Add comprehensive test cases from real configs |
| NAT schema change breaks existing state | Optional fields with nil defaults |
