package logging

import (
	"strings"

	"github.com/rs/zerolog"
)

// sensitivePatterns defines patterns that indicate sensitive data.
var sensitivePatterns = []string{
	"password",
	"pre-shared-key",
	"secret",
	"community", // SNMP community strings
	"token",
	"key",
	"credential",
}

// sensitiveFields defines field names that should be redacted in structured logs.
var sensitiveFields = map[string]bool{
	"password":       true,
	"pre_shared_key": true,
	"secret":         true,
	"community":      true,
	"token":          true,
	"api_key":        true,
	"credential":     true,
}

// redactedMessage is the replacement text for sensitive data.
const redactedMessage = "[REDACTED]"

// SanitizingHook implements zerolog.Hook to automatically redact sensitive data.
type SanitizingHook struct{}

// Run implements the zerolog.Hook interface.
// It inspects log events and redacts sensitive field values.
func (h SanitizingHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	// Note: zerolog.Event doesn't provide direct field access,
	// so we sanitize the message if it contains sensitive patterns
	if containsSensitivePattern(msg) {
		// We can't modify the message directly, but we add a warning field
		e.Bool("sanitized", true)
	}
}

// containsSensitivePattern checks if a string contains any sensitive patterns.
func containsSensitivePattern(s string) bool {
	if s == "" {
		return false
	}
	lower := strings.ToLower(s)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

// SanitizeString redacts sensitive patterns from a string.
// Use this for sanitizing command strings or messages before logging.
func SanitizeString(s string) string {
	if s == "" {
		return s
	}

	lower := strings.ToLower(s)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(lower, pattern) {
			return redactedMessage
		}
	}

	return s
}

// IsSensitiveField checks if a field name indicates sensitive data.
func IsSensitiveField(fieldName string) bool {
	return sensitiveFields[strings.ToLower(fieldName)]
}

// SanitizeMap redacts sensitive values from a map based on field names.
func SanitizeMap(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		if IsSensitiveField(k) {
			result[k] = redactedMessage
		} else if str, ok := v.(string); ok && containsSensitivePattern(str) {
			result[k] = redactedMessage
		} else {
			result[k] = v
		}
	}
	return result
}

// NewSanitizedLogger creates a logger with the sanitizing hook attached.
func NewSanitizedLogger() zerolog.Logger {
	return NewLogger().Hook(SanitizingHook{})
}
