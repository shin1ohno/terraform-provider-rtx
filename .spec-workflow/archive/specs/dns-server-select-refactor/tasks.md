# Tasks Document: DNS Server Select Schema Refactor

## Phase 1: Parser Layer

- [x] 1. Update DNSServerSelect struct in parser layer
  - File: internal/rtx/parsers/dns.go
  - Modify DNSServerSelect struct to include new fields: EDNS, RecordType, QueryPattern, OriginalSender, RestrictPP
  - Remove Domains field
  - Purpose: Define new data model for DNS server select entries
  - _Leverage: Existing struct pattern in internal/rtx/parsers/dns.go_
  - _Requirements: 1, 2, 3, 4, 5, 6_
  - _Prompt: Implement the task for spec dns-server-select-refactor, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in data modeling | Task: Update DNSServerSelect struct in internal/rtx/parsers/dns.go to replace Domains []string with new fields: EDNS bool, RecordType string, QueryPattern string, OriginalSender string, RestrictPP int. Follow JSON tag conventions. | Restrictions: Do not modify other structs in the file, maintain backward compatibility considerations, keep existing comments style | _Leverage: internal/rtx/parsers/dns.go existing struct patterns | _Requirements: 1, 2, 3, 4, 5, 6 | Success: Struct compiles without errors, JSON tags are correct, fields match design document. Instructions: Mark task as in-progress in tasks.md before starting, log implementation with log-implementation tool after completion, then mark as complete._

- [x] 2. Update ParseDNSConfig function for new schema
  - File: internal/rtx/parsers/dns.go
  - Modify regex pattern for `dns server select` command
  - Parse EDNS option, record type, query pattern, original sender, restrict pp
  - Purpose: Parse RTX CLI output into new DNSServerSelect struct
  - _Leverage: Existing ParseDNSConfig function, regexp patterns_
  - _Requirements: 7_
  - _Prompt: Implement the task for spec dns-server-select-refactor, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with regex expertise | Task: Update ParseDNSConfig in internal/rtx/parsers/dns.go to parse RTX command format: `dns server select <id> <server> [edns=on] [type] <pattern> [sender] [restrict pp n]`. Extract all new fields correctly. | Restrictions: Handle optional fields gracefully, maintain existing parsing logic for other DNS commands, test with various input formats | _Leverage: internal/rtx/parsers/dns.go existing regex patterns | _Requirements: 7 | Success: Parser correctly extracts all fields from various command formats, handles edge cases. Instructions: Mark task as in-progress in tasks.md before starting, log implementation with log-implementation tool after completion, then mark as complete._

- [x] 3. Update BuildDNSServerSelectCommand function
  - File: internal/rtx/parsers/dns.go
  - Generate RTX CLI command from new DNSServerSelect struct
  - Handle optional fields (edns, original_sender, restrict_pp) correctly
  - Purpose: Build correct RTX command from Terraform configuration
  - _Leverage: Existing BuildDNSServerSelectCommand function_
  - _Requirements: 7_
  - _Prompt: Implement the task for spec dns-server-select-refactor, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Update BuildDNSServerSelectCommand in internal/rtx/parsers/dns.go to generate command format: `dns server select <id> <servers> [edns=on] [type] <pattern> [sender] [restrict pp n]`. Only include optional parts when values are set. | Restrictions: Maintain correct RTX command order, handle empty strings and zero values properly, ensure command is valid for RTX router | _Leverage: internal/rtx/parsers/dns.go BuildDNSServerSelectCommand | _Requirements: 7 | Success: Generated commands match RTX syntax exactly, optional fields omitted when not set. Instructions: Mark task as in-progress in tasks.md before starting, log implementation with log-implementation tool after completion, then mark as complete._

- [x] 4. Update ValidateDNSConfig function
  - File: internal/rtx/parsers/dns.go
  - Add validation for new fields: RecordType must be valid value, QueryPattern required
  - Purpose: Validate configuration before sending to router
  - _Leverage: Existing ValidateDNSConfig function_
  - _Requirements: 2, 3_
  - _Prompt: Implement the task for spec dns-server-select-refactor, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Update ValidateDNSConfig in internal/rtx/parsers/dns.go to validate new fields: RecordType must be one of (a, aaaa, ptr, mx, ns, cname, any), QueryPattern must not be empty. | Restrictions: Maintain existing validation logic, provide clear error messages, follow existing error format | _Leverage: internal/rtx/parsers/dns.go ValidateDNSConfig | _Requirements: 2, 3 | Success: Validation catches invalid record types and empty patterns, error messages are actionable. Instructions: Mark task as in-progress in tasks.md before starting, log implementation with log-implementation tool after completion, then mark as complete._

