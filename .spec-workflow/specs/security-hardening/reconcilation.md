# Reconciliation

## Product principles
- Hardening aims to protect credentials while keeping existing Cisco-aligned provider behavior; state stays config-only.

## Implementation alignment
- Sanitization utilities exist (`internal/logging/sanitizer.go`) and zerolog hook marks events containing sensitive patterns; `SanitizeString` can redact commands.
- `skip_host_key_check` option is exposed in provider schema.
- Gaps: sanitization does not replace logged commands with redacted text uniformly (adds `sanitized` flag only), sensitive fields may still reach log.Printf call sites, and no warnings are emitted when `skip_host_key_check` is true or host key sources are absent. `.gitignore` does not enforce the spec’s secret patterns. Sensitive prompt responses are not explicitly suppressed.
