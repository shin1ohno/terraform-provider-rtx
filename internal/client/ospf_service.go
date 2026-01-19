package client

import (
	"context"
	"fmt"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// OSPFService handles OSPF configuration operations
type OSPFService struct {
	executor Executor
	client   *rtxClient
}

// NewOSPFService creates a new OSPF service
func NewOSPFService(executor Executor, client *rtxClient) *OSPFService {
	return &OSPFService{
		executor: executor,
		client:   client,
	}
}

// Get retrieves the current OSPF configuration
func (s *OSPFService) Get(ctx context.Context) (*OSPFConfig, error) {
	output, err := s.executor.Run(ctx, parsers.BuildShowOSPFConfigCommand())
	if err != nil {
		return nil, fmt.Errorf("failed to get OSPF config: %w", err)
	}

	parser := parsers.NewOSPFParser()
	parsed, err := parser.ParseOSPFConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse OSPF config: %w", err)
	}

	// Convert from parser type to client type
	config := &OSPFConfig{
		Enabled:               parsed.Enabled,
		ProcessID:             parsed.ProcessID,
		RouterID:              parsed.RouterID,
		Distance:              parsed.Distance,
		DefaultOriginate:      parsed.DefaultOriginate,
		RedistributeStatic:    parsed.RedistributeStatic,
		RedistributeConnected: parsed.RedistributeConnected,
		Networks:              make([]OSPFNetwork, len(parsed.Networks)),
		Areas:                 make([]OSPFArea, len(parsed.Areas)),
		Neighbors:             make([]OSPFNeighbor, len(parsed.Neighbors)),
	}

	for i, n := range parsed.Networks {
		config.Networks[i] = OSPFNetwork{
			IP:       n.IP,
			Wildcard: n.Wildcard,
			Area:     n.Area,
		}
	}

	for i, a := range parsed.Areas {
		config.Areas[i] = OSPFArea{
			ID:        a.ID,
			Type:      a.Type,
			NoSummary: a.NoSummary,
		}
	}

	for i, n := range parsed.Neighbors {
		config.Neighbors[i] = OSPFNeighbor{
			IP:       n.IP,
			Priority: n.Priority,
			Cost:     n.Cost,
		}
	}

	return config, nil
}

// Create creates a new OSPF configuration
func (s *OSPFService) Create(ctx context.Context, config OSPFConfig) error {
	// Validate configuration
	parserConfig := convertToParserOSPFConfig(config)
	if err := parsers.ValidateOSPFConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid OSPF config: %w", err)
	}

	commands := []string{}

	// 1. Set router ID
	commands = append(commands, parsers.BuildOSPFRouterIDCommand(config.RouterID))

	// 2. Configure areas
	for _, area := range config.Areas {
		parserArea := parsers.OSPFArea{
			ID:        area.ID,
			Type:      area.Type,
			NoSummary: area.NoSummary,
		}
		commands = append(commands, parsers.BuildOSPFAreaCommand(parserArea))
	}

	// 3. Configure networks (interface to area assignments)
	for _, network := range config.Networks {
		if network.IP != "" && network.Area != "" {
			commands = append(commands, parsers.BuildIPOSPFAreaCommand(network.IP, network.Area))
		}
	}

	// 4. Configure redistribution
	if config.RedistributeStatic {
		commands = append(commands, parsers.BuildOSPFImportCommand("static"))
	}
	if config.RedistributeConnected {
		commands = append(commands, parsers.BuildOSPFImportCommand("connected"))
	}

	// 5. Enable OSPF
	commands = append(commands, parsers.BuildOSPFEnableCommand())

	// Execute all commands
	for _, cmd := range commands {
		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to execute OSPF command '%s': %w", cmd, err)
		}
		if containsError(string(output)) {
			return fmt.Errorf("OSPF command '%s' failed: %s", cmd, string(output))
		}
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save OSPF config: %w", err)
	}

	return nil
}

