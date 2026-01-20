// Package logging provides structured logging utilities for the terraform-provider-rtx.
package logging

import (
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// globalLogger is the default logger used when no logger is in context.
var globalLogger zerolog.Logger

func init() {
	// Initialize global logger with default configuration
	globalLogger = NewLogger()
}

// NewLogger creates a new zerolog logger configured based on environment variables.
// It reads TF_LOG for log level (debug, info, warn, error) and TF_LOG_JSON for output format.
func NewLogger() zerolog.Logger {
	level := parseLogLevel(os.Getenv("TF_LOG"))

	var output io.Writer
	if shouldUseJSON() {
		output = os.Stderr
	} else {
		output = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
		}
	}

	return zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()
}

// parseLogLevel parses the TF_LOG environment variable into a zerolog.Level.
// Supports: debug, info, warn, error. Defaults to warn if unset or invalid.
func parseLogLevel(tfLog string) zerolog.Level {
	switch strings.ToLower(tfLog) {
	case "trace", "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		// Default to warn level
		return zerolog.WarnLevel
	}
}

// shouldUseJSON returns true if JSON output format should be used.
// JSON is used when TF_LOG_JSON=1 or when running in CI environment.
func shouldUseJSON() bool {
	if os.Getenv("TF_LOG_JSON") == "1" {
		return true
	}
	// Detect common CI environments
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" || os.Getenv("JENKINS_URL") != "" {
		return true
	}
	return false
}

// Global returns a pointer to the global logger.
func Global() *zerolog.Logger {
	return &globalLogger
}

// SetGlobal sets the global logger.
func SetGlobal(logger zerolog.Logger) {
	globalLogger = logger
}

// WithFields returns a logger with the specified fields attached.
func WithFields(logger zerolog.Logger, fields map[string]interface{}) zerolog.Logger {
	ctx := logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return ctx.Logger()
}
