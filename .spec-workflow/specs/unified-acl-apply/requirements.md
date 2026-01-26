# Requirements Document: Unified ACL Apply Design

## Introduction

This specification unifies the ACL (Access Control List) design pattern across all ACL types in the Terraform RTX provider. The goal is to establish a single, consistent pattern for ACL definition and interface application across all ACL types, with automatic sequence management and comprehensive validation.

This is a breaking change that prioritizes clean design over backward compatibility.

## Alignment with Product Vision

This feature directly supports the product principles outlined in product.md:

1. **Cisco-Compatible Syntax**: The unified design aligns with Cisco IOS XE provider patterns where ACLs are defined as groups and applied to interfaces.
2. **Follow RTX CLI Semantics**: The apply block mirrors how RTX CLI commands work (`ip lan1 secure filter in ...`).
3. **State Clarity**: All ACL configuration including interface application is managed in one resource.

## Breaking Changes

This specification introduces the following breaking changes:

| Change | Impact |
|--------|--------|
| Remove `access_list_*_in/out` from `rtx_interface` | ACL application moves to ACL resources only |
| Remove `access_list_*_in/out` from `rtx_pp_interface` | ACL application moves to ACL resources only |
| Remove `access_list_*_in/out` from `rtx_ipv6_interface` | ACL application moves to ACL resources only |
| Redesign `rtx_access_list_ip` | Individual filter → Group resource |
| Redesign `rtx_access_list_ipv6` | Individual filter → Group resource |
| `rtx_access_list_mac` apply block | Single → Multiple |

## Target ACL Types

All ACL types will adopt the unified design:

| ACL Resource | Current Design | Target Design |
|--------------|----------------|---------------|
| `rtx_access_list_mac` | Group + single apply | Group + multiple apply + auto sequence |
| `rtx_access_list_extended` | Group only | Group + multiple apply + auto sequence |
| `rtx_access_list_ipv6` | Individual filter | Group + multiple apply + auto sequence |
| `rtx_access_list_ip` | Individual filter | Group + multiple apply + auto sequence |
| `rtx_access_list_ip_dynamic` | TBD | Group + multiple apply + auto sequence |
| `rtx_access_list_ipv6_dynamic` | TBD | Group + multiple apply + auto sequence |

## Target Interfaces for Apply

ACLs can be applied to the following interface types:

| Interface Type | Example | IP ACL | IPv6 ACL | MAC ACL | Dynamic ACL |
|----------------|---------|--------|----------|---------|-------------|
| LAN | `lan1`, `lan2` | Yes | Yes | Yes | Yes |
| Bridge | `bridge1` | Yes | Yes | Yes | Yes |
| PP (PPPoE) | `pp1` | Yes | Yes | No | Yes |
| Tunnel | `tunnel1` | Yes | Yes | No | Yes |

## Requirements

### Requirement 1: Automatic Sequence Management

**User Story:** As a network administrator, I want ACL entry sequences to be automatically calculated based on definition order, so that I don't have to manually manage sequence numbers.

#### Acceptance Criteria

1. WHEN a user specifies `sequence_start` attribute THEN the provider SHALL automatically calculate sequence numbers for all entries based on definition order.
2. IF `sequence_step` is specified THEN the provider SHALL use that value as the increment; IF omitted THEN the provider SHALL use default value of 1.
3. WHEN `sequence_start` is specified THEN entry-level `sequence` attributes SHALL be prohibited (validation error).
4. WHEN `sequence_start` is NOT specified THEN entry-level `sequence` attributes SHALL be required for each entry.
5. WHEN reading back the resource THEN the provider SHALL return the calculated sequence values in the state.

#### Example Usage

