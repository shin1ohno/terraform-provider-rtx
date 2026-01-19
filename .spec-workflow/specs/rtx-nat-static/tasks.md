# Tasks Document: rtx_nat_static

## Phase 1: Parser Layer

- [ ] 1. Create NATStatic data model and parser
  - File: internal/rtx/parsers/nat_static.go
  - Define NATStatic and NATStaticEntry structs
  - Implement ParseNATStaticConfig() to parse RTX output
  - Purpose: Parse "show config | grep nat descriptor" output for static NAT
  - _Leverage: internal/rtx/parsers/nat_masquerade.go for patterns (if exists), internal/rtx/parsers/dhcp_scope.go_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-nat-static, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create NATStatic struct with DescriptorID, StaticEntries, Interface fields. Create NATStaticEntry struct with InsideLocal, InsideLocalPort, OutsideGlobal, OutsideGlobalPort, Protocol fields. Implement ParseNATStaticConfig() function to parse RTX router output from "show config | grep nat descriptor" command for static type NAT. Follow patterns from existing parsers | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line configurations | _Leverage: internal/rtx/parsers/dhcp_scope.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (1:1 Mapping), Requirement 3 (Port-based NAT) | Success: Parser correctly extracts all static NAT attributes from sample RTX output, handles both 1:1 and port-based mappings | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 2. Create command builder functions for static NAT
  - File: internal/rtx/parsers/nat_static.go (continue)
  - Implement BuildNATDescriptorTypeStaticCommand() for descriptor type
  - Implement BuildNATStaticMappingCommand() for 1:1 mapping
  - Implement BuildNATStaticPortMappingCommand() for port-based mapping
  - Implement BuildDeleteNATStaticCommand() for deletion
  - Implement BuildShowNATStaticCommand() for reading
  - Purpose: Generate RTX CLI commands for static NAT management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go command builder patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-nat-static, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "nat descriptor type <id> static", "nat descriptor static <id> <outer_ip>=<inner_ip>", "nat descriptor static <id> <outer_ip>:<port>=<inner_ip>:<port> <protocol>", "no nat descriptor type <id>". Build commands for both 1:1 and port-based static NAT | Restrictions: Follow existing command builder patterns exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (1:1 Mapping), Requirement 3 (Port-based) | Success: All commands generate valid RTX CLI syntax for static NAT configuration | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 3. Create parser unit tests
  - File: internal/rtx/parsers/nat_static_test.go
  - Test ParseNATStaticConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: missing fields, port-based NAT, multiple entries
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-nat-static, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for nat_static.go. Include test cases for parsing static NAT config output, command building for 1:1 mapping, port-based mapping with tcp/udp protocols, edge cases like multiple entries per descriptor | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [ ] 4. Add NATStatic type to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add NATStatic struct with all fields
  - Add NATStaticEntry struct
  - Extend Client interface with static NAT methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing struct patterns in interfaces.go_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-nat-static, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add NATStatic struct (DescriptorID int, StaticEntries []NATStaticEntry, Interface string). Add NATStaticEntry struct (InsideLocal, OutsideGlobal string, InsideLocalPort, OutsideGlobalPort int, Protocol string). Extend Client interface with: GetNATStatic(ctx, descriptorID) (*NATStatic, error), CreateNATStatic(ctx, nat) error, UpdateNATStatic(ctx, nat) error, DeleteNATStatic(ctx, descriptorID) error, ListNATStatics(ctx) ([]NATStatic, error) | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 5. Create NATStaticService implementation
  - File: internal/client/nat_static_service.go (new)
  - Implement NATStaticService struct with executor reference
  - Implement Create() with validation and command execution
  - Implement Get() to parse static NAT configuration
  - Implement Update() for modifying entries
  - Implement Delete() for removing NAT descriptor
  - Implement List() to retrieve all static NAT descriptors
  - Purpose: Service layer for static NAT CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-nat-static, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create NATStaticService following existing service patterns. Include input validation (descriptor ID 1-65535, valid IP addresses, valid ports 1-65535, protocol tcp/udp required when ports specified). Use parsers.BuildNATDescriptorTypeStaticCommand and related functions. Call client.SaveConfig() after modifications. Handle update by deleting entries and recreating | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, maintain separation from other services | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-4 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 6. Integrate NATStaticService into rtxClient
  - File: internal/client/client.go (modify)
  - Add natStaticService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface static NAT methods delegating to service
  - Purpose: Wire up static NAT service to main client
  - _Leverage: existing service integration pattern in client.go_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-nat-static, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add natStaticService *NATStaticService field to rtxClient. Initialize in Dial(): c.natStaticService = NewNATStaticService(c.executor, c). Implement GetNATStatic, CreateNATStatic, UpdateNATStatic, DeleteNATStatic, ListNATStatics methods delegating to service | Restrictions: Follow existing service integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go existing service integrations | _Requirements: Requirement 1 | Success: Client compiles, all static NAT methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 7. Create service unit tests
  - File: internal/client/nat_static_service_test.go (new)
  - Test Create with valid and invalid inputs
  - Test Get parsing for 1:1 and port-based NAT
  - Test Update behavior
  - Test Delete
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-nat-static, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for NATStaticService. Mock Executor interface to simulate RTX responses. Test validation (invalid descriptor ID, invalid IPs, invalid ports, missing protocol for port NAT). Test successful CRUD operations for both 1:1 and port-based NAT. Test error handling | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [ ] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_nat_static.go (new)
  - Define resourceRTXNATStatic() with full schema
  - Add id (Required, ForceNew, Int) for descriptor ID
  - Add static_entries (Required, List of Object)
    - inside_local (Required, String with IP validation)
    - inside_local_port (Optional, Int, 1-65535)
    - outside_global (Required, String with IP validation)
    - outside_global_port (Optional, Int, 1-65535)
    - protocol (Optional, String, tcp/udp, required if ports specified)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-nat-static, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXNATStatic() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Add ValidateFunc for id (1-65535), IP addresses, ports. Set ForceNew on id. Use TypeList for static_entries with nested schema. Add custom validation: protocol required when ports are specified | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1-3 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_nat_static.go (continue)
  - Implement resourceRTXNATStaticCreate()
  - Implement resourceRTXNATStaticRead()
  - Implement resourceRTXNATStaticUpdate()
  - Implement resourceRTXNATStaticDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-nat-static, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build NATStatic from ResourceData, call client.CreateNATStatic, set ID to descriptor_id). Read (call GetNATStatic, update ResourceData, handle not found by clearing ID). Update (call UpdateNATStatic for static entries). Delete (call DeleteNATStatic). Follow rtx_dhcp_scope patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 10. Implement import functionality
  - File: internal/provider/resource_rtx_nat_static.go (continue)
  - Implement resourceRTXNATStaticImport()
  - Parse descriptor_id from import ID string
  - Validate descriptor exists on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_scope.go import pattern_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec rtx-nat-static, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXNATStaticImport(). Parse import ID as descriptor_id integer. Call GetNATStatic to verify existence. Populate all ResourceData fields from retrieved static NAT config. Call Read to ensure state consistency | Restrictions: Handle invalid import ID format, non-existent descriptor errors gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go import function | _Requirements: Requirement 5 (Import) | Success: terraform import rtx_nat_static.example 10 works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_nat_static" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-nat-static, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_nat_static": resourceRTXNATStatic() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 12. Create resource unit tests
  - File: internal/provider/resource_rtx_nat_static_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-nat-static, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_nat_static.go. Test schema validation (invalid descriptor ID, invalid IPs, invalid ports, missing protocol). Test CRUD operations with mocked client for both 1:1 and port-based NAT. Test import with valid and invalid IDs | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [ ] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_nat_static_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test 1:1 static NAT creation
  - Test port-based static NAT creation
  - Test static NAT update
  - Test static NAT import
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_acc_test.go acceptance test patterns_
  - _Requirements: 1, 5, 6_
  - _Prompt: Implement the task for spec rtx-nat-static, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with 1:1 static NAT creation, port-based static NAT with tcp/udp, update entries, import existing descriptor. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5, 6 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 14. Add example Terraform configurations
  - File: examples/nat_static/main.tf (new)
  - Basic 1:1 static NAT example
  - Port-based static NAT example
  - Multiple entries example
  - Purpose: User documentation and testing
  - _Leverage: examples/ existing examples_
  - _Requirements: 1, 6_
  - _Prompt: Implement the task for spec rtx-nat-static, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: 1:1 static NAT with single entry. Port-based: static NAT with tcp port forwarding. Multiple: static NAT with multiple entries including both 1:1 and port-based mappings | Restrictions: Use realistic IP addresses, include comments explaining options | _Leverage: examples/ | _Requirements: Requirements 1, 6 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-nat-static, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works for both 1:1 and port-based static NAT. Check terraform import functionality. Ensure no regressions in existing resources | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
