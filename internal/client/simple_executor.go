package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/sh1/terraform-provider-rtx/internal/logging"
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
	logger := logging.FromContext(ctx)

	// Log command with resource context if available
	logEvent := logger.Info().Str("command", logging.SanitizeString(cmd))
	if res := logging.ResourceFromContext(ctx); res != nil {
		logEvent = logEvent.Str("resource", res.Type)
		if res.ID != "" {
			logEvent = logEvent.Str("id", res.ID)
		}
	}
	logEvent.Msg("RTX command")

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
		logger.Debug().Msg("SimpleExecutor: Command requires administrator privileges, authenticating...")
		if err := e.authenticateAsAdmin(ctx, session); err != nil {
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
		logger.Debug().Str("output", string(output)).Msg("SimpleExecutor: Prompt detection failed")
		return nil, fmt.Errorf("%w: output does not contain expected prompt", ErrPrompt)
	}
	logger.Debug().Str("prompt", prompt).Msg("SimpleExecutor: Prompt detected")

	return output, nil
}

// requiresAdminPrivileges checks if a command requires administrator privileges.
// Read-only commands (show, console) do not require admin privileges.
// Configuration commands require admin authentication when password is configured.
func (e *simpleExecutor) requiresAdminPrivileges(cmd string) bool {
	hasConfig := e.rtxConfig != nil
	hasPassword := hasConfig && e.rtxConfig.AdminPassword != ""
	if !hasPassword {
		return false
	}

	// Normalize command for checking
	cmdLower := strings.ToLower(strings.TrimSpace(cmd))

	// Read-only commands do not require admin privileges
	readOnlyPrefixes := []string{
		"show ",    // show commands (show config, show status, show sshd host key, etc.)
		"console ", // console display commands
		"less ",    // pager commands
	}
	for _, prefix := range readOnlyPrefixes {
		if strings.HasPrefix(cmdLower, prefix) {
			logging.Global().Debug().
				Str("command", cmd).
				Msg("SimpleExecutor: read-only command, no admin required")
			return false
		}
	}

	// All other commands require admin when password is configured
	logging.Global().Debug().
		Str("command", cmd).
		Msg("SimpleExecutor: admin required for this command")
	return true
}

// authenticateAsAdmin authenticates as administrator using the administrator command
func (e *simpleExecutor) authenticateAsAdmin(ctx context.Context, session Session) error {
	logger := logging.FromContext(ctx)
	logger.Debug().Msg("SimpleExecutor: Authenticating as administrator")

	// Cast session to workingSession to access low-level methods
	ws, ok := session.(*workingSession)
	if !ok {
		return fmt.Errorf("session type not supported for administrator authentication")
	}

	// Send administrator command and wait for password prompt
	if err := e.sendAdministratorCommand(ctx, ws); err != nil {
		return fmt.Errorf("failed to authenticate as administrator: %w", err)
	}

	logger.Debug().Msg("SimpleExecutor: Administrator authentication completed")
	return nil
}

// sendAdministratorCommand sends the administrator command and handles password prompt
func (e *simpleExecutor) sendAdministratorCommand(ctx context.Context, ws *workingSession) error {
	logger := logging.FromContext(ctx)
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return fmt.Errorf("session is closed")
	}

	logger.Debug().Msg("SimpleExecutor: Sending administrator command")

	// Send administrator command
	if _, err := fmt.Fprintf(ws.stdin, "administrator\r"); err != nil {
		return fmt.Errorf("failed to send administrator command: %w", err)
	}

	// Read until we get password prompt
	_, err := ws.readUntilString("Password:", 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get password prompt: %w", err)
	}
	logger.Debug().Msg("SimpleExecutor: Password prompt received")

	// Send password
	logger.Debug().Msg("SimpleExecutor: Sending administrator password")
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", e.rtxConfig.AdminPassword); err != nil {
		return fmt.Errorf("failed to send password: %w", err)
	}

	// Read response after password - look for administrator prompt (# instead of >)
	response, err := ws.readUntilPrompt(10 * time.Second)
	if err != nil {
		return fmt.Errorf("failed to read password response: %w", err)
	}

	responseStr := string(response)
	logger.Debug().Str("response", responseStr).Msg("SimpleExecutor: Password authentication response received")

	// Check for authentication failure
	if strings.Contains(responseStr, "incorrect") || strings.Contains(responseStr, "failed") || strings.Contains(responseStr, "Invalid") ||
		strings.Contains(responseStr, "エラー") || strings.Contains(responseStr, "パスワードが違います") {
		return fmt.Errorf("administrator authentication failed: %s", responseStr)
	}

	// Verify we actually got the admin prompt (#) not user prompt (>)
	if !strings.Contains(responseStr, "#") {
		return fmt.Errorf("administrator authentication failed: did not get admin prompt (#), got: %s", responseStr)
	}

	logger.Debug().Msg("SimpleExecutor: Administrator authentication successful")
	return nil
}

