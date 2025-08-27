package client

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// DHCPService handles DHCP-related operations
type DHCPService struct {
	executor Executor
	client   *rtxClient // Reference to the main client for save functionality
}

// NewDHCPService creates a new DHCP service instance
func NewDHCPService(executor Executor, client *rtxClient) *DHCPService {
	return &DHCPService{
		executor: executor,
		client:   client,
	}
}

// CreateBinding creates a new DHCP binding
func (s *DHCPService) CreateBinding(ctx context.Context, binding DHCPBinding) error {
	// Validate input
	if err := validateDHCPBinding(binding); err != nil {
		return fmt.Errorf("invalid binding: %w", err)
	}
	
	// Check context before expensive operations
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	
	// Convert client.DHCPBinding to parsers.DHCPBinding
	parserBinding := parsers.DHCPBinding{
		ScopeID:             binding.ScopeID,
		IPAddress:           binding.IPAddress,
		MACAddress:          binding.MACAddress,
		ClientIdentifier:    binding.ClientIdentifier,
		UseClientIdentifier: binding.UseClientIdentifier,
	}
	
	cmd := parsers.BuildDHCPBindCommand(parserBinding)
	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to create DHCP binding: %w", err)
	}
	
	// Check if there's an error in the output
	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}
	
	// Save configuration after successful creation
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("binding created but failed to save configuration: %w", err)
		}
	}
	
	return nil
}

// DeleteBinding removes a DHCP binding
func (s *DHCPService) DeleteBinding(ctx context.Context, scopeID int, ipAddress string) error {
	cmd := parsers.BuildDHCPUnbindCommand(scopeID, ipAddress)
	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to delete DHCP binding: %w", err)
	}
	
	// Check if there's an error in the output
	if len(output) > 0 && containsError(string(output)) {
		return fmt.Errorf("command failed: %s", string(output))
	}
	
	// Save configuration after successful deletion
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("binding deleted but failed to save configuration: %w", err)
		}
	}
	
	return nil
}

// ListBindings retrieves all DHCP bindings for a scope
func (s *DHCPService) ListBindings(ctx context.Context, scopeID int) ([]DHCPBinding, error) {
	cmd := parsers.BuildShowDHCPBindingsCommand(scopeID)
	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list DHCP bindings: %w", err)
	}
	
	log.Printf("[DEBUG] DHCP bindings raw output for scope %d: %q", scopeID, string(output))
	
	// Parse the output
	parser := parsers.NewDHCPBindingsParser()
	parserBindings, err := parser.ParseBindings(string(output), scopeID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DHCP bindings: %w", err)
	}
	
	// Convert parsers.DHCPBinding to client.DHCPBinding
	bindings := make([]DHCPBinding, len(parserBindings))
	for i, pb := range parserBindings {
		bindings[i] = DHCPBinding{
			ScopeID:             pb.ScopeID,
			IPAddress:           pb.IPAddress,
			MACAddress:          pb.MACAddress,
			ClientIdentifier:    pb.ClientIdentifier,
			UseClientIdentifier: pb.UseClientIdentifier,
		}
	}
	
	return bindings, nil
}

// containsError checks if the output contains an error message
func containsError(output string) bool {
	// More specific patterns for RTX router errors
	errorPatterns := []string{
		"Error:",
		"% Error:",
		"Command failed:",
		"Invalid parameter",
		"Permission denied",
		"Connection timeout",
		"already exists",
		"not found",
	}
	
	outputLower := strings.ToLower(output)
	for _, pattern := range errorPatterns {
		if strings.Contains(outputLower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// validateDHCPBinding validates DHCP binding parameters
func validateDHCPBinding(binding DHCPBinding) error {
	if binding.ScopeID <= 0 {
		return fmt.Errorf("scope_id must be positive")
	}
	
	if net.ParseIP(binding.IPAddress) == nil {
		return fmt.Errorf("invalid IP address: %s", binding.IPAddress)
	}
	
	// Validate either MAC address or client identifier
	if binding.ClientIdentifier != "" {
		// Validate client identifier format
		parts := strings.Split(binding.ClientIdentifier, ":")
		if len(parts) < 2 {
			return fmt.Errorf("client_identifier must be in format 'type:data' (e.g., '01:aa:bb:cc:dd:ee:ff')")
		}
		
		// Validate each part is valid hex
		for i, part := range parts {
			if len(part) != 2 {
				return fmt.Errorf("client_identifier must contain 2-character hex octets at position %d, got %q", i, part)
			}
			
			for _, c := range part {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
					return fmt.Errorf("client_identifier contains invalid hex character '%c' at position %d", c, i)
				}
			}
		}
	} else if binding.MACAddress != "" {
		// Validate MAC address
		_, err := parsers.NormalizeMACAddress(binding.MACAddress)
		if err != nil {
			return fmt.Errorf("invalid MAC address: %w", err)
		}
	} else {
		return fmt.Errorf("either mac_address or client_identifier must be specified")
	}
	
	return nil
}

