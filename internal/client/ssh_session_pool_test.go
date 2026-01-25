package client

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Helper functions and mock connection factory
// =============================================================================

// mockConnectionFactory creates a simple connection factory for testing
func mockConnectionFactory() ConnectionFactory {
	return func() (*PooledConnection, error) {
		return &PooledConnection{
			client:      nil, // nil is fine for tests that don't use actual SSH
			session:     nil,
			adminMode:   false,
			poolID:      "mock-conn",
			lastUsed:    time.Now(),
			useCount:    0,
			initialized: false,
		}, nil
	}
}

// errorConnectionFactory creates a connection factory that always returns an error
func errorConnectionFactory(err error) ConnectionFactory {
	return func() (*PooledConnection, error) {
		return nil, err
	}
}

// countingConnectionFactory creates a connection factory that counts creations
func countingConnectionFactory(counter *int32) ConnectionFactory {
	return func() (*PooledConnection, error) {
		atomic.AddInt32(counter, 1)
		return &PooledConnection{
			client:      nil,
			session:     nil,
			adminMode:   false,
			poolID:      "counting-conn",
			lastUsed:    time.Now(),
			useCount:    0,
			initialized: false,
		}, nil
	}
}

// createTestPool creates a pool with mock connection factory and no idle cleanup
func createTestPool(config SSHPoolConfig) *SSHConnectionPool {
	return NewSSHConnectionPoolWithOptions(
		nil, // sshConfig not needed with factory
		"",  // address not needed with factory
		config,
		WithConnectionFactory(mockConnectionFactory()),
		WithoutIdleCleanup(),
	)
}

// createTestPoolWithFactory creates a pool with custom connection factory
func createTestPoolWithFactory(config SSHPoolConfig, factory ConnectionFactory) *SSHConnectionPool {
	return NewSSHConnectionPoolWithOptions(
		nil,
		"",
		config,
		WithConnectionFactory(factory),
		WithoutIdleCleanup(),
	)
}

// =============================================================================
// Task 7: Basic Unit Tests
// =============================================================================

func TestDefaultSSHPoolConfig(t *testing.T) {
	config := DefaultSSHPoolConfig()

	assert.Equal(t, 2, config.MaxSessions, "default MaxSessions should be 2")
	assert.Equal(t, 5*time.Minute, config.IdleTimeout, "default IdleTimeout should be 5m")
	assert.Equal(t, 30*time.Second, config.AcquireTimeout, "default AcquireTimeout should be 30s")
}

func TestNewSSHConnectionPool_DefaultConfig(t *testing.T) {
	config := DefaultSSHPoolConfig()
	pool := createTestPool(config)
	defer pool.Close()

	stats := pool.Stats()
	assert.Equal(t, 0, stats.TotalCreated, "new pool should have 0 connections created")
	assert.Equal(t, 0, stats.InUse, "new pool should have 0 connections in use")
	assert.Equal(t, 0, stats.Available, "new pool should have 0 available connections")
	assert.Equal(t, config.MaxSessions, stats.MaxSessions, "MaxSessions should match config")
}

func TestNewSSHConnectionPool_CustomConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      SSHPoolConfig
		wantMax     int
		wantIdle    time.Duration
		wantAcquire time.Duration
	}{
		{
			name: "single connection pool",
			config: SSHPoolConfig{
				MaxSessions:    1,
				IdleTimeout:    1 * time.Minute,
				AcquireTimeout: 5 * time.Second,
			},
			wantMax:     1,
			wantIdle:    1 * time.Minute,
			wantAcquire: 5 * time.Second,
		},
		{
			name: "large pool",
			config: SSHPoolConfig{
				MaxSessions:    10,
				IdleTimeout:    30 * time.Minute,
				AcquireTimeout: 60 * time.Second,
			},
			wantMax:     10,
			wantIdle:    30 * time.Minute,
			wantAcquire: 60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := createTestPool(tt.config)
			defer pool.Close()

			stats := pool.Stats()
			assert.Equal(t, tt.wantMax, stats.MaxSessions)
		})
	}
}

func TestSSHConnectionPool_Acquire_EmptyPool(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()
	conn, err := pool.Acquire(ctx)

	require.NoError(t, err, "Acquire should succeed on empty pool")
	require.NotNil(t, conn, "acquired connection should not be nil")
	assert.Equal(t, 1, conn.useCount, "first acquisition should set useCount to 1")
	assert.True(t, conn.initialized, "connection should be initialized")

	stats := pool.Stats()
	assert.Equal(t, 1, stats.TotalCreated, "should have created 1 connection")
	assert.Equal(t, 1, stats.InUse, "should have 1 connection in use")
	assert.Equal(t, 0, stats.Available, "should have 0 available connections")
}

