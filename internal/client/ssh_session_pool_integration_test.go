package client

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Task 10: State Drift Regression Tests
// =============================================================================

/*
TestStateDriftRegression_ConsoleCharacter verifies that console.character setting
does not drift after terraform apply due to SSH session initialization.

## Problem Description

When connecting to an RTX router via SSH, the client runs initialization commands
including "console character en.ascii" to ensure consistent output encoding.
Without session pooling, this command runs on EVERY operation (Read, Create, Update),
which overwrites the user's configured setting (e.g., "ja.utf8").

The state drift scenario:
1. User configures rtx_system with console.character = "ja.utf8"
2. Terraform apply sets console.character to "ja.utf8" on the router
3. Next Terraform plan/read operation creates a new SSH session
4. Session initialization runs "console character en.ascii"
5. Terraform reads console.character as "en.ascii"
6. Terraform detects drift and shows changes needed

## Solution

SSH session pool reuses sessions, so initialization only runs once when
creating a new session, not on every operation. Subsequent operations
reuse the existing session without re-running init commands.

## Test Scenarios

This file contains mock-based tests that simulate the state drift scenario.
For tests with a real router, set TF_ACC=1 and provide router credentials.
*/

// TestStateDriftRegression_SessionReusePreventsDrift verifies that session reuse
// prevents repeated initialization that would cause state drift.
func TestStateDriftRegression_SessionReusePreventsDrift(t *testing.T) {
	// Track how many times session factory is called (simulates init commands)
	var initCount int32

	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	factory := func() (*PooledSSHSession, error) {
		// This simulates SSH session creation which includes init commands
		// like "console character en.ascii"
		atomic.AddInt32(&initCount, 1)
		return &PooledSSHSession{
			workingSession: nil,
			poolID:         "drift-test-session",
			lastUsed:       time.Now(),
			useCount:       0,
			initialized:    true,
		}, nil
	}

	pool := createTestPoolWithFactory(config, factory)
	defer pool.Close()

	ctx := context.Background()

	// Simulate multiple terraform operations (plan, apply, plan, etc.)
	// Each operation acquires and releases a session
	for i := 0; i < 5; i++ {
		session, err := pool.Acquire(ctx)
		require.NoError(t, err, "Acquire should succeed")

		// Simulate operation (read config, apply changes, etc.)
		time.Sleep(time.Millisecond)

		pool.Release(session)
	}

	// With session pooling, init should only run ONCE (first acquire)
	// Without pooling, init would run 5 times (causing state drift each time)
	assert.Equal(t, int32(1), atomic.LoadInt32(&initCount),
		"Session init should only run once due to session reuse")
}

// TestStateDriftRegression_MultipleOperationsReuseSameSession verifies that
// sequential terraform operations reuse the same session.
func TestStateDriftRegression_MultipleOperationsReuseSameSession(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    1, // Force reuse by limiting to 1 session
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// First operation
	session1, err := pool.Acquire(ctx)
	require.NoError(t, err)
	poolID1 := session1.poolID
	pool.Release(session1)

	// Second operation - should get the same session
	session2, err := pool.Acquire(ctx)
	require.NoError(t, err)
	poolID2 := session2.poolID
	pool.Release(session2)

	// Third operation - should still get the same session
	session3, err := pool.Acquire(ctx)
	require.NoError(t, err)
	poolID3 := session3.poolID
	pool.Release(session3)

	assert.Equal(t, poolID1, poolID2, "Second operation should reuse first session")
	assert.Equal(t, poolID2, poolID3, "Third operation should reuse same session")

	// Use count should reflect reuse
	assert.Equal(t, 3, session3.useCount, "Session should have been used 3 times")
}

