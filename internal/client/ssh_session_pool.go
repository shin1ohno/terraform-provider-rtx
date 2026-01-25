package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/sh1/terraform-provider-rtx/internal/logging"
)

// SSHPoolConfig configures the SSH connection pool
type SSHPoolConfig struct {
	MaxSessions    int           // Maximum concurrent SSH connections (default: 2)
	IdleTimeout    time.Duration // Close SSH connections after idle time (default: 5m)
	AcquireTimeout time.Duration // Max wait for SSH connection acquisition (default: 30s)
}

// DefaultSSHPoolConfig returns sensible defaults for SSH connection pool
func DefaultSSHPoolConfig() SSHPoolConfig {
	return SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 30 * time.Second,
	}
}

// PooledConnection wraps an SSH connection with its session and pool metadata
type PooledConnection struct {
	client      *ssh.Client     // Independent SSH connection
	session     *workingSession // Single session on this connection
	adminMode   bool            // Whether admin authenticated on this connection
	poolID      string
	lastUsed    time.Time
	useCount    int
	initialized bool
}

// ConnectionFactory is a function that creates a new PooledConnection.
// This is used for dependency injection in testing.
type ConnectionFactory func() (*PooledConnection, error)

// SSHConnectionPool manages a pool of SSH connections to RTX routers
type SSHConnectionPool struct {
	mu                sync.Mutex
	cond              *sync.Cond
	sshConfig         *ssh.ClientConfig // SSH configuration for new connections
	address           string            // host:port for dial
	config            SSHPoolConfig
	available         []*PooledConnection
	inUse             map[*PooledConnection]bool
	pendingCreations  int // Number of connections currently being created
	totalCreated      int
	totalAcquisitions int // Total number of successful acquisitions
	waitCount         int // Number of times an acquire had to wait
	closed            bool
	connectionFactory ConnectionFactory // Optional: custom factory for testing
	skipIdleCleanup   bool              // For testing: skip idle cleanup goroutine
}

// SSHConnectionPoolOption is a functional option for configuring SSHConnectionPool
type SSHConnectionPoolOption func(*SSHConnectionPool)

// WithConnectionFactory sets a custom connection factory for testing
func WithConnectionFactory(factory ConnectionFactory) SSHConnectionPoolOption {
	return func(p *SSHConnectionPool) {
		p.connectionFactory = factory
	}
}

// WithoutIdleCleanup disables the idle cleanup goroutine (for testing)
func WithoutIdleCleanup() SSHConnectionPoolOption {
	return func(p *SSHConnectionPool) {
		p.skipIdleCleanup = true
	}
}

// NewSSHConnectionPool creates a new SSH connection pool
func NewSSHConnectionPool(sshConfig *ssh.ClientConfig, address string, config SSHPoolConfig) *SSHConnectionPool {
	return NewSSHConnectionPoolWithOptions(sshConfig, address, config)
}

// NewSSHConnectionPoolWithOptions creates a new SSH connection pool with options
func NewSSHConnectionPoolWithOptions(sshConfig *ssh.ClientConfig, address string, config SSHPoolConfig, opts ...SSHConnectionPoolOption) *SSHConnectionPool {
	logger := logging.Global()
	logger.Info().
		Int("max_connections", config.MaxSessions).
		Dur("idle_timeout", config.IdleTimeout).
		Dur("acquire_timeout", config.AcquireTimeout).
		Str("address", address).
		Msg("SSH connection pool created")

	pool := &SSHConnectionPool{
		sshConfig: sshConfig,
		address:   address,
		config:    config,
		available: make([]*PooledConnection, 0, config.MaxSessions),
		inUse:     make(map[*PooledConnection]bool),
	}
	pool.cond = sync.NewCond(&pool.mu)

	// Apply options
	for _, opt := range opts {
		if opt != nil {
			opt(pool)
		}
	}

	// Start idle SSH connection cleanup goroutine (unless disabled for testing)
	if !pool.skipIdleCleanup {
		go pool.idleCleanup()
	}

	return pool
}

