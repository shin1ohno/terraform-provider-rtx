package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConfigPathResolver_Resolve_Japanese(t *testing.T) {
	tests := []struct {
		name         string
		mockSetup    func(*MockExecutor)
		expectedPath string
		expectedErr  bool
	}{
		{
			name: "Japanese output - config 0",
			mockSetup: func(m *MockExecutor) {
				output := `RTX1300 Rev.23.00.02 (Wed Feb  7 11:02:53 2024)
  ブートROM: 1.02
  CPU: (unknown) (800MHz)
  メモリ: 2048MByte(Free: 1605M Byte)
  ファームウェア: internal (original)
  デフォルト設定ファイル: config0
  リビジョンアップ実行の保留: なし
  起動時の設定ファイル: config0
  起動時刻: 2024/01/15 10:30:00 +09:00
  起動からの経過時間: 5日10時間30分15秒
  セキュリティクラス: 1
`
				m.On("Run", mock.Anything, "show environment").Return([]byte(output), nil)
			},
			expectedPath: "/system/config0",
			expectedErr:  false,
		},
		{
			name: "Japanese output - config 1",
			mockSetup: func(m *MockExecutor) {
				output := `RTX1300 Rev.23.00.02 (Wed Feb  7 11:02:53 2024)
  ブートROM: 1.02
  CPU: (unknown) (800MHz)
  メモリ: 2048MByte(Free: 1605M Byte)
  ファームウェア: internal (original)
  デフォルト設定ファイル: config1
  リビジョンアップ実行の保留: なし
  起動時の設定ファイル: config1
  起動時刻: 2024/01/15 10:30:00 +09:00
`
				m.On("Run", mock.Anything, "show environment").Return([]byte(output), nil)
			},
			expectedPath: "/system/config1",
			expectedErr:  false,
		},
		{
			name: "Japanese output - config 5 (higher number)",
			mockSetup: func(m *MockExecutor) {
				output := `RTX1300 Rev.23.00.02 (Wed Feb  7 11:02:53 2024)
  デフォルト設定ファイル: config5
  起動時刻: 2024/01/15 10:30:00 +09:00
`
				m.On("Run", mock.Anything, "show environment").Return([]byte(output), nil)
			},
			expectedPath: "/system/config5",
			expectedErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			resolver := NewConfigPathResolver(mockExecutor)
			path, err := resolver.Resolve(context.Background())

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPath, path)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestConfigPathResolver_Resolve_English(t *testing.T) {
	tests := []struct {
		name         string
		mockSetup    func(*MockExecutor)
		expectedPath string
		expectedErr  bool
	}{
		{
			name: "English output - config 0",
			mockSetup: func(m *MockExecutor) {
				output := `RTX1300 Rev.23.00.02 (Wed Feb  7 11:02:53 2024)
  Boot ROM: 1.02
  CPU: (unknown) (800MHz)
  Memory: 2048MByte(Free: 1605M Byte)
  Firmware: internal (original)
  Default config file: config0
  Pending revision update: none
  Boot time config file: config0
  Start time: 2024/01/15 10:30:00 +09:00
  Uptime: 5 days 10 hours 30 minutes 15 seconds
  Security class: 1
`
				m.On("Run", mock.Anything, "show environment").Return([]byte(output), nil)
			},
			expectedPath: "/system/config0",
			expectedErr:  false,
		},
		{
			name: "English output - config 2",
			mockSetup: func(m *MockExecutor) {
				output := `RTX1300 Rev.23.00.02 (Wed Feb  7 11:02:53 2024)
  Boot ROM: 1.02
  Default config file: config2
  Start time: 2024/01/15 10:30:00 +09:00
`
				m.On("Run", mock.Anything, "show environment").Return([]byte(output), nil)
			},
			expectedPath: "/system/config2",
			expectedErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			resolver := NewConfigPathResolver(mockExecutor)
			path, err := resolver.Resolve(context.Background())

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPath, path)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestConfigPathResolver_Resolve_Fallback(t *testing.T) {
	tests := []struct {
		name         string
		mockSetup    func(*MockExecutor)
		expectedPath string
		expectedErr  bool
	}{
		{
			name: "Command execution error - fallback to config0",
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "show environment").
					Return(nil, errors.New("connection failed"))
			},
			expectedPath: "/system/config0",
			expectedErr:  false,
		},
		{
			name: "Empty output - fallback to config0",
			mockSetup: func(m *MockExecutor) {
				m.On("Run", mock.Anything, "show environment").Return([]byte(""), nil)
			},
			expectedPath: "/system/config0",
			expectedErr:  false,
		},
		{
			name: "Missing config file field - fallback to config0",
			mockSetup: func(m *MockExecutor) {
				output := `RTX1300 Rev.23.00.02 (Wed Feb  7 11:02:53 2024)
  Boot ROM: 1.02
  CPU: (unknown) (800MHz)
  Memory: 2048MByte(Free: 1605M Byte)
`
				m.On("Run", mock.Anything, "show environment").Return([]byte(output), nil)
			},
			expectedPath: "/system/config0",
			expectedErr:  false,
		},
		{
			name: "Malformed config number - fallback to config0",
			mockSetup: func(m *MockExecutor) {
				output := `RTX1300 Rev.23.00.02 (Wed Feb  7 11:02:53 2024)
  デフォルト設定ファイル: configXYZ
`
				m.On("Run", mock.Anything, "show environment").Return([]byte(output), nil)
			},
			expectedPath: "/system/config0",
			expectedErr:  false,
		},
		{
			name: "Config field with no value - fallback to config0",
			mockSetup: func(m *MockExecutor) {
				output := `RTX1300 Rev.23.00.02 (Wed Feb  7 11:02:53 2024)
  Default config file:
`
				m.On("Run", mock.Anything, "show environment").Return([]byte(output), nil)
			},
			expectedPath: "/system/config0",
			expectedErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			resolver := NewConfigPathResolver(mockExecutor)
			path, err := resolver.Resolve(context.Background())

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPath, path)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestParseConfigNumber(t *testing.T) {
	tests := []struct {
		name           string
		output         string
		expectedNumber int
		expectedFound  bool
	}{
		{
			name:           "Japanese デフォルト設定ファイル config0",
			output:         "  デフォルト設定ファイル: config0\n",
			expectedNumber: 0,
			expectedFound:  true,
		},
		{
			name:           "Japanese デフォルト設定ファイル config3",
			output:         "  デフォルト設定ファイル: config3\n",
			expectedNumber: 3,
			expectedFound:  true,
		},
		{
			name:           "English Default config file config0",
			output:         "  Default config file: config0\n",
			expectedNumber: 0,
			expectedFound:  true,
		},
		{
			name:           "English Default config file config4",
			output:         "  Default config file: config4\n",
			expectedNumber: 4,
			expectedFound:  true,
		},
		{
			name:           "Mixed case output (should still work)",
			output:         "  default config file: config2\n",
			expectedNumber: 2,
			expectedFound:  true,
		},
		{
			name:           "Extra whitespace",
			output:         "  デフォルト設定ファイル:   config1  \n",
			expectedNumber: 1,
			expectedFound:  true,
		},
		{
			name:           "No match in output",
			output:         "  Some other line\n  Another line\n",
			expectedNumber: 0,
			expectedFound:  false,
		},
		{
			name:           "Empty string",
			output:         "",
			expectedNumber: 0,
			expectedFound:  false,
		},
		{
			name:           "Config with double digits",
			output:         "  デフォルト設定ファイル: config10\n",
			expectedNumber: 10,
			expectedFound:  true,
		},
		{
			name:           "Full show environment output (Japanese)",
			output:         "RTX1300 Rev.23.00.02 (Wed Feb  7 11:02:53 2024)\n  ブートROM: 1.02\n  CPU: (unknown) (800MHz)\n  メモリ: 2048MByte(Free: 1605M Byte)\n  ファームウェア: internal (original)\n  デフォルト設定ファイル: config0\n  リビジョンアップ実行の保留: なし\n  起動時の設定ファイル: config0\n  起動時刻: 2024/01/15 10:30:00 +09:00\n",
			expectedNumber: 0,
			expectedFound:  true,
		},
		{
			name:           "Full show environment output (English)",
			output:         "RTX1300 Rev.23.00.02 (Wed Feb  7 11:02:53 2024)\n  Boot ROM: 1.02\n  CPU: (unknown) (800MHz)\n  Memory: 2048MByte(Free: 1605M Byte)\n  Firmware: internal (original)\n  Default config file: config0\n  Pending revision update: none\n  Boot time config file: config0\n  Start time: 2024/01/15 10:30:00 +09:00\n",
			expectedNumber: 0,
			expectedFound:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			number, found := parseConfigNumber(tt.output)
			assert.Equal(t, tt.expectedFound, found, "found mismatch")
			if tt.expectedFound {
				assert.Equal(t, tt.expectedNumber, number, "number mismatch")
			}
		})
	}
}
