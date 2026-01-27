package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// ACLApplyService handles ACL application to interfaces
type ACLApplyService struct {
	executor Executor
	client   *rtxClient
}

// NewACLApplyService creates a new ACL apply service instance
func NewACLApplyService(executor Executor, client *rtxClient) *ACLApplyService {
	return &ACLApplyService{
		executor: executor,
		client:   client,
	}
}

// ApplyFiltersToInterface binds filter sequences to an interface
// The aclType determines which command format to use:
// - ip: "ip <interface> secure filter <direction> <filter_numbers...>"
// - ipv6: "ipv6 <interface> secure filter <direction> <filter_numbers...>"
// - mac: "ethernet <interface> filter <direction> <filter_numbers...>"
func (s *ACLApplyService) ApplyFiltersToInterface(ctx context.Context, iface, direction string, aclType ACLType, filterIDs []int) error {
	logger := logging.FromContext(ctx)

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Validate inputs
	if err := s.ValidateInterface(ctx, iface, aclType); err != nil {
		return err
	}

	if direction != "in" && direction != "out" {
		return fmt.Errorf("invalid direction %q: must be 'in' or 'out'", direction)
	}

	if len(filterIDs) == 0 {
		return fmt.Errorf("at least one filter ID is required")
	}

	// Build command based on ACL type
	cmd := s.buildApplyCommand(iface, direction, aclType, filterIDs)
	if cmd == "" {
		return fmt.Errorf("unsupported ACL type: %s", aclType)
	}

	logger.Debug().
		Str("service", "ACLApplyService").
		Str("operation", "ApplyFiltersToInterface").
		Str("interface", iface).
		Str("direction", direction).
		Str("acl_type", string(aclType)).
		Ints("filter_ids", filterIDs).
		Msgf("Applying filters with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to apply filters to interface %s: %w", iface, err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("filters applied but failed to save configuration: %w", err)
		}
	}

	return nil
}

// RemoveFiltersFromInterface removes filter bindings from an interface
func (s *ACLApplyService) RemoveFiltersFromInterface(ctx context.Context, iface, direction string, aclType ACLType) error {
	logger := logging.FromContext(ctx)

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if direction != "in" && direction != "out" {
		return fmt.Errorf("invalid direction %q: must be 'in' or 'out'", direction)
	}

	// Build delete command based on ACL type
	cmd := s.buildDeleteCommand(iface, direction, aclType)
	if cmd == "" {
		return fmt.Errorf("unsupported ACL type: %s", aclType)
	}

	logger.Debug().
		Str("service", "ACLApplyService").
		Str("operation", "RemoveFiltersFromInterface").
		Str("interface", iface).
		Str("direction", direction).
		Str("acl_type", string(aclType)).
		Msgf("Removing filters with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		// Ignore "not found" errors since filter may not be applied
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return nil
		}
		return fmt.Errorf("failed to remove filters from interface %s: %w", iface, err)
	}

	if len(output) > 0 && containsError(string(output)) {
		// Ignore "not found" errors
		if strings.Contains(strings.ToLower(string(output)), "not found") {
			return nil
		}
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("filters removed but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetInterfaceFilters returns current filter bindings for an interface
func (s *ACLApplyService) GetInterfaceFilters(ctx context.Context, iface, direction string, aclType ACLType) ([]int, error) {
	logger := logging.FromContext(ctx)

	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if direction != "in" && direction != "out" {
		return nil, fmt.Errorf("invalid direction %q: must be 'in' or 'out'", direction)
	}

	// Build show command and parse output based on ACL type
	var cmd string
	var parseFunc func(string) (map[string]map[string][]int, error)

	switch aclType {
	case ACLTypeIP, ACLTypeExtended, ACLTypeIPDynamic:
		cmd = parsers.BuildShowIPFilterCommand()
		parseFunc = parsers.ParseInterfaceSecureFilter
	case ACLTypeIPv6, ACLTypeIPv6Dynamic:
		cmd = parsers.BuildShowIPv6FilterCommand()
		parseFunc = parsers.ParseInterfaceIPv6SecureFilter
	case ACLTypeMAC:
		cmd = parsers.BuildShowInterfaceEthernetFilterCommand()
		parseFunc = parsers.ParseInterfaceEthernetFilter
	default:
		return nil, fmt.Errorf("unsupported ACL type: %s", aclType)
	}

	logger.Debug().
		Str("service", "ACLApplyService").
		Str("operation", "GetInterfaceFilters").
		Str("interface", iface).
		Str("direction", direction).
		Str("acl_type", string(aclType)).
		Msgf("Getting interface filters with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get interface filters: %w", err)
	}

	result, err := parseFunc(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse interface filters: %w", err)
	}

	// Extract filters for specific interface and direction
	if ifaceFilters, ok := result[iface]; ok {
		if filters, ok := ifaceFilters[direction]; ok {
			return filters, nil
		}
	}

	// No filters found for this interface/direction
	return []int{}, nil
}

// ValidateInterface checks if interface exists and supports the ACL type
func (s *ACLApplyService) ValidateInterface(ctx context.Context, iface string, aclType ACLType) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Validate interface name format
	if iface == "" {
		return fmt.Errorf("interface name is required")
	}

	// Extract interface type prefix
	ifaceType := getInterfaceType(iface)
	if ifaceType == "" {
		return fmt.Errorf("invalid interface name %q: expected format like lan1, pp1, bridge1, tunnel1", iface)
	}

	// Validate ACL type compatibility with interface type
	return s.validateACLTypeCompatibility(ifaceType, aclType)
}

