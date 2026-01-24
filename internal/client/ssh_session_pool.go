package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/sh1/terraform-provider-rtx/internal/logging"
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

// SessionFactory is a function that creates a new PooledSSHSession.
// This is used for dependency injection in testing.
type SessionFactory func() (*PooledSSHSession, error)

// SSHSessionPool manages a pool of SSH sessions to RTX routers
type SSHSessionPool struct {
	mu                sync.Mutex
	cond              *sync.Cond
	sshClient         *ssh.Client
	config            SSHPoolConfig
	available         []*PooledSSHSession
	inUse             map[*PooledSSHSession]bool
	totalCreated      int
	totalAcquisitions int // Total number of successful acquisitions
	waitCount         int // Number of times an acquire had to wait
	closed            bool
	sessionFactory    SessionFactory // Optional: custom factory for testing
	skipIdleCleanup   bool           // For testing: skip idle cleanup goroutine
}

// SSHSessionPoolOption is a functional option for configuring SSHSessionPool
type SSHSessionPoolOption func(*SSHSessionPool)

// WithSessionFactory sets a custom session factory for testing
func WithSessionFactory(factory SessionFactory) SSHSessionPoolOption {
	return func(p *SSHSessionPool) {
		p.sessionFactory = factory
	}
}

// WithoutIdleCleanup disables the idle cleanup goroutine (for testing)
func WithoutIdleCleanup() SSHSessionPoolOption {
	return func(p *SSHSessionPool) {
		p.skipIdleCleanup = true
	}
}

// NewSSHSessionPool creates a new SSH session pool
func NewSSHSessionPool(sshClient *ssh.Client, config SSHPoolConfig) *SSHSessionPool {
	return NewSSHSessionPoolWithOptions(sshClient, config)
}

// NewSSHSessionPoolWithOptions creates a new SSH session pool with options
func NewSSHSessionPoolWithOptions(sshClient *ssh.Client, config SSHPoolConfig, opts ...SSHSessionPoolOption) *SSHSessionPool {
	logger := logging.Global()
	logger.Info().
		Int("max_sessions", config.MaxSessions).
		Dur("idle_timeout", config.IdleTimeout).
		Dur("acquire_timeout", config.AcquireTimeout).
		Msg("SSH session pool created")

	pool := &SSHSessionPool{
		sshClient: sshClient,
		config:    config,
		available: make([]*PooledSSHSession, 0, config.MaxSessions),
		inUse:     make(map[*PooledSSHSession]bool),
	}
	pool.cond = sync.NewCond(&pool.mu)

	// Apply options
	for _, opt := range opts {
		if opt != nil {
			opt(pool)
		}
	}

	// Start idle SSH session cleanup goroutine (unless disabled for testing)
	if !pool.skipIdleCleanup {
		go pool.idleCleanup()
	}

	return pool
}

// Acquire gets an SSH session from the pool or creates a new one
func (p *SSHSessionPool) Acquire(ctx context.Context) (*PooledSSHSession, error) {
	logger := logging.Global()
	p.mu.Lock()
	defer p.mu.Unlock()

	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		deadline = time.Now().Add(p.config.AcquireTimeout)
	}

	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

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

			p.totalAcquisitions++

			logger.Debug().
				Str("pool_id", session.poolID).
				Int("use_count", session.useCount).
				Int("available", len(p.available)).
				Int("in_use", len(p.inUse)).
				Int("total_acquisitions", p.totalAcquisitions).
				Msg("Acquired existing SSH session from pool")

			return session, nil
		}

		// Can we create a new SSH session?
		totalSessions := len(p.inUse)
		if totalSessions < p.config.MaxSessions {
			logger.Debug().
				Int("total_sessions", totalSessions).
				Int("max_sessions", p.config.MaxSessions).
				Msg("Creating new SSH session for pool")

			session, err := p.createSSHSession()
			if err != nil {
				logger.Error().Err(err).Msg("Failed to create new SSH session")
				return nil, err
			}
			p.inUse[session] = true
			p.totalAcquisitions++

			logger.Debug().
				Str("pool_id", session.poolID).
				Int("in_use", len(p.inUse)).
				Int("total_acquisitions", p.totalAcquisitions).
				Msg("Created and acquired new SSH session")

			return session, nil
		}

		// Wait for SSH session to become available
		if time.Now().After(deadline) {
			logger.Warn().
				Int("in_use", len(p.inUse)).
				Int("max_sessions", p.config.MaxSessions).
				Msg("Timeout waiting for available SSH session")
			return nil, fmt.Errorf("timeout waiting for available SSH session")
		}

		p.waitCount++
		logger.Debug().
			Int("in_use", len(p.inUse)).
			Int("max_sessions", p.config.MaxSessions).
			Int("wait_count", p.waitCount).
			Msg("SSH pool exhausted, waiting for session to become available")

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
	logger := logging.Global()
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.inUse[session]; !ok {
		// SSH session not from this pool or already released
		logger.Warn().
			Str("pool_id", session.poolID).
			Msg("Attempted to release unknown or already released SSH session")
		return
	}

	delete(p.inUse, session)

	if p.closed {
		logger.Debug().
			Str("pool_id", session.poolID).
			Msg("Pool is closed, closing released SSH session")
		// Only close if workingSession is not nil (for test safety)
		if session.workingSession != nil {
			session.Close()
		}
		return
	}

	session.lastUsed = time.Now()
	p.available = append(p.available, session)
	p.cond.Signal()

	logger.Debug().
		Str("pool_id", session.poolID).
		Int("use_count", session.useCount).
		Int("available", len(p.available)).
		Int("in_use", len(p.inUse)).
		Msg("Released SSH session back to pool")
}

