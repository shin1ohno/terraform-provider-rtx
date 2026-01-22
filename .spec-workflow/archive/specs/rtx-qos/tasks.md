# Tasks Document: rtx_qos

## Phase 1: Parser Layer

- [x] 1. Create QoS data models and parser
  - File: internal/rtx/parsers/qos.go
  - Define QoSConfig, QoSClass structs
  - Implement ParseQoSConfig() to parse RTX output
  - Purpose: Parse "show config | grep queue" and "show queue" output
  - _Leverage: internal/rtx/parsers/dhcp_scope.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-qos, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in text parsing and regex | Task: Create QoSConfig struct with Interface, QueueType, Classes, ShapeAverage, ShapeBurst fields. Create QoSClass struct with Name, Filter, Priority, BandwidthPercent, PoliceCIR, QueueLimit fields. Implement ParseQoSConfig() function to parse RTX router output from "show config | grep queue" command. Follow patterns from dhcp_scope.go | Restrictions: Do not modify existing parser files, use standard library regexp, handle multi-line queue configurations | _Leverage: internal/rtx/parsers/dhcp_scope.go, internal/rtx/parsers/registry.go | _Requirements: Requirement 1 (CRUD), Requirement 2 (Queue Type), Requirement 3 (Traffic Classes), Requirement 4 (Shaping) | Success: Parser correctly extracts all QoS attributes from sample RTX output, handles edge cases like missing optional fields | After completing, mark task as [-] in progress, use log-implementation tool to record details, then mark as [x] complete_

- [x] 2. Create command builder functions for QoS configuration
  - File: internal/rtx/parsers/qos.go (continue)
  - Implement BuildQueueTypeCommand() for queue type
  - Implement BuildQueueClassFilterCommand() for class filter association
  - Implement BuildQueueClassPriorityCommand() for priority setting
  - Implement BuildSpeedCommand() for interface bandwidth/shaping
  - Implement BuildQueueLengthCommand() for queue depth
  - Implement BuildDeleteQoSCommand() for deletion
  - Purpose: Generate RTX CLI commands for QoS management
  - _Leverage: internal/rtx/parsers/dhcp_scope.go BuildDHCPScopeCommand pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-qos, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with RTX router CLI knowledge | Task: Create command builder functions following RTX command syntax: "queue <interface> type priority", "queue <interface> class filter <n> <filter>", "queue <interface> class priority <class> <priority>", "speed <interface> <bandwidth>", "queue <interface> length <class> <length>", "no queue <interface>". Handle bandwidth units (m, k suffixes) | Restrictions: Follow existing BuildDHCPScopeCommand pattern exactly, validate inputs before building commands | _Leverage: internal/rtx/parsers/dhcp_scope.go | _Requirements: Requirement 1 (CRUD Operations) | Success: All commands generate valid RTX CLI syntax, bandwidth unit conversion works correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 3. Create parser unit tests
  - File: internal/rtx/parsers/qos_test.go
  - Test ParseQoSConfig with various RTX output formats
  - Test all command builder functions
  - Test edge cases: missing fields, malformed input, bandwidth units
  - Purpose: Ensure parser reliability
  - _Leverage: internal/rtx/parsers/dhcp_scope_test.go for test patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-qos, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for qos.go. Include test cases for parsing queue config output, command building with various parameter combinations, edge cases like different queue types (priority, cbq), multiple classes, bandwidth units (k, m) | Restrictions: Use table-driven tests, do not require actual RTX router, use testdata fixtures if needed | _Leverage: internal/rtx/parsers/dhcp_scope_test.go | _Requirements: All functional requirements | Success: All tests pass, coverage > 80%, edge cases handled | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 2: Client Layer

