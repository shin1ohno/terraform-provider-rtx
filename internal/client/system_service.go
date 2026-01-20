package client

import (
	"context"
	"fmt"
	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// SystemService handles system configuration operations
type SystemService struct {
	executor Executor
	client   *rtxClient // Reference to the main client for save functionality
}

// NewSystemService creates a new system service instance
func NewSystemService(executor Executor, client *rtxClient) *SystemService {
	return &SystemService{
		executor: executor,
		client:   client,
	}
}

// Configure sets system configuration
func (s *SystemService) Configure(ctx context.Context, config SystemConfig) error {
	// Convert client.SystemConfig to parsers.SystemConfig
	parserConfig := s.toParserConfig(config)

	// Validate input
	if err := parsers.ValidateSystemConfig(&parserConfig); err != nil {
		return fmt.Errorf("invalid system config: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Apply timezone
	if config.Timezone != "" {
		cmd := parsers.BuildTimezoneCommand(config.Timezone)
		logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Setting timezone with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to set timezone: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("timezone command failed: %s", string(output))
		}
	}

	// Apply console settings
	if config.Console != nil {
		if config.Console.Character != "" {
			cmd := parsers.BuildConsoleCharacterCommand(config.Console.Character)
			logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Setting console character with command: %s", cmd)

			output, err := s.executor.Run(ctx, cmd)
			if err != nil {
				return fmt.Errorf("failed to set console character: %w", err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("console character command failed: %s", string(output))
			}
		}

		if config.Console.Lines != "" {
			cmd := parsers.BuildConsoleLinesCommand(config.Console.Lines)
			logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Setting console lines with command: %s", cmd)

			output, err := s.executor.Run(ctx, cmd)
			if err != nil {
				return fmt.Errorf("failed to set console lines: %w", err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("console lines command failed: %s", string(output))
			}
		}

		if config.Console.Prompt != "" {
			cmd := parsers.BuildConsolePromptCommand(config.Console.Prompt)
			logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Setting console prompt with command: %s", cmd)

			output, err := s.executor.Run(ctx, cmd)
			if err != nil {
				return fmt.Errorf("failed to set console prompt: %w", err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("console prompt command failed: %s", string(output))
			}
		}
	}

	// Apply packet buffer settings
	for _, pb := range config.PacketBuffers {
		parserPB := parsers.PacketBufferConfig{
			Size:      pb.Size,
			MaxBuffer: pb.MaxBuffer,
			MaxFree:   pb.MaxFree,
		}
		cmd := parsers.BuildPacketBufferCommand(parserPB)
		logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Setting packet buffer with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to set packet buffer %s: %w", pb.Size, err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("packet buffer command failed: %s", string(output))
		}
	}

	// Apply statistics settings
	if config.Statistics != nil {
		cmd := parsers.BuildStatisticsTrafficCommand(config.Statistics.Traffic)
		logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Setting statistics traffic with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to set statistics traffic: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("statistics traffic command failed: %s", string(output))
		}

		cmd = parsers.BuildStatisticsNATCommand(config.Statistics.NAT)
		logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Setting statistics NAT with command: %s", cmd)

		output, err = s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to set statistics NAT: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("statistics NAT command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("system config set but failed to save configuration: %w", err)
		}
	}

	return nil
}

// Get retrieves system configuration
func (s *SystemService) Get(ctx context.Context) (*SystemConfig, error) {
	cmd := parsers.BuildShowSystemConfigCommand()
	logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Getting system config with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get system config: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "system").Msgf("System config raw output: %q", string(output))

	parser := parsers.NewSystemParser()
	parserConfig, err := parser.ParseSystemConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse system config: %w", err)
	}

	// Convert parsers.SystemConfig to client.SystemConfig
	config := s.fromParserConfig(*parserConfig)
	return &config, nil
}

// Update updates system configuration
func (s *SystemService) Update(ctx context.Context, config SystemConfig) error {
	// Convert client.SystemConfig to parsers.SystemConfig
	parserConfig := s.toParserConfig(config)

	// Validate input
	if err := parsers.ValidateSystemConfig(&parserConfig); err != nil {
		return fmt.Errorf("invalid system config: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get current configuration for comparison
	current, err := s.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current system config: %w", err)
	}

	// Update timezone if changed
	if config.Timezone != current.Timezone {
		if config.Timezone != "" {
			cmd := parsers.BuildTimezoneCommand(config.Timezone)
			logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Updating timezone with command: %s", cmd)

			output, err := s.executor.Run(ctx, cmd)
			if err != nil {
				return fmt.Errorf("failed to update timezone: %w", err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("timezone command failed: %s", string(output))
			}
		} else if current.Timezone != "" {
			// Remove timezone setting
			cmd := parsers.BuildDeleteTimezoneCommand()
			logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Removing timezone with command: %s", cmd)
			_, _ = s.executor.Run(ctx, cmd)
		}
	}

	// Update console settings
	if config.Console != nil {
		currentConsole := current.Console
		if currentConsole == nil {
			currentConsole = &ConsoleConfig{}
		}

		if config.Console.Character != currentConsole.Character {
			if config.Console.Character != "" {
				cmd := parsers.BuildConsoleCharacterCommand(config.Console.Character)
				logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Updating console character with command: %s", cmd)

				output, err := s.executor.Run(ctx, cmd)
				if err != nil {
					return fmt.Errorf("failed to update console character: %w", err)
				}

				if len(output) > 0 && containsError(string(output)) {
					return fmt.Errorf("console character command failed: %s", string(output))
				}
			} else if currentConsole.Character != "" {
				cmd := parsers.BuildDeleteConsoleCharacterCommand()
				_, _ = s.executor.Run(ctx, cmd)
			}
		}

		if config.Console.Lines != currentConsole.Lines {
			if config.Console.Lines != "" {
				cmd := parsers.BuildConsoleLinesCommand(config.Console.Lines)
				logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Updating console lines with command: %s", cmd)

				output, err := s.executor.Run(ctx, cmd)
				if err != nil {
					return fmt.Errorf("failed to update console lines: %w", err)
				}

				if len(output) > 0 && containsError(string(output)) {
					return fmt.Errorf("console lines command failed: %s", string(output))
				}
			} else if currentConsole.Lines != "" {
				cmd := parsers.BuildDeleteConsoleLinesCommand()
				_, _ = s.executor.Run(ctx, cmd)
			}
		}

		if config.Console.Prompt != currentConsole.Prompt {
			if config.Console.Prompt != "" {
				cmd := parsers.BuildConsolePromptCommand(config.Console.Prompt)
				logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Updating console prompt with command: %s", cmd)

				output, err := s.executor.Run(ctx, cmd)
				if err != nil {
					return fmt.Errorf("failed to update console prompt: %w", err)
				}

				if len(output) > 0 && containsError(string(output)) {
					return fmt.Errorf("console prompt command failed: %s", string(output))
				}
			} else if currentConsole.Prompt != "" {
				cmd := parsers.BuildDeleteConsolePromptCommand()
				_, _ = s.executor.Run(ctx, cmd)
			}
		}
	} else if current.Console != nil {
		// Remove all console settings
		if current.Console.Character != "" {
			cmd := parsers.BuildDeleteConsoleCharacterCommand()
			_, _ = s.executor.Run(ctx, cmd)
		}
		if current.Console.Lines != "" {
			cmd := parsers.BuildDeleteConsoleLinesCommand()
			_, _ = s.executor.Run(ctx, cmd)
		}
		if current.Console.Prompt != "" {
			cmd := parsers.BuildDeleteConsolePromptCommand()
			_, _ = s.executor.Run(ctx, cmd)
		}
	}

	// Update packet buffer settings
	// Create maps for easier comparison
	currentPBMap := make(map[string]PacketBufferConfig)
	for _, pb := range current.PacketBuffers {
		currentPBMap[pb.Size] = pb
	}

	newPBMap := make(map[string]PacketBufferConfig)
	for _, pb := range config.PacketBuffers {
		newPBMap[pb.Size] = pb
	}

	// Remove packet buffers that are no longer in the new config
	for size := range currentPBMap {
		if _, exists := newPBMap[size]; !exists {
			cmd := parsers.BuildDeletePacketBufferCommand(size)
			logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Removing packet buffer with command: %s", cmd)
			_, _ = s.executor.Run(ctx, cmd)
		}
	}

	// Add or update packet buffers
	for _, pb := range config.PacketBuffers {
		currentPB, exists := currentPBMap[pb.Size]
		if !exists || currentPB.MaxBuffer != pb.MaxBuffer || currentPB.MaxFree != pb.MaxFree {
			parserPB := parsers.PacketBufferConfig{
				Size:      pb.Size,
				MaxBuffer: pb.MaxBuffer,
				MaxFree:   pb.MaxFree,
			}
			cmd := parsers.BuildPacketBufferCommand(parserPB)
			logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Updating packet buffer with command: %s", cmd)

			output, err := s.executor.Run(ctx, cmd)
			if err != nil {
				return fmt.Errorf("failed to update packet buffer %s: %w", pb.Size, err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("packet buffer command failed: %s", string(output))
			}
		}
	}

	// Update statistics settings
	if config.Statistics != nil {
		currentStats := current.Statistics
		if currentStats == nil {
			currentStats = &StatisticsConfig{}
		}

		if config.Statistics.Traffic != currentStats.Traffic {
			cmd := parsers.BuildStatisticsTrafficCommand(config.Statistics.Traffic)
			logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Updating statistics traffic with command: %s", cmd)

			output, err := s.executor.Run(ctx, cmd)
			if err != nil {
				return fmt.Errorf("failed to update statistics traffic: %w", err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("statistics traffic command failed: %s", string(output))
			}
		}

		if config.Statistics.NAT != currentStats.NAT {
			cmd := parsers.BuildStatisticsNATCommand(config.Statistics.NAT)
			logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Updating statistics NAT with command: %s", cmd)

			output, err := s.executor.Run(ctx, cmd)
			if err != nil {
				return fmt.Errorf("failed to update statistics NAT: %w", err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("statistics NAT command failed: %s", string(output))
			}
		}
	} else if current.Statistics != nil {
		// Remove statistics settings
		cmd := parsers.BuildDeleteStatisticsTrafficCommand()
		_, _ = s.executor.Run(ctx, cmd)
		cmd = parsers.BuildDeleteStatisticsNATCommand()
		_, _ = s.executor.Run(ctx, cmd)
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("system config updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// Reset resets system configuration to defaults
func (s *SystemService) Reset(ctx context.Context) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get current configuration to know what to remove
	current, err := s.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current system config: %w", err)
	}

	// Build and execute delete commands
	parserConfig := s.toParserConfig(*current)
	commands := parsers.BuildDeleteSystemCommands(&parserConfig)

	for _, cmd := range commands {
		logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Resetting system config with command: %s", cmd)
		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			// Log but continue - some settings might not exist
			logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Warning: command failed: %v", err)
			continue
		}

		if len(output) > 0 && containsError(string(output)) {
			// Log but continue
			logging.FromContext(ctx).Debug().Str("service", "system").Msgf("Warning: command output indicates error: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("system config reset but failed to save configuration: %w", err)
		}
	}

	return nil
}

// toParserConfig converts client.SystemConfig to parsers.SystemConfig
func (s *SystemService) toParserConfig(config SystemConfig) parsers.SystemConfig {
	parserConfig := parsers.SystemConfig{
		Timezone:      config.Timezone,
		PacketBuffers: []parsers.PacketBufferConfig{},
	}

	if config.Console != nil {
		parserConfig.Console = &parsers.ConsoleConfig{
			Character: config.Console.Character,
			Lines:     config.Console.Lines,
			Prompt:    config.Console.Prompt,
		}
	}

	for _, pb := range config.PacketBuffers {
		parserConfig.PacketBuffers = append(parserConfig.PacketBuffers, parsers.PacketBufferConfig{
			Size:      pb.Size,
			MaxBuffer: pb.MaxBuffer,
			MaxFree:   pb.MaxFree,
		})
	}

	if config.Statistics != nil {
		parserConfig.Statistics = &parsers.StatisticsConfig{
			Traffic: config.Statistics.Traffic,
			NAT:     config.Statistics.NAT,
		}
	}

	return parserConfig
}

// fromParserConfig converts parsers.SystemConfig to client.SystemConfig
func (s *SystemService) fromParserConfig(pc parsers.SystemConfig) SystemConfig {
	config := SystemConfig{
		Timezone:      pc.Timezone,
		PacketBuffers: []PacketBufferConfig{},
	}

	if pc.Console != nil {
		config.Console = &ConsoleConfig{
			Character: pc.Console.Character,
			Lines:     pc.Console.Lines,
			Prompt:    pc.Console.Prompt,
		}
	}

	for _, pb := range pc.PacketBuffers {
		config.PacketBuffers = append(config.PacketBuffers, PacketBufferConfig{
			Size:      pb.Size,
			MaxBuffer: pb.MaxBuffer,
			MaxFree:   pb.MaxFree,
		})
	}

	if pc.Statistics != nil {
		config.Statistics = &StatisticsConfig{
			Traffic: pc.Statistics.Traffic,
			NAT:     pc.Statistics.NAT,
		}
	}

	return config
}