// RunBatch executes multiple commands and returns the combined output
func (e *simpleExecutor) RunBatch(ctx context.Context, cmds []string) ([]byte, error) {
	var allOutput []byte

	for _, cmd := range cmds {
		output, err := e.Run(ctx, cmd)
		if err != nil {
			return allOutput, fmt.Errorf("batch command '%s' failed: %w", cmd, err)
		}
		allOutput = append(allOutput, output...)
	}

	return allOutput, nil
}

// SetAdministratorPassword sets the administrator password using interactive prompts
// RTX prompts: Old_Password: -> New_Password: -> New_Password: -> Password Strength: ...
func (e *simpleExecutor) SetAdministratorPassword(ctx context.Context, oldPassword, newPassword string) error {
	logger := logging.FromContext(ctx)
	logger.Debug().Msg("SimpleExecutor: Setting administrator password")

	// Create a new SSH connection for the interactive password command
	client, err := ssh.Dial("tcp", e.addr, e.config)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	// Create a working session
	ws, err := newWorkingSession(client)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer ws.Close()

	// Authenticate as administrator first (required for password commands)
	if e.rtxConfig != nil && e.rtxConfig.AdminPassword != "" {
		if err := e.authenticateAsAdminWithSession(ctx, ws); err != nil {
			return fmt.Errorf("failed to authenticate as administrator: %w", err)
		}
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Send administrator password command
	if _, err := fmt.Fprintf(ws.stdin, "administrator password\r"); err != nil {
		return fmt.Errorf("failed to send administrator password command: %w", err)
	}

	// Wait for Old_Password: prompt
	_, err = ws.readUntilString("Old_Password:", 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get Old_Password prompt: %w", err)
	}
	logger.Debug().Msg("SimpleExecutor: Old_Password prompt received")

	// Send old password
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", oldPassword); err != nil {
		return fmt.Errorf("failed to send old password: %w", err)
	}

	// Wait for first New_Password: prompt
	_, err = ws.readUntilString("New_Password:", 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get first New_Password prompt: %w", err)
	}
	logger.Debug().Msg("SimpleExecutor: First New_Password prompt received")

	// Send new password
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", newPassword); err != nil {
		return fmt.Errorf("failed to send new password: %w", err)
	}

	// Wait for second New_Password: prompt (confirmation)
	_, err = ws.readUntilString("New_Password:", 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get second New_Password prompt: %w", err)
	}
	logger.Debug().Msg("SimpleExecutor: Second New_Password prompt received")

	// Send new password again for confirmation
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", newPassword); err != nil {
		return fmt.Errorf("failed to send password confirmation: %w", err)
	}

	// Wait for completion (Password Strength or prompt)
	response, err := ws.readUntilPrompt(10 * time.Second)
	if err != nil {
		return fmt.Errorf("failed to read password change response: %w", err)
	}

	responseStr := string(response)
	logger.Debug().Str("response", responseStr).Msg("SimpleExecutor: Password change response received")

	// Check for errors
	if strings.Contains(responseStr, "incorrect") || strings.Contains(responseStr, "failed") || strings.Contains(responseStr, "Invalid") ||
		strings.Contains(responseStr, "エラー") || strings.Contains(responseStr, "パスワードが違います") {
		return fmt.Errorf("administrator password change failed: %s", responseStr)
	}

	logger.Debug().Msg("SimpleExecutor: Administrator password changed successfully")
	return nil
}

