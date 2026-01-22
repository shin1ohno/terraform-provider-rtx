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
