# Master Requirements: DHCP Resources

## Overview

DHCP resources provide comprehensive management of DHCP (Dynamic Host Configuration Protocol) services on Yamaha RTX series routers. This includes DHCP scope management for dynamic IP address allocation and DHCP binding management for static IP reservations.

## Alignment with Product Vision

These resources directly support the product goal of enabling Infrastructure as Code management of RTX router configurations. DHCP is a fundamental network service that benefits from:
- Version-controlled configuration
- Reproducible deployments across multiple sites
- Automated management through CI/CD pipelines

The implementation follows the Cisco-compatible naming conventions outlined in product.md, using terms like `scope`, `routers`, and `dns_servers` that align with industry standards.

---

# Resource: rtx_dhcp_scope

## Resource Summary

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_dhcp_scope` |
| Type | Collection (indexed by scope_id) |
| Import Support | Yes |
| Last Updated | 2026-01-23 |
| Source Specs | Implementation code analysis |

## Functional Requirements

### Core Operations

#### Create
- Creates a new DHCP scope on the RTX router
- Validates network format is valid CIDR notation
- Configures optional DNS servers, routers (default gateways), and domain name
- Configures optional IP exclusion ranges
- Saves configuration to persistent memory after successful creation
- Returns the scope_id as the resource ID

#### Read
- Retrieves DHCP scope configuration by scope_id
- Parses output from `show config | grep "dhcp scope"` command
- Returns all configured attributes including options and exclusions
- Marks resource as deleted if scope not found

#### Update
- Updates existing DHCP scope settings
- Supports changing: lease_time, DNS servers, routers, domain_name, exclusion ranges
- Network and scope_id cannot be changed (ForceNew)
- Performs differential update for exclusion ranges (adds new, removes old)
- Saves configuration to persistent memory after successful update

#### Delete
- Removes the DHCP scope from the router
- Gracefully handles already-deleted resources
- Saves configuration to persistent memory after successful deletion

### Requirement 1: Scope Definition

**User Story:** As a network administrator, I want to define DHCP scopes with network ranges and lease times, so that I can automatically assign IP addresses to network clients.

#### Acceptance Criteria

1. WHEN a valid CIDR network is provided THEN the system SHALL create a DHCP scope with that network range
2. IF scope_id is less than 1 THEN the system SHALL reject the configuration with a validation error
3. WHEN lease_time is specified THEN the system SHALL convert Go duration format (e.g., "72h") to RTX format (e.g., "72:00")
4. IF lease_time is "infinite" THEN the system SHALL configure permanent leases

### Requirement 2: DHCP Options (RFC 2132)

**User Story:** As a network administrator, I want to configure DHCP options like DNS servers and default gateways, so that clients receive complete network configuration.

#### Acceptance Criteria

1. WHEN dns_servers are specified THEN the system SHALL configure up to 3 DNS server addresses
2. IF more than 3 DNS servers are provided THEN the system SHALL reject with validation error
3. WHEN routers (default gateways) are specified THEN the system SHALL configure up to 3 gateway addresses
4. IF routers or DNS servers contain invalid IP addresses THEN the system SHALL reject with validation error
5. WHEN domain_name is specified THEN the system SHALL configure the DNS domain for clients

### Requirement 3: IP Exclusion Ranges

**User Story:** As a network administrator, I want to exclude specific IP ranges from DHCP allocation, so that I can reserve addresses for static assignments.

#### Acceptance Criteria

1. WHEN exclude_ranges are specified THEN the system SHALL prevent those IPs from being allocated
2. IF exclude_range has invalid start or end IP THEN the system SHALL reject with validation error
3. WHEN updating exclusion ranges THEN the system SHALL add new ranges and remove old ones differentially

### Requirement 4: Import Existing Scopes

**User Story:** As a network administrator, I want to import existing DHCP scope configurations, so that I can manage pre-existing configurations with Terraform.

#### Acceptance Criteria

1. WHEN importing with scope_id THEN the system SHALL retrieve and populate all scope attributes
2. IF scope_id does not exist THEN the system SHALL return an import error
3. WHEN import succeeds THEN all attributes (network, options, exclusions) SHALL be populated in state

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Parser handles output parsing, Service handles CRUD operations, Resource handles Terraform integration
- **Modular Design**: DHCP scope logic is isolated in dedicated service and parser files
- **Dependency Management**: Service depends on Executor interface for testability
- **Clear Interfaces**: Client interface defines CRUD contract

### Performance
- Batch command execution for create/update operations to minimize SSH round-trips
- Parser efficiently handles multi-line configuration output

### Security
- No credentials stored in Terraform state
- Configuration saved to persistent memory to survive router reboots

### Reliability
- Context cancellation support for long-running operations
- Graceful handling of "not found" errors during delete
- Configuration saved after each successful operation

### Validation
- CIDR notation validation for network addresses
- IPv4 address validation for all IP fields
- Maximum count validation for DNS servers (3) and routers (3)
- Positive integer validation for scope_id

## RTX Commands Reference

```
# Create/Update scope
dhcp scope <id> <network>/<prefix> [expire <time>]

