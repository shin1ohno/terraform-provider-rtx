package client

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	
	"golang.org/x/crypto/ssh"
)

// sshDialer is the default SSH connection dialer
type sshDialer struct{}

// Dial creates an SSH connection to the router
func (d *sshDialer) Dial(ctx context.Context, host string, config *Config) (Session, error) {
	sshConfig := &ssh.ClientConfig{
		User: config.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Consider stricter host key verification
		Timeout:         time.Duration(config.Timeout) * time.Second,
	}
	
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	
	// Create a channel to handle the connection result
	type result struct {
		client *ssh.Client
		err    error
	}
	ch := make(chan result, 1)
	
	go func() {
		client, err := ssh.Dial("tcp", addr, sshConfig)
		ch <- result{client, err}
	}()
	
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("%w: %v", ErrTimeout, ctx.Err())
	case res := <-ch:
		if res.err != nil {
			// Check if it's an authentication error by examining the error message
			errMsg := res.err.Error()
			if strings.Contains(errMsg, "auth") || strings.Contains(errMsg, "permission denied") {
				return nil, fmt.Errorf("%w: %v", ErrAuthFailed, res.err)
			}
			return nil, res.err
		}
		return &sshSession{client: res.client}, nil
	}
}

// sshSession wraps an SSH client to implement the Session interface
type sshSession struct {
	client *ssh.Client
	mu     sync.Mutex
}

// Send executes a command and returns the output
func (s *sshSession) Send(cmd string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	session, err := s.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()
	
	// Execute the command
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		// Still return the output even if there's an error
		// RTX routers might return non-zero exit codes for valid commands
		return output, fmt.Errorf("%w: %v", ErrCommandFailed, err)
	}
	
	return output, nil
}

// Close closes the SSH connection
func (s *sshSession) Close() error {
	return s.client.Close()
}