func TestSSHConnectionPool_Acquire_ReusesAvailableConnection(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// Acquire and release a connection
	conn1, err := pool.Acquire(ctx)
	require.NoError(t, err)
	pool.Release(conn1)

	// Acquire again - should get the same connection
	conn2, err := pool.Acquire(ctx)
	require.NoError(t, err)

	assert.Equal(t, conn1, conn2, "should reuse the same connection")
	assert.Equal(t, 2, conn2.useCount, "reused connection should have incremented useCount")

	stats := pool.Stats()
	assert.Equal(t, 1, stats.TotalCreated, "should only have created 1 connection total")
}

func TestSSHConnectionPool_Release_ReturnsToPool(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)

	stats := pool.Stats()
	assert.Equal(t, 1, stats.InUse)
	assert.Equal(t, 0, stats.Available)

	pool.Release(conn)

	stats = pool.Stats()
	assert.Equal(t, 0, stats.InUse, "released connection should not be in use")
	assert.Equal(t, 1, stats.Available, "released connection should be available")
}

func TestSSHConnectionPool_Close_ClosesAllConnections(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    3,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)

	ctx := context.Background()

	// Acquire and release multiple connections
	var conns []*PooledConnection
	for i := 0; i < 3; i++ {
		conn, err := pool.Acquire(ctx)
		require.NoError(t, err)
		conns = append(conns, conn)
	}

	// Release all
	for _, c := range conns {
		pool.Release(c)
	}

	stats := pool.Stats()
	assert.Equal(t, 3, stats.Available)

	// Close the pool
	err := pool.Close()
	require.NoError(t, err)

	stats = pool.Stats()
	assert.Equal(t, 0, stats.Available, "close should clear available connections")
}

func TestSSHConnectionPool_Close_Idempotent(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)

	err := pool.Close()
	require.NoError(t, err, "first close should succeed")

	err = pool.Close()
	require.NoError(t, err, "second close should succeed (idempotent)")
}

func TestSSHConnectionPool_Stats_ReturnsCorrectValues(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    5,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// Initial state
	stats := pool.Stats()
	assert.Equal(t, 0, stats.TotalCreated)
	assert.Equal(t, 0, stats.InUse)
	assert.Equal(t, 0, stats.Available)
	assert.Equal(t, 5, stats.MaxSessions)

	// Acquire 3 connections
	var conns []*PooledConnection
	for i := 0; i < 3; i++ {
		c, _ := pool.Acquire(ctx)
		conns = append(conns, c)
	}

	stats = pool.Stats()
	assert.Equal(t, 3, stats.TotalCreated)
	assert.Equal(t, 3, stats.InUse)
	assert.Equal(t, 0, stats.Available)

	// Release 2 connections
	pool.Release(conns[0])
	pool.Release(conns[1])

	stats = pool.Stats()
	assert.Equal(t, 3, stats.TotalCreated)
	assert.Equal(t, 1, stats.InUse)
	assert.Equal(t, 2, stats.Available)
}

func TestSSHConnectionPool_DoubleRelease_HandledGracefully(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)

	// First release
	pool.Release(conn)

	stats := pool.Stats()
	assert.Equal(t, 1, stats.Available)

	// Second release should be ignored
	pool.Release(conn)

	stats = pool.Stats()
	assert.Equal(t, 1, stats.Available, "double release should not add duplicate connection")
}

func TestSSHConnectionPool_ReleaseUnknownConnection_Ignored(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	// Create a connection that was never acquired from this pool
	unknownConn := &PooledConnection{
		client:      nil,
		session:     nil,
		adminMode:   false,
		poolID:      "unknown-conn",
		lastUsed:    time.Now(),
		useCount:    1,
		initialized: true,
	}

	// Release should be handled gracefully (no panic)
	pool.Release(unknownConn)

	stats := pool.Stats()
	assert.Equal(t, 0, stats.Available, "unknown connection should not be added to pool")
}

// =============================================================================
// Task 8: Concurrent Access Tests
// =============================================================================