// Close closes all SSH sessions and the pool
func (p *SSHSessionPool) Close() error {
	logger := logging.Global()
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		logger.Debug().Msg("SSH session pool already closed")
		return nil
	}

	logger.Debug().
		Int("available", len(p.available)).
		Int("in_use", len(p.inUse)).
		Int("total_created", p.totalCreated).
		Msg("Closing SSH session pool")

	p.closed = true

	// Close available SSH sessions
	for _, session := range p.available {
		logger.Debug().
			Str("pool_id", session.poolID).
			Msg("Closing available SSH session")
		// Only close if workingSession is not nil (for test safety)
		if session.workingSession != nil {
			session.Close()
		}
	}
	p.available = nil

	// In-use SSH sessions will be closed when released
	p.cond.Broadcast()

	logger.Info().
		Int("total_created", p.totalCreated).
		Int("total_acquisitions", p.totalAcquisitions).
		Int("wait_count", p.waitCount).
		Msg("SSH session pool closed")

	return nil
}

// createSSHSession creates a new pooled SSH session (must hold lock)
func (p *SSHSessionPool) createSSHSession() (*PooledSSHSession, error) {
	logger := logging.Global()

	// Increment totalCreated before releasing lock to ensure unique ID
	p.totalCreated++
	sessionID := p.totalCreated

	p.mu.Unlock() // Release lock during SSH session creation
	defer p.mu.Lock()

	// Use custom factory if provided (for testing)
	if p.sessionFactory != nil {
		session, err := p.sessionFactory()
		if err != nil {
			return nil, err
		}
		// Override poolID with sequential ID
		session.poolID = fmt.Sprintf("ssh-session-%d", sessionID)
		session.lastUsed = time.Now()
		session.useCount = 1
		session.initialized = true
		return session, nil
	}

	logger.Debug().
		Int("session_id", sessionID).
		Msg("Creating working session for pool")

	ws, err := newWorkingSession(p.sshClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create working session: %w", err)
	}

	pooledSession := &PooledSSHSession{
		workingSession: ws,
		poolID:         fmt.Sprintf("ssh-session-%d", sessionID),
		lastUsed:       time.Now(),
		useCount:       1,
		initialized:    true, // newWorkingSession already runs init commands
	}

	logger.Debug().
		Str("pool_id", pooledSession.poolID).
		Msg("Created new pooled SSH session")

	return pooledSession, nil
}

// idleCleanup periodically closes idle SSH sessions
func (p *SSHSessionPool) idleCleanup() {
	logger := logging.Global()
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()
			logger.Debug().Msg("SSH session pool closed, stopping idle cleanup")
			return
		}

		// Find and close idle SSH sessions (keep at least 1)
		now := time.Now()
		remaining := make([]*PooledSSHSession, 0, len(p.available))
		closedCount := 0

		for _, session := range p.available {
			if len(remaining) == 0 || now.Sub(session.lastUsed) < p.config.IdleTimeout {
				remaining = append(remaining, session)
			} else {
				logger.Debug().
					Str("pool_id", session.poolID).
					Dur("idle_time", now.Sub(session.lastUsed)).
					Msg("Closing idle SSH session")
				// Only close if workingSession is not nil (for test safety)
				if session.workingSession != nil {
					session.Close()
				}
				closedCount++
			}
		}

		if closedCount > 0 {
			logger.Debug().
				Int("closed_count", closedCount).
				Int("remaining", len(remaining)).
				Msg("Completed idle SSH session cleanup")
		}

		p.available = remaining
		p.mu.Unlock()
	}
}

// SSHPoolStats contains SSH session pool statistics
type SSHPoolStats struct {
	TotalCreated      int
	InUse             int
	Available         int
	MaxSessions       int
	TotalAcquisitions int // Total number of successful acquisitions
	WaitCount         int // Number of times an acquire had to wait for a session
}

// Stats returns current SSH session pool statistics
func (p *SSHSessionPool) Stats() SSHPoolStats {
	p.mu.Lock()
	defer p.mu.Unlock()

	return SSHPoolStats{
		TotalCreated:      p.totalCreated,
		InUse:             len(p.inUse),
		Available:         len(p.available),
		MaxSessions:       p.config.MaxSessions,
		TotalAcquisitions: p.totalAcquisitions,
		WaitCount:         p.waitCount,
	}
}

// LogStats logs current SSH session pool statistics at Info level
func (p *SSHSessionPool) LogStats() {
	stats := p.Stats()
	logging.Global().Info().
		Int("total_created", stats.TotalCreated).
		Int("in_use", stats.InUse).
		Int("available", stats.Available).
		Int("max_sessions", stats.MaxSessions).
		Int("total_acquisitions", stats.TotalAcquisitions).
		Int("wait_count", stats.WaitCount).
		Msg("SSH session pool statistics")
}
