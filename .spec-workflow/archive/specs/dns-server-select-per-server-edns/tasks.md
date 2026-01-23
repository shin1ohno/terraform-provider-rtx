# Tasks Document: DNS Server Select Per-Server EDNS

- [x] 1. Add DNSServer struct and update DNSServerSelect in client interfaces
  - File: internal/client/interfaces.go
  - Add new `DNSServer` struct with `Address` and `EDNS` fields
  - Update `DNSServerSelect.Servers` from `[]string` to `[]DNSServer`
  - Remove `EDNS bool` field from `DNSServerSelect`
  - Purpose: Align client types with parser types for per-server EDNS
  - _Leverage: internal/rtx/parsers/dns.go (DNSServer struct as reference)_
  - _Requirements: 1, 3, 4_
  - _Prompt: Implement the task for spec dns-server-select-per-server-edns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in Terraform provider development | Task: Add DNSServer struct and update DNSServerSelect in internal/client/interfaces.go to support per-server EDNS configuration. The DNSServer struct should have Address (string) and EDNS (bool) fields. Update DNSServerSelect.Servers to use []DNSServer instead of []string, and remove the standalone EDNS bool field. Reference the parser's DNSServer struct in internal/rtx/parsers/dns.go for consistency. | Restrictions: Do not modify parser layer, maintain JSON tags for serialization, keep other fields in DNSServerSelect unchanged | Success: DNSServer struct added, DNSServerSelect updated, code compiles without errors | After completing, mark task as in-progress in tasks.md before starting, use log-implementation tool to record details, then mark as complete._

- [x] 2. Update DNS service conversion functions
  - File: internal/client/dns_service.go
  - Simplify `convertDNSServerSelectToParser` to directly map DNSServer fields
  - Simplify `convertDNSServerSelectFromParser` to directly map DNSServer fields
  - Remove the EDNS flattening/aggregation logic
  - Purpose: Clean conversion between aligned client and parser types
  - _Leverage: internal/rtx/parsers/dns.go (DNSServer, DNSServerSelect structs)_
  - _Requirements: 3_
  - _Prompt: Implement the task for spec dns-server-select-per-server-edns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with expertise in type conversion and service layers | Task: Update convertDNSServerSelectToParser and convertDNSServerSelectFromParser in internal/client/dns_service.go to directly map the new DNSServer struct fields. The old code applied a single EDNS to all servers or aggregated EDNS from multiple servers - this should now be a direct per-server copy. | Restrictions: Do not change other service methods, maintain error handling patterns, ensure bidirectional conversion preserves all data | Success: Conversion functions directly map Address and EDNS for each server, roundtrip conversion preserves per-server EDNS settings | After completing, mark task as in-progress in tasks.md before starting, use log-implementation tool to record details, then mark as complete._

- [x] 3. Update Terraform schema for server_select
  - File: internal/provider/resource_rtx_dns_server.go
  - Replace `servers` (TypeList of strings) and `edns` (TypeBool) with `server` nested block
  - Add `server` block with `address` (required string) and `edns` (optional bool, default false)
  - Add MinItems: 1 and MaxItems: 2 validation for `server` block
  - Purpose: Enable per-server EDNS in Terraform HCL configuration
  - _Leverage: existing nested block patterns in other resources (e.g., rtx_nat_masquerade static_entry)_
  - _Requirements: 1, 2_
  - _Prompt: Implement the task for spec dns-server-select-per-server-edns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer with expertise in schema design | Task: Update the server_select schema in internal/provider/resource_rtx_dns_server.go to replace the flat servers list and edns boolean with a nested server block. Each server block should have address (required, validated with validateIPAddressAny) and edns (optional, default false). Add MinItems:1 and MaxItems:2 constraints. | Restrictions: Do not change other schema fields (id, record_type, query_pattern, etc.), follow existing nested block patterns, maintain backward compatibility error messages | Success: New schema accepts server blocks with per-server EDNS, validation enforces 1-2 servers per entry | After completing, mark task as in-progress in tasks.md before starting, use log-implementation tool to record details, then mark as complete._

