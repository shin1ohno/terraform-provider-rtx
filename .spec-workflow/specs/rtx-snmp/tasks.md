# Tasks Document: rtx_snmp

## Phase 1: Parser Layer

- [ ] 1. Create SNMPConfig data model and parser
  - File: internal/rtx/parsers/snmp.go
  - Define SNMPConfig, SNMPCommunity, and SNMPHost structs
  - Implement ParseSNMPConfig() to parse RTX output
  - Purpose: Parse "show config | grep snmp" output
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-snmp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create SNMPConfig struct with Location, Contact, SysName, Communities, Hosts, EnableTraps fields. Create SNMPCommunity struct with Name, Permission (ro/rw), ACL fields. Create SNMPHost struct with Address, Community, Version fields. Implement ParseSNMPConfig() function to parse RTX router output from "show config | grep snmp" command. Follow patterns from dhcp_scope.go | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line SNMP configurations | _Leverage: internal/rtx/parsers/dhcp_scope.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (Community), Requirement 3 (Trap Host), Requirement 4 (Contact/Location) | Success: Parser correctly extracts all SNMP attributes from sample RTX output, handles edge cases like missing optional fields | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 2. Create command builder functions for SNMP
  - File: internal/rtx/parsers/snmp.go (continue)
  - Implement BuildSNMPSysLocationCommand() for location setting
  - Implement BuildSNMPSysContactCommand() for contact setting
  - Implement BuildSNMPSysNameCommand() for sysname setting
  - Implement BuildSNMPCommunityCommand() for community configuration
  - Implement BuildSNMPHostCommand() for trap destination
  - Implement BuildSNMPTrapEnableCommand() for trap types
  - Implement BuildDeleteSNMPCommand() for deletion
  - Purpose: Generate RTX CLI commands for SNMP management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go BuildDHCPScopeCommand pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-snmp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "snmp sysname <name>", "snmp syslocation <location>", "snmp syscontact <contact>", "snmp community read-only <string>", "snmp community read-write <string>", "snmp host <ip>", "snmp trap community <string>", "snmp trap enable snmp <types>", "no snmp host <ip>", "no snmp community read-only/read-write" | Restrictions: Follow existing BuildDHCPScopeCommand pattern exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 3. Create parser unit tests
  - File: internal/rtx/parsers/snmp_test.go
  - Test ParseSNMPConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: missing fields, malformed input
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-snmp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for snmp.go. Include test cases for parsing SNMP config output, command building with various parameter combinations, edge cases like empty community list, multiple trap hosts, enable traps list | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [ ] 4. Add SNMPConfig type to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add SNMPConfig struct with all fields
  - Add SNMPCommunity struct
  - Add SNMPHost struct
  - Extend Client interface with SNMP methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-snmp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add SNMPConfig struct (Location string, Contact string, SysName string, Communities []SNMPCommunity, Hosts []SNMPHost, EnableTraps []string). Add SNMPCommunity struct (Name, Permission, ACL string). Add SNMPHost struct (Address, Community, Version string). Extend Client interface with: GetSNMP(ctx) (*SNMPConfig, error), CreateSNMP(ctx, snmp) error, UpdateSNMP(ctx, snmp) error, DeleteSNMP(ctx) error | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 5. Create SNMPService implementation
  - File: internal/client/snmp_service.go (new)
  - Implement SNMPService struct with executor reference
  - Implement Create() with validation and command execution
  - Implement Get() to parse SNMP configuration
  - Implement Update() for modifying configuration
  - Implement Delete() to remove SNMP settings
  - Purpose: Service layer for SNMP CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-snmp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create SNMPService following DHCPScopeService pattern. Include input validation (IP format for hosts, valid SNMP versions 1/2c/3, community string not empty). Use parsers.BuildSNMPSysLocationCommand and related functions. Call client.SaveConfig() after modifications. SNMP is a singleton resource - handle update by deleting existing and recreating | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, mark community strings as sensitive | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-4 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 6. Integrate SNMPService into rtxClient
  - File: internal/client/client.go (modify)
  - Add snmpService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface SNMP methods delegating to service
  - Purpose: Wire up SNMP service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-snmp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add snmpService *SNMPService field to rtxClient. Initialize in Dial(): c.snmpService = NewSNMPService(c.executor, c). Implement GetSNMP, CreateSNMP, UpdateSNMP, DeleteSNMP methods delegating to service | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all SNMP methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 7. Create service unit tests
  - File: internal/client/snmp_service_test.go (new)
  - Test Create with valid and invalid inputs
  - Test Get parsing
  - Test Update behavior
  - Test Delete
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-snmp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for SNMPService. Mock Executor interface to simulate RTX responses. Test validation (invalid IP for hosts, invalid SNMP version, empty community string). Test successful CRUD operations. Test error handling | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [ ] 8. Create Terraform resource schema for rtx_snmp_server
  - File: internal/provider/resource_rtx_snmp_server.go (new)
  - Define resourceRTXSNMPServer() with full schema
  - Add location (Optional, String)
  - Add contact (Optional, String)
  - Add communities (Optional, List of Object with name/permission/acl, Sensitive)
  - Add hosts (Optional, List of Object with address/community/version)
  - Add enable_traps (Optional, List of String)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-snmp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXSNMPServer() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Use TypeList for communities with nested schema (name Sensitive, permission, acl). Use TypeList for hosts with nested schema (address with IP validation, community Sensitive, version with validation for 1/2c/3). Add ValidateFunc for trap types | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style, mark sensitive fields appropriately | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1-4 | Success: Schema compiles, validation functions work, sensitive attributes marked | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 9. Implement CRUD operations for rtx_snmp_server resource
  - File: internal/provider/resource_rtx_snmp_server.go (continue)
  - Implement resourceRTXSNMPServerCreate()
  - Implement resourceRTXSNMPServerRead()
  - Implement resourceRTXSNMPServerUpdate()
  - Implement resourceRTXSNMPServerDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-snmp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build SNMPConfig from ResourceData, call client.CreateSNMP, set ID to "snmp" as singleton). Read (call GetSNMP, update ResourceData, handle not found by clearing ID). Update (call UpdateSNMP for mutable fields). Delete (call DeleteSNMP). Follow rtx_dhcp_scope patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully, use fixed ID "snmp" for singleton resource | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 10. Implement import functionality for rtx_snmp_server
  - File: internal/provider/resource_rtx_snmp_server.go (continue)
  - Implement resourceRTXSNMPServerImport()
  - Accept "snmp" as import ID
  - Validate SNMP configuration exists on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_scope.go import pattern_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec rtx-snmp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXSNMPServerImport(). Accept "snmp" as import ID (singleton resource). Call GetSNMP to verify SNMP is configured. Populate all ResourceData fields from retrieved config. Call Read to ensure state consistency | Restrictions: Handle non-existent SNMP configuration gracefully, return clear error if not configured | _Leverage: internal/provider/resource_rtx_dhcp_scope.go import function | _Requirements: Requirement 5 (Import) | Success: terraform import rtx_snmp_server.example snmp works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_snmp_server" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-snmp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_snmp_server": resourceRTXSNMPServer() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 12. Create resource unit tests
  - File: internal/provider/resource_rtx_snmp_server_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-snmp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_snmp_server.go. Test schema validation (invalid IP for hosts, invalid SNMP version, empty community). Test CRUD operations with mocked client. Test import with valid and invalid IDs | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [ ] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_snmp_server_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test SNMP creation with communities and hosts
  - Test SNMP update (change location, add trap host)
  - Test SNMP import
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_test.go acceptance test patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-snmp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with SNMP creation including communities and trap hosts. Test update to modify location and contact. Test import existing SNMP configuration. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 14. Add example Terraform configurations
  - File: examples/snmp/main.tf (new)
  - Basic SNMP configuration example
  - SNMP with all options example (communities, trap hosts, traps enabled)
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-snmp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: SNMP with location and contact only. Full: SNMP with communities (ro and rw), trap hosts, enable_traps. Security note: Show using variables for sensitive community strings | Restrictions: Use realistic values, include comments explaining options, demonstrate security best practices with variables | _Leverage: examples/dhcp_scope/ | _Requirements: Requirement 1 | Success: Examples are valid Terraform, demonstrate all features, show security best practices | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-snmp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works. Check terraform import functionality. Ensure no regressions in existing resources | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
