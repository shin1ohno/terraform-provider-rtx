# Tasks Document: rtx_interface

## Phase 1: Parser Layer

- [x] 1. Create InterfaceConfig data model and parser
  - File: internal/rtx/parsers/interface_config.go
  - Define InterfaceConfig and InterfaceIP structs
  - Implement ParseInterfaceConfig() to parse RTX output
  - Purpose: Parse "show config | grep <interface>" output for IP addresses, filters, NAT
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4, 5_
  - _Prompt: Implement the task for spec rtx-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create InterfaceConfig struct with Name, Description, IPAddress (InterfaceIP with Address/DHCP), SecureFilterIn, SecureFilterOut, DynamicFilterOut, NATDescriptor, ProxyARP, MTU fields. Implement ParseInterfaceConfig() function to parse RTX router output from "show config | grep <interface>" command. Handle various interface types (lan1, lan2, pp1, bridge1, tunnel1). Follow patterns from dhcp_scope.go | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line interface configurations | _Leverage: internal/rtx/parsers/dhcp_scope.go, internal/rtx/parsers/registry.go | _Requirements: Requirements 1-5 (IP Address, Filters, NAT, ProxyARP, Description) | Success: Parser correctly extracts all interface attributes from sample RTX output, handles edge cases like DHCP vs static IP, multiple filters | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [x] 2. Create command builder functions for interface configuration
  - File: internal/rtx/parsers/interface_config.go (continue)
  - Implement BuildIPAddressCommand() for IP address assignment
  - Implement BuildSecureFilterCommand() for security filter configuration
  - Implement BuildNATDescriptorCommand() for NAT binding
  - Implement BuildProxyARPCommand() for ProxyARP setting
  - Implement BuildDescriptionCommand() for interface description
  - Implement BuildMTUCommand() for MTU configuration
  - Implement BuildResetInterfaceCommands() for interface reset
  - Purpose: Generate RTX CLI commands for interface management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go BuildDHCPScopeCommand pattern_
  - _Requirements: 1, 2, 3, 4, 5_
  - _Prompt: Implement the task for spec rtx-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "ip <iface> address <ip>/<prefix>", "ip <iface> address dhcp", "ip <iface> secure filter in <filter_list>", "ip <iface> secure filter out <filter_list> dynamic <dynamic_list>", "ip <iface> nat descriptor <id>", "ip <iface> proxyarp on/off", "description <iface> <desc>", "ip <iface> mtu <size>". Also implement "no" commands for removal | Restrictions: Follow existing command builder pattern exactly, validate inputs before building commands, handle filter list ordering | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirements 1-5 | Success: All commands generate valid RTX CLI syntax, filter lists maintain order | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 3. Create parser unit tests
  - File: internal/rtx/parsers/interface_config_test.go
  - Test ParseInterfaceConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: DHCP vs static IP, missing fields, multiple filters
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3, 4, 5_
  - _Prompt: Implement the task for spec rtx-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for interface_config.go. Include test cases for parsing interface config output (static IP, DHCP, with filters, with NAT, with ProxyARP), command building with various parameter combinations, edge cases like empty filter list, interface types (lan1, bridge1, pp1) | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [x] 4. Add InterfaceConfig type to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add InterfaceConfig struct with all fields
  - Add InterfaceIP struct
  - Extend Client interface with interface configuration methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add InterfaceConfig struct (Name string, Description string, IPAddress *InterfaceIP, SecureFilterIn []int, SecureFilterOut []int, DynamicFilterOut []int, NATDescriptor int, ProxyARP bool, MTU int). Add InterfaceIP struct (Address string, DHCP bool). Extend Client interface with: GetInterfaceConfig(ctx, interfaceName) (*InterfaceConfig, error), ConfigureInterface(ctx, config) error, UpdateInterfaceConfig(ctx, config) error, ResetInterface(ctx, interfaceName) error, ListInterfaceConfigs(ctx) ([]InterfaceConfig, error) | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 5. Create InterfaceService implementation
  - File: internal/client/interface_service.go (new)
  - Implement InterfaceService struct with executor reference
  - Implement Configure() with validation and command execution
  - Implement Get() to parse interface configuration
  - Implement Update() for modifying interface settings
  - Implement Reset() to remove interface configuration
  - Implement List() to retrieve all configured interfaces
  - Purpose: Service layer for interface CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4, 5_
  - _Prompt: Implement the task for spec rtx-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create InterfaceService following DHCPScopeService pattern. Include input validation (interface name format, CIDR notation, valid filter numbers, valid NAT descriptor). Use parsers.BuildIPAddressCommand and related functions. Call client.SaveConfig() after modifications. Handle interface update by removing old configuration first then applying new | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, maintain separation from other services | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-5 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 6. Integrate InterfaceService into rtxClient
  - File: internal/client/client.go (modify)
  - Add interfaceService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface methods delegating to service
  - Purpose: Wire up interface service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add interfaceService *InterfaceService field to rtxClient. Initialize in Dial(): c.interfaceService = NewInterfaceService(c.executor, c). Implement GetInterfaceConfig, ConfigureInterface, UpdateInterfaceConfig, ResetInterface, ListInterfaceConfigs methods delegating to service | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all interface methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 7. Create service unit tests
  - File: internal/client/interface_service_test.go (new)
  - Test Configure with valid and invalid inputs
  - Test Get parsing
  - Test Update behavior
  - Test Reset
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4, 5_
  - _Prompt: Implement the task for spec rtx-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for InterfaceService. Mock Executor interface to simulate RTX responses. Test validation (invalid interface name, invalid CIDR, invalid filter numbers). Test successful CRUD operations. Test error handling for connection failures | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-5 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [x] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_interface.go (new)
  - Define resourceRTXInterface() with full schema
  - Add name (Required, ForceNew, String with validation)
  - Add description (Optional, String)
  - Add ip_address (Optional, Block with address/dhcp)
  - Add secure_filter_in (Optional, List of Int)
  - Add secure_filter_out (Optional, List of Int)
  - Add dynamic_filter_out (Optional, List of Int)
  - Add nat_descriptor (Optional, Int)
  - Add proxyarp (Optional, Bool, default false)
  - Add mtu (Optional, Int)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4, 5_
  - _Prompt: Implement the task for spec rtx-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXInterface() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Add ValidateFunc for name (lan1, lan2, pp1, bridge1, tunnel1 patterns), ip_address block (ExactlyOneOf address/dhcp). Set ForceNew on name. Use TypeList for secure_filter_in/out and dynamic_filter_out | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1-5 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_interface.go (continue)
  - Implement resourceRTXInterfaceCreate()
  - Implement resourceRTXInterfaceRead()
  - Implement resourceRTXInterfaceUpdate()
  - Implement resourceRTXInterfaceDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build InterfaceConfig from ResourceData, call client.ConfigureInterface, set ID to interface name). Read (call GetInterfaceConfig, update ResourceData, handle not found by clearing ID). Update (call UpdateInterfaceConfig for mutable fields). Delete (call ResetInterface to remove configuration). Follow rtx_dhcp_scope patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 10. Implement import functionality
  - File: internal/provider/resource_rtx_interface.go (continue)
  - Implement resourceRTXInterfaceImport()
  - Parse interface name from import ID string
  - Validate interface exists on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_scope.go import pattern_
  - _Requirements: 6_
  - _Prompt: Implement the task for spec rtx-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXInterfaceImport(). Parse import ID as interface name. Call GetInterfaceConfig to verify existence. Populate all ResourceData fields from retrieved configuration. Call Read to ensure state consistency | Restrictions: Handle invalid interface name format, non-existent interface errors gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go import function | _Requirements: Requirement 6 (Import) | Success: terraform import rtx_interface.example lan2 works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_interface" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_interface": resourceRTXInterface() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 12. Create resource unit tests
  - File: internal/provider/resource_rtx_interface_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 6_
  - _Prompt: Implement the task for spec rtx-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_interface.go. Test schema validation (invalid interface name, conflicting ip_address settings). Test CRUD operations with mocked client. Test import with valid and invalid interface names | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 6 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [x] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_interface_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test interface configuration with static IP
  - Test interface configuration with DHCP
  - Test security filter application
  - Test NAT descriptor binding
  - Test interface import
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_acc_test.go acceptance test patterns_
  - _Requirements: 1, 6, 7_
  - _Prompt: Implement the task for spec rtx-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with static IP, DHCP, security filters, NAT descriptor, ProxyARP. Test update scenarios (change filters, change NAT). Test import existing interface. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 6, 7 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 14. Add example Terraform configurations
  - File: examples/interface/main.tf (new)
  - WAN interface with DHCP example
  - LAN interface with static IP example
  - Interface with security filters and NAT example
  - Bridge interface example
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1, 7_
  - _Prompt: Implement the task for spec rtx-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. WAN: interface with DHCP and NAT descriptor. LAN: interface with static IP and ProxyARP. Full: interface with security filters (in/out/dynamic), NAT, description. Bridge: bridge interface with static IP. Show dependency with rtx_ip_filter and rtx_nat_masquerade | Restrictions: Use realistic IP addresses and filter numbers, include comments explaining options | _Leverage: examples/dhcp_scope/ | _Requirements: Requirements 1, 7 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-interface, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works. Check terraform import functionality. Ensure no regressions in existing resources | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
