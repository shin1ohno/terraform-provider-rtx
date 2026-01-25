# Design Document: SSH Session Pool for State Drift Fix

## Overview

This design implements an SSH session pool to prevent SSH session initialization commands from causing Terraform state drift. SSH sessions are reused across operations, ensuring initialization occurs only once per pooled SSH session.

## Steering Document Alignment

### Technical Standards (tech.md)
- Follow existing patterns in `internal/client/`
- Use zerolog for logging SSH pool events
- Maintain consistent error handling patterns
- Use sync primitives for concurrency safety

### Project Structure (structure.md)
- SSH session pool in `internal/client/ssh_session_pool.go`
- SSH session pool tests in `internal/client/ssh_session_pool_test.go`
- Integration with `internal/client/client.go`

## Code Reuse Analysis

### Existing Components to Leverage
- **`workingSession`**: Current SSH session implementation to be pooled
- **`rtxClient`**: Integration point for SSH session pool
- **`logging.Global()`**: Consistent logging interface

### Integration Points
- **`rtxClient.getExecutor()`**: Replace direct SSH session creation with pool acquisition
- **`rtxClient.releaseExecutor()`**: Return SSH sessions to pool instead of closing

## Architecture

### Design Decision: Pool Strategy

**Option A: Simple Mutex-Based Pool**
- Single SSH session per router
- Simpler implementation
- May cause contention under parallel operations

**Option B: Bounded Pool with Wait Queue (Recommended)**
- Configurable max SSH sessions (default: 2)
- Queue for waiting acquisitions
- Better parallelism while respecting router limits

**Rationale for Option B:**
- RTX routers support multiple SSH sessions (4-8 typically)
- Terraform parallel operations benefit from multiple SSH sessions
- Prevents overwhelming the router with too many SSH sessions

### Component Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        rtxClient                             │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                  SSHSessionPool                         ││
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐                 ││
│  │  │   SSH   │  │   SSH   │  │   SSH   │  (max: N)       ││
│  │  │Session#1│  │Session#2│  │  ...    │                 ││
│  │  └────┬────┘  └────┬────┘  └────┬────┘                 ││
│  │       │            │            │                       ││
│  │  ┌────▼────────────▼────────────▼────┐                 ││
│  │  │         Available Queue           │                 ││
│  │  └───────────────────────────────────┘                 ││
│  │                    │                                    ││
│  │  ┌─────────────────▼─────────────────┐                 ││
│  │  │         Waiting Requests          │                 ││
│  │  └───────────────────────────────────┘                 ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

```
Terraform Operation                SSHSessionPool                   RTX Router
       │                                  │                              │
       │  Acquire()                       │                              │
       │─────────────────────────────────►│                              │
       │                                  │  [Pool Empty?]               │
       │                                  │───────┐                      │
       │                                  │       │ Yes: Create new      │
       │                                  │◄──────┘                      │
       │                                  │      SSH Connect             │
       │                                  │─────────────────────────────►│
       │                                  │      SSH Session Ready       │
       │                                  │◄─────────────────────────────│
       │                                  │      Init commands           │
       │                                  │─────────────────────────────►│
       │                                  │                              │
       │      PooledSSHSession            │                              │
       │◄─────────────────────────────────│                              │
       │                                  │                              │
       │      Execute commands            │                              │
       │─────────────────────────────────────────────────────────────────►
       │                                  │                              │
       │  Release(session)                │                              │
       │─────────────────────────────────►│                              │
       │                                  │  [Return to pool]            │
       │                                  │                              │

─────────── Next Operation ───────────

       │  Acquire()                       │                              │
       │─────────────────────────────────►│                              │
       │                                  │  [SSH Session Available]     │
       │      PooledSSHSession (reused!)  │  [NO re-initialization]      │
       │◄─────────────────────────────────│                              │
```

## Components and Interfaces

### Component 1: SSHSessionPool

**Purpose:** Manage reusable SSH sessions

**File:** `internal/client/ssh_session_pool.go`

