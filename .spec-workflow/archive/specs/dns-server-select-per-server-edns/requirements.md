# Requirements Document: DNS Server Select Per-Server EDNS

## Introduction

This feature fixes a schema mismatch between the provider layer and parser layer in `rtx_dns_server` resource. The parser layer correctly supports per-server EDNS settings using `DNSServer` struct with `Address` and `EDNS` fields, but the provider layer's Terraform schema uses a flat `servers` string list with a single `edns` boolean, which cannot express per-server EDNS configuration.

RTX router command syntax supports per-server EDNS:
```
dns server select 500000 . a 1.1.1.1 edns=on 1.0.0.1 edns=on
```

The current schema cannot represent this configuration correctly.

## Alignment with Product Vision

Per product.md principles:
- **Follow RTX CLI Semantics**: The RTX router CLI allows per-server EDNS, so the provider should too
- **State Clarity**: The schema should accurately reflect the router's actual configuration to avoid diffs
- **Fail Safely**: Clear validation and error messages for the new schema structure

## Requirements

### Requirement 1: Support Per-Server EDNS in Terraform Schema

**User Story:** As a network administrator, I want to configure EDNS independently for each DNS server in a server_select entry, so that I can match the exact RTX router configuration.

#### Acceptance Criteria

1. WHEN a user defines `server_select` with nested `server` blocks THEN the provider SHALL accept a structure like:
   ```hcl
   server_select {
     id            = 500000
     query_pattern = "."
     record_type   = "a"

     server {
       address = "1.1.1.1"
       edns    = true
     }
     server {
       address = "1.0.0.1"
       edns    = true
     }
   }
   ```

2. IF a `server_select` entry contains more than 2 `server` blocks THEN the provider SHALL return a validation error with message "maximum 2 servers allowed per server_select entry"

3. WHEN a `server_select` entry contains 0 `server` blocks THEN the provider SHALL return a validation error with message "at least one server block is required"

4. IF `edns` is omitted from a `server` block THEN the provider SHALL default to `false`

### Requirement 2: Backward Compatible Migration Path

**User Story:** As a user with existing configurations, I want clear guidance on migrating from the old schema to the new schema, so that I can update my configurations without confusion.

#### Acceptance Criteria

1. WHEN a user attempts to use the old `servers` list attribute THEN Terraform SHALL report "Unsupported argument" error pointing to the new `server` block syntax

2. IF documentation is generated THEN it SHALL include a migration example showing old vs new syntax

### Requirement 3: Correct RTX Command Generation

**User Story:** As a network administrator, I want the provider to generate correct RTX commands with per-server EDNS flags, so that the router configuration matches my Terraform definition.

#### Acceptance Criteria

1. WHEN a server_select has servers with different EDNS settings THEN the provider SHALL generate RTX command with per-server edns flags:
   - Input: `[{address: "1.1.1.1", edns: true}, {address: "1.0.0.1", edns: false}]`
   - Output: `dns server select 1 1.1.1.1 edns=on 1.0.0.1 .`

2. WHEN all servers have EDNS disabled THEN the provider SHALL NOT include any edns flags in the command

3. WHEN reading back router configuration THEN the provider SHALL correctly parse per-server EDNS settings into the `server` blocks

### Requirement 4: Schema Consistency

**User Story:** As a user, I want terraform plan to show no changes when my configuration matches the router's actual state, so that I have confidence in the provider's accuracy.

#### Acceptance Criteria

1. WHEN a router has `dns server select 1 1.1.1.1 edns=on . ` configured AND the Terraform config has matching `server { address = "1.1.1.1", edns = true }` THEN `terraform plan` SHALL show no changes

2. WHEN a router has `dns server select 1 1.1.1.1 . ` (no edns) configured AND the Terraform config has `server { address = "1.1.1.1", edns = false }` THEN `terraform plan` SHALL show no changes

## Non-Functional Requirements

### Code Architecture and Modularity
- **Single Responsibility Principle**: Provider schema changes isolated to `resource_rtx_dns_server.go`
- **Layer Alignment**: Provider schema matches parser layer `DNSServer` struct
- **Clear Interfaces**: Schema-to-parser conversion in service layer (`dns_service.go`)

### Performance
- No performance impact expected; schema change only

### Security
- No security implications; schema change only

### Reliability
- Existing parser tests already cover per-server EDNS parsing
- Provider tests must be updated for new schema structure

### Usability
- Clear error messages for validation failures
- Migration guide in documentation