- [x] 5. Add parser unit tests for new schema
  - File: internal/rtx/parsers/dns_test.go
  - Add test cases for parsing new field combinations
  - Test command building with various field combinations
  - Purpose: Ensure parser reliability
  - _Leverage: Existing test patterns in dns_test.go_
  - _Requirements: All_
  - _Prompt: Implement the task for spec dns-server-select-refactor, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with testing expertise | Task: Add unit tests in internal/rtx/parsers/dns_test.go for new DNSServerSelect schema. Test parsing of: edns=on/off, all record types, various query patterns, original sender with CIDR, restrict pp values. Test command building for all combinations. | Restrictions: Follow existing test patterns, use table-driven tests, cover edge cases | _Leverage: internal/rtx/parsers/dns_test.go existing test patterns | _Requirements: All | Success: All new fields tested with parsing and building, edge cases covered, tests pass. Instructions: Mark task as in-progress in tasks.md before starting, log implementation with log-implementation tool after completion, then mark as complete._

## Phase 2: Client Layer

- [x] 6. Update DNSServerSelect struct in client interfaces
  - File: internal/client/interfaces.go
  - Mirror the updated DNSServerSelect struct from parser layer
  - Purpose: Maintain struct consistency between layers
  - _Leverage: internal/client/interfaces.go DNSServerSelect_
  - _Requirements: All_
  - _Prompt: Implement the task for spec dns-server-select-refactor, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Update DNSServerSelect struct in internal/client/interfaces.go to match the parser layer struct. Fields: ID int, Servers []string, EDNS bool, RecordType string, QueryPattern string, OriginalSender string, RestrictPP int. | Restrictions: Keep JSON tags consistent with parser layer, maintain struct location in file | _Leverage: internal/client/interfaces.go DNSServerSelect | _Requirements: All | Success: Struct matches parser layer exactly, compiles without errors. Instructions: Mark task as in-progress in tasks.md before starting, log implementation with log-implementation tool after completion, then mark as complete._

- [x] 7. Update dns_service.go conversion functions
  - File: internal/client/dns_service.go
  - Update toParserConfig and fromParserConfig to handle new fields
  - Purpose: Convert between client and parser layer correctly
  - _Leverage: internal/client/dns_service.go toParserConfig, fromParserConfig_
  - _Requirements: All_
  - _Prompt: Implement the task for spec dns-server-select-refactor, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Update toParserConfig and fromParserConfig in internal/client/dns_service.go to convert all new DNSServerSelect fields between client and parser types. | Restrictions: Handle nil/empty values correctly, maintain existing conversion patterns | _Leverage: internal/client/dns_service.go conversion functions | _Requirements: All | Success: All new fields converted correctly in both directions. Instructions: Mark task as in-progress in tasks.md before starting, log implementation with log-implementation tool after completion, then mark as complete._

- [x] 8. Update dns_service_test.go
  - File: internal/client/dns_service_test.go
  - Update test cases to use new schema
  - Purpose: Ensure service layer tests pass with new schema
  - _Leverage: internal/client/dns_service_test.go existing tests_
  - _Requirements: All_
  - _Prompt: Implement the task for spec dns-server-select-refactor, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Update test cases in internal/client/dns_service_test.go to use new DNSServerSelect fields. Update mock outputs and expected values. | Restrictions: Maintain existing test coverage, follow existing test patterns | _Leverage: internal/client/dns_service_test.go | _Requirements: All | Success: All existing tests updated and passing with new schema. Instructions: Mark task as in-progress in tasks.md before starting, log implementation with log-implementation tool after completion, then mark as complete._

## Phase 3: Provider Layer

- [x] 9. Update Terraform schema for server_select
  - File: internal/provider/resource_rtx_dns_server.go
  - Remove "domains" field from schema
  - Add new fields: edns (bool), record_type (string), query_pattern (string), original_sender (string), restrict_pp (int)
  - Add validation for record_type using StringInSlice
  - Purpose: Define new Terraform schema for users
  - _Leverage: internal/provider/resource_rtx_dns_server.go server_select schema_
  - _Requirements: 1, 2, 3, 4, 5, 6_
  - _Prompt: Implement the task for spec dns-server-select-refactor, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Update server_select schema in internal/provider/resource_rtx_dns_server.go. Remove domains field. Add: edns (TypeBool, Optional, Default false), record_type (TypeString, Optional, Default "a", ValidateFunc StringInSlice), query_pattern (TypeString, Required), original_sender (TypeString, Optional), restrict_pp (TypeInt, Optional, Default 0). | Restrictions: Follow existing schema patterns, add proper descriptions, use appropriate validation functions | _Leverage: internal/provider/resource_rtx_dns_server.go schema patterns | _Requirements: 1, 2, 3, 4, 5, 6 | Success: Schema compiles, validation works correctly, descriptions are clear. Instructions: Mark task as in-progress in tasks.md before starting, log implementation with log-implementation tool after completion, then mark as complete._

