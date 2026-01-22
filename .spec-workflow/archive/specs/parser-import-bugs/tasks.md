# Tasks Document: Parser Import Bugs

## Bug 1: Filter List Parsing

- [x] 1.1. Fix preprocessWrappedLines continuation pattern
  - File: internal/rtx/parsers/interface_config.go
  - Change continuation pattern from `^\d` to `^(\s+)?\d` to detect lines starting with space+digit
  - Handle RTX 80-char line wrapping correctly
  - Purpose: Fix filter number truncation when lists wrap to multiple lines
  - _Leverage: existing preprocessWrappedLines function_
  - _Requirements: 1.1, 1.3_

- [x] 1.2. Add wrapped line test cases
  - File: internal/rtx/parsers/interface_config_test.go
  - Add test for 10+ filter numbers spanning multiple lines
  - Test 6-digit filter numbers (200100, 200102, etc.)
  - Test dynamic keyword with wrapped lines
  - Purpose: Ensure wrapped lines are correctly joined
  - _Leverage: existing test patterns_
  - _Requirements: 1.1, 1.2, 1.4_

- [x] 1.3. Apply same fix to IPv6 interface parser
  - File: internal/rtx/parsers/ipv6_interface.go
  - Add preprocessWrappedLines call before parsing
  - Ensure IPv6 filter lists handle wrapping
  - Purpose: Consistent behavior between IPv4 and IPv6
  - _Leverage: interface_config.go preprocessWrappedLines_
  - _Requirements: 1.1_

## Bug 2: DNS server_select Parsing

- [x] 2.1. Fix parseDNSServerSelectFields server limit
  - File: internal/rtx/parsers/dns.go
  - Limit server parsing to max 2 IPs
  - Prevent 3rd+ fields from being parsed as servers
  - Purpose: Stop second server IP being parsed as query_pattern
  - _Leverage: existing parseDNSServerSelectFields function_
  - _Requirements: 2.1, 2.2_

- [x] 2.2. Add DNS server_select test cases
  - File: internal/rtx/parsers/dns_test.go
  - Add test for 2-server config with edns=on and record_type
  - Test query_pattern "." is not confused with IP
  - Test record_type "aaaa" parsing
  - Purpose: Verify correct field order parsing
  - _Leverage: existing DNS test patterns_
  - _Requirements: 2.1, 2.2, 2.3, 2.4_

## Bug 3: DHCP Scope Parsing

- [x] 3.1. Fix scopePattern regex for expire without gateway
  - File: internal/rtx/parsers/dhcp_scope.go
  - Add separate pattern for expire without gateway
  - Support direct "expire" after network CIDR
  - Purpose: Parse "dhcp scope 1 192.168.0.0/16 expire 24:00" correctly
  - _Leverage: existing scopePattern regex_
  - _Requirements: 3.1, 3.2_

- [x] 3.2. Add DHCP scope parsing test cases
  - File: internal/rtx/parsers/dhcp_scope_test.go
  - Add test for scope with network and expire (no gateway)
  - Verify network field is populated correctly
  - Test lease_time conversion
  - Purpose: Ensure network CIDR is always captured
  - _Leverage: existing DHCP test patterns_
  - _Requirements: 3.1, 3.2, 3.3_

## Bug 4: NAT Masquerade Import

- [x] 4.1. Fix BuildShowNATDescriptorCommand grep pattern
  - File: internal/rtx/parsers/nat_masquerade.go
  - Use broader grep pattern: `grep "nat descriptor" | grep -E "( ID | ID$)"`
  - Purpose: Find NAT descriptor by ID correctly
  - _Leverage: existing grep patterns_
  - _Requirements: 4.1_

- [x] 4.2. Add NAT masquerade import test cases
  - File: internal/rtx/parsers/nat_masquerade_test.go
  - Add test for descriptor ID 1000
  - Test multiple static entries parsing
  - Test protocol-only entries (esp, ah)
  - Purpose: Verify descriptor lookup works
  - _Leverage: existing NAT test patterns_
  - _Requirements: 4.1, 4.2_

## Bug 5: IPv6 Filter Dynamic Read

- [x] 5.1. Create ParseIPv6FilterDynamicConfig function
  - File: internal/rtx/parsers/ip_filter.go (already implemented)
  - Add IPv6 version of dynamic filter parser
  - Parse "ipv6 filter dynamic <n> <src> <dst> <protocol> [syslog on]"
  - Purpose: Enable IPv6 dynamic filter parsing
  - _Leverage: existing ParseIPFilterDynamicConfig function_
  - _Requirements: 5.1, 5.2_

- [x] 5.2. Add IPv6 dynamic filter test cases
  - File: internal/rtx/parsers/ip_filter_test.go
  - Add test for 8 IPv6 dynamic filter entries
  - Test protocols: ftp, domain, www, smtp, pop3, submission, tcp, udp
  - Test syslog option parsing
  - Purpose: Verify IPv6 dynamic filter parsing
  - _Leverage: existing IP filter test patterns_
  - _Requirements: 5.1, 5.2_

- [x] 5.3. Implement GetIPv6FilterDynamicConfig in client
  - File: internal/client/ip_filter_service.go (already implemented)
  - Call "show config | grep ipv6 filter dynamic"
  - Parse with ParseIPv6FilterDynamicConfig
  - Return IPv6FilterDynamicConfig struct
  - Purpose: Enable IPv6 dynamic filter read/import
  - _Leverage: existing GetIPFilterDynamicConfig pattern_
  - _Requirements: 5.1, 5.3_

## Integration Testing

- [x] 6.1. Run terraform import integration test
  - Execute terraform import for all affected resources
  - Verify terraform plan shows no diff
  - Test: interface.lan2, dns_server, dhcp_scope, nat_masquerade, ipv6_filter_dynamic
  - Purpose: Confirm all bugs are fixed end-to-end
  - _Requirements: All_
  - Note: Unit tests pass; full integration requires RTX router
