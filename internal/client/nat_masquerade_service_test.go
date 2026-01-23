package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// intPtr is a helper function to create *int pointers for test data
func intPtr(i int) *int {
	return &i
}

func TestNATMasqueradeService_Create(t *testing.T) {
	tests := []struct {
		name        string
		nat         NATMasquerade
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful NAT masquerade creation",
			nat: NATMasquerade{
				DescriptorID: 1,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.1.0-192.168.1.255",
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasType := false
					hasOuter := false
					hasInner := false
					for _, cmd := range cmds {
						if cmd == "nat descriptor type 1 masquerade" {
							hasType = true
						}
						if cmd == "nat descriptor address outer 1 ipcp" {
							hasOuter = true
						}
						if cmd == "nat descriptor address inner 1 192.168.1.0-192.168.1.255" {
							hasInner = true
						}
					}
					return hasType && hasOuter && hasInner
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Successful NAT masquerade creation with static entries",
			nat: NATMasquerade{
				DescriptorID: 1,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.1.0-192.168.1.255",
				StaticEntries: []MasqueradeStaticEntry{
					{
						EntryNumber:       1,
						InsideLocal:       "192.168.1.100",
						InsideLocalPort:   intPtr(8080),
						OutsideGlobal:     "ipcp",
						OutsideGlobalPort: intPtr(80),
						Protocol:          "tcp",
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasType := false
					hasStatic := false
					for _, cmd := range cmds {
						if cmd == "nat descriptor type 1 masquerade" {
							hasType = true
						}
						if cmd == "nat descriptor masquerade static 1 1 ipcp:80=192.168.1.100:8080 tcp" {
							hasStatic = true
						}
					}
					return hasType && hasStatic
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Validation error - invalid descriptor ID",
			nat: NATMasquerade{
				DescriptorID: 0,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.1.0-192.168.1.255",
			},
			mockSetup:   func(m *MockExecutor) {},
			expectedErr: true,
			errMessage:  "descriptor ID",
		},
		{
			name: "Validation error - empty outer address",
			nat: NATMasquerade{
				DescriptorID: 1,
				OuterAddress: "",
				InnerNetwork: "192.168.1.0-192.168.1.255",
			},
			mockSetup:   func(m *MockExecutor) {},
			expectedErr: true,
			errMessage:  "outer address",
		},
		{
			name: "Execution error on type command",
			nat: NATMasquerade{
				DescriptorID: 1,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.1.0-192.168.1.255",
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.Anything).
					Return(nil, errors.New("connection failed"))
			},
			expectedErr: true,
			errMessage:  "connection failed",
		},
		{
			name: "Command error with error output",
			nat: NATMasquerade{
				DescriptorID: 1,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.1.0-192.168.1.255",
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.Anything).
					Return([]byte("Error: invalid command"), nil)
			},
			expectedErr: true,
			errMessage:  "command failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &NATMasqueradeService{executor: mockExecutor}
			err := service.Create(context.Background(), tt.nat)

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

func TestNATMasqueradeService_Get(t *testing.T) {
	tests := []struct {
		name         string
		descriptorID int
		mockSetup    func(*MockExecutor)
		expected     *NATMasquerade
		expectedErr  bool
		errMessage   string
	}{
		{
			name:         "Successful get",
			descriptorID: 1,
			mockSetup: func(m *MockExecutor) {
				output := `nat descriptor type 1 masquerade
nat descriptor address outer 1 ipcp
nat descriptor address inner 1 192.168.1.0-192.168.1.255
`
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep "nat descriptor.*1"`
				})).Return([]byte(output), nil)
			},
			expected: &NATMasquerade{
				DescriptorID:  1,
				OuterAddress:  "ipcp",
				InnerNetwork:  "192.168.1.0-192.168.1.255",
				StaticEntries: []MasqueradeStaticEntry{},
			},
			expectedErr: false,
		},
		{
			name:         "Successful get with multiple static entries",
			descriptorID: 2,
			mockSetup: func(m *MockExecutor) {
				output := `nat descriptor type 2 masquerade
nat descriptor address outer 2 ipcp
nat descriptor address inner 2 192.168.2.0-192.168.2.255
nat descriptor masquerade static 2 1 ipcp:80=192.168.2.10:8080 tcp
nat descriptor masquerade static 2 2 ipcp:443=192.168.2.10:8443 tcp
nat descriptor masquerade static 2 3 ipcp:53=192.168.2.20:53 udp
`
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep "nat descriptor.*2"`
				})).Return([]byte(output), nil)
			},
			expected: &NATMasquerade{
				DescriptorID: 2,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.2.0-192.168.2.255",
				StaticEntries: []MasqueradeStaticEntry{
					{
						EntryNumber:       1,
						InsideLocal:       "192.168.2.10",
						InsideLocalPort:   intPtr(8080),
						OutsideGlobal:     "ipcp",
						OutsideGlobalPort: intPtr(80),
						Protocol:          "tcp",
					},
					{
						EntryNumber:       2,
						InsideLocal:       "192.168.2.10",
						InsideLocalPort:   intPtr(8443),
						OutsideGlobal:     "ipcp",
						OutsideGlobalPort: intPtr(443),
						Protocol:          "tcp",
					},
					{
						EntryNumber:       3,
						InsideLocal:       "192.168.2.20",
						InsideLocalPort:   intPtr(53),
						OutsideGlobal:     "ipcp",
						OutsideGlobalPort: intPtr(53),
						Protocol:          "udp",
					},
				},
			},
			expectedErr: false,
		},
		{
			name:         "NAT masquerade not found",
			descriptorID: 99,
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep "nat descriptor.*99"`
				})).Return([]byte(""), nil)
			},
			expected:    nil,
			expectedErr: true,
			errMessage:  "not found",
		},
		{
			name:         "Execution error",
			descriptorID: 1,
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

			service := &NATMasqueradeService{executor: mockExecutor}
			result, err := service.Get(context.Background(), tt.descriptorID)

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

func TestNATMasqueradeService_Delete(t *testing.T) {
	tests := []struct {
		name         string
		descriptorID int
		mockSetup    func(*MockExecutor)
		expectedErr  bool
		errMessage   string
	}{
		{
			name:         "Successful deletion",
			descriptorID: 1,
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, []string{"no nat descriptor type 1"}).
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name:         "Execution error",
			descriptorID: 1,
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, []string{"no nat descriptor type 1"}).
					Return(nil, errors.New("connection failed"))
			},
			expectedErr: true,
			errMessage:  "connection failed",
		},
		{
			name:         "Already deleted - not found",
			descriptorID: 1,
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, []string{"no nat descriptor type 1"}).
					Return([]byte("Error: not found"), nil)
			},
			expectedErr: false, // Not found is not an error for delete
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &NATMasqueradeService{executor: mockExecutor}
			err := service.Delete(context.Background(), tt.descriptorID)

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

func TestNATMasqueradeService_List(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*MockExecutor)
		expected    []NATMasquerade
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful list with multiple NATs",
			mockSetup: func(m *MockExecutor) {
				output := `nat descriptor type 1 masquerade
nat descriptor address outer 1 ipcp
nat descriptor address inner 1 192.168.1.0-192.168.1.255
nat descriptor type 2 masquerade
nat descriptor address outer 2 pp1
nat descriptor address inner 2 10.0.0.0-10.0.0.255
`
				m.On("Run", mock.Anything, `show config | grep "nat descriptor"`).
					Return([]byte(output), nil)
			},
			expected: []NATMasquerade{
				{
					DescriptorID:  1,
					OuterAddress:  "ipcp",
					InnerNetwork:  "192.168.1.0-192.168.1.255",
					StaticEntries: []MasqueradeStaticEntry{},
				},
				{
					DescriptorID:  2,
					OuterAddress:  "pp1",
					InnerNetwork:  "10.0.0.0-10.0.0.255",
					StaticEntries: []MasqueradeStaticEntry{},
				},
			},
			expectedErr: false,
		},
		{
			name: "Empty list",
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, `show config | grep "nat descriptor"`).
					Return([]byte(""), nil)
			},
			expected:    []NATMasquerade{},
			expectedErr: false,
		},
		{
			name: "Execution error",
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, `show config | grep "nat descriptor"`).
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

			service := &NATMasqueradeService{executor: mockExecutor}
			result, err := service.List(context.Background())

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.expected, result)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestNATMasqueradeService_Update(t *testing.T) {
	tests := []struct {
		name        string
		nat         NATMasquerade
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful update outer address",
			nat: NATMasquerade{
				DescriptorID:  1,
				OuterAddress:  "pp1",
				InnerNetwork:  "192.168.1.0-192.168.1.255",
				StaticEntries: []MasqueradeStaticEntry{},
			},
			mockSetup: func(m *MockExecutor) {
				// Get current config
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep "nat descriptor.*1"`
				})).Return([]byte(`nat descriptor type 1 masquerade
nat descriptor address outer 1 ipcp
nat descriptor address inner 1 192.168.1.0-192.168.1.255
`), nil)
				// Update outer address using RunBatch
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					for _, cmd := range cmds {
						if cmd == "nat descriptor address outer 1 pp1" {
							return true
						}
					}
					return false
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Successful update with static entry changes",
			nat: NATMasquerade{
				DescriptorID: 1,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.1.0-192.168.1.255",
				StaticEntries: []MasqueradeStaticEntry{
					{
						EntryNumber:       1,
						InsideLocal:       "192.168.1.100",
						InsideLocalPort:   intPtr(8080),
						OutsideGlobal:     "ipcp",
						OutsideGlobalPort: intPtr(80),
						Protocol:          "tcp",
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				// Get current config
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep "nat descriptor.*1"`
				})).Return([]byte(`nat descriptor type 1 masquerade
nat descriptor address outer 1 ipcp
nat descriptor address inner 1 192.168.1.0-192.168.1.255
`), nil)
				// Add static entry using RunBatch
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					for _, cmd := range cmds {
						if cmd == "nat descriptor masquerade static 1 1 ipcp:80=192.168.1.100:8080 tcp" {
							return true
						}
					}
					return false
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Validation error",
			nat: NATMasquerade{
				DescriptorID: 0,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.1.0-192.168.1.255",
			},
			mockSetup:   func(m *MockExecutor) {},
			expectedErr: true,
			errMessage:  "descriptor ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &NATMasqueradeService{executor: mockExecutor}
			err := service.Update(context.Background(), tt.nat)

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

func TestNATMasqueradeService_ContextCancellation(t *testing.T) {
	tests := []struct {
		name   string
		action func(ctx context.Context, service *NATMasqueradeService) error
	}{
		{
			name: "Create cancelled",
			action: func(ctx context.Context, service *NATMasqueradeService) error {
				return service.Create(ctx, NATMasquerade{
					DescriptorID: 1,
					OuterAddress: "ipcp",
					InnerNetwork: "192.168.1.0-192.168.1.255",
				})
			},
		},
		{
			name: "Get cancelled",
			action: func(ctx context.Context, service *NATMasqueradeService) error {
				_, err := service.Get(ctx, 1)
				return err
			},
		},
		{
			name: "Update cancelled",
			action: func(ctx context.Context, service *NATMasqueradeService) error {
				return service.Update(ctx, NATMasquerade{
					DescriptorID: 1,
					OuterAddress: "ipcp",
					InnerNetwork: "192.168.1.0-192.168.1.255",
				})
			},
		},
		{
			name: "Delete cancelled",
			action: func(ctx context.Context, service *NATMasqueradeService) error {
				return service.Delete(ctx, 1)
			},
		},
		{
			name: "List cancelled",
			action: func(ctx context.Context, service *NATMasqueradeService) error {
				_, err := service.List(ctx)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			service := &NATMasqueradeService{executor: mockExecutor}

			// Create a cancelled context
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := tt.action(ctx, service)
			assert.Error(t, err)
			assert.Equal(t, context.Canceled, err)
		})
	}
}

func TestNATMasqueradeService_UsesRunBatch(t *testing.T) {
	t.Run("Create uses RunBatch for all commands", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		var capturedCommands []string
		mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				capturedCommands = args.Get(1).([]string)
			}).
			Return([]byte(""), nil)

		service := &NATMasqueradeService{executor: mockExecutor}
		err := service.Create(context.Background(), NATMasquerade{
			DescriptorID: 1,
			OuterAddress: "ipcp",
			InnerNetwork: "192.168.1.0-192.168.1.255",
			StaticEntries: []MasqueradeStaticEntry{
				{
					EntryNumber:       1,
					InsideLocal:       "192.168.1.100",
					InsideLocalPort:   intPtr(8080),
					OutsideGlobal:     "ipcp",
					OutsideGlobalPort: intPtr(80),
					Protocol:          "tcp",
				},
			},
		})

		assert.NoError(t, err)

		// Verify essential commands are present
		hasType := false
		hasOuter := false
		hasInner := false
		hasStatic := false
		for _, cmd := range capturedCommands {
			if cmd == "nat descriptor type 1 masquerade" {
				hasType = true
			}
			if cmd == "nat descriptor address outer 1 ipcp" {
				hasOuter = true
			}
			if cmd == "nat descriptor address inner 1 192.168.1.0-192.168.1.255" {
				hasInner = true
			}
			if cmd == "nat descriptor masquerade static 1 1 ipcp:80=192.168.1.100:8080 tcp" {
				hasStatic = true
			}
		}
		assert.True(t, hasType, "Expected type command to be included")
		assert.True(t, hasOuter, "Expected outer address command to be included")
		assert.True(t, hasInner, "Expected inner network command to be included")
		assert.True(t, hasStatic, "Expected static entry command to be included")
	})

	t.Run("Delete uses RunBatch", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		mockExecutor.On("RunBatch", mock.Anything, []string{"no nat descriptor type 1"}).
			Return([]byte(""), nil)

		service := &NATMasqueradeService{executor: mockExecutor}
		err := service.Delete(context.Background(), 1)

		assert.NoError(t, err)
		mockExecutor.AssertExpectations(t)
	})
}

func TestNATMasqueradeService_ProtocolOnlyEntries(t *testing.T) {
	testCases := []struct {
		name     string
		protocol string
	}{
		{"ESP protocol", "esp"},
		{"AH protocol", "ah"},
		{"GRE protocol", "gre"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)

			var capturedCommands []string
			mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) {
					capturedCommands = args.Get(1).([]string)
				}).
				Return([]byte(""), nil)

			service := &NATMasqueradeService{executor: mockExecutor}
			err := service.Create(context.Background(), NATMasquerade{
				DescriptorID: 1,
				OuterAddress: "ipcp",
				InnerNetwork: "192.168.1.0-192.168.1.255",
				StaticEntries: []MasqueradeStaticEntry{
					{EntryNumber: 1, InsideLocal: "192.168.1.10", Protocol: tc.protocol},
				},
			})

			assert.NoError(t, err)

			// Verify protocol-only entry has no port
			hasProtocolEntry := false
			expectedCmd := "nat descriptor masquerade static 1 1 192.168.1.10 " + tc.protocol
			for _, cmd := range capturedCommands {
				if cmd == expectedCmd {
					hasProtocolEntry = true
					break
				}
			}
			assert.True(t, hasProtocolEntry, "Expected protocol-only entry command: %s", expectedCmd)
		})
	}
}

func TestNATMasqueradeService_TypeConversions(t *testing.T) {
	service := &NATMasqueradeService{}

	t.Run("toParserNAT", func(t *testing.T) {
		nat := NATMasquerade{
			DescriptorID: 1,
			OuterAddress: "ipcp",
			InnerNetwork: "192.168.1.0-192.168.1.255",
			StaticEntries: []MasqueradeStaticEntry{
				{
					EntryNumber:       1,
					InsideLocal:       "192.168.1.100",
					InsideLocalPort:   intPtr(8080),
					OutsideGlobal:     "ipcp",
					OutsideGlobalPort: intPtr(80),
					Protocol:          "tcp",
				},
			},
		}

		parserNAT := service.toParserNAT(nat)

		assert.Equal(t, 1, parserNAT.DescriptorID)
		assert.Equal(t, "ipcp", parserNAT.OuterAddress)
		assert.Equal(t, "192.168.1.0-192.168.1.255", parserNAT.InnerNetwork)
		assert.Len(t, parserNAT.StaticEntries, 1)
		assert.Equal(t, 1, parserNAT.StaticEntries[0].EntryNumber)
		assert.Equal(t, "192.168.1.100", parserNAT.StaticEntries[0].InsideLocal)
		assert.NotNil(t, parserNAT.StaticEntries[0].InsideLocalPort)
		assert.Equal(t, 8080, *parserNAT.StaticEntries[0].InsideLocalPort)
		assert.Equal(t, "ipcp", parserNAT.StaticEntries[0].OutsideGlobal)
		assert.NotNil(t, parserNAT.StaticEntries[0].OutsideGlobalPort)
		assert.Equal(t, 80, *parserNAT.StaticEntries[0].OutsideGlobalPort)
		assert.Equal(t, "tcp", parserNAT.StaticEntries[0].Protocol)
	})
}
