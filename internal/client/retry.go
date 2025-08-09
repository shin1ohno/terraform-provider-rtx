package client

import (
	"crypto/rand"
	"math"
	"math/big"
	"time"
)

// noRetry is a retry strategy that never retries
type noRetry struct{}

func (r *noRetry) Next(retry int) (delay time.Duration, giveUp bool) {
	return 0, true
}

// ExponentialBackoff implements an exponential backoff retry strategy
type ExponentialBackoff struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
	MaxRetries int
}

// NewExponentialBackoff creates a new exponential backoff strategy with defaults
func NewExponentialBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   10 * time.Second,
		MaxRetries: 5,
	}
}

// Next calculates the next retry delay
func (r *ExponentialBackoff) Next(retry int) (delay time.Duration, giveUp bool) {
	if retry >= r.MaxRetries {
		return 0, true
	}
	
	// Calculate exponential delay with jitter
	delay = time.Duration(float64(r.BaseDelay) * math.Pow(2, float64(retry)))
	if delay > r.MaxDelay {
		delay = r.MaxDelay
	}
	
	// Add jitter (±10%) using cryptographically secure random
	jitterMax := int64(float64(delay) * 0.1)
	if jitterMax > 0 {
		n, err := rand.Int(rand.Reader, big.NewInt(jitterMax*2))
		if err == nil {
			jitter := time.Duration(n.Int64() - jitterMax)
			delay += jitter
		}
	}
	
	return delay, false
}

// LinearBackoff implements a linear backoff retry strategy
type LinearBackoff struct {
	Delay      time.Duration
	MaxRetries int
}

// NewLinearBackoff creates a new linear backoff strategy
func NewLinearBackoff(delay time.Duration, maxRetries int) *LinearBackoff {
	return &LinearBackoff{
		Delay:      delay,
		MaxRetries: maxRetries,
	}
}

// Next returns a constant delay for each retry
func (r *LinearBackoff) Next(retry int) (delay time.Duration, giveUp bool) {
	if retry >= r.MaxRetries {
		return 0, true
	}
	return r.Delay, false
}

// RetryableError marks an error as retryable
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// IsRetryable checks if an error should trigger a retry
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	
	// Check if it's explicitly marked as retryable
	if _, ok := err.(*RetryableError); ok {
		return true
	}
	
	// Check for specific error conditions that are retryable
	switch err {
	case ErrTimeout:
		return true
	default:
		return false
	}
}