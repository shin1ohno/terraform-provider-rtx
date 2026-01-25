package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sh1/terraform-provider-rtx/internal/logging"
)

const (
	// maxRetries is the maximum number of retry attempts for failed commands
	maxRetries = 2
	// retryBaseDelay is the base delay between retries
	retryBaseDelay = 100 * time.Millisecond
)

// PooledExecutor executes commands using connections from the SSH connection pool
type PooledExecutor struct {
	pool           *SSHConnectionPool
	promptDetector PromptDetector
	config         *Config
}

// NewPooledExecutor creates a new pooled executor
func NewPooledExecutor(pool *SSHConnectionPool, promptDetector PromptDetector, config *Config) Executor {
	return &PooledExecutor{
		pool:           pool,
		promptDetector: promptDetector,
		config:         config,
	}
}

// Run executes a command using a session from the pool with retry logic
func (e *PooledExecutor) Run(ctx context.Context, cmd string) ([]byte, error) {
	logger := logging.FromContext(ctx)

	// Log command with resource context if available
	logEvent := logger.Info().Str("command", logging.SanitizeString(cmd))
	if res := logging.ResourceFromContext(ctx); res != nil {
		logEvent = logEvent.Str("resource", res.Type)
		if res.ID != "" {
			logEvent = logEvent.Str("id", res.ID)
		}
	}
	logEvent.Msg("RTX command (pooled)")

	return e.executeWithRetry(ctx, cmd, maxRetries)
}

// executeWithRetry executes a command with retry logic on connection failure
func (e *PooledExecutor) executeWithRetry(ctx context.Context, cmd string, retries int) ([]byte, error) {
	logger := logging.FromContext(ctx)
	var lastErr error

	for attempt := 0; attempt <= retries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Acquire connection from pool
		conn, err := e.pool.Acquire(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to acquire SSH connection: %w", err)
		}

		// Prepare connection (admin authentication if needed)
		needsAdmin := e.requiresAdminPrivileges(cmd)
		if err := e.prepareConnection(ctx, conn, needsAdmin); err != nil {
			logger.Warn().
				Err(err).
				Int("attempt", attempt+1).
				Msg("PooledExecutor: Failed to prepare connection, discarding")
			e.pool.Discard(conn)
			lastErr = fmt.Errorf("failed to prepare connection: %w", err)
			if attempt < retries {
				time.Sleep(retryBaseDelay * time.Duration(attempt+1))
			}
			continue
		}

		// Execute command on connection
		output, err := e.executeOnConnection(ctx, conn, cmd)
		if err != nil {
			logger.Warn().
				Err(err).
				Int("attempt", attempt+1).
				Int("max_retries", retries).
				Msg("PooledExecutor: Command execution failed, discarding connection")
			e.pool.Discard(conn)
			lastErr = err
			if attempt < retries {
				time.Sleep(retryBaseDelay * time.Duration(attempt+1))
			}
			continue
		}

		// Success - release connection back to pool
		e.pool.Release(conn)
		return output, nil
	}

	return nil, fmt.Errorf("command failed after %d attempts: %w", retries+1, lastErr)
}

// executeOnConnection executes a command on the given connection
func (e *PooledExecutor) executeOnConnection(ctx context.Context, conn *PooledConnection, cmd string) ([]byte, error) {
	logger := logging.FromContext(ctx)

	// Execute the command
	output, err := conn.Send(cmd)
	if err != nil {
		return nil, fmt.Errorf("command execution failed: %w", err)
	}

	// Check for prompt
	matched, prompt := e.promptDetector.DetectPrompt(output)
	if !matched {
		logger.Debug().Str("output", string(output)).Msg("PooledExecutor: Prompt detection failed")
		return nil, fmt.Errorf("%w: output does not contain expected prompt", ErrPrompt)
	}
	logger.Debug().Str("prompt", prompt).Msg("PooledExecutor: Prompt detected")

	return output, nil
}

// requiresAdminPrivileges checks if a command requires administrator privileges.
// If admin password is configured, always use administrator mode since RTX routers
// provide more complete information in administrator mode.
func (e *PooledExecutor) requiresAdminPrivileges(cmd string) bool {
	hasConfig := e.config != nil
	hasPassword := hasConfig && e.config.AdminPassword != ""
	return hasPassword
}

// prepareConnection prepares a connection for command execution, including admin authentication if needed
func (e *PooledExecutor) prepareConnection(ctx context.Context, conn *PooledConnection, needsAdmin bool) error {
	if !needsAdmin {
		return nil
	}

	// Check if connection is already in admin mode
	if conn.adminMode {
		logging.FromContext(ctx).Debug().
			Str("pool_id", conn.poolID).
			Msg("PooledExecutor: Connection already in admin mode")
		return nil
	}

	// Authenticate as administrator
	if err := e.authenticateAsAdmin(ctx, conn); err != nil {
		return err
	}

	conn.SetAdminMode(true)
	return nil
}

