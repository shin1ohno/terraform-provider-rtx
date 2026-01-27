# Design Document: SSH Connection Pool Redesign

## Overview

This design transforms `SSHSessionPool` from a session-based pool (multiple sessions on one connection) to a connection-based pool (multiple independent connections, each with one session). This addresses the RTX router limitation where only one SSH session is allowed per connection.

## Current Architecture (Problem)

```
┌─────────────────────────────────────────┐
│           SSHSessionPool                │
├─────────────────────────────────────────┤
│  sshClient *ssh.Client  (1 connection)  │
│                                         │
│  available []*PooledSSHSession          │
│     └─> workingSession (ssh.Session #1) │  ← RTX: OK
│     └─> workingSession (ssh.Session #2) │  ← RTX: REJECTED!
│     └─> workingSession (ssh.Session #3) │  ← RTX: REJECTED!
└─────────────────────────────────────────┘
```

**Problem**: RTX routers reject additional sessions on the same connection with `ssh: rejected: connect failed (open failed)`.

## New Architecture (Solution)

```
┌─────────────────────────────────────────┐
│         SSHConnectionPool               │
├─────────────────────────────────────────┤
│  sshConfig  *ssh.ClientConfig           │
│  address    string                      │
│                                         │
│  available []*PooledConnection          │
│     └─> ssh.Client #1 + workingSession  │  ← Own TCP connection
│     └─> ssh.Client #2 + workingSession  │  ← Own TCP connection
│                                         │
│  inUse map[*PooledConnection]bool       │
│     └─> ssh.Client #3 + workingSession  │  ← Currently executing
└─────────────────────────────────────────┘
```

Each `PooledConnection` has its own:
- TCP connection to the RTX router
- `ssh.Client` (SSH handshake completed)
- `workingSession` (single ssh.Session with terminal initialization)

## Data Structures

### PooledConnection (Replaces PooledSSHSession)

```go
// PooledConnection wraps a complete SSH connection with its session
type PooledConnection struct {
    client         *ssh.Client      // Independent SSH connection
    session        *workingSession  // Single session on this connection
    adminMode      bool             // Admin authentication state
    poolID         string           // Unique identifier for logging
    lastUsed       time.Time        // For idle timeout tracking
    useCount       int              // Number of times this connection was used
}
```

### SSHConnectionPool (Replaces SSHSessionPool)

```go
// SSHConnectionPool manages a pool of SSH connections
type SSHConnectionPool struct {
    mu                sync.Mutex
    cond              *sync.Cond

    // Connection factory
    sshConfig         *ssh.ClientConfig
    address           string           // host:port

    // Pool configuration
    config            SSHPoolConfig

    // Connection tracking
    available         []*PooledConnection
    inUse             map[*PooledConnection]bool

    // Statistics
    totalCreated      int
    totalAcquisitions int
    waitCount         int

    // State
    closed            bool

    // For testing
    connectionFactory ConnectionFactory
    skipIdleCleanup   bool
}
```

### ConnectionFactory

```go
// ConnectionFactory creates a new PooledConnection
// Used for dependency injection in tests
type ConnectionFactory func() (*PooledConnection, error)
```

## Key Changes from Current Implementation

| Aspect | Current (Session Pool) | New (Connection Pool) |
|--------|----------------------|----------------------|
| Pool holds | 1 ssh.Client, N sessions | N ssh.Clients, N sessions |
| Session creation | `newWorkingSession(p.sshClient)` | `ssh.Dial()` + `newWorkingSession(client)` |
| Connection reuse | Session reused, connection shared | Connection reused, session stays open |
| RTX compatibility | ❌ Rejected after 1st session | ✅ Each connection has 1 session |
| Admin state | Per session | Per connection (preserved) |
| Resource cost | 1 TCP connection, N sessions | N TCP connections, N sessions |

## Component Interactions

