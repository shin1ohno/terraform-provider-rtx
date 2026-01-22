# Requirements Document: Import Fidelity Fix

## Introduction

This specification addresses critical bugs in the terraform import functionality where imported resources do not accurately reflect the actual router configuration. When comparing the router's live config (`show config` output) with the Terraform-generated `main.tf`, significant discrepancies were identified. These bugs prevent users from reliably importing existing RTX router configurations into Terraform management.

## Bug Report Summary

**Discovered**: 2026-01-20
**Severity**: High
**Impact**: Import functionality produces incomplete/incorrect Terraform configurations

### Identified Discrepancies

| Category | Issue | Severity |
|----------|-------|----------|
| Static Routes | Multiple gateways not imported (only first gateway captured) | High |
| DHCP Scope | Entire resource not being imported | High |
| Admin User | `login_timer` value incorrect (0 instead of 3600) | Medium |
| NAT Masquerade | Static NAT entries not imported | High |
| DNS Server | `edns=on` flags mixed into domains list (parser issue) | Medium |
| IP Filters | Resources not imported at all | Medium |
| IPv6 Filters | Resources not imported at all | Medium |
| Ethernet Filters | Resources not imported at all | Medium |
| L2TP Service | Global service enablement not captured | Low |
| IPsec Transport | Transport mode mappings not imported | Medium |

## Alignment with Product Vision

From `product.md`:
- "Enable complete IaC management of Yamaha RTX router configurations"
- "Support enterprise deployment patterns with proper state management"
- "Import Support: Import existing configurations into Terraform state"

The current import bugs directly violate these objectives. Users cannot trust that an imported configuration accurately represents their router state.

## Requirements

### REQ-1: Static Route Multi-Gateway Import

**User Story:** As a network administrator, I want static routes with multiple gateways to be fully imported, so that my failover routing configurations are accurately represented in Terraform.

#### Acceptance Criteria

1. WHEN a static route has multiple gateways defined THEN the parser SHALL capture all gateways in the `next_hops` list
2. WHEN importing `ip route 10.0.0.0/8 gateway 192.168.1.20 gateway 192.168.1.21` THEN the resource SHALL contain both gateways with appropriate distance values
3. IF a route has N gateways THEN the imported resource SHALL have exactly N entries in `next_hops`

### REQ-2: DHCP Scope Import

**User Story:** As a network administrator, I want DHCP scope configurations to be imported, so that my IP address management is fully represented in Terraform.

#### Acceptance Criteria

1. WHEN `dhcp scope N` is defined in router config THEN a `rtx_dhcp_scope` resource SHALL be created
2. WHEN scope has static bindings (`dhcp scope bind`) THEN they SHALL be included in the resource
3. WHEN scope has options (`dhcp scope option`) THEN they SHALL be included in the resource
4. WHEN scope defines gateway, expire, and maxexpire THEN these SHALL be captured accurately

### REQ-3: Admin User Attribute Import

**User Story:** As an administrator, I want user attributes to be imported correctly, so that access control settings are accurately maintained.

#### Acceptance Criteria

1. WHEN `login-timer=N` is set for a user THEN `login_timer` attribute SHALL equal N
2. WHEN `login-timer` is not explicitly set THEN the default value SHALL be used (not 0)
3. WHEN user has `connection=serial,telnet,remote,ssh,sftp,http` THEN all connection methods SHALL be imported

### REQ-4: NAT Masquerade Static Entries Import

**User Story:** As a network administrator, I want NAT masquerade static entries to be imported, so that port forwarding and protocol mappings are preserved.

#### Acceptance Criteria

1. WHEN `nat descriptor masquerade static N M addr protocol` is defined THEN it SHALL be included in the resource
2. WHEN multiple static entries exist THEN all entries SHALL be imported as a list
3. WHEN static entry specifies protocol (esp, udp, tcp) and port THEN these SHALL be captured

### REQ-5: DNS Server Select Parsing

**User Story:** As a network administrator, I want DNS server select statements to be parsed correctly, so that DNS forwarding rules work properly.

#### Acceptance Criteria

1. WHEN `dns server select N server edns=on domain` is parsed THEN `edns=on` SHALL NOT appear in domains list
2. WHEN server has edns flag THEN it SHALL be captured as a separate boolean attribute
3. WHEN multiple servers are specified THEN each server's edns flag SHALL be individually captured

### REQ-6: IP Filter Import (Future)

**User Story:** As a network administrator, I want IP filter rules to be importable, so that firewall configurations can be managed via Terraform.

#### Acceptance Criteria

1. WHEN `ip filter N action src dst proto sport dport` is defined THEN a filter resource SHALL be creatable
2. IF filter references are used in interfaces THEN they SHALL be resolvable

**Note:** This may require new resource types. Mark as future enhancement if scope is too large.

### REQ-7: Ethernet Filter Import (Future)

**User Story:** As a network administrator, I want Ethernet filter rules to be importable, so that Layer 2 filtering can be managed.

#### Acceptance Criteria

1. WHEN `ethernet filter N action mac-src mac-dst` is defined THEN a filter resource SHALL be creatable
2. WHEN `ethernet lanN filter in/out` is defined THEN interface filter bindings SHALL be captured

**Note:** May require new resource types. Mark as future enhancement if scope is too large.

## Non-Functional Requirements

### Code Architecture and Modularity

- **Parser Accuracy**: Each parser must produce output that, when re-serialized to RTX commands, produces equivalent configuration
- **Round-Trip Fidelity**: Import → Export → Import should produce identical Terraform state
- **Test Coverage**: Each parser fix must include unit tests with real config samples

### Performance

- Import of full router configuration should complete within 60 seconds

### Reliability

- Parser failures should produce clear error messages indicating which line failed to parse
- Partial import failures should not corrupt already-imported resources

### Usability

- Imported resources should require minimal manual adjustment before being usable
- Documentation should clearly indicate any known import limitations

## Out of Scope (Phase 1)

The following items are documented but deferred:

- IPv6 filter import (requires new resource type)
- IP filter import (requires new resource type)
- Ethernet filter import (requires new resource type)
- L2TP service global settings
- IPsec transport mode import

These will be addressed in a future specification after core import fidelity is restored.

## Priority Order

1. **P0 - Critical**: Static Routes, DHCP Scope, NAT Masquerade Static Entries
2. **P1 - High**: Admin User attributes, DNS Server parsing
3. **P2 - Medium**: Filter resources (new resource types - future spec)
