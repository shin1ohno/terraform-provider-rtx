package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"github.com/sh1/terraform-provider-rtx/internal/logging"
)

// ErrSFTPClosed indicates the SFTP client has been closed
var ErrSFTPClosed = errors.New("sftp client is closed")

// SFTPClient provides SFTP file download functionality
type SFTPClient interface {
	// Download downloads a file from the remote server and returns its content in memory
	Download(ctx context.Context, path string) ([]byte, error)

	// Close closes the SFTP connection
	Close() error
}

// sshClientInterface abstracts the SSH client for testing
type sshClientInterface interface {
	NewSession() (sshSessionInterface, error)
	Close() error
}

// sshSessionInterface abstracts the SSH session for testing
type sshSessionInterface interface {
	RequestSubsystem(subsystem string) error
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.Reader, error)
	Close() error
}

// sftpClientInterface abstracts the SFTP client for testing
type sftpClientInterface interface {
	Open(path string) (sftpFileInterface, error)
	Close() error
}

// sftpFileInterface abstracts the SFTP file for testing
type sftpFileInterface interface {
	io.Reader
	Close() error
}

// sftpClientImpl is the implementation of SFTPClient
type sftpClientImpl struct {
	sshClient  sshClientInterface
	sftpClient sftpClientInterface
	mu         sync.Mutex
	closed     bool
}

// sshClientWrapper wraps *ssh.Client to implement sshClientInterface
type sshClientWrapper struct {
	client *ssh.Client
}

func (w *sshClientWrapper) NewSession() (sshSessionInterface, error) {
	sess, err := w.client.NewSession()
	if err != nil {
		return nil, err
	}
	return &sshSessionWrapper{session: sess}, nil
}

func (w *sshClientWrapper) Close() error {
	return w.client.Close()
}

// sshSessionWrapper wraps *ssh.Session to implement sshSessionInterface
type sshSessionWrapper struct {
	session *ssh.Session
}

func (w *sshSessionWrapper) RequestSubsystem(subsystem string) error {
	return w.session.RequestSubsystem(subsystem)
}

func (w *sshSessionWrapper) StdinPipe() (io.WriteCloser, error) {
	return w.session.StdinPipe()
}

func (w *sshSessionWrapper) StdoutPipe() (io.Reader, error) {
	return w.session.StdoutPipe()
}

func (w *sshSessionWrapper) Close() error {
	return w.session.Close()
}

// sftpClientWrapper wraps *sftp.Client to implement sftpClientInterface
type sftpClientWrapper struct {
	client *sftp.Client
}

func (w *sftpClientWrapper) Open(path string) (sftpFileInterface, error) {
	return w.client.Open(path)
}

func (w *sftpClientWrapper) Close() error {
	return w.client.Close()
}

// NewSFTPClient creates a new SFTP client connected to the router
func NewSFTPClient(ctx context.Context, config *Config) (SFTPClient, error) {
	logger := logging.FromContext(ctx)

	// Build SSH config
	dialer := &sshDialer{}
	hostKeyCallback := dialer.getHostKeyCallback(config)

	sshConfig := &ssh.ClientConfig{
		User: config.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				answers := make([]string, len(questions))
				for i := range questions {
					answers[i] = config.Password
				}
				return answers, nil
			}),
		},
		HostKeyCallback: hostKeyCallback,
		Timeout:         time.Duration(config.Timeout) * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// Establish SSH connection
	logger.Debug().Str("addr", addr).Msg("Establishing SSH connection for SFTP")
	sshClient, err := DialContext(ctx, "tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to establish SSH connection for SFTP: %w", err)
	}

	// Create SFTP client
	logger.Debug().Msg("Creating SFTP client")
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	logger.Debug().Msg("SFTP client created successfully")

	return &sftpClientImpl{
		sshClient:  &sshClientWrapper{client: sshClient},
		sftpClient: &sftpClientWrapper{client: sftpClient},
	}, nil
}

// Download downloads a file from the remote server and returns its content in memory
func (c *sftpClientImpl) Download(ctx context.Context, path string) ([]byte, error) {
	// Check context cancellation first
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, ErrSFTPClosed
	}
	c.mu.Unlock()

	logger := logging.FromContext(ctx)
	logger.Debug().Str("path", path).Msg("Downloading file via SFTP")

	// Open the remote file
	file, err := c.sftpClient.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", path, err)
	}
	defer file.Close()

	// Read entire content into memory
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", path, err)
	}

	logger.Debug().Int("bytes", len(content)).Msg("File downloaded successfully")
	return content, nil
}

// Close closes the SFTP connection
func (c *sftpClientImpl) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	var errs []error

	if c.sftpClient != nil {
		if err := c.sftpClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close SFTP client: %w", err))
		}
	}

	if c.sshClient != nil {
		if err := c.sshClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close SSH connection: %w", err))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
