package client

import (
	"context"
	"errors"
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

func (m *mockServiceExecutor) SetAdministratorPassword(ctx context.Context, oldPassword, newPassword string) error {
	return nil
}

func (m *mockServiceExecutor) SetLoginPassword(ctx context.Context, password string) error {
	return nil
}

func (m *mockServiceExecutor) GenerateSSHDHostKey(ctx context.Context) error {
	return nil
}

func (m *mockServiceExecutor) setOutput(pattern, output string) {
	m.outputs[pattern] = output
}

func (m *mockServiceExecutor) getCommands() []string {
	return m.commands
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
	hasDeleteAuthMethod := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "no sshd service") {
			hasDeleteService = true
		}
		if strings.Contains(cmd, "no sshd auth method") {
			hasDeleteAuthMethod = true
		}
	}

	if !hasDeleteService {
		t.Errorf("expected 'no sshd service' command")
	}
	if !hasDeleteAuthMethod {
		t.Errorf("expected 'no sshd auth method' command")
	}
}

func TestServiceManager_ConfigureSSHD_WithAuthMethod(t *testing.T) {
	tests := []struct {
		name            string
		authMethod      string
		expectedCmdPart string
	}{
		{
			name:            "password auth",
			authMethod:      "password",
			expectedCmdPart: "sshd auth method password",
		},
		{
			name:            "publickey auth",
			authMethod:      "publickey",
			expectedCmdPart: "sshd auth method publickey",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := newMockServiceExecutor()
			manager := NewServiceManager(executor, nil)

			config := SSHDConfig{
				Enabled:    true,
				Hosts:      []string{"lan1"},
				AuthMethod: tt.authMethod,
			}

			err := manager.ConfigureSSHD(context.Background(), config)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			commands := executor.getCommands()
			hasAuthCmd := false
			for _, cmd := range commands {
				if strings.Contains(cmd, tt.expectedCmdPart) {
					hasAuthCmd = true
					break
				}
			}

			if !hasAuthCmd {
				t.Errorf("expected command containing '%s', got commands: %v", tt.expectedCmdPart, commands)
			}
		})
	}
}

