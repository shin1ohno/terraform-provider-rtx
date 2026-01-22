package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// DHCPScopeService handles DHCP scope operations
type DHCPScopeService struct {
	executor Executor
	client   *rtxClient // Reference to the main client for save functionality
}

// NewDHCPScopeService creates a new DHCP scope service instance
func NewDHCPScopeService(executor Executor, client *rtxClient) *DHCPScopeService {
	return &DHCPScopeService{
		executor: executor,
		client:   client,
	}
}

// CreateScope creates a new DHCP scope
func (s *DHCPScopeService) CreateScope(ctx context.Context, scope DHCPScope) error {
	// Convert client.DHCPScope to parsers.DHCPScope
	parserScope := s.toParserScope(scope)

	// Validate input
	if err := parsers.ValidateDHCPScope(parserScope); err != nil {
		return fmt.Errorf("invalid scope: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Collect all commands
	commands := []string{}

	// Build scope creation command
	cmd := parsers.BuildDHCPScopeCommand(parserScope)
	logging.FromContext(ctx).Debug().Str("service", "dhcp_scope").Msgf("Creating DHCP scope with command: %s", cmd)
	commands = append(commands, cmd)

	// Configure DHCP options (DNS, routers, domain) if any are specified
	if len(scope.Options.DNSServers) > 0 || len(scope.Options.Routers) > 0 || scope.Options.DomainName != "" {
		optsCmd := parsers.BuildDHCPScopeOptionsCommand(scope.ScopeID, parserScope.Options)
		logging.FromContext(ctx).Debug().Str("service", "dhcp_scope").Msgf("Setting DHCP options with command: %s", optsCmd)
		commands = append(commands, optsCmd)
	}

	// Configure exclusion ranges
	for _, excludeRange := range scope.ExcludeRanges {
		parserRange := parsers.ExcludeRange{
			Start: excludeRange.Start,
			End:   excludeRange.End,
		}
		exceptCmd := parsers.BuildDHCPScopeExceptCommand(scope.ScopeID, parserRange)
		logging.FromContext(ctx).Debug().Str("service", "dhcp_scope").Msgf("Adding exclusion range with command: %s", exceptCmd)
		commands = append(commands, exceptCmd)
	}

	// Execute all commands in batch
	output, err := s.executor.RunBatch(ctx, commands)
	if err != nil {
		return fmt.Errorf("failed to create DHCP scope: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("scope created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetScope retrieves a DHCP scope configuration
func (s *DHCPScopeService) GetScope(ctx context.Context, scopeID int) (*DHCPScope, error) {
	cmd := parsers.BuildShowDHCPScopeCommand(scopeID)
	logging.FromContext(ctx).Debug().Str("service", "dhcp_scope").Msgf("Getting DHCP scope with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get DHCP scope: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "dhcp_scope").Msgf("DHCP scope raw output: %q", string(output))

	parser := parsers.NewDHCPScopeParser()
	parserScope, err := parser.ParseSingleScope(string(output), scopeID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DHCP scope: %w", err)
	}

	// Convert parsers.DHCPScope to client.DHCPScope
	scope := s.fromParserScope(*parserScope)
	return &scope, nil
}

// UpdateScope updates an existing DHCP scope
// Note: network and scope_id changes require recreation
func (s *DHCPScopeService) UpdateScope(ctx context.Context, scope DHCPScope) error {
	parserScope := s.toParserScope(scope)

	// Validate input
	if err := parsers.ValidateDHCPScope(parserScope); err != nil {
		return fmt.Errorf("invalid scope: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get current scope configuration
	currentScope, err := s.GetScope(ctx, scope.ScopeID)
	if err != nil {
		return fmt.Errorf("failed to get current scope: %w", err)
	}

	// If network changed, we need to recreate the scope
	if currentScope.Network != scope.Network {
		return fmt.Errorf("network cannot be changed without recreating the scope")
	}

	// Collect all commands
	commands := []string{}

	// Update gateway and lease time by recreating the base scope command
	// RTX routers allow re-running the scope command to update these values
	cmd := parsers.BuildDHCPScopeCommand(parserScope)
	logging.FromContext(ctx).Debug().Str("service", "dhcp_scope").Msgf("Updating DHCP scope with command: %s", cmd)
	commands = append(commands, cmd)

	// Update DHCP options (DNS, routers, domain)
	// First, remove existing options configuration
	hasCurrentOptions := len(currentScope.Options.DNSServers) > 0 || len(currentScope.Options.Routers) > 0 || currentScope.Options.DomainName != ""
	if hasCurrentOptions {
		deleteCmd := parsers.BuildDeleteDHCPScopeOptionsCommand(scope.ScopeID)
		logging.FromContext(ctx).Debug().Str("service", "dhcp_scope").Msgf("Removing existing options with command: %s", deleteCmd)
		commands = append(commands, deleteCmd)
	}

	// Set new options if specified
	hasNewOptions := len(scope.Options.DNSServers) > 0 || len(scope.Options.Routers) > 0 || scope.Options.DomainName != ""
	if hasNewOptions {
		optsCmd := parsers.BuildDHCPScopeOptionsCommand(scope.ScopeID, parserScope.Options)
		logging.FromContext(ctx).Debug().Str("service", "dhcp_scope").Msgf("Setting DHCP options with command: %s", optsCmd)
		commands = append(commands, optsCmd)
	}

	// Update exclusion ranges
	// Remove old ranges that are not in new configuration
	for _, oldRange := range currentScope.ExcludeRanges {
		found := false
		for _, newRange := range scope.ExcludeRanges {
			if oldRange.Start == newRange.Start && oldRange.End == newRange.End {
				found = true
				break
			}
		}
		if !found {
			parserRange := parsers.ExcludeRange{
				Start: oldRange.Start,
				End:   oldRange.End,
			}
			deleteCmd := parsers.BuildDeleteDHCPScopeExceptCommand(scope.ScopeID, parserRange)
			logging.FromContext(ctx).Debug().Str("service", "dhcp_scope").Msgf("Removing exclusion range with command: %s", deleteCmd)
			commands = append(commands, deleteCmd)
		}
	}

	// Add new ranges that are not in old configuration
	for _, newRange := range scope.ExcludeRanges {
		found := false
		for _, oldRange := range currentScope.ExcludeRanges {
			if oldRange.Start == newRange.Start && oldRange.End == newRange.End {
				found = true
				break
			}
		}
		if !found {
			parserRange := parsers.ExcludeRange{
				Start: newRange.Start,
				End:   newRange.End,
			}
			exceptCmd := parsers.BuildDHCPScopeExceptCommand(scope.ScopeID, parserRange)
			logging.FromContext(ctx).Debug().Str("service", "dhcp_scope").Msgf("Adding exclusion range with command: %s", exceptCmd)
			commands = append(commands, exceptCmd)
		}
	}

	// Execute all commands in batch
	output, err := s.executor.RunBatch(ctx, commands)
	if err != nil {
		return fmt.Errorf("failed to update DHCP scope: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("scope updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// DeleteScope removes a DHCP scope
func (s *DHCPScopeService) DeleteScope(ctx context.Context, scopeID int) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cmd := parsers.BuildDeleteDHCPScopeCommand(scopeID)
	logging.FromContext(ctx).Debug().Str("service", "dhcp_scope").Msgf("Deleting DHCP scope with command: %s", cmd)

	// Execute command in batch
	output, err := s.executor.RunBatch(ctx, []string{cmd})
	if err != nil {
		return fmt.Errorf("failed to delete DHCP scope: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		// Check if it's already gone
		if strings.Contains(strings.ToLower(string(output)), "not found") {
			return nil
		}
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("scope deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ListScopes retrieves all DHCP scopes
func (s *DHCPScopeService) ListScopes(ctx context.Context) ([]DHCPScope, error) {
	cmd := parsers.BuildShowAllDHCPScopesCommand()
	logging.FromContext(ctx).Debug().Str("service", "dhcp_scope").Msgf("Listing DHCP scopes with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list DHCP scopes: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "dhcp_scope").Msgf("DHCP scopes raw output: %q", string(output))

	parser := parsers.NewDHCPScopeParser()
	parserScopes, err := parser.ParseScopeConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse DHCP scopes: %w", err)
	}

	// Convert parsers.DHCPScope to client.DHCPScope
	scopes := make([]DHCPScope, len(parserScopes))
	for i, ps := range parserScopes {
		scopes[i] = s.fromParserScope(ps)
	}

	return scopes, nil
}

// toParserScope converts client.DHCPScope to parsers.DHCPScope
func (s *DHCPScopeService) toParserScope(scope DHCPScope) parsers.DHCPScope {
	excludeRanges := make([]parsers.ExcludeRange, len(scope.ExcludeRanges))
	for i, r := range scope.ExcludeRanges {
		excludeRanges[i] = parsers.ExcludeRange{
			Start: r.Start,
			End:   r.End,
		}
	}

	return parsers.DHCPScope{
		ScopeID:       scope.ScopeID,
		Network:       scope.Network,
		LeaseTime:     scope.LeaseTime,
		ExcludeRanges: excludeRanges,
		Options: parsers.DHCPScopeOptions{
			DNSServers: scope.Options.DNSServers,
			Routers:    scope.Options.Routers,
			DomainName: scope.Options.DomainName,
		},
	}
}

// fromParserScope converts parsers.DHCPScope to client.DHCPScope
func (s *DHCPScopeService) fromParserScope(ps parsers.DHCPScope) DHCPScope {
	excludeRanges := make([]ExcludeRange, len(ps.ExcludeRanges))
	for i, r := range ps.ExcludeRanges {
		excludeRanges[i] = ExcludeRange{
			Start: r.Start,
			End:   r.End,
		}
	}

	return DHCPScope{
		ScopeID:       ps.ScopeID,
		Network:       ps.Network,
		LeaseTime:     ps.LeaseTime,
		ExcludeRanges: excludeRanges,
		Options: DHCPScopeOptions{
			DNSServers: ps.Options.DNSServers,
			Routers:    ps.Options.Routers,
			DomainName: ps.Options.DomainName,
		},
	}
}
