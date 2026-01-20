# Tasks Document: RTX Command Parser Test Coverage

## Phase 1: Foundation

- [x] 1. Define pattern catalog YAML schema
  - File: internal/rtx/testdata/patterns/schema.yaml
  - Create CommandPattern, Parameter, Example structures as YAML schema
  - Document field definitions and validation rules
  - Purpose: Establish consistent format for all command pattern files
  - _Leverage: Design document data models_
  - _Requirements: REQ-1, REQ-8_
  - _Prompt: Role: Go Developer specializing in configuration schemas | Task: Define YAML schema for RTX command patterns following REQ-1 and REQ-8, supporting syntax patterns, parameters with types/ranges, and documentation references | Restrictions: Must be human-readable, support all parameter types from RTX documentation (ipv4, ipv6, int, bool, interface, etc.) | Success: Schema validates all required fields, supports optional parameters, enables test fixture generation_

- [x] 2. Create test fixture directory structure
  - File: internal/rtx/testdata/patterns/, internal/rtx/testdata/fixtures/
  - Create directory hierarchy matching design document
  - Add README explaining organization
  - Purpose: Establish consistent organization for patterns and fixtures
  - _Leverage: Design document Test Fixture Organization section_
  - _Requirements: REQ-8_
  - _Prompt: Role: Go Developer | Task: Create directory structure for pattern catalogs and test fixtures following REQ-8 design | Restrictions: Follow existing testdata conventions, maintain separation between patterns and fixtures | Success: Directories created, README documents the organization_

## Phase 2: High Priority Parsers (Existing Terraform Resources)

- [x] 3. DNS parser pattern extraction (Chapter 24)
  - File: internal/rtx/testdata/patterns/dns.yaml
  - Extract patterns: dns server, dns server select, dns domain, dns static, dns private address, dns notice, dns syslog, dns cache
  - Include examples from documentation 設定例 sections
  - Purpose: Create comprehensive DNS command pattern catalog
  - _Leverage: docs/RTX-commands/24*.md, internal/rtx/parsers/dns.go_
  - _Requirements: REQ-1, REQ-2, REQ-3_
  - _Prompt: Role: Go Developer with RTX expertise | Task: Extract all DNS command patterns from Chapter 24 documentation for REQ-1, categorize for REQ-2, map to existing dns.go parser | Restrictions: Include all format variations and no-form commands, capture parameter ranges from 設定値 sections | Success: All DNS commands cataloged with examples, tagged with parser name_

