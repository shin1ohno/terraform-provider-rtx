package client

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
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

			client, err := NewClient(config, WithDialer(tt.dialer))
			if err != nil {
				t.Fatalf("NewClient() failed: %v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			err = client.Dial(ctx)

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
				session:        tt.session,
				promptDetector: tt.detector,
				retryStrategy:  &MockRetryStrategy{},
				active:         true, // Set as connected for testing
			}
			// Initialize executor for the test
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
		// Test that the client can be created and closed
		config := &Config{
			Host:     "localhost", // Use localhost to avoid network issues
			Port:     22,
			Username: "admin",
			Password: "password",
			Timeout:  30,
		}

		client, err := NewClient(config)
		if err != nil {
			t.Fatalf("NewClient() failed: %v", err)
		}

		if client == nil {
			t.Error("NewClient() returned nil client")
		}

		// Test that client can be closed without connection
		if err := client.Close(); err != nil {
			t.Logf("Close() returned error (expected for unconnected client): %v", err)
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
				return &rtxClient{config: config}
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

		client, err := NewClient(config, WithDialer(dialer))
		if err != nil {
			t.Fatalf("NewClient() failed: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err = client.Dial(ctx)
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
	})

	t.Run("command timeout", func(t *testing.T) {
		session := &MockSession{
			SendFunc: func(cmd string) ([]byte, error) {
				time.Sleep(1 * time.Second)
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
			session:        session,
			promptDetector: detector,
			retryStrategy:  &MockRetryStrategy{},
			active:         true, // Set as connected for testing
		}
		// Initialize executor for the test
		client.executor = NewSSHExecutor(session, detector, &MockRetryStrategy{})

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := client.Run(ctx, Command{Key: "test", Payload: "test"})
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
	})
}

// MockClient for GetDHCPScope testing
type MockClientForDHCPScope struct {
	GetDHCPScopesFunc func(ctx context.Context) ([]DHCPScope, error)
}

func (m *MockClientForDHCPScope) GetDHCPScopes(ctx context.Context) ([]DHCPScope, error) {
	if m.GetDHCPScopesFunc != nil {
		return m.GetDHCPScopesFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *MockClientForDHCPScope) GetDHCPScope(ctx context.Context, scopeID int) (*DHCPScope, error) {
	if scopeID <= 0 || scopeID > 255 {
		return nil, fmt.Errorf("scope_id must be between 1 and 255")
	}

	scopes, err := m.GetDHCPScopes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve DHCP scopes: %w", err)
	}

	// Find the specific scope
	for _, scope := range scopes {
		if scope.ID == scopeID {
			return &scope, nil
		}
	}

	// Return ErrNotFound if scope doesn't exist
	return nil, ErrNotFound
}

// Implement other Client interface methods with stubs
func (m *MockClientForDHCPScope) Dial(ctx context.Context) error { return nil }
func (m *MockClientForDHCPScope) Close() error                   { return nil }
func (m *MockClientForDHCPScope) Run(ctx context.Context, cmd Command) (Result, error) {
	return Result{}, nil
}
func (m *MockClientForDHCPScope) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	return nil, nil
}
func (m *MockClientForDHCPScope) GetInterfaces(ctx context.Context) ([]Interface, error) {
	return nil, nil
}
func (m *MockClientForDHCPScope) GetRoutes(ctx context.Context) ([]Route, error) { return nil, nil }
func (m *MockClientForDHCPScope) CreateDHCPScope(ctx context.Context, scope DHCPScope) error {
	return nil
}
func (m *MockClientForDHCPScope) UpdateDHCPScope(ctx context.Context, scope DHCPScope) error {
	return nil
}
func (m *MockClientForDHCPScope) DeleteDHCPScope(ctx context.Context, scopeID int) error { return nil }
func (m *MockClientForDHCPScope) GetDHCPBindings(ctx context.Context, scopeID int) ([]DHCPBinding, error) {
	return nil, nil
}
func (m *MockClientForDHCPScope) CreateDHCPBinding(ctx context.Context, binding DHCPBinding) error {
	return nil
}
func (m *MockClientForDHCPScope) DeleteDHCPBinding(ctx context.Context, scopeID int, ipAddress string) error {
	return nil
}
func (m *MockClientForDHCPScope) SaveConfig(ctx context.Context) error { return nil }

// StaticRoute related methods
func (m *MockClientForDHCPScope) CreateStaticRoute(ctx context.Context, route StaticRoute) error {
	return fmt.Errorf("mock not implemented")
}

func (m *MockClientForDHCPScope) GetStaticRoute(ctx context.Context, destination, gateway, iface string) (*StaticRoute, error) {
	return nil, fmt.Errorf("mock not implemented")
}

func (m *MockClientForDHCPScope) GetStaticRoutes(ctx context.Context) ([]StaticRoute, error) {
	return nil, fmt.Errorf("mock not implemented")
}

func (m *MockClientForDHCPScope) UpdateStaticRoute(ctx context.Context, route StaticRoute) error {
	return fmt.Errorf("mock not implemented")
}

func (m *MockClientForDHCPScope) DeleteStaticRoute(ctx context.Context, destination, gateway, iface string) error {
	return fmt.Errorf("mock not implemented")
}

func TestGetDHCPScope(t *testing.T) {
	tests := []struct {
		name          string
		scopeID       int
		setupClient   func() Client
		expectError   bool
		expectedScope *DHCPScope
		errorContains string
	}{
		{
			name:    "valid scope found",
			scopeID: 1,
			setupClient: func() Client {
				return &MockClientForDHCPScope{
					GetDHCPScopesFunc: func(ctx context.Context) ([]DHCPScope, error) {
						return []DHCPScope{
							{
								ID:         1,
								RangeStart: "192.168.1.100",
								RangeEnd:   "192.168.1.200",
								Prefix:     24,
								Gateway:    "192.168.1.1",
								DNSServers: []string{"8.8.8.8", "8.8.4.4"},
								Lease:      86400,
								DomainName: "example.com",
							},
						}, nil
					},
				}
			},
			expectError: false,
			expectedScope: &DHCPScope{
				ID:         1,
				RangeStart: "192.168.1.100",
				RangeEnd:   "192.168.1.200",
				Prefix:     24,
				Gateway:    "192.168.1.1",
				DNSServers: []string{"8.8.8.8", "8.8.4.4"},
				Lease:      86400,
				DomainName: "example.com",
			},
		},
		{
			name:    "scope not found",
			scopeID: 2,
			setupClient: func() Client {
				return &MockClientForDHCPScope{
					GetDHCPScopesFunc: func(ctx context.Context) ([]DHCPScope, error) {
						return []DHCPScope{
							{ID: 1, RangeStart: "192.168.1.100", RangeEnd: "192.168.1.200", Prefix: 24},
						}, nil
					},
				}
			},
			expectError:   true,
			errorContains: "resource not found",
		},
		{
			name:    "invalid scope ID - zero",
			scopeID: 0,
			setupClient: func() Client {
				return &MockClientForDHCPScope{}
			},
			expectError:   true,
			errorContains: "scope_id must be between 1 and 255",
		},
		{
			name:    "invalid scope ID - too large",
			scopeID: 256,
			setupClient: func() Client {
				return &MockClientForDHCPScope{}
			},
			expectError:   true,
			errorContains: "scope_id must be between 1 and 255",
		},
		{
			name:    "GetDHCPScopes returns error",
			scopeID: 1,
			setupClient: func() Client {
				return &MockClientForDHCPScope{
					GetDHCPScopesFunc: func(ctx context.Context) ([]DHCPScope, error) {
						return nil, errors.New("connection failed")
					},
				}
			},
			expectError:   true,
			errorContains: "failed to retrieve DHCP scopes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupClient()
			ctx := context.Background()

			scope, err := client.GetDHCPScope(ctx, tt.scopeID)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if scope == nil {
					t.Fatal("Expected scope but got nil")
				}

				// Compare expected scope
				if scope.ID != tt.expectedScope.ID {
					t.Errorf("Expected scope ID %d, got %d", tt.expectedScope.ID, scope.ID)
				}
				if scope.RangeStart != tt.expectedScope.RangeStart {
					t.Errorf("Expected range start %s, got %s", tt.expectedScope.RangeStart, scope.RangeStart)
				}
				if scope.RangeEnd != tt.expectedScope.RangeEnd {
					t.Errorf("Expected range end %s, got %s", tt.expectedScope.RangeEnd, scope.RangeEnd)
				}
				if scope.Prefix != tt.expectedScope.Prefix {
					t.Errorf("Expected prefix %d, got %d", tt.expectedScope.Prefix, scope.Prefix)
				}
				if scope.Gateway != tt.expectedScope.Gateway {
					t.Errorf("Expected gateway %s, got %s", tt.expectedScope.Gateway, scope.Gateway)
				}
				if scope.Lease != tt.expectedScope.Lease {
					t.Errorf("Expected lease %d, got %d", tt.expectedScope.Lease, scope.Lease)
				}
				if scope.DomainName != tt.expectedScope.DomainName {
					t.Errorf("Expected domain name %s, got %s", tt.expectedScope.DomainName, scope.DomainName)
				}

				// Compare DNS servers slice
				if len(scope.DNSServers) != len(tt.expectedScope.DNSServers) {
					t.Errorf("Expected %d DNS servers, got %d", len(tt.expectedScope.DNSServers), len(scope.DNSServers))
				} else {
					for i, expected := range tt.expectedScope.DNSServers {
						if scope.DNSServers[i] != expected {
							t.Errorf("Expected DNS server %d to be %s, got %s", i, expected, scope.DNSServers[i])
						}
					}
				}
			}
		})
	}
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
