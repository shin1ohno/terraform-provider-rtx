package client

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/sh1/terraform-provider-rtx/internal/logging"
)

// workingSession implements a working SSH session for RTX routers
// This is based on the successful expect script test
type workingSession struct {
	client    *ssh.Client
	session   *ssh.Session
	stdin     io.WriteCloser
	stdout    io.Reader
	mu        sync.Mutex
	closed    bool
	adminMode bool // Track if we're in administrator mode
}

// newWorkingSession creates a new working session
func newWorkingSession(client *ssh.Client) (*workingSession, error) {
	logger := logging.Global()
	logger.Debug().Msg("Creating new working session")

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

	// Use wide terminal to prevent line wrapping for long filter lists
	// RTX config lines can exceed 200 characters (e.g., secure filter with 13+ IDs)
	// RequestPty parameters: term, height, width, modes
	if err := session.RequestPty("vt100", 40, 512, modes); err != nil {
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
	logger.Debug().Msg("Waiting for initial prompt")
	initialOutput, err := s.readUntilPrompt(10 * time.Second)
	if err != nil {
		s.Close()
		return nil, fmt.Errorf("failed to get initial prompt: %w", err)
	}
	logger.Debug().Int("bytes", len(initialOutput)).Msg("Got initial output")
	logger.Debug().Str("content", string(initialOutput)).Msg("Initial output content")

	// Optional: Set character encoding (some routers don't support this)
	logger.Debug().Msg("Setting character encoding")
	if _, err := s.executeCommand("console character en.ascii", 5*time.Second); err != nil {
		logger.Warn().Err(err).Msg("Failed to set character encoding (continuing anyway)")
	}

	// Disable paging to get full output from commands like "show config"
	logger.Debug().Msg("Disabling console paging")
	if _, err := s.executeCommand("console lines infinity", 5*time.Second); err != nil {
		logger.Warn().Err(err).Msg("Failed to disable paging (continuing anyway)")
	}

	return s, nil
}

// Send executes a command and returns the output
func (s *workingSession) Send(cmd string) ([]byte, error) {
	logger := logging.Global()
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.Debug().Str("command", SanitizeCommandForLog(cmd)).Bool("closed", s.closed).Msg("workingSession.Send called")

	if s.closed {
		return nil, fmt.Errorf("session is closed")
	}

	// The executor expects the raw output including the prompt
	// So we return the raw output without cleaning
	// Use reasonable timeout for commands
	timeout := 15 * time.Second
	if strings.Contains(cmd, "show config") {
		timeout = 120 * time.Second // show config produces large output
	} else if strings.Contains(cmd, "show status dhcp") {
		timeout = 30 * time.Second
	} else if strings.Contains(cmd, "show environment") {
		timeout = 20 * time.Second
	}
	output, err := s.executeCommandRaw(cmd, timeout)
	if err != nil {
		logger.Error().Err(err).Msg("workingSession.Send failed")
		return nil, err
	}

	return output, nil
}

// executeCommand sends command and reads response (cleaned)
func (s *workingSession) executeCommand(cmd string, timeout time.Duration) ([]byte, error) {
	logger := logging.Global()
	logger.Debug().Str("command", cmd).Msg("Executing command")

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

	logger.Debug().Int("bytes", len(cleanOutput)).Msg("Command completed")
	return []byte(cleanOutput), nil
}

// executeCommandRaw sends command and returns raw response including prompt
func (s *workingSession) executeCommandRaw(cmd string, timeout time.Duration) ([]byte, error) {
	logger := logging.Global()
	logger.Debug().Str("command", cmd).Msg("Executing command (raw)")

	// Send command with carriage return
	if _, err := fmt.Fprintf(s.stdin, "%s\r", cmd); err != nil {
		return nil, fmt.Errorf("failed to send command: %w", err)
	}

	// Read response until prompt
	output, err := s.readUntilPrompt(timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	logger.Debug().Int("bytes", len(output)).Msg("Command completed (raw)")
	return output, nil
}

// readUntilPrompt reads until we see a prompt character
func (s *workingSession) readUntilPrompt(timeout time.Duration) ([]byte, error) {
	logger := logging.Global()
	var buffer bytes.Buffer
	deadline := time.Now().Add(timeout)
	buf := make([]byte, 1)

	for {
		if time.Now().After(deadline) {
			logger.Debug().Str("buffer", buffer.String()).Msg("readUntilPrompt: Timeout waiting for prompt")
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
				// Detect RTX prompt generically without depending on hostname
				// Conditions:
				// 1. Line is short (prompts are typically < 100 chars)
				// 2. Line ends with "> " or "# " (with trailing space)
				// 3. Line doesn't start with whitespace (config content is often indented)
				if len(lastLine) > 0 && len(lastLine) < 100 {
					trimmedLeft := strings.TrimLeft(lastLine, "\r")
					// Skip if line starts with # (config comment line, not a prompt)
					// RTX config comments start with "#", but prompts end with "# "
					if strings.HasPrefix(trimmedLeft, "#") {
						continue
					}
					// Skip if line starts with whitespace (indented config content)
					if strings.HasPrefix(trimmedLeft, " ") || strings.HasPrefix(trimmedLeft, "\t") {
						continue
					}
					// Check for user mode prompt ending with "> "
					if strings.HasSuffix(lastLine, "> ") {
						return buffer.Bytes(), nil
					}
					// Check for admin mode prompt ending with "# "
					if strings.HasSuffix(lastLine, "# ") {
						return buffer.Bytes(), nil
					}
					// Also check without trailing space (some terminals)
					// Require minimum length to avoid matching single "#" or ">"
					if len(trimmedLeft) >= 3 &&
						(strings.HasSuffix(lastLine, ">") || strings.HasSuffix(lastLine, "#")) {
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
			_ = conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
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
	logger := logging.Global()
	s.mu.Lock()
	defer s.mu.Unlock()

	logger.Debug().Bool("already_closed", s.closed).Msg("workingSession.Close() called")

	if s.closed {
		return nil
	}

	s.closed = true

	// Send appropriate exit commands based on current mode
	if s.adminMode {
		logger.Debug().Msg("Session is in administrator mode, sending two exit commands")
		// First exit: leave administrator mode (back to user mode)
		if err := s.exitAdminMode(); err != nil {
			logger.Warn().Err(err).Msg("Failed to exit administrator mode properly")
		}
		s.adminMode = false

		// Small delay before second exit
		time.Sleep(500 * time.Millisecond)

		// Second exit: disconnect from router
		if _, err := fmt.Fprintf(s.stdin, "exit\r"); err != nil {
			logger.Warn().Err(err).Msg("Failed to send second exit command")
		}
		time.Sleep(300 * time.Millisecond)
	} else {
		logger.Debug().Msg("Session is in user mode, sending one exit command")
		if _, err := fmt.Fprintf(s.stdin, "exit\r"); err != nil {
			logger.Warn().Err(err).Msg("Failed to send exit command")
		}
		time.Sleep(300 * time.Millisecond)
	}

	// Close session
	if s.session != nil {
		logger.Debug().Msg("Closing SSH session")
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
	logger := logging.Global()
	logger.Debug().Msg("Exiting administrator mode")

	// Send exit command
	if _, err := fmt.Fprintf(s.stdin, "exit\r"); err != nil {
		return fmt.Errorf("failed to send exit command: %w", err)
	}

	// Read response and check for configuration save prompt
	response, err := s.readUntilPromptOrSaveConfirmation(5 * time.Second)
	if err != nil {
		logger.Warn().Err(err).Msg("Error reading response after exit")
		return err
	}

	responseStr := string(response)
	logger.Debug().Str("response", responseStr).Msg("Exit response")

	// Check if we got a configuration save confirmation prompt
	if s.isSaveConfigurationPrompt(responseStr) {
		logger.Debug().Msg("Configuration save prompt detected, responding with 'N'")
		// Respond with 'N' to not save configuration
		if _, err := fmt.Fprintf(s.stdin, "N\r"); err != nil {
			return fmt.Errorf("failed to respond to save prompt: %w", err)
		}

		// Read final response after save confirmation
		finalResponse, err := s.readUntilPrompt(3 * time.Second)
		if err != nil {
			logger.Warn().Err(err).Msg("Error reading final response")
			return err
		}
		logger.Debug().Str("response", string(finalResponse)).Msg("Final exit response")
	}

	return nil
}

// readUntilPromptOrSaveConfirmation reads until we see a prompt or save confirmation
func (s *workingSession) readUntilPromptOrSaveConfirmation(timeout time.Duration) ([]byte, error) {
	logger := logging.Global()
	var buffer bytes.Buffer
	deadline := time.Now().Add(timeout)
	buf := make([]byte, 1)

	for {
		if time.Now().After(deadline) {
			logger.Debug().Str("buffer", buffer.String()).Msg("readUntilPromptOrSaveConfirmation: Timeout")
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
