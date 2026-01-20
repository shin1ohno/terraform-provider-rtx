# Tasks Document: Parser Edge Case Tests

## Phase 1: IPsec Edge Cases

- [x] 1. Add IPsec encryption algorithm edge case tests
  - File: internal/rtx/parsers/ipsec_tunnel_test.go
  - Add TestParseIPsecEncryptionAlgorithms with all algorithm variants
  - Test: DES-CBC, 3DES-CBC, AES-128-CBC, AES-256-CBC, AES-GCM-128, AES-GCM-256
  - Purpose: Ensure all encryption algorithms parse correctly
  - _Leverage: internal/rtx/testdata/patterns/ipsec.yaml, existing ipsec_tunnel_test.go patterns_
  - _Requirements: REQ-1.1_
  - _Prompt: Role: Go Test Developer | Task: Implement TestParseIPsecEncryptionAlgorithms covering all encryption algorithm variants from REQ-1.1, using table-driven tests referencing ipsec.yaml patterns | Restrictions: Do not modify parser code, only add tests, maintain test independence | Success: All encryption algorithms tested, tests pass, edge cases documented_

- [x] 2. Add IPsec hash algorithm edge case tests
  - File: internal/rtx/parsers/ipsec_tunnel_test.go
  - Add TestParseIPsecHashAlgorithms with all hash variants
  - Test: MD5, SHA-1, SHA-256, SHA-384, SHA-512
  - Purpose: Ensure all hash algorithms parse correctly
  - _Leverage: internal/rtx/testdata/patterns/ipsec.yaml_
  - _Requirements: REQ-1.1_
  - _Prompt: Role: Go Test Developer | Task: Implement TestParseIPsecHashAlgorithms covering all hash algorithm variants | Restrictions: Table-driven tests only | Success: All hash algorithms tested_

- [x] 3. Add IPsec IKE version edge case tests
  - File: internal/rtx/parsers/ipsec_tunnel_test.go
  - Add TestParseIPsecIKEVersions for IKEv1/IKEv2 configurations
  - Test: Main mode, aggressive mode, IKEv2 specific options
  - Purpose: Ensure IKE version configurations parse correctly
  - _Leverage: internal/rtx/testdata/patterns/ipsec.yaml_
  - _Requirements: REQ-1.2_
  - _Prompt: Role: Go Test Developer | Task: Implement TestParseIPsecIKEVersions for all IKE version configurations from REQ-1.2 | Restrictions: Include both IKEv1 and IKEv2 edge cases | Success: All IKE versions tested_

- [x] 4. Add IPsec NAT traversal edge case tests
  - File: internal/rtx/parsers/ipsec_tunnel_test.go
  - Add TestParseIPsecNATTraversal for NAT-T options
  - Test: NAT-T enabled/disabled, keep-alive intervals, port configurations
  - Purpose: Ensure NAT traversal configurations parse correctly
  - _Leverage: internal/rtx/testdata/patterns/ipsec.yaml_
  - _Requirements: REQ-1.3_
  - _Prompt: Role: Go Test Developer | Task: Implement TestParseIPsecNATTraversal for NAT-T configurations from REQ-1.3 | Success: All NAT-T options tested_

- [x] 5. Add IPsec authentication edge case tests
  - File: internal/rtx/parsers/ipsec_tunnel_test.go
  - Add TestParseIPsecAuthentication for auth methods
  - Test: PSK with special characters, certificate references, XAUTH
  - Purpose: Ensure authentication configurations parse correctly
  - _Leverage: internal/rtx/testdata/patterns/ipsec.yaml_
  - _Requirements: REQ-1.4_
  - _Prompt: Role: Go Test Developer | Task: Implement TestParseIPsecAuthentication for all authentication methods from REQ-1.4 | Restrictions: Test special characters in PSK | Success: All auth methods tested_

## Phase 2: OSPF Edge Cases

