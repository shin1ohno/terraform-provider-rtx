# Tasks Document: Terraform Provider Testing Patterns

## Phase 1: Test Infrastructure

- [x] 1. Create acctest package foundation
  - File: `internal/provider/acctest/acctest.go`
  - Implement PreCheck function for router connectivity verification
  - Implement RandomName helper for unique test resource names
  - Add provider factory setup for acceptance tests
  - Purpose: Establish core test infrastructure reused by all acceptance tests
  - _Leverage: `terraform-plugin-sdk/v2/helper/acctest`, existing provider setup_
  - _Requirements: 5, 7_
  - _Prompt: Implement the task for spec terraform-provider-testing-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Infrastructure Developer | Task: Create acctest package with PreCheck, RandomName, and provider factory functions as specified in design.md | Restrictions: Do not import production code unnecessarily, follow existing test patterns, ensure PreCheck verifies RTX_HOST and RTX_USER env vars | Success: Package compiles, PreCheck correctly identifies missing prerequisites, RandomName generates unique prefixed names | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [ ] 2. Create reusable check functions
  - File: `internal/provider/acctest/checks.go`
  - Implement CheckResourceAttrNotEmpty for non-empty attribute verification
  - Implement CheckNoPlannedChanges wrapper for SDK v2
  - Implement CheckResourceImportState for import verification
  - Purpose: Reduce test boilerplate with reusable assertion helpers
  - _Leverage: `terraform-plugin-sdk/v2/helper/resource`_
  - _Requirements: 1, 3, 6_
  - _Prompt: Implement the task for spec terraform-provider-testing-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create reusable check functions in acctest/checks.go for common test assertions | Restrictions: Follow terraform-plugin-sdk patterns, return proper TestCheckFunc signatures, handle edge cases gracefully | Success: All check functions compile, work with resource.TestCase, provide clear failure messages | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [ ] 3. Create test step builders
  - File: `internal/provider/acctest/steps.go`
  - Implement BasicCreateStep for standard resource creation
  - Implement UpdateStep for update testing
  - Implement ImportStep for import testing
  - Implement NoChangeStep for perpetual diff verification
  - Purpose: Standardize test step construction across all resource tests
  - _Leverage: `terraform-plugin-sdk/v2/helper/resource`_
  - _Requirements: 1, 3, 5_
  - _Prompt: Implement the task for spec terraform-provider-testing-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create test step builder functions that return configured resource.TestStep structs | Restrictions: Ensure steps are composable, support optional check functions, follow SDK patterns | Success: Step builders create valid TestStep structs, NoChangeStep correctly verifies empty plans | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [x] 4. Create config builder utilities
  - File: `internal/provider/acctest/config.go`
  - Implement ConfigBuilder struct with fluent interface
  - Add methods for setting resource attributes
  - Implement Build() to generate valid HCL
  - Purpose: Generate test configurations programmatically
  - _Leverage: Standard Go string formatting_
  - _Requirements: 5, 7_
  - _Prompt: Implement the task for spec terraform-provider-testing-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Create ConfigBuilder with fluent interface for building HCL test configurations | Restrictions: Generate valid HCL syntax, support all Terraform attribute types, escape special characters properly | Success: ConfigBuilder generates valid HCL, supports nested blocks, handles lists and maps correctly | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

## Phase 2: State Migration Test Support

- [x] 5. Create state migration test helpers
  - File: `internal/provider/acctest/migration.go`
  - Implement StateMigrationTestCase struct
  - Implement RunStateMigrationTests function for unit testing upgraders
  - Add helper for cross-version testing with ExternalProviders
  - Purpose: Enable systematic testing of state version upgrades
  - _Leverage: `terraform-plugin-sdk/v2/helper/schema`, `google/go-cmp`_
  - _Requirements: 8_
  - _Prompt: Implement the task for spec terraform-provider-testing-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create state migration test helpers as specified in design.md section on State Migration Testing | Restrictions: Support multiple schema versions, provide clear diff output on failure, handle nil state gracefully | Success: RunStateMigrationTests correctly validates upgrader functions, failures show clear diffs | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

## Phase 3: Unit Tests for Test Infrastructure

- [ ] 6. Add unit tests for acctest package
  - File: `internal/provider/acctest/acctest_test.go`
  - Test RandomName generates unique names
  - Test ConfigBuilder produces valid HCL
  - Test check functions with mock data
  - Purpose: Ensure test infrastructure itself is reliable
  - _Leverage: `testing`, `testify/assert`_
  - _Requirements: 7_
  - _Prompt: Implement the task for spec terraform-provider-testing-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for the acctest package functions | Restrictions: Use table-driven tests, test edge cases, do not require external resources | Success: All acctest functions have test coverage, tests pass reliably | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

