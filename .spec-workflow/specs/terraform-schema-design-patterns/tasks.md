# Tasks Document: Terraform Schema Design Patterns

## Phase 1: Core Utilities

- [x] 1. Create schema helpers module
  - File: `internal/provider/schema_helpers.go`
  - Implement WriteOnlyStringSchema function for SDK v2
  - Implement SensitiveStringSchema function
  - Add documentation comments explaining SDK v2 limitations
  - Purpose: Provide reusable schema patterns for sensitive fields
  - _Leverage: `terraform-plugin-sdk/v2/helper/schema`_
  - _Requirements: 2_
  - _Prompt: Implement the task for spec terraform-schema-design-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create schema helper functions for sensitive and write-only fields as specified in design.md | Restrictions: Follow SDK v2 patterns, document limitations vs plugin-framework, return *schema.Schema | Success: Helper functions compile and produce valid schema definitions | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [x] 2. Create DiffSuppressFunc library
  - File: `internal/provider/diff_suppress.go`
  - Implement SuppressCaseDiff for case-insensitive comparison
  - Implement SuppressWhitespaceDiff for whitespace normalization
  - Implement SuppressJSONDiff for semantic JSON comparison
  - Implement SuppressEquivalentIPDiff for IP address comparison
  - Purpose: Provide reusable diff suppression functions
  - _Leverage: `encoding/json`, `net`, `strings`_
  - _Requirements: 4_
  - _Prompt: Implement the task for spec terraform-schema-design-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Create DiffSuppressFunc library with all functions specified in design.md | Restrictions: Handle edge cases (nil, empty, invalid input), return false on errors (safe fallback), add documentation | Success: All suppress functions compile and correctly identify equivalent values | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [x] 3. Create StateFunc normalizers
  - File: `internal/provider/state_funcs.go`
  - Implement normalizeIPAddress for canonical IP format
  - Implement normalizeLowercase for case normalization
  - Implement normalizeJSON for consistent JSON formatting
  - Purpose: Normalize values before storing in state to prevent diffs
  - _Leverage: `encoding/json`, `net`, `strings`_
  - _Requirements: 4, 5_
  - _Prompt: Implement the task for spec terraform-schema-design-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Create StateFunc normalizer functions as specified in design.md | Restrictions: Handle invalid input gracefully, return input unchanged if normalization fails, document behavior | Success: Normalizers produce consistent canonical output for equivalent inputs | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

## Phase 2: State Migration Infrastructure

- [x] 4. Create state upgraders directory structure
  - File: `internal/provider/state_upgraders/upgraders.go`
  - Create package with common types and utilities
  - Implement helper for creating StateUpgrader entries
  - Add documentation on state migration patterns
  - Purpose: Establish structure for state version management
  - _Leverage: `terraform-plugin-sdk/v2/helper/schema`_
  - _Requirements: 9_
  - _Prompt: Implement the task for spec terraform-schema-design-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create state_upgraders package with common utilities | Restrictions: Follow SDK v2 StateUpgrader patterns, document version numbering convention | Success: Package provides reusable utilities for state migration | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

## Phase 3: Unit Tests

- [x] 5. Add unit tests for DiffSuppressFunc library
  - File: `internal/provider/diff_suppress_test.go`
  - Test each suppress function with equivalent and different values
  - Test edge cases: nil, empty, invalid format
  - Use table-driven tests
  - Purpose: Ensure diff suppression works correctly
  - _Leverage: `testing`, `testify/assert`_
  - _Requirements: 4_
  - _Prompt: Implement the task for spec terraform-schema-design-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for DiffSuppressFunc library | Restrictions: Use table-driven tests, cover edge cases, test both positive and negative cases | Success: All suppress functions have full test coverage, tests document expected behavior | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [-] 6. Add unit tests for StateFunc normalizers
  - File: `internal/provider/state_funcs_test.go`
  - Test each normalizer with various input formats
  - Test invalid input handling
  - Verify idempotency (normalizing twice produces same result)
  - Purpose: Ensure normalizers work correctly
  - _Leverage: `testing`, `testify/assert`_
  - _Requirements: 4, 5_
  - _Prompt: Implement the task for spec terraform-schema-design-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for StateFunc normalizers | Restrictions: Test idempotency, cover edge cases, verify invalid input handling | Success: All normalizers tested, idempotency verified, edge cases covered | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

## Phase 4: Reference Implementation

