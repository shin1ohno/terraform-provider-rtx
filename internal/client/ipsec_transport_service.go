package client

import (
	"context"
	"fmt"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// IPsecTransportService handles IPsec transport configuration operations
type IPsecTransportService struct {
	executor Executor
	client   *rtxClient
}

// NewIPsecTransportService creates a new IPsec transport service
func NewIPsecTransportService(executor Executor, client *rtxClient) *IPsecTransportService {
	return &IPsecTransportService{
		executor: executor,
		client:   client,
	}
}

// Get retrieves a specific IPsec transport configuration
func (s *IPsecTransportService) Get(ctx context.Context, transportID int) (*parsers.IPsecTransport, error) {
	transports, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, transport := range transports {
		if transport.TransportID == transportID {
			return &transport, nil
		}
	}

	return nil, fmt.Errorf("IPsec transport %d not found", transportID)
}

// List retrieves all IPsec transport configurations
func (s *IPsecTransportService) List(ctx context.Context) ([]parsers.IPsecTransport, error) {
	output, err := s.executor.Run(ctx, parsers.BuildShowIPsecTransportCommand())
	if err != nil {
		return nil, fmt.Errorf("failed to get IPsec transport config: %w", err)
	}

	parsed, err := parsers.ParseIPsecTransportConfig(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse IPsec transport config: %w", err)
	}

	return parsed, nil
}

// Create creates a new IPsec transport configuration
func (s *IPsecTransportService) Create(ctx context.Context, t parsers.IPsecTransport) error {
	// Validate configuration
	if err := parsers.ValidateIPsecTransport(t); err != nil {
		return fmt.Errorf("invalid IPsec transport config: %w", err)
	}

	cmd := parsers.BuildIPsecTransportCommand(t)
	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to execute IPsec transport command '%s': %w", cmd, err)
	}
	if containsError(string(output)) {
		return fmt.Errorf("IPsec transport command '%s' failed: %s", cmd, string(output))
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save IPsec transport config: %w", err)
	}

	return nil
}

// Update modifies an existing IPsec transport configuration
func (s *IPsecTransportService) Update(ctx context.Context, t parsers.IPsecTransport) error {
	// Validate configuration
	if err := parsers.ValidateIPsecTransport(t); err != nil {
		return fmt.Errorf("invalid IPsec transport config: %w", err)
	}

	// Delete existing and recreate (RTX routers typically require this for updates)
	cmd := parsers.BuildIPsecTransportCommand(t)
	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to execute IPsec transport command '%s': %w", cmd, err)
	}
	if containsError(string(output)) {
		return fmt.Errorf("IPsec transport command '%s' failed: %s", cmd, string(output))
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save IPsec transport config: %w", err)
	}

	return nil
}

// Delete removes an IPsec transport configuration
func (s *IPsecTransportService) Delete(ctx context.Context, transportID int) error {
	cmd := parsers.BuildDeleteIPsecTransportCommand(transportID)
	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to execute IPsec transport delete command '%s': %w", cmd, err)
	}
	if containsError(string(output)) {
		return fmt.Errorf("IPsec transport delete command '%s' failed: %s", cmd, string(output))
	}

	// Save configuration
	if err := s.client.SaveConfig(ctx); err != nil {
		return fmt.Errorf("failed to save config after IPsec transport delete: %w", err)
	}

	return nil
}
