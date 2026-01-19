package client

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// IPv6InterfaceService handles IPv6 interface configuration operations
type IPv6InterfaceService struct {
	executor Executor
	client   *rtxClient // Reference to the main client for save functionality
}

// NewIPv6InterfaceService creates a new IPv6 interface service instance
func NewIPv6InterfaceService(executor Executor, client *rtxClient) *IPv6InterfaceService {
	return &IPv6InterfaceService{
		executor: executor,
		client:   client,
	}
}

// Configure creates a new IPv6 interface configuration
func (s *IPv6InterfaceService) Configure(ctx context.Context, config IPv6InterfaceConfig) error {
	// Convert client.IPv6InterfaceConfig to parsers.IPv6InterfaceConfig
	parserConfig := s.toParserConfig(config)

	// Validate input
	if err := parsers.ValidateIPv6InterfaceConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid IPv6 interface configuration: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Configure IPv6 addresses
	for _, addr := range config.Addresses {
		addrCmd := parsers.BuildIPv6AddressCommand(config.Interface, parsers.IPv6Address{
			Address:     addr.Address,
			PrefixRef:   addr.PrefixRef,
			InterfaceID: addr.InterfaceID,
		})
		if addrCmd != "" {
			log.Printf("[DEBUG] Setting IPv6 address with command: %s", addrCmd)
			output, err := s.executor.Run(ctx, addrCmd)
			if err != nil {
				return fmt.Errorf("failed to set IPv6 address: %w", err)
			}
			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("IPv6 address command failed: %s", string(output))
			}
		}
	}

	// Configure RTADV
	if config.RTADV != nil && config.RTADV.Enabled {
		rtadvCmd := parsers.BuildIPv6RTADVCommand(config.Interface, parsers.RTADVConfig{
			Enabled:  config.RTADV.Enabled,
			PrefixID: config.RTADV.PrefixID,
			OFlag:    config.RTADV.OFlag,
			MFlag:    config.RTADV.MFlag,
			Lifetime: config.RTADV.Lifetime,
		})
		log.Printf("[DEBUG] Setting RTADV with command: %s", rtadvCmd)
		output, err := s.executor.Run(ctx, rtadvCmd)
		if err != nil {
			return fmt.Errorf("failed to set RTADV: %w", err)
		}
		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("RTADV command failed: %s", string(output))
		}
	}

	// Configure DHCPv6 service
	if config.DHCPv6Service != "" && config.DHCPv6Service != "off" {
		dhcpCmd := parsers.BuildIPv6DHCPv6Command(config.Interface, config.DHCPv6Service)
		log.Printf("[DEBUG] Setting DHCPv6 service with command: %s", dhcpCmd)
		output, err := s.executor.Run(ctx, dhcpCmd)
		if err != nil {
			return fmt.Errorf("failed to set DHCPv6 service: %w", err)
		}
		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("DHCPv6 service command failed: %s", string(output))
		}
	}

	// Configure MTU
	if config.MTU > 0 {
		mtuCmd := parsers.BuildIPv6MTUCommand(config.Interface, config.MTU)
		log.Printf("[DEBUG] Setting IPv6 MTU with command: %s", mtuCmd)
		output, err := s.executor.Run(ctx, mtuCmd)
		if err != nil {
			return fmt.Errorf("failed to set IPv6 MTU: %w", err)
		}
		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("IPv6 MTU command failed: %s", string(output))
		}
	}

	// Configure inbound security filter
	if len(config.SecureFilterIn) > 0 {
		filterCmd := parsers.BuildIPv6SecureFilterInCommand(config.Interface, config.SecureFilterIn)
		log.Printf("[DEBUG] Setting IPv6 inbound filter with command: %s", filterCmd)
		output, err := s.executor.Run(ctx, filterCmd)
		if err != nil {
			return fmt.Errorf("failed to set IPv6 inbound filter: %w", err)
		}
		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("IPv6 inbound filter command failed: %s", string(output))
		}
	}

	// Configure outbound security filter (with optional dynamic filters)
	if len(config.SecureFilterOut) > 0 {
		filterCmd := parsers.BuildIPv6SecureFilterOutCommand(config.Interface, config.SecureFilterOut, config.DynamicFilterOut)
		log.Printf("[DEBUG] Setting IPv6 outbound filter with command: %s", filterCmd)
		output, err := s.executor.Run(ctx, filterCmd)
		if err != nil {
			return fmt.Errorf("failed to set IPv6 outbound filter: %w", err)
		}
		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("IPv6 outbound filter command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("IPv6 interface configured but failed to save configuration: %w", err)
		}
	}

	return nil
}

