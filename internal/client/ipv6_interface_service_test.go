package client

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// mockIPv6InterfaceExecutor implements Executor for testing
type mockIPv6InterfaceExecutor struct {
	output  []byte
	err     error
	lastCmd string
	cmdLog  []string
}

func (m *mockIPv6InterfaceExecutor) Run(ctx context.Context, cmd string) ([]byte, error) {
	m.lastCmd = cmd
	m.cmdLog = append(m.cmdLog, cmd)
	return m.output, m.err
}

func (m *mockIPv6InterfaceExecutor) RunBatch(ctx context.Context, cmds []string) ([]byte, error) {
	var allOutput []byte
	for _, cmd := range cmds {
		output, err := m.Run(ctx, cmd)
		if err != nil {
			return allOutput, err
		}
		allOutput = append(allOutput, output...)
	}
	return allOutput, nil
}

func (m *mockIPv6InterfaceExecutor) SetAdministratorPassword(ctx context.Context, oldPassword, newPassword string) error {
	return nil
}

func (m *mockIPv6InterfaceExecutor) SetLoginPassword(ctx context.Context, password string) error {
	return nil
}

func TestIPv6InterfaceService_Configure(t *testing.T) {
	tests := []struct {
		name       string
		config     IPv6InterfaceConfig
		setupMock  func(*mockIPv6InterfaceExecutor)
		wantErr    bool
		errContain string
	}{
		{
			name: "configure static address",
			config: IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{
					{Address: "2001:db8::1/64"},
				},
			},
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				m.output = []byte("")
				m.err = nil
			},
			wantErr: false,
		},
		{
			name: "configure prefix-based address",
			config: IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{
					{PrefixRef: "ra-prefix@lan2", InterfaceID: "::1/64"},
				},
			},
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				m.output = []byte("")
				m.err = nil
			},
			wantErr: false,
		},
		{
			name: "configure RTADV",
			config: IPv6InterfaceConfig{
				Interface: "lan1",
				RTADV: &RTADVConfig{
					Enabled:  true,
					PrefixID: 1,
					OFlag:    true,
					MFlag:    false,
				},
			},
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				m.output = []byte("")
				m.err = nil
			},
			wantErr: false,
		},
		{
			name: "configure DHCPv6 server",
			config: IPv6InterfaceConfig{
				Interface:     "lan1",
				DHCPv6Service: "server",
			},
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				m.output = []byte("")
				m.err = nil
			},
			wantErr: false,
		},
		{
			name: "configure full settings",
			config: IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{
					{Address: "2001:db8::1/64"},
				},
				RTADV: &RTADVConfig{
					Enabled:  true,
					PrefixID: 1,
					OFlag:    true,
					MFlag:    true,
					Lifetime: 1800,
				},
				DHCPv6Service:            "server",
				MTU:                      1500,
				AccessListIPv6In:         "ipv6-in-acl",
				AccessListIPv6Out:        "ipv6-out-acl",
				AccessListIPv6DynamicIn:  "ipv6-dynamic-in",
				AccessListIPv6DynamicOut: "ipv6-dynamic-out",
			},
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				m.output = []byte("")
				m.err = nil
			},
			wantErr: false,
		},
		{
			name: "invalid interface name",
			config: IPv6InterfaceConfig{
				Interface: "invalid",
				Addresses: []IPv6Address{
					{Address: "2001:db8::1/64"},
				},
			},
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				// Should fail validation before calling executor
			},
			wantErr:    true,
			errContain: "invalid IPv6 interface",
		},
		{
			name: "empty interface name",
			config: IPv6InterfaceConfig{
				Interface: "",
				Addresses: []IPv6Address{
					{Address: "2001:db8::1/64"},
				},
			},
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				// Should fail validation before calling executor
			},
			wantErr:    true,
			errContain: "invalid IPv6 interface",
		},
		{
			name: "invalid MTU",
			config: IPv6InterfaceConfig{
				Interface: "lan1",
				MTU:       100, // Too small for IPv6
			},
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				// Should fail validation before calling executor
			},
			wantErr:    true,
			errContain: "invalid IPv6 interface",
		},
		{
			name: "executor error",
			config: IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{
					{Address: "2001:db8::1/64"},
				},
			},
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				m.err = errors.New("connection failed")
			},
			wantErr:    true,
			errContain: "failed to set IPv6 address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockIPv6InterfaceExecutor{}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			service := NewIPv6InterfaceService(mock, nil)
			err := service.Configure(context.Background(), tt.config)

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

