# Tasks Document: RTX PPP/PPPoE Resource

## Phase 1: Parser Layer

- [x] 1. Create PPP parser core structures
  - File: internal/rtx/parsers/ppp.go
  - Define PPPoEConfig, PPInterfaceConfig, LCPEchoConfig structs
  - Add validation functions for auth methods, MTU ranges
  - Purpose: Establish data models for PPP/PPPoE
  - _Leverage: internal/rtx/parsers/interface_config.go for patterns_
  - _Requirements: REQ-1, REQ-2, REQ-3_
  - _Prompt: Role: Go Developer | Task: Create ppp.go with core data structures for PPPoE and PP interface configurations from REQ-1, REQ-2, REQ-3 | Restrictions: Follow existing parser patterns | Success: Structs defined, validation implemented_

- [x] 2. Implement PPPoE config parser
  - File: internal/rtx/parsers/ppp.go
  - Add ParsePPPoEConfig function
  - Parse: pp bind, pp auth myname, pp auth accept, pppoe service-name, pp always-on
  - Purpose: Parse existing PPPoE configurations
  - _Leverage: RTX command documentation Chapter 10_
  - _Requirements: REQ-1_
  - _Prompt: Role: Go Developer | Task: Implement ParsePPPoEConfig to parse all PPPoE settings from REQ-1 | Success: All PPPoE commands parsed correctly_

- [x] 3. Implement PP interface config parser
  - File: internal/rtx/parsers/ppp.go
  - Add ParsePPInterfaceConfig function
  - Parse: ip pp address, ip pp mtu, ip pp tcp mss, ip pp nat descriptor
  - Purpose: Parse PP interface IP configurations
  - _Leverage: RTX command documentation Chapter 10_
  - _Requirements: REQ-2_
  - _Prompt: Role: Go Developer | Task: Implement ParsePPInterfaceConfig to parse all PP interface settings from REQ-2 | Success: All PP interface commands parsed_

- [x] 4. Implement PPP command builders
  - File: internal/rtx/parsers/ppp.go
  - Add BuildPPPoECommand, BuildPPInterfaceCommand functions
  - Generate RTX commands from config structs
  - Purpose: Generate configuration commands
  - _Leverage: Existing command builder patterns_
  - _Requirements: REQ-1, REQ-2, REQ-3_
  - _Prompt: Role: Go Developer | Task: Implement command builders for PPPoE and PP interface configurations | Success: Commands generate correctly_

- [x] 5. Create PPP parser tests
  - File: internal/rtx/parsers/ppp_test.go
  - Add tests for ParsePPPoEConfig, ParsePPInterfaceConfig
  - Add tests for command builders
  - Purpose: Ensure parser reliability
  - _Leverage: Existing test patterns_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Role: Go Test Developer | Task: Create comprehensive tests for PPP parser | Success: All parser functions tested_

- [x] 6. Create PPP pattern catalog
  - File: internal/rtx/testdata/patterns/ppp.yaml
  - Document all PPP/PPPoE command patterns
  - Include examples from RTX documentation
  - Purpose: Document command formats
  - _Leverage: internal/rtx/testdata/patterns/schema.yaml_
  - _Requirements: REQ-1, REQ-2, REQ-3_
  - _Prompt: Role: Documentation Engineer | Task: Create ppp.yaml pattern catalog | Success: All PPP commands documented_

## Phase 2: Client Layer

- [x] 7. Create PPP service interface
  - File: internal/client/ppp_service.go
  - Define PPPService interface with all methods
  - Purpose: Establish service contract
  - _Leverage: internal/client/interface_service.go for patterns_
  - _Requirements: REQ-1, REQ-2, REQ-4_
  - _Prompt: Role: Go Developer | Task: Define PPPService interface with CRUD methods | Success: Interface defined_

- [x] 8. Implement PPPoE service methods
  - File: internal/client/ppp_service.go
  - Implement GetPPPoE, ConfigurePPPoE, UpdatePPPoE, DeletePPPoE
  - Add error handling for authentication failures
  - Purpose: Enable PPPoE operations via SSH
  - _Leverage: internal/client/client.go executor_
  - _Requirements: REQ-1, REQ-4_
  - _Prompt: Role: Go Developer | Task: Implement PPPoE CRUD operations | Success: All PPPoE operations work_

