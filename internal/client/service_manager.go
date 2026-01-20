package client

import (
	"context"
	"fmt"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// ServiceManager handles network service daemon operations (HTTPD, SSHD, SFTPD)
type ServiceManager struct {
	executor Executor
	client   *rtxClient // Reference to the main client for save functionality
}

// NewServiceManager creates a new service manager instance
func NewServiceManager(executor Executor, client *rtxClient) *ServiceManager {
	return &ServiceManager{
		executor: executor,
		client:   client,
	}
}

// ========== HTTPD Methods ==========

// GetHTTPD retrieves the current HTTPD configuration
func (s *ServiceManager) GetHTTPD(ctx context.Context) (*HTTPDConfig, error) {
	cmd := parsers.BuildShowHTTPDConfigCommand()
	logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Getting HTTPD config with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get HTTPD configuration: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("HTTPD config raw output: %q", string(output))

	parser := parsers.NewServiceParser()
	parserConfig, err := parser.ParseHTTPDConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTTPD configuration: %w", err)
	}

	// Convert parsers.HTTPDConfig to client.HTTPDConfig
	config := &HTTPDConfig{
		Host:        parserConfig.Host,
		ProxyAccess: parserConfig.ProxyAccess,
	}

	return config, nil
}

// ConfigureHTTPD creates a new HTTPD configuration
func (s *ServiceManager) ConfigureHTTPD(ctx context.Context, config HTTPDConfig) error {
	// Validate input
	parserConfig := parsers.HTTPDConfig{
		Host:        config.Host,
		ProxyAccess: config.ProxyAccess,
	}
	if err := parsers.ValidateHTTPDConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid HTTPD configuration: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Set host
	cmd := parsers.BuildHTTPDHostCommand(config.Host)
	logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Setting HTTPD host with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to set HTTPD host: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Set proxy access
	proxyCmd := parsers.BuildHTTPDProxyAccessCommand(config.ProxyAccess)
	logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Setting HTTPD proxy access with command: %s", proxyCmd)

	output, err = s.executor.Run(ctx, proxyCmd)
	if err != nil {
		return fmt.Errorf("failed to set HTTPD proxy access: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("HTTPD configured but failed to save configuration: %w", err)
		}
	}

	return nil
}

// UpdateHTTPD updates the HTTPD configuration
func (s *ServiceManager) UpdateHTTPD(ctx context.Context, config HTTPDConfig) error {
	// For HTTPD, update is the same as configure (idempotent)
	return s.ConfigureHTTPD(ctx, config)
}

