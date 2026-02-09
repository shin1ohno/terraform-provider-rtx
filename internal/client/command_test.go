package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSaveConfig(t *testing.T) {
	t.Run("nil client returns nil", func(t *testing.T) {
		err := saveConfig(context.Background(), nil, "operation succeeded")
		assert.NoError(t, err)
	})

	t.Run("non-nil client saves successfully", func(t *testing.T) {
		executor := new(MockExecutor)
		executor.On("Run", mock.Anything, "save").Return([]byte(""), nil)
		client := &rtxClient{executor: executor, active: true}

		err := saveConfig(context.Background(), client, "config applied")
		assert.NoError(t, err)
		executor.AssertExpectations(t)
	})

	t.Run("non-nil client save error includes operation description", func(t *testing.T) {
		executor := new(MockExecutor)
		executor.On("Run", mock.Anything, "save").Return(nil, errors.New("connection lost"))
		client := &rtxClient{executor: executor, active: true}

		err := saveConfig(context.Background(), client, "user created")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user created")
		assert.Contains(t, err.Error(), "failed to save configuration")
		executor.AssertExpectations(t)
	})
}

func TestCheckOutputError(t *testing.T) {
	tests := []struct {
		name          string
		output        []byte
		operationDesc string
		expectedErr   bool
		errContains   string
	}{
		{
			name:          "empty output returns nil",
			output:        []byte{},
			operationDesc: "command failed",
			expectedErr:   false,
		},
		{
			name:          "nil output returns nil",
			output:        nil,
			operationDesc: "command failed",
			expectedErr:   false,
		},
		{
			name:          "normal output returns nil",
			output:        []byte("timezone +09:00\n"),
			operationDesc: "command failed",
			expectedErr:   false,
		},
		{
			name:          "output with Error: returns error",
			output:        []byte("Error: invalid parameter"),
			operationDesc: "set timezone",
			expectedErr:   true,
			errContains:   "set timezone",
		},
		{
			name:          "output with Japanese error returns error",
			output:        []byte("エラー: 不正なパラメータ"),
			operationDesc: "command failed",
			expectedErr:   true,
			errContains:   "command failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkOutputError(tt.output, tt.operationDesc)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckOutputErrorIgnoringNotFound(t *testing.T) {
	tests := []struct {
		name        string
		output      []byte
		expectedErr bool
	}{
		{
			name:        "empty output returns nil",
			output:      []byte{},
			expectedErr: false,
		},
		{
			name:        "not found error is ignored",
			output:      []byte("Error: not found"),
			expectedErr: false,
		},
		{
			name:        "Not Found with mixed case is ignored",
			output:      []byte("Error: entry Not Found in table"),
			expectedErr: false,
		},
		{
			name:        "other error is returned",
			output:      []byte("Error: invalid parameter"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkOutputErrorIgnoringNotFound(tt.output, "command failed")
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunCommand(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		setupMock   func(*MockExecutor)
		expectedErr bool
		errContains string
	}{
		{
			name: "successful command",
			cmd:  "timezone +09:00",
			setupMock: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "timezone +09:00").
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "executor error propagated",
			cmd:  "timezone +09:00",
			setupMock: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "timezone +09:00").
					Return(nil, errors.New("connection timeout"))
			},
			expectedErr: true,
			errContains: "connection timeout",
		},
		{
			name: "output error detected",
			cmd:  "invalid-command",
			setupMock: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "invalid-command").
					Return([]byte("Error: invalid parameter"), nil)
			},
			expectedErr: true,
			errContains: "command failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := new(MockExecutor)
			tt.setupMock(executor)

			err := runCommand(context.Background(), executor, tt.cmd)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			executor.AssertExpectations(t)
		})
	}
}

func TestRunBatchCommands(t *testing.T) {
	tests := []struct {
		name        string
		commands    []string
		setupMock   func(*MockExecutor)
		expectedErr bool
		errContains string
	}{
		{
			name:        "empty commands returns nil without executing",
			commands:    []string{},
			setupMock:   func(m *MockExecutor) {},
			expectedErr: false,
		},
		{
			name:     "successful batch",
			commands: []string{"timezone +09:00", "console character ascii"},
			setupMock: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, []string{"timezone +09:00", "console character ascii"}).
					Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name:     "executor error propagated",
			commands: []string{"timezone +09:00"},
			setupMock: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, []string{"timezone +09:00"}).
					Return(nil, errors.New("batch execution failed"))
			},
			expectedErr: true,
			errContains: "batch execution failed",
		},
		{
			name:     "output error detected",
			commands: []string{"invalid-command"},
			setupMock: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, []string{"invalid-command"}).
					Return([]byte("Error: invalid parameter"), nil)
			},
			expectedErr: true,
			errContains: "batch command failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := new(MockExecutor)
			tt.setupMock(executor)

			err := runBatchCommands(context.Background(), executor, tt.commands)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			executor.AssertExpectations(t)
		})
	}
}
