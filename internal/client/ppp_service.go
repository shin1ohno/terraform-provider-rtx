package client

import (
	"context"
	"fmt"
	"log"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// PPPService handles PPP/PPPoE operations
type PPPService struct {
	executor Executor
	client   *rtxClient
}

// NewPPPService creates a new PPP service instance
func NewPPPService(executor Executor, client *rtxClient) *PPPService {
	return &PPPService{
		executor: executor,
		client:   client,
	}
}

// ============================================================================
// PPPoE Operations
// ============================================================================

// List retrieves all PPPoE configurations
func (s *PPPService) List(ctx context.Context) ([]PPPoEConfig, error) {
	cmd := "show config"
	log.Printf("[DEBUG] Getting PPPoE configs with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get PPPoE config: %w", err)
	}

	log.Printf("[DEBUG] PPPoE raw output length: %d", len(output))

	parser := parsers.NewPPPParser()
	parserConfigs, err := parser.ParsePPPoEConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse PPPoE config: %w", err)
	}

	// Convert parser configs to client configs
	configs := make([]PPPoEConfig, len(parserConfigs))
	for i, pc := range parserConfigs {
		configs[i] = s.fromParserPPPoEConfig(pc)
	}

	return configs, nil
}

// Get retrieves PPPoE configuration by PP number
func (s *PPPService) Get(ctx context.Context, ppNum int) (*PPPoEConfig, error) {
	configs, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, cfg := range configs {
		if cfg.Number == ppNum {
			return &cfg, nil
		}
	}

	return nil, fmt.Errorf("PPPoE config not found for PP %d", ppNum)
}