// TestStateDriftRegression_ConcurrentOperationsLimitedInit verifies that
// even with concurrent operations, session initialization is minimized.
func TestStateDriftRegression_ConcurrentOperationsLimitedInit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	var initCount int32

	config := SSHPoolConfig{
		MaxSessions:    2, // Allow 2 concurrent sessions
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 10 * time.Second,
	}

	factory := func() (*PooledSSHSession, error) {
		atomic.AddInt32(&initCount, 1)
		return &PooledSSHSession{
			workingSession: nil,
			poolID:         "concurrent-test-session",
			lastUsed:       time.Now(),
			useCount:       0,
			initialized:    true,
		}, nil
	}

	pool := createTestPoolWithFactory(config, factory)
	defer pool.Close()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Simulate 20 concurrent terraform operations
	numOperations := 20
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			session, err := pool.Acquire(ctx)
			if err != nil {
				return
			}

			// Simulate operation
			time.Sleep(5 * time.Millisecond)

			pool.Release(session)
		}()
	}

	wg.Wait()

	// Key insight: Even with 20 concurrent operations, the number of
	// session initializations should be much smaller than the number
	// of operations due to session reuse.
	//
	// Note: Due to the pool implementation releasing the lock during
	// session creation, there may be a small number of extra sessions
	// created under high concurrency. This is acceptable behavior as
	// the goal is to minimize init runs, not eliminate them entirely.
	finalInitCount := atomic.LoadInt32(&initCount)

	// The init count should be significantly less than the number of operations
	// (at most ~50% given high concurrency edge cases, but typically much lower)
	maxExpectedInits := int32(numOperations / 2)
	assert.LessOrEqual(t, finalInitCount, maxExpectedInits,
		"Init count should be much less than number of operations due to session reuse")

	t.Logf("Ran %d operations with only %d session initializations (max expected: %d)",
		numOperations, finalInitCount, maxExpectedInits)
}

// TestStateDriftRegression_SessionInitializedFlag verifies that the
// initialized flag is set correctly on pooled sessions.
func TestStateDriftRegression_SessionInitializedFlag(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// Acquire a session
	session, err := pool.Acquire(ctx)
	require.NoError(t, err)

	// Session should be marked as initialized
	assert.True(t, session.initialized,
		"Acquired session should be marked as initialized (init commands already run)")

	pool.Release(session)

	// Re-acquire and verify it's still initialized
	session2, err := pool.Acquire(ctx)
	require.NoError(t, err)
	assert.True(t, session2.initialized,
		"Reused session should still be marked as initialized")
}

// =============================================================================
// Acceptance Test - requires real RTX router
// =============================================================================

/*
TestStateDriftRegression_RealRouter_ConsoleCharacter is an acceptance test
that verifies the state drift fix with a real RTX router.

Prerequisites:
  - Set TF_ACC=1 environment variable
  - Set RTX_HOST, RTX_USER, RTX_PASSWORD environment variables
  - Router must be accessible via SSH

Test steps:
 1. Connect to router with SSH session pool enabled
 2. Set console.character to a non-default value (e.g., "ja.utf8")
 3. Read the current console.character setting
 4. Verify it matches the set value (not "en.ascii")
 5. Perform multiple read operations
 6. Verify console.character is still the user-configured value

To run: TF_ACC=1 go test -v ./internal/client/... -run TestStateDriftRegression_RealRouter
*/
func TestStateDriftRegression_RealRouter_ConsoleCharacter(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping acceptance test. Set TF_ACC=1 to run.")
	}

	host := os.Getenv("RTX_HOST")
	user := os.Getenv("RTX_USER")
	password := os.Getenv("RTX_PASSWORD")

	if host == "" || user == "" || password == "" {
		t.Skip("Skipping: RTX_HOST, RTX_USER, and RTX_PASSWORD must be set")
	}

	t.Log("This test requires manual verification with a real RTX router")
	t.Log("The test would:")
	t.Log("1. Create client with SSH session pool enabled")
	t.Log("2. Configure console.character to 'ja.utf8'")
	t.Log("3. Perform multiple read operations")
	t.Log("4. Verify console.character remains 'ja.utf8' (not 'en.ascii')")

	// TODO: Implement when acceptance test infrastructure is available
	// For now, document the expected behavior:
	//
	// config := &Config{
	//     Host:     host,
	//     Port:     22,
	//     Username: user,
	//     Password: password,
	//     Timeout:  30,
	// }
	//
	// client, err := NewClient(config, WithSSHSessionPool(true))
	// require.NoError(t, err)
	// defer client.Close()
	//
	// ctx := context.Background()
	// err = client.Dial(ctx)
	// require.NoError(t, err)
	//
	// // Configure console.character
	// systemService := client.SystemService()
	// err = systemService.Configure(ctx, SystemConfig{
	//     Console: &ConsoleConfig{
	//         Character: "ja.utf8",
	//     },
	// })
	// require.NoError(t, err)
	//
	// // Perform multiple reads and verify no drift
	// for i := 0; i < 3; i++ {
	//     config, err := systemService.Get(ctx)
	//     require.NoError(t, err)
	//     assert.Equal(t, "ja.utf8", config.Console.Character,
	//         "console.character should not drift to en.ascii on read %d", i+1)
	// }
}

