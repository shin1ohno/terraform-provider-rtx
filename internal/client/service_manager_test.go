package client

import (
	"context"
	"strings"
	"testing"
)

// mockServiceExecutor is a mock executor for testing ServiceManager
type mockServiceExecutor struct {
	commands []string
	outputs  map[string]string
	errors   map[string]error
}

func newMockServiceExecutor() *mockServiceExecutor {
	return &mockServiceExecutor{
		commands: []string{},
		outputs:  make(map[string]string),
		errors:   make(map[string]error),
	}
}

func (m *mockServiceExecutor) Run(ctx context.Context, cmd string) ([]byte, error) {
	m.commands = append(m.commands, cmd)

	// Check for errors
	for pattern, err := range m.errors {
		if strings.Contains(cmd, pattern) {
			return nil, err
		}
	}

	// Check for specific outputs
	for pattern, output := range m.outputs {
		if strings.Contains(cmd, pattern) {
			return []byte(output), nil
		}
	}

	return []byte{}, nil
}

func (m *mockServiceExecutor) RunBatch(ctx context.Context, cmds []string) ([]byte, error) {
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

func (m *mockServiceExecutor) setOutput(pattern, output string) {
	m.outputs[pattern] = output
}

func (m *mockServiceExecutor) getCommands() []string {
	return m.commands
}

func (m *mockServiceExecutor) clearCommands() {
	m.commands = []string{}
}

func TestServiceManager_GetHTTPD(t *testing.T) {
	executor := newMockServiceExecutor()
	executor.setOutput("grep httpd", `httpd host lan1
httpd proxy-access l2ms permit on`)

	manager := NewServiceManager(executor, nil)

	config, err := manager.GetHTTPD(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.Host != "lan1" {
		t.Errorf("expected host 'lan1', got '%s'", config.Host)
	}
	if !config.ProxyAccess {
		t.Errorf("expected proxy access enabled")
	}
}

func TestServiceManager_ConfigureHTTPD(t *testing.T) {
	executor := newMockServiceExecutor()
	manager := NewServiceManager(executor, nil)

	config := HTTPDConfig{
		Host:        "any",
		ProxyAccess: true,
	}

	err := manager.ConfigureHTTPD(context.Background(), config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commands := executor.getCommands()
	if len(commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(commands))
	}

	if !strings.Contains(commands[0], "httpd host any") {
		t.Errorf("expected httpd host command, got %s", commands[0])
	}
	if !strings.Contains(commands[1], "httpd proxy-access l2ms permit on") {
		t.Errorf("expected httpd proxy-access command, got %s", commands[1])
	}
}

func TestServiceManager_ResetHTTPD(t *testing.T) {
	executor := newMockServiceExecutor()
	manager := NewServiceManager(executor, nil)

	err := manager.ResetHTTPD(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commands := executor.getCommands()
	if len(commands) < 1 {
		t.Fatalf("expected at least 1 command, got %d", len(commands))
	}

	hasDeleteHost := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "no httpd host") {
			hasDeleteHost = true
		}
	}

	if !hasDeleteHost {
		t.Errorf("expected 'no httpd host' command")
	}
}

func TestServiceManager_GetSSHD(t *testing.T) {
	executor := newMockServiceExecutor()
	executor.setOutput("grep sshd", `sshd service on
sshd host lan1 lan2`)

	manager := NewServiceManager(executor, nil)

	config, err := manager.GetSSHD(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !config.Enabled {
		t.Errorf("expected enabled true")
	}
	if len(config.Hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(config.Hosts))
	}
	if config.Hosts[0] != "lan1" || config.Hosts[1] != "lan2" {
		t.Errorf("expected hosts ['lan1', 'lan2'], got %v", config.Hosts)
	}
}

func TestServiceManager_ConfigureSSHD(t *testing.T) {
	executor := newMockServiceExecutor()
	manager := NewServiceManager(executor, nil)

	config := SSHDConfig{
		Enabled: true,
		Hosts:   []string{"lan1", "lan2"},
	}

	err := manager.ConfigureSSHD(context.Background(), config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commands := executor.getCommands()
	if len(commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(commands))
	}

	hasHostCmd := false
	hasServiceCmd := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "sshd host lan1 lan2") {
			hasHostCmd = true
		}
		if strings.Contains(cmd, "sshd service on") {
			hasServiceCmd = true
		}
	}

	if !hasHostCmd {
		t.Errorf("expected sshd host command")
	}
	if !hasServiceCmd {
		t.Errorf("expected sshd service command")
	}
}

