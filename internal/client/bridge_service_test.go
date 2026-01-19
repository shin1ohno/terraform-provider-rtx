package client

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// mockBridgeExecutor implements Executor for testing
type mockBridgeExecutor struct {
	output    []byte
	err       error
	lastCmd   string
	cmdLog    []string
}

func (m *mockBridgeExecutor) Run(ctx context.Context, cmd string) ([]byte, error) {
	m.lastCmd = cmd
	m.cmdLog = append(m.cmdLog, cmd)
	return m.output, m.err
}

func TestBridgeService_CreateBridge(t *testing.T) {
	tests := []struct {
		name       string
		bridge     BridgeConfig
		setupMock  func(*mockBridgeExecutor)
		wantErr    bool
		errContain string
	}{
		{
			name: "create bridge with single member",
			bridge: BridgeConfig{
				Name:    "bridge1",
				Members: []string{"lan1"},
			},
			setupMock: func(m *mockBridgeExecutor) {
				m.output = []byte("")
				m.err = nil
			},
			wantErr: false,
		},
		{
			name: "create bridge with multiple members",
			bridge: BridgeConfig{
				Name:    "bridge1",
				Members: []string{"lan1", "tunnel1"},
			},
			setupMock: func(m *mockBridgeExecutor) {
				m.output = []byte("")
				m.err = nil
			},
			wantErr: false,
		},
		{
			name: "invalid bridge name",
			bridge: BridgeConfig{
				Name:    "br1",
				Members: []string{"lan1"},
			},
			setupMock: func(m *mockBridgeExecutor) {
				// Should fail validation before calling executor
			},
			wantErr:    true,
			errContain: "invalid bridge",
		},
		{
			name: "invalid member",
			bridge: BridgeConfig{
				Name:    "bridge1",
				Members: []string{"eth0"},
			},
			setupMock: func(m *mockBridgeExecutor) {
				// Should fail validation before calling executor
			},
			wantErr:    true,
			errContain: "invalid bridge",
		},
		{
			name: "duplicate members",
			bridge: BridgeConfig{
				Name:    "bridge1",
				Members: []string{"lan1", "lan1"},
			},
			setupMock: func(m *mockBridgeExecutor) {
				// Should fail validation before calling executor
			},
			wantErr:    true,
			errContain: "duplicate member",
		},
		{
			name: "executor error",
			bridge: BridgeConfig{
				Name:    "bridge1",
				Members: []string{"lan1"},
			},
			setupMock: func(m *mockBridgeExecutor) {
				m.err = errors.New("connection failed")
			},
			wantErr:    true,
			errContain: "failed to create bridge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockBridgeExecutor{}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			service := NewBridgeService(mock, nil)
			err := service.CreateBridge(context.Background(), tt.bridge)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContain)
					return
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errContain)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBridgeService_GetBridge(t *testing.T) {
	tests := []struct {
		name       string
		bridgeName string
		setupMock  func(*mockBridgeExecutor)
		wantErr    bool
		wantBridge *BridgeConfig
	}{
		{
			name:       "get existing bridge",
			bridgeName: "bridge1",
			setupMock: func(m *mockBridgeExecutor) {
				m.output = []byte("bridge member bridge1 lan1 tunnel1")
			},
			wantErr: false,
			wantBridge: &BridgeConfig{
				Name:    "bridge1",
				Members: []string{"lan1", "tunnel1"},
			},
		},
		{
			name:       "bridge not found",
			bridgeName: "bridge99",
			setupMock: func(m *mockBridgeExecutor) {
				m.output = []byte("")
			},
			wantErr: true,
		},
		{
			name:       "invalid bridge name",
			bridgeName: "br1",
			setupMock: func(m *mockBridgeExecutor) {
				// Should fail validation
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockBridgeExecutor{}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			service := NewBridgeService(mock, nil)
			bridge, err := service.GetBridge(context.Background(), tt.bridgeName)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if bridge == nil {
				t.Error("expected bridge, got nil")
				return
			}

			if bridge.Name != tt.wantBridge.Name {
				t.Errorf("Name = %q, want %q", bridge.Name, tt.wantBridge.Name)
			}

			if len(bridge.Members) != len(tt.wantBridge.Members) {
				t.Errorf("Members count = %d, want %d", len(bridge.Members), len(tt.wantBridge.Members))
			}
		})
	}
}

func TestBridgeService_UpdateBridge(t *testing.T) {
	tests := []struct {
		name       string
		bridge     BridgeConfig
		setupMock  func(*mockBridgeExecutor)
		wantErr    bool
		errContain string
	}{
		{
			name: "update bridge members",
			bridge: BridgeConfig{
				Name:    "bridge1",
				Members: []string{"lan1", "tunnel1", "tunnel2"},
			},
			setupMock: func(m *mockBridgeExecutor) {
				// First call: get existing bridge
				m.output = []byte("bridge member bridge1 lan1 tunnel1")
			},
			wantErr: false,
		},
		{
			name: "bridge does not exist",
			bridge: BridgeConfig{
				Name:    "bridge1",
				Members: []string{"lan1"},
			},
			setupMock: func(m *mockBridgeExecutor) {
				m.output = []byte("")
			},
			wantErr:    true,
			errContain: "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockBridgeExecutor{}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			service := NewBridgeService(mock, nil)
			err := service.UpdateBridge(context.Background(), tt.bridge)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContain)
					return
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errContain)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBridgeService_DeleteBridge(t *testing.T) {
	tests := []struct {
		name       string
		bridgeName string
		setupMock  func(*mockBridgeExecutor)
		wantErr    bool
		errContain string
	}{
		{
			name:       "delete existing bridge",
			bridgeName: "bridge1",
			setupMock: func(m *mockBridgeExecutor) {
				m.output = []byte("bridge member bridge1 lan1")
			},
			wantErr: false,
		},
		{
			name:       "delete non-existent bridge (idempotent)",
			bridgeName: "bridge99",
			setupMock: func(m *mockBridgeExecutor) {
				m.output = []byte("")
			},
			wantErr: false, // Should be idempotent
		},
		{
			name:       "invalid bridge name",
			bridgeName: "br1",
			setupMock: func(m *mockBridgeExecutor) {
				// Should fail validation
			},
			wantErr:    true,
			errContain: "invalid bridge name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockBridgeExecutor{}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			service := NewBridgeService(mock, nil)
			err := service.DeleteBridge(context.Background(), tt.bridgeName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContain)
					return
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errContain)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBridgeService_ListBridges(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockBridgeExecutor)
		wantErr     bool
		wantBridges int
	}{
		{
			name: "list multiple bridges",
			setupMock: func(m *mockBridgeExecutor) {
				m.output = []byte("bridge member bridge1 lan1 tunnel1\nbridge member bridge2 lan2")
			},
			wantBridges: 2,
		},
		{
			name: "empty list",
			setupMock: func(m *mockBridgeExecutor) {
				m.output = []byte("")
			},
			wantBridges: 0,
		},
		{
			name: "executor error",
			setupMock: func(m *mockBridgeExecutor) {
				m.err = errors.New("connection failed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockBridgeExecutor{}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			service := NewBridgeService(mock, nil)
			bridges, err := service.ListBridges(context.Background())

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(bridges) != tt.wantBridges {
				t.Errorf("got %d bridges, want %d", len(bridges), tt.wantBridges)
			}
		})
	}
}
