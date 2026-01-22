# Tasks Document: DNS Server Select Per-Server EDNS Support

## Task Overview

| Task | Description | Files | Requirements |
|------|-------------|-------|--------------|
| 1 | Add DNSServer struct to parser layer | parsers/dns.go | REQ-1 |
| 2 | Update DNSServerSelect struct | parsers/dns.go, interfaces.go | REQ-1 |
| 3 | Update parser for per-server EDNS | parsers/dns.go | REQ-3 |
| 4 | Update command builder | parsers/dns.go | REQ-4 |
| 5 | Add parser unit tests | parsers/dns_test.go | REQ-3 |
| 6 | Update provider schema | resource_rtx_dns_server.go | REQ-1, REQ-2 |
| 7 | Update data conversion functions | resource_rtx_dns_server.go | REQ-1, REQ-2 |
| 8 | Add provider tests | resource_rtx_dns_server_test.go | REQ-1, REQ-2 |
| 9 | Integration test with RTX | manual verification | REQ-3 |

---

## Tasks

- [x] 1. Add DNSServer struct to parser layer
  - File: `internal/rtx/parsers/dns.go`
  - Add new `DNSServer` struct with `Address` and `EDNS` fields
  - Place before `DNSServerSelect` struct definition
  - Purpose: Enable per-server EDNS storage in parser layer
  - _Leverage: Existing struct patterns in dns.go_
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec dns-server-select-edns-parsing-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in data structures | Task: Add DNSServer struct to internal/rtx/parsers/dns.go with Address (string) and EDNS (bool) fields, following existing struct patterns and JSON tags | Restrictions: Do not modify other structs yet, maintain consistent JSON tag naming, add appropriate comments | Success: DNSServer struct compiles, has proper JSON tags, follows project conventions | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 2. Update DNSServerSelect struct in both layers
  - Files: `internal/rtx/parsers/dns.go`, `internal/client/interfaces.go`
  - Change `Servers []string` to `Servers []DNSServer`
  - Remove `EDNS bool` field (now per-server)
  - Update JSON tags if needed
  - Purpose: Enable per-server EDNS in data model
  - _Leverage: DNSServer struct from task 1_
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec dns-server-select-edns-parsing-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Update DNSServerSelect struct in both parsers/dns.go and client/interfaces.go - change Servers from []string to []DNSServer, remove the EDNS bool field | Restrictions: Keep both files in sync, maintain all other fields unchanged, update comments to reflect new structure | Success: Both structs compile, Servers field uses DNSServer type, EDNS field removed | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3. Update parser for per-server EDNS parsing
  - File: `internal/rtx/parsers/dns.go`
  - Modify `parseDNSServerSelectFields` function
  - Parse `edns=on`/`edns=off` after each server IP
  - Handle both interleaved and trailing EDNS formats
  - Purpose: Correctly parse RTX output with per-server EDNS
  - _Leverage: Existing isValidIPForDNS function, current parser logic_
  - _Requirements: REQ-3_
  - _Prompt: Implement the task for spec dns-server-select-edns-parsing-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in parsing | Task: Update parseDNSServerSelectFields in parsers/dns.go to parse per-server EDNS. For each IP, check if next token is edns=on/edns=off and set accordingly. Create DNSServer{Address, EDNS} for each server. Handle format: `<server1> edns=on <server2> edns=on <type> <pattern>` | Restrictions: Must handle edns=off, must not break when edns is omitted (default false), must correctly identify record_type and query_pattern after servers | Success: Parser correctly extracts servers with per-server EDNS, record_type and query_pattern are correct | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 4. Update command builder for per-server EDNS
  - File: `internal/rtx/parsers/dns.go`
  - Modify `BuildDNSServerSelectCommand` function
  - Generate `edns=on` after each server that has EDNS enabled
  - Purpose: Generate correct RTX commands with per-server EDNS
  - _Leverage: Current builder logic_
  - _Requirements: REQ-4_
  - _Prompt: Implement the task for spec dns-server-select-edns-parsing-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Update BuildDNSServerSelectCommand in parsers/dns.go to generate per-server EDNS. For each server in sel.Servers, append server.Address, then if server.EDNS is true, append "edns=on" | Restrictions: Only add edns=on when EDNS is true (don't add edns=off), maintain correct command order: servers then type then pattern | Success: Generated command matches format `dns server select <id> <srv1> [edns=on] [<srv2> [edns=on]] [<type>] <pattern>` | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 5. Add parser unit tests for per-server EDNS
  - File: `internal/rtx/parsers/dns_test.go`
  - Update existing tests to use new struct format
  - Add tests for: interleaved EDNS, mixed EDNS, no EDNS, single server
  - Add builder tests for per-server EDNS output
  - Purpose: Ensure parser and builder work correctly
  - _Leverage: Existing test patterns in dns_test.go_
  - _Requirements: REQ-3, REQ-4_
  - _Prompt: Implement the task for spec dns-server-select-edns-parsing-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Update dns_test.go tests to use new DNSServer struct. Add test cases: (1) two servers both edns=on, (2) server1 edns=on server2 edns=off, (3) no edns specified, (4) single server with edns. Update builder tests similarly | Restrictions: Update expected values in existing tests, don't remove test coverage, follow existing test patterns | Success: All tests pass, coverage for all EDNS combinations | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 6. Update provider schema with new server block
  - File: `internal/provider/resource_rtx_dns_server.go`
  - Add deprecated markers to `servers` and `edns` attributes
  - Add new `server` nested block with `address` and `edns`
  - Set MaxItems: 2 for server block (RTX limit)
  - Purpose: Enable per-server EDNS in Terraform configuration
  - _Leverage: Existing schema patterns, validateIPAddressAny_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Implement the task for spec dns-server-select-edns-parsing-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Update server_select schema in resource_rtx_dns_server.go. Add Deprecated to existing "servers" and "edns". Add new "server" block (TypeList, MaxItems 2) with nested schema: "address" (Required, TypeString, validateIPAddressAny) and "edns" (Optional, TypeBool, Default false) | Restrictions: Keep deprecated attributes working for backward compatibility, don't change other schema fields | Success: Schema compiles, new server block available, deprecation warnings on old attributes | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 7. Update data conversion functions
  - File: `internal/provider/resource_rtx_dns_server.go`
  - Update `buildDNSConfigFromResourceData` to handle both old and new schema
  - Update `resourceRTXDNSServerRead` to populate new server blocks
  - Update `resourceRTXDNSServerImport` similarly
  - Add validation: error if both old and new schema used
  - Purpose: Enable bidirectional data conversion
  - _Leverage: Existing conversion patterns_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Implement the task for spec dns-server-select-edns-parsing-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Update data conversion in resource_rtx_dns_server.go. In buildDNSConfigFromResourceData: detect if new "server" blocks or old "servers" list is used, convert appropriately to client.DNSServerSelect with []DNSServer. In Read/Import: always populate new "server" format in state. Add validation error if both formats used | Restrictions: Must maintain backward compatibility, Read should always return new format to encourage migration | Success: Both old and new schema work, state uses new format, validation prevents mixing | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 8. Update client layer conversion functions
  - File: `internal/client/dns_service.go`
  - Update `toParserConfig` and `fromParserConfig` for new struct
  - Purpose: Sync client and parser layer struct conversions
  - _Leverage: Existing conversion patterns_
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec dns-server-select-edns-parsing-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Update toParserConfig and fromParserConfig in dns_service.go to handle new DNSServerSelect with []DNSServer. Map client.DNSServer to parsers.DNSServer and vice versa | Restrictions: Keep conversion logic symmetric, handle empty Servers slice | Success: Conversions work correctly between client and parser layers | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 9. Add provider tests for new schema
  - File: `internal/provider/resource_rtx_dns_server_test.go`
  - Add tests for new `server` block schema
  - Test deprecation warning on old schema
  - Test validation error when mixing schemas
  - Purpose: Ensure provider works correctly with new schema
  - _Leverage: Existing provider test patterns_
  - _Requirements: REQ-1, REQ-2_
  - _Prompt: Implement the task for spec dns-server-select-edns-parsing-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Add tests in resource_rtx_dns_server_test.go for: (1) Create with new server blocks, (2) Create with deprecated servers (check warning), (3) Validation error mixing old/new. Use existing test patterns | Restrictions: Don't remove existing tests, follow existing test naming conventions | Success: All new tests pass, deprecation and validation work as expected | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 10. Integration verification with RTX router
  - Manual verification step
  - Run `terraform import` on existing RTX configuration
  - Verify server_select[500100] shows correct values:
    - query_pattern = "."
    - record_type = "aaaa"
    - servers with per-server EDNS
  - Run `terraform plan` to verify no drift
  - Purpose: Confirm fix works with real RTX output
  - _Requirements: REQ-3_
  - _Prompt: Implement the task for spec dns-server-select-edns-parsing-fix, first run spec-workflow-guide to get the workflow guide then implement the task: Role: DevOps Engineer | Task: Manually verify fix with RTX router. Build provider, run terraform import rtx_dns_server.main dns, check terraform.tfstate for server_select[500100] values. Verify query_pattern=".", record_type="aaaa". Run terraform plan to confirm no unexpected changes | Restrictions: This is verification only, don't modify RTX config | Success: State shows correct values, plan shows no changes for existing config | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._
