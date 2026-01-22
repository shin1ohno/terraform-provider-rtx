# Tasks Document: Comprehensive CRUD Tests

## Task Overview

| Phase | Tasks | Description | Files | Requirements |
|-------|-------|-------------|-------|--------------|
| 0 | 1.1-1.3 (3 tasks) | Infrastructure: Batch execution | executor.go, interfaces.go, session.go | REQ-3 |
| 1 | 2.x.1-2.x.3 (24 tasks) | Client layer TDD: Write test → Fix impl → Verify green (8 services × 3 steps) | *_service_test.go, *_service.go | REQ-1 |
| 2 | 3.x.1-3.x.3 (45 tasks) | Provider layer TDD: Write test → Fix impl → Verify green (15 resources × 3 steps) | resource_rtx_*_test.go, resource_rtx_*.go | REQ-2 |
| 3 | 4.1 (1 task) | Final verification: Run all tests, verify coverage | - | All |

**Total: 73 tasks**

## TDD Cycle

Each test task follows the TDD (Test-Driven Development) cycle:

```
┌─────────────────────────────────────────────────────────────┐
│ Step 1: Write Test (Red)                                    │
│ - Create test file with expected behavior                   │
│ - Test may fail initially (expected)                        │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ Step 2: Fix Implementation (Green)                          │
│ - Modify service/resource code to pass tests                │
│ - Add RunBatch support if needed                            │
│ - Add configure refresh / VPN safe update if needed         │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ Step 3: Verify Green                                        │
│ - Run: go test ./path/to/package -v -run TestName           │
│ - All tests must pass                                       │
│ - Mark task complete only after green                       │
└─────────────────────────────────────────────────────────────┘
```

---

## Phase 0: Infrastructure (Batch Execution)

- [x] 1.1. Add RunBatch to Executor interface
  - File: `internal/client/executor.go`, `internal/client/interfaces.go`
  - Add `RunBatch(ctx context.Context, cmds []string) ([]byte, error)` method
  - Implement in sshExecutor using Session.SendBatch
  - Purpose: Enable batch command execution for all resources
  - _Leverage: Existing Executor.Run implementation_
  - _Requirements: REQ-3_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in interface design | Task: Add RunBatch method to Executor interface and implement in sshExecutor. The method should accept []string commands and delegate to Session.SendBatch | Restrictions: Do not modify existing Run method behavior, maintain backward compatibility | Success: RunBatch method compiles, follows existing patterns | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 1.2. Implement SendBatch in Session
  - File: `internal/client/simple_session.go`, `internal/client/interfaces.go`
  - Add `SendBatch(cmds []string) ([]byte, error)` to Session interface
  - Implement: send all commands without waiting for individual prompts
  - Wait only for final prompt after all commands sent
  - Purpose: Enable fast batch command execution
  - _Leverage: Existing Send implementation, sendCommand function_
  - _Requirements: REQ-3_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with SSH expertise | Task: Add SendBatch method to Session interface and implement in simpleRTXSession. Send all commands at once using fmt.Fprintf, then read until final prompt | Restrictions: Must handle connection drops gracefully, do not wait for prompt between commands | Success: SendBatch sends commands in batch, reads final output correctly | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 1.3. Implement VPN Safe Update pattern helpers
  - File: `internal/client/tunnel_helper.go` (new)
  - Create `ExecuteTunnelUpdate(ctx, executor, tunnelID, commands []string) error`
  - Phase 1: SendBatch (commands + disable + enable, no save)
  - Phase 2: WaitForReconnection with polling
  - Phase 3: Reconnect and save
  - Purpose: Enable safe tunnel updates over VPN
  - _Leverage: Executor.RunBatch, existing SSH connection logic_
  - _Requirements: REQ-3, REQ-4_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with network programming expertise | Task: Create tunnel_helper.go with ExecuteTunnelUpdate function implementing VPN safe update pattern: (1) SendBatch commands without save, (2) Poll for SSH reconnection with timeout, (3) Reconnect and save | Restrictions: Must handle connection drops as expected not error, implement configurable timeout for polling | Success: Tunnel updates work safely over VPN, reconnection polling works | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

---

## Phase 1: Client Layer Tests (TDD Cycle)

