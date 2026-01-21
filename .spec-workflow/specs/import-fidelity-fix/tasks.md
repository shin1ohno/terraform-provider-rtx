# Tasks Document: Import Fidelity Fix

## Phase 1: P0 - Critical Parser Fixes

- [x] 1. Fix DNS server_select field parsing order (REQ-1)
  - File: `internal/rtx/parsers/dns.go`
  - Refactor `parseDNSServerSelectFields` to parse fields in correct order:
    1. servers (1-2 IPs) from beginning
    2. `edns=on` if present (literal match)
    3. record_type if in validRecordTypes AND not `.`
    4. query_pattern (required, first non-IP/non-keyword)
    5. original_sender (optional, IP/CIDR after query_pattern)
    6. `restrict pp n` if present
  - Fix: Second server incorrectly parsed as original_sender
  - Fix: record_type defaulting to `a` instead of preserving `aaaa`
  - Fix: query_pattern `.` misinterpreted as other fields
  - Purpose: Ensure DNS forwarding rules are accurately imported
  - _Leverage: `internal/rtx/parsers/dns.go` existing isIPOrCIDR, validRecordTypes_
  - _Requirements: REQ-1_
  - _Completed: 2026-01-21 - Refactored parseDNSServerSelectFields with strict phase-based field order parsing. Added TestParseDNSServerSelectStrictOrder test with 8 test cases._

- [x] 2. Fix Interface secure_filter array truncation (REQ-2)
  - File: `internal/rtx/parsers/interface_config.go`
  - Investigate `parseFilterList` for truncation cause
  - Check SSH output buffer size in `internal/client/`
  - Verify regex captures entire line without truncation
  - Test with 13+ filter IDs to reproduce issue
  - Fix: Filter arrays truncated (e.g., `20010` instead of `200100`)
  - Purpose: Ensure all firewall filter IDs are captured
  - _Leverage: `internal/rtx/parsers/interface_config.go` parseFilterList_
  - _Requirements: REQ-2_
  - _Completed: 2026-01-21 - Root cause was PTY terminal width (40 cols). Fixed in commit bd748e3 by changing all SSH session implementations to use width=512. parseFilterList and regex were correct. Added comprehensive test cases for 13+ filter IDs._

## Phase 2: P1 - High Priority Parser Fixes

- [x] 3. Fix Static route multi-gateway import (REQ-3)
  - File: `internal/rtx/parsers/static_route.go`, `internal/client/static_route_service.go`
  - Verify grep command captures all gateway lines for same prefix
  - Confirm `ParseRouteConfig` groups by routeKey correctly
  - Check if service query filters too narrowly
  - Fix: Only first gateway captured for routes with multiple gateways
  - Purpose: Ensure ECMP/failover routing configurations are preserved
  - _Leverage: `internal/rtx/parsers/static_route.go` existing routeKey grouping_
  - _Requirements: REQ-3_
  - _Completed: 2026-01-20 - Implementation verified correct. Added comprehensive service tests._

- [x] 4. Fix Admin user attribute parsing (REQ-5)
  - File: `internal/rtx/parsers/admin.go`, `internal/provider/resource_rtx_admin_user.go`
  - Verify `parseUserAttributeString` handles `login-timer=N` correctly
  - Check `gui-page=` parsing (comma-separated values)
  - Verify schema Default values don't override parsed values
  - Fix: login_timer=0 instead of actual value, gui_pages empty
  - Purpose: Ensure user permissions are accurately imported
  - _Leverage: `internal/rtx/parsers/admin.go` parseUserAttributeString_
  - _Requirements: REQ-5_
  - _Completed: 2026-01-20 - Verified parser is correct. RTX uses `login-timer=` (hyphen) and `gui-page=` (singular). Added REQ-5 unit tests._

## Phase 3: P2 - Medium Priority Fixes

