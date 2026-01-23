# Master Requirements: IPv6 Prefix Resource

## Overview

The IPv6 Prefix resource provides management of IPv6 prefix definitions on Yamaha RTX series routers. IPv6 prefixes are fundamental building blocks for IPv6 addressing, supporting three acquisition methods: static assignment, Router Advertisement (RA) prefix delegation, and DHCPv6 Prefix Delegation (DHCPv6-PD).

## Alignment with Product Vision

This resource directly supports the product goal of enabling Infrastructure as Code management of RTX router IPv6 configurations. IPv6 prefix management is essential for:
- Dual-stack network deployments
- ISP connectivity with delegated prefixes
- SLAAC (Stateless Address Autoconfiguration) setups
- Enterprise IPv6 network segmentation

The implementation follows consistent patterns established by other resources in this provider, with clear separation between static and dynamic prefix sources.

---

# Resource: rtx_ipv6_prefix

## Resource Summary

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_ipv6_prefix` |
| Type | Collection (indexed by prefix_id) |
| Import Support | Yes |
| Last Updated | 2026-01-23 |
| Source Specs | Implementation code analysis |

## Functional Requirements

### Core Operations

#### Create
- Creates a new IPv6 prefix definition on the RTX router
- Validates prefix configuration based on source type (static, ra, dhcpv6-pd)
- For static source: requires explicit prefix value
- For ra/dhcpv6-pd sources: requires source interface
- Saves configuration to persistent memory after successful creation
- Returns the prefix_id as the resource ID

#### Read
- Retrieves IPv6 prefix configuration by prefix_id
- Parses output from `show config | grep "ipv6 prefix"` command
- Returns all configured attributes including derived source type
- Marks resource as deleted if prefix not found

#### Update
- Updates existing IPv6 prefix settings
- Supports changing: prefix value (static only), prefix_length
- source and interface cannot be changed (ForceNew)
- Re-issues the prefix command with new values
- Saves configuration to persistent memory after successful update

#### Delete
- Removes the IPv6 prefix from the router using `no ipv6 prefix <id>`
- Gracefully handles already-deleted resources
- Saves configuration to persistent memory after successful deletion

### Requirement 1: Static Prefix Definition

**User Story:** As a network administrator, I want to define static IPv6 prefixes, so that I can configure predictable IPv6 addressing for my network segments.

#### Acceptance Criteria

1. WHEN source is "static" THEN the system SHALL require the prefix attribute
2. IF prefix is empty for static source THEN the system SHALL reject with validation error "'prefix' is required when source is 'static'"
3. WHEN a valid IPv6 prefix is provided THEN the system SHALL create the prefix definition
4. IF interface is specified for static source THEN the system SHALL reject with validation error "'interface' should not be set when source is 'static'"

### Requirement 2: Router Advertisement (RA) Prefix

**User Story:** As a network administrator, I want to use prefixes learned from upstream Router Advertisements, so that I can automatically adapt to ISP-assigned addressing.

#### Acceptance Criteria

1. WHEN source is "ra" THEN the system SHALL require the interface attribute
2. IF interface is empty for ra source THEN the system SHALL reject with validation error "'interface' is required when source is 'ra'"
3. IF prefix is specified for ra source THEN the system SHALL reject with validation error "'prefix' should not be set when source is 'ra'"
4. WHEN configured THEN the system SHALL generate command format `ipv6 prefix <id> ra-prefix@<interface>::/<length>`

### Requirement 3: DHCPv6 Prefix Delegation

**User Story:** As a network administrator, I want to use prefixes delegated via DHCPv6-PD from my ISP, so that I can implement proper IPv6 prefix hierarchy.

#### Acceptance Criteria

1. WHEN source is "dhcpv6-pd" THEN the system SHALL require the interface attribute
2. IF interface is empty for dhcpv6-pd source THEN the system SHALL reject with validation error "'interface' is required when source is 'dhcpv6-pd'"
3. IF prefix is specified for dhcpv6-pd source THEN the system SHALL reject with validation error "'prefix' should not be set when source is 'dhcpv6-pd'"
4. WHEN configured THEN the system SHALL generate command format `ipv6 prefix <id> dhcp-prefix@<interface>::/<length>`

### Requirement 4: Import Existing Prefixes

**User Story:** As a network administrator, I want to import existing IPv6 prefix configurations, so that I can manage pre-existing configurations with Terraform.

#### Acceptance Criteria

1. WHEN importing with prefix_id THEN the system SHALL retrieve and populate all prefix attributes
2. IF prefix_id does not exist THEN the system SHALL return an import error
3. WHEN import succeeds THEN all attributes (prefix, prefix_length, source, interface) SHALL be populated in state
4. WHEN importing THEN the system SHALL correctly determine the source type from the configuration format

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Parser handles output parsing, Service handles CRUD operations, Resource handles Terraform integration
- **Modular Design**: IPv6 prefix logic is isolated in dedicated service and parser files
- **Dependency Management**: Service depends on Executor interface for testability
- **Clear Interfaces**: Client interface defines CRUD contract for IPv6 prefixes

### Performance
- Single command execution for create/update/delete operations
- Parser efficiently handles multi-line configuration output
- Context cancellation support for long-running operations

### Security
- No credentials stored in Terraform state
- Configuration saved to persistent memory to survive router reboots

### Reliability
- Context cancellation support for all operations
- Graceful handling of "not found" errors during read/delete
- Configuration saved after each successful operation
- Update operation validates source/interface changes are not attempted

### Validation
- Prefix ID range validation (1-255)
- Prefix length range validation (1-128)
- IPv6 prefix format validation for static sources
- Source type validation (static, ra, dhcpv6-pd)
- Interface name validation for ra/dhcpv6-pd sources
- Cross-field validation via CustomizeDiff

## RTX Commands Reference

```
# Create static prefix
ipv6 prefix <id> <prefix>::/<length>
Example: ipv6 prefix 1 2001:db8::/<64>

