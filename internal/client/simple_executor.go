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
		return nil, fmt.Errorf("command execution failed: %w", err)
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
	}
	
	cmdLower := strings.ToLower(strings.TrimSpace(cmd))
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
		return fmt.Errorf("failed to get password prompt: %w", err)
	}
	log.Printf("[DEBUG] SimpleExecutor: Password prompt received: %q", string(passwordPrompt))
	
	// Send password
	log.Printf("[DEBUG] SimpleExecutor: Sending administrator password")
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", e.rtxConfig.AdminPassword); err != nil {
		return fmt.Errorf("failed to send password: %w", err)
	}
	
	// Read response after password - look for administrator prompt (# instead of >)
	response, err := ws.readUntilPrompt(10 * time.Second)
	if err != nil {
		return fmt.Errorf("failed to read password response: %w", err)
	}
	
	responseStr := string(response)
	log.Printf("[DEBUG] SimpleExecutor: Password response: %q", responseStr)
	
	// Check for authentication success (look for # prompt or no error message)
	if strings.Contains(responseStr, "incorrect") || strings.Contains(responseStr, "failed") || strings.Contains(responseStr, "Invalid") {
		return fmt.Errorf("administrator authentication failed: %s", responseStr)
	}
	
	log.Printf("[DEBUG] SimpleExecutor: Administrator authentication successful")
	return nil
}