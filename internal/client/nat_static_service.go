package client

import (
	"context"
	"fmt"
	"log"
	"strings"

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

	// Build and execute the NAT descriptor type command
	typeCmd := parsers.BuildNATDescriptorTypeStaticCommand(nat.DescriptorID)
	log.Printf("[DEBUG] Creating NAT static with command: %s", typeCmd)

	output, err := s.executor.Run(ctx, typeCmd)
	if err != nil {
		return fmt.Errorf("failed to create NAT static: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Add each NAT mapping entry
	for _, entry := range nat.Entries {
		parserEntry := s.toParserEntry(entry)

		var entryCmd string
		if parsers.IsPortBasedNAT(parserEntry) {
			entryCmd = parsers.BuildNATStaticPortMappingCommand(nat.DescriptorID, parserEntry)
		} else {
			entryCmd = parsers.BuildNATStaticMappingCommand(nat.DescriptorID, parserEntry)
		}

		log.Printf("[DEBUG] Adding NAT static entry with command: %s", entryCmd)

		output, err = s.executor.Run(ctx, entryCmd)
		if err != nil {
			return fmt.Errorf("failed to add NAT static entry: %w", err)
		}

		if len(output) > 0 && containsError(string(output)) {
			return fmt.Errorf("entry command failed: %s", string(output))
		}
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
	log.Printf("[DEBUG] Getting NAT static with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get NAT static: %w", err)
	}

	log.Printf("[DEBUG] NAT static raw output: %q", string(output))

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
			log.Printf("[DEBUG] Removing NAT static entry with command: %s", deleteCmd)
			_, _ = s.executor.Run(ctx, deleteCmd) // Ignore errors for cleanup
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
			log.Printf("[DEBUG] Adding NAT static entry with command: %s", entryCmd)

			output, err := s.executor.Run(ctx, entryCmd)
			if err != nil {
				return fmt.Errorf("failed to add NAT static entry: %w", err)
			}

			if len(output) > 0 && containsError(string(output)) {
				return fmt.Errorf("entry command failed: %s", string(output))
			}
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
	log.Printf("[DEBUG] Deleting NAT static with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
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
	log.Printf("[DEBUG] Listing NAT statics with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list NAT statics: %w", err)
	}

	log.Printf("[DEBUG] NAT statics raw output: %q", string(output))

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
	return parsers.NATStaticEntry{
		InsideLocal:       entry.InsideLocal,
		InsideLocalPort:   entry.InsideLocalPort,
		OutsideGlobal:     entry.OutsideGlobal,
		OutsideGlobalPort: entry.OutsideGlobalPort,
		Protocol:          entry.Protocol,
	}
}

// fromParserEntry converts parsers.NATStaticEntry to client.NATStaticEntry
func (s *NATStaticService) fromParserEntry(pe parsers.NATStaticEntry) NATStaticEntry {
	return NATStaticEntry{
		InsideLocal:       pe.InsideLocal,
		InsideLocalPort:   pe.InsideLocalPort,
		OutsideGlobal:     pe.OutsideGlobal,
		OutsideGlobalPort: pe.OutsideGlobalPort,
		Protocol:          pe.Protocol,
	}
}

// entriesEqual checks if two NATStaticEntry instances are equal
func (s *NATStaticService) entriesEqual(a, b NATStaticEntry) bool {
	return a.InsideLocal == b.InsideLocal &&
		a.InsideLocalPort == b.InsideLocalPort &&
		a.OutsideGlobal == b.OutsideGlobal &&
		a.OutsideGlobalPort == b.OutsideGlobalPort &&
		strings.EqualFold(a.Protocol, b.Protocol)
}