- [x] 6. Add OSPF area type edge case tests
  - File: internal/rtx/parsers/ospf_test.go
  - Add TestParseOSPFAreaTypes with all area type variants
  - Test: Normal areas, stub, NSSA, totally stubby
  - Purpose: Ensure all OSPF area types parse correctly
  - _Leverage: internal/rtx/testdata/patterns/ospf.yaml, existing ospf_test.go patterns_
  - _Requirements: REQ-2.1_
  - _Prompt: Role: Go Test Developer | Task: Implement TestParseOSPFAreaTypes covering all area type variants from REQ-2.1 | Success: All area types tested_

- [x] 7. Add OSPF authentication edge case tests
  - File: internal/rtx/parsers/ospf_test.go
  - Add TestParseOSPFAuthentication with all auth options
  - Test: No auth, simple password, MD5 with key-id
  - Purpose: Ensure OSPF authentication parses correctly
  - _Leverage: internal/rtx/testdata/patterns/ospf.yaml_
  - _Requirements: REQ-2.2_
  - _Prompt: Role: Go Test Developer | Task: Implement TestParseOSPFAuthentication for all authentication options from REQ-2.2 | Success: All auth options tested_

- [x] 8. Add OSPF redistribution edge case tests
  - File: internal/rtx/parsers/ospf_test.go
  - Add TestParseOSPFRedistribution with all redistribution options
  - Test: Static, connected, with metrics, metric-type, prefix-lists
  - Purpose: Ensure OSPF redistribution parses correctly
  - _Leverage: internal/rtx/testdata/patterns/ospf.yaml_
  - _Requirements: REQ-2.3_
  - _Prompt: Role: Go Test Developer | Task: Implement TestParseOSPFRedistribution for all redistribution options from REQ-2.3 | Success: All redistribution options tested_

- [x] 9. Add OSPF interface settings edge case tests
  - File: internal/rtx/parsers/ospf_test.go
  - Add TestParseOSPFInterfaceSettings with all interface options
  - Test: Cost, priority, hello/dead intervals, network types
  - Purpose: Ensure OSPF interface settings parse correctly
  - _Leverage: internal/rtx/testdata/patterns/ospf.yaml_
  - _Requirements: REQ-2.4_
  - _Prompt: Role: Go Test Developer | Task: Implement TestParseOSPFInterfaceSettings for all interface options from REQ-2.4 | Success: All interface settings tested_

## Phase 3: NAT Edge Cases

- [x] 10. Add NAT port range edge case tests
  - File: internal/rtx/parsers/nat_masquerade_test.go
  - Add TestParseNATPortRanges with port mapping variants
  - Test: Single port, port ranges (1000-2000), multiple protocols
  - Purpose: Ensure NAT port ranges parse correctly
  - _Leverage: internal/rtx/testdata/patterns/nat.yaml, existing nat_masquerade_test.go patterns_
  - _Requirements: REQ-3.1_
  - _Prompt: Role: Go Test Developer | Task: Implement TestParseNATPortRanges for all port mapping variants from REQ-3.1 | Success: All port range formats tested_

- [x] 11. Add NAT protocol edge case tests
  - File: internal/rtx/parsers/nat_masquerade_test.go
  - Add TestParseNATProtocols with all protocol variants
  - Test: TCP-only, UDP-only, ICMP, any/all protocols
  - Purpose: Ensure NAT protocol handling parses correctly
  - _Leverage: internal/rtx/testdata/patterns/nat.yaml_
  - _Requirements: REQ-3.2_
  - _Prompt: Role: Go Test Developer | Task: Implement TestParseNATProtocols for all protocol variants from REQ-3.2 | Success: All protocols tested_

- [x] 12. Add NAT dynamic pool edge case tests
  - File: internal/rtx/parsers/nat_masquerade_test.go
  - Add TestParseNATDynamic with dynamic NAT options
  - Test: Pool configurations, address ranges, overload (PAT)
  - Purpose: Ensure dynamic NAT configurations parse correctly
  - _Leverage: internal/rtx/testdata/patterns/nat.yaml_
  - _Requirements: REQ-3.3_
  - _Prompt: Role: Go Test Developer | Task: Implement TestParseNATDynamic for all dynamic NAT options from REQ-3.3 | Success: All dynamic NAT options tested_

