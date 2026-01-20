package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// simpleRTXSession implements a simple shell session for RTX routers
type simpleRTXSession struct {
	client  *ssh.Client
	session *ssh.Session
	stdin   io.WriteCloser
	stdout  io.Reader
	reader  *bufio.Reader
	mu      sync.Mutex
	closed  bool
}

// newSimpleRTXSession creates a new simple shell session
func newSimpleRTXSession(client *ssh.Client) (*simpleRTXSession, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
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

	if err := session.Shell(); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to start shell: %w", err)
	}

	s := &simpleRTXSession{
		client:  client,
		session: session,
		stdin:   stdin,
		stdout:  stdout,
		reader:  bufio.NewReader(stdout),
	}

	// Read initial banner and prompt
	if err := s.waitForPrompt(); err != nil {
		s.Close()
		return nil, fmt.Errorf("failed to read initial prompt: %w", err)
	}

	// Set character encoding
	if _, err := s.sendCommand("console character en.ascii"); err != nil {
		s.Close()
		return nil, fmt.Errorf("failed to set encoding: %w", err)
	}

	return s, nil
}

// Send executes a command and returns the output
func (s *simpleRTXSession) Send(cmd string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, fmt.Errorf("session is closed")
	}

	log.Printf("[DEBUG] Sending RTX command: %s", cmd)
	
	output, err := s.sendCommand(cmd)
	if err != nil {
		log.Printf("[ERROR] RTX command failed: %v", err)
		return nil, err
	}
	
	log.Printf("[DEBUG] RTX command output length: %d", len(output))
	return output, nil
}

// sendCommand sends a command and waits for output
func (s *simpleRTXSession) sendCommand(cmd string) ([]byte, error) {
	// Send command
	if _, err := fmt.Fprintf(s.stdin, "%s\r\n", cmd); err != nil {
		return nil, fmt.Errorf("failed to write command: %w", err)
	}

	// Read until we see the command echo
	var output bytes.Buffer
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read: %w", err)
		}
		
		// Skip the command echo line
		if strings.TrimSpace(line) == cmd {
			break
		}
	}

	// Now read the actual output until the next prompt
	if err := s.readUntilPrompt(&output); err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}

// waitForPrompt waits for a prompt without capturing output
func (s *simpleRTXSession) waitForPrompt() error {
	var discard bytes.Buffer
	return s.readUntilPrompt(&discard)
}

// readUntilPrompt reads output until a prompt is found
func (s *simpleRTXSession) readUntilPrompt(output io.Writer) error {
	timeout := time.After(10 * time.Second)
	
	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for prompt")
		default:
			// Try to read a byte
			b, err := s.reader.ReadByte()
			if err != nil {
				if err == io.EOF {
					return fmt.Errorf("connection closed")
				}
				return fmt.Errorf("read error: %w", err)
			}

			// Write to output
			if output != nil {
				output.Write([]byte{b})
			}

			// Check for prompt character
			if b == '>' || b == '#' {
				// Peek at the next character
				nextBytes, err := s.reader.Peek(1)
				if err == nil && len(nextBytes) > 0 {
					if nextBytes[0] == ' ' {
						// This is a prompt, consume the space
						s.reader.ReadByte()
						if output != nil {
							output.Write([]byte{' '})
						}
						return nil
					}
				} else if err == io.EOF {
					// End of stream after prompt character
					return nil
				}
			}
		}
	}
}

// Close closes the session
func (s *simpleRTXSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true

	// Send exit command
	fmt.Fprintln(s.stdin, "exit")
	
	// Close the session
	if s.session != nil {
		return s.session.Close()
	}
	
	return nil
}