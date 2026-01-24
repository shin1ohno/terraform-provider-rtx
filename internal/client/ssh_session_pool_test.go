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
	"golang.org/x/crypto/ssh"
)

// =============================================================================
// Helper functions and mock session factory
// =============================================================================

// mockSessionFactory creates a simple session factory for testing
func mockSessionFactory() SessionFactory {
	return func() (*PooledSSHSession, error) {
		return &PooledSSHSession{
			workingSession: nil, // nil is fine for tests that don't use actual SSH
			poolID:         "mock-session",
			lastUsed:       time.Now(),
			useCount:       0,
			initialized:    false,
		}, nil
	}
}

// errorSessionFactory creates a session factory that always returns an error
func errorSessionFactory(err error) SessionFactory {
	return func() (*PooledSSHSession, error) {
		return nil, err
	}
}

// countingSessionFactory creates a session factory that counts creations
func countingSessionFactory(counter *int32) SessionFactory {
	return func() (*PooledSSHSession, error) {
		atomic.AddInt32(counter, 1)
		return &PooledSSHSession{
			workingSession: nil,
			poolID:         "counting-session",
			lastUsed:       time.Now(),
			useCount:       0,
			initialized:    false,
		}, nil
	}
}

// createTestPool creates a pool with mock session factory and no idle cleanup
func createTestPool(config SSHPoolConfig) *SSHSessionPool {
	return NewSSHSessionPoolWithOptions(
		nil, // sshClient not needed with factory
		config,
		WithSessionFactory(mockSessionFactory()),
		WithoutIdleCleanup(),
	)
}

