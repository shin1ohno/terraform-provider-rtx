# Tasks Document: rtx_ip_filter

## Phase 1: Parser Layer

- [ ] 1. Create IPFilter data model and parser
  - File: internal/rtx/parsers/ip_filter.go
  - Define IPFilter and IPFilterEntry structs
  - Define IPFilterDynamic struct for stateful filters
  - Define IPv6Filter and IPv6FilterDynamic structs
  - Implement ParseIPFilterConfig() to parse RTX output
  - Purpose: Parse "show config | grep ip filter" output
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create IPFilter struct with Number, Name, Entries fields. Create IPFilterEntry struct with Sequence, Remark, Action (permit/deny), Protocol (tcp/udp/icmp/ip), SourcePrefix, SourcePrefixMask, SourceAny, SourcePort, DestPrefix, DestPrefixMask, DestAny, DestPort, Log, Established fields. Create IPFilterDynamic struct with Number, Source, Dest, Protocol, SyslogOn fields. Implement ParseIPFilterConfig() to parse RTX router output from "show config | grep ip filter" command. Follow patterns from dhcp_scope.go | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line filter configurations | _Leverage: internal/rtx/parsers/dhcp_scope.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (Filter Rules), Requirement 3 (Protocol), Requirement 4 (Ports) | Success: Parser correctly extracts all filter attributes from sample RTX output, handles edge cases like missing optional fields | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 2. Create command builder functions for IP filter
  - File: internal/rtx/parsers/ip_filter.go (continue)
  - Implement BuildIPFilterCommand() for static filter creation
  - Implement BuildIPFilterDynamicCommand() for dynamic filter creation
  - Implement BuildIPFilterSetCommand() for filter set creation
  - Implement BuildInterfaceSecureFilterCommand() for interface binding
  - Implement BuildDeleteIPFilterCommand() for deletion
  - Purpose: Generate RTX CLI commands for filter management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go BuildDHCPScopeCommand pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "ip filter <n> <action> <src> <dst> <protocol> [<src_port>] [<dst_port>]", "ip filter dynamic <n> <src> <dst> <protocol> [syslog=on|off]", "ip filter set <set_n> <filter_list>", "ip <interface> secure filter <direction> <set>", "no ip filter <n>". Map Cisco permit/deny to RTX pass/reject | Restrictions: Follow existing BuildDHCPScopeCommand pattern exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax, action mapping works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 3. Create IPv6 filter command builder functions
  - File: internal/rtx/parsers/ip_filter.go (continue)
  - Implement BuildIPv6FilterCommand() for static IPv6 filter creation
  - Implement BuildIPv6FilterDynamicCommand() for dynamic IPv6 filter creation
  - Implement BuildIPv6InterfaceSecureFilterCommand() for IPv6 interface binding
  - Implement BuildDeleteIPv6FilterCommand() for IPv6 filter deletion
  - Purpose: Generate RTX CLI commands for IPv6 filter management
  - _Leverage: IPv4 command builder functions_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create IPv6 command builder functions following RTX command syntax: "ipv6 filter <n> <action> <src> <dst> <protocol> [<src_port>] [<dst_port>]", "ipv6 filter dynamic <n> <src> <dst> <protocol> [syslog=on|off]", "ipv6 <interface> secure filter <direction> <filter_list>", "no ipv6 filter <n>". Support icmp6 protocol for IPv6 | Restrictions: Follow IPv4 command builder patterns, validate IPv6 addresses | _Leverage: IPv4 command builders in same file | _Requirements: Requirement 1 (CRUD Operations) | Success: All IPv6 commands generate valid RTX CLI syntax | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 4. Create parser unit tests
  - File: internal/rtx/parsers/ip_filter_test.go
  - Test ParseIPFilterConfig with various RTX output formats
  - Test all command builder functions for IPv4 and IPv6
  - Test edge cases: missing fields, malformed input, wildcard addresses
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for ip_filter.go. Include test cases for parsing filter config output, command building with various protocol/port combinations, edge cases like wildcard addresses (*), action mapping (permit->pass, deny->reject), dynamic filter parsing | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [ ] 5. Add IPFilter types to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add IPFilter struct with all fields
  - Add IPFilterEntry struct
  - Add IPFilterDynamic struct
  - Add IPv6Filter and IPv6FilterDynamic structs
  - Extend Client interface with IP filter methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add IPFilter struct (Number int, Name string, Entries []IPFilterEntry). Add IPFilterEntry struct (Sequence int, Remark string, Action string, Protocol string, SourcePrefix string, SourcePrefixMask string, SourceAny bool, SourcePort string, DestPrefix string, DestPrefixMask string, DestAny bool, DestPort string, Log bool, Established bool). Add IPFilterDynamic struct (Number int, Source string, Dest string, Protocol string, SyslogOn bool). Extend Client interface with: GetIPFilter(ctx, filterNum) (*IPFilter, error), CreateIPFilter(ctx, filter) error, UpdateIPFilter(ctx, filter) error, DeleteIPFilter(ctx, filterNum) error, ListIPFilters(ctx) ([]IPFilter, error) | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 6. Create IPFilterService implementation
  - File: internal/client/ip_filter_service.go (new)
  - Implement IPFilterService struct with executor reference
  - Implement Create() with validation and command execution
  - Implement Get() to parse filter configuration
  - Implement Update() for modifying filter entries
  - Implement Delete() with dependency check warning
  - Implement List() to retrieve all filters
  - Purpose: Service layer for IP filter CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4, 5_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create IPFilterService following DHCPScopeService pattern. Include input validation (filter number 1-65535, valid IP addresses, valid protocols tcp/udp/icmp/ip, valid ports 1-65535). Use parsers.BuildIPFilterCommand and related functions. Call client.SaveConfig() after modifications. Handle filter update by deleting and recreating if entries change | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, maintain separation from other services | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-5 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 7. Create IPFilterDynamicService implementation
  - File: internal/client/ip_filter_service.go (continue)
  - Implement methods for dynamic (stateful) filter management
  - Implement CreateDynamicFilter() with protocol validation
  - Implement GetDynamicFilter() to parse dynamic filter configuration
  - Implement DeleteDynamicFilter() for removal
  - Purpose: Service layer for dynamic filter CRUD operations
  - _Leverage: IPFilterService pattern in same file_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Extend IPFilterService with dynamic filter methods. Validate protocol against supported list (ftp, domain, www, smtp, pop3, submission, tcp, udp). Use parsers.BuildIPFilterDynamicCommand. Handle syslog option | Restrictions: Follow existing service patterns, reuse validation utilities | _Leverage: IPFilterService implementation | _Requirements: Requirement 1 | Success: Dynamic filter CRUD operations work correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 8. Integrate IPFilterService into rtxClient
  - File: internal/client/client.go (modify)
  - Add ipFilterService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface IP filter methods delegating to service
  - Purpose: Wire up filter service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add ipFilterService *IPFilterService field to rtxClient. Initialize in Dial(): c.ipFilterService = NewIPFilterService(c.executor, c). Implement GetIPFilter, CreateIPFilter, UpdateIPFilter, DeleteIPFilter, ListIPFilters methods delegating to service | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all IP filter methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 9. Create service unit tests
  - File: internal/client/ip_filter_service_test.go (new)
  - Test Create with valid and invalid inputs
  - Test Get parsing for various filter configurations
  - Test Update behavior
  - Test Delete
  - Test dynamic filter operations
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for IPFilterService. Mock Executor interface to simulate RTX responses. Test validation (invalid filter number, invalid IP, invalid protocol, invalid port). Test successful CRUD operations. Test error handling. Test dynamic filter operations | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [ ] 10. Create AccessListExtended Terraform resource schema
  - File: internal/provider/resource_rtx_access_list_extended.go (new)
  - Define resourceRTXAccessListExtended() with full schema
  - Add name (Required, ForceNew, String)
  - Add entries (Required, List of Object)
  - Entry schema: sequence, remark, ace_rule_action, ace_rule_protocol, source fields, destination fields, log, established
  - Purpose: Define Terraform resource structure for IPv4 access list
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXAccessListExtended() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Add entries as TypeList with nested schema: sequence (int), remark (string, optional), ace_rule_action (string, permit/deny), ace_rule_protocol (string), source_prefix (string, optional), source_prefix_mask (string, optional), source_any (bool, optional), source_port (string, optional), destination_prefix (string, optional), destination_prefix_mask (string, optional), destination_any (bool, optional), destination_port_equal (string, optional), log (bool, optional), established (bool, optional) | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1-4 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 11. Implement CRUD operations for AccessListExtended resource
  - File: internal/provider/resource_rtx_access_list_extended.go (continue)
  - Implement resourceRTXAccessListExtendedCreate()
  - Implement resourceRTXAccessListExtendedRead()
  - Implement resourceRTXAccessListExtendedUpdate()
  - Implement resourceRTXAccessListExtendedDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build IPFilter from ResourceData, call client.CreateIPFilter, set ID to filter name). Read (call GetIPFilter, update ResourceData, handle not found by clearing ID). Update (call UpdateIPFilter for entry changes). Delete (call DeleteIPFilter). Map Terraform permit/deny to RTX pass/reject. Follow rtx_dhcp_scope patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 12. Create IPFilterDynamic Terraform resource
  - File: internal/provider/resource_rtx_ip_filter_dynamic.go (new)
  - Define resourceRTXIPFilterDynamic() with full schema
  - Add number (Required, ForceNew, Int)
  - Add source (Required, String)
  - Add destination (Required, String)
  - Add protocol (Required, String)
  - Add syslog (Optional, Bool, default false)
  - Implement CRUD operations
  - Purpose: Terraform resource for dynamic (stateful) IPv4 filters
  - _Leverage: resource_rtx_access_list_extended.go patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXIPFilterDynamic() returning *schema.Resource. Define schema with number (int, ForceNew), source (string), destination (string), protocol (string, validate against ftp/domain/www/smtp/pop3/submission/tcp/udp), syslog (bool, default false). Implement CRUD operations following existing patterns | Restrictions: Follow Terraform SDK v2 patterns, validate protocol values | _Leverage: internal/provider/resource_rtx_access_list_extended.go | _Requirements: Requirement 1 | Success: Schema compiles, CRUD operations work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 13. Create IPv6 filter Terraform resources
  - File: internal/provider/resource_rtx_access_list_extended_ipv6.go (new)
  - File: internal/provider/resource_rtx_ipv6_filter_dynamic.go (new)
  - Define resourceRTXAccessListExtendedIPv6() with schema similar to IPv4
  - Define resourceRTXIPv6FilterDynamic() with schema similar to IPv4
  - Implement CRUD operations for both resources
  - Purpose: Terraform resources for IPv6 filters
  - _Leverage: IPv4 resource patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXAccessListExtendedIPv6() and resourceRTXIPv6FilterDynamic() following IPv4 patterns. Support icmp6 protocol for IPv6 access lists. Implement CRUD operations for both resources | Restrictions: Follow IPv4 resource patterns exactly, validate IPv6 addresses | _Leverage: IPv4 resource implementations | _Requirements: Requirement 1 | Success: Both resources compile, CRUD operations work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 14. Create InterfaceACL Terraform resource
  - File: internal/provider/resource_rtx_interface_acl.go (new)
  - Define resourceRTXInterfaceACL() with full schema
  - Add interface (Required, ForceNew, String)
  - Add ip_access_group_in (Optional, String)
  - Add ip_access_group_out (Optional, String)
  - Add ipv6_access_group_in (Optional, String)
  - Add ipv6_access_group_out (Optional, String)
  - Implement CRUD operations
  - Purpose: Terraform resource for applying ACLs to interfaces
  - _Leverage: other resource patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXInterfaceACL() returning *schema.Resource. Define schema with interface (string, ForceNew), ip_access_group_in (string, optional), ip_access_group_out (string, optional), ipv6_access_group_in (string, optional), ipv6_access_group_out (string, optional). Generate RTX commands: "ip <interface> secure filter in <acl>" for applying ACLs. Implement CRUD operations | Restrictions: Follow Terraform SDK v2 patterns, at least one access_group must be specified | _Leverage: other resource implementations | _Requirements: Requirement 1 | Success: Schema compiles, CRUD operations work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 15. Implement import functionality for all resources
  - File: internal/provider/resource_rtx_access_list_extended.go (continue)
  - File: internal/provider/resource_rtx_ip_filter_dynamic.go (continue)
  - File: internal/provider/resource_rtx_access_list_extended_ipv6.go (continue)
  - File: internal/provider/resource_rtx_ipv6_filter_dynamic.go (continue)
  - File: internal/provider/resource_rtx_interface_acl.go (continue)
  - Implement import functions for each resource
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_scope.go import pattern_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement import functions for all IP filter resources. Parse import ID as filter name/number or interface. Call Get methods to verify existence. Populate all ResourceData fields from retrieved configuration. Call Read to ensure state consistency | Restrictions: Handle invalid import ID format, non-existent resource errors gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go import function | _Requirements: Requirement 5 (Import) | Success: terraform import works correctly for all resources | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 16. Register resources in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_access_list_extended" to ResourcesMap
  - Add "rtx_access_list_extended_ipv6" to ResourcesMap
  - Add "rtx_ip_filter_dynamic" to ResourcesMap
  - Add "rtx_ipv6_filter_dynamic" to ResourcesMap
  - Add "rtx_interface_acl" to ResourcesMap
  - Purpose: Make resources available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entries to ResourcesMap in provider.go: "rtx_access_list_extended": resourceRTXAccessListExtended(), "rtx_access_list_extended_ipv6": resourceRTXAccessListExtendedIPv6(), "rtx_ip_filter_dynamic": resourceRTXIPFilterDynamic(), "rtx_ipv6_filter_dynamic": resourceRTXIPv6FilterDynamic(), "rtx_interface_acl": resourceRTXInterfaceACL() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with all new resources registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 17. Create resource unit tests
  - File: internal/provider/resource_rtx_access_list_extended_test.go (new)
  - File: internal/provider/resource_rtx_ip_filter_dynamic_test.go (new)
  - File: internal/provider/resource_rtx_access_list_extended_ipv6_test.go (new)
  - File: internal/provider/resource_rtx_ipv6_filter_dynamic_test.go (new)
  - File: internal/provider/resource_rtx_interface_acl_test.go (new)
  - Test schema validation for each resource
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for all IP filter resources. Test schema validation (invalid protocol, invalid port, invalid IP). Test CRUD operations with mocked client. Test import with valid and invalid IDs | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [ ] 18. Create acceptance tests
  - File: internal/provider/resource_rtx_access_list_extended_acc_test.go (new)
  - File: internal/provider/resource_rtx_ip_filter_dynamic_acc_test.go (new)
  - File: internal/provider/resource_rtx_interface_acl_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test filter creation with various protocols and ports
  - Test filter update
  - Test filter import
  - Test interface ACL binding
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_acc_test.go acceptance test patterns_
  - _Requirements: 1, 5, 6_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with access list creation, update entries, import existing filter. Test rtx_interface_acl depends_on rtx_access_list_extended. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5, 6 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 19. Add example Terraform configurations
  - File: examples/ip_filter/main.tf (new)
  - Basic access list example
  - Access list with multiple entries example
  - Dynamic filter example
  - Interface ACL binding example
  - IPv6 filter examples
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1, 6_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: access list with single permit rule. Full: access list with multiple entries (permit HTTP, permit HTTPS, deny all). Dynamic: stateful filter for common protocols (ftp, www, tcp). Interface: rtx_interface_acl with access list dependency. IPv6: equivalent examples for IPv6 filters | Restrictions: Use realistic IP addresses, include comments explaining options | _Leverage: examples/dhcp_scope/ | _Requirements: Requirements 1, 6 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 20. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-ip-filter, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works. Check terraform import functionality. Ensure no regressions in existing resources | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
