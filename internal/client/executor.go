package client

import (
	"context"
	"fmt"
	"time"

	"github.com/sh1/terraform-provider-rtx/internal/logging"
)

// Executor interface defines how commands are executed
type Executor interface {
	Run(ctx context.Context, cmd string) ([]byte, error)
	RunBatch(ctx context.Context, cmds []string) ([]byte, error)
	SetLoginPassword(ctx context.Context, newPassword string) error
	SetAdministratorPassword(ctx context.Context, oldPassword, newPassword string) error
}

// sshExecutor implements Executor using SSH session
type sshExecutor struct {
	session        Session
	promptDetector PromptDetector
	retryStrategy  RetryStrategy
}

// NewSSHExecutor creates a new SSH executor
func NewSSHExecutor(session Session, promptDetector PromptDetector, retryStrategy RetryStrategy) Executor {
	return &sshExecutor{
		session:        session,
		promptDetector: promptDetector,
		retryStrategy:  retryStrategy,
	}
}

// Run executes a command via SSH and returns the raw output
func (e *sshExecutor) Run(ctx context.Context, cmd string) ([]byte, error) {
	var raw []byte
	var err error

	for retry := 0; ; retry++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrTimeout, ctx.Err())
		default:
		}

		// Execute Send operation with context timeout handling
		type sendResult struct {
			data []byte
			err  error
		}

		sendCh := make(chan sendResult, 1)
		go func() {
			data, sendErr := e.session.Send(cmd)
			sendCh <- sendResult{data: data, err: sendErr}
		}()

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrTimeout, ctx.Err())
		case result := <-sendCh:
			raw, err = result.data, result.err
		}

		if err == nil {
			break
		}

		// Check if we should retry
		delay, giveUp := e.retryStrategy.Next(retry)
		if giveUp {
			return nil, fmt.Errorf("command execution failed: %w", err)
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %v", ErrTimeout, ctx.Err())
		case <-time.After(delay):
			// Continue to retry
		}
	}

	// Check for prompt
	matched, prompt := e.promptDetector.DetectPrompt(raw)
	if !matched {
		logging.FromContext(ctx).Debug().Str("component", "executor").Msgf("Prompt detection failed. Output: %q", string(raw))
		return nil, fmt.Errorf("%w: output does not contain expected prompt", ErrPrompt)
	}
	logging.FromContext(ctx).Debug().Str("component", "executor").Msgf("Prompt detected: %q", prompt)

	return raw, nil
}

// RunBatch executes multiple commands via SSH and returns the combined output
func (e *sshExecutor) RunBatch(ctx context.Context, cmds []string) ([]byte, error) {
	var allOutput []byte

	for _, cmd := range cmds {
		output, err := e.Run(ctx, cmd)
		if err != nil {
			return allOutput, fmt.Errorf("batch command '%s' failed: %w", cmd, err)
		}
		allOutput = append(allOutput, output...)
	}

	return allOutput, nil
}

// SetLoginPassword sets the login password via interactive prompt
func (e *sshExecutor) SetLoginPassword(ctx context.Context, newPassword string) error {
	logger := logging.FromContext(ctx)
	logger.Debug().Str("component", "executor").Msg("Setting login password")

	// Cast session to workingSession to access low-level methods
	ws, ok := e.session.(*workingSession)
	if !ok {
		return fmt.Errorf("session type not supported for login password setting")
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return fmt.Errorf("session is closed")
	}

	// Send login password command
	if _, err := fmt.Fprintf(ws.stdin, "login password\r"); err != nil {
		return fmt.Errorf("failed to send login password command: %w", err)
	}

	// Read until we get password prompt
	_, err := ws.readUntilString("Password:", 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get password prompt: %w", err)
	}

	// Send new password
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", newPassword); err != nil {
		return fmt.Errorf("failed to send new password: %w", err)
	}

	// Read response after password
	_, err = ws.readUntilPrompt(10 * time.Second)
	if err != nil {
		return fmt.Errorf("failed to read password response: %w", err)
	}

	logger.Debug().Str("component", "executor").Msg("Login password set successfully")
	return nil
}

// SetAdministratorPassword sets the administrator password via interactive prompt
func (e *sshExecutor) SetAdministratorPassword(ctx context.Context, oldPassword, newPassword string) error {
	logger := logging.FromContext(ctx)
	logger.Debug().Str("component", "executor").Msg("Setting administrator password")

	// Cast session to workingSession to access low-level methods
	ws, ok := e.session.(*workingSession)
	if !ok {
		return fmt.Errorf("session type not supported for administrator password setting")
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return fmt.Errorf("session is closed")
	}

	// Send administrator password command
	if _, err := fmt.Fprintf(ws.stdin, "administrator password\r"); err != nil {
		return fmt.Errorf("failed to send administrator password command: %w", err)
	}

	// Read until we get old password prompt
	_, err := ws.readUntilString("Old_Password:", 10*time.Second)
	if err != nil {
		// Some versions may use different prompt text
		_, err = ws.readUntilString("Password:", 10*time.Second)
		if err != nil {
			return fmt.Errorf("failed to get old password prompt: %w", err)
		}
	}

	// Send old password
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", oldPassword); err != nil {
		return fmt.Errorf("failed to send old password: %w", err)
	}

	// Read until we get new password prompt
	_, err = ws.readUntilString("New_Password:", 10*time.Second)
	if err != nil {
		_, err = ws.readUntilString("Password:", 10*time.Second)
		if err != nil {
			return fmt.Errorf("failed to get new password prompt: %w", err)
		}
	}

	// Send new password
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", newPassword); err != nil {
		return fmt.Errorf("failed to send new password: %w", err)
	}

	// Read response after password
	_, err = ws.readUntilPrompt(10 * time.Second)
	if err != nil {
		return fmt.Errorf("failed to read password response: %w", err)
	}

	logger.Debug().Str("component", "executor").Msg("Administrator password set successfully")
	return nil
}
