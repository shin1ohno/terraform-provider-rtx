package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDHCPService_CreateScope(t *testing.T) {
	tests := []struct {
		name        string
		scope       DHCPScope
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful basic scope creation",
			scope: DHCPScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
			},
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "dhcp scope 1 192.168.1.100-192.168.1.200/24").
					Return([]byte(""), nil)
				// Mock for state verification after creation
				m.On("Run", mock.Anything, "show config | grep \"dhcp scope\"").
					Return([]byte("dhcp scope 1 192.168.1.100-192.168.1.200/24"), nil)
			},
			expectedErr: false,
		},
		{
			name: "Successful scope creation with all options",
			scope: DHCPScope{
				ID:         10,
				RangeStart: "192.168.100.10",
				RangeEnd:   "192.168.100.100",
				Prefix:     24,
				Gateway:    "192.168.100.1",
				DNSServers: []string{"192.168.100.1", "8.8.8.8"},
				Lease:      7200,
				DomainName: "internal.local",
			},
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "dhcp scope 10 192.168.100.10-192.168.100.100/24 gateway 192.168.100.1 dns 192.168.100.1 8.8.8.8 lease 7200 domain internal.local").
					Return([]byte(""), nil)
				// Mock for state verification after creation
				m.On("Run", mock.Anything, "show config | grep \"dhcp scope\"").
					Return([]byte("dhcp scope 10 192.168.100.10-192.168.100.100/24 gateway 192.168.100.1 dns 192.168.100.1 8.8.8.8 lease 7200 domain internal.local"), nil)
			},
			expectedErr: false,
		},
		{
			name: "Creation with conflict, then success on retry",
			scope: DHCPScope{
				ID:         2,
				RangeStart: "10.0.0.10",
				RangeEnd:   "10.0.0.100",
				Prefix:     24,
			},
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "dhcp scope 2 10.0.0.10-10.0.0.100/24").
					Return([]byte("Error: Scope already exists"), nil).Once()
				m.On("Run", mock.Anything, "dhcp scope 2 10.0.0.10-10.0.0.100/24").
					Return([]byte(""), nil).Once()
				// Mock for state verification after creation
				m.On("Run", mock.Anything, "show config | grep \"dhcp scope\"").
					Return([]byte("dhcp scope 2 10.0.0.10-10.0.0.100/24"), nil)
			},
			expectedErr: false,
		},
		{
			name: "Execution error",
			scope: DHCPScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
			},
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "dhcp scope 1 192.168.1.100-192.168.1.200/24").
					Return(nil, errors.New("connection failed"))
			},
			expectedErr: true,
			errMessage:  "connection failed",
		},
		{
			name: "Invalid scope validation",
			scope: DHCPScope{
				ID:         0, // Invalid ID
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
			},
			mockSetup:   func(m *MockExecutor) {}, // No setup needed, validation fails first
			expectedErr: true,
			errMessage:  "scope_id must be between 1 and 255",
		},
		{
			name: "Command error with non-retryable output",
			scope: DHCPScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
			},
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "dhcp scope 1 192.168.1.100-192.168.1.200/24").
					Return([]byte("Error: Invalid configuration"), nil)
			},
			expectedErr: true,
			errMessage:  "command failed: Error: Invalid configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &DHCPService{executor: mockExecutor}
			err := service.CreateScope(context.Background(), tt.scope)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestDHCPService_DeleteScope(t *testing.T) {
	tests := []struct {
		name        string
		scopeID     int
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name:    "Successful deletion",
			scopeID: 1,
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "no dhcp scope 1").
					Return([]byte(""), nil)
				// Mock for state verification after deletion - scope should not be found
				m.On("Run", mock.Anything, "show config | grep \"dhcp scope\"").
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name:    "Deletion with transient error, then success on retry",
			scopeID: 2,
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "no dhcp scope 2").
					Return([]byte("Error: System busy, please try again"), nil).Once()
				m.On("Run", mock.Anything, "no dhcp scope 2").
					Return([]byte(""), nil).Once()
				// Mock for state verification after deletion - scope should not be found
				m.On("Run", mock.Anything, "show config | grep \"dhcp scope\"").
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name:    "Invalid scope ID - zero",
			scopeID: 0,
			mockSetup: func(m *MockExecutor) {
				// No setup needed, validation fails first
			},
			expectedErr: true,
			errMessage:  "scope_id must be between 1 and 255",
		},
		{
			name:    "Invalid scope ID - too high",
			scopeID: 256,
			mockSetup: func(m *MockExecutor) {
				// No setup needed, validation fails first
			},
			expectedErr: true,
			errMessage:  "scope_id must be between 1 and 255",
		},
		{
			name:    "Execution error",
			scopeID: 1,
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "no dhcp scope 1").
					Return(nil, errors.New("connection failed"))
			},
			expectedErr: true,
			errMessage:  "connection failed",
		},
		{
			name:    "Command error with non-retryable output",
			scopeID: 1,
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "no dhcp scope 1").
					Return([]byte("Error: Scope not found"), nil)
			},
			expectedErr: true,
			errMessage:  "command failed: Error: Scope not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &DHCPService{executor: mockExecutor}
			err := service.DeleteScope(context.Background(), tt.scopeID)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestDHCPService_UpdateScope(t *testing.T) {
	tests := []struct {
		name        string
		scope       DHCPScope
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful update",
			scope: DHCPScope{
				ID:         1,
				RangeStart: "192.168.1.50",
				RangeEnd:   "192.168.1.150",
				Prefix:     24,
				Gateway:    "192.168.1.1",
			},
			mockSetup: func(m *MockExecutor) {
				// Delete command
				m.On("Run", mock.Anything, "no dhcp scope 1").
					Return([]byte(""), nil)
				// Create command
				m.On("Run", mock.Anything, "dhcp scope 1 192.168.1.50-192.168.1.150/24 gateway 192.168.1.1").
					Return([]byte(""), nil)
				// Mock for state verification after update
				m.On("Run", mock.Anything, "show config | grep \"dhcp scope\"").
					Return([]byte("dhcp scope 1 192.168.1.50-192.168.1.150/24 gateway 192.168.1.1"), nil)
			},
			expectedErr: false,
		},
		{
			name: "Update with delete failing (scope doesn't exist) - should continue",
			scope: DHCPScope{
				ID:         2,
				RangeStart: "10.0.0.10",
				RangeEnd:   "10.0.0.100",
				Prefix:     24,
			},
			mockSetup: func(m *MockExecutor) {
				// Delete command fails because scope doesn't exist
				m.On("Run", mock.Anything, "no dhcp scope 2").
					Return([]byte("Error: Scope not found"), nil)
				// Create command still executes
				m.On("Run", mock.Anything, "dhcp scope 2 10.0.0.10-10.0.0.100/24").
					Return([]byte(""), nil)
				// Mock for state verification after update
				m.On("Run", mock.Anything, "show config | grep \"dhcp scope\"").
					Return([]byte("dhcp scope 2 10.0.0.10-10.0.0.100/24"), nil)
			},
			expectedErr: false,
		},
		{
			name: "Invalid scope validation",
			scope: DHCPScope{
				ID:         0, // Invalid ID
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
			},
			mockSetup:   func(m *MockExecutor) {}, // No setup needed, validation fails first
			expectedErr: true,
			errMessage:  "scope_id must be between 1 and 255",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &DHCPService{executor: mockExecutor}
			err := service.UpdateScope(context.Background(), tt.scope)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestDHCPService_CreateScope_RetryBehavior(t *testing.T) {
	t.Run("Retry on conflict with exponential backoff", func(t *testing.T) {
		mockExecutor := new(MockExecutor)
		
		// First 3 attempts fail with conflict
		mockExecutor.On("Run", mock.Anything, "dhcp scope 1 192.168.1.100-192.168.1.200/24").
			Return([]byte("Error: Scope already exists"), nil).Times(3)
		// Fourth attempt succeeds
		mockExecutor.On("Run", mock.Anything, "dhcp scope 1 192.168.1.100-192.168.1.200/24").
			Return([]byte(""), nil).Once()
		// Mock for state verification after creation
		mockExecutor.On("Run", mock.Anything, "show config | grep \"dhcp scope\"").
			Return([]byte("dhcp scope 1 192.168.1.100-192.168.1.200/24"), nil)

		service := &DHCPService{executor: mockExecutor}
		scope := DHCPScope{
			ID:         1,
			RangeStart: "192.168.1.100",
			RangeEnd:   "192.168.1.200",
			Prefix:     24,
		}

		startTime := time.Now()
		err := service.CreateScope(context.Background(), scope)
		duration := time.Since(startTime)

		assert.NoError(t, err)
		// Should have some delay due to retries (at least 100ms + 200ms + 400ms = 700ms)
		assert.Greater(t, duration, 500*time.Millisecond)
		mockExecutor.AssertExpectations(t)
	})

	t.Run("Give up after max retries", func(t *testing.T) {
		mockExecutor := new(MockExecutor)
		
		// All attempts fail with conflict
		mockExecutor.On("Run", mock.Anything, "dhcp scope 1 192.168.1.100-192.168.1.200/24").
			Return([]byte("Error: Scope conflict"), nil).Times(6) // Max retries + 1

		service := &DHCPService{executor: mockExecutor}
		scope := DHCPScope{
			ID:         1,
			RangeStart: "192.168.1.100",
			RangeEnd:   "192.168.1.200",
			Prefix:     24,
		}

		err := service.CreateScope(context.Background(), scope)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command failed after 6 attempts")
		mockExecutor.AssertExpectations(t)
	})
}

func TestDHCPService_ValidateDHCPScope(t *testing.T) {
	tests := []struct {
		name        string
		scope       DHCPScope
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Valid basic scope",
			scope: DHCPScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
			},
			expectedErr: false,
		},
		{
			name: "Valid scope with all options",
			scope: DHCPScope{
				ID:         255,
				RangeStart: "10.0.0.10",
				RangeEnd:   "10.0.0.100",
				Prefix:     8,
				Gateway:    "10.0.0.1",
				DNSServers: []string{"8.8.8.8", "8.8.4.4"},
				Lease:      3600,
				DomainName: "example.com",
			},
			expectedErr: false,
		},
		{
			name: "Invalid scope ID - zero",
			scope: DHCPScope{
				ID:         0,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
			},
			expectedErr: true,
			errMessage:  "scope_id must be between 1 and 255",
		},
		{
			name: "Invalid scope ID - too high",
			scope: DHCPScope{
				ID:         256,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
			},
			expectedErr: true,
			errMessage:  "scope_id must be between 1 and 255",
		},
		{
			name: "Invalid range_start",
			scope: DHCPScope{
				ID:         1,
				RangeStart: "invalid.ip",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
			},
			expectedErr: true,
			errMessage:  "invalid range_start IP address",
		},
		{
			name: "Invalid range_end",
			scope: DHCPScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "invalid.ip",
				Prefix:     24,
			},
			expectedErr: true,
			errMessage:  "invalid range_end IP address",
		},
		{
			name: "Invalid prefix - too low",
			scope: DHCPScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     7,
			},
			expectedErr: true,
			errMessage:  "prefix must be between 8 and 32",
		},
		{
			name: "Invalid prefix - too high",
			scope: DHCPScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     33,
			},
			expectedErr: true,
			errMessage:  "prefix must be between 8 and 32",
		},
		{
			name: "Invalid gateway",
			scope: DHCPScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
				Gateway:    "invalid.gateway",
			},
			expectedErr: true,
			errMessage:  "invalid gateway IP address",
		},
		{
			name: "Invalid DNS server",
			scope: DHCPScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
				DNSServers: []string{"8.8.8.8", "invalid.dns"},
			},
			expectedErr: true,
			errMessage:  "invalid DNS server IP address at index 1",
		},
		{
			name: "Invalid lease - negative",
			scope: DHCPScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
				Lease:      -1,
			},
			expectedErr: true,
			errMessage:  "lease must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDHCPScope(tt.scope)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}