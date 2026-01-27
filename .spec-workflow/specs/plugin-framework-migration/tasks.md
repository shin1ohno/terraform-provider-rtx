# Tasks Document: Plugin Framework Migration

## Phase 1: Infrastructure Setup

- [ ] 1.1. Add Plugin Framework dependencies to go.mod
  - File: go.mod
  - Add terraform-plugin-framework, terraform-plugin-testing dependencies
  - Remove terraform-plugin-sdk/v2 references (after migration complete)
  - Purpose: Enable Framework development
  - _Leverage: existing go.mod structure_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with Terraform provider experience | Task: Update go.mod to add terraform-plugin-framework v1.x and terraform-plugin-testing dependencies following requirement 1 | Restrictions: Do not remove SDK v2 yet (needed during migration), use latest stable versions | Success: go mod tidy succeeds, both Framework and SDK v2 compile | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts (dependencies added), then mark as [x]_

- [ ] 1.2. Create new directory structure
  - Files: internal/provider/resources/, internal/provider/datasources/, internal/provider/validators/, internal/provider/planmodifiers/, internal/provider/fwhelpers/
  - Create directory skeleton for Framework code
  - Purpose: Organize migrated code separately from SDK v2
  - _Leverage: design.md architecture diagram_
  - _Requirements: 1, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Create directory structure as specified in design.md | Restrictions: Do not move existing files yet | Success: All directories exist with .gitkeep or placeholder files | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion, then mark as [x]_

- [ ] 1.3. Implement Framework provider skeleton
  - File: internal/provider/provider_framework.go
  - Create new provider implementing provider.Provider interface
  - Include schema with WriteOnly attributes for sensitive fields
  - Purpose: Foundation for Framework migration
  - _Leverage: internal/provider/provider.go (SDK v2 version), internal/client/client.go_
  - _Requirements: 1, 2, 5_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer with Framework expertise | Task: Create provider_framework.go implementing provider.Provider with WriteOnly attributes for password, private_key, admin_password following requirements 1, 2, 5 | Restrictions: Must require Terraform 1.11+, use typed schema, do not break SDK v2 provider yet | Success: Provider compiles, schema has WriteOnly sensitive fields, Configure creates client | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts (provider struct, schema), then mark as [x]_

- [ ] 1.4. Setup acceptance test infrastructure for Framework
  - File: internal/provider/acctest/framework_acctest.go
  - Create ProtoV6ProviderFactories for Framework testing
  - Setup test configuration helpers
  - Purpose: Enable testing of migrated resources
  - _Leverage: internal/provider/acctest/acctest.go (SDK v2 version)_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer with Terraform provider testing expertise | Task: Create Framework acceptance test infrastructure with ProtoV6ProviderFactories following requirement 1 | Restrictions: Keep SDK v2 test infrastructure for unmigrated resources | Success: Framework tests can run alongside SDK v2 tests | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion, then mark as [x]_

## Phase 2: High-Priority Resources (Sensitive Attributes)

- [ ] 2.1. Migrate rtx_ipsec_tunnel resource
  - Files: internal/provider/resources/ipsec_tunnel/resource.go, model.go, schema.go
  - Implement Resource interface with Create, Read, Update, Delete, ImportState
  - pre_shared_key as WriteOnly attribute
  - Purpose: Enable secure IPsec configuration without state exposure
  - _Leverage: internal/provider/resource_rtx_ipsec_tunnel.go, internal/client/ipsec_tunnel_service.go_
  - _Requirements: 2, 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_ipsec_tunnel to Framework with WriteOnly pre_shared_key following requirements 2, 3, 4. Create typed Model struct, Schema with nested blocks for ikev2_proposal and ipsec_transform | Restrictions: Must maintain exact same Terraform schema for users, pre_shared_key must be WriteOnly | Success: Resource compiles, acceptance tests pass, pre_shared_key not in state | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts (resource, model, schema), then mark as [x]_