### 2.1 BGPService

- [x] 2.1.1. Write BGPService tests (Red)
  - File: `internal/client/bgp_service_test.go`
  - Write tests for Get/Create/Update/Delete
  - Use MockExecutor.On("RunBatch", ctx, expectedCommands)
  - Verify commands include `bgp configure refresh`
  - Tests may fail initially (expected)
  - _Leverage: MockExecutor pattern, parsers/bgp.go_
  - _Requirements: REQ-1, REQ-4_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Write comprehensive tests for BGPService CRUD operations. Use RunBatch mock for batch commands. Verify bgp configure refresh is included | Restrictions: Follow existing test patterns, tests may fail initially | Success: Test file created with all CRUD test cases | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 2.1.2. Fix BGPService implementation
  - File: `internal/client/bgp_service.go`
  - Modify Create/Update/Delete to use RunBatch
  - Add `bgp configure refresh` to Create/Update
  - Purpose: Make tests pass
  - _Leverage: Executor.RunBatch, parsers.BuildBGPCommands_
  - _Requirements: REQ-1, REQ-4_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Modify BGPService to use RunBatch for batch execution. Add bgp configure refresh after Create/Update | Restrictions: Do not break existing functionality | Success: Service uses batch execution, includes configure refresh | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 2.1.3. Verify BGPService tests green
  - Run: `go test ./internal/client/... -v -run TestBGP`
  - All tests must pass
  - Fix any remaining issues
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run BGPService tests, verify all pass. Fix any failures | Restrictions: Do not skip tests | Success: All BGPService tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 2.2 DHCPScopeService

- [x] 2.2.1. Write DHCPScopeService tests (Red)
  - File: `internal/client/dhcp_scope_service_test.go`
  - Test both CIDR and IP range formats
  - Use MockExecutor.On("RunBatch", ctx, expectedCommands)
  - _Leverage: MockExecutor pattern, parsers/dhcp_scope.go_
  - _Requirements: REQ-1, REQ-4_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Write tests for DHCPScopeService. Include both CIDR and IP range format tests | Restrictions: Follow existing test patterns | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 2.2.2. Fix DHCPScopeService implementation
  - File: `internal/client/dhcp_scope_service.go`
  - Modify to use RunBatch
  - _Leverage: Executor.RunBatch_
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Modify DHCPScopeService to use RunBatch | Restrictions: Do not break existing functionality | Success: Service uses batch execution | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 2.2.3. Verify DHCPScopeService tests green
  - Run: `go test ./internal/client/... -v -run TestDHCP`
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run DHCPScopeService tests, verify all pass | Restrictions: Do not skip tests | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 2.3 IPsecTunnelService

- [x] 2.3.1. Write IPsecTunnelService tests (Red)
  - File: `internal/client/ipsec_tunnel_service_test.go`
  - Test VPN safe update pattern for Update
  - Mock reconnection behavior
  - _Leverage: MockExecutor pattern, tunnel_helper.go_
  - _Requirements: REQ-1, REQ-4_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Write tests for IPsecTunnelService including VPN safe update pattern | Restrictions: Mock reconnection | Success: Test file created with VPN pattern tests | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 2.3.2. Fix IPsecTunnelService implementation
  - File: `internal/client/ipsec_tunnel_service.go`
  - Use RunBatch, implement VPN safe update for Update
  - _Leverage: tunnel_helper.go_
  - _Requirements: REQ-1, REQ-4_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Modify IPsecTunnelService to use VPN safe update pattern | Restrictions: Handle reconnection properly | Success: VPN safe update implemented | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 2.3.3. Verify IPsecTunnelService tests green
  - Run: `go test ./internal/client/... -v -run TestIPsec`
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run IPsecTunnelService tests, verify all pass | Restrictions: Do not skip tests | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 2.4 NATMasqueradeService

