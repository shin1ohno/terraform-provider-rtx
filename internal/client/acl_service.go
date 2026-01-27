package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// ACLType represents the type of Access Control List
type ACLType string

const (
	// ACLTypeIP represents IPv4 static filters
	ACLTypeIP ACLType = "ip"
	// ACLTypeIPv6 represents IPv6 static filters
	ACLTypeIPv6 ACLType = "ipv6"
	// ACLTypeMAC represents MAC/Ethernet filters
	ACLTypeMAC ACLType = "mac"
	// ACLTypeIPDynamic represents IPv4 dynamic filters
	ACLTypeIPDynamic ACLType = "ip_dynamic"
	// ACLTypeIPv6Dynamic represents IPv6 dynamic filters
	ACLTypeIPv6Dynamic ACLType = "ipv6_dynamic"
	// ACLTypeExtended represents IPv4 extended access list (alias for ip)
	ACLTypeExtended ACLType = "extended"
)

// ACLEntry represents a generic ACL entry that can be used across different ACL types
type ACLEntry struct {
	Sequence int                    // Sequence number for this entry
	Fields   map[string]interface{} // ACL-type specific fields
}

// ACLService handles unified ACL entry operations across all ACL types
type ACLService struct {
	executor Executor
	client   *rtxClient
}

// NewACLService creates a new ACL service instance
func NewACLService(executor Executor, client *rtxClient) *ACLService {
	return &ACLService{
		executor: executor,
		client:   client,
	}
}

// CreateACLEntries creates multiple ACL entries in batch
// Supports all ACL types: ip, ipv6, mac, extended
func (s *ACLService) CreateACLEntries(ctx context.Context, aclType ACLType, entries []ACLEntry) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("service", "acl").
		Str("acl_type", string(aclType)).
		Int("entry_count", len(entries)).
		Msg("Creating ACL entries")

	for i, entry := range entries {
		cmd, err := s.buildCreateCommand(aclType, entry)
		if err != nil {
			return fmt.Errorf("failed to build command for entry %d (sequence %d): %w", i, entry.Sequence, err)
		}

		logger.Debug().
			Str("service", "acl").
			Str("command", cmd).
			Msg("Executing ACL create command")

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to create ACL entry %d (sequence %d): %w", i, entry.Sequence, err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed for entry %d (sequence %d): %s", i, entry.Sequence, string(output))
		}
	}

	// Save configuration after all entries are created
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("ACL entries created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ReadACLEntries reads all ACL entries of a specific type
// Returns entries filtered by the provided sequence numbers, or all entries if sequences is empty
func (s *ACLService) ReadACLEntries(ctx context.Context, aclType ACLType, sequences []int) ([]ACLEntry, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("service", "acl").
		Str("acl_type", string(aclType)).
		Msg("Reading ACL entries")

	cmd := s.buildShowCommand(aclType)
	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to read ACL entries: %w", err)
	}

	entries, err := s.parseEntries(aclType, string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ACL entries: %w", err)
	}

	// Filter by sequences if provided
	if len(sequences) > 0 {
		sequenceSet := make(map[int]bool)
		for _, seq := range sequences {
			sequenceSet[seq] = true
		}

		filtered := make([]ACLEntry, 0)
		for _, entry := range entries {
			if sequenceSet[entry.Sequence] {
				filtered = append(filtered, entry)
			}
		}
		return filtered, nil
	}

	return entries, nil
}

