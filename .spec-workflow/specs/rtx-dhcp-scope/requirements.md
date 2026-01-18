# Requirements Document: rtx_dhcp_scope

## Introduction

This feature enables Terraform-based management of DHCP scopes on Yamaha RTX series routers. A DHCP scope defines the IP address range and associated network parameters that the router's DHCP server uses to allocate IP addresses to client devices. This is the parent resource for the already-implemented `rtx_dhcp_binding` resource, enabling complete Infrastructure as Code (IaC) management of DHCP configuration.

## Alignment with Product Vision

The `rtx_dhcp_scope` resource aligns with the provider's goal of enabling comprehensive network infrastructure management for Yamaha RTX routers:

- **IaC for Network Infrastructure**: Complete DHCP configuration as code
- **Dependency Management**: Parent resource for `rtx_dhcp_binding` enables proper Terraform dependency graphs
- **Enterprise Readiness**: Scope management is essential for enterprise DHCP deployments
- **Consistency**: Follows existing patterns established by `rtx_dhcp_binding`

## Requirements

### Requirement 1: DHCP Scope CRUD Operations

**User Story:** As a network administrator, I want to create, read, update, and delete DHCP scopes via Terraform, so that I can manage IP address allocation as code.

#### Acceptance Criteria

1. WHEN a user applies a Terraform configuration with an `rtx_dhcp_scope` resource THEN the system SHALL execute the appropriate RTX `dhcp scope` command to create the scope
2. WHEN a user runs `terraform plan` on an existing scope THEN the system SHALL read the current scope configuration and compare it with the desired state
3. WHEN a user modifies scope parameters and applies THEN the system SHALL update the scope configuration on the RTX router
4. WHEN a user removes an `rtx_dhcp_scope` resource and applies THEN the system SHALL delete the scope from the RTX router
5. IF a scope has active leases or bindings THEN the system SHALL warn the user before deletion

### Requirement 2: IP Address Range Configuration

**User Story:** As a network administrator, I want to define the IP address range for a DHCP scope, so that client devices receive addresses within a specified range.

#### Acceptance Criteria

1. WHEN creating a scope THEN the system SHALL require a network address with CIDR notation (e.g., `192.168.1.0/24`)
2. IF the network address is invalid THEN the system SHALL return a validation error before attempting configuration
3. WHEN the scope is created THEN the system SHALL configure the start and end addresses based on the network definition
4. IF exclude ranges are specified THEN the system SHALL configure the RTX router to exclude those addresses from allocation

### Requirement 3: Gateway and DNS Configuration

**User Story:** As a network administrator, I want to configure gateway and DNS servers for a DHCP scope, so that clients receive complete network configuration.

#### Acceptance Criteria

1. WHEN a gateway address is specified THEN the system SHALL configure DHCP option 3 (default gateway)
2. WHEN DNS server addresses are specified THEN the system SHALL configure DHCP option 6 (DNS servers)
3. IF multiple DNS servers are specified (up to 3) THEN the system SHALL configure all servers in order
4. IF gateway or DNS addresses are outside the scope network THEN the system SHALL allow it (common for routed networks)

### Requirement 4: Lease Time Configuration

**User Story:** As a network administrator, I want to configure lease duration, so that I can control how long clients retain their IP addresses.

#### Acceptance Criteria

1. WHEN a lease time is specified THEN the system SHALL configure the lease duration on the scope
2. IF lease time is not specified THEN the system SHALL use the RTX router's default (typically 72 hours)
3. WHEN lease time is updated THEN the system SHALL apply the new time to new leases (existing leases are not affected)
4. IF lease time is set to `infinite` THEN the system SHALL configure permanent leases

### Requirement 5: Scope Import Functionality

**User Story:** As a network administrator with existing RTX configurations, I want to import existing DHCP scopes into Terraform state, so that I can manage them with IaC without recreating them.

#### Acceptance Criteria

1. WHEN a user runs `terraform import rtx_dhcp_scope.example <scope_id>` THEN the system SHALL read the existing scope configuration
2. IF the scope exists THEN the system SHALL populate all Terraform state attributes from the router configuration
3. IF the scope does not exist THEN the system SHALL return an appropriate error message
4. WHEN imported THEN all attributes including gateway, DNS, and lease time SHALL be correctly populated

### Requirement 6: Integration with rtx_dhcp_binding

**User Story:** As a network administrator, I want DHCP scopes and bindings to have proper dependency relationships, so that Terraform creates and destroys resources in the correct order.

#### Acceptance Criteria

1. WHEN an `rtx_dhcp_binding` references a scope_id THEN Terraform dependency graph SHALL ensure the scope exists first
2. WHEN destroying resources THEN bindings SHALL be removed before the scope
3. IF a user tries to reference a non-existent scope_id THEN the system SHALL return a clear error during apply

## Non-Functional Requirements

### Code Architecture and Modularity

- **Single Responsibility Principle**: DHCPScopeService separate from DHCPService (bindings)
- **Modular Design**: Parser, service, and resource layers clearly separated
- **Dependency Management**: Reuse existing client infrastructure and patterns
- **Clear Interfaces**: Extend Client interface with scope management methods

### Performance

- Scope operations should complete within the standard provider timeout (30 seconds default)
- Minimize SSH round-trips by batching related commands where possible
- Cache scope information during a single Terraform operation

### Security

- Scope configurations do not contain sensitive data by default
- Gateway and DNS addresses should be validated for format but not connectivity
- Document security implications of DHCP in network documentation

### Reliability

- Handle RTX router command execution errors gracefully
- Provide clear error messages for common configuration mistakes
- Support retry logic for transient connection failures
- Implement proper locking to prevent concurrent scope modifications

### Usability

- Schema should mirror RTX CLI terminology where practical
- Provide sensible defaults for optional parameters
- Include comprehensive examples in documentation
- Validation errors should reference RTX documentation when helpful

## RTX Router Commands Reference

Based on Yamaha RTX documentation:

### Create/Configure Scope
```
dhcp scope <scope-id> <network>/<prefix> [gateway <gateway>] [expire <time>]
```
Example: `dhcp scope 1 192.168.1.0/24 gateway 192.168.1.1 expire 3:00`

### Configure DNS Servers
```
dhcp scope option <scope-id> dns=<dns1>[,<dns2>[,<dns3>]]
```
Example: `dhcp scope option 1 dns=8.8.8.8,8.8.4.4`

### Configure Exclude Range
```
dhcp scope <scope-id> except <start-ip>-<end-ip>
```
Example: `dhcp scope 1 except 192.168.1.1-192.168.1.10`

### Show Scope Configuration
```
show dhcp scope [<scope-id>]
```

### Delete Scope
```
no dhcp scope <scope-id>
```

## Proposed Terraform Schema

```hcl
resource "rtx_dhcp_scope" "example" {
  scope_id = 1  # Required, ForceNew

  # Network configuration
  network = "192.168.1.0/24"  # Required, ForceNew

  # Optional parameters
  gateway     = "192.168.1.1"
  dns_servers = ["8.8.8.8", "8.8.4.4"]

  # Lease configuration
  lease_time = "72h"  # Go duration format, or "infinite"

  # Exclusions
  exclude_ranges = [
    {
      start = "192.168.1.1"
      end   = "192.168.1.10"
    }
  ]
}
```

## Dependencies

- Existing `internal/client/` infrastructure
- Existing `internal/rtx/parsers/` registry pattern
- `rtx_dhcp_binding` resource (for integration testing)
