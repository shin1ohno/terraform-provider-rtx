package client

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/sh1/terraform-provider-rtx/internal/rtx/parsers"
)

// MockSession implements the Session interface for testing
type MockSession struct {
	SendFunc  func(cmd string) ([]byte, error)
	CloseFunc func() error
	closed    bool
}

func (m *MockSession) Send(cmd string) ([]byte, error) {
	if m.closed {
		return nil, errors.New("session closed")
	}
	if m.SendFunc != nil {
		return m.SendFunc(cmd)
	}
	return []byte("mock response"), nil
}

func (m *MockSession) Close() error {
	m.closed = true
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockSession) SetAdminMode(enabled bool) {
	// Mock implementation - do nothing
}

// MockConnDialer implements the ConnDialer interface for testing
type MockConnDialer struct {
	DialFunc func(ctx context.Context, host string, config *Config) (Session, error)
}

func (m *MockConnDialer) Dial(ctx context.Context, host string, config *Config) (Session, error) {
	if m.DialFunc != nil {
		return m.DialFunc(ctx, host, config)
	}
	return &MockSession{}, nil
}

// MockPromptDetector implements the PromptDetector interface for testing
type MockPromptDetector struct {
	DetectFunc func(output []byte) (matched bool, prompt string)
}

func (m *MockPromptDetector) DetectPrompt(output []byte) (matched bool, prompt string) {
	if m.DetectFunc != nil {
		return m.DetectFunc(output)
	}
	// Default behavior: assume prompt is found
	return true, "RTX1200>"
}

// MockRetryStrategy implements the RetryStrategy interface for testing
type MockRetryStrategy struct {
	NextFunc func(retry int) (delay time.Duration, giveUp bool)
}