- [x] 2.4.1. Write NATMasqueradeService tests (Red)
  - File: `internal/client/nat_masquerade_service_test.go`
  - Include protocol-only entry tests (ESP, AH, GRE)
  - _Leverage: MockExecutor pattern, parsers/nat_masquerade.go_
  - _Requirements: REQ-1, REQ-4_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Write tests for NATMasqueradeService including protocol-only entries | Restrictions: Cover edge cases | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 2.4.2. Fix NATMasqueradeService implementation
  - File: `internal/client/nat_masquerade_service.go`
  - Use RunBatch
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Modify NATMasqueradeService to use RunBatch | Restrictions: Do not break existing functionality | Success: Service uses batch execution | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 2.4.3. Verify NATMasqueradeService tests green
  - Run: `go test ./internal/client/... -v -run TestNATMasquerade`
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Restrictions: Do not skip tests | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 2.5 NATStaticService

- [x] 2.5.1. Write NATStaticService tests (Red)
  - File: `internal/client/nat_static_service_test.go`
  - _Leverage: MockExecutor pattern_
  - _Requirements: REQ-1, REQ-4_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Write tests for NATStaticService | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 2.5.2. Fix NATStaticService implementation
  - File: `internal/client/nat_static_service.go`
  - Use RunBatch
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Modify NATStaticService to use RunBatch | Success: Service uses batch | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 2.5.3. Verify NATStaticService tests green
  - Run: `go test ./internal/client/... -v -run TestNATStatic`
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 2.6 OSPFService

- [x] 2.6.1. Write OSPFService tests (Red)
  - File: `internal/client/ospf_service_test.go`
  - Verify `ospf configure refresh` included
  - _Leverage: MockExecutor pattern, parsers/ospf.go_
  - _Requirements: REQ-1, REQ-4_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Write tests for OSPFService. Verify ospf configure refresh | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 2.6.2. Fix OSPFService implementation
  - File: `internal/client/ospf_service.go`
  - Use RunBatch, add `ospf configure refresh`
  - _Requirements: REQ-1, REQ-4_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Modify OSPFService to use RunBatch, add ospf configure refresh | Success: Service uses batch with refresh | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 2.6.3. Verify OSPFService tests green
  - Run: `go test ./internal/client/... -v -run TestOSPF`
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 2.7 PPTPService

- [x] 2.7.1. Write PPTPService tests (Red)
  - File: `internal/client/pptp_service_test.go`
  - Test VPN safe update pattern for Update
  - _Leverage: MockExecutor pattern, tunnel_helper.go_
  - _Requirements: REQ-1, REQ-4_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Write tests for PPTPService including VPN safe update | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 2.7.2. Fix PPTPService implementation
  - File: `internal/client/pptp_service.go`
  - Use RunBatch, VPN safe update for Update
  - _Requirements: REQ-1, REQ-4_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Modify PPTPService to use VPN safe update | Success: VPN safe update implemented | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 2.7.3. Verify PPTPService tests green
  - Run: `go test ./internal/client/... -v -run TestPPTP`
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 2.8 SystemService

- [x] 2.8.1. Write SystemService tests (Red)
  - File: `internal/client/system_service_test.go`
  - Test timezone, console, packet-buffer
  - _Leverage: MockExecutor pattern, parsers/system.go_
  - _Requirements: REQ-1, REQ-4_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Test Developer | Task: Write tests for SystemService | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 2.8.2. Fix SystemService implementation
  - File: `internal/client/system_service.go`
  - Use RunBatch
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer | Task: Modify SystemService to use RunBatch | Success: Service uses batch | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 2.8.3. Verify SystemService tests green
  - Run: `go test ./internal/client/... -v -run TestSystem`
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

---

## Phase 2: Provider Layer Tests (TDD Cycle)

### 3.1 rtx_access_list_extended

- [x] 3.1.1. Write rtx_access_list_extended tests (Red)
  - File: `internal/provider/resource_rtx_access_list_extended_test.go`
  - Test schema to config conversion
  - Test config to state population
  - Test validation rules
  - Tests may fail initially (expected)
  - _Leverage: schema.TestResourceDataRaw, existing provider test patterns_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create tests for rtx_access_list_extended using schema.TestResourceDataRaw. Test buildAccessListFromResourceData and state population | Restrictions: Follow existing provider test patterns, tests may fail initially | Success: Test file created with CRUD test cases | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.1.2. Fix rtx_access_list_extended implementation
  - File: `internal/provider/resource_rtx_access_list_extended.go`
  - Fix any issues found by tests
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Fix any issues in rtx_access_list_extended to make tests pass | Restrictions: Do not break existing functionality | Success: Implementation matches test expectations | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.1.3. Verify rtx_access_list_extended tests green
  - Run: `go test ./internal/provider/... -v -run TestAccListExt`
  - All tests must pass
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 3.2 rtx_access_list_extended_ipv6

