# Tasks Document: rtx_pptp

## Phase 1: Parser Layer

- [x] 1. Create PPTPConfig data model and parser
  - File: internal/rtx/parsers/pptp.go
  - Define PPTPConfig, PPTPAuth, PPTPEncryption, PPTPIPPool structs
  - Implement ParsePPTPConfig() to parse RTX output
  - Purpose: Parse "show config | grep pptp" output
  - _Leverage: internal/rtx/parsers/l2tp.go for VPN patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-pptp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create PPTPConfig struct with Shutdown, ListenAddress, MaxConnections, Authentication, Encryption, IPPool, DisconnectTime, KeepaliveEnabled fields. Create PPTPAuth struct with Method, Username, Password. Create PPTPEncryption struct with MPPEBits, Required. Create PPTPIPPool struct with Start, End. Implement ParsePPTPConfig() function to parse RTX router output from "show config | grep pptp" command. Follow patterns from l2tp.go | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line PPTP configurations | _Leverage: internal/rtx/parsers/l2tp.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (Authentication), Requirement 3 (Encryption), Requirement 4 (IP Pool) | Success: Parser correctly extracts all PPTP attributes from sample RTX output, handles edge cases like missing optional fields | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [x] 2. Create command builder functions for PPTP
  - File: internal/rtx/parsers/pptp.go (continue)
  - Implement BuildPPTPServiceCommand() for enabling/disabling PPTP
  - Implement BuildPPTPTunnelDisconnectTimeCommand() for idle timeout
  - Implement BuildPPTPKeepaliveCommand() for keepalive settings
  - Implement BuildPPAuthAcceptCommand() for authentication method
  - Implement BuildPPPCCPTypeCommand() for MPPE encryption
  - Implement BuildDeletePPTPCommand() for deletion
  - Purpose: Generate RTX CLI commands for PPTP management
  - _Leverage: internal/rtx/parsers/l2tp.go command builder patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-pptp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "pptp service on/off", "pptp tunnel disconnect time <n>", "pptp keepalive use on/off", "pp auth accept <method>", "ppp ccp type mppe-128/mppe-any", "ip pp remote address pool <start>-<end>". Handle PP anonymous configuration | Restrictions: Follow existing command builder patterns exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/l2tp.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 3. Create parser unit tests
  - File: internal/rtx/parsers/pptp_test.go
  - Test ParsePPTPConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: missing fields, malformed input
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/l2tp_test.go for test patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-pptp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for pptp.go. Include test cases for parsing PPTP config output, command building with various parameter combinations, edge cases like missing authentication, different MPPE bit settings (40, 56, 128) | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/l2tp_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [x] 4. Add PPTP types to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add PPTPConfig struct with all fields
  - Add PPTPAuth, PPTPEncryption, PPTPIPPool structs
  - Extend Client interface with PPTP methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing L2TPConfig struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-pptp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add PPTPConfig struct (Shutdown bool, ListenAddress string, MaxConnections int, Authentication PPTPAuth, Encryption PPTPEncryption, IPPool *PPTPIPPool, DisconnectTime int, KeepaliveEnabled bool). Add PPTPAuth struct (Method, Username, Password string). Add PPTPEncryption struct (MPPEBits int, Required bool). Add PPTPIPPool struct (Start, End string). Extend Client interface with: GetPPTP(ctx) (*PPTPConfig, error), CreatePPTP(ctx, pptp) error, UpdatePPTP(ctx, pptp) error, DeletePPTP(ctx) error | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 5. Create PPTPService implementation
  - File: internal/client/pptp_service.go (new)
  - Implement PPTPService struct with executor reference
  - Implement Create() with validation and command execution
  - Implement Get() to parse PPTP configuration
  - Implement Update() for modifying settings
  - Implement Delete() to remove PPTP configuration
  - Purpose: Service layer for PPTP CRUD operations
  - _Leverage: internal/client/l2tp_service.go for VPN service pattern_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-pptp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create PPTPService following L2TPService pattern. Include input validation (authentication method is pap/chap/mschap/mschap-v2, MPPE bits is 40/56/128, valid IP addresses). Use parsers.BuildPPTPServiceCommand and related functions. Call client.SaveConfig() after modifications. Handle PP anonymous configuration | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, PPTP is singleton resource | _Leverage: internal/client/l2tp_service.go | _Requirements: Requirements 1-4 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 6. Integrate PPTPService into rtxClient
  - File: internal/client/client.go (modify)
  - Add pptpService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface PPTP methods delegating to service
  - Purpose: Wire up PPTP service to main client
  - _Leverage: existing l2tpService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-pptp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add pptpService *PPTPService field to rtxClient. Initialize in Dial(): c.pptpService = NewPPTPService(c.executor, c). Implement GetPPTP, CreatePPTP, UpdatePPTP, DeletePPTP methods delegating to service | Restrictions: Follow existing l2tpService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go l2tpService integration | _Requirements: Requirement 1 | Success: Client compiles, all PPTP methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 7. Create service unit tests
  - File: internal/client/pptp_service_test.go (new)
  - Test Create with valid and invalid inputs
  - Test Get parsing
  - Test Update behavior
  - Test Delete
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/l2tp_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-pptp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for PPTPService. Mock Executor interface to simulate RTX responses. Test validation (invalid auth method, invalid MPPE bits, invalid IPs). Test successful CRUD operations. Test error handling | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/l2tp_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [x] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_pptp.go (new)
  - Define resourceRTXPPTP() with full schema
  - Add shutdown (Optional, Bool, default false)
  - Add listen_address (Optional, String, default "0.0.0.0")
  - Add max_connections (Optional, Int)
  - Add authentication block (Required, with method, username, password)
  - Add encryption block (Optional, with mppe_bits, required)
  - Add ip_pool block (Optional, with start, end)
  - Add disconnect_time (Optional, Int)
  - Add keepalive_enabled (Optional, Bool)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_l2tp.go for VPN resource patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-pptp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXPPTP() returning *schema.Resource. Define schema following rtx_l2tp patterns. Add ValidateFunc for authentication method (pap/chap/mschap/mschap-v2), mppe_bits (40/56/128). Mark password as Sensitive. Use TypeList with MaxItems 1 for authentication, encryption, ip_pool blocks | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_l2tp.go | _Requirements: Requirements 1-4 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_pptp.go (continue)
  - Implement resourceRTXPPTPCreate()
  - Implement resourceRTXPPTPRead()
  - Implement resourceRTXPPTPUpdate()
  - Implement resourceRTXPPTPDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_l2tp.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-pptp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build PPTPConfig from ResourceData, call client.CreatePPTP, set ID to "pptp"). Read (call GetPPTP, update ResourceData, handle not found by clearing ID). Update (call UpdatePPTP). Delete (call DeletePPTP). Follow rtx_l2tp patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully, PPTP is singleton so ID is always "pptp" | _Leverage: internal/provider/resource_rtx_l2tp.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 10. Implement import functionality
  - File: internal/provider/resource_rtx_pptp.go (continue)
  - Implement resourceRTXPPTPImport()
  - Parse import ID (should be "pptp")
  - Validate PPTP configuration exists on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_l2tp.go import pattern_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec rtx-pptp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXPPTPImport(). Import ID should be "pptp" (singleton resource). Call GetPPTP to verify existence. Populate all ResourceData fields from retrieved config. Call Read to ensure state consistency | Restrictions: Handle non-existent configuration errors gracefully | _Leverage: internal/provider/resource_rtx_l2tp.go import function | _Requirements: Requirement 5 (Import) | Success: terraform import rtx_pptp.vpn_server pptp works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_pptp" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-pptp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_pptp": resourceRTXPPTP() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 12. Create resource unit tests
  - File: internal/provider/resource_rtx_pptp_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_l2tp_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-pptp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_pptp.go. Test schema validation (invalid auth method, invalid MPPE bits). Test CRUD operations with mocked client. Test import with valid and invalid IDs | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_l2tp_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [x] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_pptp_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test PPTP creation with all parameters
  - Test PPTP update
  - Test PPTP import
  - Test with different authentication methods
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_l2tp_test.go acceptance test patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-pptp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with PPTP creation, update authentication method, change MPPE settings, import existing PPTP. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 14. Add example Terraform configurations
  - File: examples/pptp/main.tf (new)
  - Basic PPTP server example
  - PPTP with full options example
  - PPTP with MPPE encryption example
  - Include security warning about PPTP deprecation
  - Purpose: User documentation and testing
  - _Leverage: examples/l2tp/ existing examples_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-pptp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: PPTP server with authentication only. Full: PPTP with all options including MPPE-128, IP pool, keepalive. Add comments warning about PPTP security limitations and recommending L2TP/IPsec or IKEv2 instead | Restrictions: Use realistic IP addresses, include comments explaining options and security considerations | _Leverage: examples/l2tp/ | _Requirements: Requirement 1 | Success: Examples are valid Terraform, demonstrate all features, include security warnings | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-pptp, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works. Check terraform import functionality. Ensure no regressions in existing resources | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
