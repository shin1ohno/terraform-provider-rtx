package client

import (
	"context"
	"fmt"
	"strings"
)

// saveConfig persists the running configuration to flash memory.
func saveConfig(ctx context.Context, client *rtxClient, operationDesc string) error {
	if client == nil {
		return nil
	}
	if err := client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("%s but failed to save configuration: %w", operationDesc, err)
	}
	return nil
}

// checkOutputError returns an error if the output contains RTX error patterns.
func checkOutputError(output []byte, operationDesc string) error {
	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("%s: %s", operationDesc, string(output))
	}
	return nil
}

// checkOutputErrorIgnoringNotFound returns an error if the output contains RTX error patterns, ignoring "not found".
func checkOutputErrorIgnoringNotFound(output []byte, operationDesc string) error {
	if len(output) > 0 && containsError(string(output)) {
		if strings.Contains(strings.ToLower(string(output)), "not found") {
			return nil
		}
		return fmt.Errorf("%s: %s", operationDesc, string(output))
	}
	return nil
}

// runCommand executes a single command via the executor.
func runCommand(ctx context.Context, executor Executor, cmd string) error {
	output, err := executor.Run(ctx, cmd)
	if err != nil {
		return err
	}
	return checkOutputError(output, "command failed")
}

// runBatchCommands executes multiple commands via the executor.
func runBatchCommands(ctx context.Context, executor Executor, commands []string) error {
	if len(commands) == 0 {
		return nil
	}
	output, err := executor.RunBatch(ctx, commands)
	if err != nil {
		return err
	}
	return checkOutputError(output, "batch command failed")
}
