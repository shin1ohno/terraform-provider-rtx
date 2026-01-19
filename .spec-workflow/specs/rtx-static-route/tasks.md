# Tasks Document: rtx_static_route

## Phase 1: Parser Layer

- [ ] 1. Create StaticRoute data model and parser
  - File: internal/rtx/parsers/static_route.go
  - Define StaticRoute and NextHop structs
  - Implement ParseRouteConfig() to parse RTX output
  - Purpose: Parse "show config | grep ip route" output
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-static-route, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create StaticRoute struct with Prefix, Mask, NextHops fields. Create NextHop struct with NextHop, Interface, Distance, Name, Permanent, Filter fields. Implement ParseRouteConfig() function to parse RTX router output from "show config | grep ip route" command. Handle various route formats including gateway IP, pp interface, and tunnel interface | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-path routes with same prefix | _Leverage: internal/rtx/parsers/dhcp_scope.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (Multi-path), Requirement 3 (Interface types) | Success: Parser correctly extracts all route attributes from sample RTX output, handles multiple next hops for same prefix | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 2. Create command builder functions for static routes
  - File: internal/rtx/parsers/static_route.go (continue)
  - Implement BuildIPRouteCommand() for route creation
  - Implement BuildDeleteIPRouteCommand() for route deletion
  - Implement BuildShowIPRouteCommand() for reading routes
  - Handle gateway, pp, tunnel, and dhcp interface types
  - Purpose: Generate RTX CLI commands for route management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go BuildDHCPScopeCommand pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-static-route, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "ip route <network> gateway <gateway> [weight <n>] [hide] [filter <n>]", "ip route <network> gateway pp <n>", "ip route <network> gateway tunnel <n>", "no ip route <network>". Handle special case where prefix=0.0.0.0 and mask=0.0.0.0 maps to "ip route default" | Restrictions: Follow existing BuildDHCPScopeCommand pattern exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax, default route conversion works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 3. Create parser unit tests
  - File: internal/rtx/parsers/static_route_test.go
  - Test ParseRouteConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: default route, multi-hop, interface routes
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3_
  - _Prompt: Implement the task for spec rtx-static-route, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for static_route.go. Include test cases for parsing route config output, command building with various parameter combinations, edge cases like default route (0.0.0.0/0), multi-hop routes with different weights, tunnel and pp interface routes | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [ ] 4. Add StaticRoute type to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add StaticRoute struct with all fields
  - Add NextHop struct
  - Extend Client interface with route methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-static-route, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add StaticRoute struct (Prefix string, Mask string, NextHops []NextHop). Add NextHop struct (NextHop string, Interface string, Distance int, Name string, Permanent bool, Filter int). Extend Client interface with: GetStaticRoute(ctx, network) (*StaticRoute, error), CreateStaticRoute(ctx, route) error, UpdateStaticRoute(ctx, route) error, DeleteStaticRoute(ctx, network) error, ListStaticRoutes(ctx) ([]StaticRoute, error) | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 5. Create StaticRouteService implementation
  - File: internal/client/static_route_service.go (new)
  - Implement StaticRouteService struct with executor reference
  - Implement CreateRoute() with validation and multi-hop command execution
  - Implement GetRoute() to parse route configuration by network
  - Implement UpdateRoute() by deleting and recreating all next hops
  - Implement DeleteRoute() to remove all next hops for a network
  - Implement ListRoutes() to retrieve all static routes
  - Purpose: Service layer for route CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-static-route, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create StaticRouteService following DHCPScopeService pattern. Include input validation (IP address format, mask format, valid interface names). Each next_hop becomes a separate RTX command. Handle update by deleting all hops then recreating. Use parsers.BuildIPRouteCommand and related functions. Call client.SaveConfig() after modifications | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, maintain separation from other services | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-4 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 6. Integrate StaticRouteService into rtxClient
  - File: internal/client/client.go (modify)
  - Add staticRouteService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface route methods delegating to service
  - Purpose: Wire up route service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-static-route, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add staticRouteService *StaticRouteService field to rtxClient. Initialize in Dial(): c.staticRouteService = NewStaticRouteService(c.executor, c). Implement GetStaticRoute, CreateStaticRoute, UpdateStaticRoute, DeleteStaticRoute, ListStaticRoutes methods delegating to service | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all route methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 7. Create service unit tests
  - File: internal/client/static_route_service_test.go (new)
  - Test CreateRoute with valid and invalid inputs
  - Test GetRoute parsing for single and multi-hop routes
  - Test UpdateRoute behavior with hop changes
  - Test DeleteRoute
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-static-route, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for StaticRouteService. Mock Executor interface to simulate RTX responses. Test validation (invalid IP, invalid mask, invalid interface name). Test successful CRUD operations including multi-hop scenarios. Test error handling for route already exists, route not found | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [ ] 8. Create Terraform resource schema
  - File: internal/provider/resource_rtx_static_route.go (new)
  - Define resourceRTXStaticRoute() with full schema
  - Add prefix (Required, ForceNew, String with IP validation)
  - Add mask (Required, ForceNew, String with IP validation)
  - Add next_hops (Required, List of Object)
    - next_hop (Optional, String - gateway IP)
    - interface (Optional, String - pp, tunnel, etc.)
    - distance (Optional, Int, default 1)
    - name (Optional, String)
    - permanent (Optional, Bool, default false)
    - filter (Optional, Int)
  - Purpose: Define Terraform resource structure
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-static-route, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXStaticRoute() returning *schema.Resource. Define schema following rtx_dhcp_scope patterns. Add ValidateFunc for prefix and mask (valid IP addresses). Set ForceNew on prefix and mask. Use TypeList for next_hops with nested schema containing next_hop, interface, distance, name, permanent, filter. Require at least one of next_hop or interface in each hop | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1-4 | Success: Schema compiles, validation functions work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 9. Implement CRUD operations for resource
  - File: internal/provider/resource_rtx_static_route.go (continue)
  - Implement resourceRTXStaticRouteCreate()
  - Implement resourceRTXStaticRouteRead()
  - Implement resourceRTXStaticRouteUpdate()
  - Implement resourceRTXStaticRouteDelete()
  - Purpose: Terraform lifecycle management
  - _Leverage: resource_rtx_dhcp_scope.go CRUD patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-static-route, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement Create (build StaticRoute from ResourceData including all next_hops, call client.CreateStaticRoute, set ID to prefix/mask). Read (call GetStaticRoute, update ResourceData including next_hops list, handle not found by clearing ID). Update (call UpdateStaticRoute for next_hops changes). Delete (call DeleteStaticRoute). Use prefix/mask as Terraform resource ID (e.g., "0.0.0.0/0.0.0.0" for default route) | Restrictions: Use diag.Diagnostics for errors, handle partial failures gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: All CRUD operations work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 10. Implement import functionality
  - File: internal/provider/resource_rtx_static_route.go (continue)
  - Implement resourceRTXStaticRouteImport()
  - Parse prefix/mask from import ID string
  - Validate route exists on router
  - Purpose: Support terraform import command
  - _Leverage: resource_rtx_dhcp_scope.go import pattern_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec rtx-static-route, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Implement resourceRTXStaticRouteImport(). Parse import ID as "prefix/mask" format (e.g., "10.0.0.0/255.0.0.0" or "0.0.0.0/0.0.0.0" for default). Call GetStaticRoute to verify existence. Populate all ResourceData fields including next_hops from retrieved route. Call Read to ensure state consistency | Restrictions: Handle invalid import ID format, non-existent route errors gracefully | _Leverage: internal/provider/resource_rtx_dhcp_scope.go import function | _Requirements: Requirement 5 (Import) | Success: terraform import rtx_static_route.example "10.0.0.0/255.0.0.0" works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 11. Register resource in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_static_route" to ResourcesMap
  - Purpose: Make resource available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-static-route, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entry to ResourcesMap in provider.go: "rtx_static_route": resourceRTXStaticRoute() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with new resource registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 12. Create resource unit tests
  - File: internal/provider/resource_rtx_static_route_test.go (new)
  - Test schema validation
  - Test CRUD operations with mock client
  - Test import functionality
  - Test multi-hop route handling
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-static-route, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for resource_rtx_static_route.go. Test schema validation (invalid IP prefix, invalid mask). Test CRUD operations with mocked client including multi-hop scenarios. Test import with valid and invalid IDs (prefix/mask format) | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [ ] 13. Create acceptance tests
  - File: internal/provider/resource_rtx_static_route_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test default route creation (0.0.0.0/0.0.0.0)
  - Test network route with multiple next hops
  - Test tunnel/PP interface routes
  - Test route update (weight/distance changes)
  - Test route import
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_acc_test.go acceptance test patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-static-route, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with default route creation, multi-hop network route, tunnel interface route, update route weights, import existing route. Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 14. Add example Terraform configurations
  - File: examples/static_route/main.tf (new)
  - Default route example with primary and backup gateways
  - Network route with multiple next hops example
  - Tunnel interface route example
  - PP interface route example
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-static-route, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Default route: 0.0.0.0/0.0.0.0 with primary (distance 1) and backup (distance 10) gateways. Network route: 10.0.0.0/255.0.0.0 with multi-path routing. Tunnel route: 172.16.0.0/255.255.0.0 via tunnel 1. Include comments explaining distance/weight for load balancing and failover | Restrictions: Use realistic IP addresses, include comments explaining options | _Leverage: examples/dhcp_scope/ | _Requirements: Requirement 1 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [ ] 15. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-static-route, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works for all route types (default, network, tunnel, pp). Check terraform import functionality with prefix/mask format. Ensure no regressions in existing resources | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