// Create creates a new PPPoE configuration
func (s *PPPService) Create(ctx context.Context, config PPPoEConfig) error {
	// Convert to parser config for validation
	parserConfig := s.toParserPPPoEConfig(config)

	// Validate input
	if err := parsers.ValidatePPPoEConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid PPPoE config: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Build and execute commands
	commands := parsers.BuildPPPoECommand(parserConfig)
	for _, cmd := range commands {
		log.Printf("[DEBUG] Executing PPPoE command: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to execute command %q: %w", cmd, err)
		}
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// Update updates an existing PPPoE configuration
func (s *PPPService) Update(ctx context.Context, config PPPoEConfig) error {
	// Convert to parser config for validation
	parserConfig := s.toParserPPPoEConfig(config)

	// Validate input
	if err := parsers.ValidatePPPoEConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid PPPoE config: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Select PP interface
	selectCmd := parsers.BuildPPSelectCommand(config.Number)
	log.Printf("[DEBUG] Selecting PP interface: %s", selectCmd)
	if _, err := s.executor.Run(ctx, selectCmd); err != nil {
		return fmt.Errorf("failed to select PP interface: %w", err)
	}

	// Update description if changed
	if config.Name != "" {
		cmd := parsers.BuildPPDescriptionCommand(config.Name)
		if cmd != "" {
			log.Printf("[DEBUG] Updating description: %s", cmd)
			if _, err := s.executor.Run(ctx, cmd); err != nil {
				return fmt.Errorf("failed to set description: %w", err)
			}
		}
	}

	// Update PPPoE interface
	if config.Interface != "" {
		cmd := parsers.BuildPPPoEUseCommand(config.Interface)
		if cmd != "" {
			log.Printf("[DEBUG] Updating PPPoE interface: %s", cmd)
			if _, err := s.executor.Run(ctx, cmd); err != nil {
				return fmt.Errorf("failed to set PPPoE interface: %w", err)
			}
		}
	}

	// Update authentication
	if config.Authentication != nil {
		if config.Authentication.Method != "" {
			cmd := parsers.BuildPPPAuthAcceptCommand(config.Authentication.Method)
			if cmd != "" {
				log.Printf("[DEBUG] Updating auth accept: %s", cmd)
				if _, err := s.executor.Run(ctx, cmd); err != nil {
					return fmt.Errorf("failed to set auth accept: %w", err)
				}
			}
		}
		if config.Authentication.Username != "" && config.Authentication.Password != "" {
			cmd := parsers.BuildPPPAuthMynameCommand(config.Authentication.Username, config.Authentication.Password)
			if cmd != "" {
				log.Printf("[DEBUG] Updating auth myname")
				if _, err := s.executor.Run(ctx, cmd); err != nil {
					return fmt.Errorf("failed to set auth myname: %w", err)
				}
			}
		}
	}

	// Update always-on
	cmd := parsers.BuildPPAlwaysOnCommand(config.AlwaysOn)
	log.Printf("[DEBUG] Updating always-on: %s", cmd)
	if _, err := s.executor.Run(ctx, cmd); err != nil {
		return fmt.Errorf("failed to set always-on: %w", err)
	}

	// Update IP config
	if config.IPConfig != nil {
		if config.IPConfig.Address != "" {
			cmd := parsers.BuildIPPPAddressCommand(config.IPConfig.Address)
			if cmd != "" {
				log.Printf("[DEBUG] Updating IP address: %s", cmd)
				if _, err := s.executor.Run(ctx, cmd); err != nil {
					return fmt.Errorf("failed to set IP address: %w", err)
				}
			}
		}
		if config.IPConfig.MTU > 0 {
			cmd := parsers.BuildIPPPMTUCommand(config.IPConfig.MTU)
			if cmd != "" {
				log.Printf("[DEBUG] Updating MTU: %s", cmd)
				if _, err := s.executor.Run(ctx, cmd); err != nil {
					return fmt.Errorf("failed to set MTU: %w", err)
				}
			}
		}
		if config.IPConfig.TCPMSSLimit > 0 {
			cmd := parsers.BuildIPPPTCPMSSLimitCommand(config.IPConfig.TCPMSSLimit)
			if cmd != "" {
				log.Printf("[DEBUG] Updating TCP MSS: %s", cmd)
				if _, err := s.executor.Run(ctx, cmd); err != nil {
					return fmt.Errorf("failed to set TCP MSS: %w", err)
				}
			}
		}
		if config.IPConfig.NATDescriptor > 0 {
			cmd := parsers.BuildIPPPNATDescriptorCommand(config.IPConfig.NATDescriptor)
			if cmd != "" {
				log.Printf("[DEBUG] Updating NAT descriptor: %s", cmd)
				if _, err := s.executor.Run(ctx, cmd); err != nil {
					return fmt.Errorf("failed to set NAT descriptor: %w", err)
				}
			}
		}
	}

	// Enable/disable PP interface
	if config.Enabled {
		cmd := parsers.BuildPPEnableCommand(config.Number)
		log.Printf("[DEBUG] Enabling PP interface: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to enable PP interface: %w", err)
		}
	} else {
		cmd := parsers.BuildPPDisableCommand(config.Number)
		log.Printf("[DEBUG] Disabling PP interface: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to disable PP interface: %w", err)
		}
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// Delete removes a PPPoE configuration
func (s *PPPService) Delete(ctx context.Context, ppNum int) error {
	if ppNum < 1 {
		return fmt.Errorf("invalid PP number: %d", ppNum)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Build and execute delete commands
	commands := parsers.BuildDeletePPPoECommand(ppNum)
	for _, cmd := range commands {
		log.Printf("[DEBUG] Deleting PPPoE config with command: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			// Log but continue - some commands may fail if config doesn't exist
			log.Printf("[DEBUG] Command %q returned error (may be normal): %v", cmd, err)
		}
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// ============================================================================
// PP Interface IP Operations
// ============================================================================

// GetIPConfig retrieves PP interface IP configuration
func (s *PPPService) GetIPConfig(ctx context.Context, ppNum int) (*PPIPConfig, error) {
	cmd := "show config"
	log.Printf("[DEBUG] Getting PP IP config with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get PP IP config: %w", err)
	}

	parser := parsers.NewPPPParser()
	parserConfig, err := parser.ParsePPInterfaceConfig(string(output), ppNum)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PP IP config: %w", err)
	}

	config := s.fromParserPPIPConfig(parserConfig)
	return &config, nil
}

// ConfigureIPConfig configures PP interface IP settings
func (s *PPPService) ConfigureIPConfig(ctx context.Context, ppNum int, config PPIPConfig) error {
	// Validate input
	parserConfig := s.toParserPPIPConfig(config)
	if err := parsers.ValidatePPIPConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid PP IP config: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Select PP interface
	selectCmd := parsers.BuildPPSelectCommand(ppNum)
	log.Printf("[DEBUG] Selecting PP interface: %s", selectCmd)
	if _, err := s.executor.Run(ctx, selectCmd); err != nil {
		return fmt.Errorf("failed to select PP interface: %w", err)
	}

	// Configure IP address
	if config.Address != "" {
		cmd := parsers.BuildIPPPAddressCommand(config.Address)
		log.Printf("[DEBUG] Setting IP address: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to set IP address: %w", err)
		}
	}

	// Configure MTU
	if config.MTU > 0 {
		cmd := parsers.BuildIPPPMTUCommand(config.MTU)
		log.Printf("[DEBUG] Setting MTU: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to set MTU: %w", err)
		}
	}

	// Configure TCP MSS limit
	if config.TCPMSSLimit > 0 {
		cmd := parsers.BuildIPPPTCPMSSLimitCommand(config.TCPMSSLimit)
		log.Printf("[DEBUG] Setting TCP MSS: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to set TCP MSS: %w", err)
		}
	}

	// Configure NAT descriptor
	if config.NATDescriptor > 0 {
		cmd := parsers.BuildIPPPNATDescriptorCommand(config.NATDescriptor)
		log.Printf("[DEBUG] Setting NAT descriptor: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to set NAT descriptor: %w", err)
		}
	}

	// Configure secure filters
	if len(config.SecureFilterIn) > 0 {
		cmd := parsers.BuildIPPPSecureFilterInCommand(config.SecureFilterIn)
		log.Printf("[DEBUG] Setting secure filter in: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to set secure filter in: %w", err)
		}
	}
	if len(config.SecureFilterOut) > 0 {
		cmd := parsers.BuildIPPPSecureFilterOutCommand(config.SecureFilterOut)
		log.Printf("[DEBUG] Setting secure filter out: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to set secure filter out: %w", err)
		}
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// ============================================================================
// Connection Status Operations
// ============================================================================

