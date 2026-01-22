# Tasks Document: RTX DDNS Resource

## Phase 1: Parser Layer

- [x] 1. Create DDNS parser core structures
  - File: internal/rtx/parsers/ddns.go
  - Define NetVolanteConfig, DDNSConfig, DDNSStatus structs
  - Add validation functions for hostnames, URLs
  - Purpose: Establish data models for DDNS
  - _Leverage: internal/rtx/parsers/dns.go for patterns_
  - _Requirements: REQ-1, REQ-3_
  - _Prompt: Role: Go Developer | Task: Create ddns.go with core data structures for DDNS configurations from REQ-1, REQ-3 | Restrictions: Follow existing parser patterns | Success: Structs defined, validation implemented_

- [x] 2. Implement NetVolante DNS parser
  - File: internal/rtx/parsers/ddns.go
  - Add ParseNetVolanteDNS function
  - Parse: netvolante-dns hostname, server, go, timeout, ipv6, auto hostname
  - Purpose: Parse existing NetVolante configurations
  - _Leverage: RTX command documentation Chapter 27_
  - _Requirements: REQ-3_
  - _Prompt: Role: Go Developer | Task: Implement ParseNetVolanteDNS to parse all NetVolante settings from REQ-3 | Success: All NetVolante commands parsed_

- [x] 3. Implement custom DDNS parser
  - File: internal/rtx/parsers/ddns.go
  - Add ParseDDNSConfig function
  - Parse: ddns server url, hostname, user, password, go
  - Purpose: Parse custom DDNS provider configurations
  - _Leverage: RTX command documentation Chapter 27_
  - _Requirements: REQ-1_
  - _Prompt: Role: Go Developer | Task: Implement ParseDDNSConfig for custom DDNS providers from REQ-1 | Success: All DDNS commands parsed_

- [x] 4. Implement DDNS status parser
  - File: internal/rtx/parsers/ddns.go
  - Add ParseDDNSStatus function
  - Parse: show status netvolante-dns, show status ddns
  - Purpose: Parse DDNS registration status
  - _Leverage: RTX show command output patterns_
  - _Requirements: REQ-4_
  - _Prompt: Role: Go Developer | Task: Implement ParseDDNSStatus for status monitoring from REQ-4 | Success: Status parsing works_

- [x] 5. Implement DDNS command builders
  - File: internal/rtx/parsers/ddns.go
  - Add BuildNetVolanteCommand, BuildDDNSCommand functions
  - Generate RTX commands from config structs
  - Purpose: Generate configuration commands
  - _Leverage: Existing command builder patterns_
  - _Requirements: REQ-1, REQ-2, REQ-3_
  - _Prompt: Role: Go Developer | Task: Implement command builders for DDNS configurations | Success: Commands generate correctly_

- [x] 6. Create DDNS parser tests
  - File: internal/rtx/parsers/ddns_test.go
  - Add tests for all parser and builder functions
  - Include status parsing tests
  - Purpose: Ensure parser reliability
  - _Leverage: Existing test patterns_
  - _Requirements: All_
  - _Prompt: Role: Go Test Developer | Task: Create comprehensive tests for DDNS parser | Success: All parser functions tested_

- [x] 7. Create DDNS pattern catalog
  - File: internal/rtx/testdata/patterns/ddns.yaml
  - Document all DDNS command patterns
  - Include examples from RTX documentation
  - Purpose: Document command formats
  - _Leverage: internal/rtx/testdata/patterns/schema.yaml_
  - _Requirements: REQ-1, REQ-3_
  - _Prompt: Role: Documentation Engineer | Task: Create ddns.yaml pattern catalog | Success: All DDNS commands documented_

## Phase 2: Client Layer

- [x] 8. Create DDNS service interface
  - File: internal/client/ddns_service.go
  - Define DDNSService interface with all methods
  - Purpose: Establish service contract
  - _Leverage: internal/client/dns_service.go for patterns_
  - _Requirements: REQ-1, REQ-3, REQ-4_
  - _Prompt: Role: Go Developer | Task: Define DDNSService interface with CRUD and status methods | Success: Interface defined_

- [x] 9. Implement NetVolante DNS service methods
  - File: internal/client/ddns_service.go
  - Implement GetNetVolanteDNS, ConfigureNetVolanteDNS, UpdateNetVolanteDNS, DeleteNetVolanteDNS
  - Implement TriggerNetVolanteUpdate for manual updates
  - Purpose: Enable NetVolante operations via SSH
  - _Leverage: internal/client/client.go executor_
  - _Requirements: REQ-3_
  - _Prompt: Role: Go Developer | Task: Implement NetVolante CRUD operations | Success: All NetVolante operations work_