// createTestPoolWithFactory creates a pool with custom session factory
func createTestPoolWithFactory(config SSHPoolConfig, factory SessionFactory) *SSHSessionPool {
	return NewSSHSessionPoolWithOptions(
		nil,
		config,
		WithSessionFactory(factory),
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

func TestNewSSHSessionPool_DefaultConfig(t *testing.T) {
	config := DefaultSSHPoolConfig()
	pool := createTestPool(config)
	defer pool.Close()

	stats := pool.Stats()
	assert.Equal(t, 0, stats.TotalCreated, "new pool should have 0 sessions created")
	assert.Equal(t, 0, stats.InUse, "new pool should have 0 sessions in use")
	assert.Equal(t, 0, stats.Available, "new pool should have 0 available sessions")
	assert.Equal(t, config.MaxSessions, stats.MaxSessions, "MaxSessions should match config")
}

func TestNewSSHSessionPool_CustomConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      SSHPoolConfig
		wantMax     int
		wantIdle    time.Duration
		wantAcquire time.Duration
	}{
		{
			name: "single session pool",
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

func TestSSHSessionPool_Acquire_EmptyPool(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()
	session, err := pool.Acquire(ctx)

	require.NoError(t, err, "Acquire should succeed on empty pool")
	require.NotNil(t, session, "acquired session should not be nil")
	assert.Equal(t, 1, session.useCount, "first acquisition should set useCount to 1")
	assert.True(t, session.initialized, "session should be initialized")

	stats := pool.Stats()
	assert.Equal(t, 1, stats.TotalCreated, "should have created 1 session")
	assert.Equal(t, 1, stats.InUse, "should have 1 session in use")
	assert.Equal(t, 0, stats.Available, "should have 0 available sessions")
}

func TestSSHSessionPool_Acquire_ReusesAvailableSession(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// Acquire and release a session
	session1, err := pool.Acquire(ctx)
	require.NoError(t, err)
	pool.Release(session1)

	// Acquire again - should get the same session
	session2, err := pool.Acquire(ctx)
	require.NoError(t, err)

	assert.Equal(t, session1, session2, "should reuse the same session")
	assert.Equal(t, 2, session2.useCount, "reused session should have incremented useCount")

	stats := pool.Stats()
	assert.Equal(t, 1, stats.TotalCreated, "should only have created 1 session total")
}

func TestSSHSessionPool_Release_ReturnsToPool(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()
	session, err := pool.Acquire(ctx)
	require.NoError(t, err)

	stats := pool.Stats()
	assert.Equal(t, 1, stats.InUse)
	assert.Equal(t, 0, stats.Available)

	pool.Release(session)

	stats = pool.Stats()
	assert.Equal(t, 0, stats.InUse, "released session should not be in use")
	assert.Equal(t, 1, stats.Available, "released session should be available")
}

func TestSSHSessionPool_Close_ClosesAllSessions(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    3,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)

	ctx := context.Background()

	// Acquire and release multiple sessions
	var sessions []*PooledSSHSession
	for i := 0; i < 3; i++ {
		session, err := pool.Acquire(ctx)
		require.NoError(t, err)
		sessions = append(sessions, session)
	}

	// Release all
	for _, s := range sessions {
		pool.Release(s)
	}

	stats := pool.Stats()
	assert.Equal(t, 3, stats.Available)

	// Close the pool
	err := pool.Close()
	require.NoError(t, err)

	stats = pool.Stats()
	assert.Equal(t, 0, stats.Available, "close should clear available sessions")
}

func TestSSHSessionPool_Close_Idempotent(t *testing.T) {
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

func TestSSHSessionPool_Stats_ReturnsCorrectValues(t *testing.T) {
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

	// Acquire 3 sessions
	var sessions []*PooledSSHSession
	for i := 0; i < 3; i++ {
		s, _ := pool.Acquire(ctx)
		sessions = append(sessions, s)
	}

	stats = pool.Stats()
	assert.Equal(t, 3, stats.TotalCreated)
	assert.Equal(t, 3, stats.InUse)
	assert.Equal(t, 0, stats.Available)

	// Release 2 sessions
	pool.Release(sessions[0])
	pool.Release(sessions[1])

	stats = pool.Stats()
	assert.Equal(t, 3, stats.TotalCreated)
	assert.Equal(t, 1, stats.InUse)
	assert.Equal(t, 2, stats.Available)
}

func TestSSHSessionPool_DoubleRelease_HandledGracefully(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()
	session, err := pool.Acquire(ctx)
	require.NoError(t, err)

	// First release
	pool.Release(session)

	stats := pool.Stats()
	assert.Equal(t, 1, stats.Available)

	// Second release should be ignored
	pool.Release(session)

	stats = pool.Stats()
	assert.Equal(t, 1, stats.Available, "double release should not add duplicate session")
}

func TestSSHSessionPool_ReleaseUnknownSession_Ignored(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	// Create a session that was never acquired from this pool
	unknownSession := &PooledSSHSession{
		workingSession: nil,
		poolID:         "unknown-session",
		lastUsed:       time.Now(),
		useCount:       1,
		initialized:    true,
	}

	// Release should be handled gracefully (no panic)
	pool.Release(unknownSession)

	stats := pool.Stats()
	assert.Equal(t, 0, stats.Available, "unknown session should not be added to pool")
}

// =============================================================================
// Task 8: Concurrent Access Tests
// =============================================================================

func TestSSHSessionPool_ConcurrentAcquire(t *testing.T) {
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

			session, err := pool.Acquire(ctx)
			if err != nil {
				atomic.AddInt32(&errorCount, 1)
				return
			}

			atomic.AddInt32(&successCount, 1)

			// Simulate work
			time.Sleep(5 * time.Millisecond)

			pool.Release(session)
		}()
	}

	wg.Wait()

	assert.Equal(t, int32(numGoroutines), successCount, "all goroutines should succeed")
	assert.Equal(t, int32(0), errorCount, "no errors expected")

	stats := pool.Stats()
	assert.Equal(t, 0, stats.InUse, "all sessions should be released")
}

func TestSSHSessionPool_ConcurrentRelease(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	config := SSHPoolConfig{
		MaxSessions:    10,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 10 * time.Second,
	}

	mockClient := &ssh.Client{}

	pool := &SSHSessionPool{
		sshClient: mockClient,
		config:    config,
		available: make([]*PooledSSHSession, 0, config.MaxSessions),
		inUse:     make(map[*PooledSSHSession]bool),
	}
	pool.cond = sync.NewCond(&pool.mu)

	// Create sessions that are "in use"
	sessions := make([]*PooledSSHSession, config.MaxSessions)
	for i := 0; i < config.MaxSessions; i++ {
		session := &PooledSSHSession{
			workingSession: nil,
			poolID:         "release-test-" + string(rune('0'+i)),
			lastUsed:       time.Now(),
			useCount:       1,
			initialized:    true,
		}
		sessions[i] = session
		pool.inUse[session] = true
	}

	var wg sync.WaitGroup

	// Release all sessions concurrently
	for i := 0; i < config.MaxSessions; i++ {
		wg.Add(1)
		go func(session *PooledSSHSession) {
			defer wg.Done()
			pool.Release(session)
		}(sessions[i])
	}

	wg.Wait()

	stats := pool.Stats()
	assert.Equal(t, 0, stats.InUse, "all sessions should be released")
	assert.Equal(t, config.MaxSessions, stats.Available, "all sessions should be available")
}

func TestSSHSessionPool_MixedAcquireRelease(t *testing.T) {
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

			session, err := pool.Acquire(ctx)
			if err != nil {
				return
			}

			// Variable work duration
			time.Sleep(time.Duration(id%5) * time.Millisecond)

			pool.Release(session)
			atomic.AddInt32(&successCount, 1)
		}(i)
	}

	wg.Wait()

	t.Logf("Completed %d successful operations", successCount)
	assert.Greater(t, successCount, int32(0), "should have some successful operations")

	stats := pool.Stats()
	assert.Equal(t, 0, stats.InUse, "all sessions should be released")
}

