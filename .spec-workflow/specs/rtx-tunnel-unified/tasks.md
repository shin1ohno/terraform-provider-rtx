# Implementation Tasks: rtx_tunnel (Unified Tunnel Resource)

## Task Overview

| Task | Description | Estimated Complexity | Status |
|------|-------------|---------------------|--------|
| 1 | Parser: Unified tunnel parser | Medium | ✅ Complete |
| 2 | Client: Tunnel data types | Low | ✅ Complete |
| 3 | Client: TunnelService | High | ✅ Complete |
| 4 | Provider: Tunnel resource model | Medium | ✅ Complete |
| 5 | Provider: Tunnel resource schema | Medium | ✅ Complete |
| 6 | Provider: CRUD implementation | High | ✅ Complete |
| 7 | Tests: Parser unit tests | Medium | ✅ Complete |
| 8 | Tests: Service unit tests | Medium | ✅ Complete |
| 9 | Deprecation: Old resources | Low | ⏳ Pending |
| 10 | Documentation: Update docs | Medium | ✅ Complete |
| 11 | Examples: Update main.tf | Low | ✅ Complete |

---

## Task 1: Parser - Unified Tunnel Parser

**File:** `internal/rtx/parsers/tunnel.go`

### Subtasks

1.1. Create `Tunnel` struct that embeds IPsec and L2TP components
1.2. Create `ParseTunnelConfig()` that:
   - Parses `tunnel select N` blocks
   - Determines encapsulation type
   - Delegates to existing IPsec/L2TP parsers for nested content
1.3. Create unified command builders:
   - `BuildTunnelCommands(tunnel Tunnel) []string` - orchestrates all commands
   - Reuse existing builders from `ipsec_tunnel.go` and `l2tp.go`

### Acceptance Criteria

- [x] Can parse all three encapsulation types from router config
- [x] Command builders generate correct RTX command sequences
- [x] Reuses existing parser code where possible

---

## Task 2: Client - Tunnel Data Types

**File:** `internal/client/interfaces.go`

### Subtasks

2.1. Add `Tunnel` struct with embedded `TunnelIPsec` and `TunnelL2TP`
2.2. Add helper types: `IPsecTransform`, `IPsecKeepalive`, `L2TPKeepalive`, etc.
2.3. Add `TunnelService` interface to `Client`

### Acceptance Criteria

- [x] All data types defined with proper JSON tags
- [x] Sensitive fields marked appropriately
- [x] Types align with Terraform schema design

---

## Task 3: Client - TunnelService

**File:** `internal/client/tunnel_service.go`

### Subtasks

3.1. Implement `TunnelService` struct
3.2. Implement `Create()`:
   - Generate commands in correct order
   - Handle all three encapsulation types
3.3. Implement `Get()`:
   - Parse config and return unified Tunnel
3.4. Implement `Update()`:
   - Handle partial updates
   - Clear old settings when needed
3.5. Implement `Delete()`:
   - Clean up all related config
3.6. Implement `List()`:
   - Return all configured tunnels

### Acceptance Criteria

- [x] All CRUD operations work for all encapsulation types
- [x] Commands executed in correct order within tunnel context
- [x] Config saved after modifications

---

## Task 4: Provider - Tunnel Resource Model

**File:** `internal/provider/resources/tunnel/model.go`

### Subtasks

4.1. Create `TunnelModel` with tfsdk tags
4.2. Create nested models: `TunnelIPsecModel`, `TunnelL2TPModel`, etc.
4.3. Implement `ToClient()` conversion
4.4. Implement `FromClient()` conversion
4.5. Implement `ID()` method

### Acceptance Criteria

- [x] All fields have proper tfsdk tags
- [x] Sensitive fields handled correctly
- [x] Bi-directional conversion works

---

## Task 5: Provider - Tunnel Resource Schema

**File:** `internal/provider/resources/tunnel/resource.go`

### Subtasks

5.1. Define schema with all attributes and nested blocks
5.2. Add validators for encapsulation-specific requirements
5.3. Add plan modifiers (RequiresReplace for tunnel_id)
5.4. Add cross-block validation (ipsec required for l2tp encapsulation, etc.)

### Acceptance Criteria

- [x] Schema matches design document
- [x] Validation prevents invalid configurations
- [x] Plan modifiers correctly identify replacement scenarios