func TestSSHConnectionPool_ConcurrentAcquire(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	config := SSHPoolConfig{
		MaxSessions:    5,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 10 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	numGoroutines := 20
	var wg sync.WaitGroup
	successCount := int32(0)
	errorCount := int32(0)

	ctx := context.Background()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			conn, err := pool.Acquire(ctx)
			if err != nil {
				atomic.AddInt32(&errorCount, 1)
				return
			}

			atomic.AddInt32(&successCount, 1)

			// Simulate work
			time.Sleep(5 * time.Millisecond)

			pool.Release(conn)
		}()
	}

	wg.Wait()

	assert.Equal(t, int32(numGoroutines), successCount, "all goroutines should succeed")
	assert.Equal(t, int32(0), errorCount, "no errors expected")

	stats := pool.Stats()
	assert.Equal(t, 0, stats.InUse, "all connections should be released")
}

func TestSSHConnectionPool_ConcurrentRelease(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	config := SSHPoolConfig{
		MaxSessions:    10,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 10 * time.Second,
	}

	pool := &SSHConnectionPool{
		sshConfig: nil,
		address:   "",
		config:    config,
		available: make([]*PooledConnection, 0, config.MaxSessions),
		inUse:     make(map[*PooledConnection]bool),
	}
	pool.cond = sync.NewCond(&pool.mu)

	// Create connections that are "in use"
	conns := make([]*PooledConnection, config.MaxSessions)
	for i := 0; i < config.MaxSessions; i++ {
		conn := &PooledConnection{
			client:      nil,
			session:     nil,
			adminMode:   false,
			poolID:      "release-test-" + string(rune('0'+i)),
			lastUsed:    time.Now(),
			useCount:    1,
			initialized: true,
		}
		conns[i] = conn
		pool.inUse[conn] = true
	}

	var wg sync.WaitGroup

	// Release all connections concurrently
	for i := 0; i < config.MaxSessions; i++ {
		wg.Add(1)
		go func(conn *PooledConnection) {
			defer wg.Done()
			pool.Release(conn)
		}(conns[i])
	}

	wg.Wait()

	stats := pool.Stats()
	assert.Equal(t, 0, stats.InUse, "all connections should be released")
	assert.Equal(t, config.MaxSessions, stats.Available, "all connections should be available")
}

func TestSSHConnectionPool_MixedAcquireRelease(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	config := SSHPoolConfig{
		MaxSessions:    3,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	numOperations := 50
	var wg sync.WaitGroup
	var successCount int32

	ctx := context.Background()

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			conn, err := pool.Acquire(ctx)
			if err != nil {
				return
			}

			// Variable work duration
			time.Sleep(time.Duration(id%5) * time.Millisecond)

			pool.Release(conn)
			atomic.AddInt32(&successCount, 1)
		}(i)
	}

	wg.Wait()

	t.Logf("Completed %d successful operations", successCount)
	assert.Greater(t, successCount, int32(0), "should have some successful operations")

	stats := pool.Stats()
	assert.Equal(t, 0, stats.InUse, "all connections should be released")
}

func TestSSHConnectionPool_RaceDetector(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race detector test in short mode")
	}

	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 2 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	done := make(chan bool)
	iterations := 100

	// Goroutine 1: Continuous acquire/release
	go func() {
		ctx := context.Background()
		for i := 0; i < iterations; i++ {
			conn, err := pool.Acquire(ctx)
			if err == nil && conn != nil {
				pool.Release(conn)
			}
		}
		done <- true
	}()

	// Goroutine 2: Continuous acquire/release
	go func() {
		ctx := context.Background()
		for i := 0; i < iterations; i++ {
			conn, err := pool.Acquire(ctx)
			if err == nil && conn != nil {
				pool.Release(conn)
			}
		}
		done <- true
	}()

	// Goroutine 3: Read stats concurrently
	go func() {
		for i := 0; i < iterations; i++ {
			_ = pool.Stats()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	t.Log("Race detector test completed without detecting races")
}

func TestSSHConnectionPool_ConcurrentStatsAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	config := SSHPoolConfig{
		MaxSessions:    5,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	numGoroutines := 50
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				stats := pool.Stats()
				assert.Equal(t, config.MaxSessions, stats.MaxSessions)
			}
		}()
	}

	wg.Wait()
}

