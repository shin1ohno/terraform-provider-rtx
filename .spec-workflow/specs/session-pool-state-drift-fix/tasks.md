# Tasks Document: SSH Session Pool for State Drift Fix

## Phase 1: Core SSH Session Pool Infrastructure

- [x] 1. Create SSHSessionPool struct and basic interfaces
  - File: `internal/client/ssh_session_pool.go`
  - Implement `SSHPoolConfig`, `DefaultSSHPoolConfig()`
  - Implement `PooledSSHSession` struct wrapping `workingSession`
  - Implement `SSHSessionPool` struct with mutex and condition variable
  - Implement `NewSSHSessionPool()` constructor
  - Purpose: Establish foundational SSH pool data structures
  - _Leverage: Existing `workingSession` implementation_
  - _Requirements: 1, 3_
  - _Prompt: Implement the task for spec session-pool-state-drift-fix: Role: Go Developer specializing in concurrency | Task: Create SSH session pool data structures in internal/client/ssh_session_pool.go with SSHPoolConfig, PooledSSHSession, and SSHSessionPool as specified in design.md | Restrictions: Follow existing code patterns in client package, use zerolog for logging, ensure thread-safe design | Success: Structs compile, SSH pool can be instantiated with config | Instructions: Mark task as [-] in tasks.md when starting, log implementation details, then mark as [x] when done_

- [x] 2. Implement SSH session acquisition logic
  - File: `internal/client/ssh_session_pool.go`
  - Implement `Acquire(ctx context.Context) (*PooledSSHSession, error)`
  - Handle pool empty case (create new SSH session)
  - Handle SSH session available case (return from pool)
  - Handle pool exhausted case (wait with timeout)
  - Implement context cancellation support
  - Purpose: Enable retrieving SSH sessions from the pool
  - _Leverage: `sync.Cond`, `newWorkingSession()`_
  - _Requirements: 1, 3_
  - _Prompt: Implement the task for spec session-pool-state-drift-fix: Role: Go Developer | Task: Implement Acquire method for SSHSessionPool that handles empty pool, available SSH session, and exhausted pool cases as specified in design.md | Restrictions: Support context cancellation, use mutex properly, log acquisition events | Success: Acquire returns SSH session from pool or creates new, respects timeout | Instructions: Mark task as [-] in tasks.md when starting, log implementation details, then mark as [x] when done_

- [x] 3. Implement SSH session release logic
  - File: `internal/client/ssh_session_pool.go`
  - Implement `Release(session *PooledSSHSession)`
  - Return SSH session to available pool
  - Signal waiting goroutines
  - Handle release of unknown SSH sessions gracefully
  - Purpose: Enable returning SSH sessions to the pool for reuse
  - _Leverage: `sync.Cond.Signal()`_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec session-pool-state-drift-fix: Role: Go Developer | Task: Implement Release method for SSHSessionPool that returns SSH sessions to available pool and signals waiting acquirers | Restrictions: Handle edge cases gracefully, update lastUsed timestamp, log release events | Success: Released SSH sessions become available for next Acquire | Instructions: Mark task as [-] in tasks.md when starting, log implementation details, then mark as [x] when done_

- [x] 4. Implement SSH pool close and cleanup
  - File: `internal/client/ssh_session_pool.go`
  - Implement `Close() error` method
  - Close all available SSH sessions
  - Signal blocked acquirers to unblock
  - Implement `idleCleanup()` goroutine for periodic cleanup
  - Purpose: Proper resource cleanup and idle SSH session management
  - _Leverage: `time.Ticker`, SSH session `Close()` method_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec session-pool-state-drift-fix: Role: Go Developer | Task: Implement Close method and idleCleanup goroutine for SSHSessionPool as specified in design.md | Restrictions: Ensure no goroutine leaks, handle concurrent close calls, log cleanup events | Success: SSH pool closes gracefully, idle SSH sessions are cleaned up periodically | Instructions: Mark task as [-] in tasks.md when starting, log implementation details, then mark as [x] when done_

## Phase 2: Client Integration

- [x] 5. Integrate SSH session pool with rtxClient
  - File: `internal/client/client.go`
  - Add `sshSessionPool *SSHSessionPool` field to `rtxClient`
  - Add `sshPoolEnabled bool` field
  - Initialize SSH pool in `NewClient()` or when first SSH session needed
  - Modify `getExecutor()` to use SSH pool when enabled
  - Implement fallback to non-pooled SSH session on pool failure
  - Purpose: Connect SSH session pool to existing client infrastructure
  - _Leverage: Existing `getExecutor()` pattern_
  - _Requirements: 1, 4_
  - _Prompt: Implement the task for spec session-pool-state-drift-fix: Role: Go Developer | Task: Integrate SSHSessionPool into rtxClient, modifying getExecutor to acquire from SSH pool and return release function | Restrictions: Maintain backward compatibility, implement fallback, do not break existing tests | Success: rtxClient uses SSH pool by default, fallback works when pool fails | Instructions: Mark task as [-] in tasks.md when starting, log implementation details, then mark as [x] when done_

