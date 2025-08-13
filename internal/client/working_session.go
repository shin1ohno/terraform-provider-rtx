package client

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// workingSession implements a working SSH session for RTX routers
// This is based on the successful expect script test
type workingSession struct {
	client   *ssh.Client
	session  *ssh.Session
	stdin    io.WriteCloser
	stdout   io.Reader
	mu       sync.Mutex
	closed   bool
}

// newWorkingSession creates a new working session
func newWorkingSession(client *ssh.Client) (*workingSession, error) {
	log.Printf("[DEBUG] Creating new working session")
	
	// Create session first
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Get pipes BEFORE requesting PTY or starting shell
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

	// Request PTY - same as working expect script
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("vt100", 80, 40, modes); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to request PTY: %w", err)
	}

	// Start shell
	if err := session.Shell(); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to start shell: %w", err)
	}

	s := &workingSession{
		client:  client,
		session: session,
		stdin:   stdin,
		stdout:  stdout,
	}

	// Wait for initial prompt
	log.Printf("[DEBUG] Waiting for initial prompt")
	initialOutput, err := s.readUntilPrompt(10 * time.Second)
	if err != nil {
		s.Close()
		return nil, fmt.Errorf("failed to get initial prompt: %w", err)
	}
	log.Printf("[DEBUG] Got initial output: %d bytes", len(initialOutput))
	log.Printf("[DEBUG] Initial output content: %q", string(initialOutput))

	// Optional: Set character encoding (some routers don't support this)
	log.Printf("[DEBUG] Setting character encoding")
	if _, err := s.executeCommand("console character en.ascii", 5*time.Second); err != nil {
		log.Printf("[WARN] Failed to set character encoding: %v (continuing anyway)", err)
	}

	return s, nil
}

// Send executes a command and returns the output
func (s *workingSession) Send(cmd string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[DEBUG] workingSession.Send called with command: %s, closed: %v", cmd, s.closed)
	
	if s.closed {
		return nil, fmt.Errorf("session is closed")
	}

	// The executor expects the raw output including the prompt
	// So we return the raw output without cleaning
	// Use longer timeout for commands that may have large output like "show status dhcp"
	timeout := 30 * time.Second
	if strings.Contains(cmd, "show status dhcp") {
		timeout = 60 * time.Second
	}
	output, err := s.executeCommandRaw(cmd, timeout)
	if err != nil {
		log.Printf("[ERROR] workingSession.Send failed: %v", err)
		return nil, err
	}

	return output, nil
}

// executeCommand sends command and reads response (cleaned)
func (s *workingSession) executeCommand(cmd string, timeout time.Duration) ([]byte, error) {
	log.Printf("[DEBUG] Executing command: %s", cmd)

	// Send command with carriage return (like expect script)
	if _, err := fmt.Fprintf(s.stdin, "%s\r", cmd); err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Read response until prompt
	output, err := s.readUntilPrompt(timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Clean output - remove command echo and prompt
	cleanOutput := s.cleanOutput(string(output), cmd)
	
	log.Printf("[DEBUG] Command completed, output length: %d bytes", len(cleanOutput))
	return []byte(cleanOutput), nil
}

// executeCommandRaw sends command and returns raw response including prompt
func (s *workingSession) executeCommandRaw(cmd string, timeout time.Duration) ([]byte, error) {
	log.Printf("[DEBUG] Executing command (raw): %s", cmd)

	// Send command with carriage return
	if _, err := fmt.Fprintf(s.stdin, "%s\r", cmd); err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Read response until prompt
	output, err := s.readUntilPrompt(timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("[DEBUG] Command completed (raw), output length: %d bytes", len(output))
	return output, nil
}

// readUntilPrompt reads until we see a prompt character
func (s *workingSession) readUntilPrompt(timeout time.Duration) ([]byte, error) {
	var buffer bytes.Buffer
	deadline := time.Now().Add(timeout)
	buf := make([]byte, 1)

	for {
		if time.Now().After(deadline) {
			return buffer.Bytes(), fmt.Errorf("timeout waiting for prompt")
		}

		// Read one byte at a time
		n, err := s.stdout.Read(buf)
		if err != nil && err != io.EOF {
			return buffer.Bytes(), fmt.Errorf("read error: %w", err)
		}

		if n > 0 {
			buffer.WriteByte(buf[0])
			
			// Check if we have a prompt
			content := buffer.String()
			lines := strings.Split(content, "\n")
			if len(lines) > 0 {
				lastLine := lines[len(lines)-1]
				// Check for prompt at end of line (> or # with optional space after)
				// Also check for RTX format like "[RTX1210] >"
				if len(lastLine) > 0 {
					trimmed := strings.TrimSpace(lastLine)
					if strings.HasSuffix(trimmed, ">") || strings.HasSuffix(trimmed, "#") {
						// Also check if it looks like RTX prompt format
						if strings.Contains(lastLine, "] >") || strings.HasSuffix(lastLine, "> ") {
							// Found prompt
							return buffer.Bytes(), nil
						}
					}
				}
			}
		}

		// Small delay to avoid busy loop
		time.Sleep(10 * time.Millisecond)
	}
}

// cleanOutput removes command echo and prompt from output
func (s *workingSession) cleanOutput(output, cmd string) string {
	lines := strings.Split(output, "\n")
	
	// Remove command echo (first line that matches command)
	for i, line := range lines {
		if strings.TrimSpace(line) == cmd {
			if i < len(lines)-1 {
				lines = lines[i+1:]
			}
			break
		}
	}

	// Remove prompt from last line
	if len(lines) > 0 {
		lastIdx := len(lines) - 1
		lastLine := lines[lastIdx]
		if strings.Contains(lastLine, ">") || strings.Contains(lastLine, "#") {
			// Remove the line if it's just a prompt
			trimmed := strings.TrimSpace(lastLine)
			if trimmed == ">" || trimmed == "#" || strings.HasPrefix(trimmed, "[") {
				lines = lines[:lastIdx]
			}
		}
	}

	// Join lines back and trim
	result := strings.Join(lines, "\n")
	return strings.TrimSpace(result)
}

// Close closes the session
func (s *workingSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[DEBUG] workingSession.Close() called, already closed: %v", s.closed)
	
	if s.closed {
		return nil
	}

	s.closed = true

	// Send exit command
	fmt.Fprintln(s.stdin, "exit")
	time.Sleep(100 * time.Millisecond)

	// Close session
	if s.session != nil {
		log.Printf("[DEBUG] Closing SSH session")
		return s.session.Close()
	}

	return nil
}