## Phase 4: Reference Implementation Tests

- [ ] 7. Create rtx_admin_user perpetual diff test
  - File: `internal/provider/resource_rtx_admin_user_test.go`
  - Implement TestAccAdminUser_noDiff following Pattern 1 from design
  - Verify no changes on re-apply of same config
  - Purpose: Reference implementation of perpetual diff prevention test
  - _Leverage: `internal/provider/acctest/`, existing test file if any_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec terraform-provider-testing-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create perpetual diff prevention test for rtx_admin_user resource | Restrictions: Use acctest helpers, require TF_ACC flag, clean up test resources | Success: Test passes when resource produces no diff on re-apply, fails if perpetual diff exists | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [ ] 8. Create rtx_admin_user import test
  - File: `internal/provider/resource_rtx_admin_user_test.go`
  - Implement TestAccAdminUser_import following Pattern 2 from design
  - Verify import populates all expected attributes
  - Use ImportStateVerifyIgnore for write-only fields
  - Purpose: Reference implementation of import test
  - _Leverage: `internal/provider/acctest/`_
  - _Requirements: 3_
  - _Prompt: Implement the task for spec terraform-provider-testing-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create import test for rtx_admin_user resource | Restrictions: Use ImportStep helper, ignore password field in verification, clean up resources | Success: Test passes when import correctly populates state, fails if import is broken | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [ ] 9. Create rtx_admin_user Optional+Computed preservation test
  - File: `internal/provider/resource_rtx_admin_user_test.go`
  - Implement TestAccAdminUser_preserveAdministrator following Pattern 4 from design
  - Verify administrator field preserved when not specified in update
  - Purpose: Regression test for administrator lockout scenario
  - _Leverage: `internal/provider/acctest/`_
  - _Requirements: 2_
  - _Prompt: Implement the task for spec terraform-provider-testing-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create Optional+Computed preservation test for administrator field | Restrictions: Test both setting and removing the field from config, verify state values | Success: Test passes when administrator is preserved, fails if reset to default | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [ ] 10. Create WriteOnly attribute test for password
  - File: `internal/provider/resource_rtx_admin_user_test.go`
  - Implement TestAccAdminUser_passwordHandling following Pattern 5 from design
  - Verify password is applied but not readable from state
  - Purpose: Verify sensitive field handling
  - _Leverage: `internal/provider/acctest/`_
  - _Requirements: 9_
  - _Prompt: Implement the task for spec terraform-provider-testing-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create test verifying password field is not stored or is marked sensitive in state | Restrictions: Do not log actual password values, verify behavior not implementation | Success: Test confirms password is handled securely, not exposed in state | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

## Phase 5: Documentation and Expansion

- [ ] 11. Create test pattern documentation
  - File: `docs/testing-patterns.md`
  - Document all test patterns with examples
  - Include decision tree for choosing test types
  - Add troubleshooting guide for common test failures
  - Purpose: Enable developers to write consistent tests
  - _Leverage: Design document examples_
  - _Requirements: 5, 7_
  - _Prompt: Implement the task for spec terraform-provider-testing-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer | Task: Create comprehensive testing patterns documentation | Restrictions: Use clear examples, reference actual code, keep it practical | Success: Documentation helps developers write tests correctly, covers all patterns | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [ ] 12. Add tests to rtx_admin resource
  - File: `internal/provider/resource_rtx_admin_test.go`
  - Apply all test patterns from Phase 4 to rtx_admin
  - Adapt patterns to resource-specific attributes
  - Purpose: Expand test coverage to second resource
  - _Leverage: `internal/provider/acctest/`, rtx_admin_user tests as reference_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec terraform-provider-testing-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Apply all test patterns to rtx_admin resource | Restrictions: Follow established patterns, adapt to resource specifics, maintain consistency | Success: rtx_admin has comprehensive test coverage matching rtx_admin_user | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [ ] 13. Audit and add tests to remaining resources
  - Files: All `internal/provider/resource_*_test.go`
  - Identify resources lacking test coverage
  - Apply test patterns to each resource
  - Purpose: Achieve comprehensive test coverage across provider
  - _Leverage: `internal/provider/acctest/`, existing tests as reference_
  - _Requirements: All_
  - _Prompt: Implement the task for spec terraform-provider-testing-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Audit all resources and add missing test coverage | Restrictions: Prioritize high-risk resources, maintain pattern consistency, document any resource-specific considerations | Success: All resources have at least basic, import, and no-diff tests | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_
