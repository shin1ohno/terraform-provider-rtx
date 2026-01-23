# Tasks Document: SFTP Config Parsing Fixes

## Task Overview

| Phase | Tasks | Description | Files | Requirements |
|-------|-------|-------------|-------|--------------|
| 1 | 1.1 | Write holistic tests (Red) | config_file_sftp_test.go | REQ-7 |
| 2 | 2.1-2.5 | Fix parsers (Green) | l2tp.go, l2tp_service.go, system.go, dhcp_binding.go, config_file.go | REQ-1 to REQ-5 |
| 3 | 3.1 | Update main.tf | examples/import/main.tf | REQ-6 |
| 4 | 4.1 | Verify all tests pass | - | All |

**Total: 8 tasks**

## TDD Cycle

```
┌─────────────────────────────────────────────────────────────┐
│ Phase 1: Write Tests (Red)                                  │
│ - Create config_file_sftp_test.go with actual RTX config    │
│ - Tests WILL fail initially (this is expected)              │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ Phase 2: Fix Implementation (Green)                         │
│ - Fix each parser issue one by one                          │
│ - Run tests after each fix                                  │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ Phase 3: Update Config                                      │
│ - Update main.tf dns_servers order                          │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ Phase 4: Verify Green                                       │
│ - Run: go test ./internal/rtx/parsers/... -v                │
│ - All tests must pass                                       │
└─────────────────────────────────────────────────────────────┘
```

---

## Phase 1: Write Holistic Tests (Red)

- [x] 1.1. Create holistic SFTP config tests
  - File: `internal/rtx/parsers/config_file_sftp_test.go`
  - Use actual RTX config from `.spec-workflow/specs/sftp-config-parsing-fixes/test-data/rtx_config_raw.txt`
  - Mask sensitive data in test constant
  - Write test cases for:
    - L2TP Tunnel 1: AlwaysOn, KeepaliveEnabled, Name, Version, Mode
    - L2TP Tunnel 2: Version="l2tp", Mode="lns"
    - L2TP Service: Enabled, Protocols=["l2tpv3", "l2tp"]
    - System: Timezone, Console, PacketBuffers, Statistics
    - DHCP Bindings: 5 bindings with ScopeID=1
  - Tests WILL fail initially (expected - Red phase)
  - _Leverage: Existing patterns from config_file_test.go_
  - _Requirements: REQ-7_
  - _Prompt: Implement the task for spec sftp-config-parsing-fixes: Role: Go Test Developer | Task: Create config_file_sftp_test.go with holistic tests using actual RTX config. Tests should verify all parsing issues identified in requirements. Use table-driven tests where appropriate | Restrictions: Tests will fail initially - this is expected TDD Red phase | Success: Test file created with comprehensive coverage | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

---

## Phase 2: Fix Implementation (Green)

### 2.1 Fix L2TP Tunnel Context Parsing

- [x] 2.1. Fix L2TP tunnel attributes parsing
  - File: `internal/rtx/parsers/l2tp.go`
  - Fix parsing for:
    - `l2tp always-on on` → AlwaysOn = true
    - `l2tp hostname <name>` → Name = name
    - `l2tp keepalive use on <interval> <retry>` → KeepaliveEnabled = true
  - Run: `go test ./internal/rtx/parsers/... -v -run SFTP`
  - _Requirements: REQ-1_
  - _Prompt: Implement the task for spec sftp-config-parsing-fixes: Role: Go Developer | Task: Fix l2tp.go to correctly parse l2tp always-on, l2tp hostname, and l2tp keepalive use commands from tunnel context | Restrictions: Do not break existing functionality | Success: L2TP Tunnel 1 tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 2.2 Fix L2TPv2 Mode Detection

- [x] 2.2. Fix L2TPv2 mode detection
  - File: `internal/rtx/parsers/l2tp.go`
  - When `tunnel encapsulation l2tp` is detected:
    - Set Version = "l2tp"
    - Set Mode = "lns" (not default "l2vpn")
  - Run: `go test ./internal/rtx/parsers/... -v -run SFTP`
  - _Requirements: REQ-2_
  - _Prompt: Implement the task for spec sftp-config-parsing-fixes: Role: Go Developer | Task: Fix l2tp.go so that when tunnel encapsulation l2tp is detected, Mode is set to "lns" instead of staying as default "l2vpn" | Restrictions: L2TPv3 should keep Mode="l2vpn" | Success: L2TP Tunnel 2 tests pass with Mode="lns" | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 2.3 Fix L2TP Service Protocol Parsing

- [x] 2.3. Fix L2TP service protocol parsing
  - File: `internal/rtx/parsers/l2tp_service.go`
  - Parse protocols from `l2tp service on <proto1> <proto2>` format
  - Example: `l2tp service on l2tpv3 l2tp` → Protocols = ["l2tpv3", "l2tp"]
  - Run: `go test ./internal/rtx/parsers/... -v -run SFTP`
  - _Requirements: REQ-3_
  - _Prompt: Implement the task for spec sftp-config-parsing-fixes: Role: Go Developer | Task: Fix l2tp_service.go to extract protocol list from l2tp service on command. Protocols appear after "on" keyword | Restrictions: Handle case with no protocols (empty list) | Success: L2TP Service tests pass with Protocols=["l2tpv3", "l2tp"] | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 2.4 Fix System Config Extraction

