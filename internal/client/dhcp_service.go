package client

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

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

// CreateScope creates a new DHCP scope
func (s *DHCPService) CreateScope(ctx context.Context, scope DHCPScope) error {
	// Validate input
	if err := validateDHCPScope(scope); err != nil {
		return fmt.Errorf("invalid scope: %w", err)
	}
	
	// Check context before expensive operations
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	
	// Convert client.DHCPScope to parsers.DhcpScope
	parserScope := parsers.DhcpScope{
		ID:          scope.ID,
		RangeStart:  scope.RangeStart,
		RangeEnd:    scope.RangeEnd,
		Prefix:      scope.Prefix,
		Gateway:     scope.Gateway,
		DNSServers:  scope.DNSServers,
		Lease:       scope.Lease,
		DomainName:  scope.DomainName,
	}
	
	// Build command with validation
	cmd, err := parsers.BuildDHCPScopeCreateCommandWithValidation(parserScope)
	if err != nil {
		return fmt.Errorf("invalid scope configuration: %w", err)
	}
	
	// Execute command with retry for conflict scenarios
	retryStrategy := NewExponentialBackoff()
	
	for attempt := 0; ; attempt++ {
		output, err := s.executor.Run(ctx, cmd)
		
		// If command succeeded, check output for errors
		if err == nil {
			if len(output) > 0 && containsError(string(output)) {
				outputStr := string(output)
				// Check for conflict scenarios that should be retried
				if strings.Contains(strings.ToLower(outputStr), "already exists") ||
				   strings.Contains(strings.ToLower(outputStr), "conflict") ||
				   strings.Contains(strings.ToLower(outputStr), "busy") {
					
					delay, giveUp := retryStrategy.Next(attempt)
					if giveUp {
						return fmt.Errorf("command failed after %d attempts: %s", attempt+1, outputStr)
					}
					
					log.Printf("[DEBUG] DHCP scope creation attempt %d failed with conflict, retrying in %v: %s", 
						attempt+1, delay, outputStr)
					
					select {
					case <-time.After(delay):
						continue
					case <-ctx.Done():
						return ctx.Err()
					}
				}
				return fmt.Errorf("command failed: %s", outputStr)
			}
			break
		}
		
		// If execution failed, check if it's retryable
		if !IsRetryable(err) {
			return fmt.Errorf("failed to create DHCP scope: %w", err)
		}
		
		delay, giveUp := retryStrategy.Next(attempt)
		if giveUp {
			return fmt.Errorf("failed to create DHCP scope after %d attempts: %w", attempt+1, err)
		}
		
		log.Printf("[DEBUG] DHCP scope creation attempt %d failed, retrying in %v: %v", 
			attempt+1, delay, err)
		
		select {
		case <-time.After(delay):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	// Save configuration after successful creation
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("scope created but failed to save configuration: %w", err)
		}
	}
	
	// Wait for the scope to be available
	if err := s.waitForScopeState(ctx, scope.ID, true); err != nil {
		return fmt.Errorf("scope created but failed state verification: %w", err)
	}
	
	return nil
}

