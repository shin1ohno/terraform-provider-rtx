# Tasks Document: rtx_schedule

## Phase 1: Parser Layer

- [x] 1. Create Schedule data model and parser
  - File: internal/rtx/parsers/schedule.go
  - Define Schedule struct with ID, Name, AtTime, DayOfWeek, Date, Recurring, OnStartup, PolicyList, Commands, Enabled fields
  - Implement ParseScheduleConfig() to parse RTX output
  - Purpose: Parse "show config | grep schedule" output
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-schedule, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create Schedule struct with ID (int), Name (string), AtTime (string, HH:MM format), DayOfWeek (string), Date (string, YYYY/MM/DD format), Recurring (bool), OnStartup (bool), PolicyList (string), Commands ([]string), Enabled (bool). Implement ParseScheduleConfig() function to parse RTX router output from "show config | grep schedule" command. Handle "schedule at", "schedule pp", and startup schedules | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line schedule configurations | _Leverage: internal/rtx/parsers/dhcp_scope.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (Time), Requirement 3 (Action), Requirement 4 (Days) | Success: Parser correctly extracts all schedule attributes from sample RTX output, handles edge cases like startup schedules and date-specific tasks | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [x] 2. Create command builder functions for schedule
  - File: internal/rtx/parsers/schedule.go (continue)
  - Implement BuildScheduleAtCommand() for time-based schedules
  - Implement BuildScheduleAtStartupCommand() for startup schedules
  - Implement BuildScheduleAtDateTimeCommand() for date-specific schedules
  - Implement BuildSchedulePPCommand() for PP connection schedules
  - Implement BuildDeleteScheduleCommand() for deletion
  - Implement BuildShowScheduleCommand() for reading
  - Purpose: Generate RTX CLI commands for schedule management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go BuildDHCPScopeCommand pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-schedule, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "schedule at <id> <time> <command>", "schedule at <id> startup <command>", "schedule at <id> <date> <time> <command>", "schedule pp <n> <day> <time> connect/disconnect", "no schedule at <id>". Support day ranges like "mon-fri" | Restrictions: Follow existing BuildDHCPScopeCommand pattern exactly, validate inputs before building commands, validate time format HH:MM and date format YYYY/MM/DD | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax, time and date validation works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 3. Create parser unit tests
  - File: internal/rtx/parsers/schedule_test.go
  - Test ParseScheduleConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: startup schedules, date-specific schedules, PP schedules
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-schedule, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for schedule.go. Include test cases for parsing schedule config output, command building with various parameter combinations, edge cases like recurring vs one-time schedules, startup schedules, day ranges (mon-fri) | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [x] 4. Add Schedule type to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add Schedule struct with all fields
  - Extend Client interface with schedule methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-schedule, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add Schedule struct (ID int, Name string, AtTime string, DayOfWeek string, Date string, Recurring bool, OnStartup bool, PolicyList string, Commands []string, Enabled bool). Extend Client interface with: GetSchedule(ctx, scheduleID) (*Schedule, error), CreateSchedule(ctx, schedule) error, UpdateSchedule(ctx, schedule) error, DeleteSchedule(ctx, scheduleID) error, ListSchedules(ctx) ([]Schedule, error) | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 5. Create ScheduleService implementation
  - File: internal/client/schedule_service.go (new)
  - Implement ScheduleService struct with executor reference
  - Implement Create() with validation and command execution
  - Implement Get() to parse schedule configuration
  - Implement Update() for modifying schedule
  - Implement Delete() to remove schedule
  - Implement List() to retrieve all schedules
  - Purpose: Service layer for schedule CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-schedule, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create ScheduleService following DHCPScopeService pattern. Include input validation (time format HH:MM, date format YYYY/MM/DD, valid day names). Use parsers.BuildScheduleAtCommand and related functions. Call client.SaveConfig() after modifications. Handle schedule update by deleting and recreating | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, separate handling for startup vs time-based schedules | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-4 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 6. Integrate ScheduleService into rtxClient
  - File: internal/client/client.go (modify)
  - Add scheduleService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface schedule methods delegating to service
  - Purpose: Wire up schedule service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-schedule, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add scheduleService *ScheduleService field to rtxClient. Initialize in Dial(): c.scheduleService = NewScheduleService(c.executor, c). Implement GetSchedule, CreateSchedule, UpdateSchedule, DeleteSchedule, ListSchedules methods delegating to service | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all schedule methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 7. Create service unit tests
  - File: internal/client/schedule_service_test.go (new)
  - Test Create with valid and invalid inputs
  - Test Get parsing
  - Test Update behavior
  - Test Delete
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-schedule, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for ScheduleService. Mock Executor interface to simulate RTX responses. Test validation (invalid time format, invalid date, invalid day names). Test successful CRUD operations for time-based, startup, and date-specific schedules. Test error handling | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [x] 8. Create Terraform resource schema for rtx_kron_policy
  - File: internal/provider/resource_rtx_kron_policy.go (new)
  - Define resourceRTXKronPolicy() with full schema
  - Add name (Required, ForceNew, String)
  - Add command_lines (Required, List of String)
  - Purpose: Define Terraform resource structure for policy (command list)
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-schedule, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXKronPolicy() returning *schema.Resource. Define schema with name (Required, ForceNew, String) and command_lines (Required, TypeList of String). Policy represents a named list of commands to be executed by schedules | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 (CRUD) | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 9. Implement CRUD operations for rtx_kron_policy resource
  - File: internal/provider/resource_rtx_kron_policy.go (continue)
  - Implement resourceRTXKronPolicyCreate()
  - Implement resourceRTXKronPolicyRead()
  - Implement resourceRTXKronPolicyUpdate()
  - Implement resourceRTXKronPolicyDelete()
  - Implement resourceRTXKronPolicyImport()
  - Purpose: Terraform lifecycle management for policy
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-schedule, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build policy from ResourceData, store command list, set ID to name). Read (retrieve policy, update ResourceData). Update (modify command_lines). Delete (remove policy). Import (parse name from import ID). Follow rtx_dhcp_scope patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1, 5 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 10. Create Terraform resource schema for rtx_kron_schedule
  - File: internal/provider/resource_rtx_kron_schedule.go (new)
  - Define resourceRTXKronSchedule() with full schema
  - Add id (Required, ForceNew, Int)
  - Add name (Optional, String)
  - Add at_time (Optional, String with HH:MM validation)
  - Add day_of_week (Optional, String)
  - Add date (Optional, String with YYYY/MM/DD validation)
  - Add recurring (Optional, Bool, default false)
  - Add on_startup (Optional, Bool, default false)
  - Add policy_list (Required, String)
  - Purpose: Define Terraform resource structure for schedule
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-schedule, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXKronSchedule() returning *schema.Resource. Define schema with id (Required, ForceNew, Int), name (Optional, String), at_time (Optional, String with HH:MM validation), day_of_week (Optional, String with valid day validation), date (Optional, String with YYYY/MM/DD validation), recurring (Optional, Bool), on_startup (Optional, Bool), policy_list (Required, String). Add ConflictsWith for mutually exclusive options (on_startup vs at_time) | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1-4 | Success: Schema compiles, validation functions work, conflicts properly defined | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 11. Implement CRUD operations for rtx_kron_schedule resource
  - File: internal/provider/resource_rtx_kron_schedule.go (continue)
  - Implement resourceRTXKronScheduleCreate()
  - Implement resourceRTXKronScheduleRead()
  - Implement resourceRTXKronScheduleUpdate()
  - Implement resourceRTXKronScheduleDelete()
  - Implement resourceRTXKronScheduleImport()
  - Purpose: Terraform lifecycle management for schedule
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-schedule, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build Schedule from ResourceData, call client.CreateSchedule, set ID to schedule id). Read (call GetSchedule, update ResourceData, handle not found by clearing ID). Update (call UpdateSchedule for mutable fields). Delete (call DeleteSchedule). Import (parse schedule_id from import ID string). Follow rtx_dhcp_scope patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1, 5 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 12. Register resources in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_kron_policy" to ResourcesMap
  - Add "rtx_kron_schedule" to ResourcesMap
  - Purpose: Make resources available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-schedule, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entries to ResourcesMap in provider.go: "rtx_kron_policy": resourceRTXKronPolicy(), "rtx_kron_schedule": resourceRTXKronSchedule() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resources registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 13. Create resource unit tests
  - File: internal/provider/resource_rtx_kron_policy_test.go (new)
  - File: internal/provider/resource_rtx_kron_schedule_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-schedule, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_kron_policy.go and resource_rtx_kron_schedule.go. Test schema validation (invalid time format, conflicting options). Test CRUD operations with mocked client. Test import with valid and invalid IDs | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [x] 14. Create acceptance tests
  - File: internal/provider/resource_rtx_kron_policy_acc_test.go (new)
  - File: internal/provider/resource_rtx_kron_schedule_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test policy creation with command list
  - Test daily recurring schedule
  - Test startup schedule
  - Test one-time date-specific schedule
  - Test schedule import
  - Test dependency between rtx_kron_schedule and rtx_kron_policy
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_acc_test.go acceptance test patterns_
  - _Requirements: 1, 5, 6_
  - _Prompt: Implement the task for spec rtx-schedule, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with policy creation, daily recurring schedule, startup schedule, one-time schedule, import existing schedule. Test rtx_kron_schedule depends_on rtx_kron_policy. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5, 6 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 15. Add example Terraform configurations
  - File: examples/schedule/main.tf (new)
  - Basic policy and schedule creation example
  - Daily recurring schedule example
  - Weekly schedule with day_of_week example
  - Startup schedule example
  - One-time date-specific schedule example
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1, 6_
  - _Prompt: Implement the task for spec rtx-schedule, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: policy with backup command and daily schedule. Weekly: schedule with day_of_week for weekend maintenance. Startup: schedule that runs on router boot. One-time: date-specific schedule for planned maintenance. Show dependency between rtx_kron_schedule and rtx_kron_policy | Restrictions: Use realistic commands, include comments explaining options | _Leverage: examples/dhcp_scope/ | _Requirements: Requirements 1, 6 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 16. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-schedule, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works for both rtx_kron_policy and rtx_kron_schedule. Check terraform import functionality. Ensure no regressions in existing resources | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
