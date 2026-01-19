package client

import (
	"context"
	"fmt"
)

// ConfigService handles configuration changes on RTX routers
type ConfigService struct {
	executor Executor
}

// NewConfigService creates a new configuration service
func NewConfigService(executor Executor) *ConfigService {
	return &ConfigService{
		executor: executor,
	}
}

// AddDNSHost adds a DNS host entry to the router configuration
func (s *ConfigService) AddDNSHost(ctx context.Context, host DNSHost) error {
	// Will be implemented when working on rtx_dns_host resource
	return fmt.Errorf("not implemented")
}

// UpdateDNSHost updates an existing DNS host entry
func (s *ConfigService) UpdateDNSHost(ctx context.Context, oldHost, newHost DNSHost) error {
	// Will be implemented when working on rtx_dns_host resource
	return fmt.Errorf("not implemented")
}

// DeleteDNSHost removes a DNS host entry from the router configuration
func (s *ConfigService) DeleteDNSHost(ctx context.Context, host DNSHost) error {
	// Will be implemented when working on rtx_dns_host resource
	return fmt.Errorf("not implemented")
}

// ApplyConfig commits pending configuration changes
func (s *ConfigService) ApplyConfig(ctx context.Context) error {
	// Will be implemented when working on configuration management
	return fmt.Errorf("not implemented")
}

// RevertConfig reverts to the previous configuration
func (s *ConfigService) RevertConfig(ctx context.Context) error {
	// Will be implemented when working on configuration management
	return fmt.Errorf("not implemented")
}