- [x] 4. Add QoS types to client interfaces
  - File: internal/client/interfaces.go (modify)
  - Add QoSConfig struct with all fields
  - Add QoSClass struct
  - Extend Client interface with QoS methods
  - Purpose: Define client-level data types and interface contract
  - _Leverage: existing DHCPScope struct pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-qos, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add QoSConfig struct (Interface string, QueueType string, Classes []QoSClass, ShapeAverage int, ShapeBurst int). Add QoSClass struct (Name string, Filter int, Priority string, BandwidthPercent int, PoliceCIR int, QueueLimit int). Extend Client interface with: GetQoS(ctx, iface) (*QoSConfig, error), CreateQoS(ctx, qos) error, UpdateQoS(ctx, qos) error, DeleteQoS(ctx, iface) error, ListQoS(ctx) ([]QoSConfig, error) | Restrictions: Do not break existing interface methods, maintain backward compatibility | _Leverage: internal/client/interfaces.go existing patterns | _Requirements: Requirement 1 (CRUD) | Success: Interface compiles, follows existing patterns | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 5. Create QoSService implementation
  - File: internal/client/qos_service.go (new)
  - Implement QoSService struct with executor reference
  - Implement Create() with validation and command execution
  - Implement Get() to parse QoS configuration
  - Implement Update() for modifying queue settings
  - Implement Delete() to remove QoS configuration
  - Implement List() to retrieve all QoS configurations
  - Purpose: Service layer for QoS CRUD operations
  - _Leverage: internal/client/dhcp_scope_service.go for service pattern_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-qos, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with service layer expertise | Task: Create QoSService following DHCPScopeService pattern. Include input validation (valid interface, queue type in [priority, cbq], priority in [high, medium, normal, low], bandwidth percent sum <= 100). Use parsers.BuildQueueTypeCommand and related functions. Call client.SaveConfig() after modifications. Handle QoS update by deleting and recreating if queue type changes | Restrictions: Follow existing service patterns exactly, use containsError() for output checking, maintain separation from other services | _Leverage: internal/client/dhcp_scope_service.go | _Requirements: Requirements 1-4 | Success: All CRUD operations work, validation catches invalid input, configuration is saved | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 6. Integrate QoSService into rtxClient
  - File: internal/client/client.go (modify)
  - Add qosService field to rtxClient struct
  - Initialize service in Dial() method
  - Implement Client interface QoS methods delegating to service
  - Purpose: Wire up QoS service to main client
  - _Leverage: existing dhcpScopeService integration pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-qos, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add qosService *QoSService field to rtxClient. Initialize in Dial(): c.qosService = NewQoSService(c.executor, c). Implement GetQoS, CreateQoS, UpdateQoS, DeleteQoS, ListQoS methods delegating to service | Restrictions: Follow existing dhcpScopeService integration pattern exactly, maintain mutex locking pattern | _Leverage: internal/client/client.go dhcpScopeService integration | _Requirements: Requirement 1 | Success: Client compiles, all QoS methods delegate correctly | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 7. Create service unit tests
  - File: internal/client/qos_service_test.go (new)
  - Test Create with valid and invalid inputs
  - Test Get parsing
  - Test Update behavior
  - Test Delete
  - Mock executor for isolated testing
  - Purpose: Ensure service reliability
  - _Leverage: internal/client/dhcp_scope_service_test.go for patterns_
  - _Requirements: 1, 2, 3, 4_
  - _Prompt: Implement the task for spec rtx-qos, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for QoSService. Mock Executor interface to simulate RTX responses. Test validation (invalid interface, invalid queue type, invalid priority, bandwidth > 100%). Test successful CRUD operations. Test error handling | Restrictions: Use mock executor, do not require real router, use table-driven tests | _Leverage: internal/client/dhcp_scope_service_test.go | _Requirements: Requirements 1-4 | Success: All tests pass, validation logic tested, error paths covered | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 3: Provider Layer