- [ ] 2.2. Migrate rtx_l2tp resource
  - Files: internal/provider/resources/l2tp/resource.go, model.go, schema.go
  - tunnel_auth_password and ipsec_profile.pre_shared_key as WriteOnly
  - Complex nested blocks: l2tpv3_config, ipsec_profile, authentication
  - Purpose: Enable secure L2TP configuration
  - _Leverage: internal/provider/resource_rtx_l2tp.go, internal/client/l2tp_service.go_
  - _Requirements: 2, 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_l2tp to Framework with WriteOnly sensitive attributes following requirements 2, 3, 4. Handle complex nested blocks l2tpv3_config, ipsec_profile | Restrictions: Both tunnel_auth_password and ipsec_profile.pre_shared_key must be WriteOnly | Success: Resource compiles, acceptance tests pass, sensitive values not in state | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 2.3. Migrate rtx_admin resource
  - Files: internal/provider/resources/admin/resource.go, model.go, schema.go
  - admin_password and login_password as WriteOnly
  - Purpose: Secure admin credential management
  - _Leverage: internal/provider/resource_rtx_admin.go, internal/client/admin_service.go_
  - _Requirements: 2, 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_admin to Framework with WriteOnly admin_password and login_password following requirements 2, 3, 4 | Restrictions: All password fields must be WriteOnly | Success: Resource compiles, acceptance tests pass, passwords not in state | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 2.4. Migrate rtx_admin_user resource
  - Files: internal/provider/resources/admin_user/resource.go, model.go, schema.go
  - password as WriteOnly
  - Purpose: Secure user credential management
  - _Leverage: internal/provider/resource_rtx_admin_user.go, internal/client/admin_user_service.go_
  - _Requirements: 2, 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_admin_user to Framework with WriteOnly password following requirements 2, 3, 4 | Restrictions: password must be WriteOnly | Success: Resource compiles, acceptance tests pass, password not in state | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 2.5. Migrate rtx_ddns resource
  - Files: internal/provider/resources/ddns/resource.go, model.go, schema.go
  - password as WriteOnly
  - Purpose: Secure DDNS credential management
  - _Leverage: internal/provider/resource_rtx_ddns.go, internal/client/ddns_service.go_
  - _Requirements: 2, 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_ddns to Framework with WriteOnly password following requirements 2, 3, 4 | Restrictions: password must be WriteOnly | Success: Resource compiles, acceptance tests pass, password not in state | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

## Phase 3: Normal-Priority Resources

- [ ] 3.1. Create shared validators package
  - File: internal/provider/validators/validators.go
  - IP address, CIDR, MAC address, port range validators
  - Purpose: Reusable validation across resources
  - _Leverage: existing validation in SDK v2 resources_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Create Framework validators for IP, CIDR, MAC, port validation following requirements 3, 4 | Restrictions: Use terraform-plugin-framework-validators where possible | Success: All validators compile and have unit tests | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts (validators), then mark as [x]_

