package client

import (
	"context"
	"fmt"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// DDNSService handles DDNS operations (NetVolante and custom providers)
type DDNSService struct {
	executor Executor
	client   *rtxClient
}

// NewDDNSService creates a new DDNS service instance
func NewDDNSService(executor Executor, client *rtxClient) *DDNSService {
	return &DDNSService{
		executor: executor,
		client:   client,
	}
}

// ============================================================================
// NetVolante DNS Operations
// ============================================================================

// GetNetVolante retrieves NetVolante DNS configuration
func (s *DDNSService) GetNetVolante(ctx context.Context) ([]NetVolanteConfig, error) {
	cmd := "show config | grep netvolante-dns"
	logging.FromContext(ctx).Debug().Str("service", "ddns").Msgf("Getting NetVolante config with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get NetVolante config: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "ddns").Msgf("NetVolante raw output: %q", string(output))

	parser := parsers.NewDDNSParser()
	parserConfigs, err := parser.ParseNetVolanteDNS(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse NetVolante config: %w", err)
	}

	// Convert parser configs to client configs
	configs := make([]NetVolanteConfig, len(parserConfigs))
	for i, pc := range parserConfigs {
		configs[i] = s.fromParserNetVolanteConfig(pc)
	}

	return configs, nil
}

// GetNetVolanteByInterface retrieves NetVolante DNS configuration for a specific interface
func (s *DDNSService) GetNetVolanteByInterface(ctx context.Context, iface string) (*NetVolanteConfig, error) {
	configs, err := s.GetNetVolante(ctx)
	if err != nil {
		return nil, err
	}

	for _, cfg := range configs {
		if cfg.Interface == iface {
			return &cfg, nil
		}
	}

	return nil, fmt.Errorf("NetVolante config not found for interface: %s", iface)
}

// ConfigureNetVolante creates NetVolante DNS configuration
func (s *DDNSService) ConfigureNetVolante(ctx context.Context, config NetVolanteConfig) error {
	// Convert to parser config for validation
	parserConfig := s.toParserNetVolanteConfig(config)

	// Validate input
	if err := parsers.ValidateNetVolanteConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid NetVolante config: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Build and execute commands
	commands := parsers.BuildNetVolanteCommand(parserConfig)
	for _, cmd := range commands {
		logging.FromContext(ctx).Debug().Str("service", "ddns").Msgf("Executing NetVolante command: %s", cmd)
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

// UpdateNetVolante updates NetVolante DNS configuration
func (s *DDNSService) UpdateNetVolante(ctx context.Context, config NetVolanteConfig) error {
	// Delete existing config first
	if err := s.DeleteNetVolante(ctx, config.Interface); err != nil {
		logging.FromContext(ctx).Debug().Str("service", "ddns").Msgf("Failed to delete existing NetVolante config (may not exist): %v", err)
	}

	// Configure new settings
	return s.ConfigureNetVolante(ctx, config)
}

// DeleteNetVolante removes NetVolante DNS configuration for an interface
func (s *DDNSService) DeleteNetVolante(ctx context.Context, iface string) error {
	if iface == "" {
		return fmt.Errorf("interface is required")
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cmd := parsers.BuildDeleteNetVolanteHostnameCommand(iface)
	logging.FromContext(ctx).Debug().Str("service", "ddns").Msgf("Deleting NetVolante config with command: %s", cmd)

	if _, err := s.executor.Run(ctx, cmd); err != nil {
		return fmt.Errorf("failed to delete NetVolante config: %w", err)
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// TriggerNetVolanteUpdate triggers a manual NetVolante DNS update
func (s *DDNSService) TriggerNetVolanteUpdate(ctx context.Context, iface string) error {
	if iface == "" {
		return fmt.Errorf("interface is required")
	}

	cmd := parsers.BuildNetVolanteGoCommand(iface)
	logging.FromContext(ctx).Debug().Str("service", "ddns").Msgf("Triggering NetVolante update with command: %s", cmd)

	if _, err := s.executor.Run(ctx, cmd); err != nil {
		return fmt.Errorf("failed to trigger NetVolante update: %w", err)
	}

	return nil
}

// ============================================================================
// Custom DDNS Operations
// ============================================================================

// GetDDNS retrieves custom DDNS configuration
func (s *DDNSService) GetDDNS(ctx context.Context) ([]DDNSServerConfig, error) {
	cmd := "show config | grep \"ddns server\""
	logging.FromContext(ctx).Debug().Str("service", "ddns").Msgf("Getting DDNS config with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get DDNS config: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "ddns").Msgf("DDNS raw output: %q", string(output))

	parser := parsers.NewDDNSParser()
	parserConfigs, err := parser.ParseDDNSConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse DDNS config: %w", err)
	}

	// Convert parser configs to client configs
	configs := make([]DDNSServerConfig, len(parserConfigs))
	for i, pc := range parserConfigs {
		configs[i] = s.fromParserDDNSServerConfig(pc)
	}

	return configs, nil
}

// GetDDNSByID retrieves custom DDNS configuration by server ID
func (s *DDNSService) GetDDNSByID(ctx context.Context, id int) (*DDNSServerConfig, error) {
	configs, err := s.GetDDNS(ctx)
	if err != nil {
		return nil, err
	}

	for _, cfg := range configs {
		if cfg.ID == id {
			return &cfg, nil
		}
	}

	return nil, fmt.Errorf("DDNS server config not found for ID: %d", id)
}

// ConfigureDDNS creates custom DDNS configuration
func (s *DDNSService) ConfigureDDNS(ctx context.Context, config DDNSServerConfig) error {
	// Convert to parser config for validation
	parserConfig := s.toParserDDNSServerConfig(config)

	// Validate input
	if err := parsers.ValidateDDNSServerConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid DDNS config: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Build and execute commands
	commands := parsers.BuildDDNSCommand(parserConfig)
	for _, cmd := range commands {
		logging.FromContext(ctx).Debug().Str("service", "ddns").Msgf("Executing DDNS command: %s", cmd)
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

// UpdateDDNS updates custom DDNS configuration
func (s *DDNSService) UpdateDDNS(ctx context.Context, config DDNSServerConfig) error {
	// Delete existing config first
	if err := s.DeleteDDNS(ctx, config.ID); err != nil {
		logging.FromContext(ctx).Debug().Str("service", "ddns").Msgf("Failed to delete existing DDNS config (may not exist): %v", err)
	}

	// Configure new settings
	return s.ConfigureDDNS(ctx, config)
}

// DeleteDDNS removes custom DDNS configuration
func (s *DDNSService) DeleteDDNS(ctx context.Context, id int) error {
	if id < 1 || id > 4 {
		return fmt.Errorf("invalid DDNS server ID: %d (must be 1-4)", id)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	commands := parsers.BuildDeleteDDNSCommand(id)
	for _, cmd := range commands {
		logging.FromContext(ctx).Debug().Str("service", "ddns").Msgf("Deleting DDNS config with command: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to delete DDNS config: %w", err)
		}
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// TriggerDDNSUpdate triggers a manual DDNS update
func (s *DDNSService) TriggerDDNSUpdate(ctx context.Context, id int) error {
	if id < 1 || id > 4 {
		return fmt.Errorf("invalid DDNS server ID: %d (must be 1-4)", id)
	}

	cmd := parsers.BuildDDNSGoCommand(id)
	logging.FromContext(ctx).Debug().Str("service", "ddns").Msgf("Triggering DDNS update with command: %s", cmd)

	if _, err := s.executor.Run(ctx, cmd); err != nil {
		return fmt.Errorf("failed to trigger DDNS update: %w", err)
	}

	return nil
}

// ============================================================================
// Status Operations
// ============================================================================

// GetNetVolanteStatus retrieves NetVolante DNS registration status
func (s *DDNSService) GetNetVolanteStatus(ctx context.Context) ([]DDNSStatus, error) {
	cmd := parsers.BuildShowNetVolanteStatusCommand()
	logging.FromContext(ctx).Debug().Str("service", "ddns").Msgf("Getting NetVolante status with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get NetVolante status: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "ddns").Msgf("NetVolante status raw output: %q", string(output))

	parser := parsers.NewDDNSParser()
	parserStatuses, err := parser.ParseDDNSStatus(string(output), "netvolante")
	if err != nil {
		return nil, fmt.Errorf("failed to parse NetVolante status: %w", err)
	}

	// Convert parser statuses to client statuses
	statuses := make([]DDNSStatus, len(parserStatuses))
	for i, ps := range parserStatuses {
		statuses[i] = s.fromParserDDNSStatus(ps)
	}

	return statuses, nil
}

// GetDDNSStatus retrieves custom DDNS registration status
func (s *DDNSService) GetDDNSStatus(ctx context.Context) ([]DDNSStatus, error) {
	cmd := parsers.BuildShowDDNSStatusCommand()
	logging.FromContext(ctx).Debug().Str("service", "ddns").Msgf("Getting DDNS status with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get DDNS status: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "ddns").Msgf("DDNS status raw output: %q", string(output))

	parser := parsers.NewDDNSParser()
	parserStatuses, err := parser.ParseDDNSStatus(string(output), "custom")
	if err != nil {
		return nil, fmt.Errorf("failed to parse DDNS status: %w", err)
	}

	// Convert parser statuses to client statuses
	statuses := make([]DDNSStatus, len(parserStatuses))
	for i, ps := range parserStatuses {
		statuses[i] = s.fromParserDDNSStatus(ps)
	}

	return statuses, nil
}

// ============================================================================
// Conversion Functions
// ============================================================================

func (s *DDNSService) toParserNetVolanteConfig(config NetVolanteConfig) parsers.NetVolanteConfig {
	return parsers.NetVolanteConfig{
		Hostname:     config.Hostname,
		Interface:    config.Interface,
		Server:       config.Server,
		Timeout:      config.Timeout,
		IPv6:         config.IPv6,
		AutoHostname: config.AutoHostname,
		Use:          config.Use,
	}
}

func (s *DDNSService) fromParserNetVolanteConfig(config parsers.NetVolanteConfig) NetVolanteConfig {
	return NetVolanteConfig{
		Hostname:     config.Hostname,
		Interface:    config.Interface,
		Server:       config.Server,
		Timeout:      config.Timeout,
		IPv6:         config.IPv6,
		AutoHostname: config.AutoHostname,
		Use:          config.Use,
	}
}

func (s *DDNSService) toParserDDNSServerConfig(config DDNSServerConfig) parsers.DDNSServerConfig {
	return parsers.DDNSServerConfig{
		ID:       config.ID,
		URL:      config.URL,
		Hostname: config.Hostname,
		Username: config.Username,
		Password: config.Password,
	}
}

func (s *DDNSService) fromParserDDNSServerConfig(config parsers.DDNSServerConfig) DDNSServerConfig {
	return DDNSServerConfig{
		ID:       config.ID,
		URL:      config.URL,
		Hostname: config.Hostname,
		Username: config.Username,
		Password: config.Password,
	}
}

func (s *DDNSService) fromParserDDNSStatus(status parsers.DDNSStatus) DDNSStatus {
	return DDNSStatus{
		Type:         status.Type,
		Interface:    status.Interface,
		ServerID:     status.ServerID,
		Hostname:     status.Hostname,
		CurrentIP:    status.CurrentIP,
		LastUpdate:   status.LastUpdate,
		Status:       status.Status,
		ErrorMessage: status.ErrorMessage,
	}
}