# Create RA-derived prefix
ipv6 prefix <id> ra-prefix@<interface>::/<length>
Example: ipv6 prefix 2 ra-prefix@lan2::/64

# Create DHCPv6-PD prefix
ipv6 prefix <id> dhcp-prefix@<interface>::/<length>
Example: ipv6 prefix 3 dhcp-prefix@lan2::/48

# Delete prefix
no ipv6 prefix <id>
Example: no ipv6 prefix 1

# Show prefix configuration
show config | grep "ipv6 prefix <id>"
show config | grep "ipv6 prefix"
```

## Terraform Command Support

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | Yes | Compares desired state with RTX configuration |
| `terraform apply` | Yes | Creates/updates IPv6 prefix on router |
| `terraform destroy` | Yes | Removes IPv6 prefix from router |
| `terraform import` | Yes | Imports existing prefix by prefix_id |
| `terraform refresh` | Yes | Reads current prefix state from router |
| `terraform state` | Yes | Manages prefix in local state file |

### Import Specification
- **Import ID Format**: `<prefix_id>` (integer)
- **Import Command**: `terraform import rtx_ipv6_prefix.example 1`
- **Post-Import**: All attributes populated from router configuration

## Example Usage

```hcl
# Static IPv6 prefix
resource "rtx_ipv6_prefix" "static" {
  prefix_id     = 1
  prefix        = "2001:db8::"
  prefix_length = 64
  source        = "static"
}

# Router Advertisement derived prefix
resource "rtx_ipv6_prefix" "ra" {
  prefix_id     = 2
  prefix_length = 64
  source        = "ra"
  interface     = "lan2"
}

# DHCPv6 Prefix Delegation
resource "rtx_ipv6_prefix" "dhcpv6_pd" {
  prefix_id     = 3
  prefix_length = 48
  source        = "dhcpv6-pd"
  interface     = "lan2"
}

# Using prefix with IPv6 interface address
resource "rtx_ipv6_prefix" "wan_pd" {
  prefix_id     = 10
  prefix_length = 56
  source        = "dhcpv6-pd"
  interface     = "pp1"
}

# Static documentation prefix
resource "rtx_ipv6_prefix" "documentation" {
  prefix_id     = 100
  prefix        = "2001:db8:1234::"
  prefix_length = 48
  source        = "static"
}
```

## Terraform Schema

### Attributes

| Attribute | Type | Required | Optional | ForceNew | Computed | Description |
|-----------|------|----------|----------|----------|----------|-------------|
| `prefix_id` | Int | Yes | - | Yes | - | IPv6 prefix ID (1-255) |
| `prefix` | String | - | Yes | - | - | Static IPv6 prefix value (e.g., '2001:db8::') |
| `prefix_length` | Int | Yes | - | - | - | Prefix length in bits (1-128) |
| `source` | String | Yes | - | Yes | - | Prefix source type: 'static', 'ra', or 'dhcpv6-pd' |
| `interface` | String | - | Yes | Yes | - | Source interface for 'ra' and 'dhcpv6-pd' sources |

### Validation Rules

1. **prefix_id**: Must be between 1 and 255
2. **prefix_length**: Must be between 1 and 128
3. **source**: Must be one of "static", "ra", or "dhcpv6-pd"
4. **prefix**: Required when source is "static"; must not be set when source is "ra" or "dhcpv6-pd"
5. **interface**: Required when source is "ra" or "dhcpv6-pd"; must not be set when source is "static"

## State Handling

- Only configuration attributes are persisted in Terraform state
- Operational status (actual delegated prefix values for ra/dhcpv6-pd) is not stored
- State includes: prefix_id, prefix (for static), prefix_length, source, interface (for ra/dhcpv6-pd)
- Dynamic prefix values (from RA or DHCPv6-PD) are not tracked in state

---

## Change History

| Date | Source Spec | Changes |
|------|-------------|---------|
| 2026-01-23 | Implementation analysis | Initial master spec creation from implementation code |
| 2026-01-23 | terraform-plan-differences-fix | IPv6 filter dynamic service now fully implemented (stub methods replaced with proper IPFilterService delegation) |
