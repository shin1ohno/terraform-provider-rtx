# Tasks Document: Unified ACL Apply Design

## Phase 1: Shared Components

- [x] 1. Create ACL schema common utilities
  - File: internal/provider/acl_schema_common.go
  - Implement CommonACLSchema(), CommonApplySchema(), CommonEntrySchema()
  - Add BuildACLFromResourceData() helper
  - Add ValidateACLSchema() for CustomizeDiff
  - Purpose: Provide shared schema definitions for all ACL resources
  - _Leverage: internal/provider/resource_rtx_access_list_mac.go (reference implementation)_
  - _Requirements: 1, 7_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in Terraform provider development | Task: Create acl_schema_common.go with shared schema definitions for all ACL types, including CommonACLSchema() returning name/sequence_start/sequence_step/apply attributes, CommonApplySchema() for apply blocks, and ValidateACLSchema() for mode validation | Restrictions: Do not modify existing ACL resources yet, follow existing code patterns in the provider, use terraform-plugin-sdk/v2 | _Leverage: internal/provider/resource_rtx_access_list_mac.go for apply block reference | _Requirements: Requirement 1 (auto sequence), Requirement 7 (consistent schema) | Success: All functions compile, schema matches design spec, validation correctly detects auto/manual mode mixing | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

- [x] 2. Create sequence calculator
  - File: internal/provider/sequence_calculator.go
  - Implement CalculateSequences(start, step, count int) []int
  - Implement ValidateSequenceRange() for overflow detection
  - Implement DetectSequenceMode() to determine auto vs manual
  - Purpose: Centralize sequence calculation logic
  - _Leverage: None (pure functions)_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Create sequence_calculator.go with pure functions for sequence calculation: CalculateSequences() returns slice of calculated sequences, ValidateSequenceRange() checks for overflow, DetectSequenceMode() returns SequenceMode enum | Restrictions: Keep functions pure with no side effects, handle edge cases (step=0, overflow), add comprehensive documentation | _Leverage: None | _Requirements: Requirement 1 | Success: All functions have unit tests, edge cases handled, no panics possible | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

- [x] 3. Create sequence calculator tests
  - File: internal/provider/sequence_calculator_test.go
  - Test CalculateSequences with various inputs
  - Test edge cases: step=0, large values, overflow
  - Test DetectSequenceMode
  - Purpose: Ensure sequence calculation reliability
  - _Leverage: internal/provider/sequence_calculator.go_
  - _Requirements: 1, 8_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create comprehensive unit tests for sequence_calculator.go covering normal cases (start=100, step=10, count=3), edge cases (step=0, count=0, overflow), and mode detection | Restrictions: Use standard Go testing package, test both success and error cases, achieve high coverage | _Leverage: internal/provider/sequence_calculator.go | _Requirements: Requirement 1, Requirement 8 (CRUD testing) | Success: All tests pass, edge cases covered, >90% coverage | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

- [x] 4. Create collision validator
  - File: internal/provider/collision_validator.go
  - Implement ValidateNoCollision() for Plan-time state-based validation
  - Implement CheckRouterCollision() for Apply-time router-based validation
  - Define CollisionError type
  - Purpose: Detect sequence conflicts between ACL resources
  - _Leverage: internal/provider/acl_schema_common.go, internal/client/client.go_
  - _Requirements: 2_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in Terraform provider validation | Task: Create collision_validator.go with ValidateNoCollision() that queries Terraform state for other ACLs and compares sequence ranges, CheckRouterCollision() that queries router for existing filters, and CollisionError type for detailed error reporting | Restrictions: ValidateNoCollision must work in CustomizeDiff context, CheckRouterCollision requires RTXClient, return clear error messages | _Leverage: internal/provider/acl_schema_common.go, internal/client/client.go | _Requirements: Requirement 2 | Success: Collisions detected correctly, error messages are actionable, no false positives | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

## Phase 2: Service Layer

- [x] 5. Create ACL service
  - File: internal/client/acl_service.go
  - Implement CreateACLEntries() for batch entry creation
  - Implement ReadACLEntries() for group reading
  - Implement UpdateACLEntries() for entry updates
  - Implement DeleteACLEntries() for batch deletion
  - Implement GetAllFilterSequences() for collision detection
  - Purpose: Unified service for ACL entry management
  - _Leverage: internal/client/ethernet_filter_service.go, internal/client/ip_filter_service.go_
  - _Requirements: 1, 4_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in service layer architecture | Task: Create acl_service.go with CRUD methods for ACL entries, supporting all ACL types (ip, ipv6, mac, extended) through aclType parameter, using command templates for different filter types | Restrictions: Reuse existing SSH execution patterns, handle errors gracefully, support batched operations | _Leverage: internal/client/ethernet_filter_service.go, internal/client/ip_filter_service.go | _Requirements: Requirement 1, Requirement 4 | Success: All CRUD operations work for all ACL types, proper error handling, efficient batching | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