- [x] 8. Create rtx_class_map Terraform resource
  - File: internal/provider/resource_rtx_class_map.go (new)
  - Define resourceRTXClassMap() with full schema
  - Add name (Required, ForceNew, String)
  - Add match_protocol (Optional, String)
  - Add match_destination_port (Optional, List of String)
  - Implement CRUD operations and import
  - Purpose: Define traffic classification rules
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 2_
  - _Prompt: Implement the task for spec rtx-qos, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXClassMap() returning *schema.Resource. Define schema: name (Required, ForceNew), match_protocol (Optional, ValidateFunc for tcp/udp/icmp), match_destination_port (Optional, List of String for port ranges). Implement Create, Read, Update, Delete, Import functions | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1, 2 | Success: Schema compiles, CRUD operations work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 9. Create rtx_policy_map Terraform resource
  - File: internal/provider/resource_rtx_policy_map.go (new)
  - Define resourceRTXPolicyMap() with full schema
  - Add name (Required, ForceNew, String)
  - Add class blocks (Required, List of Object with name, priority, bandwidth_percent, police_cir, queue_limit)
  - Implement CRUD operations and import
  - Purpose: Define QoS policy with class actions
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 3_
  - _Prompt: Implement the task for spec rtx-qos, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXPolicyMap() returning *schema.Resource. Define schema: name (Required, ForceNew), class (Required, TypeList with nested schema: name string, priority bool, bandwidth_percent int, police_cir int, queue_limit int). Implement Create, Read, Update, Delete, Import functions. Validate bandwidth_percent sum <= 100 | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1, 3 | Success: Schema compiles, nested class blocks work | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 10. Create rtx_service_policy Terraform resource
  - File: internal/provider/resource_rtx_service_policy.go (new)
  - Define resourceRTXServicePolicy() with full schema
  - Add interface (Required, ForceNew, String)
  - Add direction (Required, String: input/output)
  - Add policy_map (Required, String)
  - Implement CRUD operations and import
  - Purpose: Apply policy to interface
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-qos, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXServicePolicy() returning *schema.Resource. Define schema: interface (Required, ForceNew), direction (Required, ValidateFunc for input/output), policy_map (Required, reference to policy map name). Implement Create, Read, Update, Delete, Import functions. Resource ID format: interface:direction | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirement 1 | Success: Schema compiles, policy attachment works | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 11. Create rtx_shape Terraform resource
  - File: internal/provider/resource_rtx_shape.go (new)
  - Define resourceRTXShape() with full schema
  - Add interface (Required, ForceNew, String)
  - Add direction (Required, String: input/output)
  - Add shape_average (Required, Int - bps)
  - Add shape_burst (Optional, Int - bps)
  - Implement CRUD operations and import
  - Purpose: Traffic shaping configuration
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go for patterns_
  - _Requirements: 1, 4_
  - _Prompt: Implement the task for spec rtx-qos, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Create resourceRTXShape() returning *schema.Resource. Define schema: interface (Required, ForceNew), direction (Required, ValidateFunc for input/output), shape_average (Required, positive int), shape_burst (Optional, positive int). Implement Create, Read, Update, Delete, Import functions. Resource ID format: interface:direction | Restrictions: Follow Terraform SDK v2 patterns, match existing resource style | _Leverage: internal/provider/resource_rtx_dhcp_scope.go | _Requirements: Requirements 1, 4 | Success: Schema compiles, shaping configuration works | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 12. Register QoS resources in provider
  - File: internal/provider/provider.go (modify)
  - Add "rtx_class_map" to ResourcesMap
  - Add "rtx_policy_map" to ResourcesMap
  - Add "rtx_service_policy" to ResourcesMap
  - Add "rtx_shape" to ResourcesMap
  - Purpose: Make QoS resources available to Terraform
  - _Leverage: existing resource registrations_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec rtx-qos, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Add entries to ResourcesMap in provider.go: "rtx_class_map": resourceRTXClassMap(), "rtx_policy_map": resourceRTXPolicyMap(), "rtx_service_policy": resourceRTXServicePolicy(), "rtx_shape": resourceRTXShape() | Restrictions: Do not modify other resource entries, maintain alphabetical order if present | _Leverage: internal/provider/provider.go existing pattern | _Requirements: Requirement 1 | Success: Provider compiles with all QoS resources registered | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 13. Create resource unit tests
  - File: internal/provider/resource_rtx_class_map_test.go (new)
  - File: internal/provider/resource_rtx_policy_map_test.go (new)
  - File: internal/provider/resource_rtx_service_policy_test.go (new)
  - File: internal/provider/resource_rtx_shape_test.go (new)
  - Test schema validation for each resource
  - Test CRUD operations with mock client
  - Test import functionality
  - Purpose: Ensure resource reliability
  - _Leverage: resource_rtx_dhcp_scope_test.go patterns_
  - _Requirements: 1, 5_
  - _Prompt: Implement the task for spec rtx-qos, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer with Terraform SDK experience | Task: Create unit tests for all QoS resource files. Test schema validation (invalid protocols, bandwidth > 100, negative shaping values). Test CRUD operations with mocked client. Test import with valid and invalid IDs | Restrictions: Use terraform-plugin-sdk testing utilities, mock API client | _Leverage: internal/provider/resource_rtx_dhcp_scope_test.go | _Requirements: Requirements 1, 5 | Success: All tests pass, good coverage of validation and CRUD paths | After completing, use log-implementation tool to record details, then mark as [x] complete_

