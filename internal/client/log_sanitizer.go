package client

import (
	"github.com/sh1/terraform-provider-rtx/internal/logging"
	"strings"
)

// sensitivePatterns defines patterns that indicate sensitive data in commands
var sensitivePatterns = []string{
	"password",
	"pre-shared-key",
	"secret",
	"community", // SNMP community strings
}

// redactedMessage is the replacement text for sensitive commands
const redactedMessage = "[REDACTED - contains sensitive data]"

// SanitizeCommandForLog redacts sensitive patterns from command strings for safe logging.
// It performs case-insensitive matching against known sensitive patterns.
// If any pattern is found, the entire command is replaced with a redaction message.
func SanitizeCommandForLog(cmd string) string {
	if cmd == "" {
		return cmd
	}

	cmdLower := strings.ToLower(cmd)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(cmdLower, pattern) {
			return redactedMessage
		}
	}

	return cmd
}

// LogCommand logs a command with automatic sanitization for sensitive data.
// Use this helper instead of direct logging for command logging.
// Deprecated: Use logging.FromContext(ctx).Debug().Str("command", SanitizeCommandForLog(cmd)).Msg() instead
func LogCommand(prefix, cmd string) {
	logging.Global().Debug().Str("prefix", prefix).Str("command", SanitizeCommandForLog(cmd)).Msg("Command execution")
}
