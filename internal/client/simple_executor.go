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
	logger.Debug().Str("command", logging.SanitizeString(cmd)).Msg("SimpleExecutor: Running command")

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
// If admin password is configured, always use administrator mode since RTX routers
// provide more complete information (e.g., show config) in administrator mode.
func (e *simpleExecutor) requiresAdminPrivileges(cmd string) bool {
	// If admin password is configured, always use administrator mode
	if e.rtxConfig != nil && e.rtxConfig.AdminPassword != "" {
		return true
	}
	return false
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
	logger.Debug().Msg("SimpleExecutor: Password authentication response received")

	// Check for authentication success (look for # prompt or no error message)
	if strings.Contains(responseStr, "incorrect") || strings.Contains(responseStr, "failed") || strings.Contains(responseStr, "Invalid") {
		return fmt.Errorf("administrator authentication failed: %s", responseStr)
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
