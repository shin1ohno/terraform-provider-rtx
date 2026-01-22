# Tasks Document: rtx_ospf

## Phase 1: Parser Layer

- [x] 1. Create OSPFConfig data model and parser
  - File: internal/rtx/parsers/ospf.go
  - Define OSPFConfig, OSPFNetwork, OSPFArea, OSPFNeighbor structs
  - Implement ParseOSPFConfig() to parse RTX output
  - Purpose: Parse "show config | grep ospf" output
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ospf, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create OSPFConfig struct with ProcessID, RouterID, Distance, DefaultOriginate, Networks, Areas, Neighbors, RedistributeStatic, RedistributeConnected fields. Create OSPFNetwork (IP, Wildcard, Area), OSPFArea (ID, Type, NoSummary), OSPFNeighbor (IP, Priority, Cost) structs. Implement ParseOSPFConfig() function to parse RTX router output from "show config | grep ospf" command. Follow patterns from dhcp_scope.go | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line OSPF configurations | _Leverage: internal/rtx/parsers/dhcp_scope.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (Router ID), Requirement 3 (Areas), Requirement 4 (Networks) | Success: Parser correctly extracts all OSPF attributes from sample RTX output, handles edge cases like missing optional fields | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [x] 2. Create command builder functions for OSPF
  - File: internal/rtx/parsers/ospf.go (continue)
  - Implement BuildOSPFEnableCommand() for enabling OSPF
  - Implement BuildOSPFRouterIDCommand() for router ID configuration
  - Implement BuildOSPFAreaCommand() for area configuration
  - Implement BuildOSPFImportCommand() for redistribution
  - Implement BuildDeleteOSPFCommand() for deletion
  - Implement BuildShowOSPFCommand() for reading
  - Purpose: Generate RTX CLI commands for OSPF management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go BuildDHCPScopeCommand pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ospf, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "ospf use on", "ospf router id <router_id>", "ip <interface> ospf area <area>", "ospf area <area> stub [no-summary]", "ospf import from static", "ospf use off". Validate router ID as valid IPv4 format, area ID as decimal or dotted decimal | Restrictions: Follow existing BuildDHCPScopeCommand pattern exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax, validation catches invalid router ID and area formats | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 3. Create parser unit tests
  - File: internal/rtx/parsers/ospf_test.go
  - Test ParseOSPFConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: missing fields, malformed input
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ospf, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for ospf.go. Include test cases for parsing OSPF config output, command building with various parameter combinations, edge cases like empty networks list, stub areas with no-summary | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [x] 4. Add OSPFConfig type to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add OSPFConfig struct with all fields
  - Add OSPFNetwork, OSPFArea, OSPFNeighbor structs
  - Extend Client interface with OSPF methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ospf, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add OSPFConfig struct (ProcessID int, RouterID string, Distance int, DefaultOriginate bool, Networks []OSPFNetwork, Areas []OSPFArea, Neighbors []OSPFNeighbor, RedistributeStatic bool, RedistributeConnected bool). Add OSPFNetwork (IP, Wildcard, Area string), OSPFArea (ID, Type string, NoSummary bool), OSPFNeighbor (IP string, Priority int, Cost int) structs. Extend Client interface with: GetOSPF(ctx) (*OSPFConfig, error), CreateOSPF(ctx, ospf) error, UpdateOSPF(ctx, ospf) error, DeleteOSPF(ctx) error | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 5. Create OSPFService implementation
  - File: internal/client/ospf_service.go (new)
  - Implement OSPFService struct with executor reference
  - Implement Create() with validation and command execution
  - Implement Get() to parse OSPF configuration
  - Implement Update() for modifying OSPF settings
  - Implement Delete() to disable OSPF
  - Purpose: Service layer for OSPF CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4, 5_
  - _Prompt: Implement the task for spec rtx-ospf, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create OSPFService following DHCPScopeService pattern. Include input validation (router ID as valid IPv4, area ID format). Use parsers.BuildOSPFEnableCommand and related functions. Call client.SaveConfig() after modifications. Handle singleton resource (only one OSPF config per router) | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, maintain separation from other services | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-5 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 6. Integrate OSPFService into rtxClient
  - File: internal/client/client.go (modify)
  - Add ospfService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface OSPF methods delegating to service
  - Purpose: Wire up OSPF service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ospf, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add ospfService *OSPFService field to rtxClient. Initialize in Dial(): c.ospfService = NewOSPFService(c.executor, c). Implement GetOSPF, CreateOSPF, UpdateOSPF, DeleteOSPF methods delegating to service | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all OSPF methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 7. Create service unit tests
  - File: internal/client/ospf_service_test.go (new)
  - Test Create with valid and invalid inputs
  - Test Get parsing
  - Test Update behavior
  - Test Delete
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ospf, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for OSPFService. Mock Executor interface to simulate RTX responses. Test validation (invalid router ID, invalid area ID). Test successful CRUD operations. Test error handling for singleton resource | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [x] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_ospf.go (new)
  - Define resourceRTXOSPF() with full schema
  - Add process_id (Optional, Int, default 1)
  - Add router_id (Required, String with IPv4 validation)
  - Add distance (Optional, Int, default 110)
  - Add default_information_originate (Optional, Bool)
  - Add networks (Optional, List of Object with ip/wildcard/area)
  - Add areas (Optional, List of Object with id/type/no_summary)
  - Add neighbors (Optional, List of Object with ip/priority/cost)
  - Add redistribute_static (Optional, Bool)
  - Add redistribute_connected (Optional, Bool)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ospf, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXOSPF() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Add ValidateFunc for router_id (valid IPv4), area id (decimal or dotted decimal). Use TypeList for networks with nested schema (ip, wildcard, area strings), areas (id, type strings, no_summary bool), neighbors (ip string, priority/cost int) | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1-4 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_ospf.go (continue)
  - Implement resourceRTXOSPFCreate()
  - Implement resourceRTXOSPFRead()
  - Implement resourceRTXOSPFUpdate()
  - Implement resourceRTXOSPFDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ospf, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build OSPFConfig from ResourceData, call client.CreateOSPF, set ID to "ospf" as singleton). Read (call GetOSPF, update ResourceData, handle not found by clearing ID). Update (call UpdateOSPF for mutable fields). Delete (call DeleteOSPF). Follow rtx_dhcp_scope patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully, use fixed ID "ospf" for singleton resource | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 10. Implement import functionality
  - File: internal/provider/resource_rtx_ospf.go (continue)
  - Implement resourceRTXOSPFImport()
  - Validate OSPF exists on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_scope.go import pattern_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec rtx-ospf, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXOSPFImport(). For singleton resource, import ID should be "ospf". Call GetOSPF to verify existence. Populate all ResourceData fields from retrieved config. Call Read to ensure state consistency | Restrictions: Handle non-existent OSPF config errors gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go import function | _Requirements: Requirement 5 (Import) | Success: terraform import rtx_ospf.example ospf works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_ospf" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ospf, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_ospf": resourceRTXOSPF() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 12. Create resource unit tests
  - File: internal/provider/resource_rtx_ospf_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-ospf, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_ospf.go. Test schema validation (invalid router ID, invalid area ID). Test CRUD operations with mocked client. Test import with valid and invalid IDs | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [x] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_ospf_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test OSPF creation with single area
  - Test OSPF with multiple areas
  - Test stub area configuration
  - Test redistribution settings
  - Test OSPF update
  - Test OSPF import
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_acc_test.go acceptance test patterns_
  - _Requirements: 1, 5, 6_
  - _Prompt: Implement the task for spec rtx-ospf, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with single area OSPF creation, multi-area configuration, stub area settings, redistribution options, import existing OSPF. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5, 6 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 14. Add example Terraform configurations
  - File: examples/ospf/main.tf (new)
  - Basic OSPF with single area example
  - OSPF with multiple areas example
  - OSPF with stub area and redistribution example
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1, 6_
  - _Prompt: Implement the task for spec rtx-ospf, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: OSPF with router_id and single backbone area. Multi-area: OSPF with backbone and additional areas. Full: OSPF with stub areas, redistribution, and neighbors | Restrictions: Use realistic IP addresses, include comments explaining OSPF concepts | _Leverage: examples/dhcp_scope/ | _Requirements: Requirements 1, 6 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-ospf, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works. Check terraform import functionality. Ensure no regressions in existing resources | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
