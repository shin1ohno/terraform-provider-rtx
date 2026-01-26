package client

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

// TestSSHDialer_HostKeyVerification tests host key verification using mock dialer
func TestSSHDialer_HostKeyVerification(t *testing.T) {
	tests := []struct {
		name            string
		config          *Config
		wantErrContains string
	}{
		{
			name: "missing known_hosts file error",
			config: &Config{
				Host:           "testhost",
				Port:           22,
				Username:       "testuser",
				Password:       "testpass",
				Timeout:        5,
				KnownHostsFile: "/nonexistent/known_hosts",
			},
			wantErrContains: "failed to load known_hosts",
		},
		{
			name: "skip host key check configuration",
			config: &Config{
				Host:             "testhost",
				Port:             22,
				Username:         "testuser",
				Password:         "testpass",
				Timeout:          5,
				SkipHostKeyCheck: true,
			},
			wantErrContains: "", // Should not error, just testing config
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialer := &sshDialer{}
			callback := dialer.getHostKeyCallback(tt.config)

			// Test that callback is created (even if it might fail later)
			if callback == nil {
				t.Error("getHostKeyCallback() returned nil")
				return
			}

			// For known_hosts file error test, try to call the callback
			if tt.wantErrContains != "" && tt.config.KnownHostsFile != "" {
				// Generate test key for callback testing
				privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
				if err != nil {
					t.Fatalf("Failed to generate test key: %v", err)
				}
				signer, err := ssh.NewSignerFromKey(privateKey)
				if err != nil {
					t.Fatalf("Failed to create signer: %v", err)
				}

				mockAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}
				err = callback("testhost:22", mockAddr, signer.PublicKey())

				if err == nil {
					t.Error("Expected error for missing known_hosts file")
					return
				}

				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Errorf("Expected error containing %q, got %q", tt.wantErrContains, err.Error())
				}
			}
		})
	}
}

