package client

import (
	"context"
	"fmt"
	"time"

	"github.com/sh1/terraform-provider-rtx/internal/logging"
)

// TunnelUpdateConfig holds configuration for VPN safe tunnel updates
type TunnelUpdateConfig struct {
	TunnelID           int
	Commands           []string
	ReconnectTimeout   time.Duration
	ReconnectPollDelay time.Duration
	SkipDisable        bool // Skip tunnel disable/enable (for non-disruptive changes)
}

// DefaultTunnelUpdateConfig returns default configuration for tunnel updates
func DefaultTunnelUpdateConfig(tunnelID int, commands []string) *TunnelUpdateConfig {
	return &TunnelUpdateConfig{
		TunnelID:           tunnelID,
		Commands:           commands,
		ReconnectTimeout:   5 * time.Minute,
		ReconnectPollDelay: 5 * time.Second,
		SkipDisable:        false,
	}
}

// ExecuteTunnelUpdate performs a VPN-safe tunnel update
// This is designed for updating tunnel configuration when connected via the same tunnel (VPN)
//
// The pattern works as follows:
// 1. Send all configuration commands in batch (without save)
// 2. If needed, disable then enable the tunnel (still no save)
// 3. Wait for SSH connection to be re-established (polling)
// 4. Once reconnected, save the configuration
//
// If the configuration is incorrect and the connection cannot be re-established,
// a manual router restart will restore the previous (saved) configuration.
func ExecuteTunnelUpdate(ctx context.Context, client Client, config *TunnelUpdateConfig) error {
	logger := logging.FromContext(ctx)
	logger.Info().Int("tunnel", config.TunnelID).Msg("Starting VPN-safe tunnel update")

	// Phase 1: Send all configuration commands in batch (no save)
	allCommands := make([]string, 0, len(config.Commands)+2)
	allCommands = append(allCommands, config.Commands...)

	if !config.SkipDisable {
		// Add tunnel disable and enable commands
		allCommands = append(allCommands,
			fmt.Sprintf("tunnel select %d", config.TunnelID),
			"tunnel disable",
			"tunnel enable",
		)
	}

	logger.Debug().Strs("commands", allCommands).Msg("Sending batch commands (no save)")

	// Execute batch - this may cause connection to drop
	_, err := client.RunBatch(ctx, allCommands)
	if err != nil {
		// Connection drop is expected if we're connected via this tunnel
		logger.Debug().Err(err).Msg("Batch execution returned error (expected if connected via tunnel)")
	}

	// Phase 2: Wait for reconnection
	if !config.SkipDisable {
		logger.Info().Dur("timeout", config.ReconnectTimeout).Msg("Waiting for SSH reconnection")

		reconnectCtx, cancel := context.WithTimeout(ctx, config.ReconnectTimeout)
		defer cancel()

		if err := waitForReconnection(reconnectCtx, client, config.ReconnectPollDelay); err != nil {
			return fmt.Errorf("failed to reconnect after tunnel update: %w", err)
		}

		logger.Info().Msg("SSH connection re-established")
	}

	// Phase 3: Save configuration
	logger.Debug().Msg("Saving configuration")
	_, err = client.Run(ctx, Command{Key: "save", Payload: "save"})
	if err != nil {
		return fmt.Errorf("failed to save configuration after tunnel update: %w", err)
	}

	logger.Info().Int("tunnel", config.TunnelID).Msg("VPN-safe tunnel update completed successfully")
	return nil
}

// waitForReconnection polls until SSH connection can be established
func waitForReconnection(ctx context.Context, client Client, pollDelay time.Duration) error {
	logger := logging.FromContext(ctx)
	attempt := 0

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("reconnection timeout: %w", ctx.Err())
		default:
			attempt++
			logger.Debug().Int("attempt", attempt).Msg("Attempting to reconnect")

			// Try to establish new connection
			dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			err := client.Dial(dialCtx)
			cancel()

			if err == nil {
				// Test connection with a simple command
				testCtx, testCancel := context.WithTimeout(ctx, 10*time.Second)
				_, testErr := client.Run(testCtx, Command{Key: "test", Payload: "show environment"})
				testCancel()

				if testErr == nil {
					return nil // Successfully reconnected
				}
				logger.Debug().Err(testErr).Msg("Connection test failed")
			} else {
				logger.Debug().Err(err).Msg("Dial failed")
			}

			// Wait before next attempt
			select {
			case <-ctx.Done():
				return fmt.Errorf("reconnection timeout: %w", ctx.Err())
			case <-time.After(pollDelay):
				continue
			}
		}
	}
}

// ExecuteIPsecTunnelUpdate is a convenience wrapper for IPsec tunnel updates
func ExecuteIPsecTunnelUpdate(ctx context.Context, client Client, tunnelID int, commands []string) error {
	config := DefaultTunnelUpdateConfig(tunnelID, commands)
	return ExecuteTunnelUpdate(ctx, client, config)
}

// ExecuteL2TPTunnelUpdate is a convenience wrapper for L2TP tunnel updates
func ExecuteL2TPTunnelUpdate(ctx context.Context, client Client, tunnelID int, commands []string) error {
	config := DefaultTunnelUpdateConfig(tunnelID, commands)
	return ExecuteTunnelUpdate(ctx, client, config)
}

// ExecutePPTPTunnelUpdate is a convenience wrapper for PPTP tunnel updates
func ExecutePPTPTunnelUpdate(ctx context.Context, client Client, tunnelID int, commands []string) error {
	config := DefaultTunnelUpdateConfig(tunnelID, commands)
	return ExecuteTunnelUpdate(ctx, client, config)
}

// ExecuteNonDisruptiveUpdate executes commands that don't require tunnel restart
func ExecuteNonDisruptiveUpdate(ctx context.Context, client Client, commands []string) error {
	config := &TunnelUpdateConfig{
		Commands:    commands,
		SkipDisable: true,
	}
	return ExecuteTunnelUpdate(ctx, client, config)
}
