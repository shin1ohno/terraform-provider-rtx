# Requirements Document: DNS Server Select Per-Server EDNS Support

## Introduction

This spec addresses a parsing issue with `dns server select` entries and extends the schema to support per-server EDNS configuration, aligning with RTX router capabilities.

**Current Problem**: `server_select[500100]` shows `query_pattern: "edns"` (should be ".") and `record_type: ""` (should be "aaaa") after import from RTX router.

**Root Cause**: The current schema uses a single `edns` boolean for all servers, but RTX routers support per-server EDNS options. The parser incorrectly handles the interleaved `edns=on` tokens in the command output.

**Solution**: Restructure the schema to support per-server EDNS configuration, which will also fix the parsing issue.

## RTX Command Syntax Reference

Per [Yamaha RTX Manual](https://www.rtpro.yamaha.co.jp/RT/manual/rt-common/dns/dns_server_select.html):

```
dns server select id server [edns=sw] [server2 [edns=sw]] [type] query
```

Each server can have its own `edns=on` or `edns=off` setting independently.

## Problem Evidence

### Current Terraform State (Incorrect)
```json
{
  "id": 500100,
  "servers": ["2606:4700:4700::1111", "2606:4700:4700::1001"],
  "edns": true,
  "query_pattern": "edns",
  "record_type": ""
}
```

### Expected Values
```json
{
  "id": 500100,
  "servers": [
    {"address": "2606:4700:4700::1111", "edns": true},
    {"address": "2606:4700:4700::1001", "edns": true}
  ],
  "query_pattern": ".",
  "record_type": "aaaa"
}
```

## Alignment with Product Vision

Per product.md:
- **Cisco-Compatible Syntax**: Nested blocks for complex configurations
- **Follow RTX CLI Semantics**: Mirror RTX router terminology
- **State Clarity**: Accurately represent router configuration in state

This change enables full fidelity import of DNS server select configurations.

## Requirements

### REQ-1: Per-Server EDNS Schema

**User Story:** As a Terraform user, I want to specify EDNS settings for each DNS server independently, so that I can accurately represent RTX configurations where servers have different EDNS settings.

#### Acceptance Criteria

1. WHEN defining a server_select block THEN user SHALL be able to specify EDNS per server:
   ```hcl
   server_select {
     id            = 500100
     record_type   = "aaaa"
     query_pattern = "."

     server {
       address = "2606:4700:4700::1111"
       edns    = true
     }
     server {
       address = "2606:4700:4700::1001"
       edns    = true
     }
   }
   ```

2. WHEN a server block omits `edns` THEN it SHALL default to `false`

3. WHEN more than 2 server blocks are specified THEN validation SHALL fail with clear error (RTX limit)

4. WHEN building RTX command THEN output SHALL be:
   ```
   dns server select 500100 2606:4700:4700::1111 edns=on 2606:4700:4700::1001 edns=on aaaa .
   ```

### REQ-2: Backward Compatibility (Deprecation)

**User Story:** As a Terraform user with existing configurations, I want my current `servers` and `edns` attributes to continue working during migration, so that I don't have breaking changes.

#### Acceptance Criteria

1. WHEN using deprecated `servers` attribute THEN provider SHALL emit deprecation warning

2. WHEN using deprecated `servers` with `edns=true` THEN all servers SHALL have EDNS enabled

3. WHEN both `servers` (deprecated) and `server` blocks are specified THEN validation SHALL fail

4. IF `servers` is used without `edns` THEN all servers SHALL have EDNS disabled (default)

5. WHEN reading state with old schema THEN provider SHALL convert to new schema format

### REQ-3: Correct Parsing of RTX Output

**User Story:** As a Terraform user, I want DNS server select entries to be parsed correctly from RTX output, so that imported state matches the actual router configuration.

#### Acceptance Criteria

1. WHEN RTX output contains `dns server select <id> <server1> edns=on <server2> edns=on <type> <pattern>` THEN parser SHALL extract:
   - server1 with edns=true
   - server2 with edns=true
   - record_type correctly
   - query_pattern correctly

2. WHEN RTX output contains `dns server select <id> <server1> <server2> edns=on <pattern>` (trailing EDNS) THEN parser SHALL apply EDNS to all servers

3. WHEN RTX output contains `dns server select <id> <server1> edns=off <server2> edns=on <pattern>` THEN parser SHALL correctly capture per-server settings

4. WHEN RTX output is line-wrapped THEN parser SHALL preprocess lines before parsing

5. WHEN running `terraform refresh` after implementation THEN server_select[500100] SHALL show correct values

### REQ-4: Command Building

**User Story:** As a Terraform user, I want the provider to generate correct RTX commands when applying configurations, so that my intended settings are applied to the router.

#### Acceptance Criteria

1. WHEN all servers have `edns=true` THEN command SHALL include `edns=on` after each server

2. WHEN all servers have `edns=false` THEN command SHALL omit `edns=on` options

3. WHEN servers have mixed EDNS settings THEN command SHALL include appropriate option after each server

4. WHEN only one server is specified THEN command SHALL be valid single-server format

## Non-Functional Requirements

### Code Architecture and Modularity
- Update DNSServerSelect struct to use nested Server struct
- Maintain separation between parser, builder, and client layers
- Follow existing code patterns in `internal/rtx/parsers/`

### Migration
- Terraform state upgrade path from old to new schema
- Clear deprecation warnings with migration guidance
- Documentation for migration steps

### Testing
- Unit tests for parser with all EDNS variations
- Unit tests for command builder
- Integration test with actual RTX router
- State migration tests

### Reliability
- Graceful handling of malformed input
- Clear error messages for validation failures
- No data loss during state migration
