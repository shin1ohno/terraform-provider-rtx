package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// sshDialer is the default SSH connection dialer
type sshDialer struct{}

// Dial creates an SSH connection to the router
func (d *sshDialer) Dial(ctx context.Context, host string, config *Config) (Session, error) {
	logger := logging.FromContext(ctx)
	hostKeyCallback := d.getHostKeyCallback(config)

	sshConfig := &ssh.ClientConfig{
		User: config.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				// RTXルーターは通常パスワードプロンプトに対して単一の応答を期待
				answers := make([]string, len(questions))
				for i := range questions {
					logger.Debug().Int("question_index", i).Str("question", questions[i]).Msg("Keyboard interactive question")
					answers[i] = config.Password
				}
				return answers, nil
			}),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         time.Duration(config.Timeout) * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// Use DialContext to prevent goroutine leaks
	logger.Debug().Str("addr", addr).Msg("Dialing SSH")
	client, err := DialContext(ctx, "tcp", addr, sshConfig)
	if err != nil {
		// Check if it's an authentication error by examining the error message
		errMsg := err.Error()
		if strings.Contains(errMsg, "auth") || strings.Contains(errMsg, "permission denied") {
			return nil, fmt.Errorf("%w: %v", ErrAuthFailed, err)
		}
		return nil, err
	}
	logger.Debug().Msg("SSH connection established")
	
	// Use the working session implementation that matches our successful test
	session, err := newWorkingSession(client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create RTX session: %w", err)
	}
	
	return session, nil
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
	logger := logging.Global()
	logger.Warn().Msg("SSH host key verification is disabled. " +
		"This makes the connection vulnerable to man-in-the-middle attacks. " +
		"Consider using 'known_hosts_file' or 'host_key' for production environments.")
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