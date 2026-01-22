package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// IPFilterService handles IP filter operations
type IPFilterService struct {
	executor Executor
	client   *rtxClient // Reference to the main client for save functionality
}

// NewIPFilterService creates a new IP filter service instance
func NewIPFilterService(executor Executor, client *rtxClient) *IPFilterService {
	return &IPFilterService{
		executor: executor,
		client:   client,
	}
}

// CreateFilter creates a new IP filter
func (s *IPFilterService) CreateFilter(ctx context.Context, filter IPFilter) error {
	logger := logging.FromContext(ctx)

	// Convert client.IPFilter to parsers.IPFilter
	parserFilter := s.toParserFilter(filter)

	// Validate input
	if err := parsers.ValidateIPFilter(parserFilter); err != nil {
		return fmt.Errorf("invalid IP filter: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Build and execute filter creation command
	cmd := parsers.BuildIPFilterCommand(parserFilter)
	logger.Debug().Str("service", "IPFilterService").Str("operation", "CreateFilter").Msgf("Creating IP filter with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to create IP filter: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("filter created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetFilter retrieves an IP filter configuration
func (s *IPFilterService) GetFilter(ctx context.Context, number int) (*IPFilter, error) {
	logger := logging.FromContext(ctx)

	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cmd := parsers.BuildShowIPFilterByNumberCommand(number)
	logger.Debug().Str("service", "IPFilterService").Str("operation", "GetFilter").Msgf("Getting IP filter with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get IP filter: %w", err)
	}

	logger.Debug().Str("service", "IPFilterService").Str("operation", "GetFilter").Msgf("IP filter raw output: %q", string(output))

	// Parse the output
	parserFilters, err := parsers.ParseIPFilterConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IP filter: %w", err)
	}

	// Find the specific filter
	for _, pf := range parserFilters {
		if pf.Number == number {
			filter := s.fromParserFilter(pf)
			return &filter, nil
		}
	}

	return nil, fmt.Errorf("IP filter %d not found", number)
}

// UpdateFilter updates an existing IP filter
func (s *IPFilterService) UpdateFilter(ctx context.Context, filter IPFilter) error {
	parserFilter := s.toParserFilter(filter)

	// Validate input
	if err := parsers.ValidateIPFilter(parserFilter); err != nil {
		return fmt.Errorf("invalid IP filter: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// For RTX routers, update is done by re-running the filter command
	// This will overwrite the existing filter with the same number
	cmd := parsers.BuildIPFilterCommand(parserFilter)
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Updating IP filter with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to update IP filter: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("filter updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// DeleteFilter removes an IP filter
func (s *IPFilterService) DeleteFilter(ctx context.Context, number int) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cmd := parsers.BuildDeleteIPFilterCommand(number)
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Deleting IP filter with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete IP filter: %w", err)
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
			return fmt.Errorf("filter deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ListFilters retrieves all IP filters
func (s *IPFilterService) ListFilters(ctx context.Context) ([]IPFilter, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cmd := parsers.BuildShowIPFilterCommand()
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Listing IP filters with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list IP filters: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("IP filters raw output: %q", string(output))

	// Parse the output
	parserFilters, err := parsers.ParseIPFilterConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IP filters: %w", err)
	}

	// Convert parsers.IPFilter to client.IPFilter
	filters := make([]IPFilter, len(parserFilters))
	for i, pf := range parserFilters {
		filters[i] = s.fromParserFilter(pf)
	}

	return filters, nil
}

// CreateDynamicFilter creates a new dynamic IP filter
func (s *IPFilterService) CreateDynamicFilter(ctx context.Context, filter IPFilterDynamic) error {
	// Convert client.IPFilterDynamic to parsers.IPFilterDynamic
	parserFilter := s.toParserDynamicFilter(filter)

	// Validate input - only validate Form 1 filters (those with protocol)
	// Form 2 filters use filter lists instead of protocol
	if len(filter.FilterList) == 0 {
		if err := parsers.ValidateIPFilterDynamic(parserFilter); err != nil {
			return fmt.Errorf("invalid dynamic IP filter: %w", err)
		}
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Build and execute filter creation command using extended builder
	cmd := parsers.BuildIPFilterDynamicCommandExtended(parserFilter)
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Creating dynamic IP filter with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to create dynamic IP filter: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("dynamic filter created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetDynamicFilter retrieves a dynamic IP filter configuration
func (s *IPFilterService) GetDynamicFilter(ctx context.Context, number int) (*IPFilterDynamic, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Use the same command as list but filter by number
	cmd := parsers.BuildShowIPFilterCommand()
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Getting dynamic IP filter with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get dynamic IP filter: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Dynamic IP filter raw output: %q", string(output))

	// Parse the output using extended parser to handle both Form 1 and Form 2
	parserFilters, err := parsers.ParseIPFilterDynamicConfigExtended(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse dynamic IP filter: %w", err)
	}

	// Find the specific filter
	for _, pf := range parserFilters {
		if pf.Number == number {
			filter := s.fromParserDynamicFilter(pf)
			return &filter, nil
		}
	}

	return nil, fmt.Errorf("dynamic IP filter %d not found", number)
}

// DeleteDynamicFilter removes a dynamic IP filter
func (s *IPFilterService) DeleteDynamicFilter(ctx context.Context, number int) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cmd := parsers.BuildDeleteIPFilterDynamicCommand(number)
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Deleting dynamic IP filter with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete dynamic IP filter: %w", err)
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
			return fmt.Errorf("dynamic filter deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ListDynamicFilters retrieves all dynamic IP filters
func (s *IPFilterService) ListDynamicFilters(ctx context.Context) ([]IPFilterDynamic, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cmd := parsers.BuildShowIPFilterCommand()
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Listing dynamic IP filters with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list dynamic IP filters: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Dynamic IP filters raw output: %q", string(output))

	// Parse the output using extended parser to handle both Form 1 and Form 2
	parserFilters, err := parsers.ParseIPFilterDynamicConfigExtended(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse dynamic IP filters: %w", err)
	}

	// Convert parsers.IPFilterDynamic to client.IPFilterDynamic
	filters := make([]IPFilterDynamic, len(parserFilters))
	for i, pf := range parserFilters {
		filters[i] = s.fromParserDynamicFilter(pf)
	}

	return filters, nil
}

// toParserFilter converts client.IPFilter to parsers.IPFilter
func (s *IPFilterService) toParserFilter(filter IPFilter) parsers.IPFilter {
	return parsers.IPFilter{
		Number:        filter.Number,
		Action:        filter.Action,
		SourceAddress: filter.SourceAddress,
		SourceMask:    filter.SourceMask,
		DestAddress:   filter.DestAddress,
		DestMask:      filter.DestMask,
		Protocol:      filter.Protocol,
		SourcePort:    filter.SourcePort,
		DestPort:      filter.DestPort,
		Established:   filter.Established,
	}
}

// fromParserFilter converts parsers.IPFilter to client.IPFilter
func (s *IPFilterService) fromParserFilter(pf parsers.IPFilter) IPFilter {
	return IPFilter{
		Number:        pf.Number,
		Action:        pf.Action,
		SourceAddress: pf.SourceAddress,
		SourceMask:    pf.SourceMask,
		DestAddress:   pf.DestAddress,
		DestMask:      pf.DestMask,
		Protocol:      pf.Protocol,
		SourcePort:    pf.SourcePort,
		DestPort:      pf.DestPort,
		Established:   pf.Established,
	}
}

// toParserDynamicFilter converts client.IPFilterDynamic to parsers.IPFilterDynamic
func (s *IPFilterService) toParserDynamicFilter(filter IPFilterDynamic) parsers.IPFilterDynamic {
	return parsers.IPFilterDynamic{
		Number:        filter.Number,
		Source:        filter.Source,
		Dest:          filter.Dest,
		Protocol:      filter.Protocol,
		SyslogOn:      filter.SyslogOn,
		FilterList:    filter.FilterList,
		InFilterList:  filter.InFilterList,
		OutFilterList: filter.OutFilterList,
		Timeout:       filter.Timeout,
	}
}

// fromParserDynamicFilter converts parsers.IPFilterDynamic to client.IPFilterDynamic
func (s *IPFilterService) fromParserDynamicFilter(pf parsers.IPFilterDynamic) IPFilterDynamic {
	return IPFilterDynamic{
		Number:        pf.Number,
		Source:        pf.Source,
		Dest:          pf.Dest,
		Protocol:      pf.Protocol,
		SyslogOn:      pf.SyslogOn,
		FilterList:    pf.FilterList,
		InFilterList:  pf.InFilterList,
		OutFilterList: pf.OutFilterList,
		Timeout:       pf.Timeout,
	}
}

// CreateAccessListExtended creates a new IPv4 extended access list
func (s *IPFilterService) CreateAccessListExtended(ctx context.Context, acl AccessListExtended) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create each entry as an IP filter
	for _, entry := range acl.Entries {
		parserEntry := s.toParserExtendedEntry(entry)
		cmd := parsers.BuildAccessListExtendedEntryCommand(parserEntry)
		logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Creating ACL entry with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to create ACL entry: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("ACL created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetAccessListExtended retrieves an IPv4 extended access list
func (s *IPFilterService) GetAccessListExtended(ctx context.Context, name string) (*AccessListExtended, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Get all IP filters and filter by name/sequence pattern
	cmd := parsers.BuildShowIPFilterCommand()
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Getting ACL with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get ACL: %w", err)
	}

	filters, err := parsers.ParseIPFilterConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ACL: %w", err)
	}

	// Convert filters to ACL entries
	acl := &AccessListExtended{
		Name:    name,
		Entries: make([]AccessListExtendedEntry, 0),
	}

	for _, filter := range filters {
		acl.Entries = append(acl.Entries, s.fromParserFilterToExtendedEntry(filter))
	}

	if len(acl.Entries) == 0 {
		return nil, fmt.Errorf("access list %s not found", name)
	}

	return acl, nil
}

// UpdateAccessListExtended updates an existing IPv4 extended access list
func (s *IPFilterService) UpdateAccessListExtended(ctx context.Context, acl AccessListExtended) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Update each entry (re-create)
	for _, entry := range acl.Entries {
		parserEntry := s.toParserExtendedEntry(entry)
		cmd := parsers.BuildAccessListExtendedEntryCommand(parserEntry)
		logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Updating ACL entry with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to update ACL entry: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("ACL updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// DeleteAccessListExtended removes an IPv4 extended access list
func (s *IPFilterService) DeleteAccessListExtended(ctx context.Context, name string, filterNums []int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Delete each filter by number
	for _, num := range filterNums {
		cmd := parsers.BuildDeleteIPFilterCommand(num)
		logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Deleting ACL entry with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to delete ACL entry: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			if strings.Contains(strings.ToLower(string(output)), "not found") {
				continue
			}
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("ACL deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// CreateAccessListExtendedIPv6 creates a new IPv6 extended access list
func (s *IPFilterService) CreateAccessListExtendedIPv6(ctx context.Context, acl AccessListExtendedIPv6) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create each entry as an IPv6 filter
	for _, entry := range acl.Entries {
		parserEntry := s.toParserExtendedIPv6Entry(entry)
		cmd := parsers.BuildAccessListExtendedIPv6EntryCommand(parserEntry)
		logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Creating IPv6 ACL entry with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to create IPv6 ACL entry: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("IPv6 ACL created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetAccessListExtendedIPv6 retrieves an IPv6 extended access list
func (s *IPFilterService) GetAccessListExtendedIPv6(ctx context.Context, name string) (*AccessListExtendedIPv6, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cmd := parsers.BuildShowIPv6FilterCommand()
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Getting IPv6 ACL with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPv6 ACL: %w", err)
	}

	filters, err := parsers.ParseIPv6FilterConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IPv6 ACL: %w", err)
	}

	acl := &AccessListExtendedIPv6{
		Name:    name,
		Entries: make([]AccessListExtendedIPv6Entry, 0),
	}

	for _, filter := range filters {
		acl.Entries = append(acl.Entries, s.fromParserFilterToExtendedIPv6Entry(filter))
	}

	if len(acl.Entries) == 0 {
		return nil, fmt.Errorf("IPv6 access list %s not found", name)
	}

	return acl, nil
}

// UpdateAccessListExtendedIPv6 updates an existing IPv6 extended access list
func (s *IPFilterService) UpdateAccessListExtendedIPv6(ctx context.Context, acl AccessListExtendedIPv6) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for _, entry := range acl.Entries {
		parserEntry := s.toParserExtendedIPv6Entry(entry)
		cmd := parsers.BuildAccessListExtendedIPv6EntryCommand(parserEntry)
		logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Updating IPv6 ACL entry with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to update IPv6 ACL entry: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("IPv6 ACL updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// DeleteAccessListExtendedIPv6 removes an IPv6 extended access list
func (s *IPFilterService) DeleteAccessListExtendedIPv6(ctx context.Context, name string, filterNums []int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for _, num := range filterNums {
		cmd := parsers.BuildDeleteIPv6FilterCommand(num)
		logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Deleting IPv6 ACL entry with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to delete IPv6 ACL entry: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			if strings.Contains(strings.ToLower(string(output)), "not found") {
				continue
			}
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("IPv6 ACL deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// CreateIPFilterDynamicConfig creates the IP filter dynamic configuration
func (s *IPFilterService) CreateIPFilterDynamicConfig(ctx context.Context, config IPFilterDynamicConfig) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for _, entry := range config.Entries {
		filter := IPFilterDynamic{
			Number:        entry.Number,
			Source:        entry.Source,
			Dest:          entry.Dest,
			Protocol:      entry.Protocol,
			SyslogOn:      entry.Syslog,
			FilterList:    entry.FilterList,
			InFilterList:  entry.InFilterList,
			OutFilterList: entry.OutFilterList,
			Timeout:       entry.Timeout,
		}
		if err := s.CreateDynamicFilter(ctx, filter); err != nil {
			return err
		}
	}

	return nil
}

// GetIPFilterDynamicConfig retrieves all dynamic IP filters
func (s *IPFilterService) GetIPFilterDynamicConfig(ctx context.Context) (*IPFilterDynamicConfig, error) {
	filters, err := s.ListDynamicFilters(ctx)
	if err != nil {
		return nil, err
	}

	config := &IPFilterDynamicConfig{
		Entries: make([]IPFilterDynamicEntry, 0, len(filters)),
	}

	for _, filter := range filters {
		config.Entries = append(config.Entries, IPFilterDynamicEntry{
			Number:        filter.Number,
			Source:        filter.Source,
			Dest:          filter.Dest,
			Protocol:      filter.Protocol,
			Syslog:        filter.SyslogOn,
			FilterList:    filter.FilterList,
			InFilterList:  filter.InFilterList,
			OutFilterList: filter.OutFilterList,
			Timeout:       filter.Timeout,
		})
	}

	return config, nil
}

// UpdateIPFilterDynamicConfig updates the IP filter dynamic configuration
func (s *IPFilterService) UpdateIPFilterDynamicConfig(ctx context.Context, config IPFilterDynamicConfig) error {
	// Re-create all entries
	for _, entry := range config.Entries {
		filter := IPFilterDynamic{
			Number:        entry.Number,
			Source:        entry.Source,
			Dest:          entry.Dest,
			Protocol:      entry.Protocol,
			SyslogOn:      entry.Syslog,
			FilterList:    entry.FilterList,
			InFilterList:  entry.InFilterList,
			OutFilterList: entry.OutFilterList,
			Timeout:       entry.Timeout,
		}

		parserFilter := s.toParserDynamicFilter(filter)
		cmd := parsers.BuildIPFilterDynamicCommandExtended(parserFilter)
		logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Updating dynamic filter with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to update dynamic filter: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("dynamic filters updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// DeleteIPFilterDynamicConfig removes all IP filter dynamic entries
func (s *IPFilterService) DeleteIPFilterDynamicConfig(ctx context.Context, filterNums []int) error {
	for _, num := range filterNums {
		if err := s.DeleteDynamicFilter(ctx, num); err != nil {
			if !strings.Contains(err.Error(), "not found") {
				return err
			}
		}
	}
	return nil
}

// CreateIPv6FilterDynamicConfig creates the IPv6 filter dynamic configuration
func (s *IPFilterService) CreateIPv6FilterDynamicConfig(ctx context.Context, config IPv6FilterDynamicConfig) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for _, entry := range config.Entries {
		parserFilter := parsers.IPFilterDynamic{
			Number:   entry.Number,
			Source:   entry.Source,
			Dest:     entry.Dest,
			Protocol: entry.Protocol,
			SyslogOn: entry.Syslog,
		}
		cmd := parsers.BuildIPv6FilterDynamicCommand(parserFilter)
		logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Creating IPv6 dynamic filter with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to create IPv6 dynamic filter: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("IPv6 dynamic filters created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetIPv6FilterDynamicConfig retrieves all IPv6 dynamic filters
func (s *IPFilterService) GetIPv6FilterDynamicConfig(ctx context.Context) (*IPv6FilterDynamicConfig, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cmd := parsers.BuildShowIPv6FilterCommand()
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Getting IPv6 dynamic filters with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPv6 dynamic filters: %w", err)
	}

	filters, err := parsers.ParseIPv6FilterDynamicConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IPv6 dynamic filters: %w", err)
	}

	config := &IPv6FilterDynamicConfig{
		Entries: make([]IPv6FilterDynamicEntry, 0, len(filters)),
	}

	for _, filter := range filters {
		config.Entries = append(config.Entries, IPv6FilterDynamicEntry{
			Number:   filter.Number,
			Source:   filter.Source,
			Dest:     filter.Dest,
			Protocol: filter.Protocol,
			Syslog:   filter.SyslogOn,
		})
	}

	return config, nil
}

// UpdateIPv6FilterDynamicConfig updates the IPv6 filter dynamic configuration
func (s *IPFilterService) UpdateIPv6FilterDynamicConfig(ctx context.Context, config IPv6FilterDynamicConfig) error {
	for _, entry := range config.Entries {
		parserFilter := parsers.IPFilterDynamic{
			Number:   entry.Number,
			Source:   entry.Source,
			Dest:     entry.Dest,
			Protocol: entry.Protocol,
			SyslogOn: entry.Syslog,
		}
		cmd := parsers.BuildIPv6FilterDynamicCommand(parserFilter)
		logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Updating IPv6 dynamic filter with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to update IPv6 dynamic filter: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("IPv6 dynamic filters updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// DeleteIPv6FilterDynamicConfig removes all IPv6 filter dynamic entries
func (s *IPFilterService) DeleteIPv6FilterDynamicConfig(ctx context.Context, filterNums []int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for _, num := range filterNums {
		cmd := parsers.BuildDeleteIPv6FilterDynamicCommand(num)
		logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Deleting IPv6 dynamic filter with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to delete IPv6 dynamic filter: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			if strings.Contains(strings.ToLower(string(output)), "not found") {
				continue
			}
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("IPv6 dynamic filters deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// CreateInterfaceACL creates ACL bindings for an interface
func (s *IPFilterService) CreateInterfaceACL(ctx context.Context, acl InterfaceACL) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Apply IPv4 inbound filters
	if acl.IPAccessGroupIn != "" || len(acl.DynamicFiltersIn) > 0 {
		// Convert ACL name to filter numbers (assuming they're stored separately)
		// For now, use dynamic filters directly
		if len(acl.DynamicFiltersIn) > 0 {
			cmd := parsers.BuildInterfaceSecureFilterWithDynamicCommand(acl.Interface, "in", nil, acl.DynamicFiltersIn)
			logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Applying interface ACL inbound with command: %s", cmd)

			output, err := s.executor.Run(ctx, cmd)
			if err != nil {
				return fmt.Errorf("failed to apply interface ACL inbound: %w", err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("command failed: %s", string(output))
			}
		}
	}

	// Apply IPv4 outbound filters
	if acl.IPAccessGroupOut != "" || len(acl.DynamicFiltersOut) > 0 {
		if len(acl.DynamicFiltersOut) > 0 {
			cmd := parsers.BuildInterfaceSecureFilterWithDynamicCommand(acl.Interface, "out", nil, acl.DynamicFiltersOut)
			logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Applying interface ACL outbound with command: %s", cmd)

			output, err := s.executor.Run(ctx, cmd)
			if err != nil {
				return fmt.Errorf("failed to apply interface ACL outbound: %w", err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("command failed: %s", string(output))
			}
		}
	}

	// Apply IPv6 inbound filters
	if acl.IPv6AccessGroupIn != "" || len(acl.IPv6DynamicFiltersIn) > 0 {
		if len(acl.IPv6DynamicFiltersIn) > 0 {
			cmd := parsers.BuildInterfaceIPv6SecureFilterWithDynamicCommand(acl.Interface, "in", nil, acl.IPv6DynamicFiltersIn)
			logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Applying interface IPv6 ACL inbound with command: %s", cmd)

			output, err := s.executor.Run(ctx, cmd)
			if err != nil {
				return fmt.Errorf("failed to apply interface IPv6 ACL inbound: %w", err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("command failed: %s", string(output))
			}
		}
	}

	// Apply IPv6 outbound filters
	if acl.IPv6AccessGroupOut != "" || len(acl.IPv6DynamicFiltersOut) > 0 {
		if len(acl.IPv6DynamicFiltersOut) > 0 {
			cmd := parsers.BuildInterfaceIPv6SecureFilterWithDynamicCommand(acl.Interface, "out", nil, acl.IPv6DynamicFiltersOut)
			logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Applying interface IPv6 ACL outbound with command: %s", cmd)

			output, err := s.executor.Run(ctx, cmd)
			if err != nil {
				return fmt.Errorf("failed to apply interface IPv6 ACL outbound: %w", err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("command failed: %s", string(output))
			}
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("interface ACL created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetInterfaceACL retrieves ACL bindings for an interface
func (s *IPFilterService) GetInterfaceACL(ctx context.Context, iface string) (*InterfaceACL, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Get IP secure filter configuration
	cmd := parsers.BuildShowIPFilterCommand()
	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get interface ACL: %w", err)
	}

	ipFilters, err := parsers.ParseInterfaceSecureFilter(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse interface ACL: %w", err)
	}

	acl := &InterfaceACL{
		Interface: iface,
	}

	// Get IPv4 filters for the interface
	if ifaceFilters, ok := ipFilters[iface]; ok {
		if inFilters, ok := ifaceFilters["in"]; ok {
			acl.DynamicFiltersIn = inFilters
		}
		if outFilters, ok := ifaceFilters["out"]; ok {
			acl.DynamicFiltersOut = outFilters
		}
	}

	// Get IPv6 secure filter configuration
	cmd = parsers.BuildShowIPv6FilterCommand()
	output, err = s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get interface IPv6 ACL: %w", err)
	}

	ipv6Filters, err := parsers.ParseInterfaceIPv6SecureFilter(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse interface IPv6 ACL: %w", err)
	}

	// Get IPv6 filters for the interface
	if ifaceFilters, ok := ipv6Filters[iface]; ok {
		if inFilters, ok := ifaceFilters["in"]; ok {
			acl.IPv6DynamicFiltersIn = inFilters
		}
		if outFilters, ok := ifaceFilters["out"]; ok {
			acl.IPv6DynamicFiltersOut = outFilters
		}
	}

	return acl, nil
}

// UpdateInterfaceACL updates ACL bindings for an interface
func (s *IPFilterService) UpdateInterfaceACL(ctx context.Context, acl InterfaceACL) error {
	// Delete existing and re-create
	if err := s.DeleteInterfaceACL(ctx, acl.Interface); err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return err
		}
	}
	return s.CreateInterfaceACL(ctx, acl)
}

// DeleteInterfaceACL removes ACL bindings from an interface
func (s *IPFilterService) DeleteInterfaceACL(ctx context.Context, iface string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Remove IPv4 inbound filters
	cmd := parsers.BuildDeleteInterfaceSecureFilterCommand(iface, "in")
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Removing interface ACL inbound with command: %s", cmd)
	_, _ = s.executor.Run(ctx, cmd) // Ignore error if not configured

	// Remove IPv4 outbound filters
	cmd = parsers.BuildDeleteInterfaceSecureFilterCommand(iface, "out")
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Removing interface ACL outbound with command: %s", cmd)
	_, _ = s.executor.Run(ctx, cmd) // Ignore error if not configured

	// Remove IPv6 inbound filters
	cmd = parsers.BuildDeleteInterfaceIPv6SecureFilterCommand(iface, "in")
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Removing interface IPv6 ACL inbound with command: %s", cmd)
	_, _ = s.executor.Run(ctx, cmd) // Ignore error if not configured

	// Remove IPv6 outbound filters
	cmd = parsers.BuildDeleteInterfaceIPv6SecureFilterCommand(iface, "out")
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Removing interface IPv6 ACL outbound with command: %s", cmd)
	_, _ = s.executor.Run(ctx, cmd) // Ignore error if not configured

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("interface ACL deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// toParserExtendedEntry converts client.AccessListExtendedEntry to parsers.AccessListExtendedEntry
func (s *IPFilterService) toParserExtendedEntry(entry AccessListExtendedEntry) parsers.AccessListExtendedEntry {
	return parsers.AccessListExtendedEntry{
		Sequence:              entry.Sequence,
		AceRuleAction:         entry.AceRuleAction,
		AceRuleProtocol:       entry.AceRuleProtocol,
		SourceAny:             entry.SourceAny,
		SourcePrefix:          entry.SourcePrefix,
		SourcePrefixMask:      entry.SourcePrefixMask,
		SourcePortEqual:       entry.SourcePortEqual,
		SourcePortRange:       entry.SourcePortRange,
		DestinationAny:        entry.DestinationAny,
		DestinationPrefix:     entry.DestinationPrefix,
		DestinationPrefixMask: entry.DestinationPrefixMask,
		DestinationPortEqual:  entry.DestinationPortEqual,
		DestinationPortRange:  entry.DestinationPortRange,
		Established:           entry.Established,
		Log:                   entry.Log,
	}
}

// fromParserFilterToExtendedEntry converts parsers.IPFilter to client.AccessListExtendedEntry
func (s *IPFilterService) fromParserFilterToExtendedEntry(filter parsers.IPFilter) AccessListExtendedEntry {
	entry := AccessListExtendedEntry{
		Sequence:        filter.Number,
		AceRuleProtocol: filter.Protocol,
		Established:     filter.Established,
	}

	// Map action
	if filter.Action == "reject" {
		entry.AceRuleAction = "deny"
	} else {
		entry.AceRuleAction = "permit"
	}

	// Map source
	if filter.SourceAddress == "*" {
		entry.SourceAny = true
	} else {
		entry.SourcePrefix = filter.SourceAddress
		entry.SourcePrefixMask = filter.SourceMask
	}

	// Map destination
	if filter.DestAddress == "*" {
		entry.DestinationAny = true
	} else {
		entry.DestinationPrefix = filter.DestAddress
		entry.DestinationPrefixMask = filter.DestMask
	}

	// Map ports
	entry.SourcePortEqual = filter.SourcePort
	entry.DestinationPortEqual = filter.DestPort

	return entry
}

// toParserExtendedIPv6Entry converts client.AccessListExtendedIPv6Entry to parsers.AccessListExtendedIPv6Entry
func (s *IPFilterService) toParserExtendedIPv6Entry(entry AccessListExtendedIPv6Entry) parsers.AccessListExtendedIPv6Entry {
	return parsers.AccessListExtendedIPv6Entry{
		Sequence:                entry.Sequence,
		AceRuleAction:           entry.AceRuleAction,
		AceRuleProtocol:         entry.AceRuleProtocol,
		SourceAny:               entry.SourceAny,
		SourcePrefix:            entry.SourcePrefix,
		SourcePrefixLength:      entry.SourcePrefixLength,
		SourcePortEqual:         entry.SourcePortEqual,
		SourcePortRange:         entry.SourcePortRange,
		DestinationAny:          entry.DestinationAny,
		DestinationPrefix:       entry.DestinationPrefix,
		DestinationPrefixLength: entry.DestinationPrefixLength,
		DestinationPortEqual:    entry.DestinationPortEqual,
		DestinationPortRange:    entry.DestinationPortRange,
		Established:             entry.Established,
		Log:                     entry.Log,
	}
}

// fromParserFilterToExtendedIPv6Entry converts parsers.IPFilter to client.AccessListExtendedIPv6Entry
func (s *IPFilterService) fromParserFilterToExtendedIPv6Entry(filter parsers.IPFilter) AccessListExtendedIPv6Entry {
	entry := AccessListExtendedIPv6Entry{
		Sequence:        filter.Number,
		AceRuleProtocol: filter.Protocol,
	}

	// Map action
	if filter.Action == "reject" {
		entry.AceRuleAction = "deny"
	} else {
		entry.AceRuleAction = "permit"
	}

	// Map source
	if filter.SourceAddress == "*" {
		entry.SourceAny = true
	} else {
		entry.SourcePrefix = filter.SourceAddress
	}

	// Map destination
	if filter.DestAddress == "*" {
		entry.DestinationAny = true
	} else {
		entry.DestinationPrefix = filter.DestAddress
	}

	// Map ports
	entry.SourcePortEqual = filter.SourcePort
	entry.DestinationPortEqual = filter.DestPort

	return entry
}

// CreateIPv6Filter creates a new static IPv6 filter
func (s *IPFilterService) CreateIPv6Filter(ctx context.Context, filter IPFilter) error {
	// Convert client.IPFilter to parsers.IPFilter
	parserFilter := s.toParserFilter(filter)

	// Validate input
	if err := parsers.ValidateIPFilterNumber(parserFilter.Number); err != nil {
		return fmt.Errorf("invalid IPv6 filter: %w", err)
	}
	if err := parsers.ValidateIPFilterAction(parserFilter.Action); err != nil {
		return fmt.Errorf("invalid IPv6 filter: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Build and execute filter creation command
	cmd := parsers.BuildIPv6FilterCommand(parserFilter)
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Creating IPv6 filter with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to create IPv6 filter: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("IPv6 filter created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetIPv6Filter retrieves an IPv6 filter configuration by number
func (s *IPFilterService) GetIPv6Filter(ctx context.Context, number int) (*IPFilter, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cmd := parsers.BuildShowIPv6FilterByNumberCommand(number)
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Getting IPv6 filter with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get IPv6 filter: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("IPv6 filter raw output: %q", string(output))

	// Parse the output
	parserFilters, err := parsers.ParseIPv6FilterConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IPv6 filter: %w", err)
	}

	// Find the specific filter
	for _, pf := range parserFilters {
		if pf.Number == number {
			filter := s.fromParserFilter(pf)
			return &filter, nil
		}
	}

	return nil, fmt.Errorf("IPv6 filter %d not found", number)
}

// UpdateIPv6Filter updates an existing IPv6 filter
func (s *IPFilterService) UpdateIPv6Filter(ctx context.Context, filter IPFilter) error {
	parserFilter := s.toParserFilter(filter)

	// Validate input
	if err := parsers.ValidateIPFilterNumber(parserFilter.Number); err != nil {
		return fmt.Errorf("invalid IPv6 filter: %w", err)
	}
	if err := parsers.ValidateIPFilterAction(parserFilter.Action); err != nil {
		return fmt.Errorf("invalid IPv6 filter: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// For RTX routers, update is done by re-running the filter command
	// This will overwrite the existing filter with the same number
	cmd := parsers.BuildIPv6FilterCommand(parserFilter)
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Updating IPv6 filter with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to update IPv6 filter: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("IPv6 filter updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// DeleteIPv6Filter removes an IPv6 filter
func (s *IPFilterService) DeleteIPv6Filter(ctx context.Context, number int) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cmd := parsers.BuildDeleteIPv6FilterCommand(number)
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Deleting IPv6 filter with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete IPv6 filter: %w", err)
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
			return fmt.Errorf("IPv6 filter deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ListIPv6Filters retrieves all static IPv6 filters
func (s *IPFilterService) ListIPv6Filters(ctx context.Context) ([]IPFilter, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cmd := parsers.BuildShowIPv6FilterCommand()
	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("Listing IPv6 filters with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list IPv6 filters: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "UipUfilterService").Msgf("IPv6 filters raw output: %q", string(output))

	// Parse the output
	parserFilters, err := parsers.ParseIPv6FilterConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IPv6 filters: %w", err)
	}

	// Convert parsers.IPFilter to client.IPFilter
	filters := make([]IPFilter, len(parserFilters))
	for i, pf := range parserFilters {
		filters[i] = s.fromParserFilter(pf)
	}

	return filters, nil
}
