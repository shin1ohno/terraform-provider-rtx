# Tasks Document: rtx_bgp

## Phase 1: Parser Layer

- [x] 1. Create BGP data model and parser
  - File: internal/rtx/parsers/bgp.go
  - Define BGPConfig, BGPNeighbor, and BGPNetwork structs
  - Implement ParseBGPConfig() to parse RTX output
  - Purpose: Parse "show config | grep bgp" output
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-bgp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create BGPConfig struct with Enabled, ASN, RouterID, DefaultIPv4Unicast, LogNeighborChanges, Neighbors, Networks, RedistributeStatic, RedistributeConnected fields. Create BGPNeighbor struct with ID, IP, RemoteAS, HoldTime, Keepalive, Multihop, Password, LocalAddress fields. Create BGPNetwork struct with Prefix, Mask fields. Implement ParseBGPConfig() function to parse RTX router output from "show config | grep bgp" command. Follow patterns from dhcp_scope.go | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line BGP configurations | _Leverage: internal/rtx/parsers/dhcp_scope.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (AS/Router ID), Requirement 3 (Neighbors), Requirement 4 (Networks) | Success: Parser correctly extracts all BGP attributes from sample RTX output, handles edge cases like missing optional fields | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [x] 2. Create command builder functions for BGP
  - File: internal/rtx/parsers/bgp.go (continue)
  - Implement BuildBGPUseCommand() for enable/disable
  - Implement BuildBGPASNCommand() for AS number configuration
  - Implement BuildBGPRouterIDCommand() for router ID configuration
  - Implement BuildBGPNeighborCommand() for neighbor configuration
  - Implement BuildBGPNetworkCommand() for network announcements
  - Implement BuildBGPRedistributeCommand() for route redistribution
  - Purpose: Generate RTX CLI commands for BGP management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go BuildDHCPScopeCommand pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-bgp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "bgp use on/off", "bgp autonomous-system <asn>", "bgp router id <router_id>", "bgp neighbor <n> address <ip> as <asn>", "bgp neighbor <n> hold-time <time>", "bgp neighbor <n> local-address <ip>", "bgp neighbor <n> password <password>", "bgp import filter <n> include <network>/<mask>", "bgp import from static/connected" | Restrictions: Follow existing command builder patterns exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 3. Create parser unit tests
  - File: internal/rtx/parsers/bgp_test.go
  - Test ParseBGPConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: missing fields, malformed input, 4-byte ASN
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-bgp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for bgp.go. Include test cases for parsing BGP config output, command building with various parameter combinations, edge cases like empty neighbors list, 4-byte ASN, eBGP vs iBGP detection | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [x] 4. Add BGPConfig type to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add BGPConfig struct with all fields
  - Add BGPNeighbor struct
  - Add BGPNetwork struct
  - Extend Client interface with BGP methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-bgp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add BGPConfig struct (Enabled bool, ASN string, RouterID string, DefaultIPv4Unicast bool, LogNeighborChanges bool, Neighbors []BGPNeighbor, Networks []BGPNetwork, RedistributeStatic bool, RedistributeConnected bool). Add BGPNeighbor struct (ID int, IP string, RemoteAS string, HoldTime int, Keepalive int, Multihop int, Password string, LocalAddress string). Add BGPNetwork struct (Prefix string, Mask string). Extend Client interface with: GetBGPConfig(ctx) (*BGPConfig, error), ConfigureBGP(ctx, config) error, UpdateBGPConfig(ctx, config) error, ResetBGP(ctx) error | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 5. Create BGPService implementation
  - File: internal/client/bgp_service.go (new)
  - Implement BGPService struct with executor reference
  - Implement Configure() with validation and command execution
  - Implement Get() to parse BGP configuration
  - Implement Update() for modifying BGP settings
  - Implement Reset() to disable BGP
  - Purpose: Service layer for BGP CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4, 5_
  - _Prompt: Implement the task for spec rtx-bgp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create BGPService following DHCPScopeService pattern. Include input validation (ASN range 1-4294967295, valid IPv4 for router_id, valid neighbor IPs). Use parsers.BuildBGP* functions. Call client.SaveConfig() after modifications. Handle neighbor updates by removing old neighbors and adding new ones. Ensure correct order of operations: enable BGP first, then configure AS, router ID, neighbors, networks | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, maintain separation from other services | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-5 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 6. Integrate BGPService into rtxClient
  - File: internal/client/client.go (modify)
  - Add bgpService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface BGP methods delegating to service
  - Purpose: Wire up BGP service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-bgp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add bgpService *BGPService field to rtxClient. Initialize in Dial(): c.bgpService = NewBGPService(c.executor, c). Implement GetBGPConfig, ConfigureBGP, UpdateBGPConfig, ResetBGP methods delegating to service | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all BGP methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 7. Create service unit tests
  - File: internal/client/bgp_service_test.go (new)
  - Test Configure with valid and invalid inputs
  - Test Get parsing
  - Test Update behavior
  - Test Reset
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-bgp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for BGPService. Mock Executor interface to simulate RTX responses. Test validation (invalid ASN, invalid router ID, invalid neighbor IPs). Test successful CRUD operations. Test error handling. Test neighbor configuration with various parameters (hold-time, multihop, password) | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [x] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_bgp.go (new)
  - Define resourceRTXBGP() with full schema
  - Add asn (Required, String for 4-byte ASN support)
  - Add router_id (Optional, String with IPv4 validation)
  - Add default_ipv4_unicast (Optional, Bool, default true)
  - Add log_neighbor_changes (Optional, Bool, default true)
  - Add neighbors (Optional, List of Object with ip, remote_as, hold_time, keepalive, multihop, password, local_address)
  - Add networks (Optional, List of Object with prefix, mask)
  - Add redistribute_static (Optional, Bool)
  - Add redistribute_connected (Optional, Bool)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-bgp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXBGP() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Add ValidateFunc for asn (1-4294967295), router_id (valid IPv4). Use TypeList for neighbors with nested schema (ip, remote_as required; hold_time, keepalive, multihop, password, local_address optional). Mark password as Sensitive. Use TypeList for networks with nested schema (prefix, mask). Resource is singleton with fixed ID "bgp" | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1-4 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_bgp.go (continue)
  - Implement resourceRTXBGPCreate()
  - Implement resourceRTXBGPRead()
  - Implement resourceRTXBGPUpdate()
  - Implement resourceRTXBGPDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-bgp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build BGPConfig from ResourceData, call client.ConfigureBGP, set ID to "bgp"). Read (call GetBGPConfig, update ResourceData, handle not found by clearing ID). Update (call UpdateBGPConfig for mutable fields). Delete (call ResetBGP to disable BGP). Follow rtx_dhcp_scope patterns for apiClient access and error handling. Handle neighbors array conversion. Mark password fields as sensitive | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 10. Implement import functionality
  - File: internal/provider/resource_rtx_bgp.go (continue)
  - Implement resourceRTXBGPImport()
  - Use fixed ID "bgp" since BGP is a singleton resource
  - Validate BGP is configured on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_scope.go import pattern_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec rtx-bgp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXBGPImport(). Accept "bgp" as the import ID (singleton resource). Call GetBGPConfig to verify BGP is configured. Populate all ResourceData fields from retrieved config. Call Read to ensure state consistency | Restrictions: Handle non-existent BGP configuration errors gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go import function | _Requirements: Requirement 5 (Import) | Success: terraform import rtx_bgp.main bgp works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_bgp" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-bgp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_bgp": resourceRTXBGP() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 12. Create resource unit tests
  - File: internal/provider/resource_rtx_bgp_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-bgp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_bgp.go. Test schema validation (invalid ASN, invalid router_id). Test CRUD operations with mocked client. Test import with valid and invalid IDs. Test neighbor configuration handling. Test sensitive password field handling | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [x] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_bgp_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test BGP configuration with AS number and router ID
  - Test neighbor configuration
  - Test network announcements
  - Test redistribution settings
  - Test BGP import
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_test.go acceptance test patterns_
  - _Requirements: 1, 5, 6_
  - _Prompt: Implement the task for spec rtx-bgp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with BGP enable, AS number, router ID, neighbors with various options (hold-time, multihop, password), network announcements, redistribution. Test update neighbors. Test import existing BGP config | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5, 6 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 14. Add example Terraform configurations
  - File: examples/bgp/main.tf (new)
  - Basic BGP configuration example
  - BGP with all options example
  - Multi-neighbor BGP example
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1, 6_
  - _Prompt: Implement the task for spec rtx-bgp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: BGP with AS number only. Full: BGP with router_id, neighbors with hold_time and password, networks, redistribution. Multi-peer: BGP with multiple eBGP and iBGP neighbors demonstrating different configurations | Restrictions: Use realistic AS numbers and IP addresses, include comments explaining options, demonstrate eBGP vs iBGP | _Leverage: examples/dhcp_scope/ | _Requirements: Requirements 1, 6 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-bgp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works. Check terraform import functionality. Ensure no regressions in existing resources | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
