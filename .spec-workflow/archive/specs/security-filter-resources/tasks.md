# Tasks Document: Security and Filter Resources

## Summary

Implementation of 4 new Terraform resources for RTX router security and filter management:
- `rtx_access_list_ip` - Static IP filters
- `rtx_access_list_ipv6` - Static IPv6 filters
- `rtx_l2tp_service` - L2TP service on/off
- `rtx_ipsec_transport` - IPsec transport mode mappings

---

## Phase 1: rtx_access_list_ip Resource

- [x] 1.1. Extend IP filter parser for static filters
  - File: `internal/rtx/parsers/ip_filter.go`
  - Add parsing support for static `ip filter N action src dst proto sport dport` commands
  - Verify `BuildIPFilterCommand` and `BuildDeleteIPFilterCommand` work correctly
  - Add unit tests for new parsing scenarios
  - Purpose: Enable reading/writing static IP filter configurations
  - _Leverage: Existing `IPFilter` struct and `ParseIPFilterConfig` function_
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in parser implementations | Task: Extend internal/rtx/parsers/ip_filter.go to support static IP filter parsing per REQ-1, verify existing BuildIPFilterCommand works for create/delete | Restrictions: Do not modify existing IPFilter struct fields unless necessary, maintain backward compatibility with dynamic filter parsing | Success: Static IP filters parse correctly, commands generate valid RTX CLI syntax, unit tests pass | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

- [x] 1.2. Extend IP filter service for static filter CRUD
  - File: `internal/client/ip_filter_service.go`
  - Add `CreateIPFilter`, `GetIPFilter`, `UpdateIPFilter`, `DeleteIPFilter` methods
  - Implement import support via `GetIPFilterByNumber`
  - Purpose: Provide client-side CRUD operations for static IP filters
  - _Leverage: Existing service patterns in `internal/client/`_
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with SSH client experience | Task: Extend internal/client/ip_filter_service.go with CRUD methods for static IP filters per REQ-1, following existing service patterns | Restrictions: Use existing SSH session management, follow error handling patterns | Success: CRUD operations work via SSH, proper error handling, methods follow existing patterns | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

- [x] 1.3. Create rtx_access_list_ip resource
  - File: `internal/provider/resource_rtx_access_list_ip.go`
  - Implement full resource with Create, Read, Update, Delete, Import
  - Define schema matching design.md specification
  - Purpose: Terraform resource for managing static IP filters
  - _Leverage: `resource_rtx_ip_filter_dynamic.go` as pattern_
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create internal/provider/resource_rtx_access_list_ip.go implementing full CRUD and import per REQ-1, following patterns from resource_rtx_ip_filter_dynamic.go | Restrictions: Follow terraform-plugin-sdk/v2 patterns, maintain consistency with existing resources | Success: Resource compiles, schema matches design.md, CRUD operations work correctly | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

- [x] 1.4. Add unit tests for rtx_access_list_ip
  - File: `internal/provider/resource_rtx_access_list_ip_test.go`
  - Test schema validation, CRUD operations with mocked client
  - Purpose: Ensure resource reliability
  - _Leverage: Existing test patterns in `internal/provider/*_test.go`_
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for rtx_access_list_ip resource covering schema validation and CRUD with mocked dependencies | Restrictions: Use testify for assertions, mock external dependencies | Success: Tests pass, good coverage of success and error scenarios | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

---

## Phase 2: rtx_access_list_ipv6 Resource

- [x] 2.1. Verify IPv6 filter parser support
  - File: `internal/rtx/parsers/ip_filter.go`
  - Confirm `ParseIPv6FilterConfig`, `BuildIPv6FilterCommand`, `BuildDeleteIPv6FilterCommand` work
  - Add any missing parsing scenarios for static IPv6 filters
  - Purpose: Ensure IPv6 filter parsing is complete
  - _Leverage: Existing IPv6 functions in ip_filter.go_
  - _Requirements: REQ-3_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Verify and extend IPv6 filter parsing in internal/rtx/parsers/ip_filter.go per REQ-3, add tests for static IPv6 filter scenarios | Restrictions: Maintain consistency with IPv4 filter patterns | Success: IPv6 static filters parse correctly, commands generate valid syntax | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

