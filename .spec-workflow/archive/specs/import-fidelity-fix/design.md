# Design Document: Import Fidelity Fix

## Overview

This design addresses critical import fidelity bugs where Terraform import does not accurately capture the complete router configuration. The bugs affect multiple resources including static routes, DHCP scopes, admin users, NAT masquerade, and DNS server select. The fix requires targeted improvements to parsers, services, and resource import handlers.

## Steering Document Alignment

### Technical Standards (tech.md)

- **Parser Accuracy**: Each parser modification must produce output that re-serializes to equivalent RTX commands
- **Parser Registry Pattern**: All parser changes follow the established pattern in `internal/rtx/parsers/`
- **Test Coverage**: Each fix includes unit tests with real configuration samples from RTX routers
- **Error Handling**: Parser failures produce clear error messages with line-level detail

### Project Structure (structure.md)

- **Parsers**: Modifications in `internal/rtx/parsers/*.go`
- **Services**: Modifications in `internal/client/*_service.go`
- **Resources**: Modifications in `internal/provider/resource_rtx_*.go`
- **Tests**: Corresponding `*_test.go` files with testdata fixtures

## Code Reuse Analysis

### Existing Components to Leverage

- **`parsers.StaticRouteParser`**: Already supports multiple NextHops; verify `show config | grep` command captures all gateway lines
- **`parsers.DHCPScopeParser`**: Existing parser structure supports scope options and exclude ranges
- **`parsers.adminParser`**: `parseUserAttributeString()` correctly parses `login-timer=N`
- **`parsers.ParseNATMasqueradeConfig`**: Has `staticPattern` regex for static entries
- **`parsers.DNSParser`**: `parseDNSServerSelectFields()` handles `edns=on` flag

### Integration Points

- **Config Service**: Uses `show config` command for full configuration retrieval
- **Individual Service GetXxx**: Uses filtered `show config | grep` for specific resources
- **Resource Import**: Calls service layer Get methods to populate Terraform state

## Architecture

### Issue Analysis and Root Causes

```mermaid
graph TD
    subgraph "Import Flow"
        A[terraform import] --> B[resourceRtxXxxImport]
        B --> C[client.GetXxx]
        C --> D[service.GetXxx]
        D --> E[executor.Run]
        E --> F[show config | grep ...]
        F --> G[parser.ParseXxx]
        G --> H[Return to Terraform]
    end

    subgraph "Identified Issues"
        I1[REQ-1: Static Route grep pattern]
        I2[REQ-2: DHCP Scope not in importer list]
        I3[REQ-3: Admin default value handling]
        I4[REQ-4: NAT static entries not populated]
        I5[REQ-5: DNS edns flag in wrong field]
    end

    F -.->|may miss lines| I1
    B -.->|not implemented| I2
    G -.->|default vs zero| I3
    D -.->|incomplete query| I4
    G -.->|parsing order| I5
```

### Modular Design Principles

- **Single Issue per Fix**: Each requirement addressed independently to minimize regression risk
- **Test-First Approach**: Add failing tests before implementing fixes
- **Backward Compatibility**: Existing working imports must continue functioning

## Components and Interfaces

### Component 1: Static Route Multi-Gateway Fix (REQ-1)

- **Purpose**: Ensure all gateways for a route are captured during import
- **Files**: `internal/rtx/parsers/static_route.go`, `internal/client/static_route_service.go`
- **Issue**: The `show config | grep "ip route <network>"` may not capture all gateway lines if formatting differs
- **Fix Strategy**:
  1. Verify grep pattern matches all gateway variations
  2. Add test cases with multi-gateway configurations
  3. Ensure parser groups all gateways under same prefix/mask key
- **Dependencies**: Existing `ParseRouteConfig` already handles multi-gateway grouping

### Component 2: DHCP Scope Import (REQ-2)

- **Purpose**: Enable import of existing DHCP scope configurations
- **Files**:
  - `internal/provider/resource_rtx_dhcp_scope.go`
  - `internal/client/dhcp_scope_service.go`
  - `internal/rtx/parsers/dhcp_scope.go`
- **Issue**: DHCP scope resources may not have import function or may not parse all attributes
- **Fix Strategy**:
  1. Verify `resourceRtxDhcpScopeImport` exists and is registered
  2. Ensure `DHCPScopeParser.ParseSingleScope` captures all config lines (`dhcp scope N`, `dhcp scope option N`, `dhcp scope N except`)
  3. Add test with complete scope configuration
- **Interfaces**: `GetScope(ctx, scopeID)` returns `*DHCPScope`

### Component 3: Admin User Attributes (REQ-3)

- **Purpose**: Correctly import login_timer and other user attributes
- **Files**:
  - `internal/rtx/parsers/admin.go`
  - `internal/client/admin_service.go`
  - `internal/provider/resource_rtx_admin_user.go`
- **Issue**: `login_timer=0` in Terraform when router has `login-timer=3600`
- **Fix Strategy**:
  1. Verify `parseUserAttributeString` correctly parses `login-timer=N`
  2. Check if default value handling causes zero override
  3. Ensure schema uses proper default vs computed
