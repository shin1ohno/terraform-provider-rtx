package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBGPService_Get(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*MockExecutor)
		expected    *BGPConfig
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful get with basic config",
			mockSetup: func(m *MockExecutor) {
				output := `bgp use on
bgp autonomous-system 65000
bgp router id 192.168.1.1
bgp neighbor 1 65001 192.168.1.2
`
				m.On("Run", mock.Anything, `show config | grep bgp`).
					Return([]byte(output), nil)
			},
			expected: &BGPConfig{
				Enabled:  true,
				ASN:      "65000",
				RouterID: "192.168.1.1",
				Neighbors: []BGPNeighbor{
					{
						ID:       1,
						IP:       "192.168.1.2",
						RemoteAS: "65001",
					},
				},
				Networks: []BGPNetwork{},
			},
			expectedErr: false,
		},
		{
			name: "Successful get with networks and redistribution",
			mockSetup: func(m *MockExecutor) {
				output := `bgp use on
bgp autonomous-system 65000
bgp router id 192.168.1.1
bgp neighbor 1 65001 192.168.1.2
bgp network 1 192.168.0.0/24
bgp import filter 1 include static
bgp import filter 2 include connected
`
				m.On("Run", mock.Anything, `show config | grep bgp`).
					Return([]byte(output), nil)
			},
			expected: &BGPConfig{
				Enabled:               true,
				ASN:                   "65000",
				RouterID:              "192.168.1.1",
				RedistributeStatic:    true,
				RedistributeConnected: true,
				Neighbors: []BGPNeighbor{
					{
						ID:       1,
						IP:       "192.168.1.2",
						RemoteAS: "65001",
					},
				},
				Networks: []BGPNetwork{
					{
						Prefix: "192.168.0.0",
						Mask:   "255.255.255.0",
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

			service := &BGPService{executor: mockExecutor}
			result, err := service.Get(context.Background())

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ASN, result.ASN)
				assert.Equal(t, tt.expected.RouterID, result.RouterID)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestBGPService_Configure(t *testing.T) {
	tests := []struct {
		name        string
		config      BGPConfig
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful basic configuration with batch",
			config: BGPConfig{
				ASN:      "65000",
				RouterID: "192.168.1.1",
				Neighbors: []BGPNeighbor{
					{
						ID:       1,
						IP:       "192.168.1.2",
						RemoteAS: "65001",
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				// Expected batch commands including ASN, router ID, neighbor, and enable
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasASN := false
					hasNeighbor := false
					hasEnable := false
					for _, cmd := range cmds {
						if cmd == "bgp autonomous-system 65000" {
							hasASN = true
						}
						if cmd == "bgp neighbor 1 address 192.168.1.2 as 65001" {
							hasNeighbor = true
						}
						if cmd == "bgp use on" {
							hasEnable = true
						}
					}
					return hasASN && hasNeighbor && hasEnable && len(cmds) > 0
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Successful configuration with redistribution",
			config: BGPConfig{
				ASN:                   "65000",
				RouterID:              "192.168.1.1",
				RedistributeStatic:    true,
				RedistributeConnected: true,
				Neighbors: []BGPNeighbor{
					{
						ID:       1,
						IP:       "192.168.1.2",
						RemoteAS: "65001",
					},
				},
				Networks: []BGPNetwork{
					{
						Prefix: "10.0.0.0",
						Mask:   "255.0.0.0",
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasRedistributeStatic := false
					hasRedistributeConnected := false
					for _, cmd := range cmds {
						if cmd == "bgp import from static" {
							hasRedistributeStatic = true
						}
						if cmd == "bgp import from connected" {
							hasRedistributeConnected = true
						}
					}
					return hasRedistributeStatic && hasRedistributeConnected
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Validation error - missing ASN",
			config: BGPConfig{
				ASN:      "",
				RouterID: "192.168.1.1",
			},
			mockSetup:   func(m *MockExecutor) {},
			expectedErr: true,
			errMessage:  "invalid BGP config",
		},
		{
			name: "Batch execution error",
			config: BGPConfig{
				ASN:      "65000",
				RouterID: "192.168.1.1",
				Neighbors: []BGPNeighbor{
					{
						ID:       1,
						IP:       "192.168.1.2",
						RemoteAS: "65001",
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

			service := &BGPService{executor: mockExecutor}
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

func TestBGPService_Update(t *testing.T) {
	tests := []struct {
		name        string
		config      BGPConfig
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful update with new neighbor",
			config: BGPConfig{
				ASN:      "65000",
				RouterID: "192.168.1.1",
				Neighbors: []BGPNeighbor{
					{
						ID:       1,
						IP:       "192.168.1.2",
						RemoteAS: "65001",
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				// First: Get current config (no neighbors)
				currentOutput := `bgp use on
bgp autonomous-system 65000
bgp router id 192.168.1.1
`
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep bgp`
				})).Return([]byte(currentOutput), nil)

				// Then: RunBatch with neighbor commands
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasNeighbor := false
					for _, cmd := range cmds {
						if cmd == "bgp neighbor 1 address 192.168.1.2 as 65001" {
							hasNeighbor = true
						}
					}
					return hasNeighbor && len(cmds) > 0
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Update changes ASN",
			config: BGPConfig{
				ASN:      "65001",
				RouterID: "192.168.1.1",
				Neighbors: []BGPNeighbor{
					{
						ID:       1,
						IP:       "192.168.1.2",
						RemoteAS: "65000",
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				// Get current config with different ASN
				currentOutput := `bgp use on
bgp autonomous-system 65000
bgp router id 192.168.1.1
`
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep bgp`
				})).Return([]byte(currentOutput), nil)

				// RunBatch should include ASN change and neighbor update
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasASN := false
					for _, cmd := range cmds {
						if cmd == "bgp autonomous-system 65001" {
							hasASN = true
						}
					}
					return hasASN && len(cmds) > 0
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Validation error",
			config: BGPConfig{
				ASN:      "",
				RouterID: "192.168.1.1",
			},
			mockSetup: func(m *MockExecutor) {
				// Update first calls Get, so we need to mock that
				currentOutput := `bgp use on
bgp autonomous-system 65000
bgp router id 192.168.1.1
`
				m.On("Run", mock.Anything, `show config | grep bgp`).
					Return([]byte(currentOutput), nil)
			},
			expectedErr: true,
			errMessage:  "invalid BGP config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &BGPService{executor: mockExecutor}
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

func TestBGPService_Reset(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful reset",
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, []string{"bgp use off"}).
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Execution error",
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, []string{"bgp use off"}).
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

			service := &BGPService{executor: mockExecutor}
			err := service.Reset(context.Background())

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

func TestBGPService_ContextCancellation(t *testing.T) {
	t.Run("Get with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		mockExecutor := new(MockExecutor)
		mockExecutor.On("Run", mock.Anything, mock.Anything).
			Return(nil, context.Canceled)

		service := &BGPService{executor: mockExecutor}
		_, err := service.Get(ctx)

		assert.Error(t, err)
	})

	t.Run("Configure with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		mockExecutor := new(MockExecutor)
		// RunBatch returns context.Canceled
		mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
			Return(nil, context.Canceled)

		service := &BGPService{executor: mockExecutor}
		err := service.Configure(ctx, BGPConfig{
			ASN:      "65000",
			RouterID: "192.168.1.1",
			Neighbors: []BGPNeighbor{
				{ID: 1, IP: "192.168.1.2", RemoteAS: "65001"},
			},
		})

		assert.Error(t, err)
	})
}

func TestBGPService_UsesRunBatch(t *testing.T) {
	t.Run("Configure uses RunBatch for all commands", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		// Capture commands to verify RunBatch is used
		var capturedCommands []string
		mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				capturedCommands = args.Get(1).([]string)
			}).
			Return([]byte(""), nil)

		service := &BGPService{executor: mockExecutor}
		err := service.Configure(context.Background(), BGPConfig{
			ASN:      "65000",
			RouterID: "192.168.1.1",
			Neighbors: []BGPNeighbor{
				{ID: 1, IP: "192.168.1.2", RemoteAS: "65001"},
			},
		})

		assert.NoError(t, err)

		// Verify essential commands are present
		hasASN := false
		hasNeighbor := false
		hasEnable := false
		for _, cmd := range capturedCommands {
			if cmd == "bgp autonomous-system 65000" {
				hasASN = true
			}
			if cmd == "bgp neighbor 1 address 192.168.1.2 as 65001" {
				hasNeighbor = true
			}
			if cmd == "bgp use on" {
				hasEnable = true
			}
		}
		assert.True(t, hasASN, "Expected ASN command to be included")
		assert.True(t, hasNeighbor, "Expected neighbor command to be included")
		assert.True(t, hasEnable, "Expected enable command to be included")
	})

	t.Run("Update uses RunBatch for all commands", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		// Get current config (without the neighbor)
		currentOutput := `bgp use on
bgp autonomous-system 65000
bgp router id 192.168.1.1
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

		service := &BGPService{executor: mockExecutor}
		err := service.Update(context.Background(), BGPConfig{
			ASN:      "65000",
			RouterID: "192.168.1.1",
			Neighbors: []BGPNeighbor{
				{ID: 1, IP: "192.168.1.2", RemoteAS: "65001"},
			},
		})

		assert.NoError(t, err)

		// Verify neighbor command is present (since it was added)
		hasNeighbor := false
		for _, cmd := range capturedCommands {
			if cmd == "bgp neighbor 1 address 192.168.1.2 as 65001" {
				hasNeighbor = true
				break
			}
		}
		assert.True(t, hasNeighbor, "Expected neighbor command to be included in update")
	})
}