```go
package client

import (
    "context"
    "fmt"
    "sync"
    "time"

    "golang.org/x/crypto/ssh"
)

// SSHPoolConfig configures the SSH session pool
type SSHPoolConfig struct {
    MaxSessions    int           // Maximum concurrent SSH sessions (default: 2)
    IdleTimeout    time.Duration // Close SSH sessions after idle time (default: 5m)
    AcquireTimeout time.Duration // Max wait for SSH session acquisition (default: 30s)
}

// DefaultSSHPoolConfig returns sensible defaults for SSH session pool
func DefaultSSHPoolConfig() SSHPoolConfig {
    return SSHPoolConfig{
        MaxSessions:    2,
        IdleTimeout:    5 * time.Minute,
        AcquireTimeout: 30 * time.Second,
    }
}

// PooledSSHSession wraps workingSession with pool metadata
type PooledSSHSession struct {
    *workingSession
    poolID      string
    lastUsed    time.Time
    useCount    int
    initialized bool
}

// SSHSessionPool manages a pool of SSH sessions to RTX routers
type SSHSessionPool struct {
    mu           sync.Mutex
    cond         *sync.Cond
    sshClient    *ssh.Client
    config       SSHPoolConfig
    available    []*PooledSSHSession
    inUse        map[*PooledSSHSession]bool
    totalCreated int
    closed       bool
}

// NewSSHSessionPool creates a new SSH session pool
func NewSSHSessionPool(sshClient *ssh.Client, config SSHPoolConfig) *SSHSessionPool {
    pool := &SSHSessionPool{
        sshClient: sshClient,
        config:    config,
        available: make([]*PooledSSHSession, 0, config.MaxSessions),
        inUse:     make(map[*PooledSSHSession]bool),
    }
    pool.cond = sync.NewCond(&pool.mu)

    // Start idle SSH session cleanup goroutine
    go pool.idleCleanup()

    return pool
}

// Acquire gets an SSH session from the pool or creates a new one
func (p *SSHSessionPool) Acquire(ctx context.Context) (*PooledSSHSession, error) {
    p.mu.Lock()
    defer p.mu.Unlock()

    deadline, hasDeadline := ctx.Deadline()
    if !hasDeadline {
        deadline = time.Now().Add(p.config.AcquireTimeout)
    }

    for {
        // Check if closed
        if p.closed {
            return nil, fmt.Errorf("SSH session pool is closed")
        }

        // Try to get available SSH session
        if len(p.available) > 0 {
            session := p.available[len(p.available)-1]
            p.available = p.available[:len(p.available)-1]
            p.inUse[session] = true
            session.lastUsed = time.Now()
            session.useCount++
            return session, nil
        }

        // Can we create a new SSH session?
        totalSessions := len(p.inUse)
        if totalSessions < p.config.MaxSessions {
            session, err := p.createSSHSession()
            if err != nil {
                return nil, err
            }
            p.inUse[session] = true
            return session, nil
        }

        // Wait for SSH session to become available
        if time.Now().After(deadline) {
            return nil, fmt.Errorf("timeout waiting for available SSH session")
        }

        // Use Cond.Wait with timeout emulation
        done := make(chan struct{})
        go func() {
            time.Sleep(100 * time.Millisecond)
            p.cond.Signal()
            close(done)
        }()
        p.cond.Wait()
    }
}

// Release returns an SSH session to the pool
func (p *SSHSessionPool) Release(session *PooledSSHSession) {
    p.mu.Lock()
    defer p.mu.Unlock()

    if _, ok := p.inUse[session]; !ok {
        // SSH session not from this pool or already released
        return
    }

    delete(p.inUse, session)

    if p.closed {
        session.Close()
        return
    }

    session.lastUsed = time.Now()
    p.available = append(p.available, session)
    p.cond.Signal()
}

// Close closes all SSH sessions and the pool
func (p *SSHSessionPool) Close() error {
    p.mu.Lock()
    defer p.mu.Unlock()

    p.closed = true

    // Close available SSH sessions
    for _, session := range p.available {
        session.Close()
    }
    p.available = nil

    // In-use SSH sessions will be closed when released
    p.cond.Broadcast()

    return nil
}

// createSSHSession creates a new pooled SSH session (must hold lock)
func (p *SSHSessionPool) createSSHSession() (*PooledSSHSession, error) {
    p.mu.Unlock() // Release lock during SSH session creation
    defer p.mu.Lock()

    ws, err := newWorkingSession(p.sshClient)
    if err != nil {
        return nil, err
    }

    p.totalCreated++
    return &PooledSSHSession{
        workingSession: ws,
        poolID:         fmt.Sprintf("ssh-session-%d", p.totalCreated),
        lastUsed:       time.Now(),
        useCount:       1,
        initialized:    true, // newWorkingSession already runs init commands
    }, nil
}

// idleCleanup periodically closes idle SSH sessions
func (p *SSHSessionPool) idleCleanup() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        p.mu.Lock()
        if p.closed {
            p.mu.Unlock()
            return
        }

        // Find and close idle SSH sessions (keep at least 1)
        now := time.Now()
        remaining := make([]*PooledSSHSession, 0, len(p.available))
        for _, session := range p.available {
            if len(remaining) == 0 || now.Sub(session.lastUsed) < p.config.IdleTimeout {
                remaining = append(remaining, session)
            } else {
                session.Close()
            }
        }
        p.available = remaining
        p.mu.Unlock()
    }
}

// Stats returns current SSH session pool statistics
type SSHPoolStats struct {
    TotalCreated int
    InUse        int
    Available    int
    MaxSessions  int
}

func (p *SSHSessionPool) Stats() SSHPoolStats {
    p.mu.Lock()
    defer p.mu.Unlock()

    return SSHPoolStats{
        TotalCreated: p.totalCreated,
        InUse:        len(p.inUse),
        Available:    len(p.available),
        MaxSessions:  p.config.MaxSessions,
    }
}
```

### Component 2: Modified rtxClient Integration

**Purpose:** Integrate SSH session pool with existing client

**File:** `internal/client/client.go` (modifications)