- [x] 3.2.1. Write rtx_access_list_extended_ipv6 tests (Red)
  - File: `internal/provider/resource_rtx_access_list_extended_ipv6_test.go`
  - Test schema to config conversion for IPv6
  - Test IPv6 prefix validation
  - _Leverage: schema.TestResourceDataRaw, existing provider test patterns_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create tests for rtx_access_list_extended_ipv6. Test IPv6-specific validation | Restrictions: Follow existing provider test patterns | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.2.2. Fix rtx_access_list_extended_ipv6 implementation
  - File: `internal/provider/resource_rtx_access_list_extended_ipv6.go`
  - Fix any issues found by tests
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Fix any issues to make tests pass | Success: Implementation matches test expectations | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.2.3. Verify rtx_access_list_extended_ipv6 tests green
  - Run: `go test ./internal/provider/... -v -run TestAccListExtIPv6`
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 3.3 rtx_access_list_mac

- [x] 3.3.1. Write rtx_access_list_mac tests (Red)
  - File: `internal/provider/resource_rtx_access_list_mac_test.go`
  - Test schema to config conversion for MAC
  - Test MAC address format validation
  - _Leverage: schema.TestResourceDataRaw, existing provider test patterns_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create tests for rtx_access_list_mac. Test MAC address format validation | Restrictions: Follow existing provider test patterns | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.3.2. Fix rtx_access_list_mac implementation
  - File: `internal/provider/resource_rtx_access_list_mac.go`
  - Fix any issues found by tests
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Fix any issues to make tests pass | Success: Implementation matches test expectations | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.3.3. Verify rtx_access_list_mac tests green
  - Run: `go test ./internal/provider/... -v -run TestAccListMAC`
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 3.4 rtx_bgp

- [x] 3.4.1. Write rtx_bgp Provider tests (Red)
  - File: `internal/provider/resource_rtx_bgp_test.go`
  - Test schema to config conversion
  - Test neighbor configuration handling
  - Test AS number validation
  - _Leverage: schema.TestResourceDataRaw, existing provider test patterns_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create tests for rtx_bgp. Test neighbor block handling, AS validation | Restrictions: Follow existing provider test patterns | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.4.2. Fix rtx_bgp implementation
  - File: `internal/provider/resource_rtx_bgp.go`
  - Fix any issues found by tests
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Fix any issues to make tests pass | Success: Implementation matches test expectations | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.4.3. Verify rtx_bgp tests green
  - Run: `go test ./internal/provider/... -v -run TestBGP`
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 3.5 rtx_dhcp_scope

- [x] 3.5.1. Write rtx_dhcp_scope Provider tests (Red)
  - File: `internal/provider/resource_rtx_dhcp_scope_test.go`
  - Test schema to config conversion
  - Test IP range fields (range_start, range_end)
  - Test options handling
  - _Leverage: schema.TestResourceDataRaw, existing provider test patterns_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create tests for rtx_dhcp_scope. Test both CIDR and IP range formats, options | Restrictions: Follow existing provider test patterns | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.5.2. Fix rtx_dhcp_scope implementation
  - File: `internal/provider/resource_rtx_dhcp_scope.go`
  - Fix any issues found by tests
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Fix any issues to make tests pass | Success: Implementation matches test expectations | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.5.3. Verify rtx_dhcp_scope tests green
  - Run: `go test ./internal/provider/... -v -run TestDHCPScope`
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 3.6 rtx_interface_acl

- [x] 3.6.1. Write rtx_interface_acl Provider tests (Red)
  - File: `internal/provider/resource_rtx_interface_acl_test.go`
  - Test schema to config conversion
  - Test interface and direction validation
  - _Leverage: schema.TestResourceDataRaw, existing provider test patterns_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create tests for rtx_interface_acl. Test interface types and direction (in/out) | Restrictions: Follow existing provider test patterns | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.6.2. Fix rtx_interface_acl implementation
  - File: `internal/provider/resource_rtx_interface_acl.go`
  - Fix any issues found by tests
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Fix any issues to make tests pass | Success: Implementation matches test expectations | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.6.3. Verify rtx_interface_acl tests green
  - Run: `go test ./internal/provider/... -v -run TestInterfaceACL`
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 3.7 rtx_interface_mac_acl