// Acquire gets an SSH connection from the pool or creates a new one
func (p *SSHConnectionPool) Acquire(ctx context.Context) (*PooledConnection, error) {
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
			return nil, fmt.Errorf("SSH connection pool is closed")
		}

		// Try to get available SSH connection
		if len(p.available) > 0 {
			conn := p.available[len(p.available)-1]
			p.available = p.available[:len(p.available)-1]
			p.inUse[conn] = true
			conn.lastUsed = time.Now()
			conn.useCount++

			p.totalAcquisitions++

			logger.Debug().
				Str("pool_id", conn.poolID).
				Int("use_count", conn.useCount).
				Bool("admin_mode", conn.adminMode).
				Int("available", len(p.available)).
				Int("in_use", len(p.inUse)).
				Int("total_acquisitions", p.totalAcquisitions).
				Msg("Acquired existing SSH connection from pool")

			return conn, nil
		}

		// Can we create a new SSH connection?
		// Include pending creations to prevent race conditions when lock is released during dial
		totalConnections := len(p.inUse) + p.pendingCreations
		if totalConnections < p.config.MaxSessions {
			logger.Debug().
				Int("total_connections", totalConnections).
				Int("pending_creations", p.pendingCreations).
				Int("max_connections", p.config.MaxSessions).
				Msg("Creating new SSH connection for pool")

			p.pendingCreations++
			conn, err := p.createConnection()
			p.pendingCreations--

			if err != nil {
				logger.Error().Err(err).Msg("Failed to create new SSH connection")
				return nil, err
			}
			p.inUse[conn] = true
			p.totalAcquisitions++

			logger.Debug().
				Str("pool_id", conn.poolID).
				Int("in_use", len(p.inUse)).
				Int("total_acquisitions", p.totalAcquisitions).
				Msg("Created and acquired new SSH connection")

			return conn, nil
		}

		// Wait for SSH connection to become available
		if time.Now().After(deadline) {
			logger.Warn().
				Int("in_use", len(p.inUse)).
				Int("max_connections", p.config.MaxSessions).
				Msg("Timeout waiting for available SSH connection")
			return nil, fmt.Errorf("timeout waiting for available SSH connection")
		}

		p.waitCount++
		logger.Debug().
			Int("in_use", len(p.inUse)).
			Int("max_connections", p.config.MaxSessions).
			Int("wait_count", p.waitCount).
			Msg("SSH pool exhausted, waiting for connection to become available")

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

// Release returns an SSH connection to the pool (session stays open)
func (p *SSHConnectionPool) Release(conn *PooledConnection) {
	logger := logging.Global()
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.inUse[conn]; !ok {
		// SSH connection not from this pool or already released
		logger.Warn().
			Str("pool_id", conn.poolID).
			Msg("Attempted to release unknown or already released SSH connection")
		return
	}

	delete(p.inUse, conn)

	if p.closed {
		logger.Debug().
			Str("pool_id", conn.poolID).
			Msg("Pool is closed, closing released SSH connection")
		p.closeConnection(conn)
		return
	}

	// Return connection to pool (session stays open, adminMode preserved)
	conn.lastUsed = time.Now()
	p.available = append(p.available, conn)
	p.cond.Signal()

	logger.Debug().
		Str("pool_id", conn.poolID).
		Int("use_count", conn.useCount).
		Bool("admin_mode", conn.adminMode).
		Int("available", len(p.available)).
		Int("in_use", len(p.inUse)).
		Msg("Released SSH connection back to pool")
}

// Discard removes a failed connection from the pool without returning it to the available queue.
// Use this when a connection has failed and should not be reused.
func (p *SSHConnectionPool) Discard(conn *PooledConnection) {
	logger := logging.Global()
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.inUse[conn]; !ok {
		// SSH connection not from this pool or already released/discarded
		logger.Warn().
			Str("pool_id", conn.poolID).
			Msg("Attempted to discard unknown or already released SSH connection")
		return
	}

	delete(p.inUse, conn)

	// Close both session and client without returning to pool
	p.closeConnection(conn)

	logger.Debug().
		Str("pool_id", conn.poolID).
		Int("use_count", conn.useCount).
		Int("available", len(p.available)).
		Int("in_use", len(p.inUse)).
		Msg("Discarded failed SSH connection from pool")
}

// closeConnection closes both the session and client of a connection
func (p *SSHConnectionPool) closeConnection(conn *PooledConnection) {
	if conn.session != nil {
		conn.session.Close()
	}
	if conn.client != nil {
		conn.client.Close()
	}
}

// Close closes all SSH connections and the pool
func (p *SSHConnectionPool) Close() error {
	logger := logging.Global()
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		logger.Debug().Msg("SSH connection pool already closed")
		return nil
	}

	logger.Debug().
		Int("available", len(p.available)).
		Int("in_use", len(p.inUse)).
		Int("total_created", p.totalCreated).
		Msg("Closing SSH connection pool")

	p.closed = true

	// Close available SSH connections
	for _, conn := range p.available {
		logger.Debug().
			Str("pool_id", conn.poolID).
			Msg("Closing available SSH connection")
		p.closeConnection(conn)
	}
	p.available = nil

	// In-use SSH connections will be closed when released
	p.cond.Broadcast()

	logger.Info().
		Int("total_created", p.totalCreated).
		Int("total_acquisitions", p.totalAcquisitions).
		Int("wait_count", p.waitCount).
		Msg("SSH connection pool closed")

	return nil
}

