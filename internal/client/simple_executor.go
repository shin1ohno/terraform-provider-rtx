package client

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// simpleExecutor executes commands by creating a new SSH session for each command
type simpleExecutor struct {
	config         *ssh.ClientConfig
	addr           string
	promptDetector PromptDetector
	rtxConfig      *Config // RTX router configuration including admin password
}

// NewSimpleExecutor creates a new simple executor
func NewSimpleExecutor(config *ssh.ClientConfig, addr string, promptDetector PromptDetector, rtxConfig *Config) Executor {
	return &simpleExecutor{
		config:         config,
		addr:           addr,
		promptDetector: promptDetector,
		rtxConfig:      rtxConfig,
	}
}

// Run executes a command by creating a new SSH connection
func (e *simpleExecutor) Run(ctx context.Context, cmd string) ([]byte, error) {
	log.Printf("[DEBUG] SimpleExecutor: Running command: %s", cmd)

	// Create a new SSH connection for each command
	client, err := ssh.Dial("tcp", e.addr, e.config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	// Create a working session
	session, err := newWorkingSession(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Check if this command requires administrator privileges
	if e.requiresAdminPrivileges(cmd) {
		log.Printf("[DEBUG] SimpleExecutor: Command requires administrator privileges, authenticating...")
		if err := e.authenticateAsAdmin(session); err != nil {
			return nil, fmt.Errorf("failed to authenticate as administrator: %w", err)
		}

		// Mark session as being in administrator mode
		session.SetAdminMode(true)
	}

	// Execute the command
	output, err := session.Send(cmd)
	if err != nil {
		log.Printf("[DEBUG] SimpleExecutor: Command execution failed: %v", err)
		return nil, fmt.Errorf("command execution failed: %w", err)
	}

	// Check for "Administrator use only" error in output
	outputStr := string(output)
	if strings.Contains(outputStr, "Administrator use only") {
		log.Printf("[DEBUG] SimpleExecutor: Administrator use only error detected in output: %q", outputStr)
		return nil, fmt.Errorf("administrator privileges required for command: %s", cmd)
	}

	// Check for other error patterns in output
	if containsCommandError(outputStr) {
		log.Printf("[DEBUG] SimpleExecutor: Command error detected in output: %q", outputStr)
		return nil, fmt.Errorf("command failed: %s", outputStr)
	}

	// Check for prompt
	matched, prompt := e.promptDetector.DetectPrompt(output)
	if !matched {
		log.Printf("[DEBUG] SimpleExecutor: Prompt detection failed. Output: %q", string(output))
		return nil, fmt.Errorf("%w: output does not contain expected prompt", ErrPrompt)
	}
	log.Printf("[DEBUG] SimpleExecutor: Prompt detected: %q", prompt)

	return output, nil
}

// requiresAdminPrivileges checks if a command requires administrator privileges
func (e *simpleExecutor) requiresAdminPrivileges(cmd string) bool {
	adminCommands := []string{
		"dhcp scope bind",
		"dhcp scope unbind",
		"no dhcp scope bind",
		"show config",
		"show dhcp scope bind",
		// NOTE: "show environment" typically does NOT require admin privileges
		// Temporarily removing it to test if this was the cause of the hang
		// "show environment",
		"ip host",
		"no ip host",
		"ip route",
		"no ip route",
		"save", // Save configuration command requires admin privileges
	}

	cmdLower := strings.ToLower(strings.TrimSpace(cmd))

	// Check for DHCP scope creation, modification, and deletion commands
	// These commands start with "dhcp scope" followed by a number (scope ID)
	if strings.HasPrefix(cmdLower, "dhcp scope ") {
		// Extract the part after "dhcp scope "
		parts := strings.Fields(cmdLower)
		if len(parts) >= 3 {
			// Check if the third part is a number (scope ID)
			// This catches commands like "dhcp scope 1 ..." for creation
			// and "no dhcp scope 1" for deletion
			return true
		}
	}

	// Check for "no dhcp scope" commands for deletion
	if strings.HasPrefix(cmdLower, "no dhcp scope ") {
		return true
	}

	// Check for existing admin commands
	for _, adminCmd := range adminCommands {
		if strings.Contains(cmdLower, adminCmd) {
			return true
		}
	}

	return false
}

// authenticateAsAdmin authenticates as administrator using the administrator command
func (e *simpleExecutor) authenticateAsAdmin(session Session) error {
	log.Printf("[DEBUG] SimpleExecutor: Authenticating as administrator")

	// Cast session to workingSession to access low-level methods
	ws, ok := session.(*workingSession)
	if !ok {
		return fmt.Errorf("session type not supported for administrator authentication")
	}

	// Send administrator command and wait for password prompt
	if err := e.sendAdministratorCommand(ws); err != nil {
		return fmt.Errorf("failed to authenticate as administrator: %w", err)
	}

	log.Printf("[DEBUG] SimpleExecutor: Administrator authentication completed")
	return nil
}

// sendAdministratorCommand sends the administrator command and handles password prompt
func (e *simpleExecutor) sendAdministratorCommand(ws *workingSession) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return fmt.Errorf("session is closed")
	}

	log.Printf("[DEBUG] SimpleExecutor: Sending administrator command")

	// Send administrator command
	if _, err := fmt.Fprintf(ws.stdin, "administrator\r"); err != nil {
		return fmt.Errorf("failed to send administrator command: %w", err)
	}

	// Read until we get password prompt
	passwordPrompt, err := ws.readUntilString("Password:", 10*time.Second)
	if err != nil {
		// Log the received content for debugging
		log.Printf("[DEBUG] SimpleExecutor: Failed to get password prompt, received: %q", string(passwordPrompt))
		return fmt.Errorf("failed to get password prompt: %w", err)
	}
	log.Printf("[DEBUG] SimpleExecutor: Password prompt received successfully: %q", string(passwordPrompt))

	// Send password
	log.Printf("[DEBUG] SimpleExecutor: Sending administrator password")
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", e.rtxConfig.AdminPassword); err != nil {
		return fmt.Errorf("failed to send password: %w", err)
	}

	// Read response after password - look for administrator prompt (# instead of >)
	response, err := ws.readUntilPrompt(15 * time.Second) // Increased timeout
	if err != nil {
		log.Printf("[DEBUG] SimpleExecutor: Failed to read password response: %v", err)
		return fmt.Errorf("failed to read password response: %w", err)
	}

	responseStr := string(response)
	log.Printf("[DEBUG] SimpleExecutor: Password response received: %q", responseStr)

	// Check for authentication failure patterns
	responseStrLower := strings.ToLower(responseStr)
	if strings.Contains(responseStrLower, "incorrect") ||
		strings.Contains(responseStrLower, "failed") ||
		strings.Contains(responseStrLower, "invalid") ||
		strings.Contains(responseStrLower, "authentication failed") ||
		strings.Contains(responseStrLower, "access denied") {
		return fmt.Errorf("administrator authentication failed: %s", responseStr)
	}

	// Check for successful authentication (admin prompt with #)
	if strings.Contains(responseStr, "# ") || strings.HasSuffix(strings.TrimSpace(responseStr), "#") {
		log.Printf("[DEBUG] SimpleExecutor: Administrator authentication successful - admin prompt detected")
		return nil
	}

	// If we reach here, we got a response but couldn't confirm success or failure
	// Log this for debugging but continue - some routers may not show clear success indicators
	log.Printf("[DEBUG] SimpleExecutor: Administrator authentication completed - prompt detection unclear")
	return nil
}

// containsCommandError checks if the output contains a command error
func containsCommandError(output string) bool {
	// RTX router error patterns
	errorPatterns := []string{
		"Error:",
		"% Error:",
		"Command failed:",
		"Invalid parameter",
		"Permission denied",
		"Connection timeout",
		"Syntax error",
		"Unknown command",
		"Parameter error",
		"Configuration error",
	}

	outputLower := strings.ToLower(output)
	for _, pattern := range errorPatterns {
		if strings.Contains(outputLower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}
