# Tasks Document: Import Fidelity Fix v2

## Phase 1: DHCP Scope IP Range

- [ ] 1.1. Update DHCPScope struct with range fields
  - File: internal/rtx/parsers/dhcp_scope.go
  - Add RangeStart, RangeEnd string fields to DHCPScope struct
  - Purpose: Support IP range format in DHCP scope
  - _Leverage: existing DHCPScope struct_
  - _Requirements: REQ-1_

- [ ] 1.2. Update scope regex for IP range format
  - File: internal/rtx/parsers/dhcp_scope.go
  - Add pattern for `<start_ip>-<end_ip>/<mask>` format
  - Parse: `192.168.1.20-192.168.1.99/16 gateway 192.168.1.253`
  - Purpose: Extract range_start, range_end, netmask, gateway
  - _Leverage: existing scopePattern regex_
  - _Requirements: REQ-1_

- [ ] 1.3. Add DHCP scope IP range test cases
  - File: internal/rtx/parsers/dhcp_scope_test.go
  - Test: `dhcp scope 1 192.168.1.20-192.168.1.99/16 gateway 192.168.1.253 expire 12:00`
  - Verify all fields extracted correctly
  - Purpose: Ensure IP range format works
  - _Leverage: existing DHCP test patterns_
  - _Requirements: REQ-1_

- [ ] 1.4. Update DHCP scope resource schema
  - File: internal/provider/resource_rtx_dhcp_scope.go
  - Add range_start, range_end fields to schema
  - Update Read function to populate range fields
  - Purpose: Expose range fields in Terraform
  - _Leverage: existing resource schema_
  - _Requirements: REQ-1_

## Phase 2: User Attribute Parsing

- [ ] 2.1. Fix login-timer parsing in ParseUserAttributeString
  - File: internal/rtx/parsers/admin.go
  - Ensure `login-timer=3600` extracts as integer 3600
  - Handle variations: `login-timer=0`, `login-timer=300`
  - Purpose: Correct login_timer import
  - _Leverage: existing ParseUserAttributeString function_
  - _Requirements: REQ-2_

- [ ] 2.2. Fix gui-page parsing in ParseUserAttributeString
  - File: internal/rtx/parsers/admin.go
  - Parse `gui-page=dashboard,lan-map,config` as string array
  - Handle empty gui-page and gui-page=none
  - Purpose: Correct gui_pages import
  - _Leverage: existing ParseUserAttributeString function_
  - _Requirements: REQ-3_

- [ ] 2.3. Add user attribute test cases from real config
  - File: internal/rtx/parsers/admin_test.go
  - Test: `user attribute shin1ohno connection=serial,telnet,remote,ssh,sftp,http gui-page=dashboard,lan-map,config login-timer=3600`
  - Verify login_timer=3600, gui_pages=["dashboard","lan-map","config"]
  - Purpose: Verify parsing matches real RTX config
  - _Leverage: existing admin parser tests_
  - _Requirements: REQ-2, REQ-3_

## Phase 3: NAT Masquerade Protocol-Only Entry

- [ ] 3.1. Update MasqueradeStaticEntry for protocol-only
  - File: internal/rtx/parsers/nat_masquerade.go
  - Make port fields optional (use pointers or omitempty)
  - Add IsProtocolOnly helper function
  - Purpose: Support ESP/AH/GRE/ICMP entries
  - _Leverage: existing MasqueradeStaticEntry struct_
  - _Requirements: REQ-4_

- [ ] 3.2. Add protocol-only entry parsing
  - File: internal/rtx/parsers/nat_masquerade.go
  - Parse: `nat descriptor masquerade static 1000 1 192.168.1.253 esp`
  - Detect protocol-only format (no port specification)
  - Purpose: Parse VPN passthrough entries
  - _Leverage: existing static entry parsing_
  - _Requirements: REQ-4_

- [ ] 3.3. Add protocol-only entry command builder
  - File: internal/rtx/parsers/nat_masquerade.go
  - Build: `nat descriptor masquerade static <id> <entry> <ip> <protocol>`
  - Skip port fields for protocol-only entries
  - Purpose: Create VPN passthrough entries
  - _Leverage: existing BuildStaticEntryCommand_
  - _Requirements: REQ-4_

- [ ] 3.4. Add NAT protocol-only test cases
  - File: internal/rtx/parsers/nat_masquerade_test.go
  - Test parsing: ESP, AH, GRE, ICMP protocol entries
  - Test building: protocol-only command format
  - Test round-trip: parse → build → parse
  - Purpose: Verify protocol-only entry support
  - _Leverage: existing NAT test patterns_
  - _Requirements: REQ-4_

- [ ] 3.5. Update NAT masquerade resource schema
  - File: internal/provider/resource_rtx_nat_masquerade.go
  - Make inside_local_port, outside_global_port Optional
  - Add validation: ports required unless protocol is esp/ah/gre/icmp
  - Purpose: Allow protocol-only entries in Terraform
  - _Leverage: existing resource schema_
  - _Requirements: REQ-4_

## Phase 4: Verification

- [ ] 4.1. Update main.tf with correct values
  - File: examples/import/main.tf
  - Update DHCP scope with range fields
  - Update admin_user with login_timer=3600, gui_pages
  - Add NAT ESP entry (entry_number=1)
  - Purpose: Align main.tf with config.txt
  - _Requirements: All_

- [ ] 4.2. Run build and tests
  - Execute: `go build ./... && go test ./internal/rtx/parsers/... -v`
  - Verify all new tests pass
  - Verify no regressions
  - Purpose: Confirm fixes work
  - _Requirements: All_