```hcl
# Automatic sequence mode
resource "rtx_access_list_extended" "web_traffic" {
  name           = "allow_web"
  sequence_start = 100
  sequence_step  = 10   # Optional, default is 1

  entry {
    # sequence = 100 (auto-calculated)
    ace_rule_action        = "permit"
    ace_rule_protocol      = "tcp"
    source_any             = true
    destination_any        = true
    destination_port_equal = "80"
  }

  entry {
    # sequence = 110 (auto-calculated)
    ace_rule_action        = "permit"
    ace_rule_protocol      = "tcp"
    source_any             = true
    destination_any        = true
    destination_port_equal = "443"
  }

  entry {
    # sequence = 120 (auto-calculated)
    ace_rule_action   = "deny"
    ace_rule_protocol = "ip"
    source_any        = true
    destination_any   = true
  }

  apply {
    interface = "lan1"
    direction = "in"
  }
}

# Manual sequence mode
resource "rtx_access_list_extended" "manual_policy" {
  name = "manual_policy"
  # sequence_start not specified = manual mode

  entry {
    sequence          = 10    # Required in manual mode
    ace_rule_action   = "permit"
    ace_rule_protocol = "tcp"
    ...
  }

  entry {
    sequence          = 9999  # Required in manual mode
    ace_rule_action   = "deny"
    ace_rule_protocol = "ip"
    ...
  }
}

# Invalid: mixing modes
resource "rtx_access_list_extended" "invalid" {
  name           = "invalid"
  sequence_start = 100

  entry {
    sequence = 50  # ERROR: sequence cannot be specified in auto mode
    ...
  }
}
```

### Requirement 2: Sequence Collision Detection

**User Story:** As a network administrator, I want the provider to detect sequence collisions between ACL resources, so that I can avoid configuration conflicts.

#### Acceptance Criteria

1. WHEN applying an ACL THEN the provider SHALL check if the sequence numbers conflict with existing filters on the router.
2. IF a sequence collision is detected during Apply THEN the provider SHALL return an error with details about the conflicting sequences and their owners.
3. WHEN planning changes THEN the provider SHALL detect potential sequence collisions with other ACLs in the same Terraform state and return an error.
4. WHEN multiple ACL resources use overlapping sequence ranges THEN the provider SHALL detect and report the conflict as an error.

#### Validation Behavior

| Phase | Detection | Action |
|-------|-----------|--------|
| Plan | State-based | Error |
| Apply | Router-based | Error |

### Requirement 3: Unified ACL Schema with Multiple Apply Blocks

**User Story:** As a network administrator, I want to define ACL rules and their interface applications in a single resource with multiple apply targets, so that I can manage access control configuration atomically.

#### Acceptance Criteria

1. WHEN a user defines any ACL resource THEN the provider SHALL support multiple `apply` blocks.
2. WHEN an `apply` block specifies `interface` and `direction` THEN the provider SHALL apply the ACL entries to that interface.
3. IF `filter_ids` is omitted in the `apply` block THEN the provider SHALL automatically apply all entry sequences in order.
4. IF `filter_ids` is specified THEN the provider SHALL apply only the specified filter IDs in the given order.
5. WHEN multiple `apply` blocks are defined THEN the provider SHALL apply the ACL to each interface independently.

#### Example Usage

```hcl
resource "rtx_access_list_extended" "multi_interface" {
  name           = "multi_interface_policy"
  sequence_start = 100

  entry {
    ace_rule_action        = "permit"
    ace_rule_protocol      = "tcp"
    destination_port_equal = "80"
    ...
  }

  entry {
    ace_rule_action   = "deny"
    ace_rule_protocol = "ip"
    ...
  }

  # Apply to multiple interfaces
  apply {
    interface = "lan1"
    direction = "in"
    # filter_ids omitted: applies all entries [100, 101]
  }

  apply {
    interface = "lan2"
    direction = "in"
  }

  apply {
    interface = "bridge1"
    direction = "out"
    filter_ids = [100]  # Only apply first entry
  }
}
```

### Requirement 4: Redesign Individual Filter ACLs as Group Resources

**User Story:** As a network administrator, I want to manage IPv4 and IPv6 ACL entries as named groups, so that I have a consistent experience across all ACL types.

#### Acceptance Criteria

1. WHEN a user defines `rtx_access_list_ipv6` or `rtx_access_list_ip` THEN the provider SHALL accept `name` and multiple `entry` blocks (group-based design).
2. WHEN the resource includes `sequence_start` THEN sequences SHALL be auto-calculated; OTHERWISE each entry SHALL require explicit `sequence`.
3. WHEN reading the resource THEN the provider SHALL return all entries belonging to the named group AND any interface applications.
4. WHEN deleting the resource THEN the provider SHALL remove interface applications first, then delete all filter entries.

