# Tasks Document: Parser Pattern Catalogs

## Phase 1: Core Pattern Catalogs

- [x] 1. Create ip_filter.yaml pattern catalog
  - File: internal/rtx/testdata/patterns/ip_filter.yaml
  - Document ip filter, ip filter set, ip filter dynamic commands
  - Include parameter types, ranges, defaults
  - Add examples from RTX documentation
  - Purpose: Document all IP filter command patterns
  - _Leverage: internal/rtx/testdata/patterns/schema.yaml, internal/rtx/parsers/ip_filter.go_
  - _Requirements: REQ-1_
  - _Prompt: Role: Documentation Engineer | Task: Create ip_filter.yaml pattern catalog following schema.yaml format, documenting all IP filter commands from REQ-1 | Restrictions: Follow existing catalog format exactly | Success: Catalog validates against schema, all commands documented_

- [x] 2. Create ethernet_filter.yaml pattern catalog
  - File: internal/rtx/testdata/patterns/ethernet_filter.yaml
  - Document ethernet filter commands with MAC patterns
  - Include VLAN and EtherType filtering
  - Add examples for MAC address formats
  - Purpose: Document all Ethernet filter command patterns
  - _Leverage: internal/rtx/testdata/patterns/schema.yaml, internal/rtx/parsers/ethernet_filter.go_
  - _Requirements: REQ-2_
  - _Prompt: Role: Documentation Engineer | Task: Create ethernet_filter.yaml pattern catalog for all MAC filter commands from REQ-2 | Success: All Ethernet filter commands documented_

- [x] 3. Create bridge.yaml pattern catalog
  - File: internal/rtx/testdata/patterns/bridge.yaml
  - Document bridge member, bridge group, ip bridge commands
  - Include interface type constraints
  - Add examples for L2VPN configurations
  - Purpose: Document all bridge command patterns
  - _Leverage: internal/rtx/testdata/patterns/schema.yaml, internal/rtx/parsers/bridge.go_
  - _Requirements: REQ-3_
  - _Prompt: Role: Documentation Engineer | Task: Create bridge.yaml pattern catalog for all bridge commands from REQ-3 | Success: All bridge commands documented_

- [x] 4. Create service.yaml pattern catalog
  - File: internal/rtx/testdata/patterns/service.yaml
  - Document httpd, sshd, sftpd commands
  - Include host access patterns
  - Add examples for service configurations
  - Purpose: Document all service command patterns
  - _Leverage: internal/rtx/testdata/patterns/schema.yaml, internal/rtx/parsers/service.go_
  - _Requirements: REQ-4_
  - _Prompt: Role: Documentation Engineer | Task: Create service.yaml pattern catalog for HTTPD/SSHD/SFTPD commands from REQ-4 | Success: All service commands documented_

## Phase 2: System Pattern Catalogs

- [x] 5. Create system.yaml pattern catalog
  - File: internal/rtx/testdata/patterns/system.yaml
  - Document timezone, console, statistics, packet-buffer commands
  - Include UTC offset formats and encoding options
  - Add examples for system configurations
  - Purpose: Document all system command patterns
  - _Leverage: internal/rtx/testdata/patterns/schema.yaml, internal/rtx/parsers/system.go_
  - _Requirements: REQ-5_
  - _Prompt: Role: Documentation Engineer | Task: Create system.yaml pattern catalog for all system commands from REQ-5 | Success: All system commands documented_

- [x] 6. Create ipv6_interface.yaml pattern catalog
  - File: internal/rtx/testdata/patterns/ipv6_interface.yaml
  - Document ipv6 address, rtadv, dhcp service, mtu, filter commands
  - Include IPv6 address formats (full, compressed, link-local)
  - Add examples for IPv6 interface configurations
  - Purpose: Document all IPv6 interface command patterns
  - _Leverage: internal/rtx/testdata/patterns/schema.yaml, internal/rtx/parsers/ipv6_interface.go_
  - _Requirements: REQ-6_
  - _Prompt: Role: Documentation Engineer | Task: Create ipv6_interface.yaml pattern catalog for all IPv6 interface commands from REQ-6 | Success: All IPv6 interface commands documented_

- [x] 7. Create ipv6_prefix.yaml pattern catalog
  - File: internal/rtx/testdata/patterns/ipv6_prefix.yaml
  - Document ipv6 prefix, ra-prefix, dhcp-prefix commands
  - Include prefix type variations
  - Add examples for prefix configurations
  - Purpose: Document all IPv6 prefix command patterns
  - _Leverage: internal/rtx/testdata/patterns/schema.yaml, internal/rtx/parsers/ipv6_prefix.go_
  - _Requirements: REQ-7_
  - _Prompt: Role: Documentation Engineer | Task: Create ipv6_prefix.yaml pattern catalog for all IPv6 prefix commands from REQ-7 | Success: All IPv6 prefix commands documented_

## Phase 3: IPsec Transport Tests

- [x] 8. Create ipsec_transport_test.go test file
  - File: internal/rtx/parsers/ipsec_transport_test.go
  - Add TestParseIPsecTransportConfig for transport mode parsing
  - Follow patterns from ipsec_tunnel_test.go
  - Purpose: Add test coverage for IPsec transport parser
  - _Leverage: internal/rtx/parsers/ipsec_tunnel_test.go, internal/rtx/parsers/ipsec_transport.go_
  - _Requirements: REQ-8_
  - _Prompt: Role: Go Test Developer | Task: Create ipsec_transport_test.go with tests for transport mode configurations from REQ-8 | Restrictions: Follow existing ipsec_tunnel_test.go patterns | Success: All transport mode configs tested_

- [x] 9. Add IPsec transport SA configuration tests
  - File: internal/rtx/parsers/ipsec_transport_test.go
  - Add TestParseIPsecTransportSA for SA parsing
  - Test key lifetime, peer address configurations
  - Purpose: Complete IPsec transport test coverage
  - _Leverage: internal/rtx/parsers/ipsec_transport.go_
  - _Requirements: REQ-8_
  - _Prompt: Role: Go Test Developer | Task: Add SA configuration tests to ipsec_transport_test.go from REQ-8 | Success: All SA configs tested_

- [x] 10. Add IPsec transport command builder tests
  - File: internal/rtx/parsers/ipsec_transport_test.go
  - Add TestBuildIPsecTransportCommands for command generation
  - Test round-trip (parse → build → parse)
  - Purpose: Ensure command builder works correctly
  - _Leverage: internal/rtx/parsers/ipsec_transport.go_
  - _Requirements: REQ-8_
  - _Prompt: Role: Go Test Developer | Task: Add command builder tests to ipsec_transport_test.go from REQ-8 | Success: All commands build correctly_

## Phase 4: Validation

- [x] 11. Validate all catalogs against schema
  - File: (all new YAML files)
  - Run schema validation on all new catalogs
  - Fix any schema compliance issues
  - Purpose: Ensure all catalogs are valid
  - _Leverage: internal/rtx/testdata/patterns/schema.yaml_
  - _Requirements: All_
  - _Prompt: Role: QA Engineer | Task: Validate all new pattern catalogs against schema.yaml | Success: All catalogs pass validation_

- [x] 12. Run tests and verify coverage
  - File: internal/rtx/parsers/ipsec_transport_test.go
  - Run go test with coverage
  - Verify ipsec_transport.go has adequate coverage
  - Purpose: Ensure test quality
  - _Leverage: go test -cover_
  - _Requirements: REQ-8_
  - _Prompt: Role: QA Engineer | Task: Run tests and verify coverage for ipsec_transport.go | Success: Tests pass, coverage adequate_
