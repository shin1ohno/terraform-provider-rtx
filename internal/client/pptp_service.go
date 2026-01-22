package client

import (
	"context"
	"fmt"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// PPTPService handles PPTP configuration operations
type PPTPService struct {
	executor Executor
	client   *rtxClient
}

// NewPPTPService creates a new PPTP service
func NewPPTPService(executor Executor, client *rtxClient) *PPTPService {
	return &PPTPService{
		executor: executor,
		client:   client,
	}
}

// Get retrieves the current PPTP configuration
func (s *PPTPService) Get(ctx context.Context) (*PPTPConfig, error) {
	output, err := s.executor.Run(ctx, parsers.BuildShowPPTPConfigCommand())
	if err != nil {
		return nil, fmt.Errorf("failed to get PPTP config: %w", err)
	}

	parser := parsers.NewPPTPParser()
	parsed, err := parser.ParsePPTPConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse PPTP config: %w", err)
	}

	// Convert from parser type to client type
	config := convertFromParserPPTPConfig(*parsed)
	return &config, nil
}

// Create creates a new PPTP configuration
func (s *PPTPService) Create(ctx context.Context, config PPTPConfig) error {
	// Validate configuration
	parserConfig := convertToParserPPTPConfig(config)
	if err := parsers.ValidatePPTPConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid PPTP config: %w", err)
	}

	commands := []string{}

	// Enable PPTP service
	commands = append(commands, parsers.BuildPPTPServiceCommand(true))

	// Configure authentication
	if config.Authentication != nil {
		if config.Authentication.Method != "" {
			commands = append(commands, parsers.BuildPPTPAuthAcceptCommand(config.Authentication.Method))
		}
		if config.Authentication.Username != "" && config.Authentication.Password != "" {
			commands = append(commands, parsers.BuildPPTPAuthMynameCommand(
				config.Authentication.Username,
				config.Authentication.Password,
			))
		}
	}

	// Configure MPPE encryption
	if config.Encryption != nil {
		parserEnc := parsers.PPTPEncryption{
			MPPEBits: config.Encryption.MPPEBits,
			Required: config.Encryption.Required,
		}
		commands = append(commands, parsers.BuildPPPCCPTypeCommand(parserEnc))
	}

	// Configure IP pool
	if config.IPPool != nil {
		commands = append(commands, parsers.BuildPPTPIPPoolCommand(config.IPPool.Start, config.IPPool.End))
	}

	// Configure disconnect time
	if config.DisconnectTime > 0 {
		commands = append(commands, parsers.BuildPPTPTunnelDisconnectTimeCommand(config.DisconnectTime))
	}

	// Configure keepalive
	commands = append(commands, parsers.BuildPPTPKeepaliveCommand(config.KeepaliveEnabled))

	// Execute all commands in batch
	output, err := s.executor.RunBatch(ctx, commands)
	if err != nil {
		return fmt.Errorf("failed to execute PPTP commands: %w", err)
	}
	if containsError(string(output)) {
		return fmt.Errorf("PPTP commands failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("failed to save PPTP config: %w", err)
		}
	}

	return nil
}

