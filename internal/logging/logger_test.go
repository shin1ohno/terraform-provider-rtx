package logging

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/rs/zerolog"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		tfLog    string
		expected zerolog.Level
	}{
		{
			name:     "trace maps to debug",
			tfLog:    "trace",
			expected: zerolog.DebugLevel,
		},
		{
			name:     "debug level",
			tfLog:    "debug",
			expected: zerolog.DebugLevel,
		},
		{
			name:     "DEBUG uppercase",
			tfLog:    "DEBUG",
			expected: zerolog.DebugLevel,
		},
		{
			name:     "info level",
			tfLog:    "info",
			expected: zerolog.InfoLevel,
		},
		{
			name:     "warn level",
			tfLog:    "warn",
			expected: zerolog.WarnLevel,
		},
		{
			name:     "error level",
			tfLog:    "error",
			expected: zerolog.ErrorLevel,
		},
		{
			name:     "empty defaults to warn",
			tfLog:    "",
			expected: zerolog.WarnLevel,
		},
		{
			name:     "invalid defaults to warn",
			tfLog:    "invalid",
			expected: zerolog.WarnLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLogLevel(tt.tfLog)
			if result != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.tfLog, result, tt.expected)
			}
		})
	}
}

func TestShouldUseJSON(t *testing.T) {
	// Save original env vars
	origJSON := os.Getenv("TF_LOG_JSON")
	origCI := os.Getenv("CI")
	origGHA := os.Getenv("GITHUB_ACTIONS")
	origJenkins := os.Getenv("JENKINS_URL")

	// Clean up after test
	defer func() {
		os.Setenv("TF_LOG_JSON", origJSON)
		os.Setenv("CI", origCI)
		os.Setenv("GITHUB_ACTIONS", origGHA)
		os.Setenv("JENKINS_URL", origJenkins)
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name:     "no env vars returns false",
			envVars:  map[string]string{},
			expected: false,
		},
		{
			name:     "TF_LOG_JSON=1 returns true",
			envVars:  map[string]string{"TF_LOG_JSON": "1"},
			expected: true,
		},
		{
			name:     "CI=true returns true",
			envVars:  map[string]string{"CI": "true"},
			expected: true,
		},
		{
			name:     "GITHUB_ACTIONS=true returns true",
			envVars:  map[string]string{"GITHUB_ACTIONS": "true"},
			expected: true,
		},
		{
			name:     "JENKINS_URL set returns true",
			envVars:  map[string]string{"JENKINS_URL": "http://jenkins"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant env vars
			os.Unsetenv("TF_LOG_JSON")
			os.Unsetenv("CI")
			os.Unsetenv("GITHUB_ACTIONS")
			os.Unsetenv("JENKINS_URL")

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			result := shouldUseJSON()
			if result != tt.expected {
				t.Errorf("shouldUseJSON() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWithContext(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Str("test", "value").Logger()

	ctx := context.Background()
	ctx = WithContext(ctx, logger)

	retrieved := FromContext(ctx)

	// Test that the logger works by logging something
	retrieved.Info().Msg("test message")

	output := buf.String()
	if output == "" {
		t.Error("expected log output, got empty string")
	}
	if !bytes.Contains(buf.Bytes(), []byte("test message")) {
		t.Errorf("expected log to contain 'test message', got %q", output)
	}
}

func TestFromContext_NilContext(t *testing.T) {
	// Should return global logger without panic
	logger := FromContext(nil)
	// Just verify it doesn't panic and returns a valid logger
	logger.Info().Msg("test")
}

func TestFromContext_NoLogger(t *testing.T) {
	ctx := context.Background()
	logger := FromContext(ctx)
	// Should return global logger
	logger.Info().Msg("test")
}

func TestGlobalLogger(t *testing.T) {
	var buf bytes.Buffer
	testLogger := zerolog.New(&buf).With().Str("custom", "logger").Logger()

	// Save original
	original := Global()

	// Set custom global
	SetGlobal(testLogger)

	// Verify it's set
	current := Global()
	current.Info().Msg("test")

	if !bytes.Contains(buf.Bytes(), []byte("custom")) {
		t.Error("expected custom logger field in output")
	}

	// Restore original
	SetGlobal(original)
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	fields := map[string]interface{}{
		"resource_type": "nat_masquerade",
		"resource_id":   123,
	}

	loggerWithFields := WithFields(logger, fields)
	loggerWithFields.Info().Msg("test")

	output := buf.String()
	if !bytes.Contains(buf.Bytes(), []byte("resource_type")) {
		t.Errorf("expected 'resource_type' in output, got %q", output)
	}
	if !bytes.Contains(buf.Bytes(), []byte("nat_masquerade")) {
		t.Errorf("expected 'nat_masquerade' in output, got %q", output)
	}
}
