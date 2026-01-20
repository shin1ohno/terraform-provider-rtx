# Tasks Document: Import Fidelity Fix

## P0 - Critical Priority

- [x] 1. Fix static route multi-gateway import
  - File: `internal/rtx/parsers/static_route.go`, `internal/rtx/parsers/static_route_test.go`
  - Verify `BuildShowSingleRouteConfigCommand` grep pattern captures all gateway lines
  - Add test case with multi-gateway configuration (2-3 gateways for same prefix/mask)
  - Ensure `ParseRouteConfig` correctly groups all NextHops under same route key
  - Purpose: Ensure all gateways are captured during import, not just the first one
  - _Leverage: `internal/rtx/parsers/static_route.go` (existing ParseRouteConfig logic)_
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec import-fidelity-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex patterns | Task: Fix static route multi-gateway import by verifying grep patterns and parser grouping for REQ-1, leveraging existing ParseRouteConfig logic in internal/rtx/parsers/static_route.go | Restrictions: Do not change the StaticRoute struct, maintain backward compatibility with existing single-gateway routes, ensure parser handles varying whitespace | _Leverage: internal/rtx/parsers/static_route.go for existing pattern matching | _Requirements: REQ-1 (Static Route Multi-Gateway Import) | Success: Test case with 3 gateways passes, all gateways appear in NextHops slice, existing tests still pass | After completing implementation, mark this task as [-] in progress before starting, use log-implementation tool to record what was implemented, then mark as [x] complete_

- [x] 2. Implement DHCP scope import functionality
  - File: `internal/provider/resource_rtx_dhcp_scope.go`, `internal/client/dhcp_scope_service.go`
  - Verify `resourceRtxDhcpScopeImport` function exists and is registered in schema
  - Ensure service `GetScope` queries all related config lines (scope, option, except)
  - Verify parser captures `dhcp scope option N dns=... router=... domain=...` correctly
  - Purpose: Enable terraform import of existing DHCP scope configurations
  - _Leverage: `internal/rtx/parsers/dhcp_scope.go` (existing DHCPScopeParser)_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec import-fidelity-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer with expertise in import state handling | Task: Implement DHCP scope import functionality for REQ-2, ensuring import function is registered and service queries capture all scope configuration | Restrictions: Follow existing resource import patterns from other resources like rtx_static_route, ensure Computed fields are properly set | _Leverage: internal/provider/resource_rtx_static_route.go for import pattern reference | _Requirements: REQ-2 (DHCP Scope Import) | Success: terraform import rtx_dhcp_scope.test 1 populates all fields including options and exclude_ranges, import produces no plan diff | After completing implementation, mark this task as [-] in progress before starting, use log-implementation tool to record what was implemented, then mark as [x] complete_

- [x] 3. Fix NAT masquerade static entries import
  - File: `internal/rtx/parsers/nat_masquerade.go`, `internal/client/nat_masquerade_service.go`, `internal/provider/resource_rtx_nat_masquerade.go`
  - Verify `BuildShowNATDescriptorCommand` grep pattern captures static entry lines
  - Ensure `ParseNATMasqueradeConfig` staticPattern regex matches actual RTX output format
  - Verify resource schema maps `StaticEntries` from parser to Terraform state
  - Add test case with multiple static port mappings
  - Purpose: Ensure NAT static port forwarding rules are captured during import
  - _Leverage: `internal/rtx/parsers/nat_masquerade.go` (existing staticPattern regex)_
  - _Requirements: REQ-4_
  - _Prompt: Implement the task for spec import-fidelity-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with expertise in NAT configuration parsing | Task: Fix NAT masquerade static entries import for REQ-4, verifying grep patterns and ensuring static_entries list is populated | Restrictions: Do not change MasqueradeStaticEntry struct fields, ensure protocol field handles tcp/udp/empty cases | _Leverage: internal/rtx/parsers/nat_masquerade.go for existing static entry parsing | _Requirements: REQ-4 (NAT Masquerade Static Entries Import) | Success: Import of NAT descriptor with 3 static entries shows all entries in state, static entries include correct protocol and port mappings | After completing implementation, mark this task as [-] in progress before starting, use log-implementation tool to record what was implemented, then mark as [x] complete_

## P1 - High Priority

- [x] 4. Fix admin user login_timer import
  - File: `internal/rtx/parsers/admin.go`, `internal/provider/resource_rtx_admin_user.go`
  - Verify `parseUserAttributeString` correctly parses `login-timer=N`
  - Check if schema uses `Default: 0` which overrides parsed value
  - Ensure `login_timer` is Computed or has correct default behavior
  - Add test case with explicit login-timer value
  - Purpose: Ensure login_timer reflects actual router config, not default 0
  - _Leverage: `internal/rtx/parsers/admin.go` (parseUserAttributeString function)_
  - _Requirements: REQ-3_
  - _Prompt: Implement the task for spec import-fidelity-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer specializing in schema design and defaults | Task: Fix admin user login_timer import for REQ-3, ensuring parsed values are not overridden by schema defaults | Restrictions: Do not break existing user creation flows, ensure login_timer=0 is still valid when explicitly set | _Leverage: internal/rtx/parsers/admin.go for existing attribute parsing | _Requirements: REQ-3 (Admin User Attribute Import) | Success: Import of user with login-timer=3600 shows 3600 in state (not 0), user with no explicit timer uses router default | After completing implementation, mark this task as [-] in progress before starting, use log-implementation tool to record what was implemented, then mark as [x] complete_