## Phase 4: Integration and Documentation

- [x] 14. Create acceptance tests
  - File: internal/provider/resource_rtx_class_map_acc_test.go (new)
  - File: internal/provider/resource_rtx_policy_map_acc_test.go (new)
  - File: internal/provider/resource_rtx_service_policy_acc_test.go (new)
  - File: internal/provider/resource_rtx_shape_acc_test.go (new)
  - Test full lifecycle with real RTX router (when TF_ACC=1)
  - Test resource creation with all parameters
  - Test resource updates
  - Test resource import
  - Test dependencies between resources
  - Purpose: Validate end-to-end functionality
  - _Leverage: resource_rtx_dhcp_scope_acc_test.go acceptance test patterns_
  - _Requirements: 1, 5, 6_
  - _Prompt: Implement the task for spec rtx-qos, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Acceptance Test Developer | Task: Create acceptance tests using terraform-plugin-sdk/v2/helper/resource. Test config with class_map creation, policy_map with classes, service_policy attachment, shape configuration. Test resource dependencies (service_policy depends_on policy_map depends_on class_map). Use TF_ACC environment check | Restrictions: Tests require real RTX router, use skip if TF_ACC not set | _Leverage: existing acceptance test patterns | _Requirements: Requirements 1, 5, 6 | Success: Acceptance tests pass against real RTX router | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 15. Add example Terraform configurations
  - File: examples/qos/main.tf (new)
  - Basic class map and policy map example
  - Complete QoS configuration with all resources
  - Traffic shaping example
  - Purpose: User documentation and testing
  - _Leverage: examples/dhcp_scope/ existing examples_
  - _Requirements: 1, 6_
  - _Prompt: Implement the task for spec rtx-qos, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer with Terraform expertise | Task: Create example Terraform configurations. Basic: class_map for VoIP traffic. Full: class_map + policy_map + service_policy showing priority queuing for VoIP. Shaping: rtx_shape for bandwidth limiting. Include resource dependencies using depends_on | Restrictions: Use realistic IP addresses and ports, include comments explaining QoS concepts | _Leverage: examples/dhcp_scope/ | _Requirements: Requirements 1, 6 | Success: Examples are valid Terraform, demonstrate all features | After completing, use log-implementation tool to record details, then mark as [x] complete_

- [x] 16. Final integration testing and cleanup
  - Run all tests: go test ./...
  - Run acceptance tests: TF_ACC=1 go test ./... -v
  - Build provider: go build -o terraform-provider-rtx
  - Test examples manually
  - Purpose: Ensure everything works together
  - _Leverage: Makefile or existing build scripts_
  - _Requirements: All_
  - _Prompt: Implement the task for spec rtx-qos, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, fix any failures. Build provider and test with examples. Verify terraform plan/apply/destroy cycle works for all QoS resources. Check terraform import functionality. Ensure no regressions in existing resources | Restrictions: All tests must pass, no compiler warnings | _Leverage: Makefile, existing CI scripts | _Requirements: All requirements | Success: All tests pass, provider builds cleanly, examples work end-to-end | After completing, use log-implementation tool to record details, then mark as [x] complete_
