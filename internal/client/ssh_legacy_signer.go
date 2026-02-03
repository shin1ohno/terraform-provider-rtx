package client

import (
	"crypto"
	"crypto/rsa"
	"io"

	"golang.org/x/crypto/ssh"
)

// legacyRSASigner wraps an ssh.Signer to force the ssh-rsa signature algorithm
// for RSA keys. RTX routers (particularly RTX1210 with OpenSSH_7.7) only support
// the legacy ssh-rsa algorithm, not the newer rsa-sha2-256 or rsa-sha2-512.
type legacyRSASigner struct {
	signer ssh.Signer
}

// PublicKey returns the public key associated with the signer
func (s *legacyRSASigner) PublicKey() ssh.PublicKey {
	return s.signer.PublicKey()
}

// Sign signs the data using the legacy ssh-rsa algorithm
func (s *legacyRSASigner) Sign(rand io.Reader, data []byte) (*ssh.Signature, error) {
	// Get the underlying signer
	algorithmSigner, ok := s.signer.(ssh.AlgorithmSigner)
	if ok {
		// Use SignWithAlgorithm to explicitly request ssh-rsa
		return algorithmSigner.SignWithAlgorithm(rand, data, ssh.KeyAlgoRSA)
	}

	// Fallback: try to get the crypto signer directly
	cryptoSigner, ok := s.signer.(interface{ CryptoSigner() crypto.Signer })
	if ok {
		cs := cryptoSigner.CryptoSigner()
		if rsaKey, ok := cs.(*rsa.PrivateKey); ok {
			// Sign using SHA-1 (ssh-rsa algorithm)
			h := crypto.SHA1.New()
			h.Write(data)
			digest := h.Sum(nil)

			sig, err := rsa.SignPKCS1v15(rand, rsaKey, crypto.SHA1, digest)
			if err != nil {
				return nil, err
			}

			return &ssh.Signature{
				Format: ssh.KeyAlgoRSA,
				Blob:   sig,
			}, nil
		}
	}

	// Last resort: use default signing and hope for the best
	return s.signer.Sign(rand, data)
}

// SignWithAlgorithm implements ssh.AlgorithmSigner
// For RTX compatibility, we always use ssh-rsa regardless of the requested algorithm
func (s *legacyRSASigner) SignWithAlgorithm(randReader io.Reader, data []byte, algorithm string) (*ssh.Signature, error) {
	// Force ssh-rsa algorithm for RTX compatibility
	algorithmSigner, ok := s.signer.(ssh.AlgorithmSigner)
	if ok {
		return algorithmSigner.SignWithAlgorithm(randReader, data, ssh.KeyAlgoRSA)
	}
	return s.Sign(randReader, data)
}

// wrapSignerForRTX wraps an ssh.Signer to use the legacy ssh-rsa algorithm
// if it's an RSA key. Non-RSA keys (ed25519, ecdsa) are returned unchanged.
func wrapSignerForRTX(signer ssh.Signer) ssh.Signer {
	if signer == nil {
		return nil
	}

	pubKey := signer.PublicKey()
	if pubKey == nil {
		return signer
	}

	// Only wrap RSA keys - other key types don't have this compatibility issue
	if pubKey.Type() == ssh.KeyAlgoRSA {
		return &legacyRSASigner{signer: signer}
	}

	return signer
}

// Verify interface compliance at compile time
var _ ssh.Signer = (*legacyRSASigner)(nil)
var _ ssh.AlgorithmSigner = (*legacyRSASigner)(nil)
