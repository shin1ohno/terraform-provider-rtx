# Requirements Document: Terraform Provider Testing Patterns

## Introduction

This specification establishes comprehensive testing patterns and best practices for the Terraform provider for RTX routers. The goal is to ensure provider reliability, prevent regressions, and catch common issues such as perpetual diffs, state inconsistencies, and import failures before they reach users.

These patterns are based on HashiCorp's official recommendations (2025-2026) and lessons learned from mature providers like AWS and GCP.

## Alignment with Product Vision

This feature supports the product vision by:
- **Reliability**: Ensuring users can trust the provider to manage their critical network infrastructure
- **Maintainability**: Establishing patterns that make it easier to add new resources with confidence
- **Quality**: Catching bugs early in development through comprehensive test coverage

## Requirements

### Requirement 1: Perpetual Diff Prevention Tests

**User Story:** As a provider developer, I want tests that verify resources don't produce unexpected diffs on re-apply, so that users don't see confusing "changes" when nothing has actually changed.

#### Acceptance Criteria

1. WHEN a resource is created and `terraform plan` is run again without config changes THEN the system SHALL report no changes (empty plan)
2. WHEN a resource has Optional+Computed fields not specified in config THEN the system SHALL preserve state values without showing diffs
3. IF a resource uses semantic equality for custom types (e.g., JSON normalization) THEN tests SHALL verify equivalent values don't produce diffs
4. WHEN testing for empty plans THEN the system SHALL use `plancheck.ExpectEmptyPlan()` or equivalent SDK v2 pattern

### Requirement 2: State Consistency Tests

**User Story:** As a provider developer, I want tests that verify state matches the actual resource configuration, so that Terraform state accurately reflects reality.

#### Acceptance Criteria

1. WHEN a resource is created THEN the Read function SHALL populate all Computed fields in state
2. WHEN a resource is updated THEN state values SHALL reflect the actual post-update configuration
3. IF a discrepancy exists between state and actual resource THEN tests SHALL detect and fail
4. WHEN using Optional+Computed fields THEN tests SHALL verify state contains either user-specified or router-returned values

### Requirement 3: Import Testing

**User Story:** As a provider developer, I want tests that verify resources can be imported correctly, so that users can adopt existing infrastructure into Terraform management.

#### Acceptance Criteria

1. WHEN a resource is imported THEN the system SHALL populate state with all current attribute values
2. WHEN imported state is used for a subsequent plan THEN the system SHALL show no changes (if config matches)
3. FOR each resource type THEN at least one import test SHALL exist
4. WHEN import ID format changes THEN tests SHALL verify backward compatibility or clear migration path

### Requirement 4: Plan Check Patterns

**User Story:** As a provider developer, I want structured plan checks that verify resource actions and values, so that I can assert precise behavior in tests.

#### Acceptance Criteria

1. WHEN testing resource creation THEN tests SHALL verify `plan.ResourceActionCreate`
2. WHEN testing resource updates THEN tests SHALL verify `plan.ResourceActionUpdate`
3. WHEN testing resource deletion THEN tests SHALL verify `plan.ResourceActionDestroy`
4. WHEN testing no-op scenarios THEN tests SHALL verify `plan.ResourceActionNoop`
5. WHEN specific attribute values are expected THEN tests SHALL use `plancheck.ExpectKnownValue()`

### Requirement 5: Acceptance Test Structure

**User Story:** As a provider developer, I want a consistent test structure across all resources, so that tests are easy to write, understand, and maintain.

#### Acceptance Criteria

1. WHEN creating acceptance tests THEN the system SHALL follow the pattern: Basic → Update → Import → Error cases
2. WHEN tests require real router access THEN PreCheck functions SHALL verify connectivity and credentials
3. WHEN tests create resources THEN CheckDestroy functions SHALL verify cleanup
4. WHEN tests use random values THEN acctest helpers or equivalent SHALL be used for uniqueness

### Requirement 6: Value Comparison and State Checks

**User Story:** As a provider developer, I want to verify that specific values in state match expected values, so that I can catch subtle bugs in value handling.

#### Acceptance Criteria