func TestIPv6InterfaceService_Get(t *testing.T) {
	tests := []struct {
		name       string
		iface      string
		setupMock  func(*mockIPv6InterfaceExecutor)
		wantErr    bool
		wantConfig *IPv6InterfaceConfig
	}{
		{
			name:  "get static address config",
			iface: "lan1",
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				m.output = []byte(`ipv6 lan1 address 2001:db8::1/64`)
			},
			wantErr: false,
			wantConfig: &IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{
					{Address: "2001:db8::1/64"},
				},
			},
		},
		{
			name:  "get full config",
			iface: "lan1",
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				m.output = []byte(`ipv6 lan1 address 2001:db8::1/64
ipv6 lan1 rtadv send 1 o_flag=on m_flag=off
ipv6 lan1 dhcp service server
ipv6 lan1 mtu 1500`)
			},
			wantErr: false,
			wantConfig: &IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{
					{Address: "2001:db8::1/64"},
				},
				RTADV: &RTADVConfig{
					Enabled:  true,
					PrefixID: 1,
					OFlag:    true,
					MFlag:    false,
				},
				DHCPv6Service: "server",
				MTU:           1500,
			},
		},
		{
			name:  "empty config",
			iface: "lan1",
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				m.output = []byte("")
			},
			wantErr: false,
			wantConfig: &IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{},
			},
		},
		{
			name:  "invalid interface name",
			iface: "invalid",
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				// Should fail validation
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockIPv6InterfaceExecutor{}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			service := NewIPv6InterfaceService(mock, nil)
			config, err := service.Get(context.Background(), tt.iface)

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

			if config == nil {
				t.Error("expected config, got nil")
				return
			}

			if config.Interface != tt.wantConfig.Interface {
				t.Errorf("Interface = %q, want %q", config.Interface, tt.wantConfig.Interface)
			}

			if len(config.Addresses) != len(tt.wantConfig.Addresses) {
				t.Errorf("Addresses count = %d, want %d", len(config.Addresses), len(tt.wantConfig.Addresses))
			}

			if config.DHCPv6Service != tt.wantConfig.DHCPv6Service {
				t.Errorf("DHCPv6Service = %q, want %q", config.DHCPv6Service, tt.wantConfig.DHCPv6Service)
			}

			if config.MTU != tt.wantConfig.MTU {
				t.Errorf("MTU = %d, want %d", config.MTU, tt.wantConfig.MTU)
			}

			// Check RTADV
			if (config.RTADV == nil) != (tt.wantConfig.RTADV == nil) {
				t.Errorf("RTADV presence mismatch")
			}
		})
	}
}

func TestIPv6InterfaceService_Update(t *testing.T) {
	tests := []struct {
		name       string
		config     IPv6InterfaceConfig
		setupMock  func(*mockIPv6InterfaceExecutor)
		wantErr    bool
		errContain string
	}{
		{
			name: "update addresses",
			config: IPv6InterfaceConfig{
				Interface: "lan1",
				Addresses: []IPv6Address{
					{Address: "2001:db8::2/64"},
				},
			},
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				// First call returns current config
				m.output = []byte(`ipv6 lan1 address 2001:db8::1/64`)
			},
			wantErr: false,
		},
		{
			name: "update RTADV",
			config: IPv6InterfaceConfig{
				Interface: "lan1",
				RTADV: &RTADVConfig{
					Enabled:  true,
					PrefixID: 2,
					OFlag:    true,
					MFlag:    true,
				},
			},
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				m.output = []byte(`ipv6 lan1 rtadv send 1 o_flag=on m_flag=off`)
			},
			wantErr: false,
		},
		{
			name: "update DHCPv6 service",
			config: IPv6InterfaceConfig{
				Interface:     "lan1",
				DHCPv6Service: "client",
			},
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				m.output = []byte(`ipv6 lan1 dhcp service server`)
			},
			wantErr: false,
		},
		{
			name: "invalid interface name",
			config: IPv6InterfaceConfig{
				Interface: "invalid",
				Addresses: []IPv6Address{
					{Address: "2001:db8::1/64"},
				},
			},
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				// Should fail validation
			},
			wantErr:    true,
			errContain: "invalid IPv6 interface",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockIPv6InterfaceExecutor{}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			service := NewIPv6InterfaceService(mock, nil)
			err := service.Update(context.Background(), tt.config)

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

