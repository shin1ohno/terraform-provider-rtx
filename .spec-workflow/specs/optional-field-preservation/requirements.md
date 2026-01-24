# Requirements Document: Optional Field Preservation

## Introduction

This specification addresses a critical issue in the Terraform provider for RTX routers: when optional fields are not explicitly specified in the Terraform configuration, the current implementation sends default (zero) values to the router, which can unintentionally modify existing settings.

The correct behavior should be: if a user does not specify a field, the current value on the router should be preserved, not overwritten with a default value.

This principle must apply consistently across all resources in the provider.

## Problem Statement

### Current Behavior (Problematic)

```hcl
resource "rtx_admin_user" "example" {
  username = "admin"
  # administrator not specified
}
```

Current implementation:
```go
Administrator: d.Get("administrator").(bool)  // Returns false (zero value)
```

Result: `user attribute admin administrator=off` is sent, removing admin privileges.

### Expected Behavior

If `administrator` is not specified:
1. On **Update**: Preserve the current value from the router
2. On **Create**: Either require the field or use a safe default

## Alignment with Product Vision

This feature ensures that the Terraform provider follows the principle of least surprise and prevents accidental configuration changes that could lock users out of their routers (as experienced with the `administrator` attribute).

## Scope

### Fields This Principle Applies To

This optional field preservation principle SHALL apply to **all optional fields** in all resources, without exception. This includes:

- Boolean fields (e.g., `administrator`, `enabled`)
- Integer fields (e.g., `timeout`, `port`)
- String fields (e.g., `description`, `comment`)
- List/Set fields (e.g., `allowed_hosts`, `filters`)
- Nested block fields (e.g., configuration sub-blocks)

### Rationale for Universal Application

1. **Consistency**: Treating all optional fields the same way makes the provider behavior predictable
2. **Safety**: Any field could potentially have security or operational implications
3. **Terraform Conventions**: This aligns with how other mature Terraform providers handle optional fields
4. **User Expectations**: Users expect unspecified fields to remain unchanged

### No Exclusions

There are **no fields excluded** from this principle. If a field is marked as `Optional: true` in the schema, it must follow the preservation behavior described in this specification.

Note: Fields marked as `Required: true` are not affected by this specification, as they must always be specified by the user.

## Requirements

### Requirement 1: Distinguish Between Unset and Explicitly Set Values

**User Story:** As a Terraform user, I want fields I don't specify to remain unchanged on the router, so that I don't accidentally modify settings I didn't intend to change.

#### Acceptance Criteria

1. WHEN a field is not specified in the Terraform configuration AND the resource already exists THEN the system SHALL preserve the current value on the router
2. WHEN a field is explicitly set to a value (including false/0/empty) THEN the system SHALL apply that value to the router
3. WHEN a field is removed from the configuration after being previously set THEN the system SHALL preserve the last known value (not revert to default)

### Requirement 2: Safe Defaults for Resource Creation

**User Story:** As a Terraform user, I want new resources to have safe defaults when optional fields are not specified, so that I don't accidentally create resources with dangerous configurations.

#### Acceptance Criteria

1. IF a new resource is created AND a critical field (e.g., `administrator`) is not specified THEN the system SHALL either:
   - Use a safe default value (e.g., `administrator=true` for admin users), OR
   - Omit that field from the command (let the router use its default), OR
   - Return a validation error requiring the field to be specified
2. WHEN creating a new resource THEN the system SHALL NOT send `administrator=off` or equivalent dangerous defaults

### Requirement 3: Consistent Behavior Across All Resources

**User Story:** As a Terraform user, I want all resources to behave consistently regarding optional fields, so that I can predict the behavior of the provider.

#### Acceptance Criteria

1. WHEN any resource type has optional fields THEN the system SHALL apply the same preservation logic
2. IF a field uses `Optional: true` and `Computed: true` THEN the system SHALL track whether the field was explicitly set
3. WHEN building commands for the router THEN the system SHALL only include fields that were explicitly specified or have changed

### Requirement 4: Clear Terraform Plan Output

**User Story:** As a Terraform user, I want to see clearly in the plan what will change, so that I can review changes before applying.

#### Acceptance Criteria

1. WHEN `terraform plan` is run THEN the system SHALL show only fields that will actually change
2. WHEN a field is not specified THEN the plan SHALL NOT show a change from current value to default value
3. IF a field will be preserved THEN the plan SHALL show "(unchanged)" or not show the field at all

### Requirement 5: Audit and Identify Affected Resources

**User Story:** As a developer, I want to identify all resources and fields affected by this issue, so that I can ensure complete coverage of the fix.

#### Acceptance Criteria

1. WHEN implementing this specification THEN the system SHALL audit all existing resources for optional fields
2. FOR each resource THEN documentation SHALL list which fields are affected
3. WHEN the fix is complete THEN all resources SHALL be verified to follow the new behavior

## Non-Functional Requirements

### Code Architecture and Modularity

- **Single Responsibility Principle**: Create a shared utility for handling optional field tracking
- **Modular Design**: The solution should be reusable across all resources without code duplication
- **Clear Interfaces**: Define a consistent pattern that all resources must follow

### Backward Compatibility

- Existing Terraform state files must continue to work
- Existing configurations with explicitly set values must continue to work as expected
- The change should not cause unexpected diffs on existing resources

### Performance

- The additional tracking should not significantly impact plan/apply performance
- No additional API calls to the router should be required

### Security

- The fix must prevent accidental removal of security-critical settings (like `administrator` privilege)
- Sensitive fields should continue to be handled appropriately

### Testing

- Unit tests must cover all combinations: field set, field unset, field removed
- Integration tests must verify actual router behavior
- Regression tests must ensure the `administrator` lockout scenario cannot recur

## Affected Resources (To Be Audited)

The following resources need to be audited for optional fields:

- [ ] rtx_admin_user
- [ ] rtx_admin
- [ ] rtx_system
- [ ] rtx_interface
- [ ] rtx_dhcp_scope
- [ ] rtx_l2tp_service
- [ ] rtx_ethernet_filter
- [ ] (All other resources in the provider)

## References

- Incident: User locked out of router due to `administrator=false` being applied when not specified
- Terraform SDK documentation on `Optional` and `Computed` fields