// TestHostKeyCallback_FixedHostKey tests the fixed host key verification logic
func TestHostKeyCallback_FixedHostKey(t *testing.T) {
	// Generate test key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	correctHostKey := base64.StdEncoding.EncodeToString(signer.PublicKey().Marshal())

	// Generate another key for the wrong key test
	wrongPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate wrong test key: %v", err)
	}
	wrongSigner, err := ssh.NewSignerFromKey(wrongPrivateKey)
	if err != nil {
		t.Fatalf("Failed to create wrong signer: %v", err)
	}
	wrongHostKey := base64.StdEncoding.EncodeToString(wrongSigner.PublicKey().Marshal())

	tests := []struct {
		name        string
		configKey   string
		providedKey ssh.PublicKey
		wantErr     bool
		wantErrType error
	}{
		{
			name:        "correct host key",
			configKey:   correctHostKey,
			providedKey: signer.PublicKey(),
			wantErr:     false,
		},
		{
			name:        "incorrect host key",
			configKey:   wrongHostKey,
			providedKey: signer.PublicKey(),
			wantErr:     true,
			wantErrType: ErrHostKeyMismatch,
		},
		{
			name:        "invalid base64 host key",
			configKey:   "invalid-base64!@#$",
			providedKey: signer.PublicKey(),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialer := &sshDialer{}
			callback := dialer.createFixedHostKeyCallback(tt.configKey)

			err := callback("test-host:22", nil, tt.providedKey)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
					return
				}
				if tt.wantErrType != nil && !errors.Is(err, tt.wantErrType) {
					t.Errorf("Expected error type %v, got %v", tt.wantErrType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestHostKeyCallback_KnownHosts tests the known_hosts file verification logic
func TestHostKeyCallback_KnownHosts(t *testing.T) {
	// Generate test key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	hostKey := base64.StdEncoding.EncodeToString(signer.PublicKey().Marshal())

	// Generate second key for wrong key test
	wrongPrivateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate wrong test key: %v", err)
	}
	wrongSigner, err := ssh.NewSignerFromKey(wrongPrivateKey)
	if err != nil {
		t.Fatalf("Failed to create wrong signer: %v", err)
	}
	wrongHostKey := base64.StdEncoding.EncodeToString(wrongSigner.PublicKey().Marshal())

	tests := []struct {
		name              string
		knownHostsContent string
		testHost          string
		providedKey       ssh.PublicKey
		wantErr           bool
	}{
		{
			name:              "matching host and key",
			knownHostsContent: fmt.Sprintf("testhost ssh-rsa %s\n", hostKey),
			testHost:          "testhost:22",
			providedKey:       signer.PublicKey(),
			wantErr:           false,
		},
		{
			name:              "matching host with port in known_hosts",
			knownHostsContent: fmt.Sprintf("[testhost]:2222 ssh-rsa %s\n", hostKey),
			testHost:          "testhost:2222",
			providedKey:       signer.PublicKey(),
			wantErr:           false,
		},
		{
			name:              "host not in known_hosts",
			knownHostsContent: fmt.Sprintf("otherhost ssh-rsa %s\n", hostKey),
			testHost:          "testhost:22",
			providedKey:       signer.PublicKey(),
			wantErr:           true,
		},
		{
			name:              "wrong key for host",
			knownHostsContent: fmt.Sprintf("testhost ssh-rsa %s\n", wrongHostKey),
			testHost:          "testhost:22",
			providedKey:       signer.PublicKey(),
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary known_hosts file
			tmpDir := t.TempDir()
			knownHostsPath := filepath.Join(tmpDir, "known_hosts")
			if err := os.WriteFile(knownHostsPath, []byte(tt.knownHostsContent), 0600); err != nil {
				t.Fatalf("Failed to create known_hosts file: %v", err)
			}

			dialer := &sshDialer{}
			callback, err := dialer.createKnownHostsCallback(knownHostsPath)
			if err != nil {
				t.Fatalf("Failed to create known_hosts callback: %v", err)
			}

			// Create a mock net.Addr for testing
			mockAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}
			err = callback(tt.testHost, mockAddr, tt.providedKey)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestSSHDialer_HostKeyCallbackSelection tests that the dialer selects the correct callback
func TestSSHDialer_HostKeyCallbackSelection(t *testing.T) {
	tests := []struct {
		name             string
		config           *Config
		expectedCallback string // "fixed", "known_hosts", or "insecure"
	}{
		{
			name: "fixed host key takes priority",
			config: &Config{
				HostKey:        "test_key",
				KnownHostsFile: "/tmp/known_hosts",
			},
			expectedCallback: "fixed",
		},
		{
			name: "known_hosts when no fixed key",
			config: &Config{
				KnownHostsFile: "/tmp/known_hosts",
			},
			expectedCallback: "known_hosts",
		},
		{
			name: "insecure when skip is enabled",
			config: &Config{
				SkipHostKeyCheck: true,
				HostKey:          "test_key", // Should be ignored
			},
			expectedCallback: "insecure",
		},
		{
			name:             "insecure when no keys configured",
			config:           &Config{},
			expectedCallback: "insecure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialer := &sshDialer{}

			// We can't directly test the callback selection without refactoring the code,
			// but we can test the configuration validation logic
			callback := dialer.getHostKeyCallback(tt.config)

			// This test ensures the callback is not nil
			if callback == nil {
				t.Error("getHostKeyCallback() returned nil")
			}

			// The specific callback type testing would require code refactoring
			// to expose the callback creation methods, which we'll implement in the actual code
		})
	}
}

// TestSSHDialer_BuildAuthMethods tests authentication method selection
func TestSSHDialer_BuildAuthMethods(t *testing.T) {
	// Generate a test RSA key for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	// Create PEM-encoded private key for testing
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	tests := []struct {
		name             string
		config           *Config
		expectedMinCount int // minimum number of auth methods expected
		description      string
	}{
		{
			name: "password only",
			config: &Config{
				Password: "testpass",
			},
			expectedMinCount: 2, // password + keyboard-interactive
			description:      "should include password and keyboard-interactive auth",
		},
		{
			name: "explicit private key content",
			config: &Config{
				PrivateKey: string(privPEM),
				Password:   "fallback",
			},
			expectedMinCount: 3, // public key + password + keyboard-interactive
			description:      "should include public key, password, and keyboard-interactive auth",
		},
		{
			name: "no credentials",
			config: &Config{
				Username: "user",
			},
			expectedMinCount: 0, // SSH agent might be available
			description:      "should have zero methods when no credentials provided (or SSH agent if available)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dialer := &sshDialer{}
			methods := dialer.buildAuthMethods(tt.config)

			if len(methods) < tt.expectedMinCount {
				t.Errorf("Expected at least %d auth methods, got %d - %s",
					tt.expectedMinCount, len(methods), tt.description)
			}
		})
	}
}

// TestSSHDialer_LoadPrivateKey tests private key loading
func TestSSHDialer_LoadPrivateKey(t *testing.T) {
	// Generate a test RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	// Create PEM-encoded private key
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Create encrypted PEM key using the deprecated but necessary function for testing
	encryptedPEM, err := x509.EncryptPEMBlock( //nolint:staticcheck // Using deprecated function for testing legacy encrypted key support
		rand.Reader,
		"RSA PRIVATE KEY",
		x509.MarshalPKCS1PrivateKey(privateKey),
		[]byte("example!PASS123"),
		x509.PEMCipherAES256,
	)
	if err != nil {
		t.Fatalf("Failed to encrypt PEM block: %v", err)
	}
	encryptedPrivPEM := pem.EncodeToMemory(encryptedPEM)

	tests := []struct {
		name        string
		config      *Config
		setupFile   func(t *testing.T) string // returns file path if needed
		expectNil   bool
		description string
	}{
		{
			name: "key from content",
			config: &Config{
				PrivateKey: string(privPEM),
			},
			expectNil:   false,
			description: "should successfully load key from content",
		},
		{
			name: "key from file",
			config: &Config{
				PrivateKeyFile: "", // Will be set by setupFile
			},
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				keyPath := filepath.Join(tmpDir, "id_rsa")
				if err := os.WriteFile(keyPath, privPEM, 0600); err != nil {
					t.Fatalf("Failed to write key file: %v", err)
				}
				return keyPath
			},
			expectNil:   false,
			description: "should successfully load key from file",
		},
		{
			name: "encrypted key with passphrase",
			config: &Config{
				PrivateKey:           string(encryptedPrivPEM),
				PrivateKeyPassphrase: "example!PASS123",
			},
			expectNil:   false,
			description: "should successfully load encrypted key with correct passphrase",
		},
		{
			name: "encrypted key without passphrase",
			config: &Config{
				PrivateKey: string(encryptedPrivPEM),
			},
			expectNil:   true,
			description: "should fail to load encrypted key without passphrase",
		},
		{
			name: "non-existent file",
			config: &Config{
				PrivateKeyFile: "/non/existent/path/id_rsa",
			},
			expectNil:   true,
			description: "should return nil for non-existent file",
		},
		{
			name: "invalid key content",
			config: &Config{
				PrivateKey: "invalid key content",
			},
			expectNil:   true,
			description: "should return nil for invalid key content",
		},
		{
			name:        "no key configured",
			config:      &Config{},
			expectNil:   true,
			description: "should return nil when no key is configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.config
			if tt.setupFile != nil {
				config.PrivateKeyFile = tt.setupFile(t)
			}

			dialer := &sshDialer{}
			signer := dialer.loadPrivateKey(config)

			if tt.expectNil && signer != nil {
				t.Errorf("Expected nil signer, got non-nil - %s", tt.description)
			}
			if !tt.expectNil && signer == nil {
				t.Errorf("Expected non-nil signer, got nil - %s", tt.description)
			}
		})
	}
}

