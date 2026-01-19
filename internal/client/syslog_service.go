package client

import (
	"context"
	"fmt"
	"log"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// SyslogService handles syslog configuration operations
type SyslogService struct {
	executor Executor
	client   *rtxClient // Reference to the main client for save functionality
}

// NewSyslogService creates a new syslog service instance
func NewSyslogService(executor Executor, client *rtxClient) *SyslogService {
	return &SyslogService{
		executor: executor,
		client:   client,
	}
}

// Configure creates syslog configuration
func (s *SyslogService) Configure(ctx context.Context, config SyslogConfig) error {
	// Convert client.SyslogConfig to parsers.SyslogConfig
	parserConfig := s.toParserConfig(config)

	// Validate input
	if err := parsers.ValidateSyslogConfig(&parserConfig); err != nil {
		return fmt.Errorf("invalid syslog config: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Apply hosts
	for _, host := range config.Hosts {
		parserHost := parsers.SyslogHost{
			Address: host.Address,
			Port:    host.Port,
		}
		cmd := parsers.BuildSyslogHostCommand(parserHost)
		log.Printf("[DEBUG] Adding syslog host with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to add syslog host %s: %w", host.Address, err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("syslog host command failed: %s", string(output))
		}
	}

	// Apply local address if specified
	if config.LocalAddress != "" {
		cmd := parsers.BuildSyslogLocalAddressCommand(config.LocalAddress)
		log.Printf("[DEBUG] Setting syslog local address with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to set syslog local address: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("syslog local address command failed: %s", string(output))
		}
	}

	// Apply facility if specified
	if config.Facility != "" {
		cmd := parsers.BuildSyslogFacilityCommand(config.Facility)
		log.Printf("[DEBUG] Setting syslog facility with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to set syslog facility: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("syslog facility command failed: %s", string(output))
		}
	}

	// Apply log level settings
	if config.Notice {
		cmd := parsers.BuildSyslogNoticeCommand(true)
		log.Printf("[DEBUG] Enabling syslog notice with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to enable syslog notice: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("syslog notice command failed: %s", string(output))
		}
	}

	if config.Info {
		cmd := parsers.BuildSyslogInfoCommand(true)
		log.Printf("[DEBUG] Enabling syslog info with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to enable syslog info: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("syslog info command failed: %s", string(output))
		}
	}

	if config.Debug {
		cmd := parsers.BuildSyslogDebugCommand(true)
		log.Printf("[DEBUG] Enabling syslog debug with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to enable syslog debug: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("syslog debug command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("syslog config set but failed to save configuration: %w", err)
		}
	}

	return nil
}

// Get retrieves syslog configuration
func (s *SyslogService) Get(ctx context.Context) (*SyslogConfig, error) {
	cmd := parsers.BuildShowSyslogConfigCommand()
	log.Printf("[DEBUG] Getting syslog config with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get syslog config: %w", err)
	}

	log.Printf("[DEBUG] Syslog config raw output: %q", string(output))

	parser := parsers.NewSyslogParser()
	parserConfig, err := parser.ParseSyslogConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse syslog config: %w", err)
	}

	// Convert parsers.SyslogConfig to client.SyslogConfig
	config := s.fromParserConfig(*parserConfig)
	return &config, nil
}

// Update updates syslog configuration
func (s *SyslogService) Update(ctx context.Context, config SyslogConfig) error {
	// Convert client.SyslogConfig to parsers.SyslogConfig
	parserConfig := s.toParserConfig(config)

	// Validate input
	if err := parsers.ValidateSyslogConfig(&parserConfig); err != nil {
		return fmt.Errorf("invalid syslog config: %w", err)
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
		return fmt.Errorf("failed to get current syslog config: %w", err)
	}

	// Update hosts - remove hosts not in new config, add new hosts
	currentHostMap := make(map[string]SyslogHost)
	for _, host := range current.Hosts {
		currentHostMap[host.Address] = host
	}

	newHostMap := make(map[string]SyslogHost)
	for _, host := range config.Hosts {
		newHostMap[host.Address] = host
	}

	// Remove hosts that are no longer in the new config
	for address := range currentHostMap {
		if _, exists := newHostMap[address]; !exists {
			cmd := parsers.BuildDeleteSyslogHostCommand(address)
			log.Printf("[DEBUG] Removing syslog host with command: %s", cmd)
			_, _ = s.executor.Run(ctx, cmd) // Ignore errors for cleanup
		}
	}

	// Add or update hosts
	for _, host := range config.Hosts {
		currentHost, exists := currentHostMap[host.Address]
		if !exists || currentHost.Port != host.Port {
			// If host exists but port changed, remove first then add
			if exists && currentHost.Port != host.Port {
				cmd := parsers.BuildDeleteSyslogHostCommand(host.Address)
				log.Printf("[DEBUG] Removing syslog host for port update with command: %s", cmd)
				_, _ = s.executor.Run(ctx, cmd)
			}

			parserHost := parsers.SyslogHost{
				Address: host.Address,
				Port:    host.Port,
			}
			cmd := parsers.BuildSyslogHostCommand(parserHost)
			log.Printf("[DEBUG] Adding syslog host with command: %s", cmd)

			output, err := s.executor.Run(ctx, cmd)
			if err != nil {
				return fmt.Errorf("failed to add syslog host %s: %w", host.Address, err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("syslog host command failed: %s", string(output))
			}
		}
	}

	// Update local address if changed
	if config.LocalAddress != current.LocalAddress {
		if config.LocalAddress != "" {
			cmd := parsers.BuildSyslogLocalAddressCommand(config.LocalAddress)
			log.Printf("[DEBUG] Updating syslog local address with command: %s", cmd)

			output, err := s.executor.Run(ctx, cmd)
			if err != nil {
				return fmt.Errorf("failed to update syslog local address: %w", err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("syslog local address command failed: %s", string(output))
			}
		} else if current.LocalAddress != "" {
			// Remove local address setting
			cmd := parsers.BuildDeleteSyslogLocalAddressCommand()
			log.Printf("[DEBUG] Removing syslog local address with command: %s", cmd)
			_, _ = s.executor.Run(ctx, cmd)
		}
	}

	// Update facility if changed
	if config.Facility != current.Facility {
		if config.Facility != "" {
			cmd := parsers.BuildSyslogFacilityCommand(config.Facility)
			log.Printf("[DEBUG] Updating syslog facility with command: %s", cmd)

			output, err := s.executor.Run(ctx, cmd)
			if err != nil {
				return fmt.Errorf("failed to update syslog facility: %w", err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("syslog facility command failed: %s", string(output))
			}
		} else if current.Facility != "" {
			// Remove facility setting
			cmd := parsers.BuildDeleteSyslogFacilityCommand()
			log.Printf("[DEBUG] Removing syslog facility with command: %s", cmd)
			_, _ = s.executor.Run(ctx, cmd)
		}
	}

	// Update log level settings
	if config.Notice != current.Notice {
		cmd := parsers.BuildSyslogNoticeCommand(config.Notice)
		log.Printf("[DEBUG] Updating syslog notice with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to update syslog notice: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("syslog notice command failed: %s", string(output))
		}
	}

	if config.Info != current.Info {
		cmd := parsers.BuildSyslogInfoCommand(config.Info)
		log.Printf("[DEBUG] Updating syslog info with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to update syslog info: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("syslog info command failed: %s", string(output))
		}
	}

	if config.Debug != current.Debug {
		cmd := parsers.BuildSyslogDebugCommand(config.Debug)
		log.Printf("[DEBUG] Updating syslog debug with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to update syslog debug: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("syslog debug command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("syslog config updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// Reset removes syslog configuration
func (s *SyslogService) Reset(ctx context.Context) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get current configuration to know what to remove
	current, err := s.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current syslog config: %w", err)
	}

	// Build and execute delete commands
	parserConfig := s.toParserConfig(*current)
	commands := parsers.BuildDeleteSyslogCommand(&parserConfig)

	for _, cmd := range commands {
		log.Printf("[DEBUG] Resetting syslog config with command: %s", cmd)
		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			// Log but continue - some settings might not exist
			log.Printf("[DEBUG] Warning: command failed: %v", err)
			continue
		}

		if len(output) > 0 && containsError(string(output)) {
			// Log but continue
			log.Printf("[DEBUG] Warning: command output indicates error: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("syslog config reset but failed to save configuration: %w", err)
		}
	}

	return nil
}

// toParserConfig converts client.SyslogConfig to parsers.SyslogConfig
func (s *SyslogService) toParserConfig(config SyslogConfig) parsers.SyslogConfig {
	parserConfig := parsers.SyslogConfig{
		Hosts:        []parsers.SyslogHost{},
		LocalAddress: config.LocalAddress,
		Facility:     config.Facility,
		Notice:       config.Notice,
		Info:         config.Info,
		Debug:        config.Debug,
	}

	for _, host := range config.Hosts {
		parserConfig.Hosts = append(parserConfig.Hosts, parsers.SyslogHost{
			Address: host.Address,
			Port:    host.Port,
		})
	}

	return parserConfig
}

// fromParserConfig converts parsers.SyslogConfig to client.SyslogConfig
func (s *SyslogService) fromParserConfig(pc parsers.SyslogConfig) SyslogConfig {
	config := SyslogConfig{
		Hosts:        []SyslogHost{},
		LocalAddress: pc.LocalAddress,
		Facility:     pc.Facility,
		Notice:       pc.Notice,
		Info:         pc.Info,
		Debug:        pc.Debug,
	}

	for _, host := range pc.Hosts {
		config.Hosts = append(config.Hosts, SyslogHost{
			Address: host.Address,
			Port:    host.Port,
		})
	}

	return config
}
