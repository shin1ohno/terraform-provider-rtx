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
	client     *ssh.Client
	session    *ssh.Session
	stdin      io.WriteCloser
	stdout     io.Reader
	mu         sync.Mutex
	closed     bool
	adminMode  bool // Track if we're in administrator mode
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
	// Use reasonable timeout for commands
	timeout := 15 * time.Second
	if strings.Contains(cmd, "show status dhcp") {
		timeout = 30 * time.Second
	} else if strings.Contains(cmd, "show environment") {
		timeout = 20 * time.Second // Reduced from 60s to 20s
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
			log.Printf("[DEBUG] readUntilPrompt: Timeout waiting for prompt. Buffer content: %q", buffer.String())
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
				// RTX format: "[RTX1210] >" for user mode or "[RTX1210] # " for admin mode
				if len(lastLine) > 0 {
					trimmed := strings.TrimSpace(lastLine)
					// Check for user mode prompt: "[RTX1210] >"
					if strings.Contains(lastLine, "] >") || strings.HasSuffix(lastLine, "> ") {
						return buffer.Bytes(), nil
					}
					// Check for admin mode prompt: "[RTX1210] # "
					if strings.Contains(lastLine, "] # ") || strings.HasSuffix(lastLine, "# ") {
						return buffer.Bytes(), nil
					}
					// Fallback: check if line ends with > or # (with possible trailing spaces)
					if strings.HasSuffix(trimmed, ">") || strings.HasSuffix(trimmed, "#") {
						return buffer.Bytes(), nil
					}
				}
			}
		}

		// Small delay to avoid busy loop
		time.Sleep(10 * time.Millisecond)
	}
}

// readUntilString reads from stdout until the specified string appears
func (s *workingSession) readUntilString(target string, timeout time.Duration) ([]byte, error) {
	var buffer bytes.Buffer
	start := time.Now()

	for time.Since(start) < timeout {
		// Read some data
		chunk := make([]byte, 1024)
		
		// Set read deadline if possible
		if conn, ok := s.stdout.(interface{ SetReadDeadline(time.Time) error }); ok {
			conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		}
		
		n, err := s.stdout.Read(chunk)
		if err != nil && err != io.EOF {
			// Timeout is expected, continue
			if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline") {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			return buffer.Bytes(), fmt.Errorf("read error: %w", err)
		}

		if n > 0 {
			buffer.Write(chunk[:n])
			// Check if target string appears in buffer
			if strings.Contains(buffer.String(), target) {
				return buffer.Bytes(), nil
			}
		}

		// Small delay to avoid busy loop
		time.Sleep(10 * time.Millisecond)
	}

	return buffer.Bytes(), fmt.Errorf("timeout waiting for %q (got: %q)", target, buffer.String())
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

	// Send appropriate exit commands based on current mode
	if s.adminMode {
		log.Printf("[DEBUG] Session is in administrator mode, sending two exit commands")
		// First exit: leave administrator mode (back to user mode)
		if err := s.exitAdminMode(); err != nil {
			log.Printf("[WARN] Failed to exit administrator mode properly: %v", err)
		}
		s.adminMode = false
		
		// Small delay before second exit
		time.Sleep(500 * time.Millisecond)
		
		// Second exit: disconnect from router
		if _, err := fmt.Fprintf(s.stdin, "exit\r"); err != nil {
			log.Printf("[WARN] Failed to send second exit command: %v", err)
		}
		time.Sleep(300 * time.Millisecond)
	} else {
		log.Printf("[DEBUG] Session is in user mode, sending one exit command")
		if _, err := fmt.Fprintf(s.stdin, "exit\r"); err != nil {
			log.Printf("[WARN] Failed to send exit command: %v", err)
		}
		time.Sleep(300 * time.Millisecond)
	}

	// Close session
	if s.session != nil {
		log.Printf("[DEBUG] Closing SSH session")
		return s.session.Close()
	}

	return nil
}

// SetAdminMode sets the administrator mode flag
func (s *workingSession) SetAdminMode(admin bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.adminMode = admin
}

