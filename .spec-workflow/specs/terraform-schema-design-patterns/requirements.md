# Requirements Document: Terraform Schema Design Patterns

## Introduction

This specification establishes comprehensive schema design patterns and best practices for the Terraform provider for RTX routers. Proper schema design is fundamental to creating a provider that behaves predictably, integrates well with Terraform's workflow, and provides an excellent user experience.

These patterns are based on HashiCorp's official terraform-plugin-framework and terraform-plugin-sdk/v2 documentation (2025-2026), covering attribute configurability, plan modification, custom types, and migration strategies.

## Alignment with Product Vision

This feature supports the product vision by:
- **Predictability**: Users can understand and predict provider behavior based on schema design
- **Consistency**: All resources follow the same patterns, reducing cognitive load
- **Future-proofing**: Patterns prepare for eventual migration to terraform-plugin-framework
- **Safety**: Proper schema design prevents accidental data loss or security issues

## Requirements

### Requirement 1: Attribute Configurability Matrix

**User Story:** As a provider developer, I want clear guidelines for choosing between Required, Optional, and Computed flags, so that I select the correct combination for each attribute.

#### Acceptance Criteria

1. WHEN an attribute must always be provided by the user THEN it SHALL be marked `Required: true`
2. WHEN an attribute may be provided by the user but has a sensible default THEN it SHALL be marked `Optional: true`
3. WHEN an attribute is read-only and populated by the provider/API THEN it SHALL be marked `Computed: true`
4. WHEN an attribute can be provided by the user OR populated by the API THEN it SHALL be marked `Optional: true, Computed: true`
5. WHEN combining flags THEN `Required: true` and `Computed: true` SHALL NOT be combined (contradictory)
6. WHEN combining flags THEN `Required: true` and `Optional: true` SHALL NOT be combined (contradictory)

### Requirement 2: WriteOnly Attributes for Sensitive Data

**User Story:** As a provider developer, I want to properly handle sensitive write-only data like passwords, so that credentials are not stored in state and cannot be read back.

#### Acceptance Criteria

1. WHEN an attribute contains sensitive credentials (passwords, API keys, tokens) THEN it SHALL be considered for WriteOnly treatment
2. WHEN using WriteOnly attributes THEN `Computed: true` SHALL NOT be combined (WriteOnly values cannot be read)
3. WHEN a WriteOnly attribute is set THEN the value SHALL be sent to the router but NOT stored in state
4. WHEN implementing WriteOnly in SDK v2 THEN the system SHALL use `Sensitive: true` with appropriate handling
5. WHEN migrating to plugin-framework THEN the system SHALL use the `WriteOnly: true` flag

### Requirement 3: RequiresReplace Patterns

**User Story:** As a provider developer, I want to correctly identify attributes that force resource recreation when changed, so that users understand the impact of their changes.

#### Acceptance Criteria

1. WHEN an attribute is immutable after creation (e.g., username, resource ID) THEN it SHALL use RequiresReplace
2. WHEN an attribute change requires recreation only under certain conditions THEN it SHALL use RequiresReplaceIf
3. WHEN an attribute change requires recreation only if user explicitly configured it THEN it SHALL use RequiresReplaceIfConfigured
4. WHEN implementing in SDK v2 THEN the system SHALL use `ForceNew: true`
5. WHEN implementing in plugin-framework THEN the system SHALL use plan modifiers
6. WHEN a RequiresReplace attribute is changed THEN terraform plan SHALL clearly show resource will be destroyed and recreated

### Requirement 4: DiffSuppressFunc and Semantic Equality

**User Story:** As a provider developer, I want to suppress irrelevant diffs when values are semantically equal, so that users don't see noise in their plans.

#### Acceptance Criteria

1. WHEN two values are semantically equal but syntactically different (e.g., JSON with different key order) THEN the system SHALL suppress the diff
2. WHEN implementing in SDK v2 THEN the system SHALL use `DiffSuppressFunc`
3. WHEN implementing in plugin-framework THEN the system SHALL use custom types with semantic equality or plan modifiers
4. WHEN normalizing values THEN the preferred approach SHALL be to canonicalize before storing in state
5. WHEN case-insensitive comparison is needed THEN the system SHALL normalize to lowercase before storing
6. WHEN whitespace differences should be ignored THEN the system SHALL trim/normalize whitespace

### Requirement 5: Custom Types for Domain-Specific Values

**User Story:** As a provider developer, I want to use custom types for domain-specific values, so that validation and equality checks are handled automatically.

#### Acceptance Criteria

1. WHEN an attribute represents an IP address or CIDR THEN the system SHALL use or create a custom IP/CIDR type
2. WHEN an attribute represents JSON data THEN the system SHALL use a custom JSON type with semantic equality
3. WHEN an attribute represents a timestamp THEN the system SHALL use a custom time type with format normalization
4. WHEN an attribute has domain-specific validation (MAC address, RTX command format) THEN the system SHALL consider a custom type
5. WHEN implementing custom types THEN the system SHALL prefer existing community packages (nettypes, jsontypes, timetypes)
6. WHEN custom validation is needed THEN custom types SHALL implement the Validate() method

