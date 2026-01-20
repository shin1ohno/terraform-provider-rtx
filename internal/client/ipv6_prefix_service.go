package client

import (
	"context"
	"fmt"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// IPv6PrefixService handles IPv6 prefix operations
type IPv6PrefixService struct {
	executor Executor
	client   *rtxClient // Reference to the main client for save functionality
}

// NewIPv6PrefixService creates a new IPv6 prefix service instance
func NewIPv6PrefixService(executor Executor, client *rtxClient) *IPv6PrefixService {
	return &IPv6PrefixService{
		executor: executor,
		client:   client,
	}
}

// CreatePrefix creates a new IPv6 prefix
func (s *IPv6PrefixService) CreatePrefix(ctx context.Context, prefix IPv6Prefix) error {
	// Convert client.IPv6Prefix to parsers.IPv6Prefix
	parserPrefix := s.toParserPrefix(prefix)

	// Validate input
	if err := parsers.ValidateIPv6Prefix(parserPrefix); err != nil {
		return fmt.Errorf("invalid prefix: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Build and execute prefix creation command
	cmd := parsers.BuildIPv6PrefixCommand(parserPrefix)
	logging.FromContext(ctx).Debug().Str("service", "ipv6_prefix").Msgf("Creating IPv6 prefix with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to create IPv6 prefix: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("prefix created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetPrefix retrieves an IPv6 prefix configuration
func (s *IPv6PrefixService) GetPrefix(ctx context.Context, prefixID int) (*IPv6Prefix, error) {
	cmd := parsers.BuildShowIPv6PrefixCommand(prefixID)
	logging.FromContext(ctx).Debug().Str("service", "ipv6_prefix").Msgf("Getting IPv6 prefix with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPv6 prefix: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "ipv6_prefix").Msgf("IPv6 prefix raw output: %q", string(output))

	parser := parsers.NewIPv6PrefixParser()
	parserPrefix, err := parser.ParseSinglePrefix(string(output), prefixID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse IPv6 prefix: %w", err)
	}

	// Convert parsers.IPv6Prefix to client.IPv6Prefix
	prefix := s.fromParserPrefix(*parserPrefix)
	return &prefix, nil
}

// UpdatePrefix updates an existing IPv6 prefix
// Note: source changes require ForceNew (recreation), only prefix_length and prefix value can be updated
func (s *IPv6PrefixService) UpdatePrefix(ctx context.Context, prefix IPv6Prefix) error {
	parserPrefix := s.toParserPrefix(prefix)

	// Validate input
	if err := parsers.ValidateIPv6Prefix(parserPrefix); err != nil {
		return fmt.Errorf("invalid prefix: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get current prefix configuration to check for changes
	currentPrefix, err := s.GetPrefix(ctx, prefix.ID)
	if err != nil {
		return fmt.Errorf("failed to get current prefix: %w", err)
	}

	// Check if source changed (requires ForceNew)
	if currentPrefix.Source != prefix.Source {
		return fmt.Errorf("source cannot be changed without recreating the prefix")
	}

	// For ra/dhcpv6-pd, check if interface changed (requires ForceNew)
	if (prefix.Source == "ra" || prefix.Source == "dhcpv6-pd") && currentPrefix.Interface != prefix.Interface {
		return fmt.Errorf("interface cannot be changed without recreating the prefix")
	}

	// Update by re-issuing the prefix command
	// RTX routers allow re-running the prefix command to update values
	cmd := parsers.BuildIPv6PrefixCommand(parserPrefix)
	logging.FromContext(ctx).Debug().Str("service", "ipv6_prefix").Msgf("Updating IPv6 prefix with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to update IPv6 prefix: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("prefix updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// DeletePrefix removes an IPv6 prefix
func (s *IPv6PrefixService) DeletePrefix(ctx context.Context, prefixID int) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cmd := parsers.BuildDeleteIPv6PrefixCommand(prefixID)
	logging.FromContext(ctx).Debug().Str("service", "ipv6_prefix").Msgf("Deleting IPv6 prefix with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete IPv6 prefix: %w", err)
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
			return fmt.Errorf("prefix deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ListPrefixes retrieves all IPv6 prefixes
func (s *IPv6PrefixService) ListPrefixes(ctx context.Context) ([]IPv6Prefix, error) {
	cmd := parsers.BuildShowAllIPv6PrefixesCommand()
	logging.FromContext(ctx).Debug().Str("service", "ipv6_prefix").Msgf("Listing IPv6 prefixes with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list IPv6 prefixes: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "ipv6_prefix").Msgf("IPv6 prefixes raw output: %q", string(output))

	parser := parsers.NewIPv6PrefixParser()
	parserPrefixes, err := parser.ParseIPv6PrefixConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IPv6 prefixes: %w", err)
	}

	// Convert parsers.IPv6Prefix to client.IPv6Prefix
	prefixes := make([]IPv6Prefix, len(parserPrefixes))
	for i, pp := range parserPrefixes {
		prefixes[i] = s.fromParserPrefix(pp)
	}

	return prefixes, nil
}

// toParserPrefix converts client.IPv6Prefix to parsers.IPv6Prefix
func (s *IPv6PrefixService) toParserPrefix(prefix IPv6Prefix) parsers.IPv6Prefix {
	return parsers.IPv6Prefix{
		ID:           prefix.ID,
		Prefix:       prefix.Prefix,
		PrefixLength: prefix.PrefixLength,
		Source:       prefix.Source,
		Interface:    prefix.Interface,
	}
}

// fromParserPrefix converts parsers.IPv6Prefix to client.IPv6Prefix
func (s *IPv6PrefixService) fromParserPrefix(pp parsers.IPv6Prefix) IPv6Prefix {
	return IPv6Prefix{
		ID:           pp.ID,
		Prefix:       pp.Prefix,
		PrefixLength: pp.PrefixLength,
		Source:       pp.Source,
		Interface:    pp.Interface,
	}
}
