package client

import (
	"context"
	"fmt"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// SNMPService handles SNMP configuration operations
type SNMPService struct {
	executor Executor
	client   *rtxClient
}

// NewSNMPService creates a new SNMP service
func NewSNMPService(executor Executor, client *rtxClient) *SNMPService {
	return &SNMPService{
		executor: executor,
		client:   client,
	}
}

// Get retrieves the current SNMP configuration
func (s *SNMPService) Get(ctx context.Context) (*SNMPConfig, error) {
	output, err := s.executor.Run(ctx, parsers.BuildShowSNMPConfigCommand())
	if err != nil {
		return nil, fmt.Errorf("failed to get SNMP config: %w", err)
	}

	parser := parsers.NewSNMPParser()
	parsed, err := parser.ParseSNMPConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse SNMP config: %w", err)
	}

	// Convert from parser type to client type
	config := &SNMPConfig{
		SysName:     parsed.SysName,
		SysLocation: parsed.SysLocation,
		SysContact:  parsed.SysContact,
		Communities: make([]SNMPCommunity, len(parsed.Communities)),
		Hosts:       make([]SNMPHost, len(parsed.Hosts)),
		TrapEnable:  make([]string, len(parsed.TrapEnable)),
	}

	for i, c := range parsed.Communities {
		config.Communities[i] = SNMPCommunity{
			Name:       c.Name,
			Permission: c.Permission,
			ACL:        c.ACL,
		}
	}

	for i, h := range parsed.Hosts {
		config.Hosts[i] = SNMPHost{
			Address:   h.Address,
			Community: h.Community,
			Version:   h.Version,
		}
	}

	copy(config.TrapEnable, parsed.TrapEnable)

	return config, nil
}

// Create creates a new SNMP configuration
func (s *SNMPService) Create(ctx context.Context, config SNMPConfig) error {
	// Validate configuration
	parserConfig := s.toParserConfig(config)
	if err := parsers.ValidateSNMPConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid SNMP config: %w", err)
	}

	// Build and execute commands in order
	commands := []string{}

	// 1. Set system information
	if config.SysName != "" {
		commands = append(commands, parsers.BuildSNMPSysNameCommand(config.SysName))
	}
	if config.SysLocation != "" {
		commands = append(commands, parsers.BuildSNMPSysLocationCommand(config.SysLocation))
	}
	if config.SysContact != "" {
		commands = append(commands, parsers.BuildSNMPSysContactCommand(config.SysContact))
	}

	// 2. Configure communities
	for _, community := range config.Communities {
		parserCommunity := parsers.SNMPCommunity{
			Name:       community.Name,
			Permission: community.Permission,
			ACL:        community.ACL,
		}
		commands = append(commands, parsers.BuildSNMPCommunityCommand(parserCommunity))
	}

	// 3. Configure trap community (use first community if available)
	var trapCommunity string
	for _, community := range config.Communities {
		if community.Permission == "ro" {
			trapCommunity = community.Name
			break
		}
	}
	if trapCommunity == "" && len(config.Communities) > 0 {
		trapCommunity = config.Communities[0].Name
	}
	if trapCommunity != "" {
		commands = append(commands, parsers.BuildSNMPTrapCommunityCommand(trapCommunity))
	}

	// 4. Configure trap hosts
	for _, host := range config.Hosts {
		parserHost := parsers.SNMPHost{
			Address:   host.Address,
			Community: host.Community,
			Version:   host.Version,
		}
		commands = append(commands, parsers.BuildSNMPHostCommand(parserHost))
	}

	// 5. Configure trap enable
	if len(config.TrapEnable) > 0 {
		commands = append(commands, parsers.BuildSNMPTrapEnableCommand(config.TrapEnable))
	}

	// Execute all commands
	for _, cmd := range commands {
		logging.FromContext(ctx).Debug().Str("service", "snmp").Msgf("Executing SNMP command: %s", cmd)
		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to execute SNMP command '%s': %w", cmd, err)
		}
		if containsError(string(output)) {
			return fmt.Errorf("SNMP command '%s' failed: %s", cmd, output)
		}
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save SNMP config: %w", err)
	}

	return nil
}