- [x] 6. Create ACL apply service
  - File: internal/client/acl_apply_service.go
  - Implement ApplyFiltersToInterface() with command mapping
  - Implement RemoveFiltersFromInterface()
  - Implement GetInterfaceFilters() for reading current state
  - Implement ValidateInterface() for interface/ACL type compatibility
  - Purpose: Manage filter-interface bindings
  - _Leverage: internal/client/ethernet_filter_service.go (existing apply logic)_
  - _Requirements: 3, 5_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Create acl_apply_service.go with methods to apply/remove filters from interfaces, using InterfaceFilterCommands mapping for different ACL types (ip: "ip %s secure filter", ipv6: "ipv6 %s secure filter", mac: "ethernet %s filter") | Restrictions: Validate interface exists before applying, check ACL type compatibility (no MAC on PP), handle partial failures | _Leverage: internal/client/ethernet_filter_service.go | _Requirements: Requirement 3, Requirement 5 | Success: All interface types supported, proper validation, clean error messages | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

## Phase 3: Update Existing ACL Resources

- [x] 7. Update rtx_access_list_mac for multiple apply
  - File: internal/provider/resource_rtx_access_list_mac.go
  - Change apply from MaxItems=1 to unlimited
  - Add sequence_start and sequence_step attributes
  - Integrate with shared components
  - Update CRUD handlers to use new services
  - Purpose: Extend MAC ACL to unified design
  - _Leverage: internal/provider/acl_schema_common.go, internal/client/acl_service.go_
  - _Requirements: 1, 3, 7_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Update resource_rtx_access_list_mac.go to support multiple apply blocks (remove MaxItems=1), add sequence_start/sequence_step attributes, integrate ValidateACLSchema in CustomizeDiff, update CRUD to use acl_service and acl_apply_service | Restrictions: Maintain backward compatibility for existing single-apply configs, preserve existing entry schema | _Leverage: internal/provider/acl_schema_common.go, internal/client/acl_service.go | _Requirements: Requirement 1, 3, 7 | Success: Multiple applies work, auto sequence works, existing tests pass | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

- [x] 8. Update rtx_access_list_extended with apply block
  - File: internal/provider/resource_rtx_access_list_extended.go
  - Add apply block schema from CommonApplySchema()
  - Add sequence_start and sequence_step attributes
  - Add CustomizeDiff for validation
  - Update CRUD handlers
  - Purpose: Add apply capability to Extended ACL
  - _Leverage: internal/provider/acl_schema_common.go, internal/client/acl_apply_service.go_
  - _Requirements: 1, 3, 7_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Update resource_rtx_access_list_extended.go to add apply block using CommonApplySchema(), add sequence_start/sequence_step, integrate validation in CustomizeDiff, update Create/Read/Update/Delete to handle applies | Restrictions: Keep existing entry schema intact, apply is optional | _Leverage: internal/provider/acl_schema_common.go, internal/client/acl_apply_service.go | _Requirements: Requirement 1, 3, 7 | Success: Apply blocks work, auto sequence works, existing functionality preserved | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

- [x] 9. Redesign rtx_access_list_ip as group resource
  - File: internal/provider/resource_rtx_access_list_ip.go
  - Change from individual filter to group-based design
  - Add name attribute as identifier
  - Add entry block for multiple filters
  - Add apply block
  - Add sequence_start and sequence_step
  - Update CRUD handlers completely
  - Purpose: Convert IP ACL to unified group design
  - _Leverage: internal/provider/acl_schema_common.go, internal/provider/resource_rtx_access_list_extended.go_
  - _Requirements: 1, 3, 4, 7_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Completely redesign resource_rtx_access_list_ip.go from individual filter (sequence as ID) to group resource (name as ID), add entry block with sequence/action/source/destination/protocol attributes, add apply block, implement new CRUD handlers | Restrictions: This is a breaking change, do not attempt backward compatibility | _Leverage: internal/provider/acl_schema_common.go, internal/provider/resource_rtx_access_list_extended.go as reference | _Requirements: Requirement 1, 3, 4, 7 | Success: Group-based design works, multiple entries supported, apply works | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

- [x] 10. Redesign rtx_access_list_ipv6 as group resource
  - File: internal/provider/resource_rtx_access_list_ipv6.go
  - Change from individual filter to group-based design
  - Add name attribute as identifier
  - Add entry block for multiple filters
  - Add apply block
  - Add sequence_start and sequence_step
  - Update CRUD handlers completely
  - Purpose: Convert IPv6 ACL to unified group design
  - _Leverage: internal/provider/acl_schema_common.go, internal/provider/resource_rtx_access_list_ip.go_
  - _Requirements: 1, 3, 4, 7_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Completely redesign resource_rtx_access_list_ipv6.go from individual filter to group resource, following the same pattern as rtx_access_list_ip, with entry block supporting IPv6-specific attributes (icmp6 protocol) | Restrictions: This is a breaking change, do not attempt backward compatibility | _Leverage: internal/provider/acl_schema_common.go, internal/provider/resource_rtx_access_list_ip.go | _Requirements: Requirement 1, 3, 4, 7 | Success: Group-based design works, IPv6-specific features preserved | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