- [x] 6. Update client Close to cleanup SSH pool
  - File: `internal/client/client.go`
  - Modify `Close()` method to close SSH session pool
  - Ensure SSH pool is closed before SSH client
  - Handle partial initialization (SSH pool may be nil)
  - Purpose: Proper cleanup of pooled SSH resources
  - _Leverage: Existing `Close()` method pattern_
  - _Requirements: 1_
  - _Prompt: Implement the task for spec session-pool-state-drift-fix: Role: Go Developer | Task: Update rtxClient.Close() to properly cleanup SSHSessionPool before closing SSH client | Restrictions: Handle nil SSH pool gracefully, log closure events, maintain existing cleanup order | Success: Client closes cleanly with SSH pool, no resource leaks | Instructions: Mark task as [-] in tasks.md when starting, log implementation details, then mark as [x] when done_

## Phase 3: Unit Testing

- [x] 7. Create SSH session pool unit tests
  - File: `internal/client/ssh_session_pool_test.go`
  - Test SSH pool creation with default config
  - Test SSH pool creation with custom config
  - Test Acquire when pool empty
  - Test Acquire when SSH session available (reuse)
  - Test Release returns SSH session to pool
  - Test Close behavior
  - Purpose: Verify SSH pool logic works correctly
  - _Leverage: Standard Go testing, testify/mock if needed_
  - _Requirements: 5_
  - **Completed**: Added 12 basic unit tests including TestDefaultSSHPoolConfig, TestNewSSHSessionPool_DefaultConfig, TestNewSSHSessionPool_CustomConfig (table-driven), TestSSHSessionPool_Acquire_EmptyPool, TestSSHSessionPool_Acquire_ReusesAvailableSession, TestSSHSessionPool_Release_ReturnsToPool, TestSSHSessionPool_Close_ClosesAllSessions, TestSSHSessionPool_Close_Idempotent, TestSSHSessionPool_Stats_ReturnsCorrectValues, TestSSHSessionPool_DoubleRelease_HandledGracefully, TestSSHSessionPool_ReleaseUnknownSession_Ignored. Also added SessionFactory for dependency injection and WithoutIdleCleanup option.

- [x] 8. Add concurrent access tests
  - File: `internal/client/ssh_session_pool_test.go`
  - Test concurrent Acquire calls
  - Test concurrent Release calls
  - Test mixed Acquire/Release under load
  - Verify no data races (run with -race flag)
  - Purpose: Ensure thread safety of SSH pool implementation
  - _Leverage: `sync.WaitGroup`, `go test -race`_
  - _Requirements: 3_
  - **Completed**: Added 7 concurrent tests including TestSSHSessionPool_ConcurrentAcquire, TestSSHSessionPool_ConcurrentRelease, TestSSHSessionPool_MixedAcquireRelease, TestSSHSessionPool_RaceDetector, TestSSHSessionPool_ConcurrentStatsAccess, TestSSHSessionPool_ConcurrentClose, TestSSHSessionPool_HighContention. All tests pass with -race flag.

- [x] 9. Add timeout and error handling tests
  - File: `internal/client/ssh_session_pool_test.go`
  - Test Acquire timeout when SSH pool exhausted
  - Test context cancellation
  - Test behavior when SSH session creation fails
  - Test SSH pool closed error
  - Purpose: Verify error handling works correctly
  - _Leverage: `context.WithTimeout`, `context.WithCancel`_
  - _Requirements: 1, 3_
  - **Completed**: Added 8 timeout/error handling tests including TestSSHSessionPool_AcquireTimeout_PoolExhausted, TestSSHSessionPool_AcquireTimeout_WithContextDeadline, TestSSHSessionPool_ContextCancellation, TestSSHSessionPool_PoolClosedError, TestSSHSessionPool_SessionCreationFailure, TestSSHSessionPool_SessionCreationFailure_CountedCorrectly, TestSSHSessionPool_ReleaseAfterClose, TestSSHSessionPool_AcquireBlocksUntilReleased. Also added additional edge case tests.

## Phase 4: Integration Testing