- [x] 4. DNS parser test fixture generation and validation
  - File: internal/rtx/testdata/fixtures/dns/*.txt, internal/rtx/parsers/dns_test.go
  - Generate test fixtures from dns.yaml patterns
  - Add table-driven tests covering all patterns
  - Run tests and fix any parser gaps
  - Purpose: Achieve 100% DNS command coverage
  - _Leverage: internal/rtx/testdata/patterns/dns.yaml, existing dns_test.go_
  - _Requirements: REQ-3, REQ-4, REQ-5, REQ-6_
  - _Prompt: Role: Go Test Developer | Task: Generate test fixtures from DNS patterns for REQ-3, verify against existing parser for REQ-4, include interface variations for REQ-5 and value types for REQ-6 | Restrictions: Use table-driven tests, maintain existing test patterns, fix any parser failures | Success: All DNS patterns tested, failures fixed, coverage report shows 100% for DNS category_

- [x] 5. Static Route parser pattern extraction
  - File: internal/rtx/testdata/patterns/static_route.yaml
  - Extract patterns: ip route, ipv6 route with all gateway types
  - Include patterns for pp, tunnel, null gateway types
  - Purpose: Create comprehensive static route command pattern catalog
  - _Leverage: docs/RTX-commands/8*.md (IP routing), internal/rtx/parsers/static_route.go_
  - _Requirements: REQ-1, REQ-2, REQ-5_
  - _Prompt: Role: Go Developer with RTX expertise | Task: Extract all static route patterns from Chapter 8 documentation, including gateway variations (gateway IP, pp, tunnel, null interface) for REQ-5 | Restrictions: Cover all metric, filter, hide options | Success: All ip route/ipv6 route variations cataloged_

- [x] 6. Static Route parser test fixture generation and validation
  - File: internal/rtx/testdata/fixtures/static_route/*.txt, internal/rtx/parsers/static_route_test.go
  - Generate test fixtures from static_route.yaml patterns
  - Add tests for gateway type variations and optional parameters
  - Purpose: Achieve 100% static route command coverage
  - _Leverage: internal/rtx/testdata/patterns/static_route.yaml, existing static_route_test.go_
  - _Requirements: REQ-3, REQ-4, REQ-5, REQ-6_
  - _Prompt: Role: Go Test Developer | Task: Generate comprehensive test fixtures for static route patterns, test all gateway types and interface naming conventions | Restrictions: Fix any parser gaps discovered | Success: All static route patterns tested and passing_

- [x] 7. VLAN parser pattern extraction (Chapter 38)
  - File: internal/rtx/testdata/patterns/vlan.yaml
  - Extract patterns: vlan lan1/*, switch control function vlan, vlan port mapping
  - Purpose: Create comprehensive VLAN command pattern catalog
  - _Leverage: docs/RTX-commands/38*.md, internal/rtx/parsers/vlan.go_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Role: Go Developer with RTX expertise | Task: Extract all VLAN command patterns from Chapter 38 documentation, including tagged VLAN interface naming | Restrictions: Cover all VLAN-related commands | Success: All VLAN commands cataloged_

- [x] 8. VLAN parser test fixture generation and validation
  - File: internal/rtx/testdata/fixtures/vlan/*.txt, internal/rtx/parsers/vlan_test.go
  - Generate test fixtures from vlan.yaml patterns
  - Test tagged VLAN interface naming (lan1/1, lan1/2)
  - Purpose: Achieve 100% VLAN command coverage
  - _Leverage: internal/rtx/testdata/patterns/vlan.yaml, existing vlan_test.go_
  - _Requirements: REQ-3, REQ-4, REQ-5_
  - _Prompt: Role: Go Test Developer | Task: Generate VLAN test fixtures, verify tagged VLAN interface patterns | Restrictions: Fix any parser gaps | Success: All VLAN patterns tested and passing_

- [x] 9. Interface parser pattern extraction (Chapter 8)
  - File: internal/rtx/testdata/patterns/interface.yaml
  - Extract patterns: ip lan* address, ip pp address, ip tunnel address, description
  - Include LAN division (lan1.1) and all interface types
  - Purpose: Create comprehensive interface command pattern catalog
  - _Leverage: docs/RTX-commands/8*.md, internal/rtx/parsers/interface_config.go_
  - _Requirements: REQ-1, REQ-2, REQ-5_
  - _Prompt: Role: Go Developer with RTX expertise | Task: Extract all interface configuration patterns from Chapter 8, covering all interface naming conventions from REQ-5 | Restrictions: Include physical LAN, LAN division, PP, tunnel, loopback, bridge interfaces | Success: All interface commands cataloged_

- [x] 10. Interface parser test fixture generation and validation
  - File: internal/rtx/testdata/fixtures/interface/*.txt, internal/rtx/parsers/interface_config_test.go
  - Generate test fixtures from interface.yaml patterns
  - Test all interface naming conventions
  - Purpose: Achieve 100% interface command coverage
  - _Leverage: internal/rtx/testdata/patterns/interface.yaml, existing interface_config_test.go_
  - _Requirements: REQ-3, REQ-4, REQ-5, REQ-6_
  - _Prompt: Role: Go Test Developer | Task: Generate interface test fixtures, verify all interface types and IP address formats | Restrictions: Fix any parser gaps | Success: All interface patterns tested and passing_

- [x] 11. IPsec parser pattern extraction (Chapter 15)
  - File: internal/rtx/testdata/patterns/ipsec.yaml
  - Extract patterns: ipsec ike, ipsec sa, ipsec tunnel, tunnel enable/disable
  - Include all encryption/hash algorithm options
  - Purpose: Create comprehensive IPsec command pattern catalog
  - _Leverage: docs/RTX-commands/15*.md, internal/rtx/parsers/ipsec_tunnel.go_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Role: Go Developer with RTX expertise | Task: Extract all IPsec/IKE command patterns from Chapter 15 | Restrictions: Cover all algorithm variations and tunnel configurations | Success: All IPsec commands cataloged_

- [x] 12. IPsec parser test fixture generation and validation
  - File: internal/rtx/testdata/fixtures/ipsec/*.txt, internal/rtx/parsers/ipsec_tunnel_test.go
  - Generate test fixtures from ipsec.yaml patterns
  - Test encryption algorithm variations
  - Purpose: Achieve 100% IPsec command coverage
  - _Leverage: internal/rtx/testdata/patterns/ipsec.yaml, existing ipsec_tunnel_test.go_
  - _Requirements: REQ-3, REQ-4_
  - _Prompt: Role: Go Test Developer | Task: Generate IPsec test fixtures, verify all encryption options | Restrictions: Fix any parser gaps | Success: All IPsec patterns tested and passing_

- [x] 13. OSPF parser pattern extraction (Chapter 28)
  - File: internal/rtx/testdata/patterns/ospf.yaml
  - Extract patterns: ospf use, ospf area, ospf router id, ospf network, ospf import
  - Include all area types and authentication options
  - Purpose: Create comprehensive OSPF command pattern catalog
  - _Leverage: docs/RTX-commands/28*.md, internal/rtx/parsers/ospf.go_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Role: Go Developer with RTX expertise | Task: Extract all OSPF command patterns from Chapter 28 | Restrictions: Cover all area types, authentication modes, and import options | Success: All OSPF commands cataloged_

- [x] 14. OSPF parser test fixture generation and validation
  - File: internal/rtx/testdata/fixtures/ospf/*.txt, internal/rtx/parsers/ospf_test.go
  - Generate test fixtures from ospf.yaml patterns
  - Test area type variations and authentication
  - Purpose: Achieve 100% OSPF command coverage
  - _Leverage: internal/rtx/testdata/patterns/ospf.yaml, existing ospf_test.go_
  - _Requirements: REQ-3, REQ-4_
  - _Prompt: Role: Go Test Developer | Task: Generate OSPF test fixtures, verify all configurations | Restrictions: Fix any parser gaps | Success: All OSPF patterns tested and passing_

- [x] 15. BGP parser pattern extraction (Chapter 29)
  - File: internal/rtx/testdata/patterns/bgp.yaml
  - Extract patterns: bgp use, bgp router id, bgp autonomous-system, bgp neighbor, bgp import, bgp export
  - Include all AS path and community options
  - Purpose: Create comprehensive BGP command pattern catalog
  - _Leverage: docs/RTX-commands/29*.md, internal/rtx/parsers/bgp.go_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Role: Go Developer with RTX expertise | Task: Extract all BGP command patterns from Chapter 29 | Restrictions: Cover all neighbor configurations and route policy options | Success: All BGP commands cataloged_

- [x] 16. BGP parser test fixture generation and validation
  - File: internal/rtx/testdata/fixtures/bgp/*.txt, internal/rtx/parsers/bgp_test.go
  - Generate test fixtures from bgp.yaml patterns
  - Test AS number and neighbor configurations
  - Purpose: Achieve 100% BGP command coverage
  - _Leverage: internal/rtx/testdata/patterns/bgp.yaml, existing bgp_test.go_
  - _Requirements: REQ-3, REQ-4_
  - _Prompt: Role: Go Test Developer | Task: Generate BGP test fixtures, verify all configurations | Restrictions: Fix any parser gaps | Success: All BGP patterns tested and passing_

- [x] 17. NAT parser pattern extraction (Chapter 23)
  - File: internal/rtx/testdata/patterns/nat.yaml
  - Extract patterns: nat descriptor type, nat descriptor address outer/inner, nat descriptor masquerade static
  - Include all NAT types (masquerade, static, dynamic)
  - Purpose: Create comprehensive NAT command pattern catalog
  - _Leverage: docs/RTX-commands/23*.md, internal/rtx/parsers/nat_masquerade.go, nat_static.go_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Role: Go Developer with RTX expertise | Task: Extract all NAT command patterns from Chapter 23 | Restrictions: Cover all NAT types and port mapping options | Success: All NAT commands cataloged_

- [x] 18. NAT parser test fixture generation and validation
  - File: internal/rtx/testdata/fixtures/nat/*.txt, internal/rtx/parsers/nat_*_test.go
  - Generate test fixtures from nat.yaml patterns
  - Test static, masquerade, and dynamic NAT configurations
  - Purpose: Achieve 100% NAT command coverage
  - _Leverage: internal/rtx/testdata/patterns/nat.yaml, existing nat_*_test.go_
  - _Requirements: REQ-3, REQ-4, REQ-6_
  - _Prompt: Role: Go Test Developer | Task: Generate NAT test fixtures, verify all configurations | Restrictions: Fix any parser gaps | Success: All NAT patterns tested and passing_

## Phase 3: Medium Priority Parsers

- [x] 19. DHCP parser pattern extraction (Chapter 12)
  - File: internal/rtx/testdata/patterns/dhcp.yaml
  - Extract patterns: dhcp service, dhcp scope, dhcp scope bind, dhcp scope option
  - Purpose: Create comprehensive DHCP command pattern catalog
  - _Leverage: docs/RTX-commands/12*.md, internal/rtx/parsers/dhcp_scope.go, dhcp_bindings.go_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Role: Go Developer with RTX expertise | Task: Extract all DHCP command patterns from Chapter 12 | Restrictions: Cover all scope options and binding configurations | Success: All DHCP commands cataloged_

- [x] 20. DHCP parser test fixture generation and validation
  - File: internal/rtx/testdata/fixtures/dhcp/*.txt, internal/rtx/parsers/dhcp_*_test.go
  - Generate test fixtures from dhcp.yaml patterns
  - Purpose: Achieve 100% DHCP command coverage
  - _Leverage: internal/rtx/testdata/patterns/dhcp.yaml_
  - _Requirements: REQ-3, REQ-4_
  - _Prompt: Role: Go Test Developer | Task: Generate DHCP test fixtures | Success: All DHCP patterns tested and passing_

- [x] 21. L2TP parser pattern extraction (Chapter 16)
  - File: internal/rtx/testdata/patterns/l2tp.yaml
  - Extract patterns: l2tp service, pp auth, tunnel enable
  - Purpose: Create comprehensive L2TP command pattern catalog
  - _Leverage: docs/RTX-commands/16*.md, internal/rtx/parsers/l2tp.go_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Role: Go Developer with RTX expertise | Task: Extract all L2TP command patterns from Chapter 16 | Success: All L2TP commands cataloged_

- [x] 22. L2TP parser test fixture generation and validation
  - File: internal/rtx/testdata/fixtures/l2tp/*.txt, internal/rtx/parsers/l2tp_test.go
  - Generate test fixtures from l2tp.yaml patterns
  - Purpose: Achieve 100% L2TP command coverage
  - _Leverage: internal/rtx/testdata/patterns/l2tp.yaml_
  - _Requirements: REQ-3, REQ-4_
  - _Prompt: Role: Go Test Developer | Task: Generate L2TP test fixtures | Success: All L2TP patterns tested and passing_

- [x] 23. PPTP parser pattern extraction (Chapter 17)
  - File: internal/rtx/testdata/patterns/pptp.yaml
  - Extract patterns: pptp service, pptp hostname
  - Purpose: Create comprehensive PPTP command pattern catalog
  - _Leverage: docs/RTX-commands/17*.md, internal/rtx/parsers/pptp.go_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Role: Go Developer with RTX expertise | Task: Extract all PPTP command patterns from Chapter 17 | Success: All PPTP commands cataloged_

- [x] 24. PPTP parser test fixture generation and validation
  - File: internal/rtx/testdata/fixtures/pptp/*.txt, internal/rtx/parsers/pptp_test.go
  - Generate test fixtures from pptp.yaml patterns
  - Purpose: Achieve 100% PPTP command coverage
  - _Leverage: internal/rtx/testdata/patterns/pptp.yaml_
  - _Requirements: REQ-3, REQ-4_
  - _Prompt: Role: Go Test Developer | Task: Generate PPTP test fixtures | Success: All PPTP patterns tested and passing_

- [x] 25. Schedule parser pattern extraction (Chapter 37)
  - File: internal/rtx/testdata/patterns/schedule.yaml
  - Extract patterns: schedule at, schedule cron
  - Purpose: Create comprehensive schedule command pattern catalog
  - _Leverage: docs/RTX-commands/37*.md, internal/rtx/parsers/schedule.go_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Role: Go Developer with RTX expertise | Task: Extract all schedule command patterns from Chapter 37 | Success: All schedule commands cataloged_

- [x] 26. Schedule parser test fixture generation and validation
  - File: internal/rtx/testdata/fixtures/schedule/*.txt, internal/rtx/parsers/schedule_test.go
  - Generate test fixtures from schedule.yaml patterns
  - Purpose: Achieve 100% schedule command coverage
  - _Leverage: internal/rtx/testdata/patterns/schedule.yaml_
  - _Requirements: REQ-3, REQ-4_
  - _Prompt: Role: Go Test Developer | Task: Generate schedule test fixtures | Success: All schedule patterns tested and passing_

- [x] 27. Admin parser pattern extraction (Chapter 4)
  - File: internal/rtx/testdata/patterns/admin.yaml
  - Extract patterns: login user, administrator, login password
  - Purpose: Create comprehensive admin command pattern catalog
  - _Leverage: docs/RTX-commands/4*.md, internal/rtx/parsers/admin.go_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Role: Go Developer with RTX expertise | Task: Extract all admin/user command patterns from Chapter 4 | Success: All admin commands cataloged_

- [x] 28. Admin parser test fixture generation and validation
  - File: internal/rtx/testdata/fixtures/admin/*.txt, internal/rtx/parsers/admin_test.go
  - Generate test fixtures from admin.yaml patterns
  - Purpose: Achieve 100% admin command coverage
  - _Leverage: internal/rtx/testdata/patterns/admin.yaml_
  - _Requirements: REQ-3, REQ-4_
  - _Prompt: Role: Go Test Developer | Task: Generate admin test fixtures | Success: All admin patterns tested and passing_

## Phase 4: Low Priority Parsers

- [x] 29. SNMP parser pattern extraction (Chapter 21)
  - File: internal/rtx/testdata/patterns/snmp.yaml
  - Extract patterns: snmpv2c host, snmpv2c community, snmpv2c trap host
  - Purpose: Create comprehensive SNMP command pattern catalog
  - _Leverage: docs/RTX-commands/21*.md, internal/rtx/parsers/snmp.go_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Role: Go Developer with RTX expertise | Task: Extract all SNMP command patterns from Chapter 21 | Success: All SNMP commands cataloged_

- [x] 30. SNMP parser test fixture generation and validation
  - File: internal/rtx/testdata/fixtures/snmp/*.txt, internal/rtx/parsers/snmp_test.go
  - Generate test fixtures from snmp.yaml patterns
  - Purpose: Achieve 100% SNMP command coverage
  - _Leverage: internal/rtx/testdata/patterns/snmp.yaml_
  - _Requirements: REQ-3, REQ-4_
  - _Prompt: Role: Go Test Developer | Task: Generate SNMP test fixtures | Success: All SNMP patterns tested and passing_

- [x] 31. Syslog parser pattern extraction (Chapter 59)
  - File: internal/rtx/testdata/patterns/syslog.yaml
  - Extract patterns: syslog host, syslog facility, syslog notice
  - Purpose: Create comprehensive syslog command pattern catalog
  - _Leverage: docs/RTX-commands/59*.md, internal/rtx/parsers/syslog.go_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Role: Go Developer with RTX expertise | Task: Extract all syslog command patterns from Chapter 59 | Success: All syslog commands cataloged_

- [x] 32. Syslog parser test fixture generation and validation
  - File: internal/rtx/testdata/fixtures/syslog/*.txt, internal/rtx/parsers/syslog_test.go
  - Generate test fixtures from syslog.yaml patterns
  - Purpose: Achieve 100% syslog command coverage
  - _Leverage: internal/rtx/testdata/patterns/syslog.yaml_
  - _Requirements: REQ-3, REQ-4_
  - _Prompt: Role: Go Test Developer | Task: Generate syslog test fixtures | Success: All syslog patterns tested and passing_

- [x] 33. QoS parser pattern extraction (Chapter 26)
  - File: internal/rtx/testdata/patterns/qos.yaml
  - Extract patterns: queue class filter, queue lan* type, speed lan*
  - Purpose: Create comprehensive QoS command pattern catalog
  - _Leverage: docs/RTX-commands/26*.md, internal/rtx/parsers/qos.go_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Role: Go Developer with RTX expertise | Task: Extract all QoS command patterns from Chapter 26 | Success: All QoS commands cataloged_

- [x] 34. QoS parser test fixture generation and validation
  - File: internal/rtx/testdata/fixtures/qos/*.txt, internal/rtx/parsers/qos_test.go
  - Generate test fixtures from qos.yaml patterns
  - Purpose: Achieve 100% QoS command coverage
  - _Leverage: internal/rtx/testdata/patterns/qos.yaml_
  - _Requirements: REQ-3, REQ-4_
  - _Prompt: Role: Go Test Developer | Task: Generate QoS test fixtures | Success: All QoS patterns tested and passing_

## Phase 5: Coverage Analysis and Reporting

- [x] 35. Generate comprehensive coverage report
  - File: docs/coverage-report.md
  - Aggregate test results from all parsers
  - Calculate coverage percentages by category
  - List patterns without parser coverage
  - Purpose: Identify remaining gaps and prioritize future work
  - _Leverage: All pattern catalogs, test results_
  - _Requirements: REQ-7_
  - _Prompt: Role: QA Engineer | Task: Generate coverage report following REQ-7, showing pass/fail/skip counts by category, identifying commands without parser implementation | Restrictions: Include prioritization based on Terraform resource usage | Success: Report shows coverage percentages, gaps are prioritized_

- [x] 36. Create gap analysis and improvement roadmap
  - File: docs/parser-improvement-roadmap.md
  - Document commands with partial/no parser coverage
  - Prioritize based on Terraform resource usage
  - Estimate effort for each gap
  - Purpose: Guide future parser development
  - _Leverage: docs/coverage-report.md, existing Terraform resources_
  - _Requirements: REQ-7_
  - _Prompt: Role: Technical Lead | Task: Create roadmap for parser improvements based on gap analysis, prioritizing commands used by existing Terraform resources | Success: Clear roadmap with prioritized items_