- [x] 10. Implement custom DDNS service methods
  - File: internal/client/ddns_service.go
  - Implement GetDDNS, ConfigureDDNS, DeleteDDNS
  - Add error handling for network failures
  - Purpose: Enable custom DDNS operations via SSH
  - _Leverage: internal/client/client.go executor_
  - _Requirements: REQ-1_
  - _Prompt: Role: Go Developer | Task: Implement custom DDNS CRUD operations | Success: All DDNS operations work_

- [x] 11. Implement DDNS status service methods
  - File: internal/client/ddns_service.go
  - Implement GetDDNSStatus for status monitoring
  - Parse last update time, registered IP, status
  - Purpose: Enable status monitoring
  - _Leverage: Status parsing from ddns.go_
  - _Requirements: REQ-4_
  - _Prompt: Role: Go Developer | Task: Implement DDNS status monitoring | Success: Status retrieval works_

- [x] 12. Create DDNS service tests
  - File: internal/client/ddns_service_test.go
  - Add tests with mock executor
  - Test CRUD operations and status retrieval
  - Purpose: Ensure service reliability
  - _Leverage: Existing client test patterns_
  - _Requirements: All_
  - _Prompt: Role: Go Test Developer | Task: Create DDNS service tests with mocks | Success: All service methods tested_

## Phase 3: Provider Layer

- [x] 13. Create NetVolante DNS Terraform resource
  - File: internal/provider/resource_rtx_netvolante_dns.go
  - Define schema with hostname, server, interface, update_interval, ipv6_enabled
  - Implement Create, Read, Update, Delete functions
  - Purpose: Enable Terraform management of NetVolante DNS
  - _Leverage: internal/provider/resource_rtx_dns_server.go for singleton patterns_
  - _Requirements: REQ-3_
  - _Prompt: Role: Terraform Provider Developer | Task: Create rtx_netvolante_dns resource with full CRUD | Success: Resource CRUD works_

- [x] 14. Create custom DDNS Terraform resource
  - File: internal/provider/resource_rtx_ddns.go
  - Define schema with provider_url, hostname, username, password, interface
  - Implement Create, Read, Update, Delete functions
  - Purpose: Enable Terraform management of custom DDNS
  - _Leverage: Existing resource patterns_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Role: Terraform Provider Developer | Task: Create rtx_ddns resource with full CRUD | Restrictions: Mark password as sensitive | Success: Resource CRUD works_

- [x] 15. Create DDNS status data source
  - File: internal/provider/data_source_rtx_ddns_status.go
  - Define schema for status output
  - Implement Read function
  - Purpose: Enable status monitoring in Terraform
  - _Leverage: Existing data source patterns_
  - _Requirements: REQ-4_
  - _Prompt: Role: Terraform Provider Developer | Task: Create rtx_ddns_status data source | Success: Data source works_

- [x] 16. Add import support to resources
  - File: internal/provider/resource_rtx_netvolante_dns.go, resource_rtx_ddns.go
  - Implement Importer functions
  - Handle credential preservation
  - Purpose: Enable import of existing configs
  - _Leverage: Existing import patterns_
  - _Requirements: REQ-1, REQ-3_
  - _Prompt: Role: Terraform Provider Developer | Task: Add import support for DDNS resources | Success: Import works_

- [x] 17. Register resources in provider
  - File: internal/provider/provider.go
  - Add rtx_netvolante_dns, rtx_ddns to ResourcesMap
  - Add rtx_ddns_status to DataSourcesMap
  - Purpose: Make resources available in Terraform
  - _Leverage: Existing provider.go registration_
  - _Requirements: All_
  - _Prompt: Role: Terraform Provider Developer | Task: Register DDNS resources in provider.go | Success: Resources registered_

## Phase 4: Documentation and Examples

- [x] 18. Create DDNS example configurations
  - File: examples/ddns/main.tf
  - Add examples for NetVolante DNS and common providers
  - Include multi-WAN failover example
  - Purpose: Provide user documentation
  - _Leverage: Existing examples format_
  - _Requirements: All_
  - _Prompt: Role: Technical Writer | Task: Create DDNS example configurations | Success: Examples are clear_

- [x] 19. Create provider tests
  - File: internal/provider/resource_rtx_netvolante_dns_test.go, resource_rtx_ddns_test.go, data_source_rtx_ddns_status_test.go
  - Add unit tests with mock client
  - Purpose: Ensure provider reliability
  - _Leverage: Existing provider test patterns_
  - _Requirements: All_
  - _Prompt: Role: Go Test Developer | Task: Create provider resource tests | Success: All functions tested_

## Phase 5: Integration

- [x] 20. Build and test full stack
  - Run go build ./...
  - Run go test ./...
  - Fix any integration issues
  - Purpose: Ensure everything works together
  - _Leverage: Existing build/test infrastructure_
  - _Requirements: All_
  - _Prompt: Role: Integration Engineer | Task: Build and test full DDNS implementation | Success: Build succeeds, all tests pass_