```
┌─────────────────┐      ┌──────────────────────┐
│PooledExecutor   │──────│ SSHConnectionPool    │
│                 │      │                      │
│ Run(cmd)        │      │ Acquire() → conn     │
│ RunBatch(cmds)  │      │ Release(conn)        │
│ SetAdminPwd()   │      │ Discard(conn)        │
│ SetLoginPwd()   │      │ Close()              │
└─────────────────┘      └──────────────────────┘
                                   │
                                   ▼
                         ┌──────────────────────┐
                         │ PooledConnection     │
                         │                      │
                         │ client (*ssh.Client) │
                         │ session (*working-   │
                         │          Session)    │
                         │ adminMode (bool)     │
                         └──────────────────────┘
                                   │
                                   ▼
                         ┌──────────────────────┐
                         │ RTX Router           │
                         │                      │
                         │ Accepts N TCP conns  │
                         │ 1 session per conn   │
                         └──────────────────────┘
```

## Connection Lifecycle

### 1. Connection Creation

```go
func (p *SSHConnectionPool) createConnection() (*PooledConnection, error) {
    // Step 1: Establish new TCP connection + SSH handshake
    client, err := ssh.Dial("tcp", p.address, p.sshConfig)
    if err != nil {
        return nil, err
    }

    // Step 2: Create working session on the new connection
    session, err := newWorkingSession(client)
    if err != nil {
        client.Close()
        return nil, err
    }

    // Step 3: Wrap in PooledConnection
    return &PooledConnection{
        client:    client,
        session:   session,
        adminMode: false,
        poolID:    fmt.Sprintf("conn-%d", p.totalCreated),
        lastUsed:  time.Now(),
        useCount:  1,
    }, nil
}
```

### 2. Acquire Connection

```go
func (p *SSHConnectionPool) Acquire(ctx context.Context) (*PooledConnection, error) {
    p.mu.Lock()
    defer p.mu.Unlock()

    // Try available pool first (reuse existing connection)
    if len(p.available) > 0 {
        conn := p.available[len(p.available)-1]
        p.available = p.available[:len(p.available)-1]
        p.inUse[conn] = true
        conn.lastUsed = time.Now()
        conn.useCount++
        return conn, nil
    }

    // Create new if under limit
    if len(p.inUse) < p.config.MaxSessions {
        conn, err := p.createConnection()
        if err != nil {
            return nil, err
        }
        p.inUse[conn] = true
        return conn, nil
    }

    // Wait for available connection (with timeout)
    // ...
}
```

### 3. Release Connection

```go
func (p *SSHConnectionPool) Release(conn *PooledConnection) {
    p.mu.Lock()
    defer p.mu.Unlock()

    delete(p.inUse, conn)

    // Return to available pool (session stays open, adminMode preserved)
    conn.lastUsed = time.Now()
    p.available = append(p.available, conn)
    p.cond.Signal()
}
```

### 4. Discard Connection

```go
func (p *SSHConnectionPool) Discard(conn *PooledConnection) {
    p.mu.Lock()
    defer p.mu.Unlock()

    delete(p.inUse, conn)

    // Close session and connection - do not return to pool
    if conn.session != nil {
        conn.session.Close()
    }
    if conn.client != nil {
        conn.client.Close()
    }
}
```

### 5. Close Pool

```go
func (p *SSHConnectionPool) Close() error {
    p.mu.Lock()
    defer p.mu.Unlock()

    p.closed = true

    // Close all available connections
    for _, conn := range p.available {
        conn.session.Close()
        conn.client.Close()
    }
    p.available = nil

    // In-use connections closed on release
    p.cond.Broadcast()
    return nil
}
```

## Admin State Preservation

Since the `workingSession` stays open across acquisitions, admin authentication state is naturally preserved:

```go
// PooledExecutor.prepareSession (unchanged)
func (e *PooledExecutor) prepareSession(ctx context.Context, conn *PooledConnection, needsAdmin bool) error {
    if !needsAdmin {
        return nil
    }

    // Check if already authenticated (persists across acquisitions)
    if conn.adminMode {
        return nil  // Reuse admin state
    }

    // Authenticate on the persistent session
    if err := e.authenticateAsAdmin(ctx, conn); err != nil {
        return err
    }

    conn.adminMode = true  // Preserved for future acquisitions
    return nil
}
```

