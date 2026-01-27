# Requirements Document: Plugin Framework Migration

## Introduction

This specification defines the requirements for migrating the terraform-provider-rtx from Terraform Plugin SDK v2 to Terraform Plugin Framework. The migration will be a complete replacement (not incremental) since the provider is not yet widely adopted. The primary goals are:

1. Enable true write-only attributes for sensitive values (no state storage)
2. Adopt modern Terraform provider patterns
3. Simplify codebase with typed schema definitions

## Background

### Current State (SDK v2)
- Provider uses `Sensitive: true` which only masks output but still stores values in state
- Pre-shared keys, passwords, and other secrets are stored in `terraform.tfstate`
- Import operations cannot read sensitive values from router (by design)
- Results in perpetual plan diffs for sensitive attributes after import

### Target State (Plugin Framework)
- True write-only attributes using `WriteOnly: true` modifier
- Sensitive values not stored in state at all
- Support for ephemeral values (Terraform 1.11+)
- Cleaner separation of concerns with typed schema definitions
- No SDK v2 code remaining

## Requirements

### Requirement 1: Complete Framework Migration

**User Story:** As a provider maintainer, I want to fully migrate to Plugin Framework without SDK v2 remnants, so that the codebase is modern and maintainable.

#### Acceptance Criteria

1. WHEN migration is complete THEN no SDK v2 imports SHALL remain in provider code
2. WHEN provider builds THEN only `terraform-plugin-framework` SHALL be used (not `terraform-plugin-sdk`)
3. IF any resource uses SDK v2 patterns THEN it SHALL be rewritten to Framework patterns
4. WHEN all resources are migrated THEN `terraform-plugin-sdk` dependency SHALL be removed from go.mod

### Requirement 2: Write-Only Sensitive Attributes

**User Story:** As a Terraform user, I want sensitive values (passwords, pre-shared keys) to not be stored in state, so that my secrets are not exposed in backend storage.

#### Acceptance Criteria

1. WHEN `pre_shared_key` is set on `rtx_ipsec_tunnel` THEN the value SHALL NOT appear in state
2. WHEN `tunnel_auth_password` is set on `rtx_l2tp` THEN the value SHALL NOT appear in state
3. WHEN `admin_password` is set on `rtx_admin` THEN the value SHALL NOT appear in state
4. WHEN `password` is set on any resource THEN the value SHALL NOT appear in state
5. WHEN a write-only attribute is configured THEN terraform plan SHALL NOT show perpetual diffs
6. IF Terraform version < 1.11 THEN provider SHALL emit a clear error about version requirement

### Requirement 3: Typed Schema Definitions

**User Story:** As a provider maintainer, I want type-safe schema definitions, so that compile-time errors catch schema mistakes.

#### Acceptance Criteria

1. WHEN defining resource schemas THEN typed attribute definitions SHALL be used
2. IF attribute type mismatch occurs THEN Go compiler SHALL catch it (not runtime)
3. WHEN reading/writing state THEN typed models SHALL be used instead of `map[string]interface{}`

### Requirement 4: Resource Structure Modernization

**User Story:** As a provider maintainer, I want resources to follow Plugin Framework patterns, so that they are consistent and testable.

#### Acceptance Criteria

1. WHEN implementing a resource THEN it SHALL implement `resource.Resource` interface
2. WHEN implementing CRUD operations THEN they SHALL use `*resource.CreateRequest`/`*resource.CreateResponse` patterns
3. IF resource supports import THEN it SHALL implement `resource.ResourceWithImportState`
4. WHEN defining validators THEN Framework validators SHALL be used

### Requirement 5: Minimum Terraform Version

**User Story:** As a user, I want clear version requirements, so that I know what Terraform version to use.

#### Acceptance Criteria

1. WHEN provider initializes THEN minimum Terraform version SHALL be 1.11
2. IF user runs older Terraform THEN clear error message SHALL indicate version requirement
3. WHEN documenting provider THEN version requirement SHALL be prominently displayed

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Each resource in its own file with typed models
- **Modular Design**: Shared schema helpers, validators, and plan modifiers in separate packages
- **Clear Interfaces**: All resources implement standard Framework interfaces

### Performance
- Migration SHALL NOT introduce measurable performance regression
- Provider startup time SHALL remain under 1 second

### Security
- Write-only attributes SHALL be the default for all password/key fields
- State files SHALL NOT contain any sensitive credential after migration
- Logging SHALL sanitize sensitive values (existing behavior maintained)

### Reliability
- All existing acceptance tests SHALL be migrated and pass
- New Framework resources SHALL have equivalent or better test coverage

### Breaking Changes (Accepted)
- State format changes are acceptable (users will re-import)
- Schema attribute name changes are acceptable if documented
- Minimum Terraform version increase to 1.11 is acceptable

## Resources to Migrate

| Resource | Sensitive Attributes | Priority |
|----------|---------------------|----------|
| `rtx_ipsec_tunnel` | pre_shared_key | High |
| `rtx_l2tp` | tunnel_auth_password, ipsec_profile.pre_shared_key | High |
| `rtx_admin` | admin_password, login_password | High |
| `rtx_admin_user` | password | High |
| `rtx_ddns` | password | High |
| `rtx_interface` | - | Normal |
| `rtx_ipv6_interface` | - | Normal |
| `rtx_dhcp_scope` | - | Normal |
| `rtx_dhcp_binding` | - | Normal |
| `rtx_bridge` | - | Normal |
| `rtx_static_route` | - | Normal |
| `rtx_nat_masquerade` | - | Normal |
| `rtx_access_list_*` | - | Normal |
| `rtx_sshd` | - | Normal |
| `rtx_sshd_*` | - | Normal |
| `rtx_system` | - | Normal |
| `rtx_syslog` | - | Normal |
| `rtx_httpd` | - | Normal |
| `rtx_sftpd` | - | Normal |
| `rtx_dns_server` | - | Normal |
| `rtx_l2tp_service` | - | Normal |
| `rtx_ipsec_transport` | - | Normal |
| `rtx_ipv6_prefix` | - | Normal |

## References

- [Terraform Plugin Framework Migration Guide](https://developer.hashicorp.com/terraform/plugin/framework/migrating)
- [Write-Only Arguments Documentation](https://developer.hashicorp.com/terraform/plugin/framework/resources/write-only-arguments)
- [Plugin Framework Benefits](https://developer.hashicorp.com/terraform/plugin/framework-benefits)
