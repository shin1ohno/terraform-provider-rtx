package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPPTPService_Get(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*MockExecutor)
		expected    *PPTPConfig
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful get with basic config",
			mockSetup: func(m *MockExecutor) {
				output := `pptp service on
pp auth accept mschap-v2
pp auth myname user password
ppp ccp type mppe-any
`
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep pptp` || cmd == `show config | grep "pp auth"` || cmd == `show config | grep ppp`
				})).Return([]byte(output), nil)
			},
			expected: &PPTPConfig{
				Enabled: true,
				Authentication: &PPTPAuth{
					Method:   "mschap-v2",
					Username: "user",
					Password: "password",
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

			service := &PPTPService{executor: mockExecutor}
			result, err := service.Get(context.Background())

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Enabled, result.Enabled)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestPPTPService_Create(t *testing.T) {
	tests := []struct {
		name        string
		config      PPTPConfig
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful creation with batch",
			config: PPTPConfig{
				Enabled: true,
				Authentication: &PPTPAuth{
					Method:   "mschap-v2",
					Username: "testuser",
					Password: "testpass",
				},
				Encryption: &PPTPEncryption{
					MPPEBits: 128,
					Required: true,
				},
				IPPool: &PPTPIPPool{
					Start: "192.168.10.100",
					End:   "192.168.10.200",
				},
				KeepaliveEnabled: true,
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasService := false
					hasAuth := false
					hasMyname := false
					for _, cmd := range cmds {
						if cmd == "pptp service on" {
							hasService = true
						}
						if cmd == "pp auth accept mschap-v2" {
							hasAuth = true
						}
						if cmd == "pp auth myname testuser testpass" {
							hasMyname = true
						}
					}
					return hasService && hasAuth && hasMyname
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Batch execution error",
			config: PPTPConfig{
				Enabled: true,
				Authentication: &PPTPAuth{
					Method:   "mschap-v2",
					Username: "testuser",
					Password: "testpass",
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

			// Create service without client to skip SaveConfig
			service := &PPTPService{executor: mockExecutor, client: nil}
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

func TestPPTPService_Update(t *testing.T) {
	tests := []struct {
		name        string
		config      PPTPConfig
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful update with batch",
			config: PPTPConfig{
				Enabled: true,
				Authentication: &PPTPAuth{
					Method:   "mschap-v2",
					Username: "newuser",
					Password: "newpass",
				},
				KeepaliveEnabled: false,
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasMyname := false
					for _, cmd := range cmds {
						if cmd == "pp auth myname newuser newpass" {
							hasMyname = true
						}
					}
					return hasMyname && len(cmds) > 0
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			// Create service without client to skip SaveConfig
			service := &PPTPService{executor: mockExecutor, client: nil}
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

func TestPPTPService_Delete(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful delete with batch",
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasServiceOff := false
					for _, cmd := range cmds {
						if cmd == "pptp service off" || cmd == "no pptp service" {
							hasServiceOff = true
						}
					}
					return hasServiceOff
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Execution error",
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

			// Create service without client to skip SaveConfig
			service := &PPTPService{executor: mockExecutor, client: nil}
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

func TestPPTPService_UsesRunBatch(t *testing.T) {
	t.Run("Create uses RunBatch for all commands", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		var capturedCommands []string
		mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				capturedCommands = args.Get(1).([]string)
			}).
			Return([]byte(""), nil)

		// Create service without client to skip SaveConfig
		service := &PPTPService{executor: mockExecutor, client: nil}
		err := service.Create(context.Background(), PPTPConfig{
			Enabled: true,
			Authentication: &PPTPAuth{
				Method:   "mschap-v2",
				Username: "user",
				Password: "pass",
			},
			KeepaliveEnabled: true,
		})

		assert.NoError(t, err)

		// Verify essential commands are present
		hasService := false
		hasAuth := false
		for _, cmd := range capturedCommands {
			if cmd == "pptp service on" {
				hasService = true
			}
			if cmd == "pp auth accept mschap-v2" {
				hasAuth = true
			}
		}
		assert.True(t, hasService, "Expected service command to be included")
		assert.True(t, hasAuth, "Expected auth command to be included")
	})

	t.Run("Update uses RunBatch for all commands", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		var capturedCommands []string
		mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				capturedCommands = args.Get(1).([]string)
			}).
			Return([]byte(""), nil)

		// Create service without client to skip SaveConfig
		service := &PPTPService{executor: mockExecutor, client: nil}
		err := service.Update(context.Background(), PPTPConfig{
			Authentication: &PPTPAuth{
				Method:   "mschap-v2",
				Username: "newuser",
				Password: "newpass",
			},
		})

		assert.NoError(t, err)
		assert.Greater(t, len(capturedCommands), 0, "Expected commands to be captured")
	})
}

func TestPPTPService_ContextCancellation(t *testing.T) {
	t.Run("Get with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		mockExecutor := new(MockExecutor)
		mockExecutor.On("Run", mock.Anything, mock.Anything).
			Return(nil, context.Canceled)

		service := &PPTPService{executor: mockExecutor}
		_, err := service.Get(ctx)

		assert.Error(t, err)
	})

	t.Run("Create with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		mockExecutor := new(MockExecutor)
		mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
			Return(nil, context.Canceled)

		// Create service without client to skip SaveConfig
		service := &PPTPService{executor: mockExecutor, client: nil}
		err := service.Create(ctx, PPTPConfig{
			Enabled: true,
			Authentication: &PPTPAuth{
				Method:   "mschap-v2",
				Username: "user",
				Password: "pass",
			},
		})

		assert.Error(t, err)
	})
}
