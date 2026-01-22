package client

import (
	"context"
	"fmt"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// BGPService handles BGP configuration operations
type BGPService struct {
	executor Executor
	client   *rtxClient
}

// NewBGPService creates a new BGP service
func NewBGPService(executor Executor, client *rtxClient) *BGPService {
	return &BGPService{
		executor: executor,
		client:   client,
	}
}

// Get retrieves the current BGP configuration
func (s *BGPService) Get(ctx context.Context) (*BGPConfig, error) {
	output, err := s.executor.Run(ctx, parsers.BuildShowBGPConfigCommand())
	if err != nil {
		return nil, fmt.Errorf("failed to get BGP config: %w", err)
	}

	parser := parsers.NewBGPParser()
	parsed, err := parser.ParseBGPConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse BGP config: %w", err)
	}

	// Convert from parser type to client type
	config := &BGPConfig{
		Enabled:               parsed.Enabled,
		ASN:                   parsed.ASN,
		RouterID:              parsed.RouterID,
		DefaultIPv4Unicast:    parsed.DefaultIPv4Unicast,
		LogNeighborChanges:    parsed.LogNeighborChanges,
		RedistributeStatic:    parsed.RedistributeStatic,
		RedistributeConnected: parsed.RedistributeConnected,
		Neighbors:             make([]BGPNeighbor, len(parsed.Neighbors)),
		Networks:              make([]BGPNetwork, len(parsed.Networks)),
	}

	for i, n := range parsed.Neighbors {
		config.Neighbors[i] = BGPNeighbor{
			ID:           n.ID,
			IP:           n.IP,
			RemoteAS:     n.RemoteAS,
			HoldTime:     n.HoldTime,
			Keepalive:    n.Keepalive,
			Multihop:     n.Multihop,
			Password:     n.Password,
			LocalAddress: n.LocalAddress,
		}
	}

	for i, net := range parsed.Networks {
		config.Networks[i] = BGPNetwork{
			Prefix: net.Prefix,
			Mask:   net.Mask,
		}
	}

	return config, nil
}

// Configure creates a new BGP configuration
func (s *BGPService) Configure(ctx context.Context, config BGPConfig) error {
	// Validate configuration
	parserConfig := convertToParserBGPConfig(config)
	if err := parsers.ValidateBGPConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid BGP config: %w", err)
	}

	// Build and execute commands in order
	commands := []string{}

	// 1. Set AS number first
	commands = append(commands, parsers.BuildBGPASNCommand(config.ASN))

	// 2. Set router ID if specified
	if config.RouterID != "" {
		commands = append(commands, parsers.BuildBGPRouterIDCommand(config.RouterID))
	}

	// 3. Configure neighbors
	for _, neighbor := range config.Neighbors {
		parserNeighbor := parsers.BGPNeighbor{
			ID:       neighbor.ID,
			IP:       neighbor.IP,
			RemoteAS: neighbor.RemoteAS,
		}
		commands = append(commands, parsers.BuildBGPNeighborCommand(parserNeighbor))

		if neighbor.HoldTime > 0 {
			commands = append(commands, parsers.BuildBGPNeighborHoldTimeCommand(neighbor.ID, neighbor.HoldTime))
		}
		if neighbor.Keepalive > 0 {
			commands = append(commands, parsers.BuildBGPNeighborKeepaliveCommand(neighbor.ID, neighbor.Keepalive))
		}
		if neighbor.Multihop > 0 {
			commands = append(commands, parsers.BuildBGPNeighborMultihopCommand(neighbor.ID, neighbor.Multihop))
		}
		if neighbor.Password != "" {
			commands = append(commands, parsers.BuildBGPNeighborPasswordCommand(neighbor.ID, neighbor.Password))
		}
		if neighbor.LocalAddress != "" {
			commands = append(commands, parsers.BuildBGPNeighborLocalAddressCommand(neighbor.ID, neighbor.LocalAddress))
		}
	}

	// 4. Configure networks
	for i, network := range config.Networks {
		parserNetwork := parsers.BGPNetwork{
			Prefix: network.Prefix,
			Mask:   network.Mask,
		}
		commands = append(commands, parsers.BuildBGPNetworkCommand(i+1, parserNetwork))
	}

	// 5. Configure redistribution
	if config.RedistributeStatic {
		commands = append(commands, parsers.BuildBGPRedistributeCommand("static"))
	}
	if config.RedistributeConnected {
		commands = append(commands, parsers.BuildBGPRedistributeCommand("connected"))
	}

	// 6. Enable BGP
	commands = append(commands, parsers.BuildBGPUseCommand(true))

	// Execute all commands in batch
	output, err := s.executor.RunBatch(ctx, commands)
	if err != nil {
		return fmt.Errorf("failed to execute BGP batch commands: %w", err)
	}
	if containsError(string(output)) {
		return fmt.Errorf("BGP batch commands failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("failed to save BGP config: %w", err)
		}
	}

	return nil
}