// Update modifies the existing OSPF configuration
func (s *OSPFService) Update(ctx context.Context, config OSPFConfig) error {
	// Get current config
	current, err := s.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current OSPF config: %w", err)
	}

	// Validate new configuration
	parserConfig := convertToParserOSPFConfig(config)
	if err := parsers.ValidateOSPFConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid OSPF config: %w", err)
	}

	commands := []string{}

	// Update router ID if changed
	if config.RouterID != current.RouterID {
		commands = append(commands, parsers.BuildOSPFRouterIDCommand(config.RouterID))
	}

	// Remove old areas not in new config
	for _, oldArea := range current.Areas {
		found := false
		for _, newArea := range config.Areas {
			if oldArea.ID == newArea.ID {
				found = true
				break
			}
		}
		if !found {
			commands = append(commands, parsers.BuildDeleteOSPFAreaCommand(oldArea.ID))
		}
	}

	// Add/update areas
	for _, area := range config.Areas {
		parserArea := parsers.OSPFArea{
			ID:        area.ID,
			Type:      area.Type,
			NoSummary: area.NoSummary,
		}
		commands = append(commands, parsers.BuildOSPFAreaCommand(parserArea))
	}

	// Update networks
	for _, network := range config.Networks {
		if network.IP != "" && network.Area != "" {
			commands = append(commands, parsers.BuildIPOSPFAreaCommand(network.IP, network.Area))
		}
	}

	// Handle redistribution changes
	if config.RedistributeStatic && !current.RedistributeStatic {
		commands = append(commands, parsers.BuildOSPFImportCommand("static"))
	} else if !config.RedistributeStatic && current.RedistributeStatic {
		commands = append(commands, parsers.BuildDeleteOSPFImportCommand("static"))
	}

	if config.RedistributeConnected && !current.RedistributeConnected {
		commands = append(commands, parsers.BuildOSPFImportCommand("connected"))
	} else if !config.RedistributeConnected && current.RedistributeConnected {
		commands = append(commands, parsers.BuildDeleteOSPFImportCommand("connected"))
	}

	// Execute all commands
	for _, cmd := range commands {
		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to execute OSPF command '%s': %w", cmd, err)
		}
		if containsError(string(output)) {
			return fmt.Errorf("OSPF command '%s' failed: %s", cmd, string(output))
		}
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save OSPF config: %w", err)
	}

	return nil
}

// Delete disables OSPF and removes configuration
func (s *OSPFService) Delete(ctx context.Context) error {
	// Disable OSPF
	output, err := s.executor.Run(ctx, parsers.BuildOSPFDisableCommand())
	if err != nil {
		return fmt.Errorf("failed to disable OSPF: %w", err)
	}
	if containsError(string(output)) {
		return fmt.Errorf("failed to disable OSPF: %s", string(output))
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save config after OSPF delete: %w", err)
	}

	return nil
}

// Configure is an alias for Create
func (s *OSPFService) Configure(ctx context.Context, config OSPFConfig) error {
	return s.Create(ctx, config)
}

// Reset is an alias for Delete
func (s *OSPFService) Reset(ctx context.Context) error {
	return s.Delete(ctx)
}

// convertToParserOSPFConfig converts client OSPFConfig to parser OSPFConfig
func convertToParserOSPFConfig(config OSPFConfig) parsers.OSPFConfig {
	parserConfig := parsers.OSPFConfig{
		Enabled:               config.Enabled,
		ProcessID:             config.ProcessID,
		RouterID:              config.RouterID,
		Distance:              config.Distance,
		DefaultOriginate:      config.DefaultOriginate,
		RedistributeStatic:    config.RedistributeStatic,
		RedistributeConnected: config.RedistributeConnected,
		Networks:              make([]parsers.OSPFNetwork, len(config.Networks)),
		Areas:                 make([]parsers.OSPFArea, len(config.Areas)),
		Neighbors:             make([]parsers.OSPFNeighbor, len(config.Neighbors)),
	}

	for i, n := range config.Networks {
		parserConfig.Networks[i] = parsers.OSPFNetwork{
			IP:       n.IP,
			Wildcard: n.Wildcard,
			Area:     n.Area,
		}
	}

	for i, a := range config.Areas {
		parserConfig.Areas[i] = parsers.OSPFArea{
			ID:        a.ID,
			Type:      a.Type,
			NoSummary: a.NoSummary,
		}
	}

	for i, n := range config.Neighbors {
		parserConfig.Neighbors[i] = parsers.OSPFNeighbor{
			IP:       n.IP,
			Priority: n.Priority,
			Cost:     n.Cost,
		}
	}

	return parserConfig
}
