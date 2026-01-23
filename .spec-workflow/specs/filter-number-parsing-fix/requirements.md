# Requirements Document: Filter Number Parsing Fix

## Introduction

During terraform plan reconciliation testing, several parsing issues were discovered where filter numbers are being incorrectly split or truncated. This causes false positives in terraform plan output, showing changes that don't actually exist.

## Problem Statement

The parser incorrectly handles filter number lists in several contexts:

1. **IPv4 secure filter parsing**: Numbers like `200100` are being read as `20010`, and `200027` is split into `2000` and `27`
2. **IPv6 dynamic filter parsing**: Numbers like `101085` are split into `1` and `1085`

This appears to be a tokenization or regex issue in the interface configuration parser.

## Requirements

### REQ-1: Fix IPv4 Secure Filter Number Parsing
- **Priority**: P0 (Critical)
- **Description**: The parser must correctly read all filter numbers from `ip <interface> secure filter in/out` commands
- **Acceptance Criteria**:
  - Filter number `200100` is read as `200100`, not `20010`
  - Filter number `200027` is read as `200027`, not split into `2000` and `27`
  - All 6-digit filter numbers are preserved correctly
  - Round-trip test: parse -> generate -> parse produces identical results

### REQ-2: Fix IPv6 Dynamic Filter Number Parsing
- **Priority**: P0 (Critical)
- **Description**: The parser must correctly read all filter numbers from `ipv6 <interface> secure filter out ... dynamic` commands
- **Acceptance Criteria**:
  - Filter number `101085` is read as `101085`, not split into `1` and `1085`
  - All filter numbers in dynamic filter lists are preserved correctly
  - Round-trip test produces identical results

### REQ-3: Add Regression Tests
- **Priority**: P1 (High)
- **Description**: Add unit tests specifically for parsing large filter numbers
- **Acceptance Criteria**:
  - Test case for 6-digit filter numbers in secure_filter_in
  - Test case for 6-digit filter numbers in secure_filter_out
  - Test case for 6-digit filter numbers in dynamic_filter_out
  - Tests cover both IPv4 and IPv6 contexts

## Non-Functional Requirements

### NFR-1: No Breaking Changes
- The fix must not change the behavior for correctly-parsed configurations
- Existing tests must continue to pass

## Out of Scope

- L2TP tunnel_auth_enabled mismatch (main.tf configuration error, not parser issue)
- NAT masquerade state import issues (separate import functionality)
- DHCP scope network replacement (separate state management issue)