### Requirement 6: UseStateForUnknown Pattern (Framework Migration Prep)

**User Story:** As a provider developer, I want to understand the UseStateForUnknown pattern, so that I can prepare for eventual migration to terraform-plugin-framework.

#### Acceptance Criteria

1. WHEN an Optional+Computed attribute value is not expected to change frequently THEN UseStateForUnknown SHALL be used (in framework)
2. WHEN UseStateForUnknown is used on list/set nested attributes THEN special care SHALL be taken due to index instability
3. WHEN the attribute is at the top level or in a single-nested block THEN UseStateForUnknown is safe to use
4. WHEN preparing SDK v2 code for migration THEN documentation SHALL note where UseStateForUnknown will be needed
5. WHEN attribute value is truly computed fresh each time THEN UseStateForUnknown SHALL NOT be used

### Requirement 7: Nested Block Design Patterns

**User Story:** As a provider developer, I want clear guidelines for choosing between list, set, and single nested blocks, so that I select the appropriate structure for each use case.

#### Acceptance Criteria

1. WHEN elements have a defined order that matters THEN ListNestedAttribute/Block SHALL be used
2. WHEN elements are order-independent and unique THEN SetNestedAttribute/Block SHALL be used
3. WHEN exactly one nested object exists THEN SingleNestedAttribute/Block SHALL be used
4. WHEN nested blocks have optional fields THEN each child attribute SHALL be marked appropriately (Optional/Computed)
5. WHEN a nested element needs a stable identifier for updates THEN an `id` or `name` field SHALL be included
6. WHEN elements lack a stable server-side ID THEN a synthetic identifier strategy SHALL be documented

### Requirement 8: Zero Value Handling

**User Story:** As a provider developer, I want to correctly distinguish between "not set" and "explicitly set to zero/empty", so that user intent is preserved.

#### Acceptance Criteria

1. WHEN a user explicitly sets a value to `""`, `0`, or `false` THEN the system SHALL treat it as intentional
2. WHEN a user does not specify a value THEN the system SHALL NOT send a zero value to the API
3. WHEN using SDK v2 THEN `GetOk()` SHALL be used to distinguish set from unset
4. WHEN using plugin-framework THEN `IsNull()` and `IsUnknown()` SHALL be used appropriately
5. WHEN an API returns empty string to mean "not configured" THEN the state SHALL store null, not empty string
6. WHEN converting between Terraform types and Go types THEN null/unknown checks SHALL precede value access

### Requirement 9: Schema Versioning and State Migration

**User Story:** As a provider developer, I want to safely evolve schema over time without breaking existing state files, so that users can upgrade smoothly.

#### Acceptance Criteria

1. WHEN schema layout changes (rename, restructure) THEN SchemaVersion SHALL be incremented
2. WHEN SchemaVersion increases THEN a StateUpgrader function SHALL be provided
3. WHEN writing StateUpgrader THEN the function SHALL handle all previous versions
4. WHEN testing state migration THEN unit tests SHALL verify upgrader logic
5. WHEN testing state migration THEN acceptance tests SHALL verify upgrade from previous provider versions
6. WHEN breaking changes are unavoidable THEN migration guide documentation SHALL be provided

## Non-Functional Requirements

### Code Architecture and Modularity

- **Single Responsibility Principle**: Each custom type should handle one domain concept
- **Modular Design**: Custom types and validators should be in dedicated packages
- **Reusability**: Patterns should be easily applicable to new resources
- **Clear Interfaces**: Schema builders should have consistent signatures

### Documentation

- Each pattern should be documented with examples
- Decision rationale should be recorded for non-obvious choices
- Migration notes should be maintained for framework transition

### Backward Compatibility

- Schema changes must not break existing state files without migration
- New optional fields should have sensible defaults
- Changing Optional+Computed to Computed-only is a breaking change

### Performance

- Custom type validation should not significantly impact plan time
- Semantic equality checks should be efficient
- State migration should handle large state files gracefully

## Scope

### Patterns to Document and Implement

1. **Attribute Configurability**: Required/Optional/Computed decision tree
2. **WriteOnly**: Sensitive credential handling
3. **RequiresReplace**: Immutable attribute handling
4. **DiffSuppressFunc**: Semantic equality patterns
5. **Custom Types**: IP/CIDR, JSON, Time, RTX-specific types
6. **UseStateForUnknown**: Framework migration preparation
7. **Nested Blocks**: List vs Set vs Single selection
8. **Zero Values**: Null vs Zero distinction
9. **State Migration**: Version upgrade patterns

### Resources to Apply Patterns

All existing and future resources should follow these patterns. Priority:
- [ ] rtx_admin_user (reference implementation)
- [ ] rtx_admin
- [ ] Resources with password/credential fields (WriteOnly)
- [ ] Resources with immutable identifiers (RequiresReplace)

## References

- HashiCorp terraform-plugin-framework Handling Data documentation
- HashiCorp terraform-plugin-sdk/v2 Schema documentation
- terraform-plugin-framework-jsontypes package
- terraform-plugin-framework-nettypes package
- terraform-plugin-framework-timetypes package
- Existing `optional-field-preservation` spec
- Existing `terraform-provider-testing-patterns` spec
