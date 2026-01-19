# Tasks Document: rtx_ipv6_prefix

## Phase 1: Parser Layer

- [ ] 1. Create IPv6Prefix data model and parser
  - File: internal/rtx/parsers/ipv6_prefix.go
  - Define IPv6Prefix struct with ID, Prefix, PrefixLength, Source, Interface fields
  - Implement ParseIPv6PrefixConfig() to parse RTX output
  - Purpose: Parse "show config | grep ipv6 prefix" output
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-ipv6-prefix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create IPv6Prefix struct with ID (int, 1-255), Prefix (string), PrefixLength (int), Source (string: "static", "ra", "dhcpv6-pd"), Interface (string, optional). Implement ParseIPv6PrefixConfig() function to parse RTX router output from "show config | grep ipv6 prefix" command. Handle three prefix types: static (e.g., "ipv6 prefix 1 2001:db8:1234::/64"), RA-derived (e.g., "ipv6 prefix 1 ra-prefix@lan2::/64"), DHCPv6-PD (e.g., "ipv6 prefix 1 dhcp-prefix@lan2::/48") | Restrictions: Do not modify existing parser files, use standard library regexp, handle all three source types | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (Static Prefix), Requirement 3 (Dynamic Prefix) | Success: Parser correctly extracts all prefix attributes from sample RTX output, handles all three source types | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 2. Create command builder functions for IPv6 prefix
  - File: internal/rtx/parsers/ipv6_prefix.go (continue)
  - Implement BuildIPv6PrefixCommand() for prefix creation
  - Implement BuildDeleteIPv6PrefixCommand() for deletion
  - Implement BuildShowIPv6PrefixCommand() for reading
  - Purpose: Generate RTX CLI commands for prefix management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go BuildDHCPScopeCommand pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ipv6-prefix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "ipv6 prefix <id> <prefix>::/<length>" for static, "ipv6 prefix <id> ra-prefix@<interface>::/<length>" for RA-derived, "ipv6 prefix <id> dhcp-prefix@<interface>::/<length>" for DHCPv6-PD, "no ipv6 prefix <id>" for deletion, "show config | grep ipv6 prefix" for reading | Restrictions: Follow existing command builder patterns exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax for all three prefix source types | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 3. Create parser unit tests
  - File: internal/rtx/parsers/ipv6_prefix_test.go
  - Test ParseIPv6PrefixConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: all three source types, invalid prefix IDs
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-ipv6-prefix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for ipv6_prefix.go. Include test cases for parsing static prefixes, RA-derived prefixes, DHCPv6-PD prefixes, command building for all source types, edge cases like invalid prefix length (>128), invalid prefix ID (>255) | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, all three source types tested | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [ ] 4. Add IPv6Prefix type to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add IPv6Prefix struct with all fields
  - Extend Client interface with IPv6 prefix methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ipv6-prefix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add IPv6Prefix struct (ID int, Prefix string, PrefixLength int, Source string, Interface string). Extend Client interface with: GetIPv6Prefix(ctx, prefixID) (*IPv6Prefix, error), CreateIPv6Prefix(ctx, prefix) error, UpdateIPv6Prefix(ctx, prefix) error, DeleteIPv6Prefix(ctx, prefixID) error, ListIPv6Prefixes(ctx) ([]IPv6Prefix, error) | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 5. Create IPv6PrefixService implementation
  - File: internal/client/ipv6_prefix_service.go (new)
  - Implement IPv6PrefixService struct with executor reference
  - Implement Create() with validation and command execution
  - Implement Get() to parse prefix configuration
  - Implement Update() for modifying prefix settings
  - Implement Delete() for prefix removal
  - Implement List() to retrieve all prefixes
  - Purpose: Service layer for prefix CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ipv6-prefix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create IPv6PrefixService following DHCPScopeService pattern. Include input validation (prefix ID 1-255, prefix length 1-128, valid IPv6 format for static, valid interface for ra/dhcpv6-pd). Use parsers.BuildIPv6PrefixCommand and related functions. Call client.SaveConfig() after modifications. Handle prefix update by deleting and recreating since source type change requires ForceNew | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, validate source-specific requirements (interface required for ra/dhcpv6-pd) | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-4 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 6. Integrate IPv6PrefixService into rtxClient
  - File: internal/client/client.go (modify)
  - Add ipv6PrefixService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface IPv6 prefix methods delegating to service
  - Purpose: Wire up prefix service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ipv6-prefix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add ipv6PrefixService *IPv6PrefixService field to rtxClient. Initialize in Dial(): c.ipv6PrefixService = NewIPv6PrefixService(c.executor, c). Implement GetIPv6Prefix, CreateIPv6Prefix, UpdateIPv6Prefix, DeleteIPv6Prefix, ListIPv6Prefixes methods delegating to service | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all IPv6 prefix methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 7. Create service unit tests
  - File: internal/client/ipv6_prefix_service_test.go (new)
  - Test Create with valid and invalid inputs for all source types
  - Test Get parsing for all prefix types
  - Test Update behavior
  - Test Delete
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-ipv6-prefix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for IPv6PrefixService. Mock Executor interface to simulate RTX responses. Test validation (invalid prefix ID >255, prefix length >128, missing interface for ra/dhcpv6-pd, invalid IPv6 format). Test successful CRUD operations for all three source types. Test error handling | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-3 | Success: All tests pass, validation logic tested, error paths covered for all source types | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [ ] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_ipv6_prefix.go (new)
  - Define resourceRTXIPv6Prefix() with full schema
  - Add prefix_id (Required, ForceNew, Int, range 1-255)
  - Add source (Required, ForceNew, String: "static", "ra", "dhcpv6-pd")
  - Add prefix (Optional, String, required for static source)
  - Add prefix_length (Required, Int, range 1-128)
  - Add interface (Optional, String, required for ra/dhcpv6-pd sources)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-ipv6-prefix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXIPv6Prefix() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Add ValidateFunc for prefix_id (1-255), prefix_length (1-128), source (enum: static, ra, dhcpv6-pd). Set ForceNew on prefix_id and source. Add ConflictsWith for prefix (conflicts with interface for static) and interface (required for ra/dhcpv6-pd) | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style, implement custom validation for source-dependent required fields | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1-3 | Success: Schema compiles, validation functions work for all source types | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_ipv6_prefix.go (continue)
  - Implement resourceRTXIPv6PrefixCreate()
  - Implement resourceRTXIPv6PrefixRead()
  - Implement resourceRTXIPv6PrefixUpdate()
  - Implement resourceRTXIPv6PrefixDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ipv6-prefix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build IPv6Prefix from ResourceData based on source type, call client.CreateIPv6Prefix, set ID to prefix_id). Read (call GetIPv6Prefix, update ResourceData, handle not found by clearing ID). Update (call UpdateIPv6Prefix for mutable fields like prefix_length for static). Delete (call DeleteIPv6Prefix). Follow rtx_dhcp_scope patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully, validate source-specific fields in Create | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end for all three source types | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 10. Implement import functionality
  - File: internal/provider/resource_rtx_ipv6_prefix.go (continue)
  - Implement resourceRTXIPv6PrefixImport()
  - Parse prefix_id from import ID string
  - Validate prefix exists on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_scope.go import pattern_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec rtx-ipv6-prefix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXIPv6PrefixImport(). Parse import ID as prefix_id integer. Call GetIPv6Prefix to verify existence. Populate all ResourceData fields from retrieved prefix including source type detection. Call Read to ensure state consistency | Restrictions: Handle invalid import ID format, non-existent prefix errors gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go import function | _Requirements: Requirement 5 (Import) | Success: terraform import rtx_ipv6_prefix.example 1 works correctly for all source types | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_ipv6_prefix" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ipv6-prefix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_ipv6_prefix": resourceRTXIPv6Prefix() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 12. Create resource unit tests
  - File: internal/provider/resource_rtx_ipv6_prefix_test.go (new)
  - Test schema validation for all source types
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-ipv6-prefix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_ipv6_prefix.go. Test schema validation (invalid prefix ID, invalid prefix length, missing interface for ra source). Test CRUD operations with mocked client for all three source types. Test import with valid and invalid IDs | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths for all source types | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [ ] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_ipv6_prefix_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test static prefix creation
  - Test RA-derived prefix creation
  - Test DHCPv6-PD prefix creation
  - Test prefix import
  - Test dependency with rtx_ipv6_interface
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_acc_test.go acceptance test patterns_
  - _Requirements: 1, 5, 6_
  - _Prompt: Implement the task for spec rtx-ipv6-prefix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with static prefix creation, RA-derived prefix creation, DHCPv6-PD prefix creation, update prefix length, import existing prefix. Test rtx_ipv6_interface depends_on rtx_ipv6_prefix. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5, 6 | Success: Acceptance tests pass against real RTX router for all source types | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 14. Add example Terraform configurations
  - File: examples/ipv6_prefix/main.tf (new)
  - Static prefix example
  - RA-derived prefix example (for IPoE/MAP-E ISP integration)
  - DHCPv6-PD prefix example
  - Integration with rtx_ipv6_interface example
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1, 6_
  - _Prompt: Implement the task for spec rtx-ipv6-prefix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Static: prefix with explicit IPv6 address. RA: prefix derived from upstream router (Japanese ISP IPoE/MAP-E scenario). DHCPv6-PD: prefix delegation from ISP. Integration: rtx_ipv6_prefix with rtx_ipv6_interface showing prefix reference in interface address assignment | Restrictions: Use realistic IPv6 addresses (2001:db8::/32 documentation range), include comments explaining Japanese ISP integration scenarios | _Leverage: examples/dhcp_scope/ | _Requirements: Requirements 1, 6 | Success: Examples are valid Terraform, demonstrate all three source types and interface integration | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-ipv6-prefix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works for all three source types. Check terraform import functionality. Ensure no regressions in existing resources including rtx_ipv6_interface integration | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end for static, RA, and DHCPv6-PD prefixes | After completing, use log-implementation tool to record details, then mark as [x] complete_
