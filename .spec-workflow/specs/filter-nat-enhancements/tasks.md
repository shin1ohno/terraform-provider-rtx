# Tasks Document: Filter & NAT Enhancements

## Phase 1: IP Filter Extensions (REQ-1, REQ-2)

- [x] 1. Extend IP filter action validation list
  - File: internal/rtx/parsers/ip_filter.go
  - Add `restrict-nolog` to `ValidIPFilterActions` slice
  - Purpose: Enable restrict action without logging for IP filters
  - _Leverage: Existing ValidIPFilterActions pattern_
  - _Requirements: REQ-1_

- [x] 2. Extend IP filter protocol validation list
  - File: internal/rtx/parsers/ip_filter.go
  - Add `tcpfin`, `tcprst`, `tcpsyn`, `established` to `ValidIPFilterProtocols` slice
  - Purpose: Enable TCP flag-based filtering protocols
  - _Leverage: Existing ValidIPFilterProtocols pattern_
  - _Requirements: REQ-2_

- [x] 3. Add parser tests for new actions and protocols
  - File: internal/rtx/parsers/ip_filter_test.go
  - Add test cases for `restrict-nolog` action parsing
  - Add test cases for `tcpfin`, `tcprst`, `tcpsyn`, `established` protocol parsing
  - Purpose: Ensure parser correctly handles new action and protocol values
  - _Leverage: Existing test patterns in ip_filter_test.go_
  - _Requirements: REQ-1, REQ-2_

- [x] 4. Update resource schema validation for rtx_access_list_ip
  - File: internal/provider/resource_rtx_access_list_ip.go
  - Update action enum validation to include `restrict-nolog`
  - Update protocol enum validation to include TCP flag protocols
  - Purpose: Allow Terraform users to specify new values in HCL
  - _Leverage: Existing schema validation patterns_
  - _Requirements: REQ-1, REQ-2_

## Phase 2: Dynamic IP Filter Resource (REQ-3)

- [x] 5. Extend ValidDynamicProtocols list
  - File: internal/rtx/parsers/ip_filter.go
  - Add missing services: tftp, submission, https, imap, ldap, bgp, sip, ipsec-nat-t, etc.
  - Purpose: Support all RTX dynamic filter protocols
  - _Leverage: Existing ValidDynamicProtocols slice_
  - _Requirements: REQ-3_

- [x] 6. Extend IPFilterDynamic struct for filter-reference form
  - File: internal/rtx/parsers/ip_filter.go
  - Add fields: FilterList, InFilterList, OutFilterList []int
  - Add Timeout *int field for optional timeout parameter
  - Purpose: Support both dynamic filter syntax forms
  - _Leverage: Existing IPFilterDynamic struct_
  - _Requirements: REQ-3_

- [x] 7. Create extended dynamic filter parser function
  - File: internal/rtx/parsers/ip_filter.go
  - Implement `ParseIPFilterDynamicConfigExtended()` to handle both forms
  - Handle Form 1: `ip filter dynamic <id> <src> <dst> <protocol> [options]`
  - Handle Form 2: `ip filter dynamic <id> <src> <dst> filter <list> [in <list>] [out <list>] [options]`
  - Purpose: Parse all dynamic filter configurations from router output
  - _Leverage: Existing ParseIPFilterDynamicConfig function_
  - _Requirements: REQ-3_

- [x] 8. Create extended dynamic filter command builder
  - File: internal/rtx/parsers/ip_filter.go
  - Implement `BuildIPFilterDynamicCommandExtended()` for both forms
  - Handle syslog=on/off and timeout=N options
  - Purpose: Generate correct RTX commands for dynamic filters
  - _Leverage: Existing BuildIPFilterDynamicCommand function_
  - _Requirements: REQ-3_

- [x] 9. Add parser tests for extended dynamic filters
  - File: internal/rtx/parsers/ip_filter_test.go
  - Add test cases for Form 1 with various protocols
  - Add test cases for Form 2 with filter/in/out lists
  - Add test cases for syslog and timeout options
  - Purpose: Verify parser handles all dynamic filter variations
  - _Leverage: Existing dynamic filter test patterns_
  - _Requirements: REQ-3_

- [ ] 10. Create rtx_ip_filter_dynamic resource
  - File: internal/provider/resource_rtx_ip_filter_dynamic.go
  - Implement resourceRTXIPFilterDynamic() with schema definition
  - Schema fields: number, source, destination, protocol (Form 1)
  - Schema fields: filter_list, in_filter_list, out_filter_list (Form 2)
  - Schema fields: syslog (bool), timeout (optional int)
  - Purpose: New Terraform resource for dynamic IP filters
  - _Leverage: resource_rtx_access_list_ip.go as template_
  - _Requirements: REQ-3_