func (m *MockRetryStrategy) Next(retry int) (delay time.Duration, giveUp bool) {
	if m.NextFunc != nil {
		return m.NextFunc(retry)
	}
	// Default: give up after first retry
	return time.Millisecond, retry >= 1
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: &Config{
				Host:     "192.168.1.1",
				Port:     22,
				Username: "admin",
				Password: "password",
				Timeout:  30,
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: &Config{
				Port:     22,
				Username: "admin",
				Password: "password",
				Timeout:  30,
			},
			wantErr: true,
		},
		{
			name: "missing username",
			config: &Config{
				Host:     "192.168.1.1",
				Port:     22,
				Password: "password",
				Timeout:  30,
			},
			wantErr: true,
		},
		{
			name: "missing password",
			config: &Config{
				Host:     "192.168.1.1",
				Port:     22,
				Username: "admin",
				Timeout:  30,
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: &Config{
				Host:     "192.168.1.1",
				Port:     0,
				Username: "admin",
				Password: "password",
				Timeout:  30,
			},
			wantErr: true,
		},
		{
			name: "zero timeout gets default",
			config: &Config{
				Host:     "192.168.1.1",
				Port:     22,
				Username: "admin",
				Password: "password",
				Timeout:  0, // Should get default value
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewClient() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewClient() unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestClient_Dial(t *testing.T) {
	tests := []struct {
		name    string
		dialer  ConnDialer
		timeout time.Duration
		wantErr bool
		errType error
	}{
		{
			name: "successful connection",
			dialer: &MockConnDialer{
				DialFunc: func(ctx context.Context, host string, config *Config) (Session, error) {
					return &MockSession{}, nil
				},
			},
			timeout: 5 * time.Second,
			wantErr: false,
		},
		{
			name: "connection failure",
			dialer: &MockConnDialer{
				DialFunc: func(ctx context.Context, host string, config *Config) (Session, error) {
					return nil, ErrDial
				},
			},
			timeout: 5 * time.Second,
			wantErr: true,
			errType: ErrDial,
		},
		{
			name: "authentication failure",
			dialer: &MockConnDialer{
				DialFunc: func(ctx context.Context, host string, config *Config) (Session, error) {
					return nil, ErrAuthFailed
				},
			},
			timeout: 5 * time.Second,
			wantErr: true,
			errType: ErrAuthFailed,
		},
		{
			name: "context timeout",
			dialer: &MockConnDialer{
				DialFunc: func(ctx context.Context, host string, config *Config) (Session, error) {
					select {
					case <-ctx.Done():
						return nil, ctx.Err()
					case <-time.After(2 * time.Second):
						return &MockSession{}, nil
					}
				},
			},
			timeout: 100 * time.Millisecond, // Short timeout
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Host:     "192.168.1.1",
				Port:     22,
				Username: "admin",
				Password: "password",
				Timeout:  30,
			}

			client := &rtxClient{
				config: config,
				dialer: tt.dialer,
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			err := client.Dial(ctx)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Dial() expected error, got nil")
				}
				if tt.errType != nil && !errors.Is(err, tt.errType) {
					t.Errorf("Dial() expected error type %v, got %v", tt.errType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Dial() unexpected error: %v", err)
			}
		})
	}
}

func TestClient_Run(t *testing.T) {
	tests := []struct {
		name        string
		session     *MockSession
		detector    PromptDetector
		command     Command
		wantErr     bool
		wantRawData []byte
	}{
		{
			name: "successful command execution",
			session: &MockSession{
				SendFunc: func(cmd string) ([]byte, error) {
					return []byte("show version\nRTX1200 Rev.10.01.76\nRTX1200>"), nil
				},
			},
			detector: &MockPromptDetector{
				DetectFunc: func(output []byte) (matched bool, prompt string) {
					return true, "RTX1200>"
				},
			},
			command: Command{
				Key:     "show_version",
				Payload: "show version",
			},
			wantErr:     false,
			wantRawData: []byte("show version\nRTX1200 Rev.10.01.76\nRTX1200>"),
		},
		{
			name: "session send error",
			session: &MockSession{
				SendFunc: func(cmd string) ([]byte, error) {
					return nil, errors.New("send failed")
				},
			},
			detector: &MockPromptDetector{},
			command: Command{
				Key:     "show_version",
				Payload: "show version",
			},
			wantErr: true,
		},
		{
			name: "prompt not detected",
			session: &MockSession{
				SendFunc: func(cmd string) ([]byte, error) {
					return []byte("incomplete output"), nil
				},
			},
			detector: &MockPromptDetector{
				DetectFunc: func(output []byte) (matched bool, prompt string) {
					return false, ""
				},
			},
			command: Command{
				Key:     "show_version",
				Payload: "show version",
			},
			wantErr: true,
		},
		{
			name: "context timeout during command",
			session: &MockSession{
				SendFunc: func(cmd string) ([]byte, error) {
					// Return an error to simulate command failure
					time.Sleep(50 * time.Millisecond) // Short delay to ensure timeout occurs
					return nil, errors.New("command timeout")
				},
			},
			detector: &MockPromptDetector{},
			command: Command{
				Key:     "show_version",
				Payload: "show version",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Host:     "192.168.1.1",
				Port:     22,
				Username: "admin",
				Password: "password",
				Timeout:  30,
			}

			client := &rtxClient{
				config:         config,
				promptDetector: tt.detector,
				retryStrategy:  &MockRetryStrategy{},
				active:         true, // Set as connected for testing
				semaphore:      make(chan struct{}, 1),
			}
			// Initialize executor for the test (session is nil so executor will be used)
			client.executor = NewSSHExecutor(tt.session, tt.detector, &MockRetryStrategy{})

			timeout := 5 * time.Second
			if tt.name == "context timeout during command" {
				timeout = 100 * time.Millisecond
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			result, err := client.Run(ctx, tt.command)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Run() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Run() unexpected error: %v", err)
				return
			}

			if result.Raw == nil {
				t.Error("Run() result.Raw is nil")
				return
			}

			if tt.wantRawData != nil && string(result.Raw) != string(tt.wantRawData) {
				t.Errorf("Run() result.Raw = %q, want %q", string(result.Raw), string(tt.wantRawData))
			}
		})
	}
}

func TestClient_Close(t *testing.T) {
	tests := []struct {
		name    string
		session *MockSession
		wantErr bool
	}{
		{
			name: "successful close",
			session: &MockSession{
				CloseFunc: func() error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "close error",
			session: &MockSession{
				CloseFunc: func() error {
					return errors.New("close failed")
				},
			},
			wantErr: true,
		},
		{
			name:    "no active session",
			session: nil,
			wantErr: false, // Should not error when no session exists
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Host:     "192.168.1.1",
				Port:     22,
				Username: "admin",
				Password: "password",
				Timeout:  30,
			}

			client := &rtxClient{
				config:  config,
				session: tt.session,
				active:  tt.session != nil, // Set active based on session presence
			}

			err := client.Close()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Close() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Close() unexpected error: %v", err)
			}
		})
	}
}

