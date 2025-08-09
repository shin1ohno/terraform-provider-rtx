package client

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// rtxClient is the concrete implementation of the Client interface
type rtxClient struct {
	config         *Config
	dialer         ConnDialer
	promptDetector PromptDetector
	parsers        map[string]Parser
	retryStrategy  RetryStrategy
	
	mu      sync.Mutex
	session Session
	active  bool
}

// NewClient creates a new RTX client instance
func NewClient(config *Config, opts ...Option) (Client, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	
	c := &rtxClient{
		config:         config,
		dialer:         &sshDialer{},
		promptDetector: &defaultPromptDetector{},
		parsers:        make(map[string]Parser),
		retryStrategy:  &noRetry{},
	}
	
	// Apply options
	for _, opt := range opts {
		opt(c)
	}
	
	return c, nil
}

// Option is a functional option for configuring the client
type Option func(*rtxClient)

// WithDialer sets a custom connection dialer
func WithDialer(dialer ConnDialer) Option {
	return func(c *rtxClient) {
		c.dialer = dialer
	}
}

// WithPromptDetector sets a custom prompt detector
func WithPromptDetector(detector PromptDetector) Option {
	return func(c *rtxClient) {
		c.promptDetector = detector
	}
}

// WithParser registers a parser for a specific command
func WithParser(cmdKey string, parser Parser) Option {
	return func(c *rtxClient) {
		c.parsers[cmdKey] = parser
	}
}

// WithRetryStrategy sets the retry strategy for transient failures
func WithRetryStrategy(strategy RetryStrategy) Option {
	return func(c *rtxClient) {
		c.retryStrategy = strategy
	}
}

// Dial establishes a connection to the RTX router
func (c *rtxClient) Dial(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if c.active {
		return nil // Already connected
	}
	
	session, err := c.dialer.Dial(ctx, c.config.Host, c.config)
	if err != nil {
		// Check if it's a specific error type we want to preserve
		if errors.Is(err, ErrAuthFailed) {
			return err
		}
		return fmt.Errorf("%w: %v", ErrDial, err)
	}
	
	c.session = session
	c.active = true
	return nil
}

// Close terminates the connection
func (c *rtxClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if !c.active {
		return nil
	}
	
	err := c.session.Close()
	c.active = false
	c.session = nil
	
	if err != nil {
		return fmt.Errorf("failed to close session: %w", err)
	}
	return nil
}

// Run executes a command and returns the result
func (c *rtxClient) Run(ctx context.Context, cmd Command) (Result, error) {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return Result{}, fmt.Errorf("client not connected")
	}
	session := c.session
	c.mu.Unlock()
	
	// Execute with retry logic
	var raw []byte
	var err error
	
	for retry := 0; ; retry++ {
		select {
		case <-ctx.Done():
			return Result{}, fmt.Errorf("%w: %v", ErrTimeout, ctx.Err())
		default:
		}
		
		// Execute Send operation with context timeout handling
		type sendResult struct {
			data []byte
			err  error
		}
		
		sendCh := make(chan sendResult, 1)
		go func() {
			data, sendErr := session.Send(cmd.Payload)
			sendCh <- sendResult{data: data, err: sendErr}
		}()
		
		select {
		case <-ctx.Done():
			return Result{}, fmt.Errorf("%w: %v", ErrTimeout, ctx.Err())
		case result := <-sendCh:
			raw, err = result.data, result.err
		}
		
		if err == nil {
			break
		}
		
		// Check if we should retry
		delay, giveUp := c.retryStrategy.Next(retry)
		if giveUp {
			return Result{}, fmt.Errorf("command execution failed: %w", err)
		}
		
		select {
		case <-ctx.Done():
			return Result{}, fmt.Errorf("%w: %v", ErrTimeout, ctx.Err())
		case <-time.After(delay):
			// Continue to retry
		}
	}
	
	// Check for prompt
	matched, _ := c.promptDetector.DetectPrompt(raw)
	if !matched {
		return Result{}, fmt.Errorf("%w: output does not contain expected prompt", ErrPrompt)
	}
	
	result := Result{Raw: raw}
	
	// Parse if parser is available
	if parser, ok := c.parsers[cmd.Key]; ok {
		parsed, err := parser.Parse(raw)
		if err != nil {
			return result, fmt.Errorf("%w: %v", ErrParse, err)
		}
		result.Parsed = parsed
	}
	
	return result, nil
}

// validateConfig checks if the configuration is valid
func validateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if config.Host == "" {
		return fmt.Errorf("host is required")
	}
	if config.Username == "" {
		return fmt.Errorf("username is required")
	}
	if config.Password == "" {
		return fmt.Errorf("password is required")
	}
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", config.Port)
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 // Default timeout
	}
	return nil
}