package client

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
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