func TestClient_Integration(t *testing.T) {
	// Integration test simulating the full client lifecycle
	t.Run("full lifecycle", func(t *testing.T) {
		session := &MockSession{
			SendFunc: func(cmd string) ([]byte, error) {
				responses := map[string][]byte{
					"show version": []byte("RTX1200 Rev.10.01.76\nRTX1200>"),
					"show config":  []byte("# RTX1200 Rev.10.01.76\n! configuration\nRTX1200>"),
				}
				if response, ok := responses[cmd]; ok {
					return response, nil
				}
				return []byte(fmt.Sprintf("Unknown command: %s\nRTX1200>", cmd)), nil
			},
		}

		dialer := &MockConnDialer{
			DialFunc: func(ctx context.Context, host string, config *Config) (Session, error) {
				return session, nil
			},
		}

		detector := &MockPromptDetector{
			DetectFunc: func(output []byte) (matched bool, prompt string) {
				return true, "RTX1200>"
			},
		}

		config := &Config{
			Host:     "192.168.1.1",
			Port:     22,
			Username: "admin",
			Password: "password",
			Timeout:  30,
		}

		client := &rtxClient{
			config:         config,
			dialer:         dialer,
			promptDetector: detector,
			semaphore:      make(chan struct{}, 1),
		}

		ctx := context.Background()

		// Test Dial
		if err := client.Dial(ctx); err != nil {
			t.Fatalf("Dial() failed: %v", err)
		}

		// Test multiple commands
		commands := []Command{
			{Key: "show_version", Payload: "show version"},
			{Key: "show_config", Payload: "show config"},
		}

		for _, cmd := range commands {
			result, err := client.Run(ctx, cmd)
			if err != nil {
				t.Errorf("Run(%s) failed: %v", cmd.Key, err)
				continue
			}

			if result.Raw == nil {
				t.Errorf("Run(%s) returned nil Raw data", cmd.Key)
			}
		}

		// Test Close
		if err := client.Close(); err != nil {
			t.Errorf("Close() failed: %v", err)
		}
	})
}