// authenticateAsAdmin authenticates as administrator on the given connection
func (e *PooledExecutor) authenticateAsAdmin(ctx context.Context, conn *PooledConnection) error {
	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("pool_id", conn.poolID).
		Msg("PooledExecutor: Authenticating as administrator")

	ws := conn.session
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return fmt.Errorf("session is closed")
	}

	// Send administrator command
	if _, err := fmt.Fprintf(ws.stdin, "administrator\r"); err != nil {
		return fmt.Errorf("failed to send administrator command: %w", err)
	}

	// Read until we get password prompt
	_, err := ws.readUntilString("Password:", 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get password prompt: %w", err)
	}
	logger.Debug().Msg("PooledExecutor: Password prompt received")

	// Send password
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", e.config.AdminPassword); err != nil {
		return fmt.Errorf("failed to send password: %w", err)
	}

	// Read response after password - look for administrator prompt (# instead of >)
	response, err := ws.readUntilPrompt(10 * time.Second)
	if err != nil {
		return fmt.Errorf("failed to read password response: %w", err)
	}

	responseStr := string(response)
	logger.Debug().Str("response", responseStr).Msg("PooledExecutor: Password authentication response received")

	// Check for authentication failure
	if strings.Contains(responseStr, "incorrect") || strings.Contains(responseStr, "failed") || strings.Contains(responseStr, "Invalid") {
		return fmt.Errorf("administrator authentication failed: %s", responseStr)
	}

	// Verify we actually got the admin prompt (#) not user prompt (>)
	if !strings.Contains(responseStr, "#") {
		return fmt.Errorf("administrator authentication failed: did not get admin prompt (#), got: %s", responseStr)
	}

	logger.Debug().Msg("PooledExecutor: Administrator authentication successful")
	return nil
}

// RunBatch executes multiple commands using a single connection and returns the combined output
func (e *PooledExecutor) RunBatch(ctx context.Context, cmds []string) ([]byte, error) {
	logger := logging.FromContext(ctx)

	if len(cmds) == 0 {
		return nil, nil
	}

	// Acquire connection once for all commands
	conn, err := e.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire SSH connection: %w", err)
	}

	// Determine if any command needs admin privileges
	needsAdmin := false
	for _, cmd := range cmds {
		if e.requiresAdminPrivileges(cmd) {
			needsAdmin = true
			break
		}
	}

	// Prepare connection once
	if err := e.prepareConnection(ctx, conn, needsAdmin); err != nil {
		e.pool.Discard(conn)
		return nil, fmt.Errorf("failed to prepare connection for batch: %w", err)
	}

	var allOutput []byte
	for _, cmd := range cmds {
		logger.Info().Str("command", logging.SanitizeString(cmd)).Msg("RTX batch command (pooled)")

		output, err := e.executeOnConnection(ctx, conn, cmd)
		if err != nil {
			// On failure, discard connection and return partial output
			e.pool.Discard(conn)
			return allOutput, fmt.Errorf("batch command '%s' failed: %w", cmd, err)
		}
		allOutput = append(allOutput, output...)
	}

	// Release connection after all commands complete
	e.pool.Release(conn)
	return allOutput, nil
}

