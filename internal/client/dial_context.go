package client

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// DialContext creates an SSH connection with context support to prevent goroutine leaks
func DialContext(ctx context.Context, network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	// Step 1: Context-aware TCP dial
	d := &net.Dialer{
		Timeout: config.Timeout,
	}
	
	conn, err := d.DialContext(ctx, network, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial TCP: %w", err)
	}
	
	// Ensure connection is closed if SSH handshake fails
	var sshClient *ssh.Client
	defer func() {
		if sshClient == nil && conn != nil {
			conn.Close()
		}
	}()
	
	// Step 2: Upgrade the raw net.Conn to an SSH client connection
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, fmt.Errorf("SSH handshake failed: %w", err)
	}
	
	sshClient = ssh.NewClient(c, chans, reqs)
	
	// Step 3: Tie the lifetime of the SSH client to the context
	// This ensures all goroutines exit if the context is cancelled
	go func() {
		<-ctx.Done()
		sshClient.Close() // This also closes the underlying net.Conn
	}()
	
	return sshClient, nil
}

// RunCommandContext executes a command with context support
func RunCommandContext(ctx context.Context, client *ssh.Client, cmd string) ([]byte, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	
	// Set up a channel to receive the command result
	type result struct {
		output []byte
		err    error
	}
	resultCh := make(chan result, 1)
	
	// Run the command in a goroutine
	go func() {
		output, err := session.CombinedOutput(cmd)
		resultCh <- result{output: output, err: err}
		session.Close()
	}()
	
	// Wait for either the command to complete or context to be cancelled
	select {
	case <-ctx.Done():
		// Context cancelled, close the session
		session.Close()
		return nil, fmt.Errorf("command cancelled: %w", ctx.Err())
	case res := <-resultCh:
		if res.err != nil {
			// Still return output even on error
			// RTX routers might return non-zero exit codes for valid commands
			return res.output, fmt.Errorf("command execution failed: %w", res.err)
		}
		return res.output, nil
	}
}

// WithTimeout creates a context with timeout from seconds
func WithTimeout(ctx context.Context, timeoutSeconds int) (context.Context, context.CancelFunc) {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30 // Default timeout
	}
	return context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
}