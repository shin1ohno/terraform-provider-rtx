// Package logging provides structured logging utilities for the terraform-provider-rtx.
package logging

import (
	"context"

	"github.com/rs/zerolog"
)

// ctxKey is the context key for storing the logger.
type ctxKey struct{}

// resourceCtxKey is the context key for storing resource information.
type resourceCtxKey struct{}

// ResourceInfo contains information about the current Terraform resource.
type ResourceInfo struct {
	Type string // Resource type (e.g., "rtx_system", "rtx_interface")
	ID   string // Resource ID
}

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

// WithResource returns a new context with resource information attached.
func WithResource(ctx context.Context, resourceType, resourceID string) context.Context {
	return context.WithValue(ctx, resourceCtxKey{}, ResourceInfo{
		Type: resourceType,
		ID:   resourceID,
	})
}

// ResourceFromContext retrieves resource information from context.
// Returns nil if no resource info is attached.
func ResourceFromContext(ctx context.Context) *ResourceInfo {
	if ctx == nil {
		return nil
	}
	if info, ok := ctx.Value(resourceCtxKey{}).(ResourceInfo); ok {
		return &info
	}
	return nil
}