# Configure options
dhcp scope option <id> dns=<dns1>[,<dns2>[,<dns3>]] [router=<gw1>[,<gw2>]] [domain=<domain>]

# Add exclusion range
dhcp scope <id> except <start_ip>-<end_ip>

# Delete exclusion range
no dhcp scope <id> except <start_ip>-<end_ip>

# Delete options
no dhcp scope option <id>

# Delete scope
no dhcp scope <id>

# Show scope configuration
show config | grep "dhcp scope"
```

## Terraform Command Support

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | Yes | Compares desired state with RTX configuration |
| `terraform apply` | Yes | Creates/updates DHCP scope on router |
| `terraform destroy` | Yes | Removes DHCP scope from router |
| `terraform import` | Yes | Imports existing scope by scope_id |
| `terraform refresh` | Yes | Reads current scope state from router |
| `terraform state` | Yes | Manages scope in local state file |

### Import Specification
- **Import ID Format**: `<scope_id>` (integer)
- **Import Command**: `terraform import rtx_dhcp_scope.example 1`
- **Post-Import**: All attributes populated from router configuration

## Example Usage

```hcl
# Basic DHCP scope with required attributes
resource "rtx_dhcp_scope" "office" {
  scope_id = 1
  network  = "192.168.1.0/24"
}

# Full DHCP scope with all options
resource "rtx_dhcp_scope" "guest" {
  scope_id   = 2
  network    = "192.168.100.0/24"
  lease_time = "24h"

  options {
    routers     = ["192.168.100.1"]
    dns_servers = ["8.8.8.8", "8.8.4.4"]
    domain_name = "guest.local"
  }

  exclude_ranges {
    start = "192.168.100.1"
    end   = "192.168.100.10"
  }

  exclude_ranges {
    start = "192.168.100.250"
    end   = "192.168.100.254"
  }
}

# Scope with infinite lease
resource "rtx_dhcp_scope" "servers" {
  scope_id   = 3
  network    = "10.0.0.0/24"
  lease_time = "infinite"

  options {
    routers     = ["10.0.0.1"]
    dns_servers = ["10.0.0.2"]
  }
}
```

## State Handling

- Only configuration attributes are persisted in Terraform state
- Operational status (current leases, lease counts) is not stored
- State includes: scope_id, network, lease_time, exclude_ranges, options

---

# Resource: rtx_dhcp_binding

## Resource Summary

| Attribute | Value |
|-----------|-------|
| Resource Name | `rtx_dhcp_binding` |
| Type | Collection (indexed by scope_id + identifier) |
| Import Support | Yes |
| Last Updated | 2026-01-23 |
| Source Specs | Implementation code analysis |

## Functional Requirements

### Core Operations

#### Create
- Creates a static DHCP binding (IP reservation) on the RTX router
- Associates an IP address with either a MAC address or DHCP client identifier
- Validates client identification method (exactly one required)
- Saves configuration to persistent memory after successful creation
- Returns composite ID: `<scope_id>:<mac_address>`

#### Read
- Retrieves DHCP binding by parsing scope bindings list
- Supports both MAC address and legacy IP address ID formats for backward compatibility
- Marks resource as deleted if binding not found

#### Update
- DHCP bindings do not support in-place updates
- All attributes are ForceNew (changes require recreation)

#### Delete
- Removes the DHCP binding from the router
- Looks up IP address from MAC address before deletion
- Gracefully handles already-deleted resources
- Saves configuration to persistent memory after successful deletion

### Requirement 1: MAC Address Binding

**User Story:** As a network administrator, I want to assign static IP addresses based on device MAC addresses, so that specific devices always receive the same IP address.

#### Acceptance Criteria

1. WHEN mac_address is provided THEN the system SHALL create a binding for that MAC
2. IF mac_address format is invalid THEN the system SHALL reject with validation error
3. WHEN mac_address is in any valid format (colons, hyphens, no separator) THEN the system SHALL normalize to colon-separated lowercase
4. IF use_mac_as_client_id is true THEN the system SHALL use "ethernet" prefix in RTX command

### Requirement 2: Client Identifier Binding

**User Story:** As a network administrator, I want to assign static IP addresses based on DHCP client identifiers, so that I can support devices that don't use MAC-based identification.

#### Acceptance Criteria

1. WHEN client_identifier is provided THEN the system SHALL create a binding using client-id mode
2. IF client_identifier format is invalid THEN the system SHALL reject with validation error
3. WHEN client_identifier prefix is 01 THEN the system SHALL treat it as MAC-based
4. WHEN client_identifier prefix is 02 THEN the system SHALL treat it as ASCII-based
5. WHEN client_identifier prefix is FF THEN the system SHALL treat it as vendor-specific
6. IF unsupported prefix is used THEN the system SHALL reject with validation error

### Requirement 3: Mutual Exclusivity

**User Story:** As a network administrator, I want clear validation of client identification methods, so that I don't accidentally configure conflicting bindings.

#### Acceptance Criteria

1. IF both mac_address and client_identifier are specified THEN the system SHALL reject with validation error
2. IF neither mac_address nor client_identifier is specified THEN the system SHALL reject with validation error
3. IF use_mac_as_client_id is set without mac_address THEN the system SHALL reject with validation error
4. IF use_mac_as_client_id is set with client_identifier THEN the system SHALL reject with validation error

### Requirement 4: Import Existing Bindings

**User Story:** As a network administrator, I want to import existing DHCP bindings, so that I can manage pre-existing reservations with Terraform.

#### Acceptance Criteria

1. WHEN importing with "scope_id:mac_address" THEN the system SHALL find and import the binding
2. WHEN importing with "scope_id:ip_address" (legacy) THEN the system SHALL find and import the binding
3. IF binding not found THEN the system SHALL return an import error
4. WHEN import succeeds THEN all attributes SHALL be populated including discovered use_mac_as_client_id value

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Parser handles output parsing, Service handles CRUD operations
- **Modular Design**: DHCP binding logic is isolated in dedicated service and parser files
- **Clear Interfaces**: DHCPBindingsParser interface enables mock testing

### Performance
- Single command execution for create/delete operations
- Parser handles multiple output formats (RTX830, RTX1210, config format)

### Security
- No sensitive data in bindings (MAC addresses are not credentials)
- Configuration saved to persistent memory

### Reliability
- Context cancellation support for operations
- Graceful handling of "not found" errors
- MAC address normalization ensures consistent state matching

### Validation
- IPv4 address validation for ip_address
- MAC address format validation (12 hex digits)
- Client identifier format validation (type:data format, valid hex octets)
- Client identifier length validation (max 255 bytes)
- Positive integer validation for scope_id

## RTX Commands Reference

```
# Create binding with MAC address
dhcp scope bind <scope_id> <ip_address> <mac_address>

