package client

import (
	"context"
	"fmt"
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// EthernetFilterService handles Ethernet filter operations
type EthernetFilterService struct {
	executor Executor
	client   *rtxClient // Reference to the main client for save functionality
}

// NewEthernetFilterService creates a new Ethernet filter service instance
func NewEthernetFilterService(executor Executor, client *rtxClient) *EthernetFilterService {
	return &EthernetFilterService{
		executor: executor,
		client:   client,
	}
}

// CreateFilter creates a new Ethernet filter
func (s *EthernetFilterService) CreateFilter(ctx context.Context, filter EthernetFilter) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Convert client.EthernetFilter to parsers.EthernetFilter
	parserFilter := s.toParserFilter(filter)

	// Validate input
	if err := parsers.ValidateEthernetFilter(parserFilter); err != nil {
		return fmt.Errorf("invalid filter: %w", err)
	}

	// Build and execute filter creation command
	cmd := parsers.BuildEthernetFilterCommand(parserFilter)
	logging.FromContext(ctx).Debug().Str("service", "ethernet_filter").Msgf("Creating Ethernet filter with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to create Ethernet filter: %w", err)
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

// GetFilter retrieves an Ethernet filter configuration
func (s *EthernetFilterService) GetFilter(ctx context.Context, number int) (*EthernetFilter, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Validate filter number
	if err := parsers.ValidateEthernetFilterNumber(number); err != nil {
		return nil, fmt.Errorf("invalid filter number: %w", err)
	}

	cmd := parsers.BuildShowEthernetFilterCommand(number)
	logging.FromContext(ctx).Debug().Str("service", "ethernet_filter").Msgf("Getting Ethernet filter with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get Ethernet filter: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "ethernet_filter").Msgf("Ethernet filter raw output: %q", string(output))

	parserFilter, err := parsers.ParseSingleEthernetFilter(string(output), number)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Ethernet filter: %w", err)
	}

	// Convert parsers.EthernetFilter to client.EthernetFilter
	filter := s.fromParserFilter(*parserFilter)
	return &filter, nil
}

// UpdateFilter updates an existing Ethernet filter
func (s *EthernetFilterService) UpdateFilter(ctx context.Context, filter EthernetFilter) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Convert client.EthernetFilter to parsers.EthernetFilter
	parserFilter := s.toParserFilter(filter)

	// Validate input
	if err := parsers.ValidateEthernetFilter(parserFilter); err != nil {
		return fmt.Errorf("invalid filter: %w", err)
	}

	// RTX routers allow re-running the filter command to update values
	cmd := parsers.BuildEthernetFilterCommand(parserFilter)
	logging.FromContext(ctx).Debug().Str("service", "ethernet_filter").Msgf("Updating Ethernet filter with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to update Ethernet filter: %w", err)
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

// DeleteFilter removes an Ethernet filter
func (s *EthernetFilterService) DeleteFilter(ctx context.Context, number int) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Validate filter number
	if err := parsers.ValidateEthernetFilterNumber(number); err != nil {
		return fmt.Errorf("invalid filter number: %w", err)
	}

	cmd := parsers.BuildDeleteEthernetFilterCommand(number)
	logging.FromContext(ctx).Debug().Str("service", "ethernet_filter").Msgf("Deleting Ethernet filter with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete Ethernet filter: %w", err)
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

// ListFilters retrieves all Ethernet filters
func (s *EthernetFilterService) ListFilters(ctx context.Context) ([]EthernetFilter, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cmd := parsers.BuildShowAllEthernetFiltersCommand()
	logging.FromContext(ctx).Debug().Str("service", "ethernet_filter").Msgf("Listing Ethernet filters with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list Ethernet filters: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "ethernet_filter").Msgf("Ethernet filters raw output: %q", string(output))

	parserFilters, err := parsers.ParseEthernetFilterConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse Ethernet filters: %w", err)
	}

	// Convert parsers.EthernetFilter to client.EthernetFilter
	filters := make([]EthernetFilter, len(parserFilters))
	for i, pf := range parserFilters {
		filters[i] = s.fromParserFilter(pf)
	}

	return filters, nil
}

// toParserFilter converts client.EthernetFilter to parsers.EthernetFilter
func (s *EthernetFilterService) toParserFilter(filter EthernetFilter) parsers.EthernetFilter {
	return parsers.EthernetFilter{
		Number:    filter.Number,
		Action:    filter.Action,
		SourceMAC: filter.SourceMAC,
		DestMAC:   filter.DestMAC,
		EtherType: filter.EtherType,
		VlanID:    filter.VlanID,
	}
}

// fromParserFilter converts parsers.EthernetFilter to client.EthernetFilter
func (s *EthernetFilterService) fromParserFilter(pf parsers.EthernetFilter) EthernetFilter {
	return EthernetFilter{
		Number:    pf.Number,
		Action:    pf.Action,
		SourceMAC: pf.SourceMAC,
		DestMAC:   pf.DestMAC,
		EtherType: pf.EtherType,
		VlanID:    pf.VlanID,
	}
}

// CreateAccessListMAC creates a new MAC access list
func (s *EthernetFilterService) CreateAccessListMAC(ctx context.Context, acl AccessListMAC) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create each entry as an Ethernet filter
	for _, entry := range acl.Entries {
		parserEntry := s.toParserMACEntry(entry)
		cmd := parsers.BuildAccessListMACEntryCommand(parserEntry)
		logging.FromContext(ctx).Debug().Str("service", "ethernet_filter").Msgf("Creating MAC ACL entry with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to create MAC ACL entry: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Apply to interface if requested
	if acl.Apply != nil {
		filterNums := acl.Apply.FilterIDs
		if len(filterNums) == 0 {
			for _, entry := range acl.Entries {
				num := entry.FilterID
				if num == 0 {
					num = entry.Sequence
				}
				if num > 0 {
					filterNums = append(filterNums, num)
				}
			}
		}
		cmd := parsers.BuildMACAccessListInterfaceCommand(acl.Apply.Interface, acl.Apply.Direction, filterNums)
		if cmd != "" {
			logging.FromContext(ctx).Debug().Str("service", "ethernet_filter").Msgf("Applying MAC ACL to interface with command: %s", cmd)
			if _, err := s.executor.Run(ctx, cmd); err != nil {
				return fmt.Errorf("failed to apply MAC ACL: %w", err)
			}
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("MAC ACL created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetAccessListMAC retrieves a MAC access list
func (s *EthernetFilterService) GetAccessListMAC(ctx context.Context, name string) (*AccessListMAC, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Get all Ethernet filters
	cmd := parsers.BuildShowAllEthernetFiltersCommand()
	logging.FromContext(ctx).Debug().Str("service", "ethernet_filter").Msgf("Getting MAC ACL with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get MAC ACL: %w", err)
	}

	filters, err := parsers.ParseEthernetFilterConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse MAC ACL: %w", err)
	}

	// Convert filters to ACL entries
	acl := &AccessListMAC{
		Name:    name,
		Entries: make([]AccessListMACEntry, 0),
	}

	for _, filter := range filters {
		acl.Entries = append(acl.Entries, s.fromParserFilterToMACEntry(filter))
	}

	// Attempt to populate apply from interface filter config
	if len(acl.Entries) > 0 {
		allIDs := make(map[int]struct{}, len(acl.Entries))
		var ordered []int
		for _, e := range acl.Entries {
			id := e.FilterID
			if id == 0 {
				id = e.Sequence
			}
			if id > 0 {
				allIDs[id] = struct{}{}
				ordered = append(ordered, id)
			}
		}

		intfCmd := parsers.BuildShowInterfaceEthernetFilterCommand()
		logging.FromContext(ctx).Debug().Str("service", "ethernet_filter").Msgf("Getting interface ethernet filter bindings with command: %s", intfCmd)
		if output, err := s.executor.Run(ctx, intfCmd); err == nil {
			if bindings, err := parsers.ParseInterfaceEthernetFilter(string(output)); err == nil {
				for iface, dirs := range bindings {
					for dir, nums := range dirs {
						match := true
						for _, n := range nums {
							if _, ok := allIDs[n]; !ok {
								match = false
								break
							}
						}
						if match {
							acl.Apply = &MACApply{
								Interface: iface,
								Direction: dir,
								FilterIDs: nums,
							}
							break
						}
					}
					if acl.Apply != nil {
						break
					}
				}
			}
		}
	}

	if len(acl.Entries) > 0 {
		acl.FilterID = acl.Entries[0].FilterID
	}

	if len(acl.Entries) == 0 {
		return nil, fmt.Errorf("MAC access list %s not found", name)
	}

	return acl, nil
}

// UpdateAccessListMAC updates an existing MAC access list
func (s *EthernetFilterService) UpdateAccessListMAC(ctx context.Context, acl AccessListMAC) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for _, entry := range acl.Entries {
		parserEntry := s.toParserMACEntry(entry)
		cmd := parsers.BuildAccessListMACEntryCommand(parserEntry)
		logging.FromContext(ctx).Debug().Str("service", "ethernet_filter").Msgf("Updating MAC ACL entry with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to update MAC ACL entry: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Apply to interface if requested
	if acl.Apply != nil {
		filterNums := acl.Apply.FilterIDs
		if len(filterNums) == 0 {
			for _, entry := range acl.Entries {
				num := entry.FilterID
				if num == 0 {
					num = entry.Sequence
				}
				if num > 0 {
					filterNums = append(filterNums, num)
				}
			}
		}
		cmd := parsers.BuildMACAccessListInterfaceCommand(acl.Apply.Interface, acl.Apply.Direction, filterNums)
		if cmd != "" {
			logging.FromContext(ctx).Debug().Str("service", "ethernet_filter").Msgf("Applying MAC ACL to interface with command: %s", cmd)
			if _, err := s.executor.Run(ctx, cmd); err != nil {
				return fmt.Errorf("failed to apply MAC ACL: %w", err)
			}
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("MAC ACL updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// DeleteAccessListMAC removes a MAC access list
func (s *EthernetFilterService) DeleteAccessListMAC(ctx context.Context, name string, filterNums []int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for _, num := range filterNums {
		cmd := parsers.BuildDeleteEthernetFilterCommand(num)
		logging.FromContext(ctx).Debug().Str("service", "ethernet_filter").Msgf("Deleting MAC ACL entry with command: %s", cmd)

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to delete MAC ACL entry: %w", err)
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
			return fmt.Errorf("MAC ACL deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// CreateInterfaceMACACL creates MAC ACL bindings for an interface
func (s *EthernetFilterService) CreateInterfaceMACACL(ctx context.Context, acl InterfaceMACACL) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// This would require parsing the ACL name to get filter numbers
	// For now, we'll use a placeholder implementation
	logging.FromContext(ctx).Debug().Str("service", "ethernet_filter").Msgf("CreateInterfaceMACACL: interface=%s, in=%s, out=%s",
		acl.Interface, acl.MACAccessGroupIn, acl.MACAccessGroupOut)

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("interface MAC ACL created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetInterfaceMACACL retrieves MAC ACL bindings for an interface
func (s *EthernetFilterService) GetInterfaceMACACL(ctx context.Context, iface string) (*InterfaceMACACL, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Get interface Ethernet filter configuration
	cmd := parsers.BuildShowInterfaceEthernetFilterCommand()
	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get interface MAC ACL: %w", err)
	}

	ifaceFilters, err := parsers.ParseInterfaceEthernetFilter(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse interface MAC ACL: %w", err)
	}

	acl := &InterfaceMACACL{
		Interface: iface,
	}

	// Get filters for the interface
	if filters, ok := ifaceFilters[iface]; ok {
		if inFilters, ok := filters["in"]; ok && len(inFilters) > 0 {
			acl.MACAccessGroupIn = "configured"
		}
		if outFilters, ok := filters["out"]; ok && len(outFilters) > 0 {
			acl.MACAccessGroupOut = "configured"
		}
	}

	return acl, nil
}

// UpdateInterfaceMACACL updates MAC ACL bindings for an interface
func (s *EthernetFilterService) UpdateInterfaceMACACL(ctx context.Context, acl InterfaceMACACL) error {
	// Delete existing and re-create
	if err := s.DeleteInterfaceMACACL(ctx, acl.Interface); err != nil {
		if !strings.Contains(err.Error(), "not found") {
			return err
		}
	}
	return s.CreateInterfaceMACACL(ctx, acl)
}

// DeleteInterfaceMACACL removes MAC ACL bindings from an interface
func (s *EthernetFilterService) DeleteInterfaceMACACL(ctx context.Context, iface string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Remove inbound filters
	cmd := parsers.BuildDeleteInterfaceEthernetFilterCommand(iface, "in")
	logging.FromContext(ctx).Debug().Str("service", "ethernet_filter").Msgf("Removing interface MAC ACL inbound with command: %s", cmd)
	_, _ = s.executor.Run(ctx, cmd) // Ignore error if not configured

	// Remove outbound filters
	cmd = parsers.BuildDeleteInterfaceEthernetFilterCommand(iface, "out")
	logging.FromContext(ctx).Debug().Str("service", "ethernet_filter").Msgf("Removing interface MAC ACL outbound with command: %s", cmd)
	_, _ = s.executor.Run(ctx, cmd) // Ignore error if not configured

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("interface MAC ACL deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// toParserMACEntry converts client.AccessListMACEntry to parsers.AccessListMACEntry
func (s *EthernetFilterService) toParserMACEntry(entry AccessListMACEntry) parsers.AccessListMACEntry {
	return parsers.AccessListMACEntry{
		Sequence:               entry.Sequence,
		FilterID:               entry.FilterID,
		AceAction:              entry.AceAction,
		SourceAny:              entry.SourceAny,
		SourceAddress:          entry.SourceAddress,
		SourceAddressMask:      entry.SourceAddressMask,
		DestinationAny:         entry.DestinationAny,
		DestinationAddress:     entry.DestinationAddress,
		DestinationAddressMask: entry.DestinationAddressMask,
		EtherType:              entry.EtherType,
		VlanID:                 entry.VlanID,
		Log:                    entry.Log,
		DHCPType:               entry.DHCPType,
		DHCPScope:              entry.DHCPScope,
		Offset:                 entry.Offset,
		ByteList:               entry.ByteList,
	}
}

// fromParserFilterToMACEntry converts parsers.EthernetFilter to client.AccessListMACEntry
func (s *EthernetFilterService) fromParserFilterToMACEntry(filter parsers.EthernetFilter) AccessListMACEntry {
	entry := AccessListMACEntry{
		Sequence:  filter.Number,
		FilterID:  filter.Number,
		EtherType: filter.EtherType,
		VlanID:    filter.VlanID,
		DHCPType:  filter.DHCPType,
		DHCPScope: filter.DHCPScope,
		Offset:    filter.Offset,
		ByteList:  filter.ByteList,
	}

	// Map action
	entry.AceAction = filter.Action
	if filter.Action == "reject" {
		entry.AceAction = "deny"
	}
	if filter.Action == "pass" {
		entry.AceAction = "permit"
	}

	// Map source
	if filter.DHCPType != "" || filter.SourceMAC == "*" {
		entry.SourceAny = true
	} else {
		entry.SourceAddress = filter.SourceMAC
	}

	// Map destination
	if filter.DHCPType != "" || filter.DestMAC == "*" || filter.DestinationMAC == "*" {
		entry.DestinationAny = true
	} else {
		entry.DestinationAddress = filter.DestMAC
	}

	return entry
}