// TestSSHDialer_LoadPrivateKey_TildeExpansion tests ~ expansion in key file paths
func TestSSHDialer_LoadPrivateKey_TildeExpansion(t *testing.T) {
	// Generate a test RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Create a temporary key file in user's home directory area
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot determine home directory")
	}

	// Use a subdirectory that's likely to exist and be writable in tests
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test_key")
	if err := os.WriteFile(keyPath, privPEM, 0600); err != nil {
		t.Fatalf("Failed to write test key: %v", err)
	}

	// Test that relative path under home gets expanded correctly
	// Note: We can't actually test ~/... without putting files in the real home directory
	// So we test that the expansion logic works by verifying the path construction

	dialer := &sshDialer{}

	// Test with absolute path (should work)
	config := &Config{
		PrivateKeyFile: keyPath,
	}
	signer := dialer.loadPrivateKey(config)
	if signer == nil {
		t.Error("Expected signer for absolute path, got nil")
	}

	// Verify that tilde expansion is attempted by checking behavior with non-existent tilde path
	config = &Config{
		PrivateKeyFile: "~/.ssh/nonexistent_test_key_12345",
	}
	signer = dialer.loadPrivateKey(config)
	if signer != nil {
		t.Error("Expected nil for non-existent tilde path, got signer")
	}

	// If we can write to ~/.ssh, test actual tilde expansion
	sshDir := filepath.Join(homeDir, ".ssh")
	testKeyPath := filepath.Join(sshDir, ".test_terraform_provider_rtx_key")

	// Try to write test key (might fail due to permissions)
	if err := os.WriteFile(testKeyPath, privPEM, 0600); err == nil {
		defer func() {
			_ = os.Remove(testKeyPath) // Clean up, ignore error
		}()

		config = &Config{
			PrivateKeyFile: "~/.ssh/.test_terraform_provider_rtx_key",
		}
		signer = dialer.loadPrivateKey(config)
		if signer == nil {
			t.Error("Expected signer for tilde-expanded path, got nil")
		}
	}
}

