# Tasks Document: rtx_dhcp_scope

## Phase 1: Parser Layer

- [x] 1. Create DHCPScope data model and parser
  - File: internal/rtx/parsers/dhcp_scope.go
  - Define DHCPScope and ExcludeRange structs
  - Implement ParseScopeConfig() to parse RTX output
  - Purpose: Parse "show config | grep dhcp scope" output
  - _Leverage: internal/rtx/parsers/dhcp_bindings.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-dhcp-scope, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create DHCPScope struct with ScopeID, Network, Gateway, DNSServers, LeaseTime, ExcludeRanges fields. Implement ParseScopeConfig() function to parse RTX router output from "show config | grep dhcp scope" command. Follow patterns from dhcp_bindings.go | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line scope configurations | _Leverage: internal/rtx/parsers/dhcp_bindings.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (IP Range), Requirement 3 (Gateway/DNS), Requirement 4 (Lease Time) | Success: Parser correctly extracts all scope attributes from sample RTX output, handles edge cases like missing optional fields | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [x] 2. Create command builder functions for DHCP scope
  - File: internal/rtx/parsers/dhcp_scope.go (continue)
  - Implement BuildDHCPScopeCommand() for scope creation
  - Implement BuildDHCPScopeOptionsCommand() for DNS configuration
  - Implement BuildDHCPScopeExceptCommand() for exclusion ranges
  - Implement BuildDeleteDHCPScopeCommand() for deletion
  - Implement BuildShowDHCPScopeCommand() for reading
  - Purpose: Generate RTX CLI commands for scope management
  - _Leverage: internal/rtx/parsers/dhcp_bindings.go BuildDHCPBindCommand pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-dhcp-scope, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "dhcp scope <id> <network>/<prefix> [gateway <gw>] [expire <time>]", "dhcp scope option <id> dns=<dns1>,<dns2>", "dhcp scope <id> except <start>-<end>", "no dhcp scope <id>". Convert Go duration to RTX format (72h -> 72:00) | Restrictions: Follow existing BuildDHCPBindCommand pattern exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_bindings.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax, lease time conversion works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 3. Create parser unit tests
  - File: internal/rtx/parsers/dhcp_scope_test.go
  - Test ParseScopeConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: missing fields, malformed input
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_bindings_test.go for test patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-dhcp-scope, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for dhcp_scope.go. Include test cases for parsing scope config output, command building with various parameter combinations, edge cases like empty DNS list, infinite lease time | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_bindings_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [x] 4. Add DHCPScope type to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add DHCPScope struct with all fields
  - Add ExcludeRange struct
  - Extend Client interface with scope methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPBinding struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-dhcp-scope, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add DHCPScope struct (ScopeID int, Network string, Gateway string, DNSServers []string, LeaseTime string, ExcludeRanges []ExcludeRange). Add ExcludeRange struct (Start, End string). Extend Client interface with: GetDHCPScope(ctx, scopeID) (*DHCPScope, error), CreateDHCPScope(ctx, scope) error, UpdateDHCPScope(ctx, scope) error, DeleteDHCPScope(ctx, scopeID) error, ListDHCPScopes(ctx) ([]DHCPScope, error) | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 5. Create DHCPScopeService implementation
  - File: internal/client/dhcp_scope_service.go (new)
  - Implement DHCPScopeService struct with executor reference
  - Implement CreateScope() with validation and command execution
  - Implement GetScope() to parse scope configuration
  - Implement UpdateScope() for modifying options
  - Implement DeleteScope() with binding check warning
  - Implement ListScopes() to retrieve all scopes
  - Purpose: Service layer for scope CRUD operations
  - _Leverage: internal/client/dhcp_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4, 5_
  - _Prompt: Implement the task for spec rtx-dhcp-scope, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create DHCPScopeService following DHCPService pattern. Include input validation (CIDR format, IP addresses, max 3 DNS servers). Use parsers.BuildDHCPScopeCommand and related functions. Call client.SaveConfig() after modifications. Handle scope update by deleting and recreating if network changes | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, maintain separation from DHCPService | _Leverage: internal/client/dhcp_service.go | _Requirements: Requirements 1-5 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 6. Integrate DHCPScopeService into rtxClient
  - File: internal/client/client.go (modify)
  - Add dhcpScopeService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface scope methods delegating to service
  - Purpose: Wire up scope service to main client
  - _Leverage: existing dhcpService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-dhcp-scope, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add dhcpScopeService *DHCPScopeService field to rtxClient. Initialize in Dial(): c.dhcpScopeService = NewDHCPScopeService(c.executor, c). Implement GetDHCPScope, CreateDHCPScope, UpdateDHCPScope, DeleteDHCPScope, ListDHCPScopes methods delegating to service | Restrictions: Follow existing dhcpService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpService integration | _Requirements: Requirement 1 | Success: Client compiles, all scope methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 7. Create service unit tests (partial - mock updates for interface compatibility)
  - File: internal/client/dhcp_scope_service_test.go (new)
  - Test CreateScope with valid and invalid inputs
  - Test GetScope parsing
  - Test UpdateScope behavior
  - Test DeleteScope
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-dhcp-scope, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for DHCPScopeService. Mock Executor interface to simulate RTX responses. Test validation (invalid CIDR, too many DNS servers, invalid IPs). Test successful CRUD operations. Test error handling | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [x] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_dhcp_scope.go (new)
  - Define resourceRTXDHCPScope() with full schema
  - Add scope_id (Required, ForceNew, Int)
  - Add network (Required, ForceNew, String with CIDR validation)
  - Add gateway (Optional, String)
  - Add dns_servers (Optional, List of String, max 3)
  - Add lease_time (Optional, String, default "72h")
  - Add exclude_ranges (Optional, List of Object with start/end)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_binding.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-dhcp-scope, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXDHCPScope() returning *schema.Resource. Define schema following rtx_dhcp_binding patterns. Add ValidateFunc for network (CIDR), dns_servers (max 3, valid IPs). Set ForceNew on scope_id and network. Use TypeList for exclude_ranges with nested schema (start, end strings) | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_binding.go | _Requirements: Requirements 1-4 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_dhcp_scope.go (continue)
  - Implement resourceRTXDHCPScopeCreate()
  - Implement resourceRTXDHCPScopeRead()
  - Implement resourceRTXDHCPScopeUpdate()
  - Implement resourceRTXDHCPScopeDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_binding.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-dhcp-scope, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build DHCPScope from ResourceData, call client.CreateDHCPScope, set ID to scope_id). Read (call GetDHCPScope, update ResourceData, handle not found by clearing ID). Update (call UpdateDHCPScope for mutable fields). Delete (call DeleteDHCPScope). Follow rtx_dhcp_binding patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully | _Leverage: internal/provider/resource_rtx_dhcp_binding.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 10. Implement import functionality
  - File: internal/provider/resource_rtx_dhcp_scope.go (continue)
  - Implement resourceRTXDHCPScopeImport()
  - Parse scope_id from import ID string
  - Validate scope exists on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_binding.go import pattern_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec rtx-dhcp-scope, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXDHCPScopeImport(). Parse import ID as scope_id integer. Call GetDHCPScope to verify existence. Populate all ResourceData fields from retrieved scope. Call Read to ensure state consistency | Restrictions: Handle invalid import ID format, non-existent scope errors gracefully | _Leverage: internal/provider/resource_rtx_dhcp_binding.go import function | _Requirements: Requirement 5 (Import) | Success: terraform import rtx_dhcp_scope.example 1 works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_dhcp_scope" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-dhcp-scope, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_dhcp_scope": resourceRTXDHCPScope() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 12. Create resource unit tests (mock updates for interface compatibility)
  - File: internal/provider/resource_rtx_dhcp_scope_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_binding_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-dhcp-scope, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_dhcp_scope.go. Test schema validation (invalid CIDR, too many DNS servers). Test CRUD operations with mocked client. Test import with valid and invalid IDs | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_binding_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [x] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_dhcp_scope_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test scope creation with all parameters
  - Test scope update
  - Test scope import
  - Test dependency with rtx_dhcp_binding
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_binding_test.go acceptance test patterns_
  - _Requirements: 1, 5, 6_
  - _Prompt: Implement the task for spec rtx-dhcp-scope, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with scope creation, update DNS servers, import existing scope. Test rtx_dhcp_binding depends_on rtx_dhcp_scope. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5, 6 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 14. Add example Terraform configurations
  - File: examples/dhcp_scope/main.tf (new)
  - Basic scope creation example
  - Scope with all options example
  - Scope with binding dependency example
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp/ existing examples_
  - _Requirements: 1, 6_
  - _Prompt: Implement the task for spec rtx-dhcp-scope, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: scope with network only. Full: scope with gateway, dns_servers, lease_time, exclude_ranges. Integration: rtx_dhcp_scope with rtx_dhcp_binding showing dependency | Restrictions: Use realistic IP addresses, include comments explaining options | _Leverage: examples/dhcp/ | _Requirements: Requirements 1, 6 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-dhcp-scope, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works. Check terraform import functionality. Ensure no regressions in existing resources | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
