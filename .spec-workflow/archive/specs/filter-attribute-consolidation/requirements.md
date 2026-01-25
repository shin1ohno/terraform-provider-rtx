# Requirements: Filter Attribute Consolidation

## Overview

Consolidate and simplify filter management by:
1. Grouping all filter rules into access list resources (including dynamic filters)
2. Referencing access lists by name from `rtx_interface`
3. Removing redundant ACL binding resources

## Problem Statement

### Current State

**Filter definition resources:**
- `rtx_access_list_ip` - static IP filter rules (grouped by name)
- `rtx_access_list_ipv6` - static IPv6 filter rules (grouped by name)
- `rtx_access_list_mac` - MAC filter rules (grouped by name)
- `rtx_ip_filter_dynamic` - **individual** dynamic filter (one resource per rule)
- `rtx_ipv6_filter_dynamic` - **singleton** IPv6 dynamic filters

**Filter binding resources:**
- `rtx_interface` - has filter number attributes (`secure_filter_in`, `ethernet_filter_in`, etc.)
- `rtx_interface_acl` - binds IP/IPv6 ACLs to interfaces
- `rtx_interface_mac_acl` - binds MAC ACLs to interfaces

### Issues

1. **Inconsistent patterns**: Static filters use grouped access lists, dynamic filters are individual resources
2. **Multiple binding methods**: Filter numbers in `rtx_interface` vs named ACLs in `rtx_interface_acl`
3. **Resource fragmentation**: Three resources for interface filter binding
4. **Manual sequence management**: Users must manage filter numbers across resources

### Impact

- User confusion about which approach to use
- Inconsistent configuration patterns
- Error-prone manual filter number management
- Maintenance burden across multiple resources

## Requirements

### Requirement 1: Unified Access List Pattern for Dynamic Filters

**User Story:** As a Terraform user, I want dynamic filters to follow the same pattern as static filters, so that my configuration is consistent and predictable.

#### Acceptance Criteria

1. WHEN defining dynamic IP filters THEN users SHALL create `rtx_access_list_ip_dynamic` resource with multiple entries
2. WHEN defining dynamic IPv6 filters THEN users SHALL create `rtx_access_list_ipv6_dynamic` resource with multiple entries
3. WHEN entries are defined THEN each entry SHALL have a sequence number for ordering
4. WHEN access list is referenced THEN it SHALL be referenced by name

### Requirement 2: Simplified Interface Filter Binding

**User Story:** As a Terraform user, I want to bind filters to interfaces using simple name references, so that I don't need to manage filter numbers.

#### Acceptance Criteria

1. WHEN binding filters to interface THEN users SHALL reference access lists by name in `rtx_interface` resource
2. WHEN multiple filter types are needed THEN all SHALL be specified as attributes of `rtx_interface`
3. WHEN filter order matters THEN it SHALL be determined by sequence numbers within the access list

### Requirement 3: Remove Redundant Resources

**User Story:** As a Terraform user, I want fewer resources to manage, so that my configuration is simpler.

#### Acceptance Criteria

1. WHEN upgrading provider THEN `rtx_interface_acl` resource SHALL be removed
2. WHEN upgrading provider THEN `rtx_interface_mac_acl` resource SHALL be removed
3. WHEN upgrading provider THEN `rtx_ip_filter_dynamic` resource SHALL be removed (replaced by `rtx_access_list_ip_dynamic`)
4. WHEN upgrading provider THEN `rtx_ipv6_filter_dynamic` resource SHALL be removed (replaced by `rtx_access_list_ipv6_dynamic`)

### Requirement 4: Update rtx_interface Attributes

**User Story:** As a Terraform user, I want all filter bindings in the interface resource, so that configuration is centralized.

#### Acceptance Criteria

1. WHEN binding IP filters THEN `rtx_interface` SHALL support `access_list_ip_in` and `access_list_ip_out` attributes
2. WHEN binding IPv6 filters THEN `rtx_interface` SHALL support `access_list_ipv6_in` and `access_list_ipv6_out` attributes
3. WHEN binding dynamic IP filters THEN `rtx_interface` SHALL support `access_list_ip_dynamic_in` and `access_list_ip_dynamic_out` attributes
4. WHEN binding dynamic IPv6 filters THEN `rtx_interface` SHALL support `access_list_ipv6_dynamic_in` and `access_list_ipv6_dynamic_out` attributes
5. WHEN binding MAC filters THEN `rtx_interface` SHALL support `access_list_mac_in` and `access_list_mac_out` attributes
6. WHEN old filter attributes are used THEN they SHALL be removed (`secure_filter_*`, `ethernet_filter_*`, `dynamic_filter_*`)

### Requirement 5: Migration Documentation

**User Story:** As an existing user, I want clear migration documentation, so that I can update my configurations smoothly.

#### Acceptance Criteria

1. WHEN breaking changes are released THEN documentation SHALL include migration examples
2. WHEN resources are removed THEN CHANGELOG SHALL clearly document the changes
3. WHEN users need to migrate THEN examples SHALL show before/after configurations

## Proposed Resource Structure

### New Resources

| Resource | Purpose |
|----------|---------|
| `rtx_access_list_ip_dynamic` | Group dynamic IP filter rules |
| `rtx_access_list_ipv6_dynamic` | Group dynamic IPv6 filter rules |

### Modified Resources

| Resource | Changes |
|----------|---------|
| `rtx_interface` | Add `access_list_*` attributes, remove `secure_filter_*`, `ethernet_filter_*`, `dynamic_filter_*` |
| `rtx_access_list_ip` | No changes (already supports grouped entries) |
| `rtx_access_list_ipv6` | No changes |
| `rtx_access_list_mac` | No changes |