- [x] 10. Update buildDNSConfigFromResourceData function
  - File: internal/provider/resource_rtx_dns_server.go
  - Update to extract new fields from Terraform resource data
  - Build client.DNSServerSelect with new fields
  - Purpose: Convert Terraform config to client types
  - _Leverage: internal/provider/resource_rtx_dns_server.go buildDNSConfigFromResourceData_
  - _Requirements: All_
  - _Prompt: Implement the task for spec dns-server-select-refactor, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Update buildDNSConfigFromResourceData in internal/provider/resource_rtx_dns_server.go to extract new server_select fields: edns, record_type, query_pattern, original_sender, restrict_pp. Build client.DNSServerSelect with these fields. | Restrictions: Handle type assertions safely, follow existing patterns for extracting nested blocks | _Leverage: internal/provider/resource_rtx_dns_server.go buildDNSConfigFromResourceData | _Requirements: All | Success: All new fields extracted correctly, no panics on type assertions. Instructions: Mark task as in-progress in tasks.md before starting, log implementation with log-implementation tool after completion, then mark as complete._

- [x] 11. Update Read function flatten logic
  - File: internal/provider/resource_rtx_dns_server.go
  - Update resourceRTXDNSServerRead to flatten new fields into state
  - Update resourceRTXDNSServerImport similarly
  - Purpose: Read state correctly from router
  - _Leverage: internal/provider/resource_rtx_dns_server.go resourceRTXDNSServerRead_
  - _Requirements: All_
  - _Prompt: Implement the task for spec dns-server-select-refactor, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Update resourceRTXDNSServerRead and resourceRTXDNSServerImport in internal/provider/resource_rtx_dns_server.go to flatten new server_select fields into Terraform state: edns, record_type, query_pattern, original_sender, restrict_pp. | Restrictions: Follow existing flatten patterns, handle nil values correctly | _Leverage: internal/provider/resource_rtx_dns_server.go Read and Import functions | _Requirements: All | Success: All fields flattened correctly into state, import works correctly. Instructions: Mark task as in-progress in tasks.md before starting, log implementation with log-implementation tool after completion, then mark as complete._

- [x] 12. Update provider tests
  - File: internal/provider/resource_rtx_dns_server_test.go
  - Update test configurations to use new schema
  - Add tests for new field combinations
  - Purpose: Ensure provider works correctly with new schema
  - _Leverage: internal/provider/resource_rtx_dns_server_test.go_
  - _Requirements: All_
  - _Prompt: Implement the task for spec dns-server-select-refactor, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Update tests in internal/provider/resource_rtx_dns_server_test.go to use new server_select schema. Update test HCL configurations and expected values. Add tests for various field combinations. | Restrictions: Maintain existing test coverage, follow existing test patterns | _Leverage: internal/provider/resource_rtx_dns_server_test.go | _Requirements: All | Success: All tests updated and passing, new field combinations tested. Instructions: Mark task as in-progress in tasks.md before starting, log implementation with log-implementation tool after completion, then mark as complete._

## Phase 4: Documentation and Verification

- [x] 13. Update examples/dns_server/main.tf
  - File: examples/import/main.tf
  - Update rtx_dns_server resource to use new schema
  - Purpose: Provide working example for users
  - _Leverage: examples/import/main.tf_
  - _Requirements: All_
  - _Prompt: Implement the task for spec dns-server-select-refactor, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer | Task: Update rtx_dns_server resource in examples/import/main.tf to use new server_select schema. Convert existing domains field to edns, record_type, query_pattern fields. | Restrictions: Ensure example is realistic and matches actual router state | _Leverage: examples/import/main.tf | _Requirements: All | Success: Example uses new schema correctly, is clear and understandable. Instructions: Mark task as in-progress in tasks.md before starting, log implementation with log-implementation tool after completion, then mark as complete._

- [x] 14. Run all tests and fix any issues
  - Files: All test files
  - Run `go test ./...` and fix any failing tests
  - Purpose: Ensure all tests pass before completion
  - _Leverage: All test files_
  - _Requirements: All_
  - _Prompt: Implement the task for spec dns-server-select-refactor, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run `go test ./...` in project root and fix any failing tests related to DNS server select changes. | Restrictions: Only fix test failures, do not change unrelated code | _Leverage: All test files | _Requirements: All | Success: All tests pass, no regressions introduced. Instructions: Mark task as in-progress in tasks.md before starting, log implementation with log-implementation tool after completion, then mark as complete._

- [x] 15. Build and verify provider works
  - Files: All
  - Run `make install` and test with terraform plan
  - Purpose: Final verification of the implementation
  - _Leverage: Makefile_
  - _Requirements: All_
  - _Prompt: Implement the task for spec dns-server-select-refactor, first run spec-workflow-guide to get the workflow guide then implement the task: Role: DevOps Engineer | Task: Run `make install` to build and install provider. Run `terraform plan` in examples/import to verify no schema errors. | Restrictions: Do not modify code unless absolutely necessary for build | _Leverage: Makefile | _Requirements: All | Success: Provider builds successfully, terraform plan runs without schema errors. Instructions: Mark task as in-progress in tasks.md before starting, log implementation with log-implementation tool after completion, then mark as complete._
