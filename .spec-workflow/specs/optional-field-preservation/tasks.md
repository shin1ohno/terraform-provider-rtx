# Tasks Document: Optional Field Preservation

## Phase 1: Core Infrastructure

- [x] 1. Create field tracking helper utilities
  - File: `internal/provider/field_helpers.go`
  - Implement helper functions for extracting values from ResourceData with state fallback
  - Functions: `GetBoolValue`, `GetIntValue`, `GetStringValue`, `GetStringListValue`
  - Add pointer helper functions: `BoolPtr`, `IntPtr`, `StringPtr`
  - Purpose: Provide reusable utilities for all resources to handle optional fields consistently
  - _Leverage: `github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema`_
  - _Requirements: 1, 3_
  - _Prompt: Implement the task for spec optional-field-preservation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in Terraform Provider SDK | Task: Create field helper utilities in internal/provider/field_helpers.go that extract values from schema.ResourceData with automatic state fallback, following the design patterns in design.md | Restrictions: Do not modify existing files, follow project code conventions from memory, use Zerolog for any logging | Success: All helper functions compile, handle both config and state values correctly, include proper documentation comments | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [x] 2. Update AdminUserAttributes to use pointer types
  - File: `internal/client/interfaces.go`
  - Change `Administrator bool` to `Administrator *bool`
  - Change `LoginTimer int` to `LoginTimer *int`
  - Keep slice types as-is (nil vs empty is distinguishable)
  - Purpose: Enable distinguishing between "not set" and "explicitly set to zero/false"
  - _Leverage: Existing `AdminUserAttributes` struct_
  - _Requirements: 1, 2_
  - _Prompt: Implement the task for spec optional-field-preservation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with expertise in type systems | Task: Update AdminUserAttributes struct in internal/client/interfaces.go to use pointer types for bool and int fields as specified in design.md | Restrictions: Only modify the specified struct, maintain JSON tags with omitempty, do not change slice fields | Success: Struct compiles, pointer types allow nil detection, JSON serialization works correctly | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

## Phase 2: Command Builder Updates

- [x] 3. Update BuildUserAttributeCommand for pointer types
  - File: `internal/rtx/parsers/admin.go`
  - Modify function to handle `*bool` and `*int` types
  - Only include attributes in command when pointer is non-nil
  - Purpose: Generate RTX commands that only include explicitly set fields
  - _Leverage: Existing `BuildUserAttributeCommand` function_
  - _Requirements: 1, 3_
  - _Prompt: Implement the task for spec optional-field-preservation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router expertise | Task: Update BuildUserAttributeCommand in internal/rtx/parsers/admin.go to handle pointer types, only including attributes when non-nil as specified in design.md | Restrictions: Maintain existing command format, handle nil gracefully, do not change function signature unnecessarily | Success: Command builder correctly handles nil pointers, generates valid RTX commands, existing tests pass after update | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [x] 4. Update UserAttributes struct in parsers package
  - File: `internal/rtx/parsers/admin.go`
  - Change corresponding fields to pointer types to match client.AdminUserAttributes
  - Update toParserUser/fromParserUser conversion functions in admin_service.go
  - Purpose: Maintain consistency between client and parser types
  - _Leverage: Existing `UserAttributes` struct, `toParserUser`, `fromParserUser` functions_
  - _Requirements: 3_
  - _Prompt: Implement the task for spec optional-field-preservation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Update UserAttributes struct in internal/rtx/parsers/admin.go to use pointer types, and update conversion functions in internal/client/admin_service.go | Restrictions: Maintain consistency between client and parser types, ensure bidirectional conversion works | Success: Both structs use compatible types, conversion functions handle nil correctly | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

## Phase 3: Resource Layer Updates

