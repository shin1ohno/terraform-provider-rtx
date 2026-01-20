package client

import (
	"context"
	"fmt"
	"log"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// DNSService handles DNS operations
type DNSService struct {
	executor Executor
	client   *rtxClient // Reference to the main client for save functionality
}

// NewDNSService creates a new DNS service instance
func NewDNSService(executor Executor, client *rtxClient) *DNSService {
	return &DNSService{
		executor: executor,
		client:   client,
	}
}

// Get retrieves the DNS configuration
func (s *DNSService) Get(ctx context.Context) (*DNSConfig, error) {
	cmd := parsers.BuildShowDNSConfigCommand()
	log.Printf("[DEBUG] Getting DNS config with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS config: %w", err)
	}

	log.Printf("[DEBUG] DNS raw output: %q", string(output))

	parser := parsers.NewDNSParser()
	parserConfig, err := parser.ParseDNSConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse DNS config: %w", err)
	}

	// Convert parsers.DNSConfig to client.DNSConfig
	config := s.fromParserConfig(parserConfig)
	return &config, nil
}

// Configure creates DNS configuration
func (s *DNSService) Configure(ctx context.Context, config DNSConfig) error {
	// Convert client.DNSConfig to parsers.DNSConfig for validation
	parserConfig := s.toParserConfig(config)

	// Validate input
	if err := parsers.ValidateDNSConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid DNS config: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Configure domain lookup
	if !config.DomainLookup {
		cmd := parsers.BuildDNSDomainLookupCommand(false)
		log.Printf("[DEBUG] Setting domain lookup with command: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to set domain lookup: %w", err)
		}
	}

	// Configure domain name
	if config.DomainName != "" {
		cmd := parsers.BuildDNSDomainNameCommand(config.DomainName)
		log.Printf("[DEBUG] Setting domain name with command: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to set domain name: %w", err)
		}
	}

	// Configure name servers
	if len(config.NameServers) > 0 {
		cmd := parsers.BuildDNSServerCommand(config.NameServers)
		log.Printf("[DEBUG] Setting DNS servers with command: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to set DNS servers: %w", err)
		}
	}

	// Configure server select entries
	for _, sel := range config.ServerSelect {
		parserSel := parsers.DNSServerSelect{
			ID:             sel.ID,
			Servers:        sel.Servers,
			EDNS:           sel.EDNS,
			RecordType:     sel.RecordType,
			QueryPattern:   sel.QueryPattern,
			OriginalSender: sel.OriginalSender,
			RestrictPP:     sel.RestrictPP,
		}
		cmd := parsers.BuildDNSServerSelectCommand(parserSel)
		if cmd == "" {
			continue
		}
		log.Printf("[DEBUG] Setting DNS server select with command: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to set DNS server select %d: %w", sel.ID, err)
		}
	}

	// Configure static hosts
	for _, host := range config.Hosts {
		parserHost := parsers.DNSHost{
			Name:    host.Name,
			Address: host.Address,
		}
		cmd := parsers.BuildDNSStaticCommand(parserHost)
		if cmd == "" {
			continue
		}
		log.Printf("[DEBUG] Setting DNS static host with command: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to set DNS static host %s: %w", host.Name, err)
		}
	}

	// Configure DNS service
	cmd := parsers.BuildDNSServiceCommand(config.ServiceOn)
	log.Printf("[DEBUG] Setting DNS service with command: %s", cmd)
	if _, err := s.executor.Run(ctx, cmd); err != nil {
		return fmt.Errorf("failed to set DNS service: %w", err)
	}

	// Configure private address spoof
	cmd = parsers.BuildDNSPrivateSpoofCommand(config.PrivateSpoof)
	log.Printf("[DEBUG] Setting DNS private spoof with command: %s", cmd)
	if _, err := s.executor.Run(ctx, cmd); err != nil {
		return fmt.Errorf("failed to set DNS private spoof: %w", err)
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("DNS config created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// Update updates DNS configuration
func (s *DNSService) Update(ctx context.Context, config DNSConfig) error {
	// Convert client.DNSConfig to parsers.DNSConfig for validation
	parserConfig := s.toParserConfig(config)

	// Validate input
	if err := parsers.ValidateDNSConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid DNS config: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get current configuration
	currentConfig, err := s.Get(ctx)
	if err != nil {
		log.Printf("[DEBUG] Could not get current DNS config, proceeding with update: %v", err)
		currentConfig = &DNSConfig{}
	}

	// Update domain lookup
	if config.DomainLookup != currentConfig.DomainLookup {
		cmd := parsers.BuildDNSDomainLookupCommand(config.DomainLookup)
		log.Printf("[DEBUG] Updating domain lookup with command: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to update domain lookup: %w", err)
		}
	}

	// Update domain name
	if config.DomainName != currentConfig.DomainName {
		if currentConfig.DomainName != "" {
			cmd := parsers.BuildDeleteDNSDomainNameCommand()
			log.Printf("[DEBUG] Removing old domain name with command: %s", cmd)
			_, _ = s.executor.Run(ctx, cmd) // Ignore errors for cleanup
		}
		if config.DomainName != "" {
			cmd := parsers.BuildDNSDomainNameCommand(config.DomainName)
			log.Printf("[DEBUG] Setting domain name with command: %s", cmd)
			if _, err := s.executor.Run(ctx, cmd); err != nil {
				return fmt.Errorf("failed to set domain name: %w", err)
			}
		}
	}

	// Update name servers
	if !slicesEqual(config.NameServers, currentConfig.NameServers) {
		// Remove old servers
		cmd := parsers.BuildDeleteDNSServerCommand()
		log.Printf("[DEBUG] Removing old DNS servers with command: %s", cmd)
		_, _ = s.executor.Run(ctx, cmd) // Ignore errors for cleanup

		// Set new servers
		if len(config.NameServers) > 0 {
			cmd = parsers.BuildDNSServerCommand(config.NameServers)
			log.Printf("[DEBUG] Setting DNS servers with command: %s", cmd)
			if _, err := s.executor.Run(ctx, cmd); err != nil {
				return fmt.Errorf("failed to set DNS servers: %w", err)
			}
		}
	}

	// Update server select entries
	// First, remove entries that are no longer needed
	for _, currentSel := range currentConfig.ServerSelect {
		found := false
		for _, newSel := range config.ServerSelect {
			if newSel.ID == currentSel.ID {
				found = true
				break
			}
		}
		if !found {
			cmd := parsers.BuildDeleteDNSServerSelectCommand(currentSel.ID)
			log.Printf("[DEBUG] Removing DNS server select %d with command: %s", currentSel.ID, cmd)
			_, _ = s.executor.Run(ctx, cmd)
		}
	}
	// Add/update new entries
	for _, sel := range config.ServerSelect {
		parserSel := parsers.DNSServerSelect{
			ID:             sel.ID,
			Servers:        sel.Servers,
			EDNS:           sel.EDNS,
			RecordType:     sel.RecordType,
			QueryPattern:   sel.QueryPattern,
			OriginalSender: sel.OriginalSender,
			RestrictPP:     sel.RestrictPP,
		}
		cmd := parsers.BuildDNSServerSelectCommand(parserSel)
		if cmd == "" {
			continue
		}
		log.Printf("[DEBUG] Setting DNS server select with command: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to set DNS server select %d: %w", sel.ID, err)
		}
	}

	// Update static hosts
	// First, remove hosts that are no longer needed
	for _, currentHost := range currentConfig.Hosts {
		found := false
		for _, newHost := range config.Hosts {
			if newHost.Name == currentHost.Name {
				found = true
				break
			}
		}
		if !found {
			cmd := parsers.BuildDeleteDNSStaticCommand(currentHost.Name)
			log.Printf("[DEBUG] Removing DNS static host %s with command: %s", currentHost.Name, cmd)
			_, _ = s.executor.Run(ctx, cmd)
		}
	}
	// Add/update new entries
	for _, host := range config.Hosts {
		parserHost := parsers.DNSHost{
			Name:    host.Name,
			Address: host.Address,
		}
		cmd := parsers.BuildDNSStaticCommand(parserHost)
		if cmd == "" {
			continue
		}
		log.Printf("[DEBUG] Setting DNS static host with command: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to set DNS static host %s: %w", host.Name, err)
		}
	}

	// Update DNS service
	if config.ServiceOn != currentConfig.ServiceOn {
		cmd := parsers.BuildDNSServiceCommand(config.ServiceOn)
		log.Printf("[DEBUG] Updating DNS service with command: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to update DNS service: %w", err)
		}
	}

	// Update private address spoof
	if config.PrivateSpoof != currentConfig.PrivateSpoof {
		cmd := parsers.BuildDNSPrivateSpoofCommand(config.PrivateSpoof)
		log.Printf("[DEBUG] Updating DNS private spoof with command: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to update DNS private spoof: %w", err)
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("DNS config updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// Reset removes DNS configuration
func (s *DNSService) Reset(ctx context.Context) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get current configuration to clean up server select and static hosts
	currentConfig, err := s.Get(ctx)
	if err != nil {
		log.Printf("[DEBUG] Could not get current DNS config, proceeding with reset: %v", err)
		currentConfig = &DNSConfig{}
	}

	// Remove server select entries
	for _, sel := range currentConfig.ServerSelect {
		cmd := parsers.BuildDeleteDNSServerSelectCommand(sel.ID)
		log.Printf("[DEBUG] Removing DNS server select %d with command: %s", sel.ID, cmd)
		_, _ = s.executor.Run(ctx, cmd)
	}

	// Remove static hosts
	for _, host := range currentConfig.Hosts {
		cmd := parsers.BuildDeleteDNSStaticCommand(host.Name)
		log.Printf("[DEBUG] Removing DNS static host %s with command: %s", host.Name, cmd)
		_, _ = s.executor.Run(ctx, cmd)
	}

	// Execute delete commands
	deleteCommands := parsers.BuildDeleteDNSCommand()
	for _, cmd := range deleteCommands {
		log.Printf("[DEBUG] Resetting DNS with command: %s", cmd)
		if _, err := s.executor.Run(ctx, cmd); err != nil {
			log.Printf("[DEBUG] Command %s failed: %v (continuing)", cmd, err)
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("DNS config reset but failed to save configuration: %w", err)
		}
	}

	return nil
}

// toParserConfig converts client.DNSConfig to parsers.DNSConfig
func (s *DNSService) toParserConfig(config DNSConfig) parsers.DNSConfig {
	serverSelect := make([]parsers.DNSServerSelect, len(config.ServerSelect))
	for i, sel := range config.ServerSelect {
		serverSelect[i] = parsers.DNSServerSelect{
			ID:             sel.ID,
			Servers:        sel.Servers,
			EDNS:           sel.EDNS,
			RecordType:     sel.RecordType,
			QueryPattern:   sel.QueryPattern,
			OriginalSender: sel.OriginalSender,
			RestrictPP:     sel.RestrictPP,
		}
	}

	hosts := make([]parsers.DNSHost, len(config.Hosts))
	for i, host := range config.Hosts {
		hosts[i] = parsers.DNSHost{
			Name:    host.Name,
			Address: host.Address,
		}
	}

	return parsers.DNSConfig{
		DomainLookup: config.DomainLookup,
		DomainName:   config.DomainName,
		NameServers:  config.NameServers,
		ServerSelect: serverSelect,
		Hosts:        hosts,
		ServiceOn:    config.ServiceOn,
		PrivateSpoof: config.PrivateSpoof,
	}
}

// fromParserConfig converts parsers.DNSConfig to client.DNSConfig
func (s *DNSService) fromParserConfig(parserConfig *parsers.DNSConfig) DNSConfig {
	serverSelect := make([]DNSServerSelect, len(parserConfig.ServerSelect))
	for i, sel := range parserConfig.ServerSelect {
		serverSelect[i] = DNSServerSelect{
			ID:             sel.ID,
			Servers:        sel.Servers,
			EDNS:           sel.EDNS,
			RecordType:     sel.RecordType,
			QueryPattern:   sel.QueryPattern,
			OriginalSender: sel.OriginalSender,
			RestrictPP:     sel.RestrictPP,
		}
	}

	hosts := make([]DNSHost, len(parserConfig.Hosts))
	for i, host := range parserConfig.Hosts {
		hosts[i] = DNSHost{
			Name:    host.Name,
			Address: host.Address,
		}
	}

	return DNSConfig{
		DomainLookup: parserConfig.DomainLookup,
		DomainName:   parserConfig.DomainName,
		NameServers:  parserConfig.NameServers,
		ServerSelect: serverSelect,
		Hosts:        hosts,
		ServiceOn:    parserConfig.ServiceOn,
		PrivateSpoof: parserConfig.PrivateSpoof,
	}
}

// slicesEqual compares two string slices for equality
func slicesEqual(a, b []string) bool {
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