- [ ] 3.2. Create shared plan modifiers package
  - File: internal/provider/planmodifiers/planmodifiers.go
  - UseStateForUnknown, RequiresReplace modifiers
  - Purpose: Consistent plan behavior across resources
  - _Leverage: existing DiffSuppressFunc patterns in SDK v2_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Create Framework plan modifiers following requirements 3, 4 | Restrictions: Match existing SDK v2 diff suppress behavior | Success: Plan modifiers compile and work correctly | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 3.3. Migrate rtx_interface resource
  - Files: internal/provider/resources/interface/resource.go, model.go, schema.go
  - Purpose: Basic interface configuration
  - _Leverage: internal/provider/resource_rtx_interface.go_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_interface to Framework following requirements 3, 4 | Restrictions: Maintain schema compatibility | Success: Resource compiles, acceptance tests pass | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 3.4. Migrate rtx_ipv6_interface resource
  - Files: internal/provider/resources/ipv6_interface/resource.go, model.go, schema.go
  - _Leverage: internal/provider/resource_rtx_ipv6_interface.go_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_ipv6_interface to Framework following requirements 3, 4 | Restrictions: Maintain schema compatibility | Success: Resource compiles, acceptance tests pass | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 3.5. Migrate rtx_bridge resource
  - Files: internal/provider/resources/bridge/resource.go, model.go, schema.go
  - _Leverage: internal/provider/resource_rtx_bridge.go_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_bridge to Framework following requirements 3, 4 | Restrictions: Maintain schema compatibility | Success: Resource compiles, acceptance tests pass | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 3.6. Migrate rtx_dhcp_scope resource
  - Files: internal/provider/resources/dhcp_scope/resource.go, model.go, schema.go
  - _Leverage: internal/provider/resource_rtx_dhcp_scope.go_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_dhcp_scope to Framework following requirements 3, 4. Handle options nested block | Restrictions: Maintain schema compatibility | Success: Resource compiles, acceptance tests pass | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 3.7. Migrate rtx_dhcp_binding resource
  - Files: internal/provider/resources/dhcp_binding/resource.go, model.go, schema.go
  - _Leverage: internal/provider/resource_rtx_dhcp_binding.go_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_dhcp_binding to Framework following requirements 3, 4 | Restrictions: Maintain schema compatibility | Success: Resource compiles, acceptance tests pass | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 3.8. Migrate rtx_static_route resource
  - Files: internal/provider/resources/static_route/resource.go, model.go, schema.go
  - _Leverage: internal/provider/resource_rtx_static_route.go_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_static_route to Framework following requirements 3, 4 | Restrictions: Maintain schema compatibility | Success: Resource compiles, acceptance tests pass | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 3.9. Migrate rtx_nat_masquerade resource
  - Files: internal/provider/resources/nat_masquerade/resource.go, model.go, schema.go
  - _Leverage: internal/provider/resource_rtx_nat_masquerade.go_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_nat_masquerade to Framework following requirements 3, 4. Handle static_entry nested blocks | Restrictions: Maintain schema compatibility | Success: Resource compiles, acceptance tests pass | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 3.10. Migrate rtx_access_list_ip resource
  - Files: internal/provider/resources/access_list_ip/resource.go, model.go, schema.go
  - _Leverage: internal/provider/resource_rtx_access_list_ip.go_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_access_list_ip to Framework following requirements 3, 4. Handle entry and apply nested blocks | Restrictions: Maintain schema compatibility | Success: Resource compiles, acceptance tests pass | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 3.11. Migrate rtx_access_list_ipv6 resource
  - Files: internal/provider/resources/access_list_ipv6/resource.go, model.go, schema.go
  - _Leverage: internal/provider/resource_rtx_access_list_ipv6.go_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_access_list_ipv6 to Framework following requirements 3, 4 | Restrictions: Maintain schema compatibility | Success: Resource compiles, acceptance tests pass | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 3.12. Migrate rtx_access_list_mac resource
  - Files: internal/provider/resources/access_list_mac/resource.go, model.go, schema.go
  - _Leverage: internal/provider/resource_rtx_access_list_mac.go_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_access_list_mac to Framework following requirements 3, 4 | Restrictions: Maintain schema compatibility | Success: Resource compiles, acceptance tests pass | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 3.13. Migrate remaining access_list resources
  - Files: internal/provider/resources/access_list_*/
  - Includes: access_list_ip_dynamic, access_list_ipv6_dynamic, access_list_*_apply
  - _Leverage: existing SDK v2 implementations_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate remaining access_list resources (ip_dynamic, ipv6_dynamic, *_apply) to Framework following requirements 3, 4 | Restrictions: Maintain schema compatibility | Success: All access_list resources compile, acceptance tests pass | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 3.14. Migrate rtx_sshd and related resources
  - Files: internal/provider/resources/sshd/, sshd_host_key/, sshd_authorized_keys/
  - _Leverage: existing SDK v2 implementations_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_sshd, rtx_sshd_host_key, rtx_sshd_authorized_keys to Framework following requirements 3, 4 | Restrictions: Maintain schema compatibility | Success: All sshd resources compile, acceptance tests pass | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 3.15. Migrate rtx_system resource
  - Files: internal/provider/resources/system/resource.go, model.go, schema.go
  - _Leverage: internal/provider/resource_rtx_system.go_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_system to Framework following requirements 3, 4 | Restrictions: Maintain schema compatibility | Success: Resource compiles, acceptance tests pass | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 3.16. Migrate remaining singleton resources
  - Files: internal/provider/resources/syslog/, httpd/, sftpd/, dns_server/, l2tp_service/
  - _Leverage: existing SDK v2 implementations_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_syslog, rtx_httpd, rtx_sftpd, rtx_dns_server, rtx_l2tp_service to Framework following requirements 3, 4 | Restrictions: Maintain schema compatibility | Success: All singleton resources compile, acceptance tests pass | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 3.17. Migrate rtx_ipsec_transport resource
  - Files: internal/provider/resources/ipsec_transport/resource.go, model.go, schema.go
  - _Leverage: internal/provider/resource_rtx_ipsec_transport.go_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_ipsec_transport to Framework following requirements 3, 4 | Restrictions: Maintain schema compatibility | Success: Resource compiles, acceptance tests pass | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 3.18. Migrate rtx_ipv6_prefix resource
  - Files: internal/provider/resources/ipv6_prefix/resource.go, model.go, schema.go
  - _Leverage: internal/provider/resource_rtx_ipv6_prefix.go_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate rtx_ipv6_prefix to Framework following requirements 3, 4 | Restrictions: Maintain schema compatibility | Success: Resource compiles, acceptance tests pass | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 3.19. Migrate data sources
  - Files: internal/provider/datasources/
  - Includes: rtx_interfaces, rtx_ddns_status
  - _Leverage: existing SDK v2 data source implementations_
  - _Requirements: 3, 4_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Migrate all data sources to Framework following requirements 3, 4 | Restrictions: Maintain schema compatibility | Success: All data sources compile, tests pass | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