#### Example Usage

```hcl
# IPv6 ACL with auto sequence
resource "rtx_access_list_ipv6" "internal_v6" {
  name           = "allow_internal_v6"
  sequence_start = 200

  entry {
    action      = "pass"
    source      = "2001:db8::/32"
    destination = "*"
    protocol    = "tcp"
  }

  entry {
    action      = "reject"
    source      = "*"
    destination = "*"
    protocol    = "*"
  }

  apply {
    interface = "lan1"
    direction = "in"
  }
}

# IPv4 ACL with auto sequence
resource "rtx_access_list_ip" "internal_v4" {
  name           = "allow_internal_v4"
  sequence_start = 300

  entry {
    action      = "pass"
    source      = "192.168.0.0/16"
    destination = "*"
    protocol    = "tcp"
  }

  entry {
    action      = "reject"
    source      = "*"
    destination = "*"
    protocol    = "*"
  }

  apply {
    interface = "lan1"
    direction = "in"
  }
}
```

### Requirement 5: Separate Apply Resource for Advanced Use Cases

**User Story:** As an advanced user, I want to manage ACL interface applications separately from ACL definitions, so that I can reuse common ACLs across multiple configurations with fine-grained control.

#### Acceptance Criteria

1. WHEN a user creates `rtx_access_list_ip_apply`, `rtx_access_list_ipv6_apply`, or `rtx_access_list_mac_apply` THEN the provider SHALL apply the referenced ACL to the specified interface.
2. WHEN using `for_each` with an apply resource THEN the provider SHALL create independent apply configurations for each instance.
3. IF both inline `apply` block and separate apply resource target the same interface/direction THEN the provider SHALL return an error during plan.

#### Example Usage

```hcl
resource "rtx_access_list_extended" "common_security" {
  name           = "common_security"
  sequence_start = 500

  entry { ... }
  # No inline apply - managed separately
}

locals {
  acl_targets = {
    "lan1_in" = { interface = "lan1", direction = "in" }
    "lan2_in" = { interface = "lan2", direction = "in" }
    "pp1_in"  = { interface = "pp1", direction = "in" }
  }
}

resource "rtx_access_list_ip_apply" "common" {
  for_each = local.acl_targets

  access_list = rtx_access_list_extended.common_security.name
  interface   = each.value.interface
  direction   = each.value.direction
}
```

### Requirement 6: Remove ACL Attributes from Interface Resources

**User Story:** As a Terraform user, I want a single source of truth for ACL-interface bindings, so that there are no conflicting configurations.

#### Acceptance Criteria

1. WHEN using `rtx_interface` THEN the attributes `access_list_ip_in`, `access_list_ip_out`, `access_list_ipv6_in`, `access_list_ipv6_out`, `access_list_ip_dynamic_in`, `access_list_ip_dynamic_out`, `access_list_ipv6_dynamic_in`, `access_list_ipv6_dynamic_out`, `access_list_mac_in`, `access_list_mac_out` SHALL NOT be available.
2. WHEN using `rtx_pp_interface` THEN the attributes `access_list_ip_in`, `access_list_ip_out` SHALL NOT be available.
3. WHEN using `rtx_ipv6_interface` THEN the ACL-related attributes SHALL NOT be available.
4. WHEN a user attempts to use removed attributes THEN the provider SHALL return a clear error indicating to use ACL resource `apply` blocks instead.

### Requirement 7: Consistent Schema Design Across All ACL Types

**User Story:** As a Terraform user, I want all ACL resources to have the same schema structure, so that I can apply knowledge from one ACL type to another.

#### Acceptance Criteria

1. WHEN defining any ACL resource THEN the user SHALL use the same attribute names for common concepts:
   - `name`: ACL group identifier
   - `sequence_start`: Starting sequence for auto-calculation (optional)
   - `sequence_step`: Increment for auto-calculation (optional, default 1)
   - `entry`: List of ACL entries
   - `apply`: List of interface applications
   - `sequence`: Entry evaluation order (within entry block, required if sequence_start not set)
2. WHEN using the `apply` block THEN all ACL types SHALL accept the same structure:
   - `interface`: Target interface name (e.g., "lan1", "bridge1", "pp1")
   - `direction`: Traffic direction ("in" or "out")
   - `filter_ids`: Optional list of specific filter IDs to apply
