package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// NATStaticService handles NAT static operations
type NATStaticService struct {
	executor Executor
	client   *rtxClient // Reference to the main client for save functionality
}

// NewNATStaticService creates a new NAT static service instance
func NewNATStaticService(executor Executor, client *rtxClient) *NATStaticService {
	return &NATStaticService{
		executor: executor,
		client:   client,
	}
}

// Create creates a new NAT static configuration
func (s *NATStaticService) Create(ctx context.Context, nat NATStatic) error {
	// Convert client.NATStatic to parsers.NATStatic
	parserNAT := s.toParserNATStatic(nat)

	// Validate input
	if err := parsers.ValidateNATStatic(parserNAT); err != nil {
		return fmt.Errorf("invalid NAT static: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Collect all commands
	commands := []string{}

	// Build the NAT descriptor type command
	typeCmd := parsers.BuildNATDescriptorTypeStaticCommand(nat.DescriptorID)
	logging.FromContext(ctx).Debug().Str("service", "nat_static").Msgf("Creating NAT static with command: %s", typeCmd)
	commands = append(commands, typeCmd)

	// Add each NAT mapping entry
	for _, entry := range nat.Entries {
		parserEntry := s.toParserEntry(entry)

		var entryCmd string
		if parsers.IsPortBasedNAT(parserEntry) {
			entryCmd = parsers.BuildNATStaticPortMappingCommand(nat.DescriptorID, parserEntry)
		} else {
			entryCmd = parsers.BuildNATStaticMappingCommand(nat.DescriptorID, parserEntry)
		}

		logging.FromContext(ctx).Debug().Str("service", "nat_static").Msgf("Adding NAT static entry with command: %s", entryCmd)
		commands = append(commands, entryCmd)
	}

	// Execute all commands in batch
	output, err := s.executor.RunBatch(ctx, commands)
	if err != nil {
		return fmt.Errorf("failed to create NAT static: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("NAT static created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// Get retrieves a NAT static configuration
func (s *NATStaticService) Get(ctx context.Context, descriptorID int) (*NATStatic, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cmd := parsers.BuildShowNATStaticCommand(descriptorID)
	logging.FromContext(ctx).Debug().Str("service", "nat_static").Msgf("Getting NAT static with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get NAT static: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "nat_static").Msgf("NAT static raw output: %q", string(output))

	parser := parsers.NewNATStaticParser()
	parserNAT, err := parser.ParseSingleNATStatic(string(output), descriptorID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse NAT static: %w", err)
	}

	// Convert parsers.NATStatic to client.NATStatic
	nat := s.fromParserNATStatic(*parserNAT)
	return &nat, nil
}

// Update updates an existing NAT static configuration
func (s *NATStaticService) Update(ctx context.Context, nat NATStatic) error {
	parserNAT := s.toParserNATStatic(nat)

	// Validate input
	if err := parsers.ValidateNATStatic(parserNAT); err != nil {
		return fmt.Errorf("invalid NAT static: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get current NAT static configuration
	currentNAT, err := s.Get(ctx, nat.DescriptorID)
	if err != nil {
		return fmt.Errorf("failed to get current NAT static: %w", err)
	}

	// Collect all commands
	commands := []string{}

	// Delete old entries that are not in the new configuration
	for _, oldEntry := range currentNAT.Entries {
		found := false
		for _, newEntry := range nat.Entries {
			if s.entriesEqual(oldEntry, newEntry) {
				found = true
				break
			}
		}
		if !found {
			parserOldEntry := s.toParserEntry(oldEntry)
			var deleteCmd string
			if parsers.IsPortBasedNAT(parserOldEntry) {
				deleteCmd = parsers.BuildDeleteNATStaticPortMappingCommand(nat.DescriptorID, parserOldEntry)
			} else {
				deleteCmd = parsers.BuildDeleteNATStaticMappingCommand(nat.DescriptorID, parserOldEntry)
			}
			logging.FromContext(ctx).Debug().Str("service", "nat_static").Msgf("Removing NAT static entry with command: %s", deleteCmd)
			commands = append(commands, deleteCmd)
		}
	}

	// Add new entries that are not in the old configuration
	for _, newEntry := range nat.Entries {
		found := false
		for _, oldEntry := range currentNAT.Entries {
			if s.entriesEqual(oldEntry, newEntry) {
				found = true
				break
			}
		}
		if !found {
			parserNewEntry := s.toParserEntry(newEntry)
			var entryCmd string
			if parsers.IsPortBasedNAT(parserNewEntry) {
				entryCmd = parsers.BuildNATStaticPortMappingCommand(nat.DescriptorID, parserNewEntry)
			} else {
				entryCmd = parsers.BuildNATStaticMappingCommand(nat.DescriptorID, parserNewEntry)
			}
			logging.FromContext(ctx).Debug().Str("service", "nat_static").Msgf("Adding NAT static entry with command: %s", entryCmd)
			commands = append(commands, entryCmd)
		}
	}

	// Execute all commands in batch
	if len(commands) > 0 {
		output, err := s.executor.RunBatch(ctx, commands)
		if err != nil {
			return fmt.Errorf("failed to update NAT static: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("command failed: %s", string(output))
		}
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("NAT static updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// Delete removes a NAT static configuration
func (s *NATStaticService) Delete(ctx context.Context, descriptorID int) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cmd := parsers.BuildDeleteNATStaticCommand(descriptorID)
	logging.FromContext(ctx).Debug().Str("service", "nat_static").Msgf("Deleting NAT static with command: %s", cmd)

	// Execute command in batch
	output, err := s.executor.RunBatch(ctx, []string{cmd})
	if err != nil {
		return fmt.Errorf("failed to delete NAT static: %w", err)
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
			return fmt.Errorf("NAT static deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// List retrieves all NAT static configurations
func (s *NATStaticService) List(ctx context.Context) ([]NATStatic, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cmd := parsers.BuildShowAllNATStaticCommand()
	logging.FromContext(ctx).Debug().Str("service", "nat_static").Msgf("Listing NAT statics with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list NAT statics: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "nat_static").Msgf("NAT statics raw output: %q", string(output))

	parserNATs, err := parsers.ParseNATStaticConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse NAT statics: %w", err)
	}

	// Convert parsers.NATStatic to client.NATStatic
	nats := make([]NATStatic, len(parserNATs))
	for i, pn := range parserNATs {
		nats[i] = s.fromParserNATStatic(pn)
	}

	return nats, nil
}

// toParserNATStatic converts client.NATStatic to parsers.NATStatic
func (s *NATStaticService) toParserNATStatic(nat NATStatic) parsers.NATStatic {
	entries := make([]parsers.NATStaticEntry, len(nat.Entries))
	for i, e := range nat.Entries {
		entries[i] = s.toParserEntry(e)
	}

	return parsers.NATStatic{
		DescriptorID: nat.DescriptorID,
		Entries:      entries,
	}
}

// fromParserNATStatic converts parsers.NATStatic to client.NATStatic
func (s *NATStaticService) fromParserNATStatic(pn parsers.NATStatic) NATStatic {
	entries := make([]NATStaticEntry, len(pn.Entries))
	for i, pe := range pn.Entries {
		entries[i] = s.fromParserEntry(pe)
	}

	return NATStatic{
		DescriptorID: pn.DescriptorID,
		Entries:      entries,
	}
}

// toParserEntry converts client.NATStaticEntry to parsers.NATStaticEntry
func (s *NATStaticService) toParserEntry(entry NATStaticEntry) parsers.NATStaticEntry {
	pe := parsers.NATStaticEntry{
		InsideLocal:   entry.InsideLocal,
		OutsideGlobal: entry.OutsideGlobal,
		Protocol:      entry.Protocol,
	}
	if entry.InsideLocalPort != nil {
		pe.InsideLocalPort = *entry.InsideLocalPort
	}
	if entry.OutsideGlobalPort != nil {
		pe.OutsideGlobalPort = *entry.OutsideGlobalPort
	}
	return pe
}

// fromParserEntry converts parsers.NATStaticEntry to client.NATStaticEntry
func (s *NATStaticService) fromParserEntry(pe parsers.NATStaticEntry) NATStaticEntry {
	entry := NATStaticEntry{
		InsideLocal:   pe.InsideLocal,
		OutsideGlobal: pe.OutsideGlobal,
		Protocol:      pe.Protocol,
	}
	if pe.InsideLocalPort != 0 {
		port := pe.InsideLocalPort
		entry.InsideLocalPort = &port
	}
	if pe.OutsideGlobalPort != 0 {
		port := pe.OutsideGlobalPort
		entry.OutsideGlobalPort = &port
	}
	return entry
}

// entriesEqual checks if two NATStaticEntry instances are equal
func (s *NATStaticService) entriesEqual(a, b NATStaticEntry) bool {
	if a.InsideLocal != b.InsideLocal || a.OutsideGlobal != b.OutsideGlobal {
		return false
	}
	if !strings.EqualFold(a.Protocol, b.Protocol) {
		return false
	}
	// Compare pointer values
	if !intPtrEqual(a.InsideLocalPort, b.InsideLocalPort) {
		return false
	}
	if !intPtrEqual(a.OutsideGlobalPort, b.OutsideGlobalPort) {
		return false
	}
	return true
}

// intPtrEqual compares two *int values for equality
func intPtrEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