func TestClient_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() Client
		testFunc  func(c Client) error
		wantErr   error
	}{
		{
			name: "run command without dialing",
			setupFunc: func() Client {
				config := &Config{
					Host:     "192.168.1.1",
					Port:     22,
					Username: "admin",
					Password: "password",
					Timeout:  30,
				}
				return &rtxClient{config: config, semaphore: make(chan struct{}, 1)}
			},
			testFunc: func(c Client) error {
				_, err := c.Run(context.Background(), Command{Key: "test", Payload: "test"})
				return err
			},
			wantErr: errors.New("client not connected"),
		},
		{
			name: "double close",
			setupFunc: func() Client {
				session := &MockSession{}
				config := &Config{
					Host:     "192.168.1.1",
					Port:     22,
					Username: "admin",
					Password: "password",
					Timeout:  30,
				}
				return &rtxClient{
					config:  config,
					session: session,
					active:  true,
				}
			},
			testFunc: func(c Client) error {
				if err := c.Close(); err != nil {
					return err
				}
				// Second close should not error
				return c.Close()
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupFunc()
			err := tt.testFunc(client)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Expected error %v, got nil", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestClientTimeoutHandling(t *testing.T) {
	t.Run("dial timeout", func(t *testing.T) {
		dialer := &MockConnDialer{
			DialFunc: func(ctx context.Context, host string, config *Config) (Session, error) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(1 * time.Second):
					return &MockSession{}, nil
				}
			},
		}

		config := &Config{
			Host:     "192.168.1.1",
			Port:     22,
			Username: "admin",
			Password: "password",
			Timeout:  30,
		}

		client := &rtxClient{
			config: config,
			dialer: dialer,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := client.Dial(ctx)
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
	})

	t.Run("command timeout", func(t *testing.T) {
		session := &MockSession{
			SendFunc: func(cmd string) ([]byte, error) {
				<-time.After(1 * time.Second)
				return []byte("RTX1200>"), nil
			},
		}

		detector := &MockPromptDetector{
			DetectFunc: func(output []byte) (matched bool, prompt string) {
				return true, "RTX1200>"
			},
		}

		config := &Config{
			Host:     "192.168.1.1",
			Port:     22,
			Username: "admin",
			Password: "password",
			Timeout:  30,
		}

		client := &rtxClient{
			config:         config,
			promptDetector: detector,
			retryStrategy:  &MockRetryStrategy{},
			active:         true, // Set as connected for testing
			semaphore:      make(chan struct{}, 1),
		}
		// Initialize executor for the test (session is nil so executor will be used)
		client.executor = NewSSHExecutor(session, detector, &MockRetryStrategy{})

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := client.Run(ctx, Command{Key: "test", Payload: "test"})
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
	})
}

// MockSFTPClientForTest implements SFTPClient for testing
type MockSFTPClientForTest struct {
	DownloadFunc func(ctx context.Context, path string) ([]byte, error)
	CloseFunc    func() error
}

func (m *MockSFTPClientForTest) Download(ctx context.Context, path string) ([]byte, error) {
	if m.DownloadFunc != nil {
		return m.DownloadFunc(ctx, path)
	}
	return []byte("# mock config"), nil
}

func (m *MockSFTPClientForTest) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// MockExecutorForCache implements Executor for testing
type MockExecutorForCache struct {
	RunFunc      func(ctx context.Context, cmd string) ([]byte, error)
	RunBatchFunc func(ctx context.Context, cmds []string) ([]byte, error)
}

func (m *MockExecutorForCache) Run(ctx context.Context, cmd string) ([]byte, error) {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, cmd)
	}
	return []byte("mock response"), nil
}

func (m *MockExecutorForCache) RunBatch(ctx context.Context, cmds []string) ([]byte, error) {
	if m.RunBatchFunc != nil {
		return m.RunBatchFunc(ctx, cmds)
	}
	return []byte("mock batch response"), nil
}

func TestClient_GetCachedConfig_CacheHit(t *testing.T) {
	// Setup client with valid cache
	config := &Config{
		Host:        "192.168.1.1",
		Port:        22,
		Username:    "admin",
		Password:    "password",
		Timeout:     30,
		SFTPEnabled: true,
	}

	cache := NewConfigCache()
	// Pre-populate the cache with valid data
	rawContent := "ip route default gateway 192.168.1.1\n"
	parser := parsers.NewConfigFileParser()
	parsed, _ := parser.Parse(rawContent)
	cache.Set(rawContent, parsed)

	client := &rtxClient{
		config:      config,
		configCache: cache,
		active:      true,
		semaphore:   make(chan struct{}, 1),
	}

	ctx := context.Background()
	result, err := client.GetCachedConfig(ctx)

	if err != nil {
		t.Errorf("GetCachedConfig() unexpected error: %v", err)
	}
	if result == nil {
		t.Error("GetCachedConfig() returned nil on cache hit")
	}
	if result != nil && result.Raw != rawContent {
		t.Errorf("GetCachedConfig() raw = %q, want %q", result.Raw, rawContent)
	}
}