## Phase 4: Separate Apply Resources

- [x] 11. Create rtx_access_list_ip_apply resource
  - File: internal/provider/resource_rtx_access_list_ip_apply.go
  - Schema: access_list, interface, direction, filter_ids (optional)
  - Implement CRUD handlers
  - Add conflict detection with inline applies
  - Purpose: Separate resource for IP ACL interface binding
  - _Leverage: internal/client/acl_apply_service.go_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Create resource_rtx_access_list_ip_apply.go with access_list/interface/direction/filter_ids schema, ID format "interface:direction", CRUD using acl_apply_service, conflict detection in CustomizeDiff checking for inline applies | Restrictions: ForceNew on interface and direction, filter_ids optional (defaults to all) | _Leverage: internal/client/acl_apply_service.go | _Requirements: Requirement 5 | Success: Resource works with for_each, conflicts detected, proper state management | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

- [x] 12. Create rtx_access_list_ipv6_apply resource
  - File: internal/provider/resource_rtx_access_list_ipv6_apply.go
  - Same pattern as IP apply
  - Purpose: Separate resource for IPv6 ACL interface binding
  - _Leverage: internal/provider/resource_rtx_access_list_ip_apply.go_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Create resource_rtx_access_list_ipv6_apply.go following the same pattern as rtx_access_list_ip_apply, using ACLTypeIPv6 for service calls | Restrictions: Consistent with IP apply resource | _Leverage: internal/provider/resource_rtx_access_list_ip_apply.go | _Requirements: Requirement 5 | Success: Resource works identically to IP apply | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

- [x] 13. Create rtx_access_list_mac_apply resource
  - File: internal/provider/resource_rtx_access_list_mac_apply.go
  - Same pattern as IP apply
  - Add validation: no PP/Tunnel interfaces for MAC
  - Purpose: Separate resource for MAC ACL interface binding
  - _Leverage: internal/provider/resource_rtx_access_list_ip_apply.go_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Create resource_rtx_access_list_mac_apply.go following IP apply pattern, add interface validation to reject PP and Tunnel interfaces (MAC not supported) | Restrictions: Validate interface type in CustomizeDiff | _Leverage: internal/provider/resource_rtx_access_list_ip_apply.go | _Requirements: Requirement 5 | Success: Resource works, PP/Tunnel rejected with clear error | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

## Phase 5: Remove ACL from Interface Resources

- [x] 14. Remove ACL attributes from rtx_interface
  - File: internal/provider/resource_rtx_interface.go
  - Remove access_list_ip_in, access_list_ip_out
  - Remove access_list_ipv6_in, access_list_ipv6_out
  - Remove access_list_ip_dynamic_in, access_list_ip_dynamic_out
  - Remove access_list_ipv6_dynamic_in, access_list_ipv6_dynamic_out
  - Remove access_list_mac_in, access_list_mac_out
  - Update CRUD handlers to not read/write ACL state
  - Purpose: Single source of truth for ACL bindings
  - _Leverage: None_
  - _Requirements: 6_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Remove all access_list_* attributes from resource_rtx_interface.go schema, remove related code from Create/Read/Update/Delete handlers, update buildInterfaceConfigFromResourceData | Restrictions: This is a breaking change, clean removal only | _Leverage: None | _Requirements: Requirement 6 | Success: No ACL attributes in schema, CRUD works without ACL handling | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

- [x] 15. Remove ACL attributes from rtx_pp_interface
  - File: internal/provider/resource_rtx_pp_interface.go
  - Remove access_list_ip_in, access_list_ip_out
  - Update CRUD handlers
  - Purpose: Single source of truth for ACL bindings
  - _Leverage: None_
  - _Requirements: 6_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Remove access_list_ip_in and access_list_ip_out from resource_rtx_pp_interface.go, update CRUD handlers | Restrictions: Clean removal | _Leverage: None | _Requirements: Requirement 6 | Success: No ACL attributes, tests updated | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

- [x] 16. Remove ACL attributes from rtx_ipv6_interface
  - File: internal/provider/resource_rtx_ipv6_interface.go
  - Remove any ACL-related attributes
  - Update CRUD handlers
  - Purpose: Single source of truth for ACL bindings
  - _Leverage: None_
  - _Requirements: 6_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Remove any ACL-related attributes from resource_rtx_ipv6_interface.go, update CRUD handlers | Restrictions: Clean removal | _Leverage: None | _Requirements: Requirement 6 | Success: No ACL attributes | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

