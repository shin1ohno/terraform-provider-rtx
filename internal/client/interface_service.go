package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// InterfaceService handles interface configuration operations
type InterfaceService struct {
	executor Executor
	client   *rtxClient // Reference to the main client for save functionality
}

// NewInterfaceService creates a new interface service instance
func NewInterfaceService(executor Executor, client *rtxClient) *InterfaceService {
	return &InterfaceService{
		executor: executor,
		client:   client,
	}
}

// Configure creates a new interface configuration
func (s *InterfaceService) Configure(ctx context.Context, config InterfaceConfig) error {
	// Convert client.InterfaceConfig to parsers.InterfaceConfig
	parserConfig := s.toParserConfig(config)

	// Validate input
	if err := parsers.ValidateInterfaceConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid interface configuration: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Configure IP address
	if config.IPAddress != nil {
		ipCmd := parsers.BuildIPAddressCommand(config.Name, parsers.InterfaceIP{
			Address: config.IPAddress.Address,
			DHCP:    config.IPAddress.DHCP,
		})
		if ipCmd != "" {
			logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Setting IP address with command: %s", ipCmd)
			output, err := s.executor.Run(ctx, ipCmd)
			if err != nil {
				return fmt.Errorf("failed to set IP address: %w", err)
			}
			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("IP address command failed: %s", string(output))
			}
		}
	}

	// Configure description
	if config.Description != "" {
		descCmd := parsers.BuildDescriptionCommand(config.Name, config.Description)
		logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Setting description with command: %s", descCmd)
		output, err := s.executor.Run(ctx, descCmd)
		if err != nil {
			return fmt.Errorf("failed to set description: %w", err)
		}
		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("description command failed: %s", string(output))
		}
	}

	// Note: Access list bindings are managed by separate ACL resources
	// (rtx_interface_acl, rtx_interface_mac_acl)

	// Configure NAT descriptor
	if config.NATDescriptor > 0 {
		natCmd := parsers.BuildNATDescriptorCommand(config.Name, config.NATDescriptor)
		logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Setting NAT descriptor with command: %s", natCmd)
		output, err := s.executor.Run(ctx, natCmd)
		if err != nil {
			return fmt.Errorf("failed to set NAT descriptor: %w", err)
		}
		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("NAT descriptor command failed: %s", string(output))
		}
	}

	// Configure ProxyARP
	if config.ProxyARP {
		proxyCmd := parsers.BuildProxyARPCommand(config.Name, true)
		logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Enabling ProxyARP with command: %s", proxyCmd)
		output, err := s.executor.Run(ctx, proxyCmd)
		if err != nil {
			return fmt.Errorf("failed to enable ProxyARP: %w", err)
		}
		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("ProxyARP command failed: %s", string(output))
		}
	}

	// Configure MTU
	if config.MTU > 0 {
		mtuCmd := parsers.BuildMTUCommand(config.Name, config.MTU)
		logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Setting MTU with command: %s", mtuCmd)
		output, err := s.executor.Run(ctx, mtuCmd)
		if err != nil {
			return fmt.Errorf("failed to set MTU: %w", err)
		}
		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("MTU command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("interface configured but failed to save configuration: %w", err)
		}
	}

	return nil
}

// Get retrieves an interface configuration
func (s *InterfaceService) Get(ctx context.Context, interfaceName string) (*InterfaceConfig, error) {
	// Validate interface name
	if err := parsers.ValidateInterfaceName(interfaceName); err != nil {
		return nil, err
	}

	cmd := parsers.BuildShowInterfaceConfigCommand(interfaceName)
	logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Getting interface config with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get interface configuration: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Interface config raw output: %q", string(output))

	parserConfig, err := parsers.ParseInterfaceConfig(string(output), interfaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse interface configuration: %w", err)
	}

	// Convert parsers.InterfaceConfig to client.InterfaceConfig
	config := s.fromParserConfig(*parserConfig)
	return &config, nil
}