- [x] 5. Fix L2TP tunnel_auth parsing (REQ-4)
  - File: `internal/rtx/parsers/l2tp.go`
  - Add debug logging to track currentTunnelID during parsing
  - Verify `tunnel select N` sets currentTunnelID correctly
  - Ensure L2TPv3Config initialization at tunnel select
  - Fix l2tpTunnelAuthPattern association with correct tunnel
  - Fix: tunnel_auth_enabled=false when router has `l2tp tunnel auth on`
  - Purpose: Ensure L2TPv3 VPN security settings are preserved
  - _Leverage: `internal/rtx/parsers/l2tp.go` l2tpTunnelAuthPattern_
  - _Requirements: REQ-4_
  - _Completed: 2026-01-20 - Parser was correct. Fixed resourceRTXL2TPImport to set l2tpv3_config including tunnel_auth. Added REQ-4 test cases._

- [x] 6. Relax schema constraints for import compatibility (REQ-6)
  - File: Various `internal/provider/resource_rtx_*.go`
  - Audit schemas for `Required: true` on optional router attributes
  - Change to `Optional: true` where appropriate for import
  - Add `Computed: true` for attributes with router defaults
  - Remove `Default:` that overrides imported values
  - Purpose: Enable import without validation errors
  - _Leverage: Existing resource schemas_
  - _Requirements: REQ-6_
  - _Completed: 2026-01-21 - Changed Default to Computed in rtx_static_route (distance, permanent, filter), rtx_admin_user (encrypted, administrator), and rtx_l2tp (local_router_id, remote_router_id in l2tpv3_config)._

## Phase 4: Testing and Validation

- [ ] 7. Create test fixtures for parser fixes
  - File: `internal/rtx/testdata/import_fidelity/`
  - Create `dns_server_select_multi_server.txt` - two servers, edns, aaaa
  - Create `interface_filter_long_list.txt` - 13+ filter IDs
  - Create `static_route_multi_gateway.txt` - same prefix, 2 gateways
  - Create `l2tp_tunnel_auth.txt` - tunnel auth on with password
  - Create `admin_user_full_attributes.txt` - all attribute types
  - Purpose: Provide test data for parser unit tests
  - _Leverage: Existing testdata patterns_
  - _Requirements: REQ-1 through REQ-5_
  - _Prompt: Implement the task for spec import-fidelity-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer creating test fixtures | Task: Create test fixture files with realistic RTX configurations | Restrictions: Use valid RTX command syntax, include edge cases | _Leverage: internal/rtx/testdata/ | _Requirements: All | Success: Fixture files created with valid syntax | After completing, use log-implementation tool, then mark [x]_

- [ ] 8. Add parser unit tests using fixtures
  - File: `internal/rtx/parsers/*_test.go`
  - Add tests for DNS multi-server and record_type parsing
  - Add tests for interface filter long lists
  - Add tests for static route multi-gateway grouping
  - Add tests for L2TP tunnel auth context
  - Add tests for admin user attributes
  - Purpose: Ensure parser fixes work correctly
  - _Leverage: Test fixtures from task 7_
  - _Requirements: REQ-1 through REQ-5_
  - _Prompt: Implement the task for spec import-fidelity-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer writing parser unit tests | Task: Add parser unit tests using fixtures for all requirements | Restrictions: Table-driven tests, no external mocks | _Leverage: internal/rtx/parsers/*_test.go patterns | _Requirements: All | Success: All tests pass, cover edge cases | After completing, use log-implementation tool, then mark [x]_

- [ ] 9. Validate import round-trip fidelity
  - File: Manual testing / acceptance tests
  - Test: `terraform import` → `terraform plan` shows no changes
  - Verify each fixed resource type imports correctly
  - Document any remaining known limitations
  - Purpose: Confirm end-to-end import fidelity achieved
  - _Leverage: RTX router or simulator_
  - _Requirements: All_
  - _Prompt: Implement the task for spec import-fidelity-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer validating import functionality | Task: Validate import round-trip produces no plan diff | Restrictions: Document steps for reproducibility | _Leverage: Existing acceptance test patterns | _Requirements: All | Success: Import → plan shows no changes for all resource types | After completing, use log-implementation tool, then mark [x]_
