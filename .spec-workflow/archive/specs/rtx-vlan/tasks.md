# Tasks Document: rtx_vlan

## Phase 1: Parser Layer

- [x] 1. Create VLAN data model and parser
  - File: internal/rtx/parsers/vlan.go
  - Define VLAN struct with VlanID, Name, Interface, VlanInterface, IPAddress, IPMask, Shutdown fields
  - Implement ParseVLANConfig() to parse RTX output
  - Purpose: Parse "show config | grep vlan" output for VLAN configuration
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-vlan, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create VLAN struct with VlanID (int), Name (string), Interface (string, parent like lan1), VlanInterface (string, computed like lan1/1), IPAddress (string), IPMask (string), Shutdown (bool) fields. Implement ParseVLANConfig() function to parse RTX router output from "show config | grep vlan" command. Handle "vlan lan1/1 802.1q vid=10" format and associated IP configuration | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line VLAN configurations | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (Interface/VLAN ID), Requirement 3 (IP Configuration) | Success: Parser correctly extracts all VLAN attributes from sample RTX output, handles edge cases like missing optional fields | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [x] 2. Create command builder functions for VLAN
  - File: internal/rtx/parsers/vlan.go (continue)
  - Implement BuildVLANCommand() for VLAN interface creation with 802.1q vid
  - Implement BuildVLANIPCommand() for IP address configuration
  - Implement BuildVLANDescriptionCommand() for name/description
  - Implement BuildDeleteVLANCommand() for deletion
  - Implement BuildShowVLANCommand() for reading configuration
  - Purpose: Generate RTX CLI commands for VLAN management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go BuildDHCPScopeCommand pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-vlan, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "vlan <interface>/<n> 802.1q vid=<vlan_id>", "ip <vlan_interface> address <ip>/<prefix>", "description <vlan_interface> <name>", "<vlan_interface> enable / no <vlan_interface> enable", "no vlan <interface>/<n>". Convert IP mask to prefix length format | Restrictions: Follow existing BuildDHCPScopeCommand pattern exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax, IP mask conversion works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 3. Create parser unit tests
  - File: internal/rtx/parsers/vlan_test.go
  - Test ParseVLANConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: missing fields, malformed input, VLAN ID validation (1-4094)
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-vlan, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for vlan.go. Include test cases for parsing VLAN config output, command building with various parameter combinations, edge cases like VLAN with no IP, shutdown state variations, invalid VLAN ID (out of 1-4094 range) | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [x] 4. Add VLAN type to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add VLAN struct with all fields
  - Extend Client interface with VLAN methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-vlan, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add VLAN struct (VlanID int, Name string, Interface string, VlanInterface string, IPAddress string, IPMask string, Shutdown bool). Extend Client interface with: GetVLAN(ctx, iface string, vlanID int) (*VLAN, error), CreateVLAN(ctx, vlan VLAN) error, UpdateVLAN(ctx, vlan VLAN) error, DeleteVLAN(ctx, iface string, vlanID int) error, ListVLANs(ctx) ([]VLAN, error) | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 5. Create VLANService implementation
  - File: internal/client/vlan_service.go (new)
  - Implement VLANService struct with executor reference
  - Implement Create() with validation and command execution
  - Implement Get() to parse VLAN configuration
  - Implement Update() for modifying name, IP, shutdown state
  - Implement Delete() to remove VLAN interface
  - Implement List() to retrieve all VLANs
  - Purpose: Service layer for VLAN CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-vlan, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create VLANService following DHCPScopeService pattern. Include input validation (VLAN ID 1-4094, valid interface names like lan1/lan2, IP address format). Use parsers.BuildVLANCommand and related functions. Call client.SaveConfig() after modifications. Handle interface slot auto-assignment or explicit specification | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, maintain separation from other services | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-4 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 6. Integrate VLANService into rtxClient
  - File: internal/client/client.go (modify)
  - Add vlanService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface VLAN methods delegating to service
  - Purpose: Wire up VLAN service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-vlan, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add vlanService *VLANService field to rtxClient. Initialize in Dial(): c.vlanService = NewVLANService(c.executor, c). Implement GetVLAN, CreateVLAN, UpdateVLAN, DeleteVLAN, ListVLANs methods delegating to service | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all VLAN methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 7. Create service unit tests
  - File: internal/client/vlan_service_test.go (new)
  - Test Create with valid and invalid inputs
  - Test Get parsing
  - Test Update behavior
  - Test Delete
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-vlan, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for VLANService. Mock Executor interface to simulate RTX responses. Test validation (invalid VLAN ID, invalid interface name, invalid IP). Test successful CRUD operations. Test error handling for VLAN already exists, VLAN not found scenarios | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [x] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_vlan.go (new)
  - Define resourceRTXVLAN() with full schema
  - Add vlan_id (Required, ForceNew, Int, ValidateFunc 1-4094)
  - Add interface (Required, ForceNew, String for parent interface like lan1)
  - Add name (Optional, String for description)
  - Add ip_address (Optional, String with IP validation)
  - Add ip_mask (Optional, String, required if ip_address set)
  - Add shutdown (Optional, Bool, default false)
  - Add vlan_interface (Computed, String, derived like lan1/1)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-vlan, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXVLAN() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Add ValidateFunc for vlan_id (1-4094), ip_address (valid IP). Set ForceNew on vlan_id and interface. Set vlan_interface as Computed attribute. Add custom validation: ip_mask required when ip_address is set | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1-3 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_vlan.go (continue)
  - Implement resourceRTXVLANCreate()
  - Implement resourceRTXVLANRead()
  - Implement resourceRTXVLANUpdate()
  - Implement resourceRTXVLANDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-vlan, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build VLAN from ResourceData, call client.CreateVLAN, set ID to interface/vlan_id format like "lan1/10"). Read (call GetVLAN, update ResourceData including computed vlan_interface, handle not found by clearing ID). Update (call UpdateVLAN for name, ip_address, ip_mask, shutdown). Delete (call DeleteVLAN). Follow rtx_dhcp_scope patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 10. Implement import functionality
  - File: internal/provider/resource_rtx_vlan.go (continue)
  - Implement resourceRTXVLANImport()
  - Parse interface and vlan_id from import ID string (format: "lan1/10")
  - Validate VLAN exists on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_scope.go import pattern_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec rtx-vlan, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXVLANImport(). Parse import ID as "interface/vlan_id" format (e.g., "lan1/10"). Call GetVLAN to verify existence. Populate all ResourceData fields from retrieved VLAN. Call Read to ensure state consistency | Restrictions: Handle invalid import ID format, non-existent VLAN errors gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go import function | _Requirements: Requirement 5 (Import) | Success: terraform import rtx_vlan.example lan1/10 works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_vlan" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-vlan, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_vlan": resourceRTXVLAN() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 12. Create resource unit tests
  - File: internal/provider/resource_rtx_vlan_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-vlan, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_vlan.go. Test schema validation (invalid VLAN ID, missing ip_mask when ip_address set). Test CRUD operations with mocked client. Test import with valid and invalid IDs (lan1/10 format) | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [x] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_vlan_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test VLAN creation with all parameters
  - Test VLAN update (name, IP, shutdown)
  - Test VLAN import
  - Test multiple VLANs on same interface
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_acc_test.go acceptance test patterns_
  - _Requirements: 1, 5, 6_
  - _Prompt: Implement the task for spec rtx-vlan, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with VLAN creation on lan1, update name and IP configuration, toggle shutdown state, import existing VLAN. Test creating multiple VLANs (vid=10, vid=20) on same parent interface. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5, 6 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 14. Add example Terraform configurations
  - File: examples/vlan/main.tf (new)
  - Basic VLAN creation example
  - VLAN with IP configuration example
  - Multiple VLANs on same interface example
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1, 6_
  - _Prompt: Implement the task for spec rtx-vlan, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: VLAN with ID only. Full: VLAN with name, IP address, mask, shutdown=false. Multi-VLAN: management VLAN (vid=10) and users VLAN (vid=20) on same lan1 interface showing 802.1Q segmentation | Restrictions: Use realistic IP addresses (192.168.x.x), include comments explaining RTX VLAN concepts | _Leverage: examples/dhcp_scope/ | _Requirements: Requirements 1, 6 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-vlan, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works for VLAN resources. Check terraform import functionality with interface/vlan_id format. Ensure no regressions in existing resources (dhcp_scope, dhcp_binding) | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
