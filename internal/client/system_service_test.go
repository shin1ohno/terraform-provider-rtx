package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSystemService_Get(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*MockExecutor)
		expected    *SystemConfig
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful get with basic config",
			mockSetup: func(m *MockExecutor) {
				output := `timezone +09:00
console character ascii
console lines 24
`
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep "(timezone|console|packet-buffer|statistics)"`
				})).Return([]byte(output), nil)
			},
			expected: &SystemConfig{
				Timezone: "+09:00",
				Console: &ConsoleConfig{
					Character: "ascii",
					Lines:     "24",
				},
			},
			expectedErr: false,
		},
		{
			name: "Successful get with packet buffers",
			mockSetup: func(m *MockExecutor) {
				output := `timezone +09:00
system packet-buffer small max-buffer=2048 max-free=256
system packet-buffer medium max-buffer=512 max-free=1024
`
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep "(timezone|console|packet-buffer|statistics)"`
				})).Return([]byte(output), nil)
			},
			expected: &SystemConfig{
				Timezone: "+09:00",
				PacketBuffers: []PacketBufferConfig{
					{Size: "small", MaxBuffer: 2048, MaxFree: 256},
					{Size: "medium", MaxBuffer: 512, MaxFree: 1024},
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

			service := &SystemService{executor: mockExecutor}
			result, err := service.Get(context.Background())

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Timezone, result.Timezone)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestSystemService_Configure(t *testing.T) {
	tests := []struct {
		name        string
		config      SystemConfig
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful configuration with batch",
			config: SystemConfig{
				Timezone: "+09:00",
				Console: &ConsoleConfig{
					Character: "ascii",
					Lines:     "24",
					Prompt:    "router",
				},
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasTimezone := false
					hasConsoleChar := false
					hasConsoleLines := false
					hasConsolePrompt := false
					for _, cmd := range cmds {
						if cmd == "timezone +09:00" {
							hasTimezone = true
						}
						if cmd == "console character ascii" {
							hasConsoleChar = true
						}
						if cmd == "console lines 24" {
							hasConsoleLines = true
						}
						if cmd == "console prompt router" {
							hasConsolePrompt = true
						}
					}
					return hasTimezone && hasConsoleChar && hasConsoleLines && hasConsolePrompt
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Successful configuration with packet buffers",
			config: SystemConfig{
				Timezone: "+09:00",
				PacketBuffers: []PacketBufferConfig{
					{Size: "small", MaxBuffer: 2048, MaxFree: 256},
				},
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasPacketBuffer := false
					for _, cmd := range cmds {
						if cmd == "system packet-buffer small max-buffer=2048 max-free=256" {
							hasPacketBuffer = true
						}
					}
					return hasPacketBuffer
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Successful configuration with statistics",
			config: SystemConfig{
				Timezone: "+09:00",
				Statistics: &StatisticsConfig{
					Traffic: true,
					NAT:     true,
				},
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasTraffic := false
					hasNAT := false
					for _, cmd := range cmds {
						if cmd == "statistics traffic on" {
							hasTraffic = true
						}
						if cmd == "statistics nat on" {
							hasNAT = true
						}
					}
					return hasTraffic && hasNAT
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Batch execution error",
			config: SystemConfig{
				Timezone: "+09:00",
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

			service := &SystemService{executor: mockExecutor}
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

func TestSystemService_Update(t *testing.T) {
	tests := []struct {
		name        string
		config      SystemConfig
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful update with batch",
			config: SystemConfig{
				Timezone: "+00:00",
				Console: &ConsoleConfig{
					Character: "ja.utf8",
				},
			},
			mockSetup: func(m *MockExecutor) {
				// First: Get current config
				currentOutput := `timezone +09:00
console character ascii
`
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep "(timezone|console|packet-buffer|statistics)"`
				})).Return([]byte(currentOutput), nil)

				// Then: RunBatch with updated commands
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasTimezoneUpdate := false
					hasConsoleCharUpdate := false
					for _, cmd := range cmds {
						if cmd == "timezone +00:00" {
							hasTimezoneUpdate = true
						}
						if cmd == "console character ja.utf8" {
							hasConsoleCharUpdate = true
						}
					}
					return hasTimezoneUpdate && hasConsoleCharUpdate
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &SystemService{executor: mockExecutor}
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

func TestSystemService_Reset(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful reset with batch",
			mockSetup: func(m *MockExecutor) {
				// First: Get current config
				currentOutput := `timezone +09:00
console character ascii
`
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep "(timezone|console|packet-buffer|statistics)"`
				})).Return([]byte(currentOutput), nil)

				// Then: RunBatch with delete commands
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasNoTimezone := false
					hasNoConsole := false
					for _, cmd := range cmds {
						if cmd == "no timezone" {
							hasNoTimezone = true
						}
						if cmd == "no console character" {
							hasNoConsole = true
						}
					}
					return hasNoTimezone || hasNoConsole
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Reset continues even on RunBatch error",
			mockSetup: func(m *MockExecutor) {
				currentOutput := `timezone +09:00
`
				m.On("Run", mock.Anything, mock.Anything).
					Return([]byte(currentOutput), nil)

				// Reset method logs the error but continues
				m.On("RunBatch", mock.Anything, mock.Anything).
					Return(nil, errors.New("connection failed"))
			},
			expectedErr: false, // Reset swallows RunBatch errors and continues
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &SystemService{executor: mockExecutor}
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

func TestSystemService_UsesRunBatch(t *testing.T) {
	t.Run("Configure uses RunBatch for all commands", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		var capturedCommands []string
		mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				capturedCommands = args.Get(1).([]string)
			}).
			Return([]byte(""), nil)

		service := &SystemService{executor: mockExecutor}
		err := service.Configure(context.Background(), SystemConfig{
			Timezone: "+09:00",
			Console: &ConsoleConfig{
				Character: "ascii",
				Lines:     "24",
			},
			PacketBuffers: []PacketBufferConfig{
				{Size: "small", MaxBuffer: 2048, MaxFree: 256},
			},
		})

		assert.NoError(t, err)

		// Verify essential commands are present
		hasTimezone := false
		hasConsoleChar := false
		hasConsoleLines := false
		hasPacketBuffer := false
		for _, cmd := range capturedCommands {
			if cmd == "timezone +09:00" {
				hasTimezone = true
			}
			if cmd == "console character ascii" {
				hasConsoleChar = true
			}
			if cmd == "console lines 24" {
				hasConsoleLines = true
			}
			if cmd == "system packet-buffer small max-buffer=2048 max-free=256" {
				hasPacketBuffer = true
			}
		}
		assert.True(t, hasTimezone, "Expected timezone command to be included")
		assert.True(t, hasConsoleChar, "Expected console character command to be included")
		assert.True(t, hasConsoleLines, "Expected console lines command to be included")
		assert.True(t, hasPacketBuffer, "Expected packet buffer command to be included")
	})

	t.Run("Update uses RunBatch for all commands", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		// Get current config
		currentOutput := `timezone +09:00
`
		mockExecutor.On("Run", mock.Anything, mock.Anything).
			Return([]byte(currentOutput), nil)

		var capturedCommands []string
		mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				capturedCommands = args.Get(1).([]string)
			}).
			Return([]byte(""), nil)

		service := &SystemService{executor: mockExecutor}
		err := service.Update(context.Background(), SystemConfig{
			Timezone: "+00:00",
		})

		assert.NoError(t, err)

		// Verify timezone update command is present
		hasTimezone := false
		for _, cmd := range capturedCommands {
			if cmd == "timezone +00:00" {
				hasTimezone = true
				break
			}
		}
		assert.True(t, hasTimezone, "Expected timezone update command to be included")
	})
}

func TestSystemService_ContextCancellation(t *testing.T) {
	t.Run("Get with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		mockExecutor := new(MockExecutor)
		mockExecutor.On("Run", mock.Anything, mock.Anything).
			Return(nil, context.Canceled)

		service := &SystemService{executor: mockExecutor}
		_, err := service.Get(ctx)

		assert.Error(t, err)
	})

	t.Run("Configure with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		mockExecutor := new(MockExecutor)
		mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
			Return(nil, context.Canceled)

		service := &SystemService{executor: mockExecutor}
		err := service.Configure(ctx, SystemConfig{
			Timezone: "+09:00",
		})

		assert.Error(t, err)
	})
}
