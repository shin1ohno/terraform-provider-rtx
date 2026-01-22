# Requirements Document: Import Fidelity Fix v2

## Introduction

This spec addresses remaining import fidelity issues discovered during reconciliation of `config.txt` (actual RTX router configuration) with `examples/import/main.tf` (Terraform configuration). These bugs cause discrepancies between the router's actual configuration and what Terraform reads/manages.

## Problem Summary

After fixing parser-import-bugs (5 bugs), additional issues remain:

1. **DHCP Scope IP Range**: IP range format (`192.168.1.20-192.168.1.99/16`) parsed incorrectly
2. **User Attributes**: `login-timer` and `gui-page` values not parsed
3. **NAT Masquerade ESP Entry**: Protocol-only entries without ports not supported

## Alignment with Product Vision

Import fidelity is critical for:
- Accurate drift detection (terraform plan shows real differences)
- Safe infrastructure-as-code adoption
- Predictable state management

## Requirements

### REQ-1: DHCP Scope IP Range Parsing

**User Story:** As a Terraform user, I want DHCP scope IP ranges to be parsed correctly, so that the Terraform state matches the router configuration.

#### Acceptance Criteria

1. WHEN config contains `dhcp scope 1 192.168.1.20-192.168.1.99/16 gateway 192.168.1.253 expire 12:00` THEN parser SHALL extract:
   - `start_ip = "192.168.1.20"`
   - `end_ip = "192.168.1.99"`
   - `netmask = "/16"`
   - `gateway = "192.168.1.253"`
   - `expire = "12:00"`

2. WHEN reading DHCP scope with IP range format THEN resource SHALL populate both `network` and `range` fields appropriately

3. WHEN building DHCP scope command THEN builder SHALL output correct IP range format

### REQ-2: User Attribute login-timer Parsing

**User Story:** As a Terraform user, I want user login_timer values to be imported correctly, so that session timeout settings are preserved.

#### Acceptance Criteria

1. WHEN config contains `user attribute shin1ohno ... login-timer=3600` THEN parser SHALL extract `login_timer = 3600`

2. IF login-timer is not present THEN parser SHALL default to `0` (no timeout)

3. WHEN user has `login-timer=300` AND another user has `login-timer=3600` THEN each user SHALL have their correct value

### REQ-3: User Attribute gui-page Parsing

**User Story:** As a Terraform user, I want user gui_pages to be imported correctly, so that web interface permissions are preserved.

#### Acceptance Criteria

1. WHEN config contains `gui-page=dashboard,lan-map,config` THEN parser SHALL extract `gui_pages = ["dashboard", "lan-map", "config"]`

2. IF gui-page is not present THEN parser SHALL default to empty list

3. WHEN gui-page contains `dashboard,lan-map,config` THEN all three pages SHALL be included in the list

### REQ-4: NAT Masquerade Protocol-Only Entry

**User Story:** As a Terraform user, I want NAT masquerade protocol-only entries (like ESP, AH) to be supported, so that VPN passthrough configurations work correctly.

#### Acceptance Criteria

1. WHEN config contains `nat descriptor masquerade static 1000 1 192.168.1.253 esp` THEN parser SHALL create entry with:
   - `entry_number = 1`
   - `inside_local = "192.168.1.253"`
   - `protocol = "esp"`
   - No port fields (nil/omitted)

2. WHEN building NAT static entry with protocol-only THEN builder SHALL output `<ip> <protocol>` format (without ports)

3. IF protocol is "esp", "ah", "gre", or "icmp" THEN entry SHALL NOT require port fields

## Non-Functional Requirements

### Code Architecture and Modularity
- Extend existing parsers without breaking changes
- Maintain backward compatibility with existing configurations
- Follow existing code patterns in `internal/rtx/parsers/`

### Testing
- Add unit tests for each parser fix
- Include round-trip tests (parse → build → parse)
- Test edge cases and variations

### Reliability
- Graceful handling of malformed input
- Clear error messages for unsupported formats
