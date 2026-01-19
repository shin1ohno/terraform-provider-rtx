package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSNMPService_Get(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*MockExecutor)
		expected    *SNMPConfig
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful get with full config",
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "show config | grep snmp").
					Return([]byte(`snmp sysname RTX830-Main
snmp syslocation Tokyo Data Center
snmp syscontact admin@example.com
snmp community read-only public
snmp community read-write private
snmp trap community public
snmp host 192.168.1.100
snmp trap enable snmp coldstart warmstart`), nil)
			},
			expected: &SNMPConfig{
				SysName:     "RTX830-Main",
				SysLocation: "Tokyo Data Center",
				SysContact:  "admin@example.com",
				Communities: []SNMPCommunity{
					{Name: "public", Permission: "ro"},
					{Name: "private", Permission: "rw"},
				},
				Hosts: []SNMPHost{
					{Address: "192.168.1.100", Community: "public"},
				},
				TrapEnable: []string{"coldstart", "warmstart"},
			},
			expectedErr: false,
		},
		{
			name: "Empty config",
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "show config | grep snmp").
					Return([]byte(""), nil)
			},
			expected: &SNMPConfig{
				Communities: []SNMPCommunity{},
				Hosts:       []SNMPHost{},
				TrapEnable:  []string{},
			},
			expectedErr: false,
		},
		{
			name: "Execution error",
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "show config | grep snmp").
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

			service := &SNMPService{executor: mockExecutor}
			result, err := service.Get(context.Background())

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.SysName, result.SysName)
				assert.Equal(t, tt.expected.SysLocation, result.SysLocation)
				assert.Equal(t, tt.expected.SysContact, result.SysContact)
				assert.Len(t, result.Communities, len(tt.expected.Communities))
				assert.Len(t, result.Hosts, len(tt.expected.Hosts))
				assert.Len(t, result.TrapEnable, len(tt.expected.TrapEnable))
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestSNMPService_Create(t *testing.T) {
	tests := []struct {
		name        string
		config      SNMPConfig
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Validation error - invalid community permission",
			config: SNMPConfig{
				Communities: []SNMPCommunity{
					{Name: "public", Permission: "invalid"},
				},
			},
			mockSetup:   func(m *MockExecutor) {},
			expectedErr: true,
			errMessage:  "community permission",
		},
		{
			name: "Validation error - invalid host IP",
			config: SNMPConfig{
				Communities: []SNMPCommunity{
					{Name: "public", Permission: "ro"},
				},
				Hosts: []SNMPHost{
					{Address: "invalid-ip"},
				},
			},
			mockSetup:   func(m *MockExecutor) {},
			expectedErr: true,
			errMessage:  "invalid host IP",
		},
		{
			name: "Execution error on sysname command",
			config: SNMPConfig{
				SysName: "TestRouter",
				Communities: []SNMPCommunity{
					{Name: "public", Permission: "ro"},
				},
			},
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "snmp sysname TestRouter").
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

			service := &SNMPService{
				executor: mockExecutor,
				client:   nil, // client is nil since we only test validation errors
			}

			err := service.Create(context.Background(), tt.config)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

// Note: Delete tests require a full rtxClient mock to test SaveConfig.
// For now, we test the validation and parser logic which are fully unit testable.

func TestStringSlicesEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected bool
	}{
		{
			name:     "equal slices",
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "c"},
			expected: true,
		},
		{
			name:     "different length",
			a:        []string{"a", "b"},
			b:        []string{"a", "b", "c"},
			expected: false,
		},
		{
			name:     "different content",
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "x", "c"},
			expected: false,
		},
		{
			name:     "empty slices",
			a:        []string{},
			b:        []string{},
			expected: true,
		},
		{
			name:     "nil slices",
			a:        nil,
			b:        nil,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringSlicesEqual(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