// exitAdminMode safely exits administrator mode handling configuration save prompts
func (s *workingSession) exitAdminMode() error {
	log.Printf("[DEBUG] Exiting administrator mode")
	
	// Send exit command
	if _, err := fmt.Fprintf(s.stdin, "exit\r"); err != nil {
		return fmt.Errorf("failed to send exit command: %w", err)
	}
	
	// Read response and check for configuration save prompt
	response, err := s.readUntilPromptOrSaveConfirmation(5 * time.Second)
	if err != nil {
		log.Printf("[WARN] Error reading response after exit: %v", err)
		return err
	}
	
	responseStr := string(response)
	log.Printf("[DEBUG] Exit response: %q", responseStr)
	
	// Check if we got a configuration save confirmation prompt
	if s.isSaveConfigurationPrompt(responseStr) {
		log.Printf("[DEBUG] Configuration save prompt detected, responding with 'N'")
		// Respond with 'N' to not save configuration
		if _, err := fmt.Fprintf(s.stdin, "N\r"); err != nil {
			return fmt.Errorf("failed to respond to save prompt: %w", err)
		}
		
		// Read final response after save confirmation
		finalResponse, err := s.readUntilPrompt(3 * time.Second)
		if err != nil {
			log.Printf("[WARN] Error reading final response: %v", err)
			return err
		}
		log.Printf("[DEBUG] Final exit response: %q", string(finalResponse))
	}
	
	return nil
}

// readUntilPromptOrSaveConfirmation reads until we see a prompt or save confirmation
func (s *workingSession) readUntilPromptOrSaveConfirmation(timeout time.Duration) ([]byte, error) {
	var buffer bytes.Buffer
	deadline := time.Now().Add(timeout)
	buf := make([]byte, 1)

	for {
		if time.Now().After(deadline) {
			log.Printf("[DEBUG] readUntilPromptOrSaveConfirmation: Timeout. Buffer content: %q", buffer.String())
			return buffer.Bytes(), fmt.Errorf("timeout waiting for prompt or save confirmation")
		}

		// Read one byte at a time
		n, err := s.stdout.Read(buf)
		if err != nil && err != io.EOF {
			return buffer.Bytes(), fmt.Errorf("read error: %w", err)
		}

		if n > 0 {
			buffer.WriteByte(buf[0])
			
			content := buffer.String()
			
			// Check for save configuration prompt
			if s.isSaveConfigurationPrompt(content) {
				return buffer.Bytes(), nil
			}
			
			// Check for normal prompt (user mode or admin mode)
			lines := strings.Split(content, "\n")
			if len(lines) > 0 {
				lastLine := lines[len(lines)-1]
				if len(lastLine) > 0 {
					trimmed := strings.TrimSpace(lastLine)
					// Check for user mode prompt: "[RTX1210] >"
					if strings.Contains(lastLine, "] >") || strings.HasSuffix(lastLine, "> ") {
						return buffer.Bytes(), nil
					}
					// Check for admin mode prompt: "[RTX1210] # "
					if strings.Contains(lastLine, "] # ") || strings.HasSuffix(lastLine, "# ") {
						return buffer.Bytes(), nil
					}
					// Fallback: check if line ends with > or # (with possible trailing spaces)
					if strings.HasSuffix(trimmed, ">") || strings.HasSuffix(trimmed, "#") {
						return buffer.Bytes(), nil
					}
				}
			}
		}

		// Small delay to avoid busy loop
		time.Sleep(10 * time.Millisecond)
	}
}

// isSaveConfigurationPrompt checks if the text contains a configuration save prompt
func (s *workingSession) isSaveConfigurationPrompt(text string) bool {
	lowerText := strings.ToLower(text)
	
	// Common RTX router save configuration prompts
	savePrompts := []string{
		"save configuration?",
		"設定を保存しますか",
		"save config?",
		"(y/n)",
		"(y/n):",
		"(yes/no)",
		"save changes?",
		"保存しますか",
	}
	
	for _, prompt := range savePrompts {
		if strings.Contains(lowerText, prompt) {
			return true
		}
	}
	
	return false
}