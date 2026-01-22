//go:build ignore
// +build ignore

// This file contains unused session implementation that was kept for reference.
// It is excluded from builds via the ignore build tag.

package client

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"golang.org/x/crypto/ssh"
)

// rtxExpectSession implements a simple expect-like session for RTX routers
type rtxExpectSession struct {
	client  *ssh.Client
	session *ssh.Session
	stdin   io.WriteCloser
	stdout  io.Reader
	mu      sync.Mutex
	closed  bool
	buffer  bytes.Buffer
}

// newRTXExpectSession creates a new expect-style session
func newRTXExpectSession(client *ssh.Client) (*rtxExpectSession, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Set up pipes before requesting PTY
	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to get stdin: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to get stdout: %w", err)
	}

	// Request PTY
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// Use wide terminal to prevent line wrapping for long filter lists
	// RTX config lines can exceed 200 characters (e.g., secure filter with 13+ IDs)
	// RequestPty parameters: term, height, width, modes
	if err := session.RequestPty("vt100", 40, 512, modes); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to request PTY: %w", err)
	}

	// Start shell without waiting
	if err := session.Shell(); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to start shell: %w", err)
	}

	s := &rtxExpectSession{
		client:  client,
		session: session,
		stdin:   stdin,
		stdout:  stdout,
	}

	// Start reading stdout in background
	go s.readLoop()

	// Wait for initial prompt
	if err := s.expectPrompt(10 * time.Second); err != nil {
		s.Close()
		return nil, fmt.Errorf("failed to get initial prompt: %w", err)
	}

	// Set character encoding
	if err := s.sendLine("console character en.ascii"); err != nil {
		logging.Global().Warn().Str("component", "rtx-expect-session").Msgf("Failed to set character encoding: %v", err)
		// Continue anyway as some RTX models might not support this
	} else {
		s.expectPrompt(5 * time.Second)
	}

	return s, nil
}

// readLoop continuously reads from stdout
func (s *rtxExpectSession) readLoop() {
	buf := make([]byte, 4096)
	for {
		n, err := s.stdout.Read(buf)
		if n > 0 {
			s.mu.Lock()
			s.buffer.Write(buf[:n])
			s.mu.Unlock()
		}
		if err != nil {
			break
		}
	}
}

// expectPrompt waits for a prompt
func (s *rtxExpectSession) expectPrompt(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		s.mu.Lock()
		content := s.buffer.String()
		s.mu.Unlock()

		// Check for prompt
		if strings.Contains(content, ">") || strings.Contains(content, "#") {
			// Check if it's at the end of a line
			lines := strings.Split(content, "\n")
			if len(lines) > 0 {
				lastLine := strings.TrimSpace(lines[len(lines)-1])
				if strings.HasSuffix(lastLine, ">") || strings.HasSuffix(lastLine, "#") {
					return nil
				}
			}
		}

		time.Sleep(50 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for prompt")
}

// sendLine sends a command line
func (s *rtxExpectSession) sendLine(cmd string) error {
	_, err := fmt.Fprintf(s.stdin, "%s\r\n", cmd)
	return err
}

// Send executes a command and returns the output
func (s *rtxExpectSession) Send(cmd string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, fmt.Errorf("session is closed")
	}

	// Clear buffer
	s.buffer.Reset()

	// Send command
	if err := s.sendLine(cmd); err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Wait for prompt
	if err := s.expectPrompt(30 * time.Second); err != nil {
		return nil, fmt.Errorf("failed to get response: %w", err)
	}

	// Get output
	output := s.buffer.String()

	// Remove command echo and prompt
	lines := strings.Split(output, "\n")
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == cmd {
		lines = lines[1:]
	}
	if len(lines) > 0 {
		lastLine := lines[len(lines)-1]
		if strings.HasSuffix(strings.TrimSpace(lastLine), ">") ||
			strings.HasSuffix(strings.TrimSpace(lastLine), "#") {
			lines = lines[:len(lines)-1]
		}
	}

	return []byte(strings.Join(lines, "\n")), nil
}

// Close closes the session
func (s *rtxExpectSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true

	// Send exit
	fmt.Fprintln(s.stdin, "exit")
	time.Sleep(100 * time.Millisecond)

	// Close session
	if s.session != nil {
		return s.session.Close()
	}

	return nil
}