// UpdateScope updates an existing DHCP scope
func (s *DHCPService) UpdateScope(ctx context.Context, scope DHCPScope) error {
	// Validate input
	if err := validateDHCPScope(scope); err != nil {
		return fmt.Errorf("invalid scope: %w", err)
	}
	
	// Check context before expensive operations
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	
	// Convert client.DHCPScope to parsers.DhcpScope
	parserScope := parsers.DhcpScope{
		ID:          scope.ID,
		RangeStart:  scope.RangeStart,
		RangeEnd:    scope.RangeEnd,
		Prefix:      scope.Prefix,
		Gateway:     scope.Gateway,
		DNSServers:  scope.DNSServers,
		Lease:       scope.Lease,
		DomainName:  scope.DomainName,
	}
	
	// Build update commands (delete + create)
	updateCmds, err := parsers.BuildDHCPScopeUpdateCommand(parserScope)
	if err != nil {
		return fmt.Errorf("invalid scope configuration: %w", err)
	}
	
	// Execute commands with retry for conflict scenarios
	retryStrategy := NewExponentialBackoff()
	
	for attempt := 0; ; attempt++ {
		// Execute all commands in sequence
		allSuccess := true
		var lastError error
		
		for i, cmdStr := range updateCmds {
			output, err := s.executor.Run(ctx, cmdStr)
			
			// If command failed to execute
			if err != nil {
				lastError = err
				allSuccess = false
				break
			}
			
			// Check output for errors
			if len(output) > 0 && containsError(string(output)) {
				outputStr := string(output)
				
				// For delete command failures, check if scope doesn't exist (acceptable for update)
				if i == 0 && (strings.Contains(strings.ToLower(outputStr), "not found") ||
					strings.Contains(strings.ToLower(outputStr), "does not exist")) {
					// Scope doesn't exist for deletion, that's fine for update - continue with creation
					log.Printf("[DEBUG] DHCP scope %d doesn't exist for deletion during update, continuing with creation", scope.ID)
					continue
				}
				
				lastError = fmt.Errorf("command %d failed: %s", i+1, outputStr)
				allSuccess = false
				
				// Check for conflict scenarios that should be retried
				if strings.Contains(strings.ToLower(outputStr), "already exists") ||
				   strings.Contains(strings.ToLower(outputStr), "conflict") ||
				   strings.Contains(strings.ToLower(outputStr), "busy") {
					break // Will retry the whole sequence
				}
				
				// Non-retryable error
				return fmt.Errorf("scope update failed: %w", lastError)
			}
		}
		
		// If all commands succeeded, break
		if allSuccess {
			break
		}
		
		// If execution failed, check if it's retryable
		if lastError != nil && !IsRetryable(lastError) {
			return fmt.Errorf("failed to update DHCP scope: %w", lastError)
		}
		
		delay, giveUp := retryStrategy.Next(attempt)
		if giveUp {
			return fmt.Errorf("failed to update DHCP scope after %d attempts: %w", attempt+1, lastError)
		}
		
		log.Printf("[DEBUG] DHCP scope update attempt %d failed, retrying in %v: %v", 
			attempt+1, delay, lastError)
		
		select {
		case <-time.After(delay):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	// Save configuration after successful update
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("scope updated but failed to save configuration: %w", err)
		}
	}
	
	// Wait for the updated scope to be available
	if err := s.waitForScopeState(ctx, scope.ID, true); err != nil {
		return fmt.Errorf("scope updated but failed state verification: %w", err)
	}
	
	return nil
}

