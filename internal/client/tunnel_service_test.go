package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTunnelService_Get(t *testing.T) {
	tests := []struct {
		name        string
		tunnelID    int
		mockSetup   func(*MockExecutor)
		expected    *Tunnel
		expectError bool
		errMessage  string
	}{
		{
			name:     "Successful get IPsec tunnel",
			tunnelID: 1,
			mockSetup: func(m *MockExecutor) {
				output := `tunnel select 1
 ipsec tunnel 1
 ipsec sa policy 1 1 esp aes-cbc sha-hmac
 ipsec ike local address 1 192.168.1.1
 ipsec ike remote address 1 192.168.2.1
 ipsec ike pre-shared-key 1 text secret123
 ipsec ike keepalive use 1 on dpd 30 3
 tunnel enable 1
`
				m.On("Run", mock.Anything, "show config").Return([]byte(output), nil)
			},
			expected: &Tunnel{
				ID:            1,
				Encapsulation: "ipsec",
				Enabled:       true,
				IPsec: &TunnelIPsec{
					IPsecTunnelID: 1,
					LocalAddress:  "192.168.1.1",
					RemoteAddress: "192.168.2.1",
					PreSharedKey:  "secret123",
				},
			},
			expectError: false,
		},
		{
			name:     "Tunnel not found",
			tunnelID: 99,
			mockSetup: func(m *MockExecutor) {
				output := `tunnel select 1
 ipsec tunnel 1
`
				m.On("Run", mock.Anything, "show config").Return([]byte(output), nil)
			},
			expected:    nil,
			expectError: true,
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
			expectError: true,
			errMessage:  "connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &TunnelService{executor: mockExecutor}
			result, err := service.Get(context.Background(), tt.tunnelID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.ID, result.ID)
				assert.Equal(t, tt.expected.Encapsulation, result.Encapsulation)
				assert.Equal(t, tt.expected.Enabled, result.Enabled)
				if tt.expected.IPsec != nil {
					assert.NotNil(t, result.IPsec)
					assert.Equal(t, tt.expected.IPsec.LocalAddress, result.IPsec.LocalAddress)
					assert.Equal(t, tt.expected.IPsec.RemoteAddress, result.IPsec.RemoteAddress)
				}
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestTunnelService_Create(t *testing.T) {
	tests := []struct {
		name        string
		tunnel      Tunnel
		mockSetup   func(*MockExecutor)
		expectError bool
		errMessage  string
	}{
		{
			name: "Successful creation of IPsec tunnel",
			tunnel: Tunnel{
				ID:            1,
				Encapsulation: "ipsec",
				Enabled:       true,
				IPsec: &TunnelIPsec{
					IPsecTunnelID: 1,
					LocalAddress:  "192.168.1.1",
					RemoteAddress: "192.168.2.1",
					PreSharedKey:  "secret123",
					IKEv2Proposal: IKEv2Proposal{
						EncryptionAES128: true,
						IntegritySHA1:    true,
						GroupFourteen:    true,
					},
					Transform: IPsecTransform{
						Protocol:         "esp",
						EncryptionAES128: true,
						IntegritySHA1:    true,
					},
					Keepalive: &TunnelIPsecKeepalive{
						Enabled:  true,
						Mode:     "dpd",
						Interval: 30,
						Retry:    3,
					},
				},
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
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
			expectError: false,
		},
		{
			name: "Successful creation of L2TPv3 tunnel",
			tunnel: Tunnel{
				ID:            1,
				Encapsulation: "l2tpv3",
				Enabled:       true,
				IPsec: &TunnelIPsec{
					IPsecTunnelID: 101,
					LocalAddress:  "192.168.1.253",
					RemoteAddress: "itm.ohno.be",
					PreSharedKey:  "secret123",
					IKEv2Proposal: IKEv2Proposal{
						EncryptionAES128: true,
						IntegritySHA1:    true,
						GroupFourteen:    true,
					},
					Transform: IPsecTransform{
						Protocol:         "esp",
						EncryptionAES128: true,
						IntegritySHA1:    true,
					},
				},
				L2TP: &TunnelL2TP{
					Hostname:       "ebisu-RTX1210",
					LocalRouterID:  "192.168.1.253",
					RemoteRouterID: "192.168.1.254",
					RemoteEndID:    "shin1",
					AlwaysOn:       true,
				},
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasEncapsulation := false
					hasIPsecTunnel := false
					hasL2TPHostname := false
					hasL2TPLocalRouterID := false
					for _, cmd := range cmds {
						if cmd == "tunnel encapsulation l2tpv3" {
							hasEncapsulation = true
						}
						if cmd == "ipsec tunnel 101" {
							hasIPsecTunnel = true
						}
						if cmd == "l2tp hostname ebisu-RTX1210" {
							hasL2TPHostname = true
						}
						if cmd == "l2tp local router-id 192.168.1.253" {
							hasL2TPLocalRouterID = true
						}
					}
					return hasEncapsulation && hasIPsecTunnel && hasL2TPHostname && hasL2TPLocalRouterID
				})).Return([]byte(""), nil)
			},
			expectError: false,
		},
		{
			name: "Validation error - missing ipsec block for ipsec encapsulation",
			tunnel: Tunnel{
				ID:            1,
				Encapsulation: "ipsec",
				Enabled:       true,
			},
			mockSetup:   func(m *MockExecutor) {},
			expectError: true,
			errMessage:  "ipsec block is required",
		},
		{
			name: "Batch execution error",
			tunnel: Tunnel{
				ID:            1,
				Encapsulation: "ipsec",
				Enabled:       true,
				IPsec: &TunnelIPsec{
					PreSharedKey: "secret123",
				},
			},
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.Anything).
					Return(nil, errors.New("connection failed"))
			},
			expectError: true,
			errMessage:  "connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &TunnelService{executor: mockExecutor}
			err := service.Create(context.Background(), tt.tunnel)

			if tt.expectError {
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

func TestTunnelService_Delete(t *testing.T) {
	tests := []struct {
		name        string
		tunnelID    int
		mockSetup   func(*MockExecutor)
		expectError bool
		errMessage  string
	}{
		{
			name:     "Successful delete",
			tunnelID: 1,
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.MatchedBy(func(cmds []string) bool {
					hasDeleteIPsec := false
					hasDeleteTunnel := false
					for _, cmd := range cmds {
						if cmd == "no ipsec tunnel 1" {
							hasDeleteIPsec = true
						}
						if cmd == "no tunnel select 1" {
							hasDeleteTunnel = true
						}
					}
					return hasDeleteIPsec && hasDeleteTunnel
				})).Return([]byte(""), nil)
			},
			expectError: false,
		},
		{
			name:     "Execution error",
			tunnelID: 1,
			mockSetup: func(m *MockExecutor) {
				m.On("RunBatch", mock.Anything, mock.Anything).
					Return(nil, errors.New("connection failed"))
			},
			expectError: true,
			errMessage:  "connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockExecutor)
			tt.mockSetup(mockExecutor)

			service := &TunnelService{executor: mockExecutor}
			err := service.Delete(context.Background(), tt.tunnelID)

			if tt.expectError {
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

func TestTunnelService_Update(t *testing.T) {
	t.Run("Successful update of IPsec tunnel", func(t *testing.T) {
		mockExecutor := new(MockExecutor)

		var capturedCommands []string
		mockExecutor.On("RunBatch", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				capturedCommands = args.Get(1).([]string)
			}).
			Return([]byte(""), nil)

		service := &TunnelService{executor: mockExecutor}
		err := service.Update(context.Background(), Tunnel{
			ID:            1,
			Encapsulation: "ipsec",
			Enabled:       true,
			IPsec: &TunnelIPsec{
				IPsecTunnelID: 1,
				LocalAddress:  "192.168.1.1",
				RemoteAddress: "192.168.2.1",
				PreSharedKey:  "newsecret",
				IKEv2Proposal: IKEv2Proposal{
					EncryptionAES128: true,
					IntegritySHA1:    true,
					GroupFourteen:    true,
				},
				Transform: IPsecTransform{
					Protocol:         "esp",
					EncryptionAES128: true,
					IntegritySHA1:    true,
				},
			},
		})

		assert.NoError(t, err)
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

		mockExecutor.AssertExpectations(t)
	})
}