func TestSSHConnectionPool_ConcurrentClose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	config := SSHPoolConfig{
		MaxSessions:    3,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 1 * time.Second,
	}
	pool := createTestPool(config)

	var wg sync.WaitGroup
	ctx := context.Background()

	// Start some acquire operations
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := pool.Acquire(ctx)
			if err == nil && conn != nil {
				time.Sleep(50 * time.Millisecond)
				pool.Release(conn)
			}
		}()
	}

	// Close the pool while operations are in progress
	time.Sleep(10 * time.Millisecond)
	err := pool.Close()
	require.NoError(t, err)

	wg.Wait()

	// Verify pool is closed
	_, err = pool.Acquire(ctx)
	assert.Error(t, err, "acquire should fail on closed pool")
}

func TestSSHConnectionPool_HighContention(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high contention test in short mode")
	}

	// Small pool to maximize contention
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 10 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	numGoroutines := 20
	numIterations := 10
	var wg sync.WaitGroup
	var totalOperations int32

	ctx := context.Background()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				conn, err := pool.Acquire(ctx)
				if err != nil {
					continue
				}

				time.Sleep(time.Millisecond)

				pool.Release(conn)
				atomic.AddInt32(&totalOperations, 1)
			}
		}()
	}

	wg.Wait()

	total := atomic.LoadInt32(&totalOperations)
	t.Logf("Completed %d operations under high contention", total)
	assert.Greater(t, total, int32(0), "should complete some operations")

	stats := pool.Stats()
	assert.Equal(t, 0, stats.InUse, "all connections should be released")
}

// =============================================================================
// Task 9: Timeout and Error Handling Tests
// =============================================================================

func TestSSHConnectionPool_AcquireTimeout_PoolExhausted(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    1,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 200 * time.Millisecond, // Short timeout for test
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// Acquire the only available connection
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)

	// Try to acquire another - should timeout
	start := time.Now()
	_, err = pool.Acquire(ctx)
	elapsed := time.Since(start)

	assert.Error(t, err, "should timeout waiting for connection")
	assert.Contains(t, err.Error(), "timeout", "error should mention timeout")
	assert.GreaterOrEqual(t, elapsed, 200*time.Millisecond, "should wait at least AcquireTimeout")
	assert.Less(t, elapsed, 500*time.Millisecond, "should not wait too long")

	// Cleanup
	pool.Release(conn)
}

func TestSSHConnectionPool_AcquireTimeout_WithContextDeadline(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    1,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 10 * time.Second, // Long default timeout
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// Acquire the only available connection
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)

	// Try to acquire with a short context deadline
	shortCtx, cancel := context.WithTimeout(ctx, 150*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err = pool.Acquire(shortCtx)
	elapsed := time.Since(start)

	assert.Error(t, err, "should timeout with context deadline")
	assert.GreaterOrEqual(t, elapsed, 150*time.Millisecond)
	assert.Less(t, elapsed, 500*time.Millisecond)

	// Cleanup
	pool.Release(conn)
}

func TestSSHConnectionPool_ContextCancellation(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    1,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// Acquire the only available connection
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)

	// Try to acquire with a cancelled context
	cancelCtx, cancel := context.WithCancel(ctx)

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err = pool.Acquire(cancelCtx)
	elapsed := time.Since(start)

	assert.ErrorIs(t, err, context.Canceled, "should return context.Canceled")
	assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond)
	assert.Less(t, elapsed, 300*time.Millisecond)

	// Cleanup
	pool.Release(conn)
}

func TestSSHConnectionPool_PoolClosedError(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)

	// Close the pool
	err := pool.Close()
	require.NoError(t, err)

	// Try to acquire from closed pool
	ctx := context.Background()
	_, err = pool.Acquire(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed", "error should mention pool is closed")
}

func TestSSHConnectionPool_ConnectionCreationFailure(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	creationError := errors.New("SSH connection creation failed")
	pool := createTestPoolWithFactory(config, errorConnectionFactory(creationError))
	defer pool.Close()

	ctx := context.Background()
	_, err := pool.Acquire(ctx)

	assert.Error(t, err)
	assert.ErrorIs(t, err, creationError, "should return the factory error")
}