// SetLoginPassword sets the login password using interactive prompts
// RTX prompts: Old_Password: (if exists) -> New_Password: -> New_Password: -> Password Strength: ...
func (e *simpleExecutor) SetLoginPassword(ctx context.Context, newPassword string) error {
	logger := logging.FromContext(ctx)
	logger.Debug().Msg("SimpleExecutor: Setting login password")

	// Create a new SSH connection for the interactive password command
	client, err := ssh.Dial("tcp", e.addr, e.config)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	// Create a working session
	ws, err := newWorkingSession(client)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer ws.Close()

	// Authenticate as administrator first (required for password commands)
	if e.rtxConfig != nil && e.rtxConfig.AdminPassword != "" {
		if err := e.authenticateAsAdminWithSession(ctx, ws); err != nil {
			return fmt.Errorf("failed to authenticate as administrator: %w", err)
		}
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Send login password command
	if _, err := fmt.Fprintf(ws.stdin, "login password\r"); err != nil {
		return fmt.Errorf("failed to send login password command: %w", err)
	}

	// Wait for New_Password: prompt (login password may not have old password prompt if not set)
	_, err = ws.readUntilString("New_Password:", 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get first New_Password prompt: %w", err)
	}
	logger.Debug().Msg("SimpleExecutor: First New_Password prompt received")

	// Send new password
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", newPassword); err != nil {
		return fmt.Errorf("failed to send new password: %w", err)
	}

	// Wait for second New_Password: prompt (confirmation)
	_, err = ws.readUntilString("New_Password:", 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to get second New_Password prompt: %w", err)
	}
	logger.Debug().Msg("SimpleExecutor: Second New_Password prompt received")

	// Send new password again for confirmation
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", newPassword); err != nil {
		return fmt.Errorf("failed to send password confirmation: %w", err)
	}

	// Wait for completion (Password Strength or prompt)
	response, err := ws.readUntilPrompt(10 * time.Second)
	if err != nil {
		return fmt.Errorf("failed to read password change response: %w", err)
	}

	responseStr := string(response)
	logger.Debug().Str("response", responseStr).Msg("SimpleExecutor: Password change response received")

	// Check for errors
	if strings.Contains(responseStr, "incorrect") || strings.Contains(responseStr, "failed") || strings.Contains(responseStr, "Invalid") ||
		strings.Contains(responseStr, "エラー") || strings.Contains(responseStr, "パスワードが違います") {
		return fmt.Errorf("login password change failed: %s", responseStr)
	}

	logger.Debug().Msg("SimpleExecutor: Login password changed successfully")
	return nil
}

// authenticateAsAdminWithSession authenticates as administrator using the given session
// Handles the case where the session is already in administrator mode
func (e *simpleExecutor) authenticateAsAdminWithSession(ctx context.Context, ws *workingSession) error {
	logger := logging.FromContext(ctx)
	logger.Debug().Msg("SimpleExecutor: Authenticating as administrator")

	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return fmt.Errorf("session is closed")
	}

	// Send administrator command
	if _, err := fmt.Fprintf(ws.stdin, "administrator\r"); err != nil {
		return fmt.Errorf("failed to send administrator command: %w", err)
	}

	// Read until we get password prompt or admin prompt (already administrator)
	response, err := ws.readUntilPasswordPromptOrAdminMode(10 * time.Second)
	if err != nil {
		return fmt.Errorf("failed to get response after administrator command: %w", err)
	}

	responseStr := string(response)
	logger.Debug().Str("response", responseStr).Msg("SimpleExecutor: Response after administrator command")

	// Check if already in administrator mode
	if strings.Contains(responseStr, "すでに管理レベル") || strings.Contains(strings.ToLower(responseStr), "already") {
		logger.Debug().Msg("SimpleExecutor: Already in administrator mode, skipping authentication")
		return nil
	}

	// Check if we got admin prompt directly (# at end of line)
	// This can happen if the session started in admin mode
	if strings.Contains(responseStr, "# ") || strings.HasSuffix(strings.TrimSpace(responseStr), "#") {
		// Check if this is NOT the password prompt case
		if !strings.Contains(responseStr, "Password:") && !strings.Contains(responseStr, "password:") {
			logger.Debug().Msg("SimpleExecutor: Session appears to be in administrator mode already")
			return nil
		}
	}

	// We should have received Password: prompt
	if !strings.Contains(responseStr, "Password:") && !strings.Contains(responseStr, "password:") {
		return fmt.Errorf("unexpected response after administrator command: %s", responseStr)
	}

	logger.Debug().Msg("SimpleExecutor: Password prompt received")

	// Send password
	if _, err := fmt.Fprintf(ws.stdin, "%s\r", e.rtxConfig.AdminPassword); err != nil {
		return fmt.Errorf("failed to send password: %w", err)
	}

	// Read response after password - look for administrator prompt (# instead of >)
	response, err = ws.readUntilPrompt(10 * time.Second)
	if err != nil {
		return fmt.Errorf("failed to read password response: %w", err)
	}

	responseStr = string(response)

	// Check for authentication success
	if strings.Contains(responseStr, "incorrect") || strings.Contains(responseStr, "failed") || strings.Contains(responseStr, "Invalid") ||
		strings.Contains(responseStr, "エラー") || strings.Contains(responseStr, "パスワードが違います") {
		return fmt.Errorf("administrator authentication failed: %s", responseStr)
	}

	logger.Debug().Msg("SimpleExecutor: Administrator authentication successful")
	return nil
}