- [x] 2.2. Extend IP filter service for IPv6 static filter CRUD
  - File: `internal/client/ip_filter_service.go`
  - Add `CreateIPv6Filter`, `GetIPv6Filter`, `UpdateIPv6Filter`, `DeleteIPv6Filter` methods
  - Purpose: Client-side CRUD for static IPv6 filters
  - _Leverage: IPv4 static filter methods from task 1.2_
  - _Requirements: REQ-3_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add IPv6 static filter CRUD methods to ip_filter_service.go per REQ-3, following IPv4 patterns from task 1.2 | Restrictions: Maintain consistent error handling with IPv4 methods | Success: IPv6 CRUD operations work correctly | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

- [x] 2.3. Create rtx_access_list_ipv6 resource
  - File: `internal/provider/resource_rtx_access_list_ipv6.go`
  - Implement full resource with Create, Read, Update, Delete, Import
  - Schema for IPv6-specific attributes (icmp6 protocol, etc.)
  - Purpose: Terraform resource for managing static IPv6 filters
  - _Leverage: `resource_rtx_access_list_ip.go` from task 1.3_
  - _Requirements: REQ-3_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create internal/provider/resource_rtx_access_list_ipv6.go per REQ-3, based on rtx_access_list_ip patterns | Restrictions: Handle IPv6-specific protocols like icmp6 | Success: Resource compiles, handles IPv6 specifics correctly | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

- [x] 2.4. Add unit tests for rtx_access_list_ipv6
  - File: `internal/provider/resource_rtx_access_list_ipv6_test.go`
  - Test IPv6-specific scenarios (icmp6, IPv6 addresses)
  - Purpose: Ensure IPv6 resource reliability
  - _Leverage: rtx_access_list_ip tests from task 1.4_
  - _Requirements: REQ-3_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for rtx_access_list_ipv6 resource with IPv6-specific test cases | Restrictions: Cover IPv6-specific edge cases | Success: Tests pass, IPv6 specifics are well tested | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

---

## Phase 3: rtx_l2tp_service Resource

- [x] 3.1. Extend L2TP parser for service on/off
  - File: `internal/rtx/parsers/l2tp.go`
  - Add `L2TPService` struct with `Enabled` and `Protocols` fields
  - Add `ParseL2TPServiceConfig` function to parse `l2tp service on/off` output
  - Verify `BuildL2TPServiceCommand` generates correct syntax
  - Purpose: Enable parsing L2TP service state
  - _Leverage: Existing `L2TPParser` and `BuildL2TPServiceCommand`_
  - _Requirements: REQ-6_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Extend internal/rtx/parsers/l2tp.go with L2TPService struct and ParseL2TPServiceConfig per REQ-6 | Restrictions: Do not break existing L2TP tunnel parsing | Success: L2TP service state parses correctly, commands generate valid syntax | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

- [x] 3.2. Extend L2TP service for service CRUD
  - File: `internal/client/l2tp_service.go`
  - Add `EnableL2TPService`, `DisableL2TPService`, `GetL2TPServiceState` methods
  - Purpose: Client-side operations for L2TP service management
  - _Leverage: Existing L2TP service patterns_
  - _Requirements: REQ-6_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add L2TP service state methods to l2tp_service.go per REQ-6 | Restrictions: Handle singleton nature of service resource | Success: Service enable/disable works correctly | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

- [x] 3.3. Create rtx_l2tp_service resource
  - File: `internal/provider/resource_rtx_l2tp_service.go`
  - Implement singleton resource (only one instance per router)
  - Schema: `enabled` (bool), `protocols` (list of strings)
  - Import ID: "default"
  - Purpose: Terraform resource for L2TP service configuration
  - _Leverage: Singleton resource patterns_
  - _Requirements: REQ-6_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create internal/provider/resource_rtx_l2tp_service.go as singleton resource per REQ-6 | Restrictions: Handle singleton semantics (only one instance), use "default" as import ID | Success: Resource works as singleton, enables/disables L2TP service correctly | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

- [x] 3.4. Add unit tests for rtx_l2tp_service
  - File: `internal/provider/resource_rtx_l2tp_service_test.go`
  - Test singleton behavior, enable/disable scenarios
  - Purpose: Ensure L2TP service resource reliability
  - _Leverage: Existing resource test patterns_
  - _Requirements: REQ-6_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for rtx_l2tp_service singleton resource | Restrictions: Test singleton semantics properly | Success: Tests pass, singleton behavior verified | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

