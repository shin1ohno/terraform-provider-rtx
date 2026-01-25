# Requirements Document: State Drift Fix for Access List Resources

## Introduction

Fix persistent terraform plan diffs that occur after `terraform apply` for access list related resources. These diffs indicate state drift where the Read function doesn't properly retrieve the values that were written, causing Terraform to believe the infrastructure has changed when it hasn't.

**Affected Resources:**
- `rtx_access_list_mac` - MAC address access list with entry-level attributes
- `rtx_interface` - Interface configuration with access list bindings
- `rtx_ipv6_interface` - IPv6 interface configuration with access list bindings

## Alignment with Product Vision

From `product.md`:
> **State Clarity**: Persist only configuration in Terraform state; do not store operational/runtime status to avoid perpetual diffs

From `tech.md`:
> **Terraform State Handling**: Persist only configuration attributes in Terraform state; Do not store operational/runtime status values to avoid perpetual diffs

This fix ensures that configuration attributes (access list bindings) are properly read back from the router, eliminating perpetual diffs. These ARE configuration values, not runtime status.

## Requirements

### Requirement 1: Fix rtx_access_list_mac State Drift

**User Story:** As a network administrator, I want `rtx_access_list_mac` resources to have stable state after apply, so that `terraform plan` shows no changes when the router configuration matches my Terraform code.

#### Acceptance Criteria

1. WHEN `filter_id` is set in configuration THEN the system SHALL preserve the value in state after Read
2. WHEN `ace_action` is "pass" or "pass-nolog" THEN the system SHALL treat them as equivalent and suppress diff
3. WHEN `ace_action` is "reject" or "reject-nolog" THEN the system SHALL treat them as equivalent and suppress diff
4. WHEN `source_any = true` THEN the system SHALL NOT show diff for `source_address` field
5. WHEN `destination_any = true` THEN the system SHALL NOT show diff for `destination_address` field
6. WHEN `source_address = "*:*:*:*:*:*"` THEN the system SHALL normalize to `source_any = true`

### Requirement 2: Fix rtx_interface Access List Binding State Drift

**User Story:** As a network administrator, I want `rtx_interface` access list bindings to persist in state after apply, so that `terraform plan` shows no changes when the router has the correct filters applied.

#### Acceptance Criteria

1. WHEN `access_list_ip_in` is configured THEN the system SHALL read the value back from router configuration
2. WHEN `access_list_ip_out` is configured THEN the system SHALL read the value back from router configuration
3. WHEN `access_list_ip_dynamic_in` is configured THEN the system SHALL read the value back from router configuration
4. WHEN `access_list_ip_dynamic_out` is configured THEN the system SHALL read the value back from router configuration
5. WHEN `access_list_mac_in` is configured THEN the system SHALL read the value back from router configuration
6. WHEN `access_list_mac_out` is configured THEN the system SHALL read the value back from router configuration
7. IF access list binding is removed from router config THEN the system SHALL show the attribute as empty in state

### Requirement 3: Fix rtx_ipv6_interface Access List Binding State Drift

**User Story:** As a network administrator, I want `rtx_ipv6_interface` access list bindings to persist in state after apply, so that `terraform plan` shows no changes when the router has the correct IPv6 filters applied.

#### Acceptance Criteria

1. WHEN `access_list_ipv6_in` is configured THEN the system SHALL read the value back from router configuration
2. WHEN `access_list_ipv6_out` is configured THEN the system SHALL read the value back from router configuration
3. WHEN `access_list_ipv6_dynamic_in` is configured THEN the system SHALL read the value back from router configuration
4. WHEN `access_list_ipv6_dynamic_out` is configured THEN the system SHALL read the value back from router configuration
5. IF access list binding is removed from router config THEN the system SHALL show the attribute as empty in state

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Service layer handles data retrieval; resource layer handles schema mapping
- **Modular Design**: Parser, service, and resource layers remain separated
- **Dependency Management**: No new dependencies required
- **Clear Interfaces**: Service methods return complete configuration including access list bindings

### Performance
- No additional SSH commands required; access list bindings should be parsed from existing interface configuration output

### Security
- No security implications; this is a state management fix

### Reliability
- Read functions must be idempotent and return consistent results
- Missing access list bindings should be represented as empty strings, not errors

### Usability
- After `terraform apply`, subsequent `terraform plan` must show "No changes" when configuration matches
- Clear error messages if access list name resolution fails