- [ ] 11. Implement CRUD operations for rtx_ip_filter_dynamic
  - File: internal/provider/resource_rtx_ip_filter_dynamic.go
  - Implement Create, Read, Update, Delete functions
  - Implement Import function for existing dynamic filters
  - Purpose: Full lifecycle management of dynamic filters
  - _Leverage: Existing resource CRUD patterns_
  - _Requirements: REQ-3_

- [ ] 12. Register rtx_ip_filter_dynamic in provider
  - File: internal/provider/provider.go
  - Add resource to ResourcesMap
  - Purpose: Make resource available to Terraform users
  - _Leverage: Existing resource registration pattern_
  - _Requirements: REQ-3_

- [ ] 13. Add acceptance tests for rtx_ip_filter_dynamic
  - File: internal/provider/resource_rtx_ip_filter_dynamic_test.go
  - Test Create/Read/Update/Delete lifecycle
  - Test Import functionality
  - Test both Form 1 and Form 2 configurations
  - Purpose: Verify resource works correctly with real RTX router
  - _Leverage: Existing acceptance test patterns_
  - _Requirements: REQ-3_

## Phase 3: NAT Protocol-Only Support (REQ-4)

- [x] 14. Modify MasqueradeStaticEntry struct for optional ports
  - File: internal/rtx/parsers/nat_masquerade.go
  - Change InsideLocalPort from int to *int (pointer for optional)
  - Change OutsideGlobalPort from int to *int (pointer for optional)
  - Purpose: Support protocol-only entries (ESP, AH, GRE) without ports
  - _Leverage: Existing MasqueradeStaticEntry struct_
  - _Requirements: REQ-4_

- [x] 15. Update NAT masquerade static command builder
  - File: internal/rtx/parsers/nat_masquerade.go
  - Modify BuildNATMasqueradeStaticCommand() to handle nil port values
  - Generate `nat descriptor masquerade static <id> <num> <ip> <protocol>` when port is nil
  - Generate full command with ports when port is provided
  - Purpose: Build correct commands for protocol-only entries
  - _Leverage: Existing command builder logic_
  - _Requirements: REQ-4_

- [x] 16. Update NAT masquerade parser for protocol-only entries
  - File: internal/rtx/parsers/nat_masquerade.go
  - Modify ParseNATMasqueradeConfig() to handle entries without port
  - Set port fields to nil when not present in config
  - Purpose: Correctly parse existing protocol-only static entries
  - _Leverage: Existing parser logic_
  - _Requirements: REQ-4_

- [x] 17. Add parser tests for protocol-only NAT entries
  - File: internal/rtx/parsers/nat_masquerade_test.go
  - Test parsing `nat descriptor masquerade static 1000 1 192.168.1.253 esp`
  - Test parsing entries with ah, gre, icmp protocols
  - Test command building for protocol-only entries
  - Purpose: Verify parser handles protocol-only entries correctly
  - _Leverage: Existing NAT masquerade test patterns_
  - _Requirements: REQ-4_

- [x] 18. Update rtx_nat_masquerade resource schema
  - File: internal/provider/resource_rtx_nat_masquerade.go
  - Make inside_local_port optional in static_entry block
  - Make outside_global_port optional in static_entry block
  - Add validation: port required for tcp/udp, not allowed for esp/ah/gre
  - Purpose: Allow Terraform users to create protocol-only entries
  - _Leverage: Existing static_entry schema_
  - _Requirements: REQ-4_

- [x] 19. Update expandStaticEntries and flattenStaticEntries
  - File: internal/provider/resource_rtx_nat_masquerade.go
  - Handle nil port values in expansion and flattening
  - Purpose: Correctly convert between Terraform state and Go structs
  - _Leverage: Existing helper functions_
  - _Requirements: REQ-4_

- [x] 20. Add acceptance tests for protocol-only NAT entries
  - File: internal/provider/resource_rtx_nat_masquerade_test.go
  - Test creating static entry with ESP protocol only
  - Test import of existing protocol-only entries
  - Purpose: Verify protocol-only entries work correctly
  - _Leverage: Existing acceptance test patterns_
  - _Requirements: REQ-4_

## Phase 4: Ethernet Filter Resource (REQ-5)

- [x] 21. Create ethernet filter parser module
  - File: internal/rtx/parsers/ethernet_filter.go
  - Define EthernetFilter struct with fields: Number, Action, SourceMAC, DestinationMAC, DHCPType, DHCPScope, Offset, ByteList
  - Define ValidEthernetFilterActions: pass-log, pass-nolog, reject-log, reject-nolog
  - Purpose: Data structures for ethernet filter parsing
  - _Leverage: ip_filter.go as template_
  - _Requirements: REQ-5_

