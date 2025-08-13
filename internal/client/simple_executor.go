package client

import (
	"context"
	"fmt"
	"log"

	"golang.org/x/crypto/ssh"
)

// simpleExecutor executes commands by creating a new SSH session for each command
type simpleExecutor struct {
	config         *ssh.ClientConfig
	addr           string
	promptDetector PromptDetector
}

// NewSimpleExecutor creates a new simple executor
func NewSimpleExecutor(config *ssh.ClientConfig, addr string, promptDetector PromptDetector) Executor {
	return &simpleExecutor{
		config:         config,
		addr:           addr,
		promptDetector: promptDetector,
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