// GetConnectionStatus retrieves PP interface connection status
func (s *PPPService) GetConnectionStatus(ctx context.Context, ppNum int) (*PPConnectionStatus, error) {
	cmd := fmt.Sprintf("show status pp %d", ppNum)
	log.Printf("[DEBUG] Getting PP connection status with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get PP status: %w", err)
	}

	// Parse basic status (simplified)
	status := &PPConnectionStatus{
		PPNumber:  ppNum,
		RawStatus: string(output),
	}

	// Parse connection state from output
	outputStr := string(output)
	if contains(outputStr, "PP[ON]") || contains(outputStr, "接続中") {
		status.Connected = true
		status.State = "connected"
	} else if contains(outputStr, "PP[OFF]") || contains(outputStr, "切断") {
		status.Connected = false
		status.State = "disconnected"
	} else {
		status.State = "unknown"
	}

	return status, nil
}

// ============================================================================
// Conversion Functions
// ============================================================================

func (s *PPPService) toParserPPPoEConfig(config PPPoEConfig) parsers.PPPoEConfig {
	parserConfig := parsers.PPPoEConfig{
		Number:            config.Number,
		Name:              config.Name,
		Interface:         config.Interface,
		BindInterface:     config.BindInterface,
		ServiceName:       config.ServiceName,
		ACName:            config.ACName,
		AlwaysOn:          config.AlwaysOn,
		Enabled:           config.Enabled,
		DisconnectTimeout: config.DisconnectTimeout,
	}

	if config.Authentication != nil {
		parserConfig.Authentication = &parsers.PPPAuth{
			Method:   config.Authentication.Method,
			Username: config.Authentication.Username,
			Password: config.Authentication.Password,
		}
	}

	if config.IPConfig != nil {
		parserConfig.IPConfig = &parsers.PPIPConfig{
			Address:         config.IPConfig.Address,
			MTU:             config.IPConfig.MTU,
			TCPMSSLimit:     config.IPConfig.TCPMSSLimit,
			NATDescriptor:   config.IPConfig.NATDescriptor,
			SecureFilterIn:  config.IPConfig.SecureFilterIn,
			SecureFilterOut: config.IPConfig.SecureFilterOut,
		}
	}

	return parserConfig
}

func (s *PPPService) fromParserPPPoEConfig(config parsers.PPPoEConfig) PPPoEConfig {
	clientConfig := PPPoEConfig{
		Number:            config.Number,
		Name:              config.Name,
		Interface:         config.Interface,
		BindInterface:     config.BindInterface,
		ServiceName:       config.ServiceName,
		ACName:            config.ACName,
		AlwaysOn:          config.AlwaysOn,
		Enabled:           config.Enabled,
		DisconnectTimeout: config.DisconnectTimeout,
	}

	if config.Authentication != nil {
		clientConfig.Authentication = &PPPAuth{
			Method:   config.Authentication.Method,
			Username: config.Authentication.Username,
			Password: config.Authentication.Password,
		}
	}

	if config.IPConfig != nil {
		clientConfig.IPConfig = &PPIPConfig{
			Address:         config.IPConfig.Address,
			MTU:             config.IPConfig.MTU,
			TCPMSSLimit:     config.IPConfig.TCPMSSLimit,
			NATDescriptor:   config.IPConfig.NATDescriptor,
			SecureFilterIn:  config.IPConfig.SecureFilterIn,
			SecureFilterOut: config.IPConfig.SecureFilterOut,
		}
	}

	return clientConfig
}

func (s *PPPService) toParserPPIPConfig(config PPIPConfig) parsers.PPIPConfig {
	return parsers.PPIPConfig{
		Address:         config.Address,
		MTU:             config.MTU,
		TCPMSSLimit:     config.TCPMSSLimit,
		NATDescriptor:   config.NATDescriptor,
		SecureFilterIn:  config.SecureFilterIn,
		SecureFilterOut: config.SecureFilterOut,
	}
}

func (s *PPPService) fromParserPPIPConfig(config *parsers.PPIPConfig) PPIPConfig {
	if config == nil {
		return PPIPConfig{}
	}
	return PPIPConfig{
		Address:         config.Address,
		MTU:             config.MTU,
		TCPMSSLimit:     config.TCPMSSLimit,
		NATDescriptor:   config.NATDescriptor,
		SecureFilterIn:  config.SecureFilterIn,
		SecureFilterOut: config.SecureFilterOut,
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
