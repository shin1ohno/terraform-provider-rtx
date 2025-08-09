package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
	
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// sshDialer is the default SSH connection dialer
type sshDialer struct{}

// Dial creates an SSH connection to the router
func (d *sshDialer) Dial(ctx context.Context, host string, config *Config) (Session, error) {
	hostKeyCallback := d.getHostKeyCallback(config)
	
	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         time.Duration(config.Timeout) * time.Second,
	}
	
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	
	// Use DialContext to prevent goroutine leaks
	client, err := DialContext(ctx, "tcp", addr, sshConfig)
	if err != nil {
		// Check if it's an authentication error by examining the error message
		errMsg := err.Error()
		if strings.Contains(errMsg, "auth") || strings.Contains(errMsg, "permission denied") {
			return nil, fmt.Errorf("%w: %v", ErrAuthFailed, err)
		}
		return nil, err
	}
	
	return &sshSession{
		client: client,
		ctx:    ctx,
	}, nil
}

// sshSession wraps an SSH client to implement the Session interface
type sshSession struct {
	client *ssh.Client
	ctx    context.Context
	mu     sync.Mutex
}

// Send executes a command and returns the output
func (s *sshSession) Send(cmd string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Use RunCommandContext to prevent goroutine leaks
	output, err := RunCommandContext(s.ctx, s.client, cmd)
	if err != nil {
		// Check if it's a command execution error vs other errors
		if strings.Contains(err.Error(), "command execution failed") {
			// Still return the output even if there's an error
			// RTX routers might return non-zero exit codes for valid commands
			return output, fmt.Errorf("%w: %v", ErrCommandFailed, err)
		}
		return output, err
	}
	
	return output, nil
}

// Close closes the SSH connection
func (s *sshSession) Close() error {
	return s.client.Close()
}

// getHostKeyCallback returns the appropriate host key callback based on configuration
func (d *sshDialer) getHostKeyCallback(config *Config) ssh.HostKeyCallback {
	// If skip host key check is enabled, use insecure callback
	if config.SkipHostKeyCheck {
		return ssh.InsecureIgnoreHostKey()
	}
	
	// If a fixed host key is provided, use it for verification
	if config.HostKey != "" {
		return d.createFixedHostKeyCallback(config.HostKey)
	}
	
	// If known_hosts file is provided, use it for verification
	if config.KnownHostsFile != "" {
		callback, err := d.createKnownHostsCallback(config.KnownHostsFile)
		if err != nil {
			// Return a callback that will fail with the error
			return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return fmt.Errorf("failed to load known_hosts file %q: %w", config.KnownHostsFile, err)
			}
		}
		return callback
	}
	
	// Default to insecure (backward compatibility)
	return ssh.InsecureIgnoreHostKey()
}

// createFixedHostKeyCallback creates a callback that verifies against a fixed host key
func (d *sshDialer) createFixedHostKeyCallback(expectedKeyB64 string) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		// Decode the expected host key from base64
		expectedKeyData, err := base64.StdEncoding.DecodeString(expectedKeyB64)
		if err != nil {
			return fmt.Errorf("invalid host key format: %w", err)
		}
		
		// Get the provided key data
		providedKeyData := key.Marshal()
		
		// Compare the keys
		if len(expectedKeyData) != len(providedKeyData) {
			return fmt.Errorf("%w: host key mismatch for %s", ErrHostKeyMismatch, hostname)
		}
		
		for i, b := range expectedKeyData {
			if providedKeyData[i] != b {
				return fmt.Errorf("%w: host key mismatch for %s", ErrHostKeyMismatch, hostname)
			}
		}
		
		return nil
	}
}

// createKnownHostsCallback creates a callback that verifies against a known_hosts file
func (d *sshDialer) createKnownHostsCallback(knownHostsPath string) (ssh.HostKeyCallback, error) {
	callback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		return nil, err
	}
	
	// Wrap the callback to convert knownhosts errors to our custom error type
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := callback(hostname, remote, key)
		if err != nil {
			// Check if it's a key-related error and wrap it
			errStr := err.Error()
			if strings.Contains(errStr, "key") && (strings.Contains(errStr, "mismatch") || 
				strings.Contains(errStr, "changed") || strings.Contains(errStr, "unknown")) {
				return fmt.Errorf("%w: %s", ErrHostKeyMismatch, err.Error())
			}
			return err
		}
		return nil
	}, nil
}