## Migration Strategy

### Phase 1: Rename and Restructure

1. Rename `SSHSessionPool` → `SSHConnectionPool`
2. Rename `PooledSSHSession` → `PooledConnection`
3. Add `client *ssh.Client` field to `PooledConnection`
4. Add `sshConfig` and `address` fields to pool

### Phase 2: Update Connection Creation

1. Remove single `sshClient` field from pool
2. Update `createConnection()` to dial new connection
3. Each connection gets its own `ssh.Client`

### Phase 3: Update PooledExecutor

1. Update type references: `*PooledSSHSession` → `*PooledConnection`
2. Access session via `conn.session` instead of embedding
3. Access adminMode via `conn.adminMode`

### Phase 4: Update Tests

1. Update mock factory to create full connections
2. Update test assertions for new structure
3. Add integration tests verifying multiple connections

## Files to Modify

| File | Changes |
|------|---------|
| `internal/client/ssh_session_pool.go` | Major rewrite: rename types, add connection factory |
| `internal/client/pooled_executor.go` | Update type references, access patterns |
| `internal/client/pooled_executor_test.go` | Update mocks and assertions |
| `internal/client/ssh_session_pool_test.go` | Update for new connection model |
| `internal/client/client.go` | Pass sshConfig and address to pool constructor |

## Interface Compatibility

The `Executor` interface remains unchanged:

```go
type Executor interface {
    Run(ctx context.Context, cmd string) ([]byte, error)
    RunBatch(ctx context.Context, cmds []string) ([]byte, error)
    SetAdministratorPassword(ctx context.Context, oldPassword, newPassword string) error
    SetLoginPassword(ctx context.Context, newPassword string) error
}
```

`PooledExecutor` signature changes:
- Constructor: `NewPooledExecutor(pool *SSHConnectionPool, ...)` (type change)
- Internal methods: Use `*PooledConnection` instead of `*PooledSSHSession`

## Configuration

No changes to `SSHPoolConfig`:

```go
type SSHPoolConfig struct {
    MaxSessions    int           // Max concurrent connections (renamed semantically)
    IdleTimeout    time.Duration // Close idle connections after this time
    AcquireTimeout time.Duration // Max wait time for connection acquisition
}
```

## Error Handling

### Connection Failure Scenarios

| Scenario | Handling |
|----------|----------|
| `ssh.Dial` fails | Return error, don't add to pool |
| `newWorkingSession` fails | Close client, return error |
| Command fails | Discard connection, retry with new |
| Connection lost | Detected on next use, discarded |
| Idle timeout | Close connection, remove from available |

## Performance Considerations

### Tradeoffs

| Aspect | Impact |
|--------|--------|
| Connection establishment | Higher latency per new connection (~100-200ms) |
| Connection reuse | Eliminates repeated SSH handshakes |
| Memory per connection | Higher (each has own TCP buffer, SSH state) |
| RTX connection limit | Uses N connections instead of 1 |

### RTX Connection Limits

RTX routers typically allow 4-8 concurrent SSH connections. With `MaxSessions: 2` (default), this leaves room for other SSH users/tools while still enabling parallelism.

### Recommended Defaults

```go
DefaultSSHPoolConfig() SSHPoolConfig {
    return SSHPoolConfig{
        MaxSessions:    2,              // 2 concurrent connections
        IdleTimeout:    5 * time.Minute,
        AcquireTimeout: 30 * time.Second,
    }
}
```

## Testing Strategy

### Unit Tests

1. **Connection creation**: Mock `ssh.Dial`, verify connection wrapping
2. **Acquire/Release**: Verify pool state transitions
3. **Discard**: Verify connection cleanup
4. **Concurrent access**: Verify mutex safety
5. **Admin state preservation**: Verify state persists across acquisitions

### Integration Tests

1. **Real RTX connection**: Verify multiple connections work
2. **Connection reuse**: Verify same connection serves multiple commands
3. **Parallel commands**: Verify concurrent execution uses different connections
