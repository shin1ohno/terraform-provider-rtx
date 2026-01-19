# Tasks Document: rtx_ethernet_filter

## Phase 1: Parser Layer

- [ ] 1. Create EthernetFilter data model and parser
  - File: internal/rtx/parsers/ethernet_filter.go
  - Define EthernetFilter and EthernetFilterEntry structs
  - Implement ParseEthernetFilterConfig() to parse RTX output
  - Implement NormalizeMAC() for MAC address format normalization
  - Purpose: Parse "show config | grep ethernet filter" output
  - _Leverage: internal/rtx/parsers/dhcp_bindings.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ethernet-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create EthernetFilter struct with Number, Name, Entries fields. Create EthernetFilterEntry struct with Sequence, Action, SourceMAC, SourceMACMask, SourceAny, DestMAC, DestMACMask, DestAny, EtherType, VlanID fields. Implement ParseEthernetFilterConfig() to parse RTX router output from "show config | grep ethernet filter" command. Implement NormalizeMAC() to convert between MAC formats (00:11:22:33:44:55 vs 0011.2233.4455). Follow patterns from dhcp_bindings.go | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line filter configurations | _Leverage: internal/rtx/parsers/dhcp_bindings.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (MAC Filtering), Requirement 3 (EtherType), Requirement 4 (Action Mapping) | Success: Parser correctly extracts all filter attributes from sample RTX output, handles edge cases like wildcard MAC addresses | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 2. Create command builder functions for Ethernet filter
  - File: internal/rtx/parsers/ethernet_filter.go (continue)
  - Implement BuildEthernetFilterCommand() for filter creation
  - Implement BuildInterfaceEthernetFilterCommand() for interface binding
  - Implement BuildDeleteEthernetFilterCommand() for deletion
  - Implement BuildShowEthernetFilterCommand() for reading
  - Purpose: Generate RTX CLI commands for Ethernet filter management
  - _Leverage: internal/rtx/parsers/dhcp_bindings.go BuildDHCPBindCommand pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ethernet-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "ethernet filter <n> <action> <src_mac> <dst_mac> [<eth_type>]", "ethernet <interface> filter <direction> <filter_list>", "no ethernet filter <n>". Map Cisco permit/deny actions to RTX pass/reject. Handle wildcard MAC addresses with * | Restrictions: Follow existing BuildDHCPBindCommand pattern exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_bindings.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax, action mapping works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 3. Create parser unit tests
  - File: internal/rtx/parsers/ethernet_filter_test.go
  - Test ParseEthernetFilterConfig with various RTX output formats
  - Test all command builder functions
  - Test NormalizeMAC() with different MAC formats
  - Test edge cases: wildcard MACs, missing fields, malformed input
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_bindings_test.go for test patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ethernet-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for ethernet_filter.go. Include test cases for parsing filter config output, command building with various parameter combinations, MAC address normalization, edge cases like wildcard MACs (*), EtherType values (0x0800, 0x0806, 0x86DD) | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_bindings_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [ ] 4. Add EthernetFilter type to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add EthernetFilter struct with all fields
  - Add EthernetFilterEntry struct
  - Extend Client interface with Ethernet filter methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPBinding struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ethernet-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add EthernetFilter struct (Number int, Name string, Entries []EthernetFilterEntry). Add EthernetFilterEntry struct (Sequence int, Action string, SourceMAC string, SourceMACMask string, SourceAny bool, DestMAC string, DestMACMask string, DestAny bool, EtherType string, VlanID int). Extend Client interface with: GetEthernetFilter(ctx, filterNum) (*EthernetFilter, error), CreateEthernetFilter(ctx, filter) error, UpdateEthernetFilter(ctx, filter) error, DeleteEthernetFilter(ctx, filterNum) error, ListEthernetFilters(ctx) ([]EthernetFilter, error) | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 5. Create EthernetFilterService implementation
  - File: internal/client/ethernet_filter_service.go (new)
  - Implement EthernetFilterService struct with executor reference
  - Implement Create() with validation and command execution
  - Implement Get() to parse filter configuration
  - Implement Update() for modifying filter entries
  - Implement Delete() for removing filters
  - Implement List() to retrieve all filters
  - Purpose: Service layer for Ethernet filter CRUD operations
  - _Leverage: internal/client/dhcp_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ethernet-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create EthernetFilterService following DHCPService pattern. Include input validation (filter number 1-65535, MAC address format, EtherType hex format, VLAN ID 1-4094). Use parsers.BuildEthernetFilterCommand and related functions. Call client.SaveConfig() after modifications. Handle filter update by deleting and recreating if needed | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, maintain separation from other services | _Leverage: internal/client/dhcp_service.go | _Requirements: Requirements 1-4 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 6. Integrate EthernetFilterService into rtxClient
  - File: internal/client/client.go (modify)
  - Add ethernetFilterService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface filter methods delegating to service
  - Purpose: Wire up Ethernet filter service to main client
  - _Leverage: existing dhcpService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ethernet-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add ethernetFilterService *EthernetFilterService field to rtxClient. Initialize in Dial(): c.ethernetFilterService = NewEthernetFilterService(c.executor, c). Implement GetEthernetFilter, CreateEthernetFilter, UpdateEthernetFilter, DeleteEthernetFilter, ListEthernetFilters methods delegating to service | Restrictions: Follow existing dhcpService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpService integration | _Requirements: Requirement 1 | Success: Client compiles, all filter methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 7. Create service unit tests
  - File: internal/client/ethernet_filter_service_test.go (new)
  - Test Create with valid and invalid inputs
  - Test Get parsing
  - Test Update behavior
  - Test Delete
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ethernet-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for EthernetFilterService. Mock Executor interface to simulate RTX responses. Test validation (invalid filter number, invalid MAC format, invalid EtherType). Test successful CRUD operations. Test error handling | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [ ] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_access_list_mac.go (new)
  - Define resourceRTXAccessListMac() with full schema
  - Add name (Required, String)
  - Add entries (Required, List of Object)
    - sequence (Required, Int)
    - ace_action (Required, String: permit/deny)
    - source_address (Optional, String with MAC validation)
    - source_address_mask (Optional, String)
    - source_any (Optional, Bool)
    - destination_address (Optional, String with MAC validation)
    - destination_address_mask (Optional, String)
    - destination_any (Optional, Bool)
    - ethertype (Optional, String with hex validation)
    - vlan_id (Optional, Int, range 1-4094)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_binding.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ethernet-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXAccessListMac() returning *schema.Resource. Define schema following rtx_dhcp_binding patterns. Add ValidateFunc for MAC addresses, EtherType (hex format), VLAN ID (1-4094). Use TypeList for entries with nested schema. Ensure ace_action validates permit/deny values | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_binding.go | _Requirements: Requirements 1-4 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_access_list_mac.go (continue)
  - Implement resourceRTXAccessListMacCreate()
  - Implement resourceRTXAccessListMacRead()
  - Implement resourceRTXAccessListMacUpdate()
  - Implement resourceRTXAccessListMacDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_binding.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ethernet-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build EthernetFilter from ResourceData, call client.CreateEthernetFilter, set ID to filter number). Read (call GetEthernetFilter, update ResourceData, handle not found by clearing ID). Update (call UpdateEthernetFilter for entry changes). Delete (call DeleteEthernetFilter). Follow rtx_dhcp_binding patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully | _Leverage: internal/provider/resource_rtx_dhcp_binding.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 10. Implement import functionality
  - File: internal/provider/resource_rtx_access_list_mac.go (continue)
  - Implement resourceRTXAccessListMacImport()
  - Parse filter number from import ID string
  - Validate filter exists on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_binding.go import pattern_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec rtx-ethernet-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXAccessListMacImport(). Parse import ID as filter number integer. Call GetEthernetFilter to verify existence. Populate all ResourceData fields from retrieved filter. Call Read to ensure state consistency | Restrictions: Handle invalid import ID format, non-existent filter errors gracefully | _Leverage: internal/provider/resource_rtx_dhcp_binding.go import function | _Requirements: Requirement 5 (Import) | Success: terraform import rtx_access_list_mac.example 1 works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 11. Create interface MAC ACL binding resource
  - File: internal/provider/resource_rtx_interface_mac_acl.go (new)
  - Define resourceRTXInterfaceMacACL() with schema
    - interface (Required, String)
    - mac_access_group_in (Optional, String)
    - mac_access_group_out (Optional, String)
  - Implement CRUD operations for interface binding
  - Purpose: Bind MAC ACL to interface
  - _Leverage: resource_rtx_dhcp_binding.go patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ethernet-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXInterfaceMacACL() for binding MAC access lists to interfaces. Schema includes interface name and optional in/out access group names. Implement CRUD to execute "ethernet <interface> filter <direction> <filter_list>" commands. Handle update by removing old binding before adding new | Restrictions: Follow Terraform SDK v2 patterns, coordinate with rtx_access_list_mac resource | _Leverage: internal/provider/resource_rtx_dhcp_binding.go | _Requirements: Requirement 1 | Success: Interface binding works end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 12. Register resources in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_access_list_mac" to ResourcesMap
  - Add "rtx_interface_mac_acl" to ResourcesMap
  - Purpose: Make resources available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ethernet-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entries to ResourcesMap in provider.go: "rtx_access_list_mac": resourceRTXAccessListMac(), "rtx_interface_mac_acl": resourceRTXInterfaceMacACL() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resources registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 13. Create resource unit tests
  - File: internal/provider/resource_rtx_access_list_mac_test.go (new)
  - File: internal/provider/resource_rtx_interface_mac_acl_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_binding_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-ethernet-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_access_list_mac.go and resource_rtx_interface_mac_acl.go. Test schema validation (invalid MAC, invalid EtherType, invalid VLAN). Test CRUD operations with mocked client. Test import with valid and invalid IDs | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_binding_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [ ] 14. Create acceptance tests
  - File: internal/provider/resource_rtx_access_list_mac_acc_test.go (new)
  - File: internal/provider/resource_rtx_interface_mac_acl_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test MAC ACL creation with all parameters
  - Test MAC ACL update
  - Test MAC ACL import
  - Test interface binding
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_binding_test.go acceptance test patterns_
  - _Requirements: 1, 5, 6_
  - _Prompt: Implement the task for spec rtx-ethernet-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with MAC ACL creation, update entries, import existing filter. Test rtx_interface_mac_acl depends_on rtx_access_list_mac. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5, 6 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 15. Add example Terraform configurations
  - File: examples/ethernet_filter/main.tf (new)
  - Basic MAC ACL creation example
  - MAC ACL with EtherType filtering example
  - Interface binding example
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp/ existing examples_
  - _Requirements: 1, 6_
  - _Prompt: Implement the task for spec rtx-ethernet-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: MAC ACL with permit/deny rules. Full: MAC ACL with EtherType filtering (0x0800 IPv4, 0x0806 ARP, 0x86DD IPv6). Integration: rtx_access_list_mac with rtx_interface_mac_acl showing interface binding | Restrictions: Use realistic MAC addresses, include comments explaining options | _Leverage: examples/dhcp/ | _Requirements: Requirements 1, 6 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 16. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-ethernet-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works. Check terraform import functionality. Ensure no regressions in existing resources | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