- [x] 9. Implement PP interface service methods
  - File: internal/client/ppp_service.go
  - Implement GetPPInterface, ConfigurePPInterface
  - Implement GetConnectionStatus for status monitoring
  - Purpose: Enable PP interface operations via SSH
  - _Leverage: internal/client/client.go executor_
  - _Requirements: REQ-2, REQ-4_
  - _Prompt: Role: Go Developer | Task: Implement PP interface operations | Success: All PP interface operations work_

- [x] 10. Create PPP service tests
  - File: internal/client/ppp_service_test.go
  - Add tests with mock executor
  - Test CRUD operations and error handling
  - Purpose: Ensure service reliability
  - _Leverage: Existing client test patterns_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Role: Go Test Developer | Task: Create PPP service tests with mocks | Success: All service methods tested_

## Phase 3: Provider Layer

- [ ] 11. Create PPPoE Terraform resource
  - File: internal/provider/resource_rtx_pppoe.go
  - Define schema with all PPPoE attributes
  - Implement Create, Read, Update, Delete functions
  - Purpose: Enable Terraform management of PPPoE
  - _Leverage: internal/provider/resource_rtx_interface.go for patterns_
  - _Requirements: REQ-1, REQ-3_
  - _Prompt: Role: Terraform Provider Developer | Task: Create rtx_pppoe resource with full CRUD | Restrictions: Mark password as sensitive | Success: Resource CRUD works_

- [ ] 12. Add PPPoE resource import support
  - File: internal/provider/resource_rtx_pppoe.go
  - Implement Importer function
  - Handle encrypted password preservation
  - Purpose: Enable import of existing configs
  - _Leverage: Existing import patterns_
  - _Requirements: REQ-4_
  - _Prompt: Role: Terraform Provider Developer | Task: Add import support for rtx_pppoe | Success: Import works correctly_

- [ ] 13. Create PP interface Terraform resource
  - File: internal/provider/resource_rtx_pp_interface.go
  - Define schema with IP, MTU, NAT attributes
  - Implement Create, Read, Update, Delete functions
  - Purpose: Enable Terraform management of PP interfaces
  - _Leverage: internal/provider/resource_rtx_interface.go for patterns_
  - _Requirements: REQ-2, REQ-5_
  - _Prompt: Role: Terraform Provider Developer | Task: Create rtx_pp_interface resource with full CRUD | Success: Resource CRUD works_

- [ ] 14. Register resources in provider
  - File: internal/provider/provider.go
  - Add rtx_pppoe and rtx_pp_interface to ResourcesMap
  - Purpose: Make resources available in Terraform
  - _Leverage: Existing provider.go registration_
  - _Requirements: All_
  - _Prompt: Role: Terraform Provider Developer | Task: Register PPP resources in provider.go | Success: Resources registered_

## Phase 4: Documentation and Examples

- [ ] 15. Create PPPoE example configurations
  - File: examples/pppoe/main.tf
  - Add examples for NTT FLET'S, common Japanese ISPs
  - Include multi-WAN failover example
  - Purpose: Provide user documentation
  - _Leverage: Existing examples format_
  - _Requirements: All_
  - _Prompt: Role: Technical Writer | Task: Create PPPoE example configurations | Success: Examples are clear and complete_

- [ ] 16. Create provider tests
  - File: internal/provider/resource_rtx_pppoe_test.go, resource_rtx_pp_interface_test.go
  - Add unit tests with mock client
  - Purpose: Ensure provider reliability
  - _Leverage: Existing provider test patterns_
  - _Requirements: All_
  - _Prompt: Role: Go Test Developer | Task: Create provider resource tests | Success: All provider functions tested_

## Phase 5: Integration

- [ ] 17. Build and test full stack
  - Run go build ./...
  - Run go test ./...
  - Fix any integration issues
  - Purpose: Ensure everything works together
  - _Leverage: Existing build/test infrastructure_
  - _Requirements: All_
  - _Prompt: Role: Integration Engineer | Task: Build and test full PPP implementation | Success: Build succeeds, all tests pass_
