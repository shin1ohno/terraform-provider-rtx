# Tasks Document: rtx_ipsec_tunnel

## Phase 1: Parser Layer

- [ ] 1. Create IPsecTunnel data model and parser
  - File: internal/rtx/parsers/ipsec_tunnel.go
  - Define IPsecTunnel, IKEv2Proposal, and IPsecTransform structs
  - Implement ParseIPsecTunnelConfig() to parse RTX output
  - Purpose: Parse "show config | grep ipsec" and "show ipsec sa" output
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ipsec-tunnel, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create IPsecTunnel struct with ID, Name, LocalAddress, RemoteAddress, PreSharedKey, IKEv2Proposal, IPsecTransform, LocalNetwork, RemoteNetwork, DPDEnabled, DPDInterval fields. Create IKEv2Proposal struct with EncryptionAES256, IntegritySHA256, GroupFourteen, LifetimeSeconds. Create IPsecTransform struct with Protocol, EncryptionAES256, IntegritySHA256, PFSGroupFourteen, LifetimeSeconds. Implement ParseIPsecTunnelConfig() function to parse RTX router output from "show config | grep ipsec" command. Follow patterns from dhcp_scope.go | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line IPsec configurations | _Leverage: internal/rtx/parsers/dhcp_scope.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (IKE Phase 1), Requirement 3 (IPsec Phase 2), Requirement 4 (DPD) | Success: Parser correctly extracts all tunnel attributes from sample RTX output, handles edge cases like missing optional fields | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 2. Create command builder functions for IPsec tunnel
  - File: internal/rtx/parsers/ipsec_tunnel.go (continue)
  - Implement BuildTunnelSelectCommand() for tunnel selection
  - Implement BuildIPsecTunnelCommand() for tunnel creation
  - Implement BuildIPsecSAPolicyCommand() for SA policy configuration
  - Implement BuildIPsecIKEPreSharedKeyCommand() for PSK configuration
  - Implement BuildIPsecIKERemoteAddressCommand() for remote address
  - Implement BuildIPsecIKEEncryptionCommand() for IKE encryption
  - Implement BuildIPsecIKEGroupCommand() for DH group configuration
  - Implement BuildIPsecIKEKeepaliveCommand() for DPD configuration
  - Implement BuildDeleteIPsecTunnelCommand() for deletion
  - Purpose: Generate RTX CLI commands for IPsec tunnel management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go BuildDHCPScopeCommand pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ipsec-tunnel, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "tunnel select <n>", "ipsec tunnel <n>", "ipsec ike pre-shared-key <n> text <key>", "ipsec ike remote address <n> <ip>", "ipsec ike encryption <n> aes-cbc", "ipsec ike hash <n> sha256", "ipsec ike group <n> modp2048", "ipsec sa policy <n> <tunnel> esp aes-cbc sha-hmac", "ipsec ike keepalive use <n> on dpd <interval>", "no tunnel select <n>", "no ipsec tunnel <n>". Map algorithm names (encryption_aes_cbc_256 -> aes-cbc, group_fourteen -> modp2048) | Restrictions: Follow existing command builder patterns exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax, algorithm mapping works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 3. Create parser unit tests
  - File: internal/rtx/parsers/ipsec_tunnel_test.go
  - Test ParseIPsecTunnelConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: missing fields, malformed input, various algorithm combinations
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ipsec-tunnel, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for ipsec_tunnel.go. Include test cases for parsing IPsec config output, command building with various parameter combinations, edge cases like different encryption algorithms, DH groups, DPD enabled/disabled | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [ ] 4. Add IPsecTunnel types to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add IPsecTunnel struct with all fields
  - Add IKEv2Proposal struct
  - Add IPsecTransform struct
  - Extend Client interface with IPsec tunnel methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ipsec-tunnel, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add IPsecTunnel struct (ID int, Name string, LocalAddress string, RemoteAddress string, PreSharedKey string, IKEv2Proposal IKEv2Proposal, IPsecTransform IPsecTransform, LocalNetwork string, RemoteNetwork string, DPDEnabled bool, DPDInterval int). Add IKEv2Proposal struct (EncryptionAES256 bool, IntegritySHA256 bool, GroupFourteen bool, LifetimeSeconds int). Add IPsecTransform struct (Protocol string, EncryptionAES256 bool, IntegritySHA256 bool, PFSGroupFourteen bool, LifetimeSeconds int). Extend Client interface with: GetIPsecTunnel(ctx, tunnelID) (*IPsecTunnel, error), CreateIPsecTunnel(ctx, tunnel) error, UpdateIPsecTunnel(ctx, tunnel) error, DeleteIPsecTunnel(ctx, tunnelID) error, ListIPsecTunnels(ctx) ([]IPsecTunnel, error) | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 5. Create IPsecTunnelService implementation
  - File: internal/client/ipsec_tunnel_service.go (new)
  - Implement IPsecTunnelService struct with executor reference
  - Implement Create() with validation and multi-command execution
  - Implement Get() to parse tunnel configuration
  - Implement Update() for modifying tunnel settings
  - Implement Delete() for tunnel removal
  - Implement List() to retrieve all tunnels
  - Purpose: Service layer for IPsec tunnel CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ipsec-tunnel, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create IPsecTunnelService following DHCPScopeService pattern. Include input validation (IP addresses, tunnel ID, algorithm validity). Use parsers.BuildTunnelSelectCommand, BuildIPsecTunnelCommand, BuildIPsecIKEPreSharedKeyCommand, and related functions. Execute multiple commands in sequence for full tunnel creation. Call client.SaveConfig() after modifications. Handle tunnel update by clearing SA and reconfiguring | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, maintain separation from other services | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-4 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 6. Integrate IPsecTunnelService into rtxClient
  - File: internal/client/client.go (modify)
  - Add ipsecTunnelService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface IPsec methods delegating to service
  - Purpose: Wire up IPsec tunnel service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ipsec-tunnel, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add ipsecTunnelService *IPsecTunnelService field to rtxClient. Initialize in Dial(): c.ipsecTunnelService = NewIPsecTunnelService(c.executor, c). Implement GetIPsecTunnel, CreateIPsecTunnel, UpdateIPsecTunnel, DeleteIPsecTunnel, ListIPsecTunnels methods delegating to service | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all IPsec methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 7. Create service unit tests
  - File: internal/client/ipsec_tunnel_service_test.go (new)
  - Test Create with valid and invalid inputs
  - Test Get parsing
  - Test Update behavior
  - Test Delete
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ipsec-tunnel, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for IPsecTunnelService. Mock Executor interface to simulate RTX responses. Test validation (invalid tunnel ID, invalid IP addresses, invalid algorithms). Test successful CRUD operations. Test error handling for command failures | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [ ] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_ipsec_tunnel.go (new)
  - Define resourceRTXIPsecTunnel() with full schema
  - Add id (Required, ForceNew, Int)
  - Add name (Optional, String)
  - Add local_address (Required, String with IP validation)
  - Add remote_address (Required, String with IP validation)
  - Add pre_shared_key (Required, Sensitive, String)
  - Add ikev2_proposal (Required, Block with nested attributes)
  - Add ipsec_transform (Required, Block with nested attributes)
  - Add local_network (Required, String with CIDR validation)
  - Add remote_network (Required, String with CIDR validation)
  - Add dpd_enabled (Optional, Bool, default true)
  - Add dpd_interval (Optional, Int, default 30)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-ipsec-tunnel, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXIPsecTunnel() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Add ValidateFunc for local_address, remote_address (valid IPs), local_network, remote_network (valid CIDR). Set ForceNew on id. Mark pre_shared_key as Sensitive. Use TypeList with MaxItems 1 for ikev2_proposal with nested schema (encryption_aes_cbc_256 bool, integrity_sha256 bool, group_fourteen bool, lifetime_seconds int). Use TypeList with MaxItems 1 for ipsec_transform with nested schema (protocol string, encryption_aes_cbc_256 bool, integrity_sha256_hmac bool, pfs_group_fourteen bool, lifetime_seconds int) | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1-4 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_ipsec_tunnel.go (continue)
  - Implement resourceRTXIPsecTunnelCreate()
  - Implement resourceRTXIPsecTunnelRead()
  - Implement resourceRTXIPsecTunnelUpdate()
  - Implement resourceRTXIPsecTunnelDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ipsec-tunnel, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build IPsecTunnel from ResourceData including nested ikev2_proposal and ipsec_transform blocks, call client.CreateIPsecTunnel, set ID to tunnel id). Read (call GetIPsecTunnel, update ResourceData including nested blocks, handle not found by clearing ID). Update (call UpdateIPsecTunnel for mutable fields). Delete (call DeleteIPsecTunnel). Follow rtx_dhcp_scope patterns for apiClient access and error handling. Handle sensitive pre_shared_key appropriately | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully, do not expose PSK in logs | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 10. Implement import functionality
  - File: internal/provider/resource_rtx_ipsec_tunnel.go (continue)
  - Implement resourceRTXIPsecTunnelImport()
  - Parse tunnel_id from import ID string
  - Validate tunnel exists on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_scope.go import pattern_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec rtx-ipsec-tunnel, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXIPsecTunnelImport(). Parse import ID as tunnel_id integer. Call GetIPsecTunnel to verify existence. Populate all ResourceData fields from retrieved tunnel including nested blocks. Note: pre_shared_key may not be retrievable from router config, handle appropriately | Restrictions: Handle invalid import ID format, non-existent tunnel errors gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go import function | _Requirements: Requirement 5 (Import) | Success: terraform import rtx_ipsec_tunnel.example 1 works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_ipsec_tunnel" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-ipsec-tunnel, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_ipsec_tunnel": resourceRTXIPsecTunnel() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 12. Create resource unit tests
  - File: internal/provider/resource_rtx_ipsec_tunnel_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Test sensitive attribute handling
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-ipsec-tunnel, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_ipsec_tunnel.go. Test schema validation (invalid IP, invalid CIDR, invalid algorithm combinations). Test CRUD operations with mocked client. Test import with valid and invalid IDs. Test sensitive attribute (pre_shared_key) is marked sensitive | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [ ] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_ipsec_tunnel_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test tunnel creation with all parameters
  - Test tunnel update
  - Test tunnel import
  - Test various algorithm combinations
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_acc_test.go acceptance test patterns_
  - _Requirements: 1, 5, 6_
  - _Prompt: Implement the task for spec rtx-ipsec-tunnel, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with tunnel creation including IKE and IPsec settings, update DPD settings, import existing tunnel. Test various encryption/integrity algorithm combinations. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set, use test PSK values | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5, 6 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 14. Add example Terraform configurations
  - File: examples/ipsec_tunnel/main.tf (new)
  - Basic site-to-site VPN example
  - Full tunnel with all options example
  - Multiple tunnel example for hub-and-spoke
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1, 6_
  - _Prompt: Implement the task for spec rtx-ipsec-tunnel, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: site-to-site VPN tunnel with minimal settings. Full: tunnel with all IKE, IPsec, and DPD settings explicitly configured. Hub-spoke: multiple tunnels demonstrating branch office connectivity pattern. Include variable for pre_shared_key to demonstrate secure handling | Restrictions: Use realistic but non-routable IP addresses (documentation ranges), include comments explaining security considerations | _Leverage: examples/dhcp_scope/ | _Requirements: Requirements 1, 6 | Success: Examples are valid Terraform, demonstrate all features, follow security best practices | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-ipsec-tunnel, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works for IPsec tunnels. Check terraform import functionality. Ensure no regressions in existing resources. Verify sensitive attribute handling for pre_shared_key | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
