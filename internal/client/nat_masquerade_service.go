package client

import (
	"context"
	"fmt"

	"github.com/sh1/terraform-provider-rtx/internal/logging"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// NATMasqueradeService handles NAT masquerade operations
type NATMasqueradeService struct {
	executor Executor
	client   *rtxClient // Reference to the main client for save functionality
}

// NewNATMasqueradeService creates a new NAT masquerade service instance
func NewNATMasqueradeService(executor Executor, client *rtxClient) *NATMasqueradeService {
	return &NATMasqueradeService{
		executor: executor,
		client:   client,
	}
}

// Create creates a new NAT masquerade configuration
func (s *NATMasqueradeService) Create(ctx context.Context, nat NATMasquerade) error {
	// Convert client.NATMasquerade to parsers.NATMasquerade
	parserNAT := s.toParserNAT(nat)

	// Validate input
	if err := parsers.ValidateNATMasquerade(parserNAT); err != nil {
		return fmt.Errorf("invalid NAT masquerade: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Collect all commands
	commands := []string{}

	// Step 1: Set NAT descriptor type to masquerade
	cmd := parsers.BuildNATDescriptorTypeMasqueradeCommand(nat.DescriptorID)
	logging.FromContext(ctx).Debug().Str("service", "nat_masquerade").Msgf("Creating NAT masquerade with command: %s", cmd)
	commands = append(commands, cmd)

	// Step 2: Set outer address
	cmd = parsers.BuildNATDescriptorAddressOuterCommand(nat.DescriptorID, nat.OuterAddress)
	logging.FromContext(ctx).Debug().Str("service", "nat_masquerade").Msgf("Setting outer address with command: %s", cmd)
	commands = append(commands, cmd)

	// Step 3: Set inner network
	cmd = parsers.BuildNATDescriptorAddressInnerCommand(nat.DescriptorID, nat.InnerNetwork)
	logging.FromContext(ctx).Debug().Str("service", "nat_masquerade").Msgf("Setting inner network with command: %s", cmd)
	commands = append(commands, cmd)

	// Step 4: Configure static entries
	for i, entry := range nat.StaticEntries {
		parserEntry := parsers.MasqueradeStaticEntry{
			EntryNumber:       entry.EntryNumber,
			InsideLocal:       entry.InsideLocal,
			InsideLocalPort:   entry.InsideLocalPort,
			OutsideGlobal:     entry.OutsideGlobal,
			OutsideGlobalPort: entry.OutsideGlobalPort,
			Protocol:          entry.Protocol,
		}
		cmd = parsers.BuildNATMasqueradeStaticCommand(nat.DescriptorID, entry.EntryNumber, parserEntry)
		logging.FromContext(ctx).Debug().Str("service", "nat_masquerade").Msgf("Adding static entry %d with command: %s", i+1, cmd)
		commands = append(commands, cmd)
	}

	// Execute all commands in batch
	if err := runBatchCommands(ctx, s.executor, commands); err != nil {
		return fmt.Errorf("failed to create NAT masquerade: %w", err)
	}

	return saveConfig(ctx, s.client, "NAT masquerade created")
}

// Get retrieves a NAT masquerade configuration by descriptor ID
func (s *NATMasqueradeService) Get(ctx context.Context, descriptorID int) (*NATMasquerade, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cmd := parsers.BuildShowNATDescriptorCommand(descriptorID)
	logging.FromContext(ctx).Debug().Str("service", "nat_masquerade").Msgf("Getting NAT masquerade with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get NAT masquerade: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "nat_masquerade").Msgf("NAT masquerade raw output: %q", string(output))

	parserNATs, err := parsers.ParseNATMasqueradeConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse NAT masquerade: %w", err)
	}

	// Find the matching descriptor
	for _, parserNAT := range parserNATs {
		if parserNAT.DescriptorID == descriptorID {
			nat := s.fromParserNAT(parserNAT)
			return &nat, nil
		}
	}

	return nil, fmt.Errorf("NAT masquerade with descriptor ID %d not found", descriptorID)
}

// Update updates an existing NAT masquerade configuration
func (s *NATMasqueradeService) Update(ctx context.Context, nat NATMasquerade) error {
	// Convert client.NATMasquerade to parsers.NATMasquerade
	parserNAT := s.toParserNAT(nat)

	// Validate input
	if err := parsers.ValidateNATMasquerade(parserNAT); err != nil {
		return fmt.Errorf("invalid NAT masquerade: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Get current configuration
	currentNAT, err := s.Get(ctx, nat.DescriptorID)
	if err != nil {
		return fmt.Errorf("failed to get current NAT masquerade: %w", err)
	}

	// Collect all commands
	commands := []string{}

	// Update outer address if changed
	if currentNAT.OuterAddress != nat.OuterAddress {
		cmd := parsers.BuildNATDescriptorAddressOuterCommand(nat.DescriptorID, nat.OuterAddress)
		logging.FromContext(ctx).Debug().Str("service", "nat_masquerade").Msgf("Updating outer address with command: %s", cmd)
		commands = append(commands, cmd)
	}

	// Update inner network if changed
	if currentNAT.InnerNetwork != nat.InnerNetwork {
		cmd := parsers.BuildNATDescriptorAddressInnerCommand(nat.DescriptorID, nat.InnerNetwork)
		logging.FromContext(ctx).Debug().Str("service", "nat_masquerade").Msgf("Updating inner network with command: %s", cmd)
		commands = append(commands, cmd)
	}

	// Handle static entries: remove old entries that are not in new configuration
	for _, oldEntry := range currentNAT.StaticEntries {
		found := false
		for _, newEntry := range nat.StaticEntries {
			if oldEntry.EntryNumber == newEntry.EntryNumber {
				found = true
				break
			}
		}
		if !found {
			cmd := parsers.BuildDeleteNATMasqueradeStaticCommand(nat.DescriptorID, oldEntry.EntryNumber)
			logging.FromContext(ctx).Debug().Str("service", "nat_masquerade").Msgf("Removing static entry with command: %s", cmd)
			commands = append(commands, cmd)
		}
	}

	// Add/update new entries
	for i, entry := range nat.StaticEntries {
		parserEntry := parsers.MasqueradeStaticEntry{
			EntryNumber:       entry.EntryNumber,
			InsideLocal:       entry.InsideLocal,
			InsideLocalPort:   entry.InsideLocalPort,
			OutsideGlobal:     entry.OutsideGlobal,
			OutsideGlobalPort: entry.OutsideGlobalPort,
			Protocol:          entry.Protocol,
		}
		cmd := parsers.BuildNATMasqueradeStaticCommand(nat.DescriptorID, entry.EntryNumber, parserEntry)
		logging.FromContext(ctx).Debug().Str("service", "nat_masquerade").Msgf("Setting static entry %d with command: %s", i+1, cmd)
		commands = append(commands, cmd)
	}

	// Execute all commands in batch
	if err := runBatchCommands(ctx, s.executor, commands); err != nil {
		return fmt.Errorf("failed to update NAT masquerade: %w", err)
	}

	return saveConfig(ctx, s.client, "NAT masquerade updated")
}

// Delete removes a NAT masquerade configuration
func (s *NATMasqueradeService) Delete(ctx context.Context, descriptorID int) error {
	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cmd := parsers.BuildDeleteNATMasqueradeCommand(descriptorID)
	logging.FromContext(ctx).Debug().Str("service", "nat_masquerade").Msgf("Deleting NAT masquerade with command: %s", cmd)

	// Execute command in batch
	output, err := s.executor.RunBatch(ctx, []string{cmd})
	if err != nil {
		return fmt.Errorf("failed to delete NAT masquerade: %w", err)
	}

	if err := checkOutputErrorIgnoringNotFound(output, "failed to delete NAT masquerade"); err != nil {
		return err
	}

	return saveConfig(ctx, s.client, "NAT masquerade deleted")
}

// List retrieves all NAT masquerade configurations
func (s *NATMasqueradeService) List(ctx context.Context) ([]NATMasquerade, error) {
	// Check context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	cmd := parsers.BuildShowAllNATDescriptorsCommand()
	logging.FromContext(ctx).Debug().Str("service", "nat_masquerade").Msgf("Listing NAT masquerades with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list NAT masquerades: %w", err)
	}

	logging.FromContext(ctx).Debug().Str("service", "nat_masquerade").Msgf("NAT masquerades raw output: %q", string(output))

	parserNATs, err := parsers.ParseNATMasqueradeConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse NAT masquerades: %w", err)
	}

	// Convert parsers.NATMasquerade to client.NATMasquerade
	nats := make([]NATMasquerade, len(parserNATs))
	for i, parserNAT := range parserNATs {
		nats[i] = s.fromParserNAT(parserNAT)
	}

	return nats, nil
}

// toParserNAT converts client.NATMasquerade to parsers.NATMasquerade
func (s *NATMasqueradeService) toParserNAT(nat NATMasquerade) parsers.NATMasquerade {
	staticEntries := make([]parsers.MasqueradeStaticEntry, len(nat.StaticEntries))
	for i, entry := range nat.StaticEntries {
		staticEntries[i] = parsers.MasqueradeStaticEntry{
			EntryNumber:       entry.EntryNumber,
			InsideLocal:       entry.InsideLocal,
			InsideLocalPort:   entry.InsideLocalPort,
			OutsideGlobal:     entry.OutsideGlobal,
			OutsideGlobalPort: entry.OutsideGlobalPort,
			Protocol:          entry.Protocol,
		}
	}

	return parsers.NATMasquerade{
		DescriptorID:  nat.DescriptorID,
		OuterAddress:  nat.OuterAddress,
		InnerNetwork:  nat.InnerNetwork,
		StaticEntries: staticEntries,
	}
}

// fromParserNAT converts parsers.NATMasquerade to client.NATMasquerade
func (s *NATMasqueradeService) fromParserNAT(parserNAT parsers.NATMasquerade) NATMasquerade {
	staticEntries := make([]MasqueradeStaticEntry, len(parserNAT.StaticEntries))
	for i, entry := range parserNAT.StaticEntries {
		staticEntries[i] = MasqueradeStaticEntry{
			EntryNumber:       entry.EntryNumber,
			InsideLocal:       entry.InsideLocal,
			InsideLocalPort:   entry.InsideLocalPort,
			OutsideGlobal:     entry.OutsideGlobal,
			OutsideGlobalPort: entry.OutsideGlobalPort,
			Protocol:          entry.Protocol,
		}
	}

	return NATMasquerade{
		DescriptorID:  parserNAT.DescriptorID,
		OuterAddress:  parserNAT.OuterAddress,
		InnerNetwork:  parserNAT.InnerNetwork,
		StaticEntries: staticEntries,
	}
}
