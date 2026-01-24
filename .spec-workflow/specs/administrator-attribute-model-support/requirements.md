# Requirements: Administrator Attribute Model Support

## Overview

RTX routers have different `administrator` attribute value formats depending on the model. The current implementation only supports `on`/`off` format, but some models use `1`/`2`/`off` format. This causes incorrect parsing and command generation for those models.

## Problem Statement

The `user attribute` command's `administrator` parameter has different valid values depending on the RTX model:

### Model Group A (on/off format)
- RTX1220, RTX1210, RTX830, RTX5000, RTX3500, vRX Amazon EC2, vRX VMware ESXi

| Value | Description |
|-------|-------------|
| `on` | Can elevate to administrator mode. Can access GUI admin pages. Can SFTP with admin password. |
| `off` | Cannot elevate to administrator mode. Cannot access GUI admin pages. Cannot SFTP with admin password. |

### Model Group B (1/2/off format)
- RTX840, RTX1300, RTX3510, and others

| Value | Description |
|-------|-------------|
| `2` | Can elevate to administrator mode WITHOUT password. Can access GUI admin pages. Can SFTP with admin password. |
| `1` | Can elevate to administrator mode WITH password. Cannot access GUI admin pages. Can SFTP with admin password. |
| `off` | Cannot elevate to administrator mode. Cannot access GUI admin pages. Cannot SFTP with admin password. |

### Current Implementation Issues

1. **Parsing**: Only `on` is recognized as `true`. Values `1` and `2` are treated as `false`.
2. **Command Generation**: Only outputs `on`/`off`. Models requiring `1`/`2` receive incorrect commands.

## Requirements

### Requirement 1: Parse All Administrator Attribute Formats

**User Story:** As a Terraform user managing RTX routers, I want the provider to correctly parse `administrator=1`, `administrator=2`, and `administrator=on` as administrator-enabled states, so that imported resources reflect the actual router configuration.

#### Acceptance Criteria

1. WHEN parsing `administrator=on` THEN system SHALL set Administrator to `true`
2. WHEN parsing `administrator=1` THEN system SHALL set Administrator to `true`
3. WHEN parsing `administrator=2` THEN system SHALL set Administrator to `true` AND set a flag indicating password-less elevation
4. WHEN parsing `administrator=off` THEN system SHALL set Administrator to `false`

### Requirement 2: Generate Model-Appropriate Commands

**User Story:** As a Terraform user, I want the provider to generate the correct `user attribute` command format for my RTX model, so that configurations are applied correctly.

#### Acceptance Criteria

1. WHEN generating command for Model Group A THEN system SHALL use `administrator=on` or `administrator=off`
2. WHEN generating command for Model Group B THEN system SHALL use `administrator=2`, `administrator=1`, or `administrator=off`
3. IF model information is not available THEN system SHALL default to `on`/`off` format (most common)

### Requirement 3: Terraform Schema Enhancement

**User Story:** As a Terraform user with Model Group B routers, I want to specify whether administrator elevation requires a password, so that I can configure the appropriate access level.

#### Acceptance Criteria

1. WHEN `administrator = true` AND `administrator_password_required = false` (or not specified) THEN system SHALL generate `administrator=on` (Group A) or `administrator=2` (Group B)
2. WHEN `administrator = true` AND `administrator_password_required = true` THEN system SHALL generate `administrator=on` (Group A) or `administrator=1` (Group B)
3. WHEN `administrator = false` THEN system SHALL generate `administrator=off`

### Requirement 4: Backward Compatibility

**User Story:** As an existing Terraform user, I want my current configurations to continue working without modification.

#### Acceptance Criteria

1. IF `administrator_password_required` is not specified THEN system SHALL use default behavior (equivalent to current `on`/`off`)
2. WHEN upgrading provider THEN existing state files SHALL remain valid
3. WHEN running `terraform plan` with existing config THEN no unexpected changes SHALL appear

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Separate model detection logic from command generation
- **Modular Design**: Create a model-specific configuration helper
- **Clear Interfaces**: Define clean contracts for administrator value handling

### Performance
- Model detection should not require additional router queries

### Reliability
- Fallback to `on`/`off` format when model is unknown
- Log warnings when model-specific behavior is applied

### Usability
- Clear documentation on which models use which format
- Helpful error messages when invalid combinations are specified

## Out of Scope

- Automatic model detection from router (would require additional implementation)
- GUI page access level differences between `1` and `2` (documentation only)

## References

- RTX Command Reference: Section 4.10 "Setting User Attributes"
- Affected file: `internal/rtx/parsers/admin.go`