// Update modifies the existing BGP configuration
func (s *BGPService) Update(ctx context.Context, config BGPConfig) error {
	// Get current config to determine what needs to change
	current, err := s.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current BGP config: %w", err)
	}

	// Validate new configuration
	parserConfig := convertToParserBGPConfig(config)
	if err := parsers.ValidateBGPConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid BGP config: %w", err)
	}

	commands := []string{}

	// Update AS number if changed
	if config.ASN != current.ASN {
		commands = append(commands, parsers.BuildBGPASNCommand(config.ASN))
	}

	// Update router ID if changed
	if config.RouterID != current.RouterID {
		commands = append(commands, parsers.BuildBGPRouterIDCommand(config.RouterID))
	}

	// Remove old neighbors that are not in new config
	for _, oldNeighbor := range current.Neighbors {
		found := false
		for _, newNeighbor := range config.Neighbors {
			if oldNeighbor.ID == newNeighbor.ID {
				found = true
				break
			}
		}
		if !found {
			commands = append(commands, parsers.BuildDeleteBGPNeighborCommand(oldNeighbor.ID))
		}
	}

	// Add/update neighbors
	for _, neighbor := range config.Neighbors {
		parserNeighbor := parsers.BGPNeighbor{
			ID:       neighbor.ID,
			IP:       neighbor.IP,
			RemoteAS: neighbor.RemoteAS,
		}
		commands = append(commands, parsers.BuildBGPNeighborCommand(parserNeighbor))

		if neighbor.HoldTime > 0 {
			commands = append(commands, parsers.BuildBGPNeighborHoldTimeCommand(neighbor.ID, neighbor.HoldTime))
		}
		if neighbor.Keepalive > 0 {
			commands = append(commands, parsers.BuildBGPNeighborKeepaliveCommand(neighbor.ID, neighbor.Keepalive))
		}
		if neighbor.Multihop > 0 {
			commands = append(commands, parsers.BuildBGPNeighborMultihopCommand(neighbor.ID, neighbor.Multihop))
		}
		if neighbor.Password != "" {
			commands = append(commands, parsers.BuildBGPNeighborPasswordCommand(neighbor.ID, neighbor.Password))
		}
		if neighbor.LocalAddress != "" {
			commands = append(commands, parsers.BuildBGPNeighborLocalAddressCommand(neighbor.ID, neighbor.LocalAddress))
		}
	}

	// Handle redistribution changes
	if config.RedistributeStatic && !current.RedistributeStatic {
		commands = append(commands, parsers.BuildBGPRedistributeCommand("static"))
	} else if !config.RedistributeStatic && current.RedistributeStatic {
		commands = append(commands, parsers.BuildDeleteBGPRedistributeCommand("static"))
	}

	if config.RedistributeConnected && !current.RedistributeConnected {
		commands = append(commands, parsers.BuildBGPRedistributeCommand("connected"))
	} else if !config.RedistributeConnected && current.RedistributeConnected {
		commands = append(commands, parsers.BuildDeleteBGPRedistributeCommand("connected"))
	}

	// Execute all commands in batch
	if len(commands) > 0 {
		output, err := s.executor.RunBatch(ctx, commands)
		if err != nil {
			return fmt.Errorf("failed to execute BGP batch commands: %w", err)
		}
		if containsError(string(output)) {
			return fmt.Errorf("BGP batch commands failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("failed to save BGP config: %w", err)
		}
	}

	return nil
}

// Reset disables BGP and removes configuration
func (s *BGPService) Reset(ctx context.Context) error {
	// Disable BGP
	commands := []string{parsers.BuildBGPUseCommand(false)}
	output, err := s.executor.RunBatch(ctx, commands)
	if err != nil {
		return fmt.Errorf("failed to disable BGP: %w", err)
	}
	if containsError(string(output)) {
		return fmt.Errorf("failed to disable BGP: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("failed to save config after BGP reset: %w", err)
		}
	}

	return nil
}

// convertToParserBGPConfig converts client BGPConfig to parser BGPConfig
func convertToParserBGPConfig(config BGPConfig) parsers.BGPConfig {
	parserConfig := parsers.BGPConfig{
		Enabled:               config.Enabled,
		ASN:                   config.ASN,
		RouterID:              config.RouterID,
		DefaultIPv4Unicast:    config.DefaultIPv4Unicast,
		LogNeighborChanges:    config.LogNeighborChanges,
		RedistributeStatic:    config.RedistributeStatic,
		RedistributeConnected: config.RedistributeConnected,
		Neighbors:             make([]parsers.BGPNeighbor, len(config.Neighbors)),
		Networks:              make([]parsers.BGPNetwork, len(config.Networks)),
	}

	for i, n := range config.Neighbors {
		parserConfig.Neighbors[i] = parsers.BGPNeighbor{
			ID:           n.ID,
			IP:           n.IP,
			RemoteAS:     n.RemoteAS,
			HoldTime:     n.HoldTime,
			Keepalive:    n.Keepalive,
			Multihop:     n.Multihop,
			Password:     n.Password,
			LocalAddress: n.LocalAddress,
		}
	}

	for i, net := range config.Networks {
		parserConfig.Networks[i] = parsers.BGPNetwork{
			Prefix: net.Prefix,
			Mask:   net.Mask,
		}
	}

	return parserConfig
}