func TestSSHConnectionPool_ConnectionCreationFailure_CountedCorrectly(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	callCount := int32(0)
	failAfter := int32(2)

	factory := func() (*PooledConnection, error) {
		count := atomic.AddInt32(&callCount, 1)
		if count > failAfter {
			return nil, errors.New("max connections reached simulation")
		}
		return &PooledConnection{
			client:      nil,
			session:     nil,
			adminMode:   false,
			poolID:      "test-conn",
			lastUsed:    time.Now(),
			useCount:    0,
			initialized: false,
		}, nil
	}

	pool := createTestPoolWithFactory(config, factory)
	defer pool.Close()

	ctx := context.Background()

	// First two should succeed
	conn1, err1 := pool.Acquire(ctx)
	conn2, err2 := pool.Acquire(ctx)
	require.NoError(t, err1)
	require.NoError(t, err2)

	pool.Release(conn1)
	pool.Release(conn2)

	stats := pool.Stats()
	assert.Equal(t, 2, stats.TotalCreated, "should have created 2 connections")
}

func TestSSHConnectionPool_ReleaseAfterClose(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)

	ctx := context.Background()
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)

	// Close pool while connection is in use
	err = pool.Close()
	require.NoError(t, err)

	// Release should still work (connection gets closed)
	pool.Release(conn)

	stats := pool.Stats()
	assert.Equal(t, 0, stats.InUse, "connection should be removed from in use")
	assert.Equal(t, 0, stats.Available, "connection should not be added to available (pool closed)")
}

func TestSSHConnectionPool_AcquireBlocksUntilReleased(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    1,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// Acquire the only connection
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)

	// Start goroutine to release after delay
	go func() {
		time.Sleep(200 * time.Millisecond)
		pool.Release(conn)
	}()

	// Try to acquire - should block and then succeed
	start := time.Now()
	conn2, err := pool.Acquire(ctx)
	elapsed := time.Since(start)

	require.NoError(t, err, "should succeed after connection is released")
	assert.Equal(t, conn, conn2, "should get the same connection")
	assert.GreaterOrEqual(t, elapsed, 200*time.Millisecond, "should have waited for release")
	assert.Less(t, elapsed, 1*time.Second, "should not wait too long")
}

// =============================================================================
// Additional Edge Case Tests
// =============================================================================

func TestSSHConnectionPool_ConnectionFactoryCalledCorrectly(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    3,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	var callCount int32
	pool := createTestPoolWithFactory(config, countingConnectionFactory(&callCount))
	defer pool.Close()

	ctx := context.Background()

	// Acquire 3 connections (should create 3)
	var conns []*PooledConnection
	for i := 0; i < 3; i++ {
		c, err := pool.Acquire(ctx)
		require.NoError(t, err)
		conns = append(conns, c)
	}

	assert.Equal(t, int32(3), callCount, "factory should be called 3 times")

	// Release all
	for _, c := range conns {
		pool.Release(c)
	}

	// Acquire again - should reuse, not create new
	for i := 0; i < 3; i++ {
		c, err := pool.Acquire(ctx)
		require.NoError(t, err)
		pool.Release(c)
	}

	assert.Equal(t, int32(3), callCount, "factory should not be called again for reused connections")
}

func TestSSHConnectionPool_UseCountIncrementsOnReuse(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    1,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	var conn *PooledConnection
	var err error

	// Acquire and release 5 times
	for i := 1; i <= 5; i++ {
		conn, err = pool.Acquire(ctx)
		require.NoError(t, err)
		assert.Equal(t, i, conn.useCount, "useCount should increment on each acquisition")
		pool.Release(conn)
	}
}

func TestSSHConnectionPool_LastUsedUpdated(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    1,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// First acquire
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)
	firstAcquireTime := conn.lastUsed
	pool.Release(conn)

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Second acquire
	conn, err = pool.Acquire(ctx)
	require.NoError(t, err)
	secondAcquireTime := conn.lastUsed

	assert.True(t, secondAcquireTime.After(firstAcquireTime), "lastUsed should be updated on acquisition")
}

func TestSSHConnectionPool_WithoutIdleCleanupOption(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	// Create pool with WithoutIdleCleanup
	pool := NewSSHConnectionPoolWithOptions(
		nil,
		"",
		config,
		WithoutIdleCleanup(),
		WithConnectionFactory(mockConnectionFactory()),
	)
	defer pool.Close()

	// Verify pool works correctly without idle cleanup goroutine
	ctx := context.Background()
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)
	pool.Release(conn)

	stats := pool.Stats()
	assert.Equal(t, 1, stats.Available)
}

