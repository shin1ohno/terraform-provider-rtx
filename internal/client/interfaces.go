package client

import (
	"context"
	"time"
)

// Client is the main interface for interacting with RTX routers
type Client interface {
	// Dial establishes a connection to the RTX router
	Dial(ctx context.Context) error
	
	// Close terminates the connection
	Close() error
	
	// Run executes a command and returns the result
	Run(ctx context.Context, cmd Command) (Result, error)
	
	// GetSystemInfo retrieves system information from the router
	GetSystemInfo(ctx context.Context) (*SystemInfo, error)
}

// Command represents a command to be executed on the router
type Command struct {
	Key     string // Command identifier for parser lookup
	Payload string // Actual command string to send
}

// Result wraps the raw output and any parsed representation
type Result struct {
	Raw    []byte      // Raw command output
	Parsed interface{} // Parsed/typed representation
}

// Parser converts raw command output into structured data
type Parser interface {
	Parse(raw []byte) (interface{}, error)
}

// PromptDetector identifies when the router prompt appears in output
type PromptDetector interface {
	DetectPrompt(output []byte) (matched bool, prompt string)
}

// RetryStrategy determines retry behavior for transient failures
type RetryStrategy interface {
	Next(retry int) (delay time.Duration, giveUp bool)
}

// Session represents an SSH session with the router
type Session interface {
	Send(cmd string) ([]byte, error)
	Close() error
}

// ConnDialer abstracts SSH connection creation
type ConnDialer interface {
	Dial(ctx context.Context, host string, config *Config) (Session, error)
}

// Config holds the SSH connection configuration
type Config struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Timeout         int    // seconds
	HostKey         string // Fixed host key for verification (base64 encoded)
	KnownHostsFile  string // Path to known_hosts file
	SkipHostKeyCheck bool  // Skip host key verification (insecure)
}