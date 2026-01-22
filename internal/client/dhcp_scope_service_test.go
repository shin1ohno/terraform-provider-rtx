package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDHCPScopeService_GetScope(t *testing.T) {
	tests := []struct {
		name        string
		scopeID     int
		mockSetup   func(*MockExecutor)
		expected    *DHCPScope
		expectedErr bool
		errMessage  string
	}{
		{
			name:    "Successful get with CIDR network",
			scopeID: 1,
			mockSetup: func(m *MockExecutor) {
				output := `dhcp scope 1 192.168.1.0/24 gateway 192.168.1.1 expire 72:00
dhcp scope option 1 dns=8.8.8.8,8.8.4.4
`
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep "dhcp scope"`
				})).Return([]byte(output), nil)
			},
			expected: &DHCPScope{
				ScopeID:   1,
				Network:   "192.168.1.0/24",
				LeaseTime: "72:00",
			},
			expectedErr: false,
		},
		{
			name:    "Successful get with IP range",
			scopeID: 2,
			mockSetup: func(m *MockExecutor) {
				output := `dhcp scope 2 192.168.2.100-192.168.2.200/24 gateway 192.168.2.1 expire 24:00
`
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep "dhcp scope"`
				})).Return([]byte(output), nil)
			},
			expected: &DHCPScope{
				ScopeID:   2,
				Network:   "192.168.2.100-192.168.2.200/24",
				LeaseTime: "24:00",
			},
			expectedErr: false,
		},
		{
			name:    "Execution error",
			scopeID: 1,
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, mock.Anything).
					Return(nil, errors.New("connection failed"))
			},
			expected:    nil,
			expectedErr: true,
			errMessage:  "connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &DHCPScopeService{executor: mockExecutor}
			result, err := service.GetScope(context.Background(), tt.scopeID)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ScopeID, result.ScopeID)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestDHCPScopeService_CreateScope(t *testing.T) {
	tests := []struct {
		name        string
		scope       DHCPScope
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful CIDR scope creation with batch",
			scope: DHCPScope{
				ScopeID:   1,
				Network:   "192.168.1.0/24",
				LeaseTime: "72:00",
				Options: DHCPScopeOptions{
					DNSServers: []string{"8.8.8.8", "8.8.4.4"},
					Routers:    []string{"192.168.1.1"},
				},
			},
			mockSetup: func(m *MockExecutor) {
				// Expected batch commands for scope creation
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasScopeCmd := false
					for _, cmd := range cmds {
						if cmd == "dhcp scope 1 192.168.1.0/24 expire 72:00" {
							hasScopeCmd = true
						}
					}
					return hasScopeCmd && len(cmds) > 0
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Successful scope creation with different network",
			scope: DHCPScope{
				ScopeID:   2,
				Network:   "10.0.0.0/24",
				LeaseTime: "24:00",
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					for _, cmd := range cmds {
						if cmd == "dhcp scope 2 10.0.0.0/24 expire 24:00" {
							return true
						}
					}
					return false
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Batch execution error",
			scope: DHCPScope{
				ScopeID:   1,
				Network:   "192.168.1.0/24",
				LeaseTime: "72:00",
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.Anything).
					Return(nil, errors.New("connection failed"))
			},
			expectedErr: true,
			errMessage:  "connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &DHCPScopeService{executor: mockExecutor}
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

func TestDHCPScopeService_UpdateScope(t *testing.T) {
	tests := []struct {
		name        string
		scope       DHCPScope
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful update with batch",
			scope: DHCPScope{
				ScopeID:   1,
				Network:   "192.168.1.0/24",
				LeaseTime: "48:00",
				Options: DHCPScopeOptions{
					DNSServers: []string{"1.1.1.1"},
				},
			},
			mockSetup: func(m *MockExecutor) {
				// First: Get current config
				currentOutput := `dhcp scope 1 192.168.1.0/24 gateway 192.168.1.1 expire 72:00
`
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep "dhcp scope"`
				})).Return([]byte(currentOutput), nil)

				// Then: RunBatch with updated commands
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasLeaseTimeUpdate := false
					for _, cmd := range cmds {
						if cmd == "dhcp scope 1 192.168.1.0/24 expire 48:00" {
							hasLeaseTimeUpdate = true
						}
					}
					return hasLeaseTimeUpdate && len(cmds) > 0
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &DHCPScopeService{executor: mockExecutor}
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

func TestDHCPScopeService_DeleteScope(t *testing.T) {
	tests := []struct {
		name        string
		scopeID     int
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name:    "Successful delete with batch",
			scopeID: 1,
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, []string{"no dhcp scope 1"}).
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name:    "Execution error",
			scopeID: 1,
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, []string{"no dhcp scope 1"}).
					Return(nil, errors.New("connection failed"))
			},
			expectedErr: true,
			errMessage:  "connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &DHCPScopeService{executor: mockExecutor}
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

func TestDHCPScopeService_UsesRunBatch(t *testing.T) {
	t.Run("CreateScope uses RunBatch for all commands", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		// Capture commands to verify RunBatch is used
		var capturedCommands []string
		mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				capturedCommands = args.Get(1).([]string)
			}).
			Return([]byte(""), nil)

		service := &DHCPScopeService{executor: mockExecutor}
		err := service.CreateScope(context.Background(), DHCPScope{
			ScopeID:   1,
			Network:   "192.168.1.0/24",
			LeaseTime: "72:00",
			Options: DHCPScopeOptions{
				DNSServers: []string{"8.8.8.8"},
			},
		})

		assert.NoError(t, err)

		// Verify scope command is present
		hasScopeCmd := false
		for _, cmd := range capturedCommands {
			if cmd == "dhcp scope 1 192.168.1.0/24 expire 72:00" {
				hasScopeCmd = true
				break
			}
		}
		assert.True(t, hasScopeCmd, "Expected scope command to be included")
	})

	t.Run("DeleteScope uses RunBatch", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		mockExecutor.On("RunBatch", mock.Anything, []string{"no dhcp scope 1"}).
			Return([]byte(""), nil)

		service := &DHCPScopeService{executor: mockExecutor}
		err := service.DeleteScope(context.Background(), 1)

		assert.NoError(t, err)
		mockExecutor.AssertExpectations(t)
	})
}
