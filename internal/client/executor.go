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
	// SetAdministratorPassword sets the administrator password using interactive prompts
	// oldPassword: current password, newPassword: new password to set
	SetAdministratorPassword(ctx context.Context, oldPassword, newPassword string) error
	// SetLoginPassword sets the login password using interactive prompts
	// oldPassword: current password (empty if not set), newPassword: new password to set
	SetLoginPassword(ctx context.Context, newPassword string) error
	// GenerateSSHDHostKey generates SSHD host key with interactive prompt handling
	// Handles confirmation prompt if host key already exists
	// Timeout: up to 10 minutes for key generation on slower hardware
	GenerateSSHDHostKey(ctx context.Context) error
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

// SetAdministratorPassword is not supported by sshExecutor
// Use simpleExecutor for interactive password commands
func (e *sshExecutor) SetAdministratorPassword(ctx context.Context, oldPassword, newPassword string) error {
	return fmt.Errorf("SetAdministratorPassword is not supported by sshExecutor, use simpleExecutor")
}

// SetLoginPassword is not supported by sshExecutor
// Use simpleExecutor for interactive password commands
func (e *sshExecutor) SetLoginPassword(ctx context.Context, newPassword string) error {
	return fmt.Errorf("SetLoginPassword is not supported by sshExecutor, use simpleExecutor")
}

// GenerateSSHDHostKey is not supported by sshExecutor
// Use simpleExecutor or pooledExecutor for interactive commands
func (e *sshExecutor) GenerateSSHDHostKey(ctx context.Context) error {
	return fmt.Errorf("GenerateSSHDHostKey is not supported by sshExecutor, use simpleExecutor or pooledExecutor")
}