func TestSSHSessionPool_RaceDetector(t *testing.T) {
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
			session, err := pool.Acquire(ctx)
			if err == nil && session != nil {
				pool.Release(session)
			}
		}
		done <- true
	}()

	// Goroutine 2: Continuous acquire/release
	go func() {
		ctx := context.Background()
		for i := 0; i < iterations; i++ {
			session, err := pool.Acquire(ctx)
			if err == nil && session != nil {
				pool.Release(session)
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

func TestSSHSessionPool_ConcurrentStatsAccess(t *testing.T) {
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

func TestSSHSessionPool_ConcurrentClose(t *testing.T) {
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
			session, err := pool.Acquire(ctx)
			if err == nil && session != nil {
				time.Sleep(50 * time.Millisecond)
				pool.Release(session)
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

func TestSSHSessionPool_HighContention(t *testing.T) {
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
				session, err := pool.Acquire(ctx)
				if err != nil {
					continue
				}

				time.Sleep(time.Millisecond)

				pool.Release(session)
				atomic.AddInt32(&totalOperations, 1)
			}
		}()
	}

	wg.Wait()

	total := atomic.LoadInt32(&totalOperations)
	t.Logf("Completed %d operations under high contention", total)
	assert.Greater(t, total, int32(0), "should complete some operations")

	stats := pool.Stats()
	assert.Equal(t, 0, stats.InUse, "all sessions should be released")
}

// =============================================================================
// Task 9: Timeout and Error Handling Tests
// =============================================================================

func TestSSHSessionPool_AcquireTimeout_PoolExhausted(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    1,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 200 * time.Millisecond, // Short timeout for test
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// Acquire the only available session
	session, err := pool.Acquire(ctx)
	require.NoError(t, err)

	// Try to acquire another - should timeout
	start := time.Now()
	_, err = pool.Acquire(ctx)
	elapsed := time.Since(start)

	assert.Error(t, err, "should timeout waiting for session")
	assert.Contains(t, err.Error(), "timeout", "error should mention timeout")
	assert.GreaterOrEqual(t, elapsed, 200*time.Millisecond, "should wait at least AcquireTimeout")
	assert.Less(t, elapsed, 500*time.Millisecond, "should not wait too long")

	// Cleanup
	pool.Release(session)
}

func TestSSHSessionPool_AcquireTimeout_WithContextDeadline(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    1,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 10 * time.Second, // Long default timeout
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// Acquire the only available session
	session, err := pool.Acquire(ctx)
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
	pool.Release(session)
}

func TestSSHSessionPool_ContextCancellation(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    1,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// Acquire the only available session
	session, err := pool.Acquire(ctx)
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
	pool.Release(session)
}

func TestSSHSessionPool_PoolClosedError(t *testing.T) {
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

func TestSSHSessionPool_SessionCreationFailure(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	creationError := errors.New("SSH session creation failed")
	pool := createTestPoolWithFactory(config, errorSessionFactory(creationError))
	defer pool.Close()

	ctx := context.Background()
	_, err := pool.Acquire(ctx)

	assert.Error(t, err)
	assert.ErrorIs(t, err, creationError, "should return the factory error")
}

func TestSSHSessionPool_SessionCreationFailure_CountedCorrectly(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	callCount := int32(0)
	failAfter := int32(2)

	factory := func() (*PooledSSHSession, error) {
		count := atomic.AddInt32(&callCount, 1)
		if count > failAfter {
			return nil, errors.New("max sessions reached simulation")
		}
		return &PooledSSHSession{
			workingSession: nil,
			poolID:         "test-session",
			lastUsed:       time.Now(),
			useCount:       0,
			initialized:    false,
		}, nil
	}

	pool := createTestPoolWithFactory(config, factory)
	defer pool.Close()

	ctx := context.Background()

	// First two should succeed
	session1, err1 := pool.Acquire(ctx)
	session2, err2 := pool.Acquire(ctx)
	require.NoError(t, err1)
	require.NoError(t, err2)

	pool.Release(session1)
	pool.Release(session2)

	stats := pool.Stats()
	assert.Equal(t, 2, stats.TotalCreated, "should have created 2 sessions")
}

func TestSSHSessionPool_ReleaseAfterClose(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)

	ctx := context.Background()
	session, err := pool.Acquire(ctx)
	require.NoError(t, err)

	// Close pool while session is in use
	err = pool.Close()
	require.NoError(t, err)

	// Release should still work (session gets closed)
	pool.Release(session)

	stats := pool.Stats()
	assert.Equal(t, 0, stats.InUse, "session should be removed from in use")
	assert.Equal(t, 0, stats.Available, "session should not be added to available (pool closed)")
}

func TestSSHSessionPool_AcquireBlocksUntilReleased(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    1,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// Acquire the only session
	session, err := pool.Acquire(ctx)
	require.NoError(t, err)

	// Start goroutine to release after delay
	go func() {
		time.Sleep(200 * time.Millisecond)
		pool.Release(session)
	}()

	// Try to acquire - should block and then succeed
	start := time.Now()
	session2, err := pool.Acquire(ctx)
	elapsed := time.Since(start)

	require.NoError(t, err, "should succeed after session is released")
	assert.Equal(t, session, session2, "should get the same session")
	assert.GreaterOrEqual(t, elapsed, 200*time.Millisecond, "should have waited for release")
	assert.Less(t, elapsed, 1*time.Second, "should not wait too long")
}

// =============================================================================
// Additional Edge Case Tests
// =============================================================================

func TestSSHSessionPool_SessionFactoryCalledCorrectly(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    3,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	var callCount int32
	pool := createTestPoolWithFactory(config, countingSessionFactory(&callCount))
	defer pool.Close()

	ctx := context.Background()

	// Acquire 3 sessions (should create 3)
	var sessions []*PooledSSHSession
	for i := 0; i < 3; i++ {
		s, err := pool.Acquire(ctx)
		require.NoError(t, err)
		sessions = append(sessions, s)
	}

	assert.Equal(t, int32(3), callCount, "factory should be called 3 times")

	// Release all
	for _, s := range sessions {
		pool.Release(s)
	}

	// Acquire again - should reuse, not create new
	for i := 0; i < 3; i++ {
		s, err := pool.Acquire(ctx)
		require.NoError(t, err)
		pool.Release(s)
	}

	assert.Equal(t, int32(3), callCount, "factory should not be called again for reused sessions")
}

func TestSSHSessionPool_UseCountIncrementsOnReuse(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    1,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	var session *PooledSSHSession
	var err error

	// Acquire and release 5 times
	for i := 1; i <= 5; i++ {
		session, err = pool.Acquire(ctx)
		require.NoError(t, err)
		assert.Equal(t, i, session.useCount, "useCount should increment on each acquisition")
		pool.Release(session)
	}
}

func TestSSHSessionPool_LastUsedUpdated(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    1,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}
	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// First acquire
	session, err := pool.Acquire(ctx)
	require.NoError(t, err)
	firstAcquireTime := session.lastUsed
	pool.Release(session)

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Second acquire
	session, err = pool.Acquire(ctx)
	require.NoError(t, err)
	secondAcquireTime := session.lastUsed

	assert.True(t, secondAcquireTime.After(firstAcquireTime), "lastUsed should be updated on acquisition")
}

func TestSSHSessionPool_WithoutIdleCleanupOption(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	// Create pool with WithoutIdleCleanup
	pool := NewSSHSessionPoolWithOptions(
		nil,
		config,
		WithoutIdleCleanup(),
		WithSessionFactory(mockSessionFactory()),
	)
	defer pool.Close()

	// Verify pool works correctly without idle cleanup goroutine
	ctx := context.Background()
	session, err := pool.Acquire(ctx)
	require.NoError(t, err)
	pool.Release(session)

	stats := pool.Stats()
	assert.Equal(t, 1, stats.Available)
}

func TestSSHSessionPool_WithSessionFactoryOption(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	customPoolID := "custom-factory-session"
	factory := func() (*PooledSSHSession, error) {
		return &PooledSSHSession{
			workingSession: nil,
			poolID:         customPoolID,
			lastUsed:       time.Now(),
			useCount:       0,
			initialized:    false,
		}, nil
	}

	pool := NewSSHSessionPoolWithOptions(
		nil,
		config,
		WithSessionFactory(factory),
		WithoutIdleCleanup(),
	)
	defer pool.Close()

	ctx := context.Background()
	session, err := pool.Acquire(ctx)
	require.NoError(t, err)

	// PoolID should be overwritten with sequential ID
	assert.Contains(t, session.poolID, "ssh-session-", "factory session should get sequential poolID")
}

// =============================================================================
// Task 12: Statistics and Observability Tests
// =============================================================================

func TestSSHSessionPool_TotalAcquisitions(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	pool := NewSSHSessionPoolWithOptions(
		nil,
		config,
		WithoutIdleCleanup(),
		WithSessionFactory(mockSessionFactory()),
	)
	defer pool.Close()

	ctx := context.Background()

	// Initial state: no acquisitions
	stats := pool.Stats()
	assert.Equal(t, 0, stats.TotalAcquisitions, "initial acquisitions should be 0")

	// Acquire and release 5 times
	for i := 0; i < 5; i++ {
		session, err := pool.Acquire(ctx)
		require.NoError(t, err)
		pool.Release(session)
	}

	// Verify total acquisitions
	stats = pool.Stats()
	assert.Equal(t, 5, stats.TotalAcquisitions, "should track 5 acquisitions")
}

func TestSSHSessionPool_WaitCount(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    1, // Only 1 session to force waiting
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	pool := NewSSHSessionPoolWithOptions(
		nil,
		config,
		WithoutIdleCleanup(),
		WithSessionFactory(mockSessionFactory()),
	)
	defer pool.Close()

	ctx := context.Background()

	// Initial state: no waits
	stats := pool.Stats()
	assert.Equal(t, 0, stats.WaitCount, "initial wait count should be 0")

	// Acquire the only session
	session1, err := pool.Acquire(ctx)
	require.NoError(t, err)

	// Start a goroutine that will wait for a session
	var session2 *PooledSSHSession
	var acquireErr error
	done := make(chan struct{})
	go func() {
		session2, acquireErr = pool.Acquire(ctx)
		close(done)
	}()

	// Give the goroutine time to start waiting
	time.Sleep(150 * time.Millisecond)

	// Release the first session to unblock the waiting goroutine
	pool.Release(session1)

	// Wait for the second acquire to complete
	select {
	case <-done:
		require.NoError(t, acquireErr)
		pool.Release(session2)
	case <-time.After(2 * time.Second):
		t.Fatal("second acquire should have completed")
	}

	// Verify wait count was incremented
	stats = pool.Stats()
	assert.GreaterOrEqual(t, stats.WaitCount, 1, "wait count should be at least 1")
}

func TestSSHSessionPool_LogStats(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    3,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	pool := NewSSHSessionPoolWithOptions(
		nil,
		config,
		WithoutIdleCleanup(),
		WithSessionFactory(mockSessionFactory()),
	)
	defer pool.Close()

	ctx := context.Background()

	// Acquire a few sessions
	session1, _ := pool.Acquire(ctx)
	session2, _ := pool.Acquire(ctx)
	pool.Release(session1)

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

	pool.Release(session2)
}