---

## Task 6: Provider - CRUD Implementation

**File:** `internal/provider/resources/tunnel/resource.go`

### Subtasks

6.1. Implement `Create()`:
   - Validate encapsulation-specific requirements
   - Call TunnelService.Create
   - Read back and update state
6.2. Implement `Read()`:
   - Call TunnelService.Get
   - Update model from client
6.3. Implement `Update()`:
   - Detect changes
   - Call TunnelService.Update
   - Read back and update state
6.4. Implement `Delete()`:
   - Call TunnelService.Delete
6.5. Implement `ImportState()`:
   - Parse tunnel ID from import string
   - Call Read to populate state

### Acceptance Criteria

- [x] Full CRUD lifecycle works
- [x] Import works
- [x] Error handling is comprehensive

---

## Task 7: Tests - Parser Unit Tests

**File:** `internal/rtx/parsers/tunnel_test.go`

### Subtasks

7.1. Test `ParseTunnelConfig()` for IPsec-only tunnels
7.2. Test `ParseTunnelConfig()` for L2TPv3 tunnels
7.3. Test `ParseTunnelConfig()` for L2TPv2 tunnels
7.4. Test command builders for all encapsulation types
7.5. Test edge cases (missing fields, multiple tunnels)

### Acceptance Criteria

- [x] All encapsulation types tested
- [x] Real router config samples used
- [x] Edge cases covered

---

## Task 8: Tests - Service Unit Tests

**File:** `internal/client/tunnel_service_test.go`

### Subtasks

8.1. Test Create for all encapsulation types
8.2. Test Get with mock executor
8.3. Test Update scenarios
8.4. Test Delete
8.5. Test error handling

### Acceptance Criteria

- [x] All CRUD operations tested
- [x] Mock executor used
- [x] Error scenarios covered

---

## Task 9: Deprecation - Old Resources

**Files:**
- `internal/provider/resources/ipsec_tunnel/resource.go`
- `internal/provider/resources/l2tp/resource.go`

### Subtasks

9.1. Add deprecation warning to `rtx_ipsec_tunnel` schema
9.2. Add deprecation warning to `rtx_l2tp` schema
9.3. Log deprecation notice on resource use

### Acceptance Criteria

- [ ] Deprecation warnings visible in terraform plan
- [ ] Documentation updated with migration guide

> **Status:** Not started - deprecation warnings not yet implemented in code

---

## Task 10: Documentation - Update Docs

**Files:** `docs/resources/tunnel.md`, `docs/index.md`

### Subtasks

10.1. Create `docs/resources/tunnel.md` with full documentation
10.2. Update `docs/index.md` to reference new resource
10.3. Add migration guide for old resources
10.4. Run `tfplugindocs` to regenerate

### Acceptance Criteria

- [x] Documentation complete and accurate
- [ ] Migration guide clear and actionable
- [x] Examples included

> **Status:** Partially complete - docs/resources/tunnel.md generated by tfplugindocs

---

## Task 11: Examples - Update main.tf

**File:** `examples/import/main.tf`

### Subtasks

11.1. Replace `rtx_ipsec_tunnel` and `rtx_l2tp` with `rtx_tunnel`
11.2. Update any references
11.3. Test with `terraform plan` to ensure no changes

### Acceptance Criteria

- [x] main.tf uses only rtx_tunnel
- [x] terraform plan shows no changes (after state migration)
- [x] All tunnel configurations preserved

> **Status:** Complete - examples/tunnel/main.tf updated with new resource

---

## Execution Order

```
1. Task 2 (Data types) - Foundation
2. Task 1 (Parser) - Core parsing logic
3. Task 7 (Parser tests) - Verify parsing
4. Task 3 (Service) - Business logic
5. Task 8 (Service tests) - Verify service
6. Task 4 (Model) - Terraform model
7. Task 5 (Schema) - Terraform schema
8. Task 6 (CRUD) - Resource implementation
9. Task 11 (Examples) - Update examples
10. Task 9 (Deprecation) - Mark old resources
11. Task 10 (Documentation) - Final docs
```

## Notes

- Reuse existing code from `ipsec_tunnel.go` and `l2tp.go` parsers
- Consider extracting common validation functions
- Test with real router config from `examples/import/`