// Update modifies the existing SNMP configuration
func (s *SNMPService) Update(ctx context.Context, config SNMPConfig) error {
	// Get current config to determine what needs to change
	current, err := s.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current SNMP config: %w", err)
	}

	// Validate new configuration
	parserConfig := s.toParserConfig(config)
	if err := parsers.ValidateSNMPConfig(parserConfig); err != nil {
		return fmt.Errorf("invalid SNMP config: %w", err)
	}

	commands := []string{}

	// Update system information
	if config.SysName != current.SysName {
		if config.SysName == "" {
			commands = append(commands, parsers.BuildDeleteSNMPSysNameCommand())
		} else {
			commands = append(commands, parsers.BuildSNMPSysNameCommand(config.SysName))
		}
	}
	if config.SysLocation != current.SysLocation {
		if config.SysLocation == "" {
			commands = append(commands, parsers.BuildDeleteSNMPSysLocationCommand())
		} else {
			commands = append(commands, parsers.BuildSNMPSysLocationCommand(config.SysLocation))
		}
	}
	if config.SysContact != current.SysContact {
		if config.SysContact == "" {
			commands = append(commands, parsers.BuildDeleteSNMPSysContactCommand())
		} else {
			commands = append(commands, parsers.BuildSNMPSysContactCommand(config.SysContact))
		}
	}

	// Remove old communities that are not in new config
	for _, oldCommunity := range current.Communities {
		found := false
		for _, newCommunity := range config.Communities {
			if oldCommunity.Name == newCommunity.Name && oldCommunity.Permission == newCommunity.Permission {
				found = true
				break
			}
		}
		if !found {
			parserCommunity := parsers.SNMPCommunity{
				Name:       oldCommunity.Name,
				Permission: oldCommunity.Permission,
			}
			commands = append(commands, parsers.BuildDeleteSNMPCommunityCommand(parserCommunity))
		}
	}

	// Add new communities
	for _, newCommunity := range config.Communities {
		found := false
		for _, oldCommunity := range current.Communities {
			if oldCommunity.Name == newCommunity.Name && oldCommunity.Permission == newCommunity.Permission && oldCommunity.ACL == newCommunity.ACL {
				found = true
				break
			}
		}
		if !found {
			parserCommunity := parsers.SNMPCommunity{
				Name:       newCommunity.Name,
				Permission: newCommunity.Permission,
				ACL:        newCommunity.ACL,
			}
			commands = append(commands, parsers.BuildSNMPCommunityCommand(parserCommunity))
		}
	}

	// Update trap community
	var newTrapCommunity string
	for _, community := range config.Communities {
		if community.Permission == "ro" {
			newTrapCommunity = community.Name
			break
		}
	}
	if newTrapCommunity == "" && len(config.Communities) > 0 {
		newTrapCommunity = config.Communities[0].Name
	}

	var currentTrapCommunity string
	for _, community := range current.Communities {
		if community.Permission == "ro" {
			currentTrapCommunity = community.Name
			break
		}
	}
	if currentTrapCommunity == "" && len(current.Communities) > 0 {
		currentTrapCommunity = current.Communities[0].Name
	}

	if newTrapCommunity != currentTrapCommunity {
		if newTrapCommunity == "" {
			commands = append(commands, parsers.BuildDeleteSNMPTrapCommunityCommand())
		} else {
			commands = append(commands, parsers.BuildSNMPTrapCommunityCommand(newTrapCommunity))
		}
	}

	// Remove old hosts that are not in new config
	for _, oldHost := range current.Hosts {
		found := false
		for _, newHost := range config.Hosts {
			if oldHost.Address == newHost.Address {
				found = true
				break
			}
		}
		if !found {
			commands = append(commands, parsers.BuildDeleteSNMPHostCommand(oldHost.Address))
		}
	}

	// Add new hosts
	for _, newHost := range config.Hosts {
		found := false
		for _, oldHost := range current.Hosts {
			if oldHost.Address == newHost.Address {
				found = true
				break
			}
		}
		if !found {
			parserHost := parsers.SNMPHost{
				Address:   newHost.Address,
				Community: newHost.Community,
				Version:   newHost.Version,
			}
			commands = append(commands, parsers.BuildSNMPHostCommand(parserHost))
		}
	}

	// Update trap enable settings
	if !stringSlicesEqual(config.TrapEnable, current.TrapEnable) {
		if len(current.TrapEnable) > 0 {
			commands = append(commands, parsers.BuildDeleteSNMPTrapEnableCommand())
		}
		if len(config.TrapEnable) > 0 {
			commands = append(commands, parsers.BuildSNMPTrapEnableCommand(config.TrapEnable))
		}
	}

	// Execute all commands
	for _, cmd := range commands {
		logging.FromContext(ctx).Debug().Str("service", "snmp").Msgf("Executing SNMP command: %s", cmd)
		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to execute SNMP command '%s': %w", cmd, err)
		}
		if containsError(string(output)) {
			return fmt.Errorf("SNMP command '%s' failed: %s", cmd, output)
		}
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save SNMP config: %w", err)
	}

	return nil
}