// DeleteScope removes a DHCP scope
func (s *DHCPService) DeleteScope(ctx context.Context, scopeID int) error {
	if scopeID <= 0 || scopeID > 255 {
		return fmt.Errorf("scope_id must be between 1 and 255")
	}
	
	// Check context before expensive operations
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	
	// Build command to delete DHCP scope
	cmd := parsers.BuildDHCPScopeDeleteCommand(scopeID)
	
	// Execute command with retry for transient failures
	retryStrategy := NewExponentialBackoff()
	
	for attempt := 0; ; attempt++ {
		output, err := s.executor.Run(ctx, cmd)
		
		// If command succeeded, check output for errors
		if err == nil {
			if len(output) > 0 && containsError(string(output)) {
				outputStr := string(output)
				// Check for transient scenarios that should be retried
				if strings.Contains(strings.ToLower(outputStr), "busy") ||
				   strings.Contains(strings.ToLower(outputStr), "timeout") {
					
					delay, giveUp := retryStrategy.Next(attempt)
					if giveUp {
						return fmt.Errorf("command failed after %d attempts: %s", attempt+1, outputStr)
					}
					
					log.Printf("[DEBUG] DHCP scope deletion attempt %d failed with transient error, retrying in %v: %s", 
						attempt+1, delay, outputStr)
					
					select {
					case <-time.After(delay):
						continue
					case <-ctx.Done():
						return ctx.Err()
					}
				}
				return fmt.Errorf("command failed: %s", outputStr)
			}
			break
		}
		
		// If execution failed, check if it's retryable
		if !IsRetryable(err) {
			return fmt.Errorf("failed to delete DHCP scope: %w", err)
		}
		
		delay, giveUp := retryStrategy.Next(attempt)
		if giveUp {
			return fmt.Errorf("failed to delete DHCP scope after %d attempts: %w", attempt+1, err)
		}
		
		log.Printf("[DEBUG] DHCP scope deletion attempt %d failed, retrying in %v: %v", 
			attempt+1, delay, err)
		
		select {
		case <-time.After(delay):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	// Save configuration after successful deletion
	if s.client != nil {
		if err := s.client.SaveConfig(ctx); err != nil {
			return fmt.Errorf("scope deleted but failed to save configuration: %w", err)
		}
	}
	
	// Wait for the scope to be deleted
	if err := s.waitForScopeState(ctx, scopeID, false); err != nil {
		return fmt.Errorf("scope deleted but failed state verification: %w", err)
	}
	
	return nil
}

// waitForScopeState waits for a DHCP scope to reach the expected state after creation/update/deletion
func (s *DHCPService) waitForScopeState(ctx context.Context, scopeID int, expectExists bool) error {
	retryStrategy := NewLinearBackoff(500*time.Millisecond, 10) // 5 seconds total
	
	for attempt := 0; ; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		// Check current state
		_, err := s.getScope(ctx, scopeID)
		
		if expectExists {
			// We expect the scope to exist
			if err == nil {
				// Scope exists, success
				return nil
			}
			if !IsNotFoundError(err) {
				// Unexpected error, not a "not found" error
				return fmt.Errorf("unexpected error while checking scope state: %w", err)
			}
			// Scope doesn't exist yet, continue waiting
		} else {
			// We expect the scope to not exist
			if IsNotFoundError(err) {
				// Scope doesn't exist, success
				return nil
			}
			if err != nil {
				// Unexpected error
				return fmt.Errorf("unexpected error while checking scope state: %w", err)
			}
			// Scope still exists, continue waiting
		}
		
		// Check if we should retry
		delay, giveUp := retryStrategy.Next(attempt)
		if giveUp {
			if expectExists {
				return fmt.Errorf("timeout waiting for DHCP scope %d to be created", scopeID)
			} else {
				return fmt.Errorf("timeout waiting for DHCP scope %d to be deleted", scopeID)
			}
		}
		
		// Wait before next attempt
		select {
		case <-time.After(delay):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// getScope retrieves a specific DHCP scope (internal helper)
func (s *DHCPService) getScope(ctx context.Context, scopeID int) (*DHCPScope, error) {
	// Use the existing client interface to get the scope
	if s.client != nil {
		return s.client.GetDHCPScope(ctx, scopeID)
	}
	
	// Fallback: use executor directly if client is not available
	cmd := parsers.BuildShowDHCPScopesCommand()
	output, err := s.executor.Run(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list DHCP scopes: %w", err)
	}
	
	// Parse the output line by line to find our specific scope
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Parse each DHCP scope line
		scope, err := parsers.ParseDhcpScope(line)
		if err != nil {
			continue // Skip invalid lines
		}
		
		// Check if this is the scope we're looking for
		if scope.ID == scopeID {
			// Convert parsers.DhcpScope to client.DHCPScope
			clientScope := &DHCPScope{
				ID:          scope.ID,
				RangeStart:  scope.RangeStart,
				RangeEnd:    scope.RangeEnd,
				Prefix:      scope.Prefix,
				Gateway:     scope.Gateway,
				DNSServers:  scope.DNSServers,
				Lease:       scope.Lease,
				DomainName:  scope.DomainName,
			}
			return clientScope, nil
		}
	}
	
	return nil, ErrNotFound
}

// IsNotFoundError checks if an error indicates that a resource was not found
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for specific error types
	if err == ErrNotFound {
		return true
	}
	
	// Check for error messages that indicate "not found"
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "not found") ||
		   strings.Contains(errMsg, "does not exist") ||
		   strings.Contains(errMsg, "no such")
}

// validateDHCPScope validates DHCP scope parameters
func validateDHCPScope(scope DHCPScope) error {
	if scope.ID <= 0 || scope.ID > 255 {
		return fmt.Errorf("scope_id must be between 1 and 255")
	}
	
	if net.ParseIP(scope.RangeStart) == nil {
		return fmt.Errorf("invalid range_start IP address: %s", scope.RangeStart)
	}
	
	if net.ParseIP(scope.RangeEnd) == nil {
		return fmt.Errorf("invalid range_end IP address: %s", scope.RangeEnd)
	}
	
	if scope.Prefix < 8 || scope.Prefix > 32 {
		return fmt.Errorf("prefix must be between 8 and 32")
	}
	
	if scope.Gateway != "" && net.ParseIP(scope.Gateway) == nil {
		return fmt.Errorf("invalid gateway IP address: %s", scope.Gateway)
	}
	
	for i, dns := range scope.DNSServers {
		if net.ParseIP(dns) == nil {
			return fmt.Errorf("invalid DNS server IP address at index %d: %s", i, dns)
		}
	}
	
	if scope.Lease < 0 {
		return fmt.Errorf("lease must be non-negative")
	}
	
	return nil
}