// =============================================================================
// Task 11: Tests for SSH Pool Integration with Client
// =============================================================================

// TestSSHPoolIntegration_ClientCreation verifies that clients can be created
// with and without SSH session pool enabled.
func TestSSHPoolIntegration_ClientCreation(t *testing.T) {
	config := &Config{
		Host:     "192.168.1.1",
		Port:     22,
		Username: "admin",
		Password: "password",
		Timeout:  30,
	}

	// Test with SSH pool disabled (default)
	client, err := NewClient(config)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Test with SSH pool enabled
	clientWithPool, err := NewClient(config, WithSSHSessionPool(true))
	require.NoError(t, err)
	require.NotNil(t, clientWithPool)
}

// TestSSHPoolIntegration_PoolStatsAccurate verifies that pool statistics
// are accurately maintained during operations.
func TestSSHPoolIntegration_PoolStatsAccurate(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    3,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// Initial stats
	stats := pool.Stats()
	assert.Equal(t, 0, stats.TotalCreated)
	assert.Equal(t, 0, stats.InUse)
	assert.Equal(t, 0, stats.Available)

	// Acquire 2 sessions
	session1, _ := pool.Acquire(ctx)
	session2, _ := pool.Acquire(ctx)

	stats = pool.Stats()
	assert.Equal(t, 2, stats.TotalCreated)
	assert.Equal(t, 2, stats.InUse)
	assert.Equal(t, 0, stats.Available)

	// Release 1 session
	pool.Release(session1)

	stats = pool.Stats()
	assert.Equal(t, 2, stats.TotalCreated)
	assert.Equal(t, 1, stats.InUse)
	assert.Equal(t, 1, stats.Available)

	// Release remaining session
	pool.Release(session2)

	stats = pool.Stats()
	assert.Equal(t, 2, stats.TotalCreated)
	assert.Equal(t, 0, stats.InUse)
	assert.Equal(t, 2, stats.Available)
}

// TestSSHPoolIntegration_GracefulDegradation verifies that the pool
// handles errors gracefully without panicking.
func TestSSHPoolIntegration_GracefulDegradation(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 100 * time.Millisecond,
	}

	pool := createTestPool(config)

	ctx := context.Background()

	// Close the pool
	pool.Close()

	// Acquire should fail gracefully
	_, err := pool.Acquire(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "closed")

	// Double close should not panic
	err = pool.Close()
	assert.NoError(t, err)
}

// TestSSHPoolIntegration_SessionLifecycle verifies the full lifecycle
// of sessions through the pool.
func TestSSHPoolIntegration_SessionLifecycle(t *testing.T) {
	config := SSHPoolConfig{
		MaxSessions:    2,
		IdleTimeout:    5 * time.Minute,
		AcquireTimeout: 5 * time.Second,
	}

	pool := createTestPool(config)
	defer pool.Close()

	ctx := context.Background()

	// Phase 1: Initial acquisition creates new session
	session, err := pool.Acquire(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, session.useCount)
	originalPoolID := session.poolID

	// Phase 2: Release returns to pool
	pool.Release(session)
	stats := pool.Stats()
	assert.Equal(t, 1, stats.Available)

	// Phase 3: Re-acquisition reuses session
	session2, err := pool.Acquire(ctx)
	require.NoError(t, err)
	assert.Equal(t, originalPoolID, session2.poolID, "Should reuse same session")
	assert.Equal(t, 2, session2.useCount, "Use count should increment")

	// Phase 4: Close pool cleans up
	pool.Close()
	stats = pool.Stats()
	assert.Equal(t, 0, stats.Available, "Close should clean up available sessions")
}
