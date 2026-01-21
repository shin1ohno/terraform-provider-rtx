# Tasks Document: Filter Enhancements

## Feature 1: Ethernet Filter Interface Application

- [ ] 1. Add parser functions for ethernet filter application
  - File: internal/rtx/parsers/interface.go (or new interface_filter.go)
  - Implement `ParseEthernetFilterApplication()` to parse `ethernet <if> filter in/out` commands
  - Implement `BuildEthernetFilterApplicationCommand()` to generate CLI commands
  - Purpose: Enable parsing and generation of ethernet filter application commands
  - _Leverage: internal/rtx/parsers/ethernet_filter.go, internal/rtx/parsers/interface.go_
  - _Requirements: 1.1, 1.2_

- [ ] 2. Add parser unit tests for ethernet filter application
  - File: internal/rtx/parsers/interface_test.go
  - Test parsing of `ethernet lan1 filter in 1 100` format
  - Test command generation with multiple filter numbers
  - Test edge cases: empty filters, single filter, max filters
  - Purpose: Ensure parser correctness and reliability
  - _Leverage: internal/rtx/parsers/ethernet_filter_test.go patterns_
  - _Requirements: 1.1, 1.2_

- [ ] 3. Extend InterfaceConfig struct with ethernet filter fields
  - File: internal/client/interfaces.go
  - Add `EthernetFilterIn []int` and `EthernetFilterOut []int` fields
  - Purpose: Support ethernet filter data in client layer
  - _Leverage: existing filter fields pattern (SecureFilterIn/Out)_
  - _Requirements: 1.1_

- [ ] 4. Extend InterfaceService to handle ethernet filter commands
  - File: internal/client/interface_service.go
  - Add command generation for `ethernet <if> filter in/out` in Create/Update
  - Add command generation for `no ethernet <if> filter in/out` in Delete/Update
  - Add parsing in GetInterfaceConfig to read existing filter application
  - Purpose: Enable CRUD operations for ethernet filter application
  - _Leverage: existing secure_filter handling patterns_
  - _Requirements: 1.1, 1.2, 1.3_

- [ ] 5. Add ethernet_filter_in/out schema attributes to rtx_interface resource
  - File: internal/provider/resource_rtx_interface.go
  - Add `ethernet_filter_in` and `ethernet_filter_out` schema definitions
  - Add validation: IntBetween(1, 512) for filter numbers
  - Update `buildInterfaceConfigFromResourceData()` to include new fields
  - Update `flattenInterfaceConfigToResourceData()` to read new fields
  - Purpose: Expose ethernet filter application in Terraform schema
  - _Leverage: existing secure_filter_in/out pattern_
  - _Requirements: 1.1, 1.4, 1.5_

- [ ] 6. Add acceptance tests for ethernet filter application
  - File: internal/provider/resource_rtx_interface_test.go
  - Test: Create interface with ethernet_filter_in/out
  - Test: Update ethernet filter application
  - Test: Import interface with existing ethernet filter application
  - Test: Remove ethernet filter application
  - Purpose: Validate end-to-end functionality
  - _Requirements: 1.1, 1.2, 1.3, 1.4_

## Feature 2: IPv6 Dynamic Filter Protocol Extensions

- [ ] 7. Add `submission` protocol to IPv6 dynamic filter validation
  - File: internal/provider/resource_rtx_ipv6_filter_dynamic.go
  - Update line 52: Add "submission" to StringInSlice validator
  - Update description to include submission protocol
  - Purpose: Enable submission protocol in IPv6 dynamic filters
  - _Requirements: 2.1, 2.2_

- [ ] 8. Add acceptance test for IPv6 dynamic filter with submission
  - File: internal/provider/resource_rtx_ipv6_filter_dynamic_test.go
  - Test: Create filter entry with protocol = "submission"
  - Test: Import filter with submission protocol
  - Purpose: Validate submission protocol support
  - _Requirements: 2.1, 2.3_

## Feature 3: Restrict Action Support (Already Implemented)

- [x] 9. Verify restrict action support in IP filters
  - File: internal/rtx/parsers/ip_filter.go (line 44)
  - Status: ALREADY IMPLEMENTED - `restrict`, `restrict-log`, `restrict-nolog` in ValidIPFilterActions
  - Purpose: Confirm existing implementation meets requirements
  - _Requirements: 3.1, 3.3_

- [x] 10. Verify tcpfin/tcprst protocol support
  - File: internal/rtx/parsers/ip_filter.go (line 47)
  - Status: ALREADY IMPLEMENTED - `tcpfin`, `tcprst` in ValidIPFilterProtocols
  - Purpose: Confirm existing implementation meets requirements
  - _Requirements: 3.2, 3.4_

- [x] 11. Verify resource schema support for restrict actions
  - File: internal/provider/resource_rtx_access_list_ip.go (line 38-39)
  - File: internal/provider/resource_rtx_access_list_ipv6.go (line 41-42)
  - Status: ALREADY IMPLEMENTED - Both resources accept restrict/restrict-log
  - Purpose: Confirm existing implementation meets requirements
  - _Requirements: 3.1_

## Documentation

- [ ] 12. Update rtx_interface resource documentation
  - File: docs/resources/interface.md
  - Add documentation for ethernet_filter_in and ethernet_filter_out attributes
  - Add usage examples showing ethernet filter application
  - Purpose: Enable users to discover and use new feature
  - _Requirements: 1.1_

- [ ] 13. Update rtx_ipv6_filter_dynamic resource documentation
  - File: docs/resources/ipv6_filter_dynamic.md
  - Add `submission` to valid protocols list
  - Add example with submission protocol
  - Purpose: Enable users to discover and use submission protocol
  - _Requirements: 2.1_

## Summary

| Task | Feature | Effort | Status |
|------|---------|--------|--------|
| 1-6 | Ethernet Filter Application | Medium | Pending |
| 7-8 | IPv6 submission protocol | Trivial | Pending |
| 9-11 | Restrict Action Support | None | Complete |
| 12-13 | Documentation | Low | Pending |

**Total new implementation tasks**: 8 (Tasks 1-8, 12-13)
**Already complete**: 3 (Tasks 9-11)
