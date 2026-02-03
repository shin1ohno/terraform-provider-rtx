package client

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestWrapSignerForRTX_RSAKey(t *testing.T) {
	// Generate an RSA key for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	// Wrap the signer
	wrapped := wrapSignerForRTX(signer)

	// Verify it's wrapped (should be a legacyRSASigner)
	if _, ok := wrapped.(*legacyRSASigner); !ok {
		t.Errorf("RSA signer should be wrapped as legacyRSASigner, got %T", wrapped)
	}

	// Verify public key is still accessible
	if wrapped.PublicKey() == nil {
		t.Error("Wrapped signer should return public key")
	}

	// Verify public key type is still ssh-rsa
	if wrapped.PublicKey().Type() != ssh.KeyAlgoRSA {
		t.Errorf("Public key type = %v, want %v", wrapped.PublicKey().Type(), ssh.KeyAlgoRSA)
	}
}

func TestWrapSignerForRTX_Ed25519Key(t *testing.T) {
	// Generate an Ed25519 key for testing
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate Ed25519 key: %v", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	// Wrap the signer
	wrapped := wrapSignerForRTX(signer)

	// Verify it's NOT wrapped (Ed25519 doesn't need legacy algorithm handling)
	if _, ok := wrapped.(*legacyRSASigner); ok {
		t.Error("Ed25519 signer should NOT be wrapped as legacyRSASigner")
	}

	// Verify it returns the original signer
	if wrapped != signer {
		t.Error("Ed25519 signer should be returned unchanged")
	}
}

func TestWrapSignerForRTX_NilSigner(t *testing.T) {
	wrapped := wrapSignerForRTX(nil)
	if wrapped != nil {
		t.Error("wrapSignerForRTX(nil) should return nil")
	}
}

func TestLegacyRSASigner_Sign(t *testing.T) {
	// Generate an RSA key for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	wrapped := wrapSignerForRTX(signer)

	// Test signing
	testData := []byte("test data to sign")
	sig, err := wrapped.Sign(rand.Reader, testData)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	// Verify signature format is ssh-rsa (legacy)
	if sig.Format != ssh.KeyAlgoRSA {
		t.Errorf("Signature format = %v, want %v", sig.Format, ssh.KeyAlgoRSA)
	}

	// Verify signature blob is not empty
	if len(sig.Blob) == 0 {
		t.Error("Signature blob should not be empty")
	}
}

func TestLegacyRSASigner_SignWithAlgorithm(t *testing.T) {
	// Generate an RSA key for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	wrapped := wrapSignerForRTX(signer)

	// Verify wrapped signer implements AlgorithmSigner
	algSigner, ok := wrapped.(ssh.AlgorithmSigner)
	if !ok {
		t.Fatal("Wrapped signer should implement ssh.AlgorithmSigner")
	}

	// Test signing with different algorithms - should always use ssh-rsa
	testData := []byte("test data to sign")

	testCases := []struct {
		name      string
		algorithm string
	}{
		{"rsa-sha2-256", ssh.KeyAlgoRSASHA256},
		{"rsa-sha2-512", ssh.KeyAlgoRSASHA512},
		{"ssh-rsa", ssh.KeyAlgoRSA},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sig, err := algSigner.SignWithAlgorithm(rand.Reader, testData, tc.algorithm)
			if err != nil {
				t.Fatalf("SignWithAlgorithm(%q) error = %v", tc.algorithm, err)
			}

			// Should always return ssh-rsa format for RTX compatibility
			if sig.Format != ssh.KeyAlgoRSA {
				t.Errorf("Signature format = %v, want %v (legacy ssh-rsa for RTX)", sig.Format, ssh.KeyAlgoRSA)
			}
		})
	}
}

func TestSSHConfig_HostKeyAlgorithmsIncludesLegacy(t *testing.T) {
	// This test verifies the documentation/intent rather than actual config creation
	// The actual HostKeyAlgorithms are set in ssh_dialer.go and sftp_client.go

	expectedAlgorithms := []string{
		ssh.KeyAlgoRSA, // ssh-rsa (legacy, required by RTX)
	}

	for _, algo := range expectedAlgorithms {
		// Just verify the constants are valid
		if algo == "" {
			t.Errorf("Algorithm constant should not be empty")
		}
	}

	// Verify ssh.KeyAlgoRSA is "ssh-rsa"
	if ssh.KeyAlgoRSA != "ssh-rsa" {
		t.Errorf("ssh.KeyAlgoRSA = %q, want \"ssh-rsa\"", ssh.KeyAlgoRSA)
	}
}