func TestServiceManager_ConfigureSSHD_AnyAuthMethod(t *testing.T) {
	// When auth_method is "any" or empty, no auth method command should be sent
	executor := newMockServiceExecutor()
	manager := NewServiceManager(executor, nil)

	config := SSHDConfig{
		Enabled:    true,
		Hosts:      []string{"lan1"},
		AuthMethod: "any",
	}

	err := manager.ConfigureSSHD(context.Background(), config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commands := executor.getCommands()
	for _, cmd := range commands {
		if strings.Contains(cmd, "sshd auth method") {
			t.Errorf("unexpected auth method command for 'any' auth: %s", cmd)
		}
	}
}

func TestServiceManager_UpdateSSHD_AuthMethodChange(t *testing.T) {
	tests := []struct {
		name            string
		currentAuth     string
		newAuth         string
		expectedCmdPart string
		expectAuthCmd   bool
	}{
		{
			name:            "change from any to password",
			currentAuth:     "any",
			newAuth:         "password",
			expectedCmdPart: "sshd auth method password",
			expectAuthCmd:   true,
		},
		{
			name:            "change from password to publickey",
			currentAuth:     "password",
			newAuth:         "publickey",
			expectedCmdPart: "sshd auth method publickey",
			expectAuthCmd:   true,
		},
		{
			name:            "change from publickey to any",
			currentAuth:     "publickey",
			newAuth:         "any",
			expectedCmdPart: "no sshd auth method",
			expectAuthCmd:   true,
		},
		{
			name:            "no change (same auth method)",
			currentAuth:     "password",
			newAuth:         "password",
			expectedCmdPart: "",
			expectAuthCmd:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := newMockServiceExecutor()
			// Mock GetSSHD response with current auth method
			executor.setOutput("grep sshd", "sshd service on\nsshd host lan1\nsshd auth method "+tt.currentAuth)
			manager := NewServiceManager(executor, nil)

			config := SSHDConfig{
				Enabled:    true,
				Hosts:      []string{"lan1"},
				AuthMethod: tt.newAuth,
			}

			err := manager.UpdateSSHD(context.Background(), config)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			commands := executor.getCommands()
			hasAuthCmd := false
			for _, cmd := range commands {
				if tt.expectedCmdPart != "" && strings.Contains(cmd, tt.expectedCmdPart) {
					hasAuthCmd = true
					break
				}
			}

			if tt.expectAuthCmd && !hasAuthCmd {
				t.Errorf("expected auth method command containing '%s', got commands: %v", tt.expectedCmdPart, commands)
			}
			if !tt.expectAuthCmd && hasAuthCmd {
				t.Errorf("did not expect auth method command, but got one in: %v", commands)
			}
		})
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

func TestServiceManager_GetSSHDAuthorizedKeys(t *testing.T) {
	tests := []struct {
		name         string
		output       string
		expectedKeys int
		expectedType string
		expectedFP   string
		expectedComm string
		expectedErr  bool
	}{
		{
			name:         "no keys",
			output:       "",
			expectedKeys: 0,
		},
		{
			name:         "single ed25519 key",
			output:       "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample user@host\n",
			expectedKeys: 1,
			expectedType: "ssh-ed25519",
			expectedFP:   "AAAAC3NzaC1lZDI1NTE5AAAAIExample",
			expectedComm: "user@host",
		},
		{
			name:         "multiple keys",
			output:       "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample user@host\nssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQExample admin@pc\n",
			expectedKeys: 2,
		},
		{
			name:         "key without comment",
			output:       "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAINoComment\n",
			expectedKeys: 1,
			expectedType: "ssh-ed25519",
			expectedFP:   "AAAAC3NzaC1lZDI1NTE5AAAAINoComment",
			expectedComm: "",
		},
		{
			name: "wrapped key (RTX line wrap)",
			output: `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDCmJ1iJqU/Vcd2orBBidRnzkt6v4hzVwNCCaUGtrQ
YlhH7Tlx24Qom8w/V+pzOBFZiIbbYbBUHDNkmx8BDxV9WnvcgHtWoJYmDwjC user@host`,
			expectedKeys: 1,
			expectedType: "ssh-rsa",
			expectedFP:   "AAAAB3NzaC1yc2EAAAADAQABAAABgQDCmJ1iJqU/Vcd2orBBidRnzkt6v4hzVwNCCaUGtrQYlhH7Tlx24Qom8w/V+pzOBFZiIbbYbBUHDNkmx8BDxV9WnvcgHtWoJYmDwjC",
			expectedComm: "user@host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := newMockServiceExecutor()
			executor.setOutput("show sshd authorized-keys", tt.output)
			manager := NewServiceManager(executor, nil)

			keys, err := manager.GetSSHDAuthorizedKeys(context.Background(), "testuser")

			if tt.expectedErr && err == nil {
				t.Errorf("expected error but got nil")
				return
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(keys) != tt.expectedKeys {
				t.Errorf("expected %d keys, got %d", tt.expectedKeys, len(keys))
				return
			}

			if tt.expectedKeys > 0 && tt.expectedType != "" {
				if keys[0].Type != tt.expectedType {
					t.Errorf("expected type %s, got %s", tt.expectedType, keys[0].Type)
				}
				if keys[0].Fingerprint != tt.expectedFP {
					t.Errorf("expected fingerprint %s, got %s", tt.expectedFP, keys[0].Fingerprint)
				}
				if keys[0].Comment != tt.expectedComm {
					t.Errorf("expected comment %s, got %s", tt.expectedComm, keys[0].Comment)
				}
			}

			// Verify command was executed
			commands := executor.getCommands()
			if len(commands) < 1 {
				t.Errorf("expected at least 1 command to be executed")
			}
			if len(commands) > 0 && !strings.Contains(commands[0], "show sshd authorized-keys testuser") {
				t.Errorf("expected show sshd authorized-keys command, got %s", commands[0])
			}
		})
	}
}

func TestServiceManager_SetSSHDAuthorizedKeys(t *testing.T) {
	executor := newMockServiceExecutor()
	manager := NewServiceManager(executor, nil)

	keys := []string{
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample user@host",
		"ssh-rsa AAAAB3NzaC1yc2EAAAAExample admin@pc",
	}

	err := manager.SetSSHDAuthorizedKeys(context.Background(), "testuser", keys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commands := executor.getCommands()

	// Should have: 1 delete + 2 keys * (import + key + empty line)
	// But RunBatch may collapse commands
	hasDelete := false
	hasImport := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "delete /ssh/authorized_keys/testuser") {
			hasDelete = true
		}
		if strings.Contains(cmd, "import sshd authorized-keys testuser") {
			hasImport = true
		}
	}

	if !hasDelete {
		t.Errorf("expected delete command to be executed")
	}
	if !hasImport {
		t.Errorf("expected import command to be executed")
	}
}

func TestServiceManager_SetSSHDAuthorizedKeys_Empty(t *testing.T) {
	executor := newMockServiceExecutor()
	manager := NewServiceManager(executor, nil)

	// Setting empty keys should just delete existing keys
	err := manager.SetSSHDAuthorizedKeys(context.Background(), "testuser", []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commands := executor.getCommands()

	// Should have delete command but no import
	hasDelete := false
	hasImport := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "delete /ssh/authorized_keys/testuser") {
			hasDelete = true
		}
		if strings.Contains(cmd, "import sshd authorized-keys") {
			hasImport = true
		}
	}

	if !hasDelete {
		t.Errorf("expected delete command to be executed")
	}
	if hasImport {
		t.Errorf("unexpected import command for empty keys")
	}
}