### Removed Resources (Including All Documentation)

| Resource | Files to Delete | Replacement |
|----------|-----------------|-------------|
| `rtx_interface_acl` | `resource_rtx_interface_acl.go`, `resource_rtx_interface_acl_test.go`, `docs/resources/interface_acl.md` | Use `rtx_interface` attributes |
| `rtx_interface_mac_acl` | `resource_rtx_interface_mac_acl.go`, `resource_rtx_interface_mac_acl_test.go`, `docs/resources/interface_mac_acl.md` | Use `rtx_interface` attributes |
| `rtx_ip_filter_dynamic` | `resource_rtx_ip_filter_dynamic.go`, `resource_rtx_ip_filter_dynamic_test.go`, `docs/resources/ip_filter_dynamic.md` | Use `rtx_access_list_ip_dynamic` |
| `rtx_ipv6_filter_dynamic` | `resource_rtx_ipv6_filter_dynamic.go`, `resource_rtx_ipv6_filter_dynamic_test.go`, `docs/resources/ipv6_filter_dynamic.md` | Use `rtx_access_list_ipv6_dynamic` |

## rtx_interface Attribute Changes

### Removed Attributes

- `secure_filter_in`
- `secure_filter_out`
- `dynamic_filter_out`
- `ethernet_filter_in`
- `ethernet_filter_out`

### New Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `access_list_ip_in` | String | Inbound IP access list name |
| `access_list_ip_out` | String | Outbound IP access list name |
| `access_list_ip_dynamic_in` | String | Inbound dynamic IP access list name |
| `access_list_ip_dynamic_out` | String | Outbound dynamic IP access list name |
| `access_list_ipv6_in` | String | Inbound IPv6 access list name |
| `access_list_ipv6_out` | String | Outbound IPv6 access list name |
| `access_list_ipv6_dynamic_in` | String | Inbound dynamic IPv6 access list name |
| `access_list_ipv6_dynamic_out` | String | Outbound dynamic IPv6 access list name |
| `access_list_mac_in` | String | Inbound MAC access list name |
| `access_list_mac_out` | String | Outbound MAC access list name |

## Non-Functional Requirements

### Backward Compatibility

- This is a **breaking change** requiring major version bump
- Clear deprecation warnings and migration guide required

### Documentation

- Update all affected resource documentation
- Provide migration examples
- Update master specs

## Out of Scope

- Changes to underlying RTX commands
- Changes to `rtx_access_list_ip`, `rtx_access_list_ipv6`, `rtx_access_list_mac` resources
- Extended access list resources (`rtx_access_list_extended`, `rtx_access_list_extended_ipv6`)

## Files to Delete (Complete List)

### Resource Implementation Files
- `internal/provider/resource_rtx_interface_acl.go`
- `internal/provider/resource_rtx_interface_acl_test.go`
- `internal/provider/resource_rtx_interface_mac_acl.go`
- `internal/provider/resource_rtx_interface_mac_acl_test.go`
- `internal/provider/resource_rtx_ip_filter_dynamic.go`
- `internal/provider/resource_rtx_ip_filter_dynamic_test.go`
- `internal/provider/resource_rtx_ipv6_filter_dynamic.go`
- `internal/provider/resource_rtx_ipv6_filter_dynamic_test.go`

### Documentation Files
- `docs/resources/interface_acl.md`
- `docs/resources/interface_mac_acl.md`
- `docs/resources/ip_filter_dynamic.md`
- `docs/resources/ipv6_filter_dynamic.md`

### Provider Registration
- Remove from `internal/provider/provider.go` ResourcesMap:
  - `"rtx_interface_acl"`
  - `"rtx_interface_mac_acl"`
  - `"rtx_ip_filter_dynamic"`
  - `"rtx_ipv6_filter_dynamic"`

## Files to Create (Complete List)

### Resource Implementation Files
- `internal/provider/resource_rtx_access_list_ip_dynamic.go`
- `internal/provider/resource_rtx_access_list_ip_dynamic_test.go`
- `internal/provider/resource_rtx_access_list_ipv6_dynamic.go`
- `internal/provider/resource_rtx_access_list_ipv6_dynamic_test.go`

### Documentation Files
- `docs/resources/access_list_ip_dynamic.md`
- `docs/resources/access_list_ipv6_dynamic.md`

### Provider Registration
- Add to `internal/provider/provider.go` ResourcesMap:
  - `"rtx_access_list_ip_dynamic"`
  - `"rtx_access_list_ipv6_dynamic"`

## Files to Modify

- `internal/provider/resource_rtx_interface.go`
  - Remove: `secure_filter_in`, `secure_filter_out`, `dynamic_filter_out`, `ethernet_filter_in`, `ethernet_filter_out`
  - Add: `access_list_ip_in`, `access_list_ip_out`, `access_list_ipv6_in`, `access_list_ipv6_out`, `access_list_ip_dynamic_in`, `access_list_ip_dynamic_out`, `access_list_ipv6_dynamic_in`, `access_list_ipv6_dynamic_out`, `access_list_mac_in`, `access_list_mac_out`
- `internal/provider/resource_rtx_interface_test.go` - update tests
- `docs/resources/interface.md` - update attribute documentation
- `internal/client/interface_service.go` - remove old filter handling, add access list lookup
- `internal/client/interfaces.go` - update InterfaceConfig struct