// buildApplyCommand builds the appropriate apply command based on ACL type
func (s *ACLApplyService) buildApplyCommand(iface, direction string, aclType ACLType, filterIDs []int) string {
	switch aclType {
	case ACLTypeIP, ACLTypeExtended:
		return parsers.BuildInterfaceSecureFilterCommand(iface, direction, filterIDs)
	case ACLTypeIPv6:
		return parsers.BuildInterfaceIPv6SecureFilterCommand(iface, direction, filterIDs)
	case ACLTypeMAC:
		return parsers.BuildInterfaceEthernetFilterCommand(iface, direction, filterIDs)
	case ACLTypeIPDynamic:
		// Dynamic filters are applied as part of secure filter command with "dynamic" keyword
		return parsers.BuildInterfaceSecureFilterWithDynamicCommand(iface, direction, nil, filterIDs)
	case ACLTypeIPv6Dynamic:
		return parsers.BuildInterfaceIPv6SecureFilterWithDynamicCommand(iface, direction, nil, filterIDs)
	default:
		return ""
	}
}

// buildDeleteCommand builds the appropriate delete command based on ACL type
func (s *ACLApplyService) buildDeleteCommand(iface, direction string, aclType ACLType) string {
	switch aclType {
	case ACLTypeIP, ACLTypeExtended, ACLTypeIPDynamic:
		return parsers.BuildDeleteInterfaceSecureFilterCommand(iface, direction)
	case ACLTypeIPv6, ACLTypeIPv6Dynamic:
		return parsers.BuildDeleteInterfaceIPv6SecureFilterCommand(iface, direction)
	case ACLTypeMAC:
		return parsers.BuildDeleteInterfaceEthernetFilterCommand(iface, direction)
	default:
		return ""
	}
}

// validateACLTypeCompatibility checks if the ACL type is compatible with the interface type
func (s *ACLApplyService) validateACLTypeCompatibility(ifaceType string, aclType ACLType) error {
	switch aclType {
	case ACLTypeMAC:
		// MAC filters are only supported on Ethernet interfaces (lan, bridge)
		// They are NOT supported on PP (Point-to-Point) or Tunnel interfaces
		if ifaceType == "pp" || ifaceType == "tunnel" {
			return fmt.Errorf("MAC ACL type is not supported on %s interfaces", ifaceType)
		}
	case ACLTypeIP, ACLTypeIPv6, ACLTypeExtended, ACLTypeIPDynamic, ACLTypeIPv6Dynamic:
		// IP-based filters are supported on all interface types
		// (lan, pp, bridge, tunnel)
	default:
		return fmt.Errorf("unknown ACL type: %s", aclType)
	}

	return nil
}

