# Tasks Document: rtx_dns_server

## Phase 1: Parser Layer

- [ ] 1. Create DNS data model and parser
  - File: internal/rtx/parsers/dns.go
  - Define DNSConfig, DNSServerSelect, and DNSHost structs
  - Implement ParseDNSConfig() to parse RTX output
  - Purpose: Parse "show config | grep dns" output
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-dns-server, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create DNSConfig struct with DomainLookup, DomainName, NameServers, ServerSelect, Hosts, ServiceOn, PrivateSpoof fields. Create DNSServerSelect struct with ID, Servers, Domains fields. Create DNSHost struct with Name, Address fields. Implement ParseDNSConfig() function to parse RTX router output from "show config | grep dns" command. Follow patterns from dhcp_scope.go | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line DNS configurations | _Leverage: internal/rtx/parsers/dhcp_scope.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (Name Servers), Requirement 3 (Static Hosts), Requirement 4 (Domain Routing) | Success: Parser correctly extracts all DNS attributes from sample RTX output, handles edge cases like missing optional fields | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 2. Create command builder functions for DNS
  - File: internal/rtx/parsers/dns.go (continue)
  - Implement BuildDNSServerCommand() for DNS server configuration
  - Implement BuildDNSServerSelectCommand() for domain-based routing
  - Implement BuildDNSStaticCommand() for static host entries
  - Implement BuildDNSServiceCommand() for DNS service on/off
  - Implement BuildDNSPrivateSpoofCommand() for private address spoofing
  - Implement BuildDeleteDNSCommand() for deletion
  - Purpose: Generate RTX CLI commands for DNS management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go BuildDHCPScopeCommand pattern_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-dns-server, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "dns server <ip1> [<ip2>]", "dns server select <id> <server> <domain>...", "dns static <hostname> <ip>", "dns service on/off", "dns private address spoof on/off", "no dns server", "no dns static <hostname>", "no dns server select <id>" | Restrictions: Follow existing command builder patterns exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirements 1-4 | Success: All commands generate valid RTX CLI syntax | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 3. Create parser unit tests
  - File: internal/rtx/parsers/dns_test.go
  - Test ParseDNSConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: missing fields, malformed input
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-dns-server, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for dns.go. Include test cases for parsing DNS config output with various configurations, command building with different parameter combinations, edge cases like empty server list, multiple server select rules, multiple static hosts | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [ ] 4. Add DNS types to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add DNSConfig struct with all fields
  - Add DNSServerSelect struct
  - Add DNSHost struct
  - Extend Client interface with DNS methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-dns-server, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add DNSConfig struct (DomainLookup bool, DomainName string, NameServers []string, ServerSelect []DNSServerSelect, Hosts []DNSHost, ServiceOn bool, PrivateSpoof bool). Add DNSServerSelect struct (ID int, Servers []string, Domains []string). Add DNSHost struct (Name string, Address string). Extend Client interface with: GetDNS(ctx) (*DNSConfig, error), CreateDNS(ctx, dns) error, UpdateDNS(ctx, dns) error, DeleteDNS(ctx) error | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 5. Create DNSService implementation
  - File: internal/client/dns_service.go (new)
  - Implement DNSService struct with executor reference
  - Implement Create() with validation and command execution
  - Implement Get() to parse DNS configuration
  - Implement Update() for modifying DNS settings
  - Implement Delete() to remove DNS configuration
  - Purpose: Service layer for DNS CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-dns-server, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create DNSService following DHCPScopeService pattern. Include input validation (IP addresses, domain names, hostnames). Use parsers.BuildDNSServerCommand and related functions. Call client.SaveConfig() after modifications. Handle DNS update by deleting old entries and creating new ones. Note: DNS is a singleton resource | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, maintain separation from other services | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-4 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 6. Integrate DNSService into rtxClient
  - File: internal/client/client.go (modify)
  - Add dnsService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface DNS methods delegating to service
  - Purpose: Wire up DNS service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-dns-server, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add dnsService *DNSService field to rtxClient. Initialize in Dial(): c.dnsService = NewDNSService(c.executor, c). Implement GetDNS, CreateDNS, UpdateDNS, DeleteDNS methods delegating to service | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all DNS methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 7. Create service unit tests
  - File: internal/client/dns_service_test.go (new)
  - Test Create with valid and invalid inputs
  - Test Get parsing
  - Test Update behavior
  - Test Delete
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-dns-server, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for DNSService. Mock Executor interface to simulate RTX responses. Test validation (invalid IPs, invalid domains, invalid hostnames). Test successful CRUD operations. Test error handling | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [ ] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_dns_server.go (new)
  - Define resourceRTXDNSServer() with full schema
  - Add domain_lookup (Optional, Bool, default true)
  - Add domain_name (Optional, String)
  - Add name_servers (Optional, List of String)
  - Add server_select (Optional, List of Object with id, servers, domains)
  - Add hosts (Optional, List of Object with name, address)
  - Add private_address_spoof (Optional, Bool, default false)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-dns-server, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXDNSServer() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Add ValidateFunc for name_servers (valid IPs), hosts (valid hostname and IP). Use TypeList for server_select with nested schema (id int, servers list, domains list). Use TypeList for hosts with nested schema (name, address strings). Note: This is a singleton resource | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1-4 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_dns_server.go (continue)
  - Implement resourceRTXDNSServerCreate()
  - Implement resourceRTXDNSServerRead()
  - Implement resourceRTXDNSServerUpdate()
  - Implement resourceRTXDNSServerDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-dns-server, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build DNSConfig from ResourceData, call client.CreateDNS, set ID to "dns" as singleton). Read (call GetDNS, update ResourceData, handle not found by clearing ID). Update (call UpdateDNS for mutable fields). Delete (call DeleteDNS). Follow rtx_dhcp_scope patterns for apiClient access and error handling | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 10. Implement import functionality
  - File: internal/provider/resource_rtx_dns_server.go (continue)
  - Implement resourceRTXDNSServerImport()
  - Accept "dns" as import ID (singleton resource)
  - Validate DNS configuration exists on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_scope.go import pattern_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec rtx-dns-server, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXDNSServerImport(). Accept "dns" as import ID (singleton resource). Call GetDNS to verify existence. Populate all ResourceData fields from retrieved configuration. Call Read to ensure state consistency | Restrictions: Handle non-existent DNS configuration errors gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go import function | _Requirements: Requirement 5 (Import) | Success: terraform import rtx_dns_server.main dns works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_dns_server" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-dns-server, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_dns_server": resourceRTXDNSServer() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 12. Create resource unit tests
  - File: internal/provider/resource_rtx_dns_server_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-dns-server, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_dns_server.go. Test schema validation (invalid IPs, invalid hostnames). Test CRUD operations with mocked client. Test import with valid and invalid IDs | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [ ] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_dns_server_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test DNS creation with name servers
  - Test DNS with server select rules
  - Test DNS with static hosts
  - Test DNS update
  - Test DNS import
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_test.go acceptance test patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-dns-server, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with DNS server creation, update name servers, add server select rules, add static hosts, import existing DNS config. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 14. Add example Terraform configurations
  - File: examples/dns_server/main.tf (new)
  - Basic DNS server configuration example
  - DNS with server select rules example
  - DNS with static hosts example
  - Full configuration example
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-dns-server, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: DNS with name_servers only. Server Select: DNS with domain-based routing. Static Hosts: DNS with static host entries. Full: DNS with all options including name_servers, server_select, hosts, private_address_spoof | Restrictions: Use realistic IP addresses and domain names, include comments explaining options | _Leverage: examples/dhcp_scope/ | _Requirements: Requirement 1 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-dns-server, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works. Check terraform import functionality. Ensure no regressions in existing resources | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