- [x] 4. Update buildDNSConfigFromResourceData function
  - File: internal/provider/resource_rtx_dns_server.go
  - Update server_select parsing to extract `server` blocks
  - Create `[]client.DNSServer` from nested block data
  - Handle edns default (false when not specified)
  - Purpose: Convert Terraform state to client types correctly
  - _Leverage: existing nested block parsing patterns in the same file_
  - _Requirements: 1, 3_
  - _Prompt: Implement the task for spec dns-server-select-per-server-edns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Update buildDNSConfigFromResourceData in internal/provider/resource_rtx_dns_server.go to parse the new server nested blocks. Extract address and edns from each server block and create []client.DNSServer. Handle the case where edns is not specified (default to false). | Restrictions: Do not change parsing of other fields, maintain existing error handling patterns | Success: Function correctly extracts per-server EDNS from Terraform state, handles defaults properly | After completing, mark task as in-progress in tasks.md before starting, use log-implementation tool to record details, then mark as complete._

- [x] 5. Update Read and Import flatten logic
  - File: internal/provider/resource_rtx_dns_server.go
  - Update `resourceRTXDNSServerRead` to flatten `[]client.DNSServer` to `server` blocks
  - Update `resourceRTXDNSServerImport` with the same flatten logic
  - Ensure each server's address and edns are preserved
  - Purpose: Correctly populate Terraform state from router configuration
  - _Leverage: existing flatten patterns for nested blocks_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec dns-server-select-per-server-edns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Update resourceRTXDNSServerRead and resourceRTXDNSServerImport in internal/provider/resource_rtx_dns_server.go to flatten []client.DNSServer to the new server block structure. Each server in the list should become a map with address and edns keys. | Restrictions: Do not change flatten logic for other fields, ensure both Read and Import use consistent logic | Success: Read and Import correctly populate server blocks with per-server EDNS values, terraform plan shows no changes when config matches state | After completing, mark task as in-progress in tasks.md before starting, use log-implementation tool to record details, then mark as complete._

- [x] 6. Update provider tests for new schema
  - File: internal/provider/resource_rtx_dns_server_test.go
  - Update test configurations to use new `server` block syntax
  - Add test case for mixed EDNS settings (first=true, second=false)
  - Verify roundtrip (create -> read -> plan with no changes)
  - Purpose: Ensure provider works correctly with new schema
  - _Leverage: existing test patterns in the same file_
  - _Requirements: 1, 3, 4_
  - _Prompt: Implement the task for spec dns-server-select-per-server-edns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with expertise in Terraform provider testing | Task: Update tests in internal/provider/resource_rtx_dns_server_test.go to use the new server block syntax. Add a test case with mixed EDNS settings (one server with edns=true, another with edns=false). Verify that the configuration round-trips correctly. | Restrictions: Do not remove existing test coverage, follow existing test patterns | Success: All tests pass, new schema is properly tested, mixed EDNS scenario is covered | After completing, mark task as in-progress in tasks.md before starting, use log-implementation tool to record details, then mark as complete._

- [x] 7. Run tests and verify no regressions
  - Run `go test ./internal/client/... ./internal/provider/...`
  - Verify parser tests still pass (no parser changes)
  - Run `go build` to ensure compilation succeeds
  - Purpose: Ensure all changes work together without regressions
  - _Leverage: existing test infrastructure_
  - _Requirements: All_
  - _Prompt: Implement the task for spec dns-server-select-per-server-edns, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer with Go testing expertise | Task: Run the test suite to verify all changes work correctly. Execute go test ./internal/client/... ./internal/provider/... and go test ./internal/rtx/parsers/... to ensure parser tests still pass. Run go build to verify compilation. Fix any failing tests or compilation errors. | Restrictions: Do not skip failing tests, ensure all packages compile | Success: All tests pass, no compilation errors, parser tests confirm no parser-layer regressions | After completing, mark task as in-progress in tasks.md before starting, use log-implementation tool to record details, then mark as complete._
