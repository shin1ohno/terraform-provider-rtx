# Tasks Document: rtx_ipv6_interface

## Phase 1: Parser Layer

- [ ] 1. Create IPv6InterfaceConfig data model and parser
  - File: internal/rtx/parsers/ipv6_interface.go
  - Define IPv6InterfaceConfig, IPv6Address, and RTADVConfig structs
  - Implement ParseIPv6InterfaceConfig() to parse RTX output
  - Purpose: Parse "show config | grep ipv6 <interface>" output
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4, 5_
  - _Prompt: Implement the task for spec rtx-ipv6-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create IPv6InterfaceConfig struct with Interface, Addresses, RTADV, DHCPv6Service, MTU, SecureFilterIn, SecureFilterOut, DynamicFilterOut fields. Create IPv6Address struct with Address, PrefixRef, InterfaceID. Create RTADVConfig struct with Enabled, PrefixID, OFlag, MFlag, Lifetime. Implement ParseIPv6InterfaceConfig() function to parse RTX router output from "show config | grep ipv6" command. Handle prefix-based addresses like "ra-prefix@lan2::2/64" | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line configurations | _Leverage: internal/rtx/parsers/dhcp_scope.go, internal/rtx/parsers/registry.go | _Requirements: All functional requirements | Success: Parser correctly extracts all IPv6 interface attributes from sample RTX output, handles edge cases like missing optional fields | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 2. Create command builder functions for IPv6 interface
  - File: internal/rtx/parsers/ipv6_interface.go (continue)
  - Implement BuildIPv6AddressCommand() for address configuration
  - Implement BuildIPv6RTADVCommand() for Router Advertisement
  - Implement BuildIPv6DHCPv6Command() for DHCPv6 service
  - Implement BuildIPv6MTUCommand() for MTU setting
  - Implement BuildIPv6SecureFilterCommand() for security filters
  - Implement BuildDeleteIPv6InterfaceCommands() for removal
  - Purpose: Generate RTX CLI commands for IPv6 interface management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go BuildDHCPScopeCommand pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ipv6-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "ipv6 <interface> address <address>", "ipv6 <interface> rtadv send <prefix_id> [o_flag=on|off] [m_flag=on|off]", "ipv6 <interface> dhcp service server|client", "ipv6 <interface> mtu <size>", "ipv6 <interface> secure filter in|out <filters>". Handle prefix-based address syntax like "ra-prefix@lan2::2/64" | Restrictions: Follow existing command builder pattern exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 3. Create parser unit tests
  - File: internal/rtx/parsers/ipv6_interface_test.go
  - Test ParseIPv6InterfaceConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: missing fields, prefix references, filter lists
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3, 4, 5_
  - _Prompt: Implement the task for spec rtx-ipv6-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for ipv6_interface.go. Include test cases for parsing IPv6 config output with static addresses, prefix-based addresses, RTADV configuration, DHCPv6 service, MTU, and security filters. Test command building with various parameter combinations | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [ ] 4. Add IPv6InterfaceConfig type to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add IPv6InterfaceConfig struct with all fields
  - Add IPv6Address struct
  - Add RTADVConfig struct
  - Extend Client interface with IPv6 interface methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ipv6-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add IPv6InterfaceConfig struct (Interface string, Addresses []IPv6Address, RTADV *RTADVConfig, DHCPv6Service string, MTU int, SecureFilterIn []int, SecureFilterOut []int, DynamicFilterOut []int). Add IPv6Address struct (Address, PrefixRef int, InterfaceID). Add RTADVConfig struct (Enabled bool, PrefixID int, OFlag bool, MFlag bool, Lifetime int). Extend Client interface with: GetIPv6InterfaceConfig(ctx, interfaceName) (*IPv6InterfaceConfig, error), ConfigureIPv6Interface(ctx, config) error, UpdateIPv6InterfaceConfig(ctx, config) error, ResetIPv6Interface(ctx, interfaceName) error, ListIPv6InterfaceConfigs(ctx) ([]IPv6InterfaceConfig, error) | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 5. Create IPv6InterfaceService implementation
  - File: internal/client/ipv6_interface_service.go (new)
  - Implement IPv6InterfaceService struct with executor reference
  - Implement Configure() with validation and command execution
  - Implement Get() to parse interface configuration
  - Implement Update() for modifying settings
  - Implement Reset() to remove IPv6 configuration
  - Implement List() to retrieve all IPv6 interfaces
  - Purpose: Service layer for IPv6 interface CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4, 5_
  - _Prompt: Implement the task for spec rtx-ipv6-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create IPv6InterfaceService following DHCPScopeService pattern. Include input validation (interface name format, IPv6 address format, filter numbers). Use parsers.BuildIPv6AddressCommand and related functions. Call client.SaveConfig() after modifications. Handle updates by clearing and reconfiguring as needed | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, maintain separation from other services | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-5 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 6. Integrate IPv6InterfaceService into rtxClient
  - File: internal/client/client.go (modify)
  - Add ipv6InterfaceService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface IPv6 methods delegating to service
  - Purpose: Wire up IPv6 interface service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ipv6-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add ipv6InterfaceService *IPv6InterfaceService field to rtxClient. Initialize in Dial(): c.ipv6InterfaceService = NewIPv6InterfaceService(c.executor, c). Implement GetIPv6InterfaceConfig, ConfigureIPv6Interface, UpdateIPv6InterfaceConfig, ResetIPv6Interface, ListIPv6InterfaceConfigs methods delegating to service | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all IPv6 interface methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 7. Create service unit tests
  - File: internal/client/ipv6_interface_service_test.go (new)
  - Test Configure with valid and invalid inputs
  - Test Get parsing for various configurations
  - Test Update behavior
  - Test Reset
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4, 5_
  - _Prompt: Implement the task for spec rtx-ipv6-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for IPv6InterfaceService. Mock Executor interface to simulate RTX responses. Test validation (invalid interface name, invalid IPv6 address, invalid filter numbers). Test successful CRUD operations. Test error handling | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-5 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [ ] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_ipv6_interface.go (new)
  - Define resourceRTXIPv6Interface() with full schema
  - Add interface (Required, ForceNew, String)
  - Add address block (Optional, List) with address/prefix_ref/interface_id
  - Add rtadv block (Optional, MaxItems 1) with enabled/prefix_id/o_flag/m_flag/lifetime
  - Add dhcpv6_service (Optional, String: "server", "client", "off")
  - Add mtu (Optional, Int)
  - Add secure_filter_in (Optional, List of Int)
  - Add secure_filter_out (Optional, List of Int)
  - Add dynamic_filter_out (Optional, List of Int)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4, 5_
  - _Prompt: Implement the task for spec rtx-ipv6-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXIPv6Interface() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Add ValidateFunc for interface name, IPv6 address format, dhcpv6_service values. Set ForceNew on interface. Use TypeList for address block with nested schema (address string, prefix_ref int, interface_id string). Use TypeList with MaxItems 1 for rtadv block | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1-5 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_ipv6_interface.go (continue)
  - Implement resourceRTXIPv6InterfaceCreate()
  - Implement resourceRTXIPv6InterfaceRead()
  - Implement resourceRTXIPv6InterfaceUpdate()
  - Implement resourceRTXIPv6InterfaceDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ipv6-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build IPv6InterfaceConfig from ResourceData, call client.ConfigureIPv6Interface, set ID to interface name). Read (call GetIPv6InterfaceConfig, update ResourceData, handle not found by clearing ID). Update (call UpdateIPv6InterfaceConfig for mutable fields). Delete (call ResetIPv6Interface). Follow rtx_dhcp_scope patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 10. Implement import functionality
  - File: internal/provider/resource_rtx_ipv6_interface.go (continue)
  - Implement resourceRTXIPv6InterfaceImport()
  - Parse interface name from import ID string
  - Validate interface has IPv6 configuration on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_scope.go import pattern_
  - _Requirements: 6_
  - _Prompt: Implement the task for spec rtx-ipv6-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXIPv6InterfaceImport(). Parse import ID as interface name. Call GetIPv6InterfaceConfig to verify existence. Populate all ResourceData fields from retrieved config. Call Read to ensure state consistency | Restrictions: Handle invalid import ID format, non-existent configuration errors gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go import function | _Requirements: Requirement 6 (Import) | Success: terraform import rtx_ipv6_interface.example lan1 works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_ipv6_interface" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ipv6-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_ipv6_interface": resourceRTXIPv6Interface() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 12. Create resource unit tests
  - File: internal/provider/resource_rtx_ipv6_interface_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 6_
  - _Prompt: Implement the task for spec rtx-ipv6-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_ipv6_interface.go. Test schema validation (invalid interface name, invalid IPv6 address, invalid dhcpv6_service value). Test CRUD operations with mocked client. Test import with valid and invalid interface names | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 6 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [ ] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_ipv6_interface_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test IPv6 interface creation with static address
  - Test IPv6 interface with prefix-based address
  - Test Router Advertisement configuration
  - Test DHCPv6 service configuration
  - Test security filter configuration
  - Test import
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_acc_test.go acceptance test patterns_
  - _Requirements: 1, 6, 7_
  - _Prompt: Implement the task for spec rtx-ipv6-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with IPv6 interface creation including address, RTADV, DHCPv6 service, MTU, and filters. Test updates. Test import. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 6, 7 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 14. Add example Terraform configurations
  - File: examples/ipv6_interface/main.tf (new)
  - WAN interface with DHCPv6 client example
  - LAN interface with RA and DHCPv6 server example
  - Bridge interface with prefix-based address example
  - Security filter configuration example
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1, 7_
  - _Prompt: Implement the task for spec rtx-ipv6-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. WAN: interface with dhcpv6_service client and security filters. LAN: interface with prefix-based address, rtadv block, and dhcpv6_service server. Bridge: interface with static address. Include comments explaining IPv6 concepts like SLAAC, RA flags, DHCPv6 modes | Restrictions: Use realistic IPv6 addresses and prefix references, include comments explaining options | _Leverage: examples/dhcp_scope/ | _Requirements: Requirements 1, 7 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-ipv6-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works. Check terraform import functionality. Ensure no regressions in existing resources | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
