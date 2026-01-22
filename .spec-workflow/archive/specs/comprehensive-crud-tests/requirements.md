# Requirements Document: Comprehensive CRUD Tests

## Introduction

Terraform Provider for Yamaha RTX routers needs complete test coverage for all implemented resources. Currently, 15 resources lack unit tests at the Provider and/or Client layer. This specification defines requirements for adding comprehensive CRUD (Create, Read, Update, Delete) tests for all missing resources.

## Alignment with Product Vision

Complete test coverage ensures:
- Reliability of resource CRUD operations
- Regression prevention during future development
- Confidence in provider stability for production deployments
- Compliance with Terraform provider best practices

## Current Test Coverage Gap

| Resource | Client Test | Provider Test | RTX Commands |
|----------|-------------|---------------|--------------|
| rtx_access_list_extended | N/A | ❌ | ip access-list extended |
| rtx_access_list_extended_ipv6 | N/A | ❌ | ipv6 access-list extended |
| rtx_access_list_mac | N/A | ❌ | mac access-list extended |
| rtx_bgp | ❌ | ❌ | bgp neighbor, bgp local-as |
| rtx_dhcp_scope | ❌ | ❌ | dhcp scope |
| rtx_interface_acl | N/A | ❌ | ip access-group |
| rtx_interface_mac_acl | N/A | ❌ | mac access-group |
| rtx_ipsec_tunnel | ❌ | ❌ | ipsec tunnel, ipsec sa policy |
| rtx_l2tp | N/A | ❌ | l2tp tunnel |
| rtx_nat_masquerade | ❌ | ❌ | nat descriptor masquerade |
| rtx_nat_static | ❌ | ❌ | nat descriptor static |
| rtx_ospf | ❌ | ❌ | ospf area, ospf network |
| rtx_pptp | ❌ | ❌ | pptp service, pptp client |
| rtx_static_route | N/A | ❌ | ip route |
| rtx_system | ❌ | ❌ | timezone, console, packet-buffer |

## Requirements

### REQ-1: Client Layer Tests

**User Story:** As a developer, I want unit tests for all Client service methods, so that I can verify SSH command generation and response parsing work correctly.

#### Acceptance Criteria

1. WHEN a Client service method is called THEN the test SHALL verify correct RTX commands are generated
2. WHEN the RTX router returns a successful response THEN the test SHALL verify the response is parsed correctly
3. WHEN the RTX router returns an error THEN the test SHALL verify the error is handled appropriately
4. WHEN the Client service performs CRUD operations THEN tests SHALL cover Create, Read, Update, and Delete methods
5. IF a service has edge cases (empty config, partial config) THEN tests SHALL cover those scenarios

**Coverage Required:**
- bgp_service.go
- dhcp_scope_service.go
- ipsec_tunnel_service.go
- nat_masquerade_service.go
- nat_static_service.go
- ospf_service.go
- pptp_service.go
- system_service.go

### REQ-2: Provider Layer Tests

**User Story:** As a developer, I want unit tests for all Provider resource functions, so that I can verify Terraform schema handling works correctly.

#### Acceptance Criteria

1. WHEN resource data is provided via Terraform schema THEN the test SHALL verify correct conversion to internal types
2. WHEN a resource is created THEN the test SHALL verify all required attributes are set correctly
3. WHEN a resource is read THEN the test SHALL verify state is populated correctly
4. WHEN a resource is updated THEN the test SHALL verify changed attributes are handled
5. WHEN a resource is deleted THEN the test SHALL verify cleanup is performed
6. IF a resource has validation rules THEN tests SHALL verify validation works

**Coverage Required:**
- resource_rtx_access_list_extended.go
- resource_rtx_access_list_extended_ipv6.go
- resource_rtx_access_list_mac.go
- resource_rtx_bgp.go
- resource_rtx_dhcp_scope.go
- resource_rtx_interface_acl.go
- resource_rtx_interface_mac_acl.go
- resource_rtx_ipsec_tunnel.go
- resource_rtx_l2tp.go
- resource_rtx_nat_masquerade.go
- resource_rtx_nat_static.go
- resource_rtx_ospf.go
- resource_rtx_pptp.go
- resource_rtx_static_route.go
- resource_rtx_system.go

### REQ-3: Test Patterns Compliance

**User Story:** As a developer, I want all new tests to follow existing patterns, so that the codebase remains consistent and maintainable.

#### Acceptance Criteria

1. WHEN writing Client tests THEN tests SHALL use MockExecutor pattern with testify/mock
2. WHEN writing Provider tests THEN tests SHALL use schema.TestResourceDataRaw pattern
3. WHEN testing commands THEN tests SHALL include real RTX command format examples
4. WHEN testing parsing THEN tests SHALL include real RTX response format examples

### REQ-4: RTX Command Coverage

**User Story:** As a developer, I want tests to cover all RTX command variations, so that the provider handles real-world configurations.

#### Acceptance Criteria

1. WHEN an RTX command has optional parameters THEN tests SHALL cover both with and without optional parameters
2. WHEN an RTX command has multiple formats THEN tests SHALL cover all formats
3. WHEN an RTX feature has dependencies THEN tests SHALL verify dependency handling

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Each test file tests one service/resource
- **Modular Design**: Test helper functions should be reusable
- **Clear Interfaces**: Tests should mock only external dependencies

### Performance
- Unit tests SHALL complete within 10 seconds total
- No external network calls in unit tests

### Reliability
- Tests SHALL be deterministic (no flaky tests)
- Tests SHALL not depend on execution order

### Maintainability
- Tests SHALL use table-driven test patterns where appropriate
- Test names SHALL clearly describe the scenario being tested