// UpdateACLEntries updates existing ACL entries
// For RTX routers, this is done by re-creating entries with the same sequence numbers
func (s *ACLService) UpdateACLEntries(ctx context.Context, aclType ACLType, entries []ACLEntry) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("service", "acl").
		Str("acl_type", string(aclType)).
		Int("entry_count", len(entries)).
		Msg("Updating ACL entries")

	// For RTX routers, we can simply re-run the create command to overwrite
	for i, entry := range entries {
		cmd, err := s.buildCreateCommand(aclType, entry)
		if err != nil {
			return fmt.Errorf("failed to build command for entry %d (sequence %d): %w", i, entry.Sequence, err)
		}

		logger.Debug().
			Str("service", "acl").
			Str("command", cmd).
			Msg("Executing ACL update command")

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to update ACL entry %d (sequence %d): %w", i, entry.Sequence, err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed for entry %d (sequence %d): %s", i, entry.Sequence, string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("ACL entries updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// DeleteACLEntries removes multiple ACL entries by sequence numbers
func (s *ACLService) DeleteACLEntries(ctx context.Context, aclType ACLType, sequences []int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("service", "acl").
		Str("acl_type", string(aclType)).
		Interface("sequences", sequences).
		Msg("Deleting ACL entries")

	for _, seq := range sequences {
		cmd := s.buildDeleteCommand(aclType, seq)

		logger.Debug().
			Str("service", "acl").
			Str("command", cmd).
			Msg("Executing ACL delete command")

		output, err := s.executor.Run(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to delete ACL entry (sequence %d): %w", seq, err)
		}

		if len(output) > 0 && containsError(string(output)) {
			// Check if it's already gone (not found)
			if strings.Contains(strings.ToLower(string(output)), "not found") {
				continue
			}
			return fmt.Errorf("command failed for sequence %d: %s", seq, string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("ACL entries deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetAllFilterSequences returns all used sequence numbers for a specific ACL type
// This is used for collision detection before creating new entries
// Returns a map of sequence number to owner identifier (empty string if unknown)
func (s *ACLService) GetAllFilterSequences(ctx context.Context, aclType ACLType) (map[int]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	logger := logging.FromContext(ctx)
	logger.Debug().
		Str("service", "acl").
		Str("acl_type", string(aclType)).
		Msg("Getting all filter sequences")

	entries, err := s.ReadACLEntries(ctx, aclType, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to read entries for collision detection: %w", err)
	}

	result := make(map[int]string)
	for _, entry := range entries {
		// We don't know the owner from router output, so use empty string
		result[entry.Sequence] = ""
	}

	return result, nil
}

// buildCreateCommand builds the CLI command to create an ACL entry
func (s *ACLService) buildCreateCommand(aclType ACLType, entry ACLEntry) (string, error) {
	switch aclType {
	case ACLTypeIP, ACLTypeExtended:
		return s.buildIPFilterCommand(entry)
	case ACLTypeIPv6:
		return s.buildIPv6FilterCommand(entry)
	case ACLTypeMAC:
		return s.buildMACFilterCommand(entry)
	default:
		return "", fmt.Errorf("unsupported ACL type: %s", aclType)
	}
}

// buildIPFilterCommand builds an IP filter command from an ACL entry
func (s *ACLService) buildIPFilterCommand(entry ACLEntry) (string, error) {
	filter := parsers.IPFilter{
		Number: entry.Sequence,
	}

	// Extract fields from the entry
	if v, ok := entry.Fields["action"].(string); ok {
		filter.Action = v
	} else {
		return "", fmt.Errorf("action is required for IP filter")
	}

	if v, ok := entry.Fields["source_address"].(string); ok {
		filter.SourceAddress = v
	} else {
		filter.SourceAddress = "*"
	}

	if v, ok := entry.Fields["source_mask"].(string); ok {
		filter.SourceMask = v
	}

	if v, ok := entry.Fields["dest_address"].(string); ok {
		filter.DestAddress = v
	} else {
		filter.DestAddress = "*"
	}

	if v, ok := entry.Fields["dest_mask"].(string); ok {
		filter.DestMask = v
	}

	if v, ok := entry.Fields["protocol"].(string); ok {
		filter.Protocol = v
	} else {
		filter.Protocol = "*"
	}

	if v, ok := entry.Fields["source_port"].(string); ok {
		filter.SourcePort = v
	}

	if v, ok := entry.Fields["dest_port"].(string); ok {
		filter.DestPort = v
	}

	if v, ok := entry.Fields["established"].(bool); ok {
		filter.Established = v
	}

	return parsers.BuildIPFilterCommand(filter), nil
}

// buildIPv6FilterCommand builds an IPv6 filter command from an ACL entry
func (s *ACLService) buildIPv6FilterCommand(entry ACLEntry) (string, error) {
	filter := parsers.IPFilter{
		Number: entry.Sequence,
	}

	// Extract fields from the entry
	if v, ok := entry.Fields["action"].(string); ok {
		filter.Action = v
	} else {
		return "", fmt.Errorf("action is required for IPv6 filter")
	}

	if v, ok := entry.Fields["source_address"].(string); ok {
		filter.SourceAddress = v
	} else {
		filter.SourceAddress = "*"
	}

	if v, ok := entry.Fields["dest_address"].(string); ok {
		filter.DestAddress = v
	} else {
		filter.DestAddress = "*"
	}

	if v, ok := entry.Fields["protocol"].(string); ok {
		filter.Protocol = v
	} else {
		filter.Protocol = "*"
	}

	if v, ok := entry.Fields["source_port"].(string); ok {
		filter.SourcePort = v
	}

	if v, ok := entry.Fields["dest_port"].(string); ok {
		filter.DestPort = v
	}

	return parsers.BuildIPv6FilterCommand(filter), nil
}

// buildMACFilterCommand builds an Ethernet filter command from an ACL entry
func (s *ACLService) buildMACFilterCommand(entry ACLEntry) (string, error) {
	filter := parsers.EthernetFilter{
		Number: entry.Sequence,
	}

	// Extract fields from the entry
	if v, ok := entry.Fields["action"].(string); ok {
		filter.Action = v
	} else {
		return "", fmt.Errorf("action is required for MAC filter")
	}

	if v, ok := entry.Fields["source_mac"].(string); ok {
		filter.SourceMAC = v
	} else {
		filter.SourceMAC = "*"
	}

	if v, ok := entry.Fields["dest_mac"].(string); ok {
		filter.DestMAC = v
		filter.DestinationMAC = v
	} else {
		filter.DestMAC = "*"
		filter.DestinationMAC = "*"
	}

	if v, ok := entry.Fields["ether_type"].(string); ok {
		filter.EtherType = v
	}

	if v, ok := entry.Fields["vlan_id"].(int); ok {
		filter.VlanID = v
	}

	if v, ok := entry.Fields["dhcp_type"].(string); ok {
		filter.DHCPType = v
	}

	if v, ok := entry.Fields["dhcp_scope"].(int); ok {
		filter.DHCPScope = v
	}

	if v, ok := entry.Fields["offset"].(int); ok {
		filter.Offset = v
	}

	if v, ok := entry.Fields["byte_list"].([]string); ok {
		filter.ByteList = v
	}

	return parsers.BuildEthernetFilterCommand(filter), nil
}

// buildShowCommand builds the CLI command to show ACL entries
func (s *ACLService) buildShowCommand(aclType ACLType) string {
	switch aclType {
	case ACLTypeIP, ACLTypeExtended:
		return parsers.BuildShowIPFilterCommand()
	case ACLTypeIPv6:
		return parsers.BuildShowIPv6FilterCommand()
	case ACLTypeMAC:
		return parsers.BuildShowAllEthernetFiltersCommand()
	default:
		return ""
	}
}

// buildDeleteCommand builds the CLI command to delete an ACL entry
func (s *ACLService) buildDeleteCommand(aclType ACLType, sequence int) string {
	switch aclType {
	case ACLTypeIP, ACLTypeExtended:
		return parsers.BuildDeleteIPFilterCommand(sequence)
	case ACLTypeIPv6:
		return parsers.BuildDeleteIPv6FilterCommand(sequence)
	case ACLTypeMAC:
		return parsers.BuildDeleteEthernetFilterCommand(sequence)
	default:
		return ""
	}
}

// parseEntries parses the CLI output into ACL entries
func (s *ACLService) parseEntries(aclType ACLType, output string) ([]ACLEntry, error) {
	switch aclType {
	case ACLTypeIP, ACLTypeExtended:
		return s.parseIPFilters(output)
	case ACLTypeIPv6:
		return s.parseIPv6Filters(output)
	case ACLTypeMAC:
		return s.parseMACFilters(output)
	default:
		return nil, fmt.Errorf("unsupported ACL type: %s", aclType)
	}
}

// parseIPFilters parses IP filter output into ACL entries
func (s *ACLService) parseIPFilters(output string) ([]ACLEntry, error) {
	filters, err := parsers.ParseIPFilterConfig(output)
	if err != nil {
		return nil, err
	}

	entries := make([]ACLEntry, len(filters))
	for i, f := range filters {
		entries[i] = ACLEntry{
			Sequence: f.Number,
			Fields: map[string]interface{}{
				"action":         f.Action,
				"source_address": f.SourceAddress,
				"source_mask":    f.SourceMask,
				"dest_address":   f.DestAddress,
				"dest_mask":      f.DestMask,
				"protocol":       f.Protocol,
				"source_port":    f.SourcePort,
				"dest_port":      f.DestPort,
				"established":    f.Established,
			},
		}
	}

	return entries, nil
}

// parseIPv6Filters parses IPv6 filter output into ACL entries
func (s *ACLService) parseIPv6Filters(output string) ([]ACLEntry, error) {
	filters, err := parsers.ParseIPv6FilterConfig(output)
	if err != nil {
		return nil, err
	}

	entries := make([]ACLEntry, len(filters))
	for i, f := range filters {
		entries[i] = ACLEntry{
			Sequence: f.Number,
			Fields: map[string]interface{}{
				"action":         f.Action,
				"source_address": f.SourceAddress,
				"dest_address":   f.DestAddress,
				"protocol":       f.Protocol,
				"source_port":    f.SourcePort,
				"dest_port":      f.DestPort,
			},
		}
	}

	return entries, nil
}

// parseMACFilters parses MAC/Ethernet filter output into ACL entries
func (s *ACLService) parseMACFilters(output string) ([]ACLEntry, error) {
	filters, err := parsers.ParseEthernetFilterConfig(output)
	if err != nil {
		return nil, err
	}

	entries := make([]ACLEntry, len(filters))
	for i, f := range filters {
		entries[i] = ACLEntry{
			Sequence: f.Number,
			Fields: map[string]interface{}{
				"action":     f.Action,
				"source_mac": f.SourceMAC,
				"dest_mac":   f.DestinationMAC,
				"ether_type": f.EtherType,
				"vlan_id":    f.VlanID,
				"dhcp_type":  f.DHCPType,
				"dhcp_scope": f.DHCPScope,
				"offset":     f.Offset,
				"byte_list":  f.ByteList,
			},
		}
	}

	return entries, nil
}

// ValidateACLType validates that the ACL type is supported
func ValidateACLType(aclType ACLType) error {
	switch aclType {
	case ACLTypeIP, ACLTypeIPv6, ACLTypeMAC, ACLTypeExtended, ACLTypeIPDynamic, ACLTypeIPv6Dynamic:
		return nil
	default:
		return fmt.Errorf("unsupported ACL type: %s, must be one of: ip, ipv6, mac, extended, ip_dynamic, ipv6_dynamic", aclType)
	}
}
