package client

import (
	"context"
	"fmt"
	"log"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// L2TPService handles L2TP/L2TPv3 configuration operations
type L2TPService struct {
	executor Executor
	client   *rtxClient
}

// NewL2TPService creates a new L2TP service
func NewL2TPService(executor Executor, client *rtxClient) *L2TPService {
	return &L2TPService{
		executor: executor,
		client:   client,
	}
}

// Get retrieves a specific L2TP tunnel configuration
func (s *L2TPService) Get(ctx context.Context, tunnelID int) (*L2TPConfig, error) {
	tunnels, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, tunnel := range tunnels {
		if tunnel.ID == tunnelID {
			return &tunnel, nil
		}
	}

	return nil, fmt.Errorf("L2TP tunnel %d not found", tunnelID)
}

// List retrieves all L2TP tunnel configurations
func (s *L2TPService) List(ctx context.Context) ([]L2TPConfig, error) {
	cmd := parsers.BuildShowL2TPConfigCommand()
	log.Printf("[DEBUG] L2TP List: executing command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get L2TP config: %w", err)
	}

	log.Printf("[DEBUG] L2TP List: output length: %d bytes", len(output))
	if len(output) < 1000 {
		log.Printf("[DEBUG] L2TP List: full output: %s", string(output))
	} else {
		log.Printf("[DEBUG] L2TP List: output preview (first 500 chars): %s", string(output[:500]))
	}

	parser := parsers.NewL2TPParser()
	parsed, err := parser.ParseL2TPConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse L2TP config: %w", err)
	}

	log.Printf("[DEBUG] L2TP List: parsed %d tunnels", len(parsed))
	for _, t := range parsed {
		log.Printf("[DEBUG] L2TP List: tunnel ID=%d, Version=%s, Mode=%s, Enabled=%v", t.ID, t.Version, t.Mode, t.Enabled)
	}

	// Convert from parser types to client types
	tunnels := make([]L2TPConfig, len(parsed))
	for i, p := range parsed {
		tunnels[i] = convertFromParserL2TPConfig(p)
	}

	return tunnels, nil
}

// Create creates a new L2TP/L2TPv3 tunnel
func (s *L2TPService) Create(ctx context.Context, config L2TPConfig) error {
	// Validate configuration
	parserConfig := convertToParserL2TPConfig(config)
	if err := parsers.ValidateL2TPConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid L2TP config: %w", err)
	}

	commands := []string{}

	if config.Version == "l2tpv3" && config.Mode == "l2vpn" {
		// L2TPv3 L2VPN configuration
		commands = append(commands, fmt.Sprintf("tunnel select %d", config.ID))
		commands = append(commands, parsers.BuildTunnelEncapsulationCommand(config.ID, "l2tpv3"))

		if config.TunnelSource != "" && config.TunnelDest != "" {
			commands = append(commands, parsers.BuildTunnelEndpointCommand(config.TunnelSource, config.TunnelDest))
		}

		if config.L2TPv3Config != nil {
			if config.L2TPv3Config.LocalRouterID != "" {
				commands = append(commands, parsers.BuildL2TPLocalRouterIDCommand(config.L2TPv3Config.LocalRouterID))
			}
			if config.L2TPv3Config.RemoteRouterID != "" {
				commands = append(commands, parsers.BuildL2TPRemoteRouterIDCommand(config.L2TPv3Config.RemoteRouterID))
			}
			if config.L2TPv3Config.RemoteEndID != "" {
				commands = append(commands, parsers.BuildL2TPRemoteEndIDCommand(config.L2TPv3Config.RemoteEndID))
			}
		}

		if config.AlwaysOn {
			commands = append(commands, parsers.BuildL2TPAlwaysOnCommand(true))
		}

		if config.KeepaliveEnabled && config.KeepaliveConfig != nil {
			commands = append(commands, parsers.BuildL2TPKeepaliveCommand(
				config.KeepaliveConfig.Interval,
				config.KeepaliveConfig.Retry,
			))
		}

		if config.DisconnectTime > 0 {
			commands = append(commands, parsers.BuildL2TPDisconnectTimeCommand(config.DisconnectTime))
		}
	} else if config.Version == "l2tp" && config.Mode == "lns" {
		// L2TPv2 LNS configuration
		commands = append(commands, parsers.BuildL2TPServiceCommand(true))
		commands = append(commands, parsers.BuildPPSelectAnonymousCommand())
		commands = append(commands, parsers.BuildPPBindTunnelCommand(config.ID))

		if config.Authentication != nil {
			if config.Authentication.Method != "" {
				commands = append(commands, parsers.BuildPPAuthAcceptCommand(config.Authentication.Method))
			}
			if config.Authentication.Username != "" && config.Authentication.Password != "" {
				commands = append(commands, parsers.BuildPPAuthMynameCommand(
					config.Authentication.Username,
					config.Authentication.Password,
				))
			}
		}

		if config.IPPool != nil {
			commands = append(commands, parsers.BuildIPPPRemotePoolCommand(config.IPPool.Start, config.IPPool.End))
		}
	}

	// Execute all commands
	for _, cmd := range commands {
		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to execute L2TP command '%s': %w", cmd, err)
		}
		if containsError(string(output)) {
			return fmt.Errorf("L2TP command '%s' failed: %s", cmd, string(output))
		}
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save L2TP config: %w", err)
	}

	return nil
}

