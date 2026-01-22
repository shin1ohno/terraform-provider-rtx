package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIPsecTunnelService_Get(t *testing.T) {
	tests := []struct {
		name        string
		tunnelID    int
		mockSetup   func(*MockExecutor)
		expected    *IPsecTunnel
		expectedErr bool
		errMessage  string
	}{
		{
			name:     "Successful get",
			tunnelID: 1,
			mockSetup: func(m *MockExecutor) {
				output := `tunnel select 1
 ipsec tunnel 1
 ipsec ike local address 1 192.168.1.1
 ipsec ike remote address 1 192.168.2.1
 ipsec ike pre-shared-key 1 text secret123
 ipsec ike keepalive use 1 on dpd 30 3
`
				m.On("Run", mock.Anything, mock.MatchedBy(func(cmd string) bool {
					return cmd == `show config | grep ipsec`
				})).Return([]byte(output), nil)
			},
			expected: &IPsecTunnel{
				ID:            1,
				LocalAddress:  "192.168.1.1",
				RemoteAddress: "192.168.2.1",
				DPDEnabled:    true,
				DPDInterval:   30,
				DPDRetry:      3,
			},
			expectedErr: false,
		},
		{
			name:     "Tunnel not found",
			tunnelID: 99,
			mockSetup: func(m *MockExecutor) {
				output := `tunnel select 1
 ipsec tunnel 1
`
				m.On("Run", mock.Anything, mock.Anything).Return([]byte(output), nil)
			},
			expected:    nil,
			expectedErr: true,
			errMessage:  "not found",
		},
		{
			name:     "Execution error",
			tunnelID: 1,
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

			service := &IPsecTunnelService{executor: mockExecutor}
			result, err := service.Get(context.Background(), tt.tunnelID)

			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ID, result.ID)
				assert.Equal(t, tt.expected.LocalAddress, result.LocalAddress)
				assert.Equal(t, tt.expected.RemoteAddress, result.RemoteAddress)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestIPsecTunnelService_Create(t *testing.T) {
	tests := []struct {
		name        string
		tunnel      IPsecTunnel
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful creation with batch",
			tunnel: IPsecTunnel{
				ID:            1,
				LocalAddress:  "192.168.1.1",
				RemoteAddress: "192.168.2.1",
				PreSharedKey:  "secret123",
				DPDEnabled:    true,
				DPDInterval:   30,
				DPDRetry:      3,
				IKEv2Proposal: IKEv2Proposal{
					EncryptionAES256: true,
					IntegritySHA256:  true,
					GroupFourteen:    true,
				},
				IPsecTransform: IPsecTransform{
					Protocol:         "esp",
					EncryptionAES256: true,
					IntegritySHA256:  true,
				},
			},
			mockSetup: func(m *MockExecutor) {
				// Expect RunBatch to be called with all configuration commands
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					// Verify commands include tunnel select and ipsec tunnel
					hasTunnelSelect := false
					hasIPsecTunnel := false
					hasLocalAddr := false
					hasRemoteAddr := false
					hasPSK := false
					for _, cmd := range cmds {
						if cmd == "tunnel select 1" {
							hasTunnelSelect = true
						}
						if cmd == "ipsec tunnel 1" {
							hasIPsecTunnel = true
						}
						if cmd == "ipsec ike local address 1 192.168.1.1" {
							hasLocalAddr = true
						}
						if cmd == "ipsec ike remote address 1 192.168.2.1" {
							hasRemoteAddr = true
						}
						if cmd == "ipsec ike pre-shared-key 1 text secret123" {
							hasPSK = true
						}
					}
					return hasTunnelSelect && hasIPsecTunnel && hasLocalAddr && hasRemoteAddr && hasPSK
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Validation error - missing local address",
			tunnel: IPsecTunnel{
				ID:            1,
				LocalAddress:  "",
				RemoteAddress: "192.168.2.1",
				PreSharedKey:  "secret123",
			},
			mockSetup:   func(m *MockExecutor) {},
			expectedErr: true,
			errMessage:  "invalid IPsec tunnel config",
		},
		{
			name: "Batch execution error",
			tunnel: IPsecTunnel{
				ID:            1,
				LocalAddress:  "192.168.1.1",
				RemoteAddress: "192.168.2.1",
				PreSharedKey:  "secret123",
				IKEv2Proposal: IKEv2Proposal{
					EncryptionAES256: true,
					IntegritySHA256:  true,
					GroupFourteen:    true,
				},
				IPsecTransform: IPsecTransform{
					Protocol:         "esp",
					EncryptionAES256: true,
					IntegritySHA256:  true,
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

			service := &IPsecTunnelService{executor: mockExecutor}
			err := service.Create(context.Background(), tt.tunnel)

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

func TestIPsecTunnelService_Update(t *testing.T) {
	tests := []struct {
		name        string
		tunnel      IPsecTunnel
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name: "Successful update with batch",
			tunnel: IPsecTunnel{
				ID:            1,
				LocalAddress:  "192.168.1.1",
				RemoteAddress: "192.168.2.1",
				PreSharedKey:  "newsecret",
				DPDEnabled:    true,
				DPDInterval:   60,
				DPDRetry:      5,
				IKEv2Proposal: IKEv2Proposal{
					EncryptionAES256: true,
					IntegritySHA256:  true,
					GroupFourteen:    true,
				},
				IPsecTransform: IPsecTransform{
					Protocol:         "esp",
					EncryptionAES256: true,
					IntegritySHA256:  true,
				},
			},
			mockSetup: func(m *MockExecutor) {
				// Expect RunBatch to be called with update commands
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					// Verify commands include tunnel select
					hasTunnelSelect := false
					hasLocalAddr := false
					for _, cmd := range cmds {
						if cmd == "tunnel select 1" {
							hasTunnelSelect = true
						}
						if cmd == "ipsec ike local address 1 192.168.1.1" {
							hasLocalAddr = true
						}
					}
					return hasTunnelSelect && hasLocalAddr
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name: "Validation error",
			tunnel: IPsecTunnel{
				ID:            1,
				LocalAddress:  "",
				RemoteAddress: "",
				PreSharedKey:  "",
			},
			mockSetup:   func(m *MockExecutor) {},
			expectedErr: true,
			errMessage:  "invalid IPsec tunnel config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &IPsecTunnelService{executor: mockExecutor}
			err := service.Update(context.Background(), tt.tunnel)

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

func TestIPsecTunnelService_Delete(t *testing.T) {
	tests := []struct {
		name        string
		tunnelID    int
		mockSetup   func(*MockExecutor)
		expectedErr bool
		errMessage  string
	}{
		{
			name:     "Successful delete with batch",
			tunnelID: 1,
			mockSetup: func(m *MockExecutor) {
				// Expect RunBatch to be called with delete commands
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasDeleteIPsec := false
					hasDeleteTunnel := false
					for _, cmd := range cmds {
						if cmd == "no ipsec tunnel 1" || cmd == "ipsec tunnel delete 1" {
							hasDeleteIPsec = true
						}
						if cmd == "no tunnel select 1" || cmd == "tunnel delete 1" {
							hasDeleteTunnel = true
						}
					}
					return hasDeleteIPsec && hasDeleteTunnel
				})).Return([]byte(""), nil)
			},
			expectedErr: false,
		},
		{
			name:     "Execution error",
			tunnelID: 1,
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

			service := &IPsecTunnelService{executor: mockExecutor}
			err := service.Delete(context.Background(), tt.tunnelID)

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

func TestIPsecTunnelService_CreateUsesRunBatch(t *testing.T) {
	t.Run("Create uses RunBatch for all commands", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		// Capture commands to verify RunBatch is used
		var capturedCommands []string
		mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				capturedCommands = args.Get(1).([]string)
			}).
			Return([]byte(""), nil)

		service := &IPsecTunnelService{executor: mockExecutor}
		err := service.Create(context.Background(), IPsecTunnel{
			ID:            1,
			LocalAddress:  "192.168.1.1",
			RemoteAddress: "192.168.2.1",
			PreSharedKey:  "secret123",
			DPDEnabled:    true,
			DPDInterval:   30,
			DPDRetry:      3,
			IKEv2Proposal: IKEv2Proposal{
				EncryptionAES256: true,
				IntegritySHA256:  true,
				GroupFourteen:    true,
			},
			IPsecTransform: IPsecTransform{
				Protocol:         "esp",
				EncryptionAES256: true,
				IntegritySHA256:  true,
			},
		})

		assert.NoError(t, err)

		// Verify that commands were passed to RunBatch
		assert.NotEmpty(t, capturedCommands, "Expected commands to be passed to RunBatch")

		// Verify key commands are present
		hasTunnelSelect := false
		hasIPsecTunnel := false
		for _, cmd := range capturedCommands {
			if cmd == "tunnel select 1" {
				hasTunnelSelect = true
			}
			if cmd == "ipsec tunnel 1" {
				hasIPsecTunnel = true
			}
		}
		assert.True(t, hasTunnelSelect, "Expected 'tunnel select 1' command")
		assert.True(t, hasIPsecTunnel, "Expected 'ipsec tunnel 1' command")
	})
}

func TestIPsecTunnelService_UpdateUsesRunBatch(t *testing.T) {
	t.Run("Update uses RunBatch for all commands", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		// Capture commands to verify RunBatch is used
		var capturedCommands []string
		mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				capturedCommands = args.Get(1).([]string)
			}).
			Return([]byte(""), nil)

		service := &IPsecTunnelService{executor: mockExecutor}
		err := service.Update(context.Background(), IPsecTunnel{
			ID:            1,
			LocalAddress:  "192.168.1.1",
			RemoteAddress: "192.168.2.1",
			PreSharedKey:  "secret123",
			IKEv2Proposal: IKEv2Proposal{
				EncryptionAES256: true,
				IntegritySHA256:  true,
				GroupFourteen:    true,
			},
			IPsecTransform: IPsecTransform{
				Protocol:         "esp",
				EncryptionAES256: true,
				IntegritySHA256:  true,
			},
		})

		assert.NoError(t, err)

		// Verify that commands were passed to RunBatch
		assert.NotEmpty(t, capturedCommands, "Expected commands to be passed to RunBatch")
	})
}

func TestIPsecTunnelService_DeleteUsesRunBatch(t *testing.T) {
	t.Run("Delete uses RunBatch for all commands", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		// Capture commands to verify RunBatch is used
		var capturedCommands []string
		mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				capturedCommands = args.Get(1).([]string)
			}).
			Return([]byte(""), nil)

		service := &IPsecTunnelService{executor: mockExecutor}
		err := service.Delete(context.Background(), 1)

		assert.NoError(t, err)

		// Verify that commands were passed to RunBatch
		assert.NotEmpty(t, capturedCommands, "Expected commands to be passed to RunBatch")
	})
}