// GenerateSSHDHostKey generates SSHD host key with interactive prompt handling
// RTX may prompt for confirmation if a host key already exists
func (e *simpleExecutor) GenerateSSHDHostKey(ctx context.Context) error {
	logger := logging.FromContext(ctx)
	logger.Debug().Msg("SimpleExecutor: Generating SSHD host key")

	// Create a new SSH connection for the interactive command
	client, err := ssh.Dial("tcp", e.addr, e.config)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	// Create a working session
	ws, err := newWorkingSession(client)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer ws.Close()

	// Authenticate as administrator first (required for sshd commands)
	if e.rtxConfig != nil && e.rtxConfig.AdminPassword != "" {
		if err := e.authenticateAsAdminWithSession(ctx, ws); err != nil {
			return fmt.Errorf("failed to authenticate as administrator: %w", err)
		}
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Send sshd host key generate command
	logger.Debug().Msg("SimpleExecutor: Sending sshd host key generate command")
	if _, err := fmt.Fprintf(ws.stdin, "sshd host key generate\r"); err != nil {
		return fmt.Errorf("failed to send sshd host key generate command: %w", err)
	}

	// Key generation can take several minutes on RTX hardware
	// Read response - either:
	// 1. Confirmation prompt (Y/N) if host key already exists
	// 2. Direct completion with prompt if no existing key
	keyGenTimeout := 10 * time.Minute
	response, err := ws.readUntilPromptOrConfirmation(keyGenTimeout)
	if err != nil {
		return fmt.Errorf("failed to read sshd host key generate response: %w", err)
	}

	responseStr := string(response)
	logger.Debug().Str("response", responseStr).Msg("SimpleExecutor: Host key generate response received")

	// Check if we got a confirmation prompt (existing key)
	if ws.isHostKeyUpdatePrompt(responseStr) {
		logger.Info().Msg("SimpleExecutor: Host key update prompt detected, responding with 'N' to preserve existing key")
		// Respond with 'N' to abort regeneration and preserve existing key
		// This is a safety measure to prevent accidental host key regeneration
		if _, err := fmt.Fprintf(ws.stdin, "N\r"); err != nil {
			return fmt.Errorf("failed to respond to host key update prompt: %w", err)
		}

		// Wait for prompt after aborting
		_, err := ws.readUntilPrompt(keyGenTimeout)
		if err != nil {
			return fmt.Errorf("failed to read response after aborting host key generation: %w", err)
		}

		// Return informational error
		return fmt.Errorf("host key already exists; generation aborted to preserve existing key")
	}

	// Check for errors in response
	if strings.Contains(strings.ToLower(responseStr), "error") ||
		strings.Contains(strings.ToLower(responseStr), "failed") {
		return fmt.Errorf("sshd host key generation failed: %s", responseStr)
	}

	logger.Debug().Msg("SimpleExecutor: SSHD host key generated successfully")
	return nil
}