func TestServiceManager_ResetSSHD(t *testing.T) {
	executor := newMockServiceExecutor()
	manager := NewServiceManager(executor, nil)

	err := manager.ResetSSHD(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commands := executor.getCommands()
	if len(commands) < 1 {
		t.Fatalf("expected at least 1 command, got %d", len(commands))
	}

	hasDeleteService := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "no sshd service") {
			hasDeleteService = true
		}
	}

	if !hasDeleteService {
		t.Errorf("expected 'no sshd service' command")
	}
}

func TestServiceManager_GetSFTPD(t *testing.T) {
	executor := newMockServiceExecutor()
	executor.setOutput("grep sftpd", `sftpd host lan1 lan2`)

	manager := NewServiceManager(executor, nil)

	config, err := manager.GetSFTPD(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(config.Hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(config.Hosts))
	}
	if config.Hosts[0] != "lan1" || config.Hosts[1] != "lan2" {
		t.Errorf("expected hosts ['lan1', 'lan2'], got %v", config.Hosts)
	}
}

func TestServiceManager_ConfigureSFTPD(t *testing.T) {
	executor := newMockServiceExecutor()
	manager := NewServiceManager(executor, nil)

	config := SFTPDConfig{
		Hosts: []string{"lan1"},
	}

	err := manager.ConfigureSFTPD(context.Background(), config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commands := executor.getCommands()
	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commands))
	}

	if !strings.Contains(commands[0], "sftpd host lan1") {
		t.Errorf("expected sftpd host command, got %s", commands[0])
	}
}

func TestServiceManager_ResetSFTPD(t *testing.T) {
	executor := newMockServiceExecutor()
	manager := NewServiceManager(executor, nil)

	err := manager.ResetSFTPD(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commands := executor.getCommands()
	if len(commands) < 1 {
		t.Fatalf("expected at least 1 command, got %d", len(commands))
	}

	hasDeleteHost := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "no sftpd host") {
			hasDeleteHost = true
		}
	}

	if !hasDeleteHost {
		t.Errorf("expected 'no sftpd host' command")
	}
}

func TestServiceManager_ValidateHTTPD_InvalidHost(t *testing.T) {
	executor := newMockServiceExecutor()
	manager := NewServiceManager(executor, nil)

	config := HTTPDConfig{
		Host: "invalid",
	}

	err := manager.ConfigureHTTPD(context.Background(), config)
	if err == nil {
		t.Errorf("expected error for invalid host")
	}
	if !strings.Contains(err.Error(), "invalid host") {
		t.Errorf("expected invalid host error, got: %v", err)
	}
}

func TestServiceManager_ValidateSSHD_InvalidInterface(t *testing.T) {
	executor := newMockServiceExecutor()
	manager := NewServiceManager(executor, nil)

	config := SSHDConfig{
		Enabled: true,
		Hosts:   []string{"invalid"},
	}

	err := manager.ConfigureSSHD(context.Background(), config)
	if err == nil {
		t.Errorf("expected error for invalid interface")
	}
	if !strings.Contains(err.Error(), "invalid interface") {
		t.Errorf("expected invalid interface error, got: %v", err)
	}
}

func TestServiceManager_ValidateSFTPD_EmptyHosts(t *testing.T) {
	executor := newMockServiceExecutor()
	manager := NewServiceManager(executor, nil)

	config := SFTPDConfig{
		Hosts: []string{},
	}

	err := manager.ConfigureSFTPD(context.Background(), config)
	if err == nil {
		t.Errorf("expected error for empty hosts")
	}
	if !strings.Contains(err.Error(), "at least one host") {
		t.Errorf("expected 'at least one host' error, got: %v", err)
	}
}

func TestStringSliceEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected bool
	}{
		{
			name:     "equal slices",
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "c"},
			expected: true,
		},
		{
			name:     "different length",
			a:        []string{"a", "b"},
			b:        []string{"a", "b", "c"},
			expected: false,
		},
		{
			name:     "different content",
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "d"},
			expected: false,
		},
		{
			name:     "empty slices",
			a:        []string{},
			b:        []string{},
			expected: true,
		},
		{
			name:     "nil vs empty",
			a:        nil,
			b:        []string{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringSliceEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("stringSliceEqual(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