1. WHEN testing computed values (IDs, timestamps) THEN tests SHALL use `statecheck.ExpectKnownValue()`
2. WHEN testing value stability across steps THEN tests SHALL use Value Comparers
3. IF a value should remain constant after updates THEN tests SHALL verify with `compare.ValuesSame()`
4. IF a value should change after updates THEN tests SHALL verify with `compare.ValuesDiffer()`

### Requirement 7: Test Helpers and Utilities

**User Story:** As a provider developer, I want reusable test helpers that reduce boilerplate, so that writing new tests is fast and consistent.

#### Acceptance Criteria

1. WHEN multiple tests need similar configurations THEN shared config builder functions SHALL be available
2. WHEN tests need to verify specific RTX router behavior THEN RTX-specific assertion helpers SHALL be available
3. WHEN tests run in parallel THEN resource names SHALL be unique to avoid conflicts
4. WHEN test fixtures are needed THEN they SHALL be stored in `testdata/` directories

### Requirement 8: State Migration Tests

**User Story:** As a provider developer, I want tests that verify state upgrades work correctly when schema versions change, so that users can safely upgrade the provider without losing state.

#### Acceptance Criteria

1. WHEN SchemaVersion is incremented THEN unit tests SHALL verify StateUpgrader functions
2. WHEN a StateUpgrader is added THEN tests SHALL verify transformation from old format to new format
3. WHEN testing state migration THEN acceptance tests SHALL use ExternalProviders to apply with old provider version
4. WHEN testing cross-version upgrade THEN the system SHALL verify empty plan after upgrade (no unexpected changes)
5. WHEN StateUpgrader handles multiple versions THEN each version path SHALL be tested
6. WHEN state migration fails THEN clear error messages SHALL be provided

### Requirement 9: WriteOnly and Sensitive Attribute Tests

**User Story:** As a provider developer, I want tests that verify sensitive and write-only attributes are handled correctly, so that credentials are not exposed in state or logs.

#### Acceptance Criteria

1. WHEN a WriteOnly attribute is set THEN tests SHALL verify the value is NOT stored in state
2. WHEN a Sensitive attribute is used THEN tests SHALL verify the value is masked in plan output
3. WHEN testing password fields THEN tests SHALL verify the password is sent to router but not readable from state
4. WHEN updating a WriteOnly attribute THEN tests SHALL verify the update is applied without exposing the value
5. WHEN importing a resource with WriteOnly attributes THEN tests SHALL verify appropriate handling (null or require re-specification)

## Non-Functional Requirements

### Code Architecture and Modularity

- **Single Responsibility Principle**: Test helpers should have focused, specific purposes
- **Modular Design**: Shared test utilities should be in `internal/provider/acctest/` or similar
- **Dependency Management**: Test code should not leak into production builds
- **Clear Interfaces**: Test helper functions should have self-documenting signatures

### Performance

- Tests should leverage `t.Parallel()` where safe
- Test suites should complete within reasonable CI time limits
- Network-dependent tests should use appropriate timeouts

### Reliability

- Tests should be deterministic (no flaky tests)
- Tests should clean up resources even on failure
- Tests should handle router connectivity issues gracefully

### Maintainability

- Each resource should have its own test file (`resource_*_test.go`)
- Test code should follow the same style guidelines as production code
- Complex test scenarios should be documented with comments

## Scope

### Included Resources for Initial Implementation

The following resources will receive comprehensive test coverage first:

- [ ] rtx_admin_user (as reference implementation)
- [ ] rtx_admin
- [ ] Resources with Optional+Computed fields (identified in optional-field-preservation spec)

### Test Categories to Implement

1. **Unit Tests**: Field helpers, parsers, command builders
2. **Acceptance Tests**: Full CRUD lifecycle with real router
3. **Regression Tests**: Specific bug prevention (e.g., administrator lockout)

## References

- HashiCorp Testing Patterns Guide (2025-2026)
- terraform-plugin-testing v1.14+ documentation
- terraform-plugin-sdk/v2 testing documentation
- Existing `optional-field-preservation` spec for related patterns