// SetAdministratorPassword sets the administrator password using interactive prompts
// RTX prompts: Old_Password: -> New_Password: -> New_Password: -> Password Strength: ...
func (e *PooledExecutor) SetAdministratorPassword(ctx context.Context, oldPassword, newPassword string) error {
	logger := logging.FromContext(ctx)
	logger.Debug().Msg("PooledExecutor: Setting administrator password")

	// Acquire connection from pool
	conn, err := e.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire SSH connection: %w", err)
	}

	// Authenticate as administrator first (required for password commands)
	if e.config != nil && e.config.AdminPassword != "" {
		if err := e.authenticateAsAdmin(ctx, conn); err != nil {
			e.pool.Discard(conn)
			return fmt.Errorf("failed to authenticate as administrator: %w", err)
		}
		conn.SetAdminMode(true)
	}

	ws := conn.session
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Send administrator password command
	if _, err := fmt.Fprintf(ws.stdin, "administrator password\r"); err != nil {
		e.pool.Discard(conn)
		return fmt.Errorf("failed to send administrator password command: %w", err)
	}

	// Wait for Old_Password: prompt
	_, err = ws.readUntilString("Old_Password:", 10*time.Second)
	if err != nil {
		e.pool.Discard(conn)
		return fmt.Errorf("failed to get Old_Password prompt: %w", err)
	}
	logger.Debug().Msg("PooledExecutor: Old_Password prompt received")

	// Send old password
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", oldPassword); err != nil {
		e.pool.Discard(conn)
		return fmt.Errorf("failed to send old password: %w", err)
	}

	// Wait for first New_Password: prompt
	_, err = ws.readUntilString("New_Password:", 10*time.Second)
	if err != nil {
		e.pool.Discard(conn)
		return fmt.Errorf("failed to get first New_Password prompt: %w", err)
	}
	logger.Debug().Msg("PooledExecutor: First New_Password prompt received")

	// Send new password
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", newPassword); err != nil {
		e.pool.Discard(conn)
		return fmt.Errorf("failed to send new password: %w", err)
	}

	// Wait for second New_Password: prompt (confirmation)
	_, err = ws.readUntilString("New_Password:", 10*time.Second)
	if err != nil {
		e.pool.Discard(conn)
		return fmt.Errorf("failed to get second New_Password prompt: %w", err)
	}
	logger.Debug().Msg("PooledExecutor: Second New_Password prompt received")

	// Send new password again for confirmation
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", newPassword); err != nil {
		e.pool.Discard(conn)
		return fmt.Errorf("failed to send password confirmation: %w", err)
	}

	// Wait for completion (Password Strength or prompt)
	response, err := ws.readUntilPrompt(10 * time.Second)
	if err != nil {
		e.pool.Discard(conn)
		return fmt.Errorf("failed to read password change response: %w", err)
	}

	responseStr := string(response)
	logger.Debug().Str("response", responseStr).Msg("PooledExecutor: Password change response received")

	// Check for errors
	if strings.Contains(responseStr, "incorrect") || strings.Contains(responseStr, "failed") || strings.Contains(responseStr, "Invalid") {
		e.pool.Discard(conn)
		return fmt.Errorf("administrator password change failed: %s", responseStr)
	}

	// Release connection back to pool (note: connection state may have changed)
	e.pool.Release(conn)
	logger.Debug().Msg("PooledExecutor: Administrator password changed successfully")
	return nil
}

// SetLoginPassword sets the login password using interactive prompts
// RTX prompts: Old_Password: (if exists) -> New_Password: -> New_Password: -> Password Strength: ...
func (e *PooledExecutor) SetLoginPassword(ctx context.Context, newPassword string) error {
	logger := logging.FromContext(ctx)
	logger.Debug().Msg("PooledExecutor: Setting login password")

	// Acquire connection from pool
	conn, err := e.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire SSH connection: %w", err)
	}

	// Authenticate as administrator first (required for password commands)
	if e.config != nil && e.config.AdminPassword != "" {
		if err := e.authenticateAsAdmin(ctx, conn); err != nil {
			e.pool.Discard(conn)
			return fmt.Errorf("failed to authenticate as administrator: %w", err)
		}
		conn.SetAdminMode(true)
	}

	ws := conn.session
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Send login password command
	if _, err := fmt.Fprintf(ws.stdin, "login password\r"); err != nil {
		e.pool.Discard(conn)
		return fmt.Errorf("failed to send login password command: %w", err)
	}

	// Wait for New_Password: prompt (login password may not have old password prompt if not set)
	_, err = ws.readUntilString("New_Password:", 10*time.Second)
	if err != nil {
		e.pool.Discard(conn)
		return fmt.Errorf("failed to get first New_Password prompt: %w", err)
	}
	logger.Debug().Msg("PooledExecutor: First New_Password prompt received")

	// Send new password
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", newPassword); err != nil {
		e.pool.Discard(conn)
		return fmt.Errorf("failed to send new password: %w", err)
	}

	// Wait for second New_Password: prompt (confirmation)
	_, err = ws.readUntilString("New_Password:", 10*time.Second)
	if err != nil {
		e.pool.Discard(conn)
		return fmt.Errorf("failed to get second New_Password prompt: %w", err)
	}
	logger.Debug().Msg("PooledExecutor: Second New_Password prompt received")

	// Send new password again for confirmation
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", newPassword); err != nil {
		e.pool.Discard(conn)
		return fmt.Errorf("failed to send password confirmation: %w", err)
	}

	// Wait for completion (Password Strength or prompt)
	response, err := ws.readUntilPrompt(10 * time.Second)
	if err != nil {
		e.pool.Discard(conn)
		return fmt.Errorf("failed to read password change response: %w", err)
	}

	responseStr := string(response)
	logger.Debug().Str("response", responseStr).Msg("PooledExecutor: Password change response received")

	// Check for errors
	if strings.Contains(responseStr, "incorrect") || strings.Contains(responseStr, "failed") || strings.Contains(responseStr, "Invalid") {
		e.pool.Discard(conn)
		return fmt.Errorf("login password change failed: %s", responseStr)
	}

	// Release connection back to pool
	e.pool.Release(conn)
	logger.Debug().Msg("PooledExecutor: Login password changed successfully")
	return nil
}