# Create binding with ethernet (client-id) mode
dhcp scope bind <scope_id> <ip_address> ethernet <mac_address>

# Create binding with custom client identifier
dhcp scope bind <scope_id> <ip_address> client-id <type:hex:hex:...>

# Delete binding
no dhcp scope bind <scope_id> <ip_address>

# Show bindings for a scope
show config | grep "dhcp scope bind <scope_id>"
```

## Terraform Command Support

| Command | Support | Description |
|---------|---------|-------------|
| `terraform plan` | Yes | Compares desired state with RTX configuration |
| `terraform apply` | Yes | Creates DHCP binding on router |
| `terraform destroy` | Yes | Removes DHCP binding from router |
| `terraform import` | Yes | Imports existing binding |
| `terraform refresh` | Yes | Reads current binding state from router |
| `terraform state` | Yes | Manages binding in local state file |

### Import Specification
- **Import ID Format**: `<scope_id>:<mac_address>` or `<scope_id>:<ip_address>` (legacy)
- **Import Command**: `terraform import rtx_dhcp_binding.server1 1:00:11:22:33:44:55`
- **Post-Import**: All attributes populated, ID normalized to MAC address format

## Example Usage

```hcl
# Basic MAC address binding
resource "rtx_dhcp_binding" "server1" {
  scope_id   = 1
  ip_address = "192.168.1.100"
  mac_address = "00:11:22:33:44:55"
}

# MAC address binding with client identifier mode (ethernet prefix)
resource "rtx_dhcp_binding" "server2" {
  scope_id            = 1
  ip_address          = "192.168.1.101"
  mac_address         = "00:aa:bb:cc:dd:ee"
  use_mac_as_client_id = true
}

# Custom client identifier binding (MAC-based, 01 prefix)
resource "rtx_dhcp_binding" "printer" {
  scope_id          = 1
  ip_address        = "192.168.1.200"
  client_identifier = "01:00:11:22:33:44:55"
}

# Custom client identifier binding (ASCII-based, 02 prefix)
resource "rtx_dhcp_binding" "custom_device" {
  scope_id          = 1
  ip_address        = "192.168.1.201"
  client_identifier = "02:68:6f:73:74:6e:61:6d:65"  # "hostname" in hex
}

# With optional metadata
resource "rtx_dhcp_binding" "workstation" {
  scope_id    = 1
  ip_address  = "192.168.1.50"
  mac_address = "aa:bb:cc:dd:ee:ff"
  hostname    = "workstation1"
  description = "Reception desk workstation"
}
```

## State Handling

- Configuration attributes persisted in Terraform state
- Resource ID format: `<scope_id>:<normalized_mac_address>`
- MAC addresses normalized to lowercase colon-separated format
- Operational status (lease status) is not stored

---

## Change History

| Date | Source Spec | Changes |
|------|-------------|---------|
| 2026-01-23 | Implementation analysis | Initial master spec creation from implementation code |
| 2026-01-23 | terraform-plan-differences-fix | Network address is calculated from IP range start and prefix length; range_start/range_end fields added to support IP range format parsing |
