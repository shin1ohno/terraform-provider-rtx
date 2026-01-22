package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockExecutor is a mock implementation of Executor interface
type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) Run(ctx context.Context, cmd string) ([]byte, error) {
	args := m.Called(ctx, cmd)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockExecutor) RunBatch(ctx context.Context, cmds []string) ([]byte, error) {
	args := m.Called(ctx, cmds)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func TestDHCPService_CreateBinding(t *testing.T) {
	tests := []struct {
		name        string
		binding     DHCPBinding
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful MAC binding creation",
			binding: DHCPBinding{
				ScopeID:             1,
				IPAddress:           "192.168.1.100",
				MACAddress:          "00:11:22:33:44:55",
				UseClientIdentifier: false,
			},
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "dhcp scope bind 1 192.168.1.100 00:11:22:33:44:55").
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Successful ethernet binding creation",
			binding: DHCPBinding{
				ScopeID:             1,
				IPAddress:           "192.168.1.101",
				MACAddress:          "00:aa:bb:cc:dd:ee",
				UseClientIdentifier: true,
			},
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "dhcp scope bind 1 192.168.1.101 ethernet 00:aa:bb:cc:dd:ee").
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Execution error",
			binding: DHCPBinding{
				ScopeID:             1,
				IPAddress:           "192.168.1.100",
				MACAddress:          "00:11:22:33:44:55",
				UseClientIdentifier: false,
			},
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "dhcp scope bind 1 192.168.1.100 00:11:22:33:44:55").
					Return(nil, errors.New("connection failed"))
			},
			expectedErr: true,
			errMessage:  "connection failed",
		},
		{
			name: "Command error with error output",
			binding: DHCPBinding{
				ScopeID:             1,
				IPAddress:           "192.168.1.100",
				MACAddress:          "00:11:22:33:44:55",
				UseClientIdentifier: false,
			},
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "dhcp scope bind 1 192.168.1.100 00:11:22:33:44:55").
					Return([]byte("Error: IP address already bound"), nil)
			},
			expectedErr: true,
			errMessage:  "command failed: Error: IP address already bound",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &DHCPService{executor: mockExecutor}
			err := service.CreateBinding(context.Background(), tt.binding)

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

func TestDHCPService_DeleteBinding(t *testing.T) {
	tests := []struct {
		name        string
		scopeID     int
		ipAddress   string
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name:      "Successful deletion",
			scopeID:   1,
			ipAddress: "192.168.1.100",
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "no dhcp scope bind 1 192.168.1.100").
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name:      "Execution error",
			scopeID:   1,
			ipAddress: "192.168.1.100",
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "no dhcp scope bind 1 192.168.1.100").
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

			service := &DHCPService{executor: mockExecutor}
			err := service.DeleteBinding(context.Background(), tt.scopeID, tt.ipAddress)

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

func TestDHCPService_ListBindings(t *testing.T) {
	tests := []struct {
		name        string
		scopeID     int
		mockSetup   func(*MockExecutor)
		expected    []DHCPBinding
		expectedErr bool
		errMessage  string
	}{
		{
			name:    "Successful list with bindings",
			scopeID: 1,
			mockSetup: func(m *MockExecutor) {
				output := `dhcp scope bind 1 192.168.1.100 00:11:22:33:44:55
dhcp scope bind 1 192.168.1.101 ethernet 00:aa:bb:cc:dd:ee
`
				m.On("Run", mock.Anything, `show config | grep "dhcp scope bind 1"`).
					Return([]byte(output), nil)
			},
			expected: []DHCPBinding{
				{
					ScopeID:             1,
					IPAddress:           "192.168.1.100",
					MACAddress:          "00:11:22:33:44:55",
					UseClientIdentifier: false,
				},
				{
					ScopeID:             1,
					IPAddress:           "192.168.1.101",
					MACAddress:          "00:aa:bb:cc:dd:ee",
					UseClientIdentifier: true,
				},
			},
			expectedErr: false,
		},
		{
			name:    "Empty bindings",
			scopeID: 1,
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, `show config | grep "dhcp scope bind 1"`).
					Return([]byte(""), nil)
			},
			expected:    []DHCPBinding{},
			expectedErr: false,
		},
		{
			name:    "Execution error",
			scopeID: 1,
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, `show config | grep "dhcp scope bind 1"`).
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

			service := &DHCPService{executor: mockExecutor}
			result, err := service.ListBindings(context.Background(), tt.scopeID)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}