- [x] 5. Update rtx_admin_user resource schema
  - File: `internal/provider/resource_rtx_admin_user.go` (if exists) or appropriate resource file
  - Add `Computed: true` to optional fields that need preservation
  - Ensure schema allows reading values back from router
  - Purpose: Enable Terraform to track current router values for merge
  - _Leverage: Existing resource schema patterns_
  - _Requirements: 4_
  - _Prompt: Implement the task for spec optional-field-preservation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Update resource schema to add Computed: true to optional fields as specified in design.md, search for admin user resource files first | Restrictions: Only add Computed flag, do not change Required/Optional settings, maintain existing descriptions | Success: Schema allows both user input and computed values, terraform plan shows correct diffs | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [x] 6. Update resource Create function to use field helpers
  - File: Resource file for admin user (find in internal/provider/)
  - Use new field helper functions to build AdminUserAttributes
  - Pass complete attribute set to CreateAdminUser
  - Purpose: Apply new pattern to resource creation
  - _Leverage: `internal/provider/field_helpers.go`, existing Create function_
  - _Requirements: 2, 3_
  - _Prompt: Implement the task for spec optional-field-preservation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Update resource Create function to use field helpers for building AdminUserAttributes, ensuring all fields are properly handled | Restrictions: Follow existing error handling patterns, use Zerolog for logging, do not change function signature | Success: Create function uses field helpers, passes complete attributes to client, handles errors correctly | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [x] 7. Update resource Update function for state-based merge
  - File: Resource file for admin user
  - Use d.Get() to obtain merged values (config or state)
  - Build complete AdminUserAttributes with all fields
  - Send full configuration to router
  - Purpose: Implement the core Read-Merge-Write pattern
  - _Leverage: `internal/provider/field_helpers.go`, existing Update function_
  - _Requirements: 1, 3_
  - _Prompt: Implement the task for spec optional-field-preservation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Update resource Update function to implement Read-Merge-Write pattern as specified in design.md, using d.Get() for merged values | Restrictions: Do not add extra router API calls, use existing state values, maintain error handling | Success: Update function sends complete config to router, unspecified fields preserve current values | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [x] 8. Update resource Read function for complete state
  - File: Resource file for admin user
  - Ensure Read function populates all attribute fields in state
  - Handle nil values from router appropriately
  - Purpose: Ensure state contains current values for merge during Update
  - _Leverage: Existing Read function, parser output_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec optional-field-preservation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Update resource Read function to populate all attribute fields in Terraform state from router values | Restrictions: Handle nil/missing values gracefully, do not modify router, use existing parser | Success: Read populates all fields, state reflects actual router configuration | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

## Phase 4: Testing

- [x] 9. Add unit tests for field helpers
  - File: `internal/provider/field_helpers_test.go`
  - Test each helper function with: value in config, value only in state, zero values
  - Test pointer helper functions
  - Purpose: Ensure field helpers work correctly for all scenarios
  - _Leverage: Standard Go testing, testify if available_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec optional-field-preservation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for field helper functions in internal/provider/field_helpers_test.go | Restrictions: Test all edge cases, use table-driven tests, do not require external dependencies | Success: All helper functions tested, edge cases covered, tests pass | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [x] 10. Add unit tests for command builder changes
  - File: `internal/rtx/parsers/admin_test.go`
  - Test BuildUserAttributeCommand with nil pointers
  - Test with mix of set and unset fields
  - Test complete attribute set
  - Purpose: Verify command generation handles pointer types correctly
  - _Leverage: Existing test patterns in parsers package_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec optional-field-preservation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Add or update tests for BuildUserAttributeCommand to cover pointer type handling | Restrictions: Follow existing test patterns, test nil handling, ensure backward compatibility | Success: All command builder scenarios tested, nil handling verified | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [x] 11. Add integration test for administrator preservation
  - File: `internal/provider/resource_rtx_admin_user_test.go` or appropriate test file
  - Test scenario: Create user with administrator=true, update without specifying administrator
  - Verify administrator privilege is preserved (not reset to false)
  - Purpose: Regression test for the original lockout scenario
  - _Leverage: Existing acceptance test patterns_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec optional-field-preservation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create integration test that verifies administrator privilege is preserved when not specified in update | Restrictions: May require TF_ACC flag, use existing test helpers, clean up test resources | Success: Test verifies preservation behavior, catches regression if pattern breaks | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

## Phase 5: Documentation and Audit

- [x] 12. Audit all resources for optional field issues
  - Files: All files in `internal/provider/resource_*.go`
  - Identify resources with optional fields using same problematic pattern
  - Document which resources need updates
  - Purpose: Ensure complete coverage of the fix across all resources
  - _Leverage: grep/search tools, existing resource files_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec optional-field-preservation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Code Auditor | Task: Search all resource files for optional fields that may have the same preservation issue, document findings | Restrictions: Read-only audit, do not modify files, create comprehensive list | Success: All affected resources identified, documented with specific fields | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts including list of affected resources, then mark as [x] when done_

- [x] 13. Apply pattern to remaining affected resources
  - Files: Resources identified in task 12
  - Apply same field helper and schema pattern to each resource
  - Update corresponding parsers and client code as needed
  - Purpose: Complete the fix across the entire provider
  - _Leverage: Patterns established in tasks 1-8_
  - _Requirements: 3, 5_
  - _Prompt: Implement the task for spec optional-field-preservation, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Apply optional field preservation pattern to all remaining resources identified in the audit | Restrictions: Follow established patterns, one resource at a time, verify each works | Success: All resources updated, consistent pattern across provider | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_
