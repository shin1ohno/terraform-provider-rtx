# Tasks Document: rtx_admin

## Phase 1: Parser Layer

- [ ] 1. Create AdminConfig data model and parser
  - File: internal/rtx/parsers/admin.go
  - Define AdminConfig, UserConfig, and UserAttributes structs
  - Implement ParseAdminConfig() to parse RTX output
  - Purpose: Parse "show config | grep -E (login|administrator|user)" output
  - _Leverage: internal/rtx/parsers/dhcp_bindings.go for patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-admin, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create AdminConfig struct with LoginPassword, AdminPassword, Users fields. Create UserConfig struct with Username, Password, Encrypted, Attributes fields. Create UserAttributes struct with Administrator, Connection, GUIPages, LoginTimer fields. Implement ParseAdminConfig() function to parse RTX router output from "show config | grep -E (login|administrator|user)" command. Follow patterns from dhcp_bindings.go | Restrictions: Do not modify existing parser files, use standard library regexp, handle sensitive password fields (passwords shown as asterisks in show config) | _Leverage: internal/rtx/parsers/dhcp_bindings.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (User Attributes), Requirement 3 (Security) | Success: Parser correctly extracts admin settings and user configurations from sample RTX output, handles encrypted password markers | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 2. Create command builder functions for admin configuration
  - File: internal/rtx/parsers/admin.go (continue)
  - Implement BuildLoginPasswordCommand() for console login password
  - Implement BuildAdminPasswordCommand() for administrator password
  - Implement BuildUserCommand() for user account creation
  - Implement BuildUserAttributeCommand() for user attributes
  - Implement BuildDeleteUserCommand() for user deletion
  - Purpose: Generate RTX CLI commands for admin management
  - _Leverage: internal/rtx/parsers/dhcp_bindings.go BuildDHCPBindCommand pattern_
  - _Requirements: 1, 2_
  - _Prompt: Implement the task for spec rtx-admin, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "login password <password>", "administrator password <password>", "login user <username> <password>", "login user <username> encrypted <encrypted_password>", "user attribute <username> administrator=on|off connection=<types> gui-page=<pages> login-timer=<seconds>", "no login user <username>", "no user attribute <username>". Handle connection types (serial, telnet, remote, ssh, sftp, http) and GUI pages (dashboard, lan-map, config) | Restrictions: Follow existing BuildDHCPBindCommand pattern exactly, validate inputs before building commands, handle sensitive password data | _Leverage: internal/rtx/parsers/dhcp_bindings.go | _Requirements: Requirement 1 (CRUD Operations), Requirement 2 (User Attributes) | Success: All commands generate valid RTX CLI syntax | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 3. Create parser unit tests
  - File: internal/rtx/parsers/admin_test.go
  - Test ParseAdminConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: missing fields, encrypted passwords, empty user list
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_bindings_test.go for test patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-admin, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for admin.go. Include test cases for parsing admin config output, command building with various parameter combinations, edge cases like empty user attributes, all connection type combinations, all GUI page combinations | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_bindings_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [ ] 4. Add Admin types to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add AdminConfig, UserConfig, UserAttributes structs
  - Extend Client interface with admin methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-admin, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add AdminConfig struct (LoginPassword, AdminPassword string, Users []UserConfig). Add UserConfig struct (Username, Password string, Encrypted bool, Attributes UserAttributes). Add UserAttributes struct (Administrator bool, Connection []string, GUIPages []string, LoginTimer int). Extend Client interface with: GetAdminConfig(ctx) (*AdminConfig, error), ConfigureAdmin(ctx, config) error, UpdateAdminConfig(ctx, config) error, CreateUser(ctx, user) error, DeleteUser(ctx, username) error, ListUsers(ctx) ([]UserConfig, error) | Restrictions: Do not break existing interface methods, maintain backward compatibility, mark password fields as sensitive | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 5. Create AdminService implementation
  - File: internal/client/admin_service.go (new)
  - Implement AdminService struct with executor reference
  - Implement Configure() for setting passwords
  - Implement Get() to parse admin configuration
  - Implement Update() for modifying admin settings
  - Implement CreateUser() for user account creation
  - Implement DeleteUser() for user removal
  - Implement ListUsers() to retrieve all users
  - Purpose: Service layer for admin CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-admin, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create AdminService following DHCPScopeService pattern. Include input validation (username format, valid connection types, valid GUI pages, login timer range). Use parsers.BuildLoginPasswordCommand, BuildAdminPasswordCommand, BuildUserCommand, BuildUserAttributeCommand, BuildDeleteUserCommand functions. Call client.SaveConfig() after modifications. Handle sensitive password data appropriately | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, do not log passwords | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-3 | Success: All CRUD operations work, validation catches invalid input, configuration is saved, passwords are handled securely | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 6. Integrate AdminService into rtxClient
  - File: internal/client/client.go (modify)
  - Add adminService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface admin methods delegating to service
  - Purpose: Wire up admin service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-admin, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add adminService *AdminService field to rtxClient. Initialize in Dial(): c.adminService = NewAdminService(c.executor, c). Implement GetAdminConfig, ConfigureAdmin, UpdateAdminConfig, CreateUser, DeleteUser, ListUsers methods delegating to service | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all admin methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 7. Create service unit tests
  - File: internal/client/admin_service_test.go (new)
  - Test Configure with valid and invalid inputs
  - Test Get parsing
  - Test CreateUser with various user configurations
  - Test DeleteUser
  - Test validation for connection types and GUI pages
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-admin, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for AdminService. Mock Executor interface to simulate RTX responses. Test validation (invalid username, invalid connection types, invalid GUI pages, out of range login timer). Test successful CRUD operations for both passwords and users. Test error handling | Restrictions: Use mock executor, do not require real router, use table-driven tests, do not include actual passwords in test output | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-3 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [ ] 8. Create Terraform resource schema for rtx_admin
  - File: internal/provider/resource_rtx_admin.go (new)
  - Define resourceRTXAdmin() with schema for passwords
  - Add login_password (Optional, Sensitive, String)
  - Add admin_password (Optional, Sensitive, String)
  - Purpose: Define Terraform resource structure for admin passwords (singleton)
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 3_
  - _Prompt: Implement the task for spec rtx-admin, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXAdmin() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Add login_password (Optional, Sensitive: true). Add admin_password (Optional, Sensitive: true). This is a singleton resource with fixed ID "admin" | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style, mark all password fields as Sensitive: true | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1, 3 | Success: Schema compiles, sensitive fields properly marked | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 9. Create Terraform resource schema for rtx_admin_user
  - File: internal/provider/resource_rtx_admin_user.go (new)
  - Define resourceRTXAdminUser() with schema for user accounts
  - Add username (Required, ForceNew, String)
  - Add password (Required, Sensitive, String)
  - Add encrypted (Optional, Bool, default false)
  - Add attributes block (Optional, nested)
    - administrator (Optional, Bool, default false)
    - connection (Optional, List of String)
    - gui_pages (Optional, List of String)
    - login_timer (Optional, Int)
  - Purpose: Define Terraform resource structure for user accounts
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-admin, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXAdminUser() returning *schema.Resource. Define schema with username (Required, ForceNew), password (Required, Sensitive), encrypted (Optional, Bool). Add attributes block with administrator (Optional, Bool), connection (Optional, List of String with ValidateFunc for valid types), gui_pages (Optional, List of String with ValidateFunc), login_timer (Optional, Int). Valid connection types: serial, telnet, remote, ssh, sftp, http. Valid GUI pages: dashboard, lan-map, config | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style, validate connection types and GUI pages | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1, 2, 3 | Success: Schema compiles, validation functions work, nested attributes defined correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 10. Implement CRUD operations for rtx_admin resource
  - File: internal/provider/resource_rtx_admin.go (continue)
  - Implement resourceRTXAdminCreate()
  - Implement resourceRTXAdminRead()
  - Implement resourceRTXAdminUpdate()
  - Implement resourceRTXAdminDelete()
  - Purpose: Terraform lifecycle management for admin passwords
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1, 3_
  - _Prompt: Implement the task for spec rtx-admin, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build AdminConfig from ResourceData, call client.ConfigureAdmin, set ID to "admin"). Read (call GetAdminConfig, note passwords cannot be read back from router). Update (call UpdateAdminConfig). Delete (reset passwords or no-op based on design decision). Handle sensitive data - passwords won't be readable from router | Restrictions: Use diag.Diagnostics for errors, handle sensitive data appropriately, passwords cannot be read back from device | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1, 3 | Success: All CRUD operations work, sensitive data handled appropriately | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 11. Implement CRUD operations for rtx_admin_user resource
  - File: internal/provider/resource_rtx_admin_user.go (continue)
  - Implement resourceRTXAdminUserCreate()
  - Implement resourceRTXAdminUserRead()
  - Implement resourceRTXAdminUserUpdate()
  - Implement resourceRTXAdminUserDelete()
  - Purpose: Terraform lifecycle management for user accounts
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1, 2_
  - _Prompt: Implement the task for spec rtx-admin, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build UserConfig from ResourceData including attributes, call client.CreateUser, set ID to username). Read (call ListUsers, find user by username, update ResourceData, handle not found by clearing ID). Update (handle attribute changes, password changes may require delete/recreate). Delete (call DeleteUser). Follow rtx_dhcp_scope patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully, password changes may require special handling | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1, 2 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 12. Implement import functionality for both resources
  - File: internal/provider/resource_rtx_admin.go (continue)
  - File: internal/provider/resource_rtx_admin_user.go (continue)
  - Implement resourceRTXAdminImport() (singleton, always "admin")
  - Implement resourceRTXAdminUserImport() (parse username from import ID)
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_scope.go import pattern_
  - _Requirements: 4_
  - _Prompt: Implement the task for spec rtx-admin, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXAdminImport() for singleton resource (ID is always "admin"). Implement resourceRTXAdminUserImport() parsing username from import ID string. Verify resources exist on router. Note: passwords cannot be imported as they're not readable from device | Restrictions: Handle import limitations gracefully (passwords not readable), non-existent user errors | _Leverage: internal/provider/resource_rtx_dhcp_scope.go import function | _Requirements: Requirement 4 (Import) | Success: terraform import works for both resources with appropriate limitations noted | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 13. Register resources in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_admin" to ResourcesMap
  - Add "rtx_admin_user" to ResourcesMap
  - Purpose: Make resources available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-admin, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entries to ResourcesMap in provider.go: "rtx_admin": resourceRTXAdmin(), "rtx_admin_user": resourceRTXAdminUser() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resources registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 14. Create resource unit tests
  - File: internal/provider/resource_rtx_admin_test.go (new)
  - File: internal/provider/resource_rtx_admin_user_test.go (new)
  - Test schema validation for both resources
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 2, 4_
  - _Prompt: Implement the task for spec rtx-admin, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_admin.go and resource_rtx_admin_user.go. Test schema validation (invalid connection types, invalid GUI pages). Test CRUD operations with mocked client. Test import with valid and invalid IDs. Test sensitive field handling | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client, do not include actual passwords in test assertions | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 2, 4 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [ ] 15. Create acceptance tests
  - File: internal/provider/resource_rtx_admin_acc_test.go (new)
  - File: internal/provider/resource_rtx_admin_user_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test password configuration
  - Test user account creation with attributes
  - Test user attribute updates
  - Test user deletion
  - Test import functionality
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_acc_test.go acceptance test patterns_
  - _Requirements: 1, 2, 4_
  - _Prompt: Implement the task for spec rtx-admin, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with admin password setup, user creation with all attribute options, user attribute updates, import existing user. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set, be careful with password testing to not expose credentials | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 2, 4 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 16. Add example Terraform configurations
  - File: examples/admin/main.tf (new)
  - Basic admin password configuration example
  - User account creation example
  - User with full attributes example
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1, 2_
  - _Prompt: Implement the task for spec rtx-admin, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: rtx_admin with login_password and admin_password (using variables, not hardcoded). User: rtx_admin_user with username and password. Full: rtx_admin_user with all attribute options (administrator, connection types, gui_pages, login_timer) | Restrictions: Use variable references for passwords, never hardcode credentials, include comments explaining security considerations | _Leverage: examples/dhcp_scope/ | _Requirements: Requirements 1, 2 | Success: Examples are valid Terraform, demonstrate all features, follow security best practices | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 17. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-admin, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works. Check terraform import functionality. Ensure no regressions in existing resources. Pay special attention to sensitive data handling | Restrictions: All tests must pass, no compiler warnings, sensitive data must not appear in logs | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end, security requirements met | After completing, use log-implementation tool to record details, then mark as [x] complete_