```go
// rtxClient modifications
type rtxClient struct {
    // ... existing fields ...
    sshSessionPool *SSHSessionPool
    sshPoolEnabled bool
}

// NewClient modifications
func NewClient(config *Config) (Client, error) {
    // ... existing code ...

    client := &rtxClient{
        // ... existing fields ...
        sshPoolEnabled: true, // Enable by default
    }

    // Initialize SSH session pool
    if client.sshPoolEnabled {
        poolConfig := DefaultSSHPoolConfig()
        client.sshSessionPool = NewSSHSessionPool(client.sshClient, poolConfig)
    }

    return client, nil
}

// getExecutor modification
func (c *rtxClient) getExecutor(ctx context.Context) (Executor, func(), error) {
    if c.sshPoolEnabled && c.sshSessionPool != nil {
        session, err := c.sshSessionPool.Acquire(ctx)
        if err != nil {
            // Fallback to non-pooled SSH session
            logging.Global().Warn().Err(err).Msg("Failed to acquire pooled SSH session, creating new session")
            return c.createNonPooledSession(ctx)
        }

        releaseFunc := func() {
            c.sshSessionPool.Release(session)
        }
        return session, releaseFunc, nil
    }

    return c.createNonPooledSession(ctx)
}

// Close modification
func (c *rtxClient) Close() error {
    if c.sshSessionPool != nil {
        c.sshSessionPool.Close()
    }
    // ... rest of existing close logic ...
}
```

### Component 3: Modified SSH Session Initialization (Alternative Approach)

**Purpose:** Preserve original console settings during initialization

Instead of modifying the pool, we can modify `newWorkingSession` to save and restore settings:

```go
// Alternative: Save/Restore approach in working_session.go
func newWorkingSession(client *ssh.Client) (*workingSession, error) {
    // ... existing SSH session creation code ...

    s := &workingSession{
        client:  client,
        session: session,
        stdin:   stdin,
        stdout:  stdout,
    }

    // Wait for initial prompt
    initialOutput, err := s.readUntilPrompt(10 * time.Second)
    if err != nil {
        s.Close()
        return nil, fmt.Errorf("failed to get initial prompt: %w", err)
    }

    // Save original console character (for restore on close)
    // Note: This only affects in-memory SSH session state, not needed for pool

    // Set character encoding for SSH session communication
    // Use en.ascii to ensure command parsing works correctly
    if _, err := s.executeCommand("console character en.ascii", 5*time.Second); err != nil {
        logger.Warn().Err(err).Msg("Failed to set character encoding (continuing anyway)")
    }

    // Disable paging
    if _, err := s.executeCommand("console lines infinity", 5*time.Second); err != nil {
        logger.Warn().Err(err).Msg("Failed to disable paging (continuing anyway)")
    }

    return s, nil
}
```

## Error Handling

### Error Scenarios

1. **SSH session creation failure:**
   - **Handling:** Log error, return error to caller
   - **Recovery:** Caller may retry with exponential backoff

2. **SSH pool exhausted (all SSH sessions in use):**
   - **Handling:** Wait with timeout, then return error
   - **User Impact:** Operation fails with clear message about SSH session limits

3. **SSH session becomes invalid (connection lost):**
   - **Handling:** Detect on next use, remove from pool, create new
   - **User Impact:** Slightly increased latency for one operation

4. **SSH pool shutdown while SSH sessions in use:**
   - **Handling:** SSH sessions closed when released, blocked acquires unblocked
   - **User Impact:** Graceful shutdown, no data loss

## Testing Strategy

### Unit Testing

1. **SSH Pool Tests** (`ssh_session_pool_test.go`):
   - Test SSH session acquisition when pool empty (creates new)
   - Test SSH session acquisition when SSH session available (reuses)
   - Test concurrent acquisitions with wait
   - Test release returns SSH session to pool
   - Test SSH pool close behavior
   - Test idle cleanup

2. **Mock SSH Session Tests**:
   - Test SSH pool behavior without real SSH connections
   - Test timeout handling
   - Test error scenarios

### Integration Testing

1. **State Drift Test**:
   - Create rtx_system with console.character = "ja.utf8"
   - Apply configuration
   - Run terraform plan multiple times
   - Verify no state drift detected

2. **Concurrent Operation Test**:
   - Run multiple terraform operations in parallel
   - Verify all complete successfully
   - Verify SSH session pool statistics

### Performance Testing

1. **Reuse Efficiency**:
   - Measure time savings from SSH session reuse
   - Compare 10 operations with/without SSH pool

## Migration Guide

### For Existing Users

No action required. SSH session pooling is enabled by default and is transparent to users.

### Configuration Options (Future Enhancement)

```hcl
provider "rtx" {
  # SSH session pool configuration (optional)
  ssh_session_pool {
    enabled       = true  # default: true
    max_sessions  = 2     # default: 2
    idle_timeout  = "5m"  # default: 5m
  }
}
```

## Implementation Order

1. Create SSHSessionPool struct and basic acquire/release
2. Add idle cleanup goroutine
3. Integrate SSH pool with rtxClient.getExecutor()
4. Add unit tests for SSH pool
5. Add integration test for state drift
6. Update documentation