## Phase 4: Cleanup and Release

- [ ] 4.1. Update provider.go to use only Framework
  - File: internal/provider/provider.go
  - Replace SDK v2 provider with Framework provider
  - Remove all SDK v2 resource registrations
  - Purpose: Complete the migration
  - _Leverage: internal/provider/provider_framework.go_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Update main provider.go to use Framework provider, remove SDK v2 code following requirement 1 | Restrictions: Ensure all resources are registered | Success: Provider compiles with only Framework | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 4.2. Remove SDK v2 dependencies
  - File: go.mod
  - Remove terraform-plugin-sdk/v2
  - Run go mod tidy
  - Purpose: Clean dependencies
  - _Requirements: 1_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Remove terraform-plugin-sdk/v2 from go.mod following requirement 1 | Restrictions: Ensure no SDK v2 imports remain | Success: go mod tidy succeeds, no SDK v2 in go.sum | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 4.3. Delete old SDK v2 resource files
  - Files: internal/provider/resource_rtx_*.go (old), schema_helpers.go, state_funcs.go
  - Purpose: Remove deprecated code
  - _Requirements: 1_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Developer | Task: Delete all old SDK v2 resource files following requirement 1 | Restrictions: Verify no imports before deletion | Success: Only Framework code remains in provider package | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts (files deleted), then mark as [x]_

- [ ] 4.4. Update documentation
  - Files: docs/index.md, README.md, docs/resources/*.md
  - Add Terraform 1.11+ requirement
  - Update examples if needed
  - Purpose: Inform users of requirements
  - _Requirements: 5_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer | Task: Update all documentation to reflect Terraform 1.11+ requirement following requirement 5 | Restrictions: Keep existing content accurate | Success: Docs clearly state version requirement | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with artifacts, then mark as [x]_

- [ ] 4.5. Run full acceptance test suite
  - Verify all tests pass with Framework implementation
  - Test write-only attributes don't appear in state
  - Purpose: Final validation
  - _Requirements: All_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full acceptance test suite and verify all requirements are met | Restrictions: All tests must pass | Success: TF_ACC=1 go test ./... passes, sensitive values not in state | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion with test results, then mark as [x]_

- [ ] 4.6. Update Master Spec documentation
  - Files: docs/index.md, CLAUDE.md
  - Document Plugin Framework architecture
  - Update development commands if changed
  - Purpose: Keep project documentation current
  - _Requirements: All_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Technical Writer | Task: Update Master Spec documentation (docs/index.md, CLAUDE.md) to reflect Plugin Framework architecture | Restrictions: Keep existing useful content | Success: Documentation accurately describes new architecture | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion, then mark as [x]_

- [ ] 4.7. Release as v0.7.0
  - Update version in Makefile, README, docs
  - Create release notes documenting Terraform 1.11+ requirement
  - Purpose: Publish migration
  - _Requirements: All_
  - _Prompt: Implement the task for spec plugin-framework-migration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Release Engineer | Task: Prepare v0.7.0 release with release notes documenting Terraform 1.11+ requirement | Restrictions: Follow existing release process in CLAUDE.md, minor version increment only | Success: Version bumped to 0.7.0, release notes ready | Instructions: Mark task as [-] in tasks.md before starting, use log-implementation tool after completion, then mark as [x]_
