package client

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// BridgeService handles bridge operations
type BridgeService struct {
	executor Executor
	client   *rtxClient // Reference to the main client for save functionality
}

// NewBridgeService creates a new bridge service instance
func NewBridgeService(executor Executor, client *rtxClient) *BridgeService {
	return &BridgeService{
		executor: executor,
		client:   client,
	}
}

// CreateBridge creates a new bridge
func (s *BridgeService) CreateBridge(ctx context.Context, bridge BridgeConfig) error {
	// Convert client.BridgeConfig to parsers.BridgeConfig for validation
	parserBridge := s.toParserBridge(bridge)

	// Validate input
	if err := parsers.ValidateBridge(parserBridge); err != nil {
		return fmt.Errorf("invalid bridge: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Check if bridge already exists
	existingBridge, err := s.GetBridge(ctx, bridge.Name)
	if err == nil && existingBridge != nil {
		return fmt.Errorf("bridge %s already exists", bridge.Name)
	}

	// Build and execute bridge creation command
	cmd := parsers.BuildBridgeMemberCommand(bridge.Name, bridge.Members)
	log.Printf("[DEBUG] Creating bridge with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to create bridge: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("bridge created but failed to save configuration: %w", err)
		}
	}

	return nil
}

// GetBridge retrieves a bridge configuration
func (s *BridgeService) GetBridge(ctx context.Context, name string) (*BridgeConfig, error) {
	// Validate bridge name
	if err := parsers.ValidateBridgeName(name); err != nil {
		return nil, fmt.Errorf("invalid bridge name: %w", err)
	}

	cmd := parsers.BuildShowBridgeCommand(name)
	log.Printf("[DEBUG] Getting bridge with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to get bridge: %w", err)
	}

	log.Printf("[DEBUG] Bridge raw output: %q", string(output))

	parser := parsers.NewBridgeParser()
	parserBridge, err := parser.ParseSingleBridge(string(output), name)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bridge: %w", err)
	}

	// Convert parsers.BridgeConfig to client.BridgeConfig
	bridge := s.fromParserBridge(*parserBridge)
	return &bridge, nil
}

// UpdateBridge updates an existing bridge
func (s *BridgeService) UpdateBridge(ctx context.Context, bridge BridgeConfig) error {
	parserBridge := s.toParserBridge(bridge)

	// Validate input
	if err := parsers.ValidateBridge(parserBridge); err != nil {
		return fmt.Errorf("invalid bridge: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Verify bridge exists
	_, err := s.GetBridge(ctx, bridge.Name)
	if err != nil {
		return fmt.Errorf("bridge %s does not exist: %w", bridge.Name, err)
	}

	// Update by replacing the bridge member command
	// RTX routers replace the entire bridge member list with a new command
	cmd := parsers.BuildBridgeMemberCommand(bridge.Name, bridge.Members)
	log.Printf("[DEBUG] Updating bridge with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to update bridge: %w", err)
	}

	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}

	// Save configuration
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("bridge updated but failed to save configuration: %w", err)
		}
	}

	return nil
}

// DeleteBridge removes a bridge
func (s *BridgeService) DeleteBridge(ctx context.Context, name string) error {
	// Validate bridge name
	if err := parsers.ValidateBridgeName(name); err != nil {
		return fmt.Errorf("invalid bridge name: %w", err)
	}

	// Check context
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Check if bridge exists (optional, but helps with idempotency)
	_, err := s.GetBridge(ctx, name)
	if err != nil {
		// Check if it's already gone
		if strings.Contains(err.Error(), "not found") {
			return nil
		}
		// For other errors, we try to delete anyway
		log.Printf("[DEBUG] Could not verify bridge existence: %v, attempting delete anyway", err)
	}

	cmd := parsers.BuildDeleteBridgeCommand(name)
	log.Printf("[DEBUG] Deleting bridge with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete bridge: %w", err)
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
			return fmt.Errorf("bridge deleted but failed to save configuration: %w", err)
		}
	}

	return nil
}

// ListBridges retrieves all bridges
func (s *BridgeService) ListBridges(ctx context.Context) ([]BridgeConfig, error) {
	cmd := parsers.BuildShowAllBridgesCommand()
	log.Printf("[DEBUG] Listing bridges with command: %s", cmd)

	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list bridges: %w", err)
	}

	log.Printf("[DEBUG] Bridges raw output: %q", string(output))

	parser := parsers.NewBridgeParser()
	parserBridges, err := parser.ParseBridgeConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse bridges: %w", err)
	}

	// Convert parsers.BridgeConfig to client.BridgeConfig
	bridges := make([]BridgeConfig, len(parserBridges))
	for i, pb := range parserBridges {
		bridges[i] = s.fromParserBridge(pb)
	}

	return bridges, nil
}

// toParserBridge converts client.BridgeConfig to parsers.BridgeConfig
func (s *BridgeService) toParserBridge(bridge BridgeConfig) parsers.BridgeConfig {
	return parsers.BridgeConfig{
		Name:    bridge.Name,
		Members: bridge.Members,
	}
}

// fromParserBridge converts parsers.BridgeConfig to client.BridgeConfig
func (s *BridgeService) fromParserBridge(pb parsers.BridgeConfig) BridgeConfig {
	return BridgeConfig{
		Name:    pb.Name,
		Members: pb.Members,
	}
}