func TestServiceManager_DeleteSSHDAuthorizedKeys(t *testing.T) {
	executor := newMockServiceExecutor()
	manager := NewServiceManager(executor, nil)

	err := manager.DeleteSSHDAuthorizedKeys(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	commands := executor.getCommands()
	if len(commands) < 1 {
		t.Fatalf("expected at least 1 command, got %d", len(commands))
	}

	hasDelete := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "delete /ssh/authorized_keys/testuser") {
			hasDelete = true
		}
	}

	if !hasDelete {
		t.Errorf("expected delete command")
	}
}

func TestServiceManager_DeleteSSHDAuthorizedKeys_NotFound(t *testing.T) {
	executor := newMockServiceExecutor()
	executor.outputs["delete"] = "not found"
	manager := NewServiceManager(executor, nil)

	// Should not return error for "not found"
	err := manager.DeleteSSHDAuthorizedKeys(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("unexpected error for not found case: %v", err)
	}
}

// mockSFTPClientForServiceManager is a mock SFTP client for testing ServiceManager
type mockSFTPClientForServiceManager struct {
	writeFileFunc func(ctx context.Context, path string, content []byte) error
	writtenFiles  map[string][]byte
}

func newMockSFTPClient() *mockSFTPClientForServiceManager {
	return &mockSFTPClientForServiceManager{
		writtenFiles: make(map[string][]byte),
	}
}

func (m *mockSFTPClientForServiceManager) Download(ctx context.Context, path string) ([]byte, error) {
	return nil, nil
}

func (m *mockSFTPClientForServiceManager) ListDir(ctx context.Context, path string) ([]string, error) {
	return nil, nil
}

func (m *mockSFTPClientForServiceManager) WriteFile(ctx context.Context, path string, content []byte) error {
	if m.writeFileFunc != nil {
		return m.writeFileFunc(ctx, path, content)
	}
	m.writtenFiles[path] = content
	return nil
}

func (m *mockSFTPClientForServiceManager) Close() error {
	return nil
}

func TestServiceManager_SetSSHDAuthorizedKeys_ViaSFTP(t *testing.T) {
	executor := newMockServiceExecutor()
	mockSFTP := newMockSFTPClient()
	manager := NewServiceManager(executor, nil)
	manager.SetSFTPClient(mockSFTP)

	keys := []string{
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample user@host",
		"ssh-rsa AAAAB3NzaC1yc2EAAAAExample admin@pc",
	}

	err := manager.SetSSHDAuthorizedKeys(context.Background(), "testuser", keys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that file was written via SFTP
	expectedPath := "/ssh/authorized_keys/testuser"
	writtenContent, ok := mockSFTP.writtenFiles[expectedPath]
	if !ok {
		t.Fatalf("expected file to be written at %s", expectedPath)
	}

	// Should contain both keys
	contentStr := string(writtenContent)
	if !strings.Contains(contentStr, "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample user@host") {
		t.Errorf("expected first key in content, got: %s", contentStr)
	}
	if !strings.Contains(contentStr, "ssh-rsa AAAAB3NzaC1yc2EAAAAExample admin@pc") {
		t.Errorf("expected second key in content, got: %s", contentStr)
	}

	// Check that link command was executed
	commands := executor.getCommands()
	hasLinkCmd := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "sshd authorized-keys filename testuser") {
			hasLinkCmd = true
			break
		}
	}
	if !hasLinkCmd {
		t.Errorf("expected sshd authorized-keys filename command to be executed")
	}
}

func TestServiceManager_SetSSHDAuthorizedKeys_SFTPError(t *testing.T) {
	executor := newMockServiceExecutor()
	mockSFTP := newMockSFTPClient()
	mockSFTP.writeFileFunc = func(ctx context.Context, path string, content []byte) error {
		return errors.New("SFTP write failed")
	}
	manager := NewServiceManager(executor, nil)
	manager.SetSFTPClient(mockSFTP)

	keys := []string{"ssh-ed25519 AAAAC3... user@host"}

	err := manager.SetSSHDAuthorizedKeys(context.Background(), "testuser", keys)
	if err == nil {
		t.Fatal("expected error from SFTP write failure")
	}
	if !strings.Contains(err.Error(), "SFTP") {
		t.Errorf("expected SFTP-related error, got: %v", err)
	}
}

func TestServiceManager_SetSSHDAuthorizedKeys_FallbackToCommand(t *testing.T) {
	executor := newMockServiceExecutor()
	// No SFTP client set, should fall back to command method
	manager := NewServiceManager(executor, nil)

	keys := []string{"ssh-ed25519 AAAAC3... user@host"}

	err := manager.SetSSHDAuthorizedKeys(context.Background(), "testuser", keys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use command method (import command)
	commands := executor.getCommands()
	hasImport := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "import sshd authorized-keys") {
			hasImport = true
			break
		}
	}
	if !hasImport {
		t.Errorf("expected import command to be executed when SFTP is not available")
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
