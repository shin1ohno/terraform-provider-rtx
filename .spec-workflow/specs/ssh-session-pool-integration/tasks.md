# Tasks Document: SSH Session Pool Integration

- [x] 1. Create PooledExecutor struct and basic Run method
  - File: internal/client/pooled_executor.go
  - Create `PooledExecutor` struct with pool, promptDetector, and config dependencies
  - Implement `Run(ctx, cmd)` method that acquires session, executes command, and releases session
  - Add basic error handling for acquisition failures
  - Purpose: Establish core pooled execution mechanism
  - _Leverage: internal/client/ssh_session_pool.go (Acquire/Release), internal/client/working_session.go (Send)_
  - _Requirements: 1.1, 1.2_
  - _Prompt: Implement the task for spec ssh-session-pool-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in concurrent systems and SSH integration | Task: Create PooledExecutor struct implementing the Executor interface's Run method, acquiring sessions from SSHSessionPool, executing commands via workingSession.Send(), and releasing sessions back to pool. Reference existing simpleExecutor patterns in internal/client/simple_executor.go | Restrictions: Do not modify SSHSessionPool or workingSession, follow existing error handling patterns, use Zerolog for logging | Success: Run method successfully acquires session, executes command, releases session, and returns output. Unit test passes with mock pool. | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [x] 2. Implement retry logic with session discard
  - File: internal/client/pooled_executor.go (continue from task 1)
  - Add `executeWithRetry(ctx, cmd, maxRetries)` helper method
  - On session failure: discard session instead of releasing, retry with new session
  - Add exponential backoff between retries (100ms base delay)
  - Purpose: Handle transient session failures gracefully
  - _Leverage: internal/client/ssh_session_pool.go (Discard method needed)_
  - _Requirements: 3.1, 3.2_
  - _Prompt: Implement the task for spec ssh-session-pool-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with expertise in error handling and retry patterns | Task: Implement executeWithRetry method in PooledExecutor with configurable retry count and exponential backoff. Failed sessions should be discarded (not released) to prevent returning broken sessions to pool | Restrictions: Maximum 2 retries, use context for cancellation, log retry attempts with Zerolog | Success: Failed command execution triggers retry with new session, backoff delays are applied, after max retries error is returned with context | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [x] 3. Add Discard method to SSHSessionPool
  - File: internal/client/ssh_session_pool.go (modify existing)
  - Implement `Discard(session *PooledSSHSession)` method
  - Remove session from inUse tracking without returning to available queue
  - Close the underlying workingSession
  - Purpose: Allow executor to discard failed sessions without polluting the pool
  - _Leverage: internal/client/ssh_session_pool.go (existing Release method as reference)_
  - _Requirements: 3.1_
  - _Prompt: Implement the task for spec ssh-session-pool-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in connection pooling and resource management | Task: Add Discard method to SSHSessionPool that removes a session from inUse tracking and closes it without returning to available queue. Follow Release method pattern for locking | Restrictions: Must hold mutex during operation, log discard with session ID, do not signal condition variable (no waiting goroutines need notification) | Success: Discarded sessions are properly closed and removed from tracking, pool state remains consistent, existing tests pass | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [x] 4. Implement administrator mode authentication
  - File: internal/client/pooled_executor.go (continue)
  - Add `prepareSession(ctx, session, needsAdmin)` method
  - Port `authenticateAsAdmin()` logic from simpleExecutor
  - Check session's adminMode state before authenticating
  - Purpose: Support commands requiring administrator privileges
  - _Leverage: internal/client/simple_executor.go (authenticateAsAdmin), internal/client/working_session.go (SetAdminMode)_
  - _Requirements: 2.1, 2.2_
  - _Prompt: Implement the task for spec ssh-session-pool-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with expertise in SSH session management and authentication | Task: Implement prepareSession method that checks if admin privileges are needed (via requiresAdminPrivileges helper) and authenticates on the session if not already in admin mode. Port authentication logic from simpleExecutor.authenticateAsAdmin | Restrictions: Do not modify workingSession, reuse existing prompt detection patterns, track admin state via session.session.SetAdminMode | Success: Commands requiring admin privileges work correctly, sessions already in admin mode are reused without re-authentication | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [x] 5. Implement RunBatch method
  - File: internal/client/pooled_executor.go (continue)
  - Implement `RunBatch(ctx, cmds []string)` method
  - Acquire session once, execute all commands sequentially, release once
  - Aggregate output from all commands
  - Purpose: Efficient batch command execution with single session
  - _Leverage: existing Run method pattern_
  - _Requirements: 1.1, 1.2_
  - _Prompt: Implement the task for spec ssh-session-pool-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in batch processing and resource optimization | Task: Implement RunBatch method that acquires a single session, executes all commands in sequence, aggregates outputs, and releases the session. Handle partial failures appropriately | Restrictions: Single session for entire batch, return aggregated output even if one command fails, use same admin mode preparation as Run | Success: Batch of commands executes with single session acquisition, output from all commands is returned, session is released after batch completes | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [x] 6. Implement SetAdministratorPassword method
  - File: internal/client/pooled_executor.go (continue)
  - Implement `SetAdministratorPassword(ctx, oldPassword, newPassword)` method
  - Port interactive password change sequence from simpleExecutor
  - Handle "Old_Password:", "New_Password:" prompts
  - Purpose: Support administrator password management via pooled sessions
  - _Leverage: internal/client/simple_executor.go (SetAdministratorPassword)_
  - _Requirements: 2.1_
  - _Prompt: Implement the task for spec ssh-session-pool-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with expertise in interactive command sequences and password management | Task: Implement SetAdministratorPassword method following simpleExecutor pattern - acquire session, authenticate as admin, send "administrator password" command, handle interactive prompts for old and new password | Restrictions: Must handle prompt detection correctly, do not log passwords, release session after operation | Success: Administrator password can be changed via pooled session, interactive prompts are handled correctly, session is returned to pool after completion | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [x] 7. Implement SetLoginPassword method
  - File: internal/client/pooled_executor.go (continue)
  - Implement `SetLoginPassword(ctx, newPassword)` method
  - Port login password change sequence from simpleExecutor
  - Handle case where old password prompt may not appear
  - Purpose: Support login password management via pooled sessions
  - _Leverage: internal/client/simple_executor.go (SetLoginPassword)_
  - _Requirements: 2.1_
  - _Prompt: Implement the task for spec ssh-session-pool-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with expertise in interactive command sequences | Task: Implement SetLoginPassword method following simpleExecutor pattern - acquire session, optionally authenticate as admin if configured, send "login password" command, handle password prompts (note: Old_Password prompt may not appear if not previously set) | Restrictions: Handle missing Old_Password prompt gracefully, do not log passwords, release session after operation | Success: Login password can be changed via pooled session, handles both cases (with/without old password), session is returned to pool | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [x] 8. Wire PooledExecutor into rtxClient
  - File: internal/client/client.go (modify existing)
  - Modify executor initialization to use PooledExecutor when pool is enabled
  - Keep simpleExecutor as fallback when pool is disabled
  - Add constructor `NewPooledExecutor(pool, promptDetector, config)`
  - Purpose: Integrate pooled execution into the client
  - _Leverage: internal/client/client.go (existing executor initialization at line 221)_
  - _Requirements: 4.1, 4.2_
  - _Prompt: Implement the task for spec ssh-session-pool-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in dependency injection and client architecture | Task: Modify rtxClient initialization in client.go to use PooledExecutor when sshPoolEnabled is true. Add conditional: if pool enabled use NewPooledExecutor(pool, promptDetector, config), else use existing NewSimpleExecutor | Restrictions: Do not modify Executor interface, maintain backward compatibility, ensure pool is created before PooledExecutor | Success: PooledExecutor is used by default when pool is enabled, simpleExecutor remains available as fallback, existing behavior unchanged when pool disabled | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [x] 9. Create unit tests for PooledExecutor
  - File: internal/client/pooled_executor_test.go
  - Create mock SSHSessionPool for testing
  - Test Run method with successful execution
  - Test retry logic on session failure
  - Test admin mode preparation
  - Purpose: Ensure PooledExecutor reliability
  - _Leverage: internal/client/ssh_session_pool_test.go (testing patterns)_
  - _Requirements: Testing requirement from requirements.md_
  - _Prompt: Implement the task for spec ssh-session-pool-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer with expertise in unit testing and mocking | Task: Create comprehensive unit tests for PooledExecutor covering: successful command execution, retry on failure, admin mode authentication, batch execution. Create mock pool that returns controlled sessions | Restrictions: Use standard testing package, mock at pool interface level, test edge cases (pool exhaustion, session failure) | Success: All PooledExecutor methods have test coverage, tests verify correct pool interaction, admin mode logic tested | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._

- [x] 10. Create integration test verifying session reuse
  - File: internal/client/pooled_executor_integration_test.go
  - Test that multiple commands reuse the same pooled session
  - Verify pool statistics (acquisitions vs new sessions created)
  - Test concurrent command execution
  - Purpose: Verify connection reuse works as expected
  - _Leverage: internal/client/ssh_session_pool_test.go (integration test patterns)_
  - _Requirements: Performance requirement from requirements.md_
  - _Prompt: Implement the task for spec ssh-session-pool-integration, first run spec-workflow-guide to get the workflow guide then implement the task: Role: Go Developer specializing in integration testing and concurrent systems | Task: Create integration tests that verify session reuse by checking pool statistics after multiple command executions. Test concurrent execution to verify pool handles parallel access | Restrictions: May require test build tag for integration tests, use pool's Stats() method if available, test realistic scenarios | Success: Tests demonstrate session reuse (fewer connections than commands), concurrent execution works without errors, pool maintains consistency | After implementation: Mark task [-] as in progress in tasks.md before starting, use log-implementation tool to record artifacts, then mark [x] when complete._