// Get retrieves an IPv6 interface configuration
func (s *IPv6InterfaceService) Get(ctx context.Context, interfaceName string) (*IPv6InterfaceConfig, error) {
	// Validate interface name
	if err := parsers.ValidateIPv6InterfaceName(interfaceName); err != nil {
		return nil, err
	}

	cmd := parsers.BuildShowIPv6InterfaceConfigCommand(interfaceName)
	log.Printf("[DEBUG] Getting IPv6 interface config with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPv6 interface configuration: %w", err)
	}

	log.Printf("[DEBUG] IPv6 interface config raw output: %q", string(output))

	parserConfig, err := parsers.ParseIPv6InterfaceConfig(string(output), interfaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse IPv6 interface configuration: %w", err)
	}

	// Convert parsers.IPv6InterfaceConfig to client.IPv6InterfaceConfig
	config := s.fromParserConfig(*parserConfig)
	return &config, nil
}

// Update updates an existing IPv6 interface configuration
func (s *IPv6InterfaceService) Update(ctx context.Context, config IPv6InterfaceConfig) error {
	// Convert client.IPv6InterfaceConfig to parsers.IPv6InterfaceConfig
	parserConfig := s.toParserConfig(config)

	// Validate input
	if err := parsers.ValidateIPv6InterfaceConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid IPv6 interface configuration: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get current configuration
	currentConfig, err := s.Get(ctx, config.Interface)
	if err != nil {
		// If not found, treat as new configuration
		if strings.Contains(err.Error(), "not found") {
			return s.Configure(ctx, config)
		}
		return fmt.Errorf("failed to get current configuration: %w", err)
	}

	// Update addresses
	if !ipv6AddressesEqual(currentConfig.Addresses, config.Addresses) {
		// Remove old addresses first
		for _, oldAddr := range currentConfig.Addresses {
			parserAddr := &parsers.IPv6Address{
				Address:     oldAddr.Address,
				PrefixRef:   oldAddr.PrefixRef,
				InterfaceID: oldAddr.InterfaceID,
			}
			deleteCmd := parsers.BuildDeleteIPv6AddressCommand(config.Interface, parserAddr)
			log.Printf("[DEBUG] Removing old IPv6 address with command: %s", deleteCmd)
			_, _ = s.executor.Run(ctx, deleteCmd)
		}
		// Add new addresses
		for _, addr := range config.Addresses {
			addrCmd := parsers.BuildIPv6AddressCommand(config.Interface, parsers.IPv6Address{
				Address:     addr.Address,
				PrefixRef:   addr.PrefixRef,
				InterfaceID: addr.InterfaceID,
			})
			if addrCmd != "" {
				log.Printf("[DEBUG] Setting IPv6 address with command: %s", addrCmd)
				output, err := s.executor.Run(ctx, addrCmd)
				if err != nil {
					return fmt.Errorf("failed to set IPv6 address: %w", err)
				}
				if len(output) > 0 && containsError(string(output)) {
					return fmt.Errorf("IPv6 address command failed: %s", string(output))
				}
			}
		}
	}

	// Update RTADV
	if !rtadvConfigsEqual(currentConfig.RTADV, config.RTADV) {
		// Remove old RTADV if it was configured
		if currentConfig.RTADV != nil && currentConfig.RTADV.Enabled {
			deleteCmd := parsers.BuildDeleteIPv6RTADVCommand(config.Interface)
			log.Printf("[DEBUG] Removing old RTADV with command: %s", deleteCmd)
			_, _ = s.executor.Run(ctx, deleteCmd)
		}
		// Set new RTADV
		if config.RTADV != nil && config.RTADV.Enabled {
			rtadvCmd := parsers.BuildIPv6RTADVCommand(config.Interface, parsers.RTADVConfig{
				Enabled:  config.RTADV.Enabled,
				PrefixID: config.RTADV.PrefixID,
				OFlag:    config.RTADV.OFlag,
				MFlag:    config.RTADV.MFlag,
				Lifetime: config.RTADV.Lifetime,
			})
			log.Printf("[DEBUG] Setting RTADV with command: %s", rtadvCmd)
			output, err := s.executor.Run(ctx, rtadvCmd)
			if err != nil {
				return fmt.Errorf("failed to set RTADV: %w", err)
			}
			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("RTADV command failed: %s", string(output))
			}
		}
	}

	// Update DHCPv6 service
	if currentConfig.DHCPv6Service != config.DHCPv6Service {
		// Remove old DHCPv6 service
		if currentConfig.DHCPv6Service != "" {
			deleteCmd := parsers.BuildDeleteIPv6DHCPv6Command(config.Interface)
			log.Printf("[DEBUG] Removing old DHCPv6 service with command: %s", deleteCmd)
			_, _ = s.executor.Run(ctx, deleteCmd)
		}
		// Set new DHCPv6 service
		if config.DHCPv6Service != "" && config.DHCPv6Service != "off" {
			dhcpCmd := parsers.BuildIPv6DHCPv6Command(config.Interface, config.DHCPv6Service)
			log.Printf("[DEBUG] Setting DHCPv6 service with command: %s", dhcpCmd)
			output, err := s.executor.Run(ctx, dhcpCmd)
			if err != nil {
				return fmt.Errorf("failed to set DHCPv6 service: %w", err)
			}
			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("DHCPv6 service command failed: %s", string(output))
			}
		}
	}

	// Update MTU
	if currentConfig.MTU != config.MTU {
		if currentConfig.MTU > 0 {
			deleteCmd := parsers.BuildDeleteIPv6MTUCommand(config.Interface)
			log.Printf("[DEBUG] Removing old IPv6 MTU with command: %s", deleteCmd)
			_, _ = s.executor.Run(ctx, deleteCmd)
		}
		if config.MTU > 0 {
			mtuCmd := parsers.BuildIPv6MTUCommand(config.Interface, config.MTU)
			log.Printf("[DEBUG] Setting IPv6 MTU with command: %s", mtuCmd)
			output, err := s.executor.Run(ctx, mtuCmd)
			if err != nil {
				return fmt.Errorf("failed to set IPv6 MTU: %w", err)
			}
			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("IPv6 MTU command failed: %s", string(output))
			}
		}
	}

	// Update inbound security filter
	if !intSliceEqual(currentConfig.SecureFilterIn, config.SecureFilterIn) {
		if len(currentConfig.SecureFilterIn) > 0 {
			deleteCmd := parsers.BuildDeleteIPv6SecureFilterCommand(config.Interface, "in")
			log.Printf("[DEBUG] Removing old IPv6 inbound filter with command: %s", deleteCmd)
			_, _ = s.executor.Run(ctx, deleteCmd)
		}
		if len(config.SecureFilterIn) > 0 {
			filterCmd := parsers.BuildIPv6SecureFilterInCommand(config.Interface, config.SecureFilterIn)
			log.Printf("[DEBUG] Setting IPv6 inbound filter with command: %s", filterCmd)
			output, err := s.executor.Run(ctx, filterCmd)
			if err != nil {
				return fmt.Errorf("failed to set IPv6 inbound filter: %w", err)
			}
			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("IPv6 inbound filter command failed: %s", string(output))
			}
		}
	}

	// Update outbound security filter
	if !intSliceEqual(currentConfig.SecureFilterOut, config.SecureFilterOut) ||
		!intSliceEqual(currentConfig.DynamicFilterOut, config.DynamicFilterOut) {
		if len(currentConfig.SecureFilterOut) > 0 {
			deleteCmd := parsers.BuildDeleteIPv6SecureFilterCommand(config.Interface, "out")
			log.Printf("[DEBUG] Removing old IPv6 outbound filter with command: %s", deleteCmd)
			_, _ = s.executor.Run(ctx, deleteCmd)
		}
		if len(config.SecureFilterOut) > 0 {
			filterCmd := parsers.BuildIPv6SecureFilterOutCommand(config.Interface, config.SecureFilterOut, config.DynamicFilterOut)
			log.Printf("[DEBUG] Setting IPv6 outbound filter with command: %s", filterCmd)
			output, err := s.executor.Run(ctx, filterCmd)
			if err != nil {
				return fmt.Errorf("failed to set IPv6 outbound filter: %w", err)
			}
			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("IPv6 outbound filter command failed: %s", string(output))
			}
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("IPv6 interface updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// Reset removes IPv6 interface configuration (resets to defaults)
func (s *IPv6InterfaceService) Reset(ctx context.Context, interfaceName string) error {
	// Validate interface name
	if err := parsers.ValidateIPv6InterfaceName(interfaceName); err != nil {
		return err
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Execute all delete commands
	commands := parsers.BuildDeleteIPv6InterfaceCommands(interfaceName)
	for _, cmd := range commands {
		log.Printf("[DEBUG] Resetting IPv6 interface with command: %s", cmd)
		_, _ = s.executor.Run(ctx, cmd)
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("IPv6 interface reset but failed to save configuration: %w", err)
		}
	}

	return nil
}

// List retrieves all IPv6 interface configurations
func (s *IPv6InterfaceService) List(ctx context.Context) ([]IPv6InterfaceConfig, error) {
	// Get configuration for common interface names
	interfaces := []string{"lan1", "lan2", "lan3", "bridge1", "pp1", "tunnel1"}
	var configs []IPv6InterfaceConfig

	for _, iface := range interfaces {
		config, err := s.Get(ctx, iface)
		if err != nil {
			// Skip interfaces that don't have configuration
			continue
		}
		// Only include interfaces with actual IPv6 configuration
		if len(config.Addresses) > 0 || config.RTADV != nil ||
			config.DHCPv6Service != "" || config.MTU > 0 ||
			len(config.SecureFilterIn) > 0 || len(config.SecureFilterOut) > 0 {
			configs = append(configs, *config)
		}
	}

	return configs, nil
}

// toParserConfig converts client.IPv6InterfaceConfig to parsers.IPv6InterfaceConfig
func (s *IPv6InterfaceService) toParserConfig(config IPv6InterfaceConfig) parsers.IPv6InterfaceConfig {
	parserConfig := parsers.IPv6InterfaceConfig{
		Interface:        config.Interface,
		DHCPv6Service:    config.DHCPv6Service,
		MTU:              config.MTU,
		SecureFilterIn:   config.SecureFilterIn,
		SecureFilterOut:  config.SecureFilterOut,
		DynamicFilterOut: config.DynamicFilterOut,
	}

	// Convert addresses
	for _, addr := range config.Addresses {
		parserConfig.Addresses = append(parserConfig.Addresses, parsers.IPv6Address{
			Address:     addr.Address,
			PrefixRef:   addr.PrefixRef,
			InterfaceID: addr.InterfaceID,
		})
	}

	// Convert RTADV
	if config.RTADV != nil {
		parserConfig.RTADV = &parsers.RTADVConfig{
			Enabled:  config.RTADV.Enabled,
			PrefixID: config.RTADV.PrefixID,
			OFlag:    config.RTADV.OFlag,
			MFlag:    config.RTADV.MFlag,
			Lifetime: config.RTADV.Lifetime,
		}
	}

	return parserConfig
}

// fromParserConfig converts parsers.IPv6InterfaceConfig to client.IPv6InterfaceConfig
func (s *IPv6InterfaceService) fromParserConfig(pc parsers.IPv6InterfaceConfig) IPv6InterfaceConfig {
	config := IPv6InterfaceConfig{
		Interface:        pc.Interface,
		DHCPv6Service:    pc.DHCPv6Service,
		MTU:              pc.MTU,
		SecureFilterIn:   pc.SecureFilterIn,
		SecureFilterOut:  pc.SecureFilterOut,
		DynamicFilterOut: pc.DynamicFilterOut,
	}

	// Convert addresses
	for _, addr := range pc.Addresses {
		config.Addresses = append(config.Addresses, IPv6Address{
			Address:     addr.Address,
			PrefixRef:   addr.PrefixRef,
			InterfaceID: addr.InterfaceID,
		})
	}

	// Convert RTADV
	if pc.RTADV != nil {
		config.RTADV = &RTADVConfig{
			Enabled:  pc.RTADV.Enabled,
			PrefixID: pc.RTADV.PrefixID,
			OFlag:    pc.RTADV.OFlag,
			MFlag:    pc.RTADV.MFlag,
			Lifetime: pc.RTADV.Lifetime,
		}
	}

	return config
}

// ipv6AddressesEqual compares two slices of IPv6Address for equality
func ipv6AddressesEqual(a, b []IPv6Address) bool {
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

// rtadvConfigsEqual compares two RTADVConfig pointers for equality
func rtadvConfigsEqual(a, b *RTADVConfig) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Enabled == b.Enabled &&
		a.PrefixID == b.PrefixID &&
		a.OFlag == b.OFlag &&
		a.MFlag == b.MFlag &&
		a.Lifetime == b.Lifetime
}