---

## Phase 4: rtx_ipsec_transport Resource

- [x] 4.1. Create IPsec transport parser
  - File: `internal/rtx/parsers/ipsec_transport.go`
  - Create `IPsecTransport` struct with `TransportID`, `TunnelID`, `Protocol`, `Port`
  - Implement `ParseIPsecTransportConfig` to parse `ipsec transport N tunnel proto port`
  - Implement `BuildIPsecTransportCommand` and `BuildDeleteIPsecTransportCommand`
  - Add unit tests
  - Purpose: Parser for IPsec transport mode configurations
  - _Leverage: `parsers/ipsec_tunnel.go` as reference_
  - _Requirements: REQ-7_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Create internal/rtx/parsers/ipsec_transport.go with parser and command builders per REQ-7, following patterns from ipsec_tunnel.go | Restrictions: Follow existing parser patterns, register in parser registry | Success: Parser works correctly, commands generate valid syntax, unit tests pass | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

- [x] 4.2. Create IPsec transport service
  - File: `internal/client/ipsec_transport_service.go`
  - Implement `CreateIPsecTransport`, `GetIPsecTransport`, `UpdateIPsecTransport`, `DeleteIPsecTransport`
  - Purpose: Client-side CRUD for IPsec transport mappings
  - _Leverage: Existing service patterns_
  - _Requirements: REQ-7_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Create internal/client/ipsec_transport_service.go with CRUD methods per REQ-7 | Restrictions: Follow existing service patterns, proper error handling | Success: CRUD operations work correctly | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

- [x] 4.3. Create rtx_ipsec_transport resource
  - File: `internal/provider/resource_rtx_ipsec_transport.go`
  - Implement full resource with Create, Read, Update, Delete, Import
  - Schema: `transport_id`, `tunnel_id`, `protocol`, `port`
  - Purpose: Terraform resource for IPsec transport mode mappings
  - _Leverage: Existing resource patterns_
  - _Requirements: REQ-7_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create internal/provider/resource_rtx_ipsec_transport.go per REQ-7 | Restrictions: Validate tunnel_id references existing tunnel | Success: Resource compiles, CRUD works correctly | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

- [x] 4.4. Add unit tests for rtx_ipsec_transport
  - File: `internal/provider/resource_rtx_ipsec_transport_test.go`
  - Test CRUD operations, validation
  - Purpose: Ensure IPsec transport resource reliability
  - _Leverage: Existing resource test patterns_
  - _Requirements: REQ-7_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for rtx_ipsec_transport resource | Restrictions: Mock external dependencies | Success: Tests pass with good coverage | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

---

## Phase 5: Integration

- [x] 5.1. Register new resources in provider
  - File: `internal/provider/provider.go`
  - Add `rtx_access_list_ip`, `rtx_access_list_ipv6`, `rtx_l2tp_service`, `rtx_ipsec_transport` to ResourcesMap
  - Purpose: Make resources available in Terraform
  - _Leverage: Existing provider registration pattern_
  - _Requirements: All_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Register all 4 new resources in internal/provider/provider.go ResourcesMap | Restrictions: Maintain alphabetical ordering | Success: Provider compiles, resources are accessible | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

- [x] 5.2. Run full test suite and fix issues
  - Run `go test ./...` and fix any failures
  - Run `go build` to verify compilation
  - Purpose: Ensure all components work together
  - _Leverage: Existing test infrastructure_
  - _Requirements: All_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Run full test suite and build, fix any failures | Restrictions: Do not skip failing tests | Success: All tests pass, build succeeds | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_

- [x] 5.3. Update documentation
  - Add example configurations to `examples/` directory
  - Update any relevant documentation
  - Purpose: Provide usage examples for new resources
  - _Leverage: Existing example patterns_
  - _Requirements: All_
  - _Prompt: Implement the task for spec security-filter-resources, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer | Task: Create example Terraform configurations for all 4 new resources in examples/ directory | Restrictions: Follow existing example patterns | Success: Examples are clear and functional | After completion: Mark task as [-] in-progress before starting, use log-implementation tool to record artifacts, then mark as [x] complete_
