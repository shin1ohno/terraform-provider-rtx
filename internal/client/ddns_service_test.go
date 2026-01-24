package client

import (
	"context"
	"errors"
	"testing"
)

// mockDDNSExecutor implements Executor interface for DDNS testing
type mockDDNSExecutor struct {
	responses map[string][]byte
	errors    map[string]error
	commands  []string
}

func newMockDDNSExecutor() *mockDDNSExecutor {
	return &mockDDNSExecutor{
		responses: make(map[string][]byte),
		errors:    make(map[string]error),
		commands:  []string{},
	}
}

func (m *mockDDNSExecutor) Run(ctx context.Context, cmd string) ([]byte, error) {
	m.commands = append(m.commands, cmd)
	if err, ok := m.errors[cmd]; ok {
		return nil, err
	}
	if resp, ok := m.responses[cmd]; ok {
		return resp, nil
	}
	return []byte{}, nil
}

func (m *mockDDNSExecutor) RunBatch(ctx context.Context, cmds []string) ([]byte, error) {
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

func (m *mockDDNSExecutor) SetAdministratorPassword(ctx context.Context, oldPassword, newPassword string) error {
	return nil
}

func (m *mockDDNSExecutor) SetLoginPassword(ctx context.Context, password string) error {
	return nil
}

func (m *mockDDNSExecutor) setResponse(cmd string, resp []byte) {
	m.responses[cmd] = resp
}

func (m *mockDDNSExecutor) setError(cmd string, err error) {
	m.errors[cmd] = err
}

// mockDDNSClient provides SaveConfig for testing
type mockDDNSClient struct {
	saveConfigCalled bool
	saveConfigError  error
}

func (m *mockDDNSClient) SaveConfig(ctx context.Context) error {
	m.saveConfigCalled = true
	return m.saveConfigError
}

// ============================================================================
// NetVolante DNS Tests
// ============================================================================

func TestDDNSService_GetNetVolante(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		wantConfigs int
		wantErr     bool
	}{
		{
			name: "single config",
			output: `netvolante-dns hostname host pp 1 myhost.aa0.netvolante.jp
netvolante-dns use pp 1 on
`,
			wantConfigs: 1,
			wantErr:     false,
		},
		{
			name: "multiple configs",
			output: `netvolante-dns hostname host pp 1 host1.aa0.netvolante.jp
netvolante-dns hostname host pp 2 host2.aa0.netvolante.jp
netvolante-dns use pp 1 on
netvolante-dns use pp 2 on
`,
			wantConfigs: 2,
			wantErr:     false,
		},
		{
			name:        "no config",
			output:      "",
			wantConfigs: 0,
			wantErr:     false,
		},
		{
			name: "config with IPv6 enabled",
			output: `netvolante-dns hostname host pp 1 myhost.aa0.netvolante.jp
netvolante-dns use pp 1 on
netvolante-dns use ipv6 pp 1 on
`,
			wantConfigs: 1,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := newMockDDNSExecutor()
			executor.setResponse("show config | grep netvolante-dns", []byte(tt.output))

			mockClient := &mockDDNSClient{}
			service := &DDNSService{
				executor: executor,
				client:   &rtxClient{}, // Use real struct, but we won't use it for GetNetVolante
			}
			// Override client for tests that need SaveConfig
			_ = mockClient

			configs, err := service.GetNetVolante(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNetVolante() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(configs) != tt.wantConfigs {
				t.Errorf("GetNetVolante() got %d configs, want %d", len(configs), tt.wantConfigs)
			}
		})
	}
}

func TestDDNSService_GetNetVolanteByInterface(t *testing.T) {
	executor := newMockDDNSExecutor()
	executor.setResponse("show config | grep netvolante-dns", []byte(`netvolante-dns hostname host pp 1 myhost.aa0.netvolante.jp
netvolante-dns use pp 1 on
`))

	service := &DDNSService{
		executor: executor,
		client:   &rtxClient{},
	}

	// Test found
	config, err := service.GetNetVolanteByInterface(context.Background(), "pp 1")
	if err != nil {
		t.Errorf("GetNetVolanteByInterface() unexpected error: %v", err)
	}
	if config == nil {
		t.Error("GetNetVolanteByInterface() returned nil config")
	}

	// Test not found
	_, err = service.GetNetVolanteByInterface(context.Background(), "pp 2")
	if err == nil {
		t.Error("GetNetVolanteByInterface() expected error for non-existent interface")
	}
}

func TestDDNSService_ConfigureNetVolante_ExecutorError(t *testing.T) {
	executor := newMockDDNSExecutor()
	executor.setError("netvolante-dns hostname host pp 1 test.aa0.netvolante.jp", errors.New("connection failed"))

	service := &DDNSService{
		executor: executor,
		client:   &rtxClient{},
	}

	config := NetVolanteConfig{
		Hostname:  "test.aa0.netvolante.jp",
		Interface: "pp 1",
		Use:       true,
	}

	err := service.ConfigureNetVolante(context.Background(), config)
	if err == nil {
		t.Error("ConfigureNetVolante() expected error when executor fails")
	}
}

func TestDDNSService_DeleteNetVolante_EmptyInterface(t *testing.T) {
	service := &DDNSService{
		executor: newMockDDNSExecutor(),
		client:   &rtxClient{},
	}

	err := service.DeleteNetVolante(context.Background(), "")
	if err == nil {
		t.Error("DeleteNetVolante() expected error for empty interface")
	}
}

