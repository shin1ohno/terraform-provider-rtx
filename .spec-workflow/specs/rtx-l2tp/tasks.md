# Tasks Document: rtx_l2tp

## Phase 1: Parser Layer

- [ ] 1. Create L2TPConfig data model and parser
  - File: internal/rtx/parsers/l2tp.go
  - Define L2TPConfig, L2TPAuth, L2TPIPPool, L2TPIPsec, L2TPv3Config structs
  - Define L2TPTunnelAuth, L2TPKeepalive structs for extended L2TPv3 settings
  - Implement ParseL2TPConfig() to parse RTX output for L2TP/L2TPv3 tunnels
  - Purpose: Parse "show config" output for L2TP configuration
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-l2tp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create L2TPConfig struct with ID, Name, Version, Mode, Shutdown, TunnelSource, TunnelDest, Authentication, IPPool, IPsecProfile, L2TPv3Config, KeepaliveEnabled, DisconnectTime fields. Create L2TPAuth, L2TPIPPool, L2TPIPsec, L2TPv3Config, L2TPTunnelAuth, L2TPKeepalive nested structs. Implement ParseL2TPConfig() to parse RTX router output from "show config" for L2TP/L2TPv3 tunnel configuration. Support both L2TPv2 (pp select anonymous) and L2TPv3 (tunnel select) configurations | Restrictions: Do not modify existing parser files, use standard library regexp, handle both L2TP versions | _Leverage: internal/rtx/parsers/dhcp_scope.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (Tunnel Config), Requirement 3 (Authentication), Requirement 4 (L2TPv3 Config) | Success: Parser correctly extracts all L2TP attributes from sample RTX output, handles L2TPv2 and L2TPv3 configurations | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 2. Create command builder functions for L2TP
  - File: internal/rtx/parsers/l2tp.go (continue)
  - Implement BuildL2TPServiceCommand() for enabling L2TP service
  - Implement BuildPPSelectAnonymousCommand() for L2TPv2 LNS
  - Implement BuildPPBindTunnelCommand() for tunnel binding
  - Implement BuildPPAuthAcceptCommand() for authentication method
  - Implement BuildPPAuthMynameCommand() for credentials
  - Implement BuildTunnelEncapsulationCommand() for tunnel type
  - Implement BuildTunnelEndpointCommand() for L2TPv3 endpoints
  - Implement BuildL2TPRouterIDCommands() for L2TPv3 router IDs
  - Implement BuildL2TPKeepaliveCommand() for keepalive settings
  - Implement BuildL2TPDisconnectTimeCommand() for idle timeout
  - Implement BuildDeleteL2TPCommand() for deletion
  - Purpose: Generate RTX CLI commands for L2TP management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go command builder pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-l2tp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax for L2TPv2: "l2tp service on", "pp select anonymous", "pp bind tunnel<n>", "pp auth accept <method>", "pp auth myname <name> <pass>", "tunnel encapsulation <n> l2tp". For L2TPv3: "tunnel encapsulation l2tpv3", "tunnel endpoint address <local> <remote>", "l2tp local router-id <ip>", "l2tp remote router-id <ip>", "l2tp remote end-id <string>", "l2tp always-on on/off", "l2tp keepalive use on <interval> <retry>", "l2tp tunnel disconnect time <n>" | Restrictions: Follow existing command builder patterns exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax for both L2TPv2 and L2TPv3 | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 3. Create parser unit tests
  - File: internal/rtx/parsers/l2tp_test.go
  - Test ParseL2TPConfig with various RTX output formats
  - Test L2TPv2 LNS configuration parsing
  - Test L2TPv3 L2VPN configuration parsing
  - Test all command builder functions
  - Test edge cases: missing fields, malformed input
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-l2tp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for l2tp.go. Include test cases for parsing L2TPv2 LNS config output, L2TPv3 L2VPN config output, command building with various parameter combinations, edge cases like missing optional fields, different authentication methods | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases for both L2TP versions handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [ ] 4. Add L2TP types to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add L2TPConfig struct with all fields
  - Add L2TPAuth, L2TPIPPool, L2TPIPsec structs
  - Add L2TPv3Config, L2TPTunnelAuth, L2TPKeepalive structs
  - Extend Client interface with L2TP methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-l2tp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add L2TPConfig struct (ID int, Name string, Version string, Mode string, Shutdown bool, TunnelSource string, TunnelDest string, Authentication L2TPAuth, IPPool *L2TPIPPool, IPsecProfile *L2TPIPsec, L2TPv3Config *L2TPv3Config, KeepaliveEnabled bool, DisconnectTime int). Add nested structs for authentication, IP pool, IPsec, L2TPv3 config. Extend Client interface with: GetL2TP(ctx, tunnelID) (*L2TPConfig, error), CreateL2TP(ctx, l2tp) error, UpdateL2TP(ctx, l2tp) error, DeleteL2TP(ctx, tunnelID) error, ListL2TPs(ctx) ([]L2TPConfig, error) | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 5. Create L2TPService implementation
  - File: internal/client/l2tp_service.go (new)
  - Implement L2TPService struct with executor reference
  - Implement Create() with validation for L2TPv2/L2TPv3
  - Implement Get() to parse L2TP configuration
  - Implement Update() for modifying settings
  - Implement Delete() with proper cleanup of tunnel and PP
  - Implement List() to retrieve all L2TP tunnels
  - Purpose: Service layer for L2TP CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4, 5_
  - _Prompt: Implement the task for spec rtx-l2tp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create L2TPService following DHCPScopeService pattern. Include input validation (tunnel ID, IP addresses, authentication method, L2TPv3 router IDs). Handle L2TPv2 vs L2TPv3 configuration differences. Use parsers.BuildL2TP* command functions. Call client.SaveConfig() after modifications. Handle tunnel update by reconfiguring specific settings | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, handle both L2TP versions properly | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-5 | Success: All CRUD operations work for L2TPv2 and L2TPv3, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 6. Integrate L2TPService into rtxClient
  - File: internal/client/client.go (modify)
  - Add l2tpService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface L2TP methods delegating to service
  - Purpose: Wire up L2TP service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-l2tp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add l2tpService *L2TPService field to rtxClient. Initialize in Dial(): c.l2tpService = NewL2TPService(c.executor, c). Implement GetL2TP, CreateL2TP, UpdateL2TP, DeleteL2TP, ListL2TPs methods delegating to service | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all L2TP methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 7. Create service unit tests
  - File: internal/client/l2tp_service_test.go (new)
  - Test Create with valid and invalid inputs for L2TPv2
  - Test Create with valid and invalid inputs for L2TPv3
  - Test Get parsing for both versions
  - Test Update behavior
  - Test Delete
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-l2tp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for L2TPService. Mock Executor interface to simulate RTX responses. Test validation (invalid tunnel ID, invalid mode, invalid auth method, invalid IPs). Test successful CRUD operations for L2TPv2 LNS and L2TPv3 L2VPN. Test error handling | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested for both L2TP versions, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [ ] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_l2tp.go (new)
  - Define resourceRTXL2TP() with full schema
  - Add id (Required, ForceNew, Int) - tunnel ID
  - Add name (Optional, String) - description
  - Add version (Required, ForceNew, String) - "l2tp" or "l2tpv3"
  - Add mode (Required, ForceNew, String) - "lns" or "l2vpn"
  - Add shutdown (Optional, Bool, default false)
  - Add tunnel_source (Required, String)
  - Add tunnel_destination (Required, String)
  - Add tunnel_dest_type (Optional, String) - "ip" or "fqdn"
  - Add authentication block (Optional) with method, username, password
  - Add ip_pool block (Optional) with start, end
  - Add ipsec_profile block (Optional) with enabled, pre_shared_key
  - Add l2tpv3_config block (Optional) with L2TPv3-specific settings
  - Add keepalive_enabled (Optional, Bool)
  - Add disconnect_time (Optional, Int)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-l2tp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXL2TP() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Add ValidateFunc for version (l2tp/l2tpv3), mode (lns/l2vpn), auth method (pap/chap). Set ForceNew on id, version, mode. Use TypeList for authentication, ip_pool, ipsec_profile, l2tpv3_config blocks. Mark password and pre_shared_key as Sensitive | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1-4 | Success: Schema compiles, validation functions work, sensitive fields marked | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_l2tp.go (continue)
  - Implement resourceRTXL2TPCreate()
  - Implement resourceRTXL2TPRead()
  - Implement resourceRTXL2TPUpdate()
  - Implement resourceRTXL2TPDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-l2tp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build L2TPConfig from ResourceData, call client.CreateL2TP, set ID to tunnel ID). Read (call GetL2TP, update ResourceData, handle not found by clearing ID). Update (call UpdateL2TP for mutable fields). Delete (call DeleteL2TP). Follow rtx_dhcp_scope patterns for apiClient access and error handling. Handle version-specific fields (authentication/ip_pool for L2TPv2, l2tpv3_config for L2TPv3) | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end for both L2TP versions | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 10. Implement import functionality
  - File: internal/provider/resource_rtx_l2tp.go (continue)
  - Implement resourceRTXL2TPImport()
  - Parse tunnel_id from import ID string
  - Validate L2TP tunnel exists on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_scope.go import pattern_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec rtx-l2tp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXL2TPImport(). Parse import ID as tunnel_id integer. Call GetL2TP to verify existence. Populate all ResourceData fields from retrieved L2TP config. Call Read to ensure state consistency | Restrictions: Handle invalid import ID format, non-existent tunnel errors gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go import function | _Requirements: Requirement 5 (Import) | Success: terraform import rtx_l2tp.example 1 works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_l2tp" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-l2tp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_l2tp": resourceRTXL2TP() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 12. Create resource unit tests
  - File: internal/provider/resource_rtx_l2tp_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client for L2TPv2
  - Test CRUD operations with mock client for L2TPv3
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-l2tp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_l2tp.go. Test schema validation (invalid version, invalid mode, invalid auth method). Test CRUD operations with mocked client for L2TPv2 LNS and L2TPv3 L2VPN configurations. Test import with valid and invalid IDs | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths for both versions | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [ ] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_l2tp_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test L2TPv2 LNS creation with all parameters
  - Test L2TPv3 L2VPN creation with all parameters
  - Test L2TP update
  - Test L2TP import
  - Test L2TP with IPsec integration
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_test.go acceptance test patterns_
  - _Requirements: 1, 5, 6_
  - _Prompt: Implement the task for spec rtx-l2tp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test L2TPv2 LNS config with authentication and IP pool. Test L2TPv3 L2VPN config with router IDs and bridge interface. Test update of mutable fields. Test import existing L2TP tunnel. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5, 6 | Success: Acceptance tests pass against real RTX router for both L2TP versions | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 14. Add example Terraform configurations
  - File: examples/l2tp/main.tf (new)
  - L2TPv2 remote access VPN server example
  - L2TPv3 site-to-site L2VPN example
  - L2TPv3 with IPsec encryption example
  - L2TPv3 without IPsec (unencrypted) example
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1, 6_
  - _Prompt: Implement the task for spec rtx-l2tp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. L2TPv2 LNS: remote access VPN server with authentication, IP pool, IPsec. L2TPv3 L2VPN: site-to-site with router IDs, bridge interface. L2TPv3 with IPsec: encrypted L2VPN tunnel. L2TPv3 unencrypted: internal network L2VPN without IPsec | Restrictions: Use realistic IP addresses, include comments explaining L2TPv2 vs L2TPv3 differences and use cases | _Leverage: examples/dhcp_scope/ | _Requirements: Requirements 1, 6 | Success: Examples are valid Terraform, demonstrate all L2TP features for both versions | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually for both L2TPv2 and L2TPv3
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-l2tp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works for L2TPv2 LNS and L2TPv3 L2VPN. Check terraform import functionality for both versions. Ensure no regressions in existing resources | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end for both L2TP versions | After completing, use log-implementation tool to record details, then mark as [x] complete_