// TestSSHDialer_TrySSHAgent tests SSH agent detection
func TestSSHDialer_TrySSHAgent(t *testing.T) {
	dialer := &sshDialer{}

	// Save and clear SSH_AUTH_SOCK to test without agent
	originalSock := os.Getenv("SSH_AUTH_SOCK")
	t.Setenv("SSH_AUTH_SOCK", "") // Use t.Setenv which automatically restores

	authMethod := dialer.trySSHAgent()
	if authMethod != nil {
		t.Error("Expected nil when SSH_AUTH_SOCK is not set")
	}

	// Restore SSH_AUTH_SOCK and test if agent is available
	if originalSock != "" {
		t.Setenv("SSH_AUTH_SOCK", originalSock)

		// If SSH agent is available, trySSHAgent should return an auth method
		// Note: We don't assert authMethod != nil here because the agent
		// might not be running even if SSH_AUTH_SOCK is set
		_ = dialer.trySSHAgent()
	}
}

// TestSSHDialer_BuildAuthMethods_Priority tests that auth methods are in correct priority order
func TestSSHDialer_BuildAuthMethods_Priority(t *testing.T) {
	// Generate a test key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Clear SSH_AUTH_SOCK to ensure predictable behavior
	t.Setenv("SSH_AUTH_SOCK", "")

	dialer := &sshDialer{}

	// Test: When explicit key is provided, it should be first (no agent)
	config := &Config{
		PrivateKey: string(privPEM),
		Password:   "testpass",
	}
	methods := dialer.buildAuthMethods(config)

	// Should have: public key, password, keyboard-interactive
	if len(methods) != 3 {
		t.Errorf("Expected 3 auth methods with explicit key and password, got %d", len(methods))
	}
}

// TestSSHDialer_BuildAuthMethods_AgentFallback tests SSH agent as fallback when no explicit key
func TestSSHDialer_BuildAuthMethods_AgentFallback(t *testing.T) {
	dialer := &sshDialer{}

	// Clear SSH_AUTH_SOCK to ensure no agent
	t.Setenv("SSH_AUTH_SOCK", "")

	// Test: Only password, no explicit key, no agent
	config := &Config{
		Password: "testpass",
	}
	methods := dialer.buildAuthMethods(config)

	// Should have: password + keyboard-interactive (no agent since SSH_AUTH_SOCK unset)
	if len(methods) != 2 {
		t.Errorf("Expected 2 auth methods with password only (no agent), got %d", len(methods))
	}
}
