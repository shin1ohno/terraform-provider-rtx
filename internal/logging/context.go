// Package logging provides structured logging utilities for the terraform-provider-rtx.
package logging

import (
	"context"

	"github.com/rs/zerolog"
)

// ctxKey is the context key for storing the logger.
type ctxKey struct{}

// WithContext returns a new context with the logger attached.
func WithContext(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

// FromContext retrieves the logger from context.
// Returns a pointer to the global logger if no logger is attached to the context.
func FromContext(ctx context.Context) *zerolog.Logger {
	if ctx == nil {
		return &globalLogger
	}
	if logger, ok := ctx.Value(ctxKey{}).(zerolog.Logger); ok {
		return &logger
	}
	return &globalLogger
}