- [x] 13. Add NAT static edge case tests
  - File: internal/rtx/parsers/nat_static_test.go
  - Add TestParseNATStaticEdgeCases with static NAT variants
  - Test: 1:1 mappings with ports, inside/outside inversions, hairpinning
  - Purpose: Ensure static NAT edge cases parse correctly
  - _Leverage: internal/rtx/testdata/patterns/nat.yaml_
  - _Requirements: REQ-3.4_
  - _Prompt: Role: Go Test Developer | Task: Implement TestParseNATStaticEdgeCases for all static NAT variants from REQ-3.4 | Success: All static NAT edge cases tested_

## Phase 4: BGP Edge Cases

- [x] 14. Add BGP route-map edge case tests
  - File: internal/rtx/parsers/bgp_test.go
  - Add TestParseBGPRouteMap with route-map options
  - Test: In/out route-maps, multiple clauses, match/set actions
  - Purpose: Ensure BGP route-map configurations parse correctly
  - _Leverage: internal/rtx/testdata/patterns/bgp.yaml, existing bgp_test.go patterns_
  - _Requirements: REQ-4.1_
  - _Prompt: Role: Go Test Developer | Task: Implement TestParseBGPRouteMap for all route-map options from REQ-4.1 | Success: All route-map options tested_

- [x] 15. Add BGP 4-byte ASN edge case tests
  - File: internal/rtx/parsers/bgp_test.go
  - Add TestParseBGP4ByteASN with ASN format variants
  - Test: asdot notation (1.65535), asplain notation (65536), private ASN
  - Purpose: Ensure 4-byte ASN formats parse correctly
  - _Leverage: internal/rtx/testdata/patterns/bgp.yaml_
  - _Requirements: REQ-4.2_
  - _Prompt: Role: Go Test Developer | Task: Implement TestParseBGP4ByteASN for all ASN format variants from REQ-4.2 | Success: All ASN formats tested_

- [x] 16. Add BGP community edge case tests
  - File: internal/rtx/parsers/bgp_test.go
  - Add TestParseBGPCommunity with community attribute variants
  - Test: Standard (AA:NN), well-known (no-export, no-advertise), extended
  - Purpose: Ensure BGP community attributes parse correctly
  - _Leverage: internal/rtx/testdata/patterns/bgp.yaml_
  - _Requirements: REQ-4.3_
  - _Prompt: Role: Go Test Developer | Task: Implement TestParseBGPCommunity for all community attribute variants from REQ-4.3 | Success: All community formats tested_

- [x] 17. Add BGP neighbor advanced edge case tests
  - File: internal/rtx/parsers/bgp_test.go
  - Add TestParseBGPNeighborAdvanced with advanced neighbor options
  - Test: eBGP multihop, TTL security, update-source, password with special chars, timers
  - Purpose: Ensure advanced BGP neighbor options parse correctly
  - _Leverage: internal/rtx/testdata/patterns/bgp.yaml_
  - _Requirements: REQ-4.4_
  - _Prompt: Role: Go Test Developer | Task: Implement TestParseBGPNeighborAdvanced for all advanced neighbor options from REQ-4.4 | Success: All advanced options tested_

## Phase 5: Validation and Cleanup

- [x] 18. Run full test suite and verify coverage
  - File: (all test files)
  - Run go test with coverage report
  - Verify all new tests pass
  - Document any discovered parser issues
  - Purpose: Ensure all edge case tests are valid and complete
  - _Leverage: go test -cover, existing test infrastructure_
  - _Requirements: All_
  - _Prompt: Role: QA Engineer | Task: Run full test suite with coverage, verify all edge case tests pass, document results | Success: All tests pass, coverage improved, no parser bugs discovered_

- [x] 19. Update pattern catalogs with edge case examples
  - File: internal/rtx/testdata/patterns/ipsec.yaml, ospf.yaml, nat.yaml, bgp.yaml
  - Add edge case examples to pattern catalogs
  - Document test case to pattern mapping
  - Purpose: Maintain documentation parity with tests
  - _Leverage: Existing pattern catalog format_
  - _Requirements: All_
  - _Prompt: Role: Technical Writer | Task: Update pattern catalogs with edge case examples matching new tests | Success: Catalogs updated with all edge cases documented_
