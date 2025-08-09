package client

import (
	"context"
	"fmt"
	"time"
)

// Executor interface defines how commands are executed
type Executor interface {
	Run(ctx context.Context, cmd string) ([]byte, error)
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
	matched, _ := e.promptDetector.DetectPrompt(raw)
	if !matched {
		return nil, fmt.Errorf("%w: output does not contain expected prompt", ErrPrompt)
	}

	return raw, nil
}