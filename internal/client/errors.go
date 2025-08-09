package client

import "errors"

// Common errors that can occur when interacting with RTX routers
var (
	// ErrAuthFailed indicates authentication failure
	ErrAuthFailed = errors.New("authentication failed")
	
	// ErrDial indicates connection establishment failure
	ErrDial = errors.New("dial failed")
	
	// ErrParse indicates command output parsing failure
	ErrParse = errors.New("parse failed")
	
	// ErrPrompt indicates the router prompt was not found in output
	ErrPrompt = errors.New("prompt not found")
	
	// ErrTimeout indicates operation timeout
	ErrTimeout = errors.New("operation timeout")
	
	// ErrCommandFailed indicates the command execution failed on the router
	ErrCommandFailed = errors.New("command failed")
	
	// ErrHostKeyMismatch indicates SSH host key verification failed
	ErrHostKeyMismatch = errors.New("host key verification failed")
)