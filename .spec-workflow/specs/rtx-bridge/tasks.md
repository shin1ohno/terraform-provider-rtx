# Tasks Document: rtx_bridge

## Phase 1: Parser Layer

- [ ] 1. Create BridgeConfig data model and parser
  - File: internal/rtx/parsers/bridge.go
  - Define BridgeConfig struct with Name and Members fields
  - Implement ParseBridgeConfig() to parse RTX output
  - Purpose: Parse "show config | grep bridge" output
  - _Leverage: internal/rtx/parsers/dhcp_bindings.go for patterns_
  - _Requirements: 1, 2_
  - _Prompt: Implement the task for spec rtx-bridge, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create BridgeConfig struct with Name (string, e.g., "bridge1") and Members ([]string, e.g., ["lan1", "tunnel1"]) fields. Implement ParseBridgeConfig() function to parse RTX router output from "show config | grep bridge" command. Parse "bridge member bridge1 lan1 tunnel1" format | Restrictions: Do not modify existing parser files, use standard library regexp, handle multiple bridge configurations | _Leverage: internal/rtx/parsers/dhcp_bindings.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (Member Interfaces) | Success: Parser correctly extracts bridge name and member interfaces from sample RTX output, handles bridges with no members | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 2. Create command builder functions for bridge
  - File: internal/rtx/parsers/bridge.go (continue)
  - Implement BuildBridgeMemberCommand() for bridge creation/update
  - Implement BuildDeleteBridgeCommand() for deletion
  - Implement BuildShowBridgeCommand() for reading
  - Purpose: Generate RTX CLI commands for bridge management
  - _Leverage: internal/rtx/parsers/dhcp_bindings.go BuildDHCPBindCommand pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-bridge, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "bridge member <bridge_name> <member1> [<member2> ...]" for creation/update, "no bridge member <bridge_name>" for deletion, "show config | grep bridge" for reading. Handle variable number of member interfaces | Restrictions: Follow existing BuildDHCPBindCommand pattern exactly, validate bridge name format (bridge1, bridge2, etc.) | _Leverage: internal/rtx/parsers/dhcp_bindings.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 3. Create parser unit tests
  - File: internal/rtx/parsers/bridge_test.go
  - Test ParseBridgeConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: empty members, multiple bridges, invalid names
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_bindings_test.go for test patterns_
  - _Requirements: 1, 2_
  - _Prompt: Implement the task for spec rtx-bridge, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for bridge.go. Include test cases for parsing bridge config output with single member, multiple members, no members. Test command building with various member combinations. Test edge cases like empty member list, invalid bridge names | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_bindings_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [ ] 4. Add BridgeConfig type to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add BridgeConfig struct with Name and Members fields
  - Extend Client interface with bridge methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPBinding struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-bridge, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add BridgeConfig struct (Name string, Members []string). Extend Client interface with: GetBridge(ctx, bridgeName) (*BridgeConfig, error), CreateBridge(ctx, bridge) error, UpdateBridge(ctx, bridge) error, DeleteBridge(ctx, bridgeName) error, ListBridges(ctx) ([]BridgeConfig, error) | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 5. Create BridgeService implementation
  - File: internal/client/bridge_service.go (new)
  - Implement BridgeService struct with executor reference
  - Implement Create() with validation and command execution
  - Implement Get() to parse bridge configuration
  - Implement Update() for modifying members
  - Implement Delete() to remove bridge
  - Implement List() to retrieve all bridges
  - Purpose: Service layer for bridge CRUD operations
  - _Leverage: internal/client/dhcp_service.go for service pattern_
  - _Requirements: 1, 2_
  - _Prompt: Implement the task for spec rtx-bridge, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create BridgeService following DHCPService pattern. Include input validation (bridge name format, valid member interface names). Use parsers.BuildBridgeMemberCommand and related functions. Call client.SaveConfig() after modifications. Handle bridge update by replacing member list | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, validate member types (lan*, tunnel*, pp*) | _Leverage: internal/client/dhcp_service.go | _Requirements: Requirements 1-2 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 6. Integrate BridgeService into rtxClient
  - File: internal/client/client.go (modify)
  - Add bridgeService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface bridge methods delegating to service
  - Purpose: Wire up bridge service to main client
  - _Leverage: existing dhcpService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-bridge, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add bridgeService *BridgeService field to rtxClient. Initialize in Dial(): c.bridgeService = NewBridgeService(c.executor, c). Implement GetBridge, CreateBridge, UpdateBridge, DeleteBridge, ListBridges methods delegating to service | Restrictions: Follow existing dhcpService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpService integration | _Requirements: Requirement 1 | Success: Client compiles, all bridge methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 7. Create service unit tests
  - File: internal/client/bridge_service_test.go (new)
  - Test Create with valid and invalid inputs
  - Test Get parsing
  - Test Update behavior
  - Test Delete
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_service_test.go for patterns_
  - _Requirements: 1, 2_
  - _Prompt: Implement the task for spec rtx-bridge, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for BridgeService. Mock Executor interface to simulate RTX responses. Test validation (invalid bridge name, invalid member interface names). Test successful CRUD operations. Test error handling for member already in another bridge | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_service_test.go | _Requirements: Requirements 1-2 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [ ] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_bridge.go (new)
  - Define resourceRTXBridge() with full schema
  - Add name (Required, ForceNew, String, validated as bridge*)
  - Add members (Optional, List of String, validated interface names)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_binding.go for patterns_
  - _Requirements: 1, 2_
  - _Prompt: Implement the task for spec rtx-bridge, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXBridge() returning *schema.Resource. Define schema following rtx_dhcp_binding patterns. Add ValidateFunc for name (must match bridge[0-9]+), members (valid interface names: lan*, tunnel*, pp*). Set ForceNew on name. Use TypeList for members | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_binding.go | _Requirements: Requirements 1-2 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_bridge.go (continue)
  - Implement resourceRTXBridgeCreate()
  - Implement resourceRTXBridgeRead()
  - Implement resourceRTXBridgeUpdate()
  - Implement resourceRTXBridgeDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_binding.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-bridge, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build BridgeConfig from ResourceData, call client.CreateBridge, set ID to name). Read (call GetBridge, update ResourceData, handle not found by clearing ID). Update (call UpdateBridge for member changes). Delete (call DeleteBridge). Follow rtx_dhcp_binding patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully | _Leverage: internal/provider/resource_rtx_dhcp_binding.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 10. Implement import functionality
  - File: internal/provider/resource_rtx_bridge.go (continue)
  - Implement resourceRTXBridgeImport()
  - Parse bridge name from import ID string
  - Validate bridge exists on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_binding.go import pattern_
  - _Requirements: 3_
  - _Prompt: Implement the task for spec rtx-bridge, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXBridgeImport(). Parse import ID as bridge name (e.g., "bridge1"). Call GetBridge to verify existence. Populate all ResourceData fields from retrieved config. Call Read to ensure state consistency | Restrictions: Handle invalid import ID format, non-existent bridge errors gracefully | _Leverage: internal/provider/resource_rtx_dhcp_binding.go import function | _Requirements: Requirement 3 (Import) | Success: terraform import rtx_bridge.example bridge1 works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_bridge" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-bridge, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_bridge": resourceRTXBridge() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 12. Create resource unit tests
  - File: internal/provider/resource_rtx_bridge_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_binding_test.go patterns_
  - _Requirements: 1, 3_
  - _Prompt: Implement the task for spec rtx-bridge, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_bridge.go. Test schema validation (invalid bridge name, invalid member names). Test CRUD operations with mocked client. Test import with valid and invalid bridge names | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_binding_test.go | _Requirements: Requirements 1, 3 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [ ] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_bridge_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test bridge creation with single member
  - Test bridge creation with multiple members
  - Test member addition/removal
  - Test bridge import
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_binding_test.go acceptance test patterns_
  - _Requirements: 1, 3, 4_
  - _Prompt: Implement the task for spec rtx-bridge, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with bridge creation with lan1 member, add tunnel1 member, import existing bridge. Test L2VPN scenario with multiple tunnel members. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 3, 4 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 14. Add example Terraform configurations
  - File: examples/bridge/main.tf (new)
  - Basic bridge creation with single member example
  - Bridge with multiple members example
  - L2VPN bridge with L2TPv3 tunnels example
  - Bridge with rtx_interface IP configuration example
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp/ existing examples_
  - _Requirements: 1, 4_
  - _Prompt: Implement the task for spec rtx-bridge, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: bridge with single lan member. Multi-member: bridge with lan1 and tunnel1. L2VPN: bridge with multiple L2TPv3 tunnels for site-to-site L2 connectivity. Integration: rtx_bridge with rtx_interface for IP assignment | Restrictions: Use realistic interface names, include comments explaining bridge use cases | _Leverage: examples/dhcp/ | _Requirements: Requirements 1, 4 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-bridge, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works. Check terraform import functionality. Ensure no regressions in existing resources. Test L2TPv3 integration scenario | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
