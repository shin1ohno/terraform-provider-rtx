package client

import (
	"context"
	"errors"
	"testing"
)

// mockPPPExecutor implements Executor interface for PPP testing
type mockPPPExecutor struct {
	responses map[string][]byte
	errors    map[string]error
	commands  []string
}

func newMockPPPExecutor() *mockPPPExecutor {
	return &mockPPPExecutor{
		responses: make(map[string][]byte),
		errors:    make(map[string]error),
		commands:  []string{},
	}
}

func (m *mockPPPExecutor) Run(ctx context.Context, cmd string) ([]byte, error) {
	m.commands = append(m.commands, cmd)
	if err, ok := m.errors[cmd]; ok {
		return nil, err
	}
	if resp, ok := m.responses[cmd]; ok {
		return resp, nil
	}
	return []byte{}, nil
}

func (m *mockPPPExecutor) RunBatch(ctx context.Context, cmds []string) ([]byte, error) {
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

func (m *mockPPPExecutor) setResponse(cmd string, resp []byte) {
	m.responses[cmd] = resp
}

func (m *mockPPPExecutor) setError(cmd string, err error) {
	m.errors[cmd] = err
}

// ============================================================================
// PPPoE List/Get Tests
// ============================================================================

func TestPPPService_List(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		wantConfigs int
		wantErr     bool
	}{
		{
			name: "single PPPoE config",
			output: `pp select 1
 description pp ISP-Connection
 pp bind lan2
 pppoe use lan2
 pp auth accept pap chap
 pp auth myname user@provider.jp password123
 pp always-on on
 pppoe auto disconnect off
 ip pp address ipcp
 ip pp mtu 1454
 ip pp tcp mss limit auto
 pp enable 1
`,
			wantConfigs: 1,
			wantErr:     false,
		},
		{
			name: "multiple PPPoE configs",
			output: `pp select 1
 description pp ISP-1
 pp bind lan2
 pppoe use lan2
 pp enable 1
pp select 2
 description pp ISP-2
 pp bind lan3
 pppoe use lan3
 pp enable 2
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
			executor := newMockPPPExecutor()
			executor.setResponse("show config", []byte(tt.output))

			service := &PPPService{
				executor: executor,
				client:   &rtxClient{},
			}

			configs, err := service.List(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(configs) != tt.wantConfigs {
				t.Errorf("List() got %d configs, want %d", len(configs), tt.wantConfigs)
			}
		})
	}
}

func TestPPPService_Get(t *testing.T) {
	executor := newMockPPPExecutor()
	executor.setResponse("show config", []byte(`pp select 1
 description pp ISP-Connection
 pp bind lan2
 pppoe use lan2
 pp enable 1
`))

	service := &PPPService{
		executor: executor,
		client:   &rtxClient{},
	}

	// Test found
	config, err := service.Get(context.Background(), 1)
	if err != nil {
		t.Errorf("Get() unexpected error: %v", err)
	}
	if config == nil {
		t.Error("Get() returned nil config")
	}
	if config.Number != 1 {
		t.Errorf("Get() config.Number = %d, want 1", config.Number)
	}

	// Test not found
	_, err = service.Get(context.Background(), 2)
	if err == nil {
		t.Error("Get() expected error for non-existent PP number")
	}
}

func TestPPPService_Get_ExecutorError(t *testing.T) {
	executor := newMockPPPExecutor()
	executor.setError("show config", errors.New("connection failed"))

	service := &PPPService{
		executor: executor,
		client:   &rtxClient{},
	}

	_, err := service.Get(context.Background(), 1)
	if err == nil {
		t.Error("Get() expected error when executor fails")
	}
}

// ============================================================================
// PPPoE Create Tests
// ============================================================================

func TestPPPService_Create_InvalidConfig(t *testing.T) {
	executor := newMockPPPExecutor()
	service := &PPPService{
		executor: executor,
		client:   &rtxClient{},
	}

	// Empty config should fail validation
	config := PPPoEConfig{}
	err := service.Create(context.Background(), config)
	if err == nil {
		t.Error("Create() expected error for invalid config")
	}
}

func TestPPPService_Create_ContextCanceled(t *testing.T) {
	executor := newMockPPPExecutor()
	service := &PPPService{
		executor: executor,
		client:   &rtxClient{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := PPPoEConfig{
		Number:        1,
		BindInterface: "lan2",
	}

	err := service.Create(ctx, config)
	if err == nil {
		t.Error("Create() expected error when context is canceled")
	}
}

// ============================================================================
// PPPoE Delete Tests
// ============================================================================

func TestPPPService_Delete_InvalidPPNumber(t *testing.T) {
	executor := newMockPPPExecutor()
	service := &PPPService{
		executor: executor,
		client:   &rtxClient{},
	}

	err := service.Delete(context.Background(), 0)
	if err == nil {
		t.Error("Delete() expected error for invalid PP number 0")
	}

	err = service.Delete(context.Background(), -1)
	if err == nil {
		t.Error("Delete() expected error for negative PP number")
	}
}

func TestPPPService_Delete_ContextCanceled(t *testing.T) {
	executor := newMockPPPExecutor()
	service := &PPPService{
		executor: executor,
		client:   &rtxClient{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := service.Delete(ctx, 1)
	if err == nil {
		t.Error("Delete() expected error when context is canceled")
	}
}

// ============================================================================
// PP Interface IP Config Tests
// ============================================================================

func TestPPPService_GetIPConfig(t *testing.T) {
	executor := newMockPPPExecutor()
	executor.setResponse("show config", []byte(`pp select 1
 ip pp address ipcp
 ip pp mtu 1454
 ip pp tcp mss limit auto
 ip pp nat descriptor 1
`))

	service := &PPPService{
		executor: executor,
		client:   &rtxClient{},
	}

	config, err := service.GetIPConfig(context.Background(), 1)
	if err != nil {
		t.Errorf("GetIPConfig() unexpected error: %v", err)
	}
	if config == nil {
		t.Error("GetIPConfig() returned nil config")
	}
}

func TestPPPService_ConfigureIPConfig_ContextCanceled(t *testing.T) {
	executor := newMockPPPExecutor()
	service := &PPPService{
		executor: executor,
		client:   &rtxClient{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := PPIPConfig{
		Address: "ipcp",
		MTU:     1454,
	}

	err := service.ConfigureIPConfig(ctx, 1, config)
	if err == nil {
		t.Error("ConfigureIPConfig() expected error when context is canceled")
	}
}

// ============================================================================
// Connection Status Tests
// ============================================================================

func TestPPPService_GetConnectionStatus(t *testing.T) {
	tests := []struct {
		name          string
		output        string
		wantConnected bool
		wantState     string
	}{
		{
			name:          "connected",
			output:        "PP[ON]\nPPPoE connection active",
			wantConnected: true,
			wantState:     "connected",
		},
		{
			name:          "disconnected",
			output:        "PP[OFF]\nNo connection",
			wantConnected: false,
			wantState:     "disconnected",
		},
		{
			name:          "connected Japanese",
			output:        "接続中\nactive session",
			wantConnected: true,
			wantState:     "connected",
		},
		{
			name:          "disconnected Japanese",
			output:        "切断\nno session",
			wantConnected: false,
			wantState:     "disconnected",
		},
		{
			name:          "unknown state",
			output:        "Some other output",
			wantConnected: false,
			wantState:     "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := newMockPPPExecutor()
			executor.setResponse("show status pp 1", []byte(tt.output))

			service := &PPPService{
				executor: executor,
				client:   &rtxClient{},
			}

			status, err := service.GetConnectionStatus(context.Background(), 1)
			if err != nil {
				t.Errorf("GetConnectionStatus() unexpected error: %v", err)
				return
			}
			if status.Connected != tt.wantConnected {
				t.Errorf("GetConnectionStatus() connected = %v, want %v", status.Connected, tt.wantConnected)
			}
			if status.State != tt.wantState {
				t.Errorf("GetConnectionStatus() state = %q, want %q", status.State, tt.wantState)
			}
		})
	}
}

func TestPPPService_GetConnectionStatus_Error(t *testing.T) {
	executor := newMockPPPExecutor()
	executor.setError("show status pp 1", errors.New("connection failed"))

	service := &PPPService{
		executor: executor,
		client:   &rtxClient{},
	}

	_, err := service.GetConnectionStatus(context.Background(), 1)
	if err == nil {
		t.Error("GetConnectionStatus() expected error when executor fails")
	}
}

// ============================================================================
// Conversion Function Tests
// ============================================================================

func TestPPPService_ConversionFunctions(t *testing.T) {
	service := &PPPService{}

	// Test toParserPPPoEConfig
	clientConfig := PPPoEConfig{
		Number:        1,
		Name:          "Test Connection",
		Interface:     "pp 1",
		BindInterface: "lan2",
		ServiceName:   "FLETS",
		AlwaysOn:      true,
		Enabled:       true,
		Authentication: &PPPAuth{
			Method:   "chap",
			Username: "user",
			Password: "pass",
		},
		IPConfig: &PPIPConfig{
			Address:       "ipcp",
			MTU:           1454,
			TCPMSSLimit:   1414,
			NATDescriptor: 1,
		},
	}

	parserConfig := service.toParserPPPoEConfig(clientConfig)
	if parserConfig.Number != 1 {
		t.Errorf("toParserPPPoEConfig() Number = %d, want 1", parserConfig.Number)
	}
	if parserConfig.Authentication == nil {
		t.Error("toParserPPPoEConfig() Authentication is nil")
	}
	if parserConfig.IPConfig == nil {
		t.Error("toParserPPPoEConfig() IPConfig is nil")
	}

	// Test fromParserPPPoEConfig (round-trip)
	backConfig := service.fromParserPPPoEConfig(parserConfig)
	if backConfig.Number != clientConfig.Number {
		t.Errorf("fromParserPPPoEConfig() Number = %d, want %d", backConfig.Number, clientConfig.Number)
	}
	if backConfig.Name != clientConfig.Name {
		t.Errorf("fromParserPPPoEConfig() Name = %q, want %q", backConfig.Name, clientConfig.Name)
	}
}

func TestPPPService_PPIPConfigConversion(t *testing.T) {
	service := &PPPService{}

	// Test toParserPPIPConfig
	config := PPIPConfig{
		Address:         "ipcp",
		MTU:             1454,
		TCPMSSLimit:     1414,
		NATDescriptor:   1,
		SecureFilterIn:  []int{1, 2, 3},
		SecureFilterOut: []int{4, 5, 6},
	}

	parserConfig := service.toParserPPIPConfig(config)
	if parserConfig.Address != "ipcp" {
		t.Errorf("toParserPPIPConfig() Address = %q, want %q", parserConfig.Address, "ipcp")
	}
	if len(parserConfig.SecureFilterIn) != 3 {
		t.Errorf("toParserPPIPConfig() SecureFilterIn length = %d, want 3", len(parserConfig.SecureFilterIn))
	}

	// Test fromParserPPIPConfig
	backConfig := service.fromParserPPIPConfig(&parserConfig)
	if backConfig.MTU != config.MTU {
		t.Errorf("fromParserPPIPConfig() MTU = %d, want %d", backConfig.MTU, config.MTU)
	}

	// Test nil input
	nilConfig := service.fromParserPPIPConfig(nil)
	if nilConfig.Address != "" {
		t.Error("fromParserPPIPConfig(nil) should return empty config")
	}
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", true},
		{"hello", "hello", true},
		{"hello", "world", false},
		{"", "test", false},
		{"test", "", true}, // Empty substring is always contained
		{"PP[ON]", "PP[ON]", true},
		{"PP[OFF]", "PP[ON]", false},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}