// Update modifies an existing L2TP tunnel
func (s *L2TPService) Update(ctx context.Context, config L2TPConfig) error {
	// Validate configuration
	parserConfig := convertToParserL2TPConfig(config)
	if err := parsers.ValidateL2TPConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid L2TP config: %w", err)
	}

	commands := []string{}

	if config.Version == "l2tpv3" {
		commands = append(commands, fmt.Sprintf("tunnel select %d", config.ID))

		if config.TunnelSource != "" && config.TunnelDest != "" {
			commands = append(commands, parsers.BuildTunnelEndpointCommand(config.TunnelSource, config.TunnelDest))
		}

		if config.L2TPv3Config != nil {
			if config.L2TPv3Config.LocalRouterID != "" {
				commands = append(commands, parsers.BuildL2TPLocalRouterIDCommand(config.L2TPv3Config.LocalRouterID))
			}
			if config.L2TPv3Config.RemoteRouterID != "" {
				commands = append(commands, parsers.BuildL2TPRemoteRouterIDCommand(config.L2TPv3Config.RemoteRouterID))
			}
		}

		commands = append(commands, parsers.BuildL2TPAlwaysOnCommand(config.AlwaysOn))

		if config.KeepaliveEnabled && config.KeepaliveConfig != nil {
			commands = append(commands, parsers.BuildL2TPKeepaliveCommand(
				config.KeepaliveConfig.Interval,
				config.KeepaliveConfig.Retry,
			))
		} else {
			commands = append(commands, parsers.BuildL2TPKeepaliveOffCommand())
		}

		if config.DisconnectTime > 0 {
			commands = append(commands, parsers.BuildL2TPDisconnectTimeCommand(config.DisconnectTime))
		}
	}

	// Execute all commands
	for _, cmd := range commands {
		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to execute L2TP command '%s': %w", cmd, err)
		}
		if containsError(string(output)) {
			return fmt.Errorf("L2TP command '%s' failed: %s", cmd, string(output))
		}
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save L2TP config: %w", err)
	}

	return nil
}

// Delete removes an L2TP tunnel
func (s *L2TPService) Delete(ctx context.Context, tunnelID int) error {
	cmd := parsers.BuildDeleteL2TPTunnelCommand(tunnelID)
	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete L2TP tunnel: %w", err)
	}
	if containsError(string(output)) {
		return fmt.Errorf("failed to delete L2TP tunnel: %s", string(output))
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save config after L2TP delete: %w", err)
	}

	return nil
}

// convertToParserL2TPConfig converts client L2TPConfig to parser L2TPConfig
func convertToParserL2TPConfig(config L2TPConfig) parsers.L2TPConfig {
	parserConfig := parsers.L2TPConfig{
		ID:               config.ID,
		Name:             config.Name,
		Version:          config.Version,
		Mode:             config.Mode,
		Shutdown:         config.Shutdown,
		TunnelSource:     config.TunnelSource,
		TunnelDest:       config.TunnelDest,
		TunnelDestType:   config.TunnelDestType,
		KeepaliveEnabled: config.KeepaliveEnabled,
		DisconnectTime:   config.DisconnectTime,
		AlwaysOn:         config.AlwaysOn,
		Enabled:          config.Enabled,
	}

	if config.Authentication != nil {
		parserConfig.Authentication = &parsers.L2TPAuth{
			Method:   config.Authentication.Method,
			Username: config.Authentication.Username,
			Password: config.Authentication.Password,
		}
	}

	if config.IPPool != nil {
		parserConfig.IPPool = &parsers.L2TPIPPool{
			Start: config.IPPool.Start,
			End:   config.IPPool.End,
		}
	}

	if config.IPsecProfile != nil {
		parserConfig.IPsecProfile = &parsers.L2TPIPsec{
			Enabled:      config.IPsecProfile.Enabled,
			PreSharedKey: config.IPsecProfile.PreSharedKey,
			TunnelID:     config.IPsecProfile.TunnelID,
		}
	}

	if config.L2TPv3Config != nil {
		parserConfig.L2TPv3Config = &parsers.L2TPv3Config{
			LocalRouterID:   config.L2TPv3Config.LocalRouterID,
			RemoteRouterID:  config.L2TPv3Config.RemoteRouterID,
			RemoteEndID:     config.L2TPv3Config.RemoteEndID,
			SessionID:       config.L2TPv3Config.SessionID,
			CookieSize:      config.L2TPv3Config.CookieSize,
			BridgeInterface: config.L2TPv3Config.BridgeInterface,
		}
		if config.L2TPv3Config.TunnelAuth != nil {
			parserConfig.L2TPv3Config.TunnelAuth = &parsers.L2TPTunnelAuth{
				Enabled:  config.L2TPv3Config.TunnelAuth.Enabled,
				Password: config.L2TPv3Config.TunnelAuth.Password,
			}
		}
	}

	if config.KeepaliveConfig != nil {
		parserConfig.KeepaliveConfig = &parsers.L2TPKeepalive{
			Interval: config.KeepaliveConfig.Interval,
			Retry:    config.KeepaliveConfig.Retry,
		}
	}

	return parserConfig
}