// Update updates an existing interface configuration
func (s *InterfaceService) Update(ctx context.Context, config InterfaceConfig) error {
	// Convert client.InterfaceConfig to parsers.InterfaceConfig
	parserConfig := s.toParserConfig(config)

	// Validate input
	if err := parsers.ValidateInterfaceConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid interface configuration: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get current configuration
	currentConfig, err := s.Get(ctx, config.Name)
	if err != nil {
		// If not found, treat as new configuration
		if strings.Contains(err.Error(), "not found") {
			return s.Configure(ctx, config)
		}
		return fmt.Errorf("failed to get current configuration: %w", err)
	}

	// Update IP address if changed
	if s.ipAddressChanged(currentConfig.IPAddress, config.IPAddress) {
		// Remove old IP configuration first
		if currentConfig.IPAddress != nil {
			deleteCmd := parsers.BuildDeleteIPAddressCommand(config.Name)
			logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Removing old IP address with command: %s", deleteCmd)
			_, _ = s.executor.Run(ctx, deleteCmd)
		}

		// Set new IP address
		if config.IPAddress != nil {
			ipCmd := parsers.BuildIPAddressCommand(config.Name, parsers.InterfaceIP{
				Address: config.IPAddress.Address,
				DHCP:    config.IPAddress.DHCP,
			})
			if ipCmd != "" {
				logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Setting IP address with command: %s", ipCmd)
				output, err := s.executor.Run(ctx, ipCmd)
				if err != nil {
					return fmt.Errorf("failed to set IP address: %w", err)
				}
				if len(output) > 0 && containsError(string(output)) {
					return fmt.Errorf("IP address command failed: %s", string(output))
				}
			}
		}
	}

	// Update description if changed
	if currentConfig.Description != config.Description {
		if config.Description != "" {
			descCmd := parsers.BuildDescriptionCommand(config.Name, config.Description)
			logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Setting description with command: %s", descCmd)
			output, err := s.executor.Run(ctx, descCmd)
			if err != nil {
				return fmt.Errorf("failed to set description: %w", err)
			}
			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("description command failed: %s", string(output))
			}
		} else {
			deleteCmd := parsers.BuildDeleteDescriptionCommand(config.Name)
			logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Removing description with command: %s", deleteCmd)
			_, _ = s.executor.Run(ctx, deleteCmd)
		}
	}

	// Note: Access list bindings are managed by separate ACL resources
	// (rtx_interface_acl, rtx_interface_mac_acl)

	// Update NAT descriptor if changed
	if currentConfig.NATDescriptor != config.NATDescriptor {
		if currentConfig.NATDescriptor > 0 {
			deleteCmd := parsers.BuildDeleteNATDescriptorCommand(config.Name)
			logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Removing old NAT descriptor with command: %s", deleteCmd)
			_, _ = s.executor.Run(ctx, deleteCmd)
		}
		if config.NATDescriptor > 0 {
			natCmd := parsers.BuildNATDescriptorCommand(config.Name, config.NATDescriptor)
			logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Setting NAT descriptor with command: %s", natCmd)
			output, err := s.executor.Run(ctx, natCmd)
			if err != nil {
				return fmt.Errorf("failed to set NAT descriptor: %w", err)
			}
			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("NAT descriptor command failed: %s", string(output))
			}
		}
	}

	// Update ProxyARP if changed
	if currentConfig.ProxyARP != config.ProxyARP {
		proxyCmd := parsers.BuildProxyARPCommand(config.Name, config.ProxyARP)
		logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Setting ProxyARP with command: %s", proxyCmd)
		output, err := s.executor.Run(ctx, proxyCmd)
		if err != nil {
			return fmt.Errorf("failed to set ProxyARP: %w", err)
		}
		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("ProxyARP command failed: %s", string(output))
		}
	}

	// Update MTU if changed
	if currentConfig.MTU != config.MTU {
		if currentConfig.MTU > 0 {
			deleteCmd := parsers.BuildDeleteMTUCommand(config.Name)
			logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Removing old MTU with command: %s", deleteCmd)
			_, _ = s.executor.Run(ctx, deleteCmd)
		}
		if config.MTU > 0 {
			mtuCmd := parsers.BuildMTUCommand(config.Name, config.MTU)
			logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Setting MTU with command: %s", mtuCmd)
			output, err := s.executor.Run(ctx, mtuCmd)
			if err != nil {
				return fmt.Errorf("failed to set MTU: %w", err)
			}
			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("MTU command failed: %s", string(output))
			}
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("interface updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// Reset removes interface configuration (resets to defaults)
func (s *InterfaceService) Reset(ctx context.Context, interfaceName string) error {
	// Validate interface name
	if err := parsers.ValidateInterfaceName(interfaceName); err != nil {
		return err
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Remove IP address
	ipCmd := parsers.BuildDeleteIPAddressCommand(interfaceName)
	logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Removing IP address with command: %s", ipCmd)
	_, _ = s.executor.Run(ctx, ipCmd)

	// Remove description
	descCmd := parsers.BuildDeleteDescriptionCommand(interfaceName)
	logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Removing description with command: %s", descCmd)
	_, _ = s.executor.Run(ctx, descCmd)

	// Note: Access list bindings are managed by separate ACL resources
	// (rtx_interface_acl, rtx_interface_mac_acl)

	// Remove NAT descriptor
	natCmd := parsers.BuildDeleteNATDescriptorCommand(interfaceName)
	logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Removing NAT descriptor with command: %s", natCmd)
	_, _ = s.executor.Run(ctx, natCmd)

	// Disable ProxyARP
	proxyCmd := parsers.BuildProxyARPCommand(interfaceName, false)
	logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Disabling ProxyARP with command: %s", proxyCmd)
	_, _ = s.executor.Run(ctx, proxyCmd)

	// Remove MTU
	mtuCmd := parsers.BuildDeleteMTUCommand(interfaceName)
	logging.FromContext(ctx).Debug().Str("service", "interface").Msgf("Removing MTU with command: %s", mtuCmd)
	_, _ = s.executor.Run(ctx, mtuCmd)

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("interface reset but failed to save configuration: %w", err)
		}
	}

	return nil
}

// List retrieves all interface configurations
func (s *InterfaceService) List(ctx context.Context) ([]InterfaceConfig, error) {
	// Get configuration for common interface names
	interfaces := []string{"lan1", "lan2", "lan3", "bridge1", "pp1", "tunnel1"}
	var configs []InterfaceConfig

	for _, iface := range interfaces {
		config, err := s.Get(ctx, iface)
		if err != nil {
			// Skip interfaces that don't have configuration
			continue
		}
		// Only include interfaces with actual configuration
		if config.IPAddress != nil || config.Description != "" ||
			config.NATDescriptor > 0 || config.ProxyARP || config.MTU > 0 {
			configs = append(configs, *config)
		}
	}

	return configs, nil
}

// toParserConfig converts client.InterfaceConfig to parsers.InterfaceConfig
func (s *InterfaceService) toParserConfig(config InterfaceConfig) parsers.InterfaceConfig {
	parserConfig := parsers.InterfaceConfig{
		Name:          config.Name,
		Description:   config.Description,
		NATDescriptor: config.NATDescriptor,
		ProxyARP:      config.ProxyARP,
		MTU:           config.MTU,
		// Note: Access list fields are managed by separate ACL resources
	}

	if config.IPAddress != nil {
		parserConfig.IPAddress = &parsers.InterfaceIP{
			Address: config.IPAddress.Address,
			DHCP:    config.IPAddress.DHCP,
		}
	}

	return parserConfig
}

// fromParserConfig converts parsers.InterfaceConfig to client.InterfaceConfig
func (s *InterfaceService) fromParserConfig(pc parsers.InterfaceConfig) InterfaceConfig {
	config := InterfaceConfig{
		Name:          pc.Name,
		Description:   pc.Description,
		NATDescriptor: pc.NATDescriptor,
		ProxyARP:      pc.ProxyARP,
		MTU:           pc.MTU,
		// Note: Access list fields are managed by separate ACL resources
	}

	if pc.IPAddress != nil {
		config.IPAddress = &InterfaceIP{
			Address: pc.IPAddress.Address,
			DHCP:    pc.IPAddress.DHCP,
		}
	}

	return config
}

// ipAddressChanged checks if IP address configuration has changed
func (s *InterfaceService) ipAddressChanged(old, new *InterfaceIP) bool {
	if old == nil && new == nil {
		return false
	}
	if old == nil || new == nil {
		return true
	}
	return old.Address != new.Address || old.DHCP != new.DHCP
}
