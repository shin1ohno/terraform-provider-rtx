# Tasks Document: rtx_system

## Phase 1: Parser Layer

- [ ] 1. Create SystemConfig data model and parser
  - File: internal/rtx/parsers/system.go
  - Define SystemConfig, ConsoleConfig, PacketBufferConfig, StatisticsConfig structs
  - Implement ParseSystemConfig() to parse RTX output
  - Purpose: Parse "show config | grep -E (timezone|console|packet-buffer|statistics)" output
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-system, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create SystemConfig struct with Timezone, Console, PacketBuffers, Statistics fields. Create nested structs ConsoleConfig (Character, Lines, Prompt), PacketBufferConfig (Size, MaxBuffer, MaxFree), StatisticsConfig (Traffic, NAT). Implement ParseSystemConfig() function to parse RTX router output. Follow patterns from dhcp_scope.go | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line configurations | _Leverage: internal/rtx/parsers/dhcp_scope.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (Timezone), Requirement 3 (Console), Requirement 4 (Statistics) | Success: Parser correctly extracts all system attributes from sample RTX output, handles edge cases like missing optional fields | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 2. Create command builder functions for system configuration
  - File: internal/rtx/parsers/system.go (continue)
  - Implement BuildTimezoneCommand() for timezone setting
  - Implement BuildConsoleCommand() for console settings (character, lines, prompt)
  - Implement BuildPacketBufferCommand() for packet buffer tuning
  - Implement BuildStatisticsCommand() for statistics collection settings
  - Implement BuildDeleteSystemCommands() for removing configuration
  - Implement BuildShowSystemConfigCommand() for reading configuration
  - Purpose: Generate RTX CLI commands for system management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go BuildDHCPScopeCommand pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-system, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "timezone <offset>", "console character <encoding>", "console lines <number|infinity>", "console prompt <prompt>", "system packet-buffer <size> max-buffer=<n> max-free=<n>", "statistics traffic on|off", "statistics nat on|off". Handle prompt quoting for strings with spaces | Restrictions: Follow existing BuildDHCPScopeCommand pattern exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 3. Create parser unit tests
  - File: internal/rtx/parsers/system_test.go
  - Test ParseSystemConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: missing fields, malformed input, infinity lines
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-system, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for system.go. Include test cases for parsing system config output, command building with various parameter combinations, edge cases like missing console settings, infinity lines, various timezone formats | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [ ] 4. Add SystemConfig type to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add SystemConfig struct with all fields
  - Add ConsoleConfig, PacketBufferConfig, StatisticsConfig structs
  - Extend Client interface with system methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-system, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add SystemConfig struct (Timezone string, Console *ConsoleConfig, PacketBuffers []PacketBufferConfig, Statistics *StatisticsConfig). Add nested config structs. Extend Client interface with: GetSystemConfig(ctx) (*SystemConfig, error), ConfigureSystem(ctx, config) error, UpdateSystemConfig(ctx, config) error, ResetSystem(ctx) error | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 5. Create SystemService implementation
  - File: internal/client/system_service.go (new)
  - Implement SystemService struct with executor reference
  - Implement Configure() with validation and command execution
  - Implement Get() to parse system configuration
  - Implement Update() for modifying settings
  - Implement Reset() for removing custom configuration
  - Purpose: Service layer for system CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-system, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create SystemService following DHCPScopeService pattern. Include input validation (timezone format +/-HH:MM, valid character encodings, valid packet buffer sizes). Use parsers.BuildTimezoneCommand and related functions. Call client.SaveConfig() after modifications. Handle partial updates for individual settings | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, singleton resource pattern (fixed ID "system") | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-4 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 6. Integrate SystemService into rtxClient
  - File: internal/client/client.go (modify)
  - Add systemService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface system methods delegating to service
  - Purpose: Wire up system service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-system, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add systemService *SystemService field to rtxClient. Initialize in Dial(): c.systemService = NewSystemService(c.executor, c). Implement GetSystemConfig, ConfigureSystem, UpdateSystemConfig, ResetSystem methods delegating to service | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all system methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 7. Create service unit tests
  - File: internal/client/system_service_test.go (new)
  - Test Configure with valid and invalid inputs
  - Test Get parsing
  - Test Update behavior
  - Test Reset
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-system, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for SystemService. Mock Executor interface to simulate RTX responses. Test validation (invalid timezone format, invalid character encoding, invalid packet buffer size). Test successful CRUD operations. Test error handling | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [ ] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_system.go (new)
  - Define resourceRTXSystem() with full schema
  - Add timezone (Optional, String with UTC offset validation)
  - Add console block (Optional, nested: character, lines, prompt)
  - Add packet_buffer block (Optional, List: size, max_buffer, max_free)
  - Add statistics block (Optional, nested: traffic, nat booleans)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-system, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXSystem() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Add ValidateFunc for timezone (UTC offset format). Use TypeList for console block with nested schema (character, lines, prompt). Use TypeList for packet_buffer with nested schema (size enum, max_buffer int, max_free int). Use TypeList for statistics block (traffic bool, nat bool). Singleton resource with fixed ID "system" | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1-4 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_system.go (continue)
  - Implement resourceRTXSystemCreate()
  - Implement resourceRTXSystemRead()
  - Implement resourceRTXSystemUpdate()
  - Implement resourceRTXSystemDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-system, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build SystemConfig from ResourceData, call client.ConfigureSystem, set ID to "system"). Read (call GetSystemConfig, update ResourceData, handle not found gracefully). Update (call UpdateSystemConfig for modified fields). Delete (call ResetSystem to restore defaults). Follow rtx_dhcp_scope patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully, singleton resource pattern | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 10. Implement import functionality
  - File: internal/provider/resource_rtx_system.go (continue)
  - Implement resourceRTXSystemImport()
  - Accept "system" as import ID (singleton)
  - Validate system config exists on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_scope.go import pattern_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec rtx-system, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXSystemImport(). Accept import ID "system" (singleton resource). Call GetSystemConfig to retrieve existing configuration. Populate all ResourceData fields from retrieved config. Call Read to ensure state consistency | Restrictions: Handle empty/default configuration gracefully, only accept "system" as valid import ID | _Leverage: internal/provider/resource_rtx_dhcp_scope.go import function | _Requirements: Requirement 5 (Import) | Success: terraform import rtx_system.main system works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_system" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-system, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_system": resourceRTXSystem() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 12. Create resource unit tests
  - File: internal/provider/resource_rtx_system_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-system, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_system.go. Test schema validation (invalid timezone format, invalid packet buffer size). Test CRUD operations with mocked client. Test import with valid ID "system" and invalid IDs | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [ ] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_system_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test system configuration with all parameters
  - Test system update
  - Test system import
  - Test singleton behavior
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_acc_test.go acceptance test patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-system, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with timezone setting, console configuration, packet buffer tuning, statistics settings. Test updating individual settings. Test import existing system configuration. Verify singleton behavior. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 14. Add example Terraform configurations
  - File: examples/system/main.tf (new)
  - Basic timezone configuration example
  - Console settings example
  - Packet buffer tuning example
  - Full configuration example with all options
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-system, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: timezone only. Console: character encoding and prompt. Performance: packet buffer tuning for high-traffic. Full: all options including statistics. Include comments explaining each option and its impact | Restrictions: Use realistic values, include comments explaining options | _Leverage: examples/dhcp_scope/ | _Requirements: Requirement 1 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-system, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works. Check terraform import functionality. Ensure no regressions in existing resources. Verify singleton resource behavior | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