3. IF the user specifies invalid values THEN the provider SHALL return consistent error messages across all ACL types.

### Requirement 8: Comprehensive CRUD Testing

**User Story:** As a developer, I want comprehensive tests covering all CRUD operations and edge cases, so that I can ensure the reliability of the ACL implementation.

#### Acceptance Criteria

1. WHEN implementing any ACL resource THEN the following test scenarios SHALL be covered:

**Create Tests:**
- Create ACL with auto sequence mode (sequence_start specified)
- Create ACL with manual sequence mode (entry-level sequence)
- Create ACL with single apply block
- Create ACL with multiple apply blocks
- Create ACL without apply block
- Create ACL with filter_ids specified in apply
- Create ACL with filter_ids omitted in apply (auto-apply all)
- Validation error: sequence specified in auto mode
- Validation error: sequence missing in manual mode
- Validation error: sequence collision with existing filter

**Read Tests:**
- Read ACL and verify all entries returned with correct sequences
- Read ACL and verify apply blocks returned correctly
- Read ACL after router-side changes (drift detection)
- Read non-existent ACL (should remove from state)

**Update Tests:**
- Update ACL: add new entry (auto sequence recalculation)
- Update ACL: remove entry
- Update ACL: modify entry content (same sequence)
- Update ACL: change sequence_start (all sequences recalculated)
- Update ACL: change sequence_step
- Update ACL: add apply block
- Update ACL: remove apply block
- Update ACL: modify apply block (change interface/direction)
- Update ACL: switch from auto to manual mode (and vice versa)

**Delete Tests:**
- Delete ACL with no apply blocks
- Delete ACL with single apply block (removes apply first)
- Delete ACL with multiple apply blocks (removes all applies first)

**Import Tests:**
- Import existing ACL entries as group
- Import ACL with interface applications
- Import and detect sequence mode (auto vs manual)

**Edge Cases:**
- Large number of entries (performance)
- Maximum sequence value handling
- Concurrent modifications to same interface
- Apply to non-existent interface (error handling)

## Non-Functional Requirements

### Code Architecture and Modularity

- **Single Responsibility Principle**: The `apply` logic and sequence calculation should be extracted into shared helpers to avoid code duplication across ACL resources.
- **Modular Design**: Service methods for interface filter application should be reusable across IP, IPv6, MAC, and Dynamic ACL types.
- **Dependency Management**: Minimize changes to existing parser infrastructure; extend rather than rewrite.
- **Clear Interfaces**: Define a common interface for ACL resources that support the unified pattern.

### Performance

- Apply operations should be batched where possible to minimize SSH round-trips.
- Reading ACL state should leverage SFTP cache when available (already implemented for some resources).
- Sequence collision detection should be efficient even with many ACL resources.

### Security

- No changes to existing security model; credentials continue to use Terraform's secure variable mechanism.

### Reliability

- If apply operation fails, the ACL entries should still be created (partial success) with clear error messaging.
- Terraform plan should accurately reflect the expected interface filter state.
- Deleting an ACL with multiple apply blocks should remove all interface applications before deleting entries.
- Sequence collision detection should prevent configuration conflicts.

### Usability

- Documentation should be updated with examples showing both auto and manual sequence modes.
- Migration guide should be provided for users converting from old ACL resources to the new unified design.
- Clear error messages when sequence collisions or `apply` conflicts are detected.
- Recommended sequence range patterns should be documented for multi-ACL scenarios.

### Recommended Sequence Range Patterns

```hcl
# Documentation example: organizing sequence ranges by function
locals {
  sequence_ranges = {
    security_base  = { start = 100,  step = 1 }   # 100-199
    web_services   = { start = 200,  step = 1 }   # 200-299
    database       = { start = 300,  step = 1 }   # 300-399
    monitoring     = { start = 400,  step = 1 }   # 400-499
    default_policy = { start = 9000, step = 1 }   # 9000-9099
  }
}

resource "rtx_access_list_extended" "security_base" {
  name           = "security_base"
  sequence_start = local.sequence_ranges.security_base.start
  sequence_step  = local.sequence_ranges.security_base.step
  entry { ... }
}
```
