# Tasks Document: rtx_nat_masquerade

## Phase 1: Parser Layer

- [x] 1. Create NATMasquerade data model and parser
  - File: internal/rtx/parsers/nat_masquerade.go
  - Define NATMasquerade and StaticEntry structs
  - Implement ParseNATDescriptorConfig() to parse RTX output
  - Purpose: Parse "show config | grep nat descriptor" output
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-nat-masquerade, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create NATMasquerade struct with DescriptorID, OuterAddress, InnerNetwork, Interface, StaticEntries fields. Create StaticEntry struct with InsideLocal, InsideLocalPort, OutsideGlobal, OutsideGlobalPort, Protocol fields. Implement ParseNATDescriptorConfig() function to parse RTX router output from "show config | grep nat descriptor" command. Handle masquerade type descriptors, address outer/inner configs, and static mappings | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line configurations | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (Descriptor Config), Requirement 3 (Interface Binding) | Success: Parser correctly extracts all NAT descriptor attributes from sample RTX output, handles edge cases like missing optional fields | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [x] 2. Create command builder functions for NAT masquerade
  - File: internal/rtx/parsers/nat_masquerade.go (continue)
  - Implement BuildNATDescriptorTypeCommand() for descriptor type
  - Implement BuildNATDescriptorAddressOuterCommand() for outer address
  - Implement BuildNATDescriptorAddressInnerCommand() for inner network
  - Implement BuildNATDescriptorStaticCommand() for static mappings
  - Implement BuildInterfaceNATCommand() for interface binding
  - Implement BuildDeleteNATDescriptorCommand() for deletion
  - Purpose: Generate RTX CLI commands for NAT descriptor management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go BuildDHCPScopeCommand pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-nat-masquerade, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "nat descriptor type <id> masquerade", "nat descriptor address outer <id> <interface/ip>", "nat descriptor address inner <id> <network>", "nat descriptor masquerade static <id> <n> <outer_ip>:<port>=<inner_ip>:<port> [<protocol>]", "ip <interface> nat descriptor <id>", "no nat descriptor type <id>". Convert CIDR to RTX range format (192.168.1.0/24 -> 192.168.1.0-192.168.1.255) | Restrictions: Follow existing command builder pattern exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax, network conversion works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 3. Create parser unit tests
  - File: internal/rtx/parsers/nat_masquerade_test.go
  - Test ParseNATDescriptorConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: missing fields, malformed input, static entries
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-nat-masquerade, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for nat_masquerade.go. Include test cases for parsing NAT descriptor config output, command building with various parameter combinations, edge cases like empty static entries, ipcp outer address, different protocol types | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [x] 4. Add NATMasquerade type to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add NATMasquerade struct with all fields
  - Add StaticEntry struct
  - Extend Client interface with NAT masquerade methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-nat-masquerade, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add NATMasquerade struct (DescriptorID int, OuterAddress string, InnerNetwork string, Interface string, StaticEntries []StaticEntry). Add StaticEntry struct (InsideLocal string, InsideLocalPort int, OutsideGlobal string, OutsideGlobalPort int, Protocol string). Extend Client interface with: GetNATMasquerade(ctx, descriptorID) (*NATMasquerade, error), CreateNATMasquerade(ctx, nat) error, UpdateNATMasquerade(ctx, nat) error, DeleteNATMasquerade(ctx, descriptorID) error, ListNATMasquerades(ctx) ([]NATMasquerade, error) | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 5. Create NATMasqueradeService implementation
  - File: internal/client/nat_masquerade_service.go (new)
  - Implement NATMasqueradeService struct with executor reference
  - Implement Create() with validation and command execution
  - Implement Get() to parse NAT descriptor configuration
  - Implement Update() for modifying descriptor options
  - Implement Delete() with interface unbinding
  - Implement List() to retrieve all NAT masquerade descriptors
  - Purpose: Service layer for NAT masquerade CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-nat-masquerade, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create NATMasqueradeService following DHCPScopeService pattern. Include input validation (descriptor ID 1-65535, valid IP/CIDR format, interface names). Use parsers.BuildNATDescriptorTypeCommand and related functions. Call client.SaveConfig() after modifications. Handle descriptor update by modifying existing config. Ensure interface binding/unbinding is handled properly | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, maintain separation from other services | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-4 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 6. Integrate NATMasqueradeService into rtxClient
  - File: internal/client/client.go (modify)
  - Add natMasqueradeService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface NAT masquerade methods delegating to service
  - Purpose: Wire up NAT masquerade service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-nat-masquerade, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add natMasqueradeService *NATMasqueradeService field to rtxClient. Initialize in Dial(): c.natMasqueradeService = NewNATMasqueradeService(c.executor, c). Implement GetNATMasquerade, CreateNATMasquerade, UpdateNATMasquerade, DeleteNATMasquerade, ListNATMasquerades methods delegating to service | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all NAT masquerade methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 7. Create service unit tests
  - File: internal/client/nat_masquerade_service_test.go (new)
  - Test Create with valid and invalid inputs
  - Test Get parsing
  - Test Update behavior
  - Test Delete with interface unbinding
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-nat-masquerade, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for NATMasqueradeService. Mock Executor interface to simulate RTX responses. Test validation (invalid descriptor ID, invalid CIDR, invalid interface). Test successful CRUD operations. Test error handling for port conflicts | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [x] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_nat_masquerade.go (new)
  - Define resourceRTXNATMasquerade() with full schema
  - Add id (Required, ForceNew, Int) for descriptor ID
  - Add inside_source block (Required)
    - Add acl (Required, String with CIDR validation) for inner network
    - Add interface (Required, String) for outside interface
    - Add overload (Optional, Bool, default true)
  - Add static_entries (Optional, List of Object)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-nat-masquerade, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXNATMasquerade() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Add ValidateFunc for acl (CIDR format), id (1-65535 range). Set ForceNew on id. Use TypeList for inside_source block with nested schema (acl, interface, overload). Use TypeList for static_entries with nested schema (inside_local, inside_local_port, outside_global, outside_global_port, protocol) | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1-3 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_nat_masquerade.go (continue)
  - Implement resourceRTXNATMasqueradeCreate()
  - Implement resourceRTXNATMasqueradeRead()
  - Implement resourceRTXNATMasqueradeUpdate()
  - Implement resourceRTXNATMasqueradeDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-nat-masquerade, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build NATMasquerade from ResourceData, call client.CreateNATMasquerade, set ID to descriptor_id). Read (call GetNATMasquerade, update ResourceData, handle not found by clearing ID). Update (call UpdateNATMasquerade for mutable fields like static_entries). Delete (call DeleteNATMasquerade). Follow rtx_dhcp_scope patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 10. Implement import functionality
  - File: internal/provider/resource_rtx_nat_masquerade.go (continue)
  - Implement resourceRTXNATMasqueradeImport()
  - Parse descriptor_id from import ID string
  - Validate descriptor exists on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_scope.go import pattern_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec rtx-nat-masquerade, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXNATMasqueradeImport(). Parse import ID as descriptor_id integer. Call GetNATMasquerade to verify existence. Populate all ResourceData fields from retrieved NAT config. Call Read to ensure state consistency | Restrictions: Handle invalid import ID format, non-existent descriptor errors gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go import function | _Requirements: Requirement 5 (Import) | Success: terraform import rtx_nat_masquerade.example 1 works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_nat_masquerade" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-nat-masquerade, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_nat_masquerade": resourceRTXNATMasquerade() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 12. Create resource unit tests
  - File: internal/provider/resource_rtx_nat_masquerade_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-nat-masquerade, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_nat_masquerade.go. Test schema validation (invalid CIDR, invalid descriptor ID range). Test CRUD operations with mocked client. Test import with valid and invalid IDs | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [x] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_nat_masquerade_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test NAT masquerade creation with all parameters
  - Test NAT masquerade update (static entries)
  - Test NAT masquerade import
  - Test interface binding
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_acc_test.go acceptance test patterns_
  - _Requirements: 1, 5, 6_
  - _Prompt: Implement the task for spec rtx-nat-masquerade, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with NAT masquerade creation, update static entries, import existing descriptor. Test interface binding with pp1. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5, 6 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 14. Add example Terraform configurations
  - File: examples/nat_masquerade/main.tf (new)
  - Basic NAT masquerade creation example
  - NAT masquerade with static port mappings example
  - NAT masquerade with PPPoE interface example
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1, 6_
  - _Prompt: Implement the task for spec rtx-nat-masquerade, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: NAT masquerade with inner network and interface only. Full: NAT masquerade with static port mappings for web server (80->8080) and SSH (22). PPPoE: NAT masquerade using ipcp for outer address on pp1 interface | Restrictions: Use realistic IP addresses, include comments explaining options | _Leverage: examples/dhcp_scope/ | _Requirements: Requirements 1, 6 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-nat-masquerade, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works. Check terraform import functionality. Ensure no regressions in existing resources | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