// Delete removes SNMP configuration
func (s *SNMPService) Delete(ctx context.Context) error {
	// Get current configuration to know what to delete
	current, err := s.Get(ctx)
	if err != nil {
		// If we can't read config, assume it's already clean
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "empty") {
			return nil
		}
		return fmt.Errorf("failed to get current SNMP config: %w", err)
	}

	commands := []string{}

	// Remove trap settings first
	if len(current.TrapEnable) > 0 {
		commands = append(commands, parsers.BuildDeleteSNMPTrapEnableCommand())
	}

	// Remove hosts
	for _, host := range current.Hosts {
		commands = append(commands, parsers.BuildDeleteSNMPHostCommand(host.Address))
	}

	// Remove trap community
	commands = append(commands, parsers.BuildDeleteSNMPTrapCommunityCommand())

	// Remove communities
	for _, community := range current.Communities {
		parserCommunity := parsers.SNMPCommunity{
			Name:       community.Name,
			Permission: community.Permission,
		}
		commands = append(commands, parsers.BuildDeleteSNMPCommunityCommand(parserCommunity))
	}

	// Remove system information
	if current.SysName != "" {
		commands = append(commands, parsers.BuildDeleteSNMPSysNameCommand())
	}
	if current.SysLocation != "" {
		commands = append(commands, parsers.BuildDeleteSNMPSysLocationCommand())
	}
	if current.SysContact != "" {
		commands = append(commands, parsers.BuildDeleteSNMPSysContactCommand())
	}

	// Execute all commands
	for _, cmd := range commands {
		logging.FromContext(ctx).Debug().Str("service", "snmp").Msgf("Executing SNMP delete command: %s", cmd)
		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			// Ignore errors for delete commands - resource may already be gone
			logging.FromContext(ctx).Debug().Str("service", "snmp").Msgf("SNMP delete command '%s' failed (may already be deleted): %v", cmd, err)
			continue
		}
		if containsError(string(output)) {
			// Ignore errors for delete - may already be deleted
			logging.FromContext(ctx).Debug().Str("service", "snmp").Msgf("SNMP delete command '%s' returned error (may already be deleted): %s", cmd, output)
		}
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save config after SNMP delete: %w", err)
	}

	return nil
}

// toParserConfig converts client SNMPConfig to parser SNMPConfig
func (s *SNMPService) toParserConfig(config SNMPConfig) parsers.SNMPConfig {
	parserConfig := parsers.SNMPConfig{
		SysName:     config.SysName,
		SysLocation: config.SysLocation,
		SysContact:  config.SysContact,
		Communities: make([]parsers.SNMPCommunity, len(config.Communities)),
		Hosts:       make([]parsers.SNMPHost, len(config.Hosts)),
		TrapEnable:  make([]string, len(config.TrapEnable)),
	}

	for i, c := range config.Communities {
		parserConfig.Communities[i] = parsers.SNMPCommunity{
			Name:       c.Name,
			Permission: c.Permission,
			ACL:        c.ACL,
		}
	}

	for i, h := range config.Hosts {
		parserConfig.Hosts[i] = parsers.SNMPHost{
			Address:   h.Address,
			Community: h.Community,
			Version:   h.Version,
		}
	}

	copy(parserConfig.TrapEnable, config.TrapEnable)

	return parserConfig
}

// stringSlicesEqual compares two string slices for equality
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