- **Reuses**: `UserAttributes.LoginTimer` field already exists

### Component 4: NAT Masquerade Static Entries (REQ-4)

- **Purpose**: Import static NAT port mappings within masquerade descriptors
- **Files**:
  - `internal/rtx/parsers/nat_masquerade.go`
  - `internal/client/nat_masquerade_service.go`
  - `internal/provider/resource_rtx_nat_masquerade.go`
- **Issue**: `static_entries` list not populated during import
- **Fix Strategy**:
  1. Verify `ParseNATMasqueradeConfig` captures `nat descriptor masquerade static N M ...` lines
  2. Ensure service `Get` method queries with pattern that includes static lines
  3. Verify resource maps `StaticEntries` to Terraform schema
- **Interfaces**: `MasqueradeStaticEntry` struct with `EntryNumber`, `InsideLocal`, `Protocol`, etc.

### Component 5: DNS Server Select EDNS Parsing (REQ-5)

- **Purpose**: Correctly parse EDNS flag without contaminating domain list
- **Files**:
  - `internal/rtx/parsers/dns.go`
  - `internal/client/dns_service.go`
  - `internal/provider/resource_rtx_dns_server.go`
- **Issue**: `edns=on` appearing in domains list instead of separate boolean
- **Fix Strategy**:
  1. Verify `parseDNSServerSelectFields` extracts `edns=on` before processing domains
  2. Confirm field parsing order: servers → edns → record_type → query_pattern
  3. Add test case with `edns=on` in various positions
- **Reuses**: `DNSServerSelect.EDNS` boolean field exists

## Data Models

### StaticRoute (existing)

```go
type StaticRoute struct {
    Prefix   string    `json:"prefix"`
    Mask     string    `json:"mask"`
    NextHops []NextHop `json:"next_hops"`  // All gateways captured here
}

type NextHop struct {
    NextHop   string `json:"next_hop,omitempty"`
    Interface string `json:"interface,omitempty"`
    Distance  int    `json:"distance"`
    Permanent bool   `json:"permanent"`
    Filter    int    `json:"filter,omitempty"`
}
```

### DHCPScope (existing)

```go
type DHCPScope struct {
    ScopeID       int              `json:"scope_id"`
    Network       string           `json:"network"`
    LeaseTime     string           `json:"lease_time,omitempty"`
    ExcludeRanges []ExcludeRange   `json:"exclude_ranges,omitempty"`
    Options       DHCPScopeOptions `json:"options,omitempty"`
}
```

### NATMasquerade (verify static_entries populated)

```go
type NATMasquerade struct {
    DescriptorID  int                     `json:"descriptor_id"`
    OuterAddress  string                  `json:"outer_address"`
    InnerNetwork  string                  `json:"inner_network"`
    StaticEntries []MasqueradeStaticEntry `json:"static_entries,omitempty"`  // Must be populated
}
```

### DNSServerSelect (verify EDNS separate from domains)

```go
type DNSServerSelect struct {
    ID           int      `json:"id"`
    Servers      []string `json:"servers"`
    EDNS         bool     `json:"edns"`           // Must not appear in QueryPattern
    QueryPattern string   `json:"query_pattern"`  // Domain pattern only
}
```

## Error Handling

### Error Scenarios

1. **Multi-gateway grep misses lines**
   - **Handling**: Expand grep pattern or use broader config query
   - **User Impact**: All gateways will be captured; no silent data loss

2. **DHCP scope lines incomplete**
   - **Handling**: Query all related lines (`dhcp scope N`, `dhcp scope option N`, `dhcp scope N except`)
   - **User Impact**: Complete scope configuration in Terraform

3. **Zero value vs default value confusion**
   - **Handling**: Use `Computed: true` for values with router defaults
   - **User Impact**: Correct values shown; no perpetual diff

4. **Parser regex mismatch**
   - **Handling**: Add support for additional RTX CLI output variations
   - **User Impact**: Clear error message if parsing fails

## Testing Strategy

### Unit Testing

- Add test cases to each parser's `*_test.go` with real router config samples
- Test multi-line configurations for grouped resources
- Test edge cases: empty values, maximum field counts, special characters

### Test Fixtures Required

```
internal/rtx/testdata/
├── static_route_multi_gateway.txt      # REQ-1
├── dhcp_scope_complete.txt             # REQ-2
├── admin_user_with_timer.txt           # REQ-3
├── nat_masquerade_with_static.txt      # REQ-4
└── dns_server_select_edns.txt          # REQ-5
```

### Integration Testing

- Verify import → read → plan produces no changes
- Test round-trip: import existing config, export to HCL, re-import

### Acceptance Testing

- Requires real RTX router or simulator
- Import actual router configurations
- Verify Terraform state matches `show config` output

## Implementation Order

1. **P0 - Critical** (Core import functionality broken):
   - REQ-1: Static Routes multi-gateway
   - REQ-2: DHCP Scope import
   - REQ-4: NAT Masquerade static entries

2. **P1 - High** (Incorrect values imported):
   - REQ-3: Admin User login_timer
   - REQ-5: DNS Server EDNS parsing
