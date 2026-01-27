# Requirements Document: SSH Connection Pool Redesign

## Introduction

This feature redesigns the SSH session pool to manage multiple independent SSH connections instead of multiple sessions on a single connection. RTX routers only allow one session per SSH connection, making the current design non-functional.

## Problem Statement

**Current Behavior:**
- `SSHSessionPool` creates one persistent `ssh.Client` (connection)
- Attempts to create multiple `ssh.Session` objects on that single connection
- RTX routers reject additional sessions: `ssh: rejected: connect failed (open failed)`

**Root Cause:**
```go
// ssh_session_pool.go: Pool holds one sshClient
type SSHSessionPool struct {
    sshClient *ssh.Client  // Single connection - RTX only allows 1 session here
    ...
}

// createSSHSession attempts to create new session on same connection
func (p *SSHSessionPool) createSSHSession() (*PooledSSHSession, error) {
    ws, err := newWorkingSession(p.sshClient)  // Fails on RTX after first session
    ...
}
```

**RTX Router Limitation:**
- SSH multiplexing (multiple channels on one TCP connection) is not supported
- Each command execution requires its own TCP connection + SSH handshake

## Alignment with Product Vision

This redesign directly enables the original goal of the ssh-session-pool-integration spec:
- Eliminate SSH connection errors during Terraform operations
- Reduce latency by reusing established connections
- Support larger Terraform configurations without hitting RTX connection limits

## Requirements

### Requirement 1: Connection-Based Pool

**User Story:** As a Terraform user, I want the provider to pool SSH connections (not sessions), so that connection reuse works with RTX routers.

#### Acceptance Criteria

1. WHEN the pool is initialized THEN it SHALL manage independent SSH connections (each with its own `ssh.Client`)
2. WHEN a command is executed THEN the system SHALL acquire a connection from the pool, use its single session, and return the connection
3. IF the pool has no available connections AND pool is at capacity THEN the system SHALL wait until a connection becomes available (with timeout)
4. WHEN a connection is returned to the pool THEN it SHALL remain open for reuse (session already closed or kept idle)

### Requirement 2: Connection Lifecycle Management

**User Story:** As a Terraform user, I want connections to be properly managed, so that resources are not leaked.

#### Acceptance Criteria

1. WHEN a connection is acquired THEN the system SHALL create a new session on that connection for command execution
2. WHEN command execution completes THEN the system SHALL close the session but keep the connection open
3. IF a connection fails THEN the system SHALL discard it and create a new one
4. WHEN idle timeout is reached THEN the system SHALL close and remove idle connections

### Requirement 3: Session State Management

**User Story:** As a Terraform user managing RTX routers with admin passwords, I want admin authentication state to persist across commands on the same connection.

#### Acceptance Criteria

1. WHEN a session is created on an existing connection THEN the system SHALL track whether that connection has been authenticated as admin
2. IF connection is already admin-authenticated THEN the new session SHALL inherit admin mode automatically
3. WHEN a connection is returned to the pool THEN its admin state SHALL be preserved for the next acquisition

### Requirement 4: Backward Compatibility

**User Story:** As an existing Terraform user, I want the provider to work seamlessly with the new pool design.

#### Acceptance Criteria

1. WHEN using the redesigned pool THEN the `PooledExecutor` interface SHALL remain unchanged
2. IF pool operations fail THEN error messages SHALL clearly indicate the cause
3. WHEN the pool is disabled THEN the system SHALL fall back to `simpleExecutor` behavior

## Non-Functional Requirements

### Performance

- Connection acquisition latency should be < 10ms when connections are available
- Connection reuse should reduce overall terraform plan time by 50%+ for large configurations
- Pool should support concurrent access without lock contention

### Reliability

- Failed connections must be detected and discarded
- Connection loss during command execution must trigger retry with new connection
- Pool state must remain consistent under concurrent access

### Resource Management

- Idle connections should be closed after configurable timeout (default: 5 minutes)
- Maximum concurrent connections should be configurable (default: 2)
- All connections must be properly closed when pool is destroyed