// convertFromParserL2TPConfig converts parser L2TPConfig to client L2TPConfig
func convertFromParserL2TPConfig(p parsers.L2TPConfig) L2TPConfig {
	config := L2TPConfig{
		ID:               p.ID,
		Name:             p.Name,
		Version:          p.Version,
		Mode:             p.Mode,
		Shutdown:         p.Shutdown,
		TunnelSource:     p.TunnelSource,
		TunnelDest:       p.TunnelDest,
		TunnelDestType:   p.TunnelDestType,
		KeepaliveEnabled: p.KeepaliveEnabled,
		DisconnectTime:   p.DisconnectTime,
		AlwaysOn:         p.AlwaysOn,
		Enabled:          p.Enabled,
	}

	if p.Authentication != nil {
		config.Authentication = &L2TPAuth{
			Method:   p.Authentication.Method,
			Username: p.Authentication.Username,
			Password: p.Authentication.Password,
		}
	}

	if p.IPPool != nil {
		config.IPPool = &L2TPIPPool{
			Start: p.IPPool.Start,
			End:   p.IPPool.End,
		}
	}

	if p.IPsecProfile != nil {
		config.IPsecProfile = &L2TPIPsec{
			Enabled:      p.IPsecProfile.Enabled,
			PreSharedKey: p.IPsecProfile.PreSharedKey,
			TunnelID:     p.IPsecProfile.TunnelID,
		}
	}

	if p.L2TPv3Config != nil {
		config.L2TPv3Config = &L2TPv3Config{
			LocalRouterID:   p.L2TPv3Config.LocalRouterID,
			RemoteRouterID:  p.L2TPv3Config.RemoteRouterID,
			RemoteEndID:     p.L2TPv3Config.RemoteEndID,
			SessionID:       p.L2TPv3Config.SessionID,
			CookieSize:      p.L2TPv3Config.CookieSize,
			BridgeInterface: p.L2TPv3Config.BridgeInterface,
		}
		if p.L2TPv3Config.TunnelAuth != nil {
			config.L2TPv3Config.TunnelAuth = &L2TPTunnelAuth{
				Enabled:  p.L2TPv3Config.TunnelAuth.Enabled,
				Password: p.L2TPv3Config.TunnelAuth.Password,
			}
		}
	}

	if p.KeepaliveConfig != nil {
		config.KeepaliveConfig = &L2TPKeepalive{
			Interval: p.KeepaliveConfig.Interval,
			Retry:    p.KeepaliveConfig.Retry,
		}
	}

	return config
}

// GetL2TPServiceState retrieves the current L2TP service state
func (s *L2TPService) GetL2TPServiceState(ctx context.Context) (*L2TPServiceState, error) {
	cmd := parsers.BuildShowL2TPConfigCommand()
	log.Printf("[DEBUG] L2TP GetServiceState: executing command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get L2TP service state: %w", err)
	}

	parsed, err := parsers.ParseL2TPServiceConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse L2TP service state: %w", err)
	}

	log.Printf("[DEBUG] L2TP GetServiceState: enabled=%v, protocols=%v", parsed.Enabled, parsed.Protocols)

	return &L2TPServiceState{
		Enabled:   parsed.Enabled,
		Protocols: parsed.Protocols,
	}, nil
}

// SetL2TPServiceState enables or disables the L2TP service with optional protocols
func (s *L2TPService) SetL2TPServiceState(ctx context.Context, enabled bool, protocols []string) error {
	cmd := parsers.BuildL2TPServiceCommandWithProtocols(enabled, protocols)
	log.Printf("[DEBUG] L2TP SetServiceState: executing command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to set L2TP service state: %w", err)
	}

	if containsError(string(output)) {
		return fmt.Errorf("failed to set L2TP service state: %s", string(output))
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save L2TP service config: %w", err)
	}

	log.Printf("[DEBUG] L2TP SetServiceState: successfully set enabled=%v, protocols=%v", enabled, protocols)
	return nil
}