- [x] 3.7.1. Write rtx_interface_mac_acl Provider tests (Red)
  - File: `internal/provider/resource_rtx_interface_mac_acl_test.go`
  - Test schema to config conversion
  - Test MAC ACL binding
  - _Leverage: schema.TestResourceDataRaw, existing provider test patterns_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create tests for rtx_interface_mac_acl | Restrictions: Follow existing provider test patterns | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.7.2. Fix rtx_interface_mac_acl implementation
  - File: `internal/provider/resource_rtx_interface_mac_acl.go`
  - Fix any issues found by tests
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Fix any issues to make tests pass | Success: Implementation matches test expectations | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.7.3. Verify rtx_interface_mac_acl tests green
  - Run: `go test ./internal/provider/... -v -run TestInterfaceMACACL`
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 3.8 rtx_ipsec_tunnel

- [x] 3.8.1. Write rtx_ipsec_tunnel Provider tests (Red)
  - File: `internal/provider/resource_rtx_ipsec_tunnel_test.go`
  - Test schema to config conversion
  - Test IKE settings handling
  - Test SA policy configuration
  - _Leverage: schema.TestResourceDataRaw, existing provider test patterns_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create tests for rtx_ipsec_tunnel. Test IKE and SA policy nested blocks | Restrictions: Follow existing provider test patterns | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.8.2. Fix rtx_ipsec_tunnel implementation
  - File: `internal/provider/resource_rtx_ipsec_tunnel.go`
  - Fix any issues found by tests
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Fix any issues to make tests pass | Success: Implementation matches test expectations | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.8.3. Verify rtx_ipsec_tunnel tests green
  - Run: `go test ./internal/provider/... -v -run TestIPsecTunnel`
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 3.9 rtx_l2tp

- [x] 3.9.1. Write rtx_l2tp Provider tests (Red)
  - File: `internal/provider/resource_rtx_l2tp_test.go`
  - Test schema to config conversion
  - Test L2TPv3 tunnel parameters
  - _Leverage: schema.TestResourceDataRaw, existing provider test patterns_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create tests for rtx_l2tp. Test L2TPv3 parameters | Restrictions: Follow existing provider test patterns | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.9.2. Fix rtx_l2tp implementation
  - File: `internal/provider/resource_rtx_l2tp.go`
  - Fix any issues found by tests
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Fix any issues to make tests pass | Success: Implementation matches test expectations | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.9.3. Verify rtx_l2tp tests green
  - Run: `go test ./internal/provider/... -v -run TestL2TP`
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 3.10 rtx_nat_masquerade

- [x] 3.10.1. Write rtx_nat_masquerade Provider tests (Red)
  - File: `internal/provider/resource_rtx_nat_masquerade_test.go`
  - Test schema to config conversion
  - Test static entry handling
  - Test protocol-only entries (optional ports)
  - _Leverage: schema.TestResourceDataRaw, existing provider test patterns_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create tests for rtx_nat_masquerade. Include protocol-only entry tests | Restrictions: Follow existing provider test patterns | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.10.2. Fix rtx_nat_masquerade implementation
  - File: `internal/provider/resource_rtx_nat_masquerade.go`
  - Fix any issues found by tests
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Fix any issues to make tests pass | Success: Implementation matches test expectations | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.10.3. Verify rtx_nat_masquerade tests green
  - Run: `go test ./internal/provider/... -v -run TestNATMasquerade`
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 3.11 rtx_nat_static

- [x] 3.11.1. Write rtx_nat_static Provider tests (Red)
  - File: `internal/provider/resource_rtx_nat_static_test.go`
  - Test schema to config conversion
  - Test static mapping entries
  - _Leverage: schema.TestResourceDataRaw, existing provider test patterns_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create tests for rtx_nat_static | Restrictions: Follow existing provider test patterns | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.11.2. Fix rtx_nat_static implementation
  - File: `internal/provider/resource_rtx_nat_static.go`
  - Fix any issues found by tests
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Fix any issues to make tests pass | Success: Implementation matches test expectations | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.11.3. Verify rtx_nat_static tests green
  - Run: `go test ./internal/provider/... -v -run TestNATStatic`
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 3.12 rtx_ospf