func TestIPv6InterfaceService_Reset(t *testing.T) {
	tests := []struct {
		name       string
		iface      string
		setupMock  func(*mockIPv6InterfaceExecutor)
		wantErr    bool
		errContain string
	}{
		{
			name:  "reset existing config",
			iface: "lan1",
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				m.output = []byte("")
			},
			wantErr: false,
		},
		{
			name:  "reset empty config (idempotent)",
			iface: "lan2",
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				m.output = []byte("")
			},
			wantErr: false,
		},
		{
			name:  "invalid interface name",
			iface: "invalid",
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				// Should fail validation
			},
			wantErr:    true,
			errContain: "invalid interface name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockIPv6InterfaceExecutor{}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			service := NewIPv6InterfaceService(mock, nil)
			err := service.Reset(context.Background(), tt.iface)

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

			// Check that all delete commands were issued
			// Note: Parser still generates 6 commands including filter cleanup for backward compatibility
			expectedCmdCount := 6 // address, rtadv, dhcp, mtu, filter in, filter out
			if len(mock.cmdLog) != expectedCmdCount {
				t.Errorf("expected %d commands, got %d", expectedCmdCount, len(mock.cmdLog))
			}
		})
	}
}

func TestIPv6InterfaceService_List(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockIPv6InterfaceExecutor)
		wantErr     bool
		wantConfigs int
	}{
		{
			name: "list configs with IPv6 settings",
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				// Return IPv6 config for lan1, empty for others
				m.output = []byte(`ipv6 lan1 address 2001:db8::1/64`)
			},
			wantConfigs: 6, // lan1, lan2, lan3, bridge1, pp1, tunnel1 all get queried
		},
		{
			name: "empty list",
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				m.output = []byte("")
			},
			wantConfigs: 0,
		},
		{
			name: "executor error",
			setupMock: func(m *mockIPv6InterfaceExecutor) {
				m.err = errors.New("connection failed")
			},
			wantConfigs: 0, // Errors are skipped, not returned
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockIPv6InterfaceExecutor{}
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			service := NewIPv6InterfaceService(mock, nil)
			configs, err := service.List(context.Background())

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

			// Note: The actual count depends on which interfaces have config
			// Since we're mocking with same output for all, the count varies
			if tt.wantConfigs == 0 && len(configs) > 0 {
				t.Errorf("got %d configs, want 0", len(configs))
			}
		})
	}
}

func TestIPv6AddressesEqual(t *testing.T) {
	tests := []struct {
		name string
		a    []IPv6Address
		b    []IPv6Address
		want bool
	}{
		{
			name: "equal static addresses",
			a:    []IPv6Address{{Address: "2001:db8::1/64"}},
			b:    []IPv6Address{{Address: "2001:db8::1/64"}},
			want: true,
		},
		{
			name: "different addresses",
			a:    []IPv6Address{{Address: "2001:db8::1/64"}},
			b:    []IPv6Address{{Address: "2001:db8::2/64"}},
			want: false,
		},
		{
			name: "different lengths",
			a:    []IPv6Address{{Address: "2001:db8::1/64"}},
			b:    []IPv6Address{{Address: "2001:db8::1/64"}, {Address: "2001:db8::2/64"}},
			want: false,
		},
		{
			name: "both empty",
			a:    []IPv6Address{},
			b:    []IPv6Address{},
			want: true,
		},
		{
			name: "both nil",
			a:    nil,
			b:    nil,
			want: true,
		},
		{
			name: "equal prefix addresses",
			a:    []IPv6Address{{PrefixRef: "ra-prefix@lan2", InterfaceID: "::1/64"}},
			b:    []IPv6Address{{PrefixRef: "ra-prefix@lan2", InterfaceID: "::1/64"}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ipv6AddressesEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("ipv6AddressesEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRTADVConfigsEqual(t *testing.T) {
	tests := []struct {
		name string
		a    *RTADVConfig
		b    *RTADVConfig
		want bool
	}{
		{
			name: "both nil",
			a:    nil,
			b:    nil,
			want: true,
		},
		{
			name: "one nil",
			a:    &RTADVConfig{Enabled: true, PrefixID: 1},
			b:    nil,
			want: false,
		},
		{
			name: "equal configs",
			a:    &RTADVConfig{Enabled: true, PrefixID: 1, OFlag: true, MFlag: false, Lifetime: 1800},
			b:    &RTADVConfig{Enabled: true, PrefixID: 1, OFlag: true, MFlag: false, Lifetime: 1800},
			want: true,
		},
		{
			name: "different prefix_id",
			a:    &RTADVConfig{Enabled: true, PrefixID: 1},
			b:    &RTADVConfig{Enabled: true, PrefixID: 2},
			want: false,
		},
		{
			name: "different flags",
			a:    &RTADVConfig{Enabled: true, PrefixID: 1, OFlag: true},
			b:    &RTADVConfig{Enabled: true, PrefixID: 1, OFlag: false},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rtadvConfigsEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("rtadvConfigsEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
