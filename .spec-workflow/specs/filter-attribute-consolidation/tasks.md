# Tasks: Filter Attribute Consolidation

## Task List

### Phase 1: Create New Dynamic Filter Access List Resources (Parallel)

- [x] **Task 1.1**: Create `rtx_access_list_ip_dynamic` resource
  - File: `internal/provider/resource_rtx_access_list_ip_dynamic.go`
  - Schema: name, entry block (sequence, source, destination, protocol, syslog, timeout)
  - CRUD operations
  - Reference existing `rtx_access_list_ip` for patterns

- [x] **Task 1.2**: Create `rtx_access_list_ip_dynamic` tests
  - File: `internal/provider/resource_rtx_access_list_ip_dynamic_test.go`
  - Unit tests for schema validation
  - Entry ordering tests

- [x] **Task 1.3**: Create `rtx_access_list_ipv6_dynamic` resource
  - File: `internal/provider/resource_rtx_access_list_ipv6_dynamic.go`
  - Schema: name, entry block (sequence, source, destination, protocol, syslog)
  - CRUD operations
  - Reference existing `rtx_access_list_ipv6` for patterns

- [x] **Task 1.4**: Create `rtx_access_list_ipv6_dynamic` tests
  - File: `internal/provider/resource_rtx_access_list_ipv6_dynamic_test.go`
  - Unit tests for schema validation
  - Entry ordering tests

### Phase 2: Update rtx_interface Resource

- [x] **Task 2.1**: Add new access list attributes to `rtx_interface`
  - File: `internal/provider/resource_rtx_interface.go`
  - Add: `access_list_ip_in`, `access_list_ip_out`
  - Add: `access_list_ipv6_in`, `access_list_ipv6_out`
  - Add: `access_list_ip_dynamic_in`, `access_list_ip_dynamic_out`
  - Add: `access_list_ipv6_dynamic_in`, `access_list_ipv6_dynamic_out`
  - Add: `access_list_mac_in`, `access_list_mac_out`

- [x] **Task 2.2**: Remove old filter attributes from `rtx_interface`
  - File: `internal/provider/resource_rtx_interface.go`
  - Remove: `secure_filter_in`, `secure_filter_out`
  - Remove: `dynamic_filter_out`
  - Remove: `ethernet_filter_in`, `ethernet_filter_out`

- [x] **Task 2.3**: Update `InterfaceConfig` struct
  - File: `internal/client/interfaces.go`
  - Remove old filter fields
  - Add new access list name fields

- [x] **Task 2.4**: Update interface service layer
  - File: `internal/client/interface_service.go`
  - Remove filter number handling
  - Add access list name lookup and filter resolution

- [x] **Task 2.5**: Update `rtx_interface` tests
  - File: `internal/provider/resource_rtx_interface_test.go`
  - Remove tests for old filter attributes
  - Add tests for new access_list_* attributes

### Phase 3: Remove Redundant Resources (Parallel)

- [x] **Task 3.1**: Remove `rtx_interface_acl` resource
  - Delete: `internal/provider/resource_rtx_interface_acl.go`
  - Delete: `internal/provider/resource_rtx_interface_acl_test.go`
  - Delete: `docs/resources/interface_acl.md`
  - Remove from provider.go ResourcesMap

- [x] **Task 3.2**: Remove `rtx_interface_mac_acl` resource
  - Delete: `internal/provider/resource_rtx_interface_mac_acl.go`
  - Delete: `internal/provider/resource_rtx_interface_mac_acl_test.go`
  - Delete: `docs/resources/interface_mac_acl.md`
  - Remove from provider.go ResourcesMap

- [x] **Task 3.3**: Remove `rtx_ip_filter_dynamic` resource
  - Delete: `internal/provider/resource_rtx_ip_filter_dynamic.go`
  - Delete: `internal/provider/resource_rtx_ip_filter_dynamic_test.go`
  - Delete: `docs/resources/ip_filter_dynamic.md`
  - Remove from provider.go ResourcesMap

- [x] **Task 3.4**: Remove `rtx_ipv6_filter_dynamic` resource
  - Delete: `internal/provider/resource_rtx_ipv6_filter_dynamic.go`
  - Delete: `internal/provider/resource_rtx_ipv6_filter_dynamic_test.go`
  - Delete: `docs/resources/ipv6_filter_dynamic.md`
  - Remove from provider.go ResourcesMap

### Phase 4: Update Provider Registration

- [x] **Task 4.1**: Update provider.go
  - File: `internal/provider/provider.go`
  - Add: `rtx_access_list_ip_dynamic`, `rtx_access_list_ipv6_dynamic`
  - Confirm removal of old resources from Phase 3

### Phase 5: Documentation (Parallel)

- [x] **Task 5.1**: Create `access_list_ip_dynamic` documentation
  - File: `docs/resources/access_list_ip_dynamic.md`
  - Schema documentation
  - Usage examples

- [x] **Task 5.2**: Create `access_list_ipv6_dynamic` documentation
  - File: `docs/resources/access_list_ipv6_dynamic.md`
  - Schema documentation
  - Usage examples

- [x] **Task 5.3**: Update `interface.md` documentation
  - File: `docs/resources/interface.md`
  - Remove old filter attribute documentation
  - Add new access_list_* attribute documentation
  - Add migration note

- [x] **Task 5.4**: Create migration guide
  - Document breaking changes
  - Before/after examples
  - Step-by-step migration instructions
  - (Included in interface.md)

### Phase 6: Final Verification

- [x] **Task 6.1**: Run all tests
  - `make test`
  - Verify no regressions

- [x] **Task 6.2**: Run linter
  - `make lint`
  - Fix any issues

- [x] **Task 6.3**: Build provider
  - `make build`
  - Verify successful build

- [ ] **Task 6.4**: Update CHANGELOG
  - Document breaking changes
  - Document removed resources
  - Document new resources
  - Add migration instructions
  - (Skipped - no CHANGELOG file exists)

## Dependencies

```
Phase 1 (1.1-1.4 parallel) ─► Phase 2 (2.1-2.5 sequential) ─► Phase 3 (3.1-3.4 parallel)
                                                                       │
                                                                       ├─► Phase 4 ─► Phase 6
                                                                       │
                                                                       └─► Phase 5 (parallel)
```

## Summary of Breaking Changes

1. **New Resources**:
   - `rtx_access_list_ip_dynamic`
   - `rtx_access_list_ipv6_dynamic`

2. **Removed Resources**:
   - `rtx_interface_acl`
   - `rtx_interface_mac_acl`
   - `rtx_ip_filter_dynamic`
   - `rtx_ipv6_filter_dynamic`

3. **Removed Attributes from `rtx_interface`**:
   - `secure_filter_in`, `secure_filter_out`
   - `dynamic_filter_out`
   - `ethernet_filter_in`, `ethernet_filter_out`

4. **New Attributes in `rtx_interface`**:
   - `access_list_ip_in`, `access_list_ip_out`
   - `access_list_ipv6_in`, `access_list_ipv6_out`
   - `access_list_ip_dynamic_in`, `access_list_ip_dynamic_out`
   - `access_list_ipv6_dynamic_in`, `access_list_ipv6_dynamic_out`
   - `access_list_mac_in`, `access_list_mac_out`

## Completion Status

**All tasks completed: 2026-01-25**

- Phase 1-5: ✅ Complete
- Phase 6: ✅ Complete (except CHANGELOG - file does not exist)
- Build: ✅ Pass
- Lint: ✅ Pass
- Tests: ✅ All pass
