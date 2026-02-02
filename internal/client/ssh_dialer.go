package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"

	"github.com/sh1/terraform-provider-rtx/internal/logging"
)

// sshDialer is the default SSH connection dialer
type sshDialer struct{}

// Dial creates an SSH connection to the router
func (d *sshDialer) Dial(ctx context.Context, host string, config *Config) (Session, error) {
	logger := logging.FromContext(ctx)
	hostKeyCallback := d.getHostKeyCallback(config)
	authMethods := d.buildAuthMethods(config)

	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         time.Duration(config.Timeout) * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// Use DialContext to prevent goroutine leaks
	logger.Debug().Str("addr", addr).Int("auth_methods_count", len(authMethods)).Msg("Dialing SSH")
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

// BuildAuthMethods builds authentication methods in priority order.
// This is the exported version for use by other components (e.g., client.go).
// Priority: 1) Explicit private key, 2) SSH agent (if no explicit key), 3) Password auth as fallback
func BuildAuthMethods(config *Config) []ssh.AuthMethod {
	d := &sshDialer{}
	return d.buildAuthMethods(config)
}

// buildAuthMethods builds authentication methods in priority order.
// Priority: 1) Explicit private key, 2) SSH agent (if no explicit key), 3) Password auth as fallback
func (d *sshDialer) buildAuthMethods(config *Config) []ssh.AuthMethod {
	logger := logging.Global()
	var methods []ssh.AuthMethod

	hasExplicitKey := config.PrivateKey != "" || config.PrivateKeyFile != ""

	logger.Debug().
		Bool("has_explicit_key", hasExplicitKey).
		Bool("has_private_key", config.PrivateKey != "").
		Bool("has_private_key_file", config.PrivateKeyFile != "").
		Bool("has_password", config.Password != "").
		Int("private_key_len", len(config.PrivateKey)).
		Msg("buildAuthMethods: checking authentication options")

	// If no explicit key is provided, try SSH agent first
	if !hasExplicitKey {
		if agentAuth := d.trySSHAgent(); agentAuth != nil {
			logger.Debug().Msg("SSH agent authentication available")
			methods = append(methods, agentAuth)
		}
	}

	// If explicit private key is provided, use it
	if hasExplicitKey {
		signer := d.loadPrivateKey(config)
		if signer != nil {
			fingerprint := ssh.FingerprintSHA256(signer.PublicKey())
			logger.Debug().Str("fingerprint", fingerprint).Msg("Private key authentication configured")
			methods = append(methods, ssh.PublicKeys(signer))
		} else {
			logger.Error().Msg("Failed to load private key - signer is nil")
		}
	}

	// Always include password authentication as fallback if password is set
	if config.Password != "" {
		logger.Debug().Msg("Password authentication configured")
		methods = append(methods, ssh.Password(config.Password))
		// Also add keyboard-interactive for RTX router compatibility
		methods = append(methods, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
			// RTX routers typically expect a single response to password prompts
			answers := make([]string, len(questions))
			for i := range questions {
				logger.Debug().Int("question_index", i).Str("question", questions[i]).Msg("Keyboard interactive question")
				answers[i] = config.Password
			}
			return answers, nil
		}))
	}

	if len(methods) == 0 {
		logger.Warn().Msg("No authentication methods available")
	}

	return methods
}

// loadPrivateKey loads a private key from configuration.
// Returns nil if loading fails (auth will fall back to other methods).
func (d *sshDialer) loadPrivateKey(config *Config) ssh.Signer {
	logger := logging.Global()

	var keyData []byte
	var err error

	if config.PrivateKey != "" {
		// Use key content directly
		keyData = []byte(config.PrivateKey)
		logger.Debug().Msg("Using private key from content")
	} else if config.PrivateKeyFile != "" {
		// Read key from file, handling ~ expansion
		keyPath := config.PrivateKeyFile
		if strings.HasPrefix(keyPath, "~/") {
			homeDir, homeErr := os.UserHomeDir()
			if homeErr != nil {
				logger.Error().Err(homeErr).Msg("Failed to get user home directory")
				return nil
			}
			keyPath = homeDir + keyPath[1:]
		}

		keyData, err = os.ReadFile(keyPath)
		if err != nil {
			logger.Error().Err(err).Str("file", keyPath).Msg("Failed to read private key file")
			return nil
		}
		logger.Debug().Str("file", keyPath).Msg("Using private key from file")
	} else {
		return nil
	}

	var signer ssh.Signer
	if config.PrivateKeyPassphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(keyData, []byte(config.PrivateKeyPassphrase))
		if err != nil {
			logger.Error().Err(err).Msg("Failed to parse encrypted private key")
			return nil
		}
	} else {
		signer, err = ssh.ParsePrivateKey(keyData)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to parse private key")
			return nil
		}
	}

	return signer
}

// trySSHAgent attempts to connect to SSH agent and returns an auth method.
// Returns nil if SSH agent is not available.
func (d *sshDialer) trySSHAgent() ssh.AuthMethod {
	logger := logging.Global()

	socketPath := os.Getenv("SSH_AUTH_SOCK")
	if socketPath == "" {
		logger.Debug().Msg("SSH_AUTH_SOCK not set, SSH agent not available")
		return nil
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		logger.Debug().Err(err).Str("socket", socketPath).Msg("Failed to connect to SSH agent")
		return nil
	}

	agentClient := agent.NewClient(conn)
	return ssh.PublicKeysCallback(agentClient.Signers)
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