// getInterfaceType extracts the interface type from an interface name
// e.g., "lan1" -> "lan", "pp1" -> "pp", "bridge2" -> "bridge", "tunnel1" -> "tunnel"
func getInterfaceType(iface string) string {
	iface = strings.ToLower(iface)

	prefixes := []string{"bridge", "tunnel", "lan", "pp"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(iface, prefix) {
			// Verify that there's a number after the prefix
			rest := iface[len(prefix):]
			if len(rest) > 0 && rest[0] >= '0' && rest[0] <= '9' {
				return prefix
			}
		}
	}

	return ""
}

// ApplyFiltersToInterfaceWithDynamic applies both static and dynamic filters to an interface
// This is useful for combined configurations where both static and dynamic filters are needed
func (s *ACLApplyService) ApplyFiltersToInterfaceWithDynamic(ctx context.Context, iface, direction string, aclType ACLType, staticFilterIDs, dynamicFilterIDs []int) error {
	logger := logging.FromContext(ctx)

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Validate inputs
	if err := s.ValidateInterface(ctx, iface, aclType); err != nil {
		return err
	}

	if direction != "in" && direction != "out" {
		return fmt.Errorf("invalid direction %q: must be 'in' or 'out'", direction)
	}

	if len(staticFilterIDs) == 0 && len(dynamicFilterIDs) == 0 {
		return fmt.Errorf("at least one filter ID (static or dynamic) is required")
	}

	// Build command based on ACL type
	var cmd string
	switch aclType {
	case ACLTypeIP, ACLTypeExtended:
		cmd = parsers.BuildInterfaceSecureFilterWithDynamicCommand(iface, direction, staticFilterIDs, dynamicFilterIDs)
	case ACLTypeIPv6:
		cmd = parsers.BuildInterfaceIPv6SecureFilterWithDynamicCommand(iface, direction, staticFilterIDs, dynamicFilterIDs)
	default:
		return fmt.Errorf("dynamic filters not supported for ACL type: %s", aclType)
	}

	logger.Debug().
		Str("service", "ACLApplyService").
		Str("operation", "ApplyFiltersToInterfaceWithDynamic").
		Str("interface", iface).
		Str("direction", direction).
		Str("acl_type", string(aclType)).
		Ints("static_filter_ids", staticFilterIDs).
		Ints("dynamic_filter_ids", dynamicFilterIDs).
		Msgf("Applying filters with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to apply filters to interface %s: %w", iface, err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("filters applied but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetAllInterfaceFiltersForType retrieves all interface filter bindings for a specific ACL type
// Returns a map of interface -> direction -> filter IDs
func (s *ACLApplyService) GetAllInterfaceFiltersForType(ctx context.Context, aclType ACLType) (map[string]map[string][]int, error) {
	logger := logging.FromContext(ctx)

	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var cmd string
	var parseFunc func(string) (map[string]map[string][]int, error)

	switch aclType {
	case ACLTypeIP, ACLTypeExtended, ACLTypeIPDynamic:
		cmd = parsers.BuildShowIPFilterCommand()
		parseFunc = parsers.ParseInterfaceSecureFilter
	case ACLTypeIPv6, ACLTypeIPv6Dynamic:
		cmd = parsers.BuildShowIPv6FilterCommand()
		parseFunc = parsers.ParseInterfaceIPv6SecureFilter
	case ACLTypeMAC:
		cmd = parsers.BuildShowInterfaceEthernetFilterCommand()
		parseFunc = parsers.ParseInterfaceEthernetFilter
	default:
		return nil, fmt.Errorf("unsupported ACL type: %s", aclType)
	}

	logger.Debug().
		Str("service", "ACLApplyService").
		Str("operation", "GetAllInterfaceFiltersForType").
		Str("acl_type", string(aclType)).
		Msgf("Getting all interface filters with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get interface filters: %w", err)
	}

	result, err := parseFunc(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse interface filters: %w", err)
	}

	return result, nil
}