- [x] 3.12.1. Write rtx_ospf Provider tests (Red)
  - File: `internal/provider/resource_rtx_ospf_test.go`
  - Test schema to config conversion
  - Test area configuration (backbone, stub, nssa)
  - Test network interface assignment
  - _Leverage: schema.TestResourceDataRaw, existing provider test patterns_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create tests for rtx_ospf. Test area types and interface assignment | Restrictions: Follow existing provider test patterns | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.12.2. Fix rtx_ospf implementation
  - File: `internal/provider/resource_rtx_ospf.go`
  - Fix any issues found by tests
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Fix any issues to make tests pass | Success: Implementation matches test expectations | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.12.3. Verify rtx_ospf tests green
  - Run: `go test ./internal/provider/... -v -run TestOSPF`
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 3.13 rtx_pptp

- [x] 3.13.1. Write rtx_pptp Provider tests (Red)
  - File: `internal/provider/resource_rtx_pptp_test.go`
  - Test schema to config conversion
  - Test server settings
  - _Leverage: schema.TestResourceDataRaw, existing provider test patterns_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create tests for rtx_pptp | Restrictions: Follow existing provider test patterns | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.13.2. Fix rtx_pptp implementation
  - File: `internal/provider/resource_rtx_pptp.go`
  - Fix any issues found by tests
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Fix any issues to make tests pass | Success: Implementation matches test expectations | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.13.3. Verify rtx_pptp tests green
  - Run: `go test ./internal/provider/... -v -run TestPPTP`
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 3.14 rtx_static_route

- [x] 3.14.1. Write rtx_static_route Provider tests (Red)
  - File: `internal/provider/resource_rtx_static_route_test.go`
  - Test schema to config conversion
  - Test gateway types (IP, pp, tunnel)
  - Test optional parameters (metric, weight, hide)
  - _Leverage: schema.TestResourceDataRaw, existing provider test patterns_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create tests for rtx_static_route. Test various gateway types | Restrictions: Follow existing provider test patterns | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.14.2. Fix rtx_static_route implementation
  - File: `internal/provider/resource_rtx_static_route.go`
  - Fix any issues found by tests
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Fix any issues to make tests pass | Success: Implementation matches test expectations | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.14.3. Verify rtx_static_route tests green
  - Run: `go test ./internal/provider/... -v -run TestStaticRoute`
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 3.15 rtx_system

- [x] 3.15.1. Write rtx_system Provider tests (Red)
  - File: `internal/provider/resource_rtx_system_test.go`
  - Test schema to config conversion
  - Test timezone, console, packet-buffer settings
  - _Leverage: schema.TestResourceDataRaw, existing provider test patterns_
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Test Developer | Task: Create tests for rtx_system. Test all system settings | Restrictions: Follow existing provider test patterns | Success: Test file created | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.15.2. Fix rtx_system implementation
  - File: `internal/provider/resource_rtx_system.go`
  - Fix any issues found by tests
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Terraform Provider Developer | Task: Fix any issues to make tests pass | Success: Implementation matches test expectations | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 3.15.3. Verify rtx_system tests green
  - Run: `go test ./internal/provider/... -v -run TestSystem`
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run tests, verify all pass | Success: All tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

---

## Phase 3: Verification

- [x] 4.1. Run all tests and verify coverage
  - Execute: `go test ./internal/client/... ./internal/provider/... -v -cover`
  - Verify no regressions in existing tests
  - Verify new tests pass
  - Check coverage metrics
  - Purpose: Ensure all tests work and meet coverage goals
  - _Requirements: All_
  - _Prompt: Implement the task for spec comprehensive-crud-tests, first run spec-workflow-guide to get the workflow guide then implement the task: Role: QA Engineer | Task: Run full test suite, verify coverage, fix any failures | Restrictions: Do not skip failing tests, address all issues | Success: All tests pass, coverage meets 80%+ target | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._