## Phase 6: Register New Resources

- [x] 17. Register new resources in provider
  - File: internal/provider/provider.go
  - Register rtx_access_list_ip_apply
  - Register rtx_access_list_ipv6_apply
  - Register rtx_access_list_mac_apply
  - Purpose: Make new resources available
  - _Leverage: internal/provider/provider.go_
  - _Requirements: 5_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Add new apply resources to ResourcesMap in provider.go | Restrictions: Follow existing naming conventions | _Leverage: internal/provider/provider.go | _Requirements: Requirement 5 | Success: All new resources available via terraform | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

## Phase 7: Testing

- [x] 18. Create ACL common component tests
  - File: internal/provider/acl_schema_common_test.go
  - Test CommonACLSchema()
  - Test ValidateACLSchema() with various scenarios
  - Purpose: Ensure shared components work correctly
  - _Leverage: internal/provider/acl_schema_common.go_
  - _Requirements: 8_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for acl_schema_common.go testing schema generation and validation logic | Restrictions: Mock ResourceDiff for validation tests | _Leverage: internal/provider/acl_schema_common.go | _Requirements: Requirement 8 | Success: High coverage, all validation cases tested | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

- [x] 19. Create collision validator tests
  - File: internal/provider/collision_validator_test.go
  - Test no collision scenarios
  - Test collision detection scenarios
  - Test error message formatting
  - Purpose: Ensure collision detection reliability
  - _Leverage: internal/provider/collision_validator.go_
  - _Requirements: 2, 8_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create unit tests for collision_validator.go testing non-overlapping ranges pass, overlapping ranges fail, error messages contain useful info | Restrictions: Mock state queries | _Leverage: internal/provider/collision_validator.go | _Requirements: Requirement 2, 8 | Success: All collision scenarios covered | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

- [x] 20. Create ACL resource acceptance tests
  - File: internal/provider/resource_rtx_access_list_extended_test.go (update)
  - Add TestAccRTXAccessListExtended_AutoSequence
  - Add TestAccRTXAccessListExtended_MultipleApply
  - Add TestAccRTXAccessListExtended_UpdateAddEntry
  - Add TestAccRTXAccessListExtended_Import
  - Purpose: End-to-end testing of ACL resources
  - _Leverage: Existing test patterns in the codebase_
  - _Requirements: 8_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create/update acceptance tests for rtx_access_list_extended covering auto sequence, multiple apply, entry updates, and import | Restrictions: Use TF_ACC=1 pattern, follow existing test conventions | _Leverage: Existing test patterns | _Requirements: Requirement 8 | Success: All acceptance tests pass against real/simulated router | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

- [x] 21. Create apply resource acceptance tests
  - File: internal/provider/resource_rtx_access_list_ip_apply_test.go
  - Test basic apply
  - Test for_each pattern
  - Test conflict with inline apply
  - Purpose: End-to-end testing of apply resources
  - _Leverage: internal/provider/resource_rtx_access_list_ip_apply.go_
  - _Requirements: 5, 8_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Create acceptance tests for rtx_access_list_ip_apply testing basic apply, for_each usage, and conflict detection | Restrictions: Follow existing test patterns | _Leverage: internal/provider/resource_rtx_access_list_ip_apply.go | _Requirements: Requirement 5, 8 | Success: All tests pass | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

## Phase 8: Documentation

- [x] 22. Update ACL resource documentation
  - Files: docs/resources/access_list_*.md
  - Document new schema with examples
  - Document auto vs manual sequence mode
  - Document multiple apply blocks
  - Purpose: User documentation
  - _Leverage: Existing docs structure_
  - _Requirements: All_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer | Task: Update documentation for all ACL resources showing new schema, auto/manual sequence examples, multiple apply examples | Restrictions: Run tfplugindocs to generate, then enhance with examples | _Leverage: Existing docs structure | _Requirements: All | Success: Clear, complete documentation with working examples | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_

- [x] 23. Create migration guide
  - File: docs/guides/acl-migration.md
  - Document breaking changes
  - Provide before/after examples
  - Step-by-step migration instructions
  - Purpose: Help existing users migrate
  - _Leverage: None_
  - _Requirements: All_
  - _Prompt: Implement the task for spec unified-acl-apply, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer | Task: Create migration guide documenting breaking changes, showing before/after terraform configs, providing step-by-step migration instructions for each ACL type | Restrictions: Be thorough about state migration | _Leverage: None | _Requirements: All | Success: Users can successfully migrate existing configs | Instructions: Mark task as [-] in tasks.md when starting, use log-implementation tool after completion, mark as [x] when done_