- [x] 10. Create state drift regression test
  - File: `internal/client/ssh_session_pool_integration_test.go`
  - Test scenario:
    1. Create rtx_system with console.character = "ja.utf8"
    2. Apply configuration
    3. Run terraform plan
    4. Verify no changes detected
  - May require TF_ACC flag
  - Purpose: Verify the original state drift issue is fixed
  - _Leverage: Existing acceptance test patterns_
  - _Requirements: 2, 5_
  - **Completed**: Created integration tests in `internal/client/ssh_session_pool_integration_test.go`:
    - `TestStateDriftRegression_SessionReusePreventsDrift`: Verifies session init only runs once due to reuse
    - `TestStateDriftRegression_MultipleOperationsReuseSameSession`: Confirms sequential operations reuse sessions
    - `TestStateDriftRegression_ConcurrentOperationsLimitedInit`: Verifies init count is minimized under concurrency
    - `TestStateDriftRegression_SessionInitializedFlag`: Checks initialized flag is set correctly
    - `TestStateDriftRegression_RealRouter_ConsoleCharacter`: Documented acceptance test (skipped without TF_ACC)
  - Comprehensive documentation explains the state drift problem and how session pooling solves it

- [x] 11. Update existing tests to work with SSH pool
  - Files: Various test files in `internal/client/`
  - Ensure mock executors work with SSH pool integration
  - Update test setup to handle SSH pool initialization
  - Verify all existing tests pass
  - Purpose: Maintain backward compatibility with existing test suite
  - _Leverage: Existing mock patterns_
  - _Requirements: 4_
  - **Completed**: Verified all existing tests pass with SSH pool integration:
    - Ran `go test ./internal/client/... -count=1` - all tests pass
    - Ran `go test ./internal/client/... -count=1 -race` - no race conditions detected
    - Added integration tests in `ssh_session_pool_integration_test.go`:
      - `TestSSHPoolIntegration_ClientCreation`: Verifies clients work with/without pool
      - `TestSSHPoolIntegration_PoolStatsAccurate`: Verifies stats are maintained correctly
      - `TestSSHPoolIntegration_GracefulDegradation`: Verifies graceful error handling
      - `TestSSHPoolIntegration_SessionLifecycle`: Verifies full session lifecycle
    - Existing MockExecutor and other mocks work correctly with pool integration

## Phase 5: Documentation and Polish

- [x] 12. Add SSH pool statistics and observability
  - File: `internal/client/ssh_session_pool.go`
  - Add `Stats()` method returning SSH pool statistics
  - Track: total created, current active, current available, total acquisitions
  - Log statistics periodically or on demand
  - Purpose: Enable debugging and monitoring of SSH pool behavior
  - _Requirements: Non-functional (Observability)_
  - **Completed**: Enhanced `SSHPoolStats` with `TotalAcquisitions` and `WaitCount` fields. Added `LogStats()` method for on-demand statistics logging at Info level. Updated pool creation log to Info level. Added statistics counters in `Acquire()` method. Pool close now logs total acquisitions and wait count. Added 3 new tests: `TestSSHSessionPool_TotalAcquisitions`, `TestSSHSessionPool_WaitCount`, `TestSSHSessionPool_LogStats`.

- [x] 13. Add provider-level SSH pool configuration (optional enhancement)
  - File: `internal/provider/provider.go`
  - Add optional `ssh_session_pool` block to provider schema
  - Support: `enabled`, `max_sessions`, `idle_timeout`
  - Pass configuration to client creation
  - Purpose: Allow users to tune SSH pool behavior if needed
  - _Requirements: 4_
  - **Completed**: Added `ssh_session_pool` configuration block to provider schema with:
    - `enabled` (bool, default: true) - Enable/disable SSH session pooling
    - `max_sessions` (int, default: 2) - Maximum concurrent SSH sessions
    - `idle_timeout` (string, default: "5m") - Idle session timeout in Go duration format
  - Added `SSHPoolEnabled`, `SSHPoolMaxSessions`, `SSHPoolIdleTimeout` fields to `client.Config`
  - Updated `providerConfigure()` to read and pass SSH pool config to client
  - Updated `NewClient()` to initialize `sshPoolEnabled` from config
  - Updated `Dial()` to parse `idle_timeout` and apply custom pool settings

## Notes

### Dependencies
- Tasks 1-4: Can be done in sequence, each depends on previous
- Task 5: Requires tasks 1-4 complete
- Task 6: Requires task 5
- Tasks 7-9: Can run in parallel after task 4
- Tasks 10-11: Require task 5-6 complete
- Tasks 12-13: Can be done after main implementation

### Testing Requirements
- Run `go test -race ./internal/client/...` after tasks 7-9
- Run full test suite after task 11
- Manual testing recommended after task 10

### File Naming Convention
- Main implementation: `ssh_session_pool.go`
- Tests: `ssh_session_pool_test.go`
- Types: `SSHSessionPool`, `SSHPoolConfig`, `PooledSSHSession`, `SSHPoolStats`
