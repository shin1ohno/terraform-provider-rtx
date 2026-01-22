package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDNSService_Get(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*MockExecutor)
		expected    *DNSConfig
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful get - basic config",
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "show config | grep dns").
					Return([]byte(`dns server 8.8.8.8 8.8.4.4
dns domain example.com
dns service recursive
dns private address spoof on
`), nil)
			},
			expected: &DNSConfig{
				DomainLookup: true,
				DomainName:   "example.com",
				NameServers:  []string{"8.8.8.8", "8.8.4.4"},
				ServerSelect: []DNSServerSelect{},
				Hosts:        []DNSHost{},
				ServiceOn:    true,
				PrivateSpoof: true,
			},
			expectedErr: false,
		},
		{
			name: "Successful get - full config",
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "show config | grep dns").
					Return([]byte(`dns server 8.8.8.8
dns server select 1 192.168.1.1 internal.example.com
dns static router 192.168.1.1
dns service recursive
`), nil)
			},
			expected: &DNSConfig{
				DomainLookup: true,
				DomainName:   "",
				NameServers:  []string{"8.8.8.8"},
				ServerSelect: []DNSServerSelect{
					{ID: 1, Servers: []DNSServer{{Address: "192.168.1.1"}}, RecordType: "a", QueryPattern: "internal.example.com"},
				},
				Hosts: []DNSHost{
					{Name: "router", Address: "192.168.1.1"},
				},
				ServiceOn:    true,
				PrivateSpoof: false,
			},
			expectedErr: false,
		},
		{
			name: "Empty config",
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "show config | grep dns").
					Return([]byte(""), nil)
			},
			expected: &DNSConfig{
				DomainLookup: true,
				NameServers:  []string{},
				ServerSelect: []DNSServerSelect{},
				Hosts:        []DNSHost{},
				ServiceOn:    false,
				PrivateSpoof: false,
			},
			expectedErr: false,
		},
		{
			name: "Execution error",
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "show config | grep dns").
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

			service := &DNSService{executor: mockExecutor}
			config, err := service.Get(context.Background())

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.DomainLookup, config.DomainLookup)
				assert.Equal(t, tt.expected.DomainName, config.DomainName)
				assert.Equal(t, tt.expected.NameServers, config.NameServers)
				assert.Equal(t, len(tt.expected.ServerSelect), len(config.ServerSelect))
				assert.Equal(t, len(tt.expected.Hosts), len(config.Hosts))
				assert.Equal(t, tt.expected.ServiceOn, config.ServiceOn)
				assert.Equal(t, tt.expected.PrivateSpoof, config.PrivateSpoof)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestDNSService_Configure(t *testing.T) {
	tests := []struct {
		name        string
		config      DNSConfig
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful basic configuration",
			config: DNSConfig{
				DomainLookup: true,
				NameServers:  []string{"8.8.8.8", "8.8.4.4"},
				ServiceOn:    true,
				PrivateSpoof: false,
			},
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "dns server 8.8.8.8 8.8.4.4").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "dns service recursive").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "dns private address spoof off").
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Configuration with domain name",
			config: DNSConfig{
				DomainLookup: true,
				DomainName:   "example.com",
				NameServers:  []string{"8.8.8.8"},
				ServiceOn:    true,
				PrivateSpoof: true,
			},
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "dns domain example.com").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "dns server 8.8.8.8").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "dns service recursive").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "dns private address spoof on").
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Configuration with domain lookup disabled",
			config: DNSConfig{
				DomainLookup: false,
				NameServers:  []string{"8.8.8.8"},
				ServiceOn:    false,
				PrivateSpoof: false,
			},
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "no dns domain lookup").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "dns server 8.8.8.8").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "dns service off").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "dns private address spoof off").
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Configuration with server select and hosts",
			config: DNSConfig{
				DomainLookup: true,
				NameServers:  []string{"8.8.8.8"},
				ServerSelect: []DNSServerSelect{
					{ID: 1, Servers: []DNSServer{{Address: "192.168.1.1"}}, QueryPattern: "internal.example.com"},
				},
				Hosts: []DNSHost{
					{Name: "router", Address: "192.168.1.1"},
				},
				ServiceOn:    true,
				PrivateSpoof: false,
			},
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "dns server 8.8.8.8").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "dns server select 1 192.168.1.1 internal.example.com").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "dns static router 192.168.1.1").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "dns service recursive").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "dns private address spoof off").
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Validation error - invalid name server",
			config: DNSConfig{
				NameServers: []string{"invalid"},
			},
			mockSetup:   func(m *MockExecutor) {},
			expectedErr: true,
			errMessage:  "invalid",
		},
		{
			name: "Validation error - too many name servers",
			config: DNSConfig{
				NameServers: []string{"8.8.8.8", "8.8.4.4", "1.1.1.1", "1.0.0.1"},
			},
			mockSetup:   func(m *MockExecutor) {},
			expectedErr: true,
			errMessage:  "maximum 3",
		},
		{
			name: "Execution error",
			config: DNSConfig{
				DomainLookup: true,
				NameServers:  []string{"8.8.8.8"},
				ServiceOn:    true,
				PrivateSpoof: false,
			},
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "dns server 8.8.8.8").
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

			service := &DNSService{executor: mockExecutor}
			err := service.Configure(context.Background(), tt.config)

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

func TestDNSService_Reset(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*MockExecutor)
		expectedErr bool
	}{
		{
			name: "Successful reset - empty config",
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "show config | grep dns").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "no dns server").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "no dns domain").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "dns service off").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "dns private address spoof off").
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Successful reset - with server select and hosts",
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "show config | grep dns").
					Return([]byte(`dns server 8.8.8.8
dns server select 1 192.168.1.1 internal.example.com
dns static router 192.168.1.1
`), nil)
				m.On("Run", mock.Anything, "no dns server select 1").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "no dns static router").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "no dns server").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "no dns domain").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "dns service off").
					Return([]byte(""), nil)
				m.On("Run", mock.Anything, "dns private address spoof off").
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &DNSService{executor: mockExecutor}
			err := service.Reset(context.Background())

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestSlicesEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected bool
	}{
		{
			name:     "Equal slices",
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "c"},
			expected: true,
		},
		{
			name:     "Different lengths",
			a:        []string{"a", "b"},
			b:        []string{"a", "b", "c"},
			expected: false,
		},
		{
			name:     "Different content",
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "x", "c"},
			expected: false,
		},
		{
			name:     "Both empty",
			a:        []string{},
			b:        []string{},
			expected: true,
		},
		{
			name:     "One nil",
			a:        nil,
			b:        []string{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slicesEqual(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}
