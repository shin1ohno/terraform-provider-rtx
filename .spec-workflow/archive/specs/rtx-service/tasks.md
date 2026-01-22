# Tasks Document: rtx_service

## Phase 1: Parser Layer

- [x] 1. Create Service data models and parser
  - File: internal/rtx/parsers/service.go
  - Define HTTPDConfig, SSHDConfig, SFTPDConfig structs
  - Implement ParseHTTPDConfig() to parse RTX output
  - Implement ParseSSHDConfig() to parse RTX output
  - Implement ParseSFTPDConfig() to parse RTX output
  - Purpose: Parse "show config | grep httpd/sshd/sftpd" output
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create HTTPDConfig struct with Host (string) and ProxyAccess (bool) fields. Create SSHDConfig struct with Enabled (bool), Hosts ([]string), HostKey (string) fields. Create SFTPDConfig struct with Hosts ([]string) field. Implement ParseHTTPDConfig(), ParseSSHDConfig(), ParseSFTPDConfig() functions to parse RTX router output from "show config" command. Follow patterns from dhcp_scope.go | Restrictions: Do not modify existing parser files, use standard library regexp, handle various output formats | _Leverage: internal/rtx/parsers/dhcp_scope.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (HTTPD), Requirement 2 (SSHD), Requirement 3 (SFTPD) | Success: Parsers correctly extract all service attributes from sample RTX output, handles edge cases like missing optional fields | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [x] 2. Create command builder functions for services
  - File: internal/rtx/parsers/service.go (continue)
  - Implement BuildHTTPDHostCommand() for HTTPD host configuration
  - Implement BuildHTTPDProxyAccessCommand() for proxy access setting
  - Implement BuildSSHDServiceCommand() for SSHD enable/disable
  - Implement BuildSSHDHostCommand() for SSHD host interfaces
  - Implement BuildSFTPDHostCommand() for SFTPD host interfaces
  - Implement delete command builders for each service
  - Purpose: Generate RTX CLI commands for service management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go BuildDHCPScopeCommand pattern_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "httpd host any|<interface>", "httpd proxy-access l2ms permit on|off", "sshd service on|off", "sshd host <interface1> [<interface2> ...]", "sftpd host <interface1> [<interface2> ...]". Also create delete commands: "no httpd host", "no httpd proxy-access", "no sshd service", "no sshd host", "no sftpd host" | Restrictions: Follow existing BuildDHCPScopeCommand pattern exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirements 1, 2, 3 | Success: All commands generate valid RTX CLI syntax | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 3. Create parser unit tests
  - File: internal/rtx/parsers/service_test.go
  - Test ParseHTTPDConfig, ParseSSHDConfig, ParseSFTPDConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: missing fields, malformed input, empty hosts
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for service.go. Include test cases for parsing each service config output, command building with various parameter combinations, edge cases like empty host list, service disabled state | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [x] 4. Add Service types to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add HTTPDConfig, SSHDConfig, SFTPDConfig structs
  - Extend Client interface with service methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add HTTPDConfig struct (Host string, ProxyAccess bool). Add SSHDConfig struct (Enabled bool, Hosts []string, HostKey string). Add SFTPDConfig struct (Hosts []string). Extend Client interface with: GetHTTPDConfig(ctx) (*HTTPDConfig, error), ConfigureHTTPD(ctx, config) error, GetSSHDConfig(ctx) (*SSHDConfig, error), ConfigureSSHD(ctx, config) error, GetSFTPDConfig(ctx) (*SFTPDConfig, error), ConfigureSFTPD(ctx, config) error | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirements 1, 2, 3 | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 5. Create ServiceManager implementation
  - File: internal/client/service_manager.go (new)
  - Implement ServiceManager struct with executor reference
  - Implement ConfigureHTTPD() and GetHTTPD() methods
  - Implement ConfigureSSHD() and GetSSHD() methods
  - Implement ConfigureSFTPD() and GetSFTPD() methods
  - Purpose: Service layer for service CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create ServiceManager following DHCPScopeService pattern. Include input validation (valid interface names, "any" keyword for HTTPD). Use parsers.BuildHTTPDHostCommand and related functions. Call client.SaveConfig() after modifications. Handle singleton resource pattern (each service has only one configuration) | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, maintain separation of concerns | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1, 2, 3 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 6. Integrate ServiceManager into rtxClient
  - File: internal/client/client.go (modify)
  - Add serviceManager field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface service methods delegating to manager
  - Purpose: Wire up service manager to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add serviceManager *ServiceManager field to rtxClient. Initialize in Dial(): c.serviceManager = NewServiceManager(c.executor, c). Implement GetHTTPDConfig, ConfigureHTTPD, GetSSHDConfig, ConfigureSSHD, GetSFTPDConfig, ConfigureSFTPD methods delegating to manager | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirements 1, 2, 3 | Success: Client compiles, all service methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 7. Create service manager unit tests
  - File: internal/client/service_manager_test.go (new)
  - Test ConfigureHTTPD and GetHTTPD with valid and invalid inputs
  - Test ConfigureSSHD and GetSSHD
  - Test ConfigureSFTPD and GetSFTPD
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for ServiceManager. Mock Executor interface to simulate RTX responses. Test validation (invalid interface names). Test successful CRUD operations for all three services. Test error handling | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1, 2, 3 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [x] 8. Create HTTPD Terraform resource schema
  - File: internal/provider/resource_rtx_httpd.go (new)
  - Define resourceRTXHTTPD() with full schema
  - Add host (Required, String, "any" or interface name)
  - Add proxy_access (Optional, Bool, default false)
  - Purpose: Define Terraform resource structure for HTTPD
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXHTTPD() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Add ValidateFunc for host (must be "any" or valid interface name). Set ID to "httpd" (singleton resource) | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 9. Create SSHD Terraform resource schema
  - File: internal/provider/resource_rtx_sshd.go (new)
  - Define resourceRTXSSHD() with full schema
  - Add enabled (Required, Bool)
  - Add hosts (Optional, List of String)
  - Add host_key (Optional, Sensitive, String)
  - Purpose: Define Terraform resource structure for SSHD
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 2_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXSSHD() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Mark host_key as Sensitive. Set ID to "sshd" (singleton resource) | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 2 | Success: Schema compiles, host_key is properly marked sensitive | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 10. Create SFTPD Terraform resource schema
  - File: internal/provider/resource_rtx_sftpd.go (new)
  - Define resourceRTXSFTPD() with full schema
  - Add hosts (Required, List of String)
  - Purpose: Define Terraform resource structure for SFTPD
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 3_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXSFTPD() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Set ID to "sftpd" (singleton resource) | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 3 | Success: Schema compiles | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 11. Implement CRUD operations for HTTPD resource
  - File: internal/provider/resource_rtx_httpd.go (continue)
  - Implement resourceRTXHTTPDCreate()
  - Implement resourceRTXHTTPDRead()
  - Implement resourceRTXHTTPDUpdate()
  - Implement resourceRTXHTTPDDelete()
  - Implement resourceRTXHTTPDImport()
  - Purpose: Terraform lifecycle management for HTTPD
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build HTTPDConfig from ResourceData, call client.ConfigureHTTPD, set ID to "httpd"). Read (call GetHTTPDConfig, update ResourceData). Update (call ConfigureHTTPD with updated values). Delete (disable HTTPD configuration). Import (set ID to "httpd", call Read) | Restrictions: Use diag.Diagnostics for errors, handle singleton resource pattern | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 12. Implement CRUD operations for SSHD resource
  - File: internal/provider/resource_rtx_sshd.go (continue)
  - Implement resourceRTXSSHDCreate()
  - Implement resourceRTXSSHDRead()
  - Implement resourceRTXSSHDUpdate()
  - Implement resourceRTXSSHDDelete()
  - Implement resourceRTXSSHDImport()
  - Purpose: Terraform lifecycle management for SSHD
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 2_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build SSHDConfig from ResourceData, call client.ConfigureSSHD, set ID to "sshd"). Read (call GetSSHDConfig, update ResourceData). Update (call ConfigureSSHD with updated values). Delete (disable SSHD service). Import (set ID to "sshd", call Read). Handle host_key sensitivity | Restrictions: Use diag.Diagnostics for errors, warn about self-lockout when disabling | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 2 | Success: All CRUD operations work, sensitive fields handled correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 13. Implement CRUD operations for SFTPD resource
  - File: internal/provider/resource_rtx_sftpd.go (continue)
  - Implement resourceRTXSFTPDCreate()
  - Implement resourceRTXSFTPDRead()
  - Implement resourceRTXSFTPDUpdate()
  - Implement resourceRTXSFTPDDelete()
  - Implement resourceRTXSFTPDImport()
  - Purpose: Terraform lifecycle management for SFTPD
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 3_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build SFTPDConfig from ResourceData, call client.ConfigureSFTPD, set ID to "sftpd"). Read (call GetSFTPDConfig, update ResourceData). Update (call ConfigureSFTPD with updated values). Delete (remove SFTPD configuration). Import (set ID to "sftpd", call Read) | Restrictions: Use diag.Diagnostics for errors, handle singleton resource pattern | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 3 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 14. Register resources in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_httpd" to ResourcesMap
  - Add "rtx_sshd" to ResourcesMap
  - Add "rtx_sftpd" to ResourcesMap
  - Purpose: Make resources available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entries to ResourcesMap in provider.go: "rtx_httpd": resourceRTXHTTPD(), "rtx_sshd": resourceRTXSSHD(), "rtx_sftpd": resourceRTXSFTPD() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirements 1, 2, 3 | Success: Provider compiles with new resources registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 15. Create resource unit tests
  - File: internal/provider/resource_rtx_httpd_test.go (new)
  - File: internal/provider/resource_rtx_sshd_test.go (new)
  - File: internal/provider/resource_rtx_sftpd_test.go (new)
  - Test schema validation for each resource
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for all three service resources. Test schema validation (invalid host values). Test CRUD operations with mocked client. Test import with valid IDs. Test sensitive field handling for SSHD host_key | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 2, 3 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [x] 16. Create acceptance tests
  - File: internal/provider/resource_rtx_httpd_acc_test.go (new)
  - File: internal/provider/resource_rtx_sshd_acc_test.go (new)
  - File: internal/provider/resource_rtx_sftpd_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test service configuration with various parameters
  - Test service update
  - Test service import
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_acc_test.go acceptance test patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with HTTPD, SSHD, SFTPD creation and update. Test import existing configurations. Use TF_ACC environment check. Be careful with SSHD tests to avoid self-lockout | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 2, 3 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 17. Add example Terraform configurations
  - File: examples/services/httpd/main.tf (new)
  - File: examples/services/sshd/main.tf (new)
  - File: examples/services/sftpd/main.tf (new)
  - Basic service configuration examples
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. HTTPD: host with "any" and specific interface, proxy_access enabled. SSHD: enabled with multiple host interfaces. SFTPD: with host interface list. Include security recommendations in comments | Restrictions: Use realistic interface names, include comments explaining options, add security warnings | _Leverage: examples/dhcp_scope/ | _Requirements: Requirements 1, 2, 3 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 18. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-service, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works for all three services. Check terraform import functionality. Ensure no regressions in existing resources. Verify security warnings are displayed when appropriate | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