func TestSSHConnectionPool_WithConnectionFactoryOption(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	customPoolID := "custom-factory-conn"
	factory := func() (*PooledConnection, error) {
		return &PooledConnection{
			client:      nil,
			session:     nil,
			adminMode:   false,
			poolID:      customPoolID,
			lastUsed:    time.Now(),
			useCount:    0,
			initialized: false,
		}, nil
	}

	pool := NewSSHConnectionPoolWithOptions(
		nil,
		"",
		config,
		WithConnectionFactory(factory),
		WithoutIdleCleanup(),
	)
	defer pool.Close()

	ctx := context.Background()
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)

	// PoolID should be overwritten with sequential ID
	assert.Contains(t, conn.poolID, "ssh-conn-", "factory connection should get sequential poolID")
}

// =============================================================================
// Task 12: Statistics and Observability Tests
// =============================================================================

func TestSSHConnectionPool_TotalAcquisitions(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	pool := NewSSHConnectionPoolWithOptions(
		nil,
		"",
		config,
		WithoutIdleCleanup(),
		WithConnectionFactory(mockConnectionFactory()),
	)
	defer pool.Close()

	ctx := context.Background()

	// Initial state: no acquisitions
	stats := pool.Stats()
	assert.Equal(t, 0, stats.TotalAcquisitions, "initial acquisitions should be 0")

	// Acquire and release 5 times
	for i := 0; i < 5; i++ {
		conn, err := pool.Acquire(ctx)
		require.NoError(t, err)
		pool.Release(conn)
	}

	// Verify total acquisitions
	stats = pool.Stats()
	assert.Equal(t, 5, stats.TotalAcquisitions, "should track 5 acquisitions")
}

func TestSSHConnectionPool_WaitCount(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    1, // Only 1 connection to force waiting
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	pool := NewSSHConnectionPoolWithOptions(
		nil,
		"",
		config,
		WithoutIdleCleanup(),
		WithConnectionFactory(mockConnectionFactory()),
	)
	defer pool.Close()

	ctx := context.Background()

	// Initial state: no waits
	stats := pool.Stats()
	assert.Equal(t, 0, stats.WaitCount, "initial wait count should be 0")

	// Acquire the only connection
	conn1, err := pool.Acquire(ctx)
	require.NoError(t, err)

	// Start a goroutine that will wait for a connection
	var conn2 *PooledConnection
	var acquireErr error
	done := make(chan struct{})
	go func() {
		conn2, acquireErr = pool.Acquire(ctx)
		close(done)
	}()

	// Give the goroutine time to start waiting
	time.Sleep(150 * time.Millisecond)

	// Release the first connection to unblock the waiting goroutine
	pool.Release(conn1)

	// Wait for the second acquire to complete
	select {
	case <-done:
		require.NoError(t, acquireErr)
		pool.Release(conn2)
	case <-time.After(2 * time.Second):
		t.Fatal("second acquire should have completed")
	}

	// Verify wait count was incremented
	stats = pool.Stats()
	assert.GreaterOrEqual(t, stats.WaitCount, 1, "wait count should be at least 1")
}

func TestSSHConnectionPool_LogStats(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    3,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	pool := NewSSHConnectionPoolWithOptions(
		nil,
		"",
		config,
		WithoutIdleCleanup(),
		WithConnectionFactory(mockConnectionFactory()),
	)
	defer pool.Close()

	ctx := context.Background()

	// Acquire a few connections
	conn1, _ := pool.Acquire(ctx)
	conn2, _ := pool.Acquire(ctx)
	pool.Release(conn1)

	// LogStats should not panic and should log correctly
	// We can't easily verify log output in unit tests, but we can ensure no panic
	assert.NotPanics(t, func() {
		pool.LogStats()
	})

	// Verify stats are accurate
	stats := pool.Stats()
	assert.Equal(t, 2, stats.TotalCreated)
	assert.Equal(t, 1, stats.InUse)
	assert.Equal(t, 1, stats.Available)
	assert.Equal(t, 2, stats.TotalAcquisitions)

	pool.Release(conn2)
}

// =============================================================================
// Admin Mode Tests
// =============================================================================

func TestSSHConnectionPool_AdminModePersists(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    1,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// Acquire connection, set admin mode, release
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)
	assert.False(t, conn.adminMode, "new connection should not be in admin mode")
	conn.SetAdminMode(true)
	pool.Release(conn)

	// Acquire again - should get same connection with admin mode preserved
	conn2, err := pool.Acquire(ctx)
	require.NoError(t, err)
	assert.True(t, conn2.adminMode, "reused connection should preserve admin mode")
}