- [x] 2.4. Fix system config extraction and parsing
  - Files: `internal/rtx/parsers/config_file.go`, `internal/rtx/parsers/system.go`
  - Debug ExtractSystem to verify commands are collected
  - Fix ParseSystemConfig to handle all formats:
    - timezone, console character/lines/prompt
    - system packet-buffer, statistics traffic/nat
  - Run: `go test ./internal/rtx/parsers/... -v -run SFTP`
  - _Requirements: REQ-4_
  - _Prompt: Implement the task for spec sftp-config-parsing-fixes: Role: Go Developer | Task: Fix system config extraction and parsing. First debug ExtractSystem in config_file.go to verify commands are collected. Then fix ParseSystemConfig in system.go to handle timezone, console, packet-buffer, and statistics commands | Restrictions: Preserve existing functionality | Success: System config tests pass with all fields populated | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### 2.5 Fix DHCP Binding Scope ID

- [x] 2.5. Fix DHCP binding scope ID extraction
  - Files: `internal/rtx/parsers/config_file.go`, `internal/rtx/parsers/dhcp_binding.go`
  - Extract scope ID from each `dhcp scope bind <id> <ip> <mac>` line
  - Don't use hardcoded scopeID=0
  - Run: `go test ./internal/rtx/parsers/... -v -run SFTP`
  - _Requirements: REQ-5_
  - _Prompt: Implement the task for spec sftp-config-parsing-fixes: Role: Go Developer | Task: Fix DHCP binding parsing to extract scope ID from each line instead of using hardcoded value. In config_file.go ExtractDHCPBindings or dhcp_binding.go ParseBindings, extract scope ID from line format: dhcp scope bind <scope_id> <ip> <mac> | Restrictions: Maintain backward compatibility | Success: DHCP binding tests pass with correct ScopeID=1 for all bindings | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

---

## Phase 3: Update Configuration

- [x] 3.1. Update main.tf dns_servers order
  - File: `examples/import/main.tf`
  - Change: `dns_servers = ["1.0.0.1", "1.1.1.1"]` → `dns_servers = ["1.1.1.1", "1.0.0.1"]`
  - This matches the actual RTX config order
  - _Requirements: REQ-6_
  - _Prompt: Implement the task for spec sftp-config-parsing-fixes: Role: Configuration Manager | Task: Update dns_servers order in examples/import/main.tf to match actual RTX config. The correct order is ["1.1.1.1", "1.0.0.1"] | Restrictions: Only change this specific value | Success: dns_servers order matches RTX config | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

---

## Phase 4: Verification

- [x] 4.1. Run all parser tests
  - Execute: `go test ./internal/rtx/parsers/... -v`
  - Verify no regressions in existing tests
  - Verify all new SFTP config tests pass
  - _Requirements: All_
  - _Prompt: Implement the task for spec sftp-config-parsing-fixes: Role: QA Engineer | Task: Run full parser test suite, verify all tests pass including new SFTP config tests | Restrictions: Do not skip failing tests | Success: All parser tests pass | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 4.2. Build provider and run terraform plan
  - Execute in project root: `go build -o terraform-provider-rtx`
  - Execute in examples/import: `terraform plan`
  - Verify no unexpected differences
  - _Requirements: All_
  - _Prompt: Implement the task for spec sftp-config-parsing-fixes: Role: QA Engineer | Task: Build the provider binary and run terraform plan in examples/import directory. Document any differences found | Restrictions: Use latest built binary | Success: terraform plan output captured | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

- [x] 4.3. Iterate until no differences (if needed)
  - IF terraform plan shows differences:
    1. Analyze the differences
    2. Create new spec for each issue (requirements.md, design.md, tasks.md)
    3. Execute fix tasks in parallel where possible
    4. Re-run terraform plan
    5. Repeat until no differences
  - Continue this cycle until `terraform plan` shows no changes
  - _Requirements: All_
  - _Prompt: Implement the task for spec sftp-config-parsing-fixes: Role: QA Engineer + Developer | Task: If terraform plan shows differences, analyze each difference, create new spec if needed, implement fixes, and re-run terraform plan. Repeat until no differences remain. Execute independent fixes in parallel | Restrictions: Each new issue needs its own spec if not covered by existing requirements | Success: terraform plan shows no unexpected changes | Instructions: Before starting, mark this task as in-progress in tasks.md by changing [ ] to [-]. After completing, use log-implementation tool to record what was done, then mark as complete [x]._

### Iteration Cycle Diagram

```
┌─────────────────────────────────────────────────────────────┐
│ Step 1: Build & Plan                                        │
│ - go build -o terraform-provider-rtx                        │
│ - cd examples/import && terraform plan                      │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ Step 2: Check Differences                                   │
│ - No differences? → DONE ✓                                  │
│ - Has differences? → Continue to Step 3                     │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ Step 3: Analyze & Create Spec                               │
│ - Identify root cause for each difference                   │
│ - Create new spec (requirements/design/tasks)               │
│ - Or add to existing spec if related                        │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ Step 4: Implement Fixes (Parallel)                          │
│ - Execute independent tasks in parallel                     │
│ - Run unit tests after each fix                             │
└─────────────────────────────────────────────────────────────┘
                           ↓
              (Return to Step 1)
```