- [x] 7. Apply WriteOnly pattern to password fields
  - File: `internal/provider/resource_rtx_admin_user.go`
  - Update password field to use WriteOnlyStringSchema helper
  - Ensure Read function does not attempt to read password from router
  - Verify password is sent during Create/Update but not stored in state
  - Purpose: Reference implementation of WriteOnly pattern
  - _Leverage: `internal/provider/schema_helpers.go`_
  - _Requirements: 2_
  - _Prompt: Implement the task for spec terraform-schema-design-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Apply WriteOnly pattern to password field in rtx_admin_user resource | Restrictions: Do not break existing functionality, ensure password is still sent to router, verify with tests | Success: Password is applied to router but not stored in state | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [x] 8. Apply RequiresReplace to username field
  - File: `internal/provider/resource_rtx_admin_user.go`
  - Add ForceNew: true to username field if not already present
  - Update schema description to indicate immutability
  - Verify terraform plan shows replacement when username changes
  - Purpose: Reference implementation of RequiresReplace pattern
  - _Leverage: Existing schema definition_
  - _Requirements: 3_
  - _Prompt: Implement the task for spec terraform-schema-design-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Apply RequiresReplace pattern to username field | Restrictions: Only if username is truly immutable on RTX router, update description, verify with test | Success: Changing username triggers resource replacement in plan | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [x] 9. Apply DiffSuppressFunc to appropriate fields
  - Files: Resource files with fields needing diff suppression
  - Identify fields that may have equivalent but different representations
  - Apply appropriate suppress function from library
  - Document why each suppression is needed
  - Purpose: Reduce noise in terraform plans
  - _Leverage: `internal/provider/diff_suppress.go`_
  - _Requirements: 4_
  - _Prompt: Implement the task for spec terraform-schema-design-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Identify and apply DiffSuppressFunc to fields that need semantic equality | Restrictions: Only apply where truly needed, document rationale, verify with tests | Success: Semantically equivalent values no longer produce diffs | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

## Phase 5: Documentation

- [x] 10. Create schema design patterns guide
  - File: `docs/schema-patterns.md`
  - Document attribute configurability decision tree
  - Include examples for each pattern
  - Add migration notes for plugin-framework
  - Purpose: Enable developers to choose correct patterns
  - _Leverage: Design document, code examples_
  - _Requirements: 1, 6_
  - _Prompt: Implement the task for spec terraform-schema-design-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer | Task: Create comprehensive schema design patterns documentation | Restrictions: Use real examples from codebase, include decision trees, keep practical | Success: Documentation helps developers choose correct schema patterns | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [x] 11. Document zero value handling patterns
  - File: `docs/zero-value-handling.md`
  - Explain difference between GetOk and Get
  - Document when to use GetOkExists
  - Include examples for bool, int, string fields
  - Purpose: Prevent common mistakes with zero values
  - _Leverage: Design document section on Zero Value Handling_
  - _Requirements: 8_
  - _Prompt: Implement the task for spec terraform-schema-design-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer | Task: Create documentation for zero value handling patterns | Restrictions: Include practical examples, explain SDK v2 vs Framework differences | Success: Documentation prevents zero value mishandling bugs | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

## Phase 6: Audit and Expansion

- [-] 12. Audit all resources for schema pattern compliance
  - Files: All `internal/provider/resource_*.go`
  - Check each resource against pattern decision tree
  - Identify fields needing pattern updates
  - Create list of required changes
  - Purpose: Identify gaps in pattern application
  - _Leverage: Schema patterns documentation_
  - _Requirements: All_
  - _Prompt: Implement the task for spec terraform-schema-design-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Code Auditor | Task: Audit all resources for schema pattern compliance | Restrictions: Read-only audit, document findings, prioritize by risk | Success: Complete list of resources and fields needing updates | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_

- [-] 13. Apply patterns to remaining resources
  - Files: Resources identified in audit
  - Apply WriteOnly to all credential fields
  - Apply RequiresReplace to all immutable fields
  - Apply DiffSuppressFunc where needed
  - Add StateFunc normalizers where appropriate
  - Purpose: Achieve consistent patterns across provider
  - _Leverage: All pattern implementations from previous tasks_
  - _Requirements: All_
  - _Prompt: Implement the task for spec terraform-schema-design-patterns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Apply all schema patterns to remaining resources | Restrictions: Follow established patterns, verify each change with tests, maintain backward compatibility | Success: All resources follow consistent schema patterns | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion with detailed artifacts, then mark as [x] when done_