// createConnection creates a new pooled SSH connection (must hold lock)
func (p *SSHConnectionPool) createConnection() (*PooledConnection, error) {
	logger := logging.Global()

	// Increment totalCreated before releasing lock to ensure unique ID
	p.totalCreated++
	connectionID := p.totalCreated

	p.mu.Unlock() // Release lock during SSH connection creation
	defer p.mu.Lock()

	// Use custom factory if provided (for testing)
	if p.connectionFactory != nil {
		conn, err := p.connectionFactory()
		if err != nil {
			return nil, err
		}
		// Override poolID with sequential ID
		conn.poolID = fmt.Sprintf("ssh-conn-%d", connectionID)
		conn.lastUsed = time.Now()
		conn.useCount = 1
		conn.initialized = true
		return conn, nil
	}

	logger.Debug().
		Int("connection_id", connectionID).
		Str("address", p.address).
		Msg("Dialing new SSH connection for pool")

	// Establish new TCP connection + SSH handshake
	client, err := ssh.Dial("tcp", p.address, p.sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial SSH: %w", err)
	}

	logger.Debug().
		Int("connection_id", connectionID).
		Msg("Creating working session on new connection")

	// Create working session on the new connection
	session, err := newWorkingSession(client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create working session: %w", err)
	}

	pooledConn := &PooledConnection{
		client:      client,
		session:     session,
		adminMode:   false,
		poolID:      fmt.Sprintf("ssh-conn-%d", connectionID),
		lastUsed:    time.Now(),
		useCount:    1,
		initialized: true,
	}

	logger.Debug().
		Str("pool_id", pooledConn.poolID).
		Msg("Created new pooled SSH connection")

	return pooledConn, nil
}

// idleCleanup periodically closes idle SSH connections
func (p *SSHConnectionPool) idleCleanup() {
	logger := logging.Global()
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()
			logger.Debug().Msg("SSH connection pool closed, stopping idle cleanup")
			return
		}

		// Find and close idle SSH connections (keep at least 1)
		now := time.Now()
		remaining := make([]*PooledConnection, 0, len(p.available))
		closedCount := 0

		for _, conn := range p.available {
			if len(remaining) == 0 || now.Sub(conn.lastUsed) < p.config.IdleTimeout {
				remaining = append(remaining, conn)
			} else {
				logger.Debug().
					Str("pool_id", conn.poolID).
					Dur("idle_time", now.Sub(conn.lastUsed)).
					Msg("Closing idle SSH connection")
				p.closeConnection(conn)
				closedCount++
			}
		}

		if closedCount > 0 {
			logger.Debug().
				Int("closed_count", closedCount).
				Int("remaining", len(remaining)).
				Msg("Completed idle SSH connection cleanup")
		}

		p.available = remaining
		p.mu.Unlock()
	}
}

// SSHPoolStats contains SSH connection pool statistics
type SSHPoolStats struct {
	TotalCreated      int
	InUse             int
	Available         int
	MaxSessions       int
	TotalAcquisitions int // Total number of successful acquisitions
	WaitCount         int // Number of times an acquire had to wait for a connection
}

// Stats returns current SSH connection pool statistics
func (p *SSHConnectionPool) Stats() SSHPoolStats {
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

// LogStats logs current SSH connection pool statistics at Info level
func (p *SSHConnectionPool) LogStats() {
	stats := p.Stats()
	logging.Global().Info().
		Int("total_created", stats.TotalCreated).
		Int("in_use", stats.InUse).
		Int("available", stats.Available).
		Int("max_connections", stats.MaxSessions).
		Int("total_acquisitions", stats.TotalAcquisitions).
		Int("wait_count", stats.WaitCount).
		Msg("SSH connection pool statistics")
}

// SetAdminMode sets the admin mode flag for this connection
func (c *PooledConnection) SetAdminMode(admin bool) {
	c.adminMode = admin
}

// Send sends a command to the session and returns the output
func (c *PooledConnection) Send(cmd string) ([]byte, error) {
	if c.session == nil {
		return nil, fmt.Errorf("connection has no active session")
	}
	return c.session.Send(cmd)
}

// Close closes the session (but not the client connection)
func (c *PooledConnection) Close() error {
	if c.session != nil {
		return c.session.Close()
	}
	return nil
}
