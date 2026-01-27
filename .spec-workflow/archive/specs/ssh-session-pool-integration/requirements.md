# Requirements Document: SSH Session Pool Integration

## Introduction

This feature integrates the existing `SSHSessionPool` into the executor layer to eliminate redundant SSH connections during Terraform operations. Currently, the `simpleExecutor` creates a new SSH connection for each command, causing connection exhaustion and rate limiting errors on RTX routers when managing many resources.

## Problem Statement

**Current Behavior:**
- `simpleExecutor.Run()` creates a new SSH connection per command (dial → execute → close)
- A terraform plan with 60+ resources results in 60+ SSH connections in rapid succession
- RTX routers have connection rate limits and session limits, causing `ssh: handshake failed: EOF` errors

**Root Cause:**
```go
// client.go line 213: Pool is created
c.sshSessionPool = NewSSHSessionPool(sshClient, poolConfig)

// client.go line 221: But simpleExecutor is used (pool unused)
c.executor = NewSimpleExecutor(sshConfig, addr, c.promptDetector, c.config)
```

The `SSHSessionPool` exists but is not wired into the command execution path.

## Alignment with Product Vision

This feature directly supports provider reliability and user experience by:
- Eliminating SSH connection errors during normal operations
- Reducing latency by reusing established connections
- Supporting larger Terraform configurations without hitting RTX limits

## Requirements

### Requirement 1: Pooled Executor Implementation

**User Story:** As a Terraform user, I want the provider to reuse SSH connections, so that I don't encounter connection errors when managing many resources.

#### Acceptance Criteria

1. WHEN a command is executed THEN the system SHALL acquire a session from the pool instead of creating a new SSH connection
2. WHEN command execution completes THEN the system SHALL return the session to the pool for reuse
3. IF the pool has no available sessions AND pool is at capacity THEN the system SHALL wait until a session becomes available (with timeout)
4. IF session acquisition times out THEN the system SHALL return a clear error message

### Requirement 2: Administrator Mode Support

**User Story:** As a Terraform user managing RTX routers with admin passwords, I want pooled sessions to support administrator authentication, so that configuration commands work correctly.

#### Acceptance Criteria

1. WHEN a command requires administrator privileges THEN the system SHALL authenticate on the pooled session before executing
2. IF a session is already in administrator mode THEN the system SHALL reuse it without re-authenticating
3. WHEN a session is returned to the pool THEN the system SHALL track its administrator mode state for reuse

### Requirement 3: Session Health Management

**User Story:** As a Terraform user, I want the provider to handle stale or broken sessions automatically, so that operations don't fail due to connection issues.

#### Acceptance Criteria

1. WHEN a session fails during command execution THEN the system SHALL discard the session and retry with a new session
2. IF a session has been idle beyond the configured timeout THEN the system SHALL close and remove it from the pool
3. WHEN the SSH client connection is lost THEN the system SHALL invalidate all pooled sessions

### Requirement 4: Backward Compatibility

**User Story:** As an existing Terraform user, I want the provider to work without configuration changes, so that my workflows are not disrupted.

#### Acceptance Criteria

1. WHEN upgrading to the new version THEN existing Terraform configurations SHALL work without modification
2. IF session pooling is disabled via configuration THEN the system SHALL fall back to the current simpleExecutor behavior
3. WHEN errors occur THEN error messages SHALL be consistent with existing error formats

## Non-Functional Requirements

### Code Architecture and Modularity

- **Single Responsibility Principle**: `PooledExecutor` handles only command execution via pooled sessions
- **Modular Design**: Pool management remains in `SSHSessionPool`, executor uses pool as dependency
- **Dependency Management**: `PooledExecutor` depends on `SSHSessionPool` interface, not concrete implementation
- **Clear Interfaces**: Define `Executor` interface that both `simpleExecutor` and `PooledExecutor` implement

### Performance

- Session acquisition latency should be < 10ms when sessions are available
- Connection reuse should reduce overall terraform plan time by 50%+ for large configurations
- Pool should support concurrent access without lock contention

### Reliability

- Failed sessions must not corrupt pool state
- Connection loss must be detected and handled gracefully
- Retry logic must prevent cascading failures

### Testing

- Unit tests for `PooledExecutor` with mock pool
- Integration tests verifying connection reuse
- Stress tests with concurrent command execution
