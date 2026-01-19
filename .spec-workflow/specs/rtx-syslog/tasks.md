# Tasks Document: rtx_syslog

## Phase 1: Parser Layer

- [ ] 1. Create SyslogConfig data model and parser
  - File: internal/rtx/parsers/syslog.go
  - Define SyslogConfig and SyslogHost structs
  - Implement ParseSyslogConfig() to parse RTX output
  - Purpose: Parse "show config | grep syslog" output
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-syslog, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create SyslogConfig struct with Hosts ([]SyslogHost), LocalAddress, Facility, Notice, Info, Debug fields. Create SyslogHost struct with Address and Port fields. Implement ParseSyslogConfig() function to parse RTX router output from "show config | grep syslog" command. Parse syslog host, syslog local address, syslog facility, syslog notice/info/debug settings | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line configurations | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (Host Configuration), Requirement 3 (Facility), Requirement 4 (Log Levels) | Success: Parser correctly extracts all syslog attributes from sample RTX output, handles edge cases like missing optional fields | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 2. Create command builder functions for syslog
  - File: internal/rtx/parsers/syslog.go (continue)
  - Implement BuildSyslogHostCommand() for host configuration
  - Implement BuildSyslogLocalAddressCommand() for local address
  - Implement BuildSyslogFacilityCommand() for facility setting
  - Implement BuildSyslogLevelCommand() for log level on/off
  - Implement BuildDeleteSyslogCommand() for deletion
  - Purpose: Generate RTX CLI commands for syslog management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go BuildDHCPScopeCommand pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-syslog, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "syslog host <address>", "syslog host <address> <port>", "syslog local address <ip>", "syslog facility <facility>", "syslog notice on|off", "syslog info on|off", "syslog debug on|off", "no syslog host <address>", "no syslog local address", etc. | Restrictions: Follow existing command builder pattern exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 3. Create parser unit tests
  - File: internal/rtx/parsers/syslog_test.go
  - Test ParseSyslogConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: missing fields, malformed input, multiple hosts
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-syslog, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for syslog.go. Include test cases for parsing syslog config output, command building with various parameter combinations, edge cases like multiple hosts, custom ports, all facilities | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [ ] 4. Add SyslogConfig type to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add SyslogConfig struct with all fields
  - Add SyslogHost struct
  - Extend Client interface with syslog methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-syslog, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add SyslogConfig struct (Hosts []SyslogHost, LocalAddress string, Facility string, Notice bool, Info bool, Debug bool). Add SyslogHost struct (Address string, Port int). Extend Client interface with: GetSyslogConfig(ctx) (*SyslogConfig, error), ConfigureSyslog(ctx, config) error, UpdateSyslogConfig(ctx, config) error, ResetSyslog(ctx) error | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 5. Create SyslogService implementation
  - File: internal/client/syslog_service.go (new)
  - Implement SyslogService struct with executor reference
  - Implement Configure() with validation and command execution
  - Implement Get() to parse syslog configuration
  - Implement Update() for modifying settings
  - Implement Reset() to remove syslog configuration
  - Purpose: Service layer for syslog CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-syslog, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create SyslogService following DHCPScopeService pattern. Include input validation (IP address format, valid facility names, port range 1-65535). Use parsers.BuildSyslogHostCommand and related functions. Call client.SaveConfig() after modifications. Handle update by removing old config then applying new | Restrictions: Follow existing service patterns exactly, use containsError() for output checking | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-4 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 6. Integrate SyslogService into rtxClient
  - File: internal/client/client.go (modify)
  - Add syslogService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface syslog methods delegating to service
  - Purpose: Wire up syslog service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-syslog, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add syslogService *SyslogService field to rtxClient. Initialize in Dial(): c.syslogService = NewSyslogService(c.executor, c). Implement GetSyslogConfig, ConfigureSyslog, UpdateSyslogConfig, ResetSyslog methods delegating to service | Restrictions: Follow existing service integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all syslog methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 7. Create service unit tests
  - File: internal/client/syslog_service_test.go (new)
  - Test Configure with valid and invalid inputs
  - Test Get parsing
  - Test Update behavior
  - Test Reset
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-syslog, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for SyslogService. Mock Executor interface to simulate RTX responses. Test validation (invalid IP, invalid facility, invalid port). Test successful CRUD operations. Test error handling | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [ ] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_syslog.go (new)
  - Define resourceRTXSyslog() with full schema
  - Add host block (Required, Set of Object with address and port)
  - Add local_address (Optional, String with IP validation)
  - Add facility (Optional, String, default "user")
  - Add notice (Optional, Bool, default true)
  - Add info (Optional, Bool, default false)
  - Add debug (Optional, Bool, default false)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-syslog, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXSyslog() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Add ValidateFunc for local_address (valid IP), facility (valid syslog facility). Use TypeSet for host with nested schema (address string required, port int optional default 514). Syslog is singleton resource, use fixed ID "syslog" | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1-4 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_syslog.go (continue)
  - Implement resourceRTXSyslogCreate()
  - Implement resourceRTXSyslogRead()
  - Implement resourceRTXSyslogUpdate()
  - Implement resourceRTXSyslogDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-syslog, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build SyslogConfig from ResourceData, call client.ConfigureSyslog, set ID to "syslog"). Read (call GetSyslogConfig, update ResourceData, handle not found by clearing ID). Update (call UpdateSyslogConfig for changed fields). Delete (call ResetSyslog). Follow singleton resource patterns - only one syslog config per router | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 10. Implement import functionality
  - File: internal/provider/resource_rtx_syslog.go (continue)
  - Implement resourceRTXSyslogImport()
  - Accept fixed ID "syslog" for import
  - Validate syslog configuration exists on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_scope.go import pattern_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec rtx-syslog, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXSyslogImport(). Accept import ID as "syslog" (singleton). Call GetSyslogConfig to verify configuration exists. Populate all ResourceData fields from retrieved config. Call Read to ensure state consistency | Restrictions: Handle non-existent config errors gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go import function | _Requirements: Requirement 5 (Import) | Success: terraform import rtx_syslog.main syslog works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_syslog" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-syslog, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_syslog": resourceRTXSyslog() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 12. Create resource unit tests
  - File: internal/provider/resource_rtx_syslog_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-syslog, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_syslog.go. Test schema validation (invalid IP, invalid facility). Test CRUD operations with mocked client. Test import with valid and invalid IDs | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [ ] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_syslog_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test syslog configuration with single host
  - Test syslog configuration with multiple hosts
  - Test local address, facility, and log levels
  - Test syslog import
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_test.go acceptance test patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-syslog, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with single host, multiple hosts, custom ports, all facilities, log level combinations. Test import existing syslog config. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 14. Add example Terraform configurations
  - File: examples/syslog/main.tf (new)
  - Basic syslog configuration example (single host)
  - Full syslog configuration example (multiple hosts, all options)
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-syslog, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: syslog with single host. Full: syslog with multiple hosts, custom ports, local_address, facility, all log levels enabled | Restrictions: Use realistic IP addresses, include comments explaining options | _Leverage: examples/dhcp_scope/ | _Requirements: Requirement 1 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-syslog, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works. Check terraform import functionality. Ensure no regressions in existing resources | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