- [x] 22. Implement ethernet filter parser function
  - File: internal/rtx/parsers/ethernet_filter.go
  - Implement ParseEthernetFilterConfig() to parse `show config` output
  - Handle MAC-based filters: `ethernet filter <id> <action> <src_mac> [<dst_mac>]`
  - Handle DHCP-based filters: `ethernet filter <id> <action> dhcp-bind|dhcp-not-bind [scope]`
  - Handle optional offset and byte_list parameters
  - Purpose: Parse existing ethernet filter configurations
  - _Leverage: ParseIPFilterConfig pattern_
  - _Requirements: REQ-5_

- [x] 23. Implement ethernet filter command builders
  - File: internal/rtx/parsers/ethernet_filter.go
  - Implement BuildEthernetFilterCommand() for MAC-based and DHCP-based filters
  - Implement BuildDeleteEthernetFilterCommand() for filter removal
  - Implement BuildEthernetInterfaceFilterCommand() for interface application
  - Purpose: Generate correct RTX commands for ethernet filters
  - _Leverage: Existing command builder patterns_
  - _Requirements: REQ-5_

- [x] 24. Implement ethernet filter validators
  - File: internal/rtx/parsers/ethernet_filter.go
  - Implement ValidateMACAddress() with regex for xx:xx:xx:xx:xx:xx or * format
  - Implement ValidateEthernetFilterNumber() for range 1-512
  - Implement ValidateEthernetFilter() for complete filter validation
  - Purpose: Ensure valid ethernet filter configurations
  - _Leverage: Existing validation patterns_
  - _Requirements: REQ-5_

- [x] 25. Add ethernet filter parser tests
  - File: internal/rtx/parsers/ethernet_filter_test.go
  - Test parsing MAC-based filters with various MAC formats
  - Test parsing DHCP-based filters with and without scope
  - Test command building for all filter types
  - Test MAC address and filter number validation
  - Purpose: Verify parser handles all ethernet filter variations
  - _Leverage: ip_filter_test.go as template_
  - _Requirements: REQ-5_

- [ ] 26. Create rtx_ethernet_filter resource
  - File: internal/provider/resource_rtx_ethernet_filter.go
  - Implement resourceRTXEthernetFilter() with schema definition
  - Schema fields for MAC-based: number, action, source_mac, destination_mac
  - Schema fields for DHCP-based: dhcp_type, dhcp_scope
  - Schema fields for advanced: offset, byte_list
  - Add ConflictsWith between MAC and DHCP fields
  - Purpose: New Terraform resource for ethernet filters
  - _Leverage: resource_rtx_access_list_ip.go as template_
  - _Requirements: REQ-5_

- [ ] 27. Implement CRUD operations for rtx_ethernet_filter
  - File: internal/provider/resource_rtx_ethernet_filter.go
  - Implement Create, Read, Update, Delete functions
  - Implement Import function for existing ethernet filters
  - Purpose: Full lifecycle management of ethernet filters
  - _Leverage: Existing resource CRUD patterns_
  - _Requirements: REQ-5_

- [ ] 28. Register rtx_ethernet_filter in provider
  - File: internal/provider/provider.go
  - Add resource to ResourcesMap
  - Purpose: Make resource available to Terraform users
  - _Leverage: Existing resource registration pattern_
  - _Requirements: REQ-5_

- [ ] 29. Add acceptance tests for rtx_ethernet_filter
  - File: internal/provider/resource_rtx_ethernet_filter_test.go
  - Test Create/Read/Update/Delete lifecycle for MAC-based filters
  - Test DHCP-based filters
  - Test Import functionality
  - Purpose: Verify resource works correctly with real RTX router
  - _Leverage: Existing acceptance test patterns_
  - _Requirements: REQ-5_

## Final Integration

- [ ] 30. Update examples/import/main.tf with new resources
  - File: examples/import/main.tf
  - Add examples for rtx_ip_filter_dynamic resource
  - Add examples for rtx_ethernet_filter resource
  - Update existing IP filter examples with new protocols
  - Purpose: Provide usage examples for new features
  - _Leverage: Existing example patterns_
  - _Requirements: All_

- [ ] 31. Update import.sh with new resource imports
  - File: examples/import/import.sh
  - Add import commands for dynamic filters
  - Add import commands for ethernet filters
  - Purpose: Support importing new resource types
  - _Leverage: Existing import script patterns_
  - _Requirements: All_

- [ ] 32. Run full test suite and fix any issues
  - Run `go test ./...` to verify all tests pass
  - Run acceptance tests with `TF_ACC=1`
  - Fix any integration issues discovered
  - Purpose: Ensure all components work together correctly
  - _Leverage: Existing test infrastructure_
  - _Requirements: All_