// Update modifies the existing PPTP configuration
func (s *PPTPService) Update(ctx context.Context, config PPTPConfig) error {
	// Validate configuration
	parserConfig := convertToParserPPTPConfig(config)
	if err := parsers.ValidatePPTPConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid PPTP config: %w", err)
	}

	commands := []string{}

	// Update authentication
	if config.Authentication != nil {
		if config.Authentication.Method != "" {
			commands = append(commands, parsers.BuildPPTPAuthAcceptCommand(config.Authentication.Method))
		}
		if config.Authentication.Username != "" && config.Authentication.Password != "" {
			commands = append(commands, parsers.BuildPPTPAuthMynameCommand(
				config.Authentication.Username,
				config.Authentication.Password,
			))
		}
	}

	// Update MPPE encryption
	if config.Encryption != nil {
		parserEnc := parsers.PPTPEncryption{
			MPPEBits: config.Encryption.MPPEBits,
			Required: config.Encryption.Required,
		}
		commands = append(commands, parsers.BuildPPPCCPTypeCommand(parserEnc))
	}

	// Update IP pool
	if config.IPPool != nil {
		commands = append(commands, parsers.BuildPPTPIPPoolCommand(config.IPPool.Start, config.IPPool.End))
	}

	// Update disconnect time
	if config.DisconnectTime > 0 {
		commands = append(commands, parsers.BuildPPTPTunnelDisconnectTimeCommand(config.DisconnectTime))
	}

	// Update keepalive
	commands = append(commands, parsers.BuildPPTPKeepaliveCommand(config.KeepaliveEnabled))

	// Execute all commands in batch
	output, err := s.executor.RunBatch(ctx, commands)
	if err != nil {
		return fmt.Errorf("failed to execute PPTP commands: %w", err)
	}
	if containsError(string(output)) {
		return fmt.Errorf("PPTP commands failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("failed to save PPTP config: %w", err)
		}
	}

	return nil
}

// Delete removes the PPTP configuration
func (s *PPTPService) Delete(ctx context.Context) error {
	commands := parsers.BuildDeletePPTPCommand()

	// Execute all commands in batch
	output, err := s.executor.RunBatch(ctx, commands)
	if err != nil {
		return fmt.Errorf("failed to execute PPTP delete commands: %w", err)
	}
	if containsError(string(output)) {
		return fmt.Errorf("PPTP delete commands failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("failed to save config after PPTP delete: %w", err)
		}
	}

	return nil
}

// convertToParserPPTPConfig converts client PPTPConfig to parser PPTPConfig
func convertToParserPPTPConfig(config PPTPConfig) parsers.PPTPConfig {
	parserConfig := parsers.PPTPConfig{
		Shutdown:         config.Shutdown,
		ListenAddress:    config.ListenAddress,
		MaxConnections:   config.MaxConnections,
		DisconnectTime:   config.DisconnectTime,
		KeepaliveEnabled: config.KeepaliveEnabled,
		Enabled:          config.Enabled,
	}

	if config.Authentication != nil {
		parserConfig.Authentication = &parsers.PPTPAuth{
			Method:   config.Authentication.Method,
			Username: config.Authentication.Username,
			Password: config.Authentication.Password,
		}
	}

	if config.Encryption != nil {
		parserConfig.Encryption = &parsers.PPTPEncryption{
			MPPEBits: config.Encryption.MPPEBits,
			Required: config.Encryption.Required,
		}
	}

	if config.IPPool != nil {
		parserConfig.IPPool = &parsers.PPTPIPPool{
			Start: config.IPPool.Start,
			End:   config.IPPool.End,
		}
	}

	return parserConfig
}

// convertFromParserPPTPConfig converts parser PPTPConfig to client PPTPConfig
func convertFromParserPPTPConfig(p parsers.PPTPConfig) PPTPConfig {
	config := PPTPConfig{
		Shutdown:         p.Shutdown,
		ListenAddress:    p.ListenAddress,
		MaxConnections:   p.MaxConnections,
		DisconnectTime:   p.DisconnectTime,
		KeepaliveEnabled: p.KeepaliveEnabled,
		Enabled:          p.Enabled,
	}

	if p.Authentication != nil {
		config.Authentication = &PPTPAuth{
			Method:   p.Authentication.Method,
			Username: p.Authentication.Username,
			Password: p.Authentication.Password,
		}
	}

	if p.Encryption != nil {
		config.Encryption = &PPTPEncryption{
			MPPEBits: p.Encryption.MPPEBits,
			Required: p.Encryption.Required,
		}
	}

	if p.IPPool != nil {
		config.IPPool = &PPTPIPPool{
			Start: p.IPPool.Start,
			End:   p.IPPool.End,
		}
	}

	return config
}