// ResetHTTPD removes the HTTPD configuration
func (s *ServiceManager) ResetHTTPD(ctx context.Context) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Remove host configuration
	cmd := parsers.BuildDeleteHTTPDHostCommand()
	logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Removing HTTPD host with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to remove HTTPD host: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		// Ignore "not found" errors
		if !strings.Contains(strings.ToLower(string(output)), "not found") {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Disable proxy access
	proxyCmd := parsers.BuildDeleteHTTPDProxyAccessCommand()
	logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Disabling HTTPD proxy access with command: %s", proxyCmd)

	_, _ = s.executor.Run(ctx, proxyCmd) // Ignore errors for cleanup

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("HTTPD reset but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ========== SSHD Methods ==========

// GetSSHD retrieves the current SSHD configuration
func (s *ServiceManager) GetSSHD(ctx context.Context) (*SSHDConfig, error) {
	cmd := parsers.BuildShowSSHDConfigCommand()
	logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Getting SSHD config with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSHD configuration: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("SSHD config raw output: %q", string(output))

	parser := parsers.NewServiceParser()
	parserConfig, err := parser.ParseSSHDConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSHD configuration: %w", err)
	}

	// Convert parsers.SSHDConfig to client.SSHDConfig
	config := &SSHDConfig{
		Enabled: parserConfig.Enabled,
		Hosts:   parserConfig.Hosts,
		HostKey: parserConfig.HostKey,
	}

	return config, nil
}

// ConfigureSSHD creates a new SSHD configuration
func (s *ServiceManager) ConfigureSSHD(ctx context.Context, config SSHDConfig) error {
	// Validate input
	parserConfig := parsers.SSHDConfig{
		Enabled: config.Enabled,
		Hosts:   config.Hosts,
		HostKey: config.HostKey,
	}
	if err := parsers.ValidateSSHDConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid SSHD configuration: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Set hosts if specified
	if len(config.Hosts) > 0 {
		cmd := parsers.BuildSSHDHostCommand(config.Hosts)
		logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Setting SSHD hosts with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to set SSHD hosts: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Enable/disable service
	serviceCmd := parsers.BuildSSHDServiceCommand(config.Enabled)
	logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Setting SSHD service with command: %s", serviceCmd)

	output, err := s.executor.Run(ctx, serviceCmd)
	if err != nil {
		return fmt.Errorf("failed to set SSHD service: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("SSHD configured but failed to save configuration: %w", err)
		}
	}

	return nil
}

// UpdateSSHD updates the SSHD configuration
func (s *ServiceManager) UpdateSSHD(ctx context.Context, config SSHDConfig) error {
	// Get current config for comparison
	currentConfig, err := s.GetSSHD(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current SSHD configuration: %w", err)
	}

	// Validate input
	parserConfig := parsers.SSHDConfig{
		Enabled: config.Enabled,
		Hosts:   config.Hosts,
		HostKey: config.HostKey,
	}
	if err := parsers.ValidateSSHDConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid SSHD configuration: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Update hosts if changed
	hostsChanged := !stringSliceEqual(currentConfig.Hosts, config.Hosts)
	if hostsChanged {
		// Remove old hosts first if there were any
		if len(currentConfig.Hosts) > 0 {
			deleteCmd := parsers.BuildDeleteSSHDHostCommand()
			logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Removing old SSHD hosts with command: %s", deleteCmd)
			_, _ = s.executor.Run(ctx, deleteCmd) // Ignore errors for cleanup
		}

		// Set new hosts if specified
		if len(config.Hosts) > 0 {
			cmd := parsers.BuildSSHDHostCommand(config.Hosts)
			logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Setting SSHD hosts with command: %s", cmd)

			output, err := s.executor.Run(ctx, cmd)
			if err != nil {
				return fmt.Errorf("failed to set SSHD hosts: %w", err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("command failed: %s", string(output))
			}
		}
	}

	// Update service state if changed
	if currentConfig.Enabled != config.Enabled {
		serviceCmd := parsers.BuildSSHDServiceCommand(config.Enabled)
		logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Setting SSHD service with command: %s", serviceCmd)

		output, err := s.executor.Run(ctx, serviceCmd)
		if err != nil {
			return fmt.Errorf("failed to set SSHD service: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("SSHD updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ResetSSHD removes the SSHD configuration (disables service)
func (s *ServiceManager) ResetSSHD(ctx context.Context) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Disable service first
	serviceCmd := parsers.BuildDeleteSSHDServiceCommand()
	logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Disabling SSHD service with command: %s", serviceCmd)

	output, err := s.executor.Run(ctx, serviceCmd)
	if err != nil {
		return fmt.Errorf("failed to disable SSHD service: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		// Ignore "not found" errors
		if !strings.Contains(strings.ToLower(string(output)), "not found") {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Remove host configuration
	hostCmd := parsers.BuildDeleteSSHDHostCommand()
	logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Removing SSHD hosts with command: %s", hostCmd)

	_, _ = s.executor.Run(ctx, hostCmd) // Ignore errors for cleanup

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("SSHD reset but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ========== SFTPD Methods ==========

// GetSFTPD retrieves the current SFTPD configuration
func (s *ServiceManager) GetSFTPD(ctx context.Context) (*SFTPDConfig, error) {
	cmd := parsers.BuildShowSFTPDConfigCommand()
	logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Getting SFTPD config with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get SFTPD configuration: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("SFTPD config raw output: %q", string(output))

	parser := parsers.NewServiceParser()
	parserConfig, err := parser.ParseSFTPDConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse SFTPD configuration: %w", err)
	}

	// Convert parsers.SFTPDConfig to client.SFTPDConfig
	config := &SFTPDConfig{
		Hosts: parserConfig.Hosts,
	}

	return config, nil
}

// ConfigureSFTPD creates a new SFTPD configuration
func (s *ServiceManager) ConfigureSFTPD(ctx context.Context, config SFTPDConfig) error {
	// Validate input
	parserConfig := parsers.SFTPDConfig{
		Hosts: config.Hosts,
	}
	if err := parsers.ValidateSFTPDConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid SFTPD configuration: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Set hosts
	cmd := parsers.BuildSFTPDHostCommand(config.Hosts)
	logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Setting SFTPD hosts with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to set SFTPD hosts: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("SFTPD configured but failed to save configuration: %w", err)
		}
	}

	return nil
}

// UpdateSFTPD updates the SFTPD configuration
func (s *ServiceManager) UpdateSFTPD(ctx context.Context, config SFTPDConfig) error {
	// Get current config for comparison
	currentConfig, err := s.GetSFTPD(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current SFTPD configuration: %w", err)
	}

	// Validate input
	parserConfig := parsers.SFTPDConfig{
		Hosts: config.Hosts,
	}
	if err := parsers.ValidateSFTPDConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid SFTPD configuration: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Update hosts if changed
	hostsChanged := !stringSliceEqual(currentConfig.Hosts, config.Hosts)
	if hostsChanged {
		// Remove old hosts first if there were any
		if len(currentConfig.Hosts) > 0 {
			deleteCmd := parsers.BuildDeleteSFTPDHostCommand()
			logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Removing old SFTPD hosts with command: %s", deleteCmd)
			_, _ = s.executor.Run(ctx, deleteCmd) // Ignore errors for cleanup
		}

		// Set new hosts
		cmd := parsers.BuildSFTPDHostCommand(config.Hosts)
		logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Setting SFTPD hosts with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to set SFTPD hosts: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("SFTPD updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ResetSFTPD removes the SFTPD configuration
func (s *ServiceManager) ResetSFTPD(ctx context.Context) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Remove host configuration
	cmd := parsers.BuildDeleteSFTPDHostCommand()
	logging.FromContext(ctx).Debug().Str("component", "service-manager").Msgf("Removing SFTPD hosts with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to remove SFTPD hosts: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		// Ignore "not found" errors
		if !strings.Contains(strings.ToLower(string(output)), "not found") {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("SFTPD reset but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ========== Helper Functions ==========

// stringSliceEqual compares two string slices for equality
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