func TestClient_GetCachedConfig_CacheMiss_SFTPSuccess(t *testing.T) {
	// Setup client with empty cache and mock SFTP
	config := &Config{
		Host:        "192.168.1.1",
		Port:        22,
		Username:    "admin",
		Password:    "password",
		Timeout:     30,
		SFTPEnabled: true,
	}

	mockSFTP := &MockSFTPClientForTest{
		DownloadFunc: func(ctx context.Context, path string) ([]byte, error) {
			return []byte("ip route default gateway 192.168.1.1\n"), nil
		},
	}

	mockExecutor := &MockExecutorForCache{
		RunFunc: func(ctx context.Context, cmd string) ([]byte, error) {
			if cmd == "show environment" {
				return []byte("デフォルト設定ファイル: config0\n"), nil
			}
			return []byte("OK"), nil
		},
	}

	cache := NewConfigCache()

	client := &rtxClient{
		config:      config,
		configCache: cache,
		sftpClient:  mockSFTP,
		executor:    mockExecutor,
		active:      true,
		semaphore:   make(chan struct{}, 1),
	}

	ctx := context.Background()
	result, err := client.GetCachedConfig(ctx)

	if err != nil {
		t.Errorf("GetCachedConfig() unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("GetCachedConfig() returned nil after SFTP download")
	}

	// Verify cache is now populated
	cached, ok := cache.Get()
	if !ok {
		t.Error("Cache should be populated after GetCachedConfig()")
	}
	if cached == nil {
		t.Error("Cached value should not be nil")
	}
}

func TestClient_GetCachedConfig_SFTPError_SSHFallback(t *testing.T) {
	// Setup client with SFTP that fails
	config := &Config{
		Host:        "192.168.1.1",
		Port:        22,
		Username:    "admin",
		Password:    "password",
		Timeout:     30,
		SFTPEnabled: true,
	}

	sftpError := errors.New("SFTP connection failed")

	mockSFTP := &MockSFTPClientForTest{
		DownloadFunc: func(ctx context.Context, path string) ([]byte, error) {
			return nil, sftpError
		},
	}

	sshConfigOutput := "ip route default gateway 192.168.1.1\n"
	mockExecutor := &MockExecutorForCache{
		RunFunc: func(ctx context.Context, cmd string) ([]byte, error) {
			if cmd == "show environment" {
				return []byte("デフォルト設定ファイル: config0\n"), nil
			}
			if cmd == "show config" {
				return []byte(sshConfigOutput), nil
			}
			return []byte("OK"), nil
		},
	}

	cache := NewConfigCache()

	client := &rtxClient{
		config:      config,
		configCache: cache,
		sftpClient:  mockSFTP,
		executor:    mockExecutor,
		active:      true,
		semaphore:   make(chan struct{}, 1),
	}

	ctx := context.Background()
	result, err := client.GetCachedConfig(ctx)

	if err != nil {
		t.Errorf("GetCachedConfig() should not return error on SSH fallback, got: %v", err)
	}
	if result == nil {
		t.Fatal("GetCachedConfig() should return config from SSH fallback")
	}
}

func TestClient_GetCachedConfig_CacheDirty_RefetchesConfig(t *testing.T) {
	// Setup client with dirty cache
	config := &Config{
		Host:        "192.168.1.1",
		Port:        22,
		Username:    "admin",
		Password:    "password",
		Timeout:     30,
		SFTPEnabled: true,
	}

	downloadCount := 0
	mockSFTP := &MockSFTPClientForTest{
		DownloadFunc: func(ctx context.Context, path string) ([]byte, error) {
			downloadCount++
			return []byte("ip route default gateway 192.168.1.1\n"), nil
		},
	}

	mockExecutor := &MockExecutorForCache{
		RunFunc: func(ctx context.Context, cmd string) ([]byte, error) {
			if cmd == "show environment" {
				return []byte("デフォルト設定ファイル: config0\n"), nil
			}
			return []byte("OK"), nil
		},
	}

	cache := NewConfigCache()
	// Pre-populate cache
	rawContent := "old config\n"
	parser := parsers.NewConfigFileParser()
	parsed, _ := parser.Parse(rawContent)
	cache.Set(rawContent, parsed)
	// Mark as dirty (simulating a write operation)
	cache.MarkDirty()

	client := &rtxClient{
		config:      config,
		configCache: cache,
		sftpClient:  mockSFTP,
		executor:    mockExecutor,
		active:      true,
		semaphore:   make(chan struct{}, 1),
	}

	ctx := context.Background()
	_, err := client.GetCachedConfig(ctx)

	if err != nil {
		t.Errorf("GetCachedConfig() unexpected error: %v", err)
	}

	// Should have downloaded fresh config since cache was dirty
	if downloadCount == 0 {
		t.Error("GetCachedConfig() should have fetched fresh config when cache is dirty")
	}

	// Dirty flag should be cleared
	if cache.IsDirty() {
		t.Error("Cache dirty flag should be cleared after refetch")
	}
}

func TestClient_InvalidateCache(t *testing.T) {
	config := &Config{
		Host:        "192.168.1.1",
		Port:        22,
		Username:    "admin",
		Password:    "password",
		Timeout:     30,
		SFTPEnabled: true,
	}

	cache := NewConfigCache()
	// Pre-populate the cache
	rawContent := "test config\n"
	parser := parsers.NewConfigFileParser()
	parsed, _ := parser.Parse(rawContent)
	cache.Set(rawContent, parsed)

	client := &rtxClient{
		config:      config,
		configCache: cache,
		active:      true,
		semaphore:   make(chan struct{}, 1),
	}

	// Verify cache is populated
	_, ok := cache.Get()
	if !ok {
		t.Fatal("Cache should be populated before test")
	}

	// Call InvalidateCache
	client.InvalidateCache()

	// Verify cache is now empty
	_, ok = cache.Get()
	if ok {
		t.Error("Cache should be empty after InvalidateCache()")
	}
}

func TestClient_MarkCacheDirty(t *testing.T) {
	config := &Config{
		Host:        "192.168.1.1",
		Port:        22,
		Username:    "admin",
		Password:    "password",
		Timeout:     30,
		SFTPEnabled: true,
	}

	cache := NewConfigCache()

	client := &rtxClient{
		config:      config,
		configCache: cache,
		active:      true,
		semaphore:   make(chan struct{}, 1),
	}

	// Initially not dirty
	if cache.IsDirty() {
		t.Error("Cache should not be dirty initially")
	}

	// Mark dirty
	client.MarkCacheDirty()

	// Should be dirty now
	if !cache.IsDirty() {
		t.Error("Cache should be dirty after MarkCacheDirty()")
	}
}

func TestClient_GetCachedConfig_SFTPDisabled_UsesSSH(t *testing.T) {
	// Setup client with SFTP disabled
	config := &Config{
		Host:        "192.168.1.1",
		Port:        22,
		Username:    "admin",
		Password:    "password",
		Timeout:     30,
		SFTPEnabled: false, // SFTP disabled
	}

	sshConfigOutput := "ip route default gateway 192.168.1.1\n"
	mockExecutor := &MockExecutorForCache{
		RunFunc: func(ctx context.Context, cmd string) ([]byte, error) {
			if cmd == "show config" {
				return []byte(sshConfigOutput), nil
			}
			return []byte("OK"), nil
		},
	}

	cache := NewConfigCache()

	client := &rtxClient{
		config:      config,
		configCache: cache,
		executor:    mockExecutor,
		active:      true,
		semaphore:   make(chan struct{}, 1),
	}

	ctx := context.Background()
	result, err := client.GetCachedConfig(ctx)

	if err != nil {
		t.Errorf("GetCachedConfig() unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("GetCachedConfig() should return config from SSH")
	}
}

func TestClient_GetCachedConfig_NotConnected(t *testing.T) {
	config := &Config{
		Host:        "192.168.1.1",
		Port:        22,
		Username:    "admin",
		Password:    "password",
		Timeout:     30,
		SFTPEnabled: true,
	}

	cache := NewConfigCache()

	client := &rtxClient{
		config:      config,
		configCache: cache,
		active:      false, // Not connected
		semaphore:   make(chan struct{}, 1),
	}

	ctx := context.Background()
	_, err := client.GetCachedConfig(ctx)

	if err == nil {
		t.Error("GetCachedConfig() should return error when not connected")
	}
}
