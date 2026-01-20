package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// rtxTerminalSession implements a terminal session for RTX routers
// Based on Ansible RTX collection implementation
type rtxTerminalSession struct {
	client       *ssh.Client
	session      *ssh.Session
	stdin        io.WriteCloser
	stdout       io.Reader
	stderr       io.Reader
	reader       *bufio.Reader
	mu           sync.Mutex
	closed       bool
	promptRegex  *regexp.Regexp
	errorRegex   *regexp.Regexp
}

// newRTXTerminalSession creates a new terminal session for RTX router
func newRTXTerminalSession(client *ssh.Client) (*rtxTerminalSession, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Request PTY - RTX requires terminal
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

	// Get pipes
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

	stderr, err := session.StderrPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to get stderr: %w", err)
	}

	// Try to start shell with retry
	log.Printf("[DEBUG] Starting shell session")
	
	// Some RTX routers may need a small delay after PTY request
	time.Sleep(100 * time.Millisecond)
	
	if err := session.Shell(); err != nil {
		log.Printf("[ERROR] Failed to start shell: %v", err)
		
		// Try without stderr pipe
		session.Close()
		session2, err2 := client.NewSession()
		if err2 != nil {
			return nil, fmt.Errorf("failed to create second session: %w", err2)
		}
		
		// Use wide terminal (same as above)
		if err2 := session2.RequestPty("vt100", 40, 512, modes); err2 != nil {
			session2.Close()
			return nil, fmt.Errorf("failed to request PTY on second attempt: %w", err2)
		}
		
		stdin2, _ := session2.StdinPipe()
		stdout2, _ := session2.StdoutPipe()
		
		if err2 := session2.Shell(); err2 != nil {
			session2.Close()
			return nil, fmt.Errorf("failed to start shell on second attempt: %w (first error: %w)", err2, err)
		}
		
		log.Printf("[DEBUG] Shell session started on second attempt")
		session = session2
		stdin = stdin2
		stdout = stdout2
		stderr = nil
	} else {
		log.Printf("[DEBUG] Shell session started successfully")
	}

	// Create session object
	s := &rtxTerminalSession{
		client:      client,
		session:     session,
		stdin:       stdin,
		stdout:      stdout,
		stderr:      stderr,
		reader:      bufio.NewReader(stdout),
		promptRegex: regexp.MustCompile(`[>#]\s*$`),     // Matches RTX prompts
		errorRegex:  regexp.MustCompile(`(?i)Error:\s*`), // Matches error messages
	}

	// Wait for initial prompt
	log.Printf("[DEBUG] Waiting for initial RTX prompt")
	initialOutput, err := s.readUntilPrompt(10 * time.Second)
	if err != nil {
		s.Close()
		return nil, fmt.Errorf("failed to read initial prompt: %w", err)
	}
	log.Printf("[DEBUG] Initial output: %q", string(initialOutput))

	// Set character encoding for compatibility
	log.Printf("[DEBUG] Setting character encoding")
	if _, err := s.executeCommand("console character en.ascii", 5*time.Second); err != nil {
		// Non-fatal error - some RTX models might not support this
		log.Printf("[WARN] Failed to set character encoding: %v", err)
	}

	return s, nil
}

// Send executes a command and returns the output
func (s *rtxTerminalSession) Send(cmd string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, fmt.Errorf("session is closed")
	}

	// Use a reasonable timeout for commands
	output, err := s.executeCommand(cmd, 30*time.Second)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// executeCommand sends a command and waits for response
func (s *rtxTerminalSession) executeCommand(cmd string, timeout time.Duration) ([]byte, error) {
	log.Printf("[DEBUG] Executing RTX command: %s", cmd)

	// Send command
	if _, err := fmt.Fprintf(s.stdin, "%s\n", cmd); err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Read response
	output, err := s.readUntilPrompt(timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Remove command echo from output
	lines := bytes.Split(output, []byte("\n"))
	if len(lines) > 0 {
		firstLine := string(lines[0])
		// Remove the command echo line
		if strings.TrimSpace(firstLine) == cmd {
			if len(lines) > 1 {
				output = bytes.Join(lines[1:], []byte("\n"))
			} else {
				output = []byte{}
			}
		}
	}

	// Check for errors in output
	if s.errorRegex.Match(output) {
		return output, fmt.Errorf("command returned error: %s", string(output))
	}

	log.Printf("[DEBUG] Command output length: %d bytes", len(output))
	return output, nil
}

// readUntilPrompt reads output until a prompt is detected
func (s *rtxTerminalSession) readUntilPrompt(timeout time.Duration) ([]byte, error) {
	var output bytes.Buffer
	buffer := make([]byte, 4096)
	deadline := time.Now().Add(timeout)

	for {
		// Check timeout
		if time.Now().After(deadline) {
			return output.Bytes(), fmt.Errorf("timeout waiting for prompt")
		}

		// Set read deadline
		s.reader.Reset(s.stdout) // Reset to clear any buffered data
		
		// Try to read available data
		n, err := s.reader.Read(buffer)
		if err != nil && err != io.EOF {
			return output.Bytes(), fmt.Errorf("read error: %w", err)
		}

		if n > 0 {
			output.Write(buffer[:n])
			
			// Check if we have a prompt in the output
			currentOutput := output.String()
			if s.promptRegex.MatchString(currentOutput) {
				log.Printf("[DEBUG] Found prompt in output")
				return output.Bytes(), nil
			}
		}

		// Small delay to avoid busy waiting
		time.Sleep(10 * time.Millisecond)
	}
}

// Close closes the session
func (s *rtxTerminalSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true

	// Try to send exit command
	fmt.Fprintln(s.stdin, "exit")
	time.Sleep(100 * time.Millisecond)

	// Close the session
	if s.session != nil {
		return s.session.Close()
	}

	return nil
}