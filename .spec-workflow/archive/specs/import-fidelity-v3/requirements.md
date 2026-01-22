# Requirements Document: Import Fidelity Fix v3

## Introduction

This spec addresses remaining import fidelity issues discovered during reconciliation of `config.txt` (actual RTX router configuration) with `examples/import/main.tf` (Terraform configuration). These bugs cause critical discrepancies between the router's actual configuration and what Terraform reads/manages.

**Note**: Some issues in terraform state (DHCP scope, user attribute) are due to stale state from pre-v2 binary. These are already fixed in v2 and will resolve after `terraform refresh`.

## Problem Summary

After implementing import-fidelity-v2, additional issues remain:

1. **Filter List Line-Wrap Truncation**: RTX wraps long lines at ~80 chars, but when a number is split across lines (e.g., `20010\n0` for `200100`), the parser only captures the truncated value
2. **DNS Service Recursive**: `dns service recursive` is not recognized (only `on|off` supported)
3. **DNS Server Select Multi-EDNS**: When each server has its own `edns=on`, parsing fails

## Evidence from Terraform State

| Resource | Expected | Actual (State) | Status |
|----------|----------|----------------|--------|
| lan2.secure_filter_in | 13 values ending `...200099` | 8 values ending `...20010` | **NEW BUG** |
| lan2.secure_filter_out | values ending `...200099` | values ending `...2000` | **NEW BUG** |
| lan2.dynamic_filter_out | `[200080-200085]` | `[]` (empty) | **NEW BUG** |
| ipv6_lan2.dynamic_filter_out | 8 values `[101080-101099]` | 6 values ending `...1` | **NEW BUG** |
| dns_server.service_on | `true` (recursive) | `false` | **NEW BUG** |
| dns_server.server_select[500100] | query_pattern=".", record_type="aaaa" | query_pattern="edns", record_type="a" | **NEW BUG** |
| dhcp_scope.network | `192.168.1.20-192.168.1.99/16` | `null` | Fixed in v2 (stale state) |
| admin_user.login_timer | `3600` | `0` | Fixed in v2 (stale state) |

## Alignment with Product Vision

Import fidelity is critical for:
- Accurate drift detection (terraform plan shows real differences)
- Safe infrastructure-as-code adoption
- Predictable state management

## Requirements

### REQ-1: Filter List Line-Wrap Handling

**User Story:** As a Terraform user, I want filter lists to be parsed correctly even when RTX output wraps long lines, so that all filter numbers are captured.

#### Root Cause

RTX routers wrap long output lines at approximately 80 characters. When a filter number spans the line break (e.g., `200100` becomes `20010\n0`), the current `preprocessWrappedLines` function joins lines but `parseFilterList` only parses complete numbers separated by whitespace.

#### Acceptance Criteria

1. WHEN config contains `ip lan2 secure filter in 200020 200021 200022 200023 200024 200025 200103 20010\n0 200102 200104 200101 200105 200099` THEN parser SHALL reconstruct `200100` and capture all 13 filter numbers

2. WHEN a filter number is split across lines (e.g., `2000\n27`) THEN `preprocessWrappedLines` SHALL detect the split and join the number fragments correctly

3. WHEN parsing `dynamic` keyword followed by filter numbers THEN all dynamic filter numbers SHALL be captured even if line-wrapped

### REQ-2: DNS Service Recursive Support

**User Story:** As a Terraform user, I want `dns service recursive` to be recognized, so that the DNS service state is imported correctly.

#### Root Cause

Current regex pattern: `^\s*dns\s+service\s+(on|off)\s*$`
Does not match: `dns service recursive`

#### Acceptance Criteria

1. WHEN config contains `dns service recursive` THEN parser SHALL set `ServiceOn = true`

2. WHEN config contains `dns service on` THEN parser SHALL set `ServiceOn = true`

3. WHEN config contains `dns service off` THEN parser SHALL set `ServiceOn = false`

4. WHEN building DNS config with ServiceOn=true THEN builder SHALL output `dns service recursive` (preferred form)

### REQ-3: DNS Server Select Multi-Server EDNS Parsing

**User Story:** As a Terraform user, I want DNS server select entries with multiple servers and EDNS options to be parsed correctly, so that all servers and options are captured.

#### Root Cause

Config line: `dns server select 500100 2606:4700:4700::1111 edns=on 2606:4700:4700::1001 edns=on aaaa .`

Current parser:
- Phase 1: Parses first server, stops at `edns=on` (not an IP)
- Phase 2: Captures `edns=on`
- Phase 4: Sets `query_pattern = "2606:4700:4700::1001"` (wrong!)

#### Acceptance Criteria

1. WHEN config contains `dns server select <id> <server1> edns=on <server2> edns=on <record_type> <query_pattern>` THEN parser SHALL extract:
   - servers = [server1, server2]
   - edns = true
   - record_type correctly
   - query_pattern correctly

2. WHEN EDNS option appears after each server THEN parser SHALL handle interleaved format

3. WHEN building DNS server select command THEN output SHALL follow RTX standard format

## Non-Functional Requirements

### Code Architecture and Modularity
- Extend existing parsers without breaking changes
- Maintain backward compatibility with existing configurations
- Follow existing code patterns in `internal/rtx/parsers/`

### Testing
- Add unit tests for each parser fix
- Include tests with actual RTX output samples (line-wrapped)
- Test edge cases and variations

### Reliability
- Graceful handling of malformed input
- Clear error messages for unsupported formats