- [x] 5. Verify DNS server select EDNS parsing
  - File: `internal/rtx/parsers/dns.go`, `internal/rtx/parsers/dns_test.go`
  - Verify `parseDNSServerSelectFields` extracts `edns=on` before domain pattern
  - Add test case with `edns=on` in various positions within the command
  - Ensure EDNS flag is stored in `DNSServerSelect.EDNS` not in `QueryPattern`
  - Purpose: Ensure EDNS flag is correctly captured as boolean, not mixed into domain list
  - _Leverage: `internal/rtx/parsers/dns.go` (parseDNSServerSelectFields function)_
  - _Requirements: REQ-5_
  - _Prompt: Implement the task for spec import-fidelity-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with expertise in DNS configuration parsing | Task: Verify DNS server select EDNS parsing for REQ-5, ensuring edns=on is captured as boolean field | Restrictions: Do not change DNSServerSelect struct, maintain compatibility with existing DNS configurations | _Leverage: internal/rtx/parsers/dns.go for existing parseDNSServerSelectFields logic | _Requirements: REQ-5 (DNS Server Select Parsing) | Success: Test with edns=on before domain passes, EDNS field is true, QueryPattern contains only domain pattern without edns=on string | After completing implementation, mark this task as [-] in progress before starting, use log-implementation tool to record what was implemented, then mark as [x] complete_

## Test Fixtures and Validation

- [x] 6. Create test fixtures for import fidelity tests
  - File: `internal/rtx/testdata/import_fidelity/`
  - Create `static_route_multi_gateway.txt` with sample multi-gateway config
  - Create `dhcp_scope_complete.txt` with scope, options, and except ranges
  - Create `nat_masquerade_with_static.txt` with multiple static entries
  - Create `admin_user_with_timer.txt` with explicit login-timer value
  - Create `dns_server_select_edns.txt` with edns=on flag
  - Purpose: Provide real-world configuration samples for parser testing
  - _Leverage: Existing testdata patterns in `internal/rtx/testdata/`_
  - _Requirements: All (REQ-1 through REQ-5)_
  - _Prompt: Implement the task for spec import-fidelity-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer creating test fixtures for parser validation | Task: Create test fixture files with realistic RTX router configurations covering all requirements | Restrictions: Use realistic RTX command syntax, include edge cases like maximum field counts | _Leverage: Existing testdata files for format reference | _Requirements: REQ-1, REQ-2, REQ-3, REQ-4, REQ-5 | Success: All fixture files created with valid RTX syntax, each fixture covers the specific requirement's edge cases | After completing implementation, mark this task as [-] in progress before starting, use log-implementation tool to record what was implemented, then mark as [x] complete_

- [x] 7. Add parser unit tests using fixtures
  - File: `internal/rtx/parsers/*_test.go`
  - Add test cases that load fixtures from task 6
  - Verify multi-gateway grouping for static routes
  - Verify complete DHCP scope parsing
  - Verify NAT static entries parsing
  - Verify admin login_timer parsing
  - Verify DNS EDNS flag isolation
  - Purpose: Ensure parsers handle all identified edge cases correctly
  - _Leverage: Test fixtures from task 6, existing test patterns_
  - _Requirements: All (REQ-1 through REQ-5)_
  - _Prompt: Implement the task for spec import-fidelity-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer writing comprehensive parser unit tests | Task: Add parser unit tests using created fixtures for all requirements | Restrictions: Use table-driven tests, ensure each test is independent, mock no external dependencies | _Leverage: internal/rtx/parsers/*_test.go for existing test patterns | _Requirements: REQ-1 through REQ-5 | Success: All new tests pass, tests cover both success and edge cases, existing tests still pass | After completing implementation, mark this task as [-] in progress before starting, use log-implementation tool to record what was implemented, then mark as [x] complete_

## Integration Validation

- [x] 8. Validate import round-trip fidelity
  - File: Integration test documentation
  - Document manual testing procedure for import validation
  - Test import → plan should show no changes
  - Verify exported HCL matches router configuration
  - Purpose: Confirm end-to-end import fidelity is achieved
  - _Leverage: Existing acceptance test patterns if available_
  - _Requirements: All_
  - _Prompt: Implement the task for spec import-fidelity-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer validating import functionality | Task: Create integration test documentation and validate import round-trip produces no plan diff | Restrictions: Document steps clearly for reproducibility, note any known limitations | _Leverage: Existing documentation patterns | _Requirements: All requirements validation | Success: Import of each resource type produces no plan diff, documentation is clear and complete | After completing implementation, mark this task as [-] in progress before starting, use log-implementation tool to record what was implemented, then mark as [x] complete_
