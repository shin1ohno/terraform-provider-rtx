package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOSPFService_Get(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*MockExecutor)
		expected    *OSPFConfig
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful get with basic config",
			mockSetup: func(m *MockExecutor) {
				output := `ospf use on
ospf router id 192.168.1.1
ospf area 0.0.0.0
`
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep ospf`
				})).Return([]byte(output), nil)
			},
			expected: &OSPFConfig{
				Enabled:  true,
				RouterID: "192.168.1.1",
				Areas: []OSPFArea{
					{
						ID: "0.0.0.0",
					},
				},
			},
			expectedErr: false,
		},
		{
			name: "Execution error",
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

			service := &OSPFService{executor: mockExecutor}
			result, err := service.Get(context.Background())

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.RouterID, result.RouterID)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestOSPFService_Create(t *testing.T) {
	tests := []struct {
		name        string
		config      OSPFConfig
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful basic configuration with batch",
			config: OSPFConfig{
				RouterID: "192.168.1.1",
				Areas: []OSPFArea{
					{
						ID:   "0.0.0.0",
						Type: "normal",
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				// Expected batch commands including router ID, area, and use on
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasRouterID := false
					hasArea := false
					hasEnable := false
					for _, cmd := range cmds {
						if cmd == "ospf router id 192.168.1.1" {
							hasRouterID = true
						}
						if cmd == "ospf area 0.0.0.0" {
							hasArea = true
						}
						if cmd == "ospf use on" {
							hasEnable = true
						}
					}
					return hasRouterID && hasArea && hasEnable && len(cmds) > 0
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Successful configuration with redistribution",
			config: OSPFConfig{
				RouterID:              "192.168.1.1",
				RedistributeStatic:    true,
				RedistributeConnected: true,
				Areas: []OSPFArea{
					{
						ID:   "0.0.0.0",
						Type: "normal",
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasImportStatic := false
					hasImportConnected := false
					for _, cmd := range cmds {
						if cmd == "ospf import from static" {
							hasImportStatic = true
						}
						if cmd == "ospf import from connected" {
							hasImportConnected = true
						}
					}
					return hasImportStatic && hasImportConnected
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Validation error - missing router ID",
			config: OSPFConfig{
				RouterID: "",
			},
			mockSetup:   func(m *MockExecutor) {},
			expectedErr: true,
			errMessage:  "invalid OSPF config",
		},
		{
			name: "Batch execution error",
			config: OSPFConfig{
				RouterID: "192.168.1.1",
				Areas: []OSPFArea{
					{
						ID:   "0.0.0.0",
						Type: "normal",
					},
				},
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

			service := &OSPFService{executor: mockExecutor}
			err := service.Create(context.Background(), tt.config)

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

func TestOSPFService_Update(t *testing.T) {
	tests := []struct {
		name        string
		config      OSPFConfig
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful update with batch",
			config: OSPFConfig{
				RouterID:           "192.168.1.2",
				RedistributeStatic: true,
				Areas: []OSPFArea{
					{
						ID:   "0.0.0.0",
						Type: "normal",
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				// First: Get current config
				currentOutput := `ospf use on
ospf router id 192.168.1.1
ospf area 0.0.0.0
`
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep ospf`
				})).Return([]byte(currentOutput), nil)

				// Then: RunBatch with updated commands
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasRouterID := false
					hasImportStatic := false
					for _, cmd := range cmds {
						if cmd == "ospf router id 192.168.1.2" {
							hasRouterID = true
						}
						if cmd == "ospf import from static" {
							hasImportStatic = true
						}
					}
					return hasRouterID && hasImportStatic && len(cmds) > 0
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Validation error",
			config: OSPFConfig{
				RouterID: "",
			},
			mockSetup: func(m *MockExecutor) {
				// Update first calls Get
				currentOutput := `ospf use on
ospf router id 192.168.1.1
`
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep ospf`
				})).Return([]byte(currentOutput), nil)
			},
			expectedErr: true,
			errMessage:  "invalid OSPF config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &OSPFService{executor: mockExecutor}
			err := service.Update(context.Background(), tt.config)

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

func TestOSPFService_Delete(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful delete",
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, []string{"ospf use off"}).
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Execution error",
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, []string{"ospf use off"}).
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

			service := &OSPFService{executor: mockExecutor}
			err := service.Delete(context.Background())

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

func TestOSPFService_UsesRunBatch(t *testing.T) {
	t.Run("Create uses RunBatch for all commands", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		// Capture commands to verify RunBatch is used
		var capturedCommands []string
		mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				capturedCommands = args.Get(1).([]string)
			}).
			Return([]byte(""), nil)

		service := &OSPFService{executor: mockExecutor}
		err := service.Create(context.Background(), OSPFConfig{
			RouterID: "192.168.1.1",
			Areas: []OSPFArea{
				{ID: "0.0.0.0", Type: "normal"},
			},
		})

		assert.NoError(t, err)

		// Verify essential commands are present
		hasRouterID := false
		hasArea := false
		hasEnable := false
		for _, cmd := range capturedCommands {
			if cmd == "ospf router id 192.168.1.1" {
				hasRouterID = true
			}
			if cmd == "ospf area 0.0.0.0" {
				hasArea = true
			}
			if cmd == "ospf use on" {
				hasEnable = true
			}
		}
		assert.True(t, hasRouterID, "Expected router ID command to be included")
		assert.True(t, hasArea, "Expected area command to be included")
		assert.True(t, hasEnable, "Expected enable command to be included")
	})

	t.Run("Update uses RunBatch for all commands", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		// Get current config (different router ID)
		currentOutput := `ospf use on
ospf router id 192.168.1.2
`
		mockExecutor.On("Run", mock.Anything, mock.Anything).
			Return([]byte(currentOutput), nil)

		// Capture commands to verify RunBatch is used
		var capturedCommands []string
		mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				capturedCommands = args.Get(1).([]string)
			}).
			Return([]byte(""), nil)

		service := &OSPFService{executor: mockExecutor}
		err := service.Update(context.Background(), OSPFConfig{
			RouterID: "192.168.1.1",
			Areas: []OSPFArea{
				{ID: "0.0.0.0", Type: "normal"},
			},
		})

		assert.NoError(t, err)

		// Verify router ID change command is present
		hasRouterID := false
		for _, cmd := range capturedCommands {
			if cmd == "ospf router id 192.168.1.1" {
				hasRouterID = true
				break
			}
		}
		assert.True(t, hasRouterID, "Expected router ID command to be included in update")
	})
}
