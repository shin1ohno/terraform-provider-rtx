# Requirements Document: DNS Server Select Schema Refactor

## Introduction

This feature refactors the `rtx_dns_server` resource's `server_select` block to separate mixed concerns (EDNS options, record types, and domain patterns) into distinct schema fields. Currently, the `domains` field contains a mix of metadata (`edns=on`, `any`) and actual domain patterns, which is confusing and doesn't align with Cisco IOS XE conventions.

The refactoring follows the Cisco IOS XE approach where EDNS, domain matching, and server selection are independent configuration concerns.

## Alignment with Product Vision

This feature directly supports the product principle of **Cisco-Compatible Syntax**: aligning resource and attribute naming with Cisco IOS XE Terraform provider conventions. By separating concerns into distinct fields, we:

1. Reduce cognitive load for network administrators familiar with Cisco
2. Enable proper validation of each parameter type
3. Provide clearer, more maintainable configuration

## Requirements

### Requirement 1: Separate EDNS Option

**User Story:** As a network administrator, I want to configure EDNS (Extension mechanisms for DNS) as a dedicated boolean field, so that I can clearly enable/disable EDNS without mixing it with domain patterns.

#### Acceptance Criteria

1. WHEN a `server_select` block is defined THEN the schema SHALL include an `edns` boolean field (optional, default: false)
2. IF `edns = true` THEN the provider SHALL generate `edns=on` in the RTX command
3. IF `edns = false` or omitted THEN the provider SHALL omit the edns parameter (RTX default is off)

### Requirement 2: Separate Record Type Field

**User Story:** As a network administrator, I want to specify the DNS record type as a dedicated field, so that I can clearly define which query types this rule applies to.

#### Acceptance Criteria

1. WHEN a `server_select` block is defined THEN the schema SHALL include a `record_type` string field (optional, default: "a")
2. IF `record_type` is specified THEN the system SHALL validate it is one of: `a`, `aaaa`, `ptr`, `mx`, `ns`, `cname`, `any`
3. WHEN `record_type = "any"` THEN the rule SHALL match all DNS record types
4. IF `record_type` is omitted THEN the system SHALL default to `a` (matching RTX default behavior)

### Requirement 3: Single Query Pattern Field

**User Story:** As a network administrator, I want to specify the domain matching pattern as a single string field, so that the configuration clearly shows which domains this rule targets.

#### Acceptance Criteria

1. WHEN a `server_select` block is defined THEN the schema SHALL include a `query_pattern` string field (required)
2. IF `query_pattern = "."` THEN the rule SHALL match all domain queries
3. IF `query_pattern` starts with `*.` THEN the rule SHALL match suffix patterns (e.g., `*.example.com`)
4. IF `query_pattern` ends with `.*` THEN the rule SHALL match prefix patterns (e.g., `internal.*`)
5. IF `query_pattern` is an exact domain THEN the rule SHALL match only that domain

### Requirement 4: Optional Original Sender Field

**User Story:** As a network administrator, I want to optionally restrict DNS server selection based on the source IP of the query, so that I can apply different DNS servers for different network segments.

#### Acceptance Criteria

1. WHEN a `server_select` block is defined THEN the schema SHALL include an optional `original_sender` string field
2. IF `original_sender` is specified THEN the system SHALL validate it is a valid IPv4/IPv6 address, CIDR notation, or IP range
3. IF `original_sender` is omitted THEN the rule SHALL apply to queries from any source

### Requirement 5: Optional PP Restriction Field

**User Story:** As a network administrator, I want to optionally restrict DNS server selection to when a specific PPP session is active, so that I can have different DNS servers for different WAN connections.

#### Acceptance Criteria

1. WHEN a `server_select` block is defined THEN the schema SHALL include an optional `restrict_pp` integer field
2. IF `restrict_pp > 0` THEN the rule SHALL only be active when the specified PP session is UP
3. IF `restrict_pp = 0` or omitted THEN the rule SHALL apply regardless of PP session state

### Requirement 6: Breaking Change - Remove Legacy `domains` Field

**User Story:** As a maintainer, I want to remove the legacy `domains` field that mixed different concerns, so that the schema is clean and unambiguous.

#### Acceptance Criteria

1. WHEN upgrading to this version THEN the `domains` field SHALL be removed from the schema
2. IF a configuration uses the old `domains` field THEN Terraform SHALL report a schema error with migration guidance
3. WHEN the provider documentation is generated THEN it SHALL include migration instructions from old to new schema

### Requirement 7: RTX Command Generation

**User Story:** As a network administrator, I want the provider to correctly generate RTX CLI commands from the new schema, so that my configuration is applied correctly to the router.

#### Acceptance Criteria

1. WHEN generating a DNS server select command THEN the format SHALL be:
   `dns server select <id> <server> [edns=on] [type] <query-pattern> [original-sender] [restrict pp n]`
2. IF multiple servers are specified THEN they SHALL be included in order (primary, secondary)
3. WHEN parsing RTX output THEN the provider SHALL correctly map CLI values to schema fields

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Parser logic, command generation, and schema definition shall be in separate functions
- **Modular Design**: The DNS parser shall be registered in the parser registry pattern
- **Clear Interfaces**: Client interface shall define clear methods for DNS CRUD operations

### Performance
- No additional performance requirements beyond existing SSH connection constraints

### Security
- No additional security requirements; existing SSH credential handling applies

### Reliability
- Parser shall handle variations in RTX CLI output format across firmware versions
- Command generation shall validate all inputs before sending to router

### Usability
- Schema documentation shall clearly explain each field's purpose and valid values
- Error messages shall provide actionable guidance for configuration mistakes
- Migration guide shall be included in documentation

## Example Configuration

### Before (Legacy Schema)
```hcl
server_select {
  id      = 1
  servers = ["1.1.1.1", "1.0.0.1"]
  domains = ["edns=on", "any", "."]  # Mixed concerns!
}
```

### After (New Schema)
```hcl
server_select {
  id            = 1
  servers       = ["1.1.1.1", "1.0.0.1"]
  edns          = true
  record_type   = "any"
  query_pattern = "."
}
```

### Advanced Example
```hcl
server_select {
  id              = 10
  servers         = ["10.0.0.53"]
  edns            = false
  record_type     = "a"
  query_pattern   = "*.corp.example.com"
  original_sender = "192.168.1.0/24"
  restrict_pp     = 1
}
```