// ============================================================================
// Custom DDNS Tests
// ============================================================================

func TestDDNSService_GetDDNS(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		wantConfigs int
		wantErr     bool
	}{
		{
			name: "single server",
			output: `ddns server url 1 https://dynupdate.no-ip.com/nic/update
ddns server hostname 1 myhost.no-ip.org
ddns server user 1 myuser mypassword
`,
			wantConfigs: 1,
			wantErr:     false,
		},
		{
			name: "multiple servers",
			output: `ddns server url 1 https://dynupdate.no-ip.com/nic/update
ddns server hostname 1 host1.no-ip.org
ddns server url 2 https://www.dyndns.org/nic/update
ddns server hostname 2 host2.dyndns.org
`,
			wantConfigs: 2,
			wantErr:     false,
		},
		{
			name:        "no config",
			output:      "",
			wantConfigs: 0,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := newMockDDNSExecutor()
			executor.setResponse("show config | grep \"ddns server\"", []byte(tt.output))

			service := &DDNSService{
				executor: executor,
				client:   &rtxClient{},
			}

			configs, err := service.GetDDNS(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDDNS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(configs) != tt.wantConfigs {
				t.Errorf("GetDDNS() got %d configs, want %d", len(configs), tt.wantConfigs)
			}
		})
	}
}

func TestDDNSService_GetDDNSByID(t *testing.T) {
	executor := newMockDDNSExecutor()
	executor.setResponse("show config | grep \"ddns server\"", []byte(`ddns server url 1 https://dynupdate.no-ip.com/nic/update
ddns server hostname 1 myhost.no-ip.org
`))

	service := &DDNSService{
		executor: executor,
		client:   &rtxClient{},
	}

	// Test found
	config, err := service.GetDDNSByID(context.Background(), 1)
	if err != nil {
		t.Errorf("GetDDNSByID() unexpected error: %v", err)
	}
	if config == nil {
		t.Error("GetDDNSByID() returned nil config")
	}

	// Test not found
	_, err = service.GetDDNSByID(context.Background(), 2)
	if err == nil {
		t.Error("GetDDNSByID() expected error for non-existent ID")
	}
}

func TestDDNSService_DeleteDDNS_InvalidID(t *testing.T) {
	service := &DDNSService{
		executor: newMockDDNSExecutor(),
		client:   &rtxClient{},
	}

	tests := []struct {
		name string
		id   int
	}{
		{"zero", 0},
		{"negative", -1},
		{"too high", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteDDNS(context.Background(), tt.id)
			if err == nil {
				t.Errorf("DeleteDDNS(%d) expected error for invalid ID", tt.id)
			}
		})
	}
}

func TestDDNSService_TriggerDDNSUpdate_InvalidID(t *testing.T) {
	service := &DDNSService{
		executor: newMockDDNSExecutor(),
		client:   &rtxClient{},
	}

	err := service.TriggerDDNSUpdate(context.Background(), 0)
	if err == nil {
		t.Error("TriggerDDNSUpdate() expected error for invalid ID")
	}

	err = service.TriggerDDNSUpdate(context.Background(), 5)
	if err == nil {
		t.Error("TriggerDDNSUpdate() expected error for ID > 4")
	}
}

func TestDDNSService_TriggerNetVolanteUpdate_EmptyInterface(t *testing.T) {
	service := &DDNSService{
		executor: newMockDDNSExecutor(),
		client:   &rtxClient{},
	}

	err := service.TriggerNetVolanteUpdate(context.Background(), "")
	if err == nil {
		t.Error("TriggerNetVolanteUpdate() expected error for empty interface")
	}
}

// ============================================================================
// Context Cancellation Tests
// ============================================================================

func TestDDNSService_ConfigureNetVolante_ContextCanceled(t *testing.T) {
	executor := newMockDDNSExecutor()
	service := &DDNSService{
		executor: executor,
		client:   &rtxClient{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := NetVolanteConfig{
		Hostname:  "test.aa0.netvolante.jp",
		Interface: "pp 1",
		Use:       true,
	}

	err := service.ConfigureNetVolante(ctx, config)
	if err == nil {
		t.Error("ConfigureNetVolante() expected error when context is canceled")
	}
}

func TestDDNSService_DeleteNetVolante_ContextCanceled(t *testing.T) {
	executor := newMockDDNSExecutor()
	service := &DDNSService{
		executor: executor,
		client:   &rtxClient{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := service.DeleteNetVolante(ctx, "pp 1")
	if err == nil {
		t.Error("DeleteNetVolante() expected error when context is canceled")
	}
}

func TestDDNSService_ConfigureDDNS_ContextCanceled(t *testing.T) {
	executor := newMockDDNSExecutor()
	service := &DDNSService{
		executor: executor,
		client:   &rtxClient{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := DDNSServerConfig{
		ID:       1,
		URL:      "https://example.com/update",
		Hostname: "test.example.com",
	}

	err := service.ConfigureDDNS(ctx, config)
	if err == nil {
		t.Error("ConfigureDDNS() expected error when context is canceled")
	}
}

func TestDDNSService_DeleteDDNS_ContextCanceled(t *testing.T) {
	executor := newMockDDNSExecutor()
	service := &DDNSService{
		executor: executor,
		client:   &rtxClient{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := service.DeleteDDNS(ctx, 1)
	if err == nil {
		t.Error("DeleteDDNS() expected error when context is canceled")
	}
}
